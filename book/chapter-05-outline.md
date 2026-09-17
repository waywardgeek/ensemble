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

Second instance, same failure, same week — and it is the stronger one because
nobody chose it as a design: **the public surface did not exist until a real
consumer forced it.**

- The bulk of the framework lives under `internal/`, where Go forbids other
  modules from importing it. `internal/agent` alone is 112 files.
- A root facade was retrofitted 2026-08-24 through 08-28, the first commit
  titled *"root facade with a go/ast guard that changed the design."* Note the
  guard did not merely enforce a decision; it CHANGED it.
- It runs straight into the GUI commits of the same week, one of which is *"no
  GUI left in the public API"* — so the GUI had leaked into the public surface
  too, not only into the tests.
- **The facade was measured, not designed from taste.** A real consumer
  (homebrew-vtt, 141 Go files) was ported and the compiler enumerated what it
  actually touched: 21 symbols, 455 references — tool authoring 300 (66%),
  sandboxed file I/O 80 (18%), media 52 (11%), agent lifecycle 23 (5%).

That measurement is the chapter's thesis arriving as evidence instead of
assertion: **two-thirds of a real consumer's use of an "agent framework" is
declaring tools, and five percent is controlling agents.** It is a
tool-authoring SDK. Design the seam accordingly.

Do NOT write that the framework "cannot be imported by anyone" — that was true
before the facade and is false now. The honest version is better: it went months
without a public seam, and cutting one took a five-day campaign that is still
ratcheting.

This is deliberately dropped here and paid off in §5.7.

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

**Chapter 5 does not introduce agent state. It EXPOSES it.** The reader has had a
state machine since Chapter 2 — `TurnState` with `Idle`, `InputPending`,
`InFlight`, `ToolsPending` — and Chapter 4 added `Interrupted`, deliberately as a
STATE rather than a flag, "or replay re-executes tool calls that were cancelled."
This is another reveal, not an introduction, and it should be written that way.

- The framework holds that state machine. Agent state is a FACT, not something
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

**Rule 1 — `internal/` for the bulk, plus a MEASURED facade.** The naive version
of this rule ("don't use internal/") is wrong, and CodeRhapsody demonstrates the
right one. Hide the implementation where Go forbids others from reaching it, and
export a deliberately small public surface. Then the hard part, which is the
actual lesson:

> Do not design the facade from taste. Measure it.

CodeRhapsody's was derived by porting a real consumer and letting the compiler
enumerate what it touched — 21 symbols, 455 references, two-thirds of them tool
authoring. Taste would have produced a symmetrical API with an elegant agent
lifecycle; the measurement said agent lifecycle is 5% and tool authoring is 66%.

Two corollaries worth printing, both from that facade:

- **Aliases, never wrappers.** `type Tool = common.Tool`, not a struct that
  copies it. An alias IS the same type, so a value crossing the seam needs no
  conversion and cannot drift from the implementation it names. A wrapper is two
  types that must be kept in sync forever, and buys nothing here.
- **Enforce it mechanically.** The rule that no exported symbol may mention an
  unaliased internal type is a TEST, not a convention. Third instance in this
  codebase of architecture enforced by a guard rather than by discipline — and
  the guard on the facade is documented as having CHANGED the design, not merely
  checked it.

**Rule 2 — the seam package imports nothing outside the standard library.**
This is what actually minimizes exposure. It is also checkable: `go list -deps`
on the seam package returns stdlib only. A package that imports nothing cannot
participate in a cycle — the `gui_independence` guard generalized from "do not
import the GUI" to "do not import anything."

Corollary worth stating, because Chapter 2 already earned it: vendor types must
not appear in the seam. Vendor types in a signature were the tell in Chapter 2's
cold open; here the rule is mechanical instead of stylistic.

**Interfaces for behavior, concrete data for data — with a precision that
matters.** Observers, tool handlers and vendor render/parse are interfaces.
Events and parts use Chapter 2's existing pattern: a SEALED union (`isPart()`)
plus an ordered-field JSON envelope. A sealed union serializes and replays fine;
what cannot be replayed is an OPEN behavioral interface. Do not state the rule as
"no interfaces in the data" — Chapter 2's `Part` IS an interface, and the book
must not contradict code the reader already has.

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

**RULED (Bill, 2026-09-13): the exercise gives NO code.** The exercise is a
problem statement written in enough detail that a good AI coding agent, handed
it verbatim, could succeed. Every prior chapter's exercise handed over a
contract and let the student find the design; this one hands over the design in
prose and lets the student's agent find the code. The chapter states the reason
out loud: the student is meant to become the expert in the code base, and there
are exactly two ways to do that with an AI writing the code. Type the prompts
yourself, one at a time, guiding the agent and reading what comes back; the
book asks, as close to begging as prose allows, for this one. Or paste the whole
statement and review every generated line, which is what most working engineers
do today and is the fallback, not the goal. What the book will not endorse is
the third option nobody admits to: paste, run the grader, ship.

**RULED (author): the reviewer is a third actor, not a second editor pass.** The
chapter's checkable claim is that a new agent adds nothing to the framework.
Two actors cannot test that claim, because the second one IS the framework's design
cost. Only the third shows whether the cost was paid once.

**Exercise contract** (the only thing that is fixed; contents are the student's):

- Directory `seam/` holds the seam types from §5.7, typed in from the chapter.
  It imports nothing outside the standard library (§5.11 property 2).
- Directory `framework/` holds the actor loop. It imports `seam/` and nothing
  from any agent (§5.11 property 3).
- One binary, `./ch05`, runs the three-agent workflow. Inbound events arrive as
  JSON lines on stdin; observations leave as JSON lines on stdout; the event log
  is written to the path given by `CH05_LOG`. Stdin and stdout are the two seams
  from §5.1 made into pipes, and they are the door the grader walks through to
  prove the agent hears while a tool runs. The GUI chapter will grade the same
  door, so nothing here is thrown away.
- Vendor: the fake from chapters 2–4, unchanged, so the grader controls what the
  model "says" and can request a slow tool on cue.

The problem statement itself (what author/editor/reviewer do, what tools each
has, what "done" means) is the last section of the chapter and is written to be
handed to an agent. The seam-draft Go from §5.7 appears in the chapter body with
its prose; the student's agent will need the student to point it at the right
page.

## §5.11 Grading

Properties, not pixels. Each must be checkable by deletion (P9), and every
check carries points (P9 corollary, `5001a76`). Points are the author's
allocation (ruled: author's choice). 100 total, seven checks.

| Check | Pts | What the grader does |
|---|---|---|
| `ch4parity` | 10 | Re-runs the ch4 grader against the student's tree. A regression gate is worth 10 in every chapter that has one (ch2→ch3 precedent). |
| `not-deaf` | 25 | Fake vendor requests a slow tool (a job that takes seconds). While it runs, the grader writes a `Hint` to stdin. The observation stream on stdout must show the hint observed BEFORE the tool's `ToolCompleted`. The chapter's thesis; it carries the most weight. A blocking select passes every other check and fails this one. |
| `replay-is-live` | 15 | Byte-compare the context rendered from the event log against the context the live run sent to the fake. Deltas must not be in the log; finalized parts must be. Ch2's discipline carried forward. |
| `framework-blind` | 15 | `go list -deps ./framework/` contains no package from any agent directory. Dependency DIRECTION is the mechanical form of "a new agent adds zero lines to the framework". Line counting is not checkable without a before, and the grader has no before. |
| `wake-once` | 15 | Two children finish in the same turn; the grader counts the workflow's wakeups on the observation stream. Exactly one. Two is the polling loop the mailbox was supposed to replace; zero is a timeout or a text marker. |
| `seam-stdlib` | 10 | `go list -deps ./seam/` is standard library only. Split from `framework-blind` because a student can get the direction right and still let a vendor type leak into the seam; different mistake, different fix. |
| `loud-refusal` | 10 | Present a `BlobPart` whose `Media` the model's table row does not declare. Renderer must refuse with an error naming the model and the media. Dropping the part or substituting text scores zero here. |

Budget per skill, then split across ids (P9 corollary): mailbox 40 (`not-deaf`,
`wake-once`), dependency discipline 25 (`framework-blind`, `seam-stdlib`),
logging 15, rendering 10, regression 10.

Deletion audit: `not-deaf` and `wake-once` are the checks most likely to be
graded by nothing if the fake's slow tool is not actually slow. The audit must
include a mutant that makes the tool instant and confirm a blocking student
still fails both. Ch1–3 each shipped with their loudest rule ungraded; this is
the chapter whose loudest rule is hardest to grade, so the audit goes first,
not last.

Grader cost: the fake vendor from ch2–4 already serves three vendors on one
path, and the ch3 grader already scripts a tool request for its ordering check.
Nothing new is served. The observation stream is JSON lines on stdout by
contract (§5.10), so every check except the two `go list` checks reads one
file.

Fake vendor is SHARED with earlier chapters — regression-check the earlier
graders after any change to it.

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

All four closed 2026-09-13.

1. **Title: "The Agent Framework."** RULED (Bill).
2. **Exercise scope.** RULED (Bill): no code given; a problem statement good
   enough to be a prompt; the book asks the student to type the prompts, not
   paste them. Reviewer as third actor is the author's consequence (§5.10).
3. **Seam location.** RULED (Bill): in the chapter, not the repo. The seam is
   read and discussed as prose; the student's agent types it in. The only fixed
   thing is the contract directory name `seam/`. Question 3 as asked
   (a repo path) is dissolved, not answered.
4. **Points.** RULED (Bill): author's choice. Allocated in §5.11.

Remaining, carried from the seam draft (Go-level, not outline-level):
`Chunk` string vs `[]byte`; is `Submitted` an `Observation`; `partJSON` field
names. `Wait` on Agent vs Watcher was resolved by the workflow seam
(`f7cc54e`): the condition is data, invoked as a tool.
