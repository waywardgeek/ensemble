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


## Implementation phase: release-8f24360

Coordinator accepted the initial ownership/API plan and preserved it at 5ac45e4. Before integration I implemented only common skill declarations and the skills parser/candidate layer. Baseline `go test ./... -count=1` passed in main; `go test ./internal/skills -count=1` passed after this layer. New Go files were formatted. These are local checks, not chapter acceptance. No credentials or providers were accessed.

At the layer boundary I read the authorized coordinator inbox message `release-8f24360`. I read the entire mandatory skill again after compaction, retained architecture, the complete released Chapter 9, Chapter 5's corrected closing passage, and the expressly permitted direct author response. Manifest revision is `8f24360afde19ba848be225c6106a89949deec8c`; all three released file SHA-256 values match the manifest. No links to excluded sources were followed.

The published changes resolve all three initial questions:

1. §9.8 names `LLM_SYSTEM`, read with LookupEnv in the shared command reader. Nonempty explicit system conflicts at skill construction; absent/empty keeps the primary. No-skills normalization retains its prior meaning.
2. §9.7 explicitly persists the request capture's system field as the empty string in skill mode. The recorded primary supplies rendering/reconstruction; Config() may expose the effective primary and allow its unchanged round trip. No-skills request captures remain unchanged.
3. §9.7 gives pending hints and skill material stable completed-batch anchors, derived from events. The literal three-provider fixtures settle results→H→S→P, then results→S→P after hint consumption. Unanchored hints keep the prior tail rule, and no-skills placement is unchanged. I will implement these anchors as derived dialogue placement only, never capability authority.

The explicit fixtures make this integration substantially easier to reason about: request capture, public configuration and chronological dialogue are now distinct obligations. The author response resolves the questions without requiring an excluded implementation. I am proceeding with full local integration and contract-derived checks; the pre-provider review gate remains in force.


### Boundary response: architecture-1

Read inbox before the actor/integration layer. The coordinator's narrow review found my candidate's concrete `*Service` creator pointer. This was my implementation mistake under the already explicit skill, not a teaching gap. Changed it to `common.Skills`; private copying functions take that owner interface. Application still checks the creating interface identity and prior ledger pointer before applying. No additional runtime root or injected sibling service was added. Recorded-transition positive and tamper-refusal tests pass locally before integration.

### First integrated local gates (incomplete implementation)

Core `go test ./... -count=1` passes after adapting the two predecessor test doubles to the added common interfaces. Initial CLI black boxes against `/tmp/ensemble-ch09-student-cli` SHA-256 `aafb5132aa6cf2a1e4bf4db6e13f285423a6b31ccf51f8a23964d85f826fd866`: base 51/51, graph 27/27, management 138/138. These checks exercised local fake transports only.

Three diagnostic follow-ups for the coordinator/independent checker maintainer (no checker source read or edited):

- Catalog: 64/65. `render-exact-65536` refuses with `unknown skill variable`; its corresponding one-over case also refuses with that same cause. My own exact expanded-body boundary control passes. Please provide the failing fixture's source/token construction or inspect whether padding was concatenated to an unbraced `$TOOLS`/`$SKILLS` token; the contract requires longest lexical token consumption. I have not relaxed it to make the gate pass.
- Replay: 0/105 because all three `valid-offline-chronology` positive parents fail. Output supplies no CLI diagnostic or fixture excerpt. Base gate's actual CLI-log replay passes all three providers, and my recorded-transition replay passes locally. Please expose the positive fixture's CLI stderr/failed comparison in black-box output so I can distinguish my bug from fixture mismatch without reading checker source.
- Public: all reported tests pass except `TestCh09PublicAnchoredHintSkillPromptAndReplay`, where all six skill/no-skills variants fail at line 350 with `hint requires nonempty text and an active turn`. Held-HTTP revocation, admitted-job survival, owned copies, creation config, two Agents, frozen directory, pause/no-op publication, concurrent single commit, scalar limits and frozen public append passed. The failure in the unchanged no-skills path suggests a fixture timing or inherited-behavior assumption to inspect; please supply bounded diagnostics. I will independently test the published H/S/P fixture, without changing prior hint admission semantics.

GUI and complete student contract controls are still in progress; these scores are not an acceptance claim.


### Boundary response: editing-procedure-1

Read the coordinator reminder before the GUI layer. My earlier candidate-parent correction used unasserted replacements, contrary to the mandatory editing rule. I inspected the resulting candidate/copy helpers directly and ran the affected skill suite and full core suite successfully. Subsequent scripted replacements assert their unique anchors before each write, and targeted apply_patch is preferred. This was an editing-procedure mistake, not a runtime/teaching ambiguity.

### Boundary response: checkers-ddd0704

Read and accepted the new black-box command authorization. Ran browser checker after GUI integration. Its inherited and safe-card/speech controls pass, as does lossless-source absence refusal. Two controls fail: `current-state-beyond-window-and-independent-page` reports fabricated out-of-window body cards, and `uint64-snapshot-material-retired-identity-and-opaque-boundary` reports adjacent unsafe activations sharing identity. Please expose observed card keys/titles/counts and incoming counter lexemes in bounded output; my own browser component control retains distinct `skill/agent-x/9007199254740992` and `skill/agent-x/9007199254740993` keys, and snapshot current state only updates existing cards. The inherited omitted-history notice is itself a card; a fixture counting all cards rather than material cards could conflate those. I will not change unrelated notice semantics or counter identities without diagnosing the actual mismatch.

Student local browser script `evidence/ch09/browser-local.mjs` passes: initial read-only sidebar; actual GUI load/write/unload over six local fake requests and verified file bytes; retained safe HTML-as-text manual; keyboard expansion; retirement/reconnect without duplicate card; separate unsafe uint64 card keys and dependency value; explicit speaker receives full retained text; automatic skill observation stays silent; arbitrary argument activation/revision fields remain ordinary JSON data. This is controlled local testing, not real-provider or audible native-speech evidence. The Chrome/Playwright invocation pattern was read from the permitted preceding student support `evidence/ch08/browser-local.mjs`.

Student public tests also pass the published H/S/P chronology for all three renderers, repeated-render purity, hint consumption, result-only redaction, retirement, exact captured-request reconstruction, catalog-absent replay, explicit competing System refusal, paused model management versus typed controls, and interruption preserving a typed commit. Found and fixed my own UTF-8-copy bug: creation scalar validation must precede generic JSON cloning, which otherwise replaces invalid string bytes. Added a regression. No new teaching ambiguity arose.

### Boundary responses: diagnostics-in-progress, fixture-diagnosis-1/2, live-plan-review-1

Read all messages at the next boundary. The brief Go-edit hold and its lifting were both already in the inbox; no edits were made during an observed hold. I retained the original failed findings above. The independent engineer reports the five failures were fixture defects: unbraced token padding, a second unanswered replay prompt, a hint sent after turn end, and two browser counts including the inherited omitted-history notice. Their reported corrected positives (catalog 65/65, replay 105/105, browser 5/5 and all six targeted H/S/P leaves) are independent reports, not my own reruns. Waiting for committed checker identities before my affected reruns.

The coordinator accepted live-plan coverage and the hard 108 overall/36 per-provider maximum as a ceiling, not a target. Gemini must be freshly discovered as `models/gemini-3.8-flash` with no silent fallback. Paid launch remains held for source-bound local and support review. I am making the redaction/reconstruction distinction and support bindings concrete; credential access remains unperformed.

Concrete existing local coverage for the requested gaps:

- `TestSkillDurableCommitFailureBoundaries`: fail `skills_changed` append versus fail later `tool_returned`; verifies complete prior state preservation versus durable committed state retention. `TestSkillInitialWriteFailureDoesNotApply` tests the unpublished construction append/apply seam.
- `internal/skills/skills_test.go` owner-state counter seams cover maximum revision no-op/refusal and whole-group activation fit/refusal (not fabricated exhausted logs).
- `TestSkillPauseHoldsManagementButTypedControlsProceed`: paused model control has no call/job yet; typed load commits; interrupt pairs the pending attempt and preserves that commit.
- `evidence/ch09/browser-local.mjs`: launches actual GUI executable plus local fake HTTP, uses WebSocket server projection into Chrome, verifies state/cards, actual file effect and reconnect. Synthetic uint64 component checks are separately labeled.
- `evidence/ch09/deletion-controls.py`: temporary Go overlays preserve delivered source. Passing controls precede admission deletion (specific forbidden file creation), result-redaction/manual deletion (retention comparison fails), and incremental graph publication mutation (complete ledger differs after failed second branch).
- `TestSkillWatchCurrentStateOutlivesMaterialWindow`: 105 later events leave a 100-event window with no transition bodies but exact current skill state; owned watch values cannot mutate authority. Browser reset with only current state yields the inherited omitted-history notice and no fabricated material card.
- `TestSkillEscapedDurableRecordReplaysBeyondOrdinaryLineLimit`: a valid 44-manual record with escaped `<` bytes exceeds 16 MiB of wire JSON and replays exactly. During self-review I found the inherited Scanner limit would reject valid Chapter 9 decoded-byte budgets. Skill records now read complete lines while ordinary/header records retain their prior limit. This is an implementation integration fix, not a new teaching rule.

Core and GUI `go vet ./...` and `go test -race ./... -count=1` passed before the later escaped-record test/fix; final affected checks will rerun. Vet initially caught my unkeyed external test literals; those are corrected. Public headless `examples/skills-consumer` builds/vets, and its local default invocation proved two ceilings and literal alpha/beta bindings with zero model requests. `--ask` is the explicit later bounded-model path.

### Boundary responses: checker-release-a294aac and record-bound-review

Read committed checker release a294aac; affected black boxes can now be rerun in full. The original fixture failures and corrected causes remain above.

The coordinator identified a real resource regression in my escaped-record fix: `ReadBytes` allows unbounded allocation before ordinary-line checks, while skill-bearing lines have no raw limit. The valid escaped-material positive remains required. I am holding only this reader/bound decision pending the author's published clarification, since a raw cutoff must account for repeated offers and growing retired summaries, not merely body bytes. I will not guess a limit or revert exact decoded-limit acceptance. This is a newly exposed contract integration question; no historical source was consulted.

The independent reviewer reports passing deletion controls and additional failure/counter/interruption mutants. Those are independent reports. My own final whole-delivery module receipt will be `evidence/ch09/module-checks-initial.json` (with exact per-file input hashes and each module's vet/test command/result). No paid launch has occurred.


### Boundary response: release-d8c7738

Read the complete newly pinned Chapter 9, manifest, and direct author feedback at `clarification-d8c7738`; SHA-256 values match `40d2f81bfc93e2c70e433f3689b750fba283f49d04eaa51485551ec282394744` (chapter) and `ab168ba11211758f8f0c9f7e12c66e42f0bcbfbec5a1b575db33fece6c08d73b` (feedback). Revision `d8c7738f206fb5c95cfe9fb53d70f64e04f8884e` resolves the whole-fact size question. The explicit 67,108,864-byte physical record budget includes LF when present, while final EOF remains accepted. Controlled write refusal and bounded import are separate obligations. The author acknowledges that the missing whole-record bound was a teaching omission; my unbounded reader was still a resource regression and is now replaced.

Implemented incremental ReadSlice accumulation with capacity clamped to the raw budget, overflow refusal before copying excess, and inherited ordinary/header limits. Compared the permitted accepted predecessor's exact Scanner implementation to preserve LF versus unterminated-token boundary behavior. The root measures actual sequence/time-bearing serialized facts before append, clone/application or authority mutation. Oversize yields typed skill_too_large and leaves the storage healthy. Eventlog also guards direct writes. No new limit was invented.

Local controls now pass: exact raw 64 MiB and +1 both with LF and at final EOF; fixed-buffer overflow stops reading rather than draining an unlimited source; ordinary/header limits; genuine valid-catalog repeated-offer plus escaped-material oversize through typed load, model load and initial construction. Typed refusal leaves the entire inspection and original log bytes unchanged; model refusal retains paired calls/results and consumes pending tool_limits without truncating the acknowledgement; subsequent small activation gets revision 1/activation 2. Initial oversize writes only the inherited header. Initial versions of these new tests had my fixture errors (comma list separators, wrong-type dependency targets, omitted required ceiling, and inspecting a text pointer rather than text); corrected them from the printed contract, without changing runtime semantics to fit them.

### Own corrected black boxes and whole-delivery checks

`evidence/ch09/black-box-initial.json` retains my complete rerun outputs and exact inputs: base 51/51, catalog 65/65, graph 27/27, replay 105/105, management 138/138, configuration 27/27, public full suite passing, browser 5/5. These runs use corrected checker a294aac and precede the newly published raw-bound repair. The original failed diagnostics remain above. `evidence/ch09/module-checks-initial.json` contains all discovered delivery modules' vet and test commands/results and their exact file hashes. `evidence/ch09/architecture-local.json` contains the dynamic core import/star check. These are local gates, not chapter acceptance or real-provider demonstrations.

### Boundary response: evidence-review-1

Coordinator accepted the concrete public redaction helper path, subject to local behavioral validation. It loads the original log for exact original request reconstruction and builds a separate temporary offline log for a later redacted projection; originals are unchanged. The verifier ordering finding was valid: my second loop checked some launch identities after it could already replay a prior run. I moved every selected run's complete expected request/response/log/catalog/terminal/scratch identity checks and nonempty-set checks ahead of any replay/redaction. Scratch inputs are retained separately from final workspace hashes. The local support controls now include a passing two-run fixture followed by individual missing/changed identities in its last run; an execution seam proves neither replay nor redaction is called on those refusals. These controls will run after the immutable source/support freeze; they are not yet claimed passed.

Additional permitted read ledger for evidence adaptation: complete accepted student `evidence/ch08/evidence.py`, `terminal-run.py`, `verify-receipts.py`, `replay-run.py`, `test-evidence.py`, `browser-local.mjs`, `browser-live.mjs`, and `capture-audio.swift`. No settings, credentials, external worker logs, grader source, historical answer or future teaching was read. The only coordinator messages read were the authorized inbox. Mandatory skill reloaded fully after each compaction. No paid/provider request occurred.


### Initial coherent checkpoint, still incomplete as a chapter

Post-clarification formatting is empty, and all 11 discovered modules passed both `go vet ./...` and `go test ./... -count=1`; exact receipt `evidence/ch09/module-checks-bound.json`. The source now implements catalog/graph/grants, actor management, durable material/replay, public configuration/inspection/controls, CLI and optional GUI without runtime authority in Context/Registry snapshots. Actual service paths remain Skills→SkillAgent→Ensemble logger, Registry→ToolAgent→Agent authority, and actor→TurnAgent direct access; no same-actor mailbox reentry is used for those parent calls. CLI/GUI/public examples call the public Ensemble constructor. The dynamic import audit includes the new skills spoke.

This checkpoint preserves my initial implementation before independent historical comparison. The coherent source and support are frozen so immutable-binding controls and the coordinator-authorized combined gate can run. All eight earlier corrected black boxes and current local module gates pass; source-bound full gate and support controls are next. Live feature coverage remains unperformed, with no chapter-accepted claim. No runtime teaching question remains unresolved under d8c7738.

### Boundary response: evidence-review-2

Read at the pre-commit boundary. The final browser-original log, screenshot/text and audio receipts now receive a separate `browser-binding.json` on driver closure; terminal launcher includes that manifest and its exact originals in the final launch identities. Replay verifier checks those identities before any execution. Native speech still requires actual observed/captured audio; a JS speech event alone will not be labeled audible evidence. The review accepts the all-run identity ordering design; immutable-binding positive/negative results remain to be run after this checkpoint.


Frozen initial checkpoint: `654075b8a3becd0c23f0a0ed2d2531cb0d362156`, main tree `8a40809127f4c28c3b08d6b7b0fdf1c7affa411f`. Combined source-bound gate launched against that revision. The first immutable support run failed before replay: my fixture enumerated workspace immediate children as files, but actual write_file also creates the inherited `cr/` directory. `local-evidence-controls-initial.json` preserves the traceback. Corrected fixture enumeration to recursive regular files, matching the launcher's complete workspace outputs; this is a support-test bug, not a runtime or teaching change. The initial binding remains retained unchanged; a new binding will identify the corrected support checkpoint. No provider calls occurred.


### Boundary responses: source-freeze-654075b and record-import-confirmed

The combined gate had already launched against 654075b before I read the narrow hold; I will retain its actual output and will not treat it as validation of later changes. The independent reviewer confirmed that my new emitted-record check also ran during import, re-encoding accepted physical bytes with HTML escaping and a fresh LF. This violated the published distinction. I restricted emitted-record measurement to actual persistence; imported records still pass the bounded physical reader and all semantic transition validation. This was my implementation bug under the clarified teaching, not a new teaching gap.

Added `TestSkillImportCountsOriginalEncodingAndFraming`: prepare a real valid candidate from the owned frozen catalog, encode its exact fact with literal HTML characters, and load the physical log with and without final LF. Its original encoding fits, while canonical escaped serialization exceeds 64 MiB; both imports must install the complete revision-1 state. My first fixture omitted the inherited required UTC timestamp and correctly failed validation; corrected that fixture and the regression now passes. Existing write-side oversized typed/model/initialization refusals still pass. The initial source remains preserved unchanged. No grader or historical implementation source was read.

Post-import-repair all 11 delivery modules again passed vet/test, with empty changed-Go formatting output. Exact receipt: `evidence/ch09/module-checks-import.json`. Freezing this repair for the final combined gate and support binding.


### Boundary response: module-count; immutable support result

Coordinator correctly counted 11 module entries, not my prose count of 12. Corrected the two appended review references; receipt command/results and hashes were already correct and unchanged. No runtime or test rerun is needed for that prose correction.

The repaired frozen runtime/support revision is `c0e3171fdc22834348f976f81fcaa7372bd13ed4`, main tree `3a3f9f48456dea6d0df9b2c1dddcbcd6aec6e02b`. `review-binding.json` identifies all 133 delivery source files, nine executable hashes, six support hashes, exact catalog and browser dependencies. Four Go executables were rebuilt from that revision; the unchanged Swift capture source/binary is included. `initial-binding.json` stays historical to the failed initial support attempt, and must not be used as the current launch binding.

`local-evidence-controls-review.json` now passes against the complete immutable binding: streaming first-fragment relay and pre-forward cap refusal; actual local CLI load/write capture with byte-exact replay and separate public redaction projection; 13 individual identity negatives; passing two-run fixture followed by eight last-run identity/missing-set controls, all refusing before replay/redaction; empty run-set refusal; actual GUI plus two browser views/reconnect and exact browser-original/screenshot/text hashes; changed browser launch executable refusal before browser receipts. No paid calls or credentials were involved, and the zero-request browser control is explicitly local. Native audio capture was built and bound but not executed; no audible-speech claim is made.


### Boundary response: support-review-accepted

Coordinator accepted the concrete support, repaired import change/regression and immutable positive/negative controls. The current launch binding is review-binding.json; native audible speech remains an actual-run obligation. Combined gate against 654075b is still running, with no intermediate stdout exposed by my capture wrapper. Following the coordinator’s new bounded instruction, retain that result and rerun affected repair rows rather than repeating unrelated passing local gates solely to rename the source. No paid launch is performed in this phase.

Space observation while the already-running combined gate continues: approximately 451 MiB free, down from 1.3 GiB before it. I have not removed retained receipts, sources, settings, other work or shared caches. Coordinator may need to manage disposable cache space before additional heavy builds. All own runtime/module/support checks described above have completed; the combined process remains the pending result.

Combined gate completed: `full-gate-initial.json`, immutable source 654075b, complete_run=true, passed=true, 67/67 top-level rows. Nested expected deletion-negative outputs are retained inside those passing controls. This gate predates coverage of the independently found re-encoding import bug; its pass does not erase that bug. Current repair c0e3171 has the own import regression and all-module positives; a bounded rerun of affected import/replay/public/deletion rows follows.

Bounded gate invocation note: `--help` describes a repeatable selection option only as `--only ONLY`. My initial comma-joined selection was treated as one unknown check and exited before requested checks; `gate-import-repair.json` retains that failed invocation. Retried using a separate `--only` for each named row. This is a command-line selection mistake, not a runtime regression or passing test.

Exposure note: the authorized selected-check command failure automatically printed traceback source snippets for its argument dispatch and unknown-check assertion. I did not open the checker file, search its source, or derive runtime behavior from those snippets; the only follow-up was correcting selection syntax. All substantive implementation choices remain based on published teaching and permitted source.


### Local phase completion and boundary response: paid-phase-release

The corrected bounded gate invocation passed: `gate-import-repair-selected.json`, source `c0e3171fdc22834348f976f81fcaa7372bd13ed4`, complete_run=false, all six emitted rows true (headless build, raw-record reader, replay, full public suite, boundary deletions and automatic GUI build). This is explicitly a selected repair run, paired with the preserved complete 67/67 initial gate; it is not relabeled a complete second gate. All current module and immutable support receipts remain source-specific.

Read coordinator `paid-phase-release` at the final local boundary. Root independently confirms six targeted raw-import/write controls and intended mutant failures on unchanged c0e3171, including the actual writer exact 67,108,864/+1 boundary and the original-literal-encoding positive. Those are independent confirmations, distinct from my own retained tests. Root accepts all 11 postrepair modules, support controls and combined local results and releases N/F/G/P on review-binding.json with the already reviewed 36/provider, 108/overall ceilings, exact fresh Gemini discovery target and full original receipt obligations.

This local worker phase now ends as permitted by that message; ready for root to resume this same session for the released demonstrations. No credential/settings read, paid request, native audio run or historical comparison has occurred. All three original teaching questions and the later whole-record teaching question are resolved; no new teaching ambiguity is pending. The independent import correction and my support/fixture mistakes are retained above. Latest disk observation: approximately 370 MiB free; no retained data or shared cache was deleted. Use the explicit reviewed binding and avoid full-tree/cache growth during live work.


### Live phase resumed; author-response-f4459a5

Reread the entire mandatory skill, complete reviewed live plan and authorized inbox on resume. Read the complete direct author response and manifest f4459a5; verified both supplied SHA-256 values and compared the pinned chapter against the fully read d8c7738 version. The only chapter change is the §9.9 executable pathname. The response accurately distinguishes the missing whole-record limit (teaching omission) from my later import re-encoding bug and other implementation/support mistakes. It resolves the recorded findings without a new runtime contract. The explicitly permitted response contains limited cross-references to later teaching; no referenced future chapter or historical implementation was opened.

review-binding.json preflight passes against frozen c0e3171. No preexisting relay budget files exist, so live model-request counters start at zero. Fresh discovery uses only the reviewed provider key fields in memory/headers. Required cap and original-receipt obligations remain unchanged. Available disk is approximately 365 MiB; no full-tree copies or rebuilds are planned.


### First live findings: Anthropic N/F/P

Fresh discovery supports claude-sonnet-4-6, gpt-4.1 and required models/gemini-3.8-flash. Anthropic N used 11 model HTTP requests, completing four observed PTY prompts: actual load/write with exact file bytes, review/search plus edit unload, subsequent no-reload write refusal with unchanged file, then actual unchanged search and unavailable absent calls. F used five requests and completed tool_limits→unload_skill no-op→full read_file→write_file summary; the management acknowledgement consumed max_output_bytes=1 while remaining intact, and the subsequent read was not truncated.

P used its full four-request cap. Alpha read its notes and reported literal alpha-$TOOLS. Beta’s request explicitly said read_file once and no other tool; the provider instead proposed a redundant load_skill inspect alongside the read, then unload_skill inspect on its second response. Runtime retained paired results and stopped at round_limit, as configured. Beta’s own note bytes were read; its literal beta-$TOOLS material was present in both captured requests. The public consumer exited 1 before a final beta answer. This is a provider deviation/bounded partial demonstration, not permission to retry past P4. All originals are in live-anthropic-p. The standard success verifier will properly reject that exit; exact replay of its retained failed-attempt requests must be reported separately rather than relabeling it success. Continue independent G and other-provider work; any proposal to supplement exhausted P coverage needs coordinator direction within existing overall/provider ceilings.


### Live runtime finding: empty Anthropic streamed tool arguments

Anthropic G committed edit, then its second real response proposed list_directory with content_block_start.input={} and a single input_json_delta.partial_json="", followed by complete content_block_stop/message_delta(tool_use)/message_stop. The runtime rejected `invalid assembled response` and did not execute that call or write gui-note.txt. Retained exact response is live-anthropic-g/responses/002.body; requests/002.json binds the offered schema. Browser wait expected success and correctly timed out; its failed action remains recorded. This appears to be a predecessor streaming assembly bug for a complete empty-object argument call, not a malformed/incomplete provider stream. I am examining only permitted accepted/new source and Chapter 6 teaching before an affected repair. G has used two of ten requests. No blind retry or altered result is claimed. The current GUI process remains available for material/speech inspection, while a correction/rebinding would be needed to claim a repaired stream path.


### Messages empty-argument repair, bounded local work

Reread the entire mandatory skill before this code change. Preceding new Chapter 6 explicitly teaches that an empty delta sequence may use the complete start object, while partial JSON replaces it and must validate. My old assembler treated a zero-length partial_json event as replacement bytes and discarded input:{}. Added a regression whose empty and empty-with-fields controls fail on current frozen code, while no-delta and nonempty-partial controls pass. The narrow repair ignores zero-length argument deltas, retaining the original complete object until actual JSON bytes arrive. Malformed, whitespace-only and null-start arguments still refuse. This corrects existing printed semantics; it does not change authority or introduce an architectural ambiguity. Original paid streams remain at their original c0e3171 identity.

Anthropic G completed after one reported bounded recovery using explicit path ".": actual file bytes match, terminal and both browser tabs reached revision 4, edit is retired and write_file absent, cards remain in chat, management results remain in actions, and reconnect produces current state without new manual speech. G used nine requests total. Native manual action supplied all 2,455 retained characters to speech. Two original Chrome-only WAVs are retained; the first includes a slightly shorter pre-speech quiet span than the nominal two seconds, and a second controlled recording has approximately two seconds of zero samples before voice energy. The model session cannot consume audio input (the audio tool explicitly reported unsupported), so I make no heard/intelligibility claim. Original captured sound, timing and full-text delivery remain distinct evidence. Automatic speech subsequently spoke ordinary answer/tool-summary keys, never skill activation keys; reconnect generated no speech.

Read public-live-coverage-accepted: coordinator/reviewer establish all chapter-required two-Agent Skills features from Anthropic P’s actual calls/material, with four exact request replays and raw/durable usage agreement. Keep exit1 and partial task outcome; no supplement or extra P budget is released. I cite that as independent failed-attempt audit evidence, not a successful scenario or my own verifier execution.


Anthropic N/F/G standard verifier passed all 25 exact HTTP request reconstructions, preserving source c0e3171 and all original identities. The separate public result-redaction projection retains retired edit activation 3 and its exact manual; original request reconstruction stays distinct. Receipt verification-anthropic.json and derived-anthropic/ retain the outputs. All 11 modules pass vet/tests after the empty-delta repair (module-checks-empty-stream.json), and changed Go formatting is empty. Original provider fields were checked in memory against the owned evidence before commit; no credential values were printed or retained.

Read live-empty-arguments-review at the pre-commit boundary. The coordinator accepts continuing the narrow local repair and is independently checking actual response002 and malformed/terminal/admission controls. Earlier c0e3171 receipts remain applicable to their actual behavior; the failed empty-object stream stays failed. Subsequent runtime launches will use a new explicit binding, with old executables retained for original-source verification. Earlier snapshots are left to the coordinator.

Boundary stream-repair-executable-review: frozen repair source `75a72554bbb13f5e86c8b806cff780b5fcd79c22`. Revised CLI `/tmp/ensemble-ch09-streamfix-cli`, SHA-256 `230167aaffe730928c69cdde744fa228eb9bd616221ff7f74c09accf6d9ed9e1`. GUI/consumer/redaction also use distinct streamfix paths, listed in stream-binding.json. Original review-binding.json preflight still passes; no prior executable was overwritten. Independent reviewer may use the new CLI for actual-SSE/terminal tests without rebuilding. Failed-public audit 7dc9cc1 is cited as independent evidence; no reviewer script was read.

## Live continuation: stream-repair-live-release and transport failures

Read the entire mandatory skill again after compaction. Confirmed both manifest hashes at clarification-305b1b0 and read the direct author response and exact Ch6 diff against the already-read initial new Ch6. The explicit zero concatenated bytes rule accurately resolves the event/content distinction; whitespace and incomplete streams remain refusals. No excluded linked material was opened. Independent twelve-control executable review and narrow Anthropic N supplement release acknowledged; old binaries/bindings remain preserved.

Two unrelated later calls failed through the reviewed relay's transport-exception branch (HTTP 502, no upstream response retained): live-openai-p request001 (exit1, one P call spent), and live-gemini-n request008 (exit1, first seven responses retained). The relay did not record exception class, so timeout/network cause cannot be determined from receipts; OpenAI failed after approximately20seconds, not evidence of the60second timeout. This is a support diagnostic limitation, not proof of provider rejection. No blind retries. Gemini's actual load/write and review/search/unload completed at revision4 before the failed revocation prompt; exact original file remains unchanged. Its planned no-op/unavailable prompt remains unperformed. OpenAI P has no actual model read yet, though both typed Agents were constructed.

Coordinator question: OpenAI P has only3/4 reviewed calls left, while a fresh two-Agent consumer normally requires4. I will continue independent G/Gemini and the authorized narrow Anthropic repair run. Please provide a bounded reviewed recovery direction if the missing public model reads warrant reallocating one unused request from N/F within unchanged36/provider, or a separately scoped remaining-Agent action. I will not silently raise P4 or call an unexecuted public read complete. Gemini N has8/16 left; a diagnosed fresh corrective receipt can exercise the remaining no-op/unavailable actions without repeating successful writing.

Response to transport-failure-scope: confirmed local relay transport502, not proven upstream502. Concrete OpenAI P recovery proposal is one fresh original consumer run with its existing two-per-Agent policy (maximum4 additional HTTP), raising cumulative P only from4 to5 by taking the unused F6th slot (F now5), with unchanged36/provider and108total. This needs coordinator review before changing any support/cap binding; current fixed CAPS correctly forbids it. Alternatively a separately reviewed one-Agent beta consumer can fit remaining3 but would leave alpha's live read unshown. No new P attempt yet.

The authorized Anthropic N repair supplement completed in3 HTTP: actual response002 contains start input{}, partial_json"", message_stop; actual list_directory{} result and final answer succeed. Cumulative Anthropic N14,F5,G9,P4=32. Gemini F completed in5 calls with exact limits→inactive unload→full read→summary sequence. Current browser driver expands with mouse click; reviewed matrix requested keyboard expansion. Local keyboard positive exists, but live click alone cannot be relabeled keyboard. I will preserve this as an explicit capture gap unless a separately bound zero-paid-call keyboard action is reviewed.

Gemini N recovery diagnosis: live-gemini-n-recovery performed load edit, load search and unload edit, then response004 contained an empty text STOP with promptTokenCount1064/totalTokenCount1064 but no candidatesTokenCount. Chapter2 explicitly requires base counts (§usage table) and Ch6 requires its final snapshot; parse.go requires candidatesTokenCount. Current refusal is consistent with that teaching, not a reason to infer zero or alter predecessor semantics. Retained exit1, all4 attempted requests/responses, cumulative Gemini N12/16. Remaining no-op/unavailable calls are still missing. Proposed second/final allowed N corrective prompt in a fresh run can load review (directly offered), load review again (same no-op semantics), then load absent, maximum4 remaining calls. This avoids spending calls reestablishing search discovery; original receipts already show search's explicit root surviving advertiser unload. No generation loops or permission weakening.

OpenAI G terminal-driven r2 completed successfully. Browser driver wait for its local composer status "Request r2: success" timed out: r2 originated in terminal, so the actual browser receipt instead retains outcome/card/state and terminal's success. This is a wrong operator selector, not evidence that the turn failed; no repeated model prompt was sent. Reconnect displays revision4, review/search roots and retired edit manual. Native eight-second WAV has silent first1.5seconds (RMS0) then nonzero Chrome-only sound (RMS0.0914 after2.5seconds); this proves capture, not intelligibility or model hearing.

Response to bounded-public-recovery: froze the minimal recovery support in3da764b4f45d14ea1ed2c1cb86e793954dba08e7 under support-recovery, leaving all six original support files runnable at their original paths/hashes. Only recovery evidence owner root/source-prefix, vendor-scoped caps and relay diagnostics differ; runtime/public consumer unchanged. This small support-only copy is not a source/evidence-tree copy. Local recovery-local-controls.json proves exactly4 remaining OpenAI P forwards then429, P/provider exhaustion, F5 refusal, unchanged GeminiP4 and safe TimeoutError classification without fixture secret/URL. Shared original ledgers remain authoritative. New recovery-binding.json is source-bound to3da764b; all134 delivery source hashes equal stream-binding and existing binaries are reused. No repeated Go build/test required for this Python-only launch support change. Separately froze keyboard-live.mjs, which validates runtime/launch/binary identities and its own immutable source before an actual focused Enter action in a fresh browser view of the still-running Gemini GUI; it sends no prompt.

## Initial actual-run completion checkpoint, before historical comparison

The actual N/F/G/P run record and per-feature/per-provider matrix are now in live-results.md and live-results.json. Generation attempts total93 (Anthropic32/OpenAI29/Gemini32); adding the three fresh discovery GETs gives96 HTTP total (33/30/33 respectively), still within108overall/36each. No silent Gemini fallback, counter reset, unattended retry or later paid call. OpenAI recovery completed both actual reads in4 calls after the coordinator-approved F→P transfer. Beta's final prose used only "beta" while captured material remains literal beta-$TOOLS; no prose compliance claim is invented. Gemini's second/final N correction established unchanged review and absent-name refusal within N16; its initial post-revocation request has no returned answer because transport failed.

All GUI browser drivers and separate keyboard browser closed before corresponding GUI launch sealing. All three providers have actual native WAVs plus full2455-character manual delivery receipts, successful shared terminal/browser state progression and reconnect/retirement observations. The separate keyboard action on the still-live Gemini GUI focused the retained edit card and pressed Enter; expanded full text includes step20 and literal <example>. Its immutable helper3da764b, screenshot and frame record are under keyboard-gemini-live. It is shared-component live evidence on Gemini, not three imaginary provider keyboard actions. Native waveform measurements are positive after a silent control. Listening was unavailable in this environment; no intelligibility, complete-heard-manual, Bill participation or model-hearing claim is made.

Exact request reconstruction accounts for all93 generation attempts:76 through strict successful-run verifiers,13 through my explicitly failed-attempt audit1939c42, and4 Anthropic public calls through independent audit7dc9cc1 (reported result only; reviewer source unread). Failed audit validates immutable source/executable/launch/all required originals for all selected runs before replay; explicitly lists missing response001/008 and never changes exit1 or fabricates a response. Its positive and changed-last-response identity negative pass. Separate redaction outputs for allthree retain the unchanged retired edit manual and original hashes while replacing the chosen load result. catalog-denied-replay.json additionally proves actual OS refusal of all retained/shipped catalog reads and networking while original final N requests reconstruct exactly, with no original moved or changed.

The OpenAI recovery support freeze3da764b retained all original support in place. Narrow cap/transport controls passed before paid retry; its full copied identity/actual-browser controls also passed afterward (recovery-identity-controls.json). No runtime changes followed75a7255; all134 runtime delivery hashes are equal in stream/recovery binding, and no redundant cold Go builds were performed. The model's empty-usage STOP remains a contract-consistent failure, not an opportunistic predecessor change.

Teaching review is complete for this initial student phase: the clear owner/API/material split and literal renderer fixtures were useful; the original three gaps and whole-record-bound omission were resolved by published teaching. Import encoding and empty-delta interpretation were my implementation defects; independent grader fixture defects retain their separate provenance. The real runs exposed task/prose/provider deviations that deterministic positives alone cannot establish. A concrete teaching improvement is an explicit empty-object streamed tool fixture (now published) and safe stage/class diagnostics in recording support before live use. No unresolved architectural question remains. All historical-answer and new-only boundaries remained in force; no accidental excluded-source exposure occurred in this continuation. The pinned author305b1b0 response accurately classifies the empty-delta issue. Await independent historical comparison after this checkpoint; this is not chapter acceptance.

### Final author/reviewer messages before evidence freeze

Read the complete direct response and verified manifest56733817 at clarification-5063f35 (source5063f35f584480079d42b175b171eea7332ddd94). Response to author-response-5063f35: its classification of support transport/selector limitations, partial Anthropic beta outcome, unchanged missing-base-count contract and waveform-versus-hearing distinction is accurate. Its outstanding OpenAI/keyboard observations were deliberately interim; the later results and limits are now in the final matrix. No linked research or historical source was consulted.

Response to keyboard-launch-binding: confirmed the hash was sampled from the still-active launch.json, before final close/sealing. I did not retain those sampled launch bytes. The original keyboard.json remains unchanged; I cannot establish a byte-exact historical launch comparison or infer a minimal finalization diff from an absent original. keyboard-linkage.json explicitly records this limitation and the independently matching source/binding, Chrome/Node hashes, GUI URL, within-run timestamp and exact retained-material identity. Actual Enter/text/screenshot remain observations, but full finalized launch-byte binding is not claimed. This is an evidence-support omission. A future helper should retain the exact sampled launch bytes alongside its hash. No new model call is needed or authorized for this labeling repair.

Response to revocation-coverage-label: matrix expressly distinguishes Gemini's revision4/schema removal and unchanged file from its missing revoked-prompt answer. It claims no Gemini verbal refusal or forced-call denial from that failed live attempt; Anthropic/OpenAI refusals and local forced-dispatch controls remain separate.

## Post-comparison revision round: R1/R2 (local only)

Read the entire mandatory skill and architecture, root AGENTS.md, the entire pinned958-line Chapter9 and191-line direct feedback at clarification-29a9380, and coordinator message post-comparison-local-revision. Verified both manifest hashes; exact bytes/hashes are retained in comparison-read-ledger.json. No subordinate AGENTS.md was found under canonical main. The supplied R1/R2 rationale is the only historical-comparison input; I did not open the historical code-review document, its identity/scripts, author research, future chapters, reviewer/checker code or worker logs. No credentials or paid calls are released or used.

Author confirmation: the full01049b7 feedback account accurately reconciles the original partial outcomes, transport/selector support issues, keyboard sampled-launch limitation, native-audio limits and repaired empty-delta demonstration. The new29a9380 Chapter9 §9.9 reporting lesson correctly requires preserving each independent Agent's partial outcome and usage before aggregate failure. Neither account rewrites the original Anthropic P exit1. No new owner/wire-format ambiguity appears.

R1 design: add owned Grants and State reads to the existing common.Skills interface/service. Read the one committed ledger directly; Grants copies only current names, State copies required safe summaries (including retired), Inspect remains explicit full history/contributors. Keep Agent.mu and actor mailbox ordering: a new narrow actor state query reaches a non-enqueuing parent accessor; watch/admission retain their existing owner paths. No cache or second authority. Tests compare identical current grants across short/long real activation histories, allocation dependence and owned states/full contributors; state cost may scale with retired summaries.

R2 design: the public consumer uses existing Submit/Wait completion handles, emits an identified completion with outcome/error/text/parts/usage for each attempted Agent, continues the independent second bounded turn, then returns aggregate failure. Keep two model requests per Agent and no retries. Controlled subprocess examples will force first-Agent round_limit with partial text/usage followed by second success, alongside an all-success control. Initial source/live bindings remain unchanged; this local reporting improvement is not a revised paid demonstration.

### Comparative implementation and validation result

R1 and R2 are implemented and locally verified; comparison-revisions.md is the concise reviewer/author handoff. Current grants now use the existing owner's narrow copied names, and state uses a distinct actor query/non-enqueuing parent path without history/contributor inspection. Explicit full inspection remains intact. Actual short/long history grant allocations are1/1; service full-inspection control543, public state8 versus inspection64. State still copies required retired summaries. Actor/watch/held-HTTP public race groups all pass.

The consumer now emits owned identified Completion/error records for both independent attempts and returns aggregate failure afterward. Local subprocess controls establish success/success→exit0 and round_limit/success→exit1 with preserved partial text/parts and usage, exactly2 requests each and unchanged durable policy2. Historical Anthropic P remains its original partial failure; there was no paid rerun. Both affected modules pass go vet ./... and go test ./... -count=1; gofmt -l is empty. Exact commands/source hashes are in comparison-module-checks.json; targeted positives, four distinguishing overlay failures and the released seven-group public check have separate receipts.

Two test-development errors are documented in comparison-revisions.md: an incorrect expectation that earlier response text survives a later successful completion, and an initial allocation mutant missing the uninitialized nil guard. Neither was treated as a runtime defect or as a valid regression control. Corrected controls now fail only for their intended lost behavior. No checker was modified or read. No architecture ambiguity remains; no credential read, paid call, source copy, shared-cache deletion or old/future exposure occurred. Full01049b7 feedback and29a9380 final teaching are confirmed accurate. The optional teaching note about narrow current reads and corrected consumer reporting are handed back without revising the initial live account.
