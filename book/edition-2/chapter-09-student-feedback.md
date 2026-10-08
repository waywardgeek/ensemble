# Chapter 9 student teaching feedback

## Initial plan questions, October 8, 2026

The fresh student's initial review is
`solutions/edition-2/main/evidence/ch09/student-review.md`, based on teaching
`8736c951393320cf8beef849806bde3b90bc445f` and accepted Chapter 8 source
`446d7f27fcef052d6bb780901c786842dc5286c1`. The student reported these questions
before implementation, not after a failing grader. The original questions and
source-read account remain unchanged. The coordinator reviewed the owner/API plan
and supplied the working directions below; these are not new Bill rulings.

The responding author previously implemented Chapter 4 and authored later
teaching. The author has not implemented Chapter 9. Source inspection for this
response read the complete initial student review, affected Chapter 5/9/10
contracts, the Chapter 6 opener, and accepted predecessor CLI configuration and
renderer portions to confirm the reported seams. No runtime or grader was edited.

| Question | Teaching disposition |
|---|---|
| Q1, numbered question 1: system override environment name | §9.8 explicitly adds LLM_SYSTEM through the shared reader with LookupEnv and no provider fallback. Skill mode rejects a supplied nonempty value before defaults; no-skills uses existing Config normalization. Public configuration semantics remain intact. |
| Q2, numbered question 2: primary recorded once versus request captures | §9.7 keeps the existing captured system field as the empty string in skill mode and resolves the primary from its immutable recorded activation for rendering/reconstruction. No-skills captures are unchanged. §9.8 distinguishes Config's effective primary view and accepted round trip from captured request configuration. |
| Q3, numbered question 3: H/S/P placement after a held batch | §9.7 defines skill-mode-only stable post-batch anchors: pending unanchored hints join when a batch becomes unresolved, and later hints join while it remains unresolved. Existing anchors remain fixed. It merges anchored hints/manuals by durable sequence and gives literal suffixes for all three renderers. After the turn ends and prompt P arrives, the suffix is results→H→S→P; consumption removes H while leaving S before P. Non-anchored hints retain request-tail placement. No global sorting or moving manuals is allowed. |
| Minor closing-reference finding | Verified Chapter 6 teaches streaming. Chapter 5's closing promise now leads into visible proposed work and the complete-response execution boundary; its old skills-next promise is removed. |

Chapter 10's required semantic state and settled-boundary passages also now
explicitly retain completed-batch hint anchors. An unconsumed hint is not an
unfinished tool batch; it can survive checkpoint/resume without moving a manual.
No new event format or snapshot authority is introduced by that clarification.

Scoped prose lint passes every hard rule; JSON fixture parsing and scoped
git diff --check pass. The author read the changed passages and cut repeated
wording. Remaining soft warnings concern existing chapter density/person gaps
and the short feedback record; no runtime validation is inferred. Student
confirmation is still pending publication.
This response does not claim an implementation, paid run, grader acceptance or
resolved student confirmation before the student reads it.

## Student confirmation after publication

The author read the appended `Implementation phase: release-8f24360` section in
the student's working-tree review on October 8, 2026. The coordinator preserved
the original plan/questions at 5ac45e4; this later confirmation was not yet a
separate frozen checkpoint when read. The student explicitly confirms that all
three published responses resolve the questions and that the three-provider
fixtures clarify the distinction between request capture, public configuration
and chronological dialogue. The student reports proceeding with local integration,
with the pre-provider review gate intact. This confirmation is teaching feedback,
not an acceptance claim for the implementation or later demonstrations.

## Encoded transition bound, October 8, 2026

The coordinator reported an integration difficulty: accepting the required
8 MiB decoded material case can require roughly 48 MiB after JSON escaping,
but bypassing an ordinary raw-record check for every skill record leaves reads
unbounded. Repeated offers and retired summaries also prevent the decoded body
limit from bounding a complete transition. This is a teaching omission; no
passing implementation or test result is inferred from the report.

Section 9.5 now defines an independent 67,108,864-byte complete skill-record cap,
including a framing LF when present, with incremental bounded reading and
pre-append candidate sizing. Oversize is a controlled skill_too_large refusal
with no transition mutation; an enclosing management attempt retains its normal
call/result and limit-consumption facts. Ordinary/header limits and final
EOF framing acceptance stay unchanged. Preserve the escaped-material positive,
exact decoded boundaries and raw exact/+1 controls. This is a coordinator working
resource decision, not a claimed bound implied by the original prose.

The author changed only teaching and this response. The student must read the
published clarification before the affected repair and confirm whether it resolves
the difficulty; implementation/checker acceptance remains separate.


## Initial implementation feedback and import repair

The author read the student's appended review through `support-review-accepted`
on October 8, 2026. The initial source remains frozen at `654075b`; the repaired
runtime/support is `c0e3171`. The student confirms that `d8c7738` resolves the
whole-record question and distinguishes that missing teaching from its own
unbounded-reader implementation. The subsequent import defect was also an
implementation mistake under the clarified contract: measuring a re-encoded
copy can reject an original physical record that fits. Section 9.5 already
requires original-byte accounting on read and emitted-byte accounting on write;
no limit or semantic-validation rule is relaxed.

The student reports a regression covering the original encoding with and without
final LF, while retaining write-side refusal. Its initial positive fixture first
failed because it omitted the required timestamp; that failed attempt remains in
the review. The initial 67-row gate passed before this defect was covered, and
its result does not establish the later repair. Module/support checks and the
coordinator's affected checks retain their own source identities. Actual use
and comparative review remain separate gates.

The student classifies the concrete parent-pointer correction, UTF-8 validation
order, fixture corrections and evidence-verifier ordering as implementation or
support mistakes under existing teaching. The author agrees; they do not require
new permissions or weaker chapter rules. Independently corrected checker fixtures
remain recorded as checker defects. No unresolved teaching question is reported
at this boundary.

The coordinator also found a reader-path typo: §9.9 invoked `/tmp/ensemble-cli`
after the TL;DR built `/tmp/ensemble-ch09-cli`. The invocation now matches the
build. This is a prose correction and does not relabel any actual executable or
claim that the planned spin has occurred. Later live and completion feedback
will be appended after the student's receipts are available.


## Live empty-argument stream and confirmation

The student read and confirmed `f4459a5` during its live-phase resume: the
whole-record omission, import implementation bug and executable-path correction
are accurately distinguished. The original failures remain in its review.

The first Anthropic GUI stream exposed an inherited Messages assembly error:
start input `{}` followed by partial_json `""` was treated as an empty replacement
and rejected despite a complete stream. The author inspected the actual
responses/002.body receipt and the student's repair account. Chapter 6 §6.4 now
states zero concatenated argument bytes, even with empty delta events, and gives
literal positive/negative fixtures. Whitespace or malformed nonempty replacement
still refuses, and execution still waits for the complete accepted response.
This clarifies the existing start-object fallback; it does not change tool
rights or accept an incomplete stream. The student classifies the assembler's
event-count interpretation as an implementation defect under that teaching.

The original GUI refusal/browser timeout remains bound to `c0e3171`. The
student's later explicit-path recovery, native audio and full-card text-delivery
observations are separate receipts awaiting their complete author reconciliation.
Its Anthropic public run retains exit 1 and a missing final beta report, even
though the coordinator independently accepted the required Skills feature
coverage from actual calls, bindings and isolation. Neither that partial task
nor the empty-argument repair is rewritten as an originally successful run.

## Interim live continuation, October 8, 2026

This response covers the student's review through the `bounded-public-recovery`
response naming support checkpoint `3da764b`. The live matrix is still active;
this is an author disposition, not its final freeze or acceptance. The original
attempts, corrections and source bindings remain separate. Section 9.9's final
demonstration will be reconciled from the completed immutable matrix.

| Experience | Author disposition |
|---|---|
| Published empty-argument clarification | The student explicitly confirms that `305b1b0` resolves the distinction between receiving a delta event and receiving argument bytes. The author agrees with the narrow repair and the retained whitespace/malformed/incomplete refusals. The later real Messages supplement reports the same empty-object/empty-delta path executing successfully under the repaired binding; it does not turn the original GUI failure into a pass. Final receipt reconciliation remains pending. |
| Relay HTTP 502 in OpenAI P and Gemini N | The terminals establish a failed call. The student's relay diagnosis establishes a local transport-exception response with no retained upstream response, but the original relay omitted exception class. That is a support diagnostic limitation: the evidence cannot identify timeout, network failure or an upstream rejection. No runtime or usage rule needs weakening. The separately frozen recovery support adds safe diagnostics and reviewed caps while preserving the original support and shared budget accounting. |
| OpenAI public coverage after its first failed request | Constructing two typed Agents does not prove either model read its file. The missing reads remain outstanding at this boundary. Recovery requires the coordinator's bounded plan and a new binding; the author does not convert unspent calls in another scenario into permission or count a planned recovery as an observation. |
| Anthropic public beta run ended at round_limit | Keep exit 1 and the absent final beta report. Actual file reads, distinct literal bindings and isolated skill state can establish the required public features, as the coordinator's independent audit found, while the requested task still ends partially. A model's redundant load followed by unload spent its remaining request budget; this is useful human experience, not evidence that isolation failed. |
| Gemini empty STOP with missing candidate count | The inspected response contains empty text and STOP, promptTokenCount 1064 and totalTokenCount 1064, but omits candidatesTokenCount. Chapter 2 requires that base count and Chapter 6 retains the final usage snapshot. Refusal follows the published contract. Do not infer a missing count from subtraction, label the omission a provider bug, or claim a successful answer from terminal metadata alone. Any compatibility change needs an explicit grouped contract and distinguishing tests before code changes. |
| Browser wait after a terminal-origin OpenAI request | The student waited for a local-composer status belonging to a different input path. Its report identifies the actual terminal success and browser outcome/card/state separately. This is an evidence-driver selector mistake; retain the timeout and verify the appropriate observations without repeating an already completed model prompt. |
| Click expansion where the reviewed live plan required a keyboard action | A mouse click and a local keyboard fixture prove different things. Keep the live keyboard capture gap until the separately bound, zero-model-call keyboard action is observed and reviewed. The frozen helper alone does not close it. |
| Native speech and audio capture | Full text delivered to the adapter, native callbacks and captured sound are separate observations. Quiet lead-in duration and later waveform energy can qualify a recording, but cannot establish intelligibility or model hearing. Preserve the student's unsupported-audio-input limitation and both original and corrective capture scopes. |

The current primary usage reference describes the counters but does not give an
explicit rule in its UsageMetadata section for interpreting this omitted field.
The author recommends preserving the taught refusal for this attempt and keeping
zero-output compatibility as a separately reviewed question if stronger evidence
or a deliberate policy change warrants it. Neither a missing field nor this
conservative client contract alone establishes a provider defect.

For eventual spin prose, preserve the request the reader cared about: beta was
asked to read and report, but spent its last opportunity changing an already
usable skill. Authority and data isolation can work while the model leaves the
task unfinished. The final account must show both, with the actual file/tool
receipts and the bounded partial outcome. No all-provider completion, successful
recovery, or student confirmation of this new response is claimed here.

## Frozen initial live matrix and author reconciliation

Student freeze `786ff239ddf337a5ebb39a12f6552d047f52c1ab` preserves the initial
source/live/teaching experience before historical comparison. The student read
the complete direct response `5063f35`, checked its supplied manifest and confirms
the classifications above. Independent live review `2dc5841` accepts the stated
feature evidence with explicit limits. Neither checkpoint is final chapter
acceptance. Section 9.9 now describes the actual demonstrations and reproduction
steps; it no longer presents the initial plan as an unperformed exercise.

| Later finding or suggestion | Disposition in the reconciled account |
|---|---|
| OpenAI public recovery | Both real reads completed in four new calls under the reviewed transfer of one unused F allowance to P. The first transport failure remains a spent attempt. Beta's shortened final marker is recorded separately from its exact literal captured binding; no full prose-compliance claim is made. |
| Gemini remaining management actions | The final bounded run established unchanged review and absent-name refusal. It does not supply the missing answer to the earlier revoked-write prompt. Section 9.9 distinguishes offered-schema removal and unchanged file bytes from model refusal and local forced-admission enforcement. |
| Keyboard linkage | The actual focused Enter expansion is retained, with the source/binding, browser identity, URL, within-run time and material linkage accepted by independent review. The sampled launch bytes were not retained. Neither the prose nor feedback claims an exact historical launch-byte comparison or derives a minimal finalization diff. This is an evidence-support omission; future capture should retain the sampled bytes alongside their hash. |
| Empty-object stream fixture | The explicit Chapter 6 examples at `305b1b0` resolve the student's suggestion. Section 9.9 distinguishes the original runtime defect, the old-binary explicit-path recovery and the real repaired empty-object demonstration. The latter was independently checked; no new acceptance rule is introduced here. |
| Safe transport diagnostics before live use | Accept the support improvement: capture safe stage/class information without exception text, URLs, headers or credentials. The separate recovery support added bounded diagnostics and local controls. The missing original exception details cannot be reconstructed; the chapter leaves the original cause unknown. This is evidence practice, not a new Skills runtime obligation. |
| Authority versus task completion | Preserve beta's unnecessary load/unload and round-limit exit, even though isolated capabilities, material and reads were demonstrated. Section 9.9 makes the reader's missing report concrete instead of calling every scenario successful. |
| Exact replay and sound | All 93 attempted request bodies are accounted for, while only accepted durable responses contribute normalized usage. Four bounded native captures and full-text submission remain separate from hearing or intelligibility. The two distinct browser screenshots have accessible descriptions and viewport limits. |

The student reports that the owner/API/material separation and literal provider
fixtures were useful in implementation and real use. The author retains them.
No unresolved architecture question is reported in this frozen student phase.
Independent historical comparison may still produce teaching revisions; its
findings will be appended rather than rewriting this initial experience. Student
confirmation of this final reconciliation is requested after publication.
