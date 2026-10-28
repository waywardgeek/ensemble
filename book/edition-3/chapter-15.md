# Chapter 15: The Invisible Invoice

Bill reported $2,475 in model spend on the first edition through its recall
chapter. He attributed a few hundred dollars to a caching bug that produced
correct answers, on time, while quietly losing the discount on repeated input.
The answers gave him no reason to look for it.

Every request the agent sends carries a prefix: tool declarations, the system
prompt including skill manuals, then retained conversation. Much of that prefix is
identical from one turn to the next. A provider cache can make those repeated
tokens cheaper to read. Lose the cache and the same useful answer arrives with
a different bill.

The bug left no obvious trace in the agent's behavior. No test failed. No log
line turned red. An aggregated invoice is a poor place to discover which turn
first lost its prefix. Correct output can conceal a cost defect indefinitely.

This chapter builds the instrument that detects prefix changes, the markers
that identify reusable boundaries, and the meter that reports the cost while
the operator can still do something about it.

---

## TL;DR

Engine owns an optional `CacheLens` and its process-local meter. History retains
canonical response counts, never a mutable dollar balance. The lens observes
actual sends without changing them; an immutable prediction follows each request.
Marker selection runs on final provider JSON, after tool-result adjacency.
Anthropic support is explicit opt-in; OpenAI uses implicit caching with an
optional routing key. Exact response model/tier selects a conservative price.
Unknown values remain unknown. GUI projects one coherent Engine snapshot.

### In plain words

First ask, “Did we send the same reusable beginning?” Then ask, “Did the provider
report reading it from cache?” Finally ask, “Do we know the applicable rate?”
Those are three questions, not one. A correct answer to the user's task answers
none of them. Keep the evidence separate so a student can find which assumption
failed without rewriting conversation history or guessing a cheaper bill.

### Prerequisites and scope

Continue the C# Chapter 14 endpoint in `solutions/edition-3/main`. Preserve
`solutions/edition-3/ch14`; it is a runnable predecessor, not a directory to edit.
Use C# 14 and .NET 10. The CLI, GUI and public examples reference the same core
assembly. No new package or caching service is required.

This chapter adds three instruments: an Engine-owned cache lens, explicit
Anthropic marker placement, and a process-local usage meter. OpenAI Chat
Completions keeps its implicit caching and may carry a documented routing key.
Gemini keeps its existing implicit path. A fake can exercise the instruments;
it cannot award evidence of a real cache hit, model availability or billing.

## §15.1 Correct answers can hide waste

A cache miss can produce a completely correct answer. The tool works, the edit
lands, and the test passes. Nothing in that behavior tells the operator whether
the provider read a stable prefix from cache. A cost problem can remain invisible
until an aggregated invoice arrives, long after the request that caused it.

Build the instruments in an order that makes diagnosis possible:

1. **The lens** shows the actual request and a byte-preserving approximation of
   its provider-ordered prompt. Compare consecutive requests before guessing
   which content changed.
2. **The markers** express supported cache boundaries. A legal marker is not
   proof that a provider accepted the request or reused its prefix.
3. **The meter** shows reported tokens and a conservative price estimate. It
   must not disguise an unknown price as zero dollars.

The distinction between prediction and observation governs all three. A lens
predicts reusable bytes. A response reports tokens served from cache. A price
row estimates a charge. None is a substitute for the provider's invoice.

Do not store dollars in History. Record usage in the existing canonical response
facts, while Engine separately accumulates successful sends for this process.
Restoring yesterday's conversation must not silently charge yesterday's tokens
to today's display.

## §15.2 Vocabulary and ownership

Read `Common/Cache.cs` and `Common/Meter.cs` before their implementations in
`Engine`. Common declares shared state and callable partial members; the
functional files supply their bodies. The compiler combines these parts into
one type, not linked wrapper objects.

| Object or value | Responsibility |
|---|---|
| `Engine` | HTTP exchange, optional lens ownership, process-local usage rows |
| `CacheLens` | Consecutive send captures, byte-prefix prediction and diagnostic files |
| `CacheCapture` | Value-like wire, prompt and options byte arrays |
| `CachePrediction` | Immutable measurement belonging to one admitted request |
| `UsageIdentity` | Value-like vendor, actual response model and service tier |
| `MeterSnapshot` | Coherent immutable latest response, grouped rows and derived totals |
| `TokenPrices` | Conservative exact-identity price lookup; no mutable balance |

A lens constructor receives its real Engine parent and registers both ends.
`Engine.Lens()` reaches the child; `CacheLens.Parent()` reaches the owner. A
second lens cannot silently replace an attached one. Explicit `Destroy()`
detaches both directions, and Engine destruction destroys its lens. There is
no global lens registry or copied logger: diagnostics reach the application
through the existing parent relationships.

Snapshots and identity records are explicitly value-like. They need no parent
back-pointers merely because an Engine returns them. A retained snapshot remains
unchanged after another response completes.

The library never reads host environment variables. CLI and GUI composition
roots translate their flags into explicit configuration and construct the lens.
Engine owns the resulting object's lifetime.

## §15.3 The lens: observe without rewriting

`new CacheLens(agent.Engine, directory)` attaches private diagnostics. Its
`Begin` operation runs at actual transport admission, not during offline
`Render`, parser tests or restored-history replay. Pre-cancelled sends do not
rotate the captures.

The selected directory contains three pairs:

| Current | Previous | Meaning |
|---|---|---|
| `request.wire.json` | `request.prev.wire.json` | Exact UTF-8 request bodies, including provider markers |
| `request.body` | `request.prev.body` | Marker-stripped prompt properties in provider order |
| `request.options.json` | `request.prev.options.json` | Remaining model/generation/routing properties |

`request.prediction.json` also records the latest sequence and byte comparison.
On completion, the existing application logger receives that request's prediction
alongside its reported usage. The wire/body/options files are consecutive-send
diagnostics, not a complete per-request archive.

There are no authorization headers in these files. Nevertheless, prompt text,
file content and tool arguments can be private. Keep captures outside Git and
publication artifacts. Selecting a diagnostic directory is not permission to
publish its contents.

### Separate the prompt from options

Raw JSON property order is not necessarily the provider's context order. The
C# diagnostic uses these explicit sequences:

- Anthropic: tools, system, messages.
- OpenAI Chat Completions: tools, messages, including the system message.
- Gemini: systemInstruction, tools, contents.

The outer diagnostic envelope is new. Each selected property's lexical bytes
are retained rather than parsed into a DOM and serialized afresh. Options such
as model and sampling settings belong in the separate options view, not ahead
of the prompt where they would spoil a prefix comparison.

This is a local approximation, not access to a provider's tokenizer or internal
cache key. Compare the raw wire too. A prompt-only view cannot diagnose a changed
routing option by itself.

### Strip only genuine markers

The moving stable-end marker changes location as history grows. Comparing it as
conversation content would report a changed prefix on a healthy request. The
comparison therefore strips genuine provider `cache_control` fields while the
actual wire remains untouched.

Do not recursively delete every property named `cache_control`. A user's tool
argument or JSON schema may legitimately use that name. `CacheBytes.cs` walks
tokens with their structural paths, identifies legal provider marker locations,
and copies the untouched byte spans. It handles the adjacent commas without
normalizing escaped text, whitespace, key order or number spelling.

The first-edition account describes a useful failure here. Its initial strip
matched one compact spelling of a marker, while the actual formatter inserted
spaces. Hand-built compact fixtures agreed with the broken implementation.
A fixture built by the actual formatter exposed the mistake. Those historical
measurements are not results from this C# implementation; the lesson survives:
**test the artifact the application produces, not a convenient imitation.**

### Keep each prediction with its request

The lens lock serializes capture rotation, comparison and file writes, not
HTTP. More importantly, the predicted fraction travels as an immutable value
with the exchange that produced it. A late response cannot consult a shared
“latest prediction” and borrow a newer request's result.

The actor normally serializes its turn decisions, but Engine also has direct
callers, and asynchronous completion deserves its own identity. Merely locking
one mutable latest value fixes a data race without fixing attribution.

A capture failure is a diagnostic failure, not an inference failure. An
unwritable directory or malformed caller-supplied request disables that capture
and logs the problem without changing the bytes sent or inventing a cache
measurement. The files are a debugging instrument, not a transactional archive;
inspect reported capture failures before treating a pair as complete.

## §15.4 Mark the final projection

Marker placement belongs in disposable provider JSON. It must not mutate
History, skill snapshots, tool definitions or provenance-bearing parts.

The order of work matters:

1. Render the frozen Context and its allowed declarations.
2. Complete provider-specific tool-call/result adjacency and turn merging.
3. Retain the distinction between stable blocks and request-only ephemera.
4. Choose legal boundaries in that final shape.

Selecting a marker on the last neutral Context entry before rendering can be
wrong: the provider adapter may move a result next to its call or merge user
turns. Likewise, the final user message can contain both stable text and
transient ephemera. Message position alone does not identify a stable block.

### Anthropic: explicit opt-in, at most four boundaries

`Configuration` accepts `cacheBreakpoints: true`. It is off by default; neither
a model-name prefix nor a test fake establishes support. The C# renderer may
mark:

1. The last **non-deferred** tool declaration, if one exists.
2. Nonempty SYSTEM instructions.
3. The surviving marked handoff note, if present.
4. The last legal stable conversation block before transient ephemera.

Coincident boundaries collapse. Missing content does not require placeholder
text merely to use four slots. Empty text, thinking and unknown opaque blocks
are not explicit targets. A tool declaration with `defer_loading: true` must
not receive `cache_control`; if all declarations are deferred, omit that
boundary rather than choosing an illegal one.

The [Anthropic prompt-caching guide](https://platform.claude.com/docs/en/build-with-claude/prompt-caching)
and [tool-search guide](https://platform.claude.com/docs/en/agents-and-tools/tool-use/tool-search-tool)
are the primary references checked for this implementation. Check the current
selected model and route before live use. No TTL is sent here: the marker uses
the documented default, not a client promise of retention.

The tool boundary remains useful when SYSTEM changes: an earlier tool prefix
can still be reusable. The rolling history boundary covers more than the first
user message; fixing it at the beginning would leave the growing conversation
outside the intended boundary.

Chapter 9's rule still holds: loaded manuals belong once in SYSTEM. A dynamic
activation legitimately changes that instruction prefix. Do not move manuals
into dialogue to conceal the change. Pending manuals remain SYSTEM instructions
through Chapter 14 cuts, although their pending-only tool authority is revoked.

A handoff changes the working context, so an old suffix may cease to match. The
marker describes retained bytes after the recorded reduction; it cannot keep
removed tool traffic alive. This chapter implements the surviving handoff
boundary, not a future memory-compaction policy.

### OpenAI and Gemini: do not copy an analogy onto the wire

This endpoint uses OpenAI **Chat Completions**, not Responses. It retains
implicit caching and optionally supplies `prompt_cache_key`, documented in the
[Chat Completions request reference](https://developers.openai.com/api/reference/resources/chat/subresources/completions/methods/create).
A routing key is not a breakpoint and is not a credential. Do not put secrets in
it. Configuration refuses this option on a different provider.

The earlier manuscript describes experiments with `prompt_cache_breakpoint`
and `prompt_cache_options.mode`. Those descriptions are historical records,
not current API contracts verified for this C# route. This implementation does
not emit them or turn off implicit caching by analogy with another provider.

Gemini retains its existing implicit path. No explicit cache object, TTL
manager or cache registry is introduced. Its
[caching guide](https://ai.google.dev/gemini-api/docs/caching) is a starting point
for an authorized live experiment, not proof of a hit here.

### Eligibility is separate from stability

Cache minimums belong to a particular model and route, and are measured in
tokens. A 6,000-byte fixture does not establish a 1,024-token minimum. A fake can
award reads to any prompt size. Zero reads can mean ineligibility, expiry,
routing, changed content or another provider decision; do not pick a cause
without evidence.

Earlier accounts compared fixed large references with rapidly growing history
and got different warm-up behavior. A fixed-reference probe cannot refute a
claim about growing history. Similarly, changing content encoding and removing
a marker in one probe does not isolate either variable. These are lessons in
controlled experiments, not universal cache thresholds or token-block rules.

## §15.5 Count first, estimate second

Engine records usage exactly once for each successfully completed send, after
its completed response facts have been produced. Plain and streaming requests
share final-response decoding. Cumulative streaming updates are not added as
separate purchases.

Parsing a fixture, rendering offline, replaying History, failed HTTP and
cancelled exchanges do not advance this meter. A provider may charge for a
request whose successful usage never reached the client. The GUI explicitly
calls this a token-price estimate, not an authoritative account balance.

### Four disjoint buckets

`Usage` retains nullable ordinary input, cache write, cache read and output
counts. Missing required usage is unknown. Omitted optional cache buckets may
be zero; malformed supplied values must not become reassuring zeroes.

Normalize before pricing:

- Anthropic input/write/read buckets are disjoint.
- OpenAI's prompt total includes cached reads; subtract them to obtain ordinary
  input. The ordinary Chat Completions pricing fixture has no cache-write bucket.
- Gemini's prompt total includes cached reads. Its output accounting includes
  separately reported thinking without counting that thinking twice.

`AllInput()` sums the three disjoint input buckets. Unknown operands or overflow
produce unknown, not wraparound. Session hit rate is:

```text
session cache-read tokens / session ALL-input tokens
```

A zero or unknown denominator yields unknown. This differs from the lens's
byte-prefix fraction and from the latest response's cache-read count.

### The actual response identity owns the rate

`UsageIdentity` retains vendor, actual response model and service tier. Engine
groups counts by that identity under one short lock. A later model or tier must
not retroactively reprice earlier requests. `Meter()` captures latest and rows
together; `MeterSnapshot` derives totals, hit rate and estimated dollars.

Prices are intentionally narrow. The implementation verified these standard
text-token rows against the primary [GPT-4o mini](https://developers.openai.com/api/docs/models/gpt-4o-mini)
and [GPT-4o](https://developers.openai.com/api/docs/models/gpt-4o) pages:

| Response model or documented snapshot | Ordinary input | Cached read | Output |
|---|---:|---:|---:|
| `gpt-4o-mini`, `gpt-4o-mini-2024-07-18` | $0.15 | $0.075 | $0.60 |
| `gpt-4o`, `gpt-4o-2024-08-06` | $2.50 | $1.25 | $10.00 |

Amounts are USD per million tokens, with explicit response tier `default` only.
A page can list a deprecated model: a price row is not availability evidence.
Missing tier may imply project-level routing, so it is not silently standard.
Unknown identity, a different tier, unknown counts or unverified cache-write
charges make the estimate unknown. Other providers retain valid counts without
invented prices. Update verified pricing deliberately, not by broad model-name
matching or aliases invented to make fixtures pass.

The calculation uses decimal arithmetic at view time:

```text
(input × input rate + cached reads × read rate + output × output rate)
/ 1,000,000
```

The general four-bucket formula also includes cache writes when a verified
write rate exists. This table does not supply one; nonzero or unknown writes
must not be treated as free. If any contributing session row is unpriced, the
whole session estimate stays unknown rather than displaying a partial subtotal
as its total cost.

## §15.6 A GUI projection, not another meter

The header shows **COST**, **ALL**, **CACHE** and **HITRATE**:

- COST: derived estimate for this Engine's current process.
- ALL: input total for its latest completed response.
- CACHE: cache reads for that latest response.
- HITRATE: the session ratio, including its cold first response.

`Server/Meter.cs` takes one coherent snapshot. The browser formats acknowledged
values; it does not accumulate tokens or own a price table. Both connected
clients receive the same projection. Refresh does not create usage. Periodic
catch-up repairs missed observations, and a cached serialized frame suppresses
identical notifications without becoming an accounting authority.

The published WebSocket shape is:

```json
{
  "type": "usage",
  "agent_id": "root",
  "usage": {
    "input_tokens": 400000,
    "cache_write_tokens": 0,
    "cache_read_tokens": 600000,
    "output_tokens": 100000,
    "all_input": 1000000,
    "totals": {
      "input_tokens": 400000,
      "cache_write_tokens": 0,
      "cache_read_tokens": 600000,
      "output_tokens": 100000
    }
  },
  "total_cost_usd": 0.165,
  "hit_rate": 0.6
}
```

That example uses the mini/default price row. The display rounds its estimate
to `$0.17`; the frame retains the more precise value. Unknown is JSON null and
visible `?`, never silently zero. Expanded details expose session counts and
explain the estimate and restart limits without requiring hover alone.

Before any response, latest counts and hit rate are null, session counts are
zero, and estimated spend is zero. A new Engine starts this way even after
restoring a conversation. An HTTP error does not erase or advance its prior
successful meter. Usage notifications are silent to the speech queue.

## §15.7 Build, exercise and grade

From `solutions/edition-3/main`, with the approved SDK and owned offline
environment:

```sh
dotnet run --project Tests/Tests.csproj
dotnet build Cli/Ch04.csproj
dotnet build Examples/Gui/Gui.csproj
dotnet run --project Examples/Cache/Cache.csproj -- /new/owned/cache-exercise
```

The separate Cache example uses only public library APIs. It runs two actual
Agent turns against an owned loopback provider, checks exact captures and
immutable snapshots, changes from a priced tier to an unpriced tier, restores
saved conversation into a new empty meter, and destroys both directions of the
lens relationship. It does not access a paid provider.

Both hosts accept:

| Flag | Environment fallback | Meaning |
|---|---|---|
| `--cache-lens DIR` | `EN_CACHE_LENS_DIR` | Opt-in private capture directory |
| `--cache-breakpoints` | `EN_CACHE_BREAKPOINTS=1` | Explicit Anthropic marker support |
| `--prompt-cache-key KEY` | `EN_PROMPT_CACHE_KEY` | OpenAI routing key |

Explicit value flags override environment values. These are host configuration,
not saved browser preferences. Missing values and a routing key on the wrong
provider refuse startup. No flag establishes that a selected model accepts the
request.

From the repository root, use the existing Go grader's narrow C# adapter:

```sh
grade-edition2 -ch 15 -executable-only -json \
  -gui /absolute/path/to/built/Gui \
  -cache-prices .15,0,.075,.6 solutions/edition-3/main
```

Use the actual built GUI apphost, with its runtime dependencies and .NET root
available, or an equivalent launcher. Unlike early chapters, `-gui` here names
an executable, not static assets. The adapter creates owned workspaces, sends
real `/ws` prompts, waits for completed turns, and joins SIGTERM shutdown.
The original Go route and its historical defaults remain unchanged. In this
C# invocation supply the shown exact mini/default fixture rates; the original
defaults are not the C# pricing contract. The adapter uses the C# fixed lens
filenames, not the Go route's `-cache-current`/`-cache-prior` paths.

| Check | Points | C# executable scope |
|---|---:|---|
| `cache-lens-exists` | 15 | Current/prior full wire equal the actual sent bodies |
| `prefix-is-stable` | 20 | Unchanged tools and SYSTEM across a growing conversation |
| `request-has-breakpoints` | 15 | Legal Anthropic tool/SYSTEM/advancing stable-history markers |
| `cost-is-computed-not-stored` | 15 | Count-only events, independently priced OpenAI fixture, unknown model/tier |
| `usage-is-session-scoped` | 10 | Prior dialogue restored while new latest counts stay unknown and totals zero |
| `cache-read-rate` | 10 | Disjoint cold/warm fixture counts produce the expected session ratio |
| `all-vendors-report-usage` | 15 | Actual adapter requests and normalized final counts for all three providers |

The marker session and priced session are deliberately separate. An unpriced
Anthropic fixture must not acquire a fictional tariff merely because the
historical grader combined those tests.

The marker grade has no handoff or ephemera. Focused component tests separately
cover final adjacency, deferred declarations, empty instructions, retained
handoffs and transient tails. Out-of-order completions, malformed/unknown usage,
streaming, capture failures and ownership also need their own real-component
checks. A numeric grade is not an architectural review.

### Personally use the application

Run the GUI from a fresh owned workspace on a free loopback port, not a working
user's port or browser profile. Supply provider values as in earlier chapters;
keep real credentials out of commands, source files and captures. For example,
when an authorized Anthropic route has already been configured:

```sh
dotnet /absolute/path/to/Gui.dll --port 18085 \
  --cache-lens ./private-cache --cache-breakpoints
```

For an OpenAI route, omit `--cache-breakpoints`; optionally use
`--prompt-cache-key chapter15-owned-demo`. For a fake, configure its loopback
base URL and dummy key instead. Choose the actual model explicitly; the older
manuscript's model labels are not recommendations or availability checks.

Send two short questions with a genuinely stable reference between them. Wait
for each completed turn. Inspect the header and expanded totals, then compare
`request.body` against `request.prev.body` and the raw bodies.
Check the provider's usage rather than concluding that a plausible answer means
a cache hit. With a real provider, first verify its current eligibility rules
and use a small bounded authorized exercise.

Refresh and open a second owned tab: both must show the same meter without
another request. Stop the owned host gracefully with Ctrl-C, then restart it in
the same workspace. Conversation should return, but current-process latest
usage and hit rate should be unknown and cost zero. No inference is required
to restore the display. Do not use EOF as the GUI's shutdown contract.

### What this endpoint establishes

The preserved C# boundary has 99 passing component groups, a scoped executable
100/100, a separate public consumer, CLI flag/capture exercises, and owned GUI
exercises using plain and SSE responses. Browser checks cover two clients,
refresh, restart isolation, unknown pricing/counts, failed HTTP, speech silence
and a 360-pixel-wide layout. DOM/geometry evidence is not pixel assessment or
a human accessibility audit.

Live API testing was authorized, but no approved credential source was available
in the test session. Thus no C# live-provider cache grant, billing, current model
availability, native speech or human-accessibility result is claimed. Historical
cache counts and dollar tables from previous implementations cannot fill that
gap. Correct fake replies establish the client contract; only an authorized
real exchange can establish what the real provider did.

The enduring lesson is not “the cache works.” It is that useful engineering
keeps the wire, prediction, reported counts, estimated cost and evidence limits
separate enough to diagnose when one of them is wrong.
