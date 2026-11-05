# Chapter 3 retained-result metadata supplement

This outside-assembly consumer tests the early metadata contract in Chapter 3.
It references only the candidate `Ensemble.dll`: no candidate test assembly,
reflection, Jobs, Sandbox, GUI, live offer collection, or HTTP request is used.
It awards no chapter points and does not replace `grade-edition2 -ch 3`.

Use the existing .NET 10 SDK. Build the candidate first, then supply its absolute
assembly path. Use fresh owned build and working directories, an isolated HOME,
TMPDIR and DOTNET_CLI_HOME, and disable first-run certificate generation.
The supplied NuGet configuration clears package sources; no package installation
is needed for this consumer.

```sh
/path/to/dotnet build /bundle/ch03-metadata-consumer/Chapter3Metadata.csproj \
  --configfile /bundle/ch03-metadata-consumer/NuGet.Config \
  --artifacts-path /your/owned/metadata-build \
  -p:CandidateAssembly=/absolute/candidate/Ensemble.dll --tl:off -v:q
cd /your/owned/empty-workspace
/path/to/dotnet /your/owned/metadata-build/bin/Chapter3Metadata/debug/Chapter3Metadata.dll
```

Success requires exit 0, `contract: chapter3-metadata-v1`, `passed: true`, and all
41 named checks from this consumer. A failed build is not a behavioral failure receipt.
Keep stdout and build logs. Check names distinguish validation, codecs, numbering,
reduction, redaction, and three early provider-rendering surfaces. Randomized
ordinary payloads must remain visible in rendered requests; whole-request equality
then checks that adding retained metadata changes none of their contents.

The chapter specifies the public value and optional named constructor arguments.
The consumer owns the 64 MiB bound independently of any candidate constant.
Legacy absent/null metadata remains absent; pathname prose does not manufacture
an offer. Result redaction removes the description without mutating the original
Context. These are data-boundary checks, not evidence that a file exists or that
any Agent has authority to download it. Actual synchronous dispatch propagation
still requires source review and focused candidate tests; this value consumer
alone does not certify an effectful tool loop.

No live-provider compatibility, running-offer behavior, browser download,
restart-expiry, student reconstruction, or full Chapter 3 acceptance is implied.

## Library registry and ownership mode

Use the same build and fresh owned environment, adding `library`:

```sh
/path/to/dotnet /your/owned/metadata-build/bin/Chapter3Metadata/debug/Chapter3Metadata.dll library
```

Require exit0, `contract: chapter3-library-v1`, `passed: true`, and all48 named
checks. The default metadata mode retains its separate41-check contract.
Unsupported arguments exit2. For each early provider route, this mode constructs
two Agents, registers a concrete `ReadFile` Tool with only one, and verifies the
actual declaration and the empty peer's omission. It mutates returned JSON trees,
then requires fresh projections to remain unchanged. Destroying the last Tool
must restore the original empty request; destroying its owner must detach its
child without changing the peer's independent Tool.

Only `Tool(Agent, ToolKind)`, `Parent`, `Name`, `Declaration`, `Destroy`,
`Agent.Tools`, and the existing renderer/owner APIs are required. There is no
custom callback, Job, live skill, command execution, private access, or network.
These checks supplement the CLI grader, whose installed registry is never empty.
They do not certify all schemas, nested-schema projection, effectful dispatch,
thread safety, or the historical chapter score.

## Separate-process metadata log replay

After building, enter a fresh owned workspace and run these commands sequentially:

```sh
/path/to/dotnet /your/owned/metadata-build/bin/Chapter3Metadata/debug/Chapter3Metadata.dll log-write
/path/to/dotnet /your/owned/metadata-build/bin/Chapter3Metadata/debug/Chapter3Metadata.dll log-read
```

Require the writer to exit0 before starting the reader. Both receipts use
`contract: chapter3-metadata-log-v1`, their exact `mode`, and `passed: true`.
The writer has4 named checks; the reader has7. Keep both receipts. The reader
has no input-generation fallback: missing writer files must fail.

The writer uses public `History.StartLog`, `Append`, and `Dump`, then destroys
its owner before exit. It compares log and dump at the same lifecycle boundary,
since destruction may append a shutdown fact. An independent manifest records
random call identity, ordinary payload, and expected description fields. The
reader uses public `History.Load` and checks canonical events, restored Context,
and redumped descriptions for both actual log files. No later snapshot API is
required. This tests clean process exit and replay of data, not crash recovery,
restored file authority, or synchronous tool execution.
