# Ch8 Coder Brief: Everything Is an Artifact

## Context

Chapter 7 built streaming. The terminal renders reasoning, chat, and
tool arguments as they arrive, in three colours with no layout.
Chapter 8 builds the GUI: a WebSocket transport, a browser component,
TTS, and a pause mechanism. The agent's existing source changes only
in two places: adding two new Observation types, and checking a pause
flag between tool calls.

This brief specifies both Go (graded) and JS/HTML/CSS (not graded but
required for the exercise to be usable). The grader tests the
WebSocket protocol from a Go WebSocket client — it never opens a
browser.

## Architecture (non-negotiable)

1. **The agent does not know the GUI exists.** The GUI is an Observer.
   No existing agent package may import the WebSocket handler package.
   Verify with `go list -deps ./internal/llm/... ./internal/tools/...
   ./internal/jobs/...` — no `ws` package.

2. **Observer MUST NOT BLOCK.** The hub's Observe method buffers and
   returns. A slow WebSocket client cannot park the actor.

3. **Multiple clients, fan-out.** Any number of browsers connect to
   the same agent. Each sees the same observation stream.

4. **Reconnection from the observation buffer.** The hub keeps an
   append-only list of serialized messages, each with a seq. On
   subscribe(cursor), replay messages with seq > cursor. For this
   exercise (single short session), unbounded is fine.

5. **gui.log is transport-layer logging.** It does not go through the
   Host interface. The hub writes it directly.

6. **No React, no build system, no npm.** Vanilla JS. External
   libraries (markdown renderer, ANSI converter) via CDN `<script>`
   tags.

## What to build

### 1. Observer extension (internal/common/observer.go)

Two new Observation types. Names must not collide with:
- `ToolCompleted` (mailbox Inbound type)
- `ToolCalled` / `ToolReturned` (Event types)

Suggested names (author can change these):

```go
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

### 2. Fire observations in the actor (internal/llm/actor.go)

In `dispatchTool`: fire `ToolDispatched` before starting the tool.
In `waitForTools` (or wherever `ToolCompleted` mailbox messages are
processed): fire `ToolFinished` after each tool completes.

The actor already calls `a.notify(obs)` for other observations. Same
pattern.

### 3. Pause gate (internal/llm/actor.go)

Add a pausable interface or a direct mechanism:

```go
// Pauser is checked by the actor between tool calls.
type Pauser interface {
    Paused() bool
    WaitUnpaused(ctx context.Context) // blocks until unpaused or ctx done
}
```

Between tool calls in the main loop (before `dispatchTool`), if
`paused`, the actor calls `WaitUnpaused`. Interrupt must still work
while paused — if the actor is interrupted while waiting for unpause,
it should stop.

The hub implements Pauser. It sets/clears a flag on receiving
`pause`/`unpause` WebSocket messages from any client. The WaitUnpaused
method blocks on a condition variable or channel.

The Pauser is passed to the Actor at construction or set via a method
after creation.

### 4. WebSocket hub (new package: internal/ws/)

**handler.go** — the core:

```go
type Hub struct {
    mu       sync.Mutex
    seq      uint64
    messages []Message       // append-only buffer
    clients  map[*Client]bool
    paused   bool
    pauseCh  chan struct{}    // signaled on unpause
    guiLog   *GuiLogger
}

// Observe implements common.Observer. Must not block.
func (h *Hub) Observe(obs common.Observation) { ... }

// ServeWS upgrades an HTTP request to WebSocket.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) { ... }

// Paused / WaitUnpaused implement the Pauser interface.
func (h *Hub) Paused() bool { ... }
func (h *Hub) WaitUnpaused(ctx context.Context) { ... }
```

Each Client has a write goroutine and a buffered send channel.
Messages dropped if the channel is full (slow client does not block
the hub).

**Message format (on the wire):**

```go
type Message struct {
    Type string          `json:"type"`
    Seq  uint64          `json:"seq,omitempty"`
    // Remaining fields vary by type. Use either:
    // (a) json.RawMessage Data field, or
    // (b) embed observation structs and rely on JSON tags
}
```

### 5. WebSocket protocol

**Server → Client** (every message has `type` and `seq`):

| type | additional fields | source |
|---|---|---|
| `part_delta` | part_id, kind, chunk, agent | PartDelta observation |
| `part_final` | part_id, seq (the Part's seq), part | PartFinal observation |
| `state_changed` | from, to, agent | StateChanged observation |
| `turn_ended` | text, error, agent | TurnEnded observation |
| `tool_dispatched` | call_id, name, input, agent | ToolDispatched observation |
| `tool_finished` | call_id, result, is_error, agent | ToolFinished observation |

Note: the `seq` in the WebSocket envelope is the WebSocket message
sequence number (monotonically increasing per hub). The `seq` in
`part_final` is the Part's sequence number from the seam. Different
sequences, different purposes.

**Client → Server** (no seq):

| type | fields | action |
|---|---|---|
| `subscribe` | cursor (uint64, default 0) | replay from seq > cursor, then live |
| `prompt` | text (string) | call Agent.Ask or Actor.Ask |
| `hint` | text (string) | call Agent.Hint or Actor.Hint |
| `interrupt` | (none) | call Agent.Interrupt or Actor.Interrupt |
| `pause` | (none) | set paused flag |
| `unpause` | (none) | clear paused flag, wake WaitUnpaused |

### 6. gui.log (internal/ws/gui_log.go)

One log line per WebSocket message, both directions:

```
2026-09-17T14:32:01.123Z > {"type":"subscribe","cursor":0}
2026-09-17T14:32:01.456Z < {"type":"part_delta","seq":1,"part_id":3,"kind":"text","chunk":"Hello"}
```

`>` = client-to-server, `<` = server-to-client. ISO 8601 timestamps.
Written by the hub on every message send/receive.

The gui.log path is passed at hub construction, alongside the other
log paths (api.log, debug.log, event.log). Open the file in the hub
constructor.

### 7. HTTP server (agent/main.go)

Add an HTTP server to main.go:

- `GET /` and static files → serve from `static/` directory
- `GET /ws` → upgrade to WebSocket (hub.ServeWS)
- Port from `CH08_PORT` env var, default `8088`

The HTTP server runs alongside the existing CLI stdin loop. Both work
simultaneously: the user can type prompts in the terminal OR in the
browser. The hub and CLI both call Agent.Ask/Hint/Interrupt.

### 8. Client-side: ArtifactScroll (static/artifact-scroll.js)

The reusable scroll view component. Not graded, but the exercise
needs a working browser interface.

```javascript
class ArtifactScroll {
    constructor(container, options = {}) { ... }

    // Called for every parsed WebSocket message
    handleMessage(msg) { ... }
}
```

On `part_delta`: find or create an Artifact element by part_id.
Stream the chunk through the appropriate renderer based on `kind`:
- `"thinking"` → markdown renderer (dim CSS class)
- `"text"` → markdown renderer
- `"tool_call"` → JSON renderer

On `part_final`: replace streaming content with finalized HTML.

On `tool_dispatched`: create a tool execution card showing name +
abbreviated input.

On `tool_finished`: update the tool card with result (or error).

On `state_changed`: update a status indicator.

On `turn_ended`: scroll to bottom, update status to idle.

### 9. Built-in renderers (static/renderers.js)

- **Markdown**: Use marked.js via CDN. On each chunk, append to
  accumulated text and re-render the whole block (simple, correct).
- **ANSI**: Use ansi_up.js via CDN. Convert escape codes to styled
  HTML spans.
- **JSON**: Pretty-print with indentation. Collapsible sections for
  large objects.
- **Diff** (finalized only): For edit_file results, show old/new text
  with red/green line styling.

### 10. TTS (static/tts.js)

- Auto-speak thinking and chat text as chunks arrive (queue sentences).
- Auto-speak tool dispatches as abbreviated summaries:
  `read_file("main.go", start_line=42)` → "read file: main.go,
  line 42"
- Silent for tool results (too verbose).
- Speaker icon (🔊) on every Artifact card. Clicking speaks the full
  accessible text.
- Chrome wake-up: zero-volume utterance on `visibilitychange`.
- Pause integration: when TTS starts speaking or user starts typing,
  send `{"type":"pause"}`. When both clear, send `{"type":"unpause"}`.
- `onerror` must mirror `onend` and continue the queue.

### 11. Page (static/index.html)

Single-pane layout:
- ArtifactScroll container (full width, scrolling)
- Input field at bottom (prompt when idle, hint when working)
- Status bar (agent state, model name, token count)
- CDN imports: marked.js, ansi_up.js
- Dark theme: background #0d0d0d, surface #1a1a1a, text #e0e0e0

### 12. Exercise binary (ch08/main.go)

Imports the agent library. Starts one agent with standard tools.
Starts HTTP server serving `ch08/static/` and handling `/ws`.
Same behaviour as the reference `agent/main.go`. Students who build
the full agent can also build a standalone exercise.

Copy `agent/static/` to `ch08/static/` (or serve from agent/static
with a path override).

## Grading

### Checks

| check | points | what it tests |
|---|---|---|
| `websocket-streams` | 25 | subscribe, prompt via WS; receive part_delta, part_final, state_changed, turn_ended, tool_dispatched, tool_finished; seq is monotonically increasing |
| `event-replay` | 20 | disconnect after events; reconnect with cursor = last-seen seq; receive exactly the missed messages; replayed + continued matches full replay |
| `gui-log` | 15 | gui.log exists; contains JSON lines; timestamps present; both `>` and `<` directions present |
| `pause-holds-tools` | 20 | during a multi-tool turn, pause prevents the next tool_dispatched; unpause resumes; all tools eventually complete |
| `ch7-parity` | 20 | every Chapter 7 check still passes |
| **total** | **100** | |

### Harness (internal/grade/ch08_harness.go)

The ch8 harness is new territory: it tests WebSocket, not CLI.

1. Build the submission directory (`go build -o agent .`)
2. Start the binary as a subprocess (with `CH08_PORT` set to a
   random available port)
3. Wait for the HTTP server to be ready (poll `GET /` until 200)
4. Connect via WebSocket to `ws://localhost:<port>/ws`
5. Send `subscribe{cursor:0}`, then `prompt{text:"..."}` using the
   fake vendor
6. Read messages, validate structure and seq ordering
7. For replay: disconnect, reconnect with cursor, compare
8. For pause: send pause between tools, verify no tool_dispatched
   within a timeout, send unpause, verify completion
9. Check gui.log file

The harness needs a WebSocket client library. Use the same one the
student solution uses (gorilla/websocket is the safe choice — it's
already in the ecosystem).

### Testing the pause check

The pause check is timing-sensitive. To make it robust:

1. Configure the fake vendor to return a response with 3 tool calls
   (e.g., think(2s), think(2s), think(2s) — using ch6's think tool)
2. Wait for the first `tool_dispatched`
3. Send `pause`
4. Wait for the first `tool_finished` (the in-flight tool completes)
5. Wait 2 seconds — verify NO second `tool_dispatched` arrives
6. Send `unpause`
7. Wait for remaining `tool_dispatched` and `tool_finished` messages
8. Verify all 3 tools completed

The 2-second tool duration gives ample time for the pause message to
be processed before the first tool finishes.

### Mutation tests (P9 audit)

At least 5 mutations, each deleting exactly one behaviour:

| mutant | what it deletes | expected failing checks |
|---|---|---|
| `no-ws-seq` | remove seq from server messages | {event-replay} |
| `no-tool-obs` | remove ToolDispatched/ToolFinished observations | {websocket-streams} |
| `no-replay` | subscribe always replays from seq 0 | {event-replay} |
| `no-gui-log` | disable gui.log writing | {gui-log} |
| `no-pause-gate` | ignore pause, always execute tools | {pause-holds-tools} |

Each must fail its targeted check(s) and pass the rest.

## Files to create

```
agent/internal/ws/handler.go          — Hub, Client, WebSocket handler
agent/internal/ws/gui_log.go          — gui.log writer
agent/static/index.html               — single-page GUI
agent/static/artifact-scroll.js       — reusable scroll component
agent/static/renderers.js             — markdown, ANSI, JSON, diff
agent/static/tts.js                   — TTS + Chrome fix + pause
agent/static/style.css                — dark theme
ch08/main.go                          — exercise binary
ch08/static/                          — exercise static files
internal/grade/ch08_checks.go         — grader checks
internal/grade/ch08_harness.go        — WebSocket test harness
internal/grade/ch08_grader_test.go     — mutation tests
```

## Files to modify

```
agent/internal/common/observer.go     — add ToolDispatched, ToolFinished
agent/internal/llm/actor.go           — fire tool observations; check pause
agent/main.go                         — add HTTP server
cmd/grade/main.go                     — add case 8
Makefile                              — add grade8, grade-dir for ch8
```

## Constraints

- All ch1-ch7 graders must pass 100/100 after changes.
- `go vet ./...` clean.
- No existing agent package imports `internal/ws/`.
- Observer.Observe must not block.
- The pause check must not deadlock when combined with interrupt.
- gui.log uses `>` for client-to-server, `<` for server-to-client.
- WebSocket library: gorilla/websocket (recommended) or equivalent.
  Add to go.mod.

## Open questions for Bill

1. **Observation type names**: `ToolDispatched`/`ToolFinished` avoid
   collision with mailbox `ToolCompleted` and event `ToolCalled`/
   `ToolReturned`. Better names welcome.

2. **Pause mechanism**: Should the actor check a `Pauser` interface
   (clean, testable) or should `Pause`/`Unpause` be mailbox Inbound
   types processed by the drain loop (consistent with existing hint/
   interrupt pattern)?

3. **WebSocket seq vs Part seq**: Two different sequences share the
   name `seq`. The WebSocket envelope seq is the message counter. The
   PartFinal seq is the part counter from the seam. Rename one?

4. **Static file location**: `agent/static/` or `agent/web/` or
   `static/` at repo root?

5. **HTTP port**: env var only (`CH08_PORT`) or also a CLI flag?
