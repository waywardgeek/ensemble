# Chapter 6 validation

Started 2026-10-07. Implementation is in progress; this chapter is not validated.
The accepted predecessor is `edition-2-ch05-r1`, an annotated immutable tag at
`7a6ef036322e1cf362894bd30b073fc8399a3c30`. Its main source matches the export
from `185ba767545af567d3d2d3917e807222ff0b2771`.

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Contract and structure | Author, coordinator, student | Root accepted the student's plan in `main/evidence/ch06/student-review.md` against new teaching before affected implementation | Check implementation retains the stated owners, bounds and lifetime |
| Initial implementation and local checks | `/root/coder_ch06` | Fresh `fork_turns="none"` student implementing from accepted plan and new-only reading boundary | Retain read ledger and initial teaching review; run module checks and inherited grader |
| Independent acceptance and mutations | `/root/grader_ch05` | `c6272a4`: full coverage plan and initial six CLI barrier cases; harness controls pass; archived Chapter 5 fails the new feature checks as expected | Complete public, semantic, framing, lifecycle and ownership checks; no full Chapter 6 acceptance result yet |
| Initial live use | Student, coordinator | Pending; no Chapter 6 paid calls claimed | Review complete feature/provider/public-consumer plan, then exercise real human PTYs and public clients |
| Historical comparison and revisions | Independent reviewer, student | Pending initial implementation and actual runs | Freeze initial attempt, then compare old Chapter 7 at matching scope; return rationale and review revisions |
| Final live evidence | Student, reviewer | Pending | Bind original receipts to source/executables; repeat only behavior affected by corrections |
| Manuscript and teaching feedback | Author, student, proofreader | Author resumed; first Gemini unknown-text-part ambiguity taught at `f085b95` before affected code | Reconcile student response and actual receipts, then independently proofread |
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
