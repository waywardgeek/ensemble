# Chapter 7 initial student review

## Source and read ledger

Started 2026-10-07 from accepted `edition-2-ch06-r1`, commit
`1c6b1f065d11bd3a532c94bc53305394e17e7cc2`, main source tree
`91f6da877ab76fcda6887bc3ec523332a499fbd0`.

Read the entire repository `AGENTS.md`,
`book/edition-2/skills/ensemble-coding/SKILL.md`, and
`book/edition-2/architecture.md`. Read the complete new Chapter 7 and Chapter 1
architecture sections (1.1–1.3) from the coordinator's committed-only copies in
`/Users/bill/projects/ensemble-edition-2-revisions/ch07-student-inputs/` and its
manifest. Initial large combined tool output truncated; repeated the skill and
Chapter 7 in smaller reads to retain their complete instructions. Inspected
accepted main's `ensemble.go`, `internal/common/actor.go`,
`internal/llm/actor.go`, GUI stub and module. Listed paths under main but did not
read grader implementation, old/future chapters or solutions, author notes,
reviewer notes, or historical links. No forbidden-source exposure.

## Ownership and concurrency plan (awaiting coordinator contract check)

- Ensemble remains application/logger owner. Agent owns its actor, Engine,
  Jobs, authoritative durable events/configuration, watch projection and pause
  registrations. Common declares public transport-independent snapshot, state,
  tail and registration values/interfaces. Actor implements their behavior in
  `internal/llm`; root adapters expose public Agent and ClientOwner operations.
- Actor alone changes the registration cause map, aggregate counts, watch
  revision, active-operation metadata and incremental partial buffers. A
  registration carries its actor parent and immutable identity. Updates and
  close enter the mailbox; acknowledgement follows projection publication.
  Closing an already closed registration is harmless; later updates fail.
- Tool admission checks the actor-owned aggregate immediately before the normal
  call lifecycle. A held call remains the current serial index; clearing the
  final cause schedules dispatch, while hint/interrupt/close/job facts continue.
  Interrupted unstarted calls retain the inherited refusal/pairing behavior.
- Durable publication is routed synchronously through actor observation while
  actor-owned mutations run. Watch capture selects the last 100 renderable
  immutable events under Agent synchronization and registers its tail at that
  actor boundary. Current configuration supplies only safe model identity;
  Engine supplies usage under its lock. No network operation enters capture.
- Recoverable partials use appendable owned byte buffers per part/channel and
  an incremental byte count, with only active operation retained. Snapshot
  materialization creates owned values. Observations have one Agent-local
  revision; subscribers retain owned payloads in independent bounded queues.
  A queue mutex protects its bytes/items/status; publication never waits for a
  reader. A done signal closes once; data channels are never closed beneath
  producers. Overflow invalidates the watch and subsequent reads report loss.
- Optional GUI Server owns connections, trace and immutable local origin;
  Server's public parent is ClientOwner. Connection reaches Server for watch,
  registration, trace and Ensemble logging. One connection lifetime cancels
  workers, closes its socket/watch/registration and joins reader/writer work.
  One writer owns data writes and deadlines. Queue operations never hold a
  recipient-set lock across I/O; snapshot enqueue may wait cancellably outside
  actor, live enqueue overflows explicitly. Trace lock serializes its writer.
- Public Go Server/Connector and browser Connector/ArtifactScroll are reusable.
  Page controller owns connector, cards, input causes and speech queue; children
  reach diagnostics/control through their actual parent. Browser derived cards
  and speech cursors use complete part identities and authoritative final keys.
  No browser state becomes conversation authority. Shared CLI extraction keeps
  existing parsing/control behavior in one public client, with explicit EOF
  detach versus quit semantics for combined serving.
- Creation-only settings: connection Agent identity, server origin/listener,
  trace destination, immutable queue limits; inherited workspace/log/builtins
  remain creation-only on Agent. Registration causes and connection generations
  are transient; reconnect creates both afresh. Server close joins connections;
  application close joins Agent; tab close never cancels an accepted request.

## Initial teaching observations

The chapter specifies ordering, ownership and exact wire examples together;
the snapshot watermark, count-sensitive pause observation and typed projection
examples make otherwise ambiguous boundaries concrete. No architectural gap
identified at this initial read. Implementation and live difficulties will be
recorded here as they occur, before comparative review.

## Implementation progress before comparison

Coordinator accepted the owner/concurrency plan, including synchronous actor
publication, and the public CLI placement. CLI implementation and its existing
unit tests moved to `example.com/ensemble/cli`; the command is a small public
entry point. The public client imports the public library, and no internal
spoke imports a client or GUI. The inherited core and GUI test suites passed
before changes, and core tests still pass after initial watch/pause changes and
CLI extraction. No baseline failures occurred.

Consulted the official Gorilla WebSocket documentation at
https://pkg.go.dev/github.com/gorilla/websocket for its reader/writer and control
concurrency rules. Optional module pins v1.5.3. No example implementation was
copied. Additional permitted reads: inherited `internal/common/types.go`,
`internal/llm/engine.go`, GUI stub tests, and CLI implementation/tests.

### Command refusal ambiguity and resolution

The initial local transport checker reported 10/15. Two failures demanded a
same-socket pause observation before ack (stronger than the actor publication
contract); two expected semantic JSON errors to close despite the correlated
unknown-field error example; oversized sending reported BrokenPipeError while
the server enforced its read limit. This was reported immediately to root.
Read the author/coordinator clarification in new Chapter 7 §7.4 at `1bba1bc`
and `book/edition-2/chapter-07-student-feedback.md`. It resolves the ambiguity:
transport failures/unusable IDs close, semantic errors with usable IDs receive
correlated refusals and remain usable. Changed unusable-ID handling accordingly.
The original failed attempt remains a failure, not a retrospective pass.

Coordinator's initial ownership inspection also found idle closed-watch
retention and an overflow branch that could offer a truncated partial snapshot.
The watch now posts removal to the actor even without another publication,
clears removed backing-array references, and refuses capture while projection
is invalid until model end. These are implementation corrections from current
contract inspection, not historical comparison. The independent public checker
passed all five initial groups under the race detector after core changes.

### Browser and local acceptance before paid source freeze

The independent updated wire gate passes 23/23 after the clarified semantic
error policy. Exact oversized transport initially lost its explanatory reason
because a header-level Gorilla read-limit close preceded TCP teardown. The
reader now retains one extra byte, rejects the 65,537-byte boundary after
consuming it, and emits an explanatory static close. The old failed receipts
remain. Independent component checks pass watch bytes, publication order,
projection invalidation, selected-recipient close and recursive projection.
Independent Chrome checks pass six groups covering cards/identity, content and
accessibility, job lifecycle, speech callbacks/cursors, Connector reset and
actual Page input/uncertain acceptance. The grader identified and I fixed a
Page promise-order bug: pause rejection had hidden an uncertain prompt receipt.
Both promises are now owned immediately, and submission uncertainty wins.

Real Chrome 154.0.8037.93 speech produced start/end callbacks for a fixture card.
The initial headless ScreenCaptureKit attempt found no shareable Chrome app;
headful capture succeeded. A stronger retained run selected exactly the test
browser PID 75583, with a two-second silent lead-in. First 1.8 seconds measured
at the -91 dB floor; the 2–4 second speech window measured mean -22.7 dB and peak
-3.8 dB. Original audio and timing/PID receipts are retained. This proves produced
audio in that isolated browser; it does not claim Bill listened or transcribe it.
macOS screen-capture preflight was already granted; no permissions were changed,
and no microphone or screen-video stream was captured.

Additional permitted reads: Chapter 6 `evidence.py`, `terminal-run.py`,
`verify-receipts.py` and `test-evidence.py` to adapt its validated evidence chain.
Consulted official Apple ScreenCaptureKit documentation for audio capture API.
No historical solution or grader implementation was read. Local source tests
and `go vet` pass in affected modules; `gofmt -l` is empty. Paid sessions and
initial student-experience conclusion are still pending.
