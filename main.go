package main

import (
	"embed"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gofr.dev/pkg/gofr"
	resp "gofr.dev/pkg/gofr/http/response"
)

//go:embed web
var webFS embed.FS

const maxToolsPerSession = 40

// appVersion is bumped with each asset change; the client force-reloads on mismatch.
const appVersion = "113"

func main() {
	addr := flag.String("addr", "127.0.0.1:7480", "listen address (loopback only by default)")
	notifyOn := flag.Bool("notify", true, "desktop notification when an agent starts waiting")
	token := flag.String("token", "", "if set, non-loopback clients (e.g. over Tailscale) must present this token via X-Rook-Token header or ?token=")
	flag.Parse()

	port := portOf(*addr)
	// GoFr is configured via env: run its HTTP server on our port, and disable
	// the extra metrics server (this is a loopback single-user tool).
	os.Setenv("HTTP_PORT", port)
	os.Setenv("METRICS_PORT", "0")

	if err := initDB(); err != nil {
		log.Printf("summary store disabled: %v", err)
	}
	if *notifyOn {
		startNotifier(3 * time.Second)
	}
	startSummaryScheduler("http://127.0.0.1:" + port)
	startAuditIngester(60 * time.Second)

	app := gofr.New()

	// Enforce loopback-only (or token for remote) + no-store caching at the app
	// layer — GoFr binds all interfaces, so the guard is what keeps rook local.
	app.UseMiddleware(accessGuard(*token))
	app.UseMiddleware(termWSMiddleware) // intercept /ws/term for the interactive PTY bridge
	app.UseMiddleware(noCache)

	// static UI (embedded). The Operator console is the front door; the
	// original dashboard stays reachable at /classic during the migration.
	app.GET("/", staticHandler("operator.html", "text/html; charset=utf-8"))
	app.GET("/operator.html", staticHandler("operator.html", "text/html; charset=utf-8"))
	app.GET("/operator.css", staticHandler("operator.css", "text/css; charset=utf-8"))
	app.GET("/operator.js", staticHandler("operator.js", "text/javascript; charset=utf-8"))
	for _, v := range []string{"github", "summaries", "dev", "audit", "workspace"} {
		name := "operator-" + v + ".js"
		app.GET("/"+name, staticHandler(name, "text/javascript; charset=utf-8"))
	}
	app.GET("/classic", staticHandler("index.html", "text/html; charset=utf-8"))
	app.GET("/index.html", staticHandler("index.html", "text/html; charset=utf-8"))
	app.GET("/app.js", staticHandler("app.js", "text/javascript; charset=utf-8"))
	app.GET("/style.css", staticHandler("style.css", "text/css; charset=utf-8"))
	app.GET("/favicon.svg", staticHandler("favicon.svg", "image/svg+xml"))
	app.GET("/manifest.webmanifest", staticHandler("manifest.webmanifest", "application/manifest+json"))
	app.GET("/sw.js", staticHandler("sw.js", "text/javascript; charset=utf-8"))
	app.GET("/vendor/xterm.js", staticHandler("vendor/xterm.js", "text/javascript; charset=utf-8"))
	app.GET("/vendor/addon-fit.js", staticHandler("vendor/addon-fit.js", "text/javascript; charset=utf-8"))
	app.GET("/vendor/xterm.css", staticHandler("vendor/xterm.css", "text/css; charset=utf-8"))
	app.GET("/vendor/fonts/Geist-Regular.woff2", staticHandler("vendor/fonts/Geist-Regular.woff2", "font/woff2"))
	app.GET("/vendor/fonts/Geist-Medium.woff2", staticHandler("vendor/fonts/Geist-Medium.woff2", "font/woff2"))
	app.GET("/vendor/fonts/Geist-SemiBold.woff2", staticHandler("vendor/fonts/Geist-SemiBold.woff2", "font/woff2"))
	app.GET("/vendor/fonts/Geist-Bold.woff2", staticHandler("vendor/fonts/Geist-Bold.woff2", "font/woff2"))
	app.GET("/vendor/fonts/GeistMono-Regular.woff2", staticHandler("vendor/fonts/GeistMono-Regular.woff2", "font/woff2"))
	app.GET("/vendor/fonts/GeistMono-Medium.woff2", staticHandler("vendor/fonts/GeistMono-Medium.woff2", "font/woff2"))
	app.GET("/board.js", staticHandler("board.js", "text/javascript; charset=utf-8"))
	app.GET("/board.css", staticHandler("board.css", "text/css; charset=utf-8"))
	app.GET("/charts.js", staticHandler("charts.js", "text/javascript; charset=utf-8"))
	app.GET("/charts.css", staticHandler("charts.css", "text/css; charset=utf-8"))
	app.GET("/diffview.js", staticHandler("diffview.js", "text/javascript; charset=utf-8"))
	app.GET("/diffview.css", staticHandler("diffview.css", "text/css; charset=utf-8"))
	app.GET("/vendor/highlight.min.js", staticHandler("vendor/highlight.min.js", "text/javascript; charset=utf-8"))
	app.GET("/vendor/highlight.css", staticHandler("vendor/highlight.css", "text/css; charset=utf-8"))

	// API
	app.GET("/api/state", handleState)
	app.GET("/api/usage", handleUsageBreakdown)
	app.GET("/api/context", handleCodeContext)
	app.POST("/api/reflect", handleReflect)
	app.POST("/api/devserver/stop", handleStopDevServer)
	app.GET("/api/logs", handleLogs)
	app.POST("/api/respond", handleRespond)
	app.POST("/api/spawn", handleSpawn)
	app.POST("/api/kill", handleKill)
	app.POST("/api/send", handleSend)
	app.GET("/api/pane", handlePaneCapture)
	app.GET("/api/config", handleConfigGet)
	app.POST("/api/config", handleConfigPost)
	app.POST("/api/summary", handleSummaryPost)
	app.GET("/api/summary", handleSummaryGet)
	app.DELETE("/api/summary", handleSummaryDelete)
	app.GET("/api/summaries", handleSummaries)
	app.POST("/api/summary/generate", handleSummaryGenerate)
	app.GET("/api/claude-activity", handleClaudeActivity)
	app.GET("/api/repos", handleRepos)
	app.GET("/api/diff", handleDiff)
	app.POST("/api/hook", handleHook)
	app.GET("/api/hooks/events", handleHookEvents)
	app.GET("/api/hooks/status", handleHooksStatus)
	app.POST("/api/hooks/install", handleHooksInstall)
	app.POST("/api/hooks/uninstall", handleHooksUninstall)
	app.POST("/api/review", handleReview)
	app.POST("/api/verify", handleVerify)
	app.POST("/api/chain", handleChainCreate)
	app.GET("/api/chains", handleChains)
	app.POST("/api/chain/advance", handleChainAdvance)
	app.POST("/api/pr/create", handlePRCreate)
	app.POST("/api/pr/merge", handlePRMerge)
	app.POST("/api/webhook/test", handleWebhookTest)
	app.POST("/api/open-editor", handleOpenEditor)
	app.POST("/api/tracker/fetch", handleTrackerFetch)
	app.POST("/api/clone", handleClone)
	app.GET("/api/session-pr", handleSessionPR)
	app.GET("/api/search", handleSearch)
	app.GET("/api/worktrees", handleWorktreesGet)
	app.DELETE("/api/worktrees", handleWorktreesDelete)
	app.GET("/api/audit-history", handleAuditHistory)
	app.GET("/api/github/orgs", handleGHOrgs)
	app.GET("/api/github/repos", handleGHRepos)
	app.GET("/api/github/issues", handleGHIssues)
	app.GET("/api/github/prs", handleGHPRs)
	app.GET("/api/github/pr", handleGHPRDetail)
	app.GET("/api/github/issue", handleGHIssueDetail)

	log.Printf("rook listening on http://%s", *addr)
	log.Printf("watching %s", claudeDir())
	if *token != "" {
		log.Printf("remote access enabled — non-loopback clients need the token")
	}
	app.Run()
}

// portOf extracts the port from an addr like 127.0.0.1:7480.
func portOf(addr string) string {
	if _, p, err := net.SplitHostPort(addr); err == nil && p != "" {
		return p
	}
	return "7480"
}

// staticHandler serves an embedded web asset.
func staticHandler(name, ctype string) gofr.Handler {
	return func(ctx *gofr.Context) (any, error) {
		b, err := webFS.ReadFile("web/" + name)
		if err != nil {
			return nil, errf(http.StatusNotFound, "not found")
		}
		return resp.File{Content: b, ContentType: ctype}, nil
	}
}

// noCache stops the browser caching the embedded HTML/CSS/JS.
func noCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, must-revalidate")
		next.ServeHTTP(w, r)
	})
}

// accessGuard allows loopback clients always. Non-loopback clients (e.g. over
// Tailscale) are allowed only when a token is configured and they present it via
// the X-Rook-Token header or ?token= query — otherwise rejected, so the default
// (no token) stays loopback-only even though GoFr binds all interfaces.
func accessGuard(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}
			host = strings.Trim(host, "[]")
			loopback := host == "127.0.0.1" || host == "::1" || host == "localhost"
			if loopback {
				next.ServeHTTP(w, r)
				return
			}
			if token != "" {
				given := r.Header.Get("X-Rook-Token")
				if given == "" {
					given = r.URL.Query().Get("token")
				}
				if given == token {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, "forbidden: loopback only (set --token for remote access)", http.StatusForbidden)
		})
	}
}

func handleState(ctx *gofr.Context) (any, error) {
	now := time.Now()
	sessions := ScanAllSessions(maxToolsPerSession)
	annotateHealth(sessions, now.UnixMilli())
	return rawJSON(State{
		AppVersion:    appVersion,
		TmuxAvailable: tmuxBin != "",
		Now:           now.UnixMilli(),
		Sessions:      sessions,
		Windows:       TokenWindows(now),
		Trends:        ScanTrends(30),
		DevServers:    DevServers(portFromListenFlag()),
		Env:           ScanEnvironment(),
	})
}

// portFromListenFlag reads the -addr flag's port so rook excludes its own port
// from the dev-server list.
func portFromListenFlag() int {
	f := flag.Lookup("addr")
	if f == nil {
		return 7480
	}
	if p, err := strconv.Atoi(portOf(f.Value.String())); err == nil {
		return p
	}
	return 7480
}

type stopReq struct {
	PID int `json:"pid"`
}

func handleStopDevServer(ctx *gofr.Context) (any, error) {
	var req stopReq
	if err := ctx.Bind(&req); err != nil || req.PID <= 0 {
		return nil, errf(http.StatusBadRequest, "bad request")
	}
	proc, err := os.FindProcess(req.PID)
	if err != nil {
		return nil, errf(http.StatusNotFound, "process not found")
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return nil, errf(http.StatusInternalServerError, "%v", err)
	}
	return rawJSON(map[string]any{"ok": true, "pid": req.PID})
}

type respondReq struct {
	SessionID string `json:"sessionId"`
	Action    string `json:"action"`
	Value     string `json:"value"`
}

// handleRespond sends keystrokes into a session's tmux pane.
func handleRespond(ctx *gofr.Context) (any, error) {
	var req respondReq
	if err := ctx.Bind(&req); err != nil || !validSessionID(req.SessionID) {
		return nil, errf(http.StatusBadRequest, "bad request")
	}
	var pane string
	for _, s := range ScanSessions(0) {
		if s.SessionID == req.SessionID {
			pane = s.TmuxPane
			break
		}
	}
	if pane == "" {
		return nil, errf(http.StatusConflict, "session is not in a tmux pane")
	}
	if code, err := applyKeyAction(pane, req.Action, req.Value); err != nil {
		return nil, errf(code, "%v", err)
	}
	return rawJSON(map[string]any{"ok": true, "pane": pane, "action": req.Action})
}

func isMenuKey(s string) bool {
	return len(s) == 1 && s[0] >= '1' && s[0] <= '9'
}

// validSessionID guards against path traversal — session ids are UUID-shaped.
func validSessionID(id string) bool {
	if id == "" || len(id) > 80 {
		return false
	}
	for _, r := range id {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}
	return true
}

// handleLogs serves a session's raw transcript (bounded to the tail).
func handleLogs(ctx *gofr.Context) (any, error) {
	id := ctx.Param("session")
	if !validSessionID(id) {
		return nil, errf(http.StatusBadRequest, "bad session id")
	}
	path := findTranscript(id)
	if path == "" {
		return nil, errf(http.StatusNotFound, "transcript not found")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errf(http.StatusInternalServerError, "%v", err)
	}
	const maxTail = 2 << 20 // 2 MiB
	if len(data) > maxTail {
		data = data[len(data)-maxTail:]
	}
	return textResp(data, "text/plain; charset=utf-8")
}
