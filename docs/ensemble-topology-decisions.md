# Ensemble topology: naming and boundary decisions

Status: decided in conversation with Bill (AUTHOR), 2026-10-06, during ch22.
Not yet implemented. Ch22 is explicitly scoped to NOT do this work.

## The three objects

| Object | Role | Exists today? |
|---|---|---|
| **Ensemble** | Top-level root. Creates agents, owns their lifetimes, routes between them. One per application, many agents. | **No.** Its code exists as straight-line wiring inside `agent/cmd/main.go` (1,154 lines) and has never been named. |
| **GUIServer** | Owns the websocket, the client set, tab syncing, frame reassembly, the usage meter. One per GUI surface. | Yes, as `agent/internal/ws.Hub` (`handler.go:34`, 994 lines). Misplaced: it is a GUI server living in the agent library's internal tree. |
| **Gateway** | Channels (chat, webhooks, voice) plus a scheduler. Remote, with a trust boundary. | No, and deliberately separate. `book/chapters.md:46` reserves it for its own chapter. |

**Ruling (Bill): the Ensemble is a gateway CLIENT, not the gateway itself.**

A Gateway is an edge object: it sits at a boundary, authenticates, and
translates. It owns no agents. The Ensemble is a root object: it sits at the
centre and owns everything. Opposite topologies, and they will coexist in one
running system (the house-security use case puts a monitoring agent and a
main agent behind chat notifications), so they must not share a name.

Bill's own description already separated them -- he described the top-level
object as routing "to gateways for remote agents". Gateways are a thing it
TALKS TO.

## This is the completion of ch6, not a new idea

Two pieces of evidence that the design is the natural extension of what is
already built, rather than a rewrite:

1. `internal/common/observer.go:15` -- "AgentID names which agent an
   observation came from" -- and SEVEN observation variants already carry
   `Agent AgentID`. The observation stream is already per-agent tagged. "One
   GUI, many agents" needs no new data model.

2. `book/chapter-06.md:238` already names the attachment points:
   "nothing inside the framework knows what the observers do. A GUI, a
   gateway, and a sub-agent..." Gateway was always a peer observer.

So: the Ensemble owns agents and exposes N surfaces. GUIServer, Gateway and
CLI are all surfaces. Each attaches observers and injects inbound messages.
Uniform story, already-cut seam.

## Direction of the chain

Agents hold a back-pointer UP to the Ensemble. The Ensemble holds its agents
by ID (it must, to route and to manage lifetimes). Bidirectional, exactly as
Agent/Engine already are. The chain extends by one level:

    Call.Engine.Agent().Ensemble()

Recommendation: every application gets an Ensemble, even single-agent
headless ones. Make it cheap rather than optional. An optional parent means
`if parent != nil` guards, and ch22 deleted one of those that was silently
losing an entire run's token counts in precisely the bare configuration that
tests use.

## Known costs and open questions

1. **The 44-type promotion. This is the gating cost.** `internal/ws`
   references 44 distinct `internal/common` types -- the whole observation
   union plus `Log`, `Seq`, `PauseGate`, `Inbound`, `SettingsSource`,
   `UsageSource`, `Tool`. Go's `internal/` rule means a GUIServer living
   outside `agent/` cannot import any of them. They must be promoted to the
   public API first. The alias layer already exists in `agent/agent.go`
   (`WSHub`, `PauseGate`, `Inbound`, `Log`, `SettingsStore`) but is
   incomplete. This promotion is the real project and is worth doing on its
   own merits.

2. **MCP is multiplexed over the GUI websocket.** `ServeMCP` is a raw TCP
   listener; `BroadcastJSONRPC`/`SetMCPReceiver` route by source tag over the
   same socket as GUI frames. MCP and gateways are Ensemble-level concerns
   (headless, remote); websocket framing is GUIServer-level. That coupling
   has to break. Messiest part of the split -- scope it explicitly.

3. **Settings must become per-agent.** Bill: "Model is a property of the
   agent." Today `settings` is one global store the Hub syncs to all clients,
   and `effectiveModel()` falls back to it. With N agents at different
   models, bands and targets, a single store is wrong. Framework defaults
   plus per-agent overrides, probably.

4. **Switching an observer needs history, not just live events.** Clicking
   agent B must show B's past. Replay comes from the append-only event log
   with a per-client `lastSeq` cursor, and each agent will have its own log.
   So the Ensemble must expose per-agent log access, and "switch" is really
   detach-from-A, attach-to-B, replay-B-from-cursor. Same two-channel hazard
   (log for replay, observations for live) that previously bit the reset
   button.

## Consequence for ch22 (already applied)

Ch22 does NOT restructure the Hub. Giving today's `ws.Hub` a
`common.Agent` back-pointer would wire it INTO the library at the moment it
is due to leave, and the ch22 grader would then enforce a shape about to be
inverted.

Ch22 keeps: the process fix, the ch5 check repair, the `Host` to `Agent`
rename, `common.Engine` + `Call.Engine`, per-model usage, the cost bug,
`agent_status`, and collapsing `cliHost` into `NewAgent`. That last item is
MORE justified under this design, not less: the flat root is the embryonic
Ensemble.

Ch22 fixes only the Hub's usage source -- one line -- so the meter is not
broken by moving the counter onto the Engine.

The arc this sets up: ch22 repairs the chain, the multi-agent chapter spends
it. `spawn_sub_agent` becomes a walk up the parent chain with no new
plumbing, which demonstrates the rule instead of asserting it.
