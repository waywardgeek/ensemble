# Chapter writing procedure

The second-edition workflow for *The Art of Building AI Coding Agents*.
Updated for Bill's instructions in this session. This replaces the old
code-first outline procedure and its per-chapter approval ceremony.

## Outcome and scope

The current plan is **22 chapters, numbered 0–21**. Bill subsequently
authorized chapter adjustments when they improve the teaching and explicitly
requested a Chapter 0 guide to executing the second edition. Enhance that
chapter in `book/edition-2/chapter-00.md`; preserve its first-edition source
and the preface. Bill also requests a new final chapter comparing the editions;
see `book/edition-2/ending-chapter-note.md` for the evaluation plan. Rewrite
the exercises from Chapter 1. Absorb the lessons of first-edition Chapters 5
and 22 into the chapters where the decisions first matter; neither survives
as a standalone repair chapter. Chapters 1–4 retain their numbers, and old
Chapters 6–21 become new Chapters 5–20 unless a dependency requires an
explicitly documented adjustment with a concrete teaching benefit. Do not
restore an avoidable architecture-repair chapter merely to preserve history.

Preserve each edition as a historical artifact in *The Singularity as it
Happened*, including its text, code, graders, and available evidence. Correct
and explain earlier decisions in the new edition instead of editing the old
record to match the new account.

When the second edition is ready for publication, rewrite its epilogue
entirely in Codex's own first-person voice, as Bill explicitly requests.
This exception applies to that epilogue, not ordinary chapter bodies or the
first-edition epilogue. Write from the completed comparison and actual record;
do not invent the outcome now.

The manuscript lives in `book/edition-2/`. The sole development tree is
`solutions/edition-2/main/`, tracked by the outer Ensemble Git repository.
Initialize Chapter 1's Go module there from scratch; subsequent chapters
extend its accepted second-edition predecessor. Do not create nested Git
repositories. The `solutions/edition-2/chNN/` directories are frozen, ordinary
tracked exports of exact accepted source versions, with module files, tests
and evidence. They are references for readers, not parallel development trees.

The coordinator prepares the student's main tree at the permitted preceding
chapter version. A cold student must not inspect later main history to infer
an answer. Work on an earlier correction uses an isolated checkout/branch at
that chapter's accepted source, then carries the validated correction forward;
do not copy a later implementation backward into an earlier exercise.

**Do not edit existing Ensemble implementation in `agent/` or first-edition
solutions.** They are historical evidence for the author and global reviewer,
not code for the student to copy. Graders may be enhanced, but coverage cannot
be weakened and the existing tests must remain passing.

Bill authorizes autonomous drafting, implementation, review, and correction.
Do not require a new outline approval for every routine decision. Ask real
questions when his judgment changes the outcome; continue independent work.
Draft validation and Bill's editorial approval are different statuses. Never
claim he approved a chapter merely because its checks passed. Push is Bill's.

## Roles

- **Author:** owns outlines, exercise contracts, chapter prose, and evidence
  reconciliation. Reads history to find lessons that belong earlier. Does not
  implement the student's solution.
- **Student coder:** builds from the new chapter and earlier new baseline,
  reports gaps, tests the result, and actually uses the CLI with real models.
  Does not edit chapter prose or consult first-edition chapters or old answers.
- **Grader engineer:** derives checks from the taught contract, independently
  of the student's implementation. The coordinator can fill this role.
- **Global reviewer:** reads the complete textbook, records dependencies and
  later lessons, and answers questions about the whole-book consequences.
- **Code reviewer:** after the student's first implementation and runs,
  compares it with the corresponding first-edition standard solution. Gives
  concrete quality feedback to the coder and teaching feedback to the author,
  then reviews their revisions. This reviewer is independent of the coder.
- **Reviewer / proofreader:** reads the author's work independently for voice,
  clarity, organization, consistency, and evidence. Returns concrete findings
  to the author, then reviews the revision. It is distinct from authorship.
- **Coordinator:** sequences the work, resolves source conflicts, relays
  questions to Bill, enforces architectural corrections, and records progress.

Both **Author and Reviewer must read all of `book/voice.md` and this procedure**
before working, and reload after compaction or a rule change. Use the current
voice file where older procedural wording disagrees with it.

Agent slots may require scheduling roles in phases. In this session the
global review thread also serves the code-review and proofreading phases after
saving its whole-book findings. Keep role and review evidence explicit; do
not describe that as a separately spawned fourth worker.

Every coding agent, including the grader engineer, must read the entire
`book/edition-2/skills/ensemble-coding/SKILL.md` before each coding task and
after compaction. Every handoff names that path and requires the read before
editing. Root `AGENTS.md` carries the same instruction. A remembered summary
and automatic skill discovery do not fulfill the requirement.

## Architecture precedes implementation

Chapter 1 must teach these rules in plain English **before any code**:

- Imports form a star. Core data structures and interfaces, including parent
  interfaces, live in `internal/common`; implementation spokes never import
  one another, and common imports no implementation.
- Behavior belongs in the directory responsible for the work. If a type in
  common prevents methods in `internal/llm`, use free functions in llm.
  Go's same-package method restriction does not decide the architecture.
  Retain only the narrow shared-type methods needed for standard-library
  interface dispatch.
- Child objects receive interface back-pointers to their actual owners.
  Follow the chain to owned data and services; do not inject sibling
  dependency bags or attach closures to repair unreachable facilities.
  Every line likely to need diagnostics can reach the logger.
- One Hub/Ensemble owns potentially many Agents. Shared vocabulary at the
  center of the import star is a package, not this runtime owner.
- Observer delivers streaming and real-time display events. Parent interfaces
  also support explicit requests and service access.
- GUI code, including WebSocket transport, belongs in a separate optional Go
  module, never `agent/internal`. Clients use the public agent interfaces.

Starting in Chapter 2, clean core structures support **two clients** of
Hub/Ensemble: CLI and browser GUI through WebSocket. The GUI can be a stub in
that chapter; the boundary and interface must already exist.

Read `book/edition-2/architecture.md` for current owner decisions. A small
Chapter 1 conversation does not excuse postponing the methodology. Add only
the mechanisms the chapter needs, while placing them correctly from birth.

If any agent detects architectural drift, stop the affected implementation,
name the violated rule and the concrete code, and direct the coder to fix it.
Strengthen a missing check and rerun relevant tests. Do not weaken the chapter,
skill, or grader to accommodate drift, or defer repair to an extra chapter.

## 1. Derive the chapter from evidence

Before an outline, the author reads:

1. The current chapter and relevant predecessor contracts.
2. Relevant git history, prior briefs, coder reviews, and grader audits.
3. Current grader fixtures, checks, and observed reference behavior.
4. The global reviewer's map of later lessons needed at this point.
5. Current architecture decisions, course policy, voice, and this procedure.

Distinguish current source from stale notes and superseded examples. A brief
claiming a bug is open may predate its fix. Record commit IDs for findings.
First-edition fixtures are evidence of the old contract; they do not override
Bill's new requirements or compel the new student to repeat old architecture.

Measure every published figure. Do not invent private anecdotes, quotes,
model identifiers, timing, output, or test results. Verify current external
claims against primary sources. Existing model IDs are dated facts, not safe
defaults taken from memory.

The chapter outline records its thesis, teaching order, full exercise
contract, checks, preserved prior behavior, and genuinely unresolved choices.
Record a short voice plan using the current voice guide. Ask Bill only where
his answer materially changes the design.

## 2. Write the student-facing contract

Write enough of the chapter before implementation that the student can build
from the teaching. Chapter 1's plain-English rules precede all code. Each
chapter explains the idea in plain words before its detailed mechanism.

The second-edition journey is learn the rules, build correctly, then extend.
Motivate a rule before demanding compliance: briefly describe the failure it
prevents, then explain the design. A paragraph about a timeout that cannot
reach the right logger can teach the reason for a parent chain without making
the student build a flawed version first. Keep the opener warm and direct
within `voice.md`; a precise contract does not excuse specification-only prose.
Use documented experience, never invented Bill stories or unsupported numbers.

Chapter 1 explicitly tells the student to load
`book/edition-2/skills/ensemble-coding/SKILL.md` before writing code and explains
why that instruction is part of the working method. The skill supplies the
rules during implementation; checks and independent review detect violations.
Do not claim that loading a text skill mechanically prevents invalid code.

The TL;DR includes:

- Data structures and wire/on-disk contracts the student needs.
- Numbered requirements, including failure and lifecycle behavior.
- Choices genuinely left to the student.
- A working build/grader command for the new snapshot.
- A mapping from required behaviors to acceptance checks.

Keep the TL;DR concise, following the voice/procedure tradition of at most
700 prose words when feasible; code and necessary literal fixtures are separate.
Print exact fixture bytes when those bytes are part of the public contract.
Never grade an unpublished private assumption.

For a cold student evaluation, provide the architecture rules from Chapter 1,
the mandatory coding skill, current and earlier second-edition TL;DRs, the preceding new
solution, and the grader command. Do not omit the architecture from a
"TL;DR-only" test. Do not provide future or first-edition solution code or
grader internals. In a shared workspace this is an instruction boundary,
not a claimed filesystem sandbox. Disclose accidental answer exposure.

Use a fresh student context for each chapter, with no inherited coordinator
conversation (`fork_turns="none"` in the current orchestration API). The
coordinator has read material the student must not see. Explicitly exclude old
chapters, old solutions, and author/reviewer research notes, including historical
sources linked from the coding skill. Record actual source reads. An instruction
not to open an old file does not undo old text already inherited in context.
Keep post-run comparison findings separate from the initial student attempt;
the reviewer supplies rationale rather than old answer code.

Method and field spelling may be student decisions where behavior and
structure are equivalent. Unspecified spelling alone does not require a
user ruling. Unresolved ownership or conflicting requirements do.

## 3. Build, grade, and repair the specification

The student implements from scratch in Chapter 1 and from the preceding new
baseline thereafter. Before editing, load the full coding skill and identify
any contract ambiguity. Report gaps to the author rather than silently filling
missing teaching with knowledge from another solution.

The grader engineer builds checks from the contract. Test behavior and
structural properties, not identifier vocabulary or code lineage. Include
each newly added spoke and every executable in architecture checks.

Run the new solution's build, module tests, vet, and relevant grader against
`solutions/edition-2/main/`. A frozen `chNN/` export can be graded read-only
at its own feature scope. Preserve historical grader numbering where the
second-edition chapter explicitly maps to a different inherited grader.
Then audit the advertised properties by deletion:

- Keep a correct positive control.
- Delete each protected behavior and require the intended failure and score
  loss; prefer exact failing-check sets.
- Use negative controls for checks asserting absence.
- Plant independent fixtures so one broken feature does not falsely fail
  an unrelated check.
- Distinguish check coverage from coverage of all promises inside the check.

The student may discover and enhance deficient graders; Bill explicitly
allows it. Prefer routing findings to the independent grader engineer so the
student remains blind. If the student reads or changes grader internals,
record that exposure and do not call a subsequent attempt a cold evaluation.

Before changing a shared grader, establish its legacy baseline. After changes,
run the old tests and old reference checks plus the new ones. Coverage may not
be weakened. A genuine test defect needs a stated reason and a stronger
replacement, not a deleted assertion to get green. Record pre-existing
failures separately; do not claim a passing legacy regression suite if it
did not pass.

Check fixture model IDs against every implementation the fixture exercises.
Frozen solutions may carry their own model tables; updating the main model
catalog does not update those snapshots. Keep local fixture identity separate
from discovered live model identity, and test compatibility in both trees
before changing a shared fixture. The coding skill cannot prove that agreement.

When behavior is missing, update the teaching and its check, then rebuild.
Fixes must survive regeneration by the next student. Green checks alone are
not evidence that the complete feature is usable.

## 4. Actually take it for a spin

**The coder must run the chapter's real user interface with a real model
backend and exercise every chapter feature. Initially that interface is the
CLI.** This happens before the author writes the chapter's demonstration.
From Chapter 2, the CLI has a human chat mode: ordinary text, visible prompts,
readable answers, and usage. Exercise that mode in an actual terminal/PTY,
waiting for its prompts and reading its answers. The coder may drive the PTY,
but may not replace this interaction with JSON-lines protocol input or call it
a session Bill personally ran. Retain machine-protocol runs as separately
labeled interface evidence. Human chat cannot be deferred to the jobs chapter.
Fake-server grading, internal function calls, scripted JSON conversations,
and plausible output cannot substitute for the human-mode live interaction.

Create a feature checklist with a concrete user action and observable result
for each capability. Run the actual snapshot's executable. Record:

- The snapshot/commit or exact source state and build command.
- Actual run date, provider, model, mode, and sanitized invocation.
- User inputs, observed outputs, and relevant measured usage.
- The result for each feature, limitations, and failures needing correction.

Test all three initially supported vendors, Anthropic, OpenAI, and Gemini, as
their adapters are introduced. A single-provider chapter tests its supported
provider; the multi-provider chapter tests all three. Use provider discovery
and current official documentation rather than remembered model names.

Bill authorizes API keys in `~/.cr/settings.json`. Read only needed values
programmatically. Never print the file or keys, put secrets in command
arguments or prompts, record credential-bearing request logs, copy keys into
repository files, or commit them. Keep secrets in memory or child process
environments. Sanitize retained evidence. Keep demonstrations bounded.

Deterministic fault injection remains necessary for failures that a live
provider does not reliably produce. Label those tests separately from the
live run; never invent a provider failure or silently replace required live
capabilities with fake coverage. Label the Chapter 2 GUI stub honestly.

Fix real-user failures through the teaching and regression checks, then repeat
the affected demonstration. If provider access or another external blocker
prevents a required live check, leave the chapter's validation incomplete and
record the blocker. Do not write a successful transcript anyway.

At the caching chapter, verify the current official OpenAI subscription-access
method and re-test the OAuth caching defect Bill reported. Use fresh,
repeatable measurements and API-key controls where comparable. Record endpoint,
model, credential mode, cache counters, and uncertainty without recording
credentials. A historical report and a closed bug are not measurements.

## 5. Compare with the first edition and improve the answer

Passing is necessary, but it is not the objective of the second edition.
**After the coder has implemented and run the chapter, an independent code
reviewer must compare the new solution with the corresponding first-edition
standard solution.** Preserve the initial student attempt before this review.
The reviewer reads the old implementation; the student receives findings and
rationale, not an answer key to copy. This keeps the initial build a test of
the new teaching while allowing the completed answer to improve through review.

Record the two snapshot commits and the corresponding chapter numbers; old
Chapters 6–21 map to new Chapters 5–20. Review at the same feature scope, using
later first-edition corrections as evidence where they bear on today's design.
The corrected architecture remains authoritative: an old shortcut does not
become desirable merely because the standard solution used it.

Compare:

- Design, data ownership, dependency direction, and room for later features.
- Code clarity and economy: unnecessary machinery, duplicated work, awkward
  control flow, and avoidable complexity. Fewer lines alone is not better.
- Comments: explain intent, invariants, and surprising choices accurately;
  preserve useful reasoning and remove misleading or redundant narration.
- Behavior, error handling, public usability, tests, and diagnostic access.
- Teaching: whether the prose and instructions enable a student to produce
  the better design without private hints from the reviewer.

Give actionable findings with locations, reasons, and expected benefits.
Identify strengths worth retaining from either version, regressions in the
new answer, and opportunities beyond test compliance. Record concrete
improvements over the first edition, plus any justified tradeoffs. Do not
manufacture edits when a comparison supports keeping the new code as written.

The coder revises the answer; the author incorporates teaching findings into
the chapter and contract. Rerun affected checks and live demonstrations when
the revision changes demonstrated behavior. The code reviewer then checks
the revisions and records which findings were fixed or declined, with reasons.
A score of 100 does not close this stage. Unresolved material quality or
teaching findings prevent chapter validation.

## 6. Author reconciliation and independent proofreading

The coder returns a post-build review: actual implementation, discovered
contract gaps, grader changes, checks run, exact failures, and live receipts.
The author reconciles outline and chapter with that evidence.

Write "Taking it for a spin" using the actual run. An abridged transcript must
preserve what occurred. Explain how the reader reproduces it. Where a GUI is
part of the chapter's demonstrated feature, capture the actual interface and
provide an accessible text description. Do not fabricate screenshots or require
a GUI screenshot for a chapter whose GUI is explicitly a stub.

The Reviewer independently reads the complete chapter, current voice rules,
this procedure, and the coder's evidence. Check:

- Plain-English explanation, logical order, terminology, and necessary detail.
- Motivation for the rules and a reader-facing reason to build the mechanism;
  specifications and a clean lint run do not by themselves establish voice.
- Voice budgets, repetition, unsupported claims, and stale cross-references.
- Agreement among TL;DR, prose, exercises, checks, and demonstrated behavior.
- Architectural consistency and forward lessons, consulting the global map.
- Every live demonstration against receipts and the all-feature checklist.

Send actionable findings with source locations to the author. The author
revises; the Reviewer checks resolution. A draft can be proofread while live
work is pending, but cannot be labeled complete without required evidence.
Run prose lint against the new chapter, not every unrelated first-edition file.

## 7. Checkpoint and continue

A validated chapter has a successful independent student build, required
checks and deletion audits, retained legacy coverage, actual live feature
demonstrations, a recorded comparison with the first-edition standard,
completed code/teaching revisions, and resolved reviewer findings. Record
what improved rather than inferring quality from a passing score.
Bill's editorial approval is
recorded separately.

After every validated chapter, make a dedicated commit in the outer Ensemble
repository and add an immutable annotated tag, for example
`edition-2-ch03-r1`. The chapter checkpoint includes its manuscript, the matching
main source and chapter snapshot, changed graders and skills, and a validation
manifest linking source identities and evidence. Keep the worktree free of
unrelated changes in that commit; stage the intended files explicitly.

Export `solutions/edition-2/chNN/` from the exact validated main source version;
exclude nested Git metadata, credentials and unrelated runtime files. Record
the source tree identity, export hashes, contract/check versions and evidence
binding in the validation manifest. The dedicated outer commit and annotated
tag preserve the matching manuscript, main source and export together. A later
validated export may update that chapter's tracked directory in a new revision;
the previous commit and tag remain unchanged. Students never edit the export.

The earlier standalone student repositories and any unfinished work must be
preserved before consolidation moves them. Record the migration mapping and
original commit IDs; do not rewrite old transcripts, source hashes or local
repository identities to make them look like later outer commits. This policy
describes the target layout, not evidence that migration has completed.
Consolidating unfinished source does not validate it: preserve its pending
status and withhold the validated chapter tag until its remaining gates pass.

Preserve the initial student attempt before comparative review. Work-in-progress
commits are welcome, but a validated chapter tag requires every chapter gate.
A later correction gets a new commit and revision tag, such as
`edition-2-ch03-r2`; never move the earlier tag or relabel its evidence.
Carry earlier corrections forward through affected later chapters and the
main source, revalidating each affected version before tagging it.

Record status, exact commands, remaining issues, and the next action in
`book/edition-2/progress.md`. Checkpoint unfinished work honestly if interrupted.
Local commits and tags do not authorize a push or a published release.

Stage only owned changes explicitly. Never `git add -A`, stage Bill's
unrelated files, mutate first-edition solutions for an audit, or push.
Continue autonomously to the next chapter when the current work meets these
conditions. Ask Bill when a real unresolved decision requires his judgment.
