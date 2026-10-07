# Second-edition checkpoint

Date: 2026-10-07. Baseline: `ec41c6e` (Bill's publishing update).
Status: PAUSED at Bill's request for a Codex restart. Read `CHECKPOINT.md`.
Chapter 1 contract is ready and its fresh student implementation is in progress.
Chapter 2 outline/contract is drafted but not yet ready for a student handoff.
No second-edition chapter has yet been validated or approved by Bill.

## Current work

- Author: Chapter 1 body/contract ready, live demonstration still pending;
  Chapter 2 outline, evidence, and partial contract saved.
- Global reviewer: reports canonical Chapters 1–22 read; forward-lesson map.
- Student coder: fresh repository at `solutions/edition-2/ch01/` contains
  library, CLI, common/llm packages, and tests. Paused for checkpoint; see its
  own `CHECKPOINT.md` for exact validated and unfinished work.
- Coordinator: coding skill now states the mandatory settled methodology;
  root `AGENTS.md` and workflow require loading it for every coding task.

## Latest user ruling

Code likely to need debug logging must have access to the logger. Stateless
helpers are not automatically exempt. Bill suggests the logger probably
belongs on Hub and agent configuration on Agent. A registry may belong on
Agent, or be shared on Hub with per-agent tool visibility. The earlier absolute
"per-agent registries are mandatory" interpretation is superseded.
Hub/Ensemble now clearly means one application-level owner managing many
Agents and communication with the GUI or external gateway. Final naming and
logger ownership remain open. GUI code, including its WebSocket detail, must
live in a separate optional Go module, never `agent/internal`. The second
edition must teach these boundaries at their first introduction, not repair
them in appended chapters. See `architecture.md` for the reconciled record.
Agent holds a back-pointer to Hub/Ensemble, like all children to their owners,
and uses Observer to deliver streaming events and real-time GUI updates to
Hub/Ensemble. Bill explicitly clarified that Observer is not exclusive.
Action requests (such as sub-agent creation and messaging) and shared-service
access can use parent-interface methods. Exact methods remain to be defined.

Chapter 1 must teach the rules in plain English before any code. Core data
and interfaces go in `internal/common`, behavior in responsible spokes, using
free functions when shared types prevent method declarations there. Parent
interfaces preserve access to data and logging. These rules govern the first
implementation, not a later refactor.

Starting in Chapter 2, clean core data structures support both CLI and browser
GUI clients of Hub/Ensemble; the GUI may be stubbed then. Every chapter's
"Taking it for a spin" requires actual coder-run user-facing demonstrations
with a real model, initially CLI, covering all chapter features. Local fake
grading is complementary. Bill authorizes credentials from
`~/.cr/settings.json`; never print, log, copy into repo, or commit secrets.
No credential file has been read and no live call made for this docs task.
Do not write code with architectural ambiguity. Clarify with Bill first.
The current `agent/skills/ensemble/SKILL.md` explicitly requires the immediate
creator chain, `Call.Engine.Agent()` rather than a redundant `Call.Agent`, and
capabilities on the object that owns the fact (usage on Engine).

Working paths: `book/edition-2/` and `solutions/edition-2/chNN/`.
Existing first-edition artifacts and untracked files remain untouched.

## Questions pending

1. Final application-root name (working name: Ensemble). Coordinator selects
   Ensemble as logger owner, following Bill's preference, for the student
   contract; exact method signatures are student design choices.
3. Registry ownership: Agent-owned, or shared application registry with
   per-agent visibility. Both are acceptable in principle to Bill.
4. Define explicit Hub/Ensemble request capabilities (sub-agent creation and
   messaging are candidates), and shared logger/registry access, separately
   from observer notifications.

Bill endorsed separate solution directories and authorized full Git repos and
snapshots in each `solutions/edition-2/chNN/`. No existing Ensemble code may
be edited. New code comes from the second-edition student exercise. Graders
may be enhanced, never weakened; legacy tests must remain passing.
Bill has clarified
that lessons belong at their original points of introduction so structures
and organization improve from the start. The exact revised chapter sequence
is not yet drafted. Retaining P1's sacrificial Chapter 1 remains provisional
where it might conflict with that objective.

## Findings to carry forward

- Chapter 5's global registration example conflicts with its per-agent rule.
- Chapter 22's parent-chain failure must be prevented by persistent property
  checks, not merely moved to an earlier paragraph.
- Old Chapter 1 audit briefs are stale: commits `898f3b3` and `7eb3e24`
  already addressed several purported gaps. Reconcile history before fixing.
- Existing `make grade-dir CH=1 DIR=...` supports isolated student directories;
  `make grade1` does not exist. Go 1.25.4 is available.

## Next action

On restart read `CHECKPOINT.md`, restore author/student/reviewer roles, and
resume Chapter 1 independent acceptance and live validation. Main chapter
procedure is rewritten, proofreader has checked it, and mandatory coder skill
is installed in the repository. Target exactly 21 chapters numbered 0–20,
removing old Chapters 5 and 22 as standalone chapters.

## Methodology validation

- Global reviewer checked Chapter 1, skill, workflow, architecture record,
  and `AGENTS.md`: no material architectural contradiction found.
- Student coder read the rules and reasoned through four no-code scenarios:
  common data/llm behavior, stateless parser diagnostics, GUI events versus
  requests, and real-model evidence after fake grading. Responses preserved
  the intended boundaries and identified remaining contract decisions.
- Author ran prose lint: no hard failures, two soft incomplete-draft warnings.
- Bundled skill validator could not start because its Python lacks PyYAML.
  This is a tooling limitation, not a successful validation result.
- Ruby parsed skill frontmatter successfully; behavioral read-only validation
  passed. Existing agent module full tests PASS; legacy Chapter 1 reference
  and mutation tests PASS. Full root baseline interrupted for restart, so
  remains unvalidated. No grader changes have been made yet.
