# Edition 3 — Chapter 4 C# endpoint

C# 14 / .NET 10. All core partial types compile into one assembly. Common declares
persistent state, relationships and callable operations; the functional directories
supply bodies. The CLI, tests and example consumer are separate assemblies.

## Read in this order

1. `Common/Owners.cs`, `Common/Tools.cs`, `Common/Jobs.cs`: Ensemble owns Agents;
   Agent owns History, Engine, Jobs and Tools; Jobs owns Job. Parent add/remove
   methods complete both sides. Ensemble allocates handles; each Jobs collection
   restricts lookup to its own children.
2. `Agent/Agent.cs`, `Agent/Tools.cs`: one HTTP exchange per Engine call, one
   dispatcher per tool call, sequential model-visible results. Even a refusal
   follows admission; it never permits the rejected effect.
3. `Jobs/Jobs.cs`: consume the next-call observation policy, reserve the ordinary
   job, record ToolCalled, then start work. Controls do not allocate jobs.
4. `Jobs/Job.cs`: full output on disk, serialized unseen-byte cursor, head/tail
   inline bounds, completion notifications and explicit cancellation.
5. `Jobs/Pty.cs`: macOS `/usr/bin/script` supplies a real 50×200 PTY. C# owns its
   Process and stdin, forwards bytes, strips provider keys, kills the child
   process group and drains it. This endpoint does not claim Linux support.
6. `History` and `Serialization`: job snapshots are immutable recorded values.
   Result redaction retains the actual output locator. History remains the only
   authoritative conversation; live processes are not serialized into it.
7. `Examples/Consumer`: a separate public-API client registering `shout`. Its
   callback receives the actual Job and reaches diagnostics through parents.
   This concrete callback is required for external custom behavior; built-ins
   remain direct methods, not a callback registry.

`ai_callback_delay` is a wake delay, never an execution timeout. The regex and
inline byte cap also control observation only. `tool_limits` sets the three
values for exactly the next dispatch, including controls and failed calls.
Call arguments override the pending values. The returned notice identifies
which call consumed them. No deadline or hard output cap is added.

Explicit kill retains `killed`, records its reason and wakes observers. Shutdown
kills remaining work before removing History or other parents. .NET cannot
forcibly stop arbitrary managed callbacks: they receive a cancellation token;
late output/completion cannot resurrect a killed job. PTY processes are killed
and joined. Finished OS process handles are closed promptly, while output files
remain available after logical destruction.

## Build and verify

Use the approved .NET 10 SDK in an owned environment with first-run setup and
telemetry disabled; no test framework or new package installation is needed.

```sh
dotnet run --project Tests
dotnet publish Cli/Ch04.csproj -c Release -o /your/owned/ch04-build
dotnet build Examples/Consumer/Consumer.csproj
```

The CLI keeps the `chat`, `dump`, `render`, JSON-line stdin and `CH02_LOG`
contracts. Its job output is `cr/io/<handle>` under the working directory.
Library Agents default to independent temporary directories; callers may supply
an explicit output directory. Use a private workspace per concurrent CLI.

From the repository root:

```sh
go run ./cmd/grade-edition2 -ch 4 -jobs-only -json /your/owned/ch04-build/ch04
```

This scope grades the executable job protocol and Chapter 3/2 compatibility. It
explicitly excludes Go consumer checks. Independently exercise the C# consumer
against a loopback fake using the environment described in the chapter. Source
inspection and the actual CLI debugger exercise are also required; a score
cannot establish architectural compliance.

Evidence and limitations belong in `docs/edition-3-csharp-progress.md`. Real
provider use remains incomplete: local protocol fakes do not establish model
judgment, provider acceptance or production reliability. Earlier chapter
endpoints remain unchanged.

## Chapter 5 actor and executable inputs

`SubmitAsync` returns a reliable request task. `PostAsync` acknowledges live
input/control admission. `WaitAsync` and root `WaitAnyAsync` consume bounded,
lossy progress; competing readers are not independent subscriptions. HTTP receives
frozen bytes and posts its result back; only the actor mutates conversation state.
Caller cancellation detaches a wait. Explicit interrupt ends the turn but preserves
independent Jobs. `FlushAsync` waits for foreground quiescence without closing
admission; it does not resume paused work or wait for background Jobs.

The JSON-lines CLI keeps `user` and `ephemeral` compatibility. `user` requests
queue separate ordered reliable replies while stdin stays responsive. `parts`,
when supplied, overrides user text using the public codec. Live inputs use `kind`:
`prompt`, `hint`, `pause`, `resume`, `interrupt` (mode `now` or `after_tool`),
`tool` (`id`, `name`, `arguments`) and `shutdown`. Admission produces `ack`; live
progress is separately labeled `observation`. EOF flushes turns then shuts down
Jobs; explicitly paused queued work requires resume or shutdown before EOF.

Local-path images render for the exact declared Anthropic/OpenAI/Gemini models;
Gemini also declares audio, video and documents. Unknown models can send text but
not attachments. URI and handle resolution are deferred. These are offline
renderer declarations, not real-provider certification.

`Examples/Workflow/Workflow.csproj` is a separate consumer assembly. It builds
three named Agents with external role prompts and toolsets, sharing one synchronized
value-like manuscript. Its public JSON protocol is documented in Chapter 5.
Use an owned loopback Anthropic-protocol fake, `LLM_*` environment and `CH06_LOG`.
It writes tagged progress and reliable pipeline markers, keeping each Agent's
History in a separate sibling log. EOF awaits the submitted workflow.

The retained Go grader's `-ch 5 -executable-only` route reuses only request replay
and media checks (25 points). It excludes Go consumer imports, workflow, topology
and Chapter 4 parity. The separate C# consumer, component tests and architecture
inspection are distinct evidence. Real-provider execution remains incomplete.
