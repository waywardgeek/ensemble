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

Post-clarification formatting is empty, and all 12 discovered modules passed both `go vet ./...` and `go test ./... -count=1`; exact receipt `evidence/ch09/module-checks-bound.json`. The source now implements catalog/graph/grants, actor management, durable material/replay, public configuration/inspection/controls, CLI and optional GUI without runtime authority in Context/Registry snapshots. Actual service paths remain Skills→SkillAgent→Ensemble logger, Registry→ToolAgent→Agent authority, and actor→TurnAgent direct access; no same-actor mailbox reentry is used for those parent calls. CLI/GUI/public examples call the public Ensemble constructor. The dynamic import audit includes the new skills spoke.

This checkpoint preserves my initial implementation before independent historical comparison. The coherent source and support are frozen so immutable-binding controls and the coordinator-authorized combined gate can run. All eight earlier corrected black boxes and current local module gates pass; source-bound full gate and support controls are next. Live feature coverage remains unperformed, with no chapter-accepted claim. No runtime teaching question remains unresolved under d8c7738.

### Boundary response: evidence-review-2

Read at the pre-commit boundary. The final browser-original log, screenshot/text and audio receipts now receive a separate `browser-binding.json` on driver closure; terminal launcher includes that manifest and its exact originals in the final launch identities. Replay verifier checks those identities before any execution. Native speech still requires actual observed/captured audio; a JS speech event alone will not be labeled audible evidence. The review accepts the all-run identity ordering design; immutable-binding positive/negative results remain to be run after this checkpoint.
