# Chapter 6: Streaming, or the Same Answer in Pieces

The Agent can listen while busy now. It can accept a hint, interrupt an observation without killing the job, and tell an application what happened. Ask it a difficult question, though, and the terminal still goes quiet until the entire model response arrives.

Chapter 5 prepared `PartDelta`; it did not make the provider stream. Its whole-part notifications were useful scaffolding, not evidence of words arriving early. Streaming earns that distinction by carrying pieces all the way from an open HTTP response to a display while more of the response remains unread.

The central bargain is small: change delivery without changing the answer’s durable meaning. A reply can arrive in one piece or many. History still records completed facts, tools still execute complete calls, and replay still reconstructs the same conversation. The human gets to see work sooner without making unfinished work authoritative.

## Let the Response Stay Open

An HTTP body is already a stream. Reading it all before parsing is a decision to hide that property. Streaming moves the parser onto the open response: the HTTP worker hands it the response, the parser recognizes arriving fragments, and those fragments travel through the Agent’s mailbox to observers. Worker messages carry the turn identity, so late progress from an interrupted request cannot decorate a newer turn.

There are three visible kinds of fragment: readable thinking, ordinary reply text, and tool-call content. The separation matters beyond colour. Reasoning must not leak into the reply, and a tool’s partly written arguments must not be mistaken for executable instructions. A provider that exposes no readable reasoning does not owe the display an invented thinking stream.

The parser accumulates the complete native response as it goes. On successful completion, the same semantic decoding used for plain delivery builds the finalized parts, with canonical arguments, provenance, usage, and opaque replay material intact. The worker returns facts to the actor; it never appends directly to History. The actor records and reduces the response before publishing final-part observations.

This preserves the existing ownership seam rather than creating a special streaming conversation loop. Rendering remains Engine’s work, configuration comes from its Agent parent, and transport constructs the URL and headers. Asking for a stream changes those existing request decisions. It does not require a second orchestration API.

The [source chapter](../edition-3/chapter-06.md) argues for one `Parse`, not `Parse` plus `ParseStream`. Two entry points look modest until each vendor has two definitions of a tool call, two places to interpret stop reasons, and two opportunities to lose usage. The path used in production gets repaired; the other becomes the mode selected when somebody wants fewer variables while debugging.

One method name alone cannot prevent that drift. Shared final construction does. The accepted implementation accumulates streamed provider messages before passing them through the semantic decoder used by plain responses. Framing differs; what counts as a completed answer does not.

Plain delivery also emits a delta: one whole-content notification per part, followed by its finalizer. Otherwise every observer would need both a delta renderer and a separate renderer for non-streaming responses. “A stream of length one” is useful because consumers can actually treat it that way.

Parity here means the same scripted content produces the same completed parts and usage through either delivery mode. Two independent live generations need not choose the same words. Turning streaming off is not a seed for model determinism.

## Framing Is Boring Until It Eats the Ending

The streaming dialects in this chapter use server-sent events, or SSE, over ordinary HTTP. A frame has field lines, including an optional event type and one or more data lines; a blank line ends the event. A shared reader turns this framing into an event type and payload before vendor interpretation begins.

A scanner that looks for `data: ` handles a pleasant demonstration. Real framing adds a few less pleasant cases. Multiple data lines belong to one event. A colon-prefixed comment may be a keepalive, not empty JSON. CRLF must still produce a separator. A completion sentinel such as `[DONE]` is not a JSON object. An event left unterminated at EOF must not simply disappear.

These faults often leave plausible output. Losing the final event may preserve all the visible prose while dropping its stop information. The resulting problem surfaces later, wearing the wrong subsystem’s name badge.

A small shared framing reader keeps those deterministic failures out of each vendor parser. Once framing is resolved, the provider-specific work can start with a real payload rather than another interpretation of line endings.

The native request still has to be valid. Anthropic and OpenAI request streaming in their bodies; Gemini uses its streaming endpoint with SSE selected. OpenAI additionally needs its usage-inclusion option. Without it, streamed usage is absent, and an accumulator that treats absence as measured zero can make every answer appear free while everything else works. Missing accounting remains missing accounting.

Streaming also does not rescue a rejected request parameter. The source’s accepted OpenAI renderer uses `max_completion_tokens`, while Anthropic and Gemini retain their own native token-limit fields. A parser cannot demonstrate liveness on an answer the provider refused to generate.

## A Part ID Means “This One Again”

Imagine a display receiving fragments for reasoning, reply text, and a tool call. It needs to append each fragment to the right place, then replace the provisional rendering with the completed part. The identity must survive both operations.

The first-edition brief initially proposed a fresh atomic ID for every chunk. The IDs would all have been unique. They would also have told a consumer nothing about which chunks belonged together. Forty numbered envelopes are not forty answers.

The next temptation is to use each part’s position in the finished response. OpenAI exposes the flaw: tool arguments can start arriving before the parser knows whether a text part will occupy the first final position. A caller choosing IDs from that future list is guessing about a response that has not finished existing.

The parser has the information, so it chooses the identities and repeats them at finalization. Anthropic can retain its content-block indices. The accepted OpenAI parser reserves zero for text and maps tool indices into a separate range. Gemini needs a local continuation mapping. Zero is a valid identity; positivity is not the guarantee.

Within one response, one part has one ID, different parts have different IDs, and an ID does not change content kinds. The finalizer carries the parser’s ID, not a position reconstructed from the final list. That is enough to connect an approximation to the thing that replaces it without claiming the number means “where.”

The scope matters. Parser IDs may restart with every response. A turn that calls a tool usually includes another model response afterward, and its part zero is a different part. A consumer needs the pair of response identity and part ID, not a turn-wide map of bare numbers.

The current protocol exposes the transition to `in_flight` before the response’s deltas, allowing a connected consumer to separate responses. That transition must remain visible even if the CLI also supplies explicit response identifiers. It is a live boundary, not by itself a durable identity that a reconnecting GUI can recover. Chapter 7 still has work to do there.

There is another order to preserve at the far end. Parsing reports the completed event before its part finals; the worker retains both the facts and the identity mapping. The actor then records the response and publishes `PartFinal` from the stored parts. An observer refreshing from History sees the fact it was just told about, rather than a log one event behind.

Text and readable thinking fragments should reconstruct their own parts, grouped by response and ID. Comparing only the concatenated text of an entire turn misses broken identities completely: the words may be right while every widget is wrong. Requiring a part to arrive in several pieces also distinguishes real incremental delivery from whole-part notifications with an ambitious name.

Tool displays have a slightly different boundary. They include arriving name fragments as well as arguments, emitting each fragment once rather than repeating the accumulated name on every update. A consumer appending the ordinary chunk field must see the tool name; hiding it in supplemental metadata does not accomplish that. Plain delivery supplies the complete name and arguments once.

The raw argument fragments may retain wire whitespace, while the authoritative tool part holds canonical JSON. At that boundary, decoded argument values are the meaningful comparison. Cosmetic differences in the display must not become differences in what the tool receives.

## Seeing a Call Is Not Running It

Early tool arguments are useful to a human. They show what the model is preparing to do. They are not permission to begin doing it.

A path may arrive before the rest of a write request. Opening the file at that point would let an incomplete JSON fragment cause an effect. Even a completed-looking part is not enough: dispatch waits until the response completes successfully. An error after visible fragments must not leave behind a tool operation started on speculation.

Streaming also makes the order of explanation observable. For models whose policy requires ordinary explanatory text before a tool batch, the implementation must retain whether nonblank text arrived before the first tool-call delta. Thinking does not count as that explanation. A tool-first stream followed by reassuring prose has still shown the call first.

This cannot be recovered solely from the finalized response. A decoder may arrange final text before tools even when the wire delivered them in the opposite order. Accepting that arrangement as evidence would let the model explain after announcing the action and receive credit for explaining before it.

The final-response check therefore stays, including for plain JSON, and streaming retains the earlier arrival evidence as well. A tool-first stream or reasoning-only prefix gets corrective tool results instead of execution when the explanation policy applies. A text-first stream may dispatch after successful completion. The streaming display makes more of the process visible; it does not relax the existing gate on effects.

## Local Differences, Not Imaginary Uniformity

A common observer vocabulary does not mean the providers send the same thing.

Anthropic identifies content blocks and distinguishes text, thinking, and input-JSON deltas. OpenAI has its own indexed tool fragments and accounting option. Gemini sends response-shaped frames with new content rather than the same named delta events. It has no equivalent block indices, so its parser must infer continuation from the shape and position of what arrived before.

For Gemini, a new frame is not automatically a new part. Conversely, two text entries explicitly present in one frame’s parts array are not automatically one paragraph. The first entry of a later frame may continue a compatible unsigned part. That inference must stop at a signature boundary.

A trailing signature can finish unsigned text, thought, or tool content. Once attached, it binds to that exact content. Appending later words would change what it signed; replacing it would discard the original binding. A signature-only frame is therefore meaningful even when the display has nothing new to print. Ordinary text can carry signatures too.

The accumulator must preserve those distinctions: standalone signature material stays standalone, signed text stays separate from later text, and signed thought does not absorb a separate opaque part. Canonical parts retain the native payload needed for exact-source replay. Treating everything with no visible words as empty is a tidy way to break the next request.

In the reference implementation and observations, Gemini’s function arguments arrived as complete objects even while its text streamed. The correct display response was one whole tool-object chunk, not manufactured subdivisions to make the stream look busier. The actor and terminal did not need a Gemini exception; the parser expressed what the wire actually supplied through the common vocabulary.

The source retains a historical capability bitmask distinguishing text, thinking, and tool-argument streaming. The accepted restart uses a simpler SSE/plain/unknown capability, with native parsers handling the kinds of chunks available. These are different representations of the same obligation, not two simultaneous requirements. Neither promises that every model exposes every kind of content incrementally.

Capability rows also age. A table update can enable a dialect the translator already understands. It cannot teach a parser a new wire format. The design contains provider differences inside translation; it does not establish that providers are interchangeable.

## An Unknown Model Can Still Answer

The capability table’s conservative fallback is plain delivery. An unlisted model is not a reason to reject an otherwise valid text request.

This deliberately differs from Chapter 5’s media refusal. Guessing about an attachment can remove content the model needs while leaving a plausible answer behind. Falling back from streaming preserves the content and gives up early visibility. One risks changing the conversation; the other changes how long the terminal stays blank.

Model names form an open set, including test models and deployments newer than the table. Requiring every name to be known before any request can proceed would make the capability catalog an accidental access-control list. Unknown streaming capability means do not ask for streaming, not do not work.

The explicit opt-out is `DisableStreaming`. The negative name buys the intended Go zero value: a hand-built configuration streams when the model’s capabilities permit it, unless the caller disables that behavior. A positive `Stream` boolean would default to false unless every caller used a constructor. In a public library, callers are allowed to construct structs. A default that works only through the preferred entrance is a suggestion.

## Three Accounts of the Same Response

History, the wire trace, and the display answer different questions.

History records finalized conversation facts. `ResponseEnded` still carries the complete response, including its parts, provenance, and usage. Deltas never become additional conversation events. Persisting them as content would either duplicate the final parts or require replay to assemble a second representation of the answer. The existing log remains enough to reconstruct the live context.

Chronology in that log need not be literal message order in a vendor request. A busy hint can arrive between an assistant’s tool calls and their results. OpenAI requires the paired tool results immediately after the assistant call message, so its renderer moves the complete result group ahead of the hint in the disposable wire projection. History keeps the hint where it arrived. Rendering may rearrange real, complete results to satisfy the API; it may not invent results or rewrite the historical facts.

The root-owned `api.log` answers a different question: what payloads passed through translation? The first-edition engine logged response JSON after reading the body. Streaming removed that convenient whole-body moment. Without replacement plumbing, the debugging trace would have lost its response half precisely when responses became harder to inspect.

Frame notifications restore that visibility as payloads arrive, in order, in both modes. They are not a byte-for-byte HTTP capture: SSE comments and framing are not the payload, and authentication headers do not belong in the trace. The accepted Engine redacts a reflected configured key before logging or forwarding a frame and excludes credential-bearing URLs. Logging follows the existing parent chain to Ensemble, which owns and serializes the writer; parser callbacks carry output rather than a collection of borrowed logging services.

Observers provide immediate, provisional visibility. A display can append chunks quickly and later render the authoritative part under the same identity. It need not pretend its temporary rendering was complete. Tool-result finalizers remain distinct from model-response finalizers, or a consumer could count job output as part of the provider’s stream.

These observations still do not replace reliable request completion. A completed assistant answer comes from the request’s private reply, not an acknowledgment that input was admitted and not a progress queue that may drop entries. A busy hint acknowledgment is not another completed answer. Streaming adds more things worth watching, not more authorities over whether the request finished.

## The Last Buffer Can Undo Everything

The human CLI sends reply text to stdout, with thinking and tool-call content on stderr. Terminal-only styling makes those channels readable without contaminating redirected output with escape codes. A pipeline receives the reply; a human sees the surrounding activity too.

That arrangement can be correct and still fail at the last inch. A buffered writer may accept every chunk immediately, then hold all the bytes until completion. The parser streams, the actor streams, the observer streams, and the human gets a blank terminal followed by the whole answer. The buffer has done an excellent job of removing the feature.

Flushing each chunk fixes that last boundary. The accepted CLI instead writes chunks directly to unbuffered stdout. Its observer queues display work without waiting for a slow terminal or dropping reply chunks. Once the Agent has joined producers, the display drains before the summary. Finalizers add separators, not a second copy of the reply.

That queue is a deliberate resource tradeoff. Optional progress may be dropped when a consumer falls behind, but a terminal promising the full reply must preserve those bytes without holding up the actor. Buffering moves the pressure into retained display work; it does not abolish pressure. Making the actor wait would recreate Chapter 5’s deafness through the output side.

The source records a particularly revealing CLI failure. The initial student implementation treated any chunk beginning `hint:` as a human-hint acknowledgment and routed it to stderr. A real parser-to-actor-to-CLI test supplied assistant text in two pieces, `hint:early` and `late`. Only `late` reached stdout. The assistant had accidentally used a reserved prefix that nobody had reserved.

The correction marks actual acknowledgments by origin, using `PartDelta.Hint`, rather than inferring origin from prose. The content kinds remain thinking, text, and tool call. A fourth colour would not fix a missing distinction between who said something and what characters they used.

The corrected test also supplies the chapter’s strongest liveness evidence. It withholds the rest of the HTTP response until stdout contains `hint:early`, then requires the complete reply `hint:earlylate`. The first piece must become externally visible while the response is still unfinished, or the test cannot advance. Restoring the prefix heuristic breaks that same test; removing the origin marker breaks acknowledgment routing.

Several reads after completion would prove much less. A network reader, a pipe, or a terminal can split already-finished output into many reads. End-to-end streaming is a relation between visibility and unfinished work, not a pleasing chunk count.

## What the Evidence Actually Shows

The source’s separate public-library application observes stream shape: delta counts by kind, finalized parts, time to first delta, and the completed answer. It uses the ordinary Agent and observer seam, rather than a private demonstration path. Reliable request completion determines when its in-memory measurements can be summarized.

Those measurements are useful for different reasons. Counts reveal whether streaming produced multiple fragments; reconstruction by kind checks that reasoning stayed separate from the reply. Grouping reveals identity errors. First-delta latency describes the initial wait. Final content and usage establish plain/streaming parity for a controlled response. None of these alone proves that a human CLI flushed before completion; that is what the gated test establishes.

A first-edition audit removed only the request’s streaming flag. Final text, parts, and event log stayed identical, while the incremental-delivery check failed because too few deltas arrived. That is the chapter’s thesis expressed as a fault: content correctness can survive the complete loss of early delivery.

The audit also found limits in its own checks. A per-chunk-ID mutation initially used a package-level counter and failed an unrelated prohibition on mutable globals as well as the intended identity check. It had to be rewritten to isolate the identity fault. Removing thinking deltas did not fail a reconstruction check that examined only text. A failed check’s name is not enough to establish why it failed, and one content kind’s coverage does not silently cover the others.

The retained live shape pair used `gpt-4o-mini` and happened to return the same numbered text in both modes. Streaming delivered its first delta sooner but took longer overall in that small pair. That is evidence of earlier visibility in those runs, not a throughput improvement or a general latency benchmark. The model still has to produce the answer.

The human demonstrations exposed other boundaries. An Anthropic launcher watched file modification times and labeled an input as a new turn. A log rewrite changed the timestamp while the prior turn was still active; the durable record showed that the input was actually a busy hint. A changing file is evidence of a write, not of request completion.

The Gemini demonstration required a different correction. Bill had required Gemini 3.0 Flash or newer and specifically requested 3.8 Flash for hints. The student selected 2.5 from discovery, and the handoff and first review missed the restriction. The run worked, but success on the wrong model did not establish the requested behavior.

One bounded run on the actual `gemini-3.8-flash` CLI supplied the missing evidence. A real command created a waiting marker, paused for a release marker, and then printed its result. While it was waiting, a busy hint changed a requested lighthouse paragraph from red to blue and required a recognizable answer prefix. The launcher observed the actual stderr acknowledgment before releasing the command. The next provider request contained both the successful tool result and the hint; the answer used the prefix and described a blue lighthouse. Tool and CLI exited successfully.

That record establishes delivered steering, not merely accepted input. It does not establish leisurely token-by-token typing: its three nonempty stdout reads spanned only about a millisecond. The gated CLI test remains the stronger evidence that bytes become visible before response completion. No production change or repeat of the other provider demonstrations was needed to repair the model-selection gap.

Historical grading failures also need their causes kept separate. One fixture selected an unknown model while demanding streaming, contradicting the intended plain fallback. A reference run that supposedly disabled streaming still streamed. Neither result can be explained honestly by saying simply that streaming passed or failed. The source preserves those records; they are not a claim that this prose revision reran them.

## Enough Visibility to Want a Better View

The terminal now exposes reasoning, reply text, and half-written arguments across two file descriptors. It appends rather than redrawing completed parts. That is useful development tooling, and not much of a layout engine.

The next interface can do better without changing who owns the conversation. It has kinds for routing, response-local part IDs for grouping, and finalizers that replace approximations with stored facts. It still needs durable widget identity across reconnection, connection lifetime, and a sensible arrangement of the content. Those are GUI problems, not reasons to pull display policy into the parser.
