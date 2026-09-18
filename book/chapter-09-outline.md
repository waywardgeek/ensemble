# Chapter 9 Outline: Build Your Dream GUI

## Through-line stake

The student built one scroll pane and a dark page. Now they discover
that pane is a reusable unit: drop two of them in a layout, add a
settings panel that syncs through the same WebSocket, and the thing
becomes a workbench that survives restarts. The stake is whether the
Artifact abstraction from Chapter 8 actually composes, or whether
adding real workspace features breaks it. It composes.

## Wild facts per section

- §9.2 (three panes): a drag bar is 8 lines of JavaScript with three
  mouse events. The demo that convinces a PM takes longer to explain
  than to code.
- §9.3 (routing): "chat" and "actions" differ by one CSS class and a
  filter function. Everything else is ArtifactScroll running twice.
- §9.4 (settings): the server round-trips settings to every client
  through the WebSocket that already exists, and persists them to one
  JSON file. No database, no API, no /settings endpoint.
- §9.5 (theming): ~15 CSS variables redraw the entire GUI. Dark by
  default because the agent's primary user reads by speech and does
  not need brightness.
- §9.6 (agent tree): a single node that says "idle" or "thinking."
  Honest: until sub-agents exist, the tree is a status display with
  a future-proof data structure.

## Sections

### §9.0 Opening

The Chapter 8 GUI works and is honest about its limitations: one
scroll pane, one input field, no settings, no memory of preferences.
A tool the human can watch is the safety floor. A tool the human can
customize is the one they actually use.

### §9.1 TL;DR

The grader table. What the code does. Settings round-trip, persist,
and broadcast. Routing. Layout. Theming. Everything client-side is
yours; the graded surface is the server-side settings protocol.

Grader table:
| check | points | what it tests |
|---|---|---|
| settings-roundtrip | 25 | update_settings → settings_changed echo with matching values |
| settings-on-connect | 20 | current_settings sent on subscribe |
| settings-broadcast | 20 | second client receives settings_changed |
| settings-persist | 15 | settings survive server restart |
| ch8-parity | 20 | every Chapter 8 check still passes |

### §9.2 Three panes and a drag bar

The layout: left sidebar (agent tree, settings tabs), center (chat
ArtifactScroll), right (actions ArtifactScroll). Two drag bars. CSS
grid or flexbox. Minimum widths enforced. The drag bar is `mousedown`
→ `mousemove` → `mouseup`, 8 lines, no library.

### §9.3 Routing messages to the right pane

`part_delta`, `part_final`, `state_changed`, `turn_ended` → center
(chat). `tool_dispatched`, `tool_finished` → right (actions). Two
`ArtifactScroll` instances, one filter function. The reuse that
Chapter 8's Artifact abstraction promised.

### §9.4 Settings over the wire

`SettingsStore`: a mutex, a struct, a file path. `ApplyRaw` parses
JSON, updates fields, persists, returns the current state.
`update_settings` on the WebSocket: apply, broadcast
`settings_changed` to all clients. `current_settings` on subscribe:
the new client sees the world as it is. No REST endpoint. The
WebSocket carries state changes the same way it carries observations.

### §9.5 Theming

~15 CSS custom properties. `--bg`, `--fg`, `--surface`, `--border`,
`--accent`, plus variants. `[data-theme="dark"]` and
`[data-theme="light"]` on the body. Default is dark because the agent
was built for a user who reads by speech. A `system` option tracks
`prefers-color-scheme`.

### §9.6 The agent tree

One node: the agent's name and state (idle, thinking, tool_use). A
`<ul>` with one `<li>`. The data model is a tree (children array)
because sub-agents will appear in a later chapter. Until then, the
tree is a status indicator with a future-proof shape.

### §9.7 Sidebar tabs

Left sidebar cycles between "Agents" and "Settings." Tabs at the top.
Settings panel: theme selector, TTS toggle, TTS speed slider, font
size selector. Each change sends `update_settings`, receives
`settings_changed`, applies locally. The state lives on the server.
Open a second tab: same settings.

### §9.8 What is not graded and why

Layout, theming, agent tree, sidebar, TTS settings. All client-side.
The grader tests Go because the grader runs Go. The client-side work
is not less important; it is less gradable. The student knows it works
because they can see it.

### §9.9 What Chapter 10 does with this

Sub-agents appear in the tree. The settings panel grows a model
selector, an API key field. The right pane routes sub-agent artifacts
by agent id. The layout survives because the Artifact abstraction
survived.
