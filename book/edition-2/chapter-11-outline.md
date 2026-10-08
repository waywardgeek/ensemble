# Chapter 11 outline: Tools across a boundary

**Stake:** a reader wants to add a useful external tool without rebuilding the
Agent, while keeping control of what it may execute and keeping the interface
usable when the external program stalls or disappears.

Preparation only, October 8, 2026. New Chapter 11 maps to first-edition
Chapter 12, MCP. The current global-review table calls old Chapter 12 “voice”;
that label is stale against the actual chapter. Follow the workflow's numeric
map and the actual source, with old Chapter 13's GUI self-observation as the
next chapter's consumer. Chapter 10's contract is accepted for checker
preparation, but its implementation/checkpoint is still a prerequisite. No
Chapter 11 student, grader pass or actual spin is claimed.

This outline includes coordinator-accepted direction and explicitly marked
proposals requiring review. It is not a student-facing complete contract.

## What the chapter preserves from the old telling

Old Chapter 12 begins with the inconvenience of compiling every new tool into
the Agent. Keep that concrete reason for an extension protocol. Its strongest
mechanical explanation is that a tool's transport changes while ordinary
Agent call/result handling should not. There is no documented personal incident
in that chapter to invent into an opener.

Later old Chapter 21 supplies a useful consequence: an MCP configuration could
parse on a primary skill yet never connect, because connection happened only
in a later load callback. Use that recorded architectural failure to motivate
explicit preparation and visible connection state. Do not copy its old callback
wiring or historical service/pricing claims. Actual Chapter 11 use will supply
the demonstration; a successful discovery reply cannot substitute for a tool
effect or a cancellation observed through the reader's interface.

Voice plan: open with the reader trying to add a tool; follow its path from
configuration through discovery, authorization, remote work and a usable result.
Return to the same task after the failure and shutdown rules. Explain the owner
plan before interfaces. Keep wire fixtures next to the rule they establish;
put source archaeology and grader limitations in the evidence file.

## Settled requirements and proposed scope

Bill explicitly requires MCP tunneled over WebSocket so the Agent can see and
control its GUI in the following chapter: “The MCP transport should not be
hard-coded.” The coordinator recorded this requirement at `ad2b2e8` in the
architecture ledger. It is a required extension seam now, not an optional
refactor when the browser needs it.

The coordinator accepts a clearly versioned **MCP 2026-07-28 tool-client subset**
for this outline. The official versioning page names that revision current on
the research date. Earlier initialize-based revisions remain a separately named
compatibility topic; they are not described as current MCP. The first edition's
`ephemeral` tool field and reverse `tools/call` are not inherited execution
rights. See the primary-source links in the evidence record.

Required in this chapter:

- Complete-message transport interface with public construction/embedding seam,
  declared ownership, cancellation and close semantics. Protocol/discovery/call
  code cannot assume line framing, a subprocess, stdin or a WebSocket.
- A production stdio adapter and a headless in-memory message adapter, exercised
  by the same protocol checks. A separate public consumer implements a custom
  message adapter without importing internal implementation packages.
- Current-version discovery, paginated tool listing, exact response correlation,
  authorized calls, tool errors, malformed peer handling, timeout/cancel and
  teardown. Ordinary model-issued remote calls use existing Jobs/report rules.
- Per-Agent installed aliases and grants; no automatic installation from a peer's
  discovered names. Skills may grant those preinstalled aliases using Chapter 9's
  unchanged authority rules. A remote description never becomes an instruction.
- Human CLI, ordinary browser display/control and public headless/multi-Agent
  demonstrations with all three model APIs. The browser can submit an external
  tool task through the ordinary Agent interface without implementing MCP itself.
- Historical replay and session restart preserve recorded results and current
  permission checks without calling the external server to reproduce history.

**Proposed next-chapter boundary:** Chapter 12 implements the actual WebSocket
adapter in the optional GUI module and makes the Agent observe/control its GUI.
Automatic round/turn context collection and any skill-triggered connection
lifecycle should be taught there with their first concrete consumer. This
moves old Chapter 12's forward sketch beside old Chapter 13's actual wiring;
it does not waive either the transport seam or the future GUI capability.
No microphone input, hosted HTTP/OAuth transport, sampling, elicitation, tasks,
remote skill installation or general server-initiated Agent tool execution is
added incidentally. These exclusions need coordinator confirmation before draft.

## Proposed teaching order

1. A connection can advertise a tool without authorizing it for an Agent.
2. One owner for each connection, codec, message transport and admitted call.
3. A versioned MCP subset: exact requests, results, errors and refusal of another
   protocol era. Contrast one small historical handshake example if useful.
4. Complete-message I/O first; stdio is an adapter with a process lifetime.
5. Discovery as candidate data, then explicit aliases and immutable installation.
6. Route an authorized call through Jobs; cancellation is local lifecycle, not
   proof that an external effect was undone.
7. Record useful result bytes without turning metadata or remote text into rights.
8. Prove alternative transport and public embedding before building browser MCP.
9. Actual user exercise, then inherited grader gaps and distinguishing checks.

## Owner and data plan for coordinator review

The root service direction is plausible to the coordinator, but exact lifetime
and preparation/resume choices below remain proposals.

| Owner | Authoritative data and responsibility | Actual parent route |
|---|---|---|
| Ensemble | Configured named MCP service and application shutdown | Application logger lives here |
| MCP service | Explicit connection identities, connection creation/close, current connection generations; no Agent grant state | Service → Ensemble |
| MCP Connection | Frozen discovered descriptor candidate, RPC identity allocator, pending requests, protocol state and child transport lifetime | Connection → Service → Ensemble |
| Codec or parser child | Correlation/framing-independent validation and bounded pending operations, if split from Connection | Codec → Connection |
| Concrete transport | Its logical endpoint, adapter I/O workers and resources it actually owns | Transport → Connection |
| Stdio transport | Child process, pipe ends, stdout framing and bounded stderr drain; joins child once | Transport → Connection → Service |
| Agent | Creation-only selected remote bindings and installed local aliases | Agent → Ensemble → MCP service when service access is needed |
| Registry/Skills | Existing handler definitions, ceiling, active grants and ordinary argument/admission logic | Registry or Skills → Agent |
| Job | Admitted remote operation, local output artifact/report state and cancellation association | Job → Jobs → Agent → Ensemble → MCP service |
| Actor | Durable call/result acceptance, grant transitions and watch publication | Actor → Agent |

Shared serializable values and owner interfaces stay in common. MCP behavior
belongs in its spoke; Tools retains ordinary tool decoding/admission and calls
the typed service through Agent's actual parent. Neither imports the other's
implementation. Private runtime structs remain allowed when interfaces expose
these routes. Parser, timeout, framing and process diagnostics retain the same
route; “stateless helper” is not an exemption.

Construction proposal: the composition root selects a concrete adapter
constructor, then Connection invokes it with its actual parent interface before
publishing readiness. The constructor creates a transport child with that
Connection back-pointer. An external public adapter gets the same parent,
can reach diagnostics, and returns a resource with complete-message behavior.
This is a construction interface with owned configuration, not a runtime bag of
Send/Recv/Log callbacks or an injected sibling service. Exact method spelling
is left to the student after the lifecycle is reviewed.

A transport owns its **logical endpoint**. A future GUI adapter may share a
physical WebSocket owned by a GUI Connection; disposing the MCP endpoint must
remove its routing/subscription without closing unrelated chat, settings or
speech traffic. The optional GUI module's physical-socket bridge retains its
own GUI owner chain. The complete-message boundary must accommodate that split
without adding GUI imports to the core. Chapter 12 must publish the concrete
multiplexing/generation/close contract before implementing it.

Root connection lifetime is independent of an Agent turn or skill load. Closing
one Agent cancels/joins that Agent's pending operations without closing a
connection still used by another Agent. Unloading a skill revokes subsequent
admission while already admitted Jobs retain their captured permission. Root
close ends admission, cancels owned operations, closes adapters and joins their
workers. No background worker synchronously calls the Actor while Actor waits
for that worker.

## Candidate wire and transport contract

These are proposed application bounds/choices, not claims that MCP mandates
these limits. The full chapter must publish complete literal fixtures and
stable safe errors after review.

- Use version `2026-07-28` on every request. Client capabilities are explicitly
  empty for this subset. Send `server/discover` first as an application choice,
  require that version and the tools capability, then list tools. A recognized
  unsupported version or old-era reply fails clearly; no silent initialize
  fallback and no automatic replay of a call that might have had an effect.
- Proposed outgoing IDs are monotonically allocated strings `rpc-1`, `rpc-2`,
  scoped to a connection generation. A uint64 cursor refuses exhaustion before
  send; IDs are never reused in that generation. Matching is exact, not a
  float64 conversion. A response is delivered once only to its own waiter.
- Reject malformed UTF-8/JSON, duplicate members, invalid envelopes, both result
  and error, unusable IDs and an unexpected server request. Current MCP has no
  server-initiated JSON-RPC request direction. Close that connection with a safe
  diagnostic; never dispatch a peer's `tools/call` into Agent tools.
- A late response after local cancellation is ignored; a duplicate response
  cannot settle twice. Closing/EOF unblocks every pending caller with its own
  safe error and fences old callbacks from a replacement generation.
- `Send` and `Receive` exchange one owned complete UTF-8 JSON message; no trailing
  LF is part of the generic contract. Stdio encodes/decodes LF framing itself.
  Whole-message ownership prevents reused caller buffers from altering a send.
  Close is idempotent, unblocks I/O and joins owned adapter work. A bounded write
  queue cannot let a peer that stops reading freeze cancellation or shutdown.
  Include a request-scoped abandon operation in the adapter contract: stdio and
  the proposed message tunnel send the cancellation notification; a later
  request-stream binding may terminate its own stream. Codec never assumes that
  cancellation means closing stdin or the whole connection.
- Proposed bounds: 8 MiB per encoded message; 64 pending RPCs per connection;
  at most 64 discovery pages, 1,024 tools and 16 MiB total validated descriptor
  data; each opaque cursor at most 4 KiB. Detect a repeated cursor and duplicate
  remote name across pages. No partial catalog is published on any failure.
- Proposed time bounds: discovery 10 seconds per operation and 30 seconds for the
  complete enumeration; tool-call absolute deadline 120 seconds with explicit
  caller configuration bounded to 1–600 seconds; progress never extends it.
  Stdio close first closes input, then waits up to one second, terminates,
  waits one more second, then kills and joins. Supported process platforms must
  match the preceding macOS/Linux lifecycle teaching; portable custom transport
  must not inherit POSIX assumptions.
- Send stdio `notifications/cancelled` for a still-pending request on deadline or
  local cancellation. A remote server may already have completed an effect.
  Mark local state truthfully; neither notification delivery nor process exit
  proves external rollback. Custom adapters implement the same logical abandon
  operation without hard-coded stdin access in Codec.

The full draft must distinguish JSON-RPC protocol errors, MCP `isError:true`
tool results, local timeout/closed errors and malformed results. No receiver
turns a 200-equivalent transport success into a successful tool effect by itself.
Unknown optional metadata is not granted authority; unsupported required result
modes must produce a visible tool failure rather than disappearing.

## Discovery, aliases and current authority

Proposed preparation sequence:

1. The trusted application deliberately configures named connections and launches
   a server or constructs another transport. A skill body/tool result cannot
   choose a command, endpoint or environment.
2. The service collects and validates a complete descriptor candidate. Peer
   self-reported names are display data, not unique routing or trust identities.
3. The caller selects explicit bindings `(local alias, connection name, remote
   tool name)` before Agent construction. Local aliases retain Chapter 9's
   `[a-z][a-z0-9_]{0,63}` grammar; remote names are separate opaque identifiers.
   Reject duplicate aliases and collisions with compiled handlers. Never silently
   truncate/prefix remote names into collisions.
4. Copy the selected definition into that Agent's installed ceiling. In skill
   mode the active grant union still determines declarations and dispatch. In
   plain mode apply the preceding explicit visible-set rules. A server offering
   extra tools or a future credential widening its list changes neither set.
5. The connection's disappearance causes calls through that binding to fail
   unavailable. It does not edit the Agent's committed skill history or pretend
   those grants were unloaded. Explicit reconnect rediscovers and validates the
   selected definitions before use; mismatches refuse instead of hot-swapping.

No automatic server instructions, annotations, `ephemeral` fields or list-change
notifications append instructions, request tools or extend the ceiling. Declared
capabilities are descriptions of the peer, not permission from the application.
A default example must use the same selection path as an external public caller;
an isolated grader-only path cannot conceal a shipped configuration that does
nothing.

Schema candidate: support JSON Schema 2020-12 with bounded validation and local
references, refusing unsupported dialects or unresolved external references.
Never fetch a `$ref` over the network while preparing a handler. A real validator
is preferable to silently treating unknown keywords as permission. The chosen
schema/structured-output subset and precise rejection behavior need review
before a library dependency or fixture assumes them.

## Calls, data and persistence candidates

An ordinary MCP call is an admitted Job. Allocate its existing artifact before
starting remote work, consume pending limits once, and preserve call/result
pairing. Wait/report limits are distinct from the RPC's absolute execution
deadline. `send_input` has no stdin semantics for an RPC Job and returns the
existing unsupported-operation form. Killing this Job abandons its own RPC;
it cannot cancel another Agent's call or kill their shared server process.

Proposed result subset: text plus structured JSON, with an explicit visible
unsupported-content error for images/audio/resource blocks instead of silent
loss. Never dereference returned resource links. Preserve text ordering and
`isError`; keep raw peer `_meta`, annotations and connection credentials out
of the model projection. The full draft must specify exact artifact and report
bytes, structured-value normalization, optional output-schema validation and
whether mixed supported/unsupported content refuses the whole result. That
choice is still open rather than hidden inside a grader.

The call's accepted result is historical data. Offline replay, request
reconstruction and imported snapshot inspection perform no discovery, RPC,
subprocess start or credential lookup. Chapter 10 unfinished-boundary refusal
still applies after an interrupted persistence attempt. Resuming never restores
RPC waiters, remote tasks, subprocess ownership or permission to use stale
server-side handles merely because their text was recorded.

Session identity is a real design question: Chapter 10 currently binds installed
local handler definitions, while handler code and current route are supplied by
the caller. Decide whether remote alias→logical-connection→remote-name mapping
must additionally be creation identity, and publish any format version change.
Do not silently add a field to Chapter 10's strict identity object. Recommendation:
bind stable logical routing and definition semantics, keep resolved executable
paths/current credentials out, and permit same definitions at a deliberately
reselected physical location only under that published rule. This needs
coordinator review before the full draft.

## Proposed actual-use plan

No result below has been observed. Establish a bounded provider plan before the
student spends on it, with the usual source/binary/support bindings and original
failed outcomes retained.

| Surface/action | Concrete receipt to require |
|---|---|
| Human CLI on all three APIs asks a separately launched stdio server to look up a scratch catalog entry, then append a bounded note | Actual PTY prompts/answers; remote wire name/arguments; resulting note bytes; ordinary Job/artifact/result; normalized provider usage; unchanged Agent binary across an external server change |
| Human uses a slow remote tool, then its Job handle to wait/cancel | Running report before finish, exact cancellation request ID, later-call usability, unchanged unrelated call, no remote rollback claim |
| Load/unload an ordinary skill granting a preinstalled remote alias | Same declaration/admission ceiling; forced post-unload call locally refused without remote send; earlier admitted Job's lifetime retained |
| Ordinary browser uses that same Agent and external tool | Real cards, pause/interruption responsiveness, reconnect without duplicate remote effect, actual screenshot with alt text; no claim that GUI self-observation is implemented yet |
| Public headless consumer creates two Agents with different aliases/grants on one root service | Real model tasks with isolated effects; one Agent close/cancel leaves the other usable; safe state copies cannot mutate owner data |
| Public external consumer supplies a custom complete-message transport | Same version/discovery/call/cancel/close behavior with no process or line reader; separate module imports only public packages; all three vendor continuations reconstruct |
| Restart/inspection | Same definition/routing policy enforced; recorded results replay with endpoint disabled and zero new calls; no old Job/RPC becomes live |

Local fault fixtures, not paid requests, cover malformed frames, timeouts at
precise races, partial writes, unsupported content, exhausted IDs, capacity,
slow receiver and process-leak controls. The actual alternate transport is a
public integration exercise; label an in-memory server as such, not as an
independent third-party service. Use a useful scratch-file stdio utility for the
real UI path, with only its intended child environment and no provider key
passed to the tool process by default.

## Acceptance and inherited limitations

The future TL;DR must publish an independent Chapter 11 command before student
release. Retain `make grade-dir CH=12 DIR=solutions/edition-2/main` as a historical
diagnostic; it assumes an old binary/wire/module layout and an older protocol.
Its 35 tool-call/ephemera points can pass from discovery alone. Its WebSocket
check searches source vocabulary. Neither proves an external effect or a
replaceable transport. Do not weaken or rewrite the old checks to claim parity.

The new matrix must include:

- Same complete-message suite on stdio and memory transports; independent public
  adapter; parser/correlation code unchanged when transport changes; intended
  mutation that bypasses the adapter fails alternative-transport proof.
- Out-of-order responses, cancellation/response races, EOF, duplicate/late IDs,
  capacity and writer stalls, followed by real close/join and no leaked process.
- Full pagination and failures before publication; explicit alias collisions;
  declaration filtering plus a distinct forced unauthorized-call/no-effect test.
- Remote success, `isError`, protocol failure, supported/unsupported result
  content and output-size boundaries with readable retained artifacts.
- Skills/policy/hints/interrupt/Jobs/public multi-Agent behavior through existing
  owners; terminal persistence failure cannot retry an already attempted effect.
- Historical replay and session compatibility without automatic RPCs; exact
  numeric values survive common/persistence/browser projections where introduced.
- All affected delivered modules, public examples, star imports, actual parent
  routes, race tests, inherited checks and deletion controls with passing positives.

## Decisions to settle before the full chapter

1. Approve Ensemble-owned service/connections plus child transport constructor
   contract, including shared-connection Agent close and the future GUI logical
   endpoint versus physical socket split.
2. Approve external preparation/explicit alias installation, keeping Chapter 9
   skills declarative, and moving automatic context/skill-triggered server startup
   beside the actual Chapter 12 GUI consumer. No parse-success/no-connect path.
3. Choose persisted remote-binding identity and exact compatibility/versioning
   behavior under Chapter 10, including when discovery may occur around live resume.
4. Choose the bounded schema/result subset and exact loss/error policy. Confirm
   proposed resource/time limits before checker authors encode them.

Protocol version and transport independence are already settled above. These
remaining choices affect ownership or observable behavior, not merely method
spelling. Bill's new first-edition sandboxing work remains untouched and pending;
this outline neither claims it was reviewed nor changes the map around it.
