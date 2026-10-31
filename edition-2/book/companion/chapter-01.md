# Chapter 1: One Loop, Sixty Billion Dollars

## The logger that arrived late

The first version could talk to a model before it could answer a smaller question: where does a low-level helper get the application's logger?

In the [instructional chapter's account](../edition-3/chapter-01.md), the program had grown to thousands of lines before a refactor confronted that omission. A global variable would make the logger reachable. Passing it through every intervening function would also work. Neither answered the underlying question particularly well: which application does this helper belong to, and how does it reach the services that application owns?

The later design starts with that answer. Ensemble owns its Agents and the application logger. Each Agent owns a conversation and its configuration. Its Engine owns the HTTP exchange and measured usage. The command-line program constructs Ensemble through the public library, asks it for an Agent, and uses that Agent even when there is only one conversation to manage.

That can look like ceremony around a POST request. The point is not to make the first request easier. It is to let the second conversation use the same path without inheriting the first conversation's history, and to let a helper find the right logger without acquiring a second source of truth. The architecture earns its place through those ordinary changes, not through a forecast of every feature the agent might eventually have.

The initial behavior is deliberately modest. A user supplies text. The program sends that text with the conversation so far, receives text back, retains the completed exchange, and repeats. There are no tools, streaming responses, or concurrent turns yet. This is the part of a coding agent that can fit comfortably in someone's head. It deserves to remain understandable after the rest arrives.

## Why own the request?

A framework is often the economical way to build an agent. Storage, retries, model discovery, and tool plumbing are useful work that somebody else has already done. Bill's objection concerns a narrower problem: a coding agent whose behavior depends on exactly how context reaches the model.

Storage and delivery are different responsibilities. A framework may retain every message correctly while offering too little control over their arrangement in the next request. For ordinary chat, that distinction can be invisible. For steering and caching, the arrangement is part of the behavior.

The source describes Bill's discovery that a hint could accompany a tool result. In the Anthropic Messages API shape used by that design, the result comes first in the user message; a text block carrying the human's hint follows it. The model then has both the result and the new direction when work resumes. If a framework consumes the tool result and immediately constructs the next request, with no insertion point between those actions, the operator cannot use that moment to steer. A feature that sounds like a sophisticated interaction can depend on access to one position in one payload.

Bill's account also reports that the hint did not always take effect on the request that carried it. That is a useful limit on the story: legal bytes and prompt delivery do not establish timely model behavior. The client controls what it sends. Whether a particular API and model act on that hint at the intended point requires observation. The historical account is not a promise about every model or the service as it exists today.

Placement matters for a less conversational reason too. Prefix caching reuses an unchanged beginning of a request. A changing timestamp near the front ends the common prefix there, even if the large body following it is unchanged. Bill reports that moving a timestamp transformed CodeRhapsody's cache behavior. The mechanism is more useful than a percentage: volatile material belongs after the stable context it should not invalidate. A current screen state or status snapshot also needs only its current copy, rather than a new retained entry on every turn. Cache breakpoints make the same ownership question explicit: who knows which context is stable, and who gets to mark the boundary?

These are reasons to own request construction, not proof that every SDK prevents it. An abstraction that exposes the necessary ordering and cache controls may be adequate. An abstraction that hides them transfers control over behavior and cost to its author. Ensemble's instructional design chooses a plain HTTP client so those decisions stay visible. The first loop needs none of the advanced capabilities yet; it establishes the place where they will belong.

## Ownership is not the import graph

The runtime has a downward ownership chain: Ensemble, Agent, Engine. Access to an owner's services travels back up that chain. Engine receives and retains its Agent parent through an interface; Agent does the same for Ensemble. Those required parents cannot be nil. An Engine needing to log a diagnostic follows its Agent to Ensemble's logger.

The Go packages do not import each other in the same shape. Shared values and interfaces live in `internal/common`. Agent and Engine implementations import that vocabulary rather than importing one another. The public composition root imports the implementations and connects them. This separates two questions that are easy to blur: which code must know a type, and which live object owns another object?

`common` is consequently neither the runtime Ensemble nor a drawer for functions that have become inconvenient to place. HTTP behavior belongs with Engine. Go cannot attach an Engine-package method to a type declared in `common`; a free function can keep that behavior in its responsible package without moving all the surrounding machinery into the shared package. Small private runtime structs can stay with their implementations. Methods that participate in standard-library dispatch, such as JSON encoding methods, can remain beside the shared types they serve.

Bill's missing logger gives these interfaces a concrete purpose. They still create coupling: they make a child's dependence on its owner explicit rather than eliminating it. Their advantage here is a stable route to application services without copying those services and their state into every descendant. Configuration remains with Agent. Usage remains with Engine. Copying either into a sibling invites two answers to a question that should have one.

A parent relationship is also more specific than a dependency. An HTTP client used by Engine is not therefore Engine's parent. Nor does the design need to pretend every object has exactly one owner: the source uses a pin that belongs to an IC package and may also be connected to a wire. Actual relationships matter, including optional ones. Adding a child must establish both the parent's record of the child and the child's back-pointer; Ensemble, as the root, needs no invented parent above it. Helpers that need diagnostics must have a route to the relevant owner too. A bag of callbacks can provide services while concealing these relationships rather than clarifying them.

This is enough structure for the small program. Later, observations for display can reach Ensemble through an Observer, while explicit parent methods provide services. The intended GUI and its WebSocket transport live in an optional separate Go module using the public core API. CLI and GUI use the same Agent. Those are boundaries to preserve, not a reason to build a display system before the conversation works.

## A turn crosses two boundaries

An ordinary turn starts at Agent. It rejects an empty or whitespace-only question and stages the new user message alongside its retained history. Engine receives that pending conversation, sends it, and returns either an answer or an error. Only a successful exchange lets Agent retain the new user message and assistant reply together.

The ordering does real work. If Agent appends the user's message permanently before the request succeeds, a network failure leaves an unanswered user turn in history. The next attempt can then produce two user messages in a row, violating this loop's conversation model. Staging the question and committing the pair after success avoids that repair problem. Engine does not need to edit history, and a failed exchange does not become conversational memory.

The calls are synchronous at this stage: one turn finishes before another begins. That restriction makes the commit rule simple. Concurrent calls would need a decision about which history each request sees and which reply may extend it; adding a mutex without deciding the semantics would merely serialize an accident.

The other boundary is the wire. This Messages API exchange does not maintain the conversation for the client. Each request carries the history needed to answer it. The retained slice in Agent is functional state: it determines what the model will see. If a reply is displayed but never appended, a human sees a conversation while the next model request sees something else. Remembering user messages alone is insufficient: facts introduced by the assistant must travel forward too.

This property of the request interface says nothing about a provider's logging or data-retention policies. Stateless conversation handling means the client cannot rely on a previous call supplying context to the next one.

For this initial loop, the retained representation is a role and a text string per message. Requests begin with a user message, alternate user and assistant, end with the pending user question, and contain no empty messages. Those are the program's deliberately narrow invariants. They are not a complete description of everything the API permits; same-role inputs and assistant prefills belong to a wider surface than this program needs.

The request goes to `/v1/messages` with the API key, the selected API version, and the JSON content type in headers. The body names the model, supplies a required output-token limit, places a fixed system string outside the messages array, and includes the pending conversation. This first transport uses non-streaming responses. Engine marshals the request, performs the POST, requires the expected successful HTTP status, reads the body, and decodes it. A transport or decoding failure remains an error rather than becoming an assistant answer.

An explicit model setting protects the request from stale recollection. The source records a build in which an agent selected a familiar model identifier from its training data. The request worked, but the choice was stale. Success at the HTTP boundary did not validate the intended model choice. Discovery through `/v1/models` supplies available identifiers; it does not make any identifier printed in a book permanently current. The initial program requires an explicit model rather than silently substituting one.

## The reply is not a string

The request may use a bare string for message content. The response uses a list of typed blocks. The shared field name is a small act of misdirection.

Engine walks the response blocks in order and concatenates the text from text blocks. Reading only the first block can pass casual testing because a single text block is a perfectly ordinary answer. It fails as soon as an answer is split across two. Filtering by block type prepares the parser for non-text content without pretending that every possible future block already has a meaning in this loop.

The source's controlled server splits a reply across multiple text blocks to expose the first-block shortcut. That establishes whether the parser collects the whole answer. It does not establish that the parser handles non-text blocks correctly when the test never supplies any. A prudent implementation and evidence for that implementation are separate things.

The response also carries a stop reason and measured token usage. The initial path handles a completed text answer, represented here by `end_turn`; later tool use will change both what content can contain and what ending a response means. Treating an HTTP success as the whole semantic result would erase that distinction before the agent needs it.

No usable text is an error for this text-only conversation. Missing usage is also an error, but for a different reason: absent accounting is not a measurement of zero. Engine retains measured usage even when a response contains no usable answer. Agent can reject the exchange as conversation while Engine records that the attempt consumed tokens. History describes successful dialogue; accounting describes measured expenditure. Making them share one success flag would lose one of those truths.

Diagnostics travel through the owner chain. They omit credentials and remote error bodies, which can contain material that does not belong in local logs. An error needs enough context to identify the failed operation without turning the logger into a second transcript or a credential archive.

## A long request is not necessarily a stuck request

Bill's account reports that a blanket HTTP client deadline killed valid model requests during later development. Generation can take a long time, including before response headers arrive. A total-response timeout chosen for an ordinary web request can therefore become an accidental limit on the model's work.

This design retains connection and TLS protections but lets the owning context cancel the model request rather than imposing an arbitrary total-response deadline. The distinction does not ban deadlines. A caller with a real latency budget can express it through that context. It prevents a transport default from silently becoming the product's policy about how long thought is allowed to take.

Waiting for progress, deciding that a user should hear an update, and cancelling the request are different actions. Later runtime features will need to preserve that distinction. The first synchronous loop already benefits from getting the lifetime right.

It also makes exactly one API call for each valid question. There is no warm-up call, hidden summary, retry, or backoff. A rate-limit response is reported as failure rather than concealed behind another attempt. That narrow starting design leaves retry policy for a more capable client. A retry introduces policy about cost, delay, cancellation, and what may already have happened remotely. Leaving it out keeps the first loop's behavior inspectable.

## The history has a running cost

Every successful turn extends the history sent on the next request. With a fixed system prompt and no pruning, the amount of retained conversation in the request grows even when the next question is short. For similarly sized turns, the cumulative amount of conversation sent grows roughly quadratically with the number of turns: early messages are paid for repeatedly.

Actual billing depends on the provider's tokenization, caching, and pricing. Request growth alone does not establish a dollar figure, nor does a fake server's character counter establish real token counts. The mechanism is nevertheless enough to explain why accounting belongs in the first implementation rather than in a future billing feature.

Engine sums the input and output usage reported by each response. Totals come from measured values rather than estimates derived from text length; missing data cannot be made reassuring by printing zero. Input and output remain separate because they represent different work and may have different prices. Even these totals are not omniscient: if an attempt fails without returning usage, the client has no measurement for it. A local counter is not a reconciliation with the provider's invoice.

Raw API access also comes with provider throughput limits and a pricing model that may differ from a packaged coding product. Owning the bytes buys control, not automatically cheaper operation. The design keeps the base URL configurable, alongside the key and model, so the same transport can address a compatible test server, a proxy, or the provider endpoint. Configuration changes the destination; it does not prove that every destination has identical behavior or trust properties.

## Two front ends, one conversation

The CLI has two ways to use the same Agent. A human-facing chat mode supplies prompts and readable replies. A machine-facing mode accepts one JSON object per physical input line and emits one JSON answer per successful round, followed by cumulative usage at normal end-of-input. Neither needs its own conversation implementation.

For the machine interface, stdout belongs exclusively to the protocol. Logs and diagnostics go to stderr. A friendly greeting or a debug print on stdout is not harmless decoration; it is an extra message in somebody else's language. The initial CLI consequently makes machine mode the default and chat an explicit choice. That convention serves its callers; the more durable boundary is that presentation never leaks into the data channel.

Input handling has two unglamorous consequences worth retaining. Go's default `bufio.Scanner` imposes a token-size limit that can reject a valid long JSON question. A line reader avoids that accidental restriction. It can also return data together with EOF: a final line without a newline is still a line to process, exactly once. EOF with no remaining data ends the conversation. Other read errors end in failure, with a diagnostic and a nonzero exit, not the terminal usage record that denotes normal completion. Otherwise a damaged input stream can look like a successful session.

The source's live-chat account includes a particularly economical false success. A launcher consumed stdin; chat immediately encountered EOF and exited cleanly without making a request. The program had demonstrated that it could stop. It had not demonstrated that it could converse.

A real chat probes the whole route through the front end, Agent, Engine, and service. Controlled requests answer different questions: whether the exact prior messages survive unchanged, whether an assistant-introduced fact remains in later history, whether every returned text block reaches the caller, and whether usage is summed rather than invented. Neither a cheerful terminal exchange nor a green protocol check substitutes for the other.

The model's own account of its environment is another limited observation. In the source's preserved conversation, it could describe the raw-HTTP setup supplied in its system prompt but could not confirm how many times the program had sent the conversation. That restraint was appropriate. The prompt told it what kind of assistant it was; the program had not exposed its request counter. Inspecting outgoing requests establishes what the client sends. Asking the model what the client sends establishes what the model can infer from what it has been told.

The result is still a small chat client. Its conversation representation will soon be too simple. The useful inheritance is the separation underneath it: Agent decides what becomes history, Engine owns transport and measured usage, Ensemble owns application services, and front ends present the same conversation through different interfaces. Chapter 2 can replace the simple record without rediscovering who is responsible for it.
