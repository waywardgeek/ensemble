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
	// File is set on band entries only, and it is what they are ordered by.
	// A pointer so that every other entry serializes exactly as it did
	// before bands existed.
	//
	// Band entries carry Seq 0 rather than the Seq of the event that landed
	// them. Seq exists so a redaction span can match an entry, and no span
	// ever reaches a band (see inSpan). Storing the landing event's Seq
	// would put the ORDER THE BANDS WERE RESTORED IN into the context, so
	// restoring the same files in a different order would produce a
	// different context — the one thing a band restore must never do.
	File *MemoryFileID `json:"file,omitempty"`
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

	// The five memory bands, one kind each. Not one parameterized
	// KindBand{Band}, because the removal rule above wants every kind
	// literally checkable where tool clearing decides what it may touch.
	KindSoul    // SOUL.md, from BandPopulated
	KindMemory  // MEMORY.md, from BandPopulated
	Kind64x     // bucket-1, from BandPopulated
	Kind8x      // bucket-0, from BandPopulated
	KindSession // session memories, from BandPopulated

	// Auto-recalled memories, from RecallAttached. Its own kind rather
	// than a flavour of KindDialogue for three reasons that all point the
	// same way: the user's typed words stay distinguishable from
	// machine-retrieved text, the log records who produced each byte, and
	// a checkpoint can delete recall by kind without touching a user turn.
	//
	// Note what it is NOT. It is not ephemera. Ephemera is cleared on
	// every request, including every tool round, which is right for a
	// timestamp and wrong for a memory. Recall is attached once per user
	// message and then left alone, because a prompt cache is a prefix
	// property: bytes added once are paid for once, and bytes deleted from
	// the middle invalidate everything after them.
	KindRecall // auto-recalled memories, from RecallAttached
)

var entryKindNames = map[EntryKind]string{
	KindDialogue: "dialogue",
	KindHandoff:  "handoff",
	KindSkill:    "skill",
	KindTools:    "tools",
	KindSoul:     "soul",
	KindMemory:   "memory",
	Kind64x:      "64x",
	Kind8x:       "8x",
	KindSession:  "session",
	KindRecall:   "recall",
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
