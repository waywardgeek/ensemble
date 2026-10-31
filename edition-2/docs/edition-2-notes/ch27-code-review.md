# Original Chapter 27 / Edition 2 Chapter 26: independent code review

Reviewer: `reviewer_astra`, GPT-6 Astra/high. October 12, 2026.
**Accepted with one narrow original-grader fixture exception at initial review.
Zero formal correction rounds. Final source `b5c35a39`; see disposition below.**

## Reading and source boundary

Reread all469 lines of original [Chapter27](../../edition-1/book/chapter-27.md), SHA-256
`5463f060485fc838858ebe44115c70ac062582ba9f3179281df95ddcce9d974d`, and all
current [student-safe preparation/contract notes](ch27.md). Earlier complete
reading of the historical brief, coder review and working notes remains recorded
there. The final chapter requires actual Ensemble actor integration and a
three-process model interview; the historical external subject is not that proof.
No original answer code was read before the student's implementation.

Read all original grader files at revision
`558b29fb098a1f79a60794f467b75762177fbeee`:

| File | Lines | SHA-256 |
| --- | ---: | --- |
| `internal/grade/ch27_harness.go` |315|`cab0928523a1ee489980d65ac571f125d10b9718987996036ae99f2ae0834689`|
| `internal/grade/ch27_probe.go` |553|`1d1d1bf9cbad8c389f1aa553a1c3bf48a0028ddc4f637e39cd95b68b73d8e4ec`|
| `internal/grade/ch27_checks.go` |90|`eeec8a960e14677be504d2399c791c4505b3927f9763ddf11be5125f549aee07`|

The complete workflow/carryover/voice/procedure remain active together from the
full reload recorded at the Chapter23 gate; no intervening compaction. Entire
`.agents/skills/ensemble-coding/SKILL.md` and AGENTS instructions remain active.
The student extends committed main tree `ada66c1572f1e5d22efb6ac59f23ab2b61dd03ac`
in the coordinator's isolated worktree. Shared user main/GUI/fleet/skills changes
and frozen exports are outside this reviewer's mutation scope.

## Grader preflight

The seven original rows total100. The coordinator's public launch/config/API
erratum matches the actual probe calls. No additional input contract or old
implementation recipe is needed at this stage. Run the original grader unchanged
on the frozen student submission; no score or exception is preaccepted.

The external module aliases the submitted source to the first-edition module
name (harness73; probe import). This is a **potential** fixture mismatch, not an
inevitable compile failure: a self-contained public package may still compile
through Go's replacement. Diagnose the actual result and imports rather than
renaming the student's module or installing a compatibility facade in advance.
The dependency detector at harness51–54 only searches `/agent/internal`; it
cannot detect the student's `ensemble/internal/...` paths. Any awarded C6 credit
therefore still needs the actual resolved graph and real outside consumer.

Several labels exceed what the original probe observes. Its provenance check
requires only a header prefix and contained body, not the exact authenticated
identity/header fields. Its delivery handler returns immediately; later queue
inspection after sleeps does not establish that acknowledgment waits for handler
return. Reconnect exercises the running server, not restoration of dynamic state
after a server restart. The probe does not run the interviewer against Ensemble
or inspect actual model input/output. These are coverage limits, not reasons to
invent more grader points, edit the original harness or waive chapter obligations.
The eventual source/tests/live record must establish those existing promises.

## Independent review focus after a coherent freeze

Follow actual owners and construction in gateway server, public client and each
binary. The gateway owns routing/authorization/delivery, never an Agent or model
request. The chapter's public Handler callbacks are the required external API;
they do not authorize internal sibling-service callback bags. Inspect necessary
transport/library ownership under the skill without imposing a preferred package
design. Public client reuse must avoid private runtime imports in the real graph.

Verify authenticated sender provenance independently of group authorization, a
successful authorized route beside refusal, subset registration beside over-grant
refusal, and immutable configured versus persisted runtime policy. Inspect real
handler-return acknowledgment, offline replay/resume and dynamic-state restart,
with observations tied to actual effects and source versions. At-least-once is
not exactly-once; periodic saves do not promise zero loss after abrupt termination.

The actual gateway, Ensemble and interviewer must run as separate processes.
Follow received envelopes into the existing actor and responses back through the
public client, with real model interview evidence and normal process completion.
An external stand-in or a plausible report file cannot establish that integration.
Keep security claims scoped to tested boundaries; primary Noise explanation
corrections already in ch27 notes do not prescribe new cryptography.

After implementation, compare the student's complete review, production/comments/
tests and own-predecessor delta with the matching original answer. Reuse valid
unchanged evidence and require meaningful targeted defects for new claims; inspect
whole failure outputs and cleanup. No broad or paid freshness run is requested by
this preparation. Gateway authorship remains gated on implementation acceptance;
Chapter25 remains reserved.


Student design question, before implementation: whether the server may consume
public `gateway` wire/Noise types, or must split a second public protocol package.
Reviewer response: no split is required solely to change the package diagram.
The chapter explicitly requires the public boundary to avoid private runtime
imports. Server use of its protocol/crypto API is not by itself sibling-service
wiring; actual behavior placement and ownership still require source review.
Keep server policy/queues with their owner rather than putting them in Client.
This is a bounded architecture interpretation, not acceptance of unwritten code
or a supplied implementation recipe. Coordinator informed. Mandatory
`.agents/skills/ensemble-coding/SKILL.md` included in the student handoff.


## Initial formal disposition, October 12

**Accept original27 / Edition2 Gateway26 with the field-spelling exception below.**
Final student source/evidence commit `b5c35a39fce89fdb83bf65682f3b715077a650dd`,
main tree `6d3e3d2e54a4def135f9d5c23ce3242d7d4c0313`, resolves the two
pre-disposition findings. No formal correction round was issued. The student
review was read in full, including its final grader explanation and follow-up;
its earlier pending statements remain dated history. Gateway manuscript/grader
enhancement and the coordinator's integration/export are still separate work.

After compaction, reread the entire mandatory
`.agents/skills/ensemble-coding/SKILL.md` (373 lines) and all four documents:
workflow1–2293, carryover1–981, voice1–632 and procedure1–599, in complete bounded
ranges. The complete original chapter, new implementation and matching original
answer were read; old answers were accessed only after the student's implementation.
No answer code was sent to the student. No credentials or provider were accessed
by this reviewer, and user main/crossover/GUI/fleet/skill changes remain untouched.

### Original result and precise exception

The reviewer executed the unchanged original grader on exact archive `15e0386b`,
tree `f3f75227ad870a1e86402724b96b461425df8c5d`: **0/100 FAIL**, make2/grader1,
1.544 seconds, no timeout. All seven rows stop while compiling the external
probe's `WelcomePayload.QueuedCount`; the submitted public field is `QueueDepth`.
[Raw output](../../solutions/evidence/ch27/reviewer-original-grade.txt)
and [provenance](../../solutions/evidence/ch27/reviewer-original-grade.json)
remain unchanged. No original row established runtime behavior.

Accept the student's narrow explanation: the chapter specifies welcome queue
count behavior but neither it nor the supplied public erratum specifies that Go
field name. Renaming it or adding a grader-only alias is unnecessary. The actual
public client checks authenticated welcome identity/version and empty/nonempty
counts; a count-always-zero defect fails the nonempty assertion. The earlier
suspected module-path mismatch did not cause this run's failure and is not an
accepted exception. No authorization, cryptographic, delivery, persistence,
architecture or actual-Ensemble obligation is waived. Future grader correction
should publish/parameterize the callable contract rather than rewrite history.

### Ownership, behavior and comparison

The actual `cmd/gateway` constructs the routing root, listener and persistence
ticker. Its connection children retain a non-nil common parent interface;
policy, grants, queues, receipts and diagnostics remain at that root. The public
`gateway` package contains independent protocol/client/Noise services with no
private imports. The server uses that public wire boundary, not another spoke's
runtime service. No Agent is owned by the Gateway. The separate adapter constructs
`ensemble.New` and its actual Agent, submits incoming stamped content through the
existing reliable actor request path, then correlates its result back to the
sender/conversation. Shutdown joins callbacks before reply workers and saves the
Agent. The interviewer is a genuine public-only consumer. Resolved dependency
positive controls and the outside-module execution establish reuse beyond a
source-name detector.

Authentication pins keys and checks the encrypted identity; routing independently
requires capability and accepted group membership. Provenance uses authenticated
state and the strongest accepted group, preserving body and conversation identity.
`register-child` refuses replacement and widened grants. Runtime administration is
a distinct configured `group-admin` authority; it can edit runtime definitions
and additional memberships, never bootstrap groups/memberships. This is not an
exception to parent registration's subset rule. The public README accurately
states this administrative boundary and plaintext persistence limit.

The single policy/queue mutex makes authorization and admission one operation;
network writes stay outside it. Connection writes serialize Noise nonces as well
as frames. Queues retain allocated sequence identities after draining, deliver
unacknowledged work again, and apply cumulative receipts only after sequential
handler return. Resume and orderly restart preserve queues and dynamic policy.
The periodic individual-file snapshots do not promise a transactional power-loss
boundary or durable completion of admitted actor work. Those explicit limitations
fit the original periodic/at-least-once scope.

The original answer separates registry, queue, connection and group services into
more independently wired aggregates; the new root keeps the small shared policy
transaction together without a callback service bag. The original source's reply
routing uses a volatile origin map; the new reply carries its conversation/target
and rechecks current authorization. Original Client returns before welcome and
closes its done notification before callback completion; the new public contract
waits for authenticated welcome and joins callbacks. The new real Ensemble adapter
also supplies the final manuscript's integration absent from the historical
stand-in. These are source comparisons, not claims that this review reproduced
every old failure. No speculative scheduler, dashboard, channels framework or new
cryptography was added. No remaining scope or architecture blocker was found.

### Code, comments and tests

Independent counts compare each edition's Gateway additions against its own
predecessor. Method: physical Go lines, blanks and comment-only lines separated;
all other lines (including braces) count active. Mixed code/comment lines count
active. All measured Gateway files are additions, with no moves.

| Gateway scope | Physical | Active | Comment-only | Blank |
| --- | ---: | ---: | ---: | ---: |
| Edition2 production, including three commands/common interface |1871|1475|306|90|
| Original production, including two commands |3058|2155|561|342|
| Edition2 tests |592|543|29|20|
| Original tests |1326|1067|111|148|

[Per-file counts and predecessor identities](../../solutions/evidence/ch27/reviewer-counts.json)
retain the complete original delta list. Original ch26→ch27 also imports unrelated
MCP-WebSocket Go (109 physical/87 active/13 comment lines), speech JS (418 physical),
a shell harness (186 physical), a 2,124-line OAuth design and skill documents;
these are separate carry-in/support artifacts, not Gateway production savings.
New README, dependency files and evidence are likewise excluded from Go counts.

New production comments are17.2% of nonblank lines, approximately one per4.8
active lines; original Gateway comments are20.7%. The difference is not a quota
failure. Exported contracts and adjacent explanations cover ownership, callback
ordering, receipt/admission semantics, routing atomicity, identity publication and
snapshot limits. The initial acknowledgment-barrier comment was inaccurate and
is corrected below. Dense code or removed explanations did not produce the
reduction; fewer policy aggregates and reuse of the existing actor explain it.
The original's more numerous unit cases remain useful historical coverage;
smaller tests are not inherently better or evidence of exhaustive security.

The new tests run actual Noise/WebSocket/client/router and real Ensemble actor
behavior, with a small HTTP inference fake only at the external boundary. They
protect authenticated provenance/refusal, subset registration, handler-held
unacked reconnect, monotonic/offline delivery, dynamic definitions/memberships/
contactability/queues across restart, corrupt-state refusal and actual actor
input/reply. No mocks or production test-only adapter were introduced.

Read all mutation outcomes and their actual failure reasons. The original23
ultimately detected defects retain four initial compile-invalid, broad or
mislabelled attempts. Broad persistence/authentication failures are coupled;
they do not prove every narrower state property. Narrow follow-ups cover those
properties separately. The two review follow-up defects add root diagnostic and
nonzero welcome sensitivity, for25 distinct eventually detected defects, not25
independent proofs or a security certification. Baselines/restorations pass;
cleanup/timeout and historical-source limits remain explicit in the student record.

### Reviewer corrections and final verification

The initial parent interface exposed queues but not diagnostics. The student
accepted this genuine architecture gap and added the narrow root `Logf` service;
connection decode/decrypt/write failures now reach it. A real authenticated peer
sends damaged ciphertext and the test reads the actual root log after teardown.
Removing that call fails the intended assertion. No child-private logger or
callback bag was added.

The original resume test received a message through a channel before its handler
returned, so its subsequent RPC was not necessarily ordered after ack despite
the comment. It now waits for `LastAckedSeq`, whose update follows the actual ack
write, before that RPC. Production delivery behavior was unchanged. These were
pre-disposition corrections with attribution, not an unrecorded formal round.

The reviewer ran exact final source plus a small independent authenticated-client/
silent-handshake shutdown probe under race detection. All public Gateway, server
and actual actor tests pass (1.996 seconds wall); the independent shutdown joins
both peers in916.708 microseconds, rather than waiting for the10-second handshake
deadline. [Command/output/binding](../../solutions/evidence/ch27/reviewer-final-check.json),
[raw output](../../solutions/evidence/ch27/reviewer-final-check.txt) and
[probe source](../../solutions/evidence/ch27/reviewer-shutdown-probe.txt)
are retained. This is one bounded lifetime observation, not an exhaustive timing
claim. Student affected vet/format and baseline/mutant/restored race receipts also
pass. Earlier full core/GUI checks keep their `ab6d0516` provenance; there was no
unchanged broad or paid rerun merely to relabel them final.

### Actual live result and limits

Independently checked the retained three-process interview: Gateway, actual
Ensemble adapter and outside interviewer have distinct PIDs, all exit0 without
forced kill, and the Agent save contains four actual message/request/response
triplets. Every request/completion names GPT-6 Astra; stamped questions and exact
answers agree with the report, and the retained binary hashes match their receipts.
The live run took11.739 seconds on `ab6d0516`. Its source is not relabelled as the
final diagnostic/test correction. [Independent artifact check](../../solutions/evidence/ch27/reviewer-live-check.json)
records the exact hashes and limits.

High reasoning and the explicit Codex route follow the recorded launch and source;
this is not an independently captured full provider wire trace. The four model
answers demonstrate an actual encrypted, authorized Ensemble interview and saved
conversation, not coding-tool capability, measured workflow friction, pricing,
cache grants or general reliability. The no-tools interview's impressions are
model statements. No new paid call was necessary for the local diagnostic,
welcome, schema or corrupt-state corrections.

The implementation gate is complete with the stated original-grader exception.
Coordinator may integrate these owned changes and export exact direct `ch26`,
then release its separate author/grader phase. Chapter25 remains reserved; current
user work, first-edition sources and existing frozen exports remain preserved.
Mandatory handoff: `.agents/skills/ensemble-coding/SKILL.md`.
