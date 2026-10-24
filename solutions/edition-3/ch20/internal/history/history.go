// Package history owns the append-only log and every derived conversation view.
package history

import (
	"bufio"
	"encoding/json"
	"ensemble/internal/common"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
)

type history struct {
	// writers serializes admission staging, append, save, and journal rotation.
	// mu protects committed state and is never held across an external save callback.
	writers  sync.Mutex
	mu       sync.Mutex
	parent   common.Agent
	events   []common.Event
	through  common.Seq
	dialogue []common.DialogueRecord
	journal  string
	archive  string
	context  common.Context
}

func New(parent common.Agent) common.History {
	return &history{parent: parent, context: empty()}
}
func empty() common.Context {
	return common.Context{Turn: common.Idle, Dialogue: []common.Entry{}, Ephemera: common.PartList{}}
}
func (h *history) Parent() common.Agent   { return h.parent }
func clone[T any](v T) T                  { b, _ := json.Marshal(v); var c T; _ = json.Unmarshal(b, &c); return c }
func (h *history) Events() []common.Event { return h.EventsRange(0, ^common.Seq(0)) }
func (h *history) LastSeq() common.Seq {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.through
}
func (h *history) EventsSince(after common.Seq) []common.Event {
	return h.EventsRange(after, ^common.Seq(0))
}

// Appended events are immutable. Capture the range under lock, then clone
// outside it: a slow view must not block the actor's next append.
func (h *history) EventsRange(after, through common.Seq) []common.Event {
	h.mu.Lock()
	all, err := h.canonicalLocked()
	if err != nil {
		h.diagnoseArchive(err)
		all = h.events
	}
	i := sort.Search(len(all), func(i int) bool { return all[i].Seq > after })
	j := sort.Search(len(all), func(i int) bool { return all[i].Seq > through })
	if i > j {
		i = j
	}
	events := all[i:j]
	h.mu.Unlock()
	return clone(events)
}
func (h *history) Recent(since time.Time, minimum int) []common.Event {
	return h.RecentThrough(^common.Seq(0), since, minimum)
}
func (h *history) RecentThrough(through common.Seq, since time.Time, minimum int, eligible ...func(common.Event) bool) []common.Event {
	h.mu.Lock()
	all, err := h.canonicalLocked()
	if err != nil {
		h.diagnoseArchive(err)
		all = h.events
	}
	end := sort.Search(len(all), func(i int) bool { return all[i].Seq > through })
	events := all[:end]
	h.mu.Unlock()
	// The caller owns display eligibility; neither formatting nor scanning holds
	// History's append lock. Events in this captured suffix are immutable.
	include := func(common.Event) bool { return true }
	if len(eligible) > 0 && eligible[0] != nil {
		include = eligible[0]
	}
	start, count := end, 0
	for start > 0 {
		e := events[start-1]
		if count >= minimum && e.Time.Before(since) {
			break
		}
		start--
		if include(e) {
			count++
		}
	}
	return clone(events[start:end])
}
func (h *history) Context() common.Context { h.mu.Lock(); defer h.mu.Unlock(); return clone(h.context) }
func (h *history) prepare(e common.Event) (common.Event, error) {
	next := h.through + 1
	if next == 0 {
		return e, fmt.Errorf("sequence exhausted")
	}
	if e.Seq == 0 {
		e.Seq = next
	}
	if e.Seq < next {
		return e, fmt.Errorf("sequence must increase")
	}
	if e.Time.IsZero() {
		e.Time = time.Now().UTC()
	}
	if _, err := json.Marshal(e); err != nil {
		return e, err
	}
	if err := validate(e); err != nil {
		return e, err
	}
	return clone(e), nil
}
func (h *history) appendPrepared(e common.Event) {
	h.events = append(h.events, e)
	h.through = e.Seq
	if err := semantic(e, h.context); err != nil && h.parent != nil {
		h.parent.Parent().Logger().Printf("history: skip %s at %d: %v", e.Type, e.Seq, err)
	}
	Apply(&h.context, e)
}
func (h *history) Append(e common.Event) error {
	h.writers.Lock()
	defer h.writers.Unlock()
	h.mu.Lock()
	defer h.mu.Unlock()
	e, err := h.prepare(e)
	if err != nil {
		return err
	}
	if err := h.appendJournal(e); err != nil {
		return err
	}
	h.appendPrepared(e)
	return nil
}

// Hold the writer reservation across the staged save, without publishing the
// admission or blocking committed readers. A failed save leaves the old frontier
// intact; later writers and snapshot rotation cannot overtake this reservation.
func (h *history) AppendSaved(e common.Event, save func(func(io.Writer) error) error) error {
	h.writers.Lock()
	defer h.writers.Unlock()
	h.mu.Lock()
	e, err := h.prepare(e)
	var events []common.Event
	if err == nil {
		events, err = h.canonicalLocked()
		if err == nil {
			events = append(events[:len(events):len(events)], e)
		}
	}
	h.mu.Unlock()
	if err != nil {
		return err
	}
	if save != nil {
		if err = save(func(w io.Writer) error { return dumpEvents(w, events) }); err != nil {
			return err
		}
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if err = h.appendJournal(e); err != nil {
		return err
	}
	h.appendPrepared(e)
	return nil
}
func (h *history) Load(r io.Reader) error {
	h.writers.Lock()
	defer h.writers.Unlock()
	// Unbounded lines, optional version header, atomic publication on success.
	reader := bufio.NewReader(r)
	next := &history{parent: h.parent, context: empty()}
	n := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}
		if len(line) > 0 {
			n++
			if strings.TrimSpace(line) == "" {
				return fmt.Errorf("empty history line %d", n)
			}
			var header struct {
				Version *int `json:"log_version"`
			}
			if x := json.Unmarshal([]byte(line), &header); x != nil {
				return x
			}
			if header.Version != nil {
				if n != 1 || *header.Version != 1 {
					return fmt.Errorf("unsupported history version")
				}
			} else {
				var e common.Event
				if x := json.Unmarshal([]byte(line), &e); x != nil {
					return fmt.Errorf("history line %d: %w", n, x)
				}
				if e.Seq == 0 {
					return fmt.Errorf("missing sequence")
				}
				if x := next.Append(e); x != nil {
					return fmt.Errorf("history line %d: %w", n, x)
				}
			}
		}
		if err == io.EOF {
			break
		}
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events, h.context, h.through = next.events, next.context, next.through
	return nil
}
func (h *history) Dump(w io.Writer) error {
	h.mu.Lock()
	events, err := h.canonicalLocked()
	h.mu.Unlock()
	if err != nil {
		return err
	}
	return dumpEvents(w, events)
}
func dumpEvents(w io.Writer, events []common.Event) error {
	enc := json.NewEncoder(w)
	if err := enc.Encode(struct {
		Version int `json:"log_version"`
	}{1}); err != nil {
		return err
	}
	for _, e := range events {
		if err := enc.Encode(e); err != nil {
			return err
		}
	}
	return nil
}
func provenance(p common.Provenance) error {
	if p.Vendor.String() == "" || p.Surface.String() == "" {
		return fmt.Errorf("incomplete provenance")
	}
	return nil
}
func parts(ps common.PartList) error {
	for _, p := range ps {
		switch p.Type {
		case "text", "reasoning_summary":
		case "tool_decls":
			var delta common.ToolsData
			if err := json.Unmarshal(p.Data, &delta); err != nil {
				return err
			}
			if err := payloadValid(common.Event{Type: common.ToolsChanged, Tools: &delta}); err != nil {
				return err
			}
		case "blob", "image", "audio", "video":
			if p.Ref == nil {
				return fmt.Errorf("blob missing ref")
			}
		case "opaque":
			if p.From == nil || len(p.Data) == 0 {
				return fmt.Errorf("opaque missing provenance or data")
			}
		case "tool_call":
			if p.From == nil || p.Name == "" || len(p.Args) == 0 {
				return fmt.Errorf("incomplete tool call")
			}
		case "tool_result":
			if err := parts(p.Parts); err != nil {
				return err
			}
		case "redacted":
		default:
			return fmt.Errorf("unknown part %q", p.Type)
		}
		if p.Ref != nil {
			switch p.Ref.Kind {
			case 1, 2, 3:
			default:
				return fmt.Errorf("unknown ref kind")
			}
		}
		if p.From != nil {
			if err := provenance(*p.From); err != nil {
				return err
			}
		}
	}
	return nil
}
func validate(e common.Event) error {
	switch e.Type {
	case common.MessageReceived, common.HintReceived, common.TurnEnded, common.RequestSent, common.ResponseStarted, common.ResponseEnded, common.JobReported, common.JobKilled, common.ToolCalled, common.ToolReleased, common.ToolReturned, common.Redacted, common.ErrorOccurred, common.TurnInterrupted, common.SkillLoaded, common.SkillUnloaded, common.ToolsChanged, common.MicroHandoff:
		return nil
	default:
		return fmt.Errorf("unknown event type %q", e.Type)
	}
}
func payloadValid(e common.Event) error {
	n := 0
	for _, yes := range []bool{e.Message != nil, e.Request != nil, e.Response != nil, e.Tool != nil, e.Redact != nil, e.Error != nil, e.Job != nil, e.Skill != nil, e.Tools != nil, e.Handoff != nil} {
		if yes {
			n++
		}
	}
	if n != 1 {
		return fmt.Errorf("event requires exactly one typed payload")
	}
	switch e.Type {
	case common.SkillLoaded, common.SkillUnloaded:
		if e.Skill != nil && strings.TrimSpace(e.Skill.Name) != "" && e.Skill.ActiveSeq <= e.Seq && (e.Type == common.SkillUnloaded || strings.TrimSpace(e.Skill.Body) != "") {
			return nil
		}
	case common.ToolsChanged:
		if e.Tools != nil {
			seen := map[string]bool{}
			for _, t := range e.Tools.Added {
				if t.Name == "" || seen[t.Name] || !objectJSON(t.Schema) {
					return fmt.Errorf("invalid tool delta")
				}
				seen[t.Name] = true
			}
			for _, n := range e.Tools.Removed {
				if n == "" {
					return fmt.Errorf("invalid tool delta")
				}
				seen[n] = true
			}
			return nil
		}
	case common.MicroHandoff:
		if e.Handoff != nil && strings.TrimSpace(e.Handoff.Text) != "" {
			return nil
		}
	case common.MessageReceived, common.HintReceived, common.TurnEnded:
		if e.Message != nil {
			switch e.Message.Actor {
			case common.Human, common.Assistant, common.System, common.Tool:
			default:
				return fmt.Errorf("unknown actor")
			}
			return parts(e.Message.Parts)
		}
	case common.RequestSent:
		if e.Request != nil {
			return provenance(e.Request.To)
		}
	case common.ResponseStarted, common.ResponseEnded:
		if e.Response != nil {
			if err := provenance(e.Response.From); err != nil {
				return err
			}
			u := e.Response.Usage
			if u.Input < -1 || u.Output < -1 || u.CacheRead < -1 || u.CacheWrite < -1 {
				return fmt.Errorf("negative usage")
			}
			return parts(e.Response.Parts)
		}
	case common.JobReported:
		if e.Job != nil && e.Job.Handle > 0 {
			return nil
		}
		if e.Tool != nil {
			return parts(e.Tool.Parts)
		}
	case common.JobKilled:
		if e.Job != nil && e.Job.Handle > 0 && e.Job.Status == common.JobKilledStatus && (e.Job.Reason == "kill_job" || e.Job.Reason == "shutdown") {
			return nil
		}
	case common.ToolCalled, common.ToolReleased:
		if e.Tool != nil && e.Tool.Name != "" && objectJSON(e.Tool.Args) && (e.TurnID == 0 || e.Tool.CallID != "") {
			return nil
		}
	case common.ToolReturned:
		if e.Tool != nil && (e.Tool.Automatic || e.Tool.CallID != "") {
			return parts(e.Tool.Parts)
		}
	case common.Redacted:
		if e.Redact != nil {
			r := e.Redact
			for _, index := range r.Parts {
				if index < 0 {
					return fmt.Errorf("negative part selector")
				}
			}
			if r.From == 0 || r.To < r.From {
				return fmt.Errorf("invalid redaction span")
			}
			switch r.Level {
			case common.RedactResult, common.RedactTool, common.RedactDialogue, common.RedactSummary:
				return parts(r.Replacement)
			}
		}
	case common.ErrorOccurred, common.TurnInterrupted:
		if e.Error != nil {
			return nil
		}
	default:
		return fmt.Errorf("unknown event type %q", e.Type)
	}
	return fmt.Errorf("invalid %s payload", e.Type)
}
func add(a *common.Usage, b common.Usage) { a.Add(b) }

// Apply is total for known events. Dialogue changes and turn transitions are
// separate: a known event can add a fact without changing this turn's state.
func Apply(c *common.Context, e common.Event) {
	c.AsOf = e.Seq
	if semantic(e, *c) != nil {
		return
	}
	if e.Stale {
		if e.Type == common.ResponseEnded && e.Response != nil {
			add(&c.Usage, e.Response.Usage)
		}
		return
	}
	switch e.Type {
	case common.SkillLoaded:
		skill := *e.Skill
		if skill.Origin != "initial" && skill.ActiveSeq == 0 {
			skill.ActiveSeq = e.Seq
		}
		c.Held = append(c.Held, common.Entry{Kind: common.KindSkill, Seq: e.Seq, Actor: common.System, Skill: &skill, Parts: common.PartList{common.Text(skill.Body)}})
	case common.SkillUnloaded:
		for i := range c.ActiveSkills {
			if c.ActiveSkills[i].Name == e.Skill.Name {
				c.ActiveSkills[i].PendingUnload = true
			}
		}
		for i := range c.Held {
			if c.Held[i].Skill != nil && c.Held[i].Skill.Name == e.Skill.Name {
				c.Held[i].Skill.PendingUnload = true
			}
		}
	case common.ToolsChanged:
		first := !c.ToolsInitialized
		c.ToolsInitialized = true
		removed := map[string]bool{}
		for _, n := range e.Tools.Removed {
			removed[n] = true
		}
		for _, t := range e.Tools.Added {
			removed[t.Name] = true
		}
		kept := c.Tools[:0]
		for _, t := range c.Tools {
			if !removed[t.Name] {
				kept = append(kept, t)
			}
		}
		c.Tools = append(kept, e.Tools.Added...)
		if first {
			c.BaseTools = clone(c.Tools)
			c.BaseThrough = e.Seq
		}
		data, _ := json.Marshal(e.Tools)
		c.Held = append(c.Held, common.Entry{Kind: common.KindTools, Seq: e.Seq, Actor: common.System, Tools: e.Tools, Parts: common.PartList{{Type: "tool_decls", Data: data}}})
	case common.MicroHandoff:
		c.PendingHandoff, c.PendingHandoffSeq = e.Handoff.Text, e.Seq
	case common.MessageReceived, common.HintReceived:
		if e.Message.Actor == common.System {
			// The event log is append-only; this pending projection is a replacement.
			c.Ephemera = append(common.PartList(nil), e.Message.Parts...)
			if len(c.Ephemera) == 1 && c.Ephemera[0].Type == "text" && c.Ephemera[0].Text == "" {
				c.Ephemera = nil
			}
		} else {
			c.Dialogue = append(c.Dialogue, common.Entry{Seq: e.Seq, Actor: e.Message.Actor, Parts: e.Message.Parts})
		}
		if c.Turn == common.Idle {
			c.Turn = common.InputPending
		}
	case common.RequestSent:
		c.PreviousSentThrough = c.SentThrough
		c.SentThrough = e.Seq - 1
		c.Ephemera = common.PartList{}
		c.Turn = common.InFlight
	case common.ResponseEnded:
		c.Dialogue = append(c.Dialogue, common.Entry{Seq: e.Seq, Actor: common.Assistant, Parts: e.Response.Parts})
		add(&c.Usage, e.Response.Usage)
		if c.Turn == common.InFlight {
			c.Pending = nil
			for _, p := range e.Response.Parts {
				if p.Type == "tool_call" {
					c.Pending = append(c.Pending, p.CallID)
				}
			}
			if len(c.Pending) > 0 {
				c.Turn = common.ToolsPending
			} else {
				c.Turn = common.Idle
			}
		}
	case common.ToolCalled:
		// Bind only the most recent matching assistant call. Never fabricate a
		// call for manual/unpaired admissions or rewrite the canonical response.
	admission:
		for i := len(c.Dialogue) - 1; i >= 0; i-- {
			if c.Dialogue[i].Actor != common.Assistant {
				continue
			}
			for j, p := range c.Dialogue[i].Parts {
				if p.Type != "tool_call" || p.CallID != e.Tool.CallID || p.Name != e.Tool.Name {
					continue
				}
				if p.AdmissionSeq == 0 {
					c.Dialogue[i].Parts = append(common.PartList(nil), c.Dialogue[i].Parts...)
					c.Dialogue[i].Parts[j].AdmissionSeq = e.Seq
				}
				break admission
			}
		}
	case common.ToolReturned:
		if e.Tool.FleetCompletion {
			c.FleetCompletion = true
		}
		if e.Tool.Name == "keep_tool_results" && !e.Tool.IsError {
			c.KeepAt = c.SentThrough
			for _, entry := range c.Dialogue {
				if entry.Seq > c.SentThrough {
					continue
				}
				for _, p := range entry.Parts {
					if p.Type == "tool_call" {
						c.KeptBatch = entry.Seq
					}
				}
			}
		}
		if e.Tool.Automatic {
			break
		}
		c.Dialogue = append(c.Dialogue, common.Entry{Seq: e.Seq, Actor: common.Tool, Parts: common.PartList{{Type: "tool_result", CallID: e.Tool.CallID, Name: e.Tool.Name, Parts: e.Tool.Parts, IsError: e.Tool.IsError}}})
		if c.Turn == common.ToolsPending {
			for i, id := range c.Pending {
				if id == e.Tool.CallID {
					c.Pending = append(c.Pending[:i], c.Pending[i+1:]...)
					break
				}
			}
			if len(c.Pending) == 0 {
				c.Turn = common.InputPending
			}
		}
	case common.TurnEnded:
		c.Turn = common.Idle
		c.Pending = nil
	case common.TurnInterrupted:
		c.Turn = common.Interrupted
	case common.ErrorOccurred:
		// An error changes turn state, not projected causal facts. Only an
		// explicit redaction may remove or replace retained dialogue.
		if c.Turn == common.InFlight {
			c.Turn = common.Idle
		}
	case common.Redacted:
		redact(c, *e.Redact)
	}
	flushSurvivors(c)
}
func stub(p common.Part) common.Part {
	size := len([]byte(p.Text))
	if p.Type != "text" {
		b, _ := json.Marshal(p)
		size = len(b)
	}
	return common.Part{Type: "redacted", Stub: fmt.Sprintf("[redacted %d bytes]", size), Ref: p.Ref}
}
func redact(c *common.Context, r common.RedactData) {
	if r.Level == common.RedactTool {
		removeToolTraffic(c, r)
		return
	}
	if r.Level == common.RedactSummary {
		out := []common.Entry{}
		inserted := false
		for _, e := range c.Dialogue {
			if dialogue(e) && e.Seq >= r.From && e.Seq <= r.To {
				if !inserted {
					out = append(out, common.Entry{Seq: r.From, Actor: common.System, Parts: r.Replacement})
					inserted = true
				}
			} else {
				out = append(out, e)
			}
		}
		c.Dialogue = out
		return
	}
	for i := range c.Dialogue {
		e := &c.Dialogue[i]
		if !dialogue(*e) || e.Seq < r.From || e.Seq > r.To {
			continue
		}
		out := common.PartList{}
		for index, p := range e.Parts {
			selected := len(r.Parts) == 0
			for _, n := range r.Parts {
				if n == index {
					selected = true
				}
			}
			if !selected {
				out = append(out, p)
				continue
			}
			switch r.Level {
			case common.RedactResult:
				if p.Type == "tool_result" {
					encoded, _ := json.Marshal(p)
					if len(r.Parts) == 0 && len(encoded) < 256 {
						out = append(out, p)
						continue
					}
					p.Parts = common.PartList{common.Text("[tool result cleared]")}
					p.Stub = "cleared"
				}
				out = append(out, p)
			case common.RedactTool:
				if p.Type != "tool_call" && p.Type != "tool_result" {
					out = append(out, p)
				}
			case common.RedactDialogue:
				if p.Type == "text" || p.Type == "opaque" || p.Type == "blob" {
					out = append(out, stub(p))
				} else {
					out = append(out, p)
				}
			}
		}
		e.Parts = out
	}
}

// A published snapshot owns the old prefix. Reading its raw events for explicit
// display/export does not re-apply reducers or replace the current projection.
func (h *history) canonicalLocked() ([]common.Event, error) {
	if h.archive == "" {
		return h.events, nil
	}
	data, err := h.parent.Sandbox().ReadFile(h.archive)
	if err != nil {
		return nil, err
	}
	var saved common.SaveFile
	if err = json.Unmarshal(data, &saved); err != nil {
		return nil, err
	}
	out := saved.Log
	var last common.Seq
	if len(out) > 0 {
		last = out[len(out)-1].Seq
	}
	for _, event := range h.events {
		if event.Seq > last {
			out = append(out, event)
		}
	}
	return out, nil
}
func (h *history) diagnoseArchive(err error) {
	if h.parent != nil {
		h.parent.Parent().Logger().Printf("history archive: %v", err)
	}
}
