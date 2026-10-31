# Background compaction and structured submissions

Coordinator: Astra (Codex), October 14, 2026. Current status: Stages B and C
runtime implementation accepted. Final source `7feb910a` and independent review `09e68149`
close the background actor gate, including the factory-refusal and GUI corrections.
Stage D author and grader plans are independently accepted for bounded edits.
Accounting through Sandbox reconstruction is accepted through new19. Stage D
narrative is checkpointed at `e4b7480c`, with reviewed clarification `8057f854`.
Grader slices D1–D5 are checkpointed through `8430d4b6`; remaining memory/recall
checks and final TL;DR contracts are in progress. Independent grader review and
new20 implementation are active. No exports, renumbering, deployment or publication by this workstream.
Earlier observations retain their actual source/results.

Bill authorized proceeding after reading Cadence's
[handoff](codex-compaction-subagents-plan.md). He separately required full
Chapters 14 and 15 and both governing documents before surgery. Those full
reads are complete, including recovery of truncated output. The coordinator
also read AGENTS.md, the entire `.agents/skills/ensemble-coding/SKILL.md`,
deployment HANDOFF.md, voice and chapter procedure, chapter map, enhancement
record, publication record and crossover notes. Reload full governing/craft
documents after compaction before formal authorship/review.
After coordinator compaction, all four governing/craft documents and the entire
coding skill were reloaded in full before assessing the completed audit notes.
The coordinator repeated that full reload after the next compaction before
accepting the accounting reconstruction plan; the truncated workflow range was
read separately to recover the missing text.

## Isolation and ownership

- Baseline: `56ec71cbde8477b6959ad2fbc65f9237a0fa8346` on main.
- Worktree: `/Users/bill/projects/ensemble-compaction`, branch
  `codex/background-compaction`. Main checkout remains untouched.
- At capture there were no tracked modifications. Existing untracked publication,
  historical, runtime and evidence paths remain untouched. Full status is in
  `/tmp/ensemble-compaction-baseline/initial-status.txt`.
- Cadence owns its data-directory migration and memory repository. No edits to
  deployment, credentials, launcher, runtime settings, binary or shared crossover
  notes. Reconcile any overlapping application changes before integration.
- Codex owns new progress/audit notes and the eventual submission/compaction
  changes in this separate worktree. Do not stage unrelated files or push.

Later main reached `f910e8f5`; its two new commits change only Cadence's migration
note, with no application-code overlap. The coordinator read the full updated
note: Cadence records Bill applied the move and verified runtime output under
`/Users/bill/projects/cadence-memories`, preserving the explicit save path and
all six disabled bands. These are Cadence's observations, not a Codex live test.
Do not rerun migration or assume the original deployment data path remains active.
Reconcile those note-only commits at a clean checkpoint; preserve both workstreams.

## Review assignments

- `compaction_dependency_audit`: independent chapter/fixture/parity/snapshot
  dependency audit, read-only except its own audit note.
- `compaction_runtime_audit`: independent runtime ownership, schema and lifecycle
  audit, read-only except its own audit note.
- Both are GPT-6 Astra/high, fresh contexts, with explicit full-reading and
  `.agents/skills/ensemble-coding/SKILL.md` handoffs. Their initial Stage A was
  read-only; the separately assigned implementation roles are listed below.
  This is maintenance of the existing runtime, not a new blind student evaluation.

## Stage A observations

Chapter 15 explicitly requires a finite synchronous direct compressor and says
it is not a sub-agent. Background children deliberately replace that scheduling
contract; they must preserve its source-retention and memory semantics. Current
full governing records contain historical versions/pauses; their current header
and Bill's new direction govern this maintenance.

Sequential baseline commands use an environment stripped of inherited `LLM_`,
`OPENAI_`, `ANTHROPIC_`, `GEMINI_`, `EN_` and `CODEX_` variables, only in the test
process. Raw output and command/timing records are in
`/tmp/ensemble-compaction-baseline/`. No live provider requests are authorized
by these fake-provider checks.

- `go run ./cmd/grade-edition2 -ch 14 -json ./agent`: 100/100, exited 0.
- Same for 15 and 16: each 100/100, exited 0.
- Same for 22 and 23: build failed before grading. `grade.Build` selects
  `agent/cmd/`, which now has subcommands but no Go source of its own. No runtime
  behavior was assessed by those invocations. Preserve this result; investigate
  a narrow explicit-entry-point fixture repair separately.
- Supplying a CLI built with `go build -o /tmp/ensemble-compaction-baseline/ensemble-cli .`
  from `agent/` to the unchanged `gradeedition2.Chapter22` and `Chapter23`
  functions gives 200/200 and 100/100 respectively. A temporary root-module
  driver invoked these functions with explicit core/GUI paths, then was removed.
  These diagnostic passes bypass only build selection; the advertised CLI
  invocations above still fail. No grader source was changed.
- Core `go test -short -race ./...` exited 1; root package reproduced
  `TestFleetCLIHumanReplyAndEscalation/true: human sink was unreachable` before
  any edits. Other reported packages passed. No all-green suite claim.
- Core and GUI `go vet ./...` each passed.
- GUI `go test -race ./...` initially failed because this new worktree lacked
  Playwright. After `npm ci --ignore-scripts --no-audit --no-fund` in its GUI
  directory, the same suite passed (11.916 seconds). No tracked dependency
  files changed.
- Root `go test ./internal/gradeedition2 ./internal/grade` exited 1 (21.281
  seconds): `TestChapter13Speech` cannot import Playwright from the frozen
  `solutions/ch13/gui` fixture. Installing current GUI dependencies does not
  install that separate snapshot's dependencies. This setup failure is retained;
  no speech-runtime regression is established by it.
- The separately run response-lifetime/cancellation baseline now passes:
  `go test -race ./ensemble -run 'TestModelResponseBeyondOneMinute|TestBlockedModelResponseCancellation' -count=1`
  from `agent/`, exit0 in63.110 seconds. Both real61-second cases were exercised.
  New compaction/schema tests and live work have not run. The baseline is not
  acceptance of the proposed implementation.

## Accepted prerequisite: grader entry-point selection

Coordinator changed `internal/grade/harness.go` to prefer `cmd/` only when Go's
package metadata identifies `main`. A directory containing subcommands or a
library no longer hides the root executable. A selected command's compilation
error still fails; it does not fall back to a healthy root. No grading predicate
or application behavior changed.

Five real temporary-module tests cover root, nested command container, command
preference, library fallback and broken preferred command. Before the fix, the
container and library scenarios failed; all pass afterward. Independent review
accepted with zero correction rounds. Its scratch mutations (always-root and
restored directory-only selection) fail the intended checks, with the correct
control passing; exact hashes and comparison are in the dependency audit.

Actual advertised explicit-source commands now pass: Chapter22 200/200 and
Chapter23 100/100, exits0 in2.489 and1.125 seconds. Raw reports use
`grader-{22,23}-buildfix` under the existing temporary evidence directory.
After installing locked dependencies with `npm ci --ignore-scripts --no-audit
--no-fund` in the isolated frozen ch13 GUI fixture (only ignored node_modules),
`go test ./internal/gradeedition2 ./internal/grade` passes in30.429 seconds.
No frozen tracked source changed. Earlier failures remain recorded above.

## Reviewed dependency result and approved mapping

Both independent audit notes are complete:
[chapter dependencies](codex-compaction-dependency-audit.md) and
[runtime boundaries](codex-compaction-runtime-audit.md). Their findings are
source review, separate from the coordinator's executed baseline above.

Recommended mapping: current17–23 become15–21; current15/16 become22/23;
0–14 and24–26 keep their numbers. The accounting chapter currently requires
compressor/judge usage and memory cache boundaries before the proposed new
memory chapters. Stage those obligations at new22/23 while keeping ordinary
accounting and handoff boundaries in new15/18. Nothing is dropped. Memory's
existing explicit Chapter14 parity stays Chapter14 parity.

The cumulative references are the material consequence: old17 already includes
memory/recall. Renaming it new15 cannot demonstrate the new sequence. Recommend
reviewed staged reconstruction from accepted14 through new23, then affected
carry-forward/validation for24 and26. This is maintenance using reviewed source,
not a fresh blind student run. Replace book-facing exports only after validation;
preserve existing tags, source/manifests/results and Edition1. Historical grading
can use the immutable old checkout and its own grader without a new dual-mode
framework. Chapter25 remains a placeholder.

Bill approved this mapping, staged accounting and reviewed replacement of current
exports on October14: "Yes, approved." This resolves the handoff's mapping and
historical-export decision. Apply the move only after the behavior/review gates;
the approval is not acceptance of unbuilt intermediate solutions.

Stage B is assigned to `structured_submission_coder` (fresh GPT-6 Astra/high),
with `compaction_runtime_audit` as independent design/code reviewer. The coder
must present concrete schema/API/persistence design before implementation.
Coordinator owns the narrow grader build-selection repair, independently reviewed
by `compaction_dependency_audit`. Work remains in the isolated worktree.
The Stage B design is now accepted by the coordinator and independent runtime
reviewer: see [the coder's contract](codex-structured-submissions.md). Implementation
is released. The design uses a pinned genuine validator with an explicitly
admitted Draft2020-12 subset and no references; admission bounds reject rather
than truncate input. This is design acceptance, not implementation acceptance.

Separate preparation branch `codex/reordered-foundations` at
`/Users/bill/projects/ensemble-reordered` starts from `a8eec84a`.
`reordered_foundations_coder` (fresh GPT-6 Astra/high) owns its staged `agent/`
and note, with the dependency reviewer checking it independently. Its first scope
is accepted14 plus accounting only, with future memory/recall obligations deferred.
This isolated preparation does not replace runtime main or publish new snapshots.
The coordinator and dependency reviewer accepted its concrete source plan before
implementation: accepted14 plus the selected16-to17 accounting delta, with
memory/recall integrations deferred to their new introduction points. Existing
supported-model and response-lifetime corrections carry forward explicitly;
affected separate example modules join the core/GUI validation. Implementation
was released in that branch. The resulting `288fc6e8` (agent tree
`2b671b45dc5f7b760f2faaff2a77c83e6c7c09c8`) is now independently accepted at
initial review, zero correction rounds. Its old14/17 graders each score100/100;
affected checks and five targeted defects pass their intended observations.
The original concurrent EOF timing failure remains unexplained, alongside passing
exact14 and isolated staged controls. No export, renumber or auth stage is accepted.

Validation encountered host disk exhaustion at142 MiB free. The coordinator
paused new builds across both coders and the reviewer, verified no Go build/test
process remained, and ran `go clean -cache` against the11 GiB regenerable build
cache. The command completed successfully and restored11 GiB free. Source,
module downloads and evidence were preserved. Link/build failures during ENOSPC
are retained as setup failures; affected checks require clean reruns.

## Stage B acceptance

Independent runtime review accepts the frozen structured-submission implementation
at initial formal review, zero correction rounds. The coordinator read the full
disposition and accepts the stage. Common holds vocabulary, Fleet owns the compiled
contract and durable results, and the composition root refuses invalid construction
before factory effects. Legacy APIs/rows remain compatible; custom results retain
JSON data through public retrieval and GUI live/replay paths. The accepted bounded
dialect and admission ceilings are in the coder's contract.

Reviewer feedback corrected canonical-size/reload inconsistency, multiplicative
validation-work exposure, unsafe shutdown on a wrong-owner factory result, missing
required-parent admission and an incomplete GUI assertion. These were genuine
pre-submission assistance, not independent student discoveries. No Stage C cleanup
or memory behavior is claimed by this acceptance.

Final focused races, full GUI race and both vets pass. Full ensemble race passes
in91.778s, with the later parent guard separately covered by focused races. The
full core suite retains its reproduced CLI human-sink failure. An unnecessary
browser glob invocation lacked required harness environment; its separate speech
assertion also fails on exact `a8eec84a`. Failed fixture/setup attempts and the
mutation evidence's coupled/repeated claims remain in the coder/reviewer notes.

Coordinator ran advertised explicit-source old22/23 commands sequentially on the
frozen stage:200/200 in2.356s and100/100 in1.072s, exits0. All214 source-manifest
files remained unchanged during the runs and matched at independent review.
The reviewer verifies21 distinct targeted defects across24 executions with stated
limits. Changed-file active production grows306 lines; the reviewed explanation
and runtime coverage justify that scope. No new result browser or scheduler.

## Runtime decisions to carry into the implementation plan

Coordinator preparation reproduced both source-identified disk/event windows on
an exact `a8eec84a:agent/` archive. The temporary external module uses real
Ensemble, Memory and History with an HTTP provider fake. It saves at the selected
boundary, creates a fresh Agent, restores that saved history and starts an ordinary
turn. This is a controlled restart-boundary probe, not a forced process kill.

- After replacement `Memory.Write`, before cut/population, restart injects the
  uncommitted candidate alongside the retained conversation. The source-retention
  control passes, then the candidate-absence assertion fails.
- After paired retirement/population, before `Memory.Archive`, the pre-restart
  context has retired the source, but restart reinjects those old source files.
- Controls omit the first replacement write and complete the second archive;
  both pass on the same unchanged application source.

`go test -mod=mod -v -count=1 ./...` in the temporary probe module: baseline
exit1, two intended failures (0.010s); controls exit0, two passes (0.011s).
The original probe was restored after controls. Raw files are under
`/tmp/ensemble-compaction-window-probe/`; probe SHA-256
`ef869c4e0945e84c763c837dfd81f7a9f1ca85941b4aa4139d1418b5c5fae2af`.
An initial run against the in-progress Stage B tree reproduced the same failures;
the immutable-archive repeat establishes that they precede Stage B. These results
support the narrow Stage C recovery work, not a claim that recovery is fixed.

A separate exact-baseline range probe creates two source files whose parsed Thru
precedes File. Scan accepts them; a completed write/paired retirement/population/
archive retires no context entries, and Graduation immediately selects the same
source again. `TestReversedRangesCannotRepeatFold` fails for that intended reason
(0.008s, exit1). Replacing only the filenames with valid single-file identities
passes (0.008s, exit0). The original probe was restored; SHA-256
`a513c5c1fde6fa7bbd26ddc7ea0924394117f0d9264b41a421784be4056b27dd`, file
`ranges_test.go` and logs in the same temporary directory. This establishes the
separate no-progress defect before repair; it does not explain Cadence's observed
Conversation compressor failure.

- Preserve the current accounting promise: compaction cost belongs in owner
  session totals exactly once, attributed to the actual model/tier, while main
  last-request counts remain separate. A separately retained child receipt is
  necessary but does not replace that existing promise. Review the concrete
  auxiliary import boundary before implementation; never count it again through
  a fleet aggregate or replay it as new spending.
- Ordinary parent-turn completion and interruption leave useful compaction alive,
  following the handoff's ordinary-child preference. Explicit worker kill,
  owner shutdown and relevant memory disable cancel/invalidate it. This changes
  the old interrupted-direct-compressor expectation deliberately.
- A submission is a candidate. Acceptance waits for reliable child completion,
  including save, and a safe parent boundary with selected-source/settings
  revalidation. Automatic cleanup removes Fleet, root and GUI references and
  preserves durable outcomes. No inference on replay.
- Retain default string submission compatibility and add an explicit raw JSON
  custom path; a string convenience call must obey the same custom validator.
  The accepted Stage B contract specifies the schema subset, pinned validator and
  durable envelope. Bounded worker request policy and Memory-owned crash recovery
  remain Stage C design work before that implementation.
- The audit initially identified disk/event crash windows from source; the
  controlled restart probes above now reproduce them before implementation.

## October 14: next-stage preparation

The coordinator reloaded all four governing/craft documents and the complete
coding skill after compaction. Primary maintenance HEAD is `0a732ed4`, which
merges Cadence's documentation-only migration commits through `f910e8f5`.
Cadence reports the applied data directory as
`/Users/bill/projects/cadence-memories`; its explicit conversation save remains
at the deployment's `session/save.json`. This is Cadence's recorded live
verification, not a new coordinator deployment test. No application overlap or
live changes were introduced by the merge.

The independent dependency reviewer accepted authentication source plan
`185a1e31969eb94c2cbec54a9437455988ce1fba`. The coordinator read the full plan
and disposition and released that bounded reconstruction from the accepted
accounting tree, with implementation review still required. It retains the
single Send path, resolved-credential sanitization, once-only successful renewal
accounting and the no-total-response-deadline correction. Recall/judge integration
remains deferred. No export, renumbering or new live use is implied.

Coordinator source inspection found two Stage D fixture issues at this HEAD:

- `internal/gradeedition2/ch15_driver.go:memoryCompressor` checks only that the
  sole tool is named `submit`. The handoff's claim that it also inspects the
  `memory` property is inaccurate. Revised classification must distinguish an
  ordinary submit-only child from the compaction worker's actual contract.
- `internal/fakevendor/fake.go:handle` holds its server-wide mutex while calling
  `Options.Route`. A held memory route therefore blocks main requests and
  request inspection too. Releasing only the memory router's own mutex cannot
  establish background progress. Audit other route consumers before a narrow
  shared-fake correction, or use an appropriate external fake; preserve archived
  Edition 1 graders.

The coordinator then reproduced that fake-server serialization through real HTTP
in `/tmp/ensemble-fakevendor-concurrency-probe`. On the unchanged helper,
`go test -mod=mod -race -v -count=1 ./...` fails both intended assertions:
the unrelated main request and `Requests()` remain blocked while the routed
auxiliary response is held (exit1,0.612s). Releasing only that gate before the
main request gives the passing control (exit0,1.017s). Original test restored,
SHA-256 `13511e65951acc094fc04f8e72faa6a4e18e060dcf469b3ff96bd4ba592ffaaa`;
raw `baseline.log` and `control.log` remain beside it. This proves a fixture
limitation, not a new runtime compaction result. No fake implementation changed.

Further source inspection found an existing narrow option:
`ch23_driver.go:managementGUI` holds selected HTTP requests in a reverse-proxy
boundary outside the same capture mutex. Stage D should consider that established
pattern before changing shared fake concurrency semantics for every route caller.
This is a design option, not an implemented correction or extra requirement.

The current main-turn completion barrier also does not establish background
worker completion. Stage D needs observable separate outcomes and causal
barriers while retaining all eleven existing weighted obligations. No grader,
scoring or chapter change has been made or approved by this preparation.

## Stage C design accepted; implementation released

Coordinator and independent runtime reviewer accept the revised complete
[`codex-background-compaction.md`](codex-background-compaction.md). The design
uses one owner-lifetime attempt, restricted real children, reliable completion
and save, selected-source validation, Memory-owned hidden staging and journal
receipts, and bounded child/reference cleanup. Three Engine.Send invocations
are distinct from existing bounded authentication retries. Ordinary parent
completion/interruption preserves useful work; explicit kill/shutdown and
relevant policy changes cancel or invalidate it. The stated near-window wait
is limited to an outstanding attempt at the existing 95% main-input threshold.

Pre-implementation review clarified bounded attempt-owned accounting instead
of an accumulating deduplication set, default tool registration versus actual
safety ceilings, and the trusted custom-factory callback boundary. A valid JSON
submit accompanied by a provider length/incomplete terminal reason still fails
before dispatch. Committed-but-unfinalized disk work blocks recall, corpus scans
and main rendering until recovery; it cannot be mislabeled abandonment. The
runtime audit records the independent disposition and required evidence.

The coder is released to implement coherent tested slices and checkpoint them,
with independent implementation review afterward. No runtime correctness or
semantic memory quality is established by design acceptance. Grader/chapter
changes remain Stage D, and renumbering/export/live work retain their later gates.

Parallel authentication reconstruction is accepted at `970676abeebce8f156d5cdc7ee0dd47a40708136`,
Agent tree `21b15bc6cd1c24a2ee3e03873a87a9c080744402`, initial independent review
with zero correction rounds. The reviewer verified all reported checks and source
counts, old18/17/14 each100/100, both61-second regressions and four intended
mutation failures; its independent frozen-source focused race run passes1.236s.
The coordinator read and accepted the full disposition, then released a source
plan only for current19/new17, the still-ungraded historical crossover. Historical
live evidence keeps its source identity; no paid repeat, export or renumbering.

Main advanced to documentation-only `67060b6e` while this work continued. Its new
`cadence-session-directory-migration.md`, read in full by the coordinator, describes
a prepared and fixture-tested move of application session files into
`cadence-memories/session/`; it explicitly says not applied. It adds no application
code conflict. Merge that note at a clean checkpoint, and re-establish actual
deployment paths before any later authorized live operation. Do not treat the
proposed session location as active or run Cadence's migration.

## October 14: coordinator review and reconstruction resumed

After answering status questions, the coordinator resumed active work at Bill's
"Please proceed." Earlier replies had ended coordination while runtime delegates
continued; that delayed the crossover source-plan gate. This repeats the workflow's
handoff lesson and is coordinator delay, not a technical blocker.

Before formal review, reloaded all four complete governing/craft documents and
the entire coding skill, then reread all current Chapters14/15 and the complete
handoff. Root inspected the historical crossover's complete three-path delta,
plan484c7234 and independent dispositionfa29b7ee. The bounded current19/new17
implementation is released with its stated two mutations and affected checks.
It remains ungraded; no paid task, new export or whole-tree merge is authorized.

Runtime slices C1–C9 are checkpointed throughf111f1b3: range admission,
restricted finite workers, provider sizing/terminal handling, selected-source
validation, correlated journal receipts, staged filesystem recovery, auxiliary
usage buckets, child completion/removal, and GUI dismissal. Each has scoped
independent review; none alone establishes the complete background actor.
The coder's C10 working tree is separate from this exact review baseline.
Root's overall review is underway, not accepted, and uses the immutable source.

### Reproduced reverse-purpose factory refusal defect

Root's real-runtime external-boundary test on exact f111f1b3 calls
StartCompaction with a trusted constructor-only custom factory. The factory
returns a real owned child with Compaction=false and LogPath pointing to a
preexisting sentinel file. CreateChild correctly refuses the purpose mismatch,
but its Shutdown writes the empty log header over that file; ForgetCompaction
then retains the rejected ordinary child because its purpose is false.
The same redirected-path factory retaining Compaction=true is a passing control:
refusal preserves the log and removes that managed child. No inference or mock
was used, and arbitrary callback side effects are outside this finding.

Actual command: `go test -race -run '^TestRootFactoryPurposeRefusal$' -count=1 ./ensemble`
in an isolated f111f1b3 Agent archive, with provider/runtime variables stripped
and a fresh EN_HOME. Exit1,0.022s package time; the ordinary-return case fails
both exact-byte preservation and root-membership checks. Raw output:
`/tmp/ensemble-root-review-factory-refusal.txt`; scratch location is recorded in
`/tmp/ensemble-root-review-path.txt`, with test `agent/ensemble/root_factory_refusal_test.go`.
The archive extraction emitted a Python tarfile deprecation warning; compilation
and test execution succeeded before the intended behavioral failures.

Coder and runtime reviewer received the unchanged probe. The existing C8 purpose
regression covers the opposite ordinary-request/managed-return case and missed
this direction. Prior scoped acceptance remains history. Correction must preserve
ordinary valid lifecycle while rejecting the returned child without writing into
its rejected paths or retaining its new registry entry. Recheck the unchanged
probe and appropriate ordinary/managed controls before accepting the correction.
This is overall review of accepted support, not a claim C10 has been submitted.

## October 14: integration review and next reconstruction handoff

After the next compaction, the coordinator again reloaded the complete skill,
workflow, carryover, voice and chapter procedure, then all Chapters 14/15 and the
handoff before resuming formal review. The initial combined tool output was
truncated; separate complete reads recovered it. Summaries were not substituted.

The unchanged root factory-refusal test body now passes on the C10 correction:
`go test -race -run '^TestCoordinatorFactoryPurposeRefusal$' -count=1 ./ensemble`,
exit0,1.030s. Root overlaid current changed/new Agent files onto its exact f111f1b3
scratch, binding file hashes in `/tmp/ensemble-root-review-fixed-source.json` and
checking those source bytes did not change during capture. The first attempt
failed to compile because the coder had retained the same test name; root renamed
only its independent function/comment, preserving the body and assertions.
That setup output remains `/tmp/ensemble-root-review-factory-refusal-fixed-setup.txt`;
the actual passing result is `/tmp/ensemble-root-review-factory-refusal-fixed.txt`.

Root inspected the explicit construction-discard path and exact-identity removal,
then the complete C10 production diff. The earlier source-only completion-count
concern is corrected: attempt identity and first-completion admission precede
worker decrement and auxiliary import. No duplicate-delivery runtime reproduction
is claimed. Independent runtime review separately identified the GUI consequence
of removing a rejected ordinary child; the small membership correction and actual
refusal test remain pending. Full Stage C acceptance is withheld until that final
source and the independent review are complete. The coder's reported full core
run retains its reproduced CLI human-sink failure; the Ensemble package and full
GUI pass. Exact commands, timings and mutation evidence live in the coder/reviewer
notes rather than being relabeled as coordinator executions.

Historical crossover reconstruction `078cccbd` is accepted after independent `003ae75d`,
Agent tree `f94f414deb977d9f106b8d749a57faa934987049`. The next plan `68793cad` is accepted
by coordinator and independent reviewer `076048cd`: current20 → new18 combined
web/status, including only the necessary earlier GUI settings-lock correction.
Root read the complete plan, exact `05b43a0` lock diff, 23-path historical delta
inventory and predecessor locking sites. The new synchronous model acknowledgment
requires that lock ordering now; memory itself remains deferred. Supported fixture
and unknown-price limits stay explicit. Implementation is released in the isolated
reconstruction branch, with focused causal regression, grader serialization and
independent implementation review before advancement. No whole-tree merge or
new accepted export is authorized at this boundary.

## Next gates

Stage C's complete runtime implementation is accepted at initial formal review,
zero formal correction rounds: source `7feb910a`, independent audit `09e68149`.
Root read the complete final disposition and final changed source/tests, verified
all 14 source-manifest hashes and diff hygiene, and retained its independent
factory-refusal before/after reproduction. The independent frozen-source final
race results are core 5.324s and GUI 4.434s; the coder's full core run retains the
baseline CLI human-sink failure, with final full GUI 8.367s passing. Scope,
ownership, comment and production/test comparisons are in the runtime audit.
Passing local fixtures establishes mechanics, not semantic memory quality.

Stage D is released for concrete plans first. `structured_submission_coder` owns
grader checks and final TL;DR contracts; fresh GPT-6 Astra/high
`compaction_chapter_author` owns narrative/explanations and a source-relative
editorial plan. Both received the exact mandatory skill path and full-context
rules; no chapter or grader implementation is released before their plans agree
and receive review. Runtime reviewer remains independent for that next gate.
Keep old chapter numbers until Stage E, preserve all 11 memory weights/obligations
or explain any deliberate specification replacement, and distinguish turn-ended
from actual background outcomes. The shared fake's locked callback must not
supply a barrier that serializes main traffic.

Main documentation-only changes through `bd4d907c` were merged at clean checkpoint
`a322050b`. Root read the full session-migration record and artifact investigation.
Cadence now records Bill's applied/live-verified session migration at `9d9ef6b8`:
active configuration `/Users/bill/projects/cadence-memories/session/settings.json`.
This supersedes the earlier prepared-only observation, not its historical record.
The separate restored artifact-ID collision is reproduced but unfixed; it adds no
current runtime change or deployment authorization. No private state was read or
migration run by this workstream.

Continue each reconstructed cumulative stage only after its independent acceptance.
No old accepted snapshot can become a new-order result by changing its directory
name. Current20's unchanged 155/155 run is coder evidence on the developing new18
source, not yet implementation acceptance. Chapter prose, frozen exports, live
configuration and publishing artifacts remain unchanged by this workstream.

Subsequent new18 acceptance: independent `387b79b2` accepts source `9af38643`,
Agent tree `6cb2eb95e7a7ad5d5fe9afc8ca2d4f5a7ad3e99d`, at initial review with
zero correction rounds. Root verified that exact clean tree and read the full
coder submission and independent disposition. The unchanged grader remains
155/155; the reviewer's separate causal GUI race check passes in1.025s. Source
selection, supported-price limits, actual browser evidence and the lock mutant's
blocked cleanup remain explicit in the review. Release a source plan only for
current21/new19 Sandbox from this predecessor; no cumulative source import or
new export is implied.

Stage D plan review also found old22's root-submit instruction conflicts with
the already approved main-chat/delegated-submit behavior. Root inspected the
actual Fleet.Idle root exemption and the exact TL;DR passage, then supplied a
bounded, attributed plan correction to the author and coder. Preserve ordinary
delegated three-wake correction and historical transcripts. Root has reloaded
the complete skill, four governing/craft documents, Chapters14/15 and handoff
after compaction; the independent reviewer separately reloaded all required
documents and read all five affected chapters before the plan review.

The final author plan (SHA-256 `07a70665…`) and grader plan (`ae5c0f15…`)
are now independently accepted in the runtime audit. Coordinator accepts their
agreed ownership, preservation and scored-versus-source-evidence distinctions,
and releases bounded Stage D implementation. The author owns narrative; coder
owns final TL;DRs and graded sections, with shared example wording exchanged
before edits. This is plan acceptance only, with no formal correction round
consumed. Old numbers remain until Stage E. Full independent prose/grader review
still follows the actual implementation, examples and targeted checks.

## Stage F live checklist prepared during Stage D

Coordinator preparation only, October14. No model request, credential read, live
configuration change or deployment is performed by this checklist. Execution
follows the chapter/sequence gates and independent checklist review. Use the
accepted runtime then, recording its actual source tree and binary hashes before
launch. Earlier live scripts are historical examples, not safe defaults: the old
memory launcher selects excluded Haiku5.5, treats a changed log as turn completion,
and can close input before a background child finishes. Do not reuse those rules.

Run one isolated human-interface session with GPT-6 Astra/high through the
existing read-only Codex credential adapter. Select the route explicitly and
verify the actual model, effort, endpoint and visible-explanation policy from
sanitized request metadata. Use only the needed existing credential path, never
copy the credential file or print its values. Do not change the active Cadence
profile. Give the test its own temporary workspace, settings/save/log/data paths,
port and GUI process; its browser must never connect to the live application.
Retain the effective memory budgets/watermarks and UTF-8 byte sizes of the selected
source, submitted memory and committed file alongside those safe run settings.
Use a small primary skill and only synthetic public case text, preserving the
normal visible explanation before tool calls. No production conversation enters
the test or evidence.

| Action | Receipt and honest limit |
| --- | --- |
| Analyze the chapter's Cedar Lantern case and request one completed-work checkpoint | Retain the exact public input, visible explanation, checkpoint fact and complete selected source. Set a small conversation budget and a large Session budget so this bounded run requests one conversation compression, without intentionally triggering a graduation cascade. |
| Observe the real child and submit an unrelated ordinary question while it runs | Record monotonic launch, main-input admission, main completion and child completion times from actual events. A screenshot/text observation identifies the child and owner. Main completion before child completion demonstrates progress in this run; a child finishing too quickly leaves this live overlap unestablished, without erasing deterministic barrier evidence or spending another run merely for a better story. |
| Inspect the worker's completion and parent acceptance | Verify the exact declared memory schema, dedicated role and source-owner first-person instructions, visible explanation before submit, actual structured payload, terminal outcome and committed bytes. Compare decision, reason for rejection, owner/date and unresolved attachment question with the original case. Record omissions or inventions instead of treating valid JSON as memory quality. |
| Select the child in the browser before dismissal, when observable | Observe completion/save followed by removal from the tree and selected-conversation fallback to the owner. Retain durable outcome and trace identity. If timing prevents selection, report that limit and retain the separate deterministic GUI result; do not claim an unobserved transition. |
| Ask the owner to recall the supplied case without tools after committed outcome | Inspect actual request memory as well as the reply. Distinguish correct delivery, preservation and model use. Record owner usage before and after auxiliary-only completion, proving main Last is unchanged across that boundary; keep the subsequent recall turn's new main usage separate. Counts and cache grants are observations, not an invoice. |
| Close normally after the terminal outcome and reopen the isolated save without a prompt | Verify process exit, retained memory/outcome, no respawn or new inference at startup, and clean shutdown. Never use main turn-ended alone as the compaction-completion signal. |

Before paid execution, validate changed launch/capture code against the existing
local provider fixture and its failure controls. Check complete source/binary/path
bindings and that empty or partial receipts cannot pass. Build the GUI first and
supervise that binary directly. The launcher has a bounded observation budget and
explicit cancellation/cleanup on exhaustion; this is not a new runtime HTTP
deadline. Keep partial/failure receipts and do not retry automatically. Do not
collect or commit hidden reasoning, raw credential-bearing traffic or the private
profile. Retain only synthetic visible dialogue, relevant memory bytes and safe
metadata needed for the stated observations. The final review decides whether
any uncovered live requirement needs further action.


## October 14: Sandbox accepted; narrative checkpoint reviewed next

Independent review `40372118` accepts reconstructed current21/new19 source
`922caeb4be404c335d209348a9d36aca4f88645e`, Agent tree
`06c7edc9489eff9cf1bf071a8b6cccba303fcc8b`, at initial review with zero correction
rounds. Root read the full disposition and verified the review's one-file commit.
The unchanged seven-row grader scores100/100; actual core/GUI checks and two
restored-defect failures retain coder provenance. No duplicate root execution
is claimed. Release the concrete source plan for combined current22/new20,
including structured submissions at child construction and the applicable
main-chat/delegated-submit policy. Later memory/recall remain deferred.

Author checkpoint `e4b7480c` contains narrative changes to current15/16/22/23 and
its existing note. Original stories and historical live results remain; final
TL;DR and graded contracts deliberately await the coder. Independent reviewer
`compaction_runtime_audit` now reads the narrative while grader implementation
continues. Full combined Stage D acceptance follows the completed contracts,
checks and examples, without an extra per-slice approval gate. Root's interim
correction preserves recall of retired originals: staged recovery protects
committed corpus consistency, not an invented ban on returning retired source.
Private staging filenames are labeled implementation examples.

The grader's first synchronization checkpoint is `1acfd47e`. It recognizes
correlated terminal outcomes, incomplete journal tails and pending cascades;
main turn completion alone is no longer a compression barrier. Focused race
checks and targeted defects are recorded in the coder's plan. This is not yet
a full revised memory score. The next actual-binary fixture holds a child
response while exercising further main turns and observes acceptance/cleanup.

After coordinator compaction, the entire skill and all four governing/craft
documents were reloaded, followed by full Chapters14/15 and the handoff.
Truncated carryover/handoff output was recovered in separate reads before
formal review resumed. No summary substituted for the required source.


## Stage E preparation: references and executable identity

Coordinator source inspection only; no renumbering or export is performed yet.
The current command dispatch (`cmd/grade-edition2/main.go`) selects GUI for
current14–20 except ungraded19, core for21–23 and the dedicated Gateway command
for26. Under the approved mapping, GUI remains14,15,16,18,22,23; core becomes
19,20,21;17 stays ungraded;26 retains its separate command. Update help strings,
Makefile grade targets, default solution paths and report titles with those
semantics. Memory still invokes Chapter14 parity. Original rubric IDs, course
model IDs and immutable historical evidence keep their identities.

The chapter index, solutions table/manifest links and map need the current
sequence, while narrative references require contextual review. Invoice's
current `Extend solutions/ch16` must become new14, not a mechanical mapping to
new23. Historical original-numbered constructor lists in current20 and the
Chapter19 grader/bisect story in current22 remain original-source references;
label that provenance where needed. Current Chapter0 and epilogue references
to the bounded crossover task move to new17, while Bill's preface remains his
own. A preliminary121-line reference inventory lives at
`/tmp/ensemble-stage-e-reference-inventory.txt`; regenerate after final D edits.
It is an editing aid, not proof that every reference is covered.

`scripts/assemble-book.sh` already consumes sorted chapter filenames; verify
its inputs after the move, without rebuilding the committed assembled Markdown,
PDF or EPUB automatically. Existing export tooling accepts committed `agent/`
source from either worktree through an explicit repository/revision and writes
to absent destinations. Preserve exact source identities and old immutable tags;
only promote freshly validated reconstructed exports. Current24/26 still need
the affected carry-forward and validation. No accepted25 export is introduced.


## October 14: structured graders reviewed; live launcher rehearsed

Reconstruction plan `53958bcc` passed root and independent review `ab511273`.
New20 implementation is released. Actual compilation/execution exposed two
dependencies missed by both plan reviews: `df38abe0`'s after-tool interruption
branch/test belongs with original26/new21 (`55cbef23`), and the successful
`turn_ended` dispatch fact requires the matching zero-payload validation guard
from `450f7a9`. Independent corrections `3257a817` and `924ba4fe` record the
precise staging, restoration obligation and actual initial failures. The earlier
memory-only course-row assertion remains deferred until new22. These are
source-plan corrections, not acceptance of the unfinished new20 implementation.

Grader checkpoints D2 `2b17971e`, D3 `5f7bc621`, D4 `5391d5eb` and D5
`8430d4b6` follow D1 `1acfd47e`. They exercise a genuinely held memory worker,
main-chat/delegated-submit compatibility, nested typed submissions and actual GUI
retrieval/reload. Coder reports current22 200/200 and current23 100/100 on their
respective checkpoint sources, with focused race/mutation evidence in the
existing grader plan. Full memory/recall and chapter contracts remain pending.
Independent reviewer resumed on these coherent slices and identified a build-
failure scoring attribution gap in23; correction is underway. No full Stage D
acceptance is claimed.

Root reloaded the entire coding skill, workflow, carryover, voice and chapter
procedure after compaction, then all current14/15 and the handoff before this
work. Root prepared `scripts/edition2/background-memory-live.mjs` for the
accepted Stage F checklist. It builds committed Agent source, binds the exact
source tree and binary, uses an isolated file-configured GUI/Chrome instance
and scratch paths, and defaults to an external HTTP fake. The live mode is
explicit; no live request has been made. Raw provider logs stay in private
scratch. The retained receipt permits only synthetic visible text, schema,
source memory, usage and lifecycle metadata; opaque/provider reasoning parts
are excluded. The tree-only screenshot shows the actual isolated interface.

Final local preflight at source `924ba4fe`, Agent tree
`ec61c01d3f9841364843c1b7e1998d9c573790a4`, passed with launcher SHA-256
`2b3198ce220b3781e968583d0835af44e3dfa442b8e1bb14c19128130326cd9b`.
Command: `node scripts/edition2/background-memory-live.mjs REPOSITORY preflight`
(the executed identical script was still at `/tmp/ensemble-background-live.mjs`).
[Sanitized receipt](../../solutions/evidence/background-compaction/preflight.json)
and [actual tree](../../solutions/evidence/background-compaction/preflight-running-tree.png)
retain the exact source/binary/launcher identity, scratch location and timings.

The real main actor answered the unrelated question while the external worker
response remained held. Its selected child then committed and disappeared, with
selection returning to the owner. The selected source was2034 UTF-8 bytes;
submitted memory and actual file were847 bytes each. This is fixture propagation,
not a semantic compression measurement. Auxiliary-only completion added20
uncached input,11 cached input and17 output tokens once; latest-main counts
remained60,41 and7. Recall received the exact committed memory. Both application
exits were0; restart without a prompt made zero inference requests. Seven
negative controls refused empty/partial receipts, missing/wrong source identity,
wrong binary identity, missing worker input, changed committed bytes and absent
restart completion. No provider credentials were read by the launcher itself.

The initial successful rehearsal at scratch suffix `mFnZxl` sampled the child
title before selection acknowledgment, so it could not establish actual selected-
child fallback. Root noticed the title was still main. The corrected launcher
waits for the explicit child title before claiming selection; `lpI9zh` verifies
that correction, and final `ELW14J` adds the tree capture and smaller receipt.
Those earlier receipts remain in their scratch directories, with their limited
claims preserved here. Live execution still waits for the required sequence
gates and review; no runtime deployment, settings change or publication occurred.


## October 14: final local capture corrections and interim review

Independent interim review `2355ad58` closes the Chapter23 build-failure
scoring correction `ce43d75e` and checks the launcher against its checklist.
The initial source guard ignored untracked Agent inputs. It now rejects them
before and after building: a temporary untracked Go source was rejected at
the intended pre-build assertion (exit1), then removed. The raw negative
control is `/tmp/ensemble-live-source-binding-negative.txt`. No paid run was
used. Review `e9fd0676` separately accepts the reconstructed speech/MCP submit
identity correction; full new20 and combined Stage D acceptance remain pending.

The latest actual preflight is scratch suffix `C8xXXZ`, source `ce43d75e`,
unchanged Agent tree `ec61c01d3f9841364843c1b7e1998d9c573790a4`, launcher SHA-256
`1c7986c22a62ca9011ef28ca83ab984e23bb4ef40b9a2347e49d0dfee2aaad3b`.
The retained receipt and tree image now contain that run; the previous
`ELW14J` artifacts remain in commit `6fa42fec` and their scratch directory.
Capture now records the actual `message_received`, `request_sent` and
`response_ended` event types, replacing nonexistent prompt/turn-end journal
names. The owner admitted the second input at sequence14 after worker launch11
and answered before committed outcome19. Selection/fallback, source2034 bytes,
submitted/file847 bytes, auxiliary accounting, both normal exits and restart
without inference still pass. Seven receipt negative controls still reject
their intended omissions or mismatches.

Review also clarified configuration provenance. These receipts retain launch
configuration, not a GUI readback of effective settings. The omitted conversation
watermarks resolve to high1200/low600 bytes in `internal/memory.Settings`, from
its600-byte budget. Session values are explicitly high200000/low100000 with
budget100000; other bands and recall are disabled. The launcher now labels that
distinction for future receipts. This last metadata-only label has syntax/source
validation, without claiming another application rehearsal or changing the
retained run's launcher hash. No live inference, deployment or publication has
occurred; the sequence gates still precede Stage F.


## October 14: submit identity carried into the full runtime

Reconstruction coder carried only the independently reviewed tool-identity
correction (`e9fd0676`, reconstructed source `f3009a05`) onto primary
`80434965`, predecessor Agent tree `ec61c01d3f9841364843c1b7e1998d9c573790a4`.
This is maintenance with reviewed-source exposure. No cumulative reconstructed
Agent tree was merged. The three production regions install fleet definitions
at composition, bind the immutable schema to the actual compiled submit handler,
and uniformly copy stored declarations instead of classifying tools by name.
Real MCP/custom submit retains its own contract and invocation, including when
a public factory is enabled after construction. Compiled tools retain precedence.

The current-runtime adaptation preserves restricted compaction construction:
its normalized configuration disables ordinary builtins; its internal worker
gets only compiled submit even without public delegation; its early admission
filter remains. Default/custom factories, SafeMode, memory/recall, management and
Gateway behavior are retained. The focused regression is copied unchanged from
the reviewed reconstruction; no additional public API or origin framework exists.
Root independently inspected these production paths before checkpoint.

All commands strip provider/application environment and use temporary EN_HOME.
Raw outputs below are in `/tmp/ensemble-submit-identity-checks/`:

| Check | Actual result (launcher wall seconds) |
| --- | --- |
| `baseline.txt`: new identity regression on unchanged production | FAIL1.342s: missing MCP declaration, custom admission refused, and late-factory MCP schema replaced. |
| `focused-race.txt`: identity/structured plus construction, safety, finite execution, factory, background/lifetime/accounting controls | PASS3.725s (`ensemble/ensemble`1.799s). |
| `core-library.txt`: `go test -short -count=1 -timeout 120s ./ensemble ./internal/...` | PASS32.959s, including memory/recall and Gateway paths; this is not a root CLI suite claim. |
| `gui.txt`: separate GUI `go test -short -count=1 -timeout 120s ./...` | PASS9.289s. |
| `core-vet.txt`, `gui-vet.txt`: `go vet ./...` in each module | PASS0.325s /0.266s. |
| `browser-speech.txt`: `node --test speech.test.mjs` | PASS3.473s, real browser/Agent and blind capability path. |
| `mutant-worker-registration.txt`: remove internal-purpose alternative from composition guard in isolated scratch | Intended FAIL0.989s: both no-public-factory cases report `worker tools=[]`; trusted custom-factory controls still pass. |

Baseline failures are preserved, not called setup failures. The two reviewed
reconstruction restored-name-gate/schema mutants retain their original evidence
and are reused for unchanged behavior; the additional primary mutation checks
the actual memory-only adaptation. Scratch mutant removed. Formatting/diff checks
pass. Existing separately recorded root CLI failure remains unchanged and is not
relabeled by these passing library tests. No grader, paid run or deployment was
performed; the grader coder held application builds during this source change.

Comparing against primary80434965, production changes total0 physical/-3 active/
+3 comment-only Go lines; the focused regression adds140/129/10. Active excludes
blank and `//`-only lines. These are a small composition correction and a real
external-peer regression, preserving owner contracts and explaining binding and
worker lifetime without introducing new machinery. Exact changed source hashes:

- `agent/ensemble/ensemble.go`: `14eccb60037f15ad4ac14dd53492bdc3dbfc9e21ee5b499e9aefdb85abce15af`.

- `agent/internal/agent/agent.go`: `b39f9c60c4a598791813c8a49af047f1b6fa8b53d38d4706b93d1bacdcade19a`.

- `agent/internal/tools/tools.go`: `3c66f33f2cc95a5aac52d5d92d635e267f2f3f6400ac163f0a74dfe324d9ae44`.

- `agent/ensemble/submission_test.go`: `6b0a245cebc69d67115189ca680f430734cd7a2f0d133a61a0fb5b49139250c4`.

This checkpoint awaits its own final carry-forward review; reconstructed new20
acceptance is separate. Mandatory handoff: `.agents/skills/ensemble-coding/SKILL.md`;
the complete skill and four governing documents were reloaded after compaction.


## October 14: full-runtime identity correction accepted; new21 planning released

Root accepts the scoped carry-forward `62c1e781`, Agent tree
`4871633ee3e63dc6aa2cbc0a853f6ef22b4ddfa3`, after independently inspecting the
three production diffs, the real-MCP/custom/compiled regression, and all eight
actual raw outputs listed above. The compaction configuration disables ordinary
builtins before tool installation; internal workers still receive only compiled
submit without enabling public fleet tools, and the early execution filter remains.
Source/result binding and child-owned schema copies agree. No further correction
is required. This review reuses the coder's actual checks; root did not rerun them.
The existing root CLI limitation is separate from the passing library suite.

Independent full new20 review `2a9e7ed9` accepts reconstruction `b4708839`, Agent
tree `39449051049e6d7325b4c368e627739f97eab28f`, after one documentation-only
correction round. The README now distinguishes main chat from delegated submission;
production and tests remain those reviewed at `f3009a05`. Root read the complete
comparative disposition and releases new21 source planning from that exact
predecessor. Restore the deferred AfterTool explanation branch/test with the actual
management API; memory and recall stay deferred to new22/23. Final revised grading
and replacement exports still belong to Stage E.

Root reloaded all four complete governing/craft documents and the mandatory skill,
then read complete current14/15/16/22/23 and the handoff. That prose comparison
found current22 §22.6 teaching default child construction before its implementation
exists in new20. The corrected paragraph uses the actual application-supplied
factory; the later default-construction capability remains in the memory chapter.
The fleet demonstration now ends after collected child results and the main chat
answer. Independent scoped consistency review in `2a9e7ed9` accepts both changes;
this is not full Stage D acceptance. Chapter22 prose lint has no hard failures
(7140 prose words); its long mechanism/person-gap soft warning remains. No code
fence changed and no runtime rerun is needed for these two prose corrections.

## October 14: Stage D and new21 accepted; first live run exposes a lost acknowledgment

Independent review `8935f486` accepts Stage D (grader/contracts) with no unresolved
findings; `d085bd63` accepts reconstructed new21 `aae1e230`. Bill requests continuous
automatic progression. New22/23 cumulative assembly and Stage E numbering are in
progress. Historical source/result bindings remain unchanged.

Root ran the reviewed isolated live launcher from `d085bd63`, Agent tree
`0d4f2acc499cbd35717deca22c9d108f7600d849`, against GPT-6 Astra/high via Codex.
It used synthetic Cedar Lantern data, scratch config/data/ports and a separate
browser. No live production settings, files or process were changed.
The preserved `solutions/evidence/background-compaction/live-first-attempt.json`
is **failed**, not a successful full exercise. The main model called handoff
four times and hit the fixture's three-tool-round limit before its first turn
completed. Thus no second human prompt, recall prompt or restart check occurred.

The background worker nevertheless made one valid submission and its owner
committed that exact first-person memory. The separate sanitized
`live-first-attempt-compaction.json` records this partial result without changing
the failed receipt. The memory retains CL-721, Morgan/October 18 with no invented
year, thirty-day retention, both rejected approaches and their reasons, source-ID
and failure-preservation requirements, and unresolved attachment placement. It
distinguishes planned work from completed implementation. It also includes exercise
constraints, so this is useful retention evidence, not an optimum-size claim.

Inspection of the actual next request found the missing fact: clearing tools
and opaque reasoning leaves an assistant announcement and the unlabelled note,
but removes the tool acknowledgment that the checkpoint completed. The note even
repeats the original instruction to call handoff once. That motivates a narrow
repair; it does not prove model causality or guarantee future compliance.

Root and the independent reviewer chose the existing shared `engine.entries`
renderer helper. It prepends a host-owned completion label to a copied rendered
KindHandoff, preserving stored note bytes, context, fingerprints and event logs.
Existing saved handoffs receive the label too. No event, state, retry mechanism
or tool-limit policy was added. The user's suggestion about removing the limit
was discussed; this repair addresses lost completion information instead.

Before the source fix, the Responses/Gemini HTTP regression failed specifically
for the absent marker. An earlier test compile error used a nonexistent public
KindHandoff alias; it was corrected to existing common vocabulary and is not
counted as a behavioral failure. After the fix, focused context/handoff/Responses/
persistence race checks pass (2.401s), affected vet passes, and the final added
Anthropic assertion plus Responses/Gemini race checks pass (1.245s). Assertions
cover one rendered label, unchanged stored note/context, repeated rendering,
full-log and snapshot restoration. The reviewer accepts the small correction.
Final mapped graders and any corrected live exercise retain their own subsequent
source/results; no live success is claimed here.

## October 14: corrected isolated live exercise completes

The single corrected exercise at `57920dd3`, Agent
`9f421ae303b6d63ef8c9734dc3767587dff6340c`, completed successfully; receipt
`solutions/evidence/background-compaction/live.json` retains its exact source,
launcher and executable hashes. An initial sparse-worktree setup invocation had
no checked-out launcher and made no model request; populating that new isolated
worktree corrected the setup. No production binary, settings or process changed.

Five model requests ran at GPT-6 Astra/high: four main requests and one worker
request. The main agent called handoff once and finished the turn. It answered
the arithmetic prompt and later recalled the supplied case. The worker submitted
1257 UTF-8 bytes from2669 selected source bytes; the recorded population, submitted
memory and committed file match exactly. Both GUI exits were normal, and restart
sent zero model requests. No additional paid run is needed to improve the story.

Root manually compared the retained memory and subsequent answer with the public
synthetic case. Both preserve the project/code, owner/date without inventing a
year, thirty-day retention, source identifiers, failure preservation, the two
rejected approaches and their reasons, and unresolved attachment placement.
The memory is in the source owner's first person and distinguishes a design
decision from completed implementation. It includes exercise constraints, so
this demonstrates useful retention rather than optimal compression. The final
main request received the exact committed memory; recall was disabled, making
this a memory-delivery check, not an Auto-Recall retrieval claim.

The worker finished before browser selection and before the second human input.
Therefore this paid run does **not** establish selected-child fallback, unrelated
main progress during a pending worker, or usage changing across the sampled
meters. Those claims retain the separate deterministic held-provider and grader
evidence; the receipt states both timing limitations. Completed workers disappeared
from the tree, and the retained terminal outcome survives restart.

New22 assembly also exposed a test-only compatibility adjustment: the legacy
OpenAI Chat Completions renderer concatenates the new status label with the note
in its cache-marked text block. The exact expected string in `TestCacheMarkers`
now includes that prefix while retaining the stable-tail/volatile exclusions and
marker-count checks. The source still places the checkpoint marker at the same
boundary. The initial failing assertion is retained by the reconstruction coder;
this adjustment does not alter the live executable or weaken marker placement.

Independent review confirms the actual memory/file/submission equality, all named
facts/reasons, one handoff, four main request/response pairs, normal exits and
restart result. Its clarification: the receipt's shorthand "No main answer before
worker outcome" refers to the unrelated second turn. The first main response at
Seq13 does precede the worker outcome at16; second human input arrives at17.
The original receipt remains verbatim. The launcher's future limitation text now
names the second turn explicitly. This wording correction changes no behavior or
past evidence. Root's final cache/handoff focused race check passes1.295s.


## October 14: reconstructed source and deployment handoff reviewed

Independent reviewer compaction_runtime_audit accepts the complete bounded
carry-forwards through Gateway 26, with zero formal correction rounds. The
[runtime review](codex-compaction-runtime-audit.md) records actual source identities,
necessary transport adaptations and reused historical evidence. Corrected 14–21
sources are in the [carry-forward table](codex-handoff-carryforward.md); memory,
recall, optional Codex and Gateway sources are in the
[reconstruction note](codex-reordered-foundations.md). This is reviewed maintenance,
not another cold student run. The missing sparse-worktree precommit hook was
subsequently checked against all eight exact correction diffs, all passing;
that retrospective check does not rewrite the original commit procedure.

Current solutions/ch14–ch24 and ch26 now contain the exact accepted cumulative
sources, with manifests and fresh edition-2-background-chNN-r1 source tags.
The former directories/manifests remain verbatim at
edition-2-before-background-compaction-20261014 (5d15954e), alongside all existing
immutable tags. The [solution index](../../solutions/README.md) follows the new
chapter order and preserves historical score/source distinctions. Final default-path
grades pass: Chapters 14/15/16/19/21/22/23/26 score 100/100, 18 scores 155/155,
and 20 scores 200/200. Root's grader package tests pass in 36.390s and vet passes.
The [final receipt](../../solutions/evidence/background-compaction/reordered-final.json)
retains commands, timing, full reports and source hashes. Independent review
confirms the eight test-path substitutions, all 12 source tags/manifests, all
1,960 installed file hashes/modes and all 155 grader-source hashes. No behavioral
check was changed for adoption.

Root staged `/tmp/ensemble-gui-background-compaction-b4746ac9` from clean primary
b4746ac985bd84a38d59778ecfa6a3c7f3c3bc7e, Agent 4ff207e66df6d49cbdea462df2f6753d421779d8.
SHA-256: db81a99f17f50a13dd05c98fe620af7a2bd92ce49e7cafa58ce46c2564a4325c.
Build and --help exit 0 on Go 1.25.4 darwin/arm64. It is an optional deployment
handoff artifact, not the earlier live exercise's executable. Nothing was installed
or restarted; no settings or production memory changed. The successful live
receipt and its failed predecessor remain separate, including the absence of
live unrelated-turn overlap, selected-child fallback and sampled accounting-delta
proof. Existing deterministic evidence establishes those mechanisms.
No publication artifact was regenerated, no preface text changed, and no push or
upload occurred. Publication and production deployment remain separate decisions.

## October 14: authorized deployment and publishing rebuild complete

Bill stopped Ensemble and authorized the remaining work. Root rebuilt the GUI
from main `274d3a49` (Agent tree `4ff207e66df6d49cbdea462df2f6753d421779d8`),
backed up the installed binary and three saved-state files, and installed it at
`/Users/bill/ensemble-edition2-codex/bin/ensemble-gui`. The existing launcher still
uses `/Users/bill/projects/cadence-memories/session/settings.json`; configuration
and launcher hashes are unchanged. Astra/high and the Codex route are preserved.

The first detached launch exited normally on closed terminal input. Starting via
Terminal succeeded, with a running process and HTTP 200 at localhost:8088. No chat
prompt or new paid compaction exercise was submitted. This is installed/startup
verification, separate from the earlier isolated semantic-memory evidence.
Rollback material is local under
`/Users/bill/ensemble-edition2-codex/backups/before-background-compaction-20261014-191451/`.
Stop the application before replacing a binary. The saved-state backup is for
recovery; restoring it after new work would discard later conversation.

All five publishing artifacts were regenerated. Independent review accepted the
577-page interior, 578-page reading PDF, chapter order, EPUB references and
representative pages at the retained 7×10, 12pt/10pt settings. Review caught and
verified a repair to the reading PDF's contents links; the print builder now
preserves them when prepending the cover. The preface and chapters are unchanged.
See the [release receipt](../../solutions/evidence/background-compaction/deployment-publication.json)
and [publication record](../edition-2-publication.md). No upload or push occurred.
