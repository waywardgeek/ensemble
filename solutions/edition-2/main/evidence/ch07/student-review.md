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

### Pre-launch freeze and checker boundaries

Initial source is preserved at `da162e821369784b37cd6a6544610429af5f1cbe`.
Evidence-only support repair `ba902b7322a9cb5dd71c8352d45d6975663409c0`
requires the complete browser dependency map and launch equality before writes;
17 local adapter controls pass, including valid paths and intended mismatches.
No runtime source changed for that repair. The old zero-call browser launch and
its original binding remain retained separately.

The retained Chapter 6 gate failed because its source copier omitted embedded
assets and its CLI test injected into the former command package. The Chapter 7
adapter retains the assertions and accepts the public client location. Its
prior CLI checks pass. The initial full Chapter 7 gate passes 32/33 groups;
the remaining inherited structural heuristic mistakes an immutable error
sentinel and embedded asset filesystem for mutable session globals. The grader
is correcting its declaration/write analysis; no runtime repair is requested.
These checker limitations are not chapter teaching failures and their failed
receipts remain historical.

After context restoration I reread the entire coding skill, AGENTS.md,
architecture ledger, all 648 lines of the committed-only new Chapter 7 input,
and the current clarified §7.4 transport paragraph. No old solutions, grader
implementations or historical reviewer notes were read.

### First live provider difficulty

Anthropic completed the initial planned matrix in 9 prompts and 15 forwarded
model requests. Exact replay reproduced all 15 raw request bodies. OpenAI's
third prompt described its intention to write/read the file but made no tool
calls. The first two prompts had requested no tools, which may have influenced
that response; this is an interpretation, not proven causation. Preserved the
response and reported it to root. A deliberate human correction explicitly
ends that restriction and requests the missing calls within the original cap;
there is no automatic replay or claim that descriptive text proves tool use.

### Initial live experience completed before historical comparison

All nine planned sessions completed on the frozen binding: Messages with
`claude-sonnet-4-6`, Chat Completions with `gpt-4.1-mini-2025-04-14`, and
explicitly mapped generateContent with `models/gemini-3.8-flash`. Totals were
9/10/9 admitted prompts and 15/14/15 forwarded model requests respectively.
The cap was not reached. All application exits were zero, all 44 original
request bodies exactly matched independent replay, and actual browser drivers
recorded no page errors or failed actions. See `live-receipts.json` and the
separate `reconstructed-*-r1` derivatives; raw originals are unchanged.

The shared terminal/browser path was usable. I entered ordinary terminal text,
observed the answer, then `/history` and a follow-up. New tabs and reloads caught
up while operations were active. Anthropic happened to expose two nonempty
partial snapshots; OpenAI/Gemini cuts captured active-operation metadata before
fragments. Those timing differences are not substituted for the deterministic
held-response tests. Each provider wrote/read the exact 13-byte scratch file,
started a delayed job, received a browser hint, and was interrupted from the
terminal while typing held the next supervision admission. The next prompt
observed the job still running and deliberately killed it. Terminal EOF detached,
then a browser-only prompt completed. Standalone plain CLI and the custom
public two-Agent embedding also completed for all three providers.

The correction marker reached the next raw request for every provider.
Anthropic and Gemini printed it; OpenAI continued waiting without printing it.
The distinction between delivery and model compliance in §7.8 was useful here.
Likewise, models' explanations of reports sometimes overstated that a result
means a process finished; the UI's actual running report and later killed
lifecycle are stronger evidence than that explanatory prose.

Actual Chrome 154.0.8037.93 on macOS 26.3.1(a) synthesized selected cards. Each
provider's answer has an exact-browser-PID ScreenCaptureKit audio receipt, a
silent first second of PCM at the -91 dB analysis floor, and a later 2–6 second
speech window around -20 dB mean. Capturing starts asynchronously: the requested
two-second wall-time lead does not imply two whole silent seconds in the WAV.
One initial Gemini selection matched its user prompt, so it remains retained
and is labeled separately; a second deliberate selection captured the model
answer. No human-listening or transcription claim is made. Cross-tab speech
cancellation caused Chrome's real interrupted callbacks, while typing remained
registered; local controlled callbacks prove the stronger queue fences.

I visually inspected actual initial, provisional-answer and tool-result
screenshots. Anthropic `browser-26.png` and OpenAI `browser-25.png` show the
refused pending call, interrupted outcome, still-typed correction and connected
status. Gemini `browser-25.png` shows the running report, supervision call and
combined typing/speaking pause. The cards are readable and input focus survives
streaming. Text receipts retain offscreen cards; screenshots do not pretend to
show the whole scrollable history at once.

Teaching suggestions: keep the explicit receipt/model-behavior distinction;
state no-tools restrictions as applying to one turn in demonstrations; explain
that inherited deliberate cleanup emits `job_killed` while normal completion
emits `job_ended`. The first is a prompt-writing lesson, and the last is a small
terminology reconciliation with the inherited lifecycle, not a reason to
rewrite accepted core behavior. The transport-error clarification resolved the
one initial chapter ambiguity. Other corrections above were implementation or
checker issues. I have not read historical answers or comparative reviewer
notes; this is the preserved initial student experience, ready for that review.
