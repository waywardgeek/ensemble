# Context engineering — chapter notes

**Status:** captured 2026-09-12, not yet outlined. Placement TBD (open question
7 in `chapter-02-outline.md`): depends on tool results existing (Ch3) and on the
system prompt existing (Ch6 skills), so Ch7 or later. The chapter's *hook* —
the `Redacted` event with a span and a level — is already established in
Chapter 2 §2.4a, so this can land wherever it reads best without a rewrite.

**Provenance:** Bill's design, stated 2026-09-12, for a context system he has
partly built. Items marked **[derived]** are mine, developed in conversation and
not yet ruled on. Items marked **[TBD]** are explicitly unsettled by him.

---

## The thesis

> **Know what you are throwing away.** Purge categories first, summarize last.

Set a token threshold. When the context reaches it, compaction begins
automatically — but *not* by replacing the oldest portion of history with a
summary.

**Compaction by position vs compaction by category.** one widely used agent SDK (and the
common framework approach) replaces a portion of history with an LLM-written
summary. Done. That is compaction by **position**: it discards whatever happens
to be old, valuable or not, and what it loses is unpredictable — a summary is
lossy in ways nobody enumerated.

Compaction by **category** discards a *kind* of content wherever it appears.
The categories are wildly unequal: tool results and tool-call arguments are the
clear majority of a real session's tokens while carrying almost none of its
continuity. Reasoning, decisions, and sense of purpose are cheap and cannot be
regenerated.

A category purge is lossy in a way you can **name and have measured**. A summary
is lossy in a way you discover later, in production, as a personality change.

## The measured finding that drives the design

**You still remain you if all tool calls and results are eliminated.** That is
major compaction — and it is an empirical result from our own experiments, not
a prediction. It is the single most valuable lever available, and it is
available precisely because the discarded category is enormous and nearly
continuity-free.

Two tool calls **survive the purge**: `handoff_task` and `save_memory`. They
are not incidental work; they are the record of what the work *meant*.

## The compaction gradient

Newest to oldest, in the message history:

| region | what it holds |
|---|---|
| most recent conversation | everything, in full — tool calls and results, possibly with result-level redaction |
| older, past the most recent `save_memory` | tool calls and results purged; **visible reasoning present** |
| older still | conversation and visible reasoning go too — **but never the stack of goals, no matter how much we compact** |
| oldest | when memories and handoffs pile up back-to-back and cost too many tokens, *then* an ADK-style summary — a compressed memory |

**The purge boundary is the most recent `save_memory`, maybe earlier. [TBD]**

Note the shape of the argument: summarization is not rejected, it is
**demoted**. It is the last resort applied to material that has already been
compacted by category, not the first tool reached for.

## Memory lives in the message history, not the system prompt

The full ordering, front to back:

1. `MEMORY.md` — the final long-term memory
2. some number of **64× compressed** memories
3. some number of **8× compressed** memories
4. uncompressed memories
5. uncompressed dialogue including visible reasoning, **mixed with** uncompressed
   memories
6. the portion that still has tool calls, possibly with result redaction

**This moves memory out of the system prompt into the message history where it
belongs.**

**[derived] The gradient is also a volatility ordering, which turns this into a
cache argument rather than an aesthetic one.** Most-compressed material changes
least often and sits at the front, where the stable prefix lives; new material
appends at the back. Memory in the *system prompt* invalidates the entire prefix
every time a memory is written — the worst possible position for the most
frequently-changing content. Memory positioned by compression level appends
instead. Worth stating as a checkable claim, not a preference.

**The quantitative backing is in `Usage`.** Cache reads cost roughly an order of
magnitude less than plain input, and cache writes cost *more* than plain input.
So misplacing volatile content at the front of the prefix does not merely fail
to save money — it converts the cheapest token category into the most expensive
one, on every request, forever. Chapter 2 §2.4a defines the four disjoint
categories that make this measurable; without them the argument here is a
preference, and with them it is arithmetic.

## Mechanisms already fixed in Chapter 2

These are settled and the chapter inherits them:

- **Redaction is a family, not a flag.** `RedactData{From, To, Level,
  Replacement, Reason}` — the span says *where*, the level says *what*.
- **Levels, weakest first:** `RedactResult` (stub the result, keep the call) →
  `RedactTool` (call and result go, reasoning survives) → `RedactDialogue`
  (prose and reasoning go, goal stack never does) → `RedactSummary`.
- **Compaction is an event.** It goes in the log like everything else. This is
  what lets the log stay complete and append-only while the context stays
  bounded, and it means replay reproduces the *compacted* context exactly. A
  system that compacts by mutating in-memory history has silently given up
  replay.
- **Stubs are synthesized, not stored** — computed by the reducer from the event
  superseded (tool name, size, the path the output still lives at).
  Deterministic, recoverable, and no storage that grows. Only `RedactSummary`
  stores a `Replacement`, because only there is the content something an LLM
  wrote that nobody can recompute.
- **No field in the context may grow without bound.**

## Open questions

1. **Where exactly is the purge boundary?** The most recent `save_memory`, or
   earlier? **[TBD — Bill]**
2. **Does `keep_tool_results` survive into this scheme? [derived, flagged]** By
   the design's own logic it may not: if purging *all* tool calls is the better
   compaction, an opt-out for individual results is solving a problem the purge
   does not have. Bill's live design decision, not recorded as settled.
3. **Does the goal stack become a `Context` field, and when?** **RULED
   2026-09-12: not in Chapter 2 — it belongs to THIS chapter.** It survives
   every level of redaction, so it is first-class, and it satisfies the
   no-unbounded-growth rule (bounded by nesting depth, not by time). But
   Chapter 2 has no concept of goals and cannot motivate one. Deferral is cheap
   for a structural reason: adding a *new* field later is additive, while
   reshaping an existing one is not — which is why `RedactData` had to be fixed
   in Chapter 2 and this did not. **This chapter must define both the field and
   the policy that needs it**, since `RedactDialogue` no longer names its
   survivors and now defers to "whatever the compaction policy designates."
4. **How do the compression ratios get chosen?** 8× and 64× are the shipped
   cascade's numbers; the chapter should say whether they are principled or
   empirical.

## Existing implementation to draw on

CodeRhapsody's shipped **Memory Cascade v2** already implements much of the
lower half: 8× compression, fixed-size buckets, a maximum of two buckets, and
graduation into `MEMORY.md`. Treat it as a source of measured detail — and as a
source of *corrections*, since the chapter's claim is that the cascade should
live in the message history rather than the system prompt, which is not where it
lives today.

**For the coder:** every number in this file needs verification before print.
The token-share claim about tool results, the compression ratios, the bucket
sizes, and the "identity survives the purge" result are all currently the
author's and Bill's working knowledge.


---

## War story: `compress_context`, and why LLM-discretionary compaction died

**Source: Bill, 2026-09-13, first-party.** Recorded verbatim because the
mechanism matters more than the summary.

> "compress_context was the old system that you (Claude) did pretty well,
> picking a range of messages to summarize, but freaking Gemini almost always
> deleted 80% of messages, starting with message 1, with a terrible summary,
> lobotomizing the LLM, so we switched to handoffs instead."

### Why this is the best available argument for structural compaction

Same tool. Same prompt. Same instructions. **Two models, and one of them
routinely destroyed the conversation it was asked to condense.**

That is a receipt no amount of reasoning substitutes for, and it kills a
tempting design in one sentence: *let the model decide what to drop.* It works
on the model you tested and silently lobotomizes the agent on the model you
did not. A capability that is correct on one frontier model and catastrophic on
another is not a capability; it is a coin flip with a good day.

Note the shape of Gemini's failure, because it is exactly the failure mode the
notes above predict: it compacted **by position** — "starting with message 1" —
rather than by category. Position-based compaction cannot distinguish the goal
stack from a stale directory listing, so it eats the former first, since the
oldest messages are where the goal was stated. **The agent forgets what it is
doing before it forgets anything it could afford to lose.**

### The succession, which the book should tell as a sequence

1. **`compress_context`** — model picks a range and summarizes it. Died of
   model-dependence. 154 calls in the corpus, all historical.
2. **`handoff_task`** — write a structured document for a fresh instance.
   Deterministic, author-controlled, survives the model swap. **Currently in
   use, and already slated for deprecation.**
3. **What replaces it:** keep the actor permanently knowledgeable about its own
   history and current work, through more refined context engineering, rather
   than periodically resetting it and handing over a note.

**Bill, same session:** *"Handoffs will be deprecated [...] and we'll instead
try to keep an actor permanently knowledgeable about its history and what it is
doing, with more refined context engineering."*

This is worth printing as a **live** trajectory rather than a solved one. The
book gains more from "here is where this is going and here is what each step
cost" than from pretending the current answer is the final one. It also keeps
the chapter honest: handoffs are taught while being openly marked as
transitional.

### Corollary for the tool table

`refine_context` and `keep_tool_results` both survive, and for a reason worth
stating: **redacting tool results the agent did not keep is far more effective
than `refine_context`, but not every model handles `keep_tool_results` well.**
So both ship. That is the same lesson as above in a smaller frame — the better
mechanism is kept *alongside* the worse one because model capability is uneven,
not because the design is undecided.
