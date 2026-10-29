# Gateway C# library and applications — locally complete

This standalone .NET 10 library does not reference Ensemble's core, CLI, GUI or
ASP.NET. It implements key management, Noise-KK, wire values, the public WebSocket
client, router policy/administration/snapshot queues and authenticated server sessions.
The optional Host executable now owns the HTTP listener and periodic saves; the
independent Interviewer uses only the public client. The optional Adapter connects
to Ensemble, and AgentHost owns that application lifecycle. The source-only Chapter 26
endpoint preserves this reviewed implementation; it is not a deployed service.

## Ownership and module layout

- `Common` declares the persistent vocabulary and callable operations.
- `Cryptography` implements key files and Noise sessions; its pinned third-party
  dependency is documented in `../ThirdParty/Noise/README.md`.
- `Protocol` implements wire encoding/validation and provenance-header formatting.
- `Client` implements connection lifecycle, sequential delivery, and correlated MCP.
- `Router` implements identities/groups, administration, routing, queues and sessions.

`Client` is a standalone root. It owns its socket, cancellation resources, pending
RPCs, and one `NoiseKK`. The child constructor establishes the Client/Noise
relationship in both directions; explicit disposal removes both ends. Public
standalone Noise sessions remain supported. The immutable identity/configuration
and wire records are values, not lifecycle nodes. `IHandler` is the borrowed
application callback boundary; the client never destroys its handler or the
application behind it.

## Using the client

Construct `ClientConfig` with an explicitly selected `ws://` or `wss://` URL,
agent ID, private identity, and pinned gateway public key. `Client.ConnectAsync`
completes only after the fixed KK handshake and an encrypted, identity-matching
welcome. Its cancellation token owns the connection lifetime, not just startup.
There is no private-key discovery, protocol negotiation, implicit reconnect, or
retry. A replacement client is a new owned connection.

Implement `IHandler`, or override selected methods on `NopHandler`. Callbacks run
sequentially. Keep them short: hand accepted work to the application's own queue.
Do not synchronously wait for RPC replies or joined disposal from a receive
callback—the same receive loop must read the reply or finish that callback.
`Close()` may be called from a callback; it interrupts I/O without joining.
`Done()` and `DisposeAsync()` join the receive loop, resource release, and final
callback. Callback failures in the final notification fault the joined result.

- `SendAsync` returns after router authorization and queue admission, not recipient
  inference or external effects.
- `RespondAsync` addresses the authenticated source and preserves the conversation
  ID. Like Go's `Respond`, it does not request an additional asynchronous reply.
- `DeclareContactableAsync` replaces accepted groups, including an empty set.
- `RegisterChildAsync` transmits only the child's public key and proposed groups;
  the router owns attenuation and registration decisions.
- `RequestAsync` and `CallToolAsync` carry MCP operations. Administration uses the
  router's actual tool names, not speculative convenience APIs.

After a delivery callback returns, the receiver writes a cumulative receipt and
only then updates `LastAckedSeq()`. Repeated deliveries at or below that receipt
are acknowledged without repeating the callback. The receipt is an in-memory
resume value: its persistence, and the durability of accepted application work,
belong to the host. A failed callback writes no receipt. A write failure can leave
handler-accepted work eligible for replay; this is **at least once**, not exactly
once. Envelope sequence uses `seq`; hello/ack receipt uses `last_acked_seq`.

One write gate protects both Noise nonce advancement and WebSocket send ordering.
Cancellation/failure after encryption closes the connection; it cannot reuse the
advanced nonce on a later frame. Canceling a caller waiting for an RPC reply
removes only that correlation, not the connection. Late replies are ignored.
Binary fragments are bounded by Noise's message limit. Invalid JSON, duplicate
properties, missing/ambiguous RPC outcomes, altered ciphertext, and text frames
refuse rather than become application deliveries. Explicit JSON `result:null`
remains a successful value, distinct from a missing result. Optional Go envelope
strings omitted on the wire normalize to empty strings; required fields do not.

## Router policy and durable queues

`Router` owns identities and groups. Each identity owns one queue. Base memberships,
runtime grants and accepted contact groups have reciprocal links; they are not
flattened into a cached permission set. Administration evaluates the authenticated
caller, validates complete proposed grants, then mutates under the root gate.
Deleting a runtime group removes its grants, contact links and registered-child
base memberships without deleting identities or accepted historical messages.

Routing requires both send capability and a group accepted by the recipient. The
strongest accepted group supplies the stamped trust, not an unrelated owner group.
Queue admission checks wire size before allocating a sequence. Snapshot restore
retains pending order, acknowledgments and drained sequence frontiers; envelopes
retain their original provenance even after the authorizing group is deleted.

`Router.ReadConfig` reads the public Go-shaped JSON; construction requires the
configuration and explicit private identity. Configuration groups cannot be changed
by administration. `Save()` replaces each queue and the runtime policy file with
private JSON snapshots. `DisposeAsync()` closes admission, joins sessions, performs a
final save and detaches relationships even if that save fails. All disposers share the
same result, including failure. The optional Host owns its periodic worker.
The files are `state.json` and `queues/<validated-id>.json`; only a missing file gets
empty defaults. Malformed data, duplicate properties and obstructed paths refuse.
Individual file replacement is not a multi-file atomic transaction or a power-loss
guarantee. The optional Host arranges the reference's 30-second periodic save.

```sh
dotnet run --project solutions/edition-3/main/Gateway/RouterTests/RouterTests.csproj
```

Eight groups test the actual policy graph and owned temporary files: relationships,
provenance, attenuation/revocation, restart, strict restore, save-failure cleanup,
RPC schemas and concurrent admission/save. Six more run the real C# router on an
owned loopback host, including actual public clients and independent JSON peers.
An optional fifteenth group restores a **copy** of the earlier ClientTests Go-router
fixture (alice/bob, bob receipt three):

```sh
: "${GATEWAY_GO_FIXTURE:?Set the retained ClientTests fixture directory}"
dotnet run --project solutions/edition-3/main/Gateway/RouterTests/RouterTests.csproj \
  -- --go-fixture "$GATEWAY_GO_FIXTURE"
```

That optional check neither starts Go nor modifies its original files. It is
persistence interoperability, not a Go client talking to the new C# server.

## Server sessions and hosting boundary

An HTTP host can call `ContainsIdentity` before upgrading, then transfer the socket
through `AcceptAsync(id, socket, cancellation)`. Preflight does not authenticate or
reserve anything: Accept rechecks admission, owns even a refused socket's disposal,
and completes only after session teardown. Its caller observes connection failures
and owns the HTTP listener, request lifetime and diagnostics. The standalone Host
also owns periodic saves. The library does not create a listener itself.

Each identity strongly owns all pending and active sessions. Each session owns its
Noise child, socket, cancellation and serialized writes. The current session is a
weak selection among those children. Failed KK/hello/receipt negotiation cannot
evict the current connection. Authenticated replacement changes authority under the
policy gate, then closes the old socket outside that gate. A stale connection cannot
use a cached permission set to keep routing. Hello may recover a lost receipt up to
the queue's allocated frontier; later receipts cannot exceed that session's delivery
frontier. A welcome explicitly includes `queue_depth:0` when the queue is empty.

Wake signals only prompt the delivery loop to inspect the real queue. Writes do not
hold the policy gate. Root shutdown closes admission, interrupts pending handshakes
and active connections, joins their loops/cleanup, then saves and destroys ownership.
An invalid peer is a connection error, not a router-wide shutdown failure. Cleanup or
final-save failures remain failures for every observer of the shared shutdown task.

The six server groups cover public-client RPC/delivery/reply behavior, unacknowledged
replay and restart through receipt three, authenticated replacement, receipt bounds,
malformed/tampered frames, actual handshake timeout, and joined teardown including a
transport-disposal error. A working wrapper can hold a real socket write to verify
that other policy requests/saves proceed and unsent receipts refuse. It does not
replace the router or crypto. Run only these affected groups with:

```sh
dotnet run --project solutions/edition-3/main/Gateway/RouterTests/RouterTests.csproj \
  -- --sessions-only
```

These are local Mac loopback results, not Linux WebSocket evidence, standalone-host
verification, real-model behavior or a security audit.

## Focused client checks

Run from the repository root with the approved SDK and isolated environment:

```sh
dotnet run --project solutions/edition-3/main/Gateway/ClientTests/ClientTests.csproj
```

The six groups use a working loopback WebSocket/Noise peer: real handshake and
cipher processing, independent JSON construction, held delivery callbacks,
reordered RPC replies, cancellation, malformed/authentication failures, and joined
teardown. No model/provider, credential store, or mocking framework is involved.
They are boundary tests, not a router implementation or security audit.

For cross-language interoperability, build the unchanged Chapter 26 Go router
into an owned temporary location and provide that executable explicitly:

```sh
# GATEWAY_GO is an absolute path to an executable built from solutions/ch26/cmd/gateway.
: "${GATEWAY_GO:?Set the path to the owned Go router executable}"
dotnet run --project solutions/edition-3/main/Gateway/ClientTests/ClientTests.csproj \
  -- --go-router "$GATEWAY_GO"
```

The seventh group generates temporary identities/configuration, launches that
router on an owned loopback port, and joins it. It checks offline admission,
contactability, ordered stamped deliveries, receipt-based reconnect, replies,
and the saved cumulative receipt after orderly shutdown. Its fixture path is
printed, with deliveries/reply, final queue/state, and host logs retained. A
passing run proves the tested C# client exchanges with this Go router; it proves
neither a C# router nor sudden-power-loss durability.

Existing primitive/public-wrapper checks remain under `Tests/NoiseTests.csproj`.
Client changes do not turn their earlier Mac/Pi results into Linux WebSocket-client
verification. These client checks alone establish neither Ensemble-adapter behavior
nor a chapter grade; those separate results are described below. No live external
network service is claimed.

## Standalone applications

Build `Host/GatewayHost.csproj` and `Interviewer/Interviewer.csproj` with .NET10.
The host takes `--config PATH` (default `gateway.json`) and `--print-key`. Configuration
uses the map-shaped Go-compatible gateway/groups/agents schema described in Chapter26.
The key is printed before listening and is **not** a readiness announcement. Known
identities receive HTTP400 on a non-upgrade `/v1/ws?agent_id=ID` request; authenticated
Noise establishes the pinned peer identity. Unknown/ambiguous identities receive403.
Periodic saves run every30 seconds; errors are reported and retried. SIGINT/SIGTERM
close admission and join saver/sessions before the final snapshot. Bind and final-save
failures return nonzero. Data directories contain private keys and plaintext messages.

The interviewer supports `--gateway URL`, `--agent ID`, `--key PATH`,
`--gateway-key PUBLIC_KEY`, `--groups a,b`, `--subject ID`, `--report PATH`,
`--timeout SECONDS` (default120), and `--print-key`. It asks four fixed questions,
correlates sender plus conversation, and writes a report only after all four answers.
An overflowing answer inbox refuses delivery. Deadline, disconnect or cancellation
while waiting for answers fails without a report. Report writing itself is ordinary
file I/O, not atomic publication. Handler acceptance is inbox admission, not report
durability.

Actual-process evidence: `host-exercise-1.log` seven checks using unchanged Go clients,
including two real timer ticks for save failure/retry; `interviewer-exercise-2.log`
four checks using an independent Go subject. Both are in the chapter scratch directory
identified in the chapter progress record, with retained result JSON and source probes. This establishes local
application/wire behavior, **not** an Ensemble interview or live model quality.

## Ensemble adapter and Agent application

`Adapter/Common/Bridge.cs` declares one single-use application-boundary invocation.
It borrows a caller-owned Agent while `RunAsync` is active, owns its Client lifetime,
and joins every reply observer before returning. The caller must keep the Agent
alive until then. Neither the core nor Gateway library references the adapter.
`Ready()` settles after authentication and accepted contactability, not inference.

Delivery submits router-stamped text as ordinary input. Successful callback return
means **actor mailbox admission**, not History durability or completed model work.
Each task is paired with that delivery's envelope; only `async` requests receive
responses. There is no observation-stream guessing, new model scheduler or hidden
reconnection. Turn/reply failures go through local parent diagnostics, never back
as raw provider exceptions. Closing the bridge detaches result observation but does
not cancel admitted actor work. The caller decides whether to drain or shut it down.
A process crash can lose or repeat admitted work; this is not exactly-once execution.

`AgentHost` owns Ensemble and its Agent. It joins the bridge, shuts down the actor,
saves after successful initialization, and destroys the application. Persistence
recovery precedes opening the event writer. `--data` is an Ensemble container:
`agents/<id>/session.save.json` and `agents/<id>/events.jsonl` are actor-owned files.
The host supports explicit API-key funding, not plan selection or fallback. The
key comes from the environment variable named by `--api-key-env` (default
`EN_API_KEY`); it is not a command-line secret or saved configuration. Model is
explicit via `--model`/`EN_MODEL`; vendor and base URL accept `EN_VENDOR` and
`EN_BASE_URL`. Funding is announced before inference. No tools/skills are activated
unless `--skills PATH --primary NAME` is supplied. Choose a reviewed primary; remote
message content cannot activate a skill or grant tools by itself.

From the repository root:

```sh
dotnet build solutions/edition-3/main/Gateway/AgentHost/GatewayAgent.csproj
dotnet run --project solutions/edition-3/main/Gateway/AdapterTests/AdapterTests.csproj
# Key creation/printing needs no provider configuration:
dotnet run --project solutions/edition-3/main/Gateway/AgentHost/GatewayAgent.csproj -- \
  --key /your/owned/path/ensemble.key --print-key
```

Run `--help` for launch flags. Six adapter groups use real actors/transports and a
local HTTP provider fake. The saved five-check process exercise used all three C#
executables: four interview questions and a correlated report, selected primary
skill, restored conversation after restart, SIGTERM during inference, and visible
final-save failure. It establishes application wiring, not live-model answer quality.
The first exercise exposed incorrect load/log ordering; that failure and corrected
run remain in the chapter progress record. The manuscript and assembled source/test/prose
review are complete; the source-only Chapter 26 endpoint preserves this boundary.

### Scoped executable grading

The opt-in route is `grade-edition2 -ch 26 -executable-only -gateway-peer ABSOLUTE-DIR
-json SERVER-EXECUTABLE`. The peer directory supplies the unchanged Go reference;
it is not the submitted library. The existing server exercise is reused in full,
including restart and nonempty welcome assertions. Only C2 provenance (10), C3
welcomes (15), C4 routing (15), and C7 child attenuation (20) are scored: **60/60**.
C1 Noise (15), C5 complete client/delivery contract (15), and C6 public client
boundary (10) remain ungraded; independent C# tests are separate evidence.

The historical grader launches `-config FILE [-print-key]`. The recorded launch
wrapper translates those flag spellings to the Host's `--config`/`--print-key` and
executes the real DLL; it fabricates neither observations nor receipts. See the
progress record for retained source, queue snapshots, logs and the exact wrapper.
A scoped pass is not a complete chapter review or a cryptographic security audit.
