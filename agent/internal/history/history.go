// Package history owns the append-only facts and their disposable projection.
package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"ensemble/internal/common"
)

// Facts and their projection share the Agent lifetime, but not mutable parts
// after a rewrite. This owner is separate from the Engine's HTTP lifetime:
// offline loading does not manufacture a transport exchange.
type history struct {
	mu      sync.Mutex
	parent  common.Agent
	events  []common.Event
	context common.Context
	asOf    uint64
	journal string
}

// An empty history starts Idle. Explicit initialization keeps the zero string
// from becoming an accidental fifth turn state on the first replayed event.
func New(parent common.Agent) common.History {
	if parent == nil {
		panic("History requires Agent")
	}
	return &history{parent: parent, context: common.Context{Turn: "idle"}}
}

// Agent returns the required owning Agent rather than a copied service bundle.
func (h *history) Agent() common.Agent { return h.parent }

// Readers receive their own snapshot: live clients cannot race with reduction.
func (h *history) Context() *common.Context {
	h.mu.Lock()
	defer h.mu.Unlock()
	data, _ := json.Marshal(h.context)
	var snapshot common.Context
	json.Unmarshal(data, &snapshot)
	return &snapshot
}

// Append is the single live entry point. Sequence assignment belongs here so
// parsers and human clients cannot disagree about ordering.
func (h *history) Append(e common.Event) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	e.Seq = h.asOf + 1
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
	if err := validateFor(&h.context, e); err != nil {
		return err
	}
	if err := h.journalEvent(e); err != nil {
		return err
	}
	h.events = append(h.events, e)
	apply(&h.context, e)
	h.asOf = e.Seq
	return nil
}

// Load stages the entire replay: refusing malformed input must not leave a
// half-loaded context. Headerless logs are current; unknown versions are not.
func (h *history) Load(r io.Reader) error {
	h.mu.Lock()
	defer h.mu.Unlock()
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
				// Version distinguishes a current log header from unsupported or absent metadata.
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
				if !knownEvent(e.Type) {
					return fmt.Errorf("unknown event type %q", e.Type)
				}
				// Preserve recorded numbers instead of renumbering imported facts.
				// Gaps are harmless; duplicates or reversal make redaction ambiguous.
				if e.Seq == 0 || (len(next.events) > 0 && e.Seq <= next.events[len(next.events)-1].Seq) {
					return errors.New("log seq must ascend")
				}
				next.events = append(next.events, e)
				next.replay(e)
				next.asOf = e.Seq
			}
			first = false
		}
		if err == io.EOF {
			break
		}
	}
	h.events, h.context, h.asOf = next.events, next.context, next.asOf
	return nil
}

// Dump serializes original events, never the redacted projection. A fresh
// process must recover both the historical evidence and the current view.
// Encoding the version separately also keeps the event vocabulary frozen.
func (h *history) Dump(w io.Writer) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	enc := json.NewEncoder(w)
	if err := enc.Encode(struct {
		// Version distinguishes a current log header from unsupported or absent metadata.
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
	// Require one matching payload, except for successful turn completion.
	count := 0
	for _, present := range []bool{e.Message != nil, e.Request != nil, e.Response != nil, e.Tool != nil, e.Redact != nil, e.Error != nil, e.Turn != nil, e.Job != nil, e.Skill != nil, e.Handoff != nil, e.Tools != nil, e.Populated != nil, e.Depopulated != nil, e.Compactor != nil, e.CompactionOutcome != nil, e.Recall != nil} {
		if present {
			count++
		}
	}
	// Completion is also a fact when there is no failure to attach. Keep the
	// existing error payload for failed endings and reject other payloads below.
	if e.Type == "turn_ended" && count == 0 {
		return nil
	}
	if count != 1 {
		return errors.New("event requires one payload")
	}
	var parts []common.Part
	switch e.Type {
	case "message_received", "hint_received", "fleet_notice":
		if e.Message == nil {
			return errors.New("message payload missing")
		}
		switch e.Message.Actor {
		case common.Human, common.AgentActor, common.System, common.Tool:
		default:
			return fmt.Errorf("unknown actor %q", e.Message.Actor)
		}
		parts = e.Message.Parts
	case "band_populated":
		if e.Populated == nil || e.Populated.Band < common.BandSoul || e.Populated.Band > common.BandSession {
			return errors.New("invalid memory population")
		}
	case "band_depopulated":
		if e.Depopulated == nil || e.Depopulated.Band < common.BandSoul || e.Depopulated.Band > common.BandSession {
			return errors.New("invalid memory retirement")
		}
	case "compaction_outcome":
		if e.CompactionOutcome == nil || e.CompactionOutcome.TaskID == "" {
			return errors.New("invalid compaction outcome")
		}
		switch e.CompactionOutcome.State {
		case "committed":
		case "failed", "canceled", "invalidated", "abandoned":
			if strings.TrimSpace(e.CompactionOutcome.Reason) == "" {
				return errors.New("compaction failure requires reason")
			}
		default:
			return errors.New("invalid compaction outcome state")
		}
	case "compactor_launched":
		if e.Compactor == nil || e.Compactor.Band < common.BandMemory || e.Compactor.Band > common.BandSession {
			return errors.New("invalid compressor launch")
		}
	case "skill_loaded":
		if e.Skill == nil || e.Skill.Name == "" {
			return errors.New("skill payload missing")
		}
	case "recall_attached":
		if e.Recall == nil {
			return errors.New("recall payload missing")
		}
		parts = e.Recall.Parts
		for _, p := range parts {
			if p.Type != "text" {
				return errors.New("recall must contain text")
			}
		}
	case "micro_handoff":
		if e.Handoff == nil || strings.TrimSpace(e.Handoff.Text) == "" {
			return errors.New("handoff text missing")
		}
	case "tools_changed":
		if e.Tools == nil {
			return errors.New("tools payload missing")
		}
		return validateDeclarations(e.Tools.Added)
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
	case "tool_called", "tool_returned", "job_report", "job_finished":
		if e.Tool == nil {
			return errors.New("tool payload missing")
		}
		parts = e.Tool.Parts
	case "job_killed":
		if e.Job == nil || e.Job.Status != "killed" {
			return errors.New("killed job payload missing")
		}
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
	case "turn_ended":
		if e.Turn == nil {
			if e.Error == nil {
				return errors.New("turn payload missing")
			}
			break
		}
		if timing := e.Turn.Timing; timing != nil {
			if timing.Started.IsZero() || timing.TotalMS < 0 || timing.ModelMS < 0 || timing.ToolMS < 0 || timing.OtherMS < 0 || timing.Calls < 0 || timing.Requests < 0 {
				return errors.New("invalid turn timing")
			}
			if timing.ModelMS > timing.TotalMS || timing.ToolMS > timing.TotalMS-timing.ModelMS || timing.OtherMS != timing.TotalMS-timing.ModelMS-timing.ToolMS {
				return errors.New("turn timing does not reconcile")
			}
			if timing.OtherPhasesMS != nil {
				remaining := timing.OtherMS
				for name, milliseconds := range timing.OtherPhasesMS {
					if name == "" || milliseconds < 0 || milliseconds > remaining {
						return errors.New("invalid turn timing phase")
					}
					remaining -= milliseconds
				}
				if remaining != 0 {
					return errors.New("turn timing phases do not reconcile")
				}
			}
		}
	case "error_occurred", "turn_interrupted":
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
	if p.Vendor < common.Anthropic || p.Vendor > common.Gemini || p.Surface < common.Messages || p.Surface > common.Responses || p.Model == "" {
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
		case "tool_decls":
			if err := validateDeclarations(p.Added); err != nil {
				return err
			}
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

// EventsAfter snapshots only the requested suffix. Append-only slices still
// require locking: append can replace their backing array while readers run.
func (h *history) EventsAfter(seq uint64) []common.Event {
	h.mu.Lock()
	defer h.mu.Unlock()
	start := len(h.events)
	for start > 0 && h.events[start-1].Seq > seq {
		start--
	}
	data, _ := json.Marshal(h.events[start:])
	var events []common.Event
	json.Unmarshal(data, &events)
	return events
}

// LastSeq lets dispatch retain provenance without copying the whole event log.
func (h *history) LastSeq() uint64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.asOf
}

// Unknown vocabulary is refused; malformed familiar facts are diagnosed and
// skipped atomically. Validation never mutates the projected context.
func knownEvent(kind string) bool {
	switch kind {
	case "fleet_notice", "message_received", "hint_received", "request_sent", "response_started", "response_ended", "tool_called", "tool_returned", "job_report", "job_finished", "job_killed", "redacted", "error_occurred", "turn_ended", "turn_interrupted", "skill_loaded", "micro_handoff", "tools_changed", "band_populated", "band_depopulated", "compactor_launched", "compaction_outcome", "recall_attached":
		return true
	}
	return false
}
func validateDeclarations(declarations []common.ToolDeclaration) error {
	for _, d := range declarations {
		if d.Name == "" || !json.Valid(d.Schema) {
			return errors.New("invalid tool declaration")
		}
	}
	return nil
}
func validateFor(c *common.Context, e common.Event) error {
	if err := validate(e); err != nil {
		return err
	}
	if err := validateCompaction(c, e); err != nil {
		return err
	}
	if e.Redact != nil {
		if e.Redact.From > e.Redact.To {
			return errors.New("reversed redaction span")
		}
		for _, entry := range c.Dialogue {
			if entry.Seq >= e.Redact.From && entry.Seq <= e.Redact.To {
				return nil
			}
		}
		return errors.New("redaction names no surviving entry")
	}
	return nil
}
func (h *history) replay(e common.Event) {
	if err := validateFor(&h.context, e); err != nil {
		h.parent.Ensemble().Logf("restore: skipped event %d: %v", e.Seq, err)
		return
	}
	apply(&h.context, e)
}
