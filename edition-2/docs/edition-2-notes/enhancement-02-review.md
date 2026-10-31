# Enhancement review: Chapter 2 — accepted

**Current decision: accepted at initial combined prose/code review**, October
10, 2026, by `enhancement_reviewer_resumed`. All three preparation findings are
resolved; zero formal correction rounds were needed. Final evidence and limits
are recorded below. Mandatory handoff: `.agents/skills/ensemble-coding/SKILL.md`.

## Original handoff checkpoint — historical status

Reviewer: `enhancement_reviewer` (`/root/enhancement_reviewer`), October 10,
2026. Paused at coordinator handoff. **Chapter 2 is not accepted.** Initial
prose review found no blockers; code validation and final code review remain
pending. Mandatory resumption/handoff:
`.agents/skills/ensemble-coding/SKILL.md` (read completely before code review
and after compaction), plus the carryover and bounded enhancement plan.

No formal completed code-review/correction cycle has occurred. The findings
below were sent while the coder was preparing its first submission; preserve
them and their resolution rather than resetting any eventual review round.

## Reading completed and initial prose assessment

Read all 1,189 lines of original Chapter 2, the complete Chapter 2 student and
comparative review records, and enhanced Chapter 2 via its complete source
diff against the original. Read author dispositions in
`enhancement-02.md`. Inspected accepted CLI, root, Agent, Engine, History loader
and reducer, Gemini and Anthropic translators, relevant replay tests, original
reference Gemini/reducer, original Chapter 2 checks/Ref checks, and retained live
observations. Prior Chapter 1 review read the full governing skill, carryover,
workflow, voice, chapter procedure and enhancement plan.

The prose remains an enhancement. Its larger additions supply the missing
public command/environment/directive/log contract that stopped the student,
then explain demonstrated architecture and actual corrections. It preserves
the Claude-shaped interface story, migration cost, Gemini compaction failure,
summary multiplication, conflicting accounting, empty-versus-absent failure
and the historical seam bet. CodeRhapsody's first-person opening is explicitly
labeled historical and deliberately preserved under Bill's instruction.

Verified the new environment/directive/nested schema descriptions against the
accepted CLI/loader. The complete illustrative nine-event sample is structurally
consistent on inspection; **coder execution of the exact published fixture is
still pending**. Verified 324/23 Anthropic totals and 718/650 render-byte figures
against `solutions/edition-2/evidence/ch02/live.txt`; these are reused historical
results, not new paid runs. Original 92/100 FAIL and accepted media exception
remain correctly distinguished. Ref preservation is separated from media
rendering and internal reduced-Context facts from black-box CLI observations.
No initial prose correction requested. Final alignment with actual grader
weights, fixtures and results remains to check.

## Preliminary code findings sent to coder

The draft separate Chapter 2 wrapper preserves original checks, substitutes
scope-correct Ref redaction observations, carries the exact-version Chapter 1
correction forward, and adds result-name/capture/ordering probes. Source read:
`internal/gradeedition2/ch02.go`, `ch02_ref.go`, and command dispatch diff.
**This draft has not received final approval or independent execution here.**

1. **Fixture completeness (coordinator finding, confirmed).** Draft
   `interleavedLog` omitted event time, call-level provenance, response usage
   and request/dispatch context. The accepted loader rejected missing call
   provenance during the coder's first baseline. Complete the fixture against
   the public contract, not permissive zero defaults. Coordinator reports
   coder is repairing this; reviewer has not yet verified the final source.
2. **Empty stub can falsely pass (reviewer finding).** Draft `refRedaction`
   tests only `len(result.Response) > 0`, accepting `{result:""}` or
   `{output:""}`. Require actual nonempty stub content in a shape-tolerant way
   and show a targeted empty-stub defect fails for that reason. Current log
   assertions establish original secret, exact Ref and redaction-event presence;
   do not overclaim complete byte-for-byte log identity from those assertions.
3. **New exported field documentation (coordinator reminder, sent to coder).**
   Anonymous wire-decode structs in the new grader have exported fields without
   useful comments. Bill's explicit rule applies to new grader code, including
   internal symbols; add concise contract/purpose explanations and check new
   exported test functions. Do not edit frozen student exports.

The coder was also warned before implementation that the accepted Chapter 2
renderer emits a redaction Stub, not the retained Ref's locator. Exact Kind and
Locator survive in the reducer Context and existing real-source tests. A new
generic grader must not invent a context-dump command, require this student's
public Go spelling, or demand media/locator rendering just to expose that
internal fact. The draft explicitly leaves exact reduced Ref preservation to
source tests and review; this is the appropriate boundary.

## Required next steps

- Obtain coder's final `enhancement-02-code.md`, final grader/test source,
  baseline/legacy results, focused mutation results and execution of the exact
  chapter sample. Initial sandbox/approval delay was reported by coordinator;
  it is not an application result.
- Verify all three preparation findings above, public fixture validity,
  preservation of valid inherited checks and the new checks' actual sensitivity.
  Read relevant existing source/test coverage rather than asserting black-box
  proof of internal Ref state.
- Check new code comments/counts/simplicity, and affected command compatibility.
  Run focused independent checks as warranted; do not repeat broad paid tests.
- Reconcile prose/check weights and evidence, record any formal correction
  rounds (maximum three after initial review), then give combined prose/code
  decision. Progression to Chapter 3 waits for acceptance and checkpoint.

Only this review record was written by the reviewer. No implementation,
first-edition text/grader or frozen source was changed; no commit was made.

## Resumed independent review — in progress

Reviewer replacement: `enhancement_reviewer_resumed`, October 10, 2026. Read
the complete current skill, AGENTS, carryover, enhancement plan, workflow,
voice and chapter procedure after recovering `git status`/`git log`. Read the
complete original Chapter 2 and enhanced source diff, author dispositions,
student review, historical comparative review and live observations. Inspected
the new grader, original checks/harness, accepted History/reducer, translators
and focused replay tests, and matching first-edition reducer/Gemini code.
This replacement preserves the preparation findings above; no formal completed
initial code-review round has yet occurred.

The source comparison confirms the original story and argument survive. The
new contract solves the actual missing-input stop without prescribing a decoding
framework; ownership, Ref/media scope, capture versus replay and the two-result
example explain demonstrated failures. The disclosed historical opener is a
deliberate preservation choice under Bill's enhancement instruction. No new
prose blocker found. Independent lint again reports no hard failures; original
8,473 versus enhanced 9,793 prose words and both person-gap warnings remain
visible. These soft warnings do not justify deleting useful original stories.

Independent focused source checks passed, uncached, in frozen Chapter 2:

```text
go test -count=1 ./ensemble -run 'TestRefCarryforwardAndMissingCallIDs|TestReducerStateAndRedactionReplay|TestGeminiCaptureAndResultName|TestAnthropicInterleavedResultsPrecedeText'
ok ensemble/ensemble
```

They exercise actual reduction/capture/rendering and support the explicitly
separate exact-Ref source obligation. No paid calls or broad mutation sweep was
repeated. Grader execution waits for the coder's stable submission to avoid
overlapping harnesses. Final source comments, mutation evidence, published
fixture execution, counts and acceptance remain pending.

### Source assessment before final execution

The wrapper calls the original runner/evaluator, preserves the existing checks
except the two disclosed media-scope replacements, and folds the exact-version
Chapter 1 correction into parity. The stronger seam checks observe real captured
logs and vendor requests. They do not require a new public Go API, a context-dump
command or the student's struct spelling. The returned-identity probe deliberately
separates the routing alias from the fake's actual model identity; the new
interleaved fixture now includes request/dispatch facts, timestamps, usage and
call provenance. Ref inspection still distinguishes observable dumped facts
from private reduced Context state.

The new support code is a bounded extension of the existing harness, with no
new framework, mocks or product functionality. Both accepted and reference
renderers have different wrappers for Gemini result text, which the nonempty
payload walk accepts without prescribing either. Anonymous decode fields and
the exported entry/test functions now have useful contract comments. Comments
explain the test's evidence boundary, not just its mechanics.

Counting physical Go lines as blank, leading-`//` comment-only, or active
(including imports/braces and fixture literal lines), the two new grader files
have **371 active / 50 comments / 441 physical** lines. The command grows from
48 to 56 active lines, with 3 comments unchanged (55 to 63 physical). The new
integration test has **19 active / 3 comments / 25 physical** lines. These are
grader/support costs: the enhancement adds **zero product-code lines**, and
the accepted Chapter 2 runtime comparison in `ch02-code-review.md` remains the
applicable comparison against each edition's own predecessor. No original
grader was edited or copied wholesale. The grader's 11.9% comment share is
below the approximate teaching target, but its explanations and exported-field
coverage are useful; padding them would not improve the check logic.

## Final combined decision and independent validation

Read the final [coder report](enhancement-02-code.md), its nine pairs of raw
original/revised mutation reports and four raw baseline reports, and the
author's final evidence reconciliation. Accept the stated scope correction:
URI media rendering was premature, while exact Ref preservation remains
required. The original 92/100 student result and 100/100 legacy result remain
historical facts. The revised grader returns **100/100 PASS for both accepted
and legacy Chapter 2**. The published table matches its 13 checks and 100-point
total: 12 inherited checks (three strengthened) plus the replacement redaction
check. Only the deferred media-render check is removed.

Independent completed checks:

- `go test -count=1 ./internal/gradeedition2 ./cmd/grade-edition2` passed,
  including the accepted-solution 100/100 integration test and Chapter 1
  regression. `go vet` on the same packages passed. Gofmt and scoped whitespace
  checks were clean.
- `go run ./cmd/grade-edition2 -ch 2 -json solutions/ch02` independently
  returned **100/100 PASS**, confirming compatibility with the other result
  wrapper and the original solution's ordering implementation.
- In an isolated accepted-source copy, replaced only the reducer's synthesized
  stub text with an empty string via `fmt.Sprintf("%.0s", encoded)`. The mutant
  compiled and the real revised CLI grader returned **92/100 FAIL**, with only
  `ref-redaction` failing for all three kinds: “must retain a result stub without
  the removed text.” This independently resolves the earlier false-pass finding;
  the otherwise successful request census establishes that the intended path ran.
- Independently extracted the exact chapter JSONL block, confirmed its SHA-256
  `06fc7b51af868ce7386008bdee4faf7daf2987fac11901b720b26c5848f9bff9`, built
  the frozen accepted source, and ran every vendor twice without credentials
  against an unreachable endpoint. All pairs were identical; sizes were
  **718/685/756 bytes** for Anthropic/OpenAI/Gemini. The original secret remained
  in every ten-line dump (header plus nine events), disappeared from rendering,
  and the second result and human text survived. Anthropic emitted both result
  IDs before text. The input stayed unchanged. Reviewer scratch was removed.

The nine coder mutations reach the intended checks; their actual scores and
failures match the durable table. The result-name defect correctly triggers
both seam rendering and the redaction result-name observation; it is not
misreported as one isolated failure. The two setup errors are explicitly
excluded from mutation kills. Original-grader survivors remain visible.
Existing source tests and the historical independent Ref mutations supply the
internal state evidence that this CLI cannot expose. No generic grader claim
has been stretched into proof of exact reduced Context or complete dump identity.

The broader root test attempt is **incomplete**, with unrelated script-package
setup errors reproduced by the coder from unchanged HEAD and an unrelated
grader child stopped by the coordinator. Inspection confirms those script
files have no current changes. This review relies on the completed affected
checks and preserved original coverage; it does not claim all root tests pass.
No runtime changes require new paid live runs. Historical human chat, ephemera
and nine-way replay remain reused with their original provenance and caveats.

Teaching and code now agree. The chapter retains the migration costs, compaction
failure, summary multiplication, empty-versus-absent failure and seam bet, while
making the previously missing contract executable. Its additions explain why
the tested boundaries matter, and its examples preserve the reader's concrete
reason to build them. No unresolved prose, architecture, scope, bloat or
correctness finding remains. This accepts the bounded enhancement for the
coordinator's checkpoint; it neither rewrites the student's original grade nor
authorizes Chapter 10 development. Only this review record was edited here.
