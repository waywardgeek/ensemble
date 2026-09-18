# Chapter 8: Everything Is an Artifact

The terminal from Chapter 7 is honest and ugly: three streams in three
colours, wrapping at terminal width, no memory of what scrolled past.
A tool the human cannot comfortably watch is a tool the human cannot
steer, and steering is the entire safety argument. This chapter builds
one reusable component, wires it to the agent through a WebSocket, and
the agent never learns the GUI exists.

---

## TL;DR

The agent writes observations. The GUI reads them. The agent has no
import, no dependency, no field, no flag that mentions the GUI.
Chapter 6's Observer seam is the entire interface.

### The observer, extended

Two new observation types, fired by the actor at tool boundaries:

```go
// New Observation types in internal/common/observer.go.
// Names must not collide with the existing mailbox ToolCompleted
// or event ToolCalled/ToolReturned.

type ToolDispatched struct {
    Agent  AgentID         `json:"agent,omitempty"`
    CallID string          `json:"call_id"`
    Name   string          `json:"name"`
    Input  json.RawMessage `json:"input"`
}

type ToolFinished struct {
    Agent   AgentID `json:"agent,omitempty"`
    CallID  string  `json:"call_id"`
    Result  string  `json:"result"`
    IsError bool    `json:"is_error"`
}

func (ToolDispatched) isObservation() {}
func (ToolFinished) isObservation()   {}
```

`ToolDispatched` fires before execution. `ToolFinished` fires after.
Together with the existing four observations, the Observer interface
now carries the complete lifecycle of a turn: state change, streaming
content, finalized parts, tool execution, and completion.

### The wire

A WebSocket handler implements Observer, serializes each observation
to JSON, and fans it out to every connected client:

```json
{"type":"part_delta","part_id":3,"kind":"text","chunk":"Hello"}
{"type":"state_changed","from":"idle","to":"thinking"}
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

No cursor. No sequence numbers. The event log has its own ordering
(the Seq from Chapter 6). On each live observation, the hub checks
for new event-log entries since the client's last-seen Seq and sends
them alongside the streaming content. The client never parses SSE
frames. Every message is a well-formatted JSON event with the
artifact id, type, and payload ready to render.

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
WebSocket messages, creates or updates Artifacts by part id, streams
chunks through the format-appropriate renderer, and replaces them
with finalized HTML when the part completes. Different visual
treatment comes from CSS classes, not separate component types.

### TTS

Two channels. **Auto-speak** fires during streaming: full text for
thinking and chat, an abbreviated summary for tool calls ("read file:
main.go, line 42"), silence for tool results. **On-demand** via a
speaker icon on every Artifact card, speaking the full accessible
text on click.

Chrome's `speechSynthesis` loses its voice on tab switch. Fix: a
zero-volume empty utterance on every `visibilitychange` event. Every
student building TTS in Chrome will hit this.

### Pause

```
paused = tts_speaking OR user_typing
```

Checked at every tool call boundary. The engine checks a shared
pause gate before starting each tool; the server sets and clears it
on receiving `pause`/`unpause` from any client. Rules:

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
| event.log | agent events (append-only) | ch6 |
| api.log | LLM wire JSON | ch6 |
| debug.log | diagnostics | ch6 |
| **gui.log** | **WebSocket JSON, both directions** | **ch8** |

### Yours

Markdown rendering library. ANSI-to-HTML approach. Visual styling.
TTS voice and speed. WebSocket library. Whether finalized Artifacts
replace or augment the streamed view. How the speaker icon looks.

### The exercise

`ch08/main.go`: one agent, one HTTP server on `--port` (default 8088),
serving static files from `ch08/web/gui/`. The binary accepts prompts
over WebSocket, streams observations to all connected clients,
supports reconnection, gates tool calls on pause, and logs to gui.log.

Configuration uses flags (e.g. `--port`, `--model`). No environment
variables.

```
make grade8
```

| check | points | what it tests |
|---|---|---|
| `websocket-streams` | 25 | subscribe, prompt via WS; receive part_delta, part_final, state_changed, turn_ended, tool_dispatched, tool_finished with correct fields |
| `event-replay` | 20 | disconnect after a turn completes; reconnect with a fresh subscribe; receive event-log history covering the completed turn; tool and text events present |
| `gui-log` | 15 | gui.log contains JSON lines with timestamps; both server-to-client and client-to-server messages present |
| `pause-holds-tools` | 20 | during a multi-tool turn, pause prevents the next tool from starting; unpause resumes; all tools eventually complete |
| `ch7-parity` | 20 | every Chapter 7 check still passes |
| **total** | **100** | |

---

## 8.1 The idea in plain words

A radio station transmits regardless of how many receivers are tuned
in. Add a radio and it hears from that moment forward. Unplug one and
the station does not pause. Play the recording from an earlier
timestamp and hear what you missed.

The agent is the station. Observations are the broadcast: streaming
chunks, finalized parts, state changes, tool starts, tool completions.
The GUI is a radio. The observation buffer is the recording.

**The agent does not know the GUI exists.** No WebSocket import, no
rendering code, no GUI flag in the configuration. The agent writes
observations to the Observer interface from Chapter 6. A Go WebSocket
handler implements that interface, serializes each observation to JSON,
and fans it out to every connected browser. Attach zero browsers and
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

## 8.2 Extending the broadcast

Chapter 6 built four observations: `PartDelta`, `PartFinal`,
`StateChanged`, `TurnEnded`. They carry the lifecycle of a response
but between responses the observer goes dark. The model requests a
tool call and the next thing any observer hears is the result, wrapped
in the following response.

`ToolDispatched` fires before execution: call id, tool name, raw
input. `ToolFinished` fires after: call id, result, error flag.
The observer now sees the complete turn lifecycle with no gaps.

The names avoid a silent collision. The event log already has
`ToolCalled` and `ToolReturned` for replay. The mailbox has
`ToolCompleted` as the internal signal from tool to actor. Three
systems, three names for what looks like the same thing: the event
is the record, the mailbox message is the signal, the observation
is the broadcast. Using the same name for two of them compiles fine
and produces a conversation about why something fired twice.

---

## 8.3 One tool at a time

The actor dispatches tools serially: start one, wait for completion,
start the next. This was true in Chapter 6 and the code has not
changed. What changes is that the property now matters.

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

## 8.4 The pause gate

Pausing is not a message. There is no `Pause` in the mailbox, no
event in the log, no observation. A paused agent is one whose actor
is blocked on a condition variable before starting the next tool.

`PauseGate` lives in `internal/common`. Three methods: `Pause`,
`Unpause`, `WaitIfPaused(ctx) bool`. Created in `cmd/main.go`,
passed to both the engine and the WebSocket hub at construction.
Neither imports the other. The star topology holds.

The return value is the design. `true` means someone called `Unpause`.
`false` means the context was cancelled: interrupt. A paused agent
that receives an interrupt wakes and stops immediately. Without
context awareness, pause is a trap: the human would have to remember
to unpause before killing a turn, under exactly the pressure where
remembering extra steps fails.

The gate bypasses the mailbox deliberately. A mailbox message waits
in FIFO order behind pending tool completions, hints, and interrupts.
Pause is about immediacy, and a shared variable checked at dispatch
time has no queue to wait in.

---

## 8.5 The receiver

The hub implements `Observer`. Its `Observe` method handles two
tiers: streaming content (deltas) accumulates in an in-flight map
keyed by part id; everything else fans out to connected clients as
a well-formatted JSON message. Then it returns. If a client's send
channel is full, the message is dropped for that client. The actor
never blocks.

Each client has a write goroutine draining a buffered channel.
`Observe` iterates the set and does a non-blocking send on each.
The hub is a broadcaster, and the design that makes
it safe is the one the opening metaphor promised: the station does
not wait for the radio.

Prompts and hints from the browser reach the agent through the same
`Ask` and `Hint` methods the terminal calls. The agent cannot
distinguish the two, and that is the proof the seam works.

---

## 8.6 The wire

No sequence numbers. No cursor. A `subscribe` message carries no
state at all. The server decides what history to send based on its
own event-log window (configurable: last hour or last 100 renderable
events, whichever is larger), then sends any in-flight streaming
content, then switches to live delivery.

A reconnecting client and a fresh client use the same path. The
server does not need to know whether this is a first connection or a
tenth. The event-log window is the same either way. This is simpler
than tracking per-connection offsets and produces a better result:
the client always gets a useful amount of context, never an empty
pane.

---

## 8.7 Two tiers of state

The hub holds two kinds of state, matching the two things a
connecting client needs.

**Tier 1: The event log.** Completed events from Chapter 6's
append-only log. Finalized parts, tool calls and returns, state
transitions, user messages. On subscribe, the hub reads the recent
window and converts each event to a well-formatted wire message the
client can render directly. These are durable: they survive
disconnection, restart, anything.

**Tier 2: In-flight partials.** Accumulated streaming content for
anything the agent is currently producing. A map from part id to
the concatenated chunks received so far. On subscribe, the hub sends
the current partial for each active part. On `PartFinal`, the
partial is cleared. These are ephemeral: they exist only while
content is streaming.

The event log is safe to read without locking. It is append-only:
once an element is written, it never changes. The hub snapshots the
current length during each `Observe` call (which runs on the engine
goroutine) and stores it under its own mutex. Subscribers read that
snapshot. A slow WebSocket send never blocks the engine.

For this chapter the event-log window is hardcoded. A configurable
window (time-based, count-based, or both) is future work, because it
is a policy decision the chapter should not make for the student.

---

## 8.8 The fourth log

```
2026-09-17T14:32:01.123Z > {"type":"subscribe"}
2026-09-17T14:32:01.456Z < {"type":"part_delta","part_id":3,"kind":"text","chunk":"Hello"}
```

`>` is client-to-server. `<` is server-to-client. The hub writes
gui.log directly, bypassing the Host interface. This is
transport-layer tracing, the same tool as `api.log` for a different
wire.

A log missing one direction lies about what happened. The lie
surfaces at the worst moment: the message that was not delivered,
invisible in the record because the record only shows one side.

---

## 8.9 Everything is an Artifact

The chapter is named for this.

An Artifact streams, then finalizes. Thinking is an Artifact. Chat
is an Artifact. A tool call's arguments, a tool's result, a user
message: all Artifacts. They differ in CSS class and renderer, not
in kind.

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

None of this is graded. The grader tests Go. But an exercise whose
output is invisible is an exercise nobody finishes.

---

## 8.10 The voice

TTS auto-speaks the stream. Thinking and chat in full. Tool
dispatches abbreviated: "read file: main.go, line 42." Tool results
silent. A speaker icon on every Artifact card speaks the accessible
text on click.

Chrome's `speechSynthesis` sleeps on tab switch and refuses to speak
when you return. Fix: a zero-volume empty utterance on every
`visibilitychange` event. Ugly, and nobody invents it independently.

`onerror` must mirror `onend`. Chrome fires `error` with reason
`interrupted` when one utterance cancels another, which is normal queue
behaviour. An error handler that does not advance the queue deadlocks
it: silence, permanent, no error message. Both callbacks mean "this
utterance is done, advance."

---

## 8.11 Pause in the browser

```
paused = tts_speaking OR user_typing
```

TTS queues an utterance: send `pause`. Queue empties: send `unpause`,
unless the input has text. User types: `pause`. User sends: stay
paused if TTS is speaking. ESC with empty input: cancel TTS,
re-evaluate.

Edge-triggered. Send `pause` on the transition, send `unpause` on
the reverse. The WebSocket carries transitions; the server holds state.
A client sending `pause` per queued utterance floods the wire with
messages that carry no information.

Streaming into a running tool continues. The gate is between tools,
not inside them.

---

## Taking it for a spin

```
$ go run ./agent/cmd --port 8088
```

Open `http://localhost:8088`. Dark page, input field, "idle." Type a
prompt. Reasoning appears dim, the reply streams, TTS reads it. Ask
for a tool and a card appears with the arguments, the result fills
in, the reply continues.

Open a second tab. Both show the same stream. Close one, reopen:
it subscribes, receives the event-log window, catches up in a burst. Type a hint in the terminal; both tabs see
it. Start typing in the browser: "paused," no new tool starts.
Interrupt from the terminal: the agent stops even while paused.

Terminal and browser, two views of one agent, neither aware of the
other.

---

## What Chapter 9 does with this

One scroll pane, one input field, one dark page. Chapter 9 adds the
workspace: a second `ArtifactScroll` for tool calls and results, an
agent tree, drag bars, settings, and a sidebar that grows as chapters
add features. The Artifact is the reusable unit. The layout is what
remains.
