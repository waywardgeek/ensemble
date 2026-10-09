# Chapter 10 outline: Remembering without restarting the work

Current reconciliation, October 8 PDT / October 9 UTC, 2026. New Chapter 10 maps
to first-edition Chapter 11. The initial actual-use experience is frozen at
`44d7627`, with runtime `57d4aac` and support `9822b2b`. The [chapter](chapter-10.md)
now describes those observations; independent review and later quality revisions
remain open in the [gate record](chapter-10-validation.md).

The original preparation sections below retain their proposed decisions and
then-pending predecessor chronology. They are historical research; the complete
published contract and current gate record govern the student, not those old
open questions.

## Stake and story preservation

The reader returns to yesterday's work and needs the conversation back without
accidentally repeating yesterday's command. Resolution belongs in an actual
second-process CLI/browser session that retains the dialogue, refuses corrupt
state, and clearly separates remembered jobs from controllable processes.

Keep old Chapter 11's default-load incident: Bill asked, “Let's load by default
without a flag.” Preserve the consequence that makes it matter: starting fresh
after an unreadable save can overwrite the very conversation the reader came
back for. Keep the snapshot/anchor explanation and the spliced older-snapshot,
newer-log example. The old goldfish opener now needs narrowing because the new
edition already has durable event logs and offline rendering; it lacks an
explicit, safe live-resume boundary rather than all memory of prior work.

Bring forward two documented later failures, with their source limits clear:
old Chapter 20's SIGINT path never reached the save code, and replay omitted
part identity so the browser lost assistant cards despite recorded events.
Use the first to teach one coordinated shutdown path and the second to connect
saved semantic identity to the existing public watch/Artifact contracts. Do not
reproduce the old hint/prompt classification design: the new actor contract
already supplies explicit requests. No invented duration, user dialogue or
successful run accompanies these historical incidents.

Voice plan: begin with returning to work, explain the unsafe retry it could
cause, then give the snapshot boundary and owner. Keep exact bytes and failure
cases near the decisions. Avoid a generic claim that persistence saves the
whole process; the user needs to know what actually survives.

## Proposed teaching order

1. Durable facts can resume a conversation; they cannot resurrect a worker.
2. Agent owns session identity and persistence location, separately from current
   model credentials/route and Chapter 8 policy/preferences files.
3. Define a versioned semantic snapshot, anchor and log relationship.
4. Capture through the actor; write an owned checkpoint through one SessionStore.
5. Validate the complete candidate before exposing a resumed Agent or writing.
6. Restore durable counters, usage and instruction/skill facts without double
   counting or replaying effects; reject an unfinished live-resume boundary.
7. Admit a new process with compatible caller authority, fresh clients and no
   restored jobs, queues, pauses or provisional speech/model work.
8. Exercise default restart through CLI, browser and public multi-Agent use;
   compare snapshot/tail/full-rebuild requests independently and inject faults.

Retain the accepted Chapter 9 primary/skill entry contract, Chapter 8 independent
settings authorities, Chapter 7 optional browser module, and the single actor
append path. No compactor, log-retention policy, automatic retry, distributed
writer protocol or crash-work recovery state machine is added incidentally.

## Ownership direction accepted for outlining

Root accepts this direction for preparation; full contract review still follows.

| Owner | Facts and work | Parent/lifetime |
|---|---|---|
| Ensemble | Application logger, live Agent/session identity and path reservations | Root; rejects two live owners of the same session store within the application |
| Agent | Authoritative creation configuration, session identity/DataDir, SessionStore, Actor and child services | Agent→Ensemble; identity/location are creation-only |
| Actor | Orders checkpoint capture and continuation with conversation mutations | Actor→Agent; exactly one durable event append path |
| SessionStore | Resolved persistence file identities, validation/loading, one checkpoint writer and atomic replacement | SessionStore→Agent→Ensemble; storage helpers retain this owner for diagnostics |
| Engine | Producing-model usage and current request transport state | Engine→Agent; restore accounting facts once, never transport workers |
| Skills | Frozen parsed catalog and committed activation/material state under Chapter 9 | Skills→Agent; resume compares caller configuration, restores recorded facts through the actor |
| Jobs | Fresh live-job ownership and any explicitly retained Agent-level next-call setting | Jobs→Agent; no saved process or stale handle becomes a live Job |
| GUI Server/client | Existing public presentation and preferences, fresh connections/Pages | Optional module; no parsing private core snapshots or writing Agent files |

Shared snapshot/configuration data and interfaces stay in core common; validation,
loading and file behavior belong in the persistence spoke. Reducer behavior stays
with its responsible llm implementation, invoked through common owner interfaces.
Do not move reducer/Skills/Jobs behavior into common merely to serialize a struct.
No SessionStore→Engine shortcut, sibling dependency bundle or closure containing
other services repairs missing reachability.

Root also accepts the first chapter's fail-closed direction for interrupted
work: an unfinished accepted turn/call boundary is inspectable offline but
refuses live continuation. Do not retry an effect, invent an interrupted result
or mark a job killed just to make a save loadable. The full draft must state an
actual recovery limitation and useful diagnostic instead of implying broad
crash recovery. Root will review exact boundary/error rules before drafting.

## Location, identity and one append path

Propose a creation-only DataDir containing a versioned checkpoint and the one
Agent event log. CLI and GUI should select the same default session location in
a workspace and load it automatically, preserving the historical usability
lesson. A public caller supplies distinct directories for multiple Agents.
Resolve paths once without changing process cwd; SessionStore owns their file
identities. Exact default/flags remain a coordinator decision below.

Reconcile Chapter 2's CH02_LOG/immutable LogPath before code. Persistence cannot
report a new destination while still appending to the original writer. A resumed
Agent intentionally opens its validated prior log for continuation; ordinary
configuration changes still cannot rotate it. An explicit conflicting path must
fail before mutation rather than silently creating another history authority.
Do not add an independent journal while leaving the existing append path alive.

Persist a stable session identity separately from vendor/model configuration.
An application cannot mount that same identity twice concurrently. Fresh runtime
connections, watch generations and operation workers remain transient. A new
session/fork needs a new identity and explicit user/application intent; copying
a save does not authorize two live writers to the original store.

Retain monotone durable sequence, request/call/activation and any other persisted
identity allocators needed to avoid reuse. A snapshot-only save needs its high
watermarks even when the corresponding events were omitted. Restore identities
through their actual owners, not by resetting an arbitrary process-global counter.
Job artifact handles retain Chapter 4's exclusive allocation/occupied-name skip;
existing `cr/io` contents cannot be truncated or repurposed on restart.

Cross-process writer protection needs a declared policy. Proposed chapter scope:
one application owns a store at a time, with nonblocking platform-supported file
locking for the live store; a second process refuses without writing. A stale
lock file must not permanently prevent restart after its owner dies. Do not
quietly claim in-process path reservations provide cross-process exclusion.

## Snapshot and tail proposal

Use a versioned file containing safe recorded configuration/provenance, stable
session identity, an anchor sequence, an optional semantic snapshot and recorded
history/tail metadata. The full draft must publish exact fields, supported
versions, byte bounds and anchor validation before checker work. Do not serialize
private runtime objects, mutexes, channels, sockets or function values as “state.”

A semantic snapshot must be complete for its promised restore mode: ordered
entries/parts and provenance; redaction/instruction/ephemeral selection state;
call pairing and durable outcomes; sequence/identity high-watermarks; per-model
usage; primary and all active/retired skill activation/material facts; and any
explicitly retained pending-next-call setting. It is owned data captured at one
actor boundary, not a collection of independently timed getter results.

Keep the historical anchor property. With a validated snapshot at sequence S,
apply only later events. Without a snapshot, a complete event history rebuilds
from the origin. Snapshot plus empty historical log remains an intended supported
case, provided the snapshot carries all required facts and high-watermarks;
do not claim full-rebuild equivalence for a missing prefix. A tail must belong
to the same identity and join the declared anchor without duplicate or conflicting
sequence contents. Existing imported gaps require an explicit rule; the new
writer itself remains contiguous.

For a complete-history save, snapshot-plus-tail and full replay must produce
identical next requests under fixed current render inputs, plus the same public
state and usage. Compare all three adapters, including skill material and mixed
parts. Equality of one plain-text prompt is too weak to validate the whole
snapshot. An independently produced fixture must have a real nonempty tail.

Retain Chapter 2's refusal of a partially written final log record; this chapter
does not silently discard the last line to manufacture a resumable boundary.
Propose strict semantic validation rather than old Chapter 15's skip-and-continue
policy for malformed parsed events. A syntax-valid but impossible skill transition,
unknown version, duplicate field, invalid identity, impossible anchor or mismatched
snapshot is corruption, not permission to start fresh. Preserve original bytes
and expose a safe diagnostic. Offline inspection should explain the problem
without silently converting damaged history into a new live session.

Full-history snapshot consistency needs a deliberate cost policy: recommend
validating the snapshot against reduction of its available prefix before use.
That is extra replay work; do not sell this implementation as a startup speedup.
A snapshot-only import cannot make the same comparison and must instead satisfy
its complete structural/semantic invariants and resume compatibility checks.
No unsigned hash proves authenticity of an operator-editable save.

## Capture, save and shutdown

Expose public explicit checkpoint/save and inspect/load operations without
importing the GUI. Proposed checkpoint capture is available only at a quiescent
actor boundary, with a clear busy refusal while a turn/batch is unsettled.
Whether historical running jobs count as quiescent is addressed separately
below; their workers are never part of the saved executable state.

Capture owned values under the responsible actor/owner synchronization, then
perform disk I/O in a SessionStore-owned worker. Only one checkpoint write may
be active. Controls and observation cannot wait on a lock held across disk I/O.
An acknowledgement names the persisted anchor; later appended events remain in
the authoritative log and are not falsely claimed inside that checkpoint.

Write a new private temporary file beside the destination, check full writes,
sync/close and atomically replace. Preserve the previous valid file on failures
before replacement and remove uncommitted temporary files. A failure after the
replacement commit cannot be reported as a rollback. Keep the append log intact;
no trimming or backup-rotation policy is required to introduce safe snapshots.
Explicitly distinguish process-crash recovery of complete durable records from
power-loss guarantees and from recovery of unfinished work.

Normal application shutdown refuses new admissions, follows Chapter 5's active
and queued completion rules, stops/joins its actual workers and process groups,
then captures the settled state and finishes its write before returning. A save
failure is visible and causes unsuccessful application exit; it cannot overwrite
a corrupt file that startup already refused. Route SIGINT/SIGTERM through that
same owned shutdown path instead of calling os.Exit from a signal callback.
A second forced signal/SIGKILL is outside orderly-shutdown completion.

Combined GUI --terminal EOF still detaches the terminal without closing the
Agent/server. It is not an application save-and-exit trigger in that mode.
Standalone CLI EOF and /quit retain their application-close meanings. Browser
view close releases only its client. Test these actual surfaces rather than
reusing an old harness that treated every stdin close as process exit.

## What a new process may restore

After full candidate validation, install the resumed conversation through a
construction-only entry point before exposing the Agent. Public event append
cannot turn a no-skills live Agent into a skill-enabled one or adopt another
session identity. Ordinary active configuration rules remain unchanged.

A complete idle boundary can include a historical job last observed running.
Root requires this distinction explicitly: saved status records are history,
never live OS process ownership. Do not recreate handles, readers, report cursors,
input channels or process IDs from a save. Waiting, sending input or killing
such a stale handle refuses as unavailable; it cannot target a coincidentally
reused process or another Agent. The browser should display “last recorded
running; no live owner” or equivalent provenance, without inventing a terminal
job event. Existing artifact references remain historical locators; never fetch
or re-execute them automatically.

Accepted but unsettled requests/calls, partial model responses and an active turn
block live resume under the proposed first-chapter boundary. Offline rendering
retains its own pending-call incompatibility rules. No startup completion result
is invented for a request whose caller belonged to a dead process. Fresh CLI and
browser clients start with fresh registrations, no queued prompts, no typing or
speaking causes and no resumed HTTP. Historical accepted final cards retain their
identities; provisional fragments are not promoted to accepted content.

Preserve Engine usage from accepted response facts exactly once, keyed by
producing provenance. Current selected model/credentials do not reprice or erase
old totals. Current route is chosen by the caller; foreign opaque material keeps
Chapter 2's explicit incompatibility behavior. No generic promise says every
saved conversation can continue on every provider/model.

## Skills, settings and credentials

Root's working direction requires explicit compatibility of the new caller's
skill-mode choice, primary, immutable catalog contents, scalar bindings and
installed ceiling with the saved identity. Propose canonical identity inputs:
normalized definitions including exact body bytes and graph/grant fields,
primary ID, sorted binding name/value pairs and sorted installed tool definitions
or their declared stable schema identities. Path names alone are insufficient.
The full draft must distinguish raw-file sameness from semantic catalog equality.

Recommended policy: same-content catalog at a different location can resume;
changed primary/body/graph/binding or incompatible ceiling refuses live resume
before writes. Missing catalog can still support historical inspection/rendering
from recorded activations, but not future dynamic loads or silently reconstructed
live grants. Persist compatibility descriptors/digests, not secret credentials.
No remote catalog refresh or changed-source migration is implied. Retain inactive
definition identity too, because a future load depends on those frozen choices.

Restore recorded active/retired roots, contributors, body bytes and primary base
without rerendering variables, reloading manuals or duplicating the primary in
context. Construction compatibility is a necessary additional gate; it cannot
replace Chapter 9's transition validation and fixed-ceiling checks.

Chapter 8 policy files remain Agent policy's authority, with their current
revision and configured path supplied at construction. Old turn.policy captures
remain historical evidence; they must not overwrite today's policy file.
GUI preferences remain Server-owned and independently persisted. A session save
must not introduce a competing copy whose load overwrites either domain.

Current credentials, endpoints, headers and transport clients come only from the
new application's configuration. Persist a deliberately safe request-config
record for diagnosis, never entire Config. Saved model/system/tool fields do not
authorize credentials or bypass the caller's current tool ceiling. A historical
base system instruction and durable supplements need an explicit precedence rule
rather than “current config always wins” silently replacing recorded identity.

## Open decisions for coordinator review before full drafting

1. Exact default store layout and CLI/public path selector; how CH02_LOG and
   existing LogPath identity interact with explicit resume and imported snapshots.
   Proposed one DataDir and one log avoids double journals; it needs a concrete
   compatibility/error table before implementation.
2. Exact clean boundary/error policy, including whether checkpoint requests during
   active turns refuse immediately or wait for a bounded settled boundary. Root
   prefers refusing unfinished live resume without a recovery protocol; define
   complete-tail validation and retain Chapter 2's truncated-record refusal.
3. Snapshot-only import and full-history consistency policy. Recommended semantic
   snapshot validation plus prefix equivalence when the complete prefix exists;
   keep the cost visible and define authority restoration independent of hashes.
4. Pending tool_limits: recommend preserving the one-shot Agent setting in its
   Jobs-owned typed state, without reviving any Job. Silently dropping it would
   change the literal-next-attempt rule across restart. This needs an explicit
   recorded/snapshot reconstruction rule before being required.
5. Exact catalog compatibility identity and no-skills base-System precedence.
   Recommend exact semantic catalog/binding/ceiling equality for skill mode;
   changing current credentials/model remains permitted subject to provenance
   compatibility, while changing conversation identity requires an explicit fork.
6. Cross-process exclusion mechanism and bounded parser/resource limits. No claim
   of crash/power-loss durability or safe concurrent writers without a taught
   implementation contract and distinguishing fault tests.

## Acceptance and actual-use plan

| Promise | Distinguishing control |
|---|---|
| Default restart | Two actual processes share a chosen session store; second request includes prior dialogue without an explicit load command |
| Anchor | Old snapshot/newer same-session log has a real tail; each event applies once; snapshot-only and complete-log-only paths retain correct numbering |
| Consistency | Same fixed rendering inputs give equal requests on all three adapters, safe watch state, skill state and producing-model usage |
| Corruption | Invalid JSON and semantic corruption, duplicate fields, wrong version/identity/anchor and incompatible skills refuse before writes/HTTP; bytes unchanged |
| Atomic write | Fault write/sync/close/replace independently; previous checkpoint survives precommit failure; tail remains authoritative |
| Shutdown | Standalone EOF and actual SIGINT/SIGTERM use one close/save path; combined terminal EOF detaches and browser remains live |
| No resumed work | Unfinished turn/call refuses live resume; historical running-job handles cannot wait/input/kill another process; no startup HTTP or tool effect |
| Authority | Current route credentials win, safe config serialization omits secret canaries, changed catalog/ceiling cannot acquire restored grants |
| Identity | Session, event, request/call/activation IDs continue without collision; regenerated browser cards match historical identities |
| Isolation | Two Agents and stores coexist; duplicate live identity/path refused; failure in one does not mutate the other |
| Public/browser | External headless save/resume, optional GUI startup/reconnect and reusable components; policy/preferences remain separate |

Before paid use, freeze a complete feature/action matrix on Messages, Chat
Completions and supported generateContent. Drive actual human CLI PTYs across
process exits, then observe answers before follow-up; run the browser and a
public two-Agent embedding with independent stores. Have the Agent remember a
unique ordinary marker and perform one checked scratch edit, close cleanly,
restart and ask about the recorded work without repeating the effect. Show
loaded skill state and a still-available permitted tool under compatible catalog
identity. Exercise current policy/preferences after restart without restoring
stale client pause or speaking history.

Malformed snapshots, write faults, process crashes and forced stale-handle calls
belong in deterministic controls, with exact original-file hashes and zero-effect
assertions. A live model declining a tool request is not proof of dispatch denial.
Retain source/binary/store/catalog bindings, two-process transcripts, actual files,
request bodies, usage, real browser images and all failures. No successful spin
is written until those observations exist.

Historical `make grade-dir CH=11 DIR=solutions/edition-2/main` remains a diagnostic.
Publish an independent acceptance command after the full new contract; retain
legacy positives and add mutation-sensitive tests without trusting a student's
own MATCH message. Accepted Chapter 9, contract review and independent checks
remain required before the next fresh student handoff.

## Initial actual-use and story reconciliation

After freeze `44d7627`, the coordinator released comparative teaching finding Q3.
The current opener now names the missing safe live mount instead of saying that
the preceding Agent forgets everything. Its durable logs and offline rendering
already exist. The returning-next-morning stake and original-author default-load
exchange remain; this is a continuity correction, not removal of the human story.

Section 10.9 now follows the actual two-process marker exercise, with exact
Chat Completions status excerpts and source-bound evidence links. It then moves
through the real headed GUI, current policy/preferences, public two-Agent and
snapshot-only import paths, and retained Skills/one-shot settings. The manuscript
keeps the invalid skill-name proposals, mistaken report-size explanation and
Gemini empty-card observation. All 33 generations and 3 discoveries belong to the
initial actual-use source, not to a later quality correction.

Actual commands were driven by the coder. Native speech claims are limited to
measured restored-work/admission boundaries; no hearing or Bill participation is
claimed. The accessible screenshot description accompanies the actual retained
image. Detailed file/hash/replay ledgers are linked instead of filling the reader
path. Student feedback dispositions are recorded separately; independent live
review and the presentation question remain open.
