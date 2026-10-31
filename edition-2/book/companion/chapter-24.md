# Chapter 24: While the Cache Is Broken

Chapter 23 preserves a stable request prefix so the provider can reuse it. That creates an opportunity, not a cache grant. An agent can produce correct answers, execute tools successfully, and still consume far more input budget than expected because the repeated context receives no cached-token credit.

In the [source chapter's account](../edition-3/chapter-24.md#why-this-chapter-exists), Bill wanted to use a ChatGPT subscription with Ensemble. His paired requests reused the same model and request bytes, yet the public OAuth route produced fewer cache hits than the API-key route. Support replied that caching was not guaranteed. That did not resolve the practical problem: a coding agent repeatedly sends its working context, and successful inference alone does not establish an economical connection.

The chapter's response is an optional detour through Codex's endpoint using an existing Codex sign-in. It is not a replacement authentication system or a prerequisite for later features. If Ensemble's public OAuth route already gives satisfactory reuse for the workload, there is no reason to take it.

## A different connection, not a different Agent

The source defines three explicit OpenAI connection choices. `OPENAI_BILLING_MODE=chatgpt` selects Ensemble's public OAuth registration, `api` selects an API key, and `codex` selects this adapter. Public OAuth and Codex credentials belong to different access paths; changing only the destination URL is insufficient. The adapter does not send Ensemble's public OAuth grant to Codex's endpoint.

The ordinary path keeps the existing ownership boundaries. The authentication component supplies credentials, the Engine sends HTTP and parses responses, and the Agent owns the conversation. Tools still become jobs, results still enter local history, and that history supplies the next request. A second agent loop would duplicate machinery unrelated to the connection problem.

For Codex mode, the CLI selects `EN_CODEX_AUTH_FILE` first, then `CODEX_HOME/auth.json`, then `~/.codex/auth.json`. The adapter reads `tokens.access_token` and `tokens.account_id` together before each request. Rereading lets it pick up credentials renewed by Codex without restarting Ensemble. Codex alone owns renewal: Ensemble neither writes the file nor refreshes another application's token. An API-key Codex profile or a sign-in held only in the operating-system keychain does not provide the file-backed ChatGPT credential this adapter needs.

A missing file or rejected credential produces an actionable sign-in error, not automatic fallback to public OAuth or an API key. In particular, selecting an API key after a Codex failure would turn a repairable authentication problem into an unannounced spending decision. Ensemble's own `auth` commands continue to manage its public registration, not Codex's sign-in.

Supported Responses models go to `https://chatgpt.com/backend-api/codex/responses` with bearer authorization and `ChatGPT-Account-ID`. Incompatible models are rejected before credentials are sent, redirects are refused, and tokens never enter prompts, saves, or logs. These boundaries keep a narrowly selected connection from becoming a general credential-forwarding mechanism.

## Why the cache key was not enough

The [implementation account cited by the source](../../docs/edition-2-notes/codex-endpoint-chapter.md) records an initially convincing failure. A terminal session created a Python function, ran its assertion, and submitted the result. Every request reported zero cached tokens. A second run added a longer stable reference and still received none. The tool loop worked; the purpose of the detour did not.

The missing detail came from Codex's client source: its backend uses the HTTP `session-id` header for cache affinity. The described adapter therefore sends the same stable identity in that header and the body's `prompt_cache_key`. It derives the identity from the Agent's data directory and model, hashing it so the directory itself stays local. A fresh identity on each request would defeat affinity. The body key alone had been insufficient in the recorded runs; neither field guarantees a hit.

The request carries `store: false`, `stream: true`, instructions, and the complete local input. Explicit cache options and breakpoints are omitted on this route. Cached context is still transmitted; the provider's response reports how much received reuse credit. Local history must preserve streamed text, tool results, and eligible encrypted reasoning across requests, while the existing streaming parser, tool loop, and usage accounting remain in place.

A response counts only when its terminal completion arrives. An overload error or interrupted stream is not a completed request with zero cached usage. Confusing those cases would make a connection failure look like evidence about cache behavior.

## What the recorded runs establish

The source reports an October 11, 2026 [paired endpoint probe](../../docs/codex-endpoint-probe.md) using the same GPT-5.6 Sol request body. Both completed public OAuth requests reported zero cached tokens out of 4,998; both completed Codex requests reported 4,864 cached tokens out of 4,998. A further request on each route failed with overload and supplied no cache measurement. This was evidence about those credentials, endpoints, and requests at that time, not a current service-wide comparison.

The implementation account also records a corrected terminal run with GPT-6 Astra at high reasoning. An inert reference supplied a substantial stable prefix while the Agent wrote a function, ran an assertion, and submitted its result through the real tool loop. The first request received no cached tokens. The next two received 8,064 of 8,297 and 8,192 of 8,351 input tokens as cached grants. These were live provider calls around a synthetic task, not a representative production workload or a permanent cache guarantee. The earlier misses matter because they separate successful routing and tool execution from successful reuse.

An external HTTP fake can expose route selection, credential rereading, refusal behavior, stream handling, replay, and accounting without establishing anything about the backend's cache policy. The recorded live run answers a different question: whether that particular connection actually received cache grants during tool use. Neither kind of evidence substitutes for the other, and no new verification is claimed here.

Returning to `chatgpt` selects Ensemble's existing public OAuth flow, with base-URL overrides unset and the Codex credential file untouched. The reason to retire the workaround is satisfactory measured reuse on that route with a stable prefix, not merely a successful login or a plausible answer.
