#!/usr/bin/env bash
# Live cache check: run the agent against the real vendors and compare what the
# lens PREDICTED against what the provider actually GRANTED.
#
# Why this exists. Everything else in chapter 18 is graded against a fake
# vendor, which is correct for a grader: it is deterministic, it needs no
# network, and it costs nothing. But a fake vendor cannot falsify the one claim
# the whole chapter rests on, because the fake grants whatever cache figures we
# tell it to. Only a real provider can confirm that a prefix we believe is
# cacheable is a prefix it will actually serve from cache.
#
# One assumption in particular needs a real vendor to settle. Our second
# breakpoint MOVES: it rolls along the message history so each turn banks the
# exchange before it. That only works if the marker is not part of the cache
# key, and if it were, every turn would miss and the agent would silently pay
# full price forever while the lens reported a healthy prefix. This script is
# how that question gets an answer rather than an argument.
#
# THIS SPENDS REAL MONEY. A handful of very short turns per vendor, so cents
# rather than dollars, but it is not free and it is not hermetic. It is
# deliberately NOT part of the grader and NOT part of the sweep.
#
# Usage:
#   scripts/live-cache-check.sh                 # all three vendors
#   scripts/live-cache-check.sh anthropic       # just one
#
# Keys are read from ~/.cr/settings.json and passed to the child through the
# environment. They are never printed, never logged, and never written to the
# run directory.

set -uo pipefail
cd "$(dirname "$0")/.."
ROOT="$PWD"
SETTINGS="$HOME/.cr/settings.json"

# Model IDs are the real frontier models, and they are quadruple checked. Do not
# "helpfully" look them up and substitute something that looks more plausible.
#
# Lookups are case statements rather than associative arrays because macOS ships
# bash 3.2, where `declare -A` is a syntax error waiting in a branch you have not
# taken yet. Indexed arrays are fine and are used below.
VENDORS="anthropic openai gemini"

model_for() {
	case "$1" in
	anthropic) echo "claude-sonnet-5" ;;
	openai) echo "gpt-5.6-sol" ;;
	gemini) echo "gemini-3.8-flash" ;;
	esac
}

keyname_for() {
	case "$1" in
	anthropic) echo "directClaudeAPIKey" ;;
	openai) echo "directOpenAIAPIKey" ;;
	gemini) echo "directGeminiAPIKey" ;;
	esac
}

if [ ! -f "$SETTINGS" ]; then
	echo "no $SETTINGS: cannot run a live check without keys" >&2
	exit 1
fi
command -v jq >/dev/null || { echo "jq required" >&2; exit 1; }

BIN="${TMPDIR:-/tmp}/live-cache-agent"
echo "building agent..."
(cd agent && go build -o "$BIN" ./cmd) || exit 1

# Three turns, not two. Turn 1 is cold and writes the cache. Turn 2 reads the
# system prefix. Turn 3 is the one that matters: by then the ROLLING marker has
# had a chance to bank turn 2's exchange, so if the history breakpoint works, the
# cache read should cover more than the system prompt alone.
PROMPTS=(
	"Reply with exactly the word one, nothing else."
	"Reply with exactly the word two, nothing else."
	"Reply with exactly the word three, nothing else."
)

# The agent is driven over stdin as JSON lines, one object per message, which is
# the same door the GUI and the graders use. Turn completion is detected by
# counting response_ended entries in the event log rather than by sleeping,
# because a fixed sleep either flakes or wastes time, and against a real vendor
# the turn length is genuinely unpredictable.
wait_for_turns() {
	local log="$1" want="$2" deadline=$((SECONDS + 120))
	while [ $SECONDS -lt $deadline ]; do
		if [ -f "$log" ] && [ "$(grep -c '"type":"response_ended"' "$log" 2>/dev/null || echo 0)" -ge "$want" ]; then
			return 0
		fi
		sleep 1
	done
	return 1
}

run_vendor() {
	local v="$1"
	local model keyname
	model=$(model_for "$v")
	keyname=$(keyname_for "$v")
	local key
	key=$(jq -r --arg k "$keyname" '.[$k] // empty' "$SETTINGS")

	echo
	echo "=============================================================="
	echo "$v / $model"
	echo "=============================================================="
	if [ -z "$key" ]; then
		echo "  SKIPPED: $keyname not present in $SETTINGS"
		return
	fi

	local work
	work=$(mktemp -d "${TMPDIR:-/tmp}/livecache.$v.XXXXXX")
	local evlog="$work/events.jsonl"
	local fifo="$work/in"
	mkfifo "$fifo"

	# Open the FIFO read-write from this shell. Read-write and not write-only:
	# opening a FIFO for writing BLOCKS until a reader arrives, and the reader
	# here is the agent, which has not been launched yet. Write-only deadlocks
	# the script before the agent ever starts, and the symptom is a run
	# directory containing the FIFO and nothing else.
	#
	# Holding it open also stops the agent seeing EOF after the first prompt and
	# shutting down before the cache has anything to hit.
	exec 9<>"$fifo"

	(
		cd "$work" || exit 1
		LLM_VENDOR="$v" \
		LLM_MODEL="$model" \
		LLM_API_KEY="$key" \
		CH02_LOG="$evlog" \
		EN_SKILLS_DIR="$ROOT/agent/skills" \
		EN_PRIMARY_SKILL=ensemble \
			"$BIN" --port 0 --gui-dir "$ROOT/agent/web/gui" <"$fifo" >"$work/out.log" 2>&1
	) &
	local apid=$!

	local n=0
	for p in "${PROMPTS[@]}"; do
		n=$((n + 1))
		printf '%s\n' "$(jq -nc --arg t "$p" '{kind:"prompt",text:$t}')" >&9
		if ! wait_for_turns "$evlog" "$n"; then
			echo "  turn $n did not complete within the deadline"
			break
		fi
		echo "  turn $n complete"
	done

	# Close our end of the FIFO, then stop the agent ourselves rather than
	# waiting for it to notice. Closing stdin does NOT make it exit: it stays up
	# serving its GUI socket, which is correct behaviour for a long-lived agent
	# and a hang for a script that plans to `wait`. Give it a moment to finish
	# flushing its logs, then take the process group down.
	exec 9>&-
	sleep 2
	kill -TERM "-$apid" 2>/dev/null || kill -TERM "$apid" 2>/dev/null
	sleep 1
	kill -KILL "-$apid" 2>/dev/null || true
	wait "$apid" 2>/dev/null

	report "$v" "$work"
	echo "  run dir: $work"
}

# The comparison is the whole point, so it is printed side by side.
#
# These two numbers are NOT the same kind of thing, and the chapter is emphatic
# about not mixing them. The prediction is a fraction of BYTES, measured from the
# request before it was sent. The grant is a fraction of TOKENS, reported by the
# provider afterwards. Comparing them as ratios is legitimate; adding or
# subtracting them is not. They should land close to each other, and a large gap
# is the finding, not a rounding error.
report() {
	local v="$1" work="$2"
	echo
	echo "  --- what the lens predicted (bytes, from the request) ---"
	if [ -f "$work/debug.log" ]; then
		grep -o "cachelens: .*" "$work/debug.log" | tail -3 | sed 's/^/    /' || echo "    (no lens lines)"
	else
		echo "    (no debug.log)"
	fi

	echo "  --- what the provider granted (tokens, from the reply) ---"
	if [ -f "$work/events.jsonl" ]; then
		jq -r 'select(.type=="response_ended") | .response.usage
		       | "    input=\(.input) cache_write=\(.cache_write) cache_read=\(.cache_read) output=\(.output)"' \
			"$work/events.jsonl" 2>/dev/null || echo "    (could not parse usage)"

		# The verdict. A cache_read of zero on every turn means nothing was ever
		# served from cache, which is the failure this chapter is about and is
		# invisible in the agent's behaviour.
		local reads
		reads=$(jq -r 'select(.type=="response_ended") | .response.usage.cache_read' "$work/events.jsonl" 2>/dev/null | paste -sd, -)
		echo "  --- verdict ---"
		echo "    cache_read per turn: ${reads:-none}"
		local maxread
		maxread=$(jq -rs 'map(select(.type=="response_ended") | .response.usage.cache_read) | max // 0' "$work/events.jsonl" 2>/dev/null)
		if [ "${maxread:-0}" = "0" ] || [ -z "${maxread:-}" ]; then
			echo "    NOTHING WAS CACHED. Either the prefix never reached the provider's"
			echo "    minimum cacheable size, or the breakpoints are not doing their job."
		else
			echo "    the provider served $maxread cached tokens at peak, so the"
			echo "    breakpoints are real and the moving marker does not break them"
		fi
	else
		echo "    (no event log)"
	fi
}

ONLY="${1:-}"
for v in $VENDORS; do
	if [ -n "$ONLY" ] && [ "$ONLY" != "$v" ]; then continue; fi
	run_vendor "$v"
done
echo
echo "done. These runs cost real money and are not part of the grader."
