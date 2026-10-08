# Chapter 8 validation

Preferences and execution policy preparation, October 8, 2026. The required
predecessor is accepted checkpoint `edition-2-ch07-r1` at
`ed5667b38362875c578e88609f4a0fce61946c53`. Fresh student `/root/coder_ch08` has
started with `fork_turns="none"`. No Chapter 8 implementation or live result
is claimed. Bill's editorial approval is separate.

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Contract | Author, coordinator | Complete draft `cfb0d87`; reviewed clarification `7200f17`; `chapter-08-review.md` | Preserve corrected Chapter 7 browser ownership in the handoff |
| Independent checks | Coordinator, grader engineer | Initial partial checker `b3f3137`; strict validation fixtures `f6ccbf5`; 11 assertion controls pass; no end-to-end positive yet | Extend §8.9 coverage and obtain student positive before acceptance |
| Student and ownership plan | `/root/coder_ch08`, coordinator | Fresh new-only handoff delivered; initial ownership plan pending | Check owners, persistence workers, actor ordering and lifetime before affected implementation |
| Initial implementation and live use | `/root/coder_ch08` | Not started | Implement, validate, use CLI/browser/public clients on all three providers |
| Historical comparison and revisions | Independent reviewer | Not started | Preserve initial source, live receipts and teaching review first |
| Manuscript and feedback | Author, student, proofreader | Draft explicitly labels actual spin pending | Reconcile actual receipts and resolve student findings |
| Export and checkpoint | Coordinator | Not started | Complete all gates before immutable chapter export/tag |

Initial invocation: `python3 scripts/edition2/accept_ch08.py GUI_BINARY`.
Strict validation: `python3 scripts/edition2/accept_ch08_validation.py GUI_BINARY`.
Its scope and remaining independent controls are recorded in
`chapter-08-grader-review.md`. Wire/persistence success alone does not establish
browser usability, audible speech, concurrency or actual policy enforcement.

The student receives committed-only new Chapters 1–8 under
`/Users/bill/projects/ensemble-edition-2-revisions/ch08-student-inputs/`, with
hashes and predecessor identity in its manifest. It reads the mandatory coding
skill and architecture directly. Its preceding main source is
`48976424ab75924f98f4ad01f75f8e046ec26872`, tree
`8fb543cc78eccdd9737bd3a690ee9601193fecba`, identical to the Chapter 7 export.
Old chapters/answers, future teaching, author research and grader implementation
remain outside the student context. This is an instruction boundary in the
shared workspace, not a filesystem sandbox.

Independent grader `/root/grader_ch05` develops remaining contract checks in
parallel without consulting the historical Chapter 9 implementation before the
student's initial freeze. Temporary author `/root/coder_ch04` is preparing the
Chapter 10 persistence outline and remains available for teaching questions.
Root owns orchestration, plan review, validation ledgers and checkpointing.
