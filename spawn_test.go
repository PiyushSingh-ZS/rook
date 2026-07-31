package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func gitInitRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	repo := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "t@example.com")
	run("config", "user.name", "tester")
	if err := os.WriteFile(filepath.Join(repo, "f.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-q", "-m", "init")
	return repo
}

func TestGitToplevel(t *testing.T) {
	repo := gitInitRepo(t)
	if got := gitToplevel(repo); got == "" {
		t.Error("expected a toplevel for a git repo")
	}
	if got := gitToplevel(t.TempDir()); got != "" {
		t.Errorf("expected empty toplevel for a non-repo, got %q", got)
	}
}

func TestCreateWorktree_IsolatesCheckout(t *testing.T) {
	repo := gitInitRepo(t)
	home := t.TempDir()
	t.Setenv("HOME", home)

	wt, err := createWorktree(repo, "review-pr-1", 12345)
	if err != nil {
		t.Fatalf("createWorktree: %v", err)
	}
	// worktree is a separate directory containing the repo's files
	if _, err := os.Stat(filepath.Join(wt, "f.txt")); err != nil {
		t.Errorf("worktree missing repo files: %v", err)
	}
	if wt == repo {
		t.Error("worktree must not be the repo itself")
	}
	if !strings.Contains(wt, filepath.Join(".rook", "worktrees")) {
		t.Errorf("worktree should live under ~/.rook/worktrees, got %q", wt)
	}
	// git recognizes it as a linked worktree of the repo
	out, err := exec.Command("git", "-C", repo, "worktree", "list").CombinedOutput()
	if err != nil {
		t.Fatalf("worktree list: %s", out)
	}
	if !strings.Contains(string(out), wt) {
		t.Errorf("worktree not registered with repo:\n%s", out)
	}
}

// TestSendInitialPrompt_WaitsForBoot verifies the prompt is delivered only after
// the TUI stops changing (booting), not lost during the splash.
func TestSendInitialPrompt_WaitsForBoot(t *testing.T) {
	if tmuxBin == "" {
		t.Skip("tmux not available")
	}
	target := "rook_boottest"
	_, _ = runTmux("kill-session", "-t", target)
	// "boots" for ~3s (output keeps changing) then settles at a cat prompt.
	if _, err := runTmux("new-session", "-d", "-s", target, "-x", "100", "-y", "30",
		"bash", "-lc", "for i in 1 2 3; do clear; echo boot $i; sleep 1; done; clear; echo READY; cat"); err != nil {
		t.Fatalf("new-session: %v", err)
	}
	defer runTmux("kill-session", "-t", target)

	sendInitialPrompt(target, "HELLO_ROOK_PROMPT")
	time.Sleep(600 * time.Millisecond)
	out, _ := runTmux("capture-pane", "-p", "-t", target)
	if !strings.Contains(string(out), "HELLO_ROOK_PROMPT") {
		t.Errorf("prompt not delivered after boot settle; pane:\n%s", out)
	}
}
