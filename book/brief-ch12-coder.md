# Ch12 Coder Brief — MCP

## Chapter thesis

MCP is the tool seam. External tools join the agent's registry at
runtime via a standard protocol. Same internal tool interface,
different origin. And some tools are *ephemeral* — auto-called by
the engine each round, their results replacing the previous
snapshot in a volatile context slot.

The chapter builds MCP client infrastructure (JSON-RPC 2.0,
transport abstraction, tool discovery, bidirectional calls) and
a browser-hosted MCP server that gives the LLM eyes and hands
on the GUI: a markdown snapshot of the DOM, click/input actions,
and a TTS queue so the agent can "hear" what the human hears.

## Starting point

`solutions/ch11` (or the current `agent/` HEAD — they should
be identical).

## What to build

### 1. JSON-RPC 2.0 core (`internal/mcp/jsonrpc.go`, ~100 lines)

Types for JSON-RPC 2.0 messages:

```go
type Request struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      any             `json:"id,omitempty"`      // int or string; nil = notification
    Method  string          `json:"method"`
    Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      any             `json:"id"`
    Result  json.RawMessage `json:"result,omitempty"`
    Error   *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
    Code    int             `json:"code"`
    Message string          `json:"message"`
    Data    json.RawMessage `json:"data,omitempty"`
}
```

Correlation: outgoing requests get monotonic int IDs. A background
goroutine reads responses and routes them to waiting callers via
a `map[int]chan Response` protected by a mutex.

### 2. Transport interface (`internal/mcp/transport.go`, ~30 lines)

```go
type Transport interface {
    Send(msg json.RawMessage) error
    Recv() (json.RawMessage, error)
    Close() error
}
```

Three implementations cover every MCP server scenario:

| Transport | Use case | SKILL.md `transport:` |
|-----------|----------|----------------------|
| StdioTransport | Spawn a local subprocess (Python/Go MCP server) | `stdio` |
| WSTransport | Tunnel over the GUI WebSocket | `websocket` |
| PipeTransport | In-process testing (grader, unit tests) | — (test only) |

A future fourth (URL/port-based) can be added later — the
interface is the seam.

### 3. StdioTransport (`internal/mcp/stdio.go`, ~80 lines)

Spawns a subprocess and talks newline-delimited JSON-RPC over
stdin/stdout. This is the standard MCP transport. The subprocess
is a managed process — same lifecycle discipline as background
jobs from `run_command`.

```go
func NewStdioTransport(command string, args []string, env []string) (*StdioTransport, error)
```

- Starts the process immediately
- `Send` writes JSON + newline to stdin
- `Recv` reads one JSON line from stdout
- `Close` sends EOF on stdin, waits for exit with bounded timeout
  (same as Job.awaitReaped — don't hang on a stuck server)

### 4. PipeTransport (`internal/mcp/pipe.go`, ~60 lines)

In-process transport using `io.Pipe`. Both ends get a Transport.
Used for testing — the grader runs a fake MCP server on one end.

```go
func NewPipeTransport() (client Transport, server Transport)
```

Messages are newline-delimited JSON (one JSON object per line).

### 5. WSTransport (`internal/mcp/ws_transport.go`, ~50 lines)

Tunnels JSON-RPC over the existing WebSocket connection. The Hub
routes messages with `"type":"jsonrpc"` to/from this transport.

Hub changes: `handleClientMessage` gains a `"jsonrpc"` case that
forwards `msg.Data` to the WSTransport's receive channel. Outgoing
messages from WSTransport are sent as `{"type":"jsonrpc","data":{...}}`.

### 5. MCP Client (`internal/mcp/client.go`, ~200 lines)

```go
type Client struct {
    transport Transport
    // correlation map, next ID, etc.
}

func NewClient(t Transport) *Client
func (c *Client) Initialize(ctx context.Context) error
func (c *Client) ListTools(ctx context.Context) ([]ToolInfo, error)
func (c *Client) CallTool(ctx context.Context, name string, args json.RawMessage) (*ToolResult, error)
func (c *Client) Close() error
```

**Initialize** sends `initialize` request, waits for response,
sends `notifications/initialized`.

**ListTools** sends `tools/list`, returns tool schemas.

**CallTool** sends `tools/call`, waits for result.

**ToolInfo** carries enough to build a `common.ToolDecl`:

```go
type ToolInfo struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    InputSchema json.RawMessage `json:"inputSchema"`
    Ephemeral   string          `json:"ephemeral,omitempty"` // "round", "turn", or ""
}
```

### 6. Reverse tool handler (`internal/mcp/reverse.go`, ~80 lines)

The client also listens for incoming `tools/call` requests FROM
the MCP server (bidirectional). When the server calls a tool:

1. Look up the tool in the agent's registry
2. If found and the MCP connection is trusted: execute it
3. Send the result back as a JSON-RPC response

```go
func (c *Client) SetReverseHandler(handler func(name string, args json.RawMessage) (string, error))
```

The handler is wired by the engine to route through the normal
tool execution path.

### 7. Ephemeral tool mechanism (`internal/common/context.go` + engine changes, ~60 lines)

**ToolDecl gains a field:**

```go
type ToolDecl struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Schema      json.RawMessage `json:"input_schema"`
    Ephemeral   string          `json:"ephemeral,omitempty"` // "round" or "turn"
}
```

**Engine changes (in the round loop, before building the request):**

1. Collect all tools where `Ephemeral == "round"`
2. Call each one through the normal job infrastructure
   (context-cancellable — if the turn is interrupted,
   pending ephemeral calls are cancelled too)
3. Build a combined markdown string from their results
4. Set `context.Ephemera` to that string

For `"turn"` ephemeral tools, call them once when the turn starts
(after the prompt arrives, before the first round).

Ephemeral tools are NOT included in the tool declarations sent to
the LLM. The LLM never sees them as callable tools — it sees their
output as context data. (This prevents the LLM from calling
gui_snapshot manually, which would produce duplicate snapshots.)

**Ephemera format:**

```markdown
## GUI State (auto-updated)

[markdown from gui_snapshot]

## TTS Queue (auto-updated)

[json from tts_queue]
```

### 8. MCP tool bridge (`internal/mcp/bridge.go`, ~80 lines)

Converts MCP-discovered tools into the agent's tool registry:

```go
func Bridge(client *Client, tools []ToolInfo) []common.Tool
```

**MCP tools are processes, not functions.** Each bridged tool's
handler is a normal tool handler — takes `context.Context`,
returns result. The framework's existing job infrastructure wraps
every call, giving it a handle, cancellation, and wait/kill.

The handler calls `client.CallTool(ctx, name, args)`. If the
job is killed (context cancelled), the client stops waiting for
the JSON-RPC response and sends a `$/cancelRequest` notification
to the MCP server.

This is critical: MCP servers are external processes. Their tool
calls can hang, time out, or fail. The same job control that
manages `run_command` manages MCP tool calls. The framework sees
no difference — a tool is a tool.

Non-ephemeral MCP tools are registered as normal tools. Ephemeral
MCP tools are registered with the `Ephemeral` field set.

### 9. Browser MCP server (JavaScript, `web/gui/mcp.js`, ~200 lines)

Runs inside the browser. Speaks JSON-RPC 2.0 over the WebSocket.

**Implements these MCP methods:**

- `initialize` → returns capabilities (tools, reverse_tools)
- `tools/list` → returns tool schemas
- `tools/call` → dispatches to tool handlers

**Tools provided:**

| Tool | Ephemeral | Description |
|------|-----------|-------------|
| `gui_snapshot` | round | Returns markdown of the current DOM |
| `gui_click` | — | Clicks an element by CSS selector |
| `gui_input` | — | Sets text of an input by CSS selector |
| `tts_queue` | round | Returns pending TTS utterances as JSON |

**gui_snapshot** walks the visible DOM and produces a markdown
tree of elements with:
- Tag, text content (truncated), visible state
- CSS selectors for interactive elements (buttons, inputs, tabs)
- Pane layout (which panes visible, approximate sizes)
- Artifact summaries (type, first few lines)

Keep it under ~4KB to avoid flooding context.

**tts_queue** returns an array of objects:
```json
[
  {"text": "File saved successfully", "state": "playing", "startedAt": 1695...},
  {"text": "Running tests...", "state": "pending", "startedAt": null}
]
```

### 10. Integration wiring (`cmd/main.go` + `agent.go`, ~50 lines)

- `agent.go` gains `ConnectMCP(transport Transport) error` which:
  1. Creates an MCP Client
  2. Calls Initialize
  3. Calls ListTools
  4. Bridges discovered tools into the registry
  5. Sets up reverse tool handler
  6. Stores the client for cleanup

- `cmd/main.go` passes the WSTransport from Hub to ConnectMCP

- Hub gains `MCPTransport() Transport` method that returns a
  WSTransport wired to the client's WebSocket

## Grader checks (7, 100 points)

| Check | Points | Tests |
|-------|--------|-------|
| mcp-handshake | 15 | Fake MCP server via PipeTransport; verify initialize/initialized exchange |
| tool-discovery | 15 | tools/list populates agent registry; declarations include MCP tools |
| mcp-tool-call | 15 | Call a discovered MCP tool; result appears in dialogue as ToolReturned event |
| ephemeral-round | 20 | Ephemeral tool auto-called each round; result in context.Ephemera, NOT in dialogue |
| reverse-call | 15 | MCP server calls an agent tool; result flows back as JSON-RPC response |
| ws-tunnel | 10 | JSON-RPC messages routed through WebSocket hub (WSTransport) |
| ch11-parity | 10 | All ch11 behavior preserved |

### Fake MCP server (for grader)

The grader includes a `fakemcp` package that implements a minimal
MCP server over PipeTransport:

```go
type FakeServer struct {
    Tools    []mcp.ToolInfo
    OnCall   func(name string, args json.RawMessage) (string, error)
}

func (s *FakeServer) Serve(transport mcp.Transport)
```

It handles:
- `initialize` → responds with capabilities
- `tools/list` → returns s.Tools
- `tools/call` → calls s.OnCall, returns result

For `reverse-call` check, the fake server sends a `tools/call`
request TO the client after initialization.

## SKILL.md `mcp_servers` field

Skills can declare MCP servers. Loading the skill starts the
server; unloading stops it. Three transport types:

```yaml
---
name: browser-debug
description: GUI inspection and control
mcp_servers:
  - name: gui
    transport: websocket          # tunneled over GUI WebSocket
---
```

```yaml
---
name: code-search
description: Semantic code search via MCP
mcp_servers:
  - name: search-server
    transport: stdio              # spawn subprocess
    command: python3
    args: ["mcp_server.py"]
---
```

```yaml
---
name: remote-tools
description: Tools from a remote MCP server
mcp_servers:
  - name: remote
    transport: url                # connect to running server (future)
    url: "http://localhost:8080"
---
```

**Parsing**: `SkillProperties` gains `MCPServers []MCPServerConfig`.
`MCPServerConfig` has `Name`, `Transport` (enum: stdio/websocket/url),
`Command`, `Args`, `Env`, `URL`. The skill registry calls
`ConnectMCP` on load and `DisconnectMCP` on unload.

## Exercise contract

```
./ch12 prompt "hello"         # normal agent behavior (ch11 parity)
./ch12 --mcp-pipe             # connect MCP client to stdin/stdout pipe
                              # (grader uses this to inject fake MCP server)
```

The `--mcp-pipe` flag creates a StdioTransport reading/writing
the process's own stdin/stdout so the grader can drive the MCP
protocol externally. In normal operation (no flag), the agent
connects to MCP servers declared in loaded skills.

## Files to create/modify

### New files
- `internal/mcp/jsonrpc.go` — JSON-RPC 2.0 types + correlation
- `internal/mcp/transport.go` — Transport interface
- `internal/mcp/pipe.go` — PipeTransport
- `internal/mcp/ws_transport.go` — WSTransport
- `internal/mcp/client.go` — MCP Client
- `internal/mcp/reverse.go` — Reverse tool handler
- `internal/mcp/bridge.go` — MCP → agent tool bridge
- `web/gui/mcp.js` — Browser MCP server

### Modified files
- `internal/common/tool.go` — add Ephemeral field to ToolDecl
- `internal/common/context.go` — ephemeral tool calling in engine
- `internal/ws/handler.go` — jsonrpc message routing
- `agent.go` — ConnectMCP method
- `cmd/main.go` — wiring

### Grader files
- `internal/grade/ch12_checks.go`
- `internal/grade/ch12_grader_test.go`
- `internal/grade/ch12_harness.go` (if needed)

## Non-goals (deferred to later chapters)

- URL-based transport (connect to running server) — future
- GUI bug detection/fixing — ch13 (self-wielding)
- Audio/TTS capture for the LLM — aspirational
- MCP resource subscriptions, sampling, logging — future
