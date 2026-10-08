# Optional browser client

This module depends on the public `example.com/ensemble` library. The core
library has no GUI dependency. Build the local demo from this directory:

```sh
go build -o /tmp/ensemble-gui ./cmd/ensemble-gui
/tmp/ensemble-gui --port 0 --terminal
```

Use the inherited LLM environment configuration and a fresh `CH02_LOG` path.
The command prints its actual local URL. `--terminal` attaches the existing
public `ensemble/cli.Chat` client to the same Agent. EOF drains that terminal's
requests and detaches; `/quit` shuts down the application. Optional
`--gui-log PATH` creates an exclusive conversation trace with distinct received,
queued and written stages. It never records headers or configuration keys.

`NewServer(parent, agentID, origin, trace)` returns an `http.Handler` and lifetime
owner. Bind a listener on 127.0.0.1 first, then pass its exact HTTP origin. Host
and WebSocket Origin must match including port. `Assets()` exposes the reusable
browser modules. Each public `Connector` owns its socket, watch, independent
pause registration, bounded outgoing queue and workers; its `Server()` parent
provides the path to the application logger and public services.

Browser ES modules are served directly: `connector.js` exports `Connector`,
`artifacts.js` exports `Artifact` and `ArtifactScroll`, `speech.js` exports
`SpeechQueue`, and `page.js` exports the composed `Page`. Components receive their
actual owner. Connector's owner accepts snapshots, observations, replies and
connection/diagnostic notices. ArtifactScroll's owner supplies speech and
logging. SpeechQueue's owner supplies pause reconciliation, diagnostics and
speech-event recording. Page owns those components, input and one speech queue.
`examples/browser-consumer` replaces the page layout and reuses these modules
with two independent Agents in one application.

The browser stages a whole new snapshot before replacing history. Lost watches
or outgoing queues close the generation; reconnect creates a fresh registration
and never retries a prompt. An uncertain submission is visibly marked. Pausing
holds later tool admission; already admitted work continues. Typing and speech
are separate causes, and another tab's cause cannot be released by this one.

Speech starts disabled. Card Speak reads the full retained text. Auto speech
reads new text/thinking and concise tool summaries; history and results remain
silent. Cancel invalidates old callbacks and reconciles the owned causes.
