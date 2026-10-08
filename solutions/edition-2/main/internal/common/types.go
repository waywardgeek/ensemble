// Package common declares shared vocabulary; implementations live in their spokes.
package common

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type Conversation []Message

type Config struct {
	Skills           *SkillConfig
	PolicyPath       string
	DisableStreaming bool
	APIKey           string
	Model            string
	BaseURL          string
	Vendor           string
	ResolvedModel    string
	System           string
	MaxTokens        int
	Tools            []ToolDefinition
	Workspace        string
	Builtins         []string
	LogPath          string
}
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Schema      json.RawMessage `json:"input_schema"`
}
type Usage struct {
	Input      int64 `json:"input"`
	CacheWrite int64 `json:"cache_write"`
	CacheRead  int64 `json:"cache_read"`
	Output     int64 `json:"output"`
}

// JSON decoding must distinguish a missing count from an observed zero.
func (u *Usage) UnmarshalJSON(data []byte) error {
	var wire struct {
		Input      *int64 `json:"input"`
		CacheWrite *int64 `json:"cache_write"`
		CacheRead  *int64 `json:"cache_read"`
		Output     *int64 `json:"output"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return fmt.Errorf("invalid normalized usage")
	}
	if wire.Input == nil || wire.CacheWrite == nil || wire.CacheRead == nil || wire.Output == nil || *wire.Input < 0 || *wire.CacheWrite < 0 || *wire.CacheRead < 0 || *wire.Output < 0 {
		return fmt.Errorf("missing or invalid normalized usage")
	}
	*u = Usage{*wire.Input, *wire.CacheWrite, *wire.CacheRead, *wire.Output}
	return nil
}

type Provenance struct {
	Vendor  string `json:"vendor"`
	Model   string `json:"model"`
	Surface string `json:"surface"`
}
type Ref struct {
	Kind    int    `json:"kind"`
	Locator string `json:"locator"`
}
type Part struct {
	Type    string          `json:"type"`
	Text    *string         `json:"text,omitempty"`
	From    *Provenance     `json:"from,omitempty"`
	Opaque  json.RawMessage `json:"opaque,omitempty"`
	CallID  string          `json:"call_id,omitempty"`
	Name    string          `json:"name,omitempty"`
	Args    json.RawMessage `json:"args,omitempty"`
	Parts   []Part          `json:"parts,omitempty"`
	IsError bool            `json:"is_error,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
	MIME    string          `json:"mime,omitempty"`
	Ref     *Ref            `json:"ref,omitempty"`
	Stub    string          `json:"stub,omitempty"`
}
type Entry struct {
	SkillName  string `json:"skill_name,omitempty"`
	Activation uint64 `json:"activation,omitempty"`
	Anchor     uint64 `json:"anchor,omitempty"`
	Seq        uint64 `json:"seq,omitempty"`
	Actor      string `json:"actor"`
	Purpose    string `json:"purpose"`
	Parts      []Part `json:"parts"`
}
type TurnEvent struct {
	Policy    *TurnPolicy `json:"policy,omitempty"`
	RequestID string      `json:"request_id"`
	Outcome   string      `json:"outcome,omitempty"`
}
type HintEvent struct {
	RequestID string `json:"request_id"`
	Text      string `json:"text"`
}
type RequestEvent struct {
	Delivery      string         `json:"delivery,omitempty"`
	Hints         []uint64       `json:"hints"`
	Configuration *RequestConfig `json:"configuration,omitempty"`
	To            Provenance     `json:"to"`
	Ephemera      []uint64       `json:"ephemera"`
}
type Response struct {
	StopReason    string          `json:"stop_reason,omitempty"`
	From          Provenance      `json:"from"`
	Requested     *Provenance     `json:"requested,omitempty"`
	ModelReported *bool           `json:"model_reported,omitempty"`
	Parts         []Part          `json:"parts,omitempty"`
	Usage         *Usage          `json:"usage,omitempty"`
	RawUsage      json.RawMessage `json:"raw_usage,omitempty"`
}

// ParsedResponse carries temporary parser facts, never public event metadata.
type ParsedResponse struct {
	PartIDs        []int
	Response       Response
	MissingCallIDs []int
}

type ToolEvent struct {
	Job     *JobSnapshot    `json:"job,omitempty"`
	CallID  string          `json:"call_id"`
	Name    string          `json:"name,omitempty"`
	Args    json.RawMessage `json:"args,omitempty"`
	Parts   []Part          `json:"parts,omitempty"`
	IsError bool            `json:"is_error,omitempty"`
}
type Redaction struct {
	From   uint64 `json:"from"`
	To     uint64 `json:"to"`
	Level  string `json:"level"`
	Reason string `json:"reason"`
}
type EventError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type Event struct {
	Skills   *SkillTransition `json:"skills,omitempty"`
	Turn     *TurnEvent       `json:"turn,omitempty"`
	Hint     *HintEvent       `json:"hint,omitempty"`
	Job      *JobSnapshot     `json:"job,omitempty"`
	Seq      uint64           `json:"seq"`
	Type     string           `json:"type"`
	Time     string           `json:"time"`
	Message  *Entry           `json:"message,omitempty"`
	Request  *RequestEvent    `json:"request,omitempty"`
	Response *Response        `json:"response,omitempty"`
	Tool     *ToolEvent       `json:"tool,omitempty"`
	Redact   *Redaction       `json:"redact,omitempty"`
	Error    *EventError      `json:"error,omitempty"`
}
type CallState struct {
	JobHandle  uint64
	Part       Part
	Dispatched bool
	Returned   bool
}
type Context struct {
	// Dialogue projection only. Runtime grants and material authority belong to Skills.
	SkillMode      bool
	SkillPrimary   string
	SkillBatch     uint64
	DeferredSkills []Entry
	PendingSkills  []Entry
	Hints          []Entry
	TurnID         string
	TurnIDs        map[string]bool
	ExplicitTurns  bool
	FinalResponse  bool
	Jobs           map[uint64]JobSnapshot
	Entries        []Entry
	Instructions   []Entry
	Ephemera       []Entry
	Pending        *Entry
	Active         bool
	Continuation   bool
	Calls          map[string]CallState
	LastSeq        uint64
}
type Observation struct {
	Skills          *SkillState     `json:"skills,omitempty"`
	ExecutionPolicy *PolicySnapshot `json:"execution_policy,omitempty"`
	Paused          bool            `json:"paused"`
	TypingClients   int             `json:"typing_clients"`
	SpeakingClients int             `json:"speaking_clients"`
	OperationID     string          `json:"operation_id,omitempty"`
	Delivery        string          `json:"delivery,omitempty"`
	PartID          int             `json:"part_id,omitempty"`
	Channel         string          `json:"channel,omitempty"`
	Text            string          `json:"text,omitempty"`
	ResponseSeq     uint64          `json:"response_seq,omitempty"`
	PartIndex       int             `json:"part_index"`
	Accepted        bool            `json:"accepted"`
	Code            string          `json:"code,omitempty"`
	Message         string          `json:"message,omitempty"`
	RequestID       string          `json:"request_id,omitempty"`
	OldState        string          `json:"old_state,omitempty"`
	State           string          `json:"state,omitempty"`
	Position        int             `json:"position,omitempty"`
	Part            *Part           `json:"part,omitempty"`
	AgentID         string          `json:"agent_id"`
	Seq             uint64          `json:"seq"`
	Kind            string          `json:"kind"`
	Event           Event           `json:"event"`
}
type Observer interface{ Observe(Observation) }
type ClientRequest struct {
	AgentID   string
	Prompt    *string
	Ephemeral *string
	Redact    *Redaction
}
type ClientResult struct {
	Text  string
	Parts []Part
	Usage Usage
}

type Ensemble interface {
	ClaimSettingsPath(string) (string, error)
	ReleaseSettingsPath(string)
	Logf(string, ...any)
	Publish(string, Event)
	AllocateHandle() uint64
	Observe(Observation)
	Collect([]RequestHandle) Collection
}
type Agent interface {
	Ensemble() Ensemble
	Config() Config
	Workspace() string
	ModelReady(ModelOperation)
}
type Engine interface{ Agent() Agent }
type EventLog interface {
	Agent() Agent
	Append(Event) error
	Close() error
}

// ClientOwner is public through an alias; optional clients never import internal packages.
type ClientOwner interface {
	SkillState(string) (*SkillState, error)
	InspectSkills(string) (SkillInspection, error)
	LoadSkill(string, string) (SkillResult, error)
	UnloadSkill(string, string) (SkillResult, error)
	ClaimSettingsPath(string) (string, error)
	ReleaseSettingsPath(string)
	ExecutionPolicy(string) (PolicySnapshot, error)
	UpdatePolicy(string, uint64, json.RawMessage) (PolicyAck, error)
	Watch(string) (WatchSnapshot, Watch, error)
	RegisterPause(string) (PauseRegistration, error)
	Logf(string, ...any)
	Submit(context.Context, ClientRequest) (ClientResult, error)
	SubmitPrompt(string, string) (RequestHandle, error)
	Hint(string, string) (ControlAck, error)
	Interrupt(string) (ControlAck, error)
	Collect([]RequestHandle) Collection
	SubscriptionStatus(uint64) string
	Subscribe(string, Observer) (uint64, error)
	Unsubscribe(uint64)
}

// TurnAgent exposes the owning Agent's serialized event path to its Engine.
// The composition root's private adapter prevents clients bypassing admission.
type TurnAgent interface {
	SkillView() SkillInspection
	ChangeSkill(SkillOperation) (SkillResult, error)
	Agent
	TurnSnapshot() Context
	RecordTurn(Event) error
	RecordResponse(ParsedResponse) error
	NextSequence() uint64
	Registry() Registry
	Jobs() Jobs
}
type Registry interface {
	Management(string) bool
	SkillOperation(Part, uint64) (SkillOperation, error)
	SkillAcknowledgement(Part, SkillResult, error, string) ToolEvent
	Agent() Agent
	Declarations() []ToolDefinition
	Execute(Part) ToolEvent
	Kind(string) (available, supervision bool)
	ExecuteJob(Part, Job) *ExecutionResult
	ResolveLimits(Part) (Limits, string, error)
	Supervise(Part, Limits, string) error
	BeginSupervision(Part, Limits, string) (*ToolEvent, *ReportTask)
}

// Job snapshots are durable facts; live handles never reattach during replay.
type JobSnapshot struct {
	Handle   uint64 `json:"handle"`
	Status   string `json:"status"`
	Output   Ref    `json:"output"`
	Bytes    int64  `json:"bytes"`
	ExitCode *int   `json:"exit_code,omitempty"`
	IsError  bool   `json:"is_error,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Cwd      string `json:"cwd,omitempty"`
}
type Limits struct {
	Delay    time.Duration
	Pattern  *regexp.Regexp
	MaxBytes int
}
type LimitOverrides struct {
	Delay      *time.Duration
	Pattern    *regexp.Regexp
	PatternSet bool
	MaxBytes   *int
}
type JobAgent interface {
	Agent
	RecordJob(Event) error
	Registry() Registry
	Fault(error)
}
type ToolAgent interface {
	GrantedTools() []string
	Agent
	Jobs() Jobs
	RecordTool(Event) error
}
type ExecutionResult struct {
	Text    string
	Note    string // Presentation facts are never bytes in the retained artifact.
	IsError bool
}
type JobReport struct {
	CallID     string
	Limits     Limits
	Note       string
	Original   bool
	MatchStart int64 // -1 starts at the unconsumed report cursor.
}
type Jobs interface {
	Agent() JobAgent
	Resolve(LimitOverrides) (Limits, bool)
	SetLimits(LimitOverrides)
	Create() (Job, error)
	Lookup(uint64) (Job, error)
	Send(Job, string) (int64, error)
	SendContext(context.Context, Job, string) (int64, int, error)
	Kill(Job, string) error
	RequestKill(Job, string) (<-chan struct{}, error)
	Start(Job, Part)
	Abort(Job)
	Report(Job, JobReport) error
	PrepareReport(context.Context, Job, JobReport) (PreparedReport, error)
	CommitReport(PreparedReport) error
	Close() error
}
type Job interface {
	Jobs() Jobs
	Snapshot() JobSnapshot
	StartProcess(command, cwd string) error
}
