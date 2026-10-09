// Package ensemble is the public composition root used by applications.
package ensemble

import (
	"errors"
	"io"
	"log"

	"ensemble/internal/agent"
	"ensemble/internal/common"
	"ensemble/internal/engine"
	"ensemble/internal/history"
	"ensemble/internal/tools"
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

// Public vendor names allow optional clients to configure the same core.
type Vendor = common.Vendor

const (
	Anthropic = common.Anthropic
	OpenAI    = common.OpenAI
	Gemini    = common.Gemini
)

func (e *Ensemble) NewAgent(cfg Config) (Agent, error) {
	// A zero vendor retains Chapter 1's public API default. Provenance itself
	// always receives a concrete nonzero vendor before any event is captured.
	if cfg.Vendor == 0 {
		cfg.Vendor = Anthropic
	}
	if cfg.Model == "" {
		return nil, errors.New("provider model is not set")
	}
	if cfg.BaseURL == "" {
		switch cfg.Vendor {
		case Anthropic:
			cfg.BaseURL = "https://api.anthropic.com"
		case OpenAI:
			cfg.BaseURL = "https://api.openai.com"
		case Gemini:
			cfg.BaseURL = "https://generativelanguage.googleapis.com"
		default:
			return nil, errors.New("unknown vendor")
		}
	}
	// Offline render/dump require no key. Live transport rejects missing keys
	// at the application boundary, so offline clients remain network independent.
	a := agent.New(e, cfg)
	a.AttachEngine(engine.New(a))
	a.AttachHistory(history.New(a))
	if cfg.BuiltinTools {
		a.SetTools(tools.Builtins())
	}
	e.agents = append(e.agents, a)
	return a, nil
}
