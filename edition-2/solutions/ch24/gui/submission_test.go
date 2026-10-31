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
	cfg := ensemble.Config{Model: "fake", BaseURL: provider.URL, DataDir: dir, SavePath: filepath.Join(dir, "save.json"), BuiltinTools: true}
	spec := ensemble.AgentSpec{SubmitSchema: json.RawMessage(`{"type":"object","properties":{"memory":{"type":"string"},"facts":{"type":"array"}},"required":["memory","facts"],"additionalProperties":false}`)}
	for _, reload := range []bool{false, true} {
		app := ensemble.New(io.Discard)
		app.SetAgentFactory(func(parent ensemble.Agent, spec ensemble.AgentSpec) (ensemble.Agent, error) {
			return app.NewChildAgent(parent, parent.Config(), spec)
		})
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
			frames = until(t, conn, kind("turn_ended"))
		}
		if got := a.History().Context().Turn; got != "idle" {
			t.Errorf("reload=%v: submitted turn remains %s", reload, got)
		}
		if reload {
			state := frames[len(frames)-1]
			if state["to"] != "idle" {
				t.Errorf("completed replay state: %v", state)
			}
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

// A genuinely held provider request must still replay as in-flight. This guards
// against making the completed-submission control pass by forcing every view idle.
func TestStructuredSubmissionGUIRunningControl(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		close(entered)
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	defer provider.Close()
	// Release the peer even when an assertion fails before request cancellation.
	defer close(release)
	app := ensemble.New(io.Discard)
	a, err := app.NewAgent(ensemble.Config{Model: "fake", BaseURL: provider.URL, DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	hub, err := gui.New(app, a, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Close()
	server := httptest.NewServer(hub.Handler(""))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := a.Submit(ctx, "HELD_TASK")
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	conn := dial(t, server.URL)
	defer conn.Close()
	frames := until(t, conn, kind("state_changed"))
	if state := frames[len(frames)-1]; state["to"] != "in_flight" || state["from"] != "in_flight" {
		t.Errorf("held request reported complete: %v", state)
	}
	cancel()
	<-result
}
