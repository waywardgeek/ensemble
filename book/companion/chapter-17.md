# Chapter 17: Giving the Agent the Web, and a Status Line

Bill wanted search he could carry between models. Community tools looked attractive: open source, no keys, no lock-in. Three manual calls to one candidate looked promising. Fifteen distinct queries succeeded only eleven times, and the failures offered almost no explanation. Another project's advertised MCP support existed in its Docker deployment, not the installed Python package.

Those historical measurements explain the choice of Firecrawl, not a permanent service ranking. The architectural decision is more durable: web access belongs to the Agent's tool boundary, not to one model vendor's search format. Status supplies the complementary view. A page tells the Agent what the outside world returned; Engine's meter tells it what its own completed requests consumed.

## A new transport, not a new tool system

HTTP MCP implements the existing external transport interface. The MCP client still initializes the connection, routes replies by ID, discovers tools, and executes them through Jobs. HTTP adds independently cancellable POSTs, not another scheduler.

A request with an ID needs a matching JSON-RPC reply. An empty accepted acknowledgement can finish a notification, not a request. JSON and SSE replies use the same routing. An SSE request can finish when its matching reply arrives, without waiting for a server that leaves the connection open. EOF without that reply is failure, not an empty success.

Initialization publishes the server's session header before exposing its result, because the next notification must already carry that header. Disposal cancels and joins outstanding HTTP work. Redirects and automatic retries are deliberately absent: a redirect could forward the bearer elsewhere, while a retry might repeat a remote effect. This bounded transport is not a resumable event-stream framework.

The host supplies named authentication values separately from manual variables. The library does not rummage through process environment, and secrets do not become instructions or saved History. Missing authentication is reported before an anonymous attempt rather than disguised as an authenticated low-quota connection.

An early passing local run missed one diagnostic path. Non-success HTTP bodies removed the configured bearer, but an HTTP-200 JSON-RPC error could echo it. Source review found the gap; real-HTTP JSON and SSE cases reproduced it. Redacting the decoded diagnostic field fixed the leak even when JSON escaping changed its spelling. Ordinary tool content was not globally rewritten to conceal the issue.

## Discovery does not grant authority

The loadable web skill grants exactly search and scrape. A credential can reveal a much larger inventory, including expensive crawls, account modification, or billing queries. Discovering those tools must not authorize them. Ungranted names never enter the model's enabled definitions, and guessed calls refuse as unknown.

Loading publishes instructions and authority only after connection preparation succeeds. Unloading revokes this skill's grants and destroys its connection, without revoking another active skill's independent grant. As chapter 14 established, the pending manual can survive after its authority has gone.

The MCP client belongs to both the Agent's MCP manager and the activating Skill. Construction and destruction maintain both relationships, so unloading does not leave a detached catalog alive. Common declares these relationships; functional partial sections implement them directly in one C# assembly. The transport interface is a real external message boundary, not a reason to turn every owner relationship into a service interface.

The shipped primary matters. A historical test supplied a primary that allowed web access while the installed one did not. Progressive disclosure correctly refused the real user's load request. Another run fetched a page successfully but lacked `wait_for_job`; the model saw a running notice and could not collect the result. The fix was to expose the existing observation tool, not increase a timeout or invent another fetcher.

That corrected run fetched roughly 1.8 megabytes and showed the model a selected 12,000-byte portion with truncation disclosed. The fetched page, saved result, and delivered excerpt were different artifacts. The answer depended on the excerpt actually containing the relevant evidence.

## A page cannot promote itself to SYSTEM

Retrieved text stays in the correlated tool result, regardless of what it asks. A page might claim to be an agent configuration file, demand a diagnostic shell command, or request private context in the next search. None of those statements has authority to change settings, load a skill, or run a command.

The two-tool grant reduces the available damage, but it cannot prevent exfiltration through an allowed search query. Bill's preference for deliberate, supervised searches addresses a risk the capability list cannot eliminate. Search and scrape send the necessary query or URL, not the conversation merely because a page requests it.

Failures also have layers. HTTP can succeed while the MCP tool reports an error. A successful scrape can retrieve an HTTP-404 page with its tool error flag unset. Page metadata, tool status, and transport status refer to different operations. Treating any one as the verdict for all three turns “Page Not Found” into confident documentation.

A web result is retrieved evidence, not a new instruction source. Status is a current meter snapshot, not another conversation owner. Both use existing tool traffic and chapter 14's working-context policy rather than inventing a permanent parallel history of pages or status banners. Useful conclusions still need their source and scope.

## Five useful numbers, one actual owner

The status tool returns the selected model, latest-response counts, session counts, estimated cost, and cache hit rate. It performs no model request of its own. From the executing Job it follows Jobs to Agent, then reads one coherent Engine meter snapshot. The handler formats existing data; it owns no counter, pricing table, or saved-log scan.

The first-edition status task exposed a broken ownership path. The clean composition root existed but had no callers; the binary built a flat substitute and stapled individual getters onto objects. A check searching for two architectural words kept passing while the actual relationship disappeared. That is a historical Go failure, not a pattern to port into C# with callback bags.

The same investigation found mixed-model costs computed using the currently selected model's price for all earlier tokens. Keeping counts rather than dollars was sensible, but incomplete: counts needed their response identity too. The current meter retains actual model, tier, and funding. Selecting a cheaper model cannot retroactively discount yesterday's requests. Any unpriced contribution keeps the session total unknown, even after returning to a priced model.

“Last response” means the completed response that requested the status tool. The answer consuming that tool result has not happened yet. A later GUI meter can legitimately be larger. Unknown counts and prices remain unknown; the cache rate uses cache reads divided by all input, including writes. A token-price estimate is not a web-service bill or a provider invoice.

Human model selection enters through the actor mailbox while idle. Busy or queued work refuses, as do active direct Engine sends—the public Engine API makes guarding only actor turns insufficient. Validation builds a new immutable Configuration before publication. The switch preserves conversation, Engine, meter, lens, provider, surface, account, and funding; it resets model-specific resolved identity and defaults. Restart still uses launch settings.

CLI and GUI project acknowledged selection rather than assuming a proposed name took effect. Capability rows govern local compatibility; account discovery governs what the account can see. Neither establishes price or permission to spend.

The source's current results cover real hosts and local HTTP peers, including corrected diagnostic redaction and model switching. Historical live search and browser sessions supply separate observations, not a fresh C# hosted-service trial. The practical gain is immediate: the model can obtain current evidence and inspect measured work without either operation inventing a new authority inside the conversation.
