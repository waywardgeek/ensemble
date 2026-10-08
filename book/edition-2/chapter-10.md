# Chapter 10: Return to the conversation

The useful test of persistence comes the next morning. The Agent remembers the
file it changed, the reason for the change and the tool result that confirmed
it. The reader can continue the work without asking it to perform the edit
again just to recover the explanation.

The first edition made loading an option. Bill's response, recorded in old
Chapter 11, was short: “Let's load by default without a flag.” Automatic loading
makes the application easier to use and makes one error much worse. If an
unreadable save causes a fresh start, that new conversation can overwrite the
history the reader wanted back. Refusing to open it preserves a chance to repair
it. Starting over silently removes that chance.

This edition already has recorded events and offline rendering. The missing
piece is a safe way to mount those facts as a live conversation. A saved shell
job cannot become a process again. A retired skill cannot acquire new permission
because its manual survived. Remembering requires fewer powers than resuming.

**Contract draft:** the coordinator accepted the outline and its construction,
ownership and storage direction. The detailed format remains under review.
The Chapter 9 implementation and Chapter 10 independent
checker are not yet accepted. No student implementation, successful restart or
actual spin is claimed.

## TL;DR

Extend only the accepted Chapter 9 source when released by the coordinator. Read
the architecture ledger and the entire
`book/edition-2/skills/ensemble-coding/SKILL.md` before code. The optional GUI
continues to consume the public library from its separate module.

1. Keep ordinary fresh Agent construction and offline Load. Add an explicit
   public open-or-resume session operation with creation-only DataDir and an
   Agent-owned SessionStore. Human CLI and GUI choose it by default. Explicit
   standalone-log selection retains its fresh behavior; never migrate it silently.
2. Give the persisted conversation a SessionID separate from runtime Agent.ID.
   A mounted Agent gets a fresh runtime identity. Restore historical sequence,
   part, call and activation identities, and prevent identity/handle reuse.
3. Keep one actor-ordered durable append path. Publish the exact outer format
   below and a strict versioned semantic snapshot codec. Snapshot plus later
   events and independent reduction must agree on context, usage and public facts.
4. Checkpoint at an idle, settled actor boundary; otherwise return busy. Capture
   owned values, then write through one off-actor worker and atomic replacement.
   A completed save reports its anchor. It cannot claim later events were inside.
5. Validate the entire candidate before exposing a live Agent or sending HTTP.
   Corrupt, incompatible, partial or unfinished state refuses live resume and
   remains available for honest offline diagnosis. No retry or invented terminal
   outcome makes it resumable.
6. Restore recorded history and pending one-shot tool limits. Never restore jobs,
   process ownership, queued prompts, pause registrations or provisional speech.
   New limit set/consume facts make complete-log reconstruction possible.
7. Require compatible caller-selected skills, bindings and handler definitions.
   Recorded base System belongs to session identity; current model, credentials,
   route, policy and GUI preferences retain their separate authorities.
8. Hold a nonblocking OS lock throughout live store ownership. Exercise restart,
   close, signals, imports, corruption, writer exclusion and stale handles through
   public consumers and the actual CLI/browser, on all three supported APIs.

Build the main CLI with `go build ./cmd` and the optional GUI from `main/gui`
with `go build ./cmd/ensemble-gui`. The inherited diagnostic is
`make grade-dir CH=11 DIR=solutions/edition-2/main`. It does not cover this
contract. The independent Chapter 10 acceptance command is pending publication;
that absence blocks student release. Section 10.10 maps the required checks.
Public Go names and private payload field spelling remain student choices;
shared ownership, outer bytes, semantic requirements and errors do not.

## 10.1 A session has an owner

A file path used to identify an output log. It now also selects work to resume.
Those are different construction requests, and guessing between them would
make existing embedding applications load a conversation they never requested.

Keep the preceding fresh constructor: it creates its explicit log exclusively,
refuses an existing destination and returns a new Agent. Keep offline Load and
render as read-only operations with no actor, HTTP, process or tool execution.
The new session-opening API deliberately selects automatic creation or resume.
Its public configuration supplies DataDir, ordinary request configuration,
optional policy path and any Chapter 9 skill configuration. It returns either
a completely validated live Agent or an error. A public checkpoint exporter and
explicit snapshot-import operation work without the GUI.

Agent owns DataDir, SessionID, its authoritative creation configuration and
SessionStore. SessionStore owns resolved file identities, the lock and checkpoint
worker; it reaches its Agent parent and the Ensemble logger. Actor orders state
capture and every durable mutation. Engine retains usage and request transport;
Skills retains its frozen catalog and material state; Jobs retains live jobs and
pending limits. Their parent interfaces make those facts reachable without
injecting siblings into the store.

Shared serializable values and interfaces belong in common. Storage behavior
belongs in the persistence spoke; reducers and service-specific validation stay
in their responsible spokes. A snapshot is an owned value of existing state,
not a second conversation whose fields another worker can change. Private runtime
implementations remain permitted under the architecture ledger.

Ensemble reserves both the resolved store and SessionID while mounted. A second
mount of either in the same application fails. Across applications the OS lock
protects the store. A copied store keeps its SessionID; copying files does not
create a fork operation or permission to mount that identity twice in one root.
This chapter does not add a fork or migration command.
Reserve/release under a brief root lock; validation and filesystem work must not
hold that lock or block unrelated Agents.

A runtime Agent.ID still identifies a child of the current Ensemble. Never
restore the old application counter as that ID. Keep the durable identities
inside events, and expose SessionID separately so a client can recognize the
conversation after a process restart. Watch generations, operation IDs,
completion handles and connection registrations remain new runtime objects.
A new application may allocate the same Agent.ID text as an old application;
its scope remains that Ensemble. Do not promise global uniqueness for an
application counter. SessionID supplies the durable conversation identity.

## 10.2 Choose the directory before opening anything

Human chat and the GUI default to `.ensemble/session` under their launch
workspace. Both accept `--session-dir PATH`; resolve it once without changing
process cwd. A public caller supplies DataDir explicitly. A relative selection
uses the application's declared launch workspace. Resolve it before tool work.

The store contains `events.log`, `checkpoint.json` and `owner.lock`. An explicitly
imported snapshot-only store also has immutable `origin.json`, explained below.
These are ordinary files under the resolved directory. Reject symlinked store
leaves and nonregular files; canonicalize the existing directory portion for
path reservations. This is file-identity hygiene, not protection against another
program concurrently replacing directories or against an enabled shell tool.

| Entry path | Selection | Behavior |
|---|---|---|
| Human CLI or GUI | No log/session selector | Open or resume `.ensemble/session` |
| Human CLI or GUI | `--session-dir PATH`, no CH02_LOG | Open or resume PATH |
| Human CLI or GUI | Explicit nonempty CH02_LOG, no session selector | Preserve legacy fresh, exclusive standalone log creation; no automatic session loading/checkpoint |
| Any live CLI/GUI mode | Both CH02_LOG and `--session-dir`, even if paths appear related | Refuse conflicting selectors before opening state |
| Existing machine protocol | No session selector | Preserve the preceding fresh-log default and protocol bytes |
| Existing machine protocol | Explicit `--session-dir PATH` | Deliberately open/resume that session; emit only its existing protocol records |
| Public fresh constructor | Explicit LogPath, no DataDir | Preserve fresh semantics; a nonempty DataDir refuses with session_conflict and names the separate session operation |
| Public session opener | DataDir, no explicit LogPath | Resolve the writer to DataDir/events.log |
| Public session opener | Any explicit nonempty LogPath | Refuse conflicting selectors; do not import or redirect the standalone log |

An explicitly supplied blank CLI selector is invalid. Shared configuration must
retain presence before filling defaults, especially for System. Returned live
Config reports the actual event-log destination. Its unchanged round trip stays
valid; SetConfig still refuses changing LogPath or DataDir. Offline `dump` and
`render LOG` retain their preceding standalone selection and read-only meaning.
The new `session inspect DIR` command reads a session store without live keys,
prints safe identity/boundary information and returns nonzero on corruption.
It does not turn an unfinished conversation into a live one.

Acquire the store lock before reading files for live opening or importing.
Use an exclusive nonblocking OS advisory lock on owner.lock, held until close
has joined owned writers and closed the log. Support macOS and Linux local
filesystems; other platforms return `session_unsupported`. The lock descriptor
must not be inherited by tool processes. Do not unlink or replace owner.lock on
close: another opener must lock the same file, rather than a newly created inode.
The descriptor-based lifetime and nonblocking operation are documented by the
[Linux manual](https://man7.org/linux/man-pages/man2/flock.2.html) and
[Apple's manual](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/flock.2.html).

A process dying releases its noninherited descriptor; the remaining filename
is not a stale ownership claim. An existing holder returns `session_in_use`
immediately. Advisory locks require cooperating writers. Network filesystems,
malicious replacement, distributed leases and power-loss recovery are outside
this chapter's guarantees; do not infer them from a successful local test.
Offline inspection acquires a shared nonblocking lock while reading a store and
refuses an active writer. Inspecting a separately exported immutable checkpoint
needs no live-store lock.

After locking, use this table. A failed read never means missing.

| Files found | Result |
|---|---|
| Directory absent, or directory containing only owner.lock | Create a fresh session exclusively |
| Neither checkpoint nor log, but other contents | Refuse ambiguous/incomplete store; do not erase contents |
| Full-origin events.log, no checkpoint | Validate and rebuild; a checkpoint is optional |
| events.log and checkpoint | Validate identity, prefix/anchor and tail together |
| checkpoint without events.log | Refuse incomplete store; explicit import is a separate operation |
| Anchor-origin log without its origin.json | Refuse missing origin |
| origin.json without its matching anchor-origin log | Refuse incomplete import |

Uncommitted temporary files in an otherwise valid store do not become state and
are not silently promoted. Partial initialization/import may leave diagnostic
artifacts. A subsequent open refuses incomplete state rather than guessing which
step committed. No automatic destructive migration from first-edition save.json
or an earlier standalone log is provided.

## 10.3 Bind the snapshot to the conversation

A snapshot without a boundary cannot tell whether the last tool result is
already inside it. Applying the whole log duplicates it; ignoring the log loses
what arrived afterward. The checkpoint names that boundary and the history it
belongs to.

The outer checkpoint object has exactly these required fields:

| Field | Value |
|---|---|
| `version` | Integer 1 |
| `session_id` | 32 lowercase hexadecimal characters, freshly generated from 16 cryptographically random bytes for a new session |
| `identity` | Creation identity object specified in §10.7 |
| `as_of` | Positive integer sequence folded into state |
| `high_watermarks` | Object with nonnegative integer `event`, `request`, `activation`, `job` |
| `state_version` | Integer 1 for the student's documented strict semantic codec |
| `state` | Complete semantic snapshot object, or null for complete-history reconstruction |
| `state_sha256` | SHA-256 of canonical state bytes, or null exactly when state is null |

`high_watermarks.event` equals as_of. Every other watermark is at least the
largest corresponding allocated durable identity represented by state; imported
snapshots cannot reset one below a recorded use. Exact safe-integer bounds appear
in §10.8. Hashes are lowercase 64-digit hex. A hash detects mismatched bytes; it
does not authenticate a file that its operator can edit.

For session turns, add positive request_index to turn_started.turn. It is the
request allocator ordinal assigned at admission, independent of the public
request ID's string spelling. Queued cancellation may burn an ordinal without
ever producing a turn event. Snapshot request is the captured allocator cursor;
complete replay establishes at least the maximum recorded request_index. The
snapshot may be higher, but never lower. Resumed allocation advances from its
validated cursor and also rejects every previously recorded request ID. No new
admission event or synchronous disk write is required merely to enqueue a prompt.
Older standalone turn records keep their existing validation without this field.
Session request_index values strictly increase in recorded turn order; gaps
from canceled queued requests are permitted. A tail index must exceed the
captured cursor it follows, and overflow refuses admission before reuse.

Define canonical JSON for these hashes: valid UTF-8, no duplicate members, no
insignificant whitespace, object keys sorted by UTF-8 byte order, array order
preserved, strings escaped with JSON's short control escapes and lowercase
`\u00xx` for remaining controls, otherwise literal UTF-8. Do not HTML-escape
`<`, `>` or `&`, escape `/`, or normalize Unicode. Integers use shortest base-10
notation. Noninteger numbers use shortest round-trippable binary64 decimal,
with negative zero written as 0. Reject nonfinite numbers. Parse raw numbers
without first rounding identity integers through a float. The public exporter
and importer use this same documented codec.

The student publishes the state object's required fields/types in the source's
persistence-format document before independent checks. Its names may follow the
existing shared data. Strict decoding refuses missing, unknown, duplicate or
inconsistent required structural state, including nested objects. Explicit
arbitrary-JSON fields such as tool arguments retain their inherited schemas;
their data keys are not new snapshot fields. Required contents are:

- Session metadata agreeing with the outer session_id, identity, as_of and
  watermarks; copying another session's state cannot silently relabel it.
- Ordered accepted entries and typed parts with original sequence, call/part
  identity, purpose, provenance, opaque material and exact text/reference bytes.
- Effective redaction spans, pending ephemera and their consumed status, enduring
  instructions, base identity and any deferred ordered material needed by the
  inherited reducer. A settled live candidate has no deferred unresolved batch.
- Seen calls/requests, pairing and terminal turn outcomes needed to reject reuse
  and invalid subsequent appends. Dropping rendered text never erases identities.
- Per-producing-model normalized usage and associated accepted-response facts
  needed to restore totals exactly once, including raw/requested provenance.
- All active and retired skill activation records, immutable material, roots,
  offers, contributors, revision and activation allocator state, or explicit
  no-skills state. Retain ordered transition provenance sufficient to validate
  each prior root's visibility, exact action and retirement under Chapter 9.
  It references immutable activation records by ID; the primary/material bodies
  appear once. Current active roots alone cannot prove a past load was allowed.
- Historical job snapshots and tool result references, plus pending typed
  one-shot limits. Historical jobs carry no process, report cursor or live right.
- Allocator state and enough recent durable-event data/counts to reproduce the
  Chapter 7 window without pretending that 100 events are the whole history.

A serialization of event history named state does not satisfy the semantic
snapshot requirement. Loading a snapshot must install validated reduced state;
it cannot secretly depend on the omitted log prefix. Equally, a second mutable
Context maintained only for saving is not the solution. Export owned values from
the actual authorities at one actor boundary.

Session logs retain the exact `{"log_version":1}` header. A fresh store begins
with session_initialized at sequence 1, payload session containing exactly
session_id and identity. It is construction-only, precedes skills_initialized
and ordinary conversation facts, and occurs once. Skill mode requires its matching
initializer before the new Agent is exposed. An incomplete skill initialization
refuses live resume. A standalone log without session facts retains its old
offline/fresh contract.

Public append cannot initialize/adopt a session, replace its identity, insert
an anchor or reset a counter on an exposed Agent. Offline session reduction
recognizes these facts only in their allowed initial position. An interior or
second initializer/anchor is corruption before any candidate is published.

## 10.4 Apply the tail once

For a full-origin store, events.log starts with session_initialized and contains
the complete history. Validate increasing sequences and every inherited event
transition. Imported standalone version-1 gaps keep their old offline meaning;
new session writers are contiguous. A full-origin session prefix must begin at
1 and have no gap, including across the checkpoint boundary.

With state at S, validate the complete available prefix through S, reduce it
independently and compare its semantic projection with the decoded snapshot.
Then apply each event above S once, in order. An anchor beyond the log, a missing
boundary event, inconsistent identity or mismatched state refuses. With state
null, reduce the complete log from its initializer; the envelope's as_of and
watermarks still must describe the corresponding prefix. Nulling state and
state_sha256 in an exported complete-history checkpoint is the documented
rebuild control. All other envelope validation still applies.

This implementation checks its snapshot against the available prefix. That costs
replay work and is not advertised as faster startup. Stream the log once, build
indexed validation state, and compare owned semantic values without rereading
every earlier event for each new one. Rebuilding for a test should not accidentally
become quadratic merely because histories grow.

A snapshot-only import has a different proof boundary. Its explicit public
operation accepts a non-null complete checkpoint and a destination with no
existing session files. It validates the whole snapshot, caller compatibility,
identities and settled boundary before creating an imported store. Preserve
SessionID; import is relocation/inspection support, not a fork. It neither reads
missing skill files during historical rendering nor grants live authority without
the compatible current configuration.

Retain those exact accepted checkpoint bytes as immutable origin.json, written
exclusively. events.log begins with its normal header and one session_anchor at
sequence S+1. Its session payload contains exactly session_id, origin_as_of,
origin_sha256 and high_watermarks. origin_sha256 hashes the entire original
checkpoint file bytes, including any final LF; the other fields exactly match
that origin's as_of and watermarks. No history before S is manufactured. The
first ordinary new event receives S+2. The anchor itself advances the event
watermark and cannot be injected through public append.

A full-origin store must not contain origin.json. An imported store must retain
it unchanged even when checkpoint.json is replaced later. To validate a newer
checkpoint at C, seed from origin, validate its first matching anchor, reduce
the available tail through C, and compare that projection with the newer state.
Continue above C once. With no newer checkpoint, start from origin directly.
All event sequences after origin are contiguous. The anchor's SessionID, hash,
watermarks and sequence must all agree; matching only its filename is insufficient.

A newer checkpoint cannot erase or supersede an inconsistent origin. A null-state
checkpoint still requires that origin in an imported store. Log-only standalone
loading of session_anchor refuses `session_origin_required`: it cannot rebuild
the missing conversation. Offline session inspection supplies the recorded
origin without claiming complete-history equivalence for events that are absent.

Compare snapshot and rebuilt paths under identical current render inputs. Require
byte-identical next requests for Messages, Chat Completions and generateContent,
plus equal usage, skill state, durable identities and safe watch projection.
Exclude runtime Agent IDs, watch generations and other explicitly transient
fields. Allocate safely even where transient canceled handles burned an unused
request number; never reuse a request ID present in recorded history. Compare
the request cursor's lower-bound/no-reuse invariant, not equality of invisible
ordinals burned by dead callers. Event, activation and historical-job watermarks
must equal their corresponding reduced durable maxima.

An older snapshot with a genuinely newer log is the distinguishing fixture.
Testing only the application's latest checkpoint produces an empty tail and
lets an implementation that ignores every tail pass. Keep a second fixture
with no preceding log and a complete snapshot, then prove a subsequent save and
restart still find its immutable origin.

## 10.5 Capture without holding the conversation hostage

A user who asks to save during an unresolved tool batch needs a clear answer.
Saving half the batch and inventing the other half on restart would decide
whether an effect should run again. Serialization cannot choose that for its caller.

Checkpoint capture runs as an actor control. It succeeds only when there is no
active turn, queued admitted prompt, active response slot, unresolved call,
model/report/input worker operation, or deferred batch material. Otherwise return
`session_busy` immediately; do not queue an unbounded series of save requests or
wait for a paused model call to finish. A running independent job whose initial
report was paired does not itself make the conversation unsettled. Its current
accepted lifecycle facts may be captured as history.

At that boundary, copy the required owned state and the last durable sequence.
Jobs must supply the durable accepted job projection, not an unrecorded worker
status that raced ahead of its event. The same rule applies to skill candidates
and usage. Copying has a bounded cost proportional to retained state; do not claim
that a large snapshot takes no actor time. Encoding, hashing and disk I/O run in
SessionStore's worker after the owned capture. The actor can then accept new work.

Only one checkpoint worker may be active per store. A second save returns busy.
No lock needed by controls, job facts or observation spans disk I/O. The worker
returns its actual result to the actor, which acknowledges the anchor it wrote.
An observer failure cannot undo persistence. An event arriving after capture
belongs to the append-log tail, even if it arrives before the acknowledgement.

Write a private temporary file beside checkpoint.json, check complete writes,
sync and close, then atomically replace the destination. The successful replace
is the commit. Remove uncommitted temporary files on ordinary failure. Failures
before replacement preserve the previous checkpoint; a failure after replacement
cannot claim rollback. A checkpoint failure does not erase a healthy append log
or turn later durable events into unrecorded ones. Report its safe error and
allow a deliberate later checkpoint if the Agent itself remains healthy.

Keep the append log intact. There is no truncation, rotation, retention or
multi-file transaction protocol in this chapter. Existing event persistence
still orders write before apply/observe, and a failed append faults the Agent.
An incomplete final record is refused under Chapter 2; do not drop the line or
invent an outcome so that a damaged session appears settled. Checked writes and
whole-file replacement do not promise survival of every filesystem or power-loss
failure. State exactly which fault was injected when demonstrating recovery.

Orderly close first stops admission and follows Chapter 5: settle queued/active
requests, stop and join cancellable model/report workers, kill managed process
groups, drain output and record truthful terminal facts. Nonkillable Go handlers
retain their inherited suppression/lifetime limitation. Join an already started
checkpoint worker, then capture/write the final settled state before closing the
log and releasing ownership. Repeated close is safe. A save error is returned
and causes the standalone application to exit unsuccessfully.
If the durable append path already faulted, clean up without replacing the last
checkpoint from an incomplete candidate. Return the persistence failure and
preserve the refused log for diagnosis.

Route SIGINT and SIGTERM through that close path. Old Chapter 20 recorded a save
that existed in code but was unreachable from its signal exit. A deferred write
after os.Exit cannot rescue it. A second forced signal and SIGKILL do not promise
orderly completion. The lock must still become available after process death
without deleting owner.lock.

Standalone CLI EOF and /quit close the application. Combined GUI --terminal EOF
still detaches that client while browser/server/Agent remain alive. Closing a
browser Page closes its registration and speech work, not the session. Preserve
these distinctions when adding automatic saves; an old harness's stdin close is
not authority to shut down today's GUI.

## 10.6 Restore a setting, never a process

The saved command may say running because its first report returned before the
process finished. That report correctly completed the original tool call. It
says nothing about whether an OS process exists in a later application.

Live resume requires a settled reduced conversation: no active turn or response
slot, unanswered call, pending human dialogue awaiting completion, or deferred
batch material. Errors and interruptions already committed under their original
rules can be settled. An unfinished accepted turn refuses with
`session_unfinished`; offline inspection can still report the last accepted
boundary. No startup HTTP request, tool invocation, completion handle or terminal
event fills the gap. The reader must preserve/inspect that history or choose a
new session explicitly. Automatic crash-work recovery is deferred.

Restore neither process IDs nor input channels, readers, output cursors, workers,
queued prompts, canceled caller contexts, typing/speaking causes or provisional
model/speech fragments. Historical accepted cards retain their event/part/call
identities. Old job records stay available as evidence, with no live owner.
wait_for_job, send_input and kill_job against such a handle return unavailable;
they cannot target an unrelated process or a newly admitted job with the same
number. Reserve the restored job high-watermark through Ensemble's actual
allocator before permitting any new job. Continue occupied-artifact skipping
and exclusive creation under Chapter 4. Do not fetch artifact locators on load.
Moving a store does not move the referenced files. Retain their recorded locator
bytes and provenance; the interface must not promise that a historical relative
path still identifies available content in a different workspace.

Pending tool_limits is different: it is an Agent setting for the literal next
attempt, not a worker. Jobs retains its typed optional fields across a clean
restart. To reconstruct it without parsing old human-readable result text, extend
the existing version-1 vocabulary with these exact facts:

| Event | Payload key | Required payload |
|---|---|---|
| `tool_limits_set` | `limits` | `call_id`, `overrides` containing exactly the nonempty subset of the three valid limit fields supplied to the setter |
| `tool_limits_consumed` | `limits` | `call_id`, `name`, `overrides` exactly equal to the pending typed setting being cleared |

The fields are ai_callback_delay, ai_callback_pattern and max_output_bytes with
Chapter 4's types, ranges and presence semantics. An explicit empty pattern
remains present. Do not serialize compiled regex/runtime objects or silently
clamp a previously valid value. Tool wire parsing stays in Tools; Jobs supplies
typed candidates through its owner, and Actor commits their facts through the
one append path before applying them.

At the literal next attempted accepted model call, commit consumption before
validation/permission refusal or execution. That includes an unknown/disabled
name, invalid limits, invalid setter, supervision call, skill-management call,
and an interruption's paired attempted-call refusal under inherited rules. A
malformed response created no accepted call and consumes nothing. A paused call
has not yet been attempted; pause alone cannot clear the setting. Typed public
skill controls and checkpoints are not model calls and consume nothing.

A successful tool_limits attempt commits tool_limits_set after its tool_called
and before its successful tool_returned. A second setter first consumes the old
setting, then may set the new one. An invalid setter consumes the old state and
stores nothing. Consumption keeps the existing visible note and precedence:
defaults, then pending fields, then legal explicit fields on the consuming call.
The fact records what was consumed even when argument validation prevents an
effective setting from being calculated. No pending state means no consumed fact.

Validate each fact against its accepted unanswered call, name/arguments, prior
pending setting and event order. A call can consume only once. Set requires a
valid admitted tool_limits call and its exact supplied overrides; an unrelated
public append cannot create a new setting. Reject duplicate/set-without-call,
wrong copied overrides and missing required consumption before tool_called can
admit an effect. Reject a successful setter result lacking its set fact before
committing that result. As with Skills, live append must match the typed
candidate; plausible replacement state is insufficient.

A durable consumption followed by a log failure remains consumed. The terminal
append-failure rules and unfinished-resume refusal prevent a later process from
pretending the failed attempt never happened. Replay applies these settings facts
without dispatching a call. Newly created session logs require them; legacy
standalone logs remain offline-readable under their previous rules, with no claim
that their prose receipts establish a resumable pending setting.

## 10.7 Check today's authority against yesterday's identity

The file can explain why a tool was available yesterday. It cannot install that
tool in today's application. Reopening must compare recorded identity with what
the caller deliberately supplies now.

The reader should be able to move an unchanged catalog to a different directory
and continue the conversation. Changing the manual or its grants is a different
request. Refusing that mismatch gives the reader a chance to choose the intended
catalog before another tool is admitted.

The identity object has exactly mode, system, skills and handlers. mode is
`plain` or `skills`. In plain mode, system is the exact nonempty creation-time
base instruction and skills is null. In skill mode, system is null and skills
contains exactly primary, catalog_sha256 and bindings_sha256. The primary body
remains in its Chapter 9 activation record; identity must not duplicate it into
another provider prefix.

handlers is the sorted installed ceiling, each object having exactly name,
description and schema with its original semantic JSON value. It contains every
installed selectable handler, including management tools when skills are enabled.
Require unique names and exact canonical definition equality on resume. In plain
mode this also fixes the visible handler set; the application cannot silently
replace it while reopening. Existing fresh constructors retain their earlier
configuration behavior. Handler implementation code remains the application's
responsibility: a matching schema is not a code signature or sandbox guarantee.

Catalog identity covers every frozen definition, including inactive and currently
unreachable definitions. Form one object per definition containing name,
description, type, sorted tools, sorted depends, sorted loadable-skills and exact
body text; hash the canonical array of those objects, sorted by name, as
catalog_sha256. Empty sets are explicit empty arrays. List syntax
and frontmatter quoting choices disappear after parsing. Body bytes do not:
changing a final newline changes identity. Bindings_sha256 hashes the canonical
object of all supplied name/value pairs. Empty bindings hash an empty object.
Primary selection is compared separately. Do not hash directory names, mtimes,
absolute paths or the order files happened to be discovered.

| New caller selection | Live resume | Historical inspect/render |
|---|---|---|
| Same semantic catalog/bindings/handlers at another path | Allowed after ordinary validation | Uses recorded facts |
| Equivalent frontmatter spelling, exact same body bytes | Allowed | Uses recorded facts |
| Changed body, primary, graph, offer, binding or handler definition | Refuse session_incompatible | Remains available from recorded material |
| Missing catalog or mismatched skill-mode choice | Refuse session_incompatible | Does not require files or skill environment |
| Plain mode, System omitted | Adopt recorded base bytes | Use recorded base |
| Plain mode, explicit System exactly equal | Allowed | Allowed |
| Plain mode, explicitly different or empty System | Refuse session_incompatible | Refuse competing render base |
| Skill mode, competing nonempty System | Refuse under Chapter 9 | Retain Chapter 9 recorded-primary rule |

Presence matters. A shared CLI configuration reader must not fill the ordinary
default before it knows whether a session already supplies a plain base. For a
new plain session, omission selects the preceding nonempty default. For a resume,
omission adopts recorded bytes. Public options must distinguish omission from an
explicit request, rather than guessing from a zero-filled Config. A mounted
session rejects a changed base System before configuration mutation. Deliberate
enduring instruction appends remain supplements under Chapter 2; they do not
replace that base identity.

Install the recorded skill activations and exact material bytes. Live compatibility
validation also checks each activation's grants, dependencies and offers against
the caller's matching frozen definitions. Using its ordered transition provenance,
verify the rendered bytes against the original candidate and copied scalar
bindings; never replace a mismatch with newly generated text. This validation
does not reload files or invoke callbacks. Offline inspection/rendering continues
to validate recorded facts without catalogs or expansion. Enforce Chapter 9's active closure, grant union,
mandatory management pair, fixed ceiling, retirement and never-reused activation
rules. Comparing hashes alone does not validate those facts. A valid-looking
snapshot cannot load a hidden root, replace primary bytes or acquire a new grant.
The public live append boundary retains candidate validation against the frozen
current catalog. Offline projection confers no executable authority.

Model, vendor, endpoint, credentials and transport client are current application
inputs. Do not serialize Config wholesale. Session files contain the deliberately
listed identity plus historical producing provenance, never credential fields,
authorization headers or base URLs copied from transport configuration. Persisted
text is still user/tool/model content and may itself be sensitive; the library
cannot promise to redact a secret someone put in a prompt. Test secret canaries
in configuration without placing them in conversation content.

Keep same-target opaque-material restrictions. Selecting a new model is allowed
as configuration, but a historical request containing incompatible opaque parts
still refuses projection rather than dropping or sending them to another target.
Restore each producing model's usage once; current model selection does not
reprice earlier totals or charge replay as a new response.

Chapter 8's current policy file and its revision remain Agent policy's authority.
Historical turn.policy explains past decisions and never overwrites today's file.
GUI preferences remain Server-owned and independently persisted. The session
checkpoint contains no competing authoritative copy of either domain. Their
startup validation still applies; failure leaves session files unchanged. No
cross-domain “save everything” acknowledgement claims an atomic commit.

## 10.8 Refuse bounded, diagnosable failures

Loading an ordinary file should not ask the program to allocate whatever size
a number inside it names. Apply byte limits while reading, before decoding into
large values, and validate collections before reserving their declared size.
When the reader reaches a storage bound, the diagnostic must identify the bound
and preserve the last usable files. An unexplained fresh conversation is not a
reasonable substitute for that decision.

| Boundary | Limit |
|---|---|
| events.log file | 1 GiB, including header and line endings |
| One encoded event record | 64 MiB, excluding its LF |
| checkpoint.json or origin.json file | 512 MiB each, including whitespace |
| Canonical semantic state | 256 MiB |
| Event count; total entries; total parts; total recorded activations; seen request/call IDs | 1,000,000 each |
| JSON container nesting | 128 levels |
| Installed handler definitions | 1,024, with at most 16 MiB canonical total |
| Event/request/activation/job watermarks and durable identity integers | 0 through 9,007,199,254,740,991; required identities remain positive |

Here MiB and GiB mean powers of 1024. Exact boundaries are accepted when all
other rules hold. The outer object is nesting level 1; each contained object or
array adds one. Retain earlier smaller limits: human input, WebSocket messages,
policy files, skill source/catalog sizes and per-transition rendered material
are unchanged. A session limit is a storage admission bound, not permission to
enlarge an earlier tool or client message. Chapter 4 numeric tool-limit values
retain their existing type/range; they are not identity counters.

Check newly produced encoded record size before append. Exceeding it or the log
file/count bound returns a storage-limit error under the terminal persistence
rules before that event's effect is admitted. If a tool already executed before
a later result exceeded the limit, report the durable failure honestly; do not
claim the effect rolled back. Exceeding snapshot limits refuses checkpoint/export
without damaging the prior checkpoint or healthy log. No automatic compaction
or truncation conceals the limit. Diagnose early enough for an operator to choose
a new session rather than promise unbounded storage.

Strict outer/identity/state parsing rejects duplicate or unknown fields, invalid
UTF-8, trailing JSON, wrong types and unsupported versions. Existing event metadata
compatibility remains Chapter 2's; a new outer parser must not reject an otherwise
allowed harmless annotation on an old event. Detect duplicate event members and
invalid known payloads without echoing their contents. Validate integer ranges
before conversion or allocation. A JSON object can be syntactically valid and
still name an impossible call, anchor, primary or transition.

Use stable public error codes: session_in_use, session_unsupported,
session_conflict, session_corrupt, session_incompatible, session_unfinished,
session_origin_required, session_busy and session_io. A size/collection failure
uses session_corrupt when reading existing state and session_limit when admitting
new storage. Return safe field/record/sequence context, without dumping invalid
JSON, bodies or credentials. CLI errors may name the operator-selected path;
browser errors and snapshots expose no absolute paths.

Construction validation finishes before exposing an Agent, opening tool admission
or sending HTTP. On refusal, no durable source file is rewritten and no fresh
conversation is saved over it. Opening the lock file is not a history mutation.
Close failed construction resources and release reservations. Distinguish a
startup failure from a failure after an initial durable append: existing terminal
write semantics remain, and partial created artifacts stay diagnosable.

## 10.9 Take it through the interfaces

A public consumer must be able to open/resume, inspect, capture/export, import
and explicitly checkpoint using the same library the binaries use. Return owned
snapshots and typed acknowledgements; changing one caller's returned buffers
cannot mutate the live Agent or another consumer. An exported checkpoint includes
its anchor and exact versioned bytes, so an independent importer need not call
private implementation functions. Historical inspection never registers a live
Agent with tools.

Add `/session` and `/checkpoint` to human chat. The read-only `/session` reports the
SessionID, whether this mount resumed, current durable sequence and last committed
checkpoint sequence, or says that this is a standalone fresh-log Agent.
`/checkpoint` saves at the permitted boundary and displays the acknowledged anchor
or useful refusal. It makes no model request. A standalone Agent returns
session_conflict rather than silently choosing a directory. Machine protocol
output retains its previous defaults; a selected session does not inject banners
or save acknowledgements into that stream.

Add `session` to the safe public/watch state: null for standalone Agents, otherwise
an object with exactly id, resumed and checkpoint_seq. id is SessionID, resumed
is true for any successful existing-store mount and false for new creation or
explicit import's first mount, and checkpoint_seq is null until a checkpoint
commit exists. Existing log_seq still reports the current durable sequence.
State changes use the same atomic watch boundary and revision stream.
On WebSocket, put this object at snapshot_begin.state.session. Session and limit
bookkeeping facts do not become renderable cards or consume slots in the existing
100-event renderable window.

The browser labels the conversation with its session identity and provides a
Checkpoint control. Its subscribed command is exactly:

```json
{"type":"checkpoint","id":"save-1"}
```

It follows Chapter 7's command identity, strict validation and subscribe rules.
On success send a command_ack with that id, status `saved`, as_of and the applied
watch revision; refusal uses command_error and a code from §10.8. After checkpoint
commit, publish session_changed with the next watch revision, agent_id and the
complete safe session object before acknowledging. A second in-flight save gets
busy, and a disconnected client cannot cancel or roll back a committed write.
The browser updates the displayed anchor only from acknowledged/applied facts.

Expose historical job availability separately from its durable status. At the
watch boundary, include `job_access`, a sorted array of objects with handle and
live Boolean for historical jobs represented in the current event window. A
currently owned Job has live true; a restored historical job has live false,
including one last recorded done or running. Artifact displays the recorded
status and “historical; no live owner” when false, disabling any supervision
control. Live tool admission still enforces ownership if a caller forges a command.
This is presentation metadata; never append job_killed to tidy the screen.
Its initial wire location is snapshot_begin.state.job_access. Newly admitted job
facts belong to this live mount; later snapshot metadata still derives ownership
from Jobs rather than from the recorded status string.

On restart the page receives new runtime Agent identity and watch generation,
but retained event/part/call/activation coordinates. It replaces its old runtime
view with the new snapshot, including the existing omitted-event count. It does
not claim to display the complete history, restore old provisional fragments,
resend prompts, restart speech or reuse old input registrations. Actual store
resume and WebSocket reconnect are different operations.

### Taking it for a spin: demonstration to be filled from receipts

No Chapter 10 implementation or live transcript exists yet. The following is a
reproduction plan, not output from an actual session. Bind the validated source,
executable, store format and catalog first, and use provider discovery rather
than a remembered model ID. Credentials remain in the inherited memory/environment
path, never in argv, saved configuration or the transcript.

In a fresh scratch workspace, build the binaries to explicit absolute paths.
Run human chat with `--session-dir` pointing inside that workspace. Ask the Agent
to remember a unique ordinary marker and make one checked scratch-file edit.
Observe its answer and the file. Use /session and /checkpoint, then /quit. Start
a second actual process with the same selection and ask about the recorded work
without repeating the edit. Retain both PTY transcripts, event/checkpoint hashes,
request bodies, actual file bytes and usage. The second process must know the
conversation without reexecuting its tools.

Repeat on Messages, Chat Completions and generateContent. In skill mode, load a
narrow capability, leave a pending next-call setting through a controlled public
fixture where necessary, close and resume with the same semantic catalog at a
new path. Observe retained manual/grants and the literal next attempted call's
consumption. Preserve a model's refusal or unexpected extra call as an observed
outcome; use deterministic call fixtures to establish exact failure semantics.

Use the optional GUI on the same store after the CLI exits. Capture the actual
interface with accessible descriptions of the resumed session, prior cards,
checkpoint acknowledgement and current settings. Change today's policy and verify
that an ensuing real turn follows it rather than an old turn.policy value.
Show independent public Agents in separate directories, one resumed and one fresh,
without private imports. A busy checkpoint and a second writer must refuse while
the original Agent remains usable.

Malformed files, write/sync/replace faults, killed processes, missing origins and
stale-handle attacks require deterministic controls alongside the live use. Take
copies for destructive corruption tests; retain the untouched original hashes.
A clean real-model snapshot-only export/import must also resume through the public
consumer, with no invented pre-origin history. The final prose will report which
features were actually exercised, exact source identities, failures, limitations
and bounded usage. It will not present this plan as a completed spin.

## 10.10 Check the promises independently

The old seven-check persistence grader has a useful core: it compares actual
requests and creates a nonempty tail rather than trusting a MATCH message from
the program being graded. Retain those controls and the earlier chapter coverage.
Its old startup, fake Messages backend and EOF assumptions are insufficient for
this contract. Publish the independent command before releasing the student.

| Property | Required independent control |
|---|---|
| Construction compatibility | Fresh NewAgent and offline Load remain unchanged; default human/GUI resume and every selector-table row behave distinctly |
| Ownership | Headless public consumer, optional GUI module, actual SessionStore parent/logger, no second writer or saved Context authority |
| Identity | Fresh runtime IDs with stable SessionID; duplicate mount/path refusal; changed session/anchor/hash/watermarks reject; no forged public adoption |
| Snapshot | Public export/import, owned buffers, snapshot without prefix, later checkpoint/reopen retaining origin; missing/duplicate/inconsistent semantic fields refuse |
| Tail/rebuild | Old checkpoint/newer real tail, complete-history rebuild, imported-origin replay; all three exact next requests, usage, skill state and watch facts agree |
| Authority | Same-content different-path catalog passes; changed/missing source, primary, binding or ceiling refuses; plain-System presence table; no credential canary on disk |
| One-shot limits | Set/consume/restart, second and invalid setters, unknown/disabled call, skill control, pause, malformed response, append failure; complete-log reconstruction agrees |
| No resumed work | Unfinished boundary refuses; zero startup HTTP/tools; historical running/finished handles cannot supervise a new Job; transient queues/pauses/speech absent |
| Capture/write | Immediate busy refusal; independent job event during slow write belongs to tail; single worker; write/sync/close/replace faults preserve truthful commit state |
| Lifecycle | Actual EOF/quit/SIGINT/SIGTERM; GUI terminal detachment; join/repeated close; process death releases lock without file deletion or child-held descriptor |
| Limits/corruption | Exact and one-over boundaries; invalid UTF-8/JSON/version/duplicates/identities/transitions; partial final record; unchanged source bytes and zero side effects |
| Clients | Public multi-Agent isolation, browser checkpoint and historical-job display, bounded reconnect window, independent current policy/preferences |

Run all affected modules' build, vet and tests, including suitable race checks.
Keep legacy baseline failures separate. Use passing controls and deliberate
mutations for each material promise; a wrong hash rejected only because its test
path was invalid proves nothing about hashing. External checks inspect the
student's published codec and public exports instead of imposing undocumented
private field names.

Freeze the initial source, actual interface runs and student teaching review
before first-edition comparative feedback. The reviewer examines design, code
clarity, comments, cost and useful extension seams as well as behavior. Reconcile
findings into the teaching, repeat affected evidence, and leave failed/pending
attempts in the record. The independent proofreading and validation records,
not a successful save message, determine when this chapter is ready.
