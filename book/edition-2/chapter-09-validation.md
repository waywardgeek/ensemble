# Chapter 9 validation

Skills chapter work, October 8, 2026. A fresh student is implementing from
accepted `edition-2-ch08-r1` (`bfdadaf5`), source `446d7f2` and tree
`58d3fbf4`. Bill's editorial approval remains separate from technical validation.

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Contract | Author, coordinator | Earlier contract/three-question responses accepted and confirmed; complete encoded skill-record boundary `d8c7738` confirmed by the student | Preserve original-raw import measurement separately from newly emitted write size |
| Independent checks | Initial `/root/coder_ch08`; continuing `/root/grader_ch05` | Complete initial gate passes all 67 rows on `654075b`; reviewer confirms six targeted write/import controls on repair `c0e3171`, including exact-size and intended deletion failures; original fixture failures retained | Reconcile any affected selected rerun and retain each result's original source identity |
| Student and ownership plan | Fresh CLI student `01a11c0d-47b8-7241-8834-5ddf57ac5009` | Initial new-only read ledger and plan `5ac45e4`; owners/API accepted; author responses `8f24360` confirmed, recorded at `080fbca` | Retain initial experience and append new findings |
| Initial implementation and live use | Same fresh student | Initial source `654075b` preserved; raw-import repair `c0e3171` passes all 11 modules and independent reproduction; immutable support controls pass in `local-evidence-controls-review.json`; coordinator released bounded real runs using `review-binding.json` | Execute actual CLI/browser/public matrix, preserving original failures and the 36/provider, 108-total HTTP ceilings |
| Historical comparison and revisions | Independent code reviewer | Not started | Preserve initial source, live receipts and teaching review first |
| Manuscript and feedback | Author, student, proofreader | Draft explicitly labels actual spin pending | Reconcile actual receipts and resolve student findings |
| Export and checkpoint | Coordinator | Not started | Complete all gates before immutable chapter export/tag |

The draft commit also captured concurrently staged Chapter 7 student files.
The author disclosed the shared-index race; the Chapter 7 student inspected its
own paths and preserved that history. This gives no Chapter 9 implementation
or acceptance evidence. Future commits use explicit owned commit pathspecs.

The initial combined gate's pass did not cover the discovered import-encoding
defect. A valid 13,088,007-byte input was rejected after alternate serialization
grew it to 75,674,887 bytes. Repair `c0e3171` separates physical import admission
from emitted-write measurement. The independent original fixture now passes;
write-side exact 67,108,864-byte admission, one-byte overflow and atomic refusal
remain protected. Student regression includes both LF and final EOF framing.
These are local results. Live run release does not claim a completed live spin
or waive the later first-edition comparison and manuscript reconciliation.

## Fresh worker launch

The managed spawn tool returned `agent thread limit reached` twice. Interrupting
a completed worker did not release a slot. The coordinator started a new local
`codex exec` session instead, without resume/fork or inherited conversation, with
memory injection and generation disabled for that invocation. Existing user
model settings and authentication were retained; no credentials were copied.
This is a separately supervised CLI worker, not a managed collaboration thread.

The handoff permits extracted new Chapters 1–9, the mandatory coding skill,
architecture, and the accepted preceding second-edition source. It excludes old
chapters/solutions, future chapters, author/reviewer notes and grader source.
Isolation remains an instruction boundary, not a filesystem sandbox. The student
must record actual reads and exposure in its teaching review. First phase:
owner/state/API plan only, with no runtime changes or paid model demonstrations.
The coordinator reviews that file and resumes the exact session for the next
phase. Launch prompt, events and result are retained outside the repository in
`/Users/bill/projects/ensemble-edition-2-revisions/ch09-student-inputs/`.
