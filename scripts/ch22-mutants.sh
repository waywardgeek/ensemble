#!/usr/bin/env bash
#
# ch22 mutation audit.
#
# A grader that has never been observed to fail is evidence of nothing. Each
# mutant below deletes exactly ONE behavior the chapter teaches, and the run
# asserts the exact set of check IDs that go red. A mutant that kills more
# checks than expected is as much a defect as one that kills none: it means a
# check is coupled to something other than the property it names.
#
# Two rules the mutants obey:
#
#   1. Every mutant must COMPILE. A build failure proves the compiler works,
#      not that the grader is sensitive, so the harness treats a non-compiling
#      mutant as a script bug and refuses to score it.
#
#   2. Every mutation must be VERIFIED to have landed. A botched edit reads
#      exactly like a surviving mutant, which sends you off to strengthen a
#      check that was never weak.
#
# One check has no mutant, deliberately:
#
#   agent-builds    There is no valid mutation. Breaking the build is not a
#                   deletion of a taught behavior, and every other mutant
#                   already depends on the tree compiling, so this check is
#                   exercised on every run as a precondition.
#
set -uo pipefail
cd "$(dirname "$0")/.."
ROOT="$PWD"
OUT="${TMPDIR:-/tmp}/ch22-mutants"
mkdir -p "$OUT"

# The guard is scoped to the subtrees this script actually mutates. It reverts
# with 'git checkout --', so uncommitted work there would be destroyed; work
# anywhere else is none of its business.
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

	"$GRADER" -ch 22 -json ./agent >"$OUT/m$n.json" 2>"$OUT/m$n.err"
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
		echo "  PASS"
		PASSED=$((PASSED + 1))
	else
		echo "  FAIL - expected exactly: $expect"
		PASSED=$PASSED
		FAILED=$((FAILED + 1))
	fi
}

# --------------------------------------------------------------------------
# 1. back-pointer-chain
#
# Clause 3 of the chain check is NO CLOSURES: the dispatch struct a tool is
# handed must carry no func-typed fields. A closure is how the staple gets back
# in - it looks like a dependency but it is a private wire that bypasses the
# chain, and it is exactly the shape ch22 removed from cmd/main.go.
#
# Adding an unused struct field compiles in Go, which is what makes this a
# clean single-behavior deletion rather than a compile error in disguise.
# --------------------------------------------------------------------------
run_mutant 1 "dispatch struct carries a closure again" \
	"back-pointer-chain" \
	"agent/internal/common/interfaces.go" \
	"perl -0pi -e 's/^type Call struct \{/type Call struct {\n\tModelName func() string\n/m' agent/internal/common/interfaces.go" \
	"grep -q 'ModelName func() string' agent/internal/common/interfaces.go"

# --------------------------------------------------------------------------
# 2. reaches-through-the-chain
#
# Tool packages must reach the rest of the system through the hub interface
# and nothing else. Importing the engine directly is the decay this check
# exists to catch: it compiles, it works, and it quietly turns a leaf into a
# second edge in the dependency graph.
#
# The blank-identifier reference is required - an unused import does not
# compile, and a mutant that does not compile proves nothing.
# --------------------------------------------------------------------------
run_mutant 2 "a tool package imports the engine directly" \
	"reaches-through-the-chain" \
	"agent/internal/tools/statustools.go" \
	"perl -0pi -e 's{^package tools$}{package tools\n\nimport mutantllm \"github.com/waywardgeek/ensemble/agent/internal/llm\"\n\nvar _ = mutantllm.Engine{}}m' agent/internal/tools/statustools.go" \
	"grep -q 'mutantllm' agent/internal/tools/statustools.go"

# --------------------------------------------------------------------------
# 3. ch21-parity
#
# Every chapter is strictly additive, so ch22 must not be gradeable on a tree
# that has lost ch21's web access. Removing the search tool from the shipped
# skill file is the same fault ch21 itself shipped and caught: the code is all
# still there, and the feature is dead because nothing declares it.
# --------------------------------------------------------------------------
run_mutant 3 "ch21's web-search tool is undeclared" \
	"ch21-parity" \
	"agent/skills/web-search/SKILL.md" \
	"perl -0pi -e 's/firecrawl_search/firecrawl_DISABLED/g' agent/skills/web-search/SKILL.md" \
	"grep -q 'firecrawl_DISABLED' agent/skills/web-search/SKILL.md"

# --------------------------------------------------------------------------
# 4. single-composition-root
#
# The chapter's thesis: one place assembles the agent. Two call sites is how
# the tree got into the state ch22 repaired - cmd/main.go hand-built a second
# root beside the library's, and the two drifted until the library's could no
# longer build a working agent at all.
#
# This mutant restores the shape, not the original code: a second root call in
# the same binary.
# --------------------------------------------------------------------------
run_mutant 4 "a second composition root in the same binary" \
	"single-composition-root" \
	"agent/cmd/main.go" \
	"perl -0pi -e 's/^\thost := a$/\thost := a\n\t_, _ = agent.NewAgent(cfg, spec)/m' agent/cmd/main.go" \
	"grep -q '_, _ = agent.NewAgent(cfg, spec)' agent/cmd/main.go"

# --------------------------------------------------------------------------
# 5. agent-status-tool
#
# The tool must report what it claims to report. Deleting one reported figure
# leaves a tool that still exists, still runs, and still returns plausible
# text - which is why the check reads the tool's actual output out of the
# transcript rather than asserting the tool is registered.
# --------------------------------------------------------------------------
run_mutant 5 "agent_status stops reporting the cache hit rate" \
	"agent-status-tool" \
	"agent/internal/tools/statustools.go" \
	"perl -0pi -e 's/cache hit rate/cache hit RETIRED/' agent/internal/tools/statustools.go" \
	"grep -q 'cache hit RETIRED' agent/internal/tools/statustools.go"

# --------------------------------------------------------------------------
# 6. per-model-cost
#
# The bug this check exists for, restored exactly: price the whole session at
# ONE model's rate. Note the mutant does not change a single number in the
# report - it changes which price sheet the arithmetic consults, and the
# reported total stays a plausible dollar figure. That is why the check
# recomputes the expected cost from an independent copy of the price table
# instead of asserting the number is non-zero or well-formed.
# --------------------------------------------------------------------------
run_mutant 6 "every model is billed at one model's rate" \
	"per-model-cost" \
	"agent/internal/common/usage.go" \
	"perl -0pi -e 's/f, ok := LookupModel\(model\)/f, ok := LookupModel(\"claude-opus-4-6\")/' agent/internal/common/usage.go" \
	"grep -q 'LookupModel(\"claude-opus-4-6\")' agent/internal/common/usage.go"

echo
echo "=============================================================="
echo "mutants killed as expected: $PASSED"
echo "mutants that did not:       $FAILED"
echo "=============================================================="
[ "$FAILED" -eq 0 ]
