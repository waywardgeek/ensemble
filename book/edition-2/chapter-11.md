# Chapter 11: Tools across a boundary

Every tool the Agent has used so far was compiled into the binary. Adding a new
one means writing Go, rebuilding and restarting. But a project catalog, a test
service or a small utility already has an owner and a working interface.
Rewriting it as an Agent package buys another copy to maintain. MCP lets the
Agent call that program through a published tool description instead.

The inconvenient part arrives after discovery. Two Agents may share the same
program. One cancels a request while the other waits for an answer. The program
may finish an edit and lose its connection before returning the receipt. A useful
integration has to explain those outcomes without freezing the conversation or
pretending the edit never happened.

The chapter traces one external tool from discovery to an ordinary Job and back
to a readable result. Along the way, the transport becomes a seam: permission,
correlation and cancellation stay the same whether messages travel over stdin or
a WebSocket. The next chapter uses that seam to give the Agent access to its own
GUI.

*Contract draft. The predecessor chapter must be accepted before student
handoff. No Chapter 11 implementation is claimed yet.*

## TL;DR

Read the full [coding skill](skills/ensemble-coding/SKILL.md) before coding,
plus [architecture](architecture.md). Extend the accepted Chapter 10 source.

1. Implement the explicitly limited **MCP 2026-07-28 tool client** below. Require
   per-request metadata, server/discover and complete paginated tools/list before
   publishing a connection. No legacy fallback or automatic effect retry.
2. Ensemble owns MCP service, service owns Connections, and each Connection owns
   its complete-message transport. Constructors receive actual parents. Common
   declares shared values/interfaces; behavior stays in responsible spokes.
   Expose public custom-transport construction. Implement stdio and an in-memory
   message adapter; prove the same protocol behavior through both.
3. Explicitly prepare connections and install frozen per-Agent aliases before
   Agent construction. Discovery grants nothing. Ordinary visibility, Skills,
   handler ceilings and dispatch checks govern remote aliases. Agent close
   cancels its calls; Ensemble owns shared connection shutdown.
4. Keep framing in adapters and RPC parsing, IDs, pending state and cancellation
   decisions in Connection. Use canonical rpc-N string IDs, bounded pending work,
   absolute deadlines and close that unblocks workers even when the peer stops.
5. Validate the exact schema profile in §11.5 and the whole result in §11.7.
   Preserve text order and structured JSON, reject unsupported mixed content,
   and propagate isError. Remote calls use Jobs, artifacts, tool_limits and the
   Actor's durable event path. Interrupt retains its earlier meaning.
6. Publish strict version-2 session identity for MCP bindings under §11.8.
   Keep plain Chapter 10 version-1 stores compatible. Validate stored history
   before remote preparation; replay never invokes an external tool. Current
   routes and credentials remain caller inputs.
7. Provide human CLI configuration/status, ordinary browser cards, public safe
   state and independent public embedding. Exercise all three model APIs through
   the actual interfaces. Retain initial attempts before comparative review.

Method names, private types and choice of bounded schema library remain student
decisions. The JSON formats, authority, lifecycle and literal result bytes here
are observable requirements. Document the public API and semantic snapshot codec.

From `solutions/edition-2/main/`:

```sh
go build -o /tmp/ensemble-ch11-cli ./cmd
go vet ./...
go test ./... -count=1
```

Repeat vet/tests in every affected module, including optional GUI and public
consumers; require empty gofmt output. Run the Chapter 11 independent checker
once its command is published. From the repository root, the inherited
`make grade-dir CH=12 DIR=solutions/edition-2/main` is a historical diagnostic,
not the new acceptance gate. Its protocol and module assumptions differ.
Section 11.10 defines the required distinguishing checks.

## 11.1 An advertised tool has no authority yet

A server can return a thousand tool names. The application may want two of them.
Installing every discovery result would let a server update change the Agent's
powers the next time a connection opens. Put a deliberate selection between the
server's catalog and the Agent's Registry.

The application chooses a logical connection, a remote name and a local alias.
For example, connection `catalog`, remote tool `notes.append` and local alias
`append_note` identify different things. The remote name routes the RPC; the
alias is what the model sees and what a skill may grant. Server-reported names
and titles never supply routing identity.

Use the existing local alias grammar, `[a-z][a-z0-9_]{0,63}`, for aliases and
logical connection keys. This chapter accepts remote names matching
`[A-Za-z0-9_.-]{1,128}`. That is a deliberate interoperable subset, not a claim
that every possible remote name fits. Reject duplicate keys, duplicate aliases,
collisions with installed local handlers and missing selections before exposing
an Agent. Two remote tools with the same name on different connections remain
distinct. Multiple aliases for one remote tool are allowed when selected
explicitly; every alias remains a separate grant.

Freeze each selected definition into the Agent's creation configuration. Its
local handler has the alias, the remote description and the validated input
schema. Keep the validated output schema and logical routing in the associated
binding. The installed ceiling is the selected local handlers plus all explicitly
bound remote aliases, within Chapter 10's total handler bound. Plain mode makes
that installed set visible; skill mode applies the existing active grant union
and mandatory management pair. No MCP callback can mutate Registry or Skills. A skill's tools list can
name a preinstalled alias, using Chapter 9's ordinary closure and ceiling rules;
unload revokes future admission while an already admitted Job keeps its lifetime.
Do not add `mcp_servers` frontmatter in this chapter.

The first edition found a subtle trap here: a configuration that parsed cleanly
on a primary skill but never actually connected, because startup depended on a
load callback that ran later. This edition avoids it by making preparation
explicit. It finishes before construction succeeds. A configured tool either
has a validated connection or gives the caller a useful failure. Automatic
context collection and skill-triggered connection lifetime belong with the
actual GUI consumer in Chapter 12.

A remote description is untrusted tool documentation. A result is untrusted tool
data. Neither can append system instructions, install tools, load skills, read
credentials or invoke an Agent handler by sending a message back. The current
protocol's tool-client direction is sufficient for this exercise; there is no
default server-to-client execution authority.

## 11.2 Give the connection an owner

Closing one Agent must not kill a shared catalog process. Leaving that process
unowned until the application exits is equally unhelpful in a reusable library.
The application root owns the shared resource; each Agent owns its selected
bindings and its admitted work.

| Object | Owns | Actual parent |
|---|---|---|
| Ensemble | MCP service and application shutdown | Application logger lives here |
| MCP service | Named connections, generations and explicit prepare/reopen/close | Ensemble |
| Connection | Protocol state, discovered definitions, ID cursor, bounded pending operations and transport | MCP service |
| Transport | Logical endpoint, framing and its own I/O workers/resources | Connection |
| Agent | Frozen bindings, Registry, Skills, Jobs and Actor | Ensemble |
| Job | Remote-call association, artifact and terminal/report state | Jobs, then Agent |

A parser or timer helper split into an object receives its responsible owner,
with a common parent interface providing the whole route to logging. Tools
reaches MCP through Job/Jobs/Agent/Ensemble; MCP never imports the Tools
implementation. Observers publish ordinary Agent facts. They are not an RPC
return channel or a hidden second conversation authority.

The service accepts a public transport-constructor interface. Its construction
method receives the actual new Connection parent and returns a transport child.
An external package can implement that interface using only public declarations,
including diagnostic access through its parent. Avoid a collection of unrelated
Send, Receive and Log closures. A concrete constructor may carry its immutable
configuration and an application-owned endpoint resource; it cannot replace the
required Connection parent with that resource.

Expose explicit prepare and close operations, plus immutable connection and
Agent-binding snapshots. Snapshot fields include logical key, connection state
(`preparing`, `ready`, `closed` or `failed`), positive generation, protocol version,
discovered count, selected alias/remote name and safe error code. Generation is
uint64, allocated before creating each transport and never reused within that
service; refuse exhaustion before mutation. Copies cannot alter live maps,
schemas or pending work. Browser projections preserve exact newly introduced
numeric identities as Chapter 8 teaches for counters.

Preparation happens off the Actor. A service may retain a failed connection's
safe diagnosis, but it publishes no usable partial discovery. Explicit reopen
creates a new generation after the old transport has closed. Existing Agent
bindings become usable only if their selected definitions still match exactly.
Check all bindings attached to that logical connection before publishing the
replacement ready state. Changed unselected tools may appear in discovery; they
cannot enter a frozen Agent ceiling.

An Agent close cancels only operations associated with that Agent, then joins
its owned workers through existing shutdown. The connection remains available
to other Agents. Service close first ends new admission, cancels all outstanding
operations, closes transports and joins workers. These operations are idempotent.
A transport failure naturally affects everyone sharing that transport; report
that failure rather than attributing it to another Agent's authority.

## 11.3 Put messages across the seam

Protocol code receives one complete owned UTF-8 JSON message at a time. It knows
nothing about stdin, lines, frames or browser sockets. A transport sends one
complete owned message and receives complete messages; neither operation retains
caller-mutable byte slices. The API must support cancellation, bounded I/O and
close that wakes a blocked sender or receiver. State which side owns each worker
and prove that close joins it.

Outgoing request delivery includes its opaque correlation ID as transport data.
That lets a future request-stream adapter associate a stream with a request
without parsing MCP JSON. The Connection constructs every protocol envelope.
For request-specific abandonment, it supplies the ID and an already encoded
cancellation notification. A shared-channel adapter delivers those bytes; an
adapter with a dedicated response stream can close that stream instead. The
adapter chooses its delivery mechanism, never the protocol reason or the pending
request to cancel. This chapter's stdio and memory adapters both deliver the
provided notification.

For stdio, launch the caller-selected executable directly with an argv array,
without a shell. The adapter adds one LF to each compact outgoing message and
removes the framing LF from incoming messages. Reject a nonempty EOF fragment;
an empty EOF closes the connection. Actual LF bytes inside a JSON string are
invalid JSON; escaped `\n` remains part of the message. Reject invalid UTF-8,
empty lines and overlong lines. Drain stderr separately so diagnostics cannot
block stdout, and never interpret stderr as tool output. These framing choices
follow the [stdio binding](https://modelcontextprotocol.io/specification/2026-07-28/basic/transports/stdio).

The in-memory adapter transfers complete messages without adding or removing
bytes. Run the same discovery, call, out-of-order reply, error, cancellation and
shutdown tests through both adapters. A separate public consumer must implement
its own message transport without importing internals or relying on a process.
The purpose is to discover a hard-coded pipe assumption before the GUI depends
on that assumption being absent.

On supported macOS/Linux stdio, close stdin, allow one second for normal exit,
then terminate the owned process group, allow one more second, and kill/join
remaining owned processes and pipe workers. Root close can force this sequence;
a single Job cancellation cannot. Document unsupported process platforms
explicitly. A read or write must be interruptible by transport close, including
when the peer has stopped reading. Use one bounded delivery worker per connection,
with request/cancellation state held by the 64 operation permits described below.
Allow at most one second from staging a message or cancellation notice through
completed delivery, including queue time. A stall faults and closes the connection
instead of keeping a shared writer stuck forever. Root close aborts staged work
and closes the endpoint without waiting for that queue to drain.

The future WebSocket adapter belongs in the optional GUI module. Its resource is
a logical MCP channel on a GUI-owned socket. Closing that endpoint must unblock
its MCP workers without closing unrelated channels, Pages or the physical socket.
The public constructor and alternative transport proof are required now. The
actual tunnel and GUI observation/control are required Chapter 12 work.

## 11.4 Choose a protocol version, then discover

As checked on October 8, 2026, the official specification identifies
2026-07-28 as current. That revision puts version and client capabilities on each
request. The initialize/initialized exchange belongs to earlier revisions.
This chapter deliberately implements a current, limited tool client and refuses
an incompatible server without silently retrying a different protocol.
[Versioning](https://modelcontextprotocol.io/docs/2026-07-28/learn/versioning)
records the distinction.

Every request has jsonrpc `2.0`, a canonical string ID, a method and object params
with these metadata entries. Use the fixed client information shown here; other
required request fixtures have the same metadata, even when abbreviated in an
explanation. The first request is:

```json
{"jsonrpc":"2.0","id":"rpc-1","method":"server/discover","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28","io.modelcontextprotocol/clientCapabilities":{},"io.modelcontextprotocol/clientInfo":{"name":"Ensemble","version":"edition-2-ch11"}}}}
```

A minimal accepted response is:

```json
{"jsonrpc":"2.0","id":"rpc-1","result":{"resultType":"complete","supportedVersions":["2026-07-28"],"capabilities":{"tools":{}}}}
```

Require supportedVersions to be a nonempty array of unique strings containing
the selected version, and require an object tools capability.
The client makes discovery mandatory for its own preparation even though the
protocol permits a client to skip that probe. Ignore instructions, cache hints,
serverInfo and additional capabilities as authority. No discovery text becomes
model context. See the current
[server/discover method](https://modelcontextprotocol.io/specification/2026-07-28/server/discover).

Next send tools/list with the same metadata and no cursor. Each reply has a tools
array; a present nextCursor must be a nonempty string. Send it back unchanged in
the next request's cursor. Missing nextCursor ends the list. Refuse a repeated
cursor, duplicate remote name across pages, malformed descriptor, unsupported
schema or limit violation. The reader should never see an apparently ready
append_note while a later discovery page can still invalidate its catalog.
Do not expose page one while page two is still being validated. A failed
preparation closes its transport and leaves no selected handler executable.

For a single-page catalog, the request and response are:

```json
{"jsonrpc":"2.0","id":"rpc-2","method":"tools/list","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28","io.modelcontextprotocol/clientCapabilities":{},"io.modelcontextprotocol/clientInfo":{"name":"Ensemble","version":"edition-2-ch11"}}}}
```

```json
{"jsonrpc":"2.0","id":"rpc-2","result":{"resultType":"complete","tools":[{"name":"notes.append","description":"Append a note to the selected scratch notebook.","inputSchema":{"type":"object","properties":{"text":{"type":"string","maxLength":2000}},"required":["text"],"additionalProperties":false},"outputSchema":{"type":"object","properties":{"written":{"type":"integer","minimum":0}},"required":["written"],"additionalProperties":false}}]}}
```

Description is optional and normalizes to the empty string; name and inputSchema
are required. Input root must explicitly have type object. Output schema is
optional. Descriptive title/icons/annotations and `_meta` are ignored; other
unknown descriptor fields refuse this bounded profile. Definition equality uses
exact remote name, normalized description, complete input schema and either the
complete output schema or an explicit absence. Schema annotations retained under
§11.5 participate in equality. Chapter 10's lossless semantic JSON canonicalization
compares them; map order and equivalent number spellings do not create a change.

Parse envelopes strictly: one JSON object, no duplicate keys, no trailing data,
valid UTF-8, bounded nesting and exactly one response result or error. Reject
unpaired surrogate escapes rather than silently replacing their text. Responses
require the exact issued ID. Result must be an object; absent resultType means
complete, and only complete is supported. An input_required result produces a
visible unsupported-result failure without a follow-up exchange. An unknown
resultType or malformed envelope faults the connection. A well-formed JSON-RPC
error has a numeric code with an exact mathematical integer value in signed
int32 (-2,147,483,648 through 2,147,483,647), string message and optional data.
Validate that bound losslessly before formatting the short decimal code; never
expand an arbitrary exponent just to print it. An out-of-domain code is a
protocol error. Expose the valid code with a safe local explanation, excluding
arbitrary remote error data. Reject
unknown top-level envelope members; the permitted request/notification members
are jsonrpc, method, params and optional id, and response members are jsonrpc, id
and exactly one of result/error. Optional error data and known result metadata
remain bounded JSON, even when excluded from the model projection.

A server request with method and id faults the connection without dispatching
anything or responding with an Agent result. Well-formed notifications without
an id are ignored within the message bounds; they cannot alter descriptors,
extend deadlines, grant tools or append context. In particular, this chapter
neither subscribes to list changes nor follows a peer's resource links.

## 11.5 Validate the schema the Agent will actually use

A schema that gets displayed but never checked makes the external boundary
unreliable. Conversely, fetching a remote `$ref` while validating a tool would
turn discovery into another network client with its own hidden credentials and
failure modes. Keep validation local and bounded.

This is a documented profile of JSON Schema 2020-12, not a full implementation
claim. The ordinary keyword meanings come from the 2020-12
[validation vocabulary](https://json-schema.org/draft/2020-12/json-schema-validation)
and [core specification](https://json-schema.org/draft/2020-12/json-schema-core).
A real validator library is allowed, provided preparation rejects features
outside this profile and disables external resolution. Schemas are objects or
booleans except the required object input root. Support exactly these keywords:

| Keywords | Required meaning and shape |
|---|---|
| `$schema` | Optional exact `https://json-schema.org/draft/2020-12/schema`; omission selects that dialect |
| `$defs`, `$ref` | Object of schemas; local `#` or `#/...` JSON Pointer reference only, resolving inside this same root |
| `type` | One type name or nonempty unique array from object, array, string, number, integer, boolean, null |
| `properties`, `required`, `additionalProperties` | Object of schemas, unique string array, boolean/schema; absent additionalProperties permits extras |
| `items`, `minItems`, `maxItems` | One schema and nonnegative integer bounds |
| `minLength`, `maxLength` | Nonnegative integer bounds in Unicode scalar values |
| `minimum`, `maximum`, `exclusiveMinimum`, `exclusiveMaximum` | Exact JSON numeric bounds; integer accepts mathematically integral JSON numbers |
| `minProperties`, `maxProperties` | Nonnegative integer bounds |
| `enum`, `const` | Nonempty array of unique semantic JSON values, or any JSON value |
| `allOf`, `anyOf`, `oneOf`, `not` | Nonempty schema arrays or one schema; ordinary boolean validation semantics |
| `title`, `description`, `default`, `examples` | String, string, any JSON value, array; annotations only, no insertion or execution |

Reject every other keyword, including pattern, format, dynamic references,
external/relative references and unevaluated vocabulary. Reject malformed keyword
values, unresolved pointers and cycles reached by traversing schema containment
and reference expansion together. For example, a child property with `{"$ref":"#"}`
returns to its containing root and is refused. `$ref` siblings apply too. Enforce the bounds below on the physical schema and on
validation work so repeated references cannot create exponential work unnoticed.
Validation failure or budget exhaustion cannot fall through to sending a call.
Numeric comparisons and enum equality retain arbitrary accepted JSON number
precision; binary64 conversion cannot merge distinct values. Apply Chapter 10's
bounded lossless number-token rules and compare coefficient/exponent forms
without expanding a huge exponent into an equally huge allocation.

Before a remote send, validate the complete arguments object against the frozen
input schema through the normal tool-admission path. For the descriptor above,
`{"text":"checked"}` passes, while `{"text":7}` and
`{"text":"checked","path":"other"}` fail locally with zero RPCs. Do not
silently strip the extra member or fill default values. A connection being ready
does not exempt a forced call from the current Agent's visibility check.

Output schema validation uses the same profile. If one was declared, require
structuredContent and validate it, including for an isError result. This course
choice is intentionally strict: a server whose error results violate its declared
output shape is incompatible with the selected definition. Without an output
schema, structuredContent may be any bounded JSON value, including null.

These are maximum accepted resource sizes, with exact-boundary acceptance and
one-over refusal. They apply independently; a smaller inherited call/log limit
still applies before allocating or publishing the corresponding value.

| Resource | Limit |
|---|---|
| Configured connections per root / installed remote aliases per Agent | 32 / 1,024 |
| Incoming or outgoing complete message | 8 MiB, excluding transport delimiter |
| Parsed JSON nesting / total value nodes per message | 64 / 100,000 |
| Admitted remote operations per connection, including staged delivery/cancellation | 64; excess refuses immediately |
| Discovery pages / remote definitions / accumulated canonical retained-definition bytes | 64 / 1,024 / 16 MiB |
| Opaque cursor / description / each schema | 4 KiB / 64 KiB / 256 KiB UTF-8 bytes |
| Physical schema nodes / reference-expanded nesting / validation steps | 4,096 / 64 / 100,000 |
| Result text blocks | 1,024, within the message limit |
| Retained stderr tail per connection | 16 KiB; older bytes discarded with a truncation indication |
| Discovery request / entire preparation | 10 seconds / 30 seconds |
| Remote call absolute deadline | Default 120 seconds; caller-selected integer seconds 1–600 |

A schema node is each schema object/boolean visited during compilation, counting
shared physical targets once. Validation steps count each schema/value pair
visit, including repeated reference/composition visits. Generic JSON nodes count
each object, array and scalar, including schema annotations. Check limits while
reading/compiling, before unbounded allocation. Calls do not grow a second,
unbounded pre-send queue; capacity includes work admitted for delivery. All stated
byte counts use powers of 1024. Generic nesting counts the outer object/array as
level 1.

Measure each schema's 256 KiB bound on its Chapter 10 canonical JSON encoding.
An absent output schema consumes zero schema bytes; a boolean schema consumes
its actual encoded bytes. For discovery's 16 MiB total, normalize each definition
to exactly name, description, inputSchema and outputSchema, using the empty
description and null absent output described above. Sum the canonical byte length
of each such object, with no surrounding array brackets or separators. Ignored
descriptor metadata is absent from this sum; retained schema annotations count.
The 8 MiB received-message bound includes ignored metadata and original wire
whitespace. The generic-node bound includes ignored metadata values, but whitespace
is not a JSON node. Discarding metadata cannot evade either input bound.

## 11.6 A remote call is still a Job

The reader sees append_note in the same tool list as a local operation. Give it
the same admission checks, durable tool_called event, exclusive artifact creation,
limits and initial report behavior. Only the handler's execution crosses the
connection boundary. An artifact or pre-dispatch append failure must prevent the
RPC exactly as it prevents a local process start.

After admission, a worker reaches the named Connection through its actual owner
chain. The Actor never waits for discovery, a pipe write or a remote answer.
A later response updates the owned Job and commits ordinary terminal facts through
the Actor. Tool limits consume once at the Chapter 10 next-attempt boundary; they
control reports, not the maximum time an external server may run. A short report
wait can return running while the absolute RPC deadline remains in force.
Start that deadline when the worker begins the admitted remote attempt, before
transport delivery; writing, waiting and progress notifications cannot reset it.

Suppose discovery used rpc-1 and rpc-2. The call is:

```json
{"jsonrpc":"2.0","id":"rpc-3","method":"tools/call","params":{"name":"notes.append","arguments":{"text":"checked"},"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28","io.modelcontextprotocol/clientCapabilities":{},"io.modelcontextprotocol/clientInfo":{"name":"Ensemble","version":"edition-2-ch11"}}}}
```

Each remote attempt first acquires one of 64 shared operation permits. The
permit covers its pending association, any staged request and any staged
cancellation. Removing pending correlation does not free the permit while
delivery is unfinished. If cancellation wins before issue, discard the unsent
request and release the permit without allocating an ID or sending a notice.
After issue, retain the permit until cancellation delivery completes or the
transport closes. A normal response releases it only after its request delivery
has also completed. Use bounded per-operation flags/records with the single
writer, not a goroutine or new queue entry for every cancellation. Thus a stream
of canceled calls cannot bypass admission while earlier notices wait to be sent.

Connection IDs are canonical `rpc-N`, with N a positive base-10 uint64, no sign
or leading zero. Allocate the next ordinal only at the irrevocable local issue
boundary, immediately before handing an owned request to the transport; insert
its pending association before any reply can arrive. Merely waiting for capacity
or failing local validation does not reserve an ID. Once issuance starts, burn
the ID even if delivery fails before the peer receives a byte. “Issued” means
attempted transport delivery, not proof of remote receipt. Refuse overflow before
allocation; a caller may explicitly reopen a new generation.

Keep an issued high-watermark and a pending map, rather than a set of every
canceled ID. A valid response for a pending ID settles that operation once.
A canonical ordinal at or below the issued watermark with no pending entry is
stale, canceled or duplicate: discard it without state change. An ordinal above
the watermark, a noncanonical spelling, a number ID or a different namespace is
unknown and faults this connection. Every receive worker belongs to one generation;
late work from a closed generation cannot enter a replacement's pending map.
All valid IDs through the watermark were issued under this rule, including
failed deliveries. There are no secretly reserved holes to classify.

A cancellation and a response may race. The Connection chooses one local
settlement and removes the pending association once; Jobs retains its own
terminal-event ordering. A late result cannot replace a killed status or append
a second result. Keep cancellation bookkeeping bounded by the 64 admitted operation permits,
even after millions of sequential canceled calls. Logging stale
replies must be bounded too; retain aggregate counts rather than every payload.

For a still-pending rpc-3, cancellation bytes are exactly this compact message:

```json
{"jsonrpc":"2.0","method":"notifications/cancelled","params":{"requestId":"rpc-3","reason":"Cancelled by client"}}
```

Use the same safe reason for user kill, deadline and owner close; local error/status
records distinguish those causes. A response that already won produces no
cancellation notification. On a healthy shared channel, killing one Job sends
only that request's notice and leaves another Agent's call alive. It does not
terminate the stdio process. A stalled cancellation write can fault the shared
transport under §11.3; all affected calls then report transport loss.
The [cancellation specification](https://modelcontextprotocol.io/specification/2026-07-28/basic/patterns/cancellation)
explains why advisory cancellation cannot undo a completed external effect.

An explicit kill produces the inherited killed lifecycle; deadline, capacity,
remote/protocol failure and transport loss complete with done plus is_error.
`wait` reports retained bytes and terminal state; `send_input` refuses for an MCP
Job. Chapter 5 interruption stops model continuation without killing admitted
Jobs. EOF, timeout or loss after sending must say that the external outcome may
be unknown. Never automatically resend tools/call after reconnect, even when
its arguments appear harmless.

## 11.7 Keep the whole accepted result

A tool can return text and structured data together. Flattening one into the
other loses information; silently dropping an unsupported image can make a
partial answer look complete. This client accepts an ordered array of text
blocks and an optional structured JSON value. It rejects the whole result if any
content block has another type. No URI is fetched to repair that refusal.

The current tool specification distinguishes execution errors from JSON-RPC
errors and allows any JSON value in structuredContent. This chapter's accepted
content types are narrower than that protocol's full set.
[Tool results](https://modelcontextprotocol.io/specification/2026-07-28/server/tools)
define those wire meanings.

A complete call result requires content as an array. Each accepted item has type
text and a string text; optional annotations and `_meta` are ignored. Other item
fields refuse. Result fields may be resultType, content, structuredContent,
isError and `_meta`; unknown fields refuse. Absent isError means false; a present
value must be boolean. Validate the entire result and output schema before
writing accepted bytes. An empty content array is valid, and a present null
structuredContent differs from absence.

This response to the example call passes:

```json
{"jsonrpc":"2.0","id":"rpc-3","result":{"resultType":"complete","content":[{"type":"text","text":"Appended.\n"}],"structuredContent":{"written":7},"isError":false,"_meta":{"diagnostic":"not model context"}}}
```

The artifact is exactly the following compact canonical JSON followed by one LF:

```text
{"content":[{"text":"Appended.\n","type":"text"}],"isError":false,"structuredContent":{"written":7}}
```

Use Chapter 10's canonical JSON rules, including lossless numbers. Always include
content and isError; include structuredContent exactly when present. Block order
and text values are preserved. Exclude resultType, `_meta` and annotations.
The ordinary ToolResult is text containing these bytes, subject to inherited
report limits; it does not add a second copy of structuredContent to a provider
request. All three renderers receive the same recorded text and their existing
tool-call association.

The reader can recover the whole accepted object from cr/io/HANDLE even when
the model receives a capped report. The JSON wrapper makes an empty text block,
a missing structured value and an explicit null distinguishable without
guessing from a line of prose.

For an output-schema-free tool, `{"content":[],"structuredContent":null}`
produces `{"content":[],"isError":false,"structuredContent":null}` plus LF.
A content array containing text followed by an image produces no accepted text
prefix. It writes only the safe local failure record below and marks the Job's
execution as an error. A result with isError true retains its accepted content
and structured value while setting the ordinary tool/job error flag. A later
successful wait operation does not erase that historical execution failure.

Local MCP failures produce a compact canonical object plus LF with exactly
error and message, for example:

```text
{"error":"mcp_unsupported_content","message":"Remote result contains unsupported content."}
```

Use stable error codes: mcp_unavailable, mcp_capacity, mcp_timeout,
mcp_transport, mcp_protocol, mcp_remote_error, mcp_invalid_result,
mcp_unsupported_result, mcp_unsupported_content and mcp_limit. Input-schema
failure retains the inherited invalid-arguments disposition and sends nothing.
For remote JSON-RPC errors the safe message includes the decimal remote code,
but never untrusted data or stderr. For transport loss after issue and timeout,
the message must explicitly say the external outcome may be unknown. Exact
English wording beyond the fixture is student choice. Failure records themselves
must fit the ordinary artifact/report path, even after an oversized result.

Do not persist credentials, raw diagnostic envelopes or arbitrary stderr as
conversation material. Safe logs may name connection key, generation, local
alias, remote name, RPC ID and local error code. Stderr is available only through
an explicit application diagnostic operation, with its bounded/truncated status;
it is absent from model, public watch and default CLI output. The application
must still treat the tool's intended text as data that may contain private
project information.

## 11.8 Resume the conversation without reviving the connection

A saved alias called append_note is not enough to identify its behavior. A new
application could bind that name to a different server method with the same
schema. Persist the logical mapping and definition semantics as creation identity,
then require the new caller to supply a compatible selection deliberately.

MCP-enabled sessions use this explicit extension of Chapter 10. Existing formats
are never changed by silently adding a field:

| Store | Checkpoint version / state_version | Identity and initial session payload |
|---|---|---|
| No remote bindings, including existing Chapter 10 stores | 1 / 1 | Exact Chapter 10 identity and payload, unchanged |
| One or more frozen remote bindings | 2 / 2 | Identity adds required nonempty mcp_bindings; initialization/anchor session payload adds version:2 |

The log header remains `{"log_version":1}` and the event envelope is unchanged.
For version 2, session_initialized.session has exactly version, session_id and
identity; version is integer 2. Identity has exactly the old mode, system, skills
and handlers fields plus mcp_bindings. It remains sequence 1, once, construction-only.
A version-2 imported session_anchor.session has exactly version, session_id,
origin_as_of, origin_sha256 and high_watermarks, with version 2 and all inherited
anchor checks. Absence of version in those payloads selects the exact old
Chapter 10 shape; explicit version 1 or any other version refuses. Checkpoint
version, state_version, initializer/anchor and any immutable origin must agree.
All these metadata facts remain construction-only; public append cannot insert
an initializer/anchor, adopt bindings or change a session's format/identity.

The version-2 checkpoint has exactly the Chapter 10 outer fields and uses their
same hash/sequence meanings. Its semantic codec extends required state with the
frozen remote bindings and validates agreement with identity; the student must
document the exact private field spelling before checker integration. No live
Connection, pending RPC, ID cursor, process, credential or transport resource is
semantic conversation state. Snapshot-only import has the same absent-history
limit and cannot claim that it validated a missing prefix.

Each mcp_bindings item has exactly alias, connection, remote_name and definition.
Sort by alias and require uniqueness. Definition has exactly name, description,
inputSchema and outputSchema; outputSchema is null exactly when absent on the
wire. Name equals remote_name. All values satisfy §§11.1/11.5. The corresponding
installed handler in identity.handlers has name equal to alias, the same
description and schema equal to inputSchema. Require this correspondence in
both directions for remote bindings, with no alias collision with a local
handler. A representative item is:

```json
{"alias":"append_note","connection":"catalog","remote_name":"notes.append","definition":{"name":"notes.append","description":"Append a note.","inputSchema":{"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false},"outputSchema":null}}
```

Compare all stored remote bindings, including currently ungranted aliases.
Semantic equality permits different object-member order and equivalent numeric
spellings. A changed connection key, remote name, local alias, description or
schema refuses session_incompatible. A same-content executable at another path
can be deliberately selected; paths, argv, environment values and credentials
are current physical configuration and are excluded from identity. Matching
metadata cannot prove that two executables have the same implementation.

A version-1 store accepts only zero remote bindings on resume. Adding them
requires an explicitly fresh session directory; do not import or rewrite the old
store automatically. A version-2 store requires its exact nonempty selection.
Offline inspection and request reconstruction work without servers or current
MCP configuration, using recorded tool results and declarations. Historical
remote Jobs remain unavailable under Chapter 10, even if a new server happens
to recognize an old handle printed in a result.

Order live resume carefully. Acquire the session reservation/lock, then validate
all stored bytes, version, transitions, hashes, settled boundary and local caller
compatibility before this operation starts a transport or sends discovery. Check
the caller's logical selections against recorded bindings in that phase. Only
then prepare the required current connections and compare their selected
validated definitions. Publish the resumed Agent after both phases succeed.
A failed second phase leaves the session unchanged and releases construction
resources; an already shared connection remains owned by the service.

The public API may expose a prepared, immutable resume plan or perform both phases
inside session opening. It must retain the SessionStore/Actor append authority
and actual owner routes. A caller may already have prepared a shared connection
for another Agent; the resume operation still performs no new remote I/O before
validating its stored state. Reuse a ready generation only after checking the
frozen definitions. Historical reduction never sends tools/call, and no protocol
retry reconstructs an unfinished effect.

## 11.9 Taking it for a spin

The implementation and actual runs are pending. The steps here define the user
exercise; they are not a transcript. Preserve failed attempts and replace this
section's planned outcomes with source-bound observations after the student runs
all three model APIs.

Provide a strict CLI/GUI-launcher option `--mcp-config FILE`. The UTF-8 JSON file
is at most 1 MiB, rejects duplicate/unknown fields and has version 1, connections
and bindings exactly. A connection contains key, transport, command, args, cwd,
env_allowlist and call_timeout_seconds; transport is stdio, command/cwd are
nonempty paths, args and env_allowlist are string arrays, and timeout follows
§11.5. Relative paths resolve against the configuration file's directory; command
is a path, not a shell expression or PATH search. Each binding contains exactly
alias, connection and remote_name. Keys/aliases obey §11.1. Example:

```json
{"version":1,"connections":[{"key":"catalog","transport":"stdio","command":"./scratch-catalog","args":["--stdio"],"cwd":".","env_allowlist":[],"call_timeout_seconds":120}],"bindings":[{"alias":"append_note","connection":"catalog","remote_name":"notes.append"}]}
```

Start children with only explicitly allowlisted environment entries, copied from
the caller at preparation. Require unique valid environment names and present
values; never copy provider credentials by default. Other public adapters have
their own typed constructor configuration; this file does not expose arbitrary
plugin loading. Omission selects no MCP connections. Prepare only connections
referenced by selected bindings, in key order; unused entries grant and launch
nothing. Missing referenced connections refuse before starting any process.

Keep Chapter 10's session-directory selector and explicit standalone-log behavior.
Validate existing sessions before preparation, and never redirect CH02_LOG because
an MCP file was supplied. The launcher shares this reader with the CLI. Add
`/mcp` to human chat: show selected aliases, logical keys, ready/unavailable state,
generation and safe errors without argv, paths, environment or stderr. Public
consumers can obtain the same owned safe information. A browser uses ordinary
ToolCall/ToolResult/Job cards and a small binding-status view through public
Agent interfaces; it does not implement MCP or receive diagnostic secrets.

Extend safe WatchState with required `mcp`, an array sorted by alias, empty when
there are no remote bindings. Each item has exactly alias, connection, remote_name,
state, generation, protocol_version and error_code. State uses §11.2's four
values; generation is a positive uint64 or null when there is no current runtime
connection. Protocol_version is `2026-07-28`; error_code is empty or a safe local
code from §11.7. A ready binding appears as:

```json
{"alias":"append_note","connection":"catalog","remote_name":"notes.append","state":"ready","generation":1,"protocol_version":"2026-07-28","error_code":""}
```

Route owned connection-state facts through Ensemble to the bound Agents' Actors,
after releasing service locks. Actor maintains the derived presentation state,
advances its watch revision and publishes `mcp_changed` with agent_id and the
complete safe mcp array. This is a transient observation, never a durable
conversation event or another grant authority. Watch capture and subsequent
changes share the existing actor boundary; no dropped state fact or late old
generation may leave the replacement displayed as ready. Coalesce identical
projections. Offline watch uses recorded bindings with closed state, null
generation and mcp_unavailable; it neither prepares a connection nor invents
a live generation. The page labels this state and treats generation losslessly
before accepting a frame, using the inherited unsafe-counter refusal policy.

Build a separate useful scratch-catalog server that can look up a note, append
a bounded note and delay a reply. Document its build/start command and its data
file. Extend the example configuration with explicit aliases for that utility's
lookup and delay tools before asking the model to use them; the single binding
above illustrates the file format. The utility can use the same protocol subset,
but must perform the intended file
operation and retain a sanitized server receipt. The server is an exercise
utility, not a fake model endpoint. Keep its source/binary identity separate from
the Agent executable.

In an actual human PTY, ask the model to look up a known note and append a new
one. Read the answer before following up. Inspect the actual notebook bytes and
Job artifact; a model saying “done” is insufficient. Change the external utility's
implementation in a controlled way and repeat with the same Agent binary in a
fresh compatible session, distinguishing metadata compatibility from actual
implementation behavior. Record the exact provider request and usage.

Then request delayed work, observe a running report, and ask for wait and kill
through the ordinary tools. Keep another Agent's call active on the shared
connection and show it completes after the first Agent closes. Load/unload a
skill that grants an installed alias, including a forced unauthorized-call
control with zero remote sends. Verify that ordinary interruption remains
responsive while a remote Job runs.

Use the browser for the same task, capture actual cards/status with an accessible
screenshot description, and reconnect without repeating the effect. The public
headless consumer must use two Agents and the public custom message adapter
with real model requests on all three provider paths. Label the in-memory
external service honestly; it proves the transport seam, not third-party service
compatibility. Checkpoint and restart the offline inspector with endpoints disabled to prove
historical rendering makes zero MCP calls. Separately resume a compatible live
session: discovery may prepare its new connection, but no historical tools/call
is repeated.

Review the complete action/result/provider matrix before spending. Retain exact
source/build bindings, sanitized launches, PTY prompts/answers, RPC IDs and
result bytes, final files, process cleanup, screenshots and normalized usage.
Keep local faults, real-model generations and endpoint-disabled replay separate.
Do not spend paid requests to manufacture malformed JSON or a stopped pipe.

## 11.10 What passing has to establish

The inherited grader maps to old Chapter 12. Its tool-call and ephemeral flags
can become true after discovery alone, and its WebSocket check searches source
strings. Preserve it as historical evidence; those checks cannot establish the
new transport seam or an external effect. Publish the independent new checker
command before a student starts, without giving that student grader internals.

| Required property | Distinguishing acceptance evidence |
|---|---|
| Transport independence | Same protocol suite on stdio/memory; external public adapter; replacing pipe assumptions breaks an intended control |
| Protocol and bounds | Literal metadata/discovery/list/call; pagination; exact/one-over limits; malformed envelopes; unknown/stale IDs; no reverse execution |
| Authority | Two Agents with distinct aliases/grants; forced hidden calls produce zero sends; unload changes future admission only |
| Schema/results | Local refs, precision, composition budget, arguments before send; exact canonical artifacts; isError; unsupported mixed result refused whole |
| Lifecycle | Out-of-order/racing replies, many sequential cancels with bounded memory, peer stops reading, EOF, late generation, healthy peer isolation and joined shutdown |
| Persistence | v1/v2 fixtures, incompatible identities before preparation, snapshot/tail equivalence, no replay effects or restored remote workers |
| Existing behavior | Jobs/report limits, policy, hints, interrupt, Skills material, all three renderer continuations and optional GUI boundary |
| Public usability | Human CLI, browser and independent headless/custom adapter with actual model use and safe immutable snapshots |

Test exact uint64 exhaustion through a disclosed valid owner-state seam where
constructing that many calls is impractical. Keep passing controls for every
mutation, and require the intended failure rather than an unrelated setup error.
No assertion may depend on an unpublished private identifier spelling. Run race
checks for shared connection calls, cancel/close/reopen and snapshot readers.
Preserve inherited tests and the complete delivered-module build boundary.

After the initial implementation and actual run checkpoint, an independent
reviewer compares the corresponding first-edition standard. The student receives
findings and rationale, not old source to copy. Reconcile teaching difficulties,
repeat affected demonstrations, and retain unsuccessful outcomes with their
original identities. A green discovery test does not finish this chapter.

The reader can now attach a tool without moving its implementation into the
Agent. The same owner and transport boundaries will carry the GUI tunnel next:
a logical endpoint can disappear while the rest of the interface keeps working.
