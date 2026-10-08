# Chapter 7 outline: A browser you can steer from

Status: accepted contract with initial student source preserved. The
[validation record](chapter-07-validation.md) owns current local/live/review
gate status. Initial actual browser/PTY/public demonstrations are preserved at
`8cc87f2`; comparative revisions and final acceptance remain separate.
Maps to first-edition Chapter 8. Development stays in `solutions/edition-2/main`;
only the coordinator exports a validated `ch07`.

## Stake and story decision

The reader sees a wrong path appear, starts composing a correction, and needs
the next tool to wait without losing the running job, the answer, or another
browser's pause. Resolution belongs in an actual two-tab browser exercise,
with visible admission acknowledgement, correction, interruption and reconnect.

Retain old Chapter 8's radio/receiver explanation and reusable Artifact lifecycle,
but qualify the recording as durable events, not a lossy broadcast buffer.
Retain its concrete attention problem rather than its claim that comfortable
steering is the entire safety argument. Explain late Chapter 13's input-clear
bug here: assignment to an input's value did not emit the event on which
unpause depended. Teach explicit state reconciliation before a new student can
repeat it. Use the Chapter 22 disconnect crash as a short ownership lesson,
with commit `352b590` as its historical receipt. Do not reuse unverified Chrome
universals, send-site counts, timings or an invented Bill/browser session.

Preserve `0f05359`'s personal motivation: `book/gui-design.md` records the
750-wpm listening workflow and macular dystrophy. No receipt in the consulted
material establishes Eloquence as the voice used, a comparison with other
readers' speed, or how long the teardown crash took to diagnose. Omit those
additions while retaining the concrete need to hear, distinguish and steer
the work. The first browser is a single ArtifactScroll; separate panes,
draggable dividers and saved preferences belong to Chapter 8. Reconnect
promises the declared retained window and omissions, not a complete history.

## Teaching order

1. A readable screen matters only when its controls still reach the Agent.
2. Optional GUI module and reusable Server/Connector/ArtifactScroll components;
   core owns conversation, pause causes and atomic watch; GUI owns transport.
3. Snapshot plus ordered live tail at one actor boundary, bounded queues and
   explicit resync; immutable copies, full identities, no invented durability.
4. Per-client pause causes, acknowledged tool-admission boundary, responsive
   interrupt and disconnect cleanup. Job reports do not mean process exit.
5. Small explicit wire protocol, safe local server exposure and transport trace.
6. Artifact identity, safe DOM rendering, keyboard-accessible expansion and
   separate provisional/final state. Typed values survive later formats.
7. Speech queue cancellation and typing derive one pause predicate; replay is
   silent and canceled speech cannot restart through a stale callback.
8. Actual browser/PTY/public-client acceptance, adversarial fixtures and
   deletion controls; record real spin only after use.

## Accepted working architecture choices

Coordinator accepted Agent-owned pause registrations and actor-ordered updates,
with acknowledgement defining admission order. This supersedes historical
injected shared PauseGate mechanics under today's parent-chain rules, not a
new claimed Bill ruling. Actor never waits on the pause gate. Public atomic
watch returns an owned snapshot and bounded tail at a single actor boundary.
Agent owns transient presentation projection; Engine still owns assembly,
Jobs owns output and lifecycle, Ensemble owns logger. GUI server has a public
Ensemble parent; connection has Server parent; browser Connector and artifact
components have their actual client/controller owner.

Snapshot retention is a window of the last 100 renderable durable events,
with an explicit omitted-prefix marker and current active partials. It is not
restart recovery or a complete history download. A result whose call is outside
the window becomes a clearly labeled earlier-call card. Full old observations
are not retained as a second unbounded log. The optional module has public
component seams, and the headless module acquires no browser dependencies.

## Checks and evidence plan

The historical `CH=8` grader remains a diagnostic at its historical scope;
its fixture tool, binary layout and wire shapes differ from this contract.
New independent checks must exercise real browser state as well as transport.
The full immutable gate is now
`python3 scripts/edition2/accept_ch07_gate.py SOURCE_COMMIT`. Initial source
passed 32/33 groups; a corrected retained structural check closed the remaining
immutable-declaration false positive. The original failed receipt is preserved.
`python3 scripts/edition2/accept_ch07.py GUI_BINARY` now has 23 wire checks;
its original 15-check scope at `a5a07ba` belongs to the dated evidence entry.
The historical score cannot substitute for public, concurrency and browser
coverage, or the distinguishing regression for a subsequently corrected contract.

Local fixtures cover snapshot/live races at partial/final/tool boundaries;
two Agents reusing local part IDs; slow-client close/resync; sender/teardown;
two-client pause causes and paused interruption; safe DOM and expansion;
speech completion/error/cancellation; origin/header/message limits; trace
stage honesty. Positive controls must pass before each deletion/mutation.

Live plan: all three current API paths in actual browser and human CLI;
two tabs, streamed answer, tool proposal/result, typing pause, hint, interrupt,
reconnect during streaming and after completion, independent Agent public
embedding, actual speech when available and explicit cancellation. Use local
fixtures for schedule-sensitive races and synthesis failure. Capture real
screenshots plus text descriptions; source/binary binding precedes verification.
Section 7.8 now records the initial actual use, including tool-free narration
before the corrected file operation, delivered but unheeded hint text, and the
first audio selection of a user card. Runtime `da162e8`, launch binding
`ba902b7` and initial experience freeze `8cc87f2` remain distinct from repairs.


## Coordinator review clarifications

The coordinator accepted shared public CLI reuse for `--terminal` and the last
100 renderable-event window. Subsequent narrow review requested executable
pause transitions and recursive browser-safe projection. The draft now defines
`pause_changed` aggregate/count payloads, actor publication/revision ordering,
count-only changes, disconnect release and snapshot watermark coverage. Default
CLI protocol and the existing observe filter are preserved. Snapshot state
names an active begun operation even before its first fragment; reconnect does
not fabricate a second begin.

Projection recurses through response/message/result part lists, preserves order
and empty text, strips bound signatures and replaces raw opaque payloads with
placeholders. Exact valid payload fixtures cover response opaque data, nested
result signatures and an argument whose ordinary key is named opaque. No new
neutral result child kind or runtime architecture is introduced. These
clarifications were accepted before the Chapter 7 student handoff.


## Initial actual-use reconciliation, October 8

The recorded initial spin completes the intended story: typing holds the next
supervision admission, terminal interruption settles the turn, and an explicit
later cleanup kills the still-running job. Two actual screenshots retain that
distinction and visible pause ownership. Model explanations that conflate a
report with process exit remain recorded, with the tool facts beside them.
All three API paths also exercised terminal EOF and the independent-Agent
public embedding; produced browser audio is distinguished from human listening.

Student feedback identified a real teaching omission: the renderable-event list
named `job_ended` but omitted inherited `job_killed`. Root accepted the narrow
correction before affected code. Sections 7.3 and 7.6 now include both normal
completion and deliberate/shutdown kill, preserving existing interruption
semantics. Coder confirms the initial live card supports both but the snapshot
filter follows the old list; its repair requires a reconnect regression. This
is a later correction, not a successful property retroactively assigned to
the original source.


Subsequent browser review adds two explicit ownership requirements before repair:
Page/Scroll/Artifact teardown removes owned DOM handlers and cancels reconnect
callbacks, permitting replacement on the same root. An explicit document-level
application owns the native FIFO speech service; Pages retain their own queues,
input and pause causes through the actual parent chain. Selective cancellation
cannot stop another Page's utterance. Required replacement/two-Page regressions
and affected live repeats follow; the original demonstrations are not relabeled.
