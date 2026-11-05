# Chapter 7: Everything Is an Artifact

The terminal from Chapter 6 is honest but limited. Text, thinking, and tool
arguments compete for the same screen. A tool the human cannot comfortably
follow is a tool the human cannot steer. This chapter adds a browser view
without making the Agent depend on that view.

## TL;DR

Build the browser as an **optional application** over the public C# API. The
Agent still owns conversation, scheduling, tools, and jobs. The GUI owns
presentation and connection lifetime, not a second transcript or actor loop.

- A locked `History.Snapshot()` supplies immutable events and context from one
  instant. Appending to a list is not a synchronization strategy.
- Final response events retain Chapter 6's response sequence and parser part
  IDs. A response-local ID of zero is valid. Provider call IDs and opaque
  signature data remain unchanged.
- Browser IDs combine Agent, response, and part identities. Historical records
  without parser metadata get deterministic event/position IDs instead.
- The host sends canonical history plus current partials on connection, then
  live frames. An internal history frontier is not a cursor supplied by the
  browser.
- Progress observations are bounded and lossy. Periodic snapshot catch-up
  recovers finalized facts. Each WebSocket has its own bounded output queue;
  a slow browser is disconnected, not allowed to block the actor.
- `ToolCalled` is an audit of a request. `ToolStarted` means admission succeeded
  and execution was permitted. A refused call must never look dispatched.
- Each Agent has one pause state. Pause and resume commands from any tab
  control that state in mailbox order; the server keeps no device hold registry.
  A tab addresses only its selected Agent. Speech preferences belong to the browser.
- Full-content speaker buttons are separate from automatic narration.
  Collapsed or scrollable presentation must not discard the rest of the text.

The implementation is `Examples/Gui`, a separate .NET 10 web project using the
installed ASP.NET Core shared framework. The core library references neither
that project nor ASP.NET. No package installation is required.

---

## 7.1 Why the browser is outside the Agent

The earlier chapters established the useful boundary: submit inputs through
an Agent's public operations, await reliable completion when completion
matters, and consume observations when displaying progress.

The browser adds no new scheduling authority. A button sends an input to the
same mailbox that a non-GUI caller uses. HTTP workers still return immutable
outcomes. Only the actor records conversation facts and schedules the next
step. Closing a tab does not cancel a turn, terminate a job, or destroy the
Agent.

The optional application's composition root constructs the Ensemble, Agent,
selected-model configuration, tools, logs, and hosting operation. Its `Common`
directory declares the GUI functions and value-like projection. `Server`
implements them. Static browser files live in `wwwroot`.

The server operation holds its connection queues and partial text locally for
that operation's lifetime. It does not introduce a persistent GUI owner into
the core graph. Those values are derived display and transport state, not
alternative authoritative conversation objects. Core destruction remains the
existing Ensemble → Agent → owned children cascade.

This distinction also limits the claim: the GUI is a trusted local control
surface, not a security sandbox for the tools. Choose a suitable workspace.
Binding to loopback and rejecting foreign origins does not make arbitrary
model-generated commands safe.

## 7.2 Give reconnect a trustworthy source

A browser can miss progress. It may disconnect during a response, reconnect
after completion, or simply consume its bounded queue too slowly. A list of
recent observations cannot reconstruct completed history reliably.

Use two sources:

1. **Recorded facts:** finalized response parts, human messages, tool admission
   and results, pause/resume, and terminal facts. These come from History.
2. **Transient partials:** accumulated chunks for an unfinished response. These
   belong only to the hosting operation and disappear when finalized or ended.

History exposes a value-like atomic read:

```csharp
public sealed record HistorySnapshot(
    ImmutableArray<Event> Events,
    Context Context);

// Declared in Common; implemented by History.
public partial HistorySnapshot Snapshot();
```

History locks its mutation and snapshot boundaries. Append still validates,
sequences, writes, and flushes before publishing the new state. Only afterward
does it publish an `EventRecorded` observation. That observation can be dropped;
the recorded event cannot be replaced by a claim that the observer saw it.

The host serializes subscriber registration, replay construction, and queueing
under one short handoff lock. It never holds that lock during a network send.
A connection receives its replay before becoming eligible for subsequent
frames. Its internal frontier prevents already-replayed durable facts from
being delivered again as newly recorded history.

There is a subtle ordering trap. By the time an old delta reaches the display
reader, History may already contain the response's final. Taking the newest
snapshot on every delta would publish that final first and erase genuine live
progress. The live path therefore projects through the observation's recorded
sequence. Full catch-up happens on idle polls and connection establishment.
A reconnect that already received the final must not receive old queued chunks
as if the part were unfinished again.

On a quiet observation stream, a short cancellable poll catches up finalized
facts. This is display recovery, not an Agent timer or turn deadline. It does
not schedule model work or cancel it.

Writing an event log does not implement application restoration. This chapter
reconnects to the same live Agent. Loading a saved application after process
restart belongs to later chapters.

## 7.3 One identity from streaming through replay

Chapter 6's parser IDs are local to one response. Part `0` in the next response
is a different part. A tool provider's call ID is a different kind of identity
again: it correlates a call with a result and must retain its provider meaning.

For finalized responses, `Event` now has:

```csharp
public readonly long ResponseID;
public readonly ImmutableArray<int> PartIDs;
public partial Event WithDisplay(long responseID, IEnumerable<int> partIDs);
```

`ResponseID` is the actual `RequestSent` sequence. `PartIDs` follows canonical
part order and contains the parser's IDs, not inferred array indices. Metadata
must describe every part, use nonnegative distinct IDs, and belong to a response
event. Numbering and serialization preserve it. Older logs without this
metadata remain valid.

The GUI combines these facts into a string such as:

```text
root:r27:p0
```

Human messages, results, and historical responses without metadata use a
separate event/position namespace. They never collide with a streamed response.
No global mutable display counter, identifier registry, or signature rewriting
is needed.

The browser uses the ID as the card key. A final replaces that card's partial
text; replay reconstructs the same card, not a duplicate. Repeated final frames
are idempotent. The server and browser both reject resurrection of completed
content by stale chunks.

### Keep the mount-local identity distinct

An empty `agent_id` can identify the original unnamed Agent borrowed by a GUI
mount. It is not a global address and must resolve directly to that actual
mounted object. Other selections need a live qualified address. When a selected
Agent is destroyed and loses its address, stop its observer instead of converting
it into the empty alias. Queue tracing must follow the same distinction.

Test the full lifecycle with a named selection replaced under the same name,
and repeat with an unnamed mounted source. A successful interaction followed by
a faulted observer during joined shutdown is a failed test, even if the browser
already received its answer or downloaded its file.

## 7.4 Truthful tool cards

There are three distinct claims:

1. The model requested a call.
2. The application admitted execution.
3. An observation of the job or control returned.

`ToolCalled` records the first. It is not proof that a handler ran. The dispatcher
validates schema arguments after separating observation options, preserves
explanation-policy refusals, and records `ToolStarted` only when admission
succeeds. It records that permission before starting effects. If recording
fails, the effect must not start.

A started fact does not mean an operating-system command succeeded. A command
may fail to start, a file may be missing, or a tool may report an expected error.
Those are outcomes, not reasons to falsify the admission history.

The GUI projects admitted execution as `tool_dispatched`. Refused calls have
no such frame. They still get their paired result and visible error status.
Retain the wire-level states `planned`, `started`, `finished` and `error` for
machine behavior, but label them **Requested**, **Dispatched: awaiting result**,
**Result returned** and **Error reported**. These labels describe admission and
observation without portraying a refusal as real execution.

`tool_finished` means the foreground observation returned. A running-job report
can be a legitimate result: waiting stopped, while the independent job remains
alive. Preserve the job handle and locator. Do not relabel that observation as
process termination.

When a result includes typed Job metadata, show its handle and **state when
observed** separately: running, finished or killed, with the recorded exit code
when available. Finished does not imply success, and an absent exit code is not
zero. Keep earlier snapshots intact when a later observation arrives. Reopening
an old transcript does not prove that a once-running Job is still running now.
Unknown metadata stays unknown; prose that resembles a Job report must not create
status facts. A live Job monitor would be a separate feature, not an inference
from this retained result. Error reported likewise does not by itself distinguish
a failed observation from unsuccessful execution; retain the exact result detail.
This also applies to `wait_for_job`, `send_input`, and `kill_job`: their results
can describe the existing Job. Preserve those typed observations through Agent
recording and replay, even though the controls create no new Jobs.

### Display recorded time and observation duration

Project finalized facts with their recorded timestamp. The timestamp belongs to
the retained Entry, not the browser's first paint or the diagnostic tail. Show a
local date/time and expose its exact offset-bearing value in accessible detail.
A streamed draft has no finalized timestamp yet. Keep a call's original recorded
time when its dispatch status arrives; each result has its own recorded time.
Missing historical values remain unknown.

Show a returned tool's optional duration as **Observation**, explicitly identifying
dispatch-to-return elapsed time rather than background Job runtime. Format the
recorded value; never start a live counter on a retained running result. Invalid
or absent timing metadata cannot become zero, current time or an inferred value.
Test rerender, snapshot-only replay, exact timestamp detail and the distinction
between an untimed draft and its finalized card.

### Offer controls consume typed facts

The later work-product feature renders a Download control only for a successful
typed offer, never a pathname inferred from text. Bind the control to the actual
card's Agent and opaque offer ID. Rendering or replaying an offer must make no
download request. A human click starts the transfer, and the server rechecks live
authority and the Sandbox. History can describe an expired offer without granting
access to it.

Retain the same control and feedback during ordinary rerenders. Give an active
transfer a Cancel action. Bound both declared and streamed byte counts before
creating a download Blob; an absent Content-Length does not remove the cap.
Abort on transcript replacement, and check the current card object again before
starting the browser download. A late response from an old card must not save a
file under a replacement card's identity. Expose expired/unavailable and cancelled
states without retrying automatically. Use download-only bytes, not inline HTML.

### Preserve paired controls and reading position

Use the result frame's `tool_part_id` to link a result to its actual call card.
Names, Job handles and result prose cannot establish that relationship. Each
result offers **View call**; its call offers **View latest result**. Earlier
results retain their own call link. An absent or wrong-kind counterpart leaves
an ordinary readable result without a guessed link.

Keep these as relationships between the current presentation objects, not a
second transcript. On activation, check both objects against the current card
map and selected Agent, and require connected nodes. A detached button from an
old replay must not target a new card merely because its ID was reused. Explicit
navigation focuses and scrolls the counterpart; its controls have at least a
44-pixel touch target.

Preserve the actual heading, details summary and download controls when a result
updates its call. Restoring only `details.open` after replacing the DOM loses
keyboard focus and pending-control feedback. Update raw text only when changed,
retain unchanged attachment and output controls, and defer repaint while selected
text intersects the card. Selection release schedules the deferred render.
Incoming output preserves pane position while the reader has focus or selection
inside that pane. Otherwise, only a reader already near the bottom follows new
output. Test both focus on the details summary and focus on the pair button when
another result arrives; those exercise different update paths.

`turn_ended` carries the completed answer when the turn completed normally, or
an explicit terminal reason otherwise. It is not inferred from a transient
return to Idle between tool rounds. The GUI derives the answer from the last
recorded response; it does not keep another authoritative answer field.

## 7.5 The wire protocol

The server sends JSON over `/ws`. The page and static assets are served from
the same loopback host. There is no SSE parser in the browser: vendor streams
have already passed through the core Engine.

Important outbound frames are:

| Type | Purpose |
|---|---|
| `replay` | recorded display frames, current partials, and Agent state |
| `part_delta` | Agent ID, part ID, kind, and a chunk string |
| `part_final` | stable part ID and a typed finalized part |
| `tool_dispatched` | admitted call ID, name, arguments, and job metadata |
| `tool_finished` | result identity, corresponding call card, result/error/job |
| `state_changed` | recorded effective-state transition |
| `agent_state` | current effective state, including pause |
| `turn_ended` | completed answer or terminal reason |
| `lifecycle` / `error` | interruption or other visible control/error information |

An empty chunk is a real empty string, not an omitted `chunk` property. Part
zero is equally real. Streamed tool names and arguments remain distinct at the
core boundary; their GUI projection includes readable name fragments.

Inbound commands are small mappings to existing public operations:

```json
{"type":"subscribe"}
{"type":"prompt","text":"Explain the selected function."}
{"type":"hint","text":"Keep the public signature unchanged."}
{"type":"pause"}
{"type":"unpause"}
{"type":"interrupt","mode":"now"}
{"type":"interrupt","mode":"after_tool"}
{"type":"shutdown"}
```

Connection registration queues replay automatically. `subscribe` is idempotent;
it does not start a new Agent or replay work. The server validates each addressed
pause/resume command against that connection's selected Agent. It does not track
which device caused a pause. The GUI has no separate Pause/Resume button pair;
Stop speech and Escape send explicit resume through the same public operation.

The host validates message shape and bounds input size. It rejects foreign
WebSocket origins and non-loopback hostnames. The configured listener is
loopback-only. This is not permission to expose it on a public interface.

Each connection has a bounded outgoing queue and one network writer. If the
queue fills, disconnect that client rather than blocking all clients or the
actor. Reconnect uses the same replay path. Dropped progress is not repaired by
inventing extra deltas; finalized content is repaired from History.

## 7.6 Render one kind of card well

Every message is an artifact: a stable card that may stream and then finalize.
Use one card map and vary rendering through type and role, not through separate
conversation models for chat, tools, and thinking.

- Human and assistant text use safe Markdown: headings, paragraphs, nested lists,
  tables, blockquotes, nested emphasis, links, inline code, and fenced code.
- Supplied thinking and reasoning summaries use the same readable prose renderer.
  Tool details remain expandable. Complete text stays in each backing card;
  provider signatures and opaque metadata are not readable thinking. Empty or
  whitespace-only thinking/summary cards occupy no visible space; reveal the same
  card when readable text arrives without deleting its backing data.
- Tool plans show arguments and lifecycle status. Results retain full readable
  output; terminal escape sequences do not become browser control syntax.
- Successful `edit_file` results can show the applied exact replacement from
  the admitted arguments. Label it accurately: it is not a whole-file diff.
- Search results keep their path/line output readable. Do not delete content
  merely to make a card short.

Use a pinned, locally bundled Markdown parser rather than growing a collection
of regular expressions. The current browser uses markdown-it 14.1.0 with its MIT
license and source integrity recorded alongside the vendor files. Do not add a
runtime CDN, fetch a parser on page load from another site, or silently enable
plugins. Tables and nested formatting are required; arbitrary HTML and every
possible Markdown extension are not.

Keep raw HTML disabled. Only HTTP(S) links are clickable; reject unsafe schemes
after parsing and normalization. Render Markdown images as escaped descriptions,
not automatically fetched remote resources. Explicit image attachments have their
own admission path. A filename, fetched document, or source fragment may contain
hostile markup. Fenced HTML must appear as code, not execute.

The parser's generated HTML may enter the message body only through this reviewed
policy. Never concatenate provider or tool text directly into `innerHTML`.
Continue using DOM nodes and `textContent` for controls, labels, tool output and
code highlighting. Test malicious markup, encoded unsafe links and remote image
syntax, including the absence of network requests. Library configuration is a
security boundary, not a substitute for these tests.

### Copy the logical code, not the preview

Give explicitly closed fenced code in assistant text a **Copy code** button.
Here, exact means the parser's logical code payload: omit the fences and Markdown
quote/list prefixes; retain code indentation, Unicode, blank lines, trailing spaces
and the payload's terminal newline. Accept the pinned parser's normalization to
LF line endings. The original Markdown remains unchanged in the backing card and
History. This is code copying, not a byte-for-byte export of the Markdown source.
Junien's [snippet contribution](https://github.com/waywardgeek/ensemble/pull/7)
informs this requirement and its clipboard-denial and literal-content tests.

An unfinished fence renders readable code but has no Copy button, even after the
response ends. Inline code, indented code, human text and reasoning cards have no
snippet-copy controls. Reuse the existing parser rather than add a competing
fence grammar. In the pinned parser, a fence token's source span includes its
opening line, every body line, and a closing line only if consumed. Container
prefix removal preserves body-line count. Count an unterminated final body line
too; end-of-input must not masquerade as a closer. Test this assumption against
empty blocks, nested quotes/lists, mismatched fences and indentation whenever the
parser changes.

Use an explicit button click and the clipboard write API. Capture the complete
logical payload at that click; neither highlighting nor a collapsed preview is
the source. The control stays outside long-code expansion, so copying a collapsed
snippet still copies all its code. Report success only after the write completes.
When the API is absent or denies access, show an honest failure status and a
labelled read-only field containing the same payload for manual selection/copy.
Do not read the clipboard, execute the snippet or submit it as a prompt.

Keep unchanged snippet controls when later prose streams. Retaining the node
preserves pending-copy state, success feedback, expansion and manual selection;
restore keyboard focus if reparenting blurs that same retained control. A changed
payload gets a new control. Late clipboard completion belongs to the old snippet
and must not update a replacement card or redirect its copied bytes. These are
browser interaction lifetimes, not additional persistent artifact ownership.

### Stream Markdown, render the accumulated text

Keep the complete original Markdown in the backing card. Send text deltas over
the existing connection; do not send a growing HTML document from the server.
Reparse only the changed card's accumulated Markdown in the browser. Coalesce
bursts of deltas with a bounded delay, currently 50 milliseconds, instead of
replacing the DOM for every token. The first visible card keeps arrival order;
a final event flushes the latest text promptly. This cadence is not a hard
real-time guarantee when the browser is busy or backgrounded.

Incomplete syntax can change interpretation: a table becomes a table after its
separator arrives, and a link becomes clickable when complete. Test prefixes of
real examples rather than inventing a second incremental Markdown grammar.
Retain speaker labels and playback controls while updating the body. Speech
consumes the original readable text through its existing path, never text scraped
back out of generated HTML.

Do not replace a card while a nonempty text selection intersects it, including
cards between the selection's endpoints. Retain the latest update and paint it
when the selection clears. Discard queued renders on replay or conversation
replacement so stale cards cannot reappear. Preserve the reader's scroll position
when away from the bottom; follow new output only when already near the bottom.

Tables use semantic header cells and a labelled, keyboard-scrollable wrapper so
wide content cannot widen the chat pane. Long code retains the complete source
behind its preview; preserve expansion and table scroll across updates. Measure
whole-message parsing and DOM replacement on realistic content before adding
incremental machinery. Desktop timing is not evidence of phone performance.

Display limits are acceptable only if the rest remains reachable. Collapsing
or scrolling is different from truncating the stored value. A speaker button
uses the full backing text, not the visible summary or the first screenful.

The layout must work at narrow widths. Controls wrap, code can scroll, and
focus remains visible. Geometry checks and browser DOM assertions are useful,
but they do not establish every person's visual comfort or screen-reader
experience.

## 7.7 Speech is part of steering

For someone who cannot comfortably read a busy transcript, speech is a control
surface, not ornamental animation.

Store narration enabled, rate, voice choice, and pause-while-speaking in one
origin-local browser preference, independent of the selected Agent. Same-origin
tabs share local-storage changes; separate profiles and devices keep independent
choices. Listening on a phone must not enable a laptop's narration. A fresh or
invalid preference defaults to speech off, rate1, default voice, and speech pausing
on. Unlike theme migration, never seed narration from a legacy server-wide enable
flag: an unattended browser must remain quiet. Server settings and replay cannot
overwrite this local preference, and speech controls send no settings patch.

Controls work offline. If storage is blocked, keep the choice for this page only;
removing the saved preference returns other tabs to silent defaults. Turning
narration off cancels that browser's current and queued speech. A rate or voice
change affects the next constructed utterance, not audio already speaking.
Microphone recording and transcription funding are separate from speech output.
The hold-to-talk shortcut is a browser-profile preference, not Agent policy;
Chapter8 specifies its local storage and Chapter25 implements its input guards.
Automatic narration reads only new content for the Agent selected in that tab.
A selected tab may be in the background; selection is not foreground visibility.


**Speak this artifact** is always available when the browser supports speech:

- human and assistant text: full readable text;
- thinking: full readable thought text, never signatures or opaque metadata;
- tool call: full name and formatted arguments;
- tool result: full readable result, not a collapsed preview.

Clicking the active artifact again stops it. Choosing another artifact cancels
the old queue and starts the new request. A global Stop speech button and Escape
even with a draft provide explicit escape paths. Speech failures are not
allowed to strand the Agent in a permanently queued-speech state.

**Automatic narration** is a separate opt-in preference. It narrates new
assistant text and thinking as complete clauses arrive, then flushes the final
remainder. It announces a concise tool-call summary. It does not automatically
read human messages, tool results, or reconnect history.

A single queue owns pending utterances and the current utterance. Both `onend`
and `onerror` settle the current item and advance the queue. A callback from a
cancelled utterance is ignored by identity. Queue state counts as busy before
the browser actually starts audio; relying on `speechSynthesis.speaking` alone
creates a pause race.

The earlier browser implementation recorded a Chrome visibility workaround:
when the page becomes visible, issue an empty zero-volume utterance through
`speechSynthesis.speak`. Preserve it as a local compatibility measure, outside
the narration queue. It is not a claim about every browser version, and it must
not itself cause a busy-speech pause.

A fake speech boundary can verify selected text, queue transitions, and error
cleanup. It cannot prove audible output, voice quality, or human hearing. Those
remain separate user exercises.

## 7.8 Pause scheduling, not the mailbox

Chapter 5 already supplies the right C# mechanism. Pause prevents **new** model
requests and tool dispatch, while the actor continues processing controls and
worker completions. Do not add a second GUI-specific gate or block the actor
thread while it waits for unpause.

A request already in flight may finish while paused. Its response is recorded;
a planned tool remains undispatched until unpause. A hint is accepted promptly.
Immediate interrupt still cancels the active turn without killing independent
jobs, and remains usable while paused. The two interrupt modes retain Chapter
5's different boundary semantics.

The browser tracks genuine human typing and current or queued speech. Whitespace
typed by the human counts; programmatic tool drafts do not pretend the human is
typing. Its local activity is:

```text
local cause = human typing OR (pause while speaking AND speech busy)
```

Only a change in that cause sends `pause` or `unpause`. The Agent mailbox applies
the last command to its single pause state. There are no device owners, vetoes,
or separate manual and automatic flags. Another device's resume takes effect
immediately; unchanged typing or speech must not repeatedly reassert pause.
Normal speech completion sends resume when neither local cause remains. Clearing
a draft while speech continues does not change the cause and sends no resume.

Use one **Stop speech** button beside the microphone in the composer, accessible
on phones. Keep both controls grouped when narrow layouts wrap. It uses
`type="button"`, never the form-submit default. Stop speech and Escape cancel
current and queued speech, invalidate late callbacks, and explicitly resume the
selected Agent even if this or another device still has a draft or speech.
Preserve the draft and establish its unchanged local activity as the edge baseline.
Escape must still allow native dialog dismissal. Neither action interrupts work
or cancels Jobs. Do not build a Pause/Play button pair.

A tab may send pause/resume only to its selected Agent, never an unselected child
or every Agent. Before a selection change, cancel outgoing speech and send any
outgoing resume while the old Agent is still selected. Late callbacks and replay
must not affect the incoming Agent. Disconnect sends no server-invented resume;
reconnect neither resumes blindly nor reasserts stale local activity. Any connected
device can use Stop speech to resume. A saved child retains the single admitted
pause state across restart; there is no transient device-hold state to serialize.

Enter submits a prompt; Shift+Enter inserts a newline. Ctrl+Enter or Alt+H sends
a hint. Submission or clearing updates the typing cause. Stop speech does not
interrupt Agent work, cancel Jobs, or disable future opt-in narration.

## 7.9 Build and inspect locally

Build the optional project separately from the core:

```sh
dotnet build Examples/Gui/Gui.csproj
```

Run it from an owned workspace, selecting a provider and model explicitly:

```sh
dotnet run --project /path/to/Examples/Gui -- \
  --port 8097 --workspace /path/to/owned-work \
  --vendor anthropic --model SELECTED-MODEL --base-url PROVIDER-URL \
  --event-log event.log
```

Credentials come from the existing `LLM_API_KEY` environment contract, not from
page content or command-line arguments. `LLM_VENDOR`, `LLM_MODEL`, and
`LLM_BASE_URL` provide defaults for their flags. `--no-stream` disables streaming;
unknown stream capability still follows Chapter 6's plain fallback. `--gui-dir`
selects static assets when not using the copied build output.

Keep four kinds of evidence distinct:

| Log | Contents |
|---|---|
| `event.log` / `--event-log` / `CH02_LOG` | canonical recorded facts |
| `api.log` / `CH06_API_LOG` | existing sanitized request/payload trace |
| `gui.log` / `CH07_GUI_LOG` | UTC timestamp, direction, and GUI JSON frame |
| `debug.log` | ordinary application diagnostics |

GUI tracing writes `>` for received and `<` for sent frames. API and GUI traces
redact the configured credential. They are still private working artifacts:
conversation and tool output can contain other sensitive information. Do not
publish them merely because one key was redacted.

Use an owned port, workspace, browser profile, and fake provider while developing.
Do not borrow another running Agent or change someone's connected browser.
Closing the browser leaves work running. Stopping the host explicitly tears
down connections, stops observation, and lets the composition root destroy the
Agent graph. The existing cooperative-callback shutdown limitation still applies.

## 7.10 Verification and taking it for a spin

Test real components with small external fakes. Important cases include:

1. Round-trip finalized display IDs, including zero and noncontiguous IDs;
   reject malformed metadata and retain historical-log compatibility.
2. Read snapshots during concurrent appends and verify one coherent frontier.
3. Refuse malformed and policy-disallowed tool calls without a started fact;
   admit a valid tool and inspect its real effect and recorded result.
4. Gate a provider after a meaningful delta. Observe it in the browser before
   releasing completion, then reconnect during the partial response and after
   finalization. IDs must remain stable and cards must not duplicate.
5. Pause an active turn before a tool plan completes. Verify that no tool effect
   occurs until unpause, while hints and interrupts still work.
6. Verify a fresh prompt works after interruption. Reject a foreign WebSocket
   origin without granting it actor input capability.
7. Inspect actual narrow-layout controls and safe rendering of hostile-looking
   code. Exercise full-content speaker buttons through a fake speech boundary.
8. Compare copied bytes for complete fences, including Unicode, CRLF input,
   quoted/list-contained code, empty blocks and a long collapsed preview. Refuse
   unfinished fences; test denied/missing clipboard access, manual selection,
   focus during streaming and completion after the original card disappears.
   A deliberately preview-only copy and an unfinished-fence copy must fail.
9. Exercise queued speech, interrupted-utterance cleanup, automatic typing pause,
   cross-tab resume, Stop speech/Escape preserving a draft, disconnect/reconnect,
   selection cleanup, transient restart state, and silent visibility wakeup.

The existing executable grader can test the optional server without compiling
Go consumers:

```sh
grade-edition2 -ch 7 -executable-only \
  -gui /absolute/path/to/wwwroot -gui-model SELECTED-STREAM-MODEL \
  /absolute/path/to/gui-launcher
```

This route covers 80 points: wire behavior, reconnect replay, GUI trace, and
pause semantics. It excludes inherited Chapter 6 scoring and Go topology
checks. The full historical Go route remains available and unchanged in scope.
Neither route establishes accessible speech quality or real provider acceptance.
These historical points do not certify the added snippet-copy obligations.
Focused browser tests and the forthcoming named Artifacts grader cover that
contract separately; student reconstruction and human usability remain distinct.

For a human exercise, use a bounded task in the owned workspace. Watch a
response stream, inspect thinking and tool details, send a hint, pause and
unpause, and interrupt while paused. Reconnect while work continues. Listen to
individual full artifacts and try automatic narration if authorized. Keep the
actual logs and describe failures honestly.

The C# endpoint's offline protocol, browser, and fake-speech evidence is recorded
in `docs/edition-3-csharp-progress.md`. Paid-provider work and audible-user
verification must remain explicitly incomplete when they have not been done.
A green grade is evidence for its checks, not a replacement for these boundaries
or for an inspection of the ownership and call paths.
