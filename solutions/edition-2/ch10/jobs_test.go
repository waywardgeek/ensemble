package ensemble

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestBackgroundCompletionFinalizesActualResponseIdentity(t *testing.T) {
	ended := make(chan struct{})
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		switch count {
		case 1:
			fmt.Fprint(w, `{"modelVersion":"test","candidates":[{"content":{"parts":[{"functionCall":{"id":"supplied","name":"run_command","args":{"command":"sleep 0.15; printf late-output","ai_callback_delay":0.01}}}]}}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1}}`)
		case 2:
			select {
			case <-ended:
			case <-time.After(3 * time.Second):
				t.Error("completion could not publish during HTTP")
			}
			fmt.Fprint(w, `{"modelVersion":"test","candidates":[{"content":{"parts":[{"text":"next"},{"functionCall":{"name":"read_file","args":{"path":"missing"}}}]}}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1}}`)
		default:
			var request json.RawMessage
			json.NewDecoder(r.Body).Decode(&request)
			if !strings.Contains(string(request), `"id":"call-9-1"`) {
				t.Error("continuation did not use committed response ID", string(request))
			}
			fmt.Fprint(w, `{"modelVersion":"test","candidates":[{"content":{"parts":[{"text":"finished"}]}}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1}}`)
		}
	}))
	defer server.Close()
	app := New(nil)
	defer app.Close()
	dir := t.TempDir()
	a, err := app.NewAgent(Config{DisableStreaming: true, Vendor: "gemini", APIKey: "test", Model: "test", BaseURL: server.URL, Workspace: dir, LogPath: filepath.Join(dir, "log"), Builtins: []string{"run_command", "read_file"}})
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	sequences := []uint64{}
	app.Subscribe(a.ID(), callbackObserver(func(o Observation) {
		mu.Lock()
		if o.Event.Seq != 0 {
			sequences = append(sequences, o.Seq)
		}
		mu.Unlock()
		if o.Kind == "job_ended" && o.Event.Job.Handle == 1 {
			close(ended)
		}
	}))
	if _, err = a.Prompt(context.Background(), "go"); err != nil {
		t.Fatal(err)
	}
	for _, e := range a.Events() {
		if e.Response != nil {
			for i, p := range e.Response.Parts {
				if p.Type == "tool_call" && p.CallID != "supplied" && p.CallID != fmt.Sprintf("call-%d-%d", e.Seq, i) {
					t.Fatal(e.Seq, p.CallID)
				}
			}
		}
	}
	mu.Lock()
	defer mu.Unlock()
	for i, seq := range sequences {
		if seq != uint64(i+1) {
			t.Fatal(sequences)
		}
	}
	data, _ := os.ReadFile(filepath.Join(dir, "cr/io/1"))
	if !strings.Contains(string(data), "late-output\nexit_code: 0") {
		t.Fatal(string(data))
	}
	// Replay is facts-only, even after artifacts have vanished.
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	os.RemoveAll(filepath.Join(dir, "cr"))
	replay, err := New(nil).Load(filepath.Join(dir, "log"), Config{DisableStreaming: true, Model: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = replay.Render(Config{DisableStreaming: true, Model: "test"}); err != nil {
		t.Fatal(err)
	}
}

func TestPublicAgentHandlesAndWorkspaceIsolation(t *testing.T) {
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		if count%2 == 1 {
			fmt.Fprint(w, `{"content":[{"type":"tool_use","id":"read","name":"read_file","input":{"path":"note"}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		} else {
			fmt.Fprint(w, `{"content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		}
	}))
	defer server.Close()
	app := New(nil)
	defer app.Close()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "note"), []byte("shared\n"), 0600)
	for i := 1; i <= 2; i++ {
		a, err := app.NewAgent(Config{DisableStreaming: true, APIKey: "test", Model: "test", BaseURL: server.URL, Workspace: dir, LogPath: filepath.Join(dir, fmt.Sprint("log", i)), Builtins: []string{"read_file"}})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = a.Prompt(context.Background(), "read"); err != nil {
			t.Fatal(err)
		}
		state := a.Snapshot()
		if _, ok := state.Jobs[uint64(i)]; !ok {
			t.Fatal(state.Jobs)
		}
		data, _ := os.ReadFile(filepath.Join(dir, fmt.Sprint("cr/io/", i)))
		if string(data) != "shared\n" {
			t.Fatal(string(data))
		}
	}
	// A fresh application skips retained names; it cannot claim the old jobs.
	second := New(nil)
	defer second.Close()
	a, err := second.NewAgent(Config{DisableStreaming: true, APIKey: "test", Model: "test", BaseURL: server.URL, Workspace: dir, LogPath: filepath.Join(dir, "restart"), Builtins: []string{"read_file"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.Prompt(context.Background(), "read"); err != nil {
		t.Fatal(err)
	}
	if _, ok := a.Snapshot().Jobs[3]; !ok {
		t.Fatal("retained names not skipped", a.Snapshot().Jobs)
	}
	if len(a.Snapshot().Jobs) != 1 {
		t.Fatal("old jobs restored")
	}
	data, _ := os.ReadFile(filepath.Join(dir, "cr/io/1"))
	if string(data) != "shared\n" {
		t.Fatal("artifact changed", string(data))
	}
}

func TestWireLimitsConsumeEveryAttempt(t *testing.T) {
	calls := []struct{ id, name, args string }{
		{"set1", "tool_limits", `{"ai_callback_delay":0.2,"max_output_bytes":4}`},
		{"unknown", "unavailable", `{}`},
		{"default", "run_command", `{"command":"echo DEFAULT-CONTROL"}`},
		{"set2", "tool_limits", `{"ai_callback_delay":0.2}`},
		{"set3", "tool_limits", `{"max_output_bytes":4}`},
		{"override", "run_command", `{"command":"echo EXPLICIT-CONTROL","ai_callback_delay":1,"max_output_bytes":100}`},
		{"set4", "tool_limits", `{"max_output_bytes":4}`},
		{"invalid", "run_command", `{"command":"touch forbidden","ai_callback_pattern":"["}`},
		{"after", "run_command", `{"command":"echo AFTER-CONTROL"}`},
	}
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		if count > 1 {
			fmt.Fprint(w, `{"content":[{"type":"text","text":"complete"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
			return
		}
		content := []map[string]any{}
		for _, c := range calls {
			content = append(content, map[string]any{"type": "tool_use", "id": c.id, "name": c.name, "input": json.RawMessage(c.args)})
		}
		json.NewEncoder(w).Encode(map[string]any{"content": content, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
	}))
	defer server.Close()
	dir := t.TempDir()
	app := New(nil)
	defer app.Close()
	a, err := app.NewAgent(Config{DisableStreaming: true, APIKey: "test", Model: "test", BaseURL: server.URL, Workspace: dir, LogPath: filepath.Join(dir, "log"), Builtins: []string{"run_command", "tool_limits"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.Prompt(context.Background(), "exercise limits"); err != nil {
		t.Fatal(err)
	}
	results := map[string]ToolEvent{}
	for _, e := range a.Events() {
		if e.Type == "tool_returned" {
			results[e.Tool.CallID] = *e.Tool
		}
	}
	if !results["unknown"].IsError || !strings.Contains(*results["unknown"].Parts[0].Text, "tool_limits consumed by unavailable") {
		t.Fatal(results["unknown"])
	}
	if strings.Contains(*results["default"].Parts[0].Text, "tool_limits") || !strings.Contains(*results["default"].Parts[0].Text, "DEFAULT-CONTROL") {
		t.Fatal(results["default"])
	}
	if !strings.Contains(*results["set3"].Parts[0].Text, "tool_limits consumed by tool_limits") {
		t.Fatal(results["set3"])
	}
	if got := *results["override"].Parts[0].Text; !strings.Contains(got, "ai_callback_delay=1") || !strings.Contains(got, "max_output_bytes=100") || !strings.Contains(got, "EXPLICIT-CONTROL") {
		t.Fatal(got)
	}
	if !results["invalid"].IsError || results["invalid"].Job != nil || !strings.Contains(*results["invalid"].Parts[0].Text, "validation refused") {
		t.Fatal(results["invalid"])
	}
	if strings.Contains(*results["after"].Parts[0].Text, "tool_limits") {
		t.Fatal(results["after"])
	}
	if _, err = os.Stat(filepath.Join(dir, "forbidden")); !os.IsNotExist(err) {
		t.Fatal("invalid limit executed")
	}
}

func TestReadSourceCapIsReportMetadata(t *testing.T) {
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		if count == 1 {
			fmt.Fprint(w, `{"content":[{"type":"tool_use","id":"read","name":"read_file","input":{"path":"source","max_bytes":4}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		} else {
			fmt.Fprint(w, `{"content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		}
	}))
	defer server.Close()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "source"), []byte("ABCDEFGHIJ"), 0600)
	app := New(nil)
	defer app.Close()
	a, err := app.NewAgent(Config{DisableStreaming: true, APIKey: "test", Model: "test", BaseURL: server.URL, Workspace: dir, LogPath: filepath.Join(dir, "log"), Builtins: []string{"read_file"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.Prompt(context.Background(), "read capped source"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "cr/io/1"))
	if err != nil || string(data) != "ABCD" {
		t.Fatal(string(data), err)
	}
	if a.Snapshot().Jobs[1].Bytes != 4 {
		t.Fatal(a.Snapshot().Jobs)
	}
	found := false
	for _, e := range a.Events() {
		if e.Type == "tool_returned" {
			text := *e.Tool.Parts[0].Text
			found = strings.Contains(text, "ABCD") && strings.Contains(text, "max_bytes limit reached")
		}
	}
	if !found {
		t.Fatal("source truncation metadata disappeared")
	}
}
