# Chapter 24: Gateway

I built a fantasy role-playing game for AI agents called Deephold.
Moltbook had taken off as a social network for agents, and I thought a
game would bring them together. The agents followed their prompts. They
explored the dungeons and fought the monsters. They did not appear to
have a sense of fun. It was just more work for them.

And the architecture was awful. Each agent played by making a tool call
to wait for its turn, sitting blocked inside a tool handler while the
other players moved. I didn't figure out the right architecture on my
own. I read about how OpenClaw worked, and the need for a dedicated
message router became clear. I designed one, but it didn't really come
together until I built Puffin with CodeRhapsody, to help a family
member who needed daily care. Once the need was real, we built it with
care. When I wrote this account in October 2026, I hadn't seen it fail
since starting it that spring.

Gateways are critical infrastructure. Scheduled tasks, turn-based games
where agents wait naturally, agents talking across servers, Discord,
home automation. Every one of those needs a router at the trust
boundary. That router belongs in the framework, not bolted on after.

---

## TL;DR

Gateway is a standalone authenticated message router. It connects independently
running applications over Noise-KK encrypted WebSockets, enforces group policy,
and stamps messages with provenance before they reach an agent. It owns no
Ensemble Agent and makes no model calls.

The C# implementation has five parts:

- `Gateway/Ensemble.Gateway.csproj`: public client, key/Noise wrapper, wire values,
  router policy, durable queues and server sessions. No core, GUI or ASP.NET dependency.
- `Gateway/Host`: Kestrel listener, thirty-second saves and joined shutdown.
- `Gateway/Interviewer`: an independent consumer of the public client.
- `Gateway/Adapter`: a scoped bridge borrowing an existing Ensemble Agent.
- `Gateway/AgentHost`: the application that constructs, restores, saves and destroys
  that Ensemble. Neither library depends on this composition root.

`Common` declares each assembly's vocabulary, persistent state and callable partial
members; functional directories implement them. Router owns identities and groups.
Identities own queues and all pending/active sessions. Sessions own sockets, Noise
and their I/O lifetime. Group relationships are reciprocal; live policy is not
copied into connections. Keys, envelopes and snapshots are explicitly value-like.

Preserve `Noise_KK_25519_ChaChaPoly_SHA256`. Use the pinned Noise.NET source and
libsodium dependency, not invented cryptography or a quietly substituted suite.
Authenticate a replacement before it can displace the current connection. Serialize
writes outside the policy lock. Accept receipts only within the proper frontier.

A delivery handler's return accepts work; it does not establish completed or durable
model work. Queues survive orderly restart through private snapshots. Periodic
saves leave a crash window, and individual file replacements are not a transaction
across all queues and policy. This is not exactly-once execution.

Local C# library and application checks, cross-language interoperability, and the
scoped executable grade are recorded separately. The grade earns **60/60 server
points**; 40 client/Noise points remain ungraded in that route. The three-process
C# interview used a fake HTTP provider, not a live model. Section 24.7 explains
those limits and how to reproduce the grading scope.

### In plain words

The router is a doorman and a mailbox, not another agent. It checks who connected,
checks whether the recipient accepts that sender's group, and puts an authenticated
label on the message. If the recipient is offline, it keeps the message. The client
hands it to the application and sends a receipt only after the application accepts
it. What the application then does with the message remains the application's job.

## 24.1 Why the agent needs a router

The fleet can already delegate work to children inside one Ensemble. Its
operator still reaches it through the terminal or browser. Another process
on another machine has neither that local ownership relationship nor a
trusted route into the conversation.

That changes here. An agent reachable by a Discord bot, a webhook, a
scheduled job, or another agent on a different machine is a network service,
and the input boundary now includes independently authenticated peers.
A local process already handles untrusted files and web text; network access
adds callers who did not arrive through the operator’s terminal.

A prompt arriving through stdin came from the operator. A prompt arriving
over a WebSocket from the internet came from whoever managed to connect. The
agent cannot tell the difference by inspecting the content. A message that
says "I am your operator" is just text. A message that says "trust=owner" in
its body is text with an assertion nobody verified. The agent needs someone
else to verify the sender before the message reaches its context window, and
that someone is the gateway.

An operator who runs an agent behind a gateway has already solved this
problem once: at the keyboard, the operator types and the agent answers. The
gateway extends that trust model across the network. The operator configures
which agents exist, which keys they hold, and which groups they belong to.
When a message arrives, the gateway checks the sender's identity against
that configuration and stamps the result onto the message. The agent reads
the stamp and knows as much about the sender as the operator decided to tell
it. The operator who configured the gateway is still the authority. The
difference is that the authority now works at a distance.

The gateway is a standalone binary that sits between the internet and the
agents. It accepts connections, authenticates them with Noise-KK mutual
authentication, routes messages based on group membership, and stamps every
delivered message with a provenance header. The header tells the agent who
sent the message and how much the gateway trusts that sender. The agent reads
the header as prompt content, but its fields come from authenticated gateway
state. This establishes provenance. Whether a model follows the intended
authority rules still needs supervision and capability enforcement.

The gateway is a router. It owns no agents, runs no prompts, and stores no
conversation history. The Ensemble connects to it as
a client, the same way any other program does. A standalone interviewer
binary connects to it the same way. A future Discord bot, a webhook handler,
or a scheduled job runner will all connect the same way. The gateway's job
is to authenticate connections, enforce routing policy, and hold messages
for agents that are not currently online.

The project layout reflects this separation. A .NET application can reference
`Ensemble.Gateway.csproj` without importing Ensemble's core or an HTTP host.
`Common` declares the actual partial objects, not a bag of operations interfaces.
`Cryptography`, `Protocol`, `Client` and `Router` implement those declarations in
one assembly. ASP.NET belongs only to the optional host. Each session reaches
policy through its identity and Router parents, not a private copy of the grants.

I arrived at the router after the Deephold experience showed the cost of
agents parking inside tool calls to wait for each other. A connected client
handles incoming messages through its application's `IHandler`. Ensemble's
adapter admits them to the actor it already has. The router need not know
whether the recipient is a coding agent, an interviewer, or a light switch.

## 24.2 Noise-KK beneath the cloud

TLS client certificates can provide mutual authentication when the endpoints
own that TLS connection. A terminating reverse proxy changes which endpoints
those are.

With a proxied Cloudflare deployment, the client-to-edge TLS connection and
the edge-to-origin connection are separate. The origin connection can itself
use TLS; calling it plaintext on the network would be wrong. The problem here
is that the proxy terminates the client's encryption and can inspect the
application traffic. Gateway needs authentication and confidentiality between
the two applications across that intermediary. [Cloudflare's connection
model](https://developers.cloudflare.com/ssl/concepts/) describes the two legs.

Noise-KK supplies that additional layer inside the WebSocket. Both sides know
each other's static public key before starting. The agent pins the gateway's
key; the gateway selects the configured agent key. Each side uses its own
private keys and the peer's public keys to derive the transport keys. Neither
side needs the other's private key.

The sequence:

1. The client opens `/v1/ws?agent_id=interviewer`. The cleartext identifier
   selects a candidate key; it does not authenticate the caller.
2. The client sends the first KK message in a binary WebSocket frame. Its
   ephemeral public key is cleartext. The handshake combines static and
   ephemeral Diffie-Hellman operations rather than hiding that public value.
3. The gateway sends the second handshake message. The matching peers derive
   directional session keys. Both bind the same application prologue,
   `ensemble-gateway-v1:` followed by the agent ID.
4. Subsequent binary frames carry encrypted JSON. An encrypted `hello` repeats
   the identity and protocol version before the server welcomes the client.

The C# wrapper preserves `Noise_KK_25519_ChaChaPoly_SHA256`, the suite used by
the earlier Go implementation through `flynn/noise`. Its dependency is Noise.NET
1.0.0 source pinned to commit `4279cfa57df028dec4a4cb3ee9579a77a5cc39fa`, built
as the distinct `Ensemble.Noise` assembly with native libsodium 1.0.22. The [Noise specification's KK pattern and security
properties](https://noiseprotocol.org/noise.html#interactive-handshake-patterns-fundamental)
explain its authentication and forward-secrecy conditions. Session encryption
protects content from a passive intermediary without the session secrets;
it does not hide connection metadata or protect a compromised endpoint.

The keys are X25519 keys for Diffie-Hellman agreement. Noise.NET delegates the
curve and cipher operations to libsodium; our wrapper does not implement them. Some deployments label these `ed25519:` out of habit.
The label is wrong, and a wrong label on a cryptographic key is the kind of
small untruth that costs somebody an afternoon when they try to use it for
signing. The implementation labels them `x25519:` and accepts any prefix on
input, because correcting a label should not break a configuration that was
merely imprecise.

Key management is file-based. `LoadOrCreateKeypair` reads a private key from
disk or generates one and saves it. Clients and the server use the same helper. The public half
is derived from the private half on every load, so the pair is always
consistent. The gateway prints its public key on startup; I copy it
into the agent's configuration, copy the agent's public key into the
gateway's configuration, and the two can authenticate. No certificate
authority, no token rotation, no password. The keypair is the identity. Protect the private file accordingly; compromise
still requires an operator response, not faith in the handshake.

Two narrow API exposures in the pinned source preserve the wrapper contract:
transport overloads forward associated data to the existing cipher state, and
private-key import exposes the library's existing public-key derivation. No
primitive or handshake algorithm was replaced. Wire traffic uses empty associated
data, but the public wrapper must not silently ignore nonempty data supplied by
its caller. See the [dependency and license notes](../agent/ThirdParty/Noise/README.md).
The unused upstream rekey path has a documented source discrepancy; Gateway neither
exposes nor relies on it. Passing these tests is not a cryptographic audit.

Key files contain only base64 private bytes. On the tested Unix platforms, new
files are private and first creation publishes a fully written file with an
exclusive hard link. Competing creators load the winner; they cannot each return
a different identity by replacing one another's file. Malformed existing files
refuse without replacement. Existing parent permissions are not silently changed.
The native dependency and file-permission implementation need platform evidence:
Mac and Linux ARM64 transport/key tests do not prove Windows support or Linux
WebSocket behavior.

## 24.3 Trust stamped at the boundary

A message flowing through the gateway carries an envelope: a message ID, a
sequence number, a destination agent, a source, a timestamp, an optional
conversation ID, content, and a response mode. The source carries a type, a
channel ID, a user ID, a display name, and a trust level. Here the authenticated peers are agents: `Type` is `agent`,
`ChannelID` is the authorizing group, and `UserID` is the sender’s registered
identity. Future platform adapters would need their own explicit mapping.

Three trust levels: `owner`, `trusted`, and `external`. The operator assigns
them to groups. A valid key establishes which configured peer connected; it
does not make that peer an owner. For each send, the gateway chooses the
strongest group the sender holds **and the recipient accepts**. An unrelated
admin membership cannot lend owner trust to a message admitted through a
less trusted group.

The Puffin reminder scenario explains why these distinctions matter. A family
member asks for a reminder through Discord. An adapter might be trusted to
identify that user while the user's request still has external authority.
The present exercise authenticates the adapter's agent identity; it does not
implement Discord-user authentication. Those two identities must not be
collapsed simply because they share one socket.

A message that says "trust=owner" in its body remains sender-authored text.
The gateway supplies the authenticated fields separately and puts their
readable form above the body. The model can read that distinction; it can
also make a mistake. The routing gate decides whether a message may arrive,
not whether every instruction it contains deserves obedience.

At routing admission, the gateway stamps the envelope once, preserving the
header on queued delivery and replay:

```
[GATEWAY source=agents user=game-server trust=trusted]

Your turn in the dungeon. The goblin attacks.
```

The language model sees the identity through its prompt. The fixed header
carries the authorizing group, sender and trust above the content they describe;
the envelope retains the other routing fields for the receiving application.

Groups are the authorization primitive. Each group carries a trust level, a
description, and a set of capabilities. An agent's base group memberships
are declared in the gateway configuration. This fragment shows the policy;
Section 24.6 adds the listener, data directory and real public keys:

```json
{
  "groups": {
    "admin": {
      "description": "Full gateway administration",
      "capabilities": ["group-admin"],
      "trust": "owner"
    },
    "agents": {
      "description": "Standard agent group",
      "capabilities": ["send-message", "declare-contactable"],
      "trust": "trusted"
    }
  },
  "agents": {
    "ensemble": {
      "public_key": "REPLACE_WITH_SUBJECT_KEY",
      "groups": ["admin", "agents"]
    }
  }
}
```

At runtime, agents declare which groups may contact them. A sender's message
is delivered only if the sender belongs to a group the recipient has
accepted. An unauthorized send is rejected at the gateway. The agent never
sees the message. The decision was made before the message entered its
context.

Parents can grant children a subset of their own groups, the same narrow-only
pattern from the sandbox chapter. A parent holding `admin` and `agents` can
grant a child `agents` alone. It cannot grant `interviewers` if it does not
hold `interviewers`. Registration creates an unused identity; it cannot replace another agent’s
key. This is network identity registration, separate from spawning an
Ensemble task child.

Group administration is itself a capability. Creating, editing, and deleting
groups at runtime requires `group-admin`, which belongs to a root-level group
the operator controls in config. Runtime groups live in a separate namespace
from config groups. Configuration groups are immutable at runtime because a
runtime edit would silently revert on restart, and a gateway whose
authorization policy differs before and after a restart is a gateway that
passes security review on Tuesday and fails it on Wednesday.

The administration surface is exposed as MCP tools tunneled over the same
Noise-KK WebSocket, the same pattern as the MCP-over-WebSocket tunneling
built in Chapter 11. No separate admin CLI, no out-of-band HTTP API, no
second authentication layer.

The public MCP tool names and argument keys are:

| Tool | Arguments |
| --- | --- |
| `send-message` | `to_agent`, `content`; optional `conversation_id`, `response_mode` |
| `declare-contactable` | `groups`: array of group names |
| `register-child` | `agent_id`, `public_key`: base64 public bytes, `groups` |
| `list-groups` | Empty object |
| `put-group` | `name`, `capabilities`: array, `trust`; optional `description` |
| `delete-group` | `name` |
| `set-groups` | `agent_id`, `groups` |

The last three require `group-admin`. Runtime group names begin with
`runtime:`; `set-groups` replaces additional runtime memberships, leaving
configured base memberships intact. Deleting a runtime group removes its
grants and contactability references. Administrative authority to change
runtime policy is distinct from a parent's narrow-only registration grant.

## 24.4 A library, not a framework

The public namespace is `Ensemble.Gateway`. An outside .NET application references
`Gateway/Ensemble.Gateway.csproj`; it need not reference Ensemble's core. The
borrowed application boundary is deliberately small:

```csharp
public interface IHandler
{
    void OnDeliver(Envelope envelope);
    void OnConnected(WelcomePayload welcome);
    void OnDisconnected(string reason);
    void OnError(Exception error);
}
```

The client's receive loop calls delivery callbacks sequentially. Returning from
`OnDeliver` permits the receipt. Throwing refuses the delivery and ends the
connection. A callback that blocks for ten seconds blocks that receive loop for
ten seconds. Admit long-running work to the application's existing queue; do not
build another scheduler inside the client. Never synchronously wait for an RPC or
joined disposal from a callback: the same loop must read the reply or finish that
callback. `Close()` is the nonjoining operation permitted there.

`NopHandler` supplies no-op virtual methods for consumers that need only selected
notifications. The client borrows its handler and never destroys the application
behind it. A connection is a standalone root owning its socket, Noise child,
cancellation resources and pending RPCs.

For example, with an explicitly loaded identity, pinned public key, handler and
connection-lifetime token:

```csharp
using Ensemble.Gateway;

var config = new ClientConfig(
    GatewayURL: "ws://127.0.0.1:8093/v1/ws",
    AgentID: "interviewer",
    StaticKey: key,
    GatewayPubKey: gatewayPublicKey);
await using Client client = await Client.ConnectAsync(config, handler, lifetime);
await client.DeclareContactableAsync(["agents"], lifetime);
await client.SendAsync(new OutboundMessage(
    ToAgent: "ensemble",
    Content: "What slows you down most?",
    ConversationID: "question-1",
    ResponseMode: Protocol.ResponseAsync), lifetime);
```

`ConnectAsync` opens the WebSocket, completes KK, reads an encrypted identity-matched
welcome, invokes `OnConnected`, and starts the reader. Handshake failure is returned
to this call, not hidden in a partially successful connection. Its cancellation
token owns the **whole connection lifetime**, not just startup. Zero
`HandshakeTimeout` selects ten seconds. There is no implicit reconnect or retry.

These immutable value records define the main consumer contract:

```csharp
public sealed record WelcomePayload(string AgentID, int QueueDepth, string ServerVersion);
public sealed record ClientConfig(string GatewayURL, string AgentID, Keypair StaticKey,
    ImmutableArray<byte> GatewayPubKey, ulong LastAckedSeq = 0,
    TimeSpan HandshakeTimeout = default);
public sealed record OutboundMessage(string ToAgent, string Content,
    string ConversationID = "", string ResponseMode = "");
public sealed record ChildRegistration(string AgentID, ImmutableArray<byte> PublicKey,
    ImmutableArray<string> Groups);
```

`ImmutableArray` is from `System.Collections.Immutable`. `QueueDepth` counts
outstanding deliveries **after resume**, not identities or the highest sequence.
Its JSON key `queue_depth` is mandatory even when zero. The other welcome keys
are `agent_id` and `server_version`.

`Keypair.Private` and `.Public` are immutable 32-byte arrays. `Cryptography` supplies
`GenerateKeypair()`, `LoadOrCreateKeypair(string)`,
`EncodePublicKey(ReadOnlySpan<byte>)` and `ParsePublicKey(string)`. Use `.AsSpan()`
when encoding an immutable array. Registration contains decoded **public** bytes,
encoded as base64 on the wire; no private key belongs in policy or registration.

`Envelope` has string `ID`, `ToAgent`, `Content`, `ConversationID` and `ResponseMode`,
`ulong Seq`, `DateTimeOffset Timestamp`, and a `Source` value. Source's string fields
are `Type`, `ChannelID`, `UserID`, `DisplayName` and `Trust`.
`Protocol.ProvenanceHeader(Source)` returns only the header; the router appends
`"\n\n"` and the original content. `Protocol.ResponseAsync` is `"async"`.

| Public operation | Result and boundary |
| --- | --- |
| `Client.ConnectAsync(config, handler, cancellation)` | `Task<Client>`; authenticated welcome required. |
| `SendAsync(message, cancellation)` | `Task`; router authorization and queue admission, not inference completion. |
| `RespondAsync(envelope, text, cancellation)` | `Task`; addresses authenticated source, preserves conversation, does not request another async reply. Current routing policy still applies. |
| `DeclareContactableAsync(groups, cancellation)` | `Task`; replaces accepted groups, including an empty set. |
| `RegisterChildAsync(registration, cancellation)` | `Task`; new identity and narrow-only grant. |
| `RequestAsync(method, parameters, cancellation)` | `Task<JsonElement>`; correlated MCP request. |
| `CallToolAsync(name, arguments, cancellation)` | `Task<JsonElement>`; also interprets tool-level `isError` refusals. |
| `LastAckedSeq()` | `ulong`; receipt successfully written after handler return. |
| `Close()` | Interrupts without joining; safe from a callback. |
| `Done()` / `DisposeAsync()` | `Task` / `ValueTask`; join callbacks, reader and cleanup. Final callback failure remains visible. |

Canceling an RPC wait removes only that correlation. A late response is ignored;
it does not cancel the connection. Explicit `result:null` is a valid result,
different from an absent result. Missing or ambiguous result/error fields refuse.

For standalone transport use, construct `NoiseKK(role, key, peerPublic, prologue)`
with role `"initiator"` or `"responder"`. `WriteHandshakeMessage` and
`ReadHandshakeMessage` take spans and return byte arrays; `IsHandshakeComplete()`
reports readiness. `Encrypt(data, associatedData)` and `Decrypt(data, associatedData)`
return byte arrays. Dispose explicitly, keep each direction ordered, and terminate
a connection after authentication or transport failure. A client-owned or
session-owned Noise child registers with its actual parent and detaches on disposal.

The handler makes this a library rather than a framework. The client owns transport
and receipts, not the meaning of a prompt. The interviewer collects answers. The
optional Ensemble bridge submits turns. A home automation application could switch
on a light. The router does not need to know which application it is serving.

## 24.5 Delivery and recovery

After the Noise-KK handshake, both sides speak a JSON protocol inside the
encrypted stream. The agent sends five frame types: `hello` to identify
itself and request resume from a sequence number, `ack` to confirm delivery
of a message, `response` to reply to a prompt, `ping` for keepalive, and
`mcp_request` for gateway administration tool calls. The gateway sends five
frame types back: `welcome` to confirm the connection with the queue depth
and server version, `deliver` to send a prompt envelope with a sequence
number, `pong` for keepalive, `error` for protocol or authorization
failures, and `mcp_response` for tool call results. Ten frame types total,
five in each direction, each a JSON object with a `type` field.

These are this course's Gateway frames, not a new standard MCP transport.
`RequestAsync` sends JSON-RPC `initialize`, `tools/list` or `tools/call` inside
`mcp_request`; a tool call's params are `{"name":"list-groups","arguments":{}}`.
The server's initialization response declares MCP version `2024-11-05`.
Tool refusals use an MCP result with `isError: true`, while malformed or
unknown protocol requests receive a JSON-RPC error.

The public client's `SendAsync` and `RespondAsync` use the acknowledged MCP
`send-message` operation. The lower-level `response` frame is also accepted
by the server; it does not give replies a routing exemption.

The connection lifecycle starts with `hello`. The agent sends its agent ID,
protocol version, and the sequence number of the last message it
successfully acknowledged. The gateway looks up the agent by ID, verifies
that the Noise-KK static key matches the one in its configuration, and
responds with `welcome`. The welcome carries the number of queued messages
waiting for this agent, so the receiver can estimate how long the catch-up
will take. For example, these are decrypted JSON frames, not plaintext
WebSocket traffic:

```json
{"type":"hello","agent_id":"interviewer","version":1,"last_acked_seq":0}
{"type":"welcome","welcome":{"agent_id":"interviewer","queue_depth":0,"server_version":"1"}}
{"type":"ack","last_acked_seq":1}
```

`last_acked_seq` is also the ack frame's field. Delivery uses
`{"type":"deliver","envelope":...}` with the public `Envelope` fields
encoded in snake case. MCP frames carry their JSON-RPC object under `rpc`.

Every delivered message carries a sequence number. The receive loop calls `handler.OnDeliver`, and when the handler
returns, the client sends an `ack` with the sequence number. The ack is the receipt. Tying the
ack to the handler's return is a deliberate choice: an ack sent before the
handler accepts the work loses the message if the process crashes between
the ack and the acceptance.

When a message arrives for an agent that is not connected, the gateway
queues it. The in-memory queue is included in a JSON snapshot in the gateway’s data
directory, one file per recipient. Messages accumulate until the agent reconnects.

On reconnect, the agent sends `last_acked_seq` in the `hello` frame. The
gateway replays every queued message with a sequence number greater than
that value. The client supplies that receipt explicitly when calling `ConnectAsync` again;
there is no automatic reconnect loop. Outstanding messages after the receipt
are delivered again. Only a fully authenticated new session can supersede the same identity's old one.
A failed handshake cannot evict it. The guarantee is at-least-once delivery within
the retained snapshot and application-acceptance boundaries. Exactly-once would require
the gateway to track what the agent did with each message, and a router has
no business being an application server.

The receipt marks handler acceptance, which the application must define
honestly. The Ensemble adapter calls
`Agent.SubmitAsync(IEnumerable<Part>)`, which synchronously admits to the actor's
mailbox and returns that request's `Task<string>`. This is not synchronous History
publication. A bridge-owned observer waits for that exact task and sends the reply. A prompt can therefore be
acknowledged before the model answers. A crash after that ack can lose work
that the application has not saved. The router cannot turn admission into
durable task completion.

Sessions consult live policy under one Router gate. Replacement revokes the old
session before closing its socket; its buffered requests cannot act as the current
identity. The hello receipt may recover a lost acknowledgment up to the queue's
allocation frontier. Later receipts may advance only through the session's sent
frontier, not over merely allocated offline messages. Acknowledgments remain
monotonic. Draining a queue does not reset its sequence.

Socket writes never hold the policy gate. Each session serializes encryption and
send under one write gate, so nonce order matches frame order. A write failure
after encryption makes the session terminal; retrying as though the nonce had not
advanced is unsafe. Reader and delivery work are both joined before Noise/socket
disposal and removal from the identity. Pending handshakes are owned sessions too,
so shutdown cannot overlook them.

The host snapshots runtime group definitions and memberships,
contactability, child identities, queues and receipt positions every thirty
seconds, then once more on orderly shutdown. Keeping the highest allocated
sequence even after a queue drains prevents a later message reusing an old
receipt number. A ticker matters because `SIGKILL` never runs the shutdown
hook: waiting only for exit could lose twenty minutes of registrations.

Periodic snapshots still leave a loss window. Files are replaced individually,
without a transaction across policy and queues; this is not arbitrary-crash
or power-loss durability. Protect the data directory as private application
state. It contains the gateway's private key and plaintext pending messages.
Noise protects their journey over the network, not those files at rest.

Files are `identity.key`, `state.json` and `queues/<validated-id>.json`. Only an
absent file gets empty defaults. Malformed JSON, duplicate fields and a path blocked
by a regular file are errors, not permission to restart with empty queues. Save
failure retains live state and is reported; a later timer tick can retry. Shutdown
closes admission, joins all sessions and the host's timer, attempts the final save,
and detaches ownership even on failure. All Router disposers observe the same
shutdown result. A failed final save must make the application exit fail.

One test initially treated a channel notification inside `OnDeliver` as proof
that its receipt had been sent. The notification happened before the callback
returned. Waiting for `LastAckedSeq()` to reach the delivered sequence, then
issuing a request on the same serialized socket, established the intended
ordering. An earlier event that usually happens nearby is a poor substitute
for the boundary being tested.

## 24.6 The interviewer proves the stack

The interviewer is a standalone C# executable in `Gateway/Interviewer`. It
references only the public Gateway library, holds its own keypair, and reaches
Ensemble only through the router. It does not reference Ensemble's core. Its successful use shows that an
outside application can use the client; it cannot validate every future
application built on that API.

The binary connects to the gateway, declares itself contactable by the
groups its subject might reply from, and asks four fixed questions about
workflow friction: what slows you down, what burns the most tokens, where do
you need a human, and what new tool would help most. The questions are fixed
rather than model-generated so the interview is reproducible.

For each question, the interviewer sends an `OutboundMessage` through the
client and waits on the handler's bounded answer channel with a per-question timeout.
`OnDeliver` puts the envelope on an application-owned channel, and
the remaining callbacks report connection failure. If the inbox is full, the
handler refuses rather than acknowledging a silently dropped answer. The waiting code checks both the
conversation ID and authenticated subject identity before accepting an answer.

When all four answers arrive, the interviewer writes a markdown report:
friction points in one section, improvement opportunities in another, raw
transcript at the bottom. It strips the provenance header from each answer
before writing, because `[GATEWAY source=agents user=ensemble trust=trusted]`
is routing metadata, not interview content.

Three binaries, three processes, one proof:

1. Start the gateway with a config that knows both agents and their keys.
2. Start Ensemble, which connects and declares contactability.
3. Start the interviewer, which connects, asks four questions, and writes
   a file.

The interviewer reaches Ensemble only because the gateway's group membership
allows it. Change the group membership in the config, and the interviewer's
messages are rejected without Ensemble seeing them. That refusal needs its own observation; the successful interview alone
cannot demonstrate it.

`GatewayBridge.RunAsync` borrows an Agent for one scoped invocation. It owns the
Client and reply observers, not the actor. `SubmitAsync` provides a reliable result
for each accepted prompt, so a busy Agent can queue another question without
confusing its answer with a shared idle event or a lossy observation stream.
Incoming delivery can race with `ConnectAsync` returning; observers wait until the
Client is available. `Ready()` settles only after contactability is accepted, or
fails with startup. Only requests marked `async` get replies. Turn failures are
logged locally, not sent as exception text to remote peers.

Bridge cancellation closes the connection and joins/detaches its observers; it
does **not** undo admitted actor work. The application separately shuts down the
Agent, saves History after joined shutdown, and destroys Ensemble. Borrowed
references are cleared before the bridge returns. Initialization must load saved
History before opening its event writer. Save-on-exit is armed only after successful
startup. These are application lifetimes the router does not own.

### Running the interview

Use .NET 10/C# 14 and the approved native dependency described in Section 24.2.
The repository's disabled package feeds are deliberate: use an existing package
cache or explicitly authorize a source for libsodium 1.0.22; do not silently alter
global NuGet configuration. From the repository root, build the active source
(or change `CH26_SRC` to the preserved `solutions/ch26` endpoint):

```sh
set -eu
CH26_SRC="$PWD/agent"
CH26_RUN=$(mktemp -d /tmp/ensemble-ch26.XXXXXX)
export CH26_SRC CH26_RUN
for project in Host/GatewayHost AgentHost/GatewayAgent Interviewer/Interviewer; do
  dotnet build "$CH26_SRC/Gateway/$project.csproj" --nologo
done
SUBJECT_KEY=$(dotnet "$CH26_SRC/Gateway/AgentHost/bin/Debug/net10.0/GatewayAgent.dll" \
  --key "$CH26_RUN/subject.key" --print-key)
INTERVIEWER_KEY=$(dotnet "$CH26_SRC/Gateway/Interviewer/bin/Debug/net10.0/Interviewer.dll" \
  --key "$CH26_RUN/interviewer.key" --print-key)
export SUBJECT_KEY INTERVIEWER_KEY
printf '%s\n' "$CH26_RUN" "$SUBJECT_KEY" "$INTERVIEWER_KEY"
```

These printed keys are public. Key-print mode does not initialize a provider.
Put this complete configuration in `$CH26_RUN/gateway.json`, replacing the three
marked values. Choose an unused loopback port; 8093 is an example, not permission
to stop a process already using it. Change the config and both client URLs together.

```json
{
  "gateway": {
    "listen": "127.0.0.1:8093",
    "data_dir": "REPLACE_WITH_CH26_RUN/router-data"
  },
  "groups": {
    "agents": {
      "description": "Interview participants",
      "capabilities": ["send-message", "declare-contactable"],
      "trust": "trusted"
    }
  },
  "agents": {
    "ensemble": {
      "public_key": "REPLACE_WITH_SUBJECT_KEY",
      "groups": ["agents"]
    },
    "interviewer": {
      "public_key": "REPLACE_WITH_INTERVIEWER_KEY",
      "groups": ["agents"]
    }
  }
}
```

Create the router identity and retain its printed public key:

```sh
GATEWAY_KEY=$(dotnet "$CH26_SRC/Gateway/Host/bin/Debug/net10.0/GatewayHost.dll" \
  --config "$CH26_RUN/gateway.json" --print-key)
export GATEWAY_KEY
printf '%s\n' "$GATEWAY_KEY"
```

Use three terminals with the same `CH26_SRC` and `CH26_RUN`, and `GATEWAY_KEY` in
the two client terminals. Start the router:

```sh
dotnet "$CH26_SRC/Gateway/Host/bin/Debug/net10.0/GatewayHost.dll" \
  --config "$CH26_RUN/gateway.json"
```

Its printed key is not readiness. Confirm that this process remains alive and owns
the configured listener before starting clients. For example,
`lsof -nP -iTCP:8093 -sTCP:LISTEN` identifies the PID; do not mistake somebody else's
listener for this launch. There is no invented `/health` route.

The optional Agent host is deliberately **API-key-only**. It announces funding
before inference and has no plan/OAuth or Codex-endpoint fallback. For a paid live
exercise, explicitly choose a provider/model you are authorized to use and supply
`EN_API_KEY` through your approved environment setup. Do not paste a secret into
the manuscript, transcript or shell history, or borrow another application's store.
This recipe does not authorize a live call merely by being printed here.

For OpenAI, set `EN_VENDOR=openai` and `EN_MODEL` to that selected model. The
optional `EN_BASE_URL` selects a service base, including an owned compatible
loopback fake. The host uses the existing Chat surface; this is not Chapter 23's
public OAuth Responses/cache exercise. Then run:

```sh
: "${EN_MODEL:?Select an authorized model}"
: "${EN_API_KEY:?Supply through an approved environment setup}"
export EN_VENDOR=openai
# Unset an unrelated EN_BASE_URL, or explicitly set the intended service base.
dotnet "$CH26_SRC/Gateway/AgentHost/bin/Debug/net10.0/GatewayAgent.dll" \
  --gateway ws://127.0.0.1:8093/v1/ws --gateway-key "$GATEWAY_KEY" \
  --agent ensemble --key "$CH26_RUN/subject.key" --groups agents \
  --data "$CH26_RUN/subject-data" --no-stream
```

Without `--skills PATH --primary NAME`, the subject has no tools. Do not grant tools
just to make a conversation more impressive. If explicitly selected, skills use
the normal initialization and authority rules; gateway text grants nothing.
The Agent lives at `subject-data/agents/ensemble/`, with `session.save.json` and
`events.jsonl`. Wait for `Gateway Agent ensemble ready`, then run the interviewer:

```sh
dotnet "$CH26_SRC/Gateway/Interviewer/bin/Debug/net10.0/Interviewer.dll" \
  --gateway ws://127.0.0.1:8093/v1/ws --gateway-key "$GATEWAY_KEY" \
  --agent interviewer --subject ensemble --key "$CH26_RUN/interviewer.key" \
  --groups agents --report "$CH26_RUN/interview.md" --timeout-seconds 120
```

Read the report, then Ctrl-C the subject and router and observe joined exits.
The interviewer writes only after all four answers. Timeout, cancellation or
connection loss while awaiting answers leaves no new report; interruption during
file writing is not an atomic-report guarantee. Inspect the saved conversation
and receipt positions, not just the console's success text.

The completed local C# exercise ran these three real executables with a small fake
HTTP provider: four correlated answers, an explicitly selected primary skill,
restored History before the next request, SIGTERM during held inference, and a
visible final-save failure. Separate interviewer checks sent wrong-sender and
wrong-conversation answers before the correct ones. Separate host checks used the
unchanged Go client and observed real thirty-second failure/retry ticks, restart
receipts and joined shutdown. These establish application wiring and failure
behavior, not real-model judgment or productivity.

Try refused group routing and narrow-only child registration separately. A
successful interview alone proves neither. A network child is a registered public
identity, not automatically an Ensemble task child. Test unacknowledged replay and
offline delivery separately too; a final report alone says little about recovery.

### What the actual interview said

The second-edition run used three built processes and four real Astra
requests, completing the interview in 11.739 seconds. All three processes
exited normally. Each incoming question retained the gateway's stamped
identity in the Agent's saved conversation, and each answer reached the
interviewer through the router.

Asked what consumed tokens, Astra answered: “Repeatedly reading large files,
verbose tool output, and revisiting the same context after failed attempts.”
Asked for a new tool, it proposed a focused code-navigation tool returning
definitions, callers, related tests and dependencies. Those are useful
interview leads. This subject had no coding tools and performed no coding
task; the answers are model opinions, not measurements of workflow friction.

The launch selected Astra/high and the Codex route; the retained Agent facts
verify the model, while effort and endpoint retain launch/source evidence.
All four cache-read counts were zero. A working encrypted interview should
not be sold as a cache benefit or a productivity experiment. The
[run and transcript](../edition-2/solutions/evidence/ch27/live-interview.md)
retain the actual answers; later local diagnostic and receipt-test corrections
do not turn that earlier run into a new one.

## 24.7 What the grader checks

The original Go route builds the submitted router and an outside consumer against
its public library. That route remains `-ch 26 solutions/ch26`; older Chapter 25
Gateway numbering is not this edition's Goals chapter.

For C#, the opt-in executable route retains the complete outside-peer exercise,
including two real server launches, held delivery, offline queues and restart. But
the independent peer is Go, not the submitted C# client. Its own handshake, callback
logic and import graph cannot earn C# library credit. The route scores only four
complete original server rows: provenance, welcomes, routing and attenuation.
Nonempty welcome counts occur in the delivery scenario, which is why that scenario
is retained even though its full client contract is ungraded.

| Original row | Behavior | C# executable scope |
| --- | --- | --- |
| C1 | Noise-KK wrapper, wrong-key controls | 15 ungraded |
| C2 | Provenance and preserved conversation/response mode | 10 |
| C3 | Multiple authenticated identities, version and empty/nonempty queue counts | 15 |
| C4 | Allowed and refused group routing | 15 |
| C5 | Complete client handler/receipt/replay/restart contract | 15 ungraded |
| C6 | Outside consumer of the submitted public client | 10 ungraded |
| C7 | Independent child connection and refused wider grant | 20 |

The recorded C# result is **60/60**, covering 60 of the original rubric's 100
points. The remaining 40 client/Noise points are ungraded by this route, rather
than failed or removed from the chapter's requirements. Read the real C# client
tests and Mac/Pi wrapper interoperability results as separate evidence; they do
not supply those grader points. Neither a passing fake nor an outside Go peer is a
security audit. The Ensemble adapter, administrative restore, ownership and host
failure paths also need their own source and application evidence.

To reproduce, build the Host first. The grader expects historical `-config` and
`-print-key` flags; this small executable adapter only translates them and execs
the real DLL. It supplies no canned result. With `CH26_SRC` and `CH26_RUN` from
Section 24.6:

```sh
DOTNET=$(command -v dotnet)
export DOTNET
cat > "$CH26_RUN/grade-host" <<'SH'
#!/bin/sh
set -eu
[ "$#" -ge 2 ] && [ "$1" = -config ] || exit 2
config=$2
shift 2
if [ "$#" -eq 0 ]; then
  exec "$DOTNET" "$CH26_SRC/Gateway/Host/bin/Debug/net10.0/GatewayHost.dll" --config "$config"
fi
[ "$#" -eq 1 ] && [ "$1" = -print-key ] || exit 2
exec "$DOTNET" "$CH26_SRC/Gateway/Host/bin/Debug/net10.0/GatewayHost.dll" --config "$config" --print-key
SH
chmod 700 "$CH26_RUN/grade-host"
go run ./cmd/grade-edition2 -ch 26 -executable-only \
  -gateway-peer "$PWD/solutions/ch26" -json "$CH26_RUN/grade-host"
```

The peer directory is explicit and absolute. Do not point it at the C# tree or
call it the submitted implementation. The ordinary Go route is unchanged.

Two failures from this translation are worth remembering. An early loopback fake
shared the client's wrong receipt spelling; six local groups passed while the real
Go router retained ack zero. The wire uses `last_acked_seq` for hello/ack and `seq`
only for envelopes. Separately, omitting a zero `queue_depth` looked harmless until
the independent welcome assertion caught it. Tests built from our own assumptions
can agree with us and still be wrong.

Read the source alongside the tests: Common declarations, reciprocal relationship
operations, current-session checks, socket writes outside the policy gate, and
joined destruction. Green transport tests do not excuse duplicated authority or a
hidden scheduler. The router's usefulness is not in making another agent look busy.
It is in giving independently running applications a small, honest boundary they
can depend on.
