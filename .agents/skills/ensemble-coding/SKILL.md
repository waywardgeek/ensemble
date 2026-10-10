---
name: ensemble-coding
description: Mandatory architecture, KISS and comparative-review rules for Ensemble coders and code reviewers using first-edition chapters and original graders.
---

# Ensemble student coding

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
Bill also authorizes the reviewer to read the coder's notes and independently
accept an exception when the grader is flawed or an exercise requirement is
unreasonable. Record the requirement, the coder's reasoning, the reviewer's
agreement, actual grader/live
results, the accepted departure and any effect on later exercises. Mark the
chapter accepted with an exception and proceed to the next chapter without
another permission request. Keep the original chapter and grader unchanged;
acceptance does not turn a failing or unrun check into a passing result. Leave
the teaching/grader correction for future maintainers. Difficulty alone is not
evidence that an exercise is unreasonable; assess the substance of the objection.
This authority does not waive Bill's explicit architecture rules or add features.

Explain material growth and ask
whether a simpler design satisfies the same actual requirements. Do not impose
an arbitrary line quota or remove required behavior to win a size comparison.
Return concrete findings and rationale without exposing old answer code. State
whether the chapter is accepted, accepted with an exception, or needs correction; do not
accept unresolved scope or bloat findings. Ask Bill if a consequential tradeoff
cannot be resolved from his instructions.

The student may revise both code and its chapter review in response to feedback,
rerun affected checks and submit both for reviewer reassessment. Iterate until
accepted (including an accepted exception) or the fixed limit is reached. The
default is three revision rounds after the initial review, unless Bill sets a
different limit. If still unaccepted, declare the student unable to pass this
chapter within that limit, stop progression and bring Bill the latest code,
student review, unresolved findings, actual results and what was tried. Bill
then helps resolve chapter/grader problems or directs another attempt. Do not
restart the count with another agent. Record rounds in the existing chapter
notes; preserve failed results as the review evolves. Escalate consequential
ambiguity earlier when necessary. Keep this a code review, not another evidence framework.
Commit completed chapters and exact source snapshots with fresh tags, never
moving old tags. Chapters and graders remain unchanged during the student run;
Bill authorizes their joint revision in the later phase below.

## Author phase comes last

Finish the whole new Ensemble implementation and its comparative code reviews
before starting an author agent. A passing grader alone does not establish that
the implementation is better; the comparisons above must support that judgment.

Then the author and coder work together, chapter by chapter, using the original
chapters, all agents' notes, student reviews, reviewer findings and completed code.
The author improves the chapters; the coder applies grader corrections and
meaningful coverage improvements, checking technical claims with the author.
Reconcile prose and grader expectations through feedback in both directions.
The prose reviewer reviews chapters; the code reviewer reviews grader changes.
Run affected checks against student solutions and verify legacy compatibility
where applicable, documenting baseline failures and corrected expectations.
Preserve valid coverage and grade behavior rather than this implementation's
shape. Preserve the original book, graders and actual student-run results as
history; report revised-grader results separately. Review and test any needed
implementation corrections without overwriting frozen snapshots. Do not draft
future chapters during the student run, replace exercises wholesale, or add
features to justify prose. Keep KISS as the standard in this joint phase too.

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
