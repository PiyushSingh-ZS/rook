package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// pushNtfy sends a phone push via ntfy to the configured topic URL
// (e.g. https://ntfy.sh/my-rook-topic). Opt-in — off unless configured.
// priority is an ntfy priority ("urgent"|"high"|""), "" leaving the default.
func pushNtfy(title, body, priority string) {
	url := loadConfig().Ntfy
	if url == "" {
		return
	}
	postNtfy(url, title, body, priority)
}

// postNtfy performs the actual push and reports transport / non-2xx failures
// instead of swallowing them — a revoked topic used to fail silently forever.
func postNtfy(url, title, body, priority string) {
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		log.Printf("ntfy push: bad request: %v", err)
		return
	}
	req.Header.Set("Title", title)
	if priority != "" {
		req.Header.Set("Priority", priority)
	}
	req.Header.Set("Click", localUIURL()) // tapping the push opens the rook console
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ntfy push failed: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("ntfy push: %s returned %s", url, resp.Status)
	}
}

// localUIURL is the address the rook console is served on, for notification
// click-throughs.
func localUIURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d/", portFromListenFlag())
}

// startNotifier watches sessions and fires a quiet native banner notification
// each time one enters "waiting". The persistent, act-until-resolved surface is
// the in-page sticky card; this is just the non-intrusive heads-up.
// stuckAfter is how long an agent may sit "waiting" before rook escalates.
const stuckAfter = 10 * time.Minute

func startNotifier(interval time.Duration) {
	notifiedWait := map[string]bool{}   // waiting sessions already alerted
	lastStatus := map[string]string{}   // previous status per session
	waitingSince := map[string]int64{}  // ms when a session entered waiting
	lastEscalate := map[string]int64{}  // ms of the last stuck re-alert
	first := true
	go func() {
		for {
			time.Sleep(interval)
			current := map[string]bool{}
			nowMs := time.Now().UnixMilli()
			for _, s := range ScanSessions(1) {
				if !s.Alive {
					continue
				}
				prev := lastStatus[s.SessionID]
				lastStatus[s.SessionID] = s.Status

				// waiting: alert once per episode, and keep escalating while stuck
				if s.Status == "waiting" {
					current[s.SessionID] = true
					if _, ok := waitingSince[s.SessionID]; !ok {
						waitingSince[s.SessionID] = nowMs
					}
					if !notifiedWait[s.SessionID] {
						notifiedWait[s.SessionID] = true
						if !first {
							notifyWaiting(s)
						}
					}
					// Re-alert (urgent) every stuckAfter while it stays stuck, not
					// once — a single escalation is easy to miss.
					if !first && nowMs-waitingSince[s.SessionID] >= stuckAfter.Milliseconds() &&
						nowMs-lastEscalate[s.SessionID] >= stuckAfter.Milliseconds() {
						lastEscalate[s.SessionID] = nowMs
						notifyStuck(s, nowMs-waitingSince[s.SessionID])
					}
					continue
				}

				// finished: a session that was working OR waiting on you is now
				// idle. Waiting->idle is a real completion (you answered, it wrapped
				// up) and was previously missed.
				if !first && (prev == "busy" || prev == "waiting") && s.Status == "idle" {
					notifyFinished(s)
				}
			}
			for id := range notifiedWait {
				if !current[id] {
					delete(notifiedWait, id)
					delete(waitingSince, id)
					delete(lastEscalate, id)
				}
			}
			first = false
		}
	}()
}

// notifyStuck escalates an agent that has been waiting too long.
func notifyStuck(s Session, waitedMs int64) {
	mins := waitedMs / 60000
	title := "⚠️ " + firstNonEmpty(s.Project, "agent") + " stuck " + itoa(int(mins)) + "m"
	subtitle := firstNonEmpty(s.Title, s.Project)
	body := clip(firstNonEmpty(s.Asking, s.LastPrompt, "still waiting for your input"), 240)
	banner(title, subtitle, body, "Sosumi")
	pushNtfy(firstNonEmpty(s.Project, "agent")+" waiting "+itoa(int(mins))+"m", body, "urgent")
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func clip(s string, n int) string {
	if len(s) > n {
		return s[:n-1] + "…"
	}
	return s
}

// notifyWaiting fires a non-intrusive native banner: title = project waiting,
// subtitle = what the agent is working on, body = what it's asking to allow.
func notifyWaiting(s Session) {
	title := "⏳ " + firstNonEmpty(s.Project, "agent") + " waiting"
	subtitle := firstNonEmpty(s.Title, s.Project)
	body := clip(firstNonEmpty(s.Asking, s.LastPrompt, "waiting for your input"), 240)
	banner(title, subtitle, body, "Glass")
	pushNtfy(firstNonEmpty(s.Project, "agent")+" needs you", body, "") // phone push (opt-in)
}

// notifyFinished fires a gentle banner when a working session goes idle.
func notifyFinished(s Session) {
	title := "✅ " + firstNonEmpty(s.Project, "agent") + " finished"
	subtitle := firstNonEmpty(s.Title, s.Project)
	body := firstNonEmpty(s.Summary, "session is now idle")
	banner(title, subtitle, body, "Purr")
}

// banner shows a non-intrusive native notification. macOS uses osascript (argv
// passing avoids AppleScript string-escaping issues); Linux uses notify-send if
// present. Other platforms log so nothing is silently lost.
func banner(title, subtitle, body, sound string) {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("osascript",
			"-e", "on run {msg, ttl, sub, snd}",
			"-e", `display notification msg with title ttl subtitle sub sound name snd`,
			"-e", "end run",
			"--", body, title, subtitle, sound)
		if err := cmd.Run(); err != nil {
			log.Printf("notify: %v", err)
		}
	case "linux":
		if _, err := exec.LookPath("notify-send"); err == nil {
			msg := body
			if subtitle != "" {
				msg = subtitle + " — " + body
			}
			if err := exec.Command("notify-send", title, msg).Run(); err != nil {
				log.Printf("notify: %v", err)
			}
			return
		}
		log.Printf("notify: %s — %s: %s", title, subtitle, body)
	default:
		log.Printf("notify: %s — %s: %s", title, subtitle, body)
	}
}
