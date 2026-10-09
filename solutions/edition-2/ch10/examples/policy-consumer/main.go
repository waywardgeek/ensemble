// A headless application uses separate policy owners and proves restart without
// importing the optional GUI. Environment configuration comes from the public CLI.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"example.com/ensemble"
	"example.com/ensemble/cli"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	app := ensemble.New(os.Stderr)
	defer app.Close()
	config := cli.Configuration(app)
	config.MaxTokens = 1024
	config.Builtins = []string{"read_file"}
	configs := make([]ensemble.Config, 2)
	agents := make([]*ensemble.Agent, 2)
	for i, name := range []string{"first", "second"} {
		c := config
		c.LogPath = filepath.Join(".", name+".jsonl")
		c.PolicyPath = filepath.Join(".", name+"-policy.json")
		configs[i] = c
		a, err := app.NewAgent(c)
		if err != nil {
			return err
		}
		agents[i] = a
		snapshot := a.ExecutionPolicy()
		patch := json.RawMessage(fmt.Sprintf(`{"max_model_requests":%d}`, i+1))
		ack, err := a.UpdatePolicy(snapshot.Revision, patch)
		if err != nil {
			return err
		}
		owned := a.ExecutionPolicy()
		owned.MaxModelRequests = 200
		if a.ExecutionPolicy().MaxModelRequests != i+1 {
			return fmt.Errorf("owned policy snapshot changed authority")
		}
		json.NewEncoder(os.Stdout).Encode(map[string]any{"kind": "policy", "agent": a.ID(), "ack": ack, "policy": a.ExecutionPolicy()})
	}
	handles := make([]ensemble.RequestHandle, 2)
	for i, a := range agents {
		h, err := a.Submit("Use read_file exactly once to read notes.txt. Then report its first-line marker in your final answer. Do not use any other tools or reread the file.")
		if err != nil {
			return err
		}
		handles[i] = h
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	for i, h := range handles {
		result, err := h.Wait(ctx)
		if err != nil {
			return err
		}
		json.NewEncoder(os.Stdout).Encode(map[string]any{"kind": "completion", "agent": agents[i].ID(), "completion": result})
		calls := 0
		for _, event := range agents[i].Events() {
			if event.Type == "request_sent" {
				calls++
			}
		}
		json.NewEncoder(os.Stdout).Encode(map[string]any{"kind": "request_count", "agent": agents[i].ID(), "requests": calls})
	}
	if err := app.Close(); err != nil {
		return err
	}
	restarted := ensemble.New(os.Stderr)
	defer restarted.Close()
	for i, c := range configs {
		c.LogPath = filepath.Join(".", fmt.Sprintf("restart-%d.jsonl", i))
		a, err := restarted.NewAgent(c)
		if err != nil {
			return err
		}
		if a.ExecutionPolicy().MaxModelRequests != i+1 {
			return fmt.Errorf("policy restart mismatch")
		}
		json.NewEncoder(os.Stdout).Encode(map[string]any{"kind": "restart", "agent": a.ID(), "policy": a.ExecutionPolicy()})
	}
	return nil
}
