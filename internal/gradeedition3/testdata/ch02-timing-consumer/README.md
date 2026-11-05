# Chapter2 timing-value supplement

This grader-owned outside-assembly consumer checks the revised Chapter2 data
contract using Event, Entry/Context, History, JsonCodec and Engine.Render. It uses no
candidate test assembly, reflection, Jobs, GUI, provider requests or credentials.
It does not replace the existing language-neutral `grade-edition2 -ch 2` exercise
or award chapter points. It is a focused supplement, not a student acceptance
receipt by itself.

Required candidate API is the immutable timing shape printed in Chapter2:
nullable Event.Time and Entry.Time; nullable nonnegative ObservationDuration;
Event.Numbered; public pure History.Reduce; JsonCodec.ReadEvent/WriteEvent/Text.
Known time retains its exact offset and ticks. Zero duration differs from unknown.
Numbering changes the fact identity/time while preserving duration; reduction and
redaction preserve original timing, while a synthesized summary has the new
summary fact's time and no inherited observation duration. Invalid negative or
non-result event duration refuses.

With the existing .NET10 SDK, run in an owned isolated home/temp/CLI environment.
Use absolute paths and a separate output directory; the included NuGet.Config
clears package sources. This project has no package dependencies. These are
command templates, not permission to install or fetch dependencies:

```sh
/absolute/dotnet build /absolute/grader/Chapter2Timing.csproj \
  -p:CandidateAssembly=/absolute/candidate/Ensemble.dll \
  --configfile /absolute/grader/NuGet.Config \
  --artifacts-path /absolute/owned/artifacts --tl:off -v:q
/absolute/dotnet /absolute/owned/artifacts/bin/Chapter2Timing/debug/Chapter2Timing.dll
```

Require successful compilation, exit0 and JSON `passed:true` with contract
`chapter2-timing-values-v1`. The20 named checks cover known/zero/unknown values,
codec and reduction, numbering, refusals, three redaction levels and summary
replacement. A failure exits1 with the completed checks and named failure. A build
failure is an interface/setup failure, not a successful sensitivity test.

The consumer has passed against the current candidate core. An isolated core
that drops retained tool-result timing compiled and failed `known-reduction`
after its value and codec checks passed. This is not yet a fresh Chapter2 student
result. The frozen historical Chapter2 endpoint predates the added timing fields;
do not modify that archive or claim it passed this supplement. Separate public
CLI tests retain responsibility for the original three-vendor and log workflows.
The complete permitted teaching bundle still needs audit before student launch.


## Log-process and rendering supplement

The same executable accepts three explicit modes. Run these from one fresh,
writable grader workspace, using the isolated environment described above. Wait
for each command to exit successfully before launching the next. The reader
requires the writer's files; it never creates replacement input.

```sh
/absolute/dotnet /absolute/owned/artifacts/bin/Chapter2Timing/debug/Chapter2Timing.dll log-write
/absolute/dotnet /absolute/owned/artifacts/bin/Chapter2Timing/debug/Chapter2Timing.dll log-read
/absolute/dotnet /absolute/owned/artifacts/bin/Chapter2Timing/debug/Chapter2Timing.dll render
```

Each command must exit0 and return `passed:true`, contract
`chapter2-timing-boundaries-v1`, and its exact mode. Writer has5 named checks,
reader8, and render10. Retain every receipt, not only the last command's exit.
The writer refuses existing fixture files. It uses public StartLog/Append/Dump,
records the returned UTC timestamp in an independent expected-values manifest,
and destroys the owner before returning. Dump/log equality is checked before
shutdown, since shutdown may append a lifecycle fact. The reader checks both
actual files after the writer exits, including positive and zero durations and
exact recorded timestamps. A headerless legacy fixture separately exercises
unknown time and duration through Load and Dump. This is clean log-process
continuity, not crash recovery or the later snapshot API.

The render mode uses the three Chapter2 default routes: Messages,
ChatCompletions and GenerateContent. It checks positive random payload presence
and full JSON equality between otherwise identical timed and untimed Contexts.
No surface override or later Responses route is required; no HTTP request is
made. Synthetic unused keys and a loopback-only endpoint are configuration, not
access to real credentials. Run only in the disposable workspace and environment.

The new reader rejects the previously compiled reducer-loss candidate at
`events.jsonl-retains-12345678`. The three rendering comparisons have positive
execution evidence; no new early-route leak mutant is claimed in this increment.
These supplements still award no points and do not constitute student acceptance.
