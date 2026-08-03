package main

import "fmt"

// The quality signal is rook's read on HOW WELL an agent did a task, derived
// from process signals it already tracks: whether it looped/thrashed on the same
// tool call, how many reflexion retries it needed to pass the checks, and whether
// the watchdog saw it freeze. It is deliberately NOT a semantic judgment of
// output correctness (rook runs no LLM judge here) — it's a work-quality /
// efficiency score in 0–100, shown per task next to cost.

// computeQuality scores a session and lists what moved the score.
func computeQuality(s Session) (int, string, []string) {
	score := 100
	var reasons []string

	// looping / thrash: repeated identical tool calls = redoing the same thing.
	if n, tc := leadingRepeat(s.ToolCalls); n >= loopRunLen {
		pen := (n - loopRunLen + 1) * 10
		if pen > 30 {
			pen = 30
		}
		score -= pen
		reasons = append(reasons, fmt.Sprintf("looping on %s ×%d", toolLabel(tc), n))
	}

	// reflexion retries: it recovered, but needed N tries to pass the gate.
	if s.ReflectionAttempts > 0 {
		pen := s.ReflectionAttempts * 10
		if pen > 30 {
			pen = 30
		}
		score -= pen
		reasons = append(reasons, fmt.Sprintf("%d reflexion %s to green", s.ReflectionAttempts, plur(s.ReflectionAttempts, "retry", "retries")))
	}

	// frozen: the watchdog flagged "busy but no new output" (possible stall). A
	// "waiting for your input" alert is NOT counted — that's blocked on the human,
	// not poor work.
	if s.Health != nil && s.Health.Level == "warn" {
		score -= 12
		reasons = append(reasons, s.Health.Reason)
	}

	if score < 0 {
		score = 0
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "no problems detected")
	}
	return score, qualityLabel(score), reasons
}

func qualityLabel(score int) string {
	switch {
	case score >= 85:
		return "excellent"
	case score >= 70:
		return "good"
	case score >= 50:
		return "fair"
	default:
		return "at risk"
	}
}

func plur(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// annotateQuality attaches the work-quality score to each session in place. Call
// AFTER annotateHealth — it reads the health signal.
func annotateQuality(sessions []Session) {
	for i := range sessions {
		sc, label, reasons := computeQuality(sessions[i])
		sessions[i].QualityScore = sc
		sessions[i].QualityLabel = label
		sessions[i].QualityReasons = reasons
	}
}
