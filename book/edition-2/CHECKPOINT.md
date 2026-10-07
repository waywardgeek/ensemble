# Working checkpoint

Date: 2026-10-07. Status: ACTIVE. Bill enabled full access without restarting
and authorized continued autonomous work. This file supersedes the original
paused restart record. Read `progress.md` for the current executable next step.

## Objective and current scope

Write the second edition and regenerate Ensemble as a student of that book.
The current map has 21 chapters, 0–20, absorbing old Chapters 5 and 22 where
those lessons first matter. Bill subsequently permitted chapter adjustments
for concrete teaching benefits and explicitly requested an enhanced Chapter 0
execution guide. Author is creating `book/edition-2/chapter-00.md` from the
existing introduction. Preserve the first-edition source and preface.

New implementations live in separate full Git repositories at
`solutions/edition-2/chNN/`. Start Chapter 1 from scratch; derive each later
repository from the preceding new history. Never copy old answer keys or edit
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

## Current state

- Main workflow review requirement committed as `1d3b6c9`.
- Independent acceptance/package tooling and evidence committed as `602ae87`.
- Chapter 1 student final revision: `75542c1c73388fa1ab618dbb8b1e252e3d80816a`.
  Initial comparison/live checkpoint: `459e4ce`. Both are in the student repo.
  Chapter 1 passed implementation, required live demonstrations, scoped audits,
  code-quality comparison/revisions, and final prose review. Bill's separate
  editorial approval is not claimed.
- Chapter 2 contract passed independent review and a cold student read. The
  coder is implementing in `solutions/edition-2/ch02/`, derived from `75542c1`.
  It has no claimed successful grade or live result yet.
- Author is writing Chapter 0's execution guide. Reviewer will check it.
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
without edits. Durable logs are in `checkpoint-evidence/`. Shared legacy
graders have not changed. Do not repeat the long suite absent relevant changes.

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
