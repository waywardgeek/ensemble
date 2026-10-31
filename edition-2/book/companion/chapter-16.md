# Chapter 16: Leaving Anthropic for OpenAI

A provider migration tests whether the agent owns its conversation or merely keeps a copy of one provider’s request. Ensemble already has a neutral history, provider renderers, and a transport boundary. Moving to OpenAI still requires a new renderer, different stream parsing, and credentials that can renew themselves. It should not require rebuilding the Agent or teaching every tool how authentication works.

The reason for moving is a conflict over context control. The agent deliberately removes stale material: ephemeral views, old tool output, and context selected for compaction. A provider may legitimately protect hidden reasoning while imposing rules that make those edits incompatible with retaining its reasoning artifacts. The choice is not between security and no security. It is whether the service’s continuation contract fits the agent’s context policy.

## Hidden reasoning has an authorization boundary

Encryption does not by itself establish who may replay a reasoning artifact. Its owner may be unable to read it while another model can decode it. In August 2026, researchers replayed a frontier model’s encrypted reasoning into a cheaper model behind a jailbreak prompt. The receiving model treated the block as its own prior thinking. The recovered material included credentials; 64 of the 704 recovered artifacts were absent from visible chat history. An apparently opaque conversation attachment had carried secrets outside the visible record. These are the researchers’ findings, not credentials to try. ([Panfilov et al.](https://arxiv.org/html/2608.09867v1), §§3.1 and 4.1.)

OpenAI’s September 30 report describes stronger protections across users, workspaces, organizations, and model families, a closed replay pathway, and output checks. It does not disclose enough implementation detail to identify the mitigation as exactly the paper’s proposed cryptographic design. Protecting replay means binding an artifact to an authorized context, not merely making its bytes unreadable to the person storing them.

Anthropic’s preserved-thinking policy creates a different constraint. The Opus 5.5 launch applied it to that model and Fable 5.1 for API accounts created on or after August 31, 2026. Documentation checked October 11 ties retained thinking to the prefix against which it was produced: editing earlier messages or tool results can invalidate later thinking. It also supports server-side editing and client-side compaction that discards stale thinking. The conflict is therefore not “compaction is forbidden.” It is retaining those artifacts while independently changing their earlier context.

Ensemble depends on that editing authority. A changing screen view, an ephemeral tool result, and a recall attachment retained until a checkpoint have different lifetimes, but the client must be able to stop sending each when its policy calls for it. Current Anthropic examples elsewhere in the edition stay on supported older models such as Sonnet 5; a newer catalog entry does not establish compatibility with that behavior. This chapter takes the migration route rather than redesigning context management around preserved thinking.

## Credentials belong beside transport, not inside rendering

A static API key is constant from this client’s perspective: resolve it at startup and send the same value with each request. The service can still expire or revoke it. An OAuth access token, by contrast, expires routinely, and renewing it rotates the refresh token. Treating either token as an ordinary configuration string puts renewal and concurrency failures into every caller.

The credential seam has two operations. `GetBearerToken` returns a valid bearer or an error; `Kind` distinguishes an API key from OAuth. The API-key provider wraps a string. The OAuth provider owns protected profile state and coordinates refresh. HTTP obtains the bearer just before sending and puts it in the same `Authorization` header in either case. After an authentication failure, HTTP can request renewal without taking ownership of stored tokens.

The owner relationships matter as much as the interface. Ensemble owns the auth store. Agent selects the funding mode. Engine obtains the bearer through request-local configuration. The renderer receives only public route capabilities, such as whether tools require a namespace. Turning history into request bytes must not require access to an account or trigger a refresh. That keeps saved-conversation rendering and replay offline, and lets a new renderer concentrate on its wire format rather than reimplementing OAuth.

## Signing in without a secret hidden in the binary

A distributed local application cannot keep a shared client secret: someone can inspect or copy the binary. Authorization Code with PKCE instead proves that the application completing authorization is the one that began it.

The client generates 32 cryptographically random bytes and base64url-encodes them without padding to make a verifier. It hashes the verifier string with SHA-256 and base64url-encodes the digest to make a challenge. The authorization request carries the challenge; the token exchange carries the verifier. Intercepting the authorization code is insufficient without the verifier.

The client briefly becomes a web server, binding a loopback listener on any available port at `127.0.0.1`, with the exact path `/auth/callback`. For this contract, `localhost` is not a substitute and `/callback` is not an equivalent path. Only the port varies. The browser returns the authorization result there, and the listener shuts down after receiving the callback.

PKCE does not replace the other checks. Generate fresh `state` and `nonce` values for each authorization attempt. The state must match the callback to prevent cross-site request forgery; the nonce must match the returned ID token to prevent replay. Persist a random `urn:uuid:<v4>` as the installation’s `ext_agent_host_id`, never deriving it from username, hostname, email, or MAC address. The RFC 8707 resource value, `https://api.openai.com/v1`, appears in both the authorization request and the code exchange.

First registration uses the bootstrap client ID `dynamic_agent_client`. The callback supplies the issued `client_id` alongside the code and state. Persist that issued ID before exchanging the code: the bootstrap value is never valid for token exchange. Subsequent requests use the issued ID. A returning callback may omit it, but a different supplied ID is an error, not a new identity to accept silently. This is callback-based registration, not a separate RFC 7591 registration request. Send `agent_name_hint=Ensemble` on first registration and omit it on reauthorization. Identity scopes, offline access, `resource.invoke`, and `chatgpt.tokens.use.direct` are protocol inputs rather than descriptive labels. ([Sign-in contract](https://developers.openai.com/siwc/token-sharing-open-source/sign-in).)

### Identity and permission to spend are separate results

The exchange returns an access token for API calls, a refresh token for renewal, and an ID token for identity. Decode-and-trust is not ID-token validation. Fetch the trusted issuer’s published JWKS, verify the JWT signature, and check that the issuer is exactly `https://auth.openai.com`, the audience is the issued client ID, the token is unexpired, and the nonce matches the authorization request. The `sub` claim supplies stable account identity. Keep it with the issued client ID; the same email address or person in a different workspace registration is not grounds for merging token sets. A returning verified identity must match the account being renewed before replacing its credentials.

A valid identity still does not authorize plan inference. Inspect the scopes actually granted in the token response for `chatgpt.tokens.use.direct`. Without it, identity may be saved, but inference remains disabled. Trying an available API key instead would turn missing permission into an unauthorized funding change.

The discovery document inspected during the original implementation advertised identity and offline-access scopes but not the plan-usage scope. Implementing only what that catalog listed could therefore produce a working login that could not infer. Discovery describes available protocol metadata; the granted scopes determine what this user authorized.

## Refresh is a coordinated state change

The documented lifetimes on October 11 were about an hour for an access token and thirty days for a refresh token. Those are not expiry constants to bake into the client. Track the returned expiry and refresh before the next inference when the current time reaches five minutes before `expires_at`.

Refresh rotation creates the dangerous race. Once the server issues a replacement refresh token, the old one is dead. If two callers independently notice expiry, the first can renew successfully while the second submits the old token. The server may treat that second attempt as replay and invalidate both tokens. A mutex protects goroutines in one process; it cannot protect two clients sharing a credential file.

The accepted store serializes refresh per profile across processes, then reloads the profile under that coordination before deciding to renew. The second caller can use the replacement already saved by the first. Save the new access and refresh tokens atomically before using either, through a transaction or temporary-file rename. Updating fields separately risks leaving a mixed, unusable profile.

Atomic replacement solves partial local writes, not the entire distributed transaction. If the server rotates the token and the client crashes before persisting the response, the old token is already invalid and the new one is lost. No local rename can recover a response that never reached durable storage.

Terminal refresh errors are `invalid_grant`, `invalid_refresh_token`, `token_expired`, `refresh_token_expired`, `refresh_token_invalidated`, and `refresh_token_reused`. They establish that the credentials are unusable. Discard the tokens, retain the issued client ID and host ID, and require sign-in again. Network failures and HTTP 5xx are transient: retain credentials and retry with exponential backoff. A reachable server can have a temporary failure too.

An inference 401 alone does not prove that refresh credentials are revoked. This implementation attempts renewal once through the selected funding source. Only a terminal refresh rejection warrants clearing the tokens. Authentication recovery must not become a hidden switch to another account or another bill.

## Responses changes encoding and the meaning of completion

Responses represents the same neutral conversation with input items rather than the previous message format. System-role messages become developer-role messages, and streaming event names change. HTTP 200 means the transport opened. `response.completed` establishes completed inference. Text appearing on the connection is not enough to make a failed or incomplete response part of the conversation.

The model table selects the surface per model through `SurfaceResponses`. Earlier models remain on Chat Completions; selected newer models use Responses. A boundary test guards against flipping an entire vendor onto a surface that has not been established for every model. Historical preparation for this migration included an unwired Responses surface, separate reasoning-summary events, and recognition that reasoning with function tools needed Responses. Those were useful seams, not an already-completed migration.

Model support and route permission are different facts. On the ChatGPT-plan HTTP route, `store: false` and `stream: true` are mandatory and `input` carries the required history. The route excludes sampling controls `temperature` and `top_p`; `max_output_tokens`; server-side continuation fields `previous_response_id` and `conversation`; `background`; and `metadata`, `max_tool_calls`, `moderation`, `multi_agent`, `prompt`, `prompt_cache_retention`, `safety_identifier`, `top_logprobs`, `truncation`, and `user`. Keep these restrictions in the selected route profile, not scattered through the message loop or inferred from the model name.

Local function tools need a namespace on this route. Preserve the returned namespace separately from the function name, then carry the call ID into `function_call_output`. Producing a tool call is only the first half of the transaction. The next request must contain the actual tool result correlated with that call.

### Cache support follows the route

The top-level `instructions` field is legal on the plan route. As a plain string it cannot carry its own content-block breakpoint; that does not make its bytes inherently uncacheable, because a later reusable prefix can include earlier instructions. This renderer places the constitution in a developer-role content block so the supported API-key route can put an explicit boundary there.

The earlier student run exposed a restriction missing from the original account. Its first subscription request failed with HTTP 400 and `invalid_parameter`, naming `prompt_cache_options`. Explicit mode had worked with the API key. The accepted plan profile therefore omits that option and its associated breakpoint markers. The option’s rejection was observed; omitting the paired markers is a conservative choice, not a separately measured rejection.

The supported API-key route retains `prompt_cache_options.mode` set to `explicit` and markers on eligible content blocks within the four-marker budget. GPT-6.1 Sol standard prices verified October 11, for prompts through 272K input tokens, were $2 per million fresh input tokens, $0.10 cached input, $2.50 cache writes, and $10 output. Reads cost 0.05 times fresh input and writes 1.25 times. Those prices explain why placement matters on that route. They are neither a subscription invoice nor evidence that any particular request received cached usage.

### A completed tool item can precede an empty terminal output

Removing the unsupported cache option did not make the subscription run work immediately. The body contained SSE, but the parser chose whole-JSON decoding. Once stream selection was repaired, another assumption failed: `response.output_item.done` had already delivered a completed `read_file` item, while `response.completed` carried an empty `output` array.

The parser must detect streaming without buffering the whole body, retain completed items by output index, and recover them when terminal output is empty. A nonempty terminal output remains authoritative. In both cases, terminal success and usage must arrive before retained items become conversation facts or executable calls. Complete-looking arguments do not authorize executing a response that later fails.

The captured stream reproduced both defects without another model call. A fake that always repeats completed items in its terminal array cannot reveal the lost-item bug. The real route was less redundant than the test fixture, and the parser had mistaken redundancy for a requirement.

## Public summaries need their own live path

Reasoning summaries are model-generated descriptions of private thinking, not the private trace. Preserve encrypted reasoning for compatible replay, but do not read or speak its ciphertext. Public-summary deltas belong in their own observer event kind, separate from assistant-text deltas, with existing part identities associating each stream with its final content. Interleaved reasoning, tools, and assistant text must retain stream order without losing summaries next to tool calls.

Request `reasoning.summary: "auto"` by default on supported models; an explicit supported choice such as `detailed` must not silently become every request’s setting. A model may also require ordinary assistant explanation before tools. Public summaries do not satisfy that rule. Select it from the actual model configuration, including a worker’s model override, put it in model-facing instructions, and retain the dispatch and arrival-order checks. Whether a summary happened to arrive is not a capability signal.

This is an accessibility requirement, not just an extra panel. Bill has no macular vision and reads code at about 80 words per minute; he listens to model output at about five times speaking rate. A summary delivered after the answer misses the opportunity to hear a wrong approach developing and send a corrective hint. The GUI needs phrases while generation continues, without inserting those phrases into the assistant’s answer.

The first-edition coder investigated streaming before implementation. The October 2 report recorded 314 deltas across three parts for `gpt-6.1-sol` with high effort and a `detailed` summary, 186 with `auto`, and four with `concise`. An initial trivial prompt had produced only 18 reasoning tokens and no summary deltas, nearly leading to a false capability conclusion. A probe needs a task that exercises the capability it is meant to observe.

Even substantial reasoning does not guarantee a public summary. The earlier student API-key session produced 90 summary deltas. Its subscription session requested `detailed` and used 165 reasoning tokens on the final response, yet produced none. Completion and shutdown cannot depend on hearing a summary.

The Edition 3 restart found a separate delivery gap: final summary cards survived, but the live path missed text supplied only in the final summary. Sending the unseen suffix through the existing public-summary path makes it available to display and speech without repeating the prefix. A complete card proves neither incremental arrival nor native speech delivery; those are separate boundaries.

## Funding selection must survive failure

The proposed billing state machine had three modes: subscription-only `chatgpt`, API-key-only `api`, and `chatgpt_then_api`. The third would require explicit `allow_metered_fallback`, defaulting to false, and an announcement before the first metered request. A paused agent is preferable to a bill the user did not authorize.

Only `subscription_sharing_usage_limit_exceeded` permits that proposed quota-driven transition. The limit can be app-specific rather than exhaustion of the whole subscription. `subscription_sharing_usage_unavailable` means temporary overload and calls for backoff, not another funding source. `subscription_sharing_unsupported_capability` requires request adaptation or a reported limitation. A 401 calls for credential and permission diagnosis with bounded same-source renewal. Treating these failures as interchangeable would make billing depend on unrelated transport or feature errors.

The accepted implementation chooses the smaller contract: it rejects `chatgpt_then_api` and provides no automatic fallback. Both supported modes announce their selected funding source before the first request. Successful login selects `chatgpt` for later launches unless explicitly overridden. Logout leaves that selection in place. Removing subscription credentials must not silently activate an API key already present in the environment.

Model discovery is account-specific too. `GET /v1/models` with OAuth returns the subscription catalog, filtered by `visibility: "list"`; use `slug` in requests and `display_name` in the UI. An API key sees its tier’s catalog instead. Availability there is necessary but not sufficient: the application still needs a supported model route. This integration uses the public `https://api.openai.com/v1/responses` endpoint, not another application’s private credential store or backend.

Login, status, discovery, inference, and logout must use the same profile selection. The client uses `EN_HOME` when supplied, otherwise its own `ensemble` directory under the operating system’s user-config location. Access tokens, refresh tokens, ID tokens, and API keys belong only in protected credential storage, never in prompts, memories, conversation saves, logs, debug output, or other model-visible state. Printing an authorization URL is not permission to print tokens. Logout attempts remote revocation, removes local tokens, and retains registration and host identity. If remote revocation cannot be confirmed, report that uncertainty rather than claiming it succeeded.

## What the recorded runs establish

The local exercise drives actual clients against external OAuth and Responses fakes. Signed JWTs, published fake JWKS, rotating tokens, and actual tool-result bytes test more than whether a request-shaped object can be constructed. A login that never reaches inference cannot establish a billing rule. Copying an error code into output does not establish that Agent treated the turn as failed. Both GUI and CLI login entrypoints should delegate to the same auth owner rather than acquire independent implementations.

The synthetic `gpt-ch19-course` model declares its fixture capabilities explicitly: Responses, SSE, summaries, encrypted continuation, and API-key explicit caching. It is not a live model, and its name is no basis for inferring capabilities of other OpenAI models. Local issuer, inference endpoint, and private-profile overrides keep these exercises separate from real browser consent and production accounts. Multi-turn continuation includes encrypted reasoning rather than testing only an isolated answer.

A local score samples the contract. Owner relationships, identity-validation cases, cross-process refresh coordination, and atomic replacement still need review. Browser checks of incremental speech add evidence that visible WebSocket events alone cannot supply. Neither fake summaries nor a passing score guarantees that a live model will emit a summary.

The historical paid task required reading a scheduling file, finding a minimum schedule for two identical workers, and proving optimality. The dependency chain A–C–F–H gives a lower bound of 13 from durations 4, 5, 3, and 1. That chain fits on one worker while the other performs B from 0–3, E from 3–7, D from 7–9, and G from 9–11. Matching a feasible schedule to the lower bound establishes optimality. Knowing that answer beforehand does not establish tool use: the actual file result must reach the following request.

The earlier API-key run did read the file and produce that schedule. Its three requests included the original prompt and a hint-triggered continuation, not two independent prompted turns. The hint reached the third request after an intermediate answer, and the process exited normally in 14.51 seconds.

The corrected subscription run made two requests with no API key configured. It read the file, carried the correlated result into the second request, and completed the same optimal answer in 14.887 seconds. Both terminal output arrays were empty, directly exercising completed-item recovery. But the launcher had tied its hint and eventual EOF to seeing a public summary. None arrived. The answer finished while the idle CLI remained open until the launcher killed it at 150.03 seconds. That is evidence of a successful answer and tool loop, not graceful shutdown or a saved conversation. The corrected completion condition was checked against captured events; no new paid run was made.

The cache lens also needed repair. Its extractor recognized the older API’s messages but omitted Responses `input`, so it compared the tools while ignoring the conversation. After including the actual input, the final plan request pair shared a predicted 92.12% byte prefix, yet both provider cache counters remained zero. Each input was below 1,024 tokens. Stable bytes, a completed task, and a granted cache hit remain distinct observations.

On the October 22, 2026 Edition 3 restart, local chapter and regression checks passed. Bill deferred the live OAuth exercise. There was no fresh native GUI, browser, or listening observation accompanying those results. The earlier paid runs remain historical evidence, not proof that the restarted implementation worked with a live account or provider.
