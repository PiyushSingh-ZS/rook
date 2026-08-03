package main

import (
	"fmt"
	"sync"
)

// The quality signal is rook's read on how a task went, built from signals it
// tracks locally — no external eval service. Following how the field weights
// things (SWE-bench % resolved, LangSmith/DeepEval tool-correctness + task
// success), the backbone is OUTCOME (did build/tests pass) and TOOL RELIABILITY
// (tool-call error rate); looping/retries/stalls are minor efficiency factors.
// It is NOT a correctness judgment (that needs an LLM judge) — a high score with
// "no build/test gate run" means "nothing went visibly wrong," not "verified".

// verifyStore remembers the last verify (build/test) outcome per worktree dir,
// recorded by runVerify. Absent = never run for that dir.
var (
	verifyMu    sync.Mutex
	verifyStore = map[string]bool{}
)

func recordVerify(dir string, res verifyResult) {
	if dir == "" || !res.Ran {
		return
	}
	verifyMu.Lock()
	verifyStore[dir] = res.OK
	verifyMu.Unlock()
}

// verifyOutcomeFor returns "pass" | "fail" | "" (not run) for a worktree dir.
func verifyOutcomeFor(dir string) string {
	if dir == "" {
		return ""
	}
	verifyMu.Lock()
	defer verifyMu.Unlock()
	v, ok := verifyStore[dir]
	if !ok {
		return ""
	}
	if v {
		return "pass"
	}
	return "fail"
}

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

// computeQuality scores a session given its build/test outcome ("pass"|"fail"|"")
// and returns the factor breakdown behind it.
func computeQuality(s Session, verify string) (int, string, []qualityFactor) {
	score := 100
	factors := make([]qualityFactor, 0, 5)

	// 1) OUTCOME — build/tests (the dominant signal, like SWE-bench % resolved).
	switch verify {
	case "fail":
		score -= 45
		factors = append(factors, qualityFactor{Name: "Build & tests", OK: false, Penalty: 45, Detail: "build/tests failing"})
	case "pass":
		factors = append(factors, qualityFactor{Name: "Build & tests", OK: true, Detail: "build/tests passing"})
	default:
		// neutral: not penalized, but flagged so a 100 isn't mistaken for "verified".
		factors = append(factors, qualityFactor{Name: "Build & tests", OK: true, Detail: "no build/test gate run — enable Auto-verify for a real pass/fail"})
	}

	// 2) TOOL RELIABILITY — tool-call error rate (deterministic, from transcript).
	if s.ToolResults > 0 {
		rate := float64(s.ToolErrors) / float64(s.ToolResults)
		pen := capPenalty(int(rate*35+0.5), 30)
		score -= pen
		factors = append(factors, qualityFactor{Name: "Tool reliability", OK: pen == 0,
			Penalty: pen, Detail: fmt.Sprintf("%d of %d tool calls errored", s.ToolErrors, s.ToolResults)})
	} else {
		factors = append(factors, qualityFactor{Name: "Tool reliability", OK: true, Detail: "no tool errors"})
	}

	// 3) No looping / thrash (minor efficiency).
	loopPen, loopDetail := 0, "no repeated tool calls"
	if n, tc := leadingRepeat(s.ToolCalls); n >= loopRunLen {
		loopPen = capPenalty((n-loopRunLen+1)*6, 15)
		loopDetail = fmt.Sprintf("looping on %s ×%d", toolLabel(tc), n)
	}
	score -= loopPen
	factors = append(factors, qualityFactor{Name: "No looping", OK: loopPen == 0, Penalty: loopPen, Detail: loopDetail})

	// 4) Recovered without retries (minor) — reflexion retries to pass the gate.
	refPen, refDetail := 0, "passed checks without retries"
	if s.ReflectionAttempts > 0 {
		refPen = capPenalty(s.ReflectionAttempts*6, 15)
		refDetail = fmt.Sprintf("%d reflexion %s to green", s.ReflectionAttempts, plur(s.ReflectionAttempts, "retry", "retries"))
	}
	score -= refPen
	factors = append(factors, qualityFactor{Name: "Recovered cleanly", OK: refPen == 0, Penalty: refPen, Detail: refDetail})

	// 5) No stalls (minor) — watchdog "busy but no output". Waiting-on-human isn't counted.
	stallPen, stallDetail := 0, "kept making progress"
	if s.Health != nil && s.Health.Level == "warn" {
		stallPen = 8
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

// annotateQuality attaches the quality score + breakdown to each session in
// place. Call AFTER annotateHealth — it reads the health signal.
func annotateQuality(sessions []Session) {
	for i := range sessions {
		sc, label, factors := computeQuality(sessions[i], verifyOutcomeFor(sessions[i].CWD))
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
