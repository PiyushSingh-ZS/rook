package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// pushNtfy sends a phone push via ntfy if FOREMAN_NTFY is set to a topic URL
// (e.g. https://ntfy.sh/my-foreman-topic). Opt-in — off unless configured.
func pushNtfy(title, body string) {
	url := loadConfig().Ntfy
	if url == "" {
		url = os.Getenv("FOREMAN_NTFY")
	}
	if url == "" {
		return
	}
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Title", title)
	req.Header.Set("Tags", "construction_worker")
	client := &http.Client{Timeout: 8 * time.Second}
	if resp, err := client.Do(req); err == nil {
		_ = resp.Body.Close()
	}
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
	escalated := map[string]bool{}      // stuck-escalation already fired
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

				// waiting: alert once per episode, and escalate if stuck
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
					if !first && !escalated[s.SessionID] &&
						nowMs-waitingSince[s.SessionID] >= stuckAfter.Milliseconds() {
						escalated[s.SessionID] = true
						notifyStuck(s, nowMs-waitingSince[s.SessionID])
					}
					continue
				}

				// finished: a session that was working is now idle
				if !first && prev == "busy" && s.Status == "idle" {
					notifyFinished(s)
				}
			}
			for id := range notifiedWait {
				if !current[id] {
					delete(notifiedWait, id)
					delete(waitingSince, id)
					delete(escalated, id)
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
	pushNtfy(firstNonEmpty(s.Project, "agent")+" waiting "+itoa(int(mins))+"m", body)
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
	pushNtfy(firstNonEmpty(s.Project, "agent")+" needs you", body) // phone push (opt-in)
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
