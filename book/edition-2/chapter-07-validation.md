# Chapter 7 validation

Fresh new-only student `/root/coder_ch07` is active. Its accepted predecessor is
`edition-2-ch06-r1` at `1c6b1f065d11bd3a532c94bc53305394e17e7cc2`. The student's
ownership plan was accepted before affected implementation. Bill's editorial
approval is separate.

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Contract | Author, coordinator | Accepted at `74928f1`; incremental projection lesson `2f105ce`; published initial checker command `7b116b7` | Preserve the new teaching in the fresh student handoff |
| Structure and implementation | `/root/coder_ch07`, coordinator | Root accepted the retained owner/state/lifetime plan before affected code; initial source/read ledger records new-only inputs | Implement and test public watch, pause, optional GUI and reused CLI |
| Independent checks | Grader engineer | Initial partial checker and coverage plan at `a5a07ba`; 15 local transport/snapshot/pause cases, with assertion controls | Add public watch/admission/capacity/teardown and real-browser checks; partial checker is not full acceptance |
| Initial live use | Student, coordinator | Not started; browser driver is available outside the repository | Review complete feature/action plan before paid browser, CLI and public demonstrations on all three APIs |
| Historical comparison | Independent reviewer, student | Not started | Freeze initial implementation, runs and student teaching review before old-answer comparison |
| Revisions and final evidence | Student, reviewer | Not started | Group compatible repairs; repeat only affected live paths with exact new bindings |
| Manuscript and feedback | Author, student, proofreader | Contract prose reviewed; demonstration remains explicitly pending | Reconcile actual receipts and teaching findings, then independently proofread |
| Export and checkpoint | Coordinator | Not started | After every gate passes, export exact source and create immutable `edition-2-ch07-r1` |

The student receives a fresh context, new-edition teaching and the accepted
Chapter 6 source in `solutions/edition-2/main/`. It must read the entire
`book/edition-2/skills/ensemble-coding/SKILL.md` before code and after compaction.
First-edition chapters/solutions, future solutions and author/reviewer research
are outside the initial context. Historical links in the skill do not override
that boundary. Actual reads and help belong in the student's retained review.

The GUI remains a separate optional module. Public watch and pause authority
belong to Agent; browser connections and components use public owner interfaces.
The combined application reuses the CLI and the same Ensemble/Agent. Real
browser synthesis requires observable evidence of produced audio; mocked queue
tests and available voice names alone cannot close that gate.

The initial checker invocation is:

```sh
python3 scripts/edition2/accept_ch07.py GUI_BINARY
```

Its current scope is partial. The complete required coverage remains §7.9 and
the [grader plan](chapter-07-grader-review.md). The preserved Chapter 6 gate is
also required. No student code, browser result, live speech, successful chapter
grade, push or publication is implied by this preparation.

## Handoff and scheduling

Root supplied committed-only copies of new Chapters 1–7 at
`/Users/bill/projects/ensemble-edition-2-revisions/ch07-student-inputs/`, bound
to the Chapter 6 checkpoint by their manifest. Concurrent external edits to
earlier manuscript files are preserved and excluded from this handoff. The
mandatory repository skill and architecture remain direct required reads.

The original author is preparing Chapter 8 while the fresh student prepares its
Chapter 7 plan. Resuming `/root/grader_ch05` for full Chapter 7 engineering hit
the active thread limit. Its initial checker/coverage plan remain available;
resume that independent role when the author's next checkpoint frees a slot.
Root accepted the new student plan before implementation. After the author
checkpointed Chapter 8 at `cfb0d87`, the independent grader resumed Chapter 7
engineering. No Chapter 7 implementation acceptance is implied by its initial
partial checker.

## Early implementation invariants

During the new-contract ownership check, root accepted the public reusable CLI
client outside the internal spokes, importing only the public core and standard
library. The grader will adapt inherited fixture placement while preserving its
assertions; an old `cmd` directory layout is not an architectural requirement.

Root also asked the student to ensure idle watch close releases its actor-owned
recipient registration without requiring a later publication, and that an
overflowed partial projection cannot produce a falsely complete new snapshot.
These are work-in-progress lifetime/completeness findings, before historical
comparison or initial source acceptance; outcomes belong in the student review.

The grader clarified that pause publication precedes the public update return,
but the contract does not require a same-socket observation frame before its
acknowledgement. Wire checks require the matching applied revision/counts and
ordered watch delivery, without inventing a stronger interleaving requirement.
