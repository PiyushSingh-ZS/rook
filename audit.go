package main

import (
	"strconv"
	"strings"
	"time"

	"gofr.dev/pkg/gofr"
)

// AuditCmd is one persisted command run by an agent.
type AuditCmd struct {
	Session  string `json:"session"`
	Project  string `json:"project"`
	Provider string `json:"provider"`
	Cmd      string `json:"cmd"`
	Ts       int64  `json:"ts"`
}

// startAuditIngester periodically records Bash/Shell commands from all sessions
// into SQLite so the audit trail survives restarts and grows beyond the live
// transcript window.
func startAuditIngester(interval time.Duration) {
	if db == nil {
		return
	}
	go func() {
		for {
			ingestAudit()
			time.Sleep(interval)
		}
	}()
}

func ingestAudit() {
	if db == nil {
		return
	}
	for _, s := range ScanAllSessions(maxToolsPerSession) {
		prov := s.Provider
		if prov == "" {
			prov = "claude"
		}
		for _, t := range s.ToolCalls {
			if t.Name != "Bash" && t.Name != "Shell" {
				continue
			}
			cmd := t.Summary
			key := s.SessionID + "|" + strconv.FormatInt(t.Timestamp, 10) + "|" + firstN(cmd, 80)
			_, _ = db.Exec(
				`INSERT OR IGNORE INTO audit_cmds (dedup, session, project, provider, cmd, ts) VALUES (?, ?, ?, ?, ?, ?)`,
				key, s.SessionID, s.Project, prov, cmd, t.Timestamp,
			)
		}
	}
}

func firstN(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// handleAuditHistory returns persisted commands, newest first, optionally
// filtered by a substring (?q=).
func handleAuditHistory(ctx *gofr.Context) (any, error) {
	if db == nil {
		return rawJSON([]AuditCmd{})
	}
	q := strings.TrimSpace(ctx.Param("q"))
	var rows interface {
		Next() bool
		Scan(...any) error
		Close() error
		Err() error
	}
	var err error
	if q != "" {
		rows, err = db.Query(`SELECT session, project, provider, cmd, ts FROM audit_cmds
			WHERE cmd LIKE ? ORDER BY ts DESC LIMIT 500`, "%"+q+"%")
	} else {
		rows, err = db.Query(`SELECT session, project, provider, cmd, ts FROM audit_cmds
			ORDER BY ts DESC LIMIT 500`)
	}
	if err != nil {
		return nil, errf(500, "%v", err)
	}
	defer rows.Close()
	out := []AuditCmd{}
	for rows.Next() {
		var a AuditCmd
		if err := rows.Scan(&a.Session, &a.Project, &a.Provider, &a.Cmd, &a.Ts); err != nil {
			return nil, errf(500, "%v", err)
		}
		out = append(out, a)
	}
	return rawJSON(out)
}
