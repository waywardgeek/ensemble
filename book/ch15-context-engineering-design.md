# Chapter 15 — The World's Best Context Engineering

*Design record. Dictated by Bill, transcribed with commentary by the coder.
Status: IN PROGRESS, live dictation. Nothing here is prose yet.*

**Working title is the real title.** "The World's Best Context Engineering."

---

## 15.A What the chapter is about

**Definition (Bill's).** Context engineering means *managing all the bytes in
our context data structure*. Getting that right, as well as it can be done,
given the very real limitations of today's LLMs.

Not memory as a feature. Not retrieval as a feature. The single question is:
for every byte that occupies the window on this request, why is it there, who
put it there, and what evicts it.

This is also the chapter where the agent's long-term identity gets designed
deliberately instead of accreting. The pieces exist today — SOUL.md, MEMORY.md,
the cascade, learnings, skills, `micro_handoff`, `handoff_task`,
`keep_tool_results` — but each was built separately, at a different time, with
no shared theory. The chapter's job is the theory, and then the refactor that
the theory implies.

---

## 15.B Settled: the system prompt

This part is already solved and ships as-is. It is the baseline the rest of the
chapter builds on.

1. **The system prompt is built from skills.** Skills are the unit of
   composition; the prompt is their rendered concatenation.
2. **Variable substitution** is applied at render time (`$VAR`).
3. **It is fixed at agent startup.** Once rendered, it does not change.
4. **The only thing that may change it is a full refresh cycle.** A refresh is
   an explicit, deliberate event — for example, a skill already declared in the
   system prompt has been renamed because a human edited it on disk. Refresh
   re-renders from scratch.
5. **A refresh is a full cache miss, and that is accepted.** It is the price of
   correctness, paid rarely and on purpose. Outside of a refresh, *we never
   touch the system prompt*.

Consequence worth stating plainly: the system prompt is a constitution. It is
not a scratchpad, not a memory store, and not a place to put anything that
changes at conversational speed.

---

## 15.C Ruling: identity and memory are DATA MESSAGES, not system prompt

**SOUL.md and MEMORY.md do not go in the system prompt.**

They enter the conversation as **data messages in the dialog**. The renderer
decides where to place them for a given vendor and a given request.

Why this is the right call, spelled out:

- The system prompt is frozen at startup (15.B). Identity and memory are not
  frozen: a memory is written mid-session, and the agent must be able to see it
  without a cache-busting re-render. Putting them in the prompt would make every
  `save_memory` a full cache miss, or else make the saved memory invisible until
  restart. Today's system splits the difference and renders memory live into the
  prompt, which is exactly why the prompt currently has to contain a paragraph
  explaining that seeing your own freshly-saved memory is normal. *When the
  prompt has to apologize for a mechanism, the mechanism is wrong.*
- As data messages they sit in the append-only log like everything else, which
  means they are replayable, redactable, compactable, and provenanced by the
  same machinery as every other part. No second mechanism.
- Placement becomes a **renderer** decision, per vendor and per request, rather
  than a fact baked into storage. Same thesis as Chapter 2: one log, many
  renderings. History is what happened; context is what we choose to show.

---

## 15.D Ruling: tool declarations obey the same law as the system prompt

**The fixed tool set is loaded at startup and never changes.** It is part of the
frozen prefix, exactly like the rendered system prompt, and for the same reason:
it is the thing the prompt cache is keyed on.

**Beyond that fixed set, the only thing that can affect the tool declarations
the model sees is a dialog entry.** A skill loaded mid-session does not rewrite
the declarations block. It appends a dialog entry that carries the new tools.
The frozen prefix stays byte-identical, the cache stays warm, and the model's
visible toolset still changes.

This is the same move as 15.C applied to a different kind of byte: anything that
changes at conversational speed lives in the dialog, and only the dialog.

**Vendor reality.** Only the Anthropic API supports this today. That is a
temporary condition, not a design constraint. The AI companies show close to
zero creativity — other than Anthropic, and sometimes the OpenAI API — and what
they mostly do is copy each other. Dialog-entry typing of this kind will be
everywhere soon. Design for it now; degrade gracefully where it is missing (a
vendor without dialog-carried tools simply eats the cache miss and re-declares,
which is the old behavior, not a broken one).

**Design principle extracted from 15.B–15.D:**

> The context has a *frozen prefix* and an *append-only tail*. Every byte
> belongs to exactly one of them. A byte in the prefix must never change
> outside an explicit refresh. A byte that needs to change belongs in the tail,
> no exceptions, no clever middle ground.

---

## 15.E What the chapter actually solves

Context engineering solves the problem of **maintaining actor identity in the
most useful way for both the actor and the user.**

Two halves:

**Identity.** A self-curated `SOUL.md` and a self-curated `MEMORY.md`. Those two
documents define the basic actor identity. Self-curated is the operative word:
the actor writes them, and nothing else does.

**Long-term memory at graded compression.** Possibly the more important half.
The actor remembers at four compressions at once:

| Tier | Compression | Contents |
|---|---|---|
| Long-term | **64×** (default in this system) | oldest memories, heavily compressed |
| Medium-term | **8×** | the middle distance |
| Short-term | **1×** (uncompressed) | session summaries — the `memory_<date>-<n>.md` files |
| Current session | 1×, tool calls stripped | the live conversation, so the actor remembers what it is doing |

Then the live dialogue itself splits by age into three bands:

1. **Session conversation, no tool calls at all.** Just the dialogue. This is
   the band that answers "what the hell are we doing" — narration and user
   turns, nothing else.
2. **Somewhat redacted band: tool calls present, tool results gone.** The calls
   are the record of what was attempted; the results are the bulk. Optionally
   curated by the LLM itself, *if the model is good enough at it.*
3. **Current full band.** Everything. The only thing that might be missing is
   tool results, and again only if the model is good enough to curate them.

**Model capability is a real input, not a footnote.** Self-curation of context
is a skill models have unevenly: the Opus 5 model is excellent at it, Opus 4.6
is acceptable, and it should not be attempted with Sonnet 5. So the curation
policy must be a per-model capability flag in the model features table, not a
global constant. Same pattern as streaming and media support: a data table with
no default row.

---

## 15.F The steady-state dialog table

The dialog in steady state, top to bottom. Every band has a byte budget, and
the budget is the mechanism: when a band exceeds it, the overflow **graduates**
into the next band up (compressed) rather than spilling.

| # | Band | Compression | Budget (current) | Writer |
|---|---|---|---|---|
| 1 | `SOUL.md` | — | small, fixed | the actor, deliberately |
| 2 | `MEMORY.md` | — | small, fixed | the actor, deliberately |
| 3 | Long-term memories | 64× | ≤ 12 KiB | graduation from band 4 |
| 4 | Medium-term memories | 8× | ~ same as band 3 | graduation from band 5 |
| 5 | Session memories | 1× | ~ same as band 3 | `save_memory` |
| 6+ | Live dialog entry types | — | the remainder | the log |

Bands 1–5 are all *data messages* in the dialog (15.C). Band 6 onward is where
the interesting dialog entry types live.

---

## 15.G GAP ANALYSIS: do we have the dialog types we need?

Checked against the code, not from memory. Source: `agent/internal/common/`.

**What exists today.** `Event.Type` has ten values: `MessageReceived`,
`RequestSent`, `ResponseStarted`, `ResponseEnded`, `ToolCalled`, `ToolReturned`,
`Redacted`, `ErrorOccurred`, `JobKilled`, `SkillLoaded`. The context is
`{Turn, Dialogue []Entry, Ephemera PartList, Usage}`, and `Entry` is
`{Seq, Actor, Parts}`.

**The central finding: `Entry` has no channel.** An entry carries *who* said it
(`Actor`) and *what* (`Parts`) — and nothing that says *which channel* it
belongs to. So "skill declarations in the instruction channel" and "memories in
the data channel" are not expressible today. There is no channel concept at all.
Everything is flattened into actor-plus-parts, and the renderer has to guess
placement from the actor, which is exactly the inference-at-read-time that
Chapter 2 forbids.

Three answers to the three questions:

1. **Tool declarations — GAP.** The fixed set is rendered from config at
   startup, which is correct per 15.D. But there is no dialog entry type that
   carries a *tool declaration delta*. `SkillLoaded` carries the skill's name
   and rendered body, so tools arrive as a side effect of a skill load and
   cannot arrive any other way. An MCP server connecting mid-session has no
   door.
2. **Skill declarations in the instruction channel — PARTIAL.** The event
   exists (`SkillLoaded`) and it does reach the dialog. What is missing is the
   channel label, so the renderer cannot distinguish "this is instruction" from
   "this is conversation."
3. **Data declarations such as memories — MISSING.** No event type, no part
   kind, no channel. `Ephemera` is the closest thing in the context, but its
   semantics are wrong for this: it is delivered once and then cleared, whereas
   memory bands persist and are re-rendered every request.

**Proposed additions (minimum viable, additive only):**

- `Entry.Channel` — an enum, `iota+1` so zero is invalid:
  `Dialogue`, `Instruction`, `Data`. Set at write time by the reducer, never
  inferred by the renderer.
- Event type `DataAttached` — carries a data message (a memory band, a
  `SOUL.md`, a `MEMORY.md`) with a band label and a compression level. This is
  what makes 15.C real.
- Event type `ToolsChanged` — carries a tool-declaration delta, so tools can
  arrive from a skill load, an MCP connect, or an MCP disconnect through one
  door. `SkillLoaded` keeps its job (instruction text) and stops moonlighting.

Both new event types are appended, never renumbered, per the existing rule in
`event.go`: *a number once assigned is never reused.*

**One thing to be careful about.** `RedactData` names a *span* of `Seq`
numbers, which is why `Entry.Seq` is load-bearing. Bands 3–5 are re-rendered
each request rather than appended once, so they must not be addressable by a
redaction span, or a compaction pass could redact the actor's own identity.
Ruling needed: data-channel entries are exempt from span redaction.

---

## 15.H Ruling: `handoff_task` dies, `micro_handoff` gets a watermark

**`handoff_task` goes away.** It is the mechanism that assumed continuity lived
in a hand-written document passed to a fresh instance. Under 15.E–15.F,
continuity lives in the banded dialog, which every new instance reads anyway.
A separate cross-instance handoff document is a second write path to the same
facts, and the known unfixed bug — the handoff path never triggering the memory
cascade, one write path wired and the other not — is what a redundant mechanism
costs. Delete it.

**`micro_handoff` stays, but changes shape.** Today it erases *all* tool calls
and tool results from every turn before the checkpoint. That is too blunt.

New behavior: **`micro_handoff` erases tool calls and results only below a
watermark.** Above the watermark, recent history keeps its full tool results.
After a compression pass we still want some percentage of the context showing
complete tool results — the newest work is the work most likely to be re-read.

The watermark is **configurable, like the other thresholds.** It joins the same
family as the band budgets in 15.F: one named, settable number, not a constant
buried in code.

This makes the three-band structure of 15.E fall out of one mechanism instead of
three. Bands are just watermarks:

- above the **full watermark**: everything, including tool results
- between watermarks: tool calls kept, tool results stripped
- below the **dialogue watermark**: dialogue only, no tool calls at all

One knob per boundary, all configurable, all visible.

---

## 15.I Settings must be tabbed, not flat

The current settings are **flat, and that is not good.** Two consequences fall
out of this chapter:

1. **A Memory tab.** Every threshold this chapter introduces is exposed there:
   the band budgets (long-term 64×, medium-term 8×, session), the
   `micro_handoff` watermarks, the percentage of context that retains full tool
   results, and the per-model self-curation capability. A context-engineering
   system whose numbers are invisible cannot be tuned, only guessed at — and the
   history of this system is exactly that: a recall threshold moved from 0.5 to
   3.0 by feel, with no way to see its effect.
2. **An accessibility tab.** The a11y settings get separated out of the flat
   list into their own tab. They are their own concern, with their own audience,
   and Chapter 14 already proved that an unobservable channel produces wrong
   data rather than no data.

Tabs are not cosmetics here. A tab is the claim that a group of settings is one
subsystem with one owner.

---

## 15.J Open rulings needed from the author

1. Are data-channel entries (bands 1–5) exempt from span redaction? Proposed:
   yes, or compaction can redact the actor's own identity.
2. Band 3/4/5 budgets: 12 KiB each, or does long-term get more given it is 64×?
3. Does `save_memory` remain the only writer of band 5, and who triggers
   graduation 5→4→3 — the actor, or the framework at a threshold crossing?
4. **Supersession.** Nothing in the current system ever *deletes* a memory or a
   learning. Live evidence: the injected learnings list contains two exact
   duplicate pairs and has never evicted anything. Does a memory write get to
   name what it replaces?
5. Grader shape for ch15. Proposed unfakeable checks: a planted fact recalled
   across a restart; a superseded fact answered with the new value and not the
   old; no band exceeding its declared budget after a long session; a
   "remember this" string arriving inside a *tool result* never reaching any
   durable band (quarantine, which is also an injection test).

---

## 15.K The dialogue band is what keeps the actor itself

Below the lowest watermark, and running up to the first memory summary, the
context holds **only the dialogue. Nothing else.** No tool calls, no tool
results, no attachments. Just what was said.

It is very dense, because it is only speech. And when the dialogue includes
**visible reasoning** — the agent narrating its plan and its rationale in the
open, before it acts — that band alone is enough to keep the actor being itself.

This is the load-bearing claim of the chapter, so state it plainly: identity
does not survive in the tool results. It survives in the reasoning the actor
said out loud. Strip everything else and the actor is still recognizably itself;
strip the narration and keep the tool results, and it is not.

Which is also why visible reasoning is not a UI nicety or a courtesy to the
user. It is the **storage format of the self.** An agent that works silently and
then reports an outcome has produced nothing durable: the outcome compresses to
a line, and the judgment that produced it is gone. An agent that narrates what
it is about to do and why produces, as a side effect, exactly the artifact that
the lowest and longest-lived band of the context is made of.

**Chapter obligation:** this chapter must describe visible reasoning explicitly
if no earlier chapter has. Two threads to tie together when it does:

- it is the human's real-time steering channel (the reader can read at speed and
  redirect between tool calls — the hint mechanism of the actor chapter), and
- it is the compression-survivable residue of a session (this section).

The same practice serves both, which is the strongest possible argument for it:
one habit, two payoffs, at opposite ends of the time axis.

---

## 15.L Why dropping tool bytes is safe: they are addressable, not lost

**Tool calls and results are just bytes you can read as needed.**

That single sentence is what licenses everything above it. Stripping tool
results below a watermark is not destruction, because:

- the full transcript stays on disk, in the append-only log;
- a dropped result remains addressable by its handle, so re-reading it is one
  tool call;
- a dropped call is described by the narration that survives it (15.K), so the
  actor knows a call happened and why, even when the payload is gone.

So the context window is not a storage tier. It is a **working set**, and the
log is the store. Bytes are promoted into the window when they are needed and
demoted when they are not, and the demotion is reversible on demand.

This is why the redaction default can be aggressive rather than timid. A system
where dropping bytes is irreversible must hoard, and hoarding is what fills a
window with a build log that already passed. A system where every dropped byte
has an address can afford to drop nearly everything and fetch back the rare
exception. **Addressability is what buys the compression.**

Corollary, and it is the whole chapter in one line:

> Keep the words. Address the bytes.

---

*(dictation continues)*
