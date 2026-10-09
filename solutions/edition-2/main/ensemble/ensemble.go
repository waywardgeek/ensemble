// Package ensemble is the public composition root used by applications.
package ensemble

import (
	"errors"
	"io"
	"log"

	"ensemble/internal/agent"
	"ensemble/internal/common"
	"ensemble/internal/engine"
)

// Public aliases let clients use the core without importing internal packages.
type Config = common.Config
type Agent = common.Agent
type Usage = common.Usage

// The root retains its Agents and one logger. It has no parent or global state.
// A single-Agent executable still constructs this same ownership graph.
type Ensemble struct {
	logger *log.Logger
	agents []Agent
}

// New gives logging one owner; applications choose the diagnostic destination.
func New(diagnostics io.Writer) *Ensemble {
	return &Ensemble{logger: log.New(diagnostics, "", 0)}
}
func (e *Ensemble) Logger() *log.Logger { return e.logger }

func (e *Ensemble) NewAgent(cfg Config) (Agent, error) {
	// Refuse missing model selection rather than silently choosing an old alias.
	// Discovery belongs in setup, not in the one-call-per-turn conversation.
	if cfg.APIKey == "" {
		return nil, errors.New("ANTHROPIC_API_KEY is not set")
	}
	if cfg.Model == "" {
		return nil, errors.New("ANTHROPIC_MODEL is not set; ask GET /v1/models which models exist")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.anthropic.com"
	}
	// Only the root imports implementation spokes and constructs the graph.
	// Constructors store parent pointers before either child becomes usable.
	a := agent.New(e, cfg)
	a.AttachEngine(engine.New(a))
	e.agents = append(e.agents, a)
	return a, nil
}
