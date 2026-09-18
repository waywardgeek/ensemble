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
to JSON, and fans it out to every connected client. Every server
message carries a monotonic `seq`:

```json
{"type":"part_delta","seq":1,"part_id":3,"kind":"text","chunk":"Hello"}
{"type":"state_changed","seq":2,"from":"idle","to":"thinking"}
{"type":"tool_dispatched","seq":3,"call_id":"tc_1","name":"read_file","input":{"path":"main.go"}}
{"type":"tool_finished","seq":4,"call_id":"tc_1","result":"package main...","is_error":false}
{"type":"part_final","seq":5,"part_id":3,"part":{"type":"text","text":"The file contains..."}}
{"type":"turn_ended","seq":6,"text":"The file contains..."}
```

Client messages carry no seq:

```json
{"type":"subscribe","cursor":0}
{"type":"prompt","text":"Run the tests"}
{"type":"hint","text":"Skip the slow ones"}
{"type":"interrupt"}
{"type":"pause"}
{"type":"unpause"}
```

### Reconnection

A client subscribes with a cursor (last-seen `seq`). The server
replays every message since that seq, then switches to live delivery.
A fresh connection uses cursor 0. A client disconnected for an hour
catches up in one burst. The agent does not pause, restart, or notice.

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

Checked at every tool call boundary. The server holds the next tool
call until the client sends `unpause`. Rules:

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

`ch08/main.go`: one agent, one HTTP server on a port from `CH08_PORT`
(default 8088), serving static files from `ch08/static/`. The binary
accepts prompts over WebSocket, streams observations to all connected
clients, supports reconnection, gates tool calls on pause, and logs
to gui.log.

```
make grade8
```

| check | points | what it tests |
|---|---|---|
| `websocket-streams` | 25 | subscribe, prompt via WS; receive part_delta, part_final, state_changed, turn_ended, tool_dispatched, tool_finished; seq is monotonically increasing |
| `event-replay` | 20 | disconnect after events arrive; reconnect with cursor = last-seen seq; receive exactly the missed messages; replayed + continued = complete |
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
stream. Close a tab, reopen it an hour later, subscribe with the
last-seen sequence number, and catch up in one burst.

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
