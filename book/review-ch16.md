# review-ch16.md — the coder's report on Chapter 16

Written by the coder after building the reference implementation, the grader
and `solutions/ch16`. Everything below is either VERIFIED (I observed it) or
labelled UNMEASURED. Nothing here is reconstructed from memory.

---

## 1. The chapter does not exist yet

There is no `book/chapter-16.md`. The brief asks me to report "every TL;DR
problem", and §3.15 assumes a TL;DR to check the build against. There is
nothing to check. What exists is `brief-ch16-coder.md`, `chapter-16-outline.md`
and `ch16-memory-design-notes.md`.

So this review reports against the **outline and the design notes**, and the
usual direction of travel is reversed: rather than reconciling prose to code,
the list below is what the prose will have to say when it is written, because
it is what the code now does.

---

## 2. Where the outline and the shipped code disagree

`chapter-16-outline.md` predates Bill's later rulings and disagrees with the
brief in several places. The brief is the coder's contract, so I followed the
brief; each conflict below needs an author decision before the body is written.

### 2.1 `save_memory` is no longer the door (outline §16.2, line 86)

The outline's spine is "`save_memory`: the one door", with the cascade started
only by that tool. Bill ruled the opposite: **`micro_handoff` is the sole
trigger**, and `save_memory` stops being an agent-called tool at all. The
shipped code has no `save_memory`; compaction is measured after a checkpoint.

This is not a detail. The outline's §16.2, its TL;DR contract list (line 192)
and three of its grader checks are all phrased around `save_memory`.

**Why the new rule is better, for the prose:** two tools that end a context,
only one of which feeds the cascade, is exactly the shipped CodeRhapsody bug
the outline itself cites at line 60 (`handoff_task` never fired the cascade,
first seen 2026-04-18, still unfixed). One trigger makes that bug
unrepresentable rather than merely fixed.

### 2.2 Event names (outline ruling 2, lines 17-18)

The outline rules the names `MemorySaved`, `CompactorLaunched`,
`MemoryCompacted`, `LearningAdded`. The shipped events are
`BandPopulated`, `BandDepopulated` and `CompactorLaunched`.

The rename is not cosmetic: the band events describe *a band gaining or losing
a file*, which is what makes switching a band off, editing a file by hand and
switching it back on expressible at all. `MemorySaved` names an act of the
agent; `BandPopulated` names a change to the context. Only the second can be
replayed by a reducer that cannot read a disk.

`LearningAdded` is **not built** — see §6.

### 2.3 Grader checks the outline asks for that I did not build

The outline's table (lines 176-181) names four checks that are not in the
brief's list of eleven, and I built the brief's list:

| Outline check | Status |
|---|---|
| `supersession-both-ways` | **not built** — no check asserts the old value is *absent* after it is superseded |
| `bands-within-budget` | **not built** — nothing asserts a band ends under its budget |
| `crash-mid-compaction` | built, as `abandon-on-restart` |
| `oldest-first-graduation` | built, as `graduation-fires-oldest-first` |

The first two are real gaps and I recommend adding them; see §5.3, where the
mutation audit independently found the same hole from the other direction.

### 2.4 Forcing (outline open question 4, lines 201-202)

The outline leaves open whether ch16 or ch15 owns forcing. The brief assigned
it to ch16 and it is built here: a warning at 90% of the model's context
window and, at 95%, every tool but `micro_handoff` removed.

---

## 3. Where the design notes were wrong, and the rulings that fixed them

### 3.1 §13 said a populate event carries a `Ref`, never embedded text

This cannot work, and the reason is structural rather than a matter of taste:

- The reducer is **pure**. `internal/common/context.go` imports exactly
  `encoding/json` and `fmt`. It has no filesystem and no host. A `Ref` has
  nowhere to become text.
- The renderers cannot resolve one either. `os.ReadFile` and `os.Open` appear
  **nowhere** in `internal/llm/` or `internal/vendor/`, and
  `geminiFileParts` refuses on purpose, because "resolving one into a uri is
  an upload, which is a job for the layer that owns the bytes".

So a Ref-only event would never reach the model. Bill ruled §13 wrong; events
carry bytes, and §13 is corrected in the notes. Chapter 15 had already set the
precedent in code: `summarizeSpan` stores its replacement, "because only here
is the new content something an LLM wrote and nobody can recompute".

### 3.2 The wider "purge Ref" was narrowed, correctly

Bill's first instinct was to purge the concept of `Ref` entirely. `Ref` is
ch02 and ch15 shipped, graded behaviour — `RedactedPart{Stub, Ref}` at
`part.go:117`, and an entire grader file `internal/grade/ch02_ref_checks.go`.
The narrowed ruling is the right one and is what the code now satisfies: a stub
that **tells** the model where full output lives (`cr/io/24`) is helpful prose
and stays; what is banned is any ref the emitter or renderer must
**dereference**. I audited for those. There are none.

### 3.3 The two open items the notes handed the coder, both now closed

**Whether a band entry's `Parts` should reuse `BlobPart{Ref}` — no, and it
could not have.** A non-empty `r.Blobs` returns a hard error on Claude
(`claude.go:251`) and on OpenAI (`openai.go:136`), and `geminiFileParts`
accepts only an already-remote `RefURI`, erroring on a local path. `BlobPart`
is constructed in exactly one place, `part.go:275`, which is wire
unmarshalling — nothing in the agent ever produces one. Memory carried that
way would crash the request on two vendors out of three. Band entries are
`TextPart`, which every renderer already handles.

This deserves a line in the chapter rather than living as an implementation
detail, because it is the concrete reason the design note's original
"carries a `Ref`, never embedded text" could not work. The reducer is pure
and imports only `encoding/json` and `fmt`; the renderers dereference
nothing; so there is no point anywhere in the pipeline where a `Ref` would
have become bytes a model could read.

**Session-band `Thru` versus `FromSeq`/`ToSeq` — neither.** The new file's
own identity carries what is needed, and `Thru` is left to the cascade,
where it genuinely means "the newest source folded into this one." Putting
`FromSeq`/`ToSeq` on a memory would bury an event-log coordinate inside a
corpus meant to outlive any particular log. Which span of conversation a
memory came from is already answered by the `Redacted` event that removed
that span, in the only coordinate system where the question means anything.

---

## 4. The brief's one defect

§3.5 hands the coder a "wrinkle you need to resolve": `summarizeSpan` lands its
replacement as `KindDialogue`, and for ch16 it must land as `KindSession`,
"likely meaning a small change to `summarizeSpan`, or a field on `RedactData`
saying which kind the replacement should land as."

**Do not do this.** It would put the compressed text in the context twice: once
as the span's replacement and once as the band entry. The correct construction
needs no change to `summarizeSpan` at all, because ch15 already built the
behaviour:

```go
if placed || len(r.Replacement) == 0 {
    continue // the span collapses; only the first survivor is emitted
}
```

A `RedactSummary` with an **empty** replacement deletes the span and lands
nothing. So compaction emits two events with disjoint jobs: `Redacted` removes
the conversation, `BandPopulated` adds the memory. `summarizeSpan` is untouched.

---

## 5. The mutation audit

Nine mutants, each deleting exactly one behaviour the chapter promises. Full
detail in `book/ch16-mutation-audit.md`.

| Mutant | Behaviour deleted | Score | Checks that failed |
|---|---|---|---|
| M1 | compressor writes in the third person | 95 | `memory-is-data` |
| M2 | never measure after a checkpoint | 23 | 8 checks |
| M4 | forcing warns but never removes tools | 88 | `forced-handoff` |
| M7 | replay re-runs the compressor | 83 | `disable-enable-idempotent`, `fresh-start-populates` |
| M9 | graduate into a disabled band | 92 | `disabled-neighbour-refused` |
| M5 | a disabled band still compacts | 88 | `disable-enable-idempotent` |
| M6 | re-enable replays a snapshot instead of re-reading disk | 88 | `disable-enable-idempotent` |
| M3 | graduation folds the newest memories | 100 | **equivalent — see below** |
| M8 | an abandoned graduation is retried on restart | 100 | **open hole** |

Five died on the first pass. Four survived, which under P9 makes them holes
in the grader rather than successes. Two of the four now die after
strengthening three checks; the other two are reported rather than papered
over.

**M3 is an equivalent mutant.** Folding the oldest and folding the newest are
the same operation whenever a fold consumes the whole band, and here a fold
always does: graduation triggers on a byte watermark and then folds
`min(FoldFactor, len(files))`. Instrumenting the reference showed folds
firing with five files and taking all five, and the first fold request was
byte-identical under the mutation. Making the difference observable would
mean requiring a full batch before folding, which is the behaviour deleted
earlier in this chapter because a band over budget that cannot fill a batch
would then never fold at all.

**M8 is an open hole, and the reason is worth reading.** The abandon
scenario answers every compressor with a 500, so no memory is ever written,
the session band stays empty, and graduation — the only thing the
abandonment guard protects — is never attempted. The scenario cannot observe
the behaviour it exists for. Closing it needs a compressor that fails folds
while letting session memories succeed; that router flag was written and
then reverted, because a failed fold makes the harness's cumulative request
counts unpredictable and destabilising a passing grader to chase one mutant
was the worse trade.

**One finding that is not about a mutant.** M7 kills two checks but not
`replay-needs-no-llm`, the check named for exactly the property it violates.
That check counts vendor requests over a window that does not include the
startup sync, so it cannot see the call it forbids. A separate hole, not yet
closed.

**Two checks were found to be blind while strengthening.** The disabled-band
check asked the harness whether an extra request had appeared; the harness
counts turns, and a compressor is not a turn, so the compaction it was
looking for was invisible to it. And the hand-edit check originally searched
for a memory by its text, which after folding finds the superseded original
rather than the compressed copy the context is actually showing — editing
the wrong file correctly changed nothing. Both are the same mistake in
different clothes: inferring a thing from a proxy that does not track it.

---

## 6. What is not built

- **`LearningAdded`** (outline §16.8). No learnings channel exists. The brief
  did not ask for one.
- **The Memory tab** (outline §16.11). No GUI work.
- **Auto-recall.** Correctly out of scope: the outline moves it to its own
  chapter (ruling 1), and Bill confirmed the plan is BM25 over the
  uncompressed memories, in a later chapter.

---

## 7. Four real bugs this work exposed

These were found by building the chapter, not by reading it. Each is fixed.

### 7.1 The shipped agent had no memory at all

`eng.Memory` and `eng.Bands` were fields nobody set. Every unit test passed
because tests construct an engine directly; the binary in `cmd/main.go` never
wired either one, so the entire memory system was dead code in the product.
This is the shape worth printing: a feature can be fully built, fully tested
and entirely absent.

### 7.2 A settings file with no `memory` key switched all memory off

`BandSettings.Enabled bool` meant the zero value was "disabled". Any
`settings.json` that did not mention memory deserialized to every band off.
Flipped to `Disabled`, following ch7's `DisableStreaming` precedent: a negative
bool so that the zero value is the behaviour you want.

### 7.3 The band that most needed to fold could never fold

Graduation required a full batch of `FoldFactor` (8) files. A band over budget
holding five large memories would wait forever for a ninth that could only
arrive by growing further past its budget. It now folds what it has.

### 7.4 An edited memory file lost to the copy in the log

`BandHas` matched on file identity alone, so after a restart the log replayed
what a file said when it was first read, and the loader skipped the file as
already present. Editing a memory by hand — the ordinary way to correct
something the agent believes and should not — changed nothing, and the files on
disk were decorative. `BandHas` now compares text as well as identity.

Worth noting for the prose: the check that caught this was written **because**
Bill asked for an out-of-order-restore property. The property found a bug that
was not the property.

---

## 8. Smaller things the author should know

- **Chat mode has no memory.** There are two engine constructions in
  `cmd/main.go`. The server path (`runActorLoop`) is wired; the chat-mode loop
  is not. Deliberate for now — the grader drives the server — but it means
  `--chat` gets no bands.
- **The disabled-neighbour warning repeats every turn.** When a band is full
  and the band above it is switched off, the refusal is logged on every
  checkpoint. It is correct and loud, per the ruling that this must never be
  silent, but it should probably warn once per state change.
- **`ContextWindow` values in the model table are invented.** 200000 / 128000 /
  1000000 are plausible table data for the course's fake models. They are
  **UNMEASURED** and must not be printed as facts about real models.

---

## 9. Measurements

From `agent/internal/llm/measure_test.go`, which runs the real compaction
path against a scripted session: 40 turns, a checkpoint every 4 turns, 400
bytes of talk per turn, conversation budget 3000 bytes (high 6000), session
band budget 1500 bytes (high 3000).

| | memory ON | memory OFF |
|---|---|---|
| conversation | 1,600 bytes | 16,000 bytes |
| session band | 27 bytes | — |
| 8x band | 22 bytes | — |
| **total context** | **1,649 bytes** | **16,000 bytes** |
| memories written | 4 | 0 |
| cascade launches | 1 | 0 |

Cascade launches per checkpoint: 0.10.

**What this measures, and what it does not.** The structural numbers are
real: the conversation is held at 1,600 bytes against 16,000 unbounded, four
memories were written, and one fold fired. That is the property the chapter
claims — a context that stops growing with the length of the session.

The band sizes are **not** a compression measurement and must not be read as
one. The compressor here is a fake that returns a short fixed string, so 27
bytes and 22 bytes are artifacts of the fake's reply length. For the same
reason the headline "memory ON is 10.3% the size of memory OFF" is
**UNMEASURED** as a compression ratio: it is dominated by the fake's reply
size, not by any real summarisation. The 2x / 8x / 64x targets in the design
notes cannot be verified without a real model, and nothing here should be
cited as evidence for them.

What can be said honestly: the ladder fires when it should, folds when it
should, and the conversation stays bounded. The ratios wait for a real
compressor.
