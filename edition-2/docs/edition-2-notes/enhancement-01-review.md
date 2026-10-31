# Enhancement review: Chapters 0–1

Reviewer: `enhancement_reviewer`, October 10, 2026. **Accepted after one
correction round.** Independent review covers both enhanced prose and the
Chapter 1 grader/code. Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`.

Read the complete skill, carryover, workflow, voice guide, chapter-writing
procedure, AGENTS and bounded enhancement plan; original Chapters 0–1;
the original Chapter 1 solution, grader/harness/fake; accepted Chapter 1
implementation/tests; student and comparative reviews; historical live/mutation
evidence; enhanced prose, author dispositions and new grader/evidence.
Source revisions are recorded in [the author record](enhancement-01.md) and
[coder record](enhancement-01-code.md). No first-edition or frozen source was
edited during this review.

## Prose and teaching

This is an enhancement of the originals. Chapter 0 retains the failed design
transfer, expert steering, specification/receipt argument and self-wielding
motivation. Its new execution guide explains why live use, independent review,
mutation sensitivity and durable reasons are separate from a score. The
automatic-100 and guaranteed-improvement claims now reflect the actual
experiment and its accepted exceptions. The crossover remains a destination.

Chapter 1 preserves the full two-week bet and restart, hint-discovery story,
model-discovery failure, usage curve, historical conversation, Waywardgeek joke
and seven-check exercise. The logger ambush motivates the new architecture
section before the first code. Its owners are precisely the three objects
already used by this exercise; Observer and GUI remain later boundaries.
Parent interfaces, shared vocabulary and behavior placement are explained in
terms of reachability and conflicting copies, rather than asserted as taste.

The targeted technical changes agree with the accepted implementation:
Agent stages a conversation pair until success, Engine owns accounting,
alternating roles are the exercise's narrower contract, and ephemera are
outside retained dialogue rather than absent from every historical record.
The existing parsing excerpt is explicitly abbreviated. Historical live
figures are labeled and the new PTY receipt is separate; the model's answer
about its own implementation no longer substitutes for inspection.

Initial review found one command-status error: `go run` masks a child exit 2
as wrapper exit 1. Correction round 1 distinguishes the binary's statuses
from the wrapper diagnostic. Reinspection confirms the correction. No other
blocking prose finding remains. The additions sustain the originals' practical
engineering voice and give the reader reasons to care; no invented incident
or expanded exercise was needed.

Independent prose lint reproduces Chapter 1 with no hard failures. Chapter 0
still fails on four pointer closers and one table em-dash, exactly the inherited
hard findings documented by the author. Its first-person register and historical
comparison claims are preserved and explicitly labeled, not newly verified
product measurements. These deliberate preservation limits are accepted under
Bill's enhance-not-rewrite instruction; the combined lint run is **not** passing.
Soft person-gap warnings do not justify inserting anecdotes or padding.

## Code, comments and coverage

The production implementation is unchanged. I inspected its real CLI →
Ensemble → Agent → Engine path, owner accessors, staged history, HTTP block
walk, accounting and failure logging. The historical comparative review's
199 versus 281 active Go lines (first edition versus student) remains the
relevant Chapter 1 comparison; the extra code buys the explicitly required
ownership and public API, without tools or future execution machinery.
Frozen Chapter 1 predates the later exported-member documentation rule:
some fields and methods lack individual comments. This is an inherited
limitation, not a claim of compliance or permission to edit that export.

The new grader is a small separate command and evaluator. It directly reuses
all seven original checks, then strengthens only `wire` with the already
documented exact Anthropic version. It preserves existing failures, removes a
now-false success note only when necessary, and returns through deferred build
cleanup before exiting. It does not inspect implementation names or introduce
a new grading framework.

Measured added Go lines, classified as blank, leading-`//` comment or active
(imports, braces and declarations included): grader/command 73 active,
9 comment, 7 blank, 89 physical; test 38 active, 2 comment, 3 blank, 43 physical.
These are separate support additions, not product growth. Comments explain the
correction and cleanup contract; the exported evaluator is documented. The
short dispatch code needs no filler to meet a percentage.

The test executes the real accepted program against the original provider fake,
compares the entire unchanged check list, then varies captured header evidence
to require a precise wire failure while preserving the other six checks.
This is boundary evidence, not a mock of the behavior under test. The coder's
actual-source wrong-version mutation additionally compiled and reached all
five HTTP requests: original grader 100/100, revised grader 85/100, only `wire`
failing. Existing behavior reuses its independently reviewed mutation evidence;
no new live model call is warranted for this offline verdict correction.

Independent checks passed: focused grader tests and vet, no gofmt filenames,
accepted Chapter 1 uncached module tests, and revised grading of both accepted
and first-edition Chapter 1 sources at **100/100**, with 476 input/65 output.
Read the retained live transcript ending at 350/33, including the original
launcher failure; no repeat paid run or new live provenance is claimed.
Original grader, fake and frozen implementation paths have no diff.

The historical original score, survivor and documentation limitations remain
intact. The separate revised grader result and reviewed teaching now address
the recorded Chapter 1 gaps. No unresolved scope, architecture, coverage or
simplicity finding prevents progression to Chapter 2.
