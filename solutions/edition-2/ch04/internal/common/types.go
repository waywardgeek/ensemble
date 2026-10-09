// Package common holds shared vocabulary; implementation packages own behavior.
package common

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os/exec"
	"time"
)

// Configuration belongs to an Agent. Secrets are never serialized into events.
// Model routes the request; ResolvedModel is an optional exact replay identity.
type Config struct {
	Vendor                                Vendor
	BaseURL, APIKey, Model, ResolvedModel string
	Audio                                 bool
	BuiltinTools                          bool
	// Workspace is the default shell cwd; DataDir stores this Agent's I/O.
	// Keeping them separate lets logs live outside the project directory.
	Workspace, DataDir string
}

// These categories are disjoint after parsing. Cache counts are not extra
// input to add to an already-inclusive vendor total. Keeping token counts,
// rather than prices, allows later clients to apply current pricing.
type Usage struct {
	Input      int `json:"input"`
	CacheWrite int `json:"cache_write"`
	CacheRead  int `json:"cache_read"`
	Output     int `json:"output"`
}

// A routing alias is not proof that two replies came from the same model.
// Capture the provider's returned identity when available; renderers compare
// all three fields before replaying private model material.
type Provenance struct {
	Vendor  Vendor  `json:"vendor"`
	Model   string  `json:"model"`
	Surface Surface `json:"surface"`
}

// Authorship is a fact; roles are wire choices. In particular, a Tool remains
// a Tool even when Anthropic or Gemini must render its output as a user turn.
type Actor string

const (
	Human      Actor = "human"
	AgentActor Actor = "agent"
	System     Actor = "system"
	Tool       Actor = "tool"
)

// Ref names bytes without deciding how a future transport should obtain them.
// Redaction preserves this address even after the visible bytes become a stub.
// There is deliberately no inline/base64 reference kind in the event log.
type RefKind int

const (
	RefPath RefKind = iota + 1
	RefURI
	RefHandle
)

type Ref struct {
	Kind    RefKind `json:"kind"` // Zero is absent only on a text-only redaction stub.
	Locator string  `json:"locator"`
}

// Part is a tagged union, like Event. Keeping its disk shape explicit avoids a
// registry or custom interface decoder. Only fields for Type carry meaning.
// Opaque belongs to a call; Data belongs to a standalone opaque part.
type Part struct {
	Type    string          `json:"type"`
	Text    string          `json:"text,omitempty"`
	MIME    string          `json:"mime,omitempty"`
	Ref     Ref             `json:"ref,omitzero"`
	From    Provenance      `json:"from,omitzero"`
	Data    json.RawMessage `json:"data,omitempty"`
	CallID  string          `json:"call_id,omitempty"`
	Name    string          `json:"name,omitempty"`
	Args    json.RawMessage `json:"args,omitempty"`
	Opaque  json.RawMessage `json:"opaque,omitempty"`
	Parts   []Part          `json:"parts,omitempty"`
	IsError bool            `json:"is_error,omitempty"`
	Stub    string          `json:"stub,omitempty"`
}

// The arrival site reports the speaker and parts. Only the reducer decides
// that a System arrival is pending ephemera instead of ordinary dialogue.
type MessageData struct {
	Actor Actor  `json:"actor"`
	Parts []Part `json:"parts"`
}

// The body is reproducible output and is intentionally absent here. This fact
// marks delivery, including the consumption of pending ephemeral instructions.
type RequestData struct {
	To Provenance `json:"to"`
}

// Parts retain the model's text/call order. ToolCalled records dispatch later;
// moving calls there would lose where the model placed them among its words.
type ResponseData struct {
	Parts []Part     `json:"parts"`
	Usage Usage      `json:"usage"`
	From  Provenance `json:"from"`
}

// These shapes can be replayed before this chapter executes any tools.
// Dispatch records retain the name/arguments; returns carry content and an
// ordinary failure flag, which is distinct from an infrastructure error.
type ToolData struct {
	Job     *JobData        `json:"job,omitempty"`
	Name    string          `json:"name,omitempty"`
	Args    json.RawMessage `json:"args,omitempty"`
	CallID  string          `json:"call_id"`
	Parts   []Part          `json:"parts,omitempty"`
	IsError bool            `json:"is_error,omitempty"`
}

// The inclusive Seq span names original facts without changing those facts.
// Only a summary needs a stored replacement: a deterministic filter or stub
// can be reconstructed from the content it supersedes.
type RedactData struct {
	From        uint64 `json:"from"`
	To          uint64 `json:"to"`
	Level       string `json:"level"`
	Replacement []Part `json:"replacement,omitempty"`
	Reason      string `json:"reason,omitempty"`
}
type ErrorData struct {
	Message string `json:"message"`
}

// Seq orders facts. Time is diagnostic metadata and never influences replay.
// Exactly one payload is present, selected by Type; the loader checks it.
type Event struct {
	Seq      uint64         `json:"seq"`
	Type     string         `json:"type"`
	Time     time.Time      `json:"time"`
	Message  *MessageData   `json:"message,omitempty"`
	Request  *RequestData   `json:"request,omitempty"`
	Response *ResponseData  `json:"response,omitempty"`
	Tool     *ToolData      `json:"tool,omitempty"`
	Redact   *RedactData    `json:"redact,omitempty"`
	Error    *ErrorData     `json:"error,omitempty"`
	Job      *JobKilledData `json:"job,omitempty"`
}

// Seq connects projected content back to its original event. It survives
// category filtering and supplies stable IDs when a provider omitted one.
type Entry struct {
	Seq   uint64 `json:"seq"`
	Actor Actor  `json:"actor"`
	Parts []Part `json:"parts"`
}

// Everything here is derived by replay. Usage is the log's accounting
// projection; the Engine separately measures exchanges made in this process.
// Ephemera is pending delivery, never a second retained conversation slice.
type Context struct {
	Turn     string  `json:"turn"`
	Dialogue []Entry `json:"dialogue"`
	Ephemera []Part  `json:"ephemera"`
	Usage    Usage   `json:"usage"`
}

// These are actual ownership links, not a bag of sibling service callbacks.
type Ensemble interface {
	Logger() *log.Logger
	NextJobHandle() int
}
type Agent interface {
	Ensemble() Ensemble
	Config() Config
	Engine() Engine
	History() History
	Ask(context.Context, string) (string, error)
	Ephemeral(string) error
	Declarations() []ToolDeclaration
	Jobs() Jobs
	Shutdown() error
}

// A declaration is vendor-neutral; only Engine chooses its wire spelling.
// A tool receives the owning Engine through the dispatch context, so later
// diagnostics and owned facts need neither globals nor copied configuration.
type ToolDeclaration struct {
	Name, Description string
	Schema            json.RawMessage
}
type ToolContext interface {
	Agent() Agent
	Engine() Engine
	Job() Job
	Limits() Limits
}
type ToolDefinition struct {
	ToolDeclaration
	Supervision bool
	Run         func(ToolContext, json.RawMessage) (string, error)
}

// Job metadata is a snapshot: events must not change as a process writes more.
// Output is an address, never a blob part to be sent to a provider.
type JobData struct {
	Handle   int    `json:"handle"`
	Status   string `json:"status"`
	Output   Ref    `json:"output"`
	Bytes    int64  `json:"bytes"`
	ExitCode *int   `json:"exit_code,omitempty"`
	Cwd      string `json:"cwd,omitempty"`
}
type JobKilledData struct {
	JobData
	Reason string `json:"reason"`
}

// Limits describe this observation, never the execution lifetime or the
// amount of output that the job is allowed to retain on disk.
type Limits struct {
	Delay    time.Duration
	Pattern  string
	MaxBytes int
}

// Agent owns the collection; each job retains that parent, not copied config
// or logging services. Waiting observes lifetime; it never imposes a deadline.
type Jobs interface {
	Agent() Agent
	Start() (Job, error)
	Find(int) (Job, error)
	Consume(json.RawMessage) (Limits, string, error)
	SetLimits(json.RawMessage) error
	Shutdown() error
}
type Job interface {
	Jobs() Jobs
	Data() JobData
	Finish(string, error)
	StartProcess(*exec.Cmd) error
	Wait(Limits) (string, error)
	Send(string) error
	Kill(string) error
}

// History exposes replay for offline clients, independently of transport.
// Context is a borrowed read-only view: callers must not modify its slices.
// Mutation belongs to Append/Load so the log remains the source of truth.
type History interface {
	Agent() Agent
	Append(Event) error
	Context() *Context
	Load(io.Reader) error
	Dump(io.Writer) error
}

// The seam uses only common values. Vendor message/response structs remain
// private to Engine, so adding a vendor cannot impose its roles on clients.
// Send returns facts; the Agent appends them through the same history door.
type Engine interface {
	Agent() Agent
	Send(context.Context, *Context) ([]Event, error)
	Render(*Context) ([]byte, error)
	Usage() Usage
}
