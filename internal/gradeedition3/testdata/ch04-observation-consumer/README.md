# Chapter 4 public observation supplement

This is a separate .NET 10 / C# 14 executable referencing only the candidate
`Ensemble.dll`. Its default mode runs 23 named checks through public APIs. It uses
no reflection, candidate test assembly, model request, GUI, shell command, or
native capture.
It does not change the historical 100-point `-jobs-only` grade.

## Public interface

Use the Chapter 1 composition root and Chapter 2 values, codec and History APIs.
The C# adapter contract additionally uses these Chapter 4 members:

- `Tool(Agent, ToolKind)` with `ToolKind.WaitForJob`.
- `Tool(Agent, string, string, JsonElement,
  Func<Job, JsonElement, CancellationToken, Task<string>>)` for a custom handler.
  A compatible overload with optional trailing arguments also works.
- `Agent.DispatchAsync(ToolCallPart, CancellationToken)` returning
  `Task<ToolResultPart>` through the ordinary dispatch path.
- `Job.Handle`, `Job.Parent()`, and `Job.Snapshot()` returning immutable `JobInfo`.
- `Jobs.Parent()` and `Jobs.Children()` returning an immutable collection.
- Immutable `JobInfo` with `int Handle`, non-null `string Status` and
  `string OutputPath`, `long Bytes`, `bool Truncated` (default `false`),
  `int? ExitCode` (default `null`), and non-null `string Cwd` (default `""`).
  `Truncated` describes the bounded observation, not truncation of durable output.
  Empty `Cwd` means no per-call override; execution still uses the Agent workspace.
  The optional wire `cwd` may be absent or empty for that case.
- `ToolResultPart.Job`, `Event.Job`, `Event.ObservationDuration`, and `Event.Seq`.
- `EventKind.ToolCalled` and `ToolStarted`, plus the early paired `ToolReturned`.

Later runtimes may record a direct caller's result as `ToolObserved`. The consumer
accepts that spelling too, without requiring the early candidate to add that
vocabulary or a mailbox. It still requires exactly one real paired canonical
result for each call ID. Do not add a test-only dispatch route or fabricate facts
in an adapter. Common declares the actual types and callable members; functional
partials implement them.

## Run offline

Choose fresh absolute directories for the consumer's build and working state.
Use an existing approved SDK, isolated HOME/TMPDIR/DOTNET_CLI_HOME, disabled
first-time setup/telemetry/certificate generation, and the included cleared-source
NuGet configuration. The consumer itself has no package dependencies.

```sh
"$DOTNET" build Chapter4Observations.csproj \
  --configfile NuGet.Config --artifacts-path "$BUILD" \
  -p:CandidateAssembly="$CANDIDATE/Ensemble.dll" --tl:off -v:q
cd "$WORK"
"$DOTNET" "$BUILD/bin/Chapter4Observations/debug/Chapter4Observations.dll"
```

Require process exit zero, `passed: true`, and all 23 named checks, not just a
successful compilation. A missing candidate API is an incomplete interface, not
a behavioral failure receipt. The SDK may print terminal control sequences when
redirected; distinguish them from the JSON receipt.

## What this establishes

A real external callback stays behind a barrier while public dispatch and
`wait_for_job` return running snapshots. Both observations have recorded positive
durations before the callback completes. The public terminal snapshot then gates
the final observation. Its randomized real output must reach both the result
and owned file. Earlier duration and running state remain unchanged. Controls
retain the observed Job but allocate no wrapper; the callback reaches the real
owning Agent and application through its parents. A schema refusal does not
record admission or duration and does not invoke the callback.

The test has a 30-second overall observation deadline, not a tiny timing-accuracy
threshold. It releases its external barrier before owner destruction even on
failure. It does not prove forced termination of an arbitrary uncooperative
handler. It does not measure worker runtime, exercise process PTYs, test provider
invocation or crash recovery, or award the separate consumer/topology score.

The callback returns text. These checks therefore do **not** certify completed
structured-result metadata, first-terminal delivery semantics, or a delayed
built-in work-product offer. The integrated MCP structured-result exercise is
deferred to Chapter11, not a Chapter4 admission blocker. Chapter26 retains the
specific offer checks. No production test hook is introduced by this supplement.

## Separate-process observation replay

Run the writer in a fresh workspace and require its successful process exit before
starting the reader in that same workspace:

```sh
cd "$REPLAY_WORK"
"$DOTNET" "$BUILD/bin/Chapter4Observations/debug/Chapter4Observations.dll" log-write
"$DOTNET" "$BUILD/bin/Chapter4Observations/debug/Chapter4Observations.dll" log-read
```

The writer performs all 23 real dispatch checks plus four persistence checks
(27 total). It starts `History.StartLog` before dispatch, saves a grader-owned
manifest of the actual running, running-control, terminal-control, and refused
observations, and compares `History.Dump` with the live JSON-lines log before
destruction. Owner destruction completes before the writer reports success.
The manifest uses independent plain records rather than the candidate's event
codec as its expected-value representation.

The separate reader requires all 21 named checks. It does not seed missing files
or invoke tools. It first compares raw canonical facts with the manifest, then
calls public `History.Load` and checks the actual published events, then checks
`Dump` again. Exact timestamp offsets, duration ticks (including the refused
observation's unknown duration), all Job snapshot fields, error flags, kinds,
call IDs, and result text must survive both writer files. The reader registers
no tools. These are clean process-separated JSON-lines persistence checks, not
crash recovery or restored work execution.

Direct `ToolObserved` audit facts in later runtimes deliberately do not enter
model dialogue. This consumer checks canonical replay rather than requiring
such facts in retained Context. Model-driven `ToolReturned` reduction and the
external provider-consumer exercise remain separate obligations. The text
callback still cannot certify structured completion metadata.

## Receiver-bound execution

The Chapter3 public `Tool.ExecuteAsync(ToolCallPart, CancellationToken)` member
must retain its receiver binding after the Job refactor. In a fresh workspace run:

```sh
"$DOTNET" "$BUILD/bin/Chapter4Observations/debug/Chapter4Observations.dll" receiver
```

Require exit zero, `passed: true`, and all ten named checks. Two harmless custom
callbacks distinguish execution from a plausible result. A call naming the other
Tool returns a correlated error with no Job metadata, callback effects, new Job,
or History dispatch facts. A matching receiver executes its own callback;
`Agent.DispatchAsync` still selects the other Tool by name. Comparison is ordinal.
The check uses existing public APIs, no model requests or candidate test assembly.
It has a 20-second deadline and destroys its owner before reporting success.
Default and replay modes retain their existing counts. This adds no historical
chapter points and does not test detached receivers or arbitrary concurrent
registry replacement.

## Cleanup despite diagnostic failure

Run a separate fresh workspace with the same consumer build:

```sh
"$DOTNET" "$BUILD/bin/Chapter4Observations/debug/Chapter4Observations.dll" cleanup
```

Require exit zero, `passed: true`, and all ten named checks. The existing public
composition root accepts a borrowed TextWriter. The test appends an explicitly
synthetic ResponseEnded with unknown usage, checks that the early diagnostic path
actually writes, and admits a real custom callback. The writer then throws during
Agent destruction. Cancellation must reach the callback despite that exception;
destruction must remain pending while its cooperative cleanup is held, and finish
only after the grader releases it. The recording error remains visible, owned
Jobs detach, the unrelated Agent survives, and the borrowed writer is not disposed.

There are no candidate-test calls, private fields, reflection, model requests or
real processes. A separate grader-owned emergency cancellation source is used
only during final teardown so a rejected candidate cannot strand the callback;
it cannot satisfy the assertion on the Job's own cancellation token. A 20-second
observation deadline bounds the test. This proves the exercised cooperative
failure path, not crash recovery, all possible cleanup failures or forced stopping
of an uncooperative handler. Default/replay/receiver counts and historical scores
are unchanged. Do not suppress diagnostic exceptions merely to make destruction
look successful.
