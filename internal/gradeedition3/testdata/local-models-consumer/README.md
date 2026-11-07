# Local Models public consumer (in development)

This outside-assembly .NET 10 consumer references only a candidate Ensemble.dll.
It uses public Agent, Model, Tool and History APIs with grader-owned loopback HTTP,
random payloads and real buffered/SSE replies. It never references candidate tests,
reflects private fields, accesses an installed credential store, or calls ET.

Current coverage is 136 named checks across buffered tools, streamed tools,
ordinary tool-shaped prose, configured child restoration and executing-tool
shutdown, including separate writer/reader processes. Checks cover exact local
routing, explicit no-auth headers, configured capabilities, real callback/result
correlation, recording before the next request, partial-stream non-execution,
offering provenance, serving
context and cached-input accounting. The final fixture answer is released only
after checking the actual result request; this is protocol evidence, not a claim
that a real model understood the result. Public registration grants one harmless
custom tool. This does not certify every packaged skill's permission checks.

This component awards no chapter points. The named CLI runner now supplies an
outer process deadline and strict receipt validation. Lifecycle, configuration
loading and restoration now have bounded public checks; broader lifecycle/refusal
and unavailable-endpoint coverage remain incomplete.
No live-model usefulness or installed-GUI acceptance follows from these checks.

A separate buffered response omits usage entirely: working-context input and
Engine meter input must remain null, with unknown dollar cost. This is a fresh
Agent case, not a known-to-missing transition or streamed missing-usage check.
A compiled zero-substitution candidate fails missing-usage-input-remains-unknown.

Real conversation meters preserve known token counts while no-auth dollar cost
remains unknown. Configuration checks use the public document loader and selector,
assert late-error rollback, save a paused child, destroy the root, and restore
through fresh owners from actual files. A same-wire offering on a different
endpoint must not substitute for the missing identity. Refusal preserves the
registration; Model destruction leaves the Agent and conversation alive but
refuses inference configuration. This original probe is fresh-root restoration
in one process. The restoration phases make no
provider request. An isolated compiled resolver mutant that picks the first
endpoint fails child-missing-route-no-wire-name-fallback.

A separate restart probe writes a real paused child and random expectation
manifest, destroys its owners, and exits before a new reader process begins.
The reader loads only those saved files, reconstructs the configured graph, and
checks exact offering/funding, retained conversation and no replayed request.
The manifest is an expectation, never a source of restored dialogue. Removing
saved child History from a disposable copy fails restart-read-paused-conversation;
an empty reader directory refuses without creating seed files. This is clean
process restart, not crash recovery or a full GUI-host restart.

The writer also separates the child's original launch offering, its current saved
Agent preference, and the application's remembered funding default. After orderly
shutdown it deliberately restores its captured original launch record, without
altering the saved preference. The reader omits that old launch endpoint entirely.
Both parent and child must restore the saved offering, not the different usable
global default. This checks preference precedence across a clean process boundary;
it does not certify crash recovery or every peer/fleet restoration case.

Fifteen additional public preference checks use real temporary settings files.
They cover ambiguity, explicit saved references, independent Agent/default state,
exact remembered binding, unrelated global/key preservation, exclusive writer-lock
refusal, sibling Agent settings, selection-save refusal, missing funding without
fallback, and fresh-root restoration. A usable global decoy must not replace an
unavailable saved Agent choice. These operations initiate no inference; synthetic
key preservation is not a credential transport or live entitlement test. An
isolated compiled wrong-binding resolver fails exactly
preference-resolves-remembered-binding after the preceding preference guards pass.

The lifecycle probe delivers a real tool call over owned HTTP. Its callback
acknowledges candidate cancellation and holds cleanup behind a fixture gate.
Agent destruction must stay pending until release, then remove Model membership
without destroying a sibling. The fixture retains and joins the actual callback
Task even after failure; emergency cancellation is used only during cleanup and
cannot satisfy the candidate-token assertion. A compiled no-join candidate fails
lifecycle-destruction-joins-held-cleanup. This proves cooperative shutdown, not
forced termination of arbitrary managed callbacks.

Two additional real-SSE cases hold incomplete tool arguments, then either submit
public InterruptNow or return a provider error. Neither may authorize a call or
effect. A fresh turn on the same Agent and offering must finish with qualified
provenance and known usage, without orphan tool messages in its actual request.
The provider error must be a controlled canonical failure, not successful empty
completion. A compiled false-success parser fails provider-error-controlled-failure.
This does not certify every disconnect, malformed frame or interruption mode.

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
Setup has a two-minute limit; all three executions share a separate 75-second
limit. Both stdout and stderr are capped at 64 KiB. The runner cancels its owned process group,
waits for the direct child and bounds inherited output-pipe waits. It does not
provide a network sandbox or containment of deliberately detached descendants.
Success requires the 126 main checks, 4 writer checks and 6 reader checks
in their respective receipts, each with a null failure field;
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
Network and lifecycle scenarios have 20-second cooperative deadlines. A broken
candidate may ignore cancellation; direct callers still need an outer process-group deadline. The
public runner supplies that outer bound. A cleared environment is not a network
sandbox. The loopback port is reserved then rebound;
a competing bind causes a setup failure, not a silently chosen remote endpoint.

The no-argument invocation above runs the 126 main checks. To exercise restart,
invoke the same executable with restart-write, wait for successful exit, then
invoke restart-read from the same work directory and cleared environment. Each
phase emits its own receipt (4 and 6 checks); the public runner enforces this order.

Calibration: the current candidate passes all 136 checks. An isolated, successfully
compiled candidate that emits a synthetic bearer for explicit no-auth fails at
buffered-no-borrowed-credentials-0. That proves this wire oracle's sensitivity,
not mutation coverage for every assertion. No candidate or installed source is
modified by the consumer.
