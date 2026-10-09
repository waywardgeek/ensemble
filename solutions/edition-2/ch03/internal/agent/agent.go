// Package agent owns conversation lifetime and each Agent's configuration.
package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"ensemble/internal/common"
)

// The root retains this Agent; the Agent retains its children. Neither child
// imports the other implementation package. Their shared interfaces let the
// owner coordinate one synchronous turn without introducing a scheduler.
type agent struct {
	parent  common.Ensemble
	config  common.Config
	engine  common.Engine
	history common.History
	tools   []common.ToolDefinition
}

func New(parent common.Ensemble, cfg common.Config) *agent {
	if parent == nil {
		panic("Agent requires Ensemble")
	}
	return &agent{parent: parent, config: cfg}
}
func (a *agent) Ensemble() common.Ensemble { return a.parent }
func (a *agent) Config() common.Config     { return a.config }
func (a *agent) Engine() common.Engine     { return a.engine }
func (a *agent) History() common.History   { return a.history }

// The composition root installs this Agent's visible registry. Keeping one
// ordered slice makes declaration order stable and dispatch agree with visibility.
func (a *agent) SetTools(definitions []common.ToolDefinition) { a.tools = definitions }
func (a *agent) Declarations() []common.ToolDeclaration {
	var declarations []common.ToolDeclaration
	for _, tool := range a.tools {
		declarations = append(declarations, tool.ToolDeclaration)
	}
	return declarations
}
func (a *agent) dispatch(call common.Part) (string, error) {
	for _, tool := range a.tools {
		if tool.Name == call.Name {
			return tool.Run(a, call.Args)
		}
	}
	return "", fmt.Errorf("unknown tool %q", call.Name)
}

// The root constructs children; attachment verifies both sides of ownership
// before exposing the assembled Agent to clients.
func (a *agent) AttachEngine(e common.Engine) {
	if e == nil || e.Agent() != a {
		panic("Engine must belong to this Agent")
	}
	a.engine = e
}
func (a *agent) AttachHistory(h common.History) {
	if h == nil || h.Agent() != a {
		panic("History must belong to this Agent")
	}
	a.history = h
}

// Queueing is not a model request. Record the same arrival fact as any other
// input; the reducer classifies it and the next RequestSent consumes it.
// This keeps a pending instruction recoverable from the log alone.
func (a *agent) Ephemeral(value string) error {
	return a.history.Append(common.Event{Type: "message_received", Message: &common.MessageData{Actor: common.System, Parts: []common.Part{{Type: "text", Text: value}}}})
}

// Ask remains synchronous. Failed attempts now remain honest facts in the log,
// unlike Chapter 1's completed-pair slice; ErrorOccurred ends an in-flight turn.
func (a *agent) Ask(ctx context.Context, question string) (string, error) {
	if strings.TrimSpace(question) == "" {
		return "", errors.New("user message is empty")
	}
	if err := a.history.Append(common.Event{Type: "message_received", Message: &common.MessageData{Actor: common.Human, Parts: []common.Part{{Type: "text", Text: question}}}}); err != nil {
		return "", err
	}
	var answer strings.Builder
	// A tool call is the middle of a human turn. Record the complete reply
	// before running its calls, then return every result before requesting more.
	// Sixteen is a round bound, not a timeout: each tool blocks until it finishes.
	for round := 0; round < 16; round++ {
		// Intermediate narration remains in the log; the caller receives the
		// final reply rather than a concatenation of every tool-planning sentence.
		answer.Reset()
		events, sendErr := a.engine.Send(ctx, a.history.Context())
		var calls []common.Part
		for _, event := range events {
			if err := a.history.Append(event); err != nil {
				return "", err
			}
			if event.Response != nil {
				for _, p := range event.Response.Parts {
					if p.Type == "text" {
						answer.WriteString(p.Text)
					}
					if p.Type == "tool_call" {
						calls = append(calls, p)
					}
				}
			}
		}
		if sendErr != nil || len(calls) == 0 {
			return answer.String(), sendErr
		}
		for _, call := range calls {
			if err := a.history.Append(common.Event{Type: "tool_called", Tool: &common.ToolData{CallID: call.CallID, Name: call.Name, Args: call.Args}}); err != nil {
				return "", err
			}
			// Tool failures are testimony for the model, not failed HTTP exchanges.
			// Sequential dispatch lets each call observe the previous call's writes.
			result, toolErr := a.dispatch(call)
			if toolErr != nil {
				result = toolErr.Error()
			}
			if err := a.history.Append(common.Event{Type: "tool_returned", Tool: &common.ToolData{CallID: call.CallID, Parts: []common.Part{{Type: "text", Text: result}}, IsError: toolErr != nil}}); err != nil {
				return "", err
			}
		}
	}
	return answer.String(), errors.New("tool loop reached 16 rounds")
}
