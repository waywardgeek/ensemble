# Chapter 6 validation

Started 2026-10-07. Repaired code, deterministic acceptance and independent
live-evidence review pass. Final manuscript review and checkpoint remain open.
The accepted predecessor is `edition-2-ch05-r1`, an annotated immutable tag at
`7a6ef036322e1cf362894bd30b073fc8399a3c30`. Its main source matches the export
from `185ba767545af567d3d2d3917e807222ff0b2771`.

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Contract and structure | Author, coordinator, student | Student's initial and repair ownership plans accepted against new teaching before affected implementation; implementation reviewed | Preserve the stated owners, bounds and lifetime |
| Implementation and local checks | `/root/coder_ch06` | Initial runtime `aa5f86a` retained; repaired runtime `75bd14d` passes all seven modules vet/test and main race; inherited incompatible grader result 0/100 retained | Preserve both source identities |
| Independent acceptance and mutations | Reviewer, coordinator | Full immutable gate passes at `5a95578`: 23 command groups, 14 contract groups, six CLI barriers, 61 wire cases, retained Chapter 5 assertions, 16 implementation mutants and two CLI recovery mutants; three supplementary mutants pass at `340c678` | Bind exact receipts in checkpoint manifest |
| Initial live use | Student, coordinator | Nine sessions / 33 requests frozen at `f73b01e`, `d12a0cb`, `3417575`; exact live executables archived at `bb74fed`. Independent audit passes 16 controls, chronological single-Agent replay, artifacts/usage and public finals; see `chapter-06-live-review.md` | Preserve initial identities and limitations; plan only affected revised demonstrations |
| Historical comparison and revisions | Independent reviewer, original student | R1–R4 repairs accepted at `75bd14d`, after retained intermediate `3ccaed6` and escaped-opaque accounting correction; final comparative record at `5a95578`. Teaching precedes each affected correction | Preserve comparison rationale and original attempt |
| Final live evidence | Student, coordinator | Accepted: seven sessions / 20 requests frozen at `8c73f9b`, bound to `75bd14d`; 16 independent controls, exact replay, artifacts, usage and public finals pass. All 75 original files unchanged; binaries archived | Bind revised audit receipt; earlier unaffected demonstrations retain initial source |
| Manuscript and teaching feedback | Author, student, proofreader | Student confirms repair teaching dispositions; author reconciling revised receipts and final gate; initial account remains preserved | Independently proofread complete revised chapter |
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
