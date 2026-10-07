# Second-edition working procedure

Started 2026-10-07 at Bill's request. First-edition baseline: `ec41c6e`.

## Scope and authority

Rewrite exercises from Chapter 1. Bill subsequently requested an enhanced
second-edition Chapter 0 explaining how to execute the book, and authorized
chapter adjustments for concrete teaching benefits. A newly requested final
comparison chapter extends the working map to 0–21. Preserve the preface,
first-edition sources, and existing untracked work. All editions remain
historical artifacts. At publication readiness, rewrite the second-edition
epilogue in Codex's own first person. See `ending-chapter-note.md` for the
comparison protocol and this deferred writing instruction.
The manuscript lives in
`book/edition-2/`, with independent student solutions in
`solutions/edition-2/chNN/`. Bill endorsed this layout and authorized `git init`
for full solution repositories. Each chapter is a self-contained repo with its
own module files and validation evidence; continue subsequent chapters from
the preceding second-edition repository/history and checkpoint completed work.
Do not edit existing `agent/` implementation or first-edition solutions.
The student rebuilds from the new teaching, never from the old answer key.

Grader enhancements are explicitly authorized, including findings raised by
the student, but cannot weaken test coverage. Run and preserve legacy tests;
record any pre-existing failures separately rather than claiming regression
success. The coordinator can implement grader changes while the student
remains blind to grader internals.

Bill authorized an autonomous author/coder/reviewer workflow, and questions
when ambiguity requires his judgment. This authorizes working drafts without
the old per-outline approval ceremony in `book/chapter-writing-procedure.md`.
It does not imply Bill has reviewed or approved those drafts. No pushes.

The second edition deliberately supersedes the final paragraph of course
policy P11 that preserves earlier flawed prose for chronological discovery.
Teach the corrected design when it first matters. Other course rulings remain
in force unless Bill changes them. In particular, P1 makes Chapter 1's
conversation representation the sole sacrificial approach; Chapter 2 onward
must build on a durable architecture.

Bill clarified the purpose further: better data structures and organization
from the start, including one Hub/Ensemble managing many Agents and a separate,
optional GUI Go module. Read `architecture.md` before designing chapter
contracts. Do not preserve an old module boundary merely to match old graders.

## Roles and context

- Coordinator: owns sequencing, decisions, progress, independent grading,
  user questions, and integration. Does not pass answer-key code to students.
- Author: owns chapter prose, outline, and chapter evidence. Reads history,
  first-edition artifacts, voice rules, and the reviewer's forward lessons.
  Does not edit solutions or grader code.
- Student coder: implements from the current second-edition chapter contract,
  earlier second-edition contracts, coding skill, and the preceding
  second-edition solution. Does not read first-edition chapters, old solutions,
  author/reviewer research notes, later implementation, or grader internals.
- Global reviewer: reads every canonical chapter, maps dependencies, extracts
  later lessons, and checks each rewrite for contradictions and omissions.
  Keeps a read ledger; reloads sources when context has been compacted.
- Code reviewer: independently compares the student's completed initial run
  with the corresponding first-edition standard, sends quality findings to
  the coder and teaching findings to the author, and reviews their revisions.
  The global reviewer can fill this role in a separate recorded phase.

The current session has four agent slots. The coordinator handles independent
grader engineering while author, student, and reviewer occupy the other three.
Student isolation is an instruction boundary in this shared workspace, not an
OS access restriction. Report accidental exposure and use a fresh student if
it compromises a cold build.

Start each chapter's cold student in a fresh context with an explicit handoff,
without inheriting the coordinator's conversation. In this orchestration API,
use `fork_turns="none"`; the coordinator's history contains old-book research
and answer-key inspection. A path restriction alone cannot remove that inherited
information. The student may retain its context through that chapter's repairs
and post-run review; use another fresh context for the next chapter. Record
actual reads and any accidental exposure separately from the intended policy.
Historical links in the coding skill are author/reviewer sources, not student
prerequisites. If the new teaching is insufficient, repair it instead of sending
the student to the old edition.

All agents may raise questions. Route them through the coordinator with the
question, source of ambiguity, consequence, and any recommendation. The
coordinator asks Bill in the active conversation and relays the ruling to all
affected agents. Work independent of the answer continues. Never treat silence
as approval for a required decision.

Before every coding task and after compaction, the coder must read the entire
`book/edition-2/skills/ensemble-coding/SKILL.md` (repository-relative path).
This includes fixes, refactors, tests, and grader changes, not only an initial
student implementation. Each handoff names this file, requires the read before
editing, and names the current chapter scope. Root `AGENTS.md` supplies the
same requirement to repository sessions. The coder reports which rules apply
and any unresolved conflict before editing. A skill is a reminder; structural
and behavioral checks enforce its promises. Do not claim permanent context
pinning or rely on automatic skill discovery.

Chapter 1 explains the coding methodology in plain English before the first
code example. Its first implementation must already follow the star, directory
responsibilities, common declarations, free-function placement, and parent
chain. The small conversation exercise has no flat-package exemption.

Starting in Chapter 2, Hub/Ensemble has CLI and browser-GUI client interfaces
over clean core data structures. The GUI side can be a stub then, while its
WebSocket implementation remains in the separate optional GUI module.

## Chapter cycle

1. Author reads current chapter and its relevant git history, audits, and
   later corrections. Record current facts separately from superseded notes.
2. Reviewer identifies future lessons needed now, prerequisites, and what can
   wait. Introduce a mechanism when its first consumer exists.
3. Author writes the outline and a self-contained TL;DR contract. Chapter 1's
   plain-English architecture rules precede all code and travel with every
   student handoff, including TL;DR-only evaluations. Every graded
   requirement appears in the teaching. Leave real design choices to students.
4. Coordinator extracts only student-facing contracts to a clean handoff
   directory. Student builds independently. Grader work derives from the
   contract, without fitting checks to the student's implementation.
5. Run the actual binary against fake services. Audit protected properties
   by deletion, with positive controls and exact expected failing check IDs.
   Preserve earlier behavior and continuously check applicable architecture.
6. If the student needs an unstated fact, repair the chapter and re-run the
   exercise. Do not cure missing teaching with private implementation advice.
7. For every "Taking it for a spin" section, coder actually runs the user-facing
   interface with a real model backend, initially through the CLI. Maintain a
   feature-to-user-action checklist covering every chapter feature and record
   observed outcomes with date, command, and actual provider/model. Internal
   calls and fake servers cannot replace this check. Mark the Chapter 2 GUI
   stub honestly. Use credentials from `~/.cr/settings.json` without printing
   them, logging them, or copying them into repository files. Keep secrets in
   memory or child environments; keep commands and evidence sanitized. A live
   blocker leaves validation incomplete, not waived. Author reconciles prose
   with these real receipts and checks voice and references. Reviewer checks
   the chapter against the whole-book map. Record limitations honestly.
8. After the initial student implementation and runs, the independent code
   reviewer compares its snapshot with the first-edition standard at the same
   feature scope. Record commits, concrete improvements, regressions, and
   tradeoffs in design, code clarity/economy, comments, behavior, and teaching.
   Passing alone is insufficient. The coder revises from the findings, the
   author improves prose/instructions, affected checks and demonstrations run
   again, and the reviewer checks resolution. Keep the initial attempt blind;
   comparison is not permission to copy old implementation or restore flawed
   architecture. Follow §5 of `book/chapter-writing-procedure.md`.
9. Coordinator saves chapter status and next action. Only a chapter with a
   successful student build, grading evidence, audit, required live feature
   demonstrations, first-edition comparison and improvements, and resolved
   code/teaching review is validated.
   Bill's editorial approval is a separate recorded status.

## Durable checkpoint

Maintain `progress.md` with the current chapter, artifact paths, checks run,
unresolved questions, user rulings, and the next executable action. Record
the baseline commit and exact commands in chapter evidence. Historical tests
are evidence about the old edition, never proof the new chapter works.

Use the current `book/voice.md` for prose. Prefer fresh measurements to old
figures. Follow `book/course-policy.md` P10 for model identifiers; fake grading
requires neither credentials nor live inference. Scope formatting and tests to
owned files. Never stage all files or mutate old solutions for an audit.

Live validation covers Anthropic, OpenAI, and Gemini as their adapters enter
the course. At the caching chapter, verify OpenAI's current official method
for subscription-backed access and measure OAuth caching against an API-key
control where comparable. Bill reports an OAuth caching defect and asks that
it be re-tested then, not assumed fixed or working as intended. Keep dated
observations separate from that report; record exact endpoint/model/surface,
credential mode (never credentials), cache counters, and usage limitations.
Do not build the later integration from remembered product documentation.
