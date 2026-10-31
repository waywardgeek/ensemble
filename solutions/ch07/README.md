# Edition 3 — Chapter 7 C# endpoint

C# 14 / .NET 10. Common declares the core vocabulary, persistent fields,
relationships and cross-module operations. Functional directories implement the
matching partial types in one assembly. CLI, tests and examples are separate
assemblies using that public API; no additional packages are required.

## Read in this order

1. `Common/Owners.cs`, `Common/Actor.cs`, `Common/Streaming.cs`: Ensemble owns
   Agents; Agent owns History, Engine, Jobs and Tools. The Agent mailbox owns
   conversation mutation and scheduling. Reliable request tasks are separate
   from bounded, lossy Agent/root observation queues.
2. `Agent/Actor.cs`: controls remain responsive during HTTP and foreground Job
   waits. Workers return completion messages; they do not write History. Chunk
   messages are transient. Recorded final events precede final observations.
3. `Engine/Engine.cs`, `Engine/Streaming.cs`, `Engine/Parse.cs`: frozen request
   bytes, untimed cancellable HTTP, one JSON/SSE response seam and shared final
   decoding. Parser-supplied IDs correlate deltas with finals even when arrival
   order differs from final order. Partial arguments never authorize effects.
4. `Engine/Render.cs`, `Engine/Media.cs`: explicit model capabilities, request-only
   local-media resolution, unknown-model plain fallback, and configuration
   opt-out. Stream capability is not inferred from media support.
5. `Jobs`: reserve → record call → start effects; full output on disk, serialized
   unseen-output cursor, observation bounds, process lifetime and explicit kill.
   Job handles are Ensemble-wide but each Jobs owner limits lookup to its children.
6. `History` and `Serialization`: one authoritative conversation, exact JSON
   arguments and provenance-aware opaque data. Deltas do not inflate history or
   accounting. Host-only tool observations do not invent model dialogue.
7. `Application/Ensemble.cs`: complete two-way relationships and optional sanitized
   tracing through real parents. The caller owns the trace writer and closes it
   after the root is destroyed.
8. `Examples/Observer` and `Examples/Workflow`: delivered-stream statistics and
   multi-Agent display respectively. Reliable tasks govern completion; observations
   never advance workflow stages. `Examples/Consumer` retains the small custom
   tool example from Chapter 4.

## Build and exercise

Use the approved SDK in an owned offline environment, with first-run setup and
telemetry disabled. Build from this endpoint directory:

```sh
dotnet run --project Tests
dotnet build Cli/Ch04.csproj
dotnet build Examples/Observer/Observer.csproj
dotnet build Examples/Workflow/Workflow.csproj
dotnet build Examples/Gui/Gui.csproj
```

The CLI project retains its historical name. Its output is `ch04.dll`, but it
references this endpoint's current library. Run binaries from an owned scratch
workspace, not the source directory. Configure `LLM_VENDOR`, `LLM_MODEL`,
`LLM_BASE_URL`, and `LLM_API_KEY`; use dummy credentials with an owned loopback
fake. Job output goes to `cr/io/<handle>` under the working directory.

The CLI retains `chat`, `dump`, `render`, `-media-capabilities`, JSON stdin, and
`CH02_LOG`. JSON inputs include blocking `user`, live `kind: prompt`, `hint`,
`pause`, `resume`, `interrupt`, `shutdown`, and direct `tool` operations as
specified in Chapter 5. EOF drains accepted turns; shutdown cancels them.

Streaming is requested for declared capable models unless `-no-stream` or
`EN_DISABLE_STREAMING=1` opts out. Plain parsing emits finalized parts without
synthetic deltas. Live JSON (`kind: prompt`) emits flushed `part_delta` and
`part_final` records with response-local IDs. Legacy blocking requests keep
reliable reply-only output. Interactive `chat` previews chunks on stderr and
writes each reliable completed answer once to stdout. It suppresses ANSI styling
when stderr is redirected. `CH06_API_LOG` selects the sanitized payload trace,
which defaults to `api.log`; do not commit traces or credentials.

From the repository root, a compiled grader can exercise the CLI and external
aggregate observer launchers:

```sh
grade-edition2 -ch 6 -executable-only -json \
    -stream-model claude-3-haiku-20240307 \
    -stream-exercise /absolute/observer /absolute/cli
```

This 80-point scope excludes inherited Chapter 5 and Go-specific consumer/topology
checks. The full Go route remains intact. All-vendor parser tests, source review,
gated pre-EOF visibility, interrupts and the actual three-role workflow remain
separate evidence. See `docs/edition-3-csharp-progress.md` in the repository for
results and limitations; a green grade is not an architecture verdict.

## Lifetime and evidence limits

`ai_callback_delay`, patterns and inline caps control observation, never job
execution deadlines. Busy hints and interrupts stop foreground observation but
preserve independent Jobs. `FlushAsync` waits for accepted turns and foreground
controls, not independent background work; it does not resume a paused Agent.

Destruction stops and joins work before detaching History or parents. .NET cannot
forcibly stop an arbitrary managed callback: cooperative cancellation is required,
and a callback that ignores it can delay shutdown. PTY processes are killed and
joined. The PTY implementation uses macOS `/usr/bin/script`, not a claimed portable
Linux implementation. Completed output files remain after logical destruction.

Progress is bounded and lossy. Aggregate counts describe delivered observations,
not billing or a reliable transcript. Use History usage and reliable request
results for authoritative facts. Model capability rows and fake-provider passes
do not establish live provider availability, streaming acceptance, media decoding,
or paid integration. Those exercises remain incomplete without separate authority.


## Chapter 7 browser application

`Examples/Gui` is a separate ASP.NET Core application over the public library,
not a dependency of the core. It uses the installed shared framework; no packages
or frontend downloads are needed. Run from an owned workspace and loopback port.
See its README for flags, logs, and the WebSocket control vocabulary.

History now supplies locked immutable snapshots. Completed response events retain
response/part display identities; old logs without them still load. Admission
records `ToolStarted` only after validation and before effects. Pause, resume,
and joined shutdown are recorded, so reconnect need not invent lifecycle state.
These are canonical facts, not a GUI-owned copy of the actor state.

The server has bounded per-client queues, replay/live handoff, periodic durable
catch-up, and transient partials. The browser safely renders typed cards, preserves
full speaker payloads, and implements edge-triggered typing/queued-speech pause.
Explicit unpause from another tab wins while a local cause is unchanged.

Chapter 7 verification includes core snapshot/identity/admission/shutdown tests,
real loopback HTTP/WebSocket exercises, an isolated headless-browser exercise,
and the scoped executable GUI grader. Fake speech verifies payloads and queue
cleanup, not audible playback. Provider acceptance and human accessibility remain
separate, incomplete exercises; consult the progress record for exact evidence.
