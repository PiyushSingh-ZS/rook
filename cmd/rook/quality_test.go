package main

import "testing"

func qtc(name string) ToolCall { return ToolCall{Name: name, Summary: name} }

func TestComputeQuality(t *testing.T) {
	// healthy: no loops, no retries, no health flag → perfect, all factors OK
	sc, label, factors := computeQuality(Session{})
	if sc != 100 || label != "excellent" {
		t.Fatalf("healthy = %d/%s, want 100/excellent", sc, label)
	}
	if len(factors) != 3 {
		t.Fatalf("expected 3 factors, got %d", len(factors))
	}
	for _, f := range factors {
		if !f.OK || f.Penalty != 0 {
			t.Fatalf("healthy factor %q should be clean, got ok=%v pen=%d", f.Name, f.OK, f.Penalty)
		}
	}

	// looping: 5 identical consecutive tool calls → capped -30, that factor not OK
	loop := []ToolCall{qtc("Bash"), qtc("Bash"), qtc("Bash"), qtc("Bash"), qtc("Bash")}
	sc, label, factors = computeQuality(Session{ToolCalls: loop})
	if sc != 70 || label != "good" {
		t.Fatalf("looping = %d/%s, want 70/good", sc, label)
	}
	if factors[0].OK || factors[0].Penalty != 30 {
		t.Fatalf("looping factor should be -30 not-ok, got ok=%v pen=%d", factors[0].OK, factors[0].Penalty)
	}

	// reflexion retries: 2 → -20
	sc, _, _ = computeQuality(Session{ReflectionAttempts: 2})
	if sc != 80 {
		t.Fatalf("2 retries = %d, want 80", sc)
	}

	// waiting-for-input alert is NOT a quality problem (blocked on the human)
	sc, _, _ = computeQuality(Session{Health: &Health{Level: "alert", Reason: "waiting 12m for your input"}})
	if sc != 100 {
		t.Fatalf("waiting-for-input = %d, want 100 (not penalized)", sc)
	}

	// frozen warn → -12
	sc, _, _ = computeQuality(Session{Health: &Health{Level: "warn", Reason: "no new output for 9m"}})
	if sc != 88 {
		t.Fatalf("frozen = %d, want 88", sc)
	}

	// compounding, clamped, low score → at risk
	sc, label, _ = computeQuality(Session{ToolCalls: loop, ReflectionAttempts: 5, Health: &Health{Level: "warn", Reason: "frozen"}})
	if sc >= 50 || label != "at risk" {
		t.Fatalf("compounded = %d/%s, want low/at risk", sc, label)
	}
}
