# Brief: Chapter 15 coder (reference solution + grader + review)

You are the CODER for Chapter 15 of *The Self-Wielding Agent*, repo
`~/projects/ensemble`. The AUTHOR (a separate instance) wrote the chapter's
contract; you make it real, grade it, audit the grader, and report every place
the contract is wrong, ambiguous or ungradeable. Bill owns the repo and pushes.

Model: claude-opus-5. Expect a long job. Work in small commits.

---

## 1. Read first, in this order

1. `book/chapter-15.md`: opener, **TL;DR (the contract)**, §15.1 plain words.
   The TL;DR is authoritative for behavior. It contains a marked draft block of
   five questions (a–e) addressed to you.
2. `book/ch15-context-engineering-design.md`, Part I and Part II (Chapter A),
   plus the "Rulings, 2026-09-22 evening" section at the end. Rationale and
   file:line findings live here. Where it conflicts with the TL;DR, **the TL;DR
   wins and you report the conflict**; never resolve one silently.
3. `book/chapter-15-outline.md`, TL;DR spec and body sections, for context.
4. `book/chapter-11.md` TL;DR: the anchored save contract ch15 builds on
   (`save.json`, `as_of`, snapshot-plus-tail load, replay-equals-snapshot on
   vendor requests).
5. `book/chapter-writing-procedure.md` and `book/course-policy.md` (P9: every
   grader is audited by deleting one protected behavior from the reference and
   asserting the exact failing set).

## 2. Hard constraints

- Edit only: `agent/`, `solutions/ch14/` (step 3.0), `solutions/ch15/` (new),
  `internal/grade/ch15_*` (new), the ch15 wiring in `cmd/grade` and the
  `Makefile`, and `book/review-ch15.md` (new). **No other book file.** If the
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
  `freshRunDir` helper, added for ch11). The agent loads `./save.json` by
  default; a shared cwd makes runs load each other's conversations.
- Never run a grader while `internal/grade` is mid-edit; take baselines in a
  clean `git worktree` at your starting commit.

## 3. Work plan

### 3.0 Re-snapshot `solutions/ch14` first
`solutions/ch14` is stale: it scores 5/100 on the ch14 grader because it lacks
`scripts/tts-harness.sh`, which `agent/` has (`agent/` scores 100). Students
start ch15 from their ch14 agent, and a fresh-student test would start from
`solutions/ch14`, so it must be right. Re-snapshot it from `agent/` at your
starting commit (follow how earlier snapshots were made; Go's internal-package
rule may require path rewrites), confirm `make grade-dir CH=14
DIR=./solutions/ch14` = 100/100, commit. Report that the `ch14-solution` tag
needs moving; do not move it.

### 3.1 Build ch15 in `agent/`
Implement TL;DR rules 1–10. Things already known about the code (VERIFIED by
the author during design; re-check line numbers, they drift):

- **Skills ride in the tool result today.** The `load_skill` result carries the
  skill body (`agent/internal/tools/tools.go` around line 1016), so the first
  tool clear deletes the manual. Rule 4 moves the body into a `Skill` entry
  built by the reducer from `SkillLoaded` (whose `SkillData{Name, Body}`
  already carries the body, `event.go:201-204`); the tool result becomes an
  acknowledgement.
- **`summarizeSpan` folds everything in a span** (`context.go` ~285-305). Any
  compaction or redaction that walks a span must skip non-`Dialogue` entries,
  or it eats survivors (rule 3).
- **Text tool-result stubs carry no reference** (`stubFor`, `context.go`
  ~319-336; only `BlobPart` gets a Ref). Not a ch15 deliverable, but do not
  write prose-facing stub text that claims the bytes are addressable.
- **Two tool dispatchers exist:** `Engine.Execute` (ch3/ch4 path) and
  `Actor.dispatchTool` (ch6+). A tool-path change made in one only has broken
  ch4 before. Check both.
- **ch11's `Restore` returns an error on the first event it cannot apply.**
  Rule 8 makes the reducer total: skip plus diagnostic. Keep ch11 rule 2
  (a save that does not parse is still fatal).
- `RedactData{From, To, Level, Replacement, Reason}` (`event.go:209-215`) and
  levels `RedactResult`, `RedactTool` exist. Chapter 2 built them; no
  production code emits them yet. Rules 6 and 7 are their first producers.
- `ModelFeatures` / `LookupModel` is the per-model capability table (ch6/ch7);
  rule 6's gate is a new column there. No default row.

Design decisions that are yours, all to be written up in the review:
answers to draft questions (a)–(e); the on-disk log layout (rule 9 leaves it
to the student); the target-size → watermark/stub-threshold derivation
(rule 10), which must be simple enough to print as one formula.

Rule 6 semantics, stated precisely so you can check the TL;DR's wording:
round trip k's model response calls a tool; request k+1 carries the full
result; request k+2 carries a stub unless the model's response to k+1 called
`keep_tool_results`. The call survives in both. The stub is a `RedactResult`
event in the log, so replay reproduces it. On a model without the feature,
nothing is stubbed per round trip.

### 3.2 Snapshot `solutions/ch15`
Same procedure as ch14. `make grade-dir CH=15 DIR=./solutions/ch15` must score
100/100.

### 3.3 Grader `internal/grade/ch15_*`
The exercise table in the TL;DR is the check list; names are fixed, weights are
provisional (you set them via the audit; they must sum to 100). Grade behavior
only: **what the fake vendor receives, and the ch11 `save.json`**. Never read
the student's own on-disk log layout (rule 9 makes it theirs). Suggested
constructions (improve them; report what you changed):

| Check | Construction |
|---|---|
| skill-survives-the-ladder | load a skill, then generate enough tool traffic to push the ladder past both watermarks; later requests still contain the skill body exactly once, outside any tool part |
| micro-handoff-shape | script a `micro_handoff` call; the next request has no tool call/result parts older than it, the handoff text exactly once, and no orphaned call or result anywhere |
| keep-or-stub | two scripted round trips with large results, one kept and one not; request k+2 has the stub for the unkept one, the full result for the kept one, the call in both. Repeat on a model whose features row lacks the column: no stubs |
| ladder-is-recorded | run until cuts happen, exit (save), change the target size in `settings.json`, restart; the next request's history is byte-identical to what the old settings produced (cuts are events, not recomputed) |
| frozen-prefix | across a session with a mid-session skill load and an MCP connect (ch13's FakeMCPServer), the system prompt and startup tool declarations are byte-identical in every request; new tools appear in the dialog |
| replay-equals-snapshot | ch11's check extended to the new event types: load as written vs `context` nulled, byte-identical next request, on a save containing `MicroHandoff`, `RedactData` and `SkillLoaded` events |
| crash-recovery | SIGKILL the agent mid-session after N turns (no clean exit, so no snapshot), restart in the same dir; the next request carries all N turns |
| total-reducer | plant, in `save.json`'s log, one event that parses but cannot apply (a redaction naming a Seq no entry has; a malformed payload); the agent starts, and the next request is well-formed and carries the rest of the conversation |

Wire `make grade15` / `make grade-dir CH=15` like the neighbors. Consider
whether a `ch14-parity` check belongs in the table (ch11 had `ch10-parity`);
propose it in the review with a rebalance rather than silently adding it.

### 3.4 P9 mutation audit
Each mutant deletes exactly ONE protected behavior from `solutions/ch15`;
revert after each; a non-compiling mutant is BUILD-FAIL, not scored. Every
point-bearing check needs at least one killer. Report a table: mutant,
behavior deleted, exact failing check set, score. Required mutants include at
least: skill body left in the tool result; survivors not skipped by the span
walk; handoff that leaves an orphaned tool result; stubbing that ignores
`keep_tool_results`; stubbing on a model without the feature; ladder that
recomputes cuts from current settings instead of recorded Seqs; prefix that
re-declares tools on skill load; log written only at exit; reducer that aborts
on a bad event. If a mutant survives, strengthen the check only where the
TL;DR already promises the behavior; otherwise report it.

### 3.5 Sweep, vet, test
`scripts/gradesweep.sh` before (clean worktree at your start) and after; the
diff must be empty apart from ch14 on `solutions/ch14` (now 100) and the new
ch15 rows. `go vet ./...` clean. `go test ./... -count=1`: 19 tests under six
top-level cases were already failing (TestCh3PointsSumTo100,
TestCh6ReferenceScores100, TestCh6DeletionAudit, TestCh7ReferenceScores100,
TestCh7DeletionAudit, TestCh8Grade); prove any failure is pre-existing with a
worktree at your starting commit, name for name.

### 3.6 `book/review-ch15.md`
1. Answers to TL;DR draft questions (a)–(e), each as the exact text the author
   should print.
2. Every TL;DR problem: contradiction, ambiguity, a rule the grader cannot
   test, a rule the reference cannot meet. **Quote the TL;DR verbatim** (grep
   it back before you write it) and propose replacement wording.
3. Design-doc conflicts you hit, with file:line.
4. The mutation table and final weights.
5. Measurements you can produce cheaply (the chapter is required to print
   real numbers): bytes per request over a long scripted session with and
   without the ladder; how many ladder events fire; cache-relevant prefix
   stability. Label anything not measured as not measured.

## 4. Report

Submit via the `submit` tool: status, commits, ch15 score, other grader scores
(before/after sweep), mutation table, TL;DR problems, answers to (a)–(e),
open questions. `submitted` is a claim; the author re-runs the grade.

If you need a decision you cannot infer, use `send_message_to_parent`; the
brief authorizes reporting a bad TL;DR rule in the review instead of deviating
silently, which covers most ambiguity.
