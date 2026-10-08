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
