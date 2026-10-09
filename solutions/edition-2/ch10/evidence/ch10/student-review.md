# Chapter 10 — initial student design review

2026-10-08. **Plan only; coordinator design check pending.** No runtime changes,
builds, tests, provider calls, credential reads, or delegated work occurred.
This is the initial new-only experience record, not Chapter 10 acceptance or
the later implementation/live-experience freeze. Only this file is submitted.

## Source identities and actual reads

- Released prerequisite: annotated `edition-2-ch09-r1` resolves to
  `54d7b1dbcb448ce77b52200b61127c091283a03c`.
- Canonical predecessor: commit `ac55f641e69220a612debbcd6f75e77fa9259b36`,
  `solutions/edition-2/main` tree `94315d7fa4cbd4d26ec00ca59d608ec3757e38cf`.
  Compared the working bytes of all 144 tracked non-evidence files with that
  commit: identical. This mechanical comparison includes tests and assets;
  it did not display or interpret their contents. No baseline tests were run.
- Teaching root: `/Users/bill/projects/ensemble-edition-2-revisions/ch10-student-inputs/new-only`.
  Read `manifest.json`; verified all 12 listed SHA-256 values and exact bytes
  against its `teaching_commit`, the released commit above. Manifest SHA-256:
  `82957ccd0d3588c20aba4a7cb3f5b49429c334bb0ea5e61bc8f2e2406e08aad9`.
- Read the entire repository `book/edition-2/skills/ensemble-coding/SKILL.md`
  before analysis. Its SHA-256 is
  `0180e3eb5d4f6032936729f31b473c70c421bce22001790a9709b05c57893bd1`;
  the pinned copy was also opened and verified byte-identical. Read the full
  pinned `architecture.md` and Chapter 10, including all tables and §§10.1–10.10.
  Chapter 10 SHA-256:
  `9eef50b45bdd40cdb47d9cee6e939e80b96b5b9aeefb51206543875013460b3c`.
  Its unreleased-predecessor wording is superseded by this handoff, not by an
  inference of Chapter 10 readiness.
- Pinned earlier teaching actually read: Chapter 1 lines 1–386 (including
  rules before examples and TL;DR); TL;DRs in Chapters 2–9 (2:134–205,
  3:110–171, 4:90–144, 5:52–120, 6:21–82, 7:31–98, 8:28–81,
  9:28–82); detailed contracts 2:206–470 and 566–606; 4:145–432;
  5:121–220 and 306–432; 7:209–458; 8:82–244; 9:83–747.
  Heading discovery covered Chapters 1–10. Earlier chapters were not read
  in full; the ranges above identify the inherited obligations used here.
- Predecessor source contents read, relative to `solutions/edition-2/main/`:
  - Entire `internal/common/{types,actor,skills,watch}.go`;
    `internal/eventlog/log.go`; `internal/jobs/limits.go`;
    `internal/skills/replay.go`; `internal/tools/{registry,supervision}.go`;
    `internal/policy/policy.go`; root `watch.go`, `policy.go`;
    `cli/{main,session}.go`; `gui/cmd/ensemble-gui/main.go`;
    `gui/projection.go`.
  - `ensemble.go`: 1–300, 332–631, 689–808;
    `internal/llm/actor.go`: 1–240, 300–409, 454–562, 597–728;
    `internal/llm/events.go`: 110–552; `internal/llm/engine.go`: 1–70;
    `internal/skills/skills.go`: 1–55, 198–216, 402–446;
    `internal/jobs/jobs.go`: 1–105, 137–230;
    `gui/connector.go`: 207–370; `gui/web/gui/connector.js`: 1–78.
  - Declaration searches additionally covered the listed core source and
    `internal/llm/{skills,watch}.go`; targeted client searches covered
    `gui/connector.go`, `gui/projection.go`, `gui/web/gui/connector.js`,
    `cli/chat.go`, `cmd/main.go`. Source-path discovery listed Go/JS/modules,
    including test filenames and a few predecessor evidence filenames.
- Read the authorized coordinator inbox at opening, after source inspection,
  and before submission. It contained only the initial plan-only restriction.
  No other worker conversation was opened.

Exposure accounting: two unscoped `git status --short` invocations displayed unrelated
working-tree **path names**, including excluded directories and later manuscript
names. Their contents/diffs were not opened. Subsequent status checks were scoped
to the student tree. Large combined tool output was truncated; Chapter 10's
boundary passage and Chapter 1's opening were reread separately. The entire
mandatory skill was available from the initial separate read. No first-edition
source, frozen solution, grader/fixture implementation, author evidence,
memory, or unrelated history was inspected. Historical quotations and links
inside permitted teaching were seen; no excluded linked source was followed.

## Ownership and state

Keep the root as composition only, shared declarations/interfaces in `common`,
and implementation spokes importing common rather than one another. Extend the
accepted owners; no parallel mutable save Context and no second append writer.

| Owner / implementation | Parent interface and authoritative facts | Captured value / lifetime |
|---|---|---|
| Ensemble / public root | Application root; logger, runtime Agent IDs, global job allocator, brief store/SessionID reservations | No saved runtime root. Raise job floor monotonically; never restore the root counter as a session fact. |
| Agent / public root | `common.Ensemble`; copied creation configuration, canonical DataDir, SessionID/identity, history and its Context, owned services | Session metadata plus reduced conversation value; never serialize Config wholesale. |
| Actor / llm | `common.ActorAgent`; admission order, request cursor, durable commit order, watch revision, runtime queues/operations | Captured request cursor and durable identity facts; queues, handles, pauses, operations and watch generation are new runtime state. |
| Engine / llm | `common.Agent` (existing Agent adapter); transport and per-producing-model usage | Owned usage accounts and accepted-response provenance records; replay is not new spending. |
| Registry / tools | `common.ToolAgent`; installed definitions and wire argument parsing; current grants reached through Agent | Complete installed ceiling definitions for identity, distinct from visible declarations. No handler code serialized. |
| Skills / skills | `common.SkillAgent`; frozen definitions, committed roots/closure/material, revision and activation allocator | Complete active/retired material and ordered transition provenance. Exact bodies stored once. |
| Jobs / jobs | `common.JobAgent`; live jobs, report cursors, pending typed limits | Pending limits restored; historical accepted job facts retained without a live Job. No process or cursor restoration. |
| EventLog / eventlog | Agent creator through common interface; sole events.log descriptor and checked append I/O | Original bytes remain intact. Add bounded streaming reading and validated append-open; never rewrite the prefix. |
| SessionStore / persistence (new spoke) | `common.SessionAgent`; resolved file identities, advisory-lock descriptor, checkpoint worker and last committed anchor | Owns detached capture bytes/candidate while writing, never an independently evolving conversation. |
| GUI Server and browser owners / optional module | Existing public application/Server/Page chains | Preferences remain Server-owned. Browser connection, speech and pause lifetimes remain transient. |

Every new worker receives its actual creator interface: checkpoint worker →
SessionStore → Agent → Ensemble; no injected Engine/Skills/Jobs/logger closures.
Extend common owner interfaces with narrow operations and use private structs
where useful. Existing root adapters remain views of their actual Agent.

The historical job projection already in Context is an accepted-event view,
not Jobs' racing worker snapshot. Capture that projection through its owner;
Jobs validates its schema and owns pending limits and live lookup. Likewise
skill material in conversation is a projection of Skills authority. Do not
promote either copy into an independent mutation owner. Retain compact semantic
facts currently discarded by reduction (turn outcomes, response accounting
provenance, consumption/redaction facts, skill transition references) under
their existing responsible owners as events are accepted, not by maintaining
another complete save-only Context.

Creation-only: DataDir, resolved LogPath, SessionID/identity, workspace,
policy destination, skill mode/source selection/primary/copied bindings and
installed ceiling. On a mounted session base System and handler definitions
are fixed; unchanged `Config()` → `SetConfig()` remains valid. Model/vendor,
route, credentials, resolved model and delivery remain current application
inputs with inherited idle-only general updates. Active turns retain captured
configuration/policy. Current policy/revision and GUI preferences/revision keep
their own files and update protocols. Pending tool limits are mutable,
actor-committed Jobs settings. No credentials, endpoints or live transport
configuration enter a session snapshot.

## Public surface proposed for review

Public names below are choices, not new fixed outer wire spellings. Values and
interfaces are declared in common and publicly aliased as in the predecessor.

| Operation | Proposed contract |
|---|---|
| `Ensemble.NewAgent(Config)` / `Load(LOG, Config)` | Preserve fresh exclusive log construction and offline reading. Fresh constructor refuses nonempty DataDir with `session_conflict`. No silent standalone-log migration. |
| `Ensemble.OpenSession(SessionOptions)` | DataDir explicitly required, no explicit LogPath. Options carry Config and System presence separately (nil means omitted; pointer to empty means explicit empty). Return only a fully validated live Agent. |
| `Ensemble.InspectSession(DIR)` | Read under shared nonblocking store lock; owned historical inspection with SessionID, origin/full-history boundary, sequence, checkpoint anchor, settled/resumable status and diagnostics. No live registration, credentials, callbacks, processes or catalog required. |
| `Ensemble.InspectCheckpoint(bytes)` | Strict standalone checkpoint inspection with no store lock. Preserve owned bytes and distinguish snapshot validation from current live compatibility. |
| `Agent.Session()` | Safe nullable session state plus separate durable boundary inspection; use the actor boundary on live Agents. |
| `Agent.ExportCheckpoint()` | Settled actor capture, then bounded off-actor encoding/hash; return owned exact versioned checkpoint bytes and anchor. Does not claim a checkpoint file commit or advance checkpoint_seq. |
| `Agent.Checkpoint()` / root forwarding by Agent ID | Capture and commit one file; return typed `{as_of, watch_revision}` only after commit/application. Busy returns promptly. Standalone refuses `session_conflict`. |
| `Ensemble.ImportSession(bytes, SessionOptions)` | Explicit non-null checkpoint import into an unoccupied destination, preserving SessionID. Validate compatibility before exclusive origin/log creation; first mount has resumed=false. Same identity reservation rules apply as open. |
| `Agent.Close()` / `Ensemble.Close()` | Join owned work, truthful terminal facts, final checkpoint, log close, lock/reservation release; preserve actual failure and repeated-close result. |

Treat exporter work conservatively as using the same single store worker gate;
second export/save refuses busy, with no unbounded queue of copied state.
Captured buffers are copied at return boundaries. A canceled/disconnected
waiter cannot roll back a committed write. Inspection needs an explicit
historical Render operation accepting current render inputs; it must enforce
recorded base identity and same-target opaque rules without installing tools.

CLI/GUI share a presence-aware selector reader **before** filling defaults.
Human and GUI default to `.ensemble/session` relative to captured launch
workspace; `--session-dir` opts protocol into session mode. Explicit nonempty
CH02_LOG preserves standalone fresh construction. Conflicting or explicitly
blank selectors refuse before opening state. Keep offline dump/render selection
and default protocol bytes. Add `session inspect DIR`, `/session`, `/checkpoint`.
GUI receives exact `state.session:{id,resumed,checkpoint_seq}` or null and sorted
`state.job_access:[{handle,live}]`. Bookkeeping events never enter the 100-event
renderable window. Add subscribed `checkpoint` command; publish
`session_changed` before `command_ack` with status `saved`, as_of and applied
watch revision; refuse with `command_error`. Keep earlier commands' shapes.
The GUI transport waits for delivery of that revision before its acknowledgement.
Show historical status separately from live access; forged supervision still
fails at Jobs. Extend lossless protocol-counter handling to every newly exposed
sequence/handle/anchor, without converting arbitrary tool data to counters.

## Semantic codec sketch (state_version 1)

During implementation publish the complete nested field/type grammar in
`persistence-format.md` before independent checks. This sketch fixes the planned
contents, not an implementation or permission to omit a Chapter 10 invariant.
Every structural object has a closed field set, required arrays are present even
when empty, nullable values are explicit, and typed part variants have separate
required/forbidden fields. Snapshot structs are not default marshals of Context.

Keep the **exact outer** required fields: version=1, session_id, identity,
positive as_of, high_watermarks={event,request,activation,job}, state_version=1,
state and state_sha256. Identity has exactly mode/system/skills/handlers;
handler identity uses `schema` rather than the predecessor ToolDefinition's
`input_schema` tag. SessionID is 16 random bytes rendered as 32 lowercase hex.
Hashes are 64 lowercase hex. State/hash null together is only the documented
rebuild control; snapshot-only import requires non-null state.

| Required state field | Value and consistency checks |
|---|---|
| `session` | `{id,identity,as_of,high_watermarks}` agreeing with envelope. No relabeling another state's identity. |
| `conversation` | Ordered accepted entries/instructions, typed part positions, pending ephemera and hints with explicit consumed/pending records and completed-batch anchors; effective redaction spans; explicit pending/active/continuation/turn/batch fields. Settled live candidates have no unresolved/deferred batch, human input or response slot. |
| `calls` | Indexed seen call facts: original response sequence/part index, exact issued ID, name/args/provenance, dispatch/result coordinates, pairing flags and job reference. Cross-check parts and results; discarded text never frees an identity. |
| `turns` | Seen request IDs with positive request_index, start/end sequence and terminal outcome plus historical policy capture. Strict ordinal order, valid completion and no reuse. No reliable runtime completion handles. |
| `responses` | Accepted response coordinates, producing/requested provenance, model_reported, raw usage and normalized disjoint counters; refer to conversation parts instead of copying response bodies. Validate account totals against these records. |
| `usage` | Sorted complete per-provenance accounts exported/restored by Engine, compared with accepted response facts exactly once. |
| `skills` | Explicit null for plain mode; otherwise primary activation reference, ceiling names, complete committed safe state, contributors, all immutable activation records with creation/retirement coordinates, last activation ID, and ordered `{seq,action,name,revision,activated_ids,retired_ids}` transitions. Reconstruct each prior closure/offer set and exact candidate, not just today's roots. |
| `jobs` | Historical accepted job snapshots and call/result references, sorted by handle; pending typed one-shot overrides or null; set/consume provenance tied to accepted calls. No runtime status read, process, output cursor or fetch of a locator. |
| `window` | Last <=100 renderable durable facts, total renderable count and true sequence range; counts also retain total historical event/entry/part/identity cardinalities needed for continued storage bounds. This is a bounded presentation witness, not the saved conversation. |

Avoid duplicating primary/manual bodies in conversation and window: skill
entries reference activation IDs plus sequence/anchor; saved skill window facts
reference material IDs and retain their transition state. Decoder reconstructs
their ordinary public display shapes from immutable material, verifying those
references. Opaque/raw JSON stays byte-preserving (dedicated raw-JSON string
fields in the semantic codec, parsed and validated again); arbitrary argument
data remains arbitrary JSON under its inherited schema. Canonical equality
checks use the parsed semantic value; raw replay bytes are never replaced by
their canonical hash encoding. Document that distinction explicitly.

Use one strict bounded JSON/canonicalization implementation in persistence,
exposed by common interface through Agent for other owners that need equality.
An Agent-created stateless codec service supplies that interface even for a
standalone Agent with no SessionStore; its parent is Agent, not a sibling store.
Reject invalid UTF-8, duplicates at every object level, trailing data, excessive
nesting, wrong/null/missing/unknown structural fields. Preserve old event
metadata compatibility separately. Keys sort by UTF-8 bytes, arrays retain order,
strings use the chapter's exact escaping without HTML escaping or normalization.
Normalize coefficient/exponent decimal tokens exactly, including arbitrary
schema numbers; no binary64 round trip and no expansion of huge exponents.
Integer identity validation and allocation remain full uint64 with whole-group
overflow checks. Existing `reflect.DeepEqual` on float-decoded JSON in call and
schema matching cannot implement the new equality contract.

Enforce 1 GiB log, 512 MiB checkpoint/origin, 256 MiB canonical state,
1,000,000 collection counts, depth 128 and 1,024 handlers/16 MiB canonical
definitions while reading/encoding, before size-driven allocation. Preserve
earlier source/input/policy limits. Event-size precedence needs Q1 below.
Existing-state bound violations are session_corrupt; newly admitted storage
violations session_limit, subject to the unresolved inherited skill refusal rule.

## Cross-owner validation and replay

SessionStore decodes/hashes storage, then calls its Agent parent to validate an
owned candidate. Root orchestration calls common service interfaces:
conversation validation/reduction in llm; usage restore/validation on Engine;
catalog identity, historical transition validation and live material comparison
on Skills; ceiling-definition checks and typed wire decoding on Registry;
typed pending-limit validation/application on Jobs. Store never imports those
spokes or receives them as constructor siblings. Pure validation helpers retain
their responsible owner for logging. Installation is private construction work,
not a public replace-state operation.

For full origin, stream the complete log once with indexed call/request/job/
activation validation; independently reduce its prefix and compare semantic
values at S, then install the validated snapshot and apply >S exactly once.
Use a disposable, unexposed validation owner graph; it is not a second live
Context or writer. Null-state rebuild uses the same reducers from sequence 1.
Avoid repeated full-prefix scans/copies per event in the new loading path.
Request cursor may exceed recorded ordinal maxima; compare its lower bound,
not invisible canceled admissions. Event/activation/job watermarks equal
durable maxima; job maximum includes history outside the display window.

Snapshot-only import validates reduced state without an omitted log prefix.
Write accepted bytes unchanged as exclusive origin.json; its raw-file hash
includes any LF. Write only header and matching session_anchor at S+1,
ordinary continuation at S+2. Never manufacture prefix events. Every later
open validates immutable origin, anchor/hash/watermarks and contiguous tail,
even with a newer checkpoint. Full-origin stores reject origin.json;
log-only loading of an anchor returns session_origin_required.

Live compatibility reuses frozen parsed definitions, copied bindings and the
installed ceiling, independent of source paths. Compare catalog and bindings
canonical hashes; also validate every historical activation/transition and its
once-rendered bytes against the corresponding original candidate. Do not
replace recorded text with rerendered text on mismatch. Offline mode checks
recorded facts without catalog or environment. Keep imported Jobs empty and
raise the application's real job allocator floor before admitting new work.

For session tool limits, replace mutation-inside-Resolve with prepare/commit/
apply through Actor. On the next attempted accepted call: capture pending
typed overrides, append tool_limits_consumed **before** validation/permission
and tool_called; then clear exactly that candidate and resolve effective limits.
After a valid tool_limits tool_called, append tool_limits_set before successful
tool_returned and apply its typed candidate. Preserve explicit empty pattern;
store the serializable pattern/number/presence values, never compiled regex.
Interruption uses the same attempted-call path. Replay requires these facts in
new sessions, validates accepted-call order and copied overrides, and performs
no dispatch. Standalone legacy logs retain their old offline meaning.

## Locks, capture and lifecycle

1. Resolve DataDir once against declared workspace, canonicalizing the existing
   directory prefix. Reject symlink/nonregular leaves. Reserve canonical store
   identity under Ensemble's brief mutex, acquire nonblocking exclusive advisory
   owner.lock on macOS/Linux (unsupported elsewhere), then read bounded files.
   Keep the descriptor close-on-exec and never unlink/replace owner.lock.
2. Once identity is validated, reserve SessionID under the same short root lock.
   A copied store collides by SessionID inside this Ensemble. Validation and
   filesystem I/O hold no root lock. Reserve settings/store file identities
   consistently so policy/preferences cannot also own a session leaf. Current
   policy and GUI preferences must validate before session mutation; GUI startup
   needs a read-only preferences preflight through its own owner.
3. Construct privately, validate full candidate/compatibility and settled state,
   install each owner's values, open the one log for append, and only then expose
   fresh runtime Agent/Actor/watch identities and tool admission. Fresh creation
   commits session_initialized at 1, followed by required skills_initialized,
   before publication. No startup HTTP, tool, artifact read or repaired outcome.
4. Refusal unwinds private workers/descriptors, policy claims and both root
   reservations without rewriting source files. A failed fresh append/import
   retains diagnostic artifacts; it is not silently retried as an empty store.
   Distinguish missing files from failed reads and use the exact §10.2 table.
5. Actor accepts capture only when idle and settled, including no pending
   admitted prompt in its mailbox, active response/report/input/model worker,
   unanswered call or deferred batch material. A running independent Job with
   paired initial report is allowed. Completed-batch unsent hints are allowed.
   Check admission/cursor under the mailbox mutex at the capture cut; subsequent
   admissions belong to later work. Snapshot copy costs O(retained state).
6. Copy owned state at that cut; release Agent/Jobs/Engine/mailbox locks before
   encoding/hash/write/sync/close/replace. Store's brief gate permits one worker.
   Agent's existing append serialization remains the only log path; checkpoint
   disk work holds none of its locks. Worker completion posts to Actor without
   waiting while holding the store gate. Later job facts remain append-log tail.
   Successful replacement is commit; update last anchor and watch revision,
   publish session_changed, then acknowledge that captured as_of, not current seq.
7. Close stops admission, settles requests, joins cancellable model/report/input
   workers, kills/drains managed jobs and records their truthful terminal facts.
   Actor continues draining facts while shutdown workers run. Join/apply any
   checkpoint worker result, then capture/write final settled state off actor.
   Only afterward close log/transport and release advisory lock and root claims.
   Preserve nonkillable-handler suppression limits. A faulted append path skips
   final replacement and returns failure while still cleaning up. A save error
   remains visible on repeated close and makes the standalone process fail.
8. Wire EOF, /quit, SIGINT and SIGTERM to that close path, propagating errors
   before any os.Exit. GUI terminal EOF only detaches; Page close only releases
   its registration/speech. Forced death promises descriptor release, not a
   settled checkpoint or automatic crash recovery.

## Questions for coordinator before affected implementation

**Q1 — inherited record-size/refusal precedence (blocking).** Chapter 9 §9.5
says ordinary/header limits remain unchanged and oversized skill transitions
return controlled skill_too_large without faulting the Agent. Chapter 10 §10.8
says one encoded event record is 64 MiB, exact boundaries are accepted, and new
record/log/count overflow follows terminal persistence rules with session_limit.
The released eventlog reader has an approximately 16 MiB ordinary-record limit
and a separate 64 MiB skill limit. Does session mode replace the ordinary limit
with 64 MiB? Does the skill candidate's earlier controlled refusal still win
before session storage admission? Proposed resolution: 64 MiB session physical
records, legacy standalone limits unchanged, and retain skill_too_large at its
candidate preflight; separately terminal session_limit for log/count admission.
Please settle both precedence rules in teaching rather than tests.

**Q2 — public history after snapshot-only import (contract gap).** §10.4 forbids
inventing pre-origin events, but inherited Events(), Dump(), /history and
ReconstructRequest(seq) assume a complete event slice. Proposed behavior:
Events/Dump expose only actual imported-store anchor/tail; inspection explicitly
reports origin_as_of and incomplete raw history; /history labels that boundary;
pre-origin request reconstruction refuses a safe unavailable-history error.
For captured sends in the available tail, seed reconstruction from origin state.
The watch still includes its genuine saved recent window without adding those
records to events.log. Is this the intended public distinction, and should the
unavailable-prefix error have a specified stable code?

**Q3 — raw payload exactness versus canonical equality (design check).**
§10.3 preserves original opaque/payload bytes but requires semantic equality and
canonical number hashing. I propose raw-JSON string fields in the semantic
codec for byte-preserving replay data, validated as JSON; canonical parsed-value
comparison only where required (arguments, schemas, identity). Is retaining
those raw lexemes in snapshot state required, or is decoded semantic equality
sufficient for opaque JSON while only original log/manual bytes stay exact?
This affects the published nested codec and hash fixtures, not the fixed outer
envelope. Until answered, the conservative byte-preserving design above applies.

These questions are for the coordinator's new-only design check, not routine
permission requests to Bill. No affected code has begun. The other owner/API/
lifecycle analysis is complete independently of the answers.

## Initial teaching feedback and later gates

The selector table, origin/anchor rules, exact decimal algorithm and accepted
durable versus racing worker distinction make several dangerous choices explicit.
The central implementation difficulty is exporting enough semantic evidence to
validate omitted history without relabeling an event archive as state. The
proposed indexed facts and activation references address that; coordinator
review should specifically check sufficiency and ownership before code.

Observed predecessor extension points are expected Chapter 10 work, not failures:
no SessionStore yet; Skills initializer currently requires seq 1; pending limits
mutate without durable facts; request/job counters increment without exhaustion
checks; history/reconstruction assumes a full log; browser lossless conversion
currently targets settings/skill counters; general config drops System presence.
Targeted changes must preserve accepted earlier behavior. No failed experiment,
test result, successful restart or author response is claimed in this phase.

After design release: publish the complete codec/public surface, implement only
in main, run relevant module formatting/build/vet/tests/race checks and required
black-box commands (initial Chapter 10 command:
`python3 scripts/edition2/accept_ch10.py CLI_BINARY --receipt RECEIPT.json`).
Coordinate remaining coverage without reading checkers. Prove nonempty tails,
null rebuild, snapshot-only import followed by later save/reopen, independent
three-surface next-request equality, ownership/copy isolation, identity extremes,
one-shot limits, all selectors, writer exclusion, truthful fault boundaries and
signal/close behavior with distinguishing local controls.

Before paid runs, submit a bounded reviewed feature matrix and wait for the
separate API-access release. Then retain fresh human CLI PTYs, actual browser
and public multi-Agent runs on all three providers, including export/import and
current policy authority. Freeze the initial source/experience/evidence before
independent historical comparison, implement reviewed improvements, rerun affected
checks, and verify recorded author feedback. No fake-only or initial-subset pass
can complete this chapter. This plan commit creates no acceptance tag or snapshot.

## Phase 2 acknowledgment — 2026-10-08

Continued from plan 7cb84290af3e8bac5359339f8ab5d4e7e941bf74. Reread the entire
repository coding skill and pinned architecture. Read full revised Chapter 10
and full direct author response in clarification-af5a762; verified both manifest
hashes and exact bytes at af5a7625282f1c1fffe4a18ba508e69252b647c1. Chapter hash:
46249bc8fadd59f55c5dc12aebfe2c3ca22ce9c031d4a7972aa97e7140b22d60; feedback hash:
63bea996ae2e31aa96e9f609287d1d9c6f6c834289fef263ab4013ca73beb46c.
Read the sole coordinator inbox's phase-2 release; no excluded source opened.

Q1 resolved: 64 MiB all session events, inherited headers/standalone bounds;
controlled skill_too_large candidate preflight precedes terminal session storage
admission. Q2 resolved: actual anchor/tail raw history, labeled origin boundary,
stable history_unavailable before origin, origin-seeded available-tail replay.
Q3 resolved: exact raw JSON bytes/lexemes, validated dedicated raw strings in the
codec; canonical parsed equality only for its specified semantic purpose.
These author responses close the original questions; no Bill permission needed.

Accepted safeguards: validation owners remain inert children of the real Ensemble,
unregistered, with no Actor/process/policy write/shared allocator change before
acceptance. Restoring accepted usage is not charging replay as a new response.
GUI delivery waits belong to connection lifetime and release on disconnect/loss.
The early grammar and signatures are persistence-format.md and session-api.md;
status/evidence are in implementation-status.md. No live access is released.

## Implementation clarification acknowledgment — Unicode and semantic witnesses

Read the full Chapter 10 and direct feedback in clarification-cf73a64, verified
manifest SHA-256 and exact Git bytes at cf73a646ba665d2070bb51a77bc0b1c5c4912a2b.
Chapter SHA: 3faa154f60e875497d96594a1afdc219f398856b0a9bd46b39cf72f36ca126df;
feedback SHA: cfb93f93cf7b53440a92c7304c8cb690a65e69f887d6c758cdfa1370d5b7f264.
Strict session/canonical JSON keys and strings reject lone escaped surrogates
before decoder replacement, accept proper pairs, genuine U+FFFD and literal
escaped-backslash text; raw-wrapper inner JSON obeys this too. Inherited standalone
scope stays unchanged. This resolves the new clarification, not a runtime pass.
Reread the entire coding skill after compaction; no excluded source accessed.

Coordinator response: add CalledAt/ReturnedAt to CallState with response/dispatch/
result ordering checks. Missing dispatch evidence was my codec-design omission,
not missing teaching. The published grammar now names those required coordinates.
Change checkpointWorker's concrete Store back-pointer to common.SessionStore,
retaining its actual creator and logger chain. This corrects my implementation
of the already accepted owner plan. Both corrections are unblocked.

Actual implementation reads include main's session/ensemble/watch/CLI/GUI source,
common, eventlog, persistence, llm, skills, tools and jobs source and local tests.
Only the authorized coordinator inbox was checked for external direction. No
checker implementation, credentials, historical answer or other worker read.

## Raw record framing question — concrete local failure

The initial CLI cascade is reproducible without checker source: change only the
local fake response's usage object from compact JSON to `{ "input_tokens": 1,
"output_tokens": 1 }`. TestCLISessionRestartLocal then refuses its own second
mount at record 7, `checkpoint differs from reduced prefix`. Our byte-preserving
Clone retains provider RawUsage whitespace; inherited eventlog json.Marshal
compacts RawMessage when writing. Independent log reduction therefore has other
raw bytes. The earlier compact local response did not distinguish this case.
This is separate from the now-fixed schema canonical comparison.

Q4 for coordinator/author before the affected fix: does "exact recorded raw
bytes" mean bytes after the one durable event serialization (preserve number
lexemes, align accepted in-memory state with that recorded event), or must new
session log encoding preserve original provider raw JSON whitespace too? The
latter includes embedded insignificant LF/CR in raw JSON, which cannot be emitted
verbatim inside a one-physical-line JSONL event. It would require an explicit
validated byte-preserving event extension/string wrapper, not merely a semantic
snapshot wrapper. The former can retain existing event shapes and JSONL framing
by reducing the exact once-serialized accepted event, while all subsequent
capture/import/rebuild comparisons remain byte-exact. No canonical number
normalization would occur at admission. Request a published interpretation rather
than inventing an event extension. I have not implemented either affected choice.
Continue unblocked validation/limits/lifecycle/client work in parallel with this
contract question. Current initial binary and failing receipt remain retained.

Correction to milestone 3 diagnosis: coordinator pointed out the original CLI
receipt already had runs/exit/stderr; I initially summarized only checks. Read
my own runs diagnostics now. Public path mismatch is independently confirmed as
a macOS alias fixture defect; correct canonical paths remain unchanged. Targeted
corrected-selector checker is running. No excluded checker source opened.

## Additional semantic witness and connection API refinement

Add required Context.RequestSeqs: ordered accepted request_sent coordinates.
Without this compact set, an old Guidance.consumed_at could name any absent
sequence after its receipt; the bounded window cannot establish request existence.
This is another student codec sufficiency correction, not a teaching gap. It
retains no request body/configuration and does not turn state into an event archive.
Limits now validate set/consume/dispatch/result ordering using their compact facts.
Published watermark grammar corrected to the chapter's "at least" bound; Job
allocator floors survive snapshot installation without inventing old jobs.

CheckpointContext variants supplement the reviewed public API. Context cancellation
ends only the caller's wait; Actor's accepted operation and store worker retain
ownership. GUI uses its connection context for both save waiting and watch barrier.
This implements the coordinator's disconnect safeguard without an unjoined client
wait. Parent interfaces retain actual Store -> I/O -> file creator chains for
local checked-write fault controls. All nine nested example modules passed vet
and tests in nested-module-checks.json (before this API addition).

## Q4 acknowledgment — prepared accepted event boundary

Read complete revised Chapter 10 and complete direct feedback in
clarification-c5ad6c0; verified manifest and exact Git bytes at
c5ad6c04a67c0e2eeafde77f52d90dfdd7695d8e. Chapter SHA-256:
41ef91e7fad0308f32bd874c8cb952fddc8900d41f3534c847183adb0e66a26a;
feedback SHA-256: 0b4ccab128da68c300c9a831693e2037ebc9e5291949e25bde8bc2318a9ca8d2.
Q4 is resolved before the affected implementation: candidate validation precedes
one bounded JSONL preparation; prepared bytes define accepted raw fragments;
append those exact bytes before applying/publishing. Preserve original number
lexemes/order and decoded text/signatures, with no hash canonicalization or float
round trip. Existing imported records are already accepted; no read-side rewrite.
Skills preflight stays controlled; transient streaming remains inherited.
My local whitespace-only reproduction converged with independent preserved-binary
diagnosis, rather than supplying the only discovery. Keep both initial failures.

Q5 wording check, independent of Q4: §10.3 says every non-event watermark is "at
least" represented allocated identities, whereas §10.4 says event/activation/job
watermarks "must equal" reduced durable maxima. Proposed distinction: snapshot-only
import validates allocator lower bounds; an available prefix proves its actual
maxima (seeded from validated origin for imports), so an ordinary later checkpoint
cannot arbitrarily raise that seed without a durable allocation. Confirm whether
snapshot-only activation/job watermarks above all represented identities are
valid, or equality is required even there. Request cursor explicitly allows burned
ordinals. Pause only the ambiguous above-maximum activation/job import case;
continue normal exact-max paths and all other work.

### Q5 published read and acknowledgment

Read the complete chapter-10.md and complete chapter-10-student-feedback.md in clarification-3a9e5b7 after verifying manifest.json and exact permitted Git blobs at 3a9e5b76053535ab22906959edb14609596f8a93. SHA256 respectively 56f52e6ac8d251f6c775d226c4b6c48fec1311542d15450f3b666617584cbbaa and eb4e7faf3b711c5e628a4d2e189eb384cfed0923cb3362ab4b8bb9a902f0953e. These resolve Q5: complete semantic facts establish exact activation/job maxima, including retired/out-of-window identities in snapshot-only origins. Unsupported higher values refuse too. Only request admissions have a burned-ordinal exception. Ensemble's live job floor remains separately monotonic. I withdraw the proposed snapshot-only arbitrary job-floor relaxation and will remove it; no missing raw prefix is required.

After compaction read the entire mandatory skill again (a batched display truncated its tail, so repeated a standalone complete read). Read only the authorized coordinator inbox, published pins, main source and own receipts. No excluded implementation, credential or provider reads. Coordinator and student physical-bound checks overlapped before the inbox's avoid-duplication note was seen; both were already complete, and no repeat is planned for unchanged read-side code.

### Revised local checkpoint and initial experience (before historical comparison)

Q1–Q5 and escaped-Unicode clarification are resolved by published new teaching.
No unresolved architecture/teaching question remains at this local checkpoint.
The actual difficulty was preserving the distinction between provider text before
acceptance, exact accepted event bytes, canonical equality, and reduced state.
The published preparation boundary resolved it without weakening exact replay.
Q5's exact complete-state maxima now replace my withdrawn arbitrary-floor idea.

The first extended public run found my actual Agent/Jobs lock cycle, not another
teaching gap. Accepted one-shot limits now have a Jobs-owned mutex independent of
report workers; Watch copies Agent state before consulting live ownership. A
controlled report/Agent interleaving regression and actual job completion during
a gated checkpoint pass under race checks. The first timeout remains preserved.
The unchanged inherited diagnostic returned0/100 at its old GUI/startup interface;
coordinator must assess that harness coverage independently. No checker code was
read and its failure was not silently waived.

Further self-review corrections: indexed collection/unresolved-call admission
and limit-fact lookups avoid newly introduced repeated full-history scans; the
index is derived, omitted from the strict codec and rebuilt once on import.
Bounded state encoding no longer marshals/parses the same large semantic tree
repeatedly, and prechecks escaped strings. New session parts obey their published
closed variants. Meaningful empty tool results retain an explicit parts array at
preparation. Skills candidate validation precedes session collection admission.
Construction installs Actor before publishing the accepted Agent in Ensemble.

Initial local test-authoring failures in this correction group: used a nonexistent
Job.Write in the deterministic interleaving test, then corrected to the actual
Jobs write owner; a validation-helper receiver change missed two event parser
call sites at compile time and was corrected. No assertions were removed. All
modules and relevant race/format checks subsequently passed.

Actual additional reads since Q5: permitted main source files in ensemble.go,
session.go, watch.go, persistence-format.md, session_test.go,
session_lifecycle_test.go; internal/common/types.go; internal/jobs/{jobs.go,
jobs_test.go,limits.go,limit_facts.go,report.go,process.go}; internal/llm/{events.go,
engine.go,semantic.go,session.go,session_limits.go,watch.go,actor.go};
internal/persistence/{codec.go,json.go,store.go,event_fields.go};
internal/eventlog/{prepared.go,log_test.go}; targeted permitted Chapter2 reference
lines and own retained receipts/status. Some reads were targeted ranges/searches,
not claims to reread those whole files. Git/source hashes and go list imports are
metadata inspections. Full Q5 chapter/direct response and entire skill reads are
recorded above. No excluded source, credentials, real provider or other worker
conversation was opened. The allowed coordinator inbox supplied diagnoses and
published clarification pins; reviewer implementations were not consulted.

The live matrix is a proposal only. There has been no actual Chapter10 real-model
spin, human-participation claim, historical comparison, author feedback confirmation
of runtime results or chapter acceptance. Independent remaining fault/allocation/
structural coverage needs the immutable revised source checkpoint. Preserve this
initial student experience and original failed receipts for that later review.

### Phase 2 repair: acknowledged terminal append/close failure

Reread the entire mandatory coding skill, complete permitted pinned architecture,
and complete clarification-3a9e5b7 Chapter10/direct feedback. Verified both teaching
files against manifest SHA256 and exact3a9e5b76053535ab22906959edb14609596f8a93 blobs;
identities remain as recorded above. A batched display clipped the join between
feedback/chapter; reread the affected tail/opening so no missing text is assumed.
Read authorized inbox additions and the complete safe result fields in
results/ch10-fault-first.json and results/ch10-append-fault-diagnostic.json (not
checker implementations). The frozen122b04a one-byte partial public Append fails,
but Close returns nil and performs create=1/replace=1. Equal resulting bytes do
not satisfy Chapter10.5's explicit no-replacement requirement. This is my runtime
fault-propagation defect, not missing teaching or a reason to revise expectations.

Repair plan within existing owners: Agent retains the first terminal append
cause at actual storage failure/terminal storage-bound admission. A common Actor
parent accessor distinguishes it from ordinary rejected input. Actor enters its
existing persistence cleanup path on that terminal public append, prevents new
checkpoint capture and checks owner fault before final save. Retain Close's result
for repeat callers while releasing log/store locks and root reservations. Tests
must observe unchanged checkpoint inode or no created checkpoint, not just bytes;
include invalid-input positive recovery, terminal size admission, repeated close,
and released ownership. No second writer or mutable Context is introduced.

The supervised process stopped101 during the disk-blocked validation; uncommitted
source and originals remained. The attempted status write was aborted, so this
records the failure now: Go could not create vet.cfg/build directories (`no space
left on device`); the redirected after-fix receipt was empty, not a passing test.
No student cache deletion occurred. After coordinator maintenance, reloaded entire
mandatory skill and only appended coordinator-inbox messages; resumed exclusive
build stage without redoing completed source edits. First local test adapter's
Config-under-append-lock timeout was a fixture mistake, corrected by capturing
the log path before injection. Baseline then independently reproduced runtime
faulty checkpoint replacement and nil repeated close. No new teaching ambiguity.


Repair read/exposure ledger continued: reloaded the full mandatory skill after compaction; read authorized inbox through “Repair validation progress received.” Inspected own four-file runtime diff and new session_append_fault_test.go, own append-repair command receipts and current review/status/live plan; searched own skills.go, GUI preferences service and GUI web preference controls to bind the proposed public APIs/font_size16→18 action to actual implementation. Earlier repair reads also covered own CLI/GUI launch/configuration, session/persistence/eventlog tests and safe supplied diagnostic fields. No reviewer/checker source, old/future teaching, other worker conversation, credential or provider endpoint was accessed. Module and race checks now pass; prior baseline/fixture-timeout/disk failure remain retained. No teaching question is open; live support completeness and separate release remain future gates.


Repair outcome: immutable8882a18cf98e9a4b70afccfbe980f6344630aae6 passes the targeted independent append check and all four frozen fault groups, including checkpoint I/O faults, canceled wait/tail responsiveness and close/join/lock order. Read the newly generated result fields and hash metadata only; no overlay/checker implementation was inspected. Source/checker stayed unchanged. Local validation includes all nested modules and main/GUI race suites. Actor queries the real Agent-owned terminal failure through its common parent interface; invalid candidate refusal remains nonterminal. Last checkpoint identity, truthful repeated close and release now have distinguishing local tests. Original failure remains a runtime defect under adequate teaching; no new author clarification is requested. Build stage released for coordinator checks. The refined live plan remains proposed, with support freeze and live authorization explicitly pending.


### Support preparation after append repair (no Go build/live release)

Coordinator assistance read only through the authorized inbox: latest precise plan corrections and “Retained checks continuing; support readiness handback.” The first correction is my planning error: the actual CLI has no policy selector, so A/D1 capture raw0/effective16, with a separate proxy cap3. GUI B compares actual historical raw0/effective16 against current1. Second correction is now an explicit instrumentation plan for native speech admissions, application service admissions/pending work and Page queue/current/pause state around resume/reconnect/Page close; silence/screenshots alone do not establish absence. Public C/D policy uses the supported API. Inspecting the actual CLI additionally showed no scalar binding selection; D now deliberately uses empty bindings and literal prompt markers, preserving exact shared session identity without adding runtime capability. No teaching ambiguity requires an author answer.

Actual reads this phase: entire mandatory coding skill and complete pinned new-only architecture; clarification-3a9e5b7 manifest and SHA256 verification of both pinned files, with Chapter10 lines630–880 reread for identity/consumer/browser obligations (the complete chapter/direct feedback were already read in the preceding phase). Own session-api.md and persistence-format.md opening/watermark addendum search; own CLI main/selection, GUI command flags/settings acknowledgment searches, public policy/watch/skills-consumer code, common Config/Watch/Skills declarations, targeted session one-shot test, tool write implementation and own stream fixture tests. GUI application/app/Page/Connector/SpeechQueue/SpeechService, preference control searches and HTML were inspected to bind instrumentation to actual owners. Broad batched display clipped some joined outputs; follow-up individual reads covered the used GUI/public consumer code. Attempts to read nonexistent GUI main.js/boot.js and common/config.go/tools/definitions.go returned not-found; no alternate/excluded source was opened.

Permitted preceding support actually read: evidence/ch09/support-recovery/evidence.py, terminal-run.py and browser-live.mjs in the accepted main source. Reused their identity-first, deliberate PTY/browser and bounded relay patterns in newly owned support. terminal-run.py contains credential/discovery implementation text; neither it nor those paths were executed, and no credential file was opened. Did not read earlier receipts to infer a new runtime result. Own new support, local check outputs and source metadata were inspected. No grader/reviewer implementation, author research, historical/future solution, other worker conversation/memory, provider endpoint or credential was accessed. Local tests connected only to tiny self-owned loopback servers; synthetic keys and identity/build fixtures are explicitly labeled and cannot stand in for actual builds.

Support ownership: run/provider owns an fsync-before-transport attempt journal and its file lock; each relay owns its fixed upstream, request recorder and worker lifetime. Drivers own PTY/browser processes and their original receipts. Public consumer owns its Ensemble and uses public Agent methods only. Browser instrumentation retains the actual Page/application/service chain and introduces no runtime saved state or new writer. Source/build preflight checks precede derived evidence writes; historical source enumeration and complete current source/support/module/catalog/dependency identities refuse missing/extra/mismatched inputs. Future real build associations are produced only by actual successful commands. No final binding or built public consumer is claimed.

Validation: support-prebuild-checks.jsonl records14 Python tests passing, seven intended JS speech-probe negatives with positive/closed-Page controls, JS syntax checks and empty gofmt -l output; Python compile() checked all support scripts without bytecode writes. Earlier10/11/12-test attempts remain preserved. The12-test attempt passed but printed a local fake backend keepalive ConnectionResetError after the intended timeout; set that fake response to close its connection, retained the diagnostic and reran14 cleanly. No runtime fix or test weakening resulted. Go consumer is formatted but uncompiled; no go build/vet/test/module-fetch command ran. Browser/native speech/PTY integration has not run. Source-bound actual binary/preflight controls await the compiler slot. Root requested this readiness handback while retained runtime checks continue; no runtime correction is currently assigned.

## Grouped retained repair: dd1111e acknowledgment

Read the complete mandatory repository skill again (including after compaction), permitted pinned architecture, and complete chapter-10.md/direct chapter-10-student-feedback.md in clarification-dd1111e. Manifest verified against bytes and Git revision dd1111e0b09513ea59fd3b524d5446e8c5ec773f: chapter SHA256 51e32b2d336238255c5253f9f92fe3354e28ff778fdb7443e16498a0722bfd32 (66229 bytes), feedback 418d8d11ee401aa0f8eb551d9a3d37e531c0abe4a0a4cba4c64740d197062c36 (11442 bytes). Read permitted coordinator inbox additions through grouped retained repair release. Read safe retained-management-observations.json and retained-replay-observations.json metadata, failure summaries and all four replay mismatches; initial full receipt display was truncated, so this is not a claim to have visually inspected every large event transcript. No referenced checker/reviewer/historical sources were opened. Assistance consists of coordinator-supplied observations and the published new teaching; no other worker was started.

Accept the clarification: bounded syntactically valid object arguments with duplicate members reach typed tool validation in both modes. They produce paired controlled invalid_skill_arguments, no Skills mutation/Job, and consume pending limits once. This exception applies only to designated argument payloads, never structural duplicate fields, broken JSON, nonobjects or session non-scalar escapes. Unique-key arguments compare by exact canonical decimals; ambiguous arguments compare by original accepted text. Snapshot argument Raw wrappers retain that text and hash as strings.

The replay finding is an implementation defect: Chat Completions function.arguments is a STRING whose decoded bytes must survive independently of prepared object formatting. The existing call-bound opaque replay field can retain that string with its provider provenance; Engine owns parsing, checking its correspondence and rendering. Persistence owns contextual argument validation/codec/equality through the existing Agent-owned common interface. Other raw data and structural JSON stay strict. No new mutable Context or writer, sibling injection, allocator, or ownership change is needed. Actor continues using prepared accepted events and existing terminal-fault handling. No unresolved ownership conflict blocks this repair.

Validation correction: the first opaque-field approach preserved bytes but failed the existing reported-model/alias test, since opaque material intentionally requires the exact producing target. The original argument STRING is understood, portable data, not an opaque signature. Replaced that approach with a shared optional Part.arguments_text pointer, owned and validated by Engine; semantic grammar documents its narrow tool_call/OpenAI constraint and optional absence for prior snapshots. No opaque restriction is relaxed. The root suite failure is retained. Also found offline limit-set replay depended on live Registry installation; moved its typed syntax definition into the Registry owner, independent of selection, while live dispatch still enforces availability. This is an implementation correction, not a new teaching ambiguity. Two scripted unique-anchor assertions prevented edits because Args/Parts also occur in ToolEvent; no files were written by those failed scripts. Root-tests-2/3 were inadvertently repeated against unchanged source and retain the same failure.

Support preflight refused before building because main/test_ch06.log was an extra ignored file, outside the frozen source manifest. Its modification time predates this phase (Oct 7 14:52:13); origin is not established and it must not be described as a newly generated test result. Preserved its exact 16673 bytes under evidence/ch10/retained-repair-prior-main-artifact.log with SHA256/original-path metadata in retained-repair-root-test-generated.json (that metadata filename reflects the initial incorrect assumption, corrected in its contents). No transcript content was displayed or used; only bytes for SHA256 and filesystem metadata were read. No source-map check was weakened; first preflight refusal remains preserved.

Coordinator question: the new empty-name duplicate acknowledgment satisfies the literal dd1111e example, but independent management receipt reports the three duplicate/strict-ack failures (135/138 overall). Actual outputs and identities are recorded in implementation-status.md. Please resolve this independent check against the teaching; no reviewer/grader source was opened and no expectation changed. Other frozen Chapter9/10 commands pass.

Offline diagnosis: d161b121 diagnostics found exact public Context mismatch confined to structural empty collections (Hints/PendingSkills nil versus empty, and nested empty Part.Parts), while events/render/usage/Skills matched. Keep the strict comparison. Engine will normalize structural collections only in the owned public Context copy; raw JSON absence/text and stored event history remain untouched. Agent retains its existing lock during snapshot copying; no writer/owner/interface/lifetime changes. This is an implementation representation correction and does not require relaxing semantic acceptance or the support comparison.

The expanded snapshot test additionally distinguished identity-schema object formatting from replay bytes. Identity comparison is explicitly semantic in Chapter10; raw Context/Part replay payloads remain exact. Support now compares only Session.Identity as parsed JSON with UseNumber (fixed CLI schema numeric lexemes), preserving exact reflection comparison elsewhere. Positive identity-format and negative replay-number/whitespace/state controls protect this distinction. Runtime snapshot normalization affects only absent structural collections in returned copies, not identity-schema Raw bytes, event history, codec acceptance or persisted state. Initial and diagnostic failures remain retained.

After Context correction, offline branch controls reached a separate watch mismatch: only null/full reduction versus restored snapshot tool_called.parts differed (absent versus []). The semantic codec records unused lists as [], so restore must interpret the event variant: calls have no result-parts member; returned results retain a present list including empty results. Added this narrow codec decoding and a call/result presence regression. Original Events/Dump/logs are not rewritten. The support watch comparison remains unchanged. Browser support at 0d28ba7 completed and verify.py sealed/verified actual checkpoint-after-applied plus speech lifecycle observations before further source edits.

Grouped repair final read/exposure ledger: reloaded the entire repository coding
skill after each compaction, most recently as a separate unclipped read after a
batched display truncated; reread the complete pinned new-only architecture.
The complete dd1111e chapter/direct response and verified identities are recorded
above. Read authorized inbox through “Grouped retained repair and local support
build released”; its final reread contains no answer to the management question.
Inspected own current common/session/types, persistence json/codec, Engine parser/
renderer/events/session/semantic/Context-copy, Registry skill/limit validation,
Agent snapshot and associated tests, API/format docs and scoped diffs. Reads were
selected source ranges/searches as needed, not a claim to reread every entire
source file. Read own support README, build/identity/driver/browser/consumer/
fixture/verifier code and original local generated receipts, launches, binding
maps, command outputs, screenshots/DOM and comparison diagnostics. Inspected Go
module discovery/build metadata and exact Git-owned source identities. Final
receipt summaries, immutable launch revisions and source preflight were checked
again without invoking Go after releasing the compiler slot.

Assistance remains the coordinator-supplied new teaching, observation-only
retained receipts and sole inbox; reviewer closure was supplied as an identity,
not read. No worker was spawned, no checker/reviewer implementation or historical/
future solution, other worker conversation/memory, author research, credential or
external provider endpoint was opened. An ignored prior log's bytes were hashed
and preserved without displaying its contents; its unknown provenance and the
corrected initial origin assumption are explicitly retained above and in the
handback. Local fixture exchanges and prepared synthetic credential canaries are
not credentials or real-provider evidence.

Final runtime/support is57d4aac3d26edcb78dc8d9fd27d7d0e1cce0ebf4. All final
published commands except management pass, including all four fault groups after
a preserved ENOSPC attempt and explicitly authorized idle build-cache clear.
Management remains135/138 with the same three empty-name acknowledgment failures;
the coordinator question above remains unanswered. Actual local PTY/browser/
public support runs, source-specific associations,15 identity-negative controls,
bounded33 proxied plus6 local seed exchanges, failures and remaining gates are in
[retained-repair-handback.md](retained-repair-handback.md). This is the initial
student experience before historical comparison. The exact replay/duplicate
teaching resolved the original contract distinction; the later Context/watch
issues were implementation defects. No further teaching ambiguity is invented.
Compiler slot is released. No chapter/live/comparative acceptance is claimed.

### Provider support preparation (local only)

Reread the entire repository coding skill and book/edition-2/architecture.md;
reread complete pinned dd1111e Chapter10 and direct feedback, using targeted
follow-up ranges to recover the truncated middle of the full chapter display.
Verified both manifest SHA256s and exact Git blob identities again. Sole inbox
and this handoff report the management discrepancy resolved by independently
adapted138/138 checks and exact-argument replay on57d4aac. This is coordinator
assistance, not a checker run or source read by this student; original135/138
receipt remains unchanged. No runtime correction is requested or planned.

Support plan: fixed provider-origin/header/route adapter owns its in-memory key;
driver owns explicit local/live selection and subprocesses, leaving child keys
as a dummy loopback token. Existing relay/budget retains durable before-send
counting. Filter response credentials before client delivery as well as recording,
including split chunks; never retain raw authorization headers, redirect URLs or
exception strings. One bounded discovery request produces an identity-bound
selection receipt, with no pagination or fallback model. Gemini3.8Flash must
actually appear with generateContent support; availability is unknown.

Separate support revision is necessary: preserve full historical57d4aac build
source maps and actual build associations; compare current runtime/compiled
consumer bytes to them while binding interpreted support at its new revision.
Reject compiled-input or dependency changes until a new real build exists.
This extends evidence identity only, not runtime ownership or persisted state.
No architecture conflict blocks this support work. Grader owns compiler; no Go
commands, cache cleanup, credential reads or external API calls in this phase.

Provider preparation outcome before freeze:35 local Python checks pass, including
the retained relay/budget/identity controls and new provider/split-binding controls.
Initial38-check run failed discovery success: the response completed and closed
its socket, then the reader tried to set that closed socket's timeout on the next
iteration. Fixed completion handling, added explicit transport outcome assertions;
the original failure remains. Subsequent38/33/34 runs passed as controls were added
and duplicate imported unittest-base execution was removed from the command.
Final35 includes absolute header-drip and bounded DNS-wait controls and the mocked
driver's actual env/receipt construction. No Go or real client process is run by
those controls. The DNS worker owns resolution only; if its bounded wait expires,
it cannot later create a socket/send HTTP. Response credentials are redacted before
client or evidence output, not merely in the saved body. No acceptance assertion
was weakened. Python source compilation checks write no bytecode.

Additional actual reads: own complete support identity/relay/driver/tests/build/
verify/schedule/README and live matrix, selected browser bootstrap and public
consumer seed/selection code, and permitted Chapter9 support-recovery terminal
launcher lines1–80 plus targeted credential/discovery/launch searches. That source
contains credential-loading code but was not executed; the settings file was not
opened. Reused its explicit field/header mapping and safe exception classification,
adding bounded fixed-route transport rather than inheriting redirects or secret
child propagation. Consulted only official Anthropic/OpenAI/Google model-list/key
documentation for mechanics (links in support/provider-support.md); no API call or
availability inference resulted. All test keys, settings and selection receipts
are synthetic and are labeled so. No excluded source/reviewer/grader/other session,
research, memory or credentials were accessed; no worker was spawned.

Latest inbox assistance: response-to-history assertion diagnosis was subsequently
narrowed by the coordinator to unused Part.Parts nilness, with meaningful fields
and public JSON identical; no runtime repair requested. This is attributed
coordinator information, not a student inspection of checker/reviewer internals.
Deterministic clearance remains pending independent validation. Continue support
only and preserve original135/138 and all earlier failures; live stays unreleased.

Split binding created successfully for interpreted support065a6a824ec46a70affd85bd5b95cdc19439495d
against unchanged actual runtime/build57d4aac. All22 intended identity refusals
passed from that real build parent before derived writes, with a separately
labeled synthetic launch (no actual client execution or relabeling an old launch).
It covers historical/current/support source maps, missing maps, all seven binary
roles, build/module/browser/catalog identities, both launch revisions, executable
path and actual alternate support bytes. Original fixtures/bindings are retained.

Final own transport review tightened TCP→TLS deadline handoff: after a slow TCP
connect, the next handshake/send receives only remaining total time. Added a
synthetic socket timing control; all13 affected provider controls pass. No real
TLS endpoint, key or client was used. This is a support refinement, not an observed
provider/runtime failure. A new exact support revision/binding follows; the22
controls retain their initial065a6a8 association rather than being relabeled.

Final interpreted support9822b2b63257724001d7af195b377483baad3531 now has a complete
mixed-revision binding to actual runtime57d4aac, plus two passing focused final
support-source/launch-support refusal controls before derived writes. No binary
rebuild, client run or provider call occurred. Actual binary/module/browser file
bytes were hashed only for identity. Own tiny alternate support fixtures and
their intended mutation remain preserved, with copied Go/module files retained
as .fixture after controls to avoid apparent nested development modules.

Final inbox assistance reports deterministic checks passing: scoped unused-Part.Parts
comparison adapter14/14 inherited groups and19/19 intended deletions, joined70/70
source-bound result being finalized, original68/70 retained. Compiler released;
unchanged binaries preferred. These are coordinator reports, not newly read
checker implementations/receipts or student-executed results. No Go build is needed
for this prepared layer. Concrete support review/live release remains pending;
stop at provider-preparation-handback.md with no credential/API access or chapter
acceptance claim.


### Initial live pass: released phase6

Read complete mandatory skill, current architecture, full pinned dd1111e Chapter10 in three unclipped ranges, full final provider-support.md, live-matrix.md and support README/binding documentation. Verified pinned chapter and feedback manifest SHA256 and Git bytes again (feedback text retained from preceding full read). Sole inbox publishes support-review0a78688 and deterministic213b56b clearance and explicit live release; reviewer/checker implementations remain unread. Phase5 unreleased wording is historical and is superseded by this concrete handoff. Complete actual runtime57d4aac/support9822b2b preflight passed before keys/network. No rebuild or compiler use.

Reviewed adapter made exactly one successful discovery request per vendor, all200/no pagination. Selected returned claude-sonnet-4-6, gpt-4.1 and exact models/gemini-3.8-flash with generateContent support. Anthropic/OpenAI preference uses permitted own Chapter9 student-review selected-model evidence, not historical answers. Actual additional reads: own CLI selection/main searches, Engine ValidateConfig/Route and selected prior own Chapter9 review lines571/600/608/618 surfaced by model-name search; attempted nonexistent llm/models.go/config.go and root model* reads returned not-found. No other source was followed. Selected needed key fields were loaded only through reviewed provider.load_key in adapter processes; no settings dump, shell extraction or child key propagation. The exact66/3 ceilings and independent/local seed distinction remain. Actual interface work follows; discovery alone establishes no generation success.


### Initial real-use experience frozen before historical comparison

Actual phase6 pass used runtime57d4aac3d26edcb78dc8d9fd27d7d0e1cce0ebf4 and
reviewed interpreted support9822b2b63257724001d7af195b377483baad3531, with unchanged
actual binaries and the final mixed-revision binding. The student coder drove
human chat in real PTYs and actual headed Chromium controls; Bill did not perform
these interactions. Native speech was instrumented for admission/ownership only;
I make no hearing claim. All33 generation attempts returned200, with11/vendor
(A3/B1/C3/D4); exactly3 discovery calls succeeded. No retries, pagination,
alternative models, capability probes or repeated prompts. All planned model
turns completed; B intentionally ended round_limit after its one real write call.

Stable session identity, exact marker recall without tool replay, checkpoint ack
ordering, current policy1 versus historical raw0/effective16, font16→18, true GUI
server restart, socket reconnect, Page close/reopen and terminal-only EOF were
observable through the actual clients. Checkpoint still worked after terminal EOF.
Historical tool cards correctly said no live owner. Measured restored native and
service speech admissions, queues, work/pause ownership and provisional content
were zero. C's two actual Agents completed independently; source-bound public
imports recalled the marker, preserved origin bytes and usage once, and refused
genuine pre-origin reconstruction. Offline latest/older/null branches agreed on
render, state/history/usage/Skills/watch and requests with endpoints disabled.
Fifteen public reconstructions also match actual captured request bodies byte for
byte (five original sends per API), without argument-string normalization.

D loaded/retired/reloaded through public controls, used write_file from real PTYs,
then resumed from the unchanged relocated catalog. Retained activations1/2/3,
material coordinates2/3/5 and exact manuals survived. The explicit empty-pattern,
17-byte one-shot setting consumed exactly once before the real next call and was
absent afterward. The two seed exchanges per vendor are synthetic, separately
retained, and charge zero real requests: input2/output2 per vendor are separately
identified within cumulative usage, never attributed to the provider. OpenAI's
actual response provenance is gpt-4.1-2025-04-14 although the selected alias was
gpt-4.1. Gemini actual signature metadata was present and retained; no signature
text was treated as user-visible prose or as a credential.

The persistence teaching's distinction between history, live owners and current
policy proved useful in the clients. I found no new unresolved persistence-owner
contract conflict. Specific initial experience findings for coordinator/author:

- Gemini's resumed browser shows extra empty “Answer / Accepted” cards. My first
  screenshot suggested missing answers; preserved full DOM proves the real text
  is present in separate cards. Public completions retain empty text Parts with
  opaque provider metadata. That is a plausible presentation cause, not a proven
  reducer defect. Preserve metadata; review suppressing empty presentation cards
  in a later scoped correction. No live rerun or runtime edit was improvised.
- Anthropic D1 proposed invalid names “scratch activation 2” and “scratch
  activation 3” while also performing the requested write. Both management calls
  had paired controlled refusals; the turn completed. Its D2 explanation confused
  tool-report byte length with file size; actual after-resume.txt was verified
  exactly17 bytes. These are observed model explanations, not fabricated success
  text or changed receipts. Teaching should encourage file/effect inspection.
- On the first Anthropic GUI restart shutdown I sent server SIGTERM before EOF to
  the attached terminal; the supervising recorder waited until terminal EOF. I
  then sent EOF and observed exit0. Later shutdowns explicitly detached first.
  This is worth making explicit in the combined-terminal spin instructions.
- The initial live-observations.py summary failed with KeyError(state), because
  lifecycle probes deliberately record before/after/closed shapes. The preserved
  observations-initial.out and commands journal show exit1. The corrected reader
  examines the restored/closed side and passes; no original or runtime changed.
  A prior exploratory read also assumed every events.log line had type; the
  version header caused KeyError before any write. No runtime failure inferred.

Exposure ledger additions: reloaded the ENTIRE mandatory skill after compaction,
re-read entire permitted architecture, own driver.py/local-fixture.py/schedule,
consumer C/D/offline functions, verify.py, selected provider loader/redactor and
own live receipts/DOM/screenshots. Re-read full own live-matrix. Inbox checked at
meaningful boundaries, with no release changes. No grader/reviewer implementation,
other conversations, memories, old/future chapters/answers or author research was
opened. No delegation, compiler use, cleanup of caches or source, or runtime/support
repair. Summary/audit helpers are new evidence-only files outside the reviewed
bound support tree. Initial original evidence stays distinct from reconstruction.
No comparative feedback, chapter acceptance, tag/export or push is claimed.


Evidence-freeze tooling also hit an exclusive-create filename collision:
`gemini-B-restart-seal.out` already described GUI attachment sealing. The later
originals-sealing batch stopped rather than overwrite it; the preserved
freeze-label-collision.json records the observed error/cause retrospectively.
Resumed only unfinished seals with distinct originals-seal labels. No model call,
runtime/support source or original changed. This is evidence naming friction,
not a teaching or persistence-runtime failure.

Final freeze:30/30 source-bound original verifications passed; safe file and staged credential audits passed. Final inbox check remained the same explicit live release. All configured-key checks emitted only aggregate results. Initial real-use record is frozen for independent review; runtime/source/support identities unchanged.


### Grouped post-run quality acknowledgment: Q1/Q2/Q4

Read entire mandatory coding skill and architecture before edits. Verified pin
0e754b9cad74a3a1b44327e498fa704feea817ed: chapter SHA256
adcd127664db5a1df99d4c5090d09df292181f5b7aa890ff5511b72f3f90fcc7;
direct feedback SHA256 2a94ee3a6b03025c5be33acb7139a18dcb7b092eb48eed935fa8e0fde1eef723,
both against manifest and exact Git bytes. Read the complete chapter1121 lines
and complete direct response217 lines; repeated580–820 and821–1121 after a combined
tool response truncated. No linked reviewer/old sources followed. Authorized inbox
and user supply accepted initial-live review23b312c and historical-comparison
rationale only; reviewer implementations/reports and historical answers remain unread.

The actual walkthrough matches my initial experience: exact OpenAI identity and
usage, actual coder PTYs/headed browser, policy versus historical authority,
checkpoint after terminal detach, immutable-origin/public replay comparisons,
separate local seeds and the Anthropic incorrect explanation are faithfully
attributed. Its narrower opener correctly acknowledges prior durable logs/offline
rendering. The remaining Gemini causal wording is explicitly tentative, but Q4
now makes it precise: four empty text parts at index1 of response sequences5/10/15/20
include two signed and two ordinary empties. My initial guess did not establish
that all empties carried metadata. Preserve all original source/parts/screenshots.

Disposition before affected changes: Q1 add a constant-size committed activation
maximum read on the real Skills owner, under its existing Actor/Agent capture or
inert construction confinement. No cache, second state or new lock; failed/unchanged
candidates cannot advance it and retained retired records establish its maximum.
One full owned Snapshot remains for capture. Q2 correct the early format sentence
to exact activation/job maxima, zero absent, preserving the burned-request exception.
Q4 change only optional Artifact presentation of present exactly-empty response text:
keep durable identity/order, show compact truthful empty-text indication, omit
speech/expansion controls, restore ordinary controls on nonempty replacement.
Absent text, whitespace, tool-result lifecycle and opaque placeholders stay distinct.
No projection/reducer/raw-history edit or provider call is authorized/needed.

Additional reads: own session capture/install/open, Skills snapshot/ledger/common
interface and narrow-read tests; own GUI Artifact/style/speech/Connector/projection
tests; own prior local browser support and phase6 browser recorder. Exploratory
reads for nonexistent agent.go and internal/skills/service.go returned not-found;
actual owner source is ensemble.go and skills.go. All writes stay in main. Compiler
is released exclusively for this grouped local task; initial44d7627 receipts remain
immutable. Browser regression will use separately labeled captured-data evidence.


Grouped quality result: source70d86f7419c82fcf7cb8a394d54e472feccd2eed is frozen.
Core/GUI format/vet/tests and affected race checks pass; public checker16 groups,
Chapter9 51/51, Chapter10 initial93/93, clients9/9 and targeted append/Close faults
pass on the new source/binaries. No checker implementation read. The local headed
captured-data browser check passes43 checks. It confirms response5/20 part1 are
ordinary empties and10/15 part1 signed; all durable positions remain, with compact
truthful indication. Two resets, identity handoff, replacement, ordinary whitespace/
absence, opaque and tool-result controls and speech negatives/positives pass.
I viewed its new screenshot. Initial44d7627 directory/handback/binding diff is empty.
A valid-parent intended source mismatch refused before creating derived output;
owned source was restored exactly. That expected negative is not a runtime fault.

The first core vet attempt failed before compilation because the new local command
recorder omitted HOME. Corrected its environment and preserved initial failure;
no cache cleanup. The initial broad process-text probe matched its own command;
precise executable-name inspection then found zero active compiler/linkers before
Go work. Build associations retain full source maps, discovered12 modules and real
binary metadata. No need to repeat unaffected modules/full paid matrix. No runtime
failure was observed in this grouped check.

Walkthrough confirmation remains affirmative. One narrow prose nuance for author:
the actual OpenAI A1 continuation arguments inspected here are compact JSON text;
the15 byte-exact live comparisons preserve those actual bytes. Do not imply a new
deliberately spaced-string provider fixture; that distinguishing regression is
separate earlier local evidence. Q4 is an authorized post-run presentation quality
choice, not a retroactive chapter requirement. No new owner/format ambiguity.

Actual assistance/exposure additions: authorized inbox rationale Q1/Q2/Q4, own
common/Skills/Agent/Artifact tests and state, own served browser modules, own
captured original Gemini/OpenAI records and derived local outputs, old binding
metadata and browser dependency bytes for hashes only. No historical solution,
reviewer/grader implementation, other conversations/memory or author research was
opened. No credentials/network-provider access. New binary is not claimed live-
tested; only empty-card appearance is superseded by the local quality evidence.
No broader initial live observation is invalidated. Review of evidence reuse and
new source is still independent; compiler work complete, slot released.
