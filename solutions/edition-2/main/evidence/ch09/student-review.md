# Chapter 9 student review — initial owner and API proposal

Status: planning only, awaiting coordinator contract review. Written in the fresh
Chapter 9 student session on 2026-10-08. No runtime, tests, catalogs, graders,
snapshots or teaching files changed. No agents spawned, credentials read, model
calls made, tests/grading run, commits made or pushes performed. This review is
the sole owned deliverable for this phase; it does not claim implementation or
Chapter 9 acceptance.

## Input identities and actual read ledger

Before working, `git rev-parse` established all three main trees below equal
`58d3fbf4ec6493839985c6968a13c1e6a7bdad83`:

- Current `HEAD:solutions/edition-2/main`.
- `edition-2-ch08-r1:solutions/edition-2/main`; the tag resolves to
  `bfdadaf54b7f43a5cdc1d7983f003c8a7baa644b`.
- Predecessor source `446d7f27fcef052d6bb780901c786842dc5286c1` at the same path.

Main's path-scoped status was clean. Before this file was written,
`git diff --exit-code edition-2-ch08-r1 -- solutions/edition-2/main` also passed
with no output. Outer HEAD observed during review was
`77eafd303e1cd6bd367e1b79eeed1a46367bda6d`; its newer outer work does not change
the accepted main tree. No broad history inspection was used.

I read the entire mandatory skill first, root `AGENTS.md`, and the entire
architecture ledger. Applicable ancestor instruction checks found no additional
files; an AGENTS-only enumeration found none beneath main. Direct-read SHA-256s:

| File | SHA-256 |
|---|---|
| `AGENTS.md` | `a6e88dd804e8aabce0823d0fc48c6653f89d94870c57cd7ec0be0f0bfdfc1b4a` |
| `book/edition-2/skills/ensemble-coding/SKILL.md` | `0180e3eb5d4f6032936729f31b473c70c421bce22001790a9709b05c57893bd1` |
| `book/edition-2/architecture.md` | `9dada72e65c36fb649659f9076b4ca2bdfc8379fef919a3b575652dca0e49327` |

I read `manifest.json` and all of `chapter-01.md` through `chapter-09.md` in
`/Users/bill/projects/ensemble-edition-2-revisions/ch09-student-inputs/`, including
Chapter 1's pre-example architecture and the full Chapter 9 contract. Large
combined outputs were initially truncated; I reread the affected ranges in
smaller chunks rather than counting those missing portions as read.

The manifest's SHA-256 is
`770cb3a4fe7614aac8ac740a97318b10136b4ebcc0f89a1b5e10197edc4ce6cd`.
Its teaching revision is `8736c951393320cf8beef849806bde3b90bc445f`.
A local byte comparison verified each extracted chapter both against its
manifest digest and the exact named chapter blob at that revision:

| Chapter | SHA-256 |
|---|---|
| 01 | `714beccbd6c11e6629c7577fdb078601d0fec0ce0e8c74888c6e87711409db47` |
| 02 | `698533c7881f89c5238b2cfa1a72f64b7e3869d1c4fc5f6c646da072cb982c37` |
| 03 | `85438fad9357d80b045b9dd62fb55d9abdbc7986403bcef9b0ad53c7b66e9214` |
| 04 | `b18b22663ceb4a9bbe74663a830423b0762afaee783e5f92b706c483d5289a46` |
| 05 | `1635746aed55332bb82a26df93158f5c5b2a1a6f7bf684fd25fcb0a64b656fb8` |
| 06 | `1d3b85fb83754d810c8923daddf6bf64fb0f6287fdd1b9fcdb8cd8ca7bbe8fc9` |
| 07 | `61f70c392b981c792701d805ade8b4db2418baf369f949fdae7c321821d01206` |
| 08 | `a2b02256f0b88bf72900268477f0cd44e9d6af7de05de3f2d91d7c88c99cd7f6` |
| 09 | `bc0d4a6f162e9dde53fb351c85512050ca0c2cf25d7500125e805b6eefea0b33` |

Accepted predecessor reads, all relative to `solutions/edition-2/main/`:

- Full files: `ensemble.go`, `watch.go`, `policy.go`, `go.mod`;
  `internal/common/{types,actor,watch,policy,json}.go`;
  `internal/llm/{actor,events,watch,render}.go`;
  `internal/tools/{registry,supervision}.go`; `internal/eventlog/log.go`.
- Full client/example files: `cli/{main,session}.go`,
  `gui/projection.go`, `gui/cmd/ensemble-gui/main.go`,
  `gui/web/gui/{connector,page,artifacts}.js`,
  `examples/policy-consumer/main.go`.
- Test portions actually read: all `watch_test.go`; `actor_test.go` lines
  360–451 and 556–605 (request reconstruction and all-adapter hint ordering);
  `events_test.go` lines 201–294 (instruction roles, owned configuration,
  deferred human/result projection). Scoped searches also inspected test
  names and matching lines in `ensemble_test.go` and `boundaries_test.go`;
  those two files were not read in full.
- Scoped source inventory and identifier searches located the above seams.
  Earlier evidence/support was not needed and was not opened.

Isolation account: no excluded source contents were read. Listing the supplied
input directory exposed filenames `launch-plan.txt`, `plan-events.jsonl` and
`plan-stderr.txt`; none was opened. Historical descriptions and links embedded
in the permitted teaching/mandatory documents were visible, but no historical,
author/reviewer, future-chapter, worker-log, memory or checker link was followed.
The black-box command named in the teaching/manifest was not run or inspected.

## Owners, facts and lifetime

All shared declarations below belong in core `internal/common`, with public
aliases at the library root. The new implementation spoke is `internal/skills`.
The root remains the only composition site. The CLI and optional GUI continue
to construct/use one public Ensemble; neither wires internal services.

| Owner / parent route | Facts it owns | Changes, locks and lifetime |
|---|---|---|
| Ensemble; application root | Logger and existing Agents/handle allocation | Existing application synchronization and close; no global skills registry |
| Agent → Ensemble | Creation configuration: skill mode, source selection, primary, copied scalar bindings, installed ceiling (`Builtins`), resolved workspace; existing configuration/history/log | Creation inputs copied before publication. General config checks reject changed creation facts before mutation. Existing short config/history locks remain; no lock held while calling a mailbox getter |
| Skills → Agent → Ensemble | Frozen parsed definitions; one committed ledger of roots, active IDs, retired IDs, immutable activation records, revision and next activation identity | Construct and parse before exposing Agent. After initialization only actor-admitted commit changes the ledger. No background skills worker, publication callback or mutable Registry mirror. Retain material for Agent/history lifetime, including retirement |
| Actor → Agent → Ensemble | Mailbox order, skill-candidate acceptance, persistence ordering, tool admission, existing turn/pause state | One mutation path. Preparation is bounded memory work; no catalog I/O on load/unload. Queue lock guards queue transfer only. Getters and watch read at an actor boundary |
| Registry → Agent → Ensemble | Installed handler/schema metadata and JSON argument decoding | Installed entries fixed for this Agent. Filter declarations and decide admission using the same committed Skills authority; no second grant map. An admitted worker does not recheck a later revocation |
| Engine → Agent → Ensemble | Existing request operations, transport and usage | Each HTTP operation owns its captured declarations. New requests refresh declarations from current grants; model configuration and turn policy retain their existing captures |
| Jobs → Agent; Job → Jobs | Existing jobs, limits and report cursors | Management tools allocate no job/artifact. Model management attempts consume pending limits through Tools/Jobs; public typed controls do not. Unload leaves admitted jobs alive |
| EventLog → Agent → Ensemble | Append-only serialized bytes | Successful transition append commits. Initial write failure returns no usable Agent. Later failure faults/cleans up under existing rules without invented rollback |
| GUI Server → public application; Page → BrowserApplication | Safe projections, connection lifetime, local cards/input/speech | Optional module only. Extend known lossless counter paths. Existing page and shared native speech ownership stays intact |

Roots are intent, closure is derived reachability, contributors are reasons for
grants, and material is recorded instruction history. Effective names, available
offers, contributor lists, safe state, Config's effective System/Tools view and
request declarations are derived snapshots, never independently mutable policy.
The primary System returned by Config is a view of its committed activation;
ordinary loads never rerender it. Earlier material keeps its exact bytes/hash.

Proposed synchronization: Skills' live ledger is confined to actor execution;
there is no Skills mutex whose holder waits for the actor. Private parent
adapters expose actor-only preparation/apply operations, while public methods
enqueue and copy results. Initialization uses this same serialized transition
path before Agent publication; public controls cannot initialize an exposed
Agent. In-memory candidates and snapshot values have no independent goroutine
or close protocol. Failed/no-op candidates discard prospective IDs. Close
rejects new controls, settles admitted controls, and uses existing actor/job/log
cleanup. Post-close mutations return stopped/read-only errors.

Offline load creates historical skill facts without a catalog or a live
capability service. A common record-validator interface implemented in skills
performs pure validation/reduction over supplied historical values, reached
through Agent. It does not consult live grants or expand variables. Context's
skill entries and historical projection are derived conversation data, not an
alternate runtime authority. Request reconstruction uses a fresh local historical
projection, never the currently loaded Agent's final grants. All helpers retain
their responsible owner interface for diagnostics.

## Proposed public integration spelling

These are proposed declarations, not installed APIs. Keep existing constructor
`(*Ensemble).NewAgent(Config) (*Agent, error)` and add this Config field:

```go
// Config.Skills *SkillConfig; nil selects preceding no-skills behavior.
type SkillConfig struct {
    Directory string
    Catalog   map[string][]byte
    Primary   string
    Variables map[string]string
}
```

Exactly one source: a nonempty Directory or a non-nil Catalog. An empty provided
catalog cannot resolve a primary. Copy all map/slice inputs. Resolve a relative
Directory once against the captured Agent workspace; CLI/GUI workspace is their
launch workspace. `Config.Builtins` remains the sole installed-handler ceiling;
add no parallel ceiling setting. Skill-enabled callers explicitly include
`load_skill` and `unload_skill` in it. CLI/GUI install those two only in skill
mode, retaining their inherited ten-tool selection otherwise. `Config.Tools`
remains a derived/matched declaration input, never an independent grant.

General Config round trips compare skill creation values semantically, without
rereading files. Empty System retains the primary; the unchanged primary System
returned by Config is allowed; another nonempty value is refused. At initial
construction *any* competing nonempty System is refused, before defaulting.
Live `SetConfig` retains the predecessor's active-turn restriction. The narrow
typed controls below are the permitted in-flight capability changes.

```go
type SkillState struct {
    Revision  uint64                `json:"revision"`
    Primary   string                `json:"primary"`
    Roots     []string              `json:"roots"`
    Active    []SkillActive         `json:"active"`
    Available []SkillOffer          `json:"available"`
    Tools     []string              `json:"tools"`
    Retired   []SkillRetired         `json:"retired"`
}
type SkillActive struct {
    Name       string `json:"name"`
    Type       string `json:"type"`
    Activation uint64 `json:"activation"`
}
type SkillRetired struct {
    Name       string `json:"name"`
    Activation uint64 `json:"activation"`
}
type SkillOffer struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}
type SkillActivation struct {
    Activation   uint64       `json:"activation"`
    Name         string       `json:"name"`
    Type         string       `json:"type"`
    Body         string       `json:"body"`
    SHA256       string       `json:"sha256"`
    Tools        []string     `json:"tools"`
    Dependencies []uint64     `json:"dependencies"`
    Offers       []SkillOffer `json:"offers"`
}
type SkillMaterial struct {
    Record     SkillActivation
    EventSeq   uint64
    Retired    bool
}
type SkillContributors struct {
    Tool        string
    Activations []uint64
    Mandatory   bool
}
type SkillInspection struct {
    State        *SkillState
    Contributors []SkillContributors
    Material     []SkillMaterial
}
type SkillResult struct {
    Status   string `json:"status"`
    Name     string `json:"name"`
    Revision uint64 `json:"revision"`
    Changed  bool   `json:"changed"`
}
type SkillError struct {
    Code     string
    Name     string
    Revision uint64
    Detail   string // Safe static detail; Error() implements error.
}

func (a *Agent) SkillState() (*SkillState, error)
func (a *Agent) InspectSkills() (SkillInspection, error)
func (a *Agent) LoadSkill(name string) (SkillResult, error)
func (a *Agent) UnloadSkill(name string) (SkillResult, error)

func (e *Ensemble) SkillState(agentID string) (*SkillState, error)
func (e *Ensemble) InspectSkills(agentID string) (SkillInspection, error)
func (e *Ensemble) LoadSkill(agentID, name string) (SkillResult, error)
func (e *Ensemble) UnloadSkill(agentID, name string) (SkillResult, error)
```

Expose these Ensemble methods through the public client-owner interface too.
No JSON tool arguments are needed to use them. Live getters run in actor order;
offline loaded Agents return their owned historical inspection directly and
refuse mutation. No-skills state is `(nil, nil)`; inspection has nil State and
empty arrays. Typed operations return `*SkillError` with `skills_disabled` in
no-skills mode. Graph/argument failures use the published error codes and current
revision; transport/persistence/stopped errors remain ordinary lifecycle errors.
No-op returns unchanged/current revision with nil error. Errors do not carry a
success result. Tools encodes the contract's separate compact error shape from
SkillError, excluding Detail, or the exact SkillResult acknowledgement.

Inspection is one coherent cut. Contributors sort by tool, and activation IDs
sort numerically; they identify active recorded contributors. Mandatory=true
marks the always-on management grant so it needs no fabricated activation 0.
Actual explicit contributors may also appear for those names. Material sorts
by activation, includes the primary and retired records, and requires no file
access. Record bodies remain immutable; Retired is derived current status.
Every returned nested slice/map owns its storage. Safe SkillState alone goes
in `WatchState.Skills *SkillState` with `json:"skills"` (no omitempty).

Internally, common declares `Skills`, its Agent parent, typed `SkillOperation`,
`SkillCandidate`, recorded ledger and transition values. Proposed operations
are `Skills.Agent()`, `Prepare(SkillOperation)`, `Apply(SkillCandidate)` and
owned inspection; an actor-only Agent adapter supplies `Skills()` and a
candidate-aware durable commit entrypoint. A separate record-reduction method
accepts explicit prior historical facts. Tools adds typed decoding/classification
for management calls; llm never imports tools or skills implementations.
The precise private method set can be tightened during reviewed implementation;
none permits clients to apply candidates or inject grants.

## Commit and projection decisions

1. Parse all frozen definitions under the published byte/string/list rules.
   Resolve only relevant graph references/handler availability, including offer
   ID/type validity; leave unreachable bad dependencies diagnosable at load.
   Candidate postorder is primary first, then sorted explicit roots, sorted
   dependencies first. Expand only newly activated material once over the full
   candidate. Precheck the whole ID group and next revision before any append.
2. Tools classifies and strictly decodes exactly one management `name` field,
   rejecting duplicates. Actor checks pause/current permission, consumes next-call
   limits through Registry, records the matched call, prepares/commits locally,
   then records the short result. Invalid/no-op/refused calls still consume the
   one-shot setting; no management path creates a Job, handle or `cr/io` file.
   Public controls use that same candidate/commit logic directly, bypassing only
   model-call pairing, pause and next-call-limit consumption.
3. Append success precedes application of the same candidate, conversation
   projection, state publication and caller success. Readers/watch cannot cut
   between those steps. A failed transition append retains old authority;
   a failed later result append retains the already committed transition.
   Registry does not receive an asynchronous update after commit.
4. Refresh effective declarations at each request preparation. Check grants
   immediately before each call admission, including returned calls from older
   HTTP requests. Preserve serial batch behavior: load→write can succeed and
   unload→write can refuse in one batch. Already admitted ordinary work finishes
   under its captured permission, without a worker-side revocation check.
5. Add typed `Event.Skills` and exact initialize/change payloads. Validate every
   transition against prior facts, including closure, offers, unchanged ceiling,
   mandatory names, monotone identities, immutable records and exact hashes.
   Live public append additionally matches the frozen-catalog candidate; it
   cannot initialize exposed Agents. Do not treat a plausible snapshot as proof
   of a legal transition. Retain all arrays, including empty ones.
6. Add neutral system/skill entries with name and activation identity. They are
   derived from material facts, never `message_received` instructions or tool
   result text. Keep primary in the base system position; initial dependencies
   precede dialogue. A dynamic batch's entries wait for every matched result,
   then appear once in transition order with the exact envelope/newlines.
   Pending hint merge needs the clarification below. Empty bodies have identity
   but no provider text. Redaction only changes result parts. Unload marks
   retirement without deleting chronological instructions.
7. Skill-specific replay validation/graph reasoning stays in skills behind common
   interfaces; neutral conversation placement/rendering stays in llm. Offline
   paths skip catalog discovery, variable expansion, credentials and workers.
   Historical non-skill logs retain their previous meaning. Preserve captured
   request schema/delivery/provenance for reconstruction. Do not accidentally
   impose the inherited 16 MiB JSON-line limit on legal skill material merely
   because JSON escaping expands the chapter's decoded-byte allowance.
8. One committed change should produce an actor-published skills_changed state
   observation carrying current SkillState; the durable event supplies material
   cards. Prefer enriching its existing durable observation with Skills rather
   than publishing duplicate same-kind notifications. Watch revision is distinct
   from skill revision. Snapshots capture current Skills regardless of the last
   100 events; recent initialization/change events become renderable cards.
9. Extend Connector's typed counter handling for state revision and every
   activation/dependency identity in skill records, without touching arbitrary
   body or tool argument values. Card keys use Agent plus exact activation.
   Chat cards safely show retained text and update retirement status in place;
   management calls/results remain in actions. Reconnect/autoplay never speaks
   manuals, while explicit Speak uses the existing owned speech queue/service.
   `/skills` reads state locally with no model request. Default protocol stays
   unchanged; no browser source selector, credentials or full Config exposure.

## Teaching questions for coordinator/author review

Line numbers below refer to the extracted Chapter 9 at the pinned teaching
revision. These are initial interpretation questions, not defects established
by an implementation or by a checker. I have not used excluded sources to fill
them. Hold the affected design choices until reviewed.

1. **Which environment variable supplies a competing system override?**
   §9.8, lines 595–599: “Use the shared configuration reader for both commands.
   In skill mode it leaves the default System unset so the primary supplies the
   base; an explicitly supplied nonempty system override is an error, not
   discarded input.” The accepted shared reader reads no System environment
   variable. Proposed interpretation: add explicit `LLM_SYSTEM`, distinguish
   presence through `LookupEnv`, and add no provider-specific fallback spelling.
   Public callers already use Config.System. Please confirm the intended CLI
   spelling or state that this sentence concerns only a public configuration
   input. Consequence: silently ignoring an operator's attempted override would
   violate the refusal promise, while guessing a variable invents a user contract.

2. **Does recorded-once also exclude the existing per-request System capture?**
   §9.7, lines 447–448: “The primary body occurs only in its activation record;
   the initial event does not contain a second copy in a message or Config
   payload.” Lines 506–507 also call the base “recorded once.” The accepted
   actor records `RequestConfig.System` on every request. Proposed interpretation:
   skill-mode request captures leave that field empty (or omit it under an
   explicitly taught compatible shape), and reconstruction resolves the primary
   from the recorded activation. Config() still returns the established primary,
   and enduring supplements stay in instruction events. Please confirm whether
   the prohibition spans later request captures as well as initialization, and
   which captured-field representation to preserve. Consequence: blindly retaining
   the predecessor capture repeats the manual; deleting it without a replay rule
   risks changing historical request bytes or accepting a competing System.

3. **How far does post-batch hint/material sequence merging extend?**
   §9.7, lines 512–516: “A transition during an unresolved accepted batch records
   its fact immediately, but its material is projected after every call in that
   batch has a matched result, in transition order. This includes typed public
   changes during a held batch. Preserve Chapter 5's ordered post-batch material:
   merge deferred skill entries and hints by their durable event sequence after
   the complete results.” Chapter 5 §5.3 says consumed hints leave enduring
   dialogue, and its Gemini explanation says a pending ordinary prompt precedes
   hints. The predecessor appends pending hints at the request tail, whereas
   skill entries now persist at a chronological location. Proposed interpretation:
   keep a derived post-batch placement anchor for skill entries and unconsumed
   hints, sequence-sort material tied to that batch, and remove hints when a
   request consumes them without moving retained manuals. Please supply/confirm
   the ordering for hint H, then skill S, then completed results, followed by
   interruption/round-limit before another HTTP and a later prompt P. Is the
   later request's order results→H→S→P, results→S→P→H, or another specified
   projection? Consequence: a global sort would split call/result groups or
   relocate unrelated history; always-tail hints would reverse the expressly
   taught H/S order. This needs a literal fixture before affected rendering code.

## Initial teaching impressions and later gates

The installed/enabled/remembered separation, the narrow graph example and the
exact activation event make the intended authority much easier to reason about.
The single-pass scalar example prevents the common mistake of expanding inserted
values recursively. Exact byte budgets, full uint64 values, no-op semantics and
append-failure boundaries give useful implementation decisions without prescribing
private identifiers. The Chapter 1 rules provide a clear reason to add a skills
spoke and parent interfaces immediately rather than attach services to Engine.

The hardest integration is the distinction between durable event chronology and
provider-valid projection. The questions above are where an extra cross-chapter
example would most help. I have not tried a failing implementation or received
coordinator feedback yet. One minor editorial observation: Chapter 5's closing
paragraph promises skills next, while the supplied Chapter 6 teaches streaming;
it does not change this Chapter 9 contract.

After review and explicit continuation, deterministic controls should derive
from §§9.3–9.10, preserving predecessor regressions: full candidate atomicity,
forced disabled effects, same-batch permission changes, frozen catalog after
file changes, copied inputs/outputs, immutable material/redaction, replay without
source files, exact counters/whole-group exhaustion, held HTTP/pause controls,
durable failure, and browser identity/safe/silent cards. No acceptance results
are claimed here. No baseline failure was observed because tests were not run
in this planning phase.

Before any paid work I will submit a separate feature-to-live-action matrix for
review. It must cover the shipped ensemble primary and narrow graph, management
success/no-op/refusal and limit consumption, actual file effects/revocation,
same-Agent CLI/browser/reconnect/manual cards, and a public two-Agent consumer
with distinct ceilings and literal PROJECT bindings. Human CLI PTYs, real browser
and public demonstrations must cover all three vendors within reviewed bounds,
with source/executable/catalog identities and sanitized receipts. Forced races,
malformed graphs and persistence/counter faults remain local controls. No claim
that fake-only success or a model's assertion completes the live requirement.

This session stops here for coordinator review. The review file is left
uncommitted; implementation and the later validation/live gates remain pending.
