# Chapter 7 outline: A browser you can steer from

Status: proposed student contract, awaiting independent/coordinator review and
validated Chapter 6 predecessor. No Chapter 7 implementation or live evidence.
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

## Architecture choices for review

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
A missing checker invocation remains a handoff gate, not permission to claim
100 from the old five checks. Publish the command before student grading.

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
The author writes the observed transcript later, retaining failed attempts.
