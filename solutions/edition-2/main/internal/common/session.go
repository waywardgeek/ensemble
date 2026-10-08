package common

import "encoding/json"

type SessionError struct{ Code, Detail string }

func (e *SessionError) Error() string { return e.Code + ": " + e.Detail }

type SessionOptions struct {
	Config Config
	System *string
}
type SessionState struct {
	ID            string  `json:"id"`
	Resumed       bool    `json:"resumed"`
	CheckpointSeq *uint64 `json:"checkpoint_seq"`
}
type CheckpointAck struct {
	AsOf          uint64 `json:"as_of"`
	WatchRevision uint64 `json:"watch_revision"`
}
type CheckpointExport struct {
	AsOf  uint64
	Bytes []byte
}
type SessionBoundary struct {
	LogSeq          uint64 `json:"log_seq"`
	OriginAsOf      uint64 `json:"origin_as_of"`
	CompleteHistory bool   `json:"complete_history"`
	Settled         bool   `json:"settled"`
}
type JobAccess struct {
	Handle uint64 `json:"handle"`
	Live   bool   `json:"live"`
}
type Watermarks struct {
	Event      uint64 `json:"event"`
	Request    uint64 `json:"request"`
	Activation uint64 `json:"activation"`
	Job        uint64 `json:"job"`
}
type HandlerIdentity struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Schema      json.RawMessage `json:"schema"`
}
type SkillIdentity struct {
	Primary        string `json:"primary"`
	CatalogSHA256  string `json:"catalog_sha256"`
	BindingsSHA256 string `json:"bindings_sha256"`
}
type SessionIdentity struct {
	Mode     string            `json:"mode"`
	System   *string           `json:"system"`
	Skills   *SkillIdentity    `json:"skills"`
	Handlers []HandlerIdentity `json:"handlers"`
}
type SessionFact struct {
	SessionID      string           `json:"session_id"`
	Identity       *SessionIdentity `json:"identity,omitempty"`
	OriginAsOf     uint64           `json:"origin_as_of,omitempty"`
	OriginSHA256   string           `json:"origin_sha256,omitempty"`
	HighWatermarks *Watermarks      `json:"high_watermarks,omitempty"`
}
type LimitValues struct {
	Delay    *float64 `json:"ai_callback_delay,omitempty"`
	Pattern  *string  `json:"ai_callback_pattern,omitempty"`
	MaxBytes *int     `json:"max_output_bytes,omitempty"`
}
type LimitsEvent struct {
	CallID    string      `json:"call_id"`
	Name      string      `json:"name,omitempty"`
	Overrides LimitValues `json:"overrides"`
}
type TurnFact struct {
	Index   uint64      `json:"index"`
	Start   uint64      `json:"start"`
	End     uint64      `json:"end"`
	Outcome string      `json:"outcome"`
	Policy  *TurnPolicy `json:"policy"`
}
type ResponseFact struct {
	Seq           uint64          `json:"seq"`
	From          Provenance      `json:"from"`
	Requested     *Provenance     `json:"requested"`
	ModelReported *bool           `json:"model_reported"`
	Usage         Usage           `json:"usage"`
	RawUsage      json.RawMessage `json:"raw_usage"`
	StopReason    string          `json:"stop_reason"`
}
type RedactionFact struct {
	Seq    uint64 `json:"seq"`
	From   uint64 `json:"from"`
	To     uint64 `json:"to"`
	Level  string `json:"level"`
	Reason string `json:"reason"`
}
type GuidanceFact struct {
	Seq        uint64 `json:"seq"`
	Kind       string `json:"kind"`
	ConsumedAt uint64 `json:"consumed_at"`
	Anchor     uint64 `json:"anchor"`
}
type LimitFact struct {
	Seq       uint64      `json:"seq"`
	Kind      string      `json:"kind"`
	CallID    string      `json:"call_id"`
	Name      string      `json:"name"`
	Overrides LimitValues `json:"overrides"`
}
type SkillTransitionRef struct {
	Seq       uint64     `json:"seq"`
	Action    string     `json:"action"`
	Name      string     `json:"name"`
	State     SkillState `json:"state"`
	Activated []uint64   `json:"activated"`
}
type SkillSnapshot struct {
	State        SkillState              `json:"state"`
	Material     []SkillSnapshotMaterial `json:"material"`
	Transitions  []SkillTransitionRef    `json:"transitions"`
	LastID       uint64                  `json:"last_id"`
	Ceiling      []string                `json:"ceiling"`
	Contributors []SkillContributors     `json:"contributors"`
}
type SkillSnapshotMaterial struct {
	Record   SkillActivation `json:"record"`
	EventSeq uint64          `json:"event_seq"`
	Retired  bool            `json:"retired"`
}
type SnapshotSession struct {
	ID             string          `json:"id"`
	Identity       SessionIdentity `json:"identity"`
	AsOf           uint64          `json:"as_of"`
	HighWatermarks Watermarks      `json:"high_watermarks"`
}
type WindowEvent struct {
	Event       Event    `json:"event"`
	Activations []uint64 `json:"activations"`
}
type SnapshotWindow struct {
	Events          []WindowEvent `json:"events"`
	RenderableCount uint64        `json:"renderable_count"`
	EventCount      uint64        `json:"event_count"`
}
type SemanticState struct {
	Session SnapshotSession `json:"session"`
	Context Context         `json:"context"`
	Usage   []UsageAccount  `json:"usage"`
	Skills  *SkillSnapshot  `json:"skills"`
	Limits  *LimitValues    `json:"limits"`
	Window  SnapshotWindow  `json:"window"`
}

// State is encoded by the persistence codec, not by ordinary struct JSON dispatch.
type Checkpoint struct {
	Version        int             `json:"version"`
	SessionID      string          `json:"session_id"`
	Identity       SessionIdentity `json:"identity"`
	AsOf           uint64          `json:"as_of"`
	HighWatermarks Watermarks      `json:"high_watermarks"`
	StateVersion   int             `json:"state_version"`
	State          *SemanticState  `json:"-"`
	StateSHA256    *string         `json:"state_sha256"`
}
type SessionCodec interface {
	Agent() Agent
	Identity(SessionIdentity) error
	Canonical([]byte) ([]byte, error)
	EqualJSON([]byte, []byte) bool
	ValidateLogJSON([]byte, bool) error
	Encode(Checkpoint) ([]byte, error)
	Decode([]byte) (Checkpoint, error)
}
type SessionAgent interface {
	Agent
	Codec() SessionCodec
}

// SessionStore creates the checkpoint worker and owns its disk commit.
type SessionStore interface {
	Agent() SessionAgent
	ReplaceCheckpoint([]byte) error
}

// LogReaderAgent is the creator/owner of a streaming log read. Its acceptance
// operation reduces one validated record without injecting a sibling reducer.
type LogReaderAgent interface {
	Agent
	AcceptReadEvent(Event, int) error
}

type SessionReadState struct {
	Checkpoint  *Checkpoint
	System      *string
	Live, First bool
	OriginHash  string
}
type SessionReadAgent interface {
	Agent
	AcceptSessionRecord(Event, int, *SessionReadState) error
}

type CheckpointResult struct {
	Export CheckpointExport
	Saved  bool
	Error  error
}

// CheckpointIO and CheckpointFile preserve the Store -> I/O -> file creator
// chain while allowing deterministic checked-write failure controls.
type CheckpointIO interface {
	Store() SessionStore
	CreateTemp(string) (CheckpointFile, error)
	Rename(string, string) error
	Remove(string) error
}
type CheckpointFile interface {
	IO() CheckpointIO
	Name() string
	Write([]byte) (int, error)
	Sync() error
	Close() error
}
