# Chapter 6: Streaming, or the Same Answer in Pieces

An HTTP response is already a stream of bytes. Waiting for its entire body before showing anything is a client decision. Removing that wait improves the human’s ability to follow the Agent, but it also creates two kinds of truth: what has arrived so far, and what the provider has actually completed. The display can live with the first. Conversation history and tool execution need the second.

## One answer, two delivery paths

The Engine accepts the open HTTP response after headers arrive. For plain JSON, it reads the body and invokes the established completed-response decoder. For server-sent events, it recognizes frames, emits display chunks, and accumulates the provider’s native response. Only after successful completion does it send that accumulated response through the same decoder.

That shared finish line matters more than the callback. Separate JSON and streaming decoders would gradually disagree about exact argument values, nullable usage, signatures, and opaque replay material. Streaming changes when content becomes visible without creating a second definition of a valid answer.

Plain delivery invokes the chunk callback zero times. There is no synthetic one-chunk “stream.” It still produces authoritative final parts, so a display must create a card from a final even if no delta preceded it. When deltas did arrive, the final replaces their provisional content rather than being appended to it. Otherwise the most successful response would display twice.

Common declares the value-like chunk and parsed-response vocabulary and the required partial parsing operation. Engine supplies the implementation in the same C# 14/.NET 10 assembly. The invocation’s callback carries transient parser output; it is not a new service or an alternate owner of conversation state. Accumulators are per-parse scratch state, released when decoding finishes.

## Progress travels through the owner

The HTTP worker posts chunks to the Agent’s existing mailbox, then posts an immutable completion. It cannot append History, dispatch tools, or finish the turn. Controls and streaming progress therefore meet at the same actor rather than competing through separate loops.

The actor publishes deltas to bounded, lossy observation queues. A slow terminal may miss a fragment; it cannot stall the Agent. Final events are different: the actor records them before publishing their finalized parts. Their notifications can still be lost; the stronger promise is the record, not guaranteed delivery. An observer that checks History upon receiving a final can already find the fact there. Deltas never become extra dialogue or additional token accounting.

A reliable request task establishes completion. Neither a quiet screen nor a final-looking fragment can do that, and losing an observation does not change the request’s outcome. This separation also explains the terminal arrangement: live text, thinking, and tool fragments are a stderr preview, while the reliable completed answer goes once to stdout. Redirecting stdout must yield the whole answer even when a display queue dropped progress. JSON observation output flushes each delivered record so streaming is visible through a pipe rather than merely happening inside the process.

## Identity survives rearrangement

A chunk needs to identify the part it extends. Allocating a new ID for every fragment would describe forty unrelated objects instead of one developing paragraph. Numbering the final array is also too late: a tool fragment may arrive before text that the canonical decoder places ahead of it.

The parser that chooses an ID therefore returns it again alongside the finished parts. IDs are response-local, nonnegative, and distinct across parts; zero is valid. The actor supplies the recorded request sequence as the response identity. Across several Agents, the display key combines Agent, response, and part. A provider’s tool-call ID remains a separate correlation identity, not a convenient substitute for a display key.

Each adapter translates its own wire semantics. Anthropic supplies block indices, including gaps. OpenAI tool indices need mapping to parser-local identities, while text and thinking acquire identities on first appearance. Gemini has no equivalent block index, so continuation depends on content kind, explicit boundaries, and signature closure. Empty signed text can matter to replay even though it paints no characters.

The final mapping preserves identity when canonical ordering differs from arrival order. Arrival order itself remains important for the explanation policy: substantive text that arrives after the first tool fragment cannot retroactively explain that call. The parser records that fact before final decoding rearranges anything.

## A visible fragment cannot authorize an effect

SSE framing is modest but unforgiving. Blank lines delimit frames; multiple data lines form one payload. Comments and CRLF must work, as must a pending frame at EOF. A frame is not a part: either can span the other. Protocol completion still has to be present. A connection ending after readable text is not sufficient evidence of success.

Vendor completion rules remain local to their adapters. Anthropic cannot succeed with an open block. OpenAI may send usage after the finish reason, so the parser accepts that accounting without accepting new choice content. Gemini’s signatures must survive on the kinds of parts they actually sign. None of these rules justifies dispatching a partly received tool name or argument object.

The entire response must finish successfully and the arguments must decode before tool admission is possible. A malformed frame, provider error, missing terminator, or cancellation can leave a persuasive preview on screen without yielding a successful recorded response. Expected protocol failures become credential-safe errors; cancellation returns through the actor. Completed calls refused by the explanation policy still receive paired results, but produce no effect.

Exact JSON values survive this path. In particular, a large integer must not quietly pass through floating-point storage and emerge rounded. Comparing decoded argument values is more meaningful than demanding that streamed whitespace match canonical JSON formatting.

## Three useful records

History preserves completed conversation facts. Observers offer timely, expendable progress. The optional API trace preserves sanitized request and received-frame payloads for diagnosis. These serve different readers and should not be mistaken for interchangeable logs.

The root serializes trace writes to a caller-owned writer, reached through the Engine’s actual Agent and Ensemble relationships. Redaction removes the configured key and its escaped representation, but does not make the trace public-safe: prompts, files, and signatures may remain. It is a payload trace, not a byte-for-byte HTTP capture or a general secret detector.

Streaming selection is equally restrained. A known model must allow streaming and the caller must not disable it. Unknown models fall back to plain delivery; losing immediacy preserves the answer. Unsupported media still fails before HTTP because silently losing an attachment would change the answer’s inputs. A capability-table entry can select a known parser. It cannot teach that parser a new protocol.

Source: [Chapter 6](../chapter-06.md).
