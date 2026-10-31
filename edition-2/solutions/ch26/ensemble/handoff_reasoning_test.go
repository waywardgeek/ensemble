package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"ensemble/ensemble"
	"ensemble/internal/common"
)

// A real tool loop must replay reasoning before a checkpoint, clear both older
// and same-batch reasoning at it, and retain reasoning produced afterward.
func TestMicroHandoffClearsReasoning(t *testing.T) {
	for name, vendor := range map[string]ensemble.Vendor{"Responses": ensemble.OpenAI, "Gemini": ensemble.Gemini} {
		t.Run(name, func(t *testing.T) { microHandoffReasoning(t, vendor) })
	}
}

func microHandoffReasoning(t *testing.T, vendor ensemble.Vendor) {
	model := "gpt-ch19-course"
	if vendor == ensemble.Gemini {
		model = "gemini-3.8-flash"
	}
	var mu sync.Mutex
	var requests [][]byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		mu.Lock()
		requests = append(requests, raw)
		n := len(requests)
		mu.Unlock()
		output := []any{
			map[string]any{"type": "reasoning", "encrypted_content": "after-checkpoint"},
			map[string]any{"type": "message", "content": []any{map[string]any{"type": "output_text", "text": "Finished."}}},
		}
		switch n {
		case 1:
			output = []any{
				map[string]any{"type": "reasoning", "encrypted_content": "before-checkpoint"},
				map[string]any{"type": "message", "content": []any{map[string]any{"type": "output_text", "text": "I will inspect the first result."}}},
				map[string]any{"type": "function_call", "call_id": "first", "name": "echo", "arguments": "{}"},
			}
		case 2:
			output = []any{
				map[string]any{"type": "reasoning", "encrypted_content": "checkpoint-batch"},
				map[string]any{"type": "message", "content": []any{map[string]any{"type": "output_text", "text": "I will checkpoint and finish the sibling tool."}}},
				map[string]any{"type": "function_call", "call_id": "handoff", "name": "micro_handoff", "arguments": `{"text":"Keep this checkpoint note."}`},
				map[string]any{"type": "function_call", "call_id": "sibling", "name": "echo", "arguments": "{}"},
			}
		}
		if vendor == ensemble.Gemini {
			// Gemini binds signatures to visible text and calls too. Supply each
			// real wire shape; removing only standalone thought blocks is not enough.
			var parts []any
			signature := output[0].(map[string]any)["encrypted_content"].(string)
			for _, raw := range output {
				item := raw.(map[string]any)
				switch item["type"] {
				case "reasoning":
					parts = append(parts, map[string]any{"thought": true, "text": "Private thought", "thoughtSignature": signature})
				case "message":
					content := item["content"].([]any)[0].(map[string]any)
					parts = append(parts, map[string]any{"text": content["text"], "thoughtSignature": "text-" + signature})
				case "function_call":
					parts = append(parts, map[string]any{"functionCall": map[string]any{"id": item["call_id"], "name": item["name"], "args": json.RawMessage(item["arguments"].(string))}, "thoughtSignature": "call-" + signature})
				}
			}
			json.NewEncoder(w).Encode(map[string]any{"candidates": []any{map[string]any{"content": map[string]any{"role": "model", "parts": parts}}}, "usageMetadata": map[string]int{"promptTokenCount": 10, "candidatesTokenCount": 1}})
			return
		}
		frame(w, wireFrame{"response.completed", map[string]any{"response": map[string]any{"status": "completed", "model": model, "output": output, "usage": map[string]int{"input_tokens": 10, "output_tokens": 1}}}})
	}))
	defer server.Close()
	cfg := ensemble.Config{Vendor: vendor, Model: model, BaseURL: server.URL, BuiltinTools: true, DataDir: t.TempDir()}
	a, err := ensemble.New(io.Discard).NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	if err := a.RegisterTool(ensemble.ToolDeclaration{Name: "echo", Schema: json.RawMessage(`{"type":"object"}`)}, func(ensemble.ToolContext, json.RawMessage) (string, error) {
		return "actual tool result", nil
	}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := a.Ask(ctx, "Preserve this human request."); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(requests) != 3 {
		t.Fatalf("requests=%d, want 3", len(requests))
	}
	if !bytes.Contains(requests[1], []byte("before-checkpoint")) {
		t.Fatal("ordinary tool continuation lost reasoning before the handoff")
	}
	if vendor == ensemble.Gemini && !bytes.Contains(requests[1], []byte("text-before-checkpoint")) {
		t.Fatal("ordinary tool continuation lost Gemini's text-bound signature")
	}
	for _, unwanted := range []string{"before-checkpoint", "checkpoint-batch", `"type":"reasoning"`, `"type":"function_call"`, `"type":"function_call_output"`, `"thoughtSignature"`, `"functionCall"`, `"functionResponse"`} {
		if bytes.Contains(requests[2], []byte(unwanted)) {
			t.Errorf("post-handoff request retained %s", unwanted)
		}
	}
	for _, wanted := range []string{"Preserve this human request.", "I will inspect the first result.", "I will checkpoint and finish the sibling tool.", "Keep this checkpoint note."} {
		if !bytes.Contains(requests[2], []byte(wanted)) {
			t.Errorf("post-handoff request lost %q", wanted)
		}
	}
	if bytes.Count(requests[2], []byte("micro_handoff completed")) != 1 {
		t.Error("post-handoff request must identify the completed checkpoint once")
	}
	var log, save bytes.Buffer
	siblingRan := false
	for _, event := range a.History().EventsAfter(0) {
		if err := json.NewEncoder(&log).Encode(event); err != nil {
			t.Fatal(err)
		}
		if event.Type == "tool_returned" && event.Tool.CallID == "sibling" && !event.Tool.IsError {
			siblingRan = true
		}
	}
	if !siblingRan || !bytes.Contains(log.Bytes(), []byte("before-checkpoint")) || !bytes.Contains(log.Bytes(), []byte("checkpoint-batch")) {
		t.Fatal("checkpoint lost original event bytes or prevented its sibling tool")
	}
	if err := a.History().Save(&save); err != nil {
		t.Fatal(err)
	}
	// Recovery must apply the same cut, without erasing reasoning generated by
	// the next response. Rendering here inspects the actual next vendor payload.
	for _, restore := range []func() error{
		func() error { return nil },
		func() error { return a.History().Load(bytes.NewReader(log.Bytes())) },
		func() error { return a.History().Restore(bytes.NewReader(save.Bytes())) },
	} {
		if err := restore(); err != nil {
			t.Fatal(err)
		}
		current := a.History().Context()
		before, err := json.Marshal(current)
		if err != nil {
			t.Fatal(err)
		}
		// A stored handoff is still the original note, including in snapshots
		// produced before the completion label was added to request rendering.
		for _, entry := range current.Dialogue {
			if entry.Kind == common.KindHandoff && entry.Parts[0].Text != "Keep this checkpoint note." {
				t.Fatal("completion label changed the stored checkpoint note")
			}
		}
		raw, err := a.Engine().Render(current)
		if err != nil {
			t.Fatal(err)
		}
		again, err := a.Engine().Render(current)
		after, marshalErr := json.Marshal(current)
		if err != nil || marshalErr != nil || !bytes.Equal(raw, again) || !bytes.Equal(before, after) || bytes.Count(raw, []byte("micro_handoff completed")) != 1 {
			t.Fatal("render must label the completed checkpoint once without changing context")
		}
		if bytes.Contains(raw, []byte("before-checkpoint")) || bytes.Contains(raw, []byte("checkpoint-batch")) || !bytes.Contains(raw, []byte("after-checkpoint")) {
			t.Fatal("live or restored context has the wrong checkpoint reasoning boundary")
		}
		if vendor == ensemble.Gemini && !bytes.Contains(raw, []byte("text-after-checkpoint")) {
			t.Fatal("live or restored context lost Gemini's new text-bound signature")
		}
	}
}
