package gradeedition2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter15Executable retains the seven behavioral weights, using the C# host's
// socket, lens and nullable usage contracts. Prices are supplied independently;
// this fixture never obtains expected rates from the application under test.
func Chapter15Executable(bin string, rates [4]float64) (grade.Report, error) {
	r := invoiceExecutableChecks(bin, rates)
	return grade.NewTitledReport("Edition 3, Chapter 15 — C# executable cache and meter checks",
		grade.Ch18Checks(grade.Ch18Result{Errs: r.Errs, Fatal: r.Fatal}), ""), nil
}

func invoiceSend(p *guiPeer, kind, text string, executable bool) error {
	if !executable {
		return p.send(kind, text)
	}
	return p.conn.WriteJSON(map[string]string{"type": kind, "text": text, "agent_id": "root"})
}

func contextReadEvents(path string) ([]map[string]json.RawMessage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var events []map[string]json.RawMessage
	for i, line := range bytes.Split(data, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var event map[string]json.RawMessage
		if err := json.Unmarshal(line, &event); err != nil {
			return nil, fmt.Errorf("event line %d: %w", i+1, err)
		}
		events = append(events, event)
	}
	return events, nil
}

// Standard Chat Completions has cached reads, not an invented write tariff.
// The explicit producer and service tier make price applicability testable.
func invoiceExecutableOpenAI(model, tier, text string, u fakevendor.Canonical) string {
	b, _ := json.Marshal(map[string]any{
		"id": "invoice", "object": "chat.completion", "model": model, "service_tier": tier,
		"choices": []any{map[string]any{"index": 0, "message": map[string]any{"role": "assistant", "content": text}, "finish_reason": "stop"}},
		"usage": map[string]any{"prompt_tokens": u.Input + u.CacheRead, "completion_tokens": u.Output,
			"prompt_tokens_details": map[string]int{"cached_tokens": u.CacheRead}},
	})
	return string(b)
}

func invoiceEmptySession(m map[string]any) bool {
	u, ok := m["usage"].(map[string]any)
	if !ok {
		return false
	}
	for _, k := range []string{"input_tokens", "cache_write_tokens", "cache_read_tokens", "output_tokens", "all_input"} {
		value, exists := u[k]
		if !exists || value != nil {
			return false
		}
	}
	hit, exists := m["hit_rate"]
	return exists && hit == nil && invoiceNear(m["total_cost_usd"], 0) && invoiceCounts(m, "total", fakevendor.Canonical{})
}

func invoiceExecutableChecks(bin string, rates [4]float64) contextResult {
	r := contextResult{Errs: map[string]string{}}
	for _, c := range grade.Ch18Checks(grade.Ch18Result{}) {
		r.set(c.ID, "")
	}
	dir, err := os.MkdirTemp("", "invoice-csharp-")
	if err != nil {
		r.Fatal = err.Error()
		return r
	}
	// Keep failed request/capture evidence. Successful fixtures have no durable use.
	defer func() {
		failed := r.Fatal != ""
		for _, problem := range r.Errs {
			failed = failed || problem != ""
		}
		if failed {
			fmt.Fprintf(os.Stderr, "Chapter 15 fixtures retained: %s\n", dir)
		} else {
			os.RemoveAll(dir)
		}
	}()
	skills, gui := filepath.Join(dir, "skills"), filepath.Join(dir, "web")
	if err = recallWrite(skills, map[string]string{"base/SKILL.md": "---\nname: base\ndescription: Cache and meter fixture\ntype: primary\ntools: read_file write_file edit_file run_command list_directory\n---\nYou are a helpful agent. Keep the invoice fixture instructions stable.\n"}); err != nil {
		r.Fatal = err.Error()
		return r
	}
	if err = contextStaticFixture(gui); err != nil {
		r.Fatal = err.Error()
		return r
	}
	run := func(name, vendor, model string, prompts []string, replies []fakevendor.Reply) invoiceOut {
		cwd := filepath.Join(dir, name)
		if err := os.MkdirAll(cwd, 0755); err != nil {
			return invoiceOut{fatal: err.Error()}
		}
		return invoiceLaunch(bin, skills, gui, invoiceOpts{dir: cwd, vendor: vendor, model: model, prompts: prompts, replies: replies, executable: true})
	}
	cold := fakevendor.Canonical{Input: 1000, CacheWrite: 500, Output: 10}
	warm := fakevendor.Canonical{Input: 100, CacheRead: 1400, Output: 20}
	markers := run("markers", "anthropic", "offline-cache-fixture", []string{"Remember copper-orbit-713.", "Continue violet-invoice-924."}, []fakevendor.Reply{
		{JSONBody: invoiceAnthropicUsage("offline-cache-fixture", "Copper retained.", cold)},
		{JSONBody: invoiceAnthropicUsage("offline-cache-fixture", "Violet complete.", warm)},
	})
	if markers.fatal != "" || len(markers.reqs) != 2 || len(markers.meters) != 3 {
		r.Fatal = fmt.Sprintf("marker session: %s requests=%d meters=%d", markers.fatal, len(markers.reqs), len(markers.meters))
		return r
	}
	invoiceLens(&r, markers, filepath.Join(dir, "markers"), "lens/request.wire.json", "lens/request.prev.wire.json")
	for i, name := range []string{"request.prev.wire.json", "request.wire.json"} {
		b, e := os.ReadFile(filepath.Join(dir, "markers", "lens", name))
		if e != nil || !bytes.Equal(b, markers.reqs[i].Body) {
			r.fail("cache-lens-exists", "%s does not equal actual sent bytes", name)
		}
	}
	invoicePrefix(&r, markers.reqs)
	invoiceUsage(&r, "all-vendors-report-usage", markers, []fakevendor.Canonical{cold, warm}, rates, false)
	if markers.meters[2]["total_cost_usd"] != nil {
		r.fail("cost-is-computed-not-stored", "unknown Anthropic identity was priced")
	}
	if !invoiceNear(markers.meters[1]["hit_rate"], 0) || !invoiceNear(markers.meters[2]["hit_rate"], 1400.0/3000) {
		r.fail("cache-read-rate", "disjoint fresh/write/read hit-rate mismatch")
	}

	// Pricing is independent of the explicit-marker provider. These exact OpenAI
	// rows have verified standard rates; unknown identities never borrow them.
	model := "gpt-4o-mini-2024-07-18"
	first := fakevendor.Canonical{Input: 1000, Output: 10}
	second := fakevendor.Canonical{Input: 100, CacheRead: 1400, Output: 20}
	priced := run("session", "openai", model, []string{"Remember copper-orbit-713.", "Continue violet-invoice-924."}, []fakevendor.Reply{
		{JSONBody: invoiceExecutableOpenAI(model, "default", "Copper retained.", first)},
		{JSONBody: invoiceExecutableOpenAI(model, "default", "Violet complete.", second)},
	})
	invoiceUsage(&r, "cost-is-computed-not-stored", priced, []fakevendor.Canonical{first, second}, rates, true)
	if priced.fatal != "" || len(priced.meters) != 3 {
		r.fail("usage-is-session-scoped", "priced session: %s", priced.fatal)
		return r
	}
	total := fakevendor.Canonical{Input: 1100, CacheRead: 1400, Output: 30}
	if !invoiceEmptySession(priced.meters[0]) || !invoiceCounts(priced.meters[2], "total", total) {
		r.fail("usage-is-session-scoped", "initial or accumulated counts wrong")
	}
	fresh := fakevendor.Canonical{Input: 7, Output: 3}
	restart := run("session", "openai", model, []string{"Continue amber-restart-381."}, []fakevendor.Reply{{JSONBody: invoiceExecutableOpenAI(model, "default", "Restart complete.", fresh)}})
	if restart.fatal != "" || len(restart.reqs) != 1 || len(restart.meters) != 2 {
		r.fail("usage-is-session-scoped", "restart: %s", restart.fatal)
	} else {
		body := string(restart.reqs[0].Body)
		for _, text := range []string{"Remember copper-orbit-713.", "Copper retained.", "Violet complete."} {
			if !strings.Contains(body, text) {
				r.fail("usage-is-session-scoped", "restart lost prior dialogue %s", text)
			}
		}
		if !invoiceEmptySession(restart.meters[0]) || !invoiceCounts(restart.meters[1], "", fresh) || !invoiceCounts(restart.meters[1], "total", fresh) || !invoiceNear(restart.meters[1]["total_cost_usd"], invoiceCost(fresh, rates)) {
			r.fail("usage-is-session-scoped", "restored history seeded the new meter")
		}
	}
	for _, identity := range []struct{ model, tier string }{{"unknown-price-fixture", "default"}, {model, "priority"}} {
		out := run("unpriced-"+identity.model+identity.tier, "openai", identity.model, []string{"Unknown price control."}, []fakevendor.Reply{{JSONBody: invoiceExecutableOpenAI(identity.model, identity.tier, "Unpriced complete.", second)}})
		invoiceUsage(&r, "cost-is-computed-not-stored", out, []fakevendor.Canonical{second}, rates, false)
		if len(out.meters) == 2 {
			cost, exists := out.meters[1]["total_cost_usd"]
			if !exists || cost != nil {
				r.fail("cost-is-computed-not-stored", "unknown model/tier did not explicitly retain unknown cost")
			}
		}
	}
	usage := fakevendor.Canonical{Input: 11, CacheRead: 33, Output: 44}
	for _, vendor := range []string{"openai", "gemini"} {
		reply := fakevendor.Reply{Text: "Usage complete.", Usage: usage, GeminiThoughts: 13}
		if vendor == "openai" {
			reply.JSONBody = invoiceExecutableOpenAI(model, "default", reply.Text, usage)
		}
		out := run(vendor, vendor, model, []string{"Usage parser control."}, []fakevendor.Reply{reply})
		invoiceUsage(&r, "all-vendors-report-usage", out, []fakevendor.Canonical{usage}, rates, false)
	}
	return r
}
