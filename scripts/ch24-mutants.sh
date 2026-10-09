#!/usr/bin/env bash
#
# Chapter 24 mutation audit.
#
# A passing grader is evidence of nothing. It says the reference implementation
# satisfies the checks; it does not say the checks would notice if the
# implementation stopped satisfying them. This script answers the second
# question by deleting one behaviour at a time and asserting the grader reports
# exactly the checks that behaviour was protecting.
#
# Two rules learned the hard way and enforced below:
#
#   - The mutation must be proved to have landed. A sed that matched nothing
#     produces a clean run that reads exactly like an insensitive check, and
#     sends you off to fix a grader that was fine.
#   - The mutant must still compile. A mutant that fails to build proves only
#     that the compiler works, and it kills every check at once, which looks
#     like a thorough audit and is worthless.
#
# A surviving mutant usually means a broken harness or an invalid mutant, not an
# insensitive check. Suspect this script before suspecting the grader.
set -uo pipefail

cd "$(dirname "$0")/.."

if ! git diff --quiet -- agent || ! git diff --cached --quiet -- agent; then
	echo "refusing to run: agent/ has uncommitted changes."
	echo "this script mutates the tree and restores it with git checkout,"
	echo "which would destroy them."
	exit 1
fi

OUT=$(mktemp -d)
GRADER="$OUT/grade"
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

	local report="$OUT/report-$n.json"
	"$GRADER" -ch 24 -json ./agent >"$report" 2>"$OUT/stderr-$n.log"

	local killed
	killed=$(
		python3 - "$report" <<'PY'
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

	git checkout -- $files

	echo "  checks that failed: ${killed:-<none>}"

	if [ "$killed" = "$expect" ]; then
		echo "  PASS"
		PASSED=$((PASSED + 1))
	else
		echo "  FAIL - expected exactly: $expect"
		FAILED=$((FAILED + 1))
	fi
}

# --------------------------------------------------------------------------
# 1. The staple returns: the framework root imports the GUI again.
#
# Three checks die and all three dependencies are real: the headless
# consumer now links gorilla (headless-linkage), gorilla is back in the
# framework closure (mcp-without-gorilla), and the import string sits in a
# root file (root-clean). This is the chapter's central regression.
# --------------------------------------------------------------------------
run_mutant 1 "framework root imports the gui package again" \
	"headless-linkage mcp-without-gorilla root-clean" \
	"agent/agent.go" \
	"perl -pi -e 's|^package agent\$|package agent\nimport _ \"github.com/waywardgeek/ensemble/agent/gui\"|' agent/agent.go" \
	"grep -q 'ensemble/agent/gui\"' agent/agent.go"

# --------------------------------------------------------------------------
# 2. The embed is abandoned: StaticHandler always serves from disk.
#
# From an unrelated working directory there is nothing on disk to serve,
# which is exactly the pre-embed failure mode the check exists to catch.
# The reuse probe survives because it mounts components from the exported
# FS, not through StaticHandler.
# --------------------------------------------------------------------------
run_mutant 2 "StaticHandler never uses the embedded assets" \
	"assets-embedded" \
	"agent/gui/assets.go" \
	"sed -i '' 's|if dir != \"\" {|if true {|' agent/gui/assets.go" \
	"grep -q 'if true {' agent/gui/assets.go"

# --------------------------------------------------------------------------
# 3. The relay goes silent: no answer when no browser is attached.
#
# The agent-eyes leg of public-reuse dials the MCP port with no browser
# connected and expects the documented immediate error. A relay that stays
# quiet parks every headless agent on a timeout.
# --------------------------------------------------------------------------
run_mutant 3 "MCP relay stops answering when no GUI is connected" \
	"public-reuse" \
	"agent/gui/mcp_port.go" \
	"sed -i '' 's|reply(noGUIError(head.ID))|_ = head.ID|' agent/gui/mcp_port.go" \
	"grep -q '_ = head.ID' agent/gui/mcp_port.go"

# --------------------------------------------------------------------------
# 4. The settings frame is dropped: update_settings no longer matches.
#
# The scripted session switches models over the GUI websocket and the
# recorded vendor requests are the ground truth. Three replies still
# arrive, so a reply-count check alone would have passed this mutant;
# the models assertion is what kills it.
# --------------------------------------------------------------------------
run_mutant 4 "GUI websocket drops update_settings frames" \
	"wire-compat" \
	"agent/gui/handler.go" \
	"sed -i '' 's|case \"update_settings\":|case \"update_settings_x\":|' agent/gui/handler.go" \
	"grep -q 'update_settings_x' agent/gui/handler.go"

# --------------------------------------------------------------------------
# 5. The prompt frame is dropped: a browser prompt never reaches Send.
#
# The replay leg still works, so this isolates the inbound half of the
# consumer contract: the one hook the embedding application supplied.
# --------------------------------------------------------------------------
run_mutant 5 "GUI websocket drops prompt frames" \
	"public-reuse" \
	"agent/gui/handler.go" \
	"sed -i '' 's|case \"prompt\":|case \"promptx\":|' agent/gui/handler.go" \
	"grep -q 'case \"promptx\":' agent/gui/handler.go"

# --------------------------------------------------------------------------
# 6. A component file vanishes from the embedded FS.
#
# Three checks die and all three dependencies are real: the probe cannot
# read renderers.js from the exported FS (components-without-shell), it
# exits before serving anything (public-reuse), and the binary 404s the
# same file (assets-embedded). go:embed happily embeds the smaller tree,
# so only behaviour notices.
# --------------------------------------------------------------------------
run_mutant 6 "renderers.js deleted from the embedded assets" \
	"assets-embedded components-without-shell public-reuse" \
	"agent/gui/web/renderers.js" \
	"rm agent/gui/web/renderers.js" \
	"! test -f agent/gui/web/renderers.js"

# --------------------------------------------------------------------------
# 7. Gorilla re-enters the framework through internal/mcp.
#
# This is the exact regression the mcpws split prevents: the client
# websocket transport migrating back into the package every consumer
# links. Both linkage checks fire because both assert gorilla's absence
# from the headless closure, from opposite directions.
# --------------------------------------------------------------------------
run_mutant 7 "gorilla re-enters internal/mcp" \
	"headless-linkage mcp-without-gorilla" \
	"agent/internal/mcp/bridge.go" \
	"perl -pi -e 's|^package mcp\$|package mcp\nimport _ \"github.com/gorilla/websocket\"|' agent/internal/mcp/bridge.go" \
	"grep -q 'gorilla/websocket' agent/internal/mcp/bridge.go"

echo
echo "=============================================================="
echo "mutation audit: $PASSED passed, $FAILED failed"
echo "=============================================================="
[ "$FAILED" -eq 0 ]
