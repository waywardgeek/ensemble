# Enhancement review: Chapter 7 — accepted

> Historical paths: the `edition-2/ch04`–`ch09` paths recorded below use
> original student numbering. Those exports are now preserved under
> `solutions/edition-2/history/`; see the [solution mapping](../../solutions/README.md).

Reviewer: `enhancement_reviewer_resumed`, October 10, 2026. Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`, read entirely for this task. Full
carryover, voice/procedure and current bounded plan are read. Scope is original
Chapter 8 only after enhanced Chapter 6 checkpoint `54939b6`. The preparation
history below predates the initial formal acceptance recorded at the end. No
enhanced Chapter 8 work is released by this review.

## Baseline reading and findings

Read the complete original Chapter 8, complete `ch08.md` and comparative review,
original GUI grader/checks, accepted optional GUI server/command/browser code,
GUI integration tests and actor/dispatch/pause changes. Matching legacy handler
and boundary are inspected; remaining detailed source/evidence comparison is in
progress. Accepted endpoint is `solutions/edition-2/ch08`; legacy is
`solutions/ch08` itself, with `cmd/`, not an imagined `agent/` subdirectory.

Preserve the terminal-to-radio motivation, one ArtifactScroll, safety through
human pacing, serial-tool WHY comment, invisible log-direction failure and actual
speech-error/reload incidents. Length should fit the explanations, examples and
story; no trimming or growth target applies. Correct architecture where original
sketches put gate behavior in common or inject sibling callback services. One
Agent owns pause; a paused boundary cannot stop hints, interrupt or job reports.
An explicit unpause from any tab releases the Agent without another tab's veto.

Original actual student result remains **55/100 FAIL** on the earlier GUI binary,
after roughly 47 minutes of wrong-entrypoint inherited checks. Its missing replay
`from` was a genuine defect, fixed in original correction round 1, not a waived
field requirement. That run does not certify the later corrected binary. Original
legacy result is **100/100 PASS**; neither result is to be inferred from new
grading. Separate GUI/core launch inputs resolve the known architecture conflict
without repeating the timeout cascade or moving GUI code into core.

Accepted code stores Agent-lifetime artifact IDs before publishing finals,
snapshots history under its real lock, serializes history/partial/subscriber
handoff, and admits replay as one batch. Append-only slices still need locking.
Fixed last-hour OR last-100 renderable events is sufficient; no disk restoration
is authorized. Completed replay and live partial replay are different assertions.
Browser projection must preserve empty text and speak only readable thinking,
never signatures/redacted bytes; retain the actual Gemini metadata-card failure
and its correction. Native speech dispatch/start/end/error is real evidence, but
does not establish subjective listening or audible intelligibility.

## Proposed grader assessment

Approved coder's narrow plan before implementation: retain four GUI weights
25/20/15/20 through original predicates plus corrected Chapter 6 parity 20 on
the actual core. Grader-only GUI/model inputs select the separate executable;
default declared streaming model remains a local fixture, not live availability.
Attach failure of the existing external public-dependency check to websocket25,
but continue driving the legacy GUI so its behavioral outcomes remain visible.
No new public API or application acknowledgement protocol is justified.

Use real builtin command results, correlated inputs/call IDs and final content;
undefined `think` errors must not stand in for execution. Completed replay should
retain prior content and artifact identity. Parse every GUI log line's timestamp,
direction and JSON instead of trusting substrings. Retain faithful source tests
for active-partial/handoff/window/paused-interrupt and browser/native-speech proof
with explicit numeric-grader limits rather than duplicating all of them in a
second integration framework. Reader lifetimes must end when a connection closes.

Preparation concern: sending pause or observing its inbound log precedes applying
the gate. A causal shell-release test needs evidence the command was processed,
without inventing a pause acknowledgement. Suggested ordinary WebSocket Ping/Pong
on the same connection after pause: both existing synchronous read handlers apply
pause before reading the Ping and replying. Await coder's selected mechanism and
evidence. Coder owns the harness; reviewer has run no shared GUI grading.

Further source preparation read all new core/API/history changes, optional GUI
CSS/HTML, matching legacy transport/logger/gate/browser files and the actual
legacy command/actor integration. The external root's observer is a notification
consumer, not an invented parent of the Agent. The accepted server keeps socket
writes/logging off the observer, tracks connection lifetimes and joins writers;
the original copied broadcaster references can reach a closed send channel.

Read complete retained initial/corrected mutation outputs, browser/speech/
projection launchers, summaries, and original replay-field/window controls.
The initial cleanup hang, final-part cleanup survivor, late dispatch assertion
and truncation timeout remain separate from their corrected causal failures.
Browser wrappers forward to the native speech service; no mock supplies events.
The projection baseline passes readable positive controls and all three empty/
opaque negative cases; restored fallback fails display, automatic and speaker
input for each negative while both readable controls pass. Its initial later
control timeout is preserved.

Inspected the retained screenshot and actual three-provider wire/visible captures.
Independently counted 28/27/5 deltas, correlated all three successful command
results and reassembled each final text including Gemini's legitimate empty
artifact. The raw Gemini visible capture indeed contains the pre-fix metadata
card. Anthropic's terminal hint is a browser human message before tool finish.
These earlier live captures also retain the missing replay `from`; the final
source fix uses separate focused evidence and is not retroactively inserted.
No browser or provider rerun has been performed during preparation.

Read initial new grader files and CLI/helper diff completely. Its actual-command
dependency target is `./cmd` when that is what `grade.Build` invokes; earlier
consumer checks retain their original `.` behavior. Same-socket Ping/Pong is
implemented for pause processing before shell release. No preparation blocker
was found. The coder's initial actual full results are reported as accepted
100/100 and legacy 20/100, pending final report verification. Legacy transport
adds blank separators because WriteJSON and its logger both add a newline;
agreed to ignore whitespace-only separators while strictly parsing every actual
message record, preserving the initial score rather than relabeling it. Known
delta/final-kind validation is within the existing correct-fields promise.
Coder still owns the harness; these are preparation findings, not acceptance.

Read the entire 700-line enhanced manuscript and author dispositions. Initial
prose SHA-256 is `067f449d07535d60d952059fadb0c54e1c17f07a92504e55f22fdc5d27c4c180`.
The retained radio/ArtifactScroll/pacing explanations still motivate the work;
the observed empty Gemini card and speech interruption teach concrete failures
without invented dialogue or subjective audibility. Added ownership, replay
handoff and command contracts earn their space. There is no length target or
request to cut for a soft lint warning. Independent chapter lint passes hard
checks at 4,396 prose words, with soft negation/closer/person-gap warnings.
No prose blocker was found; exact sample execution and final coder evidence
remain pending. The parent's reader-versus-writer trace and network-wait wording
corrections are already present in the submitted draft.

## Initial formal combined review — accepted

Accepted October 10, 2026 at initial formal review, **zero correction rounds**.
Full final coder submission and reconciled author dispositions are read. The
manuscript remains SHA-256
`067f449d07535d60d952059fadb0c54e1c17f07a92504e55f22fdc5d27c4c180`.
The previous preparation concerns were resolved before formal submission; their
history remains above. No runtime, original grader or frozen source was edited
by this enhancement. The coordinator owns the checkpoint and next release.

### Prose and teaching

The chapter enhances the original argument and sequence. The radio analogy,
single ArtifactScroll, serial-tool pacing, missing log direction and Chrome
speech-error mechanism remain explanations with consequences. The actual empty
Gemini artifact and reloaded-tab speech failure make the additional contracts
understandable. Required owner access and separate public GUI are taught before
the student encounters the failure they prevent. Pause is explicitly Agent-wide;
another tab's unchanged typing or speech cannot veto an explicit unpause.

The new text distinguishes reliable request completion from display progress,
returned waits from finished jobs, Agent lifetime from turn interruption, parser
identity from persistent display identity, and completed replay from active
partial handoff. It does not promise disk restoration or subjective audibility.
The fixed OR window counts renderable events, not multipart cards. The inherited
opening, useful examples and observed incidents earn the resulting length; soft
lint warnings are not a reason to remove teaching or personality. No unresolved
voice, scope, factual or instruction mismatch remains.

### Code, coverage and comparison

Read all final new grader source, its command/helper changes, original checks
and harness, accepted GUI/browser/core changes and matching legacy integration.
The four valid original predicates/weights remain in `Ch8Evaluate`. The new
launcher supplies real evidence to those predicates, instead of rerunning a GUI
server as a finite CLI through the known recursive cascade. Public dependency
inspection follows the executable actually built, while earlier consumer calls
retain their existing root-package behavior. Legacy behavior remains observable
when its same-module architecture fails the separate-module requirement.

The new driver uses actual core, shell jobs, HTTP and WebSockets with only a
faithful provider fake. Pause processing precedes shell release through the
ordinary same-socket Ping/Pong barrier, justified by both synchronous command
readers. It checks actual correlated nonerror results in both GUI and the next
provider request. There is no invented acknowledgement or model-table API.
Completed replay compares identity and content, not just labels. Strict parsing
of each nonempty trace record preserves the valid legacy blank separators.

The stronger active-partial, atomic handoff, fixed-window, paused-interrupt,
browser-projection and native-speech claims remain source/integration obligations.
I accept the existing causal tests and retained mutation/live evidence described
above for those unchanged features. A numeric score is not being substituted for
that proof. Existing flattened opaque flags likewise do not prove provider-byte
preservation. Ownership, common-package behavior and absence of mutable global
application state remain source review, not inference from an emitted log line.

The comparative runtime measurements remain those of
[the accepted original Chapter 8 review](ch08-code-review.md), since no runtime
changed here. Against each edition's own predecessor, accepted core added
131 active/33 comment-only/179 physical Go lines and optional GUI production
397/66/476; legacy core added 781/245/1,131, including integrated transport.
Accepted core tests added 35/4/41 and GUI tests 294/14/314; legacy tests added
221/35/284. Accepted new browser JS is 320/38/361 versus legacy 247/12/299;
CSS and HTML are separately counted there. Previously missing legacy examples
are not credited as deleted production code. Smaller core growth reflects the
existing public ownership seam, not permission to discard the required GUI.

Independently recounted new grader production: **609 active/37 comment-only/665
physical** lines; its integration test is **16/2/20**. Original checks/harness
remain **687/77/867** and are not claimed as deletions. The method excludes
blank and comment-only lines from active counts; mixed lines count as active.
The command/helper diff adds 23 and removes 10 physical lines. This is a separate
versioned driver and content validation, not a new generic transport framework.

New grader comment share is approximately 6%, below the roughly 20% teaching
target. I explicitly accept it here: the comments explain the non-obvious
ownership/lifetime, ordered pause barrier, replay boundary, cleanup and actual
provider-result checks. Most remaining lines are readable literal fixtures and
direct field validation; padding each branch with narration would not teach more.
The exported Chapter7 and integration test have useful documentation; no new
exported fields or interface methods lack comments. Existing runtime comments
explain the harder ownership and synchronization decisions (historical core
19.5%, GUI 14.5%). No unresolved bloat or comment-quality finding remains.

### Evidence and independent validation

Verified all five final submitted grader-file hashes against
`validation-overlay.json` and both shared and isolated files. The isolated
checkout is based on `54939b6`; final manuscript identity also matches. Independently
ran `go test ./internal/gradeedition2 ./cmd/grade-edition2 -count=1`: PASS,
7.692 seconds, with the command package correctly reporting no tests. Affected
`go vet` passes. The coder's accepted GUI source race run passes in 2.845 seconds;
unchanged source behavior did not justify another broad or paid sweep.

Read all 19 final mutation patches and their actual check results. Each reaches
the intended runtime behavior and loses its intended check; the initial invalid
JSON mutant's compile failure remains setup failure, not a kill. The command-only
private-import and wrong-core controls actually build and pass wire behavior
before the dependency predicate rejects them. The ignored unpause timeout is
meaningful against its working real-gate baseline. Earlier historical survivors
and their corrected causal assertions are preserved, not relabeled.

An independent scratch mutation removes `from` only from state snapshots after
completed response history exists. With the real mutated optional GUI and core,
the revised grader gives websocket 25/25, replay **0/20** with
`replay state lacks from/to`, log 15/15 and pause 20/20. The focused overlay test
passes its intended-failure assertion in 1.243 seconds. This confirms that the
genuine original replay-status defect is still detected without damaging live
state delivery. No shipping test or source was changed for this control.

Independently parsed all twelve published JSON envelopes and inspected the actual
sample receipts: `Sample GUI reply.`, a distinct `turn interrupted` completion,
then `Still usable.`. Port 8088, the actual module build, explicit assets and six
client commands are recorded. The initial scratch assertion confused interruption
with the next request's completion; the corrected correlation and retained failed
output support the coder's explanation. No runtime or prose repair was invented.

Verified initial full reports: accepted **100/100 PASS**, legacy **20/100 FAIL**.
Verified actual final GUI components: **80/80** and **35/80**. Final **composed**
reports are **100/100 PASS** and **35/100 FAIL**, carrying forward unchanged
full-run parity **20/20** and **0/20**. This is not another full end-to-end run.
Legacy genuinely passes logging and pause; reused final IDs, missing replay IDs
and its integrated boundary fail. Its absent separate exercise applications are
setup limitations, not proof those tools executed unsuccessfully. Unknown-model
and other actual inherited failures remain separately visible.

Historical original **55/100 FAIL** belongs to the earlier accepted-source GUI
binary; original legacy **100/100 PASS** remains the recorded historical run.
Neither is overwritten by revised results. The three earlier live provider
captures and native speech events retain their exact source/proof limits. No new
paid or subjective listening result is claimed. Raw enhancement evidence remains
at the coder's path linked in [the submission](enhancement-07-code.md); reviewer
scratch is removed after recording this result.
