# Chapter writing procedure

## Governing method: remaining Edition 3 chapters and Edition 4

Bill has selected the [Teach Your AI system](https://github.com/waywardgeek/teach-your-ai)
for the remaining applicable Edition 3 chapters and for Edition 4. Its local checkout is `/Users/bill/projects/teach-your-ai`;
its chapter-writing procedure governs human-led chapter creation, and its
book-execution workflow governs fresh-student verification. This is the chosen
method, not a claim that Edition 4 has started or that its chapters are verified.
The dated Edition 2 phases below are historical; their useful craft guidance
remains, but they do not override this choice or restart an earlier run.

Keep the two activities distinct. During creation, the human works with separate
top-level author and reference-coder conversations: define the need, TL;DR,
plain-language idea and exercise; implement and test the reference and external
grader; then complete the chapter from demonstrated experience. A successful
reference implementation is not evidence that the completed chapter teaches it.
Verify that with a fresh student using the completed chapter and accepted
prerequisites—not the reference solution, prior target candidate, private review
history or solution-specific coaching. The student may run the supplied external
grader and inspect its reports throughout development; grader source remains
private.

The student owns a complete implementation, tests, grader feedback, authorized
exercises and one public-input reviewer’s review-and-fix cycle before submission.
No routine early compiling review or coordinator relay is required. The reviewer
uses the same input boundary as the student and recommends readiness, not acceptance.
The coordinator independently assesses the finished result and exercises the app.
Keep review consolidated, rerun only affected checks, and keep notes concise.

The student also writes and runs relevant tests and follows the chapter's
“Take it for a spin” steps through the actual CLI or browser, within authorized
limits. Producing code or passing a grader is not completing the exercise.
The coordinator independently performs the exercise for acceptance; neither
pass replaces the other. Record observed outcomes and any blocked, deferred or
unverifiable steps honestly in the existing chapter notes, not as completed work.

If the completed student work is not good enough, the chapter is not good enough
yet: diagnose the failure, improve the teaching and test it with a fresh isolated
student from accepted prerequisites. Do not repeatedly coach that candidate through
final acceptance. There is no numerical limit on teaching-improvement/fresh-student
cycles; historical correction-count stops do not govern this method. Distinguish
setup and grader failures honestly, and preserve operational/budget limits across
runs. Substantial design/contract changes, major teaching restructuring or solution-specific implementation coaching require fresh verification from accepted prerequisites. Faithfully incorporated clarifications may retain the successful run as verified with recorded Q&A under the rule below; asking a question alone does not invalidate a run.
Keep each chapter's requirements, prose, implementation and grader coherent and
adequately reviewed/validated before advancing. Grader green is not sufficient;
inspect ownership, correctness, simplicity and actual behavior, and report limits
honestly. Consequential unresolved choices and authority blockers go to Bill;
ordinary authorized teaching improvement does not require another permission loop.

Apply **KISS (Keep It Simple)** and **YAGNI (You Aren't Gonna Need It)** to code,
prose, tests and coordination. Meet the current need with existing owners,
helpers and suitable libraries; do not add speculative features or frameworks.
Preserve necessary safeguards and explanations of non-obvious decisions. Fewer
lines alone are not better code or better teaching. Repeated correction cycles
should prompt examination of the teaching and our assumptions, not more ceremony.

Edition 3 applies this method through [its execution procedure](../docs/edition-3-execution.md):
prepare the existing chapter, then verify it with a fresh isolated student who can
run the grader, and establish code/prose quality before advancing. Chapters 1–20
remain closed; Chapter 21 is the first current application, not a special quality
exception. Keep Chapter 25 historical-only and the separately installed Chapter 27
work distinct from regenerated-student acceptance. This does not restart closed
chapters or add provider, installation or publication authority.

## Separate learning from specification

Make the chapter efficient to learn from and precise to implement from. Keep two
clearly labeled, connected parts rather than mixing a tutorial, exhaustive contract
and debugging history into one stream:

- **Learning:** the problem, simple mental model, ownership, why the design works,
  tradeoffs, and a small worked example. Introduce concepts in a coherent order and
  preserve explanations of non-obvious decisions. Teach transferable judgment,
  not the chronology of every correction.
- **Specification/reference:** authoritative interfaces, formats, configuration,
  errors, invariants, compatibility and edge cases needed to implement and grade.
  Give each requirement one authoritative home and link to it from the explanation
  and exercise. A concise TL;DR can index or summarize it without becoming another
  independently maintained contract. Tests and graders must not add hidden rules.

The exercise applies the learning to that specification. Neither part replaces the
other: exhaustive fields do not teach a design, and a pleasant narrative does not
settle an implementable contract. Keep examples, specification and grader consistent;
separation is organization, not permission to omit necessary teaching or detail.
Use existing chapter sections when they suffice; do not impose a new document set,
rigid length target or wholesale rewrite on an adequate chapter.

## Student questions and recorded Q&A

Welcome questions. Asking for clarification is part of learning and sound engineering,
not a failure of independence. Do not make students guess the teacher's intentions,
ration questions, or wait for formal review. Answer assignment ambiguities from the
actual requirements and explain the reasons; distinguish clarification from supplying
an implementation or privately repairing it past acceptance.

Use a **“Questions from the student”** section for useful real exchanges, not an
invented FAQ or a transcript dump. Preserve the actual question and authoritative
answer verbatim. Check them against the saved exchange; do not reconstruct quotations
from memory or silently edit them. Clearly mark excerpts, omissions or redactions.
If an answer was wrong or withdrawn, include its correction or select the final
correct exchange; do not publish the withdrawn advice as guidance. Reconcile any
contradiction in the main teaching/specification rather than create a competing
contract. Keep concepts and reasons in the narrative and exhaustive requirements in
their authoritative specification section; Q&A helps readers navigate both.

Select reusable teaching: ambiguity, ownership, rationale and practical interpretation.
Exclude secrets, private grader/solution material and operational chatter such as
machine-local paths, progress reports and spending permissions. A chapter with no
useful actual exchanges needs no fabricated questions. Keep the exchanges in the
existing conversation/notes for the author; no new logging system or student paperwork.

If a student succeeds using the chapter plus those clarifications, and the chapter
faithfully incorporates them, retain that run as **verified with recorded Q&A**.
Check that the answers were available before the affected successful work/checks and
that the published wording carries the same information. This is verification of the
chapter with its Q&A, not an unassisted run or proof that the original page was sufficient.
Adding faithful clarifications alone does not require another student run. Existing
code-quality, review, exercise and final-acceptance requirements remain unchanged;
Q&A cannot turn unfinished, failing or unverified work into success.

Substantial design/contract changes, major restructuring of the teaching, or
solution-specific implementation coaching still require fresh verification from
accepted prerequisites. Do not disguise a supplied repair as a clarification to avoid
that test. Actionable review findings against existing public requirements and
student-authored fixes remain ordinary development, not automatically supplied-solution
coaching or a reason to restart. Explain the distinction briefly in the existing record.
If completed work remains inadequate, improve the chapter and run a fresh student; asking questions by
itself is not inadequacy. Operational budgets and permissions never reset.

## October 12: Edition 2 is the current top-level tree

Promotion commit `564e316f` moves the accepted second edition to `agent/`,
`book/` and `solutions/chNN/`. The first edition, frozen through original
Chapter 27, is preserved in `edition-1/`, including its original graders.
Develop current application code in `agent/`; do not edit frozen snapshots.
Current grading uses `cmd/grade-edition2`; original grading runs from `edition-1/`.

Gateway implementation and prose/grader enhancement are accepted as current
Chapter 26. Optional Chapter 24 is complete; crossover Chapter 25 remains a
placeholder. Publication preparation is active, with crossover begun rather
than complete. Preserve Bill's unfinished crossover edits and historical results.
See `docs/edition-2-publication.md` (repository-root path) for relocation and checks.

This current path/phase declaration supersedes older dated instructions below.
Their literal command paths and source identities describe the runs at that time;
relative evidence links have been redirected to the relocated artifacts. In
historical student restrictions, old `agent/` means today's `edition-1/agent/`,
and old `solutions/edition-2/main/` means today's `agent/`. No older instruction
to start from empty, stop at an earlier chapter or use the retired author-first
workflow is reactivated by this move. All architecture, full-context reading,
independent review and preservation rules remain in force.

## Current phase and authority, October 11

The student implementation and comparative reviews are complete through original
Chapter 26, exported as Edition 2 solutions 1–23. Bill now authorizes the
remaining manuscript/grader enhancement, beginning with Edition 2 Chapter 9;
Chapters 0–8 are already accepted. Use `docs/agentic-codebook-workflow.md` and
`docs/edition-2-enhancement.md` for the current sequence, assignments and gates,
and `docs/edition-2-chapter-map.md` for numbering. The earlier author-first
sequence, chapter count, deleted skill paths, expanded acceptance machinery and
initialization instructions below are historical and do not restart the student
run. The mandatory current skill is `.agents/skills/ensemble-coding/SKILL.md`.
Existing enhancement notes remain the chapter records; do not create parallel
validation ledgers or overwrite accepted solution exports.

Bill requires the author to hold all four full documents in active context:
this procedure, `book/voice.md`, `docs/agentic-codebook-workflow.md`, and
`docs/edition-2-carryover.md`. After every compaction, reload all four completely
before writing again. A summary is not a substitute. The independent reviewer
also reloads all four before reviewing after compaction. Preserve the human story,
explanations and right chapter length; there is no chapter-trimming target.

The remaining craft and evidence guidance applies under that current workflow.
The original text below is retained to preserve the procedure's history.

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

- **Author:** owns outlines, chapter prose — including each chapter's
  plain-words explanation required in §2 below — and evidence reconciliation.
  Reads history to find lessons that belong earlier. Does not implement the
  student's solution, and does not own the TL;DR contract (see Grader
  engineer).
- **Student coder:** builds from the new chapter and earlier new baseline,
  reports gaps, tests the result, and actually uses the CLI with real models.
  Does not edit chapter prose or consult first-edition chapters or old answers.
- **Grader engineer:** writes the TL;DR contract and derives its checks
  together, independently of the student's implementation, and freezes the
  TL;DR's final wording against the actual grader fixtures. The coordinator
  can fill this role.
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

Agent slots may require scheduling roles in phases. One reviewer can cover
global review, code review and proofreading after saving its whole-book
findings. If capacity requires reassigning a role, carry that durable record
forward and disclose the replacement's prior source exposure. The code
reviewer must remain independent of the current chapter's implementation.
Record actual assignments in the progress log; do not describe several roles
on one thread as separately spawned workers.

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

Apply the required story-preservation pass in `voice.md` §4.1. Record which
documented human incident or memorable explanation from the first edition
will survive, its source, and the design decision it motivates. If none fits,
record why and identify the concrete reader problem that carries the chapter.
Revisit already-drafted second-edition chapters under this rule too.

## 2. Write the student-facing contract

Write enough of the chapter before implementation that the student can build
from the teaching. Chapter 1's plain-English rules precede all code. Each
chapter explains the idea in plain words before its detailed mechanism.

Contract authorship is split by ownership of consequences. The grader
engineer writes the TL;DR, because the TL;DR is the page contract and the
fixture is its ground truth: the role that builds the checks is the only one
that can guarantee the published words and the graded behavior never drift
apart, and that drift is the defect class this split exists to remove. The
author writes the plain-words explanation and the narrative around the
contract: the why belongs to the author, the what belongs to
the engineer. The author may propose voice edits to the TL;DR, but any change
in meaning goes back through the grader engineer and its checks before
publication. The TL;DR's final wording is frozen against the actual grader
fixtures, never against a design doc alone.

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

Before implementing new ownership or concurrency boundaries, the student adds
a short data-structure plan to its existing student review: owners and parent
interfaces, authoritative facts versus derived snapshots, mutable versus
creation-only configuration, synchronization and lifetime. Coordinator or
reviewer checks that plan against the new contract before affected code.
Independent work continues during that check. Keep this review narrow; do not
design the implementation for the student or introduce first-edition answers.
Record assistance honestly. Historical comparative review still follows the
initial implementation and actual runs. For a chapter already underway, review
the current design at the next boundary rather than restart the exercise.

When the student has difficulty, record it immediately in
`solutions/edition-2/main/evidence/chNN/student-review.md` and notify the author
and coordinator. Name the chapter passage, attempted interpretation, observed
failure or conflict, and the clarification needed. Do not wait for a passing
build or abandon the record because the chapter cannot yet be completed.
Stop affected implementation when ownership or requirements are unresolved;
continue independent work. Preserve the original question and dated resolution.

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

Complete that checklist before the first paid demonstration, including human
CLI, public embedding, multi-Agent and optional GUI paths required by the
chapter. Name the provider coverage and expected receipt for each feature;
identify deterministic-only faults separately. Coordinator or reviewer checks
coverage before launch. This is a completeness check, not a new user-approval
ceremony. Finish known local failures before spending on their live paths.

Reuse validated launch, capture and verification machinery from permitted
earlier second-edition work where it fits. This does not open historical
answers, grader internals or reviewer-only scripts to the student. Adapt
chapter-specific actions without copying old solution code into the student's
implementation. Before paid runs, exercise changed evidence tooling
with local fixtures. A verifier must reject empty/incomplete source bindings
and check every required source, executable and launch identity before any
replay or derived write, including identities for the last provider in a batch.
Resolve repository paths from an explicit repository root. Preserve raw
receipts separately from derivatives.

Verifier negative controls need a passing fixture with valid paths. Mutate one
identity at a time and require the intended refusal, with originals unchanged.
An earlier path or setup failure cannot establish a later identity check.

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

Keep one consolidated findings list per review round in the existing code
review record. Report blocking defects immediately; group compatible fixes
before freezing the next runtime and repeating affected live demonstrations.
Do not delay a needed correction merely to fill a batch. Before each rerun,
record which changed behavior it verifies. Retain applicable earlier evidence
with its original source identity and explain its scope; never relabel it as
a run of the new binary. Prose-only or evidence-only repairs normally need
their own local checks, not another paid generation. Required chapter checks,
legacy coverage and independent review remain mandatory.

## 6. Author reconciliation and independent proofreading

The coder returns a post-build review: actual implementation, discovered
contract gaps, grader changes, checks run, exact failures, and live receipts.
The author reconciles outline and chapter with that evidence.

Require a distinct **student review of the teaching**, even when the build
passes without questions. The coder owns
`solutions/edition-2/main/evidence/chNN/student-review.md`; its purpose is to
tell the author what it was like to execute the chapter. Record the chapter,
skill and predecessor versions actually used, what explained the task well,
missing or conflicting instructions, difficult steps and failed approaches,
help received, and concrete suggestions for better explanations or examples.
Include what actual human-interface use revealed. Distinguish a teaching gap,
a bug in the student's implementation, a grader defect and a provider behavior;
do not turn every difficulty into a demand to relax the contract. No finding
is required merely to fill a template.

Preserve the initial student review before first-edition comparative feedback;
append the effects of clarification and revision rather than rewriting that
initial account. Update it at completion, or leave an interim review when
blocked. This is separate from the independent reviewer's assessment of code
quality and the final test/run summary.

The author reads that report and records each material finding's disposition
in `book/edition-2/chapter-NN-student-feedback.md`: chapter change and location,
clarification of an intentional choice, or a reasoned deferral/decline. The
coder checks whether the response resolves the original difficulty. The
coordinator verifies that this exchange exists; unresolved material teaching
findings prevent chapter validation. If a review is collected retrospectively,
label it with its actual date and exposure rather than presenting it as an
initial cold-student report.

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
- Preservation of documented human stakes and personality under `voice.md`
  §4.1: compare the old and new telling, check the outline's retention or
  removal decision, and flag a rewrite that keeps facts but loses their point.
- Voice budgets, repetition, unsupported claims, and stale cross-references.
- Agreement among TL;DR, prose, exercises, checks, and demonstrated behavior.
- Architectural consistency and forward lessons, consulting the global map.
- Every live demonstration against receipts and the all-feature checklist.

Send actionable findings with source locations to the author. The author
revises; the Reviewer checks resolution. A draft can be proofread while live
work is pending, but cannot be labeled complete without required evidence.
Run prose lint against the new chapter, not every unrelated first-edition file.

## 7. Checkpoint and continue

The coordinator creates `book/edition-2/chapter-NN-validation.md` at chapter
start as the authoritative current gate record. Keep a compact table of
contract/design review, implementation/local checks, initial live use,
comparative review/revisions, final live evidence, manuscript/feedback and
checkpoint. Each row names its responsible role, current status, supporting
revision or receipt, and next action. Update when a gate changes, not after
every command. Stages may overlap; implementation finished and chapter
validated are distinct statuses.

Other records retain their separate purposes and historical chronology.
`progress.md` and `CHECKPOINT.md` link the gate record with a short current
summary; prose and outlines should link it rather than repeat changing task
lists. Reconcile existing current-summary contradictions, but preserve dated
failed or pending attempts. No new parallel status ledger is needed.

When workers report completion, the coordinator checks the remaining gates
and promptly performs the next authorized handoff. Prepare export paths and
manifest inputs while final review runs; export/tag only after acceptance.
Once checkpointed, start the next fresh student without waiting for Bill to
ask for status. If worker capacity blocks a handoff, record the role awaiting
a slot and reuse an available role with its source exposure disclosed.

A validated chapter has a successful independent student build, required
checks and deletion audits, retained legacy coverage, actual live feature
demonstrations, a recorded comparison with the first-edition standard,
completed code/teaching revisions, and resolved reviewer findings. Record
what improved rather than inferring quality from a passing score.
A retained student teaching review and recorded author responses to its
material findings are required too; a success-only handoff is incomplete.
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

Check Go package discovery in every module of the complete delivered tree,
including retained evidence and support files, before tagging. A runtime-only
test copy that excludes evidence cannot prove that `go test ./...` works in the
reader's snapshot. Isolate standalone evidence helpers with an explicit build
constraint or their own module, and verify their documented explicit invocation
separately. The Chapter 7 first export exposed this distinction.

Commit the source first, then export it from the repository root with absent
destination and manifest paths:

```sh
python3 scripts/edition2/export_snapshot.py SOURCE_COMMIT /tmp/ensemble-chNN-export --manifest /tmp/ensemble-chNN-manifest.json
```

The tool exports committed `solutions/edition-2/main/` and records source and
file identities. It does not validate, stage, replace a snapshot or tag. After
the gates pass, verify and promote the export and manifest into the chapter
checkpoint.

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
Agents in this workspace share Git's index. An explicit `git add` does not
isolate a later unqualified commit from another agent's staged work. Commit
with the same explicit owned pathspec, and inspect the resulting file list.
Coordinate a commit that intentionally spans owners. If a race captures
another agent's work, disclose it and bind subsequent evidence to the actual
commit; do not silently rewrite history to make the handoff look cleaner.
Continue autonomously to the next chapter when the current work meets these
conditions. Ask Bill when a real unresolved decision requires his judgment.
