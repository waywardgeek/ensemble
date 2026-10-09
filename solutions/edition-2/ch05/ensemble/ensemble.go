// Package ensemble is the public composition root used by applications.
package ensemble

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"

	"ensemble/internal/agent"
	"ensemble/internal/common"
	"ensemble/internal/engine"
	"ensemble/internal/history"
	"ensemble/internal/jobs"
	"ensemble/internal/tools"
)

// Public aliases let clients use the core without importing internal packages.
type Config = common.Config
type Agent = common.Agent
type Usage = common.Usage
type ToolDeclaration = common.ToolDeclaration
type ToolContext = common.ToolContext

// The root retains its Agents and one logger. It has no parent or global state.
// A single-Agent executable still constructs this same ownership graph.
type Ensemble struct {
	logger  *log.Logger
	agents  []Agent
	nextJob atomic.Int64
}

// New gives logging one owner; applications choose the diagnostic destination.
func New(diagnostics io.Writer) *Ensemble {
	return &Ensemble{logger: log.New(diagnostics, "", log.LstdFlags)}
}
func (e *Ensemble) NextJobHandle() int { return int(e.nextJob.Add(1)) }

// Logf exposes logging without giving children the concrete logger to reconfigure.
// The standard logger serializes writers and stamps each diagnostic.
func (e *Ensemble) Logf(format string, args ...any) { e.logger.Printf(format, args...) }

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
	// Workspace is execution identity, separate from the directory retaining
	// Agent data. Resolve it once; per-call cwd never mutates this default.
	if cfg.Workspace == "" {
		cfg.Workspace, _ = os.Getwd()
	}
	workspace, err := filepath.Abs(cfg.Workspace)
	if err != nil {
		return nil, err
	}
	cfg.Workspace = workspace
	if cfg.DataDir == "" {
		cfg.DataDir = filepath.Join("cr", fmt.Sprintf("agent-%d", len(e.agents)+1))
	}
	a := agent.New(e, cfg)
	a.AttachEngine(engine.New(a))
	a.AttachHistory(history.New(a))
	a.AttachJobs(jobs.New(a))
	if cfg.BuiltinTools {
		a.SetTools(tools.Builtins())
	}
	e.agents = append(e.agents, a)
	return a, nil
}
