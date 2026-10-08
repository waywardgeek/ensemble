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
| Contract | Author, coordinator | Prepared-event clarification `c5ad6c0`/`c773b11` and exact watermark clarification `3a9e5b7`/`2a24f7c` published, pinned and acknowledged | Preserve exact representation and complete-state identity rules |
| Independent checks | `/root/grader_ch05`, coordinator, `/root/coder_ch08` | Prepared executable passes CLI 93, clients 9, DOM 6 and physical reads 38; corrected public suite passes 16 groups on later mutable source. Initial failures remain retained | Bind repaired source; finish independent fault, allocation, storage and retained-behavior gates |
| Fresh student and owner plan | Fresh CLI student `01a11cc3-9e40-7d62-a7b5-9b2ec4c928c0`, coordinator, independent reviewer | Plan `7cb8429` accepted at `72bf621`; answers `af5a762` proofread at `6766995`, fully acknowledged by student at `44056d1` | Preserve new-only read ledger and route new teaching gaps before affected code |
| Implementation and local gates | Same student, grader | Initial `41a5e72`, revised `122b04a`, append-fault repair `8882a18`, handback `e307d79`. All four independent fault groups and two lifecycle groups pass on the repair | Finish bounds, deletion controls, structural and retained-behavior gates |
| Actual use | Same student, reviewer | Unrun proposal reviewed at `c9ba32c`; five concrete plan/support corrections requested, no calls released | Finish deterministic clearance, executable schedule and immutable preflight before real demonstrations |
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

Its nine runtime groups subsequently pass in clients-initial.json. Six browser
groups pass in browser-initial.json. Physical read controls pass 38 rows against
the retained prepared CLI; see chapter-10-record-bound-review.md. These include
actual 64 MiB/one-over LF/EOF records, not allocation or write-bound evidence.

The expanded public suite first stopped at a fixture compile error, retained in
public-prepared.json. After that fixture-only correction it passed fifteen groups
but exposed a real lock cycle: Agent append held Agent.mu while waiting for Jobs
limits state, while a report worker held Jobs.mu and waited for Agent.Workspace.
The 90-second failure is retained in public-prepared-corrected.json. Independent
grader and coordinator confirmed the cycle; the fixture's request schedule was
valid. The initial attempt and binary/source manifest are frozen at `41a5e72`.

The student's grouped repair separates accepted limit synchronization, corrects
watch ownership lookup, enforces published Q5 maxima and introduces derived
validation indexes/bounded encoding. Targeted Skills replay passes all three
local API shapes, and public-lock-index-maxima.json passes all sixteen public
groups. These are mutable-source receipts, not final immutable acceptance.
Independent fixture commits `833c735` and `fe3ef1b` preserve the first failures.

Q5 arose from a real prose contradiction about allocator watermarks. Published
`3a9e5b7`, proofread at `2a24f7c`, requires exact represented activation/session-job
maxima even for snapshot-only origins; only request admissions may leave an
unrecorded higher cursor. The student confirms the complete new-only read.
Current root job allocation still preserves an already higher live floor.

The inherited old Chapter 11 diagnostic returned 0/100 with incompatible old
CLI/startup assumptions. That result is retained, not reclassified as a pass or
used to overwrite the new contract. Independent fault barriers, allocation,
complete storage/semantic and retained-behavior coverage are still being completed.
No partial local result releases paid calls or substitutes for the full chapter
matrix and later comparative quality review.

The grouped correction is now frozen at `122b04a` with status-only `1e041f5`.
Independent overlay `f572cd3` passes checked checkpoint-I/O, canceled waiting and
close/join/lock controls, but one-byte public Append failure leaves Close returning
nil and actually creating/replacing a checkpoint. The checkpoint bytes happen to
match; byte equality alone would miss the forbidden operation. Reopening correctly
refuses the partial log. Both diagnostics remain preserved. The same student has
resumed for a focused repair; no ordinary validation refusal may become terminal.

The initial retained-suite run against isolated `122b04a` exhausted disk before
emitting its final result. Its error is preserved, with no recovered pass claim.
Only regenerable build cache was cleared after compilers stopped. Remaining runs
must serialize builds and retain per-command results. Other source/artifacts and
module downloads are preserved. The live matrix remains pending its independent
`c9ba32c` corrections, local clearance and immutable support preflight.

Retained-gate orchestration is corrected at `c8353a2`: the existing gate accepts
optional progress callbacks without changing its checks, and the Chapter 10
wrapper saves each started/completed command atomically. Nine Python controls
pass, including identical default commands/results with a deliberately failing
check, interruption and callback-failure receipts, source refusal before writes,
subset labeling and explicit idle-only cache maintenance. No runtime result is
recovered from the original disk failure, and the broad rerun remains pending.
Author `4bb8a30` updates two stale manuscript status paragraphs only; it does not
change the contract or present the planned spin as completed.

The same student's resumed process completed successfully with source
`8882a18cf98e9a4b70afccfbe980f6344630aae6` and evidence handback `e307d79`.
Root read its source-bound targeted and complete independent fault receipts:
all four groups pass, including actual one-byte partial append, no subsequent
checkpoint replacement, truthful repeated Close and released ownership.
All eleven modules' vet/tests and main/GUI race checks pass. A rejected ordinary
event remains recoverable. The previous process exited 101 without a final
handback; its files remain, and the successful same-conversation resume retains
the original new-only context boundary. No provider or credential access occurred.

The independent lifecycle check also passes against an immutable extraction of
`8882a18`, with 158 runtime/asset/document files bound. A real independent job
finishes while checkpoint replacement is held; the saved anchor still describes
the earlier running state, and reopening that old checkpoint beside the newer
log restores the terminal fact exactly once. A separate invalid-append control
proves continued usability. Two initial reviewer fixture defects, System presence
and the terminal event's name, remain in their first receipts. They do not count
as runtime failures. Bounds and retained integration are the next serialized stages.
