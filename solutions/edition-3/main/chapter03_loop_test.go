package ensemble_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
	"ensemble/internal/common"
)

func c3ToolPart(id, name string, args map[string]any) common.Part {
	raw, _ := json.Marshal(args)
	return common.Part{Type: "tool_call", CallID: id, Name: name, Args: raw}
}

// The response deliberately has a terminal-looking stop reason even with calls.
// Real Engine parsing, not a replacement dispatcher, feeds the real Agent loop.
func c3Reply(v ensemble.Vendor, parts ...common.Part) any {
	blocks := []any{}
	calls := []any{}
	var text strings.Builder
	for _, p := range parts {
		switch v {
		case ensemble.Anthropic:
			switch p.Type {
			case "text":
				blocks = append(blocks, map[string]any{"type": "text", "text": p.Text})
			case "tool_call":
				blocks = append(blocks, map[string]any{"type": "tool_use", "id": p.CallID, "name": p.Name, "input": p.Args})
			case "opaque":
				blocks = append(blocks, json.RawMessage(p.Data))
			}
		case ensemble.OpenAI:
			if p.Type == "text" {
				text.WriteString(p.Text)
			}
			if p.Type == "tool_call" {
				calls = append(calls, map[string]any{"type": "function", "id": p.CallID, "function": map[string]any{"name": p.Name, "arguments": string(p.Args)}})
			}
		case ensemble.Gemini:
			if p.Type == "text" {
				blocks = append(blocks, map[string]any{"text": p.Text})
			}
			if p.Type == "tool_call" {
				blocks = append(blocks, map[string]any{"functionCall": map[string]any{"id": p.CallID, "name": p.Name, "args": p.Args}})
			}
		}
	}
	switch v {
	case ensemble.Anthropic:
		return map[string]any{"model": "local-fixture-model", "content": blocks, "stop_reason": "end_turn", "usage": map[string]any{"input_tokens": 0, "output_tokens": 0}}
	case ensemble.OpenAI:
		return map[string]any{"model": "local-fixture-model", "choices": []any{map[string]any{"message": map[string]any{"role": "assistant", "content": text.String(), "tool_calls": calls}, "finish_reason": "stop"}}, "usage": map[string]any{"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0}}
	default:
		return map[string]any{"modelVersion": "local-fixture-model", "candidates": []any{map[string]any{"content": map[string]any{"role": "model", "parts": blocks}, "finishReason": "STOP"}}, "usageMetadata": map[string]any{"promptTokenCount": 0, "candidatesTokenCount": 0, "totalTokenCount": 0}}
	}
}
func c3Peer(t *testing.T, reply func(int, map[string]any) any) (string, *atomic.Int32) {
	t.Helper()
	count := new(atomic.Int32)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(count.Add(1))
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("fixture request decode: %v", err)
			http.Error(w, "bad fixture request", 400)
			return
		}
		response := reply(n, request)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("fixture response encode: %v", err)
		}
	}))
	t.Cleanup(s.Close)
	return s.URL, count
}

type c3WireResult struct{ id, text string }

func c3Results(request map[string]any) []c3WireResult {
	results := []c3WireResult{}
	var walk func(any)
	walk = func(value any) {
		switch value := value.(type) {
		case []any:
			for _, item := range value {
				walk(item)
			}
		case map[string]any:
			id := ""
			var payload any
			if value["type"] == "tool_result" {
				id, _ = value["tool_use_id"].(string)
				payload = value["content"]
			}
			if value["role"] == "tool" {
				id, _ = value["tool_call_id"].(string)
				payload = value["content"]
			}
			if result, ok := value["functionResponse"].(map[string]any); ok {
				id, _ = result["id"].(string)
				payload = result["response"]
			}
			if id != "" {
				raw, _ := json.Marshal(payload)
				results = append(results, c3WireResult{id, string(raw)})
				return
			}
			for _, item := range value {
				walk(item)
			}
		}
	}
	// Never interpret tool declarations as messages or calls.
	walk(request["messages"])
	walk(request["contents"])
	return results
}
func c3CheckDeclarations(t *testing.T, v ensemble.Vendor, request map[string]any, want int) {
	t.Helper()
	tools, present := request["tools"]
	if want == 0 {
		if present {
			t.Error("empty registry sent a tools field")
		}
		return
	}
	entries, ok := tools.([]any)
	if !ok {
		t.Error("missing tools array")
		return
	}
	if v == ensemble.Gemini {
		if len(entries) != 1 {
			t.Errorf("Gemini tool containers=%d", len(entries))
			return
		}
		container, ok := entries[0].(map[string]any)
		if !ok {
			t.Error("invalid Gemini tools container")
			return
		}
		entries, ok = container["functionDeclarations"].([]any)
		if !ok {
			t.Error("missing Gemini functionDeclarations")
			return
		}
	}
	names := map[string]bool{}
	for _, entry := range entries {
		d, ok := entry.(map[string]any)
		if !ok {
			t.Error("invalid declaration")
			continue
		}
		schema := "parameters"
		if v == ensemble.OpenAI {
			if d["type"] != "function" {
				t.Error("OpenAI declaration type")
			}
			d, _ = d["function"].(map[string]any)
		}
		if v == ensemble.Anthropic {
			schema = "input_schema"
		}
		name, _ := d["name"].(string)
		if name == "" || names[name] {
			t.Errorf("missing/duplicate declaration name %q", name)
		}
		names[name] = true
		if d["description"] == "" || d[schema] == nil {
			t.Errorf("incomplete declaration for %q", name)
		}
	}
	if len(names) != want {
		t.Errorf("declarations=%v want count %d", names, want)
	}
	for _, name := range []string{"read_file", "list_directory", "search_files", "write_file", "edit_file", "run_command", "tool_limits", "wait_for_job", "send_input", "kill_job", "compact_context", "micro_handoff", "keep_tool_results", "agent_status", "spawn_sub_agent", "join_agents", "agents_status", "shutdown_agent", "send_message_to_parent", "respond_to_parent_message", "submit"} {
		if want == 21 && !names[name] {
			t.Errorf("missing default declaration %s", name)
		}
	}
}
func c3CheckPairs(t *testing.T, a ensemble.Agent, want int) []common.ToolData {
	t.Helper()
	calls := map[string]common.Part{}
	returns := []common.ToolData{}
	var pending *common.ToolData
	for _, e := range a.Log().Events() {
		switch e.Type {
		case common.ResponseEnded:
			for _, p := range e.Response.Parts {
				if p.Type == "tool_call" {
					if _, exists := calls[p.CallID]; exists || p.CallID == "" {
						t.Errorf("duplicate/missing call ID %q", p.CallID)
					}
					calls[p.CallID] = p
				}
			}
		case common.ToolCalled:
			if pending != nil {
				t.Error("next call started before preceding result")
			}
			p, exists := calls[e.Tool.CallID]
			if !exists || p.Name != e.Tool.Name || string(p.Args) != string(e.Tool.Args) {
				t.Errorf("call not present in a complete preceding response: %+v", e.Tool)
			}
			copy := *e.Tool
			pending = &copy
		case common.ToolReturned:
			if pending == nil || pending.CallID != e.Tool.CallID || pending.Name != e.Tool.Name {
				t.Errorf("unpaired/out-of-order return: %+v", e.Tool)
			}
			pending = nil
			returns = append(returns, *e.Tool)
		}
	}
	if pending != nil || len(calls) != want || len(returns) != want {
		t.Errorf("pair totals: calls=%d returns=%d want=%d pending=%v", len(calls), len(returns), want, pending)
	}
	return returns
}
func c3LocalAgent(t *testing.T, url string, v ensemble.Vendor, policy bool, tools []string) ensemble.Agent {
	return c3Agent(t, ensemble.Config{BaseURL: url, APIKey: "synthetic-local-fixture", Vendor: v, RequireToolExplanation: policy, Tools: tools})
}

func TestChapter3LoopOrderedBatchProviders(t *testing.T) {
	for _, v := range []ensemble.Vendor{ensemble.Anthropic, ensemble.OpenAI, ensemble.Gemini} {
		t.Run(fmt.Sprint(v), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "file")
			url, count := c3Peer(t, func(n int, request map[string]any) any {
				c3CheckDeclarations(t, v, request, 21)
				if n == 1 {
					return c3Reply(v, common.Text("I will create a fixture, try the disabled name, append, then read to check the final bytes."),
						c3ToolPart("write", "write_file", map[string]any{"path": path, "content": "A"}),
						c3ToolPart("disabled", "not_enabled", map[string]any{}),
						c3ToolPart("append", "write_file", map[string]any{"path": path, "content": "B", "append": true}),
						c3ToolPart("read", "read_file", map[string]any{"path": path}))
				}
				if n != 2 {
					t.Errorf("unexpected request %d", n)
				}
				results := c3Results(request)
				if len(results) != 4 {
					t.Errorf("wire results=%v", results)
				} else {
					for i, id := range []string{"write", "disabled", "append", "read"} {
						if results[i].id != id {
							t.Errorf("result order/ID=%v", results)
						}
					}
					if !strings.Contains(results[1].text, "unknown tool") || !strings.Contains(results[3].text, "AB") {
						t.Errorf("wrong actual outputs: %v", results)
					}
				}
				data, err := os.ReadFile(path)
				if err != nil || string(data) != "AB" {
					t.Errorf("ordered effects: %q %v", data, err)
				}
				return c3Reply(v, common.Text("Read back AB."))
			})
			a := c3LocalAgent(t, url, v, true, nil)
			reply, err := a.Ask(context.Background(), "Perform the fixture work.")
			if err != nil || reply != "Read back AB." || count.Load() != 2 {
				t.Fatalf("loop reply=%q err=%v requests=%d", reply, err, count.Load())
			}
			results := c3CheckPairs(t, a, 4)
			for i, r := range results {
				if r.IsError != (i == 1) {
					t.Errorf("result %d error=%t", i, r.IsError)
				}
			}
			if len(a.History()) != 2 {
				t.Errorf("intermediate tool replies entered text history: %+v", a.History())
			}
		})
	}
}

func TestChapter3LoopEmptyRegistryProviders(t *testing.T) {
	for _, v := range []ensemble.Vendor{ensemble.Anthropic, ensemble.OpenAI, ensemble.Gemini} {
		t.Run(fmt.Sprint(v), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "must-not-exist")
			url, count := c3Peer(t, func(n int, request map[string]any) any {
				c3CheckDeclarations(t, v, request, 0)
				if n == 1 {
					return c3Reply(v, c3ToolPart("disabled", "write_file", map[string]any{"path": path, "content": "forbidden"}))
				}
				if n != 2 {
					t.Errorf("unexpected request %d", n)
				}
				results := c3Results(request)
				if len(results) != 1 || results[0].id != "disabled" || !strings.Contains(results[0].text, "unknown tool") {
					t.Errorf("missing disabled-call result: %v", results)
				}
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Errorf("empty registry caused an effect: %v", err)
				}
				return c3Reply(v, common.Text("No tools enabled."))
			})
			a := c3LocalAgent(t, url, v, false, []string{})
			reply, err := a.Ask(context.Background(), "Try the absent tool.")
			if err != nil || reply != "No tools enabled." || count.Load() != 2 {
				t.Fatalf("empty registry reply=%q err=%v requests=%d", reply, err, count.Load())
			}
			results := c3CheckPairs(t, a, 1)
			if len(results) == 1 && !results[0].IsError {
				t.Fatal("disabled call was not an error result")
			}
		})
	}
}

func TestChapter3LoopExplanationPolicy(t *testing.T) {
	for _, kind := range []string{"blank", "late", "opaque", "prior"} {
		t.Run(kind, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "file")
			priorRequests := 0
			if kind == "prior" {
				priorRequests = 1
			}
			url, count := c3Peer(t, func(n int, request map[string]any) any {
				if kind == "prior" && n == 1 {
					return c3Reply(ensemble.Anthropic, common.Text("I previously explained a different turn."))
				}
				step := n - priorRequests
				if step == 1 {
					calls := []common.Part{c3ToolPart("refuse-write", "write_file", map[string]any{"path": path, "content": "bad"}), c3ToolPart("refuse-shell", "run_command", map[string]any{"command": fmt.Sprintf("printf bad > '%s'", path)})}
					switch kind {
					case "blank":
						calls = append([]common.Part{common.Text(" \t\n")}, calls...)
					case "late":
						calls = append(calls, common.Text("I explained only after the calls."))
					case "opaque":
						calls = append([]common.Part{{Type: "opaque", Data: json.RawMessage(`{"type":"thinking","thinking":"I will write to check the file.","signature":"fixture-signature"}`)}}, calls...)
					}
					return c3Reply(ensemble.Anthropic, calls...)
				}
				if step == 2 {
					if _, err := os.Stat(path); !os.IsNotExist(err) {
						t.Errorf("refused batch had effects: %v", err)
					}
					results := c3Results(request)
					if len(results) != 2 {
						t.Errorf("refused wire pairs=%v", results)
					} else {
						for i, id := range []string{"refuse-write", "refuse-shell"} {
							if results[i].id != id || !strings.Contains(results[i].text, "No tools ran") {
								t.Errorf("unpaired refusal: %v", results)
							}
						}
					}
					return c3Reply(ensemble.Anthropic, common.Text("I will write a new fixture and read it back to verify the retry."), c3ToolPart("retry-write", "write_file", map[string]any{"path": path, "content": "retry"}), c3ToolPart("retry-read", "read_file", map[string]any{"path": path}))
				}
				if step != 3 {
					t.Errorf("unexpected policy step %d", step)
				}
				data, err := os.ReadFile(path)
				if err != nil || string(data) != "retry" {
					t.Errorf("explained retry not executed: %q %v", data, err)
				}
				results := c3Results(request)
				if len(results) != 4 || results[3].id != "retry-read" || !strings.Contains(results[3].text, "retry") {
					t.Errorf("retry result not consumed: %v", results)
				}
				return c3Reply(ensemble.Anthropic, common.Text("Retry verified."))
			})
			a := c3LocalAgent(t, url, ensemble.Anthropic, true, nil)
			if priorRequests != 0 {
				if _, err := a.Ask(context.Background(), "Earlier turn."); err != nil {
					t.Fatal(err)
				}
			}
			reply, err := a.Ask(context.Background(), "Do the work.")
			if err != nil || reply != "Retry verified." || count.Load() != int32(3+priorRequests) {
				t.Fatalf("policy reply=%q err=%v requests=%d", reply, err, count.Load())
			}
			for i, r := range c3CheckPairs(t, a, 4) {
				if r.IsError != (i < 2) {
					t.Errorf("policy result %d error=%t", i, r.IsError)
				}
			}
		})
	}
	t.Run("disabled policy control", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "file")
		url, count := c3Peer(t, func(n int, request map[string]any) any {
			if n == 1 {
				return c3Reply(ensemble.Anthropic, c3ToolPart("bare", "write_file", map[string]any{"path": path, "content": "allowed"}))
			}
			data, err := os.ReadFile(path)
			if err != nil || string(data) != "allowed" {
				t.Errorf("disabled-policy effects=%q %v", data, err)
			}
			results := c3Results(request)
			if len(results) != 1 || results[0].id != "bare" {
				t.Errorf("control result=%v", results)
			}
			return c3Reply(ensemble.Anthropic, common.Text("Done without the policy."))
		})
		a := c3LocalAgent(t, url, ensemble.Anthropic, false, nil)
		if _, err := a.Ask(context.Background(), "Do the work."); err != nil || count.Load() != 2 {
			t.Fatalf("disabled-policy err=%v requests=%d", err, count.Load())
		}
		for _, r := range c3CheckPairs(t, a, 1) {
			if r.IsError {
				t.Fatal("disabled policy refused valid tool")
			}
		}
	})
}

func TestChapter3LoopBudgetAndFinalConsumption(t *testing.T) {
	for _, exhaust := range []bool{false, true} {
		t.Run(fmt.Sprintf("exhaust=%t", exhaust), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "file")
			url, count := c3Peer(t, func(n int, request map[string]any) any {
				if n > 17 {
					t.Errorf("unexpected request beyond result-consumption allowance: %d", n)
					return c3Reply(ensemble.Anthropic, common.Text("stop"))
				}
				if n > 1 {
					data, err := os.ReadFile(path)
					if err != nil || string(data) != strings.Repeat("x", n-1) {
						t.Errorf("batch %d effects=%q %v", n-1, data, err)
					}
					results := c3Results(request)
					if len(results) != 2*(n-1) || results[len(results)-1].id != fmt.Sprintf("read-%d", n-1) || !strings.Contains(results[len(results)-1].text, strings.Repeat("x", n-1)) {
						t.Errorf("batch %d result not consumed: %+v", n-1, results)
					}
				}
				if n == 17 && !exhaust {
					return c3Reply(ensemble.Anthropic, common.Text("Consumed the sixteenth result."))
				}
				return c3Reply(ensemble.Anthropic, common.Text("I will append one marker and read to verify this batch."), c3ToolPart(fmt.Sprintf("write-%d", n), "write_file", map[string]any{"path": path, "content": "x", "append": true}), c3ToolPart(fmt.Sprintf("read-%d", n), "read_file", map[string]any{"path": path}))
			})
			a := c3Agent(t, ensemble.Config{BaseURL: url, APIKey: "synthetic-local-fixture", Vendor: ensemble.Anthropic, RequireToolExplanation: true, MaxToolRounds: 16})
			reply, err := a.Ask(context.Background(), "Perform the bounded sequence.")
			if count.Load() != 17 || c3Contents(t, path) != strings.Repeat("x", 16) {
				t.Fatalf("budget requests=%d contents=%q", count.Load(), c3Contents(t, path))
			}
			want := 32
			if exhaust {
				want = 34
				if !errors.Is(err, ensemble.ErrToolRoundLimit) {
					t.Fatalf("missing budget error: %q %v", reply, err)
				}
			} else if err != nil || reply != "Consumed the sixteenth result." {
				t.Fatalf("last result not consumed: %q %v", reply, err)
			}
			for i, r := range c3CheckPairs(t, a, want) {
				if r.IsError != (i >= 32) {
					t.Errorf("budget result %d error=%t", i, r.IsError)
				}
			}
		})
	}
}

func TestChapter3LoopWaitsForWholeHTTPBody(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")
	bodySent := make(chan struct{})
	releaseTail := make(chan struct{})
	var requests atomic.Int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		if requests.Add(1) == 1 {
			response := c3Reply(ensemble.Anthropic, common.Text("I will write after the entire response returns."), c3ToolPart("held", "write_file", map[string]any{"path": path, "content": "complete"}))
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
			w.(http.Flusher).Flush()
			close(bodySent)
			<-releaseTail
			_, _ = io.WriteString(w, "\n")
			return
		}
		_ = json.NewEncoder(w).Encode(c3Reply(ensemble.Anthropic, common.Text("Complete body consumed.")))
	}))
	defer s.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	a := c3LocalAgent(t, s.URL, ensemble.Anthropic, true, nil)
	done := make(chan error, 1)
	go func() { _, err := a.Ask(ctx, "Write the fixture."); done <- err }()
	select {
	case <-bodySent:
	case <-ctx.Done():
		close(releaseTail)
		t.Fatal("fixture did not send body")
	}
	select {
	case err := <-done:
		close(releaseTail)
		t.Fatalf("Ask returned before HTTP body ended: %v", err)
	case <-time.After(30 * time.Millisecond):
	}
	_, statErr := os.Stat(path)
	close(releaseTail)
	if !os.IsNotExist(statErr) {
		t.Errorf("effect occurred before response EOF: %v", statErr)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if c3Contents(t, path) != "complete" || requests.Load() != 2 {
		t.Fatal("whole-body continuation failed")
	}
	c3CheckPairs(t, a, 1)
}
