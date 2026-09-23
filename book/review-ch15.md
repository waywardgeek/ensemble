# Review: Chapter 15 coder pass

Coder: CodeRhapsody (Opus). Start commit 93745ce. Brief: `book/brief-ch15-coder.md`.

## Results

| | before | after |
|---|---|---|
| ch15 on `./agent` | n/a | **100/100** |
| ch15 on `./solutions/ch15` | n/a | **100/100** |
| ch14 on `./solutions/ch14` | 5 | **100** (re-snapshot, cfc0f7f) |
| every other sweep row (`scripts/gradesweep.sh`) | | unchanged (diff shows only the ch14 row) |

`go vet ./...` is clean. `go test ./... -count=1` has the same failing set as a worktree at 93745ce, name for name, with one exception: `TestCh4DeletionAudit` is a **baseline flake**. At 93745ce a different subtest failed on each of two reruns (`pattern-matches-output-already-seen`, then `no-truncation-at-all`). In this tree it passed on both reruns. It is not in the brief's list of six.

Commits, all unpushed: cfc0f7f (solutions/ch14), 58abb9e (ch10 grader, see §3), 41930f6 (agent), 369bfb4 (grader), ab342d6 (solutions/ch15), ded8217 (grader strengthened plus mutation script).

**Tag moves needed** (I moved none): `ch14-solution` to cfc0f7f; new `ch15-solution` at ab342d6.

## 1. Answers to the draft questions

**(a)** Print:

```go
type ToolsChangedData struct {
    Added   []ToolDecl `json:"added,omitempty"`   // full declarations, ch10's shape
    Removed []string   `json:"removed,omitempty"` // names only
}
```

A delta, not the new full set. The reducer turns it into an entry of a fourth kind (see §2.1). On the Anthropic API that entry renders as a mid-conversation `system` message with `tool_addition` and `tool_removal` blocks (beta headers `mid-conversation-tool-changes-2026-07-01`, `inline-tools-2026-09-15`). Other vendors fold every delta into the startup set and re-declare, which costs a cache miss.

**(b)** Print as one formula: *with target size T bytes: stub threshold T/100; results band T/8; calls band T/16. A band is cut back to its budget when it passes twice it.* The default T is 400,000 (threshold 4,000, bands 50,000 and 25,000). The minimum T is 20,000, and smaller values are clamped. "Calls band" means the bytes of calls, and of the stubs they keep, that are older than the results band. Tool bytes older than the calls band get `RedactTool`.

**(c)** The audit (§4) keeps the TL;DR's weights unchanged: 15 / 15 / 10 / 15 / 10 / 15 / 15 / 5. Every point-bearing check has a killer, and no check's killers are all compound, so nothing argued for moving points.

**(d)** Print: `StubsToolResults`. Rule 2 needed a second column the TL;DR doesn't mention, `InlineTools`: whether a model can take tool declarations in the dialog. The two are independent (Sonnet 5 has neither; Opus 5 has both). See §2.4.

**(e)** Print: *`keep_tool_results` governs the batch before the message that calls it. Its own result is never stubbed, and it cannot keep its own batch. Calling it alongside other tools keeps the previous batch, not the one it rides in.* The grader exercises exactly this (keep-or-stub's parallel call).

## 2. TL;DR problems

### 2.1 A fourth entry kind

TL;DR:

> ```
>     KindDialogue EntryKind = iota + 1 // prompts, hints, output, tool parts
>     KindHandoff                       // from MicroHandoff
>     KindSkill                         // from SkillLoaded
> ```

Rule 2 puts declarations "in a dialog entry", and that entry must survive tool clearing and `micro_handoff`. If it were `KindDialogue`, a handoff or a ladder span could delete the only record that a tool exists, while the model still calls it. The reference adds `KindTools // from ToolsChanged`. Proposed: add that line and extend rule 3 to "A `Handoff`, `Skill` or `Tools` entry never holds a tool call or a tool result…".

### 2.2 Rule 5 and parallel calls

TL;DR:

> Its reducer removes every tool
>    call and tool result part from the context

If the model calls `micro_handoff` alongside another tool, the other call's result arrives after the `MicroHandoff` event. Clearing at the event would orphan that result. The reference holds survivor events until the batch's last result lands, then applies them. Proposed addition: "…applied when the batch's last result arrives, so a parallel call's result is cleared with it."

### 2.3 Rule 10 must name the settings key

TL;DR:

> The Context Management settings tab sets a target
>     context size in bytes.

ladder-is-recorded changes the target in `settings.json`, so the key is contract. The reference uses `context_target` (bytes; 0 means the default) and `log_retention` (events kept in the saved log; 0 keeps all). Proposed: "…sets `context_target` in `settings.json`…" and name `log_retention` the same way.

### 2.4 The grader's models are contract

keep-or-stub and frozen-prefix launch `claude-opus-5-course` (must have `StubsToolResults` and `InlineTools`) and `claude-sonnet-5-course` (neither). The first is a new row in the reference table. A student whose table lacks it scores zero on those checks, for reasons the TL;DR never states. Proposed sentence under the exercise table: "The grader runs as `claude-opus-5-course`, whose row enables rule 6 and inline tools, and `claude-sonnet-5-course`, whose row enables neither."

### 2.5 Rule 9 promises the grader cannot see

"Write the new snapshot before truncating the log; never truncate past the anchor; keep one backup of the previous snapshot." The reference does all three: `SaveRetaining` writes `save.json.bak`, then the snapshot, then resets the journal. None of it is graded. Seeing any of it would require killing the agent *during* shutdown. I left it ungraded rather than add a timing-dependent check. The author decides whether to keep the promise in the TL;DR as unenforced guidance.

### 2.6 Crash-safe means process crash

The journal appends one JSON line per event with no fsync. That survives a process crash (SIGKILL, which the grader uses) but not power loss. If the chapter says "a crash loses nothing", it should say "a process crash".

### 2.7 Skills on unload

After `unload_skill`, the `Skill` entry stays in the context: the tools are removed through `ToolsChanged`, but the body stays until a later chapter's removal verb. Deleting it would be a deep mutation (a cache miss), which contradicts ch10's lazy unload. The TL;DR is silent. One sentence would prevent a student from "fixing" it.

## 3. Conflicts with other files

- **ch10 grader (fixed by ruling).** ch10's "$VAR tokens rendered in skill body" read the rendered body out of the `load_skill` tool result, and rule 4 takes it out. agent/ dropped to 90 on ch10 and on ch12 (through ch11's parity). Bill ruled, on 2026-09-23, to revise the check: it now finds the rendered value in the next request and not in the one before (58abb9e). solutions/ch10 and agent/ both score 100. Mutant M10 (below) shows the check still catches missing rendering.
- **Design doc §A.13 versus TL;DR table.** The design doc's grader table differs from the TL;DR's (it has no keep-or-stub; skill is worth 20 and total-reducer 10). I followed the TL;DR, which is authoritative.
- **Pre-existing: `const MaxToolRounds = 16`** in `agent/internal/llm/engine.go` (present at 93745ce). The measurement session hit it: 60 reads in one prompt stopped at 16 rounds, even with `max_tool_rounds: 100` in settings.json. The grader's longest turn is 13 rounds. I didn't touch it; I'm flagging it for whoever owns the settings path.

## 4. P9 mutation audit

Script: `internal/grade/ch15_mutations.py`. Each mutant patches a scratch copy of `solutions/ch15` (M10 patches `solutions/ch10`), builds it and grades it. The table below is the verbatim output after the final grader commit.

| Mutant | Behavior deleted | Failing checks | Score |
|---|---|---|---|
| M1 | skill body left in the load_skill tool result | skill-survives-the-ladder | 85 |
| M2 | span walk does not skip survivors | **none (equivalent, see below)** | 100 |
| M2b | COMPOUND: M2 plus RedactTool also dropping text | skill-survives-the-ladder, micro-handoff-shape | 70 |
| M3 | handoff clears calls but leaves their results | micro-handoff-shape | 85 |
| M4 | per-round-trip stubbing ignores keep_tool_results | keep-or-stub | 90 |
| M5 | stubbing on a model whose row lacks the column | keep-or-stub, ladder-is-recorded | 75 |
| M6 | ladder cuts recomputed from current settings, never recorded | skill-survives-the-ladder, ladder-is-recorded, replay-equals-snapshot | 55 |
| M7 | startup tools array re-declared on skill load | frozen-prefix | 90 |
| M8 | events reach disk only at shutdown | crash-recovery, total-reducer | 80 |
| M9a | load aborts on an unappliable event | total-reducer | 95 |
| M9b | load stops at the first unappliable event, keeping what came before | total-reducer | 95 |
| M11 | a skill load emits no ToolsChanged | frozen-prefix | 90 |
| M10 | ch10: $VAR tokens not rendered in skill bodies | ch10 "$VAR tokens rendered in skill body" | 90 |

**M2 is equivalent under the levels ch15 emits.** `RedactResult` touches only result parts and `RedactTool` only call and result parts. Rule 3 keeps both out of survivors, so the Kind filter in the span walk changes nothing until a level that can reach text exists (`RedactDialogue`, `RedactSummary`: ch16). Rule 3 is the real protection, and M2b shows the grader catches a lost survivor as soon as a level can reach one. The filter stays in the reference because ch16 needs it.

**Two checks were strengthened by the audit, both within promises the TL;DR already makes:**

1. *skill-survives-the-ladder* never reached the second watermark. The fixture's short file paths never filled the calls band, so `RedactTool` never fired. M2b surviving the first run is what exposed this. The planted files now have long real paths, and the check asserts that the oldest read's call is gone (RedactTool) and a later read has its call kept and its result stubbed (RedactResult). The brief asked for "past both watermarks", and now it's graded.
2. *total-reducer* planted its bad events in the tail, so M9b (stop at the first bad event) lost nothing and survived. The events now sit before the last turn, with the context nulled so load replays through them. What's graded is rule 8's "loading continues".

Other changes from the brief's suggested constructions:

- crash-recovery first runs a clean session, then SIGKILLs a second one, so recovery has to compose a snapshot with a tail. That's rule 9's actual wording.
- ladder-is-recorded compares a restart at the same target against one at a raised target. Comparing against the old session's own last request isn't possible, because that request carried a different prompt.
- frozen-prefix uses one skill that adds `think` and one that connects ch13's fake MCP server, and requires both tool names to appear in the dialog.

**ch14-parity.** I'm not adding it. Ch15 changes no ch14 surface, and the sweep already shows ch14 at 100 on `./agent`. If the author wants it anyway, a 5-point check funded from frozen-prefix (10 → 5) would be the rebalance; I'd argue against it, since the sweep already does that job.

## 5. Measurements

These are synthetic: the fake vendor, 4 prompts × 15 `read_file` calls on 8,000-byte files, 64 requests per run. Reproduce with `cr/docs/ch15_measure_test.go.txt` (a throwaway test, not committed).

| Run | Last request (bytes) | Largest request | Total sent | Ladder events | Round-trip stubs |
|---|---|---|---|---|---|
| no ladder (T = 10^12, sonnet-5-course) | 542,600 | 542,600 | 17,470,874 | 0 | 0 |
| ladder only (T = 400,000, sonnet-5-course) | 143,152 | 146,378 | 6,172,202 | 7 (all RedactResult) | 0 |
| ladder + stubs (T = 400,000, opus-5-course) | 62,456 | — | 2,334,066 | 0 | 59 |

- The ladder alone cut total bytes sent to 35% of the unbounded run. The ladder plus per-round-trip stubs cut them to 13%.
- The last request is 26% and 12% of the unbounded run's, respectively.
- The largest ladder-only request (146,378) sits near the sum of both bands at twice their budgets, as the cut-at-2× rule predicts.
- The prefix (system plus tools, 2,991 bytes) was byte-identical across all 64 requests in all three runs.
- **Not measured:** cache hit rates on a real vendor, the effect on answer quality, and Bill's hypothesis that auto-redaction interferes with thinking. The Anthropic docs page on mid-conversation system messages (platform.claude.com, fetched 2026-09-22) says that on Fable 5.1 and Opus 5.5, deleting earlier content fails the conversation check for later preserved thinking blocks. That bears on the hypothesis, and nothing here tests it.

## Open questions

1. Should `const MaxToolRounds = 16` (engine.go) honour `max_tool_rounds`? It predates ch15.
2. Should rule 9's snapshot-ordering and backup promises stay in the TL;DR unenforced (§2.5)?
3. Should the chapter define "crash" as a process crash (§2.6)?
