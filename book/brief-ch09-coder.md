# Ch9 Coder Brief — Build Your Dream GUI

## Chapter thesis

The single-pane artifact widget from ch8 becomes a full workbench.
The server gains one new capability: settings management over
WebSocket. Everything else is client-side.

## What the server adds

### Settings management

A `Settings` struct in `internal/common` holds agent configuration
that the GUI can read and change:

```go
type Settings struct {
    Model           string  `json:"model,omitempty"`
    Temperature     float64 `json:"temperature,omitempty"`
    MaxTokens       int     `json:"max_tokens,omitempty"`
    ThinkingBudget  int     `json:"thinking_budget,omitempty"`
    MaxToolRounds   int     `json:"max_tool_rounds,omitempty"`
    SystemPrompt    string  `json:"system_prompt,omitempty"`
    Theme           string  `json:"theme,omitempty"`           // "dark", "light", "system"
    TTSEnabled      bool    `json:"tts_enabled,omitempty"`
    TTSSpeed        float64 `json:"tts_speed,omitempty"`
    FontSize        int     `json:"font_size,omitempty"`
}
```

### Wire protocol additions

Client → Server:
```json
{"type":"update_settings","settings":{"theme":"light","tts_speed":1.5}}
```

Server → Client (broadcast to ALL connected clients including sender):
```json
{"type":"settings_changed","settings":{"theme":"light","tts_speed":1.5}}
```

Server → Client (on subscribe, after event_range):
```json
{"type":"current_settings","settings":{"model":"fake-model","theme":"dark",...}}
```

### Settings persistence

Settings are written to `settings.json` in the working directory on
every `update_settings`. On startup, the hub loads `settings.json`
if it exists. Settings that affect agent behavior (model, temperature,
max_tokens, thinking_budget, max_tool_rounds, system_prompt) are
applied to the engine's Config.

### Settings file location

`settings.json` is in the binary's working directory (same as
the event log). The `--settings` flag is NOT needed — the working
directory is already the right scope.

## What the client adds

### Three-pane layout

```
┌──────┬──────────────────────┬──────────────┐
│ LEFT │     CENTER            │    RIGHT     │
│ MENU │     Chat scroll       │   Actions    │
│      │                       │   scroll     │
│      │                       │              │
│      │                       │  agent tree  │
│      ├───────────────────────┤              │
│      │ [input]        [Send] │              │
└──────┴───────────────────────┴──────────────┘
       ↕                       ↕
    drag bar                drag bar
```

- Left sidebar: hamburger-triggered, collapses to thin strip.
  Two tabs for now: Chats, Artifacts. (More added in later chapters.)
- Center: chat ArtifactScroll (thinking, chat, user messages).
- Right: actions ArtifactScroll (tool calls, tool results) + agent tree.
- Drag bars between panes for resizing.

### Artifact routing

Each message type goes to the appropriate pane:
- `part_delta` with kind=text or kind=thinking → center (chat scroll)
- `part_final` for text parts → center
- `tool_dispatched`, `tool_finished` → right (actions scroll)
- `part_delta` with kind=tool_call → right
- `message` (user messages) → center
- `state_changed`, `turn_ended` → both (state bar)

### Agent tree

Single node for ch9 (sub-agents come later). Shows:
- Agent name or "main"
- State indicator (● idle / ● thinking / ● tools)
- Located below the actions scroll in the right pane

### Settings panel

Accessible from a gear icon in the status bar or left sidebar.
Slides out as an overlay or modal. Groups:
- AI: model, temperature, max tokens, thinking budget, max tool rounds
- Accessibility: TTS on/off, TTS speed, font size
- Appearance: theme (dark/light/system)

Every change sends `update_settings` immediately. No save button.

### Theming

CSS custom properties. Dark default (#0d0d0d). Light theme via
`:root.light` class. System theme via `prefers-color-scheme` media
query.

## Grading table

| check | points | what it tests |
|---|---|---|
| `settings-roundtrip` | 25 | send `update_settings`; receive `settings_changed` with applied values |
| `settings-on-connect` | 20 | subscribe; receive `current_settings` in the handshake |
| `settings-broadcast` | 15 | two clients; one sends `update_settings`; other receives `settings_changed` |
| `settings-persist` | 20 | restart the server; settings survive in settings.json |
| `ch8-parity` | 20 | every ch8 check still passes |
| **total** | **100** | |

## Exercise contract

`ch09/main.go` (or integrated into agent/cmd/main.go): same binary as
ch8, extended with settings. The `--port` flag serves the three-pane
GUI. Settings persist to `settings.json` in the working directory.

```
make grade9
```

## Constraints

- Settings struct in internal/common (star topology holds).
- Hub handles update_settings, persists, broadcasts.
- No new observation types needed — settings changes are transport,
  not agent lifecycle.
- gui.log records settings messages in both directions.
- Agent tree renders from existing observation fields (AgentID on
  observations).
