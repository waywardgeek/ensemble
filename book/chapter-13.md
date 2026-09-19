# Chapter 13: The Agent Sees Itself

A coding agent that cannot see its own GUI is debugging blind. Every tool so far has operated on files, processes, and network responses. The GUI is a black box the user stares at while the agent types into it. This chapter closes that gap. A single skill connects the agent to its running browser interface through the MCP infrastructure from Chapter 12, and the agent begins seeing what the user sees: the DOM, the buttons, the text being spoken aloud.

## TL;DR

A `gui-debug` skill activates browser MCP tools via the skill system from Chapter 10 and the MCP infrastructure from Chapter 12. Loading the skill triggers an MCP handshake, discovers four tools, and registers them. Two are ephemeral (auto-injected every round), two are callable.

### The skill

```yaml
---
name: gui-debug
description: Debug the GUI through browser MCP tools
depends:
  - ensemble
mcp_servers:
  - name: browser-debug
    transport: stdio
    command: ./grader
    args: ["--fake-mcp"]
---
```

The `mcp_servers` field declares a server to connect when the skill loads. The `transport: stdio` entry spawns the command as a subprocess and speaks JSON-RPC 2.0 over its stdin/stdout. For grading, the command points at the grader binary itself running in fake MCP server mode. In production, the transport would be `websocket`, tunneling through the hub to the browser.

### Skill-based MCP lifecycle

Loading a skill with `mcp_servers`:

1. For each server entry, create the transport (`StdioTransport` for `stdio`, `WSTransport` for `websocket`).
2. Create an MCP `Client`, call `Initialize`, then `ListTools`.
3. Bridge discovered tools into the agent's registry via `mcp.Bridge`.
4. Store the client handle for cleanup.

Unloading the skill:

1. Remove bridged tools from the registry via `RemoveTool`.
2. Close the MCP client.
3. Close the transport (kills the subprocess for stdio).

Two callbacks on the tool registry wire this lifecycle:

```go
type Reg struct {
    // ...
    onSkillMCPConnect    func(skill string, servers []common.MCPServerConfig) ([]string, error)
    onSkillMCPDisconnect func(skill string)
}
```

The `load_skill` handler calls `onSkillMCPConnect` after loading the skill's tools. The `unload_skill` handler calls `onSkillMCPDisconnect` before removing them. The callbacks live in `cmd/main.go` where the MCP client, transport factories, and registry are all in scope.

### RemoveTool

The registry gains a `RemoveTool(name)` method. Chapter 12 added tools dynamically via `RegisterTool`; this chapter removes them dynamically when a skill unloads. The tool is deleted from the map and its declaration is removed from the cached list. The `onToolsChanged` callback fires so the engine picks up the new tool set.

### The four browser tools

These are the same four tools from Chapter 12's `mcp.js`, now activated through the skill system:

| Tool | Ephemeral | What it returns |
|------|-----------|-----------------|
| gui_snapshot | round | Markdown DOM: panes, interactive elements with selectors, artifacts |
| tts_queue | round | JSON array of pending TTS utterances |
| gui_click | no | Confirmation: "clicked #button-A" |
| gui_input | no | Confirmation: "set text on #search-box" |

The ephemeral tools are called automatically by the engine before each round. Their output appears in `Context.Ephemera`. The LLM sees the DOM snapshot and TTS queue as context data, refreshed every round, without having to ask for it.

The callable tools (`gui_click`, `gui_input`) appear in the LLM's tool declarations. The LLM decides when to click or type based on what the snapshot shows.

### --gui-debug flag

A convenience flag that auto-loads the `gui-debug` skill on startup:

```
./ensemble --gui-debug --skills-dir ./skills
```

Equivalent to the agent calling `load_skill("gui-debug")` as its first action. The `--skills-dir` flag overrides the default skills directory, which is useful for grading with temp directories containing test fixtures.

### Exercise

```
./ensemble --gui-debug --skills-dir SKILLS_DIR
```

The grader binary doubles as a fake MCP server (`--fake-mcp` flag). It writes a temporary `gui-debug` SKILL.md whose `command` points at itself. When the student binary loads the skill, it spawns the grader as a subprocess, establishing a JSON-RPC channel.

The fake MCP server maintains state: `gui_snapshot` returns "button-A: enabled" initially, then "button-A: disabled" after a `gui_click` call. This simulates the DOM changing in response to interaction, and the grader verifies that the next round's ephemeral injection reflects the updated state.

Seven checks:

| Check | Points |
|-------|--------|
| skill-loads | 15 |
| ephemeral-injected | 20 |
| snapshot-updates | 15 |
| tts-visibility | 10 |
| gui-interaction | 15 |
| skill-unload | 15 |
| ch12-parity | 10 |

```
make grade13
```
