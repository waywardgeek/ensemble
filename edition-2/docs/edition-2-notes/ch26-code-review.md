# Original26 / Edition2ch23: independent code review

**Current disposition: accepted with two narrow original-grader launch exceptions
after one formal correction round.** Final source `450f7a9599295dafb260c1650b9b5553010e2a54`,
main tree `27d7fcf9b8b31275068fa705d6969cfe3eafedc1`. Original grading remains
0/100 FAIL; full rationale and final evidence are below. Historical preparation
and failed initial review remain intact.

## Initial preparation (historical)

Reviewer `reviewer_astra`, GPT-6 Astra/high, October 11, 2026.
**Preparation only; awaiting first-green implementation. Zero formal rounds.**
Mandatory `.agents/skills/ensemble-coding/SKILL.md` was read in full, with the
current AGENTS, governing updates and student-safe [ch26 notes](ch26.md).
No original26 implementation answer or student implementation was inspected;
no grader or provider call was run.

Read all466 lines of `book/chapter-26.md` at
`b451d2f14d8170b9075027a0a6606badfb442dad`; SHA-256 remains
`8db9779e9099d971e8fb8fd5c8f531e110d6a94c0bf234ed836f029aa159c56e`.
The coordinator's assignment-boundary completion check satisfies Bill's release
condition; historical source scores are not new student results. Accepted
predecessor is original25 source `e00d5bc`, main tree
`30fb05d154b9808eb88622aaff9c1424cd7e8af1`, direct combined Edition2ch22.
The existing narrow public contracts in ch26 notes remain sourced to grader
`9bab19dc914f04f16c5f7f5fdcc6c8345cf0b3c7`; verified file SHA-256
`ec6e7f98b4576cbb2e3bd834632cee8106ecc9152bf0edaa0d9115ad28dbc287`.
No further grader inspection is needed to restate those facts.

The independent review will preserve these manuscript obligations:

- Activity transitions through the existing Observer, with useful tool/detail
  information, and lifecycle distinct from thinking/tool/waiting/idle phases.
- Real named-peer create/restart/conversation preservation and deletion that
  refuses active children, with the registry owned by the existing Ensemble.
  Spawned children remain ephemeral; persistent spawning from the old brief is
  not part of the assignment.
- Each new management operation resolves only the caller's actual children;
  absent and foreign targets must be indistinguishable. Accepted duplicate names
  across ancestors and trees remain valid, so browser selection must resolve the
  actual node rather than assume a globally unique display ID.
- Change waits cover the specified state/exit/message/pattern/stuck conditions;
  cursor reads preserve actual new output; synchronous questions and asynchronous
  hints use the real actor; submissions have their required durable fallback.
- Section26.4 distinguishes orderly shutdown from immediate kill. Both interrupt
  modes end the current turn while preserving the Agent; Bill's prior ruling
  preserves other jobs and their real results. These lifetimes need actual checks,
  not labels on one equivalent shutdown path.
- The real CLI and optional GUI use the accepted public composition root. New
  fleet GUI clients enable original25's factory protocol for root and descendants.
  Initial tree plus ongoing changes, nested activity, conversation click selection,
  named create controls (name/model/skill), deletion and visible rejection remain
  required despite the older coder review's narrower historical deferrals.

Constructor names, React, and a particular incremental-versus-full-snapshot
implementation do not override Bill's existing architecture or KISS. Either
recorded tree-message spelling is accepted; visible behavior and actual owner
relationships remain the test. Do not recreate a flat root or introduce callback
bags to match the source narrative. Source confidence claims about compile-time
wiring, fleet-scale security or independent mutations are not assumed facts.

After first green: read the student review and new code, then matching original
source/tests; compare both editions against their own predecessor, including
active code, useful comments, exported contracts, tests and support files.
Final disposition waits for actual unchanged grading, focused mutation outcomes,
real browser/live evidence and the complete student review. No exception is
preaccepted. Preserve concurrent CodeRhapsody work and exact-owned-path commits.

## Stage1 preflight, October 11 (not formal review)

Implemented stage `55cbef230e675ed0d42547b7ad21e53c5b8a554f`, tree
`967ed4ff7d9379c37ce3eeadac2800803431e442`, covers activity and management.
Read the student stage notes and all current changed files; matching original
management code is now permitted comparison context. Registry and GUI stages
are explicitly unfinished and are not treated as final omissions. No old answer
code was relayed to the student.

A real defect is independently reproduced: `WaitChange([watched])` reports
`watched`/`message` when only the unselected child `other` asks a question.
Both child requests are parked, so no watched-child activity/state/output changed.
The implementation applies a manager-wide message sequence to every selected ID;
that is not the requested set-scoped monitoring (§26.4). The race-enabled probe
fails cleanly in0.075s. Controls prove no-progress/no-question waits time out and
a selected-child question wakes with its own identity (both pass,1.458s).
[Failure and controls](../../solutions/evidence/ch26/reviewer-wait-scope.txt)
and adjacent reviewer probe sources preserve the setup and an initial HTTP-fake
cleanup timeout. This finding was sent through the coordinator as behavior and
rationale, without reference-answer code. The coder owns its correction.

Two small documentation inaccuracies were also relayed for final cleanup:
`FleetTools` still describes seven declarations after adding eight management
tools; `ChildStatus.State` omits the newly returned `idle` state. These are
accuracy corrections to exported contracts, not new capability requests.
Zero formal rounds used; final disposition remains pending full implementation.

## Stage2 preflight, October 11 (not formal review)

Reviewed exact `33c4e3315b2210bff946aa9d94bb27f7dda0c340`, main tree
`918fd9c466bdc55d7d4b245d78b0d846b043cc96`, through an isolated archive.
Moving GUI browser files were excluded. Read the student stage notes, complete
registry and tests, startup/fleet/test delta, and matching first-edition registry
and composition implementation. Mandatory `.agents/skills/ensemble-coding/SKILL.md`
and carryover were reread after context compaction.

The exact three preserved reviewer wait probes now pass under race (1.807s):
an unselected child's question no longer wakes the watched child, a selected
child's question does wake it with the correct identity, and the no-question
control times out. Per-sender revisions replace the manager-wide attribution.
The prior failure remains preserved; this is a verified correction, not a
reclassification of the old result. [Exact rerun](../../solutions/evidence/ch26/reviewer-stage2-wait.txt).
The two exported-contract inaccuracies identified in stage1 are corrected too.

The shared `Start` path installs the recursive factory and human sink before
constructing/restoring root and named peers. Both actual executables call it;
offline render/dump retain the non-loading path. Named peers stay separate from
ephemeral children, while tree keys follow actual parent paths so repeated local
names remain selectable. The registry uses the required four identity fields,
restores conversations through normal Agent loading, and writes via temporary
file/rename. Named deletion checks children, shuts down the owner, then removes
registry identity and files. Its existing real-runtime tests independently pass
under race (1.169s), covering restart/provider-history preservation, repeated-name
path identity, delete/restart absence and an actually parked child's rejection.
[Focused rerun](../../solutions/evidence/ch26/reviewer-stage2-registry.txt).
This does not claim every concurrency interleaving or final browser integration.

The modified GUI budget fake still exercises the built executable, persisted
settings, a two-batch turn bound, next-turn eighteen-batch behavior, restart and
zero/default behavior. Submitting as batch18 replaces the former eighteenth
shell call while retaining the actual dispatched-count/error assertions; it
adapts the external response to the enabled fleet completion protocol without
waiving the inherited budget contract.

Compared with the matching old registry/composition, this stage extends the
existing ownership root and Observer rather than adding callback wiring or
copying sibling services. Useful comments explain path identity, persistence
and lifecycle boundaries; the new public types/members have contract comments.
The change is proportional to the registry/selection requirements. Final whole-
chapter own-predecessor counts, management edge coverage, mutations, original
grading, browser/live evidence and final student review remain pending. No new
blocking finding in this stage, no acceptance or grader exception, zero formal
correction rounds.

## Full first-green preflight (not formal disposition)

Exact source `974f67be89272ea60fa17b460cefde0e478a137b`, main tree
`c34e436798cead5d2dc9d8899887da7b6eeda485`. Read the complete student stage3
notes, new GUI/view/browser code and tests, management/registry changes, public
contracts and README. Matching original activity, management, registry, GUI and
composition changes/tests were read after implementation, never supplied to the
student. Original GUI has no conversation-selection implementation and creates
by name/task rather than model/skill; those omissions do not waive this student's
manuscript obligations. The new implementation actually routes per-connection
selection and replay by owned Agent, reuses the Observer, and traverses trees in
a coalesced worker. Atomic child-admission closure repairs the earlier deletion
check boundary. No callback-bag or optional-module boundary regression was found.

**Blocking pre-submission finding:** repeated `select_agent` can deadlock the
GUI after its slow browser disconnects. The new selection branch sends a replay
batch to a bounded queue while holding the shared observer mutex. Once full,
a failed/exited writer leaves the socket receiver blocked on that send, retaining
the mutex needed by observations and Close. A real WebSocket peer, one64KiB
history item and320 selections reproduce it; no provider or internal mock is
involved. A one-selection positive control passes. The final race run fails
the intended bounded shutdown assertion in4.141s, with actual goroutine stacks
at `gui.go:332` (channel send) and `gui.go:465` (Close mutex). The first600ms
admission attempt passed; allowing3s exposed the full queue. Both are preserved,
not relabeled. [Probe, control and stacks](../../solutions/evidence/ch26/reviewer-selection.txt)
and adjacent exact sources record the finding sent through the coordinator.
No production code was changed; correction and its mutation belong to the coder.

Original grading actually reports0/100 FAIL: six checks cannot start inference
without a model; the GUI check supplies `--port` to the headless executable and
gets usage/refused connection. Those are the observed fixture boundaries, not
fleet behavioral failures, but they also provide no proof of the masked valid
obligations. Full exception assessment awaits the final student explanation and
supplemental evidence; no additional unobserved envelope failure is claimed.

[Provisional own-predecessor counts](../../solutions/evidence/ch26/reviewer-first-green-counts.json)
separate languages, tests and scripts. Student production Go adds1007 physical,
860 active,99 comment-only and48 blank lines; original adds1067/838/107/122.
Student Go tests add412/380/18/14 and browser tests68/66/1/1; original Go tests
add864/669/93/102. Student production JS adds34/32/2/0, HTML7/7/0/0, CSS9/7/1/1.
Original JS adds98/79/5/14, HTML changes-3/-4/+1/0, CSS adds59/53/0/6; its
removed418-line browser-support script is separate, not a production saving.
The original snapshot also removes `mcpws/mcpws.go` (109/87/13/9); its net Go
total includes that deletion, which is not an equivalent new management feature.
These are net tracked-source totals, excluding evidence/generated artifacts;
comment classification is lexical and inline code stays active. Refresh at final
freeze. Lower test totals alone are not an improvement, nor are the original
source's confirmation-only interrupt checks equivalent to observing a retained
real tool result and a later successful task. The current explanatory comments
cover ownership and critical lifetime ordering, but the new production comment
ratio is below the teaching target; assess clarity rather than pad to a quota.

Final checks still depend on correction of the demonstrated deadlock, targeted
mutation outcomes, actual live/browser provenance and the full student review.
No formal correction round consumed; no chapter acceptance or export.

## Selection correction verification (pre-submission)

Exact corrected source `9e2b72aeaddc85cbe1afadec41909b4d66ad3b34`, main tree
`38aee9dd34ca180dedeed0fe17bf4cfe2173ef84`, was archived independently. Both
preserved reviewer sources were copied unchanged: slow repeated selection and
one-selection control now pass under race in4.150s. [Exact rerun](../../solutions/evidence/ch26/reviewer-selection-corrected.txt).
The failed first-green reproduction remains above; the defect is corrected.

The source admits the complete replay as one existing queue batch, or releases
both mutexes and returns through that connection's established teardown. It
publishes the selected view only after successful admission. No buffer increase,
new limit, silent history-loss success, Agent shutdown or other-client disconnect
is added. A reconnect can replay from the authoritative history as before.
Read the coder's corresponding real-socket regression and the second-browser-tab
isolation addition; neither substitutes an internal mock. The three-child
monitor addition also retains real actor/provider history and checks returned
identity, submissions and pattern matching. Final mutation/live assessment is
still pending; zero formal rounds and no final disposition.

## Initial formal review — needs correction

Submission `944c6ee0b329f9463e0d1b0aa7a7dc28472210ae`, main tree
`edd8f498661273752353e8bd8df7a1ffc7e7ff11`. Read the complete final student
review in [ch26.md](ch26.md), all20 mutation attempts and actual failure outputs,
final source delta, live GUI frame/DOM evidence and preceding comparative source
review. Final source differs from independently tested/live `9e2b72a` only in
three documentation comments; the independent runtime reproduction below applies
to this submitted behavior. **Not accepted. The next correction is round1 of3;
zero formal correction rounds have been consumed.**

One blocking finding remains. A successful explicit fleet submission ends the
actual turn but selected-conversation replay reports `input_pending`, including
after named-peer restart. The live trace first exposed this: final submission
completes at frame1377, restart selects console at1394, tree1395 is idle/idle,
but selected replay1428 reports input_pending. Frame1429 reports paused=false;
the always-present Unpause everyone button is not evidence of a pause defect.
The coordinator identified the screenshot discrepancy; the reviewer traced it
and independently reproduced the actual behavior.

The reducer enters input_pending after a completed tool batch in
`internal/history/reduce.go`; normally another provider request consumes it.
Successful fleet submission instead finishes directly through
`internal/agent/dispatch.go`/`actor.go`, leaving that durable turn state pending.
The new `agentView.replay` presents it as current state. A small external-provider
fake exercises real Start/CreateAgent/Ask, successful submit, WebSocket selection,
Close/Start and selection again. Both completed and restarted Agents have idle
activity yet replay input_pending: clean intended race failure0.064s.
[Exact regression](../../solutions/evidence/ch26/reviewer-state.txt),
[probe source](../../solutions/evidence/ch26/reviewer-state_test.go.txt),
and [independent live trace](../../solutions/evidence/ch26/reviewer-live-state-check.json).
Correct the actual completed-turn/current-replay disagreement, preserving true
busy and interruption behavior. Demonstrate both immediate selection and saved
restart through the real runtime, with an affected regression and targeted
restored-defect mutation. No additional paid model run is needed merely to
refresh source provenance for this local state correction.

I agree with the student's two narrow original-grader explanations. Six checks
fail at an omitted model before inference; the GUI probe launches the headless
entry with GUI arguments. The actual original score remains0/100 FAIL at974f67b.
An explicit configured model and the shipped optional GUI are the valid entry
contracts; a hardcoded default or core GUI dependency would not repair a fleet
behavior. These fixture assumptions are accepted exceptions, not waivers of
activity, persistence, ownership, waiting, composition, GUI or inherited submit/
idle-correction obligations. No unobserved downstream envelope failure is claimed.
The current blocking state defect is not covered by either exception.

The other student explanations are supported: finalized cursors avoid confusing
partials with later final output; JSON-line submissions retain multiline results;
named peers and ephemeral children retain distinct lifetimes; existing non-fleet
consumers keep the previously accepted boundary while both shipped applications
enable recursive fleet construction. Scope stays within the manuscript and
Bill's actual-owner/optional-GUI rules. Real named conversation restoration,
owned lookups, recursive controls and two-tab isolation improve on the matching
reference's existence-only persistence and confirmation-only interruption checks.
The reference's missing conversation selection/model-skill controls are not
excuses for omitting them here. No bloat or architectural blocker remains.

Mutation evidence supports19 clean intended kills over20 attempts. The initial
graceful mutation ends in HTTP-fake cleanup timeout and is correctly excluded;
the repaired request-body fixture produces the intended premature-shutdown
failure. The nested-tree mutant's visible-node deadline detects missing recursive
rendering with a passing browser baseline; the selection mutant fails the
specific bounded Close assertion. Cursor, ownership, usage, phase, persistence,
message, pattern and stuck failures match their claims. This evidence establishes
those sensitivities, not exhaustive coverage or the validity of the stale-state
assumption that the live run exposed.

The independent GUI audit finds all1432 bidirectional frames, selections
main→console→console/helper→console→main→console, all four activity phases,
console/spare creation, spare deletion and restored console without helper.
Actual artifacts and per-Agent history record the child implementation, measured
three-test passes by child and parent, parent monitoring/join/submission retrieval
and final CH26_LIVE_COMPLETE. The coordinator separately audits15 Astra/high
requests/completions and raw tool/file correlation; no reviewer provider or
credential access occurred. This is bounded real fleet coding evidence, not
Edition3 authorship or a scale claim. The original live state discrepancy remains
in history even after correction.

[Initial-submission counts](../../solutions/evidence/ch26/reviewer-initial-counts.json)
refresh the same tracked lexical method: student production Go delta
1018physical/866active/104comment/48blank versus original1067/838/107/122;
student tests537/493/26/18 Go plus75/72/1/2 browser versus original864/669/93/102
Go. Other production deltas remain as above. Moved code counts once; original
MCP-transport deletion and removed script are explicitly distinguished above.
These sizes are comparable; the student implements required selection and durable
cursors omitted or weaker in the reference, while using existing ownership and
transport boundaries. Added Go comments are about10.7% of nonblank lines, below
Bill's rough20% teaching target, but they explain lifecycle, path identity,
admission and persistence decisions without padding. Exported contracts are
documented, including the final multiline-submission field comment. The review
judges those explanations useful; it does not require quota-driven expansion.

## Round1 final disposition — accepted with exceptions

Reviewed the full updated student review and exact source
`450f7a9599295dafb260c1650b9b5553010e2a54`, main tree
`27d7fcf9b8b31275068fa705d6969cfe3eafedc1` (180 tracked source files).
**Accepted after one formal correction round**, with only the two original-grader
launch exceptions justified in the initial review: omitted model for six CLI
checks, and headless executable selected for the optional GUI probe. Original
grading remains0/100 FAIL on974f67b; no altered-grader pass or additional waiver
is claimed. All required capabilities remain in scope. No unresolved defect,
architecture or bloat finding prevents the direct Edition2ch23 export.

The correction records the existing successful turn_ended fact after the
submitted batch drains and before final completion/save. Normal tool batches
still continue to their next request; interrupt keeps its distinct event/state.
Validation admits exactly the payloadless successful end while preserving the
general tagged-union rule and existing failed-end payload checks. This is a small
owner-level correction, not an unconditional idle display or new event framework.
The comments explain why a completed submit differs from ordinary tool output.

An isolated archive reran the **unchanged original independent state probe**
alongside the coder's real retained-state/WebSocket/pending-request control.
Both pass under race in1.219s: immediate and restarted completed peers replay
idle, a parked later request remains in_flight, and its subsequent completion
returns idle. [Independent round1 rerun](../../solutions/evidence/ch26/reviewer-round1-state.txt).
The coder's affected core race (5.235s), history/malformed-load/save race (1.050s),
full GUI race (11.067s), both-module vet and clean formatting support the rest
of this bounded change; unchanged broad checks were not needlessly repeated.

Both round1 mutation outputs fail cleanly for their predicted reasons: removing
the completion fact restores the completed/restarted pending-state failures;
forcing every replay to idle breaks the genuinely in-flight control. Together
with the assessed prior evidence, there are21 clean intended kills over22
mutation attempts, retaining the earlier invalid graceful-cleanup timeout.
The intermediate one-payload-validator failures remain recorded as implementation
failures, not successful tests or extra mutation kills. No new mock was introduced.

The paid live provenance stays at9e2b72a, binary
`a42e9069d3c9b123ede0f44f635b58e8646b5e78b85c615526748b2c915ebfa5`.
Its real fleet coding,15 Astra/high requests, measured tests, GUI selection and
restart evidence remain valid within the previously recorded limits. The
[coordinator's independent wire/artifact audit](../../solutions/evidence/ch26/coordinator-live-check.json)
agrees with the separately inspected GUI sequence. Its historical pending-state
capture remains defective; deterministic real-runtime evidence establishes the
correction, not a paid rerun or rewritten screenshot. Newly completed and restored
conversations are verified; no old-save migration is implemented or claimed.

[Final own-predecessor counts](../../solutions/evidence/ch26/reviewer-final-counts.json)
use the same language-separated tracked-source method and exclusions above.
Student production Go adds1027physical/870active/109comment-only/48blank versus
original1067/838/107/122. Student Go tests add636/585/30/21 and browser tests
75/72/1/2; original Go tests add864/669/93/102. JS/HTML/CSS production deltas
remain unchanged from the initial review, with original removals separately
identified. More complete selection, cursor/durable-result and lifetime evidence
justify the similar active-code size; no arbitrary line quota was used.
Useful explanations and exported contracts are present; final added Go comment
share is about11.1%, still below the rough teaching target, without warranting
syntactic narration or padding. Existing code, actual parent ownership and
optional-module boundaries carry the behavior with fewer special wiring paths
than the reference. The correction strengthens the real state/replay contract
without expanding chapter scope.

Mandatory `.agents/skills/ensemble-coding/SKILL.md` governs this review and handoff.
The coordinator may export this exact accepted source as Edition2ch23 under the
approved map, preserving CodeRhapsody history and all original results. This
acceptance does not itself claim Edition3 authorship or release an author agent.
