---
name: ensemble-coding
description: Mandatory architecture and scope rules for the fresh Ensemble student implementation using first-edition chapters and original graders.
---

# Ensemble student coding

Read this entire skill before coding and after compaction. The original chapter
defines feature scope; Bill's instructions define architecture. Do not import
requirements from the discarded edition.

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

Choose the smallest clear implementation meeting the exercise and these rules.
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

After the initial student attempt, a code reviewer may read the matching original
solution and compare correctness, ownership, simplicity and comments. It returns
findings and rationale without exposing answer code to the student. Challenge
unnecessary code; do not add requirements. The coder improves its answer and
reruns affected checks. Commit each completed chapter and preserve an exact source
snapshot using a fresh tag, never moving an old tag. Chapters and graders remain
unchanged until Bill authorizes later revision.
