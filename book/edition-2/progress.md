# Second-edition progress

Date: 2026-10-07. Status: ACTIVE. Bill enabled full access without a restart
and instructed autonomous work to continue. First-edition baseline: `ec41c6e`.
Main workflow checkpoint: `1d3b6c9`; independent tooling/evidence: `602ae87`.
Chapter 0 and the ending/preservation plan are committed as `24867ca`.

## Current chapter

Chapter 1 is validated against the current chapter contract, with independent
quality review and real-model receipts. Final student checkpoint:
`75542c1c73388fa1ab618dbb8b1e252e3d80816a` in the separate
`solutions/edition-2/ch01/` repository. Initial pre-comparison checkpoint:
`459e4ce`. Bill's editorial approval is not claimed. The validation record
explicitly distinguishes tested properties from remaining general limits of
static analysis and nonexhaustive mutation coverage.

- Formatting, vet, and tests pass in the core and external consumer modules.
- Inherited Chapter 1 grader: 100/100.
- Independent CLI acceptance: 23/23, including a stalled-response timeout.
- Revised deletion audit: passing control and eleven detected mutants with
  exact failing-check sets, plus two parser-classifier mutation tests.
  Package checks pass ten source controls and a real-student identifier rename.
  These are scoped evidence, not proof of all architectural properties.
- Real Anthropic CLI and public-library consumer succeeded on 2026-10-07.
  Final CLI usage: 261 input, 209 output; two independent consumer Agents:
  144/118 and 144/28. Model selected from discovery: `claude-sonnet-5-5`.
  The logger fault was a local transport probe, not an induced provider error.
- Sanitized live evidence and feature checklist are in the student's
  `evidence/` directory. Do not repeat paid calls just to recover context.
- Reviewer found the HTTP client's implicit global transport; coder replaced
  it with an owned transport and a regression test before final live runs.
- Author replaced the live placeholder with observed receipts and strengthened
  motivation and the visible skill-loading instruction. Reviewer accepted
  those revisions and the first-edition comparison's code/teaching corrections.

Chapter 2 is validated at `cc1bec45c3327c87728a4040f762155d8e860a0b`.
Initial student checkpoint `39a92ca27a418712832ac0dcbbfbe4e32b3bca35` is retained,
derived from the reviewed Chapter 1 history. All three real CLI and public-consumer demonstrations succeeded;
sanitized receipts include the initially rejected provider requests and their
corrections. Independent comparison/revisions and prose review are accepted.
See `chapter-02-validation.md`; Bill's editorial approval remains separate.

- Core, external consumer, and optional GUI modules pass format/vet/tests.
- Independent offline CLI acceptance passes 44 checks. A passing control and
  ten deliberate defects produce the exact expected failures. Nine independent
  checker controls exercise valid alternate JSON representations and malformed
  projections. This is partial offline coverage, not a live or ownership claim.
- An early architecture review corrected missing helper/GUI logger access and
  a loaded configuration's caller-buffer alias despite an inherited 100 score.
- The final inherited run initially scored 95 because its roundtrip fixture
  rendered unanswered calls, which the new contract correctly refuses. The
  harness now preserves the full dump and supplies explicitly labeled fixture
  results only for unanswered calls. New score: 100. Legacy Chapter 2 reference
  and mutation tests pass; full root regression passed with exit 0 after this
  change (`internal/grade` 513.863s). Root vet and formatting checks pass.
- The standard comparison requested fewer unnecessary internal history copies,
  safe field/transition-specific diagnostics, and comments explaining subtle
  ordering and copy boundaries. All requested revisions are accepted and passed
  the affected checks without repeating unchanged paid demonstrations.

Chapter 3's contract passed independent review after clarifying search
truncation, binary-file policy, and explicit resolved model identity for tool
continuations. A new student, `/root/coder_ch03`, was launched with
`fork_turns="none"` and only an explicit new-material handoff. It will derive
`solutions/edition-2/ch03/` from the reviewed Chapter 2 history and keep an actual
read ledger. Chapter 4 source research/outline is written; its contract is being
drafted and has not been released to a student.

Chapter 0's execution guide is written and independently reviewed at
`book/edition-2/chapter-00.md`. It covers roles, skills, architecture,
evidence, and checkpoints, preserving the first-edition source.
`ending-chapter-note.md` plans the requested new final comparison chapter;
all editions remain historical artifacts. The epilogue will be rewritten
in Codex's own first person at second-edition publication readiness.

## Required quality comparison

Bill reiterated that the student must succeed from the new edition without
reading the old chapters or answers. The current coder reports no direct reads
of first-edition chapters, old solutions, `agent/`, or grader source. Its context
did include inherited historical summaries, and compaction prevents a complete
raw-trace audit. Treat Chapters 1–2 as guided student builds with no known
answer-key exposure, not certified strictly blind trials. Every subsequent
chapter uses a fresh coder context with `fork_turns="none"` and an explicit
new-only handoff. Historical skill links are for author/reviewer research.
Record actual reads and route missing teaching back to the new chapter.

Bill added a mandatory independent code review after the student's initial
implementation and runs. Compare against the corresponding first-edition
standard, return feedback to the coder, allow revisions, and review them.
Improve design, code clarity/economy, comments, and chapter prose/instructions;
a passing grade is insufficient. See §5 of `../chapter-writing-procedure.md`.
The global reviewer performs this separate code-review phase, independent of
the coder and author. Preserve the original student attempt and avoid copying
old implementation into the new solution.

## Roles and next actions

- Author: reconcile Chapter 2's final status, resolve Chapter 3 review, and
  draft Chapter 4; preserve the requested voice improvements.
- Coder: Chapter 2 role complete. Fresh `/root/coder_ch03` implements Chapter 3
  from the reviewed new contract, with no coordinator-history inheritance.
- Global reviewer: review Chapter 4's contract and new student work; preserve
  independence from the coder/author and maintain whole-book guidance.
- Coordinator: independent acceptance/property audit, enforce architecture,
  integrate evidence, and advance only when the chapter meets all gates.

## Scope and enduring decisions

The current plan is 22 chapters, 0–21, including the newly requested final
edition-comparison chapter. Bill now permits chapter adjustments
for teaching quality and requests an enhanced Chapter 0 execution guide in
the second-edition directory. Preserve the first-edition source and preface.
Absorb old Chapters
5 and 22 where their lessons first matter; old 6–21 become new 5–20.
Teach architecture before Chapter 1 code. Common declares core data and
interfaces; behavior stays in owning spokes, using free functions as needed.
Children follow interface parent chains to data and logger. Ensemble is the
working application-root name and logger owner; Agent owns config/history;
Engine owns transport/usage. Observer carries live events, while explicit
parent methods can supply services and action requests. Registry ownership
remains a decision for the tools chapter; visibility must be per-agent.

Chapter 2 introduces CLI and browser client seams. GUI/WebSocket belongs in a
separate optional Go module and may be an honest stub then. Each chapter must
actually demonstrate all supported features through its user-facing interface
with real models. Credentials stay in memory/environment, never argv or repo.
At the caching chapter, verify current official OpenAI subscription guidance
and remeasure the reported OAuth caching failure against suitable controls.

New solution repositories only: `solutions/edition-2/chNN/`, derived from the
preceding new history. Preserve `agent/`, first-edition solutions, and Bill's
unrelated files. No pushes. Ask Bill when genuine architectural ambiguity
blocks affected work; routine naming does not require approval.

## Legacy evidence

Existing agent full tests and old Chapter 1 reference/mutation tests passed.
The initial resumed full-root baseline log had all passes but lost its process
exit receipt across daemon recovery. A subsequent full-root run passed with
exit 0, including `internal/grade` (508.958s). Logs are durable in
`checkpoint-evidence/`. Root vet and the latest package-checker tests pass.
The Chapter 2 roundtrip harness fixture was subsequently corrected as described
above, with a passing pre-change legacy reference baseline and passing targeted
legacy reference/mutation suite afterward. Full root regression subsequently
passed with exit 0 (`internal/grade` 513.863s); its durable log is retained.
Independent second-edition scripts are under `scripts/edition2/`.
