package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"example.com/ensemble"
)

// Live mode is a public-library client. It never imports or wires internal
// components. The credential-safe launcher supplies credentials only via env.
func runLive() error {
	workspace := os.Getenv("ENSEMBLE_LIVE_WORKSPACE")
	if workspace == "" {
		return fmt.Errorf("ENSEMBLE_LIVE_WORKSPACE is required")
	}
	if err := os.MkdirAll(workspace, 0700); err != nil {
		return err
	}
	app := ensemble.New(os.Stderr)
	defer app.Close()
	agents := make([]*ensemble.Agent, 2)
	observers := make([]*observer, 2)
	auditObservers := make([]*observer, 2)
	for i := range agents {
		config := ensemble.Config{Vendor: os.Getenv("LLM_VENDOR"), Model: os.Getenv("LLM_MODEL"), ResolvedModel: os.Getenv("LLM_RESOLVED_MODEL"), APIKey: os.Getenv("LLM_API_KEY"), Workspace: workspace, LogPath: filepath.Join(filepath.Dir(workspace), fmt.Sprintf("agent-%d.log", i+1)), Builtins: []string{"run_command", "wait_for_job"}, MaxTokens: 1024}
		a, err := app.NewAgent(config)
		if err != nil {
			return err
		}
		agents[i] = a
		o := &observer{ended: make(chan struct{})}
		observers[i] = o
		if _, err = app.Subscribe(a.ID(), o); err != nil {
			return err
		}
		auditObservers[i] = &observer{ended: make(chan struct{})}
		if _, err = app.Subscribe(a.ID(), auditObservers[i]); err != nil {
			return err
		}
	}
	// Independent Agents submit concurrent requests to the same public owner.
	// Each model is explicitly asked for one ordinary job and no supervision.
	var work sync.WaitGroup
	failures := make(chan error, 2)
	for i, a := range agents {
		work.Add(1)
		go func(i int, a *ensemble.Agent) {
			defer work.Done()
			prompt := fmt.Sprintf("This is a bounded public-library Jobs test. Use run_command exactly once with command \"printf AGENT-%d-START; sleep 2; printf AGENT-%d-END\" and ai_callback_delay 0.05. Do not call wait_for_job or start any further command. Return the running handle you observed; the host will observe completion independently.", i+1, i+1)
			_, err := app.Submit(context.Background(), ensemble.ClientRequest{AgentID: a.ID(), Prompt: &prompt})
			failures <- err
		}(i, a)
	}
	work.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			return err
		}
	}
	handles := make([]uint64, 2)
	for i, a := range agents {
		select {
		case <-observers[i].ended:
		case <-time.After(15 * time.Second):
			return fmt.Errorf("Agent %d background completion not observed", i+1)
		}
		state := a.Snapshot()
		if len(state.Jobs) != 1 {
			return fmt.Errorf("Agent %d model created %d jobs; expected exactly one", i+1, len(state.Jobs))
		}
		for handle, job := range state.Jobs {
			handles[i] = handle
			if job.Status != "done" {
				return fmt.Errorf("background job did not finish")
			}
			data, err := os.ReadFile(filepath.Join(workspace, job.Output.Locator))
			if err != nil || !strings.Contains(string(data), fmt.Sprintf("AGENT-%d-END", i+1)) {
				return fmt.Errorf("Agent artifact missing")
			}
		}
	}
	if handles[0] == handles[1] {
		return fmt.Errorf("application handles collided")
	}
	foreignPrompt := fmt.Sprintf("Deliberately test isolation: call wait_for_job once with handle %d. That handle belongs to another Agent in this Ensemble. Do not run any command or try another handle. Report the actual permission/refusal result.", handles[1])
	if _, err := app.Submit(context.Background(), ensemble.ClientRequest{AgentID: agents[0].ID(), Prompt: &foreignPrompt}); err != nil {
		return err
	}
	refused := false
	for _, event := range agents[0].Events() {
		if event.Type == "tool_returned" && event.Tool.IsError && strings.Contains(*event.Tool.Parts[0].Text, "unavailable to this Agent") {
			refused = true
		}
	}
	if !refused {
		return fmt.Errorf("foreign-handle refusal was not exercised")
	}
	rows := []map[string]any{}
	for i, a := range agents {
		o := observers[i]
		o.mu.Lock()
		sequences := append([]uint64(nil), o.seq...)
		o.mu.Unlock()
		audit := auditObservers[i]
		audit.mu.Lock()
		badCopy := audit.sawMutation
		auditedCount := len(audit.seq)
		audit.mu.Unlock()
		if badCopy || auditedCount != len(sequences) {
			return fmt.Errorf("second observer did not receive independent complete observations")
		}
		for index, seq := range sequences {
			if seq != uint64(index+1) {
				return fmt.Errorf("Agent observation order changed")
			}
		}
		for _, event := range a.Events() {
			if event.Tool != nil && event.Tool.CallID == "consumer-owned-copy" {
				return fmt.Errorf("observer owns history memory")
			}
		}
		rows = append(rows, map[string]any{"agent_id": a.ID(), "handle": handles[i], "observed_sequences": sequences, "usage": a.Usage(), "jobs": a.Snapshot().Jobs})
	}
	if err := app.Close(); err != nil {
		return err
	}
	if err := app.Close(); err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"backend": "real provider via public Ensemble clients", "provider": os.Getenv("LLM_VENDOR"), "model": os.Getenv("LLM_MODEL"), "agents": rows, "foreign_handle": "refused", "shared_workspace": "distinct retained artifacts", "observations": "ordered, owned snapshots; background terminal event delivered", "close": "idempotent"})
}
