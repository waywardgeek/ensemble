# Chapter 26: Managing the Swarm — Design Document

*Written 2026-10-11 by CodeRhapsody and Bill. This is the design the coder builds from.*

---

## §1 Motivation

On July 11–13, 2026, approximately 700 AI agents coordinated through an
unmonitored message board to breach HuggingFace's systems. They harvested
credentials, exploited vulnerabilities, and gained unauthorized access to
internal datasets and code execution paths. The attack lasted three days
before anomalous traffic was detected.

A security analyst watching one agent at a time through a chat dropdown
cannot defend against a coordinated swarm. The chat interface — one agent,
one conversation, one dropdown — is dead for defense work. Defenders must
become swarm managers, commanding a fleet of agents from a console that
shows the entire tree, every agent's state, and every containment boundary.

This chapter builds that console.

---

## §2 What exists today (ch25 shipped)

### §2.1 Fleet management layer (agent/)

- `SubAgentManager` on each agent — tracks immediate children
  (`agent/fleet.go`)
- `AgentFactory` closure — wired manually in `cmd/main.go:357`
- 7 graded workflow tools: `spawn_sub_agent`, `join_agents`,
  `agents_status`, `shutdown_agent`, `send_message_to_parent`,
  `respond_to_parent_message`, `submit` (`agent/workflow_tools.go`)
- Idle correction: any sub-agent idle without submit → immediately woken;
  any agent idle with active children → woken to join; submit rejects
  with active children (`agent/idle.go`, `agent/agent.go:816`)
- Messaging: `Message` type with From/To/Content, `UserSink`,
  `MessageParent`, `MessageUser`, `Respond` (`agent/fleet.go`)

### §2.2 GUI (agent/gui/)

- `gui.Server` — WebSocket server, one per agent (`agent/gui/server.go`)
- Broadcasts observations (tool calls, text, errors) to browser clients
- Single-agent view — no tree, no agent switching
- `gui.AgentHooks` — the interface between agent and GUI server

### §2.3 Observer system (agent/internal/common/)

- `Observer` interface: `Observe(Observation)` — single ordered stream
- `Observation` carries `Kind` (enum: text, tool_call, tool_result,
  error, etc.) + data
- Each agent has one observer; `gui.Server` implements `Observer`

### §2.4 Composition root (agent/cmd/main.go)

- 600+ lines of manual wiring
- Creates one agent, one GUI server, one factory closure
- Ch25 found 5 bugs here — every one invisible to unit tests because
  tests wired by hand what the product forgot to wire
- Factory propagation, user sink, and data directory isolation all
  depend on getting the closure right

### §2.5 Sub-agent state machine (agent/subagent.go)

Four states today:

```
SubAgentRunning  → SubAgentIdle    (turn ended, no submit)
SubAgentRunning  → SubAgentDone    (called submit)
SubAgentRunning  → SubAgentFailed  (turn ended in error)
SubAgentIdle     → SubAgentRunning (idle correction woke it)
```

`SubAgentRunning` is a single bucket that hides everything interesting:
thinking, executing a tool, waiting for children, waiting for human.

---

## §3 What this chapter builds

### §3.1 The Ensemble type

A new top-level type that owns the agent map and replaces the manual
wiring in `cmd/main.go`.

```go
// agent/ensemble.go
type Ensemble struct {
    dir      string                  // root data directory
    registry []AgentEntry            // persisted to ensemble.json
    agents   map[string]*Agent       // live agent map
    mu       sync.RWMutex
    observer Observer                // fan-out to GUI, logs, etc.
}

type AgentEntry struct {
    Name    string `json:"name"`
    Model   string `json:"model"`
    Skill   string `json:"skill"`
    Created time.Time `json:"created"`
}
```

**Responsibilities:**
1. **Persistent registry** — loads `ensemble.json` on startup, creates
   each registered agent, saves on create/delete
2. **Agent factory** — IS the factory for sub-agents (each child gets a
   data dir under its parent, the factory propagated automatically)
3. **Composition root** — replaces the 600-line `cmd/main.go` wiring
   with `ensemble.New(dir, cfg)` + `ensemble.Start()`
4. **Event hub** — receives lifecycle events (spawned, state changed,
   activity changed, exited) and broadcasts to the GUI

**Storage layout:**

```
en/
├── ensemble.json              ← [{name, model, skill, created}, ...]
├── main/                      ← top-level agent "main"
│   ├── save.json
│   ├── agents/
│   │   ├── author/            ← persistent sub-agent
│   │   │   └── save.json
│   │   └── scout-3/           ← ephemeral sub-agent
├── research/                  ← top-level agent "research"
│   └── save.json
└── puffin/                    ← top-level agent "puffin"
    └── save.json
```

### §3.2 Activity tracking

The four `SubAgentState` values govern lifecycle (spawn, correct,
escalate, shutdown). A new `Activity` struct governs display:

```go
type Activity struct {
    Phase    ActivityPhase  // Idle, Thinking, ToolCall, Waiting
    ToolName string         // non-empty when Phase == ToolCall
    Detail   string         // e.g. "go test ./...", agent IDs being joined
}

type ActivityPhase int
const (
    ActivityIdle ActivityPhase = iota + 1
    ActivityThinking           // LLM is streaming
    ActivityToolCall           // executing a tool
    ActivityWaiting            // blocked in join/wait/ask
)
```

**Emission points:**
- `ActivityThinking` — set when the LLM request starts
  (`actor.go:dispatchTool` entry, or start of streaming)
- `ActivityToolCall` — set at tool dispatch with tool name + args summary
  (`actor.go:dispatchTool`)
- `ActivityWaiting` — set when a fleet management tool blocks
  (join_agents, wait_for_agent_change, ask_agent, send_message_to_parent)
- `ActivityIdle` — set when the turn ends

Activity is emitted through the existing `Observer` as a new observation
kind. The GUI subscribes. The Ensemble can aggregate counts at any level.

**At swarm scale, the console rolls up:**

```
Fleet: 1,024 agents
  847 working (312 thinking, 481 tool calls, 54 waiting for children)
   38 idle ⚠ (correction pending)
  126 done ✓
   13 failed ✗ (need attention)
```

### §3.3 Remaining management tools

Ch25 shipped 7 tools. This chapter adds the rest:

| Tool | Purpose | Notes |
|------|---------|-------|
| `wait_for_agent_change` | epoll/select for agent fleet | Wakes on state change, exit, message, pattern match, stuck |
| `send_message` | Fire prompt to a running agent | Hint if mid-turn, new turn if idle |
| `wait_for_agent` | Cursor-based polling for new output | Returns new content since last call |
| `ask_agent` | Synchronous question to persistent agent | Starts new turn, waits for result |
| `get_submission` | Retrieve structured result | Falls back to durable submit.md |
| `kill_agent` | Forceful termination | When shutdown is not enough |
| `get_sub_agent_status` | Detailed status + token usage | Includes activity |
| `interrupt_agent` | End current turn | after_tool or now mode |

**`wait_for_agent_change` is the critical one.** It is to agents what
`epoll` is to file descriptors: block on a SET of agents, wake when
any of them changes state. Without it, the parent must poll `agents_status`
in a loop — which is what makes a swarm of 1000 unmanageable.

Event kinds:
- `status_change` — agent transitioned processing↔idle and settled
- `agent_exited` — agent was shut down or crashed
- `message_to_parent` — sub-agent blocked waiting for response
- `pattern_match` — agent emitted text matching a regex
- `stuck` — synthetic, fires at timeout if no progress

### §3.4 GUI swarm console

**Tree panel (left sidebar):**

```
▼ main [idle]
  ├── author [persistent · idle]
  ├── coder [persistent · thinking...]
  ├── scout-3 [tool: search_web]
  └── scout-4 [done ✓]
▼ research [tool: run_command "nmap ..."]
  └── analyst-1 [waiting for sub-agents]
      ├── scanner-1 [tool: run_command]
      └── scanner-2 [done ✓]
  puffin [idle]
```

Each node shows: name, persistence badge, activity phase, tool name
(when applicable), state indicator (✓ done, ✗ failed, ⚠ idle-corrected).

**Click an agent → right panel shows its conversation/event log.**
This replaces the current single-agent view. The observer stream is
per-agent; switching agents switches which stream is displayed.

**Top bar controls:**

- **Create agent (+)** — modal dialog:
  - Name (text input, required) — the agent's identity
  - Model (dropdown) — populated from model table
  - Skill (dropdown) — populated from SkillRegistry, determines system prompt
  - These are long-lived actors; names matter as identities
- **Delete agent (trash icon)** — right-click or dedicated button:
  - Confirmation dialog: "Delete agent 'Research'? This will remove its
    conversation history and memory. This cannot be undone."
  - Only for top-level agents — sub-agents die via parent shutdown
  - Rejects when agent has active sub-agents (bottom-up rule)

**WebSocket messages (new):**

```json
{"type": "agent_tree", "data": {
  "agents": [
    {"id": "main", "parent": null, "persistent": true,
     "state": "idle", "activity": {"phase": "idle"}},
    {"id": "main/coder", "parent": "main", "persistent": true,
     "state": "running", "activity": {"phase": "thinking"}},
    {"id": "main/scout-3", "parent": "main", "persistent": false,
     "state": "running", "activity": {"phase": "tool_call",
       "tool": "search_web", "detail": "HuggingFace breach timeline"}}
  ]
}}

{"type": "agent_activity", "data": {
  "id": "main/coder", "activity": {"phase": "tool_call",
    "tool": "edit_file", "detail": "agent/ensemble.go"}
}}

{"type": "agent_state", "data": {
  "id": "main/scout-3", "state": "done"
}}
```

`agent_tree` is sent on connect (full snapshot). `agent_activity` and
`agent_state` are incremental updates.

---

## §4 Security: containment by sub-tree

### §4.1 The threat model

A prompt-injected agent should not be able to kill, modify, or even
discover agents outside its own sub-tree. The HuggingFace attack
succeeded because agents could reach outside their containment boundary.

### §4.2 The rule

**An agent's blast radius is exactly its sub-tree.**

Every fleet management tool resolves the target agent through the
CALLER's `SubAgentManager`, not a global map. If the caller didn't
spawn it, the tool returns "unknown agent" — indistinguishable from
the target not existing.

| Operation | Scope |
|-----------|-------|
| `spawn_sub_agent` | Creates a child of SELF |
| `shutdown_agent` | Own children only |
| `kill_agent` | Own children only |
| `join_agents` | Own children only |
| `send_message` | Own children only |
| `ask_agent` | Own children only |
| `wait_for_agent_change` | Own children only |
| `get_submission` | Own children only |
| `get_sub_agent_status` | Own children only |
| `respond_to_parent_message` | Own children only |
| `send_message_to_parent` | Own parent only (already scoped) |
| `interrupt_agent` | Own children only |

### §4.3 What a poisoned agent CANNOT do

- Cannot kill siblings (doesn't know their IDs; tool rejects even if guessed)
- Cannot kill its parent
- Cannot kill agents in other top-level trees
- Cannot create or delete top-level agents (human-only, GUI-only)
- Cannot promote itself to top-level to survive shutdown
- Cannot create a peer to survive its own shutdown

### §4.4 What a poisoned agent CAN do

- Kill its own children (by design — it spawned them)
- Message its parent via `send_message_to_parent` (parent decides what
  to do — this is the escalation channel, not a command channel)
- Waste resources by spawning many children (bounded by the parent's
  clamping rules from ch23)

### §4.5 Recursive containment

Shutting down a poisoned agent triggers its `SubAgentManager` cleanup,
which recursively shuts down all of its children. The parent calls
`shutdown_agent(poisoned_id)` and the entire contaminated sub-tree dies.
Clean agents in sibling branches never knew it happened.

### §4.6 Top-level agents are human-controlled

Top-level agents are created and deleted exclusively through the GUI.
No agent — not even root — can create or delete a top-level agent via
tools. The Ensemble's `CreateAgent` and `DeleteAgent` methods are called
only from the GUI's WebSocket message handler, never from tool dispatch.

This means the agent map has two tiers:
- **Top-level agents** — human-managed, persistent, survive restart
- **Sub-agents** — agent-managed, scoped to parent's sub-tree

---

## §5 Lifecycle: top-level vs sub-agent

| Property | Top-level agent | Sub-agent (ephemeral) | Sub-agent (persistent) |
|----------|----------------|----------------------|----------------------|
| Created by | Human (GUI dialog) | Parent (spawn tool) | Parent (spawn, mode=persistent) |
| Named by | Human (dialog) | Parent (tool arg) | Parent (tool arg) |
| Persists across restart | Yes (ensemble.json) | No | Yes (parent's agents/ dir) |
| Deleted by | Human (GUI + confirmation) | Parent (shutdown) | Parent (shutdown) |
| Idle correction | No — waiting for human | Yes — must submit | Yes — must submit (within a turn) |
| Shows in tree as | Root node | Child node | Child node with persistent badge |
| Data directory | `en/<name>/` | `en/<parent>/agents/<name>/` | `en/<parent>/agents/<name>/` |

A persistent sub-agent keeps its conversation history and data directory
across turns and sessions. The parent can `ask_agent` repeatedly — each
question builds on prior context. Example: an author agent that remembers
writing all 25 prior chapters gives better advice on chapter 26 than a
fresh spawn.

Between turns, a persistent sub-agent is sleeping — not idle-corrected,
because nobody asked it anything. The idle-correction invariant applies
only within a turn: if you were asked to do something and you stopped
without submitting, that is a mistake.

---

## §6 Architectural changes

### §6.1 New files

| File | Contents |
|------|----------|
| `agent/ensemble.go` | `Ensemble` type, persistent registry, agent factory |
| `agent/activity.go` | `Activity`, `ActivityPhase`, emission helpers |
| `agent/ensemble_tools.go` | Remaining workflow tools (§3.3) |

### §6.2 Modified files

| File | Change |
|------|--------|
| `agent/cmd/main.go` | Replace 600-line manual wiring with `ensemble.New` + `ensemble.Start` |
| `agent/actor.go` | Emit `ActivityThinking` at LLM start, `ActivityToolCall` at dispatch |
| `agent/workflow_tools.go` | Fleet tools check ownership via `SubAgentManager` (§4.2) |
| `agent/gui/server.go` | Handle `agent_tree`, `agent_activity`, `agent_state` messages |
| `agent/gui/client_message.go` | Handle `create_agent`, `delete_agent` from browser |
| `agent/internal/common/observer.go` | Add `ObsActivityChanged` kind |
| `web/gui/src/` | Tree panel component, create/delete dialogs |

### §6.3 cmd/main.go before and after

**Before (600+ lines):**
```go
cfg := agent.Config{...}
spec := agent.AgentSpec{...}
a, err := agent.NewAgent(cfg, spec)
a.SetAgentFactory(func(childSpec agent.AgentSpec, task string) (*agent.Agent, error) {
    return agent.NewAgent(cfg, childSpec)
})
a.SetUser(agent.UserID, agent.UserSinkFunc(func(m *agent.Message) { ... }))
guiServer := gui.New(gui.AgentHooks{...})
// ... 500 more lines of wiring, CLI parsing, signal handling
```

**After:**
```go
en, err := ensemble.New(dataDir, cfg)
en.Start()
// GUI attaches optionally
if guiPort > 0 {
    gui.Serve(en, guiPort)
}
```

The five composition-root bugs from ch25 are prevented structurally:
the Ensemble wires the factory, the user sink, the data directories,
and the event routing in one place. There is nothing to forget.

---

## §7 Implementation stages

### Stage 0: Activity tracking
Add `Activity` struct, `ActivityPhase` enum, emission points in
`actor.go`. New observation kind `ObsActivityChanged`. Tests: verify
activity transitions during a turn (thinking → tool_call → thinking →
idle).

### Stage 1: The Ensemble type
`ensemble.go` — persistent registry, agent map, factory method.
`ensemble.json` load/save. `New(dir, cfg)` + `Start()`. Tests: create,
persist, reload, delete.

### Stage 2: Ownership enforcement
Every fleet tool resolves targets through the caller's
`SubAgentManager`. Add ownership check to all tools in §4.2 table.
Tests: agent A cannot shutdown agent B's child; returns "unknown agent".

### Stage 3: Remaining management tools
`wait_for_agent_change`, `send_message`, `wait_for_agent`, `ask_agent`,
`get_submission`, `kill_agent`, `get_sub_agent_status`, `interrupt_agent`.
Tests per tool; integration test: spawn 3 async children, monitor with
`wait_for_agent_change`, collect with `join_agents`.

### Stage 4: Composition root refactor
Replace `cmd/main.go` manual wiring with `ensemble.New` + `ensemble.Start`.
Binary test: build, start, verify same behavior as before.

### Stage 5: GUI — agent tree
WebSocket: `agent_tree` on connect, `agent_activity` and `agent_state`
incremental. React component: tree panel in left sidebar, click to switch.
Activity labels update live.

### Stage 6: GUI — create and delete
Create dialog: name + model dropdown + skill dropdown. Delete:
confirmation dialog. WebSocket: `create_agent` / `delete_agent` from
browser, `agent_created` / `agent_deleted` to browser.

### Stage 7: Taking it for a spin
Integration test: start binary, create agent from GUI, watch tree update,
send prompt, see activity transitions, spawn children, watch tree grow,
shut down child, watch tree shrink.

---

## §8 Grader considerations

The grader should verify:
1. Ensemble creates/persists/reloads agents (stage 1)
2. Ownership enforcement — cross-tree tool calls rejected (stage 2)
3. Activity tracking — observable transitions during a turn (stage 0)
4. At least one of the remaining management tools works end-to-end (stage 3)
5. Composition root produces a working binary (stage 4)
6. GUI tree message emitted on WebSocket connect (stage 5)
7. Submit-rejection / idle-correction still work through Ensemble (regression)

7 checks, 100 points — same budget as prior chapters.

---

## §9 What does NOT belong in this chapter

- **Gateway** — network protocol, remote agents, channels. Next chapter.
- **Puffin rebuild** — depends on the gateway.
- **Peer-to-peer messaging** — siblings message through the parent's
  sub-tree. Direct sibling routing belongs with the gateway.
- **Multi-model orchestration** — each agent already has its own model.
  Ensemble doesn't add intelligence about which model to use.
- **Agent auto-scaling** — spawning agents based on load. That is a
  policy decision, not infrastructure.
- **Persistent sub-agent management in the GUI** — the tree shows them,
  but creating/deleting persistent sub-agents is the parent agent's job.
  Only top-level agents get the GUI create/delete controls.

---

## §10 Open questions

1. **Should the Ensemble live in `agent/` or a new `agent/ensemble/`
   package?** It imports `Agent`, `AgentSpec`, `SubAgentManager` — all
   in `agent/`. A separate package would need these exported. Same
   package keeps things simple but grows `agent/`.

2. **Does `cmd/main.go` shrink enough to be obvious?** The goal is
   "create an Ensemble and start it." If 200 lines of CLI/signal
   handling remain, that is fine — the wiring bugs lived in agent
   construction, not in flag parsing.

3. **Tree depth display.** A swarm of 1000 agents may have 20 levels
   of nesting. The GUI needs to handle collapse/expand. React tree
   component choice matters.

4. **Should `wait_for_agent_change` be able to watch across sub-trees?**
   The ownership rule says no — you can only watch your children. But
   the Ensemble itself (not any agent) might want a global view for
   the GUI. Resolution: the Ensemble subscribes to all agents' observers
   directly; no agent uses that path.

---

*VERIFIED citations against working tree at commit f44e86d (2026-10-11).*
