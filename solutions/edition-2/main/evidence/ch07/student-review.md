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

## Revision phase after initial checkpoint 8cc87f2

Reread the entire skill, architecture ledger and current Chapter 7 before
revision work. The independent reviewer receives the historical standard;
I receive rationale and public failures, not its code or grader internals.
Original evidence scripts, binding, launches and receipts remain unchanged.

Read the author's published §7.3/§7.6 correction adding `job_killed` to the
snapshot window and explaining both inherited terminal facts. This resolves
the teaching omission I reported. The narrow filter now includes it. A public
regression starts and kills a real local scratch process through a controlled
model, then requires the watch snapshot to retain its running report followed
by the killed fact. The first draft fixture omitted mandatory usage and was
rejected as malformed; adding valid fixture usage resolved that test-only
mistake. The valid control passes. Removing just the filter addition in a
disposable source copy fails with running=6, killed=0 as intended; receipt is
`revision-1/killed-deletion-control.json`. No paid request was made.

The reviewer also found that `Page.close()` leaves DOM listeners installed,
so constructing a replacement Page on the same root loses a new submission
to the old closed handler. Proposed lifetime plan: Page owns and removes its
control listeners; ArtifactScroll owns and removes its scroll listener and
closes its cards; each Artifact removes its expand/speak listeners. Closing
is idempotent and preserves Connector/speech teardown. Awaiting published
teaching/coordinator check before that affected repair. This is a reusable
component lifetime defect, not an excuse to change model or provider behavior.

### Browser ownership revision plan (published clarification read)

Read the new §7.1 application parent, §7.6 idempotent listener disposal, and
§7.7 shared native speech/FIFO/cancellation paragraphs before affected code.
The proposed public `BrowserApplication` is the document root. It owns Pages,
the native synthesis/utterance constructors, a `SpeechService`, and diagnostic
logging. `createPage(root, url)` constructs a Page with its actual application
parent. Application close closes its Pages and service; there is no module
registry. Existing bootstrap and the public two-Agent embedding each create
one application and use it to construct their Page children.

Page continues owning Connector, ArtifactScroll, input causes and SpeechQueue.
SpeechQueue reaches the native service through Page.application(), not an
injected sibling. It offers its current utterance after the speaking pause ack;
other queued text remains local and still holds the cause. The service owns
FIFO requests from these Page heads, with Page identity, and submits exactly
one native utterance. A settled head allows its Page to offer the next one.
Service cancellation removes the requester's pending head and native-cancels
only when that Page owns the active utterance. Other Pages remain queued and
continue. Active-request identity and local queue generations fence stale
callbacks. Native availability and diagnostics stay reachable through the
application parent. No locks are needed in this single-threaded browser state;
await boundaries and native callbacks require explicit closed/generation checks.

Page closes idempotently: mark closed, remove its owned control listeners,
close Connector (including reconnect timer), close SpeechQueue, and dispose
ArtifactScroll/cards. Pending submit/reconcile/native callbacks cannot write
old notices, clear replacement input or move its focus. ArtifactScroll closes
its own scroll handler and child Artifact listeners. Creator back-pointers are
retained throughout; new Pages on the same root get fresh owned state.

Existing public component fixture adapters must use the new application parent;
no old behavioral assertion is removed. Local controls will distinguish active
and idle Page cancellation, FIFO pending speech, stale native callbacks,
continued typing causes, same-root Page replacement and component disposal.
Real native speech with retained answer cards can validate the changed browser
owner path without paying to regenerate identical provider output.

The coordinator accepted this owner plan, including FIFO over submitted ready
heads and cancellation fencing before invoking native cancel. Implemented the
application/service chain, updated bootstrap/public embedding, and added Close
view/Reconnect view controls to the example for actual lifetime use. Page closes
its handlers before closing children; Connector cancels reconnect work, rejects
pending promises, and detaches socket handlers. Late promises cannot update the
old Page's notice or focus. Component reset also closes discarded Artifact
handlers. Local service controls cover idle close, queued and active cancellation,
FIFO heads, synchronous/stale callbacks, and typing through an error. The
independent reviewer reports the six retained browser groups, same-root remount
and five shared-native groups passing this revision.

My own boundary audit found the initial oversized-message workaround only
covered 65,537 bytes. Larger/fragmented messages hit Gorilla's automatic empty
close first. The retained before probe passes 65,536/65,537 and fails 131,072,
1 MiB and fragmented 131,072; the revised bounded NextReader/LimitReader path
passes all five with explanatory closes and zero model requests. This proves
why the old comment about only draining the exact boundary is replaced; the
underlying general explanatory-close requirement is now handled at our bounded
reader instead of the library's header-level limit. Before/after binary hashes
and untouched failures are under `revision-1/oversize-*.json`.

Read the author's final initial-live disposition table and confirmed it resolves
all recorded teaching suggestions without altering original outcomes. The
revised shared speech/application behavior requires a narrow fresh demonstration:
two prompts per provider in the public two-Agent consumer, at most four HTTP
requests per provider, separate revision evidence/budgets. Root authorized that
scope. The initial nine sessions remain bound to their original source; no job
is regenerated merely to prove the snapshot projection repair. Retained-event
replay is labeled as a local control, not a new provider session.

After context compaction, re-read the entire coding skill, architecture, AGENTS
and current Chapter 7 (no historical links followed). Root and reviewer found
a teardown ordering flaw in the grouped R3 work: closing active Page A could
start pending Page B before the root reached B. The root now closes native
service admission before its Page children. A positive active/pending control
with a synchronous stale callback proves whole-application close starts no new
utterance; individual Page cancellation still advances healthy peers. This is
an implementation correction within the accepted ownership contract.

Revised runtime/support froze at 9ba7855 after the shared-index author commit
49ae919 had captured my staged batch; the later explicit-path commit preserves
that chronology rather than rewriting it. No Chapter 9 text was read. Final
independent checks pass 33/33 and comparative revisions 12/12. Revised evidence
adapters pass 18 local controls before paid use. Fresh public two-Agent sessions
on all three backends each used two prompts/two HTTP requests, exit 0, with six
exact independent request reconstructions and six real captured audio samples.
The revised live results detail real Page remount, idle-peer lifetime, owned
queued/active cancellation and retained typing. Initial evidence is unchanged.

The revised Gemini example hit MAX_TOKENS on both 512-token requests because
thinking consumed most of the budget. The actual short text still exercised the
changed client and produced audio; there was no retry or claim of paragraph
completion. This is a demonstration budget limitation, reported to the author,
not an invented teaching/ownership defect. Anthropic's first queue exercise
started after A had finished naturally; its repeated speaker-only selection
supplied the necessary overlapping sequence without another model call. A first
local WAV inspection using Python wave failed because the audio is IEEE float32;
explicit RIFF fmt/data decoding then measured silence and speech correctly.

Read current main README/CHECKPOINT for documentation reconciliation, including
its dated historical student records; no old implementation or grader was read.
Updated current usability/identity statements while retaining earlier checkpoint
history. The author's initial-live dispositions accurately resolved the recorded
teaching findings. The later explicit killed-event list, component disposal and
application/native-service ownership paragraphs now resolve the comparative
ambiguities and match the final implementation. Final export/tag remains the
coordinator's task, and no additional runtime or paid revision is pending.
