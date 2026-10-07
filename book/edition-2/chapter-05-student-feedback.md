# Chapter 5: author response to student feedback

Initial author response, October 7, 2026, before live demonstration or
first-edition comparison. The [student review](../../solutions/edition-2/main/evidence/ch05/student-review.md)
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
the coordinator is preparing separate contract-derived acceptance. Do not
describe that legacy score alone as a failure to implement this chapter or
silently treat it as a passing result. Publish the actual new acceptance command
before the continuation runs it; the original grader remains historical.

## Gemini model scope and explicit hint mapping

Bill selected Gemini 3.8 Flash for hints and limited new Gemini validation to
3.0 Flash and newer. Discovery confirmed `models/gemini-3.8-flash` supports
GenerateContent. The older attempts remain evidence of those exact models;
all four Gemini demonstration modes will use 3.8 Flash for the revised scope.

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
and Ensemble's actor turn. The revised 3.8 live demonstrations remain pending.

## A full terminal input buffer can block the actor

Independent comparison found that `send_input` performed its terminal write
on the actor's dispatch path. A retained local probe filled the input buffer
of a nonreading process and delayed hint acknowledgement until the process
ended. This violates the chapter's responsiveness rule even though HTTP and
report waits already run elsewhere. §5.4 now explicitly includes cancelable
process-input I/O, partial effects, writer settlement and process lifetime;
§5.9 names the corresponding deterministic controls. The coder's correction
and independent re-review are pending. No Gemini renderer defect is inferred
from this separate local finding.
