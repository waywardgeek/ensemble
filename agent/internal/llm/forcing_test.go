package llm

import (
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// ch16Window is the window of the grader model, small enough that a test can
// reach ninety percent of it without generating a hundred thousand tokens.
const ch16Window = 4096

func forcingEngine(t *testing.T, model string) *Engine {
	t.Helper()
	return &Engine{
		Log:  common.NewLog(),
		Ctx:  common.NewContext(),
		Host: testHost{t},
		Cfg: common.Config{
			Vendor: common.VendorAnthropic,
			Model:  model,
			Tools: []common.ToolDecl{
				{Name: "read_file"}, {Name: "run_command"}, {Name: handoffTool},
			},
		},
	}
}

// used records a response that consumed n prompt tokens.
func used(e *Engine, n int) {
	e.Record(common.Event{
		Type: common.ResponseEnded,
		Response: &common.ResponseData{
			Parts: common.PartList{common.TextPart{Text: "ok"}},
			Usage: common.Usage{Input: n / 2, CacheRead: n - n/2},
		},
	})
}

func toolNames(e *Engine) []string {
	var out []string
	for _, t := range common.EffectiveTools(e.Cfg.Tools, e.Ctx) {
		out = append(out, t.Name)
	}
	return out
}

// systemSays reads the ephemera channel, which is where a System message
// lands.
//
// That is the right channel for this and not a workaround. A warning about
// context pressure is volatile status: true now, meaningless in ten turns,
// and delivered exactly once. Putting it in the dialogue would make the
// warning about a full context into a permanent resident of that context.
func systemSays(e *Engine) string {
	var b strings.Builder
	for _, p := range e.Ctx.Ephemera {
		if tp, ok := p.(common.TextPart); ok {
			b.WriteString(tp.Text)
		}
	}
	return b.String()
}

// TestForcingWarnsBeforeItTakesAnything checks that the agent is asked
// before it is compelled, and asked only once.
func TestForcingWarnsBeforeItTakesAnything(t *testing.T) {
	e := forcingEngine(t, "claude-ch16-course")
	used(e, 3727) // 91% of 4096

	e.maybeForce()
	if got := len(toolNames(e)); got != 3 {
		t.Fatalf("a warning took tools away: %v", toolNames(e))
	}
	if !strings.Contains(systemSays(e), handoffTool) {
		t.Fatalf("no warning mentioning %s: %q", handoffTool, systemSays(e))
	}

	first := systemSays(e)
	e.maybeForce()
	if systemSays(e) != first {
		t.Fatal("the warning repeated on the very next turn")
	}
}

// TestForcingLeavesOnlyTheCheckpointTool is the compelling step. Asking is
// what the previous threshold was for; this one does not ask.
func TestForcingLeavesOnlyTheCheckpointTool(t *testing.T) {
	e := forcingEngine(t, "claude-ch16-course")
	used(e, 3932) // 96% of 4096

	e.maybeForce()

	got := toolNames(e)
	if len(got) != 1 || got[0] != handoffTool {
		t.Fatalf("want only %s, got %v", handoffTool, got)
	}
}

// TestForcingGivesTheToolsBack closes the loop: once a checkpoint has landed
// and compaction has run, the agent gets its toolbox back.
func TestForcingGivesTheToolsBack(t *testing.T) {
	e := forcingEngine(t, "claude-ch16-course")
	used(e, 3932) // 96% of 4096
	e.maybeForce()
	if len(toolNames(e)) != 1 {
		t.Fatalf("setup: forcing did not take hold: %v", toolNames(e))
	}

	used(e, ch16Window/8) // the context is roomy again
	e.maybeForce()

	if got := len(toolNames(e)); got != 3 {
		t.Fatalf("tools were not restored: %v", toolNames(e))
	}
}

// TestForcingDoesNothingWithoutAWindow is the degradation rule. An unknown
// window is an admission, not a reason to guess, and guessing low would
// take every tool from an agent that was doing fine.
func TestForcingDoesNothingWithoutAWindow(t *testing.T) {
	e := forcingEngine(t, "a-model-nobody-has-heard-of")
	used(e, 10_000_000)

	e.maybeForce()

	if got := len(toolNames(e)); got != 3 {
		t.Fatalf("forced on a model with no known window: %v", got)
	}
	if s := systemSays(e); s != "" {
		t.Fatalf("warned on a model with no known window: %q", s)
	}
}
