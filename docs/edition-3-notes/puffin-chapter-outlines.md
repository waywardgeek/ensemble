# Connections, Agenda, and Puffin: detailed design outline

Status: design outline. Bill authorized manuscript renumbering and the bounded
Gateway upgrade. Current order is Goals 24, Voice and Images 25, Gateway 26,
Connections 27, Agenda 28, Puffin 29, and Crossing Over 30. Chapters 27–29 are
explicit design-stage landing outlines, not completed construction assignments.
Historical endpoints and grader flags retain their original numbers. This does
not authorize deployment or resolve the open contracts below.

Sources: [current decisions](scheduling-and-puffin.md) and
[xyzzy inventory](puffin-xyzzy-inventory.md). Settled decisions are distinguished
from recommendations and contracts still requiring agreement. Earlier chapters
will be revised at their original construction points as dependencies become clear;
a fresh implementation should not have to reconstruct our repair history.

## Settled persistence and infrastructure decisions

- Gateway: one complete versioned JSON state file, rewritten for each durable
  update. **No event journal for now.** Its state includes durable routing and
  authorization facts, not merely editable startup options.
- Agenda: authoritative in-memory calendar/task/scheduling objects, recovered from
  a versioned JSON snapshot plus its own sequenced event journal. No SQLite layer
  or separate database mirror of each mutation method.
- Puffin: existing Ensemble Agent History, memory and persistence mechanisms.
  Do not create a second Puffin journal for facts those owners already hold.
- Static configuration, secret storage, recovery state and diagnostics have
  different purposes. Credentials are not embedded in ordinary snapshots/exports.
- Services have separate durable owners even when composed in one executable.
  Cross-owner operations use explicit identities/receipts and recovery; there is
  no implicit transaction across Gateway, Agenda, a provider and an Agent.
- Agenda is shared infrastructure, never hosted as a private subsystem of Puffin.
  Linking its assembly into the Gateway host is the current recommendation;
  a separate application remains an option. Final deployment composition is open.
- Continue using the loaded ensemble-coding skill for all implementation. Common
  declares state, vocabulary, relationships and callable operations; functional
  partials implement behavior. No copied service bags, duplicate authoritative
  registries, or ID-based reconstruction of live relationships to bypass ownership.

## Gateway foundation: update the design first, refine through vertical slices

Begin with a bounded design delta against the existing Gateway, not a new router
and not a speculative rewrite of all its clients. Current inspected source owns
identities/groups and per-identity queues, has separate mutable policy/queue
snapshots and periodic/orderly saving. Those are existing facts, not evidence of
the stronger complete-file acknowledgment contract proposed below.

### Establish before building Connections

1. **Owner map.** Keep Router's real identity/group/session/queue relationships.
   Define the authoritative Person directory in Gateway infrastructure and how
   verified external bindings reference it. Connection ownership must be explicit;
   decide whether its durable definitions/grants belong in Gateway's state or in a
   distinct connection owner before adding another file. Runtime sockets, provider
   clients and secrets are not part of the durable object graph.
2. **Caller and represented person.** Keep authenticated Agent/service identity,
   human Person ID and transport session distinct. An actor can act for a person
   only through an explicit grant. Never treat a caller-supplied person or Agent
   ID as authentication. Preserve Bill's existing Noise choices; do not improvise
   cryptography or assume compatibility with the older Go service.
3. **Whole-file mutation boundary.** Serialize durable writers. Validate a complete
   proposed change before writing; encode a candidate state; write/flush a temporary
   file; replace the saved state; publish/acknowledge only after the required save.
   Readers must not see uncommitted state. Choose a simple validated change plan or
   staged candidate without maintaining a second long-lived mutable owner graph.
4. **What causes a save.** Durable identity/grant changes, queue admission, and
   receipt advancement/queue removal all count. Pings, sockets, transient session
   selection and display-only observations do not. A delivery accepted only into
   RAM must not be advertised as durably queued.
5. **Save failure and startup.** Preserve the prior valid state and a validated
   backup; reject silent empty-state recovery. Define behavior when a replacement
   succeeded but subsequent durability confirmation failed: stop ambiguous writes
   rather than pretend the old file is certainly authoritative. Decode/version/
   relationship validation precedes live publication and accepting connections.
6. **Durability promise.** Specify process-crash versus power-loss guarantees and
   the platform-specific flush/directory handling needed for the chosen promise.
   Rename alone is not proof of power-loss durability. Measure whole-file write
   volume and flash impact before introducing a journal as an optimization.
7. **Routing and admission receipts.** Define queued/delivered/accepted/completed
   distinctly. Establish enough of the Agent-host admission contract for verified
   incoming events and later Agenda work requests. Do not claim receipt of a wire
   message means an external call completed or a model turn was durably recorded.
8. **Composition and lifetime.** A host owns listener, component lifetimes and
   shutdown order. Components stop admitting work and join their owned operations
   before final resource disposal. No network sends under policy locks. Avoid an
   internal network hop when components deliberately share a trust boundary, but
   preserve the same authority checks for in-process calls.

### Gateway acceptance before dependent slices

Use real Router, storage and sessions with synthetic identities. Inject failure
before/after file write and replacement; test concurrent updates, failed revocation,
queue admission, receipt advancement, orderly shutdown and abrupt restart. A complete
saved state must not mix policy and queues from different generations. Verify
backup preservation and refusal of malformed complete state. Existing crypto,
authenticated replacement, attenuation and reconnect obligations remain in force.

Before moving the chapter, check all manuscript/code construction dependencies.
The current Goals text explicitly says it does not depend on Gateway; that supports
but does not complete this check. Endpoint and external grader numbers stay fixed.

### Then update as we go

Each dependent chapter takes one real end-to-end slice through this foundation.
If a missing contract is exposed, revise Gateway's source, focused tests and original
teaching before the dependent chapter claims completion. Reuse unaffected evidence;
rerun touched behavior and relevant shared regressions. Do not hide a required
foundation upgrade in a later appendix or freeze an unsuitable API for numbering.

## Construction sequence and teaching rules

Connections establishes **who may do what**. Agenda establishes **when work
is due and what happened to it**. Puffin demonstrates **useful, bounded assistance
built from those shared services**. Puffin is the running example throughout,
not a prerequisite secretly needed to pass the earlier chapters.

Integrate Linux command confinement into the original Sandbox chapter and needed
routing/lifecycle changes into the original Gateway chapter. Specify a minimal
Agent work-admission contract before scheduling uses it. Keep the dependency graph
acyclic; do not require the completed Puffin to construct Connections or Agenda.
An assembly is a component boundary, not necessarily a process or security boundary.

Each chapter starts with TL;DR and In Plain Words, teaches the coherent ownership
model and reasons, then provides a separately labeled construction specification.
Every requirement has one authoritative home. Proposed names below are conceptual,
not frozen class names, schemas, or public methods. No new grader point allocation
is invented here. Clarifying questions are legitimate engineering input; retain
useful actual Q&A after it occurs, and reconcile answers into the main specification.

# Chapter A: Connections

Working title: **Connections: People, Accounts, and Agent Authority**

## TL;DR / In Plain Words

A person authorizes an external account once. Named Agents receive explicit,
limited permission to use it. A requested tool operation and an incoming event
can share that connection without sharing an execution path. Connecting an account
does not give every Agent its powers, and receiving an event does not automatically
authorize a model turn.

## Teaching arc

1. **A message arrives: who is it from?** Start with a synthetic Discord message.
   Distinguish the human, external account, connection, Agent and transport session.
   A display name or claimed Person ID cannot establish authority.
2. **People before logins.** Stable unique human IDs; mutable contact details;
   verified external-identity mappings. One authoritative directory in Gateway
   infrastructure, separate from Agent identities. Represent family members who
   never use the GUI. Keep single-administrator GUI access explicit; do not imply
   guest isolation or build multi-user OAuth login as an incidental prerequisite.
3. **Authorize once; grant deliberately.** A connection identifies the external
   account and owns its credential lifecycle. Agent grants select capabilities.
   Separate reading, external effects, and event/wake permissions. Explain provider
   scopes versus local enforcement, revocation, expiry, and bounded delegation.
4. **Two initiation paths.** Agent-initiated MCP tools use existing tools and Jobs
   directly where appropriate. Verified outside events enter Gateway routing and
   then host admission. One Gmail connection may eventually support both paths.
   MCP notifications are not an alternate route around wake authorization.
5. **Build the Connections view.** A proposed tab in the Agents pane shows account,
   capabilities, assigned Agents, event routes, status and disable/disconnect/revoke
   actions. Never display secrets. Assign grants at Agent creation and edit them
   later. Explain what an unrestricted generic MCP tool really permits.
6. **Concrete connectors: Discord and Twilio.** Map incoming identities and events,
   verify provider-origin requests, and handle outbound messages/calls and delivery
   callbacks. Adapt to provider-specific semantics rather than inventing a universal
   promise of delivery. Keep content private in diagnostics.
7. **Lifecycle and failure.** Authorization changes during pending work; duplicate
   incoming events; reconnects; late callbacks; uncertain outbound results. Local
   cancellation cannot undo an accepted external effect. Stop and join owned work.

## Construction specification to develop

- Person identity and external binding contracts: creation, verified association,
  updates, conflict/refusal, deletion/retirement and references from resources.
- Connection owner, secret-store boundary, safe status projection, capability
  description, grant/delegation checks and revocation semantics.
- Direct MCP access controlled by the same explicit grant model without requiring
  a Gateway proxy for every call. Define whether shared credentials permit shared
  transport sessions; do not accidentally share one Agent's tool authority.
- Ingress event envelope: provenance, external event identity, destination,
  represented person where verified, and correlation. Host admission contract
  distinguishes delivery from running/completing Agent work.
- Provider-specific Discord/Twilio verification, deduplication and delivery-status
  contracts; exact APIs chosen from current documentation during implementation.
- Persistence and restart restore configuration, not stale live authority, consent,
  in-flight effects, or automatic retries. Define operational disable separately
  from deleting a record or revoking the provider authorization.

## First concrete data and interaction design

The vocabulary below is a design checklist, not an instruction to create one
framework/interface or one database table per noun.

- **Person:** immutable identity, mutable display information and verified external
  bindings. Retire/delete behavior must account for referenced calendars and records.
- **Connection:** stable local identity, connector kind, explicit external account,
  credential reference, enabled state and safe capability/status projection. Runtime
  health is an observation, not a durable claim that a socket still exists.
- **Agent grant:** real connection/Agent relationship, permitted operations, represented
  person/resource scope where needed, and delegation limit. Constructor/refusal and
  destruction update both sides. Retained records are not live grants until restored
  under current policy. Avoid a second generic permission language if explicit sets
  and resource references suffice.
- **Event route:** connection, verified source mapping, destination Agent or approved
  launch target, event filter and wake policy. A person mapping does not automatically
  entitle every sender on a channel to initiate work.

First UI flow: Bill selects a connector, authorizes one account, inspects actual
capabilities, assigns a restricted grant to an Agent, and separately enables an
incoming-event route. Cancelling setup must not leave a seemingly usable partial
connection. The UI shows provider authorization versus locally permitted operations
rather than conflating them. Agent-created assignments use the same checked path;
only a suitably authorized administrator/Agent may widen access.

First ingress flow: verify external source, resolve its binding, deduplicate by the
provider's event identity, check route/wake authority, admit through Gateway, then
receive the host's acceptance. Determine which owner durably holds deduplication
state and how it survives restart. A provider retry must not bypass the same checks.

First tool flow: the selected Agent invokes an allowed capability through its real
connection grant. The adapter rechecks authority at the effect boundary. A direct
MCP call remains a normal Job; observation timeout is not cancellation, and a
non-cooperative remote server may continue after local cancellation. Unloading a
connection/server and revoking a grant have distinct lifecycle effects to specify.

**Persistence decision needed here:** durable connection definitions and grants
must have exactly one owner. Prefer placing them within the Gateway-owned complete
state if that is where their authority lives; do not silently create an extra
connection journal. Credential material stays separately protected. The Connections
chapter owns these specifications; Agenda references them instead of redefining them.

## Proposed exercises and acceptance

- Connect a fixture MCP service, assign read-only capability to one Agent, and
  prove another Agent and an over-privileged child cannot use it. Test the actual
  dispatch boundary, not only hidden GUI controls.
- Deliver a synthetic signed/verified provider event through real connector,
  Gateway and Agent-host admission. Reject forged identity, unauthorized routing,
  duplicates and revoked access. One admitted message produces one intended input.
- Exercise the real Connections controls: assignment, safe status, errors and
  revocation. Verify no secret reaches browser frames, diagnostics or Agent context.
- Fake a provider accepting an outbound effect before the reply is lost. Preserve
  uncertainty without silently sending a second call/message.
- Take it for a spin with synthetic people and a clearly isolated connector.
  Separately authorize any live Discord/Twilio test and its recipients/cost limits.

## Non-goals / decisions before construction

No public signup, multi-user GUI login, blanket access to all human accounts,
second people registry, or scheduler hidden inside a connector. Settle connection
hosting, grant persistence and delegation, identity-binding verification/recovery,
callback reachability, and revocation behavior for already-admitted work. Gmail is
an architectural example; decide whether its first implementation belongs here or
in the Puffin slice rather than quietly expanding the exercise scope.

# Chapter B: Agenda

Working subtitle: **Calendars, Tasks, and Reliable Scheduling**

## TL;DR / In Plain Words

Store what people intend, determine when it is due, and record what execution
actually achieved. A schedule is not a timer; a socket write is not a completed
reminder; a task's due date is not permission to call someone. One authoritative
scheduler serves many people and Agents, including Puffin.

## Teaching arc

1. **Separate the nouns.** Calendar event, participant, task, due date, reminder,
   recurrence rule, occurrence, delivery attempt and external effect. Start with
   one appointment and one task before introducing repeating schedules.
2. **Who owns the details?** Resource ownership references stable Person IDs.
   Shared participation does not reveal every private detail. The scheduler checks
   authority even after Gateway authenticates the caller.
3. **Civil time is not elapsed time.** Monthly weekdays, time zones, DST gaps/folds,
   series edits and per-occurrence exceptions. Use iCalendar concepts as a reference;
   choose the supported subset/library explicitly rather than imply full compliance.
4. **One clock authority, durable intent.** Persist schedule definitions and derive
   bounded future occurrences with an explicit next wake. A finite search horizon
   cannot lead to sleeping forever. Reconcile existing idle-goal reminders without
   removing Agent-owned admission, pause handling and History.
5. **Execute and tell the truth.** Stable occurrence identity, schedule revisions,
   expiry, duplicate protection and separate dispatch/acceptance/completion/effect
   records. A lost acknowledgment produces uncertainty, not imaginary certainty.
6. **Change your mind.** Edit one occurrence, this-and-future, or the whole series.
   Cancellation must address queued/admitted work where possible; deleting a row
   cannot retract an already-placed call. Explain race outcomes and visible receipts.
7. **Recover after absence.** Restart, network loss, clock changes and missed work.
   Specify policies for retries, overlapping runs and expiry; a missed morning
   call must not silently turn into an unwanted midday call.
8. **Delivery and launch.** Address an existing Agent or request a permitted launch
   template with bounded authority/lifetime. Hosts own Agents. Compose scheduling
   with Gateway or deploy separately as chosen; neither requires Puffin-specific code.

## Construction specification to develop

- Resource and recurrence operations, durable save-before-acknowledgment boundaries,
  revisions, query projections and privacy enforcement.
- Supported civil-time/recurrence semantics, stable occurrence keys and exception
  identity across edits. Define machine clock versus timezone-policy changes.
- Scheduler owner, authoritative store, recovery scan, due-work admission and next
  wake; cancellation and joined shutdown. No competing recurrence engine in clients.
- Execution state/receipt contract across scheduler, Gateway, Agent host and
  connector; bounded retries and explicit unknown outcomes rather than an
  unsupported exactly-once guarantee for arbitrary external effects.
- Authorized launch requests, unavailable recipients, overlapping work, execution
  budgets, cancellation propagation and reconciliation on reconnect.
- Goal-reminder integration and migration: one timing authority, with existing
  Agent semantics preserved. Resolve this before declaring the scheduler complete.

## Agenda's authoritative objects and event model

Keep calendars, events, tasks, schedule rules, exceptions and execution records in
Agenda-owned memory, linked through real relationships. Use stable IDs at persistence
and message boundaries, not as a replacement for those live relationships. Derived
indexes and next-wake calculations can be rebuilt rather than persisted as competing
facts. A person/resource can have shared relationships; do not equate one authority
for a fact with a restriction that every object must have exactly one parent.

A recurring event defines intended civil-time occurrences. A task can have a due
date without any reminder. A schedule requests a particular action; an occurrence
identifies one instance; an execution attempt records what was tried and what is
known. Agree whether these need separate objects before encoding redundant state.

### One accepted-change path

1. Admit a command through Agenda's serialized mutation boundary.
2. Check the authenticated caller, resource grant and expected revision.
3. Resolve IDs, timestamps and command choices; validate the complete operation.
4. Construct a versioned event with the facts needed for deterministic replay.
5. Append it successfully under the chosen durable-acknowledgment policy.
6. Apply it through the same event-application path used by recovery.
7. Publish the updated view and acknowledge the command.

An event may atomically describe several related changes within Agenda. Do not
split one domain operation into independently committed fragments that expose
invalid state. If application of an already committed event unexpectedly fails,
stop admission and recover; do not acknowledge success or reuse its sequence.

Journal records contain sequence, format version, operation/request identity,
relevant provenance and the resolved domain change. Exact schema is still a design
deliverable. Replay must not generate new IDs, read the current clock, rerun current
authorization as if an old command were new, contact providers, or invoke a model.
Current policy governs new work after state reconstruction.

### Snapshot, replay, and recovery contract

- Snapshot contains the complete durable state at a stated journal sequence.
- Load independently validated primary or backup state and replay only its
  contiguous journal extension, once. Rebuild and validate relationships before
  accepting work. Corrupt complete records or sequence gaps refuse recovery.
- Specify torn-final-record handling explicitly. Inspect load without mutating
  evidence; any repair belongs at an explicit validated writer-activation boundary.
- Append failure with an uncertain file outcome poisons further writes; never
  reuse a sequence as though nothing could have reached disk.
- Publish a validated snapshot before trimming, and retain the suffix needed by
  the validated backup, not merely by the new primary. No usable backup means no
  unjustified removal of the only recovery history.
- Select and test flush/power-loss semantics for the Pi. Do not inherit an unproved
  durability guarantee from earlier Ensemble process-recovery tests.
- Recovery journal, long-term audit history, operational diagnostics and backups
  are distinct. Retention must account for private family data and flash growth.
  Private Git was discussed as a backup option, not selected as the sole backup or
  as a live multiwriter database. Credentials are excluded; consistent encrypted
  off-device backup, retention and restore remain decisions to finish.

### Effects after recovery

Replaying a due-occurrence event does not send another text or place another call.
Only the live execution path acts, after reconstruction and reconciliation of
attempt state. Preserve a stable operation identity across legitimate retries.
If a provider accepted work before a crash but no outcome was saved, record or
retain uncertainty and follow the agreed reconciliation policy. No local journal
makes a remote side effect atomic with its own append.

Use one initial end-to-end scenario: a verified human requests a reminder, Puffin
submits an authorized Agenda command, Agenda commits it, its clock later makes
one occurrence eligible, Gateway durably queues the request, the intended Agent
host admits work, and the connector returns an outcome. Each receipt states only
what its owner can establish. Each restart point has an explicit expected result.

## Proposed exercises and acceptance

- Controlled-clock tests for second-Tuesday recurrence, exception edits, DST,
  cancelled occurrences, missed intervals and horizon advancement. No real sleeps
  across days, no private family schedules.
- Real persistence/restart tests: inject failures around acknowledgment and each
  execution boundary; assert durable records and actual recipient effects.
- Disconnect/reconnect and duplicate delivery tests with stable occurrence identity;
  verify the intended duplicate policy at the real admission/effect boundary.
- Race cancellation with queued and already-admitted work. Report cancellation,
  completion or uncertainty according to the agreed contract, not timing luck.
- Two people/two Agents exercise ownership and visibility. Run one calendar/task
  service for both, rather than one per-Agent schedule store.
- Take it for a spin: create, inspect, revise and cancel synthetic schedules through
  the supported client surface, then restart the isolated host and inspect recovery.

## Non-goals / decisions before construction

No full calendar-product feature catalogue, distributed consensus cluster, or
implicit phone/email retries. Settle Gateway-linked versus independent hosting,
storage format/library, recurrence subset, missed/overlap/retry/expiry policies,
launch authorization, privacy model and goal integration. Keep the calendar and
execution sections together initially; split only if a coherent dependency and
independent acceptance boundary emerges during design.

# Chapter C: Puffin

Working title: **Puffin: A Family Agent Built on Shared Infrastructure**

## TL;DR / In Plain Words

Build an assistant that helps the family through explicitly granted connections
and Agenda, the shared calendar/task/scheduler. Preserve Puffin's useful identity and memories,
without making old conversations the source of truth for today's appointments.
This chapter proves composition, usefulness and recoverability, not another router,
scheduler, human directory or secret store.

## Teaching arc

1. **Define help in human terms.** Select a small first set of useful workflows:
   reminders, a morning briefing, and bounded email triage/drafting. Define who
   benefits, what Puffin may do, when she should ask, and how to stop a mistake.
2. **Bring Puffin's records across.** Inventory live/clone provenance, preserve
   originals, translate formats deliberately, and record exclusions. Identity and
   memory continuity are not a transfer of process state, credentials or authority.
3. **Give her the current world, not stale recollection.** Retrieve a fresh,
   replaceable schedule/task view with provenance and scope. Durable service data
   remains authoritative; acknowledged operations perform changes. Define snapshot
   freshness and failure behavior rather than letting the model invent availability.
4. **Useful powers within a boundary.** Linux sandboxed workspace on xyzzy; explicit
   filesystem/network/descendant restrictions for commands; connection grants for
   external services. Keep credentials out of model-visible files. Agent shell
   confinement does not automatically sandbox a trusted connector process.
5. **Workflows through the shared services.** A morning briefing combines authorized
   calendar/task/mail facts. A verified message is attributed to the right person.
   A proposed calendar change becomes real only after service acknowledgment.
   Reading/drafting mail is distinct from sending; approve initial effect policies.
6. **Operate on a small machine.** Lean ARM64 release artifacts, supervised services,
   health/readiness, bounded diagnostics, durable-state capacity, backups and tested
   restore. Reserve rollback space. Per-file caps do not bound total disk growth.
7. **Cross over without duplicate help.** Isolated trial, explicit live acceptance,
   migrate selected resources, retire the old timing authority, reconcile pending
   effects and keep a rollback plan. Rolling back binaries cannot undo a phone call.
8. **Make the next reconstruction better.** Record friction, clarified requirements
   and practical lessons back into the original chapters. A future self-improvement
   system is later work, not an unbounded autonomous upgrade loop added here.

## Construction specification to develop

- Puffin's host composition, identity/memory inputs, workspace and connection grants;
  distinguish durable memory, replaceable service context, and canonical History.
- Memory migration manifest with provenance, format checks, original preservation
  and rollback; no wholesale journal/config copying as an unexplained shortcut.
- Minimal workflow contracts, recipient/person attribution, authorization checks,
  schedule-view freshness, confirmations and operational error reporting.
- Linux command-confinement acceptance as a prerequisite, not a permissive fallback.
- Deployment/runtime/native dependency inventory for xyzzy, startup/shutdown order,
  service monitoring, storage thresholds, backup/restore and live cutover procedure.

## Bounded first useful Puffin

Recommended first slice: two synthetic family members, one shared event, one
private task, one authorized incoming message path and one reminder output. Prove
identity, privacy and the full execution chain before adding a rich briefing.
Then add a morning briefing from authorized schedule/task data, followed by email
read/triage/draft access if that connector is selected. Sending email and placing
calls require explicit effect policies; do not conflate useful reading with broad
account authority.

Puffin should request fresh service views at defined boundaries. Agree the
freshness marker, bounded size, unavailable/stale indication, and invalidation
behavior. New views replace ephemeral context rather than append endless obsolete
schedule copies. Persist only genuinely durable conversation/memory facts in her
existing owners; an authoritative appointment remains in Agenda. A natural-language
claim is not a substitute for a service acknowledgment.

For migration, identify the live memory source versus clones, preserve originals,
record hashes/provenance and translate explicitly supported formats. Stage into a
new destination and inspect the resulting identity/memory view before first use.
Do not overwrite a running Agent's newer History or revive saved jobs/tool calls.
Copying content must not imply restoration of old credentials, grants or schedules.

Deployment keeps the shared infrastructure independent of Puffin's runtime.
Specify exact ARM64 artifacts/native libraries, service accounts and file ownership,
secret access, startup readiness and shutdown order. Preserve the current tunnel
and unrelated Gateway clients. Restore tests must use a separate destination and
cannot accidentally register a second live sender for the same identity/events.
Define the switch of scheduling authority and rollback reconciliation before
stopping any old service. Backup success means a tested consistent restore, not
just an uploaded archive. No flash upgrade is currently justified by measurement,
but deployment requires fresh capacity and write-growth checks.

## Proposed exercises and acceptance

- Import a synthetic legacy memory corpus; compare preserved originals and expected
  new records, and prove excluded credentials/grants/jobs do not become active.
- Use two synthetic people with different visibility and connection grants. Execute
  an end-to-end briefing and reminder through real Gateway/scheduler/Agent paths
  with controlled model/provider boundaries. Check inputs, effects and provenance.
- Change/cancel a schedule after an old snapshot was shown to Puffin. Verify the
  authoritative service wins and stale context cannot silently resurrect work.
- Inject service unavailability, expired connection, disk-save failure, Agent restart
  and duplicate ingress. Verify truthful state and bounded recovery, not a demo-only
  happy path. Test backup restore in a separate destination.
- Run adversarial command-confinement checks on Linux ARM64: allowed workspace use,
  forbidden host access/network/environment leakage, child processes and cancellation.
- Bill and the implementer each perform the actual "Take it for a spin" using the
  supported controls. Live family calls/mail are separately approved, not grader
  setup effects. Automated geometry/fakes do not establish human usability.

## Non-goals / decisions before construction

No quiet cutover, automatic copying of live schedules into active replacements,
family-data publication, or permission expansion during migration. Choose the first
workflow slice, mail provider/permissions, who can approve effects, data retention,
backup destination and cutover criteria. Fresh host measurements are required before
deployment; the earlier9.8GiB free-space observation is not a permanent guarantee.

# Next design deliverable

Before prose or implementation expands, approve the bounded Gateway delta above and write the minimal cross-component
contracts using one end-to-end synthetic scenario: verified person sends an authorized request,
Puffin records a reminder, the shared scheduler later delivers it, and the connector
reports its outcome. Derive the smallest first-build plan and unresolved choices
from that flow. Refine Gateway as those slices expose concrete needs, updating its
original teaching and tests each time. Bill approves implementation and deployment separately. The final
chapter numbers and external grading allocations follow the settled construction
sequence, not the other way around.
