#!/usr/bin/env bash
#
# Chapter 23 mutation audit.
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

	"$GRADER" -ch 23 -json ./agent >"$OUT/m$n.json" 2>"$OUT/m$n.err"
	git checkout -- $files

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
		echo "  PASS"
		PASSED=$((PASSED + 1))
	else
		echo "  FAIL - expected exactly: $expect"
		FAILED=$((FAILED + 1))
	fi
}

# --------------------------------------------------------------------------
# 1. The userspace half: file tools stop checking containment.
#
# This mutant also kills no-host-path-leak, and that is a real dependency
# rather than a blunt mutant. Check 7 inspects the refusal message, and an
# implementation that never refuses produces no message to inspect. A check
# that cannot be decided must not pass, so it fails.
# --------------------------------------------------------------------------
run_mutant 1 "file tools no longer check containment" \
	"no-host-path-leak path-confinement" \
	"agent/sandbox/sandbox.go" \
	"sed -i '' 's|if !withinRoot(root, resolved) {|if false {|' agent/sandbox/sandbox.go" \
	"grep -q 'if false {' agent/sandbox/sandbox.go"

# --------------------------------------------------------------------------
# 2. The kernel half: the profile stops denying reads.
#
# Deleting the sandbox-exec wrapping entirely would also restore the network,
# killing two checks at once. Removing one rule from the profile leaves the
# wrapping in place and deletes exactly one behaviour.
# --------------------------------------------------------------------------
run_mutant 2 "the Seatbelt profile stops denying reads" \
	"kernel-confinement" \
	"agent/sandbox/seatbelt.go" \
	"sed -i '' '/(deny file-read\*)/d' agent/sandbox/seatbelt.go" \
	"! grep -q '(deny file-read\*)' agent/sandbox/seatbelt.go"

# --------------------------------------------------------------------------
# 3. The profile stops denying the network.
# --------------------------------------------------------------------------
run_mutant 3 "the Seatbelt profile stops denying the network" \
	"no-network" \
	"agent/sandbox/seatbelt.go" \
	"sed -i '' '/(deny network\*)/d' agent/sandbox/seatbelt.go" \
	"! grep -q '(deny network\*)' agent/sandbox/seatbelt.go"

# --------------------------------------------------------------------------
# 4. The child environment keeps its credentials.
#
# HOME is still redirected, so this deletes the credential stripping and
# nothing else.
# --------------------------------------------------------------------------
run_mutant 4 "the child environment keeps its credentials" \
	"no-credentials" \
	"agent/sandbox/seatbelt.go" \
	"sed -i '' 's|isSensitiveEnv(name)|false|' agent/sandbox/seatbelt.go" \
	"! grep -q 'if isSensitiveEnv(name)' agent/sandbox/seatbelt.go"

# --------------------------------------------------------------------------
# 5. The wall: the clamp stops narrowing.
#
# Mutants 5 and 6 are the two halves of check 5, and they must fail
# independently. If deleting the clamp also stopped the error being reported,
# the error path would be the security boundary, which is the arrangement the
# chapter argues against: one careless early return and the child is wider.
# --------------------------------------------------------------------------
run_mutant 5 "the clamp stops narrowing a widened child" \
	"child-cannot-widen" \
	"agent/agent.go" \
	"sed -i '' '/s.SafeMode = true/d' agent/agent.go" \
	"! grep -q 's.SafeMode = true' agent/agent.go"

# --------------------------------------------------------------------------
# 6. The alarm: the clamp narrows silently.
# --------------------------------------------------------------------------
run_mutant 6 "the clamp narrows but never reports it" \
	"child-cannot-widen" \
	"agent/agent.go" \
	"sed -i '' 's|if len(widened) > 0 {|if false {|' agent/agent.go" \
	"grep -q 'if false {' agent/agent.go"

# --------------------------------------------------------------------------
# 7. Safe mode stops removing the exec tools.
#
# The bug this check exists for was real: RemoveTool deleted a normalised key
# while the registry stored a literal one, so removing any tool whose name
# contained an underscore silently did nothing, and safe mode reported itself
# enabled while run_command stayed fully callable.
# --------------------------------------------------------------------------
run_mutant 7 "safe mode stops removing the exec tools" \
	"safe-mode-absence" \
	"agent/agent.go" \
	"sed -i '' '/SafeModeWithheldTools/,/}/ s|a.reg.RemoveTool(name)|_ = name|' agent/agent.go" \
	"grep -q '_ = name' agent/agent.go"

# --------------------------------------------------------------------------
# 8. The refusal names the host path.
#
# The sandbox still confines; only the message changes. This is the mutant
# that separates a sandbox from a sandbox that tells an attacker where the
# walls are.
# --------------------------------------------------------------------------
run_mutant 8 "the refusal reveals the resolved host path" \
	"no-host-path-leak" \
	"agent/sandbox/sandbox.go" \
	"sed -i '' '/if !withinRoot(root, resolved) {/,/}/ s|&EscapeError{Path: asked}|\&EscapeError{Path: resolved}|' agent/sandbox/sandbox.go" \
	"grep -q 'EscapeError{Path: resolved}' agent/sandbox/sandbox.go"

echo
echo "=============================================================="
echo "mutants killed as expected: $PASSED"
echo "mutants that did not behave: $FAILED"
echo "=============================================================="
rm -rf "$OUT"
[ "$FAILED" -eq 0 ]
