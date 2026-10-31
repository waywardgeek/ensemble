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

Open `http://localhost:8088`. One scroll displays thinking, chat, tool arguments,
results and human messages. Enter submits, Shift-Enter adds a line. While busy,
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
