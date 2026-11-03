# Rebuilding Puffin: current design direction

This section records the latest discussion and supersedes conflicting priorities
in the historical investigation below. Puffin is the concrete proving application
for shared infrastructure, not a special-purpose owner of platform services.
Continue design first; decide the chapter boundaries afterward. "Puffin" is the
working subject, not an assigned chapter number. Existing Chapter27 remains
Crossing Over to Ensemble until an explicit manuscript reorganization.

## Settled direction

- Build a better Puffin in Ensemble and preserve her identity and useful memories.
  Stop the campaign to debug every defect in the old system. Retain established
  mechanisms as requirements/evidence, not as proven causes of historical calls.
- Bill built the first Puffin quickly to help his mother promptly. Preserve that
  useful behavior while taking time over the replacement's engineering.
- Exactly one shared scheduler/calendar/task system serves family work, many
  Agents, system work, and authorized Agent launches. It is separate from Puffin;
  Puffin is a client, not its owner or sole beneficiary.
- Gateway, scheduling, and Puffin will initially run on xyzzy, Bill's Raspberry
  Pi5 with32GB flash. See [read-only inventory](puffin-xyzzy-inventory.md).
- Connections are fundamental shared infrastructure. Discord and Twilio are
  existing integration targets; email and morning briefings are desired additions.
  A replacement must preserve relevant external behavior through tested adapters,
  not assume protocol compatibility from similar names or shared Noise primitives.
- Each represented human receives a stable unique Person ID, distinct from Agent
  IDs. IDs do not depend on display name, email, telephone number, or login.
  People may exist without ever logging into Ensemble.

## Components versus deployment

Keep explicit owners and interfaces for Gateway, scheduler/calendar/tasks,
connections, and Agent hosts. An assembly boundary does not require a separate
executable, nor provide security/crash isolation. C# assemblies can run on Debian
ARM64 with the appropriate .NET runtime and native dependencies.

Bill wants scheduling independent of Puffin, either linked into Gateway or in a
separate application. The current recommendation is a scheduler assembly composed
with Gateway in one supervised infrastructure host, while Puffin and other Agent
hosts are separate. **This deployment choice remains a proposal**, not a completed
selection or implementation. Shared hosting avoids an internal network/startup
boundary but shares crashes and upgrades. Separate hosting provides independent
availability at the cost of protocol, recovery, and deployment complexity. Do not
build both modes speculatively or collapse scheduler state into the router.

Gateway owns authenticated connections and routing. Scheduling owns authoritative
calendar/task and occurrence state, recurrence, exceptions and execution records.
Agent hosts own Agent lifetimes, work admission, execution and History. The host
composes and joins components; sending a launch request does not transfer ownership
of the launched Agent to the scheduler.

## Connections: two initiation paths, one authorization model

1. **Agent-initiated tools.** Reading mail, searching files, drafting, and querying
   calendars may use direct MCP clients and the existing tool/Job machinery.
   These calls need not pass through Gateway merely to make grants manageable.
2. **Outside events.** Incoming Discord messages, Twilio calls/texts, email changes,
   and other notifications that can initiate work enter through Gateway's verified
   ingress/routing path. The receiving host admits work or performs a separately
   authorized Agent launch. A protocol notification is not itself wake authority.

One connection can provide both interfaces. For example Gmail tools and new-mail
notifications should use one authorized connection, not duplicate credentials
or independent permission systems. The distinction is initiation and authority,
not whether the transport happens to be MCP.

A connection records the external account, available capabilities, status, and
explicit Agent grants. Read access, external effects (send/call/change), and
permission to wake an Agent are distinct. Provider authorization and enforced
adapter/tool restrictions determine the real limit; a GUI label or hidden button
cannot narrow an unrestricted generic API tool. Credentials stay with the
connection owner, outside model-visible files and memories where supported.
Revocation blocks new uses/deliveries but cannot undo already accepted effects.

A proposed **Connections tab in the Agents pane** makes these understandable:
which external account; what it can do; which Agents may use it; which events may
wake which Agent; health/expiry; and disable/disconnect/revoke controls with clear
semantics. This is a UI direction for design, not an implemented tab. At Agent
creation or later, Bill or an authorized Agent can assign specific connections
with capability subsets. Child creation does not implicitly copy every grant;
delegation cannot exceed the parent's authority and Bill's explicit policy.

## Multiple represented people, one GUI administrator for now

Ensemble's GUI remains a single-administrator trust boundary today. A trusted
friend operating that GUI acts with Bill's authority; do not imply isolated guest
access. Multiple represented family members are necessary for calendars/tasks,
but that does not require Google login, invitations, or a multi-user GUI now.
The earlier Google-first login proposal below is deferred, not current build scope.
A GUI login and an OAuth authorization to Gmail are separate credentials/purposes.

The recommended directory owner is Gateway infrastructure: one authoritative
human identity directory alongside, but distinct from, Agent identities.
Schedulers and connectors reference Person IDs rather than maintaining competing
people registries. Verified external identities map to Person IDs; caller-supplied
IDs are not proof of identity. Resource owners still enforce access rules.

Calendars/tasks have explicit ownership, participants and permitted access. Shared
event participation must not automatically disclose every private detail. An
Agent's authority to act for a person is explicit; knowing a Person ID grants no
powers. Exact ownership, delegation, mapping verification, and recovery APIs still
need design before implementation.

## Puffin's context, sandbox and migration

Bill values the family schedule as ephemeral Agent data. The working proposal is
an authoritative durable schedule service and a fresh replaceable context view
for Puffin, not appointments remembered as authoritative prose or discarded on
restart. The snapshot/invalidation mechanism still needs design. Task due dates
remain distinct from reminders and scheduled actions.

Bill authorized inspecting Puffin's identity/memory records, including private
family content, for migration planning. That is not permission to publish those
records. Preserve originals and determine live-versus-clone provenance. Imported
memories do not import credentials, grants, pending calls, or active schedules.
Sandboxed commands and explicitly granted connections should make Puffin more
capable than safe mode alone. Linux command confinement is currently missing in
Ensemble and must be implemented/tested fail-closed before granting run_command.
Bubblewrap's presence alone is not application acceptance. Do not use unrestricted
execution as an implicit fallback.

## Open design and acceptance work

- Final scheduler/Gateway hosting choice and explicit component shutdown ownership.
- Minimal Person, Connection, AgentGrant, Calendar, Task, Schedule and Occurrence
  contracts, with one authority for each fact rather than mirrored registries.
- Calendar recurrence including monthly weekdays, exceptions, time zones/DST,
  one/this-and-future/all-series changes; approved missed/overlap/retry policies.
- Durable occurrence identity, duplicate suppression, expiry, cancellation and
  distinct delivery/acceptance/completion/external-effect receipts. Preserve
  uncertainty; do not promise exactly-once phone calls from socket acknowledgments.
- Reconcile current goal reminders with the single-scheduler requirement.
- Verify connector identity mappings, permissions, inbound event routing and wake
  rules, outbound effects, and connection revocation without real family calls.
- Test Gateway compatibility and upgrade only where the new clients require it.
- Produce a migration/rollback plan with storage headroom, backups and no competing
  old/new scheduler for the same live occurrences. Stop old authority deliberately.
- Self-improvement and security-monitor workflows remain later work. Human approval
  gates upgrades; monitoring, pausing, cancelling Jobs and resuming have distinct
  authorities still to be settled. None is implicitly granted by this design.

Next deliverable: continue the design with concrete contracts and a first-build
slice, informed by Puffin's memory/capability inventory. Bill approves the plan
before implementation; deployment/restarts/live communications remain separate.
No code, live configuration or remote services changed while recording this update.

Lesson: multi-person ownership is needed now, multi-user GUI authentication is not.
Likewise, shared service ownership is necessary, separate processes are a deployment
choice, and connector protocol is distinct from permission to initiate Agent work.

---

# Historical investigation and earlier proposals

The following record preserves prior reasoning and evidence. The current direction
above supersedes its legacy-debugging priority, privacy-read restriction for the
now-authorized memory inventory, and immediate multi-user-login proposal.

# Scheduling, Puffin, and Gateway: initial source investigation

## Direction and authority

Bill selected exactly one scheduler for personal/family work, Puffin, framework
maintenance, and authorized launches of agents for different people and tasks.
Preserve the useful basic design. Bill describes Puffin as a weekend prototype
that now needs real engineering; its useful behavior is not a reason to replace
everything. Calendar-level recurrence includes second Tuesday of each month and
per-occurrence exceptions. Investigate first; Bill approves upgrade plans before
implementation. No runtime changes, live calls, schedule deletions, credentials,
or private calendar/conversation inspection in this investigation.

Bill directly supervises Cadence, the security monitor, and the self-improvement
agent. Potential security pause authority still needs explicit scope, receipts,
and resume rules; pausing an Agent does not necessarily cancel its running Jobs.
Self-improvement should gather friction, investigate, interview live/preserved
agents with provenance/permissions, and propose upgrades for Bill's approval.
These are planning decisions, not implemented capabilities or expanded grants.

## Planning discussion: decisions, reports, and open choices

### Bill's priorities and reported experience

Puffin is working well overall. Bill reported that expected scheduled events
sometimes fail to fire and that unwanted schedules need removal, including
7 a.m. calls to him and Friday calls to Laura. He has not yet established in
this discussion whether removal was requested and the schedules returned, or
whether they remain to be removed. Do not convert these reports into a proven
root cause or authorization to modify the live schedules.

Bill describes Puffin as largely vibe-coded over a weekend. He wants real
engineering behind it now, preserving its useful behavior and basic design.
This is hardening and a considered upgrade, not a mandate for a wholesale rewrite.

### Agreed direction

- Exactly one scheduler handles personal/family tasks, Puffin activities,
  framework work, and agent launches for different people at different times.
  Distinct clients or views must not become competing scheduling authorities.
- Scheduling must be calendar-level. Bill's concrete example is the second
  Tuesday of every month. Recurrence, exceptions, and controllable removal are
  part of the requirement, rather than interval timers alone.
- The scheduler must both drive existing agents such as Puffin and arrange
  authorized launches of agents for tasks. Existing Gateway infrastructure
  should route agentic messages as well as maintain connections.
- Scheduling is the proposed foundation for the next self-improvement work.
  This does not declare the current multimedia chapter complete, renumber book
  chapters, or authorize replacing preserved endpoints.
- Agents should be able to report friction, bugs, and improvement ideas. The
  self-improvement agent should investigate across Gateway, Ensemble, the
  scheduler, Puffin, other agents, and eventually security monitoring/WatchTower.
- Investigation should include interviews with live agents, explicitly Puffin
  and Cadence, and potentially restored or cloned agents when needed. Restoration,
  cloning, and review of saved records must remain distinguishable; none implies
  inherited permissions or recovery of an original agent's subjective experience.
- Bill directly supervises Cadence, the security monitoring agent, and the
  self-improvement agent. Upgrade plans require his approval before construction.
  Investigation, implementation, deployment, and restart are separate authorities.
- The proposed security monitor watches sub-agents for prompt injection and
  incorrect behavior. Bill is considering permission to pause other agents when
  needed; this is not a present grant of arbitrary interruption or cancellation.

### Recommendations discussed, not yet settled specifications

- Use the iCalendar recurrence model as a reference rather than inventing another
  calendar language. No particular library or complete standards-conformance
  obligation has been selected.
- Keep a task's due date distinct from its reminder or scheduled action. One
  scheduler need not become the owner of all task content or private calendars.
- Preserve person/purpose ownership and access controls within shared scheduling.
  A supervisor can investigate delivery failures without automatically receiving
  a child's homework details or another person's private records. The exact
  identity, delegation, and visibility model remains to be designed.
- Support explicit change scope: one occurrence, this and future occurrences,
  or an entire series. Define timezone/DST, missed-event, overlap, retry, and
  rescheduling rules; a missed morning call should not silently become a surprise
  midday call. These policies need approval, not assumptions embedded in code.
- Security pauses should carry reasons and evidence, notify Bill, distinguish
  suspicion from findings, and have explicit resume rules. Pause does not
  necessarily stop tools or Jobs already executing. Authority to interrupt,
  cancel, or resume remains an open decision; avoid automatic timed release of
  a safety hold or clearing an unrelated manual pause.
- A self-improvement loop should track report, investigation, proposed plan,
  Bill's approval, implementation, verification, and updated executable teaching.
  Reporting a problem does not itself authorize changes or expanded access.
- Reconcile existing goal-reminder timers with the one-scheduler decision while
  preserving Agent-owned admission, pause handling, and History. The separate
  scheduler executable recommended below is the working proposal; exact hosting,
  persistence, and migration details still need the upgrade plan.

Bill authorized the source investigation and asked that this discussion and its
findings be recorded. The next engineering deliverable is isolated reproductions
and a concrete upgrade proposal for approval, not an unapproved implementation.

## Follow-up discussion: executable split and family sign-in

### Proposed executable and ownership split

Bill asked whether the calendar/scheduler belongs in Gateway or should be a
separate Gateway client. Cadence recommended a **separate scheduler executable**,
connected through the existing Gateway. Exactly one scheduler means one
scheduling authority, not necessarily one combined executable. This is the
working architecture proposal for Bill's upgrade-plan review, not an implemented
split or permission to migrate services.

- **Gateway** owns authenticated connections, routing authorization, addressed
  messages, and delivery queues. It already has agentic routing to preserve.
- **Scheduler** keeps calendar definitions and timing together under one owner:
  recurrence, occurrences, exceptions, cancellation, missed-run policy, and
  execution tracking. It has one authoritative schedule store. Puffin and other
  clients must not maintain competing schedule copies or recurrence engines.
- **Agent hosts** own the agents they construct and fulfill authorized launch or
  work requests. An existing Agent retains ownership of its work and History.
  Requesting an agent launch does not transfer its lifecycle to Gateway.
- **Puffin** remains a useful family-facing interface and executor, using the
  common scheduler rather than independently deciding when the same work is due.

Reasons for separation: calendar complexity and scheduler upgrades should not
force routing restarts or disconnect unrelated conversations. It also avoids
making the router an owner of private calendar content. The tradeoff is another
process and a real service boundary that needs reliable receipts and recovery.
Initially the executables can share a computer under ordinary process supervision;
a distributed deployment is not a prerequisite.

The discussed occurrence flow was: record due occurrence; send an addressed
request carrying stable occurrence identity and expiry where appropriate; record
recipient acceptance separately from completion/external effect; retain an
explicit uncertain outcome if confirmation is lost. Retries must preserve
identity and recipients need appropriate duplicate protection. Cancellation must
reach queued or accepted work where still possible; deleting a calendar row is
insufficient. These are design requirements to refine and test, not a claim of
exactly-once phone calls. Migration must retire the old scheduling authority for
migrated occurrences rather than run both schedulers for them.

### Agreed family access direction

Bill is building help for his family, **not a public subscription service**.
He can provision or whitelist family accounts and prefers his Google login.
The agreed starting direction is **Google-first sign-in for approved family
members**, with Bill administering membership. No public signup, subscription
billing, or new local password database is needed for this scope.

Cadence proposed OpenID Connect for Google sign-in. The identity and login
component would initially live in the Gateway executable, separated logically
from routing; a third identity-service process is not needed for the first plan.
Gateway would own one authoritative user directory, account status, external-login
links, and browser sessions. Google would authenticate Google credentials;
the application would not receive users' Google passwords.

The proposed first-login flow is:

1. Bill creates an invitation or approves the family member.
2. The person signs in through Google.
3. Gateway validates that login and matches it to an approved account.
4. Gateway establishes the user's session; unapproved identities receive no access.

Use stable internal user IDs for ownership of calendars, schedules, and other
resources. Link a Google identity by its validated issuer and subject identifier,
not email alone. Verified email may help match an invitation on first login,
but later email changes must not transfer resource ownership. Other providers
can be supported later without duplicating the person or their data. Linking a
new login should require an authenticated account or Bill's explicit approval;
matching email strings alone must not automatically merge accounts.

### Authentication is distinct from resource authorization

Keep **people**, **agents/services**, and **sessions/connections** distinguishable.
Existing service-to-service authentication remains separate from browser login;
family members should not have to manage Agent keys to sign in. An authenticated
agent connection does not automatically confer a family member's authority.

Gateway establishes who is authenticated and enforces routing permissions.
The scheduler still owns calendar access rules and decides whether a person,
or an explicitly authorized agent acting for them, may perform an operation.
Puffin may own relevant user preferences, but not another account/password
registry. Central authentication does not eliminate resource-owner authorization.

Exact delegation, invitation matching, account linking/recovery, session lifetime
and revocation, administrator recovery, calendar visibility, and migration from
existing access controls remain design work. No Google application was registered,
no accounts or OAuth credentials were created, and no current authentication,
Gateway, scheduler, or live service configuration was changed by this discussion.
Bill asked that these notes be recorded; upgrades still require plan approval.

## Source map, not a deployment assertion

- `/Users/bill/projects/puffin/go.mod` replaces its Gateway module with
  `../gateway`. This identifies the local source dependency, not the binary or
  configuration currently running for Bill or Laura.
- Standalone legacy Gateway: `/Users/bill/projects/gateway`, especially
  `pkg/scheduler/scheduler.go`, `pkg/agent/manager.go`, `pkg/store/store.go`, and
  `pkg/router/router.go`.
- Puffin's active startup is `cmd/puffin/main.go`: data tools, `DataService`,
  `Notifier`, and delivery functions. `pkg/tools/schedule_tools.go` and the older
  schedule store also exist, but are not the tool set installed by that startup.
- Puffin recurrence: `pkg/data/occurrences.go`; notifier and delivery:
  `pkg/schedule/{notifier,delivery}.go`; Gateway client: `pkg/client/client.go`.
- Ensemble's existing C# Gateway: `agent/Gateway/`, with `Common/Protocol.cs`,
  `Router/Administration.cs`, `Client/Client.cs`, and `Adapter/Bridge/Bridge.cs`.
  CodeRhapsody also retains an embedded Gateway tree; it is not automatically
  authoritative for Puffin merely because its directory has the same name.

## Source-confirmed mechanisms and risks

1. **Two scheduling layers.** Puffin computes due actions in its notifier, then
   submits a Gateway one-shot scheduled job (one second later) for chat delivery.
   Gateway independently owns recurring jobs, timers, and persistence. Its cron
   calculation reads but does not use month and weekday fields. Puffin already
   has richer calendar recurrence, including monthly weekday ordinals and
   occurrence exceptions. Preserve and test that useful behavior.
2. **A success receipt stops too early.** `SendScheduleCreate` explicitly does
   not wait for `schedule_result`; it returns success after writing the frame.
   Incoming schedule results are logged. The notifier calls `markFired` after
   `fire`, including paths that log failure or suppress delivery. Thus a local
   fired record does not establish successful Gateway admission or a call.
3. **Future-window starvation is possible.** Recurring event lookup searches
   from reference minus 24 hours to reference plus 48 hours. If `nextAction`
   returns nil, the notifier waits for a data change or cancellation, rather
   than scheduling a future lookahead wake. An otherwise empty calendar whose
   next occurrence is farther away can therefore lack a timed wake. Reproduce
   with an isolated clock/fixture before claiming this explains a live incident.
4. **Cancellation spans different owners without one occurrence identity.**
   Puffin's active `calendar_remove` calls `DataService.RemoveEvent`. That method
   mutates the list, ignores `saveCollection`'s returned error, notifies change,
   and returns true. Gateway's `schedule_delete` removes a Gateway job, while
   already queued prompts are separately retained. An event deletion therefore
   must not be assumed to revoke already handed-off work. No live deletion or
   resurrection incident has yet been traced.
5. **Scheduled sub-agent routing is fragile.** Gateway labels scheduled prompts
   with source channel `scheduler`. Puffin resolves its local sub-agent by
   channel and otherwise falls back to the first entry encountered in a Go map.
   Whether a deployed mapping avoids that fallback is unverified. A person,
   recipient, conversation, and task destination need explicit routing semantics
   rather than accidental map order.
6. **Existing C# agentic routing should be retained.** Its `send-message`
   operation verifies the caller's capability and the recipient's accepted
   group, stamps authenticated provenance, and enqueues an addressed envelope.
   The client/bridge distinguish delivery acknowledgement, mailbox admission,
   completed inference, and external effects. The bridge explicitly documents
   that mailbox admission is not a durable application commit across crashes.
7. **Crypto name is not wire compatibility.** Legacy Gateway uses prologue
   `gateway-v1`; current C# uses `ensemble-gateway-v1:` plus the Agent identity,
   along with different frame/RPC/envelope contracts. No new interoperability
   probe was run. Do not repoint Puffin or change cryptographic primitives as an
   assumed migration shortcut.

These are source findings, not verified explanations of Bill's missed events,
7 a.m. calls, or Laura's Friday calls. No deployed schedule inventory was read.
The broad initial source searches and several combined reads were clipped;
relevant bounded spans were recovered. This is a focused initial investigation,
not a complete audit of any repository.

## Proposed next investigation and approval packet

- Preserve useful Puffin recurrence and per-occurrence changes with isolated
  behavior tests, including time zones, DST, missed occurrences, and horizon
  boundaries. Determine exact current semantics rather than label them iCalendar
  compliant without evidence.
- Reproduce nil-next-action sleep, failed schedule admission recorded as fired,
  deletion write failure, queued-work cancellation, reconnect replay, and
  ambiguous sub-agent selection using fake transports/clocks and no real calls.
- Map durable schedule, occurrence, delivery attempt, recipient admission, and
  external effect separately. Decide how stable occurrence identity, schedule
  revisions/cancellation, and idempotency cross the scheduler/Gateway/Agent
  boundary. Do not promise exactly-once external calls from socket receipts.
- One scheduler owns timing and occurrence state. Gateway owns authenticated
  connections and addressed message routing. Existing Agents own their work and
  History; a host owns any Agent it constructs in response to an authorized
  launch request. The scheduler need not become the Agent lifecycle owner merely
  because it requests the launch.
- Present Bill with explicit choices for scheduler hosting, identity/permissions,
  missed-run and overlap policy, family/system visibility, cancellation scope,
  Agent launch templates/lifetime/budgets, and the migration of existing schedules.
  Include Ensemble's current idle goal-reminder timers in that design so a second
  independent scheduling authority is not accidentally retained.
- Plan a migration/rollback with named ownership and reconciliation of old jobs
  and pending messages. Do not run both schedulers for the same live occurrences.
- Preserve the current accepted application behavior. Approval of this
  investigation is not approval to build, deploy, place calls, or replace a live
  Gateway. No implementation has begun.
