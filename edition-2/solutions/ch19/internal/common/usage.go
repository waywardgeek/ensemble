package common

// Pricing is USD per million tokens in disjoint usage categories. A nil model
// price means unknown, not free; no computed dollars belong in durable history.
type Pricing struct {
	// Input is the ordinary uncached input rate.
	Input float64
	// CacheWrite is the full cache creation rate, not an input surcharge.
	CacheWrite float64
	// CacheRead is the cached input rate.
	CacheRead float64
	// Output includes charged reasoning output where the provider reports it.
	Output float64
}

// UsageMeter is a disposable view of Engine's process-lifetime accounting.
type UsageMeter struct {
	// Usage contains only the latest main conversation response's disjoint counts.
	Usage
	// Total includes every measured request in this process, excluding restored history.
	Total Usage `json:"total"`
	// HitRate is session cached input divided by total session input, never bytes.
	HitRate float64 `json:"hit_rate"`
	// CostUSD is derived from counts and current model prices on each read.
	CostUSD float64 `json:"cost_usd"`
	// Priced is false if any contributing model has unknown pricing.
	Priced bool `json:"priced"`
}
