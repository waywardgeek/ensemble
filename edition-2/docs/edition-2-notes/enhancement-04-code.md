# Edition 2 Chapter 4 grader evidence

> Historical paths: the `edition-2/ch04`–`ch09` paths recorded below use
> original student numbering. Those exports are now preserved under
> `solutions/edition-2/history/`; see the [solution mapping](../../solutions/README.md).

Coder: `enhancement_coder_resumed`, October 10, 2026. Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`. Submitted for independent review;
this report does not itself accept the chapter.

## Scope and contract

Edition 2 Chapter 4 combines original Chapter 4 jobs with the external-consumer
proof from original Chapter 5. The reference endpoint is frozen
`solutions/edition-2/ch05`, accepted at `edition-2-ch05-r2`; its nested `ch05`
directory is the actual separate application. Original Chapter 4 was read at
file revision `0253737c06197d6f9b35e442916b7475c1a59a99`, SHA-256
`44e03445c850c939e331cb25c174e98df692063f824f9cedb748efa3ea1fee84`;
Chapter 5 at `f5d38c997b985deac29c5ed8b1d9e9aacd1af765`, SHA-256
`452a3edc46346bcd48a63483fc7c2ffbbb4056f29446dfc7fb6c292d3a0e4520`.
The complete student and comparative records were read before implementation:
[Chapter 4](ch04.md), [review](ch04-code-review.md),
[Chapter 5](ch05.md), [review](ch05-code-review.md).

The new `gradeedition2.Chapter4` reuses all nine original Chapter 4 checks,
preserving their 100 points. Only `ch3parity` receives the separately corrected
Chapter 3 result, with all checks required. The misleading kill-check title now
says that **subsequent** waits report killed: the original scenario does not
establish that an already-blocked waiter is notified.

The coordinator, author and reviewer agreed a separate 60-point consumer and
boundary score: `external-consumer` 25, `star-topology` 25,
`tool-diagnostic` 10. Total 160 is explicit in the report title and chapter.
Building the separate consumer belongs to its 25-point check; no extra build
points obscure the jobs score. The original Chapter 5 grader's 80% parity
threshold is not used.

The grader accepts `-consumer DIR` and `-consumer-case FILE`. The case is a small
JSON object with `name`, object `arguments`, successful `result` text and
`diagnostic` text. It describes application behavior, not Go identifiers or a
product configuration protocol. The default is the published `shout` example.
The consumer has its own module and runs one demonstration without command
arguments, using the established `LLM_*` environment. Its ordinary human-readable
output is sufficient. Both build and dependency inspection target that same
module-root program; reusing the core builder's optional `cmd/` preference would
risk inspecting one executable while running another.

The real consumer receives a provider fake's call. The grader checks its initial
declaration, nonempty description/object schema, then the exact non-error result
under the supplied call ID in the next request. The fake's final text is unrelated
to the expected result. A tool error, arbitrary tool-result marker or convincing
final answer cannot substitute for the registered handler's output. The diagnostic
check observes the expected line with a parsed timestamp, accepting standard Go
logger and ISO/RFC3339 dates.

`go list` supplies actual import edges and resolved module directories. The check
includes common's outgoing edges, rejects implementation-to-implementation imports,
requires a direct public dependency from the consumer, rejects its direct internal
imports, and verifies that the dependency resolves to the submitted core directory.
Transitive internal dependencies of the public composition root are legitimate.
`internal/common` is Bill's existing binding architecture rule, not a newly guessed
hub or an imposed public API spelling.

## What remains source review

Import edges and output cannot prove that the consumer actually uses the public
composition root, that tool dispatch reaches its actual parents, that common holds
only appropriate behavior, or that mutable application globals are absent. These
remain mandatory independent source-review obligations; the numeric PASS does not
claim to establish them. A blank import plus a hand-written HTTP client would not
satisfy the exercise merely by satisfying these observable checks.

The original Chapter 5 `no-mutable-globals` grep was retired with coordinator and
reviewer agreement, not silently dropped or automatically marked PASS. It exempts
declaration lines containing `names` or `name =`, so mutable state can pass under a
favored name, while a differently named immutable lookup can fail. A grouped `var (`
is detected as a declaration but its members and mutability are not established.
Copying that heuristic or adding a general AST/lifetime framework would not improve
the proof. The original direct-field logger heuristic also rejected the accepted
multi-hop interface chain; no such source-shape predicate is imposed here.

The accepted registration/owner/logger tests and their fifteen historical mutations
remain evidence in Chapter 5's records. Likewise, blocked-waiter notification and
late completion after kill retain the corrected source regressions and mutation
evidence from Chapter 4's review. No new CLI selector or speculative runtime feature
was introduced to expose those library internals. Independent reruns and any
failures belong in [the enhancement review](enhancement-04-review.md).

## Runs and comparisons

Built the original grader with `go build -o <scratch>/original-grade ./cmd/grade`.
Ran `-ch 5 -json` against each untouched endpoint. Then built the revised command
and ran its final source against both endpoints, using the published case explicitly
for the accepted one:

```sh
go run ./cmd/grade-edition2 -ch 4 -json \
  -consumer solutions/edition-2/ch05/ch05 \
  -consumer-case /tmp/consumer-case.json solutions/edition-2/ch05
go run ./cmd/grade-edition2 -ch 4 -json solutions/ch05
go test -count=1 ./internal/gradeedition2 ./cmd/grade-edition2
go vet ./internal/gradeedition2 ./cmd/grade-edition2
```

`/tmp/consumer-case.json` above means the exact published four-field case saved to
a scratch file; the retained run used `published-case.json` under the evidence
directory below. Final results:

| Source | Original Chapter 5 | Revised Edition 2 Chapter 4 |
| --- | --- | --- |
| Frozen accepted `solutions/edition-2/ch05` | 110/120 FAIL, logger shape; nested original jobs 90/100 | 160/160 PASS, exit 0; jobs 100 + boundary/consumer 60 |
| Legacy `solutions/ch05` | 120/120 PASS; nested original jobs 100/100 | 90/160 FAIL, exit 1; jobs 90 + boundary/consumer 0 |

Legacy corrected parity fails `readtools` at 6/10 because search context lines are
missing, the same genuine deficiency found in Chapter 3. It has no separate
`ch05` consumer module, so the three added obligations are unverified/failing.
No later legacy consumer or compatibility shim was imported to manufacture a pass.
The old 120/120 did not establish external-consumer execution. Historical original
Chapter 4's accepted 90/100/media exception and original Chapter 5's 110/120 remain
unchanged in the student records.

The reviewer's focused unchanged-source race run independently found an inherited
intermittent `jobs_test.go:207` assertion, `process-group child survived`: it failed
once initially and once during a five-count rerun. A failure-only diagnostic copy
then passed 20 and 100 iterations, so the reviewer did not capture the failing
process's state and stopped repeating it. Immediate `kill(0)` observing descendant
reaping is a hypothesis, not a verified zombie diagnosis. The blocked-waiter
subtest did not fail. Preserve this source-test limitation alongside the passing
grader/live process evidence; no frozen source, timing margin or runtime behavior
was changed. The independent review records the disposition.

Affected package tests passed uncached in 3.4 seconds; vet passed. The new retained
integration test executes the real separate consumer and import graph. No broad
root suite or paid model calls were repeated. Original provider live demonstrations
and unchanged job mutations are reused from the linked student records.

## Targeted changed-behavior mutations

Each mutation changed exactly one verified source occurrence in a scratch copy.
Both core and consumer were compiled before evaluation, including topology cases;
no compiler failure is counted as a killed behavior mutant. A temporary focused
driver called the same consumer/topology checks used by Chapter4, avoiding a
redundant jobs run for every consumer defect. Correct baseline: consumer 35/35,
topology 25/25. Each row failed exactly the named check(s).

| Defect introduced | Observed focused failure |
| --- | --- |
| Omit initial declarations while preserving dispatch | `external-consumer`: declaration absent |
| Empty the custom declaration's description | `external-consumer`: description/object declaration absent |
| Change declared schema type from object to string | `external-consumer`: object schema absent |
| Return lowercase instead of uppercase input | `external-consumer`: exact result differs |
| Mark correct result as a tool error | `external-consumer`: `is_error=true` |
| Prefix returned call ID with `wrong-` | `external-consumer`: expected ID has no result |
| Register a different name, leaving the requested tool unknown | `external-consumer` and `tool-diagnostic`; actual unknown-tool error, no handler diagnostic |
| Disconnect the handler's logger call | `tool-diagnostic` only |
| Remove logger timestamp flags | `tool-diagnostic` only |
| Add acyclic tools-to-engine import | `star-topology`: implementation imports implementation |
| Have common import a new independent implementation package | `star-topology`: common's forbidden edge |
| Point consumer at the original checkout while grading its copied core | `star-topology`: resolved module directory differs |
| Import core internal/common directly from an `ensemble/client` consumer | `star-topology`: direct private import, although Go permits the related module prefix |

A separate positive run changed the application's registered name to
`uppercase_text` and used a case with `a different sample` → `A DIFFERENT SAMPLE`.
Both consumer checks passed, showing that the default shout example is not a fixed
API/tool-name requirement. Source scratch trees and the temporary repository test
driver were removed; small raw results remain outside the repository.

## Published examples and a corrected false alarm

Extracted the actual JSON fences from `book/edition-2/chapter-04.md`, without shell
reconstruction. The killed-event example passed the accepted binary's offline
`dump`, retaining seq 9, timestamp, killed reason and kind-3 output locator. The
published consumer case passed both consumer checks. Built the accepted CLI from
its own module and ran its real `chat` in a scratch Go project: the exact debugger
argument object started `PAGER=cat dlv debug .`, reached `(dlv) `, and accepted
`send_input` quit. Total session time was 459 ms; the first result reported a
running job and the real prompt. No model call was involved.

The coder initially misread JSON-escaped tool output of Python's `repr` and reported
an overescaped debugger pattern. That claim was wrong. The author and coordinator
checked the source, and the coder then checked both the current fence and the
earlier saved decoded artifact: codepoints `[92,40,100,108,118,92,41,32]`, matching
`(dlv) `. **No regex source edit was needed or made.** The actual debugger run above
confirmed it. This repeated the historical representation trap; decoded bytes
and execution outrank visual interpretation of escaped diagnostic output.

The ordinary dispatch and public logger examples were compared with the accepted
source. The author's added `Finish` comment correctly distinguishes a handler's
return from a launched process reader's completion ownership. No exact public Go
names are imposed by these schematic examples.

## Size and review handoff

New production grader files: 319 physical lines, 256 active, 50 comment-only,
13 blank. New integration test: 19 physical, 15 active, 2 comment-only, 2 blank.
The command grows from 66 to 86 physical lines (59 to 79 active); this is dispatch,
two grader flags and corrected chapter-to-source mapping. Counting uses physical
lines, whitespace-only blanks and lines whose trimmed prefix is `//`; no block
comments occur in these new files. Useful comments explain the observation limits,
module resolution, direct/transitive imports, fake-result independence and build
target alignment. Original job checks are reused, not copied or weakened.

Raw local evidence: `/var/folders/p_/11vk6y114xq7k9g5r_pc3yv80000gn/T/ensemble-enhancement04-bh_lh7ka`:
`accepted-original05.json`, `legacy-original05.json`,
`accepted-revised04-final.json`, `legacy-revised04-final.json`,
`mutation-results.json`, `declaration-mutation-results.json`, the published-example
artifacts, and affected test/vet transcripts. The short tables and corrections here
are the durable record; scratch binaries are not required to interpret them.
Original graders, first-edition prose, frozen exports and unrelated work are
unchanged. Coordinator owns commits and acceptance; no Chapter 5 work is started.
