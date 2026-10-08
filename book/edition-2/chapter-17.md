# Chapter 17: The invisible invoice

A coding agent can make the right edit, pass the tests and still waste money on
every turn. The developer sees the edit. The repeated input charge is harder to
see, especially when a memory compressor or recall judge quietly contributes
another request before the answer appears.

In September 2026, Bill identified caching as a barrier to using Ensemble every
day. The first instrument built to explain it developed a problem of its own.
It searched for compact JSON, while the input it actually compared was indented.
The test fixtures were compact too. They passed; the instrument reported missing
markers that were present in the request. A convincing number can be an excellent
disguise for a broken measurement.

This chapter starts with that instrument. It separates the bytes the application
sent, the cache usage the provider reported and the price estimated from accepted
usage. Each has a different source of truth. The developer needs all three to
decide whether a change helped.

*Partial draft, October 8, 2026. The coordinator has selected the independent
diagnostic, marker, accounting and display design below. The subscription route
and its output-cap choice remain pending in the [validation record](chapter-17-validation.md).
This is not a complete student contract or an implementation release. No Chapter
17 build, live measurement, cache hit or saving is claimed. The historical incidents
are attributed in the [evidence ledger](chapter-17-evidence.md).*

## Partial exercise summary

Read the complete [coding skill](skills/ensemble-coding/SKILL.md),
[architecture](architecture.md) and accepted predecessor before eventual coding.
The working design here adds the following behavior:

1. Engine owns a bounded diagnostic child. Observe exact prepared request bytes
   at transport handoff after authoritative admission. Keep foreground,
   compression and recall-judge comparison streams separate. Rendering and
   reconstruction perform no diagnostic admission or paid work.
2. Report raw byte prefix, structural append/edit and directive-aware section
   comparison separately. Preserve original offsets, member order and opaque
   lexemes. An unavailable comparison must explain its gap.
3. Preserve old default request bytes. An explicitly selected new marker policy
   uses at most four Messages manual boundaries or three Chat Completions
   explicit boundaries alongside implicit placement. Generate Content retains
   its implicit path. Preserve hints, grant changes and signed material.
4. Capture rendering policy before serialization. New cache-affecting captures
   require the forthcoming strict v7 profile; v1–v6 stay exact. The unfinished
   identity/initializer matrix in §17.5 is a release blocker, not a student choice.
5. Keep accepted usage on Engine. Display durable, current-mount, purpose and
   last-accepted views; distinguish an incomplete latest attempt. Estimate prices
   with exact arithmetic and dated producing-identity rates. Unknown stays unknown.
6. Publish copied status to the human CLI, optional browser module and public
   embedding clients. Exact bodies require a separate explicit export. Ordinary
   displays contain no request text, credentials or account identifiers.

The complete exercise will supply the final build/check command and route-dependent
wire fixtures before a cold student starts. The historical Chapter 18 grader is
diagnostic evidence; its seven categories do not cover this chapter's promises.

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

The first edition's compact-pattern failure is a useful test specification.
The instrument must analyze the renderer's output as it exists, including
indentation. Repeating its own simplifying assumption in a handwritten fixture
does not establish that property.

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
systemInstruction and contents. Missing and empty sections remain distinct.
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

The working policy has two choices: off and stable_v1. Off adds no new fields or
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

The proposed per-request cache capture has exactly version, policy, mechanism and
capability. Version is 1; policy is off or stable_v1. Mechanism is unchanged,
messages_manual_5m, chat_implicit_30m or gemini_implicit. Capability is null for
unchanged/Gemini, otherwise the immutable verified capability-entry ID. The
selected policy and mechanism must agree with the actual surface and fields.
No credential or price table belongs in this capture. These examples describe
new cache members only, not complete existing request events:

```json
{"version":1,"policy":"stable_v1","mechanism":"messages_manual_5m","capability":"fixture-messages-cache-v1"}
```

```json
{"version":1,"policy":"off","mechanism":"unchanged","capability":null}
```

Fixture capability names are offline identities, never live model names. Old
records without a cache member retain their original rendering behavior. V1–v6
strict codecs do not acquire optional fields. The new cache-affecting capability
uses v7, and construction, initializer, anchor, origin, semantic state and
checkpoint must agree on that version. Selecting caching must not enable memory
compression, recall or its judge as a side effect.

**Partial-contract boundary:** the complete v7 identity/member matrix, combinations
with earlier capabilities, construction initializer order, standalone recognition,
policy-file version and physical record classes remain to be printed and reviewed
before implementation. The route-dependent provenance extension must be settled
alongside §17.8. A student or grader must not fill these gaps privately. Metadata-only
diagnostics and price views do not justify another durable request archive.

Snapshot reconstruction continues to use available recorded sources. A retired
body returns history_unavailable, even when a diagnostic could once compare it.
Resume clears local request pairs and starts a new mount baseline. It does not
perform HTTP, recover exported bodies or call a reconstructed pair observed.

## 17.6 Read the meter's label before its number

The first edition's meter once showed a paid model with zero cost. The model
selection had come from the environment; the meter consulted an empty settings
field. The producing identity was available, but the display asked the wrong
owner. Changing the label to unknown would have been more honest than zero;
using the actual producing identity was the useful repair.

Keep normalized, disjoint input/write/read/output counters on Engine. The current
configuration cannot reprice yesterday's responses. Each accepted fact retains
its producing model/surface, purpose and effective rate-affecting options. A rate
table is immutable selected data with an ID, source date, currency and exact
rational rates. The initial real table will cover the identities actually chosen
for the bounded demonstration; no live price or model choice is invented here.

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
This draft adds only Messages' selected default duration and Chat Completions'
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

## 17.7 Put the explanation where the user can see it

Add a Cache and usage view to the existing optional GUI and a `/cache` inspection
command to human chat. Both use the same copied public status. Preserve `/usage`
and the existing protocol replies. No automatically installed Registry tool or
new Skills grant is needed. The model can inspect the display through the
already granted GUI MCP tools when that integration is selected.

The proposed status has exactly version, agent, mount, revision, attempts, usage,
pricing and omitted. Version is 1. Mount and revision are decimal uint64 strings;
the Agent identity is its existing safe public identity. Attempts contains one
row per purpose in foreground/compression/judge order, with null latest/pair for
an empty stream. Each row contains purpose, latest, pair and code. Latest names
the inherited operation/admission identity and outcome; pair names its two
attempts, byte lengths, hashes and the three comparison results. Request values,
account identifiers, endpoint query strings and authorization data are excluded.

Usage groups the four scopes in §17.6, with all counters as decimal strings and
explicit unavailable reasons. Pricing contains table, currency, exact known
subtotal, displayed subtotal, complete, unpriced and incomplete_attempts. Omitted
counts whole omitted rows/details; it never means a truncated numeric string.
The complete nested wire grammar remains part of the final contract gate.

Status revision advances only when its copied contents change. Use the inherited
watch snapshot-plus-tail boundary: reconnect obtains the current mount/revision,
and a late old-mount callback cannot replace it. Browser code holds view state,
not a second lifetime usage accumulator. Updates and copies obey the 256 KiB
bound and existing observer overflow/resynchronization behavior. Closing the
view detaches its watch without closing Agent, changing pause causes or exporting
bodies. A remounted view receives the correct new baseline through the same seam.

The planned explicit CLI export command is `/cache export PURPOSE DIR`, where
PURPOSE is foreground, compression or judge and DIR is a new directory. A missing
pair reports no_baseline; a retired/gapped pair reports its reason. The public
consumer has the same export capability and owns its returned values. Browser
status has no automatic raw-body download. The actual command acknowledgement
grammar will be published with the full interface contract.

A two-Agent fixture should accept rate A's example on Agent A, then construct B,
accept rate B's example on B and fail A's next attempt before accepted usage.
A still shows durable/current 0.000656, its earlier accepted response and the new
failed attempt separately. B shows 0.001312. Closing and resuming A resets only
A's current-mount amount to zero. Stale A callbacks cannot overwrite B or the
replacement A. This test catches a shared counter that a one-window demo misses.

## 17.8 The subscription route is a real design decision

Bill's requested OAuth cache retest remains an open obligation. The coordinator
has asked whether to offer an explicit plan mode without the inherited output
cap or retain mandatory caps and leave that route pending. This draft assumes
neither answer. API-key behavior and the existing helper ceilings remain intact,
including the recall judge's plain delivery and fixed output limit.

The eventual section must specify the chosen surface, full-history request
mapping, tool calls/results, opaque provenance, terminal acceptance and usage,
strict replay, public configuration and the application's own authorization
lifetime. A different authentication header alone does not provide that contract.
A separately labeled probe cannot close the Ensemble integration gate. No
credential substitution, silent funding fallback or automatic paid retry belongs
in the unresolved space.

## 17.9 Taking it for a spin: evidence still to collect

There is no demonstration transcript yet. Before paid work, the complete contract
and reviewer must agree on a finite matrix: all three human CLI paths, a public
two-Agent consumer, the actual browser view, helpers with accepted and unusable
outputs, fixed-body and growing-history comparisons, and the selected plan-route
investigation. Each action needs an attempt cap, expected receipt and stopping
condition. A zero cache-read count remains a result; it does not authorize trying
again until a positive number appears.

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
| Public clients | Human inspection, two-Agent public consumer and real browser reconnect; stale mount/shared totals fail |
| Persistence | Exact old routes, fully specified new v7 route, retired-body absence and no-call reconstruction |

The remaining route decision, complete strict schemas, live price/capability
selection, grader command, actual runs and independent review belong in the
linked validation record. This partial draft establishes the independent teaching
without claiming that those gates have been completed.
