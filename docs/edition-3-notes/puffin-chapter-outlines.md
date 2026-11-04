# Three chapters toward the new Puffin: first outline

Status: first design outline, not an implementation specification or completed
chapter. Bill approved the three-part direction. Tentative manuscript positions
are27 Connections,28 Scheduling,29 Puffin, followed by30 Crossing Over. Do not
rename existing chapters, frozen endpoints, or grader flags yet. Actual chapter
boundaries can change as the contracts and executable exercises take shape.

Sources: [current design decisions](scheduling-and-puffin.md) and
[xyzzy inventory](puffin-xyzzy-inventory.md). This outline organizes those decisions;
it does not settle every open policy by turning a proposal into a requirement.

## Construction sequence and teaching rules

Connections establishes **who may do what**. Scheduling establishes **when work
is due and what happened to it**. Puffin demonstrates **useful, bounded assistance
built from those shared services**. Puffin is the running example throughout,
not a prerequisite secretly needed to pass the earlier chapters.

Integrate Linux command confinement into the original Sandbox chapter and needed
routing/lifecycle changes into the original Gateway chapter. Specify a minimal
Agent work-admission contract before scheduling uses it. Keep the dependency graph
acyclic; do not require the completed Puffin to construct Connections or Scheduling.
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

# Chapter B: Calendars, Tasks, and Reliable Scheduling

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
and reliable shared services. Preserve Puffin's useful identity and memories,
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

Before prose or implementation expands, write the minimal cross-component contracts
and one end-to-end synthetic scenario: verified person sends an authorized request,
Puffin records a reminder, the shared scheduler later delivers it, and the connector
reports its outcome. Derive the smallest first-build plan and unresolved choices
from that flow. Bill approves implementation and deployment separately. The final
chapter numbers and external grading allocations follow the settled construction
sequence, not the other way around.
