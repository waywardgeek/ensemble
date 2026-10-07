package tools

// agent_status: what the agent currently costs and what it is running on.
//
// This tool is the payoff for the back-pointer repair, and it is worth being
// explicit about what it did NOT require. No new setting. No closure handed
// down from the wiring site. No new field threaded through four
// constructors. The model name, the price table and the token counts were
// all in scope at the moment the Call was built -- the dispatch site reached
// into the engine, took Jobs, and threw the rest away. Making them reachable
// was one interface and one struct field.
//
// Note the import list: common, and nothing else. The tools package does not
// import internal/llm and must not. A tool asks its parent, and the parent
// answers; it never reaches sideways into the package that happens to
// implement the parent. That restriction is what makes the chain worth
// having, because the moment a tool can import the engine directly there is
// no reason for any of this to exist.

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

func toolAgentStatus(c *common.Call, args json.RawMessage) (string, error) {
	// The engine is absent only in a bare test harness that builds a Call by
	// hand. Say so plainly rather than returning a zeroed report: a status
	// tool that invents zeroes is worse than one that admits it cannot see.
	if c.Engine == nil {
		return "", fmt.Errorf("agent_status: no engine on this call, so there is " +
			"nothing to report on")
	}

	eng := c.Engine
	model := eng.Model()
	usage := eng.Usage()
	last := usage.LastUsage()
	session := usage.SessionUsage()
	byModel := usage.UsageByModel()

	var b strings.Builder
	fmt.Fprintf(&b, "model: %s\n", model)

	fmt.Fprintf(&b, "last response: %d in, %d out, %d cache read, %d cache write\n",
		last.Input, last.Output, last.CacheRead, last.CacheWrite)
	fmt.Fprintf(&b, "session total: %d in, %d out, %d cache read, %d cache write\n",
		session.Input, session.Output, session.CacheRead, session.CacheWrite)

	// Cache hit rate is cached input over all input the model had to be given.
	// Output is excluded because it was never a candidate for being cached, and
	// including it would dilute the figure with tokens the cache could not have
	// helped with.
	if served := session.Input + session.CacheRead; served > 0 {
		fmt.Fprintf(&b, "cache hit rate: %.1f%% (%d of %d input tokens served from cache)\n",
			100*float64(session.CacheRead)/float64(served), session.CacheRead, served)
	} else {
		fmt.Fprintf(&b, "cache hit rate: n/a (no input tokens yet)\n")
	}

	// Per-model lines, so a session that switched models shows its working
	// rather than a single figure the reader has to trust.
	if len(byModel) > 1 {
		for _, name := range sortedModelNames(byModel) {
			u := byModel[name]
			if u == (common.Usage{}) {
				continue
			}
			fmt.Fprintf(&b, "  %s: %d in, %d out, %d cache read\n",
				name, u.Input, u.Output, u.CacheRead)
		}
	}

	cost, priced := common.CostByModel(byModel)
	if priced {
		fmt.Fprintf(&b, "session cost: $%.4f\n", cost)
	} else {
		// A missing price is not a price of zero. Report the shortfall rather
		// than a number that is confidently too low.
		fmt.Fprintf(&b, "session cost: $%.4f or more (at least one model used "+
			"has no price in the model table)\n", cost)
	}

	return strings.TrimRight(b.String(), "\n"), nil
}

// sortedModelNames gives the per-model lines a stable order. Map iteration
// order is randomised in Go, and a status report that shuffles its own rows
// between calls is hard to read and impossible to diff.
//
// Named for models rather than keys because the package already has a
// sortedKeys over a different map type.
func sortedModelNames(m map[string]common.Usage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
