---
name: ensemble-coding
description: Mandatory architecture, KISS and comparative-review rules for Ensemble coders and code reviewers using first-edition chapters and original graders.
---

# Ensemble student coding

## Current Edition 3 execution

For Edition 3, `docs/edition-3-execution.md` governs phase ordering, lean reporting and the required direct reference comparison. The orchestrator must personally compare each chapter's solution with the corresponding original under `solutions/`; material unjustified bloat or poorer clarity triggers the shared targeted TL;DR correction and fresh re-execution procedure. Historical instructions below do not authorize extra reporting campaigns, obsolete author timing, or student access to reference answers. Keep comparison findings concise in the ordinary review.

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


**October 11 supported-model restriction.** Bill requires older supported
Anthropic models such as Sonnet-5, with no 5.5 models in current examples,
defaults or future tests: he reports they do not work with ephemeral context.
Do not add models merely because discovery lists them. Preserve historical
identifiers/results and frozen exports; do not relabel earlier runs. GPT-6
Astra/high remains the explicitly selected coding-agent model.

**October 11 author-phase release.** The student implementation and independent
comparative reviews are complete through original26 / Edition2 Chapter23. Bill
explicitly releases remaining manuscript enhancement from Edition2 Chapter9,
with author/coder collaboration and independent prose/grader review. Follow
`docs/edition-2-enhancement.md`; preserve original graders and frozen solutions,
and record revised grader results separately. Earlier student-phase prohibitions
on authorship/grader revision do not prevent this authorized phase.
Bill requires the author, and the reviewer, to load the complete workflow,
carryover, `book/voice.md` and `book/chapter-writing-procedure.md` into active
context and reload all four after every compaction before resuming their work.

**October 11 merge authorization.** Bill releases the original Chapter 22/24
placement stops and delegates the merges: original 21 + 22 → Edition 2 Chapter
20; original 23 → Chapter 21; original 24 + 25 → Chapter 22; original 26 →
Chapter 23 if completed when the review of original 25 finishes. Follow
`docs/edition-2-chapter-map.md`; original exercise, review and evidence numbers
stay unchanged. Preserve every valid capability and the normal student/reviewer
gates. The accepted combined `ch20` replaces its prior original-21-only version;
the old immutable tag preserves that source. Authorship follows implementation.

**October 11 model instruction.** Bill selected GPT-6 Astra (`gpt-6-astra`) at
high reasoning for all coding tasks. Use it for new coding/code-review delegates
and Ensemble live coding work. This supersedes the earlier GPT-5 Astra name and
releases the crossover model-choice stop. Keep historical model/test evidence
intact; follow the current merge authorization above for subsequent chapters.

**October 10 bounded enhancement authorization.** Bill now authorizes an author
and coder to enhance original Chapters 1–9 into Edition 2 Chapters 1–8, with
Chapter 0 guidance and independent prose/code review for each chapter. For this
pass, the author-phase timing and no-grader-edit restrictions below are superseded
by `docs/edition-2-enhancement.md`. Preserve original graders; create separately
versioned Edition 2 corrections. Enhance existing prose without rewriting or
adding features. All architecture, KISS, no-mocks and review requirements remain.
The restrictions continue to govern the original student run beyond Chapter 9.

Read this entire skill before coding or code review and after compaction. The original chapter
defines feature scope; Bill's instructions define architecture. Do not import
requirements from the discarded edition.

Also read `docs/edition-2-carryover.md` before starting and after compaction.
It replaces the revised Chapter 1 as the student's upfront lessons document;
follow the unchanged first-edition chapters. The coordinator completes its
full-edition reading and records coverage before the student run. Carryover
lessons guide the current exercise, not early implementation of later features.

## Architecture from the first line

- Shared core data and interfaces belong in `internal/common`. Implementation
  spokes import common, not one another; common imports no implementation spoke.
  The composition root constructs the objects; clients use its public API.
- Put behavior in the responsible directory. Use free functions over common
  data when Go's receiver rule would put behavior in the wrong package.
  Standard-library dispatch methods such as Error and JSON methods may remain
  with their types. Common is not a miscellaneous utility/behavior package.
- Private runtime structs may live in their implementation packages. Each child
  constructor receives and stores interface back-pointers to all required parents;
  required parents cannot be nil. Optional parents may be absent and add the
  child later. Adding a child updates both sides: the parent records the child
  and the child sets its back-pointer to that parent. A pin must belong to an IC
  package but may float without a wire; connecting a wire establishes the second
  parent relationship. Declare parent interfaces in common and expose parent
  relationships through interface methods. Multiple parents are allowed; the
  application root has none. Service dependencies are not invented parents.
- Reach configuration, services and logging through those parent relationships.
  Keep facts with their owners, not copied into siblings or supplied through callback bags.
  Parsers and helpers likely to need diagnostics receive owner access too.
  No mutable global logger, registry or application state.
- One Ensemble owns potentially many Agents. Agent owns its configuration;
  Engine owns usage. Tool visibility is per-Agent; registry storage may belong
  to Agent or Ensemble. Introduce only objects needed by the current chapter.
- Streaming/display observations reach Ensemble through Observer. Explicit
  parent methods may provide services; Observer is not the only communication.
- GUI and WebSocket implementation belongs in a separate optional Go module
  consuming public core interfaces. CLI and GUI use the same Agent. From Chapter
  2 retain the requested client separation with a small GUI stub until its
  chapter. Later MCP permits different transports, including the GUI WebSocket
  tunnel; do not implement it early.

## Concrete KISS choices

These choices do not relax the mandatory star architecture: keep implementations and private state with their owning package; put cross-package interfaces in `internal/common` to break dependency cycles; compose and wire components at the Ensemble root. A smaller reference solution is not permission to collapse ownership or introduce sibling-package coupling.

- Inspect existing owners, helpers and dependencies before adding another mechanism. Prefer the smallest clear change satisfying the current chapter; do not build for hypothetical later chapters.
- Keep private state and behavior with their natural owner. Shared interfaces/data belong in common only when genuinely shared; avoid unrelated capability collections and plumbing that merely moves the same state through more layers.
- Reuse established protocol libraries and existing infrastructure where the chapter permits them. Hand-writing a protocol, queue, dispatcher or compatibility layer needs a concrete requirement, not merely the ability to implement it.
- Make necessary departures understandable in code and the short chapter review. Review maintainability as well as correctness; passing tests do not excuse needless machinery. Preserve useful comments and explicit control flow instead of pursuing short line counts.

## Student, review, then checkpoint

Read the original chapter, implement its exercise, and run its original grader.
The student does not consult existing answers, future chapters, discarded code
or grader implementation to infer unstated features. Later architecture-repair
chapters do not require deliberately introducing flaws now.

When public instructions omit a needed input/output contract, ask the coordinator
for the missing information. Bill authorizes the coordinator to inspect the
grader and supply a narrow, sourced erratum without asking him again. Record
exactly what the student received in `docs/edition-2-notes/chNN.md`, including
source revision and the teaching gap. This does not authorize student access to
grader internals, old answer code, implementation recipes or extra features.
Conflicting requirements still use the critical-question or review-exception
procedure; a grader expectation does not automatically override the chapter.

Both coder and reviewer use **KISS: Keep It Simple** as the yardstick. Choose
the smallest clear implementation meeting the exercise and these rules.
Explain non-obvious choices in comments; compressed statements are not simplicity.
This is teaching code. Bill asks for roughly 20% comments (about one comment
line per four code lines, excluding blanks). Explain why, ownership, lifetimes
and important ordering beside the relevant code so students need not reverse
engineer the design. Treat the percentage as a teaching target, not a padding
quota; useful comments are not bloat.
Every exported Go symbol must have a useful documentation comment, including
those in internal packages: types and aliases, functions and methods, constants,
variables, struct fields and interface methods. Explain its purpose and relevant
contract; include ownership or lifetime constraints where they matter. Merely
restating its name is not documentation. This is required independently of the
roughly 20% teaching target.
Do not introduce generic frameworks, extra protocols, invented limits, speculative
recovery or exhaustive audit machinery merely because they might be useful.
There is no author agent and no grader agent working ahead of the student.

Record teaching gaps and suspected grader defects for future maintainers; do not
rewrite either during this attempt. Ask Bill about consequential ambiguities.
Critical questions are stop points: stop work and route the question to the
coordinator, who pauses active delegates, asks Bill plainly and waits for his
answer. Do not continue under an assumption or rely on Bill watching progress
and interrupting; Codex does not support the real-time collaboration he needs.
An explicit unpause from any tab unpauses the Agent for everyone; other tabs
cannot retain independent vetoes.
Bill's interrupt ruling: end the current turn so the user can start a new one;
keep the Agent alive and other jobs running. Preserve their real results.
Interrupt is not job kill or shutdown. Apply this when Chapter 6 introduces it.

Run gofmt, go vet and go test in affected modules, and the original chapter
grader. Use focused regressions for actual bugs, not a new grading framework.

**No mocks without Bill's explicit approval.** Mocks can be useful in rare
situations, but Bill expects none in this code. Use the grader-provided fakes
where available and suitable, or small additional fakes when needed. If the
coder believes a mock is necessary, stop and explain the test need, why those
fakes cannot meet it and the proposed mock's scope. Route this critical question
to Bill and wait for approval before introducing the mock. Reviewer acceptance
of an exercise or grader exception cannot authorize one. This applies to
handwritten mocks as well as mocking libraries.

Exercise the real implementation and replace external
boundaries with small, faithful fakes when deterministic tests need them. For
example, a fake HTTP server can supply vendor responses while the real client,
parser and engine run. Assert observable results and relevant boundary behavior,
not a scripted sequence of internal method calls. Do not mock away the behavior
being tested or introduce interfaces and mock frameworks solely for tests. Keep
fakes as small as the exercised contract permits; they do not replace live runs.

A chapter boundary is not a test-double boundary. Later chapters exercise the
real earlier implementation through its interfaces. For example, Chapter 12's
coder may not know Chapter 4's implementation history; mocking its job machinery
would substitute today's assumptions for the behavior actually built. Fake the
external peer when needed, keeping the earlier machinery in the exercised path.

**Test the tests with targeted mutations.** For every distinct behavioral claim
relied on in the chapter's tests, deliberately introduce a small, plausible defect
in that behavior and verify that the intended check catches it. One mutation can
cover equivalent assertions; reuse evidence for unchanged inherited behavior.
Focus on the current chapter, not a repeated whole-book mutation sweep.

- Start with the unmodified result. Change one behavior at a time in an isolated
  scratch copy of the student's source. Confirm the mutation applied and the
  relevant code path actually ran. Keep the original grader unchanged; the
  student's restriction on reading its implementation still applies.
- A killed mutant must fail the intended behavioral check for the intended
  reason. Compilation errors, unrelated failures or broken test setup do not
  demonstrate that claim. A timeout counts only when it demonstrates the actual
  timing/liveness property under test, with a working baseline.
- Investigate survivors: the mutation may be ineffective or equivalent, the
  harness may be wrong, or the test may miss the defect. Repair the actual gap
  and rerun the focused check. Never relabel a survivor as killed. If a claim
  cannot be meaningfully mutation-checked, explain the limitation and alternative
  evidence in the student review for reviewer assessment.
- Restore the correct implementation, rerun affected checks and remove scratch
  mutants. Record claim, mutation and observed failure concisely in the existing
  review/evidence; do not build a generic mutation framework or retain large
  duplicate trees. This applies to later grader revisions as well.

Mutation checks establish sensitivity to defects, not the truth of a shared
assumption. Ground expected behavior and fixtures in the chapter, observed
artifacts and independent evidence; do not let code and test certify the same
invented protocol. Keep the existing live-run requirement.

Actually exercise the human interface with a real model as the chapter introduces
features. Bound demonstrations, avoid unattended retries, and retain short
sanitized observations. Use all three vendors as introduced. Discover available
models rather than trusting remembered names. Human chat must be usable; grader
protocol input alone is not a human-interface demonstration.

Bill authorized keys in `~/.cr/settings.json` for live tests. Read needed values
programmatically without printing them:

- Parse the JSON inside the test launcher or application process. If the schema
  is unfamiliar, inspect the settings loader or emit field names/types only,
  never values. Do not use `cat`, a raw `jq` extraction or a file-reading tool
  that returns the credentials to the conversation.
- Select only the needed provider's key. Keep it in memory or set the intended
  test child's environment programmatically. Do not interpolate it into shell
  commands or command-line arguments, and do not enable shell tracing (`set -x`).
- Do not dump settings, environments, authorization headers or credential-bearing
  URLs. Check request/error logging before a live run; redact before output is
  emitted or saved, not after it reaches the tool transcript. Print only safe
  status such as provider name, credential present/missing and test outcome.
- Do not copy the settings file or keys into the repository, fixtures, prompts,
  snapshots or evidence. Keep credentials out of unrelated tool subprocesses.
  Document field names and loading steps, never real values.

Report live access failures accurately; fake-server success does not prove live use.

Leave one concise student review per chapter: source/chapter version, what was
built, actual grader/live results, ambiguities, defects, assistance and suggested
improvements. Distinguish implementation errors from teaching/grader problems.
Preserve failed attempts without building an evidence subsystem.

Any agent may add lessons for the future author to
`docs/edition-2-notes/chNN.md`; create a chapter file when there is a finding.
Attribute the contributor and role, cite the chapter/revision and evidence, and
distinguish observations, suggestions and unresolved questions. The student
review may live there or be linked. Preserve others' findings and corrections.
Notes do not expand the exercise or authorize work. Keep old answer code and
answer-revealing comparisons out of notes supplied to the student.

## Required post-chapter code review

For remaining Edition 3 and Edition 4, the current public brief and chapter procedure
override the historical policy below. The student completes implementation, tests,
grader iterations and authorized exercises, then calls one separate-context reviewer
and resolves justified findings before final submission. No routine early compiling
review. Its reviewer uses only the student’s public inputs: no originals, private
graders or root repair history. It recommends readiness, not acceptance or exceptions.
The coordinator assesses finished work and independently exercises the application.
If the final work is inadequate, improve the chapter and rerun with a fresh student;
there is no numerical teaching/rerun limit and no reset of operational allowances.
Keep review consolidated, reuse evidence and avoid duplicate review/paperwork.
Welcome student questions and explain requirements rather than make students guess.
Use the chapter procedure’s recorded-Q&A rule: faithful published clarifications may
retain an otherwise successful run as verified with recorded Q&A. Asking alone does
not invalidate independence; solution-specific coaching, substantial design/contract
changes or major teaching restructuring still need fresh verification. Preserve exact useful exchanges in
existing notes/conversation, without private source, invented quotes or extra paperwork.

The reference-comparison guidance below applies only to privileged coordinator
acceptance work, never to the student-called reviewer. No reviewer may waive
requirements or accept a chapter on its own.

The reviewer reads the original chapter, the coder's review for the future
author, the new implementation and the matching first-edition solution. Consider
the student's difficulties and interpretations when diagnosing unnecessary work;
do not turn suggestions for future teaching into current requirements. The
reviewer's access to old solutions does not extend to the coder.

Reading the student's chapter review is mandatory before deciding whether the
student passed. Evaluate its explanations for failed tests, ambiguities and
departures alongside the implementation and actual results. State which
explanations you accept or reject and why; do not decide from grader status alone.

Leave a concise comparison covering:

- **Active code:** compare production lines excluding blank and comment-only
  lines, with physical totals for context. Measure the chapter's changes against
  its own predecessor in each edition, not just whole-tree totals. Separate
  tests, examples, generated files and support machinery; distinguish languages
  and account for moved files. State the counting method. Dense one-liners and
  deleted comments are not improvements, and line count alone is not quality.
- **Comments:** compare explanation of intent, ownership and non-obvious choices,
  as well as comment counts. Check accuracy. Preserve useful explanations;
  check that every exported symbol has a useful documentation comment, including
  exported members and symbols in internal packages.
  neither silence nor comments narrating obvious syntax improve maintainability.
  Assess Bill's roughly 20% comment target and whether a student can understand
  the reasons behind the code. Do not accept thin explanations merely because
  the program passes, or filler merely because it reaches the target.
- **Tests:** compare the behaviors and real failure modes protected. Better tests
  are valuable even when longer. Do not reward test volume, implementation-mirror
  assertions or elaborate test infrastructure for its own sake. Enforce fakes
  over mocks, require Bill's explicit approval for any mock, and check that the
  real behavior under test still runs. Review mutation evidence for the claimed
  behaviors, including survivors and limitations; passing tests alone do not
  establish that the tests can detect defects. Preserve the
  original grader's role; grader defects go into a future-maintainer review.
- **Scope and simplicity:** trace added behavior to the original exercise or an
  explicit Bill instruction. Examine data structures and abstractions for the
  current need. Reject feature creep, unjustified bloat and complexity added for
  elegance. Do not justify extra code by promises the workflow itself invented.

The student need not pass a flawed grader test. In its chapter review, identify
the failed check, explain why its expectation or implementation is wrong, and
provide evidence for the student's intended behavior. Do not contort correct
code or add features just to make a broken test pass. Fix genuine implementation
bugs and satisfy valid checks; passing alone is not sufficient.
A suspected bad check or unreasonable requirement goes to the coordinator/human
with concrete evidence. It is not permission for a reviewer to waive the requirement,
mark the chapter accepted, alter the grader or authorize another paid attempt.
A failing or unrun check remains failing or unrun until honestly resolved.

Explain material growth and ask
whether a simpler design satisfies the same actual requirements. Do not impose
an arbitrary line quota or remove required behavior to win a size comparison.
Return concrete findings and rationale without exposing old answer code. State readiness and unresolved findings; final acceptance belongs to the
coordinator/human, not the student reviewer. Do not excuse scope or bloat findings. Ask Bill if a consequential tradeoff
cannot be resolved from his instructions.

Before final submission, the student resolves justified reviewer findings, reruns
affected checks and returns to that same reviewer for verification. Collect and close
that review before submitting. Do not relay routine findings through the coordinator
or impose historical three-/ten-round stopping rules. If the completed work is still
inadequate at acceptance, improve the chapter and run another fresh isolated student.
There is no numerical teaching/rerun limit and no reset of operational allowances.
Essential missing teaching belongs on the page, not in private coaching. Keep concise
notes and useful failures; no new evidence framework. Final acceptance and explicit
commit authority remain separate from a reviewer verdict.

## Teaching preparation and fresh verification

For remaining Edition 3 and Edition 4, prepare coherent teaching before the student
runs. Improve an existing chapter from actual evidence, not a ritual rewrite. Keep
learning (mental model, ownership, reasons) distinct from clearly labeled exhaustive
specification/reference, with consistent examples and one authoritative requirement
location. Prose review checks the writing; only a fresh student using tests, grader
and its reviewer tests whether the chapter teaches success.

If final completed work is inadequate, diagnose the cause, improve teaching and
rerun from accepted prerequisites without the prior target code or private repairs.
Use the appropriate owner for broken fixtures or environment setup; never change
production to satisfy an invalid check. Preserve earlier attempts and results. A
material teaching change needs fresh verification; typo-only edits do not. Do not
accept a privately coached patch as evidence the chapter was sufficient.

## Surviving an interrupted session

Vendor APIs freeze, agents get restarted, and sessions die mid-stage. This has
already happened during ensemble work. The cost of an interruption is decided
entirely by choices made before it, so treat every stage as if the next request
will never return.

Commit each completed stage the moment it is green, with an explicit pathspec.
The reason is that a commit is the only form of work that survives a freeze
intact: a stage held as uncommitted edits plus an explanation in the
conversation loses the explanation and keeps the edits, which is the worst of
both. Pair the commit with a checkpoint at the same boundary, written from
artifacts rather than memory, because reasoning regenerates and measurements
rot. Pick the boundary when nothing is in flight; a checkpoint taken mid-edit
records a plan whose other half no longer exists.

Keep durable state in files, not in tool results. A plan, a stage list, a
decision table or an open-questions list belongs in the design doc or the brief
the moment it exists, because tool results are redacted from context on the next
turn and vaporize completely on a restart, while a file survives both. This is
why the ch24 stage plan lived in `docs/ch24-design.md` rather than in a reply.

Do not park in a long blocking call during flaky vendor weather. A long
watchdog or a blocking join converts a recoverable hiccup into a dead session
that cannot be interrupted from outside, so prefer short waits in a loop where
each iteration is a chance to notice and checkpoint.

When resuming after any interruption, re-establish state from the repository
before touching anything. Run `git log`, `git status` and look at the actual
deliverable on disk. The reason is specific and was observed directly: after a
freeze the conversation context describes the world as it was at the freeze and
reads exactly as authoritative as it did before, while the tree may have moved
many commits ahead, including commits from other agents working the same repo.
A resumed context once reported a stage as pending whose work was already
committed, finished and snapshotted. Acting on that would have redone completed
work on top of newer commits. Memory files have the same failure mode and the
same fix: believe the tree, not the narrative. If the two disagree, the tree is
right and the narrative is stale, every time.

## KISS and YAGNI

Apply **KISS (Keep It Simple)** and **YAGNI (You Aren’t Gonna Need It)** to code, tests and tooling. Implement the current authorized behavior with existing owners and helpers; do not add speculative capabilities, abstraction layers, recovery schemes or compatibility machinery for hypothetical needs. Required correctness and safeguards remain required, and useful explanations of non-obvious decisions should remain. A passing grader does not justify avoidable complexity; challenge a fixture or review demand that adds behavior without a real requirement.
