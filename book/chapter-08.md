# Chapter 8: Build Your Dream GUI

## TL;DR

Make the optional GUI comfortable to use without giving it a second Agent loop.
The server owns presentation preferences; the Agent owns its next-turn tool
budget. A browser sends a patch, the server validates and saves a complete
snapshot, and only then does it publish that snapshot to every connected client.
An active turn keeps the budget it captured when it began.

The deliverable is a working settings panel: theme, narration, speech rate,
font size, speech-related pausing, and tool batches per turn. Preferences survive
browser reload and server restart. A failed save changes neither live behavior
nor the visible acknowledged settings. This is a C#/.NET 10 extension of the
Chapter 7 endpoint, not another runtime or a Go-package translation exercise.

## In plain language

A setting is a small promise. If the page says the font is 24 pixels or the next
turn allows three tool batches, the running application should agree. If a write
fails, the page must not pretend the preference was saved.

There are three different lifetimes:

1. **The browser** displays the last server-acknowledged snapshot. Its controls
   propose changes; local storage is not a second source of truth.
2. **The GUI server** loads and saves presentation preferences and shares them
   with all clients. It reaches the actual Agent to read or change its budget.
3. **The Agent turn** captures its effective budget once. A later settings change
   applies to the next turn, not halfway through this one.

Theme and font changes affect existing transcript content immediately. Speech
rate changes affect the next utterance constructed, not audio already speaking.
Turning off “Pause Agent while speaking” releases the speech-related pause cause
without cancelling speech. Typing is still an independent reason to pause.

These distinctions are small enough to explain and important enough to test.

## 8.1 Keep the existing ownership model

Continue from `solutions/ch07`; advance `main` and preserve earlier
endpoints. Core remains independent of ASP.NET, browser APIs, settings paths,
and user-interface types. The optional `Examples/Gui` assembly references core.

Core `Common/Actor.cs` declares the runtime policy and its public operations:

```csharp
// Agent owns the next-turn policy. The actor alone uses capturedToolLimit.
private int nextTurnToolLimit;
private int capturedToolLimit;
public partial int MaxToolRounds();
public partial void SetMaxToolRounds(int value);
```

The matching implementations are in `Agent/Agent.cs`; scheduling stays in
`Agent/Actor.cs`. The getter returns the configured runtime value, including zero.
The setter accepts zero or a positive integer and rejects a negative integer
without changing anything. Synchronize reads and writes with the Agent's ingress
lock. Do not add a service wrapper or have the GUI inspect actor-private state.

Chapter 5's immutable constructor configuration seeds this live policy; it is not
updated as a competing settings object. Its existing library-only negative limit
continues to mean unlimited. The **new runtime setter and GUI** accept only
nonnegative values. This distinction preserves the older constructor contract
without making negative numbers a new public preference.

At turn start, the actor reads the current policy and resolves zero to the
default 200 batches. Keep that captured value until the turn ends. A queued turn
captures when it actually starts, not when the prompt first arrives. A hint that
continues an active turn does not create a new budget.

A batch is one model response containing tool calls, regardless of how many
calls are in it. After the last admitted batch, let the model consume its results.
Add a short **request-only** instruction to answer without more tools. Extend
`Engine.Render(Context context, string requestInstruction = "")` for this purpose;
do not append the instruction to History or modify opaque/signed replay parts.
If the model still requests another batch, pair its calls with refused results
before any effects. The limit is a scheduling boundary, not merely advice.

## 8.2 Preferences are values, not another owner graph

`Examples/Gui/Common/Settings.cs` declares a value-like `Preferences` record and
partial parse/load/save operations. It contains presentation facts only:

```csharp
public sealed record Preferences(
    string Theme = "system",
    bool TtsEnabled = false,
    double TtsSpeed = 1,
    int FontSize = 18,
    bool PauseWhileSpeaking = true);
```

A record is appropriate because a preference snapshot is a value. It needs no
parent links or logical destruction. The Agent budget deliberately is not a field
in this record: read it from the Agent when making a wire or disk snapshot.
A candidate parsed from JSON may temporarily include a budget value while being
validated and saved; that is not another authoritative live store.

The host's existing composition scope owns its current preferences, settings
path, sockets, and transport synchronization. Do not invent a global settings
registry, browser-owned Agent state, or a second scheduler.

### One validation policy for disk and wire

Use these concrete rules in this endpoint:

| Field | Default | Accepted value |
|---|---|---|
| `theme` | `"system"` | `system`, `light`, or `dark` |
| `tts_enabled` | `false` | JSON Boolean |
| `tts_speed` | `1.0` | finite number, clamped to 0.25–4.0 |
| `font_size` | `18` | 32-bit integer, clamped to 12–32 pixels |
| `pause_while_speaking` | `true` | JSON Boolean |
| `max_tool_rounds` | `0` | nonnegative 32-bit integer; zero means auto/200 |

Wrong types, invalid enum values, negative budgets, out-of-range integer
representations, and unknown fields leave their prior values unchanged. A patch
must be an object. On load, merge valid fields into defaults using the same
validation. Missing, unreadable, or malformed files fall back to defaults without
rewriting or deleting the original file. Do not coerce `"true"`, `1`, and `true`
into interchangeable values.

This deliberately changes Edition 2's recovery policy: malformed saved JSON
used to stop startup with an error. Here preferences can fail without locking
the human out of the GUI. Leaving the file untouched on load gives you a chance
to inspect or recover it; this is not a backup system, and a later successful
settings edit can replace that file. A failed save still leaves the accepted
in-memory values unchanged, as the next section requires.

Preserve zero on disk, in snapshots, and in the input control. Do not replace it
with 200 in the user-facing state. Only the actor's captured effective limit
normalizes zero.

## 8.3 Save before publication

The default path is `$HOME/.ensemble/settings.json`; `--settings PATH` supplies
an explicit path, useful for owned local exercises. These are application
preferences, not conversation events. Do not put credentials in this file.

For a patch:

1. Under the host's short settings/connection lock, merge it with current values
   and the Agent's current policy.
2. Serialize the complete candidate to a unique temporary file in the destination
   directory. Create the parent directory when necessary.
3. Replace the destination with that completed file.
4. Only after replacement succeeds, update the Agent policy and presentation
   preferences, then queue the same full snapshot to every connected client.

The lock serializes concurrent patches and their publication order. Network
sends remain outside it; the existing per-connection bounded queues carry frames.
This small file write is ordinary blocking I/O, not an invitation to build a
persistence service. Same-directory replacement prevents exposing partial JSON;
it does not promise power-loss durability, a backup archive, or recovery from
arbitrary filesystem damage.

If writing or replacing fails, send an `error` to the requester. Do not change
live values, broadcast success, or overwrite the previous good file with a
partial candidate. Clean up the owned temporary file. A later valid request may
try again; do not silently change the persistence path.

## 8.4 WebSocket contract

Keep the existing `/ws` protocol and add:

```json
{"type":"get_settings"}
{"type":"update_settings","settings":{"theme":"dark","tts_speed":1.5}}
```

A newly connected client receives a full snapshot automatically. Explicit
`get_settings` returns the same shape:

```json
{
  "type":"current_settings",
  "settings":{
    "theme":"system",
    "tts_enabled":false,
    "tts_speed":1,
    "font_size":18,
    "pause_while_speaking":true,
    "max_tool_rounds":0
  }
}
```

After successful persistence, all connected clients, including the requester,
receive `settings_updated` with the complete merged `settings` object. Do not
broadcast only the patch. New connections and reconnects obtain current values
from the server, not from another tab. Ignore unknown fields for compatibility.
The older Go grader used the label `settings_changed`; that historical label is
not a requirement to add a second product message.

## 8.5 Build a useful settings panel

The second edition built three panes, drag bars and a sidebar agent tree. This
C# endpoint keeps one transcript and adds settings to the Chapter 7 page.
There are no drag bars or agent-tree controls in this endpoint. Start with this
smaller page, then judge whether it is comfortable enough for your work.

Open `Examples/Gui/wwwroot/index.html`, `style.css` and `app.js` in the preserved
`solutions/ch08` endpoint. They divide the work without a UI framework:

- The header holds connection and Agent state, Pause/Resume/Interrupt buttons,
  and a native `details` element for Settings. Its `summary` opens the controls
  without a custom panel widget.
- The `main` element, `#transcript`, holds conversation cards. The notice below
  it reports connection and settings errors; it has `role="status"` so the
  message is not merely a color change.
- The `#compose` form holds the labelled textarea and Send message/Send hint
  buttons. Keep these separate actions visible even though keyboard shortcuts
  also exist.

The layout is ordinary document flow. The body is centered with a maximum width
of `70rem`; the transcript grows with its cards, and the composer follows it.
The `.controls` rows use wrapping flex layout, so buttons can move onto another
row instead of widening the page. Cards use `overflow-wrap: anywhere`, and
code blocks use `white-space: pre-wrap`, keeping long arguments and code lines
from forcing horizontal page scrolling. At widths below 420 pixels, the page
padding shrinks and buttons can grow to fill a row. Check those choices with
real long output, not just an empty page.

Theming changes shared CSS variables rather than recoloring each card. The root
provides the light palette; `data-theme="dark"` selects the dark one. A
`prefers-color-scheme` media query supplies the system palette only when neither
explicit light nor dark is selected. Font size is another root variable,
`--font-size`, used by the page font and inherited by inputs. Existing content
therefore changes along with the controls.

Follow one edit through the page. Selecting Dark sends the sparse patch
`{"theme":"dark"}` through `patchSettings`. The control can show the proposed
choice while “Saving settings…” is displayed, but the acknowledged settings
remain authoritative. When `settings_updated` arrives, `applySettings` updates
the root theme attribute, font variable and every settings control. Another tab
uses the same path. An error restores the acknowledged values and displays the
failure; it does not pretend the proposed change was saved. This is why the
page needs both a control value and the last accepted settings snapshot.

Put the panel in the existing optional GUI. Provide clearly labeled controls
for every table entry. Numeric input steps are integral for font and batch count;
label zero as auto/default. Keep controls usable on a narrow screen and at the
largest supported font. Group secondary preferences in a collapsible panel so
they do not dominate the transcript.

Use CSS variables for theme colors and font size. Apply them to the root so
existing messages, input, controls, and code blocks update together. Explicit
light/dark settings override the operating system. System mode follows
`prefers-color-scheme`, including changes while the page is open. Syntax colors,
links, focus indicators, and notices must remain readable in both themes.

The browser retains only its last acknowledged settings snapshot. A control
sends a patch and may retain its proposed edit while saving. Actual presentation
and speech policy change only when the server acknowledges a snapshot.
If the connection is unavailable, show a notice rather than pretending to save.
A save error restores the acknowledged control values. Remove any old local-storage
copy of these preferences.

### Speech and pause are related, not identical

Keep Chapter 7's one active utterance and FIFO pending text. When constructing
an utterance, read the **current** acknowledged speech rate. Do not recreate or
cancel the utterance already speaking just because its rate setting changed.
Automatic narration is opt-in; speaker buttons still read selected content.

Both queued and active speech count as busy. The local pause cause is:

```text
typing OR (pause_while_speaking AND speech_busy)
```

Send pause/unpause only when that cause changes. Turning speech pausing off
immediately removes that cause, even with queued audio, but must not override a
remaining typing cause. Turning it on while audio is queued or speaking pauses
immediately. Preserve symmetric onend/onerror cleanup, cancellation identity
checks, cross-tab explicit unpause, and the separate silent visibility wakeup.
Speech settings must not introduce another owner of Agent lifecycle.

## 8.6 Test the behavior, not a screenshot of a slider

Use real components and working external-boundary fakes. The following evidence
answers different questions; none substitutes for the others.

### Core budget test

Hold a real loopback HTTP response after the turn has begun. Change the runtime
policy while it is held. Release it and count actual tool effects: the current
turn must keep its old limit and the next turn must use the new one. Check the
request-only stop-tools instruction, paired refusals, negative setter rejection,
zero preservation, and unchanged History after rendering. Check the instruction
projection for all three provider dialects without contacting those providers.

### Protocol and persistence test

Use two real WebSocket connections to the GUI server. Test full snapshots,
partial merges, invalid fields, clamps, explicit zero, concurrent patches,
new-client state, and actual server restart. Load malformed JSON and a file with
invalid fields. Force a write failure using an owned temporary path; verify the
error, unchanged live snapshots, no success broadcast, and preserved prior file.
Never damage a real user's configuration to create this fixture.

### Browser exercise

Use an isolated browser profile, not another user's connected browser. Operate
the real visible controls and verify both tabs update. Check system theme changes,
existing transcript font, narrow-screen geometry, and reload from server state.
At a fake speech boundary, inspect each constructed utterance's rate, queue two
items, change rate between them, and toggle pause while the queue is busy. Test
error cleanup as well as normal completion. Such a fake establishes wiring and
selection—not audible quality, correct pronunciation, or human accessibility.

### Executable grading

The existing Go grader can launch the compiled C# GUI; its language does not
make HTTP/WebSocket tests Go-only. Use the Chapter 8 `-executable-only` route with
`-gui` pointing to the static asset directory and a GUI executable wrapper as
the positional target. It reuses four settings checks totaling **80 points**:
round trip (25), initial snapshot (20), broadcast (20), and persistence (15).
Inherited Go topology/Chapter 7 scoring is excluded, not silently awarded.
Budget execution, validation/rollback, and browser checks remain separate proof.
Require parsed `passed:true` and process exit zero before reporting a pass.

## 8.7 Use it and report honestly

With authorized provider access, launch the GUI on an owned loopback port and
ask the Agent to perform a small useful programming task. Adjust theme, speech,
font, and the next-turn budget while it works. Inspect the resulting code and
run its tests. Use the existing four private logs to diagnose behavior; do not
publish credentials or raw conversation data.

Without provider authority, exercise the real application against a local fake
and leave the provider-dependent task explicitly incomplete. Fake HTTP,
WebSockets, and speech do not prove remote model compatibility or human hearing.
Do not restart or deploy a shared application merely to demonstrate this chapter.

Preserve a source-only endpoint, explain the ownership and captured-budget
choices, record corrections and evidence limitations in the progress record, and
continue to Chapter 9. The goal is a GUI whose small promises are dependable—not
an elaborate settings framework.
