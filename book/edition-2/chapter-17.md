# Chapter 17: The invisible invoice

A coding agent can make the right edit, pass the tests and still waste money on
every turn. The developer sees the edit. The repeated input charge is harder to
see, especially when a memory compressor or recall judge quietly contributes
another request before the answer appears.

In September 2026, Bill flagged caching as the barrier to running Ensemble every
day. So the project built its first instrument to show where the cache helped
and where it did not. That instrument promptly developed a problem of its own.
It searched for compact JSON, but the renderer's actual output was indented. The
test fixtures were compact too, so every one passed. Meanwhile the instrument
reported missing markers that were sitting right there in the request, invisible
only because the comparison disagreed about whitespace. A convincing number
makes an excellent disguise for a broken measurement.

This chapter starts with that instrument. It separates three quantities that
look like they should agree and never quite do: the bytes the application sent,
the cache usage the provider reported and the price estimated from accepted
usage. Each has a different source of truth, and a developer who trusts only one
of them will eventually be surprised by whichever two were ignored.

*Contract draft for independent review, October 8, 2026. The
[validation record](chapter-17-validation.md) owns release and implementation
gates. The design is selected; no live measurement, saving or completed student
build is claimed. Historical incidents are attributed in the
[evidence ledger](chapter-17-evidence.md).*

## TL;DR

Read the complete [coding skill](skills/ensemble-coding/SKILL.md),
[architecture](architecture.md) and accepted predecessor before coding.
Extend `solutions/edition-2/main/`; preserve old rendering and resume routes.

1. Engine owns a bounded diagnostic child observing exact prepared bytes at
   actual transport handoff. Separate foreground, compression and recall-judge
   pairs. Rendering/reconstruction performs no diagnostic admission or paid work.
2. Report raw byte prefix, ordered structural change and directive-excluded
   comparison separately, preserving offsets and opaque lexemes. Publish gaps,
   finite work/copy budgets, retirement and export completion as §17.2 specifies.
3. Select off or stable_v1 explicitly. Preserve old off bytes; respect transient
   prefix boundaries. Messages permits four manual markers, Chat Completions
   three alongside implicit placement, Generate Content remains implicit.
   Responses initially accepts off only. Unknown marker capability refuses.
4. Add the explicit v7 request profile in §17.5 without enabling memory/recall.
   Capture route, funding, binding, delivery, cap and cache policy before rendering.
   Preserve exact v1–v6 shapes. Replay has no authorization or remote effects.
5. Keep accepted usage on Engine, with durable/current-mount/purpose/last-accepted
   views. Use exact rational prices from immutable dated producing-identity rows;
   unknown rates and incomplete attempts stay visible. Plan usage is not an invoice.
6. Expose one copied status through human `/cache`, optional browser and public
   consumers. Bound wire values; retain uint64 exactly. Export bodies only on an
   explicit request. No automatically installed model tool is added.
7. Add explicit Responses API-key and ChatGPT-plan routes. API-key requests retain
   a remote output cap; the selected plan route has none. All Responses purposes
   stream. Preserve legacy API-key routes, including the plain 256-token judge.
   Finite request counts, deadlines and local byte limits are not remote token or
   charge ceilings. No automatic paid retry or funding fallback occurs.
8. Ensemble owns Connections; Connection owns registration, refresh and leases.
   Asynchronous Engine operations wait without blocking Actor. Refresh isolates
   waiters; invalidation races dispatch atomically. Actor settles valid usage once
   while canceled effects remain canceled. Protected registration bindings survive
   restart; credentials never enter session state, model text or browser status.
9. Retain one authoritative raw Responses payload with checked indexed references.
   Preserve ordered reasoning, phase, exact argument strings and result pairing.
   Foreign namespaces get controlled refusals. No partial call executes. Complete
   terminal response and valid usage are required; EOF/deltas are insufficient.
   Automatic cuts/compression preserve bundles. Only an authorized base-4–6
   handoff can retire them whole; snapshots keep compact helper accounting.
10. Exercise real human CLI, public two-Agent and browser paths, all three inherited
    adapters and both Responses funding modes. Demonstrate actual plan compression,
    judge and tool continuation within the published bounds. Retain initial failures,
    source-bound requests and a finite cache comparison matrix before final review.

Private method/type names remain the student's choice. Exact JSON fields,
commands, ownership, bounds and outcomes below are the public contract. Build
with the accepted predecessor's module commands, including:

```sh
go -C solutions/edition-2/main build -o /tmp/ensemble-ch17-cli ./cmd
make grade-dir CH=18 DIR=solutions/edition-2/main
```

CH=18 is the inherited diagnostic grader number. It cannot establish all these
requirements. The coordinator must publish the new independent acceptance command
in the gate record before student release; the distinguishing matrix in §17.9
already defines its scope.

## 17.1 Follow the request that incurred the work

Suppose the Agent reads a manual, receives a one-shot hint, calls read_file and
answers a question. A request counter can advance several times during that one
turn. If memory compression runs first, comparing its input with the subsequent
foreground request produces a dramatic difference and a useless explanation.

Attach the diagnostic to the Engine's actual request path. The Actor captures
policy and configuration, the responsible renderer prepares the body, and the
durable admission records the inherited request or helper fact. Only then can
transport begin. The diagnostic observes the very same immutable body at that
handoff. It never asks a second renderer what the body probably looked like.

| Owner | Responsibility |
|---|---|
| Agent | Current execution/cache policy, Actor and Engine |
| Actor | Serialized admission and acceptance; captured policy; retirement notification |
| Engine | Actual request operations, accepted usage and diagnostic child |
| Diagnostic child | Bounded comparisons, copied summaries and explicitly requested exports |
| CLI / optional GUI Page | Display owned snapshots through public interfaces |

The diagnostic's actual parent interface is Engine, with Engine → Agent →
Ensemble providing logging and facilities. Shared values and parent interfaces
belong in common; operations belong in the responsible implementation package.
Do not pass the child a second configuration pointer, GUI callback or logger bag.
The browser remains optional, and MCP's complete-message transport remains
replaceable. Existing granted GUI inspection can observe the new display.

Each local attempt is identified by mount generation, turn, durable admission
sequence, purpose and actual operation identity. Preserve requested and resolved
model, surface, safe route name, funding-mode label and local connection generation.
An account identifier adds no useful comparison information. A transport-started
attempt means that local handoff occurred; it proves neither remote receipt nor
billing. A canceled admission that never reaches transport remains distinguishable
from a sent request and does not advance the body pair.

Acceptance stays where it was. A complete valid foreground or helper response
contributes usage once after its durable accepted fact. A compressor can incur
accepted usage even when its proposed memory is rejected. A failed stream can
show text without producing accepted usage. Keep the latter attempt visible as
incomplete rather than attaching the preceding successful response's counts.

## 17.2 A useful instrument has a finite memory

The diagnostic keeps one current-route pair for each purpose: foreground,
compression and judge. A change of surface, selected/resolved model, funding
mode or authenticated connection generation clears that purpose's pair with a
reset reason. Ordinary configuration and projection changes remain comparable.
Otherwise turning a setting would erase the evidence that the setting changed
the prefix.

These are application limits selected for this exercise, not provider limits:

| Resource, per Agent | Bound and behavior |
|---|---|
| Retained bodies | Two bodies of at most 8 MiB each per purpose; 48 MiB total |
| Comparison work | One owned worker; at most 32 MiB auxiliary live storage, including token indexes, parsing, formatting and temporary body copies |
| Pending analysis | At most one waiting observation per purpose, referencing its retained body without another copy |
| Export | One admitted export, at most two additional 8 MiB owned body copies; a second request receives diagnostic_busy |
| Internal summary metadata | 1 MiB total, including bounded reasons and candidate lists |
| Public status | At most 256 KiB encoded UTF-8; whole rows only, explicit omitted counts |
| Difference detail | At most 128 paths, each at most 512 UTF-8 bytes; no request-value excerpts |
| Export manifest | At most 64 KiB, with streaming file writes using at most 64 KiB extra buffer |

MiB means 1,048,576 bytes. Reserve these budgets before allocation. The bound
applies to simultaneously live data, including copies kept by canceled work;
returning a handle does not make its storage free. A generic JSON decoder that
expands a small wire value into an unbounded object graph is unsuitable here.
An offset scanner can retain ranges into owned bytes and stop before exceeding
its index/depth budget. Limit nesting to 128 for diagnostic analysis; a deeper
otherwise valid request still runs, with analysis_limit in the diagnostic.

The Actor must remain responsive while comparison or export runs. Handoff admits
a bounded observation without waiting for disk, GUI or a previous comparison.
When analysis cannot retain the adjacent pair because a prior worker still owns
its budget, mark diagnostic_busy, clear the affected baseline and cancel obsolete
analysis. The attempt still runs. Do not retain an unbounded queue to avoid showing
a gap. A stale worker's identity prevents it from publishing over a newer status.

An over-8-MiB body records its identity, exact byte length and reason body_limit,
without retaining a truncated body. It clears the baseline. For sends A, oversized
B and ordinary C, C reports no_baseline; it cannot claim to follow A. This rule
also applies to busy and analysis-limit gaps. Count every actual handoff even
when its detailed analysis is unavailable; bounded latest status and counters
need not retain an independent permanent history of attempts.

Retirement has the same force here as in Chapters 15 and 16. Associate retained
bodies with the captured source identities they include. When a source is retired,
cancel affected analysis/export and drop those body copies. A temporary diagnostic
cannot become a hidden archive of removed memory, judge candidates or responses.
Close refuses new work, cancels and joins the child worker/export, then releases
all copies. No callback after close can restore them.

For an explicit export, reserve an owned pair and use a caller-selected new
directory. Create before.json and after.json exclusively, write their exact bytes,
then publish manifest.json last by an exclusive final operation. The manifest
names both attempts, byte lengths, lowercase SHA-256 digests, comparison status
and its version. No authorization headers are part of these bodies. Never merge
annotations into a file labeled as the request. Failure leaves no success manifest
and returns diagnostic_export; it does not terminate otherwise valid inference.

An accepted source-retirement request invalidates an in-progress unpublished
export before its success manifest can be published. Already published exports
are explicitly selected user artifacts with independent retention; snapshots do
not import them and resume cannot use them as authority. The manifest is the
completion boundary, so a directory containing only one body is visibly incomplete.

## 17.3 Measure three different things

The compact-pattern failure from the opening of this chapter left a useful
lesson: the instrument must analyze the renderer's output as it exists,
whitespace and all. A handwritten fixture that repeats the instrument's own
simplifying assumption proves only that the assumption is consistent with
itself.

**Raw prefix** is the number of equal leading bytes in the two actual UTF-8
bodies. Report zero-based byte offsets and original body hashes. If one body ends
at the divergence, that side reports end_of_body rather than an invented byte.
This measure includes whitespace, member ordering and cache directives.

**Structural comparison** compares ordered JSON structure with original scalar
lexemes. Object member reordering is changed; numeric 1 and 1.0 differ. A JSON
string's original escape spelling remains evidence. A dialogue array is appended
when all its old elements match and only whole trailing elements are added.
Replacing text inside its last message is an edit. Moving the closing bracket
does not make an appended array an equal raw-body prefix.

**Directive-aware comparison** applies that structural comparison while ignoring
only the supported cache-directive members at their documented adapter-owned
locations. It also reports per-section results. Compare Messages tools, system
and messages; Chat Completions tools and messages; Generate Content tools,
systemInstruction and contents; Responses tools and input. Missing and empty sections remain distinct.
Top-level request controls appear in a separate controls result. Do not collapse
all surfaces into a field named messages.

The scanner recognizes member names by their decoded JSON names while preserving
the original key/value spans. Reject duplicate members for analysis with
invalid_json rather than guessing which duplicate the server might use. Ignored
members have original [start,end) byte spans; derived comparison never rewrites
an exported body. An opaque value can be scanned for bounded structure, but its
lexemes and member order cannot be normalized or edited. A schema property named
cache_control and a sentence containing that text remain ordinary data.

The following ASCII fixtures contain no trailing LF. They are offline literals,
not captured requests or models to send to a provider.

```json
{"messages":[{"role":"user","content":"A"}]}
```

Call this A. The next two bodies are B and C:

```json
{"messages":[{"role":"user","content":"A"},{"role":"user","content":"B"}]}
```

```json
{"messages":[{"role":"user","content":"Z"}]}
```

| Pair | Raw equal prefix | First differing bytes | Structural messages result |
|---|---:|---|---|
| A, A | 44 | both end_of_body | equal |
| A, B | 42 | before `]`, after `,` | appended, one new element |
| A, C | 39 | before `A`, after `Z` | edited at /messages/0/content |

Independently count the literal bytes when building checks. The expected answer
must not come from the diagnostic being graded. Add exact-prefix input pairs
such as `[]` and `[] `: raw divergence is end_of_body versus space at offset 2,
while structural comparison is equal.

These two content blocks must compare equal after directive exclusion, even
though their raw bytes differ. The displayed indentation is part of the second
literal; its final closing brace has no trailing LF.

```json
{"type":"text","cache_control":{"type":"ephemeral"},"text":"A"}
```

```json
{
  "type": "text",
  "text": "A"
}
```

Embed the first at Messages /messages/0/content/0 so its marker is eligible for
analysis. Repeat with cache_control first and last in the object. Each reports
one marker, and equal directive-excluded content. Repeat through the actual
renderer, not only these fragments. Move the same member beneath a tool's
input_schema.properties: it now reports zero markers and a real structural
change. For Chat Completions repeat with prompt_cache_breakpoint and its proper
text-content location. Unsupported lookalikes remain data.

Further positive controls preserve `{"n":1.0,"s":"\u0041"}` exactly inside
opaque material; comparisons against n:1 or literal s:"A" remain changed. A
member-order fixture `{"a":1,"b":2}` versus `{"b":2,"a":1}` is changed in
both structural modes. Empty eligible content has zero markers, rather than a
percentage computed by dividing by zero. Oversize and parser-limit fixtures
must preserve the real request body and expose a diagnostic gap.

Equal bytes answer a local question. They cannot establish which machine received
the request, whether an entry had expired or which hidden material the provider
included. When the body is stable and reported reads are zero, the report has
found something worth investigating. It has not proved a backend defect.

## 17.4 Mark a prefix that will still exist

A manual can be persistent while the prefix leading to it is temporary. Consider
the ordered projection below. B is the base instruction, H a one-shot hint, S a
loaded Skills manual and P the current human prompt. These are teaching labels;
the renderer retains the actual captured text and Chapter 9's anchors.

```text
request 1: B | H | S | P
request 2: B |     S | P | answer | follow-up
```

A boundary after S in request 1 includes H. Removing H changes that prefix even
though S is unchanged. Moving the manual ahead of the hint would make the cache
diagram prettier by changing the conversation's meaning. Preserve the ordering
and choose the boundary before H instead.

The policy has two choices: off and stable_v1. Off adds no new fields or
block splitting, preserving the preceding renderer's exact bytes; it does not
promise to disable a provider's implicit caching. Stable_v1 derives candidate
boundaries from captured projection material in this priority order:

1. End of the current complete tool declarations, where the surface permits it.
2. End of the immutable base/primary instruction prefix.
3. End of the newest represented handoff, only when its entire preceding prefix
   contains no one-shot material.
4. Last ordinary persistent content boundary before the earliest represented
   transient hint, ephemeron, current observation or new current prompt.

Keep only eligible candidates; deduplicate coincident boundaries before taking
the surface's first available slots. Report each omitted candidate's reason:
unsupported_location, empty_text, opaque, transient_prefix, duplicate or slots.
There is no permanent historical turn-start anchor. Changing Skills grants can
change declarations; replacing memory or retiring recall can change earlier
content. Record those causes rather than claiming that the old prefix survived.

The new profile may split ordinary adapter-owned text at a captured material
boundary. Preserve exact concatenated text, role and call/result order. Existing
separators remain in their original side of the split. Never split, add a member
to or remarshal a bound opaque block. If a required boundary lies inside one,
skip it. Refuse an unsupported selected marker mode before paid admission with
cache_capability; do not send a speculative request and automatically retry it
without the marker.

For Messages, the selected form is manual block-level caching with at most four
boundaries and no top-level automatic cache_control. The new policy uses the
default five-minute lifetime. A text block ends as follows:

```json
{"type":"text","text":"BASE","cache_control":{"type":"ephemeral"}}
```

The final tool declaration can carry the same cache_control member. Empty text
and thinking blocks are ineligible marker locations; their existing content stays
intact. The documented prefix order is tools → system → messages.
[Messages caching documentation](https://platform.claude.com/docs/en/build-with-claude/prompt-caching).

For Chat Completions, the selected policy retains implicit placement and adds at
most three explicit boundaries. Supported ordinary text content uses:

```json
{"type":"text","text":"BASE","prompt_cache_breakpoint":{"mode":"explicit"}}
```

The top-level control is:

```json
{"prompt_cache_options":{"mode":"implicit","ttl":"30m"}}
```

This is a fragment, not a complete request. Do not add a breakpoint to a function
definition merely because the Messages adapter supports that location. The current
reference documents this control for GPT-5.6 and later. Use an explicit verified
capability entry for the exact resolved model/surface; a lexical name comparison
or a discovery listing alone cannot establish it. Unknown capability refuses the
selected mode while off remains available.
[Chat Completions reference](https://developers.openai.com/api/reference/resources/chat/subresources/completions/methods/create).

The October 8 prompt-caching guide describes the first two and latest 50 explicit
lookup boundaries; the reference above says latest 80. Preserve that disagreement.
Neither number changes this application's three-marker limit, and neither is a
grader expectation about the hidden server.
[Prompt-caching guide](https://developers.openai.com/api/docs/guides/prompt-caching).

For Generate Content, stable_v1 adds no marker field and creates no remote cache
resource. It reports implicit as the effective mechanism, with zero emitted
markers. The guide describes implicit caching without guaranteed savings and
model-specific input floors. No fixed number of cold turns follows from it.
[Generate Content caching](https://ai.google.dev/gemini-api/docs/generate-content/caching?hl=en).

Literal causality fixtures must cover the following projections with an ordinary
eligible text boundary at each bar. In the table, tools are absent and B is the
only base instruction. The current P is excluded from the reusable candidates.

| Projection | Selected distinct content boundaries |
|---|---|
| B \| H(one-shot) \| S(persistent) \| P | after B only |
| B \| S(persistent) \| P | after B and after S |
| B \| M(newest handoff) \| D(persistent dialogue) \| P | after B, M and D |
| B \| H(one-shot) \| M \| D \| P | after B only |
| B \| bound opaque O \| P | after B only; O unchanged |

For the split fixture, one ordinary block contains `BASE\nHINT`, with the captured
boundary after the LF and HINT marked one-shot. The new Messages blocks are
exactly these two members, preserving the separator and concatenation:

```json
[{"type":"text","text":"BASE\n","cache_control":{"type":"ephemeral"}},{"type":"text","text":"HINT"}]
```

Off retains the original single block. If the original is bound opaque, stable_v1
also retains it unchanged and emits no marker inside it. A later cache read is
not required to pass this deterministic fixture; correct causal placement is.

## 17.5 Capture the policy before rendering

Foreground, compression and judge requests need the same rule: derive policy
from the captured projection/configuration, render once, validate and durably
admit that exact prepared body, then hand it to transport and the diagnostic.
Adding markers after a helper's request_body has been recorded would make the
receipt describe a different request.

Select creation-only `--request-profile` in CLI/GUI or its public equivalent.
Without it, retain the exact v1–v6 selection rules. With it, construct a fresh v7
store or standalone log; an old store is session_incompatible, not an upgrade.
The flag adds no handler, catalog or enabled maintenance setting. First derive
base_version, 1 through 6, from the unchanged preceding selection rules. The v7
identity has exactly base_version, base and request_profile. Base is that version's
complete strict creation identity, with no version fields silently inserted.
Request_profile has exactly version:1 and capabilities: either null for an empty
selection or the complete strict immutable cache table from §17.6. Its rows are
part of creation identity, so offline validation resolves captured entry IDs
without reading a current file. Compatible resume preserves that selection.
Thus a simple plain Agent can select
v7 with base_version:1; recall still requires its explicit base_version:6 choice.

Session checkpoint/origin version and state_version are 7. Their exact outer
members remain Chapter 10's, including identity, as_of, state and state_sha256.
High_watermarks has event/request/activation/job, plus memory exactly when
base_version is 5 or 6. The semantic codec is {base,requests}: base is the strict
semantic state of base_version; requests has exactly version:1, profile,
config_revision, usage_receipts and capsules. Profile is the request profile
below; config_revision equals its last Actor-applied revision. Usage_receipts is
an admission-ordered array of {admission,purpose,execution,provenance,usage,
raw_usage,disposition}. Admission is the foreground/helper start sequence; purpose
is foreground/compression/judge. Disposition is applied or suppressed; the base
helper metadata additionally retains its original detailed outcome. Raw_usage is
exactly {source:"capsule",admission} for a retained foreground Responses capsule,
or {source:"inline",raw} otherwise. Inline raw is only the exact validated
usage-object JSON string, including for helpers and canceled foreground output.

Capsules contains only accepted foreground Responses records, ordered by admission.
Every record has state, admission, response, bytes and sha256. Response is its
response_ended sequence; bytes is the raw terminal payload's UTF-8 byte count and
sha256 hashes those bytes. State selects one strict variant:

| State | Additional required members | Meaning |
|---|---|---|
| retained | raw, refs | Exact payload string and complete checked §17.8 index list |
| retired | retired_by | The context_changed sequence that removed the complete bundle; no raw or refs member |

Admission/response are unique positive same-Agent sequences, with admission less
than response. Retired_by is greater than response and no greater than the snapshot
as_of. Snapshot base response facts replace the payload with exactly {capsule:{admission,response}}. This reference
must resolve to one matching variant. A retained variant supplies derived parts;
a retired variant supplies settlement/provenance metadata only, with no represented
conversation entry, call/result index or part reference surviving the cut. Its
accounting receipt retains execution, producing identity and usage. Retirement
converts raw_usage to inline before deleting raw and refs. Reject dangling, duplicate,
mismatched or state-inappropriate references; a retired record cannot serve a
request reconstruction. Hashes do not recreate missing bytes.

Helpers never enter capsules. Their full log retains the one original raw response
and its checked refs. Their semantic snapshot follows Chapter 15: exact config,
selected identities/versions, request/response hashes and sizes, source/output
measures, provenance, accepted usage and disposition, with output referenced only
through its represented memory ID/version. It contains no helper request_body,
raw response, submit arguments or second candidate text. Retiring that memory
removes the output reference and keeps its retirement metadata. Rejected or
suppressed candidates have no represented-output reference. Full-log replay can
use its receipts; a snapshot reports history_unavailable for absent source/output.
This also applies to judge receipts: retain accounting and the validated compact
verdict metadata, without keeping its raw response as a hidden body archive.

Suppressed foreground usage has no capsule or base response placeholder. Its
model_usage_recorded fact settles the exact outstanding admission as a canceled
response slot, applying the same pending-human discard as error_occurred while
preserving previously accepted dialogue and unconsumed hints. Already consumed
hints stay consumed. No second error_occurred may close that slot; turn_ended and
any pairing of previously accepted calls still follow Chapter 5. The fact neither
counts as an accepted model response toward the turn limit nor creates parts,
calls or effects. Snapshot base state records that settled slot and completion
state, with the inline suppressed receipt in requests. A duplicate or a fact
naming another/closed slot refuses before mutation.

All new arrays retain Chapter 10's item bounds and the whole-state budget. The
student documents the strict base codec as previously required; these substitutions
and transitions are fixed here. Validate base and extensions as one candidate
before installing any owner state.

Use the ordinary log header. Session_initialized.session has exactly version:7,
session_id and the complete v7 identity. Then emit the base
initializers in their unchanged order: optional skills_initialized, then
memory_initialized for base 5/6, then recall_initialized for base 6. Finally append
construction-only request_profile_initialized with payload requests containing
exactly version:1, base_version, identity equal to the complete v7 identity,
and profile equal to the validated initial request profile below.
A standalone constructor emits the same base initializer chain and final request
initializer without inventing session facts. This last initializer establishes
v7 offline capability; an exposed Agent cannot append it, and duplicates, missing
predecessors or mismatched identity refuse. Session anchors contain exactly version:7, session_id, origin_as_of, origin_sha256
and high_watermarks, bound to the v7 origin. Event envelopes and log_version
remain unchanged; session capability versions are distinct from log framing.
A partially initialized store is inspectable but cannot resume live.

A minimal plain, zero-handler construction fixture is below. The session ID is
synthetic. It establishes capability only; no model request has occurred.

```jsonl
{"log_version":1}
{"seq":1,"type":"session_initialized","time":"2026-01-01T00:00:00Z","session":{"version":7,"session_id":"11111111111111111111111111111111","identity":{"base_version":1,"base":{"mode":"plain","system":"BASE","skills":null,"handlers":[]},"request_profile":{"version":1,"capabilities":null}}}}
{"seq":2,"type":"request_profile_initialized","time":"2026-01-01T00:00:01Z","requests":{"version":1,"base_version":1,"identity":{"base_version":1,"base":{"mode":"plain","system":"BASE","skills":null,"handlers":[]},"request_profile":{"version":1,"capabilities":null}},"profile":{"revision":0,"surface":"inherited","funding":"api_key","connection":null,"binding":null,"generation":null,"cache_policy":"off"}}}
```

A standalone counterpart omits session_initialized and numbers the request
initializer 1. An anchor-origin log instead uses its validated v7 origin and
session_anchor; it does not repeat construction initializers. Missing base memory
or recall initialization is invalid even when the final request initializer is
well formed. Header-only and half-constructed files confer no live capability.

Every v7 turn_started.turn adds execution, and every foreground request_sent and
helper config adds an execution capture. Route/model/funding/binding/cache are
copied from the admitted turn; purpose-specific delivery and caps follow §17.8.
An idle compression operation captures its own selection. The exact object:

```json
{"version":1,"route":"openai-responses","requested_model":"fixture-responses","resolved_model":null,"funding":"chatgpt_plan","binding":"0123456789abcdef0123456789abcdef","generation":1,"delivery":"stream","cap":{"mode":"absent","tokens":null},"cache":{"version":1,"policy":"off","mechanism":"unchanged","capability":null}}
```

Route is anthropic-messages, openai-chat, gemini-content or openai-responses and
must match captured vendor/surface/model. Requested_model is a nonempty routing
identifier and resolved_model is null or the explicit exact comparison identity;
no learned alias is substituted. They are captured for the whole turn. Funding
is api_key or chatgpt_plan; only openai-responses permits the latter. Binding is a 32-lowercase-hex random
registration binding for plan, null for API-key; generation is positive uint64
for plan, null for API-key. These are nonsecret local identities, not account or
OAuth client IDs. Delivery is plain/stream, with Responses always stream. Cap has
exactly mode and tokens: remote plus a positive integer equal to the actual
remote cap, or absent plus null only for plan. In every v7 captured model config,
including foreground and
helpers, max_tokens is null exactly for absent; old versions still require their
original integer. Compression clamps its derived remote cap to 8,192, judge sets
256 and plain delivery on legacy adapters; Responses helper delivery is stream.
The judge's system and verdict contract remain unchanged.

Cache has exactly version:1, policy, mechanism and capability. Policy is off or
stable_v1. Mechanism is unchanged, messages_manual_5m, chat_implicit_30m or
gemini_implicit. Capability is null for unchanged/Gemini, otherwise an immutable
verified table-entry ID. Off means unchanged; stable_v1 must match the selected
surface. Responses stable_v1 refuses cache_capability locally. It may still
benefit from the provider's implicit caching. Old captures without execution or
cache retain their old meaning; v1–v6 reject the new fields.

Agent's current request profile has exactly revision, surface, funding, connection,
binding, generation and cache_policy. Revision is uint64 initially 0; surface is inherited or responses;
Funding is api_key/chatgpt_plan; connection is null for API-key or a safe selected
local name for plan; binding/generation follow execution grammar and must match
that mounted registration. Cache_policy is off/stable_v1. Validate combinations before
application. Public replacement uses expected revision and one whole profile;
conflicts and no-ops follow the existing revision rules. Actor applies it without
I/O, and already admitted turns keep their prior capture. The selected named
connection must be mounted and validated before this update. Local auth listing supplies its nonsecret binding/generation
to authorized public callers; human route commands resolve them by name. A
profile claiming another pair refuses without writing. Setting inherited
requires api_key. These execution settings are Agent-owned; their physical
credential-store path is application creation configuration.

The complete policy file remains Chapter 16's version 4: caching and funding do
not change max_model_requests/context/memory/recall policy semantics. Request
profile is explicit launch/public configuration, not another policy writer.
Resume supplies a compatible current profile and explicit registration binding;
historical captures are never filled from today's defaults. A new profile may
change a future route but must refuse incompatible still-represented opaque data
before dispatch. It cannot strip that data to make the change appear possible.

In v7, request_profile_initialized and request_profile_changed retain the
inherited ordinary record class for their session/standalone mode. Changed
payload requests has exactly base, revision and
profile; base is previous revision, revision its checked successor, profile the
complete applied object. Configuration is current execution state, not creation
identity. Responses response_ended and terminal usage facts use a 64 MiB physical
record class, recognized only after valid v7 initialization. Preserve Chapter
10’s 64 MiB limit for all session events, the base standalone record classes
and the 256 MiB semantic-state limit. A base-1 v7 profile does not acquire
memory/recall facts or their standalone input classes. Bound original encoded
records and escaping before allocation, exactly as Chapter 10 requires.

Snapshot reconstruction uses only available recorded sources. A retired body
returns history_unavailable; a diagnostic is no hidden archive. Resume clears
pairs and starts a new mount baseline without HTTP or model discovery. Required
fixtures include all six base identities wrapped in v7, exact old rejection of
new fields, missing/out-of-order initializers, memory watermark conditionality,
wrong plan binding, profile changes between turns and pure reconstruction under
a different current default. Retain exact old requests on all v1–v6 routes.

## 17.6 Read the meter's label before its number

An earlier version of the meter once showed a paid model with zero cost. The
model selection had come from the environment; the meter consulted an empty
settings field. The producing identity was available the whole time, but the
display asked the wrong owner. Labeling the cost "unknown" would have been more
honest than "zero"; using the actual producing identity was the useful repair.

Keep normalized, disjoint input/write/read/output counters on Engine. The current
model selection cannot reprice yesterday's responses. Each accepted fact retains
its producing model/surface, purpose and effective rate-affecting options. A rate
table is immutable selected data with an ID, source date, currency and exact
rational rates. The actual demonstration must supply a cited table for its chosen
API-key identities. The constructor accepts that table without installing guessed
prices; an omitted table reports unknown rates.

| View | Start and membership |
|---|---|
| Durable | Every accepted usage fact retained by the session, including helpers |
| Current mount | Baseline after successful restoration and before the first new operation |
| Purpose | Foreground, compression or judge subset of either named scope |
| Last accepted | One identified accepted response, even if its output was unusable |
| Latest attempt | Current attempt identity/outcome, with usage absent until accepted |

A failed construction has no mounted run. Closing Agent A cannot reset B's
baseline. Resuming A preserves durable totals and starts a zero current-mount
view; a later response increases both. Diagnostic comparison failure has no
accounting effect. Replay folds accepted usage once, never by counting displayed
fragments or observing the same event twice.

The Chapter 2 normalization remains binding. Messages reports separate ordinary,
write and read input. Chat Completions subtracts reported writes and reads from
its inclusive prompt total. Generate Content has no cache-write bucket to invent.
Keep the complete accepted raw usage as evidence; unknown after an incomplete
response differs from a reported zero. Pricing a supported new write duration
would require retaining that duration before aggregating it with another rate.
This chapter adds only Messages' selected default duration and Chat Completions'
selected 30m mode, without premium tiers or a remote-cache storage tariff.

These synthetic rates exist only to test arithmetic. They are not vendor prices.
All amounts are USD per one million tokens:

| Fixture rate row | Ordinary input | Cache write | Cache read | Output |
|---|---:|---:|---:|---:|
| fixture-rate-a | 2 | 5/2 | 1/5 | 10 |
| fixture-rate-b | 4 | 5 | 2/5 | 20 |

For accepted usage input 100, write 20, read 30, output 40 under rate A, the
exact cost is `(200 + 50 + 6 + 400) / 1000000`, or 0.000656 USD. The cache-read
share is 30 / (100 + 20 + 30) = 1/5. A second identical response under rate B
adds 0.001312 USD, making the mixed-model total 0.001968 USD. Merely switching
the current selected model afterward changes neither sum.

Accumulate exact rationals, reduce before storage/display where appropriate,
and round once at presentation to six decimal places with ties to even. Two
half-microdollar contributions sum to one microdollar; rounding each to zero
first is wrong. Display exact numerator/denominator in machine/public inspection
so a small nonzero cost remains distinguishable from zero after formatting.

All source counters accept full uint64. Checked aggregation that exceeds uint64
reports counter_overflow with the affected scope unavailable; it cannot wrap,
clamp or erase accepted facts. Intermediate arithmetic must not overflow simply
because a rate multiplies a valid counter. Use exact integer/rational operations
with a published 4,096-bit intermediate ceiling; reaching it yields price_limit
and an incomplete estimate, while valid usage stays visible.

The zero-input denominator reports cache_read_share unavailable. The literal
counter 9007199254740993 must survive the browser unchanged as a decimal string.
Adding one to 18446744073709551615 must report overflow. An unknown rate alongside
rate A's example reports known_subtotal 0.000656 and complete false, with the
unpriced identity listed. It must never show that subtotal as the entire bill.

An estimate over accepted usage also excludes any unreported usage from failed
attempts. Keep a separate incomplete-attempt count. Even a complete rate lookup
cannot turn this local sum into a verified invoice or a subscription charge.

### Immutable capability and price inputs

Expose creation-only --cache-table FILE and --price-table FILE, with equivalent
public owned values. Missing files are not guessed; absent options select empty
tables. Read each at most 256 KiB, with strict scalar JSON, no duplicate/unknown
members and at most 256 rows. Do not read these files during reconstruction.
Freeze their parsed values, IDs and content digests for the mount; no global
mutable catalog or automatic network price lookup is introduced.

Cache table has exactly version:1, id, as_of and entries. Each entry has exactly
id, vendor, surface, model, mechanism, source and verified_on. IDs match the safe
name grammar; dates are YYYY-MM-DD; source is an HTTPS documentation URL at most
2,048 UTF-8 bytes. All other strings are nonempty and at most 256 bytes. Mechanism
is messages_manual_5m or chat_implicit_30m, with its matching surface. Rows are
unique by both ID and vendor/surface/model; matching is exact against the supplied
resolved identity. Selecting stable_v1 without one refuses cache_capability, except
Generate Content's implicit path. No Responses marker row is accepted here.

Price table has exactly version:1, id, as_of, currency:"USD" and entries. Each
entry has exactly vendor, surface, model, mechanism, input, write, read, output,
source and verified_on. Four rates are rational objects in USD per million tokens,
with numerator/denominator canonical decimal uint64 strings and positive denominator.
Mechanism uses the execution-cache enum. Rows are unique by the four identity fields
and require API-key funding. Validate fractions, bounds and dates before any Agent
is exposed. Synthetic fixture rows may use the explicitly fictional model names
from this chapter; live tables must carry primary-source receipts for actual rows.
A schema can validate syntax, not establish that a price is current.

At accepted usage, retain the effective cache mechanism and producing identity.
These determine lookup, never the current selected model or a sorted name prefix.
A table selected for the mount supplies a labeled current estimate over durable
usage; its ID/date remain visible when another table would price history differently.
No claim of the historical invoice follows. Plan usage always appears as unpriced
with reason subscription_not_invoice, rather than applying an API rate silently.
Complete is false for any unpriced identity, incomplete attempt or arithmetic limit.
Unpriced reasons are unknown_rate, subscription_not_invoice, counter_overflow or
price_limit. On the rational ceiling, preserve the last representable known subtotal
and mark the omitted contribution price_limit; do not wrap or silently discard it.
Test mismatched model, mechanism, date syntax, duplicate row and zero denominator
independently. The complete live evidence must retain the actual immutable files.

## 17.7 Put the explanation where the user can see it

Add a Cache and usage view to the existing optional GUI and a `/cache` inspection
command to human chat. Both use the same copied public status. Preserve `/usage`
and the existing protocol replies. No automatically installed Registry tool or
new Skills grant is needed. The model can inspect the display through the
already granted GUI MCP tools when that integration is selected.

The wire status has exactly version, agent, mount, revision, profile, attempts,
usage, pricing and omitted. Version is 1; all fields below are required. Define U
as a canonical decimal uint64 string: 0 or a nonzero digit followed by digits, within
range. Define Hash as 64 lowercase hex characters. Private numeric structs may
use integers; browser JSON must retain these strings without Number conversion.
Agent uses its existing safe public identity. Mount/revision are U. Turn is the
unchanged opaque Chapter 5 request ID, and operation the unchanged nonempty
Chapter 6 operation ID; neither is parsed as a number. Preserve their validated
UTF-8 strings and inherited bounds. Profile is a coherent copy of the current
§17.5 profile, with its revision/generation encoded as U (generation may be null).
It is captured under the same status revision as the attempt rows. Protocol
request_profile/route values use this same wire-profile form; durable profiles
keep their §17.5 integer encoding.

| Object | Exact fields and values |
|---|---|
| attempts row | purpose (foreground/compression/judge), latest (Attempt or null), pair (Pair or null), code (safe reason or null) |
| Attempt | admission (U), turn (opaque request-ID string or null for idle), operation (opaque operation-ID string), route (the §17.5 enum), requested_model (nonempty string), resolved_model (string or null), funding (api_key/chatgpt_plan), generation (U or null), state (admitted/sent/accepted/failed/canceled), bytes (U or null before handoff), sha256 (Hash or null before handoff) |
| Pair | before (Attempt), after (Attempt), raw (Raw), structural (Diff), directive (Diff), markers (Markers) |
| Raw | prefix (U), before (byte or end), after (byte or end); byte is {kind:"byte",value:integer 0..255}, end is {kind:"end_of_body",value:null} |
| Diff | status (equal/appended/edited/unavailable), added (U), paths (array of JSON-pointer strings), sections (array of Section), code (null or safe reason) |
| Section | name (adapter section or controls), status (same enum), added (U) |
| Markers | before (U), after (U), skipped (array of {reason,count:U}) |
| usage | durable (Scope), mounted (Scope), last_accepted (Receipt or null), incomplete_attempts (U) |
| Scope | total (Counters), purposes (three {purpose,counters} rows in fixed purpose order), cache_read_share (Rational or null), share_code (null, no_input or counter_overflow) |
| Counters | input, write, read, output (each U, or all null on overflow), code (null or counter_overflow) |
| Receipt | admission (U), purpose, producing (vendor/model/surface triple), funding, counters (Counters) |
| pricing | table (ID or null), table_date (YYYY-MM-DD or null with absent table), currency:"USD", known_subtotal (Rational), displayed_subtotal (six-decimal string), complete (boolean), unpriced (array of {producing,funding,reason}), incomplete_attempts (U) |
| Rational | numerator (canonical nonnegative decimal string), denominator (canonical positive decimal string), reduced fraction within §17.6 bounds |
| omitted | paths, skipped, unpriced (each U) |

Attempts always has three rows in purpose order. Empty Scope has zero counters, null cache_read_share and share_code no_input;
empty attempts have null latest/pair and no_baseline. Added counts whole appended
array elements and is zero for non-appended results; paths identifies edited
locations without request values. Sections follows the adapter's order in §17.3,
then controls. Excluded marker spans remain internal/export metadata, not request
text in ordinary status. Safe reasons are fixed application codes: no_baseline, body_limit, analysis_limit,
diagnostic_busy, invalid_json, retired, closed, unsupported_location, empty_text,
opaque, transient_prefix, duplicate, slots and the explicitly named errors below.
They contain no provider body or request text. No keys are selectively omitted
to disguise an unavailable value. Model strings are limited to 256 UTF-8 bytes in this status; a larger valid
provider identity makes that status operation fail with model_identity_limit and
no partial object, rather than truncating it into another model name.

Fit the 256 KiB status bound by omitting tail paths, then skipped-detail rows,
then unpriced rows, incrementing the corresponding omitted counts. All identities,
counters, three attempt rows and availability flags remain. Never split a UTF-8
string or integer. If required metadata alone exceeds the bound, return the safe
error diagnostic_status_limit; no partial success object is published. Status
revision advances only for changed copied contents. Snapshot-plus-tail watches,
mount fencing and overflow/resynchronization follow Chapter 7. Closing a view
only detaches its subscription. Browser has no second usage accumulator.

Human `/cache` prints current route/funding, the three comparison rows, named usage
scopes, price completeness and incomplete attempts. `/cache policy off|stable_v1`
applies a request-profile replacement against its displayed revision; it is a
local command, never model text. `/cache export PURPOSE DIR` accepts only the three
purpose names and a new directory. Success prints the manifest path; missing pair,
busy, retirement or disk failure prints the safe reason. Existing chat command
parsing/escaping applies. `/route` prints the selected profile; `/route inherited
api_key` and `/route responses api_key|chatgpt_plan [NAME]` replace selection for
future turns only, with NAME required exactly for plan. The human prompt names
plan mode and its absent remote cap. Initial flags are --request-profile,
--surface inherited|responses, --funding api_key|chatgpt_plan, --connection NAME,
--cache-policy off|stable_v1 and --connections-dir DIR. New selectors without
request-profile refuse before Agent construction; defaults are inherited/api_key,
null connection and off.

Protocol inputs are {kind:"cache",id}, {kind:"cache_export",id,purpose,dir},
{kind:"request_profile",id,expected,profile} and {kind:"route",id}. Id follows the
existing caller-ID grammar; expected is U. Output is exactly {kind,id,ok,value,
code}; kind echoes the input. On success code is null; value is Status for cache,
{manifest:PATH} for export, or the complete profile for request_profile/route.
On error value is null and code is a safe reason. Malformed controls retain the
inherited continue/error policy, without paid work. Public embedding exposes the
same operations and copied values; private Go method spelling remains free.

The GUI Cache and usage panel uses that status, labels current-mount versus durable
and plan versus API-key estimates, and exposes the two cache policy choices only
where available. No automatic body download or credential selection is hidden in
reconnect. Public export completion and profile replacement use reliable request
completion, not a display observation. Export manifest has exactly version:1,
before and after (Attempt), comparison (Pair), and files containing exactly two
{name,bytes,sha256} objects for before.json and after.json. Its attempt data and
hashes must agree; publishing it is the success boundary described in §17.2.

A two-Agent fixture should accept rate A's example on Agent A, then construct B,
accept rate B's example on B and fail A's next attempt before accepted usage.
A still shows durable/current 0.000656, its earlier accepted response and the new
failed attempt separately. B shows 0.001312. Closing and resuming A resets only
A's current-mount amount to zero. Stale A callbacks cannot overwrite B or the
replacement A. This test catches a shared counter that a one-window demo misses.

For the ID fixture, an attempt with turn "r1" and operation "op-alpha" must round-trip
unchanged. Before the first request, all three latest/pair fields are null while
profile still reports inherited/api_key. Change profile to responses/chatgpt_plan
at revision 1 before sending: status must show that selection even with no attempt.
After a prior API-key attempt, the same change leaves that attempt's funding intact.
Mount table "rates-a" dated "2026-01-01", then restore with "rates-b" dated
"2026-02-01": pricing.table and table_date change together; producing usage does not.

## 17.8 Change the funding route without changing the experiment

Comparing a subscription request with a differently rendered API-key request can
produce a difference before either reaches a cache. The control in this chapter
uses the same Responses surface and history mapping. It still has an output cap,
so the comparison must disclose that difference. A common endpoint does not make
two accounts, model offerings or service conditions identical.

Bill approved an explicitly selected ChatGPT-plan mode without a per-response
output-token cap. Finite request counts, deadlines and local received-byte limits
remain. They cannot guarantee a remote token or charge ceiling. Existing API-key
surfaces keep their behavior, including the plain 256-token recall judge. The
new API-key Responses control uses streaming with the applicable remote cap.

### Own the connection before lending a credential

Ensemble owns Connections in its own implementation spoke; Connections owns each
Connection; Connection owns refresh work and credential leases. An Engine-owned
request operation reaches it through Engine → Agent → Ensemble. Common declares
shared values and parent interfaces. Agent owns selection, Engine owns usage,
and the GUI receives only safe status. CLI and GUI attached to one Ensemble share
its service; a second process cannot independently rotate the same refresh token.

Require creation-only `--connections-dir DIR` for plan authorization or use.
Resolve it once. The directory must be outside the session, workspace and exports;
create it owner-only, reject symlinks and unsafe ownership/permissions, and acquire
an exclusive OS-backed process lock for the entire mount. Refuse contention as
connection_store_in_use, without truncating anything. Crash releases the lock;
a stale lock filename alone is not evidence of ownership. Existing API-key-only
applications need no connection store or browser dependency.

Connection records have an immutable random 32-hex binding and positive uint64
generation, saved atomically with the protected registration. Bind the record to
verified issuer, subject and issued client ID. Keep its display name separate.
Normal refresh and same-registration reauthorization preserve binding/generation.
Explicit sign-out increments generation before invalidation is acknowledged;
a newly authorized registration gets a fresh binding, even if its display name
is reused. Reject generation exhaustion. Restart loads this protected mapping;
it never reconstructs identity from a label, email or a process-local counter.
A restored session explicitly selects its matching binding/generation before live
plan use. A sign-out generation change is deliberate: start a new session or use
the explicitly authorized whole-bundle handoff below to remove incompatible
bound material. V7 bases 1–3 have no such capability: keep a compatible registration
or start a new session. V7 does not grant them context maintenance. Never
automatically strip output or impersonate the retired generation. Missing,
signed-out or different bindings refuse connection_mismatch.
Pure offline reconstruction needs no credential store.

The protected file format is implementation-owned, documented and bounded to
1 MiB per registration and 64 registrations per store. It contains the registration,
validated identity, binding/generation, token set and expiry, never conversation
history. Persist a stable random host ID separately before first sign-in. Atomic
replacement includes fsync/rename/directory-sync under the inherited file rules;
an uncertain rotation or persistence outcome marks the connection unavailable
until explicit reauthorization. Do not issue a lease from an unsaved replacement.
No secret enters argv, logs, source control, session state or GUI messages.

The application offers `auth login NAME --connections-dir DIR`, `auth list` and
`auth logout NAME` with the same required directory option. NAME matches
`[a-z][a-z0-9_-]{0,63}`. Login prints a safe waiting status and opens the browser;
never print an authorization URL containing a retained ID-token hint. All commands
have public equivalents. List returns copied {name,binding,generation,state}
rows, sorted by name; state is ready, refreshing, plan_disabled, logged_out or
unavailable. It exposes no expiry claims, tokens or account identifiers. An empty
store returns an empty list. These auth commands construct no Agent and are
exempt from the --request-profile requirement for Agent route flags. Auth is a local operator action, not a model tool or GUI
MCP permission. The browser cache view need not implement a second OAuth client.

For this public-client flow, start a loopback callback listener before opening
`https://auth.openai.com/api/accounts/authorize`. Request these parameters:

| Parameter | Value |
|---|---|
| client_id | dynamic_agent_client initially; saved issued ID on return |
| agent_name_hint | Ensemble, initial registration only |
| ext_agent_host_id | Saved host ID |
| response_type | code |
| redirect_uri | Exact `http://127.0.0.1:PORT/auth/callback` |
| scope | `openid profile email offline_access resource.invoke chatgpt.tokens.use.direct` |
| resource | `https://api.openai.com/v1` |
| state, nonce | Fresh cryptographically random attempt values |
| code_challenge, code_challenge_method | PKCE S256 challenge; S256 |

Exchange a validated callback code at
`https://auth.openai.com/api/accounts/oauth/token` using form fields grant_type
(authorization_code), issued client_id, code, code_verifier, redirect_uri and
resource. No client secret or another application's token store is involved.
[Registration guide](https://developers.openai.com/siwc/token-sharing-open-source/sign-in).

Use maintained OIDC/JWT verification with discovery at
`https://auth.openai.com/.well-known/openid-configuration`: validate issuer,
JWKS signature, issued-client audience, expiration and saved nonce. Returning
identity must match the selected registration. Check granted plan scope separately;
identity-only success reports plan_disabled and permits no inference.
[Identity validation](https://developers.openai.com/siwc/website),
[permission errors](https://developers.openai.com/siwc/token-sharing-open-source/errors-and-recovery).

Validate token_type Bearer, nonempty bounded access/refresh tokens and granted
scopes before publication. Access-token contents remain opaque credentials.
A refreshed ID token is verified against issuer/audience/selected subject and
expiry; refresh has no newly generated authorization nonce to compare. Never
accept a different registration because its display email happens to match.
A failed verification leaves the candidate unpublished.

Local limits are ten minutes for one authorization attempt, 30 seconds per auth
HTTP operation, 64 KiB callback input and 1 MiB decoded discovery/JWKS/token reply.
Redirects to a different origin are refused; token POSTs do not follow redirects.
State mismatch, duplicate callback parameters, wrong path, denial, missing new
client ID or changed returning client ID refuse before token exchange. A callback
settles once and closes its listener. Fixtures mutate each condition independently
from one valid flow; an expired attempt cannot replace a later registration.

Refresh near expiry using the saved issued client ID and current refresh token,
with grant_type refresh_token and the same resource, omitting scope. Rotation
replaces the whole token set atomically. Explicit logout attempts revocation using
the discovered revocation endpoint, then clears local tokens and reports whether
remote revocation was confirmed. Application close is not logout.
[Accounts and sessions](https://developers.openai.com/siwc/token-sharing-open-source/profiles-and-sessions).

Validate expires_in as a positive integer duration no greater than one day; compute
expiry from token-response receipt using checked arithmetic and monotonic elapsed
time while running. After restart use the saved wall expiry with a 60-second safety
margin; a clock anomaly refuses renewal-needed admission rather than extending a
lease. Start refresh when remaining time is at most 60 seconds. The documented
access lifetime fits this bound. Retain earliest_refresh_at as protected provider
metadata without inventing semantics the documentation does not define.
[Token reference](https://developers.openai.com/siwc/token-sharing-open-source/token-reference).

### Cancellation belongs to the caller; refresh belongs to Connection

Actor captures execution before starting an asynchronous Engine operation. That
operation waits for a lease under its existing deadline; neither it nor Actor
holds a service lock across browser, network or disk work. Connection owns one
refresh context/result and at most 256 waiting lease acquisitions. A further
waiter receives connection_busy without paid admission. Completed/canceled
waiters are removed; token bytes are owned by the Connection and bounded leases,
never copied into unbounded completion collections. Canceling waiter A removes A only; waiter B can receive the
replacement. Sign-out or Ensemble close cancels shared work. Late refresh completion
cannot repopulate a cleared generation. Failed refresh is visible and performs
no model retry. Waiters receive bounded safe reasons, not raw token replies.

Dispatch registration and generation invalidation share one atomic ordering point.
If invalidation wins, no transport starts. If dispatch registration wins, invalidation
cancels the registered operation and prevents further dispatches, but cannot recall
already handed-off bytes. Release locks before HTTP. A lease acquired earlier is
insufficient authority without this dispatch check. Agent close joins only its
operations; Ensemble close joins Connections work and releases the store lock.

Terminal settlement is a second boundary, owned by Actor. The operation has a
single bounded terminal-result slot. If a complete valid response reaches that slot
before Actor settles cancellation, account for its accepted usage once, even if
cancellation prevents applying output. After settlement, a late result is discarded
and the attempt remains incomplete. Never wait indefinitely for a terminal response
merely to improve accounting. Ordinary accepted foreground output uses response_ended;
accepted usage with suppressed foreground output instead uses model_usage_recorded.
Its usage payload has exactly admission, execution, provenance, usage, raw_usage and
reason; reason is canceled, provenance is the producing triple, and raw_usage is the
validated original usage-object JSON string. It names one unsettled foreground
admission and may occur once instead of response_ended. It creates no conversation
part, pending tool call or new effect. Helper accepted/failed facts retain their
own once-only path and canceled output disposition. Replay verifies uniqueness
across both paths, so duplicating an observation cannot charge it again.

Use barriers, not sleep guesses: invalidate before dispatch; invalidate after
registration before HTTP; cancel A during a shared refresh with B; delay refresh
persistence across sign-out; place a valid terminal result immediately before and
immediately after Actor cancellation settlement. Assert actual dispatch count,
accepted usage count, no canceled tool/memory/recall effects and survivor progress.
An already durable accepted result is never rolled back by a later logout.

### The Responses request and its local authority

Discover plan models through authenticated GET `https://api.openai.com/v1/models`
using the selected Connection. Preserve the returned models array's ordering,
slug and display_name. API-key discovery keeps its ordinary catalog shape. Cache
no account discovery result under another binding. A selected slug must have been
listed for that connection during explicit discovery or this mount; no lexical model-name inference or silent replacement
is allowed. Availability does not prove every tool/content capability.
[Models and inference](https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference).

Both funding modes POST to `https://api.openai.com/v1/responses`. Emit fields in
this order: model, store:false, stream:true, max_output_tokens only for API-key,
input, then tools when nonempty. Use compact JSON, no final LF; ordinary newly
constructed string escaping follows the existing canonical string encoder.
Input is the complete current authorized projection, not a server conversation.
Primary instruction becomes a developer message; hints/manuals/recall remain user
data in their inherited positions. No explicit system message, previous_response_id,
conversation, temperature or unsupported plan parameter is emitted.
[Preview limitations](https://developers.openai.com/siwc/token-sharing-open-source/preview-limitations).

Ordinary messages use type:message, role and content arrays of input_text blocks.
A portable historic assistant, including a working note, uses the documented
assistant input_text envelope; it has no invented provider id/status. Real
Responses assistant items follow the capsule rule below. This complete plan request
is a separate positive control for a portable working note:

```json
{"model":"fixture-responses","store":false,"stream":true,"input":[{"type":"message","role":"developer","content":[{"type":"input_text","text":"BASE"}]},{"type":"message","role":"assistant","content":[{"type":"input_text","text":"[working note]\nPort is 9090.\n[/working note]"}]},{"type":"message","role":"user","content":[{"type":"input_text","text":"Continue"}]}]}
```

[Responses input schema](https://developers.openai.com/api/reference/resources/responses/methods/create).

This chapter's new Responses adapter supports text dialogue and local functions. A human blob has
no newly taught attachment mapping and refuses unsupported_reference before
HTTP; retaining it in history does not authorize fetching a local path or URL.
Do not invent a data-URL field on Chapter 2's reference-only neutral part.
Tool-result children retain the existing descriptive mapping: text unchanged,
URI blob as `[<mime>] <locator>`, redacted child as its stub followed by a space
and locator when present. Join children with LF into function_call_output.output;
local path/handle blobs retain their unsupported-reference refusal. Preserve the
original call_id and results-before-deferred-human ordering. These are explicit
surface limitations, not a claim about every content type Responses can accept.

Declare local functions inside one namespace named ensemble. Each function keeps
name, description, parameters and strict:false. Helpers have only namespace memory
with submit_memory; the judge has no tools. Do not inherit provider-hosted MCP,
search or shell merely because Responses recognizes their types.
[Tools](https://developers.openai.com/api/docs/guides/tools),
[function calling](https://developers.openai.com/api/docs/guides/function-calling).

Reject duplicate members in the terminal/event/item structure; the argument-string
exception below never relaxes that rule. A completed function_call requires
nonempty unique item id and call_id, nonempty
name and a bounded syntactically valid JSON-object arguments string. Preserve the
decoded string's exact bytes. Duplicate members inside that object retain the
controlled Chapter 10 tool-argument error; malformed/non-object arguments invalidate
the response. Missing or foreign namespace is a controlled call refusal with
is_error:true and exact text `tool namespace is not permitted`, paired to the
original call_id. It consumes any pending one-shot tool limit once, executes
nothing, then permits later calls/continuation. An unknown function within ensemble
uses the inherited unknown-tool result instead. A helper namespace/name mismatch
is an unusable helper response, never a foreground tool invocation.

### Keep the received item, not a plausible reconstruction

Responses output can contain reasoning, commentary and a call before an answer.
Concatenating the text loses both order and phase. Retain one immutable raw terminal
response JSON payload. Its output array is authoritative; do not independently
persist copied text/call parts that can disagree with it.

For Responses only, v7 response_ended.response gains responses with exactly raw
and refs. Raw is the exact terminal response-object JSON string, excluding the
outer SSE event. Refs is the ordered index list described below; response.parts
is empty. Ordinary transient/public parts are derived owned views. Existing
provenance, usage and raw_usage must agree with raw. While retained, the foreground
snapshot stores only a checked raw_usage reference into that payload; retirement
converts it to the inline usage object under §17.5.
In a full log, a helper already has response.raw, so its new responses member
has refs only. It references that same payload. Semantic snapshots instead use
§17.5's compact helper metadata and represented-memory reference, with no raw
helper capsule. No second raw response or item archive exists.
The v7 base semantic codec allows these explicit response substitutions and applies
all old transition rules to the derived views, not to an independently writable
shadow conversation.

Each ref has exactly item, content and kind. Item is a zero-based output index;
content is a zero-based message content index or null for a whole item. Kind is
text, call or opaque. Message content output_text creates a text ref even when
empty; refusal or unknown content creates an opaque ref. Function_call creates
one call ref. Reasoning/unknown output creates one opaque ref. Preserve array
order, all positions and exact coverage, rejecting duplicate/out-of-range/omitted
refs. Message role must be assistant and known fields must have valid types;
unknown item types are bounded opaque, never an executor. Require at least one
ordinary text position or call for foreground acceptance under Chapter 2. A
refusal-only result is a safe response error. Helpers instead apply their specific
output grammar after accepting otherwise valid usage.

Capsule replay copies the original raw output-item spans into input in order,
without dropping id/status/phase/annotations, compacting argument strings or
re-emitting opaque lexemes. These are permitted input item fields. Preserve a
message's phase when present. Pass reasoning encrypted_content without decoding
it. Supported replay item kinds are message, function_call and reasoning. Known item
fields retain their validated provider shape; an unrecognized extra field is
preserved but refuses replay unless the documented input shape permits it.
Unknown output items are retained but cause unsupported_response_item at the
next render unless their replay mapping has been explicitly taught; retention
alone is not permission to guess a new input shape.
[Responses reference](https://developers.openai.com/api/reference/resources/responses/methods/create),
[conversation state](https://developers.openai.com/api/docs/guides/conversation-state),
[reasoning](https://developers.openai.com/api/docs/guides/reasoning).

Bound the capsule to the accepted response's existing aggregate budget; offsets
are validated against its owned bytes before any dereference. The whole output
bundle and its paired results are a protected unit, even when it contains only
reasoning and final text with no call. Chapter 14 automatic/selective removal skips
it with responses_bundle. Chapter 15 compression protects every segment containing
one; a derived text view is never an unsigned compressor source.

### Retire the complete bundle at an authorized handoff

An account change cannot quietly turn signed history into ordinary text. In v7
bases 4–6, an explicit Chapter 14 handoff can make a **zero-model cut** of the
complete settled bundle. The cut itself invokes no helper and takes no bundle
as compressor input. Existing separately enabled memory scheduling still requires
its own eligible sources and captured authority. Bases 1–3 and all old versions
keep their existing capabilities and strict schemas.

Only v7 bases 4–6 extend context_changed.context with the required member
remove_responses. It is an array of exactly {admission,response}, positive durable
sequences identifying a currently retained foreground capsule and its accepted
response. Sort by response sequence, with no duplicates. Auto requires []. Handoff
requires the complete set of represented capsules, including no-call responses
and the handoff-producing response if it used Responses. Every call in each must
have its accepted result; any unresolved call or outstanding response slot refuses
the handoff with context_busy. A partial, foreign, already retired or mismatched
target list refuses before persistence. Remove_batches still names every completed
batch required by Chapter 14, including batches belonging to these capsules; the
two lists describe one atomic cut, never two removals of a result.

The Actor builds and validates this typed candidate using the admitted intent's
captured enabled policy, or the idle public request's applied policy. JSON naming
a target confers no authority. The existing intent cancellation, revision/base,
note-size and persistence rules remain. A successful cut changes each capsule to
its retired variant, removes all its derived parts and paired results, and settles
the intent once. Other user, manual, hint and portable assistant entries keep their
anchors and relative order. Insert the single working note at the inherited
post-batch/public boundary, before the pending human prompt; consumed hints retain
their normal next-request behavior. Retire affected diagnostic pairs and pending
exports at the same accepted transition; they cannot preserve another copy.

For a valid v7/base-4 prefix through sequence 49, let admission 6/response 7 contain
reasoning, commentary and two calls whose completed batch is 7. Let admission
20/response 21 contain reasoning plus final text and no call. These are the only
represented capsules; intent 42 is an authorized portable handoff in completed
batch 31. Its exact event at sequence 50 is:

```json
{"seq":50,"type":"context_changed","time":"2026-01-01T00:00:50Z","context":{"action":"handoff","intent":42,"base":49,"policy":{"revision":2,"enabled":true,"target_bytes":400000},"stub_calls":[],"remove_batches":[7,31],"remove_responses":[{"admission":6,"response":7},{"admission":20,"response":21}],"note":"Port is 9090."}}
```

This is a transition fixture with the stated prefix, not a standalone log.
Construct both bundles from the terminal examples below. Append this second call
to resp-1's output array and pair call-1/call-2 with text results note/other in
that order; the first response's reasoning/commentary stays intact:

```json
{"id":"f-2","type":"function_call","status":"completed","namespace":"ensemble","call_id":"call-2","name":"read_file","arguments":"{\"path\": \"other.txt\"}"}
```

Prepend this reasoning item to resp-2's output array:

```json
{"id":"r-2","type":"reasoning","summary":[],"encrypted_content":"NEXT"}
```

The stable targets are event identities, never provider IDs. If their positions
are A→bundle7/results→bundle21→U→H→S→batch31/results→P,
where A is older human text, U the later handoff turn's human input, H an unconsumed
hint, S a skill manual and P a pending human prompt, the handoff leaves
A→U→H→S→N→P. After the request consumes H, it leaves A→U→S→N→P. Neither bundle's
commentary/final text survives beside N.

Run the same transition from the full log, a checkpoint at 49, and that checkpoint
plus the exact tail event. Reduced state, next request and usage must agree.
Omitting target 21, substituting admission 19, retaining a part ref to 7, or omitting
one of batch 7's results must refuse without changing state. An automatic candidate
with either target refuses; an automatic cut elsewhere leaves both raw bundles
byte-identical. Repeat on base 1: the handoff capability is absent. No account
switch relaxes these rules.

For Chapter 15 pressure, each represented Responses response contributes one
assistant message whose parts array contains exactly one measurement-only object
{type:"responses_bundle",raw_output:STRING}. STRING is the exact original JSON
output-array span, including its brackets. Use Chapter 10's canonical encoding
for the outer measurement object; never decode/re-encode that string's contents.
This counts reasoning/encrypted/unknown data and all output items once, including
string-escaping cost, without also counting derived text/call views. Paired results
remain their ordinary tool entries, once each. Provider usage and the terminal
response envelope are excluded. This object is neither a neutral Part accepted
from callers nor a compressor source. All other pressure fields stay unchanged.

The smallest arithmetic control below is 127 UTF-8 bytes in Chapter 10 canonical
form. It measures a raw output array containing one opaque item, not an acceptable
standalone foreground response. Adding one ASCII byte inside x raises it to 128;
adding a duplicate text projection must fail equality rather than inflate pressure:

```json
{"messages":[{"parts":[{"raw_output":"[{\"x\":\"q\"}]","type":"responses_bundle"}],"role":"assistant"}],"system":"","tools":[]}
```

Full logs keep original receipts after a cut. Snapshot-only reconstruction returns
history_unavailable for the removed bodies; rehydration makes no provider call.
The retired record retains size/hash, identity, usage and retirement relation,
without raw output or indexed part positions. For helpers, separately test fold A
installing memory 12 and fold B retiring it to install 19. The snapshot retains
19's text and both usage receipts, but no A raw response, submit argument text or
memory-12 text or represented-output reference; its ID may remain in retirement
metadata. An unusable helper likewise retains only compact metadata.
Test full-log, latest-checkpoint and snapshot-tail equality for both helper folds
and for a suppressed foreground terminal result, including rejection of a second
slot settlement. Hash checks must not make absent bodies available again.

The following complete offline terminal response is a fixture. Its JSON is one
line, without LF; fictionally named identities must never be sent as live models:

```json
{"id":"resp-1","object":"response","status":"completed","model":"fixture-responses","error":null,"incomplete_details":null,"output":[{"id":"r-1","type":"reasoning","summary":[],"encrypted_content":"OPAQUE"},{"id":"m-1","type":"message","status":"completed","role":"assistant","phase":"commentary","content":[{"type":"output_text","text":"Checking.","annotations":[]},{"type":"output_text","text":"","annotations":[]}]},{"id":"f-1","type":"function_call","status":"completed","namespace":"ensemble","call_id":"call-1","name":"read_file","arguments":"{\"path\": \"note.txt\"}"}],"usage":{"input_tokens":150,"input_tokens_details":{"cache_write_tokens":20,"cached_tokens":30},"output_tokens":40,"output_tokens_details":{"reasoning_tokens":10},"total_tokens":190}}
```

Its complete refs are:

```json
[{"item":0,"content":null,"kind":"opaque"},{"item":1,"content":0,"kind":"text"},{"item":1,"content":1,"kind":"text"},{"item":2,"content":null,"kind":"call"}]
```

With developer BASE, user Read, and the matching result `note`, the next plan
request is exactly the following compact JSON. The copied spans retain their
original member order and the space inside arguments:

```json
{"model":"fixture-responses","store":false,"stream":true,"input":[{"type":"message","role":"developer","content":[{"type":"input_text","text":"BASE"}]},{"type":"message","role":"user","content":[{"type":"input_text","text":"Read"}]},{"id":"r-1","type":"reasoning","summary":[],"encrypted_content":"OPAQUE"},{"id":"m-1","type":"message","status":"completed","role":"assistant","phase":"commentary","content":[{"type":"output_text","text":"Checking.","annotations":[]},{"type":"output_text","text":"","annotations":[]}]},{"id":"f-1","type":"function_call","status":"completed","namespace":"ensemble","call_id":"call-1","name":"read_file","arguments":"{\"path\": \"note.txt\"}"},{"type":"function_call_output","call_id":"call-1","output":"note"}],"tools":[{"type":"namespace","name":"ensemble","description":"Agent-visible local tools.","tools":[{"type":"function","name":"read_file","description":"Read text.","parameters":{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]},"strict":false}]}]}
```

This fixture uses its printed test declaration, not a replacement for the real
read_file schema. The API-key control inserts `"max_output_tokens":256,` between
stream and input, changing nothing else. The corresponding final response is:

```json
{"id":"resp-2","object":"response","status":"completed","model":"fixture-responses","error":null,"incomplete_details":null,"output":[{"id":"m-2","type":"message","status":"completed","role":"assistant","phase":"final_answer","content":[{"type":"output_text","text":"The file says note.","annotations":[]}]}],"usage":{"input_tokens":160,"input_tokens_details":{"cache_write_tokens":0,"cached_tokens":100},"output_tokens":8,"output_tokens_details":{"reasoning_tokens":0},"total_tokens":168}}
```

Test full-log, checkpoint and snapshot-tail next-request equality independently.
Mutate a ref or an independently checked usage/binding/hash and require refusal.
Changing raw phase or spaced argument text must change exact replay and invalidate
a retained checkpoint digest; do not claim that a forged but internally consistent
full log authenticates a real provider response. A semantic substitute for spaced
argument text does not satisfy exact reconstruction. Repeat with duplicate argument
members, structural duplicate capsule fields, missing result and wrong call_id.
Use restart with a same-named different registration as a negative control.

For a separate codec fixture, accept exactly resp-1 above at admission 6/response 7,
pair its one call, and retire it through an authorized handoff at sequence 50.
Before retirement its retained record has raw equal to the printed terminal
string and refs equal to the printed list. Afterward the exact capsule row is:

```json
{"state":"retired","admission":6,"response":7,"bytes":760,"sha256":"80846537828a37f4f1ad91005e9245cc918253c0c475188454aec7630f9380c5","retired_by":50}
```

The base metadata reference remains exactly {capsule:{admission:6,response:7}};
its represented parts and call/result indexes are gone. The matching receipt's
raw_usage becomes {source:"inline",raw:STRING}, where STRING is the exact usage
object span from resp-1. Removing retired_by, keeping raw/refs, retaining a part
index, or changing either sequence refuses snapshot validation. Full-log replay
can still read the original receipt; this snapshot alone cannot.

For suppressed usage, start foreground admission 60 with no previously accepted
response in that turn. Accept its complete valid terminal usage immediately before
cancellation settlement. Exactly one model_usage_recorded closes slot 60, the
pending human is discarded, and turn_ended completes canceled. The snapshot has
one suppressed inline usage receipt, no capsule for 60 and no pending response
slot. With an earlier accepted response in the same turn, repeat and verify that
its dialogue remains. Inject a second model_usage_recorded, response_ended or
error_occurred for slot 60: each must refuse. Compare both cases through full-log,
checkpoint and tail restoration, then prove a fresh human turn can run.

### Streaming helpers still need an honest limit

Responses uses Chapter 6's SSE framing, finite queues, operation identity and
responsive control path. Track event sequence_number monotonically, item identity
by output_index and content identity by content_index. Unknown events consume
bounds but authorize no action. Validate known event fields, prohibit index/ID
rebinding, and compare completed item/text/argument observations with the terminal
output. Missing earlier fragments may be supplied by a done/terminal snapshot;
if fragments were observed, their concatenation must agree with the completed
value. Do not emit that snapshot again as duplicate visible text.
[Streaming events](https://developers.openai.com/api/reference/resources/responses/streaming-events).

Require response.completed with response.status completed, error and
incomplete_details null, nonempty producing model and valid final usage. Incomplete,
failed, error, EOF without completion or conflicting terminal output fails safely.
Known call async must be absent/false. Caller must be absent, null, or exactly
{"type":"direct"}; a bare "direct" string is invalid. A program caller or an
asynchronous execution request refuses unsupported_response_item.
No partial call executes; no usage is invented. A completed response with an
unusable submit_memory or judge verdict still records accepted usage once.
Input 150, writes 20, reads 30 normalizes to input 100/write 20/read 30; output
40 already includes the 10 reasoning tokens, so output remains 40. Require exact
nonnegative uint64 input_tokens, output_tokens and total_tokens, with the latter
equal to their checked sum. Validate cache subtraction; absent optional cache
breakdown counts as zero. Present details must be objects with valid counters;
reasoning_tokens, when present, cannot exceed output_tokens. Malformed/present-invalid
fields refuse. Preserve additional usage fields without treating them as new costs.

| Purpose | Existing API-key adapters | Explicit API-key Responses | ChatGPT plan Responses |
|---|---|---|---|
| Foreground | Existing delivery/cap | Stream; selected positive cap | Stream; cap absent |
| Compression | Captured delivery; cap at most 8,192 | Stream; same capped selection | Stream; cap absent |
| Judge | Plain; cap 256 | Stream; cap 256 | Stream; cap absent |

Helper input/system/schema/validation, deadlines and request counts remain those
of Chapters 15/16. Compression still has 2 MiB request and 8 MiB received limits.
Judge still has a 262,144-byte request and **8,192 decoded HTTP-entity bytes** of
received data. Count after content decoding, before SSE parsing: include comments,
line terminators, repeated item data and terminal response/usage; exclude HTTP
headers and transfer-chunk framing. Bound decompression as it produces bytes.
The frame limit never overrides this smaller whole-response limit. Use at most
one bounded input buffer, bounded SSE/assembly storage and one owned terminal
payload; no repeated unbounded body copies. Exceeding the limit yields
response_limit, unknown usage and the inherited safe helper fallback if still
eligible. No retry, plain fallback or invisible bound increase rescues it.

Here is an exact-limit stream fixture recipe, not provider evidence. Let E be the
following ASCII line followed by two LF bytes:

```text
data: {"type":"response.completed","sequence_number":0,"response":{"id":"j-1","object":"response","status":"completed","model":"fixture-responses","error":null,"incomplete_details":null,"output":[{"id":"j-m","type":"message","status":"completed","role":"assistant","phase":"final_answer","content":[{"type":"output_text","text":"[]","annotations":[]}]}],"usage":{"input_tokens":1,"input_tokens_details":{"cache_write_tokens":0,"cached_tokens":0},"output_tokens":1,"output_tokens_details":{"reasoning_tokens":0},"total_tokens":2}}}
```

E is exactly 532 bytes. Construct S as a colon, exactly 7,657 ASCII x bytes,
two LF bytes,
then E. Its first event is an SSE comment; S has exactly 8,192 bytes and a complete
valid empty verdict. Add one x to that comment for 8,193 bytes: refuse before
acceptance even though the verdict is still tiny. Split both streams at every
framing/UTF-8 boundary and under arbitrary HTTP chunks; the count cannot depend
on transport segmentation. Repeat with compressed transfer yielding those exact
decoded lengths, oversized terminal data after a small delta, missing completion,
invalid usage and cancellation. A successful real plan judge is still required:
these fixtures prove the limit, not that the selected model fits it.

## 17.9 Taking it for a spin: evidence still to collect

There is no demonstration transcript yet. Before paid work, the complete contract
and reviewer must agree on a finite matrix: all three human CLI paths, a public
two-Agent consumer, the actual browser view, helpers with accepted and unusable
outputs, fixed-body and growing-history comparisons, and the selected plan-route
investigation. Each action needs an attempt cap, expected receipt and stopping
condition. A zero cache-read count remains a result; it does not authorize trying
again until a positive number appears.

Use at most 16 model starts per route and 80 total in the initial reviewed
matrix, counting startup through shutdown and every helper, failure and canceled
start. Routes are Messages API-key, Chat Completions API-key, Generate Content
API-key, Responses API-key and Responses ChatGPT-plan. Allocate those starts to
fixed-body pairs, growing-history/tool turns, helper success and public/browser
observations before launch; client reuse does not create another allowance.
A reviewer may approve a revised finite matrix when actual behavior requires it;
never reset counts or retry merely to obtain a positive cache number.

For both Responses modes, the matrix requires an actual successful foreground
call/result continuation, accepted compression with installed usable memory and
accepted judge with a valid verdict inside the stated local bounds. An 8,192-byte
judge refusal is evidence of a limitation, not completion of that feature. Stop,
retain it and revise the published bound with rationale/checks before any changed
implementation. Use the same discovered model identity across funding controls
only if both routes offer it. Otherwise retain the comparison's mismatch explicitly;
do not invent entitlement or call differing models a matched pair. Turn off markers
on both Responses controls. Their remote output-cap difference remains recorded.

Actual own-application consent precedes plan use, with no copy of another client's
token store. Use local fake authorization/refresh fixtures for rare faults and
revocation races; label them separately. The real public consumer must show two
Agents sharing a connection without sharing usage/pairs, close one while the other
continues, then close the application. The GUI must show route/funding, counts,
price completeness and reconnect behavior through the real public seam. Normal
CLI auth/status actions and the real browser cache panel need usable receipts;
none requires giving the model credential authority.

The first user exercise should load a real manual, submit a question, inspect
the first pair, add a hint, perform a file read and ask a follow-up. The reader
then sees which change came from the one-shot hint and which came from growing
history. A handoff should show a deliberate prefix change beside any reduction
in total input. Cache percentage alone cannot decide whether that trade helped.

Retain final bodies at actual handoff, raw accepted usage and source/executable
bindings. Label local fixture numbers and reconstructed requests separately from
captured wire evidence. Before inference, prove that the capture verifier rejects
one changed identity at a time on otherwise valid paths. Record a real screenshot
with a text description of the browser display. Do not ask the model whether its
cache worked; the model is another consumer of the instrument.

| Distinguishing check | Positive and defect that it separates |
|---|---|
| Capture identity | Foreground/helper interleaving; a render-only call cannot advance the pair |
| Bounded lifetime | Two Agents, busy/oversize gaps, retirement during export and close/join; no old-neighbor substitution |
| Comparison | Actual compact/indented renderer output, literal offsets and marker-like user data; constant verdicts fail |
| Policy causality | H/S/P, opaque, duplicate candidates and all purpose renderers; a marker after a transient prefix fails |
| Accounting | Mixed producing identities, accepted unusable helper and failed stream; current-model repricing/double counting fail |
| Arithmetic | Exact rationals, zero denominator, uint64 browser round trip, overflow and partial price completeness |
| Public clients | Human inspection, two-Agent consumer, opaque IDs, current profile/table date and browser reconnect; stale mount/shared totals fail |
| Persistence | Six base profiles, strict v7 codec and old-route equality; whole-bundle authorized handoff, retired/helper snapshot omission, suppressed-slot settlement and honest unavailable reconstruction |
| Credentials and cancellation | Protected restart binding; wrong-name reuse, lock contention, isolated refresh waiters, invalidation/dispatch/settlement barriers and no fallback |
| Responses | Literal ordered item/phase/argument/result fixtures; foreign namespace refusal, unknown item retention, final usage and stopped effects |
| Helper bounds | Exact 8,192/8,193 decoded entity bytes, terminal duplication, missing completion, invalid verdict versus invalid response, and successful actual plan helpers |

The contract now goes to independent review. The new acceptance invocation,
actual source-bound runs, matched-control limitations, comparative revisions and
student teaching feedback remain in the [validation record](chapter-17-validation.md).
There is no promised positive cache result at the end of the exercise. The useful
outcome is a measurement the reader can trust, including when it says unknown.
