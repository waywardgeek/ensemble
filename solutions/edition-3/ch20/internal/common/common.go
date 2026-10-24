// Package common holds the vocabulary shared by the owners, not runtime state.
package common

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

// Config belongs to one Agent. Model names must come from provider discovery.
type Config struct {
	// ID and ParentID identify fleet ownership. Prompt is task dialogue, never SYSTEM.
	ID, ParentID, Prompt string
	SubmitSchema         json.RawMessage

	// BaseDir resolves relative tool paths; DataDir retains job output.
	// Empty BaseDir uses the launch directory; empty DataDir creates an Agent-local directory.
	BaseDir         string
	DataDir         string
	SandboxRoot     string
	SafeMode        bool
	EnableWebSearch bool
	// BaseURL selects the provider or a local HTTP peer; empty means Anthropic.
	BaseURL       string
	Vendor        Vendor
	Surface       Surface
	ResolvedModel string
	// APIKey authenticates requests and must never appear in diagnostics.
	APIKey           string
	AuthMode         AuthMode
	Account          string // optional stable account selector, never a token
	ReasoningEffort  string
	ReasoningSummary string
	// Model is required; there is deliberately no remembered model default.
	Model string
	// Tools selects startup capabilities: nil enables the default catalogue;
	// a non-nil slice selects names. Empty Tools without any skill selection
	// is application-only: it disables all tools and omits the optional fleet
	// catalogue, so applications may register their own submit operation.
	// Installed core, compiled and MCP names still cannot be shadowed.
	Tools      []string
	MCPServers []MCPServer
	Skills     *SkillSpec // nil preserves inherited non-skill behavior
	// StartupSkills are admitted after the primary, before the initial SYSTEM.
	StartupSkills []string
	// RequireToolExplanation selects this model's before-effects explanation policy.
	RequireToolExplanation bool
	LogPath                string
	// SavePath selects the archive; empty uses an isolated Agent path under DataDir.
	SavePath string
	// LogMode selects full (default) or dialogue-only archives.
	LogMode LogMode
	// Nil selects 200; an explicit zero retains unlimited diagnostics.
	DiagnosticMaxRecords *int
	ContextTarget        int
	MaxToolRounds        int
	SystemPrompt         string
	DisableStreaming     bool
	PauseOnStart         bool
	Speaking             bool
	StreamCapability     StreamCapability
}

// Message is a retained user or assistant turn, sent verbatim on later requests.
type Message struct {
	// Role alternates between user and assistant, beginning with user.
	Role string `json:"role"`
	// Content is nonempty text; request content is not a response block array.
	Content string `json:"content"`
}

// Usage is measured token consumption, including unusable answers.
// -1 means unreported/unknown; zero is a reported zero. Input is uncached input.
type Usage struct {
	// Input counts input tokens reported by the provider.
	Input int64 `json:"input"`
	// Output counts output tokens reported by the provider.
	Output     int64 `json:"output"`
	CacheWrite int64 `json:"cache_write"`
	CacheRead  int64 `json:"cache_read"`
}

// Ensemble is the Agent's actual owner and the source of application services.
type Ensemble interface {
	// Logger returns this application's diagnostic destination.
	Logger() *log.Logger
	Auth() Auth
	Publish(Agent, Observation)
	Trace(TraceRecord)
	RegisterJob(Job) (int, error)
	LookupJob(int) (Job, bool)
	JobTable() []Job
	Shutdown() error
}

// Agent owns one mailbox and one conversation actor.
type Agent interface {
	ID() string
	View() ViewSnapshot
	Pause() error
	Unpause() error
	// Parent returns the Ensemble which owns this conversation.
	Parent() Ensemble
	// Config returns a value copy, leaving configuration with its owner.
	Config() Config
	// SetMaxToolRounds changes next-turn policy, not an active turn.
	// Zero selects DefaultMaxToolRounds; negative values are rejected.
	SetMaxToolRounds(int) error
	RecordFact(Event) error
	SetContextTarget(int) error
	// SetModel selects a same-provider model only while the actor is idle.
	SetModel(string) error
	CompactContext(context.Context, string, string) (string, error)
	// SetPersistencePolicy changes archive mode and diagnostic retention for future writes.
	SetSecurityPolicy(bool, bool)
	SetPersistencePolicy(LogMode, int) error
	// Engine returns this Agent's transport and accounting owner.
	Engine() Engine
	// Tools is this Agent's independently owned synchronous registry.
	Tools() ToolRegistry
	Skills() SkillRegistry
	MCP() MCPManager
	RegisterExternalTool(ToolDefinition, ToolHandler) error
	AttachMCPConnection(string, MCPTransport) error
	ServeMCPConnection(context.Context, MCPTransport) error
	RegisterVar(string, func() string)
	Sandbox() Sandbox
	Jobs() Jobs
	Shutdown() error
	// Save writes a coherent archive without stopping this Agent.
	Save() error
	// Recovery describes the file and boundary installed at startup.
	Recovery() Recovery
	Post(Inbound) error
	Submit(context.Context, string) <-chan Result
	// Complete validates a terminal string result; it does not submit a chat prompt.
	Complete(context.Context, string) <-chan Result
	// SubmitLive starts a turn if idle, or adds a hint to the current turn.
	// Joined callers receive the same reliable result and TurnID at completion.
	SubmitLive(context.Context, string) <-chan Result
	Subscribe(Observer)
	Observe() ObservationQueue
	Wait(context.Context, func(Observation) bool) (Observation, error)
	ReportJob(JobReport)
	InvokeTool(context.Context, Part, string) error
	// Ask records every attempt; the compatibility History view contains completed text pairs.
	Ask(context.Context, string) (string, error)
	// History returns a copy so callers cannot corrupt retained turns.
	History() []Message
	Log() History
	Render() ([]byte, error)
	ParseResponse(int, []byte) ([]Event, error)
	// Usage reads measured totals from Engine rather than keeping a second copy.
	Usage() Usage
}

// Engine exchanges messages; it never retains the conversation itself.
type Engine interface {
	Cache() Cache
	Status() EngineStatus
	// Parent returns the Agent which owns transport configuration.
	Parent() Agent
	// Send makes one HTTP request, with cancellation controlled by the caller.
	Send(context.Context, []Message) (string, error)
	Exchange(context.Context, Context) ([]Event, error)
	ExchangeRequest(context.Context, []byte, ...ParseCallbacks) ([]Event, error)
	Render(Context) ([]byte, error)
	// EffectiveSystemPrompt describes the same system instructions used by Render.
	EffectiveSystemPrompt(Context) (string, error)
	Parse(*http.Response, ParseCallbacks) ([]Event, error)
	// Usage returns cumulative measured totals, not estimates.
	Usage() Usage
}

// History owns causal facts and their projection; its parent is the real Agent.
type History interface {
	Parent() Agent
	Append(Event) error
	// AppendSaved commits one fact only after the supplied atomic-file save succeeds.
	AppendSaved(Event, func(func(io.Writer) error) error) error
	Events() []Event
	EventsSince(Seq) []Event
	EventsRange(Seq, Seq) []Event
	RecentThrough(Seq, time.Time, int, ...func(Event) bool) []Event
	LastSeq() Seq
	Recent(time.Time, int) []Event
	Context() Context
	Load(io.Reader) error
	Dump(io.Writer) error
	// SaveFile captures and atomically replaces a validated recovery archive.
	SaveFile(string, SavedConfig, LogMode) error
	// LoadFile stages primary/backup recovery before installing any state.
	LoadFile(string) (Recovery, error)
}

// Unknown capabilities intentionally use plain responses. Explicit declarations
// support compatible endpoints without guessing from a routing alias.
type StreamCapability string

const (
	StreamUnknown StreamCapability = ""
	StreamPlain   StreamCapability = "plain"
	StreamSSE     StreamCapability = "sse"
)

func (c Config) Streaming() bool { return !c.DisableStreaming && c.StreamCapability == StreamSSE }

// Parse callbacks are synchronous transport notifications, not History writes.
// A PartID of zero is valid. IDs are local to one response.
type ParseCallbacks struct {
	OnDelta func(PartDelta)
	OnFrame func(Frame)
	OnEvent func(Event)
}

// Frame carries a provider body or SSE payload, never request headers or URLs.
type Frame struct {
	Event string
	Data  string
}

// TraceRecord is process-local diagnostics, not durable conversation. Only root
// encodes it and redacts configured credentials; leaves never open trace files.
type TraceRecord struct {
	Stage      string `json:"stage"`
	AgentID    string `json:"agent_id,omitempty"`
	Detail     string `json:"detail,omitempty"`
	TurnID     uint64 `json:"turn_id,omitempty"`
	ResponseID uint64 `json:"response_id,omitempty"`
	PartID     uint64 `json:"part_id"`
	Kind       string `json:"kind,omitempty"`
	Bytes      int    `json:"bytes,omitempty"`
	Payload    string `json:"payload,omitempty"`
	Event      string `json:"event,omitempty"`
}
