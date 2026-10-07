// Package ensemble provides a headless conversation library.
// An Agent is used sequentially; separate Agents have independent state.
package ensemble

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"example.com/ensemble/internal/common"
	"example.com/ensemble/internal/llm"
)

type Config = common.Config
type Usage = common.Usage
type Message = common.Message
type Conversation = common.Conversation

// Ensemble owns the application logger and its Agents. It has no parent.
type Ensemble struct {
	logger *log.Logger
	agents []*Agent
}

func New(diagnostics io.Writer) *Ensemble {
	if diagnostics == nil {
		diagnostics = io.Discard
	}
	return &Ensemble{logger: log.New(diagnostics, "ensemble: ", 0)}
}

func (e *Ensemble) Logf(format string, args ...any) { e.logger.Printf(format, args...) }

// NewAgent establishes the complete owner chain before exposing the Agent.
func (e *Ensemble) NewAgent(config Config) (*Agent, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is required")
	}
	if config.Model == "" {
		return nil, fmt.Errorf("ANTHROPIC_MODEL is required; discover models with GET /v1/models")
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://api.anthropic.com"
	}
	config.BaseURL = strings.TrimRight(config.BaseURL, "/")
	a := &Agent{parent: e, config: config}
	a.engine = llm.New(a)
	e.agents = append(e.agents, a)
	return a, nil
}

type Agent struct {
	parent  common.Ensemble
	config  common.Config
	history common.Conversation
	engine  *llm.Engine
}

func (a *Agent) Ensemble() common.Ensemble      { return a.parent }
func (a *Agent) Config() Config                 { return a.config }
func (a *Agent) History() Conversation          { return append(Conversation(nil), a.history...) }
func (a *Agent) Commit(user, assistant Message) { a.history = append(a.history, user, assistant) }
func (a *Agent) Ask(ctx context.Context, question string) (string, error) {
	return a.engine.Ask(ctx, question)
}
func (a *Agent) Usage() Usage { return a.engine.Usage() }
