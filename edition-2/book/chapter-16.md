# Chapter 16: Leaving Anthropic for OpenAI

This chapter is about freedom. It is also about choosing how you pay.

I built this codebook on Anthropic's models. Eighteen chapters of tools,
steering and context management. Then Anthropic
shipped Opus 5.5 with restrictions on changing the context behind
retained thinking. My agent depends on removing stale context. I am
not giving that up to use a newer model.

Protecting private reasoning matters. So does letting the person
building an agent control what it sees. The migration tests whether
the vendor seam can preserve that choice.

So we are voting with our feet. In this chapter the agent
authenticates through your existing ChatGPT subscription. No
separate API-key signup for that route. The subscription has its own
permissions and limits; it is another funding source, not a promise
of unlimited inference or a measured tenfold saving.

---

## TL;DR

### The delta

Chapter 15 left the agent with three vendor renderers, a
ModelFeatures table, capability flags, cache breakpoints, and
per-vendor endpoints. This chapter:

1. Introduces a **credential provider** abstraction that decouples
   authentication from inference.
2. Implements **OAuth 2.0 + PKCE** for OpenAI's ChatGPT plan, so
   an existing subscription becomes the API credential.
3. Adds **Responses API** transport for selected OpenAI models,
   preserving **Chat Completions** for earlier model rows.
4. Adds a **capability profile** for the ChatGPT plan's
   restrictions.
5. Streams **reasoning summaries** as first-class events, separate
   from assistant output.
6. Implements a **billing mode** state machine with explicit,
   announced transitions.

### Types (conceptual; adapt to your codebase)

```go
// CredentialProvider returns a bearer token for inference.
// Both credential kinds produce the same Authorization header.
// Renewal belongs to the provider rather than the renderer.
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
    Subject        string    // validated OIDC account identity
    Email          string    // display only
    AccessToken    string    // secret
    RefreshToken   string    // secret; rotates on refresh
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
   `GetBearerToken()` and receives a valid bearer token without
   implementing renewal itself. Both credential kinds produce `Authorization: Bearer <token>`. The
   credential provider stays off the renderer interface. Ensemble
   owns the auth service; Engine resolves a bearer into request-local
   transport configuration. Rendering and offline replay need no token.

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
   exchanging the authorization code. Reauthorization and token
   exchange use the issued `client_id`; a conflicting returning
   callback ID must be rejected. The bootstrap is a
   registration entrypoint, never valid for token exchange.

6. **ID token validation.** Validate the returned JWT: fetch
   OpenAI's published JWKS, check the signature, verify issuer,
   audience (must match issued `client_id`), expiry, and nonce.
   The OIDC `sub` claim is the stable account identity. Retain the
   issued client ID as well, so registrations remain separate.

7. **Scope checking.** A successful login does not imply plan
   usage. The granted scopes must include
   `chatgpt.tokens.use.direct`. Without it, ChatGPT-plan
   inference is forbidden even if identity was successfully saved.
   Request `openid profile email offline_access resource.invoke
   chatgpt.tokens.use.direct`, with `resource=https://api.openai.com/v1`
   on authorization and code exchange. Inspect the granted scope.

8. **Proactive refresh.** Refresh the access token when
   `now >= expires_at - 5 minutes`. Each refresh returns a new
   refresh token; the old one is invalidated immediately.
   Serialize refresh per profile, including processes sharing its
   storage. Never allow concurrent refreshes for the same session.
   Save both new tokens atomically before using either.

9. **Terminal refresh errors.** Six errors are terminal:
   `invalid_grant`, `invalid_refresh_token`, `token_expired`,
   `refresh_token_expired`, `refresh_token_invalidated`,
   `refresh_token_reused`. Mark the credential unusable, keep
   the issued `client_id`, re-authenticate. Temporary
   network/5xx errors: retry with backoff. Never delete valid
   credentials on a transient failure.

10. **Responses API format.** ChatGPT-plan requests use the
    Responses API endpoint. They require `store: false` and
    `stream: true`, full history in `input`, and namespaced local
    function tools. The plan's HTTP route excludes fields
    including `temperature`, `top_p`, `max_output_tokens`,
    `previous_response_id`, and `top_logprobs` (not `logprobs`).
    Model restrictions as a capability data structure, not
    scattered special cases. Require terminal completion before
    committing output or executing a streamed tool call; preserve
    completed output items even when the terminal `output` is empty.

11. **Cache breakpoints.** On the supported API-key route, set
    `prompt_cache_options.mode: "explicit"` and place
    `prompt_cache_breakpoint` markers on eligible content blocks,
    including the developer-role constitution, within the four-marker
    budget. The measured ChatGPT route rejected the options field;
    its accepted capability profile omits both options and markers.
    Keep the cache lens and actual usage counters on both routes.

12. **Reasoning summaries.** Request
    `reasoning.summary: "auto"` on models that support it.
    Stream summary events incrementally. Map them to separate
    observer events: never buffer a full summary, never
    concatenate into assistant text, never drop summaries
    adjacent to tool calls.

13. **Billing mode.** `chatgpt` uses only OAuth credentials.
    `api` uses only the API key. `chatgpt_then_api` prefers
    OAuth, falls back to API only on
    `subscription_sharing_usage_limit_exceeded` and only when
    `allow_metered_fallback` is true. Announce the billing
    transition before the first metered request. Never silently
    switch billing. Deferring automatic fallback is permitted; this
    chapter's accepted implementation supports only `api` and `chatgpt`.

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

```sh
go run ./cmd/grade-edition2 -ch 16
```

Extend Chapter 15; the accepted result is `solutions/ch16`.
Run the command from the repository root. It builds the core and
its separate GUI; an explicit core path and `-gui DIR` select other
submissions. The grader provides a fake OAuth
server and a fake Responses API server. Complete the OAuth flow,
perform inference through the Responses API, refresh tokens
proactively, and stream reasoning summaries as separate events.

---

## The idea in plain words

An encrypted reasoning block looks safe to store or share because
its owner cannot read it. But the provider's model can. If another
model accepts the same block, the encryption has not established
who may ask for its contents. That distinction made a seemingly
opaque artifact a source of leaked secrets.

In August 2026, researchers demonstrated the consequence. They
replayed a frontier model's encrypted reasoning into a cheaper
model behind a jailbreak prompt, and the decoder treated the
replayed block as its own prior thinking and decoded it. From
6,708 public agent trajectories, they extracted 315,320 reasoning
blocks. The genuine user sessions contained 62 distinct API keys,
33 passwords, 24 access tokens and 7 private keys; 64 of the 704
recovered artifacts were absent from visible chat history. These are
the researchers' findings, not an invitation to test recovered
credentials. The original account's $720 figure describes the
paper's estimated cost for a separate 10,000-trace example, not
the measured cost of this extraction. ([Panfilov et al., August 2026](https://arxiv.org/html/2608.09867v1), §§3.1 and 4.1.)

Binding a reasoning artifact to its authorized context is one
way to close that gap. OpenAI's September 30 report says it
strengthened protections across users, workspaces, organizations
and model families, closed a replay pathway, and added output
checks. That report establishes the mitigation it describes; it
does not disclose enough implementation detail to call it exactly
the paper's proposed cryptographic design. ([OpenAI's incident report](https://openai.com/index/disrupting-a-coordinated-model-distillation-campaign/).)

Anthropic's preserved-thinking policy creates a different integration
constraint. The Opus 5.5 launch applied it to that model and Fable 5.1
for API accounts created on or after August 31, 2026. The current
documentation, checked October 11, ties retained thinking to the
prefix against which it was produced. Changing earlier messages or
tool results can invalidate later thinking. It also documents
server-side editing and client-side compaction that discards stale
thinking, so "no compaction at all" would be the wrong criticism.
The conflict is with keeping those artifacts while independently
editing the earlier context. ([Launch policy](https://www.anthropic.com/claude-opus-5-5),
[preserved-thinking rules](https://platform.claude.com/docs/en/build-with-claude/preserved-thinking).)

Several chapters depend on owning that editing policy: Chapter 2's
ephemera, the changing GUI and speech views in Chapters 12–13,
the cuts in Chapter 14, the later memory in Chapter 22, and Chapter 23's recorded
recall attachments, which survive until a checkpoint. They do not
all have the same lifetime. They do share the ability to stop
sending material without asking a provider to make the cut.
Beyond this codebook, any agent that needs to see its current
state ephemerally is affected by such a constraint: a family assistant
tracking a conversation, a virtual tabletop tracking game state, an
accessibility tool streaming its reasoning to a developer who
cannot see the screen.

Protecting hidden reasoning is a legitimate requirement. So is
choosing a service compatible with the context policy the agent
already implements. This exercise takes the latter route instead
of designing a new policy around preserved thinking. Current
Anthropic examples elsewhere in this edition stay on supported
older models such as Sonnet 5; a newer catalog entry is not proof
that the existing ephemeral-context behavior works with it.

This chapter migrates the agent to OpenAI. The vendor seam built
in Chapter 2 and tested in Chapter 15 against three providers now
faces the migration that motivated the first-edition account.
The investment in the agent depends on the answer: if the seam
works as designed, this is a refactor. If it was aspirational,
this is a rewrite.

It is a refactor.


## The credential seam

The existing pattern is a static API key: resolved at startup,
stored on the configuration struct, passed to every request
unchanged. This client treats it as a constant. The service can
still revoke or expire it; a constant in configuration is not a
promise from the provider.

OAuth credentials are not constants. An access token expires in
about an hour. The refresh token that renews it is invalidated
the moment a new one is issued. A concurrent refresh can lock the
account out of both tokens. The HTTP layer must handle all of this
without implementing token rotation in every renderer.

The `CredentialProvider` interface has two methods.
`GetBearerToken` returns a valid bearer token or an error.
`Kind` returns whether the credential is an API key or an OAuth
token. The HTTP layer calls `GetBearerToken` and places the
result in the `Authorization` header. Renewal stays with the
provider; HTTP can request it after an authentication failure
without taking ownership of the stored tokens.

The API-key provider is a single field wrapping a string. It
returns that string every time and never fails. The OAuth
provider holds protected profile state and coordinates token refresh.
It checks the expiry, refreshes if needed, and returns the
current access token. Both providers produce an identical header.

The credential provider stays off the renderer interface. Ensemble
owns the auth store. Agent selects a funding mode, and Engine
obtains the bearer just before HTTP through its request-local
configuration. The renderer needs only the selected route's public
capabilities: for example, whether tools need a namespace. It does
not need an access token to turn history into bytes.

This distinction also keeps offline replay offline. Rendering a
saved conversation must not refresh a credential or require an
account. A developer adding another renderer learns the neutral
conversation and that provider's wire format, without implementing
OAuth. The seam's job is to make the caller's life boring.


## OAuth from first principles

A distributed local application cannot hide a shared client secret.
A web server can keep one on its backend, but a binary
running on the user's machine can be decompiled, inspected, and
copied. Authorization Code with PKCE solves this: instead of
proving you possess a long-lived secret, you prove that the app
finishing the flow is the same app that started it.

The mechanism is a verifier and a challenge. Before redirecting
the user to the authorization server, the agent generates 32
random bytes, base64url-encodes them without padding as the
verifier, and hashes that verifier string with SHA-256. The
base64url-encoded digest goes in the authorization request; the
verifier goes in the token exchange. The server checks the same
transformation. An attacker who intercepts the code still needs
the verifier, which is sent only in the token exchange.

The agent binds a TCP socket to `127.0.0.1` on any available port,
with the path `/auth/callback`. The host must be the literal
`127.0.0.1`; the OpenAI documentation says "Do not substitute
with `localhost`." The path must be exactly `/auth/callback`;
their documentation says "`/callback` does not match
`/auth/callback`." Only the port may vary. This makes the coding
agent, for exactly one HTTP request, a web server. A tool-calling engine that reads files and
runs processes now listens on a socket and serves a redirect.

The authorization request includes several things that matter beyond
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
requests for that registration use it. A returning callback may
omit the ID; if it supplies a different one, reject it.

The course flow uses this callback-based registration rather than
a separate RFC 7591 registration request. The bootstrap
`dynamic_agent_client` is never valid for token exchange. Send
`agent_name_hint=Ensemble` on first registration and omit it when
reauthorizing. The requested scopes include identity, offline
access, `resource.invoke` and `chatgpt.tokens.use.direct`. These
small fields are part of the protocol, not labels to improvise.
([Registration and sign-in](https://developers.openai.com/siwc/token-sharing-open-source/sign-in).)

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
issuer exactly matches the trusted `https://auth.openai.com` URL,
audience matches the issued
`client_id`, the token has not expired, and the nonce matches the
one sent in the authorization request. The `sub` claim is the
stable account identity. Keep the issued client ID with it: the
same email, or the same person in another workspace registration,
is not a reason to merge token sets. A returning verified identity
must match the account being renewed before replacing its credentials.

The second check is the granted scope. The token response includes
a list of scopes the user authorized. Among them,
`chatgpt.tokens.use.direct` means plan-usage inference is
permitted. Without it, identity may be saved, but plan inference
remains disabled. Refuse that request rather than quietly trying
an API key. Login and permission to spend are separate results.

The discovery document inspected during the original implementation listed
`scopes_supported` as `openid`, `profile`, `email`, and
`offline_access`. It did not advertise the plan-usage scope. The missing catalog
entry did not make the scope optional. A developer implementing against the
discovery document alone will build a login that works and an
agent that cannot infer. Inspect the token response itself for
the granted scopes. The current [sign-in contract](https://developers.openai.com/siwc/token-sharing-open-source/sign-in)
likewise makes that grant a separate inference gate.


## Tokens that expire

The access token is valid for about one hour. The refresh token
is valid for about thirty days. Those are the documented lifetimes
on October 11, not constants to substitute for the returned expiry. The agent tracks
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
A mutex handles goroutines within one process; two clients
sharing a credential file need cross-process coordination too.
The accepted store locks and reloads the profile before deciding
whether to refresh. A second request uses the replacement already
saved by the first instead of rotating the old token again.
Save both new tokens atomically before using either of them. If the process crashes between receiving the tokens and
persisting them, the old refresh token is already dead and the
new one is lost. Atomic persistence means writing a temporary
file and renaming it, or using a transaction. Writing two
fields in sequence can leave an unusable mixed profile. Atomic
replacement prevents that partial local state; it cannot recover a
new token lost between server-side rotation and local persistence.
([Token lifetimes](https://developers.openai.com/siwc/token-sharing-open-source/token-reference),
[refresh and storage](https://developers.openai.com/siwc/token-sharing-open-source/profiles-and-sessions).)

Six refresh errors are terminal: `invalid_grant`,
`invalid_refresh_token`, `token_expired`,
`refresh_token_expired`, `refresh_token_invalidated`, and
`refresh_token_reused`. Any of these means the credential is
unusable. Keep the issued `client_id` and host ID, discard the
tokens, and re-authenticate from scratch. The user signs in
again, but the agent remembers who it is.

Network errors and HTTP 5xx are transient. Retry with
exponential backoff. The critical distinction: a terminal error
is the server saying the credential is dead. A transient exchange
failure does not establish that: a reachable server can return a
temporary error too. Retain the stored credentials. A 401 on
inference alone does not establish revoked refresh credentials;
this implementation attempts renewal once through the same funding
source, and clears tokens only after a terminal refresh rejection.
([Error recovery](https://developers.openai.com/siwc/token-sharing-open-source/errors-and-recovery).)


## The Responses API

The Responses API is a different wire format for the same
conversation. Messages become input items. System-role messages
become developer-role messages. The streaming event names change.
HTTP 200 opens the transport; `response.completed` establishes
completed inference. An incomplete or failed stream does not become
a successful conversation fact merely because text appeared.
The local conversation stays vendor-neutral while its encoding changes.

ChatGPT-plan HTTP requests carry restrictions. `store: false` and
`stream: true` are mandatory, and `input` carries the required
history. The current route profile excludes `temperature`, `top_p`, `max_output_tokens`,
`previous_response_id`, `background`, `conversation`,
`metadata`, `max_tool_calls`, `moderation`, `multi_agent`,
`prompt`, `prompt_cache_retention`, `safety_identifier`,
`top_logprobs`, `truncation`, and `user`. The original printed
list omitted `user` while including `previous_response_id`; the
HTTP route excludes both. The capability table keeps these route
restrictions separate from the model table. A renderer reads the
selected profile instead of repeating policy decisions throughout
its message loop.

Local function tools go inside a namespace on this plan route.
Preserve the returned namespace separately from the function name,
and carry the call ID into the `function_call_output`. A model
asking to read a file is only half the transaction: the next request
must contain the actual result correlated to that call.
([Current plan limitations](https://developers.openai.com/siwc/token-sharing-open-source/preview-limitations),
[models and inference](https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference).)

The top-level `instructions` field is legal on the plan route.
It is a plain string, so it cannot carry its own content-block
breakpoint. That does not make its bytes inherently uncacheable:
a later reusable prefix can include earlier instructions. This
renderer uses a developer-role content block for the constitution
so the supported API-key route can place an explicit boundary there.

The first subscription attempt exposed a route difference that the
original chapter had missed. It returned HTTP 400 with
`invalid_parameter`, naming `prompt_cache_options`. The same explicit
mode had worked on the API-key route. The accepted plan profile
therefore omits both that option and its associated breakpoint
markers. Only the option's rejection was measured; the marker
omission is a conservative choice for the paired mechanism.
Changing credentials did not change the model name, but it changed
which request was admissible.

Keep explicit caching on the supported API-key route: set
`prompt_cache_options.mode` to `"explicit"` and place the markers
on eligible content blocks, within the four-marker budget. The
published GPT-6.1 Sol standard rates for prompts through 272K input
tokens are $2 per million fresh input tokens, $0.10 cached input,
$2.50 cache writes and $10 output, verified October 11. Cache reads
are 0.05x fresh input; cache writes are 1.25x. Those API prices
explain why the boundary matters. They are not a subscription
invoice or proof that a request received a cache grant.
([September 29 model release](https://developers.openai.com/api/docs/changelog).)

The migration is per-model. `SurfaceResponses` is a fact in the
ModelFeatures table, set on specific model rows. Earlier models
stay on Chat Completions; newer models move to Responses. A test
pins the boundary so that a blanket vendor-wide flip cannot
silently move tested models onto an untested surface.

The first-edition coder found four artifacts waiting for this
migration in that implementation. `provenance.go` declared `SurfaceResponses` and never
wired it. `delta.go` documented that `StreamAll` deliberately
excluded `StreamReasoningSummary`. `openai.go` carried a comment
that encrypted reasoning was "a Responses-API concept ... for a
renderer that can use it." The `gpt-5.6-sol` row recorded that
function tools with reasoning were refused on Chat Completions,
and that the fix was the Responses API. This chapter was someone
doing it. The new implementation keeps the same lesson without
requiring those historical filenames: authentication and wire
format have separate seams, and the selected Responses models can
now use reasoning and tools together.


### The stream that had already finished its tool call

After removing the unsupported cache option, the subscription
request reached the model and still failed. The body contained SSE,
but the parser selected its whole-JSON branch. Fixing that decision
exposed the second assumption: a completed `read_file` item had
already arrived in `response.output_item.done`, while the terminal
`response.completed` carried an empty `output` array.

The parser needs both parts of the protocol. Detect the stream
without waiting to buffer its entire body, retain completed items
by output index, and use them when terminal output is empty.
A nonempty terminal output remains authoritative. In either case,
terminal success and usage must arrive before those items become
conversation facts or executable tool calls. The fact that a tool's
arguments have streamed is no permission to execute a response
that subsequently fails.

The captured stream made both defects reproducible without another
model call. A parser test that always repeats every completed item
in the terminal array cannot expose the second defect. The real
route had done less repetition than the fake, which was enough to
lose a perfectly good tool call.


## Reasoning summaries

Reasoning summaries are model-generated descriptions of private
thinking, streamed separately from the assistant's visible answer.
They are not the private trace itself. Earlier adapters already
distinguished displayable thinking from opaque replay material;
Responses needs the same distinction. Preserve encrypted reasoning
for compatible replay without reading or speaking its ciphertext.

Bill has no macular vision and reads code at about 80 words per
minute. He listens to model output at about five times speaking
rate. A reasoning summary that arrives as a single paragraph after
the answer has lost its chance to help him intervene. The
value is in the delta: hearing the model think one phrase at a
time while the answer is still forming, and sending a hint to
redirect the model before it finishes the wrong approach. This
is the accessibility payoff for the entire migration.

> The first-edition coder treated the reasoning-summary gate as a research question
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
> reasoning tokens. The coder suspected there had been too little
> reasoning to summarize. A single sample nearly produced a
> "this does not work" report.
> Measure with a prompt that makes the model think.

That October 2 report is evidence of those probes, not a promise
about every response. In the later student run, the API-key session
produced 90 public-summary deltas. The subscription session requested
`detailed` and used 165 reasoning tokens on its final response, yet
returned no public summary at all. More reasoning is not a guarantee
of a summary. Completion and shutdown must not depend on hearing one.

The observer receives summary events through the existing
observation interface. Summary deltas are a new event kind,
separate from assistant-text deltas. They are delivered
incrementally, kept separate from the assistant's visible text,
and preserved when a tool call is adjacent. A response that
interleaves reasoning, tool calls, and assistant text must emit
every event in stream order. The existing part identities keep the streams associated with their
finals. The GUI can feed summary phrases to the speech queue while
the response continues, without merging them into the answer or
speaking the encrypted item that accompanies them.


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
a plan-usage limit, which can be app-specific rather than the
whole subscription being empty; if fallback is allowed, the agent transitions
to API billing and announces the switch before sending the first
metered request. `subscription_sharing_usage_unavailable` means
temporary overload: back off, do not fall back.
`subscription_sharing_unsupported_capability` means the request
used a feature the plan route does not support: adapt the request
or report the limitation. A 401 needs credential or permission
diagnosis. Attempt the bounded same-source renewal described above;
a terminal refresh error requires reauthorization, while a transient
failure does not justify deleting the account or spending elsewhere.

The announcement rule is the mechanism that prevents silent cost.
Before the first metered request, the agent emits a clear
observer event naming the billing source. The user hears it (or
sees it, or reads the log) before any money is spent. Metered
fallback is a decision, never an accident.

The accepted implementation takes the permitted smaller choice:
it rejects `chatgpt_then_api` and implements no automatic fallback.
Both supported modes announce their selected funding source before
the first request. A successful login selects `chatgpt` for later
launches unless explicitly overridden. Logout leaves that selection
in place, so removing a subscription credential cannot silently
activate an API key already present in the environment.

Model discovery uses the OAuth token. `GET /v1/models` returns
the models available under the subscription, filtered by
`visibility: "list"`. The `slug` field goes in the request; the
`display_name` goes in the UI. The same endpoint, hit with an
API key, returns a different model list reflecting the API-key
tier's access. Two credentials, two views of the same catalog.

Choose from the account's actual catalog and the model routes the
application supports. A name being available from the provider does
not add a missing transport row to the program. The public endpoint
for this exercise is `https://api.openai.com/v1/responses`; it does
not use another application's private credential store or a backend
endpoint. ([Account-specific model discovery](https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference).)


## Exercise, graded

```sh
go run ./cmd/grade-edition2 -ch 16
```

Run from the repository root, optionally supplying the core path and
`-gui DIR` for its separate GUI. Keep the seven original weights:

| Check | Points | Observable behavior |
|---|---|---|
| credential-provider | 15 | Selected credentials authenticate actual requests; an unusable selection cannot become an API-key request |
| oauth-flow | 15 | Real login follows the loopback/PKCE flow, validates identity and checks the inference grant |
| token-refresh | 15 | Expiring credentials rotate and subsequent inference uses the replacement bearer |
| responses-format | 15 | Plan format, completed output and the correlated file result survive the real tool loop |
| cache-breakpoints | 10 | API-key developer content carries an explicit breakpoint; the plan profile omits the paired explicit mechanism |
| reasoning-summaries | 15 | Public summary deltas reach the GUI before completion and remain separate from assistant text |
| billing-mode | 15 | Actual failures preserve the selected funding source without unauthorized fallback |

**Total: 100 points, seven checks; no inherited parity row.**

The grader starts external OAuth and Responses fakes and runs the
actual clients against them. Signed JWTs, a published fake JWKS,
rotating tokens and real tool-result bytes exercise the boundaries
without accessing a personal account. A login that never reaches
inference cannot prove a billing rule; a copied error code cannot
prove that the Agent treated the turn as failed.

The numeric run is a bounded sample of the contract. Review the
actual owner relationships, the full identity-validation cases,
cross-process refresh coordination and atomic replacement separately.
The retained browser tests exercise incremental summary speech;
visible WebSocket events alone do not prove native speech delivery.
Neither a fake summary nor a passing score guarantees that a live
model will provide one.

For a human run, build the core from your implementation directory.
The matching reader snapshot is `solutions/ch16`:

```sh
go build -o /tmp/ensemble-ch18 .
```

Use a fresh workspace and a dedicated profile. This keeps the
exercise's conversation and sign-in separate from another application's
credentials. The profile directory contains secrets; the ordinary
conversation save does not.

```sh
chapter18_work=$(mktemp -d)
cd "$chapter18_work"
mkdir -p skills/base profile
chmod 700 profile
export EN_HOME="$chapter18_work/profile"
export EN_PRIMARY_SKILL=base
cat > skills/base/SKILL.md <<'EOF'
---
name: base
type: primary
description: Read the local scheduling exercise and explain the result.
tools: read_file
---
Read requested files before answering. Keep public explanations separate
from tool results. Never invent file contents or a successful operation.
EOF
cat > schedule.txt <<'EOF'
Two identical workers. Each job uses one worker and cannot be split.
Job duration prerequisites
A   4        none
B   3        none
C   5        A
D   2        A
E   4        B
F   3        C,E
G   2        D,E
H   1        F,G
EOF
```

Sign in through your browser, inspect the safe status and discover
models available to that account:

```sh
/tmp/ensemble-ch18 auth login
/tmp/ensemble-ch18 auth status
/tmp/ensemble-ch18 auth models
```

`auth login --print-url` also prints the authorization URL while
opening the browser. It does not print access or refresh tokens.
The client uses `EN_HOME` when supplied; otherwise it chooses its
own `ensemble` directory under the operating system's user-config
location. Keep the same selection for login and inference. A
verified identity without plan permission is shown as disabled for
inference, rather than silently spending through an API key.

The recorded task used `gpt-6.1-sol`. If it is available to your
account, start the supported Responses route explicitly:

```sh
LLM_VENDOR=openai LLM_MODEL=gpt-6.1-sol \
OPENAI_BILLING_MODE=chatgpt OPENAI_REASONING_EFFORT=high \
OPENAI_REASONING_SUMMARY=detailed \
/tmp/ensemble-ch18 --skills-dir skills chat
```

Enter this prompt and wait for the completed answer:

```text
Read schedule.txt. Find a minimum-duration schedule for the two workers. Give each worker's job intervals and prove that a shorter schedule is impossible.
```

A lower bound is the dependency chain A–C–F–H: 4+5+3+1 = 13.
One feasible schedule places that chain on the first worker and
B0–3, E3–7, D7–9, G9–11 on the second. The Agent must actually
read the file; knowing the answer in advance is no substitute for
checking the tool result in the next request.

A summary may arrive while the model is working. You can send a
hint through the existing chat input, but do not wait for a summary
as the condition for ending the session. After the answer completes,
EOF ends chat normally and saves the conversation. To exercise
API-key billing instead, use the same launch with
`OPENAI_BILLING_MODE=api` and the provider key configured as in
Chapter 0. That is an explicit change of funding source.

When finished with the dedicated profile:

```sh
/tmp/ensemble-ch18 auth logout
```

Logout attempts remote revocation, removes local tokens and keeps
the registration and host identity. If remote revocation cannot
be confirmed, the client reports that fact. It does not turn
subscription billing into API billing on your behalf.


## Taking it for a spin

The actual student API-key run used the scheduling task above with
GPT-6.1 Sol, high reasoning and detailed summaries. It read the
file and produced a valid schedule with makespan 13. The three
requests included one prompt and a hint-triggered continuation;
they were not two independent prompted turns. Ninety summary deltas
arrived. The hint reached the third request, after an intermediate
answer, and the process exited normally in 14.51 seconds.

The subscription path took more work. First came the HTTP400
cache-option rejection. Then came a response containing real SSE
that the parser tried to decode as whole JSON. Even with stream
framing selected, the empty terminal output hid an already-completed
tool item. The corrected parser retained that item and let the
ordinary tool loop continue.

The final subscription run made two requests with no API key
configured. It read the puzzle, put the correlated file result in
the second request, and produced the same optimal 13-unit schedule.
The response completed in 14.887 seconds. Both terminal output
arrays were empty, so this was a direct exercise of the completed-item
recovery rather than a fake that happened to include redundant data.

It also returned zero public-summary deltas despite requesting
`detailed`. The launcher had tied its hint and eventual EOF to
seeing a summary. The task finished; the idle CLI stayed open until
the launcher killed it at 150.03 seconds. That run demonstrates the
answer and tool loop, not graceful shutdown or a saved conversation.
Changing the launcher's completion condition was checked against
captured events; no new paid run was made to turn the old timeout
into a cleaner story.

The cache lens needed its own correction. Its earlier extractor
recognized messages from the old API but omitted Responses `input`.
Comparing only the remaining tools produced a reassuring prediction
while ignoring the conversation. Including the actual input fixed
the instrument. The final plan request pair shared a predicted
92.12% byte prefix, but both provider cache counters stayed zero.
Each input was below 1,024 tokens. Stable bytes, a successful task
and a granted cache hit remain separate observations.

The seam held, with repairs at the seam. The conversation and tools
kept working while authentication, request encoding and streaming
parsing changed. The migration required a Responses renderer; the
claim that no renderer changed would erase the very work the
chapter teaches. Its payoff is smaller and more useful: changing
provider and funding source did not require rebuilding the Agent.
