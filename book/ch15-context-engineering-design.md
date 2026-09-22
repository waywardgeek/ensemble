# Chapter 15 — The World's Best Context Engineering

*Design record. Dictated by Bill 2026-09-22 before breakfast, transcribed with
commentary and gap analysis by the coder.
Status: DESIGN COMPLETE, awaiting rulings in §15.O. Nothing here is prose yet.*

**Working title is the real title.** "The World's Best Context Engineering."

**One-paragraph thesis.** The context window is not storage; it is a working set
rendered from an append-only log. It has a frozen prefix (system prompt and
fixed tools, built from skills, changed only by a deliberate full-refresh cache
miss) and an append-only tail. Identity and memory ride in the tail as data
messages at four compressions — 64×, 8×, 1× session, live dialogue — each with a
byte budget and graduation upward when it overflows. The lowest band is
dialogue only, and because it contains the actor's visible reasoning, that band
alone preserves the actor. Everything else is tool bytes, and tool bytes are
addressable on disk, so discarding them is safe. Compaction itself runs beside
the conversation as sandboxed sub-agents and commits by silent swap, which is
safe for the same reason the tail is append-only. Keep the words; address the
bytes.

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

**Proposal (pending ruling): what survives above the watermark.**

- **Keep recent results, bounded in BYTES, not a percentage.** A percentage keeps
  more when the context is larger, which is backwards: the biggest contexts are
  exactly when a handoff should cut hardest. A byte budget matches the bands.
  Starting guess 16 KiB, unmeasured.
- **Why keep any.** The dominant risk after a handoff is not cost but
  reconstruction: the actor re-deriving facts from its own narration instead of
  re-reading them, which is the most reliably measured failure in this system.
  The newest results are the ones most likely to be quoted next. Stripping them
  either forces an immediate re-read (saving nothing) or invites a misquote.
- **Why it is cheap.** With auto-redaction on, most results are already stubs by
  handoff time; what remains full is the last round trip plus what the actor
  chose with `keep_tool_results` — pre-curated bytes. Under redaction,
  `micro_handoff`'s saving comes mostly from tool-call arguments, stubs, and old
  thinking, not results. Kept bytes raise `S` in the break-even formula of the
  micro-handoff design doc; they do not enlarge the cache miss, which starts near
  the top either way.
- **Capable models name what to keep.** An optional retain list on
  `micro_handoff`: the actor knows which results the next stretch needs;
  position does not. Weaker models get the positional watermark. Same
  one-shape-or-two split as §15.M.
- **Thinking is lost regardless.** The handoff document replaces it; kept results
  spare the handoff from restating facts still on the page.

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

Superseded by §15.O, which consolidates every open question in one numbered
list. Left as a heading so section references stay stable.

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

## 15.M The bottom of the layout has one section or two, by model capability

The context layout diagram ends differently depending on what the model can do.

**Weaker models: two sections.** First a band where tool results have been
redacted and only the calls remain, then the bottom band with full tool calls
and results intact. The split is mechanical — a watermark by position (15.H) —
because the model cannot be trusted to decide what matters.

**Stronger models: one unified section.** The two collapse into a single band,
and the model **self-curates** which tool results are important enough to keep.
Position stops being the criterion; relevance becomes the criterion, judged by
the actor.

This is the same capability axis as 15.E, now with a structural consequence
rather than just a policy one: **the shape of the layout is a function of the
model's features table.** A layout diagram in this chapter therefore needs two
variants, and the framework needs to render both from one description.

**Expected yield (author's estimate, not a measurement).** With a model capable
of self-curation, the estimate is that **at least 80% of the tool data gets
thrown away.** Flagged explicitly as a guesstimate so no later draft quotes it
as instrumented. If the chapter prints a number, the number must come from a
measured run — the band budgets and watermarks of 15.I make that measurable,
and once the Memory tab exposes them, measuring it is cheap.

**Why it is worth the friction.** Self-curation is not free: the actor
occasionally discards something it then has to fetch back (safe, per 15.L, but
it costs a round trip). It pays for itself twice over anyway:

1. **Cost.** Eighty percent of the bulkiest category of bytes, not re-sent on
   every subsequent request.
2. **Runway — the bigger win.** The actor runs far longer between
   `micro_handoff` calls. Every checkpoint is a discontinuity in the actor's
   working state, so fewer checkpoints means longer coherent stretches of work.
   Curation buys *continuity*, and continuity is the thing this whole chapter is
   about. Cheaper is the side effect.

---

## 15.N The layout diagram

Both variants of the steady-state context. This is the figure the chapter is
built around.

**Variant A — weaker model (mechanical curation, two bottom sections):**

```
┌─ FROZEN PREFIX ──────────────────────── never changes outside a refresh ─┐
│  system prompt, rendered from skills, $VAR substituted, fixed at start   │
│  fixed tool declarations                                                 │
└──────────────────────────────────────────────────────────────────────────┘
┌─ DATA CHANNEL ─────────────────────────────── re-rendered each request ──┐
│  1  SOUL.md                              self-curated identity          │
│  2  MEMORY.md                             self-curated permanent facts   │
│  3  long-term memories        64x         <= 12 KiB                      │
│  4  medium-term memories       8x         ~12 KiB                        │
│  5  session memories           1x         ~12 KiB                        │
└──────────────────────────────────────────────────────────────────────────┘
┌─ DIALOG ─────────────────────────────────────────────── append-only ─────┐
│  6  dialogue only: speech + visible reasoning. no tool calls at all.     │
│                                                                          │
│     ---- dialogue watermark ----                                         │
│  7  tool CALLS kept, tool RESULTS stripped                               │
│                                                                          │
│     ---- full watermark ----                                             │
│  8  everything: full tool calls and full tool results                    │
└──────────────────────────────────────────────────────────────────────────┘
```

**Variant B — capable model (self-curation, one bottom section):**

```
   ... bands 1-6 identical ...
┌──────────────────────────────────────────────────────────────────────────┐
│     ---- dialogue watermark ----                                         │
│  7  UNIFIED: tool calls kept; tool results kept only where the actor     │
│     judges them still relevant. position is not the criterion.           │
│     estimated ~80% of tool bytes discarded (unmeasured).                 │
└──────────────────────────────────────────────────────────────────────────┘
```

Read bottom-up, the ladder is one idea applied at five time scales: keep the
words forever, keep the calls for a while, keep the results only while they are
hot, and address everything else on disk.

---

## 15.O Consolidated open questions

Numbered so they can be answered by reference.

**Structure**

1. Are data-channel entries (bands 1–5) exempt from span redaction?
   *Proposed: yes.* `RedactData` names a span of `Seq`, and bands 1–5 are
   re-rendered rather than appended once, so a compaction pass could otherwise
   redact the actor's own identity.
2. Band budgets: 12 KiB each, or does the 64× band get more, since it buys the
   most history per byte?
3. Does the sum of band budgets get expressed as an absolute byte count or as a
   fraction of the model's window? Windows differ by an order of magnitude
   across vendors.
4. Where do **learnings** and **skills** live in this layout? Neither appears in
   15.F. Learnings are currently injected into the system prompt, which 15.B–15.C
   now forbids for anything that changes at conversational speed. Candidate:
   learnings become a data-channel band; skills stay instruction-channel.
5. Where does **project context** (per-repo instructions) live — frozen prefix,
   or data channel? It changes when the repo changes, not when the turn changes,
   so probably prefix.

**Mechanism**

6. Who triggers graduation 5 → 4 → 3: the actor, or the framework at a threshold
   crossing? *Leaning framework*, because a threshold crossing is objective and
   the actor should not have to remember to tidy.
7. Does `save_memory` remain the only writer of band 5?
8. **Supersession. ANSWERED by §15.S.** Nothing in the current system ever
   deletes a memory or a learning. Live evidence: the injected learnings list
   contains two exact duplicate pairs and has never evicted anything. *Ruling:
   a compaction event is a replacement mapping — naming what it replaces is the
   entire content of the event — so supersession is structural rather than an
   added field. Accumulation was only possible because compaction was a silent
   mutation that left no record of what it consumed.*
9. Provenance at write time: does every memory entry carry VERIFIED vs ASSUMED?
   The recurring failure in practice is that regenerated figures arrive with the
   confidence of copied ones, and a memory offers no way to tell them apart
   later.
10. **Quarantine.** May content that arrived in a tool result — a crawled page,
    a sub-agent's output, a file from an untrusted repo — ever be promoted into
    a durable band? *Proposed: never without an explicit act by the actor.*
    This is a context-engineering concern and an injection defense at the same
    time.
11. Does the self-curation capability flag belong in the existing model features
    table? *Proposed: yes, with no default row*, matching streaming and media.

**Deliverables**

12. Is there a `context_report` tool that prints actual bytes per band against
    budget? Today the actor can see its total token count but not the breakdown,
    so it cannot tell which band is crowding the others.
13. Grader shape. Proposed unfakeable checks:
    - a planted fact recalled across a restart (planted token, unfakeable);
    - a superseded fact answered with the new value, with the old value absent
      from the current bands — both directions, so it cannot pass by ignoring
      supersession;
    - no band exceeding its declared budget after a long session;
    - a "remember this" string arriving inside a *tool result* never reaching
      any durable band (quarantine);
    - an automated write to the identity band refused.
14. Does the chapter need to describe visible reasoning from scratch, or does an
    earlier chapter already cover it? Not yet checked against the manuscript.
15. Chapter placement of the settings work (15.I). It is GUI work in a context
    chapter. Split into its own short chapter, or carry it here because the
    numbers are meaningless without somewhere to see them?

**Asynchronous compaction (§15.R)**

16. Are bands 1 and 2 (`SOUL.md`, `MEMORY.md`) hard-forbidden to every
    automated writer? *Proposed: yes.* Self-curated means self-curated; an
    automated process that can rewrite the identity document is a personality
    drift generator.
17. What sandbox do compressor sub-agents run under? *Proposed: no shell, no
    web, no filesystem beyond their own input and output band.* They read
    untrusted tool results and write near the actor's identity, which makes
    them the highest-risk component in the design.
18. Does a compaction commit need a **validation gate** before the swap — for
    example, output must be under budget, must be non-empty, and must not have
    dropped every entry of a given kind? Without one, a bad compression silently
    becomes the actor's past.
19. Is there a floor on compaction frequency to protect the prompt cache? Each
    commit invalidates from the changed band downward, so a pathological
    trigger pattern could invalidate on nearly every turn. *Proposed: commit
    only at turn boundaries, and batch all ready bands into one commit.*
20. Who observes compaction in the GUI, and how loudly? *Proposed: silent in
    the conversation, an event in the log, and a quiet indicator plus a
    `context_report` breakdown on demand.*

**Recall (§15.Q)**

21. Does the auto-recall judge see the bands, or only the candidate fragments?
    If it cannot see what is already present, it will recall things the actor
    already knows and spend budget to say them twice.
22. Do recalled fragments enter the data channel as ephemera (delivered once,
    then cleared) or as persisting entries? *Leaning ephemera*, since relevance
    was judged for one turn and does not transfer to the next.

---

## 15.P Claims this chapter must defend

Collected so they can be attacked individually.

1. The context has a frozen prefix and an append-only tail, and **every byte
   belongs to exactly one of them.** No clever middle ground.
2. Identity and memory are **data in the dialog**, not instructions in the
   prompt. Placement is a renderer decision.
3. Memory is a **reduction over the log**, not a store. Bands are reductions at
   different compressions. (Same thesis as one log, three vendors — here it is
   one log, one self.)
4. **Visible reasoning is the storage format of the self.** The dialogue-only
   band plus narration is sufficient to preserve the actor.
5. **Tool bytes are addressable, so discarding them is safe.** Addressability is
   what buys the compression: a system where dropping is irreversible must hoard.
6. Any artifact with **two writers will rot.** Single writer per band. The
   deleted `handoff_task` and the never-wired cascade trigger are the receipt.
7. Curation buys **runway**, not just cost. Fewer checkpoints means longer
   coherent stretches of work.
8. **Model capability is a structural input.** The layout has two shapes, chosen
   from a features table, not one shape with a policy knob.


---

## 15.Q The memory system: push bands and pull recall

Two distinct mechanisms feed the data channel, and conflating them is how the
current system got confusing. Name them separately.

**Push — the bands.** Bands 1–5 of §15.F are *always present* in every request.
They are governed by byte budgets and graduation. Nothing decides whether to
include them; they are the actor's standing state.

**Pull — auto-recall.** Separately, relevant fragments are *fetched* per turn
based on what the user just said. Governed by a relevance threshold, not a byte
budget. Nothing is standing; every recalled fragment must earn its place on
this turn.

The pieces, exactly as they exist today and as they should be kept:

1. **Bucket compressor agents.** Each compression step is performed by an
   agent, not by a function. Band 5 → band 4 at 8×, band 4 → band 3 at 64×.
   Each compressor has a bounded input (one band), a bounded output
   (input ÷ ratio), and no need for the parent's context — which is what makes
   §15.R possible.
2. **`save_memory` triggers the whole cascade** if a threshold crossing has not
   already triggered it. Two triggers, one mechanism. This matters: the known
   historical bug in this system was a *second* entry point (`handoff_task`)
   that never fired the cascade at all, so memories silently failed to
   graduate. With `handoff_task` deleted (§15.H) there are exactly two triggers
   and both are wired. **One mechanism may have several triggers, but every
   trigger must be wired to the same mechanism** — the failure mode is always
   the unwired second path.
3. **BM25 plus a vector database** for searching the corpus that auto-recall
   draws from. Keyword and semantic retrieval are complementary: BM25 wins on
   exact identifiers, file paths and error strings, the vector index wins on
   paraphrase. A coding agent needs both, because half its recall queries are
   literally symbol names.
4. **A low-power LLM as the relevance judge** for auto-recall. Retrieval
   proposes, the judge disposes. The judge must be cheap by construction,
   because it runs on every turn: the cost of judging has to sit far below the
   cost of the bytes it prevents from being injected, or the mechanism is
   negative-value. A small model is not a compromise here, it is the design.

**Why the judge exists at all.** A pure-score threshold cannot distinguish "this
fragment mentions the same words" from "this fragment answers the question."
The history of this system is a recall threshold moved from 0.5 to 3.0 by feel,
which is what tuning a scalar in place of a judgment looks like. Replacing a
hand-tuned number with a cheap model that reads the fragment and the question is
the entire improvement.

---

## 15.R Asynchronous compaction: compress in parallel, swap silently

**The proposal.** Do not stop the world to compact. Run `save_memory` and the
bucket compressors **in parallel, as sub-agents**, while the conversation
continues. When they finish, **silently switch** to the newly curated context
plus every round trip that happened in the meantime.

**Verdict: correct, and correct for a structural reason.** This is
read-copy-update. RCU is safe precisely when the snapshot cannot mutate
underneath the writer, and §15.D already guarantees that: the tail is
append-only. Therefore:

- the compressor snapshots bands as of `Seq = N`;
- it produces replacement bands, also as of `N`, off to the side;
- the commit swaps bands-as-of-`N` and re-appends entries `N+1..now` unchanged.

The merge is **concatenation, not a three-way merge.** There are no conflicts to
resolve because nothing below the watermark can have changed. The feature is
cheap only because the architecture was right first; in a system whose history
could mutate, this would be a distributed-systems problem.

**The per-band single-writer rule survives, and it tells us the swap is
per-band.** Band 5's writer is `save_memory`; band 4's writer is the 5→4
compressor; band 3's writer is the 4→3 compressor. A commit replaces one band
and never touches another band's entries, so a `save_memory` landing during
compression is appended to band 5 rather than clobbered.

**A pleasant result: the layout ordering is optimal for the prompt cache, and
for the same reason it is semantically right.** Cache validity is a prefix
property, so the cheapest possible arrangement puts the least-frequently-changed
bytes earliest. Change frequency ascends exactly as the table descends:

| Band | Changes when | Cache consequence of a swap |
|---|---|---|
| `SOUL.md` | a deliberate act of self-revision | almost never invalidated |
| `MEMORY.md` | a deliberate act of curation | rarely invalidated |
| 64× long-term | a 4→3 graduation | rare |
| 8× medium-term | a 5→4 graduation | occasional |
| 1× session | every `save_memory` | frequent |
| dialogue | every turn | always |

So a compaction commit invalidates the cache from the changed band downward and
no further. Compacting band 4 costs bands 4, 5 and the tail; `SOUL.md`,
`MEMORY.md` and band 3 stay cached. **The ordering was chosen for meaning and
turns out to be the cache-optimal one too.** When two independent arguments
select the same layout, that is the strongest evidence available that the layout
is right.

Two practical consequences: **commit at a turn boundary, never mid-turn**, and
**batch the commits** — swapping three bands in one commit costs one
invalidation, swapping them separately costs three.

**On visibility — the question of whether users want to see this.** The right
answer is *silent by default, observable on demand, never hidden.* Those are
three different things:

- **Silent**: the conversation is not interrupted with progress chatter. A
  compaction is infrastructure, and infrastructure that narrates itself is
  noise.
- **Observable**: the commit emits events into the log — compaction started,
  compaction committed, with band labels and before/after byte counts — so the
  GUI can show a quiet indicator and a `context_report` can explain exactly what
  happened and when.
- **Never hidden**: if compaction is unobservable and it eats something
  important, nobody can distinguish "the agent forgot" from "the compressor
  dropped it." The accessibility chapter already established the general form of
  this: *an unobservable channel produces wrong data, not no data.* Compaction
  is a channel.

**Replay stays intact, but only if the commit records bytes rather than
intent.** A compressor is an LLM, so its output is nondeterministic and cannot
be re-derived by re-running it. The commit must therefore be an **event carrying
the produced bytes**, exactly as redaction is. Replay reads the recorded output
instead of recompressing. This is the same ruling as "compaction is an event,"
now with a sharper reason: the procedure is not reproducible, so only the result
may be authoritative.

**Failure degrades safely, which is a property of RCU rather than an accident.**
If a compressor crashes, stalls, or produces garbage that fails validation, the
commit simply never happens and the old bands remain live. The failure mode is
"context stays larger than we wanted," which is survivable. Two guards are still
needed: **at most one compressor in flight per band**, so a stuck one does not
spawn a new one every turn, and a **watchdog** that abandons rather than blocks.

**The hazard that needs an explicit ruling.** A bucket compressor is an LLM that
reads tool results — crawled web pages, sub-agent output, files from untrusted
repositories — and writes into bands that sit adjacent to the actor's identity.
That is precisely the promotion path that §15.O question 10 proposes to forbid,
except automated, unattended, and running on every session. It is the most
security-sensitive component in the entire design.

Proposed constraints, all of which narrow rather than widen:

- compressors run **sandboxed**: no shell, no web, no filesystem beyond their
  own input and output;
- a compressor may write **exactly one band** and nothing else;
- compressor output is committed with `Channel = Data`, so it can never be read
  as instruction — a third payoff for the channel field of §15.G;
- **bands 1 and 2 are not writable by any compressor.** `SOUL.md` and
  `MEMORY.md` are self-curated by definition (§15.E). An automated process that
  can rewrite the actor's identity document is not a memory system, it is a
  personality drift generator.

**What the parallelism actually buys.** Not merely latency. Today compaction
stops the actor at the worst possible moment, because the threshold is crossed
in the middle of real work. Running it beside the conversation converts a
visible stall into an invisible background cost, and converts an expensive
model's time into a cheap model's time. Combined with §15.M, the effect is
cumulative: self-curation lengthens the interval between compactions, and
asynchrony removes the cost of the ones that remain.

---

## 15.S Compaction is described by events, never performed on the context

**The ruling.** We never simply compact the context. We **create events that
describe how the context is compacted.**

A **memory graduation judge** emits, via a tool call, an event that states
*which memories are replaced by which compressed version*. The context is then
**deterministically derived** by replaying the log — including the compaction
events.

This is a sharper form of "record bytes, not intent," and the sharpening
matters. The event is not an opaque blob of new text. It is a **replacement
mapping**: these entries, identified by `Seq`, are superseded by this produced
artifact. Four consequences fall out, and each one closes something that was
open.

**1. The context becomes a pure reduction over the log.** Not "mostly a
reduction, except for memory." Compaction events are ordinary events, applied
in order by an ordinary reducer. Replay produces the identical context because
the LLM's judgment is *already recorded* rather than re-derived. The
nondeterminism is quarantined into the event payload, at the boundary, where
nondeterminism belongs. The reducer stays pure.

**2. The tool call is the only door.** The graduation judge is an agent, and
its output enters the log exactly the way every other agent action does. There
is no privileged write path into memory. This is the same principle that killed
`handoff_task`: a second way to write the same facts is a second way to get it
wrong.

**3. Supersession is solved structurally, not added as a feature.** Open
question 8 asked whether a memory write may name what it replaces. Under this
design, a compaction event *is* a supersedes record — naming what it replaces
is the whole content of the event. Accumulation was only ever possible because
compaction was a silent mutation with no record of what it consumed.

**4. The memory files stop being sources of truth.** `MEMORY.md`, the bucket
files, the session logs — these become **projections of the log**, rendered
views, not storage. That kills the two-writer problem at its root rather than
policing it: a file with two writers rots, but a file with *no* writers, derived
on demand from a single append-only log, cannot.

**The reducer must be total.** It may never refuse to replay a log. A
compaction event naming an entry that a prior compaction already superseded is
a **no-op with a diagnostic**, not an error. A malformed event is skipped and
noted, not fatal. The reason is blunt: the log is the ground truth, so a reducer
that can fail on bad input is a reducer that can brick the agent permanently.
Totality is not robustness polish here, it is the difference between a bad
event costing one memory and costing the entire past.

**On the vendor's offer to do this for us.** The Anthropic API ships
server-side context editing under the beta header
`context-management-2025-06-27`, with strategies `clear_tool_uses_20250919` and
`clear_thinking_20251015`. It is worth being precise about it, because it is
*not* incompatible with this design the way stateful conversation APIs were in
Chapter 2. Quoted from the provider documentation:

> "Your client application maintains the full, unmodified conversation history.
> **You do not need to sync your client state with the edited version.**
> Continue managing your full conversation history locally as you normally
> would."

So it is stateless, non-destructive, a per-request rendering decision rather
than a mutation of the thread, and it is even observable: the response carries
`context_management.applied_edits` with `cleared_tool_uses`,
`cleared_thinking_turns` and `cleared_input_tokens`.

We still decline it as the primary mechanism, for three honest reasons rather
than a slogan:

- **Scope.** It clears tool results and thinking. It does not do memory
  graduation, which is most of this chapter.
- **Policy shape.** Its configuration is declarative — `trigger`, `keep`,
  `clear_at_least`, `exclude_tools` — so it curates by position and rule. The
  self-curation of §15.M is relevance-based judgment by the actor, which a
  declarative trigger cannot express.
- **Determinism, which is the decisive one.** If the server decides what to
  clear, the rendered context depends on the server's behavior *at replay
  time*, which is not guaranteed to match its behavior at record time. Our
  replay property would become a claim about someone else's deployment.

It remains useful as a **backstop** for vendors and situations where our own
curation is unavailable, and the applied-edits field should be recorded in the
log when it fires, so that even the backstop is observable.

---

## 15.T Crash-safe persistence: snapshot plus tail

**The gap.** We do not currently write events to disk as they happen. That is
the piece that has to exist before any of the above is trustworthy, because a
design whose ground truth is an append-only log is only as good as the moment
that log reaches storage.

**The model.** Standard checkpoint-and-replay, with one detail that makes it
work:

- **Snapshot.** The actual rendered context, saved as of **a specific event in
  the log.** The snapshot must name the `Seq` it corresponds to. Without that
  anchor, replay cannot know where to resume and the snapshot is worthless.
- **Tail.** Events appended to the log on the fly, as they occur. After a
  crash, recovery is: load the most recent snapshot, replay every event after
  its anchor, and the most recent state is back.
- **Normal shutdown** also writes a snapshot, so the common case starts from a
  current checkpoint with an empty tail.
- **Prior events** — those from before the snapshot anchor — are needed only
  for the GUI and the audit trail. They are not required for recovery. That
  distinction sets the retention policy: retention is a *display and audit*
  question, not a correctness question.

**Retention.** The default today is roughly 100 events. That should be a
**config setting**, joining the other thresholds in the Memory tab of §15.I.
Time-based truncation of the log is a plausible later option, but for now keep
it simple: retain the number of events the GUI wants to display, and no policy
beyond that.

**The save procedure.**

1. **Save the current context to its own file**, first copying the prior
   context file to a backup and overwriting the old backup. One generation of
   history, so a corrupt write cannot destroy the last good snapshot.
2. **Truncate the log to N events**, where N is the retention number configured
   for the GUI in settings.

**The context file must carry the event `Seq` it was taken at.** This is not
bookkeeping, it is the entire recovery mechanism. On load, compare the
snapshot's `Seq` against the head of the log:

- equal — the snapshot is current, nothing to replay;
- behind — replay every event after the snapshot's `Seq` to reach the most
  recent state.

That is how a crash is survived. The process dies mid-session, the snapshot on
disk is from some earlier moment, the log has been accumulating since, and the
difference between them is exactly the work to redo.

**Ordering invariant: snapshot first, then truncate.** The two steps are not
commutative. Snapshotting at the current head means there are, at that instant,
no events after the anchor, so truncation can only remove events *before* it —
and those are display-and-audit only (§15.T above), never needed for recovery.
Reverse the order and a truncation can delete events that the not-yet-written
snapshot would have needed. Same class of bug as the unwired cascade trigger:
two steps that work individually and destroy data in the wrong sequence.

**Safety rule on N.** Truncation must never cut past the snapshot anchor. If a
snapshot is stale — say recovery has just happened and no new snapshot has been
written — then a small N could otherwise delete events required to replay
forward. So the rule is: **retain N events or everything back to the snapshot
anchor, whichever is more.** N is a display preference; the anchor is a
correctness boundary, and a preference must never be allowed to overrule a
correctness boundary.


---

## 15.U Field notes from the first Opus 5.5 session (2026-09-22)

Measured on the running system, not predicted.

**1. The cost of a mutation is proportional to its distance from the tail.**
Redaction currently stubs the previous round trip's tool result on every request.
The prediction was that this would wreck the prompt cache, since it rewrites bytes
already sent. It does not: the observed cumulative hit rate was 77% after a cold
first request that missed on over 100K tokens, which implies a steady-state
per-request rate well above that. The reason is that a prefix cache invalidates
only from the first changed byte onward, and a just-finished result sits at the
tail, so the miss is bounded by roughly one round trip of bytes.

So the frozen-prefix rule of §15.D refines to a cost model rather than a ban:

- **near the tail**: mutation is nearly free — redact freely;
- **deep in the context**: mutation invalidates everything after it — do it
  rarely and batch it (the `micro_handoff` watermark move, a compaction commit).

This is the same monotonic argument as §15.R's band ordering, applied to the
dialog: things that change often belong at the end, and changes that must reach
deep are expensive in proportion to how deep they reach.

**2. A redaction stub's address must outlive the process.** Stubs cite recovery
paths like `cr/io/26`. The job handle counter resets on server restart, so after a
restart the same path names a different command's output. §15.L's claim that
dropped bytes are addressable is true only within one process lifetime today.
Fix: session-scoped or content-addressed paths, so an address never gets reused.

**3. Cumulative hit rate hides the steady state.** A cumulative figure is
dominated by the cold start and converges slowly. Show last-request hit rate
beside it.

**4. Keep-tool gating works; the approval ratchet caught an unapproved flip.**
With the model row set to support redaction, `keep_tool_results` appeared and
retained a marked 5.1 KB result as designed. Thinking survived a four-step chain
in which an in-chain result was redacted. But the model was not added to
`redactionCapableModels`, so `TestRedactionCapableModelsAreExactlyTheApprovedSet`
fails. The ratchet is doing its job: a capability claim needs a measurement.
