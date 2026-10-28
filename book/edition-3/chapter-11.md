# Chapter 11 — MCP Tools and a Browser That Can Answer

Until now, a tool has lived inside the application. This chapter adds tools that live in another process or in the browser. The Agent still owns admission, execution records, Jobs, and conversation. A remote server supplies a capability; it does not become a second Agent scheduler.

We use the **Model Context Protocol (MCP)** for discovery and calls. Its JSON-RPC messages give us a small, inspectable boundary: initialize a connection, discover tools, call one, and correlate its reply. We use that boundary in both directions. An Agent can call a skill's server, and a peer can ask the Agent to execute an enabled local tool.

The purpose is not merely to make a remote function return a string. We must preserve the architecture built in earlier chapters:

- Common declares the vocabulary and persistent relationships; functional partial classes implement operations.
- A discovered name is not a grant to execute it.
- The actor remains the only publisher of conversation and skill state.
- A Job's lifetime is independent of a caller's observation wait.
- Destruction closes admission, joins owned work, and removes both ends of every relationship.
- Browser inspection is on demand, not an unsolicited attachment to every request.

## 1. Put the boundary in the right place

The core remains one C# assembly. `Common/Mcp.cs` and `Common/McpBrowser.cs` declare the MCP vocabulary. The `Mcp/` directory implements transport, RPC, discovery, content conversion, and connection ownership. `Skills/` prepares and commits skill activations; `Agent/` schedules requests and receives worker completions; `Jobs/` runs remote calls through the existing Job machinery.

The optional GUI project has its own Common declaration and server implementation for the WebSocket tunnel. Neither the MCP library nor the Agent depends on ASP.NET or JavaScript.

The small `IMcpTransport` interface is a real external I/O boundary, not glue between partial classes. It sends and receives JSON messages and disposes its transport resources. Production implementations cover line-oriented streams, an owned child process, and a connected WebSocket. Tests use working peers at those same boundaries.

### The relationship graph

| Object | Parents and responsibility |
| --- | --- |
| `Mcp` | Belongs to one Agent; owns its connections and optional browser registration. |
| `McpClient` | Belongs to the manager; a skill connection also belongs to `Skills` under its `SkillName`; a browser client also belongs to its `McpBrowser` registration. Owns the RPC reader, pending replies, reverse-request work, transport, and remote Tools. |
| Remote `Tool` | Belongs to the Agent and the client that supplies it. Calls still execute as Agent-owned Jobs. |
| `McpBrowser` | Belongs to the manager; owns one browser transport and at most one claimed client. |
| `McpToolDefinition` | Immutable, value-like discovery data: name, description, schema, and optional automatic frequency. It needs no holder back-pointers. |

These are actual object references. A client does not rediscover its Agent by a global ID lookup or receive a copied bag of services. It reaches the Agent through its manager and reaches its granting owner through `SkillOwner()`.

Parent add/remove methods validate before mutation and update both ends. Removing a relationship is not the same operation as destroying the child. Destroying a connection destroys its remote Tools, closes the transport, releases pending callers, joins reader and reverse work, and detaches the client from all its parents. Repeated destruction calls join the same cleanup task.

Admission closes when destruction is requested, not when a background cleanup task happens to run. Otherwise a caller could start another RPC or reattach an object that is already being dismantled.

## 2. One protocol, two explicit roles

Do not confuse an application that *serves a pipe* with an outgoing MCP client. This chapter supports both roles, but not both reader modes on the same connection.

### Outgoing skill connection

For a skill server, Ensemble initiates the handshake:

1. Send `initialize` with protocol version `2024-11-05`.
2. Validate the reply.
3. Send `notifications/initialized`.
4. Request `tools/list` and validate the complete catalog.
5. Install only the ordinary tools granted by the owning skill.

A connected transport is not yet a successfully initialized skill. Missing schemas, duplicate names, incompatible protocol replies, process exit, or a failed catalog cause preparation to fail and its resources to be joined and removed.

Requests use integer IDs and a pending-reply table. A reply completes its matching request, not whichever caller happens to be waiting first. The outgoing reader must also recognize peer requests while it awaits replies: a server may reverse-call the Agent during initialization or a tool call.

Cancelling a reply wait removes that wait without corrupting correlation when a late reply arrives. A failed or cancelled *write* is different. It may have emitted half a frame, so the connection is closed rather than reused as though framing were intact.

### Incoming CLI service

The CLI's `--mcp-pipe` mode exposes enabled Agent tools on stdin/stdout. Here the external client sends `initialize`, receives the reply, and sends `notifications/initialized` before discovery or calls.

```sh
# After building Cli/Ch04.csproj; use your approved dotnet executable.
dotnet Cli/bin/Debug/net10.0/ch04.dll --mcp-pipe --no-save
```

Normal skill and log startup still happens first. Stdout contains JSON-RPC frames only; diagnostics use stderr. `ServeAsync` uses the existing reverse-tool admission path, not an independent tool dispatcher or fabricated model turn. EOF returns from serving; the application then joins connection destruction and Agent shutdown. An incoming service cannot initiate outgoing RPC on that same transport.

This distinction resolves an easy testing mistake: an outgoing fixture waits for Ensemble to initialize it, whereas an incoming-service fixture must initialize Ensemble itself. Reversing those roles produces a deadlock, not useful protocol evidence.

## 3. Reverse calls must not wait behind themselves

Suppose the Agent is waiting for a server's `tools/call` reply. The server asks the Agent to read a file before it can finish. If that reverse request is queued behind the waiting model turn, neither side can proceed.

Reverse calls therefore enter a nonblocking actor control lane. The actor still validates the name, current skill grants, and arguments, and executes through ordinary Jobs. A disabled tool remains disabled; being on a trusted transport is not a grant.

Discovery exposes enabled local tools, excluding automatic snapshot tools. Remote tools are not reflected back as local reverse capabilities. A successful reverse call records a host observation, not an assistant tool-use exchange that never occurred. Unsupported methods return JSON-RPC errors. Expected bad arguments and denied calls return controlled failures; invariant and persistence faults must not be disguised as successful tool results.

The important test is a real held operation: leave a model turn or server call waiting, issue a reverse `write_file`, verify the file and its host observation, and only then release the original operation. A mocked assertion that a callback was invoked would not establish progress through this cycle.

## 4. Skill activation is a transaction

A skill may declare servers in its front matter:

```yaml
---
name: notes
description: Read and update project notes
type: loadable
tools: notes_read notes_write
mcp_servers:
  - name: local
    transport: stdio
    command: /absolute/path/to/notes-server
    args: ["--format", "json"]
---
Use these tools for project notes. Read before replacing existing content.
```

This is the supported front-matter subset, not arbitrary YAML. Server names, transport spelling, field types, and JSON string-array arguments are validated. Use literal argument elements; this is not a shell command string. The process inherits the configured working directory and a small credential-free environment, not the application's provider secrets. Stderr is drained separately from protocol stdout.

For a browser server, the declaration is simply:

```yaml
mcp_servers:
  - name: browser
    transport: websocket
```

A browser does not need a process command. Server names are unique within their real skill owner. Two unrelated skills may both name their server `local`; Agent-wide tool-name collisions are still refused.

### Prepare outside the actor; commit on the actor

Loading a server-backed skill requires asynchronous I/O, but I/O continuations must not publish grants or instructions. Activation proceeds as follows:

1. Resolve and validate the dependency closure on the actor.
2. Capture the variables and working directory needed by preparation.
3. Start or claim the connections, initialize them, and discover their catalogs off the actor.
4. Send one reliable commit message back to the actor.
5. Revalidate the closure, cancellation, providers, and collisions; render the frozen manuals; install only explicit allowlisted ordinary tools; append one dependency-first activation event.
6. On failure, destroy and join the prepared connections without publishing a partial activation.

Once a commit is queued, preparation awaits its definite outcome. Cancelling an observer must not race successful publication and destroy clients the actor has just committed.

The server's catalog cannot broaden the skill's `tools` list. A declared name must resolve unambiguously; discovery is evidence of availability, not authority. `$TOOLS` is rendered from usable ordinary capabilities. Loaded manuals remain in provider instruction context, not buried only in a tool-result body.

### Primary startup

Local-only primary skills keep their synchronous initialization path. A server-backed primary requires explicit `Skills.InitializeAsync` after persistence loading and diagnostic-log setup, before user work is accepted. Concurrent callers join one initialization; cancelling one observer does not cancel shared startup.

Until the primary commits, the manager is unready and grants no provisional tools. Input and direct tool work are refused; actor controls remain responsive. Reverse calls during initialization see only actually committed grants. Startup is tracked control work, not a fake `load_skill` call with an invented call ID.

A GUI must listen before a browser-backed primary can connect. `RequiresBrowserStartup()` examines the primary dependency closure: only a closure containing a WebSocket server needs that wait. A stdio-only primary initializes without any browser connection. The Agent's readiness gate remains authoritative in both cases.

### Unload and persistence

Unloading closes the selected skill's clients, processes, and remote Tools immediately. The manual's pending-unload state remains under the earlier chapter's conversation policy; retaining a manual does not retain executable authority. Independently owned dependency connections are not destroyed merely because another skill used them.

Chapter 10's saved configuration is documentary, not executable state. Shutdown may save a configuration projection after remote Tools have been removed. A later startup rebuilds connections and grants from current configuration and manuals; it does not resurrect a process or browser connection from saved tool JSON. Failed primary startup must not overwrite a previous good session save.

## 5. Remote results remain structured

An ordinary MCP call is still a Job. Observation timeouts report that it is running; they do not cancel execution. Explicit Job cancellation cancels the remote request and sends the cancellation notification. A reverse call or ordinary wait does not borrow ownership of the Job lifetime.

MCP content is an ordered array, not one lossy string. The codec maps supported text, image, and audio blocks into the existing neutral parts, preserving order and `isError`. Inline media uses validated base64 and MIME information. Unsupported blocks, including general resource blocks, are refused explicitly rather than silently dropped or fetched.

The Job retains an atomic structured result separately from incremental process-output bytes. Its diagnostic output is written before completion publication. The first completed ordinary observation returns the structured parts and error flag; later observations do not replay the same result. Automatic collection uses a separate completion read that does not consume the public observation cursor.

Provider rendering is a separate boundary. An adapter supporting inline media in tool results can render those parts; another adapter may explicitly refuse that placement. Passing an MCP codec test does not establish that every provider accepts every media form. This chapter does not add remote URI fetching or Job-handle media resolution.

## 6. Automatic tools: wait for the snapshot, not the handle

Some external tools provide changing context rather than a model-chosen action. This implementation recognizes `per_turn_` and `per_round_` name prefixes, and compatible discovery annotations:

- `per_turn_`: collect once at the start of a turn and retain its value only for that turn.
- `per_round_`: collect before each model request, including a later request after ordinary tool results.

Conflicting prefix and annotation frequencies are refused. Only definitions granted by their actual active skill owner are eligible.

Automatic tools are **not ordinary callable tools**. They never appear in provider declarations, `$TOOLS`, deferred catalogs, or discovery references. Guessed model calls and reverse calls cannot invoke them. Filtering happens at both admission and request emission, including references replayed from older recorded results.

### Collection uses real Jobs

A dangerous shortcut is to treat a timeout-limited observation as the snapshot. That returns a handle or “running” text, then sends the model request without the data it was supposed to contain.

Instead, the collector reserves and starts a real Job, records execution through the actor, and waits for its completed result. It does not replace `tool_limits` with a hidden foreground timeout. While collection is pending, the actor remains responsive to hints, pause, interrupt, and shutdown.

Interrupt abandons the collector's wait; it does not kill an independent Job. That Job can finish later and remain inspectable. Its late result is not inserted into an interrupted turn or allowed to contaminate a later one. Shutdown has the separate authority to terminate and join owned work.

Pause also matters after collection: even if all results are ready, the actor must not start HTTP until scheduling resumes.

### Ephemera is request context, not dialogue

Completed automatic values are combined with the current manual ephemera slot immediately before the request is frozen. The actor records system ephemera; the reducer replaces the slot rather than appending durable conversation. Each request consumes the slot. Diagnostic events may record what arrived, but saved reduced dialogue and provider instruction bodies must not accumulate snapshots.

The per-turn value is reused within that turn; per-round values are replaced each round. The next turn collects fresh values. Failed rendering or HTTP must not leave a stale slot waiting to leak into an unrelated request. Rendering refusal still occurs before the transport-attempt event.

A server-reported snapshot failure is nonfatal context failure, not successful snapshot data. The request receives a short diagnostic naming the automatic tool and real Job handle. Full error content remains in the Job and host observation for inspection. Do not promote failed output into facts the model is asked to rely on.

## 7. The browser is an on-demand MCP server

The GUI already has one WebSocket for display and controls. MCP shares that socket through envelopes:

```json
{"type":"mcp","agent_id":"root","data":{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}}
```

The reply uses the same envelope and Agent identity. The actual identity comes from the server's replay frame; it is not guessed from a tab title. All GUI controls must match that identity before dispatch. Nonmatching controls and MCP envelopes are refused without reaching tools or the model.

There is one socket writer. The tunnel owns a bounded inbound MCP queue and uses the existing outbound queue. It must not silently drop reliable RPC replies. Closing MCP on skill unload does not close unrelated conversation traffic on the socket.

### Registration is not activation

The first eligible browser connection registers its transport. Connecting alone installs no tools and grants no capabilities. Loading the `browser` skill claims that registration, performs the outgoing handshake, discovers the browser catalog, and commits the skill's explicit allowlist.

The registration is one-shot. It cannot be reused as a fresh RPC session after unload. Disconnect removes its claimed client and Tools; a new connection can establish a new registration. Observer tabs may view the conversation but cannot inject replies into another tab's active browser session. A stopped Agent may still accept a replay-only observer, without constructing new MCP ownership.

Browser replies remain tied to the socket that received the request. A delayed callback from an old connection must not send its response through a newly connected socket.

### The shipped browser tools

The shipped skill offers five explicit tools:

| Tool | Purpose |
| --- | --- |
| `gui_snapshot` | Return a bounded description of visible DOM content and usable selectors, with truncation reported. |
| `gui_click` | Click a uniquely selected visible element. |
| `gui_input` | Set an editable element's text and notify the page through its input events. |
| `tts_queue` | Inspect current and queued speech from the page's existing speech state. |
| `tts_say` | Add speech; with `interrupt:true`, cancel current speech and clear queued speech before the replacement. |

Inspect before acting, use selectors actually observed, and inspect again to confirm the result. A DOM snapshot is not pixel vision. A queued utterance is not proof of sound or hearing. The tools omit password values from inspection, but page content can still be sensitive; callers must choose what context to expose deliberately.

There are no automatic `per_turn_*` or `per_round_*` browser snapshots in this skill. The browser can be idle and connected without sending the page or speech queue into model requests. Automatic collection remains available to a different skill with an explicit need for it.

## 8. Build and exercise the actual boundaries

From `solutions/edition-3/main`, use the approved .NET 10 SDK:

```sh
dotnet build Ensemble.csproj
dotnet run --project Tests/Tests.csproj
dotnet build Cli/Ch04.csproj
dotnet build Examples/Gui/Gui.csproj
```

Use an offline restore source and controlled environment when network access is not authorized. Keep test homes, persistence files, ports, and browser profiles separate from a person's running application. Never infer permission to use a paid model or restart another instance from permission to run local tests.

A useful verification sequence is:

1. **Transport and correlation.** Use a real pipe peer to return replies out of order, close during pending RPC, send reverse requests, and emit a partial failed write. Check callers and ownership cleanup.
2. **Skill transaction.** Hold initialization, verify no provisional grants, then release it. Fail a later dependency and verify earlier prepared processes are gone without published manuals.
3. **Primary startup.** Test stdio without a browser, browser-backed startup after listener creation, concurrent initialization observers, shutdown during startup, and preservation of a previous save on failure.
4. **Remote Jobs.** Verify ordered parts, exact error status, one-time observation, explicit kill, and denied/colliding names without effects.
5. **Automatic context.** Capture real loopback model requests across two rounds and another turn. Check cadence, declaration exclusion, pause, interrupt, retained Job output, and ephemera consumption on rendering and HTTP failure.
6. **Incoming service.** Send the handshake over actual streams, list enabled tools, read a real file with an integer correlation ID including zero, refuse a disabled write, and close stdin normally. No model HTTP should be needed.
7. **Browser.** Use an isolated real page, load the skill, inspect and change a known element, inspect speech state, unload, and test a second observer tab's isolation. Exercise a browser-backed primary and stopped-Agent replay too.
8. **Public consumer.** Reference Ensemble from a separate C# project. Construct the graph through public APIs, activate and unload a process-backed skill, use the incoming service, then join destruction and inspect both parent collections.

Use small working fakes for external boundaries, not mocks of the components under test. Preserve failed runs and distinguish fixture mistakes from product defects. A timeout that reveals an incorrect fixture selector is not evidence that a protocol is broken; a green CLI grade is not evidence that a hidden browser tab worked.

### What this endpoint establishes

The C# chapter has offline component, executable, separate-consumer, and isolated real-browser evidence. Its scoped CLI grader covers incoming pipe behavior, MCP discovery and calls, reverse calls, model-directed skill use against local peers, and automatic context. The optional GUI is exercised separately; the scoped grade does not certify Go package topology or browser appearance.

No live paid-provider interoperability or audible/human-accessibility acceptance is claimed. Fake speech verifies queue wiring, not sound. General MCP resource mapping and broad provider media support are not silently credited. Those limits matter even when every permitted check passes.

## 9. What to carry forward

A remote boundary changes latency and failure modes, not ownership. Keep the actor responsible for publication, the Job responsible for execution lifetime, the skill responsible for grants, and the client responsible for its connection.

The most useful questions in review are concrete:

- Can a discovered or detached tool gain authority without a committed owner?
- Can shutdown leave a reader, process, pending reply, or reverse parent link behind?
- Can a snapshot wait return a handle instead of data and still allow HTTP to start?
- Can interruption consume or destroy a Job that it merely observes?
- Can a stale tab, foreign Agent identity, or old socket deliver an accepted reply?
- Does saved configuration describe state, or accidentally masquerade as authority to recreate it?

Answer those questions in code and working exercises, not by trusting an architectural diagram or a passing grade. The next chapter can build on this boundary because its responsibilities remain explicit.
