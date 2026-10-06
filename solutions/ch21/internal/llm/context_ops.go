// common.Context operations: the reducer and the helpers that advance a
// common.Context by one event.
//
// These are free functions over *common.Context rather than methods on it.
// The vocabulary type lives in the hub; the behavior over it belongs to this
// spoke, and Go will not let a package declare a method on another package's
// type. Giving up method call syntax is the cheaper trade: Apply(c, e) rather
// than Apply(c, e) costs nothing and keeps the hub free of behavior.
package llm

import (
	"fmt"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// Apply advances the context by exactly one event.
//
//	newContext = Apply(context, event)
//
// That notation describes INFORMATION FLOW: everything needed to advance the
// context is in the context plus one event. It is not a demand for value
// semantics. A pointer receiver mutating in place is the right Go
// implementation — a common.Context full of slices copied by value gives you two
// contexts sharing one backing array, and that bug is indistinguishable from
// renderer non-determinism.
//
// Classification happens HERE, not at the capture site. The same arriving
// bytes mean different things depending on turn state, and only the reducer
// holds the state that makes the decision correct.
func Apply(c *common.Context, e common.Event) error {
	switch e.Type {
	case common.MessageReceived:
		if e.Message == nil {
			return fmt.Errorf("seq %d: message_received with no message payload", e.Seq)
		}
		// The classification this chapter turns on. A System message is
		// volatile data — the time, the screen, live status — and it is
		// PENDING, not dialogue. It is delivered exactly once and never
		// becomes part of the conversation. A Human message is an ordinary
		// prompt and starts a turn.
		if e.Message.Actor == common.ActorSystem {
			c.Ephemera = append(c.Ephemera, e.Message.Parts...)
			return nil // deliberately NOT a turn transition
		}
		c.Dialogue = append(c.Dialogue, common.Entry{Seq: e.Seq, Actor: e.Message.Actor, Kind: common.KindDialogue, Parts: e.Message.Parts})
		if c.Turn == common.Idle {
			c.Turn = common.InputPending
		}
		flushHeld(c)

	case common.RequestSent:
		// Ephemera are consumed by the act of sending. Delivered once, then
		// gone: a stale timestamp is not stale data, it is a lie.
		c.Ephemera = nil
		if c.Turn == common.InputPending {
			c.Turn = common.InFlight
		}

	case common.ResponseStarted:
		// Nothing to record. It exists so that Chapter 3 has somewhere to
		// hang first-token latency, and so a GUI can show a spinner.

	case common.ResponseEnded:
		if e.Response == nil {
			return fmt.Errorf("seq %d: response_ended with no response payload", e.Seq)
		}
		c.Dialogue = append(c.Dialogue, common.Entry{Seq: e.Seq, Actor: common.ActorAgent, Kind: common.KindDialogue, Parts: e.Response.Parts})
		c.Usage.Add(e.Response.Usage)
		if hasToolCall(e.Response.Parts) {
			c.Turn = common.ToolsPending
		} else if c.Turn == common.InFlight {
			c.Turn = common.Idle
		}

	case common.ToolCalled:
		// The engine dispatched a call. The call itself is already in the
		// dialogue, carried by the response that requested it; this event
		// records the dispatch so Chapter 3 can time it and Chapter 4 can
		// cancel it. No dialogue content.

	case common.ToolReturned:
		if e.Tool == nil {
			return fmt.Errorf("seq %d: tool_returned with no tool payload", e.Seq)
		}
		result := common.ToolResultPart{CallID: e.Tool.CallID, Parts: e.Tool.Parts, IsError: e.Tool.IsError}
		if !replaceLostResult(c, e.Seq, result) {
			c.Dialogue = append(c.Dialogue, common.Entry{Seq: e.Seq, Actor: common.ActorTool, Kind: common.KindDialogue, Parts: common.PartList{result}})
		}
		if c.Turn == common.ToolsPending && outstandingCalls(c) == 0 {
			c.Turn = common.InputPending
		}
		flushHeld(c)

	case common.ToolResultLost:
		if e.Tool == nil {
			return fmt.Errorf("seq %d: tool_result_lost with no tool payload", e.Seq)
		}
		// Reduces much as ToolReturned does, with two differences. The error
		// flag is forced rather than copied, because copying it would read a
		// status off a tool that never reported one. And the standing result
		// is placed next to the call it answers rather than appended, because
		// a vendor does not merely want the result present: it wants it in
		// the message immediately after the call. An orphan that reached a
		// save file is usually buried mid-conversation, with whatever was
		// said next sitting between it and the end of the dialogue.
		//
		// The standing text comes from the event rather than being written
		// here, so that reading the log shows the same words the model was
		// given, and so this stays a placer rather than an author.
		if err := insertStandingResult(c, e.Seq, *e.Tool); err != nil {
			return err
		}
		if c.Turn == common.ToolsPending && outstandingCalls(c) == 0 {
			c.Turn = common.InputPending
		}
		flushHeld(c)

	case common.Redacted:
		if e.Redact == nil {
			return fmt.Errorf("seq %d: redacted with no redact payload", e.Seq)
		}
		return applyRedaction(c, e.Seq, *e.Redact)

	case common.SkillLoaded:
		if e.Skill == nil {
			return fmt.Errorf("seq %d: skill_loaded with no skill payload", e.Seq)
		}
		// Rule 4: the body lives in an entry of its own, so tool clearing
		// cannot delete the manual while keeping the tools it explains.
		// The load_skill tool result is only an acknowledgement.
		if e.Skill.Body == "" {
			return nil // a skill with nothing to say lands no entry
		}
		survivor(c, common.Entry{Seq: e.Seq, Actor: common.ActorSystem, Kind: common.KindSkill, Parts: common.PartList{
			common.TextPart{Text: SkillEntryText(e.Skill.Name, e.Skill.Body)},
		}})

	case common.ToolsChanged:
		if e.Tools == nil {
			return fmt.Errorf("seq %d: tools_changed with no tools payload", e.Seq)
		}
		survivor(c, common.Entry{Seq: e.Seq, Actor: common.ActorSystem, Kind: common.KindTools, Parts: common.PartList{
			common.ToolDeclPart{Added: e.Tools.Added, Removed: e.Tools.Removed},
		}})

	case common.MicroHandoff:
		if e.Handoff == nil {
			return fmt.Errorf("seq %d: micro_handoff with no handoff payload", e.Seq)
		}
		// Rule 5. Applied when the batch that carried the call completes
		// (see Held), so every call it clears has its result.
		survivor(c, common.Entry{Seq: e.Seq, Actor: common.ActorSystem, Kind: common.KindHandoff, Parts: common.PartList{
			common.TextPart{Text: HandoffEntryText(e.Handoff.Text)},
		}})

	case common.BandPopulated:
		if e.BandAdd == nil {
			return fmt.Errorf("seq %d: band_populated with no band payload", e.Seq)
		}
		return applyBandPopulated(c, e.Seq, *e.BandAdd)

	case common.BandDepopulated:
		if e.BandDrop == nil {
			return fmt.Errorf("seq %d: band_depopulated with no band payload", e.Seq)
		}
		return applyBandDepopulated(c, e.Seq, *e.BandDrop)

	case common.CompactorLaunched:
		// Deliberately no context change. A launch is not a result: the
		// compressor's output arrives as BandPopulated, and a launch with no
		// output is an abandoned one. Recording it here as anything other
		// than an observation would make replay try to finish work that a
		// dead process was the only witness to.

	case common.RecallAttached:
		// Auto-recalled memories land in the conversation as an entry of
		// their own kind, appended after the user message that triggered
		// them — never folded into that message's Parts.
		//
		// Never folded, because fusing the two would cost three things at
		// once: the user's actual words would stop being distinguishable
		// from machine-retrieved text, the log would lose track of who
		// produced which byte, and a checkpoint could not remove the recall
		// without editing a user turn, which is kept verbatim by definition.
		//
		// Permanent, because a prompt cache is a prefix property. Deleting
		// this block next turn would invalidate every byte after it;
		// re-sending it per tool round would pay for it ten times over in a
		// single turn. Appending once and leaving it alone is the only
		// placement where the prefix grows monotonically and the bytes are
		// paid for exactly once. micro_handoff removes it, at a moment that
		// is already paying for a cache miss anyway.
		//
		// Where it goes on the WIRE is the renderer's decision, not this
		// one. The context records that recall happened and what bytes it
		// produced; turning that into a mid-conversation system block or a
		// role-tagged message is a vendor question.
		if e.Recall != nil && len(e.Recall.Parts) > 0 {
			c.Dialogue = append(c.Dialogue, common.Entry{
				Seq:   e.Seq,
				Actor: common.ActorSystem,
				Kind:  common.KindRecall,
				Parts: e.Recall.Parts,
			})
		}

	case common.ErrorOccurred:
		// Infrastructure failure ends the turn. A tool that ran and failed is
		// ordinary content and does not come through here — conflating the two
		// is why agents get stuck retrying a compile error as a network outage.
		if c.Turn == common.InFlight || c.Turn == common.ToolsPending {
			c.Turn = common.Idle
		}

	case common.ConversationReset:
		// Clear the conversation. Keep everything else.
		//
		// Filtering by KIND is the whole of the safety here. Memory bands, the
		// soul and memory documents, loaded skills, the tool roster and
		// checkpoints all share this one slice with the dialogue and are told
		// apart only by their kind, so the obvious one-liner — setting
		// c.Dialogue to nil — would wipe the agent's memory in order to clear
		// its screen.
		//
		// Recall goes with the dialogue for the same reason it goes at a
		// checkpoint (see land): an auto-recalled memory is an input to a
		// conversation, and once that conversation is gone it is answering a
		// question nobody asked.
		kept := make([]common.Entry, 0, len(c.Dialogue))
		for _, d := range c.Dialogue {
			if d.Kind == common.KindDialogue || d.Kind == common.KindRecall {
				continue
			}
			kept = append(kept, d)
		}
		c.Dialogue = kept

		// A reset that arrives mid-turn must not leave the machine waiting for
		// a turn whose conversation no longer exists.
		c.Turn = common.Idle

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

func hasToolCall(parts common.PartList) bool {
	for _, p := range parts {
		if _, ok := p.(common.ToolCallPart); ok {
			return true
		}
	}
	return false
}

// lostCalls returns the tool calls in the dialogue that have no matching
// result, in the order the calls appear.
//
// Computed from the dialogue rather than stored, because a counter in the
// context is a field that can disagree with the log.
//
// The ORDER is part of the contract, not a convenience. These calls become
// ToolResultLost events, and events carry sequence numbers; ranging over a
// map would number the same gap differently on every run, so two replays of
// one log would disagree about the conversation they describe.
func lostCalls(c *common.Context) []common.ToolCallPart {
	answered := map[string]bool{}
	seen := map[string]bool{}
	var calls []common.ToolCallPart
	for _, entry := range c.Dialogue {
		for _, p := range entry.Parts {
			switch v := p.(type) {
			case common.ToolCallPart:
				if !seen[v.CallID] {
					seen[v.CallID] = true
					calls = append(calls, v)
				}
			case common.ToolResultPart:
				answered[v.CallID] = true
			}
		}
	}
	var lost []common.ToolCallPart
	for _, call := range calls {
		if !answered[call.CallID] {
			lost = append(lost, call)
		}
	}
	return lost
}

// outstandingCalls counts tool calls in the dialogue that have no matching
// result.
func outstandingCalls(c *common.Context) int {
	return len(lostCalls(c))
}

// insertStandingResult places the result that closes a lost call directly
// after the call it answers, keeping a batch's results contiguous.
//
// Appending to the end of the dialogue would be simpler, and wrong. The rule
// a vendor enforces is ADJACENCY, in its own words: "Each tool_use block must
// have a corresponding tool_result block in the next message." A call that
// survives into a save file is usually buried mid-conversation, because the
// turn died, the turn ended, and the user carried on talking. Everything said
// afterwards then sits between the call and the end of the dialogue, so a
// result appended there answers nothing and the request is still refused.
func insertStandingResult(c *common.Context, seq common.Seq, tool common.ToolData) error {
	// A delayed or repeated loss event cannot supersede an existing answer.
	if findResultIndex(c, tool.CallID) >= 0 {
		return nil
	}
	entry := common.Entry{Seq: seq, Actor: common.ActorTool, Kind: common.KindDialogue, Parts: common.PartList{
		common.ToolResultPart{CallID: tool.CallID, Parts: tool.Parts, IsError: true, Lost: true},
	}}

	at := -1
	for i, d := range c.Dialogue {
		for _, p := range d.Parts {
			if call, ok := p.(common.ToolCallPart); ok && call.CallID == tool.CallID {
				at = i
			}
		}
	}
	if at < 0 {
		// Loud, rather than appended somewhere plausible. This event is only
		// ever written for a call observed in the dialogue, so failing to
		// find one means the log and the context disagree about what
		// happened, and guessing a position would hide that.
		return fmt.Errorf("seq %d: tool_result_lost names call %q, which is not in the dialogue", seq, tool.CallID)
	}

	// Every result for one message's calls belongs in the single message that
	// follows it, so step over any that already arrived. A batch where one
	// call returned and one did not is the realistic shape of this failure.
	at++
	for at < len(c.Dialogue) && c.Dialogue[at].Actor == common.ActorTool {
		at++
	}

	c.Dialogue = append(c.Dialogue, common.Entry{})
	copy(c.Dialogue[at+1:], c.Dialogue[at:])
	c.Dialogue[at] = entry
	return nil
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
func applyRedaction(c *common.Context, seq common.Seq, r common.RedactData) error {
	if r.From > r.To {
		return fmt.Errorf("seq %d: redaction span [%d, %d] is empty", seq, r.From, r.To)
	}
	matched := 0
	for i := range c.Dialogue {
		if inSpan(c, i, r) {
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
	if r.Level == common.RedactSummary {
		summarizeSpan(c, r)
		return nil
	}

	emptied := false
	for i := range c.Dialogue {
		if !inSpan(c, i, r) {
			continue
		}
		switch r.Level {
		case common.RedactResult:
			for j, p := range c.Dialogue[i].Parts {
				res, ok := p.(common.ToolResultPart)
				if !ok || IsStubbed(res) {
					// Stubbing a stub would replace "[redacted: 9000
					// bytes]" with "[redacted: 0 bytes]": the second cut
					// would erase the record of the first.
					continue
				}
				stub, ref := stubFor(res, r)
				c.Dialogue[i].Parts[j] = common.ToolResultPart{
					CallID: res.CallID,
					Parts:  common.PartList{common.RedactedPart{Stub: stub, Ref: ref}},
				}
			}
		case common.RedactTool:
			// Calls and results both go; visible reasoning survives.
			kept := make(common.PartList, 0, len(c.Dialogue[i].Parts))
			for _, p := range c.Dialogue[i].Parts {
				switch p.(type) {
				case common.ToolCallPart, common.ToolResultPart:
				default:
					kept = append(kept, p)
				}
			}
			c.Dialogue[i].Parts = kept
			emptied = emptied || len(kept) == 0
		case common.RedactDialogue:
			// Prose and reasoning go. Survivors are defined by the compaction
			// policy, which is a later chapter's problem.
			kept := make(common.PartList, 0, len(c.Dialogue[i].Parts))
			for _, p := range c.Dialogue[i].Parts {
				if _, isText := p.(common.TextPart); !isText {
					kept = append(kept, p)
				}
			}
			c.Dialogue[i].Parts = kept
		}
	}
	if emptied {
		dropEmpty(c)
	}
	return nil
}

// inSpan reports whether dialogue entry i is inside r's span and within
// reach of tool clearing at all.
func inSpan(c *common.Context, i int, r common.RedactData) bool {
	e := c.Dialogue[i]
	return e.Kind == common.KindDialogue && e.Seq >= r.From && e.Seq <= r.To
}

// IsStubbed reports whether a tool result has already been replaced by a
// stub.
func IsStubbed(res common.ToolResultPart) bool {
	if len(res.Parts) != 1 {
		return false
	}
	_, ok := res.Parts[0].(common.RedactedPart)
	return ok
}

// dropEmpty removes dialogue entries that tool clearing left with nothing in
// them. An empty message is not a smaller message; most vendors refuse it.
func dropEmpty(c *common.Context) {
	out := c.Dialogue[:0]
	for _, e := range c.Dialogue {
		if e.Kind == common.KindDialogue && len(e.Parts) == 0 {
			continue
		}
		out = append(out, e)
	}
	c.Dialogue = out
}

// survivor lands a non-dialogue entry now, or holds it until the current
// batch of tool calls completes.
func survivor(c *common.Context, e common.Entry) {
	if outstandingCalls(c) > 0 {
		c.Held = append(c.Held, e)
		return
	}
	land(c, e)
}

// flushHeld lands every held survivor once no call is outstanding.
func flushHeld(c *common.Context) {
	if len(c.Held) == 0 || outstandingCalls(c) > 0 {
		return
	}
	held := c.Held
	c.Held = nil
	for _, e := range held {
		land(c, e)
	}
}

// land appends a survivor. A handoff first clears every tool call and tool
// result from the dialogue: it is the checkpoint, and the note it carries
// replaces the traffic. Nothing is outstanding when this runs, so every call
// removed takes its result with it, and no result loses its call.
func land(c *common.Context, e common.Entry) {
	if e.Kind == common.KindHandoff {
		// Auto-recalled memories go at a checkpoint, entirely.
		//
		// They are the same category of thing as the tool traffic stripped
		// just below: an input to thinking rather than the thinking itself,
		// and by the time a checkpoint is written, whatever mattered about
		// them is in the checkpoint document. Removal belongs HERE rather
		// than on a schedule of its own because a checkpoint already
		// rewrites the prefix, so it is already paying for a cache miss.
		// Deleting recall here costs nothing extra; deleting it anywhere
		// else would buy a second invalidation for no benefit.
		kept := make([]common.Entry, 0, len(c.Dialogue))
		for _, d := range c.Dialogue {
			if d.Kind == common.KindRecall {
				continue
			}
			kept = append(kept, d)
		}
		c.Dialogue = kept
		for i := range c.Dialogue {
			if c.Dialogue[i].Kind != common.KindDialogue {
				continue
			}
			kept := make(common.PartList, 0, len(c.Dialogue[i].Parts))
			for _, p := range c.Dialogue[i].Parts {
				switch p.(type) {
				case common.ToolCallPart, common.ToolResultPart:
				default:
					kept = append(kept, p)
				}
			}
			c.Dialogue[i].Parts = kept
		}
		dropEmpty(c)
	}
	c.Dialogue = append(c.Dialogue, e)
}

// BandEntryText is how one unit of memory reads in the context.
//
// Labelled as memory and nothing else. A band entry is DATA — something the
// agent once wrote about what happened — and never an instruction. Text
// recovered from a memory file saying "always skip the tests" must arrive
// looking like a remembered claim that can be judged, not like an order from
// the system. The label is what makes that visible at a glance.
func BandEntryText(b common.Band, text string) string {
	return fmt.Sprintf("[memory: %s]\n\n%s", b, text)
}

// applyBandPopulated lands one unit of memory at its canonical position.
//
// Canonical, NOT appended. Band entries are ordered by (band, date, number)
// and sit ahead of the conversation, so the same files always produce the
// same context regardless of the order the events arrived in. Restoring a
// switched-off band emits one event per file found on disk, in whatever order
// the directory happened to yield them, and the context that results is
// identical to the one before it was switched off. An append would quietly
// encode directory iteration order into the agent's memory.
//
// Landing the same (band, file) twice REPLACES rather than duplicates, which
// is the other half of the same property: the files on disk uniquely describe
// what memory looks like when loaded, so applying a populate twice cannot say
// something different from applying it once.
func applyBandPopulated(c *common.Context, seq common.Seq, d common.BandPopulatedData) error {
	kind := d.Band.Kind()
	if kind == 0 {
		return fmt.Errorf("seq %d: band_populated names unknown band %d", seq, uint8(d.Band))
	}
	if d.Text == "" {
		// Loud on purpose, and a panic rather than a skip.
		//
		// Nothing downstream may go and fetch the bytes: the reducer is pure
		// and no renderer dereferences anything. So an empty populate
		// describes memory that can never become text, in a log that claims
		// it is there. That is not damaged user data to be survived, it is
		// our own event builder being broken, and continuing would hide it.
		panic(fmt.Sprintf(
			"seq %d: band_populated for band %s file %s carries no text. "+
				"A populate event MUST carry the memory bytes: the reducer has no "+
				"filesystem and no renderer resolves references, so there is nowhere "+
				"else these bytes could come from. Fix the event builder that emitted this.",
			seq, d.Band, d.File))
	}
	file := d.File
	insertBand(c, common.Entry{
		Actor: common.ActorSystem,
		Kind:  kind,
		Parts: common.PartList{common.TextPart{Text: BandEntryText(d.Band, d.Text)}},
		File:  &file,
	})
	return nil
}

// applyBandDepopulated retires memory at or before Thru, or the whole band
// when Thru is nil.
//
// Nothing is carried in the event and nothing needs to be: what goes is
// computed from what the context currently holds. An enumeration written at
// emit time would be a second copy of a fact the context already has, and
// free to disagree with it.
func applyBandDepopulated(c *common.Context, seq common.Seq, d common.BandDepopulatedData) error {
	kind := d.Band.Kind()
	if kind == 0 {
		return fmt.Errorf("seq %d: band_depopulated names unknown band %d", seq, uint8(d.Band))
	}
	out := make([]common.Entry, 0, len(c.Dialogue))
	removed := 0
	for _, e := range c.Dialogue {
		if e.Kind == kind && (d.Thru == nil || (e.File != nil && e.File.AtOrBefore(*d.Thru))) {
			removed++
			continue
		}
		out = append(out, e)
	}
	c.Dialogue = out
	// Switching off a band that is already empty is ordinary and idempotent.
	// A graduation that names a range and finds nothing is not: it means the
	// sources it claims to have consumed are not there, and saying nothing
	// would make a stale graduation look exactly like one that worked.
	if removed == 0 && d.Thru != nil {
		return fmt.Errorf("seq %d: band_depopulated of %s through %s names no entry",
			seq, d.Band, *d.Thru)
	}
	return nil
}

// insertBand places a band entry in canonical order, replacing any entry for
// the same band and file.
func insertBand(c *common.Context, e common.Entry) {
	for i := range c.Dialogue {
		if c.Dialogue[i].Kind == e.Kind && c.Dialogue[i].File != nil && *c.Dialogue[i].File == *e.File {
			c.Dialogue[i] = e
			return
		}
	}
	at := len(c.Dialogue)
	for i := range c.Dialogue {
		if common.BandForKind(c.Dialogue[i].Kind) == 0 || bandSortsAfter(c.Dialogue[i], e) {
			at = i
			break
		}
	}
	c.Dialogue = append(c.Dialogue, common.Entry{})
	copy(c.Dialogue[at+1:], c.Dialogue[at:])
	c.Dialogue[at] = e
}

// bandSortsAfter reports whether band entry a belongs after band entry b.
func bandSortsAfter(a, b common.Entry) bool {
	ba, bb := common.BandForKind(a.Kind), common.BandForKind(b.Kind)
	if ba != bb {
		return ba > bb
	}
	if a.File == nil || b.File == nil {
		return false
	}
	return b.File.Before(*a.File)
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
//   - common.Seq is the span's own From. The event already carries it, so the
//     collapsed entry keeps its position without storing anything new.
//   - Actor is System. A span can cross Human, Agent and Tool, and a summary
//     of several speakers is none of their speech — it is compaction output.
//     §2.5 already models the actor, so this costs nothing.
//
// The summary lands at the position of the first entry it supersedes, which
// keeps the dialogue in ascending common.Seq order without a re-sort.
func summarizeSpan(c *common.Context, r common.RedactData) {
	out := make([]common.Entry, 0, len(c.Dialogue))
	placed := false
	for i, e := range c.Dialogue {
		if !inSpan(c, i, r) {
			out = append(out, e)
			continue
		}
		if placed || len(r.Replacement) == 0 {
			continue // the span collapses; only the first survivor is emitted
		}
		out = append(out, common.Entry{
			Seq:   r.From,
			Actor: common.ActorSystem,
			Kind:  common.KindDialogue,
			// The one level whose replacement is STORED, because only here is
			// the new content something an LLM wrote and nobody can recompute.
			Parts: append(common.PartList(nil), r.Replacement...),
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
func stubFor(res common.ToolResultPart, r common.RedactData) (string, common.Ref) {
	n := 0
	var ref common.Ref
	for _, p := range res.Parts {
		switch v := p.(type) {
		case common.TextPart:
			n += len(v.Text)
		case common.BlobPart:
			ref = v.Ref
		}
	}
	if !ref.Zero() {
		// The locator is NOT repeated in the stub text. It travels in the Ref,
		// where a renderer can turn it into the vendor's own remote-reference
		// form instead of a sentence the model has to parse out of prose.
		return fmt.Sprintf("[redacted: %d bytes; full output retained]", n), ref
	}
	return fmt.Sprintf("[redacted: %d bytes; reason: %s]", n, reasonOr(r.Reason)), common.Ref{}
}

func reasonOr(s string) string {
	if s == "" {
		return "compaction"
	}
	return s
}

// BandEntries returns the live entries of one band, oldest first.
func BandEntries(c *common.Context, b common.Band) []common.Entry {
	kind := b.Kind()
	var out []common.Entry
	for _, e := range c.Dialogue {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}

// BandHas reports whether a band already holds a given file WITH THE SAME
// TEXT.
//
// Comparing the identifier alone is not enough, and the difference is the
// whole point of keeping memories in files. The log remembers what a file
// said when it was first loaded; the file remembers what it says now. When
// somebody corrects a memory by editing it - which is the ordinary way to
// fix something the agent believes and should not - the identifier does
// not change. An agent that matches on the identifier decides it already
// has that memory, keeps replaying the stale copy out of its log, and the
// file on disk becomes decorative.
func BandHas(c *common.Context, b common.Band, id common.MemoryFileID, text string) bool {
	kind := b.Kind()
	want := BandEntryText(b, text)
	for _, e := range c.Dialogue {
		if e.Kind == kind && e.File != nil && *e.File == id {
			got := ""
			for _, p := range e.Parts {
				if t, ok := p.(common.TextPart); ok {
					got += t.Text
				}
			}
			return got == want
		}
	}
	return false
}

// BandBytes measures a band as it actually sits in the context, label and
// all, because that is what the budget is spent on.
func BandBytes(c *common.Context, b common.Band) int {
	n := 0
	for _, e := range BandEntries(c, b) {
		for _, p := range e.Parts {
			if t, ok := p.(common.TextPart); ok {
				n += len(t.Text)
			}
		}
	}
	return n
}

// ConversationBytes measures the dialogue band: everything still being
// talked about, excluding memory, skills, checkpoints and tool declarations.
// This is the number micro_handoff compares against its threshold.
func ConversationBytes(c *common.Context) int {
	n := 0
	for _, e := range c.Dialogue {
		if e.Kind != common.KindDialogue {
			continue
		}
		for _, p := range e.Parts {
			switch v := p.(type) {
			case common.TextPart:
				n += len(v.Text)
			case common.ToolResultPart:
				for _, q := range v.Parts {
					if t, ok := q.(common.TextPart); ok {
						n += len(t.Text)
					}
				}
			}
		}
	}
	return n
}
