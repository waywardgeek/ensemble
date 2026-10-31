# Enhancement review: Chapter 6 — accepted

Reviewer: `enhancement_reviewer_resumed`, October 10, 2026. Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`, read entirely before this task.
Carryover, voice/procedure and bounded enhancement plan govern. Scope is original
Chapter 7 streaming only, after Chapter 5 checkpoint `e7f40fa`. Accepted at initial
formal combined review, zero correction rounds. Preparation history below remains
dated by its stage; the final disposition supersedes its pending statements.

## Reading and teaching baseline

Read the entire 642-line original Chapter 7, complete `ch07.md` and
`ch07-code-review.md`, original grader harness/checks, accepted parser/three
stream accumulators/SSE reader, actor changes, CLI/display, external shape
application and streaming/SSE/CLI tests. Compared the corresponding legacy
stream parsers, framing, actor routing, common capabilities and shape application.
The complete historical 26-defect mutation record and round-1 routing/marker
corrections are retained, including the initial reasoning-channel survivor.

Preserve the original blank-cursor motivation, parser identity lesson,
capability-table rationale, terminal flush failure and mutation-audit mistakes.
Streaming changes delivery, not finalized facts. Parser-owned IDs identify parts
within a response; they are not always literal vendor block numbers. OpenAI's
tool-first case demonstrates why list positions are not a valid allocator.
Callbacks are ordered notifications; they do not replace actual parent access
to diagnostics/services. Worker deltas enter the actor FIFO with turn identity,
finals follow stored facts, and only complete tool arguments execute.

Preparation feedback to author corrects the original CRLF claim (trailing CR is
JSON whitespace; framing separator handling is the issue), narrows one-Parse
claims, distinguishes Gemini's response-shaped new content from cumulative text,
and limits raw tracing to sanitized payloads rather than headers/credentials.
Keep the actual `hint:early` reply-routing defect and source-origin marker fix.
Do not add a GUI response-boundary protocol or a callback service bag.

Independent retained-artifact arithmetic confirms measured input/output totals:
Anthropic 5,991/261; OpenAI 3,419/160; original Gemini 2.5 5,826/248; corrected
Gemini 3.8 2,727/936. All cache counters are zero. Anthropic's log contains two
human messages and one hint, despite the raw launcher's three-turn labels.
OpenAI's actual `print STREAM_PROBE` command exits 127 and is reported honestly.
The shape example returns identical numbered text with 40 versus one text delta.

Bill's initially missed Gemini selection remains an error by the student,
handoff and review, not a compliant 2.5 demonstration. The corrected 3.8 capture
records the tool marker, busy hint acknowledgment, then tool release; the next
request carries that hint and successful result. Three stdout reads span only
1.1 ms, so preserve the limit: streamed delivery and correct content, not
perceptibly paced typing. The gated real-CLI regression establishes delivery
before the provider finishes independently of that live timing.

## Grader-plan review

Original results remain accepted 0/100 FAIL and legacy 90/100 FAIL. The original
streaming fixture chooses `fake-model`, explicitly unknown in the chapter and
accepted table, then demands streaming. Legacy declares that name but ignores the
original driver's disabled environment convention. Do not rewrite those scores.

Supported the narrow plan: retain the five streaming-check weights and replace
parity with all corrected Chapter 5 checks; use a declared streaming model plus
a separate unknown-model plain fallback; drive the actual core CLI and separately
build/run the actual public shape example without demanding per-part records
from an aggregate-only application. No new product protocol is justified.

Original `TextByPart`/`FinalText` maps key bare IDs across responses even though
the harness stores response numbers for another check. Both accepted and legacy
actors publish existing `state_changed` to `in_flight` before the next response;
use that boundary, not a heuristic based on reused IDs or the response's terminal
state. Ignore job-result finals when comparing provider response parts. Preserve
valid reuse of the same raw ID in the next response.

Requested actual text/thinking grouping and tool-name/argument correlation,
rather than relying on the original tool total/name-substring check. Canonical
final JSON can differ from verbatim argument display in whitespace/key order;
do not turn that into a new byte-format requirement. Gated CLI, SSE and detailed
cross-vendor/frame/usage guarantees may retain their source-test evidence with
explicit limits on what the numeric grader proves.

Fixture feedback: legacy explicitly declares `fake-model`; a genuinely absent
future-model name tests missing-table fallback without banning one sentinel row.
If the original sentinel is retained, classify any failure as its table/fixture
conflict, not proof that all unknown models are refused. Await final plan/contract
reconciliation and evidence before running a shared harness.

## Draft preparation feedback

Read the entire initial 804-line enhanced prose. All baseline precision
corrections appear, the original story sequence remains, and the new real CLI
routing defect and model-selection correction add concrete teaching. No prose
blocker at this stage. Suggested “may restart” for IDs so response-scoped
uniqueness does not accidentally require every implementation to reuse IDs.
The final grader flags/contract are still pending reconciliation. Independent
lint passes hard checks (4,862 words), with soft negation/short-closer/person-gap
warnings; these do not justify rewriting retained sound explanations.

Independently reconstructed Gemini 3.8 text from all eleven raw response frames:
it equals raw captured stdout and the authoritative final part. The actual
second request includes the retained busy hint; two thinking fragments and one
complete call are present. Original human-demo OS read counts include final
newline-only reads, matching the prose's label; they are not nonempty token counts.

New grader source preparation confirms existing `in_flight` boundaries safely
scope IDs and the absent `unknown-stream-model-edition2` fixture avoids the
legacy sentinel conflict. Requested two narrow captured-request assertions:
enabled/disabled flags must actually differ, and the second request must carry
the successful `STREAM_TOOL_RESULT` rather than merely following an unknown-tool
error. Parent and reviewer also flagged the external report's limited automated
proof: first-delta time presence is not timing accuracy, and text counts do not
establish every finalized-part aggregate. Use already-common fields or disclose
the source-review limit; do not add a report protocol to suit the grader.

Final prose precision reread at SHA-256
`754129c9911367845e2e3add7904dadef98cbf400ea6171a846532c58cd44f6b`:
the existing live response boundary is distinguished from future persistent GUI
identity; IDs may restart rather than being required to restart. The aggregate
check now measures all three delta counts and discloses that timing accuracy and
final-part counts still need source review. Independent lint passes hard checks
at 5,199 words; soft negation/person-gap warnings remain. No new prose blocker.
The grader's current fixture repeats thinking/text in the second response, so
its positive path actually reuses text ID 1 across response boundaries. Captured
request flags and the real successful command result are checked. Final coder
evidence and reviewer execution remain pending; no shared harness was run early.

## Independent execution after harness release

Coordinator released harness ownership after coder execution finished. All four
submitted files match both `validation-overlay.json` and the isolated `e7f40fa`
checkout under the coder's `ensemble-enhancement06-hmxgeevc` scratch directory.
`go test ./internal/gradeedition2 ./cmd/grade-edition2 -run TestChapter6Streaming
-count=1 -v` actually ran the streaming test and passed (0.568 seconds).
Affected `go vet` passes; `gofmt -l` lists no submitted files. No inherited full
suite or paid matrix was repeated.

An independent actual-source mutant changes only the second response's text
`PartFinal` ID by +100 in the copied observation serializer. The unmodified
baseline passes; the mutant fails exactly `deltas-match-final` (0/20), reporting
the orphan delta and final-without-deltas. The other four checks retain their
full points: 36 text deltas, 54 thinking deltas, 14 tool-call deltas, real tool
execution and unchanged finalized content/usage. This checks correlation after
the repeated response-local ID, rather than breaking the first response. The
scratch core compiled and executed; no setup failure is counted as evidence.
Only a canonical-path Go overlay supplied the temporary reviewer test. All
reviewer mutation/overlay scratch is removed; original/frozen files are untouched.
Formal acceptance still awaits the complete final coder report.

## Initial formal combined disposition

**Accepted**, with no unresolved prose, scope, architecture or grader finding.
Read the complete final [coder submission](enhancement-06-code.md), reconciled
[author note](enhancement-06.md), final source and raw baseline/component/mutation
receipts. Manuscript SHA remains `754129c9911367845e2e3add7904dadef98cbf400ea6171a846532c58cd44f6b`.
This is acceptance of the bounded enhancement, not Bill's editorial approval or
a claim that either failing original grade passed. No new runtime exception is
needed; historical accepted departures and their limits remain explicit.

The blank-terminal motivation, three-colour mess, fresh-ID mistake, tool-first
counterexample, buffered-writer failure and mutation-audit corrections retain
their teaching purpose and personality. Added ownership, response-boundary,
execution-contract and live-routing explanations resolve actual difficulties.
The Gemini selection mistake and its corrected run remain candid, including the
1.1 ms timing limit. The chapter gives the reader reasons to implement the
mechanisms before asking for precise observations. Its length is appropriate to
that teaching; no shrink/growth target or soft lint budget was used to remove
useful material. The chapter remains an enhancement of original Chapter 7.

Original actual full results are accepted **0/100 FAIL**, legacy **90/100 FAIL**.
Initial revised full results are **100/100 PASS** and **70/100 FAIL**. Final
revised results are explicitly **composed 100/100 and 70/100**: final executed
streaming components 80/80 and 70/80 plus unchanged full-run parity 20/20 and
0/20. I independently summed those raw reports. Legacy's final delivery failure
records both streamed core requests despite disabling, rejection of the absent
model, and two plain tool deltas in its external application. Inherited parity
failures stay visible. These facts do not justify a compatibility shim or a
rewrite of historical evidence.

All original streaming weights and useful predicates remain. The new driver
corrects model selection and response scoping, retains actual tool execution,
and checks final/delta correlation in both directions. Measured usage includes
zero cache counts; checking each request catches switching only request two to
plain. The external module resolves to the submitted public core and runs its
own report protocol. Existing observer source confirms timing and part counting;
the numeric grader deliberately does not claim to prove those measurements,
complete opaque replay, SSE framing or human-terminal flush. The actual source
tests and their retained mutations cover those obligations. No public protocol,
callback service bag, GUI feature or generic grading framework was added.

Inspected all eighteen targeted mutation results and representative patches.
Each defect fails the intended observable; the second-request-only control
leaves the other four checks passing. The attempted missing-thinking mutation
remains a survivor: its parser delta call was suppressed while accumulated
content remained, so the legitimate whole-part fallback supplied a notification.
Suppressing the actual callback after marking the part seen removes the group
and fails correlation. Requested a factual report wording correction from
“removing accumulation” to this mechanism; no source correction was necessary.
The wrong fake request type, unmatched mutation pattern and misplaced legacy
application probe are setup mistakes, never counted as kills. The independent
second-response final-ID defect above supplies an additional isolated control.
All four exact published sample receipts show successful exits, correct answer
bytes and appropriate actual request flags. Source race checks and affected
package tests/vet pass as recorded; no redundant full or paid run was needed.

### KISS, comments and comparison

Independent nonblank/comment-only counts confirm new grader production:
333 active + 37 comment-only lines, 383 physical; integration test 18 + 2,
22 physical. Original harness/checks remain 533 + 79, 677 physical. Mixed
code/comment lines count as active. Reusing original predicates while replacing
their private contradictory launcher is simpler than a second full grader or
a configurable protocol framework. The roughly 10% new-grader comment share is
below the approximate teaching target, but comments explain the actual difficult
choices: state boundary, internal aliases, job-final exclusion, decoded arguments,
measured totals and proof limits. Exported Chapter6 and the test have useful
documentation; new private containers do not introduce undocumented public fields.
The code is understandable without adding narration to routine fixture handling.

The matching frozen implementation comparison remains applicable and was read
with the actual source: accepted core 3,919 active/946 comment-only/5,161 physical
versus legacy 4,883/1,595/7,153. Against each own predecessor, accepted streaming
adds 714 active core lines and 513 active test lines; legacy core adds zero
because streaming was already present, not because it implemented this chapter
for free. Separate shape examples are accepted 74 active versus legacy 249.
Accepted shared semantic decoding, parser-owned identity, ordered actor delivery
and root-owned trace avoid duplicated final policy; gated CLI and vendor/SSE tests
protect real failure modes. Counts do not substitute for those design benefits.
No frozen runtime was changed in this pass.

The coordinator may checkpoint the owned enhancement artifacts. This reviewer
has made no commit, changed no author/coder artifact, and released no next chapter.
Mandatory next handoff remains `.agents/skills/ensemble-coding/SKILL.md`.
