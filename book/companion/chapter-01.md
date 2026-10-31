# Chapter 1: One Loop, Sixty Billion Dollars

A model provider has done something expensive and difficult. Calling its model should not require pretending that the application has done the same thing.

At the beginning, the assistant is an HTTP client. It sends a model name, instructions, a conversation, and an output budget. It receives typed content and usage. It appends a completed exchange and does it again. The request is small enough to inspect, which is an advantage worth preserving as long as possible.

There is no agent framework yet. There are no tools, background jobs, skills, or memory summaries. Adding them before understanding the first request would make it harder to discover which layer is responsible for a failure.

## A conversation is resent

The provider does not retain this program's conversational state between ordinary requests. The application sends the relevant history again. A question about the second-highest mountain only makes sense if the earlier question and answer are included.

That yields a deceptively effective data structure: a retained list of role-and-content values. The library exposes a snapshot rather than allowing callers to mutate that list. The first chapter deliberately uses a narrow alternating user/assistant conversation. It is a useful local contract, not a claim that the provider's complete protocol forbids every other shape.

Instructions travel separately from that list. So do the requested model and the output budget. Keeping them distinct prevents a tempting early mistake: treating the exact JSON used by one provider as the application's permanent model of a conversation. The next chapter will replace the small message representation, but there is no reason to anticipate its entire vocabulary in this one.

One wire detail deserves attention immediately. A request may accept plain text where a response returns an array of typed blocks. Reading the first block's text works until the provider returns two text blocks, or a block that is not text at all. Walking the array costs almost nothing and avoids making a successful shortcut into an architectural assumption.

The same caution applies to model names. A model identifier remembered from training data can look plausible and still be obsolete. Querying the provider's model catalogue is less glamorous than guessing the newest name, and considerably more dependable. Printed identifiers age; discovery is the durable lesson.

## Put the facts where they belong

The application begins with a few concrete objects. Ensemble owns its Agents and the application logger. Each Agent owns its configuration, conversational state, and the Engine that communicates with the provider. The Engine reaches configuration through its Agent and logging through the Agent's Ensemble parent. Those paths are part of the shared vocabulary, rather than services passed separately into every constructor.

Common contains the type declarations, persistent fields, relationships, and callable signatures. Functional directories supply the matching partial implementations. C# 14 allows the constructor declaration to participate in that separation too. The compiler combines the parts into one type in one assembly; no wrapper object connects the declaration to its implementation.

This is source organization, not a dependency-injection trick. The same object has the same fields on both sides. Implementations use Common's vocabulary instead of reaching into another functional module's private helpers. Application composition is the place allowed to construct and connect the system.

The discipline matters even in a small program. If the Engine needs the Agent's configuration, a parent path expresses why it is entitled to reach it. A copied configuration reference may look harmless, but it creates another fact to preserve when ownership changes. If the path is awkward, the model deserves attention before another constructor parameter is added.

Relationship-bearing objects also know their parents. A constructor registers a new child through the parent's add operation. That operation updates the parent's collection and the child's back-reference together. Removal does the reverse. Validation happens before either side changes, so refusal leaves a consistent model.

A message value is different. It does not need to know every list that contains it. Immutable value-like data can be shared without inventing ownership relationships solely to reclaim memory. The distinction prevents a sound rule about model relationships from turning into back-pointer bureaucracy for every string and vector.

## Memory reclamation is not destruction

C#'s garbage collector can reclaim a disconnected cycle. It cannot decide what removing an Agent means to this application.

Logical destruction closes resources, recursively destroys strongly owned children, removes weak child relationships without destroying their objects, and detaches from every parent through the same complete relationship operations used elsewhere. An object's finalizer is not the place to repair the model. The program needs those postconditions when destruction returns, not whenever memory happens to be collected.

Multiple ownership makes this especially important. A child may legitimately have more than one parent. The model should represent that relationship rather than force it into a tree for convenience. Cleanup must then avoid destroying the same child twice, while still removing every relationship that pointed to it.

This may sound heavy for a first HTTP request. The alternative is to build the system on a simpler-looking ownership fiction and rediscover the missing relationships when the first fleet is introduced.

## Publish a completed exchange

A request attempt and a completed conversational turn are not the same event.

The user message is staged for the request. The reply is decoded and validated. Only a successful answer publishes the user/assistant pair into the retained conversation. A malformed reply must not leave the user half of a failed exchange behind and make the next call violate the program's own conversation contract.

Usage requires a separate decision. The provider may report measured consumption even when the answer is unusable. Those tokens were still consumed. Keeping conversation publication transactional does not justify forgetting their cost.

The implementation therefore distinguishes valid usage from valid answer content. Missing usage is not measured zero. Cumulative counters are calculated with checked arithmetic before either total is published, so an overflow cannot update one half of the accounting and leave the other half old. Error diagnostics should identify the failing operation without leaking credentials or reflecting a remote error body indiscriminately into logs.

These are small choices, but they establish an important habit: each fact has its own acceptance boundary. The conversation, the bill, and the diagnostic are related without being the same transaction.

## Slowness is not permission to retry

The first loop does not hide a failed paid request behind automatic retries. A rate limit is reported. A malformed response is reported. The application does not silently turn one user action into several generations and hope one works.

The request's lifetime is likewise explicit. A default whole-request timeout can abort an otherwise valid long generation. The Engine removes that accidental deadline and propagates the caller's cancellation through both header and body waits. Connection establishment may have a separate bound; abandoning a working generation is another decision.

Retries and supervision arrive later with enough identity and state to explain what they mean. Implementing them early would obscure a very useful observation: exactly one request was attempted, and this is how it ended.

## The growing bill

The first cost lesson follows directly from resending history. A short question near the end of a conversation can be expensive because its request includes everything retained before it. The current prompt is not the unit of input cost. The assembled request is.

Measured input and output totals make that behavior visible from the first turn. Later chapters will add caching, normalization across providers, and context reduction. None can repair accounting that began by treating missing numbers as zero or by reporting only the latest user sentence.

The tiny client is useful because there is nowhere for these facts to hide. You can inspect the outgoing conversation, the returned blocks, the retained pair, and the measured usage. When the assistant becomes more capable, those four observations should remain understandable.

The next step is not a larger loop. It is a better representation of what the loop knows.

*Implementation details: [codebook Chapter 1](../chapter-01.md).*
