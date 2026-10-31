# Chapter 11 — MCP Tools and a Browser That Can Answer

Inside the core library, Common declarations and functional partial implementations compile into one assembly. They do not need a network-shaped abstraction to call one another. MCP introduces something genuinely different: another process or browser, with its own failures, request IDs, timing, and lifetime. Here a transport interface earns its keep.

The network gets a vote about availability and completion. It does not get to decide the Agent’s permissions. A remote catalog describes what a server offers; the active skill determines which offerings become usable.

## External peers, internal owners

A skill owns its MCP server records. Those records own clients, pending requests, and a transport that is either an outgoing stdio process or a browser WebSocket tunnel. Agent owns Tools, Tool owns Jobs, and History still records facts through the actor. A remote tool is an ordinary tool with an external implementation, not a parallel execution subsystem.

Core has no ASP.NET or browser dependency. The optional GUI supplies the browser transport. The real Common interface describes sending, receiving, cancellation, and destruction across that external boundary; it does not restore interfaces between every internal module.

JSON-RPC rides these transports with deliberate limits. Stdio stdout contains protocol messages, while stderr is drained separately. Request IDs correlate pending replies; unsupported IDs, malformed traffic, or exhausted bounded queues are faults rather than excuses to lose reliable replies. Destruction closes the owned transport, faults pending work, and joins readers and processes. An unknown response after cancellation may be safely late; an incoherent protocol stream is not.

Outgoing and incoming connections have opposite initialization responsibilities. An outgoing client initializes its peer and discovers tools before activation. An incoming service waits for its peer to initialize it and advertise readiness. Confusing the two creates two sides politely waiting forever. The stdio program inherits a configured working directory and a small credential-free environment, not the application’s provider secrets; literal argument elements are not interpreted as a shell command.

## Reentrancy is a progress requirement

Suppose a remote call cannot finish until it asks Ensemble to read a file. If that reverse request waits behind the actor’s current turn, the turn waits for the call while the call waits for the turn. Nothing about asynchronous syntax alone breaks this cycle.

Reverse requests therefore enter a nonblocking actor control lane. The actor still checks the tool name, current grants, and arguments, then runs work through ordinary Jobs. A trusted transport is not a new grant. Enabled local tools can be discovered; remote tools are not reflected back as local reverse capabilities, and automatic snapshot tools remain excluded.

A successful reverse call produces a host observation. It does not fabricate an assistant tool-use exchange, because no model made that call. Expected denial and bad arguments become controlled failures, while invariant and persistence faults must remain faults. This keeps both progress and the historical explanation of progress honest.

## Activation crosses the boundary only once

Loading a server-backed skill cannot perform all its work synchronously on the actor. Connections, handshakes, and discovery may wait on external systems. They also cannot publish instructions or grants from arbitrary I/O continuations.

The actor first validates the dependency closure and captures the variables and working directory needed for preparation. Preparation starts or claims transports and discovers catalogs away from the actor. A reliable commit message then returns to the actor, which revalidates cancellation, ownership, and name collisions, renders the frozen manuals, installs only explicitly allowed ordinary tools, and records one dependency-first activation.

Failure destroys and joins prepared connections without publishing a partial skill. Once commit is queued, preparation waits for its definite outcome; cancelling an observer must not destroy clients that the actor just successfully committed. The server’s catalog never expands the manual’s allowlist. Ambiguous or colliding tool names fail rather than selecting whichever connection answered first.

A server-backed primary requires explicit asynchronous initialization after persistence loading and diagnostic setup. Concurrent callers join one shared initialization; cancelling one observer does not cancel startup for everyone. Until commit, the manager is unready and grants no provisional tools. User input and direct tool work are refused, while actor controls remain responsive. Startup is tracked control work, not a pretend model call with an invented ID.

Browser-backed startup has a necessary ordering wrinkle: the GUI must listen before the browser can connect. The primary dependency closure determines whether a browser wait is required. A stdio-only primary should not hang waiting for a tab it never needed.

Unload closes the selected skill’s remote clients, processes, and Tools immediately. Its manual can still remain pending conversational removal under Chapter 9’s policy. A retained explanation does not keep an executable connection alive, and independently owned dependency connections do not belong to the skill merely because it used them. Restart rebuilds these resources from current configuration; saved tool JSON cannot resurrect a process. Failed primary startup must not replace a previous good session save.

## A result is richer than stdout

Ordinary MCP calls remain Jobs. A foreground observation timeout means the work is still running; it does not cancel the remote request. Explicit Job cancellation sends cancellation across the boundary. Waiting for a result does not transfer ownership of its lifetime to the observer.

MCP returns ordered content blocks. Supported text, image, and audio become the existing neutral parts, preserving order and the error flag. Inline media requires valid base64 and MIME information. Unsupported resource blocks fail explicitly instead of disappearing from the answer or triggering an unrequested fetch.

The Job retains an atomic structured result separately from incremental process-output bytes. Diagnostic output is written before completion is published. The first completed ordinary observation consumes those structured parts; later observations do not replay them. Automatic collection has its own completion read so it does not steal the public observer’s cursor.

Decoding supported media does not establish that every provider accepts it in a tool-result position. Provider rendering remains another boundary with its own explicit refusals. A successful codec cannot grant an adapter a capability it does not have.

## Fresh context has a short life

Some tools supply changing context rather than model-chosen actions. Supported naming prefixes and discovery annotations identify per-turn or per-round collection. Per-turn values are reused only within that turn; per-round values are refreshed before each model request, including requests following tool results. Conflicting frequency declarations are rejected, and only the actual active owner’s grants make a definition eligible.

These are not callable tools in disguise. They stay out of provider declarations, `$TOOLS`, deferred catalogs, and discovery references. Guessed calls and reverse requests cannot invoke them. Otherwise an automatic context mechanism would quietly broaden the public action surface.

Collection creates real Jobs and waits for completed values. A timeout-limited running notice is not the snapshot the next request needs. Controls remain responsive while collection waits. Interrupt abandons that turn’s wait without killing an independent Job; its eventual result remains inspectable but cannot leak into a later turn. Shutdown has separate authority to terminate and join owned work. Pause still blocks HTTP after collection completes.

The completed values join the current manual ephemera slot just before request freezing. This slot is replaced and consumed, not appended to durable dialogue. Rendering or HTTP failure must not strand stale context for the next unrelated request. A failed snapshot contributes a short diagnostic naming the tool and real Job handle, not failed output promoted to observed facts.

## One browser, one claimed session

The browser’s existing GUI WebSocket also carries enveloped MCP traffic. Agent identity comes from server replay and must match before either controls or RPC dispatch. One socket writer preserves ordering. MCP shutdown can close its own tunnel without closing unrelated conversation display traffic.

Connecting registers a transport; it does not install tools. Loading the browser skill claims the eligible registration, completes discovery, and commits its explicit allowlist. That registration is one-shot. An observer tab can watch the conversation but cannot answer another tab’s calls. A delayed callback from an old connection stays tied to that connection instead of borrowing the newly connected socket.

The browser can report bounded visible DOM and usable selectors, click or fill a uniquely selected visible element, and expose or control its existing speech queue. Inspection precedes action, and a fresh inspection establishes its effect. A DOM snapshot is not pixel vision. Queueing speech is not evidence that anyone heard it. Omitting password values is useful, but the rest of the page may still be sensitive. A real external boundary improves reach precisely because its evidence and authority remain explicit.

Source: [Chapter 11](../chapter-11.md).
