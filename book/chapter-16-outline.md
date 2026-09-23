# Chapter 16 outline: memory (Chapter B of the context arc)

Source of truth: `book/ch15-context-engineering-design.md`, Part III, plus the
evening rulings at the end of that doc (they override Part III where they name
it). Builds on Chapter 15's channels, entry kinds, compaction events and total
reducer, and on Chapter 11's anchored save. Where this outline and the design doc
disagree, the design doc wins until Bill rules; flag, never resolve silently.

Workflow (same as ch15): design doc → author writes TL;DR + §16.1 plain words →
coder builds `solutions/ch16` and writes `book/review-ch16.md` → author revises
TL;DR + plain words against the review → author writes the body.

## Rulings (Bill, 2026-09-22 late)

1. **Auto-recall is its own chapter.** Ch16 is push only (the bands); pull
   (BM25 + vectors + relevance judge) moves out. Number TBD.
2. **Event names stand:** `MemorySaved`, `CompactorLaunched`,
   `MemoryCompacted`, `LearningAdded`.
3. **A launch with no finish at restart is abandoned.** The old band stays
   live; the next `save_memory` measures again and relaunches if still over.
4. **Graduation is oldest-first and watermark-based, not one memory at a
   time.** DERIVED reading, confirm: when a band crosses its high watermark,
   the oldest entries graduate together as one batch until the band is below
   its low watermark. (The shipped CodeRhapsody cascade moves one file at a
   time.) Open detail: may one session memory be split across batches, or do
   entries move whole?

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

The three auto-recall facts (the 0.5→3.0 threshold, the ~5.6 MB corpus, the
unmeasured recall@10) move to the auto-recall chapter with ruling 1.

## Voice plan

- Third person throughout the body. Short first-person opener allowed.
- **Bill moments (≤2):** (1) the opener if Bill supplies one; (2) the
  watermark ruling (ruling 4), if it earns its place. The "Literature? We're
  good" vector-database moment moves to the auto-recall chapter.
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
request, governed by byte budgets; this chapter. **Pull:** auto-recall, fetched
per turn by relevance; its own chapter (ruling 1). Conflating them is how the
shipped system got confusing (§B.2). One
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
Ruling 7 of the design doc, plus rulings 3 and 4 above. `save_memory` measures
the session band → launches the 8x compactor if over its high watermark → on
finish, measure 8x → 64x compactor → on finish, measure 64x → MEMORY.md
curator. **What graduates is the oldest content, as one batch, down to the low
watermark** (ruling 4): the gap between the two watermarks is what keeps the
chain from firing on every save. Two events per stage:
`CompactorLaunched{Band, AsOf}` (observable, never replayed as work) and
`MemoryCompacted{Band, Replaces []Seq, Text}` (the mapping the reducer
applies). **Replay never re-runs an LLM.** A launch with no finish at restart
is abandoned (ruling 3).

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

### §16.9 (moved) Auto-recall
Moved to its own chapter by ruling 1. Source for that chapter: design doc §B.9,
Q19, Q20, and the three facts listed under the wild-fact table.

### §16.10 Memory is data
The defense is the channel: compressor output and session memories commit with
`Channel = Data`, never instruction. The grader plants text that reads as an
instruction and checks it renders as data. Identity-band write protection is the
sandboxing chapter's (ruled, design doc §B.10); name the MEMORY.md curator there as the one
sanctioned automated writer.

### §16.11 The Memory tab
Per band: budget, high and low watermark; read-only bytes per band against
budget.

### §16.12 Measure it (mandatory; the measurement chapter was cut)
The chapter prints numbers it produced, with the command:
1. Achieved compression ratio per stage vs the 8x target.
2. Bytes per band after a long scripted session, against budget and watermarks.
3. How often the chain fires per `save_memory` over that session (the watermark
   gap is supposed to make this rare).
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
| crash-mid-compaction | a launch with no finish leaves the old band live after restart, and is not relaunched until the next `save_memory` | **new**; RCU failure property + ruling 3 |
| oldest-first-graduation | when a band crosses its high watermark, the entries replaced are the oldest ones and the band ends below its low watermark | **new**; ruling 4 |
| ch15-parity | chapter 15 still passes | |

Weights set by the coder's P9 audit, not here. The identity-write refusal check
lives in the sandboxing chapter.

## TL;DR contract items (for the author's first pass)

`save_memory` sequence and end-of-turn application; band table with budget and
watermarks; `CompactorLaunched` / `MemoryCompacted` shapes and the measure-
after-finish chain; oldest-first batch graduation; abandon-on-restart; RCU
commit rule; `LearningAdded`; the data-channel rule. Event names are ruled.

## Open questions blocking the TL;DR

1. **Ruling 4 detail:** do entries graduate whole, or may one session memory be
   split across batches? And confirm the high/low watermark reading.
2. **Compressor sandbox (Q15):** here or the sandboxing chapter?
3. **Validation gate before commit (Q16), turn-boundary + batching (Q17),
   visibility (Q18):** confirm the derived answers.
4. **Forced `save_memory` threshold:** ch15 owns the context target (Context
   Management tab); does ch16 own the forcing, or does ch15?
5. **Fresh start (Q24):** what attaches identity and bands before the first
   event? Probably ch15's, but ch16's grader depends on it.
6. **Auto-recall chapter number:** after ch16 (so goal stack becomes ch18), or
   after the goal stack?

Resolved by the late rulings: event names (Q2), launch-without-finish,
who picks what graduates (Q11), recall scope, the 0.5→3.0 conflict (it goes to
the auto-recall chapter, so ch15 §15.13 should drop it or cite forward).
