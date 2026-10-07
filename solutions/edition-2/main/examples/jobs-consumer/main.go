// A headless consumer exercises only the public library and observer seam.
package main

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
	"time"

	"example.com/ensemble"
)

type observer struct {
	mu    sync.Mutex
	seq   []uint64
	ended chan struct{}
}

func (o *observer) Observe(event ensemble.Observation) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.seq = append(o.seq, event.Seq)
	if event.Kind == "job_ended" {
		close(o.ended)
	}
	// Mutating a delivered snapshot must not rewrite the other client or history.
	if event.Event.Tool != nil {
		event.Event.Tool.CallID = "consumer-owned-copy"
	}
}
func run() error {
	workspace, err := os.MkdirTemp("", "ensemble-public-jobs-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workspace)
	counts := map[string]int{}
	var serverMu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Model string `json:"model"`
		}
		json.NewDecoder(r.Body).Decode(&request)
		serverMu.Lock()
		counts[request.Model]++
		n := counts[request.Model]
		serverMu.Unlock()
		if n == 1 {
			fmt.Fprint(w, `{"content":[{"type":"tool_use","id":"process","name":"run_command","input":{"command":"sleep 0.2; printf completed","ai_callback_delay":0.01}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		} else if n == 3 {
			fmt.Fprint(w, `{"content":[{"type":"tool_use","id":"foreign","name":"wait_for_job","input":{"handle":2}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		} else {
			fmt.Fprint(w, `{"content":[{"type":"text","text":"reported"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		}
	}))
	defer server.Close()
	app := ensemble.New(os.Stderr)
	defer app.Close()
	agents := []*ensemble.Agent{}
	observers := []*observer{}
	for i := 0; i < 2; i++ {
		a, err := app.NewAgent(ensemble.Config{APIKey: "local-fixture", Model: fmt.Sprint("model", i), BaseURL: server.URL, Workspace: workspace, LogPath: filepath.Join(workspace, fmt.Sprint("log", i)), Builtins: []string{"run_command", "wait_for_job"}})
		if err != nil {
			return err
		}
		o := &observer{ended: make(chan struct{})}
		if _, err = app.Subscribe(a.ID(), o); err != nil {
			return err
		}
		agents = append(agents, a)
		observers = append(observers, o)
		prompt := "start retained work"
		if _, err = app.Submit(context.Background(), ensemble.ClientRequest{AgentID: a.ID(), Prompt: &prompt}); err != nil {
			return err
		}
	}
	for i, o := range observers {
		select {
		case <-o.ended:
		case <-time.After(3 * time.Second):
			return fmt.Errorf("Agent %d background completion not observed", i)
		}
		o.mu.Lock()
		for index, seq := range o.seq {
			if seq != uint64(index+1) {
				o.mu.Unlock()
				return fmt.Errorf("out-of-order observations")
			}
		}
		o.mu.Unlock()
	}
	if _, err = agents[0].Prompt(context.Background(), "try foreign handle"); err != nil {
		return err
	}
	denied := false
	for _, e := range agents[0].Events() {
		if e.Tool != nil && e.Tool.CallID == "consumer-owned-copy" {
			return fmt.Errorf("observer mutated history")
		}
		if e.Type == "tool_returned" && e.Tool.CallID == "foreign" {
			denied = e.Tool.IsError && strings.Contains(*e.Tool.Parts[0].Text, "unavailable")
		}
	}
	if !denied {
		return fmt.Errorf("foreign handle not refused")
	}
	for i, a := range agents {
		handle := uint64(i + 1)
		snapshot := a.Snapshot()
		job, ok := snapshot.Jobs[handle]
		if !ok || len(snapshot.Jobs) != 1 || job.Status != "done" {
			return fmt.Errorf("Agent job facts mixed")
		}
		data, err := os.ReadFile(filepath.Join(workspace, job.Output.Locator))
		if err != nil || !strings.Contains(string(data), "completed") {
			return fmt.Errorf("artifact missing")
		}
	}
	if err = app.Close(); err != nil {
		return err
	}
	if err = app.Close(); err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"agents": 2, "handles": []int{1, 2}, "shared_workspace_artifacts": "distinct", "foreign_handle": "refused", "background_observations": "ordered and owned", "close": "idempotent", "backend": "local deterministic HTTP fixture, not paid live model"})
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
