# Chapter 5 — the seam, in Go (DRAFT)

Status: DRAFT for review. This is the source that ships INLINE in the chapter, so
every type here is a promise in print. Rationale lives in
`book/chapter-05-seed.md`; structure in `book/chapter-05-outline.md`.

Reviewed against the REAL Chapter 2 types in `solutions/ch02/` on 2026-09-13, so
this is an amendment to code the reader already has, not a parallel invention.

---

## Two rules this file obeys

1. **The seam package imports nothing outside the standard library.** Checkable:
   `go list -deps` returns stdlib only. A package that imports nothing cannot
   participate in a cycle — the `gui_independence` guard generalized from "do not
   import the GUI" to "do not import anything."
2. **Not under `internal/`.** Go forbids other modules from importing it there.
   A framework nobody can import is not a framework.

Interfaces for behavior. For data, Chapter 2's existing pattern: a SEALED union
(`isPart()`) plus an ordered-field JSON envelope. Note this corrects a too-strong
claim in an earlier outline draft — a sealed union interface serializes fine; an
OPEN behavioral interface is what cannot be replayed.

---

## 1. `Ref` — the content is elsewhere, here is how to get it

Three requirements collapse into one type: multimedia input, redacted-but-
fetchable tool payloads, and supervision that never reads a history file.

```go
// RefKind says how to resolve a Ref. Constants start at iota+1 so the zero
// value is invalid rather than accidentally meaningful.
type RefKind uint8

const (
	RefPath   RefKind = iota + 1 // a file on local disk
	RefURI                       // remote: a vendor File API uri, gs://, https://
	RefHandle                    // framework-managed output: cr/io/<handle>, or memory
)

// Ref locates content that is NOT stored in the log.
//
// There is deliberately no inline-bytes case. BlobPart's contract in Chapter 2
// is "never inline", because a log you cannot grep is a log you cannot debug.
// Base64 on the wire is a RENDERING decision, made when building a request and
// never written back into the log.
type Ref struct {
	Kind    RefKind `json:"kind"`
	Locator string  `json:"locator"`
}
```

`RefHandle` is distinct from `RefPath` on purpose: a handle is resolved by the
framework and need not be a filesystem path at all — the jobs chapter allows
in-memory buffers.

## 2. The Chapter 2 amendment

Two existing types change. Everything else in `part.go` stands.

```go
// WAS: type BlobPart struct{ MIME, Path string }
// A local path cannot express three of Gemini's four input methods (File API
// uri, gs://, external URL), nor Anthropic's file_id.
type BlobPart struct {
	MIME string
	Ref  Ref
}

// WAS: type RedactedPart struct{ Stub string }
// The stub is still SYNTHESIZED by the reducer, never stored. The Ref is
// carried forward from the part it supersedes, so redaction stays recoverable
// BY CONSTRUCTION rather than by convention.
type RedactedPart struct {
	Stub string
	Ref  Ref // zero value when the superseded content had no locator
}
```

`partJSON` gains two ordered fields — `ref_kind` and `ref` — and keeps its
existing discipline: refuse to marshal an unknown part type, refuse to load a log
containing one.

Media capability stays per MODEL, not per vendor, and refusal stays LOUD.
Chapter 2 already models this with one bool; multimedia generalizes it.

Bill's note on the built system: CodeRhapsody keeps a model-features module —
a table of supported models, what media each accepts, and the constants used to
compute cost. That is better than a method that answers questions about a model,
and the reason is worth stating, because it is a recurring shape:

**Capability is DATA about a model, not BEHAVIOUR of code.** A table is
diffable, testable, and updatable without touching logic. A function that
decides is a place for a guess to hide.

```go
// Media is a bitmask of INPUT media. It is a FIELD in the table below, not an
// interface — the bitmask is how one row answers, not how the system decides.
//
// Verified 2026-09-13. Gemini accepts video as input. OpenAI does NOT: its
// video APIs are GENERATION, a conflation that is easy to make and load-bearing
// if you make it. Anthropic accepts neither audio nor video, and flattens an
// animated GIF to its first frame — it has no time-based media at all.
type Media uint8

const (
	MediaImage Media = 1 << iota
	MediaAudio
	MediaVideo
	MediaDocument
)

// ModelFeatures is one row: everything the seam needs to know about a model.
//
// Cost lives here as integer micro-units per million tokens, NEVER as a float
// and never as money in the log. Chapter 2's rule was "record counts, never
// money"; this is its other half. The log stores what happened, the table
// interprets it, and an interpretation you can re-run against a corrected table
// is worth more than a number you cannot re-derive.
type ModelFeatures struct {
	Media           Media
	InputMicros     int64 // per million tokens
	CacheWriteMicros int64
	CacheReadMicros  int64
	OutputMicros     int64
}

// Keyed by MODEL, not by vendor: Chapter 2 already established that provenance
// is per-model and recorded at write time.
var models = map[string]ModelFeatures{ /* ... */ }

// Lookup returns the row, or false. There is deliberately no default row.
func Lookup(model string) (ModelFeatures, bool) { f, ok := models[model]; return f, ok }
```

The table earns its place by what it does with a model it has never heard of:
nothing. An unknown model is absent from the map, so the caller refuses. A
method that inspected a model name and guessed would answer "probably images"
and be wrong quietly. **A missing row is a loud failure; a default row is a
silent one.** There is no honest degraded rendering of a video part either,
which is why a model that cannot accept one must refuse rather than drop it.

One caveat belongs in print beside the table, because a book is the worst
possible place to publish a list of model names: **this table rots.** Model IDs
are retired and renamed on vendor schedules that have nothing to do with
publication dates. So the table ships with a documented way to re-verify it
against the vendor's live models endpoint, and no chapter's correctness is ever
allowed to depend on a particular model ID still existing. The table is an
example of a shape, not a reference you should trust.

## 3. The outbound seam: observers

```go
// AgentID names which agent an observation came from.
//
// Empty means "the one agent", which is the only case this chapter builds. It
// is `omitempty` everywhere, so the single-agent log is byte-identical to one
// written before this field existed. Sequence numbers stay GLOBAL: per-agent
// numbering was considered and rejected, because a reconnecting observer would
// then have to reconcile N transcripts.
type AgentID string

// PartID identifies a part across its deltas and its finalization.
type PartID uint64

// Observation is what an observer receives. Sealed union, same discipline as
// Part.
type Observation interface{ isObservation() }

// PartDelta is content arriving incrementally.
//
// THE RULE THAT KEEPS THIS SEAM STABLE: streaming is not a mode. A
// non-streaming vendor emits exactly one delta and then a final. Adding real
// streaming after the GUI chapter therefore adds NO new observation kinds —
// only a different chunk count. Tool results are simply always length one,
// which is why "observable but never streaming" needs no special case.
type PartDelta struct {
	Agent  AgentID `json:"agent,omitempty"`
	PartID PartID  `json:"part_id"`
	Chunk  string  `json:"chunk"`
}

// PartFinal is the authoritative, complete part.
type PartFinal struct {
	Agent  AgentID `json:"agent,omitempty"`
	Seq    Seq     `json:"seq"`
	PartID PartID  `json:"part_id"`
	Part   Part    `json:"part"`
}

// StateChanged is a transition of the state machine Chapter 2 already defined.
// Chapter 5 does not introduce agent state. It EXPOSES it.
type StateChanged struct {
	Agent AgentID   `json:"agent,omitempty"`
	From  TurnState `json:"from"`
	To    TurnState `json:"to"`
}

func (PartDelta) isObservation()    {}
func (PartFinal) isObservation()    {}
func (StateChanged) isObservation() {}
```

```go
// Observer watches one or more agents. It never calls back into the agent.
//
// Observe MUST NOT BLOCK. An observer that blocks parks the actor's loop and
// re-creates the exact deafness this chapter exists to remove — the GUI would
// become able to freeze the agent by being slow. A slow observer buffers, or
// drops, on its own time.
type Observer interface {
	Observe(Observation)
}
```

**The stream is not the log.** Deltas are transient and never appended; only
`PartFinal` is recorded. That is what preserves Chapter 2's replay promise and
honors "no field grows without bound."

> The log is the state. The stream is the experience.

## 4. The inbound seam: the mailbox

```go
// Inbound is anything the agent can hear. ONE queue carries all of it — that
// is the entire point. A prompt, a hint typed mid-turn, a tool finishing, and
// a child asking a question are the same kind of event to the loop.
type Inbound interface{ isInbound() }

type UserMessage struct{ Parts PartList }

// Hint is a message that arrives while a turn is already running. It is not a
// separate channel: classification happens in the reducer, by turn state, never
// at capture.
type Hint struct{ Parts PartList }

// ToolCompleted is DELIVERED INTO the queue rather than awaited in a select.
// This is the whole fix. The tool already ran on its own goroutine; the agent
// was deaf because the loop that would drain events was parked waiting for it.
type ToolCompleted struct {
	CallID string
	Result ToolResultPart
}

type Interrupt struct{ AfterTool bool }

// ChildEscalation is a sub-agent blocking on a decision. Built in a later
// chapter; the kind exists now so the queue's shape does not change when it
// arrives.
type ChildEscalation struct {
	Agent     AgentID
	MessageID string
	Parts     PartList
}

func (UserMessage) isInbound()     {}
func (Hint) isInbound()            {}
func (ToolCompleted) isInbound()   {}
func (Interrupt) isInbound()       {}
func (ChildEscalation) isInbound() {}
```

## 5. The actor

A loop with a mailbox. That is the whole definition.

Explicitly NOT built, because "actor" is a loaded word: supervision trees,
addresses, distribution, restart strategies.

```go
// Agent is the loop between the two seams.
type Agent struct{ /* unexported */ }

// Send delivers into the mailbox. It NEVER blocks. Everything the agent can
// hear arrives this way, including its own tools finishing.
func (a *Agent) Send(in Inbound)

// Attach registers an observer. Detach is by the returned func, so an observer
// cannot leak by forgetting its own identity.
func (a *Agent) Attach(o Observer) (detach func())

// State reports the current state without waiting. Cheap, and the answer is a
// FACT held by the framework — nothing parses a history file to discover it.
func (a *Agent) State() TurnState

// Wait blocks until an observation satisfies pred, or ctx is done.
//
// This is the primitive the entire waiting surface derives from.
func (a *Agent) Wait(ctx context.Context, pred func(Observation) bool) (Observation, error)
```

### Everything else is derived, not enumerated

| operation             | expressed as                                        |
|-----------------------|-----------------------------------------------------|
| blocking send         | `Send` then `Wait(turn ended)`                      |
| wait for agent        | `Wait(any state change)`                            |
| join agents           | `Wait(terminal state)` on each                      |
| wake-any, N agents    | `Wait` over a merged observation stream             |
| check progress        | `State()` — no waiting at all                       |

Join and wake-any differ only in predicate, not mechanism. When a surface derives
instead of enumerating, the primitive is usually the right one.

**Wake-any is why the mailbox is structural.** "Wake when ANY awaited agent sends
a message or completes" is epoll/select semantics and cannot be built on blocking
calls. And a parent waiting on N children is the same mechanism as an agent
draining its own queue — one primitive, two scales.

### End of turn is not a result

```go
// Submitted reports what an agent PRODUCED, which is a different fact from
// having stopped.
//
// Keep these separate. Conflating "the turn ended" with "the agent succeeded"
// is how a framework acquires a self-reported success flag, and a self-reported
// success flag invites an agent to tick its own box. Terminal state says it
// stopped. This says what came out. NEITHER says the work was correct.
type Submitted struct {
	Agent AgentID
	Data  json.RawMessage
}
```

This also retires the `DONE:`-marker heuristic: a marker in text is a guess, a
state transition is a fact.

## 6. Supervision needs nothing new

A parent watching a child is `Attach`. Steering it mid-turn is `Send(Hint{...})`.
That is the entire supervision API, and it is why the seam is worth cutting here
rather than after the GUI.

The requirement that forces it: **the parent must never read the child's history
file.** It observes current thinking, current chat, and redacted tool calls, and
resolves a `Ref` on demand for the full payload.

The failure this prevents, measured: a supervision call that returned 307,984
bytes in one call — about 13% of a context window — because supervision was built
as "read the history file." Its sibling, which has volume controls, returned
3,447 bytes for the same job.

> A new waiting primitive must inherit the volume contract, or supervision
> destroys the context window it was meant to protect.

## 7. The workflow seam — learned from the built one, not invented

Bill: "In CodeRhapsody, we put wait and other agent orchestration APIs on a
workflow seam… We definitely need to learn from that, rather than just invent
one." And: "We provide workflow APIs via reverse-MCP: workflows can call tools in
the agent running in CodeRhapsody." He notes the design came from what he learned
from dynamic workflows in Claude Cowork, and is better than his original.

Verified in `~/projects/coderhapsody`: `cr/docs/workflow-design.md`,
`internal/agent/workflow.go`, `internal/agent/subagent_api.go`,
`internal/agent/join.go`, `skills/lib/python/cr_workflow.py`.

### What it actually is

A workflow is a SKILL whose MCP server IS the orchestrator. Control is inverted:
a deterministic script holds the plan, and the model is a leaf-node worker
spawned into a fresh context per micro-task. No new runtime and no new protocol —
it reuses bidirectional MCP and reverse tool dispatch.

Its stated motivation is three failure modes observed in practice, which is
better material than any argument from elegance:

- **agentic laziness** — context fills, and the agent rationalizes stopping at
  70% done
- **self-preferential bias** — the same context that generated the work also
  verifies it
- **goal drift** from compaction

Note `adversarial_verify(result, rubric)` exists in the Python surface as a
first-class helper. It is the second failure mode answered in the API.

### THE CORRECTION TO §5 OF THIS DRAFT

The orchestrator runs in ANOTHER PROCESS. Therefore:

> Orchestration is a TOOL SURFACE, not a Go API.

Which breaks the `Wait` signature drafted above:

```go
// WRONG as the seam. A Go closure cannot cross a process boundary, so a Python
// workflow script can never call this.
func (a *Agent) Wait(ctx context.Context, pred func(Observation) bool) (Observation, error)
```

The real surface is a closed set of SERIALIZABLE wait conditions. CodeRhapsody's
is `wait_for_agent_change(agent_ids, patterns, timeout_seconds,
settling_seconds, stuck_threshold_seconds)` returning typed events —
`status_change`, `agent_exited`, `message_to_parent`, `pattern_match`, `stuck`.

So the honest design is two layers, and the chapter should say which is the seam:

| layer | form | who calls it |
|-------|------|--------------|
| in-process convenience | `Wait(ctx, predicate)` with a Go closure | Go embedders |
| **the seam** | a serializable wait condition, invoked as a tool | the model, AND a deterministic script in another process |

The derivations in §5 still hold — blocking send, wait-for-agent, join and
wake-any all come from one waiting primitive, differing only in condition. What
changes is that the condition must be DATA, not a function.

Operational details in the built version that look like scar tissue and are
probably load-bearing: `settling_seconds` debounces a rapid flap, and
`stuck_threshold_seconds` synthesizes a "stuck" event only at timeout. Neither
would occur to someone designing this from taste.

### The rule worth stealing: one escape hatch, ergonomics above the seam

`cr_workflow.py` is roughly 30 methods — `spawn`, `send_message`, `join`,
`join_strict`, `fan_out`, `agent_json(schema, retries)`, `adversarial_verify` —
and all of them are sugar over ONE primitive:

> `call(tool_name, arguments)` — "Reverse-call ANY tool enabled on the host
> agent, by name… The typed helpers above are conveniences, not the ceiling.
> Anything the host agent has enabled is reachable here — this is the floor the
> wrappers are built on."

**Keep the seam generic; put ergonomics above it, not inside it.** Thirty
convenience methods on the seam would be thirty promises to keep. Thirty
convenience methods above a generic `call` are a library, replaceable without
touching the framework.

This is the same shape as the measured facade in §5.7 of the outline — small
seam, and what is deliberately absent is part of the design.

### Consequence for chapter order

The workflow chapter is not merely "later, when we get to it." It INVERTS
control, and it needs the tool surface to already be the orchestration surface.
Chapter 5 must therefore make the waiting condition serializable even though
Chapter 5 itself only ever calls it in-process. That is the same class of
decision as the `omitempty` agent tag: cheap now, a broken promise later.

## Open

1. PARTLY ANSWERED by §7: the SEAM's wait condition must be serializable data,
   because the orchestrator may be a script in another process. A Go closure
   `Wait` can still exist as an in-process convenience. Still open: whether the
   merged multi-agent wait is a free function over agents or a `Watcher` type,
   and what the closed set of serializable conditions should be for this book —
   CodeRhapsody's five event kinds are a starting point, not a ruling.
2. `PartDelta.Chunk` as `string` vs `[]byte`. String is friendlier in print and
   correct for chat/thinking/args; media never streams through this path.
3. Whether `Submitted` is an `Observation`, an `Inbound` on the parent, or both.
4. Exact `partJSON` field names for the Ref.
