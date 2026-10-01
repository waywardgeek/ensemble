package common

import "sync"

// UsageCounter accumulates token counts for one process lifetime.
//
// Embed it in whatever implements Host and the two Host usage methods come
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
type UsageCounter struct {
	mu    sync.Mutex
	usage Usage
	last  Usage // most recent single-response usage
}

// RecordUsage adds one request's counts to the session total.
//
// Called from whichever goroutine parsed the response, so it locks. This is
// the write half of the Host usage seam.
func (c *UsageCounter) RecordUsage(u Usage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.usage.Input += u.Input
	c.usage.CacheWrite += u.CacheWrite
	c.usage.CacheRead += u.CacheRead
	c.usage.Output += u.Output
	c.last = u
}

// SessionUsage returns a copy of the totals for this process.
func (c *UsageCounter) SessionUsage() Usage {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.usage
}

// LastUsage returns the token counts from the most recent response.
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
// This is deliberately narrower than Host, which also declares SessionUsage:
// the hub needs to read the tally, never to add to it.
type UsageSource interface {
	SessionUsage() Usage
	LastUsage() Usage
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
