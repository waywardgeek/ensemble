# Chapter 10 implementation status

2026-10-08, milestone 1: Q1–Q3 acknowledged; complete nested semantic grammar
published in ../../persistence-format.md (main/persistence-format.md) and
public signatures in main/session-api.md. This documents the implementation
contract for independent checker preparation; runtime implementation and every
validation gate remain pending. Grammar updates, if needed, will be explicit.

Source predecessor ac55f641e69220a612debbcd6f75e77fa9259b36; accepted plan 7cb8429;
clarifications af5a762. Exact teaching/read ledger is in student-review.md.
No provider/credential access, tests/builds, new workers or unrelated edits.
Free disk at phase opening: 636 MiB. Preserve retained artifacts and caches.

Next: common codec/session declarations, strict byte-preserving codec and owner
validation, store lifecycle, actor controls, clients; then local/independent
checks and a bounded live matrix for separate release. The initial independent
checker remains an initial subset; additional commands await coordinator inbox.

2026-10-08, milestone 2 (continuing): contract/docs commit 44056d18846ba29e8bcd703143fb1b99ba613f62.
Codec grammar correction adds required CalledAt/ReturnedAt to CallState; Unicode
clarification cf73a64 acknowledged. Core implementation remains uncommitted.
Targeted canonical codec tests and two public plain-session resume/import tests
pass. Main module compile-only test passes. Full module/race/independent checks,
CLI/browser persistence and stronger semantic validation remain pending.

Initial command failures retained: first codec-test write used wrong working
path (no test file created); initial test invocation from repository root could
not find main packages. Corrected both. Compile failures exposed new common
interface obligations in predecessor mocks (Codec, ReleaseSession, limit methods),
one pointer receiver and a malformed reconstruction bracket; fixed without
weakening assertions. Corrected targeted persistence/llm/skills/tools tests passed;
`go test ./... -run '^$'` compiled main; `go test . -run '^TestSession' -count=1
-timeout=30s` passed the two local public tests. No actual spin claimed.

Independent public command received, not yet run:
`python3 scripts/edition2/accept_ch10_public.py solutions/edition-2/main --receipt ABSOLUTE_RECEIPT`.
No paid calls or credential reads. Additional implementation/local checking continues.

Milestone 3: docs clarification commit 81aa8cf504ff88f3156b9c815f98e8db7f260be2.
Main and GUI full `go test ./... -count=1` passed (180s/120s timeouts).
First public independent receipt public-initial.json: eight positive groups,
selectors/config failure "returned creation configuration differs" and semantic
positive "identity/equivalent-schema-number" refused at metadata comparison.
Fixed the latter's bytewise schema comparison to canonical equality. Former
needs safe diagnostic detail: canonical store paths on macOS resolve /var aliases;
returned DataDir/LogPath are canonical resolved paths under the current contract.
I have not read fixture code to guess its expectation.

Initial CLI checker initial-checker.json: 12/93; offline-inspect failed and many
negative rows blocked by that control. No subprocess stderr is included in the
receipt. Own independent TestCLISessionRestartLocal (all CLI builtins, protocol
create/resume then session inspect) passes. Coordinator response requested:
please release safe failing-command/exit/stderr diagnostics or repair diagnostic
coverage, without exposing fixture implementation. This is not a request to waive
checks. Continue other local implementation meanwhile. Disk now 354 MiB; no
unrelated artifacts removed. Retained initial binary evidence/ch10/ensemble-ch10-cli.

Milestone 4: skill export/import/resume local positive added and fixed an unsorted
owned snapshot ceiling (predecessor log ceiling already sorted). Checked-write
controls for create/write/short-write/sync/close/rename failures now pass, preserving
prior bytes/ack and releasing worker gate. Lock inode lifetime and shared export/
save gate tested locally. Pending one-shot limit survives resume, then is consumed
before an unknown tool's dispatch; snapshot inspection passes before/after.

Active Q4 in student-review.md: local whitespace-only RawUsage reproduces initial
CLI record-7 prefix mismatch. Inherited event serializer compacts RawMessage but
new owned Clone retains its input bytes. Need exact recorded-versus-provider-byte
framing interpretation before fixing that path. Continue independent local work.

Milestone 5: additional compact request-coordinate witness and CheckpointContext
API documented. Public corrected selector check passed (public-selectors-corrected.json);
initial semantic equivalence rerun passed (public-semantic-second.json). Initial
failures remain retained. Awaiting Q4 publication before acceptance-byte fix.

## Prepared-event coherent boundary and independent results

Reloaded the entire mandatory coding skill after compaction. Q4 implementation now passes the initial CLI black-box command: `cli-prepared.json`, 93/93, executable SHA256 `4ad5c584d170734c052db321fe34dcc7a256fad9243fcf731ba823abc27f3c30`. Original binary and failing receipt remain preserved. Main full tests/vet and GUI full tests/vet pass; new exact accepted-Raw regression and WebSocket publication/disconnect tests pass. These are fake/local results only.

Coordinator response: `public-prepared.json` stops at fixture vet before runtime tests: `ch10-session-limits_test.go:247:34: invalid operation: record.Observation.Event != nil (mismatched types common.Event and untyped nil)`. This is the published value-typed observation contract; please repair the independent fixture. No checker source read. This run is the requested coherent prepared-event boundary; no source changed while it ran. Q5 above-maximum watermarks remain paused awaiting the published teaching pin; other work continues. Browser, client/lifetime and physical read-bound commands are next.

Milestone 6: browser-initial.json passes6/6, clients-initial.json passes9/9 and record-bounds-initial.json passes38/38 using local controlled fixtures. All nine nested example modules passed vet/tests again (nested-prepared-checks.json). Changed-file gofmt produced empty output (format-prepared.json). Main and GUI race runs complete successfully. Inherited diagnostic is running; full corrected public fixture run is running against unchanged source. Q5 full published chapter/direct response read/verified and acknowledged in student-review.md; its policy change follows this prepared-source checkpoint. Current source is not final: indexed collection admission, strict maxima and further semantic/fault coverage remain under review.

First corrected public runtime run retained in public-prepared-corrected.json:15 groups passed, SkillsReplay/anthropic timed out. Stack evidence reveals a real new lock inversion: append holds Agent.mu and calls Jobs.PendingLimits; report holds Jobs.mu and calls Agent.Workspace. Fixing actual owner lock order, not fixture timeout. Inherited `make grade-dir CH=11 DIR=solutions/edition-2/main` returned0/100, with GUI-server startup/old invocation failures (full inherited-grade-initial.out); this is a diagnostic incompatibility to relay, not a waiver or inferred Chapter10 runtime acceptance. No checker implementation read.

Milestone 7: initial prepared runtime/evidence committed at41a5e7266570a449e7470a9c5b06da3e601f0008, with prepared-source-manifest.json binding its changed source and retained binary hashes in independent receipts. This preserves the pre-Q5/lock-fix attempt, including its timeout. Targeted `TestCh10PublicSkillsReplay` now passes all three local API variants in public-skills-lockfix.json after the Jobs accepted-limit lock was separated from report-worker state and watch releases Agent state before live ownership lookup. Deterministic interleaving regression is added (initial test compile typo j.Write corrected to actual owner write helper); no timeout extension. Q5 removes unsupported job-floor relaxation. New derived Context.Index is excluded from all wire state and rebuilt once at import, eliminating per-event full conversation scans for collection admission/unresolved calls; Jobs indexes prior limit transitions. Bounded snapshot encoding now avoids repeated marshal/parse allocations and prechecks individual escaped strings. Full public rerun currently in progress against this coherent revision. No provider/credential access.

## Revised local handback boundary

Grouped runtime corrections are complete for coordinator completeness review.
Current docs: persistence-format.md (complete strict nested state_version1 grammar),
session-api.md (public signatures), and live-matrix.md (proposed66-generation-request
ceiling across three providers, plus separately released discovery; nothing run).
Q1–Q5 are closed; no outstanding teaching question. Do not interpret this as runtime,
paid-spin, historical comparison, prose or final immutable chapter acceptance.

Latest command receipts: revised-module-checks.json includes full vet/tests in main,
GUI and all nine nested examples; revised-format.json has empty gofmt output;
revised-race.out and revised-gui-race.out pass. revised-star-imports.json checks
all present implementation spokes against common and records CLI composition.
cli-revised.json passes93/93; public-revised.json passes all16 groups;
clients-revised.json passes all9 groups. Initial browser6/6 and read-bound38/38
remain applicable to unchanged assets/read-side code; no redundant large-file
repeat. oracle-self-test.out passes. New local exact64MiB/one-over write preparation,
log-byte/count admission, checked storage failure/retry, disconnected checkpoint
waiter, accepted process-job tail, strict maxima and lock-cycle regressions pass.

Inherited diagnostic remains failed at legacy GUI startup assumptions (retained
inherited-grade-initial.out). Remaining source-bound independent fault/concurrency,
allocation and structural controls are coordinator work to release after this
checkpoint; no claim that the published initial subsets prove the full chapter.
The source will remain unchanged for that review unless a concrete local finding
requires correction. No credentials/provider/discovery calls are authorized here.

Immutable revised source/check receipts:122b04a57e7c6ea158901d00bfc0f1a3a8c75329.
Retained revised CLI SHA256 f5589f003b1b2f3a496e80d941e36749005753e603ca6178a658e929a16bb7e1;
GUI SHA256450e43048a42da9cd974615ab2ed5fff9d2d59e0c336afcb55f13114bf41d0dc.
Complete source/build mapping is revised-source-manifest.json. Earlier initial
runtime/evidence checkpoint41a5e7266570a449e7470a9c5b06da3e601f0008 and all its
failures are preserved. Contract milestones were44056d18846ba29e8bcd703143fb1b99ba613f62,
81aa8cf504ff88f3156b9c815f98e8db7f260be2 andb92692955edbf2ecf64bcfd0fc135087600176a0.
Accepted initial plan remains7cb84290af3e8bac5359339f8ab5d4e7e941bf74.
Coordinator may run remaining independent coverage against122b04a without concurrent
student source mutation. This final status-only update changes no runtime or binary.

## Append-fault repair resumed after build-space maintenance

Supervised interruption preserved the uncommitted repair and receipts. The aborted
status write did not land; this records the facts now. Local pre-fix regression
append-fault-local-baseline.out reproduced both missing-checkpoint creation and
existing-checkpoint inode replacement, explicit checkpoint acceptance after fault,
and nil repeated Close. First test adapter timed out because it called Config
while append held Agent state (append-fault-local-before.out); corrected test by
capturing its immutable selected log path before installation. No production
assertion was weakened. Post-fix build initially failed `no space left on device`
(vet.cfg/mkdir); redirected append-fault-local-after.out was empty and is not a
pass. Student removed no caches/artifacts. Coordinator cleared only build cache,
released exclusive build stage with1.4GiB; current retry is running.

Reloaded complete mandatory skill again and read only authorized inbox additions.
Full permitted architecture and3a9e5b7 chapter/direct response were already reread
and verified during this same repair; no completed source work was redone. Frozen
fault-check command f572cd3 received: run only after repaired source commit,
first TestCh10ReviewAppendFailure, then all four groups. Other graders deferred.
No live/provider/discovery/credentials released. Bounded plan review feedback is
received for a prose refinement after this scoped repair validation.


Append repair local validation complete: targeted race regression passed (append-fault-local-retry.out). append-repair-module-checks.jsonl retains all25 successful commands: empty-output gofmt on five changed Go files; vet/tests in main, GUI and all nine nested examples; full race suites in main and GUI. No further source change followed those checks. New tests distinguish absent checkpoint creation and existing checkpoint inode replacement, retain original append error on repeated Close, and prove corrupt-log refusal plus released path/SessionID ownership. Ordinary invalid input still permits a later valid append/checkpoint/clean close.

The scoped source checkpoint is next, followed by the coordinator's frozen targeted and four-group fault commands. This commit-before-check order follows that command's required immutable source identity. Old retained CLI/GUI binaries still belong to122b04a; they have not been relabeled or rebuilt for this source. Live-matrix.md now contains exact proposed prompts, process/store schedule, deadlines/attempt accounting, concrete GUI policy/preference actions, public offline branches, truthful local limit seed and pending support-freeze controls responding to c9ba32c. No live support build or live request is claimed. Other independent graders remain deferred until this build stage is released.


## Append/Close repaired handback — build stage released

Immutable repair/source/local evidence:8882a18cf98e9a4b70afccfbe980f6344630aae6.
The exclusive student Go build/test stage is COMPLETE and RELEASED. No student
compiler/test/checker is running or planned in this handback; coordinator may
start the deferred independent lifecycle/bounds stages. Last observed disk1.0GiB.
No cache maintenance was performed by the student.

Both source-bound frozen fault commands exited0 against that exact commit:

```
python3 scripts/edition2/accept_ch10_faults.py solutions/edition-2/main --source-commit 8882a18cf98e9a4b70afccfbe980f6344630aae6 --run '^TestCh10ReviewAppendFailure' --receipt /Users/bill/projects/ensemble/solutions/edition-2/main/evidence/ch10/append-repair-independent-targeted.json
python3 scripts/edition2/accept_ch10_faults.py solutions/edition-2/main --source-commit 8882a18cf98e9a4b70afccfbe980f6344630aae6 --receipt /Users/bill/projects/ensemble/solutions/edition-2/main/evidence/ch10/append-repair-independent-all.json
```

Both receipts report passed=true, source_unchanged=true, checker_unchanged=true;
format output is empty, overlay vet and race tests pass. The full command passes
all four groups: seven checkpoint I/O outcomes, canceled waiter retaining captured
write, close joining writer before unlocking, and terminal append preserving the
checkpoint. Full stdout is separately retained in append-repair-independent-all.out.
These new results supersede the specific append defect, preserving original frozen
122b04a failures. They establish this local subset, not full Chapter10 acceptance.

Other earlier coverage retains its original revision/binary association (CLI93,
public16, clients9, browser6, read-bound38); this repair did not rerun those deferred
suites. All25 local module/format/race command results are in
append-repair-module-checks.jsonl. Test-fixture timeout, pre-fix failing regression,
and disk-blocked attempt remain preserved and explained above.

No open teaching question or architecture conflict. Public API and strict codec
remain session-api.md and persistence-format.md, unchanged by this repair. The
live-matrix.md proposal now answers the five relayed c9ba32c plan requests, within
66 generation/3 discovery maximum; executable support freeze/identity-negative
controls and coordinator completeness review remain prerequisites. No credentials,
provider/discovery calls, actual spin, historical comparison, author runtime
confirmation or final immutable chapter release occurred. Retained CLI/GUI binaries
still represent122b04a and require newly bound builds before a later live stage.


## Support preparation readiness — compiler slot remains with root

Prepared only evidence/ch10/support/ and corrected live-matrix.md. Runtime8882a18
and evidencee307d79 remain preserved; no runtime edits or new runtime acceptance.
Coordinator's CLI policy correction is reflected: raw0/effective16 in A/D1, proxy
cap3 separate; GUI current1 versus actual historical default. Browser support now
measures native/service admissions and actual owned queue/pause state. D bindings
are explicitly empty because the CLI has no scalar-binding selector.

Concrete prepared files: bounded schedule/catalog/bindings; identity/source/build
preflight; durable capped loopback relay; deliberate PTY and browser recorders;
public C/D/inspection consumer module; original sealer/limited verifier; local
three-wire fixture and Python/JS controls. See support/README.md for the exact
component boundaries and next-stage commands. Driver currently refuses external
origins before launch artifacts; provider-facing activation/discovery remains a
later reviewed support step, never an implicit release.

support-prebuild-checks.jsonl:14 Python controls pass (budget restart/concurrency,
step/row/provider caps, timeout/disconnect/no-retry/redirect behavior, credential
canary redaction, identity/dependency/launch/extra-source refusal before derived
writes, and synthetic wire construction). JS owner-probe positive plus seven
intended negatives pass. JS syntax and Python compile() pass; changed Go file
format output is empty. Original earlier outputs retained, including the harmless
fake-backend teardown diagnostic fixed before the final clean run. No Go compiler,
vet/test, browser integration, credentials, discovery or real-provider action ran.

READY for the next coordinated support build/local integration stage. Still needed:
compile/vet/test the new nested consumer, build freshly bound CLI/GUI/consumer,
collect actual complete dependency/build/browser identities, exercise valid real
preflight plus intended identity mutations, then drive local PTY/browser/public
C/D integration and review any source/fixture/observation failures. No actual
binding is created from planned binaries. support/README.md is a reproduction
plan, not a success transcript. Root retains compiler ownership; this worker starts
no Go build and does not wait for another repeated fault run. Full chapter/live/
historical comparison/author confirmation/release gates remain open.

## Grouped retained runtime repair — local freeze preparation

Acknowledged verified dd1111e teaching before implementation. Runtime now permits duplicate JSON members only within designated argument objects, retains ambiguous argument text through the strict semantic Raw wrapper, and uses exact-text correspondence for duplicates. Chat Completions original argument STRING bytes occupy optional typed Part.arguments_text; opaque target restrictions remain unchanged. Registry emits empty name for ambiguous duplicate name members and can validate recorded limit syntax without installing live handlers.

Own localhost tests exercise Anthropic/OpenAI/Gemini in standalone and session modes: setter, duplicate load_skill, untruncated default-limits read, valid load, continuation, paired controlled refusal/no invalid Job/no invalid Skills transition, one consumed fact, actual captured request versus public replay, latest/older-tail/null/full-log checkpoints, checkpoint inspection, semantic import/tail and resume with once-restored usage. Strict codec tests distinguish malformed/nonobject/scalar/structural duplicate refusals, exact decimals, changed duplicate spelling and unique replacement. Public malformed replay correspondence refuses without terminal Close fault.

Command ledger: retained-repair-commands.jsonl (start/end and free bytes per command), separate original output files. First test setup omitted setter capability (fixed); next found offline limit validation wrongly required handler selection (fixed). First full suite caught opaque alias restriction (fixed with typed replay field); two inadvertently repeated unchanged attempts are preserved. New negative controls independently remove string retention or restore old strict argument equality and fail their intended behavior; restored-source root vet/tests and empty gofmt output pass. All 12 discovered Go modules, including the separate support consumer, passed vet/tests; root and persistence targeted race passed. The first race selector matched no internal/llm/tools tests, so separate complete spoke race follows. No original receipts were edited. No cache maintenance was needed.

This is a runtime freeze before revision-bound independent checks and local support build/actual-binding integration; compiler stage remains owned for those authorized tasks. No paid/live or chapter-acceptance claim.

Complete internal/llm and internal/tools race suites also passed (spoke-race receipt); the no-match caveat above is resolved by actual execution. Final root gofmt output is zero bytes and final root vet/tests pass.

## Frozen checks at 56d8aa4805ba8f093bd66c1f2b1e9cbdca64a58a

Fresh bound CLI/GUI/public-consumer built successfully (retained-repair-build/binding.json and build-commands.jsonl); complete source map, actual executable hashes/build metadata and browser dependency tree preflight passed. VCS reports modified=true for the shared outer working tree; association checks verify all owned source bytes at the immutable revision. Original older binaries are untouched.

Published black-box commands: accept_ch09 51/51; accept_ch10 93/93; accept_ch10_public, accept_ch10_clients and all accept_ch10_faults groups passed. Exact commands/start/end/receipts remain under retained-repair-*; faults confirms source_commit56d8aa4 and unchanged source/checker hashes. No checker source read.

COORDINATOR REVIEW REQUEST: accept_ch09_management passed135/138, with only anthropic/duplicate/strict-ack, openai/duplicate/strict-ack and gemini/duplicate/strict-ack failing. Each actual tool_returned sequence65 is controlled is_error:true with the consumed-limit note then {"error":"invalid_skill_arguments","name":"","revision":0}; sequence64 is its paired call. The published dd1111e teaching explicitly requires empty name for duplicate name members. No change to runtime/expectations is justified from this unexplained checker failure. Please independently review the checker against the current clarification and supply any correction/rerun command. Its source was not inspected. All subsequent default-limit/continuation checks pass. This is an unresolved independent-check outcome, not a full-pass claim. Support integration continues locally without paid access.

Local support observations: A1/A2 and D1 actual CLI PTYs succeeded with labeled localhost fixture output; C public dual-Agent/import path passed after correctly closing GUI owner, including immutable origin and pre-origin history_unavailable; D local seed/relocated resume observed one consumed pending limit. Actual browser persisted policy1/font18 across process restart, exercised reconnect/Page close/reopen/tab close, measured no restored native speech/queues/pause and checkpoint after terminal EOF. Fifteen identity negative controls passed from the real build/launch parent. Initial GUI attachment hit macOS script output buffering; adding script -F flush. Student orchestration errors were one premature C open (session_in_use) and waiting for policy_limit instead of actual round_limit, causing a bounded browser timeout; original receipts retained. No browser prompt was retried.

Offline support consumer reported branch mismatch. Existing evidence records that failure; adding detailed public-field comparison diagnostics without changing equality expectations to identify its owner before correcting it. All work stays local.

The offline mismatch is now protected by an added public regression that failed before the correction. Public Context snapshot copies normalize absent structural collections; original events and replay Raw bytes retain exact accepted representation. Identity schemas use their documented semantic comparison, separately from exact replay values. New support comparison controls reject altered Raw number lexemes, Raw whitespace and state coordinates. Root vet/tests, root targeted race, consumer vet/tests and GUI tests passed after this correction; new argument replay field also refuses inactive non-call variants. Preparing a new immutable revision and fresh binding/checks; do not associate prior binaries with this corrected representation.

## Compiler slot released — final runtime57d4aac

Current immutable runtime/support revision57d4aac3d26edcb78dc8d9fd27d7d0e1cce0ebf4, fresh complete binding retained-repair-build-watch/binding.json. The offline public consumer now passes identical render/Context/history/usage/Skills/watch/request comparisons with endpoints disabled and zero provider calls. Checkpoint and full-tail call/result presence is repaired without relaxing comparisons.

Current black-box results: Chapter9 main51/51; Chapter10 CLI93/93; public and clients passed. Fault format/vet passed but the first race link failed with no space left on device; no runtime verdict was obtained from that attempt. With no compiler/linker active, the explicitly authorized go clean -cache cleared ONLY regenerable build cache (source/binaries/evidence/module downloads untouched); retry watch-faults-after-cache passed all groups with unchanged source/checker identities. Both attempts remain. Free bytes after retry710627328. No more Go builds/tests are planned in this student stage; compiler slot is released now. Existing-binary local support runs and Python/JS receipt work continue.

Management remains135/138, the same three duplicate/strict-ack failures with documented empty-name payload. Independent coordinator/checker review remains required; no checker source was opened and no test was weakened. No live/credential/discovery release or chapter acceptance is implied.

## Final local handback — compiler remains released

The complete source-specific report is
[retained-repair-handback.md](retained-repair-handback.md). Runtime/support remains
57d4aac3d26edcb78dc8d9fd27d7d0e1cce0ebf4; final source/support/build/dependency
preflight passed again without Go. All completed final-revision local runs were
sealed/verified. Actual PTY/browser/public paths cover each localhost wire route;
33 proxied generation attempts (11/route), zero discovery, plus6 explicitly local
seed exchanges. All three final offline consumers passed exact public comparisons
with endpoints disabled. Final browser OpenAI/Gemini receipts verify applied-before-
saved acknowledgment, policy/preference acknowledgments and measured speech
boundaries. Earlier Anthropic browser/restart receipts keep their actual earlier
source identities. Final actual-build identity controls passed15 intended refusals
before derived writes, with restored-parent preflight passing. No live evidence
or source identity is inferred for an old/planned executable.

The three management acknowledgment failures remain the sole reported unresolved
published-command result here; broader independent coverage and live completeness
review remain coordinator gates. Final inbox read has no new resolution. No further
Go work is planned in this stage. Compiler slot was released before receipt work;
this handback does not reacquire it. No credentials/provider requests, push,
historical comparison, chapter export/tag or acceptance occurred.
