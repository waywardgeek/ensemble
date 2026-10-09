# Chapter 9 student implementation

A Go Agent library with human and JSON-lines CLI clients, plus a separate,
optional browser module. The browser shares the Agent with the CLI, supports
atomic reconnect snapshots, owned typing/speaking pauses, safe cards and
opt-in speech. This is the canonical second-edition source in the outer
repository; frozen chapter exports are separate.

Chapter 9 adds Agent-owned skills, per-Agent grants, retained manuals, typed
public controls and browser skill cards. Initial implementation and live evidence
are frozen at `786ff23`; comparative improvements are at `06c6787`, independently
accepted at `19d2fdf`. Current grant/state reads avoid historical inspection,
and the public two-Agent example preserves each partial result and terminal
outcome before reporting overall failure. Local revision tests complement the
original source-bound live runs.
See the [Chapter 9 validation record](../../../book/edition-2/chapter-09-validation.md)
for final acceptance, export and tag status.

Build the CLI with `go build -o ensemble ./cmd`. Select `LLM_VENDOR`, a discovered
`LLM_MODEL`, and an API key through the process environment. Use a fresh
`CH02_LOG` path: existing logs are never overwritten or resumed. Run
`ensemble chat` and type an ordinary single-line request at `You> `. With no
arguments, actual terminal stdin and stdout select chat; redirected input selects
the machine protocol. Explicit `ensemble protocol` takes one JSON directive per
line, including legacy `user`, `ephemeral` and `redact` records and Chapter 5's
correlated prompt, hint and interrupt records.

Human commands are `/help`, `/usage`, `/history`, `/skills`, `/ephemeral TEXT`,
`/redact FROM TO REASON`, `/hint TEXT`, `/interrupt`, and `/quit`. History shows `tool_returned` sequences
and call IDs so a person can select a real redaction target. EOF also exits with
four usage counts. Prefix a literal leading slash with another slash. Human
input is UTF-8, at most 1 MiB per line excluding LF or CRLF.

Without skills, the CLI enables read, list, search, write, edit, `run_command`, `wait_for_job`,
`send_input`, `kill_job`, and `tool_limits`. Managed commands use a PTY and can
continue between reports. Use a scratch working directory: this workspace is
not a sandbox.
The same public submission operation records tool results and returns the final
answer. A turn captures its Agent's applied model-request limit when activated;
zero selects the default sixteen, and positive values select 1–256. Ordinary tool errors
can be corrected by the model; provider or persistence failures end the session.

`ensemble dump` reads `CH02_LOG`; `ensemble render LOG` renders the specified log
for the selected vendor/model. Neither operation needs credentials or contacts a
provider. `LLM_RESOLVED_MODEL` explicitly identifies an alias's exact target when
replaying bound material; returned model identity is not an automatic alias map.

Applications import `example.com/ensemble`, construct one Ensemble, then create
Agents with separate log paths. Prompt/directive requests name the Agent and
return their own result. Observers receive ordered owned event snapshots. The
public append path ingests supplied tool results. Agents explicitly select their
built-in tools; an omitted set exposes none. Each Agent owns its absolute
workspace and Registry; declaration and dispatch use the same visible set.
Loading a log is read-only. A persistence failure permanently faults that Agent.

To use skills, set both `LLM_SKILLS_DIR` to the absolute `skills` directory and
`LLM_PRIMARY_SKILL` to `base` (narrow) or `ensemble` (full). Leave `LLM_SYSTEM`
absent or empty so the primary supplies the base instruction. Both skill settings
absent preserves no-skills mode; incomplete or blank selections refuse startup.
The GUI uses the same settings. `/skills` inspects current authority without a
model call. Ask the model to load or unload an offered skill through its tools;
retained manual text does not grant permission after unload.

Build `examples/skills-consumer` in its own module for a headless two-Agent
demonstration. Its default mode uses typed controls; `--ask` also performs two
bounded model turns. Prepare separate `alpha/notes.txt` and `beta/notes.txt` in a
fresh workspace. Each invocation creates exclusive logs, so use a new workspace
to try the other mode. The example reports both terminal outcomes even if one
fails. `evidence/ch09/` preserves actual CLI/browser/public runs on all three
providers, including failed attempts and the limits of keyboard/audio evidence.

The independently built program in `examples/tools-consumer` demonstrates
workspace isolation, per-Agent tools and observations. `examples/consumer`
retains the earlier controlled-result ingestion example and its independent
Agents and model switching; its second model uses `DEMO_SECOND_MODEL`.
The optional `gui` module implements that public client boundary over local
WebSocket. `examples/browser-consumer` reuses its public Server and browser
components with two independent Agents and a different layout.

Build the browser command from `gui`:

```sh
go build -o /tmp/ensemble-gui ./cmd/ensemble-gui
```

Run `/tmp/ensemble-gui --port 0 --terminal` in a scratch directory with the same
`LLM_*` environment. Open the printed local URL. Terminal EOF drains and detaches
the CLI while the browser remains usable; `/quit` closes the application.
`--gui-log PATH` optionally records conversation-bearing transport traces.

`--preferences PATH` selects the Server's display-preference file and
`--policy PATH` selects the Agent's execution-policy file. Their defaults are
`.ensemble/gui-preferences.json` and `.ensemble/agent-policy.json` beneath the
launch workspace. Settings survive a restart; conversation logs still require
a fresh path. Appearance and speech preferences do not enter core Agent code.
Headless clients use `Config.PolicyPath`, `Agent.ExecutionPolicy` and
`Agent.UpdatePolicy`; without a path policy is explicitly in memory.
The independent `examples/policy-consumer` demonstrates two Agents with
separate files, limits and fresh-log restarts.

Cards show provisional text, accepted answers, tool reports and terminal job
facts. Speech begins disabled; enable auto-speech or use a card's Speak button.
Typing and queued/current speech each hold that connection's new tool admissions.
Cancel speech preserves unfinished input. Interrupt stops the turn, not running
jobs. Closing a tab releases its causes. Reconnect restores the last 100 selected
events and current partials, with an explicit count of older omitted events.

The public `Agent.Watch` returns one atomic snapshot plus bounded tail;
`Agent.RegisterPause` creates an independently owned pause registration. Browser
embeddings create one `BrowserApplication` per document and Pages through
`createPage(root, url)`. Pages retain local queues while the application owns
one FIFO native speech service. Close a Page before replacing it on the same
DOM root. The embedding example provides Close view/Reconnect view controls.
SpeechService coordinates cooperating same-origin/storage-bucket documents
through Web Locks before native playback. It cannot isolate unrelated sites or
browser profiles. Missing coordination is reported without retaining a speech
pause. Settings use lossless numeric revision parsing; a browser lacking the
required parser support refuses visibly instead of submitting rounded revisions.

Run `go vet ./...` and `go test ./... -count=1` separately in the root,
`gui`, and each module under `examples/`; use the race detector on concurrency
checks. The independent immutable gate runs from the course repository root:

```sh
python3 scripts/edition2/accept_ch08_gate.py a06d4f3c848685d81306166279df7c977d71bddd
```

Chapter 8 receipts cover actual human CLI, browser and public consumers on all
three providers: Anthropic 18 model requests, OpenAI 14 and Gemini 15. Original
failures and bounded corrective prompts remain recorded. OpenAI/Gemini's final
speech-overlap supplements replay retained real-provider events with the endpoint
disabled and new admission timestamps; they are not fresh model generations.
Audio captures establish native output, not a human listening/transcription claim.
Earlier chapter receipts retain their original source identities and limitations.

## Chapter 5: responsive turns

`chat` accepts ordinary prompts while a turn runs and prints each request ID.
Use `/hint TEXT` for one-request guidance to the active turn, `/interrupt` to
end that turn while preserving its jobs, and `/quit` for full cleanup. EOF lets
all admitted prompts finish before cleanup. `protocol` preserves legacy user
records and also accepts explicit `kind: prompt`, `kind: hint`, and
`kind: interrupt` records with attributable reliable completions.

Embedding clients use `Agent.Submit` for a reusable request handle, `Wait` to
read its completion, and `Cancel` to cancel only that request. `Agent.Prompt`
is a blocking wrapper over the same actor. `Ensemble.Collect` creates an
independent collection that drains all currently ready handles in declared
order; cancellation of a collection wait does not cancel any request.

The optional GUI module consumes these public seams. Display subscriptions have their own bounded queues and report
`overflow` through `SubscriptionStatus`; reliable request completion does not
rely on display delivery. Chapter 6 adds the streaming observations described
below while keeping this completion boundary.

`examples/workflow` builds a public author/editor/reviewer consumer. Its
`collection` mode demonstrates two Agents, queued cancellation, reusable
handles and independent collections. Set `ENSEMBLE_RUN_DIRECTORY` to a fresh
directory and use the ordinary `LLM_*` environment configuration.

For an event log with captured request configuration,
`ensemble replay LOG SEQ` reconstructs that request from its recorded
configuration and the prefix before `request_sent`. This path performs no model
or tool effect.

## Chapter 6: streaming progress

Messages, Chat Completions and generateContent stream by default. Set
`EN_DISABLE_STREAMING=1` for ordinary JSON delivery; absent or `0` enables
streaming, and other values fail configuration. Embedding clients set
`Config.DisableStreaming`. Delivery is snapshotted for each model operation and
recorded in `request_sent`; historical requests without delivery remain plain.

Human chat flushes visible text as it arrives and labels exposed thinking and
proposed tool names/arguments separately. `/hint`, `/interrupt`, `/usage` and
queued prompts remain available. A tool proposal cannot execute until the whole
response is validated and durably accepted. Interrupted or failed output is
marked incomplete. Normal completion avoids repeating displayed text; a display
overflow is reported explicitly and the reliable completion supplies a labeled
complete answer when available. Token-limit stops carry a generation-limit
notice rather than claiming the requested task was fully completed.

`ensemble protocol --observe` opts into correlated `model_begin`, `part_delta`,
`part_final` and `model_end` JSON observations, plus an explicit observation-gap
record on overflow. Default `protocol` retains its existing output. Public
observers receive the same operation/part identities and complete accepted typed
finals; signatures and opaque material are retained for replay, not displayed as
thinking. `ensemble replay LOG SEQ` reconstructs the recorded delivery without
HTTP or provisional observations.

The separate `examples/stream-consumer` module creates two Agents with ordinary,
finals-only and deliberately stalled subscribers. Give its built executable a
fresh `ENSEMBLE_RUN_DIRECTORY` and the ordinary `LLM_*` environment. It reports
full identities, typed finals and reliable completions before releasing its
stalled callbacks. Its real-provider receipts include Gemini `MAX_TOKENS`
partial answers; no live overflow or thinking delta is claimed for those
historical Chapter 6 runs.
