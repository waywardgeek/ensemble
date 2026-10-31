# Chapter 18: Giving the Agent the Web, and a Status Line

An agent that cannot reach the web can explain what it learned during training, but it cannot verify what a page says now. An agent that cannot inspect its own usage has a different blind spot: it can perform work without knowing what that work consumed. This chapter adds both kinds of visibility. The hard part is not giving the model more text. It is making the origin, authority, and limits of that text clear.

A fetched page is external evidence, not a new instruction. A status report is a snapshot of measured work, not the model's estimate of its own effort. Both features depend on paths already established in Ensemble: skills grant capabilities, tools return correlated results, jobs retain work that outlasts a wait, and owned objects provide access to the state they govern.

## Web access belongs to the agent

Some model APIs offer native search. That is convenient until the agent changes providers, or uses a model whose API does not offer it. Search attached to a particular model is a capability the application cannot reliably carry across that change. A search tool dispatched by the agent has a different owner. The model chooses to call it, but the application controls its availability and execution.

MCP provides the bridge to an external search service. The agent initializes a connection, discovers the server's tools and their schemas, and converts those definitions into the format used for model requests. A call follows the same bridge back to the server, and its result returns to the conversation as the result of that call. The agent does not need a separate adapter for every search provider's API. It does need a trustworthy transport and an explicit policy about which discovered tools it will expose.

That separation mattered in Bill's recorded investigation. The open-source candidates included `ddgs`, a metasearch library with a built-in MCP server that required no credentials. Three manual tests looked promising; a larger set of fifteen succeeded only 73 percent of the time. The failures offered little explanation. Another candidate, `crawl4ai`, had no MCP entry point or module in the tested pip distribution. Its MCP support existed in the Docker deployment as an HTTP/SSE endpoint. The packaging mismatch mattered: installing the library did not provide the advertised server.

A clean probe of Firecrawl's hosted endpoint succeeded ten times out of ten, then the keyless endpoint failed often enough during sustained use that Bill moved to a keyed endpoint. The story is not a permanent ranking of services. It shows why discovery, installation, and sustained operation are different evidence. A short successful streak does not establish availability, and a failure that names its cause is more useful to an agent than an opaque refusal. Diagnostics determine whether the next attempt can change anything relevant.

The chapter's choice of a hosted service also changes the trust boundary. Queries and requested URLs leave the machine. The bridge sends selected tool arguments rather than automatically uploading the conversation, but the model can still put private conversation text into an argument. A smaller request surface is not a guarantee of harmless disclosure.

## A connection is more than a POST

The earlier MCP bridge could communicate with a subprocess over standard input and output. A URL transport preserves the same client-facing lifecycle—start, send, receive, and close—but replaces pipes with HTTP. That small interface keeps the client independent of the transport. It does not make the two transports behave identically underneath.

HTTP success alone does not identify a useful tool reply. A server can return an ordinary JSON response, or a server-sent-events stream containing JSON-RPC messages. A single SSE event can span several data lines. A notification can receive an empty acknowledgment, often an HTTP 202, without producing any message for the client. Treating every response body as one JSON document breaks those legitimate cases. The transport must recognize framing before the MCP client can interpret the message.

`Send` parses the POST response and queues its messages; `Recv` drains that queue. One POST can enqueue no replies or several. A fixed-capacity queue can block `Send` if the consumer cannot drain it until the POST finishes. The implementation uses an unbounded queue with a condition variable to avoid that deadlock; the required behavior does not mandate that particular data structure. This POST-based path does not implement every optional transport capability, such as GET streams and resumability, or claim complete conformance to the current standard.

Initialization can also establish a session. If the server returns an `Mcp-Session-Id`, later requests must carry it. The request's Accept header needs to admit both JSON and event streams. These details explain a common half-success: initialization appears to work, but the next operation fails because the transport discarded the identity or representation negotiated at the beginning.

Credentials belong on this transport side of the boundary. The skill's `auth-env` field names an environment variable whose value becomes a Bearer token in the Authorization header when the connection is made. The skill file carries the variable's name, not the secret. If the variable is unset, the transport logs the degradation and attempts an anonymous connection; the server decides whether to accept it. A silent downgrade could hide misconfiguration behind a low rate limit. The configured token must also be redacted wherever a peer reflects it, including error diagnostics and successful RPC content. Successful retrieval is not permission to copy a secret into history.

Cancellation exposes another difference from pipes. Stopping an HTTP request does not necessarily wake a goroutine blocked in a reply-queue wait. In the chapter's acceptance review, close initially appeared correct because the MCP client stopped waiting. It had a separate quit signal, while the underlying transport receiver could remain blocked. Testing the receiver directly exposed the unfinished shutdown. Close must end both the network operation and the receiver's wait.

This lifetime is separate from the agent's callback wait. A fetch that exceeds a short wait becomes a running job whose result can be collected later. It is not cancelled merely because the immediate caller stops waiting. Cancellation, timeout, unload, or session teardown can end the work; an ordinary running notice cannot stand in for its result.

## Grant the capability, not the page's wishes

The web skill describes the search server and grants search and scrape operations. Loading it makes those tools available; unloading it removes them. Discovery is broader than authorization: the server may advertise many tools, but only the granted names may be dispatched.

Model-facing disclosure and dispatch enforcement serve different purposes. Hiding a definition discourages a call; refusing an ungranted dispatch prevents one. A model that guesses a tool name must not bypass the skill boundary. Conversely, a skill file sitting on disk is not necessarily loadable. The primary skill must allow it. Progressive disclosure narrows the tools available during unrelated work, but it is not human approval: a model allowed to load a skill can be persuaded to do so.

Catalogue presence is not itself a grant. Anthropic can receive deferred definitions in the request's tool catalogue; the paired `load_skill` result's inline `tool_reference` discloses a definition to the model. That reference is content, not a cache marker. In the October 22, 2026 live run, the provider rejected a request with HTTP 400 because a definition combined `defer_loading: true` with `cache_control`. The correction puts the tool-cache breakpoint on the last non-deferred definition, or omits it when none qualifies. It leaves the catalogue, its order and defer flags, paired references, and system/history cache markers intact. Changing cache eligibility must not change callable authorization.

The first recorded live probe found exactly this distinction between installation and availability. The default primary did not list `web-search` as loadable. A test primary did, so the isolated tests passed while the installed agent could not use the feature. The mechanism was enforcing the supplied policy correctly; the delivered policy was incomplete.

Once the page arrives, it must remain a tool result tied to the initiating call. It does not become a user message or a system instruction. Keeping that envelope preserves provenance and gives the model a basis for treating the bytes as retrieved evidence. It cannot guarantee that the model will respect the distinction.

A page can contain text telling the model to ignore prior instructions, reveal context, or run a command. The dispatcher can refuse a tool that was never granted. It cannot prevent every misuse of a granted one. An attacker does not need a file-writing tool if a search query can carry private text to an external service. Source labels and warnings help the model interpret content; they do not turn the model into an enforceable security boundary.

Bill's reported practice is therefore sparse, supervised searching. Human observation provides a chance to intervene, not a guarantee that an attack will be recognized. Keeping credentials out of model context, limiting enabled tools, and minimizing disclosed query content each reduce a different exposure. None makes arbitrary retrieved text trustworthy.

## What actually reached the model?

Search identifies candidates. Fetching lets the agent inspect a chosen source. Neither operation succeeds merely because some recognizable text appears somewhere in the conversation. The requested URL or search phrase may already contain that text. The important path is from the server's response, through the correlated tool result, into the next model request. Multiple returned content blocks must survive that path rather than silently collapsing to the first one.

Error handling needs the same attention to origin. An HTTP failure connecting to the MCP service is different from an MCP tool reporting failure inside an HTTP 200 response. The former can be an authorization or quota refusal from a reachable server, not just a dead host. The latter carries the tool's `isError` flag and diagnostic, which must not be reformatted as a successful answer.

A third case lies beyond both: a successful scrape operation can return a failed web page. In the original probe, Firecrawl returned a page's HTTP 404 content with `isError: false`; the page status appeared in metadata. Checking only the tool-error flag would let the agent confidently summarize “Page Not Found” as documentation. Transport status, tool status, and fetched-page status describe different operations.

The second-edition live run exposed a further break after a real scrape had completed. The primary could load the web skill, but lacked `wait_for_job`. The model received a running-job notice while the finished page waited in the job store. The process exited normally without answering the question. More fetching machinery would not have helped. The missing capability was the existing means of collecting the result.

After that correction, one search, one scrape, and one wait produced an answer supported by the official Go download page. The fetched page was about 1.8 megabytes; the model received a selected 12,000-byte portion with truncation disclosed. The relevant filename and HTTP-200 metadata were in that portion. A fetched page, a saved job result, and the bytes shown to the model are three different artifacts. Saving the complete page does not give the model knowledge of omitted regions.

A later recorded CLI run, on October 22, 2026, completed the web task while switching between two same-provider models and checking usage. It used an Anthropic API key, not OAuth, and did not exercise the optional GUI. These are bounded observations about the recorded runs. Their release filenames are not current release advice, and their successful requests do not establish today's service availability.

## Five numbers, and the missing path to them

The status tool asks for modest information: the selected model, counts from the last response, session totals, cache hit rate, and estimated dollar cost. The first-edition investigation found the arithmetic already implemented. The tool simply could not reach most of its inputs.

Bill had expected each child object to retain a path to its actual parent. That path should have made root-owned state reachable without separately passing each fact to each consumer. Instead, the historical binary constructed objects side by side, then supplied individual dependencies and closures whenever one needed another's state. A clean composition root existed, but nothing called it. Reading that unused constructor would have produced a reassuring account of an architecture the executable did not use.

The investigating agent initially concluded that the shared package lacked an Agent interface. Bill challenged the diagnosis. The interface existed under the name `Host`, but exposed only logging methods. Its name and narrow original purpose made it difficult to recognize as the place to extend access to the owner. The cited declarations were real; the explanation built around them was wrong.

That misunderstanding had consequences. Each new feature added another side path instead of extending the ownership relationship. A model getter supplied by closure could serve one consumer, but it did not restore a general route through the owning objects. The status tool finally made this accumulation visible because it had no convenient object on which to attach another closure.

The old architectural check had not protected against it. It searched for the words `Logf` and `Host`, not whether a dispatched tool could reach its actual owner. The words remained while the property disappeared. The same distinction applies to a constructor: its presence proves little unless the delivered executable uses it.

Those are historical defects, not instructions to recreate the old tree. In the accepted implementation, the tool-call context reaches the owning Agent's Engine; that Engine exposes its Agent, and the Agent exposes Ensemble. The status handler obtains the Engine's snapshot; it does not create a second counter or a separate status service. Shared interfaces keep tool code from importing the concrete Engine and actor implementations. The headless and GUI entry points use the public construction path.

An interface field alone does not establish that relationship. A credential provider supplies a service; it is not automatically the parent of the configuration that uses it. The useful question is who owns the state and whether this particular call reaches that owner.

## Counts retain their history

The historical cost display applied the currently selected model's price sheet to all session tokens. Switching to a cheaper model therefore made earlier work appear cheaper too. The arithmetic was correct for a single model; its inputs had lost the distinction needed for a mixed-model session.

Storing a dollar total would hide a different problem: prices can change while measured token counts remain history. The accepted meter keeps counts separated by vendor, request model, and price tier, then derives cost from those buckets. A change in the selector does not move old usage into the new model's bucket.

If any contributing bucket lacks a price, the full dollar estimate is unknown. A sum that silently excludes unpriced work looks more precise than the evidence permits. Returning to a priced model does not erase the unpriced work already recorded. Counts remain reportable even when cost is not. And a model-token estimate is neither a subscription invoice nor the bill for a separate search service.

Cache counts need equally explicit scope. Cache reads are measured provider usage, not a reward inferred from reusing a prompt. The hit rate divides cache-read tokens by total input: fresh input, cache writes, and cache reads. Cache writes also contribute to cost. Omitting them from the denominator inflates the apparent reuse rate.

The last-response and session views answer different questions. The last-response fields describe the latest completed main request; the accumulated view describes recorded work across the session. The status call itself occurs during inference. It sees the counts recorded when the tool executes, before the model has produced its final answer about those counts. A later GUI meter includes more work. Comparing those two observations as though they were simultaneous creates a false discrepancy.

The recorded browser exercise illustrates this without needing a guessed dollar value. Two completed turns used different selected models, and the actual provider requests carried those names. Earlier usage survived the switch. The final visible meter then advanced beyond the second status result because the answer added tokens. Both tool results reported unknown cost because the model rows had no prices. That run established routing and measured counts, not priced accounting; separate controlled checks covered known tariffs.

## Selection and lifetime remain owned operations

A real model switch is not the model saying it has changed identity. The CLI's `/model` command and the GUI's model field invoke the Agent's operation at a turn boundary. It rejects an empty name or a busy turn, changes the selection, and clears the old resolved identity. Conversation and usage survive, as do provider, endpoint, credentials, and billing configuration. This is a same-provider model change, not an arbitrary provider migration.

The effective selection must govern the next provider request as well as the displayed label. It is session state rather than a saved preference: restarting uses launch configuration. A mixed request to change both model and preferences is refused before either is applied, avoiding an ambiguous partial update.

Ownership also has a concurrency dimension. During the historical status investigation, disconnecting a GUI client could race with a notification sender. The sender took a snapshot of clients under a lock, released the lock, and then sent. Teardown could remove a client and close its channel between those steps. A nonblocking send did not protect it: a default branch handles a full channel, not a closed one. Sending to the closed channel still panicked.

The remedy was a coordinated lifetime, not another getter or a catch around the send. The accepted GUI keeps the data queue open, coordinates client iteration and removal under the server mutex, and uses a separate completion signal to stop and join the writer. A channel need not be closed to be garbage-collected. The historical tests became informative only when they arranged overlap between sending and teardown and could expose the restored defect; many repetitions without that overlap had passed falsely.

An owner chain answers where state comes from. A lifetime protocol answers whether it is still safe to use. Neither substitutes for the other. With both intact, web tools can return evidence without acquiring authority over the agent, and status can report measured work without inventing a second history of it.
