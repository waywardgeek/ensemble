// A public consumer: independently configured Agents share one application owner.
package main

import (
	"context"
	"encoding/json"
	"example.com/ensemble"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type observer struct{ events []ensemble.Observation }

func (o *observer) Observe(event ensemble.Observation) { o.events = append(o.events, event) }
func run() error {
	app := ensemble.New(os.Stderr)
	dir := os.Getenv("DEMO_DIR")
	if dir == "" {
		return fmt.Errorf("DEMO_DIR required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	report := map[string]any{}
	answers := []string{}
	for i, marker := range []string{"NORTH-314", "SOUTH-927"} {
		workspace := filepath.Join(dir, fmt.Sprint(i))
		if err := os.Mkdir(workspace, 0700); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(workspace, "same.txt"), []byte(marker+"\n"), 0600); err != nil {
			return err
		}
		selected := []string{"read_file"}
		if i == 1 {
			selected = append(selected, "list_directory")
		}
		config := ensemble.Config{Vendor: os.Getenv("LLM_VENDOR"), Model: os.Getenv("LLM_MODEL"), ResolvedModel: os.Getenv("LLM_RESOLVED_MODEL"), APIKey: os.Getenv("LLM_API_KEY"), Workspace: workspace, LogPath: filepath.Join(workspace, "session.log"), Builtins: selected, MaxTokens: 2048}
		a, err := app.NewAgent(config)
		if err != nil {
			return err
		}
		defer a.Close()
		watch := &observer{}
		subscription, err := app.Subscribe(a.ID(), watch)
		if err != nil {
			return err
		}
		prompt := "Call read_file for same.txt in this workspace, then respond only with the exact marker you read."
		result, err := app.Submit(ctx, ensemble.ClientRequest{AgentID: a.ID(), Prompt: &prompt})
		if err != nil {
			return err
		}
		answers = append(answers, result.Text)
		if result.Text != marker {
			return fmt.Errorf("Agent %d returned wrong workspace marker %q", i, result.Text)
		}
		called, returned := false, false
		var seq uint64
		for _, event := range watch.events {
			if event.AgentID != a.ID() || event.Seq <= seq {
				return fmt.Errorf("observer order or identity incorrect")
			}
			seq = event.Seq
			if event.Kind == "tool_called" {
				called = true
			}
			if event.Kind == "tool_returned" {
				returned = true
			}
		}
		if !called || !returned {
			return fmt.Errorf("missing live observed tool facts")
		}
		count := len(watch.events)
		app.Unsubscribe(subscription)
		if err = a.Ephemeral("after unsubscribe"); err != nil {
			return err
		}
		if len(watch.events) != count {
			return fmt.Errorf("closed subscription received event")
		}
		report[fmt.Sprint(i)] = map[string]any{"agent": a.ID(), "answer": result.Text, "usage": a.Usage(), "observations": watch.events, "visible_tools": a.Config().Tools, "workspace": workspace, "unsubscribed_count_unchanged": true}
	}
	empty, err := app.NewAgent(ensemble.Config{Vendor: os.Getenv("LLM_VENDOR"), Model: os.Getenv("LLM_MODEL"), ResolvedModel: os.Getenv("LLM_RESOLVED_MODEL"), APIKey: os.Getenv("LLM_API_KEY"), LogPath: filepath.Join(dir, "empty.log")})
	if err != nil {
		return err
	}
	defer empty.Close()
	question := "Reply with exactly READY."
	result, err := app.Submit(ctx, ensemble.ClientRequest{AgentID: empty.ID(), Prompt: &question})
	if err != nil {
		return err
	}
	if len(empty.Config().Tools) != 0 {
		return fmt.Errorf("empty Agent advertised tools")
	}
	report["empty"] = map[string]any{"answer": result.Text, "usage": empty.Usage(), "tools": empty.Config().Tools}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(dir, "report.json"), append(data, '\n'), 0600); err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"workspace_answers": answers, "empty_agent_answer": result.Text})
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
