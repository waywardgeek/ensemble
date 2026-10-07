# Chapter writing procedure

The second-edition workflow for *The Art of Building AI Coding Agents*.
Updated for Bill's instructions in this session. This replaces the old
code-first outline procedure and its per-chapter approval ceremony.

## Outcome and scope

Write exactly **21 chapters, numbered 0–20**. Leave Chapter 0 and the preface
alone. Rewrite from Chapter 1. Absorb the lessons of first-edition Chapters 5
and 22 into the chapters where the decisions first matter; neither survives
as a standalone repair chapter. Chapters 1–4 retain their numbers, and old
Chapters 6–21 become new Chapters 5–20 unless a dependency requires an
explicitly documented adjustment that preserves the agreed total.

The manuscript lives in `book/edition-2/`. New student implementations live
in `solutions/edition-2/chNN/`, each a complete Git repository with module
files, tests, and validated snapshots. Initialize Chapter 1 from scratch.
Continue each subsequent chapter from the preceding **second-edition**
repository and history. Bill authorizes initializing these repositories.

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
  Does not edit chapter prose or consult old implementation answers.
- **Grader engineer:** derives checks from the taught contract, independently
  of the student's implementation. The coordinator can fill this role.
- **Global reviewer:** reads the complete textbook, records dependencies and
  later lessons, and answers questions about the whole-book consequences.
- **Reviewer / proofreader:** reads the author's work independently for voice,
  clarity, organization, consistency, and evidence. Returns concrete findings
  to the author, then reviews the revision. It is distinct from authorship.
- **Coordinator:** sequences the work, resolves source conflicts, relays
  questions to Bill, enforces architectural corrections, and records progress.

Both **Author and Reviewer must read all of `book/voice.md` and this procedure**
before working, and reload after compaction or a rule change. Use the current
voice file where older procedural wording disagrees with it.

Agent slots may require scheduling roles in phases. In this session the
global review thread also serves the dedicated proofreading phase after
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
the mandatory coding skill, current and earlier TL;DRs, the preceding new
solution, and the grader command. Do not omit the architecture from a
"TL;DR-only" test. Do not provide future or first-edition solution code or
grader internals. In a shared workspace this is an instruction boundary,
not a claimed filesystem sandbox. Disclose accidental answer exposure.

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

Run the new solution's build, module tests, vet, and relevant grader.
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

When behavior is missing, update the teaching and its check, then rebuild.
Fixes must survive regeneration by the next student. Green checks alone are
not evidence that the complete feature is usable.

## 4. Actually take it for a spin

**The coder must run the chapter's real user interface with a real model
backend and exercise every chapter feature. Initially that interface is the
CLI.** This happens before the author writes the chapter's demonstration.
Fake-server grading, internal function calls, and plausible output cannot
substitute for a live user exercising the program.

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

## 5. Author reconciliation and independent proofreading

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
- Voice budgets, repetition, unsupported claims, and stale cross-references.
- Agreement among TL;DR, prose, exercises, checks, and demonstrated behavior.
- Architectural consistency and forward lessons, consulting the global map.
- Every live demonstration against receipts and the all-feature checklist.

Send actionable findings with source locations to the author. The author
revises; the Reviewer checks resolution. A draft can be proofread while live
work is pending, but cannot be labeled complete without required evidence.
Run prose lint against the new chapter, not every unrelated first-edition file.

## 6. Checkpoint and continue

A validated chapter has a successful independent student build, required
checks and deletion audits, retained legacy coverage, actual live feature
demonstrations, and resolved reviewer findings. Bill's editorial approval is
recorded separately.

Commit the validated snapshot in its own solution repository. Preserve chapter
history when continuing to the next repository. Record status, exact commands,
remaining issues, and the next action in `book/edition-2/progress.md`.
Checkpoint unfinished work honestly if an interruption requires it.

Stage only owned changes explicitly. Never `git add -A`, stage Bill's
unrelated files, mutate first-edition solutions for an audit, or push.
Continue autonomously to the next chapter when the current work meets these
conditions. Ask Bill when a real unresolved decision requires his judgment.
