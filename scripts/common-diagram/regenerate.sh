#!/bin/sh
# Regenerate source-only diagrams with existing tools; no application build or downloads.
set -eu
TOOL=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO=$(CDPATH= cd -- "$TOOL/../.." && pwd)
DOTNET=${1:-dotnet}
OUT="$REPO/docs/common-diagrams"
mkdir -p "$TOOL/.build/home" "$TOOL/.build/tmp" "$OUT"
export DOTNET_CLI_HOME="$TOOL/.build/home"
export TMPDIR="$TOOL/.build/tmp"
export DOTNET_CLI_TELEMETRY_OPTOUT=1
export DOTNET_SKIP_FIRST_TIME_EXPERIENCE=1
"$DOTNET" build "$TOOL/extractor/Extractor.csproj" --configfile "$TOOL/extractor/NuGet.Config" --artifacts-path "$TOOL/.build/artifacts" -p:NuGetAudit=false -p:UseSharedCompilation=false --nologo
"$DOTNET" "$TOOL/.build/artifacts/bin/Extractor/debug/Extractor.dll" "$REPO/agent/Common" "$OUT/inventory.json"
python3 "$TOOL/generate.py" "$REPO"
