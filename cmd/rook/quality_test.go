package main

import "testing"

func qtc(name string) ToolCall { return ToolCall{Name: name, Summary: name} }

func factorByName(fs []qualityFactor, name string) *qualityFactor {
	for i := range fs {
		if fs[i].Name == name {
			return &fs[i]
		}
	}
	return nil
}

func TestComputeQuality(t *testing.T) {
	// clean run, tests passing → perfect, outcome factor OK
	sc, label, factors := computeQuality(Session{ToolResults: 20}, "pass")
	if sc != 100 || label != "excellent" {
		t.Fatalf("clean+pass = %d/%s, want 100/excellent", sc, label)
	}
	if f := factorByName(factors, "Build & tests"); f == nil || !f.OK || f.Detail != "build/tests passing" {
		t.Fatalf("expected passing build factor, got %+v", f)
	}

	// failing build is the dominant hit — drops out of "excellent" even if clean otherwise
	sc, label, factors = computeQuality(Session{ToolResults: 20}, "fail")
	if sc != 55 || label != "fair" {
		t.Fatalf("fail = %d/%s, want 55/fair", sc, label)
	}
	if f := factorByName(factors, "Build & tests"); f == nil || f.OK || f.Penalty != 45 {
		t.Fatalf("expected -45 failing build factor, got %+v", f)
	}

	// tool-call error rate penalizes proportionally
	sc, _, factors = computeQuality(Session{ToolResults: 20, ToolErrors: 10}, "")
	tr := factorByName(factors, "Tool reliability")
	if tr == nil || tr.OK || tr.Penalty == 0 {
		t.Fatalf("expected tool-error penalty, got %+v", tr)
	}
	if sc != 100-tr.Penalty {
		t.Fatalf("score %d should be 100-%d", sc, tr.Penalty)
	}

	// "no gate run" is neutral (not penalized) but flagged in the breakdown
	sc, _, factors = computeQuality(Session{ToolResults: 5}, "")
	if sc != 100 {
		t.Fatalf("no-gate clean run = %d, want 100 (neutral)", sc)
	}
	if f := factorByName(factors, "Build & tests"); f == nil || !f.OK || f.Detail == "build/tests passing" {
		t.Fatalf("no-gate build factor should be neutral+flagged, got %+v", f)
	}

	// process issues are now MINOR relative to outcome (looping capped at 15)
	loop := []ToolCall{qtc("Bash"), qtc("Bash"), qtc("Bash"), qtc("Bash"), qtc("Bash")}
	_, _, factors = computeQuality(Session{ToolResults: 10, ToolCalls: loop}, "pass")
	lf := factorByName(factors, "No looping")
	if lf == nil || lf.OK || lf.Penalty > 15 {
		t.Fatalf("looping should be a minor (<=15) penalty, got %+v", lf)
	}

	// failing build + high tool errors + looping → at risk
	sc, label, _ = computeQuality(Session{ToolResults: 10, ToolErrors: 8, ToolCalls: loop}, "fail")
	if sc >= 50 || label != "at risk" {
		t.Fatalf("compounded bad run = %d/%s, want low/at risk", sc, label)
	}
}

func TestCountToolResults(t *testing.T) {
	// two results, one errored
	content := []byte(`[{"type":"tool_result","is_error":false,"content":"ok"},{"type":"tool_result","is_error":true,"content":"boom"}]`)
	total, errs := countToolResults(content)
	if total != 2 || errs != 1 {
		t.Fatalf("countToolResults = %d/%d, want 2/1", total, errs)
	}
	// tool_use blocks (assistant) are not results
	assistant := []byte(`[{"type":"text"},{"type":"tool_use","name":"Bash"}]`)
	if tt, _ := countToolResults(assistant); tt != 0 {
		t.Fatalf("tool_use should not count as results, got %d", tt)
	}
}
