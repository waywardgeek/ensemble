# Chapter 5: author response to student feedback

Author responses, October 7, 2026. The initial response preceded live
demonstration and first-edition comparison; dated revisions below preserve
that sequence. The [student review](../../solutions/edition-2/ch05/evidence/ch05/student-review.md)
retains its initial reads, plan, failures and requested clarification. Current
validation status belongs in [the gate record](chapter-05-validation.md).
The prior story pass has independent acceptance at review commit `3d6effd`;
the resumed student has now confirmed the timing clarification below.

The student found the prompt/hint/receipt/wire distinctions, explicit collection
readiness barrier, prepared report cursor and lock-inversion warning useful.
Retain those explanations. Fixed event indexes, synchronous observer assumptions,
omitted event payload recognition and the test-server cancellation hang are
recorded implementation or fixture migration issues; they do not justify
relaxing the actor contract. The student reports local tests in six modules
and the core race suite passing; independent acceptance and live use remain
separate gates.

## Failed CLI request and automatic queue progress

The student identified a genuine timing ambiguity in §5.7. A failed request
completes on the actor, but the CLI observes that completion on another
goroutine. The actor can activate the next queued prompt before the client
submits Close. The earlier shorthand could be read as a guarantee that every
request queued at the instant of failure must be stopped without activation.

Accepted clarification, consistent with the coordinator's interpretation of
§5.2: requests still queued when Close is admitted receive stopped completions.
A request activated earlier follows the active-turn close rules; accepted
effects and completions are retained. §5.7 now says this explicitly and does
not add an actor halt-on-provider-error policy. The affected stronger guarantee
was held while independent work continued. The coordinator subsequently checked
the student's “Resumed clarification confirmation”: it accepts the wording and
records a deterministic test with the second request active, its accepted file
effect retained, and a third request still queued when Close is admitted. This
teaching finding is resolved; chapter acceptance remains a separate gate.

## Live plan and inherited grader

The author read the complete pre-run live plan. It covers all-three-API human
PTY input, queued prompts, hints and observed delivery, interruption, later
job inspection, explicit quit, orderly queued EOF, actual request replay,
the public three-Agent workflow and reliable collection. Timing-sensitive
HTTP cancellation, report supersession, overflow, nonkillable work and sequence
assignment remain explicitly deterministic controls. No missing teaching scope
was identified, and no live success is claimed from this plan.

The inherited Chapter 6 black-box grader's 10/100 result is preserved. Its
old observer/pipeline and unknown-model expectations are not the new contract;
the coordinator supplied separate contract-derived acceptance, now published
in the exercise. The new checker initially scored 90 for missing taught
model/MIME/mapping diagnostics; correcting those implementation omissions
produced 100. Neither result changes the old diagnostic into the new contract,
and the original grader remains unchanged.

## Gemini model scope and explicit hint mapping

Bill selected Gemini 3.8 Flash for hints and limited new Gemini validation to
3.0 Flash and newer. Discovery confirmed `models/gemini-3.8-flash` supports
GenerateContent. The older attempts remain evidence of those exact models;
all four revised Gemini demonstration modes now have actual 3.8 Flash receipts.

The original §5.3 prescribed literal hints after results, but required the
student to combine that instruction with Chapter 2's provider mapping. The
new subsection spells out the GenerateContent request suffix: original model
call/signature, user function responses, then separate literal hint text parts;
the neutral `hints` sequence array stays in the event log. It also separates
receipt, wire delivery, consumption and observed model compliance. Chapter 2
now links that example. This fills a teaching gap without claiming that the
earlier ambiguous diagnostic established a renderer defect. The student
confirmed that the renderer and consumption path match the expanded explanation;
the independent reviewer checked its fields, placement and signed-part rule
against the renderer, frozen request and GenerateContent signature guide.
The reviewer also requested the now-added distinction between a provider turn
and Ensemble's actor turn. The four revised 3.8 demonstrations passed at runtime `959c663`. In the human
control run, hint sequence 7 appears in request sequence 10 after the matching
function response and is absent from the next request. The model includes the
requested marker. Receipt, delivery, consumption and compliance are separately
observed; this does not establish why the older models failed.

## A full terminal input buffer can block the actor

Independent comparison found that `send_input` performed its terminal write
on the actor's dispatch path. A retained local probe filled the input buffer
of a nonreading process and delayed hint acknowledgement until the process
ended. This violates the chapter's responsiveness rule even though HTTP and
report waits already run elsewhere. §5.4 now explicitly includes cancelable
process-input I/O, partial effects, writer settlement and process lifetime;
§5.9 names the corresponding deterministic controls. The coder implemented the correction at `959c663`; independent review accepted
it at `743dca3`. Local controls establish full-buffer responsiveness, truthful
partial writes, cancellation settlement and preservation of the process after
interrupt. New real-model PTYs exercise ordinary input/echo/kill on all three
APIs, including scoped Messages and Chat Completions reruns. Ordinary 16-byte
writes do not substitute for the deterministic saturated-buffer check. The
student confirms the teaching and repair in its appended review. No Gemini
renderer defect is inferred from this separate local finding.


## Subscription lifetime after shutdown

Independent review also found that exited subscriptions retained callback and
queue references, and admission could succeed after Ensemble close. §5.6 now
states closed admission, cleanup by the exiting owned worker, and retained
terminal reason. It does not promise forced termination of a user callback.
The student confirmed the clarification and revised the implementation;
`959c663` and independent review `743dca3` resolve the finding. Deterministic
lifetime controls establish this behavior; the GUI remains a stub.

## Actual use and evidence correction

The student's appended experience distinguishes a hint rejected after the turn
ended from a hint accepted during work. §5.10 now teaches waiting for observed
work before entering the hint, and inspecting receipt, actual request and next
request separately. The initial Messages prompt's invented `timeout_ms` field
and older Gemini responses missing required usage are preserved as failures,
without weakening validation or blaming an unproven renderer defect.

The final demonstration map retains eight applicable initial-source runs and
six revised-source runs: 14 logical runs, 78 captured requests. Source bindings
and exact model identities remain distinct. §5.10 supplies the human actions,
abridged transcript, public workflow and collection commands; the full ledger
is linked rather than repeated in the reader's path. The student review
confirms the hints/close and R1/R2 dispositions. No unresolved teaching question
is reported there.

Independent review found that all 13 original negative identity controls hit
an earlier path guard, masking the checks their names advertised. Preserve
that finding as an evidence-test defect. The independent supplemental audit
uses a passing complete fixture and 13 isolated mutations that reach the
intended identity refusal before replay or derived writes; it also replays all
78 captured requests and preserves all 247 raw files. This correction changes
neither runtime nor paid receipts. Final proofreading and coordinator checkpoint
remain separate in the gate record.
