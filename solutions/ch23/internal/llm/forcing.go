package llm

import (
	"fmt"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// The two points on the way to a full context window.
//
// Warning first and forcing second is not politeness. A warning lets the
// agent finish the thought it is in the middle of and checkpoint on its own
// terms, which produces a better handoff than one written under duress. The
// second threshold exists because the first one is only a request.
const (
	warnAt  = 0.90
	forceAt = 0.95
)

// handoffTool is the one tool that survives forcing.
const handoffTool = "micro_handoff"

// maybeForce is what makes it safe for compaction to refuse to touch work in
// progress.
//
// Compaction only ever takes checkpointed segments, so an agent that never
// checkpoints would fill its window and hit the wall. This is the part that
// guarantees a checkpoint arrives first. It does it by taking the tools
// away rather than by asking, because asking relies on the agent choosing
// to comply while it is busy, and busy is exactly the state that produced
// the problem.
//
// Taking the tools away rather than writing the checkpoint on the agent's
// behalf matters just as much. The framework can see the conversation, but
// it cannot see what the agent is holding in its head, and a checkpoint is
// mostly the latter. So the framework creates the conditions for a handoff
// and lets the one party who can actually write it do the writing.
func (e *Engine) maybeForce() {
	feat, _ := common.LookupModel(e.Cfg.Model)
	win := feat.ContextWindow
	if win <= 0 {
		// No window in the table means no fraction of it. Forcing on a
		// guess would take every tool from an agent that was doing fine.
		return
	}

	used := e.promptTokens()
	if used == 0 {
		return
	}
	ratio := float64(used) / float64(win)

	restricted := e.restricted()
	switch {
	case ratio >= forceAt:
		if !restricted {
			e.restrict()
		}
	case ratio >= warnAt:
		if !restricted && !e.warned {
			e.warned = true
			e.Record(common.Event{
				Type: common.MessageReceived,
				Message: &common.MessageData{
					Actor: common.ActorSystem,
					Parts: common.PartList{common.TextPart{Text: fmt.Sprintf(
						"Context is %.0f%% full. Call %s soon: at %.0f%% it will be the only "+
							"tool left, and a checkpoint written now can say more than one "+
							"written under duress.",
						ratio*100, handoffTool, forceAt*100)}},
				},
			})
		}
	default:
		// Back under the warning line, which means a checkpoint landed and
		// compaction did its work. Give the tools back and re-arm the
		// warning for next time.
		if restricted {
			e.release()
		}
		e.warned = false
	}
}

// promptTokens is how full the window was on the last request.
//
// The last request rather than a running total, because a running total
// measures how much work has been done, not how much room is left. Prompt
// tokens are input, cache reads and cache writes together: a cached token
// is cheaper, not absent, and it occupies the window exactly like any
// other.
func (e *Engine) promptTokens() int {
	for i := len(e.Log.Events) - 1; i >= 0; i-- {
		ev := e.Log.Events[i]
		if ev.Type == common.ResponseEnded && ev.Response != nil {
			u := ev.Response.Usage
			return u.Input + u.CacheRead + u.CacheWrite
		}
	}
	return 0
}

// restricted reports whether forcing is currently in effect.
//
// Derived from the effective tool set rather than remembered in a field, so
// it is still true after a restart. The log says which tools were taken
// away; nothing has to be written down twice.
func (e *Engine) restricted() bool {
	tools := common.EffectiveTools(e.Cfg.Tools, e.Ctx)
	if len(tools) != 1 {
		return false
	}
	return tools[0].Name == handoffTool
}

func (e *Engine) restrict() {
	var removed []string
	for _, t := range common.EffectiveTools(e.Cfg.Tools, e.Ctx) {
		if t.Name != handoffTool {
			removed = append(removed, t.Name)
		}
	}
	if len(removed) == 0 {
		return
	}
	e.warn("memory: context is full; every tool but %s is withdrawn until a checkpoint lands", handoffTool)
	e.Record(common.Event{
		Type:  common.ToolsChanged,
		Tools: &common.ToolsChangedData{Removed: removed},
	})
}

func (e *Engine) release() {
	var added []common.ToolDecl
	have := map[string]bool{}
	for _, t := range common.EffectiveTools(e.Cfg.Tools, e.Ctx) {
		have[t.Name] = true
	}
	for _, t := range e.Cfg.Tools {
		if !have[t.Name] {
			added = append(added, t)
		}
	}
	if len(added) == 0 {
		return
	}
	e.warn("memory: context is back under control; tools restored")
	e.Record(common.Event{
		Type:  common.ToolsChanged,
		Tools: &common.ToolsChangedData{Added: added},
	})
}
