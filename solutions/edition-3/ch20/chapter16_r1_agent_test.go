package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"ensemble"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestCh16R1CompletedToolAdmission(t *testing.T) {
	for _, tc := range []struct {
		name, model, ending string
		ordinary            bool
		wantEffects         int32
	}{
		{"summary only selected model", "gpt-6.1-sol", "complete", false, 1},
		{"summary only course model", "gpt-ch19-course", "complete", false, 1},
		{"retained ordinary gate", "gpt-5.4", "complete", false, 0},
		{"retained explained turn", "gpt-5.4", "complete", true, 1},
		{"failed done items", "gpt-6.1-sol", "failed", false, 0},
		{"incomplete done items", "gpt-6.1-sol", "incomplete", false, 0},
		{"EOF after done", "gpt-6.1-sol", "eof", false, 0},
		{"partial done arguments", "gpt-6.1-sol", "partial", false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var requests, effects atomic.Int32
			peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/responses" || r.Header.Get("Authorization") != "Bearer dummy-r1" {
					t.Error("wrong transport route")
				}
				n := requests.Add(1)
				w.Header().Set("Content-Type", "text/event-stream")
				emit := func(v any) { b, _ := json.Marshal(v); fmt.Fprintf(w, "data: %s\n\n", b) }
				if n > 1 {
					emit(map[string]any{"type": "response.completed", "response": map[string]any{"status": "completed", "model": tc.model, "output": []any{map[string]any{"type": "message", "role": "assistant", "content": []any{map[string]string{"type": "output_text", "text": "Done."}}}}, "usage": map[string]int{"input_tokens": 3, "output_tokens": 1}}})
					return
				}
				emit(map[string]any{"type": "response.created", "response": map[string]any{"model": tc.model}})
				emit(map[string]any{"type": "response.reasoning_summary_text.delta", "output_index": 0, "summary_index": 0, "delta": "Public summary, not ordinary explanation."})
				emit(map[string]any{"type": "response.output_item.done", "output_index": 0, "item": map[string]any{"type": "reasoning", "id": "reason", "summary": []any{map[string]string{"type": "summary_text", "text": "Public summary, not ordinary explanation."}}, "encrypted_content": "R1-PRIVATE"}})
				index := 1
				if tc.ordinary {
					emit(map[string]any{"type": "response.output_text.delta", "output_index": index, "content_index": 0, "delta": "I will use the local test tool."})
					emit(map[string]any{"type": "response.output_item.done", "output_index": index, "item": map[string]any{"type": "message", "role": "assistant", "content": []any{map[string]string{"type": "output_text", "text": "I will use the local test tool."}}}})
					index++
				}
				emit(map[string]any{"type": "response.output_item.added", "output_index": index, "item": map[string]string{"type": "function_call", "id": "item", "call_id": "native-call", "name": "r1_local", "arguments": ""}})
				emit(map[string]any{"type": "response.function_call_arguments.delta", "output_index": index, "delta": `{"value":`})
				args := `{"value":"actual"}`
				if tc.ending == "partial" {
					args = `{"value":`
				}
				emit(map[string]any{"type": "response.output_item.done", "output_index": index, "item": map[string]string{"type": "function_call", "id": "item", "call_id": "native-call", "name": "r1_local", "arguments": args}})
				terminal := map[string]any{"status": "completed", "model": tc.model, "output": []any{}, "usage": map[string]int{"input_tokens": 7, "output_tokens": 3}}
				switch tc.ending {
				case "eof":
					return
				case "failed", "incomplete":
					terminal["status"] = tc.ending
					emit(map[string]any{"type": "response." + tc.ending, "response": terminal})
				default:
					emit(map[string]any{"type": "response.completed", "response": terminal})
				}
			}))
			defer peer.Close()
			app := ensemble.New(nil)
			defer app.Shutdown()
			a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: tc.model, Surface: ensemble.Responses, AuthMode: ensemble.AuthAPI, APIKey: "dummy-r1", BaseURL: peer.URL, BaseDir: t.TempDir(), DataDir: t.TempDir(), Tools: []string{}, RequireToolExplanation: true})
			if err != nil {
				t.Fatal(err)
			}
			err = a.RegisterExternalTool(ensemble.ToolDefinition{Name: "r1_local", Description: "Return a local test value", Parameters: json.RawMessage(`{"type":"object","properties":{"value":{"type":"string"}},"required":["value"]}`)}, func(ctx context.Context, call ensemble.ToolCall, args json.RawMessage) error {
				if !bytes.Equal(args, []byte(`{"value":"actual"}`)) {
					t.Errorf("wrong terminal args %s", args)
				}
				effects.Add(1)
				_, err := io.WriteString(call, "actual local result")
				return err
			})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
			defer cancel()
			result := <-a.Submit(ctx, "Use the local test tool.")
			if effects.Load() != tc.wantEffects {
				t.Fatalf("effects=%d result=%+v", effects.Load(), result)
			}
			wire, err := a.Render()
			if err != nil {
				t.Fatal(err)
			}
			if tc.ending == "complete" {
				if result.Err != nil || result.Text != "Done." || requests.Load() != 2 || a.Usage().Input != 10 || a.Usage().Output != 4 {
					t.Fatalf("wrong completed turn %+v requests=%d usage=%+v", result, requests.Load(), a.Usage())
				}
				if tc.wantEffects == 0 && !bytes.Contains(wire, []byte("No tools ran")) {
					t.Fatalf("refusal was not honestly recorded: %s", wire)
				}
				if tc.wantEffects == 1 && !bytes.Contains(wire, []byte("actual local result")) {
					t.Fatal("actual tool result missing")
				}
			} else if result.Err == nil || requests.Load() != 1 || bytes.Contains(wire, []byte("native-call")) || bytes.Contains(wire, []byte("R1-PRIVATE")) {
				t.Fatalf("failed stream committed/dispatched: %+v requests=%d", result, requests.Load())
			}
		})
	}
}
