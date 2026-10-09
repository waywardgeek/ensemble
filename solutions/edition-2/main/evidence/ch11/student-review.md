# Chapter 11 initial student review and plan

Plan only, 2026-10-08. No runtime edits, builds, tests, checker invocations,
credential reads, provider calls or workers. This is the initial experience,
before implementation, live runs or historical comparison. Request coordinator
contract review of the concrete seams below; none is an implementation release.

## Source identity and actual reads

Only baseline: accepted `edition-2-ch10-r1`, peeled commit
`d91861207f4c2e4b81bd6c239ae15b8f51a1d146` (annotated tag object
`f380466ed95945a416c1615f8d0464ba5a0727ec`). Source/evidence commit
`265fe34434c84bc695740eb09c632dc632522c02`; its main subtree is
`5bb92f0d3b30da864c1bebb351ab7065973488ec`, **not** its outer repository tree
(`f2da7e44a8211fd8a5b5bd0f7fbfe41eed00c6a5`). Released runtime identity
`70d86f7` is supplied by the handoff and pinned Chapter 10, not a new run.
At inspection HEAD was the peeled release commit; tracked main had no diff
against the source/evidence commit. Tree metadata confirms 5,792 files and 12
Go modules. Existing untracked Chapter 10 binaries/build directories appeared
in scoped status; their contents were not opened or changed. Prior evidence
keeps its original identities and is not Chapter 11 evidence.

Teaching root: `/Users/bill/projects/ensemble-edition-2-revisions/ch11-student-inputs/new-only`.
Read `manifest.json` and verified all 13 entries before reading pinned teaching.
Manifest SHA-256: `e685d8ab2b09f11489cd0efa63513318531e533d944a9f70f88c3f685512cde9`.
It declares teaching commit `d91861207f4c2e4b81bd6c239ae15b8f51a1d146` and
the predecessor identities above. The handoff attributes Chapter 11 clarification
`5d6255f` to independent acceptance `d13738e`; I did not inspect those reports.

| Pinned file | Verified SHA-256 |
|---|---|
| chapter-01.md | 714beccbd6c11e6629c7577fdb078601d0fec0ce0e8c74888c6e87711409db47 |
| chapter-02.md | 698533c7881f89c5238b2cfa1a72f64b7e3869d1c4fc5f6c646da072cb982c37 |
| chapter-03.md | 85438fad9357d80b045b9dd62fb55d9abdbc7986403bcef9b0ad53c7b66e9214 |
| chapter-04.md | b18b22663ceb4a9bbe74663a830423b0762afaee783e5f92b706c483d5289a46 |
| chapter-05.md | 6667c4a66d57de906b0b3ded91793f2d5e852bf19cf31d1244604d631419fec1 |
| chapter-06.md | d5a999c99ab97dc326bdec00d097fa272874d71272f1e3b7e2e0087da9a25eb6 |
| chapter-07.md | 61f70c392b981c792701d805ade8b4db2418baf369f949fdae7c321821d01206 |
| chapter-08.md | a2b02256f0b88bf72900268477f0cd44e9d6af7de05de3f2d91d7c88c99cd7f6 |
| chapter-09.md | a9d7d104176b347df4d4627d23748641d452a4e2a9b7a9db1a80913590096b84 |
| chapter-10.md | 705e6e49f7939a2fc6fa41778727ffa0fb68455634eeb4f57a316dfa71077833 |
| chapter-11.md | 8fb1b57aa8d30d29d529464420ef6ae704c12f794fe8b25c5425ae492f42d55e |
| architecture.md | 9dada72e65c36fb649659f9076b4ca2bdfc8379fef919a3b575652dca0e49327 |
| skills/ensemble-coding/SKILL.md | 0180e3eb5d4f6032936729f31b473c70c421bce22001790a9709b05c57893bd1 |

Content-read ledger (line numbers refer to these pinned/accepted versions):

- Entire mandatory repository `book/edition-2/skills/ensemble-coding/SKILL.md`
  first; its hash equals the pinned skill. Entire pinned skill, architecture,
  Chapter 1 and Chapter 11. Initial combined output was truncated; bounded
  reads subsequently covered both chapters completely.
- Pinned chapter heading searches for Chapters 2–10, then inherited contracts:
  Ch02 lines 90–205; Ch03 59–262; Ch04 42–432 and 470–521;
  Ch05 34–163 and 306–494; Ch06 21–116; Ch07 209–286;
  Ch08 82–116 and 245–412; Ch09 32–298 and 378–460; Ch10 1–940.
  Truncated combined inherited output was reread in bounded chunks where needed.
  Hash verification reads all teaching files as bytes; that is not a claim to
  have studied the unlisted portions of Chapters 2–10.
- Accepted main, complete: `go.mod`, `persistence-format.md`, `session-api.md`,
  `internal/common/{session,actor,watch}.go`,
  `internal/tools/{registry,supervision}.go`.
- Accepted main, excerpts: `ensemble.go` 1–240, 342–404, 830–881;
  `session.go` 126–198, 355–582; `internal/common/types.go` 18–39, 255–418;
  `internal/jobs/jobs.go` 1–228; `internal/persistence/json.go` 1–210;
  `internal/persistence/codec.go` 1–118, 298–436.
- Declaration searches (`type`/`func` lines only) additionally covered
  `internal/llm/{session,watch}.go` and the above ensemble/session/common/
  jobs/registry/codec/json files. Scoped Git status/diff, revision/tree metadata
  and accepted main filename listing only. An initial `git ls-tree` exclusion
  pathspec failed without reading file contents; retried with filename filtering.
- Coordinator inbox only:
  `/Users/bill/projects/ensemble-edition-2-revisions/ch11-student-inputs/coordinator-inbox.md`,
  at start, after contract/source inspection, before writing and before commit.
  These reads contain only the plan-only release. The final check is recorded
  in the handback if it changes that instruction.

No excluded source, report, memory or conversation was opened; no accidental
excluded content exposure is known. Historical references inside permitted
teaching were read as that teaching, without following their links. No external
specification or provider documentation was fetched. This plan follows the pinned
wire contract, not a claim of independently verifying current external standards.

## Owners, facts and lifetime

| Owner / responsible implementation | Authoritative facts and parent route |
|---|---|
| Ensemble / public composition root | Logger, MCP service, root shutdown, existing Agent/store reservations. Root has no invented parent. |
| MCP service / new `internal/mcp` | Configured logical keys, current Connections, monotone service generation allocator, attachment index to Agents. Parent is Ensemble. |
| Connection / `internal/mcp` | State, immutable generation/key, catalog, issued watermark, pending map, 64 operation slots, protocol deadlines, transport. Parent is service. |
| Transport / new stdio and memory spokes | Framing/endpoint, bounded byte buffers, stderr tail and owned I/O/process workers. Parent is actual Connection. |
| Agent / existing root | Frozen selections/definitions and creation identity. Registry holds derived executable declarations; Skills retains committed grants. Parent is Ensemble. |
| Job / existing `internal/jobs` | Remote attempt association, cancellation lifetime, artifact, terminal state and report cursor. Parent Job→Jobs→Agent; Tools reaches MCP via that chain. |
| Actor / existing `internal/llm` | Orders durable call/report/terminal facts and transient watch changes. Agent's single existing Context remains authoritative. |
| SessionStore / `internal/persistence` | Existing lock, file identities, checkpoint worker. Store→Agent→Ensemble; owned captures are values, not another mutable Context. |

Shared values/interfaces go in common; public aliases make external implementation
possible. Private runtime structs stay with the responsible behavior. Root alone
imports concrete spokes. MCP, transports, Tools, Jobs, llm and persistence do not
import one another. No protocol decoding in a transport, runtime globals, callback
service bags or extra Agent Context.

Agent selections, definition semantics, SessionID/format, logical connection key
and each generation's transport configuration are creation-only. Reopen replaces
the Connection after joining the old one; it cannot mutate bindings. Runtime
state/cursors/cancellation flags are mutable only at their owner. Catalog schemas,
compiled validators, installed declarations, snapshots and Actor status projections
are owned derivatives; no mutable shared map/slice escapes. The service attachment
index references frozen Agent facts, not an independently editable grant catalog.

Brief root/service locks protect registration, generation allocation and lookup;
never span transport construction, I/O, joins or Actor calls. Per-key lifecycle
serialization orders prepare/reopen/close without blocking other keys. A Connection
mutex orders issue/settlement/slot state; release it before I/O or owner delivery.
Jobs locks protect lifecycle/report state; remote waits and cancellation joins
occur outside them. Actor receives owned facts after service/Connection locks are
released. Attach/detach and ready publication use the service boundary so a new
binding cannot miss a concurrent reopen definition check.

Each accepted connection-state transition is delivered through Ensemble to bound
Actors before the next transition for that key is published; delivery is off the
protocol reader/writer and outside locks. Actor records the generation and ordered
state fact, advances watch revision and emits complete `mcp_changed`; identical
projections coalesce. Closed Agents refuse delivery, and older generations cannot
overwrite newer projections. Watch cut/tail uses the existing Actor boundary.
This is presentation only: dispatch consults actual service state and frozen
binding compatibility. No transient status enters durable history or Skills.
Allocate a positive service-wide uint64 generation before constructing each
transport, with overflow refusal before mutation. RPC ordinals are separately
positive uint64 values scoped to that Connection; reopening resets only that
new Connection's ordinal, never the service generation cursor.

## Proposed public seams (declarations only)

These names are proposed for checker/consumer review. Values/interfaces are public
aliases from common; external code imports only `example.com/ensemble`. `context`
below is Go cancellation context, not conversation Context.

```go
type MCPRoot interface { Logf(string, ...any) }
type MCPConnection interface {
    Service() MCPService
    Snapshot() MCPConnectionSnapshot
}
type MCPTransportConstructor interface {
    Open(context.Context, MCPConnection) (MCPTransport, error)
}
type MCPTransport interface {
    Connection() MCPConnection
    Send(context.Context, MCPMessage) error
    Receive(context.Context) ([]byte, error)
    Abandon(context.Context, string, []byte) error
    Diagnostics() MCPDiagnostics
    Close() error
}
type MCPMessage struct { RequestID string; JSON []byte }
type MCPConnectionSpec struct {
    Key string
    Transport MCPTransportConstructor
    CallTimeoutSeconds int // 0 selects 120 in Go API; explicit config file: 1..600
}
type MCPSelection struct { Alias, Connection, RemoteName string }
type MCPDefinition struct {
    Name, Description string
    InputSchema, OutputSchema json.RawMessage // nil output = absent; false is present
}
type MCPBinding struct { Selection MCPSelection; Definition MCPDefinition }
type MCPConnectionSnapshot struct {
    Key, State string
    Generation uint64
    ProtocolVersion string
    DiscoveredCount int
    ErrorCode string
}
type MCPBindingSnapshot struct {
    Alias, Connection, RemoteName, State string
    Generation *uint64
    ProtocolVersion, ErrorCode string
}
type MCPDiagnostics struct { StderrTail []byte; Truncated bool }
type MCPService interface {
    Ensemble() MCPRoot
    Configure([]MCPConnectionSpec) error
    Prepare(context.Context, string) (MCPConnectionSnapshot, error)
    Reopen(context.Context, string) (MCPConnectionSnapshot, error)
    CloseConnection(string) error
    Connections() []MCPConnectionSnapshot
    Definitions(string) ([]MCPDefinition, error)
    Diagnostics(string) (MCPDiagnostics, error)
    Close() error
}
func (e *Ensemble) MCP() MCPService
// Config gains creation-only MCPBindings []MCPSelection.
func (a *Agent) MCPBindings() ([]MCPBinding, error)
func (a *Agent) MCPState() ([]MCPBindingSnapshot, error)
func (e *Ensemble) OpenSessionContext(context.Context, SessionOptions) (*Agent, error)
func (e *Ensemble) ImportSessionContext(context.Context, []byte, SessionOptions) (*Agent, error)
```

`Configure` validates/copies an atomic batch, rejects duplicate/existing keys and
the 32-key bound, and starts nothing. Constructors must own immutable configuration;
service never serializes them. Preparation requires a configured key. Already-ready
prepare returns a snapshot without I/O; failed/closed attempts require explicit
reopen. Only attempted generations appear in `Connections`; a configured unopened
binding projects closed/null-generation/unavailable. Reopen invokes the configured
constructor again after old endpoint closure/join. Changed physical configuration
can be deliberately supplied to a new Ensemble on resume; no in-place Agent
rebinding or physical reconfiguration API is added here.

External usage: create Ensemble, configure its service with a custom constructor,
explicitly prepare selected keys, set each Agent's `Config.MCPBindings`, and call
existing `NewAgent`. The constructor receives the real Connection; diagnostics
reach `transport.Connection().Service().Ensemble().Logf`. The Agent constructor
freezes ready definitions and rejects missing selections/collisions/ceilings before
exposure. It cannot prepare implicitly. Session opening instead performs the two
stages below; existing context-free session methods wrap the context variants.

Send takes one complete compact message plus opaque correlation identity; Receive
returns one owned complete message. Empty RequestID is reserved for non-request
delivery. Abandon receives the issued ID and Connection-encoded cancellation bytes;
stdio/memory send those bytes, a future stream adapter may close only that stream.
The adapter neither picks a cancellation reason nor examines JSON to find IDs.
Inputs are copied before retention; caller must not mutate them during the call.
Returned buffers are caller-owned. Close is idempotent, wakes blocked operations
before joining all adapter workers, and is callable concurrently with Send/Receive.
Open honors cancellation and cleans partial resources on error. No adapter may
leave an uninterruptible write worker behind.

Root factory functions will return constructors, e.g.
`NewStdioMCPTransport(MCPStdioOptions) (MCPTransportConstructor, error)` with
`Command, CWD string; Args, EnvAllowlist []string`, and
`NewMemoryMCPTransport(MCPMessageEndpointSource) MCPTransportConstructor`. The source
has `Open(context.Context) (MCPMessageEndpoint, error)` and supplies a fresh endpoint
per generation. It represents the application's external peer resource. The endpoint
has `Send(context.Context, []byte) error`, `Receive(context.Context) ([]byte,error)`,
`Close() error`, complete owned bytes and the same unblock contract. It is a
caller-selected physical resource, not the transport's parent. An independent
public consumer will implement MCPTransportConstructor/MCPTransport directly.

Internal common extensions expose the service through Ensemble, frozen binding
lookup through Agent, and a Job-owned cancellation context/worker completion.
Tools calls a narrow internal MCP execution interface with the actual Job and
admitted binding; there is no public alias-bypassing RPC dispatch method.
Connection/schema/parser helpers retain their responsible owner for diagnostics.
`MCPBindings` returns complete frozen definitions; `MCPState` is the exact safe,
alias-sorted §11.9 projection. Neither returns constructors, paths, argv,
environment, raw errors or stderr. Diagnostic tail is explicit application-only
access, 16 KiB with truncation; safe logs never copy it automatically.

## Admission, delivery and shutdown

Preserve current grants at Actor admission, literal next-attempt tool_limits
consumption, exclusive artifact creation and durable tool_called before execution.
Remote input is the entire object validated against its frozen schema: add/strip
no monitoring fields and insert no defaults. Chapter 4's setter supplies remote
report limits without changing the advertised schema. Invalid/hidden calls send
zero RPCs. The Job worker reaches MCP through its parent chain; the Actor never
waits for discovery, sending or remote completion. A successful wait operation
does not erase historical execution is_error. MCP send_input refuses.

At remote-attempt start establish the absolute 1–600-second deadline (default
120), then acquire an operation slot without waiting for capacity. Excess returns
mcp_capacity; there is no extra pre-send queue. Connection holds at most 64 slot
records encompassing staged request, pending reply and staged cancellation. One
delivery worker scans ready slots; cancellation changes that slot's flags rather
than spawning another goroutine/queue entry. Queue wait counts toward the one-second
delivery deadline for each staged request/notice. A stalled delivery faults/closes
the connection and settles every affected operation.
Discovery uses these same operation slots. Adapter buffering is bounded too:
at most 64 complete outbound messages and one decoded inbound message per
Connection, with backpressure/cancellation and the independent 8 MiB message cap.

Allocate contiguous `rpc-N` only immediately before transport handoff, after the
last pre-issue cancellation check. Install pending correlation first and burn the
ordinal even on failed delivery. Before issue cancellation releases the slot with
no ID or notice. After issue, cancellation removes pending once but retains its
slot until Abandon finishes or transport closes. Normal response settlement also
retains the slot until request delivery completes. A response that wins emits no
notice. Canonical IDs <= issued watermark without pending entry are discarded;
unknown namespace/noncanonical/numeric/future IDs fault the connection. No tombstone
set grows with canceled calls. Aggregate stale-reply diagnostics only.

The fixed cancellation bytes are
`{"jsonrpc":"2.0","method":"notifications/cancelled","params":{"requestId":"rpc-N","reason":"Cancelled by client"}}`,
with the actual issued N. Job kill preserves killed; timeout/capacity/transport/
remote/protocol failure is done plus is_error. Post-issue loss and timeout explicitly
report that the external outcome may be unknown. Never retry an effect. Interrupt
only ends model continuation/report waiting; it leaves admitted remote Jobs alive.

Connection owns its receiver, single delivery worker and bounded deadline/shutdown
coordination; adapter owns its framing/read/stderr/process workers. Fault paths
signal a shutdown coordinator instead of joining themselves. Generation tags fence
every callback. Agent close cancels/joins its remote workers and settles Jobs before
log close; it does not close shared transport. Ensemble shutdown first bars new MCP
admission, asks Agents to close so their kill facts retain inherited meaning, forces
transport close to unblock stalled deliveries, then joins service/adapter workers.
No root/service lock spans these steps. Stdio uses direct executable/argv and only
allowlisted present environment values copied at preparation. Close follows the
specified stdin-close, one-second grace, process-group TERM, one-second grace,
KILL/reap/join sequence on macOS/Linux. Other stdio platforms refuse explicitly.

Stdio checks line/UTF-8/8 MiB bounds during reads, adds/removes only framing LF,
rejects empty lines and nonempty EOF fragments, and drains stderr separately.
Memory retains exact message bytes with bounded buffering/backpressure. Future GUI
tunneling owns a logical channel in the optional GUI module; its close must leave
the physical socket and unrelated Pages/channels alive. No Chapter 12 tunnel now.

## Protocol, validation and persistence extension

MCP owns strict envelopes, fixed §11.4 metadata/version/clientInfo, mandatory
server/discover then all tools/list pages, schema compilation and call/result
validation. No initialize fallback, reverse execution, list-change mutation or
remote resource fetching. Discovery publishes only a wholly validated catalog.
Reopen compares every attached frozen selection, including currently ungranted
aliases, before ready; changed unselected definitions confer no authority.

Compile exactly §11.5's schema keywords with local-only pointers, combined
containment/reference cycle refusal, reference siblings and exact decimal numeric
comparison. Input root explicitly object; optional output object/boolean schema;
require structuredContent even on isError when an output schema exists. Compiled
validators are immutable owned derivatives and usable offline without a connection.
Whole result validation precedes any accepted artifact bytes. Preserve text order,
present-null structuredContent and isError; canonical accepted object plus one LF,
or exactly the safe error/message object plus LF. No accepted text prefix survives
an unsupported mixed result. Preserve every §11.7 stable code and int32 remote-code
validation without huge exponent expansion. Provider renderers retain the ordinary
single recorded ToolResult text path.

Enforce independent limits while reading/compiling, not after unbounded allocation:
8 MiB messages; depth 64/100,000 JSON values; 64 pages/1,024 definitions/16 MiB
normalized canonical definition sum; 4 KiB cursor/64 KiB description/256 KiB each
canonical schema; 4,096 physical schema nodes/64 expanded depth/100,000 validation
visits; 1,024 text blocks; 10-second discovery request/30-second whole preparation.
All exact boundaries remain accepted subject to inherited smaller bounds, including
1,024 total installed handlers and 16 MiB handler identity. Ignored metadata still
counts toward wire bytes/nodes. Limit failures retain a bounded safe artifact.

Proposed shared behavior adjustment for review: factor the existing lossless JSON
value parsing/canonical-number operations into an Ensemble-owned common-interface
service implemented in the persistence spoke. SessionCodec and MCP reach it through
their actual parents, never by importing persistence from MCP or manufacturing an
Agent for discovery. It takes explicit bounded JSON profiles; session event/raw
argument exceptions and session errors remain in SessionCodec, MCP policy/errors
remain in MCP. This reuses accepted number/Unicode semantics while preserving
different limits and the exact raw replay wrappers. No new mutable global/cache.

Two-stage live opening/import remains in root orchestration with Agent-owned
SessionStore, not a second resume Context:

1. Reserve canonical store, acquire OS lock, validate every available byte,
   version/hash/transition/prefix/tail/origin and settled boundary. Compare caller
   local handlers, Skills/System/policy and the complete logical alias/key/remote
   selection before any new MCP I/O. For v2, recorded remote definitions may seed
   an inert candidate Registry for historical and Skills validation only; it has
   no dispatch authority. V1 plus any remote selection refuses here.
2. Only then prepare referenced current connections in key order (or inspect a
   ready shared generation), validate their complete selected definitions against
   recorded identity, freeze bindings and attach them atomically. Recheck any
   concurrent replacement before publication. Only after success open append/create
   origin/initializer as appropriate and publish the live Agent. On failure,
   stored history is unchanged, reservations/construction resources are released;
   an already shared connection remains service-owned. Replay/inspection never
   prepares, invokes tools/call, fetches artifacts or restores workers.

Exact proposed v2 codec additions, to be published in persistence-format.md only
after release (not edited now):

- Checkpoint outer fields are unchanged; `version:2`, `state_version:2`. Log
  header/event envelope unchanged. Initializer session is exactly
  `{version:2,session_id,identity}`; anchor session exactly
  `{version:2,session_id,origin_as_of,origin_sha256,high_watermarks}`. Version 1
  retains absent payload version and its exact old shapes; explicit payload
  version 1 or another value refuses. Origin and checkpoint formats must agree.
- V2 identity is exactly `{mode,system,skills,handlers,mcp_bindings}`.
  `mcp_bindings` is required, nonempty, alias-sorted/unique; each item exactly
  `{alias,connection,remote_name,definition}`; definition exactly
  `{name,description,inputSchema,outputSchema}`. Output null means absent schema,
  not boolean false. Schemas are semantic JSON values, **not** Raw strings.
  Definition name equals remote_name and corresponding handler name/description/
  schema equals alias/description/inputSchema, with no local-handler collision.
- V2 State is exactly `{session,context,usage,skills,limits,window,mcp_bindings}`.
  New state.mcp_bindings is the same required frozen array, canonically equal to
  identity.mcp_bindings. This is an owned capture of Agent bindings. Existing
  state.session fields remain `{id,identity,as_of,high_watermarks}`; its identity
  uses v2. V2 Context.Session (and any codec SessionFact position) has the five
  existing fixed codec fields plus required `version:2`; initializer unused
  anchor fields remain 0/empty/null. Context.Session.Identity uses v2. Actual
  event payloads retain the sparse shapes above. No other Context fields change.
- V1 decoding/encoding accepts only its existing field sets and zero remote
  bindings, without an empty mcp_bindings addition. All existing Raw wrappers,
  exact opaque/argument strings, number rules, state hash, correspondence checks,
  settled checks and historical-job restrictions continue unchanged. V2 schema
  fields need explicit semantic encoding rather than generic Raw string dispatch.
  Empty normalized remote descriptions must pass v2 identity validation; the
  accepted codec's blanket nonempty handler-description check needs narrowing.
- No connection state/generation, RPC cursor, endpoint, process, diagnostic tail,
  physical route or credentials is saved. Formats/bindings stay creation-only;
  public append cannot initialize/adopt/change them. Owned public/watch state adds
  required `mcp:[]` or the exact §11.9 items. Offline v2 status is closed, null
  generation, mcp_unavailable; browser numbers stay lossless under Ch08 policy.

## Questions and initial teaching feedback

The ownership table, operation-permit accounting, literal artifact and two-stage
resume ordering are useful and actionable. No failed runtime approach or provider
problem is claimed: only reading and planning occurred. The two necessary baseline
adaptations found are the codec's empty-description restriction and Jobs' current
nonkillable-function path; MCP needs cooperative cancellation and joining without
changing the truthful limitation for arbitrary local functions.

Concrete coordinator questions before affected implementation:

1. **Standalone offline binding identity:** §11.9 preserves explicit standalone
   CH02_LOG behavior and requires offline watch to use recorded bindings. §11.8
   publishes identity only for session initializers/checkpoints; existing standalone
   logs have neither. For an MCP-enabled standalone log, should offline mcp be
   empty/unknown, or is another exact creation record intended? Proposed narrow
   reading: frozen offline binding status is available for v2 session stores;
   standalone history still renders recorded calls/results but cannot invent a
   logical binding. Please settle the teaching; no new event is presumed.
2. **Shared JSON owner check:** is the proposed Ensemble-owned bounded JSON-value
   service in persistence acceptable for reuse by MCP and SessionCodec? It avoids
   a sibling import, a fake Agent, or divergent precision/Unicode implementations.
   Please review this along with transport parents, Agent binding authority and
   the exact v2 state spelling before affected code.
3. **Exhaustion seam and runtime gate:** may the required uint64 exhaustion controls
   use disclosed owner-local `_test.go` fixtures representing an exhausted valid
   generation/contiguous issued prefix, with no production cursor setter? Please
   confirm the checker-facing mechanism and publish the actual runtime invocation
   against these seams before integration; no private identifier is assumed.

The foundation command remains
`python3 scripts/edition2/accept_ch11.py --self-test --receipt PATH`; it validates
oracle/peer preparation only and was neither opened nor run. Historical
`make grade-dir CH=12 DIR=solutions/edition-2/main` remains diagnostic. This plan
does not rerun accepted predecessor tests or claim runtime acceptance. Later work
still needs full deterministic/inherited/module/race checks; a reviewed bounded
all-three-provider feature matrix; actual PTY, browser and public custom-transport
multi-Agent runs; initial experience freeze; independent historical findings;
improvements and author-response confirmation. No paid run is released here.


## Implementation release acknowledgment (2026-10-08)

Continued this same student conversation from 55c9e7a. Reloaded the entire
mandatory coding skill and pinned architecture. Verified the three-file
clarification-b43fc3f manifest before reading all three files completely.
Teaching commit: b43fc3f76f2631e70cd5d24aadb06a2927c4fdb3. Manifest SHA-256:
03748a78f19f5d66adf1dbc2aa10addd71848ccb0d8796d4b765f4888e07a7a4.
Chapter SHA-256 c9fd192912baa555ae29053bdcf2cf3a3f367d69dfcd6d4941cc3b2b4cc0531b;
direct feedback c4829a7f9eeed73ad91bc9d812b9f55c33a13a1abcc91815c52eed1ff7195130;
runtime command 3684a30215114b8b93db09a3e13a4ac75dbbd458bfbb2ccdfe01696333cb9b53.
Architecture/skill pins remain unchanged. Inbox confirms sole compiler ownership,
local implementation/checks only. Reviewer sources/reports were not read.

Grouped answers accepted before affected edits: Q1 offline standalone watch is
exactly mcp:[], without inferred identity. Q2 shared bounded JSON belongs in a
neutral `internal/jsonvalue` spoke owned by Ensemble, not persistence. Q3 tests
may seed valid owner-local near-limit state, prove the last successful allocation
and overflow before mutation/handoff, without production setters. The partial
runtime command requires an actual observed build and complete source map and
is not whole-chapter acceptance. No further plan-only cycle is requested.

Correction to the original parent/API proposal: common.Ensemble exposes JSON()
returning common.JSONService. MCPService.Ensemble() returns that same coherent
common.Ensemble interface; the public MCPRoot alias names it, rather than the
original Logf-only interface. Connection.Service().Ensemble().JSON() and
SessionCodec.Agent().Ensemble().JSON() reach the real shared owner. No injected
JSON field, sibling import, capability cast or synthetic Agent. Shared syntax
has explicit bounds, while MCP/session policy and raw argument exceptions stay
with their responsible spokes. The frozen initial plan above is preserved.

Public seams will be published in mcp-api.md and the exact version-2 appendix
in persistence-format.md before runtime integration. These documents, plus the
common declarations, are the coordinator notification surface for independent
public/custom-transport integration. Implementation remains subject to complete
local and inherited validation and a separately reviewed bounded live matrix;
credentials/providers are not released.

First local failure: gofmt on the explicit changed-file list exited 2 with
`internal/mcpstdio/stdio.go:38:199: expected ':', found '}'`.
Unformatted failed file SHA-256: 3145534d70a83db66629f0fbb5db4e4ee674f3b488752e2ac16c9da4bb231918. This is a student syntax error
in the shutdown select, not teaching/checker behavior. Other files formatted
before the failure; no test/build receipt is claimed for that invocation. Fix
is adding the missing empty-case colons. Subsequent checks receive source maps.

### Local continuation: common parents (coordinator inbox)

Reloaded the entire repository skill after compaction and reread pinned architecture and the coordinator inbox. Coordinator identified `schema.parent *Service` and `connection.parent *Service`. This was my code fault, not missing teaching: a getter returning an interface does not repair a concrete stored parent. Changed both constructors/stored parents to common interfaces; schema behavior uses the actual service's root JSON owner, and connection lifecycle publication reaches its service through `MCPConnectionOwner.ConnectionChanged`. No sibling injection or concrete cast. Compile-03 passed after v2 codec integration; this is compilation only, not behavioral validation.
