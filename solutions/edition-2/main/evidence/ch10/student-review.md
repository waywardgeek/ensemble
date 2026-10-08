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
