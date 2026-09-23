# Chapter 16 outline: memory (Chapter B of the context arc)

Source of truth: `book/ch15-context-engineering-design.md`, Part III, plus the
evening rulings at the end of that doc (they override Part III where they name
it). Builds on Chapter 15's channels, entry kinds, compaction events and total
reducer, and on Chapter 11's anchored save. Where this outline and the design doc
disagree, the design doc wins until Bill rules; flag, never resolve silently.

Workflow (same as ch15): design doc → author writes TL;DR + §16.1 plain words →
coder builds `solutions/ch16` and writes `book/review-ch16.md` → author revises
TL;DR + plain words against the review → author writes the body.

---

## Title candidates

1. **Forgetting on Purpose** (names the activity, per the ch5 rule; the chapter's
   real skill is deciding what to drop and recording that it was dropped)
2. **What Survives the Wipe** (names the test the grader runs first)
3. **Memory** (plain; loses the argument)

Recommendation: 1.

## Through-line stake (voice.md §4)

An agent with no memory meets a stranger every morning (ch11 fixed the process
restart, not the window). An agent whose memory only grows is no better: it keeps
two copies of the same fact, cannot tell which is current, and pays for both on
every request. Memory is a budget with a door in and a door out, and every
removal is an event someone can read.

## Per-section wild fact (voice.md §7.3)

Every fact below is on file; cite as shown. Nothing may be replaced by a figure
reconstructed from memory.

| § | Wild fact | Source |
|---|---|---|
| 16.0 | The author's own injected learnings list carries two duplicate pairs (fiction reading preferences; the Lyric directory rename), and nothing has ever evicted either | design doc §B.6 (VERIFIED by observation); re-verify against the live prompt before print |
| 16.2 | The shipped system's second compaction path, `handoff_task`, never fired the cascade; first seen 2026-04-18, never fixed | design doc §B.3; `cr/MEMORY.md` "Known bugs" |
| 16.3 | The shipped cascade's 8x-per-step ratio is inherited, not derived; nobody has measured whether 8x is right | design doc §B.4, Q10 |
| 16.9 | The recall threshold was moved from 0.5 to 3.0 by feel | design doc §B.9 (**conflict:** ch15 outline also claims this fact for §15.13; one chapter gets it, and auto-recall is here) |
| 16.9 | The whole recall corpus is ~5.6 MB (~213 files project + ~1,000 files global); a monthly vector-database bill for a few MB of vectors is a category error | design doc §B.9 (measured 2026-09-22, not re-measured) |
| 16.9 | BM25 recall@10 on that corpus has never been measured; misses are invisible by nature | design doc §B.9 |

## Voice plan

- Third person throughout the body. Short first-person opener allowed.
- **Bill moments (≤2):** (1) the "Literature? We're good" ruling on vector
  databases, if it earns its place; (2) the opener if Bill supplies one.
- **[STORY SLOT — opener], PROPOSED, needs Bill:** the duplicate learnings. The
  author's own system prompt, the thing that is supposed to be its long-term
  memory, says the same fact about Bill twice and has no mechanism to notice.
  Contrast with ch11's goldfish: that agent forgot everything; this one forgets
  nothing, and both are broken. Alternative: Bill's own account of what memory
  failure costs him day to day.
- Measurements printed as measured, with the command that produced them (§16.12).
  Anything unmeasured is labeled so in the prose, not softened.

## Sections

### §16.1 In Plain Words
Two mechanisms, named separately. **Push:** the memory bands, present on every
request, governed by byte budgets. **Pull:** auto-recall, fetched per turn by
relevance. Conflating them is how the shipped system got confusing (§B.2). One
door in (`save_memory`), one way up (graduation), one way out (supersession
events). The context is still a fold of the log; memory is just more events.

### §16.2 `save_memory`: the one door
- The actor writes its session memory in its own voice, as the tool argument.
- Only `save_memory` starts the cascade; there is no threshold trigger (§B.3).
  The framework may *demand* the call (ruling 5), but the door is the same.
- At the end of the turn, `MemorySaved` deletes every Dialogue and Handoff entry
  before it, absorbs the micro_handoff documents, unloads dynamic skills
  (recording which, so reload is one call), applies pending lazy removals, and
  appends the new Memory entry after the last memory.
- **Why end of turn:** mid-turn, the deletion would remove the assistant message
  holding the call and orphan its tool result (§B.3, answers Q4).
- Cost: one conversation-region cache miss per milestone. Accepted.

### §16.3 The bands
Table from §B.4 (SOUL, MEMORY, 64x, 8x, session). Budgets are absolute bytes,
configurable, starting guess ~12 KiB each (UNMEASURED, Q9). Layout order is
change-frequency order, so the most stable bands stay cached longest (§I.6).

### §16.4 Graduation is a measured chain
Ruling 7. `save_memory` measures the session band → launches the 8x compactor if
over → on finish, measure 8x → 64x compactor → on finish, measure 64x → MEMORY.md
curator. Two events per stage: `CompactorLaunched{Band, AsOf}` (observable,
never replayed as work) and `MemoryCompacted{Band, Replaces []Seq, Text}` (the
mapping the reducer applies). **Replay never re-runs an LLM.** A launch with no
finish on crash recovery is relaunched or abandoned (Q to Bill below).

### §16.5 Compressors are agents
Bounded input (one band), bounded output (input ÷ ratio), no need for the
parent's context. Upgrade a model to compact, never downgrade (ruling 6); the
switch is visible because provenance is per model (ch2). Sandbox profile (no
shell, no web, one writable band) is Q15: here, or deferred to sandboxing.

### §16.6 Compress beside the conversation
Read-copy-update (§B.7). Snapshot the band as of N, build the replacement off to
the side, commit at a turn boundary by swapping the band and re-appending N+1..now.
The merge is concatenation, because the tail is append-only. Failure is safe by
construction: a compressor that never commits leaves the old band live. One
compressor per band in flight; a watchdog that abandons rather than blocks.
Silent by default, observable on demand, never hidden (ch14's rule: an
unobservable channel produces wrong data).

### §16.7 Supersession is structural
A graduation event names what it replaces; that is its whole content. The
duplicate pairs of §16.0 were possible only because compaction was a silent
mutation. Memory files (`MEMORY.md`, band files) become projections of the log:
a file with no writers cannot rot.

### §16.8 Learnings
Startup block as one data entry; `LearningAdded` for new ones, appended at the
tail (invalidates nothing); `delete_learning` applied lazily at the next
`save_memory`. Consolidation by the graduation judge is Q21.

### §16.9 Auto-recall: pull
BM25 plus vectors fused by reciprocal rank (identifiers vs paraphrase); a small
model as relevance judge (retrieval proposes, the judge decides); local embedder
(memory holds family and financial material); no vector database product.
**Scope flag:** this section is a second mechanism with its own retriever,
judge, corpus and measurement. It may be its own chapter (see open questions).

### §16.10 Memory is data
The defense is the channel: compressor output and session memories commit with
`Channel = Data`, never instruction. The grader plants text that reads as an
instruction and checks it renders as data. Identity-band write protection is the
sandboxing chapter's (ruled, design doc §B.10); name the MEMORY.md curator there as the one
sanctioned automated writer.

### §16.11 The Memory tab
Three band budgets, recall settings (retriever mix, judge model), read-only bytes
per band against budget.

### §16.12 Measure it (mandatory; the measurement chapter was cut)
The chapter prints numbers it produced, with the command:
1. BM25 recall@10 on a labeled query set, before and after adding vectors.
2. Achieved compression ratio per stage vs the 8x target.
3. Bytes per band after a long scripted session, against budget.
4. Cache hit rate across a `save_memory`, to confirm the one-miss-per-milestone
   cost.
If a number cannot be produced, the prose says so; it does not quote a guess.

### §16.13 Taking It for a Spin
Plant a fact, call `save_memory`, watch the conversation vanish and the fact
survive. Change the fact, save again, and check the old value is gone from the
bands. Kill the process mid-compaction and restart: the old band is still live.

## Grader checks (from §B.12, plus what this outline adds)

| Check | Property | Note |
|---|---|---|
| planted-fact-survives | token recalled after `save_memory` deletes the conversation | unfakeable |
| supersession-both-ways | new value answered; old value absent from the bands | |
| bands-within-budget | after a long scripted session no band exceeds its budget | shared with ch15 |
| memory-is-data | text through `save_memory` renders in the data channel, including instruction-shaped text | |
| replay-needs-no-llm | replaying a log containing compaction events makes zero vendor calls and reaches the saved context | **new**; ties to ch11 replay-equals-snapshot |
| crash-mid-compaction | a launch with no finish leaves the old band live after restart | **new**; RCU failure property |
| ch15-parity | chapter 15 still passes | |

Weights set by the coder's P9 audit, not here. The identity-write refusal check
lives in the sandboxing chapter.

## TL;DR contract items (for the author's first pass)

`save_memory` sequence and end-of-turn application; band table and budget
config; `CompactorLaunched` / `MemoryCompacted` shapes and the measure-after-
finish chain; RCU commit rule; `LearningAdded`; the data-channel rule;
recall interface (if recall stays in this chapter). Every name above is a
proposal until Q2 is ruled.

## Open questions blocking the TL;DR

1. **Event names (Q2).** `MemorySaved`, `CompactorLaunched`, `MemoryCompacted`,
   `LearningAdded` are proposals. Alternative from §I.5a: `RedactSummary` plus a
   channel selector, zero new types. Needs a ruling.
2. **Launch with no finish on restart:** relaunch or abandon?
3. **Who chooses what graduates (Q11):** the budget (oldest overflow) or the judge?
4. **Compressor sandbox (Q15):** here or the sandboxing chapter?
5. **Validation gate before commit (Q16), turn-boundary + batching (Q17),
   visibility (Q18):** confirm the derived answers.
6. **Recall (Q19, Q20):** does the judge see the bands; are fragments delivered
   once or persisted?
7. **Scope:** keep auto-recall in ch16, or split it into its own chapter? The
   author leans split: push and pull are separate mechanisms with separate
   graders, and ch16 without recall is already the largest chapter in the arc.
8. **Forced `save_memory` threshold:** ch15 owns the context target (Context
   Management tab); does ch16 own the forcing, or does ch15?
9. **The 0.5→3.0 fact:** ch15 §15.13 or ch16 §16.9, not both.
10. **Fresh start (Q24):** what attaches identity and bands before the first
    event? Probably ch15's, but ch16's grader depends on it.
