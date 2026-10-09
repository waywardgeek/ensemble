# Chapter 11 public integration preparation

Basis: complete frozen 439-line plan `55c9e7a44364123dd28ef7967a0a3663d74de925`,
current Chapter 11 teaching `b43fc3f`, architecture and freshly reloaded entire
coding skill. The coordinator accepted the independent plan review and the
three published clarifications. The coder exclusively owns the Go compiler.
No mutable student implementation was read, compiled or edited for this work.

This is a finite integration plan and runnable fixture preparation, **not runtime
acceptance**. The full §11.10 map in `chapter-11-grader-review.md` remains required.
The original foundation and initial CLI subset are unchanged.

## Public adapter integration boundary

The frozen plan supplies sufficient proposed seams: service Configure/Prepare/
Reopen/Close; a constructor receiving the actual Connection; owned Send/Receive/
Abandon complete messages; Connection→Service→Ensemble diagnostic access; built-in
memory endpoint construction; frozen Agent bindings/state; and context-aware
session open/import. A separate consumer module will import only the public
Ensemble module and implement its own constructor/transport. It must not import
an internal package or use a process as its custom transport.

Bind that consumer to the actual published declarations when the coder reports
them. Proposed spelling is not a private assertion or a compilation target yet.
The generic JSON service must be reachable through the revised common owner
interfaces; transport tests need only their legitimate diagnostic route. No new
public-seam teaching gap was found in the plan plus published clarifications.

The consumer will load the same complete-message scripts for three runs: delivered
stdio, delivered memory adapter, and independently implemented public transport.
Only stdio adds/removes LF framing. The public transport records actual parent,
message correlation ID, owned input bytes, cancellation ID/bytes and close/join
observations. It has bounded channels and explicit barriers so the driver can
wait for a call's issue before issuing another or killing its Job. Do not infer
issue from the model's proposed call or a delay. No transport chooses protocol
metadata, cancellation reasons or grant authority.

The scripted peer is separate from the client oracle. Identical peer transcripts
alone do not establish matching Jobs, artifacts, safe state, no reverse effect,
owner isolation or client worker cleanup. Those assertions belong in the public
consumer/runtime receipt. State snapshots and returned schema byte slices need
owned-copy controls; mutations of copies must not alter later snapshots or calls.

## Prepared message and schema inputs

`scripts/edition2/ch11_message_scenarios.py` exports seven compact protocol scripts:
ordinary call; reversed two-call replies; kill one pending call then send its late
reply plus the sibling reply; safe remote error followed by a healthy call; future
unissued ID; stale duplicate; reverse server request. Fault cases permit an
optional valid cancellation during teardown: the fixture does not invent a ban
on that cleanup notice. The reverse request targets a sentinel file, whose
continued absence must be observed in the eventual client workspace.

The same file prepares six paired argument-vector groups: local escaped JSON
Pointer and `$ref` siblings; numbers around 2^53; Unicode scalar/array/property
counts; composition/enum/type union; exclusive bounds; annotations with no default
insertion. Five preparation pairs cover unsupported keyword, external/unresolved
reference, combined containment/reference cycle and semantically duplicate enum.
Raw JSON strings retain original numeric tokens; they do not pass through Python
binary64 as a convenience conversion. These known vectors are not a substitute
JSON Schema compiler or a full schema/bounds acceptance claim.

Preparation commands:

```sh
python3 scripts/edition2/test_ch11_message_scenarios.py
python3 scripts/edition2/ch11_message_scenarios.py --emit /absolute/new-corpus.json
python3 scripts/edition2/ch11_message_scenarios.py --peer ordinary --receipt /absolute/new-peer.jsonl
```

The peer reads LF-framed input and emits LF-framed output. The corpus contains
complete message strings without delimiters for the future memory/custom adapter.
All output paths must be absent. The exported corpus is small and reproducible;
it was removed after retaining its hash/size and four fixture test methods passed.
The direct fixture and subprocess produce identical replies for all seven scripts.
Targeted wrong-ID, incomplete-script, extra-message and cancellation-byte controls
reach intended refusals from valid fixture parents. These are **not** student
runtime positive/negative controls. The unchanged initial runtime checker controls
also pass. Exact commands and hashes are in `ch11-message-preparation.json`.

## Finite remaining integration groups

| Group | Positive parent and concrete distinguishing observations | Prerequisite |
|---|---|---|
| Owners/authority | Two live Agents with different local aliases and grants on one prepared connection. Correct remote mapping/effect, forced hidden call with zero sends, Skills unload changes future admission only. Agent A close cancels A while B finishes; snapshots cannot mutate service/bindings. | Published concrete public declarations; genuine created Agents |
| Lifecycle | Run the seven scripts through stdio/memory/custom. Correlate returned Job/artifact identities; retain killed state after late response; reject future/reverse messages with no local effect; inspect safe errors and healthy sibling/other-key isolation. Add barrier-controlled 64 permits, staged cancellation retention, one-second stalled delivery, absolute deadline, root close and worker joins. | Public transport declarations, documented Job/public scheduling and close operations |
| Schema/results/bounds | Use accepted declared definitions and valid argument/result parents before paired refusals. Check raw accepted artifacts, local no-send, complete result validation and isError. Extend retained foundation exact/+1 vectors into actual client observations, including accumulated definitions and compilation/validation budgets. | Public preparation/call path; no private schema function dependency |
| Persistence/watch | Create genuine v1 and v2 stores through public operations, then compare checkpoint/tail/offline reconstructed requests. Check incompatible logical selection and corrupted bytes before constructor count changes, compatible physical-route replacement without historical calls, and historical Jobs unavailable. Offline v2 uses recorded unavailable bindings; standalone log has mcp:[]. | Published actual v2 codec and public session/watch declarations |
| Owner-state limits/structure | Review disclosed owner-local test fixtures, execute last-valid and first-overflow allocation/issue, observe unchanged owner state and zero new transport/handoff. Inspect actual import/parent routes and bounded cancellation storage; no production setters. | Student-published fixture entry points and validity explanation; compiler release |

Persistence negatives start from an accepted public-created store and retain its
original files/hashes. Field mutations must recompute only required hashes and
remain otherwise valid, so failure occurs at the targeted semantic condition.
The proposed `state.mcp_bindings` spelling is recorded in the plan but will be
checked against the final published codec before integration. Do not manufacture
a v2 snapshot merely to make a corruption test convenient.

The exhaustion fixture can represent a fully settled contiguous issued prefix
without retaining a tombstone per old ID; its bounded pending/permit state still
must be valid. The last legitimate issue/generation is a required parent, not
just a cursor assignment followed by an error. Independent review of that proof
and the actual constructor/handoff observations precedes mutation credit.

Structural review includes behavior-bearing schema/parser/timer helper children,
not only the service and transport. Verify each actual stored parent interface,
constructor and route to the owned JSON/logging services. A concrete `*Service`
parent on an owned schema helper does not satisfy the common-interface rule.
Immutable syntax/value nodes with no independent behavior/lifetime are a different
case; no parent is invented merely because a data structure is a Go struct.
The coordinator reported an early implementation drift for correction, without
asking this grader to inspect or grade mutable student code.

This preparation does not waive remaining full keyword/boundary cases, byte
ownership, race/late-generation controls, renderer and inherited behavior,
optional GUI/watch integration, structural/delivered-module checks or runtime
deletions. It does not authorize paid calls. The reviewed real-provider matrix,
actual initial live experience and subsequent historical comparison retain their
separate gates. No new application feature or private API is demanded.
