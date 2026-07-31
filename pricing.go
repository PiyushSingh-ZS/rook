package main

import "strings"

// rate is USD per token for each usage category.
type rate struct {
	in, out, cacheRead, cacheWrite float64
}

// pricePerToken maps a model id to its USD-per-token rates. Values are the
// public Claude API list prices (USD per million tokens, divided to per-token).
// Matched by family substring so version bumps keep working.
func pricePerToken(model string) rate {
	m := strings.ToLower(model)
	switch {
	case strings.Contains(m, "opus"):
		return rate{in: 15e-6, out: 75e-6, cacheRead: 1.5e-6, cacheWrite: 18.75e-6}
	case strings.Contains(m, "haiku"):
		return rate{in: 0.8e-6, out: 4e-6, cacheRead: 0.08e-6, cacheWrite: 1e-6}
	case strings.Contains(m, "sonnet"):
		return rate{in: 3e-6, out: 15e-6, cacheRead: 0.3e-6, cacheWrite: 3.75e-6}
	default:
		// unknown model — assume Sonnet-class so estimates stay sane
		return rate{in: 3e-6, out: 15e-6, cacheRead: 0.3e-6, cacheWrite: 3.75e-6}
	}
}

func (r rate) cost(ev usageEvent) float64 {
	return r.in*float64(ev.input) +
		r.out*float64(ev.output) +
		r.cacheRead*float64(ev.cacheRead) +
		r.cacheWrite*float64(ev.cacheWrite)
}
