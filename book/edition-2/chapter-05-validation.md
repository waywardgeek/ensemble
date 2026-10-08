# Chapter 5 validation

Accepted for checkpoint `edition-2-ch05-r1`. Bill's editorial approval is
separate. Independent code, live-evidence and manuscript acceptance is recorded
at `c1cc0b4`; the earlier code/live audit is `9760fb2`.

The student started in a fresh context from accepted Chapter 4, using the new
teaching and preceding new source. Historical comparison followed its initial
implementation and real-provider attempts. The reviewer previously engineered
the independent grader, but authored neither student runtime nor chapter prose.
This exposure is disclosed in [the code review](chapter-05-code-review.md).

## Checkpoint identities

- Initial student runtime: `8aa40c3e840af575724a6b895cf59c060633a9b1`.
- Initial attempts: `5f2576c`; completed original handoff: `525aa10`.
- Reviewed runtime: `959c663400b74927578a3609ce58b0a51263e654`.
- Revised evidence freeze: `469730f7217e09620d471f8a71825a92d760113f`.
- Export source: `185ba767545af567d3d2d3917e807222ff0b2771`; only README and
  historical checkpoint labeling changed after the evidence freeze.
- Exact 1,254-file export: `solutions/edition-2/ch05/`; manifest:
  `solutions/edition-2/manifests/ch05-r1.json`.

## Gate record

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Contract and structure | Author/coordinator/reviewer | Accepted; common ownership interfaces, one actor, report transactions and request completion reviewed | Preserve the corrected rules in later chapters |
| Implementation and local checks | Student/grader engineer | Six modules pass vet/tests; main race and formatting pass; independent checker 100/100 with eighteen deletion controls and passing controls | Retain exact source/checker bindings |
| Actual demonstrations | Student/reviewer | Accepted: eight unchanged Anthropic/OpenAI paths plus four Gemini 3.8 modes and two revised input paths | Preserve each run's original source identity |
| Student teaching review | Student/author | Initial review and subsequent confirmations retained; close timing, Gemini mapping and lifecycle repairs reconciled | Carry lessons into subsequent teaching |
| Historical comparison | Independent reviewer/student | Accepted R1/R2 repairs and concrete design/comment comparison at the corresponding first-edition scope | Preserve the initial attempt and reasons for revision |
| Evidence verification | Independent reviewer | Fourteen runs, 78 exact replays, 38 call/result pairs, 204 receipt-manifest hashes and 59 historical sources per binding checked | Retain original masked controls plus corrected independent supplement |
| Manuscript | Author/proofreader | Four author records accepted at `18fdcb8`/`9849ad7`, bound by final review `c1cc0b4`; no hard prose-lint failure | Bill's editorial approval remains separate |
| Export and chapter checkpoint | Coordinator | All 1,254 export hashes verified; all six exported modules pass vet/tests; manifest binds the chapter checkpoint | Extend this accepted source in a fresh Chapter 6 student context |

Detailed evidence: [grader review](chapter-05-grader-review.md),
[code and manuscript review](chapter-05-code-review.md),
[student feedback](chapter-05-student-feedback.md),
[independent receipt audit](checkpoint-evidence/ch05-review-final-receipts.json),
[coordinator source audit](checkpoint-evidence/ch05-coordinator-source-audit.json)
and [export checks](checkpoint-evidence/ch05-export-checks.json).

## Improvements and retained limits

Compared with the first-edition standard, the reviewed solution has one turn
owner, request-specific reusable completions, ownership interfaces, reliable
collections independent of display callbacks, transactional report cursors and
exact request replay. Review corrected process-input writes that could park the
actor and closed subscriptions that retained client resources or admitted new
workers after application shutdown. The chapter teaches those rules before the
revised implementation. Process-input cancellation and subscription lifecycle
have independent distinguishing probes, not just nominal live demonstrations.

Gemini 3.8 Flash passed hint receipt, wire delivery, consumption and observed
compliance; queued prompts, interruption, later job supervision, input and
shutdown; orderly EOF; public three-Agent workflow; and completion collections.
Earlier Gemini failures and the inconclusive grouping diagnostic remain dated
attempts. No renderer change was made to fit that unproved hypothesis.

The student's original thirteen identity negatives stopped at a path guard,
so they did not establish their advertised identity properties. Independent
controls start with a passing valid-path replay; thirteen single mutations
reach their intended refusals before replay or derived writes. All 247 original
raw receipt hashes remain unchanged. Correcting those fixtures required no
runtime change or repeated paid generation.

The GUI is a stub. Actual terminal and process evidence is macOS, not a Linux
live claim. Scoped tests and mutation checks are not exhaustive proofs. Local
nonkillable functions and arbitrary blocked callbacks cannot be forcibly stopped.
The historical Chapter 6 grader remains unchanged; its incompatible 10/100
fixture result is preserved separately from the new contract's acceptance.
Historical reference/mutation tests and the full root regression passed. No
first-edition implementation was edited and no push is authorized.
