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

Both coder and reviewer use **KISS: Keep It Simple** as the yardstick. Choose
the smallest clear implementation meeting the exercise and these rules.
Explain non-obvious choices in comments; compressed statements are not simplicity.
This is teaching code. Bill asks for roughly 20% comments (about one comment
line per four code lines, excluding blanks). Explain why, ownership, lifetimes
and important ordering beside the relevant code so students need not reverse
engineer the design. Treat the percentage as a teaching target, not a padding
quota; useful comments are not bloat.
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

Run gofmt, go vet and go test in affected modules, and the original chapter
grader. Use focused regressions for actual bugs, not a new grading framework.
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
  neither silence nor comments narrating obvious syntax improve maintainability.
  Assess Bill's roughly 20% comment target and whether a student can understand
  the reasons behind the code. Do not accept thin explanations merely because
  the program passes, or filler merely because it reaches the target.
- **Tests:** compare the behaviors and real failure modes protected. Better tests
  are valuable even when longer. Do not reward test volume, implementation-mirror
  assertions or elaborate test infrastructure for its own sake. Preserve the
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
