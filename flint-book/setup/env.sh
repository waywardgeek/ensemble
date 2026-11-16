# Source this recipe in every fresh command: . /absolute/project/env.sh
# The parent sets these three absolute paths when preparing a fresh workspace.
export PROJECT_DIR='@PROJECT_DIR@'
export GO_EXE='@GO_EXE@'
export DOTNET_ROOT='@DOTNET_ROOT@'
set -eu
cd "$PROJECT_DIR"
export DOTNET_EXE="$DOTNET_ROOT/dotnet"
# Lookup convenience is not permission: the launch policy must separately grant
# read-only imports for the selected SDK and Go toolchain.
export PATH="$DOTNET_ROOT:$(dirname "$GO_EXE"):/usr/bin:/bin:/usr/sbin:/sbin"
export HOME="$PROJECT_DIR/.home" TMPDIR="$PROJECT_DIR/.tmp"
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off CGO_ENABLED=0
export GOCACHE="$PROJECT_DIR/.cache/go-build" GOMODCACHE="$PROJECT_DIR/.cache/go-mod"
export DOTNET_CLI_HOME="$PROJECT_DIR/.dotnet-home" NUGET_PACKAGES="$PROJECT_DIR/.packages"
export DOTNET_SKIP_FIRST_TIME_EXPERIENCE=1 DOTNET_CLI_TELEMETRY_OPTOUT=1
export DOTNET_GENERATE_ASPNET_CERTIFICATE=false MSBUILDTERMINALLOGGER=off
mkdir -p "$HOME" "$TMPDIR" "$GOCACHE" "$GOMODCACHE" "$DOTNET_CLI_HOME" "$NUGET_PACKAGES"
