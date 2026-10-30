# Chapter 24: While the Cache Is Broken — Measure Before Taking the Detour

## Why this chapter exists

I wanted to give OpenAI credit. Letting people use their ChatGPT subscription
with somebody else's agent is exactly the right direction. Then I tested it.
Same model, same request bytes, same minute: API key, fifteen cache hits out
of fifteen; OAuth, five out of fifteen. I sent the reproduction to support,
case 16440622. The answer explained that caching was not guaranteed.

Frankly, that answer made me angry. A coding agent sends its working context
over and over. Getting the answer right while repeatedly charging the ordinary
input rate does not make the integration usable. I want to work on Ensemble,
not argue with support about whether the problem in the reproduction counts
as a problem. Fix the cache, OpenAI.

Hopefully this chapter is obsolete by the time you reach it. That hope is an
engineering requirement: test the current integration before implementing yesterday's
workaround. The measurements below explain why the earlier implementation took a
detour. They do not establish that today's public OAuth cache is still broken.

**C# execution status — October 29, 2026:** real Ensemble OAuth login and
account-model discovery succeeded. Three completed public-route requests reported
zero cached reads and writes. After the separately authorized comparison, the
Codex-backed C# path was implemented in core, CLI and GUI. Its bounded actual-CLI
check returned three correct answers with cache reads of **3,456, zero, and
3,456** out of 3,641 input tokens. That supports the optional detour, with a
measured miss—not a universal hit-rate or dollar-saving promise. Chapter16's
public grant and renewal store remain separate.

## TL;DR

**This chapter is optional and has no separate grader.** If the public OAuth
route gives satisfactory measured reuse for your workload, keep it. For the
failure observed here, Edition 3 now has an explicit Codex-backed C# path in
both CLI and GUI. This is a separate grant, not a way to send Ensemble's public
OAuth token to another endpoint.

1. Select `--billing codex` (or `EN_BILLING=codex`) and an explicitly approved
   `--codex-auth-file FILE` (or `EN_CODEX_AUTH_FILE`). There is no home-directory
   search, no copied token store, and no silent API-key or public-plan fallback.
2. One optional `CodexCredentials` child belongs to Ensemble. Engine reaches it
   through Agent → Ensemble immediately before sending. One bounded file read
   returns the access token and account together. Codex alone owns login and
   renewal; Ensemble neither refreshes nor writes that file.
3. Use OpenAI streaming Responses at the pinned
   `https://chatgpt.com/backend-api/codex/responses` endpoint. Reject incompatible
   configuration before reading credentials. Disable redirects and cookies.
   Send the paired bearer/account and a stable `session-id`; keep the normal
   Engine, actor, History, Jobs and usage accounting.
4. Send `store:false`, `stream:true`, top-level instructions and local input.
   A supplied `--prompt-cache-key` also supplies the session-affinity header;
   otherwise one Engine-lifetime key serves both. Neither is a hit guarantee.
5. Require terminal completion before treating a response as successful. Safe
   credential failure tells the user to repair the Codex sign-in; it does not
   expose the path, token, parser exception or provider body in History.
6. Verify boundaries with fakes, then make a bounded real application check.
   The actual C# CLI returned `OK` three times with identical bodies: cache
   reads were **3,456, zero, and 3,456** out of 3,641 input tokens. That proves
   integrated reuse occurred, not that every request reuses its prefix.

Work in `solutions/edition-3/main`; the source-only reader endpoint is
`solutions/edition-3/ch24`. This optional integration was finished after Gateway
and Goals, so that endpoint also includes their already completed foundation.
Earlier preserved endpoints were not rewritten. The source map below follows
Common declarations and real parent relationships, not a second transport agent.

### What we actually measured

On October 29, 2026 Bill authorized a new private `~/.en/oauth` store and completed
Ensemble's real public login. Redacted status confirmed plan access and model
discovery listed `gpt-6-astra`. Three CLI requests on the documented public
Responses route completed normally, each reporting zero cache reads and writes.

Bill then explicitly approved read-only use of `~/.codex/auth.json` for a bounded
comparison. The temporary Python probe sent the same body to Codex. Its first
attempt returned HTTP200 without a `Content-Type`; the probe refused before
reading the stream. **That completion and usage are unknown, not a cache miss.**
After correcting the probe, its two remaining attempts both reported 3,456 cache
reads. No replacement fourth request was sent. That experiment justified the
integration; it did not itself implement or test the C# path.

The later integrated check used the **actual C# CLI and Engine**, with explicit
Codex funding and the approved read-only source. Its existing Responses parser
already accepted the absent-header stream shape. All three requests returned
`OK`, and the second was a genuine measured cache miss:

| Route and implementation | Input per response | Cached reads | Output per response | Cache writes |
| --- | ---: | --- | ---: | ---: |
| Public OAuth, C# CLI, three completed requests | 3,641 | 0 / 0 / 0 | 5 | 0 each |
| Codex, temporary probe, two measured responses | 3,641 | 3,456 / 3,456 | 5 | 0 each |
| Codex, integrated C# CLI, three completed requests | 3,641 | 3,456 / 0 / 3,456 | 5 | 0 each |

The first probe attempt is additional, unmeasured traffic—not included in the
second row's token totals. Each three-response C# series used 10,938 total
reported tokens. In the Codex C# series, 6,912 were cached input, 4,011 were
uncached input, and 15 were output. Two responses reused about 94.9% of input;
one reused none. These are provider usage reports, **not a subscription invoice
or a measurement of dollars saved**.

The wire payloads in all three comparisons were byte-identical: 21,438 UTF-8
bytes, SHA-256
`95eae405d65743e2b34b6cb3e03754ffea92e3b5c465a5bd95ed0cd8b1f6e756`.
A documentary `request.json` has a trailing newline; the equality check uses the
actual trace payload, not that file's bytes. The prompt was fixed harmless text,
reasoning was low, summaries off, recall and memory disabled, and no tool was
called. The body already used top-level instructions: no cross-route message
conversion was necessary. The credential source, endpoint and HTTP headers
changed. The C# process durations were roughly 4.0, 3.6 and 2.9 seconds, including
startup—not a controlled latency comparison or time-to-first-token measurement.

The public [inference guidance](https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference)
checked for this exercise specifies `https://api.openai.com/v1/responses`,
`store:false`, streaming and terminal completion, and explicitly excludes
ChatGPT backend endpoints **for the public grant**. The
[Codex authentication guidance](https://developers.openai.com/codex/auth) and
[current upstream provider source](https://github.com/openai/codex/blob/main/codex-rs/model-provider-info/src/lib.rs)
describe a separate access path. Neither account was logged out or renewed by
the Codex comparison. This sample does not identify a service-side cause,
establish a universal hit rate, or explain an internal product decision.

## Implementation status

The core, both host selectors, local failure checks, isolated GUI display, and
bounded live C# path are implemented and verified. Six Codex-specific groups,
20 public-auth groups, eight existing Responses groups, eight host checks, ten
public-funding regression scenarios, and the final 310-group shared suite provide
different evidence scopes; several overlap and must not be added into a claimed
unique total. The shared build and both actual hosts build cleanly. There is no
separate Chapter24 score. Live GUI inference, renewal/revocation, accessibility,
subscription billing and general provider reliability are not established by
these checks. See the [progress record](../../docs/edition-3-csharp-progress.md)
for retained failures, exact local evidence paths and final preservation status.

## 24.1 In plain language

A coding agent may send thirty thousand tokens of instructions and working
context on every turn. Most of that context stays the same while the agent
reads a file, makes an edit, runs a test and examines its failure. Prompt
caching lets the provider reuse the unchanged beginning of the request.
Ensemble still sends the context; the response tells it how much was reused.

The answer can be perfect while that reuse fails. Nothing on the screen has
to look broken. An agent can finish the task and consume substantially more
input budget than expected, which is why the cache meter matters independently
of the answer.

For a historical price illustration, the GPT-5.6 Sol rates recorded on October 11,
2026 are $4 per million ordinary input tokens and $0.40 per million cached
input tokens. Fifty repetitions of a 30,000-token prefix amount to 1.5 million
input tokens: $6 at the ordinary rate or $0.60 if all those tokens received
the cached rate. In practice the first request needs warming and changing
input still costs separately. These are API-rate illustrations, not a
measurement of a ChatGPT subscription invoice. The tenfold difference applies
to the affected input tokens. These recorded rates are not a current quotation or
a complete model of newer cache-write charges. [Model pricing](https://developers.openai.com/api/docs/models/gpt-5.6-sol).

The fifteen-request comparison in the opener is Bill's reported experiment.
A paired probe in that earlier implementation on October 11 used the same GPT-5.6 Sol request body
with Ensemble's public OAuth credential and with Codex's credential at its
own endpoint. Both completed public requests reported zero cached tokens out
of 4,998. Both completed Codex requests reported 4,864 cached tokens out of
4,998, about 97.3%. A third request on each route failed with an overload
error; those failures supplied no cache measurement. The [probe record](../../docs/codex-endpoint-probe.md)
contains the receipts and reproduction details. That retest gave no evidence
that the problem had been fixed then; it says nothing about a later service revision.

Changing the address alone is insufficient. Ensemble's public OAuth grant
and Codex's sign-in belong to different access paths. The documented third-party
flow uses the public Responses endpoint. The earlier workaround used a separate
Codex sign-in with Codex's endpoint; it did not send Ensemble's public grant there.
That distinction is a security boundary, not a URL substitution trick.
[Public OAuth inference](https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference).

The useful ownership boundary already exists in C#. Credentials belongs to
Ensemble; Engine reaches it through its Agent immediately before sending. The
Agent owns scheduling and conversation, History records the durable facts, and
Jobs owns tool work. A justified transport change should not introduce a second
agent loop or duplicate the credential store. Explicit funding remains explicit:
a failed plan request never silently falls back to an API key.

## 24.2 Taking the detour in C#

Keep two credential owners distinct. Chapter16's `Credentials` owns Ensemble's
public registration, protected storage, selected account and renewal protocol.
`CodexCredentials` is a separate optional child of the same Ensemble root. It
selects exactly one path; construction performs no discovery or credential I/O.
Its parent add/remove operations maintain both ends under the root's existing
credential relationship gate. Destruction cancels and joins reads, then detaches;
it does not delete or rewrite somebody else's sign-in.

A request walks Engine → Agent → Ensemble → CodexCredentials. The reader accepts
one JSON object of at most 1 MiB and rejects duplicate properties, malformed data
and missing/unsafe header values. It reads `tokens.access_token` and
`tokens.account_id` together from that one document. An omitted legacy
`auth_mode` is accepted; a supplied mode must be `chatgpt`. It does not reconstruct
identity from a JWT or use `OPENAI_API_KEY` as a substitute. The exchange-local
pair is value-like, not a second mutable token store or relationship object.

This is deliberately inconvenient: Codex must repair an expired or rejected
sign-in. Ensemble reads again on each send, so a renewed pair is picked up
without restart. A Codex HTTP401 produces a safe failure and **no renewal or
second request**. Chapter16's public-plan one-retry renewal remains specific to
that different grant. Quota, redirect and interrupted-stream failures likewise
do not choose a different payer. API keys inherited in the environment are
excluded when either plan route is explicitly selected.

Configuration appends `BillingMode.Codex` without renumbering existing values.
It requires OpenAI, Responses, streaming, no API key and no public account label.
The only accepted endpoints are the exact pinned service base or its Responses
endpoint. An explicit public/loopback/other URL is refused; there is no test-only
production endpoint exception. Tests substitute a boundary HTTP handler while
exercising the real Engine construction, headers, JSON, parser and accounting.
The normal client disables redirects and cookies for both plan routes.

The HTTP exchange sends bearer authorization, `ChatGPT-Account-ID` and
`session-id`. A configured prompt-cache key is frozen in the request and used by
the header. Otherwise Engine creates one key for its lifetime; restarting the
Engine does not promise the same fallback identity. Top-level instructions use
the existing Responses projection and namespaced tool vocabulary. The credential
file path, account and bearer are not copied into documentary configuration or
History. Exchange diagnostics mask the paired secrets. Successful provider text
is still canonical content; masking diagnostics is not a claim to detect every
possible secret an application might put in a prompt.

### CLI and GUI selection

Select a source explicitly **only with the account holder's permission**. These
commands illustrate an approved file-backed Codex sign-in; do not assume that a
file exists or that another application's private store is yours to read. Perform
login or renewal with Codex itself, following its current documentation. Do not
paste credentials into chat, a shell argument, or the book's exercises.

From the repository root, build the actual host:

```sh
dotnet build solutions/edition-3/main/Cli/Ch04.csproj
```

Then, after approving that exact credential path, launch a fresh owned workspace:

```sh
work=$(mktemp -d)
env -u LLM_BASE_URL -u EN_ACCOUNT -u LLM_API_KEY \
  LLM_VENDOR=openai LLM_MODEL=gpt-6-astra \
  dotnet solutions/edition-3/main/Cli/bin/Debug/net10.0/ch04.dll \
  --billing codex --codex-auth-file "$HOME/.codex/auth.json" \
  --data-dir "$work" chat
```

`gpt-6-astra` was verified in this exercise; the name is not a promise of future
availability or entitlement. Explicit model IDs do not require a discovery
request. The chapter's ordinary model policy and capability checks still apply.
For an isolated cache exercise, additionally use a minimal primary skill and
disable recall/memory in that temporary Agent's settings so they do not add
requests or alter the prefix. Chapter16's `auth models` lists the public account's
models, not a new Codex entitlement service.

`EN_BILLING=codex` and `EN_CODEX_AUTH_FILE=/approved/file` are the equivalent
environment selections. Explicit flags win. Supplying only a file path does
**not** select Codex funding. There is no `CODEX_HOME` or `~/.codex` search and
no persisted Codex funding choice. The CLI announces the selected funding before
inference; GUI shows the same source in its normal funding display.

The GUI uses the same selectors and pinned default:

```sh
dotnet build solutions/edition-3/main/Examples/Gui/Gui.csproj
work=$(mktemp -d)
env -u LLM_BASE_URL -u EN_ACCOUNT -u LLM_API_KEY \
  LLM_VENDOR=openai LLM_MODEL=gpt-6-astra \
  dotnet solutions/edition-3/main/Examples/Gui/bin/Debug/net10.0/Gui.dll \
  --billing codex --codex-auth-file "$HOME/.codex/auth.json" \
  --data-dir "$work" --port 8099
```

Choose an unused owned port; do not replace a running application. In Codex mode,
public account controls are disabled and explain that Codex owns renewal. Guarded
server routes refuse those operations too: disabling a button is not a security
boundary. Explicit Codex startup skips the public credential store entirely,
even if the process inherited an unrelated `EN_CREDENTIALS_DIR`. It neither reads
that store's saved selection nor mutates the running actor when login state changes.
The existing CLI `auth` commands remain public-OAuth commands, not Codex operations.

To leave the detour, explicitly launch with `--billing chatgpt`, your public
`--credentials-dir` and `--account`, and no Codex base-URL override. That keeps
Chapter16's public flow and store. Select `--billing api` only when you intend
API-key funding. Neither direction copies credentials between applications.

### Source map

| Responsibility | C# source under `solutions/edition-3/main` |
| --- | --- |
| Shared credential owner, pair and safe failure | `Common/Codex.cs` |
| Fresh reads, endpoint validation, relationships and cleanup | `Auth/Codex.cs` |
| Billing vocabulary and configuration/funding contract | `Common/Values.cs`, `Serialization/Values.cs` |
| HTTP ownership, paired headers and retry separation | `Common/Owners.cs`, `Engine/Engine.cs` |
| Responses projection and terminal parsing | `Engine/Render.cs`, `Engine/Responses.cs`, `Engine/Streaming.cs` |
| Safe actor narration and root teardown | `Agent/Actor.cs`, `Application/Ensemble.cs` |
| Human selection and account display | `Cli/Program.cs`, `Examples/Gui/Program.cs`, `Gui/Server/Auth.cs`, `Gui/wwwroot/auth.js` |
| Real components with a boundary handler fake | `Tests/CodexTests.cs` |

There is no new model scheduler, History owner, usage meter or tool loop. The
request remains a frozen value passed to the existing Engine. The source split
makes the shared contract inspectable in Common and the implementation inspectable
in its responsible module, rather than hiding policy in UI glue.

## 24.3 Taking a measurement in C#

Do this with the account holder, not as an unattended credential-discovery task.
Chapter 16 supplies the human `auth status`, `auth login` and `auth models`
commands and their private-directory setup. Use those commands to establish the
approved account and discover a model actually visible to it. A model name in an
old table is neither current availability nor permission to use it.

For the joint probe, agree on a small request budget before sending. Three
attempts are a useful first observation: one candidate warm-up and two repeats.
An incomplete or unmeasured attempt still consumes that agreed request budget.
They are not enough to establish a service-wide guarantee. Use one model, fixed
reasoning/settings, a fixed harmless reference prefix, and a short requested
answer. Meet the selected model's documented minimum cacheable length; do not
assume that a byte count or the old 4,998-token example establishes eligibility.
Keep message boundaries and any tool definitions stable. Append changing text
after the reusable material rather than inserting it ahead of the prefix.

Use `--billing chatgpt --surface responses`, the explicitly approved
`--credentials-dir` and account, and a fresh owned data directory. Leave the
public base URL at its default. Do not use `--no-stream` for this OAuth route.
The CLI already accepts `--cache-lens DIR` and `--prompt-cache-key KEY`; a stable,
nonsecret key can make the comparison reproducible, but the current guide says
its routing role varies by model. It is not a cache switch or a hit guarantee.
A clean probe should not activate recall, compaction or tools that unexpectedly
change the prefix. It need not repeat all earlier chapter exercises.

Keep the measurement record small:

| Record | Why it matters |
| --- | --- |
| Actual model, route, funding mode and request settings | Avoid comparing different configurations. Never record the bearer. |
| Stable-prefix bytes and message boundaries | Explain what was held constant; bytes are not tokens. |
| Terminal success or explicit failure | Only completed requests supply successful usage observations. |
| Total input and reported cached-read tokens | Compute reuse from provider counts, not a predicted prefix ratio. |
| Any reported cache-write fields and latency | Distinguish warming, reuse and elapsed time without inventing cost. |

The existing Engine-owned `CacheLens` saves the current and previous wire/body
projections and logs byte-prefix predictions separately from measured reuse.
Those captures can contain prompt content: use harmless test material and an
owned private output directory. They are not a credential store. Missing usage
remains unknown. A failed or interrupted stream must not become a fabricated
zero-hit successful request.

There is a specific limit in the current C# parser. `Engine/Responses.cs` splits
`input_tokens_details.cached_tokens` from total input, but does not separately
classify the newer `cache_write_tokens` and per-lifetime write fields. The lens
therefore does **not** establish their absence. For a model reporting them, retain
only the relevant raw terminal usage fields in the bounded probe record; do not
claim the existing price estimate is a complete bill. Plan usage is not an API
invoice in any case. This is a documented measurement limitation, not a reason
to guess prices or expand this optional chapter into a pricing-system rewrite.

The current [prompt caching guide](https://developers.openai.com/api/docs/guides/prompt-caching)
describes model-dependent minimums, message-boundary matching, cache writes and
key behavior. Recheck it when running the exercise. `store:false` does not by
itself establish whether prompt caching works. Conversely, a high byte-prefix
prediction does not establish that provider cache routing reused that prefix.

If the two warm repeats miss, preserve the request settings and terminal usage,
check eligibility and stability, and decide with the human whether another small
probe is justified. Do not retry indefinitely or treat one small sample as proof
that every public OAuth request is broken. If the test cannot run, write
**untested**, not **fixed**, **broken**, or **passed**.

## 24.4 What the earlier detour taught us

The following is historical implementation evidence, not the C# exercise above.
The Go adapter selected a separate `codex` connection, reread a human-selected
Codex credential file before each request, and left renewal to Codex. It refused
redirects and incompatible requests and did not fall back to another payer.
Its old environment variables and file-discovery defaults are not C# features.
No authorization to access those files follows from reading this account.

The first implementation answered correctly and still missed the cache. An
actual terminal session created a Python function, ran its assertion and
submitted the result. All three requests reported zero cached tokens. A second
run carried a longer stable reference and also reported zero. Merely having a
`prompt_cache_key` in the JSON was insufficient for those runs.

In that implementation, Codex's client source supplied the missing detail:
the backend derived cache affinity from the HTTP `session-id` header. The corrected adapter sent the same stable
identity in that header and in the body. The identity was derived from the
Agent's data directory and model, hashed so the directory itself stays local.
It needs to remain the same across calls; generating a fresh identity for every
request defeats its purpose.

The corrected terminal run used GPT-6 Astra at high reasoning, with a
200-row inert reference to exercise a substantial stable prefix. The task was
to create `triple.py`, assert that `triple(14) == 42`, and submit `LONG_OK`.
The real tool loop made these requests on October 11:

| Requested action | Total input tokens | Cached input tokens |
| --- | ---: | ---: |
| Write the function | 8,241 | 0 |
| Run the assertion | 8,297 | 8,064 |
| Submit the result | 8,351 | 8,192 |

The last two requests reused about 97.2% and 98.1% of their input. The session
finished normally in 7.844 seconds, and a separate execution verified the
generated function. These are measured cache grants during tool use. The
earlier misses remain part of the [implementation record](../../docs/edition-2-notes/codex-endpoint-chapter.md),
including the source of the header requirement. This is a bounded successful
run, not a promise of that hit rate on every request.

The lesson is not “always use this private route.” It is that a plausible request
and a correct answer were inadequate evidence. The missing affinity detail only
became visible because we measured the real tool loop. The reverse lesson matters
just as much: once the supported route works, stop maintaining an unnecessary
detour.

## 24.5 Where this execution stops, and where the book continues

The optional Codex path is working through the actual C# application. It remains
a separate, explicitly selected grant using a backend that can change; it is not
a replacement for the documented public OAuth contract. The public comparison
reported no reuse; the integrated Codex check reported two hits and one miss.
Keep both facts. A correct answer, a successful login and a plausible affinity
header are each weaker evidence than the terminal usage record.

Tests also establish the less visible behavior: paired rereads after replacement,
malformed/duplicate/oversized-file refusal, safe cancellation and destruction,
redacted actor errors, no retry on Codex401/429/redirect, and no successful usage
for incomplete streams. Host checks establish selection, public-store separation
and displayed funding. They do not make a security audit, a billing guarantee,
a live GUI test or a claim that every provider will behave well tomorrow.

The chapter's end condition is not “we have a workaround, therefore keep it
forever.” When the public route gives satisfactory measured reuse for the intended
workload, prefer it and retire unnecessary detours. Until then, preserve a clear
escape hatch without silently changing who pays. Chapter25 records a different
kind of failure: the gap between a patch, a passing test and the application a
person actually uses. Gateway and Goals continue on the same ownership model.
