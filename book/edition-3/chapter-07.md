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
- Typing and queued speech cause **edge-triggered** pause/unpause requests.
  Pause is Agent-wide. An unchanged local condition is not a tab's permanent
  veto over an explicit unpause from elsewhere.
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
The browser's plan card can therefore change from **planned** to **started**
to **finished/error**, without portraying a refusal as real execution.

`tool_finished` means the foreground observation returned. A running-job report
can be a legitimate result: waiting stopped, while the independent job remains
alive. Preserve the job handle and locator. Do not relabel that observation as
process termination.

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
it does not start a new Agent or replay work. The wire spelling is **unpause**,
although the core operation is `InputKind.Resume` and the button says “Resume.”

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

- Human and assistant text use safe Markdown: headings, paragraphs, lists,
  links, inline code, and fenced code.
- Thinking and tool details are collapsible. Their complete text remains in
  the backing card and a scrollable details view.
- Tool plans show arguments and lifecycle status. Results retain full readable
  output; terminal escape sequences do not become browser control syntax.
- Successful `edit_file` results can show the applied exact replacement from
  the admitted arguments. Label it accurately: it is not a whole-file diff.
- Search results keep their path/line output readable. Do not delete content
  merely to make a card short.

Build DOM nodes and assign `textContent`; never interpolate provider or tool
text into `innerHTML`. A filename, fetched document, or source fragment may
contain hostile markup. Fenced HTML must appear as code, not execute.

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

**Speak this artifact** is always available when the browser supports speech:

- human and assistant text: full readable text;
- thinking: full readable thought text, never signatures or opaque metadata;
- tool call: full name and formatted arguments;
- tool result: full readable result, not a collapsed preview.

Clicking the active artifact again stops it. Choosing another artifact cancels
the old queue and starts the new request. A global Stop speech button and Escape
from an empty input provide explicit escape paths. Speech failures are not
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

The browser has a local cause:

```text
input is nonempty OR speech is active OR speech is queued
```

Only its transitions send controls:

- false → true: send `pause`;
- true → false: send `unpause`.

The server's effective state is authoritative. Another tab or application caller
can explicitly unpause even while this tab is still typing. Unchanged typing
must not immediately reassert pause on every key or state frame. Only a later
local cause transition sends another request.

This is why the model is not a distributed set of tab-owned veto tokens. The
human needs a dependable explicit escape, not a hidden AND of every browser's
local conditions.

Enter submits a prompt; Shift+Enter inserts a newline. Ctrl+Enter or Alt+H sends
a hint. Input clearing updates the local pause cause. Buttons provide the same
operations without requiring keyboard shortcuts.

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
8. Exercise queued speech, interrupted-utterance cleanup, automatic typing pause,
   cross-tab explicit unpause, and silent visibility wakeup.

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
