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

## TL;DR — Specification Brief

Add OpenAI Responses and an explicit ChatGPT-plan funding mode without giving
Agents copies of account credentials. One `Credentials` object belongs to the
`Ensemble` root. An Engine reaches it through its Agent and asks for a fresh bearer
immediately before sending. API-key billing remains a separate choice; a failed
plan request never becomes an API charge.

`Common` declares the vocabulary and ownership. `Auth` implements protected
storage, PKCE login, signed identity validation, renewal, discovery and sign-out.
`Engine` implements Responses projection and streaming. CLI and GUI provide human
account controls; neither is another credential store. History records ordinary
conversation and compatible opaque replay data, not OAuth secrets.

The preserved reference endpoint has a scoped 70-point transport grade; signed
local issuer tests separately exercise the credential owner. In the October 29,
2026 joint session, Bill completed real CLI login with Ensemble's own registration.
Account-specific model discovery and three plan-funded Responses calls succeeded
through current main. Renewal, sign-out, GUI login and billing reconciliation were
not exercised live. The cache result was zero reported reuse; see Chapter 23.
These observations supplement, rather than expand, the earlier executable grade.

## The coding brief

Extend the previous chapter's C# solution; do not replace its actor, Jobs,
History, meter or speech queue. Read the current primary protocol references in
this chapter before attempting a live integration. Their changing service details
are separate from the following ownership and safety contracts.

### Ownership and authentication

- Declare `Credentials` and immutable public account values in Common. Construct
  the store with its strong `Ensemble` parent; parent add/remove operations update
  both ends. Destroying the root joins Agents before destroying shared credentials.
  Destroy cancels and joins auth operations and detaches; it does **not** log out.
- Use only an explicitly selected private directory. Never import another tool's
  account, inspect its credential files, or place OAuth state in History, skills,
  logs, ordinary configuration or browser snapshots.
- Implement authorization-code PKCE with random state, nonce and verifier, exact
  callback validation, and signed identity validation. A returned client ID is a
  registration, not proof of the user's identity. Save it before consuming the
  one-use code; only a verified ID token can bind the subject.
- Share one protected, versioned file and cross-process lock. Reread under the
  lock; serialize refresh through durable replacement. Refuse damaged data rather
  than interpreting it as an empty store. Keep accounts and registrations separate.
- Renew near expiry and allow one same-funding-source recovery after inference
  HTTP 401. Reuse a newer durable grant if another owner already rotated it.
  Bound transient refresh retries; never automatically replay an uncertain
  authorization-code exchange. Clear local tokens on terminal grant rejection.
- Distinguish local logout from remote sign-out. Clear locally before attempting
  revocation; report whether the provider confirmed it. Preserve registration and
  saved plan selection either way.

### Transport, policy and presentation

- Make `Surface.Responses` and `BillingMode` explicit immutable Configuration
  choices. Plan mode requires OpenAI Responses, streaming and an account label,
  with no API key. Keep existing provider/surface defaults for other callers.
- Project tools, instructions and results from the existing vocabulary. Preserve
  exact call identity and compatible opaque reasoning; request encrypted reasoning
  for stateless replay. Unsupported media must refuse, not vanish.
- Parse JSON and SSE. Require `response.completed`; failed/incomplete terminal
  events, unfinished fallback items and EOF without success cannot run tools.
  A nonempty terminal output wins. An empty one uses only completed streamed items.
- Preserve real arrival order for tool explanation. The exact supported
  `gpt-6-astra` capability row requires ordinary visible text before tools; a
  reasoning summary is not an explanation. Refuse tool-first calls with paired
  results before effects, even if later terminal text is reordered ahead of them.
- Keep reasoning effort and summary policy per Agent. The supported row defaults
  to `medium`/`auto`; efforts are `low`, `medium`, `high`, `xhigh`, `max`. Explicit
  `off` omits summary requests. Do not infer capabilities or prices from a name.
- Display returned summaries separately, stream them incrementally, and admit
  final-only suffixes once to the existing speech queue. Replay is silent. Browser
  projection must never expose encrypted reasoning. Expected plan-authorization
  refusal records one durable error for narration, not duplicate turn-end speech.
- Discover visible plan models through the credential owner, retaining server
  order. A catalog entry establishes neither renderer support nor price. CLI model
  flags/environment bypass discovery; the GUI catalog is read-only, not a picker.
- Announce model and funding source before inference. Successful login atomically
  saves the future plan choice; explicit launch overrides do not rewrite it.
  Never change a running Agent's model/account/billing after account management.
- Keep plan token counts, but do not apply API-dollar estimates to them. Use only
  supported cache fields: this endpoint supports an optional OpenAI
  `prompt_cache_key`, not the historical explicit-cache extensions discussed in
  earlier development records. Preserve Anthropic/Gemini behavior.

Test real owners, files, sockets, processes and host lifecycles with working
boundary fakes. Keep local component evidence, executable evidence, live provider
proof and audible/human-accessibility evaluation distinct.

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
the cuts in Chapter 14, the later memory in Chapter 21, and Chapter 22's recorded
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

There is no `AuthProvider` interface, copied service bag or host-held bearer cache.
The concrete path is:

```text
Engine → Agent → Ensemble → Credentials
```

`Common/Auth.cs`, `AuthProtocol.cs` and `AuthLogin.cs` declare the shared types,
persistent fields and partial operations. Their implementations live in `Auth` in
the same assembly. For example, the public declarations include:

```csharp
public partial Credentials(Ensemble parent, string directory);
public partial Task<string?> SelectedPlanAccountAsync(
    CancellationToken cancellation = default);
public partial Task<string> GetBearerTokenAsync(
    string label, CancellationToken cancellation = default);
public partial void Destroy();
```

These are declarations from the partial owner, not a separate interface to inject
into Engines. `new Credentials(app, directory)` registers with the real parent.
`Ensemble.RemoveCredentials` clears the reverse link itself. Public `AuthAccount`
and `PlanModel` records are explicitly value-like snapshots; they contain no
access, refresh or ID tokens and need no holder back-pointers.

`Engine.SendAsync` renders without credentials, then resolves the selected bearer
through the parent path immediately before HTTP. Token acquisition failure has a
fixed safe error. An actual plan HTTP 401 permits exactly one retry of the same
request bytes. The internal rejected-bearer operation compares that exact bearer
under the file lock, so a delayed 401 cannot unnecessarily rotate a newer grant.
API-key requests never enter this path. Other statuses and interrupted successful
streams do not trigger an inference retry.

The Engine still owns wire exchange and its process meter; the actor still owns
durable admission, tool effects and History. A successful retry yields one
completed response and one meter update, not two conversational turns.

### One protected store, not copied account state

The C# implementation uses `credentials.json` beside `credentials.lock` in the
explicit directory. It requires private Unix permissions and refuses unsupported
storage-policy platforms. Version 2 stores a nullable `selected_plan_account`,
installation identity, and account registrations/grants. Version 1 remains
readable as having no saved selection. A read does not silently migrate the file.

The decoder validates the entire candidate, including unique labels, registration
identity, dates, token completeness and selected-label membership. A damaged file
is an error, not a successful empty login. The local operation gate coordinates
one owner; the file lock coordinates processes. Each writer rereads the file under
that lock, writes a private temporary file, then replaces the primary before
publishing success. This is atomic replacement, not a hardware power-loss/fsync
guarantee.

The file is the authoritative credential and selection state. The owner retains
only an active login attempt and lifecycle resources, not a second token cache.
Refresh, registration updates and logout preserve the selected label; successful
login changes it in the same write as the validated grant. Matching email strings
do not authorize registration sharing.

Credentials are private, but they are not hidden from the person operating the
application. The user controls the store and may clear it. Ordinary snapshots and
saved Agent configuration deliberately omit it.

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
`/auth/callback`." Only the port may vary. For the duration of login, the host becomes a small callback server. A
tool-calling application that reads files and runs processes now also listens
on a socket and serves the browser callback.

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

The C# library returns the private authorization URL; it does not launch a
browser. CLI `auth login` binds an owned listener first, prints the URL in the
dedicated local command, and waits with bounded cancellation and cleanup. The
GUI uses a user-opened separate sign-in window. The URL never enters the main
DOM, WebSocket transcript, or model context: it may contain an ID-token hint.
A blocked popup fails before login starts.

An unsolicited wrong-state callback does not consume the active attempt. A
matched denial or failed exchange does. Callback pages contain no codes, state,
URL, or account information. The listener may therefore serve more than one
request before the valid callback finishes the attempt; it is not a one-request
server. Hosts ask the owner whether the attempt remains pending instead of
copying its mutable state.

Persist the issued client ID before exchanging the one-use code. Failed exchange
leaves a pending registration with no subject or tokens, ready for a later login.
Only the signed, validated identity binds the subject. Do not automatically retry
an uncertain code exchange. A changed registration from another process
invalidates the pending attempt. There is no implemented `/complete` headless
redirect; use the documented local loopback flow rather than inventing one.


## Identity is not permission

A successful login proves who the user is. It does not prove the
user has granted plan usage. Two separate checks are required, and
both must pass before inference.

The token exchange returns three tokens: an access token for API
calls, a refresh token for renewal, and an ID token for identity.
The ID token is a signed JWT. Validating it requires fetching
OpenAI's published JWKS (JSON Web Key Set), verifying the
signature against those keys, and then checking the identity claims:
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


The C# validator is deliberately narrow RS256, using .NET RSA/SHA-256 rather
than a general JWT framework. It rejects ambiguous JSON/key selection and checks
audience, authorized party, timestamps, nonce and subject continuity; a supplied
access-token hash must also match. A decoded JWT is not a verified identity.
The tests sign genuine local RSA tokens and exercise refusals, but do not prove
that a real account can complete this flow.


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

The fix is mechanical: serialize credential updates. The C# owner serializes
local operations across its shared store; two processes using that file also
coordinate through its file lock.
The owner locks and reloads the store before deciding
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

Network errors and HTTP 5xx are transient. The C# refresh path makes
at most three attempts with bounded exponential backoff. The critical distinction: a terminal error
is the server saying the credential is dead. A transient exchange
failure does not establish that: a reachable server can return a
temporary error too. Retain the stored credentials. A 401 on
inference alone does not establish revoked refresh credentials;
this implementation attempts renewal once through the same funding
source, and clears tokens only after a terminal refresh rejection.
([Error recovery](https://developers.openai.com/siwc/token-sharing-open-source/errors-and-recovery).)


After an inference HTTP 401, Engine asks Credentials about the exact rejected
bearer. Under the existing lock, the owner reuses a newer valid grant if another
process already rotated it, or refreshes the rejected token. Engine retries the
same request body once. A second 401 ends the exchange. API-key requests, quota
errors, and failed or interrupted streams do not enter this recovery path. Only
the successful completed response contributes to the Engine meter.

`LogoutAsync` clears local tokens and retains registration, host identity and
saved plan selection. `SignOutAsync` clears locally **before** attempting remote
revocation, then reports whether revocation was confirmed. A network failure
cannot undo local logout. Destroying Credentials is different again: cancel and
join operations, dispose HTTP resources and detach from Ensemble, without
logging the account out.


## The Responses API

Chat Completions, Responses and funding are separate choices. Existing OpenAI
callers keep Chat Completions unless they select Responses or an explicitly
Responses-only model row. API-key mode can use Responses too. Plan mode validates
OpenAI + Responses + streaming + nonempty account, rejects an API key, and uses a
redirect-disabled client with HTTPS endpoints except literal loopback HTTP tests.

The exact `gpt-6-astra` row is explicit capability data, not a model-name pattern.
It selects Responses, streaming, application reasoning defaults `medium`/`auto`,
and ordinary explanation before tools. It does not establish account availability,
media support or an API price. `--reasoning-effort` and `--reasoning-summary`
override their `OPENAI_REASONING_*` environment values; each Agent freezes its own
Configuration. Unsupported explicit values refuse before HTTP. `off` omits the
summary request. Other models do not inherit these settings by guesswork.

References: [model documentation](https://developers.openai.com/api/docs/models/gpt-6-astra)
and [reasoning guide](https://developers.openai.com/api/docs/guides/reasoning).

### Projection uses the existing vocabulary

`Engine/Responses.cs` projects text, tool calls, tool results and compatible opaque
items from Context. Shared final rendering combines primary and skill instructions
once; it does not create a second manual store. Human text is a user message,
ordinary assistant text is an assistant message, and system-like input uses the
developer role. The top-level `instructions` field is also supported; the current
plan limitations allow instructions **or** developer messages, contrary to an
older grader's blanket prohibition.

For plan tools, the final projection uses a named namespace:

```json
{
  "model": "gpt-6-astra",
  "instructions": "Explain tool actions before calling them.",
  "input": [
    {"type": "message", "role": "user", "content": [
      {"type": "input_text", "text": "Inspect the requested file."}
    ]}
  ],
  "tools": [{
    "type": "namespace",
    "name": "ensemble",
    "description": "Ensemble tools",
    "tools": [{
      "type": "function",
      "name": "read_file",
      "description": "Read text from a file",
      "parameters": {
        "type": "object",
        "properties": {"path": {"type": "string"}},
        "required": ["path"]
      }
    }]
  }],
  "include": ["reasoning.encrypted_content"],
  "reasoning": {"effort": "medium", "summary": "auto"},
  "store": false,
  "stream": true
}
```

This illustrates the shape, not the complete installed tool catalog. A returned
function call retains its native item and call ID. A `function_call_output` uses
that same call ID; the actor's ordinary result becomes its output text. Compatible
opaque reasoning replays with exact surface/model provenance. It is neither
assistant prose nor a tool result. A billing switch adjusts the disposable tool
namespace rather than mutating History. Blob/media projection currently refuses
explicitly; documented provider media support does not make our renderer support it.

### Streaming progress is not execution authority

The SSE reader tracks output indices, completed items, displayed text and stable
part IDs in request-local assembly state. Deltas can update the page immediately.
Only a successful `response.completed` can finalize a response. When its output
is nonempty it is authoritative; when empty, only completed item-done values may
supply the fallback. Added skeletons and unfinished argument deltas are not facts.
Failed/incomplete terminal events and missing terminal success refuse tools.

Explanation policy uses actual SSE arrival order, not reordered terminal output.
A completed visible message can explain a subsequent tool even without deltas;
a summary cannot. Tool-first responses get paired refusals through the normal
actor, with no `ToolStarted` or effect. A later model request can consume those
refusals normally.

The GUI projects `summary_text` to distinct collapsed **Reasoning summary** cards.
Encrypted reasoning remains only in canonical compatible replay data. The existing
per-part speech cursor admits incremental text and final-only suffixes once;
reconnect/replay is silent. These are public summaries, not a private reasoning
trace and not a promise that the provider always returns a summary.

### Keep usage and price separate

Responses usage maps cached input into the existing disjoint count buckets. Missing
or malformed required counts remain unknown; an omitted optional cache count is
zero. The Engine meter distinguishes billing mode, so plan usage never receives
an API-dollar estimate. Unknown models/tiers keep valid counts without invented
rates. Historical preview price tables are not production pricing authority.

The supported optional OpenAI cache-routing field is `prompt_cache_key`. Historical
`cache_control`, `prompt_cache_retain` and `prompt_cache_min_tokens` experiments are
not implemented as Responses guarantees. The executable exercise checks configured
API routing and default plan omission; it does not infer a provider prohibition
or live cache hit from that difference.

### The stream that had already finished its tool call

In the historical implementation, after removing an unsupported cache option,
the subscription
request reached the model and still failed. The body contained SSE,
but the parser selected its whole-JSON branch. Fixing that decision
exposed the second assumption: a completed `read_file` item had
already arrived in `response.output_item.done`, while the terminal
`response.completed` carried an empty `output` array.

The parser needs both parts of the protocol. Detect the stream
without waiting to buffer its entire body, retain completed items
by output index, and use them when terminal output is empty.
A nonempty terminal output remains authoritative. In either case,
terminal success must arrive before those items become conversation facts or
executable tool calls. Usage may remain unknown; it is not an execution gate. The fact that a tool's
arguments have streamed is no permission to execute a response
that subsequently fails.

That captured stream made both defects reproducible without another
model call. In the C# run, a headerless loopback stream separately exposed the
same dispatch assumption; the scoped executable grade verifies the correction
without another provider call. A parser test that always repeats every completed item
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
observation interface. Summary deltas have a distinct kind in that existing
path, separate from assistant-text deltas. They are delivered
incrementally, kept separate from the assistant's visible text,
and preserved when a tool call is adjacent. A response that
interleaves reasoning, tool calls, and assistant text must emit
every event in stream order. The existing part identities keep the streams associated with their
finals. The GUI can feed summary phrases to the speech queue while
the response continues, without merging them into the answer or
speaking the encrypted item that accompanies them.


The selected C# row supports `auto` and explicit `off`, not the historical
`detailed`/`concise` choices above. A final may include text beyond its delivered
deltas: admit that suffix once through the existing per-part speech cursor.
Replay remains silent. Local transcript tests prove queue admission, not hearing.
Reasoning summaries do not satisfy the selected model's ordinary pre-tool
explanation requirement; that is Ensemble admission policy, not a provider promise.


## Funding selection and account controls

### Funding is consent, not fallback order

The implementation has two funding modes: `api` and `chatgpt`. It deliberately
has no “try plan, then spend API credits” mode. An expired grant, denied permission,
quota refusal, server error or network failure does not change the chosen payer.
A plan HTTP 401 can renew and retry once against the same source; that is not
payment fallback.

Only an explicitly supplied `--credentials-dir` / `EN_CREDENTIALS_DIR` is consulted.
Successful login saves the selected plan label atomically with the grant. On a
later launch, absent billing selection uses that saved choice; absent account
selection uses its label. Plan launches default to OpenAI Responses only when those
choices were not explicitly supplied. An incompatible explicit choice refuses.
An explicit `--billing api` or `EN_BILLING=api` overrides the launch without rewriting
the saved plan choice. Flags override their environment settings.

Logout leaves plan selected. A signed-out next launch therefore refuses before
inference instead of discovering an ambient API key and charging it. A damaged
selection file also refuses. Identity-only login saves the choice too, but does
not grant inference permission. A running GUI Agent never changes its immutable
funding, model or account after account-management operations.

Before inference, CLI stderr and the GUI's visible funding banner identify the
model and API metering or selected plan label. The GUI uses its existing usage
frame, so reconnect shows the same running configuration. This description is
not a token, provider instruction or new accounting store.

### Human commands, not model tools

From the C# endpoint directory, the standalone commands are:

```sh
dotnet run --project Cli/Ch04.csproj -- --credentials-dir /private/path/auth auth status
dotnet run --project Cli/Ch04.csproj -- --credentials-dir /private/path/auth --account work auth models
dotnet run --project Cli/Ch04.csproj -- --credentials-dir /private/path/auth --account work auth login
dotnet run --project Cli/Ch04.csproj -- --credentials-dir /private/path/auth --account work auth logout
dotnet run --project Cli/Ch04.csproj -- --credentials-dir /private/path/auth --account work auth signout
```

Choose your own private path. These are human operations, not commands to run
without account authorization. They construct no Agent, skill loader, conversation
or inference request. `status` returns redacted metadata. `login` prints a private
URL and waits for its owned loopback callback; it does not launch a browser.
`logout` means local clearing; `signout` additionally attempts provider revocation.

The GUI's account panel offers the same concepts through guarded HTTP operations,
not WebSocket conversation messages or model-callable tools. It opens a separate
sign-in window during the user's gesture, before awaiting the URL. A blocked popup
starts no login. The main DOM never receives an authorization link, and one window
closing must not cancel another client's attempt. Fixed-safe failure messages avoid
reflecting callback or issuer data.

The model catalog is fresh account-specific discovery. `ModelsAsync` uses the
owner's protected client and current bearer, selects only `visibility: list`,
retains server order, and rejects ambiguous visible entries. Hidden entries are
not choices. The GUI renders names as text and discards a held result if its label
has changed; it does not select a model. Catalog availability, renderer capability,
application explanation policy, permission and price are five different facts.

### Narration is part of error handling

With narration enabled, public summary text is admitted incrementally under the
label “Reasoning summary.” Final-only suffixes are caught up once, and replay is
silent. Credential acquisition refusal produces one durable `error_occurred`;
the existing narrator speaks its content once, possibly as several labeled
sentences. `turn_ended.error` is not another narration source. This keeps an
unattended visible failure from becoming unexplained silence without leaking
credentials or inventing a model reply.

## Exercise, graded

Use a C# 14/.NET 10 endpoint and the existing executable grader. Run it against
an owned GUI process, private directories and loopback peer. Do not bind the user's
working port or reuse a connected browser profile. No real credential belongs in
the fake fixture.

### Know the scope before reading the score

The original Go Chapter 16 route is retained. The C# adapter is explicitly
`-executable-only`; it does not add an issuer override to the production host
merely to grade login. It seeds synthetic protected credentials for transport
checks and excludes the original OAuth-flow and token-refresh checks, 15 points
each. The remaining weights are unchanged:

| Executable check | Points | Evidence required |
|---|---:|---|
| No auth leakage | 15 | Synthetic access/refresh/ID tokens absent from model input, socket messages and scanned artifacts; encrypted replay absent from browser messages |
| Responses format | 15 | Selected bearer, streaming/stateless request, namespaced plan tools, ordinary explanation, real file bytes, exact call/result correlation and opaque reasoning replay |
| Summary stream | 15 | Separate summaries, final suffixes and a causal gate that releases the terminal only after incremental summary delivery |
| Plan limits | 15 | Quota/admission failures remain plan-funded, no API fallback; signed-out plan sends no inference request |
| Cache routing | 10 | Configured API `prompt_cache_key` preserved; default plan request omits it and historical invented cache fields |
| **Scoped total** | **70** | **Not a full 100-point OAuth grade** |

The fake's explained-tool scenario deliberately places ordinary text before the
tool. That satisfies this model's admission policy without weakening it for the
grader. Historical unexplained-tool scenarios remain available to the original
route. Likewise, the adapter accepts documented `instructions` rather than copying
an obsolete prohibition from the old fixture.

The recorded reference run achieved **70/70**, parsed `passed:true` and process exit
zero: seven clean host sessions, eight inference requests, and zero requests in
the signed-out case. Retained per-run observations include the actual requests,
socket messages, errors and exits. The gate verifies that the summary is consumed
before terminal success, not merely present somewhere in a final HTML page.

A representative repository-root invocation is:

```sh
go run ./cmd/grade-edition2 \
  -ch 16 \
  -executable-only \
  -gui /absolute/path/to/owned-gui-wrapper \
  -json \
  solutions/ch16
```

Build the GUI first. The wrapper must invoke the built `Gui.dll` with `dotnet`
and forward `"$@"`; do not put a build tool between the grader and the host lifecycle.
Use the project's controlled environment and existing dependencies. The grader
starts and joins the owned host; successful compilation alone is not an exercise.
Retain failed fixtures when diagnosing a failure, and inspect exit status as well
as the JSON report.

### Credential-owner evidence is separate

The reference component suite has 20 focused Auth groups using real protected
files, local RSA-signed issuer responses, actual loopback sockets and two processes.
It tests registration persistence before exchange, identity and callback refusals,
rotating refresh, failed-save nonpublication, local-first logout/revocation,
account separation, fresh discovery, saved selection and one-retry 401 recovery.
These are actual components with working boundary fakes, not mocks of internal
method calls. They establish local behavior, not production issuer compatibility.

Eight focused Responses groups cover JSON/SSE, terminal admission, provenance,
usage, summary delivery, explanation policy and the headerless stream correction.
The earlier full 127-group pass predates that last dispatch correction; the eight
focused groups, clean host builds and final executable grade cover the correction.
Do not silently relabel an earlier broad pass as testing a later edit.

Standalone CLI auth lifecycle exercises and actual isolated two-client GUI checks
supply further host evidence: callback guards, cleanup, read-only catalog, visible
funding before inference, silent replay, summary suffixes and narrow layout. They
do **not** establish successful host signed login. Successful signed exchanges
remain local library-issuer evidence. Speech-queue transcripts prove admission,
not human hearing or accessibility quality.

### Joint live exercise — October 29, 2026

Bill approved a dedicated private credential directory and authenticated in his own
browser. The CLI login exited successfully; a separate status command reported
signed-in plan access, and account-specific model discovery succeeded. Three
isolated CLI runs then used the visible `gpt-6-astra` model, explicit
`--billing chatgpt --surface responses`, streaming and `store:false` at the public
Responses endpoint. All three reached `response.completed` and answered `OK`.
No other application's credentials were inspected or borrowed.

Each response reported 3,641 input and 5 output tokens, with zero cached-read and
zero cache-write tokens. Chapter 23 records the controlled inputs and limits.
This establishes real CLI login, discovery and bounded plan inference in that
session—not live renewal, 401 recovery, logout/revocation, GUI login, audible
quality or human accessibility. No API-key fallback or dollar-charge claim follows
from successful plan inference. The live session deliberately left the account
signed in; exercising revocation would be a separate authorized action.

The later [Chapter 23](chapter-23.md) work completed the separate Codex-backed
C# route, including explicit CLI/GUI selection and no silent billing fallback.
Its three actual C# CLI responses measured two cache hits and one miss. Codex
alone renews that grant; this does not change the public OAuth ownership or the
limits of this chapter's live exercise.


## Taking it for a spin — historical evidence

The following is the retained earlier student account, not a C# live-provider
run. Its model names, measurements and failed shutdown belong to that run.
The C# implementation and local results above do not inherit its OAuth, model
availability, cache-hit or billing evidence.

The actual student API-key run used a two-worker precedence-scheduling
puzzle with GPT-6.1 Sol, high reasoning and detailed summaries. It read the
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


## The C# boundary

The same architectural lesson holds at the locally verified C# boundary:
credential acquisition and Responses encoding changed without replacing the
Agent scheduler, Jobs, canonical History or tool loop. Common owns the shared
vocabulary; Auth and Engine implement their respective operations. The GUI
projects public state rather than owning a second account store.

The preserved endpoint remains the original offline implementation boundary.
The later joint session on current main adds observed CLI login, account-model
discovery and three successful plan responses; it does not retroactively rerun
that endpoint's tests. Live renewal/revocation, GUI login, audible quality, human
accessibility and account billing remain unverified. A model name or an encrypted
artifact is not permission; a green fake-provider grade is not a provider bill.
