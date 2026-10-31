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
	// Total includes every measured request in this process, including auxiliaries.
	Total Usage `json:"total"`
	// HitRate is session cached input divided by total session input, never bytes.
	HitRate float64 `json:"hit_rate"`
	// CostUSD is derived from counts and current model prices on each read.
	CostUSD float64 `json:"cost_usd"`
	// Priced is false if any contributing model has unknown pricing.
	Priced bool `json:"priced"`
}

// UsageBucket preserves the model and prompt-price tier of measured work. Values
// are copied between Engines after a worker drains; money is derived on read.
type UsageBucket struct {
	// Vendor identifies the provider that incurred the counts.
	Vendor Vendor `json:"vendor"`
	// Model retains the exact configured model, independent of later switches.
	Model string `json:"model"`
	// Large records the request's price tier before aggregation loses its size.
	Large bool `json:"large,omitempty"`
	// Usage contains normalized disjoint counts for this model/tier.
	Usage Usage `json:"usage"`
}
