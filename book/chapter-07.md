# Chapter 7: Everything Is an Artifact

The terminal from Chapter 6 is honest and ugly: three streams in three
colours, wrapping at terminal width, no memory of what scrolled past.
A tool the human cannot comfortably watch is a tool the human cannot
steer, and steering is the entire safety argument. This chapter builds
one reusable component, wires it to the agent through a WebSocket, and
the agent never learns the GUI exists.

---

## TL;DR

The agent writes observations. The GUI reads them. The core has no GUI
import, dependency or display configuration. Chapter 5's public observation
and input seams connect the two. GUI and WebSocket implementation live in a
separate optional Go module; the application constructs one Ensemble and one
Agent for both terminal and browser.

### The observer, extended

Two tool-boundary observation types, plus turn-end display progress:

```go
// Shared vocabulary in internal/common; public aliases expose it to clients.
// Keep observation names distinct from mailbox reports and stored events.

// ToolDispatched reports the boundary before handler execution.
type ToolDispatched struct {
    // CallID correlates execution with its requested call.
    CallID string          `json:"call_id"`
    // Name identifies the selected tool.
    Name   string          `json:"name"`
    // Input preserves raw arguments without interpreting them.
    Input  json.RawMessage `json:"input"`
}

// ToolFinished reports a wait result, including late reports after interrupt.
type ToolFinished struct {
    // CallID identifies the call that returned this result.
    CallID  string  `json:"call_id"`
    // Result is bounded text; full job output remains recoverable.
    Result  string  `json:"result"`
    // IsError distinguishes failed tool output.
    IsError bool    `json:"is_error"`
}

// TurnEnded is progress, not a reliable request acknowledgement.
type TurnEnded struct {
    // Text is the final reply, excluding intermediate tool narration.
    Text  string `json:"text"`
    // Error is a safe diagnostic for a failed or interrupted turn.
    Error string `json:"error,omitempty"`
}

func (ToolDispatched) observation() {}
func (ToolFinished) observation()   {}
func (TurnEnded) observation()      {}
```

`ToolDispatched` fires before execution. `ToolFinished` reports the returned
tool wait result; a job can still be running after that wait. Preserve later
real job reports. Together with `PartDelta`, `PartFinal` and `StateChanged`,
these observations show the turn's progress. Ensemble identifies the source
Agent when notifying clients; tools do not acquire a copied Agent identity
or a GUI parent. Reliable request replies remain separate from display progress.

### The wire

A WebSocket server subscribes through Ensemble's public identified observer
interface and fans out JSON to connected clients. Durable content comes from
the Agent's history; transient deltas and state observations complete the view.
These are examples of individual envelopes, not one complete conversation:

```json
{"type":"part_delta","part_id":3,"kind":"text","chunk":"Hello"}
{"type":"state_changed","from":"input_pending","to":"in_flight"}
{"type":"tool_dispatched","call_id":"tc_1","name":"read_file","input":{"path":"main.go"}}
{"type":"tool_finished","call_id":"tc_1","result":"package main...","is_error":false}
{"type":"part_final","part_id":3,"part":{"type":"text","text":"The file contains..."}}
{"type":"turn_ended","text":"The file contains..."}
```

Client messages:

```json
{"type":"subscribe"}
{"type":"prompt","text":"Run the tests"}
{"type":"hint","text":"Skip the slow ones"}
{"type":"interrupt"}
{"type":"pause"}
{"type":"unpause"}
```

### Reconnection

A client subscribes. The server sends two things: a window of recent
events from the event log (the last hour or last 100 renderable
events, whichever is larger), and the accumulated partial content for
anything currently streaming. The client renders the event history as
finalized Artifacts, replays the partial content to catch up to the
live stream position, then switches to live delivery. A fresh
connection and a reconnection after an hour use the same path. The
agent does not pause, restart, or notice.

No client cursor or client sequence numbers. The event log has its own ordering.
On each live observation, the hub checks for new entries since its captured
Seq and sends their renderable content alongside the streaming content. That
internal frontier is not a cursor supplied by the browser. The client never
parses SSE frames. Content messages carry an artifact or call identity and a
payload ready to render; status messages carry the effective state.

### The Artifact

Everything the agent produces is an Artifact: a widget that streams,
then finalizes.

| artifact | streaming format | finalized rendering |
|---|---|---|
| thinking | markdown (dim) | collapsible |
| chat | markdown | the response |
| tool call params | JSON (parameters as they arrive) | name + abbreviated args |
| tool result | n/a (not streamed) | type-specific: diff for edit_file, results for search_files |
| user message | n/a | the prompt |

One component, `ArtifactScroll`, manages the list. It receives
WebSocket messages, creates or updates Artifacts by display identity, streams
chunks through the format-appropriate renderer, and replaces them
with finalized HTML when the part completes. Different visual
treatment comes from CSS classes, not separate component types.

Two rules govern what a card is allowed to do with that text, and both
were broken in the first edition's implementation until its Chapter 13 audit.

Build the card with `textContent`. Never assemble it by interpolating
values into `innerHTML`. A tool call's name and arguments carry text the
agent did not write: a filename, a fetched URL, the contents of a file
it just read. Markup arriving inside any of those executes in the page.
This is the prompt-injection surface from the security chapter, reaching
the user through the renderer rather than through the model, and it is
easy to miss because the insecure version is shorter and reads better.

Cap what a card displays, and keep the remainder reachable. A cap with
an ellipsis and no affordance deletes the rest permanently for a reader
who cannot scroll past it. Put the full value in a `title` attribute or
behind an expander, so the limit governs the display instead of the
information.

### TTS

Two channels. **Auto-speak** fires during streaming: full text for
thinking and chat, an abbreviated summary for tool calls ("read file:
main.go, line 42"), silence for tool results. **On-demand** via a
speaker icon on every Artifact card, speaking the full accessible
text on click.

The original Chrome exercise needed a speech-service wakeup after tab switches:
a zero-volume empty utterance on `visibilitychange`. Retain and exercise that
workaround in the browser being used; it is a recorded compatibility behavior,
not a claim that every browser version fails identically.

### Pause

```
local_pause_cause = tts_queued_or_speaking OR user_typing
```

The Agent checks its gate before starting each tool; the server calls the
public Agent's `Pause` or `Unpause` on receiving those commands. Pause is
Agent-wide. An explicit unpause in any tab unpauses it for everyone; unchanged
local typing or speech in another tab cannot veto that decision. Rules:

1. TTS queues an utterance: send `pause`. Set the flag when the
   utterance is queued, not when audio begins.
2. TTS queue empties: send `unpause`, unless the user is typing.
3. User starts typing in the input: send `pause`, even with TTS off.
4. User sends (Enter): re-evaluate. Stay paused if TTS is speaking.
5. ESC with empty input: cancel TTS, re-evaluate.
6. Streaming into an already-running tool continues. Pause gates new
   starts only.
7. `onerror` on utterances must mirror `onend`. A Chrome `interrupted`
   error that does not continue the queue deadlocks it.
8. Edge-triggered: send pause/unpause on transitions, not on every
   utterance boundary.

### gui.log

The fourth log. Every WebSocket message in both directions,
timestamped. One line per message.

| log | captures | added in |
|---|---|---|
| event.log | agent events (append-only) | earlier chapters |
| api.log | LLM wire JSON | earlier chapters |
| debug.log | diagnostics | earlier chapters |
| **gui.log** | **WebSocket JSON, both directions** | **this chapter** |

### Yours

Markdown rendering library. ANSI-to-HTML approach. Visual styling.
TTS voice and speed. WebSocket library. Whether finalized Artifacts
replace or augment the streamed view. How the speaker icon looks.

### The exercise

Build an optional public-library GUI module: one Agent, one HTTP server on
`--port` (default 8088), serving its static files through `--gui-dir`. The
accepted example lives in `gui/`, with `cmd/` and `web/gui/` inside it; historical
chapter IDs in evidence remain `ch08`. The binary accepts prompts
over WebSocket, streams observations to all connected clients,
supports reconnection, gates tool calls on pause, and logs to gui.log.

Configuration uses public flags, including `--port`, `--gui-dir`, `--vendor`,
`--model`, `--base-url`, `--workspace`, `--event-log` and `--no-stream`.
Explicit flags override the inherited provider environment defaults. Keep
credentials in the intended process environment rather than command arguments.
The original flags-only sentence conflicted with its own environment-configured
grader; it did not justify breaking the earlier key setup.

```
go run ./cmd/grade-edition2 -ch 7 path/to/core
```

The grader builds the GUI module separately: `-gui DIR` selects it, defaulting
to `gui/` under the core. It builds `./cmd` when present, otherwise the module
root, and launches `--port` plus `--gui-dir` pointing to that module's `web/gui`.
HTTP `/` must become available and WebSocket `/ws` accepts the messages above.
Its faithful provider peer supplies the inherited `LLM_VENDOR`, `LLM_MODEL`,
`LLM_API_KEY` and `LLM_BASE_URL` environment. `-gui-model MODEL` selects the
declared Anthropic fixture row; the historical default `claude-sonnet-5` is a
local test-table entry, not a model-discovery claim. Earlier consumer, workflow,
media and streaming flags still apply to core parity.

The optional module must resolve its public core dependency to the submitted
core. This boundary check accompanies `websocket-streams`; source review still
establishes actual ownership and shared-Agent use. Active-partial replay,
handoff ordering, the fixed window, paused interrupt and browser behavior also
retain their source/integration checks. A WebSocket score alone cannot prove
what a card displays or what the listener hears.

| check | points | what it tests |
|---|---|---|
| `websocket-streams` | 25 | subscribe, prompt via WS; receive part_delta, part_final, state_changed, turn_ended, tool_dispatched, tool_finished with correct fields |
| `event-replay` | 20 | reconnect after a completed turn; recover the same final/tool content and artifact IDs, without stale partials |
| `gui-log` | 15 | gui.log contains timestamp, direction and JSON per line; both directions present |
| `pause-holds-tools` | 20 | between real tool calls, pause prevents the next start; explicit unpause from another tab releases it; all calls complete |
| `ch6-parity` | 20 | every Chapter 6 check still passes against the headless core |
| **total** | **100** | |

---

## 7.1 The idea in plain words

A radio station transmits regardless of how many receivers are tuned
in. Add a radio and it hears from that moment forward. Unplug one and
the station does not pause. Play the recording from an earlier
timestamp and hear what you missed.

The agent is the station. Observations are the broadcast: streaming
chunks, finalized parts, state changes, tool starts, tool completions.
The GUI is a radio. The event log is the recording; an observation buffer
alone would lose completed history when progress is dropped.

**The agent has no GUI code.** No WebSocket import, no rendering
dependency, no display flag in its configuration. The agent writes
observations through Ensemble's public seam from Chapter 5. An optional Go
WebSocket server subscribes there and projects history plus progress into
browser messages. Attach zero browsers and
the agent runs identically. Attach three and they all see the same
stream. Close a tab, reopen it an hour later, subscribe, and catch up from the event log in one burst.

**Everything is an Artifact.** Thinking text, chat text, tool call
arguments, tool results, user messages: they differ in how they look,
not in what they are. Each one streams in one of three formats
(markdown, ANSI, JSON), then optionally renders finalized HTML
specific to its type. One component, `ArtifactScroll`, manages the
scroll view. CSS classes make thinking dim and tool calls compact. The
renderer per type is a detail inside the component, not a separate
architecture.

**Pace is a safety mechanism, not a preference.** A human who cannot
watch the agent comfortably cannot steer it. The hint, the interrupt,
the decision to let a tool call proceed: all require seeing what the
agent is doing before it finishes doing it. TTS auto-speaks the stream
so the human hears without looking. Pause gates tool calls while the
human absorbs what just happened. The chapter after this one builds a
full three-pane workbench; this one builds the component it sits on.

---

## 7.2 Extending the broadcast

Chapter 5 built `PartDelta`, `PartFinal` and `StateChanged`;
Chapter 6 made the content incremental. They carry response content and
stored tool-result finals, but there is no explicit tool-start announcement.
The model requests a tool call; a display should be able to distinguish
execution starting from a result arriving later.

`ToolDispatched` fires before execution: call id, tool name, raw
input. `ToolFinished` carries the result of waiting: call id, result, error
flag. Chapter 4's distinction still holds. A timeout can return a handle while
the actual job continues, and its later output must remain visible even if
the human has interrupted the turn. `TurnEnded` announces the end of the turn
to displays; it does not replace the reliable completion used by `Ask`.

The names avoid a silent collision. The event log already has
`tool_called` and `tool_returned` for replay. The mailbox has its
internal worker reports to the actor. Three
systems, three names for what looks like the same thing: the event
is the record, the mailbox message is the signal, the observation
is the broadcast. Using the same name for two of them compiles fine
and produces a conversation about why something fired twice.

---

## 7.3 One tool at a time

The actor dispatches tools serially: start one, wait for its reported result,
start the next. A continuing job may outlive that wait. It has worked this
way since Chapter 5. What changes
is that the property now matters.

A pause gate that checks before each dispatch needs a gap between
dispatches in which to check. Parallel dispatch closes that gap:
three tools fire at once and a pause arriving a millisecond later has
nothing to stop. Serial dispatch creates the gap. The gate fills it.

The comment in the actor loop states the dependency rather than what
the code does:

```go
// Serial dispatch lets the pause gate hold execution between tools.
```

A comment that states *why* an architectural choice was made survives
the instinct to parallelize that hits every engineer who reads a
serial loop.

---

## 7.4 The pause gate

Pause is a live Agent control, not another conversation message waiting for
reduction. The Agent owns its gate. Common declares the public interface;
the behavior belongs with Agent, rather than in a shared utility package.
The GUI calls that Agent interface. It does not receive an Engine sibling,
become the Agent's parent, or create a second gate.

The accepted implementation uses a mutex and a wake channel. While paused,
a small boundary waiter waits for unpause or cancellation and reports back to
the actor. The actor keeps consuming its mailbox. Blocking that consumer would
strand exactly the interrupt the human is trying to send.

The wake reason is the design. Unpause permits the next dispatch; cancellation
ends the current turn. The Agent stays alive, another prompt can start, and
other jobs keep their real outcomes. Without context awareness, pause is a
trap: the human would have to remember to unpause before interrupting a turn,
under exactly the pressure where remembering extra steps fails.

The gate bypasses the mailbox deliberately. A mailbox message waits
in FIFO order behind pending tool completions, hints, and interrupts.
Pause is about the next dispatch boundary, and the Agent-owned gate has no
mailbox queue to wait in. A tool already admitted past that boundary continues;
pause is neither retroactive cancellation nor process shutdown.

---

## 7.5 The receiver

The hub implements the public identified observer. Its callback handles two
tiers: it captures new durable history and accumulates streaming content in an
in-flight map keyed by display identity. Live deltas send only the new chunk;
a reconnect receives the accumulated prefix. Then it returns. If a client's send
channel is full, the message is dropped for that client. The actor
does not wait for network I/O.

Each client has a write goroutine draining a buffered channel.
The observation callback iterates the set and does a non-blocking send on each.
The hub is a broadcaster: the station does not wait for the
radio.

Prompts and hints from the browser use the same public `Post` entrance as
terminal input. The actor classifies a busy user message as a hint; the browser
does not guess which turn should receive it. Blocking library callers still use
reliable request completion. Neither view runs another model loop.

The accepted composition root owns Agents, configuration stays with Agent,
and Engine reaches it through its actual parent. The transport owns sockets.
Public interfaces let it reach the existing Agent without making a socket
server a fictional parent or copying services into another bag.

Network sends and their trace writes happen in the connection's writer;
incoming messages are traced by its reader. Both stay outside the observation
callback. Disconnect uses a separate lifetime signal, and
shutdown joins writers before closing logs. A nonblocking send does not make
it safe to close a data channel another broadcaster can still reach.

---

## 7.6 The wire

The protocol has no sequence numbers and no cursor. A `subscribe`
message arrives empty. The server decides what history to send based on its
own fixed event-log window (last hour or last 100 renderable
events, whichever is larger), then sends any in-flight streaming
content, then switches to live delivery.

A reconnecting client and a fresh client use the same path. The
server does not need to know whether this is a first connection or a
tenth. The event-log window is the same either way. This is simpler
than tracking per-connection offsets, and better:
the client gets recent context whenever the Agent has any to give.

A subscription also receives current turn and pause state, marked as replay.
The turn snapshot uses the existing `state_changed` envelope with both `from`
and `to` equal to the current state. Equal endpoints report a snapshot without
inventing a transition. The student initially omitted `from`; that was a real
contract defect, not an optional-field convention. A fresh-subscription test
and a restored-defect mutation now check the complete envelope.

---

## 7.7 Two tiers of state

The hub holds two kinds of state, matching the two things a
connecting client needs.

**Tier 1: The event log.** Recorded facts from the append-only log:
finalized parts, tool calls and returns, user messages. On subscribe, the hub reads the recent
window and converts each event to a wire message the
client renders directly. They survive a browser disconnect while the Agent
continues. Restoring an Agent after a process restart belongs to the later
save/load chapter; writing a log does not implement restoration.

**Tier 2: In-flight partials.** Accumulated streaming content for
anything the agent is currently producing. A map from display identity to
the concatenated chunks received so far. On subscribe, the hub sends
the current partial for each active part. On `PartFinal`, the
partial is cleared. These are ephemeral: they exist only while
content is streaming.

Append-only is a history rule, not a synchronization primitive. Appending can
replace a Go slice's backing array while another goroutine reads it. The
history owner snapshots the requested suffix under its lock. The GUI then
coordinates that history frontier, its partial map and subscriber admission
under one handoff lock. A final cannot slip between a subscriber's completed
history and its old partial, leaving the same text on two cards.

The parser IDs from Chapter 6 are still local to one model response. The GUI
must retain cards across several responses and human turns. The actor assigns
an Agent-lifetime display ID to each part, uses it in root observations, and
stores it on the final part. Replay therefore names the same card that received
the live chunks. Provider call IDs and opaque signed content keep their original
meanings. A browser identity is not permission to rewrite provider history.

When a final is captured, remove its partial before publishing it. Sort active
partials by their assigned creation identities on replay; a Go map supplies no
stream order. On an interrupted turn, clear abandoned partials that will never
receive a final. Completed facts still come from history, including actual late
job output.

Replay and live progress need different admission rules. A live progress update
can be missed by a slow receiver, but silently dropping half its reconnect
history defeats the recovery path. The accepted server admits the whole replay
as one queued batch while adding the subscriber, then lets that connection's
writer send it outside the lock. A 270-event replay test exceeds the live queue's
256-batch capacity. A truncation mutant initially timed out without proving why;
after preserving terminal status, it failed directly at 249 versus 270 facts.

Another mutation removed final-part cleanup but passed: the test reconnected
after `TurnEnded`, which had already cleared every partial. Reconnecting while
tools were still running exposed the stale partial. The assertion had to observe
the interval in which the two cleanup paths differed.

For this chapter the window is fixed: include the last hour OR the last 100
renderable events, whichever reaches farther back. An event with several
rendered parts still counts as one event. A configurable policy is future work.

---

## 7.8 The fourth log

```
2026-09-17T14:32:01.123Z > {"type":"subscribe"}
2026-09-17T14:32:01.456Z < {"type":"part_delta","part_id":3,"kind":"text","chunk":"Hello"}
```

`>` is client-to-server. `<` is server-to-client. Each line has an RFC3339
timestamp, a direction marker and one compact JSON payload. The whole line
is not itself a JSON object. The optional transport owns this trace; general
diagnostics still belong to Ensemble's logger. This is
transport-layer tracing, the same tool as `api.log` for a different
wire.

A log missing one direction lies about what happened. The lie
surfaces at the worst moment: the message that was not delivered,
invisible in the record because the record only shows one side.

---

## 7.9 Everything is an Artifact

An Artifact streams, then finalizes. Thinking, chat, tool
arguments, tool results, user messages: all Artifacts. They differ
in CSS class and renderer, not in kind.

`ArtifactScroll` manages them. On `part_delta`: find or create by
part id, stream through the format renderer (markdown for thinking
and chat, JSON for tool arguments). On `part_final`: replace with
finalized HTML. On `tool_dispatched`: create a tool card. On
`tool_finished`: update it.

The alternative, ChatMessage plus ThinkingPanel plus ToolCallCard, encodes
assumptions about what the agent produces. An Artifact that knows
three rendering formats handles anything that fits one, which is
everything, because LLMs produce text in exactly the three varieties
the terminal already distinguished by colour.

Visual treatment comes from CSS classes: `artifact--thinking` gets
dim opacity, `artifact--tool` gets a compact layout. Adding a new treatment is a CSS rule.

The Go grader cannot establish safe DOM rendering or usable speech. Exercise
the actual page as well. An exercise whose output is invisible is an exercise
nobody finishes.

The student's live Gemini page revealed a less dramatic rendering error: a
legitimate empty text part displayed its own metadata. The renderer used a
truthiness test, then fell back to JSON. An empty string lost to the fallback.
The same shortcut would display opaque signatures and redacted thinking as if
they were words for the reader. Since the speaker button and automatic speech
used the card's text, the error crossed both channels.

The correction projects readable content explicitly. Empty text stays empty.
Readable Anthropic thinking and Gemini thought text can become prose; signature
and redacted bodies remain opaque replay material. Preserve those bytes in
history without pretending they are something to read aloud.

The focused browser test reused the retained Gemini part, plus readable and
opaque controls from an external peer. It checked the card, automatic speech
input and speaker-button input separately. Restoring the fallback failed all
three checks for each empty or opaque case, while readable controls still
produced native start/end events. No new model generation was needed to diagnose
what was already visible in the saved page.

### Copyable text snippets

A reply can explain a command and offer the command itself for copying.
Completed Markdown code fences in text replies become read-only textboxes,
each with its own **Copy** button. Ordinary backtick fences work with or
without a language label; no special snippet marker is required. The boxes
sit inside the reply, alongside its explanation. There is no whole-Artifact
Copy button: each button copies only its snippet.

The textbox holds the literal payload between the enclosing fence lines.
Copy excludes those fences, their language label and the surrounding prose,
while preserving the payload's whitespace, blank lines and Unicode. Assign
the payload as the textbox's value so HTML-looking text stays text, including
a literal closing textarea tag. To include triple-backtick fences in the
payload, enclose it with four backticks. More generally, use an outer fence
longer than any backtick fence the snippet contains.

While a reply streams, an unfinished fenced block remains noninteractive
preformatted text. Its closing fence makes the snippet copyable; later prose
can keep streaming around it. The authoritative final replaces provisional
content, and replay renders the same snippet controls from the retained text.
For long replies, the truncated preview has no snippet controls. Open
**Full artifact** to reach the complete Markdown and copy its snippets.

Copy first uses the native clipboard API, then falls back to the browser's
copy command if the API is unavailable or denied. If both fail, the textbox
is selected and the page asks the reader to press Ctrl+C or Command+C.
This is GUI rendering over ordinary reply text; it needs no provider or
core changes.

---

## 7.10 The voice

TTS auto-speaks the stream. Thinking and chat in full. Tool
dispatches abbreviated: "read file: main.go, line 42." Tool results
silent. A speaker icon on every Artifact card speaks the accessible
text on click.

The original Chrome exercise found `speechSynthesis` asleep after a tab switch.
Its workaround sends a zero-volume empty utterance on `visibilitychange`.
Ugly, and worth recording where the next person building the queue can find it.
The accepted page keeps this wakeup outside the audible queue, so it does not
hold the pause gate.

`onerror` must mirror `onend`. Chrome fires `error` with reason
`interrupted` when one utterance cancels another, which is normal queue
behaviour. An error handler that does not advance the queue deadlocks
it: silence, permanent, no error message. Both callbacks mean "this
utterance is done, advance."

Queue phrases from the streamed text, not an utterance for every network chunk.
Keep unfinished phrases by artifact identity. A final flushes the remainder;
it must not read the whole answer again. Replayed history is silent until the
reader clicks its speaker. Keep a reference to the active utterance, and ignore
an old cancellation callback after a newer utterance has taken its place.

The native-browser test recorded a first utterance starting, a real
`speechSynthesis.cancel()` interruption, then the second utterance starting and
ending with pause released. A restored error-handler defect stranded the queue.
That proves the browser received and completed speech work. It does not prove
that a person heard an intelligible voice; listen as a separate part of using
the interface.

---

## 7.11 Pause in the browser

```
local_pause_cause = tts_queued_or_speaking OR user_typing
```

TTS queues an utterance: send `pause`. Queue empties: send `unpause`,
unless the input has text. User types: `pause`. User sends: stay
paused if TTS is speaking. ESC with empty input: cancel TTS,
re-evaluate.

Edge-triggered. Send `pause` on the transition, send `unpause` on
the reverse. The WebSocket carries transitions; the server holds state.
A client sending `pause` per queued utterance floods the wire with
messages that carry no information. More subtly, it can undo another tab's
explicit unpause. Local causes only send automatic commands when they change.
The **Unpause everyone** button releases the one Agent gate immediately.
An unchanged cause elsewhere cannot assert a veto, and reconnecting a tab
must not reassert its stale local pause state.

Streaming into a running tool continues. The gate is between tools,
not inside them.

---

## Taking it for a spin

Build from the optional GUI module's own directory:

```sh
go build -o /tmp/ch7gui ./cmd
```

Use the inherited provider setup: `LLM_VENDOR`, a discovered `LLM_MODEL`,
`LLM_API_KEY` in the intended child environment, and `LLM_BASE_URL` when an
endpoint override is needed. Flags override those defaults. Keep credentials
out of prompts, command arguments and logs. From a disposable working directory,
launch the binary with the actual absolute asset path:

```sh
/tmp/ch7gui --port 8088 --gui-dir /absolute/path/to/gui/web/gui
```

The working directory receives `event.log`, `api.log`, `debug.log` and
`gui.log`; the default tool workspace is that directory too. `--workspace`
can select another scratch project. The HTTP server listens on localhost and
serves the WebSocket at `/ws`. Its terminal remains another view of the same
Agent; terminal EOF detaches that input without shutting down the browser.

Open `http://localhost:8088`. Dark page, input field, "idle." Type a
prompt. Reasoning appears dim, the reply streams, TTS reads it. Ask
for a tool and a card appears with the arguments, the result fills
in, the reply continues.

Open a second tab. Both show the same stream. Close one, reopen:
it subscribes, receives the event-log window, catches up in a burst. Type a hint in the terminal; both tabs see
it. Start typing in the browser: "paused," no new tool starts.
Type `/interrupt` in the terminal: the current turn ends even while paused.
The Agent remains usable and running jobs continue. An explicit unpause from
the second tab must release the gate for both, even if the first tab still
contains unsent text.

Terminal and browser, two views of one agent, neither aware of the
other.

The retained October 9 exercise used installed Chrome and the real native
speech service. It checked two tabs, a connection during an active partial,
completed-history reload, safe markup, reachable long content, typing with
Read aloud off, another tab's unpause, speaker clicks, Escape and visibility
changes. [The browser receipt](../solutions/evidence/ch08/browser.json)
records 180 available voices and a short sentence's queued/start/end events.
That is an API observation, not a subjective listening result.

The speech check did not pass cleanly on the first try. Its first multi-tab
record contained a native `canceled` error. A later repeat waited for speech
that did not complete: reloading the second tab had restored its checked
Read aloud default, accidentally enabling overlapping speech during the long
response. Explicitly disabling that tab's voice after reload produced the
clean automatic-speech result. A separate
[interruption receipt](../solutions/evidence/ch08/speech.json)
checks that a real error advances the queue. Neither result justifies
inventing an audio-quality claim.

All three live provider turns ran an actual `run_command` and returned its
marker with exit code 0:

| Historical model used | Observed deltas | Actual tool marker |
| --- | ---: | --- |
| `claude-sonnet-4-6` | 28 | `GUI_ANTHROPIC_OK` |
| `gpt-4o-mini` | 27 | `GUI_OPENAI_OK` |
| `gemini-3.8-flash` | 5 | `GUI_GEMINI_OK` |

The Anthropic run also carried a terminal busy hint into the browser on the
same Agent, and its speaker button produced native start/end events. The
Gemini run used Bill's required model and also exposed the empty-text metadata
card described above. Successful tool execution did not make the page correct.
The [live summary](../solutions/evidence/ch08/live-summary.json)
and [projection correction](../solutions/evidence/ch08/projection.json)
retain both the success and the subsequent repair. These are observed model
destinations, not a current availability list or a new paid run of this edition.

The original student GUI grader returned **55/100 FAIL** after roughly
47 minutes. It had built an earlier binary, before the replay-status fix:
the missing `from` field cost 25 points, while inherited CLI checks launched
the continuing GUI server and cost another 20. The original legacy solution
scored 100/100. Preserve those results rather than assigning a hypothetical
score to the corrected source. Final-source race tests, browser checks and
independent review established the bounded fixes. The revised grader now
uses the actual separate GUI and headless entry points; its detailed evidence
belongs in the [chapter review](../docs/edition-2-notes/enhancement-07-code.md).

---

## What Chapter 8 does with this

One scroll pane, one input field, one dark page. Chapter 8 adds the
workspace: a second `ArtifactScroll` for tool calls and results, an
agent tree, drag bars, settings, and a sidebar that grows as chapters
add features. The Artifact is the reusable unit. The layout is what
remains.
