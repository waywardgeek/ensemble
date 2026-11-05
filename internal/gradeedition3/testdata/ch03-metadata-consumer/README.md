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
