# Chapter 3: author response to student feedback

Recorded October 7, 2026, after `edition-2-ch03-r1` was frozen. This is a
response to the explicitly retrospective
[human-client student review](chapter-03-student-review-retrospective.md),
not a reconstructed initial cold review. The student performed the human-client
integration, not the original six-tool implementation, and had subsequently
read Chapter 4 and review sources. The original tag and evidence remain unchanged.

| Finding | Author disposition | Resolution check |
|---|---|---|
| Chapter 2's tool-only notice, repeated-call fixture, and README became stale when Chapter 3 began automatic continuation. | Accepted. Added a short opening transition in Chapter 3 §3.3: replace the old notice and give the fixture a tool result followed by a final answer. This makes the already-taught loop consequence explicit without changing runtime behavior. | Student confirmed resolution on October 7. |
| Preserve a concrete model correction grounded in calls and disk contents. | Already incorporated in §3.10 before this retrospective report. The exact human corrective follow-up, missing `overwrite:true`, resulting `replacement\ntail\n` bytes, and the model's mistaken description of complete `é` remain visible. No invented repair or production-code fault is claimed. | Student confirmed the current manuscript retains the requested evidence. |
| Evidence verifiers must bind immutable source and executable identity before any derived write. | Accepted and already taught in the mandatory skill and Chapter 0 evidence section. Chapter 3's final record identifies evidence-only revision `339a2a6`; the original live source remains `a347ce3`. Reconstructed requests remain labeled separately from live wire captures. | Student confirmed resolution; prior independent identity-control acceptance remains in the human review. |

The student confirmed all three dispositions on October 7 after reading the
current §3.3, §3.10, Chapter 0 and mandatory skill, and is appending that
confirmation to the supplemental retrospective. The requested resolution
checks in the table are complete; none remains an open teaching issue.

No blocking architecture question was reported. The added transition is a
post-checkpoint editorial clarification for the next manuscript revision;
it does not relabel the original student experience or claim a new live run.
