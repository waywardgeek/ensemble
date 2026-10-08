# Second-edition progress

Date: 2026-10-07. Autonomous work is active. Source consolidation is complete
and independently verified at `8831ce2`; original histories and unfinished
migration state remain preserved in the archive and tracked bundles.
`solutions/edition-2/main/` is the authoritative source. Frozen exports belong
in `chNN/`; commits and immutable annotated tags bind each validated chapter.

## Current chapter

Chapter 5 code, actual live evidence, teaching and manuscript review are
accepted at `c1cc0b4`. The exact 1,254-file export comes from `185ba76`, with
reviewed runtime `959c663` and revised evidence `469730f`. Checkpoint:
`edition-2-ch05-r1`. See [chapter-05-validation.md](chapter-05-validation.md)
for the authoritative gate record, source identities and retained limitations.
All six exported modules pass vet/tests. Chapter 6's reviewed streaming
contract is next; its student must start in a fresh new-only context.

Chapter 4 remains accepted at `edition-2-ch04-r1` (`55e6411`); its exact
780-file export and source/evidence bindings are recorded in
[chapter-04-validation.md](chapter-04-validation.md). Bill's editorial approval
remains separate from validated checkpoints. No push or publication is claimed.

Chapter 3's human-client integration is accepted. Initial production/live
commit: `a347ce31511c4b124e486bb41ef98c07bd17cec5`. Evidence-only verifier repair:
`339a2a61e7107b921bdc7acbd704b2eb7da17931`; accepted ledger: `17b60d2`.
Reviewed source plus independent receipts: `8494fdb0e5d6c096445bfac039458b83dd225332`.
The exact 491-file source export is `solutions/edition-2/ch03/`; its manifest
is `solutions/edition-2/manifests/ch03-r1.json`. Dedicated chapter checkpoint:
`edition-2-ch03-r1` at `1a61e1f2487cdc94ee65bd0e1593065cc1e95f45`. Bill's editorial approval is separate.

All four modules passed format/vet/tests, inherited grade100, independent
human45/45, a passing control plus seven exact-failure mutants, and four
independently rerun verifier controls. Actual all-provider macOS PTY sessions
covered all six tools, refusals/recovery, redaction and later context, Unicode
bounds, silent nonzero exit, commands, slash escaping, quit and EOF. The
OpenAI model omitted one requested overwrite and needed an observed follow-up;
its mistaken description of a UTF-8 result is not treated as tool evidence.
No production change or paid repeat followed review. GUI remains a stub;
Linux runtime is not claimed. Earlier Chapter 3 machine/public-consumer and
architectural evidence retains its original bindings and scope.

Chapter 2's accepted human source is frozen at `ch02/`, matching historical
`ad0d80e33a3a2857e8e0887117d9b099f1a4786d` (42 independent cases, five mutants,
all-three-provider PTYs). Chapter 4 used a fresh `fork_turns="none"` student
with only new teaching and the accepted predecessor. Historical comparison
was supplied afterward as reviewer rationale, preserving the initial attempt.

Chapter 1 is validated against the current chapter contract, with independent
quality review and real-model receipts. Final student checkpoint:
`75542c1c73388fa1ab618dbb8b1e252e3d80816a` preserved in the historical bundle and exported to
`solutions/edition-2/ch01/`. Initial pre-comparison checkpoint:
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

Chapter 2's earlier contract passed at `cc1bec45c3327c87728a4040f762155d8e860a0b`;
its human-chat correction subsequently passed at `ad0d80e` (see above).
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

Chapter 3's earlier scope passed at reviewed snapshot
`7cbbd8e2e0101e40dee63fe18d05f48b791c55ff`; the later human gate is accepted above.
Its fresh
student, `/root/coder_ch03`, was launched with `fork_turns="none"` and only an
explicit new-material handoff. The initial implementation `590c4f4` and
initial run checkpoint `540fb4a` share identical production code and derive
from the reviewed Chapter 2 history. The source-read ledger records new-only
inputs. All three real CLI and public-consumer demonstrations succeeded.

- The frozen initial binary scored 65 with incompatible inherited fixtures,
  then 100 after grader-only corrections for Gemini JSON Schema and explicit
  fake resolved-model identity. Targeted legacy reference/mutation tests pass.
- Independent initial acceptance passed 33 of 36 scenario groups; the failures
  exposed the empty-file `end_line:0` teaching gap. The author clarified it
  before the guided fix. Initial evidence remains distinct from revisions.
- Review corrected a concrete capture parent to a common interface, replaced
  unnecessary whole-configuration copies with an Agent-owned workspace query,
  and found a UTF-8 byte-cap defect shared with the historical answer. The
  author taught complete-code-point prefixes before the student corrected it.
- Revised production passed 39 independent checks, eight checker controls,
  and a control plus five targeted mutations with exact expected failures.
  Independent code and final prose/receipt review are accepted.
  Focused live checks of the empty-file correction and silent command failure
  succeeded on all three providers. Unchanged paid demonstrations are retained.
- Full legacy regression hit linker errors from a full disk. The failed log is
  preserved; clearing only the disposable Go build cache recovered space.
  The full retry passed with exit 0 (`internal/grade` 508.374s), and root vet
  and changed-file formatting passed. The later human-mode results remain separately bound above.

Chapter 4's asynchronous events coexist with response identity assignment
under Agent append serialization; its accepted checkpoint is recorded above.

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

- Author: Chapter 5 final prose and feedback are accepted. Chapter 6's full
  streaming contract and story proofreading are reviewed and ready for handoff.
- Coder: `/root/coder_ch05` completed Chapter 5. The next student needs a fresh
  new-only context from accepted `edition-2-ch05-r1`.
- Chapter 4 reviewer: `/root/coder_ch03_chat`, independent of this chapter's
  new coder, completed lifecycle/owner/fault/debugger and comparative review.
  This agent authored the preceding human-client integration; disclose that
  context rather than claiming total source blindness. The original global-review
  thread could not resume (tool thread limit); its durable map/reviews/handoff
  remain available, and the full skill/voice/procedure reload was required.
- Coordinator: preserve initial attempts, route teaching gaps before coding,
  require actual all-provider human runs, compare/revise/review, then export and
  checkpoint each validated chapter. No pushes or publication yet.

## Scope and enduring decisions

Earlier correction to propagate: the manual review's immutable LogPath fix
first lands in Chapter 4 runtime d25. The new Chapter 2 teaching now explains
that invariant; frozen ch02/ch03 need a scoped backport and validation before
new revision tags claim the clarification is implemented there. Preserve the
old tags and receipts. The current Chapter 5 baseline already contains the fix.

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
is currently Agent-owned under Chapter 3's working choice; visibility remains
per-agent even if a later design shares storage on Ensemble.

Chapter 2 introduces CLI and browser client seams. GUI/WebSocket belongs in a
separate optional Go module and may be an honest stub then. Each chapter must
actually demonstrate all supported features through its user-facing interface
with real models. Credentials stay in memory/environment, never argv or repo.
At the caching chapter, verify current official OpenAI subscription guidance
and remeasure the reported OAuth caching failure against suitable controls.

New source only in `solutions/edition-2/main/`, derived from the preceding
accepted new source. Frozen `chNN/` exports and immutable revision tags preserve
chapter checkpoints in the outer repository. Preserve `agent/`, first-edition solutions, and Bill's
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
