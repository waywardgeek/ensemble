# Chapter 12: MCP -- The Extension Protocol

Every tool the agent has used so far was compiled into the binary. Adding a new tool means writing Go, rebuilding, and restarting. The Model Context Protocol changes that. MCP is a JSON-RPC 2.0 wire protocol for connecting an agent to external tool servers: processes, browsers, remote services. This chapter builds the client, the transport layer, and a mechanism the MCP spec does not have -- ephemeral tools that the engine calls automatically, injecting their output into the context window without the LLM ever knowing they exist.

## TL;DR

MCP is JSON-RPC 2.0. The client sends `initialize`, receives capabilities, sends `notifications/initialized`, then calls `tools/list` to discover what the server offers. Each discovered tool becomes a normal tool in the agent's registry, callable by the LLM. The bidirectional channel means the server can also call agent tools back.

### Transport

One interface, multiple implementations:

```go
type Transport interface {
    Send(msg json.RawMessage) error
    Recv() (json.RawMessage, error)
    Close() error
}
```

**PipeTransport**: in-process connected pair, for testing.
**StdioTransport**: spawns a subprocess, JSON-RPC over its stdin/stdout.
**RawTransport**: wraps an `io.Reader` and `io.Writer` -- the `--mcp-pipe` flag uses this.
**WSTransport**: tunnels JSON-RPC through the WebSocket hub to the browser.

### Codec

The `Codec` handles JSON-RPC correlation: outgoing requests get incrementing integer IDs, responses are matched by ID and delivered to the blocked caller. A background goroutine reads the transport and dispatches: responses go to pending callers, incoming requests go to an `onRequest` callback for reverse tool handling.

```go
type Codec struct {
    transport Transport
    nextID    int
    pending   map[int]chan json.RawMessage
    onRequest func(Request)
    done      chan struct{}
    mu        sync.Mutex
}
```

### MCP Client

```go
func NewClient(t Transport) *Client
func (c *Client) Initialize(ctx context.Context) error
func (c *Client) ListTools(ctx context.Context) ([]ToolInfo, error)
func (c *Client) CallTool(ctx context.Context, name string, args json.RawMessage) (*ToolResult, error)
func (c *Client) SetReverseHandler(handler func(name string, args json.RawMessage) (string, error))
func (c *Client) Done() <-chan struct{}
func (c *Client) Close() error
```

`CallTool` runs the RPC in a goroutine so context cancellation works -- if the job is killed, the context is cancelled, and the client sends `$/cancelRequest` to the server. MCP tool calls are processes, not functions. They go through the same job infrastructure as `run_command`.

### Ephemeral tools

A discovered tool may carry an `ephemeral` field: `"round"` or `"turn"`. Ephemeral tools are NOT included in the tool declarations sent to the LLM. The model never sees them as callable. Instead, the engine calls them automatically:

- **round**: called before every `RequestSent`. The GUI snapshot arrives fresh each round.
- **turn**: called once when a new turn starts. Configuration data that does not change mid-conversation.

Results are combined and injected via `Attach`, which puts them into `Context.Ephemera` -- the field the context already replaces on each round.

```go
func (e *Engine) callEphemeral(mode string) error {
    tools := e.Tools.EphemeralTools(mode)
    if len(tools) == 0 {
        return nil
    }
    var parts []string
    for _, t := range tools {
        c := &common.Call{Host: e.Host, Jobs: e.Jobs}
        out, err := t.Run(c, nil)
        if err != nil {
            e.Host.Logf("ephemeral tool %s error: %v", t.Name, err)
            continue
        }
        if out != "" {
            parts = append(parts, fmt.Sprintf("## %s (auto-updated)\n\n%s", t.Name, out))
        }
    }
    if len(parts) == 0 {
        return nil
    }
    return e.Attach(strings.Join(parts, "\n\n"))
}
```

Ephemeral errors are logged but not fatal. A GUI snapshot failure should not abort a turn.

### Bridge

The bridge converts `ToolInfo` from MCP discovery into `common.Tool` entries. The handler closure captures the MCP client and routes calls through `CallTool`:

```go
func Bridge(client *Client, tools []ToolInfo) []common.Tool {
    var result []common.Tool
    for _, t := range tools {
        info := t
        tool := common.Tool{
            Name:        info.Name,
            Description: info.Description,
            Schema:      info.InputSchema,
            Ephemeral:   info.Ephemeral,
            Run: func(c *common.Call, args json.RawMessage) (string, error) {
                r, err := client.CallTool(context.Background(), info.Name, args)
                // ...
            },
        }
        result = append(result, tool)
    }
    return result
}
```

### Reverse calls

The MCP channel is bidirectional. When the server sends a `tools/call` request, the client's reverse handler looks up the tool in the agent's registry and executes it. The result goes back as a JSON-RPC response. This means an MCP server -- a browser, a Python script, a remote service -- can call `read_file`, `run_command`, or any tool the agent has loaded, subject to trust and skill boundaries.

### WebSocket tunneling

The browser cannot open a port or spawn a subprocess. The WebSocket hub already carries `prompt`, `hint`, `interrupt`, and `settings` messages. MCP adds one more type: `jsonrpc`. The hub routes `{"type": "jsonrpc", "payload": {...}}` messages between the Go MCP client and the browser's MCP server.

On the browser side, `mcp.js` intercepts these frames and speaks the full MCP protocol: `initialize`, `tools/list`, `tools/call`. It registers four tools:

- **gui_snapshot** (ephemeral/round): walks the visible DOM and returns a markdown summary -- pane layout, interactive elements with CSS selectors, artifact previews. Capped at 4KB. The cap is reported rather than silent: the summary states how many characters it elided and how many artifacts it omitted, and it reports the state attributes of every control it lists. An observer that truncates in silence will tell you a screen looks fine when it never saw it, and a control whose state it cannot read is a control it will guess about.
- **gui_click(selector)**: dispatches a click event on the matched element.
- **gui_input(selector, text)**: sets the value and dispatches input/change events.
- **tts_queue** (ephemeral/round): returns pending TTS utterances as JSON -- text, state, timing.

The agent sees the GUI the way the user does: a snapshot of what is visible, updated every round. It can click buttons and fill text fields. And it hears what the TTS is saying, so it can catch bugs where the speech does not match the display.

### SKILL.md integration

Skills declare MCP servers in their frontmatter:

```yaml
mcp_servers:
  - name: browser-debug
    transport: websocket

  - name: code-search
    transport: stdio
    command: python3
    args: ["scripts/search_server.py"]
```

Loading a skill starts its MCP servers and discovers their tools. Unloading stops them. The `transport` field determines the wire: `stdio` spawns a subprocess, `websocket` tunnels through the hub.

### Exercise

The exercise contract:

```
./ensemble --mcp-pipe
```

Connects the MCP client to stdin/stdout. The grader acts as the MCP server on the other end: sends `initialize` response, `tools/list` response, and a reverse `tools/call` request. Seven checks:

| Check | Points |
|-------|--------|
| mcp-handshake | 15 |
| tool-discovery | 15 |
| mcp-tool-call | 15 |
| ephemeral-round | 20 |
| reverse-call | 15 |
| ws-tunnel | 10 |
| ch11-parity | 10 |

```
make grade12
```
