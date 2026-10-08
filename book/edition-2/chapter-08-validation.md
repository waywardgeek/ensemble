# Chapter 8 validation

Preferences and execution policy preparation, October 8, 2026. The required
predecessor is accepted checkpoint `edition-2-ch07-r1`. No Chapter 8 student
implementation or live result is claimed. Bill's editorial approval is separate.

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Contract | Author, coordinator | Complete draft `cfb0d87`; reviewed clarification `7200f17`; `chapter-08-review.md` | Preserve corrected Chapter 7 browser ownership in the handoff |
| Independent checks | Coordinator, grader engineer | Initial partial checker `b3f3137`; strict validation fixtures `f6ccbf5`; 11 assertion controls pass; no end-to-end positive yet | Extend §8.9 coverage and obtain student positive before acceptance |
| Student and ownership plan | Future fresh student | Handoff ready after Chapter 7 checkpoint | Supply accepted Chapter 7 and new teaching only |
| Initial implementation and live use | Future student | Not started | Implement, validate, use CLI/browser/public clients on all three providers |
| Historical comparison and revisions | Independent reviewer | Not started | Preserve initial source, live receipts and teaching review first |
| Manuscript and feedback | Author, student, proofreader | Draft explicitly labels actual spin pending | Reconcile actual receipts and resolve student findings |
| Export and checkpoint | Coordinator | Not started | Complete all gates before immutable chapter export/tag |

Initial invocation: `python3 scripts/edition2/accept_ch08.py GUI_BINARY`.
Its scope and remaining independent controls are recorded in
`chapter-08-grader-review.md`. Wire/persistence success alone does not establish
browser usability, audible speech, concurrency or actual policy enforcement.
