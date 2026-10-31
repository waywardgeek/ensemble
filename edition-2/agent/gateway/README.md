# Gateway client and applications

The optional `ensemble/gateway` package has no private-runtime imports. A Gateway
is a standalone router, not an Ensemble: it authenticates keys, enforces groups,
and retains unacknowledged messages. `cmd/gateway-agent` is a small actual Ensemble
application using the public Agent actor API; `cmd/interviewer` imports only the
public Gateway library.

From the main module:

```sh
go build -o /tmp/gateway ./cmd/gateway
go build -o /tmp/gateway-agent ./cmd/gateway-agent
go build -o /tmp/interviewer ./cmd/interviewer
/tmp/gateway-agent -key /tmp/subject.key -print-key
/tmp/interviewer -key /tmp/interviewer.key -print-key
```

Choose an unused listener and a private data directory. Put the printed public
keys in `gateway.json`; keep each private key with its own process.

```json
{
  "gateway": {"listen": "127.0.0.1:8093", "data_dir": "/tmp/gateway-data"},
  "groups": {
    "agents": {
      "description": "Interview participants",
      "capabilities": ["send-message", "declare-contactable"],
      "trust": "trusted"
    }
  },
  "agents": {
    "ensemble": {"public_key": "x25519:SUBJECT_PUBLIC_KEY", "groups": ["agents"]},
    "interviewer": {"public_key": "x25519:INTERVIEWER_PUBLIC_KEY", "groups": ["agents"]}
  }
}
```

`/tmp/gateway -config gateway.json -print-key` creates the Gateway identity and
prints its public key. Start these in three terminals, replacing `GATEWAY_KEY`
with that public value and using the same chosen URL in both clients:

```sh
/tmp/gateway -config gateway.json
LLM_VENDOR=openai LLM_MODEL=gpt-6-astra OPENAI_REASONING_EFFORT=high \
OPENAI_BILLING_MODE=codex /tmp/gateway-agent \
  -gateway ws://127.0.0.1:8093/v1/ws -key /tmp/subject.key \
  -gateway-key GATEWAY_KEY -data /tmp/gateway-subject
/tmp/interviewer -gateway ws://127.0.0.1:8093/v1/ws \
  -key /tmp/interviewer.key -gateway-key GATEWAY_KEY -report /tmp/interview.md
```

The explicit Codex route uses the already signed-in read-only credentials
explained in optional Chapter24. Other configured providers use the same
`ConfigFromEnv` interface. The subject is a real Ensemble actor with its own
saved conversation; this interview application exposes no coding tools. Its
answers describe model impressions, not independently measured workflow costs.

An outside module can import `ensemble/gateway` through a local `replace` for
this module. `Connect` takes a `ClientConfig` and a non-nil `Handler`, authenticates,
then returns after welcome. Call `DeclareContactable`, `Send`, `Respond`, and
`RegisterChild` with caller-owned contexts. `Request` exposes MCP initialize,
tools/list and tools/call on that same socket. `CallTool` interprets tool refusals.

Callbacks run sequentially on the receive goroutine. Return from `OnDeliver`
accepts the message and causes its ack. Long work belongs to application-owned
goroutines; do not call blocking client RPCs inside callbacks. Delivery may begin
before the caller assigns the Client returned by Connect. The Ensemble adapter
uses a readiness signal for its asynchronous reply workers and joins them at
shutdown.

`Close` interrupts the socket; `Done` closes when callbacks have finished.
Reconnect explicitly with the same identity and `LastAckedSeq()`. A new session
supersedes the old connection. Receipts are cumulative in recipient order;
unacknowledged messages can be delivered again. Admission to an actor is not a
promise of durable execution or exactly-once effects.

The server snapshots queues and runtime policy every30 seconds and after orderly
shutdown. Files are individually replaced, without a multi-file crash transaction.
Abrupt termination can lose changes since the previous snapshot. Protect the data
directory: it contains the Gateway private key and plaintext pending messages.
Noise encrypts transport through intermediaries; it does not encrypt these files.

The MCP tools are `send-message`, `declare-contactable`, `register-child`,
`list-groups`, `put-group`, `delete-group`, and `set-groups`. The latter three
require a configured `group-admin` capability. Runtime group names begin with
`runtime:`; configuration groups and base memberships remain immutable.
`set-groups` replaces additional runtime memberships. Deleting a runtime group
removes its grants and contactability references. Registering a child never
replaces an existing identity and grants only groups the parent already holds.
