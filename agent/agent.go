// Package agent provides a reusable framework for building LLM-powered coding
// agents. External programs import this package to create agents, register
// custom tools, and run interactive or scripted sessions.
//
// The shared type vocabulary lives in agent/internal/common and is re-exported
// here so that external callers never import internal/ directly.
package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/waywardgeek/coding-agents-course/agent/internal/common"
	"github.com/waywardgeek/coding-agents-course/agent/internal/jobs"
	"github.com/waywardgeek/coding-agents-course/agent/internal/llm"
	"github.com/waywardgeek/coding-agents-course/agent/internal/tools"
)

// Re-export the types external programs need.
type Config = common.Config
type ToolDecl = common.ToolDecl
type Vendor = common.Vendor
type Surface = common.Surface
type Usage = common.Usage

// Chapter 6: observer and mailbox types.
type Observer = common.Observer
type Observation = common.Observation
type PartDelta = common.PartDelta
type PartFinal = common.PartFinal
type StateChanged = common.StateChanged
type TurnEnded = common.TurnEnded
type AgentID = common.AgentID
type Inbound = common.Inbound
type UserMessage = common.UserMessage
type Hint = common.Hint
type ToolCompleted = common.ToolCompleted
type Interrupt = common.Interrupt
type Mailbox = common.Mailbox
type Media = common.Media
type ModelFeatures = common.ModelFeatures
type TurnState = common.TurnState
type TextPart = common.TextPart

// ToolCallPart and OpaquePart are re-exported because PartFinal now reports
// EVERY part, not just text. A consumer switching on a final needs the types
// to switch on, and before streaming there was nothing but text to see. The
// internal/ wall is what surfaced this: the ch07 exercise would not compile
// without them, which is the whole reason the exercise lives outside the
// module.
type ToolCallPart = common.ToolCallPart
type OpaquePart = common.OpaquePart

// Streaming. DeltaKind says what sort of content a chunk is; Stream is the
// set of kinds a model can actually deliver incrementally. StreamCallbacks is
// the parse-side half of the seam, exported because a caller embedding this
// framework may want to drive an Engine turn directly.
type DeltaKind = common.DeltaKind
type Stream = common.Stream
type StreamCallbacks = common.StreamCallbacks

const (
	VendorAnthropic = common.VendorAnthropic
	VendorOpenAI    = common.VendorOpenAI
	VendorGemini    = common.VendorGemini
)

// TurnState constants.
const (
	Idle             = common.Idle
	InputPending     = common.InputPending
	InFlight         = common.InFlight
	ToolsPending     = common.ToolsPending
	StateInterrupted = common.Interrupted
)

// Media constants.
const (
	MediaImage    = common.MediaImage
	MediaAudio    = common.MediaAudio
	MediaVideo    = common.MediaVideo
	MediaDocument = common.MediaDocument
)

// DeltaKind constants.
const (
	DeltaThinking = common.DeltaThinking
	DeltaText     = common.DeltaText
	DeltaToolCall = common.DeltaToolCall
)

// Stream constants: the delta kinds a model can deliver incrementally.
const (
	StreamText     = common.StreamText
	StreamThinking = common.StreamThinking
	StreamToolArgs = common.StreamToolArgs
	StreamAll      = common.StreamAll
)

// StreamingFor reports which delta kinds may be streamed for a config.
func StreamingFor(cfg Config) Stream { return common.StreamingFor(cfg) }

// DefaultSurface returns the default API surface for a vendor.
func DefaultSurface(v Vendor) Surface { return common.DefaultSurface(v) }

// ToolHandler is the signature for a custom tool: given JSON arguments,
// return the text the model will see.
type ToolHandler func(args json.RawMessage) (string, error)

// Agent is the public handle to a running agent. It implements common.Host,
// providing the logger that all internal code reaches through parent interfaces.
type Agent struct {
	eng    *llm.Engine
	reg    *tools.Reg
	Logger *Logger
}

// NewAgent creates a new agent with the given config and log path. It wires
// up the builtin tools (run_command, read_file, etc.) and any custom tools
// registered before this call. Constructors take interfaces to parents: the
// engine receives the job manager and tool registry as interfaces, not
// concrete types.
func NewAgent(cfg Config, logPath string) *Agent {
	a := &Agent{Logger: DefaultLogger()}
	j := jobs.NewJobs(a)
	a.reg = tools.NewRegistry()
	cfg.Tools = a.reg.Declarations()
	a.eng = llm.NewEngine(cfg, logPath, j, a.reg, a)
	return a
}

// RegisterTool adds a custom tool to this agent's registry. Call this before
// the first Ask. The schema is the JSON Schema for the tool's input parameters.
func (a *Agent) RegisterTool(name, description string, schema json.RawMessage, handler ToolHandler) {
	a.reg.Register(name, description, schema, handler)
	// Update the engine's config so declarations include the new tool.
	a.eng.Cfg.Tools = a.reg.Declarations()
}

// Logf logs a message through the agent's logger. This makes Agent satisfy
// common.Host, so any code that holds a Host can log.
func (a *Agent) Logf(format string, args ...any)    { a.Logger.Logf(format, args...) }
func (a *Agent) APILogf(format string, args ...any) { a.Logger.APILogf(format, args...) }
func (a *Agent) Debugf(format string, args ...any)  { a.Logger.Debugf(format, args...) }

// Ask sends a prompt and runs the full tool loop until the model replies.
func (a *Agent) Ask(prompt string) (string, error) {
	return a.eng.Ask(prompt)
}

// Shutdown ends all running jobs and saves the log.
func (a *Agent) Shutdown() error {
	return a.eng.Shutdown()
}

// Usage returns the token counts accumulated across all requests.
func (a *Agent) Usage() Usage {
	return a.eng.Ctx.Usage
}

// ----------------------------------------------------------------
// Chapter 6: Actor-based API
// ----------------------------------------------------------------

// Actor is the public handle to an actor-based agent.
type Actor = llm.Actor

// Framework manages multiple actors.
type Framework = llm.Framework

// NewActor creates an Actor wrapping the given engine.
func (a *Agent) NewActor() *Actor {
	return llm.NewActor(a.eng, a)
}

// NewFramework creates a multi-agent framework.
func NewFramework(host common.Host) *Framework {
	return llm.NewFramework(host)
}

// NewMailbox creates a new mailbox.
func NewMailbox() *Mailbox {
	return common.NewMailbox()
}

// LookupModel returns model features.
func LookupModel(model string) (ModelFeatures, bool) {
	return common.LookupModel(model)
}

// ConfigFromEnv builds a Config from environment variables (LLM_VENDOR,
// LLM_MODEL, LLM_API_KEY, LLM_BASE_URL). This is the standard way for an
// external program to configure the framework without hardcoding vendor
// details.
func ConfigFromEnv() (Config, error) {
	vendor, err := parseVendor(envOr("LLM_VENDOR", "anthropic"))
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Vendor:    vendor,
		Surface:   common.DefaultSurface(vendor),
		MaxTokens: 1024,
	}
	switch vendor {
	case VendorAnthropic:
		cfg.Model = pick("LLM_MODEL", "ANTHROPIC_MODEL", "claude-sonnet-5")
		cfg.BaseURL = pick("LLM_BASE_URL", "ANTHROPIC_BASE_URL", "https://api.anthropic.com")
		cfg.APIKey = pick("LLM_API_KEY", "ANTHROPIC_API_KEY", "")
	case VendorOpenAI:
		cfg.Model = pick("LLM_MODEL", "OPENAI_MODEL", "gpt-5")
		cfg.BaseURL = pick("LLM_BASE_URL", "OPENAI_BASE_URL", "https://api.openai.com")
		cfg.APIKey = pick("LLM_API_KEY", "OPENAI_API_KEY", "")
	case VendorGemini:
		cfg.Model = pick("LLM_MODEL", "GEMINI_MODEL", "gemini-3.8-flash")
		cfg.BaseURL = pick("LLM_BASE_URL", "GEMINI_BASE_URL", "https://generativelanguage.googleapis.com")
		cfg.APIKey = pick("LLM_API_KEY", "GEMINI_API_KEY", "")
	}
	return cfg, nil
}

func parseVendor(s string) (Vendor, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "anthropic", "claude":
		return VendorAnthropic, nil
	case "openai":
		return VendorOpenAI, nil
	case "gemini":
		return VendorGemini, nil
	}
	return 0, fmt.Errorf("unknown vendor %q (want anthropic, openai or gemini)", s)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func pick(primary, secondary, def string) string {
	if v := os.Getenv(primary); v != "" {
		return v
	}
	return envOr(secondary, def)
}
