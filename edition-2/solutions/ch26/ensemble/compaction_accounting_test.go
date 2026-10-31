package ensemble_test

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"sync/atomic"
	"testing"

	"ensemble/ensemble"
)

// TestCompactionAuxiliaryAccounting imports actual drained worker measurements,
// retaining each request's price tier and leaving the main occupancy untouched.
// Once-only attempt delivery is an actor obligation, outside this Engine seam.
func TestCompactionAuxiliaryAccounting(t *testing.T) {
	for _, mode := range []string{"priced", "failed", "unpriced"} {
		t.Run(mode, func(t *testing.T) {
			var calls atomic.Int32
			peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					Model string `json:"model"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
					http.Error(w, "bad request", 400)
					return
				}
				if request.Model == "claude-sonnet-5" {
					io.WriteString(w, `{"content":[{"type":"text","text":"main done"}],"usage":{"input_tokens":10,"cache_creation_input_tokens":2,"cache_read_input_tokens":3,"output_tokens":4}}`)
					return
				}
				n := calls.Add(1)
				prompt, arguments, stop := 200, `{"memory":17}`, "tool_calls"
				if n == 2 {
					prompt, arguments = 272001, `{"memory":"I kept the reason."}`
					if mode == "failed" {
						stop = "length"
					}
				}
				// The first complete tool call is schema-invalid. Its retry crosses
				// the existing price threshold, with both cache buckets included.
				call := map[string]any{"id": fmt.Sprintf("call-%d", n), "type": "function", "function": map[string]any{"name": "submit", "arguments": arguments}}
				json.NewEncoder(w).Encode(map[string]any{
					"choices": []any{map[string]any{"message": map[string]any{"content": "I will preserve the decision and its reason.", "tool_calls": []any{call}}, "finish_reason": stop}},
					"usage":   map[string]any{"prompt_tokens": prompt, "prompt_tokens_details": map[string]int{"cache_write_tokens": 40, "cached_tokens": 60}, "completion_tokens": 10},
				})
			}))
			t.Cleanup(peer.Close)
			app := ensemble.New(io.Discard)
			cfg := ensemble.Config{Vendor: ensemble.Anthropic, Model: "claude-sonnet-5", BaseURL: peer.URL, DataDir: t.TempDir(), DisableStreaming: true}
			parent, err := app.NewAgent(cfg)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { parent.Shutdown() })
			if _, err := parent.Ask(submissionContext(t), "main task"); err != nil {
				t.Fatal(err)
			}
			before, history := parent.Engine().Meter(), parent.History().Context()
			workerCfg := cfg
			workerCfg.Vendor, workerCfg.Model = ensemble.OpenAI, "gpt-5.6-sol"
			if mode == "unpriced" {
				workerCfg.Model = "gpt-4o"
			}
			child, err := app.NewChildAgent(parent, workerCfg, ensemble.AgentSpec{ID: "accounting-worker", Compaction: true, DataDir: filepath.Join(cfg.DataDir, "children", "accounting-worker")})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { child.Shutdown() })
			_, err = child.Ask(submissionContext(t), "source decisions")
			if (err != nil) != (mode == "failed") || calls.Load() != 2 {
				t.Fatalf("worker result: calls=%d error=%v", calls.Load(), err)
			}
			if err := child.Shutdown(); err != nil {
				t.Fatal(err)
			}
			if err := child.Save(); err != nil {
				t.Fatal(err)
			}
			buckets := child.Engine().UsageBuckets()
			wantBuckets := 2
			if mode == "unpriced" {
				wantBuckets = 1
			}
			if len(buckets) != wantBuckets {
				t.Fatalf("bucket count=%d want=%d", len(buckets), wantBuckets)
			}
			for _, b := range buckets {
				if b.Vendor != ensemble.OpenAI || b.Model != workerCfg.Model {
					t.Fatalf("bucket identity: %+v", b)
				}
				if mode != "unpriced" {
					want := ensemble.Usage{Input: 100, CacheWrite: 40, CacheRead: 60, Output: 10}
					if b.Large {
						want.Input = 271901
					}
					if b.Usage != want {
						t.Fatalf("tier counts: %+v want=%+v", b, want)
					}
				}
			}
			parent.Engine().AddAuxiliaryUsage(buckets)
			after := parent.Engine().Meter()
			imported := parent.Engine().UsageBuckets()
			if len(imported) != len(buckets)+1 {
				t.Fatalf("imported bucket count: %+v", imported)
			}
			for _, bucket := range buckets {
				if !slices.Contains(imported, bucket) {
					t.Fatalf("original bucket changed during import: %+v", bucket)
				}
			}
			wantTotal := ensemble.Usage{Input: 272011, CacheWrite: 82, CacheRead: 123, Output: 24}
			if after.Total != wantTotal || after.Usage != before.Usage || parent.History().Context().LastUsage != history.LastUsage || parent.History().Context().Usage != history.Usage {
				t.Fatalf("auxiliary polluted main occupancy or lost totals: %+v", after)
			}
			if mode == "unpriced" {
				if after.Priced || after.CostUSD != 0 {
					t.Fatalf("unpriced worker concealed: %+v", after)
				}
			} else if !after.Priced || math.Abs(after.CostUSD-before.CostUSD-2.17678) > 1e-9 {
				t.Fatalf("model/tier price lost: %+v", after)
			}
			buckets[0].Usage.Output = 999999
			unchanged := parent.Engine().Meter()
			if unchanged.Total != after.Total || unchanged.Usage != after.Usage || unchanged.Priced != after.Priced || math.Abs(unchanged.CostUSD-after.CostUSD) > 1e-12 {
				t.Fatal("receipt aliases live usage")
			}
			for _, bucket := range child.Engine().UsageBuckets() {
				if bucket.Usage.Output == 999999 {
					t.Fatal("receipt aliases child usage")
				}
			}
			if err := parent.Save(); err != nil {
				t.Fatal(err)
			}
			restored := loadStagedAgent(t, cfg)
			if restored.Engine().Usage() != (ensemble.Usage{}) || len(restored.Engine().UsageBuckets()) != 0 || restored.History().Context().LastUsage != history.LastUsage {
				t.Fatal("reload reimported prior process spend or changed main usage")
			}
		})
	}
}
