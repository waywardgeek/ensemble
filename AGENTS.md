# Ensemble coding instructions

Before writing or editing code for Ensemble, its chapter solutions, tests, or
graders, read the entire repository skill:

`book/edition-2/skills/ensemble-coding/SKILL.md`

Loading this skill is mandatory for every coding task and after context
compaction. Keep it in context while coding. Every delegated coding handoff
must include that path and require the recipient to read it before editing;
do not rely on automatic skill discovery or a remembered summary.

The methodology applies from Chapter 1: star imports, core data and interfaces
in `internal/common`, behavior in the responsible package using free functions
when necessary, and interface back-pointers providing access to owned data and
logging. Chapter 1 must teach these rules before its first code example.

For the second edition, read `book/edition-2/architecture.md` and the relevant
chapter contract as well. Resolve architectural ambiguity before affected
implementation. Historical solutions are evidence, not authority to reproduce
the flaws the second edition is correcting. Do not rewrite unrelated legacy
code merely because it differs from the new rules.

Bill explicitly requires a fresh student rewrite in `solutions/edition-2/`.
Do not edit existing `agent/` implementation or first-edition solutions.
Each new `chNN/` is a self-contained Git repository; initialize the first,
then derive subsequent chapter repositories from the preceding new history.
Checkpoint validated solutions. Never copy the first-edition implementation.
Grader enhancements are allowed, but may not weaken coverage and must retain
passing legacy tests. Record baseline failures before changes if encountered.

Every chapter's "Taking it for a spin" requires actual user-facing runs with
a real model, initially through the CLI, exercising all chapter features.
Fake-server grading alone does not fulfill that requirement. Follow the
skill's live-evidence and credential handling instructions.
