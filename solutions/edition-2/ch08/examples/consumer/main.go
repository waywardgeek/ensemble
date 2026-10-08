// This independent module demonstrates real providers through the public API.
// Supplied tool results are controlled ingestion, never tool execution.
package main

import (
	"context"
	"encoding/json"
	"example.com/ensemble"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type recorder struct{ events []ensemble.Observation }

func (r *recorder) Observe(e ensemble.Observation) { r.events = append(r.events, e) }
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	app := ensemble.New(os.Stderr)
	folder := os.Getenv("DEMO_DIR")
	config := ensemble.Config{Vendor: os.Getenv("LLM_VENDOR"), APIKey: os.Getenv("LLM_API_KEY"), Model: os.Getenv("LLM_MODEL"), LogPath: filepath.Join(folder, "tools.log"), MaxTokens: 2048, Tools: []ensemble.ToolDefinition{{Name: "inspect", Description: "Read the controlled demonstration record named by path; the example supplies the result and executes nothing.", Schema: json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"],"additionalProperties":false}`)}}}
	a, err := app.NewAgent(config)
	if err != nil {
		return err
	}
	defer a.Close()
	observed := &recorder{}
	subscription, err := app.Subscribe(a.ID(), observed)
	if err != nil {
		return err
	}
	defer app.Unsubscribe(subscription)
	encoder := json.NewEncoder(os.Stdout)
	prompt := `Call inspect exactly once with path "demo-config". Do not invent its result. Please call the tool now.`
	result, err := app.Submit(context.Background(), ensemble.ClientRequest{AgentID: a.ID(), Prompt: &prompt})
	if err != nil {
		return err
	}
	if err = encoder.Encode(map[string]any{"phase": "tool_request", "answer": result.Text, "parts": result.Parts, "usage": result.Usage}); err != nil {
		return err
	}
	calls := []ensemble.Part{}
	for _, part := range result.Parts {
		if part.Type == "tool_call" {
			calls = append(calls, part)
		}
	}
	if len(calls) != 1 {
		return fmt.Errorf("expected one actual model tool call, got %d", len(calls))
	}
	resultEvent := ensemble.Event{Type: "tool_returned", Tool: &ensemble.ToolEvent{CallID: calls[0].CallID, Parts: []ensemble.Part{ensemble.Text("CONTROLLED-RESULT-914: port=8080"), {Type: "blob", MIME: "text/plain", Ref: &ensemble.Ref{Kind: 2, Locator: "https://example.invalid/demo/result"}}}}}
	if err = a.Append(resultEvent); err != nil {
		return err
	}
	events := a.Events()
	resultSeq := events[len(events)-1].Seq
	for _, event := range events {
		if event.Response != nil && event.Response.Usage != nil {
			config.ResolvedModel = event.Response.From.Model
		}
	}
	if err = a.SetConfig(config); err != nil {
		return err
	}
	plain, err := a.Render(config)
	if err != nil {
		return err
	}
	if err = a.Redact(ensemble.Redaction{From: resultSeq, To: resultSeq, Level: "redact_result", Reason: "demonstration"}); err != nil {
		return err
	}
	redacted, err := a.Render(config)
	if err != nil {
		return err
	}
	if !strings.Contains(string(plain), "CONTROLLED-RESULT-914") || strings.Contains(string(redacted), "CONTROLLED-RESULT-914") || !strings.Contains(string(redacted), "https://example.invalid/demo/result") {
		return fmt.Errorf("redaction/reference verification failed")
	}
	if err = os.WriteFile(filepath.Join(folder, "plain-request.json"), plain, 0600); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(folder, "redacted-request.json"), redacted, 0600); err != nil {
		return err
	}
	reply, err := a.Ask(context.Background(), "The tool result was deliberately redacted. Acknowledge the redaction in one brief sentence; do not call any tool.")
	if err != nil {
		return err
	}
	if err = encoder.Encode(map[string]any{"phase": "redacted_followup", "answer": reply, "usage": a.Usage(), "redaction_verified": true}); err != nil {
		return err
	}
	for i, event := range observed.events {
		if event.AgentID != a.ID() || event.Seq != uint64(i+1) {
			return fmt.Errorf("observer attribution/order failed")
		}
	}
	seen := len(observed.events)
	app.Unsubscribe(subscription)
	if err = a.Ephemeral("Local unsubscription check; no model request."); err != nil {
		return err
	}
	// A rejected public event exercises the validator-to-root diagnostic path.
	if err = a.Append(ensemble.Event{Type: "tool_returned", Tool: &ensemble.ToolEvent{CallID: "unknown-local-control", Parts: []ensemble.Part{ensemble.Text("local fault")}}}); err == nil {
		return fmt.Errorf("invalid call control was accepted")
	}
	bConfig := config
	bConfig.Tools = nil
	bConfig.ResolvedModel = ""
	bConfig.LogPath = filepath.Join(folder, "independent.log")
	b, err := app.NewAgent(bConfig)
	if err != nil {
		return err
	}
	defer b.Close()
	if _, err = b.Ask(context.Background(), "My private code for this conversation is ORCHID-572. Reply only OK."); err != nil {
		return err
	}
	oldTotals := b.UsageByModel()
	bConfig.Model = os.Getenv("DEMO_SECOND_MODEL")
	if bConfig.Model == "" {
		return fmt.Errorf("DEMO_SECOND_MODEL must name a discovered second model")
	}
	if err = b.SetConfig(bConfig); err != nil {
		return err
	}
	answer, err := b.Ask(context.Background(), "What private code did I give you? Reply only with the code.")
	if err != nil {
		return err
	}
	if !strings.Contains(answer, "ORCHID-572") {
		return fmt.Errorf("second-model recall failed")
	}
	if len(observed.events) != seen {
		return fmt.Errorf("unsubscribed observer received another event")
	}
	for _, event := range a.Events() {
		encoded, _ := json.Marshal(event)
		if strings.Contains(string(encoded), "ORCHID-572") {
			return fmt.Errorf("Agent histories crossed")
		}
	}
	for from, usage := range oldTotals {
		if b.UsageByModel()[from] != usage {
			return fmt.Errorf("old producing model usage changed after model switch")
		}
	}
	type total struct {
		From  ensemble.Provenance
		Usage ensemble.Usage
	}
	totals := []total{}
	for from, usage := range b.UsageByModel() {
		totals = append(totals, total{from, usage})
	}
	sort.Slice(totals, func(i, j int) bool { return totals[i].From.Model < totals[j].From.Model })
	return encoder.Encode(map[string]any{"phase": "independence_and_switch", "answer": answer, "independent": true, "observer_events": seen, "unsubscribed": true, "usage_by_model": totals})
}
