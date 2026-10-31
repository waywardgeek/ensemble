# Chapter 3 — Six Tools (C#)

C# 14 / .NET 10, standard library only. This is the Chapter 3 endpoint, not an
installation or replacement of the running Go Ensemble application.

## Read the model first

Keep Chapter 2's Ensemble → Agent → History and Engine relationships. Agent now
also owns six concrete Tools. Start with `Common/Tools.cs` and the additions to
`Common/Owners.cs`, then read `Agent/Tools.cs` and `Tools/Tool.cs`. Constructors
call Agent.AddTool; add/remove complete both ends before returning. Agent.Destroy
cascades through tools. Immutable schemas and conversation parts are values,
not relationship-bearing children.

Each Tool holds its name, description and schema once. Engine projects that same
Agent-owned collection for each vendor; Agent dispatches through it. An empty
library registry emits no declarations and returns errors for unknown names.
The CLI explicitly installs the six tools. There is no global registry, service
bag, callback plumbing or plug-in framework.

Follow `Agent.SendAsync`: record input; ask Engine for one exchange; inspect the
recorded ResponseEnded; record and execute every call in order; record every
result; repeat. Return only the final reply. Pre-tool narration goes through the
parent diagnostic path to stderr, leaving JSON stdout stable. A sixteen-round
bound limits repetition, not command or HTTP duration. `LLM_REQUIRE_TOOL_EXPLANATION`
accepts true/false (default false); when enabled, a calls-only response gets
paired error results without effects so the model can explain and retry.

Then read `Tools/FilesAndShell.cs`. File edits require one exact nonempty anchor,
including overlapping-match detection. Reads default to 32768 bytes after line
selection and visibly report truncation. Search merges adjacent context windows.
Commands use a fresh shell, drain both output pipes, report separate streams and
exit_code, remove provider credential variables, and honor caller cancellation.
Nonzero exit is ordinary data; tool validation and file failures are error results.

History is still the authoritative event record. Request projection does not
rewrite it. Call IDs normalize at capture and duplicates within one response
are rejected before effects. Gemini call- and text-bound signatures survive
serialization and replay only with exact provenance.

## Build, test and exercise

Use .NET 10 from PATH or Bill's approved private SDK by explicit path. In an
owned environment, disable first-run certificate generation and telemetry.

```sh
dotnet run --project Tests
dotnet publish Cli/Ch03.csproj -c Release -o /your/owned/ch03-build
/your/owned/ch03-build/ch03 chat
/your/owned/ch03-build/ch03 dump /your/owned/events.jsonl
```

The CLI retains Chapter 2's JSON-line `user`/`ephemeral`, `chat`, `dump`, `render`
and CH02_LOG contracts. Provider configuration uses LLM_VENDOR, LLM_MODEL,
LLM_BASE_URL and LLM_API_KEY (vendor-specific fallbacks remain). Always use an
explicit loopback endpoint and dummy key for offline exercises.

From the repository root, the existing executable grader accepts the compiled
apphost; it does not require a Go solution:

```sh
go run ./cmd/grade-edition2 -ch 3 -json /your/owned/ch03-build/ch03
go run ./cmd/fakevendor -vendor anthropic
```

The grader uses the already-installed Go toolchain for small shell fixtures.
Use its serve-only fake's printed environment with our executable in a disposable
project containing go.mod; repeat for openai and gemini. Do not use its historical
`-ch 3` convenience runner as evidence for this C# solution.

## Scope and evidence

The component suite uses real files, shells, objects and HTTP exchanges; only the
remote model is faked. It covers the loop, mixed outcomes, strict types, ownership,
round bound, selected explanation retry, signatures, and cancellation, plus the
Chapter 2 regressions. The Chapter 3 executable grade and actual CLI exercises
are recorded in `docs/edition-3-csharp-progress.md` at the repository root.

Real-provider declaration acceptance and coding-work exercises remain incomplete:
the paid-provider allowance is exhausted. No streaming, background job manager,
sandbox/confinement, deployment, GUI implementation or account changes are claimed.
The synchronous command limitation is intentional groundwork for Chapter 4.
