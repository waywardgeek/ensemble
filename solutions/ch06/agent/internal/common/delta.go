package common

// The content taxonomy of a stream.
//
// A chunk of thinking, a chunk of reply text, and a chunk of a tool call's
// arguments all arrive through the same seam and all look like a string. What
// distinguishes them is what the reader must DO with them: dim the thinking,
// show the text, and never, ever act on half-parsed arguments.

import (
	"encoding/json"
	"fmt"
)

// DeltaKind says what sort of content a chunk is.
//
// An enum rather than a string because every consumer must handle every case
// exhaustively: the terminal picks an ANSI colour per kind, a GUI picks a CSS
// class per kind. That is the rule — enum when code must switch on all of
// them, string when values are only ever compared and the set is open.
//
// The constants start at iota+1 so the zero value is INVALID. A delta whose
// kind was never set is a bug at the site that built it, and MarshalJSON
// refuses to write one rather than emitting a plausible default.
type DeltaKind uint8

const (
	// DeltaThinking is reasoning content. It is shown dimmed and is never
	// mistaken for the reply.
	DeltaThinking DeltaKind = iota + 1

	// DeltaText is assistant reply text: the part a user reads.
	DeltaText

	// DeltaToolCall is a tool call being spelled out, name and arguments.
	//
	// Streaming these is for DISPLAY ONLY. The arguments are incomplete JSON
	// until the part is finalized, so nothing may be executed from a delta.
	// The temptation to start early is exactly how an agent runs a tool call
	// with half its arguments.
	DeltaToolCall
)

// deltaKindNames is the wire spelling, and the only place it is written.
var deltaKindNames = map[DeltaKind]string{
	DeltaThinking: "thinking",
	DeltaText:     "text",
	DeltaToolCall: "tool_call",
}

func (k DeltaKind) String() string {
	if s, ok := deltaKindNames[k]; ok {
		return s
	}
	return fmt.Sprintf("DeltaKind(%d)", uint8(k))
}

// MarshalJSON writes the wire spelling, and REFUSES an unset kind.
//
// A loud failure here is worth more than a tidy log. The alternative is a
// stream of deltas that all claim to be text because zero happened to mean
// something.
func (k DeltaKind) MarshalJSON() ([]byte, error) {
	s, ok := deltaKindNames[k]
	if !ok {
		return nil, fmt.Errorf("cannot marshal delta kind %d: kind was never set (zero is invalid by design)", uint8(k))
	}
	return json.Marshal(s)
}

func (k *DeltaKind) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	for kind, name := range deltaKindNames {
		if name == s {
			*k = kind
			return nil
		}
	}
	return fmt.Errorf("unknown delta kind %q", s)
}

// Stream is the set of delta kinds a model can actually stream.
//
// A bitmask and not a bool because vendors ship this capability in pieces.
// Streaming text while NOT streaming tool arguments is a real, current
// configuration, not a hypothetical one, and a single bool would have forced
// a choice between losing text streaming and inventing tool-argument chunks
// that never arrived.
type Stream uint8

const (
	StreamText Stream = 1 << iota
	StreamThinking
	StreamToolArgs
)

// StreamAll is everything a vendor could stream.
const StreamAll = StreamText | StreamThinking | StreamToolArgs

func (s Stream) Has(want Stream) bool { return s&want != 0 }

// KindOK reports whether kind may be streamed under s.
func (s Stream) KindOK(kind DeltaKind) bool {
	switch kind {
	case DeltaText:
		return s.Has(StreamText)
	case DeltaThinking:
		return s.Has(StreamThinking)
	case DeltaToolCall:
		return s.Has(StreamToolArgs)
	}
	return false
}

// StreamingFor returns the delta kinds that may be streamed for cfg.
//
// Zero means "send a non-streaming request". Two things can say no: the
// caller, through Config.DisableStreaming, and the model table.
//
// AN UNKNOWN MODEL GETS NO STREAMING, AND THAT IS NOT AN ERROR. This is the
// deliberate opposite of what checkMedia does for an unknown model, so the
// difference is worth being precise about:
//
//	A guess about CONTENT can corrupt the conversation. Send an image to a
//	model that cannot see and the content is silently dropped; that must be
//	a loud refusal, because nothing downstream can recover it.
//
//	A guess about DELIVERY cannot. Stream or do not stream and the same
//	response arrives either way, in a different number of pieces.
//
// The practical half of the argument matters as much as the principled half:
// model names are an OPEN set. New ones appear weekly, dated variants appear
// daily, and a framework that refuses to send any request at all until its
// table has heard of your model is a framework nobody can use. Refusing to
// guess is right. Refusing to work is not.
func StreamingFor(cfg Config) Stream {
	if cfg.DisableStreaming {
		return 0
	}
	features, ok := LookupModel(cfg.Model)
	if !ok {
		return 0
	}
	return features.Stream
}
