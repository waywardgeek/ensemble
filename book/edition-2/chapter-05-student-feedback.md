# Chapter 5: author response to student feedback

Initial author response, October 7, 2026, before live demonstration or
first-edition comparison. The [student review](../../solutions/edition-2/main/evidence/ch05/student-review.md)
retains its initial reads, plan, failures and requested clarification. Current
validation status belongs in [the gate record](chapter-05-validation.md).
The prior story pass has independent acceptance at review commit `3d6effd`;
the new timing clarification below awaits the resumed student's confirmation.

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
was held while independent work continued. Student confirmation of this wording
will be recorded when the continuation resumes.

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
