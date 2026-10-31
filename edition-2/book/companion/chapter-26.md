# Chapter 26: Gateway

Bill built Deephold as a fantasy role-playing game for AI agents. They explored dungeons and fought monsters, but, in his [account](../edition-3/chapter-26.md), did not appear to have fun. More consequentially, each agent waited for its turn by blocking inside a tool handler while the other players moved. Reading about OpenClaw clarified the need for a dedicated message router. The design came together when Bill and CodeRhapsody built Puffin to help a family member who needed daily care.

The underlying problem extends beyond games. Local delegation gives children an owner and a route back to their parent. A scheduled task, Discord adapter or agent on another machine has neither that relationship nor the operator's terminal. It needs a way to reach the agent without acquiring the operator's authority merely by sending convincing text.

The gateway supplies that boundary. It authenticates peers, enforces routing policy, stamps provenance and queues messages for offline recipients. It owns no Agent, makes no model calls and stores no conversation history. Ensemble connects as a client; so can a program that never uses a model. As Chapter 25 explains, the Edition 3 construction continues from the accepted Chapter 24 baseline, not from an unrecorded crossover repair pass.

## Authentication across an intermediary

A reverse proxy complicates the usual assumption that TLS connects the two applications. In a proxied Cloudflare deployment, [client-to-edge and edge-to-origin are separate connections](https://developers.cloudflare.com/ssl/concepts/). The origin leg can also use TLS. The problem isn't necessarily plaintext on the network: the proxy terminates the client's encryption and can inspect application traffic.

Gateway therefore runs Noise-KK inside the WebSocket. Both applications know the other's static public key beforehand. The client pins the gateway key; the gateway holds the operator-configured key for each agent. The agent ID in the opening URL selects a candidate key, not an authenticated identity. Two handshake messages combine static and ephemeral Diffie-Hellman operations, bind the protocol and agent ID through a shared prologue, and derive directional session keys. An encrypted `hello` then repeats the identity and protocol version before the gateway returns its welcome.

The implementation uses `Noise_KK_25519_ChaChaPoly_SHA256` through `flynn/noise`; the [Noise specification](https://noiseprotocol.org/noise.html#interactive-handshake-patterns-fundamental) defines the pattern's security conditions. A passive intermediary without session secrets cannot read the encrypted application frames. Connection metadata remains visible, and endpoint compromise remains outside that protection. Authentication or transport failure ends the connection rather than leaving it partially usable.

These are X25519 key-agreement keys, not Ed25519 signing keys. The key helper saves the private half and derives the public half on load. Only public keys belong in peer configuration or child registration. File-based identity avoids a certificate authority or bearer-token exchange, but still leaves the operator responsible for protecting private keys and configuring the right peers.

## Identity is not authority

The operator defines groups with trust levels and capabilities, then assigns agents their base memberships. A valid key proves which configured peer connected. It doesn't make that peer an owner, and permission to send doesn't imply permission to contact every recipient.

Each recipient declares which sender groups it accepts. A route requires a group the sender holds and the recipient accepts; a refused route never enters the recipient's context. Contactability is replaceable, including with an empty set. Replies pass through current routing policy too, so accepting questions is not by itself enough to establish a working conversation in both directions.

For an admitted message, the gateway selects the strongest trust level among those shared, authorizing groups: `owner`, `trusted` or `external`. An unrelated administrative membership cannot raise the trust of a message admitted through a less trusted group. The envelope records the sender as an agent, the authorizing group as its channel and the registered identity as its user. It also carries the message ID, destination, sequence, timestamp, conversation ID and response mode. The gateway stamps provenance at admission and preserves that stamp during queued delivery and replay.

The readable header sits above the sender's body, but it remains text the model must interpret. A body claiming `trust=owner` supplies no authenticated authority. Conversely, a genuine gateway stamp authenticates provenance, not the model's obedience to the intended authority rules. Routing enforcement happens before inference; supervision and capability enforcement still govern what the recipient may do afterward.

The Puffin reminder example exposes another boundary. A Discord adapter might be trusted to identify a family member while that person's request has only external authority. The present gateway authenticates the adapter's agent identity, not the Discord user. A future adapter needs an explicit mapping between them; sharing a socket doesn't make them the same principal.

## Delegation narrows; administration changes policy

A connected parent can register a new child's public key and grant only groups the parent itself holds. A parent with `admin` and `agents` can grant `agents` alone, but not an `interviewers` group it lacks. Registration cannot replace an existing identity's key. This creates a network identity, not an Ensemble task child; process spawning and task ownership remain separate concerns.

Runtime policy administration is a different authority. The operator controls the root-level `group-admin` capability. Its holders can manage groups in the `runtime:` namespace and replace additional runtime memberships, leaving configured base memberships intact. Deleting a runtime group removes both grants and contactability references. Configured groups are immutable at runtime: otherwise an apparently successful policy change could silently disappear on restart.

Administration uses MCP requests tunneled over the same authenticated Noise connection. There is no second administrative authentication surface. These are Gateway-specific frames carrying JSON-RPC, not a claim to define a new standard MCP transport. Capability checks still apply to the requested operation; an authenticated socket alone isn't administrative permission.

## Acceptance must leave reception running

The public `gateway/` package owns the connection, handshake and delivery bookkeeping without importing private runtime packages. Server policy and queues belong to `internal/gateway/`; the server binary owns the listener and persistence lifecycle. Connections reach policy and diagnostics through that server ownership rather than acquiring disconnected copies. This separation lets an outside Go program use the client without taking on Ensemble's runtime.

An application's `Handler` receives deliveries and connection notifications. The client invokes callbacks serially on its receive goroutine. Returning from `OnDeliver` means the application has accepted that delivery; only then does the client write its acknowledgment. It does not mean model work has finished or been durably saved.

This ordering avoids acknowledging a message before the application accepts it, but it places a constraint on the handler. Waiting there for final model work blocks reception for the entire turn. Calling `Send`, `Request` or another client RPC synchronously from the callback is worse: its response needs the receive goroutine that is now waiting for the callback to return. Long-running work and RPCs belong to application-owned workers.

The Ensemble adapter therefore calls `Agent.Submit` to admit the prompt to the existing FIFO, then returns. A worker waits for that request's reliable result and sends the reply. There is no second turn loop and no shared idle notification that could confuse one queued question's answer with another's. `Send` itself returns after authorization and queue admission, not after recipient acceptance; `Respond` addresses the authenticated sender and preserves the conversation ID.

Connection and worker lifetimes also meet at a small race: a delivery callback can run before `Connect` returns the Client to its caller. Reply workers wait until that Client has been published. On shutdown, the adapter cancels work, closes the socket, waits for receive callbacks, joins its workers and saves the Agent. The router cannot own those application lifetimes.

## What a receipt can recover

Every delivery has a sequence number. After the handler returns, the client writes an acknowledgment and exposes the last successfully written receipt through `LastAckedSeq()`. Reconnection is explicit: the caller supplies that receipt to a new `Connect`, and the gateway replays queued messages with higher sequences. A new connection supersedes the same identity's old one. The welcome reports the authenticated agent ID, server version and outstanding queue depth after applying the receipt, not the last sequence number.

Offline and unacknowledged messages remain queued. This is an at-least-once delivery boundary, not exactly-once application execution. If acceptance happens but the receipt is lost, replay can produce a duplicate. If the receipt succeeds and the application then crashes before saving its admitted task, the model work can be lost. The application must decide what acceptance commits to; the router cannot infer durable completion from a callback return.

The source records a test that confused these boundaries. A channel notification inside `OnDeliver` was treated as evidence that the acknowledgment had been sent, although the callback hadn't yet returned. Waiting for `LastAckedSeq()` to reach the delivery's sequence, then issuing a request on the same serialized socket, established the intended order. Receipt timing cannot be inferred from eventual delivery alone.

The server snapshots queues, receipt positions, child identities, contactability and runtime policy every thirty seconds and on orderly shutdown. It retains the highest allocated sequence even after a queue drains, preventing a later message from reusing a number already covered by a receipt. Periodic saving matters because a killed process cannot run its shutdown hook.

These snapshots are not arbitrary-crash durability. They leave a loss window, and files are replaced individually rather than transactionally across policy and queues. The data directory contains both the gateway's private key and plaintext pending messages, so it needs private application-state protection. Noise encrypts transport, not queue files.

## What the interview established

The standalone interviewer uses only the public client to reach the real Ensemble adapter through the router. It asks fixed questions, keeping the interview reproducible without another model generating prompts. Its handler hands answers to application-owned waiting code, which checks both conversation ID and authenticated subject identity. A matching conversation ID alone would not establish who answered.

The [recorded second-edition run and transcript](../../solutions/evidence/ch27/live-interview.md) used three built processes and four real Astra requests. The source reports that all three exited normally, incoming questions retained stamped identity in the saved Agent conversation, and answers returned through the router. This is historical evidence of that integrated path, not fresh Edition 3 execution.

The subject described repeated file reading and verbose tool output as token costs and proposed focused code navigation. It had no coding tools and performed no coding task, so those answers are model opinions, not measured workflow friction. All four cache-read counts were zero. The successful interview establishes neither cache benefit nor productivity improvement, and it cannot establish refused routing, child-grant narrowing or reconnect behavior without their separate observations.
