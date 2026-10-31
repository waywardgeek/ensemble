# Stage C: restricted background compaction design

Coder: `/root/structured_submission_coder`, 2026-10-14. Design accepted by the
coordinator and independent runtime reviewer on October 14; implementation is
released in coherent slices and remains subject to separate review. Stage B is
independently accepted at `a0acfb5e`; design began at `e8ddced4` and is checkpointed
after documentation-only `cd798d36`. This is maintenance with prior-answer
exposure, not a blind student run. Mandatory skill:
[`.agents/skills/ensemble-coding/SKILL.md`](../../.agents/skills/ensemble-coding/SKILL.md).
The full skill and carryover were reloaded after compaction; this design also
uses the complete workflow, current chapters 14–16 and 22–23, complete handoff,
runtime audit and progress record. No chapter, grader, export or deployment edits.

## Ownership and public seams

Keep the actor, Fleet, Memory, History and Engine as their existing owners.
Shared immutable records/interfaces belong in common; implementation spokes
continue importing common rather than each other. No new scheduler or store.

- A creation-only `AgentSpec.Compaction` purpose selects the restricted worker.
  The ordinary `Spawn` path resets it, as it resets `SubmitSchema`. Root validates
  and normalizes the specification before invoking the custom factory, and the
  normal constructor enforces it before skills, primary prompts, MCP or recall can
  initialize. A public custom factory must return an actually restricted child
  with the requested owners/schema/purpose; the checked constructor refuses a
  mismatch. Safety ceilings still clamp normally. A custom Go factory is trusted
  application code: the root can normalize its input and verify its returned child,
  but cannot prevent arbitrary side effects performed inside that callback before
  return. Test the normal constructor's pre-effect restriction and reject a
  nonconforming returned child; do not claim the callback itself is sandboxed.
- `Fleet.StartCompaction(id, task)` constructs one fresh child and returns its
  actual Agent and reliable `Result` channel. It shares ordinary child creation,
  registration, paths and observation, but does not discard Result or reload a
  previous child task. A random path-safe task ID also names the child. Source,
  policy and commit metadata remain with the parent, never model-authored.
- Root supplies its ordinary default factory when no custom factory is set;
  this does not set the public `FleetEnabled` flag. `EnableFleet` continues to
  enable public delegation and preserves an already configured custom factory.
  Internal construction must work without granting delegation to headless or
  `BuiltinTools=false` parents. `BuiltinTools` controls default registration, not
  an inherited security ceiling: register only submit in this special child while
  leaving the parent's declarations and FleetEnabled unchanged. Preserve actual
  AgentSpec safety ceilings (SafeMode, web and sandbox narrowing). If those ceilings
  prevent required child construction or trace persistence, fail explicitly and
  retain source; do not widen them to make internal delegation work.
- Memory extends its existing selection value with selected-entry identities,
  relevant policy and file/destination fingerprints. Its narrow new operations
  are `ValidateSelection`, `Stage`, `Finish` and `Recover`; they own corpus I/O.
  History projects correlated facts. Agent owns when facts may be appended.
- Fleet dismissal stops/saves an exact child identity and removes that identity
  from its map and root registry. Root issues a removal observation; GUI handles
  it before creating a view, drops the retained view, and switches selected tabs
  to the owner with the ordinary replay path. Directory allocation uses a root
  monotonic counter, rather than the now-shrinking registry length.

## Restricted construction and finite execution

The child inherits provider/model/route and safety settings, not the parent's
role or conversation. Before any startup side effect, clear primary/startup
skills and variables, bypass skill/MCP initialization, disable recall and all
memory bands, and bypass context forcing. Give it only the normal submit tool,
with Stage B's exact whole-argument schema:

```json
{"type":"object","properties":{"memory":{"type":"string"}},"required":["memory"],"additionalProperties":false}
```

Both declaration and actual execution admission accept only `submit`, including
guessed names and later registration attempts. These restrictions are immutable
for this purpose; settings cannot reopen tools/recall/memory. Internal persistence
still uses the normal sandbox and does not become a model filesystem capability.

The dedicated prompt asks for the **source owner's first-person memory**, retaining
decisions and reasons, failed approaches, uncertainty, constraints and preferences.
The task message contains delimited source data with role labels and handoff notes;
quoted instructions remain data. There is no main coding/no-submit prompt, cloned
dialogue or opaque model state. Visible explanations precede submit where current
model policy requires them; they stay outside the memory string.

One attempt permits at most **three Engine.Send invocations and three tool batches**:
the first request plus two opportunities to correct validation/visible-explanation
errors. Count invocations immediately before actual sends, including ephemera-induced
resends, across the worker's entire lifetime. Do not use ordinary missing-submit
idle wakes. A completed prose-only response fails immediately. More than one call
in a response, or a call other than submit, rejects that batch with paired errors;
no first-call partial success. A successful submit finishes after its tool-return
tail. The purpose admits one task only; messages cannot start fresh turns or reset
its budget. An invocation can include the existing bounded authentication refresh
and HTTP retry; three invocations is not a claim of three physical HTTP exchanges.
Keep that separate retry bound and test that it cannot restart the worker budget.
Failures do not automatically start another attempt on the same source.
There is no total model deadline; explicit cancellation reaches the provider.

For worker requests only, omit the optional Chat `max_completion_tokens` and
Gemini `maxOutputTokens` caps, preserving existing Responses route behavior.
Their APIs explicitly make these fields optional; Gemini's default varies by model.
Anthropic requires `max_tokens`: request **16,384** for workers, a practical allowance
over twice the observed 7,317-token failed compression, rather than the current
1,024/4,096 chat allowance. This is an execution budget, not a claimed universal
model maximum or a conversion from file bytes. Sonnet-5's documented output capacity
exceeds it. A provider/model rejecting that allowance fails visibly with source
retained; there is no silent model substitution, cap reduction or beta feature.
Ordinary requests keep their existing settings. For workers, adapters must preserve
and check actual terminal reasons in both streamed and nonstreamed responses:
Anthropic `max_tokens`, Chat `length`, Gemini `MAX_TOKENS`, and Responses incomplete
output are failures even when the returned submit arguments are syntactically valid.
Propagate the failure before any submit tool dispatch; still retain final measured
usage. Stage B's payload bounds also apply. No byte truncation forces a ratio.
Test a memory over the old caps, malformed JSON, and a valid submit paired with a
length stop; live quality/sizing evidence remains Stage F.

Sources checked October 14: [Anthropic Messages](https://platform.claude.com/docs/en/api/messages/create),
[Sonnet-5](https://platform.claude.com/docs/en/models/sonnet-5/overview),
[Chat Completions](https://developers.openai.com/api/reference/resources/chat/subresources/completions/methods/create),
[Gemini GenerationConfig](https://ai.google.dev/api/generate-content#generationconfig).
These establish parameter contracts, not new model support; Bill's excluded 5.5
models remain excluded. Existing synthetic course fixtures are not live evidence.

## Parent lifetime, scheduling and acceptance

The parent runtime holds one private attempt slot, independent of its ordinary
turn: ID, immutable selection, generation, child, cancellation state, and eventual
completion. Launch uses a background owner context, not `active.ctx`. Record launch
correlation, source boundaries and child trace path. Main rendering continues after
checkpoint cleanup while the child runs. Repeated handoffs cannot overlap this slot.

Ordinary turn finish and parent interruption preserve the child. Explicit child
kill, owner shutdown or a relevant memory policy change invalidate the attempt and
cancel/drain it. Compare source/destination band settings at the actor's existing
settings transition and increment their generation on every relevant change;
disable/re-enable cannot restore an earlier generation. Unrelated UI settings and
new dialogue do not invalidate it. Worker replies carry immutable task identity;
late replies cannot adopt a newer slot. Shutdown drains the managed completion
before the parent's final save/exit, retaining accounting and failure evidence.

Keep existing 90% warning and 95% forcing arithmetic based solely on last completed
main input, with cache buckets counted once. If an attempt is outstanding and that
last main request reached 95% of a known context window, defer the next main render
until this attempt resolves; show waiting activity and keep interruption/settings
usable. Successful compaction resumes it. If that held attempt fails, end that main
turn with the explicit compaction/headroom error rather than immediately sending the
same full context. Unknown context windows retain the ordinary provider error path.
This is the documented near-window exception to background progress, not a timer.

Completion becomes a candidate only after the child's reliable Result, full
shutdown/drain, and required Save succeed. `Status.Submitted`/Join alone is insufficient.
The completion worker copies the validated payload and final usage, then posts one
result through the parent's existing mailbox. Commit only when idle or at next-render
with the tool batch complete, never while an already-sent response is in flight.

At that boundary, Memory revalidates exactly the selected source and destination:

- Conversation selection uses whole eligible completed checkpoint segments and
  exact selected entry/part identity; independent skills/declarations, unfinished
  tail, and newly appended dialogue are outside the hash and outside the cut.
- Graduation validates selected entries against current file bytes, not only IDs
  or sizes. Hash file bytes and coverage metadata; same-name human edits invalidate.
  Validate the destination's prior bytes/absence too, including existing MEMORY.md.
- Scan rejects malformed or reversed coverage ranges. Graduation also verifies its
  retirement prefix equals the selected set and makes progress; overlapping ranges
  cannot silently retire an unselected file or repeatedly select retired-nothing work.
- Recheck relevant settings/generation. Reject empty/whitespace memory and stale
  candidates explicitly. Existing filenames are not overwritten on collision.

After a successful commit, policy may select the next graduation, one child at a
time. Do not restart failed work without a later checkpoint. A graduation must
retire its selected nonempty source set before another cascade step starts.

## Durable replacement and restart protocol

Retain paired cut/retirement and population facts. The identified baseline defects
require a hidden durable candidate and a durable commit receipt, not a new database.
Memory owns one atomic `.compaction/pending.json` record per parent, outside all
active/recall scans. It includes version, task ID, selected file paths/hashes,
destination precondition, candidate bytes/hash and population identity. All paths
derive from the owner and use its Sandbox; the model supplies none of them.

Extend the existing PendingCompression projection with TaskID and a committed
population receipt. Legacy uncorrelated events keep their current replay behavior.
For new tasks, the matching population atomically applies the held cut/retirement
and retains the receipt until disk finalization. A correlated terminal outcome
clears it and keeps only the latest task outcome in Context, so an orphan staging
record can be recognized after a crash. Do not launch another attempt until this
record has been removed successfully. This bounded receipt survives snapshots
and log retention; it is not an ever-growing job ledger. Outcome facts do not change
ordinary turn state (unlike `error_occurred`).

Before launch, establish the parent's existing snapshot/journal if the embedding
client has not called Load. Make successful History.Checkpoint establish its journal
path under the same lock; Agent.Save then supplies this boundary without replaying
or replacing live history. A persistence failure refuses launch. This also makes
plain public NewAgent/Ask memory use durable without demanding hidden client setup.

| Durable boundary | Normal action | Restart before the next action |
| --- | --- | --- |
| Launch journaled | Run fresh worker; source remains active | Mark abandoned; retain source; never respawn |
| Hidden candidate installed | Revalidate selection and append cut/retirement, then matching population | Without matching population, discard hidden candidate and abandon; source stays active |
| Cut/retirement alone journaled | Source is still in projection until matching population | Discard held cut and candidate; source stays active |
| Population journaled | History commits the pair; Memory promotes staged bytes into active destination | Recover committed receipt before any startup population or recall; complete disk work without a model |
| Destination installed | Rename selected originals to `.retired` | Verify destination hash, finish remaining expected source renames idempotently |
| All source renames complete | Journal terminal committed outcome, then remove hidden record | Receipt or latest terminal outcome recognizes installed/retired state; finish only missing steps |
| Terminal outcome journaled | Remove hidden record; release slot | Matching last outcome permits hidden-record cleanup without reapplying history |

`Memory.Recover` runs after History recovery and before **all** startup corpus
population/recall, including re-enable paths. It uses recorded bytes/decisions and
never reconstructs the replacement with an LLM. A committed receipt with missing
or mismatched staging data is an explicit recovery error; do not scan conflicting
active files into context or silently start a new attempt. A cleared staging file
with an already terminal outcome is normal. Old pending attempts retain legacy
abandonment behavior; no claim is made to repair every old ambiguous disk artifact.

Stage before appending the pair; on a pre-commit persistence failure, source remains
unchanged. After journal commit, finalization errors leave the receipt and candidate
available for retry/restart, block further compaction, corpus scans, recall and main
rendering, and report the error. Startup and re-enable must pass recovery before
any of those operations. If finalization cannot finish, return the explicit error
to the waiting/new main turn; remain interruptible and retain the recovery hold
until a successful recovery. Never send a mixture of committed projection and
unfinalized active corpus. Do not call a committed replacement an abandoned attempt. Idempotent
Finish accepts already-installed matching destination and already-retired matching
source, but never overwrites a human edit or an unrelated `.retired` file. Recheck
file preconditions immediately before each mutation. An external edit racing these
steps can require explicit conflict repair: preserve both copies and the receipt,
report failure, and stop automatic finalization. Portable filesystem operations do
not provide a transaction against arbitrary concurrent external editors. This
continues the repository's process-crash, not power-loss/fsync, durability contract.

## Completion, dismissal and usage

Capture the reliable Result, then call the existing child Shutdown and Save before
acceptance/removal. Shutdown joins actual actor/job/provider workers; even canceled
children can account usage after an interrupt Result. Do not cancel from the submit
callback. Save failures prevent candidate acceptance, retain trace files, and become
parent outcome errors. Fleet cleanup uses exact identity and a shared completion
for concurrent manual kill/automatic dismissal, so it cannot delete a replacement
child with the same name or save/remove a still-running child.

Remove Fleet/root/GUI references only after drain and required save attempt, even
on failure. Preserve journal, submit row and trace path on disk. No late observation
may recreate a dismissed GUI view. Parent outcome records completion, cancellation,
invalidation, persistence and cleanup errors; repeated dismissal is harmless.

Engine exposes a copied breakdown of its existing usage buckets (vendor, model,
large-price tier, disjoint usage) and an auxiliary addition operation. The actor's
single active attempt owns an `accounted` bit: accept completion only for that exact
attempt, add its final buckets once, then set the bit before releasing the slot.
The shared child completion handles racing kill/dismissal; duplicate mailbox
delivery cannot import twice, and a late result cannot match a subsequent attempt.
Engine retains no task-ID deduplication set. Import after actual drain, for successes
and failures, into parent session totals and corresponding model/tier buckets.
Do not alter Engine.Last,
History.LastUsage, parent response facts or latest-main context occupancy. Do not
also add child totals through Fleet aggregation. The outcome retains the usage
receipt for audit; replay does not import historical spend into fresh session
totals. Pricing remains owned by Engine, including unknown-price behavior.

The existing exported `Engine.Compress` method remains deliberately source-compatible
for external callers, with a deprecation comment explaining that it is a direct
legacy operation. The normal actor path never calls it or falls back to it.

## Implementation slices and evidence

First port root's real Memory/History restart-window and reversed-range probes into
focused source regressions, retaining their exact baseline failures and positive
controls. They model saved boundaries, not a forced process kill. Then implement
coherent reviewable pieces: restricted construction/finite execution; Memory staged
recovery and correlated projection; owner-lifetime scheduling/completion/accounting;
registry/GUI dismissal. The full protocol above governs every piece.

Use real Ensemble/actor/Engine/Fleet/Memory/History and a small concurrent HTTP peer.
Identify workers by the actual sole submit declaration with a `memory` property.
Root found the old grader classifier too broad and its fake server globally locked
around Route, so a held compressor also blocks main requests there. C tests must
release their request-capture lock before waiting at a provider barrier; graders
remain Stage D work. No mocks, live calls or production GUI connections.

Focused claims: main completes and accepts another turn while worker held; exact
role/source/schema and visible explanation; default/custom/headless construction;
no startup skill/MCP side effects and forged tool admission; finite requests despite
ephemera, prose-only and invalid submit responses; unchanged selected source with
unrelated dialogue; file/destination edits and off/on invalidation; safe tool-batch
commit; all restart boundaries and journal/write failures; no truncated/empty commit;
final submit tail and save failure; interrupt versus kill/owner shutdown; exact usage
model/tier once with unchanged main Last; repeated registry/job/GUI removal including
selected-child fallback; ordinary grandchildren and recall corpus remain compatible.

Use deterministic external barriers and real filesystem refusal cases rather than
timing guesses or mock persistence. Apply focused restored-defect mutations per
distinct claim, recording coupled failures honestly. After coherent focused checks,
run core/GUI tests, race, vet and required explicit-source old-number graders with
root; retain known baseline CLI human-sink/browser-fixture failures separately.
Strip only LLM_/OPENAI_/ANTHROPIC_/GEMINI_/EN_/CODEX_ in test subprocesses. Do not
glob browser scripts or repeatedly rerun unchanged broad suites. Semantic memory
quality, live provider acceptance and deployment remain unclaimed until Stage F.

## C1 implementation evidence: coverage admission

On design checkpoint `01dfa4ce`, source regressions first reproduced all three
root probes: uncommitted active replacement after restart, committed retirement
before archive after restart, and reversed-range no-progress reselection. All
three failed for their intended reasons in 0.015s. The initial filename
`compaction_windows_test.go` was accidentally GOOS-filtered on macOS, causing
undefined test helpers; this was setup failure, corrected before the measured run.
Raw outputs: `/tmp/ensemble-compaction-stagec-baseline{,-corrected}.txt`.

The first narrow runtime slice rejects reversed/malformed filename coverage and
checks graduation's restored entries for invalid, reversed or overlapping ranges.
It protects the prefix retirement rule without changing watermarks or selecting
different valid work. Focused race tests include a valid ranged-file scan, exact
prefix retirement/progress and inherited MemoryWaterfall; the final focused race
run passed in 1.142s; after review strengthened retained-file identity assertions,
the affected race checks passed in 1.143s. Targeted vet passed.
Restoring scan admission admits both reversed and extra-segment filenames and
fails the intended checks; restoring graduation admission fails both reversed and
overlapping restored-entry checks while the valid control passes. These are two
focused mutations, not four independent mutation proofs. A third mutation reverses
the retirement comparison: the surviving count stays one, but the strengthened
check rejects source 1 surviving instead of source 3. Scratch copies removed.
Raw outputs: `/tmp/ensemble-compaction-stagec-ranges{,-final}.txt` and
`/tmp/ensemble-compaction-stagec-mutant-{scan,graduation,prefix}.txt`; reviewed final
race output is `/tmp/ensemble-compaction-stagec-ranges-reviewed.txt`.

The two legacy Write/event/Archive restart probes are still failing and remain
outside this coverage-only checkpoint. New Stage/Finish recovery tests will test
the replacement protocol explicitly; they must not relabel these old-API failures
as unchanged passing regressions. No background scheduling or recovery acceptance
is claimed by C1. Independent source review found no production blocker and asked
for useful test comments and exact retained/selected identities; both were added
before checkpoint. The reviewer accepted C1 for checkpoint, with no production
changes requested; this is not full Stage C acceptance or a formal correction
round. Final slice disposition is recorded in the runtime audit.

## C2 implementation evidence: worker construction and finite execution

Work after `c8ec3d4b` adds creation-only Compaction purpose, normalizes its startup
specification before custom factories and in the normal constructor, replaces the
main role, disables memory/recall/forcing, and registers/admit only submit. Default
child construction no longer requires public delegation; EnableFleet preserves an
existing custom factory. Ordinary Spawn and the public AgentSpec.Child helper reset purpose as well as schema.
Ordinary Spawn still refuses when public delegation is disabled; the internal
default factory does not silently grant that public capability. The actor
admits one task, at most three Send invocations/three batches, rejects entire extra
or forged-call batches, and treats prose-only completion as failure without idle wakes.

Focused real-runtime tests cover default/custom construction with and without
builtins, normalized factory input, primary-startup refusal avoidance, unchanged
parent tools, actual safety ceilings, factory purpose mismatch, valid submission,
invalid schema correction exhaustion, extra/forged calls, a held prose response
with a pending hint, second-task refusal, and ordinary descendant defaults. This
slice's provider is a concurrent Anthropic HTTP fake with faithful usage/call IDs;
provider-specific output sizing and terminal-stop handling remain the next slice.
Tests also run the Stage B structured-submission cases. Final focused race passed
in 1.519s; affected ensemble/agent/fleet vet passed. Inherited Fleet, Named,
Management and sandbox child/safe-mode compatibility race checks passed in 3.228s
(`/tmp/ensemble-compaction-stagec-worker-compatibility.txt`). Initial tests incorrectly
expected forbidden RegisterTool to succeed so later dispatch could reject it;
registration already enforces admission. That fixture error is retained in
`/tmp/ensemble-compaction-stagec-worker-initial.txt`; corrected and race outputs
use the same prefix with `fixture-corrected`, `race-final`, and `vet` suffixes.

Twelve focused scratch mutations failed their intended assertions: skipped startup
normalization, inherited main role, unavailable default factory, overwritten custom
factory, unchecked returned purpose, reopened tool registration, four-send budget,
second task admission, successful prose-only response, accepting the first submit
of a multi-call batch, inherited main forcing, and inherited worker purpose in an
ordinary descendant. The startup mutant's nonexistent-skill failure demonstrates
unexpected startup loading, not an actual MCP connection measurement. Logs are
`/tmp/ensemble-compaction-stagec-worker-mutant-<case>.txt`; scratch copies removed.

C2 does not yet launch these workers from normal memory policy: StartCompaction,
provider sizing/terminal reasons, staged replacement/recovery, owner-lifetime
scheduling, drained completion/accounting and registry/GUI dismissal remain required
Stage C work. The legacy recovery tests still fail as recorded above. Independent
C2 slice review found a missing purpose reset in the public Child helper and weak
schema substring assertions. The helper now resets purpose; a real derived child
proves ordinary construction. Both declarations and every actual input_schema are
compared structurally with the exact memory contract. Two additional focused
mutations remove that reset and allow extra submit properties; both fail the
intended assertions. Raw outputs use the worker-mutant prefix with public-child
and exact-schema suffixes. A C2-only scratch run restores HEAD Engine and omits
C3/recovery tests: final focused race passed in 1.141s. The independent equivalent
run including all Stage B tests passed in 1.526s. Scratch copies removed. These
results exclude in-progress C3 provider changes. No full Stage C acceptance is
claimed. The independent reviewer accepted C2 for scoped checkpoint with no
production blocker and no formal Stage C correction round; its audit records the
size comparison, exact source binding and remaining exclusions.

## C3 implementation evidence: provider allowance and terminal failure

After C2 checkpoint `ce52c166`, worker requests use the accepted Anthropic16,384
allowance and omit optional Chat/Gemini caps; ordinary rendering is unchanged.
All three SSE reconstruction paths preserve actual stop reasons. Semantic parsing
rejects worker length stops before submission dispatch even when arguments parse.
Responses already rejected incomplete output; it now retains complete measured
usage on terminal failure and on malformed arguments in a completed envelope.
Classic semantic failures likewise retain validated measurements. No successful
response fact is fabricated for that accounting, and absent usage is not zero cost.
This slice does not recover missing final measurements from an interrupted stream.

The real-runtime HTTP fixtures cover four routes × streamed/nonstreamed × success/
length-stop, plus Chat and Responses malformed-argument envelopes in both delivery
modes. Success preserves the exact roughly52KB memory string; synthetic9000 output
tokens are provider fixture measurements, not a tokenizer or live quality result.
Failure checks require no submission/tool dispatch and exact measured usage; the
main Agent's usage/occupancy remains unchanged before later managed import. Initial
compile setup used a nonexistent History.Events helper and was corrected to the
public EventsAfter API; the original setup failure remains in
`/tmp/ensemble-compaction-stagec-provider-initial.txt`. Corrected16-case tests
passed0.103s. The final20-case focused race passed1.919s. Inherited streaming,
visible-explanation and engine boundary race checks passed (engine1.019s,
ensemble1.366s); targeted vet passed. Raw files use the same provider prefix with
`corrected`, `race`, `compatibility`, and `vet` suffixes.

Thirteen focused mutations each failed their intended assertion: each of three
worker caps restored; each of three semantic length refusals removed; each of
three streamed terminal reasons discarded; classic failed usage dropped; Responses
terminal failed usage dropped; ordinary Anthropic requests given the worker cap;
and completed malformed Responses usage dropped. The last mutation covers both
delivery modes through their shared semantic path, not two independent proofs.
Logs: `/tmp/ensemble-compaction-stagec-provider-mutant-<case>.txt`; scratch removed.
Existing Responses incomplete-output rejection and unchanged schema validation
reuse their inherited tests; this slice does not claim a new independent mutation
for each inherited assertion. Final combined C2/C3 race passed2.163s. Independent C3-only source reconstruction
(accepted C2 archive plus exactly the C3 diff/test, excluding C4 and recovery)
passed C3+C2 race2.030s. The reviewer accepted C3 for scoped checkpoint with no
blocker and zero formal Stage C rounds; its audit retains the exact source binding.
No parent
lifetime, disk recovery, once-only auxiliary import or live capacity claim is made.

## C4 implementation evidence: selected-source acceptance

Memory selections now capture a digest of the exact selected entries and task
boundaries, effective source/destination policy, selected source-file bytes and
prior destination bytes/absence. ValidateSelection rechecks these values. It does
not hash the growing history: new dialogue, skills and unselected later files
remain independent. Actual source bytes must match the selected context already
at launch, including existing curated MEMORY.md. Nonterminal destination collisions
are refused. Policy generations across off/on transitions remain actor work;
equal current settings alone cannot prove generation continuity.

Real Memory/History/corpus tests cover unchanged selection, unrelated growth,
selected projected-content edits, same-name/same-length source-file edits,
unselected-file edits, changed policy, destination creation, existing destination
collision, and curated-file edits before/after selection. The C1 valid-prefix
fixture now writes actual source files, since selection fingerprints them; invalid
coverage still refuses before filesystem selection. Existing range and waterfall
controls pass. Initial setup misspelled common.SkillData as SkillLoadedData and did
not compile; corrected focused tests passed0.033s, final focused race1.167s and
memory/ensemble vet passed. Outputs:
`/tmp/ensemble-compaction-stagec-selection-{initial,corrected,race,final,vet}.txt`.

Five focused mutations fail intended assertions: ignore selected entry bytes;
ignore relevant policy; permit an existing destination; accept curated file/context
disagreement; and omit selected file-byte validation. The last mutation removes
both file/context comparison and captured-file comparison, explicitly a coupled
restoration of the old unchecked path, not two independent proofs. Raw outputs:
`/tmp/ensemble-compaction-stagec-selection-mutant-{entries,policy,collision,curated,source-bytes}.txt`.
Scratch copies removed. Independent reconstruction from accepted C3 plus the five
C4 paths passed selection/C1/waterfall race1.170s. The reviewer accepted C4 for
scoped checkpoint after a comment correction clarifying that Sources holds
fingerprints, not recoverable file bytes; no behavior correction or formal Stage C
round. No staged replacement, restart-window repair or actor generation/lifetime
claim is made. The runtime audit retains exact source binding and size comparison.

## C5 implementation evidence: correlated projection and journal receipt

After C4 checkpoint `ffa0bb68`, History accepts task-correlated launch, cut/retirement,
replacement and terminal outcome facts. A matching population applies the pair and
retains its committed receipt until a matching committed outcome; abandonment
cannot clear that committed state. An unrelated disk population cannot satisfy a
managed cut. New launches cannot replace an outstanding managed receipt, including
legacy launches. The latest terminal outcome survives snapshots for subsequent
staging-file cleanup; repeated identical-state completion is harmless and does not
change the main turn. Legacy uncorrelated event behavior remains compatible.
Successful Checkpoint now establishes its journal path under the same lock, so a
fresh embedding client calling Save before Load journals subsequent launch/commit
facts. This is a projection/journal slice, not completed disk recovery.

Real public History/Agent tests cover deferred source removal, exact task/file/range/
band correlation, before-cut population refusal, premature success, duplicate
launch/replacement, committed-versus-abandoned state, terminal cleanup identity,
unchanged main-turn state, and actual snapshot+journal recovery without model work.
Graduation retains the exact unselected third file. Initial new tests passed0.013s;
affected memory/context/save compatibility plus new receipt race passed3.062s;
history/ensemble vet passed. Outputs use
`/tmp/ensemble-compaction-stagec-receipt-{initial,race,vet}.txt`.

Eleven focused restored defects fail intended checks: ignored task/file/range/band
identity, wrong terminal commit state, discarded committed receipt, disk population
satisfying a managed pair, early cut, uncleared pending receipt after terminal
outcome, legacy launch overwriting a managed receipt, and missing journal activation
on fresh Save. Logs: `/tmp/ensemble-compaction-stagec-receipt-mutant-<case>.txt`.
Scratch copies removed. Final focused receipt race passed1.042s. Independent
reconstruction from accepted C4 plus the seven C5 paths passed receipt/C4/waterfall
race1.173s; the reviewer accepted C5 for scoped checkpoint with no findings and
zero formal Stage C rounds. Its audit retains exact source binding. The two original unsafe
Write/event/Archive restart probes remain failures until Memory staging/finalization
and runtime recovery gates are implemented; these receipt tests do not relabel them.

## C6 implementation evidence: staged disk recovery and corpus holds

After C5 checkpoint `c00af62b`, Memory owns the hidden versioned staging record,
exact launch/selection/candidate binding, idempotent destination promotion and
selected-source archiving. Finish retains the record; Recover returns a terminal
fact for the actor to journal before a second call removes staging. Committed
receipts refuse missing/corrupt/mismatched records and conflicting human edits.
Startup and memory-settings synchronization recover before scanning; a failed
committed receipt also blocks Memory selection/scan, recall search and rendering
of that Context. Recall construction now defers indexing until its first query,
so NewAgent cannot scan files before its embedding client loads History. This is
the recovery slice; normal actor compaction still uses its legacy path pending
the next scheduling/lifetime slice.

The original direct-Write probes remain recorded failures of the old API. Their
source is preserved at `/tmp/ensemble-compaction-legacy-recovery-before-staging.go`
and the original raw baseline described under C1. The new regressions explicitly
use Stage/Finish and real History snapshot/journal recovery; they do not claim
unchanged legacy probes now pass. Saved-boundary cases cover hidden candidate,
held cut, committed pair, promoted destination, one source archived, all sources
archived and terminal outcome before staging cleanup. These are constructed real
saved states followed by a fresh Agent Load/Ask, not forced process kills.

Eleven damaged-record/file cases prove recovery errors preserve the committed
receipt, prevent requests, block scan/selection/render/recall/re-enable, and permit
retry after an explicit repair. Empty/stale output and a real not-a-directory
staging failure preserve active source. A real unreadable directory proves recall
construction defers indexing while the subsequent query actually reaches that
corpus; this permission test passed, rather than skipping, on this host. Initial
new tests passed0.046s; final focused race passed1.243s. Affected memory/recall/
compaction compatibility race passed ensemble18.578s; internal memory/history had
no package tests and internal recall/engine selected no tests in that command.
Targeted vet passed. Raw outputs use
`/tmp/ensemble-compaction-stagec-staging-{compile,initial,race,final,compatibility,vet}.txt`.

Nineteen focused scratch mutations failed the intended checks: putting the hidden
record in active memory; skipping startup recovery; skipping archiving; ignoring
destination, source or archive conflict; rejecting already-retired matching source;
ignoring selection digest, candidate digest, committed population identity, record
version or unknown fields; bypassing scan, render or recall hold; eager constructor
recall; deleting staging before terminal outcome; accepting empty memory; and
accepting a stale Stage task ID. The active-candidate mutation exposes serialized
record bytes to Scan, sufficient to demonstrate visibility rather than semantic
memory quality. Skip-recovery fails on the required successful restart because the
remaining scan hold refuses access. Logs use
`/tmp/ensemble-compaction-stagec-staging-mutant-<case>.txt`; each scratch tree was
removed. The stale-ID mutation initially matched two source sites and was refused
by the scratch edit assertion before building; a scoped replacement then failed
the intended Stage refusal check. That setup assertion is not a mutation kill.
Malformed JSON uses the standard decoder; write refusal uses the actual filesystem;
inherited journal ordering/selection checks retain C4/C5 evidence rather than
claiming new independent mutations for unchanged behavior.

No model was called live. No power-loss guarantee, transaction against arbitrary
external editors, full worker lifetime/accounting, or complete Stage C acceptance
is claimed. Independent C6 review accepted this scoped slice with no production
correction and zero formal Stage C rounds. Its exact predecessor-plus-eight-path
reconstruction passed the six focused test functions under race in1.262s, with
no skipped startup permission test; the runtime audit retains the source binding.

Independent review requested the distinct existing-MEMORY.md promotion branch.
Added a real staged commit/restart with exact curated replacement and a human
edit refusal preserving bytes/receipt. The final focused race including both
passed1.258s (`staging-reviewed.txt`). Two additional mutations fail the intended
curated checks: skipping an existing destination update leaves old bytes; ignoring
its prior fingerprint overwrites the human edit (`staging-mutant-curated-{replacement,conflict}.txt`).
These bring the C6 total to21 intended mutant failures, with the setup assertion
above kept separate. No production correction was required for these added cases.

## C7 implementation evidence: auxiliary accounting seam

After accepted C6 `59c71d90`, Engine exposes value-copied UsageBuckets and
AddAuxiliaryUsage. The buckets retain vendor, configured model and the original
request's large-price tier; aggregation cannot reconstruct that tier afterward.
Both ordinary accounting and imports use the same locked totals helper. Imports
leave Engine's latest-main measurement and History unchanged, with no task-ID
ledger. The actor's forthcoming active-attempt bit remains responsible for calling
once after drain; this seam deliberately does not claim duplicate-delivery safety.

A real main actor and restricted child use an HTTP peer. The child corrects one
schema-invalid response, crossing the existing small/large price threshold; both
responses include cache-write/read counts. Cases cover successful completion,
measured length failure and an unpriced supported model. After actual child
Shutdown/Save, a manual import preserves exact original buckets, totals and cost,
leaves latest-main/History occupancy untouched, and does not reimport spend after
parent Save/Load. Receipt mutation cannot affect either Engine. The returned type
contains only value fields; that copy property also follows directly from its
representation, without adding artificial pointer state for a mutation test.

Initial test execution failed one overly strict whole-meter equality assertion:
map iteration changes floating-point addition order. The corrected assertion
compares exact counts/identity and uses a small tolerance only for derived money.
The original failure is retained in
`/tmp/ensemble-compaction-stagec-accounting-initial.txt`. Affected provider/cache/
status accounting race passed2.057s and targeted vet passed. After strengthening
exact imported-bucket assertions, final focused race passed1.086s. Seven targeted
mutations fail intended assertions: omit import, lose original tier, substitute
parent model or vendor, overwrite latest-main usage, omit copied input/cache counts,
and add each imported bucket twice. Logs use the accounting prefix with `race`,
`vet`, `final` and `mutant-<case>` suffixes; scratch copies removed. Duplicate
addition inside this single operation is distinct from future duplicate mailbox
handling. Existing reload/session and provider usage semantics retain inherited
coverage; no new live pricing or semantic memory claim is made.

Independent C7 review accepted the scoped slice with no production finding and
zero formal Stage C rounds. A slice-wide copy assertion replaces the earlier
map-order-dependent first-element check; its focused race passed1.084s. The
independent exact-predecessor-plus-five-path accounting/C3 race passed1.989s;
the runtime audit retains the source binding. Managed launch, drained completion/dismissal,
owner-lifetime scheduling, generation/headroom and once-only actor import remain
required before full Stage C acceptance.

## C8 concrete lifetime refinement before implementation

StartCompaction returns the actual Agent and one buffered ChildCompletion channel.
The envelope retains the normal reliable Result, copied Submission and final usage
buckets, an explicit manual-cancellation fact, and shutdown/save errors separately
from the task error. One launch goroutine consumes Submit's reliable Result, then
uses the same child entry's once-only stop/save operation as manual ShutdownChild.
It captures payload/counts only after drain and delivers the envelope afterward.
A racing manual kill marks cancellation under the manager lock before sharing that
operation, even if the successful submit Result has already arrived. The parent
actor will reject cancellation and every task/cleanup error, but still account the
final buckets. It must count its completion receiver until consuming this envelope,
including during owner shutdown, before the parent's final Save.

The goroutine naturally retains the entry until its envelope is delivered; Fleet
and root can remove their exact child identity without a second map or tombstones.
The shared entry retains its stop/save error for both racing callers. Managed
workers start in a fresh child directory and never Load prior task history.
Ordinary Spawn retains restoration and its public enablement gate. Ordinary kill
continues preserving its historical GUI/reference behavior; automatic root/GUI
removal applies to managed workers. Root exposes current identity membership for
GUI late-observation rejection and uses a monotonic default-directory counter.
A managed removal observation carries its owner for selected-tab fallback; the GUI
must process removal before normal view construction. These are concrete existing
owner APIs, not a new task service. The runtime reviewer accepted this refinement
as consistent with the complete design before implementation.

## C8 implementation evidence: managed Fleet completion and removal

After C7 `22763ac6`, StartCompaction shares ordinary child construction but refuses
an existing task directory and never loads prior work. Its buffered envelope
follows the actual Result, shutdown/drain, final snapshot attempt and exact entry
removal. Explicit kill/owner shutdown marks cancellation before sharing the entry's
sync.Once stop/save result. Shutdown snapshots entries, rather than re-resolving
IDs after automatic removal. The entry retains its manager address independently
of a trusted custom factory's display label. Root removal is managed-only, retains
files, publishes AgentRemoved, and releases its actual reference; current membership
and a monotonic allocation sequence support later GUI cleanup without tombstones.
Ordinary child restoration/public delegation and historical root retention remain.
GUI consumption and parent-actor attempt accounting/commit are not this slice.

Real HTTP/provider and filesystem tests cover a held request, successful saved
submit tail, final snapshots containing actual job-finished facts, canceled actor
admission, log-save failure, final snapshot failure, explicit kill, owner shutdown,
five repeated fresh workers, no retained running submit jobs/root/Fleet entries,
refusal to reuse a prior directory, monotonic default allocation, initial snapshot
refusal before admission and custom-factory label versus reserved-address cleanup.
Initial compile used a nonexistent Jobs.List and was corrected to inspect actual
recorded job handles. The first cancellation peer did not consume its request body;
its handler remained in httptest.Close after the real client stopped. That run was
terminated with SIGQUIT after72.503s; its stack/output is retained, not counted as
runtime success. The corrected peer drains the body and has an explicit cleanup
release. Its new-test race passed1.111s. A later initial-save fixture used Mkdir
without its parent, so the custom callback itself returned nil/error after creating
an Agent; this was not the intended launch Save refusal. MkdirAll fixes that setup.
The accepted trusted-callback side-effect limitation still applies.

Final affected Fleet/Named/Management and new lifecycle race passed4.380s; targeted
vet passed. Subsequent test-only refinements verify exact final snapshot job facts,
all closed actor admission and the custom label case; final focused race passed1.128s.
Outputs use `/tmp/ensemble-compaction-stagec-lifetime-` with `compile`, `initial`,
`corrected`, `peer-corrected`, `race`, `vet`, `final`, `focused`, and `final-focused`
suffixes. Original fixture failures retain their original files.

Thirteen focused mutations fail intended assertions: discard reliable Result error;
discard shutdown/save error; omit final snapshot; omit shutdown; retain root or
Fleet entry; lose explicit cancellation; omit final usage; reuse an existing task
directory; skip initial Save; reuse shrinking-registry directory numbering; remove
by a custom label instead of the reserved key; and stop before waiting for Result.
The last mutation fails the working provider-admission control after its bounded
five-second test wait: it proves premature shutdown prevents the task, not a
specific submit-tail interleaving. The real success test separately checks the saved
tail/job facts. Cancellation-marker mutation changes both explicit and owner paths
as one policy defect. Logs use `lifetime-mutant-<case>.txt`; scratch copies removed.
No fake success is substituted for task Result or persistence errors.

Independent C8 review is pending. Parent shutdown's obligation to consume the
managed completion/accounting/outcome before its own final Save, near-window
waiting, safe acceptance/generation and GUI dismissal still require integration.

C8 preliminary review reproduced an actual factory-path defect: checking the
requested fresh directory did not stop a constructor-only custom factory returning
an existing actual save directory; initial Save overwrote its previous snapshot.
The independent failing probe/log are `/tmp/ensemble-compaction-c8-review-factory-path_test.go`
and `...-factory-path.txt`. CreateChild now requires each explicit managed data,
log and save path to match the returned Config before Fleet performs persistence.
A factory still chooses provider configuration, but must return before loading,
saving or starting a managed child. Arbitrary callback effects before return remain
trusted, rather than being claimed reversible. Correctly owned rejected managed
children are drained and removed; foreign ownership remains untouched. Empty
managed shutdown skips log Dump, since otherwise refusal itself truncated an old
log with an empty header. This is limited to a worker with no recorded facts.

The missing successful-Result/kill boundary now uses an actual FIFO at the final
snapshot backup path. Submitted idle is published after active-turn detachment;
joining the child then establishes that finish sent its reliable Result before
stop completed. The FIFO still holds final Save. Explicit kill publishes its
cancellation request before waiting on the shared stop/save; only then does the
external reader release Save. Both success and a real journal-reset failure retain
successful Result plus Cancelled, and return the same cleanup error to both callers.
No actor, History, Engine, job or filesystem implementation is mocked. The Unix
FIFO is a filesystem boundary fixture; this is not a timing-only repeated race.

The new three tests passed race1.053s, including both shared-save outcomes. The
initial new-test compile used AgentSpec where CreateChild exposes common.AgentSpec;
that fixture correction and raw compile failure are retained. Final affected
Fleet/Named/Management, structured factory ownership and lifecycle race passed4.468s;
targeted vet passed. A later assertion-only refinement checks exact old bytes before
reporting path-admission failure. Six additional intended mutations fail: remove
actual path validation (old save/log bytes overwritten); dump empty refused worker
log; omit rejected-worker removal; capture cancellation before held Save; discard
shared error for the second caller; execute Save twice (the second FIFO open remains
blocked and the explicit five-second shared-save assertion fails). These complement,
rather than relabel, the earlier premature-stop mutant. Outputs use the existing
`lifetime-` prefix with `review-fixes`, `review-fixes-corrected`, `review-race`,
`final-review`, `final-review-vet`, and `mutant-actual-paths/empty-log/refusal-removal/
race-cancel/shared-error/repeat-save`. Scratch copies were removed. Zero usage in
the initial-Save refusal fixture means no measured usage, not independent proof of
no HTTP attempt; source ordering and its skip-initial-Save mutant establish that
admission boundary. C8 awaits final independent disposition.

Independent C8 scoped acceptance: exact predecessor plus the seven production
paths, lifecycle test and unchanged independent overwrite probe pass race1.146s
(`/tmp/ensemble-compaction-c8-review-final.txt`). Reviewer inspected all19 intended
mutation failures and accepted the FIFO, path-preservation and rejection cleanup
corrections. Executed binding is
`9d399bc32e290904980c7115efde5351471636df3d80e4194bfe8e09cf198cfc`;
final comment-only binding is
`55761342a65f2095753e758867d49564b87dfefa4dde408718f78545135b9df8`.
Final coder assertion-only race passed1.052s. No remaining C8 blocker; zero formal
Stage C correction rounds. Full GUI/actor/accounting integration remains pending.

## C9 implementation boundary: GUI managed-worker dismissal

Reuse the existing selection/settings/replay path for explicit selection and
automatic fallback. The synchronous root observer deletes a removed managed view
and schedules the existing tree refresh worker; it must never traverse parent
Fleet or load settings there. Current root membership rejects late observations
before view construction. The refresh worker resolves selected removed workers to
their actual surviving owner, loads that owner's ordinary settings outside the
observer lock, and queues its full replay in one batch. Compare the expected old
selection so a concurrent user selection wins; recheck membership after settings
loading so an in-progress selection cannot recreate a removed worker. No tombstone
map, extra worker service or result browser. Selected connection references survive
only until this existing refresh work or connection teardown releases them.
The real socket test first fails on C8 for both direct and nested owner fallback,
with two explicit five-second read deadlines (10.105s total), retained in
`/tmp/ensemble-compaction-stagec-gui-removal-before.txt`.

C9 uses two GUI production paths: the observer handles removal before creating a
view, and the existing tree worker performs deferred selection through the same
helper as ordinary socket selection. The helper preserves indivisible replay,
normal owner preferences, slow-client disconnect behavior and a user's newer
selection. It checks managed membership both before view creation and after the
settings read. No core source or new public GUI protocol changes in this slice.
Physical production delta is100 added/40 removed lines (net60); most moved selection
code is now shared rather than duplicated. Tests use the actual runtime, HTTP peer,
WebSocket server/client and settings FIFO, with internal map inspection only to
verify retained references. The provider routes a sole submit declaration with a memory property; exact
whole-schema equality remains inherited C2 evidence. It does not hold a global
route mutex while waiting.

Both root and nested-owner cases repeat three fresh workers, select the live worker,
then verify owner selection, tree removal with owner positive control, actual owner
tool-result replay, saved font preference, released view/selection references and
rejection of a retained actual late observation. The initial late-observation fixture
captured AgentRemoved itself; inspection corrected it to retain an actual ordinary
observation instead, and its targeted mutant now recreates the view and fails.
A second test holds settings loading at a real FIFO while the worker completes:
completion remains unblocked, and delayed selection falls back instead of publishing
a removed worker or the now-inapplicable managed-settings error. This is a concrete
selection race, not a probabilistic repeated timing loop.

Initial new tests race passed1.128s; strengthened tests passed1.145s. Full GUI Go
suite passed8.292s and GUI vet passed. Final focused new/selection/structured-replay
race passed4.403s after explicit tree and owner-history assertions. Existing socket
selection and slow-receiver tests run through the shared helper. Raw evidence uses
`/tmp/ensemble-compaction-stagec-gui-removal-{before,first,races,all,vet,final}.txt`.
Focused mutations remove view deletion, late membership refusal, actual-owner
fallback, owner settings load, forced fallback replay, and the settings-error
membership retry; all six fail intended assertions. Two further fallback-only replay/tree mutants also fail intended checks. The first
broader replay mutation stopped initial settings delivery; it was narrowed so the
recorded final mutant reaches fallback and fails the owner-history read. The expected-old-selection guard and the second
membership check after a successful settings read are source-reviewed race guards;
the FIFO establishes the actual settings-error/removal interleaving, not every
possible scheduler interleaving. Full parent actor integration remains separate.


Independent C9 source review identified a narrower view-admission race: selection
could pass the first membership check, wait for the GUI mutex while removal deletes
the view, then recreate it. A failed settings read retried without deleting that
recreated entry. View admission now rechecks membership while holding the same GUI
mutex used for deletion. If root removal occurs later, its observation necessarily
deletes the admitted view after the lock releases. The FIFO test covers removal
during settings loading; this additional lock window was source-identified rather
than reproduced by that fixture. No test-only runtime hook was added.

Independent C9 scoped acceptance: exact `59d6fcfb` plus the two GUI production paths
and new test passes new/selection/structured-replay race4.409s, raw
`/tmp/ensemble-compaction-c9-review-final.txt`, binding
`10881fb96e70b994dafa367aeae9f6f36eca3827404fdf3d01ada7fd73c177bd`.
Reviewer accepts the locked-admission correction and inspected all8 final intended
mutation failures, with the broader replay limitation preserved. Final coder
admission-fix race passed4.401s. No C9 blocker; zero formal Stage C correction rounds.
This is GUI dismissal acceptance, not parent actor or full Stage C acceptance.

## C10 actor integration boundary

The first new source regression holds only the memory HTTP exchange on an
independent concurrent peer. On C9 `f111f1b3`, the owner's main turn cannot finish
while that exchange is held: the explicit two-second progress assertion fails in
2.020s (`/tmp/ensemble-compaction-stagec-actor-progress-before.txt`). This is the
current actor's direct Compress path, separate from root's earlier saved-state
write/event/archive probes. Existing memoryPeer also holds its callback under a
mutex; it cannot establish background progress and will not supply this barrier.

One runtime attempt holds immutable selection, source/destination generations,
reliable drained completion, accounted flag and any terminal outcome awaiting
persistence/cleanup. Its completion receiver counts in runtime.workers and posts
directly to the mailbox even during shutdown. Completion imports final buckets
once, then waits for idle or pre-render acceptance; ordinary finish/interrupt does
not cancel it. Settings invalidate the relevant generation synchronously and
request child shutdown outside the actor. Owner shutdown invalidates before its
existing recursive drain and consumes the completion before exiting. Terminal
write/cleanup errors keep the attempt/receipt; a committed receipt is finalized,
never relabeled abandoned. A stopped owner returns persistence errors rather than
waiting forever on unrepairable I/O. Outcome adds copied usage receipt data without
replaying spend. Near-window waiting stores the active turn identity, and a failed
held attempt fails that turn explicitly; interruption detaches only that waiter.

C10 implementation now uses the existing Fleet worker rather than Engine.Compress
from the actor. The legacy exported compressor remains source-compatible and is
marked deprecated. The selected task is validated again after hidden staging and
before the correlated pair. Parent Save establishes the journal before launch.
Main 95% occupancy uses History.LastUsage; completion imports Engine buckets without
changing that main observation. Root source review noted the completion count was
decremented before identity admission; first exact completion now gates both worker
retirement and import, without a new queue or ledger. There is one normal producer;
no artificial duplicate-message hook was introduced to claim scheduler coverage.

Focused concurrent-peer lifetime/safe-boundary race passed1.496s; cache/headroom
success, failure, below-limit and interrupt cases passed1.261s with the refusal
controls. Actor file-failure cases passed1.373s: worker log/save, staging refusal,
human destination edit, and committed promotion permission failure. The latter
holds render and settings until repair, then completes; terminal outcome/usage
survive Save/Load and historical buckets are not imported into a new session.
Raw outputs: `/tmp/ensemble-compaction-stagec-actor-lifetimes-meter.txt`,
`/tmp/ensemble-compaction-stagec-context-limit-final.txt`, and
`/tmp/ensemble-compaction-stagec-actor-persistence-first.txt`.

Inherited memory tests now wait for actual terminal outcomes before checking disk
or issue a marked verification turn before checking compressed request contents.
The waterfall waits for three real committed outcomes. The old interrupt-then-owner-
shutdown case is named for shutdown; ordinary interrupt preservation is independently
covered by the new concurrent peer. First unchanged timing tests failed, then one
leftover replay request count failed after adding a verification turn. The final
memory/lifetime/refusal focused race passes2.740s, raw
`/tmp/ensemble-compaction-stagec-refusal-correction.txt`. Fixture corrections also
retain their actual failed logs: two guessed latest-usage method names did not
compile (the real API is Meter().Usage); the ordinary-shutdown control initially
omitted LogPath; its first draft used public AgentSpec where CreateChild takes the
common type. These are test setup failures, not claimed runtime defects.

Root overall review reproduced a C8 coverage gap on exact `f111f1b3`: a managed
request whose factory returned an ordinary child could truncate a redirected old
log during refusal and remain in the root. The managed-return positive control
passed; reverse-purpose refusal failed both preservation and removal. Correction
adds DiscardConstruction to the existing counted shutdown path and suppresses only
its persistence, replacing the purpose-specific empty-log workaround. CreateChild
checks ownership first, discards rejected owned construction, then shares private
exact-identity root removal. Ordinary successful Shutdown still saves and retains
its root entry. Factories must return before Load/Save/task admission; arbitrary
trusted callback pre-return effects remain outside this contract. Root's unchanged
probe is retained with formatting only in compaction_refusal_test.go, together with
an ordinary lifecycle control. Independent review of this correction/full C10 and
mutation evidence remain pending; prior C8 acceptance history is not overwritten.

Final coherent checks: focused race across Compaction, Memory, StructuredSubmission
and both refusal controls passes5.333s; core and GUI go vet pass. Full core tests
pass the Ensemble package92.956s and all other packages except the previously
reproduced CLI human-sink failure TestFleetCLIHumanReplyAndEscalation/true
(unreachable sink, process killed). The full GUI module passes8.538s. These are
isolated local peers with all six credential-prefix families stripped only from
test subprocess environments. No live inference, graders or deployment ran.
Logs: `/tmp/ensemble-compaction-stagec-actor-focused-final.txt`,
`/tmp/ensemble-compaction-stagec-core-full.txt`,
`/tmp/ensemble-compaction-stagec-gui-full.txt`,
`/tmp/ensemble-compaction-stagec-vet.txt`.

Fourteen isolated single-defect checks fail intended behavioral assertions:

| Defect | Observed failure |
| --- | --- |
| Omit auxiliary import | Failed empty submission loses exactly80 input/20 cache/9000 output |
| Accept while main response is held | Source commits before the safe boundary |
| Omit policy generation advance | Disable/re-enable wrongly commits old work |
| Ignore child cleanup/save error | Failed child save wrongly commits memory |
| Omit 95% wait | Next main request arrives before headroom |
| Resume failed held attempt | Main returns success after failed compaction |
| Count cache twice | Below-threshold request remains incorrectly held |
| Omit terminal usage receipt | Saved outcome has no measured usage |
| Skip Memory.Finish | Promotion permission failure is falsely committed |
| Skip settings recovery hold | Settings succeeds against unfinalized corpus |
| Disable construction-discard save suppression | Root probe's existing log is overwritten |
| Skip rejected child removal | Root probe retains rejected ordinary/managed child |
| Kill worker on ordinary interrupt | Interrupted owner loses ongoing worker |
| Omit counted completion receiver | Owner Shutdown returns without terminal outcome |

Raw mutant outputs are `/tmp/ensemble-compaction-c10-mutant-<name>.txt`, with
names usage, safe-boundary, generation, cleanup-error, headroom, failed-wait,
double-cache, outcome-usage, finish, settings-hold, discard, root-removal,
interrupt and shutdown-receiver. All outputs were inspected. The first interrupt
patch did not match the actual parameter name and was not applied; the corrected
patch produced the recorded behavioral failure. Scratch copies were removed.
The cache timeout establishes the actual blocked-request property with a passing
control; the shutdown timeout establishes the missing terminal result after actual
Shutdown, not a generic hung build. The latter is one exercised schedule, not an
exhaustive scheduler proof. Exact first-completion identity/duplicate admission
remains source-reviewed because production has one completion producer; no test
hook or growing task ledger was added. C1–C9 mutation evidence covers unchanged
schema, adapter, selection, correlated history, staged recovery, accounting bucket,
shared save/removal and GUI boundaries; these were not all mutated again.

C10 size against its predecessor across the seven changed/new production Go paths:
+223 physical lines, +191 active nonblank/non-comment-only lines and +22 comment
lines (full compared files1403→1594 active,525→547 comment). This includes removal
of the old synchronous actor path and the root-found constructor refusal correction;
tests and notes are excluded. One attempt and the existing workers/mailbox carry
lifetime; Memory retains filesystem policy, History retains projection receipts,
and Engine retains accounting. No extra scheduler, task ledger or persistence store
was introduced. Whole-file totals include inherited exported-interface documentation;
the independent review must judge the actual new lifetime/commit explanations too.
The full module runs began before the final small finish-time clearing of a detached
headroom waiter reference; the final5.333s focused race and both vet runs include
that final source. Manifest `/tmp/ensemble-compaction-c10-final-source.json` binds
the final production/test paths for independent review.

The refusal correction also exposes a narrow GUI case: a selection holding a child
pointer before refusal could select the now-removed ordinary child because the
membership guards tested only managed purpose. Independent reviewer agreed with
the source finding. Existing guards now check parented children against current
root identity; valid ordinary children remain registered and preserve selection.
The real refusal/socket regression invokes the shared selection continuation with
the already-resolved pointer, without a runtime test hook. Restoring all prior
purpose-only guards selects `main/rejected` and fails the exact owner assertion
(`/tmp/ensemble-compaction-c10-mutant-gui-refusal.txt`,0.024s). This is the fifteenth
intended mutation failure. The new test plus C9/ordinary selection race passes4.327s
(`/tmp/ensemble-compaction-stagec-gui-refusal-corrected.txt`). Its first two attempts
omitted the socket subscribe request and timed out before exercising the defect;
those setup failures remain in gui-refusal-before/final logs, not claimed evidence.
Root independently reran its unchanged factory refusal assertions with only its
function/comment renamed to avoid duplicate test symbols: race1.030s, raw
`/tmp/ensemble-root-review-factory-refusal-fixed.txt`; initial duplicate-symbol
setup failure remains separate. Final manifest includes all14 core/GUI source/test
paths, including the two AgentRemoved comment clarifications. No other runtime
behavior changed during this final GUI correction.

After the final GUI correction, the complete GUI module again passes8.367s and vet
passes (`/tmp/ensemble-compaction-stagec-gui-final-all.txt`). The final core actor
bytes remain those exercised by the5.333s race; subsequent core edits only clarify
AgentRemoved comments. The manifest hash is
`70ece4a33f7fb4cd6b6ad8995053809e19955fbaedb7e01dbbaf94fc952e5640`.
The final ten production paths (including GUI and the comment-only activity type)
retain the same +223 physical/+191 active/+22 comment delta; their full-file totals
are2737→2960 physical,1986→2177 active and619→641 comment lines. Tests are separate:
the actor regression file exercises external concurrent HTTP/filesystem boundaries;
the refusal regression preserves the root probe and ordinary lifecycle control.

Independent full Stage C acceptance: initial formal review, zero formal correction
rounds. Final exact14-path independent race passes core5.324s and GUI4.434s;
review binding `387b87f643844024bf1c41722be33905da8a51af38a2cd6e39d406694ba72305`.
Reviewer inspected all15 C10 intended mutant failures, accepts the repaired root
refusal coverage gap and GUI ordinary-child membership correction, and records the
complete C1–C10 comparison and limitations in the separate runtime audit. Coordinator
also accepts runtime implementation after final source/tests/hash review and its
independent root refusal reproduction. This accepts Stage C runtime code, not
Stage D revised graders, later prose/renumbering/exports, live quality or deployment.
