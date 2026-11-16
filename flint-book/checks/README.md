# Verify the student skill without a provider

Run from the Ensemble repository root. Prerequisites: an existing .NET 10 SDK,
a compatible built Ensemble.dll, and the local framework/cache needed for offline
restore. No provider credentials, downloads, child launch or host restart is used.
The assembly must be outside any untrusted student write authority.

Set these three absolute paths for your installation:

```sh
SDK=/absolute/path/to/dotnet
ASSEMBLY=/trusted/build/Ensemble.dll
CACHE=/absolute/path/to/existing/nuget/packages
REPO="$PWD"
CHECK=$(mktemp -d /private/tmp/flint-skill-check.XXXXXX)
mkdir -p "$CHECK/home" "$CHECK/tmp" "$CHECK/cli"
printf '<configuration><packageSources><clear /></packageSources></configuration>\n' > "$CHECK/NuGet.Config"
export HOME="$CHECK/home" TMPDIR="$CHECK/tmp" DOTNET_CLI_HOME="$CHECK/cli"
export NUGET_PACKAGES="$CACHE" DOTNET_SKIP_FIRST_TIME_EXPERIENCE=1 DOTNET_CLI_TELEMETRY_OPTOUT=1
"$SDK" build flint-book/checks/SkillSurface/SkillSurface.csproj \
  --artifacts-path "$CHECK/artifacts" --configfile "$CHECK/NuGet.Config" \
  -p:NuGetAudit=false -p:EnsembleAssembly="$ASSEMBLY"
# Continue only if the build succeeded.
"$SDK" "$CHECK/artifacts/bin/SkillSurface/debug/SkillSurface.dll" \
  "$REPO" "$REPO/flint-book/skills/flint-student"
```

Require exit 0 and `PASS: real loader, exact tools, coding activation, no provider request`.
The second program argument can name an independently supplied candidate directory
containing SKILL.md. Its declared primary name must be `flint-student`.

The verifier assembles a disposable catalogue from the candidate and explicit
repository dependencies, constructs a confined Agent, checks exact enabled tools,
loads `ensemble-coding` through real tool dispatch, checks unchanged authority,
and destroys the Agent before deleting its fixture. It does not run the model.
The build directory remains for inspection; remove only your recorded `$CHECK`
directory when done. These exports affect the current shell; use a disposable
shell if you need to preserve its previous environment.

## Recorded validation

The initial check compiled with zero warnings/errors against the previously
verified Ensemble build used by the first trial. It initially failed because the
fixture guessed `agent/skills/coding`; the packaged canonical dependency is
`agent/skills/dev-driven-specs/coding`. Correcting that path yielded a pass.
Disposable candidate mutations adding `spawn_sub_agent` and removing `coding`
were each rejected for the intended reason. The candidate itself was not changed
by those mutations. These checks prove loader behavior, not student performance.
