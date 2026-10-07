# Restart checkpoint

Work paused at Bill's request so he can restart Codex with different sandbox
settings. Resume from this file and `progress.md`, not from a new plan.

## Objective and authorization

Write the second edition and regenerate Ensemble as a student of that book.
Continue autonomously; ask Bill real unresolved questions. Exactly **21
chapters, 0–20**: leave Chapter 0 and preface alone; rewrite from Chapter 1;
remove old Chapters 5 and 22 as standalone repairs and absorb their lessons
where first needed. Old 6–21 become new 5–20. Detect drift and fix the coder,
never dilute the architecture to accommodate it.

Do not edit existing `agent/` implementation or first-edition solutions.
New code belongs in `solutions/edition-2/chNN/`, each a self-contained Git
repository. Bill authorizes initializing repositories and snapshots. Start
Chapter 1 from scratch; derive later repos/history only from preceding new
solutions. Do not copy old answer-key code. Grader improvements are authorized,
including coder findings, but cannot weaken coverage and must preserve legacy
tests. Do not push or include Bill's unrelated files in commits.

## Mandatory context

Read these before resuming:

1. Root `AGENTS.md`.
2. `book/chapter-writing-procedure.md` (rewritten to reflect this session).
3. `book/edition-2/skills/ensemble-coding/SKILL.md` in full before ANY coding
   task, including tests/graders, and after compaction. Every coder handoff
   requires that read explicitly.
4. `book/edition-2/architecture.md` and `workflow.md`.
5. Author and proofreader both read all of `book/voice.md` and the procedure.
6. Current chapter, outline, evidence, and `global-review.md` as relevant.

Chapter 1 now teaches architecture in plain English before all code: star
imports; shared core structures/interfaces in `internal/common`; behavior in
the owning spoke, using free functions rather than moving it to common for
method syntax; interface parent chains; logger access for all likely-to-log
code including stateless helpers. Common's vocabulary hub is not runtime
Hub/Ensemble. Mutable application state belongs to objects, not globals.

One runtime Ensemble owns many Agents. Working implementation decision:
Ensemble owns logger, Agent config/history, Engine transport/usage. Parent
interfaces are in common. Observer carries streaming/real-time display events;
explicit parent methods also support requests and services. No observer-only
prohibition remains. Registry may be Agent-owned or shared on Ensemble with
per-agent visibility; choose when tools are introduced.

GUI/WebSocket code belongs in a separate optional Go MODULE, never
`agent/internal`. Starting Chapter 2, clean core data structures and both CLI
and browser-GUI client interfaces exist; GUI may be an honest stub then.

## Team workflow

Author, student coder, global reviewer, and independent proofreader roles were
established. Only three worker threads fit in this session: attempting a fourth
failed. The global review thread saves its map and switches to a dedicated
proofreading phase; it remains independent of the author. Recreate workers as
needed after restart; do not assume the old agent threads remain available.

The student has NOT read old solution/agent/grader implementation. It reads
the new chapter and skill only, plus earlier new snapshots. Root handles
independent grader engineering. Proofreader returns findings to the author.

## Current chapter state

### Chapter 1

`book/edition-2/chapter-01.md` has full teaching and a usable student contract.
Proofreader read full voice/procedure and resolved its material findings:
credential-safe discovery example, timeout/logger contract in TL;DR, external
public-library consumer for live ownership checks. Prose lint has no hard
failures; soft warnings do not warrant padding. Section 1.9 still has a LIVE
RECEIPT PENDING placeholder. Do not claim completed validation.

Student implementation is underway in `solutions/edition-2/ch01/`, a newly
initialized Git repo: public root library, `internal/common`, `internal/llm`,
`cmd/` JSON-lines CLI, tests. See its own `CHECKPOINT.md` for the latest coder
results and exact unfinished work. No optional chat mode was planned.

Existing `grade.Build` already discovers `cmd/` beneath a solution directory;
use `make grade-dir CH=1 DIR=solutions/edition-2/ch01`.

The new chapter adds pass/fail acceptance beyond the old seven-check score:
configuration failures, malformed/empty input, response validation, finite
timeout, exact history growth, external library use, independent Agents,
logger reachability, and architecture. Root has not yet changed graders or
implemented independent new acceptance/mutation checks. Read §1.8.

### Chapter 2

Outline, evidence, and initial contract exist in `chapter-02*.md`.
**Not ready for student implementation.** Needs literal event/data schemas,
fixtures, CLI contracts, and requested-versus-returned model provenance rules.
Author researched official current API fields. See evidence, not assumptions:
OpenAI docs currently describe both cached and cache-write input fields;
Gemini prompt counts include cache and total includes thinking+candidates.
Keep original initial surfaces Messages, Chat Completions, and generateContent
unless source evidence requires a documented change. Three-vendor live checks
are mandatory once adapters exist.

Global map is saved. Its old-number rows are historical references; suggestions
to retain old 5/22 as audit chapters must be updated to the final 21-chapter map.

## Tests and evidence

- Legacy Chapter 1 reference + all mutation cases PASS:
  `GOCACHE=/tmp/ensemble-edition2-go-cache go test ./internal/grade -run
  '^(TestReferenceSolutionPasses|TestMutationsAreCaught)$' -count=1`
  Local fake-server listening required execution outside the sandbox.
- Existing `agent/` full `go test ./... -count=1 -timeout=20m` PASS, exit 0,
  without implementation changes. Log copied to checkpoint evidence.
- Full root `go test ./... -count=1 -timeout=45m` was interrupted for this
  checkpoint before the grade package completed. **NOT a passing baseline.**
  Partial log saved; rerun on resume before shared grader enhancements.
- Bundled skill validator cannot start because Python lacks PyYAML. Ruby
  successfully parsed skill frontmatter and checked name/description; that
  does not replace behavioral validation. Read-only coder scenarios and full
  global review found no material rule contradiction.

## Live demonstrations and credentials

Every "Taking it for a spin" requires the coder actually run the user-facing
program with a REAL model, initially CLI, exercising every chapter feature.
Fake grading is complementary, never a substitute. Use a runnable external
consumer for public-library features; retain sanitized receipts for the author.
The proofreader checks prose against those receipts. No invented transcripts.

Bill authorizes API keys in `~/.cr/settings.json`. Read needed values
programmatically into memory/child environments; never dump the config,
print keys, place them in argv/prompts/logs/repo files, or commit them.
Use all three vendors as introduced. Model IDs must be discovered/verified.
Check coder checkpoint for whether any live attempt happened before pausing.

At the caching chapter, check OpenAI's then-current official subscription
access method and TEST OAuth caching against comparable API-key controls.
Bill reports a recently introduced subscription/API capability and an OAuth
caching defect; these are leads, not newly verified facts. Do not declare it
fixed because a report was closed. Record actual endpoint, model, mode, cache
counts, and date, never credentials. Do not research/implement it prematurely.

## Next executable actions

1. Read the mandatory context and the student repository checkpoint; inspect
   its current commit/status without mixing nested repos into the parent commit.
2. Resume student Chapter 1 validation and real-model demonstrations. Have the
   grader engineer add independent acceptance checks and property mutants,
   retaining a passing legacy baseline. Review actual architecture for drift.
3. Give observed receipts to the author, finish §1.9, run proofreader review,
   and commit a validated Chapter 1 snapshot only after all required evidence.
4. In parallel finish Chapter 2's concrete contract. Then continue through
   new Chapter 20 using the same author/student/grade/live/review cycle.
