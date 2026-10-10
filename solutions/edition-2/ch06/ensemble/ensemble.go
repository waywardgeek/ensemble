// Package ensemble is the public composition root used by applications.
package ensemble

import (
	"context"
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

// Agent owns one conversation, its tools and the actor that coordinates them.
type Agent = common.Agent

// Usage exposes measured, disjoint token counts without fixing prices.
type Usage = common.Usage

// ToolDeclaration is the vendor-neutral schema visible to one Agent.
type ToolDeclaration = common.ToolDeclaration

// ToolContext gives a handler its owning Engine and call-specific job limits.
type ToolContext = common.ToolContext

// The root retains its Agents and one logger. It has no parent or global state.
// A single-Agent executable still constructs this same ownership graph.
type Ensemble struct {
	logger  *log.Logger
	agents  []Agent
	nextJob atomic.Int64
	merged  chan AgentObservation
}

// New gives logging one owner; applications choose the diagnostic destination.
func New(diagnostics io.Writer) *Ensemble {
	return &Ensemble{logger: log.New(diagnostics, "", log.LstdFlags), merged: make(chan AgentObservation, 256)}
}

// NextJobHandle allocates an application-wide identity so Agents cannot collide.
func (e *Ensemble) NextJobHandle() int { return int(e.nextJob.Add(1)) }

// Logf exposes logging without giving children the concrete logger to reconfigure.
// The standard logger serializes writers and stamps each diagnostic.
func (e *Ensemble) Logf(format string, args ...any) { e.logger.Printf(format, args...) }

// Public vendor names allow optional clients to configure the same core.
type Vendor = common.Vendor

const (
	// Anthropic selects the Messages provider translator.
	Anthropic = common.Anthropic
	// OpenAI selects the Chat Completions provider translator.
	OpenAI = common.OpenAI
	// Gemini selects the GenerateContent provider translator.
	Gemini = common.Gemini
)

// NewAgent assembles and retains an Agent with all required parent relationships.
func (e *Ensemble) NewAgent(cfg Config) (Agent, error) {
	if cfg.MaxToolRounds < 0 {
		return nil, errors.New("MaxToolRounds must be nonnegative")
	}
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

// Identity is an envelope owned by the coordinator, not a field on core parts.
type AgentObservation struct {
	// Agent is the actual source child, not a copied or guessed identity.
	Agent Agent
	// Observation carries progress while its enclosing Agent field supplies identity.
	Observation Observation
}

// Observe offers progress without waiting for a consumer to receive it.
func (e *Ensemble) Observe(a Agent, o Observation) {
	select {
	case e.merged <- AgentObservation{a, o}:
	default:
	}
}

// WaitAny consumes merged, identity-tagged progress until the predicate matches.
func (e *Ensemble) WaitAny(ctx context.Context, pred func(AgentObservation) bool) (AgentObservation, error) {
	for {
		select {
		case o := <-e.merged:
			if pred(o) {
				return o, nil
			}
		case <-ctx.Done():
			return AgentObservation{}, ctx.Err()
		}
	}
}

// Observation is sealed progress vocabulary; delivery may be dropped.
type Observation = common.Observation

// Observer receives progress synchronously and must return without blocking.
type Observer = common.Observer

// PartDelta is transient display content; it is never a durable history fact.
type PartDelta = common.PartDelta

// PartFinal publishes a stored part with its history and display identities.
type PartFinal = common.PartFinal

// StateChanged reports a reducer transition after it has taken effect.
type StateChanged = common.StateChanged

// UserMessage starts an idle turn or becomes a hint during an active turn.
type UserMessage = common.UserMessage

// Hint carries live human steering; idle Agents treat it as a new turn.
type Hint = common.Hint

// Interrupt ends the current turn while keeping the Agent and other jobs alive.
type Interrupt = common.Interrupt

// Result acknowledges one submitted request, independently of progress delivery.
type Result = common.Result

// Part exposes the neutral content union for embedding clients and tools.
type Part = common.Part

// Ref names recoverable bytes without embedding them in history.
type Ref = common.Ref

// RefKind distinguishes local paths, remote URIs and managed output addresses.
type RefKind = common.RefKind

const (
	// RefPath locates a local file read by the media renderer.
	RefPath = common.RefPath
	// RefURI preserves a remote address; attachment rendering is deferred.
	RefURI = common.RefURI
	// RefHandle preserves a job output address; attachment rendering is deferred.
	RefHandle = common.RefHandle
)

// ErrStopped distinguishes Agent shutdown from a failed or interrupted turn.
var ErrStopped = common.ErrStopped

// ErrInterrupted ends this request without ending the Agent lifetime.
var ErrInterrupted = common.ErrInterrupted

// ErrRoundLimit reports a refused tool batch after the configured budget.
var ErrRoundLimit = common.ErrRoundLimit

// DefaultMaxToolRounds permits 200 dispatched batches when configuration is zero.
const DefaultMaxToolRounds = common.DefaultMaxToolRounds

// Media is a bit set of input categories a declared model accepts.
type Media = common.Media

// ModelFeatures describes the explicitly known media input capabilities.
type ModelFeatures = common.ModelFeatures

const (
	// MediaImage permits image parts on a supported model.
	MediaImage = common.MediaImage
	// MediaAudio permits audio parts on a supported model.
	MediaAudio = common.MediaAudio
	// MediaVideo permits video parts on a supported model.
	MediaVideo = common.MediaVideo
	// MediaDocument permits PDF document parts on a supported model.
	MediaDocument = common.MediaDocument
)

// LookupModel returns explicit media capabilities; there is no fallback row.
func LookupModel(model string) (ModelFeatures, bool) { return engine.LookupModel(model) }

// Embedders can submit media facts without importing the internal hub.
type Event = common.Event

// MessageData captures authorship and ordered content before reduction.
type MessageData = common.MessageData

// Context is the replayed projection used to build a vendor request.
type Context = common.Context

// Entry associates projected content with the fact that introduced it.
type Entry = common.Entry

// Human marks user authorship independently of vendor role spelling.
const Human = common.Human
