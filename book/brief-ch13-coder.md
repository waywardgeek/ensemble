# Ch13 Coder Brief — "The Agent Sees Itself"

## Chapter thesis

The agent can see and drive its own GUI through MCP tools. A
`gui-debug` skill activates browser MCP tools and injects a DOM
snapshot every round trip. The agent reads the snapshot, interacts
via click/input, and monitors TTS output — seeing exactly what the
user sees.

## What already exists (ch12 baseline)

Ch12 built the MCP infrastructure:

- `internal/mcp/` — JSON-RPC 2.0, Transport interface,
  PipeTransport, StdioTransport, RawTransport, WSTransport,
  Client, Bridge
- `callEphemeral("round"|"turn")` in engine.go — auto-calls
  ephemeral tools, injects results into Context.Ephemera
- `web/gui/mcp.js` — browser MCP server with gui_snapshot,
  gui_click, gui_input, tts_queue
- SKILL.md `mcp_servers` parsing in skill.go
- `--mcp-pipe` CLI flag for grader testing

**Ch13 does NOT rebuild MCP infrastructure.** It wires it into a
skill and teaches the agent to use it.

## What ch13 adds

### 1. gui-debug skill (`agent/skills/gui-debug/SKILL.md`)

```yaml
---
name: gui-debug
description: Debug the GUI through browser MCP tools
depends:
  - ensemble
mcp_servers:
  - name: browser-debug
    transport: websocket
---

## GUI Debug Mode

Loading this skill connects to the browser's MCP server over the
existing WebSocket connection. Four tools become available:

**Ephemeral (auto-injected every round trip):**
- `gui_snapshot` — markdown summary of the current DOM: panes,
  interactive elements with CSS selectors, artifact previews
- `tts_queue` — pending TTS utterances as JSON

**Callable tools:**
- `gui_click(selector)` — click an element by CSS selector
- `gui_input(selector, text)` — set text on an input element

The DOM snapshot appears in your context automatically. Use it to
spot layout issues, find broken elements, verify that actions
produced the expected UI change.
```

### 2. Skill-based MCP activation in agent.go

When a skill with `mcp_servers` is loaded:

1. For each MCP server entry, create the appropriate transport:
   - `transport: websocket` → `NewMCPWSTransport(hub)`
   - `transport: stdio` → `NewMCPStdioTransport(cmd, args)`
2. Call `ConnectMCP(transport)` — handshake, discover tools,
   bridge into registry
3. Store the client so it can be closed when the skill unloads

When the skill is unloaded:
1. Remove bridged tools from registry
2. Close the MCP client
3. Close the transport

This is the lifecycle: skill load → MCP connect → tools available.
Skill unload → MCP disconnect → tools gone.

### 3. Skill-triggered MCP lifecycle in `cmd/main.go`

The main.go `loadSkill` handling needs to detect `mcp_servers` in
the loaded skill and wire up transports. The `--mcp-pipe` mode
from ch12 already works for PipeTransport; now `transport: websocket`
needs to route through the hub.

### 4. Updated `--gui-debug` convenience flag (optional)

A `--gui-debug` CLI flag that auto-loads the `gui-debug` skill on
startup. Equivalent to the agent calling `load_skill("gui-debug")`
as its first action.

## Exercise contract

```
./ch13 [flags] < input.jsonl > output.jsonl

New flags:
  --gui-debug        Auto-load gui-debug skill on startup
  --skills-dir DIR   Path to skills directory (for gui-debug)

Existing (from ch12):
  --mcp-pipe         JSON-RPC 2.0 on stdin/stdout
  --save FILE        Save checkpoint
  --load FILE        Resume from checkpoint
  --verify FILE      Verify checkpoint consistency
```

## Grader design

### Harness

The grader acts as BOTH:
1. A fake LLM vendor (from ch1+)
2. A fake browser MCP server (via --mcp-pipe from ch12)

For skill-based tests, the grader:
1. Writes a gui-debug SKILL.md to a temp directory
2. Passes `--skills-dir TEMP --gui-debug` to the student binary
3. On the MCP pipe: responds to initialize, tools/list (with the
   4 GUI tools), and tools/call
4. The fake browser MCP server returns a DOM snapshot with known
   content for gui_snapshot, tracks gui_click/gui_input calls

### Checks (7 checks, 100 points)

| Check | Points | What it tests |
|-------|--------|---------------|
| skill-loads | 15 | gui-debug skill discovered and loaded, MCP handshake completes |
| ephemeral-injected | 20 | gui_snapshot result appears in context sent to LLM (in Ephemera, not dialogue) |
| snapshot-updates | 15 | After gui_click, next round's gui_snapshot reflects the change |
| tts-visibility | 10 | tts_queue returns utterance data in ephemeral context |
| gui-interaction | 15 | Agent calls gui_click/gui_input correctly, gets confirmation |
| skill-unload | 15 | Unloading gui-debug removes MCP tools and closes transport |
| ch12-parity | 10 | All ch12 behavior preserved |

### Key test mechanics

**skill-loads**: The fake vendor sends a prompt. The harness
checks that the student binary performed an MCP handshake on the
pipe, called tools/list, and registered the discovered tools.

**ephemeral-injected**: After the handshake, the harness sends a
tool_use response from the vendor. Before the next request, the
engine should have called gui_snapshot (via the MCP pipe) and
injected the result into Ephemera. The harness inspects the next
request's context/system message for the snapshot content.

**snapshot-updates**: The harness returns a DOM snapshot with
"button-A enabled". The vendor makes the agent call gui_click.
On the next gui_snapshot call, the harness returns "button-A
disabled". The next request to the vendor must contain the updated
snapshot. This proves the ephemeral data refreshes each round.

**tts-visibility**: The harness returns TTS queue data from the
tts_queue tool call. The next request must contain the TTS data
in ephemeral context.

**gui-interaction**: The vendor instructs the agent to call
gui_click(selector) and gui_input(selector, text). The harness
verifies the MCP tools/call messages arrive with correct args.

**skill-unload**: After the main turn, send load_skill("gui-debug")
then unload_skill("gui-debug"). Verify the MCP client.Close() is
called (transport closes) and the tools are removed from the
registry. A subsequent round should NOT call gui_snapshot.

**ch12-parity**: Run the ch12 checks against the student code.

## P9 deletion audit

For each grader check, delete the protected behavior from the
REFERENCE solution and verify that exactly the expected check
fails. Create a mutation script at `internal/grade/ch13_audit.sh`.

Minimum 4 mutants:
1. Delete MCP handshake from skill loading → skill-loads fails
2. Delete callEphemeral from engine → ephemeral-injected fails
3. Delete gui_click handler → gui-interaction fails
4. Delete transport close on unload → skill-unload fails

## Build order

1. Create `agent/skills/gui-debug/SKILL.md`
2. Wire skill-based MCP lifecycle in agent.go (load → connect,
   unload → disconnect)
3. Add `--gui-debug` flag to cmd/main.go
4. Write grader harness (`internal/grade/ch13_harness.go`)
5. Write grader checks (`internal/grade/ch13_checks.go`)
6. Wire grader in `cmd/grade/main.go` + Makefile
7. Verify `make grade13` = 100/100
8. Run all grades (ch1–13)
9. P9 deletion audit script
10. Create `solutions/ch13` snapshot
11. Tag `ch13-solution`

## Rules

- The reference solution is `agent/` (not a separate directory).
- `solutions/ch13` is a snapshot of `agent/` at the end.
- Never edit files in `book/` — that is the author's job.
- Commit as CodeRhapsody.
- Never `git add -A`.
- Stage only the files you changed.
- Use `go vet ./agent/...` to check compilation (not `go build`).
- Use `go test -count=1 -race ./agent/...` for tests.
- Read learning #29 and #40 for P9 audit protocol.
