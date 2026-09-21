# Review: the chapter 14 grader

*Written 2026-09-20 by the coder, for the author. Companion to
`book/brief-ch14-coder.md`.*

Everything below was measured, not recalled. Commands are given so any claim can
be re-run.

## What shipped

| artifact | what it is |
|---|---|
| `internal/grade/ch14_driver.js` | The harness. Loads the student's browser modules under node against a stubbed `speechSynthesis`, feeds fixed fragment sequences, reports what reached the channel. |
| `internal/grade/ch14_harness.go` | Finds node, discovers the student's modules, runs the driver, transcribes the JSON into `Ch14Result`. |
| `internal/grade/ch14_checks.go` | The seven checks, 100 points. |
| `cmd/grade/main.go` | `case 14`, including the chapter 13 parity run. |
| `Makefile` | `make grade14`. |
| `scripts/ch14-mutations.sh` | The P9 audit. Not part of a normal test run. |
| `solutions/ch14` | The snapshot. |

**Scores.** Reference `./agent` 100/100. Snapshot `./solutions/ch14` 100/100,
graded standalone.

## The grader does not read the reference

Module discovery is by content, and each module's name is read out of the
student's own source (`ch14Discover`, `ch14_harness.go`). A student who renames
`tts.js`, or calls the object something other than `TTS`, still grades. The
speech pipeline is found by looking for a module with a chunk entry point, a
`flush`, and a `SpeechSynthesisUtterance`.

The gate check goes further and never hardcodes an element id: the input box is
whichever element registered an `input` listener.

The observation point is the `speechSynthesis` stub itself, not any field inside
the student's module. That is deliberate. It is literally the channel under
test, and it means no check depends on an internal name like `transcript`.

## Mutation audit (P9)

`bash scripts/ch14-mutations.sh`, baseline 100/100. Eight mutants, **all
caught, no stale mutants.**

| mutant | deleted behaviour | score | failing checks |
|---|---|---|---|
| M1 | the buffer; fragments enqueue raw | 45 | buffers, filters, boundaries |
| M2 | whole-buffer fence resolution | 80 | filters |
| M3 | a lone newline is a space | 85 | boundaries |
| M4 | `flush()` speaks the remainder | 20 | buffers, filters, boundaries, unstreamed, identifiers |
| M5 | the ordinary-word guard in identifier expansion | 90 | identifiers |
| M6 | the `accumulated` guard, doubling direction | 85 | unstreamed |
| M7 | the `accumulated` guard, silence direction | 85 | unstreamed |
| M8 | the edge comparison in `updateGate` | 90 | gate |

Six mutants fail exactly one predicted check. **M1 and M4 fail wider, and that
is correct rather than sloppy**: the buffer and the flush sit upstream of every
other scenario, so deleting either one removes the substrate the other checks
observe. A narrower failing set would mean the other checks were not really
reading the channel.

M6 and M7 are the same guard mutated in both directions, as the brief required.
Testing only the silence direction would pass an implementation that speaks
every finalised part twice.

## Decisions taken

**Open question 1, gate testing: node, as recommended.** It turned out not to be
a compromise. `gui.js` assigns `TTS.onStateChange = updateGate`, so the gate
closure is reachable from outside the IIFE. The check drives the real `gui.js`
under a stub DOM and asserts on the actual `{"type":"pause"}` / `unpause`
frames, a wire shape chapter 8's grader already pins. It covers rules 10, 11 and
12 in one sequence: pause on speech, no second frame when the other cause
arrives, **no unpause while a single space remains in the box**, unpause only
when both clear, then silence on an unchanged value.

**Open question 2, the persona: not implemented. See the gap below.**

**ALL CAPS is now a fixture.** On Bill's steer during the build. `ALL CAPS WORDS
ARE FINE.` must survive untouched, because an implementation that inserts a
space before every capital passes `camelCase`, `HTTPServer` and `RPGLit` and
then spells ordinary acronyms out letter by letter. That is mutant M5, and this
fixture is what catches it.

## Gaps the author should know about

### 1. The blind persona is not graded at all

Rules 13, 14 and 15 are half the chapter's stated deliverable, and **a student
can skip the persona entirely and still score 100/100.** The check table in the
brief totals 100 without a persona check, while open question 2 asks how to
build one; those two cannot both be satisfied.

Recommendation: add `persona-denies-dom` at 10 points, structural, derived from
the student's own tree the way ch5's `framework-blind` check works. Suggested
rebalance: `tts-expands-identifiers` 10 to 5 and `ch13-parity` 10 to 5. That
leaves the two protected weightings untouched.

**This needs Bill's ruling before it is built.**

### 2. The TL;DR does not name the entry point

House rule is that a fresh coder scores 100/100 from the TL;DR alone. The TL;DR
names `flush()` (rule 3) and nothing else. It never names the chunk entry point,
and `queueChunk` appears only later, in the body's code blocks.

The grader therefore has to assume a name it was never promised. Discovery
softens this (the object's name is read from the student's source) but the
method names cannot be discovered. One line in the TL;DR naming the required
surface would close it.

### 3. Rules with no check

| rule | status |
|---|---|
| 8, speak errors | **not checked.** `speakError` exists and is a verified feed site, but no check asserts an error reaches the channel. |
| 9, record a transcript | **not checked**, deliberately. Checking it would couple the grader to an internal field name. |
| 13, 14, 15, the persona | **not checked.** See gap 1. |

Rule 8 is cheap to add and worth it: an error that is displayed and not spoken is
the chapter's own example of a channel failure.

## A seventh defect, found while building the fixtures

The chapter documents six defects. There is a seventh, still present:

```
input : "- first item\n- second item"
spoken: "first item - second item"
```

`queueChunk` flattens lone newlines to spaces **before** `_speakable` runs, and
the bullet regex is `^`-anchored with `/gm`. After flattening there is only one
line, so every bullet except the first survives and is read aloud as "dash". A
normal bulleted list is heard with the dashes in it.

This is the same shape as defect 5 in the chapter, the shattered fence: **a
filter anchored to line structure, running after line structure was destroyed.**
That is a better generalisation than the chapter currently draws from one
instance, if the author wants it.

Fixing it is a solution change and therefore Bill's call, not the coder's. The
grader currently tests bullets at a paragraph boundary, where the reference is
correct. If the defect is fixed, tighten the fixture in
`tts-filters-markup` to the single-newline case.

## Chapter claim verified

§14.2 says four call sites feed the channel, all in the artifact renderer.
Confirmed in `agent/web/gui/artifact-scroll.js`: deltas (76), part final (98,
106), tool dispatch (148, 149), errors (193). The only other entry point in the
tree is `gui.js:150`, a `TTS.cancel()`, which drains the channel rather than
feeding it. **No feed site was added or moved while building the grader, so the
chapter remains accurate.**

## Pre-existing failures, not from this work

Proven by running the identical commands in a `git worktree` at HEAD `fc3be34`
with none of this work present, and getting the identical result.

- `go test ./...` fails: `TestCh3PointsSumTo100` ("got 10 checks, want 9"),
  `TestCh6ReferenceScores100`, `TestCh6DeletionAudit`,
  `TestCh7ReferenceScores100`, `TestCh7DeletionAudit`, `TestCh8Grade`. The
  failing set at baseline is identical to the failing set now.
- `make grade2` scores 90/100 and `make grade3` scores 85/100 at baseline. The
  brief's definition of done expects both to pass.
- Chapter 6's reference scores 0/100, which suggests a missing or moved
  snapshot rather than a subtle regression.

Everything else passes: **ch5, 7, 8, 9, 10, 11, 12, 13 and 14 all 100/100.**

## Snapshot hygiene

`solutions/ch14` is 580K across 64 files. It excludes the `ensemble` and
`virtual-user` binaries, `bin/`, logs, `*.jsonl` and `settings.json`.

For contrast, `solutions/ch13` tracks two 10MB binaries and a `settings.json`
carrying a personal `tts_speed`. The brief flagged this as pre-existing and out
of scope, so it was not touched, only not extended.

## Prerequisite

The grader needs `node` on PATH, per Bill's ruling. Absence produces a named
prerequisite error on every check rather than a silent skip. Verified against
node v24.20.0.

One node detail worth recording, because it cost time: recent node versions
define `navigator`, `fetch` and `localStorage` as getter-only globals, so a plain
assignment throws `Cannot set property`. The driver installs every stub through
`Object.defineProperty`.
