# Chapter 5 seed — The Agent Framework

Status: SEED. Content ruled by Bill 2026-09-13. Not an outline yet.

Bill's ruling, verbatim:

> "Chapter 5 should be creating the agent framework. We should introduce actors,
> mention but not build sub-agent management, and the observer pattern for
> listeners, like the GUI, or history file generator. The payoff for the students
> can be writing a different agent, not an AI coding agent, as the example for
> ch 5. If we need the mailbox thingy, this would be the chapter for it."

Ordering rationale, also his: the agent/actor seam must exist BEFORE the GUI is
built on top of it. The GUI chapter follows this one.

---

## Thesis

An agent framework is exactly two seams and a loop between them.

- **Outbound — the observer seam.** What the agent says. The GUI, the history
  file writer, a logger, a metrics sink are all the same kind of thing: a
  listener on one event stream. None of them is known to the core.
- **Inbound — the mailbox.** What the agent hears. Prompts, interrupts, and tool
  completions arrive as one event kind on one queue.
- **The actor** is the loop between them. Defined minimally: *a loop with a
  mailbox*. See "Declined" below.

This is the same shape as Chapter 2 ("one log, three vendors"), which is why the
chapter REVEALS rather than introduces: students already wrote an observer in
Chapter 2 without calling it one, because `history.md` is a rendering of the
event stream. Chapter 2's renderer was the first observer.

## What the framework is FOR (Bill, 2026-09-13)

> "Like CodeRhapsody, the framework will evolve into a tool authoring system,
> more than an agent management framework."

This reframe matters for the chapter's emphasis. The framework's primary value is
authoring TOOLS, not managing agents. It also justifies the exercise: a different
tool set, plus a system prompt, plus observers, IS a different agent. Agent
management (sub-agents) stays declined; tool authoring is the through-line from
the tools chapter and the jobs chapter.

## Why the mailbox belongs here, and is not optional

Observers are outbound only. They do nothing for what the agent hears. Ship the
observer seam alone and the agent broadcasts perfectly to the GUI and history
writer while remaining deaf for the entire duration of a tool call (measured
blackout: 11.4s).

The deafness is NOT caused by the tool running inline — the tool already runs on
its own goroutine. It is caused by the caller parking in a `select` awaiting the
result, and that caller is the same loop that would drain inbound events. The
drainer is parked, so the agent goes deaf.

Minimum mailbox, as settled in cr/docs/tool-calls-as-jobs-design.md (CodeRhapsody
repo):

1. An inbound queue carrying prompts, interrupts and tool completions as one
   event kind.
2. Tool completion DELIVERED INTO that queue instead of awaited in a `select`.
   Nearly free: the jobs chapter's dispatch already registers every pending call.
3. A drain loop that never blocks on a tool. This is the real work.

Consequence for chapter order: the GUI chapter needs a hint channel to already
exist before it can give the user somewhere to type into. Candidate closing line
for this chapter: *you now have a hint channel and no way to type into it.*

## The war story (cold open candidate) — VERIFIED

This is the chapter's spec, not decoration. CodeRhapsody built the GUI before it
had an agent/framework seam, and the framework ended up depending on the GUI.

Evidence, all in the CodeRhapsody repo:

- `internal/agent/gui_independence_test.go` exists solely to enforce that no
  non-test file in the framework imports `gui_server`. Its header: the edge "was
  real until 2026-08-28: internal/agent/test_support.go built a real GUIServer
  for the ~23 tests that need a UI attached."
- The fix was the observer pattern. Five commits, 2026-08-28/29:
  - `b1ead0f9` The framework no longer depends on the GUI, and a ratchet keeps it that way
  - `94a18bc0` Move secret crypto below the seam, into pkg/common
  - `957f5df3` Test scaffolding takes a UIObserver, not a GUI server
  - `6c6f0eb1` Make the GUI conformance suite a public contract for embedders
  - `38b784d6` Ratchet the framework test files that still import the GUI
- **It is not finished.** The guard exempts test files and ratchets them downward
  instead of forbidding them, because "those edges cannot all be removed at once
  and a guard that fails today guards nothing." The header states the remaining
  edges are "a real remaining blocker for the module split, not untidiness."

The non-obvious technical point, worth the page on its own: at PACKAGE level a
test edge is harmless, because a test binary is in no other package's import
closure. At MODULE level a framework test importing the GUI forces the
framework's `go.mod` to require the GUI module, which already requires the
framework — a cycle. The seam can look clean while the dependency graph is not.

Note the fix's shape: the GUI was not deleted, it was DEMOTED — from a component
the framework knows about, to an observer the framework broadcasts at.

## The payoff / exercise: a different agent

Bill's idea, and it is stronger than a payoff — it is the chapter's falsification
test. Writing a non-coding agent is the only honest check that the seam is real.

Gradeable without grading pixels: **the second agent must add zero lines to the
framework package.** If it cannot, the seam failed and the grader says so. This
is the P9 shape — the claim is checkable by deletion, not by assertion.

Its size is itself the measurement. A good seam makes the second agent small.

Constraints on the choice: it must use genuinely different tools (or the seam is
not exercised) and stay deterministic against the fake vendor (or it is not
gradeable).

OPEN — Bill's call. Recommendation: a scheduler/reminder agent. It has a clock
instead of a filesystem, notification output instead of edits, and it exercises
the observer seam from the other side because the notifier IS an observer.
Receipt that this is not hypothetical: Puffin is a real non-coding agent built on
the CodeRhapsody agent library.

## Constraints on the seam (Bill, 2026-09-13)

### The seam ships as Go source, inline in the chapter

Not prose, not an appendix. Consequence: the type definitions get PRINTED, so
every later capability (streaming, multimedia, sub-agents) must be accommodated
by the types now. A type we have to change in a later chapter is a broken promise
in print. This is the chapter where getting the struct right matters most.

### Observable streams

Four kinds flow through the observer seam:

| stream              | streams incrementally? |
|---------------------|------------------------|
| chat                | yes                    |
| thinking            | yes                    |
| tool call parameters| yes                    |
| tool results        | NO — always whole      |

Streaming SUPPORT is added after the GUI chapter. The seam must not foreclose it.

**Design rule: do not model streaming as a mode. Model non-streaming as a stream
of length one.** A part gets a stable identity; deltas append to it; a terminal
marker finalizes it. A non-streaming vendor emits one delta plus the finalizer.
Observers that do not care about liveness ignore deltas and act on the finalizer;
the GUI renders deltas. Adding streaming later then adds NO new event kinds, only
a different chunk count.

Tool results validate this rather than complicating it: "observable but never
streaming" is just the length-one case, which needs no special branch.

**The observer stream is NOT the event log.** Deltas are transient and are never
appended to the durable log; only the finalized part is recorded. This is what
keeps Chapter 2's replay promise intact and honors "no field grows without
bound." Testable property, and the natural grader: replaying the log must produce
the same final state as the live stream, minus the typing animation.
The log is the state; the stream is the experience.

RULED: a tool's output growing over time (the jobs chapter's callbacks) is NOT
the same channel as model streaming. I proposed unifying them; Bill declined.

### Multimedia input

The seam must support audio, images, and video as INPUT.

Verified 2026-09-13 against ai.google.dev/gemini-api/docs/file-input-methods:
video input is real today, not speculative. Gemini accepts "images, audio, video,
and documents" and ships dedicated video-understanding endpoints. Its part shape
is `{type, mime_type, data | uri}`. Input methods: inline base64 (100MB/request,
50MB PDFs), File API upload (2GB, 48h), registered GCS URI, and external URLs.
NOT YET CHECKED: Anthropic and OpenAI video support. Do not write a
cross-vendor claim until both are verified.

Two consequences for Chapter 2's existing types:

1. `BlobPart{MIME, Path}` is local-path only. It cannot express a File API URI, a
   `gs://` URI, or an external URL — three of Gemini's four input methods. The
   reference needs to be *either* local bytes or a remote URI. OPEN: fix in ch5,
   or amend ch2? Ch2 is already published-shaped, and P1 says architecture is
   additive but editing prior code is allowed.
2. `seam.go` already carries `AcceptsAudio bool` and REFUSES to render an audio
   part to a model that does not accept it, rather than dropping it silently.
   Multimedia is therefore not a new mechanism: it is generalizing one bool into
   a capability set, keeping the loud-refusal rule. Good chapter beat — Chapter 2
   planted the seed with one bool and one refusal.

### One agent, without foreclosing sub-agents

This chapter assumes a SINGLE agent. Sub-agent management is mentioned, not built
(see Declined). But the seam must not have to change when sub-agents arrive.

Proposed mechanism, carried over from the Counterpoint swarm design (2026-08-14).
Marked ASSUMED for this book — the decision is verified in that context, not yet
re-derived here:

- Keep a GLOBAL sequence number. Per-agent sequence numbers were considered and
  rejected there because reconnect then has to reconcile N transcripts.
- Add an `Agent` tag, `omitempty`. Observers key blocks by `(agent, id)`.
- Because it is `omitempty`, the single-agent case costs zero bytes on the wire
  and zero concepts in the chapter. The one-agent assumption forecloses nothing.

The failure mode to avoid is a hardcoded global event bus with no agent identity
and observers that cannot be attached per agent. That is the design that breaks.

## Declined (P6 — named, not built)

- **Sub-agent management.** Mentioned only. Once actors and a mailbox exist,
  sub-agents look trivial and are not. The trap to name, from
  book/subagents-seed.md: a sub-agent is an agent you are no longer watching.
  Delegation converts a continuously-corrected process into a specified-once
  process, because the parent supervises by polling and is usually BLOCKED inside
  the call that waits for the child.
- **Actor-model machinery**: supervision trees, addresses, distribution,
  restart strategies. "Actor" is a loaded word (Erlang, Akka). Define it
  minimally as a loop with a mailbox and decline the rest explicitly.
- **Workflow APIs.** Bill: "we also must provide the workflow APIs, but maybe in
  a separate chapter." Required for the framework to be complete, deferred to
  its own chapter. OPEN: does the ch5 seam need to carry anything for workflows,
  the way it must carry the `Agent` tag for sub-agents? Answer before the Go
  types are printed — a workflow hook added later is a broken promise in print.

## The sub-agent API: previewed in full, built not at all

Bill's ruling: show readers the whole seam here, even though this chapter builds
the one-agent case. The six operations the seam must eventually carry:

| operation        | meaning                                                        |
|------------------|----------------------------------------------------------------|
| spawn agent      | create a child agent                                            |
| send message     | deliver into a child's INBOUND QUEUE                            |
| stop agent       | end a child                                                     |
| wait for agent   | block on one child                                              |
| join agents      | fan-in: block until every listed child finishes                 |
| wait for agents  | wake when ANY awaited child sends a message OR completes        |

(CodeRhapsody ships these as spawn_sub_agent, send_message, shutdown_agent,
wait_for_agent, join_agents, wait_for_agent_change — a receipt that the surface
is real, not invented for the book.)

### Why previewing this belongs in THIS chapter

The last row is the argument for the mailbox, and it is stronger than the hint
argument.

**"Wake when any awaited agent sends a message or completes" is epoll/select
semantics, and it cannot be implemented on top of blocking calls.** If the
framework's waiting primitive blocks on one thing, the whole sub-agent chapter is
unreachable later. So the preview is not decoration: it is the evidence that the
mailbox is structural rather than a convenience for hints.

Deeper identity worth making explicit in the prose: a parent waiting on N
children is the SAME mechanism as an agent draining its own inbound queue. Both
are "block until any event arrives from a set of sources." One primitive, two
scales. Build it once in this chapter and the sub-agent chapter is mostly naming.

Note also what `send message` is, mechanically: the parent writing into the
child's inbound queue — the same queue the user's hints and the tool completions
arrive on. One primitive carries all three. That is the case for the mailbox
being the center of the framework rather than an accessory to it.

### It also answers Bill's own objection to sub-agents

From book/subagents-seed.md, his words: "This is one reason I don't use
sub-agents often. I can normally send a hint quickly before an agent wastes a lot
of cycles."

The reason he cannot is mechanical, not cultural: the parent supervises by
POLLING and is usually BLOCKED inside the call that waits for the child, so
delegation converts a continuously-corrected process into a specified-once
process. `wait for agents` with wake-on-message is exactly what converts it back.
The mechanism that makes sub-agents worth using is the same mailbox this chapter
builds. That is the sub-agent chapter's thesis, set up here and paid off there.

### Evidence from a built surface: CodeRhapsody's own sub-agent tools

Bill: "You should examine your own tool surface for sub-agents for that."

The real surface is 14 calls, not six. Parent-side: spawn_sub_agent, ask_agent,
agents_status, get_sub_agent_status, send_message, wait_for_agent,
wait_for_agent_change, join_agents, get_submission, interrupt_agent,
shutdown_agent, kill_agent, respond_to_parent_message. Child-side:
send_message_to_parent.

What matters for the seam is not the list. It is six properties, each of which
constrains the one-agent design we ship in this chapter:

1. **The channel is bidirectional.** `send_message_to_parent` BLOCKS the child
   until the parent replies via `respond_to_parent_message`, and the parent
   receives that question as a `message_to_parent` EVENT in the same wait that
   delivers completions. So a mailbox is needed on both sides, and "child asks
   for a decision" is just another event kind on the one inbound queue. Design
   consequence: the event kind must be open enough to carry an escalation, not
   just prompts and completions.

2. **There is deliberately no success flag.** `join_agents` reports an `outcome`
   observed BY THE FRAMEWORK: submitted, finished, finished_without_submit,
   timed_out, killed, deactivated, crashed, unknown_agent, unobservable. The
   documentation is explicit that `submitted` does not mean the work succeeded,
   and that a framework-level success flag is withheld on purpose "because a
   self-reported one invites an agent to tick its own box." This is P9 and the
   green-dashboard theme restated at agent scale. Strong book material.

3. **You cannot cancel, only contain — again.** `kill_agent` documents its own
   limit: it "cannot stop a Go tool handler the agent is blocked inside — Go
   cannot kill a goroutine." It guarantees the agent starts no new turns, not
   that the process is idle. This is the SAME constraint as the watchdog war
   story (containment, not cancellation). The jobs chapter's central limit
   reappears one level up, which is an argument for teaching it as a property of
   supervision in general rather than a quirk of tools.

4. **Capabilities narrow, never widen.** safe_mode, enable_web_search and
   enable_reasoning are tri-state: omitted means inherit. A sandboxed agent
   cannot spawn an unsandboxed child; attempts fail loudly rather than quietly
   downgrading, "so spawning must not become a way around your own limits."
   That is a security property and it is the security chapter's hook into this
   one.

5. **Results are durable because they once were not.** `get_submission` prefers
   the live result and falls back to `submit.md` on disk, explicitly because
   payloads used to live only in the child's memory: if the parent never
   collected, or collected and then lost its own context, the work was gone.
   Same lesson as the observer seam — durable state belongs in a log, not inside
   a live actor.

6. **A live contradiction in the built surface, worth printing.**
   `agents_status` and `get_sub_agent_status` both advise that for "authoritative
   completion detection" you should grep for `DONE:` in the child's history file.
   `wait_for_agent_change` documents the opposite: it detects completion via the
   AI client's turn-complete callback, "not via fragile DONE: markers." The older
   advice survived in the sibling docs after the mechanism that made it necessary
   was replaced. Status APIs drift; the newer primitive is right.

   Compounding it: `wait_for_agent_change` returns UNBOUNDED history — measured
   at 307,984 bytes in a single call, roughly 13% of a context window — while its
   sibling `wait_for_agent` has both `max_history_chars` and `content_filter`.
   The epoll-shaped primitive did not inherit the output-volume contract. The
   jobs chapter establishes caps for tool output; the supervision surface
   re-violates them at agent scale. Lesson for the seam: a new waiting primitive
   must inherit the volume contract, or supervision itself destroys the context
   window it was meant to protect.

### Why sub-agents are DEFERRED past the skills chapter

Bill's ruling: "sub-agents need agent-defining skills, which is a later chapter,
so sub-agents will be deferred."

This is a real dependency, not a preference. `spawn_sub_agent` takes a
`skill_name`; without skills there is no way to say what the child IS, so every
spawn is a clone of the parent. The interesting cases — a narrow tool-less judge,
a domain persona, a worker with a deliberately narrowed toolset — are all
expressed as skills. Chapter order is a dependency graph: skills must precede
sub-agents.

What this chapter does instead: preview the surface, build the one-agent case,
and make sure the waiting primitive is epoll-shaped so the later chapter is
reachable.

## Risks

- **Scope.** Actors + mailbox + observers + framework extraction + a second agent
  is the largest chapter in the book. Mitigation: the body builds the two seams;
  the second agent is the exercise, not body content; sub-agents are mentioned
  only.
- **Naming.** "Creating the agent framework" describes the activity, not the
  idea. A title along the lines of the two-seams thesis would teach from the TOC,
  per the established Title+Subtitle pattern.

## Effect on earlier chapters: none

The jobs chapter already closes by naming actors as what comes next, and its
forward table already lists the mailbox as not-built, next-chapter. No reversal
is needed. Its "What this chapter deliberately cannot finish" section stands.

## Open questions

Ordered by what blocks the Go types, since those print inline and cannot be
quietly changed later.

1. **`BlobPart` reference shape.** Local `Path` cannot express a File API URI, a
   `gs://` URI, or an external URL. Fix in ch5, or amend ch2? BLOCKS THE TYPES.
2. **Does anything workflow-shaped have to exist in the seam now?** Same class of
   question as the `Agent` tag. BLOCKS THE TYPES.
3. **Anthropic and OpenAI multimedia support** — video and audio input, verified
   against live docs. Only Gemini is verified so far. Needed before any
   cross-vendor sentence is written.
4. **Which non-coding agent** for the exercise? (recommendation: scheduler /
   reminder agent, above)
5. **Chapter title.** "Creating the agent framework" names the activity, not the
   idea; the two-seams thesis is the candidate.
6. Does the observer seam precede the second agent, or does the second agent
   motivate it? (Suspect: observers first, because the second agent's notifier
   is one.)
7. Rename `book/chapter-05-actors-parking.md`? Its filename is now correct again
   after three renumberings, but its internal header still says Chapter 4.
