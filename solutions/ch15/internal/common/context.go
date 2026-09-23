package common

// The context: the vendor-independent state you get by replaying the log.
//
// Derived, reconstructible, disposable. The log is the truth; this is what the
// truth means right now. Three consumers, three needs: the renderer reads the
// context, the GUI reads the log, the auditor reads the log.

import (
	"encoding/json"
	"fmt"
)

type TurnState uint8

const (
	Idle TurnState = iota + 1
	InputPending
	InFlight
	ToolsPending
	Interrupted
)

var turnNames = map[TurnState]string{
	Idle:         "idle",
	InputPending: "input_pending",
	InFlight:     "in_flight",
	ToolsPending: "tools_pending",
	Interrupted:  "interrupted",
}

func (t TurnState) String() string { return turnNames[t] }

// Entry is one turn of dialogue.
//
// Seq is the log position of the event that produced this entry, and it is
// load-bearing rather than decorative: RedactData names a SPAN of Seq numbers,
// so without it a replayed context has nothing for a span to match against.
// It is one fixed-size field per entry, and entries are already bounded by the
// compaction policy, so it does not reintroduce unbounded growth.
type Entry struct {
	Seq   Seq       `json:"seq"`
	Actor Actor     `json:"actor"`
	Kind  EntryKind `json:"kind"`
	Parts PartList  `json:"parts"`
}

// EntryKind says why an entry is in the context and which verb removes it.
//
// Chapter 15's removal rule: every entry that is not dialogue is removed only
// by its own verb. Tool clearing (redaction, micro_handoff) walks Dialogue
// entries and nothing else, so anything that must outlive tool clearing is
// its own kind, and it is safe by construction rather than by care.
type EntryKind uint8

const (
	KindDialogue EntryKind = iota + 1 // prompts, hints, output, tool parts
	KindHandoff                       // from MicroHandoff
	KindSkill                         // from SkillLoaded
	KindTools                         // from ToolsChanged
)

var entryKindNames = map[EntryKind]string{
	KindDialogue: "dialogue",
	KindHandoff:  "handoff",
	KindSkill:    "skill",
	KindTools:    "tools",
}

func (k EntryKind) String() string { return entryKindNames[k] }

func (k EntryKind) MarshalJSON() ([]byte, error) {
	s, ok := entryKindNames[k]
	if !ok {
		return nil, fmt.Errorf("refusing to marshal invalid entry kind %d", uint8(k))
	}
	return json.Marshal(s)
}

func (k *EntryKind) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	for v, name := range entryKindNames {
		if NormalizeName(name) == NormalizeName(s) {
			*k = v
			return nil
		}
	}
	return fmt.Errorf("unknown entry kind %q: refusing to load this context", s)
}

// Context holds no field that grows without bound.
//
// An earlier draft had `Redacted map[Seq]bool` here, to remember which events
// had been superseded. It was wrong twice: it grew forever, and it was
// redundant, because the log already records every Redacted event. The context
// does not need to remember that a redaction HAPPENED; it needs to hold the
// content that redaction PRODUCED.
//
// Dialogue grows too, but it is bounded by a policy (compaction) rather than
// by its shape, and the shape survives compaction unchanged. Dialogue grows
// and has a plan; the map grew and had none.
//
// Note what is absent: there is nowhere to put system prompt text. That is
// structural. The system prompt is an OUTPUT, computed by the renderer from
// Config. Equally absent: role, content, tool_use_id, assistant.
type Context struct {
	Turn     TurnState `json:"turn"`
	Dialogue []Entry   `json:"dialogue"`
	Ephemera PartList  `json:"ephemera"` // pending; delivered once, then cleared
	Usage    Usage     `json:"usage"`    // running totals, vendor-normalized
	// Held is where a survivor waits while tool calls are outstanding. A
	// skill loaded by one of two parallel calls must not land between the
	// batch's results: a vendor wants every result right after the calls
	// that asked for them, and a checkpoint that cleared tool parts while a
	// call was still running would orphan that call's result. So survivors
	// land when the batch completes. Bounded by one batch, and empty again
	// the moment the last result arrives.
	Held []Entry `json:"held,omitempty"`
}

func NewContext() *Context {
	return &Context{Turn: Idle}
}

// Apply advances the context by exactly one event.
//
//	newContext = Apply(context, event)
//
// That notation describes INFORMATION FLOW: everything needed to advance the
// context is in the context plus one event. It is not a demand for value
// semantics. A pointer receiver mutating in place is the right Go
// implementation — a Context full of slices copied by value gives you two
// contexts sharing one backing array, and that bug is indistinguishable from
// renderer non-determinism.
//
// Classification happens HERE, not at the capture site. The same arriving
// bytes mean different things depending on turn state, and only the reducer
// holds the state that makes the decision correct.
func (c *Context) Apply(e Event) error {
	switch e.Type {
	case MessageReceived:
		if e.Message == nil {
			return fmt.Errorf("seq %d: message_received with no message payload", e.Seq)
		}
		// The classification this chapter turns on. A System message is
		// volatile data — the time, the screen, live status — and it is
		// PENDING, not dialogue. It is delivered exactly once and never
		// becomes part of the conversation. A Human message is an ordinary
		// prompt and starts a turn.
		if e.Message.Actor == ActorSystem {
			c.Ephemera = append(c.Ephemera, e.Message.Parts...)
			return nil // deliberately NOT a turn transition
		}
		c.Dialogue = append(c.Dialogue, Entry{Seq: e.Seq, Actor: e.Message.Actor, Kind: KindDialogue, Parts: e.Message.Parts})
		if c.Turn == Idle {
			c.Turn = InputPending
		}
		c.flushHeld()

	case RequestSent:
		// Ephemera are consumed by the act of sending. Delivered once, then
		// gone: a stale timestamp is not stale data, it is a lie.
		c.Ephemera = nil
		if c.Turn == InputPending {
			c.Turn = InFlight
		}

	case ResponseStarted:
		// Nothing to record. It exists so that Chapter 3 has somewhere to
		// hang first-token latency, and so a GUI can show a spinner.

	case ResponseEnded:
		if e.Response == nil {
			return fmt.Errorf("seq %d: response_ended with no response payload", e.Seq)
		}
		c.Dialogue = append(c.Dialogue, Entry{Seq: e.Seq, Actor: ActorAgent, Kind: KindDialogue, Parts: e.Response.Parts})
		c.Usage.Add(e.Response.Usage)
		if hasToolCall(e.Response.Parts) {
			c.Turn = ToolsPending
		} else if c.Turn == InFlight {
			c.Turn = Idle
		}

	case ToolCalled:
		// The engine dispatched a call. The call itself is already in the
		// dialogue, carried by the response that requested it; this event
		// records the dispatch so Chapter 3 can time it and Chapter 4 can
		// cancel it. No dialogue content.

	case ToolReturned:
		if e.Tool == nil {
			return fmt.Errorf("seq %d: tool_returned with no tool payload", e.Seq)
		}
		c.Dialogue = append(c.Dialogue, Entry{Seq: e.Seq, Actor: ActorTool, Kind: KindDialogue, Parts: PartList{
			ToolResultPart{CallID: e.Tool.CallID, Parts: e.Tool.Parts, IsError: e.Tool.IsError},
		}})
		if c.Turn == ToolsPending && c.outstandingCalls() == 0 {
			c.Turn = InputPending
		}
		c.flushHeld()

	case Redacted:
		if e.Redact == nil {
			return fmt.Errorf("seq %d: redacted with no redact payload", e.Seq)
		}
		return c.applyRedaction(e.Seq, *e.Redact)

	case SkillLoaded:
		if e.Skill == nil {
			return fmt.Errorf("seq %d: skill_loaded with no skill payload", e.Seq)
		}
		// Rule 4: the body lives in an entry of its own, so tool clearing
		// cannot delete the manual while keeping the tools it explains.
		// The load_skill tool result is only an acknowledgement.
		if e.Skill.Body == "" {
			return nil // a skill with nothing to say lands no entry
		}
		c.survivor(Entry{Seq: e.Seq, Actor: ActorSystem, Kind: KindSkill, Parts: PartList{
			TextPart{Text: SkillEntryText(e.Skill.Name, e.Skill.Body)},
		}})

	case ToolsChanged:
		if e.Tools == nil {
			return fmt.Errorf("seq %d: tools_changed with no tools payload", e.Seq)
		}
		c.survivor(Entry{Seq: e.Seq, Actor: ActorSystem, Kind: KindTools, Parts: PartList{
			ToolDeclPart{Added: e.Tools.Added, Removed: e.Tools.Removed},
		}})

	case MicroHandoff:
		if e.Handoff == nil {
			return fmt.Errorf("seq %d: micro_handoff with no handoff payload", e.Seq)
		}
		// Rule 5. Applied when the batch that carried the call completes
		// (see Held), so every call it clears has its result.
		c.survivor(Entry{Seq: e.Seq, Actor: ActorSystem, Kind: KindHandoff, Parts: PartList{
			TextPart{Text: HandoffEntryText(e.Handoff.Text)},
		}})

	case ErrorOccurred:
		// Infrastructure failure ends the turn. A tool that ran and failed is
		// ordinary content and does not come through here — conflating the two
		// is why agents get stuck retrying a compile error as a network outage.
		if c.Turn == InFlight || c.Turn == ToolsPending {
			c.Turn = Idle
		}

	default:
		// EVERY PAIR NOT LISTED IS IDENTITY. This arm, not the length of the
		// table above, is what makes the reducer total. A table enumerates the
		// transitions we thought of; the default covers the ones we did not.
		//
		// This is identity, not tolerance, and the difference matters. What
		// lands here is a KNOWN event in a state that simply does not
		// transition on it. An event type we do not recognize never reaches
		// this switch at all: the loader refuses the log, loudly (§2.7).
		// Silently skipping one would produce a context that is wrong in a way
		// nothing downstream can detect — the opposite of what this arm does.
	}
	return nil
}

func hasToolCall(parts PartList) bool {
	for _, p := range parts {
		if _, ok := p.(ToolCallPart); ok {
			return true
		}
	}
	return false
}

// outstandingCalls counts tool calls in the dialogue that have no matching
// result. Computed from the dialogue rather than stored, because a counter in
// the context is a field that can disagree with the log.
func (c *Context) outstandingCalls() int {
	answered := map[string]bool{}
	calls := map[string]bool{}
	for _, entry := range c.Dialogue {
		for _, p := range entry.Parts {
			switch v := p.(type) {
			case ToolCallPart:
				calls[v.CallID] = true
			case ToolResultPart:
				answered[v.CallID] = true
			}
		}
	}
	n := 0
	for id := range calls {
		if !answered[id] {
			n++
		}
	}
	return n
}

// applyRedaction replaces superseded content in place. A redaction is not
// metadata about content — it IS content, and the context holds the result of
// replaying the log.
//
// The span says WHERE, the level says WHAT. Only Dialogue entries are in any
// span's reach: survivors (skills, handoffs, tool declarations) are removed
// by their own verbs, never by tool clearing (Chapter 15 rule 3).
//
// A span that names no dialogue entry cannot be applied, and says so. The
// reducer's callers skip such an event with a diagnostic (rule 8); saying
// nothing would make a stale or hand-edited redaction indistinguishable from
// one that worked.
func (c *Context) applyRedaction(seq Seq, r RedactData) error {
	if r.From > r.To {
		return fmt.Errorf("seq %d: redaction span [%d, %d] is empty", seq, r.From, r.To)
	}
	matched := 0
	for i := range c.Dialogue {
		if c.inSpan(i, r) {
			matched++
		}
	}
	if matched == 0 {
		return fmt.Errorf("seq %d: redaction [%d, %d] names no dialogue entry", seq, r.From, r.To)
	}

	// RedactSummary is the one level that is NOT a per-entry filter, and it
	// has to be lifted out of the loop before the loop is written.
	//
	// The other three transform each entry in the span independently: N
	// entries in, N entries out. A summary REPLACES THE SPAN — §2.4a says the
	// span is replaced by compressed prose, singular — so it folds the span
	// into one entry. Written as a fourth case inside the loop, where it fits
	// so tidily, it copies the same summary into every entry in the span, and
	// compaction then GROWS the context it was called to shrink.
	if r.Level == RedactSummary {
		c.summarizeSpan(r)
		return nil
	}

	emptied := false
	for i := range c.Dialogue {
		if !c.inSpan(i, r) {
			continue
		}
		switch r.Level {
		case RedactResult:
			for j, p := range c.Dialogue[i].Parts {
				res, ok := p.(ToolResultPart)
				if !ok || IsStubbed(res) {
					// Stubbing a stub would replace "[redacted: 9000
					// bytes]" with "[redacted: 0 bytes]": the second cut
					// would erase the record of the first.
					continue
				}
				stub, ref := stubFor(res, r)
				c.Dialogue[i].Parts[j] = ToolResultPart{
					CallID: res.CallID,
					Parts:  PartList{RedactedPart{Stub: stub, Ref: ref}},
				}
			}
		case RedactTool:
			// Calls and results both go; visible reasoning survives.
			kept := make(PartList, 0, len(c.Dialogue[i].Parts))
			for _, p := range c.Dialogue[i].Parts {
				switch p.(type) {
				case ToolCallPart, ToolResultPart:
				default:
					kept = append(kept, p)
				}
			}
			c.Dialogue[i].Parts = kept
			emptied = emptied || len(kept) == 0
		case RedactDialogue:
			// Prose and reasoning go. Survivors are defined by the compaction
			// policy, which is a later chapter's problem.
			kept := make(PartList, 0, len(c.Dialogue[i].Parts))
			for _, p := range c.Dialogue[i].Parts {
				if _, isText := p.(TextPart); !isText {
					kept = append(kept, p)
				}
			}
			c.Dialogue[i].Parts = kept
		}
	}
	if emptied {
		c.dropEmpty()
	}
	return nil
}

// inSpan reports whether dialogue entry i is inside r's span and within
// reach of tool clearing at all.
func (c *Context) inSpan(i int, r RedactData) bool {
	e := c.Dialogue[i]
	return e.Kind == KindDialogue && e.Seq >= r.From && e.Seq <= r.To
}

// IsStubbed reports whether a tool result has already been replaced by a
// stub.
func IsStubbed(res ToolResultPart) bool {
	if len(res.Parts) != 1 {
		return false
	}
	_, ok := res.Parts[0].(RedactedPart)
	return ok
}

// dropEmpty removes dialogue entries that tool clearing left with nothing in
// them. An empty message is not a smaller message; most vendors refuse it.
func (c *Context) dropEmpty() {
	out := c.Dialogue[:0]
	for _, e := range c.Dialogue {
		if e.Kind == KindDialogue && len(e.Parts) == 0 {
			continue
		}
		out = append(out, e)
	}
	c.Dialogue = out
}

// survivor lands a non-dialogue entry now, or holds it until the current
// batch of tool calls completes.
func (c *Context) survivor(e Entry) {
	if c.outstandingCalls() > 0 {
		c.Held = append(c.Held, e)
		return
	}
	c.land(e)
}

// flushHeld lands every held survivor once no call is outstanding.
func (c *Context) flushHeld() {
	if len(c.Held) == 0 || c.outstandingCalls() > 0 {
		return
	}
	held := c.Held
	c.Held = nil
	for _, e := range held {
		c.land(e)
	}
}

// land appends a survivor. A handoff first clears every tool call and tool
// result from the dialogue: it is the checkpoint, and the note it carries
// replaces the traffic. Nothing is outstanding when this runs, so every call
// removed takes its result with it, and no result loses its call.
func (c *Context) land(e Entry) {
	if e.Kind == KindHandoff {
		for i := range c.Dialogue {
			if c.Dialogue[i].Kind != KindDialogue {
				continue
			}
			kept := make(PartList, 0, len(c.Dialogue[i].Parts))
			for _, p := range c.Dialogue[i].Parts {
				switch p.(type) {
				case ToolCallPart, ToolResultPart:
				default:
					kept = append(kept, p)
				}
			}
			c.Dialogue[i].Parts = kept
		}
		c.dropEmpty()
	}
	c.Dialogue = append(c.Dialogue, e)
}

// SkillEntryText is how a loaded skill reads in the context. One place, so
// the reducer and anything that looks for the body agree on it.
func SkillEntryText(name, body string) string {
	return fmt.Sprintf("[skill loaded: %s]\n\n%s", name, body)
}

// HandoffEntryText is how a checkpoint note reads in the context.
func HandoffEntryText(text string) string {
	return "[checkpoint note from micro_handoff]\n\n" + text
}

// summarizeSpan collapses every dialogue entry in [From, To] into a single
// entry. Survivors inside the span stay where they are: a summary of the
// conversation is not a summary of the skill manual.
//
// Two things the event does not say, decided here once so that replay is
// deterministic and so that two later chapters do not answer them differently:
//
//   - Seq is the span's own From. The event already carries it, so the
//     collapsed entry keeps its position without storing anything new.
//   - Actor is System. A span can cross Human, Agent and Tool, and a summary
//     of several speakers is none of their speech — it is compaction output.
//     §2.5 already models the actor, so this costs nothing.
//
// The summary lands at the position of the first entry it supersedes, which
// keeps the dialogue in ascending Seq order without a re-sort.
func (c *Context) summarizeSpan(r RedactData) {
	out := make([]Entry, 0, len(c.Dialogue))
	placed := false
	for i, e := range c.Dialogue {
		if !c.inSpan(i, r) {
			out = append(out, e)
			continue
		}
		if placed || len(r.Replacement) == 0 {
			continue // the span collapses; only the first survivor is emitted
		}
		out = append(out, Entry{
			Seq:   r.From,
			Actor: ActorSystem,
			Kind:  KindDialogue,
			// The one level whose replacement is STORED, because only here is
			// the new content something an LLM wrote and nobody can recompute.
			Parts: append(PartList(nil), r.Replacement...),
		})
		placed = true
	}
	c.Dialogue = out
}

// stubFor synthesizes the replacement text. It is computed from the content it
// supersedes — never stored — which is what keeps replay byte-stable and the
// context free of storage that grows. Informative on purpose: what it was, how
// big it was, and how to get it back.
// stubFor synthesizes the replacement text for a superseded tool result, and
// returns the Ref that survives it.
//
// The Ref is CARRIED FORWARD rather than recorded somewhere new. That is what
// makes a redaction recoverable by construction: the stub says how much went,
// and the Ref still says where it is. Nothing stores "a redaction happened" —
// the log already does, permanently.
func stubFor(res ToolResultPart, r RedactData) (string, Ref) {
	n := 0
	var ref Ref
	for _, p := range res.Parts {
		switch v := p.(type) {
		case TextPart:
			n += len(v.Text)
		case BlobPart:
			ref = v.Ref
		}
	}
	if !ref.Zero() {
		// The locator is NOT repeated in the stub text. It travels in the Ref,
		// where a renderer can turn it into the vendor's own remote-reference
		// form instead of a sentence the model has to parse out of prose.
		return fmt.Sprintf("[redacted: %d bytes; full output retained]", n), ref
	}
	return fmt.Sprintf("[redacted: %d bytes; reason: %s]", n, reasonOr(r.Reason)), Ref{}
}

func reasonOr(s string) string {
	if s == "" {
		return "compaction"
	}
	return s
}
