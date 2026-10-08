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
| Contract | Author, coordinator | Draft `9181683`, clarifications `6cf8660`/`b4e3fb7` and inherited hint-state clarification `8f24360` accepted; see chapter-10-review.md | Retain exact compatibility and semantic-state obligations |
| Independent checks | `/root/grader_ch05` | Initial CLI/store/lock/corruption command published at `dfb4a72`; eleven canonical examples, one outer positive and eighteen intended refusals pass; preceding CLI passes 3 standalone controls with 90 session-feature absences | Extend semantic/public/browser coverage against genuine student exports; no Chapter 10 runtime positive exists yet |
| Fresh student and owner plan | Fresh CLI student `01a11cc3-9e40-7d62-a7b5-9b2ec4c928c0`, coordinator, independent reviewer | Plan `7cb8429` accepted at `72bf621`; Q1–Q3 published at `af5a762`, narrow proofread `6766995` | Student reads pinned complete chapter/direct feedback and records acknowledgment before code |
| Implementation and local gates | Same student, grader | Phase 2 resumed for local implementation; no runtime acceptance yet | Publish complete semantic codec/public API, extend independent checks, validate full contract and inherited behavior |
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
