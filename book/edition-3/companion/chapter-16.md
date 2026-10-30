# Chapter 16: Leaving Anthropic for OpenAI

A provider migration asks an uncomfortable question: does the agent own its conversation, or merely keep a copy of somebody else's request format? Ensemble already has neutral History, provider renderers, and a transport boundary. Switching services still requires real work. It should not require rebuilding the scheduler or teaching every tool about authentication.

The motivation was a conflict over context control. Ensemble removes stale views, tool output, and selected replay material. Anthropic's newer preserved-thinking policy binds retained reasoning to the prefix that produced it; editing that prefix can invalidate later thinking. Compaction is not categorically forbidden. The conflict is retaining those artifacts while independently changing their earlier context.

Opaque reasoning also carries an authority problem. Research cited in the source recovered encrypted reasoning by replaying it through another model, including secrets absent from visible chat. OpenAI subsequently reported stronger cross-account and model-family protections and a closed replay pathway. Unreadable bytes are not automatically safe bytes: replay needs an authorized context, not merely encryption. Ensemble preserves compatible artifacts with provenance rather than treating them as portable assistant prose.

## Credentials have an owner

The current C# design does not introduce an interface just because API keys and OAuth tokens both become bearer headers. Ensemble owns one concrete Credentials store; each Agent freezes its funding and account selection in Configuration. Engine follows its parent chain to obtain credentials immediately before HTTP. The renderer sees public capabilities, not tokens, and can render saved conversations offline without refreshing an account.

Common declares the shared types and callable partial members. Auth owns their implementations and protected storage. CLI and GUI manage their own temporary callback listeners, but they do not acquire independent token caches. The explicitly configured credential directory is private, created restrictively, and never shared with another application's store. Conversation snapshots, manuals, browser messages, and cache captures are not credential storage.

A local application cannot hide a shared client secret in its binary. Authorization Code with PKCE instead binds completion to initiation: the app sends a hash of a random verifier during authorization and the verifier itself during exchange. Intercepting the one-use code is insufficient. Fresh state binds the callback to the attempt; a nonce binds the signed identity to it.

Some protocol details are unforgiving. The callback uses literal `127.0.0.1` and `/auth/callback`; only the port varies. First registration starts with a bootstrap client ID, then persists the issued ID before exchanging the code. The bootstrap cannot perform that exchange. A failed exchange can therefore leave a valid pending registration without inventing an authenticated account, and an uncertain one-use exchange is not automatically retried.

An unsolicited callback with wrong state does not consume the active attempt. A matched denial or failed exchange does. The listener may need to serve several requests before the real callback arrives. Callback pages expose no code, state, account, or authorization URL. CLI prints the private URL in its human login command; GUI opens it in a separate user-initiated window rather than placing it in the main DOM or conversation. A blocked popup starts no login.

## Identity is not permission to spend

The ID token must pass signature and claim validation, not merely decode as JSON. The narrow C# validator verifies trusted issuer, issued audience, timestamps, nonce, and subject continuity, rejecting ambiguous key selection. An account is not identified by matching email strings across registrations.

Even a valid identity does not authorize plan inference. The token response must grant the plan-usage scope. An identity-only login can be saved while inference remains disabled. Discovery metadata once omitted this scope from its advertised list; following that list alone produced a plausible login without permission to use the plan.

Refresh rotation adds concurrency. The old refresh token becomes unusable when its replacement is issued. Two callers renewing independently can turn a routine expiry into a suspected replay. Credentials serializes local operations and coordinates processes sharing its file, then reloads under the lock before deciding whether renewal is still needed. The second caller can use the replacement already saved by the first.

Both replacement tokens are saved atomically before either is used. That prevents a mixed local pair. It cannot recover a new token lost after server-side rotation but before local persistence. Terminal refresh rejection clears tokens and retains registration; transient failures retain credentials and use bounded retry.

After an inference 401, Engine asks about the exact rejected bearer. Credentials reuses a newer grant or renews under the same coordination; Engine retries the same request body once. A second 401 ends the exchange. Failed streams and quota refusals are not invitations to repeat work or find another payer.

There are only two funding modes: API and ChatGPT plan. No automatic fallback connects them. Successful login saves plan selection with the grant; logout leaves that selection intact. A signed-out next launch refuses instead of noticing an ambient API key and quietly spending it. Explicit API selection can override a launch without rewriting the saved preference.

Local logout clears tokens. Sign-out clears locally first, then attempts remote revocation and reports whether it was confirmed. Destroying the credential owner is neither: it joins operations and releases resources without logging the person out. Account controls are human operations, not model tools, and they do not mutate the running Agent's immutable funding selection.

## A stream is not yet a completed response

Responses changes the wire format, not the canonical conversation. Engine projects messages, correlated function calls and outputs, and compatible opaque reasoning into input items. Plan tools use a namespace; the returned call ID still binds the output to the exact call. A funding change alters disposable projection, not History. Unsupported media refuses explicitly rather than disappearing.

Chat Completions, Responses, and funding remain separate choices. API keys can use Responses. Plan mode requires its supported streaming route. Exact capability rows choose reasoning defaults and explanation policy; neither a model-name pattern nor account catalog visibility establishes renderer support, price, or permission.

The historical migration encountered a revealing stream: a function call was complete in an item-done event, but terminal completion contained an empty output array. A parser that looked only at the terminal array lost the call. A fake that redundantly repeated every item never exposed the bug.

The parser now retains completed items by output index. A nonempty terminal output is authoritative; an empty one can fall back to completed item-done values. Added skeletons and partial argument deltas cannot fill the gap. Only terminal success makes a response executable. A stream that looks promising and then fails authorizes no tools. Missing usage can leave accounting unknown without invalidating an otherwise completed response.

Display progress is less privileged. Text and public reasoning summaries can reach the page while generation continues. Explanation-before-tools policy uses actual arrival order, not a terminal array reordered after the fact. A tool-first response receives correlated refusals through the ordinary actor, with no tool effect. A public summary is not the ordinary explanation that policy requires.

## Summaries are for the listener, not the replay cipher

Bill has no macular vision, reads code at about 80 words per minute, and listens to model output at about five times speaking rate. A summary delivered after the answer misses the chance to interrupt a wrong approach while it develops.

Public summaries therefore have distinct typed events and part identities. They remain separate from assistant text, survive adjacency to tool calls, and can feed speech incrementally. Encrypted reasoning stays in compatible canonical replay data; it is neither browser text nor something to pronounce. A final can contain a suffix absent from its deltas, so the existing speech cursor catches up exactly once. Replay stays silent.

Historical probes demonstrated substantial streaming summaries, but a later successful plan run returned none despite doing reasoning. Its launcher waited for a summary before arranging EOF. The task finished; the process remained open until killed. Optional presentation content cannot be a completion or shutdown signal. The current selected C# row supports `auto` and explicit `off`, not every historical summary setting.

Responses usage flows into the existing Engine meter, with funding kept distinct. Plan tokens never acquire an API-dollar estimate. The cache lens must include Responses `input`: its historical omission compared tools while ignoring the conversation and produced reassuring nonsense. Even after repair, a stable predicted prefix did not establish a cache grant.

The source records local signed-issuer, component, host, and loopback transport checks. It also records a later authorized live session on October 29, 2026: Bill completed CLI login in his own browser, account discovery succeeded, and three isolated plan requests completed with `OK`. Those bounded requests reported zero cache reads and writes. They establish that session's login and inference, not live refresh, revocation, GUI login, audible quality, or a dollar charge.

The migration changed credential acquisition, encoding, and stream assembly. The Agent, Jobs, History, and tool loop remained the owners they had been before it.
