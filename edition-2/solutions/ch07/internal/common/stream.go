package common

import (
	"encoding/json"
	"errors"
	"net/http"
)

// DeltaKind distinguishes display channels without changing durable part types.
type DeltaKind uint8

const (
	// DeltaThinking displays provider reasoning separately from the reply.
	DeltaThinking DeltaKind = iota + 1
	// DeltaText displays assistant reply text.
	DeltaText
	// DeltaToolCall displays a tool name followed by its argument fragments.
	DeltaToolCall
)

// MarshalJSON refuses an unset kind rather than mislabeling a display channel.
func (k DeltaKind) MarshalJSON() ([]byte, error) {
	switch k {
	case DeltaThinking:
		return json.Marshal("thinking")
	case DeltaText:
		return json.Marshal("text")
	case DeltaToolCall:
		return json.Marshal("tool_call")
	}
	return nil, errors.New("invalid delta kind")
}

// Stream describes independently supported incremental delivery channels.
type Stream uint8

const (
	// StreamText permits incremental reply text.
	StreamText Stream = 1 << iota
	// StreamThinking permits incremental provider reasoning.
	StreamThinking
	// StreamToolArgs permits incremental function arguments.
	StreamToolArgs
	// StreamAll enables every currently defined delivery channel.
	StreamAll = StreamText | StreamThinking | StreamToolArgs
)

// StreamCallbacks are response notifications, not dependencies. Callbacks run
// synchronously; consumers must queue slow work and preserve notification order.
type StreamCallbacks struct {
	// OnDelta receives a fragment with a response-scoped identity chosen by Parse.
	OnDelta func(partID uint64, kind DeltaKind, chunk string)
	// OnEvent receives the finalized fact before any of its part finals.
	OnEvent func(Event)
	// OnPartFinal repeats the exact identity used by that part's deltas.
	OnPartFinal func(partID uint64, part Part)
	// OnFrame receives each SSE data payload, or the complete plain JSON body.
	OnFrame func(eventType string, data []byte)
}

// Parser normalizes either delivery mode through the same finalization path.
type Parser interface {
	// Parse consumes the open response body; its caller owns closing the body.
	Parse(*http.Response, StreamCallbacks) error
}
