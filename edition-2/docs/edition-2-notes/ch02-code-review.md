# Chapter 2 comparative code review

Reviewer: `review_restart_ch01`, October 9, 2026. Current verdict:
**accepted with the scoped exception below**, after revision round 1 of at most
3. R1 is resolved; no blocking findings remain. The original grader still reports
**92/100 FAIL**. Mandatory handoff: `.agents/skills/ensemble-coding/SKILL.md`.

Initial verdict was **needs correction** for R1. The initial findings and counts
are preserved below; the final reassessment and updated counts follow them.

Read the entire skill, AGENTS.md, current carryover/workflow, all 1,189 lines of
original Chapter 2, the completed [student review](ch02.md), all new source and
tests, evidence, and the matching first-edition solution. Compared each edition
with its own Chapter 1 predecessor (`solutions/ch01` and frozen
`solutions/edition-2/ch01`). Chapter 2 source SHA-256 matches the recorded
`b2ebe8bd541a9be0a025280ab594b34b5ea1378a7eecfa8ec9a314e02c8615fd`.

## Required correction

**R1 — Keep all Anthropic tool results before text in a merged user message.**
At `solutions/edition-2/main/internal/engine/anthropic.go:47`, merging always
appends the next blocks. An imported log containing a reply with calls `one`
and `two`, result `one`, a human message, then result `two` renders successfully
with the final content types:

```text
tool_result(one), text, tool_result(two)
```

I reproduced this through the actual built binary's offline `render` command,
with no key and an unreachable endpoint. Chapter 2 §2.6, Exhibit B explicitly
requires tool-result blocks before any text in that message. The nearby comment
currently promises an ordering the implementation does not guarantee.

Correct the rendering of this existing log vocabulary, preserving both result
identities and the text. Add a focused regression and demonstrate its sensitivity
to restoring the faulty ordering. Rerun affected checks and the original grader.
This finding requires no live tool execution, hint command, concurrency or new
event type. It is a translation rule in the current exercise.

## Scoped exception and explanations

I accept the student's explanation for the **actual 92/100 FAIL**, independently
reproduced in this review: only `ref-render` and `ref-redaction` fail. Section
2.1 says the blob shape precedes capability; §2.5 explicitly says no blob is
exercised here and assigns the first use to Chapter 4. The grader's added positive
Gemini URI rendering and unredacted media control conflict with that scope.
Documented `fileData` fields establish wire spelling, not chapter authorization.
The original grader and chapter must stay unchanged during this attempt.

Ref carryforward itself is expressly required and is **not waived**. The student
tests check exact Kind 1/2/3 and Locator after reducer redaction, retained calls,
removed secret text, rendered stubs and unchanged original dumped facts.
Independent uncached tests passed. Reviewer scratch mutations removing the Ref
or reusing the logged part array fail specifically with `redaction changed exact
Ref` and `redaction mutated append-only facts`. Those copies were removed.
This supplies direct evidence for the required behavior without media rendering.

The grader fixtures' `generatecontent`/`generate_content` inconsistency is real.
The student's one documented input alias and canonical output spelling are a
small acceptable contract repair; unknown surfaces still fail. No broader enum
normalization is required.

Carry this exception into later original parity checks: inherited Ref media
failures remain visible and do not silently become Chapter 3 requirements.
Reassess when original Chapter 4 actually introduces blob capability. This is a
sequencing exception, not permanent removal of that capability or a declaration
that the original grader passed.

I also accept the student's explanations for recording failed exchanges as
events, preserving empty OpenAI text, and distinguishing historical Context
usage from current Engine exchanges. The chapter changes the representation;
the seven Chapter 1 grader properties still pass. The implementation records
assistance and failed experiments honestly rather than claiming the published
chapter supplied its missing protocol.

## Comparison and simplicity

Counts classify each physical Go line as blank, comment-only (leading `//`), or
active, including declarations/imports/braces. There are no block comments.
Tests and the optional GUI module are separate; there are no moved files between
categories, generated sources or additional production languages.

| Measure | First edition Ch1 → Ch2 (delta) | Student Ch1 → Ch2 (delta) |
| --- | ---: | ---: |
| Production active Go | 199 → 1,852 (+1,653) | 281 → 1,408 (+1,127) |
| Production comment-only | 48 → 651 (+603) | 59 → 249 (+190) |
| Production physical | 278 → 2,762 (+2,484) | 369 → 1,755 (+1,386) |
| Test active Go | 0 → 188 (+188) | 123 → 453 (+330) |
| Test comment-only | 0 → 77 (+77) | 6 → 16 (+10) |
| Test physical | 0 → 288 (+288) | 133 → 488 (+355) |

Student core production is 1,400 active / 246 comments / 1,742 physical;
GUI production is 8 / 3 / 13. Core tests are 424 / 14 / 454; GUI tests are
29 / 2 / 34. Module metadata and the small text/JSON evidence are excluded.
No mutation harness or duplicate mutant source is retained.

The student adds 31.8% fewer active production lines for this chapter. That is
supported by concrete simplifications: explicit tagged data, centralized
validation/reduction, one synchronous transport path, and small owner interfaces.
No speculative services or future orchestration explain the size difference.
The comparison is not feature-identical: the scoped media behavior is deferred,
and the reference also provides extra CLI convenience behavior outside the
exercise. Neither omission should be credited as an architectural improvement.

Production comments are 15.0% of nonblank student lines versus 26.0% in the
reference. Although below the approximate 20% target, they explain the difficult
mechanisms where needed: ownership, delivery/replay ordering, disjoint accounting,
producer identity, empty content, and summary folding. The code is readable
without the reference. Do not pad the percentage. The `refuseBlob` comment has
the minor typo “There media capability”; it can be corrected with R1 but is not
an acceptance blocker.

## Architecture, tests and live evidence

The real CLI constructs the public Ensemble. Ensemble retains Agents; Agent
retains Engine and History with checked interface back-pointers. Configuration
remains Agent-owned, transport/usage Engine-owned, and log/reduction History-owned.
Implementation spokes import common rather than one another. Common holds data,
interfaces and standard JSON/string dispatch; no runtime service collection or
mutable application global is present. Vendor functions retain Engine access.
Offline rendering uses the same translators as live sending.

The separate GUI module constructs public Config and uses the same real Agent as
the ordinary client. Its test runs actual HTTP through a provider fake. Independently
checked dependencies: core has no GUI-module dependency, while the GUI module
is detected by the same check as a positive control. No browser functionality
has been added early.

The student tests use faithful external HTTP fakes, not mocks of earlier machinery.
They add useful coverage for reduction, Ref preservation, atomic rejected loads,
exact capture, large numeric arguments, empty reply replay and the public client
boundary. Reference tests concentrate on CLI guidance/default names and empty
OpenAI content; those are different coverage sets, so test volume alone proves
no superiority. The shared original grader still supplies the broad three-vendor
exercise coverage. R1 exposes a missing ordering case in the current suite.

Independent validation: formatting clean, `go vet ./...` and uncached
`go test -count=1 ./...` pass in both modules; original Chapter 2 grader is
**92/100 FAIL**. The coordinator's unchanged reference baseline is **100/100**.
Read all 55 recorded mutation executions and their five initial survivors.
The repaired tests and candid distinctions between grader gaps, weak fixtures
and redundant guards are sound. Besides the two redaction mutations above,
independently reproduced the captured-model mutation failing the intended
producer assertion. No full mutation sweep was repeated.

Read the live observations and four actual JSON logs. The three-provider PTY
recall runs have matching recorded totals: Anthropic 324/23, OpenAI 229/14,
Gemini 188/12 input/output. The initial OpenAI 404 and the ephemera response's
incorrect self-correction remain disclosed. I accept this student-run live
evidence without repeating paid calls. Independently dumped and rendered all
nine source/target pairs twice without keys; byte identity and every reported
byte count matched. No claim of live media/tool execution, nonzero cache tests
or full opaque replay follows from those text runs.

Initial handoff requested R1, updated student notes and targeted evidence for
revision round 1. No other blocking finding was identified.

## Revision round 1 — accepted with exception

Read the entire mandatory skill again, the student's appended round-1 review,
the changed renderer/comment, new regression, mutation record and separate
`grader-r1.txt`. The correction applies a stable results-first partition only
to completed Anthropic user messages. It retains original block bytes, result
IDs, outputs and within-group order. It does not mutate the context or log, add
an input protocol, or move behavior out of its responsible translator. The
20 added active production lines are a clear, proportionate correction.

`TestAnthropicInterleavedResultsPrecedeText` exercises the real loader, reducer
and renderer and checks ordering, both result identities/outputs, human text
and unchanged Context. Independently reran clean formatting, vet and uncached
core tests successfully. In an isolated scratch copy, removed the final
partition: the regression compiled and failed specifically with `tool results
must precede text`, showing the original result/text/result order. The scratch
copy was removed. R1 and its nearby explanation are now correct; the unrelated
comment typo was also fixed.

The student's round-1 original grader result remains **92/100 FAIL**, with exactly
the same `ref-render` and `ref-redaction` media prerequisites failing and every
other check passing. Its affected core/GUI checks passed. No paid live run or
whole mutation sweep was repeated for this wire-ordering-only correction; the
earlier live and nine-way replay evidence remains applicable.

Final independently counted core production: **1,420 active / 249 comment-only /
1,765 physical**; core tests: **467 / 16 / 500**. GUI remains **8 / 3 / 13**
production and **29 / 2 / 34** tests. Thus combined production is **1,428 active /
252 comments / 1,778 physical**, a Chapter 2 delta of **+1,147 / +193 / +1,409**
against the student's frozen Chapter 1. Combined tests are **496 / 18 / 534**,
deltas **+373 / +12 / +401**. Counting method is unchanged. Production comments
remain 15.0% of nonblank lines; the active chapter addition remains 30.6% below
the first edition's +1,653, with the feature differences already qualified above.

Accept Chapter 2 with the explicit media-scope exception, not a passing original
grader claim. Preserve the two actual failures in later parity results and
reassess when Chapter 4 introduces blob capability. One correction round was
used; no unresolved architecture, scope, bloat or correctness finding remains.
