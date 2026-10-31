# Chapter 8: Build Your Dream GUI

Get ready for self-wielding. This is the ignition chapter, where you
can finally start using your AI coding agent to write itself. At the
end of this chapter your agent will not compete with Claude Code, and
that is fine. What matters is making the switch as soon as you can be
productive with the new system, because living inside your own agent
is how you discover the bugs and the missing features. This chapter gives you the
freedom to create whatever GUI makes sense to you.

In my case, text-to-speech is critical, and you will find it built
into the reference solution. We are wandering into territory where
the grader cannot help. This is an AI coding agent for you, not for
the LLM, so you will have to drive. Later GUI-driving work can give
control back to the LLM so it can test the system end-to-end, but
for now, you are the one in control.

Chapter 7 gave you one scroll pane on a dark page. It works. You can
watch the agent think, see tool calls arrive, pause to read. As a
safety floor it is complete: a human can watch, and that is the
threshold that matters.

Nobody customizes a safety floor, checks whether their preferences
survived a restart, or drags a divider to put chat on one side and
tool calls on the other, unless the tool they are watching is one
they intend to use every day. This chapter
crosses that line. By the end, the page has three panes, a settings
panel that persists through the same WebSocket, theming that redraws
in fifteen CSS variables, and a sidebar with an agent tree that
currently shows exactly one node. The graded surface is small: four
server-side settings checks and a parity gate. Everything else is
client code you can see working.

## TL;DR

The optional GUI server gains a small settings store: a mutex, a value, a file
path. Browser preferences and their persistence belong in that module. The
Agent owns its effective `max_tool_rounds`; the saved envelope restores that
existing runtime policy through the public Agent interface.
Clients send `update_settings` over the WebSocket. The hub applies the
change, persists it, and broadcasts `settings_changed` to every
connected client. A new client receives `current_settings` on
subscribe.

The client gains a three-pane layout (sidebar, chat, actions), two
drag bars, theming via CSS custom properties, a settings panel, and an
agent tree with one node.

From the repository root, grade the core and its optional GUI separately:

```sh
go run ./cmd/grade-edition2 -ch 8 path/to/core
```

`-gui DIR` selects the GUI module, defaulting to `gui/` under the core.
Its existing `--port` and `--gui-dir` launch, provider environment defaults,
public dependency boundary and `-gui-model` fixture selection carry forward
from Chapter 7; earlier parity flags remain available. No model call is needed to
echo a preference. The restart check launches a new process in the same
disposable working directory, rather than asking the old process for its state.

| check | pts | what it tests |
|---|---|---|
| settings-roundtrip | 25 | `update_settings` → `settings_changed` with matching values |
| settings-on-connect | 20 | `current_settings` sent on subscribe |
| settings-broadcast | 20 | second client receives `settings_changed` |
| settings-persist | 15 | settings survive server restart |
| ch7-parity | 20 | every Chapter 7 check still passes against the appropriate core and GUI clients |

The settings fixture starts with light theme, TTS true, speed 1.5 and font 18;
it then patches dark theme/speed 2, false TTS/budget 2, budget 0, and system
theme. Each full echo and other-client snapshot must preserve the known fields.
Fresh subscription and process restart must return system, false, 2, 18 and 0.
These are test inputs, not mandated defaults or UI range endpoints. Runtime
budget enforcement, invalid-load handling, failed writes and actual browser
controls still require their source/integration evidence.

The original four checks cover the server, not the client-side layout,
theming, tree or native speech. Drive the actual browser and inspect the
effective values. Seeing a saved label is not enough, as the speech-rate
correction below demonstrates.

Four obligations survive the absence of a grader. The first three shipped
broken in the first edition's reference implementation and stayed broken
until its Chapter 13 drove the GUI with a second agent. The fourth survived
until the single-loop retirement audit.

**Validate where the value enters.** A settings field arriving over the
wire is untrusted input. Clamp it in the apply path and again on the
load-from-disk path, rather than at the point of use. The first edition's
audit recorded a temperature of -5 reaching the engine because both paths
trusted the sender. Apply the same lesson to this chapter's actual fields;
that historical example does not add a temperature setting here.

**Decide whether a settings message is a patch or a snapshot.** The two
need different handling of absence. A sparse patch changes only fields that
are present; an explicit zero or false is still a change. `omitempty` on an
ordinary Go boolean cannot express both "absent" and "set false". Decode a
sparse object over a copy of the accepted value, or represent presence
explicitly. A full snapshot must include zero and false. Dropping a clamped
zero from the broadcast leaves the client displaying the number the server
already rejected; dropping false prevents the server from turning TTS off.

**Put control state in an attribute, not a CSS class.** A class styles a
toggle and tells a screen reader nothing, and an automated observer
reads the same nothing.

**Prove that a setting changes behavior.** Persistence and broadcast can
both work while the execution path ignores the value. `max_tool_rounds`
did exactly that: two orchestration loops read a hard-coded limit of 16
while the GUI faithfully saved 200. Chapter 5 now has one owner and one
policy: zero selects the named default of 200; positive N permits N tool
batches per human turn. Read the setting at turn start so an update takes
effect on the next turn, not halfway through the current one. Test both a
small saved limit that stops execution and a large one that permits more
than 16 batches, through the actual entry point and across a restart.

The four visible controls remain theme, TTS enabled, speech speed and font
size. `max_tool_rounds` already belongs to the execution exercise; it travels
in the settings envelope without requiring another GUI control. Reject
negative budgets at the public Agent boundary; the GUI's shared validation
path may clamp them to zero before calling it. Keep those two boundaries
distinct rather than teaching the Engine to repair browser preferences.

## The Why

An agent without preferences is a tool you configure by editing
source. An agent whose preferences vanish on restart is one that
wastes the first thirty seconds of every session re-learning what you
told it. The smallest useful settings system is: accept a change over
the wire, write it to disk, tell every watcher, read it back on
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

No library, and the minimum-width clamp is already in
the `Math.max` call. If you want the right pane to resize too, copy
the same handler with a sign flip on the delta.

The accepted page uses pointer capture so dragging beyond a six-pixel divider
does not lose the gesture. Focused dividers also accept arrow keys and expose
their value as a separator attribute. These are small controls over the same
width change; they do not need another layout system.

## Routing artifacts to the right pane

Chapter 7 built `ArtifactScroll`. This chapter uses it twice.

```javascript
const chatScroll    = new ArtifactScroll(centerEl);
const actionsScroll = new ArtifactScroll(actionsEl);
```

The WebSocket message handler routes by type:

```javascript
function handleMessage(msg) {
  const tool = msg.type === 'tool_dispatched' ||
    msg.type === 'tool_finished' || msg.kind === 'tool_call' ||
    msg.part?.type === 'tool_call';
  (tool ? actionsScroll : chatScroll).receive(msg);
}
```

Text/thinking deltas and finals, and `user_message`, go to chat. Tool-call
argument deltas and finals go to actions, along with `tool_dispatched` and
`tool_finished`. Preserve their artifact and call identities so dispatch
updates the card that already showed the parameters. Status messages update
the workbench's state controls. The routing is a filter, not a fork: both panes use the same
ArtifactScroll API, the same CSS, the same streaming protocol. The
abstraction that Chapter 7 promised holds up under reuse. If it had
not, you would know it here, because the right pane would need its own
rendering logic.

## Settings over the wire

The server needs four things: a place to put settings, a way to
change them, a way to notify, and a way to persist.

```go
// Settings is a complete snapshot; false and zero must stay on the wire.
type Settings struct {
    // Theme follows dark, light or the browser's system preference.
    Theme      string  `json:"theme"`
    // TTSEnabled controls automatic speech.
    TTSEnabled bool    `json:"tts_enabled"`
    // TTSSpeed is the rate consumed by the next native utterance.
    TTSSpeed   float64 `json:"tts_speed"`
    // FontSize is the interface text size in pixels.
    FontSize   int     `json:"font_size"`
    // MaxToolRounds saves the Agent's next-turn policy; zero selects 200.
    MaxToolRounds int  `json:"max_tool_rounds"`
}
```

Keep this value in the optional GUI server, under its existing mutex. The
student first put browser settings and their file handling in the core Agent;
review of the ownership boundary moved them back out. Agent-owned configuration
does not make a theme or a slider a core concern. Only the effective tool budget
belongs to the Agent, and snapshots read it from that owner.

The original `ApplyRaw` sketch ignored decoding and write errors while mutating
the live state. Decode a sparse object onto a copy instead. This excerpt uses
the accepted example's local ranges; those endpoints are implementation choices,
not additional exercise requirements:

```go
func settingsPatch(current Settings, raw json.RawMessage) (Settings, error) {
    raw = bytes.TrimSpace(raw)
    if len(raw) == 0 || raw[0] != '{' {
        return current, fmt.Errorf("settings must be a JSON object")
    }
    next := current
    if err := json.Unmarshal(raw, &next); err != nil {
        return current, fmt.Errorf("invalid settings: %w", err)
    }
    switch next.Theme {
    case "dark", "light", "system":
    default:
        next.Theme = "dark"
    }
    next.TTSSpeed = max(0.5, min(2, next.TTSSpeed))
    next.FontSize = max(12, min(24, next.FontSize))
    next.MaxToolRounds = max(0, next.MaxToolRounds)
    return next, nil
}
```

The same helper validates disk input. A missing file starts with defaults;
malformed saved JSON reports a startup error rather than silently throwing away
preferences. A type error halfway through a patch cannot leave half an update
in the accepted value.

Persist the candidate before publishing it. The accepted server writes a small
replacement file and renames it over the destination; if either operation fails,
the previous accepted value stays active and no success snapshot is broadcast.
This is error handling for one settings file, not a transaction framework or a
claim of power-loss durability. The existing server mutex orders apply,
persistence, publication and new subscriptions.

The WebSocket request carries only the changed fields:

```json
{"type":"update_settings","settings":{"tts_enabled":false,"max_tool_rounds":0}}
```

After a successful save, update the Agent's effective budget through its public
setter, then send `settings_changed` with the full settings
object to every connected client. The sender receives it too, which
confirms the round trip. A new subscriber receives `current_settings`
during the subscribe handshake alongside history and current partials; no new
`event_range` envelope is required. For example, after the other settings have
been set to system theme, speed 2 and font size 18, the snapshot includes:

```json
{"type":"settings_changed","settings":{"theme":"system","tts_enabled":false,"tts_speed":2,"font_size":18,"max_tool_rounds":0}}
```

If decoding or writing fails, report the error instead of broadcasting success.
The accepted page displays the existing `settings_error` reply. Test with an
actual filesystem failure as well as malformed JSON; a happy round trip never
exercises the publication order. The file path is chosen at the application
boundary: the accepted command uses `cr/settings.json` under its working
directory. Restarting that command restores preferences and the next-turn
budget, not a saved conversation.

No REST endpoint. No `/api/settings`. The WebSocket already carries
observations, partial text, state changes, and event-log replays. One
more message type uses the existing ordering. A second transport would add
two sources of truth, and they will disagree the moment a tab
hibernates.

The saved budget needs a test that reaches execution. The accepted entry-point
test builds the actual two-view application, saves a two-batch limit, and
restarts it. While the first provider response is held, a settings update changes
the saved limit to eighteen. The active turn still stops after two batches;
the next terminal turn completes eighteen. Another restart retains eighteen,
and zero again permits more than sixteen by selecting the existing default 200.
Only the external provider is fake; the actor, tools, stdin, WebSocket and file
all remain real.

Reading live configuration on every tool dispatch would look more responsive
and be wrong. Restoring that defect made the active turn run eighteen batches
when it should have stopped at two. A setting can change immediately in the
panel while its documented execution policy starts with the next turn.

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

`document.body.dataset.theme = settings.theme` applies a concrete theme. One
assignment, no class toggling, no re-render. A `system` option first resolves
`prefers-color-scheme`:

```javascript
if (theme === 'system') {
  const dark = matchMedia('(prefers-color-scheme: dark)').matches;
  document.body.dataset.theme = dark ? 'dark' : 'light';
}
```

Also listen for the media query's `change` event while system theme is selected.
A one-time lookup makes the page follow the operating system only until the
operating system changes its mind.

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

The indicator pulses in the accent colour when the agent is working, dims when idle.
`state_changed` messages update `data-state`, and CSS does the rest:

```css
.agent-indicator {
  width: 8px; height: 8px;
  border-radius: 50%;
  background: var(--accent);
  opacity: 0.3;
}
[data-state="in_flight"] .agent-indicator,
[data-state="tools_pending"] .agent-indicator {
  opacity: 1;
  animation: pulse 1.5s ease-in-out infinite;
}
```

Those names must match the Agent's actual state vocabulary. The student copied
the original example's `thinking` and `tool_use` into both the CSS and its first
browser check. The reducer emitted `in_flight` and `tools_pending`; the indicator
did not pulse. The corrected check inspects the active node's computed animation,
not just an attribute the test and CSS agreed to invent. Restoring the old
selectors makes it fail.

The data model is a tree:

```javascript
agents = [{
  id: 'root',
  name: 'Ensemble',
  state: 'idle',
  children: []
}];
```

The `children` array stays empty in this exercise. The rendering code walks
the tree recursively, but there is still only one Agent. The display model
does not create children, define their lifetimes or invent a sub-agent protocol.

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
  if (msg.type === 'settings_changed' || msg.type === 'current_settings') {
    applySettings(msg.settings);
  }
});
```

`applySettings` sets the theme, updates TTS and adjusts font size from the
accepted snapshot. Handle the initial snapshot as well as later changes.
Disabling speech cancels the local queue; a changed speech rate applies to
the next utterance. A successful update reaches the sender and other connected
clients; a fresh subscription or restart restores that saved state. Existing
slow-client progress limits still apply, so this is not a new reliable-message
transport.

### The saved number that speech never used

The settings server accepted and saved speed **1.75**. The label displayed
**1.75** after restart. Native speech ran at approximately **1.8**.

The independent reviewer followed the number past the WebSocket echo. The
range input had `step="0.1"`, so Chrome sanitized its value to 1.8. The label
read the saved object; the speech queue read the slider. Both halves behaved
as written, and they disagreed about what the person had selected.

The correction was one attribute, with a reason beside it:

```html
<!-- Saved rates may fall between tenths; preserve them in the slider. -->
<input id="tts-speed" type="range" min="0.5" max="2" step="any" value="1" />
```

Those bounds are the accepted example's choice. The important relationship is
that the server's accepted values survive the control and reach the consumer.
The focused check saved 1.75, restarted the real application, inspected the
snapshot, label and slider, then submitted browser input and recorded the
native utterance's rate. It repeated with **1.2345**, so a smaller fixed step
could not merely move the defect.

| Saved value | Correct slider and label | Correct native rate | Native rate with step 0.1 restored |
| ---: | ---: | ---: | ---: |
| 1.75 | 1.75 | 1.75 | 1.7999999523162842 |
| 1.2345 | 1.2345 | 1.2345000505447388 | 1.2000000476837158 |

Both correct and defective utterances produced native start/end events.
Completion did not establish that the requested speed was used. The
[baseline](../solutions/evidence/ch09/speed-r1.json) and
[restored defect](../solutions/evidence/ch09/speed-r1-mutation.json)
capture the values before asserting, so a failing check does not erase the
wrong rate it was meant to observe.

This was separate from an earlier test precision mistake. A saved 0.6 reached
the native API as `0.6000000238418579`; an exact equality check timed out.
The focused regression uses a tolerance of 0.00001, small enough for that
representation noise and much smaller than the 1.75-to-1.8 error. An earlier
timeout lacked enough retained location detail to assign it the same cause.
Keep the failed attempts without turning every test failure into a product bug.

Native callbacks and measured rates establish API behavior. They do not claim
that a person listened or found the speech intelligible. That still belongs
to driving the interface described in the opener.

## Taking it for a spin

Build and launch the optional GUI module as in Chapter 7, using a scratch
working directory and its absolute `web/gui` path. Keep the same working
directory when restarting to test `cr/settings.json`; a fresh empty directory
would test defaults instead. Provider credentials remain in the intended child
environment, outside preferences, commands, logs and browser controls.

Open two tabs. Change theme, turn Read aloud off and change font size in one;
the other should show the accepted values. Select system theme and change the
browser's OS preference. Drag each divider, focus it and try the arrow keys.
Start a tool turn: the one tree node should pulse, arguments and output should
appear in Actions, and the reply should remain in Chat. Restart the process
and check the visible controls, then the actual next utterance's rate.

The retained headed-Chrome exercise passed twelve checks covering those
client behaviors, including both pointer dividers and visible restoration of
all four preferences. Its final generic speed-control check uses the native
End key and rate 2; the separate fractional-rate test supplies the coverage
that the earlier tenths-aligned input missed. The
[browser receipt](../solutions/evidence/ch09/browser-r1.json)
records the checks and actual native callbacks.

A bounded live `gpt-4o-mini` turn followed model discovery. The browser saved
light theme, font 18 and TTS off, then restored them after restart. The model
ran a shell command whose actual output was `CH09_LIVE_OK`, with exit code 0;
Actions showed that result and Chat showed the answer. A terminal hint to keep
the reply to one sentence appeared on the same Agent. Clicking the answer's
speaker produced native start/end events. The
[live receipt](../solutions/evidence/ch09/live.json) preserves that
interaction. Unchanged Anthropic and required Gemini 3.8 coverage reuses the
preceding chapter's runs; settings did not justify another paid provider matrix.

The original student grader returned **80/100 FAIL**: all four new settings
checks passed, inherited parity did not. Its roughly 47-minute duration was
estimated from the output file's timestamps while the nested CLI checks ran
against the continuing GUI server. The original legacy solution scored
100/100. Neither score tested the speech-rate defect. Preserve the actual
earlier-build result and the later browser correction separately; the revised
grader uses the appropriate GUI and core entry points. The
[grader review](../docs/edition-2-notes/enhancement-08-code.md) records its
new results and their limits.

## What is not graded and why

The layout, the drag bars, the theming, the agent tree, the sidebar
tabs, the TTS controls. All client-side JavaScript and CSS. The grader
runs Go and connects via WebSocket. It can verify that
`update_settings` echoes correctly, that a second client sees the
change, that a restart preserves the file. It cannot verify that the
center pane is wider than 300 pixels or that the drag bar feels
smooth.

The grader covers its chosen machine checks; actual browser tests can also
measure layout and control behavior. The student supplies the judgment about
comfort, visibility and listening. Both are
necessary, and pretending one replaces the other is how codebases end
up with a green dashboard and a broken product.

## Making the switch

One working Agent, two artifact panes, preferences that survive a restart.
The value of this workbench is that someone can live in it long enough to
discover what the next improvement should be. A tree that can draw one node
does not establish sub-agent execution, and this chapter adds neither model
selection nor API-key controls. The original next-chapter forecast ran ahead
of the available exercise; it is not a requirement to invent those features.

Bill's destination for the new Ensemble is explicit: run GPT-6 Astra at high
reasoning, then use that Ensemble agent to build Edition 3 instead of Codex.
A usable workbench is progress toward that destination. Model access, the
reasoning setting and real coding use still need their own verification before
claiming the crossover. Starting useful work early, as the opener urges, is
how the remaining gaps become visible.
