# Chapter 27 — Gateway

Status: DESIGN DRAFT. Discuss with Bill before building.

## Thesis

The gateway is core infrastructure: a standalone edge service that sits at the
trust boundary between the outside world and the agent. It routes prompts from
channels (Discord, webhooks, scheduled jobs, other agents) to agents, and routes
responses back. It authenticates with Noise-KK (mutual auth + encryption beneath
Cloudflare's TLS termination), enforces config-declared routing policy, and
queues messages when agents are offline.

The gateway owns no agents. It is a router. The Ensemble is a gateway CLIENT.

This chapter builds the gateway itself and a public client library. Channels,
the scheduler, and everything that consumes the gateway come in later chapters.

## What has to come first

Everything Bill wants to build depends on the gateway:

- Connectors (Discord, home automation, email) — need a gateway to route through
- Scheduler — fires prompts into the gateway's router
- Self-improvement loop — an interviewer agent connecting through the gateway

This chapter is the foundation. It must be solid, simple, and usable as a Go
module by anyone.

## Architecture

```
┌──────────────────────────────────────┐
│         Gateway (cmd/gateway/)       │
│                                      │
│  ┌────────────────────────────────┐  │
│  │   Config (routing policy)      │  │
│  │   channel X → agent Y         │  │
│  └──────────┬─────────────────────┘  │
│             │                        │
│  ┌──────────▼─────────────────────┐  │
│  │   Router                       │  │
│  │   Envelope in → lookup route   │  │
│  │   → deliver to agent           │  │
│  └──────────┬─────────────────────┘  │
│             │                        │
│  ┌──────────▼─────────────────────┐  │
│  │   Agent Manager                │  │
│  │   Connections, queuing,        │  │
│  │   ack tracking, resume         │  │
│  └──────────┬─────────────────────┘  │
│             │                        │
│     Noise-KK WebSocket connections   │
│             │                        │
└─────────────┼────────────────────────┘
              │
    ┌─────────┴──────────┐
    │                    │
┌───▼────┐         ┌────▼─────┐
│Ensemble│         │Interviewer│
│(agent) │         │(standalone│
│        │         │ agent)    │
└────────┘         └──────────┘
```

Key invariant: **the config file declares all routing.** No agent
self-registers what prompts it can handle.

## Package Layout

```
agent/
  gateway/              # PUBLIC client library
    client.go           # Client, Connect, Send, Respond, RegisterChild
    handler.go          # Handler interface (implemented by consumers)
    noise.go            # Noise-KK handshake + encrypted framing
    envelope.go         # Envelope, Source, trust levels, provenance
    protocol.go         # Frame types, constants
    keygen.go           # Key generation utility
    client_test.go

  internal/gateway/     # Server-side (private to cmd/gateway)
    server.go           # WebSocket server, Noise-KK accept
    connections.go      # Agent connection manager
    router.go           # Group-based routing, contactability
    groups.go           # Group definitions, membership, capabilities
    queue.go            # Offline message queue, ack tracking
    mcp.go              # MCP tool surface for gateway admin
    config.go           # Server config parsing
    persistence.go      # JSON persistence for dynamic state
    server_test.go

  cmd/gateway/
    main.go             # Gateway binary

  cmd/interviewer/      # Exercise: standalone interviewing agent
    main.go
```

The split: `agent/gateway/` is the public client library anyone imports.
It has ZERO imports from `agent/internal/` — structurally separable into
its own Go module. `agent/internal/gateway/` is the server.
`cmd/gateway/` is the binary.

## Noise-KK

### Why Noise-KK, not TLS client certs

Cloudflare terminates TLS. The gateway sees plaintext HTTP from Cloudflare's
edge. So TLS authenticates Cloudflare to the client, not the gateway to the
agent. We need end-to-end encryption and mutual authentication beneath
Cloudflare's layer.

Noise-KK provides:
- **Mutual authentication**: both sides prove identity via static ed25519 keys
- **Forward secrecy**: ephemeral keys per session
- **Simplicity**: two-message handshake, no CA, no cert rotation
- **0-RTT not needed**: WebSocket upgrade happens first, then Noise inside it

### Handshake over WebSocket

1. Agent connects to `wss://gateway.example.com/v1/ws` (TLS to Cloudflare)
2. WebSocket upgrade completes (plaintext from gateway's perspective)
3. Noise-KK handshake as the first two WebSocket binary messages
4. All subsequent frames are Noise-encrypted JSON

### Key management

- Each agent has an ed25519 keypair (X25519 derived for Noise key agreement)
- Gateway config lists every agent's public key
- Agent config has the gateway's public key
- No passwords, no API keys, no tokens, no certificates

### Implementation

Adapt from `openadp/sdk/go/common/noise_nk.go` (519 lines, proven). Changes:
- NK → KK (both sides provide StaticKeypair AND PeerStatic)
- Wrap in WebSocket binary frames instead of HTTP request/response
- `github.com/flynn/noise` library handles KK natively

## Prompt Envelope

The canonical message format flowing through the gateway:

```go
type Envelope struct {
    MessageID      string    `json:"message_id"`       // UUID v7
    Seq            uint64    `json:"seq"`               // gateway-assigned sequence
    AgentID        string    `json:"agent_id"`          // destination agent
    Source         Source    `json:"source"`
    Timestamp      time.Time `json:"timestamp"`
    ConversationID string    `json:"conversation_id"`   // optional thread
    Content        string    `json:"content"`
    ResponseMode   string    `json:"response_mode"`     // async|sync|fire_and_forget
}

type Source struct {
    Type        string `json:"type"`         // "webhook", "discord", "scheduler", "agent"
    ChannelID   string `json:"channel_id"`
    UserID      string `json:"user_id"`
    DisplayName string `json:"display_name"`
    Trust       string `json:"trust"`        // "owner"|"trusted"|"external"
}
```

### Trust Levels

| Level | Meaning | Example |
|-------|---------|---------|
| `owner` | Bill himself | Noise from Bill's device, cron jobs Bill configured |
| `trusted` | Authenticated service Bill controls | Webhook from a game server |
| `external` | Untrusted users on a trusted platform | Family member on Discord |

Trust is immutable — set by the gateway based on the channel config, never
by message content. The gateway prepends a provenance header when delivering
to the agent:

```
[GATEWAY source=webhook-game user=server trust=trusted]

Your turn in the dungeon. The goblin attacks.
```

### Response Modes

| Mode | Behavior |
|------|----------|
| `async` | Agent responds whenever ready. Gateway routes response back. |
| `sync` | Source blocks until agent responds or timeout. |
| `fire_and_forget` | No response expected. |

## WebSocket Protocol

All frames are Noise-encrypted. JSON payloads inside the encrypted stream.

### Agent → Gateway

| Frame | Fields | Purpose |
|-------|--------|---------|
| `hello` | agent_id, version, last_acked_seq | Identify + resume |
| `ack` | seq | Confirm delivery |
| `response` | message_id, content, metadata | Reply to a delivered prompt |
| `ping` | | Keepalive |

### Gateway → Agent

| Frame | Fields | Purpose |
|-------|--------|---------|
| `welcome` | agent_id, queued_count, server_version | Connection confirmed |
| `deliver` | seq, envelope | Deliver a prompt |
| `pong` | | Keepalive response |
| `error` | code, message | Protocol error |

### Delivery Semantics

- **At-least-once**: unacked messages redelivered on reconnect
- **Dedup by message_id**: agent maintains a small seen-set
- **Resume**: agent sends `last_acked_seq` in hello; gateway replays from seq+1
- **Queue**: if agent is offline, prompts are queued (JSON persistence)

## Capability Model

### Groups as ACLs

Groups are the gateway's authorization primitive. An agent's capabilities —
what it can do, who it can contact, what it can administer — flow from its
group memberships. The gateway checks group membership on every operation.

**Base memberships** live in gateway config, keyed by public key. These are
the seed. At runtime, agents can declare contactability, grant memberships to
children, and (with the right capability) administer groups — all through
MCP tools tunneled over the Noise-KK WebSocket.

### Inheritance: narrow only

When an agent spawns a child through the gateway, it can grant the child a
SUBSET of its own groups. Same pattern as sandbox (ch23) — you can narrow,
never widen. An agent can only grant memberships to agents it spawns.

### Dynamic contactability

A connected agent declares at runtime which groups can contact it. This is
the agent controlling its own inbound access, not the config hardcoding
who talks to whom.

### Self-protecting administration

Group administration is itself a capability. Editing a group requires the
`group-admin` capability, which belongs to a root-level group controlled
carefully in config. The system bootstraps from config-defined memberships
and is self-protecting from that point on.

### Gateway MCP Service

The gateway exposes its administration API as MCP tools, tunneled over the
existing Noise-KK WebSocket (same pattern as ch12/ch13 MCP-over-WS). Agents
interact with gateway capabilities through normal tool calls:

| Tool | Requires | Purpose |
|------|----------|---------|
| `gateway_declare_contactable` | connected | Accept messages from listed groups |
| `gateway_grant_membership` | parent of target + own membership | Grant child a subset of own groups |
| `gateway_revoke_membership` | parent of target | Revoke child's group membership |
| `gateway_send_message` | membership in a group the target accepts | Send a message to another agent |
| `gateway_list_agents` | connected | List agents and their contactability |
| `gateway_create_group` | `group-admin` | Create a new group |
| `gateway_edit_group` | `group-admin` | Modify group properties |
| `gateway_delete_group` | `group-admin` | Delete a group |

This means agents manage the gateway the same way they use any other MCP
server — no special protocol, no out-of-band admin CLI.

### Example: Interviewer Exercise

1. Config defines groups: `admin`, `interviewers`, `agents`
2. Config grants Ensemble base memberships: `admin`, `agents`
3. Ensemble connects, declares: "I accept contact from group `interviewers`"
4. Ensemble spawns interviewer, grants it `interviewers` membership
5. Interviewer connects, inherits `interviewers` from parent grant
6. Interviewer calls `gateway_send_message` to Ensemble → gateway checks:
   is interviewer in a group Ensemble accepts? Yes (`interviewers`) → deliver
7. Ensemble responds → gateway routes back

## Config

Gateway config defines base infrastructure: groups, agent identities, and
seed memberships. Runtime state (dynamic memberships, contactability
declarations) is persisted separately.

```json
{
  "gateway": {
    "listen": ":8093",
    "data_dir": "./data"
  },
  "agents": {
    "ensemble": {
      "display_name": "Ensemble",
      "public_key": "ed25519:base64...",
      "groups": ["admin", "agents"]
    },
    "interviewer": {
      "display_name": "Interviewer",
      "public_key": "ed25519:base64..."
    }
  },
  "groups": {
    "admin": {
      "description": "Full gateway administration",
      "capabilities": ["group-admin", "route-admin"]
    },
    "agents": {
      "description": "Standard agent group",
      "capabilities": ["send-message", "declare-contactable"]
    },
    "interviewers": {
      "description": "Agents that conduct interviews",
      "capabilities": ["send-message"]
    }
  }
}
```

Base memberships are assigned in config by public key. Dynamic memberships
(granted by parents at runtime) are persisted to the data directory and
survive gateway restarts.

## Client Library (agent/gateway/)

Public API for any Go program to connect to a gateway:

```go
// Client connects to a gateway and handles the Noise-KK handshake,
// message delivery, and response routing.
type Client struct { ... }

// Connect establishes a Noise-KK encrypted WebSocket connection.
func Connect(ctx context.Context, cfg ClientConfig) (*Client, error)

// Send delivers a message to another agent through the gateway.
func (c *Client) Send(ctx context.Context, msg OutboundMessage) error

// Receive returns the next delivered envelope (blocks).
func (c *Client) Receive(ctx context.Context) (*Envelope, error)

// Respond sends a response to a previously delivered message.
func (c *Client) Respond(ctx context.Context, messageID string, content string) error

// Close gracefully disconnects.
func (c *Client) Close() error

type ClientConfig struct {
    GatewayURL    string          // wss://gateway.example.com/v1/ws
    AgentID       string
    StaticKey     noise.DHKey     // this agent's ed25519/X25519 keypair
    GatewayPubKey []byte          // gateway's known public key
}

type OutboundMessage struct {
    ToAgent        string
    Content        string
    ConversationID string // optional
    ResponseMode   string // async|sync|fire_and_forget
}
```

This is the piece that makes the gateway reusable infrastructure rather than
just a program.

## Exercise: The Interviewer Agent

The student builds a standalone agent (`cmd/interviewer/`) that connects to
Ensemble through the gateway to interview it. The interviewer:

1. Connects to the gateway using the client library
2. Sends interview questions to Ensemble (via gateway routing)
3. Receives Ensemble's responses
4. Writes a markdown doc per interview session

### What the interviewer asks

The interviewer is a self-improvement agent. It asks Ensemble about:
- Friction points in its current workflow
- What tasks take the most turns / most tokens
- Where it gets stuck or needs human intervention
- Opportunities for new tools or skills

### What it produces

A markdown file per interview:

```markdown
# Interview: Ensemble — 2026-10-11

## Friction Points
- [Finding 1 from the interview]
- [Finding 2]

## Improvement Opportunities
- [Opportunity 1]
- [Opportunity 2]

## Raw Transcript
[Q&A pairs]
```

### Why this exercise works

1. **Proves the gateway**: two separate binaries communicating over Noise-KK
2. **Proves the client library**: the interviewer imports `agent/gateway/`
3. **Real use case**: Bill actually wants automated agent interviewing
4. **Tangible output**: the grader can verify the doc was written
5. **No mock needed**: the interviewer is a real agent with a system prompt,
   not a test script

### Grader checks

1. Gateway binary builds and starts
2. Client library connects with Noise-KK handshake
3. Ensemble connects as a second client
4. Interviewer sends a message → gateway routes to Ensemble
5. Ensemble responds → gateway routes back to Interviewer
6. Delivery semantics: ack, and queue-on-offline verified
7. Interview doc written with findings

## Implementation Stages

Left to the coding model. High-level order:

1. Noise-KK (adapt from OpenADP, KK handshake over byte streams)
2. Protocol + Envelope + trust levels
3. Gateway server core (WebSocket, connection manager, config, groups)
4. Client library with Handler interface
5. Group-based routing + delivery + ack/queue/resume
6. MCP tool surface for gateway admin
7. Ensemble integration (implement Handler)
8. Interviewer agent (cmd/interviewer/)
9. Composition root (cmd/gateway/main.go)

## Settled Decisions

1. **Routing is group-based, not static routes.** Agents declare contactability
   by group. The gateway checks membership on delivery. No config-declared
   point-to-point routes.

2. **Base memberships from config, dynamic from parent grants.** Config seeds
   the system. At runtime, parents grant children subsets of their own groups.
   Narrow only, never widen. Only children you spawn.

3. **Gateway admin is MCP tools over Noise-KK WebSocket.** No separate admin
   CLI or HTTP API. Agents manage the gateway through the same tool-call
   mechanism they use for everything else. Group administration requires the
   `group-admin` capability.

4. **Noise-KK is required.** Cloudflare terminates TLS; end-to-end encryption
   beneath it is non-negotiable.

5. **Trust derives from group properties.** Each group carries a trust level.
   When an agent sends a message, the trust level on the envelope comes from
   the group the sender belongs to. No separate trust axis.

6. **Every agent gets a keypair.** The parent registers the child's public key
   with the gateway (via MCP tool). No session vouching — every agent
   authenticates independently via Noise-KK.

7. **Client library uses a callback interface.** The gateway client code
   defines a public `Handler` interface. On initialization, the caller passes
   a Handler implementation. A goroutine in the client listens for incoming
   traffic and calls Handler methods. Ensemble implements Handler to route
   traffic to/from its agents. Any other Go program can implement Handler
   for its own purposes.

8. **Ensemble is just another client.** The Ensemble object links in the
   gateway client code and registers with the gateway over WebSocket, same as
   any other client. No special status.

## Client Library Architecture

```go
// Handler is implemented by the application that uses the gateway client.
// The client's receive goroutine calls these methods on incoming traffic.
type Handler interface {
    // OnDeliver is called when the gateway delivers a message to this agent.
    OnDeliver(envelope Envelope)

    // OnError is called on protocol or connection errors.
    OnError(err error)

    // OnConnected is called after successful Noise-KK handshake.
    OnConnected(welcome WelcomeFrame)

    // OnDisconnected is called when the connection drops.
    OnDisconnected(reason string)
}

// Client manages the Noise-KK encrypted WebSocket connection to a gateway.
type Client struct { ... }

// Connect establishes the connection, performs the Noise-KK handshake,
// and starts the receive goroutine that calls handler methods.
func Connect(ctx context.Context, cfg ClientConfig, handler Handler) (*Client, error)

// Send delivers a message to another agent through the gateway.
func (c *Client) Send(ctx context.Context, msg OutboundMessage) error

// Respond sends a response to a previously delivered envelope.
func (c *Client) Respond(ctx context.Context, messageID, content string) error

// RegisterChild registers a child agent's public key with the gateway.
// The child inherits a subset of this agent's group memberships.
func (c *Client) RegisterChild(ctx context.Context, reg ChildRegistration) error

// DeclareContactable tells the gateway which groups can send messages
// to this agent.
func (c *Client) DeclareContactable(ctx context.Context, groups []string) error

// Close gracefully disconnects.
func (c *Client) Close() error
```

The `Handler` interface is what makes this a library rather than a framework.
Ensemble implements it to route envelopes into its actor system. A standalone
interviewer agent implements it to receive interview responses. A home
automation agent implements it to dispatch commands to devices.

## Ensemble Integration

Ensemble implements `gateway.Handler`:

```go
type ensembleGatewayHandler struct {
    en *Ensemble
}

func (h *ensembleGatewayHandler) OnDeliver(env gateway.Envelope) {
    // Route the envelope to the target agent's actor as an inbound message.
    // The envelope's provenance header tells the agent who sent it and at
    // what trust level.
    agent := h.en.GetAgent(env.AgentID)
    if agent != nil {
        agent.DeliverGatewayMessage(env)
    }
}
```

This is no different from how the GUI delivers stdin messages to the actor —
it's another inbound source, using the same actor mailbox path.

## Open Questions

1. **Conversation threading**: The gateway assigns conversation_ids? Or the
   sender does? For interviews, the interviewer should own the conversation.

2. **Noise library**: Copy openadp/noise_nk.go and adapt, or depend on it
   as a module? Copy seems right for the book (self-contained).

3. **Client library module boundary**: Keep in `agent/` module for the book
   (simpler), but design with zero imports from `agent/internal/` so it's
   structurally separable. Or make it a separate Go module now?

4. **MCP tool surface for gateway admin**: The full set of MCP tools the
   gateway exposes (group CRUD, membership grants, contactability, agent
   listing). Need to enumerate before building.
