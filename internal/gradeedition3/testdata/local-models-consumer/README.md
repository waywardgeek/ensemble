# Local Models public consumer (in development)

This outside-assembly .NET 10 consumer references only a candidate Ensemble.dll.
It uses public Agent, Model, Tool and History APIs with grader-owned loopback HTTP,
random payloads and real buffered/SSE replies. It never references candidate tests,
reflects private fields, accesses an installed credential store, or calls ET.

Current coverage is 43 named checks across buffered tools, streamed tools and
ordinary tool-shaped prose. Checks cover exact local routing, explicit no-auth
headers, configured capabilities, real callback/result correlation, recording
before the next request, partial-stream non-execution, offering provenance, serving
context and cached-input accounting. The final fixture answer is released only
after checking the actual result request; this is protocol evidence, not a claim
that a real model understood the result. Public registration grants one harmless
custom tool. This does not certify every packaged skill's permission checks.

This component awards no chapter points. The named CLI runner now supplies an
outer process deadline and strict receipt validation. Lifecycle, configuration
loading, restoration, unavailable endpoint and unknown-price coverage remain incomplete.
No live-model usefulness or installed-GUI acceptance follows from these checks.

## Public runner

From the repository root, with an existing candidate assembly and .NET 10 SDK:

```sh
go run ./cmd/grade-edition3 --scope local-models -executable-only -json \
  -dotnet "$DOTNET" "$CANDIDATE"
```

The compiled grader returns 3 for passing current probes with incomplete chapter
coverage, 1 for a failed probe, and 2 for setup or API-build incompatibility.
`go run` may report the program's nonzero status through its own exit wrapper;
use a built grader when consuming exit codes directly. No chapter points are
awarded. Do not combine this scope with chapter, browser or media flags.

The macOS/Linux runner embeds this consumer, creates fresh home/work directories,
clears inherited credentials, and builds offline with cleared package sources.
Setup has a two-minute limit; execution has a separate 75-second limit. Both
stdout and stderr are capped at 64 KiB. The runner cancels its owned process group,
waits for the direct child and bounds inherited output-pipe waits. It does not
provide a network sandbox or containment of deliberately detached descendants.
Success requires exactly the 43 named checks once each and a null failure field;
missing, duplicated, unknown or trailing receipt data refuse. Temporary files are
removed after execution; capture the JSON report for retained evidence.

## Build and run against an existing candidate

Supply absolute paths to an existing .NET 10 SDK, the candidate runtime directory,
this consumer directory, and a new evidence directory. No downloads are required;
NuGet.Config clears package sources and the consumer has no package references.
The candidate assembly must implement the public Chapter27 Model contracts.

```sh
mkdir -p "$EVIDENCE/home" "$EVIDENCE/tmp" "$EVIDENCE/cli" "$EVIDENCE/work"
env -i PATH=/usr/bin:/bin HOME="$EVIDENCE/home" TMPDIR="$EVIDENCE/tmp" \
  DOTNET_CLI_HOME="$EVIDENCE/cli" DOTNET_SKIP_FIRST_TIME_EXPERIENCE=1 \
  DOTNET_CLI_TELEMETRY_OPTOUT=1 DOTNET_GENERATE_ASPNET_CERTIFICATE=false \
  "$DOTNET" build "$CONSUMER/LocalModelsConsumer.csproj" \
  --artifacts-path "$EVIDENCE/artifacts" --configfile "$CONSUMER/NuGet.Config" \
  -p:CandidateDirectory="$CANDIDATE" -p:NuGetAudit=false --tl:off -v:q
cd "$EVIDENCE/work"
env -i PATH=/usr/bin:/bin HOME="$EVIDENCE/home" TMPDIR="$EVIDENCE/tmp" \
  DOTNET_CLI_HOME="$EVIDENCE/cli" DOTNET_SKIP_FIRST_TIME_EXPERIENCE=1 \
  DOTNET_CLI_TELEMETRY_OPTOUT=1 DOTNET_GENERATE_ASPNET_CERTIFICATE=false \
  "$DOTNET" "$EVIDENCE/artifacts/bin/LocalModelsConsumer/debug/LocalModelsConsumer.dll"
```

The executable emits one JSON receipt after joining its server/turn tasks and
root destruction. Exit0 means these named checks passed; exit1 means failure.
A compile failure is setup/API incompatibility, not a behavioral negative control.
Each scenario has a 20-second cooperative deadline. A broken candidate may ignore
cancellation; direct callers still need an outer process-group deadline. The
public runner supplies that outer bound. A cleared environment is not a network
sandbox. The loopback port is reserved then rebound;
a competing bind causes a setup failure, not a silently chosen remote endpoint.

Calibration: the current candidate passes43 checks. An isolated, successfully
compiled candidate that emits a synthetic bearer for explicit no-auth fails at
buffered-no-borrowed-credentials-0. That proves this wire oracle's sensitivity,
not mutation coverage for every assertion. No candidate or installed source is
modified by the consumer.
