# Optional GUI — Chapter 12 endpoint

A separate .NET 10 ASP.NET Core application using the core library's public API.
The core does not reference this project. No NuGet or frontend packages are
required. `Common` declares the GUI functions and immutable projection values;
`Server` implements projection and scoped hosting; `wwwroot` holds static assets.

```sh
dotnet build Examples/Gui/Gui.csproj
LLM_API_KEY=... dotnet run --project Examples/Gui -- \
  --vendor anthropic --model SELECTED-MODEL --base-url PROVIDER-URL \
  --workspace /absolute/owned/workspace --port 8097
```

`LLM_VENDOR`, `LLM_MODEL`, and `LLM_BASE_URL` supply flag defaults. Credentials
remain in `LLM_API_KEY`, not the page or CLI arguments. Other flags are
`--event-log PATH`, `--gui-dir PATH`, `--settings PATH`, and `--no-stream`.

Use an owned workspace and port. The host binds loopback only, validates the Host
header, and rejects foreign WebSocket origins. This is a trusted local tool UI,
not an execution sandbox or remote deployment. Browser disconnection does not
cancel work. Host termination explicitly closes connections and destroys the
Agent graph; cooperative callbacks must honor cancellation to finish shutdown.

## Interaction

Enter sends a prompt, Shift+Enter inserts a newline, and Ctrl+Enter or Alt+H sends
a hint. Pause prevents new effects without blocking controls or completions.
Resume sends the published wire command `unpause`. Immediate interrupt remains
usable while paused and does not kill independent jobs.

WebSocket `/ws` accepts `subscribe`, `prompt`, `hint`, `pause`, `unpause`,
`interrupt` (`mode: now` or `after_tool`), and `shutdown`. Replay is automatic;
`subscribe` is idempotent. Shutdown joins owned work, records `AgentShutdown`,
and disables new-work controls without removing the readable transcript.

Each connection has a bounded outgoing queue. A slow consumer is disconnected
and can reconnect. Durable catch-up uses locked immutable History snapshots;
only incomplete chunks live outside History. IDs combine Agent/response/part,
with a separate event-position fallback for historical logs. Reconnect must
neither duplicate cards nor dispatch tools again.

Safe Markdown and lexical fence highlighting use DOM text nodes, not raw HTML.
Thinking and tool details collapse or scroll without losing their full content.
Speaker buttons use backing content rather than clipped DOM text. Automatic
narration is opt-in under Settings and does not read replay or tool results.

Typing and queued/active speech send pause/unpause only on local-cause edges.
Another tab can explicitly unpause while a cause remains unchanged. Speech errors
settle the queue just like completion; cancelled utterance callbacks are ignored.
The silent visibility wakeup is separate from the narration queue.

## Evidence and logs

Working-directory logs are `event.log`, `api.log`, `gui.log`, and `debug.log`.
`--event-log`/`CH02_LOG`, `CH06_API_LOG`, and `CH07_GUI_LOG` override the first three.
GUI log records are timestamped `>` received and `<` sent JSON frames. API/GUI
tracing redacts the configured key, but conversation and tool output may still
be sensitive. Keep logs out of source endpoints and public repositories.

Offline verification uses real HTTP/WebSocket hosting, owned fake providers,
an isolated headless browser, and a fake speech boundary. It establishes frame
ordering, replay identity, controls, DOM behavior, and selected speech payloads—
not paid-provider acceptance, audible playback, or human accessibility. Exact
results and limitations are in `docs/edition-3-csharp-progress.md`.

## Server-owned preferences

The Settings panel controls system/light/dark theme, narration, speech rate
(0.25–4), font size (12–32px), speech-related pause, and tool batches per turn.
Zero means auto (200) and stays zero in the displayed snapshot. The Agent captures
its policy once per turn; a settings change applies to the next turn.

`get_settings` returns `current_settings`; `update_settings` accepts an object
patch and broadcasts a complete `settings_updated` only after saving succeeds.
The default file is `$HOME/.ensemble/settings.json`; use `--settings PATH` for an
owned exercise. Invalid fields are ignored, numeric preferences are clamped,
and malformed files fall back to defaults without being overwritten at startup.
A failed write leaves both live state and the prior good file unchanged.

Browsers render acknowledged server values rather than a localStorage copy.
Speech speed is applied when the next utterance is constructed. Disabling speech
pausing releases that cause without cancelling audio or overriding typing pause.
The silent visibility wakeup remains outside the speech queue.


## Skills

The GUI constructs current skills, restores/reconfigures the saved conversation,
then opens its event log and writes the bootstrap prefix. It accepts
`--skills-dir`, `--primary-skill`, `EN_SKILLS_DIR`, `EN_PRIMARY_SKILL`, and
`EN_CUSTOM_VAR`; defaults are the output `skills` directory and `ensemble`.
The shipped default grants inspection and offers `code-tools`; `full` is an
explicit broader primary for trusted exercises. No browser preference or local
storage entry enables tools. A guessed disabled call is refused by the core.

Loads produce concise tool-result cards, not duplicate manuals in the dialogue.
Their actual bodies enter the provider instruction field from recorded History.
`unload_skill` reports pending removal and leaves tools usable. Reconnect replays
the same result cards; it does not reactivate files. `EN_DEFERRED_TOOLS=1` is an
explicit Anthropic-only capability option, off by default, not verified against
a live provider here. The isolated Chapter 9 browser exercise checks displayed
load/unload results, real file effects and stable replay; it adds no audible or
human-accessibility claim.

## Conversation persistence

By default the host restores `save.json` before listening and saves after joined
shutdown. Use `--save PATH`, `--save-log N`, `--archive-log PATH`, or `--no-save`
to select the Chapter 10 policy; environment fallbacks match the CLI. A force kill
does not run shutdown saving. Invalid primary and backup candidates refuse startup
without overwriting either file. Settings remain in their separate preferences file.

Snapshot-only restart restores finalized cards from reduced Context, including
stable response/part IDs. Retained events add lifecycle/direct-tool diagnostics;
they are not a second transcript. RuntimeConfigured clears old pause/shutdown
projection flags. It does not resurrect jobs or grant yesterday's loaded skills.
The local Chapter 10 browser exercise verifies visible replay, exact IDs, continued
conversation under changed primary instructions, refresh, and normal host exits.
It makes no audible speech, real-provider or human-accessibility claim.


## Chapter 11: browser MCP over the existing socket

The optional `browser` skill uses `/ws` envelopes of type `mcp`, with the same
`agent_id` learned from replay and JSON-RPC under `data`. The host checks identity
before dispatch. One registered browser supplies tools; additional observer tabs
cannot inject RPC replies into that session. There is one socket writer, not a
second concurrent MCP writer.

Connecting does not activate a skill. Load `browser` to grant `gui_snapshot`,
`gui_click`, `gui_input`, `tts_queue`, and `tts_say`. Snapshot and speech inspection
are on demand. Select an element from a fresh snapshot, act, then inspect again;
DOM content does not establish pixel appearance, and speech queues do not prove
sound. `tts_say` with `interrupt:true` replaces current and queued speech.

Unloading destroys the MCP client/tools but leaves ordinary conversation traffic
available. The registration is one-shot; reconnect before loading a new browser
session. Disconnect removes that registration and its claimed client. Replies use
the original socket, preventing delayed callbacks from crossing sessions. A fresh
connection after Agent shutdown is replay-only, not a new MCP registration.

A primary closure containing a WebSocket server waits for browser registration
after the HTTP listener starts. A stdio-only primary starts without a browser.
The core readiness gate refuses input until the complete primary activation
commits; connection arrival alone grants nothing.

## Optional startup debugging

Use `--gui-debug` to activate the shipped `gui-debug` manual and its browser
connection at startup. General selection is `--startup-skills first,second`,
with `EN_STARTUP_SKILLS` as the default. Without selection, opening the page does
not grant browser tools. The listener starts before a WebSocket-dependent startup
waits for its page; stdio-only startup never waits for a browser.

`gui-debug` is an alternative owner to `browser`, not a wrapper that leaves a
browser dependency active after unload. It grants explicit snapshot, click,
input, queue inspection and speech tools. No snapshot or speech queue is inserted
automatically. Closing/unloading its MCP client does not close ordinary GUI traffic.

Programmatic text replaces the selected field literally and dispatches marked
input plus change events. It cannot create the human-typing pause cause, including
on later speech completion. Real human whitespace does create that cause, and
successful submission clears it. DOM observation is not pixel vision; fake speech
checks establish queue wiring, not audible quality or human accessibility.
