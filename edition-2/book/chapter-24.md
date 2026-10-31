# Chapter 24: While the Cache Is Broken

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

Hopefully this chapter is obsolete by the time you reach it. Until then,
Ensemble needs a way to keep working: explicitly select the Codex endpoint
with an existing Codex sign-in, and keep measuring what comes back.

## TL;DR

**This chapter is optional and has no separate grader.** If the public OAuth
route gives satisfactory cache reuse for your workload, skip it. Keep the
existing public authentication flow. The workaround is for the failure observed
here, not a prerequisite for the next feature.

1. Add `codex` as an explicit OpenAI connection choice through
   `OPENAI_BILLING_MODE=codex`. Keep `chatgpt` for Ensemble's public OAuth
   registration and `api` for an explicitly selected API key. Never fall back
   from Codex to either of those routes automatically.
2. Read the selected Codex auth file at request time. The CLI selects
   `EN_CODEX_AUTH_FILE`, otherwise `CODEX_HOME/auth.json`, otherwise
   `~/.codex/auth.json`. Read `tokens.access_token` and `tokens.account_id`
   together. Do not copy them into prompts, saves or logs. Codex owns renewal;
   Ensemble never writes or refreshes this other application's credentials.
   A missing or rejected sign-in produces an actionable error.
3. Send supported Responses models to
   `https://chatgpt.com/backend-api/codex/responses`, with bearer authorization
   and `ChatGPT-Account-ID`. Send a stable `session-id` header equal to the
   request's cache key; Codex uses it for cache affinity. Refuse redirects. Keep the existing streaming
   parser, tool loop and usage accounting. Reject incompatible models before
   sending credentials.
4. Send `store: false`, `stream: true`, instructions and the complete local
   input. Keep a stable `prompt_cache_key` for the Agent and model. Omit the
   explicit cache options and breakpoints on this route. A key helps identify
   reusable work; it does not guarantee a cache hit.
5. Preserve streamed text, tool results and eligible encrypted reasoning across
   requests. Count a response only when its terminal completion arrives. Report
   actual cached-token grants; errors and interrupted streams are not cache
   misses with zero usage.
6. Test route selection, credential rereading, refusals, real tool execution,
   replay and usage with a small external HTTP fake, then exercise the actual
   CLI with a live model. The implementation tests are useful even though this
   optional chapter has no score.

Work in `agent/`. The reviewed reader snapshot for this
chapter is `solutions/ch24/`. Run the implementation checks from the
working tree with `go test ./...` and `go vet ./...`; the optional GUI is a
separate module and needs its own checks. Helper names and file organization
remain implementation choices within the existing ownership rules.

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

For a concrete price comparison, the GPT-5.6 Sol rates checked on October 11,
2026 are $4 per million ordinary input tokens and $0.40 per million cached
input tokens. Fifty repetitions of a 30,000-token prefix amount to 1.5 million
input tokens: $6 at the ordinary rate or $0.60 if all those tokens received
the cached rate. In practice the first request needs warming and changing
input still costs separately. These are API-rate illustrations, not a
measurement of a ChatGPT subscription invoice. The tenfold difference applies
to the affected input tokens. [Model pricing](https://developers.openai.com/api/docs/models/gpt-5.6-sol).

The fifteen-request comparison in the opener is Bill's reported experiment.
A fresh paired probe on October 11 used the same GPT-5.6 Sol request body
with Ensemble's public OAuth credential and with Codex's credential at its
own endpoint. Both completed public requests reported zero cached tokens out
of 4,998. Both completed Codex requests reported 4,864 cached tokens out of
4,998, about 97.3%. A third request on each route failed with an overload
error; those failures supplied no cache measurement. The [probe record](../docs/codex-endpoint-probe.md)
contains the receipts and reproduction details. This retest gave no evidence
that the problem had been fixed.

Changing the address alone is insufficient. Ensemble's public OAuth grant
and Codex's sign-in belong to different access paths. The documented
third-party flow uses the public Responses endpoint. This workaround uses
the existing Codex sign-in with the endpoint that Codex itself uses; it does
not send Ensemble's public grant there. [Public OAuth inference](https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference).

The useful boundary already exists. The authentication component obtains
credentials, the Engine sends HTTP and parses responses, and the Agent owns
its conversation. A different connection should not require a second agent
loop. Tool calls still become jobs, their results still enter history, and
the next request still comes from that history.

There is one deliberate inconvenience: Codex remains responsible for its
sign-in. Ensemble rereads the file before each request so a renewed token is
picked up without restarting. It does not compete with Codex to rotate a
refresh token. If the selected file is absent or the service rejects the
credential, repair the Codex sign-in and retry. A fallback API key would turn
an authentication failure into an unannounced spending decision.

Keep the escape hatch in both directions. Select `chatgpt` again to return to
Ensemble's own OAuth flow, leaving the Codex credential file untouched. Repeat
a bounded cache probe with a stable prefix before declaring the workaround
unnecessary. A successful login or a plausible answer is the same inadequate
test that let this problem hide in the first place.

## 24.2 Taking the detour

Build the Edition 2 executable, then run it from your working directory:

```sh
# From agent:
go build -o /tmp/ensemble-codex .

# From the directory where the agent should work:
export LLM_VENDOR=openai
export LLM_MODEL=gpt-6-astra
export OPENAI_REASONING_EFFORT=high
export OPENAI_BILLING_MODE=codex
unset LLM_BASE_URL OPENAI_BASE_URL
/tmp/ensemble-codex chat --no-web-search
```

This requires an existing file-backed ChatGPT sign-in in Codex. Set
`EN_CODEX_AUTH_FILE` if its auth file is somewhere other than the selected Codex
home. An API-key Codex profile or a credential stored only in the operating
system keychain does not supply the file this small adapter reads. Ensemble's
`auth` commands continue to manage its own public OAuth registration.

The first implementation answered correctly and still missed the cache. An
actual terminal session created a Python function, ran its assertion and
submitted the result. All three requests reported zero cached tokens. A second
run carried a longer stable reference and also reported zero. Merely having a
`prompt_cache_key` in the JSON was insufficient for those runs.

Codex's client source supplied the missing detail: the backend derives cache
affinity from the HTTP `session-id` header. The adapter now sends the same stable
identity in that header and in the body. The identity is derived from the
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
earlier misses remain part of the [implementation record](../docs/edition-2-notes/codex-endpoint-chapter.md),
including the source of the header requirement. This is a bounded successful
run, not a promise of that hit rate on every request.

To leave the detour, set `OPENAI_BILLING_MODE=chatgpt`, keep the base URL unset,
and use Ensemble's existing public OAuth sign-in. Neither direction changes
conversation ownership or requires copying tokens between applications. Once
that route gives satisfactory measured reuse, the workaround has done its job.
