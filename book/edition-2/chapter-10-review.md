# Chapter 10 contract review

Coordinator review, October 8, 2026, of draft `9181683`. The complete chapter,
outline and preparation evidence were read. This is a contract review, not
implementation acceptance. Chapter 9 remains an unbuilt predecessor; no
Chapter 10 student or live demonstration is released by this record.

The draft keeps fresh construction distinct from deliberate session resume,
assigns SessionStore to Agent, and preserves one durable append path. Its
shutdown, unfinished-work refusal and historical-job distinction bring the
later lessons forward without inventing process recovery. Default human
startup and explicit standalone selection have separate, reviewable behavior.

## Four proposed decisions

These are coordinator choices within the authorized rewrite, not additional
rulings attributed to Bill.

1. **Request allocation:** accept the durable request index and snapshot cursor
   lower-bound rule. A canceled caller's transient handle need not become a new
   durable admission event. Reconstruction must prevent reuse of every recorded
   request identity; it cannot claim to recover invisible canceled admissions.
   Preserve that distinction in prefix-equivalence checks.
2. **Imported origin:** accept immutable `origin.json` as the seed for a
   snapshot-only store. It is a read-only reconstruction dependency, not another
   append writer. Later checkpoint replacement must retain and validate the
   seed. Missing or partial imports refuse rather than inventing recovery.
3. **Plain-mode handlers:** accept exact installed-definition and visible-set
   compatibility for session reopening. Existing fresh Agents retain their
   earlier behavior. Definition equality does not attest implementation bytes
   or create a security boundary; the draft says this explicitly.
4. **Skill provenance:** accept compact ordered transition provenance referring
   to immutable activation material. A snapshot must validate earlier visibility
   and retirement, not infer legitimacy from the final active roots. Live import
   also checks the caller's compatible frozen catalog and candidate bytes;
   historical inspection remains independent of source files.

## Clarifications before checker preparation

- Make the job watermark explicitly session-local: the greatest historical job
  handle represented in that session, including handles outside the GUI window.
  It is not an exact serialization of Ensemble's shared allocator cursor.
  Other Agents and occupied output files can advance that allocator. Resume
  raises the live allocator to at least the restored maximum without lowering
  its existing value; snapshot/replay comparison must not require equality with
  another Agent's allocations. The current source's `Ensemble.AllocateHandle`
  confirms that the counter belongs to the application root.
- Specify canonical-number behavior sufficiently for independent importers.
  In particular, preserving identity integers must not leave handler-schema or
  arbitrary payload numbers subject to a new lossy conversion during hashing.
  Two distinct accepted values cannot gain the same compatibility identity merely
  because both pass through binary64. Publish the normalization rule and examples
  for integral decimals, exponent spellings, large integers and negative zero;
  distinguish a canonical hash representation from stored original payload bytes.

Both clarifications are resolved in `6cf8660`, reviewed by the coordinator.
The job watermark now denotes this session's historical maximum and only raises
the root allocator's floor. Canonical numbers use exact decimal coefficient and
exponent normalization, with examples preserving large adjacent integers;
original payload bytes remain unchanged. The same revision removes a draft
JavaScript-safe identity ceiling that conflicted with Chapter 9's uint64 domain.
Overflow is refused before allocation or mutation, independently of file bounds.

The complete contract is accepted for independent checker preparation. This
does not release a student: accepted Chapter 9 and a published Chapter 10 checker
remain prerequisites. No runtime, live result or historical rewrite is claimed.

## Independent student ownership-plan review

October 8, 2026, after the Chapter 9 checkpoint. `/root/grader_ch05` read all
384 lines of the fresh student's plan at
`7cb84290af3e8bac5359339f8ab5d4e7e941bf74`, recorded in
`solutions/edition-2/main/evidence/ch10/student-review.md`. This was a design
review against the new teaching, architecture and coding skill, before runtime
implementation. No first-edition persistence implementation or future chapter
requirements informed this check. The reviewer's earlier grading, Chapter 9
comparison and isolated streaming-maintenance exposure remain disclosed in the
Chapter 10 grader review; this is not a cold-student evaluation.

The plan is sufficient to release implementation **after the student rereads
and acknowledges the published Q1–Q3 clarifications**. Those concern session
record bounds and inherited skill-refusal precedence; actual imported
anchor/tail history versus the saved display window; and exact replay bytes
versus purpose-specific canonical equality. This acceptance does not resolve
those questions through unpublished fixture assumptions. No additional
architectural blocker was found.

The owner graph preserves Agent authority and the import star. An Agent-created
stateless codec can serve standalone validation without a SessionStore or
injected sibling service. Private validation owners, accepted-event job capture,
off-actor checkpoint work and close ordering fit the contract. The public
open/inspect/export/import/checkpoint seams support independent checks. Their
names and the conservative shared export/save worker gate are permissible
design choices, not newly required spellings.

Retain two implementation safeguards:

- Private validation owners remain under the real application owner but inert
  and unregistered. Validation must not start processes or an Actor, charge
  replay as usage, write policy files or mutate shared allocator authority
  before the candidate has passed validation.
- GUI delivery-before-acknowledgement waiting belongs to the connection
  lifetime. Overflow or disconnect releases that wait without parking Actor
  or the checkpoint worker; losing the connection cannot undo a committed save.

These safeguards explain how to preserve the planned boundaries; they do not
add public APIs or another owner. The complete nested semantic codec grammar
must still be published before semantic fixtures are written. This record
accepts the ownership plan conditionally, not implementation, local tests,
provider runs, historical comparison or final chapter readiness.

## Independent accepted-byte clarification review

October 8, 2026. `/root/coder_ch08` independently reviewed author freeze
`c5ad6c04a67c0e2eeafde77f52d90dfdd7695d8e`. This is a narrow teaching review
of the prepared-event boundary, separate from the preceding coordinator and
ownership-plan reviews. The reviewer previously implemented Chapter 8, prepared
early Chapter 9/11 checks and reviewed later contracts; this reviewer neither
authored the correction nor implemented Chapter 10. Earlier Q1–Q3 and Unicode
proofreading remains recorded in chapter-10-clarification-review.md.

The complete three-file author delta/direct response, surrounding §§10.3/10.8,
and the retained initial-prefix diagnostic were read. Focused predecessor reads
checked Chapter 2's JSONL, typed parts and raw usage, Chapter 6's transient/accepted
response and retained stream JSON distinctions, and Chapter 9's controlled
whole-record preflight. Full current voice/procedure remained loaded from the
immediately preceding review, without intervening compaction. Working author
files match the frozen revision. No old implementation or checker source was read.

The clarification closes the observed representation gap. New session admission
validates first, prepares one bounded final event encoding, derives owned raw
fragments from it, appends exactly those bytes, and only then applies/publishes
the accepted event. The accepted-event ordering does not turn Chapter 6's earlier
transient stream fragments into durable facts. Preparation may change formatting
and escape spelling while preserving number lexemes, decoded string/manual/
signature bytes, object-member and array order, and opaque semantics. Canonical
hashing and floating-point conversion cannot substitute for that preparation.

Once accepted, the recorded fragments define the exact state/watch/snapshot/
replay comparison. Existing imported records already define their own accepted
bytes and receive no read-side normalization; original physical bytes still
determine their read limits. The strict scalar validation and dedicated raw-JSON
wrapper rules remain applicable. Legacy standalone behavior is explicitly outside
this new session-write boundary. Skills candidate and complete-record preflight
still produces controlled skill_too_large before storage admission; bounded
preparation cannot defer that refusal into a terminal writer failure. Separate
session file/count/write failures retain their existing persistence disposition.

The direct feedback and evidence consistently attribute this integration gap to
the coordinator/grader diagnostic, separately from the student's Q1–Q3. The read
receipt shows successful local creation, untouched inspect refusal at record 7,
and a successful separate checkpoint-free rebuild, with the two raw_usage spacing
differences in context and saved watch state. It records one local HTTP request
and unchanged originals. This reviewer inspected that retained receipt rather
than rerunning the binary; neither the receipt nor the author response proves a
repair of current source or any real-provider result.

No new consequential contradiction was found in the affected teaching. The sole
existing fenced JSON fixture is unchanged by the complete diff. The retained
prose executable independently passes all hard rules at 7,681 words; length,
negation and person-gap warnings are soft. The new paragraphs explain an observed
failure and its admission boundary without inventing a successful restart.

Reviewed SHA-256 values:

- chapter-10.md: `41ef91e7fad0308f32bd874c8cb952fddc8900d41f3534c847183adb0e66a26a`.
- chapter-10-evidence.md: `b205861a3d915c7e0f96d1d118704dd7016a1b393be80da7f4c8e39e4d4fe102`.
- chapter-10-student-feedback.md: `0b4ccab128da68c300c9a831693e2037ebc9e5291949e25bde8bc2318a9ca8d2`.
- ch10-initial-prefix-diagnostic.json: `00305cf60c4cb33c7685edf4fe9a68a041e84dc62d53947233ebe31d2c4c0a1a`.

Disposition: accepted-byte clarification proofreading accepted. Implementation,
affected controls and student confirmation remain separate. No runtime edit,
new build, checker edit, provider call or implementation acceptance occurred.

## Independent declared-coverage completeness audit

October 8, 2026. `/root/coder_ch08` read the complete current Chapter 10 contract
at `c5ad6c04a67c0e2eeafde77f52d90dfdd7695d8e`, the complete grader coverage
record through `d2d84811b6338a8ed3a52e2a5ff412bb162f7539`, and the validation
record. This audit compares printed promises with **declared** checker coverage.
It did not inspect mutable student code, checker implementations or historical
persistence answers, and did not execute runtime tests. Prior exposure remains
as disclosed above. Preparation descriptions and reported runtime results are
not promoted to independently reverified acceptance by this audit.

The grader directly confirmed its current engineering scope: actual Chrome
Page/Connector session and checkpoint projection, exact counters and historical
job/no-restored-speech display; basic public limit set/consume/restart and ordering
mutations; then pre-replace filesystem refusal. More extensive I/O instrumentation
awaits coherent frozen source. This audit adds no competing implementation work.

### Existing commands versus outstanding execution

| Command | Declared protection | Current evidentiary status |
|---|---|---|
| `accept_ch10.py CLI_BINARY --receipt FILE` | 93 rows: local provider session creation/resume, selection subset, outer identity/hash/corruption, plain System, secret canaries and actual lock exclusion | Released partial command. Original student failures and Raw-usage diagnostic are retained. Corrected numeric-token fixture controls pass, but do not establish current runtime acceptance. Rerun affected rows against a bound coherent binary. |
| `accept_ch10_public.py SOURCE --receipt FILE` | Nine groups: public selectors/owned capture, reservations, prefix-free import/origin, System, inert inspection, nonempty-tail/null-rebuild/import request equivalence on three local APIs, held-HTTP busy, nested semantic cases | Released partial command with reported student positives and separately reported fixture/runtime corrections. Bind final results to final source; do not label already prepared groups missing merely because reviewer execution is pending. |
| `accept_ch10_clients.py CLI GUI --source-directory SOURCE --receipt FILE` | Nine prepared CLI/WebSocket/lifetime groups: checkpoint ordering, standalone refusal, held-HTTP busy, terminal detach, signals/lock release and a completed historical Job | Prepared but unrun according to the read record. Oracle controls are not actual CLI/browser runtime positives. This command does not claim DOM, running-job or complete fault coverage. |
| Initial self-test and semantic/client/serialization preparation controls | Canonical/shape oracles, deliberate outer/semantic cases, precise serializer regression control | Protect fixture construction. They cannot replace genuine accepted exports, intended runtime refusals or implementation deletions. |

These commands do not yet form a complete Chapter 10 gate. Local Messages,
Chat Completions and generateContent shapes test adapter behavior without real
provider calls; the later bounded real CLI/browser/public feature matrix remains
a separate obligation.

### Prioritized gaps to close

Priorities below order acceptance preparation, not permission to omit lower rows.
“Missing” means no complete protection is declared in the reviewed command
coverage, not that an inspected implementation necessarily lacks the behavior.

| Priority / contract promise | Concrete missing or incomplete distinguishing control | Command destination / coordination |
|---|---|---|
| P1: Skills and current authority (§§10.3–10.4, 10.7) | Genuine skill-mode snapshot/tail/import equivalence must compare retained manuals, active/retired material, hints/anchors, grants and usage. Same-content relocated catalog passes; changed inactive definition, binding, primary, ceiling or admitted activation material refuses. Recompute hashes for semantic forgeries. Public append cannot adopt a session, insert initializer/anchor, reset a counter or fabricate permitted state. | Public runner extension or separate public authority command; not yet covered by the declared plain-session roundtrip. Grader confirms catalog authority remains open. |
| P1: Accepted raw bytes and replay (§10.3 correction) | Legal spaces/physical JSON formatting, escape spellings, unsafe integers and alternate number lexemes reach newly prepared raw usage/arguments/opaque fields coherently in state, watch, checkpoint and replay. Imported accepted fragments remain exact, including original read-size accounting. Distinguish semantic schema equality from byte-different replay data. Include compatible opaque positive and incompatible-target refusal, with per-producing-model usage counted once. | Extend genuine public roundtrip/semantic controls after the new teaching pin; Unicode and outer-number oracle cases alone do not protect this cross-path boundary. |
| P1: Capture, writer and shutdown (§10.5) | Busy parents for queued prompt, response/report/input worker and unresolved call/material, plus settled pending-hint and paired-running-Job positives. Hold checkpoint I/O while an accepted Job fact becomes tail; controls and another Agent remain responsive. Prove one worker, joined/repeated close and disconnect survival. Inject write, sync, close and replace failures, before/after commit; append fault suppresses later accepted state/final checkpoint while releasing resources. | Grader's planned instrumentation on coherent frozen source; held-HTTP busy and pre-replace refusal each cover only a subset. No new public fault API is required. |
| P1: Strict store/import and unfinished-state boundary (§§10.2–10.4, 10.6) | Every file-combination row, symlink/nonregular leaf, partial initialization/import residue and failed-construction reservation release. Missing/mismatched origin remains refused even with a newer checkpoint; anchor hash/sequence/watermark tampering has valid parents. Active turn/response, unanswered call and pending human dialogue refuse resume with zero recovery effects; settled interruption/error/hint cases pass. Import refuses nonempty destination and unsettled candidate without damaging source. | Extend initial/public commands. Existing snapshot-only success and outer corruption do not establish this entire refusal matrix. |
| P1: Physical and allocation limits (§10.8) | Exact/+1 session versus standalone/header records, actual LF/whitespace/escaping, partial final record and bounded generic first-record detection. Whole encoded Skills overflow remains controlled before session storage exhaustion. Exercise file, checkpoint/origin, canonical-state, nesting, collection/definition limits and full uint64 allocation groups with safe positive parents. Preserve original files and correct read/admission error dispositions. | Separate bounded limits command or targeted public extension; grader confirms this remains open. Plan disk/memory use and resource controls before large fixtures. |
| P2: One-shot limits and historical Jobs (§10.6) | Beyond set/consume/restart: second/invalid setter, unknown/disabled/management/supervision/interrupted attempted calls; malformed response, pause and public controls consume nothing. Full replay and snapshot agree; dispatch/result witnesses reject forged order outside the saved window. Restored running/done/killed handles reject supervision and cannot collide with a new live Job. Test a historical maximum outside the window, an already-higher root allocator and occupied artifacts; no artifact fetch on load. | Basic limits/order engineering is underway. Keep remaining cases explicit. Client completed-job coverage is a useful positive, not all historical-work protection. |
| P2: Complete interface selection and independent authorities (§§10.1–10.2, 10.7, 10.9) | Finish every GUI/CLI/public selector row, default GUI workspace resolution, blank/conflicting flags and unchanged fresh/offline behavior. Today's model/route/policy/preferences stay authoritative; past policy cannot overwrite them. Invalid current policy/preferences refuse without rewriting session files. Confirm multi-Agent isolation and credential/base-URL absence with configuration-only canaries. | Initial/public/client commands cover subsets. Browser engineering and later live plan should identify the remaining exact rows rather than infer them from one successful resume. |
| P2: Browser and connection lifetime (§10.9) | Real DOM identity/checkpoint/error labels; strict subscribed command/reused-ID behavior; applied watch revision before ack; overlap busy and disconnect/remount fencing. Imported historical card status/access, disabled forged supervision, omitted count, restored coordinates and fresh runtime generation; no resumed prompt/pause/provisional speech. Internal bookkeeping does not consume the 100-card window. | Actual Chrome work is in progress. Existing WebSocket fixtures and completed-job metadata do not establish visible Page behavior. Retain prior browser/control coverage. |
| P2: Lock and ownership structure (§§10.1–10.2, 10.5) | Noninherited lock descriptor with a surviving tool child; same lock inode after death; root reservation lock not held during slow validation; failed opens release all resources. Review actual SessionStore/worker parent/logger, one append descriptor, authority-owned snapshot capture and off-actor encoding. Check bounded single-pass prefix reduction instead of repeated whole-prefix work. | OS-lock/signal positives exist or are prepared; child-descriptor and structural/performance distinctions remain separate. Source review waits for the appropriate freeze and is not supplied by this declaration-only audit. |

Each missing runtime refusal needs a passing implementation parent, and each
material promise needs the applicable precise deletion/negative control. A
recomputed invalid semantic fixture must not fail first because of its envelope,
path or numeric token spelling. Do not rerun unrelated full gates merely to
fill a new receipt: retain source-bound applicable evidence, and run affected
groups when behavior or fixtures change.

Before acceptance, publish an aggregate required-command map covering these
rows, complete affected module/build/vet/race and delivered-tree checks, retain
earlier chapter coverage, and distinguish student initial results from repairs.
No aggregate gate, complete runtime acceptance, live matrix or post-freeze
historical comparison is yet established by the reviewed declarations. The
validation ledger still describes public groups as unrun even though the later
grader record reports initial student integration; the coordinator should
reconcile that summary without deleting either milestone.

Source SHA-256 at this audit:

- chapter-10.md: `41ef91e7fad0308f32bd874c8cb952fddc8900d41f3534c847183adb0e66a26a`.
- chapter-10-grader-review.md: `41dc4fc4eb210423328f1f6782fc4d490f2aa300ec54695090c5820a099dbb34`.
- chapter-10-validation.md: `51472ede13492ab970c1969b65a7f18e42bac76e3a5df5e1663d73ad801d279e`.

Disposition: completeness gaps identified for independent checker preparation;
no new runtime defect or student failure inferred. No test/code edit, build,
mutable runtime read, paid call or historical implementation comparison occurred.

## Independent Q5 watermark clarification closure

October 8, 2026. `/root/coder_ch08` reviewed the complete author delta and direct
response at `3a9e5b76053535ab22906959edb14609596f8a93`, the student's actual
Q4 acknowledgment/Q5 question, and affected §§10.3/10.4/10.6/10.8. Focused reads
rechecked Chapter 4's shared allocator/collision rules and Chapter 9's activation
allocation/retention. Full voice/procedure remained loaded without compaction.
This pass reads student teaching feedback, not mutable implementation code.

Q5 is resolved consistently. Activation and historical-job watermarks equal the
maxima in complete validated semantic state, including retired activations and
jobs outside the display window. The same rule applies to a snapshot-only origin:
an earlier identity can be represented by its semantic facts without its raw
event, but an unsupported extra cursor cannot be invented. Higher and lower
unsupported values refuse with session_corrupt. Full-origin reduction and
origin-seeded tail reduction now share that rule explicitly.

The request cursor retains its distinct lower-bound exception because admitted
requests may burn ordinals without durable turn facts. The shared live Job
allocator remains different from a session's historical maximum: resume raises
its floor without lowering an already higher cursor, and occupied artifact names
still consume skipped candidates without creating session Jobs. No activation
exception is needed because failed/no-op Skills transitions allocate nothing and
retired activation records remain represented.

The direct response acknowledges the original at-least/equality contradiction
and expressly declines the student's proposed snapshot-only relaxation, without
misclassifying the question as a student defect. Its Q4 acknowledgment agrees
with the read student feedback and claims no completed repair. No contradictory
instruction remains in this narrow scope; student implementation/confirmation
and exact-high/low controls are separate work.

Working author files match the freeze. The retained linter passes all hard rules;
this invocation reports 7,819 words (the author evidence reports 7,818), with soft
length/negation/person-gap warnings. The complete diff changes no existing JSON
fixture. No runtime execution, provider call, checker change or build occurred.

Reviewed SHA-256 values:

- chapter-10.md: `56f52e6ac8d251f6c775d226c4b6c48fec1311542d15450f3b666617584cbbaa`.
- chapter-10-evidence.md: `586cf6f7787b17cfc4188816dc72d94f6a8ebfe090217ce7f37badae5ee4bf2d`.
- chapter-10-student-feedback.md: `eb4e7faf3b711c5e628a4d2e189eb384cfed0923cb3362ab4b8bb9a902f0953e`.

Disposition: Q5 clarification proofreading accepted; no remaining scoped finding.
