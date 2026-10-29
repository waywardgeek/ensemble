# Optional artifact GUI

## File-based application settings (recommended)

Copy `settings.example.json` into your session directory and edit its paths and
provider fields, then launch with:

```sh
go run ./cmd --config /absolute/path/to/session/settings.json
```

In this mode the file is authoritative: provider, model, endpoint, reasoning,
workspace, skill selection, memory/recall policy, execution limits and preference
settings are not taken from inherited environment variables. Unknown JSON fields,
missing/invalid versions and invalid settings fail clearly; there is no silent
fallback. Relative paths resolve against the file's directory, not the shell's
working directory. Saved paths are normalized to absolute paths.

The main Agent's GUI preferences and provider settings share this file. Settings
shows the file path, the active provider, and editable saved provider/model/billing
and reasoning controls. Provider edits require **Save application settings** and a
process restart. A restart indicator distinguishes saved settings from the active
connection. Ordinary theme/speech/font and memory/budget controls still apply
immediately. Changes are validated and saved through private atomic replacement;
application-only status and session model fields never appear in the file.
Children inherit the running parent's provider configuration, but retain their own
per-Agent GUI preferences; they do not rewrite the main application file.

Credentials are external: `provider.credential_file` names a plain text API key
file for `api` billing, or the existing Codex `auth.json` for `codex` billing.
`chatgpt` uses the separate `auth_directory` credential store and requires an empty
`credential_file`. Never put API keys or OAuth tokens into settings JSON. Advanced
path/endpoint fields are available in the GUI; workspace, skills, recall and auth
store setup remain file-only. Protect credential files with owner-only permissions.
Application file edits made outside the GUI require a restart before further GUI
saves; the running process intentionally holds a validated configuration snapshot.

Only `--port`, `--gui-dir` and `--gui-debug` may override a file-configured launch;
they are reported explicitly. Provider/agent flags alongside `--config` are
rejected instead of silently creating a competing configuration. Without `--config`,
the original environment/flag startup remains available for existing callers and
graders. The public three-argument `gui.New` integration remains unchanged;
file-configured applications use `gui.LoadApplicationConfig` and `gui.NewConfigured`.

## Legacy environment/flag launch

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
explicit unpause. Escape always cancels speech; each speaker button
reads that artifact's full text. Native speech availability depends on the browser;
the page displays speech errors and handles them like completion so the queue
cannot strand the gate.

Ask the Agent for a **copyable text box** containing a particular command, draft
or other snippet. The reply can contain several independent boxes with prose
between them. Each box has a read-only text field and **Copy** for only that
snippet, never the surrounding reply. Prose, thinking and tool results do not
get Copy buttons. Markdown fenced snippets in replies become these boxes.

The generated format is a normal Markdown fenced block, for example:

````text
Here is the requested command:
```text
git --no-pager diff
```
````

Only complete blocks become copyable. Their payload stays literal, including
spaces, blank lines and Markdown; the fences and their structural newlines are
excluded. Use a longer backtick fence to include shorter fences in the payload.
For replies longer than the preview limit, open **Full artifact** to access the
complete boxes; a truncated preview never offers a truncated copy.

**Copied** confirms success. HTTPS/localhost uses the browser's Clipboard API;
ordinary HTTP sessions try the older copy command. If both are blocked, the
snippet field selects its content for Ctrl+C or Command+C. Copying never submits
a prompt or runs a tool. Boxes are reconstructed on replay and agent switching.
`node --test clipboard.test.mjs` checks real Firefox clipboard behavior; install
Playwright Firefox first. Its HTTP cases require a non-loopback IPv4 interface.

Reconnect rebuilds the last hour or last 100 renderable events, whichever is
larger, plus current streaming partials. History is synchronized, finals remove
partials, and display identities survive response-local parser ID reuse. Browser
reconnect replays the selected Agent; process restart first restores saved Agent
history, from which the GUI rebuilds its view.
Every incoming and outgoing WebSocket message is timestamped in `gui.log`.
The other logs are `event.log`, `api.log`, and `debug.log`.

Settings contains theme (dark/light/system), Read aloud, approximate speech speed
(0.5–5, calibrated for Eloquence), and font size (12–24 pixels). Changes apply across
tabs and persist in the application's `gui` section in file mode, or in
`cr/settings.json` for legacy launches. System theme follows changes in the browser's OS preference.
Speech rate applies to the next utterance; disabling speech cancels its queue.
Tool turns have no round-count limit; use Interrupt or shutdown to stop ongoing
work. The settings envelope still accepts and saves the legacy `max_tool_rounds`
field for compatibility, but all values are ignored by execution. Saved conversation
history is restored separately.

`update_settings` carries a sparse `settings` object. The server validates the
patch, saves it, and broadcasts a complete `settings_changed` snapshot; subscribers
receive `current_settings`. Omitted fields stay unchanged, while explicit false
and zero are retained. Errors return `settings_error` without accepting the edit.
Malformed saved JSON fails startup instead of silently discarding preferences.
GUI preference behavior stays in this module; the core has no browser preference
fields or tool-round budget.

The module uses only the core's public `ensemble` package. The application creates
one Ensemble root with Agents. Terminal input addresses the main Agent; each
browser tab can select an Agent and share that Agent's actor. The root
observation seam
supplies progress; request completion remains separate. Socket writes run in one
writer per connection. Slow live receivers may drop progress; reconnect restores
authoritative content. Replay itself is one queued batch and is not truncated by
the progress buffer. Close joins socket writers before the application closes logs.

With the `gui-debug` skill loaded, the Agent can call `gui_snapshot` and
`tts_queue` to inspect the connected browser and pending speech on demand.
Neither observation is collected or attached automatically before model requests.
After a GUI action, request a fresh snapshot to verify its effect.

Run `go test -race ./...` and `go vet ./...` here, separately from the core module.
Browser/native speech checks and targeted mutation evidence are documented in
[the original Chapter 8 record](../../docs/edition-2-notes/ch08.md) in the outer
repository. No Node dependency is required to serve or use this page.

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


Chapter 13 introduces a transcript to the speech channel. Fragments reassemble before
phrase splitting; a single newline becomes a space, blank lines and sentence
endings separate phrases, and finals, tools and errors flush unfinished text.
Markup is filtered, fenced code is named, and identifiers are expanded. Tool
results remain silent automatically. Every utterance records its text and source,
including when Read aloud is disabled. Opening `?speech=transcript` bypasses
native audio without bypassing speech text processing. Native pronunciation and
listening comfort still require a human listener.

The blind virtual user is a separate application over the public core. It owns
an Agent with exactly `hear`, `type`, `submit`, `wait`, and `sleep`; neither DOM
inspection, arbitrary selectors, clicks, shell tools nor skill lifecycle tools
are available to that Agent. The browser peer operates the actual chat form and
reads the actual speech transcript. It starts a separate quiet browser tab, so
close other talking tabs when measuring this channel in isolation.

From this GUI directory, install the optional driver dependency and build:

```sh
npm ci
npx playwright install chromium
go build -o virtual-user ./cmd/virtual-user
```

Start the normal GUI with your chosen target model. Configure the listener's
provider using `LLM_VENDOR`, `LLM_MODEL`, and `LLM_API_KEY`, then run:

```sh
./virtual-user --agent-url http://localhost:8088 --persona blind \
  --task 'Ask the target to explain queueChunk and max_tool_rounds, then report exactly what you heard.'
```

Run from this directory or specify `--peer /absolute/path/to/virtual-user.mjs`.
`--browser /path/to/chromium` selects an existing browser instead of Playwright's
installation. The listener keeps its actual event history in
`virtual-user.jsonl`; it reports a transcript, not a judgment of audible quality.
Its browser subprocess receives no provider credentials. No new agent framework
or alternate conversation loop is involved.

`node --test speech.test.mjs` exercises the real GUI/Agent with external HTTP
and speech fakes, plus the actual narrow client. Set `CHROME_BIN` if Chrome is
not at the macOS default used by this test. The root `tts-harness.sh` implements
the original grader's separate transcript interface; it uses the same application
and a small asynchronous synthesizer fake. Neither fake establishes how speech
sounds. Test setup and scoped limitations are recorded with the chapter evidence.

The header meter shows process-session USD cost, the latest conversation input
(including cache once), latest cache reads, and session cache hit rate. Hover for
session token totals. A dash means at least one contributing model is unpriced.
Engine derives dollars from counts using the effective request model and its
prompt-size tier; restoring history never restores session charges. The bundled
price table was verified October 11, 2026 for Sonnet 5, GPT-5.6 Sol and Gemini
3.8 Flash; the latter's introductory rates need updating January 1, 2027. These
are dated application rates, not a fresh price quote. Other models retain counts
and show unknown cost. Anthropic 5.5 models are excluded from current use.

The cache lens writes `cr/cache-current.txt` and `cr/cache-prior.txt`, in prompt
section order, and reports append/change verdicts and byte/token ratios in
`debug.log`. These are sensitive request-content diagnostics, like `api.log`;
credentials are excluded. Judge/compressor prompts do not replace the main
conversation comparison. A predicted stable byte prefix does not prove provider
cache eligibility or a grant; inspect the measured token reads separately.
