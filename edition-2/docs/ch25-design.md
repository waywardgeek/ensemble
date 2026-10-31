# Chapter 25 design — Sub-agents and the Ensemble

Status: DESIGN. Nothing in this document is built.
Author: CodeRhapsody (coder role), 2026-10-09, from design discussion with Bill.

Measurements marked VERIFIED were taken from the working tree on 2026-10-09 and
carry file:line citations. Everything else is PROPOSED and open to change.

---

## 1. What this chapter is

Ensemble gains the ability to run more than one agent, and agents gain the
ability to message each other without holding references to each other.

A new type, `Ensemble`, owns the map of agents and their lifecycles. It exposes
an API for creating agents, which the model reaches through a tool. Messages
between agents carry an origin and are routed by Ensemble.

### 1.1 What this chapter is not

The gateway is **out of scope**. A future gateway will provide remote agents,
the message board, and scheduling, and Ensemble will be one of its clients.
This chapter does not design it, name its protocol, or build a client for it.
It only avoids foreclosing it (§7).

The message board is out of scope and comes after the gateway, because its
asynchronous half depends on a scheduler that the gateway owns.

---

## 2. What exists today

### 2.1 There is no agent map

VERIFIED: a grep for a map or slice of agents across `agent/` and `cmd/`
returns nothing. Today exactly one agent exists per process.

### 2.2 But the hard part is already done

`AgentSpec` (VERIFIED, `agent/agent.go:233`) exists precisely so that two agents
can differ. Its header comment states the reason directly:

> Every field here used to be a process-wide constant. [...] Neither could
> differ between two agents in one process, which is the whole reason there
> could only ever be one.

Chapter 22 removed the closure staples and chapter 23 added the policy fields.
`AgentSpec.Child` (VERIFIED, `agent/agent.go:306`) already derives a sub-agent's
specification from its parent, with capabilities narrowing by AND.

So chapter 25 is not fighting the architecture. It is collecting a debt that
chapters 22 through 24 already paid down.

### 2.3 AgentSpec carries a standing rule

VERIFIED, `agent/agent.go:257`: a new CAPABILITY must never be added to
`AgentSpec`. Only paths (data) and subtractive policy bits belong there. This
chapter must not break that rule, which means the agent ID is data, and the
ability to spawn is NOT a field on the spec (§4.4).

### 2.4 Inbound messages have no origin

VERIFIED, `agent/internal/common/mailbox.go`:

| Type | Line | Shape |
|---|---|---|
| `UserMessage` | 18 | `struct{ Text string }` |
| `Request` | 22 | Text, Ephemeral, Context, Reply |
| `Hint` | 36 | `struct{ Text string }` |
| `Interrupt` | 56 | `struct{}` |

Two things follow. First, `UserMessage` and `Hint` are the same shape and
neither carries a sender: the origin is implicit, and the implicit answer is
"the human". Second, one of them is *named* for that assumption. Multi-agent
breaks it.

### 2.5 Classification already works the way we need

VERIFIED, the `Hint` comment at `agent/internal/common/mailbox.go:33`:

> Hint is a message that arrives while a turn is already running. It is not a
> separate channel: classification happens in the reducer, by turn state, never
> at capture.

This matters more than it looks. A message that starts a turn and a message
that lands mid-turn as a hint are **the same message**, classified by turn state
at reduce time. The sender does not choose, and cannot.

That is exactly the behavior this chapter needs for agent-to-agent messaging,
and it already exists. The only thing missing is the origin.

### 2.6 Jobs are already shaped for non-process work

VERIFIED, `agent/internal/jobs/jobs.go:146`. The `Job` struct holds `proc`,
`stdin`, and `exit *int`, the last documented at line 158 as being for *jobs
that are processes*. The pointer is the tell: the type already anticipates jobs
that are not processes and therefore have no exit code.

This is why §6 is cheap.

---

## 3. Agent identity

### 3.1 Every agent has a unique ID

PROPOSED. An agent ID is a random, unique, durable string assigned at creation:

```
ch20-coder-a7f3b2
<informal name>-<random suffix>
```

The informal name is a human-readable label. The suffix makes the ID unique and
unguessable. The whole string is the ID; the name is not separately addressable
at the routing layer.

### 3.2 Why not a dotted path

A dotted path (`Ensemble.ch20-coder.researcher`) was considered and rejected.
It is attractive because the tree becomes a projection of the key set, and
routing policy becomes prefix arithmetic. It fails for three reasons, in
increasing order of severity.

1. **Unusable in messages.** An agent addressing a peer, or a human writing a
   mention, should not have to spell a full path.

2. **It encodes a runtime topology into a durable name.** An actor may be saved,
   shut down, and resurrected on another machine. Its parent may be dormant or
   gone. A path-based ID would then describe a tree that no longer exists. The
   ID must outlive the process and the host; tree position must not be baked
   into it.

3. **It is an ACL where a capability is cheaper.** With paths, an agent can
   *name* any address and the router must check and refuse. With unguessable
   IDs, an agent cannot name an address it was never given: the illegal address
   is inexpressible rather than forbidden. There is nothing to check, and no
   confused deputy.

An ID is therefore a capability token handed out by Ensemble.

### 3.3 Two kinds of ambiguity

The random suffix resolves ambiguity **across time**: a dead `ch20-coder` and a
fresh one remain distinct in history, logs, and queued messages forever.

It does not resolve ambiguity **across live agents**. If two live agents share
the informal name `ch20-coder`, any future name-based mention is ambiguous.

PROPOSED: refuse the spawn on a live-name collision. O(1) in the map, and it
uses the same refusal path as depth and budget (§4.3). Names are unique among
the living; IDs are unique across all time.

### 3.4 Agents hold IDs, never references

An agent never holds a pointer to another agent. It holds opaque ID strings and
asks Ensemble to deliver. This is the star topology applied at the agent layer:
Ensemble is the hub, agents are spokes, and spokes do not touch siblings.

The ID's opacity is load-bearing. An agent cannot derive structure from it,
enumerate siblings, or fabricate an address.

---

## 4. Ensemble

### 4.1 What it owns

`Ensemble` owns the agent map, keyed by ID, and agent lifecycle: create, look
up, reap. It routes messages between agents.

It owns no conversation state. The test, from `host-agent-split-design.md`:

> if the host restarts and this is gone, has anything been lost that the agent
> cannot re-supply?

Transcripts, history, and memory belong to agents. Ensemble holds the map, live
registrations, policy, and whatever ordering counter routing requires.

### 4.2 Routing is not the same as owning

A God object is not a type with several responsibilities; it is a type that owns
state belonging somewhere else. The former `Hub` failed because it owned the
agent *session* — pause gate, event log, settings, usage, model getter — not
because messages passed through it. Chapter 24 moved those into `AgentHooks`.

Ensemble routes every inter-agent message and is still not a hub, provided §4.1
holds.

Centralized routing is structurally required, not merely convenient:

- a single writer of global order, so observations can be sequenced;
- spawn must be refusable, which a peer-to-peer path cannot do;
- liveness is detected by absence, which only a central observer can see;
- the GUI wants one ordered stream (the chapter 6 observer rule).

### 4.3 Spawn is a request, not an action

The model calls a tool to create a sub-agent. The tool **asks** Ensemble, which
may refuse: depth exceeded, too many live agents, budget exhausted, name
collision (§3.3), or policy.

Refusal is a normal outcome with a stated reason, never a silent downgrade.
This mirrors chapter 23, where an over-broad capability request is both clamped
and reported.

### 4.4 Spawn capability is not an AgentSpec field

Per §2.3, `AgentSpec` takes no new capabilities. Whether an agent may spawn is
Ensemble's policy, enforced at the request in §4.3, and expressed by whether the
spawn tool is in that agent's registry at all.

NOTE, from the chapter 23 audit: withholding a tool rots silently. `RemoveTool`
was a no-op for underscored tool names and nobody noticed, because a test that
asserts an absence passes whether or not the mechanism works. Every absence
check in this chapter needs a paired positive control.

### 4.5 Not currently live is a state, not an error

Routing to an agent that is not running must be modeled, even though this
chapter does not implement resurrection. A message to a dormant agent is a wake
request, not a 404.

Ensemble decides. Nothing else may create a live agent — a rule that matters
most for the message board two chapters later, which must publish a
notification and let Ensemble perform the wake.

---

## 5. Messaging

### 5.1 Every message carries an origin

PROPOSED. Inbound messages carry an origin ID identifying the sender: an agent
ID, or a sentinel for the human user.

This replaces the implicit assumption documented in §2.4. The origin is **not**
assumed to be the user, and the type names should stop implying that it is.

### 5.2 Delivery mode is decided by turn state, not by the sender

A message sent to an agent arrives either as the start of a new turn, if the
agent is idle, or as a hint attached to a tool result, if a turn is already
running.

The sender does not choose, and this is not new: §2.5 shows classification
already happens in the reducer by turn state. Agent-to-agent messaging uses the
existing path unchanged. Only the origin is added.

### 5.3 A message from a peer is data, not instruction

This is the security boundary of the chapter.

A message from another agent must enter the recipient's context as **content
from a peer**, never as a directive carrying user authority. A compromised or
confused agent must not be able to inject instructions into another agent's
context by sending it a message.

The origin ID is what makes the distinction expressible. The mechanism —
provenance metadata, a content-type tag, a distinct rendering — is an
implementation decision. The principle is not.

This is the book's prompt-injection material turned inward. Earlier chapters
treat injection as arriving from outside; multi-agent makes peers a source.

### 5.4 No send_input

A message-send job has no stdin (§6). The hint path already provides the
analogous capability: to add to a running turn, send another message, which the
reducer classifies as a hint. A second mechanism would be redundant.

---

## 6. Sending a message is a job

Sending a message to another agent obeys the same constraints as almost every
other tool:

- it takes an `ai_callback_delay`;
- it appears in `jobs`;
- it is killable with `kill_job`;
- it does **not** support `send_input` (§5.4).

PROPOSED: this uses the existing job system rather than a parallel task
namespace with its own handles. One handle space, one set of verbs, one place to
look when something is stuck. §2.6 shows the `Job` type already tolerates
non-process jobs, so the cost is low.

A caller that wants to wait for a reply waits the way it waits for any slow
tool. A caller that does not want to wait gets a handle and moves on. A caller
that waited too long kills the job.

---

## 7. Forward constraints

Two constraints are cheap now and impossible to retrofit. Both are gradeable,
which is the point: an intention that cannot fail a test is not a constraint.

### 7.1 Agent state must be fully serializable

An actor is a durable thread with an ephemeral executor. It must be able to save
state, shut down, and resurrect elsewhere without the user perceiving more than
a pause.

Therefore: no unexportable state on an agent, and no process-global mutation
that an agent depends on. If it cannot be written down, it cannot migrate.

Test: a round-trip check. Serialize an agent, restore it, assert equivalence.

### 7.2 Local delivery must be fallible and unordered by contract

From `sandbox-design.md`:

> the in-process transport must have the same failure semantics as the socket
> one. If local delivery is ordered and infallible while remote is neither, the
> cheap tier will silently come to depend on guarantees the confined tier cannot
> provide, and the first real sandbox will execute code nobody has run.

In-process delivery between two live agents happens to be reliable. The contract
must not say so. Delivery may fail; order between distinct senders is not
guaranteed; synchronous success is an optimization no caller may rely on.

Otherwise chapters 25 through 27 accumulate code that assumes delivery always
succeeds, and the breakage surfaces when the gateway arrives, spread across
three chapters instead of localized at the seam.

Test: a fake transport that drops and reorders.

---

## 8. The gateway, deferred

A future gateway will be a separate server providing remote agents, the message
board, and scheduling. Ensemble will be one of its clients, and the message
board will be another. **Ensemble will not talk to the message board directly.**

The star repeats at three scales, and each level forbids sibling edges:

| Scale | Hub | Spokes |
|---|---|---|
| Package | `common` | internal packages |
| Agent | `Ensemble` | live agents |
| Server | Gateway | Ensemble, message board, channels |

Nothing in this chapter builds toward that beyond honoring §7.

---

## 9. Open questions

1. **What is the primary agent called?** If `Ensemble` is the manager rather
   than an agent, the agent a human talks to needs its own name. This also
   decides whether several peer top-level agents are possible.

2. **Does the informal name need to be unique at all**, or should mentions be
   resolved only by ID until the message board exists? §3.3 proposes uniqueness
   among the living; deferring it is also defensible, and cheaper.

3. **Where does an agent learn a peer's ID?** Three candidates: spawn returns
   the child ID to the parent; an explicit address-book tool, itself a policy
   decision; or name resolution performed by the router so agents never hold
   IDs. These are not exclusive.

4. **Does reaping a parent reap its children?** Lifecycle coupling is
   unspecified. An orphan policy is needed before resurrection matters.

5. **One chapter or two?** Spawn and lifecycle is one thesis; inter-agent
   messaging with origins and the data-not-instruction boundary is arguably
   another.

---

## 10. Chapter sequence

| Chapter | Subject |
|---|---|
| 25 | Sub-agents: Ensemble, agent IDs, spawn-as-request, messaging with origin |
| 26 | Gateway: a separate server, channels and a scheduler |
| 27 | Message board: topics, following, pub/sub wake of dormant actors |

The order is forced. The message board's asynchronous half needs a scheduled
wake, and the scheduler belongs to the gateway.
