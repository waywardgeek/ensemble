# Chapter2 timing-value supplement

This grader-owned outside-assembly consumer checks the revised Chapter2 data
contract using only Event, Entry/Context, History.Reduce and JsonCodec. It uses no
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
Provider timing exclusion, real log restart and the complete permitted teaching
bundle still need early-chapter-specific acceptance work before student launch.
