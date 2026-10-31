package common

import (
	"encoding/json"
	"errors"
)

// Progress is deliberately lossy. Completion is a separate per-request value;
// neither a shared idle event nor the last displayed text acknowledges a caller.
type Result struct {
	// Text is this request's final assistant reply, excluding intermediate tool
	// narration.
	Text string
	// Err reports this request's failure, including interrupted, stopped and save
	// errors.
	Err error
}

// ErrStopped distinguishes Agent shutdown from a failed or interrupted turn.
var ErrStopped = errors.New("agent stopped")

// ErrInterrupted ends this request without ending the Agent lifetime.
var ErrInterrupted = errors.New("turn interrupted")

// ErrRoundLimit reports a refused tool batch after the configured budget.
var ErrRoundLimit = errors.New("tool round limit reached")

// DefaultMaxToolRounds permits 200 dispatched batches when configuration is zero.
const DefaultMaxToolRounds = 200

// Observation is sealed progress vocabulary; delivery may be dropped.
type Observation interface{ observation() }

// Observer receives progress synchronously and must return without blocking.
type Observer interface {
	// Observe must return promptly; queue slow display work outside the actor.
	Observe(Observation)
}

// PartDelta is transient display content; it is never a durable history fact.
type PartDelta struct {
	// PartID correlates transient content with its final display part.
	PartID uint64 `json:"part_id"`
	// Chunk contains transient text; a hint acknowledgement begins with hint:.
	Chunk string `json:"chunk"`
}

// PartFinal publishes a stored part with its history and display identities.
type PartFinal struct {
	// Seq identifies the stored history fact underlying the finalized content.
	Seq uint64 `json:"seq"`
	// PartID correlates transient content with its final display part.
	PartID uint64 `json:"part_id"`
	// Part exposes the complete neutral content, including calls and tool results.
	Part Part `json:"part"`
}

// StateChanged reports a reducer transition after it has taken effect.
type StateChanged struct {
	// From is the turn state before this observed reduction.
	From string `json:"from"`
	// To is the turn state after this observed reduction.
	To string `json:"to"`
}

func (PartDelta) observation()    {}
func (PartFinal) observation()    {}
func (StateChanged) observation() {}

// Capture does not classify hints: the reducer sees the actual turn state.
type Inbound interface{ inbound() }

// UserMessage starts an idle turn or becomes a hint during an active turn.
type UserMessage struct {
	// Text holds visible provider or user content, including an explicitly empty string.
	Text string
}

// Hint carries live human steering; idle Agents treat it as a new turn.
type Hint struct {
	// Text holds visible provider or user content, including an explicitly empty string.
	Text string
}

// Interrupt ends the current turn while keeping the Agent and other jobs alive.
type Interrupt struct{}

func (UserMessage) inbound() {}
func (Hint) inbound()        {}
func (Interrupt) inbound()   {}

// JSON is boundary dispatch, so it stays with these sealed value types.
func (p PartDelta) MarshalJSON() ([]byte, error) {
	type value PartDelta
	return json.Marshal(struct {
		// Kind selects the input or observation protocol variant.
		Kind string `json:"observation"`
		value
	}{"part_delta", value(p)})
}

// MarshalJSON writes the canonical observation discriminator alongside its payload.
func (p PartFinal) MarshalJSON() ([]byte, error) {
	type value PartFinal
	return json.Marshal(struct {
		// Kind selects the input or observation protocol variant.
		Kind string `json:"observation"`
		value
	}{"part_final", value(p)})
}

// MarshalJSON writes the canonical observation discriminator alongside its payload.
func (p StateChanged) MarshalJSON() ([]byte, error) {
	type value StateChanged
	return json.Marshal(struct {
		// Kind selects the input or observation protocol variant.
		Kind string `json:"observation"`
		value
	}{"state_changed", value(p)})
}

// Model features are data shared by clients and renderers; lookup policy lives
// with rendering, rather than making common an application service package.
type Media uint8

const (
	// MediaImage permits image parts on a supported model.
	MediaImage Media = 1 << iota
	// MediaAudio permits audio parts on a supported model.
	MediaAudio
	// MediaVideo permits video parts on a supported model.
	MediaVideo
	// MediaDocument permits PDF document parts on a supported model.
	MediaDocument
)

// ModelFeatures describes the explicitly known media input capabilities.
type ModelFeatures struct {
	// Media is the explicitly supported input-category bit set.
	Media Media
}
