# Chapter 10 validation

Persistence implementation released, October 8, 2026. A fresh CLI student started
from accepted `edition-2-ch09-r1` at `54d7b1d`, source `ac55f64`, tree `94315d7`.
Its plan-only phase ended at `7cb8429`. Independent ownership review `72bf621`
and published answers `af5a762`, proofread at `6766995`, release local implementation
after the student reads and acknowledges those answers. Credentials and provider
calls remain outside this phase.
Bill's editorial approval is separate from technical acceptance.

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Contract | Author, coordinator | Q1–Q3 and Unicode clarification accepted; prepared-event clarification `c5ad6c0` has independent closure `c773b11` and a new-only student pin | Obtain student acknowledgment and verify the affected implementation |
| Independent checks | `/root/grader_ch05` | Initial CLI run 12/93; public run found two failures. Canonical-path fixture `e118f78` and numeric mutation serialization `d2d8481` corrected; client runner `2bec837` prepares nine additional unrun groups | Retest repaired source, complete fault/limits/DOM/bounds coverage and distinguishing controls |
| Fresh student and owner plan | Fresh CLI student `01a11cc3-9e40-7d62-a7b5-9b2ec4c928c0`, coordinator, independent reviewer | Plan `7cb8429` accepted at `72bf621`; answers `af5a762` proofread at `6766995`, fully acknowledged by student at `44056d1` | Preserve new-only read ledger and route new teaching gaps before affected code |
| Implementation and local gates | Same student, grader | Phase 2 running; grammar/API through `b926929` adds accepted request witnesses and cancellation of checkpoint waiting. Main/GUI tests and nine nested module checks reported; source remains mutable | Repair accepted-byte boundary, run complete local/independent gates and freeze initial attempt |
| Actual use | Future student, reviewer | Not started | Review bounded plan, then real CLI/browser/public all-provider demonstrations |
| Historical comparison and revisions | Independent reviewer | Not started | Preserve initial source, runs and teaching review first |
| Manuscript and feedback | Author, student, proofreader | Draft labels actual spin pending | Reconcile actual evidence and resolve student feedback |
| Export and checkpoint | Coordinator | Not started | Complete all gates before immutable export/tag |

The grader engineer has earlier grading/review exposure and no Chapter 10 runtime
authorship. See [chapter-10-grader-review.md](chapter-10-grader-review.md) for
the runnable command, preparation receipts and explicitly incomplete scope.
Neither a synthetic control nor a preceding chapter's expected refusal is a
Chapter 10 implementation positive. Private semantic codec spelling remains a
student design choice under the published complete-state requirements.

The separately supervised CLI worker uses a new conversation, not Chapter 9's
resume ID. Its prompt, twelve pinned new teaching/architecture/skill files and
hash manifest are retained under
`/Users/bill/projects/ensemble-edition-2-revisions/ch10-student-inputs/`.
It may inspect the accepted preceding second-edition source, but not old/future
chapters, historical answers, author research, grader/reviewer implementations
or other workers' conversations. The shared filesystem remains an instruction
boundary. The student records actual reads and questions in its teaching review.
The coordinator inbox is the only authorized ongoing message file. Do not start
a competing worker or release paid tests before the later reviewed feature plan.

Phase 2 resumes the same student conversation. External `phase-2.txt`, events,
stderr and result paths preserve its handoff and process evidence. Only the new
chapter and direct author response are pinned in `clarification-af5a762`; author
research and reviewer/checker internals remain excluded. The coordinator read the
complete revised chapter and plan and accepts the published Q1–Q3 resolutions.
Private validation owners must remain inert until validation succeeds, and GUI
delivery waits must release with their connection without blocking Actor or the
checkpoint worker. These are ownership safeguards, not extra public API names.

The early review found two implementation/design omissions: saved call state
needed dispatch/result sequence positions to prove limit-fact ordering, and a
checkpoint worker initially held a concrete Store parent. The student classifies
both as its own deviations from the published requirements at `81aa8cf`, adds
the required coordinates to the grammar and accepts the common parent-interface
correction. The coordinator separately found the escaped-Unicode teaching gap;
`cf73a64` and the student's new-only acknowledgment preserve that attribution.
These early corrections are not a completed runtime or historical comparison.

Additional black-box command, supplied to the student without fixture source:

```sh
python3 scripts/edition2/accept_ch10_public.py solutions/edition-2/main --receipt RECEIPT.json
```

Its nine groups do not replace the initial checker or complete the matrix. The
first student run has seven passing top-level groups and two failures: canonical
path comparison and equivalent schema identity. The former was a macOS alias
fixture defect, corrected at `e118f78`; the latter was a student bytewise-versus-
canonical equality defect. Targeted revised receipts pass both groups. These
results are local checks of mutable source, not frozen implementation acceptance.
The initial predecessor evidence remains three standalone passes and ninety
expected missing-session features, not student failures.

The student's initial CLI binary scored 12/93, with many negative rows blocked
by failed untouched-session inspection. Independent diagnosis `63dea39` proves
that the initial checkpoint kept spaced raw usage while its own log compacted
the fragment. A checkpoint-free copy rebuilds successfully. The coder separately
reproduced the same mismatch and raised Q4 before changing the acceptance rule.
Author `c5ad6c0`, independently reviewed at `c773b11`, now teaches preparing the
bounded event once, deriving accepted raw fragments from those exact bytes, then
appending before apply/observe. Imported accepted bytes remain unchanged.
Only the new chapter/direct response are pinned in `clarification-c5ad6c0`;
the affected fix is released after acknowledgment. No repaired result is yet
claimed. The original binary, receipts and difficulty remain preserved.

Separate checker repair `d2d8481` preserves numeric tokens in semantic mutation
fixtures and exposes bounded failure diagnostics. It does not explain the
untouched-session restart failure. Client/lifetime runner `2bec837` was supplied
as a black-box command for coherent repaired binaries:

```sh
python3 scripts/edition2/accept_ch10_clients.py CLI_BINARY GUI_BINARY --source-directory solutions/edition-2/main --receipt RECEIPT.json
```

Its nine runtime groups remain unrun; oracle-only controls are preparation
evidence. Independent fault, bounds, limits, DOM and retained-behavior coverage
is still being completed. None of these partial checks releases paid calls or
substitutes for the full chapter matrix and later comparative quality review.
