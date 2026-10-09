// Package agent owns conversation lifetime and each Agent's configuration.
package agent

import (
	"context"
	"errors"
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
	// Send returns ordered facts even on a failed exchange. Append those before
	// returning the error so callers can inspect what happened through replay.
	// No retry or tool dispatch is hidden behind this single-turn operation.
	events, sendErr := a.engine.Send(ctx, a.history.Context())
	var answer strings.Builder
	for _, event := range events {
		if err := a.history.Append(event); err != nil {
			return "", err
		}
		if event.Response != nil {
			for _, p := range event.Response.Parts {
				if p.Type == "text" {
					answer.WriteString(p.Text)
				}
			}
		}
	}
	return answer.String(), sendErr
}
