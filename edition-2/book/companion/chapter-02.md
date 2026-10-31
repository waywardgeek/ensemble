# Chapter 2: One Log, Three Vendors

The second vendor was supposed to be the easy one.

In the CodeRhapsody history preserved by the [original chapter](../edition-3/chapter-02.md), Bill began with a single-client agent. An interface was extracted from the Claude client, taking its existing methods as the requirements. Even the message types in its signatures belonged to Claude. When Gemini arrived, its implementation had to pretend to be Claude before it could be itself. The interface had concealed a client without supplying a common model of conversation.

The next implementation was a copy of the first, edited until it worked. OpenAI got another copy. The source account says each passed its tests when it landed; later, fixes had to visit three increasingly different homes and did not always reach them. The eventual redesign used a common context with separate vendor renderers and parsers, and the account reports substantially less code. That is useful evidence of a design problem, not proof that smaller code is always better.

Delivery supplied the qualification. The redesign shipped as a big-bang rewrite. At the point recorded in the source, capabilities still existed only in the old clients, and an audit found audio attachments broken across the new vendor paths. A cleaner seam had not completed the migration. The technical lead inherits both the better design and the obligation to account for what the replacement still cannot do.

An interface can hide a dependency without removing its assumptions.

Chapter 1 left a deliberate weakness in place. One owner kept a simple history, and one exchange established a reliable path from human input to assistant output. That settled who could change the conversation. It did not settle what a conversation was. A slice of role-and-text messages still lets the first API supply the answer.

This chapter replaces that answer without changing the established chat and stdio behavior. The durable record describes what happened. A context describes what those facts currently mean for the conversation. A renderer describes how one vendor needs to hear it. None is a synonym for the others.

That distinction matters more when an AI coding assistant can produce another client quickly. The mechanical work gets cheaper; the duplicated judgment gets harder to notice. The technical lead’s review needs to follow ownership and information loss, not admire how quickly the new package compiles.

## Three Things Wearing One Name

“History” is often asked to do three jobs at once: preserve the past, hold the current state, and serve as the next request body. Those jobs conflict.

A durable log should remember a large tool result even after the result has become useless to the model. The current context should be able to replace it with a small stub. A vendor request may then need to move that stub into a particular role or combine it with another entry. If all three are the same mutable list, a formatting fix can rewrite the past, and compaction can erase the evidence needed to debug compaction.

The source’s remedy is an append-only event log, a reducer, and a vendor-neutral context. History owns the log and its reduction. Agent coordinates changes to it. Engine owns the live exchange and its measured usage, returning facts through Agent rather than editing History itself. Vendor code gets the two edges: rendering requests and parsing responses. It does not get a private copy of the agent’s turn policy.

An ordinary turn makes those boundaries concrete. A human message becomes a recorded event. The reducer adds its content and moves the turn from idle to input-pending. Rendering uses that state and the relevant configuration to build a request. Recording the send moves the turn into flight. The response parser returns neutral events; those events enter the same log and reducer as the human message. A response without tool calls finishes the turn. A response with calls leaves tools pending, and the last returned result makes input ready for the next exchange.

The model’s reply and the tool’s execution are separate facts. A reply can *ask* for a tool without the tool ever running. The reply therefore preserves text and tool calls together, in the order the model produced them. A later `ToolCalled` event records dispatch; it does not duplicate dialogue. A tool return records the outcome. Losing that separation makes interruption history say more than the system knows.

Chapter 2’s design can represent and replay tool events before it can execute tools. That is preparation, not a reason to install a job scheduler early. The data shape must leave room for the next capability; the runtime need not pretend that capability has arrived.

The log orders events by monotonic sequence number, not wall-clock time. Timestamps remain useful evidence, but clocks can disagree or tie. Ordering must survive both. The source uses readable JSON lines so ordinary tools can inspect the record without knowing the application’s types. A damaged conversation is an unfortunate time to discover that its evidence requires the damaged application to open it.

Recorded events are never edited. A redaction is another event. So is a failure. The recorded request event identifies the destination; it need not duplicate the derived request body. Reconstructing that body depends on the relevant state and rendering configuration. Today’s configuration cannot be assumed to recreate every historical byte.

This also changes how a reviewer asks what is happening *now*. The original chapter records a test-harness bug that classified a request as mid-tool-loop whenever it contained any tool results. Full history was being resent. Once a conversation had used a tool, even later new turns still contained those results. The check mistook the past for the present. Current activity belongs to the state machine or the relevant end of the request, not a search over everything the model has ever seen.

## A Reducer, Not a Second Opinion

The reducer’s rule is small: the current context and the next event contain everything required to compute the next context. No call to the network, no secret inspection of Engine, no consultation with a second history owner.

This does not require copying every value. In Go, copying a struct containing slices can leave two apparent contexts sharing the same backing array. A renderer that edits one may quietly edit the other. A clearly owned, in-place reducer is safer than decorative immutability that is not actually immutable.

The turn states are likewise modest: idle, input-pending, in-flight, and tools-pending. A known event that does not cause a transition in the current state leaves that state alone. That makes the reducer total without turning routine replay into a panic lottery. An *unknown event type* is different. The loader refuses it before reduction, rather than silently skipping a fact whose effect it cannot understand.

These rules put classification where the necessary knowledge lives. An input receiver knows that bytes arrived. It may not know whether the agent is already working, or whether those bytes belong to ordinary dialogue or temporary context. Letting the receiver decide produces several competing versions of turn state.

The source’s temporary-context example is especially revealing. An ephemeral contribution arrives as recorded System input, but the reducer puts it in a pending slot rather than dialogue. Arrival alone does not start a model exchange. Another contribution replaces the pending one; an empty update clears it. Every arrival remains in the log, so replay can reconstruct the pending state without a second, unlogged input path. Neither replacement nor clearing changes the stable primary instructions.

The pending contribution is meant for one request. Rendering must not consume it: rendering the same snapshot twice should produce the same result. The recorded send consumes it in the reducer, including on a request that later fails. Otherwise an old hint can ride along on an unrelated turn. If the model quotes the hint in its answer, that answer becomes ordinary dialogue; seeing the words again does not prove the instruction was delivered twice.

The ordering matters. The request must be built from the state that still contains the pending contribution; then the send event advances the recorded state and the already-built request is used. Re-rendering from the consumed state would omit the very material the send was supposed to deliver. A failed HTTP exchange does not retroactively mean that no request was attempted.

This is a one-use *context* policy, not a claim of exactly-once network delivery. A log does not magically make an uncertain remote side effect certain. Its value is that the policy is explicit, replayable, and owned in one place.

Infrastructure errors and tool failures also take different paths. A connection reset or HTTP rejection is an exchange failure. A compiler that ran and reported a type error produced a legitimate tool result, marked unsuccessful. Treating that compiler result as a transport error encourages the agent to retry the same broken build as though the network might become more agreeable.

The parser cannot bypass this arrangement by updating context directly. It returns events. Otherwise it becomes a second reducer, with a second opinion about every invariant and no reliable reason to agree with the first.

## Content Is More Than a String

A vendor-neutral conversation is not a vendor’s message type with the names filed off. It preserves distinctions the agent needs, then lets the wire format compromise them locally.

The context has ordered entries attributed to human, agent, system, or tool. Each entry contains typed parts: text, tool calls, tool results, references to external bytes, opaque replay material, and redaction stubs. A single string cannot preserve those distinctions. Neither can a role borrowed from an API that calls a tool result a human message.

A tool call retains its issued correlation ID, its name and arguments, and the provenance of its producer. A result refers back to the call and keeps its error flag. Correlation is not cosmetic. The renderer cannot answer a call without identifying which call it is answering. Existing IDs survive even when their spelling looks foreign to the target vendor. Only a genuinely absent ID needs a synthesized one, and that synthesis must be deterministic from stable event identity rather than random on every render. Multiple calls still need distinct, consistently matched identities.

Arguments must survive normalization with their meaning intact. The original account describes a JSON-number regression: routing a large integer through floating point rounded it. The resulting request remained perfectly valid JSON and asked for the wrong thing. Preserving JSON number representations fixed the problem; a targeted regression test detected the rounding when that protection was removed. Canonicalization is housekeeping, not a license to alter the model’s request.

A blob part holds a reference, not the bytes themselves. Its kind distinguishes a local path, a remote URI, or a framework handle; its locator preserves the address. Base64 is a possible rendering choice, not a new truth to write into history. This chapter’s renderer can refuse an unredacted blob it cannot represent. It need not fetch a URI or open a local file merely because a record mentions one. Representing a reference is not authority to dereference it.

The distinction survives redaction. A stub can replace the expensive content while retaining the original reference kind and locator. The next request sees less; the durable log still contains the original fact; later code still knows where the bytes live. “Removed from context” does not mean “deleted from storage,” and is not a privacy-erasure guarantee.

The system prompt has a different status: it is generated output, derived from context and configuration, not a stored field inside context. A constant is an adequate generator at this stage. The ownership rule matters more than the sophistication. Hand-maintained tool descriptions tend to drift from actual tools; a mixed pile of generated and sacred handwritten text eventually becomes impossible to change with confidence. Keeping prompt generation at the rendering boundary leaves room for later capabilities without making yesterday’s prompt prose part of the permanent conversation model.

The source tests that distinction by adding renderers in a deliberately inconvenient order: Anthropic, OpenAI, then Gemini. The most alien shape comes last so that success cannot merely mean the last vendor was easy. The measure is changes forced into context, not implementation hours. A difficult renderer can leave the underlying model intact; a short renderer can conceal a missing fact.

The reported first-edition result was not a clean victory. OpenAI required no new context fields. Gemini required one. A function response needed its function’s name, but that was already available on the earlier call with the matching ID. The renderer could look it up; storing the name again would have served only the wire. A call’s thought signature was different. The context could hold opaque material belonging to a turn, but could not bind it to one particular call. That missing relationship required a field on the call itself.

The later implementation described in the source began with that lesson already incorporated and needed no further context fields. That shows successful reuse, not an independent rediscovery that the third vendor costs nothing. The durable review question is whether a proposed field preserves a fact that would otherwise be lost or merely duplicates one vendor’s formatting. The signature relationship passes that test. The repeated function name does not.

## Compaction Should Know What It Is Throwing Away

The original chapter describes Bill watching a Gemini SDK’s built-in compaction summarize away the oldest portion of a session. The task was in the first message. After compaction, the agent remained fluent and confident but could no longer say what it was doing. Bill’s workaround was a handoff document written by the agent for its successor. It became an ancestor of a later mechanism in the book.

The lesson is not that summaries never work. It is that age is a poor proxy for disposability. The oldest message may contain the reason every later message exists.

The source also reports measurements from CodeRhapsody coding sessions in which tool output and tool-call arguments dominated conversation volume. Those measurements motivate a category-based policy; they do not establish the composition of every agent’s workload. A technical lead still has to measure the actual workload and identify which information carries continuity.

The distinction suggests an escalating set of redactions. First, replace tool-result content with stubs while retaining the calls. The conversation still records what the agent asked for, without repeatedly sending the entire answer. A stronger level removes calls and results while retaining the surrounding prose. Another can remove prose and reasoning. Finally, a summary replaces a span with compressed content.

These are different losses. A category purge can name what disappeared. A summary can lose a qualification or a task boundary in a way that becomes visible only when the agent makes its next decision. Tool results can contain indispensable evidence too; the category is a basis for policy, not permission to discard indiscriminately. The shape supports these operations while leaving thresholds and selection to later context engineering.

There is a surprisingly consequential implementation difference between the first operations and the last. The non-summary redactions filter each entry in a span independently. A summary folds the whole span into one entry. The source’s reference implementation initially treated all levels as cases inside the same per-entry loop. It copied the summary into every entry it was meant to replace. Compaction made the context larger. The machinery was tidy; the operation was wrong.

The corrected summary occupies one entry at the beginning sequence of the replaced span, attributed to System rather than impersonating one of the several speakers it summarizes. Its replacement text is recorded in the redaction event because model-written prose cannot be deterministically regenerated. Ordinary stubs can be derived from the superseded content. Replay therefore reproduces both kinds of compacted state without overwriting the source facts.

This is also a memory-ownership problem. A growing map of every sequence ever redacted looks handy, but duplicates facts already preserved in the log and has no natural bound. The current context needs the *result* of redaction, not a second lifetime registry of redactions. Dialogue itself can grow, but it has an intended bounding policy: compaction. An auxiliary map without such a policy is a slow leak waiting for a sufficiently long-lived agent.

Append-only history does not mean all storage is bounded. It means durable evidence and the working set have different lifetimes. Confusing those lifetimes is how a useful cache becomes a second database nobody meant to own.

## Translation Is Allowed to Be Strange

One tool returned “ok.” Who said it?

In the API behavior documented by the source chapter, Anthropic represents the result inside a user message. OpenAI Chat Completions gives it a tool role. Gemini represents it as a function-response part in a user turn. The neutral answer remains: the tool said it. The wire gets the answer it can express.

This is where a renderer earns its keep. It is not merely a serializer. It reconciles a particular API’s constraints without exporting those constraints into the conversation model.

Anthropic’s ordering rule supplies the strongest example. A tool result must immediately follow the assistant message that requested it, and result blocks must precede ordinary text in the user message carrying them. The source distinguishes this from the tempting but inaccurate explanation that consecutive user messages are simply forbidden. The causal rule is adjacency and block order, not a generic demand for alternation.

The single-tool case makes a naive implementation look correct. The assistant requests two tools; result one arrives; a human adds an instruction; result two arrives. Appending rendered blocks in arrival order produces result, text, result. That is wrong for this wire. The renderer must finish grouping the results before placing the remaining content, preserving order within each group. History and context retain the actual arrival order. Only the disposable request receives the rearrangement.

The original account reports that an independent offline replay exposed this defect after the first tests had passed. That is a better review example than a vague request for “more coverage.” The reviewer changed the arrangement of valid facts until a hidden assumption became observable.

Missing results are a harder case. After an interruption, history may contain a requested call with no recorded return. That does not establish that the tool failed, succeeded, or even ran. A renderer must first look for an actual correlated result, including one recorded later. It must not rerun the tool to make the transcript look complete.

The bridge’s chosen recovery policy is a deterministic, wire-only error result with the exact call ID and the text “No result was recorded; execution outcome unknown.” The error flag describes the missing record, not a diagnosed execution failure. The placeholder sits with the other results immediately after the assistant’s calls, ahead of intervening human text. A later real result replaces it in a fresh rendering. Neither the placeholder nor the reordered blocks become new historical facts.

This policy is a design choice, not evidence that an interrupted tool behaved one way or another. Pairing, adjacency, and block order must be validated before sending. If the history cannot be represented safely, a local error is better than malformed HTTP or an invented success.

Gemini has a different translation trap. A function response needs an object, not a bare string. Wrapping a text result solves only half the problem. If an already-object-valued payload is passed through untouched, its neutral error flag can disappear. An envelope has to preserve both payload and failure status on either path. “It was already JSON” is not a semantic guarantee.

Response parsing deserves at least as much scrutiny as rendering. Gemini’s recorded stop signal can say `STOP` even when the response contains a function call. The reducer must inspect the parts to discover that tools are pending, rather than trusting a stop label that means something different on another wire. Equivalent responses should produce equivalent neutral content and accounting while retaining their different provenance. Making the resulting records identical in *every* field would erase who produced them.

A smaller source failure illustrates the same obligation. OpenAI returned an assistant reply with an empty text string after reaching a token limit. The parser treated empty text as not worth recording. On the next request, the renderer emitted `null` where the empty string had been. That bare assistant message was rejected. The fake service had accepted it, so ordinary tests had concealed the loss.

Empty and absent are different facts. A seam that preserves tool calls but discards an empty text part is still lossy. Unsupported media must likewise cause an explicit error rather than vanish. Silent omission makes the output look reasonable precisely when a reviewer most needs it to look broken.

This ownership is why the design declines server-side conversation continuations. They can save retransmission, but the coding agent needs control over what history contributes to its next request. Redaction and compaction operate on that projection. The choice is a product tradeoff, not a claim that hosted conversation state is useless everywhere: this agent pays the replay cost to retain editing control.

## Provenance Is Not Decorative Metadata

An opaque reasoning signature is not ordinary dialogue. It is material issued by a particular producer for replay under particular conditions. The context carries it without treating it as instruction or reconstructed prose. Readable reasoning text can remain in the log; the replay path preserves the opaque material and its origin.

That origin needs vendor, exact model, and API surface. A vendor name alone is too coarse. A routing alias requested by the client may also differ from the actual model identity returned by the provider. The client captures the actual producer when supplied, keeps the requested routing name separate, and uses an explicitly resolved replay target rather than guessing identity from a friendly alias. A target that has not been resolved is not permission to send a signature anywhere convenient.

This information has to be captured when the response arrives. A renderer looking back several turns cannot reconstruct a missing producer identity from the current model setting. A prewritten signature fixture can verify replay selection while saying nothing about whether the real parser captured the signature correctly.

The source records a useful surprise from its dated vendor probes. The initial story was that Gemini would reject another model’s signature while Anthropic would quietly discard it. The reported tests did not support that simple division. Cross-model Gemini replays were accepted at the HTTP level; corrupted signatures and certain missing call-bound signatures were rejected. Anthropic’s behavior depended on whether a receiving model could read the originating model’s block, with some incompatible material quietly dropped.

Acceptance is not proof that a backend used the foreign material. The original chapter explicitly leaves that unobservable. These are observations from the source’s September 2026 probes, not a current compatibility promise or fresh measurements by this companion.

The design consequently chooses the more conservative contract: opaque replay material goes only to the exact issuing model through the appropriate vendor surface. That policy is stronger than “the server happened to accept it.” It avoids making agent continuity depend on a silent interpretation the application cannot inspect.

Call-bound material also stays attached to its own call. One call’s missing signature cannot be repaired by copying a sibling’s signature, and mismatched material is not made safe by a vendor-documented escape hatch. If the target requires material the stored call does not have, the renderer fails locally. A rendered request should not bluff its way past an invariant.

The source’s type choices support the same discipline. Vendor and surface are closed sets where the code needs exhaustive handling; model names are an open set compared for identity, not an ever-growing switch statement. An invalid zero value catches missing provenance rather than silently selecting a vendor. Readable serialized names retain the log’s inspectability, and unknown names fail on load.

Even “opaque” requires active restraint. The original chapter recounts a replay block whose decoded text asked for praise of the course author. The package carried it back to its issuer without obeying it. The anecdote is funny because the boundary is so small: bytes being transported do not gain authority merely because they turn out to contain readable words.

## The Receipt Has to Add Up

Token accounting is another translation boundary disguised as bookkeeping.

The canonical usage record separates fresh input, cache writes, cache reads, and output. Those categories are meant to be disjoint. They have different economic effects, and stable-prefix design cannot be evaluated honestly if cached work is counted as fresh input or omitted altogether.

The source describes three different conventions. Anthropic’s input count already excludes its separately reported cache categories. OpenAI’s prompt total includes cached input, so cache reads and any reported cache writes must be subtracted to obtain fresh input. Gemini’s prompt count likewise includes cached content, while its thinking tokens are reported separately from candidate output and must be included in output accounting. One response object can use a subset convention on the input side and a disjoint convention on the output side. A parser cannot infer arithmetic from the reassuring word “total.”

The source’s probes and API examples support those formulas in the chapter’s stated context. They do not make prices permanent or eliminate the need to check a new surface’s usage contract. The durable log records counts, not dollars. Rates belong in configuration so old sessions can be re-costed without rewriting history. Token counts alone also cannot represent time-based cache-storage charges; the source leaves that extension to later caching work rather than pretending the four counters cover every bill.

Malformed usage needs narrower failure than “throw away the response.” Suppose the prompt total is known, the supplied cache-read field is invalid, and the output count is valid. Fresh input depends on the invalid field, so it is unknown too. The independently measured output remains evidence and must survive alongside a diagnostic about incomplete accounting. An omitted optional cache-write bucket can mean zero; a supplied value that cannot be parsed does not. Overflow, invalid signs, and broken subset relationships cannot be laundered into plausible totals either.

A partly broken measurement is neither wholly trustworthy nor wholly worthless. The sound accounting policy preserves what is independently known, identifies what depends on the broken value, and keeps the uncertainty visible. A confident wrong cost report can influence decisions for months without ever crashing the program.

Historical usage in a replayed context is also not usage incurred by the currently running Engine. History derives the former from response events; Engine owns measurements of its own live exchanges. Offline rendering an old session must not pretend to have sent the old requests again. The distinction keeps accounting tied to events rather than to whichever object happens to be displaying them.

## A Seam Worth Trusting

The architectural test is not whether three happy-path requests succeed. It is whether changing vendors forces the agent to change its account of what happened.

A useful review follows a fact all the way around the boundary: provider bytes become events; events reduce to context; context renders for the next exchange. The empty-string failure lives across that round trip. The large-integer failure lives inside normalization. The interleaved-results failure needs more than one call and an inconveniently timed human message. Each asks a concrete question that a pleasant demo does not.

Determinism supplies leverage. Replaying the same log through the same reducer should produce the same context. Rendering the same context with the same configuration should produce the same bytes and leave the input untouched. Random IDs, unstable map iteration, injected current timestamps, and accidental mutation all undermine that property. Ordered parts are semantic order; sorting them for prettier output would merely make the wrong result reproducible.

A valid replay record also needs a format version and strict interpretation. Current code replays the log; it does not resurrect historical binaries. The version belongs in a header rather than a sequenced event. The source allows a missing header to mean the current format, but refuses facts it cannot interpret: unknown event kinds, parts, actors, or vendors. Sequence numbers must be increasing. A legacy blob’s bare path is refused rather than guessed into a reference kind, and an unspecified kind is invalid. A local-path guess might appear to work until a URI becomes a filename that never existed.

Starting with a broken log and silently skipping the awkward line creates a context no honest sequence of events produced. Redaction changes the projection through its own appended event; it does not rewrite earlier events. Wire-only recovery changes no events at all. A later dump retains the original facts. The log is the evidence, not a workspace for the renderer.

Offline rendering uses the same target configuration as live sending but needs neither a credential nor a reachable endpoint. Credentials belong in neither the durable log nor the rendered request body. That separation makes a saved conversation useful for inspecting a proposed request without accidentally sending it. Saving events still does not establish automatic session resume or crash-safe journaling; those are additional behaviors, not properties bestowed by a JSON-lines file.

The source records a particularly revealing test gap: removing the reference implementation’s only use of call-bound replay material initially left all checks passing. The fixtures tested opaque material belonging to a turn, not to a call. The distinction that had forced the new field was the distinction the tests never exercised. A foreign-origin fixture could correctly withhold a signature and still prove nothing about replay to its issuer.

A later review found the next gap: a saved-signature fixture could pass while the real parser discarded new signatures. Capturing material, associating it with the correct call, withholding it from other targets, and replaying it to its issuer are separate obligations. Likewise, a secret-free rendered redaction stub cannot prove that the reducer preserved the original locator internally. The observation has to reach the property being claimed.

Offline fixtures are valuable precisely because these cases are awkward to provoke live. They cannot establish that a provider actually honored replay material or that a fake has enforced every vendor constraint. Live observations answer different questions, with their own limits. The strongest review does not confuse either form of evidence with the other.

The payoff is not universal compatibility. New APIs will have new limitations, and some histories cannot safely cross a particular wire. The payoff is a place to put that difficulty without spreading it through the agent.

The durable log remembers what happened. The reducer decides what those facts mean now. The renderer translates that state for a particular listener and refuses what it cannot represent honestly. Once those jobs are separate, the next capability can arrive without asking a vendor’s message struct for permission.
