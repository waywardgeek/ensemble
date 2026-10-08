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
