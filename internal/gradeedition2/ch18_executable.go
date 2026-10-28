package gradeedition2

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter18Executable keeps every web and status behavioral weight. Compilation
// and Go's import graph are not inferred from an executable grade. Skills is the
// shipped directory, not a replacement primary manufactured by this adapter.
func Chapter18Executable(cli, gui, skills string) (grade.Report, error) {
	checks := webChecksFor(cli, true, skills)
	checks[1].Title = "HTTP MCP activation reaches the local server and publishes only granted tools"
	checks = append(checks, statusExecutableChecks(gui, skills)...)
	return grade.NewTitledReport("Edition 3 C#, Chapter 18 — executable web and status scope (135 points; Go build/import checks excluded)", checks,
		"Owned local peers only; fixture paths are printed to stderr. No hosted web, live model availability or billing claim."), nil
}

func selectStatusExecutableModel(peer *guiPeer, model string) error {
	if err := invoiceSend(peer, "set_model", model, true); err != nil {
		return err
	}
	var refusal error
	_, err := peer.until(3*time.Second, func(m map[string]any) bool {
		if m["type"] == "error" {
			refusal = fmt.Errorf("model control refused: %v", m)
			return true
		}
		return m["type"] == "usage" && m["agent_id"] == "root" && m["selected_model"] == model
	})
	if err != nil {
		return err
	}
	return refusal
}

func statusExecutableReply(model string, i int, usage fakevendor.Canonical) fakevendor.Reply {
	var body map[string]any
	// Reuse the documented OpenAI usage encoder, not an invented cache-write field.
	json.Unmarshal([]byte(invoiceExecutableOpenAI(model, "default", "", usage)), &body)
	body["choices"] = []any{map[string]any{"index": 0, "finish_reason": "tool_calls", "message": map[string]any{
		"role": "assistant", "content": "I will inspect the current status.", "tool_calls": []any{map[string]any{
			"id": fmt.Sprintf("status-%d", i), "type": "function", "function": map[string]any{"name": "agent_status", "arguments": "{}"},
		}},
	}}}
	data, _ := json.Marshal(body)
	return fakevendor.Reply{JSONBody: string(data)}
}

// Exact tool_call_id correlation matters: old status text elsewhere in dialogue
// cannot substitute for the tool result being assessed.
func statusExecutableResult(c *grade.Check, r fakevendor.Recorded, id string) string {
	var found []string
	for _, raw := range list(r.JSON()["messages"]) {
		m := object(raw)
		if m["role"] == "tool" && m["tool_call_id"] == id {
			text, ok := m["content"].(string)
			if !ok || strings.TrimSpace(text) == "" {
				fail(c, "empty/nontext result %s", id)
			}
			found = append(found, text)
		}
	}
	if len(found) != 1 {
		fail(c, "result %s has %d correlated entries", id, len(found))
		return ""
	}
	return found[0]
}

func statusExecutableChecks(bin, skills string) []grade.Check {
	checks := []grade.Check{
		{ID: "agent-status-tool", Title: "Shipped status tool: current model, last/session counts including cache writes, hit rate", Points: 20, Earned: 20, Passed: true},
		{ID: "per-model-cost", Title: "Two priced models, unknown work and return without repricing or resetting", Points: 15, Earned: 15, Passed: true},
	}
	status, cost := &checks[0], &checks[1]
	models := []string{"gpt-4o-mini", "gpt-4o", "course-unpriced", "gpt-4o-mini"}
	usages := []fakevendor.Canonical{
		{Input: 10000, Output: 1000, CacheRead: 2000}, {Input: 2000, Output: 200, CacheRead: 3000},
		{Input: 1000, Output: 100, CacheRead: 500}, {Input: 500, Output: 50, CacheRead: 250},
	}
	// USD per million, exact supported OpenAI/default rows. A missing row remains
	// unknown; selecting a known row later cannot retroactively price that work.
	rates := [][4]float64{{.15, 0, .075, .60}, {2.50, 0, 1.25, 10}, {}, {.15, 0, .075, .60}}
	out := statusLaunchFor(bin, skills, models, true, true, "openai", usages)
	if out.fatal != "" {
		fail(status, "status session: %s", out.fatal)
		fail(cost, "status session: %s", out.fatal)
		return checks
	}
	var total fakevendor.Canonical
	dollars := 0.0
	for i, u := range usages {
		previous := total
		total.Input += u.Input
		total.Output += u.Output
		total.CacheRead += u.CacheRead
		total.CacheWrite += u.CacheWrite
		dollars += (float64(u.Input)*rates[i][0] + float64(u.CacheRead)*rates[i][2] + float64(u.Output)*rates[i][3]) / 1e6
		for _, r := range out.reqs[2*i : 2*i+2] {
			if r.JSON()["model"] != models[i] {
				fail(status, "turn %d model did not reach wire", i)
				fail(cost, "wrong producing model")
			}
		}
		declared := false
		for _, raw := range list(out.reqs[2*i].JSON()["tools"]) {
			if object(object(raw)["function"])["name"] == "agent_status" {
				declared = true
			}
		}
		if !declared {
			fail(status, "shipped primary does not declare agent_status")
		}
		text := statusExecutableResult(status, out.reqs[2*i+1], fmt.Sprintf("status-%d", i))
		statusReport(status, text, models[i], u, total)
		statusCost(cost, text, dollars, i < 2)
		if !invoiceCounts(out.after[i], "total", total) {
			fail(status, "GUI session counts disagree after turn %d", i)
		}
		if i < 2 {
			if !invoiceNear(out.after[i]["total_cost_usd"], dollars) {
				fail(cost, "GUI cost mismatch after turn %d", i)
			}
		} else if value, present := out.after[i]["total_cost_usd"]; !present || value != nil {
			fail(cost, "unpriced work must keep explicit null GUI cost")
		}
		if out.before[i]["selected_model"] != models[i] || !invoiceCounts(out.before[i], "total", previous) {
			fail(cost, "selection changed previous counts or failed projection")
		}
		if i > 0 {
			if out.before[i]["total_cost_usd"] != out.after[i-1]["total_cost_usd"] {
				fail(cost, "selection repriced prior work")
			}
			if !strings.Contains(string(out.reqs[2*i].Body), "Report status turn 0.") {
				fail(status, "selection discarded dialogue")
			}
		}
	}
	// Anthropic has a documented cache-write bucket. Keep this separately unpriced
	// rather than inventing an OpenAI field or an Anthropic rate for the fixture.
	model := "claude-3-haiku-20240307"
	u := fakevendor.Canonical{Input: 10000, Output: 1000, CacheRead: 2000, CacheWrite: 3000}
	writes := statusLaunchFor(bin, skills, []string{model}, true, true, "anthropic", []fakevendor.Canonical{u})
	if writes.fatal != "" {
		fail(status, "cache-write session: %s", writes.fatal)
		return checks
	}
	text := skillResult(status, writes.reqs[1], "status-0", false)
	statusReport(status, text, model, u, u)
	statusCost(cost, text, 0, false)
	if !invoiceCounts(writes.after[0], "total", u) {
		fail(status, "GUI lost cache-write counts")
	}
	return checks
}
