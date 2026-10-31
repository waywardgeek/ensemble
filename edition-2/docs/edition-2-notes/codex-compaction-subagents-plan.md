# Codex handoff: structured submissions, background compaction, and memory-chapter placement

Prepared by **Cadence** (the main Ensemble actor previously addressed as Astra), at Bill's request, 2026-10-14.

**Status:** design/execution handoff, not an implemented or accepted enhancement. No submission-schema or background-compaction implementation changes have been made. Baseline inspected: `bd921235` on `main`. Recheck HEAD and the working tree before starting; another actor will be working concurrently.

## Read this first: parallel work and ownership

**Cadence will be working in the Ensemble repository in parallel on an unrelated change: moving its data directory into its dedicated memory repository.** This is not permission to reset, clean, overwrite, or absorb Cadence's changes into your work.

- Authoritative application/book repository: `/Users/bill/projects/ensemble`.
- Codex owns this compaction/submission-schema/chapter/grader workstream.
- Cadence owns the data-directory migration, its migration notes, and `/Users/bill/projects/cadence-memories`.
- Cadence expects mainly deployment/configuration and memory-repository changes. If application changes are needed, coordinate file ownership first.
- Potential shared hotspots: `agent/ensemble/ensemble.go`, `agent/ensemble/registry.go`, `agent/internal/common/skills.go`, and the GUI configuration/startup files (`agent/gui/application_config.go`, `settings.go`, `gui.go`, `cmd/main.go`). Do not assume exclusive ownership of these paths.
- Prefer a separate worktree/branch if Bill approves it. Never switch/reset the checkout under the running actor. If using the same checkout, announce file ownership and make narrowly staged commits. Re-read status/diffs before editing or committing; HEAD may legitimately advance.
- Keep a Codex-owned progress record, e.g. `docs/edition-2-notes/codex-compaction-progress.md`, containing decisions, tests, checkpoints, and pending issues. Cadence can keep its migration record separately. Coordinate edits to the shared `ch25-crossover.md` rather than racing appends.
- Do not replace the live binary, restart Ensemble, change live memory settings, or change the live launcher/data directory without coordination and Bill's authorization. Build/test in isolated directories.
- Do not push, merge remotely, or publish artifacts automatically. Bill gates pushes.
- Preserve existing unrelated untracked publication and Edition 1 material, including `.swp`, cover assets, old notes/runtime files, and `solutions/evidence/publication-cost-r3/`. Never use broad `git add .`, `git reset --hard`, or `git clean` to make this work convenient.
- Do not assume the old runtime directory remains authoritative. Read current configuration when a live operation is explicitly authorized. Resolve worker paths from the owning Agent, not hard-coded deployment paths.

### Relevant current deployment (observational, not an edit target)

- Launcher/binary: `/Users/bill/ensemble-edition2-codex/start`, `bin/ensemble-gui`.
- File configuration: `/Users/bill/ensemble-edition2-codex/session/settings.json`.
- Last verified data directory: `/Users/bill/ensemble-edition2-codex/session/cr`; Cadence plans to migrate this.
- Live conversation persistence: `session/save.json` plus `session/save.json.journal`, with `.bak` snapshots. **`event.log` is a stale export, not the live journal.**
- Credentials are separate. Do not print, copy into evidence, or commit their contents.
- Last verified memory settings: Conversation enabled, Session disabled, Soul/Memory/8x/64x disabled. Session was restored off after the experiment. Recheck rather than relying on this historical state; do not enable bands merely to run deterministic tests.

## 1. Bill's requirements and remaining decisions

Bill wants:

1. Memory compaction performed by **real, visible background sub-agents** using the normal actor/fleet lifecycle. Main work should continue while ordinary compaction runs.
2. Completed compaction children automatically joined, stopped, and dismissed, retaining an auditable result/failure record.
3. An optional **submission JSON schema** when launching a child, analogous to ordinary tool argument schemas, with **runtime validation when the child calls `submit`**. Results must be able to be structured data, not only text.
4. Chapter TL;DR/contracts and graders updated with the implementation, and the memory exercise passed by the new scheme.
5. A chapter-order solution that does not require students to use fleet machinery before it is taught.
6. Visible action explanations and reasons in the conversation and durable records. Hidden reasoning is not the project handoff format.

**Proposed placement, not yet an independently verified dependency graph:** move current Chapters 15 and 16 together after current Chapter 23. Bill asked about moving the chapter later; the pair is the recommended solution. Audit first and obtain/record confirmation of the final mapping before renumbering. Do not treat this proposal as permission to silently break intermediate exercises.

The background execution policy deliberately revises part of current Chapter 15: its present contract is a finite direct compressor call completed before the next main request. Preserve its memory semantics, but explicitly revise and test the scheduling contract. Do not claim background execution was already required by the old chapter.

## 2. Required preparation

Read the governing documents in full and keep them active for formal author/reviewer work:

- Deployment `HANDOFF.md` at `/Users/bill/ensemble-edition2-codex/HANDOFF.md` (not the repository root).
- `AGENTS.md` and `.agents/skills/ensemble-coding/SKILL.md`.
- `docs/agentic-codebook-workflow.md`, `docs/edition-2-carryover.md`.
- `book/voice.md`, `book/chapter-writing-procedure.md`.
- `docs/edition-2-chapter-map.md`, `docs/edition-2-enhancement.md`, `docs/edition-2-publication.md`.
- Current Chapters 14, 15, 16, 22, and 23, their TL;DRs, and the actual corresponding graders. Audit intervening chapters as described below.
- `docs/edition-2-notes/ch25-crossover.md` for the observed failures and prior fixes.

Do not substitute this plan, a handoff summary, or headings for the full-context authoring prerequisite. Follow the governing coder/author/reviewer separation, model requirements, review requirements, test-first requirements, and acceptance procedure. If a rule cannot be met, stop and surface it rather than implying it was satisfied.

Current Edition 2 source is `agent/`, `book/`, and `solutions/chNN/`. Edition 1 is archived under `edition-1/`. Older literal paths and pauses in historical documents are not current work instructions. Frozen accepted solutions/manifests/evidence are historical artifacts: do not modify them to conceal regressions or pretend the new sequence was the old build order.

## 3. Evidence: what actually failed

The current auxiliary compressor already has a `submit(memory)` tool. Lack of a tool is not the defect.

### Recorded observations

- An earlier handoff returned in about26ms, then launched Session compaction. The main turn waited until Bill interrupted about93 seconds later. That is an **unfinished wait**, not a measured successful compression duration; the old API trace was lost on restart.
- With all six bands disabled, a later handoff returned in about51ms, no new compactor launched, and the exact checkpoint survived graceful Ctrl-C/restart. The next ordinary main-model response took about35.6s and was separate from compaction. Evidence recorded in `48627188`.
- In the instrumented Conversation+Session test, the handoff returned in about44ms. The actual auxiliary request used `gpt-6-astra`, high effort/detailed summaries, and roughly224K serialized input characters. It contained both the main-agent instruction not to submit and the compressor request to submit memory.
- The provider streamed and completed: reported input43,594 tokens, output7,317, reasoning130; its ordinary message contained35,476 characters and **zero function calls**.
- The runtime rejected the completion with `compressor must call submit exactly once`. No Session summary was committed; source was retained. Main resumed approximately3m13s after launch.
- No explicit `tool_choice` or output-token limit was present in that captured request. The instruction conflict is verified; the trace does not prove the model's sole reason for choosing prose.
- The metadata watcher had a five-minute observation window and did not cancel work. Its metadata-only trace is under `session/diagnostics/compaction-live-3ywxg9hp/`. Do not mistake watcher expiry for request failure or print raw conversation/provider/credential payloads.

The last test and interpretation are recorded in `bd921235` and the crossover notes. An earlier naive keyword search matched compressor source code inside ordinary prompts; correct identification checks the actual sole-`submit` declaration with a `memory` property. Never infer a compactor request from a keyword appearing anywhere in the transcript.

**Implication:** background scheduling alone would make the failure less disruptive, not correct. The worker needs a coherent role and enforced submission contract.

### Existing fixes to preserve

- `df38abe0`: `RequireVisibleReasoning` defaults false, enabled for supported real OpenAI models. Visible ordinary text must precede tool calls, including in actual stream order. No length threshold or mandatory headings. Private summaries do not qualify. Rejected calls create no jobs, errors are durable, retries are bounded, after-tool interruption works, and children do not freeze inherited model-specific policy. Synthetic course rows remain available for their historical fixture contracts.
- `09111e0b`: removed the60s total HTTP response deadline. Explicit cancellation and transport connection safeguards remain. Three-second tool callbacks are check-ins, not execution deadlines. Do not reintroduce a blanket timeout as a substitute for lifecycle correctness.
- Main actors finish in chat; delegated workers submit. The compressor must not be seeded with a contradictory main-agent role.

## 4. Chapter placement and historical compatibility

### Proposed mapping

| Current chapter | Proposed chapter | Material |
|---|---|---|
| 0–14 | unchanged | Through Context Engineering |
| 17 | 15 | Accounting |
| 18 | 16 | Provider transition |
| 19 | 17 | Earlier crossover |
| 20 | 18 | Web/status/accounting integration |
| 21 | 19 | Sandbox |
| 22 | 20 | The Ensemble / child construction and submission |
| 23 | 21 | Managing the Swarm |
| 15 | 22 | Forgetting on Purpose |
| 16 | 23 | Auto-Recall |
| 24–26 | unchanged | Cache workaround, current crossover, Gateway |

Use actual chapter titles when updating publication files; shortened labels above describe subjects, not replacement titles.

### Dependency audit before adopting the move

- Check current17–23 prose, TL;DR interfaces, examples, graders, fake providers, fixtures, and parity chains for reliance on memory or recall. In particular, inspect accounting for auxiliary-usage assumptions and the GUI/fleet exercises for memory controls or corpus construction.
- Decide where default child construction and optional submission schemas are first taught (recommended: current22 / proposed20).
- Identify cumulative-code dependencies that cannot be fixed by changing prose links. An accepted old snapshot includes features in the old order; renaming its directory does not create a valid new-order student result.
- Produce a checked old↔new mapping covering chapter filenames/index, section references, grader dispatch/help/default paths, prerequisite checks, manifests, evidence links, examples, and build/publication inputs.
- Current19 is ungraded; if moved, that property follows the exercise, not the old number. Current21/22/23 have special build/public-interface handling. Current15 uses the GUI entry point for settings. Do not mechanically shift integer switch cases and leave their semantics behind.
- Decide and document how current versus historical grading is selected. Preserve historical hashes, reports, and original source identity. If a new mapping needs fresh accepted snapshots, generate them through the governing workflow, not by overwriting frozen evidence.
- Present material dependency problems to Bill before committing to the reorder. A crossover-only extension remains an alternative if moving the foundations turns out disproportionately disruptive, but do not silently substitute it for the agreed work.

Keep old chapter IDs during initial implementation/testing. Renumber only after the mapping is agreed and the behavior passes, with the mapping in evidence reports.

## 5. Optional submission schemas: intended public contract

Recommended launch spelling: `submit_schema`. The concrete internal field/signature may follow project conventions, but the behavior must be explicit.

### Default and custom forms

Without a schema, preserve ordinary behavior and existing callers:

```json
{"agent_id":"reviewer","task":"Review the change"}
```

The default child submits the existing object `{"result":"text"}` and existing text-result consumers keep working.

With a schema, the schema describes the **entire direct argument object of the child's `submit` tool**, not a string containing JSON and not an extra `result` wrapper:

```json
{
  "agent_id":"memory-worker",
  "task":"Produce the source owner's memory from the supplied source",
  "submit_schema":{
    "type":"object",
    "properties":{"memory":{"type":"string"}},
    "required":["memory"],
    "additionalProperties":false
  }
}
```

The compaction child's declared and accepted submission is then:

```json
{"memory":"I decided ... because ..."}
```

This preserves the memory signature that the current chapter and grader recognize while generalizing the normal submission mechanism.

### Enforcement and representation

- Validate the schema at spawn, before creating a partially initialized child. Freeze a defensive copy; a caller must not mutate the live contract later.
- Declare that schema to the model and validate the received arguments against the same canonical contract **before** accepting/publishing a submission or marking the child complete.
- Invalid input returns a clear tool error, does not become a successful submission, and does not prematurely satisfy a join. Ordinary workers may correct a rejected submission within their existing bounded lifecycle.
- Preserve structured results as JSON data through storage, retrieval, joins, status, GUI, and reload. Do not solve this by double-encoding every result into a string. Keep the old string-facing API behavior where required; use an explicit typed/raw JSON representation for custom payloads.
- Audit public Go entry points as well as model-facing tools. A typed old `Submit(string)` call must not bypass a custom contract. Defaults must remain source-compatible where the book/graders rely on them.
- A child's own submission schema is not automatically the schema of its grandchildren. Grandchildren receive the default unless explicitly given their own contract.
- Supported dialect/keywords, nested validation, additional properties, type errors, arrays, enums, null handling, and schema errors must be specified and tested. Current tool declarations carry `json.RawMessage` schemas; the inspected `tools.decode` checks required fields and JSON decoding, not general JSON-Schema semantics. Metadata alone is not runtime enforcement.
- Reuse a genuine existing validator if one exists after audit. Otherwise add one shared, well-defined validation boundary rather than a compression-only pseudo-validator. Do not silently accept unsupported keywords. Do not resolve remote or local filesystem `$ref` values as an unapproved network/filesystem capability. Bound pathological schema/input complexity appropriately.
- Provider-specific declaration normalization must not silently change runtime semantics. Test real wire declarations for the supported adapters and the exact memory schema.
- Schema validity does not prove memory quality. Parent acceptance also rejects an empty/whitespace-only memory and invalid/stale task results.
- Respect `RequireVisibleReasoning`: the child can give a short visible explanation before calling `submit`; that explanation must not be inserted into the submitted memory as wrapper chatter.

## 6. Worker construction and source seeding

Use the normal actor, factory, child ownership map, observation, jobs, and submission machinery. No second hidden request/retry framework and no direct-engine fallback presented as a real child.

A compaction worker starts fresh with a dedicated purpose, immutable input, and its custom submission schema:

- No cloned main dialogue, opaque reasoning state, or main coding primary prompt.
- No auto-recall, inherited memory bands, recursive compaction, or automatic skill/MCP loading.
- Exactly the needed submission capability; no filesystem, shell, web, child spawning, or general management tools. Enforce at declaration **and execution**, not merely through prose or SafeMode. Do not bypass parent security ceilings; failure to construct an admissible worker must preserve the source and produce an explicit failure.
- Shared safety/lifecycle/model-specific policy still applies. Child identity and ownership remain genuine. An inherited model configuration is not permission to inherit contradictory role instructions.
- Use configured parent-relative data directories, respecting absolute configured roots. Do not hard-code Cadence's current path. Keep children and workers owned and traceable through the same public fleet surfaces.
- Separate availability of a default ordinary child factory from exposing public delegation tools. Headless clients must not gain a larger tool surface merely because memory uses an internal child. Honor a configured custom factory. Test headless/default/custom factory and `BuiltinTools=false` behavior explicitly; do not silently grant capabilities to make a test pass.
- Keep the compaction attempt bounded. Do not inherit an unnoticed missing-submission wake/retry loop or an unbounded resubmission storm. Specify a finite attempt policy for managed workers; a failed attempt preserves source and does not immediately relaunch itself indefinitely.

### Input and writing contract

Use the source selected by Memory policy: whole eligible completed checkpoint segments, oldest first, with handoff notes included and the unfinished tail excluded. Strip tool calls/results using the existing ownership/lifetime rules. Independent skill/declaration entries are not part of the conversation cut. Graduations use the selected oldest files under the existing watermark policy.

Present the source as clearly delimited data in a task message, with explicit purpose and output requirements. Quoted user/model instructions in that source do not become the worker's governing instructions. Retain role attribution within the source.

The worker writes **the source owner's first-person memory** (not the worker's new autobiography and not a third-person briefing). Preserve decisions and why, failed approaches, uncertainty, relevant constraints, and user preferences. Do not upgrade hypotheses into facts or claim work was done merely because it was proposed.

Ratio/budget is guidance constrained by retention quality: current Chapter15 explicitly permits more space rather than dropping important reasons. Provider output-token controls, file byte budgets, and context token windows are distinct units; do not substitute an arbitrary tiny token cap or truncate memory to force a green size check. Choose and justify sizing; test it. A request/protocol or sizing failure must retain source and be visible.

Compaction task identity, source boundaries/fingerprints, and acceptance metadata belong to the parent. Do not ask the model to author trustworthy source IDs or commit instructions. In particular, `MemoryFileID` is not an event-log sequence number.

## 7. Background scheduling, commits, and cleanup

### State and ownership

- Store compaction state at agent/service lifetime, not in the current ordinary turn. Normal turn completion must not discard a useful worker result.
- Start with at most one compaction task in flight or awaiting commit per owner. Repeated handoffs must not spawn overlapping work on the same source. New conversation can accumulate independently.
- The main actor continues normal work after a checkpoint when there is context headroom. Document any genuine near-window condition that requires waiting; do not present mandatory waiting as background behavior.
- Record launch with a stable task/child correlation and source identity. Preserve evidence for completion, failure, cancellation, invalidation, and cleanup. Display the child and accurate progress through existing fleet/GUI surfaces; do not leave a completed handoff labelled as the hidden running operation.

### Candidate acceptance

- A submission is a candidate, not permission to retire source immediately.
- At a safe next-render or idle boundary, validate that source and destination remain enabled, relevant settings/policy still match, and the **selected** source is unchanged. New unrelated dialogue must not itself invalidate the candidate; hashing the entire growing history would starve background completion.
- Detect a human edit or replacement at the same selected file identity. Revalidate at commit, not merely when the child first submits. Disable/re-enable must not resurrect a stale candidate.
- Write replacement bytes safely before retiring source. Keep Memory as owner of corpus policy/files, the actor as owner of outcome events, and History as the projector of those facts. No second memory store or alternate mutable dialogue builder.
- Preserve the existing paired cut/population behavior: source remains usable until replacement is valid; no duplicate injected text, no cut of newly appended dialogue, no partial-history mutation when persistence fails.
- Replay uses recorded bytes and decisions with zero model calls and never respawns a worker. Restart of an incomplete attempt retains source and marks the attempt abandoned; later checkpoint policy may reconsider it.

### Completion and cancellation

- **Submitted is not quiescent.** Wait until the child's submission tail, tool result/history, and required save have completed before automatic shutdown/removal. Do not trigger an interrupt inside its submit callback and then call the truncated result successful.
- Make cleanup idempotent with user shutdown/kill and application shutdown. Remove child-manager and application registry references as appropriate, preserving durable trace files. Repeated compaction must not accumulate stopped agents or outstanding goroutines/jobs.
- Define parent interruption separately from parent turn completion. Prefer the same ownership semantics as ordinary background children; explicitly document what user cancellation cancels. Disabling relevant memory should prevent commit and should stop wasted work where safely possible, while still rejecting late results.
- Failures and cancellations retain source, report why, and leave an auditable record even if the child is dismissed. Never silently discard a failed child's evidence.
- Account for child cost once. Keep context-window forcing based on the last completed **main** request's total input, cached input counted once; exclude child output and compressor usage from that threshold calculation.
- No short total model deadline. Test explicit cancellation with a controllable provider instead of relying on a timeout to make hanging tests terminate.

## 8. Staged execution and checkpoint gates

### A. Baseline, dependency graph, and contract decision

Record exact starting commits and dirty/untracked state. Run targeted baseline tests and current graders against explicit source paths. Capture existing failures; a pre-existing file is not evidence its new failure is unrelated. Audit chapter dependencies and publish the proposed old↔new map. Confirm the final placement with Bill before renumbering.

Exit: reviewed scope, ownership, schema semantics/dialect, lifecycle state machine, and acceptance changes. No broad renumbering yet.

### B. Structured child submissions

Implement the optional launch/schema specification, declaration and dispatch validation, structured persistence/retrieval, and backward-compatible defaults. Test normal child/grandchild behavior and invalid attempts. Review independently and commit as a focused stage.

Exit: schemas control the real submit path, invalid data cannot complete a child, structured data survives save/reload, and ordinary children still work.

### C. Restricted background compaction workers

Implement construction/seeding, task lifetime independent of ordinary turns, safe candidate commit, bounded attempts, observability, and automatic cleanup. Eliminate the normal compressor role conflict. Retain any externally required legacy API only deliberately and document it; do not keep the hidden direct route as the new actor's fallback implementation.

Exit: isolated tests prove main progress, correct worker role/schema, no extra permissions, source safety, and clean completion/cancellation. Inspect repeat runs for registry/goroutine/job accumulation.

### D. Grader and chapter contract enhancement

While old numbers still identify the exercises, update current Chapter15's TL;DR and grader together, and the launch/submission contract in current22/23 as necessary. Follow the full authoring and independent review process; derive narrative from tested implementation and observed evidence. Update memory/recall interactions in current16.

Exit: the revised memory contract passes on explicit current source; preserve each original behavior or document a deliberate specification replacement. Tests distinguish actual runtime behavior from self-reported progress.

### E. Adopt the validated chapter mapping

Only after the dependency and behavior gates, apply the agreed renumbering/cross-reference/grader-dispatch/publication-input mapping. Run every affected exercise/parity chain under its new identity and validate historical access/provenance. Do not silently relabel accepted old-order snapshots as new-order results.

Exit: coherent reading/build order, verified commands and links, no unexplained forward prerequisites, and historical evidence still identifies what was actually tested.

### F. Bounded live evidence, final review, and coordinated deployment

Run an explicitly authorized, isolated live compression with the normal model-specific visible-explanation policy active. Capture safe metadata and inspect the actual first-person memory for reasons/uncertainty, not just successful JSON. Record model, route, settings, source/output sizes, timings, actual submission, committed bytes, main responsiveness, and automatic dismissal. No credentials/private reasoning dumps in the evidence repository.

Finish independent review, scoped commits, crossover/progress notes, and a deployment handoff. Coordinate with Cadence's migration before installing anything; Bill controls restart and push. A compiled binary is not live verification, and a green fake-provider test is not evidence of semantic memory quality.

## 9. Test and grader requirements

### Submission-schema tests

- Absent schema preserves existing text-submit behavior and public callers.
- The exact custom argument schema reaches the real provider request.
- Valid structured result survives submit, join, retrieval, restart, and GUI formatting without double encoding.
- Wrong/missing types, required fields, nested violations, forbidden extra properties, invalid schemas, and unsupported/dangerous references fail at the correct boundary.
- Invalid submission does not mark completion; a permitted corrected submission works.
- Mutating caller-owned schema/result buffers cannot alter the child's contract or recorded result.
- A child's schema does not accidentally become its grandchildren's default.
- Actual execution admission rejects forged unadvertised tool calls; a prompt-only restriction is not sufficient.

### Background-memory tests (real actor/engine/fleet; provider is the fake)

- Hold a child response behind a barrier and prove the main agent can finish a turn and accept further work.
- Verify dedicated first-person instructions, source completeness/role attribution, exact `submit(memory)` schema, and no coding primary/no-submit contradiction.
- Verify visible-explanation policy on real flagged models; fixtures must provide a valid visible explanation, not disable enforcement to get past it.
- While held: append unrelated main dialogue (candidate can remain valid), edit selected source (must invalidate), disable/re-enable bands (must not accept stale work), and change relevant settings.
- No duplicate workers for the same source; no recursive compaction or unexpected tool/network/file capability in the child.
- Validate atomic source retention, correct paired events/canonical ordering, cut boundaries, output persistence errors, and no application to an already-sent request.
- Replay requires no LLM and never respawns pending work. Failed/empty/malformed/late results, user cancellation, shutdown, and restart preserve source and cannot apply twice.
- Observe submit completion before cleanup; test save failure and cleanup racing manual kill/shutdown. Repeated workers leave no live registry references or jobs.
- Repeat headless, GUI, default/custom-factory, and restricted configurations; use genuine public APIs and honest unsupported-case errors.
- Test graduation progress. The read-only review found a separate possible issue with reversed memory-file coverage ranges (`Thru < File`): never retire nothing and repeatedly schedule identical work. Establish the defect with a focused test before changing it; do not cite it as the cause of the observed conversation-compressor failure.

### Preserve Chapter15's existing obligations

The current eleven weighted checks cover: oldest completed-work compression, planted-fact delivery and re-injection, oldest-first graduation, disable/enable idempotence, disabled-neighbor refusal, forcing, replay without LLM, abandonment safety, first-person memory as data, startup population, and Chapter14 parity. Preserve these behaviors. Current implementation is `internal/gradeedition2/ch15.go`, with weights reused from `internal/grade/ch16_checks.go`.

Add explicit observable checks for background main progress, real child lifecycle/restriction, structured submission validation, stale-result rejection, and automatic cleanup. Version and document any scoring changes rather than casually assigning arbitrary points or deleting coverage.

The existing memory router identifies the sole submit tool's `memory` field. Preserve that useful wire distinction. Its fixed main-request ordinal assumptions must become concurrency-aware barriers/predicates with bounded waits. A fake provider may supply deterministic content, but do not fabricate successful retirement, population, lifecycle, or replay evidence. Invalid/abandoned attempts also need observable terminal outcomes so the grader does not wait only for a success event.

Run deliberate negative controls/mutations for key checks (accept invalid schema, expose an extra tool, block the main turn, apply stale memory, cut before persistence, skip cleanup). A green suite alone does not show that a grader detects the defects it claims to detect.

### Commands verified against the current entry point (old numbering)

From the repository root, explicit-source examples are:

```sh
go run ./cmd/grade-edition2 -ch 14 -json ./agent
go run ./cmd/grade-edition2 -ch 15 -json ./agent
go run ./cmd/grade-edition2 -ch 16 -json ./agent
go run ./cmd/grade-edition2 -ch 22 -json ./agent
go run ./cmd/grade-edition2 -ch 23 -json ./agent
```

The submission directory is positional, not a `--solution` flag. Omitting it selects `solutions/chNN`; do not accidentally grade a frozen snapshot while claiming current-source acceptance. The default GUI path is `<submission>/gui`; `grade.Build` prefers its `cmd/` when present. Re-verify all commands after changing chapter mapping/build routing. Chapter26 and other affected exercises also require regression coverage; select them based on the dependency audit.

Other expected checks, with appropriate fixture setup:

```sh
(cd agent && go test -short -race ./...)
(cd agent && go vet ./...)
(cd agent/gui && go test -race ./...)
(cd agent/gui && go vet ./...)
go test ./internal/gradeedition2 ./internal/grade
```

Run the deliberate long-response tests separately as well; `-short` intentionally skips them. Do not use a90s suite deadline for a suite containing61s cases plus other work and then diagnose a compactor loop from that timeout.

For fixture runs, use a child-process environment stripped of inherited provider/runtime variables (prefixes `LLM_`, `OPENAI_`, `ANTHROPIC_`, `GEMINI_`, `EN_`, `CODEX_`) while preserving normal build prerequisites. Do not unset the live application's environment or print its credentials. Set fixture values explicitly. Live Codex settings previously contaminated fake-provider tests.

Browser scripts are not a uniform glob-driven suite: components requires an isolated `ASSET_URL`, fleet requires `GUI_BINARY`. Never connect a test browser to the live GUI and steal its MCP registration. Use their actual harnesses.

Previously observed baseline failures include `TestFleetCLIHumanReplyAndEscalation/true` (human sink unreachable) and a browser virtual-user speech assertion expecting `submit` in the main persona's tool list. They were reproduced before this change; re-establish baseline on the chosen source rather than assuming every future occurrence is unrelated. Report failures honestly, separately from successful new checks.

## 10. Documentation and final delivery

Update the relevant TL;DRs, model-authored prose, grader specification, mapping, CLI help, examples, references, and evidence under the governing review procedure. Reconcile the old wording that requires synchronous direct compression; it cannot remain the normative contract while a new grader expects background children.

Keep Chapter14's invariants: log versus working set, ordered replay, stable prefix/lifetimes, visible why, tool cleanup without erasing dialogue, and the same main actor across handoff. Keep Chapter15's first-person continuity and source-retention rules. Reassess Chapter16 against the corpus/replay semantics, not merely compilation.

No automatic publication rebuild/upload or remote push. Bill can choose when to regenerate/release artifacts after reviewed chapter changes. Coordinate shared crossover-note edits with Cadence.

Deliver to Bill and Cadence:

- Final design/API/schema contract and explicit lifecycle/cancellation semantics.
- Confirmed chapter mapping and a list of affected references/graders/manifests.
- Scoped implementation, grader, chapter, and review commit IDs.
- Exact test/grader commands, source revisions, results, negative controls, and remaining failures.
- Live evidence with privacy boundaries, actual submitted memory quality assessment, and measured main responsiveness/cleanup.
- Compatibility/migration notes, including historical snapshot handling and any persisted child/schema-state changes.
- Deployment status stated precisely: source, built, installed, restarted, live-verified. Include rollback and required user actions; do not assume Cadence's paths stayed unchanged.

## 11. Stop conditions

Stop and ask Bill/coordinate rather than improvising if:

- The proposed chapter move leaves a real forward dependency or requires rewriting historical accepted evidence.
- Required full authoring/review context or approved model/role separation is unavailable.
- Shared-file changes overlap Cadence's migration or unrecognized work appears in your target files.
- Restricting a worker requires bypassing a parent's security ceiling or silently expanding headless capabilities.
- Schema enforcement would be metadata-only, silently partial, or dependent on unapproved reference fetching.
- A candidate can erase source before a durable replacement, be applied to the wrong generation, or finish cleanup before its submission tail/save completes.
- Tests only pass by weakening an unrelated contract, fabricating outcomes, skipping failure paths, or suppressing the visible-explanation gate.
- Live testing would require changing production settings, revealing credentials/private reasoning, repeated costly requests, or restarting the active application without approval.

This handoff is complete enough to start a disciplined design audit. It is not a claim that the dependency audit, new architecture, grader, or revised chapters have already passed.
