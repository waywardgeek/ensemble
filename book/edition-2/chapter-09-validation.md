# Chapter 9 validation

Skills chapter work, October 8, 2026. A fresh student has begun the ownership
plan from accepted `edition-2-ch08-r1` (`bfdadaf5`), source `446d7f2` and tree
`58d3fbf4`. Bill's editorial approval remains separate from technical validation.

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Contract | Author, coordinator | Draft `49ae919`, transition clarification `0bca41c`, exact-counter clarification `77cd638`, initial command `f303754` and student-question responses `8f24360` accepted | Student reads pinned clarification and confirms resolution |
| Independent checks | Initial `/root/coder_ch08`; continuing `/root/grader_ch05` | Initial 51-check command `e6c3406`; catalog/graph supplements `0953163`; canned controls and disclosed Chapter 8 absence baselines only | Complete remaining public/browser/concurrency/boundary matrix independently; obtain real Chapter 9 positives |
| Student and ownership plan | Fresh CLI student `01a11c0d-47b8-7241-8834-5ddf57ac5009` | Initial new-only read ledger and plan `5ac45e4`; owners and public API plan accepted; author responses `8f24360` released | Obtain student confirmation before affected implementation |
| Initial implementation and live use | Same fresh student | Core predecessor tests pass; catalog work underway and complete implementation released after clarification reads | Complete implementation/local checks, then submit bounded live plan; no paid runs yet |
| Historical comparison and revisions | Independent code reviewer | Not started | Preserve initial source, live receipts and teaching review first |
| Manuscript and feedback | Author, student, proofreader | Draft explicitly labels actual spin pending | Reconcile actual receipts and resolve student findings |
| Export and checkpoint | Coordinator | Not started | Complete all gates before immutable chapter export/tag |

The draft commit also captured concurrently staged Chapter 7 student files.
The author disclosed the shared-index race; the Chapter 7 student inspected its
own paths and preserved that history. This gives no Chapter 9 implementation
or acceptance evidence. Future commits use explicit owned commit pathspecs.

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
