# Chapter 5 validation

Current gate record, started October 7, 2026. Chapter 5 is in implementation;
it is not accepted. Bill's editorial approval remains separate.

The fresh student `/root/coder_ch05` started with `fork_turns="none"` from
accepted checkpoint `edition-2-ch04-r1` (`55e64115ee5243cac6b4958fd12e6811862ae033`).
The student reads the new teaching and preceding new source; historical
comparison follows the initial implementation and actual runs.

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Contract and predecessor | Author/coordinator | Ready: [outline](chapter-05-outline.md), [contract](chapter-05.md), [accepted Chapter 4](chapter-04-validation.md) | Route new teaching gaps before affected implementation |
| Structure plan | Student/reviewer | Narrow new-contract review passed; [student plan](../../solutions/edition-2/main/evidence/ch05/student-review.md) has no blocking ownership contradiction | Student confirms close-time internal admission, exactly-once terminal recording and helper owner/logger paths in implementation and checks |
| Implementation and local checks | Student/grader engineer | Initial implementation reported complete; [six-module vet/tests and main race receipts](../../solutions/edition-2/main/evidence/ch05/pre-live-checks.json) pass. Inherited CH6 is 10/100; independent contract/fixture diagnosis underway | Resolve CLI close-order wording and run contract-derived independent checks; retain original grader failures |
| Live coverage plan | Student/coordinator | Coordinator reviewed and accepted [plan completeness](../../solutions/edition-2/main/evidence/ch05/live-plan.md): all three human CLI/provider paths, public workflow/collection and labeled deterministic supplements | Finish local path/checker gates, freeze source and run the plan |
| Initial actual demonstrations | Student | Pending; no Chapter 5 live success claimed | Drive human PTYs and public clients, retaining exact source/binary identities and observed outcomes |
| Initial attempt and teaching review | Student | Review started; final initial-attempt record pending | Freeze source and runs before comparative feedback; record difficulty and assistance |
| Independent comparison and revisions | Code reviewer/student | Pending initial implementation and runs | Compare historical standard separately; group findings, revise and review |
| Final validation | Student/reviewer | Pending | Run required checks and affected live paths; retain unchanged evidence under its original identity |
| Manuscript reconciliation | Author/proofreader | Contract reviewed; final observed demonstration and feedback dispositions pending | Reconcile prose with actual receipts and resolve student findings |
| Export and chapter checkpoint | Coordinator | Pending acceptance | Export exact accepted source, verify manifest, commit/tag, then hand off next chapter |

Keep detailed findings in the linked student and review records. Update this
table when a gate changes; progress and restart summaries link here rather
than maintaining competing copies of the current stage.

The early plan review used only the new Chapter 5 contract, architecture and
student plan, without inspecting implementation or the old answer. It is not
code acceptance or the later historical comparison. Its three risks were
returned to the student as existing contract obligations.

Current handoff: the author is clarifying CLI error/Close ordering, and
`/root/grader_ch05` is preparing independent checks from the published
contract. Resuming the original student and starting its replacement both
failed with an orchestration thread-limit error. Its implementation and
student review remain preserved in main; the coordinator owns continuation
after these dependencies clear or a worker becomes available. No Chapter 5
paid run or acceptance is inferred from the local test results.
