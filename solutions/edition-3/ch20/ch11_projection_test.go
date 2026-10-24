package ensemble_test

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
)

type c11ProjectionDefinition struct {
	Name     string `json:"name"`
	Deferred bool   `json:"defer_loading"`
	Schema   any    `json:"input_schema"`
}

// Inspect declarations/references, not automatic context text (which legitimately
// identifies its source). Registry truth must not be confused with the wire menu.
func c11ProjectionCatalog(t *testing.T, raw []byte) map[string]c11ProjectionDefinition {
	t.Helper()
	automatic := map[string]bool{"initial_auto": true, "initial_turn": true, "dynamic_auto": true, "dormant_auto": true}
	var wire struct{ Tools []c11ProjectionDefinition }
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	catalog := map[string]c11ProjectionDefinition{}
	for _, d := range wire.Tools {
		catalog[d.Name] = d
		if automatic[d.Name] {
			t.Errorf("automatic tool declared (deferred=%v): %s", d.Deferred, d.Name)
		}
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	var walk func(any)
	walk = func(v any) {
		switch x := v.(type) {
		case []any:
			for _, e := range x {
				walk(e)
			}
		case map[string]any:
			if x["type"] == "tool_reference" {
				if name, ok := x["tool_name"].(string); ok && automatic[name] {
					t.Errorf("automatic tool_reference disclosed: %s", name)
				}
			}
			for _, e := range x {
				walk(e)
			}
		}
	}
	walk(value)
	return catalog
}

func TestChapter11AutomaticAnthropicProjection(t *testing.T) {
	for _, model := range []string{"claude-sonnet-4-6", "claude-sonnet-4"} {
		t.Run(model, func(t *testing.T) {
			supported := model == "claude-sonnet-4-6"
			// The primary skill connects during NewAgent, before the first model turn.
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { listener.Close() })
			accepted := make(chan (<-chan string), 1)
			go func() {
				conn, err := listener.Accept()
				if err != nil {
					return
				}
				accepted <- c11ServePeer(t, ensemble.NewMCPStream(conn), []map[string]any{c11Tool("initial_auto", "round"), c11Tool("initial_turn", "turn"), c11Tool("dynamic_auto", "round")})
			}()
			dir := t.TempDir()
			declaration := "mcp_servers:\n  - name: projection-peer\n    transport: socket\n    address: " + listener.Addr().String() + "\n"
			c9Skill(t, dir, "base", "primary", "tools: read_file initial_auto initial_turn\nloadable-skills: writer\n"+declaration, "BASE $TOOLS")
			c9Skill(t, dir, "writer", "loadable", "tools: write_file dynamic_auto dormant_auto\n"+declaration, "WRITER $TOOLS")
			captured := make(chan []byte, 4)
			var requests, dormantCalls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				raw, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
					http.Error(w, "read request", 400)
					return
				}
				captured <- raw
				if requests.Add(1) == 1 {
					if dormantCalls.Load() != 0 {
						t.Error("dormant automatic tool ran before admission")
					}
					ch5Reply(w, ch5Call("projection-load", "load_skill", map[string]string{"name": "writer"}))
				} else {
					// Validate the actual provider request after the real load
					// result disclosed the deferred catalog entries.
					var request struct {
						Tools []struct {
							Name         string          `json:"name"`
							Deferred     bool            `json:"defer_loading"`
							CacheControl json.RawMessage `json:"cache_control"`
						}
					}
					if err := json.Unmarshal(raw, &request); err != nil {
						t.Error(err)
					}
					lastEligible, marked, toolMarkers := -1, -1, 0
					for i, tool := range request.Tools {
						if !tool.Deferred {
							lastEligible = i
						}
						if len(tool.CacheControl) > 0 {
							marked = i
							toolMarkers++
							if tool.Deferred {
								t.Errorf("illegal provider request: %s has defer_loading and cache_control", tool.Name)
							}
						}
					}
					count := strings.Count(string(raw), `"cache_control"`)
					if supported {
						if lastEligible < 0 || marked != lastEligible || toolMarkers != 1 {
							t.Errorf("tool breakpoint: eligible=%d marked=%d count=%d", lastEligible, marked, toolMarkers)
						}
						if count < 2 || count > 4 {
							t.Errorf("cache marker budget=%d", count)
						}
					} else if count != 0 {
						t.Errorf("unsupported model gained cache markers: %s", raw)
					}
					ch5Reply(w, map[string]string{"type": "text", "text": "done"})
				}
			}))
			defer server.Close()
			app := ensemble.New(nil)
			t.Cleanup(func() { _ = app.Shutdown() })
			a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.Anthropic, Model: model, APIKey: "local-fixture-only", BaseURL: server.URL, BaseDir: t.TempDir(), DataDir: t.TempDir(), Skills: &ensemble.SkillSpec{Directory: dir, Primary: "base"}})
			if err != nil {
				t.Fatal(err)
			}
			peerCalls := <-accepted
			// An application-registered automatic handler is initially dormant, then
			// genuinely enabled by the model's load_skill job alongside the MCP handler.
			err = a.Tools().RegisterTool(ensemble.ToolSpec{ToolDefinition: ensemble.ToolDefinition{Name: "dormant_auto", Description: "Automatic application context", Parameters: json.RawMessage(`{"type":"object"}`)}, Ephemeral: "round", Handler: func(_ context.Context, c ensemble.ToolCall, _ json.RawMessage) error {
				dormantCalls.Add(1)
				_, err := io.WriteString(c, "actual dormant handler output")
				return err
			}})
			if err != nil {
				t.Fatal(err)
			}
			dormant := c9Tool(t, a, "dormant_auto")
			if dormant.Enabled || dormant.EventSeq != 0 || dormant.Ephemeral != "round" {
				t.Fatalf("not initially dormant: %+v", dormant)
			}
			initial := c9Tool(t, a, "initial_auto")
			if !initial.Enabled || initial.Origin != "initial" || initial.EventSeq != 0 || initial.Ephemeral != "round" {
				t.Fatalf("startup capability truth: %+v", initial)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if _, err := a.Ask(ctx, "Load writer and finish."); err != nil {
				t.Fatal(err)
			}
			if requests.Load() != 2 {
				t.Fatalf("requests=%d", requests.Load())
			}
			before, after := <-captured, <-captured
			if supported {
				var initialWire, disclosedWire struct{ Tools []c11ProjectionDefinition }
				if err := json.Unmarshal(before, &initialWire); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(after, &disclosedWire); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(initialWire.Tools, disclosedWire.Tools) {
					t.Fatal("disclosure changed catalog order, schemas or defer flags")
				}
			}
			firstCatalog := c11ProjectionCatalog(t, before)
			secondCatalog := c11ProjectionCatalog(t, after)
			if d, ok := firstCatalog["write_file"]; supported {
				if !ok || !d.Deferred {
					t.Error("ordinary dormant schema lost from supported catalog")
				}
			} else if ok {
				t.Error("fallback exposed disabled ordinary tool")
			}
			if d, ok := firstCatalog["read_file"]; !ok || d.Deferred {
				t.Error("ordinary initial declaration changed")
			}
			for _, text := range []string{"actual peer: initial_auto", "actual peer: initial_turn"} {
				if !strings.Contains(string(before), text) || !strings.Contains(string(after), text) {
					t.Errorf("startup automatic context missing: %s", text)
				}
			}
			if !strings.Contains(string(after), "actual peer: dynamic_auto") || !strings.Contains(string(after), "actual dormant handler output") {
				t.Error("newly enabled automatic context missing")
			}
			counts := map[string]int{}
		drain:
			for {
				select {
				case name := <-peerCalls:
					counts[name]++
				default:
					break drain
				}
			}
			if counts["initial_auto"] != 2 || counts["initial_turn"] != 1 || counts["dynamic_auto"] != 1 || dormantCalls.Load() != 1 {
				t.Fatalf("actual automatic calls: %v dormant=%d", counts, dormantCalls.Load())
			}
			events := c9LoadEvents(a)
			if len(events) != 1 || events[0].Tool.CallID != "projection-load" {
				t.Fatalf("real load admissions: %+v", events)
			}
			for _, name := range []string{"write_file", "dynamic_auto", "dormant_auto"} {
				spec := c9Tool(t, a, name)
				if !spec.Enabled || spec.Origin != "tool" || spec.EventSeq != events[0].Seq {
					t.Errorf("real enabling provenance %s: %+v", name, spec)
				}
			}
			if again := c9Tool(t, a, "initial_auto"); again.Origin != initial.Origin || again.EventSeq != initial.EventSeq || again.Source != initial.Source {
				t.Error("startup provenance rebased")
			}
			var schema any
			if err := json.Unmarshal(c9Tool(t, a, "write_file").Parameters, &schema); err != nil {
				t.Fatal(err)
			}
			if d, ok := secondCatalog["write_file"]; !ok || d.Deferred != supported || !reflect.DeepEqual(d.Schema, schema) {
				t.Error("ordinary enabling schema/defer policy changed")
			}
			if supported {
				c9CheckWire(t, after, true, "projection-load")
			} else if strings.Contains(string(after), `"tool_reference"`) {
				t.Error("fallback emitted references")
			}
			history := a.Log().Context()
			status := ""
			calls, results := 0, 0
			for _, entry := range history.Dialogue {
				for _, part := range entry.Parts {
					if part.CallID != "projection-load" {
						continue
					}
					if part.Type == "tool_call" {
						calls++
						if part.Name != "load_skill" {
							t.Error("canonical enabling call changed")
						}
					}
					if part.Type == "tool_result" {
						results++
						if part.IsError {
							t.Error("load result failed")
						}
						for _, p := range part.Parts {
							if p.Type == "text" {
								status += p.Text
							}
						}
					}
				}
			}
			if calls != 1 || results != 1 || !strings.Contains(status, "writer") {
				t.Fatalf("canonical pair/status missing: calls=%d results=%d %q", calls, results, status)
			}
			if !supported {
				var request map[string]any
				if err := json.Unmarshal(after, &request); err != nil {
					t.Fatal(err)
				}
				escaped, _ := json.Marshal(status)
				found := false
				for _, result := range c3Results(request) {
					if result.id == "projection-load" && strings.Contains(result.text, string(escaped[1:len(escaped)-1])) {
						found = true
					}
				}
				if !found {
					t.Error("fallback lost ordinary canonical status")
				}
			}
			seq := a.Log().LastSeq()
			eventsBefore, _ := json.Marshal(a.Log().EventsRange(0, seq))
			contextBefore, _ := json.Marshal(history)
			replay, err := a.Engine().Render(history)
			if err != nil {
				t.Fatal(err)
			}
			c11ProjectionCatalog(t, replay)
			absent, err := a.Engine().Render(ensemble.Context{})
			if err != nil {
				t.Fatal(err)
			}
			c11ProjectionCatalog(t, absent)
			if supported {
				c9CheckWire(t, replay, true, "projection-load")
				c9CheckWire(t, absent, false, "")
			}
			eventsAfter, _ := json.Marshal(a.Log().EventsRange(0, a.Log().LastSeq()))
			contextAfter, _ := json.Marshal(history)
			if a.Log().LastSeq() != seq || string(eventsAfter) != string(eventsBefore) || string(contextAfter) != string(contextBefore) {
				t.Fatal("projection mutated canonical History/context")
			}
		})
	}
}
