// Package agent owns conversation state and each Agent's configuration.
package agent

import (
	"context"
	"errors"
	"strings"

	"ensemble/internal/common"
)

// History lives as long as its Agent; siblings never share the slice or config.
// Engine is attached by the root before the Agent is returned to a client.
type agent struct {
	parent  common.Ensemble
	config  common.Config
	engine  common.Engine
	history []common.Message
}

// New requires the real owner even for a single, headless Agent.
func New(parent common.Ensemble, cfg common.Config) *agent {
	if parent == nil {
		panic("Agent requires Ensemble")
	}
	return &agent{parent: parent, config: cfg}
}
func (a *agent) Ensemble() common.Ensemble { return a.parent }
func (a *agent) Config() common.Config     { return a.config }
func (a *agent) Engine() common.Engine     { return a.engine }

// AttachEngine completes the relationship built by the composition root.
func (a *agent) AttachEngine(e common.Engine) {
	if e == nil || e.Agent() != a {
		panic("Engine must belong to this Agent")
	}
	a.engine = e
}

// Ask is synchronous: the caller finishes a turn before starting the next.
func (a *agent) Ask(ctx context.Context, question string) (string, error) {
	if strings.TrimSpace(question) == "" {
		return "", errors.New("user message is empty")
	}
	// Stage the question so a failed exchange cannot poison the next role sequence.
	// Appending may reuse spare capacity, but does not extend retained history
	// until the successful assignment below. No concurrent Ask calls are allowed.
	pending := append(a.history, common.Message{Role: "user", Content: question})
	answer, err := a.engine.Send(ctx, pending)
	if err != nil {
		return "", err
	}
	// The provider remembers nothing: retain its exact reply for the next POST.
	a.history = append(pending, common.Message{Role: "assistant", Content: answer})
	return answer, nil
}
