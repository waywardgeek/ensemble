package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
)

// SetTrace is best effort and Shutdown does not join its writer. The test
// owns synchronization and receipt of the continuation-bearing request trace.
type ch16TraceCapture struct {
	mu       sync.Mutex
	buf      bytes.Buffer
	replayed chan struct{}
}

func (w *ch16TraceCapture) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n, err := w.buf.Write(p)
	var record ensemble.TraceRecord
	if json.Unmarshal(p, &record) == nil && record.Stage == "request_body" && strings.Contains(record.Payload, `"type":"function_call_output"`) {
		select {
		case w.replayed <- struct{}{}:
		default:
		}
	}
	return n, err
}

func (w *ch16TraceCapture) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

func TestCh16ResponsesAgentToolRecovery(t *testing.T) {
	var calls atomic.Int32
	private := "encrypted-local-continuation-not-for-diagnostics"
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		body, _ := io.ReadAll(r.Body)
		var request map[string]json.RawMessage
		if err := json.Unmarshal(body, &request); err != nil {
			t.Error(err)
		}
		if r.URL.Path != "/chosen/responses" || r.Header.Get("Authorization") != "Bearer dummy-api" || r.Header.Get("ChatGPT-Account-Id") != "" {
			t.Errorf("wrong transport route %s", r.URL.Path)
		}
		if len(request["messages"]) > 0 || string(request["stream"]) != "true" || string(request["store"]) != "false" || !bytes.Contains(request["reasoning"], []byte(`"effort":"high"`)) || !bytes.Contains(request["reasoning"], []byte(`"summary":"auto"`)) {
			t.Errorf("wrong Responses envelope: %s", body)
		}
		output := []any{}
		if n == 1 {
			output = []any{map[string]any{"type": "reasoning", "id": "reasoning-1", "summary": []any{map[string]string{"type": "summary_text", "text": "Public plan."}}, "encrypted_content": private}, map[string]any{"type": "message", "role": "assistant", "content": []any{map[string]any{"type": "output_text", "text": "I will run the local tool to obtain the requested value."}}}, map[string]any{"type": "function_call", "id": "function-1", "call_id": "call-1", "name": "local_echo", "arguments": `{"value":"hello"}`}}
		} else {
			if !bytes.Contains(body, []byte(private)) || !bytes.Contains(body, []byte(`"type":"function_call_output"`)) || !bytes.Contains(body, []byte("local tool result")) {
				t.Errorf("continuation/tool output not replayed: %s", body)
			}
			output = []any{map[string]any{"type": "message", "role": "assistant", "status": "completed", "content": []any{map[string]string{"type": "output_text", "text": "Finished."}}}}
		}
		response := map[string]any{"id": fmt.Sprintf("response-%d", n), "status": "completed", "model": "gpt-ch19-course", "output": output, "usage": map[string]any{"input_tokens": 12, "output_tokens": 4, "input_tokens_details": map[string]int{"cached_tokens": 7}, "output_tokens_details": map[string]int{"reasoning_tokens": 2}}}
		w.Header().Set("Content-Type", "text/event-stream")
		emit := func(v any) { b, _ := json.Marshal(v); fmt.Fprintf(w, "data: %s\n\n", b) }
		if n == 1 {
			emit(map[string]any{"type": "response.reasoning_summary_text.delta", "output_index": 0, "summary_index": 0, "delta": "Public plan."})
			emit(map[string]any{"type": "response.output_text.delta", "output_index": 1, "content_index": 0, "delta": "I will run the local tool to obtain the requested value."})
			emit(map[string]any{"type": "response.output_item.added", "output_index": 2, "item": map[string]any{"type": "function_call", "call_id": "call-1", "name": "local_echo", "arguments": ""}})
			emit(map[string]any{"type": "response.function_call_arguments.delta", "output_index": 2, "delta": `{"value":"hello"}`})
		} else {
			emit(map[string]any{"type": "response.output_text.delta", "output_index": 0, "content_index": 0, "delta": "Finished."})
		}
		emit(map[string]any{"type": "response.completed", "response": response})
	}))
	defer peer.Close()
	root := ensemble.New(nil)
	trace := &ch16TraceCapture{replayed: make(chan struct{}, 1)}
	root.SetTrace(trace)
	dir := t.TempDir()
	cfg := ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-ch19-course", ReasoningEffort: "high", AuthMode: ensemble.AuthAPI, APIKey: "dummy-api", BaseURL: peer.URL + "/chosen/responses", BaseDir: dir, SavePath: filepath.Join(dir, "recovery.json"), LogPath: filepath.Join(dir, "history.jsonl"), Tools: []string{}}
	a, err := root.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var executions atomic.Int32
	err = a.RegisterExternalTool(ensemble.ToolDefinition{Name: "local_echo", Description: "Return a local test value", Parameters: json.RawMessage(`{"type":"object","properties":{"value":{"type":"string"}},"required":["value"]}`)}, func(ctx context.Context, call ensemble.ToolCall, args json.RawMessage) error {
		executions.Add(1)
		_, err := io.WriteString(call, "local tool result")
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	result := <-a.Submit(ctx, "Use the local tool.")
	if result.Err != nil || result.Text != "Finished." || calls.Load() != 2 || executions.Load() != 1 {
		t.Fatalf("result=%+v requests=%d executions=%d", result, calls.Load(), executions.Load())
	}
	if got := a.Usage(); got.Input != 10 || got.CacheRead != 14 || got.Output != 8 {
		t.Fatalf("usage double-counted: %+v", got)
	}
	if err := a.Save(); err != nil {
		t.Fatal(err)
	}
	saved, err := os.ReadFile(cfg.SavePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(saved, []byte(private)) || !bytes.Contains(saved, []byte("reasoning_summary")) || bytes.Contains(saved, []byte("dummy-api")) || bytes.Contains(saved, []byte(peer.URL)) {
		t.Fatalf("recovery privacy/retention failure")
	}
	select {
	case <-trace.replayed:
	case <-time.After(2 * time.Second):
		t.Fatal("continuation-bearing request diagnostic was not received")
	}
	if err := root.Shutdown(); err != nil {
		t.Fatal(err)
	}
	diagnostic := trace.String()
	if strings.Contains(diagnostic, private) || strings.Contains(diagnostic, "dummy-api") {
		t.Fatal("private continuation/credential leaked to diagnostics")
	}
	reopened := ensemble.New(nil)
	defer reopened.Shutdown()
	b, err := reopened.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	wire, err := b.Render()
	if err != nil || !bytes.Contains(wire, []byte(private)) {
		t.Fatal("same-source continuation lost on recovery", err)
	}
	other := cfg
	other.Model = "gpt-5.4"
	other.SavePath = filepath.Join(dir, "other-recovery.json")
	if err := os.WriteFile(other.SavePath, saved, 0600); err != nil {
		t.Fatal(err)
	}
	other.Surface = ensemble.Responses
	other.LogPath = filepath.Join(dir, "other-history.jsonl")
	c, err := reopened.NewAgent(other)
	if err != nil {
		t.Fatal(err)
	}
	wire, err = c.Render()
	if err != nil || bytes.Contains(wire, []byte(private)) || !bytes.Contains(wire, []byte("Finished.")) || !bytes.Contains(wire, []byte("Public plan.")) {
		t.Fatal("cross-model replay privacy/text loss", err)
	}
}
func TestCh16ConfigurationContract(t *testing.T) {
	cases := []struct {
		name, model, surface, stream string
		want                         ensemble.Surface
		wantStream                   bool
		reasoning                    bool
	}{{"codex", "gpt-ch19-course", "", "", ensemble.Responses, true, true}, {"legacy", "gpt-5.2", "", "", ensemble.ChatCompletions, false, false}, {"override", "gpt-ch19-course", "responses", "sse", ensemble.Responses, true, true}, {"unknown", "unknown", "responses", "sse", ensemble.Responses, true, false}, {"plain", "gpt-ch19-course", "", "off", ensemble.Responses, false, true}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			values := map[string]string{"LLM_VENDOR": "openai", "LLM_MODEL": tc.model, "OPENAI_BILLING_MODE": "api", "OPENAI_API_KEY": "dummy-api", "LLM_SURFACE": tc.surface, "LLM_STREAM_CAPABILITY": tc.stream, "OPENAI_REASONING_EFFORT": "high", "OPENAI_REASONING_SUMMARY": "auto", "LLM_BASE_URL": "http://127.0.0.1:1/override/responses"}
			cfg, err := ensemble.ConfigFromEnv(func(k string) string { return values[k] })
			if err != nil {
				t.Fatal(err)
			}
			cfg.BaseDir = t.TempDir()
			root := ensemble.New(nil)
			defer root.Shutdown()
			a, err := root.NewAgent(cfg)
			if err != nil {
				t.Fatal(err)
			}
			actual := a.Config()
			if actual.Surface != tc.want || actual.Streaming() != tc.wantStream || actual.BaseURL != values["LLM_BASE_URL"] {
				t.Fatalf("wrong config %+v", actual)
			}
			wire, err := a.Render()
			if err != nil {
				t.Fatal(err)
			}
			var body map[string]json.RawMessage
			_ = json.Unmarshal(wire, &body)
			if (len(body["reasoning"]) > 0) != tc.reasoning {
				t.Fatalf("capability guessed or lost: %s", wire)
			}
		})
	}
	if _, err := ensemble.ConfigFromEnv(func(k string) string {
		return map[string]string{"LLM_VENDOR": "anthropic", "OPENAI_BILLING_MODE": "chatgpt"}[k]
	}); err == nil {
		t.Fatal("contradictory provider accepted")
	}
}

func TestCh16DefaultResponsesEndpoint(t *testing.T) {
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-ch19-course", AuthMode: ensemble.AuthAPI, APIKey: "dummy-api", BaseDir: t.TempDir(), DataDir: t.TempDir(), Tools: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	if a.Config().BaseURL != "https://api.openai.com" || a.Config().Surface != ensemble.Responses {
		t.Fatal("invalid default public Responses route", a.Config().BaseURL)
	}
}
