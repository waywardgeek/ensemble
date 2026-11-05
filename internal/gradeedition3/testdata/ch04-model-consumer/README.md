# Chapter 4 model-driven public consumer

This separate C# executable references only the candidate `Ensemble.dll` and the
existing .NET 10 ASP.NET shared framework. It supplies an owned loopback protocol
fake, not an Agent implementation. All model turns and tool effects go through
the candidate's actual public Agent loop. No reflection, internal test access,
GUI, external provider, package installation, or downloaded dependency is used.

## Build and run

Use a fresh owned workspace, an approved existing .NET 10 SDK and installed
ASP.NET shared framework, and an isolated environment with HOME, TMPDIR and
DOTNET_CLI_HOME set. Clear inherited provider settings and credentials. Disable
first-time setup, telemetry and certificate generation. Use the included
cleared-source NuGet configuration; a missing framework or dependency is a setup
failure, not permission to install it.

```sh
"$DOTNET" build Chapter4Model.csproj \
  --configfile NuGet.Config --artifacts-path "$BUILD" \
  -p:CandidateAssembly="$CANDIDATE/Ensemble.dll" \
  -p:AgentEntryPoint=Ask --tl:off -v:q
cd "$WORK"
"$DOTNET" "$BUILD/bin/Chapter4Model/debug/Chapter4Model.dll"
```

`AgentEntryPoint=Ask` calls the accepted early prerequisite's public
`Agent.AskAsync(string, CancellationToken)` and ordinary four-argument
Configuration constructor. Chapter 4 still uses buffered requests. The default
`AgentEntryPoint=Send` is for the current runtime's public `Agent.SendAsync`; it
also sets that later runtime's `disableStreaming: true` option. Both compile-time
branches invoke the candidate exactly once and contain no exchange/dispatch loop.
Unknown selector values refuse compilation. Select the appropriate public API
explicitly; do not add a second loop or use private reflection to make it compile.
The current-runtime positive and negative receipts exercise `Send`; the early
`Ask` branch still needs the Chapter 4 candidate's actual build and execution.

The remaining interfaces are the Chapter 4 public Tool constructor, real Job
parent chain, `InstallTools`, canonical History and values, Job snapshots, and the
root's timestamped `Log` method. Those live on their real owners, as taught in the
chapter and the observation supplement. No Sandbox, mailbox, skill manager or
structured custom-result extension is required by this executable.

Run only on a supported platform for the candidate's real `run_command` path.
The current verification is macOS with the existing PTY implementation. The
synthetic shell command only prints a random marker inside the disposable
workspace; this is not a debugger acceptance run or broad OS portability proof.

## Required result

Require exit zero, `contract: chapter4-model-consumer-v1`, `passed: true`, and all
38 named checks. The fake uses an OS-assigned loopback port and refuses the
protected 8088/8089/8090 ports. Its synthetic key never comes from the environment.
A 60-second deadline bounds the normal scenario. Shutdown destroys the root and
stops/disposes the owned host before reporting; successful receipt alone is not
an authorization to leave a host running.

The first real request must declare the custom `shout` and all six existing
built-ins. The fake responds with an explained three-call batch: `shout`,
`read_file`, and `run_command`. The callback checks its actual parent chain and
record-before-effect ordering, emits its timestamped diagnostic through the
root, and uppercases randomized input. Before returning a final answer, the
second provider request must satisfy all of these:

- Exactly three successful correlated canonical results and distinct owned Jobs.
- Each real expected output in canonical result, durable output file, and actual
  provider payload; output file byte counts match the snapshots.
- A successful shell exit and exactly one custom callback invocation.
- The actual timestamped root diagnostic within the scenario's time bounds.

Only then does the fake send a random final answer. The consumer requires that
exact answer and exactly two requests. An unknown-tool result marker cannot pass.
Neither can a candidate that fabricates the right uppercase output while skipping
the handler: the invocation check is independent of output content.

This is offline buffered Anthropic-protocol integration evidence. It does not
award the historical Go-specific consumer/topology points or a new chapter
score. Review Common declarations and ownership/readability independently. Live
providers, debugger behavior, uncooperative callbacks, persistence, streaming,
and completed structured-result delivery remain separate gates. The latter is
now a Chapter11 MCP/browser exercise, not a Chapter4 admission blocker. The
text-only custom handler does not certify a delayed work-product offer; that
specific integration remains in Chapter26.
