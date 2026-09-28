#!/usr/bin/env bash
# Mutation audit for the chapter 18 grader.
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
# This is a BATCH check. It builds and runs the agent eight times over and takes
# roughly fifteen minutes. It is not on the interactive path and must never be
# put there.
#
# Usage: scripts/ch18-mutants.sh [mutant-number]
#        With no argument, runs all eight.

set -uo pipefail
cd "$(dirname "$0")/.."
ROOT="$PWD"
OUT="${TMPDIR:-/tmp}/ch18-mutants"
mkdir -p "$OUT"

if ! git diff --quiet || ! git diff --cached --quiet; then
	echo "REFUSING: working tree is dirty."
	echo "This script mutates tracked files and reverts with 'git checkout --'."
	echo "Uncommitted work would be destroyed. Commit or stash first."
	exit 1
fi

GRADER="$OUT/grade"
echo "building grader..."
go build -o "$GRADER" ./cmd/grade || exit 1

# Each mutant deletes exactly ONE behaviour, and must still COMPILE: a mutant
# that fails to build proves only that the compiler works.
#
# Fields are name|expected-check-title|files-to-revert|apply-command|verify-command
# The verify command must exit 0 only if the edit actually landed. Skipping that
# step is how a botched edit gets recorded as a surviving mutant, which reads
# exactly like an insensitive test and sends you off to fix the wrong thing.

run_mutant() {
	local n="$1" name="$2" expect="$3" files="$4" apply="$5" verify="$6"
	echo
	echo "=============================================================="
	echo "mutant $n: $name"
	echo "   expected to kill: $expect"
	echo "=============================================================="

	bash -c "$apply"
	if ! bash -c "$verify"; then
		echo "  MUTATION DID NOT LAND — skipping (this is a script bug, not a result)"
		git checkout -- $files
		return
	fi
	echo "  mutation landed."

	if ! (cd agent && go build ./... >"$OUT/build.$n.log" 2>&1); then
		echo "  MUTANT DOES NOT COMPILE — invalid mutant:"
		sed -n '1,6p' "$OUT/build.$n.log" | sed 's/^/    /'
		git checkout -- $files
		return
	fi

	# DRY=1 validates that every mutation lands and compiles without paying for
	# eight agent runs. Worth having: a botched perl substitution and an
	# insensitive check produce the same output, and only one of them is a bug
	# in the code under test.
	if [ -n "${DRY:-}" ]; then
		echo "  compiles. (DRY: not grading)"
		git checkout -- $files
		return
	fi

	# Process group watchdog: macOS has no timeout(1), and a wedged agent would
	# otherwise hang the audit. set -m puts the grader in its own group so the
	# negative pid reaches the agents it spawned, not just the grader.
	set -m
	"$GRADER" -ch 18 ./agent >"$OUT/run.$n.log" 2>&1 &
	local pid=$!
	( sleep 300; kill -TERM -$pid 2>/dev/null ) &
	local wd=$!
	wait $pid
	kill $wd 2>/dev/null
	set +m

	echo "  failing checks:"
	if grep -q '^\[FAIL\]' "$OUT/run.$n.log"; then
		grep '^\[FAIL\]' "$OUT/run.$n.log" | sed 's/^/    /'
	else
		echo "    NONE"
	fi
	grep '^score:' "$OUT/run.$n.log" | sed 's/^/  /'

	# The verdict is the whole point. A mutant that kills its target check is
	# evidence that check is load bearing. A mutant that kills a DIFFERENT
	# check is evidence the checks overlap, which is worth seeing rather than
	# averaging away, so collateral is reported separately instead of being
	# folded into pass or fail.
	local failed_count killed
	failed_count=$(grep -c '^\[FAIL\]' "$OUT/run.$n.log")
	if grep '^\[FAIL\]' "$OUT/run.$n.log" | grep -qF "$expect"; then
		killed=yes
	else
		killed=no
	fi
	if [ "$killed" = yes ] && [ "$failed_count" = 1 ]; then
		echo "  VERDICT: KILLED, cleanly (exactly its own check)"
	elif [ "$killed" = yes ]; then
		echo "  VERDICT: KILLED, with $((failed_count - 1)) collateral check(s)"
	else
		echo "  VERDICT: *** SURVIVED *** — target check is not load bearing"
	fi

	git checkout -- $files
	echo "  reverted."
}

M1_APPLY='perl -0pi -e "s/eng\.Cache = cachelens\.New\(\"\.\", func\(s string\) \{ host\.Debugf\(\"%s\", s\) \}\)/eng.Cache = common.NopCacheLens{}; _ = cachelens.New/g" agent/cmd/main.go'
M1_VERIFY='grep -q "NopCacheLens{}; _ = cachelens.New" agent/cmd/main.go'

# Destabilise the prefix with a per-request timestamp in the system prompt.
#
# The obvious mutant, shuffling tool declaration order, CANNOT work here, and
# the reason is worth knowing. cfg.Tools is assigned from reg.Declarations() at
# startup and again on a skill change (agent/cmd/main.go), never per request. So
# a shuffled toolNames() produces one shuffled order that then stays put, the
# prefix is byte-identical across turns anyway, and the mutant survives while
# the check it targets is perfectly sound. Per-request tool instability is
# impossible by construction in this architecture.
#
# A clock in the system prompt is the real defect this check exists to catch,
# and the one the chapter warns about: it looks harmless, it is invisible in
# behaviour, and it moves the first bytes of the prefix on every single request,
# so nothing after it can ever be served from cache.
M2_APPLY='perl -0pi -e "s/System:    systemBlocks\(cfg\.SystemPrompt\),/System:    systemBlocks(cfg.SystemPrompt + time.Now().String()),/g" agent/internal/llm/claude.go && perl -0pi -e "s/^\t\"strings\"$/\t\"strings\"\n\t\"time\"/m" agent/internal/llm/claude.go'
M2_VERIFY='grep -q "cfg.SystemPrompt + time.Now().String()" agent/internal/llm/claude.go'

# The marker literal is indented with TWO tabs, and there is exactly one of it.
# The verify asserts the count is ZERO afterwards: a verify that is already
# satisfied before the edit certifies nothing, which is how the first version of
# this mutant reported "landed" while changing not one byte.
M3_APPLY='perl -0pi -e "s/\t\tCacheControl: &anthCacheControl\{Type: \"ephemeral\"\},\n//g" agent/internal/llm/claude.go'
M3_VERIFY='test "$(grep -c "CacheControl: &anthCacheControl" agent/internal/llm/claude.go)" = "0"'

M4_APPLY='perl -0pi -e "s/(\tOutput     int \`json:\"output\"\`)/\$1\n\tCostUSD    float64 \`json:\"cost_usd\"\`/" agent/internal/common/event.go'
M4_VERIFY='grep -q "CostUSD    float64" agent/internal/common/event.go'

M5_APPLY='perl -0pi -e "s/(\}, \"gui\.log\", eng\.Log, settingsStore, host\))/\$1\n\thost.RecordUsage(eng.Ctx.Usage)/" agent/cmd/main.go'
M5_VERIFY='grep -q "host.RecordUsage(eng.Ctx.Usage)" agent/cmd/main.go'

# Anthropic reports usage on TWO paths: the non-streaming reply and the streaming
# message_start event. The agent streams by default, so mutating only the
# non-streaming branch would leave the live path intact and the mutant would
# survive for a reason that has nothing to do with the check.
M6_APPLY='perl -0pi -e "s/CacheRead:  resp\.Usage\.CacheReadTokens,/CacheRead:  0,/g" agent/internal/llm/claude.go && perl -0pi -e "s/usage\.CacheRead = ev\.Message\.Usage\.CacheReadTokens/usage.CacheRead = 0/g" agent/internal/llm/claude.go'
M6_VERIFY='grep -q "CacheRead:  0," agent/internal/llm/claude.go && grep -q "usage.CacheRead = 0" agent/internal/llm/claude.go'

M7_APPLY='perl -0pi -e "s/CacheRead:  u\.CachedContentTokenCount,/CacheRead:  0,/g" agent/internal/llm/gemini.go'
M7_VERIFY='grep -q "CacheRead:  0," agent/internal/llm/gemini.go'

# markCache places BOTH history markers (the anchor at the end of the previous
# exchange and the rolling one at the current stable end), so disabling it leaves
# the system marker intact and the history uncached. This is the mutant that
# separates "a breakpoint exists" from "a breakpoint that advances", and it is
# the shape the chapter's own prose would have shipped.
M8_APPLY='perl -0pi -e "s/(func markCache\(msgs \[\]anthMsg, msg, blocks int\) bool \{\n)/\$1\treturn false\n/" agent/internal/llm/claude.go'
M8_VERIFY='sed -n "163p" agent/internal/llm/claude.go | grep -q "return false"'

declare -a NAMES=(
	"lens not wired: measurement silently absent"
	"clock in the system prompt: prefix moves every request"
	"system prompt carries no cache_control"
	"cost stored in Usage instead of computed"
	"meter reports lifetime spend, not this session"
	"Anthropic parser drops cache_read"
	"Gemini parser drops cache_read"
	"history breakpoints disabled: system cached, conversation not"
)
declare -a EXPECTS=(
	"a cache lens observes"
	"byte-identical across turns"
	"both carry a cache_control marker"
	"cost is derived at display time"
	"not the lifetime restored from disk"
	"cache reads are zero on a cold turn"
	"all four usage categories"
	"both carry a cache_control marker"
)
declare -a FILES=(
	"agent/cmd/main.go"
	"agent/internal/llm/claude.go"
	"agent/internal/llm/claude.go"
	"agent/internal/common/event.go"
	"agent/cmd/main.go"
	"agent/internal/llm/claude.go"
	"agent/internal/llm/gemini.go"
	"agent/internal/llm/claude.go"
)
declare -a APPLIES=("$M1_APPLY" "$M2_APPLY" "$M3_APPLY" "$M4_APPLY" "$M5_APPLY" "$M6_APPLY" "$M7_APPLY" "$M8_APPLY")
declare -a VERIFIES=("$M1_VERIFY" "$M2_VERIFY" "$M3_VERIFY" "$M4_VERIFY" "$M5_VERIFY" "$M6_VERIFY" "$M7_VERIFY" "$M8_VERIFY")

ONLY="${1:-}"
for i in 0 1 2 3 4 5 6 7; do
	n=$((i + 1))
	if [ -n "$ONLY" ] && [ "$ONLY" != "$n" ]; then continue; fi
	run_mutant "$n" "${NAMES[$i]}" "${EXPECTS[$i]}" "${FILES[$i]}" "${APPLIES[$i]}" "${VERIFIES[$i]}"
done

echo
echo "logs in $OUT"
git status --porcelain
