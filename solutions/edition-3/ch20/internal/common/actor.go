package common

import (
	"context"
	"encoding/json"
	"errors"
)

const DefaultMaxToolRounds = 200

var ErrInterrupted = errors.New("turn interrupted")
var ErrStopped = errors.New("Agent stopped")
var ErrToolRoundLimit = errors.New("tool round limit reached")

type Inbound interface{ inbound() }
type UserMessage struct{ Text string }
type Hint struct{ Text string }
type Ephemeral struct{ Text string }

// Question preserves the application's addressing and correlation metadata.
type Question struct{ Text, Sender, Recipient, RequestID, TurnID string }
type Interrupt struct{}
type Pause struct{}
type Unpause struct{ TurnID uint64 }
type SetSpeaking struct{ Enabled bool }
type SpeakParts struct {
	Parts  PartList
	Reason string
}

func (Pause) inbound()       {}
func (Unpause) inbound()     {}
func (SetSpeaking) inbound() {}
func (SpeakParts) inbound()  {}

type Paused struct {
	Paused bool   `json:"paused"`
	Reason string `json:"reason,omitempty"`
}
type SpeakingChanged struct {
	Speaking bool `json:"speaking"`
}
type SpeechRequested struct {
	TurnID uint64
	Parts  PartList
	Reason string
}

// TurnEnded is already the durable event type name.
type TurnEndedObservation struct {
	TurnID uint64
	Text   string
}

func (Paused) observation()               {}
func (SpeakingChanged) observation()      {}
func (SpeechRequested) observation()      {}
func (TurnEndedObservation) observation() {}

func (UserMessage) inbound() {}
func (Hint) inbound()        {}
func (Ephemeral) inbound()   {}
func (Question) inbound()    {}
func (Interrupt) inbound()   {}

type Result struct {
	TurnID     uint64
	Text       string
	Submission json.RawMessage
	Err        error
}
type Observation interface{ observation() }

// ViewSnapshot is an immutable process-local publication checkpoint. Through
// never includes a pending/failed Save; Partials contain complete active text.
// Consumers may skip wakeups and recover from the latest checkpoint.
type ViewSnapshot struct {
	ContinueEpoch    uint64
	TurnID           uint64
	SpeechEpoch      uint64
	Revision         uint64
	Through          Seq
	State            TurnState
	Paused, Speaking bool
	Reason           string
	ControlEpoch     uint64
	Partials         []PartDelta
	SpeechID         uint64
	Speech           SpeechRequested
}
type ModelChanged struct {
	Model string `json:"model"`
}

func (ModelChanged) observation() {}

type BillingSelected struct {
	Mode AuthMode `json:"mode"`
}

func (BillingSelected) observation() {}

type ViewChanged struct{}

func (ViewChanged) observation() {}

type HistoryCommitted struct{ Event Event }

func (HistoryCommitted) observation() {}

type DeltaKind string

const (
	DeltaText             DeltaKind = "text"
	DeltaThinking         DeltaKind = "thinking"
	DeltaReasoningSummary DeltaKind = "reasoning_summary"
	DeltaToolCall         DeltaKind = "tool_call"
)

type PartDelta struct {
	TurnID     uint64    `json:"turn_id"`
	ResponseID uint64    `json:"response_id"`
	Kind       DeltaKind `json:"kind"`
	// Name is supplemental metadata; tool-call Chunk content carries the
	// arriving name fragments followed by argument fragments for appenders.
	Name   string `json:"name,omitempty"`
	PartID uint64 `json:"part_id"`
	Chunk  string `json:"chunk"`
}
type ToolDispatched struct {
	TurnID       uint64
	PartID       uint64
	CallID, Name string
	Input        json.RawMessage
}

func (ToolDispatched) observation() {}

type ToolFinished struct {
	TurnID       uint64
	PartID       uint64
	CallID, Name string
	Parts        PartList
	IsError      bool
}

func (ToolFinished) observation() {}

type PartFinal struct {
	TurnID     uint64 `json:"turn_id"`
	ResponseID uint64 `json:"response_id"`
	Seq        uint64 `json:"seq"`
	PartID     uint64 `json:"part_id"`
	Part       Part   `json:"part"`
}
type StateChanged struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func (PartDelta) observation()    {}
func (PartFinal) observation()    {}
func (StateChanged) observation() {}

type Observer interface{ Observe(Observation) }

type Media uint8

const (
	MediaImage Media = 1 << iota
	MediaAudio
	MediaVideo
	MediaDocument
)

// ResponsesFeatures describes only the Responses contract for an exact model.
// It does not grant the same capabilities to that model on another API.
type ResponsesFeatures struct {
	DefaultEffort          string
	Summary                bool
	RequireToolExplanation bool
}

type ModelFeatures struct {
	InlineTools      bool
	StubsToolResults bool
	Media            Media
	Streaming        StreamCapability
	Responses        *ResponsesFeatures
}

// Exact rows only. An absent row permits text, never guessed attachments.
func LookupModel(model string) (ModelFeatures, bool) {
	switch model {
	case "gpt-ch19-course", "gpt-6.1-sol", "gpt-5.6-sol":
		// These rows do not require an ordinary-text tool prelude. Public
		// reasoning summaries remain separate facts, not evidence for that gate.
		return ModelFeatures{Streaming: StreamSSE, Responses: &ResponsesFeatures{DefaultEffort: "medium", Summary: true}}, true
	case "claude-opus-5-course":
		return ModelFeatures{InlineTools: true, StubsToolResults: true}, true
	case "claude-sonnet-5-course":
		return ModelFeatures{}, true
	case "claude-3-haiku-20240307", "claude-haiku-5-5", "claude-sonnet-4-6", "claude-opus-4-6", "gpt-5.4-mini", "gpt-5.4", "gpt-5.6":
		return ModelFeatures{Media: MediaImage, Streaming: StreamSSE}, true
	case "gemini-3.8-flash":
		return ModelFeatures{Media: MediaImage | MediaAudio | MediaVideo, Streaming: StreamSSE}, true
	}
	return ModelFeatures{}, false
}

// HintObserved is emitted only after the actor classifies live input as a hint.
// Sender distinguishes local human acknowledgements from child/future clients.
type HintObserved struct {
	TurnID uint64
	Sender string
	Parts  PartList
}

func (HintObserved) observation() {}

// ObservationQueue is an opt-in lossless pull subscription. Actor publishing
// only appends to memory; slow terminal I/O cannot block the actor or HTTP.
// Close detaches it and releases retained observations.
type ObservationQueue interface {
	Wake() <-chan struct{}
	Drain() []Observation
	Close()
}

// ToolAdmission lets a low-level Prepare/Await caller use the same Agent-owned
// call-record barrier as Invoke and model dispatch. It is internal graph traffic.
type ToolAdmission struct {
	Context   context.Context
	Operation ToolOperation
	Reply     chan ToolData
}

func (ToolAdmission) inbound() {}
