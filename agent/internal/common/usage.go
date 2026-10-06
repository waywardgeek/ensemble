package common

import "sync"

// UsageCounter accumulates token counts for one process lifetime.
//
// Embed it in whatever implements Agent and the two Agent usage methods come
// along for free. Its lifespan is deliberately the lifespan of the struct
// holding it, which is what makes it a *session* figure: one run of the
// program, exactly as a human means it when asking what this session cost.
//
// That distinction matters because the authoritative total kept by the reducer
// is NOT a session figure. Context.Usage is part of the saved context and is
// restored from save.json, so it accumulates across every run the conversation
// has ever had. Both numbers are useful; only one answers "what has this run
// spent", and it is not the durable one.
//
// Counts only. No money is stored here, ever — prices change while counts are
// history, so cost is computed where it is displayed and never recorded.
//
// Counts are kept PER MODEL as well as in aggregate. A session that switches
// models spends tokens at two different prices, and a single session total
// cannot be priced correctly afterwards at either of them: multiplying the
// whole session by the current model's rate retroactively re-prices every
// token already spent, so the reported cost changes when the operator picks a
// different model without sending anything. Keeping the counts split is what
// lets cost stay a derived quantity rather than a remembered one.
type UsageCounter struct {
	mu      sync.Mutex
	usage   Usage
	last    Usage // most recent single-response usage
	byModel map[string]Usage
}

// RecordUsage adds one request's counts to the session total and to the
// running total for the model that served it.
//
// Called from whichever goroutine parsed the response, so it locks. This is
// the write half of the Engine usage seam. The model is a parameter rather
// than something read back from configuration because the configured model
// may already have changed by the time a response is parsed; the only model
// that can be credited is the one the caller actually sent to.
func (c *UsageCounter) RecordUsage(model string, u Usage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.usage.Input += u.Input
	c.usage.CacheWrite += u.CacheWrite
	c.usage.CacheRead += u.CacheRead
	c.usage.Output += u.Output
	c.last = u

	if c.byModel == nil {
		c.byModel = map[string]Usage{}
	}
	m := c.byModel[model]
	m.Input += u.Input
	m.CacheWrite += u.CacheWrite
	m.CacheRead += u.CacheRead
	m.Output += u.Output
	c.byModel[model] = m
}

// UsageByModel returns a copy of the per-model tallies.
//
// A copy, because the caller is going to price it, and pricing walks the map
// while the engine may still be recording into it.
func (c *UsageCounter) UsageByModel() map[string]Usage {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]Usage, len(c.byModel))
	for k, v := range c.byModel {
		out[k] = v
	}
	return out
}

// SessionUsage returns a copy of the totals for this process.
func (c *UsageCounter) SessionUsage() Usage {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.usage
}

// LastUsage returns the token counts from the most recent response.
//
// The GUI displays this rather than the session total so the operator can see
// how large the current prompt is and how much of it was cached. Session
// totals still go to the tooltip and the cost computation.
func (c *UsageCounter) LastUsage() Usage {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.last
}

// CacheHitRate is the fraction of input tokens served from cache.
//
// The denominator is every token fed in — fresh input plus cache writes plus
// cache reads — because those three categories are disjoint and together are
// the whole prompt. Output is excluded: it was generated, not read, and no
// cache could ever have supplied it. Dividing by a denominator that included
// output would report a rate that falls whenever the model simply says more,
// which has nothing to do with caching.
//
// Returns 0 when nothing has been sent, rather than NaN, so a fresh session
// renders as "0.0%" instead of leaking a division by zero into the UI.
func CacheHitRate(u Usage) float64 {
	in := u.Input + u.CacheWrite + u.CacheRead
	if in == 0 {
		return 0
	}
	return float64(u.CacheRead) / float64(in)
}

// UsageSource reports the running token tally for this process. The GUI holds
// one so it can render a session meter. A nil source means no meter, which is
// the right behaviour for the CLI and for graders.
//
// This is deliberately narrower than Agent, which also declares SessionUsage:
// the hub needs to read the tally, never to add to it.
type UsageSource interface {
	SessionUsage() Usage
	LastUsage() Usage
	// UsageByModel returns counts split by the model that incurred them, so
	// that a session spanning several models can be priced correctly.
	UsageByModel() map[string]Usage
}

// CostUSD is the dollar cost of a tally under a price sheet.
//
// The four Usage categories are disjoint and each has its own rate, so this is
// a dot product and nothing more. Rates are per million tokens.
//
// A caller must check Pricing.Priced first. An unpriced sheet returns zero
// here, and zero is a perfectly valid cost for a session that has sent
// nothing, so this function cannot distinguish "free" from "unknown" and does
// not try. That judgment belongs to the renderer, which has somewhere to put
// a dash.
func CostUSD(u Usage, p Pricing) float64 {
	const perMillion = 1000000.0
	return (float64(u.Input)*p.Input +
		float64(u.CacheWrite)*p.CacheWrite +
		float64(u.CacheRead)*p.CacheRead +
		float64(u.Output)*p.Output) / perMillion
}

// CostByModel prices a per-model tally: the sum over models of that model's
// counts at that model's rates.
//
// This is the whole point of keeping counts split. The obvious shortcut —
// multiply the session total by the current model's price — is wrong in a way
// that is easy to miss, because it produces a plausible number that changes
// retroactively. Switch from a cheap model to an expensive one and every
// token the cheap model spent is suddenly billed at the expensive rate, so
// the session cost jumps without a request having been sent.
//
// The second return reports whether every model that actually spent tokens
// had a price. It is false when some did not, which the renderer shows as a
// dash rather than silently under-reporting: a missing price is not a price
// of zero, and the difference matters precisely when a new model has been
// added and nobody has filled in its row yet.
func CostByModel(byModel map[string]Usage) (float64, bool) {
	total := 0.0
	complete := true
	for model, u := range byModel {
		if u == (Usage{}) {
			continue // a model that was selected but never sent to
		}
		f, ok := LookupModel(model)
		if !ok || !f.Price.Priced() {
			complete = false
			continue
		}
		total += CostUSD(u, f.Price)
	}
	return total, complete
}
