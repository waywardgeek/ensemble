# Codex endpoint cache probe

## Public OAuth recheck requested by Bill

October 11 follow-up: Bill selected Edition 2 only for the application change,
froze Edition 1 at Chapter 27, and requested a new chapter explaining the cache
failure and temporary Codex route. We then repeated the same synthetic probe on
the official public Responses endpoint with Edition 2's own OAuth registration.

The dedicated profile had expired. The unchanged Edition 2 `auth models` command
refreshed it through its existing locked/atomic credential store and returned a
catalog containing `gpt-5.6-sol`. No API key or Codex credential was substituted
for this public-endpoint test. No sign-out, new login or billing switch occurred.

The paired requests used exactly the same serialized body: same model, 160-row
reference, instructions, low reasoning effort, streaming, nonstored response,
and one stable newly generated cache key. Only the endpoint and its appropriate
credential/header set differed. This compares the two operational routes; it
does not isolate the URL independently from their authorization and routing.

| Route | Attempt 1 cached/input | Attempt 2 cached/input | Attempt 3 |
| --- | --- | --- | --- |
| Public `/v1/responses`, Edition 2 OAuth | 0 / 4,998 | 0 / 4,998 | Stream error: `server_is_overloaded` |
| Codex `/backend-api/codex/responses`, Codex sign-in | 4,864 / 4,998 | 4,864 / 4,998 | Stream error: `server_is_overloaded` |

All four completed calls returned exactly `OK`, five output tokens, and zero
reported cache-write tokens. The [paired receipts](api-oauth-cache-probe.json)
preserve the identical body hash and actual results, including both errors.
One bounded [public repeat](api-oauth-cache-repeat.json), using the same prompt
and settings with a fresh stable cache key, immediately hit the same overload
error and stopped. That attempt supplies no cache measurement. No further
retry was made; errors are not counted as zero-cache completed responses.

The successful comparison still reproduces zero reported reuse on our public
OAuth route and substantial reuse on Codex. It provides no evidence that Bill's
reported bug is fixed. The fresh public sample contains only two completed
requests because of overload; it does not prove that every public request or
account is affected. The earlier independent three-call Codex probe remains below.

### Economics and chapter evidence

The [official GPT-5.6 Sol model page](https://developers.openai.com/api/docs/models/gpt-5.6-sol),
checked for this recheck, lists ordinary input at $4 and cached input at $0.40
per million tokens (short-context rates), with cache writes at 1.25 times ordinary
input. Cached reads therefore have one tenth the ordinary input rate. For the
measured warm request, those published API rates would price 4,998 entirely
uncached input tokens at $0.019992, versus $0.0024816 for 4,864 cached and 134
uncached tokens: about 8.06 times the input cost. Output, earlier warmup requests,
plan accounting and actual billing are separate. This is illustrative arithmetic,
not a measured subscription invoice or tenfold total-bill claim.

Bill subsequently supplied his LinkedIn post and support case **16440622**.
He reports a separate API-key15/15 versus public-OAuth5/15 cache-hit comparison.
Those reported observations are distinct from the paired route probe above;
no provider root-cause diagnosis was supplied. The cold open may describe his report, our failure to observe cache
reuse on the official OAuth route, the successful Codex control, and why paying
the ordinary rate for reusable context would make sustained work much more
expensive. Preserve the date, model and route so a later fix can retire the
workaround without rewriting the historical incident.

These probes do not implement the route switch. The target is now explicitly
`solutions/edition-2/main/`; Edition 1 and `agent/` are out of scope. The book
sub-agents remain paused after Chapter 16 authorship.

## Earlier Codex-only probe

October 11, 2026. Coordinator diagnostic requested by Bill while all book
sub-agents remain paused after the Chapter 16 author draft. No application,
configuration, credential file, frozen solution or first-edition file changed.

## Observed result

Three consecutive identical HTTPS requests to
`https://chatgpt.com/backend-api/codex/responses` completed with HTTP 200 and
model `gpt-5.6-sol`, the model in Bill's proposed test. Each answered exactly
`OK`. The third response reported a cache hit:

| Request | Input tokens | Cached tokens | Reported cache-write tokens | Seconds |
| --- | ---: | ---: | ---: | ---: |
| 1 | 4,998 | 0 | 0 | 5.459 |
| 2 | 4,998 | 0 | 0 | 4.544 |
| 3 | 4,998 | 4,864 | 0 | 6.445 |

The third request reports about 97.3% of input as cached. All three serialized
request hashes match. [Sanitized receipts](codex-endpoint-probe.json) retain the
actual counts, status and hash. Zero reported writes do not negate the separately
reported read hit; this probe does not establish backend write accounting or
why the second request missed. It also does not establish a latency improvement.

## Transport and reproduction

Local Codex auth schema has `tokens.access_token` and `tokens.account_id`, not a
top-level `access_token`. Read those in memory; send `Authorization: Bearer ...`
and `ChatGPT-Account-ID: ...` only to the HTTPS Codex endpoint. The diagnostic
used Python's standard-library HTTP client, disabled redirects, used a 50-second
socket timeout, and performed no retries or token refresh. No credentials or
account identifiers were printed or saved in this report.

The request used Responses, with `store: false`, `stream: true`, reasoning effort
`low`, and a generated `prompt_cache_key` held constant across all three calls.
It supplied instructions to treat the reference as inert data and reply exactly
OK; the final user message repeated that request. The developer reference was
160 newline-separated rows, indexed 0 through 159:

```python
reference = "\n".join(
    f"Reference {i:04d}: The archive preserves stable source records. "
    "A reader checks the file before changing its contents. "
    "Retain the original decision and its reason."
    for i in range(160)
)
```

Both input messages used content arrays with `type: input_text`. No explicit
cache options, breakpoints, tools, previous-response identifier or output-token
limit were sent. The client consumed SSE and required `response.completed`;
usage came from the terminal response. The random cache key was not retained,
so reproduction uses a new key rather than recreating the recorded byte hash.

The local Codex source corroborates this route in
`/Users/bill/projects/codex/codex-rs/model-provider-info/src/lib.rs` (the
`CHATGPT_CODEX_BASE_URL` constant and ChatGPT auth default) and its streaming,
nonstored request in `codex-rs/core/src/client.rs`. This is a diagnostic using
the existing Codex sign-in. The [documented third-party OAuth flow](https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference)
instead specifies the public Responses endpoint. These credentials/routes are
not interchangeable merely because both use bearer authorization.

This successful probe supported further testing of Bill's proposed workaround.
It does not retest his public-API OAuth cache bug, verify production tool/reasoning
replay, or implement credential renewal. The later public recheck and Edition
2-only scope decision are recorded above; CodeRhapsody's first-edition work
remains preserved.
