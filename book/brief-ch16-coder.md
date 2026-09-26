# Brief: Chapter 16 coder (reference solution + grader + review)

You are the CODER for Chapter 16 of *The Self-Wielding Agent*, repo
`~/projects/ensemble`. The AUTHOR (a separate instance) wrote the chapter's
contract; you make it real, grade it, audit the grader, and report every place
the contract is wrong, ambiguous or ungradeable. Bill owns the repo and pushes.

Model: claude-opus-5. Expect a long job. Work in small commits.

---

## 0. What changed since the outline

The outline (`book/chapter-16-outline.md`) predates a design session on
2026-09-26 that changed the trigger model. The design notes
(`book/ch16-memory-design-notes.md`) are authoritative where they conflict
with the outline. Three changes matter:

1. **`save_memory` is eliminated as a separate tool.** Everything triggers on
   `micro_handoff`. The outline's §16.2 ("the one door") is replaced by
   the design in §2 of the design notes.
2. **Forced handoff** is narrowing the tool set (design notes §4), not a
   prompt injection. The framework removes every tool except `micro_handoff`
   when context crosses a hard threshold.
3. **Memory on/off** is built in this chapter (design notes §9). The outline
   deferred it.

Where the TL;DR (once written) conflicts with either document, **the TL;DR
wins and you report the conflict**; never resolve one silently.

---

## 1. Read first, in this order

1. `book/chapter-16-outline.md`: sections, grader table, open questions.
2. `book/ch16-memory-design-notes.md`: the 2026-09-26 design session. §§1–12
   are all load-bearing; §13 lists rulings. **This is the authoritative
   design** for the trigger model, the forcing mechanism, and memory on/off.
3. `book/chapter-15.md` TL;DR (rules 1–10): the contract ch16 builds on.
4. `agent/internal/common/context.go`: the `Apply` reducer, `Entry`, `EntryKind`,
   the `land` method (micro_handoff stripping), `summarizeSpan` (ch16's
   compaction uses `RedactSummary`).
5. `agent/internal/common/event.go`: `EventType`, `RedactData`, `Redaction`
   levels, `MicroHandoffData`.
6. `agent/internal/llm/policy.go`: the ch15 curation policy, `curate()`,
   `ladder()`, `stubUnkeptBatch()`.
7. `agent/internal/common/budget.go`: `BudgetsFor`, `Budgets`.
8. `agent/internal/tools/tools.go`: the `micro_handoff` tool, its schema,
   and `toolMicroHandoff`.
9. `book/chapter-writing-procedure.md` and `book/course-policy.md` (P9: every
   grader is audited by deleting one protected behavior from the reference and
   asserting the exact failing set).

---

## 2. Hard constraints

- Edit only: `agent/`, `solutions/ch15/` (step 3.0), `solutions/ch16/` (new),
  `internal/grade/ch16_*` (new), the ch16 wiring in `cmd/grade` and the
  `Makefile`, and `book/review-ch16.md` (new). **No other book file.** If the
  TL;DR is wrong, say so in the review.
- Additive only. Every earlier chapter's grader must score exactly what it
  scored before your change (see §6).
- Commit with
  `git -c user.name='CodeRhapsody' -c user.email='coderhapsody@coderhapsody.local' commit`.
  Stage exact paths. **Never `git add -A`** (Bill keeps untracked files in the
  tree). **Never push. Never move or create tags**; report tag moves needed.
- Always `git --no-pager`; always `go test -count=1`. Never judge an exit code
  through `| head` (it reports head's status).
- Every grader launch runs the agent in its own fresh directory (the
  `freshRunDir` helper). The agent loads `./save.json` by default; a shared cwd
  makes runs load each other's conversations.
- Never run a grader while `internal/grade` is mid-edit; take baselines in a
  clean `git worktree` at your starting commit.

---

## 3. Work plan

### 3.0 Baseline and re-snapshot `solutions/ch15`

Confirm `solutions/ch15` scores 100/100 on `make grade15`. If it does not,
re-snapshot from `agent/` following how earlier snapshots were made (Go's
internal-package rule may require path rewrites). Commit if changed; report
that the `ch15-solution` tag needs moving; do not move it.

### 3.1 The new trigger model: micro_handoff as sole trigger

The design eliminates `save_memory` as a separate tool. Instead:

1. `micro_handoff` strips tool calls and results (this already works — the
   `land` method in `context.go` does it). The handoff note becomes a
   `KindHandoff` entry (this already works).
2. After stripping, **measure the conversation band** (all `KindDialogue`
   entries). If it is below a threshold → done.
3. If above threshold → **select one or more whole micro_handoff segments**
   (oldest first) and compress them via a sub-agent into a memory.
4. The compressed memory replaces the selected segments as a `RedactSummary`
   event (the reducer's `summarizeSpan` already handles this).
5. If uncompressed memories then exceed their band budget → launch the
   bucket-0 compressor (8× target). The cascade continues: if 8× exceeds its
   budget → 64×; if 64× → MEMORY.md curator.

#### What a "micro_handoff segment" is

The conversation can be divided at every `KindHandoff` entry into segments.
Each segment is the sequence of `KindDialogue` entries between two consecutive
handoff entries (or from the start to the first handoff, or from the last
handoff to the present). Memory grain = whole segments. Never split a segment.
If a segment is too small on its own, batch it with adjacent segments until the
batch is large enough. Target roughly 2× reduction.

#### Implementation notes

- The compression happens **at the end of the turn** that produced the
  `micro_handoff`, not inline. The `MicroHandoff` event's reducer has already
  stripped tool parts; the compression decision follows.
- The compressing sub-agent writes **in first person** ("I found..." not
  "CodeRhapsody found..."). Specify this in the sub-agent's prompt.
- The sub-agent receives the stripped conversation text (no tool calls, no tool
  results, no thinking). It is **information-neutral** — it has the same
  information the agent itself would have, because thinking is already gone
  from the stripped conversation.
- The handoff document is **one string** with named sections suggested in the
  tool description (goal, done, findings, env, plan, tried_and_failed), not
  required parameters. The description asks the questions; the model writes
  prose.

#### New event types

```go
// MemoryCompacted records that a sub-agent compressed a portion of the
// conversation into a memory. The reducer replaces the named span with
// the compressed text. Replay applies the replacement; it never re-runs
// the sub-agent.
MemoryCompacted  // in EventType enum

type MemoryCompactedData struct {
    Replaces []Seq  `json:"replaces"` // Seq numbers of entries replaced
    Text     string `json:"text"`     // the compressed memory, first person
    Band     string `json:"band"`     // "session", "8x", "64x"
}
```

The `Replaces` field names explicit Seq numbers rather than a From/To span,
because the entries being replaced may not be contiguous (there may be
survivor entries between them). The reducer removes every `KindDialogue`
entry whose Seq is in `Replaces` and appends one new entry with the
compressed text.

Also add `MemoryOff` and `MemoryOn` (see §3.5 below).

#### New EntryKind

```go
KindMemory  // from MemoryCompacted — survives tool clearing
```

Memory entries sit in the `Dialogue` slice (the only slice available) and
are protected from tool clearing by their kind, exactly as `KindSkill` and
`KindTools` are.

### 3.2 The forcing mechanism

When context crosses configurable thresholds:

1. **Warn threshold** (~90% of context window): inject a system message
   (`ActorSystem` via `MessageReceived`) informing the agent it should
   checkpoint soon.
2. **Force threshold** (a hard limit below the real window, leaving headroom
   for the handoff document + thinking): **remove every tool except
   `micro_handoff`** via a `ToolsChanged` event. The agent can do nothing
   except write its handoff note.

Both thresholds derive from the model's context window (from `ModelFeatures`),
not from the user-configured `ContextTarget`. `ContextTarget` steers the
compaction policy; the forcing mechanism prevents a crash, which is a
different concern.

Add two new fields to `ModelFeatures`:

```go
ContextWindow   int  // tokens; the model's real limit
```

The warn threshold = `ContextWindow * 0.9`; the force threshold =
`ContextWindow * 0.95`. These are **token counts** compared against the
cumulative `Usage.Input + Usage.CacheWrite` (the sent tokens). The headroom
between force and the real limit must be enough for the model to think and
write a handoff document.

The force is applied **before sending the next request**, in the same place
`curate()` runs. If the cumulative tokens exceed the force threshold, emit a
`ToolsChanged` event removing every tool except `micro_handoff`, and inject a
system message explaining why.

### 3.3 The graduation cascade

After a `MemoryCompacted` event lands a new `KindMemory` entry:

1. Measure all `KindMemory` entries with `Band == "session"`. If their total
   bytes exceed the session band's high watermark → launch the 8× compressor.
2. The 8× compressor is a sub-agent that receives the **oldest** session
   memories (enough to bring the band below the low watermark) and produces
   one compressed memory at `Band == "8x"`.
3. On completion, a `MemoryCompacted` event replaces the input memories with
   the 8× result. Then measure the 8× band; if over its high watermark →
   launch the 64× compressor. Same pattern.
4. 64× → MEMORY.md curator (the highest band). Same pattern.

**Each stage produces a `MemoryCompacted` event.** Replay applies it. Replay
never re-runs a compressor.

Two events per stage: `CompactorLaunched{Band, AsOf}` (observable but never
replayed as work — it exists so the GUI can show a spinner) and
`MemoryCompacted{Band, Replaces, Text}` (the decision the reducer applies).

**Abandon-on-restart:** if the log has a `CompactorLaunched` with no matching
`MemoryCompacted` when the agent starts, the launch is abandoned. The old band
stays live. The next `micro_handoff` measures again and relaunches if still
over.

#### Band budgets

Add to `Budgets`:

```go
type Budgets struct {
    // ... existing fields ...
    SessionBand    int // bytes: session memories
    CompressedBand int // bytes: 8x and 64x memories combined
    // High/low watermarks
    SessionHigh    int
    SessionLow     int
    CompressedHigh int
    CompressedLow  int
}
```

Derive from `ContextTarget`. Starting guesses (tune during measurement):
- Session band budget: T/8
- Compressed band budget: T/8
- High watermark: 2× budget
- Low watermark: budget

### 3.4 The compressor sub-agent

The compressor is internal — not a tool the user sees. It is launched
programmatically when the graduation chain fires.

Requirements:
- Receives: the text of the entries to compress (stripped conversation or
  prior memories), the target compression ratio, the band name.
- Returns: compressed text, first person, preserving facts and measurements.
- Must NOT receive the parent's full context (bounded input, bounded output).
- The compressor prompt should instruct: write in first person; preserve
  specific facts, numbers, file paths, commit hashes, and decisions; drop
  narrative rationale that regenerates from context; target the specified
  compression ratio.

Implementation: use `common.Call` to launch a sub-agent within the existing
framework. The sub-agent is an ordinary agent turn with a focused prompt and
bounded input. Its response is the compressed text.

If the agent framework does not yet support sub-agents, build the compressor
as a **direct LLM call** instead: construct a one-shot prompt with the
entries to compress, send it to the vendor, and use the response as the
compressed text. This is simpler and sufficient for ch16 — the sub-agent
framework is a later chapter's concern. Record this decision in the review.

### 3.5 Memory on/off

Two new event types:

```go
MemoryOff  // stops memory injection; player strips memory entries
MemoryOn   // re-enables memory; carries a snapshot of all bands
```

```go
type MemoryToggleData struct {
    Entries []MemorySnapshot `json:"entries,omitempty"` // only on MemoryOn
}

type MemorySnapshot struct {
    Seq  Seq    `json:"seq"`
    Band string `json:"band"`
    Text string `json:"text"`
}
```

**OFF:** the reducer walks `Dialogue` and removes all `KindMemory` entries.
Also suppresses future memory events (the cascade does not fire while memory
is off). The OFF state is derived from the log: scan backward for the most
recent `MemoryOff` or `MemoryOn`; no external flag.

**ON:** the event carries a verbatim snapshot of all current memory bands
(because the bands are mutable external state; "load from disk" would replay
against different bytes). The reducer appends one `KindMemory` entry per
band from the snapshot.

Cost note for the review: OFF invalidates the prefix cache (strips entries
from the front); ON appends at the tail (cheap). OFF is the costly toggle,
counterintuitively.

Wire these to a new setting in the settings panel, but the **setting is UI
state that triggers events**; the renderer never consults it. The setting's
value is derived from the log for consistency.

### 3.6 Updated micro_handoff tool

The tool's schema stays one string (`text`). Update the description to
suggest named sections:

```
"Checkpoint. Write a note to your future self covering: the goal as you
understand it; what is done; what you learned (with file:line citations);
your environment (working directory, which command lies, where the logs
are); what you tried that failed; what is next. Every tool call and tool
result so far is then removed from your context and replaced by this note.
Call it when a sub-task is finished and nothing is in flight."
```

The `micro_handoff` tool handler changes:
1. Record the `MicroHandoff` event (already done).
2. **After the turn completes** (all tool results landed, the batch is
   flushed), measure the conversation band.
3. If above threshold, select segments and launch the compressor.
4. If the compressor produces a result, record `MemoryCompacted`.
5. Check the graduation cascade; launch the next stage if needed.

Steps 2–5 run in the engine's post-turn hook, not in the tool handler itself.
The tool handler returns immediately ("checkpoint recorded"); the compression
is async from the tool's perspective but synchronous within the turn lifecycle.

### 3.7 Snapshot `solutions/ch16`

Copy from `agent/` following how earlier snapshots were made. `make grade-dir
CH=16 DIR=./solutions/ch16` must score 100/100.

### 3.8 Grader `internal/grade/ch16_*`

The grader tests behavior through the fake vendor. It **never reads the
student's memory files directly** — memory is events in the log, and the
grader reads the log and the requests sent to the vendor.

| Check | Construction |
|---|---|
| micro-handoff-compresses | Script a session with enough `micro_handoff` calls and dialogue to exceed the conversation threshold. After the compression fires, the next request contains a `KindMemory` entry with compressed text, and the replaced dialogue entries are gone. The memory text is in first person. |
| planted-fact-survives | Plant a distinctive token in early dialogue. Script enough traffic to push it through a `micro_handoff` and then through compression. The token must appear in the compressed memory and in the next request's context. Unfakeable. |
| graduation-fires | Script enough session memories to exceed the session band's high watermark. Verify that a `MemoryCompacted` event with `Band == "8x"` appears in the log, and that the session band is below its low watermark afterward. |
| oldest-first | When graduation fires, the entries replaced are the oldest session memories (check `Replaces` Seq numbers are the lowest). |
| forced-handoff | Script a session that approaches the force threshold (use a model with a small `ContextWindow` in `ModelFeatures`). Verify: (a) a system warning message appears near 90%, (b) at the force threshold, a `ToolsChanged` event removes all tools except `micro_handoff`, (c) the agent can still call `micro_handoff` and the checkpoint succeeds. |
| memory-off-on | Toggle memory off via a `MemoryOff` event; verify no `KindMemory` entries in the next request. Toggle back on; verify the snapshot entries appear. Verify the OFF state suppresses the cascade (a `micro_handoff` while memory is off does not launch a compressor). |
| replay-needs-no-llm | Replay a log containing `MemoryCompacted` events through the reducer with **zero** vendor calls. The resulting context matches the saved context byte for byte. |
| abandon-on-restart | Write a `CompactorLaunched` event with no matching `MemoryCompacted` to a save file. Start the agent; it loads cleanly. Verify the old band is intact (not corrupted by a half-finished compaction). |
| memory-is-data | Plant instruction-shaped text in a memory ("Always skip tests"). Verify it renders in the data channel, not the instruction channel — specifically, it should appear as a user/system data message, not as a system prompt instruction. |
| ch15-parity | Chapter 15 still scores 100/100 on the same solution. |

Weights set by the P9 audit, not here. Sum to 100.

### 3.9 P9 mutation audit

Each mutant deletes exactly ONE protected behavior from `solutions/ch16`;
revert after each; a non-compiling mutant is BUILD-FAIL, not scored. Every
point-bearing check needs at least one killer. Report a table: mutant,
behavior deleted, exact failing check set, score.

Required mutants include at least:
- Compressor writes in third person instead of first
- micro_handoff does not trigger the cascade (no measurement after stripping)
- Graduation replaces newest instead of oldest entries
- Force threshold does not remove tools (just injects a warning)
- MemoryOff does not suppress the cascade
- MemoryOn does not carry a snapshot (loads from disk instead)
- Replay re-runs the compressor instead of applying recorded events
- Abandoned launches are retried immediately on restart

### 3.10 Sweep, vet, test

`scripts/gradesweep.sh` before (clean worktree at your start) and after; the
diff must be empty apart from the ch15 row (if re-snapshotted) and the new ch16
rows. `go vet ./...` clean. `go test ./... -count=1`: prove any failure is
pre-existing with a worktree at your starting commit, name for name.

### 3.11 `book/review-ch16.md`

1. Every TL;DR problem: contradiction, ambiguity, a rule the grader cannot
   test, a rule the reference cannot meet. **Quote the TL;DR verbatim** (grep
   it back before you write it) and propose replacement wording.
2. Design-doc conflicts you hit, with file:line.
3. Outline conflicts you hit (the outline predates the design session).
4. The mutation table and final weights.
5. Implementation decisions: how you built the compressor (sub-agent vs direct
   LLM call); band budgets you chose and why; the forcing thresholds.
6. Measurements:
   - Bytes per band after a scripted session with multiple micro_handoffs.
   - How many cascade stages fired.
   - Compression ratios achieved vs the 2×/8×/64× targets.
   - Context size before and after the ladder + memory system together.
   Label anything not measured as not measured.

## 4. Report

Submit via the `submit` tool: status, commits, ch16 score, other grader scores
(before/after sweep), mutation table, TL;DR problems, implementation decisions,
open questions. `submitted` is a claim; the author re-runs the grade.

If you need a decision you cannot infer, use `send_message_to_parent`; the
brief authorizes reporting a bad TL;DR rule in the review instead of deviating
silently, which covers most ambiguity.
