# Chapter 26: Gateway

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

The gateway is a standalone edge service that sits at the trust boundary,
routing messages between agents over Noise-KK encrypted WebSocket
connections. It uses group-based access control where an agent's
capabilities derive from its group memberships, and parents can grant
children a subset of their own groups (narrow only, never widen). The
gateway administration surface is exposed as MCP tools tunneled over the
same Noise-KK connection.

Build a public client library in `gateway/` with a `Handler`
interface, a `Client`, Noise-KK handshake, and envelope types. The library
must have zero imports from the module’s `internal/` packages. Build a server in
`internal/gateway/` with WebSocket connections, group-based routing,
an offline queue, and MCP administration tools. Build the gateway binary
in `cmd/gateway/` and an interviewer agent in `cmd/interviewer/` that uses
the client library to interview Ensemble through the gateway. An adapter in
`cmd/gateway-agent/` constructs the real public Ensemble root and Agent,
submits stamped messages through its existing actor, and routes completed
answers back. The router itself owns no Agent and makes no model calls.

The required public welcome fields are `AgentID string`, `QueueDepth int`
and `ServerVersion string`; their JSON keys are `agent_id`, `queue_depth`
and `server_version`. Queue depth counts outstanding deliveries after resume.
Section 26.4 supplies the rest of the public callable contract; Section 26.6
supplies the complete launch configuration and three-process exercise.

A handler return accepts delivery and triggers its acknowledgment. It does
not promise completed or durable model work. Preserve offline queues and
explicit reconnect receipts, snapshot dynamic state every thirty seconds
and at orderly shutdown, and retain the at-least-once boundary. Store pending
messages privately: transport encryption does not encrypt queue files.

Build and grade from the repository root:

```sh
go run ./cmd/grade-edition2 -ch 26 solutions/ch26
```

| # | Behavior | Check | Points |
|---|----------|-------|--------|
| 1 | Noise-KK handshake | Two peers complete mutual authentication over a byte stream. Wrong keys are rejected. | 15 |
| 2 | Prompt envelope with provenance | Delivered body and authenticated provenance agree, with conversation and response mode preserved. | 10 |
| 3 | Gateway server accepts connections | Actual binary welcomes multiple authenticated clients with identity, version and empty/nonempty queue counts. | 15 |
| 4 | Group-based routing | Agent declares contactability by group. Gateway delivers messages only from senders whose group membership the recipient has accepted. Unauthorized sends are rejected. | 15 |
| 5 | Delivery semantics | Handler return precedes ack; offline and unacknowledged messages survive reconnect with the proper sequence boundary; a clean server restart preserves queued delivery. | 15 |
| 6 | Client library with Handler interface | Outside consumer compiles and uses the actual public client; its resolved imports exclude private runtime packages. | 10 |
| 7 | Parent registers child | Connected agent registers a child's public key and grants it a subset of its own groups. Child connects independently via Noise-KK. Gateway rejects group grants the parent does not hold. | 20 |

## 26.1 Why the agent needs a router

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

The package layout reflects this separation. `gateway/` is a public client
library with zero imports from private runtime packages. Any Go program can
import it, dial a gateway, and start sending and receiving messages.
`internal/gateway/` owns server policy and queues. `cmd/gateway/` constructs
that root, its listener, and its persistence ticker. Each connection reaches
policy and diagnostics through its actual parent; it does not acquire a
private copy of the grants or a logger nobody can find.

Bill arrived at the router after the Deephold experience showed the cost of
agents parking inside tool calls to wait for each other. A connected client
handles incoming messages through its application’s `Handler`. Ensemble's
adapter admits them to the actor it already has. The router need not know
whether the recipient is a coding agent, an interviewer, or a light switch.

## 26.2 Noise-KK beneath the cloud

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

This implementation uses `Noise_KK_25519_ChaChaPoly_SHA256` through
`flynn/noise`. The [Noise specification's KK pattern and security
properties](https://noiseprotocol.org/noise.html#interactive-handshake-patterns-fundamental)
explain its authentication and forward-secrecy conditions. Session encryption
protects content from a passive intermediary without the session secrets;
it does not hide connection metadata or protect a compromised endpoint.

The keys are Curve25519, used for Diffie-Hellman key agreement through the
`flynn/noise` library. Some deployments label these `ed25519:` out of habit.
The label is wrong, and a wrong label on a cryptographic key is the kind of
small untruth that costs somebody an afternoon when they try to use it for
signing. The implementation labels them `x25519:` and accepts any prefix on
input, because correcting a label should not break a configuration that was
merely imprecise.

Key management is file-based. `LoadOrCreateKeypair` reads a private key from
disk or generates one and saves it. Clients and the server use the same helper. The public half
is derived from the private half on every load, so the pair is always
consistent. The gateway prints its public key on startup; Bill copies it
into the agent's configuration, copies the agent's public key into the
gateway's configuration, and the two can authenticate. No certificate
authority, no token rotation, no password. The keypair is the identity.

## 26.3 Trust stamped at the boundary

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
Section 26.6 adds the listener, data directory and real public keys:

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
      "public_key": "x25519:A1B2C3...",
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

## 26.4 A library, not a framework

The accepted module is `ensemble`; an outside Go module imports
`ensemble/gateway` and uses a local `replace` pointing at the submitted
module. The `Handler` interface is four methods:

```go
type Handler interface {
    OnDeliver(env Envelope)
    OnConnected(welcome WelcomePayload)
    OnDisconnected(reason string)
    OnError(err error)
}
```

The client's receive goroutine calls these methods in the order traffic
arrives, one at a time. Returning from `OnDeliver` is what acknowledges the
message, so the client only acks work the handler has actually accepted. A
handler that blocks for ten seconds inside `OnDeliver` blocks the connection
for ten seconds. Work that takes real time belongs on a goroutine the handler
owns. A callback must not synchronously call `Send`, `Request`, or another
client RPC: their responses need the same receive goroutine now waiting for
that callback to return.

`NopHandler` implements all four methods as no-ops. Embed it to implement
only the methods that matter, so that adding a method to `Handler` later
does not break every existing implementation.

The `Client` connects, handshakes, and starts the receive goroutine:

```go
client, err := gateway.Connect(ctx, gateway.ClientConfig{
    GatewayURL:    "ws://127.0.0.1:8093/v1/ws",
    AgentID:       "interviewer",
    StaticKey:     key,
    GatewayPubKey: gwPub,
}, handler)
```

That call opens the WebSocket, runs the Noise-KK handshake, waits for the
gateway's welcome frame, and starts the goroutine that calls
`handler.OnDeliver` for incoming traffic. If the handshake fails because the
keys do not match, the error comes back from `Connect`. No partial
connection, no deferred failure. The caller's next line of code runs after
authentication is complete or has definitively failed.

The public types and callable signatures below are part of this exercise's
consumer contract. `WelcomePayload.QueueDepth` is the outstanding count
**after** applying the reconnect receipt. It is an `int`, encoded as
`queue_depth`; it is not the sequence number or the number of registered
agents. An empty connection reports zero, while reconnecting to one pending
message reports one.

```go
type WelcomePayload struct {
    AgentID       string `json:"agent_id"`
    QueueDepth    int    `json:"queue_depth"`
    ServerVersion string `json:"server_version"`
}

type ClientConfig struct {
    GatewayURL       string
    AgentID          string
    StaticKey        Keypair
    GatewayPubKey    []byte
    LastAckedSeq     uint64
    HandshakeTimeout time.Duration
}

type OutboundMessage struct {
    ToAgent        string `json:"to_agent"`
    Content        string `json:"content"`
    ConversationID string `json:"conversation_id,omitempty"`
    ResponseMode   string `json:"response_mode,omitempty"`
}

type ChildRegistration struct {
    AgentID   string   `json:"agent_id"`
    PublicKey []byte   `json:"public_key"`
    Groups    []string `json:"groups"`
}
```

`Keypair` exposes `Private` and `Public` as `[]byte`; each is 32 bytes.
`GenerateKeypair()` and `LoadOrCreateKeypair(path string)` return
`(Keypair, error)`. `EncodePublicKey([]byte)` returns a labeled string;
`ParsePublicKey(string)` returns `([]byte, error)`. Child registration takes
the decoded public bytes; JSON marshaling encodes that byte slice as base64.
Private bytes never belong in registration or gateway configuration.

`Envelope` exposes `ID`, `ToAgent`, `Content`, `ConversationID` and
`ResponseMode` as strings, `Seq` as `uint64`, `Timestamp` as `time.Time`, and
`Source` as a structure with string fields `Type`, `ChannelID`, `UserID`,
`DisplayName` and `Trust`. `ProvenanceHeader(Source)` returns the displayed
header. `ResponseAsync` is the string `async`.

| Public operation | Result and boundary |
| --- | --- |
| `Connect(context.Context, ClientConfig, Handler)` | `(*Client, error)`; a non-nil handler and authenticated welcome are required. Zero handshake timeout selects ten seconds. The context also owns the connection lifetime. |
| `Send(context.Context, OutboundMessage)` | `error`; returns after authorization and queue admission, before the recipient necessarily handles it. |
| `Respond(context.Context, Envelope, string)` | `error`; addresses the authenticated sender and preserves `ConversationID`. Current routing policy applies to replies too. |
| `DeclareContactable(context.Context, []string)` | `error`; replaces accepted sender groups, including an empty set. |
| `RegisterChild(context.Context, ChildRegistration)` | `error`; registers a new key with only groups the parent holds. |
| `Request(context.Context, string, any)` | `(json.RawMessage, error)`; correlated MCP method request on the same socket. |
| `CallTool(context.Context, string, any)` | `(json.RawMessage, error)`; interprets an MCP tool result, including `isError` refusals. |
| `LastAckedSeq()` | `uint64`; last receipt successfully written after handler return. |
| `Close()` / `Done()` | `Close` interrupts I/O; `Done` returns `<-chan struct{}` and closes after the receive callbacks finish. |

For the independently tested Noise layer, `NewNoiseKK(role string, key
Keypair, peerPublic, prologue []byte)` returns `(*NoiseKK, error)`. Roles are
`initiator` and `responder`. `WriteHandshakeMessage([]byte)` and
`ReadHandshakeMessage([]byte)` return `([]byte, error)`;
`IsHandshakeComplete()` returns `bool`. After completion, `Encrypt(data,
associated []byte)` and `Decrypt(data, associated []byte)` return
`([]byte, error)`. Keep each direction's calls ordered and terminate the
connection after authentication or transport failure.

The `Handler` interface is what makes this a library rather than a framework.
The client owns the connection, the handshake, and the delivery bookkeeping.
It owns no opinion about what a message means. The `cmd/gateway-agent` application implements `Handler`
to route envelopes into its actual Ensemble actor. The interviewer implements it to
collect answers. A home automation program would implement it to switch on a
light. The gateway cannot tell the difference, and it does not need to.

## 26.5 Delivery and recovery

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
`Request` sends JSON-RPC `initialize`, `tools/list` or `tools/call` inside
`mcp_request`; a tool call's params are `{"name":"list-groups","arguments":{}}`.
The server's initialization response declares MCP version `2024-11-05`.
Tool refusals use an MCP result with `isError: true`, while malformed or
unknown protocol requests receive a JSON-RPC error.

The public client's `Send` and `Respond` use the acknowledged MCP
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

Every delivered message carries a sequence number. The receive goroutine calls `handler.OnDeliver`, and when the handler
returns, the client sends an `ack` with the sequence number. The ack is the receipt. Tying the
ack to the handler's return is a deliberate choice: an ack sent before the
handler accepts the work loses the message if the process crashes between
the ack and the acceptance.

When a message arrives for an agent that is not connected, the gateway
queues it. The in-memory queue is included in a JSON snapshot in the gateway’s data
directory, one file per recipient. Messages accumulate until the agent reconnects.

On reconnect, the agent sends `last_acked_seq` in the `hello` frame. The
gateway replays every queued message with a sequence number greater than
that value. The client supplies that receipt explicitly when calling `Connect` again;
there is no automatic reconnect loop. Outstanding messages after the receipt
are delivered again. A new connection supersedes the same identity’s old one. The guarantee is at-least-once. Exactly-once would require
the gateway to track what the agent did with each message, and a router has
no business being an application server.

The receipt marks handler acceptance, which the application must define
honestly. The Ensemble adapter calls `Agent.Submit` to admit a turn to the
existing FIFO, then returns. An application-owned worker waits for that
request's reliable result and sends the reply. A prompt can therefore be
acknowledged before the model answers. A crash after that ack can lose work
that the application has not saved. The router cannot turn admission into
durable task completion.

The server snapshots runtime group definitions and memberships,
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

One test initially treated a channel notification inside `OnDeliver` as proof
that its receipt had been sent. The notification happened before the callback
returned. Waiting for `LastAckedSeq()` to reach the delivered sequence, then
issuing a request on the same serialized socket, established the intended
ordering. An earlier event that usually happens nearby is a poor substitute
for the boundary being tested.

## 26.6 The interviewer proves the stack

The interviewer is a standalone binary in `cmd/interviewer/`. It imports
`ensemble/gateway`, holds its own keypair, and reaches Ensemble only through
the gateway's routing. It imports no private Ensemble runtime. Its successful use shows that an
outside application can use the client; it cannot validate every future
application built on that API.

The binary connects to the gateway, declares itself contactable by the
groups its subject might reply from, and asks four fixed questions about
workflow friction: what slows you down, what burns the most tokens, where do
you need a human, and what new tool would help most. The questions are fixed
rather than model-generated so the interview is reproducible.

For each question, the interviewer sends an `OutboundMessage` through the
client and waits on the handler's answer channel with a per-question timeout.
`OnDeliver` puts the envelope on an application-owned channel, and
`NopHandler` covers the remaining callbacks. The waiting code checks both the
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

The adapter adds no second turn loop. `Agent.Submit` supplies a reliable
result channel for each accepted prompt, so a busy Agent can queue another
question without confusing its answer with a shared idle event. Reply workers
wait until the returned Client has been published: incoming delivery can race
with `Connect` returning to its caller. Shutdown cancels work, closes the
socket, waits for receive callbacks to finish, then joins those workers and
saves the Agent. These are application lifetimes the router does not own.

### Running the interview

Build from the accepted module, using a fresh private scratch directory:

```sh
CH26_RUN=$(mktemp -d /tmp/ensemble-ch26.XXXXXX)
export CH26_RUN
cd solutions/ch26
go build -o "$CH26_RUN/gateway" ./cmd/gateway
go build -o "$CH26_RUN/gateway-agent" ./cmd/gateway-agent
go build -o "$CH26_RUN/interviewer" ./cmd/interviewer
"$CH26_RUN/gateway-agent" -key "$CH26_RUN/subject.key" -print-key
"$CH26_RUN/interviewer" -key "$CH26_RUN/interviewer.key" -print-key
printf '%s\n' "$CH26_RUN"
```

The two printed keys are public. Put them in `$CH26_RUN/gateway.json`, using
this complete configuration and replacing the three marked values. Choose an
unused local port; 8093 is an example. If it is occupied, leave that process
alone and use another port in the config and both client URLs.

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
      "public_key": "REPLACE_WITH_SUBJECT_PUBLIC_KEY",
      "groups": ["agents"]
    },
    "interviewer": {
      "public_key": "REPLACE_WITH_INTERVIEWER_PUBLIC_KEY",
      "groups": ["agents"]
    }
  }
}
```

Generate the router's identity and retain its printed public key:

```sh
"$CH26_RUN/gateway" -config "$CH26_RUN/gateway.json" -print-key
```

Use three terminals. Set `CH26_RUN` to the printed scratch path in each, and
set `GATEWAY_KEY` to the router's public key in the two client terminals.
Start the router first:

```sh
"$CH26_RUN/gateway" -config "$CH26_RUN/gateway.json"
```

Its printed key is not a readiness signal. Confirm that this new process
remains alive and owns the chosen listening port before starting clients.
For example, `lsof -nP -iTCP:8093 -sTCP:LISTEN` shows the listener PID; match
it to this launch, rather than merely finding somebody listening there.

Then start the real Ensemble adapter. This version uses the already
configured read-only credential route from optional Chapter 24:

```sh
LLM_VENDOR=openai LLM_MODEL=gpt-6-astra OPENAI_REASONING_EFFORT=high \
OPENAI_BILLING_MODE=codex "$CH26_RUN/gateway-agent" \
  -gateway ws://127.0.0.1:8093/v1/ws -key "$CH26_RUN/subject.key" \
  -gateway-key "$GATEWAY_KEY" -data "$CH26_RUN/subject-data"
```

Gateway itself does not require that billing route. The adapter uses the
existing `ConfigFromEnv` provider configuration; readers using another
working route can keep it. Wait for `Gateway Agent ensemble ready`, then run
the interviewer in the third terminal:

```sh
"$CH26_RUN/interviewer" -gateway ws://127.0.0.1:8093/v1/ws \
  -key "$CH26_RUN/interviewer.key" -gateway-key "$GATEWAY_KEY" \
  -report "$CH26_RUN/interview.md" -timeout 2m
```

Read the four answers in `interview.md`, then stop the subject and router
with Ctrl-C so they can finish their shutdown paths. The subject owns its
conversation save. The router owns the policy and pending-message snapshots.
Register a second public identity and try an allowed subset and a refused
over-grant separately; interviewing one authorized peer does not exercise
child registration or refusal. Likewise, disconnect a recipient with an
unacknowledged delivery and verify its sequence on reconnect.

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
[run and transcript](../solutions/evidence/ch27/live-interview.md)
retain the actual answers; later local diagnostic and receipt-test corrections
do not turn that earlier run into a new one.

## 26.7 What the grader checks

The grader builds the gateway and an outside Go consumer against the submitted
module. The public contract above supplies the names that consumer needs,
including `WelcomePayload.QueueDepth`. It also checks the resolved dependency
graph: the public client must not reach private runtime packages. A compiling
in-tree example alone can hide an `internal/` dependency.

The seven original weights remain. Successful authentication needs wrong-key
controls; authorization needs a refused route beside an allowed one. Delivery
needs a held handler as well as an offline recipient: receipt timing cannot
be inferred merely from a message eventually arriving. Parent registration
needs an independently connecting child and a refused wider grant.

The actual Ensemble interview, parent ownership and diagnostics, administrative
state restoration and stronger lifetime checks retain their separate source
and integration evidence. A targeted mutation establishes sensitivity to that
defect, not general independence or complete security coverage.

Build and grade from the repository root:

```sh
go run ./cmd/grade-edition2 -ch 26 solutions/ch26
```

| # | Behavior | Check | Points |
|---|----------|-------|--------|
| 1 | Noise-KK handshake | Two peers complete mutual authentication over a byte stream. Wrong keys are rejected. | 15 |
| 2 | Prompt envelope with provenance | Delivered body and authenticated provenance agree, with conversation and response mode preserved. | 10 |
| 3 | Gateway server accepts connections | Actual binary welcomes multiple authenticated clients with identity, version and empty/nonempty queue counts. | 15 |
| 4 | Group-based routing | Agent declares contactability by group. Gateway delivers messages only from senders whose group membership the recipient has accepted. Unauthorized sends are rejected. | 15 |
| 5 | Delivery semantics | Handler return precedes ack; offline and unacknowledged messages survive reconnect with the proper sequence boundary; a clean server restart preserves queued delivery. | 15 |
| 6 | Client library with Handler interface | Outside consumer compiles and uses the actual public client; its resolved imports exclude private runtime packages. | 10 |
| 7 | Parent registers child | Connected agent registers a child's public key and grants it a subset of its own groups. Child connects independently via Noise-KK. Gateway rejects group grants the parent does not hold. | 20 |
