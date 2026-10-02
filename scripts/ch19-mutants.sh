#!/usr/bin/env bash
# Mutation audit for the chapter 19 grader.
#
# A grader that awards 100/100 has proved nothing. It has proved that one tree
# passes, which is also true of a grader that returns 100 unconditionally. The
# question worth answering is whether each check FAILS when the specific defect
# it exists to catch is present, and whether it stays quiet otherwise.
#
# So: break one thing, run the grader, and record which checks noticed. A check
# whose mutant it does not kill is decoration. A mutant that trips several
# checks means the checks overlap, which is worth knowing but is not a failure.
#
# This is a BATCH check. It builds and runs the agent seven times over. It is
# not on the interactive path and must never be put there.
#
# Usage: scripts/ch19-mutants.sh [mutant-number]
#        With no argument, runs all seven.

set -uo pipefail
cd "$(dirname "$0")/.."
ROOT="$PWD"
OUT="${TMPDIR:-/tmp}/ch19-mutants"
mkdir -p "$OUT"

# The guard is scoped to the subtree this script actually mutates. It reverts
# with 'git checkout -- agent/...', so uncommitted work under agent/ would be
# destroyed; work anywhere else is none of its business. A whole-tree guard
# would refuse to run whenever the author had an unsaved edit in book/, which
# is both common and harmless.
if ! git diff --quiet -- agent/ || ! git diff --cached --quiet -- agent/; then
	echo "REFUSING: there are uncommitted changes under agent/."
	echo "This script mutates tracked files there and reverts with 'git checkout --'."
	echo "Uncommitted work would be destroyed. Commit or stash first."
	exit 1
fi

GRADER="$OUT/grade"
echo "building grader..."
go build -o "$GRADER" ./cmd/grade || exit 1

# Each mutant deletes exactly ONE behaviour, and must still COMPILE: a mutant
# that fails to build proves only that the compiler works.
#
# The verify command must exit 0 only if the edit actually landed. Skipping
# that step is how a botched edit gets recorded as a surviving mutant, which
# reads exactly like an insensitive check and sends you off to fix the wrong
# thing.

PASSED=0
FAILED=0

run_mutant() {
	local n="$1" name="$2" expect="$3" files="$4" apply="$5" verify="$6"
	echo
	echo "=============================================================="
	echo "mutant $n: $name"
	echo "   expected to kill: $expect"
	echo "=============================================================="

	bash -c "$apply"
	if ! bash -c "$verify"; then
		echo "  MUTATION DID NOT LAND - skipping (this is a script bug, not a result)"
		git checkout -- $files
		FAILED=$((FAILED + 1))
		return
	fi
	echo "  mutation landed."

	if ! (cd agent && go build ./... >/dev/null 2>&1); then
		echo "  MUTANT DOES NOT COMPILE - invalid mutant (this is a script bug)"
		git checkout -- $files
		FAILED=$((FAILED + 1))
		return
	fi
	echo "  mutant compiles."

	"$GRADER" -ch 19 ./agent >"$OUT/m$n.txt" 2>&1
	git checkout -- $files

	# Go prints "--- FAIL:" with a space. The grader prints "[FAIL]".
	local killed
	killed=$(grep -o '^\[FAIL\] [a-z-]*' "$OUT/m$n.txt" | sed 's/^\[FAIL\] //' | sort | tr '\n' ' ')
	killed=$(echo $killed)

	echo "  checks that failed: ${killed:-<none>}"
	if [ "$killed" = "$expect" ]; then
		echo "  RESULT: KILLED EXACTLY THE EXPECTED CHECK"
		PASSED=$((PASSED + 1))
	else
		echo "  RESULT: MISMATCH - expected '$expect'"
		FAILED=$((FAILED + 1))
	fi
}

ALL="1 2 3 4 5 6 7"
WANT="${1:-$ALL}"

for n in $WANT; do
case $n in

1) run_mutant 1 \
	"engine ignores the credential provider and always uses the configured key" \
	"credential-provider" \
	"agent/internal/llm/engine.go" \
	"perl -pi -e 's/if e\.Creds == nil \{/if e.Creds == nil || true {/' agent/internal/llm/engine.go" \
	"grep -q 'e.Creds == nil || true' agent/internal/llm/engine.go"
;;

2) run_mutant 2 \
	"PKCE downgraded from S256 to plain" \
	"oauth-flow" \
	"agent/internal/oauth/pkce.go" \
	"perl -pi -e 's/const codeChallengeMethodS256 = \"S256\"/const codeChallengeMethodS256 = \"plain\"/' agent/internal/oauth/pkce.go" \
	"grep -q 'codeChallengeMethodS256 = \"plain\"' agent/internal/oauth/pkce.go"
;;

3) run_mutant 3 \
	"credentials never report needing a refresh, so the agent waits to be told" \
	"token-refresh" \
	"agent/internal/oauth/store.go" \
	"perl -pi -e 's/return !now\.Add\(skew\)\.Before\(c\.ExpiresAt\)/return false/' agent/internal/oauth/store.go" \
	"! grep -q 'now.Add(skew).Before' agent/internal/oauth/store.go"
;;

4) run_mutant 4 \
	"store:true, asking the vendor to retain the conversation" \
	"responses-format" \
	"agent/internal/llm/openai_responses.go" \
	"perl -pi -e 's/Store:\s+false,/Store: true,/' agent/internal/llm/openai_responses.go" \
	"grep -q 'Store: true,' agent/internal/llm/openai_responses.go"
;;

5) run_mutant 5 \
	"explicit cache mode declared but no breakpoint ever attached" \
	"cache-breakpoints" \
	"agent/internal/llm/openai_responses.go" \
	"perl -pi -e 's/blocks\[len\(blocks\)-1\]\.CacheBreakpoint = &respBreakpoint\{Mode: \"explicit\"\}/_ = blocks/' agent/internal/llm/openai_responses.go" \
	"grep -q '_ = blocks' agent/internal/llm/openai_responses.go"
;;

6) run_mutant 6 \
	"reasoning summary never requested, so none is streamed" \
	"reasoning-summaries" \
	"agent/internal/llm/openai_responses.go" \
	"perl -pi -e 's/reasoning\.Summary = \"auto\"/reasoning.Summary = \"\"/' agent/internal/llm/openai_responses.go" \
	"grep -q 'reasoning.Summary = \"\"' agent/internal/llm/openai_responses.go"
;;

7) run_mutant 7 \
	"mid-stream failure ignored, so a 200 that produced nothing reads as success" \
	"billing-mode" \
	"agent/internal/llm/openai_responses.go" \
	"perl -pi -e 's/case \"response\.failed\", \"response\.incomplete\":/case \"response.failed.NEVER\", \"response.incomplete.NEVER\":/' agent/internal/llm/openai_responses.go" \
	"grep -q 'response.failed.NEVER' agent/internal/llm/openai_responses.go"
;;

esac
done

echo
echo "=============================================================="
echo "mutants killing exactly their check: $PASSED"
echo "mismatched or invalid:               $FAILED"
echo "transcripts in $OUT"
echo "=============================================================="
[ "$FAILED" -eq 0 ]
