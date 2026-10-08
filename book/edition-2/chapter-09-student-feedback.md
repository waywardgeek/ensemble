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
