# Restricted speech listener — Chapter 13

This separate C# application creates an empty library Agent and installs exactly
five custom tools. It does **not** install the coding catalog or a Skills manager.
The restriction is enforced at dispatch, not merely in the model's declarations.

| Tool | Operation |
| --- | --- |
| `say(text)` | Submit a literal target turn; refuse while the preceding turn is pending. |
| `listen(timeout_seconds)` | Return only new speech admissions, with their sources. Wait for speech, explicit turn completion, or the timeout. A timed-out pending turn is not idle. |
| `sleep(seconds)` | Cancellable wait; no browser or transcript access. |
| `set_speech_speed(rate)` | Change the target's acknowledged rate (0.25–4). |
| `save_history(path)` | Write typed submissions and heard entries as JSONL. Relative paths use the session directory. |

The tools opt into atomic completion on the existing Job owner. A long listen or
sleep must not return an inaccessible running handle: this Agent intentionally
has no `wait_for_job`. Ordinary custom tools retain the default timed-observation
behavior. Cancelling a caller still does not kill a Job; joined Agent shutdown
owns that cancellation.

## Build and run

Build `Examples/Gui/Gui.csproj` and `Examples/Listener/Listener.csproj` with the
approved .NET 10 SDK. No browser automation package is required. The listener uses
.NET HTTP/WebSocket clients for a small private Chrome DevTools boundary.

```sh
dotnet Examples/Listener/bin/Debug/net10.0/Listener.dll \
  --dotnet /absolute/path/to/dotnet \
  --gui /absolute/path/to/Gui.dll \
  --browser '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome' \
  --directory /absolute/path/to/new-owned-session \
  --timeout 180 \
  --task 'Ask for a short Markdown explanation, listen, then save heard.jsonl.'
```

The target uses `LLM_VENDOR`, `LLM_MODEL`, `LLM_BASE_URL`, and `LLM_API_KEY`.
The listener requires its own matching `LISTENER_*` variables; it does not
implicitly borrow the target configuration. Supported vendors are `anthropic`, `openai`, and `gemini`. Real use can
spend tokens on **both** Agents; the offline exercise does not authorize spending.
No credentials are put into browser arguments or semantic history.

Without `--task`, supply listener instructions on stdin, one completed turn at a
time. Replies print when ready. EOF ends the entire session, including an active
turn or sleep; SIGINT, SIGTERM and the overall deadline do likewise. Redirecting a single
line and immediately closing the pipe therefore cancels pending work. Use
`--task` for a single command that should run to completion.

Use a new or empty, caller-owned session directory. The target has its own loopback port
(never 8088), settings, diagnostic logs, and Chrome profile. The browser is
headless, extensions are disabled, and its credential store uses an isolated mock
keychain. This launch/termination example is verified on macOS with existing
Chrome; it is not a cross-platform process-control claim. It never attaches to an
existing user browser or Ensemble instance.

## What the listener can observe

The private browser boundary reads only the admitted utterance log and an
explicit `turn_ended` status. It does not return DOM content, arbitrary target
frames, tool results, or a scripting capability to the Agent. The target opens in
`?speech=transcript` mode, so formatting and queue admission are real but native
synthesis is bypassed. “Heard” is a channel-admission label, **not proof of sound**.

JSONL fields are `seq`, `kind`, `source`, and `text`. `typed` records contain the
literal target submission; `heard` records retain text/thinking/error/tool source
labels. Saving collects pending speech without consuming the next listen cursor.
The file is not the listener Agent's reasoning or tool-call history.

The Session is the application composition owner. It owns the root Ensemble,
DevTools socket, HTTP client, target/browser processes, and their output drains.
The Agent owns the Tools and Jobs through the normal core relationships. The
external-assembly callback is the supported custom-tool boundary; no copied
service bag or second execution scheduler is introduced. Destruction first joins
Agent shutdown, then closes browser resources and terminates both owned processes,
including startup-failure paths. Logs remain in the owned session directory.
