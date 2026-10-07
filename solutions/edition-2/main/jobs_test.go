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
			if !strings.Contains(string(request), `"id":"call-8-1"`) {
				t.Error("continuation did not use committed response ID", string(request))
			}
			fmt.Fprint(w, `{"modelVersion":"test","candidates":[{"content":{"parts":[{"text":"finished"}]}}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1}}`)
		}
	}))
	defer server.Close()
	app := New(nil)
	defer app.Close()
	dir := t.TempDir()
	a, err := app.NewAgent(Config{Vendor: "gemini", APIKey: "test", Model: "test", BaseURL: server.URL, Workspace: dir, LogPath: filepath.Join(dir, "log"), Builtins: []string{"run_command", "read_file"}})
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	sequences := []uint64{}
	app.Subscribe(a.ID(), callbackObserver(func(o Observation) {
		mu.Lock()
		sequences = append(sequences, o.Seq)
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
	replay, err := New(nil).Load(filepath.Join(dir, "log"), Config{Model: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = replay.Render(Config{Model: "test"}); err != nil {
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
		a, err := app.NewAgent(Config{APIKey: "test", Model: "test", BaseURL: server.URL, Workspace: dir, LogPath: filepath.Join(dir, fmt.Sprint("log", i)), Builtins: []string{"read_file"}})
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
	a, err := second.NewAgent(Config{APIKey: "test", Model: "test", BaseURL: server.URL, Workspace: dir, LogPath: filepath.Join(dir, "restart"), Builtins: []string{"read_file"}})
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
