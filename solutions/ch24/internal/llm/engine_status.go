package llm

import "github.com/waywardgeek/ensemble/agent/internal/common"

// Engine's implementation of common.Engine: the back-pointer interface that
// lets a tool reach the model, the prices and the token counts without
// importing this package.
//
// Every value here was already in scope at the point a Call is constructed.
// None of it needed new plumbing, new settings, or a closure handed down from
// the wiring site. It needed an interface to be reachable through, which is
// the difference Chapter 22 is about.

// Agent returns the engine's parent, completing the chain upward. A tool
// holding a Call reaches the logger as Call.Engine.Agent().Logf, or more
// usually through the Agent embedded in Call directly; both arrive at the
// same object, which is the property that makes one route sufficient.
func (e *Engine) Agent() common.Agent { return e.parent }

// Model is the model the engine will send the next request to.
//
// Read from Cfg rather than cached, because handleSetModel mutates Cfg in
// place instead of rebuilding the engine. Caching it here would produce a
// stale answer after a model switch — which is precisely the class of bug
// that the hub.Model closure shipped.
func (e *Engine) Model() string { return e.Cfg.Model }

// Pricing is the price table for the current model.
//
// An unknown model yields the zero Pricing, whose Priced field is false. That
// is deliberately distinguishable from a model that is genuinely free: a
// missing price must not be allowed to read as a cost of zero.
func (e *Engine) Pricing() common.Pricing {
	f, ok := common.LookupModel(e.Cfg.Model)
	if !ok {
		return common.Pricing{}
	}
	return f.Price
}

// Usage reports token counts, including the per-model split.
//
// The counter is embedded on the Engine rather than on the Agent because the
// engine is the object that spends the tokens and the only one that knows
// which model spent them. Its lifetime is equally suitable: handleSetModel
// mutates Cfg in place, so the engine is never reconstructed and lives
// exactly as long as the process does.
func (e *Engine) Usage() common.UsageSource { return &e.usage }
