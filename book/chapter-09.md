# Chapter 9: Build Your Dream GUI

Get ready for self-wielding. This is the ignition chapter, where you
can finally start using your AI coding agent to write itself. At the
end of this chapter your agent will not compete with Claude Code, and
that is fine. What matters is making the switch as soon as you can be
productive with the new system, because living inside your own agent
is how you discover the bugs and the missing features. There is
nothing like building software with a tool you built to help you
figure out what you want that tool to be. This chapter gives you the
freedom to create whatever GUI makes sense to you.

In my case, text-to-speech is critical, and you will find it built
into the reference solution. We are wandering into territory where
the grader cannot help. This is an AI coding agent for you, not for
the LLM, so you will have to drive. In the next chapter we will try
to give control back to the LLM so it can test the entire system
end-to-end, but for now, you are the one in control.

Chapter 8 gave you one scroll pane on a dark page. It works. You can
watch the agent think, see tool calls arrive, pause to read. As a
safety floor it is complete: a human can watch, and that is the
threshold that matters.

Nobody customizes a safety floor. Nobody opens a second tab to check
whether their preferences survived a restart. Nobody drags a divider
to put chat on one side and tool calls on the other, unless the tool
they are watching is one they intend to use every day. This chapter
crosses that line. By the end, the page has three panes, a settings
panel that persists through the same WebSocket, theming that redraws
in fifteen CSS variables, and a sidebar with an agent tree that
currently shows exactly one node. The graded surface is small: four
server-side settings checks and a parity gate. Everything else is
client code you can see working.

## TL;DR

The server gains a `SettingsStore`: a mutex, a struct, a file path.
Clients send `update_settings` over the WebSocket. The hub applies the
change, persists it, and broadcasts `settings_changed` to every
connected client. A new client receives `current_settings` on
subscribe.

The client gains a three-pane layout (sidebar, chat, actions), two
drag bars, theming via CSS custom properties, a settings panel, and an
agent tree with one node.

| check | pts | what it tests |
|---|---|---|
| settings-roundtrip | 25 | `update_settings` → `settings_changed` with matching values |
| settings-on-connect | 20 | `current_settings` sent on subscribe |
| settings-broadcast | 20 | second client receives `settings_changed` |
| settings-persist | 15 | settings survive server restart |
| ch8-parity | 20 | every Chapter 8 check still passes |

The client-side layout, theming, agent tree, and TTS settings are not
graded. The grader runs Go. You know the client works because you can
see it.

## The Why

An agent without preferences is a tool you configure by editing
source. An agent whose preferences vanish on restart is one that
wastes the first thirty seconds of every session re-learning what you
told it. The smallest useful settings system is: accept a change over
the wire, tell every watcher, write it to disk, read it back on
startup. That is four operations, and the grader tests each one.

The layout and theming are the other half. They are not graded because
"does it look right" is a human judgment, and you are the human. The
grader trusts you to open a browser.

## Three panes and a drag bar

The layout is three columns: a left sidebar, a center pane for chat,
and a right pane for tool output. Two vertical drag bars separate
them.

```css
.layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}
.sidebar    { width: 260px; min-width: 180px; }
.drag-bar   { width: 6px; cursor: col-resize; background: var(--border); }
.center     { flex: 1; min-width: 300px; }
.actions    { width: 380px; min-width: 200px; }
```

A drag bar is three events:

```javascript
bar.addEventListener('mousedown', e => {
  const startX = e.clientX;
  const startW = left.offsetWidth;
  const move = e2 => {
    left.style.width = Math.max(180, startW + e2.clientX - startX) + 'px';
  };
  const up = () => {
    document.removeEventListener('mousemove', move);
    document.removeEventListener('mouseup', up);
  };
  document.addEventListener('mousemove', move);
  document.addEventListener('mouseup', up);
});
```

Eight lines, no library, and the minimum-width clamp is already in
the `Math.max` call. If you want the right pane to resize too, copy
the same handler with a sign flip on the delta.

## Routing artifacts to the right pane

Chapter 8 built `ArtifactScroll`. This chapter uses it twice.

```javascript
const chatScroll    = new ArtifactScroll(centerEl);
const actionsScroll = new ArtifactScroll(actionsEl);
```

The WebSocket message handler routes by type:

```javascript
function handleMessage(msg) {
  if (msg.type === 'tool_dispatched' || msg.type === 'tool_finished') {
    actionsScroll.handle(msg);
  } else {
    chatScroll.handle(msg);
  }
}
```

`part_delta`, `part_final`, `state_changed`, `turn_ended`, and
`message` go to chat. `tool_dispatched` and `tool_finished` go to
actions. The routing is a filter, not a fork: both panes use the same
ArtifactScroll API, the same CSS, the same streaming protocol. The
abstraction that Chapter 8 promised holds up under reuse. If it had
not, you would know it here, because the right pane would need its own
rendering logic.

## Settings over the wire

The server needs four things: a place to put settings, a way to
change them, a way to notify, and a way to persist.

```go
type Settings struct {
    Theme      string  `json:"theme"`
    TTSEnabled bool    `json:"tts_enabled"`
    TTSSpeed   float64 `json:"tts_speed"`
    FontSize   int     `json:"font_size"`
}

type SettingsStore struct {
    mu   sync.Mutex
    data Settings
    path string
}
```

`ApplyRaw` takes a `json.RawMessage`, unmarshals it on top of the
current state, persists, and returns the result:

```go
func (s *SettingsStore) ApplyRaw(raw json.RawMessage) Settings {
    s.mu.Lock()
    defer s.mu.Unlock()
    json.Unmarshal(raw, &s.data)
    s.persist()
    return s.data
}
```

`persist` writes JSON to `s.path`. `NewSettingsStore` reads from
`s.path` if the file exists. The full lifecycle: startup reads, update
writes, restart reads the write.

The WebSocket handler adds one case:

```go
case "update_settings":
    updated := h.settings.ApplyRaw(msg.Settings)
    h.broadcastSettings(updated)
```

`broadcastSettings` sends `settings_changed` with the full settings
object to every connected client. The sender receives it too, which
confirms the round trip. A new subscriber receives `current_settings`
during the subscribe handshake, after `event_range` and any in-flight
partials.

No REST endpoint. No `/api/settings`. The WebSocket already carries
observations, partial text, state changes, and event-log replays. One
more message type costs nothing. A second transport costs everything:
two sources of truth, and they will disagree the moment a tab
hibernates.

## Theming

Fifteen CSS custom properties redraw the entire page:

```css
:root, [data-theme="dark"] {
  --bg: #0d0d0d;
  --fg: #e0e0e0;
  --surface: #1a1a1a;
  --border: #333;
  --accent: #5b9bd5;
  --accent-hover: #7ab3e8;
  --thinking-bg: #1a1a2e;
  --code-bg: #1e1e1e;
  --input-bg: #1a1a1a;
  --input-border: #444;
  --sidebar-bg: #111;
  --drag-bar: #333;
  --tool-bg: #1a1a1a;
  --error-bg: #2d1a1a;
  --scrollbar-thumb: #444;
}
```

Dark is the default because this agent was built for a user who reads
by speech and does not need brightness. Every element uses `var(--bg)`
instead of `#0d0d0d`, so adding a light theme is writing fifteen new
values:

```css
[data-theme="light"] {
  --bg: #ffffff;
  --fg: #1a1a1a;
  --surface: #f5f5f5;
  /* ... */
}
```

`document.body.dataset.theme = settings.theme` applies it. One
assignment, no class toggling, no re-render. A `system` option
watches `prefers-color-scheme`:

```javascript
if (theme === 'system') {
  const dark = matchMedia('(prefers-color-scheme: dark)').matches;
  document.body.dataset.theme = dark ? 'dark' : 'light';
}
```

## The agent tree

The sidebar shows one node: the agent's name and its state.

```html
<ul class="agent-tree">
  <li class="agent-node" data-state="idle">
    <span class="agent-indicator"></span>
    <span class="agent-name">Ensemble</span>
  </li>
</ul>
```

The indicator pulses green when the agent is working, dims when idle.
`state_changed` messages update `data-state`, and CSS does the rest:

```css
.agent-indicator {
  width: 8px; height: 8px;
  border-radius: 50%;
  background: var(--accent);
  opacity: 0.3;
}
[data-state="thinking"] .agent-indicator,
[data-state="tool_use"] .agent-indicator {
  opacity: 1;
  animation: pulse 1.5s ease-in-out infinite;
}
```

The data model is a tree:

```javascript
agents = [{
  id: 'root',
  name: 'Ensemble',
  state: 'idle',
  children: []
}];
```

Until sub-agents exist, the `children` array is empty. The rendering
code walks the tree recursively. When sub-agents appear in a later
chapter, they are pushed into `children` and the same render function
draws them indented. That is the point of the tree shape: it costs
nothing now and saves a rewrite later.

## Sidebar tabs

The left sidebar cycles between two tabs: **Agents** and
**Settings**. The settings panel contains:

- Theme selector (dark, light, system)
- TTS toggle
- TTS speed slider
- Font size selector

Each control sends `update_settings` with the changed field. The
server echoes `settings_changed`. The handler applies it locally:

```javascript
ws.addEventListener('message', e => {
  const msg = JSON.parse(e.data);
  if (msg.type === 'settings_changed') {
    applySettings(msg.settings);
  }
});
```

`applySettings` sets the theme, updates TTS, adjusts font size, and
every connected client converges on the same state within one WebSocket
round trip. The persistence ensures a restart converges too.

## What is not graded and why

The layout, the drag bars, the theming, the agent tree, the sidebar
tabs, the TTS controls. All client-side JavaScript and CSS. The grader
runs Go and connects via WebSocket. It can verify that
`update_settings` echoes correctly, that a second client sees the
change, that a restart preserves the file. It cannot verify that the
center pane is wider than 300 pixels or that the drag bar feels
smooth.

This is not a gap. It is a division of labor. The grader tests what a
machine can test. The student tests what a human can see. Both are
necessary, and pretending one replaces the other is how codebases end
up with a green dashboard and a broken product.

## What Chapter 10 does with this

The agent tree gains children. Sub-agents appear as nested nodes, each
with their own state indicator, each routing their artifacts to the
right pane. The settings panel grows a model selector and an API key
field. The three-pane layout survives because the Artifact abstraction
survived. Every piece this chapter added was designed to hold weight
it does not carry yet.
