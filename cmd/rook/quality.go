package main

import "fmt"

// The quality signal is rook's read on HOW WELL an agent did a task, derived
// from process signals it already tracks: whether it looped/thrashed on the same
// tool call, how many reflexion retries it needed to pass the checks, and whether
// the watchdog saw it freeze. It is deliberately NOT a semantic judgment of
// output correctness (rook runs no LLM judge here) — it's a work-quality /
// efficiency score in 0–100. It starts at 100 and only drops when a concrete
// problem is detected, so "100" means "nothing went visibly wrong," not
// "verified correct." The per-factor breakdown makes that explicit in the UI.

// qualityFactor is one line of the breakdown: a check, whether it was clean, and
// how many points it cost.
type qualityFactor struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Penalty int    `json:"penalty"`
	Detail  string `json:"detail"`
}

func capPenalty(p, max int) int {
	if p > max {
		return max
	}
	return p
}

// computeQuality scores a session and returns the factor breakdown behind it.
func computeQuality(s Session) (int, string, []qualityFactor) {
	score := 100
	factors := make([]qualityFactor, 0, 3)

	// 1) looping / thrash: repeated identical tool calls = redoing the same thing.
	loopPen, loopDetail := 0, "no repeated tool calls"
	if n, tc := leadingRepeat(s.ToolCalls); n >= loopRunLen {
		loopPen = capPenalty((n-loopRunLen+1)*10, 30)
		loopDetail = fmt.Sprintf("looping on %s ×%d", toolLabel(tc), n)
	}
	score -= loopPen
	factors = append(factors, qualityFactor{Name: "No looping / thrash", OK: loopPen == 0, Penalty: loopPen, Detail: loopDetail})

	// 2) reflexion retries: it recovered, but needed N tries to pass the gate.
	refPen, refDetail := 0, "passed checks without retries"
	if s.ReflectionAttempts > 0 {
		refPen = capPenalty(s.ReflectionAttempts*10, 30)
		refDetail = fmt.Sprintf("%d reflexion %s to green", s.ReflectionAttempts, plur(s.ReflectionAttempts, "retry", "retries"))
	}
	score -= refPen
	factors = append(factors, qualityFactor{Name: "Passed checks cleanly", OK: refPen == 0, Penalty: refPen, Detail: refDetail})

	// 3) stalls: watchdog "busy but no new output". A "waiting for your input"
	// alert is NOT counted — that's blocked on the human, not poor work.
	stallPen, stallDetail := 0, "kept making progress"
	if s.Health != nil && s.Health.Level == "warn" {
		stallPen = 12
		stallDetail = s.Health.Reason
	}
	score -= stallPen
	factors = append(factors, qualityFactor{Name: "No stalls", OK: stallPen == 0, Penalty: stallPen, Detail: stallDetail})

	if score < 0 {
		score = 0
	}
	return score, qualityLabel(score), factors
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

// annotateQuality attaches the work-quality score + breakdown to each session in
// place. Call AFTER annotateHealth — it reads the health signal.
func annotateQuality(sessions []Session) {
	for i := range sessions {
		sc, label, factors := computeQuality(sessions[i])
		sessions[i].QualityScore = sc
		sessions[i].QualityLabel = label
		sessions[i].QualityFactors = factors
		reasons := []string{}
		for _, f := range factors {
			if !f.OK {
				reasons = append(reasons, f.Detail)
			}
		}
		if len(reasons) == 0 {
			reasons = append(reasons, "no problems detected")
		}
		sessions[i].QualityReasons = reasons
	}
}
