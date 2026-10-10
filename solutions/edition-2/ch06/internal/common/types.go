// Package common holds shared vocabulary; implementation packages own behavior.
package common

import (
	"context"
	"encoding/json"
	"io"
	"os/exec"
	"time"
)

// Configuration belongs to an Agent. Secrets are never serialized into events.
// Model routes the request; ResolvedModel is an optional exact replay identity.
type Config struct {
	// Vendor chooses the request and response translator for this Agent.
	Vendor Vendor
	// BaseURL selects the endpoint; APIKey is transport-only. Model routes requests,
	// while ResolvedModel optionally supplies exact replay identity.
	BaseURL, APIKey, Model, ResolvedModel string
	// Audio retains the earlier client input flag; explicit ModelFeatures now govern
	// media rendering.
	Audio bool
	// BuiltinTools installs the standard tool set during Agent construction.
	BuiltinTools bool
	// Workspace is the default shell cwd; DataDir stores this Agent's I/O.
	// Keeping them separate lets logs live outside the project directory.
	Workspace, DataDir string
	// SystemPrompt overrides the default role instruction; LogPath selects saving before
	// completion, or remains empty for caller-owned persistence.
	SystemPrompt, LogPath string
	// MaxToolRounds permits this many dispatched batches per human turn; zero selects
	// 200.
	MaxToolRounds int
}

// These categories are disjoint after parsing. Cache counts are not extra
// input to add to an already-inclusive vendor total. Keeping token counts,
// rather than prices, allows later clients to apply current pricing.
type Usage struct {
	// Input counts ordinary input tokens after removing cache categories.
	Input int `json:"input"`
	// CacheWrite counts input tokens charged to cache creation.
	CacheWrite int `json:"cache_write"`
	// CacheRead counts input tokens served from a provider cache.
	CacheRead int `json:"cache_read"`
	// Output includes all generated tokens, including separately reported thinking.
	Output int `json:"output"`
}

// A routing alias is not proof that two replies came from the same model.
// Capture the provider's returned identity when available; renderers compare
// all three fields before replaying private model material.
type Provenance struct {
	// Vendor identifies the provider that produced or will consume these bytes.
	Vendor Vendor `json:"vendor"`
	// Model is the exact producer identity when the provider supplies it.
	Model string `json:"model"`
	// Surface prevents private material from crossing incompatible API shapes.
	Surface Surface `json:"surface"`
}

// Authorship is a fact; roles are wire choices. In particular, a Tool remains
// a Tool even when Anthropic or Gemini must render its output as a user turn.
type Actor string

const (
	// Human marks user authorship independently of vendor role spelling.
	Human Actor = "human"
	// AgentActor marks model-authored dialogue and tool calls.
	AgentActor Actor = "agent"
	// System marks one-request instructions, consumed only by request_sent.
	System Actor = "system"
	// Tool marks handler testimony even when a vendor renders it as user content.
	Tool Actor = "tool"
)

// Ref names bytes without deciding how a future transport should obtain them.
// Redaction preserves this address even after the visible bytes become a stub.
// There is deliberately no inline/base64 reference kind in the event log.
type RefKind int

const (
	// RefPath locates a local file read by the media renderer.
	RefPath RefKind = iota + 1
	// RefURI preserves a remote address; attachment rendering is deferred.
	RefURI
	// RefHandle preserves a job output address; attachment rendering is deferred.
	RefHandle
)

// Ref names recoverable bytes without embedding them in history.
type Ref struct {
	Kind RefKind `json:"kind"` // Zero is absent only on a text-only redaction stub.
	// Locator is the address interpreted according to Kind, not inline content.
	Locator string `json:"locator"`
}

// Part is a tagged union, like Event. Keeping its disk shape explicit avoids a
// registry or custom interface decoder. Only fields for Type carry meaning.
// Opaque belongs to a call; Data belongs to a standalone opaque part.
type Part struct {
	// Type selects the meaningful fields of this neutral content union.
	Type string `json:"type"`
	// Text contains visible text for a text part, including legitimate empty text.
	Text string `json:"text,omitempty"`
	// MIME identifies a blob's media format before capability checks and encoding.
	MIME string `json:"mime,omitempty"`
	// Ref preserves the address of blob bytes or recoverable redacted output.
	Ref Ref `json:"ref,omitzero"`
	// From binds opaque content and calls to their producing model and surface.
	From Provenance `json:"from,omitzero"`
	// Data retains the exact JSON of a standalone opaque provider block.
	Data json.RawMessage `json:"data,omitempty"`
	// CallID correlates a tool call and its result across vendor role conventions.
	CallID string `json:"call_id,omitempty"`
	// Name selects the declared tool for a tool_call part.
	Name string `json:"name,omitempty"`
	// Args holds canonical JSON arguments without interpreting them as prose.
	Args json.RawMessage `json:"args,omitempty"`
	// Opaque retains private replay material bound to this specific call.
	Opaque json.RawMessage `json:"opaque,omitempty"`
	// Parts holds ordered nested content for a tool_result.
	Parts []Part `json:"parts,omitempty"`
	// IsError distinguishes an ordinary tool failure from successful tool testimony.
	IsError bool `json:"is_error,omitempty"`
	// Stub is the visible replacement for bytes removed from the projection.
	Stub string `json:"stub,omitempty"`
}

// The arrival site reports the speaker and parts. Only the reducer decides
// that a System arrival is pending ephemera instead of ordinary dialogue.
type MessageData struct {
	// Actor records authorship before any vendor role translation.
	Actor Actor `json:"actor"`
	// Parts preserves the captured content order for reduction.
	Parts []Part `json:"parts"`
}

// The body is reproducible output and is intentionally absent here. This fact
// marks delivery, including the consumption of pending ephemeral instructions.
type RequestData struct {
	// To identifies the target whose request consumed pending ephemera.
	To Provenance `json:"to"`
}

// Parts retain the model's text/call order. ToolCalled records dispatch later;
// moving calls there would lose where the model placed them among its words.
type ResponseData struct {
	// Parts preserves the provider's interleaving of text, calls and opaque blocks.
	Parts []Part `json:"parts"`
	// Usage records this response's normalized, disjoint token counts.
	Usage Usage `json:"usage"`
	// From records the response's actual producer, not merely the routing alias.
	From Provenance `json:"from"`
}

// These shapes can be replayed before this chapter executes any tools.
// Dispatch records retain the name/arguments; returns carry content and an
// ordinary failure flag, which is distinct from an infrastructure error.
type ToolData struct {
	// Job is an optional immutable snapshot of supervised execution.
	Job *JobData `json:"job,omitempty"`
	// Name records the tool selected at dispatch.
	Name string `json:"name,omitempty"`
	// Args preserves dispatched arguments independently of later output.
	Args json.RawMessage `json:"args,omitempty"`
	// CallID ties dispatch or returned testimony to the original model call.
	CallID string `json:"call_id"`
	// Parts is the bounded returned content observed at this report.
	Parts []Part `json:"parts,omitempty"`
	// IsError marks a handler failure; infrastructure failures use ErrorData.
	IsError bool `json:"is_error,omitempty"`
}

// The inclusive Seq span names original facts without changing those facts.
// Only a summary needs a stored replacement: a deterministic filter or stub
// can be reconstructed from the content it supersedes.
type RedactData struct {
	// From is the inclusive first original event sequence to rewrite in projection.
	From uint64 `json:"from"`
	// To is the inclusive last original event sequence to rewrite in projection.
	To uint64 `json:"to"`
	// Level selects result stubbing, tool removal, dialogue filtering or summary.
	Level string `json:"level"`
	// Replacement stores a nondeterministic summary so replay needs no model call.
	Replacement []Part `json:"replacement,omitempty"`
	// Reason explains the rewrite without changing the original captured facts.
	Reason string `json:"reason,omitempty"`
}

// ErrorData records a safe diagnostic without retaining provider secrets.
type ErrorData struct {
	// Message is a diagnostic safe to log without request credentials or remote echoes.
	Message string `json:"message"`
}

// Seq orders facts. Time is diagnostic metadata and never influences replay.
// Exactly one payload is present, selected by Type; the loader checks it.
type Event struct {
	// Seq supplies the total fact order; timestamps never determine replay.
	Seq uint64 `json:"seq"`
	// Type selects exactly one of the payload pointers below.
	Type string `json:"type"`
	// Time records capture time for diagnostics while Seq controls ordering.
	Time time.Time `json:"time"`
	// Message captures human, system or other authored content.
	Message *MessageData `json:"message,omitempty"`
	// Request records delivery and consumption of pending ephemera.
	Request *RequestData `json:"request,omitempty"`
	// Response records provider content, provenance and usage.
	Response *ResponseData `json:"response,omitempty"`
	// Tool records dispatch, bounded reports or terminal job snapshots.
	Tool *ToolData `json:"tool,omitempty"`
	// Redact describes a projection rewrite while preserving original history.
	Redact *RedactData `json:"redact,omitempty"`
	// Error records a failed exchange or an explicit turn ending.
	Error *ErrorData `json:"error,omitempty"`
	// Job records an explicit job kill and its reason.
	Job *JobKilledData `json:"job,omitempty"`
}

// Seq connects projected content back to its original event. It survives
// category filtering and supplies stable IDs when a provider omitted one.
type Entry struct {
	// Seq refers back to the original fact even after projection rewrites.
	Seq uint64 `json:"seq"`
	// Actor retains authorship independently of provider role spelling.
	Actor Actor `json:"actor"`
	// Parts is the ordered projected content for this original fact.
	Parts []Part `json:"parts"`
}

// Everything here is derived by replay. Usage is the log's accounting
// projection; the Engine separately measures exchanges made in this process.
// Ephemera is pending delivery, never a second retained conversation slice.
type Context struct {
	// Turn is reducer state, including interrupted as a terminal turn state.
	Turn string `json:"turn"`
	// Dialogue contains retained projected entries in causal order.
	Dialogue []Entry `json:"dialogue"`
	// Ephemera contains pending one-request instructions, consumed at request_sent.
	Ephemera []Part `json:"ephemera"`
	// Usage is historical accounting derived from replayed response facts.
	Usage Usage `json:"usage"`
}

// These are actual ownership links, not a bag of sibling service callbacks.
type Ensemble interface {
	// Logf sends safe diagnostics through the application-owned logger.
	Logf(format string, args ...any)
	// NextJobHandle allocates an application-wide identity so Agents cannot collide.
	NextJobHandle() int
	// Observe offers progress without waiting for a consumer to receive it.
	Observe(Agent, Observation)
}

// Agent owns one conversation, its tools and the actor that coordinates them.
type Agent interface {
	// Ensemble returns the owning application root, including its logging services.
	Ensemble() Ensemble
	// Config returns this Agent's fixed request and workspace configuration.
	Config() Config
	// Engine reaches the transport and measured usage through the owning Agent.
	Engine() Engine
	// History exposes offline replay and snapshots of this Agent's facts.
	History() History
	// Ask blocks on its own request reply or caller cancellation, never a shared idle
	// event.
	Ask(context.Context, string) (string, error)
	// Submit queues a separate turn and returns its buffered, reliable completion reply.
	Submit(context.Context, string) <-chan Result
	// Post accepts live input without waiting for execution or dropping its message.
	Post(Inbound) error
	// Observe registers a progress consumer; its callback must not block the actor.
	Observe(Observer)
	// Wait consumes progress until its predicate matches or its context ends.
	Wait(context.Context, func(Observation) bool) (Observation, error)
	// Record sends a fact through the actor and acknowledges its reduction.
	Record(Event) error
	// Flush waits for submitted turns and saves; background job shutdown is separate.
	Flush(context.Context) error
	// Ephemeral records an instruction for the next request without retaining dialogue.
	Ephemeral(string) error
	// Declarations returns the ordered tools visible to this Agent's model.
	Declarations() []ToolDeclaration
	// RegisterTool installs one Agent's handler before runtime use; duplicate names
	// fail.
	RegisterTool(ToolDeclaration, func(ToolContext, json.RawMessage) (string, error)) error
	// Jobs reaches this Agent's collection of continuing and completed executions.
	Jobs() Jobs
	// Shutdown stops admission, stops managed processes and joins outstanding work.
	Shutdown() error
}

// A declaration is vendor-neutral; only Engine chooses its wire spelling.
// A tool receives the owning Engine through the dispatch context, so later
// diagnostics and owned facts need neither globals nor copied configuration.
type ToolDeclaration struct {
	// Name is the dispatch identifier; Description explains the tool to the model.
	Name, Description string
	// Schema is vendor-neutral JSON describing accepted arguments.
	Schema json.RawMessage
}

// ToolContext gives a handler its owning Engine and call-specific job limits.
type ToolContext interface {
	// Agent returns the required owning Agent rather than a copied service bundle.
	Agent() Agent
	// Engine reaches the transport and measured usage through the owning Agent.
	Engine() Engine
	// Job returns this call's execution handle; supervision calls have no new job.
	Job() Job
	// Limits returns this call's wait/output policy, never an execution deadline.
	Limits() Limits
}

// ToolDefinition binds a declaration to its handler and dispatch policy.
type ToolDefinition struct {
	// ToolDeclaration is the schema published only when this definition is visible.
	ToolDeclaration
	// Supervision skips creation of a new job for an operation on existing jobs.
	Supervision bool
	// Run performs the handler work using its owning call context.
	Run func(ToolContext, json.RawMessage) (string, error)
}

// Job metadata is a snapshot: events must not change as a process writes more.
// Output is an address, never a blob part to be sent to a provider.
type JobData struct {
	// Handle identifies this job across every Agent in its Ensemble.
	Handle int `json:"handle"`
	// Status distinguishes running, done and explicitly killed execution.
	Status string `json:"status"`
	// Output names the retained file, independently of bounded inline reports.
	Output Ref `json:"output"`
	// Bytes counts all retained output, including content omitted from reports.
	Bytes int64 `json:"bytes"`
	// ExitCode is present only when an actual process exit was observed.
	ExitCode *int `json:"exit_code,omitempty"`
	// Cwd records an explicit override; empty means the Agent workspace.
	Cwd string `json:"cwd,omitempty"`
}

// JobKilledData retains the terminal snapshot and the reason for an explicit stop.
type JobKilledData struct {
	// JobData is the immutable terminal snapshot captured at explicit kill.
	JobData
	// Reason distinguishes explicit kill_job from shutdown process termination.
	Reason string `json:"reason"`
}

// Limits describe this observation, never the execution lifetime or the
// amount of output that the job is allowed to retain on disk.
type Limits struct {
	// Delay bounds this wait, never the execution lifetime.
	Delay time.Duration
	// Pattern wakes on a regex match in output not yet reported.
	Pattern string
	// MaxBytes caps the inline report while the full output remains on disk.
	MaxBytes int
}

// Agent owns the collection; each job retains that parent, not copied config
// or logging services. Waiting observes lifetime; it never imposes a deadline.
type Jobs interface {
	// Agent returns the required owning Agent rather than a copied service bundle.
	Agent() Agent
	// Start creates a recoverable output file before publishing a new running job.
	Start() (Job, error)
	// Find resolves a handle even after execution ended, allowing later output recovery.
	Find(int) (Job, error)
	// Consume burns one-shot limits and applies explicit overrides for this call.
	Consume(json.RawMessage) (Limits, string, error)
	// SetLimits validates the next-call override without extending any job lifetime.
	SetLimits(json.RawMessage) error
	// Shutdown closes process admission and stops managed processes; the actor joins Go
	// handlers.
	Shutdown() error
}

// Job owns execution state and recoverable output beyond any individual wait.
type Job interface {
	// Jobs returns the owner of this execution's retained output and lookup lifetime.
	Jobs() Jobs
	// Data returns a synchronized snapshot that cannot mutate the underlying job.
	Data() JobData
	// Finish records Go handler output; a launched process retains completion ownership.
	Finish(string, error)
	// StartProcess transfers completion to the PTY reader and rejects stopped jobs.
	StartProcess(*exec.Cmd) error
	// Wait returns bounded unseen output on completion, pattern or delay without
	// canceling execution.
	Wait(Limits) (string, error)
	// Send writes a line to a running PTY job, appending a newline when absent.
	Send(string) error
	// Kill explicitly stops a process group and suppresses subsequent output for that
	// job.
	Kill(string) error
	// Shutdown prevents later process launch and stops an existing managed process
	// without discarding Go output.
	Shutdown() error
}

// History exposes replay for offline clients, independently of transport.
// Context returns an independent snapshot. Append/Load are offline operations;
// live clients send facts through Agent.Record so one actor owns their ordering.
type History interface {
	// Agent returns the required owning Agent rather than a copied service bundle.
	Agent() Agent
	// Append validates and records one fact, then updates its replay projection.
	Append(Event) error
	// Context returns an independent projection snapshot for readers and renderers.
	Context() *Context
	// Load validates an entire log before replacing history, preserving its recorded
	// order.
	Load(io.Reader) error
	// Dump writes the version header and original facts, not a redacted projection.
	Dump(io.Writer) error
}

// The seam uses only common values. Vendor message/response structs remain
// private to Engine, so adding a vendor cannot impose its roles on clients.
// Send returns facts; the Agent appends them through the same history door.
type Engine interface {
	// Agent returns the required owning Agent rather than a copied service bundle.
	Agent() Agent
	// Send performs one HTTP exchange from rendered bytes and returns normalized facts.
	Send(context.Context, []byte) ([]Event, error)
	// Target identifies the current provider, resolved model and API replay surface.
	Target() Provenance
	// Render builds provider bytes from a projection, reading supported local media
	// Refs.
	Render(*Context) ([]byte, error)
	// Usage returns measured session counts under the Engine's accounting lock.
	Usage() Usage
}
