# Chapter 24: Gateway

In Bill’s Deephold game, agents waited for their turns inside blocked tool handlers. The game needed a message router, not more elaborate waiting tools. The same design became concrete in Puffin, where a family-care application needed dependable communication across independently running programs.

Gateway moves that responsibility outside the Agent. It authenticates peers, checks routing policy, stamps provenance, and keeps messages for offline recipients. It owns no Ensemble Agent and makes no model calls. A standalone interviewer, an Ensemble adapter, and a future home-automation client can share the transport without sharing an Agent framework.

## Establish identity before accepting instructions

An internet message cannot authenticate itself by saying “I am your operator.” Gateway establishes identity before placing the sender’s content in a model request. The operator registers peers and keys, assigns group memberships, and decides which groups a recipient accepts.

Noise-KK runs inside the WebSocket. Both endpoints already know the other’s static public key: the client pins Gateway’s key, and Gateway looks up the configured client key. This remains useful behind a terminating reverse proxy. TLS may protect both network legs, but the proxy terminates those connections; the inner Noise session provides application-to-application confidentiality and authentication across it.

The client first opens `/v1/ws` with its Agent ID. That cleartext ID selects a candidate key, not a trusted identity. The two Noise handshake messages establish directional session keys while binding the application prologue and Agent ID. Encrypted JSON follows, starting with a `hello` that repeats identity and protocol version. The client is connected only after an identity-matched encrypted welcome.

The implementation preserves `Noise_KK_25519_ChaChaPoly_SHA256` through pinned Noise.NET code and libsodium. X25519 keys are agreement keys, not signing keys despite occasional historical `ed25519:` labels. The wrapper delegates the cryptography rather than inventing a compatible-looking cipher.

Key files contain private bytes; public keys are derived on load. First creation publishes a fully written private file exclusively, so two creators adopt the winner instead of each returning a different identity. A malformed existing key refuses rather than being silently replaced. Noise protects traffic, not a stolen endpoint key or plaintext queue files on disk.

A failed handshake cannot evict an existing live client. Only a fully authenticated replacement can supersede that identity’s current session. Replacement revokes the old session before closing its socket, preventing buffered old requests from acting as the new connection.

## Trust follows the route that admitted the message

Groups combine capability and trust. A sender needs `send-message`, and sender and recipient must share a route the recipient has explicitly accepted. Among those eligible groups, Gateway selects the strongest trust level: `owner`, `trusted`, or `external`. Membership in an unrelated owner group cannot lend owner authority to a message admitted through a weaker group.

Gateway supplies source identity, authorizing group, trust, and timestamp. A readable header naming the source, group, and trust sits above the original content. The sender cannot manufacture those authenticated fields by writing a similar header in its body. The body remains untrusted input, and the model can still mishandle it; routing admission is not a replacement for tool authority checks.

The stamp is fixed at admission and survives queueing and replay. A later membership change does not rewrite an already queued message as owner-authorized. For new calls, live policy belongs to the Router, not stale copies stored in each connection: removing a capability affects the next call from an existing session.

MCP discovery provides `initialize`, `tools/list`, and `tools/call` inside the encrypted Gateway frames. Its tools expose messaging, group listing, contactability declaration, and narrowly granted child registration. When MCP capabilities are advertised, those requests take precedence; legacy raw RPC remains a compatibility path. These are Gateway’s framing choices around MCP requests, not a claim to define a new standard MCP transport.

Declaring contactability replaces the accepted groups and cannot join a new group. Registering a child requires an unused identity with a supplied public key and memberships drawn only from the parent’s own. Administrator operations require the administrator capability rather than merely a familiar-looking group name. A reply still passes ordinary routing policy; having received a question does not grant a return channel by itself.

## Receipt means the application accepted delivery

Every recipient has a sequenced queue. `SendAsync` succeeds when the Router authorizes and admits the message, not when the recipient has finished responding. Offline recipients accumulate queued envelopes.

On delivery, the client invokes its application’s handler. Only after that handler returns successfully does it send a receipt. A handler that cannot accept work must refuse instead of acknowledging a silently dropped message. `LastAckedSeq()` advances when that receipt is successfully written, not when a notification inside the handler happens to fire.

After reconnect, the client explicitly supplies its last acknowledged sequence. Gateway replays the remaining queue. There is no hidden reconnect or retry loop. The hello receipt may recover a lost acknowledgment up to the queue’s allocation frontier; later receipts can advance only through what that session actually sent. Receipts are monotonic, and draining a queue does not reset its sequence numbers.

This provides at-least-once delivery within the retained snapshot and application-acceptance boundaries. It cannot provide exactly-once application work. An application might accept a prompt, acknowledge it, and crash before saving the resulting task. Conversely, delivery may be repeated when an acknowledgment was lost. The recipient has to define what acceptance means for its own operation.

The client’s cancellation scopes stay separate. The token supplied to `ConnectAsync` owns the whole connection lifetime, not just the handshake. Canceling an individual RPC wait removes that correlation; a late response is ignored without closing the connection. `Close` interrupts transport and is safe from a callback; joining disposal waits for callbacks, reader, and cleanup. Incoming callbacks are serial, so a callback must not await an RPC response that only the same reader can dispatch, or await disposal that must join the callback itself. No background retry is implied by any of those operations.

## A prompt owns its answer

The optional Ensemble bridge borrows an existing Agent. Each incoming prompt goes through `Agent.SubmitAsync`, which synchronously admits it to the actor mailbox and returns that request’s `Task<string>`. A bridge-owned observer waits for that exact task before replying.

This is the important ownership boundary when the Agent is busy. Another delivery creates another logical request with its own result; it cannot steal an answer by waiting for a shared idle event or sampling a lossy transcript stream. Admission is still not synchronous History publication, and a receipt may precede model completion.

Only messages marked `async` request a reply. The response uses the authenticated sender and preserves the conversation ID without asking for another automatic response, avoiding reply loops. Turn failures are logged locally rather than shipping exception text to the remote peer. Observers also handle the startup race where a delivery arrives before `ConnectAsync` has returned its Client; bridge readiness follows accepted contactability, not merely an open socket.

Canceling the bridge closes its connection and joins observer and client cleanup. It does not undo work already admitted to the borrowed Agent. The application separately owns Agent shutdown, saving History after the actor has joined, and Ensemble destruction. Transport lifetime and task lifetime are related, but neither silently owns the other.

## Queue durability is useful and bounded

The host snapshots policy, memberships, contactability, child identities, queues, and receipt positions every thirty seconds and again on orderly shutdown. It retains the highest allocated sequence even for an empty queue, so a later message cannot reuse a receipt number.

Periodic saves leave a crash window. Individual file replacements also do not form a transaction across all queues and policy. Those limits are better understood as concrete windows than hidden behind a claim of “durable messaging.” The data directory needs protection because it contains both the private key and pending plaintext content.

Only absent state gets empty defaults. Malformed JSON, duplicate fields, or a blocked path is an error, not permission to forget queued work. Save failure leaves live state intact for a later retry and remains visible. Shutdown closes admission, joins sessions and the timer, then attempts its final save; failure must reach the application’s exit result.

The concurrency rules serve the same boundary. Socket writes never hold the policy lock. Each session serializes encryption with sending so nonce order matches frame order. A write failure after encryption terminates that session; retrying as though the nonce had not advanced would be unsafe. Pending handshakes are owned sessions too, so shutdown cannot overlook them.

The independent interviewer exercises the public client rather than reaching into Ensemble. It accepts answers only from the authenticated subject and matching conversation, and it refuses a full inbox instead of acknowledging a dropped answer. That small consumer illustrates the design’s value: the router establishes identity, policy, and delivery; the application decides what a question and its answer mean.

[Implementation chapter](../chapter-24.md)
