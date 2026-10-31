# Edition 3 — Chapter 14 C# endpoint

C# 14 / .NET 10. Common declares the core vocabulary, persistent fields,
relationships and cross-module operations. Functional directories implement the
matching partial types in one assembly. CLI, tests and examples are separate
assemblies using that public API; no additional packages are required.

## Read in this order

1. `Common/Owners.cs`, `Common/Actor.cs`, `Common/Streaming.cs`: Ensemble owns
   Agents; Agent owns History, Engine, Jobs, Tools and optional Skills. The Agent mailbox owns
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

## Chapter 14: recorded context reduction and crash recovery

Read `Common/Persistence.cs`, `Agent/Compaction.cs`, `Agent/ResultStubbing.cs`,
`Agent/Handoff.cs`, `Agent/ClearTools.cs`, then `History/Validation.cs`,
`History/Handoff.cs`, `History/Journal.cs` and `History/Archive.cs`. Agent chooses
cuts at completed-batch boundaries; History applies recorded facts. Workers do
not edit History. Full canonical events and reduced Context serve different uses.

One Agent-owned target defaults to 400,000 bytes; result/call-stub bands derive
T/8 and T/16, with hysteresis above twice each band. The API and CLI target zero
disables automatic cuts. `--context-budget-bytes N` overrides
`EN_CONTEXT_BUDGET_BYTES`. Explicit `--stub-tool-results` (or
`EN_STUB_TOOL_RESULTS=1`) adds once-only T/100 stubbing and `keep_tool_results`;
it is off by default. A keep is not immunity from the ordinary budget ladder.

`micro_handoff` records its ordinary result and companion outcomes before the
cut. The newest marked note replaces earlier active notes; tool traffic and
private replay material leave the working Context, while dialogue and manuals
remain. `InputKind.ClearTools`, CLI `kind: clear_tools`, and the GUI button clear
tool traffic through that same actor boundary without cancelling independent Jobs.

`StartPersistence` now enables a sibling journal. Readers can access last
committed state during disk work; writers serialize. Recovery validates primary,
backup and contiguous suffixes before publication. Only an incomplete final
journal record may be ignored. Retention stays beyond a validated backup, and
corrupt primary bytes never replace a good backup. This is process-crash evidence,
not a hardware power-loss guarantee. Optional archives preserve all original
events and cuts; unavailable earlier events cannot be invented from Context.

The component suite has 89 groups. The scoped eight-check executable grade is
100/100. Separate public-API, CLI and plain/SSE browser exercises cover distinct
paths; see the chapter and progress record for evidence and limitations. The
Chapter 14 executable grader requires `-gui EXECUTABLE`, unlike earlier chapters'
static-assets argument. No real-provider or native/human-accessibility result is
credited at this endpoint.

## Chapter 13: speech and a restricted listener

`Examples/Gui/wwwroot/speech.js` formats raw part text before phrase splitting;
`app.js` owns queue admission, source labels, silent transcript mode and the
independent human-typing/speech pause causes. Whole finals and lifecycle provider
errors use that same speech path. Replay stays silent. Run the portable formatter
regression with `node Tests/SpeechTests.mjs`.

`Examples/Listener` is a separate C# assembly using the public library. Its empty
Agent receives exactly five custom tools and no Skills/builtins. Read its Common
vocabulary, Channel operations, then Browser protocol/lifetime and Program.
Its README documents explicit target/listener configuration and owned sessions.
Custom channel tools opt into `waitForCompletion`; the existing Job owns execution,
shutdown cancellation and one atomic result. Other custom tools keep timed reports.

`scripts/tts-harness.sh` exercises the real GUI against the speech grader's local
provider fixtures. Existing external test-driver paths are explicit; see its README.
Neither transcript mode nor fake native callbacks prove audible accessibility.

## Chapter 9: skill ownership and activation

Read `Common/Skills.cs`, `Agent/Skills.cs`, then `Skills/Skills.cs` and
`Skills/Parser.cs`. Agent owns its manager; recorded immutable `SkillPart` values
supply both instruction bodies and tool grants. There is no second mutable
active-tool registry. A load validates and publishes one dependency batch.
Pending unload preserves SYSTEM instructions but immediately revokes pending-only
grants on dispatch, declarations and discovery surfaces. Server ownership also
closes that skill’s external connections and remote tools. The primary is rendered once;
loaded bodies appear once in each provider's instruction field, not ordinary
tool-result dialogue. The actor performs activation before scheduling the next
request; disabled calls are refused before effects.

CLI and GUI now default to the shipped `ensemble` primary: inspection tools plus
an offered `code-tools` manual for editing and job control. For an explicitly
full local exercise use `EN_PRIMARY_SKILL=full`. Both applications support
`--skills-dir PATH`, `--primary-skill NAME`, `EN_SKILLS_DIR`, `EN_PRIMARY_SKILL`,
and the harmless example variable `EN_CUSTOM_VAR`. Files are copied beside the
executables; the library reads no environment. Open logs before attaching Skills.

`EN_DEFERRED_TOOLS=1` explicitly opts an Anthropic host into deferred declarations
and paired discovery references. It is off by default and is not a model-support
probe. Deferred schemas still travel in every request. The primary's tools stay
eager. Canonical discovery metadata survives logging/redaction; offline `render`
without a manager uses portable enabled declarations and recorded manuals rather
than rereading the catalog.

`Examples/Skills` is a separate public-API consumer. Give it a new owned directory:

```sh
dotnet run --project Examples/Skills/Skills.csproj -- /new/owned/skills-exercise
```

It performs real refused/accepted writes, variable snapshot checks, provenance,
a real dynamically granted write, and refusal after pending unload without provider access.
The Chapter 9 tests also drive actual actor/HTTP load-and-write loops in ordinary
and deferred modes. Local CLI exercises cover all three dialects and key-free
replay with the catalog unavailable; an isolated browser covers skill/result
cards and stable replay. The scoped executable grade is 85/85, excluding inherited
Go topology. None of these establishes real-provider support or human hearing.

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
`CH02_LOG`. JSON inputs include queued `user`, live `kind: prompt`, `hint`,
`pause`, `resume`, `interrupt`, `shutdown`, and direct `tool` operations as
specified in Chapter 5. EOF drains accepted turns; shutdown cancels them.

Streaming is requested for declared capable models unless `-no-stream` or
`EN_DISABLE_STREAMING=1` opts out. Plain parsing emits finalized parts without
synthetic deltas. Live JSON (`kind: prompt`) emits flushed `part_delta` and
`part_final` records with response-local IDs. Legacy queued user requests keep
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


## Chapter 8 browser application

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

## Chapter 8 settings and turn budgets

`Agent.MaxToolRounds()` reads the runtime policy; `SetMaxToolRounds(int)` accepts
nonnegative values (zero = auto/200) and refuses negatives without mutation.
The actor captures the effective limit at turn start. Constructor configuration
still supports Chapter 5's library-only negative unlimited seed. Rendering a
budget instruction is request-local and never changes canonical History.

The optional GUI persists its presentation preferences plus the Agent's policy
snapshot. It saves before applying/broadcasting, validates disk and wire values
with one policy, and has no second live budget field. See `Examples/Gui/README.md`
and `Common/Settings.cs` in that application for the protocol and ownership.

## Chapter 10 persistence

Read `Common/Persistence.cs`, `History/Persistence.cs`,
`Serialization/Persistence.cs`, and `Agent/Persistence.cs` together. History owns
one reduced Context and its folded sequence; diagnostic retention cannot reset
that sequence. Save candidates and old primaries are validated before replacement.
Corrupt primary bytes never rotate over a good backup.

CLI and GUI now load/save `save.json` in their working directory by default.
Use a separate directory per session, `--save PATH` to choose one, or `--no-save`
for an intentionally ephemeral exercise. `--save-log N` selects all (negative),
none (zero), or the newest N diagnostic events. `--archive-log PATH` optionally
writes a complete event archive before trimming; `History.Load` can replay it.
The recovery journal is independent of this optional archive. Environment fallbacks are `EN_SAVE_PATH`, `EN_SAVE_LOG`,
`EN_ARCHIVE_LOG`, and `EN_DISABLE_SAVE=1`.

Library callers explicitly call `Agent.StartPersistence` before actor admission.
Construct current skills first, load/reconfigure next, then open the diagnostic
log. StartLog writes the restored/bootstrap prefix before recording new events.
Shutdown joins work, records completion and saves before logical destruction.
Current runtime configuration overrides saved skill grants, not ordinary historic
system messages. The next primary/skill bodies come from today's explicit inputs.

`Examples/Persistence` is an independent assembly using only public APIs. It takes
one new directory and demonstrates snapshot-only restore, a runtime model change,
and continuous numbering without a network provider. Chapter 10's scoped CLI
grade and isolated GUI restart/browser exercise are separate evidence; neither
claims real-provider validation or crash-proof unsaved-event recovery.


## Chapter 11 MCP and on-demand browser tools

Start with `Common/Mcp.cs` and `Common/McpBrowser.cs`, then read `Mcp/Ownership.cs`,
`Mcp/Client.cs`, and `Mcp/Service.cs`. The Agent owns its MCP manager; clients have
real manager, optional skill, and optional browser parents. Remote Tools belong
to both Agent and client. Destruction closes admission immediately, joins owned
I/O and reverse work, and removes both directions of these relationships.

`Skills/Connections.cs` prepares connections off-actor and commits one validated
activation on the actor. Discovery never grants unlisted tools. Construct Skills,
load persistence, open the diagnostic log, then await `Skills.InitializeAsync`
before user work. For a server-backed GUI primary, start the listener first;
`RequiresBrowserStartup()` distinguishes a WebSocket dependency from stdio-only
startup, which must not wait for a browser. Failed startup cannot overwrite a
previous good save. Saved configuration is documentary, not a process-restart
recipe.

`Agent/Automatic.cs`, `Mcp/Automatic.cs`, and the Job completion path collect
`per_turn_` / `per_round_` definitions as real independent Jobs. Their completed
output becomes request-only system ephemera. Automatic names never become model
or reverse-call declarations, `$TOOLS` entries, or discovery references. Interrupt
abandons collection, not the Job. Failed snapshot bytes remain inspectable in the
Job; the model receives only a short unavailable/error diagnostic.

The CLI accepts `--mcp-pipe` with ordinary application/skill configuration. An
external client sends `initialize`, then `notifications/initialized`, before
`tools/list` or `tools/call`. Stdout is protocol-only; EOF leads to joined shutdown.
This incoming role differs from outgoing skill connections, where Ensemble sends
the handshake. File-tool service work requires no model HTTP request.

The shipped `browser` skill is loadable from the default primary. It offers
`gui_snapshot`, `gui_click`, `gui_input`, `tts_queue`, and `tts_say`; none runs
automatically. One browser registration can supply tools, while observer tabs
remain unable to inject its replies. See the GUI README for connection lifetime.
General MCP resource blocks are explicitly unsupported; media placement is still
subject to each provider renderer's capabilities.

The scoped executable grader is:

```sh
grade-edition2 -ch 11 -executable-only -json /absolute/cli-launcher
```

Its five CLI checks total 80 points. Real local component/GUI/browser exercises and
a separate public-API consumer are additional evidence, not implied by that grade.
No live-provider or audible/human-accessibility acceptance is claimed.

## Chapter 12: optional startup skills

`SkillSpec.StartupSkills` is an optional ordered list frozen at construction.
The primary and selected skills resolve and publish as one initial activation;
server-backed selections finish through `Skills.InitializeAsync`. Readiness and
browser-wait decisions cover the complete selection, not only the primary.
Loaded manuals use each provider's instruction field without invented tool calls.

Both hosts accept `--startup-skills name,other` (default `EN_STARTUP_SKILLS`) and
`--gui-debug` to append that skill. The shipped `gui-debug` owns a WebSocket
browser connection; use the GUI host for it. A CLI startup exercise needs its own
usable stdio declaration, not an imaginary GUI listener. Default startup is unchanged.

The browser's marked MCP input events do not create human typing. Speech callbacks
use the retained human cause rather than guessing from field contents. Human
whitespace still pauses; successful submission explicitly clears that cause.
Snapshots and speech-queue observations remain on demand. See
`Tests/StartupSkillTests.cs`, `Skills/Startup.cs`, and `skills/gui-debug/SKILL.md`.
