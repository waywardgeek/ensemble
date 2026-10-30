# Chapter 23: While the Cache Is Broken — Measure Before Taking the Detour

A long-lived Agent sends much of the same conversation repeatedly. Stable instructions, memory, and accepted dialogue can form a reusable prefix while new work grows at the end. Small changes near the front can instead make apparently similar requests expensive again.

Caching is therefore partly a conversation-design problem. It is also a provider measurement problem. A client can control the bytes it sends; it cannot manufacture a provider-reported cache hit or infer one from a quick response.

## Separate intention from observation

A cache key may influence provider routing, but it does not promise warm entries, stable server placement, or a particular reuse policy. A catalog price can estimate the value of credited input; it cannot establish that an account is entitled to send a request.

Local prefix information and provider usage consequently remain separate observations. Comparing two rendered requests can reveal how much shared material survived. Native usage can report what the provider credited. When reuse is absent or unknown, the display must say that rather than substituting the local estimate.

This distinction is especially useful when debugging. A locally stable prefix with no reported reuse suggests a different question from a request whose first instruction changes every round. Neither observation alone identifies the provider's internal cause.

## Keep the front stable for a reason

System instructions precede stable memory, which precedes recent conversation. Loaded manuals are recorded once, and changing a source file does not silently rewrite yesterday's active instructions. Recall attaches at turn admission instead of being recomputed on every tool round. These choices improve replay as well as prefix stability.

Handoff provides a structural cache boundary. When a surviving handoff exists, its position defines the cut; otherwise the last memory block can supply it. Without either, the application invents no boundary. Provider adapters translate that decision only where their protocol supports it. A Responses request must not inherit a Chat-specific marker merely because the desired optimization is similar.

Correctness still comes first. A committed memory replacement changes context because the older source has been safely replaced. Avoiding that change to preserve a hypothetical hit would turn an optimization into an owner of conversation state.

## Freeze one logical request

Provider activity has two related objects: the logical request the application intends to make, and the transport attempts needed to send it. A credential retry must not become an opportunity to quietly choose a different model, endpoint, funding source, tool set, or replay context.

Capture request configuration and credential-source identity before the attempt. Record actual wire behavior and the accepted response. Failed attempts can have diagnostics without acquiring authoritative completion or usage that the provider never supplied.

Secret-bearing diagnostics need their own handling. Making requests inspectable does not authorize copying credentials into a benchmark artifact. The useful evidence is the route, redacted request shape, provider outcome, and measured usage—not the reusable secret that permitted the call.

## Make a cache experiment falsifiable

The local benchmark fixture freezes prompt and configuration, captures the actual outgoing request, and makes differing provider usage reports observable. It can prove request stability, adapter behavior, and accounting semantics. It cannot establish a commercial service's current caching behavior.

The bounded public Responses series in the source accepted repeated identical requests and reported no input reuse. That is a valid measurement: the requests reached the service, but the reported outcome did not support a reuse claim. It does not explain why the provider withheld reuse or prove that the route never caches.

A separate explicitly enabled subscription-route exercise reported both hits and a miss. That result establishes the bounded path and its observations, not reliable reuse on every future request or a guarantee about subscription cost. A useful report retains both outcomes rather than treating a single hit as a stable operating guarantee.

A denied model request, failed authorization, or unavailable route would be a setup outcome rather than a cache measurement. Keeping those categories distinct prevents a great deal of confident but uninformative benchmarking.

## Authentication is not a routing guess

The ordinary API path and the optional subscription path have separate selection and funding semantics. Loading a manual does not switch the application to subscription billing. Supplying a billing mode does not authorize arbitrary account discovery.

The optional path reads an explicitly selected external credential source as data. It does not overwrite that file, renew it through an unimplemented flow, or fall back silently to an API key when it is missing or expired. Any user-supplied refresh operation remains an explicit boundary and must preserve the chosen credential-source identity. Requests carrying someone else's account identity must not survive a race between token and account selection.

These rules are not peculiar to a cache experiment. An automated assistant that quietly changes its funding route because a request failed can be operationally successful and still violate the user's decision.

## Observe without taking over

The library exposes measurements through its existing Engine and Agent relationships. A minimal external consumer can select configuration, make requests, and inspect snapshots without importing internal partial implementations or creating a second accounting system.

Child and judge usage can contribute to cumulative cost while leaving the parent's latest direct-response and cache view intact. Replayed historical usage does not become a new purchase. This is the same ownership distinction that made pressure enforcement possible in Memory.

The resulting display should help an engineer ask the next useful question: did the request prefix change, did the provider report reuse, what was actually billed or estimated, and which route was selected? A single unlabeled “cache percentage” cannot answer all four.

The conversation is the principal reusable asset. Keeping its authority, ordering, and provenance coherent gives caching a chance to help without giving it permission to rewrite what the Agent knows.

[Implementation chapter](../chapter-23.md)
