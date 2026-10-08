# Chapter 6 validation

Started 2026-10-07. Initial implementation and live runs are frozen; independent
review and repairs are in progress. This chapter is not validated.
The accepted predecessor is `edition-2-ch05-r1`, an annotated immutable tag at
`7a6ef036322e1cf362894bd30b073fc8399a3c30`. Its main source matches the export
from `185ba767545af567d3d2d3917e807222ff0b2771`.

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Contract and structure | Author, coordinator, student | Root accepted the student's plan in `main/evidence/ch06/student-review.md` against new teaching before affected implementation | Check implementation retains the stated owners, bounds and lifetime |
| Initial implementation and local checks | `/root/coder_ch06` | Initial runtime frozen at `aa5f86a`; all seven modules vet/test and main race pass; inherited incompatible grader result 0/100 retained; initial student experience frozen | Bundle review repairs and repeat affected checks |
| Independent acceptance and mutations | Reviewer, coordinator | Six CLI barrier cases and 61 independent wire cases pass; local checker controls pass; source/binary receipts retained | Complete public, concurrency, capacity, replay, ownership and implementation mutations; no full Chapter 6 acceptance result yet |
| Initial live use | Student, coordinator | Nine sessions / 33 requests frozen at `f73b01e`, `d12a0cb`, `3417575`; exact live executables archived at `bb74fed`. Independent audit passes 16 controls, chronological single-Agent replay, artifacts/usage and public finals; see `chapter-06-live-review.md` | Preserve initial identities and limitations; plan only affected revised demonstrations |
| Historical comparison and revisions | Independent reviewer, original student | R1–R4 source repairs accepted at `75bd14d`, after retained intermediate `3ccaed6` and an escaped-opaque accounting correction. Exact/over boundary independently passes; teaching clarified at `0c10608`, `5df3480`, `3792a19` | Complete full immutable gate and comparative quality record |
| Final live evidence | Student, coordinator | Seven-session affected-path plan accepted; revised helper preflight and nine local controls pass; paid runs released after targeted independent repair acceptance | Bind all revised receipts to `75bd14d`; verify actual runs and retain initial evidence at its own source |
| Manuscript and teaching feedback | Author, student, proofreader | Initial live reconciliation `ca511e4`; student confirms dispositions; first Gemini ambiguity taught at `f085b95`, incremental assembly costs at `0c10608` | Reconcile repair outcomes and independently proofread final chapter |
| Export and checkpoint | Coordinator | Pending all acceptance gates | Export exact accepted source, verify manifest, commit and create immutable `edition-2-ch06-r1` |

The student reads the complete coding skill, architecture and new Chapter 6,
with earlier second-edition contracts and source as needed. First-edition
chapters, historical implementations, later teaching, grader internals and
author/reviewer research are excluded from its initial context. Actual reads
belong in `solutions/edition-2/main/evidence/ch06/`; an instruction boundary
in a shared filesystem is not a claim of operating-system isolation.

The independent grader role previously wrote Chapter 5 checks and review
probes, not that chapter's student implementation. Its historical comparison
begins only after this chapter's initial student attempt and runs are frozen.
Root and the resumed author handle contract questions; the author also prepares
Chapter 7 while the student implements Chapter 6.
The initial independent checker and remaining coverage are documented in
[chapter-06-grader-review.md](chapter-06-grader-review.md).

New Gemini live validation uses the discovered `models/gemini-3.8-flash`
target selected by Bill. All three API adapters require actual human-mode
streaming, tools, interruption and plain-mode demonstrations, with the public
consumer paths in §6.8. Strong deterministic barriers and fault injections
complement those runs. Credentials remain in memory or subprocess environments;
the optional GUI remains an explicitly identified stub.

Local commits and tags are authorized. No push, publication, successful
Chapter 6 outcome or Bill editorial approval is implied by this handoff.
