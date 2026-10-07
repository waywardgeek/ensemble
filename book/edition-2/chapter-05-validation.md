# Chapter 5 validation

Current gate record, started October 7, 2026. Chapter 5's initial implementation
is frozen and live validation is underway; it is not accepted. Bill's editorial
approval remains separate.

The fresh student `/root/coder_ch05` started with `fork_turns="none"` from
accepted checkpoint `edition-2-ch04-r1` (`55e64115ee5243cac6b4958fd12e6811862ae033`).
The student reads the new teaching and preceding new source; historical
comparison follows the initial implementation and actual runs.

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Contract and predecessor | Author/coordinator | Ready: [outline](chapter-05-outline.md), [contract](chapter-05.md), [accepted Chapter 4](chapter-04-validation.md) | Route new teaching gaps before affected implementation |
| Structure plan | Student/reviewer | Narrow new-contract review passed; [student review](../../solutions/edition-2/main/evidence/ch05/student-review.md) records the three risks and their implementation checks | Inspect actual ownership and lifecycle during comparative review |
| Implementation and local checks | Student/grader engineer | Frozen initial source `8aa40c3`; six-module vet/tests, main race and new independent checker 100/100 pass. Inherited CH6's incompatible 10/100 remains recorded separately in [grader review](chapter-05-grader-review.md) | Finish independent checker receipt checkpoint; retain original defects and fixture failures |
| Live coverage plan | Student/coordinator | Feature coverage accepted. Bill now selects Gemini 3.8 Flash; exact `models/gemini-3.8-flash` and GenerateContent support verified in [discovery receipt](../../solutions/edition-2/main/evidence/ch05/gemini38-discovery/receipt.json) | Run all four Gemini modes on 3.8 under an addendum; retain unchanged Anthropic/OpenAI evidence |
| Initial actual demonstrations | Student/reviewer | Anthropic/OpenAI paths passed. Earlier Gemini workflow/collection/EOF successes and control failures remain historical, outside the revised 3.8 target. Latest frozen handoff `525aa10` retains 61 exact PTY request replays, plus two separate diagnostic calls | Review detailed hint teaching and any compatible code corrections before the bounded 3.8 demonstrations |
| Initial attempt and teaching review | Student | Source `8aa40c3` and actual attempts `5f2576c` preserved; [student review](../../solutions/edition-2/main/evidence/ch05/student-review.md) records clarification, diagnostic defect, live difficulty and assistance | Append final provider outcome and subsequent review effects without rewriting the initial freeze |
| Independent comparison and revisions | Code reviewer/student | Assigned to grader engineer after its checker checkpoint; independent of student implementation, with grader exposure disclosed | Compare historical standard separately; group findings, revise and review |
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

Current handoff: the student confirmed the author's CLI error/Close clarification
from manuscript checkpoint `5ae7649` and added a deterministic ordering test.
`/root/grader_ch05` is finishing checker receipts, then taking comparative
review. A separate reviewer allocation failed with a thread-limit error, so
this role reuse is explicit. The original `/root/coder_ch05` continues in its
existing student context.

The Gemini grouping diagnostic succeeded for both the exact originally failed
body and a split variant. That does not establish a grouping defect or justify
changing the renderer. The final bounded control attempt used unchanged source
and failed again; raw evidence is in the student's
`evidence/ch05/controls-gemini-final-attempt/`. Paid retries have stopped and
the reviewer has these paths for investigation. Initial summary messages
and commit prose incorrectly counted 62 replayed requests; the machine report
and manifest record 59, and the student appended the correction.

Bill subsequently narrowed Gemini validation to 3.0 Flash and newer and selected
3.8 Flash for hints. The new §5.3 wire example makes the existing result/hint
mapping explicit; §5.10 names the discovered target. No further older-model
attempts are planned. The source remains unchanged unless independent review
identifies a correction. The later `525aa10` receipt freeze adds the last failed
control attempt, bringing its PTY replay total to 61; that is a later scope,
not a correction to the earlier 59-request report.
