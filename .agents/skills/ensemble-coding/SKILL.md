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
  constructor receives and stores an interface back-pointer to its actual parent.
  Declare parent interfaces in common; expose the parent's own parent through
  an interface method. The application root has no parent.
- Reach configuration, services and logging through that chain. Keep facts with
  their owners, not copied into siblings or supplied through callback bags.
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
Do not introduce generic frameworks, extra protocols, invented limits, speculative
recovery or exhaustive audit machinery merely because they might be useful.
There is no author agent and no grader agent working ahead of the student.

Record teaching gaps and suspected grader defects for future maintainers; do not
rewrite either during this attempt. Ask Bill about consequential ambiguities.
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
programmatically without printing them. Keep secrets in memory or child
environments, never command arguments, source, logs or committed files. Report
live access failures accurately; fake-server success does not prove live use.

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
- **Tests:** compare the behaviors and real failure modes protected. Better tests
  are valuable even when longer. Do not reward test volume, implementation-mirror
  assertions or elaborate test infrastructure for its own sake. Preserve the
  original grader's role; grader defects go into a future-maintainer review.
- **Scope and simplicity:** trace added behavior to the original exercise or an
  explicit Bill instruction. Examine data structures and abstractions for the
  current need. Reject feature creep, unjustified bloat and complexity added for
  elegance. Do not justify extra code by promises the workflow itself invented.

Passing tests is necessary, not sufficient. Explain material growth and ask
whether a simpler design satisfies the same actual requirements. Do not impose
an arbitrary line quota or remove required behavior to win a size comparison.
Return concrete findings and rationale without exposing old answer code. State
whether the chapter is accepted or needs simplification/correction; do not
accept unresolved scope or bloat findings. Ask Bill if a consequential tradeoff
cannot be resolved from his instructions.

The coder addresses findings and reruns affected checks; the reviewer checks
those revisions. Keep this a code review, not another evidence framework.
Commit completed chapters and exact source snapshots with fresh tags, never
moving old tags. Chapters and graders remain unchanged until Bill authorizes
their later revision.

## Author phase comes last

Finish the whole new Ensemble implementation and its comparative code reviews
before starting an author agent. A passing grader alone does not establish that
the implementation is better; the comparisons above must support that judgment.

Then the author revises the book chapter by chapter using the original chapters,
all agents' chapter notes, the coder's reviews, reviewer findings and the completed
code. Build the second edition from demonstrated lessons and actual student experience.
Preserve the original book. Do not draft future chapters during implementation,
replace exercises wholesale, or add features to justify new prose. Keep KISS as
the standard in the author phase too.
