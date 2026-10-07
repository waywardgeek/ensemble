# Second edition: Chapter 2 outline

Status: human-client validation reopened. Prior contract validated at
`cc1bec45c3327c87728a4040f762155d8e860a0b` after independent comparison
and revision. All-three-provider machine-CLI/public-consumer
receipts retain their source binding at initial checkpoint
`39a92ca27a418712832ac0dcbbfbe4e32b3bca35`. New human-mode paid runs
are separately bound to `56dacfad01f71f2a1b20d39bca15846edef41ddd`.
Bill's editorial approval remains separate. See evidence for the pilot's
inherited-context qualification.
Student handoff includes the formal sections as well as the TL;DR.

Bill requires actual human chat now, not JSON-lines input described as chat.
The published §2.7 contract adds explicit chat/protocol modes, terminal
default selection, visible prompts, ordinary text, readable answers and
usage, discoverable history/redaction, and precise command/input failure
behavior. New all-three-API terminal receipts are reconciled into §2.10;
independent final review/acceptance remains required. No old receipt is
relabeled or attributed to Bill's participation.

## Voice plan

- Stake: a reader must be able to change provider without rewriting the
  meaning of the conversation or losing the evidence needed to reconstruct it.
- Register: mechanism first; short documented example of a provider-shaped
  interface becoming three clients if the origin story earns its space.
- Revised motivation: begin with the reader's need to explain a bad answer,
  using the concrete `port=8080` result before/after redaction. Derive log,
  context, and request from that debugging problem before presenting schemas.
- Bill moment: no new anecdote. Historical interface story is available
  in old chapter 2 but its size/timeline claims will not be recycled as
  fresh measurements.
- Confession inventory: per-call opaque fixture passed vacuously in the
  original grader, documented by `5a7dfca`.

## Thesis

History records facts, context reduces them, and each adapter renders a
request from that context. There is one meaning of the conversation and
three wire projections. Keep the Chapter 1 ownership architecture while
replacing its intentionally narrow text-message representation.

## Teaching order

1. Plain-English distinction among log, current context, and request.
2. TL;DR and command contract; preserve prior user-facing behavior.
3. Common data declarations: validated events, attributed entries, typed
   parts, references, provenance, current configuration, and usage facts.
4. One append/apply path; monotonic sequence, immutable owned bytes,
   version refusal, deterministic replay, and explicit pending ephemera.
5. Three rendering/parsing adapters; no provider types outside that seam;
   exact raw values where opaque replay requires them.
6. Token normalization with actual route provenance; no current-model
   repricing of historical counts and no cost claim from unpriced totals.
7. CLI and optional GUI stub as clients of one Ensemble; public requests
   and observations, separate GUI module, no GUI transport in the core.
8. Exercise and property-level audit; inherited and additional checks.
9. Taking it for a spin from actual CLI runs on all three vendor APIs,
   external consumer, replay/redaction/ephemera user paths, and explicitly
   labeled GUI-stub integration evidence.

## Scope and ownership

Start from the exact accepted Chapter 1 source in `solutions/edition-2/main/`,
under the outer repository's history. Export `ch02/` only from a validated
source version; do not develop inside that frozen directory. Keep its public
library root and CLI. Add an
optional GUI module beneath the snapshot with its own `go.mod`; the core
module must neither require nor import it. A public transport-neutral
client seam is real; WebSocket/browser transport may remain a documented
stub. No claim of a working browser UI.

Ensemble owns Agents, the logger, and client registration/routing. Agent
owns current configuration, its immutable event log, and current context.
Engine owns its HTTP client and per-provenance usage accounting. One event
application path advances both context and accounting. The reducer and
adapter behavior remain free functions in `internal/llm` over shared
data; no behavior warehouse in common.

CLI and GUI stub submit requests through the same public services and
receive public observations carrying Agent identity and event sequence.
Observer is not request completion: a request receives its own returned
answer/error. There is no actor scheduler or streaming transport yet.

## Required mechanisms

- Three non-streaming API surfaces: Anthropic Messages, OpenAI Chat
  Completions, Gemini generateContent. Current official sources and live
  availability determine model IDs. No fixed recommended ID.
- An append-only JSON-lines log, strict sequence/payload validation, safe
  unknown-kind/version refusal, typed provenance, and captured owned bytes.
- Text, call, result, blob-reference, opaque, and redacted parts. Preserve
  present empty text, order, call identity, and part-bound replay material,
  including signatures on visible text. Thought-marked parts remain opaque
  and do not become ordinary CLI answer text.
- An entry kind distinct from actor: dialogue, instruction, ephemeral.
  The type can gain future kinds with future schema support; do not
  implement skills/recall now or silently accept unknown kinds.
- Redaction by inclusive sequence span. Implement result-content stubs
  preserving call/result pairing and reference locators. Define the
  extensible level vocabulary without requiring later compaction policy.
- Ephemera recorded but excluded from durable dialogue; attached to one
  request and consumed by its recorded request event, not by rendering.
- Disjoint normalized usage plus vendor/model/surface for each response.
  Store raw usage observations too; aggregate by provenance on Engine.
- Offline render/dump commands that never require credentials or HTTP.
  Determinism holds for a fixed log, explicit config, and adapter version.
- External headless consumer, two independent Agents, optional GUI stub
  compile/integration via the public seam, and logger reachability.

## Prior checks and new checks

Existing chapter-2 source currently awards parity 10, logdump 5, replay
10, redaction 10, ephemera 10, usage 10, seam-render 15, seam-parse 15,
ref-roundtrip 3, ref-oldformat 2, ref-zerokind 2, ref-render 4,
ref-redaction 4. These total 100. The old chapter's parity25 table is stale.
The session gate is separately required. Preserve legacy checks and their
mutation tests; do not change their meaning to accommodate new design.

Add acceptance for sequence/payload/version rejection, no aliasing of
recorded bytes, model attribution after route change, no credential logs,
same incremental/replayed projection, mixed-model usage, entry kinds,
repeated offline render leaving ephemera unconsumed, actual ownership and
imports, both public clients, and headless optional-module independence.
Publish each assumption before it becomes a grader requirement.

## Contract choices now printed

Canonical schemas, independent literal tool/redaction/opaque fixtures,
CLI directives and acknowledgements, explicit log selection, provider
environment precedence, and usage formulas are in §§2.2–2.9. Empty and
tool-only answers are represented honestly; unresolved calls block a new
request until supplied results complete them. No tool dispatch is added.

Request and returned model identities are recorded separately. Usage keys
the returned provenance, with an explicit fallback/reporting flag when
the provider omits its model. Opaque compatibility uses exact target
identity or explicitly configured `LLM_RESOLVED_MODEL`; no inferred alias
matching. Foreign standalone opaque parts are omitted; incompatible
call-bound material causes a render error.

Failed attempts remain in the log while their unmatched inputs leave
future request context. A recorded request consumes its named ephemera
even on failure; offline rendering consumes none. One replay application
path updates Agent context and Engine accounting under their actual owners.

Independent review resolved the material findings, and the student's cold
read clarified accepted message actors and permanent fault after a failed
log write. After implementation/live use, require the newly
mandated first-edition quality comparison and resulting code/teaching
revisions, in addition to grading and proofread evidence.
