package gradeedition2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// ParseCachePrices reads four independently supplied dollars-per-million rates,
// ordered fresh input, cache write, cache read, output. Historical tables differ;
// reading expected prices from the implementation would not test arithmetic.
func ParseCachePrices(text string) ([4]float64, error) {
	var rates [4]float64
	parts := strings.Split(text, ",")
	if len(parts) != 4 {
		return rates, fmt.Errorf("cache prices require input,write,read,output")
	}
	for i, p := range parts {
		n, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil || math.IsNaN(n) || math.IsInf(n, 0) || n < 0 {
			return rates, fmt.Errorf("invalid cache price %q", p)
		}
		rates[i] = n
	}
	return rates, nil
}

// Chapter15 retains original18's seven weights. Artifact paths and dated prices
// configure the grader's observations; they do not extend the runtime protocol.
func Chapter15(bin, current, prior string, rates [4]float64) (grade.Report, error) {
	r := invoiceChecks(bin, current, prior, rates)
	return grade.NewTitledReport("Edition 2, Chapter 15 — The Invisible Invoice", grade.Ch18Checks(grade.Ch18Result{Errs: r.Errs, Fatal: r.Fatal}), ""), nil
}

func invoiceChecks(bin, current, prior string, rates [4]float64) contextResult {
	r := contextResult{Errs: map[string]string{}}
	for _, c := range grade.Ch18Checks(grade.Ch18Result{}) {
		r.set(c.ID, "")
	}
	dir, err := os.MkdirTemp("", "invoice-grade-")
	if err != nil {
		r.Fatal = err.Error()
		return r
	}
	defer os.RemoveAll(dir)
	skills, gui := filepath.Join(dir, "skills"), filepath.Join(dir, "web")
	if err = recallWrite(skills, map[string]string{"base/SKILL.md": "---\nname: base\ndescription: Cache and meter fixture\ntype: primary\ntools: read_file write_file edit_file run_command list_directory\n---\nYou are a helpful agent. Keep the invoice fixture instructions stable.\n"}); err != nil {
		r.Fatal = err.Error()
		return r
	}
	// These checks use stdin and WebSocket, not the shipped browser assets.
	if err = contextStaticFixture(gui); err != nil {
		r.Fatal = err.Error()
		return r
	}
	run := func(name, vendor, model string, prompts []string, replies []fakevendor.Reply) invoiceOut {
		cwd := filepath.Join(dir, name)
		if err := os.MkdirAll(cwd, 0755); err != nil {
			return invoiceOut{fatal: err.Error()}
		}
		if vendor == "anthropic" {
			// The priced and unknown-model controls must return their own
			// producer identity, not the shared fake's historical model name.
			for i := range replies {
				replies[i].JSONBody = invoiceAnthropicUsage(model, replies[i].Text, replies[i].Usage)
			}
		}
		return invoiceLaunch(bin, skills, gui, invoiceOpts{dir: cwd, vendor: vendor, model: model, prompts: prompts, replies: replies})
	}
	cold := fakevendor.Canonical{Input: 1000, CacheWrite: 500, Output: 10}
	warm := fakevendor.Canonical{Input: 100, CacheRead: 1400, Output: 20}
	first := run("session", "anthropic", "claude-sonnet-4-6", []string{"Remember copper-orbit-713.", "Continue violet-invoice-924."}, []fakevendor.Reply{{Text: "Copper retained.", Usage: cold}, {Text: "Violet complete.", Usage: warm}})
	if first.fatal != "" {
		r.Fatal = first.fatal
		return r
	}
	if len(first.reqs) != 2 || len(first.meters) != 3 {
		r.Fatal = "two completed turns/meter snapshots missing"
		return r
	}
	invoiceLens(&r, first, filepath.Join(dir, "session"), current, prior)
	invoicePrefix(&r, first.reqs)
	invoiceUsage(&r, "cost-is-computed-not-stored", first, []fakevendor.Canonical{cold, warm}, rates, true)
	total := fakevendor.Canonical{Input: 1100, CacheWrite: 500, CacheRead: 1400, Output: 30}
	if !invoiceNear(first.meters[1]["hit_rate"], 0) || !invoiceNear(first.meters[2]["hit_rate"], 1400.0/3000) || !invoiceCounts(first.meters[2], "", warm) {
		r.fail("cache-read-rate", "cold/warm read counts or cumulative input hit rate disagree with provider buckets")
	}
	if !invoiceCounts(first.meters[2], "total", total) {
		r.fail("usage-is-session-scoped", "two-turn totals are not the sum of disjoint response counts")
	}
	// Restoring an empty conversation would also reset the meter; require both
	// old dialogue on the next wire and zero session usage before new inference.
	fresh := fakevendor.Canonical{Input: 7, Output: 3}
	restart := run("session", "anthropic", "claude-sonnet-4-6", []string{"Continue amber-restart-381."}, []fakevendor.Reply{{Text: "Restart complete.", Usage: fresh}})
	if restart.fatal != "" {
		r.fail("usage-is-session-scoped", "restart: %s", restart.fatal)
	} else if len(restart.reqs) != 1 || len(restart.meters) != 2 {
		r.fail("usage-is-session-scoped", "restart observations missing")
	} else {
		text := memoryConversation(restart.reqs[0].Body)
		if !strings.Contains(text, "Remember copper-orbit-713.") || !strings.Contains(text, "Copper retained.") || !strings.Contains(text, "Violet complete.") {
			r.fail("usage-is-session-scoped", "restart did not retain the actual prior dialogue")
		}
		if !invoiceCounts(restart.meters[0], "", fakevendor.Canonical{}) || !invoiceCounts(restart.meters[0], "total", fakevendor.Canonical{}) || !invoiceNear(restart.meters[0]["cost_usd"], 0) || !invoiceNear(restart.meters[0]["hit_rate"], 0) {
			r.fail("usage-is-session-scoped", "restored history contaminated the initial live meter")
		}
		if !invoiceCounts(restart.meters[1], "", fresh) || !invoiceCounts(restart.meters[1], "total", fresh) || !invoiceNear(restart.meters[1]["cost_usd"], invoiceCost(fresh, rates)) {
			r.fail("usage-is-session-scoped", "new session counts/cost include old work")
		}
	}
	unknown := run("unpriced", "anthropic", "claude-sonnet-5-course", []string{"Unknown price control."}, []fakevendor.Reply{{Text: "Unpriced complete.", Usage: cold}})
	if unknown.fatal != "" || len(unknown.meters) != 2 {
		r.fail("cost-is-computed-not-stored", "unpriced model session: %s", unknown.fatal)
	} else if priced, ok := unknown.meters[1]["priced"].(bool); !ok || priced || !invoiceCounts(unknown.meters[1], "", cold) {
		r.fail("cost-is-computed-not-stored", "unknown tariff was priced or discarded its token counts")
	}
	for _, v := range []struct{ vendor, model string }{{"anthropic", "claude-sonnet-4-6"}, {"openai", "gpt-5.6"}, {"gemini", "gemini-3.8-flash"}} {
		usage := fakevendor.Canonical{Input: 11, CacheWrite: 22, CacheRead: 33, Output: 44}
		reply := fakevendor.Reply{Text: "Usage complete.", Usage: usage, GeminiThoughts: 13}
		if v.vendor == "openai" {
			// This model/API reports cache_write_input_tokens; do not change
			// the shared fake's historical cache_write_tokens packets.
			reply.JSONBody = invoiceOpenAIUsage(v.model, usage)
		}
		out := run(v.vendor, v.vendor, v.model, []string{"Usage parser control."}, []fakevendor.Reply{reply})
		if v.vendor == "gemini" {
			usage.CacheWrite = 0
		}
		invoiceUsage(&r, "all-vendors-report-usage", out, []fakevendor.Canonical{usage}, rates, false)
	}
	return r
}

// Chapter-local text-only Messages replies preserve the controlled producer.
// Other chapters keep the shared fake's intentionally distinct provenance.
func invoiceAnthropicUsage(model, text string, u fakevendor.Canonical) string {
	body, _ := json.Marshal(map[string]any{
		"id": "msg_invoice", "type": "message", "role": "assistant", "model": model,
		"content": []any{map[string]any{"type": "text", "text": text}}, "stop_reason": "end_turn",
		"usage": map[string]any{"input_tokens": u.Input, "cache_creation_input_tokens": u.CacheWrite,
			"cache_read_input_tokens": u.CacheRead, "output_tokens": u.Output},
	})
	return string(body)
}

// A plain Chat Completions response for this one usage-only control. Counts
// stay disjoint in the fixture; prompt_tokens contains all three input buckets.
func invoiceOpenAIUsage(model string, u fakevendor.Canonical) string {
	body, _ := json.Marshal(map[string]any{
		"id": "chatcmpl_invoice", "object": "chat.completion", "model": model,
		"choices": []any{map[string]any{"index": 0, "message": map[string]any{"role": "assistant", "content": "Usage complete."}, "finish_reason": "stop"}},
		"usage": map[string]any{"prompt_tokens": u.Input + u.CacheWrite + u.CacheRead, "completion_tokens": u.Output, "total_tokens": u.Input + u.CacheWrite + u.CacheRead + u.Output,
			"prompt_tokens_details": map[string]any{"cached_tokens": u.CacheRead, "cache_write_input_tokens": u.CacheWrite}},
	})
	return string(body)
}

func invoiceNear(value any, want float64) bool {
	n, ok := value.(float64)
	return ok && math.Abs(n-want) < 1e-10
}
func invoiceCost(u fakevendor.Canonical, r [4]float64) float64 {
	return (float64(u.Input)*r[0] + float64(u.CacheWrite)*r[1] + float64(u.CacheRead)*r[2] + float64(u.Output)*r[3]) / 1e6
}

// Both shipped clients expose the same counts. One groups session counts under
// total; the older one prefixes flat fields. Neither representation changes the
// distinction between last response and session usage.
func invoiceCounts(m map[string]any, scope string, u fakevendor.Canonical) bool {
	prefix := ""
	// C# publishes the same disjoint counts in its documented usage envelope.
	if usage, ok := m["usage"].(map[string]any); ok {
		m = usage
		if scope == "total" {
			m, _ = m["totals"].(map[string]any)
		}
		for key, want := range map[string]int{"input_tokens": u.Input, "cache_write_tokens": u.CacheWrite, "cache_read_tokens": u.CacheRead, "output_tokens": u.Output} {
			if !invoiceNear(m[key], float64(want)) {
				return false
			}
		}
		return true
	}
	if scope == "total" {
		if nested, ok := m["total"].(map[string]any); ok {
			m = nested
		} else {
			prefix = "session_"
		}
	}
	for key, want := range map[string]int{"input": u.Input, "cache_write": u.CacheWrite, "cache_read": u.CacheRead, "output": u.Output} {
		if !invoiceNear(m[prefix+key], float64(want)) {
			return false
		}
	}
	return true
}

func invoiceUsage(r *contextResult, id string, out invoiceOut, want []fakevendor.Canonical, rates [4]float64, priced bool) {
	if out.fatal != "" {
		r.fail(id, "session: %s", out.fatal)
		return
	}
	var counts []map[string]any
	for _, event := range out.events {
		var kind string
		json.Unmarshal(event["type"], &kind)
		if kind != "response_ended" {
			continue
		}
		var payload struct {
			// Usage retains the recorded normalized token buckets, without derived money.
			Usage map[string]any
		}
		if json.Unmarshal(event["response"], &payload) == nil && payload.Usage != nil {
			counts = append(counts, payload.Usage)
		}
	}
	if len(counts) != len(want) || len(out.meters) != len(want)+1 {
		r.fail(id, "completed usage records/meter snapshots: %d/%d, want %d/%d", len(counts), len(out.meters), len(want), len(want)+1)
		return
	}
	var total fakevendor.Canonical
	for i, u := range want {
		if !invoiceCounts(counts[i], "", u) {
			r.fail(id, "response %d durable buckets differ: %v", i+1, counts[i])
		}
		for key, value := range counts[i] {
			lower := strings.ToLower(key)
			if strings.Contains(lower, "cost") || strings.Contains(lower, "usd") || strings.Contains(lower, "dollar") || strings.Contains(lower, "price") || strings.Contains(lower, "spend") || strings.Contains(lower, "charge") {
				r.fail(id, "durable usage stores monetary field %s", key)
			}
			if number, ok := value.(float64); ok && number != math.Trunc(number) {
				r.fail(id, "durable usage stores fractional field %s", key)
			}
		}
		total.Input += u.Input
		total.CacheWrite += u.CacheWrite
		total.CacheRead += u.CacheRead
		total.Output += u.Output
		meter := out.meters[i+1]
		if !invoiceCounts(meter, "", u) || !invoiceCounts(meter, "total", total) {
			r.fail(id, "response %d last/session meter buckets differ: %v", i+1, meter)
		}
		cost, priceKnown := meter["cost_usd"], meter["priced"] == true
		if value, present := meter["total_cost_usd"]; present {
			cost, priceKnown = value, value != nil
		}
		if priced && (!priceKnown || !invoiceNear(cost, invoiceCost(total, rates))) {
			r.fail(id, "response %d cost %v priced=%v, want %.12f at supplied rates %v", i+1, cost, priceKnown, invoiceCost(total, rates), rates)
		}
	}
}

// Marker removal affects comparison metadata only. Raw section bytes otherwise
// stay raw: decoding and re-marshalling would hide serialization instability.
func invoiceStrip(raw []byte) []byte {
	marker := regexp.MustCompile(`,?\s*"cache_control"\s*:\s*\{[^{}]*\}\s*,?`)
	return marker.ReplaceAllFunc(raw, func(b []byte) []byte {
		text := strings.TrimSpace(string(b))
		if strings.HasPrefix(text, ",") && strings.HasSuffix(text, ",") {
			return []byte(",")
		}
		return nil
	})
}
func invoiceSections(raw []byte) map[string]json.RawMessage {
	var m map[string]json.RawMessage
	json.Unmarshal(raw, &m)
	return m
}

func invoicePrefix(r *contextResult, reqs []fakevendor.Recorded) {
	const stable = "prefix-is-stable"
	a, b := invoiceSections(reqs[0].Body), invoiceSections(reqs[1].Body)
	for _, key := range []string{"tools", "system"} {
		if len(a[key]) < 3 || bytes.Equal(a[key], []byte(`""`)) || !bytes.Equal(invoiceStrip(a[key]), invoiceStrip(b[key])) {
			r.fail(stable, "required %s section absent, empty or changed between actual requests", key)
		}
	}
	var tools []map[string]any
	json.Unmarshal(a["tools"], &tools)
	names := map[string]bool{}
	for _, tool := range tools {
		if name, ok := tool["name"].(string); ok {
			names[name] = true
		}
	}
	for _, name := range []string{"read_file", "write_file", "edit_file", "run_command", "list_directory"} {
		if !names[name] {
			r.fail(stable, "declared fixture tool %s absent from actual request", name)
		}
	}
	if !bytes.Contains(a["system"], []byte("invoice fixture instructions")) {
		r.fail(stable, "primary instructions absent from stable system section")
	}
	const markers = "request-has-breakpoints"
	previous := -1
	for i, req := range reqs {
		body := invoiceSections(req.Body)
		var messages []map[string]json.RawMessage
		json.Unmarshal(body["messages"], &messages)
		historyEnd := false
		if len(messages) > 0 {
			historyEnd = invoiceEndMarker(messages[len(messages)-1]["content"])
		}
		// This fixture has no ephemeral tail, handoff or memory. Its final
		// message is therefore the stable end; recursively finding metadata
		// inside a schema or a message object would accept an unusable marker.
		if !invoiceEndMarker(body["tools"]) || !invoiceEndMarker(body["system"]) || !historyEnd || len(messages)-1 <= previous || invoiceMarkers(req.Body) > 4 {
			r.fail(markers, "request %d lacks a legal last-tool/system/advancing stable-history marker or exceeds four markers", i+1)
		}
		previous = len(messages) - 1
	}
}
func invoiceEndMarker(raw []byte) bool {
	var blocks []map[string]json.RawMessage
	if json.Unmarshal(raw, &blocks) != nil || len(blocks) == 0 {
		return false
	}
	var marker map[string]any
	return json.Unmarshal(blocks[len(blocks)-1]["cache_control"], &marker) == nil && marker["type"] == "ephemeral"
}

func invoiceMarkers(raw []byte) int {
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return 0
	}
	var count func(any) int
	count = func(v any) int {
		n := 0
		switch x := v.(type) {
		case map[string]any:
			for k, v := range x {
				if k == "cache_control" {
					if m, ok := v.(map[string]any); ok && m["type"] == "ephemeral" {
						n++
					}
				} else {
					n += count(v)
				}
			}
		case []any:
			for _, v := range x {
				n += count(v)
			}
		}
		return n
	}
	return count(value)
}

// Artifact formatting is not a product API. Decode embedded JSON values so a
// pretty report, a raw request, and either section-label style remain valid.
// This proves content and rotation, not byte-offset diagnostic correctness.
func invoiceArtifactContains(artifact, want []byte) bool {
	var expected any
	if json.Unmarshal(invoiceStrip(want), &expected) != nil {
		return false
	}
	clean := invoiceStrip(artifact)
	for i, b := range clean {
		if b != '[' && b != '{' && b != '"' {
			continue
		}
		var got any
		if json.NewDecoder(bytes.NewReader(clean[i:])).Decode(&got) == nil && reflect.DeepEqual(got, expected) {
			return true
		}
	}
	return false
}
func invoiceLens(r *contextResult, out invoiceOut, dir, current, prior string) {
	const id = "cache-lens-exists"
	a, firstErr := invoiceRead(dir, current)
	b, priorErr := invoiceRead(dir, prior)
	if firstErr == nil && priorErr == nil && bytes.Equal(a, b) {
		r.fail(id, "current and prior captures are identical, not a rotated request pair")
		return
	}
	for i, path := range []string{prior, current} {
		raw, err := invoiceRead(dir, path)
		if err != nil {
			r.fail(id, "capture %s: %v", path, err)
			continue
		}
		sections := invoiceSections(out.reqs[i].Body)
		for _, key := range []string{"tools", "system", "messages"} {
			if len(sections[key]) < 3 || !invoiceArtifactContains(raw, sections[key]) {
				r.fail(id, "capture %s lacks actual request %d %s content", path, i+1, key)
			}
		}
	}
}
