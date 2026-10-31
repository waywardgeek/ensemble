# Chapter 11: MCP: The Extension Protocol

Every tool the agent has used so far was compiled into the binary. Adding a new one means writing Go, rebuilding, and restarting. MCP, the Model Context Protocol, is a JSON-RPC 2.0 wire protocol for connecting an agent to external tool servers: processes, browsers, remote services. This chapter builds the client, the transport layer, and a course extension: ephemeral tools that the engine calls automatically, injecting their output into the context window without declaring them to the LLM.

A browser snapshot makes the distinction useful. The model needs to know which
tab is selected before deciding what to click. Asking it to remember to request
a snapshot adds a tool round and leaves correctness dependent on its memory.
Automatic context puts the observation where the decision needs it. The tool
still does real work, and the request must wait for that work to finish.

## TL;DR

The client sends `initialize`, receives capabilities, sends
`notifications/initialized`, then calls `tools/list` to discover what the server
offers. Admit discovered tools through the owning skill's grants. Ordinary
admitted tools use the existing model declarations and job dispatch; hidden
automatic tools supply completed round or turn context. The bidirectional
channel also carries reverse requests to enabled Agent tools.

The exercise implements the historical wire version `2024-11-05`. Its reverse
`tools/call`, `ephemeral` field and `$/cancelRequest` notification are course
contracts. This chapter does not claim compatibility with every current MCP
server or present those extensions as current standard requirements.

Extend Chapter 10 with replaceable transports, request correlation, actual
job cancellation, hidden automatic context, browser tunneling and skill-owned
server startup/shutdown. Keep the common interfaces, real parent relationships
and optional GUI module. The exercise and check table below cover the same
capabilities through their appropriate clients; `--mcp-pipe` is a dedicated
protocol service, with no implicit model turn. From the repository root, grade
with `go run ./cmd/grade-edition2 -ch 11 path/to/core`.

## Transport

One interface, multiple implementations:

```go
// MCPTransport carries whole messages independently of GUI framing.
// Close must unblock both directions; one receiver and concurrent senders use it.
type MCPTransport interface {
    // Send writes one complete JSON-RPC message.
    Send(msg json.RawMessage) error
    // Recv reads one complete JSON-RPC message.
    Recv() (json.RawMessage, error)
    // Close releases the boundary and wakes blocked I/O.
    Close() error
}
```

**Pipe**: an in-process connected pair, for testing the actual client and codec.
**Stdio**: spawns a subprocess, JSON-RPC over its stdin/stdout.
**Raw**: wraps an `io.Reader` and `io.Writer`; the `--mcp-pipe` flag uses this.
**WebSocket**: tunnels JSON-RPC through the GUI server to the browser.

The interface belongs in `internal/common`, with a public alias for external
clients. Raw, pipe and subprocess behavior belongs in the MCP package.
WebSocket framing belongs in the separate GUI module, so a headless consumer
does not acquire a browser dependency merely to call an external tool.

Agent owns the MCP manager; the manager owns its clients. A client reaches
admission, jobs and diagnostics through manager → Agent → Ensemble. The GUI
transport keeps its actual server parent. The composition root establishes
those relationships. A reader and writer are boundary resources, and need no
invented domain parent or bag of sibling services.

Serialize whole outgoing messages: a reverse reply can arrive while another
goroutine sends a request. Closing a transport must wake blocked I/O. For a
subprocess that includes closing its pipes, stopping its process group and
reaping the child; closing only the Go wrapper leaves a server running after
its skill has gone away.

## Codec

The codec handles JSON-RPC correlation: outgoing requests get incrementing integer IDs, responses are matched by ID and delivered to the blocked caller. A background goroutine reads the transport and dispatches: responses go to pending callers, incoming requests go to reverse tool handling.

One client can own the transport, ID counter, pending-reply map and reader.
The accepted implementation keeps correlation there; a separately exported
`Codec` would add an ownership layer with no other user. The mechanism matters
more than the extra object.

Keep the reader free while executing a reverse call. That tool might itself
call the peer, whose reply needs the same reader. Running the handler on the
reader goroutine would let both sides wait forever for a message already on
the wire. Concurrent requests also need an out-of-order reply test: two correct
answers delivered to the wrong callers are two failures.

## MCP Client

The client initializes, discovers tools, calls them and closes its transport.
Its owner supplies reverse dispatch through the existing Agent interface.
The public embedding API exposes `Agent.MCP().Connect(ctx, name, transport)`;
`ensemble.NewMCPTransport(reader, writer, closer)` adapts streams, and
`ensemble.MCPPipe()` returns a connected pair. With skills enabled, the peer's
name must match a declared server. `Connect` performs initialization and discovery
before making the peer's tools available.

`tools/call` sends `params` with a string `name` and object `arguments`. Text
results arrive in ordered `content` blocks. Preserve a tool's `isError` result
as a failed tool outcome; a JSON-RPC error is a protocol failure. Neither should
be turned into a successful empty string.

MCP tool calls go through the same job infrastructure as `run_command`.
If the job is killed, its context is cancelled and the client sends
`$/cancelRequest` with the request ID to the server. A callback deadline only
returns control and a job handle. Interrupt ends the current turn while the
job continues; shutdown stops the owned work and closes peers. One context
cannot represent all four lifetimes.

## Ephemeral Tools

A discovered tool may carry an `ephemeral` field: `"round"` or `"turn"`. Ephemeral tools never appear in the tool declarations sent to the LLM. The model cannot call them. The engine calls them automatically:

- **round**: called before every `RequestSent`. The GUI snapshot arrives fresh each round.
- **turn**: called once when a new turn starts, for data that can stay fixed
  through that turn. A later turn gets a fresh call.

Results are combined and attached through the existing event/reducer path into
`Context.Ephemera`, the pending field consumed by the next request. The log
retains the arrival for replay; stale snapshots do not accumulate in dialogue.
Gather outside the actor so reverse requests, interruption and job facts can
still be processed. Send the gathered content back with its turn identity and
attach it only if that turn remains active.

Ephemeral errors are logged but not fatal. A GUI snapshot failure should not abort a turn.

The first student implementation used an ordinary job wait for this gathering.
An independent peer delayed its turn snapshot by 3.4 seconds. At three seconds,
the callback returned a running notice with `Use wait_for_job`, and the provider
request carried that notice where the snapshot should have been. The once-per-turn
flag had already been set. When the actual result arrived, it was too late to
serve as that turn's automatic context.

A hidden tool has no model-directed caller to follow up on its handle. Its
collector needs the completed content, or an actual failure, before releasing
the provider request. The correction carries the handler's completed result
through a buffered reply while preserving the ordinary callback path. Enlarging
the three-second timeout would merely select a slower peer for the next failure.

Interruption still lets the collector abandon the old turn without killing the
job. The regression holds the old peer reply, interrupts, completes a new turn,
then releases the old result. Its output remains available through the job,
and none of that stale context enters the new turn. A separate slow-peer check
holds both round and turn responses past the callback deadline and verifies that
no provider request goes out early. These tests exercise the real jobs, actor,
reducer and renderer, with only the external peers replaced.

## Bridge

The bridge is the seam between MCP's type system and the agent's. It converts
discovered names, descriptions and input schemas into ordinary tool definitions.
The engine can then dispatch an admitted remote tool through the same path as
a compiled one. Inside the accepted manager's `Tools` method, the handler
captures the client and that iteration's tool information:

```go
tool := common.ToolDefinition{
    ToolDeclaration: common.ToolDeclaration{
        Name: info.Name, Description: info.Description, Schema: info.Schema,
    },
    Ephemeral: info.Ephemeral,
    Run: func(call common.ToolContext, args json.RawMessage) (string, error) {
        return c.call(call.Job().Context(), info.Name, args)
    },
}
```

The job supplies the context. Substituting `context.Background()` here severs
explicit kill from the outstanding RPC even though the handler looks correct
on every successful short call. The original example made that mistake.

Discovery describes what a peer offers; a skill determines what this Agent
admits from that peer. Apply that grant before exposing definitions, then apply
the Agent's enabled-tool policy at declaration and dispatch. Another loaded
skill naming the same tool must not widen this server's grant. Automatic tools
remain absent from model declarations and unavailable to guessed ordinary calls.
The accepted registry also prevents an external definition from shadowing a
compiled capability with the same name.

## Reverse Calls

The MCP channel is bidirectional. When the server sends a `tools/call` request,
the client's reverse handler invokes the Agent's enabled-tool dispatch. The
result goes back with the request's ID. A browser, a Python script or a remote
service can call `read_file`, `run_command`, or another enabled ordinary tool,
subject to the Agent's trust and skill boundaries.

Treat admitting such a peer as granting access to those capabilities. A server's
own offered-tool filter limits what it exports; it is not a filesystem sandbox
for reverse calls. Only connect a peer you trust with the enabled Agent tools.
Confinement arrives later. Unknown, disabled and hidden automatic names must
still fail when guessed over this channel.

Reverse execution uses real jobs and retains their facts. It does not fabricate
a model-authored tool call in conversation just because the caller used the
same method name. Its reply may legitimately be an ordinary running-job notice;
the automatic collector described above has a different completion requirement.

## WebSocket Tunneling

The browser cannot open a port or spawn a subprocess. The GUI server already
carries `prompt`, `hint`, `interrupt`, and `settings` messages. MCP adds one more
type: `jsonrpc`. The server routes `{"type":"jsonrpc","payload":{...}}`
messages between the Go MCP client and the browser's MCP server over `/ws`.
Replies use the same envelope and retain the JSON-RPC ID. Give each peer reliable
delivery through its connection's serialized writer; a lossy progress broadcast
cannot carry an awaited RPC response. The current implementation uses one active
browser MCP target, while other tabs retain ordinary chat and settings access.

On the browser side, `mcp.js` handles these frames and implements the chapter's
`initialize`, `tools/list` and `tools/call` methods. It registers four tools:

- **gui_snapshot** (ephemeral/round): walks the visible DOM and returns a Markdown summary: pane layout, interactive elements with CSS selectors, artifact previews. Capped at 4096 UTF-8 bytes. The cap is reported rather than silent: the summary states how many characters it elided and how many artifacts it omitted, and it reports the state attributes of every control it lists. An observer that truncates in silence will tell you a screen looks fine when it never saw it, and a control whose state it cannot read is a control it will guess about.
- **gui_click(selector)**: dispatches a click event on the matched element.
- **gui_input(selector, text)**: sets the value and dispatches input/change events.
- **tts_queue** (ephemeral/round): returns pending TTS utterances as JSON: text, state, timing.

The agent receives a textual observation of the visible GUI, updated every
round. It can click buttons and fill text fields, then inspect the changed
state on the next request. The speech queue reveals what the application is
trying to say and whether an utterance has started. It cannot tell whether a
human heard intelligible audio; that still requires listening.

## SKILL.md Integration

Skills declare MCP servers in their frontmatter. For example, save this as
`skills/browser-tools/SKILL.md` and list `browser-tools` in your primary skill's
`loadable-skills` field:

```markdown
---
name: browser-tools
description: Inspect and operate the visible browser interface.
type: loadable
tools: gui_snapshot gui_click gui_input tts_queue
mcp_servers:
  - name: browser-debug
    transport: websocket
---
Use the current snapshot's selectors to operate visible controls.
Check the next snapshot before claiming the interface changed.
```

A separate code-search skill can declare this fragment, provided
`scripts/search_server.py` implements the server and offers `search_code`:

```yaml
tools: search_code
mcp_servers:
  - name: code-search
    transport: stdio
    command: python3
    args: ["scripts/search_server.py"]
```

Loading a skill starts its MCP servers and discovers their tools. The
`transport` field determines the wire: `stdio` spawns a subprocess, `websocket`
tunnels through the GUI. Browser discovery requires both a declared skill and
a connected browser; an already-open browser can serve a later skill load.
The `tools` grant must include automatic tools as well as ordinary ones.

Arguments are a JSON-style string array, passed literally without shell
expansion. Relative script paths resolve from the Agent's workspace. Keep
provider credentials with the model client; the accepted subprocess launcher
passes only basic OS environment values to an external server. The limited
frontmatter parser handles these declared fields without becoming a general
YAML implementation.

Unloading stops that skill's external servers and removes their executable
tools immediately. Chapter 9's compiled tools and retained instruction text
still follow pending removal. Those lifetimes differ because a retained manual
does not require keeping a subprocess alive. Compaction is still a later
feature; this chapter adds no new way to erase the retained manual.

## Exercise

Extend the accepted Chapter 10 agent. Implement the transport/client/bridge,
automatic round and turn tools, reverse calls, browser tools/tunnel and skill
server lifecycle above. Preserve prior behavior through the same Agent runtime.
Build from the core module with `go build -o /tmp/ensemble-ch11 .`. Begin
with a protocol peer connected to that executable:

```sh
/tmp/ensemble-ch11 --mcp-pipe
```

The flag dedicates stdin/stdout to one JSON-RPC 2.0 object per newline. The
client initiates the handshake; replies preserve integer request IDs. Keep
diagnostics off stdout. This service mode does not start a conversation, so
successful initialization alone cannot prove model-driven tool execution.

Then use normal chat with a skill-declared stdio peer. Inspect the actual
provider declarations and correlated results, including an ordinary tool error.
Verify that round content refreshes before each request, turn content runs once
per new turn, automatic names stay hidden and a failed automatic call is logged
without aborting the turn. Check reverse execution with an enabled file read
and refusals for unavailable names. Exercise the separate GUI's real tunnel
in both directions.

| Check | Points | Observable |
|-------|--------|------------|
| mcp-handshake | 15 | initialize, initialized notification and discovery sequence |
| tool-discovery | 15 | actual discovered schemas and admitted model-visible tools |
| mcp-tool-call | 15 | forwarded arguments and correlated real result/error content |
| ephemeral-round | 20 | hidden round/turn context, refresh timing and nonfatal errors |
| reverse-call | 15 | enabled tool execution and unknown-tool refusal |
| ws-tunnel | 10 | JSON-RPC through the actual separate GUI connection |
| ch10-parity | 10 | Chapter 10 through its appropriate clients |

```sh
go run ./cmd/grade-edition2 -ch 11 path/to/core
```

Omit the final path to grade `solutions/ch11`. Existing `-gui DIR`,
`-gui-model` and inherited fixture options remain available. Keep focused source
tests for ownership, out-of-order correlation, slow automatic completion,
callback versus kill/interrupt, per-server grants and subprocess unload/reaping.
The numeric checks do not establish all those properties. Browser DOM and native
speech need an actual browser in addition to a faithful JSON-RPC peer.

## Taking It for a Spin

Give the agent an external capability small enough to inspect. In a fresh
scratch directory, write a word of your choice into `marker.txt`, then save
this server as `peer.py`. It implements the historical course wire and records
its call counts locally so the model's account can be checked:

```python
import json
import pathlib
import sys

counts = {}
tools = [
    ("read_marker", "Read the exact marker file text", ""),
    ("environment_snapshot", "Current live snapshot", "round"),
    ("session_brief", "Session configuration", "turn"),
]
for line in sys.stdin:
    message = json.loads(line)
    if "id" not in message:
        continue
    method = message.get("method")
    if method == "initialize":
        result = {"protocolVersion": "2024-11-05", "capabilities": {"tools": {}}}
    elif method == "tools/list":
        result = {"tools": [
            {"name": name, "description": description, "ephemeral": mode,
             "inputSchema": {"type": "object", "properties": {}}}
            for name, description, mode in tools
        ]}
    elif method == "tools/call":
        name = message["params"]["name"]
        counts[name] = counts.get(name, 0) + 1
        pathlib.Path("peer-counts.json").write_text(json.dumps(counts))
        if name == "read_marker":
            text = pathlib.Path("marker.txt").read_text()
        elif name == "environment_snapshot":
            text = f"Snapshot number {counts[name]}"
        else:
            text = "Use read_marker to inspect the local marker."
        result = {"content": [{"type": "text", "text": text}]}
    else:
        continue
    print(json.dumps({"jsonrpc": "2.0", "id": message["id"],
                      "result": result}), flush=True)
```

This little demonstration peer serves only the listed tools. It is not a
general server implementation or a cancellation test. Save the following
primary skill in `skills/base/SKILL.md`:

```markdown
---
name: base
description: Read a local marker through the demonstration MCP peer.
type: primary
tools: read_marker environment_snapshot session_brief
mcp_servers:
  - name: local-observer
    transport: stdio
    command: python3
    args: ["-u", "peer.py"]
---
Read the marker through read_marker when requested.
Automatic snapshot data describes the current round.
```

Build in your core module, then launch from that scratch directory. Set the
usual `LLM_VENDOR`, available `LLM_MODEL` and `LLM_API_KEY` in the child
environment without putting the key in a command line or transcript:

```sh
go build -o /tmp/ensemble-ch11 .
```

```sh
EN_SKILLS_DIR=./skills EN_PRIMARY_SKILL=base /tmp/ensemble-ch11 chat
```

Enter:

> Use read_marker exactly once. Report the exact marker and current automatic
> snapshot number in one sentence. Do no other work.

After the answer, close stdin. Compare the reported marker with `marker.txt`
and inspect `peer-counts.json`. The ordinary call reads the file; the round
tool updates automatically before the initial request and before the request
carrying that result. The turn tool runs once. A model may take extra rounds,
so check the actual requests and counts rather than treating an expected count
as proof of what happened.

The retained October 10, 2026 sessions used this pattern with
`claude-haiku-5-5`, `gpt-4o-mini` and `gemini-3.8-flash`. Each successful run
made one marker read, two round snapshots and one turn snapshot, and each final
answer reported the correct marker and snapshot number 2. The
[Claude](../solutions/evidence/ch12/anthropic-live.json),
[OpenAI](../solutions/evidence/ch12/openai-live.json) and
[Gemini](../solutions/evidence/ch12/gemini-live.json) receipts retain
the actual inputs, replies and requests. These are historical model choices;
select an available model when you run the exercise. The published peer and
skill make that pattern reproducible without claiming they are byte-for-byte
copies of the original launcher's generated files.

The first Claude attempt produced no answer. Its process exited zero while
stderr reported `invalid streamed Anthropic content`. The inherited streaming
parser had replaced valid initial `{}` arguments with an empty accumulated
`input_json_delta`. Keeping the initial input when no JSON bytes arrived fixed
the call. One bounded repeat succeeded and captured a valid tool-use start block
followed by `partial_json: ""`. The initial attempt had retained requests but
not response frames; only the repeat supplies that wire receipt. A regression
using the captured stream reproduces the parser failure when the defect is
restored.

The browser has separate evidence. Fifteen checks ran actual Chrome through
the Go tunnel and Agent with a local provider fake. A tool click selected the
Settings tab; tool input changed the font value and computed size. Later
snapshots exposed the selected state, and a native utterance appeared with queue
and start timing. A large snapshot stayed within the byte cap while disclosing
elided characters and omitted artifacts. The [browser receipt](../solutions/evidence/ch12/browser.json)
preserves these observations. The font assertion proves applied and effective
state; it never reads the settings file. Disk persistence has its separate
Chapter 8 restart check. Nor do those browser checks turn the live CLI sessions
into live-provider browser demonstrations.

Check the actual result when testing the bridge. One deliberately broken
argument-forwarding implementation initially passed because the assertion
searched the whole provider request. It found the marker in the earlier tool
call's arguments even though the result was wrong. Restricting the assertion
to the correlated result blocks exposed the defect. A request contains both
what the model asked for and what the tool returned. Match the assertion to
the fact it needs to establish.

The original full grader was interrupted and unscored, with exit 143. A
separate unchanged current-chapter probe scored **80/90** on interim source;
recursive parity was unrun. Its tool-call and ephemeral credits did not execute
those behaviors, and its WebSocket check searched fixed filenames rather than
running a tunnel. Independent integration, browser and live evidence established
the valid obligations, including the later slow-peer correction. The
[original review](../docs/edition-2-notes/ch12-code-review.md) preserves those
limits and source versions; the [Edition 2 grader report](../docs/edition-2-notes/enhancement-11-code.md)
records revised results separately. A handshake opens the channel. The useful
test follows a call through it and reads what came back.
