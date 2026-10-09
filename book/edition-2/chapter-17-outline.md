# Chapter 17 outline: The invisible invoice

Through-line stake: a developer can get the right answer while paying repeatedly
for the same context; the chapter resolves that uncertainty with an instrument
whose request evidence, provider observations and cost estimates remain distinct.

October 8, 2026. The [complete contract draft](chapter-17.md) now resolves the
C1–C4 findings from review `9598665` and awaits independent closure. The coordinator accepted the integration direction and R1–R4
at `a5423ed` after Bill's cap decision at `706682c`. This is not implementation
release or runtime acceptance. The [validation record](chapter-17-validation.md)
owns remaining gates, including the new checker invocation before student release.
Historical and current primary sources are in [the evidence ledger](chapter-17-evidence.md).

The initial outline at `580ce83`, mapped from first-edition Chapter 18, preserved
D1–D7 as proposals. Their research and reasoning remain below; the dated drafting
disposition at the end distinguishes selected choices from still-open work.

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

### D6. Original subscription proposal, superseded below

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

## Partial drafting disposition, October 8

At the `b9772c3` partial-draft boundary, the coordinator had accepted D1–D5/D7
direction at `e9d78da` and D6 still awaited Bill. That historical draft assumed
no answer and released no student. The decision and integration proposal below
supersede that unresolved preference; predecessor runtime acceptance remains
a separate gate.

- D1: §17.2 specifies the selected 48 MiB retained-body cap plus separate
  auxiliary work, export, metadata and wire bounds. Busy/oversize observations
  create explicit adjacency gaps. Retirement cancels unpublished affected work;
  export success is manifest publication, while prior selected exports have
  independent lifetime.
- D2: §17.3 publishes raw A/A, A/B and A/C byte coordinates, compact/indented
  marker fixtures, marker-like schema data, member order and opaque lexemes.
  Structural comparison ignores layout only; original bodies remain exact.
- D3: §17.4 verifies and prints the selected wire fragments. Earliest transient
  material bounds the entire reusable prefix, with H/S/P and text-split fixtures.
  Default/off preserves inherited bytes. The documented lookup-window conflict
  stays visible. Capabilities must be exact verified entries, not name sorting.
- D4: §17.5 fixes the capture-before-render rule and prints proposed cache-member
  examples. Full v7 capability/initializer/record-class composition and the
  route-dependent provenance extension remain explicitly unfinished; no checker
  may fill them privately. No automatic enablement of memory or recall follows.
- D5: §17.6 uses explicitly synthetic rational-rate fixtures, mixed-model totals,
  browser integer boundaries, rounding, overflow and visible partial estimates.
  Actual price-table identities and citations await the finite live selection.
- D7: §17.7 proposes the shared copied status, bounded view, mount fencing and
  explicit local export. Exact nested command/status grammar remains a full
  contract gate. The existing GUI MCP inspection path remains available only
  under its inherited grants and transport boundary.

The historical instrument's compact/indented mismatch opens the mechanism, and
the environment-selected model's empty price lookup returns in the meter section.
The draft keeps the reader's correct-result/wrong-cost problem visible without
reusing unverified invoice amounts, claimed savings or original live percentages.
The illustrative manual/hint/read/handoff exercise is still a plan; it contains
no manufactured exchange. Original historical chapters and code are unchanged.

The next step at that boundary was coordinator review and resolution of D6.
The current next step is integration review below, followed by the complete
persistence/public grammar and capability/rate/live matrix before student handoff.
Full draft proofreading remains separate from scoped partial-draft lint.

## Subscription integration proposal after Bill's decision

October 8, 2026. Bill approved the explicitly selected uncapped subscription
mode; `706682c` records that decision. The following integration choices are
**proposals for coordinator review**, not further rulings attributed to him and
not a released student contract. Preserve the original author's story/personality
edits at `53ee8e7`; the mechanism below extends that manuscript.

### Selection and ownership

Add the Responses surface with explicit funding selection, `api_key` or
`chatgpt_plan`. Existing API-key Messages, Chat Completions and Generate Content
remain the defaults for their respective providers, with unchanged caps and
delivery. Propose an explicitly selected API-key Responses path too, to permit
the closest same-surface control. It remains capped. A comparison must disclose
that residual difference rather than claiming identical requests. No route
silently switches credentials, model, surface or funding when a request fails.

Ensemble owns a Connections service in a new implementation spoke. Connections
owns named Connection records; each Connection owns authorization, refresh work
and short-lived credential leases. Common declares the value/parent interfaces.
Agent owns the selected safe connection name and route configuration. Engine
operations reach Connections through Engine → Agent → Ensemble, retaining their
actual parent rather than an injected token supplier. The service never owns
conversation state, usage or the diagnostic pair. GUI continues to be optional.

Capture funding, surface, local connection generation and selected model for the
turn, including its helpers. A configuration change applies to the next admitted
turn; a separately admitted idle compression operation captures its own selection.
Credential rotation is a lease revision, not a different account generation.
Signing out or replacing the verified identity changes generation and invalidates
new admission against the old one. Capture no access token, account identifier,
client registration ID or credential path in an event, request body or snapshot.

The public construction/configuration seam and human CLI expose login, connection
listing, explicit selection, model discovery and sign-out. An active safe label
and `ChatGPT plan; no per-response output cap` must be visible before paid work.
Labels are local user-chosen names, never an email inferred from a token. Reopening
a session requires an explicit local binding for its safe route name; it does not
restore credentials from the snapshot or opportunistically select another account.

### Own-application authorization and leases

Follow the current [registration guide](https://developers.openai.com/siwc/token-sharing-open-source/sign-in):
own application identity and host registration, loopback listener established
before browser launch, fresh state/nonce and S256 PKCE, returned client identity
validation, code exchange and verified ID-token claims. Check granted plan scope
before enabling plan inference. Never import another application's registration
or tokens. The chapter will print the exact endpoint/parameter fixtures and
callback refusal cases, without putting secrets in command arguments.

Propose a user-selected private credential directory outside the repository,
with owner-only directory/files and atomic replacement. Reject symlink traversal
and unsafe permissions. Keep separate records by verified registration/identity.
One Ensemble holds an exclusive process lease on a mounted credential store;
other processes get `connection_store_in_use`, while its own Agents may share a
Connection. This deliberately modest rule makes refresh-token rotation serial
across processes without building a credential broker. The selected store path
is creation-only application configuration, never Agent configuration.

Before request admission, a cancellable lease acquisition refreshes near expiry
through the Connection's one refresh worker. Atomic persistence of replacements
precedes issuing new leases. Failed or uncertain renewal refuses new admission;
it does not resend a model request or fall back to an API key. Lease waiting counts
against the caller's existing deadline. Use finite auth-network timeouts and bound
the browser authorization wait. The final contract must print those local limits
and state how returned expiry metadata is validated; undocumented interpretation
of `earliest_refresh_at` is not assumed from its name.

Agent close cancels/releases its own operations, not another Agent's credential
access. Ensemble close stops admission, joins refresh/lease work and releases the
store lock without revoking the registration. Explicit sign-out first prevents new
leases and cancels users of that Connection, then attempts documented revocation
and clears local tokens. Report unconfirmed remote revocation accurately. These
lifetimes follow the [accounts/session guidance](https://developers.openai.com/siwc/token-sharing-open-source/profiles-and-sessions)
but the owner hierarchy and exclusive store lease are application design choices.

### Requests, history and effects

Use account-specific model discovery and the public `/v1/responses` route with
`store:false`, `stream:true` and the complete input array. Select a discovered
slug; discovery does not prove every content/tool capability. The plan route
omits `max_output_tokens`, server-side conversation/previous-response state and
other unsupported controls. Its documented constraints are in
[models and inference](https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference)
and [preview limitations](https://developers.openai.com/siwc/token-sharing-open-source/preview-limitations).

Map the primary instruction to a developer input message. Preserve the inherited
ordered projection of human input, hints, manuals, recall and tool results; none
becomes developer authority. Declare the currently visible ordinary functions in
one `ensemble` namespace, retaining each local name and schema with explicit
non-strict schema mode. No hosted MCP, hosted shell or alternative tool executor
is introduced. A function call must name that namespace and the existing visible
function. Results use its exact call_id. Unknown names still receive the taught
controlled tool error; a foreign namespace cannot be stripped into authority.
Compressor functions use a separate `memory` namespace containing only submit_memory.

Preserve exact function.arguments string bytes, including whitespace and controlled
invalid/duplicate argument text under Chapter 10. Introduce a **new** v7 replay-item
capsule, not a second meaning for text.opaque. Each accepted Responses output item
is stored once with its raw item JSON, producing provenance and ordered references
to its neutral display/call projections. A message preserves phase, all content
positions, annotations and empty text; a call preserves namespace, item ID,
call_id and arguments; encrypted reasoning remains opaque. Validate that those
projections exactly correspond before durable acceptance. Public/browser copies
contain only the existing safe projection. Retirement removes capsules with their
source; helper capsules never enter the foreground conversation.

Reconstruction sends retained item envelopes in order, then the corresponding
function results. It does not rebuild an assistant item from concatenated text or
substitute a newly generated call ID. A bound item requires compatible exact
producing model/surface provenance. Funding mode is captured for accounting and
diagnostic separation; cross-connection opaque reuse is conservatively refused
unless explicitly established and taught. Ordinary cross-provider projection
keeps the earlier rules, and unknown opaque material cannot silently disappear.
The final chapter must print the capsule grammar, correspondence fixtures and
retirement/snapshot rules before any implementation.

Typed SSE events use the existing Engine-owned operation and Actor acceptance
boundary. Index fragments by operation/output item/content index; no tool executes
from a partial call. Preserve complete reasoning and assistant phase in the final
accepted item sequence. A completed item or TCP EOF is insufficient: require a
consistent terminal response.completed with valid usage. Failed/incomplete terminal
events, cancellation or missing completion settle without invented successful
usage or tool effects. A valid accepted helper can still fail its result grammar.
Keep inherited frame/aggregate/queue bounds and responsive cancellation, including
the terminal response's bytes; no complete-response exception permits unbounded
allocation. Distinguish accepted provider response from completed user task.

### Per-purpose differences and accounting

| Purpose | Existing API-key path | Selected ChatGPT plan path |
|---|---|---|
| Foreground | Existing captured output cap and delivery | Responses stream; remote output cap absent |
| Compression | Captured adapter/delivery; output cap clamped to 8,192 | Responses stream; remote output cap absent; same one-call submit_memory validation and no tool execution |
| Recall judge | Plain delivery; exactly 256 requested output tokens | Responses stream; remote output cap absent; same no-tool strict verdict validation |

Compression retains its 2 MiB request, 8 MiB received aggregate, 120-second
operation, per-attempt deadline and finite helper/request counts. Judge retains
its 262,144-byte request, 8,192-byte received limit and 15-second operation.
For its plan route the received limit counts the entire actual SSE stream,
including repeated terminal material and usage frames, not merely visible text.
Exceeding it is a local refusal with unknown usage under inherited fallback rules.
No larger limit or plain-delivery fallback is inferred to obtain a successful
verdict. Timeouts and cancellation cannot promise the remote service stopped
generating at an equivalent token or charge boundary.

Propose the explicit API-key Responses control uses the same item/tool projection,
store:false and streaming, with the existing applicable cap sent as max_output_tokens.
It does not change the legacy plain Chat Completions judge. Capture each choice
before rendering in the new capability; never encode absent plan caps as zero or
pretend a recorded legacy max_tokens was enforced. v1–v6 remain exact. The strict
v7 grammar must describe capability composition without automatically enabling
memory or recall, and include each purpose's route/funding/delivery/cap mode.

Normalize Responses accepted input into ordinary input, cache writes and cache
reads from its inclusive input total; validate the subtraction. Output includes
reported reasoning tokens, so do not add them again. Preserve terminal raw usage.
API-key prices remain dated estimates of accepted usage. Plan usage shows counts
and incomplete attempts; any API-rate equivalent is separately labeled and never
displayed as a subscription charge or zero cost. The selected route's errors stop
inference without automatic paid retry or credential fallback. Credentials stay
private; safe status/code/request-ID diagnostics must not echo arbitrary error
bodies into logs.

### Before student release

Coordinator review is needed for the added API-key Responses control, owner/store
lease choices, per-turn funding capture and replay capsule. Then publish exact
authorization/admission fixtures; tool/item/SSE/usage fixtures; all strict v7 and
public-command schemas; byte/deadline/lifetime checks; old-route equality controls;
and a finite live matrix. The matrix must include foreground tool continuation,
compression, judge, expiry/renewal local controls, two-Agent connection isolation,
actual own-app consent and the requested cache investigation. Account availability
or an unsupported capability may leave that real gate open; a separate probe
does not complete Ensemble integration. No authorization or provider run occurred
while preparing this proposal.

## Complete contract disposition after integration review

October 8, 2026. The coordinator accepted the proposal direction and every R1–R4
recommendation in `a5423ed`. The full manuscript now publishes the selected rules;
the earlier proposal remains historical reasoning rather than an unresolved vote.
No additional decision is attributed to Bill.

- R1: §§17.5/17.8 fix explicit capped API-key Responses and uncapped plan funding,
  whole-turn model/funding capture, per-purpose delivery/caps and no fallback.
  Responses initially supports off only; its implicit cache investigation remains
  mandatory. A source-backed capability table is frozen in v7 identity and exact
  price inputs are mount data. Unknown prices stay visible.
- R2: §17.8 assigns asynchronous Engine waits and Connection refresh, bounded
  isolated waiters, atomic dispatch invalidation and Actor once-only terminal
  settlement. Valid usage may survive canceled output; tool/memory/recall effects
  cannot. Restart uses a protected persistent random binding and generation, not
  the display name. Store lock, sign-out/close and late-work barriers are explicit.
- R3: §17.8 gives one authoritative raw terminal payload, indexed references,
  helper reuse of its existing raw receipt, protected bundle retirement and exact
  full-log/snapshot continuation. It prints ordered reasoning/commentary/call,
  spaced arguments, result and final-answer fixtures. Foreign namespace has an
  explicit controlled paired refusal; duplicate argument text retains Chapter 10.
- R4: §17.8 counts decoded HTTP-entity bytes before SSE parsing, including framing
  and terminal repetition, with E=532 bytes and exact 8,192/8,193 constructions.
  Actual successful plan judge/compression remain future validation gates.
- §17.5 completes v7 wrapper over every old base identity, construction order,
  session/standalone/anchor rules, capture and codec substitutions, record classes
  and retirement. Its minimal two-event construction fixture is printed. Old
  routes remain strict and exact; caching implies no memory/recall enablement.
- §17.7 specifies nested public status, uint64/rational representation, bounded
  truncation, protocol/human commands, manifest and safe client lifetime. §17.9
  gives the finite initial five-route matrix, actual-user surfaces and release gates.

The text-only/local-function Responses scope deliberately does not invent new
neutral attachment fields: unsupported human blobs refuse, and result references
use the inherited descriptive mapping. Optional GUI inspection remains the public
transport-independent MCP path. New credentials never grant model authority.

The full draft preserves the original-author `53ee8e7` opening, false instrument
incident, returning compact/indented lesson and producing-model meter story. The
new comparison/connection section starts with the reader's misleading-control
problem. Its extra specification length is needed to publish the newly selected
funding route; it is not a claim that documentation review produced a live result.

Next: independent full contract/schema/voice review; resolve material findings,
publish the actual new acceptance invocation, and release a fresh student only
after the preceding source and chapter gates permit it. No build or paid work
was performed by the author.


## Grouped C1–C4 correction after full review

October 8, 2026. The coordinator selected these details after reading the complete
`9598665` review and inherited Chapters 14/15. They are working design decisions,
not new Bill quotations. The original-author story passages are unchanged.

- C1: §17.8 adds the exact remove_responses list to v7/base-4–6 context_changed.
  A captured-policy, Actor-authorized handoff cuts every settled bundle without a
  model, including reasoning plus final text without calls. Automatic/selective
  cuts and compression retain whole bundles. Bases 1–3 gain no handoff permission;
  a different connection requires compatible history or a fresh session. The
  fixture specifies two calls/results, a no-call bundle, anchored survivors,
  full-log/checkpoint-tail equality and negative target/pairing controls. Neutral
  pressure counts the raw output-array span once, with a 127/128-byte control.
- C2: §17.5 gives strict retained/retired foreground capsule variants and checked
  base references. Retirement removes conversation/part indexes and converts usage
  to inline. Helpers keep Chapter 15 compact metadata and represented-memory
  references, never raw submit capsules in semantic snapshots. Suppressed usage
  closes the exact canceled response slot once without inventing a response.
  §17.8 supplies a hashed retired-record fixture and fold-A/fold-B controls.
- C3: §17.7 preserves opaque turn/operation strings, adds a coherent current profile
  and price-table date, and tests empty/post-change status independently from old
  attempts. Numeric counters retain their lossless decimal-string grammar.
- C4: §17.8 prints the documented assistant input_text working-note request and
  exact direct-caller object; real output-item spans retain their original shape.

The four-file correction is submitted for independent closure. The new checker
command and accepted Chapter 16 source remain separate release prerequisites.
No implementation, authorization, provider experiment or live result is claimed.
