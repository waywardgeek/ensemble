# Chapter 15: The Invisible Invoice

Bill reported $2,475 in model spend on the first edition through its recall chapter, attributing a few hundred dollars to a caching bug. The answers arrived correctly and on time. Repeated input simply stopped receiving its discount.

Chapter 14 preserves useful request prefixes where possible. That does not establish that a provider caches them. Three questions need separate answers: did the application send a reusable beginning, did the provider report reading it from cache, and is the applicable price known?

Engine owns the instruments: an optional cache lens observes actual sends, provider rendering places supported markers, and a process-local meter aggregates reported usage. History retains canonical response counts, not a mutable dollar balance. Restoring yesterday's conversation must not charge yesterday's tokens to today's display.

## Look at the bytes that left

The lens captures consecutive requests at transport admission, not whenever somebody renders a request offline. It keeps exact wire bytes, a marker-stripped prompt view, and a separate options view. Provider order matters: Anthropic's prompt is compared as tools, system, messages; Chat Completions uses tools then messages; Gemini uses system instructions, tools, contents. JSON property order alone is not the provider's context order.

The diagnostic envelope is new, but the selected properties retain their lexical bytes. Parsing and reserializing them could change escapes, whitespace, key ordering, and number spelling, producing a tidy comparison of bytes that were never sent. Options such as model and routing are separate so they do not obscure prompt-prefix comparisons, but the original wire remains necessary to diagnose those changes.

Provider markers must disappear from the comparison because the stable-end marker advances as conversation grows. They must not disappear indiscriminately: a tool argument or schema can legitimately contain a field named `cache_control`. The implementation recognizes legal marker locations and copies the other byte spans unchanged.

The first-edition stripper matched one compact marker spelling. The actual formatter inserted spaces. Hand-built compact fixtures agreed with the broken stripper; a fixture produced by the real formatter exposed it. An instrument that inspects generated artifacts needs examples of those artifacts, not just examples of its author's expectation.

Each prediction travels with its own request as an immutable value. A shared “latest prediction” lets an old response arriving late borrow a newer request's measurement. Locking that variable prevents a data race without fixing the attribution error. The lens locks capture rotation and comparison, never the HTTP wait.

Capture failure is diagnostic failure, not inference failure. An unwritable directory disables that capture and reports the problem without altering the outgoing request. Captures omit authorization headers but can still contain private prompts, files, and tool arguments; a diagnostic directory is not a publication destination.

## Place markers after rendering

Cache boundaries belong in disposable provider JSON, after tool-call/result adjacency and turn merging. A neutral Context entry is not necessarily a legal final boundary. A final user message may even contain both stable material and request-only ephemera; marking the whole message would bank a tail designed to disappear.

The current Anthropic path explicitly opts in and uses at most four distinct boundaries: the last non-deferred tool declaration, nonempty SYSTEM instructions, the surviving handoff, and the last legal stable conversation block before ephemera. Coincident positions collapse. Empty text, thinking, and unknown opaque blocks are not targets, and deferred declarations cannot receive the tool marker.

The tool boundary can preserve an earlier prefix when SYSTEM changes. The rolling history boundary covers the growing conversation instead of stopping at its first user message. Loaded manuals still belong in SYSTEM even when activation changes the prefix; relocating instructions to improve a cache statistic would break the instruction model. A handoff likewise changes what remains. Markers cannot keep removed traffic alive.

These are route-specific contracts. At this chapter boundary, OpenAI uses Chat Completions with implicit caching and an optional routing key, not the historical explicit-breakpoint experiments described in earlier manuscripts. Gemini retains its implicit path without a new cache-object manager. The routing key is neither a breakpoint nor a credential.

Stability is also not eligibility. Cache minimums are model- and route-specific token counts, not request-file byte counts. Zero reads might mean a small prompt, expiry, routing, changed content, or another provider decision. A fake can award a cache read to any fixture; it cannot establish the real service's threshold. Fixed-reference and rapidly growing-history experiments can warm differently without contradicting each other.

## Counts have an authority; prices have conditions

Usage normalizes into four disjoint buckets: ordinary input, cache writes, cache reads, and output. Anthropic reports disjoint input buckets. OpenAI and Gemini prompt totals include cached reads, which must be subtracted to obtain ordinary input. Gemini output also accounts for separately reported thinking without charging it twice. Streaming usage updates can be cumulative, so they are not separate purchases.

Missing required usage stays unknown. An omitted optional cache bucket can mean zero; a malformed supplied value cannot. Overflow likewise yields unknown rather than a reassuring wrapped number. Session hit rate divides cache-read tokens by all input tokens, including the cold first request. It is neither the latest response's hit rate nor the lens's byte-prefix fraction.

Engine records each successfully completed send once. Offline rendering, fixture parsing, replay, failed HTTP, and cancelled exchanges do not advance this meter. A provider may nevertheless charge for a request whose successful usage never reaches the client. The display is a token-price estimate, not an account balance or invoice.

The actual response identity determines pricing: vendor, response model, and service tier. Counts are grouped under those identities so switching models cannot reprice earlier work as if the new model had done it. The source deliberately verifies only narrow exact price rows rather than guessing from broad model-name prefixes. Missing tier, unknown model, unsupported charges, or unknown counts leave the estimate unknown. A documented price does not establish current model availability.

Dollars are derived with decimal arithmetic from counts and applicable rates. If any contributing row lacks a price, the session total stays unknown; a partial subtotal must not masquerade as the whole bill. Other providers can still show useful token counts without fictional tariffs.

## One meter, several views

The GUI takes one coherent Engine snapshot. Its header combines session cost, latest-response all-input tokens, latest-response cache reads, and session hit rate. The browser formats acknowledged values; it owns neither an accumulator nor a price table. Two tabs and a refresh therefore cannot create different charges for the same work.

A new Engine starts with zero session totals and unknown latest-response counts, even when it restores a conversation. A failed request neither advances nor erases earlier successful usage. Usage notifications remain silent to the speech channel.

The current source establishes these client behaviors with offline and loopback exchanges, not a real C# cache grant or provider invoice. The distinction is useful operationally: the lens locates changed bytes, reported usage shows what the provider said happened, and conservative pricing estimates the consequence without pretending to settle the bill.
