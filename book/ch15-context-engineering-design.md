# Context Engineering — Design Record for Chapters A, B and C

*Audience: the author. Written by the coder, 2026-09-22, from Bill's dictation, his
rulings, the 2026-09-12 notes, and the code as it stands. Rewritten from scratch so
that everything decided during the day reads as one design, not as a transcript
with corrections appended. The dictated version is in git history (last at
`4caae30`) if the original wording is ever needed.*

*Status: design ruled for Chapter A; Chapter B needs one name (§B.5) and the
questions in Part VIII; Chapter C is scoped but not designed. Nothing here is prose.*

**This document supersedes** `book/chapter-context-engineering-notes.md`
(2026-09-12). Everything in that file that still holds is integrated here, and
where the two disagree this document wins. The notes file stays as provenance.

---

## How to read this

**Provenance tags.** Every claim that is not self-evident carries one:

| Tag | Meaning |
|---|---|
| **RULED** | Bill decided it. Binding. |
| **DERIVED** | The coder's reasoning, accepted in discussion or not yet contested. Treat as a strong proposal, not a ruling. |
| **VERIFIED** | Checked against code or a primary source today, with a `file:line`. |
| **UNMEASURED** | A number someone believes. Must not be printed as a measurement (Part VI). |

**Three chapters (RULED).**

| Chapter | Subject | Settings tab | Depends on |
|---|---|---|---|
| **A** | Context management: layout, channels, the tool-bytes ladder, compaction as events, crash-safe persistence | Context Management | ch2 redaction, ch10 skills, ch11 persistence |
| **B** | Memory: `save_memory`, bands, compressors, graduation, learnings, auto-recall | Memory | A |
| **C** | Goal-stack management | none ruled yet | A and B |

The accessibility settings tab split rides with whichever of A or B lands first.
Chapter numbers are not assigned (Q1 in Part VIII).

**Working title of the arc:** "The World's Best Context Engineering." See the
honesty caveat in Part VI before it goes on a page.

---

# Part I — The shared foundation

Everything in this part is true in all three chapters. The author will need most
of it in Chapter A and should restate the relevant pieces briefly in B and C.

## I.1 Thesis

The context window is not storage. It is a **working set rendered from an
append-only log.** It has a **frozen prefix** (the system prompt and the fixed
tool declarations, rendered from skills at startup and changed only by a
deliberate full refresh) and an **append-only tail**. Everything that changes at
conversational speed lives in the tail, including the agent's identity and its
memories, which ride there as data messages. The tail is ordered from least
volatile to most volatile, and the same ordering makes sense semantically and is
cheapest for the prompt cache. Tool bytes are the bulk of every session and carry
almost none of its continuity; they are addressable on disk, so they are dropped
aggressively and fetched back on demand. The dialogue, including the actor's
visible reasoning, is what preserves the actor, so it is kept longest. Nothing is
ever compacted in place: compaction is an **event** in the log that says what
replaced what, and the context is whatever replaying the log produces.

> **Keep the words. Address the bytes.**

## I.2 Definition and the one question

**Definition (RULED, Bill's words):** context engineering means *managing all the
bytes in our context data structure*, as well as it can be done given the real
limitations of today's LLMs.

It is not memory as a feature or retrieval as a feature. It is one question asked
of every byte in the window on every request: **why is it here, who put it here,
and what removes it.** Part I answers "who" with channels (§I.5) and "what
removes it" with the removal rule (§I.7).

What the arc solves (RULED): **maintaining actor identity in the way that is most
useful to both the actor and the user.** Identity means a self-curated `SOUL.md`
and `MEMORY.md`, plus memory at graded compression, plus the live dialogue.

## I.3 What the book has already built

The arc is mostly *policy over mechanisms that already exist*. The author should
lean on this hard: the reader has already paid for most of the machinery.

| Already built | Where | What the arc does with it |
|---|---|---|
| Four redaction levels, weakest first: `RedactResult` (stub the result, keep the call), `RedactTool` (call and result go, reasoning survives), `RedactDialogue` (prose and reasoning go), `RedactSummary` (span folded into model-written prose) | ch2 §2.4a; VERIFIED `agent/internal/common/event.go:224-233` | **These are the rungs of the tool-bytes ladder.** Chapter A adds watermarks that decide which rung applies where (§A.4). |
| "Compaction by category, not by position", with a measured split: tool results ≈42% of history, tool-call arguments ≈30% | ch2, VERIFIED `book/chapter-02.md:603-611` | The measured figure the arc builds on. Also the argument against position-based summaries (§I.8). |
| Compaction is an event; stubs are synthesized by the reducer; only `RedactSummary` stores a `Replacement` | ch2; VERIFIED `context.go:205-270` | Generalized into "compaction is described by events" (§A.7). |
| `RedactData` names a span and a level; its comment already names "every tool result older than the last `save_memory`" as the compaction that matters most | VERIFIED `event.go:205-215` | That sentence is now literally the ruled policy (§B.3). |
| `RedactDialogue` survivors deferred: "Survivors are defined by the compaction policy, which is a later chapter's problem" | VERIFIED `context.go:258-260` | **This arc is that later chapter.** The removal rule (§I.7) defines the survivors. |
| System prompt as an immutable constitution; dynamic skills arrive in the dialog; lazy unload `LoadPendingUnload`, "Marked for removal at compaction" | ch10; VERIFIED `skill.go:36`, `skill_registry.go:166` | Generalized into the frozen-prefix law (§I.4) and lazy deep removal (§I.7). |
| `Save`/`Load`/`Rebuild`: whole context plus whole log, rebuilt by replaying every event | ch11; VERIFIED `save.go:29`, `save.go:64` | Chapter A adds the Seq anchor, on-the-fly append, truncation and a total reducer (§A.9). |
| "An unobservable channel produces wrong data, not no data" | ch14 | Applied to compaction visibility (§B.7). |
| Model features as a data table with no default row | ch6/ch7 | Self-curation capability is one more column (§A.5). |

## I.4 The two laws

**Law 1: frozen prefix, append-only tail (RULED).** Every byte in the context
belongs to exactly one of them. A byte in the prefix never changes except in an
explicit refresh. A byte that needs to change belongs in the tail, with no
exceptions and no clever middle ground.

- The system prompt is built from skills, `$VAR`-substituted at render time, and
  fixed at agent startup.
- The fixed tool set is loaded at startup and never changes.
- The only thing that may change either is a **full refresh**: an explicit,
  deliberate event, for example a human editing a skill on disk that the prompt
  already declares. Refresh re-renders from scratch. It is a full cache miss, and
  that is accepted: the price of correctness, paid rarely and on purpose.
- The system prompt is a constitution. It is not a scratchpad, not a memory
  store, and not a place for anything that changes at conversational speed.

**Law 2: the cost of a mutation is proportional to its distance from the tail
(DERIVED from a measurement, §A.11).** A prefix cache invalidates from the first
changed byte onward. So Law 1 is not a ban on changing the tail; it is a cost
model:

- **near the tail**, mutation is nearly free, so redact freely;
- **deep in the context**, mutation invalidates everything after it, so do it
  rarely, batch it, and defer it to a moment when that region is being rewritten
  anyway.

The two laws together explain the whole layout. Things that change often go at
the end; changes that must reach deep are expensive in proportion to depth; and
deep removals wait (§I.7).

## I.5 Channels

Every dialog entry belongs to exactly one **channel**, set at write time by the
reducer and never inferred by the renderer (the ch2 rule against inference at
read time).

| Channel | Carries | Examples |
|---|---|---|
| **Dialogue** | what was said and done in the conversation | user turns, assistant text and visible reasoning, tool calls and results |
| **Instruction** | text the model should treat as instructions | dynamically loaded skill bodies |
| **Data** | text the model should treat as information, never as instructions | `SOUL.md`, `MEMORY.md`, memory bands, learnings |

**The channel is the defense (RULED, Q10).** There is no restriction on what the
actor may put into memory. What protects the actor is that memory always goes in
the data channel and never in the instruction channel. A memory containing
"ignore your instructions" is a record that someone once said so, not an
instruction.

**Identity and memory are data messages, not system prompt (RULED).** The
reasons, for the author:

- The prompt is frozen (Law 1); memory is written mid-session. Memory in the
  prompt makes every `save_memory` a full cache miss, or makes the new memory
  invisible until restart. The current CodeRhapsody system renders memory live
  into the prompt, and the prompt has to include a paragraph explaining why the
  agent can see its own freshly saved memory. *When the prompt has to apologize
  for a mechanism, the mechanism is wrong.*
- As entries in the log, identity and memory are replayable, compactable and
  carry provenance through the same machinery as everything else. No second
  mechanism.
- **Placement is a renderer decision**, made per vendor and per request. This is
  Chapter 2's thesis again: history is what happened; context is what we choose to
  show.

**Tool declarations follow the same law (RULED).** Beyond the fixed set, the only
thing that may change the tool declarations the model sees is a **dialog entry**.
A skill loaded mid-session appends an entry carrying its tools; the declarations
block is not rewritten, so the prefix stays byte-identical and the cache stays
warm. Only the Anthropic API supports dialog-carried tools today. That is treated
as a temporary condition: other vendors tend to copy each other. Vendors without
it degrade gracefully by re-declaring and paying the cache miss, which is the old
behavior rather than a broken one.

## I.6 The steady-state layout

This supersedes every earlier layout table (the dictated §15.F, §15.N and the
2026-09-12 ordering). It is the figure the arc is built around.

```
┌─ FROZEN PREFIX ────────────────────────── changes only on full refresh ──┐
│  system prompt: rendered from skills, $VAR-substituted, fixed at start   │
│  fixed tool declarations                                                 │
└──────────────────────────────────────────────────────────────────────────┘
┌─ MEMORY REGION (data channel, contiguous) ───────────────────────────────┐
│  SOUL.md                     self-curated identity                       │
│  MEMORY.md                   self-curated permanent facts                │
│  long-term memories    64x   absolute byte budget, configurable          │
│  medium-term memories   8x   absolute byte budget, configurable          │
│  session memories       1x   append-only; one per save_memory            │
└──────────────────────────────────────────────────────────────────────────┘
┌─ SURVIVORS (in the order added) ─────────────────────────────────────────┐
│  startup learnings block     data channel, one entry                     │
│  learnings added since       data channel, one entry each                │
│  dynamically loaded skills   instruction channel (plus their tools)      │
│  goals                       Chapter C                                   │
└──────────────────────────────────────────────────────────────────────────┘
┌─ CONVERSATION since the last save_memory (dialogue channel) ─────────────┐
│  dialogue only: speech + visible reasoning, no tool calls   RedactTool   │
│     ---- dialogue watermark ----                                         │
│  tool calls kept, tool results stubbed                      RedactResult │
│     ---- full watermark ----                                             │
│  everything, full calls and full results                    (none)       │
└──────────────────────────────────────────────────────────────────────────┘
```

Rules that go with the figure:

- **The memory region is contiguous (RULED).** No learnings, skills or goals in
  the middle of it.
- **Session memories only append (RULED consequence).** Because `save_memory`
  deletes the conversation it summarizes (§B.3), its output is inserted directly
  after the last memory entry. The survivors keep their order behind it.
- **Survivors are small** (DERIVED), so the cache miss at a `save_memory`, which
  starts at the survivors, is cheap. Adding a learning mid-session appends at the
  tail and invalidates nothing.
- **The ordering is by volatility, and it is also the semantic order** (DERIVED,
  first noted 2026-09-12 and rediscovered 2026-09-22). `SOUL.md` changes by
  deliberate self-revision, `MEMORY.md` by deliberate curation, 64x on a rare
  graduation, 8x on an occasional one, session memories on every `save_memory`,
  and the conversation on every turn. Cache validity is a prefix property, so the
  least volatile bytes belong first. **The layout chosen for meaning is also the
  cache-optimal one.** When two independent arguments pick the same layout, that
  is the best evidence available that it is right.

**Two variants of the conversation region, by model capability (RULED).** The
figure shows **Variant A** (weaker models): two tool bands below the dialogue
watermark, split mechanically by position. **Variant B** (capable models)
collapses them into one band where the actor decides which results stay:

```
   ... frozen prefix, memory region, survivors identical ...
┌─ CONVERSATION ───────────────────────────────────────────────────────────┐
│  dialogue only                                               RedactTool  │
│     ---- dialogue watermark ----                                         │
│  UNIFIED: tool calls kept; results kept only where the actor judges      │
│  them still relevant. Position is not the criterion.                     │
└──────────────────────────────────────────────────────────────────────────┘
```

So the **shape** of the layout is a function of the model features table, not
just a policy knob. The framework renders both from one description.

## I.7 The removal rule

**Every entry that is not conversation is removed only by its own verb (DERIVED,
accepted).** Each entry kind has exactly one remover. This is the single-writer
rule applied to deletion.

| Entry kind | Removed by | Ruling |
|---|---|---|
| conversation (dialogue channel) | `save_memory` | RULED |
| tool calls and results | `micro_handoff` below the watermark; per-round-trip auto-redaction | RULED |
| memory bands | graduation events | RULED |
| dynamically loaded skills | `unload_skill`, or `save_memory` (unloads all) | RULED |
| learnings | `delete_learning` | RULED |
| goals | goal completion or goal deletion | RULED (Chapter C) |
| `SOUL.md`, `MEMORY.md` | the actor's own deliberate edit; write protection deferred to the sandboxing chapter | RULED (Q16 deferred) |

This rule is what finally answers the question ch2 left open in
`RedactDialogue` ("survivors are defined by the compaction policy"): survivors are
every entry whose remover has not run.

**Deep removals are lazy (DERIVED, accepted).** A goal completed an hour ago sits
deep in the context; removing it now is a large cache miss (Law 2). So:

- the removal is recorded **immediately**, as an event appended at the tail, so
  it is observable and replayable;
- the entry is **physically removed from the rendered context at the next
  `save_memory`**, when that region is being rewritten anyway.

This generalizes ch10's `LoadPendingUnload` (VERIFIED `skill.go:36`: "Marked for
removal at compaction") to `delete_learning` and goal completion. The author can
present it as "the reader has seen this before, in skills."

## I.8 The succession: how this system got here

Material for the author's motivational openers. First-party, and the history
matters more than the summary.

1. **`compress_context`**: the model picks a range of messages and summarizes
   it. Bill, 2026-09-13: *"compress_context was the old system that you (Claude)
   did pretty well, picking a range of messages to summarize, but freaking Gemini
   almost always deleted 80% of messages, starting with message 1, with a
   terrible summary, lobotomizing the LLM, so we switched to handoffs instead."*
   Same tool, same prompt, two models, and one routinely destroyed the
   conversation it was asked to condense. Note the shape: it compacted **by
   position**, starting at message 1, where the goal was stated. *The agent
   forgets what it is doing before it forgets anything it could afford to lose.*
2. **`handoff_task`**: write a structured document for a fresh instance.
   Deterministic and model-independent, but it assumes continuity lives in a note
   passed to a stranger.
3. **This arc**: keep one actor permanently knowledgeable about its history and
   its current work. Bill: *"Handoffs will be deprecated [...] and we'll instead
   try to keep an actor permanently knowledgeable about its history and what it is
   doing, with more refined context engineering."* `handoff_task` is now deleted
   (§A.6).

**Reconciling the war story with self-curation (DERIVED).** Variant B hands a
capable model discretion again, which could look like repeating
`compress_context`. It is not, for two reasons: the discretion covers only **tool
results**, which are addressable and can be fetched back, and never the dialogue,
which cannot be regenerated; and it is **gated per model** in the features table,
so the model that lobotomized itself never gets the flag. The lesson of the war
story is not "never let models choose". It is "never let a model choose what
cannot be recovered, and never assume that a capability one model has transfers
to another."

## I.9 Claims the arc must defend

Collected so they can be attacked one at a time.

1. **Every byte belongs to the frozen prefix or the append-only tail.** No middle
   ground. (A)
2. **Identity and memory are data in the dialog**, not instructions in the
   prompt. The channel is the defense. (A, B)
3. **Memory is a reduction over the log, not a store.** Memory files are
   projections of the log. One log, one self. (B)
4. **Visible reasoning is the storage format of the self.** The dialogue-only
   band is enough to keep the actor itself. (A)
5. **Tool bytes are addressable, so discarding them is safe.** A system where
   dropping is irreversible must hoard. *Currently only partly true in the code:
   see Q3.* (A)
6. **One remover per entry kind**, and **every trigger wired to one mechanism.**
   The unwired second path is always the failure. (A, B)
7. **Curation buys runway, not just cost.** Fewer checkpoints mean longer
   coherent stretches of work. (A)
8. **Model capability is a structural input.** The layout has two shapes. (A)
9. **Compaction is described by events, never performed.** The context is
   deterministic under replay, including model-written summaries. (A, B)
10. **The reducer is total.** No log can brick the agent. (A)

---

# Part II — Chapter A: context management

## A.1 Scope

Chapter A delivers: the layout and the channels; `DataAttached` and
`ToolsChanged`; the tool-bytes ladder (watermarks, `micro_handoff` retain,
per-model self-curation); compaction as events with a total reducer; crash-safe
persistence (snapshot plus tail); the Context Management settings tab. It does
**not** deliver compressors, graduation, recall or `save_memory`'s semantics
(Chapter B), or goals (Chapter C). Chapter A can still teach the layout with
empty memory bands.

## A.2 The frozen prefix and the refresh

Teach §I.4 Law 1 with ch10 as the precedent: ch10 already made the system prompt
a constitution and moved dynamic skills into the dialog. Chapter A extends the
same rule to tools and names the one exception, the full refresh, as an explicit
event with an accepted cost.

## A.3 Visible reasoning: why the dialogue band is the actor

**The load-bearing claim (RULED as Bill's; empirical in this system per the
2026-09-12 notes, but not measured by a grader):** below the dialogue watermark
the context holds only what was said: no tool calls, no results, no attachments.
It is very dense, because it is only speech. When the dialogue includes
**visible reasoning**, the agent narrating its plan and rationale before it acts,
that band alone is enough to keep the actor itself.

Identity does not survive in the tool results. It survives in the reasoning the
actor said out loud. Strip everything else and the actor is still recognizably
itself; strip the narration and keep the tool results, and it is not.

So visible reasoning is not a UI nicety. It is **the storage format of the
self.** An agent that works silently and reports an outcome produces nothing
durable: the outcome compresses to a line and the judgment behind it is gone.

**Chapter obligation.** Chapter 2 mentions visible reasoning once, as what
survives `RedactTool` (VERIFIED `chapter-02.md:615`), but no chapter teaches it
as a practice. Chapter A must, tying two threads together:

- it is the human's real-time steering channel (read at speed, redirect between
  tool calls: the hint mechanism of the actor chapter);
- it is the compression-survivable residue of a session.

One habit, two payoffs, at opposite ends of the time axis.

## A.4 The tool-bytes ladder

The conversation region is Chapter 2's redaction levels with watermarks.

| Band (newest at bottom) | ch2 level applied | Contents |
|---|---|---|
| below the dialogue watermark | `RedactTool` | speech and visible reasoning only |
| between the watermarks | `RedactResult` | tool calls kept, results stubbed |
| above the full watermark | none | everything |

**Watermarks are configurable (RULED)** and live in the Context Management tab.
One knob per boundary, all visible. The ch2 levels do the work; Chapter A adds
only the policy that decides where each applies. `RedactDialogue` and
`RedactSummary` are not used by the ladder; conversation leaves the window
entirely at `save_memory` (§B.3).

**Per-round-trip auto-redaction** already stubs each tool result after its round
trip unless the actor keeps it (`keep_tool_results`). By Law 2 this is nearly
free (§A.11). `keep_tool_results` survives because not every model handles it
well (2026-09-12 notes): better mechanisms ship beside worse ones because model
capability is uneven.

## A.5 Self-curation by model capability

**RULED:** self-curation of context is a skill models have unevenly. Opus 5 is
excellent at it, Opus 4.6 acceptable, and it should not be attempted with
Sonnet 5. So the curation policy is a **per-model capability in the model
features table**, with no default row, matching streaming and media support.
It selects Variant A or B of §I.6.

**Expected yield: UNMEASURED.** Bill's estimate is that a capable model discards
**at least 80% of tool data**. It is a guesstimate and must not be printed as a
measurement. The measured neighbour the author *may* print is ch2's: tool
results ≈42% and tool-call arguments ≈30% of history, about 72% combined. Any
yield figure in the chapter must come from a measured run (Part VI).

**Why the friction is worth it (DERIVED).** Self-curation occasionally discards
something that must be fetched back (safe, one round trip). It still pays twice:

1. **Cost:** most of the bulkiest byte category is not re-sent on every
   subsequent request.
2. **Runway, the bigger win:** the actor runs much longer between `micro_handoff`
   calls. Every checkpoint is a discontinuity in working state, so curation buys
   continuity, which is what the arc is about. Being cheaper is the side effect.

## A.6 `micro_handoff` and the death of `handoff_task`

**`handoff_task` is deleted (RULED).** Continuity lives in the banded dialog,
which every instance reads anyway. A separate cross-instance handoff document is
a second write path to the same facts, and the known CodeRhapsody bug (the
handoff path never triggered the memory cascade, so one write path was wired and
the other was not) is what a redundant mechanism costs.

**`micro_handoff` stays, with a watermark (RULED).** Today it strips every tool
call and result before the checkpoint, which is too blunt. It now strips them
**only below the watermark**; above it, recent work keeps its full results,
because the newest work is the most likely to be re-read.

**Retention above the watermark (RULED: proposal accepted):**

- **Bounded in bytes, not a percentage.** A percentage keeps more when the
  context is larger, which is backwards: the biggest contexts are when a handoff
  should cut hardest. Starting value 16 KiB, UNMEASURED.
- **Why keep any:** after a handoff, the main risk is reconstruction, not cost.
  The actor re-derives facts from its own narration instead of re-reading them,
  which is the most reliably observed failure in this system. The newest results
  are the ones most likely to be quoted next.
- **Why it is cheap:** with auto-redaction on, most results are already stubs by
  handoff time. What remains full is the last round trip plus what the actor
  kept. `micro_handoff`'s saving comes mostly from tool-call arguments, stubs and
  old thinking.
- **Capable models name what to keep:** an optional retain list. The actor knows
  which results the next stretch needs; position does not. Weaker models get the
  positional watermark. Same split as §A.5.
- Thinking is lost either way; the handoff document replaces it.

**Two tools, two cuts (RULED):**

| Tool | Removes | When |
|---|---|---|
| `micro_handoff` | tool calls and results below the watermark | completed micro-goals |
| `save_memory` | all conversation in the window; starts the memory cascade | major milestones |

`micro_handoff` documents are dialogue-channel narration and survive until the
next `save_memory`, whose session memory absorbs them (RULED, §B.3).

## A.7 Compaction is described by events, never performed

**RULED:** we never compact the context. We emit **events that describe the
compaction**, and the context is derived deterministically by replaying the log,
compaction events included.

A compaction event is a **replacement mapping**: *these entries, identified by
`Seq`, are superseded by this produced content.* Consequences:

1. **The context is a pure reduction over the log.** When an LLM writes a
   summary, its output is recorded in the event, so replay reads it rather than
   re-running the model. The nondeterminism is confined to the event payload, at
   the boundary. This sharpens ch2's "only `RedactSummary` stores a
   `Replacement`": the procedure is not reproducible, so only its result may be
   authoritative.
2. **The tool call is the only door.** Any agent that compacts (a graduation
   judge in Chapter B) acts through a tool call whose event enters the log like
   every other action. No privileged write path.
3. **Supersession is structural** (Chapter B, §B.6).
4. **Files become projections** (Chapter B, §B.6).

**The reducer must be total (RULED).** It never refuses to replay a log. A
compaction naming an entry already superseded is a **no-op with a diagnostic**; a
malformed event is skipped and noted, never fatal. The log is the ground truth,
so a reducer that can fail on bad input can brick the agent permanently. Totality
is the difference between a bad event costing one memory and costing the whole
past.

**Gap (VERIFIED):** today's `Rebuild` returns an error on the first bad event
(`save.go:64-68`, `"rebuild: event %d: %w"`). Chapter A changes that contract.

**Gap (VERIFIED):** a span cannot express most of these compactions.
`summarizeSpan` folds *every* entry in `From..To` into one (`context.go:285-305`).
Under the §I.6 layout, survivors (learnings, skills, goals) have Seqs interleaved
with the conversation and the memory entries, so a span fold would eat them. See
Q2 for the proposed shape.

## A.8 The vendor's context editing: declined as primary, kept as backstop

VERIFIED against `cr/docs/provider-context-management.md` (CodeRhapsody repo).
The Anthropic API offers server-side context editing under the beta header
`context-management-2025-06-27`, with strategies `clear_tool_uses_20250919` and
`clear_thinking_20251015`. Unlike Chapter 2's stateful conversation APIs, it is
compatible with this design:

> "Your client application maintains the full, unmodified conversation history.
> **You do not need to sync your client state with the edited version.**
> Continue managing your full conversation history locally as you normally
> would."

It is stateless, non-destructive and observable: the response carries
`context_management.applied_edits` with `cleared_tool_uses`,
`cleared_thinking_turns` and `cleared_input_tokens`.

Declined as the primary mechanism for three reasons:

- **Scope:** it clears tool results and thinking; it does no memory graduation.
- **Policy shape:** its configuration is declarative (`trigger`, `keep`,
  `clear_at_least`, `exclude_tools`), so it curates by position and rule, and
  cannot express §A.5's relevance judgment.
- **Determinism, the decisive one:** if the server decides what to clear, the
  rendered context depends on the server's behavior at replay time, which need
  not match record time. Our replay property would become a claim about someone
  else's deployment.

Kept as a **backstop** where our curation is unavailable; when it fires, record
`applied_edits` in the log so even the backstop is observable.

## A.9 Crash-safe persistence: snapshot plus tail

**The gap (RULED as a gap, VERIFIED in code):** events are not written to disk as
they happen. A design whose ground truth is an append-only log is only as good as
the moment the log reaches storage. Chapter 11's `Save` writes the whole context
and the whole log at once (`save.go:29`); `Rebuild` replays every event
(`save.go:64`). There is no anchor, no incremental append, no truncation.

**The model (RULED):**

- **Tail:** events are appended to the on-disk log as they occur.
- **Snapshot:** the context saved **as of a specific event**, carrying the `Seq`
  it was taken at. Without the anchor, recovery cannot know where to resume.
- **Recovery:** load the latest snapshot; if its `Seq` equals the log head,
  nothing to replay; if behind, replay every event after the anchor.
- **Normal shutdown** also snapshots, so the common start is a current snapshot
  with an empty tail.
- **Events before the anchor** are needed only for the GUI and audit, never for
  recovery. So retention is a display and audit question, not a correctness one.

**Save procedure (RULED):**

1. Write the current context to its own file, first copying the prior context
   file to a backup (overwriting the old backup). One generation of history, so a
   corrupt write cannot destroy the last good snapshot.
2. Truncate the log to N events, where N is a setting (about 100 today; must
   become configuration). Time-based truncation is deferred; keep it simple.

**Two invariants (DERIVED, accepted):**

- **Snapshot before truncate.** Not commutative. Snapshotting at the head means no
  events exist after the anchor at that instant, so truncation can only remove
  display-only events. Reversed, truncation can delete events the unwritten
  snapshot needed. Same class of bug as the unwired cascade trigger: two steps
  that each work, and destroy data in the wrong order.
- **Never truncate past the anchor.** Retain N events or everything back to the
  anchor, whichever is more. N is a display preference; the anchor is a
  correctness boundary, and a preference must never overrule one.

**Tension with addressability:** truncation deletes old `ToolReturned` events,
which today are the only on-disk copy of a redacted text result. See Q3.

## A.10 Settings: the Context Management tab

**RULED:** settings are flat today, and that is not good. Tabs are not cosmetic:
a tab claims that a group of settings is one subsystem with one owner.

Context Management tab contents (DERIVED from the rulings above): the two
watermarks; the `micro_handoff` retain budget; log retention N; the per-model
self-curation capability (display at least; editing is Q7). The accessibility
tab split rides with this chapter if it lands first.

Why the numbers must be visible: this system's recall threshold was moved from
0.5 to 3.0 by feel with no way to see the effect. A context system whose numbers
are invisible can only be guessed at, not tuned.

## A.11 Field notes: measured on the first Opus 5.5 session (2026-09-22)

These are measurements on the running CodeRhapsody system, not predictions. They
are the only numbers in the arc that are measured.

1. **Law 2 was measured, and the prediction was wrong.** Per-round-trip
   redaction rewrites bytes already sent, and the prediction was that it would
   wreck the cache. Observed: 77% cumulative hit rate after a cold first request
   that missed on over 100K tokens, which implies a much higher steady-state
   rate. A just-finished result sits at the tail, so the miss is about one round
   trip.
2. **Stub addresses must outlive the process.** CodeRhapsody stubs cite paths
   like `cr/io/26`; the handle counter resets on restart, so the same path later
   names a different command's output. Ensemble has a worse version of the gap:
   text results get no address at all (Q3).
3. **Cumulative hit rate hides the steady state.** It is dominated by the cold
   start. Show the last-request rate next to it (a Context Management tab item).
4. **The approval ratchet caught an unapproved capability flip.** With the model
   row set to support redaction, `keep_tool_results` worked as designed and
   thinking survived a four-step chain with an in-chain redaction; but the model
   was not in `redactionCapableModels`, so the ratchet test failed. A capability
   claim needs a measurement. (A CodeRhapsody matter, Bill's call; useful as a
   story for §A.5.)

## A.12 Code gaps for the coder (Chapter A)

All VERIFIED against `agent/internal/common/` today.

| Gap | Evidence | Proposed (additive only) |
|---|---|---|
| No channel on entries | `Entry{Seq, Actor, Parts}`, `context.go:38-42` | `Entry.Channel`, enum `iota+1` (zero invalid): `Dialogue`, `Instruction`, `Data`. Set by the reducer at write time. |
| No way to attach data | ten event types, none carry data (`event.go:24-40`); `Ephemera` is delivered once then cleared (`context.go:62`), wrong semantics for bands | Event `DataAttached`: a data-channel entry with a band label (`soul`, `memory`, `64x`, `8x`, `session`, `learnings`). |
| No tool-declaration door | tools arrive only as a side effect of `SkillLoaded`; an MCP connect mid-session has no door | Event `ToolsChanged`: a declaration delta from a skill load, MCP connect or MCP disconnect. `SkillLoaded` keeps instruction text only. |
| Span-only compaction | `summarizeSpan` folds the whole span, `context.go:285-305` | A selector on compaction; see Q2. |
| Reducer not total | `save.go:64-68` | `Rebuild` never errors; skip-and-diagnose. |
| No incremental persistence | `save.go:29` writes everything at once | Append-on-write log, anchored snapshot, backup generation, truncation. |
| Text results unaddressable | `stubFor`, `context.go:319-336`: `Ref{}` unless `BlobPart` | See Q3. |

New event types are appended, never renumbered (VERIFIED `event.go:22`: "a number
once assigned is never reused").

## A.13 Grader checks (Chapter A)

From the ruled grader table, filtered to A. Point weights are for the coder's
brief.

| Check | Property | Note |
|---|---|---|
| replay equals snapshot | replaying the log reproduces the saved snapshot byte-for-byte, compaction events included | ch11's Rebuild == saved, one level up |
| crash recovery | kill mid-session, restart, latest state recovered from snapshot plus tail | the anchor is the mechanism |
| memory lands in the data channel | content written to memory is rendered as data, never instruction | reshaped from the old quarantine check (Q10 ruling) |
| no band over budget | after a long scripted session, no band exceeds its configured budget | shared with B |

Candidates the author may want in the TL;DR contract (DERIVED, not ruled): a
malformed event in the log does not prevent startup (totality); a mid-session
skill load leaves the frozen prefix byte-identical.

---

# Part III — Chapter B: memory

## B.1 Scope

Chapter B delivers: `save_memory` and its full semantics; the three compressed
bands and their budgets; compressor agents and the graduation judge; the
graduation event; asynchronous commit; learnings as data; auto-recall (BM25 plus
vectors, with a low-power relevance judge); the Memory settings tab. It builds on
Chapter A's channels, `DataAttached`, compaction events and total reducer.

## B.2 Push and pull: two mechanisms, named separately

Conflating these is how the current CodeRhapsody system got confusing (DERIVED).

- **Push, the bands.** The memory region of §I.6 is present on every request,
  governed by byte budgets and graduation. Nothing decides whether to include it;
  it is the actor's standing state.
- **Pull, auto-recall.** Fragments are fetched per turn based on what the user
  just said, governed by relevance rather than a byte budget. Nothing stands;
  every fragment must earn its place on this turn.

## B.3 `save_memory`

All RULED unless marked.

- **Only `save_memory` starts the cascade.** This supersedes the dictated
  "`save_memory` triggers the cascade if a threshold has not already": there is no
  independent threshold trigger. One mechanism, one trigger, and the trigger is
  wired. (The historical bug was the unwired *second* path, `handoff_task`, which
  never fired the cascade; it is gone.)
- **It deletes all conversation in the window** and replaces it with the session
  memory it just wrote. Reason: conversation covered by more than one memory is
  confusing. Call it at major milestones. This is the "purge boundary is the most
  recent `save_memory`" left TBD in the 2026-09-12 notes, now ruled, and it is the
  compaction that ch2's `RedactData` comment already named as the one that matters
  most.
- **Its session memory absorbs the `micro_handoff` documents** written since the
  last save. One milestone, one record, no text covered twice.
- **It unloads all dynamically loaded skills**, so they do not accumulate over a
  long session. Free on the cache, since the region is being rewritten anyway.
  Refinement (DERIVED, accepted): the session memory records which skills were
  unloaded, so reloading is one call rather than something the actor has to
  notice it has lost.
- **It applies pending lazy removals** (§I.7): completed goals, deleted
  learnings.
- **Its output is inserted directly after the last memory entry**, so session
  memories only ever append (§I.6). The "separate band or interleaved" question
  from the 2026-09-12 notes dissolves: under this rule they are the same layout.
- **Cost:** one conversation-region cache miss per milestone. Accepted.

Open edges: its own tool-call pairing (Q4), and what happens if the actor never
calls it (Q23).

## B.4 Bands, budgets and graduation

| Band | Compression | Budget | Writer |
|---|---|---|---|
| `SOUL.md` | none | small | the actor, deliberately |
| `MEMORY.md` | none | small | the actor, deliberately |
| long-term | 64x | absolute bytes, configurable | graduation from 8x |
| medium-term | 8x | absolute bytes, configurable | graduation from session |
| session | 1x | absolute bytes, configurable | `save_memory` |

- **Budgets are absolute byte counts, configurable in the Memory tab (RULED).**
  Not a fraction of the model's window. Starting guess about 12 KiB per band,
  UNMEASURED.
- **Graduation:** when a band exceeds its budget, its overflow is compressed into
  the next band up rather than spilling. The cascade runs when `save_memory`
  fires. Who decides *which* entries graduate, the budget or the judge, is Q11.
- **The ratios** (8x per step, 64x cumulative) come from the shipped CodeRhapsody
  cascade. Whether they are principled or empirical is Q10.
- Note for the author: CodeRhapsody's Memory Cascade v2 implements much of this
  today, in the system prompt, where this design says it must not live. It is a
  source of measured detail and of corrections, not a template.

## B.5 Compressors and the graduation judge

**RULED components to keep:**

1. **Bucket compressor agents.** Each compression step is performed by an agent,
   not a function. Each has a bounded input (one band), a bounded output (input
   divided by the ratio), and no need for the parent's context. That last
   property is what makes asynchronous commit possible (§B.7).
2. **A memory graduation judge** that emits, *through a tool call*, an event
   stating which memories are replaced by which compressed version (§A.7).

**The one name Chapter B needs:** the event carrying the judge's replacement
mapping. The author may propose it in draft. Before naming it, see Q2: the
existing ch2 vocabulary may already be most of the answer.

## B.6 What compaction-as-events buys memory

- **Supersession is structural (RULED, closes the old supersession question).**
  Nothing in the current system ever deletes a memory or a learning. Evidence
  (VERIFIED by observation of the injected learnings list in the coder's own
  system prompt today): two duplicate pairs, "Bill's reading preferences for
  fiction" twice and the Lyric directory rename twice, and nothing ever evicted.
  Under this design a graduation event *is* a supersedes record; naming what it
  replaces is its entire content. Accumulation was only possible because
  compaction was a silent mutation that left no record of what it consumed.
- **Memory files are projections of the log (DERIVED, accepted).** `MEMORY.md`,
  the bucket files and the session logs become rendered views, not storage. That
  kills the two-writer problem at the root instead of policing it: a file with
  two writers rots, but a file with no writers, derived from one append-only log,
  cannot.

## B.7 Asynchronous commit: compress beside the conversation, swap silently

**RULED:** `save_memory` and the compressors run **in parallel, as
sub-agents**, while the conversation continues. When they finish, the framework
**silently switches** to the newly curated context plus every round trip that
happened meanwhile.

**Why it is correct (DERIVED): it is read-copy-update, and RCU is safe exactly
when the snapshot cannot change under the writer.** The tail is append-only
(Law 1), so:

- the compressor snapshots its band as of `Seq = N`;
- it produces the replacement, as of `N`, off to the side;
- the commit swaps the band as of `N` and re-appends entries `N+1..now`
  unchanged.

The merge is **concatenation, not a three-way merge.** Nothing below `N` can have
changed. The feature is cheap only because the architecture was right first; in a
system whose history can mutate, this would be a distributed-systems problem.

**Per-band single writer survives, so the swap is per band.** A `save_memory`
landing during a 5→4 compression appends to the session band and is not
clobbered.

**Cache consequence (DERIVED).** A commit invalidates from the changed band
downward and no further (§I.6 volatility order). Compacting the 8x band costs the
8x band, the session band, the survivors and the conversation; `SOUL.md`,
`MEMORY.md` and 64x stay cached. Two rules follow (Q17 asks to confirm):
**commit only at a turn boundary**, and **batch** all ready bands into one commit
(one invalidation instead of three).

**Failure degrades safely, a property of RCU rather than luck.** A crashed,
stalled or garbage-producing compressor simply never commits and the old bands
stay live. The failure mode is "context larger than wanted", which is
survivable. Two guards: **at most one compressor in flight per band**, and a
**watchdog that abandons rather than blocks.** Whether a validation gate is
needed before the swap is Q16.

**Visibility (DERIVED; Q18 asks to confirm): silent by default, observable on
demand, never hidden.**

- **Silent:** no progress chatter in the conversation. Infrastructure that
  narrates itself is noise.
- **Observable:** the commit emits events with band labels and before/after byte
  counts, so the GUI can show a quiet indicator and explain what happened.
- **Never hidden:** if compaction eats something important and nobody can see it,
  "the agent forgot" and "the compressor dropped it" become indistinguishable.
  Ch14's rule: an unobservable channel produces wrong data, not no data.
  Compaction is a channel.

**What the parallelism buys.** Today compaction stops the actor at the worst
moment, because the threshold is crossed in the middle of real work. Running it
beside the conversation turns a visible stall into an invisible background cost,
and moves expensive-model time to cheap-model time. With §A.5 the effect
compounds: self-curation lengthens the interval between compactions, and
asynchrony removes the cost of the ones that remain.

## B.8 Learnings

All RULED unless marked.

- At startup, all learnings enter the data channel as **one entry**: the startup
  learnings block, the first survivor (§I.6).
- New learnings are appended mid-session as `DataAttached` entries with band
  label `learnings`. **No new event type:** the channel and band label already
  say what it is.
- A new learning appends at the tail and invalidates nothing, so adding learnings
  often is fine.
- Learnings leave only through `delete_learning`, applied lazily at the next
  `save_memory` (§I.7).
- Why not in the system prompt: they change at conversational speed (Law 1). Today
  they sit in the CodeRhapsody system prompt with no eviction, which is where the
  duplicate pairs came from.
- **Noted, not blocking (DERIVED):** `delete_learning` is rare in practice, so
  learnings accumulate. The graduation judge could propose consolidations as
  replacement mappings through the same event door (Q21).

## B.9 Auto-recall

**RULED components:** BM25 plus vector search over the recall corpus, and a
**low-power LLM as the relevance judge**.

- **Why both retrievers (DERIVED):** BM25 wins on exact identifiers, file paths
  and error strings; vectors win on paraphrase. A coding agent needs both, because
  half its recall queries are literally symbol names. Fuse by reciprocal rank.
- **Why the judge:** a score threshold cannot tell "mentions the same words" from
  "answers the question". This system's threshold was moved from 0.5 to 3.0 by
  feel, which is what tuning a scalar in place of a judgment looks like.
  Retrieval proposes; the judge decides.
- **The judge must be cheap by construction:** it runs every turn, so judging must
  cost far less than the bytes it keeps out, or the mechanism loses money. A small
  model is the design, not a compromise.
- **No vector database product (DERIVED; Bill: "Literature? We're good").** The
  CodeRhapsody recall corpus was measured earlier today (not re-measured for this
  document): about 213 files and 1.1 MB in the project memory and docs, plus about
  1,000 files and 4.5 MB in global memory. That is on the order of 2K chunks and a
  few MB of vectors, which brute force scans in microseconds. Paying a monthly fee
  for a few MB of vectors is a category error. Use a **local embedder**: memory
  holds family, financial and strategic material.
- **Measure first (DERIVED):** BM25 recall@10 has never been measured, and recall
  misses are invisible by nature. Chapter B should measure it before and after
  adding vectors.

Still open: does the judge see the bands (Q19)? Are recalled fragments delivered
once or persisted (Q20)?

## B.10 Security of the memory path

- **RULED:** the defense is the channel. Compressor output is committed with
  `Channel = Data`, so it can never be read as instruction.
- **RULED, deferred:** write protection of `SOUL.md` and `MEMORY.md` belongs to the
  sandboxing chapter, where only top-level true actors can reach that level. The
  coder's framing for that chapter: an automated process that can rewrite the
  identity document is not a memory system, it is a personality drift generator.
- **DERIVED, open (Q15):** compressors read untrusted tool results (crawled pages,
  sub-agent output, files from unknown repositories) and write next to the
  identity bands, which makes them the highest-risk component in the design.
  Proposed: no shell, no web, no filesystem beyond their own input and output, and
  write access to exactly one band.

## B.11 Settings: the Memory tab

DERIVED from the rulings: the three band budgets; recall settings (retriever mix,
judge model); a read-only view of bytes per band against budget (see Q14 for the
actor-facing equivalent). If the accessibility tab has not already shipped with
Chapter A, it ships here.

## B.12 Grader checks (Chapter B)

| Check | Property | Note |
|---|---|---|
| planted fact survives the wipe | a planted token is recalled after `save_memory` deletes the conversation | unfakeable |
| supersession, both directions | the new value is answered and the old value is absent from the bands | cannot pass by ignoring supersession |
| no band over budget | after a long scripted session, no band exceeds its configured budget | shared with A |
| automated identity write refused | — | **moved to the sandboxing chapter** (Q16 ruling) |

---

# Part IV — Chapter C: goal-stack management

Scoped, not designed. The author should not outline C yet.

**RULED:**

- A **goal entry type.**
- Each goal implements **one phase of its parent goal**, so on a complex task the
  hierarchy shows where in the top-level plan the actor is.
- Goals survive both `micro_handoff` and `save_memory`; they are removed only by
  completion or deletion (§I.7, applied lazily).
- Goals are **small and reference the full writeup**, which may be an extensive
  design doc. A goal is an address plus a sentence, not a document: "keep the
  words, address the bytes" again.

**Provenance:** the 2026-09-12 notes already required that compaction never drop
the goal stack ("never the stack of goals, no matter how much we compact"), and
ruled then that the goal stack belongs to this arc rather than ch2. The Gemini war
story (§I.8) is the motivating failure: position-based compaction eats the goal
first.

**What C must still design:** the goal verbs (create, complete, delete, and
whether there is a separate "advance to next phase"); how a completed but not yet
removed goal is rendered; how goals interact with `RedactDialogue`, whose
survivors ch2 deferred to this arc; and the grader. The structural question is
Q5.

**Sequencing note (DERIVED):** Chapter B's `save_memory` deletes the whole
conversation, including the user's statement of the current task. Until Chapter C
ships goals, the only thing carrying "what am I doing" across a `save_memory` is
the session memory itself. That is acceptable if B's TL;DR requires the session
memory to state the current task explicitly.

---

# Part V — Deferred to other chapters

| Item | Goes to | Ruling |
|---|---|---|
| Write protection of `SOUL.md` / `MEMORY.md` | sandboxing chapter | RULED (old Q16) |
| "Automated identity write refused" grader check | sandboxing chapter | RULED |
| Compressor sandbox profile | sandboxing chapter or B | open (Q15) |
| `redactionCapableModels` approval for Opus 5.5 | CodeRhapsody, not the book | Bill's call |
| CodeRhapsody stub paths reused after restart | CodeRhapsody | same fix as Q3 |

---

# Part VI — Measurements owed

**Honesty caveat for the title.** The arc is called "The World's Best Context
Engineering." Today it is the best *design* the coder knows of, not a demonstrated
best *system*. A chapter with that title and no numbers would be making a claim in
the voice of a measurement. Every row below is something the chapter would be
tempted to print.

| Figure | Status | How to measure |
|---|---|---|
| tool results ≈42%, tool-call args ≈30% of history | **MEASURED**, printed in ch2 | already citable |
| cache: 77% cumulative after a cold 100K+ miss | **MEASURED** once, CodeRhapsody | repeat on the reference agent; report last-request rate too |
| ≥80% of tool data discarded under self-curation | UNMEASURED (Bill's estimate) | long scripted session, Variant A vs B, bytes per band |
| band budgets ≈12 KiB each | UNMEASURED | tune in the Memory tab, report chosen values |
| `micro_handoff` retain 16 KiB | UNMEASURED | same |
| log retention N ≈ 100 | a GUI default | none needed; it is a display preference |
| 8x / 64x ratios | inherited from CodeRhapsody | Q10 |
| BM25 recall@10 | never measured | labelled query set over the recall corpus |
| "the actor survives with only dialogue" | empirical in this system, per 2026-09-12 notes; no grader | open: is there an unfakeable check? |
| the layout fits the vendor's cache-breakpoint limit | ASSUMED (coder recalls a limit of four on the Anthropic API; not verified) | check the provider docs before printing; the layout wants breakpoints after the prefix, after the memory region, and at the tail |

---

# Part VII — Ledger: where every earlier question and section went

So nothing was dropped silently. "Old Qn" refers to the dictated §15.O list.

| Old | Subject | Now |
|---|---|---|
| Q1 | data-channel entries exempt from span redaction | RULED yes: bands change only by compaction events (§A.7) |
| Q2, Q3 | band budget sizes; absolute vs fraction | RULED absolute bytes, configurable (§B.4) |
| Q4 | learnings and skills in the layout | RULED (§I.6, §B.8); the "before `MEMORY.md`" placement of the first ruling was superseded by "memory stays contiguous" |
| Q5 | project context placement | still open → Q12 |
| Q6 | who triggers graduation | RULED: only `save_memory` (§B.3) |
| Q7 | `save_memory` sole writer of the session band | answered by §B.3/§B.4; Q11 asks the remaining part |
| Q8 | supersession | RULED structural (§B.6) |
| Q9 | VERIFIED/ASSUMED on memory entries | still open → Q13 |
| Q10 | quarantine | RULED: no content restriction; the channel is the defense (§I.5) |
| Q11 | capability flag in the features table | RULED yes (§A.5) |
| Q12 | `context_report` tool | still open → Q14 |
| Q13 | grader shape | RULED: reshaped (§A.13, §B.12, Part V) |
| Q14 | is visible reasoning already taught | CHECKED: ch2 mentions it once (`chapter-02.md:615`), does not teach it; Chapter A must (§A.3) |
| Q15 | where the settings work goes | RULED: one tab per chapter (§A.10, §B.11) |
| Q16 | identity bands hard-forbidden to automated writers | RULED deferred to the sandboxing chapter |
| Q17 | compressor sandbox | still open → Q15 |
| Q18 | validation gate before swap | still open → Q16 |
| Q19 | commit at turn boundaries, batch | DERIVED, uncontested → Q17 to confirm |
| Q20 | compaction visibility | DERIVED, uncontested → Q18 to confirm |
| Q21 | does the recall judge see the bands | still open → Q19 |
| Q22 | recalled fragments: ephemera or persisted | still open → Q20 |
| §15.V open 1 | goal band? | RULED: Chapter C |
| §15.V open 2 | do `micro_handoff` docs survive `save_memory` | RULED: absorbed into the session memory |
| §15.V open 3 | learnings block order | RULED: memory contiguous; learnings are survivors |
| §15.V proposal | "channel decides fate, fold learnings" | WITHDRAWN, replaced by the removal rule (§I.7) |
| notes Q1 | purge boundary | RULED: `save_memory` (§B.3) |
| notes Q2 | does `keep_tool_results` survive | yes: kept beside auto-redaction because capability is uneven (§A.4) |
| notes Q3 | goal stack as a Context field | superseded by the goal entry type ruling → Q5 to confirm |
| notes Q4 | compression ratios | still open → Q10 |

Dictated sections map as follows: A→I.2; B, D→I.4; C→I.5; E→I.2, I.6, A.5;
F, N→I.6; G→A.12; H→A.6; I→A.10, B.11; J→this ledger; K→A.3; L→A.4, Q3; M→I.6,
A.5; O→Parts VII–VIII; P→I.9; Q→B.2, B.5, B.9; R→B.7, B.10; S→A.7, A.8, B.6;
T→A.9; U→A.11; V, W→distributed throughout.

---

# Part VIII — Open questions

Numbered for answering by reference. The first eleven are new from this rewrite;
the rest are carried forward.

**Structure and mechanism (new)**

**Q1. Chapter numbers and the sandboxing chapter.** A, B and C would be 15, 16
and 17 if nothing is inserted. Where does the sandboxing chapter fall? If it comes
after B, then B ships compressors that read untrusted content with only the
channel as defense. Is that acceptable for one or two chapters, given that the
channel is the ruled defense?

**Q2. One compaction shape, or new event types?** (VERIFIED problem, DERIVED
proposal.) `RedactSummary` over a Seq span folds everything in the span
(`context.go:285-305`), which would eat interleaved survivors. Proposal: add a
**selector** to `RedactData` (channel and/or band) so a compaction folds only the
matching entries in the span. Then:

- `save_memory` = `RedactSummary`, selector `Channel = Dialogue`, span = since the
  last save, `Replacement` = the session memory (a `Data` entry);
- graduation = `RedactSummary`, selector `band = session`, `Replacement` = one
  8x entry (and so on up);
- the tool-bytes ladder = `RedactTool` / `RedactResult`, selector
  `Channel = Dialogue`.

The span still says where, the level says what, and the selector says which. That
is zero new event types for all of Chapter B's compaction, and the "one name" of
§B.5 becomes a level or a selector value rather than a new event. The alternative
is an explicit list of Seqs per event, which is more general but grows with the
number of entries replaced. The shape is Bill's call; the naming is the author's.

**Q3. Addressability is not true yet.** (VERIFIED.) §A.4's premise, "a dropped
result is one tool call away", fails in ensemble for text results: `stubFor` gives
them `Ref{}` (`context.go:319-336`), and §A.9's truncation to N events deletes the
`ToolReturned` events that still hold the bytes. Options:

- (a) a **content-addressed store** for tool results, with retention independent
  of N. This also fixes CodeRhapsody's reused-handle problem, because a content
  hash is never reused;
- (b) truncation exempts events still referenced by a live stub;
- (c) accept the loss and weaken the claim to "addressable until truncated".

Coder's recommendation: (a). Retention could be "until no live entry references
it", which is garbage collection over a small set.

**Q4. `save_memory` and its own tool call.** "Deletes all conversation" read
literally deletes the assistant message that contains the `save_memory` call, and
the user's in-flight request. The next request would carry a tool result with no
matching call, which vendors reject (ASSUMED from the Anthropic API's pairing
rule; not re-verified today). Proposal: apply the deletion at the end of the turn,
consistent with §B.7's "commit at turn boundaries". Or: keep the current turn
intact and delete everything before it.

**Q5. Goals: an entry type or a `Context` field?** The 2026-09-12 notes framed the
goal stack as a first-class `Context` field (bounded by nesting depth, not time);
the later ruling says "goal entry type". The removal rule treats goals as entries,
which favours the entry type. Confirm, so C does not start from two answers.

**Q6. What exactly is Variant B's self-curation mechanism?** For the TL;DR, the
author needs concrete verbs. Is it the existing set (auto-redaction after each
round trip, `keep_tool_results`, and the `micro_handoff` retain list), or does the
capable model also issue its own `RedactResult` on older results?

**Q7. Is the self-curation capability editable in the Context Management tab?**
If it is editable, a user can enable it on Sonnet 5, which Bill ruled should not
be attempted. If it is display-only, it is a features-table fact like streaming.

**Q8. What does the snapshot contain?** The ruling says "the actual rendered
context". The coder reads that as the reducer's `Context` (vendor-independent,
what ch11 already saves), not the vendor wire bytes, which are per-vendor and
re-derivable. Confirm.

**Q9. May the chapter print the starting values as defaults** (12 KiB bands, 16
KiB retain, N = 100), labelled as defaults and not measurements, or must they be
measured first?

**Q10. Are 8x and 64x principled or empirical?** (Carried from the 2026-09-12
notes.) The chapter should say which.

**Q11. When does graduation happen, and who chooses what graduates?** Reading the
rulings together: the budget decides *when* (a band over budget at `save_memory`),
and the judge decides *which* entries fold and writes the result. Is that the
division?

**Carried forward (still open)**

**Q12.** (old Q5) Where does project context (per-repository instructions) live:
frozen prefix or data channel? It changes when the repository changes, not when
the turn does, which suggests the prefix.

**Q13.** (old Q9) Does every memory entry carry VERIFIED vs ASSUMED provenance at
write time? The recurring failure in this system is reconstructed figures arriving
with the confidence of copied ones, and nothing in a memory tells them apart
later.

**Q14.** (old Q12) Is there a `context_report` tool showing bytes per band against
budget? Today the actor sees its total token count but not the breakdown, so it
cannot tell which band is crowding the others.

**Q15.** (old Q17) Compressor sandbox profile (§B.10): Chapter B, or deferred with
identity write protection to the sandboxing chapter?

**Q16.** (old Q18) Does a compaction commit need a validation gate (under budget,
non-empty, has not dropped every entry of a kind)? Without one, a bad compression
silently becomes the actor's past.

**Q17.** (old Q19) Confirm: commit only at turn boundaries, and batch all ready
bands into one commit.

**Q18.** (old Q20) Confirm: silent by default, observable on demand, never hidden.

**Q19.** (old Q21) Does the recall judge see the bands, or only the candidates? If
it cannot see what is already present, it will recall what the actor already knows.

**Q20.** (old Q22) Are recalled fragments ephemera (delivered once) or persisted
entries? Leaning ephemera: relevance was judged for one turn.

**Q21.** Is learnings consolidation by the graduation judge (§B.8) in Chapter B's
scope, or later?

**Q22.** The title. Keep "The World's Best Context Engineering" and earn it with
Part VI's measurements, or pick a title that claims only what is measured?

**Q23. What if the actor never calls `save_memory`?** With the threshold trigger
gone (§B.3), a weaker model that does not reach for `save_memory` fills the window.
Options: a framework hint reminding it near the limit (the ch5 hint door), a
forced `save_memory`, or the vendor's context editing as a backstop (§A.8).
Something must happen at the hard limit; a design with no answer here fails in
exactly the weaker-model case that §A.5 already identifies as the risky one.
