package common

import (
	"context"
	"encoding/json"
)

// RequestConfig is the replayable portion of an immutable request configuration.
// Credentials and endpoint authorization never enter durable history.
type RequestConfig struct {
	System        string           `json:"system"`
	MaxTokens     int              `json:"max_tokens"`
	Tools         []ToolDefinition `json:"tools"`
	ResolvedModel string           `json:"resolved_model,omitempty"`
}
type Completion struct {
	StopReason   string      `json:"stop_reason,omitempty"`
	AgentID      string      `json:"agent_id,omitempty"`
	RequestID    string      `json:"request_id"`
	Outcome      string      `json:"outcome"`
	Text         string      `json:"text"`
	Parts        []Part      `json:"parts,omitempty"`
	PendingHints int         `json:"pending_hints"`
	Error        *EventError `json:"error,omitempty"`
	Usage        Usage       `json:"usage"`
}
type RequestHandle interface {
	ID() string
	Done() <-chan struct{}
	Wait(context.Context) (Completion, error)
	Cancel() error
}
type Collection interface {
	Ensemble() Ensemble
	Wait(context.Context) ([]Completion, bool, error)
}
type ControlAck struct {
	RequestID   string `json:"request_id"`
	Seq         uint64 `json:"seq,omitempty"`
	Sent        bool   `json:"sent"`
	Interrupted bool   `json:"interrupted"`
}
type ActorAgent interface {
	Policy() PolicyService
	TurnAgent
	ID() string
	Engine() ModelEngine
	FinishClose() error
	BeginTurn() error
	EndTurn()
	WatchSource() WatchSnapshot
}
type ModelEngine interface {
	Engine
	ExchangeConfig(context.Context, []byte, Config) (ParsedResponse, error)
	NewOperation(string, string, Config) ModelOperation
	ExchangeOperation(context.Context, ModelOperation, []byte, Config) (ParsedResponse, error)
	Usage() Usage
}
type Actor interface {
	UpdatePolicy(uint64, json.RawMessage) (PolicyAck, error)
	Agent() ActorAgent
	Submit(string) (RequestHandle, error)
	Hint(string) (ControlAck, error)
	Interrupt() (ControlAck, error)
	Cancel(string) error
	Record(Event) error
	Append(Event) error
	Close() error
	Watch() (WatchSnapshot, Watch, error)
	CloseWatch(uint64)
	RegisterPause() (PauseRegistration, error)
	UpdatePause(uint64, bool, bool, bool) (PauseState, error)
}
type ReportTask struct {
	Ready   <-chan struct{}
	Job     Job
	Request JobReport
	Input   *string
}
type PreparedReport struct {
	Generation int
	Event      ToolEvent
	Handle     uint64
	From, To   int64
}

// ActorMessage is the mailbox vocabulary. Private worker paths construct worker
// variants; clients use the Actor operations rather than injecting messages.
type ActorMessage struct {
	SkillOperation   SkillOperation
	Policy           PolicySnapshot
	BaseRevision     uint64
	Patch            json.RawMessage
	Handle           RequestHandle
	Kind             string
	RequestID        string
	Text             string
	Event            Event
	Model            ModelOperation
	Operation        uint64
	Response         ParsedResponse
	Report           PreparedReport
	Error            error
	Registration     uint64
	Typing, Speaking bool
	Reply            chan ActorReply
}
type ActorReply struct {
	SkillResult     SkillResult
	SkillInspection SkillInspection
	PolicyAck       PolicyAck
	Snapshot        WatchSnapshot
	Watch           Watch
	Registration    PauseRegistration
	Pause           PauseState
	Ack             ControlAck
	Handle          RequestHandle
	Error           error
}

// StoppedError lets callers distinguish closed admission without parsing text.
type StoppedError struct{}

func (StoppedError) Error() string { return "Agent stopped" }

// ModelOperation is Engine-owned temporary delivery state, never durable history.
type Fragment struct {
	PartID        int
	Channel, Text string
}
type ModelOperation interface {
	Engine() ModelEngine
	ID() string
	RequestID() string
	Delivery() string
	Exchange(context.Context, []byte) (ParsedResponse, error)
	Emit(context.Context, Fragment) error
	Drain() ([]Fragment, bool)
	Discard()
}
