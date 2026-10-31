# Compaction runtime design audit

Read-only audit by `compaction_runtime_audit`, an independent design reviewer,
2026-10-14. Baseline: `56ec71cbde8477b6959ad2fbc65f9237a0fa8346` in
`/Users/bill/projects/ensemble-compaction`. The working tree was clean when
inspected. This note is the only file written by this reviewer. No implementation,
test, grader, provider call, credential access, deployment read/change, or commit
was performed. Test names below describe inspected coverage, not new results.

Read in full: `AGENTS.md`, `.agents/skills/ensemble-coding/SKILL.md`,
`docs/agentic-codebook-workflow.md`, `docs/edition-2-carryover.md`, `book/voice.md`,
`book/chapter-writing-procedure.md`, and current Chapters 14 and 15. Read the
360-line handoff plan from the main checkout. This is the runtime portion of
Stage A; the coordinator and separate dependency reviewer own the full chapter
mapping, remaining preparation, and sequential baseline results. It is not a
formal prose review or acceptance of the whole handoff.

## Assessment

The plan is feasible through the existing composition root, Agent actor, Fleet,
Memory, History and Engine. It requires real changes at their existing boundaries;
wrapping `Engine.Compress` in a goroutine or calling `EnableFleet` everywhere does
not implement the requested contract. There is no need for a second scheduler,
event bus, mutable conversation builder, memory store, or retry framework.

Keep current chapter IDs while implementing. Chapter 15 explicitly requires a
finite direct call before the next main request (TL;DR rules 1–2 and §15.2/§15.5).
The carryover also teaches that choice. Background actor compaction deliberately
replaces that scheduling/construction contract while retaining selection,
first-person memory, replay, source preservation and ownership.

## 1. Construction and capability boundaries

Verified source:

- `agent/ensemble/fleet.go:26–57`: `FleetEnabled()` means `factory != nil`;
  `CreateChild` fails without a factory. `NewChildAgent` verifies the actual
  parent and calls `AgentSpec.Clamp` before construction.
- `agent/ensemble/registry.go:85–94`: `EnableFleet` installs a recursive factory
  using the parent's Config and replaces any prior factory. `Start` calls it.
- `agent/internal/fleet/fleet.go:72–102`: `Spawn` copies the whole parent Spec,
  sets owner-relative child paths, calls the application factory, verifies
  back-pointers, loads/saves, inserts the child, and uses its normal actor.
- `agent/ensemble/ensemble.go:274–320`: `BuiltinTools=false` still installs all
  fleet tools if fleet is enabled. Skills can install lifecycle tools, load
  primary/startup instructions and launch MCP peers. Recall defaults on.
- `agent/internal/agent/agent.go:217–242,274–281`: admission and declarations
  both use `allowedTool`/`toolEnabled`. SafeMode removes shell/supervision tools,
  not file reading, all management tools, memory or recall.

Smallest clear change: separate default child-construction availability from
public delegation-tool enablement. Preserve `SetAgentFactory` callers' existing
fleet behavior and honor a configured custom factory. Do not call `EnableFleet`
merely to make internal compaction work. A headless `NewAgent` client must retain
its existing public tool surface, including `BuiltinTools=false` clients.

Add a creation-only managed-compaction purpose alongside the optional submission
contract in the shared child specification; exact field names are the coder's
choice. Normalize that purpose in `NewChildAgent` before skill/MCP side effects.
It means fresh history, a dedicated role, no primary/startup skills or variables,
no recall, all memory bands disabled, no recursive compaction, and only `submit`
admitted by declaration **and execution**. The worker's internal journal/result
writes still use its inherited Sandbox; no filesystem tool is granted to the model.
Keep SafeMode, web and sandbox-root ceilings; if the child data path cannot be
constructed under them, fail visibly and retain source.

Do not copy the parent's submission schema or managed purpose into ordinary
grandchildren. Existing Spawn copies the entire Spec, so inheritance requires an
explicit reset. Ordinary children still restore same-ID histories; compaction
attempts need fresh, uniquely correlated child identities, including after
restart. Use the owning Agent's configured DataDir, never a deployment path.

`internal/agent/context.go:19–26` runs forcing for every Agent. A restricted worker
must also suppress checkpoint forcing, otherwise its sole submit capability can
be hidden by `forceHandoff` after a large response. Disabling memory settings alone
does not disable that code. Preserve model-specific visible-explanation enforcement.

## 2. Submission contract and persistence

All current submission boundaries are string-only: `common.ChildStatus.Result`,
`SubAgentManager.Submit(string)`, `Submission(string) (string,error)`, the manager's
`result`, and `submit.md` JSON lines containing `{"result":"..."}`. Join encodes
ChildStatus values; `get_submission` encodes another string result. Sources:
`internal/common/fleet.go`, `internal/fleet/fleet.go:330–356`,
`internal/fleet/management.go:47–70`, and `internal/tools/management.go:80–89`.

Recommended compatible boundary:

1. Keep old Spawn/Submit/string retrieval entry points as wrappers. Add an explicit
   options/raw-JSON path rather than changing every caller's return type to `any`.
2. `submit_schema` describes the complete direct argument object of `submit`.
   Absence retains the existing `{"result":"text"}` behavior. Custom compaction
   uses exactly `{"memory":"..."}` without an extra wrapper.
3. Compile/validate and defensively copy the schema before factory invocation.
   Declare that canonical contract and validate raw arguments before disk output,
   manager completion or join notification. Old `Submit(string)` must traverse the
   same validation boundary; it cannot bypass a custom schema.
4. Preserve custom payloads as owned `json.RawMessage` or an equivalent JSON value
   in status, joins, retrieval and durable records. An additional explicit payload
   field can preserve the old `Result string` surface. Do not put JSON into that
   string and call it structured support. Copy incoming and outgoing byte slices.
5. Make custom durable records distinguishable from legacy lines and retain enough
   contract identity to interpret them on reload. The exact record envelope is a
   Stage B design choice. Continue reading old `submit.md` records. A historic
   fallback submission is not proof that a newly started task has completed.

No genuine general schema validator was found in the current core/GUI dependency
lists or implementation. `RegisterTool` and MCP discovery use `json.Valid`.
`internal/tools/tools.go:77–93` checks required presence/null and Go decoding;
it ignores extra properties and does not evaluate nested schema semantics.
Metadata is insufficient. Introduce one shared validation boundary in a responsible
implementation package, reached through shared contracts, rather than separate
submit/compressor validators or behavior in `common`.

Before coding, freeze the supported dialect/keywords, root-object requirement,
null/type/integer/array/nested/enum semantics, limits and reference policy. A
well-defined supported dialect with explicit rejection of unsupported constructs
is acceptable; a silent partial validator is not. Prefer a reviewed validator
dependency over a handwritten imitation of general JSON Schema. Disable external
reference resolution; reference support and complexity limits must be tested at
spawn as well as submit. No particular dependency was researched or selected in
this source-only audit.

The renderers currently pass raw schemas through: `anthropic.go:122`,
`openai.go:142`, `gemini.go:133`, `responses.go:99` (Responses sets `strict:false`).
Wire-schema tests must cover all supported routes, including the exact memory
schema, without relying on provider enforcement. General structured submission
does not require converting every ordinary tool result into structured data.

## 3. Submission is earlier than quiescence

`fleet.Submit` writes/closes `submit.md`, then sets `submitted` and signals the
parent from inside the running tool handler. `Join` returns on that flag.
`agent/dispatch.go:13–23` waits for the tool batch to drain, records `turn_ended`,
then finishes. `actor.go:305–340` marks Fleet/activity/TurnEnded **before** its
final log save; only the reliable request Result contains that save error.
`Spawn` currently discards the Result value and retains only its notification.
Status can still say submitted after a failed request save.

Use the already-created reliable request completion for managed tasks; retain its
error. Never accept a status transition or lossy observation as the completion
barrier. Successful completion must include the submit tail and save, then stop
and join the child and checkpoint it before dismissing references. `Flush` waits
for turns and saves, not all background jobs; `Shutdown` joins workers/jobs;
`Save` checkpoints after shutdown. Existing `GracefulShutdown` supplies much of
this ordering, but alone cannot recover a prior Result error that Spawn discarded.

Make managed cleanup idempotent against explicit kill and owner/application
shutdown, using retained child identity rather than a reused string ID. Record
cleanup/save errors rather than treating removal as success. The parent needs a
terminal outcome and child trace location after dismissal.

There are two concrete retention leaks to close for repeated workers:

- `ensemble.go:322–324` appends every constructed Agent to `e.agents`; child
  shutdown removes only the manager map (`fleet.go:196–211`). There is no matching
  application removal. Also audit construction failures after root registration.
  The default DataDir uses `len(e.agents)+1`, so removal must not introduce
  accidental identity reuse.
- `gui/gui.go:98–105` calls `view(a)` for every observation. `gui/tree.go:21–27`
  retains the Agent and copied events forever, even if never selected. Tree refresh
  does not prune views. Dismissal must retire these references and handle browsers
  selecting that child, through the existing observer/tree machinery.

Do not add an archive browser or a new general registry for this task. Preserve
trace files and parent outcomes while removing runtime ownership references.

## 4. Parent lifetime and finite worker attempts

Current compression is tied to `compressed.turn`, uses the active request's
context, and drops completion when that pointer is no longer active
(`internal/agent/memory.go:15–20,76–100`). `checkpoint` sends the next main request
only after compression/cascade completion. A background task must instead live
on the runtime/service with one slot for an in-flight or commit-pending attempt.
Workers return immutable messages to the existing owner mailbox.

Recommended lifecycle: selected → launched → candidate/failed/canceled → committed
or invalidated → cleanup complete. These are task facts, not a parallel Agent
execution loop. The actor admits completion only at idle or before a new render
with no incomplete main tool batch. A candidate arriving during an HTTP request
waits; do not mutate context so that the in-flight request appears to have contained
it. Unrelated appended dialogue does not invalidate selected completed work.

Recommended cancellation contract: normal parent-turn completion **and parent-turn
interrupt** leave a useful compaction child alive, matching ordinary background
children. Explicit worker cancellation/kill, owner shutdown, and relevant memory
disable invalidate the task and stop its wasted work. The plan §7 already prefers
ordinary background-child semantics; adopting it appears within the authorized
handoff, not a new Bill approval gate. Record this deliberate replacement of
Chapter 15 and `TestMemoryInterruptedCompressor`, rather than quietly deleting a
failing old expectation. If Bill explicitly intended interrupt to cancel all
owner compaction, that contrary instruction would take precedence.

Ordinary child Idle supplies three extra corrective wakes (`fleet.go:369–385`),
and MaxToolRounds bounds tool batches, not all model-only replies. Managed workers
need an explicit finite attempt policy and no automatic idle wakes or immediate
same-source relaunch after failure. A small existing round budget can allow a
bounded correction after invalid submit; document actual maximum requests as well
as dispatched batches. A text-only finish without submit is a failed compaction.
No new wall-clock model deadline is needed. A later explicit checkpoint may retry.

## 5. Selection, generations and safe acceptance

Keep selection in Memory. `memory.go:190–235` selects oldest whole completed
checkpoint segments, including handoff notes and role labels, excluding independent
skills/declarations and the unfinished tail. Tool parts are removed by the existing
handoff path; source building keeps readable text. Graduation (`:239–295`) uses
canonical files and watermarks, at least two files, and includes existing MEMORY.md
when producing the terminal band. Keep those semantics.

Capture task identity, policy revision, selected boundaries/identities and exact
selected bytes when choosing work. At commit, verify those selected entries/files
and relevant source/destination settings. Do not rerun selection against the entire
now-growing history or hash all history. A monotonic policy revision must remember
disable/re-enable even if final settings equal the old values; comparing final
booleans alone permits stale results. Do not tie it to unrelated GUI preferences.

Disk revalidation matters even when context has not rescanned a human edit.
Fingerprint the selected graduation files and the existing terminal MEMORY.md
input. Protect a newly allocated output identity from collision with a file created
while the worker ran. A terminal summary must not overwrite a human's newer curated
memory. Reject whitespace-only memory separately from schema validation. Validate
at the owner commit boundary, not merely at child submit.

`memory.Scan` parses range endpoints but does not reject `Thru < File`;
`Graduation` trusts the last selected entry's Thru, while History retirement compares
that endpoint with each entry's File. This is a source-supported possible no-progress
fold. Establish it with a focused failing test before changing it. It is not evidence
for the reported live Conversation compressor failure. A fold must retire selected
work rather than schedule the same source indefinitely.

## 6. Disk/event ordering needs crash-window coverage

`Memory.Write` (`memory.go:299–310`) installs the replacement at its active `.md`
path before the actor appends cut/retirement and population. Archive follows those
events (`agent/memory.go:105–135`). History holds proposed removals until matching
band/Thru population (`history/reduce.go:82–105`), preserving source in pure replay.
That pairing should remain.

Two separate disk/replay windows are not solved by the existing pair:

- Crash after the active replacement file is written but before population:
  restart Scan can inject that uncommitted output alongside retained source.
- Crash after population but before Archive: restart Scan can reintroduce old
  source files whose entries the completed event pair retired.

The current abandonment test restores a launch/cut with no matching population;
the held-request test interrupts before output is written. Neither establishes
these disk windows. Add narrow Memory-owned staging/recovery for its own attempts,
with enough task metadata to reconcile completed versus abandoned work before
running startup scans. Preserve replacement bytes before retiring source and keep
staged/uncommitted outputs out of ordinary scans. Do not build a general transaction
framework, and do not claim paired history facts alone make corpus files atomic.

Extend launch/outcome correlation with a task identity; current `(band,Thru)` can
repeat on abandoned retries. Record failure/cancellation/invalidation/abandonment
and cleanup without changing the dialogue turn state. `error_occurred` currently
sets History's turn idle, so it is unsuitable as an unqualified background-outcome
event while the main request is running. Replay must apply only recorded decisions,
clear abandoned pending work and never spawn or infer a new model result.

## 7. Roles, sizing, and accounting

Current `Engine.Compress` constructs fresh dialogue but calls the same Engine's
Render, which uses the owning main SystemPrompt (`engine/compress.go:14–22`,
`engine.go:405–425`). A dedicated child fixes that role-source conflict only if
construction also removes the inherited primary. The ordinary child instructions
mention tools a restricted child lacks, so supply a coherent dedicated purpose.
Ask for the **source owner's** first-person memory, with delimited attributed data,
decisions and reasons, failed approaches, uncertainty and constraints. Retain the
ratio as guidance and exclude visible submit narration from the memory payload.

Sizing is not fully settled by using the normal Engine: Anthropic currently caps
output at 1,024 or 4,096 tokens (`anthropic.go:133–136`); Gemini at 4,096
(`gemini.go:142`); Responses has no general configured output cap here, and the
subscription capability table omits `max_output_tokens`. Do not equate byte budgets
with tokens or add a tiny universal cap. The coder should justify supported-route
sizing and test truncation/failure retains source; this audit selects no new cap.

The direct compressor currently contributes once to the parent's Engine totals
through `send/account`, without updating Engine.last or History.LastUsage.
`agent/memory.go:142` forces from History.LastUsage only. A real child naturally
accounts to its own Engine. Preserve that owner and the main-only forcing numerator;
never append child `response_ended` events to the parent's dialogue to transfer cost.

Freeze the reporting choice: parent-only meter plus separately retained worker
usage is the smallest change; if the existing owner total must still include
compaction, use an explicit auxiliary usage record/import once at completion,
including vendor/model/tier and failed attempts with measured usage. Do not add both
that import and the same child again in a fleet aggregate. Do not overwrite the
main last-request meter or restore historic usage as new process spending. Parent
outcomes should retain worker usage before its runtime is dismissed. The coordinator
owns the accounting/chapter-placement decision; this audit does not silently change
the meaning of the existing session total.

## 8. Tests and scoped file ownership

Useful existing real-runtime coverage to retain/adapt:

- `ensemble/memory_test.go`: three-provider compression, forcing, disk toggles,
  waterfall, replay abandonment, failure retention, whole segments, enabled restart
  edits and interrupted compressor. Its HTTP fake holds a global mutex through the
  response callback and uses fixed request ordinals; a held compressor can therefore
  block the fake's main request independently of runtime behavior. Use per-request
  barriers and concurrent-safe classification for new background-progress checks.
- `ensemble/fleet_test.go`: hierarchy/questions, main-vs-child role, lifecycle and
  same-ID restore, bounded idle correction. `ensemble/management_test.go` explicitly
  Flushes after Join, confirming those are different boundaries.
- `ensemble/actor_test.go`: reliable completions, save failure, interruption and
  shutdown. `ensemble/visible_reasoning_test.go`: real actor refusal, stream order,
  bounded correction. `gui/selection_test.go`: submitted turn and replay state.

New checks need real actor/engine/fleet with only the provider faked: valid/invalid
structured submission including old direct API attempts; schema/result buffer
mutation; declaration and forged-call rejection; main progress behind a child
barrier; completion after parent idle/interrupt; explicit cancellation; selected
source and terminal-destination edits; disable/re-enable; no overlap/relaunch;
save failure after submit; crash points around disk/events/archive; cleanup races;
repeated registry/GUI retention; and accounting exactly once. Test the intended
failure with targeted mutations, preserving the existing source-preservation checks.

Expected owned source areas, subject to coder's smallest implementation:

| Boundary | Existing paths |
| --- | --- |
| Shared submission/task vocabulary | `agent/internal/common/{skills,fleet,memory,types}.go` |
| Factory, construction, detach | `agent/ensemble/{fleet,ensemble,registry,spec}.go` |
| Submit tools and fleet lifecycle | `agent/internal/tools/{fleet,management}.go`, `agent/internal/fleet/{fleet,management}.go` |
| Parent scheduling/admission | `agent/internal/agent/{actor,memory,dispatch,context,agent}.go` |
| Corpus and replay | `agent/internal/memory/memory.go`, `agent/internal/history/{history,reduce}.go` |
| Role/sizing/accounting only as needed | `agent/internal/engine/{compress,engine,usage,anthropic,gemini,responses}.go` |
| Visible dismissal and structured display | `agent/gui/{gui,tree}.go` and affected renderer/tests |

`ensemble.go`, `registry.go`, and `common/skills.go` are announced shared hotspots
with Cadence. The audit also establishes a concrete need to coordinate GUI view
cleanup. No evidence here requires editing live configuration/startup files or
Cadence's memory repository. Do not claim exclusive ownership from this table.

Stage A can proceed to the coordinator's contract/mapping decision. Before Stage B/C
implementation, record schema semantics, structured persistence shape, finite worker
attempt bounds, task acceptance/recovery protocol and accounting attribution.
The model/factory/lifecycle gaps above are engineering work under the handoff;
the final chapter move and historical snapshot policy remain Bill's stated gates.

## Stage B/C independent review resumed

2026-10-14, reviewer `compaction_runtime_audit`, checkpoint `7a362ec2`.
The coordinator reports Bill approved the recommended mapping, staged accounting
and reviewed replacement exports. Those Stage A decision gates are resolved;
the original audit above retains its then-current questions. Accounting will
preserve the existing owner's session-total promise by attributing measured
auxiliary usage exactly once, while keeping latest main-request usage separate.
Stage C must retain the actual worker/model/tier identity and avoid re-importing
historical spending on restart.

After compaction the reviewer reloaded the entire mandatory
`.agents/skills/ensemble-coding/SKILL.md`, workflow, carryover, voice and chapter
procedure, recovered truncated read ranges, and checked the clean worktree at
the checkpoint. Current Chapters 22 and 23 were also read in full for the
submission and fleet contracts. The original full Chapters 14/15 reading is
recorded above. No implementation, test or acceptance result is claimed by this
preparation. The Stage B coder is separate from this reviewer; its concrete
schema/API/storage/ownership design is awaiting review before implementation.

### Stage B design acceptance

The independent reviewer accepts the design in
[`codex-structured-submissions.md`](codex-structured-submissions.md), including
the coder's subsequent `SubmissionData` refinement, for implementation. This is
a design gate, not formal implementation acceptance; zero formal code-review
rounds have occurred. The coordinator and coder were notified directly.

The accepted shape preserves `Spawn`, `Submit(string)`, legacy `Result` fields
and `{"result":"..."}` durable rows. Custom contracts cover the direct submit
argument object, use an owned raw-JSON `Payload`, and store an explicit version-1
schema/payload envelope. `SubmissionData` retains default/custom identity even
when a custom payload happens to look like a legacy result. Legacy string retrieval
reports a clear error for custom results. The existing string-valued `Ask` and
`ask_agent` completion can render custom JSON as text; structured callers keep
the raw payload. Durable fallback never marks a newly spawned task complete.

The pinned genuine validator is `github.com/santhosh-tekuri/jsonschema/v6`
`v6.0.3`, using an explicit bounded Draft 2020-12 subset. All references and
unsupported keywords are rejected; a refusing loader provides a second boundary
against resource fetching. Local dependency-source inspection confirmed the
draft/loader/resource APIs, exact-number parser and standard Go regexp default.
No online research or dependency mutation was performed by the reviewer.
The proposed size/depth/node/numeric ceilings are explicit admission limits,
not truncation or summarization targets. Their implementation and tests still
need review; declaring a limit alone does not prove bounded evaluation.

The Fleet owns the immutable compiled contract. Construction validates/copies
before side effects, Spawn validates before a custom factory, and the returned
child must retain the requested contract. Every submit entry point uses that
contract before persistence and completion; copies also protect outgoing Spec,
declaration, status and retrieval buffers. Ordinary grandchildren reset the
parent's custom schema.

Two narrow clarifications accompany acceptance: absent-schema submission retains
the old acceptance of extra keys and normalizes to its result text/legacy row;
custom submission retains the full object. The keyword admission walk visits
schema-bearing positions only, not object keys inside enum, const, default or
examples data. These clarifications prevent compatibility loss and accidental
rejection of valid data without adding another validator or result subsystem.

The coder's planned real-runtime/provider-fake checks cover corrected invalid
submissions, pending joins, public API enforcement, defensive copying, hierarchy,
ownership, durable retrieval/reload, adapter declarations and GUI replay. Review
will assess the actual assertions and targeted mutation outcomes. No new tests
have been run by this reviewer at this design boundary. Stage C's quiescence,
restricted construction, background scheduling and accounting remain separate.

### Stage B pre-submission source review

The coordinator resumed production review while the coder completed its report,
with the full governing context still active. The worktree had advanced to
`a8eec84a`; implementation was uncommitted. This review does not consume a formal
correction round. Findings were sent directly to coder and coordinator before
the consolidated submission:

- GUI result assertions initially searched for marker words instead of decoding
  the actual submitted object. They could not establish the claimed nested-data
  preservation. The coder is strengthening both live delivery and replay checks.
- Separate schema/input ceilings did not control their multiplied validation
  work. Dependency source collects errors across all array elements and retains
  each element's missing required names; formatting those errors can amplify a
  small input. No large or small executable amplification probe was run before
  this finding. The agreed correction adds an explicit conservative combined
  byte budget before validation, without replacing schema semantics or claiming
  a CPU deadline.
- Incoming byte caps preceded JSON's HTML escaping. An admitted annotation or
  payload containing many `<` characters could be serialized above the same cap
  enforced during durable rereading. The accepted shape would then reject its
  own stored record. This was initially source-identified, not reproduced; the
  coder is adding canonical-byte refusal checks and small restored-defect cases.
- Moving factory checks into the public `CreateChild` path is appropriate. Its
  inherited wrong-owner branch must not shut down an unrelated returned Agent.
  In particular, returning the caller itself from a factory invoked by Spawn
  would re-enter the manager lock during shutdown. Ownership rejection and
  disposal of a newly created, correctly owned wrong-contract child are distinct.
  The coder received this narrow finding and a real-Agent survival test suggestion.

A host disk-full event temporarily paused builds/tests. The reviewer had no
active process and ran no builds during that pause; the coordinator cleared
regenerable caches and subsequently released bounded checks. Source analysis
and later executable evidence retain their distinct statuses.

### Stage B formal implementation review: accepted

October 14, 2026. Independent reviewer `compaction_runtime_audit` accepts Stage B
at initial formal review, with zero formal correction rounds. All pre-submission
findings above remain attributed assistance. After the final context compaction,
the reviewer again read the entire `.agents/skills/ensemble-coding/SKILL.md` and
all four governing/craft documents before this decision, then reread the frozen
production diff, schema/storage implementation and consolidated coder report.
HEAD is `a92aaad9`; the Stage B source is still uncommitted for the coordinator
to checkpoint. The sorted changed/new `agent/` path-and-SHA256 manifest hashes to
`516f2d8205da583393b0fc2563f1032ddaebf9d295433985db0fd4a5f08f82ce`.
All 214 files in the coordinator's grader source manifest still match disk.

No Stage B blocking finding remains. The final public `CreateChild` boundary
also refuses nil or foreign parents before invoking the factory, matching the
existing required-parent construction rule. The real-Agent tests cover this
entry point directly and through Spawn. A factory cannot discard the promised
schema; wrong ownership does not authorize shutting down a live foreign Agent
or the caller. This closes the additional pre-submission parent-boundary finding
without adding lifecycle policy outside this stage.

The implementation follows the accepted ownership plan: common holds data and
interfaces; Fleet owns the immutable compiled contract, validation and durable
submission behavior; the composition root validates before construction effects.
The genuine validator owns schema semantics, with explicit admission restrictions
and no external resource resolution. Exact-number decoding, raw and canonical
byte checks, and the combined byte budget address the identified precision,
reload and repeated-work risks. The combined budget is deliberately conservative
and does not establish a measured CPU deadline; no amplification stress run was
performed or needed for this acceptance.

Legacy string submission and rows remain compatible. Custom JSON survives
declaration, tool dispatch, status, join, typed retrieval and durable rereading
without double encoding. The explicit envelope preserves contract identity even
when custom data resembles a legacy result. Validation and successful writing
precede completion publication; historical fallback does not complete a new task.
Spec, declaration and returned payload bytes are copied. The existing text-valued
Ask/display API renders compact custom JSON, while structured readers retain its
object. GUI production code is unchanged; the new test checks full nested data
in both actual subscribed live frames and restored replay, with no inference on
reload.

Independent reviewer executions:

- `go test -race ./internal/fleet -run TestSubmission -count=1`: pass, 1.080s.
- `go test -race ./ensemble -run TestStructuredSubmission -count=1`: pass,
  1.429s after the HTTP fixture correction. The earlier failure was the fixture's
  missing Anthropic usage, and remains a failed setup rather than an ownership
  defect. After the final parent guard,
  `go test -race ./ensemble -run 'TestStructuredSubmission(Construction|FactoryOwnership)$' -count=1`
  passes, 1.036s.
- GUI `go test -race . -run TestStructuredSubmissionGUIReplay -count=1`:
  pass, 1.079s with the exact live/replay assertion.

The consolidated coder evidence supplies passing final focused Stage B race
checks, full GUI race (11.856s), both module vets and full ensemble race
(91.778s). That full ensemble run began before the final required-parent guard;
the final focused race run covers the added guard. Full core short/race is not
claimed green: its CLI human-sink case reproduces on baseline. The coder's
standalone browser glob was an invalid suite invocation, and its separate speech
persona failure reproduces on exact `a8eec84a`; neither is silently passed or
repaired in this stage. ENOSPC link failures and the initial missing GUI
subscription are retained as setup failures. No live provider, deployment or
publication result is claimed.

The coordinator's advertised Edition 2 compatibility commands using current
chapter numbers and explicit `./agent` score Chapter 22 200/200 and Chapter 23
100/100, both exit 0. The reviewer inspected their JSON receipts and verified
the complete source manifest against disk. These are the revised compatibility
graders; historical original scores are unchanged. Receipts are under
`/tmp/ensemble-compaction-baseline/grader-{22,23}-stageb*`, with
`stageb-agent-source.json` binding the source.

Mutation metadata and failure excerpts were inspected in
`/tmp/ensemble-structured-submissions-mutations/`. They cover the reported 21
distinct defects over 24 executions, including repeated location checks and the
factory recheck. Canonical schema/payload and combined-budget omissions each
admit the prohibited value; restored foreign-agent shutdown fails with
`agent stopped`; removing parent admission invokes the factory; double encoding
fails the actual nested GUI result comparison. Raw-byte corruption is caught on
decode. The earlier factory run lacked a valid baseline until its peer supplied
usage, so the repaired baseline/recheck supplies that claim. Declaration aliasing
also causes cleanup serialization errors, and several schema-semantic assertions
overlap. These are sensitivity observations, not 24 independent proofs or proof
of every upstream keyword. The tests use real runtime objects with small external
HTTP peers, without mocks or a new test framework.

The independent size count agrees with the coder. Across 13 changed/new production
Go files, physical/active/comment-only lines move from 2212/1701/332 to
2596/2007/388, a net 384/306/56. Three new test files contain 764 physical,
712 active and 33 comment-only lines. Active counts exclude blanks and
comment-only lines; moved submission code is counted on both sides, avoiding a
false wholly-new-code claim. Module dependency files are separate. This compares
the maintenance change against its own `a8eec84a` predecessor; there is no
first-edition structured-submission implementation to claim as its counterpart.

The growth implements the requested contract and storage boundary rather than a
new scheduler, result browser or schema framework. Useful comments explain
ownership, copies, refusal versus semantic validation and the write/completion
ordering. Exported API additions document their contracts. Comment-only lines
are about 16% of nonblank production lines in this changed-file set, below the
rough 20% teaching target but with the important new reasoning present; padding
would not improve the code. Tests are longer because they exercise actual
provider adapters, factory ownership, persistence and GUI delivery, not because
they reconstruct the implementation behind mocks. No scope or KISS issue remains.

Stage C still requires its concrete design gate. Stage B's submitted state is
not worker quiescence: retain reliable turn completion through the actor tail
and save before cleanup, and remove owned root/GUI references when that lifetime
ends. The earlier disk/event findings now have coordinator-run controlled
saved-state boundary/restart reproductions on exact `a8eec84a`: uncommitted output
is reinjected before population, and retired sources return before Archive.
No-write/completed-Archive controls pass. These are not forced process kills;
the probe and logs are in `/tmp/ensemble-compaction-window-probe/`, probe SHA256
`ef869c4e0945e84c763c837dfd81f7a9f1ca85941b4aa4139d1418b5c5fae2af`.
Acceptance of structured submissions does not claim those memory/lifecycle
problems are fixed, authorize chapter moves now, or satisfy later live gates.

### Stage C concrete design: accepted for implementation

October 14, 2026. The independent reviewer accepts the revised
[`codex-background-compaction.md`](codex-background-compaction.md) for Stage C
implementation. Stage B is checkpointed at `a0acfb5e`; subsequent Cadence note
merges and accounting notes do not change that runtime baseline. The entire
mandatory `.agents/skills/ensemble-coding/SKILL.md` and four governing/craft
documents remain loaded from the final B review, with no intervening compaction.
The reviewer read both complete versions of the concrete C design and refreshed
the actual Memory, actor completion, Engine usage, factory and admission paths.
This is a design acceptance only; no C implementation, formal correction round,
runtime test or recovery success is claimed.

The proposed one owner-lifetime attempt reuses the existing actor, Fleet child,
reliable request Result and mailbox. Ordinary turn completion/interruption leaves
it useful; explicit kill, owner shutdown and relevant policy changes cancel or
invalidate it. Commit waits for a safe parent boundary and exact selected-source,
destination and policy-generation checks. Unrelated appended dialogue does not
invalidate the result. A known 95%-full main context may wait for that outstanding
attempt; failure produces an explicit error rather than another full-context send.
This is a stated exception to ordinary background progress, not a model timeout.

Restricted purpose is enforced in construction, declaration and dispatch, with
no copied primary/skills/MCP, recall, memory recursion or arbitrary tools. It grants
only the needed submit operation to the worker. Source inspection confirms
`BuiltinTools=false` currently selects default registration rather than an inherited
security ceiling; keeping the parent's declarations and FleetEnabled unchanged
honors that contract. Actual SafeMode, web and sandbox ceilings remain binding.
The revised design accurately distinguishes normalized input and checked public
construction from an arbitrary trusted custom factory's own pre-return effects.
The root cannot retrospectively prevent those effects by rejecting its result.

The finite policy is three Engine.Send invocations, including correction/resend
paths, with existing bounded authentication/HTTP retries explicitly separate.
Prose-only completion fails and the purpose cannot acquire another task to reset
its budget. The reviewed provider output policy is deliberate and worker-specific;
its live acceptance and semantic quality remain Stage F obligations. Following
review feedback, adapters must detect actual length/incomplete terminal reasons
before dispatch even when a submit's JSON is syntactically valid, while retaining
measured usage. Malformed JSON alone would not test that distinct failure.

The single hidden candidate plus correlated journal receipt is justified by the
two reproduced disk/event windows. Memory owns file staging, promotion, source
archive and recovery; History atomically projects the matched cut/population and
retains the receipt through snapshots. A latest terminal outcome recognizes the
final cleanup window, and no new attempt starts before staging removal succeeds.
Recovery precedes startup/re-enable population and recall. After journal commit,
an unfinished disk finalization is a recoverable committed state, never abandonment.
The clarified hold blocks main rendering and recall as well as corpus scans until
finalization succeeds or the turn receives an explicit error. Human-edit conflicts
preserve copies and receipt, rather than overwriting them. The design claims
process-crash ordering, not power-loss durability or transactions against external
editors. The implementation must still prove each described boundary.

Completion, cancellation and accounting share a real child drain and save attempt
before reference removal. The revised accounting proposal uses the active attempt's
identity and accounted bit instead of an unbounded Engine task-ID deduplication set.
Engine remains owner of measured model/tier buckets and adds final auxiliary usage
once, including failed/canceled work, without replacing latest-main usage or replaying
historical spending into a fresh session. Exact-child shared dismissal prevents
manual kill races from removing a replacement. Root and GUI references must go
with the Fleet entry; retained files provide the audit trail.

The reversed-range issue now has a separate coordinator baseline reproduction:
`ranges_test.go` under `/tmp/ensemble-compaction-window-probe/`, SHA256
`a513c5c1fde6fa7bbd26ddc7ea0924394117f0d9264b41a421784be4056b27dd`.
Malformed coverage retires nothing and immediately selects the same source; valid
filenames pass the control. C's range admission and exact retirement-prefix checks
address that demonstrated progress obligation, without claiming it caused Cadence's
earlier conversation-compressor failure.

Validation must hold the external request outside capture locks. The coordinator's
separate real-HTTP probe reproduces the shared fakevendor Route lock blocking main
requests and Requests(); it is Stage D fixture evidence, not a compaction result.
The existing Chapter 23 reverse-proxy barrier is a narrow later grader option.
C's independent concurrent HTTP peer can exercise actual main progress now. Tests
remain real runtime plus external fakes, with targeted mutations and genuine
filesystem failures; no new persistence mock or generic task framework is needed.
No unresolved design blocker remains. Production complexity, concrete ownership,
boundary tests and required compatibility results still face independent code
review before Stage C acceptance.

### C1 range-admission slice review

October 14, 2026. The range-only slice is accepted for its scoped checkpoint;
this is pre-submission assistance, not full Stage C acceptance or a formal C
correction round. Mandatory skill/four-document context remains active without
intervening compaction. The reviewer inspected the exact `memory.go` diff,
`compaction_ranges_test.go`, its shared real-Agent helper, the coder's C1 report
and all three targeted mutation logs. No runtime edits were made by the reviewer.

Scan now refuses reversed and extra-segment coverage names. Graduation separately
checks restored entries, whose origin need not be today's Scan, for valid identities
and ordered disjoint coverage before selection. This directly protects the existing
prefix-retirement mechanism; watermarks and valid selection policy are unchanged.
The small guards and their ordering comment belong with Memory policy and add no
new abstraction. The new tests retain actual Memory/History and filesystem behavior.
For this slice, Memory physical/active/comment-only lines move from 331/294/20
to 357/317/23; the range test and shared helper contain 171/158/7. Counts exclude
blank and `//`-only lines from active code, separating production from tests.
The shared helper will also support the later recovery slice. The added production
comment explains the failure mechanism; existing explanations remain intact.

Independent focused race execution of `TestMemoryGraduationCoverage`,
`TestMemoryScanMalformedCoverage`, `TestReversedRangesCannotRepeatFold` and
`TestMemoryWaterfall` passed in 1.143s, with test-process provider/runtime prefixes
stripped. The coder's final affected race also passes in 1.143s after the test-only
refinement below; targeted vet passes. The two original restoration mutations
fail the intended scan and restored-entry assertions. Their paired failures are
not four independent proofs.

Review requested useful documentation on two exported tests and explicit selected
and surviving file identities instead of only a survivor count. Both are fixed:
the valid control selects 1/2 and retains exactly 3. Reversing the retirement
comparison leaves one wrong survivor and now fails specifically with ID1 in
`/tmp/ensemble-compaction-stagec-mutant-prefix.txt`. This third mutation establishes
the stronger prefix assertion. The reviewer inspected that result rather than
inferring it from the final passing test.

The old Write/event/Archive restart probes remain failing and are excluded from
this checkpoint, as the coder's note explicitly records. The range correction
does not fix either disk/event recovery window or implement background workers.
No broad suite, grader, live-provider or complete Stage C success is claimed.

### C2 restricted-construction and finite-execution slice review

October 14, 2026. C2 is accepted for its scoped checkpoint, as pre-submission
slice review with no formal Stage C correction round. The full mandatory skill
and governing/craft context remains active without an intervening compaction.
Review covered the nine owned production files, worker tests, consolidated C2
evidence and fourteen mutation outputs. Concurrent C3 Engine/provider edits were
excluded from this decision and from the independent verification below.

Construction normalizes the managed purpose before invoking the trusted custom
factory and again at the actual constructor, before primary/skill loading. The
child gets its dedicated owner-memory role, disabled memory/recall/forcing and
only submit admission, while the parent's declarations stay unchanged. Actual
owner and sandbox restrictions continue through NewChildAgent. Default internal
construction and public delegation remain distinct: EnableFleet retains a supplied
factory, and ordinary Spawn explicitly refuses when FleetEnabled is false, preserving
the earlier missing-factory refusal instead of starting an unusable ordinary child.
The root's trusted factory callback is not claimed to be sandboxed.

The actor owns one-task and three-Send state for the worker lifetime. Missing-submit
prose fails without ordinary idle correction; extra or forged calls reject their
entire batch before jobs can partially succeed. Existing visible-explanation policy
still applies. The finite counter includes correction sends and cannot be reset by
a second task. No new worker scheduler or retry framework is introduced.

Review found that public AgentSpec.Child still inherited Compaction even though
ordinary Spawn reset it. The coder now resets both purpose and schema in that
public derivation helper, with an actual constructed-child control. Review also
replaced substring-based schema assertions with structural comparisons of the
declaration and actual captured Anthropic input_schema. Those corrections are
attributed reviewer assistance, not independent initial discoveries.

Independent verification used a temporary archive of exact `c8ec3d4b`, applied
only the nine C2 production diffs and copied only `compaction_worker_test.go`.
No C3 Engine edits/provider tests or intentionally red recovery probes entered
that archive. Focused C2 plus all Stage B structured-submission race tests pass
in 1.526s. The sorted C2 path-and-content SHA256 binding is
`9a2443743da6ddc05ebe4e50f14a78e1260a97e2bb53c7b35a1aef0a7a6baff7`.
Scratch was removed after completion; test-process provider/runtime prefixes were
stripped. The coder separately reports a final isolated C2 race pass in 1.141s,
earlier inherited Fleet/Named/Management/sandbox race in 3.228s and affected vet
success. Its mixed-tree run is retained separately and is not relabeled C2-only.

Mutation failures were inspected at
`/tmp/ensemble-compaction-stagec-worker-mutant-*.txt`. They detect changed startup
normalization/role, default/custom factory selection, returned purpose, tool
registration, send count, one-task enforcement, missing submission, partial
multi-call success, forcing and descendant inheritance. The final helper-reset
mutation fails the public helper assertion; changing additionalProperties to true
fails the exact schema comparison in all four constructor cases. The startup
failure proves attempted loading of a nonexistent skill, not an actual external
MCP connection. Registration refusal and forged-batch rejection remain distinct
boundaries; fourteen mutations are not proof of every future managed-worker path.
The initial test's mistaken expectation that forbidden registration should succeed
remains documented as a fixture error.

Across the nine production files, physical/active/comment-only lines move from
2759/2186/371 to 2864/2274/385, net 105/88/14. The new test has 307/293/8.
Active counts exclude blanks and comment-only lines; production and tests are
separate. Comments explain startup timing, registration versus permissions, bounded
task lifetime and whole-batch refusal. The small purpose checks extend existing
owners and preserve ordinary behavior; no scope or KISS blocker remains.

Normal memory still uses the old compressor. StartCompaction, provider sizing and
terminal reasons, staged recovery, parent-lifetime scheduling, completion/drain
accounting and automatic root/GUI cleanup remain required C work. No background
progress, recovery or full Stage C acceptance follows from this slice checkpoint.


### C3 provider allowance and terminal-failure slice review

October 14, 2026. C3 is accepted for its scoped checkpoint as pre-submission
slice review; no formal Stage C correction round is consumed. After compaction,
the reviewer reloaded the entire `.agents/skills/ensemble-coding/SKILL.md` and
all four governing/craft documents before this review. Review covers the nine
Engine files, `compaction_provider_test.go` and the coder's complete C3 evidence.
Concurrent C4 selection changes and intentionally red legacy recovery probes are
excluded. This remains maintenance work, not a blind student attempt.

Compared with accepted C2 `ce52c166`, worker rendering gets the agreed Anthropic
16,384-token allowance and omits optional Chat/Gemini output caps. Ordinary
allowances remain unchanged. The three classic SSE readers now preserve terminal
reasons for their existing semantic decoders; a length stop rejects even valid
submit JSON before completed response facts or tool dispatch. Responses retains
its existing incomplete-response rejection. No parallel provider or worker loop
is added.

Measured failure usage stays in the worker Engine: classic semantic decoders
return their already normalized counts, while Responses extracts present input
and output measurements from failed terminal or malformed completed envelopes.
The failure paths account once and do not publish a successful response event;
the successful path retains its existing event-based accounting. Request-local
configuration supplies the accounting bucket. Parent usage and latest main input
remain untouched until the forthcoming managed import. This source review does
not establish that later import or persistence protocol.

Independent verification used an archive of exact `ce52c166` plus only the nine
C3 Engine diffs and the new provider test. Focused provider/malformed-response
checks and committed C2 construction/finite-execution controls pass under race
in 2.030s. The sorted owned-path-and-content SHA256 binding is
`a8dfa0985a8a4eac2da037272fb5c3bd03350b0bde9da13bcbce7066aefaccb4`.
Provider/runtime environment prefixes were stripped, no credentials were read,
and the temporary source archive was removed. The coder separately records the
20-case focused race at 1.919s, final combined C2/C3 race at 2.163s, inherited
streaming/visible-explanation/Engine race checks and targeted vet success.

All thirteen `/tmp/ensemble-compaction-stagec-provider-mutant-*.txt` outputs were
inspected. Restored caps fail actual request assertions; removed semantic checks
or discarded SSE stop reasons allow a length-limited submission and fail that
specific assertion; dropped accounting produces measured zero where exact usage
was expected. The ordinary-cap mutation fails its ordinary rendering control.
Malformed Responses accounting fails in both delivery modes through their shared
semantic path, not two independent proofs. The initial nonexistent History.Events
fixture error remains recorded; it is not runtime defect evidence. Unchanged
Responses terminal refusal and schema enforcement retain inherited evidence.

Across the nine production files, physical/active/comment-only counts move from
1967/1602/302 to 2037/1656/315, net 70/54/13. The new test is 250/237/7.
Counts exclude blank and comment-only lines from active code, keeping tests
separate. Added comments explain the output allowance, terminal reason and
failure-accounting boundary; exported fields have useful documentation. The
small changes extend existing provider owners and avoid new protocols. No KISS,
ownership or scope blocker remains for C3.

The roughly 52KB exact-memory positive control and synthetic 9000 output tokens
prove transport/validation preservation, not live tokenizer capacity or semantic
compression quality. An interrupted stream without final measurements remains
unreconstructed. Transient display deltas are distinct from executable completed
parts. Parent scheduling, source retention, staged recovery, drain/import and
automatic Fleet/root/GUI cleanup still require implementation and full Stage C
review; this checkpoint does not establish background compaction success.


### C4 selected-source snapshot slice review

October 14, 2026. C4 is accepted for its scoped checkpoint as pre-submission
assistance, with zero formal Stage C correction rounds. The full skill and four
required governing/craft documents remain active from the C3 reload, with no
intervening compaction. Review covers common memory vocabulary, existing Memory
selectors, the new selection implementation, range-fixture adjustment, selection
tests and the coder's C4 evidence. C3 is the predecessor `0a1bb906`; old recovery
probes and future actor/recovery work are excluded.

Memory owns snapshot derivation and revalidation through its existing Agent
parent. The digest binds selected entry identities and full parts, task text and
cut boundaries; effective source/destination settings and owner-relative file
fingerprints bind the relevant policy and corpus. Returned History context is an
owned snapshot, so retained entries do not alias later reducer changes. Source
files must match projected source content already at selection. Terminal MEMORY.md
is also checked against the input context before a worker can overwrite it.
Nonterminal destination collisions refuse selection rather than silently replacing
another file. Revalidation does not hash all history or unselected files, retaining
the intended freedom for independent dialogue, skills and later files to change.

The data stays a small extension of the existing Compression value, not a second
runtime owner or general snapshot framework. Selection is an immutable caller
contract; the actor must retain its own attempt identity and relevant generation.
In particular, settings changing away and back need that later generation check.
Validation is not an atomic filesystem transaction and does not replace the
forthcoming stage/finalize/recovery ordering or final conflict checks.

Independent verification archived exact `0a1bb906` and applied only the five C4
source/test paths, excluding untracked recovery probes. Selection, C1 range and
waterfall tests pass under race in 1.170s. Provider/runtime environment prefixes
were stripped and the scratch archive removed. Final reviewed five-path SHA256
binding is `a44c4493695786884458f810069202e12dcdb5559975d94eb0855123b9051725`.
The only post-run change is the requested Sources comment correction: the field
contains fingerprints, not the file contents themselves. No behavioral rerun is
needed for that wording. Coder's final focused race is 1.167s; targeted vet passes.
Its earlier misspelled fixture type is preserved as a setup failure.

The five selection mutation outputs were inspected. Ignoring selected entries
accepts redacted conversation; ignoring policy accepts a changed source budget;
removing collision protection selects an occupied destination; omitting curated
context/file agreement accepts stale input; removing both source-byte guards
accepts an edited selected file. The last is explicitly a coupled unchecked-path
mutation, not proof of two independent checks. Positive controls retain unchanged
selections and unselected-file edits. Actual Memory, History and corpus files are
exercised; no internal mocks or live-provider claims are introduced.

Across the three production paths, physical/active/comment-only counts move from
496/386/82 to 665/516/112, net 169/130/30. Across the two test paths they move
from 138/128/6 to 327/303/14, including the existing range test and its real-file
fixture adjustment. Active counts exclude blanks and comment-only lines. Comments
explain selected-only hashing, ownership, collision protection and the generation
limit; new exported vocabulary has useful documentation. Growth implements the
approved stale-source checks without expanding policy. No C4 scope, ownership or
KISS blocker remains.

This slice does not invoke ValidateSelection from the final actor acceptance path,
repair staged disk windows, import usage, schedule background work or remove child
views. Those remain required Stage C work, with full implementation acceptance
still pending. The known startup recall scan must also respect the accepted
recovery-before-scan ordering when that next slice is implemented.


### C5 correlated History and journal slice review

October 14, 2026. C5 is accepted for its scoped checkpoint as pre-submission
slice review, with zero formal Stage C correction rounds. Full mandatory skill
and governing/craft context remains loaded without an intervening compaction.
The predecessor is C4 `ffa0bb68`. Review covers the six production paths, new
receipt test, full C5 evidence and all eleven mutation outputs; legacy recovery
probes and unimplemented Memory staging/actor integration remain excluded.

History owns the existing held-cut projection and now correlates managed launch,
cut or retirement, population and outcome by task identity. Managed population
also checks destination file/band/range and requires its previously held cut.
A legacy disk population cannot satisfy that managed pair. Commit applies source
removal with replacement and retains the receipt instead of dropping it; another
launch, duplicate replacement or abandonment cannot erase committed finalization
work. A matching terminal outcome clears the pending receipt and retains only the
latest outcome. The new event does not change main-turn state or response usage.
Legacy uncorrelated behavior remains separate and supported.

The Checkpoint addition activates the journal while holding the same History lock
after the snapshot install and journal reset succeed. This closes the fresh
embedding-client path where Save preceded Load and subsequent launch/commit facts
otherwise had no active journal. Existing snapshot-before-journal-reset ordering
and process-crash rather than power-loss scope remain intact. Recovery later still
must reconcile staged files with the exact retained receipt before allowing corpus
scans or renders; receipt persistence alone does not repair either old disk window.

Independent verification archived exact `ffa0bb68` with only the seven C5 paths,
excluding old recovery probes. Receipt, C4 selection and waterfall race checks
pass in 1.173s. Sorted owned-path/content SHA256 binding:
`07758c5018362a1646ebf22dd32ade9b847d5ba1845272c98ab0b5b659ce9a8e`.
The temporary archive was removed and provider/runtime environment prefixes were
stripped. Coder's affected memory/context/save compatibility plus receipt race
passed 3.062s, final focused receipt race passed 1.042s, and targeted
history/ensemble vet passed. The final receipt log was inspected. The actual public
Save-before-Load test restores the journal without model work; a final snapshot
then preserves the terminal outcome. These are controlled persisted-state checks,
not forced process-kill experiments.

Eleven mutation logs under `/tmp/ensemble-compaction-stagec-receipt-mutant-*.txt`
were inspected. They fail task/file/range/band identity assertions, committed-state
protection, retained-receipt checks, disk-versus-managed correlation, deferred cuts,
terminal cleanup, outstanding-receipt protection against a legacy launch, and
fresh-Save journal recovery. No compile/setup failure is counted as a kill. The
real graduation control also verifies the sole unselected survivor is file3.
The tests exercise actual History, Agent saves and filesystem journals, without
mocking the behavior whose ordering matters.

Across six production paths, physical/active/comment-only lines move from
1548/1032/445 to 1697/1150/472, net 149/118/27. The new test is 173/159/8.
Active counts exclude blank and comment-only lines. New exported fields and event
vocabulary explain their correlation and lifetime; local comments identify the
recovery receipt and journal activation reasons. A narrow validation helper keeps
managed correlation out of the general reducer without introducing another state
machine owner or event framework. No C5 KISS, scope or ownership blocker remains.

Full Stage C still requires staged candidate storage, promotion/archive/restart
reconciliation, main-render and recall holds, actor lifetime/generation handling,
reliable completion/drain, usage import and removal of child views. C5 acceptance
neither validates those absent paths nor changes the known old restart failures.


### C6 staged filesystem recovery slice review

October 14, 2026. C6 is accepted for its scoped checkpoint as pre-submission
review, with zero formal Stage C correction rounds. The full mandatory skill and
four governing/craft documents remain active without an intervening compaction.
Review covers seven production paths and `compaction_recovery_test.go` against
C5 `c00af62b`, plus the complete C6 evidence and twenty-one mutation outputs.
The normal actor's legacy compression path remains outside this slice's claim.

Memory owns one hidden versioned candidate through its existing owner/Sandbox.
Stage validates the captured selection and exact launch, writes only hidden data,
and returns the proposed population. Finish requires the exact committed receipt,
selection and candidate hashes; it promotes recorded bytes and archives only the
selected source paths. Matching installed destinations and already-retired sources
allow restart to continue. Changed destination/source/archive bytes refuse further
automatic work and retain the record and receipt. Recover returns a terminal fact
for the actor to journal; only a subsequent call removes the hidden record. An
uncommitted managed attempt is abandoned without applying its held source cut.

Startup/settings synchronization resolves this protocol before active scans.
Committed unresolved receipts also block Memory selectors/scans, recall searches
and rendering of their supplied Context. Deferring initial recall indexing removes
the constructor scan that previously preceded an embedding client's Load. The
Engine check remains based on its supplied context, preserving offline rendering
ownership. The future actor must still serialize acceptance at its promised safe
boundary, keep attempts across ordinary interruption and prevent new launches
until terminal staging cleanup succeeds. These direct guards do not substitute
for that missing scheduling/lifetime integration.

Review identified one meaningful coverage gap: existing curated MEMORY.md promotion
was not exercised by the initially absent nonterminal destinations. The coder added
an actual staged commit followed by fresh Load/Ask that checks exact replacement
bytes, and an after-commit human edit that must retain both bytes and receipt.
Two targeted curated mutations now fail because old bytes remain or the human edit
is overwritten. This was test-only reviewer assistance; production needed no
correction. It is distinct from C4's earlier curated selection checks.

Independent verification archived exact `c00af62b` and applied only these eight
C6 source/test paths. All six focused recovery/staging/startup/curated test functions
pass under race in 1.262s. The actual permission-refusal startup control passes,
with no skip. Sorted owned-path/content SHA256 binding:
`54c8cb203e88cc18e49146543897ff0e804083db866dab04ce0199742efe7831`.
Provider/runtime environment prefixes were stripped, no credentials were read and
the temporary archive was removed. Coder's final focused race passes 1.258s;
its actual memory/recall/compaction compatibility race passes in 18.578s and targeted
vet passes. The internal package no-tests/no-selected-tests limits remain explicit
in the coder's note rather than being claimed as additional test coverage.

The twenty-one staging mutation logs were inspected. They detect active exposure
of a staged record, omitted startup recovery/archive, conflict overwrites,
non-idempotent partial archival, selection/candidate/receipt corruption, unsupported
record formats, bypassed scan/render/recall holds, eager startup indexing, premature
staging deletion, empty/stale candidates and both curated branches. The failed
scratch replacement assertion before the stale-ID mutation is retained as setup
failure, not a kill. Active-record visibility establishes that hidden bytes entered
Scan, not useful semantic memory. The skip-recovery mutation fails because the
remaining scan hold refuses the otherwise valid restart; its dependency is explicit.

Across seven production paths, physical/active/comment-only counts move from
1674/1329/265 to 1963/1575/297, net 289/246/32. The new recovery test is
436/409/18. Counts exclude blanks and comment-only lines from active code and keep
tests separate. Local comments explain ownership, process-crash atomicity, exact
receipt checks, partial archival and conflict handling; exported operations state
their caller-serialization and History boundary. The extra code implements the
approved one-record recovery protocol using existing owners, not a generalized
transaction framework. No C6 scope, ownership or KISS blocker remains.

These constructed saved-boundary tests use the new Stage/Finish API. The original
direct-Write probes remain preserved baseline failures, including their no-write
and completed-archive controls; these are not unchanged tests now passing. No
forced process kill, power-loss durability, atomic transaction against concurrent
external editors or repair of arbitrary legacy disk ambiguity is claimed. Normal
background scheduling, actual child completion/drain, once-only usage import and
Fleet/root/GUI dismissal remain required before full Stage C acceptance.


### C7 auxiliary accounting API slice review

October 14, 2026. C7 is accepted for its scoped checkpoint as pre-submission
slice review, with zero formal Stage C correction rounds. The required full skill
and four governing/craft documents remain active without an intervening compaction.
Review covers four production paths and the new accounting test against accepted
C6 `59c71d90`, the full coder evidence and seven mutation logs.

Engine exposes owned value copies of its existing vendor/model/price-tier buckets
and adds auxiliary buckets through the same locked totals helper as normal request
accounting. It preserves original request tiers instead of repricing aggregated
counts as one prompt or at the parent's model. Latest-main usage, History response
facts and main occupancy remain unchanged. The public alias makes these receipt
values constructible by outside consumers. No mutable shared receipt, extra usage
owner or task-ID ledger is added. Calling once remains explicitly the owner
attempt's responsibility after actual worker drain; this method is additive, not
idempotent by itself.

The real HTTP/runtime test obtains one main response and two child responses,
including a schema retry across the existing small/large pricing threshold. It
checks successful and measured-failure imports, exact bucket identity and counts,
unknown-price behavior, unchanged main usage and fresh-session Save/Load semantics.
The test shuts down and saves the actual child before manually importing its final
buckets. The expected cost is consistent with the existing model table; this is
not new live price verification. Returned buckets contain scalar values only,
which also directly establishes their absence of aliases without inventing pointer
state solely to mutate it. Review replaced the first-element child-copy assertion
with a slice-wide check so map iteration order cannot weaken that observation.

Independent verification used an exact `59c71d90` archive plus only the five C7
source/test paths. Accounting plus C3 successful/failed provider tests pass under
race in 1.989s. Sorted owned-path/content SHA256 binding:
`bd66de002b296bc75bdbbd94332884b60dd17eb51cd9acdde1b9a309103f6676`.
The temporary archive was removed and provider/runtime environment prefixes were
stripped. Coder's affected accounting/provider/cache/status race passed 2.057s
and targeted vet passed; its focused exact-bucket test passed 1.086s before the
final slice-wide copy assertion. The initial overly strict floating-point meter
comparison remains a documented fixture failure; tolerances now apply only to
derived money, with exact count and identity comparisons retained.

All seven mutation outputs were inspected. They detect omitted imports, lost tiers,
substituted model/vendor, overwritten latest-main usage, missing copied counts and
duplicate addition within one import call. The latter does not establish immunity
to repeated future mailbox delivery, and the coder's note explicitly distinguishes
those claims. No internal mocks, new model rows or live-provider calls are used.

Across the four production paths, physical/active/comment-only counts move from
1285/767/411 to 1335/795/428, net 50/28/17. The new test is 149/141/5.
Active counts exclude blank and comment-only lines. New public vocabulary and
methods document copies, drain timing, tier preservation and once-only ownership.
This is a small reuse of existing Engine accounting, with no scope, architecture
or KISS blocker. Full Stage C still needs managed launch/completion/dismissal,
owner-lifetime scheduling, generation/headroom policy and the actual exactly-once
actor import before background compaction is accepted.


### C8 managed lifetime slice: initial review findings

October 14, 2026. Review resumes against C7 `22763ac6` after reloading the
complete `.agents/skills/ensemble-coding/SKILL.md` and all four full governing/
craft documents. C8 is pre-submission slice assistance, not formal Stage C
acceptance or a reset of its zero formal correction rounds. Core lifetime paths
and real HTTP tests are being reviewed; GUI removal consumption and parent actor
integration remain later obligations.

The initial source review found that managed launch checks the requested fresh
directory before calling the trusted custom factory, then saves the returned
Agent without checking its actual paths. An isolated exact-HEAD archive plus
the six current C8 production files reproduces overwrite of a preexisting actual
factory-selected snapshot. The real callback only constructs through NewChildAgent;
it never loads or saves. StartCompaction performs the destructive initial Save.
The intended assertion fails in 0.030s under race; raw result is
`/tmp/ensemble-compaction-c8-review-factory-path.txt`. This is an observed post-return
runtime write, distinct from arbitrary trusted callback side effects that the
root cannot prevent. Coder and coordinator received the narrow finding before
acceptance.

Additional preliminary review asks for an actual successful-submission/manual-kill
interleaving: killing a provider-held request does not exercise racing automatic
cleanup after Result. Shared sync.Once source ordering looks appropriate, but
its distinct shared-error/cancellation claim still needs the intended boundary
observation. Correct-owner rejected-factory cleanup is also under assessment;
no reproduction or broad factory hardening claim is made for that source lead.

The independent overwrite probe is retained at
`/tmp/ensemble-compaction-c8-review-factory-path_test.go`, SHA256
`5fcb14fd65e631a599ef68f3fa1d0e0548bf9e7ffa723bb97252072990356587`,
for an unchanged rerun after correction. All 13 initial C8 mutation outputs were
inspected and fail their intended checks. The early-stop test establishes lost
provider admission, as the coder honestly records; it does not establish the
post-Result shared-save interleaving. The initial-save test's zero usage is only
no measured usage, not evidence by itself that no HTTP attempt occurred; source
ordering and its targeted Save omission establish the actual launch refusal.

Initial slice disposition: correction required before C8 checkpoint for the
reproduced managed-path overwrite, with the shared-cleanup race evidence still
pending. No additional runtime issue was established from the rejected-factory
cleanup lead: the ordinary-request/factory-changes-purpose case is inherited,
and managed constructor normalization prevents the suspected schema-changing
path through the real constructor. This lead does not expand C8 into generic
factory hardening. No production changes, commits, live calls or grader runs
were performed by the reviewer.


### C8 managed lifetime slice: final disposition

C8 is accepted for its scoped checkpoint after the preliminary corrections, with
zero formal Stage C correction rounds. The full mandatory skill and all four
governing/craft documents remain active from this review's reload. This review
covers seven production files and the new lifetime test against `22763ac6`; it
does not accept the unfinished parent actor integration or GUI consumption.

The reproduced overwrite is corrected at the actual construction boundary:
CreateChild compares each explicit requested managed data/log/save path against
the returned Agent Config before Fleet may save it. Provider configuration remains
customizable. Correctly owned rejected managed workers are drained and forgotten;
unrelated Agents are not shut down. Constructor-only managed shutdown skips empty
log persistence, preventing refusal itself from replacing an existing log with an
empty header. The trusted callback must return before Load, Save or task admission;
this protection does not claim to undo arbitrary callback effects before return.
The new per-path tests check exact old bytes as well as refusal and removal.

Managed launch retains the actual child and reliable Result. A single child-entry
sync.Once shares shutdown, job drain and final Save outcome between automatic
completion and explicit cancellation. The envelope owns its copied submission and
final usage buckets, retains task and cleanup errors separately, and records a
kill requested while cleanup is still active. Exact reserved map identity avoids
deriving removal from a custom display label. Root removal applies only to managed
children, retains files and publishes the surviving owner; a monotonic allocation
counter prevents reuse caused by a shrinking live registry. Ordinary spawning,
restoration and retained ordinary root references preserve their earlier behavior.
There is no secondary dismissed-child registry or duplicated execution loop.

The new real FIFO test establishes the previously missing interleaving. Submitted
idle plus actual child Shutdown puts the successful Result before the held final
snapshot. Explicit kill publishes its cancellation request before joining that
save. An external FIFO reader then releases the real filesystem operation. Both
success and actual journal-reset failure preserve successful Result plus Cancelled;
both callers receive the same final save error. This uses actual actors, History,
jobs and persistence with an external filesystem barrier, rather than an internal
mock or a repeated timing lottery. The six additional mutation outputs were
independently inspected: all fail intended assertions, including the bounded
second-caller liveness assertion for a duplicate Save. Together with the earlier
13, these are 19 targeted defect detections, not proof of every possible race.
Initial compile, peer-cleanup and initial-save fixture failures remain recorded.

Independent verification reconstructed exact HEAD plus only the eight owned
source/test paths. All eight C8 test functions and the unchanged independent
overwrite probe pass under race in 1.146s. Raw result:
`/tmp/ensemble-compaction-c8-review-final.txt`. Executed source binding:
`9d399bc32e290904980c7115efde5351471636df3d80e4194bfe8e09cf198cfc`.
The scratch archive was removed and provider/runtime environment prefixes stripped.
Coder's final affected Fleet/Named/Management/factory race passed 4.468s and vet
passed; its final exact-byte refinement passed 1.052s. A subsequent comment-only
clarification explains that constructor refusal omits Save and may hold Fleet's
lock, so removal observers must defer owner traversal. No runtime bytes changed
after the independent test. Final eight-path binding including that comment:
`55761342a65f2095753e758867d49564b87dfefa4dde408718f78545135b9df8`.

Across seven production paths, physical/active/comment-only counts move from
2672/1886/606 to 2839/2009/640, net 167/123/34. The new test is 515/482/20.
Active counts exclude blanks and comment-only lines; tests are separate. Growth
is explained by managed launch, shared lifetime completion and narrow root
removal, with useful exported API and ordering comments. No KISS or ownership
blocker remains in this slice. No live-provider or grader run was performed.

Full Stage C still requires the GUI's late-observation/removal handling, one-slot
parent scheduling and policy generation, headroom behavior, safe candidate commit
and the parent's exactly-once usage/outcome consumption before its final Save.
In particular, Fleet shutdown alone does not yet prove the parent consumed its
completion envelope. Those requirements remain active, with normal actor memory
scheduling still on the old path until the later integration replaces it.


### C9 GUI dismissal slice review

October 14, 2026. C9 is accepted for its scoped checkpoint against C8 `59d6fcfb`,
after one preliminary source correction and with zero formal Stage C correction
rounds. The full skill and four governing/craft documents remain active without
an intervening compaction. Review covers `agent/gui/gui.go`, `tree.go`, the new
`compaction_test.go` and the coder's complete C9 evidence. Full Stage C remains
unaccepted until the parent actor protocol is integrated and reviewed.

The synchronous observer handles managed removal before view creation, deletes
the actual view and only wakes the existing tree worker. It neither traverses the
owner Fleet nor loads settings while a rejecting factory may still hold Fleet's
lock. Current root membership rejects actual late observations without storing
tombstones. Deferred fallback follows the actual parent chain and shares ordinary
selection, owner settings and indivisible replay. The expected-selection check
preserves an already-applied newer user choice; ordinary selection retains slow-
receiver disconnection and the established settings/observer lock separation.

Initial source review found one admission race: membership was checked before
acquiring the GUI view lock, allowing removal to delete a view before selection
recreated it. The settings-error retry could then leave that new map reference.
This was a source-identified ordering defect, not an independently reproduced
interleaving. A membership recheck immediately before view construction, under
the same lock used for removal, closes it: earlier removal causes retry; later
removal waits and deletes the admitted view. The successful-load postcheck remains.
No test hook or parallel view registry was added to prove a scheduler ordering.

Real socket tests cover three successive workers with both root and nested owners,
owner-positive tree checks, exact owner tool-result replay and saved font settings,
view/selection reference release and replay of an actual captured late observation.
A real settings FIFO holds selection while the child completes; it establishes
that worker completion remains unblocked and the delayed settings-error path falls
back to the owner. It does not establish every successful-load or newer-selection
interleaving; those guards retain explicit source reasoning. The HTTP fake
recognizes sole submit with a memory property, while exact schema enforcement
retains C2's separate evidence. No actor, settings or GUI implementation is mocked.

Independent verification reconstructed exact C8 plus only the three C9 paths.
New compaction, existing selection and structured-submission replay tests pass
under race in 4.409s. The GUI command package reports no selected tests (1.024s),
and virtual-user has no tests; neither is claimed as extra GUI behavior coverage.
Raw result: `/tmp/ensemble-compaction-c9-review-final.txt`. Sorted owned-path/
content binding: `10881fb96e70b994dafa367aeae9f6f36eca3827404fdf3d01ada7fd73c177bd`.
Scratch source was removed and provider/runtime environment prefixes stripped.
Coder's full GUI suite passed 8.292s before the narrow admission correction; its
post-correction affected race passed 4.401s. GUI vet passed. No full-suite result
is relabeled as a run after that final correction.

All eight final mutation logs were inspected: deleted view retention, accepted
late observations, wrong actual owner, missing owner settings, omitted forced
fallback, settings-error retry, missing tree and missing fallback history each
fail intended observations. Tree has an owner-positive control; history omission
reaches fallback before its bounded missing-tool-result read fails. The earlier
broader replay mutation failed initial settings and is not credited as proof of
fallback history. The original C8 regression failed two explicit missing-selection
read deadlines (10.105s), and the initial late-observation fixture captured removal
itself before being corrected. These limits and corrections remain in coder notes.

Across two production files, physical/active/comment-only counts move from
603/507/76 to 663/561/81, net 60/54/5. The test is 314/301/5. Active counts exclude
blanks and comment-only lines; tests are separate. The larger helper mostly moves
existing settings/replay behavior into one shared path; comments explain removal
ordering, lock ownership and indivisible replay. No new core protocol, worker
service or persistent removed-ID set is introduced. There is no remaining slice
architecture or KISS blocker. No live model, browser demonstration or grader run
is claimed. Parent scheduling, safe commit, headroom and exactly-once accounting
through shutdown remain the next required Stage C work.


### Overall-review correction to C8 refusal coverage

Root's later overall review reproduces a missed reverse-purpose refusal defect
on exact `f111f1b3`. A managed StartCompaction request uses a real custom factory
that only constructs an ordinary child (`Compaction=false`) with an existing log
path. CreateChild refuses it, but Shutdown replaces the sentinel with an empty
log header and managed-only ForgetCompaction retains that rejected child. The
same callback retaining managed purpose is the passing control. Root's race
result is the intended two-assertion failure in 0.022s, retained at
`/tmp/ensemble-root-review-factory-refusal.txt`; the unchanged test is
`agent/ensemble/root_factory_refusal_test.go` under the scratch root recorded in
`/tmp/ensemble-root-review-path.txt`. No inference is started by this fixture.

The C8 scoped acceptance remains history. Its review considered ordinary-to-
managed rejection but failed to follow managed-to-ordinary refusal through log
persistence and root removal. This is a reviewer coverage gap, not evidence that
the later C10 actor work introduced the defect. The coordinator has requested a
narrow correction preserving normal lifecycle behavior, without a new sandbox
claim for arbitrary callback effects. Independent review of that correction and
the unchanged reproduction is pending alongside C10. Formal Stage C correction
rounds have not been reset or consumed by this pre-submission finding.

The coder's proposed narrow correction is accepted in principle: use an explicit
constructor-refusal drain through the existing actor and worker joins, suppress
persistence only on that path, and share private exact-child root removal behind
the managed-only public removal method. Verify actual ownership before discarding.
Normal ordinary Shutdown persistence and retained references remain unchanged.
This explicit lifetime reason should replace the earlier purpose-specific empty-
managed-log heuristic rather than adding another empty-history rule. The unchanged
root reproduction and a normal ordinary lifecycle positive control are the focused
verification boundary; actual corrected source review is still pending.

### C10 and complete Stage C implementation disposition

Independent runtime reviewer, October 14: **Stage C implementation accepted at
initial formal review, zero formal correction rounds**. C1–C9 were scoped
pre-submission reviews, not repeated resets of that count. This closes the runtime
implementation gate, including C10 and the root-found refusal correction. It does
not claim the later grader adaptation, chapter/export mapping, live quality or
deployment gates complete. The entire exact mandatory skill
[`.agents/skills/ensemble-coding/SKILL.md`](../../.agents/skills/ensemble-coding/SKILL.md)
and all four governing/craft documents were reloaded in full after compaction;
truncated read ranges were recovered before this review resumed. Prior complete
Chapter 14/15 readings and the accepted design remain governing.

The final C10 actor holds one attempt independently of the initiating turn. Its
counted completion receiver survives shutdown admission closure. Only the first
completion for the exact active attempt decrements the receiver count and imports
final copied usage. Acceptance follows reliable Result, shared child drain/save and
removal; usage is imported before success/failure disposition. Completion during a
sent main response waits for idle or the next complete-batch render boundary.
Ordinary interruption detaches the waiting turn without killing the worker, while
explicit cancellation, owner shutdown and relevant policy generation changes
invalidate uncommitted candidates. No task-ID set, second scheduler or callback bag
was added. Successful retirement alone permits the next waterfall selection.

The 95% hold uses the latest main input's disjoint input/cache buckets, independently
of imported child totals. A failed held attempt ends that waiting turn explicitly.
Committed finalization errors retain the receipt and staging bytes, hold rendering
and re-enable, and can complete after repair. Terminal outcome/cleanup failure
retains the attempt instead of launching over it. Shutdown consumes the completion
and reports unrepairable persistence errors rather than waiting indefinitely for
filesystem repair. Historical outcome usage is audit data and is not reimported
into a fresh Engine. The old synchronous Compress API remains callable for existing
clients, with deprecation documentation; the normal actor never falls back to it.

Root's reverse-purpose factory finding is corrected by explicit construction
discard, using the existing counted actor shutdown and suppressing persistence
only on that refusal path. Ownership is checked before discarding; private exact
root removal handles both returned purposes. Normal ordinary Shutdown still writes
its log and retains its root entry. The previous empty-managed-history heuristic
is removed. Root independently reran its original assertions unchanged apart from
the test-function/comment rename needed to avoid the ported duplicate: race1.030s,
`/tmp/ensemble-root-review-factory-refusal-fixed.txt`; the duplicate-symbol setup
failure is retained separately. The independent run below also exercises the
ported root assertions and ordinary lifecycle control. This does not claim to
undo arbitrary trusted factory effects before returning a constructed Agent.

The coder then identified the associated GUI omission: removal of a rejected
ordinary child was outside guards checking only Compaction purpose. Independent
source review agreed. The six existing checks now apply to parented children
missing from the root; valid ordinary children remain registered. The actual
refusal/socket regression passes the retained pointer to the shared selection
continuation, verifies owner selection and absence of retained view references.
It models delayed pointer consumption without claiming a deterministic in-flight
socket interleaving. Restoring the old purpose guards selects main/rejected and
fails that exact assertion in0.024s. Both initial omitted-subscribe fixture failures
remain setup failures. Exported removal comments now include refused construction.

Independent verification reconstructed committed base
`7552f4276f91138b213dcf4a5c22912ac29cb676` plus exactly14 final code/test overlays.
Core Compaction, Memory, StructuredSubmission, root refusal and ordinary lifecycle
checks pass under race in5.324s. GUI compaction/refusal, inherited selection and
structured replay checks pass under race in4.434s; the GUI command package reports
no selected tests and is not extra behavior coverage. Source was unchanged after
execution. Binding: `387b87f643844024bf1c41722be33905da8a51af38a2cd6e39d406694ba72305`,
recorded in `/tmp/ensemble-compaction-c10-review-source.json`; raw output is
`/tmp/ensemble-compaction-c10-review-final.txt`. Scratch was removed and credential
prefixes stripped from test subprocesses. An earlier independent launch refused
the stale manifest before any test because the authorized removal-comment edit
changed fleet.go; it is not a behavioral failure or a test pass.

All15 final C10 mutation outputs were inspected and fail their intended assertions:
usage import, safe-boundary commit, policy generation, child cleanup refusal,
headroom wait, failed-wait refusal, disjoint cache arithmetic, terminal usage,
finalization, settings hold, construction persistence suppression, exact root
removal, interrupt preservation, counted shutdown completion, and GUI refusal
fallback. The explicit blocked-request and missing-terminal-outcome deadlines have
passing controls and establish their named liveness failures; they are not hung
builds. Exact duplicate mailbox admission remains source-reviewed because the
runtime has one completion producer. No artificial runtime hook or broader
scheduler proof is claimed. Earlier fixture mistakes and coupled slice mutations
remain recorded in the coder's report and this audit.

Coder's final core focused race passed5.333s and both vets passed. Its full core
run passed Ensemble92.956s and the other packages except the independently
baseline-reproduced CLI human-sink true case. That failure remains a failure,
not a green suite. The full run predates the final detached-waiter reference clear;
the focused race includes it. Final full GUI passes8.367s and vet after the
ordinary-refusal membership correction. All original grader and later live work
remain coordinator-owned, with their earlier source bindings preserved.

Complete Stage C review traces each added mechanism to the approved behavior:
C1 proves selection retirement makes progress; C2 restricts actual construction,
capability and finite task execution; C3 gives provider allowance and rejects
length/incomplete results while retaining measured usage; C4 captures exactly
selected source/policy/destination; C5 correlates the held history pair; C6 stages
and recovers corpus changes before scans/recall/render; C7 preserves owner totals
without changing main occupancy; C8 shares reliable lifetime/save/removal; C9
releases GUI references and restores owner replay; C10 integrates those owners.
The range and saved-state window failures were reproduced on the immutable
baseline; the staged tests are controlled saved-state/restart probes, not forced
process kills. C8's missed reverse-purpose refusal remains an explicit review gap,
now repaired, rather than being reassigned to C10 or hidden by later acceptance.

For comparison, the frozen first-edition compressor and compact controller
(`edition-1/agent/internal/llm/{compress,compact}.go`, reread in full) directly
perform isolated HTTP calls and then write memory before separate cut/population
facts. Their useful teaching about first-person memory, reasons, irreproducible
output and later-checkpoint retry is retained. The current pre-C baseline likewise
used a turn-owned direct compressor and write/event/archive sequence; it could not
meet the newly authorized independent lifetime or identified restart windows.
Neither historical implementation supplies an equivalent managed-background
chapter delta, so its smaller synchronous feature is not presented as a numeric
like-for-like win. Frozen sources are unchanged.

Against Stage B checkpoint a0acfb5e, Stage C changes36 production Go paths:
physical/active/comment-only totals for those paths move7839/6095/1338 to
9147/7150/1535, net1308/1055/197. Active excludes blank and //-only lines; counts
include moved code within the changed set. Thirteen test paths separately move
690/650/25 to3966/3742/133, net3276/3092/108. Detailed inputs are retained in
`/tmp/ensemble-compaction-stagec-review-counts.json`. C10's final ten production
paths alone add223 physical/191 active/22 comments versus their own predecessor.
The growth buys the required finite child, bounded receipt/staging state,
selection safety, accounting and lifetime reuse. Tests are larger because they
exercise real provider/filesystem/socket boundaries and preserved failure states;
there are no mocks or test-only production hooks. New comments explain ownership,
shutdown counting, safe commit, refusal persistence and recovery ordering; exported
contracts remain documented. The comment balance is near the teaching target,
without padding. No unresolved scope, architecture or KISS blocker remains.


## Stage D plan review, October 14

Independent reviewer accepts the bounded author and grader plans for execution,
not the unwritten chapters or unimplemented grader. Runtime remains7feb910a,
accepted at09e68149; Stage C closed at1d211806. This is plan assistance before
Stage D submission, with no formal correction round consumed. The exact mandatory
`.agents/skills/ensemble-coding/SKILL.md` and all four complete governing/craft
documents were reloaded after compaction before this disposition; truncated reads
were recovered. Complete affected current Chapters14/15/16/22/23 were read.

Reviewed plan SHA-256 values:

- `codex-compaction-author.md`: `07a7066526673fde3004f7c9ae065166923487d493a5dea0fa5fd14369ff7f82`.
- `codex-compaction-grader-plan.md`: `ae5c0f15457c11aebad5b0555c1f2fc92ae959ca13a57edeebeedf9fa3404e82`.

The author plan preserves the family-assistant memory incident, checkpoint gap,
watermark deadlock, decorative-disk and GUI-settings failures, provenance argument
and five historical live outcomes. It keeps recall's filing-cabinet explanation,
quota/judge and capped-snippet lessons, and fleet's wiring, steering and truthful
status stories. Changed explanations follow the accepted runtime: one owner-lived
restricted child, exact selected source and policy, safe commitment, staged disk
recovery, reliable completion/save before dismissal and once-only auxiliary usage.
These replace the old synchronous compressor explicitly. Historical measurements
and lost-reason/incorrect-recall results do not become background-worker evidence.
Current14 remains the invariant reference; its pointer changes at Stage E.

Coordinator review identified a stale role rule that the initial plans had not
explicitly corrected: current22 requires a factory-enabled root to submit, while
accepted `internal/fleet/fleet.go` Idle returns immediately when ParentAgent is nil.
I independently checked that source and agree with the bounded correction. The
coordinator's attributed author-plan amendment and coder's revised plan now
distinguish main/named-peer chat completion from delegated submission and bounded
ordinary correction. Examples and primary instructions must follow that distinction;
historical root-submit transcripts retain their original provenance. Explicit root
submit admission checks remain valid. Managed workers' finite Send budget does not
replace ordinary delegates' three corrective wakes. Actual prose remains to review.

The grader plan preserves all eleven memory IDs/weights, totaling100, verified
against the original table and current wrapper's ch14-parity rename. It retains
current22's seven GUI plus seven fleet rows/200, current23's seven rows/100 and
current16's ten rows/100. New background, structured-result and accounting predicates
are assigned to relevant existing rows without dropping their prior valid checks.
Numerical credit must come from the submitted executable and its public results,
wire bytes or persisted facts. Precise staged-window, constructor-refusal, defensive
copy, model/tier, save-race and reference-release tests remain separately identified
B/C implementation evidence; their passing results cannot award points to another
submission. Missing public observations fail or report an unexercised predicate.
This resolves the coordinator's concern about the initial replay-row attribution.

Source inspection confirms the proposed harness repair is necessary and small:
current15 classifies sole submit without its schema and waits inside the router
lock, while fakevendor serializes Route under its capture mutex. Current23 already
holds selected requests in an external reverse proxy. Reusing that boundary permits
main progress and capture reads during a held worker without changing the shared
fake. Main turn_ended and HTTP ordinals cannot acknowledge background commitment.
The revised plan instead follows correlated owner outcomes, waits for the intended
outcome before EOF, and separates worker submission, saved completion, corpus state
and GUI removal. Exact-schema comparison and an ordinary submit(result) negative
control prevent a circular classifier from certifying itself.

Scope and KISS are appropriate: reuse existing drivers, external peers and public
observations; no runtime hooks, alternate protocol, fake History/Memory or new
verification framework. New checks still need targeted defect sensitivity, with
actual failed predicates and coupled limitations retained. Explicit-source grader
commands must exercise the changed contract; old frozen scores remain historical,
and old synchronous source is only a negative control for the new lifetime rule.
The whole-TL;DR/graded-section ownership split is clear, with meaningful shared
example wording exchanged before edits. Grader coder must complete its full-four
authorship reload before changing chapter text. No tests, graders, live calls,
chapter edits, runtime edits, exports or publication were performed by this plan
review. Stage D implementation and prose review, Stage E reconstruction/mapping
and Stage F bounded live evidence remain required.


## Stage D narrative preparation review, October 14

Reviewed author-only checkpoint e4b7480c against its predecessor and accepted
runtime7feb910a. This is an intermediate review of the actual narrative changes
to current15/16/22/23, not combined Stage D acceptance or a separate D1 gate.
The coder still owns final TL;DRs, graded sections and active grader work; those
intentionally unchanged old contracts are pending reconciliation, not author
failures. No formal correction round is consumed. The full skill and four
governing/craft documents loaded for the preceding plan review remain in active
context without intervening compaction; their subsequent workflow amendment was
read. Complete affected chapters were read for that review, and the full narrative
diff and relevant surrounding passages were checked again here.

The diff preserves the original openings, memory and recall incidents, checkpoint
gap, five-versus-eight failure, settings lock cycle, delegation wiring failures
and successful-work/bad-status story. Historical tables and five memory trials
remain attributed to the synchronous source. New prose explains why independent
lifetime needs dedicated role instructions, exact direct submit data, selected
source validation, staged recovery and completion before disposal. It retains
main Last versus auxiliary totals and cache-lens separation, and distinguishes
ordinary delegated wakes from the worker's finite Send budget. Added explanation
is proportional to those changed mechanisms; it does not turn the chapters into
review ledgers or rewrite unaffected teaching. Soft length/person warnings alone
are not a reason to remove useful explanation. The author's recorded lint/parser
results retain author provenance; I did not rerun them or execute graders.

One substantive clarification was returned to author and coder: current23.5 says
“Select the child during work and after restoration,” which can imply automatic
ephemeral-child restoration despite23.2/23.7 saying those children are not revived.
Distinguish same-process live child selection/reconnect from replay of retained
structured results in a restored owner's conversation, or describe explicit public
construction/load when that is the exercised boundary. The existing GUI structured
regression constructs a schema-bearing root Agent and explicitly loads it; it does
not automatically restore a task child. My initial coordinator message incorrectly
called that test an explicit child/factory reconstruction before opening the whole
file; the subsequent source read corrected that description immediately. The
wording finding stands independently. Coder acknowledges the distinction for its
pending22/23 grader. Author correction and actual final examples remain to check.

The coordinator's existing Stage F checklist is adequate preparation for the
approved single isolated live compression. It covers actual Astra/high/route and
visible explanation, source facts and reasons/uncertainty, structured submission,
committed bytes, main progress, child dismissal and selected-owner fallback,
no-tools recall and zero-inference restart. At review request it now explicitly
retains effective memory budgets/watermarks and UTF-8 source/submitted/committed
byte sizes, and measures main Last across auxiliary-only completion separately
from the subsequent recall turn. Isolation, binary/path/source binding, synthetic
public data, no private reasoning/credential capture, local launcher controls,
bounded observation and cleanup are concrete. A worker that finishes too quickly
leaves overlap or selected fallback unobserved; the checklist correctly retains
that limit rather than promising another paid run or relabeling deterministic
evidence. This is checklist review, not live execution or proof of memory quality.

Combined Stage D disposition waits for final contracts, public examples, revised
checks and targeted sensitivity/compatibility results, followed by independent
review. The coder retains the grader execution slot. I changed only this audit
note and made no chapter, runtime, grader, export, live or deployment changes.

## Stage D grader preparation and local launch review, October 14

The author corrected the sole interim narrative finding at8057f854:23.5 now
distinguishes reconnect/reselection of a registered child in the same application
from retained structured results replayed in the restored owner's conversation.
It expressly does not promise automatic recreation of ephemeral children. I read
the exact correction; that finding is closed. Final coder-owned contracts and
verified examples remain pending, so this is still pre-submission assistance,
not combined Stage D acceptance or a formal correction round.

Reviewed the coherent D1–D5 source at1acfd47e/2b17971e/5f7bc621/5391d5eb/8430d4b6,
the coder's explanations and retained outputs. After context compaction I reloaded
the entire `.agents/skills/ensemble-coding/SKILL.md` and all four complete governing
and craft documents before further review. D6 memory work was explicitly excluded
while changing. The coder retains the grader execution slot; I ran no graders or
tests during this preparation.

The memory driver holds an actual provider response outside fakevendor's capture
lock, allowing the real main actor and request recorder to progress. It joins
complete public launch/outcome records by task identity and crosses an existing
actor boundary before certifying that a cascade has drained. It does not substitute
main turn completion, elapsed time or HTTP ordinals for memory completion. The new
background scenario checks actual simultaneous main progress, exact sole memory
schema, primary-role exclusion, forged-tool refusal with a forbidden-file control,
selected-child dismissal and the next request's retained/new memory. Structured
fleet checks compare whole nested JSON values, arrays, false and null at actual
status/join/retrieval/display boundaries, including durable schema/owner binding.
Restarted owner-history replay is separate from explicitly respawning a new pending
task that retrieves an older result; that retrieval must not complete the new task.
Root chat and ordinary delegated corrective wakes remain distinct controls.

I found one scoring defect in current23: its GUI build-failure branch omitted
row6 even though the new root-chat predicate belongs to that row. Earlier core
fleetIdle success could therefore award10 points for an unexecuted predicate.
Coder correction ce43d75e adds row6 to the failure attribution. Its real-core,
deliberately unbuildable-GUI regression passes under race2.068s; restoring the
omission fails the exact assertion with Earned10,0.670s. I inspected both complete
raw outputs and the source/test diff. This closes the finding without adding a
row or changing successful-build behavior; no fresh full score is inferred.

Inspected all retained D1–D5 mutant outputs and the relevant passing controls.
The intended liveness/schema/result/ownership assertions fail as reported. Preserve
the D4 earlier grandchild-wire mutation's constructor refusal instead of claiming
it tested the declaration; the narrower declaration mutation does. Preserve D5's
unused-variable setup failure separately from its corrected fallback failure.
Invalid early submission and generic GUI encoding cause coupled downstream
failures, not independent proof for every affected row. Actual earlier full22=200
and23=100 retain their respective source versions. Public numeric grading remains
separate from B/C implementation-specific recovery, pointer-release and accounting
evidence. No new framework or runtime test hook appears in these slices; the
additional driver code serves the necessary observable concurrency and data checks.
Whole-D size/coverage comparison and final disposition wait for the complete work.

Read the entire canonical `scripts/edition2/background-memory-live.mjs`, the
coordinator's Stage F checklist and revised synthetic receipt at scratch suffix
1nKWZM (launcher SHA f79114560a7f84560e4481b94a315fa9294514c69d39ea7a05bd671b52ba9695).
I also viewed the retained tree-only screenshot. The receipt's owner sequences11,
14,16,19 establish launch, second human admission, main response and later outcome.
It retains2034 source bytes and847 submitted/committed bytes, exact schema and public
explanation before submit, selected child then owner fallback, auxiliary delta
20/11/17 with latest-main60/41/7 unchanged, exact memory in recall and two clean
exits with zero restart inference. These are local fake-provider observations,
not live inference, semantic-quality proof or a measured invoice. The script's
allowlisted output excludes auth headers and opaque/private reasoning; raw logs
stay in private scratch. The initial early-title capture and later actual event-name
correction retain their distinct provenance in the coordinator's record.

I also found a source-binding gap in the launcher:
pre/post-build checks used `--untracked-files=no`, so an untracked Go source or
embedded asset could affect the binary without appearing in the committed Agent
tree identity. Root changed both checks to `--untracked-files=all`. I inspected
the exact correction and `/tmp/ensemble-live-source-binding-negative.txt`: the
harmless untracked source fixture is refused before building at the intended
assertion. The fixture was removed and the Agent tree is clean. That source
finding is closed. The final corrected positive rehearsal C8xXXZ is complete;
its inspected receipt records the same committed Agent tree, seven verifier
refusals, main response before outcome, clean exits and zero restart requests,
with launcher SHA1c7986c22a62ca9011ef28ca83ab984e23bb4ef40b9a2347e49d0dfee2aaad3b.
The earlier capture retains its own provenance. No live run was requested. Launch readiness retains the
existing sequence gates; this review does not authorize deployment or publication.

At the coordinator's explicit request I additionally inspected the new20 speech
collision plan in `/Users/bill/projects/ensemble-reordered`, HEAD53958bcc plus its
uncommitted candidate. Name-only fleet admission hides the real MCP `submit`, and
name-only schema substitution would overwrite its contract. Current construction
does install compiled FleetTools when a factory exists even with BuiltinTools=false.
However, a documented pre-start SetAgentFactory after constructing a custom/MCP
Agent still defeats a global FleetEnabled test of tool identity. I accepted the
narrow design: conditionally install FleetTools in composition, bind the immutable
schema to that actual built-in definition, uniformly copy declarations, and remove
the fleet name-only gate. Preserve SafeMode and compiled-tool shadow precedence.
No origin API or wider tool framework is needed. Implementation and actual
MCP/custom/built-in controls remain the reconstruction reviewer's gate; this is
source-only design assistance, not acceptance of new20 or a runtime edit here.

### Scoped new20 tool-identity correction disposition

The reconstruction coder subsequently implemented that narrow correction in the
three existing production files and appended TestSubmissionToolIdentity. I read
the changed regions and entire six-case regression. It exercises actual MCP
discovery/invocation, custom registration, compiled precedence, late factory
installation and absence/refusal of unregistered fleet tools. The unchanged
browser speech test's original missing-submit failure, accepted predecessor pass
and corrected pass are retained under `/tmp/ensemble-reordered-checks/`; I inspected
their full outputs. Corrected browser duration is3822.490ms in the raw Node output;
the focused real-MCP race check passes1.205s. The two restored defects fail intended
assertions: name-only schema substitution changes question to memory, including
late-factory/custom cases; name-only admission removes MCP submit or refuses custom
registration. Neither is a compile failure or unrelated timeout.

This scoped correction is accepted. Reviewed SHA-256 prefixes: ensemble/ensemble.go
933e1342; internal/agent/agent.go9e306096; internal/tools/tools.go3c66f33f;
ensemble/submission_test.go3b2b60a0, all in the reordered candidate above. It removes
name-based identity guessing without adding metadata, an origin API or a framework,
and preserves SafeMode and compiled-handler precedence. Full reconstruction review
and any primary-source carry-forward remain separate. No test was run by this
reviewer; the inspection and acceptance use actual coder source/evidence.

One final Stage F evidence clarification was sent to the coordinator: the launch
receipt retains configured conversation budget600 but omits its default high/low
watermarks. Those effective values are1200/600 by memory.Settings, rather than an
observed effective-settings snapshot. Label that derivation or explicitly configure
and retain the values before live use. This requires no runtime change or paid
repeat and does not change the verified rehearsal results.

## Stage D stale/failure grader preparation, October 14

Reviewed exact D6 checkpoint223d6b78 and D7 checkpoint8eb924fa, their complete
coder explanations and all retained positive, initial-fixture and mutation outputs.
The full skill and four governing/craft documents remain active from the preceding
reload, without intervening compaction. Active D8 accounting edits are excluded
from this disposition. No test, grader, application edit or live invocation was
performed by this reviewer; this continues pre-submission assistance without an
extra slice gate or a formal correction round.

D6 establishes which oldest two files were selected in the real provider request
before changing either their bytes or destination policy. It retains a third-file
control, returns policy through acknowledged off/on settings, and checks the exact
captured task's reasoned invalidation, all three surviving files byte-for-byte,
empty active destination and source-bearing next main request. The shared public
snapshot/journal reader retains its journal-first ordering and complete-line rule.
No private staged-file format or internal hook awards points. The initial low200
fixture legitimately selected all three202-byte files; corrected low225 passes.
That fixture error is retained, not described as a runtime defect.

D7 supplies whitespace, wrong-type and syntactically complete length-stopped
responses through the actual child/provider path. The wrong-type case must put a
real validation error in a subsequent child request; an eventual owner parse error
cannot satisfy it. The length case changes the provider's terminal fact while
preserving valid arguments and measured usage. Exactly one failed correlated
attempt, bounded sends and retained source ends in the next actual request are
checked independently of successful process exit. The local peer has no auth
challenges, so its observed HTTP count is not a general transport-retry bound.

The full memory predicate race runs pass3.737s and4.068s respectively; neither is
relabeled a full advertised Chapter15 score including parity. I inspected all
eight intended mutation failures. D6's bypassed admission is still stopped by
Stage and fails the promised invalidated classification; it does not prove removal
of every file guard. D7's empty-memory mutant removes four enforcement sites and
is one coupled policy claim. The other failures reach the intended stale commit,
active candidate, missing source, fourth request, missing correction and truncated
commit assertions. Source and evidence support these narrow grader additions;
there is no correction requested. Whole Stage D acceptance remains pending the
remaining predicates, final public contracts/examples and advertised checks.

### D8 accounting grader disposition

Reviewed complete frozen checkpoint4ef71dbc, its note, inherited status/invoice
comparators and every retained D8 output. The scoped grader addition is accepted;
full Stage D remains pending. A held worker completes after the main response,
then the existing public selection replay exposes the owner's current meter
before another main response could overwrite Last. Actual agent_status calls
separately establish four disjoint usage counts in owner totals once for successful
and failed workers, followed by only each subsequent main request's new counts.
A fresh restored process makes actual status requests and must not rebill retained
historical outcome usage or launch another worker. This is submitted-executable
evidence; model/tier attribution, cache-lens exclusion and defensive copies remain
separate B/C source evidence. No runtime change or numerical weight change occurs.

The original repeated-subscribe observation timed out because that operation does
not request fresh replay; it is retained as a fixture error, not an accounting
failure. The corrected public select_agent control passes2.041s; the full affected
memory/accounting race run passes4.441s on Agent tree4871633e, with vet reported
successful. This does not replace D7's earlier source identity or claim a full
advertised Chapter15 score. Complete mutant logs confirm omitted failed-worker
counts, duplicate import and historical rebilling fail the corresponding actual
totals. Overwriting Last fails specifically the first post-drain snapshot while
later status calls recover it, supporting the new observation's purpose. None of
these four outputs contains a compile/setup failure. No additional test sweep or
implementation edit is requested. I ran no tests or graders for this review.

## New21 reconstruction plan review, October 14

Reviewed the complete plan at reordered checkpoint
7a4ba6fca51cfd7ff82de8d400f8d22dc8e4b1dd against accepted new20 b4708839,
Agent tree39449051049e6d7325b4c368e627739f97eab28f. Before review I reloaded the
entire `.agents/skills/ensemble-coding/SKILL.md` and all four governing/craft
documents in the reordered tree, recovering truncated ranges. I also read current
Chapter23 completely and inspected the actual historical management, actor,
registry, GUI and test changes selected from e00d5bc through450f7a9, the remaining
Stage B management integration and df38abe0's deferred AfterTool branch/test.
This is maintenance plan review with historical-answer access, not a cold student
assessment or an implementation result.

The plan is acceptable with no blocking finding. The historical delta supplies
the existing eight management tools, named peers, activity and selectable views;
it does not import a cumulative later runtime. Replacing private findChild with
the new owned Find retains the same direct-child boundary before disk fallback.
Typed get_submission, status, change waits and output snapshots preserve raw
payloads and caller-owned copies, while existing string APIs retain their
documented behavior. The plan explicitly preserves composition-bound compiled
submit identity, ordinary root/named-peer chat completion and delegated correction.
The already-present typed dispatch completion and turn_ended validation must
remain intact instead of being overwritten by historical string-only code.

The deferred AfterTool branch belongs here because this stage introduces the
actual interruption flag and tool-report handling. Its rejection regression uses
the real actor/provider path and checks that a rejected unexplained batch does
not trigger another request after the stop. GUI view conversion keeps settings
acknowledgments outside the observer lock, tree traversal in its existing worker,
and indivisible nonblocking replay admission. The planned held-work/no-progress,
selected-question, real slow-socket, typed-output and named-peer chat controls
address the consequential integration boundaries without a new framework. Actual
results and intended mutation failures remain to be reviewed; historical results
do not establish that the reconstructed binary works.

No additional requirement or test matrix is requested. Memory, recall, managed
compaction cleanup/recovery and auxiliary accounting remain at their approved
later introductions. The coordinator retains the decision to release Agent edits;
this review recommends the plan, not implementation acceptance. I ran no tests,
graders or live calls and changed no application, chapter or grader source.
Primary Stage D checkpoints D9–D11 throughf393c893 and unfinished chapter contracts
are outside this disposition; combined Stage D review remains pending.


## Combined Stage D final disposition, October 14

Accepted at initial combined formal review, zero correction rounds. This closes
prose/grader Stage D, not reconstructed exports, renumbering or the later live
assessment. Reviewed submission7ccdf847 and the separately accepted handoff
correction1a54cf25 plus chapter15 reconciliation2c6563dd. I reloaded the entire
`.agents/skills/ensemble-coding/SKILL.md`, workflow, carryover, voice and chapter
procedure after compaction. Earlier complete narrative and D1–D8 reviews above
remain evidence; this final review covers the remaining changed contracts,
D9–D11 executable predicates and consolidated results without repeating unchanged
checks. No application, chapter or grader was edited by this reviewer.

D9 holds an actual worker response outside fakevendor's capture lock. It verifies
main waiting at the known-window threshold, live settings application, ordinary
interrupt survival and later resumption from the same attempt. Its failed-worker
case requires an explicit turn error before another main request, then retained
source on an explicit new prompt. D10 distinguishes explicit child kill, owner
EOF and a real candidate-write permission refusal. Correlated outcomes, transport
cancellation, bounded shutdown, actual owner-tree dismissal and source-bearing
next requests establish those lifetimes; the shutdown case restarts the executable.
The permission case actually ran rather than silently skipping. Neither fixture
claims the exact committed-finalization crash windows established separately in C.

D11 prepares its corpus by real successful or failed graduation, with an unchanged
third-file control. The subsequent recall requests must include retired originals,
exclude still-active/replacement duplicates and retrieve a separate positive-control
document. Removing the recalled files before another process restarts establishes
recorded attachment replay without another judge call. Its ten-row100-point rubric and
memory's eleven original weights/100 remain intact. Supporting C recovery,
model/tier, lens and marker source tests are explicitly non-scoring evidence,
never numerical credit awarded to an arbitrary submitted implementation.

I read every retained D9–D11 positive, fixture-failure and mutation output. The five
headroom, four lifetime and three recall defects fail intended assertions. Initial
mixed-settings and early duplicate-band observations retain their narrower causes;
related duplicate-band assertions are coupled consequences of one mutation. The
bounded timeout failures concern the exercised waiting/cancellation properties,
with passing controls. Together with earlier reviewed sensitivity, this is adequate
focused evidence; no extra framework, exhaustive matrix or repeated sweep is needed.

The final contracts agree with accepted B/C ownership, immutable typed submission,
default string compatibility, delegated-only corrective wakes, finite restricted
workers, staged acceptance/recovery and once-only auxiliary accounting. Main Last
and cache comparisons remain distinct from owner totals. The retained narrative
still explains why these mechanisms matter and preserves the original stories;
23.5's registered-child versus restored-owner-history clarification remains correct.
The later handoff correction removes opaque parts and signatures on retained
visible parts at the completed-batch boundary, without rewriting the event log.
It is separately reviewed and is not attributed to the older D12 binary.

All390 file hashes in `/tmp/ensemble-compaction-d-final-source.json` match7ccdf847;
I independently recomputed sorted path/space/hash/newline binding
051cd2b4c900283892503f063f0873038d703ba571359f501ad2e26e1eccd72c.
Its pre-commit HEADc80755f7 is not confused with the submitted file contents.
Application tree4871633ee3e63dc6aa2cbc0a853f6ef22b4ddfa3 matches. Complete actual
sequential advertised results are14=100/100,15=100/100,16=100/100,22=200/200,
23=100/100, including15 parity and16's actual judge timeout. Raw reports, timing
records and package test outputs were inspected: gradeedition2 passes37.608s and
shared grade0.407s; vet is recorded successful. Expected cancellation diagnostics
before15's JSON do not invalidate its decoded report. Root's later primary
handoff race1.260s and15=100/100 retain their separate post-correction provenance.

I verified the four contract hashes and all46 final fences, including the indented
memory schema. Only15's Go type sketch and newly inserted JSON schema differ;
remaining fences match their immediate predecessors. All JSON examples decode;
the coder records successful standalone type parsing/compilation. My first local
fence scan omitted the indented example; the corrected scan confirms the recorded
13/9/18/6 counts. These syntax/identity checks are not new live executions or
validation of exports awaiting Stage E. Hard prose lint passes, with its stated
soft length/person warnings preserved rather than prompting an unrelated rewrite.

The21 changed grader-path counts independently match the coder's comparison to
ac7c67b3: production physical/active/comment-only2829/2493/234 becomes
4725/4220/369; `_test.go`155/130/12 becomes440/376/38. Active means nonblank and
not `//`-only. Most growth is concrete public-process fixtures within the grader,
not application machinery. Reusing existing launches, settings, correlation and
HTTP peers keeps this substantially clearer than introducing another runner or
private runtime protocol. Comments explain ordering and evidentiary limits;
their below-target proportion is acknowledged, with no padding demanded. The
added checks protect the changed behavior while preserving valid prior coverage.
Original graders, frozen solutions and historical scores remain unchanged.

No unresolved Stage D finding remains. I ran no Go tests, graders or live calls
for this final review; acceptance rests on source inspection, independently
checked bindings/examples/counts and the actual retained execution results.


## New21 management implementation review, October 14

Accepted at initial independent implementation review, zero correction rounds:
reordered sourceaae1e23043b317d92c8ec84d35340e91902f2ce9, Agent tree
fd057880524045b1e37a65519a1d5bf7f026c1da. This maintenance reconstruction extends
1b7bb652, retaining the separately accepted handoff correction on new20. The full
mandatory `.agents/skills/ensemble-coding/SKILL.md` and four governing/craft
documents remain active from the preceding reload, with no intervening compaction.
I read the final coder explanation, changed integration and actual raw results,
reusing the approved plan and unchanged historical review instead of reopening
that entire implementation. No code, test, grader or live invocation was performed.

The new public Find replaces the private duplicate and resolves direct ownership
before durable SubmissionData fallback. Typed get_submission emits the payload
as an object; status, output and change-wait retain copied typed status. Existing
Ask/display completion remains text, and submission still records turn completion
after its batch. Composition-bound compiled submit identity is preserved rather
than overwritten by the historical name-based gate. The restored AfterTool branch
ends a rejected unexplained batch without executing its calls or requesting a
corrective response. Ordinary roots/named peers retain chat completion while
delegates owe submission. Real registry/browser controls exercise those roles.

Historical registry and tree implementations are byte-identical; management
behavior differs only by removing the duplicated old string reader and using the
accepted typed result API. Selected conversation/settings/history remain in their
owned GUI views, settings acknowledgment stays outside the observer lock, and
whole replay admission cannot block the shared observer mutex. Actual CLI/GUI
construction uses Start/EnableFleet. The preserved real held-request, selected
versus unselected message and slow-socket tests remain applicable; no managed
compaction cleanup or memory/recall service is imported prematurely.

All177 captured paths match committed Git blobs and working bytes; aggregate
path/space/SHA-256/newline binding is
dbbeba92145d679637577a96bb2fa177a75d12e477b7b92692e8a3342f29bcc5.
The four handoff source/test files match accepted1b7bb652 exactly. Actual outputs
under `/tmp/ensemble-reordered-checks/management-*` show core ensemble15.926s,
GUI7.685s, focused core/GUI race4.879s/4.249s, corrected examples0.011s/0.267s,
and recorded successful vets/builds. Real Chrome fleet capture reports10 provider
requests,111 tree frames, nested selection, two-tab isolation and the four work
phases; existing speech/Responses browser tests both pass. The unchanged current23
grader scores100/100. This is that baseline grader/source result, not a final
revised Stage D score on a reconstructed export or a paid live demonstration.

Both adaptation mutants fail the intended assertion: removing AfterTool handling
returns nil instead of interruption; stringifying retrieval produces a quoted
payload rather than an object. Correct-source restored race passes1.155s. Root's
explicitly narrowed mutation scope reuses historical evidence for unchanged
behavior; no new five-mutant sweep is claimed. The initial nonexistent ChildOutput
alias compile failure and wrong example-directory setup attempts remain recorded;
correcting a test decoder did not require an extra public type.

The coder's predecessor-relative Go production delta is973 physical/827 active/
100 comment-only lines versus historical1,027/870/109; Go tests add713/660/31
versus636/585/30. JavaScript production adds34/32/2 in both, with77/74/1 test
lines versus75/72/1; HTML adds7 and CSS9 physical lines, examples unchanged.
Active counts exclude blanks and `//`-only lines. The historical transfer and
owned-view move account for the size; typed boundaries and AfterTool supply the
additional tests. Existing comments explain ownership and ordering despite a
below-target overall proportion. No new framework or unresolved KISS finding
remains. This accepts new21 implementation only; new22 memory, new23 recall,
reordered exports, final grader mapping and live assessment remain separate.


## Live receipt, new22 reconstruction and Stage E review, October 14

Accepted with the limits below. These are maintenance reviews, with prior-answer
exposure. I reloaded the entire `.agents/skills/ensemble-coding/SKILL.md` and all
four governing/craft documents after compaction before completing this review.
No runtime edits, live calls or repeated grader/test sweep were performed.

The corrected live receipt retains source57920dd3 / Agent9f421ae303b6d63ef8c9734dc3767587dff6340c.
I compared the sanitized receipt with its original L8GeVX capture, binary and
launcher hashes, saved event ordering, accepted submission and actual memory
file. The exact1257-byte memory matches submission and commit, replacing2669
source bytes. Its first-person account retains the Cedar Lantern identifier,
owner/date,30-day choice, rejected cleanup timing and reason, and unresolved
attachment policy without turning proposed work into completed implementation.
The final answer receives the exact memory and correctly answers those facts.
This is useful retention in one bounded Astra/high run, not optimal compression
or an Auto-Recall demonstration; recall was disabled. Four main requests and
one worker request complete, each process exits0, and restart makes no inference.

The worker finished before the unrelated second input and child selection.
Consequently this live run proves no unrelated-turn overlap, selected-child
fallback or before/after auxiliary delta; the earlier held-worker tests supply
those separate claims. First-turn response13 does precede outcome16. I requested
clarification of the receipt's broader “No main answer” limit: root preserved
live.json verbatim and clarified that it means no unrelated second-turn answer.
Root e1ec1511 records that clarification, launch-versus-derived threshold limits,
and the still-failed first HXZDyB attempt. The render-only handoff completion
label fixes that attempt's lost acknowledgment without changing stored notes,
source fingerprints, private-reasoning removal or tool-round policy. The exact
OpenAI cache-test expectation still protects the stable marker boundary, volatile
tail exclusion and marker count; its test-only update does not relabel the live
binary. Earlier actual failure and local regression provenance remain separate.

New22 c67173c18ba09dc6a1d53c607d69da969293562f, Agent23e0aae52df171420e0b00b9e05d3ac447660ea6,
is accepted at initial independent review, zero correction rounds. All198 captured
source hashes match committed blobs (aggregate f0bf38db46701a2550c6b012de436ff1a7df541bf9c2887653fc2112b85b50e1).
The memory selectors/staging, receipt/history, actor compaction, fleet lifecycle
and GUI cleanup carry the accepted Stage C implementation. Recall-only hooks are
properly deferred. The necessary private main/auxiliary send flag retains new16's
resolved-credential sanitization and counts auxiliary usage without replacing
main Last or rotating the main cache lens. The accepted handoff-render fix is
carried forward, including the legacy OpenAI concatenated-text expectation.

Actual memory-* outputs show core ensemble16.959s, GUI8.380s, focused core/GUI
race7.997s/4.456s and restored auxiliary race1.057s; module examples/vet are
recorded green. Both adaptation mutants fail intended main-occupancy/cache-lens
assertions across three vendors. The new-number22 grader scores100/100; all13
recorded grader source hashes independently match the later88ce721a checkpoint.
Unchanged Stage C failure/recovery evidence remains applicable with its original
source identities. Initial fixture compile/cache-expectation failures stay in
the coder's note. Production growth2225 physical/1797 active/326 comment lines
and tests4137/3895/148 follow the historical memory plus reviewed C transfer;
active excludes blank and comment-only lines. Useful ownership/recovery comments
and reuse of existing actors explain the growth; no new framework or KISS finding
remains. Historical deltas are not additive identities or a cold-student claim.

Stage E88ce721a is accepted at initial review, zero correction rounds. I inspected
the moved chapter/grader diffs, dispatch and references against the approved
old17–23→15–21 and old15/16→22/23 permutation. The existing stories, mechanisms,
examples and valid scoring predicates survive. Invoice extends14 and explicitly
defers memory/judge accounting and cache-boundary duties to22/23, where the
requirements remain. Memory extends managed fleet21. Historical rubric IDs,
grade.ChNN calls, models, evidence paths and recorded outcomes remain historical.
The change introduces no application behavior or extra abstraction.

All308 hashes in the actual Stage E run manifest match88ce721a. I read the nine
actual reports: new14/15/16/19/21/22/23 each100/100, new18=155/155 and20=200/200.
Recorded package tests pass (gradeedition235.046s), vet and hard lint pass with
soft warnings retained. The13 default/admission observations select the intended
core/GUI/Gateway paths and reject ungraded17/24/25; the27 input/title/section and
136 local-link controls agree with the mapping. Commit inspection confirms no
preface, frozen export or publication artifact edit. These are explicit-main
mapping results, not proof that old cumulative exports have been reconstructed.
The README/map correctly state that transition. Existing Stage D sensitivity
results are reused because no behavioral predicate changed. No open finding
remains for these three scoped reviews; reconstructed recall and later exports
remain separate work.


## New23 recall carry-forward review, October 14

Accepted at initial independent review, zero correction rounds:8faca2652d6bacd41014632450d5d6798862c852,
Agent4518af327241e99543cd027e62ef9e9596e18f08. The full mandatory skill and all
four governing/craft documents remain active from the preceding reload. I read
the coder's final explanation, actual integration diff and retained outputs;
no application edits, tests, graders or live calls were performed for this review.

Recall spokes, actor recall, history/reducer and skills are byte-identical to
accepted primary57920dd3. This restores the deferred actor hooks, recorded
attachment/hash semantics, Sandbox corpus access and lazy startup/recovery hold.
The new judge uses the existing private send(false) auxiliary seam; future Codex
request fields are correctly absent. The actual main/auxiliary test now exercises
both deprecated Compress and Judge, with distinctive measured usage and unchanged
main cache artifacts. No typed-submission, ownership, OAuth or handoff correction
is overwritten. The trusted skill directory remains construction-owned and does
not enable tools through retrieval.

All206 captured hashes and Git blobs match the submitted revision; aggregate
cd2d2711dca4e7b83a4e1c913d312ae8c4ed17679bc6ea64e7efb7b071fc8243.
Actual recall-* outputs show ensemble32.128s, GUI8.087s, affected race19.191s,
restored auxiliary/recall-unit race1.078s/1.017s, examples0.011s/0.273s and recorded
successful vets. Revised new23 grading is100/100,16.417s wall. The judge-main
mutation fails the intended77/9 main-occupancy assertion on all three vendors.
The first race selector's empty recall-unit selection is preserved, with the
subsequent real unit run supplying that evidence. Historical recall, confinement
and Stage C tests retain their original identities and live/semantic limits.

Predecessor-relative Go growth851 physical/669 active/144 comment-only production
lines compares with historical825/643/143; tests706/635/46 compare with612/550/39.
Active excludes blanks and comment-only lines. The difference restores accepted
confinement/recovery/Responses integration and deferred checks, while reusing the
existing auxiliary transport. Useful ownership, cancellation, prompt ownership
and deduplication comments remain; no extra framework or KISS finding exists.
This accepts the reconstruction, not a new cold-student experiment or paid run.


## Final bounded carry-forwards and deployment handoff, October 14

Accepted at initial independent review, zero correction rounds: corrected 14–21
as identified in [the exact source table](codex-handoff-carryforward.md), optional 24
f18737f62e0f9aaf553a8d8a32334ac2191bb0a9 / Agent 6fdf4cccf54989f45829bbf04ca77ddbd062ab6c,
and Gateway 26 59e25a25dab3c2e7d41c32654013694d7de1bce7 /
Agent 6982337a032962f36626cb23f68e6488046cccdc. The full mandatory skill and all four
governing/craft documents remain active. These reviews reuse accepted historical
behavior and inspect the necessary integration, actual deltas and retained outputs;
I ran no application tests, graders or live calls.

For 14 I compared the exact old solutions/ch14 subtree, not the maintenance
commit's relocated parent Agent. The five changed files carry the accepted
opaque/signature cleanup and shared render-only completion label. Its early
Anthropic HTTP test exercises the real batch, log/snapshot replay and render
purity; a public Record/reducer test checks retained text-bound signatures.
The initially surviving signature mutant exposed a renderer-only assertion;
the added projected-context assertion detects it. The original survivor and
fixture type compile failure remain visible. From 16 onward the Responses/Gemini
HTTP regression is the accepted test with only the unavailable Recall setting
removed. 15's cache expectation remains exact. 17–19's correction hunks match 16;
20–21 already had cleanup and add only the label and corresponding assertions.
The README changes correct current chapter references without rewriting history.
All eight actual core/focused/vet outputs and seven mapped grade reports support
the source table; 17 remains ungraded. No future memory or early Responses code
is imported. Short checks 16–21 explicitly reuse 15's unchanged long lifetime test.
The maintenance checkout omitted the relative precommit hook: the coder later
ran the configured credential hook against all eight exact commit-parent diffs,
all passing. This is a post-commit audit, not a claim the original hooks ran.

Optional 24's credential reader, endpoint restriction, redirect refusal, paired
account/bearer, affinity key, Responses renderer and integration tests match
accepted primary. Existing main/auxiliary transport and resolved redaction remain;
Judge stays auxiliary after receiving the Codex-specific request fields. All 210
captured hashes/blobs match, aggregate c9edbfbc9e3a41e85e6035a71b9c13c4ffc9c2c4d886ceabceb43c75da39381d.
Actual core 32.489s, GUI 8.261s, race 3.473s/1.017s and vets pass. Production delta
193 physical/148 active/33 comment-only lines matches historical behavior
(194/148/33, one blank difference); tests 341/313/17 versus 288/263/15 retain
accepted explanation/summary fixtures. No new grader, paid run or repeated
historical mutation sweep is claimed. The original optional Chapter 24 implementation was explicitly self-reviewed;
this independent review covers its reconstruction integration.

Gateway's 13 added files and dependency union are byte-identical to accepted
8935f486. All 223 captured hashes match, aggregate
71b1a664bb1a104afaabd39d57982bd1d1e34d684e0f4c8bfcb72ababe884cbc.
Its real client/server/Agent race checks pass 1.045s/1.253s/1.045s, vet passes and
revised 26 scores 100/100. GUI verification is explicitly compile-only. The final
Go-sum ordering reconciliation changes no dependencies/checksums; original source
and failed-grade/live evidence remain historical. Production 1871/1475/306 and
tests 592/543/29 are the identical historical deltas. No runtime adaptation, new
framework or unresolved KISS finding remains in either carry-forward.

Root installed exact accepted exports and fresh source tags for 14–24/26, retaining
the complete old directories/manifests at immutable
edition-2-before-background-compaction-20261014 (5d15954e). I updated the current
solutions index from installed manifest identities and preserved original results.
Final integration is accepted: the eight grader-test diffs only replace temporary
main paths with their matching installed chapter paths, including the recall
source/import check; no predicate or weight changes. I verified all 1,960 exported
file bytes, modes and blob identities against 12 manifests and source tags, and
all 155 grader-source hashes in the [final receipt](../../solutions/evidence/background-compaction/reordered-final.json).
Its ten complete default-path reports pass: 14/15/16/19/21/22/23/26 at 100/100,
18 at 155/155 and 20 at 200/200. Root's package tests pass in 36.390s and vet passes.
These adoption results retain their own source binding, separate from earlier
explicit-main, historical-grader and paid live results.

The staged GUI binary is `/tmp/ensemble-gui-background-compaction-b4746ac9`, built
by root from b4746ac985bd84a38d59778ecfa6a3c7f3c3bc7e /
Agent 4ff207e66df6d49cbdea462df2f6753d421779d8 with Go 1.25.4 darwin/arm64.
Its reported SHA-256 is db81a99f17f50a13dd05c98fe620af7a2bd92ce49e7cafa58ce46c2564a4325c;
build and --help exit 0. This staging result is separate from the earlier live
receipt's binary/source. No install, running-process restart, settings change or
production-memory access occurred. Publication artifacts and Bill's preface remain
unchanged; adoption does not imply a published build or production deployment.
