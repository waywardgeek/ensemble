package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"ensemble/ensemble"
	"ensemble/internal/common"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
)

// TestCacheLensMainBoundary drives real HTTP and owned diagnostics. Rendering
// without a send must not replace the preceding conversation request.
func TestCacheLensMainBoundary(t *testing.T) {
	for _, vendor := range []ensemble.Vendor{ensemble.Anthropic, ensemble.OpenAI, ensemble.Gemini} {
		t.Run(vendor.String(), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch vendor {
				case ensemble.Anthropic:
					replyText(w, "ok")
				case ensemble.OpenAI:
					io.WriteString(w, `{"choices":[{"message":{"content":"ok"}}],"usage":{"prompt_tokens":10,"completion_tokens":2}}`)
				case ensemble.Gemini:
					io.WriteString(w, `{"candidates":[{"content":{"parts":[{"text":"ok"}]}}],"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":2}}`)
				}
			}))
			defer server.Close()
			var logs bytes.Buffer
			a, err := ensemble.New(&logs).NewAgent(ensemble.Config{Vendor: vendor, Model: "fake-model", BaseURL: server.URL, DataDir: t.TempDir()})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Shutdown()
			for _, q := range []string{"red", "green"} {
				if _, err := a.Ask(context.Background(), q); err != nil {
					t.Fatal(err)
				}
			}
			lens := a.Engine().CacheLens()
			if lens == nil || lens.Engine() != a.Engine() {
				t.Fatal("lens ownership")
			}
			current := filepath.Join(a.Config().DataDir, "cache-current.txt")
			before, err := os.ReadFile(current)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(a.Config().DataDir, "cache-prior.txt")); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(logs.String(), ":appended") || strings.Contains(logs.String(), "cache prefix changed") || !strings.Contains(logs.String(), "cache grant:") {
				t.Fatalf("verdicts: %s", logs.String())
			}
			a.Engine().Render(a.History().Context())
			after, _ := os.ReadFile(current)
			if !bytes.Equal(before, after) {
				t.Fatal("render without send rotated lens")
			}
			body, _ := a.Engine().Render(a.History().Context())
			body = bytes.ReplaceAll(body, []byte("helpful assistant"), []byte("different assistant"))
			lens.ObserveRequest(vendor.String()+"/fake-model", body)
			if !strings.Contains(logs.String(), "cache prefix changed") {
				t.Fatal("stable-prefix rewrite unreported")
			}
		})
	}
}

// TestCacheLensSerializedMarkers checks the actual renderer in compact and
// indented forms; marker stripping must preserve unrelated escaping and bytes.
func TestCacheLensSerializedMarkers(t *testing.T) {
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "fake", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	body, err := a.Engine().Render(&common.Context{Dialogue: []common.Entry{{Actor: common.Human, Parts: []common.Part{{Type: "text", Text: "keep <escaped> bytes"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, pretty := range []bool{false, true} {
		marked := bytes.Replace(body, []byte(`"type":"text"`), []byte(`"type":"text","cache_control":{"type":"ephemeral"}`), 1)
		plain := body
		if pretty {
			var x, y bytes.Buffer
			json.Indent(&x, marked, "", "  ")
			json.Indent(&y, plain, "", "  ")
			marked, plain = x.Bytes(), y.Bytes()
		}
		lens := a.Engine().CacheLens()
		lens.ObserveRequest("fixture", marked)
		lens.ObserveRequest("fixture", plain)
		previous, _ := os.ReadFile(filepath.Join(a.Config().DataDir, "cache-prior.txt"))
		current, _ := os.ReadFile(filepath.Join(a.Config().DataDir, "cache-current.txt"))
		if !bytes.Equal(previous, current) {
			t.Fatalf("marker changed serialized artifact:\n%s\n%s", previous, current)
		}
		if !bytes.Contains(current, []byte(`\u003cescaped\u003e`)) {
			t.Fatal("escaped request bytes rewritten")
		}
	}
}

// TestCacheMarkers follows retained history through a checkpoint, ephemera and
// another user turn. It checks wire positions, not marker counts alone.
func TestCacheMarkers(t *testing.T) {
	for _, tc := range []struct {
		vendor ensemble.Vendor
		model  string
		marker string
	}{{ensemble.Anthropic, "claude-sonnet-5", "cache_control"}, {ensemble.OpenAI, "gpt-5.6-sol", "prompt_cache_breakpoint"}, {ensemble.Gemini, "gemini-3.8-flash", ""}, {ensemble.OpenAI, "gpt-4o-mini", ""}} {
		t.Run(tc.model, func(t *testing.T) {
			a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Vendor: tc.vendor, Model: tc.model, DataDir: t.TempDir(), BuiltinTools: true})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Shutdown()
			record := func(event common.Event) {
				t.Helper()
				if err := a.Record(event); err != nil {
					t.Fatal(err)
				}
			}
			record(common.Event{Type: "message_received", Message: &common.MessageData{Actor: common.Human, Parts: []common.Part{{Type: "text", Text: "old user"}}}})
			record(common.Event{Type: "micro_handoff", Handoff: &common.MicroHandoffData{Text: "settled boundary"}})
			record(common.Event{Type: "message_received", Message: &common.MessageData{Actor: common.Human, Parts: []common.Part{{Type: "text", Text: "stable tail"}}}})
			a.Ephemeral("volatile once")
			body, err := a.Engine().Render(a.History().Context())
			if err != nil {
				t.Fatal(err)
			}
			var wire map[string]any
			json.Unmarshal(body, &wire)
			var marked []string
			var walk func(any)
			walk = func(v any) {
				switch x := v.(type) {
				case []any:
					for _, e := range x {
						walk(e)
					}
				case map[string]any:
					if tc.marker != "" && x[tc.marker] != nil {
						if text, ok := x["text"].(string); ok {
							marked = append(marked, text)
						}
					}
					for _, e := range x {
						walk(e)
					}
				}
			}
			walk(wire)
			if tc.marker == "" {
				if bytes.Contains(body, []byte("cache_control")) || bytes.Contains(body, []byte("prompt_cache_")) {
					t.Fatalf("unsupported marker: %s", body)
				}
				return
			}
			boundary := "settled boundary"
			if tc.vendor == ensemble.OpenAI {
				// Chat Completions joins the host completion label and stored note.
				boundary = "micro_handoff completed. Your checkpoint note:" + boundary
			}
			if !slices.Contains(marked, boundary) || !slices.Contains(marked, "stable tail") || slices.Contains(marked, "volatile once") {
				t.Fatalf("wrong boundaries: %v", marked)
			}
			if tc.vendor == ensemble.Anthropic {
				tools := wire["tools"].([]any)
				if tools[len(tools)-1].(map[string]any)[tc.marker] == nil {
					t.Fatal("tools not banked")
				}
				if len(marked) != 3 || bytes.Count(body, []byte(`"cache_control"`)) != 4 {
					t.Fatal("Anthropic marker budget/system")
				}
			} else if wire["prompt_cache_options"] == nil || wire["reasoning_effort"] != "none" || len(marked) != 3 {
				t.Fatal("explicit opt-in/system")
			}
			record(common.Event{Type: "request_sent", Request: &common.RequestData{To: a.Engine().Target()}})
			record(common.Event{Type: "response_ended", Response: &common.ResponseData{From: a.Engine().Target(), Parts: []common.Part{{Type: "text", Text: "next tail"}}}})
			next, _ := a.Engine().Render(a.History().Context())
			marked = nil
			json.Unmarshal(next, &wire)
			walk(wire)
			if !slices.Contains(marked, "next tail") || slices.Contains(marked, "stable tail") || bytes.Contains(next, []byte("volatile once")) {
				t.Fatalf("rolling marker did not advance: %v %s", marked, next)
			}

		})
	}
}

// TestCacheMeterSessionAndTiers uses independently specified provider usage:
// Anthropic input excludes cache, OpenAI includes it, Gemini thinking is extra.
func TestCacheMeterSessionAndTiers(t *testing.T) {
	for _, tc := range []struct {
		vendor       ensemble.Vendor
		model, reply string
		want         common.Usage
		cost         float64
	}{
		{ensemble.Anthropic, "claude-sonnet-5", `{"content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":100001,"cache_creation_input_tokens":40,"cache_read_input_tokens":60,"output_tokens":10}}`, common.Usage{Input: 100001, CacheWrite: 40, CacheRead: 60, Output: 10}, .200214},
		{ensemble.OpenAI, "gpt-5.6-sol", `{"choices":[{"message":{"content":"ok"}}],"usage":{"prompt_tokens":272001,"prompt_tokens_details":{"cache_write_tokens":40,"cached_tokens":60},"completion_tokens":10}}`, common.Usage{Input: 271901, CacheWrite: 40, CacheRead: 60, Output: 10}, 2.175956},
		{ensemble.Gemini, "gemini-3.8-flash", `{"candidates":[{"content":{"parts":[{"text":"ok"}]}}],"usageMetadata":{"promptTokenCount":100,"cachedContentTokenCount":60,"candidatesTokenCount":7,"thoughtsTokenCount":3}}`, common.Usage{Input: 40, CacheRead: 60, Output: 10}, .000072},
	} {
		t.Run(tc.model, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); io.WriteString(w, tc.reply) }))
			defer server.Close()
			cfg := ensemble.Config{Vendor: tc.vendor, Model: tc.model, BaseURL: server.URL, DataDir: t.TempDir(), DisableStreaming: true}
			a, err := ensemble.New(io.Discard).NewAgent(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := a.Ask(context.Background(), "one"); err != nil {
				t.Fatal(err)
			}
			first := a.Engine().Meter()
			if first.Usage != tc.want || first.Total != tc.want || !first.Priced || math.Abs(first.CostUSD-tc.cost) > 1e-9 {
				t.Fatalf("usage/pricing: %+v want %v cost %.9f", first, tc.want, tc.cost)
			}
			if _, err := a.Ask(context.Background(), "two"); err != nil {
				t.Fatal(err)
			}
			second := a.Engine().Meter()
			if second.Usage != tc.want || second.Total.Output != 20 || math.Abs(second.CostUSD-2*tc.cost) > 1e-9 {
				t.Fatalf("last/session counts: %+v", second)
			}
			wantRate := float64(tc.want.CacheRead) / float64(tc.want.Input+tc.want.CacheWrite+tc.want.CacheRead)
			if math.Abs(second.HitRate-wantRate) > 1e-12 {
				t.Fatal("token hit ratio")
			}
			if err := a.Save(); err != nil {
				t.Fatal(err)
			}
			a.Shutdown()
			saved, _ := os.ReadFile(filepath.Join(cfg.DataDir, "save.json"))
			if bytes.Contains(saved, []byte("cost")) {
				t.Fatal("dollars persisted")
			}
			restored, err := ensemble.New(io.Discard).NewAgent(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer restored.Shutdown()
			if err := restored.Load(); err != nil {
				t.Fatal(err)
			}
			meter := restored.Engine().Meter()
			if meter.Total != (common.Usage{}) || meter.Usage != (common.Usage{}) || meter.CostUSD != 0 || restored.History().Context().Usage.Output != 20 {
				t.Fatal("session restored lifetime spend")
			}
			if calls.Load() != 2 {
				t.Fatal("restore contacted vendor")
			}
		})
	}
}

// TestCacheLensByteCounts pins the measured unit to serialized section bytes,
// including JSON punctuation, without involving a tokenizer or a remarshal.
func TestCacheLensByteCounts(t *testing.T) {
	var logs bytes.Buffer
	a, err := ensemble.New(&logs).NewAgent(ensemble.Config{Model: "fake", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	old := `[{"role":"user","content":"one"}]`
	next := `[{"role":"user","content":"one"},{"role":"assistant","content":"two"}]`
	lens := a.Engine().CacheLens()
	lens.ObserveRequest("fixture", []byte(`{"system":"stable","messages":`+old+`}`))
	lens.ObserveRequest("fixture", []byte(`{"system":"stable","messages":`+next+`}`))
	expected := fmt.Sprintf("prefix_bytes=%d prompt_bytes=%d", len(`"stable"`)+len(old)-1, len(`"stable"`)+len(next))
	if !strings.Contains(logs.String(), expected) {
		t.Fatalf("serialized-byte accounting: %s want %s", logs.String(), expected)
	}
}

// TestCacheEmptyOptIn prevents explicit mode from disabling the provider's
// implicit cache when an intentionally empty primary has no content to mark.
func TestCacheEmptyOptIn(t *testing.T) {
	dir := skillFiles(t, map[string]string{"base": "---\nname: base\ntype: primary\n---\n"})
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-5.6-sol", DataDir: t.TempDir()}, ensemble.AgentSpec{SkillsDir: dir, PrimarySkill: "base"})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	body, err := a.Engine().Render(a.History().Context())
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(body, []byte("prompt_cache_")) {
		t.Fatalf("empty explicit opt-in: %s", body)
	}
}
