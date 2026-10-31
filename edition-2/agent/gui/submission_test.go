package gui_test

import (
	"context"
	"encoding/json"
	gui "ensemble-gui-stub"
	"ensemble/ensemble"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// The optional GUI receives structured call arguments and the actual submitted
// result after a real turn and after reload. Rendering needs no new browser API.
func TestStructuredSubmissionGUIReplay(t *testing.T) {
	var requests atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		json.NewEncoder(w).Encode(map[string]any{"content": []any{map[string]any{"type": "tool_use", "id": "custom", "name": "submit", "input": map[string]any{"memory": "I kept <reasons>", "facts": []any{true, nil, 3}}}}, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
	}))
	defer provider.Close()
	dir := t.TempDir()
	cfg := ensemble.Config{Model: "fake", BaseURL: provider.URL, DataDir: dir, SavePath: filepath.Join(dir, "save.json"), BuiltinTools: true, Recall: ensemble.RecallSettings{Disabled: true}}
	spec := ensemble.AgentSpec{SubmitSchema: json.RawMessage(`{"type":"object","properties":{"memory":{"type":"string"},"facts":{"type":"array"}},"required":["memory","facts"],"additionalProperties":false}`)}
	for _, reload := range []bool{false, true} {
		app := ensemble.New(io.Discard)
		app.EnableFleet(make(selectionInbox, 1))
		a, err := app.NewAgent(cfg, spec)
		if err != nil {
			t.Fatal(err)
		}
		if reload {
			if err := a.Load(); err != nil {
				t.Fatal(err)
			}
		}
		hub, err := gui.New(app, a, io.Discard)
		if err != nil {
			t.Fatal(err)
		}
		server := httptest.NewServer(hub.Handler(""))
		conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/ws", nil)
		if err != nil {
			t.Fatal(err)
		}
		send(t, conn, "subscribe", "")
		frames := until(t, conn, kind("state_changed"))
		if !reload {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, err = a.Ask(ctx, "TASK")
			cancel()
			if err != nil {
				t.Fatal(err)
			}
			frames = until(t, conn, kind("tool_finished"))
		}
		call, result := false, false
		for _, frame := range frames {
			if frame["type"] == "part_final" {
				part, _ := frame["part"].(map[string]any)
				if part["name"] == "submit" {
					args, ok := part["args"].(map[string]any)
					call = ok && args["memory"] == "I kept <reasons>"
					facts, ok := args["facts"].([]any)
					call = call && ok && len(facts) == 3 && facts[0] == true && facts[1] == nil && facts[2] == float64(3)
				}
			}
			if frame["type"] == "tool_finished" {
				// Tool output deliberately remains text. Decode its JSON document
				// to catch both lost nested values and accidental double encoding.
				output, _ := frame["result"].(string)
				payload, ok := strings.CutPrefix(output, "submitted: ")
				var decoded any
				if ok && json.Unmarshal([]byte(payload), &decoded) == nil {
					want := map[string]any{"memory": "I kept <reasons>", "facts": []any{true, nil, float64(3)}}
					result = reflect.DeepEqual(decoded, want)
				}
			}
		}
		conn.Close()
		hub.Close()
		server.Close()
		if err := a.Save(); err != nil {
			t.Fatal(err)
		}
		if err := a.Shutdown(); err != nil {
			t.Fatal(err)
		}
		if !call || !result {
			t.Fatalf("reload=%v: call=%v result=%v frames=%v", reload, call, result, frames)
		}
	}
	if requests.Load() != 1 {
		t.Fatal("reload requested inference", requests.Load())
	}
}
