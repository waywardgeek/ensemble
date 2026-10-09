# Chapter 6 comparative code review

Reviewer: `review_restart_ch01`, October 9, 2026. **Needs correction at initial
review; zero correction rounds used (maximum three).** Actual original grading
is **80/100 FAIL**, independently reproduced. The two failed grader checks have
justified exceptions below; acceptance is withheld for the separate findings.

Read the entire mandatory `.agents/skills/ensemble-coding/SKILL.md`, AGENTS,
carryover, workflow, original Chapter 6 and [student review](ch06.md), including
the student's clarification during initial review. Chapter SHA-256:
`fc82efd7fc0caaa9805a7bfa7f2b8b457cfb096b341f0c8f92d60b672dff2495`.
Compared working `solutions/edition-2/main` with frozen `edition-2/ch05`, and
original `solutions/ch06/agent` plus its external examples with `solutions/ch05`.

## Required corrections

1. **Implement and demonstrate the missing coalesced completion capability.**
   Original §6.6 requires a parent waiting for two Agents to receive both relevant
   completions in one wake when they finish in the same turn window. Current
   `Ensemble.WaitAny` consumes one lossy observation; private request replies
   correctly protect individual requests, but neither mechanism delivers the
   required grouped completion. The sequential three-role exercise and current
   observer test do not exercise this capability. The student now explicitly
   acknowledges the valid omission and withdraws its author-phase-only treatment.
   Keep completion tracking reliable and separate from droppable progress, retain
   request/Agent identity, and add a focused concurrent regression plus a mutation
   that loses or separates the grouped delivery. Do not infer a timer, debounce
   interval, generalized scheduler or new collaboration workflow from this wording.
   The existing sequential application remains valid. The grader's `wake once`
   PASS only reports three named completion records and does not verify §6.6.

2. **Make the media test protect the actual Gemini wire structure.**
   `ensemble/media_test.go` currently checks base64 bytes and the substring
   `inlineData`; all four Gemini cases still pass when the renderer's
   `"mimeType": p.MIME` is renamed to `"wrongMimeType": p.MIME`. I reproduced
   this one-line mutant in an isolated copy with
   `go test -race -count=1 -run TestLocalMediaIsEncodedAndUnsupportedMediaIsRefused ./ensemble`:
   exit 0, `ok ensemble/ensemble`. Current production spelling is correct. This
   is a sensitivity gap in the claimed local-media wire compatibility, not a
   request for more media capabilities. Assert the relevant decoded attachment's
   MIME and byte fields in the existing focused test, grounded in the
   [official Blob schema](https://ai.google.dev/api/generate-content#Blob), then
   show this malformed-wire mutant failing the intended assertion. No additional
   paid run or generic wire-validation framework is needed.

3. **Correct one misleading exported-field comment.**
   `internal/engine/anthropic.go`, `anthropicBlock.Input`, describes a provider
   token count before cache normalization. This field actually contains tool-use
   arguments. The new documentation requirement concerns accurate contracts as
   well as coverage; fix the copied description. The broader documentation pass
   is useful and does not need another wholesale rewrite.

## Design, scope and teaching

The implementation materially improves the actor seam. One Agent owns its
mailbox, conversation and turn lifecycle; Engine performs an HTTP exchange and
returns facts. HTTP, tool execution and callback reporting do not park the actor.
Workers report through the mailbox, while process/output state stays with Jobs.
The request's buffered reply is independent of bounded progress. Save failures
reach that caller before successful acknowledgement. Queued requests receive
distinct replies, and shutdown rejects admission, cancels transport, stops managed
processes and joins remaining work. The default and positive tool-round limits
preserve call/result pairing and allow a subsequent turn.

Bill's interrupt ruling is implemented concretely: the current wait receives an
honest interrupted result, the Agent accepts a new turn, and other jobs retain
their real output. A late callback becomes a retained `job_report`, not a second
conversational result for an already paired call. This distinction is explained
beside the code. Interrupt, job kill and shutdown remain different operations.

Ownership remains real: root retains Agents; Agent stores its Ensemble interface
and installs Engine, History and Jobs; those children retain the Agent. Dispatch
contexts reach the same Engine and root logger through actual parents. Common
contains shared vocabulary, not mailbox or rendering behavior. Independent grader
import inspection confirms five clean spokes and no common-to-spoke dependency.
CLI and both external examples use the public root; GUI remains a separate small
module. There are no new service bags, mutable application globals or mock layers.

The original snapshot keeps a separate actor wrapper and synchronous Ask path,
uses shared observations for blocking completion, and places mailbox behavior in
common. Its callback execution also waits behind the Go handler. The student
removes those competing paths and tests the failure modes directly. The original
snapshot additionally contains SSE parsing and vendor streaming hooks, despite
§6.11 deferring streaming; that extra scope confounds any simple size ranking.
The student correctly avoids importing those features.

Local `RefPath` media becomes an actual attachment containing file bytes, with
explicit model capabilities and errors naming unsupported categories. Ref metadata
survives redaction. URI/handle attachments remain deferred, not silently consumed.
The external program registers and executes the three roles' real manuscript
tools; its stored edited draft drives the next Agent. A third role needs no core
special case. The application remains ordinary sequential Go orchestration, not
an early sub-agent or workflow framework.

Core production comments are 20.8% of nonblank lines (original 24.6%). The added
explanations of admission, private replies, stale reports, joining workers and
save ordering teach the important reasons. Exported declarations, fields and
interface methods have comments, including internal types and the existing
inline Ref.Kind explanation; the accuracy defect above remains. Repeated comments
on anonymous JSON fields are less instructive than the lifecycle explanations,
but satisfy Bill's explicit documentation scope rather than justify new runtime
machinery. The external application has a lower comment ratio but explains its
ownership and handoff choices; adding filler to equalize ratios would not help.

## Comparative counts

Physical Go lines are classified as blank, comment-only (leading `//`, with block
comments accounted for), or active, including imports, braces and literal data.
Mixed code/comment lines count as active. Tests and separate example/GUI modules
are distinct. Module manifests/sums, logs, evidence, binaries and prose are
excluded; there is no generated Go or new non-Go implementation. Deltas use each
edition's own Chapter 5, accounting for the original core's move beneath `agent/`.

| Go category | Active | Comments | Physical | Chapter delta: active / comments / physical |
| --- | ---: | ---: | ---: | ---: |
| Original core production | 4,883 | 1,595 | 7,153 | +1,559 / +595 / +2,393 |
| Original ch05 example present in ch06 export | 99 | 15 | 125 | +99 / +15 / +125 in exported tree |
| Original ch06 example | 261 | 31 | 328 | +261 / +31 / +328 |
| Student core production | 3,205 | 841 | 4,303 | +809 / +444 / +1,377 |
| Student GUI production | 8 | 3 | 13 | 0 / 0 / 0 |
| Student ch05 example | 74 | 13 | 92 | 0 / +1 / +1 |
| Student ch06 example | 267 | 20 | 295 | +267 / +20 / +295 |
| Original core tests | 548 | 125 | 728 | +200 / +22 / +241 |
| Student core tests | 1,577 | 139 | 1,771 | +526 / +103 / +649 |
| Student GUI tests | 29 | 2 | 34 | 0 / 0 / 0 |
| Student ch05 example tests | 72 | 12 | 87 | +2 / +10 / +12 |
| Student ch06 example tests | 132 | 13 | 147 | +132 / +13 / +147 |

Combined student production is **3,554 active / 877 comments / 4,703 physical**,
versus original **5,243 / 1,641 / 7,606**. Combined student tests are
**1,810 / 166 / 2,039**, versus original **548 / 125 / 728**. Student growth is
1,076 active production lines and 660 active test lines; 465 new production
comment lines are documentation, not 465 lines of new behavior. Eight files have
unchanged Go token sequences; their documentation changes add 35 comment lines
and 37 physical lines. Further inherited comments are mixed into behavior-changing
files, so that is a measured subset, not the entire documentation-only increase.
Expanded anonymous structs also affect physical active-line counts; for example,
the ch05 test's +2 active lines do not add behavior.

The original's entire +200 active test growth is its SSE test file, not coverage
of the new actor lifetime or external role tools. Its standalone delta/SSE source
files add 135 active lines, with additional streaming changes mixed into existing
files. These are disclosed rather than subtracted by guesswork. The student's
growth pays for this chapter's actor, clients, media and relevant tests; no
unjustified framework or arbitrary line quota is indicated. Smaller totals alone
do not establish acceptance, particularly with the completion capability missing.

## Verification and scoped exceptions

Independent formatting, vet and uncached race tests passed in all four modules.
The independent original grader returned **80/100 FAIL**, matching the retained
[student result](../../solutions/edition-2/evidence/ch06/grade-final.txt).
Original frozen-reference baseline was coordinator-observed **100/100 PASS**.

The student's 29 runnable mutation records show intended assertion/liveness
failures; its unused-import experiment is explicitly invalid evidence. I inspected
the retained output for incidental failures and independently reproduced the
shutdown-join mutant: removing the worker-count guard fails
`TestInterruptKeepsAgentAndJobAlive` at “shutdown failed to join blocked Go
handler,” with no race or panic. The new media survivor is recorded above and
must not be relabeled killed. Tests exercise the real Agent, renderer, history,
jobs and subprocess application with small HTTP boundary fakes; no mocks replace
earlier machinery. The externally built workflow validates role tools, stored
drafts, handoffs and results rather than declarations alone.

Retained live artifacts support their bounded claims: Anthropic CLI acknowledges
the hint at 22:45:22.121672Z, interrupts at 22:45:22.172780Z, answers the new turn
at 22:45:23.081200Z, and records the job's real late output at 22:45:34.088930Z.
The initial live artifact's premature `job_finished` with running status is
preserved; the student corrected that guard afterward and does not claim a paid
rerun. Gemini's separate author/editor/reviewer logs contain actual draft tools,
paired results and edited text. OpenAI's saved image experiment answers `red`
(63 input, 4 output tokens) and its audio refusal names the model/category. These
live runs do not prove simultaneous completion grouping or every vendor media
wire shape. No paid calls were repeated during review.

I independently agree with the following narrowly scoped exceptions; these do
not yet make the chapter accepted:

- **Inherited logger-shape check:** actual parent-interface logging satisfies
  Bill's rule. The structural grader demands a direct stored-interface/log-call
  shape and rejects the real Tool → Engine → Agent → Ensemble route. Redundant
  forwarding or an invented parent is unwarranted. Preserve the actual parity
  failure and the prior Chapter 5 evidence.
- **Text-only loud-refusal probe:** inspected `internal/grade/ch06_harness.go`
  `testLoudRefusal`; it submits only `{"user":"test"}` with
  `unknown-model-xyz`. §6.9 defines media capabilities and refusal of unsupported
  media, not a closed text-model allowlist. The implementation correctly refuses
  actual unknown/unsupported media without banning otherwise usable text models.
- **Inherited URI/handle media exception:** §6.11 explicitly defers attachment
  rendering for these locator forms. Metadata and redaction retention remain
  required; local-path attachments are now required and implemented. Keep the
  earlier actual failed media parity results visible in later chapters, and
  reassess when the source actually introduces those attachment forms. Do not
  silently turn inherited failures into earlier-feature requirements.

The incomplete wake-once behavior is **not** an accepted exception. Preserve the
student's corrected explanation and this initial-review history. Reassess the
three bounded findings after correction, running affected checks and targeted
mutations without another paid matrix or broad mutation sweep. Review scratch
copies and grader-generated sidecar logs were removed; production, original
graders, frozen exports and tags were not changed by the reviewer.

## Coordinator follow-up: original check 5 and reference behavior

At Bill's explicit request, the coordinator traced the original solution and
check 5 (`wake-once`) at repository revision `6309f7a`. Bill questioned whether
sub-agent spawning belongs in Chapter 6. The student remains paused; no answer
code or implementation recipe was supplied to it.

The original application creates the author, editor and reviewer directly in
`solutions/ch06/ch06/main.go:39`, then starts their actor loops. Its
`runPipeline` at line 218 performs three blocking calls in sequence: author,
editor, reviewer. The application owns these peer Agents; no model calls a
sub-agent spawning tool. Its `pipelineDone` WaitGroup waits for application
pipeline goroutines at EOF, not a parent Agent's two child completions.

The reference `Framework.WaitAny` at
`solutions/ch06/agent/internal/llm/actor.go:513` returns one matching observation
from the merged channel. It neither gathers two completions nor waits for an
explicit set, and the exercise does not call it. The nearby comment mentioning
"wake-once semantics" overstates the mechanism.

The fifth check, `internal/grade/ch06_checks.go:157`, only requires observations
from at least two Agent IDs and `turn_ended` records from at least two Agent IDs.
It measures no simultaneity, wakeup count, grouped delivery or spawning. Its
mutation test (`ch06_grader_test.go:243`) collapses role registration to one ID;
that tests distinct Agents, not completion coalescing. This confirms the earlier
grader limitation and establishes that the old answer also lacks the behavior
described in §6.6.

Source file revisions: exercise `14961aed8c08da70ba02e2048f73d9d6ff36fd81`;
actor `515e2d883ef0994db652a031eb13e6974ba7b829`;
checks `cd10cd1ea12fa132ed3ef16ddf78591e291f6c1f`.

The coordinator's recommendation is to treat the coalescing paragraph and check
description as a source inconsistency, retaining this chapter's actual sequential
multi-Agent application and deferring agent-spawning semantics. We should have
presented this full reference cross-check before asking Bill to choose a new
completion API. The initial finding is preserved above as review history; it
must not be treated as an approved instruction to expand the exercise while
Bill's scope decision is pending. The two unrelated test/comment corrections
remain valid. No original chapter, grader or reference code was modified.
