# Global review: lessons to teach before they become repairs

Review date: 2026-10-07. This is a forward-lesson and dependency map for the
second edition, not an implementation specification or validation report.
Chapter numbers below refer to the **first edition** unless stated otherwise.
The current proposed second-edition plan is **22 chapters, numbered 0–21**:
the prior 0–20 sequence plus a newly requested final comparison chapter.
Bill requested an enhanced Chapter 0 execution guide and permits
documented organizational changes with a concrete teaching benefit. Preserve
the first-edition Chapter 0 and preface; write the new guide in this edition.
Old Chapters 1–4 keep their numbers, and old
Chapters 6–21 become new Chapters 5–20. Old Chapters 5 and 22 are absorbed
at the relevant first introductions, not retained as standalone chapters.
Any dependency-driven adjustment must document its teaching benefit.
Preserve historical editions as the series *The Singularity as It Happened*.
The final comparison is new work, not a revival of either architecture-repair
chapter. Bill also authorizes a complete first-person epilogue rewrite when
the second edition is publication-ready; no ending or outcome is invented now.

## Authority and read ledger

**Current user rulings** govern the rewrite. **Coordinator choices** are
working design decisions made within that authority. **Review recommendations**
below identify lessons to incorporate into chapter contracts; they do not
silently add requirements to a student exercise.

The reviewer read the complete canonical [Chapters 1–22](../chapter-01.md),
in manageable chunks, including gaps from truncated tool output. Also read:
[course policy](../course-policy.md),
[writing procedure](../chapter-writing-procedure.md), [voice](../voice.md),
[current agent skill](../../agent/skills/ensemble/SKILL.md),
[common-refactor design](../common-refactor-design.md),
[topology decisions](../../docs/ensemble-topology-decisions.md),
[parent-chain repair plan](../../docs/parent-chain-repair-plan.md),
[agent-status wiring analysis](../../docs/agent-status-wiring-analysis.md),
Chapter 1 grader audit, Chapter 22 coder review, and the earlier global book
review. Selected Astra review passages were read; that file was not read in
full. This ledger describes reading performed, not permanent retention of
the entire textbook in context. Reload the relevant passages before resolving
a detailed question, especially after compaction.

For the latest consistency review, the reviewer read the second-edition
[Chapter 1](chapter-01.md), [architecture](architecture.md),
[workflow](workflow.md), [coding skill](skills/ensemble-coding/SKILL.md), and
[root instructions](../../AGENTS.md). This map incorporates the coordinator's
subsequent relay of Bill's authorization to continue autonomously. Other
status files may still be catching up with that relay.

No implementation, grader run, live vendor call, or OAuth research was
performed for this review. Historical measurements remain historical.

## Current authority to carry into every chapter

**User rulings:**

- Continue a fresh, autonomous second-edition rewrite. Leave the existing
  `agent/` and first-edition implementations untouched. New student code lives
  in `solutions/edition-2/chNN/`, built from scratch or the preceding new
  chapter; old solutions are evidence for authors, not student starting code.
- Add a new final comparison chapter to the prior 0–20 sequence, provisionally
  Chapter 21. Preserve historical editions and their evidence. Bill permits
  justified organizational changes. Neither
  former repair chapter survives as an extra chapter or capstone. Incorporate
  useful extension and audit exercises into the relevant remaining chapters.
- Rewrite the second-edition epilogue in the coordinator's first person when
  the edition is publication-ready, as Bill explicitly requested. This is a
  future writing task; actual results must determine its ending.
- The initial rewrite began with Chapter 1. Bill now requests an enhanced
  second-edition Chapter 0 execution guide; the preface remains untouched.
  Teach the architecture in plain English before any Chapter 1 code. There is no initial
  flat-package exemption from the methodology.
- Coders load the entire standing coding skill before every coding task,
  including fixes and graders, and after compaction. Summaries do not replace
  that read. Architectural ambiguity affecting a task is surfaced before
  implementing the disputed choice; independent work continues.
- One application owner manages many Agents and communication with GUI or
  gateway clients. Agent holds an interface back-pointer to that owner.
  Observer carries streaming events and live display updates; explicit parent
  methods may request actions or expose services. Observer is not exclusive.
- Starting in Chapter 2, clean core data structures support CLI and browser
  GUI clients of the same owner. The GUI may be stubbed initially. GUI code,
  including WebSocket, is a separate optional Go **module**, never
  `agent/internal`. Reusable GUI components require public seams.
- Code likely to need debug logging must reach the logger, including parsers
  and stateless helpers. Tool visibility is per-agent; registry storage may
  be Agent-owned or shared on the application owner with per-agent visibility.
- Every chapter's live demonstration must actually exercise all its features
  through a real user interface, initially the CLI, with an actual model
  backend. Fake grading complements this evidence. Demonstrate the three
  vendor APIs as introduced. A stub is labeled a stub, never a tested feature.
- Bill authorized credentials from `~/.cr/settings.json` for those runs.
  Never print, log, copy into the repository, or commit secrets. The skill
  specifies safe loading and sanitized evidence.

**Coordinator choices, relayed with this assignment:** use `Ensemble` as the
working application-root name and put the logger on that root, matching Bill's
preference. Exact interface method names are routine design choices once
ownership and behavior are settled. These choices do not settle registry
storage or authorize changes to the old agent. Do not ask Bill to choose
method spelling merely because a signature is not yet printed in a chapter.

**Research queued by the user:** at the caching chapter, investigate current
official OpenAI subscription OAuth access and caching behavior. Bill reports
a recent unresolved caching bug; that is a user report to investigate, not an
established defect or verified provider claim. See the research gate below.

## Methodology inherited from the later repairs

The authoritative current explanation is the
[coding skill](skills/ensemble-coding/SKILL.md). The older examples need
reconciliation, not literal reuse.

1. A child receives an interface to its actual creator/owner. The interface
   exposes its own parent so code can walk upward. No redundant ancestor
   shortcut, dependency bag, or closure stapled on after construction to
   recover a missing route. `Call.Engine.Agent().Ensemble()` illustrates the
   route as those objects arrive; it does not require unused early objects.
2. Core shared data and interfaces live in `internal/common`. Implementation
   spokes do not import siblings; common imports no implementation. The
   composition root constructs the ownership tree through public entry points.
   An acyclic Go import can still violate the stricter star rule.
3. Operations on common types live as free functions in the responsible
   spoke. Methods needed for standard-library interface dispatch stay with
   their types. Behavior needed by several spokes gets a common interface and
   an implementation in an owning spoke. A vocabulary package must not become
   a behavior warehouse merely to retain method syntax.
4. Facts stay with the object that knows them: configuration generally on
   Agent, usage on Engine with the producing model. Reachability is not a
   reason to relocate a fact to a convenient ancestor. Mutable state belongs
   to instances, not package globals. A singleton root is an explicitly
   constructed object, not a global variable.
5. Structural checks must verify actual imports and parent/service paths for
   every present spoke and executable. Matching names such as `Host` or
   `Logf` proves neither ownership nor reachability. Pair those checks with
   independent behavioral evidence and deletion-sensitive grading.

Sources: [Chapter 5 §§5.1, 5.9–5.10](../chapter-05.md),
[common-refactor design](../common-refactor-design.md),
[current agent skill](../../agent/skills/ensemble/SKILL.md),
[Chapter 22](../chapter-22.md), and [course policy P9/P11](../course-policy.md).

Do not carry forward Chapter 5's sibling injection described as parenting,
flat `cmd` composition, global registration example, direct `Call.Host`
shortcut, or fixed Agent logger/registry placement. Chapter 22's printed
`Call` with separate Agent, Engine, and Jobs fields also conflicts with the
single parent route. Neither example overrides the current rules.

## Forward-lesson map

Each row is a **review recommendation** for contract design. Dependencies are
semantic: introduce a safeguard with the first feature it protects, not only
at the chapter number where the first edition discovered its absence.

| Original chapter | Lesson and earliest useful placement | Evidence to require when implemented |
| --- | --- | --- |
| [1: conversation](../chapter-01.md) | Teach architecture before code; keep the first behavior small. Define failure/history semantics instead of appending a user then continuing with invalid alternating history after failure. Explain cumulative usage without treating all growing input as equally billed. | Real CLI conversation, retained turns, measured totals; fake protocol/error cases; ownership/import checks. |
| [2: log and vendors](../chapter-02.md) | Separate immutable event history, reduced context, and vendor request projection. Capture exact provenance when written; normalize usage without losing producing-model identity. Add both client seams now, with optional GUI stub. | Three API paths as introduced, deterministic rendering, parser-to-event-to-reducer path, independent logs, headless build, public GUI seam. |
| [3: tools](../chapter-03.md) | Per-agent capability visibility must constrain declarations and dispatch. Preserve call identity and tool-result order. A tool failure is content; infrastructure failure is different. Shell working directory is not a filesystem sandbox. | Unknown/disabled tool rejection, all issued calls answered, truthful working-directory behavior, real tool use. |
| [4: jobs](../chapter-04.md) | Give each job a lifecycle owner. Wait wakeup is not deadline or cancellation. Process-output completion and dispatch completion are distinct. Full output can live on disk while reports are capped and cursored. | Long job start/wait/input/kill through the user path, late output, cursor advancement, independent per-call cwd and one-shot limits. |
| [5: architecture repair](../chapter-05.md) | Absorb its architecture into Chapter 1 and its extension/audit exercises where the relevant features first appear; no standalone replacement chapter. Logging and multiple-agent isolation expose missing ownership before scale does. | A separate consumer extends the public library; actual parent paths and all import edges; no mutable globals. |
| [6: actor](../chapter-06.md) | One owner mutates conversation state. Public synchronous calls enqueue requests and wait for their own completion; observations are not completion joins. Introduce lifecycle rules with the actor. | Concurrent requests receive their own replies; interrupts end turns, not the actor; shutdown joins workers; late results have a specified disposition. |
| [7: streaming](../chapter-07.md) | Incremental parser IDs are response-local; UI identity needs an explicit broader scope. Deltas are provisional; finalized events are durable. Nonstreaming and streaming parsing share semantics. | Interleaved fragments, finalization once, malformed chunks, two consecutive replies reusing provider-local IDs. |
| [8: GUI](../chapter-08.md) | The optional GUI module consumes public events and requests. Introduce safe WebSocket sender/teardown ownership immediately. Give final state a reliable delivery or resynchronization path despite lossy deltas. | Headless and GUI consumers, reconnect/resync, slow clients, disconnect during broadcasts, independent authoritative state. |
| [9: settings/UI](../chapter-09.md) | Validate input at entry and persisted-state load. Specify snapshot versus patch semantics so false/zero are meaningful. Settings must affect the running actor and survive restart where promised. | User changes a setting and observes backend behavior; invalid values; restart; pause causes individually and together. |
| [10: skills](../chapter-10.md) | Distinguish enduring instructions from dynamically loaded skill material from the first implementation. Give entries a kind that later retention rules can recognize. Keep visibility per-agent. | Real load/unload behavior and dispatch restrictions, two agents with distinct tools, required state visible in the user interface. |
| [11: persistence](../chapter-11.md) | Per-agent identity/DataDir is separate from vendor request config. Snapshot plus tail applies each event once after the boundary. Corrupt state must not be silently replaced. | Restart equivalence, exact tail boundary, corruption handling, independent agent directories, no restored credential leakage. |
| [12: voice](../chapter-12.md) | Voice is another client input/output path over the same agent ownership. Define cancellation and pause causes with the first speaking/typing interaction. | Real interface checks of the introduced voice features; cancellation, typing during output, errors returning UI to usable state. |
| [13: integrations](../chapter-13.md) | Integrations need owned lifecycles, not closures in a binary that substitute for parents. Bring UI safety, elision metadata, and useful truncation into the first component that displays external content. | Actual connection/tool use and shutdown; displayed untrusted content remains data; visible truncated output can be retrieved. |
| [14: GUI testing](../chapter-14.md) | Semantic controls and independently seeded state make GUI tests informative. Test through user paths when the UI first appears. A harness must clean up agents and clients it starts. | Accessible actions with real observable effects, false/zero settings, independent fixture state, teardown without leaked agents. |
| [15: redaction](../chapter-15.md) | Record concrete selection decisions; replay does not rerun policy. Finish a tool-call batch before cutting it. Preserve the previous batch as specified; narration and result evidence have different roles. | Recorded redaction replay, no orphan call/result pairs, exact survivors, retrievable retained output. |
| [16: compaction](../chapter-16.md) | Record the model's generated bytes so replay performs zero model calls. Live checkpoint state belongs to the creator; a compressor is stateless. Specify atomicity/recovery of the replacement transaction. | Zero-call replay, interrupted write recovery, actual current-window pressure rather than lifetime token totals, preserved live state. |
| [17: recall](../chapter-17.md) | Recalled material needs its own attributed entry kind and recorded text. Apply quotas mechanically after selection. A stateless judge owns its response parser, not conversation mutation. | Reproducible recall with zero replay calls, quota boundary cases, expiry/retention contract, one actor entry path. |
| [18: caching](../chapter-18.md) | Separate provider observations, local estimates, and billed cost. Instrument the actual serialized request. Apply context projection before cache-marker decisions. Keep per-model usage from Chapter 2. | Prefix byte comparisons, provider usage receipts, cold/warm trials with fixed conditions, unknown-price handling, queued OAuth research. |
| [19: credentials](../chapter-19.md) | Credential ownership stays outside renderers; secrets are request-local, not durable context. Route and funding transitions must be explicit. API and subscription access are separate evidence questions. | Authorized real route, sanitized receipts, coherent endpoint/model/surface/credential transitions, no silent billing fallback. |
| [20: live integration](../chapter-20.md) | Test effective state through actual entry points, not only intended config. Move discovered ID-scope, parser-finalization, content-type, and shutdown fixes to their first consumers. | Real backend user sessions and route changes, two-turn identity, correct terminal-event handling, normal shutdown on signals. |
| [21: web/skills integration](../chapter-21.md) | Retrieved content is data, not authority. Tool allowlists govern dispatch, not only declarations. Test the shipped default skill, not just a special grader fixture. | Live default feature use, disabled/guessed tool rejection, injected retrieval content cannot itself grant tools, configured integrations actually connect. |
| [22: final repair](../chapter-22.md) | Absorb parent reachability, per-model cost, and WebSocket lifecycle lessons and audits into their first consumers. No standalone repair or extra capstone chapter remains. | Deliberate parent-path breakage fails; mixed-model usage not repriced; sender/disconnect race exposed by sensitive fixtures; each binary uses the library root. |

The table does not require every hypothetical mechanism in Chapter 2. A
conversation shape must admit later features without rebuilding its meaning;
an unused implementation is not proof that the shape is correct.

## Chapter 2: recommended clean data choices

This is an author proposal, not a prescribed set of Go signatures. Its source
is [Chapter 2 §§2.3–2.8](../chapter-02.md), corrected by the ownership skill
and the later chapters named below.

1. **Keep history, context, and request distinct.** History is ordered,
   append-only events; context is a reducer projection; an HTTP request is a
   vendor-specific projection. A parser returns events and cannot independently
   mutate context. Put the reducer and model operations in `llm` as free
   functions over common data. The old pointer-receiver prescription is
   superseded; pointer mutation and free-function placement are compatible.
2. **Give recorded events stable identity and validated payloads.** Use
   monotonic sequence order, with timestamps as metadata. Validate event-kind
   and payload agreement; Go does not make a pointer-field envelope exhaustive
   automatically. Recorded slices/raw bytes must not alias buffers later
   reused by a parser. Define log versioning and unknown-kind behavior.
3. **Use attributed entries containing ordered typed parts.** Preserve text,
   calls, results, opaque replay material, and external references without
   reducing them to vendor role/content strings. A response retains the order
   of text and calls. Dispatch events do not duplicate assistant dialogue.
   Render supplied tool histories before implementing tool execution.
4. **Distinguish who spoke from what an entry is for.** Human/Agent/Tool/System
   attribution does not encode skill material, recalled memory, or a durable
   instruction. Specify an extensible entry-kind contract early so Chapters
   10, 15, and 17 can retain or remove the right material. Do not add invented
   payloads for mechanisms with no consumer yet.
5. **Capture provenance at creation.** Vendor, exact model, and API surface
   belong with responses and opaque material. Model names are open strings;
   vendor/surface support is explicitly validated. Keep issued call IDs and
   define deterministic synthesized IDs where the actual route requires them.
   The identity scope must eventually distinguish response-local part IDs,
   actor IDs, and durable replay IDs; one counter does not imply all three.
6. **Normalize usage at the vendor seam and retain provenance.** Store
   disjoint input/cache-write/cache-read/output observations with the producing
   model and surface. Engine owns accounting; common contains its data types.
   Derive aggregate views and later cost from per-model observations. Do not
   let switching the active model reprice old tokens. Historical formulas and
   rates must be reverified on the API surface actually exercised; cache
   storage duration is not represented by token totals alone.
7. **Keep current request configuration separate from recorded facts.** Agent
   owns current config; a request records the route/provenance used. Credentials
   never enter history. A deterministic render claim names the log, config,
   and renderer version/input it holds fixed; the log alone cannot recreate
   unspecified mutable settings or content behind a changed external file.
8. **Specify ownership of change and of observation.** One authoritative
   mutation path is useful before the full actor scheduler arrives. CLI and
   GUI stub call the same application services. Public transport-neutral
   observations and request methods do not expose private implementation
   types. Do not use a broadcast observation as a synchronous request's
   completion signal in the later actor design.
9. **Keep bounded projections distinct from archival history.** Avoid an
   ever-growing auxiliary redaction set or duplicate retained dialogue just
   to implement replay. Teach which fields are current state and which are
   historical facts. Later redaction/compaction choices are recorded events,
   not decisions rerun by the renderer.

Recommended Chapter 2 proof set: independently authored log fixtures; all
three vendor render/parse paths; response provenance and disjoint usage
arithmetic; repeatable rendering under fixed inputs; equivalent context from
incremental application and replay; malformed/unknown event rejection; core
headless consumption; GUI stub through a public interface in a separate
module; live CLI exercises of every implemented feature and introduced vendor.
No GUI behavior should be reported as exercised merely because the stub builds.

## Contradictions to resolve in chapter contracts

These are review findings about the first edition, not unresolved rulings
that should halt unrelated work.

- **Actor semantics:** Chapter 6's newer synchronous request section keeps an
  actor alive across turn interruption and uses request-scoped completions;
  older pseudocode exits on interruption or drops late results. Its exercise
  still uses observation-based Post/Wait patterns. Choose and teach one
  lifecycle. [Source](../chapter-06.md)
- **Tool ordering:** Chapters 3 and 8 describe serial tool handling, Chapter 6
  starts goroutines, and Chapter 15 describes parallelism. Distinguish dispatch
  order, asynchronous job execution, and report order explicitly. Also resolve
  Chapter 4's dispatch finishing a job whose process reader owns completion.
  [Chapter 4](../chapter-04.md), [Chapter 15](../chapter-15.md)
- **Prompt layers:** Chapter 10/16 instruction-versus-memory boundaries
  conflict with later passages putting recent memories in the system prompt.
  Specify each material's location, lifetime, and cache behavior once.
  [Chapter 10](../chapter-10.md), [Chapter 16](../chapter-16.md),
  [Chapter 17](../chapter-17.md), [Chapter 18](../chapter-18.md)
- **Compaction trigger and transaction:** Cumulative input plus cache-write
  is not current context-window occupancy and excludes cache-read. Define the
  actual pressure measure and crash behavior between compaction events before
  grading it. [Chapter 16](../chapter-16.md)
- **User-visible integrations:** A parsed MCP configuration that never
  connects is not a completed integration. Custom grader skills can conceal
  omissions from the shipped default. [Chapter 21](../chapter-21.md)
- **Claims exceeding evidence:** Acyclic sibling imports are legal Go; an
  append-only log is not by itself concurrency-safe; allowlisted tools alone
  do not prevent a malicious query carrying data; an external reference need
  not reproduce past bytes. State the property actually proved.

Resolved matters must not be reopened as questions: mandatory parent chains,
free-function placement, optional GUI module, nonexclusive Observer, and
per-agent tool visibility. Registry storage remains a design choice within
the allowed alternatives. Escalate genuinely conflicting ownership or user
semantics; choose ordinary method names within the settled design.

## API, caching, and subscription research gate

API behavior and pricing are time-sensitive. First-edition receipts are
useful hypotheses, not current provider promises. At each vendor's
introduction, verify official contracts/model availability and exercise the
authorized API through the actual new program. Record date, route, actual
model, relevant configuration, and sanitized observed output. Distinguish
HTTP acceptance from evidence that opaque replay material was honored.

When caching is introduced, use the OpenAI documentation skill and current
official sources to establish whether and how the intended subscription OAuth
route is officially available for this application. Do not assume an API key,
a ChatGPT subscription, and another client's OAuth credentials are
interchangeable. Document the supported route and required authorization
without bypassing access restrictions. Research has **not** been performed
for this map.

Investigate Bill's reported recent cache bug with a bounded reproducible
experiment. Hold prompt prefix, model, endpoint, authentication route, and
serialization conditions fixed; record actual request-prefix evidence and
provider usage fields without secrets. Compare cold/warm behavior under the
documented rules. Separate a local serialization/cache-key defect, an
unsupported route, a provider issue, and missing observability; do not label
one established before evidence distinguishes them. A zero reported cache
hit is an observation, not by itself a diagnosis. Record an unresolved result
as unresolved and report any real access blocker rather than fabricating a
successful demonstration.

## Review and grading discipline

For each chapter, the author consults relevant history and this map before
the student sees the contract. The student receives the chapter, standing
skill, and preceding second-edition baseline. Missing teaching is repaired in
the contract, not supplied as private answer-key guidance.

Every protected promise needs an independent fixture and a deletion mutation
that demonstrably applies, still builds when appropriate, and loses the
intended check/points. Check-level mutation coverage is not property-level
coverage. A hardcoded package list, a root check satisfied by a different
binary, or a race test that never reaches the race gives false confidence.
Sources: [course policy P3/P9](../course-policy.md) and
[Chapter 22 coder review](../chapter-22-coder-review.md).

Maintain a per-chapter feature-to-user-action live checklist. Record actual
commands, dates, routes/models, observed results, and limitations. A local
fake, compile success, or internal helper test does not stand in for a user
successfully using the feature with the real backend. Failed access leaves
that acceptance item incomplete; it does not invalidate independent evidence
already collected.

Preserve the original skill's engineering safeguards as applicable: formatting
output must be empty, run the appropriate module's vet/tests before committing,
guard scripted edit anchors, preserve explanations of WHY until their
conditions are demonstrably gone, never weaken tests merely for green, and
report surprises. Architecture and usability need evidence in addition to
green tests.

## Chapter 3: first execution, without a later ownership repair

These are review recommendations for the author to make explicit in the
contract, not additional unpublished grader requirements. Freshly reread
first-edition Chapter 3's loop/tool/edit sections, Chapter 10's registry and
provenance sections, Chapter 15's batch-preservation rules, and Chapter 21's
dispatcher authorization section for this handoff.

- A simple working choice is an Agent-owned registry implemented in
  `internal/tools`, with a `common.Agent` owner interface. Engine obtains the
  tool service through its Agent interface; the two spokes do not import
  each other. Shared types/interfaces stay in common and execution behavior
  stays in tools. This recommendation does not change Bill's permission for
  a shared root registry with per-agent visibility. Record the chosen owner
  before implementation; avoid flat captured `ToolFunc` dependency closures.
- Declarations and dispatch must use the same Agent-visible tool set. Unknown
  or disabled calls get error results without execution. A tool result is
  data and cannot itself grant a capability. Introduce this with the first
  dispatcher, rather than waiting for skills or retrieved web content.
  [Chapter 10](../chapter-10.md), [Chapter 21](../chapter-21.md)
- Persist the accepted response once, then each dispatch fact before its
  side effect, then its result. Preserve issued IDs and execute in call order.
  Ordinary tool errors still answer their calls and do not silently skip the
  remaining batch. A nonzero command exit is a command result, distinct from
  failure to start or invalid arguments. Keep the one Agent append path and
  existing public observers; tools never write directly to a GUI.
- Specify persistence failure around a side effect: no execution when the
  pre-call record fails; after execution, a failed result write cannot undo
  the external change. Fault the Agent, report the limitation, and do not
  retry automatically. Likewise a later model failure cannot roll back an
  earlier accepted response or tool side effect.
- Define the round limit's counting and final-batch behavior. A useful rule
  is at most sixteen model responses, complete the last accepted call batch,
  then refuse a seventeenth request. This preserves pairing without adding
  jobs, retries, concurrent execution, or cancellation machinery. Replay and
  offline rendering must never re-execute tools. Later batch compaction needs
  these identities and boundaries, not its policy machinery today.
  [Chapter 3](../chapter-03.md), [Chapter 15](../chapter-15.md)
- Agent workspace settings determine path resolution, not filesystem
  confinement. Avoid process-global working-directory changes; two Agents
  must retain independent workspaces. A shell can access outside its working
  directory, so neither prose nor demonstrations may call it a sandbox.
- Teach overwrite refusal and an explicit zero/multiple-match edit policy.
  Preserve the six-tool synchronous scope. Use a mixed-content fixture for
  result ordering, independent declaration/dispatch checks, and real scratch
  workspace demonstrations of all six tools, failures, recovery, multiple
  rounds, and history retention on the supported APIs. Historical private
  corpus counts remain dated observations, not new measurements.

## Chapter 4: forward checks before the job contract

Reviewer recommendations from the original Chapter 4 lifecycle and audit,
revisited October 7, 2026. These are proposed contract clarifications, not
additional user rulings. Agent-owned Jobs and Ensemble-wide handle allocation
are the coordinator's current working choice.

- Define monotone terminal transitions: a late worker cannot overwrite
  `killed` with `done`. State completed-job waits, repeated kill, and kill/exit
  race outcomes, and check structured status instead of a substring.
- Specify consumption of pending one-shot limits by the literal next call,
  including another setter, kill, or invalid call. Keep consumption visible
  and the pending state on its owner; no hidden sticky configuration.
- Distinguish spool length from report cursor. Say whether omitted middle
  bytes count as consumed while remaining retrievable on disk. Pattern waits
  inspect unseen output, handle read-boundary splits, and do not repeatedly
  wake on an already reported prompt.
- The process worker/reader owns completion and spool closure. Shutdown must
  coordinate process termination, draining, terminal status, and joining;
  a returning dispatcher cannot close a still-writing reader's file.
- Distinguish process-group kill from marking a nonkillable Go goroutine.
  Neither workspace selection nor the output locator is confinement, and
  renderers must not automatically retrieve arbitrary spool references.
- State the intended Chapter 3 changes: PTY merges streams, job spooling
  retains output that inline reports cap, and a wake timer is not cancellation.
  Use a built short parity fixture instead of depending on a cold `go run`
  finishing inside the wake interval. Required debugger prerequisites must
  fail clearly rather than yield a skipped, falsely successful check.

The author received these recommendations before the Chapter 4 contract.
No new job implementation or behavior is claimed by this review.

## Final comparison: collect the evidence before writing the verdict

Bill requires this chapter. The following experimental design is a review
recommendation; the final contract should state which comparisons will
actually be performed. Keep it separate from the publication-time epilogue.

Compare two different questions explicitly:

1. **What improved in the completed software?** Compare fixed first-edition
   and second-edition snapshots at matched feature scope. Test actual user
   tasks and public integrations, failures and diagnostics, ownership and
   extension effort, and regressions. The final corrected artifact and the
   first cold student attempt answer different questions; retain both.
2. **What improved in the teaching?** Regenerate comparable scopes from each
   edition with independent cold workers, the same model snapshot, settings,
   tools, environment, budget, and exposure rules. Freeze the handoffs and
   record every intervention. A changed edition and a changed model in the
   same experiment cannot establish which caused an improvement. Cross
   edition and model conditions where feasible; otherwise name the confound.

Freeze a concrete evaluation rubric before inspecting the outputs. Keep
grader authors and evaluation challenges independent of the student, use
hidden challenge cases and mutation controls alongside published contracts,
and inspect successful real-user paths. For judgment-based code reviews,
blind edition labels and randomize presentation order where practical.
Require evidence and record reviewer disagreements; a judge's score or prose
is not a target for repeated tuning. Do not invent a numerical quality scale
whose precision the evidence cannot support.

Assess code quality through concrete changes: simpler ownership, eliminated
duplicate behavior, legible invariants, clearer failure boundaries, usable
public interfaces, and a small matched extension task. Fewer lines alone
cannot distinguish economy from missing functionality. Preserve losses,
regressions, unsupported features, and unresolved findings alongside wins.
Historical snapshots must remain available even when a new edition fixes
their mistakes.

Begin a per-chapter evidence ledger now:

- Initial and revised commits, source/build hashes, chapter/skill/handoff
  versions, evaluator version, and corresponding first-edition snapshot.
- Builder model identity and relevant settings, supplied context, accidental
  answer exposure, attempts, interventions, and the reason for each revision.
- Exact checks and failures, independently exercised properties, applied
  mutations and intended failures, plus the actual live feature checklist.
- Model input/output/cache observations, elapsed time, and human review or
  repair effort when measured. Keep build cost separate from runtime cost;
  date price conversions and distinguish API-key versus subscription costs.
- Code-review findings, accepted and declined changes with reasons, teaching
  changes, public usability evidence, and remaining limitations.

Keep credentials out of retained artifacts. Missing first-edition model,
cost, timing, or intervention data stays unknown. A retrospective estimate
may be useful if labeled and bounded, but cannot become a measured baseline.
Replicated trials can show variability where budget allows; one success
remains one observation, not a universal regeneration guarantee. Do not add
new paid experiments merely to fill a table before their scope is decided.
