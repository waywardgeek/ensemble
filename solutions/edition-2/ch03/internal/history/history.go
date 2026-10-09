// Package history owns the append-only facts and their disposable projection.
package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"ensemble/internal/common"
)

// Facts and their projection share the Agent lifetime, but not mutable parts
// after a rewrite. This owner is separate from the Engine's HTTP lifetime:
// offline loading does not manufacture a transport exchange.
type history struct {
	parent  common.Agent
	events  []common.Event
	context common.Context
}

// An empty history starts Idle. Explicit initialization keeps the zero string
// from becoming an accidental fifth turn state on the first replayed event.
func New(parent common.Agent) common.History {
	if parent == nil {
		panic("History requires Agent")
	}
	return &history{parent: parent, context: common.Context{Turn: "idle"}}
}
func (h *history) Agent() common.Agent      { return h.parent }
func (h *history) Context() *common.Context { return &h.context }

// Append is the single live entry point. Sequence assignment belongs here so
// parsers and human clients cannot disagree about ordering.
func (h *history) Append(e common.Event) error {
	e.Seq = uint64(len(h.events) + 1)
	if len(h.events) > 0 {
		e.Seq = h.events[len(h.events)-1].Seq + 1
	}
	// Timestamps are capture metadata only. Loaded events retain theirs;
	// ordering and deterministic rendering depend exclusively on Seq.
	e.Time = time.Now().UTC()
	// Some vendors omit call IDs. Capture one identity before either dispatch
	// or replay sees the call; two calls must never share the empty-string key.
	if e.Response != nil {
		for i := range e.Response.Parts {
			p := &e.Response.Parts[i]
			if p.Type == "tool_call" && p.CallID == "" {
				p.CallID = fmt.Sprintf("call_%d_%d", e.Seq, i)
			}
		}
	}
	if err := validate(e); err != nil {
		return err
	}
	h.events = append(h.events, e)
	apply(&h.context, e)
	return nil
}

// Load stages the entire replay: refusing malformed input must not leave a
// half-loaded context. Headerless logs are current; unknown versions are not.
func (h *history) Load(r io.Reader) error {
	next := history{parent: h.parent, context: common.Context{Turn: "idle"}}
	// ReadBytes avoids Scanner's small token ceiling: one event may contain
	// a large tool result. EOF may still bring the final unterminated line.
	reader := bufio.NewReader(r)
	first := true
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}
		if len(strings.TrimSpace(string(line))) > 0 {
			// A pointer distinguishes an absent header from an explicitly invalid
			// version zero. The header is metadata and consumes no event Seq.
			var header struct {
				Version *int `json:"log_version"`
			}
			if decodeErr := json.Unmarshal(line, &header); decodeErr != nil {
				return decodeErr
			}
			if header.Version != nil {
				if !first || *header.Version != 1 {
					return errors.New("unsupported log_version")
				}
			} else {
				var e common.Event
				if err := json.Unmarshal(line, &e); err != nil {
					return err
				}
				if err := validate(e); err != nil {
					return err
				}
				// Preserve recorded numbers instead of renumbering imported facts.
				// Gaps are harmless; duplicates or reversal make redaction ambiguous.
				if e.Seq == 0 || (len(next.events) > 0 && e.Seq <= next.events[len(next.events)-1].Seq) {
					return errors.New("log seq must ascend")
				}
				next.events = append(next.events, e)
				apply(&next.context, e)
			}
			first = false
		}
		if err == io.EOF {
			break
		}
	}
	h.events, h.context = next.events, next.context
	return nil
}

// Dump serializes original events, never the redacted projection. A fresh
// process must recover both the historical evidence and the current view.
// Encoding the version separately also keeps the event vocabulary frozen.
func (h *history) Dump(w io.Writer) error {
	enc := json.NewEncoder(w)
	if err := enc.Encode(struct {
		Version int `json:"log_version"`
	}{1}); err != nil {
		return err
	}
	for _, e := range h.events {
		if err := enc.Encode(e); err != nil {
			return err
		}
	}
	return nil
}

// Validation is shared by live append and disk load. The reducer can therefore
// keep known but non-transitioning event/state pairs harmless while unknown
// vocabulary fails at the boundary instead of being silently skipped.
func validate(e common.Event) error {
	// A tagged union without this check could accept contradictory payloads.
	// Require both one payload and the payload appropriate for the tag.
	count := 0
	for _, present := range []bool{e.Message != nil, e.Request != nil, e.Response != nil, e.Tool != nil, e.Redact != nil, e.Error != nil} {
		if present {
			count++
		}
	}
	if count != 1 {
		return errors.New("event requires one payload")
	}
	var parts []common.Part
	switch e.Type {
	case "message_received":
		if e.Message == nil {
			return errors.New("message payload missing")
		}
		switch e.Message.Actor {
		case common.Human, common.AgentActor, common.System, common.Tool:
		default:
			return fmt.Errorf("unknown actor %q", e.Message.Actor)
		}
		parts = e.Message.Parts
	case "request_sent":
		if e.Request == nil {
			return errors.New("request payload missing")
		}
		return provenance(e.Request.To)
	case "response_started", "response_ended":
		if e.Response == nil {
			return errors.New("response payload missing")
		}
		if err := provenance(e.Response.From); err != nil {
			return err
		}
		parts = e.Response.Parts
	case "tool_called", "tool_returned":
		if e.Tool == nil {
			return errors.New("tool payload missing")
		}
		parts = e.Tool.Parts
	case "redacted":
		if e.Redact == nil {
			return errors.New("redact payload missing")
		}
		switch e.Redact.Level {
		case "redact_result", "redact_tool", "redact_dialogue", "redact_summary":
		default:
			return errors.New("unknown redaction level")
		}
		parts = e.Redact.Replacement
	case "error_occurred":
		if e.Error == nil {
			return errors.New("error payload missing")
		}
	default:
		return fmt.Errorf("unknown event type %q", e.Type)
	}
	return validateParts(parts)
}

// A known vendor with no model identity is still unusable provenance.
// Never infer an old producer from the Agent's current target configuration.
func provenance(p common.Provenance) error {
	if p.Vendor < common.Anthropic || p.Vendor > common.Gemini || p.Surface < common.Messages || p.Surface > common.GenerateContent || p.Model == "" {
		return errors.New("missing or invalid provenance")
	}
	return nil
}

// Tool results nest parts, so validation follows that nesting. A bad Ref or
// unknown part hidden inside a result must not become valid merely because
// the outer event and actor are familiar.
func validateParts(parts []common.Part) error {
	for _, p := range parts {
		switch p.Type {
		case "text":
		case "blob":
			if p.Ref.Kind < common.RefPath || p.Ref.Kind > common.RefHandle || p.Ref.Locator == "" {
				return errors.New("blob requires a valid Ref")
			}
		case "opaque", "tool_call":
			if err := provenance(p.From); err != nil {
				return err
			}
		case "tool_result":
			if err := validateParts(p.Parts); err != nil {
				return err
			}
		// Text-only stubs have no locator. When one exists, it obeys the same
		// nonzero-kind rule as a blob and remains available for later retrieval.
		case "redacted":
			if p.Ref.Kind != 0 && (p.Ref.Kind < common.RefPath || p.Ref.Kind > common.RefHandle || p.Ref.Locator == "") {
				return errors.New("invalid redacted Ref")
			}
		default:
			return fmt.Errorf("unknown part type %q", p.Type)
		}
	}
	return nil
}
