# Chapter 8 comparative code review

> Historical paths: the `edition-2/ch04`–`ch09` paths recorded below use
> original student numbering. Those exports are now preserved under
> `solutions/edition-2/history/`; see the [solution mapping](../../solutions/README.md).

Reviewer: `review_restart_ch01`, October 9, 2026. **Accepted with documented
exceptions after one correction round (maximum three).** All three findings are
resolved. The original GUI-directory result is **55/100 FAIL** on the earlier
binary, not a final-source pass. The final disposition below distinguishes its
corrected implementation defect from the inherited entry-point mismatch. Initial
review and pending-result stages remain preserved as history.

Read the full `.agents/skills/ensemble-coding/SKILL.md`, carryover, workflow,
original Chapter 8 and the [student review](ch08.md). Chapter SHA-256:
`207ae885a7a9d6210757b5f3b4d9773a98206a66e62710e5ce0be0a6ce180f56`.
Compared current `solutions/edition-2/main` against accepted `edition-2/ch07`
at `7c940fc`, and original `solutions/ch08` against `solutions/ch07/agent`
and the latter chapter's separately stored examples.

## Initial required corrections (resolved in round 1)

1. **Project readable part content explicitly instead of falling back to JSON
   metadata.** `gui/web/gui/app.js:partText` uses truthiness to choose text, a
   stub, or serialized part data. The retained `gemini-event.log` contains a
   legitimate empty text part, artifact 2. Evaluating the unchanged function on
   that actual part produces formatted `artifact_id`, `type` and `from` JSON.
   The corresponding `gemini-live.json` **visible browser capture already shows
   that metadata card**. A supported `redacted_thinking` part similarly displays
   its opaque replay bytes via the opaque fallback. This is not human-readable
   model content and contradicts the adjacent comment. Card text also feeds
   automatic and speaker-button speech, so the same boundary affects TTS.
   Preserve empty text as empty and render/speak only legitimate readable text
   for thinking; retain opaque material internally for replay. Add a focused
   browser regression and meaningful restored-defect mutation, covering display
   and speech input without replacing native speech or adding a renderer framework.
   Reuse the existing live artifact; another paid vendor run is unnecessary.

2. **Document the new exported test method `(*trace).Write`** in `gui/gui_test.go`.
   Its synchronized transport-log writer contract deserves a brief declaration
   comment, as required for every exported Go symbol. The embedded Buffer's
   field comment does not document this method. No broad comment rewrite is needed.

3. **Keep replay status consistent with the public `state_changed` envelope.**
   `Server.replay` emits this type with `to` but no `from`; the retained live
   captures show that same incomplete initial status. Chapter 8's public wire
   example defines both fields, and its check table requires correct fields.
   Reusing the transition type for a snapshot does not by itself justify omitting
   its required field. Preserve truthful current-state/replay semantics and the
   established envelope, without inventing a historical transition or a new
   protocol framework. Add a focused assertion for a fresh subscription's status
   fields and its corresponding defect check. This is a small current-contract
   correction, not an accepted grader-shape exception. The coordinator correctly
   challenged my preliminary suggestion that this might warrant a waiver.

## Design, ownership and scope

One application constructs one Ensemble and one Agent, and both terminal and
browser inputs enter that Agent's existing actor. The optional GUI module imports
only the public core API; core has no GUI/WebSocket dependency. The root owns
observer registration, the Agent owns the pause gate, and Engine/history/tools
keep their existing actual-parent relationships. Transport receives Agent/root
services rather than becoming a fictional parent of the Agent. Common gains data
and interfaces; pause behavior stays with Agent. No mocks or mutable application
globals were introduced.

The pause waiter is outside the mailbox consumer, so hints, cancellation and real
job results can still arrive. Serial dispatch checks the shared gate before each
new tool. Explicit unpause from either tab releases that same gate; browser-local
cause transitions do not create per-tab vetoes. Interrupt leaves the Agent and
other jobs alive. No workbench, settings subsystem, sub-agent orchestration or
disk restoration was pulled forward.

Parser IDs remain local to their response, while the actor assigns lifetime
artifact IDs and stores them with final parts. Browser deltas and replay finals
therefore identify the same cards. The stream/plain comparison removes only this
new display metadata; semantic parts, usage and replayed request comparisons
remain checked. Removing IDs from that comparison is justified by different
arrival order, not permission to hide provider-content differences.

History suffix snapshots deep-copy under its real lock. The GUI synchronizes
history capture, partial state and subscriber admission, then queues replay as one
batch so a window larger than the live buffer cannot be silently truncated.
Final facts remove partials before publication. Socket writers and logging stay
outside the observer lock. Teardown uses separate lifetime signals, keeps data
channels open and joins writers before closing logs. Slow live consumers may
drop progress as the chapter permits; reconnect recovers the authoritative view.
This is a small current-feature design, not an unnecessary transport framework.

The original solution embeds WebSocket transport in core, passes gate/send/log
dependencies into the hub, and assumes append-only history permits unlocked
slice access. Its subscriber replay has separate history/partial/live stages,
bounded per-message admission and a send-channel closure reachable by copied
broadcaster references. The student addresses those concrete synchronization and
ownership weaknesses. Original browser code retains unsafe tool-name HTML,
truncation without full-content access, response-ID/replay mismatches and incomplete
speech/pause controls. The student's single artifact component and DOM construction
are materially stronger, subject to the readable-content finding above.

Comments explain why pause waits outside the actor, IDs differ from parser IDs,
replay admission is atomic and writer lifetimes must be joined. Core production
comments are 19.5% of nonblank lines; new GUI Go is 14.2% and JavaScript 10.1%.
Those lower local ratios are not grounds for padding: the important ordering and
lifetime choices are explained. The missing exported method comment remains a
separate concrete requirement.

## Counts and chapter growth

Count each physical source line as blank, comment-only, or active, including
imports/braces/literal data. Comment-only `//`, block comments and HTML comments
are counted; mixed code/comment lines remain active. Split original inline
JavaScript out of HTML rather than hiding browser behavior in a markup total.
Exclude logs, binaries, module metadata, README prose and evidence from production.
No generated source was introduced. Values below are **active / comments / physical**.

| Category | Chapter 8 | Delta from its own Chapter 7 |
| --- | ---: | ---: |
| Original core Go production, including embedded transport | 5,664 / 1,840 / 8,284 | +781 / +245 / +1,131 |
| Original Go tests | 769 / 160 / 1,012 | +221 / +35 / +284 |
| Original browser JavaScript, including inline script | 247 / 12 / 299 | all new |
| Original CSS | 116 / 2 / 138 | all new |
| Original HTML excluding inline script | 28 / 0 / 29 | all new |
| Student core Go production | 4,050 / 979 / 5,340 | +131 / +33 / +179 |
| Student optional GUI Go production, including command | 404 / 67 / 486 | +396 / +64 / +473 |
| Student ch05 example production | 74 / 13 / 92 | unchanged |
| Student ch06 example production | 267 / 20 / 295 | unchanged |
| Student ch07 example production | 74 / 4 / 82 | unchanged |
| Student core Go tests | 2,149 / 175 / 2,396 | +35 / +4 / +41 |
| Student GUI Go tests | 311 / 13 / 332 | +282 / +11 / +298 |
| Student ch05/ch06 example tests | 204 / 25 / 234 | unchanged |
| Student browser JavaScript | 322 / 36 / 361 | all new |
| Student CSS | 139 / 0 / 139 | all new |
| Student HTML | 29 / 0 / 29 | all new |

Combined student Go production is **4,869 / 1,083 / 6,295**, growth
**527 / 97 / 652**; tests **2,664 / 213 / 2,962**, growth **317 / 15 / 339**.
The original core moved from `ch07/agent` to the `ch08` root; it was not rewritten
merely because paths changed. Its earlier separate ch05–ch07 examples total
609 / 91 / 778 and are absent from ch08; their omission is not a code reduction
to credit. Its chapter growth also includes reasoning configuration and associated
tests beyond the GUI mechanism, so total deltas are not an equal-feature contest.
Student browser growth implements missing required speech, pause, safe rendering,
full-content access and one reusable card path, not an early workbench.
Two small `.mjs` browser/speech evidence launchers total **78 active / 2 comments /
80 physical** lines of test support outside the shipping modules; their compact
formatting is not credited as production simplicity.

## Verification and actual evidence

Independently ran full uncached core and GUI race suites: PASS (core integration
31.090s, GUI 2.842s). Core/GUI vet and formatting pass. Reuse unchanged separate
example-module results. New tests run real HTTP parsing, Agent, history, tools,
shell processes and WebSockets against small external peers. They cover partial
and completed replay, more history than the live queue, both replay-window bounds,
global unpause, paused cancellation/interrupt and continued use, tool lifecycle,
logging and disconnect/shutdown. No network or speech mock replaces the behavior.

Read complete initial and corrected mutation outputs, including the long cleanup
hang and surviving finalized-part cleanup mutant. The first partial-replay failure
was obscured by failure cleanup, and the first final-cleanup assertion occurred
after TurnEnded had already cleared partials. Both are corrected, with direct
clean intended failures. The dispatch-observation check now observes the actual
announcement, and replay truncation preserves terminal status so it fails at
249-versus-270 facts rather than an unrelated receive timeout. Pause/cancellation,
identity, lifecycle, logging, window, unsafe HTML, full-value access and native
speech-error mutants all have intended failures. These do not cover the newly
found empty/opaque display boundary.

Inspected the retained browser launchers, native speech events, three-provider
wire/event/visible captures and summaries. Browser tests exercise two tabs,
partial and completed replay, safe Markdown, reachable long output, typing pause,
explicit other-tab unpause, Escape and visibility. The speech wrapper forwards to
the real native API. `speech.json` shows first utterance start, native interrupted
error, second utterance start/end and released pause. The corrected automatic
speech record has queued/start/end, and the live Anthropic speaker click does too.
Initial cancellation and the reload-restored second-tab voice toggle mistake are
preserved. No subjective listening or intelligibility claim is established; this
limitation is stated honestly and is not replaced by a DOM-only claim.

Actual model destinations are Sonnet 4.6, GPT-4o mini and required Gemini 3.8
Flash. Each real command returns its marker and exit 0; observed deltas are
28 / 27 / 5. The Anthropic terminal busy hint appears as a durable/browser human
message on the same Agent. These successful interactions do not erase the extra
Gemini metadata card, which their retained visible capture exposes.

## Original grading and departures — initial pending assessment

The original reference baseline is **100/100 PASS**, as recorded by the
coordinator. The student's one ongoing original run submits `main/gui`, not the
headless core. It recursively launches CLI exercises against a GUI server;
the harness uses the same submission directory at each chapter. Those nested
timeouts are not evidence that core CLI behavior regressed. The accepted previous
core result remains **0/100 FAIL** with the Chapter 7 unknown-model fixture and
inherited logger/text-model exceptions. No new grade is inferred from it.

The ongoing run built before the final socket tracking/writer-join and terminal
display edits. Label its eventual result as that earlier binary, not exact final
source grading. Current full GUI race/vet/build and rebuilt browser/live evidence
exercise the final source. Repeating the whole inherited timeout chain solely
for those bounded edits would not add useful evidence; reassess affected behavior
and any actual new failure instead. No second original grader was launched here.

Flags override inherited environment defaults. I agree with that narrow departure
from the chapter's literal flags-only sentence: the original harness itself uses
environment configuration, and retaining key setup avoids forcing secrets into
command arguments. The documented timestamp/direction/JSON log format follows
§8.8 and the supplied narrow contract. Ownership rules properly override the
common-package gate sketch; synchronized history correctly rejects the text's
unlocked-slice claim. No restart persistence is required yet.

Actual new GUI failures, if any, must be evaluated separately from inherited
parity before acceptance. No blanket new-grader waiver is issued while the result
is pending. The three current findings above require correction regardless of score.
Reviewer reproduction created no files or persistent processes; production,
originals, graders, frozen exports and tags are untouched.

## Correction round 1 — implementation findings resolved

Read the full mandatory skill and updated student correction review before
reassessment. Read the entire projection browser launcher, baseline, initial
failed mutation attempt, completed mutation result and replay-status mutant output.
All three original findings are resolved without new architecture or feature scope.

`partText` now preserves empty text and projects only readable Anthropic thinking
or Gemini thought text from opaque parts. Other opaque bodies become empty display
text; signed/provider data remains unchanged in history. The new browser test uses
the actual retained Gemini part and a small external WebSocket peer, running the
production page and forwarding instrumentation to native speech. It checks display,
automatic speech input and speaker-button input separately. Readable controls
pass with native start/end events; empty/redacted/signature-only inputs produce
no speech body. The restored JSON fallback fails all three intended assertions for
each bad input while the readable controls still pass. The original cancellation-
interaction timeout is retained; the reordered final mutant completes without it.
I also independently evaluated the actual projection function against the retained
empty part plus both opaque-negative and readable-positive forms: all five pass.

Replay status retains `replay:true` and reports equal current-state endpoints.
This fills the public `from`/`to` envelope truthfully without manufacturing an
earlier transition. The new test asserts that contract before any turn. Its
missing-`from` mutant fails the intended assertion cleanly. `trace.Write` now
documents synchronized writes/read access. Independently ran uncached GUI race
tests for replay status, window and streaming/pause integration: PASS (1.216s).
Reuse the student's full GUI race/vet/format results and unchanged core/module/
live checks. No paid repeat or whole mutation sweep is justified by these edits.

Count updates, using the original method: GUI Go production **405 active / 69
comments / 489 physical**, round delta **+1 / +2 / +3**; GUI tests **323 / 16 /
348**, round delta **+12 / +3 / +16**; browser JavaScript **320 / 38 / 361**,
round delta **−2 / +2 / 0**. Other shipping categories are unchanged. Combined
Go production is **4,870 / 1,085 / 6,298**, chapter growth **528 / 99 / 655**;
combined tests **2,676 / 216 / 2,978**, growth **329 / 18 / 355**. The new
projection evidence launcher adds **80 / 4 / 84** lines of non-shipping test
support. The initial qualitative comparison remains applicable.

**Exact remaining gate:** collect and independently assess the single running
original GUI grader's actual outcome. It built the earlier binary described above,
before writer-lifetime fixes, terminal display and now the replay-status correction.
Its eventual score cannot certify round-1 source or erase genuine initial defects.
Affected final-source checks establish these corrections without repeating the
unrelated inherited timeout chain. Any reported new GUI failure still requires
its own diagnosis; optional-GUI/CLI structural mismatch and previously accepted
Chapter 7 exceptions do not automatically waive it. No implementation finding is
currently open, but chapter acceptance remains withheld until this grading record
is resolved. Reviewer checks left no scratch files or processes.

## Final grading assessment and acceptance

Read the completed original output and the student's final result/provenance
append before deciding. The sole original command was
`make grade-dir CH=8 DIR=solutions/edition-2/main/gui`; actual result **55/100
FAIL**, grader exit 1 and Make/tool exit 2, after roughly 47 minutes. It graded
the earlier built GUI binary, not checkpoint `2af66bb` or the headless core.

| Original check | Actual result and disposition |
| --- | --- |
| Observation types/fields | **0/25 FAIL** for replay `state_changed.from`; genuine initial defect, fixed and independently verified in round 1, not waived |
| Event replay | **20/20 PASS** |
| GUI logging | **15/15 PASS**, 22 lines with both directions |
| Pause holds tools | **20/20 PASS**, three calls complete after unpause |
| Chapter 7 parity | **0/20 FAIL**, inherited CLI launches/timeouts and module-topology/refusal checks applied to the GUI submission |

I accept the student's requested **aggregate inherited-entry-point exception**.
Independent harness inspection confirms `Ch8Run` passes its same GUI directory to
`Ch7Run`; those inherited launchers expect finite protocol exercises and scan the
submitted module's core topology. This GUI command is a continuing HTTP server
in the required separate optional module. Five streaming check labels report
exercise timeouts; that is not five independent process launches. Those failures,
no-hub/common-import findings and text-model refusal finding do not establish a
regression in the actual core command. The real common/star owner chain, shared
Agent, independent core tests and previous core grading remain separately reviewed.
Adding a second core exercise implementation or coupling GUI into core solely to
satisfy this directory assumption is not warranted. Later grader maintenance
should accept the appropriate separate core/GUI entry points and retain meaningful
parity coverage. This is a narrowly recorded exception, not a waiver of behavior.

The prior actual core Chapter 7 **0/100 FAIL** and its fake-model streaming and
inherited logger/text-model exceptions remain visible and unchanged. No new core
score is claimed. No exception for the `think` fixture is needed: the actual pause
check passed, and the independent controlled real-shell pause evidence remains
the stronger causal test. My earlier concern about that fixture does not become
a claimed failure or additional waiver.

The missing `from` field is an implementation correction, not a grader defect.
Its initial zero score stays recorded. Current-source tests and the restored-field
mutation establish its repair; other required lifecycle types/fields are supported
by real WebSocket tests and live captures. The old grader result does not prove
all later field checks would pass. Socket writer lifetimes, terminal display and
browser projection likewise use their already-reviewed affected final-source
evidence. No exact final-source original grader was run, and no hypothetical
replacement score is assigned. Repeating the known inherited timeout cascade
would not meaningfully verify these bounded changes.

**Chapter 8 is accepted with the entry-point/inherited parity exceptions and the
documented flags/environment departure.** No architecture, current-feature,
test-sensitivity or scope finding remains open. Counts and limitations are the
round-1 figures above. Native speech events are verified without claiming human
listening or intelligibility. Original failures, mutation survivors, capture
corrections and initial findings remain in history. Student runtime logs have been
removed; current source status is clean at its progress checkpoint. Reviewer
changed only this report, with no source, grader, original answer, export or tag
edits and no paid repeats or remaining scratch/processes.
