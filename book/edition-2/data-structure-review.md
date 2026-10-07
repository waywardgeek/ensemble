# Manual data-structure review

October 7, 2026. Requested by Bill while the Chapter 4 coder continues.
Coordinator review of source `8a4794d1b166b2e43c6ea47b69f1c23c4768c03f`.
This is a review record, not chapter acceptance. The coder owns implementation
changes; the independent reviewer checks their resolution. Later changes must
be checked against these findings rather than assumed to resolve them.

## Ownership in the inspected implementation

```text
Ensemble
  logger, Agent map, observer subscriptions, application handle allocator
  Agent
    configuration, authoritative events, derived Context, admission/lifecycle
    Engine
      HTTP client, usage by model
      Call
    Registry
      permitted tool entries, schemas and implementations
    Jobs
      live job map, pending one-call limits, lifecycle synchronization
      Job
        artifact, output cursor, process/terminal resources, current status
    EventLog
      writer and persistence-fault state

CLI and optional GUI client -> public Ensemble interface
```

Shared value declarations and parent interfaces are in `internal/common`.
Concrete Ensemble/Agent declarations are in the root package; concrete Engine,
Registry, Jobs service, Job and event log are in their implementation packages.
The optional GUI is a separate module.

Several copies of a fact are intentional projections. Context derives from
events; its JobSnapshot is a historical observation, not the live Job; Engine
usage derives from accepted responses. Those are different responsibilities.
Removing them merely because their fields overlap would damage replay or
accounting. The question is which copy may change, who changes it, and when
the projection becomes visible.

## Findings

### 1. Response mixes parser bookkeeping with accepted data

**Confirmed correctness defect; fix before Chapter 4 acceptance.**
`Response.MissingCallIDs` is transient parser metadata on the same public type
used for accepted responses. Agent's append path applies its indexes before
validating the response payload (`ensemble.go:393` at the reviewed revision).

Two isolated public-consumer probes against an exact committed-source export
confirmed:

```text
Append(Event{Type: "response_ended"})
  panic: missing response payload
Append(Event{Type: "response_ended", Response:
  &Response{MissingCallIDs: []int{100}}})
  panic: out-of-range index
```

Recovering the first panic and attempting another append deadlocked because
Agent's state mutex remained locked. These are local API calls; no provider
or credential was used.

Validate shape, index bounds, part kind and legitimate missing-ID conditions
before mutation. Rejected input must leave the log, state and accounting
unchanged and leave the Agent usable. Prefer distinguishing parser output and
its bookkeeping from the accepted response at the type boundary. The final
call ID must still use the actual durable response sequence, not a guessed
sequence from before HTTP.

### 2. Mutable configuration can disagree with the resource it describes

**Confirmed public identity defect; fix before Chapter 4 acceptance.**
`Config` combines construction choices (`Workspace`, `Builtins`, `LogPath`)
with mutable model settings. SetConfig protects the first two construction
choices but accepts a changed LogPath without changing the existing writer
(`ensemble.go:264`). An isolated public-consumer probe returned:

```text
SetConfig accepted changed LogPath: true
Config reports changed path: true
Original file received event: true
Changed file absent: true
```

Make the distinction explicit. Reject an attempt to change the existing log's
identity; define what omission means. Loading a log should likewise retain
truthful source identity. No implicit log-rotation feature is requested.
A separate construction-options type and mutable model-settings type may
make later APIs clearer, but the immediate invariant does not require a large
public API rewrite.

Registry declarations copied into request configuration are a different case:
the current code derives them from the Registry and checks conflicting input.
Keep that single authority rather than treating every derived copy as a bug.

### 3. Parent interfaces hide required capabilities

**Maintainability finding, independently corroborated.** Engine stores
`common.Agent`, then asserts `common.TurnAgent` to obtain its ordinary services.
Jobs reaches Ensemble and asserts `common.HandleOwner` for normal allocation.
These are required construction relationships, not optional features.

Express the required parent contract in the constructor and stored interface
so the compiler checks it. Keep the immediate-parent chain; do not solve the
problem by injecting Registry, Jobs and logger as separate sibling references.
Private admission adapters can remain where they keep internal append paths
off the public API. Their purpose and requirements should be clear.

Job also converts its parent interface back to the concrete Jobs service to
reach private synchronization state. That is confined to one spoke, so it
differs from a sideways import. Review it with the runtime-struct placement
decision below; do not expose locks through public interfaces just to remove
the cast.

### 4. A job snapshot retains a mutable pointer into live state

**Internal ownership defect; no public observer corruption demonstrated.**
`Job.Snapshot()` copies the struct but shares its `ExitCode *int` with the live
job (`internal/jobs/jobs.go:41`). A recipient can mutate that pointee.
The public Agent snapshot and observer paths deep-copy their data, so this
finding concerns the child/service boundary, not a proven public leak.

Copy the optional exit-code value when returning the snapshot. Preserve the
distinction between an absent exit code and a real zero. Apply the same
ownership review to newly added report/result structs and nested JSON values.

### 5. Usage belongs to Engine, but its synchronization is implicit

**Maintenance concern; no race demonstrated.** Engine owns the usage map;
Agent's state mutex protects accounting during append and public reads.
The relationship is not expressed beside the map or its accessors. Agent also
has separate operation, append and close locks, and Jobs has its own mutex.

Record which lock or serialized owner protects each mutable field and the
allowed call order. Preserve atomic accepted-response accounting. Moving a
mutex onto Engine is an option, not an automatic improvement if it obscures
the acceptance transaction. Chapter 5's actor must inherit an explicit rule
for these updates instead of another convention learned from call sites.

### 6. Runtime-struct placement: clarified by Bill

The requirement
puts core data structures and interfaces in common and permits truly private
implementation details in spokes. The student has treated concrete runtime
owners as private implementations behind common interfaces.

Asked whether that exception includes Engine, Registry, Jobs and Job, Bill
answered: "Private runtime structs are acceptable if common interfaces expose
the ownership chain." The architecture ledger and mandatory skill now record
that ruling; the author is incorporating it into Chapter 1. No blanket
relocation of these implementations is required. Shared values and interfaces
remain in common, with behavior in the responsible spoke. Parent capability
and snapshot ownership findings above still apply.

## Related review work and limits

The independent Chapter 4 reviewer had already identified supervision JSON
parsing in Jobs and a redundant Agent argument to Call.dispatch. The coder's
`ee15a56` revision moves wire handling into tools and introduces typed Jobs
operations and result/report values. That is the right direction: Jobs should
supervise work through domain operations rather than interpret tool JSON.
Independent revision acceptance remains pending.

The source-cap correction similarly separates retained output bytes from
presentation metadata. Its new `ExecutionResult` and `JobReport` types should
be reviewed for that invariant alongside the existing snapshots.

Synchronous observer callbacks are allowed at the current chapter scope;
Chapter 5 explicitly introduces queued delivery and actor ownership. This
review does not label the absence of future functionality a current defect.
It also does not claim that every lifecycle path or all future chapters have
been reviewed. Initial probes used exported source in a temporary directory;
the coder and independent reviewer are adding durable regression coverage.

## Handoff

Findings 1 and 2 were sent to the coder, author and independent reviewer.
Findings 3–5 were discussed with the independent reviewer, who corroborated
their scoped descriptions. The coder is preserving the initial implementation
and run receipts while fixing defects. Author changes must teach the relevant
invariants, so the next student can reproduce the corrected design.
