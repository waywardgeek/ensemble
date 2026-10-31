# Chapter 11: MCP: The Extension Protocol

Chapter 10 lets a conversation survive its process. It does not bring the old process’s capabilities back to life. A saved tool result is still a fact about past work, not a connection to the program that did it.

MCP, the Model Context Protocol, introduces that connection as a separate responsibility. An Agent can discover tools in an external process or browser, admit them under a skill, and call them through its existing jobs. The peer can also call back into the Agent. What looks like a convenient way to avoid recompiling tools is therefore a new trust boundary and a new lifecycle to own.

The [source chapter](../edition-3/chapter-11.md) develops a client for the historical wire version `2024-11-05`, using JSON-RPC 2.0. Its reverse `tools/call`, `ephemeral` field, and `$/cancelRequest` notification are course extensions, not claims about what every current MCP server supports.

## Discovery Is an Offer

The ordinary connection begins with `initialize`. The peer replies with its capabilities, the client sends `notifications/initialized`, and `tools/list` supplies the available tools. Each offer has a name, description, and input schema. The bridge turns admitted offers into the same kind of tool definition the Agent already knows how to dispatch.

There is no need for a second execution system. When the model calls an admitted external tool, the existing job machinery invokes a handler that retains the peer connection and that tool’s information. The handler sends `tools/call` with a string `name` and an object of `arguments`; the returned text blocks retain their order. A tool result marked `isError` is a failed tool outcome. A JSON-RPC error is a protocol failure. Converting either into a successful empty string would make the bridge easy to use by making its failures impossible to understand.

Before any of this becomes model-visible, the owning skill supplies the grant. Discovery says what the server offers; the skill says what this Agent may use from that server. Another loaded skill that happens to name the same tool must not widen this server’s grant. The Agent’s enabled-tool policy then applies at both declaration and dispatch, and an external definition cannot shadow a compiled capability with the same name. Checking only the declarations would leave a guessed name able to cross a boundary the model was merely encouraged not to notice.

The public embedding path follows the same rule. `Agent.MCP().Connect(ctx, name, transport)` completes initialization and discovery before exposing tools. With skills enabled, the peer name must identify a declared server. Supplying a transport is not a way to bypass admission.

That admission has a second consequence. The channel is bidirectional: a server can send its own `tools/call` request and reach enabled ordinary Agent tools such as `read_file` or `run_command`. The client dispatches through the Agent and returns a reply bearing the server’s request ID. Unknown, disabled, and hidden automatic tools remain unavailable even if the peer guesses their names.

A server’s offered-tool filter limits what it exports. It does not confine what the server can do through reverse calls. Connecting a peer means trusting it with the Agent capabilities enabled at that boundary; a grant for a remote search tool is not a filesystem sandbox for the process offering it. Confinement is later work.

## One Channel, Several Owners

The transport carries complete JSON-RPC messages. An in-process pipe pair exercises the real client and codec without a subprocess. Stdio connects to a child process; a raw adapter wraps supplied streams; a WebSocket transport carries the same messages through the GUI to a browser.

Those variants share a small boundary: send one complete message, receive one complete message, and close. One receiver consumes the stream while concurrent senders may produce requests and reverse replies. Whole-message serialization matters because those sends can overlap.

The common transport interface and its public alias let external clients connect without importing the GUI. Raw streams, pipes, and subprocess behavior belong to the MCP package. WebSocket framing belongs to the separate GUI module. A headless application should not acquire browser machinery to call a tool in a Python process.

Ownership follows the runtime rather than a new collection of sibling service pointers. The Agent owns the MCP manager, which owns its clients. A client reaches admission, jobs, and diagnostics through manager, Agent, and Ensemble. A GUI transport retains its actual server parent. The composition root establishes those relationships; a reader and writer are boundary resources, not reasons to invent another domain object.

Within a client, one owner can hold the transport, incrementing integer request IDs, pending-reply map, and reader. Replies are matched by ID and delivered to the appropriate waiting caller. The source’s accepted design keeps correlation there rather than exporting a separate codec with no other user.

Reverse requests make the reader’s job particularly important. It must arrange execution without becoming the executor. A reverse tool can itself call the peer; if the handler occupies the sole reader, its nested call waits for a response that nobody is free to read. The peer can have answered perfectly and both sides can still wait forever. Keeping the reader available also permits out-of-order replies, which must reach their original callers rather than whichever goroutine happens to be waiting first.

Closing this boundary must wake blocked sends and receives. For a subprocess, closure includes its pipes, process group, and reaping the child. Closing only a Go wrapper can leave an external server alive after the skill that authorized it has disappeared. The lifecycle is not complete until the resources on the other side have been dealt with.

## The Job Owns the Call

The bridge has an easy-to-miss line of responsibility: the outstanding RPC uses the job’s context. The original example used `context.Background()` instead. That version can return every short successful call correctly while making explicit kill unable to reach the call still waiting on the wire.

In the chapter’s design, killing the job cancels its context and sends the peer `$/cancelRequest` with the outstanding request ID. The notification belongs to this wire contract; sending it is not, by itself, evidence that an arbitrary server has stopped its work. The Agent still has to release its own waiter and manage its owned resources.

A callback deadline means something else. It returns control with a job handle while the work continues. Interrupt ends the current turn without killing that job. Shutdown stops owned work and closes peers. These are four different decisions, even when they occur during the same slow RPC. Giving them one context would make an observation timeout cancel useful work, or make shutdown wait on work it can no longer control.

Reverse calls use those real jobs too. They leave execution facts in history without fabricating a model-authored tool call merely because the peer used the same method name. A reverse caller may legitimately receive a running-job notice and follow up. The next mechanism has no such caller, and that difference exposed a particularly revealing failure.

## A Hidden Tool Has Nobody to Ask Again

Some observations should be collected before the model decides anything. A peer can mark a discovered tool `ephemeral: round` for a fresh observation before every `RequestSent`, or `ephemeral: turn` for information collected once at the start of each new turn. The owning skill must grant these tools too. Automatic execution is not an exemption from admission.

The model neither sees their declarations nor calls them. Their completed results enter `Context.Ephemera` through the existing event and reducer path, to be consumed by the next request. The event log retains the arrival for replay, while stale observations do not pile up as dialogue. The turn-scoped call runs once for that turn; it is not a license to reuse its data in the next one.

Hiding the ordinary declaration is only the beginning. Chapter 9’s deferred catalogue and skill-loading `tool_reference` blocks can also be built from the capability snapshot. Automatic names must stay out of all those provider-facing surfaces and remain unavailable to guessed ordinary calls. The internal snapshot, however, must retain them so the engine can collect them. Deleting them from the registry would solve disclosure by disabling the feature.

The source records a first implementation that gathered automatic context using an ordinary job wait. An independent peer delayed its turn snapshot by 3.4 seconds. At three seconds the callback returned a running notice containing `Use wait_for_job`, and the provider request carried that notice where the snapshot belonged. The once-per-turn flag was already set. The actual result arrived after the request that needed it.

Every local component had done something plausible. The job kept running. The callback returned promptly. The engine attached the returned text. What failed was the assumption that returned text necessarily meant completed observation. A hidden tool has no model-directed caller who can see a handle and ask again. Increasing the timeout would merely move the failure to a slower peer.

The correction preserves the ordinary callback path while giving the automatic collector the handler’s completed result through a buffered reply. The provider request waits for actual content or an actual failure. An observation failure is logged but does not abort the turn; a running notice is neither the observation nor its failure.

Waiting must not freeze the Agent. Collection happens outside the actor so reverse requests, interruption, and job facts can still be processed. The completed collection returns with its turn identity and attaches only if that turn is still active. Interrupt can therefore abandon the collection for the old turn without killing the underlying job. The job’s eventual output remains available, but it does not become fresh context for a different turn.

The source’s regression deliberately held an old peer reply, interrupted, completed a new turn, and then released the old result. Another check held both round and turn replies beyond the callback deadline and looked for an early provider request. These tests used the real jobs, actor, reducer, and renderer with substituted external peers. They probe the relevant ordering, not every possible server or provider. The important result is a distinction the fast path conceals: completion gates a request, while turn identity decides whether that completion still belongs to it.

## The Browser Is an On-Demand Peer

A browser cannot provide the same subprocess boundary. The GUI already carries prompt, hint, interrupt, and settings traffic over `/ws`; an additional `jsonrpc` envelope tunnels MCP messages through that connection. Replies retain their JSON-RPC IDs. Each peer needs reliable delivery through its connection’s serialized writer. A lossy progress broadcast is acceptable for some views of activity, not for the only copy of an awaited RPC reply.

The implementation described in the source selects one active browser MCP target. Other tabs retain ordinary chat and settings access rather than becoming competing recipients of a tool call. Browser discovery needs both a declared skill and a connected browser; an already-open browser can become the peer when its skill is loaded later.

The browser module offers four ordinary tools. `gui_snapshot` walks the visible DOM and returns a Markdown account of pane layout, interactive elements and their CSS selectors, and artifact previews. `gui_click` dispatches a click on a matched element. `gui_input` sets a value and dispatches input and change events. `tts_queue` returns pending speech text, state, and timing as JSON.

The snapshot is bounded at 4096 UTF-8 bytes, but the missing material must not disappear silently. It reports elided characters and omitted artifacts, and includes the state attributes of every control it lists. A selector without state may identify exactly the right control while leaving the Agent to guess whether it is checked, disabled, or already set. A truncated snapshot without disclosure can turn an incomplete view into an unwarranted claim that the screen looks fine.

The intended interaction is observation, action, then observation again: inspect a visible control, use its selector, and obtain a fresh snapshot to check what changed. Speech-queue inspection is similarly explicit when pending speech or timing matters. Queue state is not evidence that a human heard intelligible audio. That requires listening.

Earlier browser observations were automatic. The source reports that the crossover found per-round snapshots and speech-queue checks noisy and unnecessary, so the revised design makes all four tools ordinary admitted declarations. This is a reason to change the scheduling of those observations, not to remove automatic context collection generally. A changing environment may justify a fresh observation before every decision; a browser snapshot is useful when the work needs a browser snapshot. Both consume work and attention, so the distinction belongs in the capability’s design.

## A Skill Starts More Than Instructions

Skill frontmatter names each server and its transport. Stdio supplies a command and argument array; WebSocket selects the browser tunnel. Loading the skill starts its servers, discovers their offers, and applies its grants. For a code-search server, for example, the skill can authorize `search_code` without admitting every capability the process advertises.

Arguments are literal strings, not shell fragments. Relative script paths resolve from the Agent’s workspace. Provider credentials remain with the model client; the subprocess launcher described in the source passes only basic operating-system environment values to external servers. Avoiding shell expansion and credential inheritance narrows accidental exposure, but neither establishes a sandbox for the peer. The frontmatter parser handles the declared fields without taking on a general YAML implementation.

Unloading removes that skill’s external executable tools and stops its servers immediately. This deliberately differs from Chapter 9’s pending removal of compiled tools and retained instruction text. A manual can remain in context without consuming a process or retaining an external connection. Keeping the manual is not a reason to keep the server alive, and stopping the server does not imply that the manual has been erased. Compaction still belongs to later work.

This also completes the distinction from persistence. Conversation state can remember that a tool was used, and retained instructions can describe how it worked. Neither fact starts its server, reestablishes trust, or makes its handler executable. Those are decisions of the current Agent and its current skills.

## Follow the Result, Not the Appearance of Success

A dedicated `--mcp-pipe` service reserves stdin and stdout for newline-delimited JSON-RPC objects, keeps diagnostics off stdout, and initiates the handshake without starting a conversation. It can establish that two peers agree on framing and initialization. It cannot establish model-driven tool execution because there is no model turn in that mode.

The source’s small demonstration peer makes the next boundary visible. An ordinary tool reads a marker file, a round tool supplies a changing snapshot, and a turn tool supplies a stable session brief. Local call counts allow comparison with actual provider requests rather than reliance on the model’s account. The round observation can run before the initial request and again before the request carrying the ordinary tool’s result; extra model rounds can add calls. An expected count is not evidence of the rounds that actually occurred. The demonstration peer serves only its listed tools and does not test cancellation.

Historical Claude, OpenAI, and Gemini receipts record successful runs of this pattern. They are evidence of those runs, not current model availability or comprehensive provider coverage. The first Claude attempt is more instructive than a list of successes: it exited zero without producing an answer. Its parser had replaced valid initial `{}` arguments with an empty `input_json_delta`. Keeping the initial input when no JSON bytes arrived fixed the call. A successful repeat captured the empty delta, and a regression reproduced the failure when the defect was restored. Protocol completion, process success, and a useful answer had been three different facts.

A separate bridge test initially accepted deliberately broken argument forwarding. Its assertion searched the whole provider request for the marker, so it found the marker in the earlier call’s arguments even though the returned result was wrong. Restricting the assertion to the correlated result blocks exposed the defect. A request can contain both what the model asked for and what the tool actually did. Evidence for one cannot be borrowed from the other.

The browser evidence has its own boundary. The source’s historical receipt used real Chrome, the Go tunnel, and the Agent with a local provider fake. It observed control changes, native speech start, and disclosed snapshot truncation, but exercised the old automatic observations rather than the revised on-demand path. Applied font state did not prove persistence, and native speech start did not prove intelligible audio. It was not a live-provider browser demonstration.

The source also records earlier checks that credited tool calls and ephemeral behavior without executing them, and a WebSocket check that searched fixed filenames instead of running a tunnel. Later reports do not retroactively make those checks adequate or establish Edition 3 acceptance.
