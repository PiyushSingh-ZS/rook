package main

import (
	"strings"
	"testing"
	"time"
)

func TestSummarizeSession(t *testing.T) {
	p := &parsedTranscript{
		tools: []ToolCall{
			{Name: "Bash"}, {Name: "Bash"}, {Name: "Edit"}, {Name: "WebSearch"},
		},
		changedFiles: []string{"a.go"},
	}
	got := summarizeSession(p)
	want := "4 tool calls · 1 file changed · 2 commands · 1 web lookup"
	if got != want {
		t.Errorf("summarizeSession() = %q, want %q", got, want)
	}
}

func TestSummarizeSession_Empty(t *testing.T) {
	if got := summarizeSession(&parsedTranscript{}); got != "No tool activity yet." {
		t.Errorf("empty summarizeSession() = %q", got)
	}
}

func TestDeriveActivity(t *testing.T) {
	withBash := &parsedTranscript{tools: []ToolCall{{Name: "Bash", Summary: "go build"}}}
	cases := []struct {
		status string
		p      *parsedTranscript
		want   string
	}{
		{"busy", withBash, "Running Bash: go build"},
		{"busy", &parsedTranscript{}, "Thinking…"},
		{"waiting", withBash, "Waiting for your input"},
		{"shell", withBash, "Shell: go build"},
		{"idle", &parsedTranscript{lastText: "All done."}, "All done."},
	}
	for _, c := range cases {
		if got := deriveActivity(c.status, c.p); got != c.want {
			t.Errorf("deriveActivity(%q) = %q, want %q", c.status, got, c.want)
		}
	}
}

func TestSortSessions_AliveFirstThenRecent(t *testing.T) {
	in := []Session{
		{SessionID: "dead-new", Alive: false, UpdatedAt: 100},
		{SessionID: "alive-old", Alive: true, UpdatedAt: 50},
		{SessionID: "alive-new", Alive: true, UpdatedAt: 80},
	}
	sortSessions(in)
	order := []string{in[0].SessionID, in[1].SessionID, in[2].SessionID}
	want := []string{"alive-new", "alive-old", "dead-new"}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("sortSessions order = %v, want %v", order, want)
		}
	}
}

func TestPlural(t *testing.T) {
	if plural(1, "cat", "cats") != "1 cat" {
		t.Error("plural(1) should use singular")
	}
	if plural(3, "cat", "cats") != "3 cats" {
		t.Error("plural(3) should use plural")
	}
}

func TestOneLine_Truncates(t *testing.T) {
	long := strings.Repeat("x", 200)
	got := oneLine(long)
	if len(got) != 120 || !strings.HasSuffix(got, "...") {
		t.Errorf("oneLine long len=%d suffix=%q", len(got), got[len(got)-3:])
	}
	if oneLine("a\nb\nc") != "a b c" {
		t.Error("oneLine should collapse newlines")
	}
}

func TestTokenWindows_Shape(t *testing.T) {
	now := time.Now()
	wins := TokenWindows(now)
	if len(wins) != 2 {
		t.Fatalf("expected 2 windows, got %d", len(wins))
	}
	if wins[0].Label != "5 hours" || wins[1].Label != "7 days" {
		t.Errorf("unexpected labels: %q, %q", wins[0].Label, wins[1].Label)
	}
	for _, w := range wins {
		if w.WindowMs <= 0 {
			t.Errorf("%s: WindowMs should be positive", w.Label)
		}
		if w.OldestMs >= now.UnixMilli() {
			t.Errorf("%s: OldestMs should be before now", w.Label)
		}
	}
}
