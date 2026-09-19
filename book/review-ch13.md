# Ch13 Review — "The Agent Sees Itself"

## What was built

Six commits: f1550e4, plus grader, actor lifecycle, P9 audit,
solutions snapshot, and tag.

### New files
- `agent/skills/gui-debug/SKILL.md` — skill activating browser MCP tools
- `internal/grade/ch13_harness.go` (588 lines) — drives skill lifecycle tests
- `internal/grade/ch13_checks.go` (112 lines) — 7 checks, 100 points

### Modified files
- `agent/internal/tools/tools.go` (+60 lines): `RemoveTool`,
  `SetOnSkillMCPConnect`, `SetOnSkillMCPDisconnect`, MCP lifecycle
  in load_skill/unload_skill handlers
- `agent/cmd/main.go` (+100 lines): `--gui-debug`, `--skills-dir` flags,
  skill-based MCP lifecycle wiring (connect/disconnect/reverse handler),
  auto-load on startup
- `agent/internal/llm/actor.go` (+12 lines): skill lifecycle events
- `agent/internal/llm/engine.go` (+8 lines): minor adjustments
- `agent/internal/mcp/stdio.go` (+2 lines): minor fix
- `agent/skills/ensemble/SKILL.md`: added `gui-debug` to loadable-skills

### Grader (7 checks, 100 points)
| Check | Points | Mechanism |
|-------|--------|-----------|
| skill-loads | 15 | Binary starts, loads gui-debug, MCP handshake completes |
| ephemeral-injected | 20 | gui_snapshot in Ephemera sent to LLM |
| snapshot-updates | 15 | After gui_click, next snapshot reflects change |
| tts-visibility | 10 | tts_queue data in ephemeral context |
| gui-interaction | 15 | gui_click/gui_input results in tool results |
| skill-unload | 15 | Source check for disconnect + RemoveTool |
| ch12-parity | 10 | Source check for MCP infrastructure |

## Review

### Must-fix: none

### Enrichment

**E1: skill-unload is a source check, not a behavioral test.**
The skill-unload check (15pts) inspects source code for
`onSkillMCPDisconnect` and `RemoveTool` strings. It does not
actually test that unloading gui-debug removes tools and closes
the transport at runtime. The brief specified a behavioral test
(send load_skill then unload_skill, verify tools disappear). This
is weaker than intended but still catches the most common failure
(forgetting to write the disconnect code at all). Bill should
decide if this is acceptable or if a behavioral test is needed.

**E2: ch12-parity is also a source check.**
Same pattern — looks for string patterns in source files rather
than running ch12 tests. This is consistent with how earlier parity
checks worked (ch11-parity in ch12 was also source-level), so it's
not a regression from the established pattern.

**E3: The grader binary doubles as fake MCP server.**
Clever design: `FakeMCPServer()` runs when the grader receives
`--fake-mcp`. The skill's SKILL.md points at `os.Executable()
--fake-mcp` for transport:stdio. This avoids a separate test
binary but means the grader binary must be on PATH or its own
absolute path must work. The implementation handles this correctly
via `os.Executable()`.

**E4: MCP lifecycle is in main.go, not agent.go.**
The skill-based MCP connect/disconnect is implemented as callbacks
registered in `cmd/main.go`, not in the `agent.go` library API.
This works but means the library user (someone importing `agent`)
must wire MCP lifecycle themselves. For the book this is fine —
main.go IS the reference. But it's worth noting as a future
refactoring opportunity.

### Do NOT add
- Do not add WebSocket transport support to the grader. The stdio
  transport through --fake-mcp is the right testing approach.
- Do not add browser automation. The grader tests the Go
  infrastructure, not the JavaScript.

### Facts verified
- VERIFIED: make grade through grade13 all 100/100 (ch5 120/120)
- VERIFIED: go vet clean
- VERIFIED: gui-debug SKILL.md has correct mcp_servers format
- VERIFIED: P9 deletion audit committed
- VERIFIED: solutions/ch13 snapshot created
- VERIFIED: Tag ch13-solution applied

### Prose direction
Dense, code-forward, ~1000 words. Match ch11/ch12 style.
Key narrative: the skill IS the debug mode — loading it connects
to the browser MCP server, ephemeral tools auto-inject DOM state,
the agent sees what the user sees. Unloading disconnects cleanly.
