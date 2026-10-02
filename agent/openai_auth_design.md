# CodeRhapsody — OpenAI ChatGPT OAuth + API-Key Provider Design

**Status:** Implementation plan  
**Date:** October 2, 2026  
**Goal:** Add first-class OpenAI support to CodeRhapsody with two independent credential/billing modes:

1. **ChatGPT plan authentication** using OpenAI's supported Sign in with ChatGPT OAuth flow.
2. **Traditional OpenAI API-key authentication** as an independent metered fallback.

The implementation must preserve CodeRhapsody's existing context engineering, memory, actor identity, tool orchestration, and real-time reasoning-summary experience. Authentication and billing must remain orthogonal to those systems.

## 1. Important eligibility constraint

OpenAI currently documents the self-service dynamic ChatGPT-plan OAuth flow for open-source/local applications. OpenAI also says ChatGPT-plan usage is available to selected private clients, while paid or remotely hosted applications need approval through OpenAI's partner process. Because CodeRhapsody is not currently open source, do **not** assume that public distribution of the dynamic OSS flow is automatically permitted.

Implement the functionality behind a capability/configuration flag so it can be exercised for an eligible local installation, while leaving room for an approved private-client configuration if required.

Do not reuse, copy, import, parse, or depend on `~/.codex/auth.json`. CodeRhapsody must own its own OAuth registration and credentials.

OpenAI's documented OSS/local flow dynamically registers the application and gives it its own `client_id`, access token, refresh token, and ID token. It does not require a client secret or API key.

---

# 2. Architectural principles

Authentication must be modeled separately from the OpenAI model provider.

Conceptually:

```text
Agent / Actor
    |
    | produces vendor-neutral context + tools
    v
OpenAI Provider
    |
    +---- OpenAIRequestBuilder
    |
    +---- CredentialProvider
             |
             +---- ChatGPTOAuthCredentials
             |
             +---- ApiKeyCredentials
```

A higher-level billing policy selects which credential provider to use:

```text
OpenAIProvider
    |
    +---- billing_mode = chatgpt
    |
    +---- billing_mode = api
    |
    +---- billing_mode = chatgpt_then_api
```

The provider should not know where OAuth tokens are stored or how OAuth works. It should request:

```text
CredentialProvider.getBearerToken()
```

and receive a credential suitable for the current request.

Likewise, the OAuth subsystem should know nothing about prompts, tools, reasoning summaries, actor memory, or conversation state.

---

# 3. Configuration model

Add an OpenAI configuration section approximately equivalent to:

```yaml
openai:
  model: gpt-6.1-sol

  billing:
    mode: chatgpt_then_api
    allow_metered_fallback: true

  chatgpt:
    enabled: true
    account: default

  api:
    key_env: OPENAI_API_KEY

  reasoning:
    effort: high
    summary: auto
```

Recommended billing modes:

```text
chatgpt
    Only use the user's ChatGPT-plan OAuth credential.
    Never incur API charges.

api
    Only use OPENAI_API_KEY or configured API credentials.

chatgpt_then_api
    Prefer ChatGPT-plan usage.
    Switch to API billing only when explicitly authorized by configuration.

api_then_chatgpt
    Optional; probably unnecessary initially.
```

For released builds, make:

```yaml
allow_metered_fallback: false
```

the safe default.

Automatic transition from included-plan usage to metered API billing must never occur merely because a request failed.

For Bill's personal configuration, `allow_metered_fallback: true` is appropriate.

---

# 4. Core credential interfaces

Define a narrow abstraction such as:

```text
interface CredentialProvider {
    GetCredential(ctx) -> Credential
    RefreshIfNeeded(ctx) -> Credential
    Kind() -> CredentialKind
}
```

Where:

```text
Credential {
    bearer_token
    kind
    account_id
    expires_at
}
```

And:

```text
CredentialKind =
    ChatGPTPlan
    OpenAIAPIKey
```

Do not expose refresh tokens outside the OAuth credential manager.

The HTTP inference layer should only ever see a current bearer credential.

---

# 5. Persistent ChatGPT profile

CodeRhapsody must maintain one or more ChatGPT account registrations.

A profile should contain roughly:

```text
ChatGPTProfile {
    profile_version

    # OpenAI registration
    client_id

    # Verified OIDC identity
    issuer
    subject
    email
    display_name

    # Authorization
    granted_scopes

    # Credentials
    access_token
    refresh_token
    id_token

    access_token_expires_at
    earliest_refresh_at

    # Local metadata
    created_at
    updated_at
    last_successful_refresh_at
}
```

Do not use email as the primary identity.

OpenAI explicitly requires the validated OIDC `sub` to be treated as the account identity. Multiple registrations can potentially have the same email address and must remain distinct.

Maintain the issued `client_id` permanently for that registration. Never replace it with `dynamic_agent_client`.

---

# 6. Host identity

Each CodeRhapsody installation needs a stable:

```text
ext_agent_host_id
```

Generate it once and persist it for the lifetime of that host.

OpenAI recommends a JWK-thumbprint URI derived from a host key pair:

```text
urn:ietf:params:oauth:jwk-thumbprint:<thumbprint>
```

A simpler supported fallback is:

```text
urn:uuid:<uuid-v4>
```

For the first implementation, a persistent UUIDv4 is acceptable and simpler.

Do not derive this value from:

- username
- machine hostname
- email address
- MAC address
- account ID
- filesystem path

It must be opaque and non-identifying.

Suggested storage:

```text
CodeRhapsody auth metadata
    host_id = urn:uuid:...
```

The host ID is not a secret.

---

# 7. First-time OAuth registration

Use OAuth 2 Authorization Code + PKCE and OpenID Connect.

Before opening the browser:

1. Start an HTTP listener bound **only** to `127.0.0.1`.
2. Select an available local port.
3. Use callback path:

```text
/auth/callback
```

4. Generate fresh cryptographically random:
   - `state`
   - `nonce`
   - PKCE verifier
5. Calculate the PKCE S256 challenge.

The redirect URI should look like:

```text
http://127.0.0.1:<port>/auth/callback
```

Do not substitute `localhost`.

OpenAI permits the port to vary between attempts, but the scheme, host, and path must remain consistent.

## Initial authorization parameters

For a new registration:

```text
client_id=dynamic_agent_client
response_type=code
redirect_uri=<exact callback>
resource=https://api.openai.com/v1
agent_name_hint=CodeRhapsody
ext_agent_host_id=<persistent host ID>

scope=
    openid
    profile
    email
    offline_access
    resource.invoke
    chatgpt.tokens.use.direct

state=<random>
nonce=<random>
code_challenge_method=S256
code_challenge=<PKCE challenge>
```

Use OpenAI's current OIDC discovery metadata rather than unnecessarily hard-coding endpoints where practical.

The production issuer is currently:

```text
https://auth.openai.com
```

The token exchange documented by OpenAI currently uses:

```text
https://auth.openai.com/api/accounts/oauth/token
```

OpenAI's OAuth flow does not require a client secret for this dynamic local-client registration.

---

# 8. Browser/callback behavior

Open the authorization URL in the user's default browser.

CodeRhapsody should announce accessibly:

```text
Opening your browser to sign in with ChatGPT.
Waiting for authorization.
```

Do not make successful use of the browser itself a prerequisite for accessible status reporting.

When the callback arrives:

1. Verify that returned `state` exactly matches the pending attempt.
2. Handle OAuth errors before examining a code.
3. For new registration, extract the **issued** `client_id`.
4. Persist that issued client ID before exchanging the code.
5. Reject a successful new-registration callback that lacks the issued client ID.
6. Exchange the authorization code using:
   - issued `client_id`
   - authorization `code`
   - PKCE `code_verifier`
   - identical `redirect_uri`
   - identical `resource`

The issued ID will normally look conceptually like:

```text
oaiapp_...
```

Never save `dynamic_agent_client` as the profile's client ID.

---

# 9. Validate the ID token

Do not treat a successful token exchange as sufficient authentication.

Validate the returned ID token using OpenAI's published JWKS.

Verify at minimum:

```text
signature
issuer
audience
expiration
nonce
```

Requirements:

```text
issuer == expected OpenAI issuer
audience contains/matches issued client_id
nonce == nonce generated for this auth attempt
exp > current time
```

After validation:

```text
subject = verified_id_token.sub
```

becomes the stable OpenAI account identity for that profile.

Only after successful validation may the new profile become active.

Never replace the credentials for one saved account with credentials whose newly validated `sub` belongs to another account.

---

# 10. Verify ChatGPT-plan permission separately

Identity login does **not** imply permission to use ChatGPT plan inference.

Inspect the token response's actually granted scopes.

CodeRhapsody may use ChatGPT-plan inference only if the granted scopes include:

```text
chatgpt.tokens.use.direct
```

If the user successfully authenticates but does not grant that scope:

```text
profile.authenticated = true
profile.chatgpt_plan_enabled = false
```

Do not attempt plan-backed inference.

Offer either:

```text
Enable ChatGPT plan usage
```

or:

```text
Use OpenAI API key
```

OpenAI explicitly distinguishes identity from plan-use permission.

---

# 11. Credential storage

Treat these as secrets:

```text
access_token
refresh_token
id_token
API key
```

Preferred storage order:

1. Native OS credential/keychain facility.
2. Existing CodeRhapsody secret-store abstraction, if already secure.
3. As a fallback only, a local credential file with owner-only permissions.

Never put tokens in:

- source control
- telemetry
- debug logs
- shell history
- crash reports
- analytics
- generated agent memories
- SOUL.md
- conversation transcripts
- prompt context
- support bundles

Never place access or refresh tokens into URLs.

Redact OAuth URLs from logs if they contain `id_token_hint`.

The agent itself should never be given credentials as model-visible data.

---

# 12. Token lifetime and refresh

Current documented lifetimes:

```text
access token: 1 hour
refresh token: 30 days
```

Each successful refresh returns:

```text
new access token
new refresh token
```

The replacement refresh token receives a fresh 30-day lifetime.

Implement proactive refresh.

Recommended policy:

```text
if now >= expires_at - 5 minutes:
    refresh
```

Also respect OpenAI's returned `earliest_refresh_at` if present.

## Critical concurrency rule

Refresh tokens rotate.

Therefore refresh for one profile must be serialized.

Use a per-profile mutex / singleflight mechanism:

```text
profile refresh lock
    |
    +-- process A waits
    +-- process B refreshes
    +-- credentials replaced atomically
    +-- process A consumes new credential
```

If CodeRhapsody can run multiple processes sharing the same credential store, use an inter-process lock or transactional compare-and-swap mechanism.

Never allow two processes to independently refresh the same refresh token.

OpenAI explicitly warns about races involving rotating refresh tokens.

---

# 13. Refresh request

Use the saved issued `client_id`, never `dynamic_agent_client`.

Conceptually:

```text
grant_type=refresh_token
client_id=<issued client id>
refresh_token=<current refresh token>
resource=https://api.openai.com/v1
```

Do not send a reduced scope list during refresh; omit `scope` to preserve the existing grant.

On success, atomically replace both:

```text
access_token
refresh_token
```

plus expiry metadata.

---

# 14. Terminal refresh errors

Treat these errors as requiring reauthentication rather than retry loops:

```text
invalid_grant
invalid_refresh_token
token_expired
refresh_token_expired
refresh_token_invalidated
refresh_token_reused
```

When encountered:

1. Mark the credential unusable.
2. Stop sending inference requests with it.
3. Retain the issued `client_id`.
4. Start normal OAuth reauthorization using that existing client registration.

Do not blindly create a new dynamic client.

Temporary network/5xx failures must not erase otherwise valid credentials.

---

# 15. Returning sign-in

For an already registered profile:

Use:

```text
client_id=<saved issued client_id>
ext_agent_host_id=<same host ID>
```

Generate new:

```text
state
nonce
PKCE verifier
PKCE challenge
```

Optionally supply:

```text
id_token_hint=<previous ID token>
login_hint=<verified email>
```

Do not include `agent_name_hint` on normal reauthorization.

After sign-in, validate the new ID token and ensure the verified identity matches the selected saved profile before replacing its credentials.

---

# 16. Sign-out

Provide:

```text
coderhapsody auth logout openai
```

or equivalent UI.

On logout:

1. Discover/use OpenAI's revocation endpoint.
2. Revoke the refresh token.
3. Stop inference with that profile.
4. Clear access/refresh/ID token values locally.
5. Retain:
   - issued `client_id`
   - non-secret account mapping if desired
   - host ID

If remote revocation cannot be confirmed because of a network error, clear the local credentials but tell the user that server-side revocation was not confirmed.

Do not delete the registered client merely because the user logs out.

---

# 17. OpenAI model discovery

Do not hard-code the ChatGPT-plan model catalog.

When ChatGPT OAuth is active:

```text
GET https://api.openai.com/v1/models
Authorization: Bearer <oauth access token>
```

Filter models where:

```text
visibility == "list"
```

Display `display_name`.

Use `slug` when making requests.

Refresh the list when the active ChatGPT account changes.

This lets a Plus, Pro, or future account receive the account-specific catalog OpenAI currently exposes.

API-key mode may use its normal API model-discovery mechanism independently.

---

# 18. Responses API transport

ChatGPT-plan inference currently uses:

```text
POST https://api.openai.com/v1/responses
Authorization: Bearer <OAuth access token>
Content-Type: application/json
```

Two fields are currently mandatory for the ChatGPT-plan route:

```json
{
  "store": false,
  "stream": true
}
```

Treat a request as successful only when the stream emits:

```text
response.completed
```

Do not treat HTTP 200 or stream creation by itself as success.

A usage failure may arrive after streaming has already begun.

---

# 19. Important ChatGPT-plan preview restrictions

The ChatGPT-plan Responses route currently differs from ordinary API-key Responses usage.

Create a transport capability profile rather than scattering special cases throughout the agent.

For ChatGPT-plan HTTP requests:

```text
store = false
stream = true
```

Do not send currently unsupported top-level fields including:

```text
background
conversation
max_output_tokens
max_tool_calls
metadata
moderation
multi_agent
prompt
prompt_cache_retention
safety_identifier
temperature
top_logprobs
top_p
truncation
user
```

Also:

```text
previous_response_id
```

must be omitted for ordinary HTTP use of this route.

CodeRhapsody therefore needs to send the context required for each request in `input`, which is consistent with CodeRhapsody owning its own memory/context construction.

Do not emit explicit system-message items of the form:

```json
{"type":"message","role":"system"}
```

Use:

```text
instructions
```

or developer-role messages instead.

These limitations are preview behavior and should be represented as capabilities, not assumptions baked into the architecture.

Suggested abstraction:

```text
ProviderCapabilities {
    supports_previous_response_id
    supports_max_output_tokens
    supports_temperature
    supports_hosted_tools
    requires_streaming
    requires_store_false
    supports_reasoning_summary
    ...
}
```

---

# 20. Context engineering integration

CodeRhapsody, not OpenAI, remains authoritative for conversation and agent memory.

For the ChatGPT-plan HTTP route:

```text
Actor memory
Project memory
Working context
Tool transcripts
Selected source files
Summaries
Developer instructions
Current user request
        |
        v
CodeRhapsody Context Renderer
        |
        v
Responses API input[]
```

Do not depend on OpenAI server-side conversation state.

This is particularly important because the current ChatGPT-plan HTTP route does not permit persistent continuation through `previous_response_id`.

Any encrypted reasoning state that CodeRhapsody already chooses to preserve should remain part of the OpenAI-specific rendering layer rather than the abstract event log.

---

# 21. Reasoning configuration

For supported reasoning models, request the desired reasoning effort through the standard Responses reasoning configuration.

Example conceptual configuration:

```json
"reasoning": {
  "effort": "high",
  "summary": "auto"
}
```

Make both independently configurable.

`summary: "auto"` asks OpenAI for the most detailed reasoning-summary mode currently available for the selected model.

OpenAI does not expose raw private reasoning tokens; it does expose model-generated reasoning summaries on supported models.

---

# 22. Real-time reasoning summaries

This is a critical CodeRhapsody feature.

The streaming parser must treat reasoning summaries as first-class output, separate from final assistant text.

Handle events including the current reasoning-summary stream events, conceptually:

```text
response.reasoning_summary_part.added
response.reasoning_summary_text.delta
response.reasoning_summary_text.done
response.reasoning_summary_part.done
```

and normal assistant output events such as:

```text
response.output_text.delta
response.completed
response.failed
response.incomplete
```

The exact event schema should be taken from the installed/current OpenAI SDK or API schema rather than copied literally from this document.

OpenAI's current API exposes incremental reasoning-summary text events.

Map them into CodeRhapsody's existing event system:

```text
ModelReasoningSummaryStarted
ModelReasoningSummaryDelta
ModelReasoningSummaryCompleted

AssistantTextDelta

ToolCallStarted
ToolCallArgumentsDelta
ToolCallCompleted

ResponseCompleted
ResponseFailed
```

Do not concatenate reasoning summaries into ordinary assistant output internally.

Preserve separate channels so speech behavior can differ.

---

# 23. Accessibility and 5× listening workflow

Reasoning summaries should be emitted immediately as they arrive.

Do not buffer a full reasoning-summary paragraph before making it available to CodeRhapsody's speech layer.

Desired pipeline:

```text
SSE delta
    |
    v
OpenAI stream decoder
    |
    v
CodeRhapsody semantic event
    |
    v
speech/event queue
    |
    v
screen reader / TTS
```

Latency between receiving a complete speakable fragment and exposing it to the accessibility layer should be minimal.

Avoid UI-only status.

Every important auth state should have a textual/event equivalent:

```text
Opening ChatGPT sign-in.
Waiting for authorization.
Signed in as <account>.
Using ChatGPT plan.
Refreshing ChatGPT credentials.
ChatGPT usage limit reached.
Switching to OpenAI API billing.
```

If automatic API fallback is enabled, announce the transition **before the first metered request**:

```text
ChatGPT plan limit reached. Switching to metered OpenAI API billing.
```

For interactive configurations, optionally require confirmation.

---

# 24. Tool calling

Keep CodeRhapsody's existing client-side tool system.

The ChatGPT-plan route currently supports function/custom tool patterns but does not support several hosted Responses tools.

Do not redesign CodeRhapsody around OpenAI-hosted tools merely to support this provider.

Continue exposing CodeRhapsody's own:

```text
filesystem tools
shell tools
code-editing tools
search tools
memory tools
agent/actor tools
```

through normal function/custom-tool definitions.

The currently documented ChatGPT-plan route does not support several hosted capabilities including image generation, file search, Code Interpreter, native computer use, hosted MCP/connectors, or Responses `tool_search`.

Represent those as provider capabilities.

---

# 25. Billing/fallback state machine

Implement a small explicit state machine.

Example:

```text
START
  |
  v
ChatGPT credentials available?
  | yes
  v
Send using ChatGPT OAuth
  |
  +--> success --------------------> DONE
  |
  +--> refreshable auth failure ---> refresh ---> retry once
  |
  +--> usage limit
  |       |
  |       +--> paid fallback allowed?
  |               |
  |               +-- no --> report limit
  |               |
  |               +-- yes
  |                     |
  |                     v
  |                API key available?
  |                     |
  |                     +-- yes --> announce fallback --> API request
  |                     +-- no  --> report limit
  |
  +--> unsupported capability
          |
          v
      adapt request if safe,
      otherwise report error
```

Never map every 429/403/503 to "use API key."

Interpret OpenAI's machine-readable error code.

---

# 26. Specific ChatGPT-plan errors

Recognize at minimum:

```text
subscription_sharing_user_not_eligible
subscription_sharing_usage_limit_exceeded
subscription_sharing_usage_unavailable
subscription_sharing_unsupported_capability
subscription_sharing_route_not_supported
subscription_sharing_invalid_user
subscription_sharing_user_unavailable
chatpass_v2_scope_not_authorized
chatpass_v2_invalid_authorization_context
```

Key behavior:

### `subscription_sharing_usage_limit_exceeded`

Do not repeatedly retry.

Do not infer a reset time solely from the error.

It may represent a plan-wide or app-specific limit.

If API fallback is explicitly enabled:

```text
switch to API billing
```

Otherwise expose a usage-management action.

### `subscription_sharing_usage_unavailable`

This is not the same as exhaustion.

Use bounded backoff rather than immediately charging the API account unless the user's configured fallback policy explicitly allows fallback on temporary availability failures.

### `subscription_sharing_unsupported_capability`

Inspect `error.param`.

Do not retry the identical request.

### 401/auth context errors

Attempt token refresh only when appropriate.

If credentials have become terminally invalid, require reauthorization.

OpenAI documents these distinctions explicitly.

---

# 27. API-key provider

Preserve normal OpenAI API-key support completely independently.

Preferred lookup:

```text
OPENAI_API_KEY
```

or CodeRhapsody's existing secret manager.

Do not store the API key in the same semantic record as ChatGPT OAuth credentials.

Conceptually:

```text
ChatGPTProfile
    OAuth account + registration

OpenAIApiProfile
    API credential + optional org/project settings
```

Both eventually produce:

```text
Authorization: Bearer <credential>
```

but their billing, lifecycle, error semantics, and supported capabilities differ.

Do not confuse them merely because both use bearer authentication.

---

# 28. Request rendering

Create one canonical CodeRhapsody OpenAI request model:

```text
OpenAIRequestIntent {
    model
    instructions
    rendered_context
    tools
    reasoning_effort
    reasoning_summary
    multimodal_inputs
}
```

Then pass it through a route-specific renderer:

```text
RenderForChatGPTPlan(intent, capabilities)
RenderForApiKey(intent, capabilities)
```

This is preferable to:

```text
if oauth ...
if oauth ...
if oauth ...
```

throughout the request-building code.

It also gives CodeRhapsody freedom to exploit richer API-key capabilities while remaining compatible with the stricter ChatGPT-plan route.

---

# 29. Provider capability negotiation

At startup/account selection:

1. Fetch the model catalog.
2. Determine the selected auth/billing mode.
3. Construct a capability object.
4. Validate CodeRhapsody's intended request against it.
5. Render accordingly.

Example:

```text
Capabilities for ChatGPTPlanHTTP:
    stream_required = true
    store_required_false = true
    persistent_response_continuation = false
    hosted_file_search = false
    hosted_code_interpreter = false
    reasoning_summary = model dependent
```

This should be data-driven enough to evolve when OpenAI relaxes preview restrictions.

---

# 30. Account UX

Commands or equivalent settings should include:

```text
coderhapsody auth login openai-chatgpt
coderhapsody auth status
coderhapsody auth accounts
coderhapsody auth switch
coderhapsody auth logout openai-chatgpt
```

Example status:

```text
OpenAI

Billing:
  ChatGPT plan

Account:
  waywardgeek@gmail.com

Authorization:
  Connected

Plan usage:
  Enabled

Fallback:
  OpenAI API billing enabled

Model:
  GPT-6.1 Sol
```

Do not print access tokens, refresh tokens, ID tokens, or API keys.

Provide a Manage usage action linking the user to ChatGPT usage settings.

OpenAI's UX guidance asks applications to clearly indicate when the ChatGPT plan is being used and to provide access to usage management.

---

# 31. Multiple accounts

Even if CodeRhapsody initially exposes only one account, make the persisted schema multiple-account capable.

Model:

```text
ChatGPTAccounts {
    active_profile_id

    profiles:
        profile-A
        profile-B
}
```

Each profile owns:

```text
issued client_id
subject
email/display info
tokens
scopes
```

The installation owns:

```text
ext_agent_host_id
```

Do not combine credentials merely because two profiles share an email address.

---

# 32. Usage limits

Do not attempt to reverse engineer ChatGPT limits from Codex's local state.

Use returned OpenAI errors as authoritative for inference admission.

OpenAI currently states that Plus's five-hour usage allowance is shared across participating apps, while that particular five-hour limit does not apply to Pro users. Other limits can still exist.

CodeRhapsody should therefore model availability as:

```text
Unknown
Available
TemporarilyUnavailable
LimitReached
Unauthorized
```

rather than maintaining a homemade token quota estimate.

---

# 33. Logging

Safe logs:

```text
auth attempt started
callback received
state validated
client registration saved
ID token validated
scope chatgpt.tokens.use.direct granted
token refreshed
request ID
HTTP status
OpenAI error code
model
billing route = chatgpt/api
```

Unsafe logs:

```text
authorization header
access token
refresh token
ID token
API key
authorization URL containing id_token_hint
raw credential-file contents
```

Preserve OpenAI request IDs on failures for diagnostics.

---

# 34. Security tests

Required tests:

## OAuth attack resistance

- wrong state rejected
- missing state rejected
- nonce mismatch rejected
- bad ID-token signature rejected
- wrong issuer rejected
- wrong audience rejected
- expired ID token rejected
- callback from non-loopback path rejected
- callback with unexpected client ID rejected
- missing issued client ID on initial registration rejected

## Credential isolation

- tokens never enter agent-visible prompt
- tokens never enter memories
- tokens never appear in logs
- API key never enters OAuth profile
- one account cannot overwrite another account's credentials

## Refresh correctness

- proactive refresh works
- only one concurrent refresh occurs
- replacement refresh token saved atomically
- process restart after refresh uses new token
- invalid/reused refresh token triggers reauthorization

---

# 35. Integration tests

Implement an end-to-end test sequence:

### Test A — first ChatGPT login

```text
fresh install
generate host ID
start callback
browser OAuth
receive issued client ID
exchange code
validate ID token
verify direct-use scope
persist credentials
```

Expected:

```text
auth status == connected
```

### Test B — first inference

Call a trivial request.

Requirements:

```text
store=false
stream=true
```

Verify:

```text
response.completed
```

### Test C — reasoning summaries

Send a coding/reasoning request with:

```text
reasoning.summary=auto
```

Verify CodeRhapsody receives and emits reasoning-summary events incrementally.

### Test D — tool call

Expose a harmless local function tool.

Verify:

```text
tool request received
CodeRhapsody executes tool
tool result returned
model continues
response completes
```

### Test E — refresh

Artificially treat access token as near expiration.

Verify:

```text
one refresh
new access token
new refresh token
successful inference
```

### Test F — process restart

Restart CodeRhapsody.

Verify:

```text
same host ID
same issued client ID
saved account recovered
no login required if refresh remains valid
```

### Test G — API-key mode

Force:

```text
billing.mode=api
```

Verify no OAuth credentials are consulted.

### Test H — fallback

Simulate:

```text
subscription_sharing_usage_limit_exceeded
```

With fallback disabled:

```text
no API request occurs
```

With fallback enabled:

```text
announce billing transition
perform API request
```

---

# 36. Unit tests for request restrictions

Ensure ChatGPT-plan rendering strips or rejects currently unsupported fields.

Examples:

```text
temperature
top_p
max_output_tokens
previous_response_id
background
conversation
```

Do not silently delete semantically important settings unless the behavior is explicitly understood.

Preferred behavior:

```text
Known harmless incompatibility:
    adapt automatically.

Potential semantic change:
    return capability error.
```

---

# 37. Reasoning-summary regression test

Because CodeRhapsody's operator workflow depends heavily on reasoning summaries, add a dedicated regression test.

Given a synthetic streamed response:

```text
reasoning summary delta A
reasoning summary delta B
tool call
reasoning summary delta C
assistant output
```

verify CodeRhapsody emits:

```text
A
B
tool notification
C
assistant output
```

in stream order.

Do not defer all reasoning summaries until completion.

Do not discard summaries surrounding tool calls.

Do not merge them into final answer text.

---

# 38. Migration strategy

Implement in phases.

## Phase 1 — transport cleanup

Before OAuth:

- isolate OpenAI credential interface
- isolate Responses API request rendering
- isolate stream event parser
- introduce provider capabilities

Run existing API-key OpenAI integration through these abstractions.

No behavior change.

## Phase 2 — OAuth subsystem

Implement:

- host ID
- PKCE
- browser launch
- loopback listener
- dynamic registration
- ID-token validation
- credential storage
- refresh
- logout
- account profile

No agent changes.

## Phase 3 — ChatGPT-plan transport

Add:

```text
ChatGPTPlanCredentialProvider
```

and constrained Responses rendering.

Get:

```text
"Hello, world"
```

working through ChatGPT-plan billing.

## Phase 4 — CodeRhapsody agent features

Enable:

- full context rendering
- tools
- images/files where supported
- reasoning effort
- streamed reasoning summaries
- actor event stream

## Phase 5 — API fallback

Implement explicit:

```text
chatgpt_then_api
```

billing policy.

Test metered fallback extensively.

## Phase 6 — distribution/eligibility

Before broadly distributing the ChatGPT-plan option in closed-source CodeRhapsody, confirm that CodeRhapsody's distribution model qualifies for OpenAI's private-client rules or obtain the appropriate approval.

---

# 39. Definition of done

The implementation is complete when all of the following are true:

- CodeRhapsody can authenticate through its own ChatGPT OAuth registration.
- It never reads Codex credentials.
- OAuth tokens survive application restart securely.
- Access-token refresh is automatic.
- Rotating refresh tokens are handled without races.
- ID tokens are cryptographically verified.
- Granted plan-use scope is explicitly checked.
- CodeRhapsody can list models available to the signed-in ChatGPT account.
- CodeRhapsody can stream a full Responses API inference using ChatGPT-plan billing.
- Existing API-key OpenAI usage still works.
- The user can explicitly choose ChatGPT or API billing.
- Optional ChatGPT-to-API fallback works only when explicitly enabled.
- ChatGPT-plan limitations are isolated in a capability layer.
- Existing CodeRhapsody context engineering and memory remain authoritative.
- Tool calls work.
- Reasoning summaries stream in real time.
- Reasoning summaries remain distinct from final assistant output.
- Authentication state and billing transitions are fully accessible through speech/text events.
- No credential ever enters prompts, memories, logs, or agent-visible state.
- Disconnect/logout behaves correctly.
- Unit/integration/security tests cover the cases above.

---

# 40. Implementation priority

The shortest safe path is:

```text
1. Refactor existing OpenAI provider around CredentialProvider.
2. Preserve current API-key behavior.
3. Add OAuthProfileStore.
4. Add stable ext_agent_host_id.
5. Implement PKCE + local callback.
6. Implement dynamic ChatGPT registration.
7. Validate ID token.
8. Store issued client_id + tokens.
9. Implement serialized refresh.
10. Implement ChatGPT-plan Responses capability profile.
11. Make one trivial streamed request.
12. Add current model discovery.
13. Add reasoning summaries.
14. Add CodeRhapsody tools/context.
15. Add explicit API fallback.
16. Harden tests/security.
```

Do not begin by modifying memory, actor identity, context engineering, or agent orchestration. Those are already CodeRhapsody's strengths and should remain independent of OpenAI authentication.

---

# 41. Relevant current OpenAI documentation

Implementation should be checked against the live documentation while coding because this flow is new and currently has preview restrictions.

Primary references:

- Sign in with ChatGPT / plan usage overview.
- Dynamic registration and OAuth sign-in.
- Account management and refresh behavior.
- Token lifetimes and claims.
- Models and Responses inference.
- Current preview restrictions.
- Error semantics and recovery.
- Reasoning summaries.
- Current OpenAI OSS integration walkthrough, published September 28, 2026.

Treat those documents as authoritative over this design if OpenAI changes the protocol.
