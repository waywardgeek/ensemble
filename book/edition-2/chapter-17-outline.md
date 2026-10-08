# Chapter 17 outline: The invisible invoice

Through-line stake: a developer can get the right answer while paying repeatedly
for the same context; the chapter resolves that uncertainty with an instrument
whose request evidence, provider observations and cost estimates remain distinct.

October 8, 2026. Research and proposed design only, mapped from first-edition
Chapter 18. This is not a student contract or an implementation release. The
consequential choices below require coordinator and independent review before a
full draft. Chapter 16 currently has a reviewed draft and grouped corrections,
not an accepted runtime. Historical and current primary sources are recorded in
[the evidence ledger](chapter-17-evidence.md).

## Story and teaching order

Retain the first edition's practical motivation: correct answers gave Bill no
warning that caching needed attention before daily use. The September 27 design
note documents that concern. Do not repeat its cold-request inference, the old
opener's dollar total or a universal savings multiplier as measured facts.

The best supported incident is the instrument's own false alarm. A marker remover
matched compact JSON while its actual input was indented. Its handwritten tests
used the same compact shape and passed. The request still contained markers, but
the report said there were none. The coder review and commit `3746095` preserve
the diagnosis; no recovered original wire transcript is claimed. Keep the desired
measurement, failed assumption, observed discrepancy and corrective test together.

Return to the reader's bill at the meter: commit `40afc7a` records a paid model
shown with zero cost because its name came from the environment while the meter
looked in an empty settings field. Chapter 2 already preserves producing-model
identity, so the rewrite teaches how to use it correctly from the start.

Proposed order:

1. Explain what a correct answer cannot reveal about repeated input cost.
2. Follow one actual request through projection, cache policy, serialization,
   durable admission, transport, accepted usage and the display.
3. Compare two requests without allowing the diagnostic to manufacture bytes or
   confuse a helper request with the foreground conversation.
4. Explain persistent versus one-shot material and legitimate prefix changes.
5. Place supported provider directives using those boundaries; keep uncertain
   backend behavior out of deterministic acceptance rules.
6. Price observed usage by producing model and distinguish durable session totals
   from the current mounted run. Show unknown pricing honestly.
7. Publish CLI, browser and public diagnostics, then demonstrate bounded real
   requests and the separately scoped subscription-route investigation.

Use a small configuration task through the middle: load a manual, append a hint,
read a file, make a handoff and recall a decision. Ask which bytes should remain,
which legitimately change and which request was actually billed. This is an
illustrative task until the later actual spin supplies receipts. No invented
dialogue, latency reduction or successful cache hit belongs in the draft.

## Inherited boundaries that must survive

| Predecessor | Consequence here |
|---|---|
| Chapters 2 and 6 | Producing provenance and disjoint usage are accepted facts; partial streaming text is not a completed response. Preserve opaque material and actual number lexemes. |
| Chapters 5 and 9 | Actor owns admission. Skills manuals have chronological anchors, primary bytes occur once, hints are consumed once, and grants may legitimately change the tool array. Cache work cannot reorder H/S/P or widen grants. |
| Chapter 10 | EventLog alone owns the append descriptor; prepared accepted bytes precede apply/observe. SessionStore owns locking/checkpoint work. A diagnostic is neither a second history writer nor a resume authority. |
| Chapters 11–12 | MCP transport stays replaceable; real WebSocket tunneling remains in the optional GUI module. No new core dependency on that socket or GUI implementation. |
| Chapters 12–13 | Captured observations and speech state have explicit owners and lifetimes. A diagnostic cannot sample the GUI, retain a typing pause or announce ordinary cache misses as system failure. |
| Chapter 14 | Cuts are deliberate recorded projection changes. Opaque keep policy stays default; no stripping signatures or changing maintenance to obtain a better hit rate. |
| Chapter 15 | Generated memory can replace earlier material, including older notes. No promise that every prefix before a handoff remains unchanged forever; no diagnostic archive of retired body copies in snapshots. |
| Chapter 16 | Recall is accepted later but placed before its admitted prompt. It persists until handoff. Judge requests use a separate purpose and share the captured effective request budget. |

Chapter 15 current pressure and immutable captured turn allowance remain separate.
Diagnostics cannot add warm-up requests, retries or paid token counting behind
the user's turn. Usage for memory/recall helpers contributes to total cost even
when their output is rejected or never attached. A low cache rate does not prove
that compaction was uneconomic: the total input may have fallen substantially.

## Proposed ownership and data plan

| Owner | Responsibility and actual parent |
|---|---|
| Agent | Current cache configuration and creation capability; existing Actor, Engine and durable authority |
| Engine | Accepted per-producing-model accounting; constructs an owned diagnostic child reachable through a common Engine parent interface |
| Diagnostic child | Bounded request comparison state, safe summaries and optional export; follows Engine → Agent → Ensemble for configuration and logging |
| Actor | Captures the request's policy/configuration and serializes authoritative admission; diagnostic failure cannot pretend a response committed |
| Optional GUI Page | Displays public copied values, owns listeners and view lifetime; no private Engine access or independent cost accumulator |
| Application credential owner, if selected | Supplies current authenticated route through an explicit parent/interface seam; credentials are absent from events, snapshots and diagnostics |

Common holds shared serializable values and interfaces; behavior belongs in its
spoke. No sibling injection, closure-based reports, mutable globals or alternative
CLI composition root. Price data is immutable selected data with a version/date;
the display does not rewrite usage history when that table changes.

Record identities should connect runtime Agent, durable SessionID when present,
turn, attempt/event sequence, purpose, route/surface, requested/resolved model
and a safe credential-mode label. An authentication token or account identifier
is unnecessary diagnostic content. Exact public names remain a student choice
after the behavior is settled.

## Consequential choices before drafting

### D1. Request evidence and comparison lifetime

Recommend observing the final prepared body at Engine transport handoff, after
all projection and cache decisions. Rendering alone must do no I/O and must not
advance the diagnostic. Capture every actual attempt, with independent comparison
streams for foreground, compression and recall judge so their unrelated prompts
cannot overwrite one another's predecessor.

Choose whether a changed model/surface/configuration starts a new comparison
baseline or produces an explicitly noncomparable result. Prefer one bounded
previous body per purpose for the current route, clearing on route change, over
an ever-growing map of every historical route. Account request and response
completion by identity, including a failed attempt; observing a send is not proof
that the server charged it. Do not recover a warm local baseline from retired
session bodies on restart.

Propose metadata by default and explicitly selected export of a rotating request
pair. Set byte budgets before implementation; a request larger than the diagnostic
budget should produce an incomplete diagnostic without truncating the actual
provider request. Decide whether to export exact bodies separately from the
annotated comparison view. Prefer exact bodies plus explicit offsets/hashes,
avoiding a rewritten file mislabeled as the request. Export failure is visible;
ordinary inference continues, while an evidence run requiring that export fails.

### D2. What the local instrument can establish

Publish independent measures: raw-body prefix bytes; comparison over provider
sections with only specified directives excluded; structural append/edit paths;
and matched eligible marker positions. Any normalized comparison must name its
transformation and retain original offsets. Strings containing marker-like text,
opaque raw values and number lexemes must survive unchanged. JSON object member
order and message-array closing delimiters are not interchangeable phenomena.

Describe provider concatenation order only where documented. Preserve section
identity for messages, contents and any selected Responses input surface rather
than hard-coding a single field. Do not advertise raw HTTP byte equality as an
exact copy of the provider's hidden tokenized cache key.

The local report should distinguish identical, appended, changed and unavailable,
with causal projection/configuration changes where known. Legitimate Skills,
memory, recall and handoff changes are visible explanations, not automatic bugs.
Provider cache counters remain token observations. Never subtract bytes from
tokens or declare two percentages proof of equal caching. A stable body followed
by zero reported reads is unresolved without further evidence.

### D3. Marker policy without changing conversation meaning

Propose a default-preserving off/implicit policy and an explicit supported marker
mode, captured per request. Prefer a small provider-specific placement algorithm
over a universal guessed cache layout. Candidate reusable boundaries are the end
of tool definitions, immutable primary/base instructions, the newest represented
handoff where present, and stable material before one-shot hints/ephemera/current
observations. Define priority and deduplication when boundaries coincide or a
provider supports fewer locations. Never insert a marker into opaque replay data.

Resolve whether block splitting to preserve a reusable boundary is allowed by the
new profile, with literal fixtures proving unchanged content/role/call order.
Historical blocks may already carry protocol material that cannot be remarshalized.
Do not revive the old permanent start-of-turn anchor as an unexplained hedge.
The documented backend lookup window has a current source disagreement; the
exercise must not grade a guessed 50-versus-80 internal limit.

Bill's documented historical preference for Gemini implicit caching is retained
as the proposed default. Explicit remote cache-resource creation and storage
billing would be another lifecycle and are excluded unless separately selected.
Do not import the old NoEphemera model rule: cache optimization cannot silently
discard steering or inherited opaque history.

### D4. Strict formats and replay

Changing marker/configuration fields changes exact reconstructed requests. Prefer
an explicit fresh v7 capability/profile, with unchanged v1–v6 construction paths,
over unversioned insertion into old strict request snapshots. Decide the narrow
identity/configuration extension and whether any durable diagnostic metadata is
needed beyond captured rendering inputs and existing response usage. Full wire
requests need not become another required semantic-history archive.

Record renderer/cache policy version and inputs sufficient to reproduce an
available historical request. Snapshot-only history retains history_unavailable
when source bytes were legitimately retired. Price-table selection and current
credential/route remain separate from immutable session identity. Standalone
admission, initializers, strict unknown/duplicate fields and physical record
classes must be printed together with any new event shapes before checks exist.

### D5. Cost and usage scopes

Keep Engine's durable per-producing-model totals. Add an explicitly named mounted
run view starting at successful construction/resume, plus last accepted response
and purpose subtotals. Avoid reusing “session” for both durable SessionID and a
process lifetime. Two Agents mounted in one application need distinct baselines;
closing one must not reset another's totals.

Propose exact checked arithmetic over full uint64 counters and rational/decimal
rates, with browser values represented losslessly. Define zero-denominator rate,
rounding and overflow dispositions. Cache-read share uses the complete disjoint
input total, including cache writes. Price by actual producing identity/surface,
not today's selected model or an empty GUI setting. Unknown prices make the
aggregate incomplete; a known partial subtotal is never the complete bill.

An estimate at a dated API rate is different from charged money, subscription
allowance or latency. A funding-mode label may be needed for a plan route.
Do not imply that reported cache tokens reveal how a subscription was billed.
Choose the initial price-source/table scope during review; no speculative rates
or API-versus-plan dollar comparison is settled here.

### D6. Supported subscription experiment and surface scope

The current official route research is material to Bill's requested OAuth cache
retest. It points to a Responses flow with application-specific authorization,
not another client's credential store. It also publishes unsupported settings
that conflict with inherited helper output caps; the evidence ledger gives the
exact sources and access date.

Two explicit options need coordinator review: add a bounded, separately taught
Responses/credential surface here so the actual Ensemble program can perform the
test; or first use a clearly separate diagnostic consumer, leaving the Ensemble
feature gate open until the supported route is integrated in the next chapter.
The latter proves only that consumer's route. Neither option silently substitutes
Codex app-server for Ensemble's loop or drops max-token/settings guarantees.

Recommend separating a caller-supplied authorized-token seam from a complete
interactive sign-in/refresh manager if that keeps this chapter coherent. Exact
lease/expiry/cancellation/configuration ownership and bounded-spend behavior must
be settled before any such seam becomes teaching. No login or inference is part
of this research; an actual user authorization step may later be necessary for
the application's own account connection.

### D7. Public and human diagnostics

Use the same public copied status through CLI and browser, with explicit inspection
of the bounded request pair. Print current model, last accepted purpose/usage,
mounted-run and durable totals, cache observations and price completeness. A later
model-facing status tool is optional pending review: automatic installation would
change the frozen handler ceiling and every skill-enabled resume identity.

The browser receives safe bounded metadata, never complete raw request bodies by
default. Define exact new wire fields/commands, full uint64 decoding, watch/reconnect
semantics and late-callback fencing in the draft. A real public two-Agent consumer
must show independent baselines and failure outcomes, not only a CLI that imports
private helpers. Existing MCP-over-WebSocket inspection can observe this display
through its already granted scope without changing the core transport seam.

## Acceptance and actual-use plan

| Layer | Required distinguishing evidence |
|---|---|
| Deterministic request analysis | Actual renderer output, compact/indented input, marker-like strings, member positions, all supported dialogue fields, append versus edit, exact offset/hash, invalid/oversize input, diagnostic failure |
| Projection and markers | H/S/P, merged content, opaque presence, changing grants, memory replacement, recall anchor/retirement, no duplicate marker or transient material inside a promised reusable boundary |
| Causality and cost | Multiple purposes and Agents, failure between sends, no diagnostic accounting commit, once-only accepted usage, producing-model switch, unknown/partial rate and exact integer overflow |
| Persistence | Off-mode old bytes, strict new profile and old routes, mounted-run reset versus durable retention, no retired-body archive, lossless historical reconstruction where available |
| Public/browser | Shared library behavior, bounded display/inspection, exact counters, real screenshot with alt text, close/remount/reconnect and separate-module checks |
| Live API-key paths | All three human CLI PTYs plus required public/browser actions; retain final body and raw usage, sequential follow-ups and actual tool result, with each purpose's spending visible |
| OAuth investigation | Official current route, real authorization/access outcome, fixed prefix/surface/model/configuration and source-bound counters; API-key control only where genuinely comparable |

Before paid work, publish a concrete request cap/action/receipt matrix and locally
verify capture/negative controls. Select prompts above the current model's stated
floor without calling byte length a token count. Distinguish fixed-body and growing
conversation probes; do not generalize one shape to all agents. Initial zero does
not prove a globally cold server, nor does a fresh process clear remote caches.
No unattended retries to obtain a positive hit. A route blocker, missing field,
refused request or repeated zero remains an observed limitation.

The historical seven-check grader is evidence of scope, not the new acceptance
matrix. Extend behavior and structural checks independently, preserve historical
baselines and use compiling deletion controls with valid positives. A fake usage
reply can establish normalization and presentation, never a provider cache hit.
Actual spin, comparative review, student feedback and final prose remain pending.
