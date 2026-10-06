#!/usr/bin/env bash
# Mutation audit for the chapter 21 grader.
#
# A grader that awards 100/100 has proved nothing. It has proved that one tree
# passes, which is also true of a grader that returns 100 unconditionally. The
# question worth answering is whether each check FAILS when the specific defect
# it exists to catch is present.
#
# So: break one thing, run the grader, and record which checks noticed. A check
# whose mutant it does not kill is decoration. A mutant that trips several
# checks means the checks overlap, which is worth knowing but is not a failure.
#
# This is a BATCH check. It builds and runs the agent several times over. It is
# not on the interactive path and must never be put there.
#
# Usage: scripts/ch21-mutants.sh [mutant-number]
#        With no argument, runs all four.
#
# ----------------------------------------------------------------------------
# ONE CHECK HAS NO MUTANT, DELIBERATELY: tools-gated-by-skill.
#
# A valid mutant deletes exactly one behaviour. There is no behaviour to delete
# here: MCP servers are connected in exactly one place, the load_skill tool
# handler, so nothing connects a server before a skill is loaded and the web
# tools cannot appear in the first request. Breaking the check would mean ADDING
# a startup-connect path, which is a feature, not a mutation.
#
# That is not an argument for dropping the check. It is reachable for a STUDENT,
# and the chapter brief proposed precisely the design that fails it — "the tools
# are always available, the servers start when the agent starts". The check is
# what distinguishes that design from the one the architecture supports, so it
# earns its fifteen points against the submissions it will actually see rather
# than against a mutant of the reference.
# ----------------------------------------------------------------------------

set -uo pipefail
cd "$(dirname "$0")/.."
ROOT="$PWD"
OUT="${TMPDIR:-/tmp}/ch21-mutants"
mkdir -p "$OUT"

# The guard is scoped to the subtree this script actually mutates. It reverts
# with 'git checkout -- agent/...', so uncommitted work under agent/ would be
# destroyed; work anywhere else is none of its business.
if ! git diff --quiet -- agent/ || ! git diff --cached --quiet -- agent/; then
	echo "REFUSING: there are uncommitted changes under agent/."
	echo "This script mutates tracked files there and reverts with 'git checkout --'."
	echo "Uncommitted work would be destroyed. Commit or stash first."
	exit 1
fi

GRADER="$OUT/grade"
echo "building grader..."
go build -o "$GRADER" ./cmd/grade || exit 1

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
	# The verify command must exit 0 only if the edit actually landed. Skipping
	# that step is how a botched edit gets recorded as a surviving mutant, which
	# reads exactly like an insensitive check and sends you off to fix the wrong
	# thing.
	if ! bash -c "$verify"; then
		echo "  MUTATION DID NOT LAND - skipping (this is a script bug, not a result)"
		git checkout -- $files
		FAILED=$((FAILED + 1))
		return
	fi
	echo "  mutation landed."

	# A mutant that fails to build proves only that the compiler works.
	if ! (cd agent && go build ./... >/dev/null 2>&1); then
		echo "  MUTANT DOES NOT COMPILE - invalid mutant (this is a script bug)"
		git checkout -- $files
		FAILED=$((FAILED + 1))
		return
	fi
	echo "  mutant compiles."

	"$GRADER" -ch 21 -json ./agent >"$OUT/m$n.json" 2>"$OUT/m$n.err"
	git checkout -- $files

	# Match on check IDs from the JSON report, not on the human-readable
	# titles. The text report prints the title only, so a grep against it
	# silently matches nothing and every mutant reads as a survivor.
	local killed
	killed=$(python3 - "$OUT/m$n.json" <<-'PY'
		import json, sys
		def walk(x, out):
		    if isinstance(x, dict):
		        if "id" in x and "passed" in x:
		            if not x["passed"]:
		                out.append(x["id"])
		        for v in x.values():
		            walk(v, out)
		    elif isinstance(x, list):
		        for v in x:
		            walk(v, out)
		try:
		    doc = json.load(open(sys.argv[1]))
		except Exception as e:
		    print("UNPARSEABLE", e)
		    sys.exit(0)
		out = []
		walk(doc, out)
		print(" ".join(sorted(set(out))))
	PY
	)
	killed=$(echo $killed)

	echo "  checks that failed: ${killed:-<none>}"
	if [ "$killed" = "$expect" ]; then
		echo "  RESULT: KILLED EXACTLY THE EXPECTED CHECKS"
		PASSED=$((PASSED + 1))
	else
		echo "  RESULT: MISMATCH - expected '$expect'"
		FAILED=$((FAILED + 1))
	fi
}

ALL="1 2 3 4"
WANT="${1:-$ALL}"

for n in $WANT; do
case $n in

# The host no longer knows how to build a URL transport, so the skill's server
# is never reached. Everything downstream of having a web tool fails, which is
# the honest consequence rather than an overlap to file down.
1) run_mutant 1 \
	"host cannot construct a URL transport, so the hosted MCP server is unreachable" \
	"fetch-returns-planted-token fetched-content-is-a-tool-result search-dispatches tool-error-is-reported transport-error-is-reported url-transport-connects" \
	"agent/cmd/main.go" \
	"perl -pi -e 's/case \"url\":/case \"url-MUTANT1\":/' agent/cmd/main.go" \
	"grep -q 'url-MUTANT1' agent/cmd/main.go"
;;

# Tool output is truncated on the way back to the model. The page arrived, the
# call succeeded, and what the model is shown is a stump. This is the failure
# that a check asserting only "a tool result exists" cannot see.
2) run_mutant 2 \
	"tool results are truncated to 30 bytes before reaching the model" \
	"fetch-returns-planted-token fetched-content-is-a-tool-result search-dispatches" \
	"agent/internal/mcp/bridge.go" \
	"perl -pi -e 's/\treturn strings\.Join\(parts, \"\\\\n\"\)/\tout := strings.Join(parts, \"\\\\n\"); if len(out) > 30 { out = out[:30] }; return out/' agent/internal/mcp/bridge.go" \
	"grep -q 'out\[:30\]' agent/internal/mcp/bridge.go"
;;

# The server said the call failed; the agent passes the body through as an
# ordinary success. The words "404 Not Found" still reach the model, attached
# to a result that claims it worked -- which is why the check asserts the
# is_error flag and not the text.
3) run_mutant 3 \
	"the server's isError flag is dropped, so a failed tool reads as a success" \
	"tool-error-is-reported" \
	"agent/internal/mcp/bridge.go" \
	"perl -pi -e 's/if tr\.IsError \{/if false \{ \/\/ MUTANT3/' agent/internal/mcp/bridge.go" \
	"grep -q 'MUTANT3' agent/internal/mcp/bridge.go"
;;

# An HTTP-level refusal keeps its status and loses its body. "500 Internal
# Server Error" is true and useless; the reason the request was refused -- the
# quota, the expired key -- lived in the bytes that were thrown away.
4) run_mutant 4 \
	"the response body is discarded on a non-2xx, so the refusal loses its reason" \
	"transport-error-is-reported" \
	"agent/internal/mcp/http.go" \
	"perl -pi -e 's/resp\.Status, snippet\(body\)\)/resp.Status, \"\") \/\/ MUTANT4/' agent/internal/mcp/http.go" \
	"grep -q 'MUTANT4' agent/internal/mcp/http.go"
;;

esac
done

echo
echo "=============================================================="
echo "mutants killing exactly their checks: $PASSED"
echo "mismatched or invalid:                $FAILED"
echo "transcripts in $OUT"
echo "=============================================================="
[ "$FAILED" -eq 0 ]
