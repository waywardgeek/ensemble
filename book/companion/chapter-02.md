# Chapter 2: One Log, Three Vendors

The first interface in CodeRhapsody looked like preparation for a second vendor. Its methods took Claude messages.

In the history preserved by the [source chapter](../chapter-02.md), Bill's first client supplied both the implementation and the vocabulary of the interface extracted from it. When Gemini arrived, the new client had to translate its conversation into Claude's terms before doing anything useful. Copying the existing client was easier. OpenAI acquired another copy, and fixes acquired several places to be forgotten.

The eventual replacement separated a common conversation model from vendor renderers and parsers. It reduced duplication, but the big-bang migration left capabilities behind in the old clients; the source records broken audio attachments in the replacement. The better boundary did not make an incomplete replacement complete.

Chapter 1 established a simple conversation with one owner. The next step preserves that behavior while removing the first provider's authority to define what a conversation is. There is still one agent, not a different agent for each API.

## The record, the working view, and the request

Three things commonly called “history” need different rules.

**History** records what happened. **Context** is the current conversational meaning derived from that history. A **request** presents that context in the form a particular provider can accept. History must preserve an original tool result; context may replace it with a redaction stub; a request may move that stub next to its call. None of those operations licenses rewriting the other two.

Agent owns History and Engine. History appends events and reduces them into context. Engine renders requests, handles the exchange, and parses responses into neutral events, which Agent records. A parser does not edit context behind the reducer's back, and Engine does not keep a competing transcript.

An ordinary turn starts with a recorded human message. Reduction adds its content to dialogue and makes input pending. Engine builds a request from that context and configuration. Recording the send marks the turn in flight; the already-built request goes out. A parsed response returns through the event log and the same reducer. A reply without calls finishes the turn. A reply containing tool calls leaves tools pending; their recorded results eventually make input ready for another exchange.

Asking for a tool and running it are separate facts. The model's reply preserves text and calls together, in order. A later dispatch event says execution began without adding another copy of the call to dialogue. The return supplies the outcome. Chapter 2 can represent and replay these facts without executing tools yet. That capability belongs to the next chapter, not to a hidden scheduler inside a vendor adapter.

The append-only log is readable JSON lines, ordered by sequence number rather than wall-clock time. Timestamps help investigation; they do not settle ordering when clocks tie or disagree. Redaction and failure add events instead of editing earlier ones. The send event identifies the destination without storing a second authoritative request body.

That last economy has a limit: replaying conversation state does not necessarily reproduce every byte of a historical HTTP request. Rendering also depends on configuration and implementation version. A current configuration cannot stand in silently for the one used at the time.

## Replay must not discover new facts

Reduction depends only on the previous context and the next event. It does not consult the network, current time, Engine's private state, or a second copy of history. Replaying a log neither sends another request nor reruns a tool.

The state machine distinguishes idle, input pending, in flight, and tools pending. A known event that has no applicable transition leaves the state unchanged rather than crashing replay. An unknown event kind is different: the loader refuses it before reduction. Ignoring an unfamiliar fact would make a readable-looking context with unknown omissions.

Input remains recordable while work is in flight. Arrival is not permission to cancel, redirect, or restart that work. Those later decisions need explicit, replayable meaning rather than guesses from an error message or the presence of old tool results. Chapter 2 establishes the recording boundary; the interrupt protocol comes later.

Temporary instructions expose why ordering matters. An ephemeral System contribution goes into a pending slot, not permanent dialogue, and its arrival alone does not start a request. A newer contribution replaces it; an empty update clears it. All arrival events remain recorded. Neither operation changes the stable primary instructions.

Rendering cannot consume the slot: rendering the same snapshot twice should give the same request. The send event consumes it, even if the exchange subsequently fails. Therefore Engine builds from the snapshot that still contains the instruction, records the send, and uses that built request. Re-rendering after the send would lose the instruction before delivering it. Keeping it after a failed exchange would let stale advice leak into an unrelated turn. This is a one-request context policy, not a promise of exactly-once delivery across a network.

The source records another confusion between history and current state: a test classified a request as being in a tool loop whenever it found any tool results. Full dialogue was being resent. After the first tool use, the test could never recognize a fresh turn again. Current activity belongs to the turn state and the relevant end of the conversation, not a search for something that once happened.

## Values without duplicate owners

The C# model makes these boundaries physical. Context entries contain ordered, immutable, value-like parts rather than mutable bags of loosely named fields. Text, tool calls, tool results, blob references, opaque replay material, and redaction stubs retain different meanings even when a provider eventually puts several of them in one JSON object.

Constructors copy mutable inputs. Part arrays are not exposed for editing; tool arguments remain privately owned JSON rather than caller-owned dictionaries or JSON trees. Accessors cannot hand out a route back into recorded state. A renderer is allowed to rearrange its output, not mutate the context it received. Sharing genuinely immutable values is harmless; sharing an array that someone still owns and can edit is not immutability.

Value-like does not mean that every part must be a C# value type. A class can have value-like semantics and need no back-pointer to every entry containing it. Garbage collection reclaims it when it is unreachable. History and Engine are different: they are relationship-bearing children of Agent, each retaining its parent reference. Parent add and remove operations establish and clear both ends of the relationship; destruction releases owned resources rather than leaving a half-detached child for some other method to repair.

Common declares the shared types, persistent state, relationships, and explicitly accessible partial operations. Matching C# 14 partial type sections put the bodies in their responsible modules, compiled into one assembly. There is no second object hiding behind an injected service field to connect a declaration to its implementation. Functional modules use the Common vocabulary, not one another's private helpers.

A neutral role records authorship: human, agent, system, or tool. A neutral part records meaning. Tool arguments must preserve JSON numbers and nesting without loss. The source's normalization bug rounded a large integer through floating point and produced valid JSON asking for the wrong thing. Syntactic validity was an unusually convincing disguise.

Calls retain their issued correlation IDs, even when those IDs look foreign to the next provider. Only a genuinely missing ID is synthesized, deterministically from stable event identity; multiple calls must remain distinct. Results refer back to those identities and retain their failure status. A result's function name can be recovered from its matching call rather than copied into permanent state just because one API needs it again.

Blob parts preserve references, not payload bytes: a local path, remote URI, or framework handle remains distinct. A request may eventually encode bytes as base64, but that encoding is not a new historical fact. Merely reading a reference does not authorize opening a file or fetching a URI. An unsupported unredacted blob causes a local rendering error instead of silently disappearing.

The primary prompt is also generated at the rendering boundary, from context and configuration. It can be a constant at this stage. Keeping it out of stored context leaves later tool descriptions and capabilities derivable from their actual definitions rather than a growing collection of hand-maintained claims.

## Translation may change shape, not authorship

The bridge targets Anthropic Messages, OpenAI Chat Completions, and Gemini GenerateContent. Their formats are boundary contracts, not alternative definitions of the agent's history. Provider behavior changes; the rules here describe the chapter's supported surfaces and conservative replay policy, not every model or future API.

A tool result remains tool-authored when Anthropic carries it inside a user message or Gemini places a function response in a user turn. OpenAI Chat Completions has a tool role. Merging several neutral entries into one wire message does not merge their authors or erase their original order in context.

Anthropic's relevant constraint is adjacency and block order, not simply alternating speakers. Results must follow the assistant message containing their calls, with result blocks before ordinary text in the carrying user message. Consider a reply asking for two tools, followed by one result, a human instruction, and the other result. Appending blocks as they arrived produces result, text, result. The renderer must collect the results first, preserving their relative order, then place the remaining content. The disposable request gets reordered. History does not.

An interrupted call is harder. No recorded return means the execution outcome is unknown, not that execution failed or never started. The renderer first searches for a real correlated result, including a later one. Where pairing is required and none exists, this bridge supplies a deterministic wire-only error result: “No result was recorded; execution outcome unknown.” It uses the exact call ID and sits beside the other results before intervening human text. The error describes a missing record, not an observed tool failure.

A subsequent rendering uses a real result if one has since arrived. The placeholder never becomes a historical event. The renderer does not execute anything to complete the picture, borrow another call's result, or guess through duplicate and ambiguous IDs. Unrepresentable pairing fails locally before HTTP.

Gemini's function response needs an object envelope. Wrapping string content is not enough: already-object-valued content must retain the result's neutral failure flag too. Otherwise a convenient pass-through path turns a failed tool into an apparently successful response. On the way back, Gemini may report `STOP` while returning a function call, so pending work is determined from the parsed parts, not that label alone.

Empty text also has meaning. In the source's OpenAI failure, a token-limited response contained an empty string. Parsing discarded it; the next rendering emitted `null` for the bare assistant turn, and the provider rejected the request. A fake service had accepted it. Empty text, absent text, and an absent assistant turn are not interchangeable. An adapter that preserves elaborate tool metadata but drops an empty reply still loses conversation state.

Offline replay can expose these transformations without credentials or a network call. It shows the request that a particular context and configuration produce. It does not prove live provider acceptance. Conversely, one successful provider response does not prove that every boundary preserved meaning.

## Reasoning is not ordinary assistant prose

The canonical model distinguishes readable thinking text from ordinary text and from opaque replay material. In this chapter, readable thinking is marked by `TextPart.Thinking`; opaque bytes have their own representation. That is semantic information, not decoration to recover later by searching for a heading in a string.

Opaque material may belong to a turn or to an individual call. The distinction became necessary when Gemini supplied call-bound thought signatures: an unattached bag of opaque parts could preserve the bytes while losing which call they accompanied. A signature therefore stays with its own call. Missing material cannot be repaired by borrowing a sibling's signature.

Every such artifact needs provenance: vendor, exact model, and API surface. The requested routing alias may differ from the actual producer named in the response, so both remain distinct. Render targets likewise need an explicitly resolved identity. A friendly alias that has not been resolved does not establish permission to replay a producer's bytes.

This bridge uses exact producer matching for opaque replay. It does not assume that a vendor-wide match is sufficient, nor that an HTTP success means a foreign signature was understood. The source's September 2026 probes undermined its initial simple account of cross-model acceptance; some foreign material was accepted, some corrupted or missing material rejected, and some incompatible material could be dropped. Those observations do not establish a general compatibility promise.

The renderer therefore emits opaque material only to its issuing model through the corresponding surface. If a call requires replay material unavailable for the target, rendering fails locally rather than fabricating it or using a bypass marker. Equivalent responses from different providers can yield equivalent dialogue and usage while retaining different provenance; making every field equal would erase a fact.

Readable reasoning can remain in history without being resent as ordinary prose when opaque replay is the appropriate channel. Opaque content is transported, not obeyed. Its bytes do not gain instruction authority because they happen to decode into readable words.

## Redaction changes the working set

The log owns evidence; context owns what is still useful to send. Redaction adds a new event describing a change to the latter. The original payload remains on disk. This is exclusion from future requests, not secure deletion.

Bill's reported experience with Gemini's SDK compaction explains why age alone is a poor selection policy. Compaction summarized the oldest portion of the session, including the task in the opening message. The agent remained fluent but lost what the work was for. A handoff document preserved continuity that the automatic summary had discarded.

The chapter instead gives content categories distinct redaction meanings. Replacing tool-result payloads with stubs retains what was asked. Removing calls as well as results loses more, while leaving surrounding prose. Removing text includes readable thinking text. Summarizing a whole span replaces its contents with compressed prose. These are different losses; a tool result can still contain indispensable evidence, so categories support a policy rather than settle it automatically.

Non-summary redactions filter entries individually. A summary replaces the entire span with one System entry at its starting sequence. The source records an implementation that used the same per-entry loop for both and inserted the summary into every entry. Compaction made the context larger. Summary replacement is a different operation, not another predicate in a filter.

Generated summary prose must be stored in the redaction event because replay cannot regenerate it deterministically. Ordinary stubs can be derived from what they replace. Blob stubs retain the original reference kind and locator, preserving where the bytes live without sending their payload. Context need not maintain a second, ever-growing registry of every redaction; the log already records them.

The same ownership explains the choice not to rely on provider-maintained conversation continuations. This agent needs local control over the next context after redaction and compaction. Resending a projection has a cost, but it leaves that control with the agent. Hosted conversation state is not inherently wrong; it serves a different ownership tradeoff.

## Errors and usage must have one authority

An exchange failure and an unsuccessful tool result mean different things. A connection reset or HTTP rejection belongs to the exchange error path. A compiler that ran and found a type error produced a real tool result marked unsuccessful. Treating that result as a network problem invites retrying broken code while waiting for the connection to improve.

Diagnostics have a different authority again. A usage warning or a local rendering explanation must not become dialogue or duplicate a state-changing exchange failure. Parsers return neutral response and error events; History applies their defined semantics. Error text is an explanation, not a substitute for a structured decision about what the turn does next.

Usage follows the same single-record rule. The provider adapter returns normalized counts in the response event. Agent records that event, and History derives cumulative accounting. Engine measures the exchange but owns no independent mutable lifetime total to reconcile with replay. A response recorded twice or an Engine counter added again during reload would charge the same fact twice.

The shared categories are fresh input, cache writes, cache reads, and output. They do not overlap. Anthropic's reported input is treated as fresh input alongside separate cache categories. OpenAI's input total includes cached reads, so fresh input excludes them. Gemini's prompt total similarly includes cached content; output must include thought tokens where the response reports them separately. These are normalization rules for the supported response shapes, not a reason to guess when another shape appears.

Omitted optional cache fields can mean zero. Supplied negative, fractional, malformed, or inconsistent counts cannot. If cached input exceeds total input, subtracting would produce nonsense; independently valid output can remain known while affected input accounting becomes unknown. A missing response leaves no basis for inventing exact token usage, even when a remote system may have done billable work.

Unknown is not zero, and incomplete usage cannot honestly yield an exact cost. Each component needs its known or unknown status. Known contributions can still be retained, but adding a later well-formed response does not fill a hole in an earlier one. Raw provider usage remains available with a bounded diagnostic for investigation rather than being replaced by a plausible total.

## An old log is still an input contract

Readable JSON does not make compatibility automatic. Malformed lines and unknown event, role, or part tags fail with a line-numbered error. A known event with optional fields absent is a different case: explicit defaults may supply meaning where the schema permits it. Silently mapping a new tag to zero merely hides an unsupported fact.

Stable wire names remain separate from incidental CLR names. New meanings need a schema-version policy: explicit support or rejection, and migrations that produce a new artifact rather than rewriting the original record. The same discipline applies to provenance enums: unknown names and invalid zero values cannot quietly select a vendor or surface. Model identifiers remain an open set compared for identity, not a growing enum of product names.

One canonical history now connects human input, model replies, redaction, accounting, and provider-specific requests. Chapter 1's simple exchange still works, but its record can also distinguish a requested action from a dispatched action and an observed result. The next chapter can build the tool turn on those distinctions without allowing a wire format to decide what actually happened.
