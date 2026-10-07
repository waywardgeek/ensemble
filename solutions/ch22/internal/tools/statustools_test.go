package tools

// Tests for agent_status.
//
// The fake engine here is the point of the exercise: it implements
// common.Engine and nothing else, which demonstrates that the tool reaches
// everything it needs through the back-pointer interface and never touches a
// concrete engine type. If this file ever needs to import internal/llm to
// compile, the chain has been bypassed.

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

type fakeAgent struct{}

func (fakeAgent) Logf(string, ...any)    {}
func (fakeAgent) APILogf(string, ...any) {}
func (fakeAgent) Debugf(string, ...any)  {}

type fakeUsage struct {
	session common.Usage
	last    common.Usage
	byModel map[string]common.Usage
}

func (f fakeUsage) SessionUsage() common.Usage            { return f.session }
func (f fakeUsage) LastUsage() common.Usage               { return f.last }
func (f fakeUsage) UsageByModel() map[string]common.Usage { return f.byModel }

type fakeEngine struct {
	model string
	usage fakeUsage
}

func (f *fakeEngine) Agent() common.Agent       { return fakeAgent{} }
func (f *fakeEngine) Model() string             { return f.model }
func (f *fakeEngine) Usage() common.UsageSource { return f.usage }
func (f *fakeEngine) Pricing() common.Pricing {
	if m, ok := common.LookupModel(f.model); ok {
		return m.Price
	}
	return common.Pricing{}
}

func runStatus(t *testing.T, eng common.Engine) string {
	t.Helper()
	out, err := toolAgentStatus(&common.Call{Agent: fakeAgent{}, Engine: eng}, json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("agent_status: %v", err)
	}
	return out
}

func TestAgentStatusReportsAllFiveValues(t *testing.T) {
	eng := &fakeEngine{
		model: "gemini-3.8-flash",
		usage: fakeUsage{
			last:    common.Usage{Input: 100, Output: 20, CacheRead: 300},
			session: common.Usage{Input: 1000, Output: 200, CacheRead: 3000, CacheWrite: 50},
			byModel: map[string]common.Usage{
				"gemini-3.8-flash": {Input: 1000, Output: 200, CacheRead: 3000, CacheWrite: 50},
			},
		},
	}
	got := runStatus(t, eng)

	// Model.
	if !strings.Contains(got, "gemini-3.8-flash") {
		t.Errorf("no model name in report:\n%s", got)
	}
	// Last response usage.
	for _, want := range []string{"100 in", "20 out", "300 cache read"} {
		if !strings.Contains(got, want) {
			t.Errorf("last-response figure %q missing:\n%s", want, got)
		}
	}
	// Session totals.
	for _, want := range []string{"1000 in", "200 out", "3000 cache read"} {
		if !strings.Contains(got, want) {
			t.Errorf("session figure %q missing:\n%s", want, got)
		}
	}
	// Cache hit rate: 3000 cached of 4000 input served = 75.0%.
	if !strings.Contains(got, "75.0%") {
		t.Errorf("cache hit rate wrong or missing, want 75.0%%:\n%s", got)
	}
	// Cost, against the real table rather than a number copied from a run.
	f, ok := common.LookupModel("gemini-3.8-flash")
	if !ok {
		t.Fatal("gemini-3.8-flash missing from the model table")
	}
	want, _ := common.CostByModel(eng.usage.byModel)
	if want != common.CostUSD(eng.usage.session, f.Price) {
		t.Fatal("fixture inconsistent: single-model cost should equal the session cost")
	}
	if !strings.Contains(got, "session cost: $") {
		t.Errorf("no session cost in report:\n%s", got)
	}
}

// The chapter's bug, asserted at the tool's output. Two models at very
// different prices: the report must show the sum of each model's own cost,
// not the whole session priced at the current model's rate.
func TestAgentStatusPricesPerModelNotAtCurrentRate(t *testing.T) {
	cheap := common.Usage{Input: 1_000_000, Output: 200_000}
	dear := common.Usage{Input: 10_000, Output: 2_000}
	byModel := map[string]common.Usage{"gemini-3.8-flash": cheap, "gpt-6-astra": dear}

	eng := &fakeEngine{
		model: "gpt-6-astra", // the expensive one is selected NOW
		usage: fakeUsage{
			last: dear,
			session: common.Usage{
				Input:  cheap.Input + dear.Input,
				Output: cheap.Output + dear.Output,
			},
			byModel: byModel,
		},
	}
	got := runStatus(t, eng)

	cheapF, _ := common.LookupModel("gemini-3.8-flash")
	dearF, ok := common.LookupModel("gpt-6-astra")
	if !ok {
		t.Fatal("gpt-6-astra missing from the model table")
	}
	correct := common.CostUSD(cheap, cheapF.Price) + common.CostUSD(dear, dearF.Price)
	buggy := common.CostUSD(eng.usage.session, dearF.Price)

	if sameMoney(correct, buggy) {
		t.Fatal("fixture is useless: both pricing methods agree, so this test " +
			"could not catch the bug it exists for")
	}
	if !strings.Contains(got, money(correct)) {
		t.Errorf("report does not show the per-model cost %s:\n%s", money(correct), got)
	}
	if strings.Contains(got, money(buggy)) {
		t.Errorf("report shows the session-total-times-current-price figure %s:\n%s",
			money(buggy), got)
	}
	// Both models should appear, so the reader can check the arithmetic.
	for _, m := range []string{"gemini-3.8-flash", "gpt-6-astra"} {
		if !strings.Contains(got, m) {
			t.Errorf("per-model breakdown missing %s:\n%s", m, got)
		}
	}
}

// An unpriced model must not be billed at zero.
func TestAgentStatusFlagsUnpricedModels(t *testing.T) {
	if f, ok := common.LookupModel("gpt-ch19-course"); !ok {
		t.Fatal("gpt-ch19-course missing from the model table")
	} else if f.Price.Priced() {
		t.Skip("gpt-ch19-course has gained a price; pick another unpriced model")
	}
	eng := &fakeEngine{
		model: "gpt-ch19-course",
		usage: fakeUsage{
			session: common.Usage{Input: 1000, Output: 100},
			byModel: map[string]common.Usage{"gpt-ch19-course": {Input: 1000, Output: 100}},
		},
	}
	got := runStatus(t, eng)
	if !strings.Contains(got, "or more") {
		t.Errorf("an unpriced model must be reported as a lower bound:\n%s", got)
	}
}

// A Call built by hand with no engine must say so rather than report zeroes.
func TestAgentStatusRefusesWithoutAnEngine(t *testing.T) {
	_, err := toolAgentStatus(&common.Call{Agent: fakeAgent{}}, json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected an error when no engine is reachable")
	}
}

// money renders a figure exactly as the tool does, so comparisons are made
// against the string the operator actually sees rather than against a float.
func money(v float64) string { return fmt.Sprintf("$%.4f", v) }

// sameMoney asks whether two costs are indistinguishable once rendered. A
// fixture whose two pricing methods round to the same string could not
// demonstrate the bug, so the test checks that first.
func sameMoney(a, b float64) bool { return money(a) == money(b) }
