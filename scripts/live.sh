#!/usr/bin/env bash
#
# Run a chapter's reference solution against a REAL vendor API.
#
#   scripts/live.sh [chapter] [vendor] [rounds|chat|models]
#
#   chapter   1, 2 or 3                      default 3
#   vendor    anthropic, openai or gemini    default anthropic  (chapter 1 speaks only anthropic)
#   mode      rounds  three scripted turns, then the token bill   (default)
#             chat    interactive REPL — talk to it yourself
#             models  list the model IDs your key can actually use
#
# Arguments may come in any order; each is recognised by shape.
#
#   scripts/live.sh                      chapter 3, anthropic, rounds
#   scripts/live.sh 3 gemini chat        chapter 3 REPL on Gemini
#   scripts/live.sh openai models        which OpenAI models can this key see?
#   scripts/live.sh 1 chat               the chapter 1 REPL, as before
#
# This costs real tokens. A `rounds` run of chapter 3 is roughly six to eight
# requests as the history grows — on the order of 20k input tokens, a few
# cents at 2026 prices. `chat` costs whatever you type.
#
# To run with NO key and NO cost, use the fake instead:
#
#   go run ./cmd/fakevendor -ch 3 -vendor gemini chat
#
# The API key is taken from, in order (VENDOR is ANTHROPIC, OPENAI or GEMINI):
#   1. $LLM_API_KEY
#   2. $VENDOR_API_KEY
#   3. the file named by $VENDOR_API_KEY_FILE
#   4. ~/.cr/settings.json (CodeRhapsody's config, if you use it): directClaudeAPIKey,
#      directOpenAIAPIKey or directGeminiAPIKey
#
# The key is never echoed. Leave $VENDOR_BASE_URL unset for the vendor's own
# endpoint (a full URL, scheme included), or point it at a proxy.

set -euo pipefail
cd "$(dirname "$0")/.."

chapter=3
vendor=anthropic
mode=rounds
for arg in "$@"; do
	case "$arg" in
	1 | 2 | 3) chapter="$arg" ;;
	anthropic | openai | gemini) vendor="$arg" ;;
	rounds | chat | models) mode="$arg" ;;
	*)
		echo "usage: scripts/live.sh [1|2|3] [anthropic|openai|gemini] [rounds|chat|models]" >&2
		exit 2
		;;
	esac
done

if [ "$chapter" = 1 ] && [ "$vendor" != anthropic ]; then
	echo "chapter 1 speaks only the Anthropic dialect; use chapter 2 or 3 for $vendor" >&2
	exit 2
fi

case "$vendor" in
anthropic) V=ANTHROPIC default_base=https://api.anthropic.com ;;
openai) V=OPENAI default_base=https://api.openai.com ;;
gemini) V=GEMINI default_base=https://generativelanguage.googleapis.com ;;
esac
key_var="${V}_API_KEY"
file_var="${V}_API_KEY_FILE"
base_var="${V}_BASE_URL"
model_var="${V}_MODEL"

# --- resolve the key, never printing it ------------------------------------
key="${LLM_API_KEY:-${!key_var:-}}"
if [ -z "$key" ] && [ -n "${!file_var:-}" ]; then
	key="$(tr -d '[:space:]' <"${!file_var}")"
fi
if [ -z "$key" ] && [ -f "$HOME/.cr/settings.json" ]; then
	case "$vendor" in
	anthropic) field=directClaudeAPIKey ;;
	openai) field=directOpenAIAPIKey ;;
	gemini) field=directGeminiAPIKey ;;
	esac
	key="$(python3 -c "
import json, os, sys
p = os.path.expanduser('~/.cr/settings.json')
print(json.load(open(p)).get(sys.argv[1], ''))
" "$field")"
fi
if [ -z "$key" ]; then
	echo "No API key found for $vendor. Try:" >&2
	echo "    export $key_var=..." >&2
	echo "or run against the fake, which needs none:" >&2
	echo "    go run ./cmd/fakevendor -ch $chapter -vendor $vendor chat" >&2
	exit 1
fi
export "$key_var=$key"
export LLM_VENDOR="$vendor"
base="${!base_var:-$default_base}"
solution="./solutions/ch0$chapter"

# --- modes -----------------------------------------------------------------
case "$mode" in
models)
	case "$vendor" in
	anthropic)
		curl -sS "$base/v1/models?limit=40" -H "x-api-key: $key" -H "anthropic-version: 2023-06-01" |
			python3 -c "
import json, sys
for m in json.load(sys.stdin).get('data', []):
    print(f\"{m['id']:<40} {m.get('display_name','')}\")
"
		;;
	openai)
		curl -sS "$base/v1/models" -H "Authorization: Bearer $key" |
			python3 -c "
import json, sys
for m in sorted(json.load(sys.stdin).get('data', []), key=lambda m: m['id']):
    print(m['id'])
"
		;;
	gemini)
		curl -sS "$base/v1beta/models?pageSize=100&key=$key" |
			python3 -c "
import json, sys
for m in json.load(sys.stdin).get('models', []):
    print(f\"{m['name'].removeprefix('models/'):<40} {m.get('displayName','')}\")
"
		;;
	esac
	;;

chat)
	echo "chapter $chapter, $vendor, model: ${LLM_MODEL:-${!model_var:-(solution default)}}"
	exec go run "$solution" chat
	;;

rounds)
	echo "chapter $chapter, $vendor, model: ${LLM_MODEL:-${!model_var:-(solution default)}}"
	if [ "$chapter" = 3 ]; then
		echo "three live rounds — round 2 needs a tool, round 3 asks it to recall round 1,"
		echo "which only works if the tool loop ran and the whole history is being resent."
		echo
		printf '%s\n' \
			'{"user":"Hello! I am starting a new project. Give it a one-word codename and remember it."}' \
			'{"user":"Use your list_directory tool on the current directory and tell me how many entries end in .md. Do not guess; call the tool."}' \
			'{"user":"What codename did you give my project?"}' |
			go run "$solution"
	else
		echo "three live rounds — round 3 asks it to recall what it said in round 1,"
		echo "which only works if the whole history is being resent."
		echo
		printf '%s\n' \
			'{"user":"Hello! I am starting a new project. Give it a codename and remember it."}' \
			'{"user":"What is the capital of France?"}' \
			'{"user":"What codename did you give my project?"}' |
			go run "$solution"
	fi
	;;
esac
