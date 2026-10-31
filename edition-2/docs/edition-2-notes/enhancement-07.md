# Enhancement notes: Chapter 7 (original Chapter 8)

> Historical paths: the `edition-2/ch04`–`ch09` paths recorded below use
> original student numbering. Those exports are now preserved under
> `solutions/edition-2/history/`; see the [solution mapping](../../solutions/README.md).

Author: `enhancement_author_resumed`, October 10, 2026. Final prose submitted;
coder samples/results reconciled, independent combined review pending.
Mandatory handoff: `.agents/skills/ensemble-coding/SKILL.md`.

## Source and preservation

Read all 495 lines of original `book/chapter-08.md`, file revision
`994bf218b31033b7dbf2740dd0187d509d2927dc`, SHA-256
`207ae885a7a9d6210757b5f3b4d9773a98206a66e62710e5ce0be0a6ce180f56`.
Read full ch08 student/coordinator history and comparative review, accepted
GUI/browser/public/actor/history source, and retained browser/native-speech/live
receipts. Endpoint is `solutions/edition-2/ch08`, `edition-2-ch08-r3`, source
unchanged from reviewed `2af66bb`. Historical IDs, tags, scores and originals
remain unchanged. Governing skill/carryover/voice/procedure were read; current
bounded enhancement supersedes retired author-first rules. Bill's right-length
instruction is binding: no trimming or padding target, no cuts merely to satisfy
a soft count.

The through-line remains the human's ability to watch and steer the Agent.
Keep original title, opening and section order: radio/receiver analogy, one
Artifact component, pace as safety, serial-dispatch WHY comment, interruptible
pause, two tiers of replay, missing log direction, safe markup/full-value access,
Chrome wakeup workaround, error-advancing speech queue, local transition pacing,
two views of one Agent and the workbench cliffhanger. Corrections replace false
or superseded mechanism claims in place; they do not require a new architecture
exercise. Add actual student browser/projection and speech findings where they
explain the existing behavior.

## Findings and dispositions

| Finding | Enhancement and reason |
| --- | --- |
| Optional GUI and owner access | Separate module consumes public Ensemble/Agent; root identifies observations, Agent owns config/gate, Engine retains actual parent. Transport owns sockets, not Agent. No sibling bags or common behavior. |
| Observation vocabulary | Actual tool boundaries plus TurnEnded display progress; existing PartDelta/PartFinal/StateChanged preserved. ToolFinished means wait result, not proof continuing job ended. Reliable replies remain distinct. |
| Serial dispatch | Keep rationale and comment, correct completion wording to reported result so it does not contradict continuing jobs. |
| Pause and interruption | Gate lives with Agent; boundary waiter outside mailbox, cancellation ends current turn while Agent/jobs remain usable. Explicit any-tab unpause releases everyone; no stale local veto on reconnect. |
| Reconnect synchronization | History locks suffix copy; handoff coordinates captured facts/partials/subscriber admission. Replay batch cannot silently truncate at live-queue capacity. Preserve lossy-live distinction. |
| Identity scope | Parser IDs remain response-local; actor allocates Agent-lifetime display identity and stores it with finals. No disk-restoration or new persistent protocol promise. |
| Fixed replay policy | Last hour OR last 100 renderable events, not a configurable feature; multipart event counted once. Process restart remains later work. |
| Initial replay status defect | Include from==to truthful state snapshot; original missing-from failure remains actual rather than waived. |
| Logging and launch contradictions | Timestamp/direction/JSON line format; flags override inherited environment, secrets stay out of arguments. Build actual optional cmd from its module; grader core parity separated. |
| Final cleanup/replay mutation weaknesses | Retain initial survivor after TurnEnded and corrected mid-tool assertion; distinguish initial timeout from direct 249-versus-270 replay evidence. |
| Empty/opaque projection | Actual Gemini empty-part metadata card motivates explicit readable-text projection; signatures/redacted bodies remain in history without becoming display/speech text. Reuse actual artifact, no paid repeat. |
| Native speech evidence | Real browser queued/start/end/error events and positive controls; no subjective listening claim. Preserve initial canceled event and reloaded-tab Read aloud mistake. |
| Three-provider live exercise | Preserve actual 28/27/5 delta turns, exit-0 markers, Anthropic same-Agent terminal hint and speaker events. Gemini 3.8 follows Bill's restriction. No current-availability or new-run claim. |

No workbench/settings, disk restore, parallel tools, GUI-dependent core, second
Agent loop, new speech service or sub-agent work is added. Stable source and
browser evidence remain necessary independently of numeric grading.

## Grader and validation

Coder/coordinator agreed original GUI weights 25/20/15/20 plus corrected
`ch6-parity` 20. `-gui DIR` selects separate optional module (default core/gui);
`-gui-model` selects the historical declared Anthropic fake fixture, default
`claude-sonnet-5`. Existing earlier flags carry forward for real core parity.
Completed replay checks actual final/tool content and identity; log lines check
timestamp/direction/JSON; real gated tools verify pause and other-tab unpause.
Public submitted-core boundary is checked alongside websocket-streams. Active
partial/handoff/window/paused interrupt/browser/native speech remain source and
integration obligations, not falsely claimed numeric coverage.

Original student 55/100 FAIL graded an earlier binary before replay-field,
writer-lifetime and terminal corrections. Original legacy 100/100 stays separate.
No final-source original pass or hypothetical replacement score is assigned.
Final source/race/browser repairs were reviewed independently in one correction
round; the bounded enhancement's independent combined decision remains pending.

Final [coder evidence](enhancement-07-code.md) reports revised composed results
of accepted 100/100 PASS and legacy 35/100 FAIL: actual rerun GUI components
80/80 and 35/80, combined with unchanged full-run corrected parity 20/20 and
0/20. Initial revised full runs scored 100/100 and 20/100. The initial trace
parser wrongly rejected blank separators produced by the legacy logger's
newline handling; final parsing ignores those separators and validates every
actual nonempty timestamp/direction/JSON record. Preserve the earlier result
and correction, rather than calling the composed score another full run.

Legacy earns logging and actual pause points, but has observed final-identity
and replay-correspondence failures in addition to its integrated GUI boundary.
Its inherited missing separate exercise applications are setup limitations,
not evidence those tools executed and failed. Nineteen changed-behavior mutants
failed their intended checks; an initial malformed mutation failed compilation
and is excluded. Source race tests preserve the stronger causal replay,
paused-interrupt and window evidence. No new browser or paid live result is
claimed in this pass.

Independently counted the retained raw provider wire records: 28/27/5 deltas,
all three exact markers and exit 0. Gemini's actual empty final is artifact 2,
with type/from metadata and no text, confirming the display correction's source.
Read native browser/speech and all five projection cases directly.

Prose lint passes every hard check: 4,396 words versus original 2,558. Inherited
and remaining negation/short-closer/person-gap warnings stay visible for review;
the count is a measurement, not a target. Exact-line comparison retains 367 of
495 original lines, replaces/removes 128 and adds 333, including numbering,
corrected schemas and required evidence. All local links resolve; twelve JSON
records parse. Submitted SHA-256:
`067f449d07535d60d952059fadb0c54e1c17f07a92504e55f22fdc5d27c4c180`.
The coder checked this exact hash, twelve JSON records and the Go observation
vocabulary. The own-module build and published port-8088/asset-path launch passed
against a faithful provider; only the binary destination was made private and
the explicit asset placeholder filled. All six client envelopes ran, followed
by another successful prompt. An initial scratch assertion mistook the expected
interrupted TurnEnded for that next reply; correlating the reply text showed
the interrupt followed by `Still usable.`. The initial failed assertion remains
in evidence, and no manuscript or runtime correction was needed. Server examples
are explicitly individual illustrative envelopes, not a claimed transcript.

Section 7.5
distinguishes inbound-reader and outbound-writer logging and promises no network
wait in the observer, not absence of brief synchronization waits.
No original/frozen/runtime edits, commits, mocks, paid calls or next-chapter work.
