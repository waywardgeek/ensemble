# Chapter 5 outline — The Agent Framework

Working title. "Creating the agent framework" names the activity, not the idea.
Candidate in the established Title+Subtitle pattern, so the TOC teaches:

> **Two Seams and a Loop: Turning an Agent Into a Framework**

Status: OUTLINE. Rulings are recorded in book/chapter-05-seed.md; this file is
the structure. Everything here derives from a ruling in the seed or is marked
OPEN.

Confirmed by Bill 2026-09-13:
- ch5 is the agent framework; it precedes the GUI chapter.
- The gateway is a SEPARATE chapter, AFTER the GUI chapter.
- Exercise is author/editor/reviewer. The chat/Discord agent moves to the
  gateway chapter.
- The seam ships as Go SOURCE, INLINE in this chapter.
- The seam must live where other projects can import it.

---

## The chapter's promise

By the end, the reader's coding agent is a FRAMEWORK: the same core runs a
completely different agent, defined by different tools, a different prompt and
different observers, without one line changing inside the framework package.

That claim is falsifiable, and the exercise falsifies it.

## Thesis

**A framework is two seams and a loop.**

- Outbound: the **observer** seam. What the agent says.
- Inbound: the **mailbox**. What the agent hears.
- Between them: the **actor** — defined minimally as a loop with a mailbox.

Both seams are the shape the reader already met in Chapter 2: one stream, many
renderings. So this chapter mostly REVEALS structure rather than introducing it.

---

## §5.0 Cold open — the framework that depended on its own GUI

The war story. All of it is verified in the CodeRhapsody tree; see the seed for
commit-level receipts.

Beats:

1. CodeRhapsody built the GUI before it had a framework seam. The framework
   ended up importing the GUI.
2. It was not noticed as an architecture failure. It surfaced as test
   scaffolding: `test_support.go` built a real GUI server for the ~23 tests that
   needed a UI attached. That is the tell — the dependency hid inside
   convenience.
3. The fix was not deletion. **The GUI was DEMOTED to an observer.** Five commits
   across two days, the load-bearing one titled *"Test scaffolding takes a
   UIObserver, not a GUI server."*
4. It is not finished. The guard exempts test files and RATCHETS them downward,
   on the stated principle that "a guard that fails today guards nothing." The
   remaining edges are called "a real remaining blocker for the module split."
5. The non-obvious part, and the reason it is still unfinished: at PACKAGE level
   a test edge is harmless, because a test binary is in no other package's import
   closure. At MODULE level, a framework test importing the GUI forces the
   framework's `go.mod` to require the GUI module, which already requires the
   framework — a cycle. **The seam can look clean while the dependency graph is
   not.**

Second instance, same failure, stated by the compiler rather than by argument:
the entire framework lives under `internal/`. `pkg/` holds one package;
`internal/agent` alone is 112 files. Go forbids any other module from importing
anything under `internal/`. An agent library that no other project can import is
not a library. This is deliberately dropped here and paid off in §5.7.

Transition: both failures are the same mistake — deciding what the core knows
about, too late. This chapter cuts the seam first.

## §5.1 You have already written an observer

Cheap win that orients the reader before any new machinery.

Chapter 2 built a renderer that turned the event log into `history.md`. That is
an observer: it watches the stream and produces a view. The GUI the reader has
not built yet is the same kind of thing. So is a logger, a metrics sink, and — as
§5.6 shows — a parent supervising a child.

Point to land: the reader is not learning a new pattern, they are learning that
they already used it and can now name it.

## §5.2 The outbound seam: observers

**Go source, inline.** Interface for behavior; concrete structs for data.

Design content:

- An observer receives events. It does not call back into the agent.
- Two kinds of thing flow: **content** (thinking, chat, tool call parameters,
  tool results) and **state transitions** (turn started, turn ended, idle,
  blocked awaiting a decision). One stream carries both.
- Registration carries an agent identity (see §5.5 on not foreclosing
  sub-agents).

**The streaming rule**, which is the reason this seam survives the next two
chapters unchanged:

> Do not model streaming as a mode. Model non-streaming as a stream of length
> one.

A part has a stable identity; deltas append to it; a terminal marker finalizes
it. A non-streaming vendor emits one delta plus the finalizer. Observers that do
not care about liveness ignore deltas and act on finalizers; the GUI renders
deltas. Adding streaming later therefore adds NO new event kinds.

Tool results are the proof the rule is right: Bill's requirement is that they are
observable but never stream, which needs no special case at all — they are simply
always length one.

Receipt that streaming tool-call parameters is real and not anticipated here for
fun: Anthropic ships a "fine-grained tool streaming" feature.

**The stream is not the log.** Deltas are transient and are never appended to the
durable log; only finalized parts are recorded. This is what preserves Chapter
2's replay promise and honors "no field grows without bound."

> The log is the state. The stream is the experience.

Gradeable property, and the chapter's strongest: **replaying the log must produce
the same final state as the live run, minus the animation.**

## §5.3 The inbound seam: the mailbox

Why the agent is deaf, stated precisely, because the obvious explanation is
wrong:

- It is NOT that the tool runs inline. The tool already runs on its own
  goroutine.
- It is that the caller parks in a `select` awaiting the result, and that caller
  is the same loop that would drain inbound events. **The drainer is parked.**
- Measured blackout in a real agent: 11.4 seconds.

Minimum mailbox:

1. One inbound queue carrying prompts, hints, interrupts, tool completions and
   child escalations as ONE event kind family.
2. Tool completion DELIVERED INTO the queue rather than awaited in a `select`.
   Nearly free — the jobs chapter's dispatch already registers every pending
   call.
3. A drain loop that never blocks on a tool. This is the real work.

Point to land: the hint channel is not a feature bolted onto the agent. It is
what falls out of not blocking.

Receipt: OpenAI ships "mid-turn steering" as a vendor primitive. The idea is not
exotic.

## §5.4 The actor: the loop between the seams

Deliberately small section. Define the actor as **a loop with a mailbox**, and
say plainly what is NOT being built: supervision trees, addresses, distribution,
restart strategies. "Actor" is a loaded word and the book should decline the
rest of the Erlang/Akka surface explicitly rather than imply it.

## §5.5 State is explicit, and waitable

Bill's requirement, and the primitive the rest of the surface derives from.

- The framework holds a state machine. Agent state is a FACT, not something
  parsed out of a history file.
- Observers can both receive transitions (push) and BLOCK until one matching a
  predicate occurs (wait). "Wait for end of turn" is the canonical case.

Then derive the entire waiting surface rather than enumerating it:

| operation                  | expressed as                                   |
|----------------------------|------------------------------------------------|
| blocking send message      | write to mailbox, then wait for turn-end       |
| wait for agent             | wait for next state change on one agent        |
| join agents                | wait for terminal state on every agent         |
| wake-any                   | wait for ANY state change or message, N agents |
| check progress             | read current state; no waiting at all          |

Join and wake-any differ only in predicate, not mechanism. When a surface derives
instead of enumerating, the primitive is usually right — worth saying out loud,
since the book keeps arguing that mechanism convergence is the signal.

**`submit_result` vs end-of-turn.** Keep these distinct in the prose: "the turn
ended" and "the agent produced a result" are different facts. Conflating them is
how a framework acquires a self-reported success flag. Terminal state says the
agent stopped. Submit says what it produced. NEITHER says the work was correct —
which is the same argument as the grader chapters, one level up.

**What this retires:** the `DONE:`-marker heuristic. A marker in text is a guess;
a state transition is a fact. (CodeRhapsody's own tool docs contradict each other
on exactly this point — one recommends grepping for `DONE:`, another calls it
fragile. Good, short, honest example.)

## §5.6 Supervision is observation

The chapter's biggest simplification, and it should be stated as a headline:

> The GUI, the history-file writer, and a parent watching a child are the same
> seam.

- A parent attaching to a child is an observer registration carrying an agent
  identity.
- Injecting a hint is a write into that child's mailbox — the same queue the
  user's hints arrive on.
- Therefore supervision needs NO new mechanism.

Bill's requirement, which is the design rule: **the parent must not have to read
the child's history file.** It observes current thinking, current chat, and
redacted tool calls, and can fetch the full payload on demand.

The failure this prevents, with the measurement: a supervision call that returns
unbounded history — 307,984 bytes in one call, roughly 13% of a context window —
because supervision was built as "read the history file." Its sibling call, which
has volume controls, returns 3,447 bytes for the same purpose.

Lesson, and it generalizes past this chapter: **a new waiting primitive must
inherit the volume contract, or supervision destroys the context window it was
meant to protect.** This is the jobs chapter's cap rule, re-applied at agent
scale.

## §5.7 Where the seam lives, and what it may import

Two rules, both checkable, both earned by the cold open.

**Rule 1 — not under `internal/`.** Go forbids other modules from importing
anything under `internal/`. A framework nobody can import is not a framework.
This is the compiler stating the architecture rule for us.

**Rule 2 — the seam package imports nothing outside the standard library.**
This is what actually minimizes exposure. It is also checkable: `go list -deps`
on the seam package returns stdlib only. A package that imports nothing cannot
participate in a cycle — the `gui_independence` guard generalized from "do not
import the GUI" to "do not import anything."

Corollary worth stating, because Chapter 2 already earned it: vendor types must
not appear in the seam. Vendor types in a signature were the tell in Chapter 2's
cold open; here the rule is mechanical instead of stylistic.

**Interfaces for behavior, concrete structs for data.** Observers, tool handlers
and vendor render/parse are interfaces. Events and parts stay concrete
serializable structs, because replay is a byte comparison and you cannot replay a
log of interfaces. Over-interfacing the data would silently break Chapter 2.

### The `Ref` type, and the Chapter 2 amendment

Three requirements collapse into one type, which is the section's payoff:

| case                      | reference form                        |
|---------------------------|----------------------------------------|
| multimedia, inline        | bytes                                  |
| multimedia, local         | file path                              |
| multimedia, remote        | URI — File API, `gs://`, external URL  |
| redacted tool call/result | handle into the io files or memory     |

All four are "the content is elsewhere, here is how to get it."

RULED: **amend Chapter 2** rather than patch around it. `BlobPart{MIME, Path}` is
local-path-only and cannot express three of Gemini's four input methods, nor
Anthropic's `file_id`. Chapter 2's `RedactedPart{Stub}` becomes a stub carrying a
`Ref`, at which point redaction is recoverable by construction instead of by
convention.

Media capability is per MODEL, not per vendor. Chapter 2 already models this with
one bool and a loud refusal; multimedia generalizes the bool into a capability
set and keeps the refusal. Verified matrix (2026-09-13): Gemini takes video in,
OpenAI does not — its video APIs are generation — and Anthropic takes neither
audio nor video, flattening animated GIFs to the first frame.

There is no honest degraded rendering of a video part, which is why refusal must
be loud.

## §5.8 Preview: the sub-agent surface, built not at all

Show the whole surface; build the one-agent case.

| operation       | meaning                                              |
|-----------------|------------------------------------------------------|
| spawn agent     | create a child                                        |
| send message    | deliver into a child's inbound queue                  |
| stop agent      | end a child                                           |
| wait for agent  | block on one child                                    |
| join agents     | block until every listed child finishes               |
| wait for agents | wake when ANY child sends a message or completes      |

**Why previewing belongs here:** the last row is epoll/select semantics and
cannot be implemented on top of blocking calls. If the framework's waiting
primitive blocks on one thing, the sub-agent chapter is unreachable later. The
preview is evidence that the mailbox is structural, not a convenience for hints.

The identity worth stating: **a parent waiting on N children is the same
mechanism as an agent draining its own queue.** One primitive, two scales.

**Why sub-agents are deferred past the skills chapter** — a real dependency, not
a preference: spawning takes a skill name. Without skills there is no way to say
what a child IS, so every spawn is a clone of the parent. Chapter order is a
dependency graph.

**Not foreclosing them, concretely:** a global sequence number plus an `Agent`
tag with `omitempty`; observers key by `(agent, id)`. Per-agent sequence numbers
were considered and rejected because reconnect then has to reconcile N
transcripts. Because the tag is `omitempty`, the single-agent case costs zero
bytes and zero concepts — the one-agent assumption forecloses nothing.

## §5.9 What this chapter deliberately cannot finish

The book's established closing move. Named, not built:

| mechanism             | why not here                             | where |
|-----------------------|------------------------------------------|-------|
| streaming             | needs a renderer that shows deltas       | after the GUI chapter |
| the GUI               | needs the observer seam to exist first   | next chapter |
| the gateway           | security and deployment material         | after the GUI chapter |
| dynamic workflows     | the exercise motivates it                | its own chapter |
| skills                | agent definition as data, not Go         | later |
| sub-agent management  | requires skills                          | later |

Closing line candidate:

> You now have a hint channel and nowhere to type into it.

## §5.10 Exercise: author, editor, reviewer

Three agents, a hardcoded Go workflow, each with its own tools and prompt.

**The point is the framework, not the writing.** If the seam is right, agents two
and three are nearly free. If they are not, the seam is wrong. The exercise
measures the chapter's own thesis.

Honest framing, per Bill's correction: the real version **worked OK**. The limit
is not that it breaks — it is that it is HARDCODED. Every new collaboration
pattern is new Go code. That is a better argument for dynamic workflows than
breakage, because it cannot be dismissed as a bug someone should have fixed.

Teach the distinction that keeps this consistent with deferring sub-agents: a Go
program constructing three agents and passing messages is NOT an agent spawning
children through its own tool surface. Multi-agent does not require a sub-agent
API — you can just write Go. That is worth a paragraph, and it makes the later
chapter's value obvious.

Technical note available from the real example (1,544 lines, verified): it
coordinates two agents with a tool that BLOCKS on a Go channel — "to the LLMs it
looks like a normal synchronous tool call." Elegant, and worth showing as a
contrast with the mailbox. Present it as an observation, NOT as the reason it
underperformed. It also enforces turn-taking in Go, returning an error if an
agent calls out of turn.

OPEN: how much of the three-agent workflow is given to the student vs written by
them, and whether the reviewer is a third actor or a second pass by the editor.

## §5.11 Grading

Properties, not pixels. Each must be checkable by deletion (P9).

1. **Replay equals live.** Replaying the log produces the same final state as the
   live run, minus deltas. Byte comparison.
2. **The seam imports only stdlib.** `go list -deps` on the seam package.
3. **The second agent adds zero lines to the framework package.** The chapter's
   thesis, mechanically checked.
4. **The agent is not deaf.** A message delivered while a tool is running is
   observed before the tool completes. This is the mailbox's whole point and it
   must be graded, or it is decoration.
5. **A state predicate wakes exactly once.** Wait-for-turn-end returns on the
   transition, not on a timeout and not on a text marker.
6. **Loud refusal.** A media part a model cannot accept fails loudly rather than
   being dropped.

Fake vendor is SHARED with earlier chapters — regression-check the earlier
graders after any change to it.

OPEN: point allocation.

## Figures used in this chapter, and their verification status

Re-derived 2026-09-13, not proofread from memory. This book has already shipped
several wrong numbers that survived proofreading, so the rule is to re-measure.

VERIFIED against the artifact:
- `internal/agent` is 112 `.go` files; `pkg/` holds exactly one package.
- `examples/author_editor/main.go` is 1,544 lines.
- The supervision call returned 307,984 bytes in one call (~13% of a 575K
  context window); its sibling with volume controls returned 3,447 bytes.
  Receipt: cr/BUGS lines 153 and 158 in the CodeRhapsody repo.
- The multimedia matrix — Anthropic read from body text, OpenAI from nav only.

ASSUMED, needs a receipt before it is printed:
- The **11.4 second** blackout. Consistent across session memory and reported as
  measured, but no artifact was located this session. Either find the
  measurement or state it qualitatively.

NAV-LEVEL ONLY (a page exists; body not read):
- Anthropic "fine-grained tool streaming"; OpenAI "mid-turn steering". Fine as
  supporting receipts, but read the body before quoting specifics in prose.

## Open questions

1. Chapter title. Candidate above; Bill said the two-seams framing SGTM but has
   not ruled on wording.
2. Exercise scope — see §5.10 OPEN.
3. Package path for the seam in the course repo. Rules are settled (§5.7); the
   literal path is not.
4. Point allocation for §5.11.
