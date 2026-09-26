# Brief: Chapter 16 coder (reference solution + grader + review)

You are the CODER for Chapter 16 of *The Self-Wielding Agent*, repo
`~/projects/ensemble`. The AUTHOR (a separate instance) wrote the chapter's
contract; you make it real, grade it, audit the grader, and report every place
the contract is wrong, ambiguous or ungradeable. Bill owns the repo and pushes.

Model: claude-opus-5. Expect a long job. Work in small commits.

---

## 0. What changed since the outline

The outline (`book/chapter-16-outline.md`) predates two design sessions on
2026-09-26 that changed the trigger model and the event shapes. The design
notes (`book/ch16-memory-design-notes.md`) are authoritative where they
conflict with the outline. Four changes matter:

1. **`save_memory` is eliminated as a separate tool.** Everything triggers on
   `micro_handoff`. The outline's §16.2 ("the one door") is replaced by the
   design in §2 of the design notes.
2. **Forced handoff** is narrowing the tool set (design notes §4), not a
   prompt injection. The framework removes every tool except `micro_handoff`
   when context crosses a hard threshold.
3. **Memory is five independently-controlled bands, not one on/off switch.**
   Design notes §13 supersedes §9's `MemoryOff`/`MemoryOn` pair with a
   `Band` enum and two events, `BandPopulated`/`BandDepopulated`, following
   ch15's `Redacted`+`Redaction` precedent (one event type, an enum field
   selecting the variant).
4. **Range identifiers are file-based, not log-based.** Compressed memory
   files are named and referenced by `MemoryFileID{Date, Num}` — the same
   `YYYY-MM-DD-N.md` convention the existing memory system already uses —
   never by event-log `Seq`. The corpus is meant to outlive any one log.

Where the TL;DR (once written) conflicts with either document, **the TL;DR
wins and you report the conflict**; never resolve one silently.

---

## 1. Read first, in this order

1. `book/chapter-16-outline.md`: sections, grader table, open questions.
2. `book/ch16-memory-design-notes.md`: **§13 is the authoritative design**
   for bands, events, and on-disk layout. §§1–8 (trigger model, forcing,
   grain, handoff schema, channel, compressor voice) still stand. §9 is
   superseded by §13 but keep reading it — it has the cost-inversion
   argument (§9's "costs invert" point) which §13 doesn't repeat.
3. `book/chapter-15.md` TL;DR (rules 1–10): the contract ch16 builds on.
4. `agent/internal/common/context.go`: the `Apply` reducer, `Entry`,
   `EntryKind`, the `land` method (micro_handoff stripping), `summarizeSpan`.
5. `agent/internal/common/event.go`: `EventType`, `RedactData`, `Redaction`
   levels (**the precedent for `Band`/one-event-per-verb**), `MicroHandoffData`.
   Also find and read the definition and every use of `Ref` and `BlobPart` —
   **verify whether `BlobPart` already resolves to bytes at render time for
   every vendor path before deciding whether band entries can reuse it.**
   This is flagged OPEN in design notes §13 and is the first thing to
   resolve; it changes whether you need a new Part type.
6. `agent/internal/llm/policy.go`: the ch15 curation policy, `curate()`,
   `ladder()`, `stubUnkeptBatch()`.
7. `agent/internal/common/budget.go`: `BudgetsFor`, `Budgets`.
8. `agent/internal/tools/tools.go`: the `micro_handoff` tool, its schema,
   `toolMicroHandoff`.
9. `book/chapter-writing-procedure.md` and `book/course-policy.md` (P9).

---

## 2. Hard constraints

- Edit only: `agent/`, `solutions/ch15/` (step 3.0), `solutions/ch16/` (new),
  `internal/grade/ch16_*` (new), the ch16 wiring in `cmd/grade` and the
  `Makefile`, and `book/review-ch16.md` (new). **No other book file.** If the
  TL;DR is wrong, say so in the review.
- Additive only. Every earlier chapter's grader must score exactly what it
  scored before your change (see §3.13).
- Commit with
  `git -c user.name='CodeRhapsody' -c user.email='coderhapsody@coderhapsody.local' commit`.
  Stage exact paths. **Never `git add -A`** (Bill keeps untracked files in the
  tree). **Never push. Never move or create tags**; report tag moves needed.
- Always `git --no-pager`; always `go test -count=1`. Never judge an exit code
  through `| head` (it reports head's status).
- Every grader launch runs the agent in its own fresh directory (the
  `freshRunDir` helper). The agent loads `./save.json` by default; a shared
  cwd makes runs load each other's conversations.
- Never run a grader while `internal/grade` is mid-edit; take baselines in a
  clean `git worktree` at your starting commit.

---

## 3. Work plan

### 3.0 Baseline and re-snapshot `solutions/ch15`

Confirm `solutions/ch15` scores 100/100 on `make grade15`. If it does not,
re-snapshot from `agent/`. Commit if changed; report that the
`ch15-solution` tag needs moving; do not move it.

### 3.1 The band types

```go
// Band names a memory tier. Each gets its own settings row: enabled or not,
// a budget, and — for the four that cascade — high/low watermarks. BandSoul
// is the one exception: no cascade, no watermarks, only enabled/disabled
// and refresh-on-edit.
type Band uint8

const (
    BandSoul    Band = iota + 1 // SOUL.md — identity, no cascade
    BandMemory                  // MEMORY.md — curated, terminal
    Band64x                     // 64x compression tier
    Band8x                      // 8x compression tier
    BandSession                 // session memories, from micro_handoff
)

// MemoryFileID identifies one memory file the way save_memory already names
// it on disk: a date and a same-day sequence number. Independent of the
// event log's Seq — the memory corpus outlives any one log.
type MemoryFileID struct {
    Date string `json:"date"` // "2026-09-26"
    Num  int    `json:"num"`  // 1, 2, 3... within that date
}
```

New `EntryKind` values: `KindSoul`, `KindMemory`, `Kind64x`, `Kind8x`,
`KindSession`. Five specific kinds, not one parameterized `KindBand{Band}` —
ch15's removal rule wants each kind checkable in the tool-clearing path the
same way `KindSkill`/`KindTools` already are.

### 3.2 The two events

```go
// BandPopulated adds one immutable unit to a band. Carries a Ref, never
// embedded text: startup load, graduation output, and re-enable all resolve
// to the same on-disk file.
type BandPopulatedData struct {
    Band   Band         `json:"band"`
    Ref    Ref          `json:"ref"`
    Thru   MemoryFileID `json:"thru,omitempty"` // graduation/curator only
    Source string       `json:"source"` // "startup","compaction","graduation","curator","reenable"
}

// BandDepopulated retires everything in a band at or before Thru. For
// "disabled" it names nothing further: the reducer wipes every live entry of
// that Kind, computed at apply time.
type BandDepopulatedData struct {
    Band   Band          `json:"band"`
    Thru   *MemoryFileID `json:"thru,omitempty"` // nil + reason=disabled means all
    Reason string        `json:"reason"` // "graduation", "disabled"
}
```

A graduation is one `BandDepopulated` + one `BandPopulated`, both naming the
same `Thru`. Disabling a band is one `BandDepopulated` with no `Thru`.
Re-enabling is one `BandPopulated` per file found on a directory rescan,
`Source:"reenable"`.

**Open detail, yours to resolve:** `BandPopulated` for `BandSession`
(`Source:"compaction"`) is sourced from a conversation Seq range, not a prior
memory file, so `Thru` (a `MemoryFileID`) doesn't quite fit. Either add
`FromSeq`/`ToSeq common.Seq` fields used only for this source value, or let
the newly-created file's own identity double as `Thru`. Document your choice
in the review.

### 3.3 On-disk layout

```
memory/                # session band — one permanent file per micro_handoff-
  2026-09-26-1.md        triggered compression, same directory + naming
  2026-09-26-2.md        save_memory already uses (never invent a second
                          numbering scheme over the same directory)
memory/bucket-0/        # 8x band — one immutable file per graduation,
  2026-09-20-1_2026-09-26-3.md   named by the file range it spans
memory/bucket-1/        # 64x band — same idea, coarser range
MEMORY.md               # terminal — the one band with a real writer; the
                          curator edits it in place, unlike the other four
SOUL.md                 # identity — read once at startup if BandSoul enabled
```

Bucket-0/bucket-1 files are retired (deleted, not archived) once folded
upward — matching the existing shipped cascade's "remove the oldest
sections... discarded." Uncompressed session files in `memory/` are never
retired; they stay on disk forever.

### 3.4 Fresh start

At construction, if `BandSoul.Enabled`, read `SOUL.md` once (if present) and
emit `BandPopulated{BandSoul, Ref: ..., Source:"startup"}`. Same for
`BandMemory`/`MEMORY.md`. This is the whole mechanism — no special-casing
elsewhere, because every later band arrives through the identical event.
Resolves the ch15 outline's open question 5 (what attaches identity/memory
bands before the first event).

### 3.5 micro_handoff triggers session-band population

1. `micro_handoff` strips tool calls and results (already works — the `land`
   method in `context.go`). The handoff note becomes a `KindHandoff` entry
   (already works).
2. After stripping, **measure the conversation band** (all `KindDialogue`
   entries). If below threshold, done.
3. If above threshold, **select one or more whole micro_handoff segments**
   (the dialogue between consecutive `KindHandoff` entries), oldest first.
   Never split a segment. Batch adjacent segments if one alone is too small.
   Target roughly 2x reduction.
4. Compress the selected segments via a direct LLM call (see §3.9) into a new
   session-band file, written to `memory/YYYY-MM-DD-N.md`.
5. Record the compression as two events, not one: an ordinary `Redacted`
   event with `Level: RedactSummary` collapsing the selected conversation
   span (ch15's existing mechanism — `summarizeSpan` in `context.go` already
   folds a span into one entry), and a `BandPopulated{BandSession, Ref: new
   file, Source:"compaction"}` landing the compressed content as its own
   entry.

   **The wrinkle you need to resolve:** `summarizeSpan` today always lands
   its replacement as `Kind: KindDialogue`. For ch16, the replacement needs
   `Kind: KindSession` instead, so it survives tool clearing and future
   redaction the way `KindSkill`/`KindTools` entries already do. This likely
   means a small change to `summarizeSpan` (or a `RedactData` field saying
   which kind the replacement should land as) rather than reusing it
   unmodified. Report what you changed and why in the review; if the TL;DR
   doesn't already describe this interaction between `RedactSummary` and the
   band kinds, say so — it needs to.

The compressor writes **in first person** ("I found..."
not "CodeRhapsody found..."). It receives only the stripped conversation text
(no tool calls, no tool results, no thinking) — information-neutral, since
thinking is already gone from stripped conversation.

The handoff document itself (the `micro_handoff` tool's `text` argument)
stays **one string** with suggested sections in the tool description (goal,
done, findings, env, plan, tried_and_failed), not required parameters.

### 3.6 The forcing mechanism

When context crosses configurable thresholds:

1. **Warn threshold** (~90% of the model's context window): inject a system
   message (`ActorSystem` via `MessageReceived`) telling the agent to
   checkpoint soon.
2. **Force threshold** (a hard limit below the real window, leaving headroom
   for the handoff document + thinking): emit a `ToolsChanged` event removing
   every tool except `micro_handoff`.

Add `ContextWindow int` (tokens) to `ModelFeatures`. Warn at 90% of it, force
at 95%, compared against cumulative `Usage.Input + Usage.CacheWrite`. Applied
in the same place `curate()` runs, before sending the next request.

### 3.7 The graduation cascade

After a `BandPopulated{BandSession}` lands:

1. Measure all `KindSession` entries. If total bytes exceed `SessionBand`'s
   high watermark, launch the 8x compressor.
2. The compressor receives the **oldest** session files (enough to bring the
   band to its low watermark once removed) and produces one compressed file
   in `memory/bucket-0/`, named by the file range it spans.
3. On completion: `BandDepopulated{BandSession, Thru: X, "graduation"}` +
   `BandPopulated{Band8x, Ref: new file, Thru: X, "graduation"}`. Then
   measure `Band8x`; if over its high watermark, launch the 64x compressor
   the same way. Then `Band64x` → the MEMORY.md curator, same pattern,
   `Source:"curator"`.

**Crash recovery:** emit `CompactorLaunched{Band, Thru}` when a compressor
starts (observable, never replayed as work — exists so a GUI can show a
spinner). If the log has a `CompactorLaunched` with no matching
`BandPopulated` carrying the same `Band`+`Thru` when the agent starts, the
launch is abandoned. The old band stays live untouched. The next
`micro_handoff` measures again and relaunches if still over.

### 3.8 Band settings

```go
type BandSettings struct {
    Enabled       bool `json:"enabled"`
    Budget        int  `json:"budget,omitempty"`
    HighWatermark int  `json:"high_watermark,omitempty"` // 0 = 2x budget
    LowWatermark  int  `json:"low_watermark,omitempty"`  // 0 = 1x budget
}
```

Five instances in `Settings`: `SoulBand`, `MemoryBand`, `Band64x`, `Band8x`,
`SessionBand`. Each is a row in the memory settings tab. `SoulBand`'s
watermark fields are present but unused — flag in the review whether this
reads as a wart or is acceptable (design notes §13 open item 8).

**Disabling a band whose downward neighbor is enabled and over budget is a
configuration error, not silent behavior.** If `Band8x` is disabled while
`BandSession` crosses its high watermark, refuse the graduation and surface a
warning (a log line is sufficient for this chapter; a GUI toast is not
required). Do not silently drop the compression or let the band grow
unbounded.

### 3.9 The compressor implementation

Bounded input (one band's worth of files), bounded output (input ÷ ratio), no
tools, no memory of its own. Use a **direct LLM call** — construct a one-shot
prompt with the content to compress, send it to the vendor, use the response
as the compressed text. Do not build a sub-agent framework for this; that is
the sub-agents chapter's job, and it will upgrade this call site later
(design notes §14 item 7, ruled: sub-agents chapter comes after ch16).

Compressor prompt should instruct: write in first person; preserve specific
facts, numbers, file paths, commit hashes, and decisions; drop narrative
rationale that regenerates from context; target the specified compression
ratio.

### 3.10 Updated micro_handoff tool description

```
"Checkpoint. Write a note to your future self covering: the goal as you
understand it; what is done; what you learned (with file:line citations);
your environment (working directory, which command lies, where the logs
are); what you tried that failed; what is next. Every tool call and tool
result so far is then removed from your context and replaced by this note.
Call it when a sub-task is finished and nothing is in flight."
```

Tool handler: (1) record `MicroHandoff` (already done); (2) after the batch
flushes, measure the conversation band; (3) if above threshold, select
segments and run the compressor; (4) record the resulting events; (5) check
the graduation cascade. Steps 2–5 run in the engine's post-turn hook, not
inside the tool handler — the handler returns immediately ("checkpoint
recorded").

### 3.11 Snapshot `solutions/ch16`

`make grade-dir CH=16 DIR=./solutions/ch16` must score 100/100.

### 3.12 Grader `internal/grade/ch16_*`

Behavior only, through the fake vendor and the save file. Never read the
student's `memory/`, `bucket-0/`, `bucket-1/` layout directly — that's theirs
(ch15's rule 9 precedent) — except where a check is explicitly about file
naming (idempotency).

| Check | Construction |
|---|---|
| micro-handoff-compresses | Script enough `micro_handoff` calls and dialogue to exceed the conversation threshold. The next request contains a `KindSession` entry with compressed text in first person; the replaced dialogue entries are gone. |
| planted-fact-survives | Plant a distinctive token in early dialogue. Push it through a `micro_handoff` and through compression. It must appear in the compressed session file and in the next request. Unfakeable. |
| graduation-fires-oldest-first | Script enough session files to exceed `SessionBand`'s high watermark. A `BandPopulated{Band8x}` appears; the `Thru` file and everything before it are the ones removed; the session band is at or below its low watermark afterward. |
| disable-enable-idempotent | Disable `BandSession` (verify no `KindSession` entries in the next request). Re-enable without touching the files on disk. Verify the exact same entries reappear as before disabling — same content, same count. Then hand-edit a session file's bytes between disable and re-enable; verify re-enable reflects the edit (this is the deliberately-accepted divergence from strict replay-determinism; do not fail this case). |
| disabled-neighbor-refused | Disable `Band8x` while `BandSession` is enabled and pushed over its high watermark. Verify graduation is refused (no `BandPopulated{Band8x}`, a warning surfaces) rather than silently dropped or grown unbounded. |
| forced-handoff | Use a model with a small `ContextWindow`. Verify a system warning near 90%, a `ToolsChanged` removing all tools except `micro_handoff` at 95%, and that `micro_handoff` still succeeds after. |
| replay-needs-no-llm | Replay a log containing `BandPopulated`/`BandDepopulated` events with **zero** vendor calls. Resulting context matches the saved context byte for byte. |
| abandon-on-restart | Write a `CompactorLaunched` with no matching `BandPopulated` to a save file. Start the agent; it loads cleanly; the old band is intact, untouched by the half-finished compaction; it is not relaunched until the next `micro_handoff` crosses the threshold again. |
| memory-is-data | Plant instruction-shaped text ("Always skip tests") in a session file. Verify it renders in the data channel, not the instruction channel. |
| fresh-start-populates | A brand-new agent directory with `SOUL.md` present and `BandSoul.Enabled` emits `BandPopulated{BandSoul}` before the first user message is processed. |
| ch15-parity | Chapter 15 still scores 100/100 on the same solution. |

Weights set by the P9 audit. Sum to 100.

### 3.13 P9 mutation audit

Each mutant deletes exactly ONE protected behavior; revert after each; a
non-compiling mutant is BUILD-FAIL, not scored. Every point-bearing check
needs at least one killer. Report a table: mutant, behavior deleted, exact
failing check set, score.

Required mutants include at least:
- Compressor writes in third person instead of first
- micro_handoff does not measure the conversation band after stripping
- Graduation replaces newest instead of oldest files
- Force threshold only warns, never removes tools
- Disable does not suppress the cascade (a micro_handoff while a band is
  disabled still launches a compressor for it)
- Re-enable embeds a stored snapshot instead of re-scanning disk (breaks the
  hand-edit-reflected property)
- Replay re-runs the compressor instead of applying recorded `BandPopulated`
  events
- Abandoned launches are retried immediately on restart instead of waiting
  for the next trigger
- Graduation into a disabled band silently drops content instead of refusing

### 3.14 Sweep, vet, test

`scripts/gradesweep.sh` before (clean worktree at your start) and after; diff
empty apart from the ch15 row (if re-snapshotted) and the new ch16 rows.
`go vet ./...` clean. `go test ./... -count=1`: prove any failure is
pre-existing with a worktree at your starting commit, name for name.

### 3.15 `book/review-ch16.md`

1. Every TL;DR problem: contradiction, ambiguity, an untestable rule, a rule
   the reference cannot meet. Quote the TL;DR verbatim; propose replacement
   wording.
2. Design-doc conflicts, with file:line.
3. Outline conflicts (the outline predates both design sessions).
4. Resolution of the two open items handed to you: `BlobPart` reuse (§1
   item 5) and the session-band `Thru`/`FromSeq` question (§3.2).
5. The mutation table and final weights.
6. Measurements: bytes per band after a scripted multi-`micro_handoff`
   session; how many cascade stages fired; compression ratios achieved vs.
   2x/8x/64x targets; context size before/after the ladder + band system
   together. Label anything unmeasured as unmeasured.

## 4. Report

Submit via the `submit` tool: status, commits, ch16 score, other grader
scores (before/after sweep), mutation table, TL;DR problems, the two
resolved open items, open questions. `submitted` is a claim; the author
re-runs the grade.

If you need a decision you cannot infer, use `send_message_to_parent`; the
brief authorizes reporting a bad TL;DR rule in the review instead of
deviating silently, which covers most ambiguity.
