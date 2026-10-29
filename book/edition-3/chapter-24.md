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

**C# execution status:** the current official route has been checked in the
documentation; real OAuth and the controlled cache exercise remain deferred until
Bill can authenticate. No Codex credential adapter has been implemented in this
edition. Bill authorized continuing the independent chapters while these live
checks remain open. This is neither a cache success nor a reproduced failure.

## TL;DR

**This chapter is optional and has no separate grader.** It adds a decision gate,
not an unconditional new connection mode. The C# implementation continues from
Chapter 23 with Chapter 16's public OAuth flow. Gateway does not require the old
detour.

1. With the human present, finish Chapter 16's real OAuth exercise using an
   explicitly approved Ensemble credential directory. Do not search Codex's files,
   borrow a running application's configuration, or copy tokens between clients.
2. Recheck the official third-party inference documentation. At this revision it
   specifies `POST https://api.openai.com/v1/responses`, `store:false`, streaming,
   and `response.completed` as the success boundary. It explicitly excludes
   ChatGPT's `backend-api` endpoints for this public OAuth integration.
3. Run a small controlled probe with one available model, one funding source,
   stable instructions and tools, and a reusable prefix long enough for that
   model's current cache rules. Measure actual completed-response usage. A login,
   a correct answer, or a large byte overlap is not a cache measurement.
4. If reuse is satisfactory, skip the workaround. If it fails reproducibly,
   first check request stability, eligibility and current provider requirements.
   A failure does not itself authorize a private endpoint or another application's
   credentials. Establish an authorized, currently supported alternative before
   implementing one.
5. If authentication or a conclusive probe must wait, record that uncertainty and
   continue the independent work. Do not label an unperformed exercise passed,
   and do not manufacture a Chapter 24 code change to fill the gap.

The runnable C# foundation is `solutions/edition-3/ch23/`; subsequent work continues
in `solutions/edition-3/main/`. There is no new Chapter 24 implementation endpoint
or executable score. Historical Go files are evidence for the account below, not
C# acceptance or current setup instructions.

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

## 24.2 Taking a measurement in C#

Do this with the account holder, not as an unattended credential-discovery task.
Chapter 16 supplies the human `auth status`, `auth login` and `auth models`
commands and their private-directory setup. Use those commands to establish the
approved account and discover a model actually visible to it. A model name in an
old table is neither current availability nor permission to use it.

For the joint probe, agree on a small request budget before sending. Three
completed requests are a useful first observation: one warm-up and two repeats.
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

## 24.3 What the earlier detour taught us

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

## 24.4 Where this execution stops, and where the book continues

The official inference and caching pages were fetched successfully for this
revision. No real OAuth exchange or inference/cache probe was performed in this
C# chapter. No Codex adapter, credential lookup, endpoint change, or new pricing
claim was added. The historical measurements remain attributed to their earlier
runs. Existing local Responses, authentication and cache tests prove their own
bounded contracts, not today's provider cache behavior.

We will return to the live exercise together. Meanwhile, Chapter 25 is a historical
interlude and Chapter 26's Gateway can be built on the unchanged Chapter 23 C#
foundation. Independent progress does not erase the open check. The best code for
an unproven workaround is, for now, no workaround at all.
