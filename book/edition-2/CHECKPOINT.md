# Working checkpoint

Date: 2026-10-08. Autonomous work is active. Chapter 5 code, live evidence,
student feedback and manuscript are accepted at `c1cc0b4`. Its exact 1,254-file
export uses source `185ba76`, reviewed runtime `959c663` and revised evidence
`469730f`, with checkpoint `edition-2-ch05-r1`. All six exported modules pass
vet/tests. See [chapter-05-validation.md](chapter-05-validation.md).

Chapter 5's tag points to `7a6ef036322e1cf362894bd30b073fc8399a3c30`.
Chapter 6 is accepted at `edition-2-ch06-r1`. Reviewed runtime `75bd14d`,
revised live evidence `8c73f9b`, independent acceptance `37a37f7` and exact
1,550-file export from `c3fa758` are bound in its manifest. All required gates
pass; original attempts and limitations remain recorded. See
[chapter-06-validation.md](chapter-06-validation.md). Fresh new-only
student `/root/coder_ch07` completed Chapter 7, accepted at `edition-2-ch07-r2`.
Reviewed runtime `9ba7855` passes 33 full-gate and 12 comparative groups.
Initial `8cc87f2` and revised `341f15d` evidence retain 44 and six exactly
reconstructed requests, respectively. Independent live review `ffbad61` and
final proofread `121b65d` close remaining gates. Corrected export source `9ec94ef`
has 2,229 files across eight modules; it isolates a standalone evidence helper
missed by the runtime-only gate. Actual full-tree checks pass and r1 is preserved.
Fresh `/root/coder_ch08` used `fork_turns="none"`, new Chapters 1–8 and that
accepted source. Its final student work is frozen through `7f517d8`, with
actual all-provider receipts; the initial attempt remains separately preserved.
Independent code review `270c3b6` accepts runtime `a06d4f3` with 52 accounted-for
groups. Final live audit, author reconciliation and proofread are accepted at
`d528419`, with student confirmation `446d7f2`. Chapter 8 is checkpointed as
`edition-2-ch08-r1`, an exact 3,308-file export from that confirmation source.
Chapter 9's initial checker is published at `e6c3406`. Fresh local CLI student
`01a11c0d-47b8-7241-8834-5ddf57ac5009` is reading new-only inputs at `8736c95`
and preparing its ownership plan. Managed spawn hit its thread limit; this
separate session inherits no conversation and disables memories. See
[chapter-09-validation.md](chapter-09-validation.md) for its handoff and limits.
Source identities and retained limitations are tracked in
[chapter-07-validation.md](chapter-07-validation.md).
Earlier log-destination corrections are independently accepted and checkpointed
at `edition-2-ch02-r2` (`0a375d3`) and `edition-2-ch03-r2` (`799940c`). Their
revised frozen exports and manifests are integrated here; canonical main stays
on the Chapter 8 working implementation. See `logpath-backport-review.md` and the chapter validation records.
Chapter 4 remains checkpointed at
`edition-2-ch04-r1` (`55e6411`); its detailed record is historical below.

Chapter 3 human integration is
accepted, with dedicated checkpoint `edition-2-ch03-r1`. Reviewed source and
independent receipts are at `8494fdb0e5d6c096445bfac039458b83dd225332`; exact
491-file export and source/evidence manifest are under `solutions/edition-2/`.
See `progress.md`, `chapter-03-validation.md` and `chapter-03-human-review.md`.

`/root/coder_ch04` completed Chapter 4 from a fresh new-only context, including
actual model-driven Delve through Ensemble on all three providers. The subsequent
Chapter 5 checkpoint now supplies Chapter 6's accepted predecessor.
`/root/coder_ch03_chat` completed independent Chapter 4 and story-restoration
prose review; status reconciliation is complete. Its previous client authorship must be
disclosed. The old
global-review thread could not resume because of tool thread capacity. Read its
durable map/reviews and current progress, not assumptions about live threads.

Consolidation is complete and verified at `8831ce2`. Original nested histories,
ignored/untracked files, and the old staged merge remain safely archived;
tracked bundles and `history/migration.json` preserve their source identities.
No root regression tests remain running. Do not repeat paid demos to recover
context. No pushes; preserve Bill's unrelated untracked files.

## Objective and current scope

Write the second edition and regenerate Ensemble as a student of that book.
The current map has 22 chapters, 0–21, absorbing old Chapters 5 and 22 where
those lessons first matter and adding Bill's requested final edition-comparison
chapter. Bill permits chapter adjustments for concrete teaching benefits.
The enhanced Chapter 0 execution guide is written and independently reviewed.
Preserve all historical editions, their source/code/evidence, and the preface.
`ending-chapter-note.md` records the comparison plan and Bill's instruction to
rewrite the second-edition epilogue in Codex's own first person when the edition
is ready for publication. Do not write a fictional completed outcome now.

The authoritative implementation lives in `solutions/edition-2/main/` within
the outer Git repository. Each `chNN/` is a frozen source export with a source
manifest. Extend the preceding accepted new source and checkpoint each chapter
with an outer commit and immutable annotated revision tag. Never copy old answer keys or edit
existing `agent/` implementation and first-edition solutions. Grader changes
cannot weaken coverage and must retain legacy passes. Do not push or stage
Bill's unrelated files.

## Mandatory context and roles

Read root `AGENTS.md`, `book/chapter-writing-procedure.md`, and this directory's
`architecture.md`, `workflow.md`, and relevant chapter contract. Every coding
agent must freshly read the entire `skills/ensemble-coding/SKILL.md` before a
coding task and after compaction. Every coding handoff explicitly names
`book/edition-2/skills/ensemble-coding/SKILL.md` and requires that read.
Author and proofreader also read the full `book/voice.md`.

Root coordinates and independently grades. Worker roles are author, student
coder, and global reviewer; the reviewer performs independent code-review and
proofreading phases. Three worker threads fit in this session; scheduling
roles in phases is not a requirement of every reader's platform. Inspect
actual agent status after interruptions rather than assuming thread state.

The initial student attempt uses the new teaching and preceding new solution,
not old implementation or grader internals. AFTER the first implementation
and runs, an independent code reviewer compares it with the first-edition
standard, sends findings to coder and author, and reviews their revisions.
Passing alone is insufficient: improve code, design, comments, and teaching.
Preserve the initial attempt and record concrete gains and tradeoffs.

For every next chapter, start a fresh coder with `fork_turns="none"`; never
inherit coordinator research/answer inspection into the student. Exclude old
chapters as well as old solutions and author/reviewer notes. The current coder
reports no direct old-material reads, but inherited summaries and compacted
history prevent certification of Chapters 1–2 as strictly blind trials. Preserve
that limitation; their implementation validation is a separate claim.

## Current state

- **Latest accepted source:** Chapter 8 r1, source `446d7f2`; see its gate record.
- Earlier Chapter 4 checkpoint, runtime d25/final evidence e1c6488.
  Five-module checks, grade100, independent53/53 and public3/3, eleven mutations,
  race/fault/boundary controls, actual all-provider PTYs and public consumers,
  evidence/source-binding controls and final prose all accepted.
- Chapter 3 human integration, as recorded above:
  All four modules format/vet/test; inherited100; human45/45; passing control
  plus seven exact-failure mutations; four verifier controls independently rerun.
  Actual all-three-provider human PTYs and prose/comparison review accepted.
  Production/live initiala347ce3 is unchanged by evidence-only repair339a2a6.
- Chapter 3 earlier tool-scope snapshot `7cbbd8e2` remains historically bound to
  independent39/39, eight checker controls and five mutants, all-three machine
  CLI/public-consumer runs and comparison. Root legacy retry passed exit0
  (`internal/grade`508.374s); earlier disk-exhaustion failure remains recorded.

- Main workflow review requirement committed as `1d3b6c9`.
- Independent acceptance/package tooling and evidence committed as `602ae87`.
- Chapter 1 student final revision: `75542c1c73388fa1ab618dbb8b1e252e3d80816a`.
  Initial comparison/live checkpoint: `459e4ce`. Both are in the student repo.
  Chapter 1 passed implementation, required live demonstrations, scoped audits,
  code-quality comparison/revisions, and final prose review. Bill's separate
  editorial approval is not claimed.
- Chapter 2 prior-contract checkpoint: `cc1bec45c3327c87728a4040f762155d8e860a0b` in
  `solutions/edition-2/ch02/`, derived from `75542c1`. Initial student checkpoint
  `39a92ca27a418712832ac0dcbbfbe4e32b3bca35` is retained. All three live CLI and
  public-consumer paths succeeded; receipts are in `evidence/ch02/`. Independent
  comparison revisions and live prose review are accepted.
- Chapter 2 offline acceptance passes 44 checks, with a passing control and ten
  detected defects. Scoped package check passes. Initial inherited95 records
  an invalid roundtrip fixture; the corrected fixture produces100 and retains
  passing legacy reference/mutation tests. Full root regression passed exit 0,
  with `internal/grade` 513.863s. See `chapter-02-validation.md`.
- Chapter 0's execution guide is written and independently reviewed.
  Chapter 3's fresh coder used `fork_turns="none"` and new-only sources; its
  reviewed snapshot is recorded above. Chapter 4's student completed the
  accepted checkpoint above; Chapter 5's current gates have their own record.
  Chapters 1 and 2 now motivate the rules and visibly teach skill loading,
  following Bill's forwarded CodeRhapsody advice.

Chapter 1 checks: both modules format/vet/test clean; inherited score100;
23 independent CLI checks; eleven CLI mutants plus passing control; two
parser-classifier mutations; ten package-boundary controls and actual student
identifier-rename control. Structural analysis is scoped, not a proof of every
possible ownership violation. See `chapter-01-validation.md` and
`chapter-01-code-review.md` for evidence and limitations.

Full legacy root tests passed with exit0, including grade508.958s; root vet
and latest targeted package-checker tests pass. Existing agent suite passed
without edits. Durable logs are in `checkpoint-evidence/`. The Chapter 2
roundtrip harness now explicitly completes unanswered calls with fixture results
before rendering, preserving every original dumped byte. Targeted legacy tests
pass; the subsequent full-root run passed with exit 0. Do not duplicate it.

## Live evidence and credentials

The existing Chapter 1 paid CLI and public-consumer receipts are saved under
`solutions/edition-2/ch01/evidence/`; do not repeat paid calls to recover state.
They used discovered model `claude-sonnet-5-5`, with final CLI usage261/209.
Independent consumer Agents retained their own histories and totals. Logger
failure evidence is a deliberate local probe. Receipts are bound to459e4ce;
75542c1 adds reviewed diagnostics/comments/named fields, validated locally.

Every chapter must exercise every supported feature through actual user paths
with real models, initially CLI; three vendors as introduced. Fake checks and
GUI-stub integration must be labeled honestly. Bill authorizes keys in
`~/.cr/settings.json`: programmatic memory/environment only, never dump keys,
put them in argv/prompts/logs, or commit them. Record safe source/run provenance.

At the caching chapter, verify then-current official subscription-access
methods and measure the reported OpenAI OAuth caching issue against suitable
API-key controls. Do not assume a historical claim or closed issue proves it.

Bill also asked for an opinion on `book/the-improvement-loop.md`. Root read it
and discussed using attempts, review, revisions, and distinguishing tests as
potential future training data. No model weight training or training-speed
improvement has been performed or authorized as implementation work here.
