# Optional artifact GUI

From this directory, run:

```sh
go run ./cmd --vendor anthropic --model claude-sonnet-4-6 --port 8088
```

Choose an available model and supply its credential through the inherited
`LLM_API_KEY` environment variable. Public flags include `--model`, `--vendor`,
`--base-url`, `--workspace`, `--event-log`, `--gui-dir`, and `--no-stream`.
Provider environment defaults retain earlier CLI setup and the original grader's
launch contract; explicit flags win. Avoid putting secrets in shell history.

Open `http://localhost:8088`. Chat displays human messages, thinking and replies;
Actions displays tool arguments and results. Drag either divider to resize a pane
(or focus it and use the arrow keys). The sidebar has Agents and Settings tabs. Enter submits, Shift-Enter adds a line. While busy,
input becomes a hint. The terminal accepts the same human messages plus
`/interrupt`; its EOF detaches that view. Ctrl-C stops
the application. Browser **Interrupt turn** leaves the Agent and other jobs alive.

Typing or queued speech pauses subsequent tools, including with Read aloud off.
**Unpause everyone** explicitly releases the Agent-wide gate from any tab.
Running tools keep running. A tab's local cause does not veto another tab's
explicit unpause. Escape in an empty input cancels speech; each speaker button
reads that artifact's full text. Native speech availability depends on the browser;
the page displays speech errors and handles them like completion so the queue
cannot strand the gate.

Reconnect rebuilds the last hour or last 100 renderable events, whichever is
larger, plus current streaming partials. History is synchronized, finals remove
partials, and display identities survive response-local parser ID reuse. This is
in-memory replay for the current Agent, not restoration after process restart.
Every incoming and outgoing WebSocket message is timestamped in `gui.log`.
The other logs are `event.log`, `api.log`, and `debug.log`.

Settings contains theme (dark/light/system), Read aloud, speech speed (0.5–2),
and font size (12–24 pixels). Changes apply across tabs and persist in
`cr/settings.json`. System theme follows changes in the browser's OS preference.
Speech rate applies to the next utterance; disabling speech cancels its queue.
The same settings envelope accepts `max_tool_rounds` over the WebSocket: zero
selects 200; positive N permits N tool batches. It affects the next turn, including
terminal turns, and is restored on restart. This is not conversation restoration.

`update_settings` carries a sparse `settings` object. The server validates the
patch, saves it, and broadcasts a complete `settings_changed` snapshot; subscribers
receive `current_settings`. Omitted fields stay unchanged, while explicit false
and zero are retained. Errors return `settings_error` without accepting the edit.
Malformed saved JSON fails startup instead of silently discarding preferences.
GUI preference behavior stays in this module; the core owns the effective tool
budget and has no browser preference fields.

The module uses only the core's public `ensemble` package. One application creates
one Agent, and terminal/browser inputs share its actor. The root observation seam
supplies progress; request completion remains separate. Socket writes run in one
writer per connection. Slow live receivers may drop progress; reconnect restores
authoritative content. Replay itself is one queued batch and is not truncated by
the progress buffer. Close joins socket writers before the application closes logs.

Run `go test -race ./...` and `go vet ./...` here, separately from the core module.
Browser/native speech checks and targeted mutation evidence are documented in
`docs/edition-2-notes/ch08.md` in the outer repository. No Node dependency is
required to serve or use this page.

The GUI command accepts the same `EN_SKILLS_DIR`, `EN_PRIMARY_SKILL` and
`EN_CUSTOM_VAR` environment setup as the core CLI. The application passes an
explicit per-Agent skill specification to the public constructor. Browser and
terminal inputs see the same enabled tools and frozen primary prompt.

The browser now serves MCP over `{"type":"jsonrpc","payload":{...}}` frames on
its existing `/ws` connection. `gui_click` and `gui_input` operate visible controls;
`gui_snapshot` and `tts_queue` run automatically before each model request. The
snapshot reports control states and discloses any omissions within its 4KB cap.
Speech observations contain the actual native utterance state and timing.

Without skills, the first connected browser provides these tools. With skills,
include their names in `tools:` and declare a server with `transport: websocket`;
connection waits until a browser is available. A browser already present can
also serve a subsequently loaded skill. Additional tabs still share chat and
settings; the active MCP peer targets one browser, not a broadcast click.

To activate browser tools through the supplied skill, run from this module with
`--gui-debug --skills-dir ../skills`. This selects the `ensemble` primary unless
`EN_PRIMARY_SKILL` overrides it; `--skills-dir` wins over `EN_SKILLS_DIR`.
Open the page before the first prompt so its declared browser peer can connect.
The same skill can be loaded later through `load_skill` when using the primary
without `--gui-debug`. Unload removes executable browser tools immediately.
Programmatic prompt input does not take the typing pause, allowing the Agent to
click Send on its next tool call. Human input keeps the existing pause behavior.
