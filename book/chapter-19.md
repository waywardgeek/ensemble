# Chapter 19: Leaving Anthropic for OpenAI

This chapter is about freedom. It is also about saving you money.

I built this codebook on Anthropic's models. Eighteen chapters, each
one graded, each one tested against three vendors. Then Anthropic
shipped a frontier model that forbids modifications to the context
window. No ephemeral content. No compaction. No memory-band
watermarks. The mechanism this book teaches is the thing their latest
model will not let you do.

The fix was simple: bind the encrypted reasoning to the model that
created it, so it cannot be replayed across models or users. OpenAI
shipped that fix. Anthropic froze the context window instead.

So we are voting with our feet. In this chapter the agent
authenticates through your existing ChatGPT subscription. No
separate billing account. No API-key signup. And the cost drops by
an order of magnitude compared to metered API billing.

---

## TL;DR

### The delta

Chapter 18 left the agent with three vendor renderers, a
ModelFeatures table, capability flags, cache breakpoints, and
per-vendor endpoints. This chapter:

1. Introduces a **credential provider** abstraction that decouples
   authentication from inference.
2. Implements **OAuth 2.0 + PKCE** for OpenAI's ChatGPT plan, so
   an existing subscription becomes the API credential.
3. Migrates the OpenAI transport from **Chat Completions** to the
   **Responses API**.
4. Adds a **capability profile** for the ChatGPT plan's
   restrictions.
5. Streams **reasoning summaries** as first-class events, separate
   from assistant output.
6. Implements a **billing mode** state machine with explicit,
   announced transitions.

### Types (conceptual; adapt to your codebase)

```go
// CredentialProvider returns a bearer token for inference.
// The HTTP layer never knows whether the token is an API key
// or an OAuth access token.
type CredentialProvider interface {
    GetBearerToken(ctx context.Context) (string, error)
    Kind() CredentialKind
}

type CredentialKind int
const (
    CredentialAPIKey CredentialKind = iota + 1
    CredentialChatGPTOAuth
)

// OAuthProfile stores the state of one ChatGPT OAuth
// registration. The OIDC sub claim is the account identity.
type OAuthProfile struct {
    ClientID       string    // issued by OpenAI
    Subject        string    // OIDC sub — stable identity
    Email          string    // display only
    AccessToken    string    // secret
    RefreshToken   string    // secret — rotates on refresh
    IDToken        string    // secret — validated on receipt
    ExpiresAt      time.Time // access token expiry
    GrantedScopes  []string  // must include plan-use scope
    HostID         string    // stable per-installation UUID
}

type BillingMode int
const (
    BillingChatGPT BillingMode = iota + 1
    BillingAPI
    BillingChatGPTThenAPI
)
```

### Rules

1. **Credential abstraction.** The HTTP inference layer calls
   `GetBearerToken()` and receives a valid bearer token. It never
   knows whether the token is a static API key or an OAuth access
   token. Both produce `Authorization: Bearer <token>`. The
   credential provider stays off the renderer interface: the engine
   resolves the token into the existing API-key field, so zero
   interface churn across vendors.

2. **PKCE.** Generate a cryptographically random verifier, compute
   the S256 challenge, send the challenge in the authorization
   request, and the verifier in the token exchange. Generate fresh
   `state` and `nonce` for every authorization attempt.

3. **Loopback listener.** Bind to `127.0.0.1` (not `localhost`),
   any available port, path `/auth/callback`. The path must be
   exactly `/auth/callback` and the host must be the literal
   `127.0.0.1`. Only the port may vary.

4. **Host identity.** Generate a stable `urn:uuid:<v4>` once per
   installation. Persist it. Send as `ext_agent_host_id` in every
   authorization request. Never derive from username, hostname,
   email, or MAC address.

5. **Dynamic registration.** The first authorization uses
   `client_id=dynamic_agent_client`. OpenAI returns an issued
   `client_id` on the authorization callback. Persist it before
   exchanging the authorization code. All subsequent requests use
   the issued `client_id`. This is not RFC 7591: `auth.openai.com`
   advertises no `registration_endpoint`. The bootstrap is a
   registration entrypoint, never valid for token exchange.

6. **ID token validation.** Validate the returned JWT: fetch
   OpenAI's published JWKS, check the signature, verify issuer,
   audience (must match issued `client_id`), expiry, and nonce.
   The OIDC `sub` claim is the stable account identity.

7. **Scope checking.** A successful login does not imply plan
   usage. The granted scopes must include
   `chatgpt.tokens.use.direct`. Without it, ChatGPT-plan
   inference is forbidden. Note: the discovery document's
   `scopes_supported` does not advertise this scope.

8. **Proactive refresh.** Refresh the access token when
   `now >= expires_at - 5 minutes`. Each refresh returns a new
   refresh token; the old one is invalidated immediately.
   Serialize refresh per profile. Never allow concurrent
   refreshes for the same profile. Save both new tokens
   atomically before using either.

9. **Terminal refresh errors.** Six errors are terminal:
   `invalid_grant`, `invalid_refresh_token`, `token_expired`,
   `refresh_token_expired`, `refresh_token_invalidated`,
   `refresh_token_reused`. Mark the credential unusable, keep
   the issued `client_id`, re-authenticate. Temporary
   network/5xx errors: retry with backoff. Never delete valid
   credentials on a transient failure.

10. **Responses API format.** ChatGPT-plan requests use the
    Responses API endpoint. They require `store: false` and
    `stream: true`. Fifteen fields are banned on the plan route,
    including `temperature`, `top_p`, `max_output_tokens`,
    `previous_response_id`, and `top_logprobs` (not `logprobs`).
    Model restrictions as a capability data structure, not
    if-branches. `response.completed` is the success signal.

11. **Cache breakpoints.** Set
    `prompt_cache_options.mode: "explicit"` and place
    `prompt_cache_breakpoint` markers on developer-role content
    blocks. Up to four cache writes per request. The top-level
    `instructions` field cannot carry a breakpoint, so the
    constitution must be a developer-role content block. No
    regression from chapter 18's caching behavior.

12. **Reasoning summaries.** Request
    `reasoning.summary: "auto"` on models that support it.
    Stream summary events incrementally. Map them to separate
    observer events: never buffer a full summary, never
    concatenate into assistant text, never drop summaries
    adjacent to tool calls.

13. **Billing mode.** `chatgpt` uses only OAuth credentials.
    `api` uses only the API key. `chatgpt_then_api` prefers
    OAuth, falls back to API only on
    `usage_limit_exceeded` and only when
    `allow_metered_fallback` is true. Announce the billing
    transition before the first metered request. Never silently
    switch billing.

14. **Credential isolation.** Access tokens, refresh tokens, ID
    tokens, and API keys never appear in prompts, agent memories,
    conversation context, logs, debug output, or any
    model-visible state.

15. **Sign-out.** Revoke the refresh token via OpenAI's
    revocation endpoint. Clear local tokens. Retain the issued
    `client_id` and host ID. If remote revocation fails, clear
    locally and inform the user.

### Yours

- File layout for the OAuth subsystem
- Profile storage format (JSON, keychain, or your existing
  secret store)
- How auth status is displayed to the user
- Whether `chatgpt_then_api` is implemented now or deferred
- Multiple-account support: the schema should support it even if
  the UX exposes only one account

### Exercise

```
make grade19
```

Build from `solutions/ch18`. The grader provides a fake OAuth
server and a fake Responses API server. Complete the OAuth flow,
perform inference through the Responses API, refresh tokens
proactively, and stream reasoning summaries as separate events.

---

## The idea in plain words

Every major LLM vendor encrypts reasoning traces so competitors
cannot distill from them. The encryption is a reasonable defense.
What was unreasonable, as of mid-2026, was that none of the three
major vendors bound the encrypted block to the user, session, or
model that produced it.

In August 2026, researchers demonstrated the consequence. They
replayed a frontier model's encrypted reasoning into a cheaper
model behind a jailbreak prompt, and the decoder treated the
replayed block as its own prior thinking and decoded it. From
6,708 public agent trajectories, they extracted 315,320 reasoning
blocks and recovered 62 active API keys, 33 passwords, 24 access
tokens, 7 private keys, and 367 items of personal information.
Sixty-four of those artifacts existed only inside hidden reasoning
and never appeared in the visible response. The total cost was
$720. (Source: Panfilov et al., "Stealing Reasoning Traces from
Proprietary LLM APIs," arXiv:2608.09867, August 2026.)

The proportionate fix is cryptographic: bind the encrypted
reasoning to the user, session, and model that created it. Reject
replays across any of those boundaries. OpenAI shipped exactly
this. Their September 2026 blog post: "We strengthened protections
for hidden reasoning across users, workspaces, organizations, and
model families."

Anthropic chose a different response. Rather than binding the
signature to its source, they froze the context window entirely.
Their latest frontier model, Opus 5.5, permits zero modifications
to the prefix of the conversation. No context engineering. No
memory-band compaction. No micro-handoff watermarks.

Five chapters of this book break under that rule: chapter 2
(ephemeral entries), chapter 13 (ephemeral DOM snapshots), chapter
14 (ephemeral speech-channel view), chapter 15 (context
engineering), and chapter 17 (ephemeral auto-recall). All share
one mechanism: ephemera, temporary content placed in the context
and removed when stale. Freezing the prefix kills ephemera, and
ephemera is the mechanism the second half of this book teaches.
Beyond this codebook, any agent that needs to see its current
state ephemerally is affected: a family assistant tracking a
conversation, a virtual tabletop tracking game state, an
accessibility tool streaming its reasoning to a developer who
cannot see the screen.

The counter-argument, stated fairly: the original disclosure
included "conversation-compaction vulnerabilities," so
prefix-immutability is a genuine mitigation for that vector. The
rebuttal: both vendors received the same disclosure. One concluded
the fix was signature binding. The other concluded it was freezing
the customer's context. The binding fix is narrower, was
independently shipped, and does not break context engineering.
OpenAI's September 2026 post also disclosed the related
"conversation-compaction" surface, and they closed the specific
replay pathway rather than removing the feature entirely.

This chapter migrates the agent to OpenAI. The vendor seam built
in chapter 2, refined across sixteen chapters, tested in chapter
18 against three providers, now faces its first forced migration.
Bill's $2,475 of model spend and eighteen chapters of architecture
depend on the answer: if the seam works as designed, this is a
refactor. If it was aspirational, this is a rewrite.

It is a refactor.


## The credential seam

The existing pattern is a static API key: resolved at startup,
stored on the configuration struct, passed to every request
unchanged. An API key is valid forever, rotates only by manual
replacement, and requires no round trip to validate. The HTTP
layer treats it as a constant.

OAuth credentials are not constants. An access token expires in
about an hour. The refresh token that renews it is invalidated
the moment a new one is issued. A concurrent refresh can lock the
account out of both tokens. The HTTP layer must handle all of this
without knowing any of it.

The `CredentialProvider` interface has two methods.
`GetBearerToken` returns a valid bearer token or an error.
`Kind` returns whether the credential is an API key or an OAuth
token. The HTTP layer calls `GetBearerToken` and places the
result in the `Authorization` header. It has no other
responsibilities.

The API-key provider is a single field wrapping a string. It
returns that string every time and never fails. The OAuth
provider holds a profile, a mutex, and a token-refresh function.
It checks the expiry, refreshes if needed, and returns the
current access token. Both providers produce an identical header.

The credential provider stays off the renderer interface. The
engine resolves the token into the existing `APIKey` field on
the configuration before the renderer ever sees it. A developer
who adds a fourth vendor writes a renderer that calls
`GetBearerToken`. They do not learn that OAuth exists. Three
vendor renderers, built across seventeen chapters, change zero
lines. The seam's job is to make the caller's life boring.


## OAuth from first principles

A local application cannot keep a secret. A web server can store
a client secret on its backend, out of reach, but a binary
running on the user's machine can be decompiled, inspected, and
copied. Authorization Code with PKCE solves this: instead of
proving you possess a long-lived secret, you prove that the app
finishing the flow is the same app that started it.

The mechanism is a verifier and a challenge. Before redirecting
the user to the authorization server, the agent generates 32
random bytes and computes their SHA-256 hash. The hash goes in
the authorization request. The original bytes go in the token
exchange. The authorization server compares them. An attacker who
intercepts the redirect code still needs the original bytes, and
those never left the agent's memory.

The agent binds a TCP socket to `127.0.0.1` on any available port,
with the path `/auth/callback`. The host must be the literal
`127.0.0.1`; the OpenAI documentation says "Do not substitute
with `localhost`." The path must be exactly `/auth/callback`;
their documentation says "`/callback` does not match
`/auth/callback`." Only the port may vary. This makes the coding
agent, for exactly one HTTP request, a web server. A tool-calling engine that reads files and
runs processes now listens on a socket and serves a redirect.

The authorization request includes six things that matter beyond
PKCE. A `state` parameter prevents cross-site request forgery: the
agent generates a random value, sends it in the request, and
rejects any callback that returns a different one. A `nonce`
prevents replay: it is embedded in the returned ID token and
validated on receipt. A stable `ext_agent_host_id` identifies the
installation. An RFC 8707 `resource` parameter
(`https://api.openai.com/v1`) appears on the authorization request
and is repeated at code exchange.

On first registration, the `client_id` is the bootstrap value
`dynamic_agent_client`. The authorization callback returns an
issued `client_id` alongside the authorization code and state.
Persist the issued ID before exchanging the code. All subsequent
requests use it.

This is not RFC 7591. `auth.openai.com` advertises no
`registration_endpoint` in its OIDC discovery metadata. A reader
familiar with OAuth will look for dynamic client registration and
find nothing. The bootstrap `dynamic_agent_client` is the
registration mechanism, and it is never valid for token exchange.

The user's browser opens. They sign in. OpenAI redirects to the
loopback listener. The agent extracts the code, the state, and the
issued client ID, then shuts down the HTTP server. One request
served. The agent returns to reading files. A reader who has
followed this book from chapter 1 will notice the irony: the tool
that runs shell commands now runs a shell command's inverse,
listening for network input instead of producing it.


## Identity is not permission

A successful login proves who the user is. It does not prove the
user has granted plan usage. Two separate checks are required, and
both must pass before inference.

The token exchange returns three tokens: an access token for API
calls, a refresh token for renewal, and an ID token for identity.
The ID token is a signed JWT. Validating it requires fetching
OpenAI's published JWKS (JSON Web Key Set), verifying the
signature against those keys, and then checking five claims:
issuer matches `auth.openai.com`, audience matches the issued
`client_id`, the token has not expired, and the nonce matches the
one sent in the authorization request. The `sub` claim is the
stable account identity. Two profiles can share an email address
and be distinct accounts; the `sub` distinguishes them.

The second check is the granted scope. The token response includes
a list of scopes the user authorized. Among them,
`chatgpt.tokens.use.direct` means plan-usage inference is
permitted. Without it, the user authenticated but authorized
nothing.

The OIDC discovery document at `auth.openai.com` lists
`scopes_supported` as `openid`, `profile`, `email`, and
`offline_access`. It does not advertise the plan-usage scope at
all. The scope is real, it is granted, and the discovery metadata
does not mention it. A developer implementing against the
discovery document alone will build a login that works and an
agent that cannot infer. Inspect the token response itself for
the granted scopes.


## Tokens that expire

The access token is valid for about one hour. The refresh token
is valid for about thirty days. Both numbers are promises the
authorization server can break at any time, so the agent tracks
the expiry and refreshes proactively: when the current time
passes five minutes before `expires_at`, the agent sends a
refresh request before the next inference call.

Refresh-token rotation is a security feature that punishes
mistakes. Each refresh response returns a new access token and
a new refresh token. The old refresh token is dead the instant
the server issues the replacement. If two threads both see an
expired access token and both attempt to refresh, the first one
succeeds and the second one submits a dead token. The server
treats that as a potential replay attack and may invalidate both
tokens.

The fix is mechanical: serialize refresh operations per profile.
A mutex or a single-flight gate ensures that exactly one
goroutine refreshes at a time. The second thread waits and uses
the result. Save both new tokens atomically before using either
of them. If the process crashes between receiving the tokens and
persisting them, the old refresh token is already dead and the
new one is lost. Atomic persistence means writing a temporary
file and renaming it, or using a transaction. Writing two
fields in sequence is one crash away from an unusable profile.

Six refresh errors are terminal: `invalid_grant`,
`invalid_refresh_token`, `token_expired`,
`refresh_token_expired`, `refresh_token_invalidated`, and
`refresh_token_reused`. Any of these means the credential is
unusable. Keep the issued `client_id` and host ID, discard the
tokens, and re-authenticate from scratch. The user signs in
again, but the agent remembers who it is.

Network errors and HTTP 5xx are transient. Retry with
exponential backoff. The critical distinction: a terminal error
is the server saying the credential is dead. A transient error
is the network saying the server is unreachable. The credentials
stay on disk until the server itself rejects them.


## The Responses API

The Responses API is a different wire format for the same
conversation. Messages become input items. System-role messages
become developer-role messages. The streaming event names change.
The success signal changes from HTTP 200 to
`response.completed` in the event stream. The content is
identical; the encoding is not.

ChatGPT-plan requests carry restrictions. `store: false` and
`stream: true` are mandatory. Fifteen fields are banned on the
plan route: `temperature`, `top_p`, `max_output_tokens`,
`previous_response_id`, `background`, `conversation`,
`metadata`, `max_tool_calls`, `moderation`, `multi_agent`,
`prompt`, `prompt_cache_retention`, `safety_identifier`,
`top_logprobs`, `truncation`. Note that `top_logprobs` is
banned, not `logprobs`. When OpenAI relaxes a restriction, one
row in a data table changes. The renderer reads the table; there
are no if-branches for the plan route.

The top-level `instructions` field is legal on the plan route.
It is also useless for caching. `instructions` is typed as a
plain string, and cache breakpoints attach to content blocks.
A constitution placed in `instructions` cannot carry a
breakpoint, which means every turn pays the full input price for
the system prompt. Encode the constitution as a developer-role
content block with a `prompt_cache_breakpoint`, exactly as
chapter 18 placed markers on Anthropic's system blocks. The
Responses API supports explicit caching: set
`prompt_cache_options.mode` to `"explicit"`, place up to four
breakpoints, and the economics match chapter 18's analysis.
GPT-6.1 Sol charges $2 per million input tokens, $0.10 cached,
$2.50 for cache writes, $10 output. Cache reads are 0.05x
input; cache writes are 1.25x. The same ratios.

The migration is per-model. `SurfaceResponses` is a fact in the
ModelFeatures table, set on specific model rows. Earlier models
stay on Chat Completions; newer models move to Responses. A test
pins the boundary so that a blanket vendor-wide flip cannot
silently move tested models onto an untested surface.

Four artifacts, named in earlier chapters, were waiting for this
migration. `provenance.go` declared `SurfaceResponses` and never
wired it. `delta.go` documented that `StreamAll` deliberately
excluded `StreamReasoningSummary`. `openai.go` carried a comment
that encrypted reasoning was "a Responses-API concept ... for a
renderer that can use it." The `gpt-5.6-sol` row recorded that
function tools with reasoning were refused on Chat Completions,
and that the fix was the Responses API. Chapter 19 is someone
doing it. The migration does not only change authentication: it
unlocks reasoning-plus-tools on OpenAI models.


## Reasoning summaries

The model's reasoning has been invisible for the entire book.
Encrypted traces flow through the wire, and the agent strips
them. Reasoning summaries are the first window: model-generated
descriptions of private thinking, streamed as events, separate
from the assistant's visible answer.

Bill has no macular vision and reads code at about 80 words per
minute. He listens to model output through a screen reader at
about 750 words per minute. A reasoning summary that arrives as
a single paragraph after the answer is useless at that rate. The
value is in the delta: hearing the model think one phrase at a
time while the answer is still forming, and sending a hint to
redirect the model before it finishes the wrong approach. This
is the accessibility payoff for the entire migration.

> The reasoning-summary gate was treated as a research question
> before any code was written. Live measurement against
> `gpt-6.1-sol` with `reasoning.summary: "detailed"` and effort
> set to high produced 314 summary deltas across 3 parts. The
> `auto` setting produced 186 deltas. The `concise` setting
> produced 4. Summaries stream, they produce enough content for
> real-time listening, and the variation across settings is
> large enough to warrant a configuration knob.
>
> The first probe returned zero summary deltas and looked like a
> capability gap. The prompt was trivial and generated 18
> reasoning tokens. Summaries scale with reasoning performed. A
> single sample nearly produced a "this does not work" report.
> Measure with a prompt that makes the model think.

The observer receives summary events through the existing
observation interface. Summary deltas are a new event kind,
separate from assistant-text deltas. They are delivered
incrementally, kept separate from the assistant's visible text,
and preserved when a tool call is adjacent. A response that
interleaves reasoning, tool calls, and assistant text must emit
every event in stream order. The observer does not have to reassemble anything.


## The billing state machine

The agent now has two funding sources: a ChatGPT subscription
and a metered API key. Switching between them is a state
transition that the user must authorize and the agent must
announce.

Three modes. `chatgpt`: all inference uses the OAuth credential.
`api`: all inference uses the API key. `chatgpt_then_api`:
inference starts on the subscription and falls back to the API
key only when the subscription quota is exhausted. The fallback
requires an explicit flag (`allow_metered_fallback`) to be true.
The safe default is false: a surprised bill is worse than a
paused agent.

The error codes that drive transitions are specific.
`subscription_sharing_usage_limit_exceeded` is the signal for
quota exhaustion; if fallback is allowed, the agent transitions
to API billing and announces the switch before sending the first
metered request. `subscription_sharing_usage_unavailable` means
temporary overload: back off, do not fall back.
`subscription_sharing_unsupported_capability` means the request
used a feature the plan route does not support: adapt the request
or report the limitation. A 401 means the credential is invalid:
refresh the token, and if refresh fails, re-authenticate.

The announcement rule is the mechanism that prevents silent cost.
Before the first metered request, the agent emits a clear
observer event naming the billing source. The user hears it (or
sees it, or reads the log) before any money is spent. Metered
fallback is a decision, never an accident.

Model discovery uses the OAuth token. `GET /v1/models` returns
the models available under the subscription, filtered by
`visibility: "list"`. The `slug` field goes in the request; the
`display_name` goes in the UI. The same endpoint, hit with an
API key, returns a different model list reflecting the API-key
tier's access. Two credentials, two views of the same catalog.

While this chapter was being written, OpenAI shipped first-class
conversation compaction on the Responses API:
`response_compact_params`, `compacted_response`,
`response.compaction.compacting`. One vendor added a
context-management capability in the same period the other
removed one.


## Exercise, graded

```
make grade19
```

The grader starts a fake OAuth server and a fake Responses API
server, then runs the agent against them. The fake OAuth server
issues real JWTs signed by an ephemeral RSA key pair, serves a
JWKS endpoint, validates PKCE, enforces state and nonce, rotates
refresh tokens, and can simulate terminal errors. The fake
Responses API server validates request format, streams reasoning
summaries and assistant text, returns error codes on demand, and
rejects banned fields. If the student's code does not validate
the ID token signature against the JWKS endpoint, the grader
catches it.

| Check | Points | Section |
|---|---|---|
| credential-provider | 15 | §19.2 |
| oauth-flow | 15 | §19.3-4 |
| token-refresh | 15 | §19.5 |
| responses-format | 15 | §19.6 |
| cache-breakpoints | 10 | §19.6 |
| reasoning-summaries | 15 | §19.7 |
| billing-mode | 15 | §19.8 |

**Total: 100 points, 7 checks.**


## Taking it for a spin

```
$ agent auth login openai-chatgpt
Opening browser for OpenAI sign-in...
Listening on 127.0.0.1:54821/auth/callback

  [browser opens, user signs in, OpenAI redirects]

Signed in as bill@example.com (ChatGPT Pro)
Client ID: oai_cid_a1b2c3...
Billing mode: chatgpt

$ agent ask "List the Go files in agent/"
```

The reasoning summary streams first. On a screen reader at 5x
speed, it sounds like the model talking to itself: working
through the question, considering the tool, deciding on the
approach. Then the tool call fires, the directory listing
returns, and the assistant text arrives with the answer. Two
streams, interleaved, each with its own voice.

```
[reasoning] Looking at the directory structure...
[reasoning] The user wants Go files specifically...
[tool_call] list_directory("agent/", pattern="*.go")
[assistant] The agent/ directory contains 14 Go files:
            main.go, engine.go, actor.go, ...
```

The cache lens from chapter 18 confirms the caching. The
developer-role constitution carries a breakpoint. The tools
carry a breakpoint. The compaction bound carries a breakpoint.
Three of four markers land; the tool-array marker is inexpressible
on the Responses API because breakpoints attach to content blocks,
not to the tool array. The system marker covers both.

```
$ agent auth status
Account:  bill@example.com
Subject:  sub_0x7f...
Billing:  chatgpt
Token:    valid (expires in 47m)
Scopes:   openid profile email offline_access
          chatgpt.tokens.use.direct
```

The seam held. Eighteen chapters of architecture, one forced
migration, and the renderers did not change. The conversation
did not change. The tools did not change. The wire format
changed, the authentication changed, the streaming event names
changed, and the agent kept working.
