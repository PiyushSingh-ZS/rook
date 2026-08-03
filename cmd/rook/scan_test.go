package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDescribeWaiting_PendingBash(t *testing.T) {
	p := &parsedTranscript{hasPending: true, pendingTool: ToolCall{Name: "Bash", Summary: "go test ./..."}}
	if got := describeWaiting(p); got != "Allow Bash to run: go test ./..." {
		t.Fatalf("got %q", got)
	}
}

func TestDescribeWaiting_FullBashCommandNotTruncated(t *testing.T) {
	long := "GOOS=linux GOARCH=amd64 go build -o dist/app ./cmd/server && docker build -t myapp:latest . && docker push registry.example.com/myapp:latest && kubectl rollout restart deployment/app -n production"
	enc, _ := json.Marshal(long)
	dir := t.TempDir()
	f := filepath.Join(dir, "long.jsonl")
	os.WriteFile(f, []byte(
		`{"type":"assistant","timestamp":"2026-07-03T10:00:00Z","message":{"model":"m","content":[{"type":"tool_use","name":"Bash","input":{"command":`+string(enc)+`}}]}}`+"\n"), 0o644)
	p := readTranscript(f)
	if got := describeWaiting(p); !strings.Contains(got, long) {
		t.Fatalf("full command was truncated; got %q", got)
	}
}

func TestDescribeWaiting_NoPendingFallsBackToText(t *testing.T) {
	p := &parsedTranscript{hasPending: false, lastText: "Here is the summary."}
	if got := describeWaiting(p); got != "Here is the summary." {
		t.Fatalf("got %q", got)
	}
}

// A completed tool (result line present) must NOT read as a pending permission.
func TestReadTranscript_PendingClearedByResult(t *testing.T) {
	dir := t.TempDir()

	pending := filepath.Join(dir, "pending.jsonl")
	os.WriteFile(pending, []byte(
		`{"type":"assistant","timestamp":"2026-07-02T10:00:00Z","message":{"model":"m","content":[{"type":"tool_use","name":"Bash","input":{"command":"rm -rf build"}}]}}`+"\n"), 0o644)
	if p := readTranscript(pending); !p.hasPending || p.pendingTool.Name != "Bash" {
		t.Fatalf("expected pending Bash, got hasPending=%v tool=%q", p.hasPending, p.pendingTool.Name)
	} else if got := describeWaiting(p); got != "Allow Bash to run:\nrm -rf build" {
		t.Fatalf("pending phrasing wrong: %q", got)
	}

	completed := filepath.Join(dir, "done.jsonl")
	os.WriteFile(completed, []byte(
		`{"type":"assistant","timestamp":"2026-07-02T10:00:00Z","message":{"model":"m","content":[{"type":"tool_use","name":"Bash","input":{"command":"ls"}}]}}`+"\n"+
			`{"type":"user","timestamp":"2026-07-02T10:00:01Z","toolUseResult":{"stdout":"a b c"}}`+"\n"+
			`{"type":"assistant","timestamp":"2026-07-02T10:00:02Z","message":{"model":"m","content":[{"type":"text","text":"Done listing."}]}}`+"\n"), 0o644)
	if p := readTranscript(completed); p.hasPending {
		t.Fatalf("expected no pending after result line")
	} else if got := describeWaiting(p); got != "Done listing." {
		t.Fatalf("completed fallback wrong: %q", got)
	}
}
