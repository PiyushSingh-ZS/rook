package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gofr.dev/pkg/gofr"
)

// execWithTimeout runs a command with a hard deadline, returning combined output.
func execWithTimeout(bin string, d time.Duration, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	return exec.CommandContext(ctx, bin, args...).CombinedOutput()
}

var (
	tmuxNameRe   = regexp.MustCompile(`^[A-Za-z0-9_-]{1,40}$`)
	paneTargetRe = regexp.MustCompile(`^[%A-Za-z0-9_:.@=-]{1,60}$`)
)

// agentCmd maps a provider to the shell command that launches it.
func agentCmd(agent string) string {
	switch agent {
	case "codex":
		return "codex"
	case "aider":
		return "aider"
	case "gemini":
		return "gemini"
	default:
		return "claude"
	}
}

type spawnReq struct {
	Name     string `json:"name"`
	CWD      string `json:"cwd"`
	Agent    string `json:"agent"`
	Prompt   string `json:"prompt"`
	Worktree bool   `json:"worktree"` // isolate in a git worktree (don't touch the user's checkout)
	Resume   string `json:"resume"`   // Claude session id to resume (restores full context; claude only)
}

// gitToplevel returns the repo root for a directory, or "" if not a git repo.
func gitToplevel(dir string) string {
	out, err := execWithTimeout("git", 8*time.Second, "-C", dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// createWorktree makes a detached-HEAD git worktree off the repo so an agent can
// check out a PR branch / create a branch without disturbing the user's working
// tree. Returns the new worktree path.
func createWorktree(repoRoot, name string, ts int64) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	base := filepath.Join(home, ".rook", "worktrees")
	if err := os.MkdirAll(base, 0o755); err != nil {
		return "", err
	}
	dir := filepath.Join(base, fmt.Sprintf("%s-%d", name, ts))
	if out, err := execWithTimeout("git", 30*time.Second, "-C", repoRoot, "worktree", "add", "--detach", dir); err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return dir, nil
}

// handleSpawn launches an agent inside a fresh tmux session so Foreman can watch
// and drive it. This is the one write-ish action in orchestration; the UI
// confirms before calling it.
func handleSpawn(ctx *gofr.Context) (any, error) {
	var req spawnReq
	if err := ctx.Bind(&req); err != nil {
		return nil, errf(http.StatusBadRequest, "bad request")
	}
	req.Name = strings.TrimSpace(req.Name)
	cmd, worktree, code, err := spawnAgentSession(req)
	if err != nil {
		return nil, errf(code, "%v", err)
	}
	return rawJSON(map[string]any{"ok": true, "session": req.Name, "agent": cmd, "worktree": worktree})
}

// spawnAgentSession creates the tmux session, optional worktree, and schedules
// the initial prompt. Shared by the HTTP handler and the summary scheduler.
// Returns (agentCmd, worktreePath, httpStatusOnError, error).
func spawnAgentSession(req spawnReq) (string, string, int, error) {
	if tmuxBin == "" {
		return "", "", http.StatusServiceUnavailable, fmt.Errorf("tmux not installed — needed to spawn/control agents")
	}
	if !tmuxNameRe.MatchString(req.Name) {
		return "", "", http.StatusBadRequest, fmt.Errorf("name must be 1-40 chars [A-Za-z0-9_-]")
	}
	fi, err := os.Stat(req.CWD)
	if err != nil || !fi.IsDir() {
		return "", "", http.StatusBadRequest, fmt.Errorf("cwd is not a directory")
	}
	cmd := agentCmd(req.Agent)

	// Resume restores an existing session's full context/history in place (same
	// session id). Only Claude supports it; the initial prompt is skipped because
	// the conversation is picked up where it left off.
	launch := cmd
	if req.Resume != "" {
		if !validSessionID(req.Resume) {
			return "", "", http.StatusBadRequest, fmt.Errorf("invalid resume session id")
		}
		if req.Agent != "" && req.Agent != "claude" {
			return "", "", http.StatusBadRequest, fmt.Errorf("resume is only supported for Claude sessions")
		}
		launch = cmd + " --resume " + req.Resume
	}

	// Isolate in a git worktree so the agent can `gh pr checkout` / branch without
	// switching the user's current branch (GitHub review/work handoffs).
	runDir := req.CWD
	worktree := ""
	if req.Worktree {
		root := gitToplevel(req.CWD)
		if root == "" {
			return "", "", http.StatusBadRequest, fmt.Errorf("worktree requested but this is not a git repo")
		}
		wt, werr := createWorktree(root, req.Name, time.Now().Unix())
		if werr != nil {
			return "", "", http.StatusConflict, fmt.Errorf("worktree failed: %v", werr)
		}
		runDir = wt
		worktree = wt
	}

	if out, serr := runTmux("new-session", "-d", "-s", req.Name, "-x", "220", "-y", "50", "-c", runDir, launch); serr != nil {
		return "", "", http.StatusConflict, fmt.Errorf("spawn failed: %s", tmuxErr(serr, out))
	}
	if worktree != "" {
		rememberWorktree(req.Name, worktree)
	}

	// send the initial prompt once the agent's UI has settled (non-blocking);
	// collapse to one line so a multi-line prompt doesn't submit early.
	if p := strings.Join(strings.Fields(req.Prompt), " "); p != "" {
		go sendInitialPrompt(req.Name, p)
	}
	return cmd, worktree, 0, nil
}

// sendInitialPrompt waits for the agent's TUI to finish booting (the pane output
// stops changing) before typing the prompt, so keystrokes aren't lost during the
// splash/spinner. Falls back to sending after a max wait.
func sendInitialPrompt(target, prompt string) {
	deadline := time.Now().Add(25 * time.Second)
	var last string
	stable := 0
	for time.Now().Before(deadline) {
		time.Sleep(700 * time.Millisecond)
		out, err := runTmux("capture-pane", "-p", "-t", target)
		if err != nil {
			continue // pane not ready yet
		}
		cur := string(out)
		if strings.TrimSpace(cur) != "" && cur == last {
			stable++
			if stable >= 2 { // unchanged across two polls → booted
				break
			}
		} else {
			stable = 0
		}
		last = cur
	}
	time.Sleep(500 * time.Millisecond)
	if err := tmuxSendKeys(target, true, prompt); err != nil {
		return
	}
	time.Sleep(200 * time.Millisecond)
	_ = tmuxSendKeys(target, false, "Enter")
}

// applyKeyAction sends a control action's keystrokes to a tmux target. Returns
// the HTTP status to use on error (0 on success).
func applyKeyAction(target, action, value string) (int, error) {
	switch action {
	case "allow":
		return 0, tmuxSendKeys(target, false, "Enter")
	case "deny":
		return 0, tmuxSendKeys(target, false, "Escape")
	case "interrupt":
		return 0, tmuxSendKeys(target, false, "C-c")
	case "key":
		if !isMenuKey(value) {
			return http.StatusBadRequest, fmt.Errorf("key must be 1-9")
		}
		return 0, tmuxSendKeys(target, false, value)
	case "text":
		if strings.TrimSpace(value) == "" {
			return http.StatusBadRequest, fmt.Errorf("empty text")
		}
		if err := tmuxSendKeys(target, true, value); err != nil {
			return http.StatusInternalServerError, err
		}
		return 0, tmuxSendKeys(target, false, "Enter")
	default:
		return http.StatusBadRequest, fmt.Errorf("unknown action")
	}
}

type sendReq struct {
	Target string `json:"target"`
	Action string `json:"action"`
	Value  string `json:"value"`
}

// handleSend drives a live terminal by sending keys straight to its tmux target.
// This is what makes the Terminal interactive for any session — including
// handed-off agents whose Claude session id isn't resolved yet.
func handleSend(ctx *gofr.Context) (any, error) {
	if tmuxBin == "" {
		return nil, errf(http.StatusServiceUnavailable, "tmux not installed")
	}
	var req sendReq
	if err := ctx.Bind(&req); err != nil || !paneTargetRe.MatchString(req.Target) {
		return nil, errf(http.StatusBadRequest, "bad request")
	}
	if code, err := applyKeyAction(req.Target, req.Action, req.Value); err != nil {
		return nil, errf(code, "%v", err)
	}
	return rawJSON(map[string]any{"ok": true, "target": req.Target, "action": req.Action})
}

type killReq struct {
	SessionID string `json:"sessionId"`
	Target    string `json:"target"`
}

// handleKill terminates a tmux-controlled agent by killing its pane. Accepts
// either a session id (resolved to its pane) or a direct tmux target.
func handleKill(ctx *gofr.Context) (any, error) {
	if tmuxBin == "" {
		return nil, errf(http.StatusServiceUnavailable, "tmux not installed")
	}
	var req killReq
	if err := ctx.Bind(&req); err != nil {
		return nil, errf(http.StatusBadRequest, "bad request")
	}
	var pane string
	if req.Target != "" && paneTargetRe.MatchString(req.Target) {
		pane = req.Target
	} else if validSessionID(req.SessionID) {
		for _, s := range ScanSessions(0) {
			if s.SessionID == req.SessionID {
				pane = s.TmuxPane
				break
			}
		}
	} else {
		return nil, errf(http.StatusBadRequest, "bad request")
	}
	if pane == "" {
		return nil, errf(http.StatusConflict, "session is not in a tmux pane")
	}
	if out, err := runTmux("kill-pane", "-t", pane); err != nil {
		return nil, errf(http.StatusInternalServerError, "kill failed: %s", tmuxErr(err, out))
	}
	// auto-remove the isolated worktree this agent ran in, if any
	if wt := worktreeForTarget(pane); wt != "" {
		forgetWorktree(pane)
		go removeWorktree(wt)
	}
	return rawJSON(map[string]any{"ok": true, "pane": pane})
}

// handlePaneCapture returns the live text of a tmux pane (read-only peek).
func handlePaneCapture(ctx *gofr.Context) (any, error) {
	if tmuxBin == "" {
		return nil, errf(http.StatusServiceUnavailable, "tmux not installed")
	}
	target := ctx.Param("target")
	if !paneTargetRe.MatchString(target) {
		return nil, errf(http.StatusBadRequest, "bad target")
	}
	lines := 200
	if n, err := strconv.Atoi(ctx.Param("lines")); err == nil && n > 0 && n <= 1000 {
		lines = n
	}
	// -e preserves ANSI colour/attribute escape sequences so the UI can render
	// the agent's output in colour (like a real terminal).
	out, err := runTmux("capture-pane", "-p", "-e", "-t", target, "-S", "-"+strconv.Itoa(lines))
	if err != nil {
		return nil, errf(http.StatusNotFound, "no such pane")
	}
	return textResp(out, "text/plain; charset=utf-8")
}

// runTmux runs a tmux subcommand and returns combined output.
func runTmux(args ...string) ([]byte, error) {
	return execWithTimeout(tmuxBin, 10*time.Second, args...)
}

func tmuxErr(err error, out []byte) string {
	if len(out) > 0 {
		return strings.TrimSpace(string(out))
	}
	return err.Error()
}
