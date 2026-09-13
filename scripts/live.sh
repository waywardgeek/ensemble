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
# The agent runs in a FRESH TEMP DIRECTORY, not in the course repo. Chapter 3's
# agent can write files and run commands, and its working directory is the only
# thing deciding where a relative path lands — the scripted demo once left a
# PROJECT_CODENAME.txt in the repo root. The temp directory is seeded with a few
# small files (four of them ending in .md, which is what round 2 counts) and is
# left behind after the run so you can see what the agent did in it. The path is
# printed at startup.
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

# --- a scratch directory, OUTSIDE the repository ----------------------------
#
# Chapter 3's agent has write_file and run_command, and its working directory is
# the only thing deciding where a relative path lands. Run it in the repo and it
# writes into the repo: the scripted demo once invented a codename and left
# PROJECT_CODENAME.txt sitting in the repo root. Nothing was sandboxed; it was
# simply standing in the wrong place. So the agent now runs in a fresh temp
# directory and the repository that holds the book is never the thing it is
# standing in.
#
# The seeded files also make the demo SELF-CONTAINED. Round 2 asks how many
# entries end in .md. Pointed at the course repo that answer changed every time
# a chapter was added — the demo was already non-deterministic, and a bare
# scratch directory would have made the answer zero.
#
# The files being counted live in notes/, one level down, and NOT in the
# directory the agent is standing in. That is deliberate and was learned the
# hard way: on the first run of this script the agent answered round 1 by
# writing its codename to PROJECT_CODENAME.md in the working directory, which
# is a .md file, which made round 2 answer 5 instead of 4. The agent was
# perturbing the very thing it was about to be asked to measure. Counting a
# directory it has no reason to write into leaves it free to demonstrate that it
# CAN write — the point of the exercise — without moving the answer.
md_fixtures=4
setup_scratch() {
	run_root="$(mktemp -d "${TMPDIR:-/tmp}/live-ch0$chapter.XXXXXX")"
	bin="$run_root/bin/ch0$chapter"
	work="$run_root/work"
	mkdir -p "$run_root/bin" "$work/notes"

	go build -o "$bin" "$solution"

	cat >"$work/notes/README.md" <<'EOF'
# Scratch project

A throwaway project for a live agent demo. Nothing here is precious, and
nothing here is in the course repository.
EOF
	cat >"$work/notes/design-notes.md" <<'EOF'
# Design notes

The agent is standing in a temp directory, not in the repo that holds the book.
EOF
	cat >"$work/notes/meeting-minutes.md" <<'EOF'
# Meeting minutes

Decided: keep the demo self-contained, so its correct answer does not drift
every time the course gains a chapter.
EOF
	cat >"$work/notes/glossary.md" <<'EOF'
# Glossary

scratch directory — somewhere an agent may write without consequence.
EOF
	# Two distractors, so "how many end in .md" is a real filter over six
	# entries rather than a head count of the directory.
	cat >"$work/notes/config.json" <<'EOF'
{"name": "scratch", "version": 1}
EOF
	cat >"$work/notes/run.sh" <<'EOF'
#!/usr/bin/env bash
echo "a distractor whose name does not end in .md"
EOF

	echo "working directory: $work"
	echo "(a scratch directory outside the repo; it is left behind so you can see"
	echo " what the agent did in it)"
}

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
	setup_scratch
	cd "$work"
	exec "$bin" chat
	;;


rounds)
	echo "chapter $chapter, $vendor, model: ${LLM_MODEL:-${!model_var:-(solution default)}}"
	setup_scratch
	cd "$work"
	if [ "$chapter" = 3 ]; then
		echo "three live rounds — round 2 needs a tool, round 3 asks it to recall round 1,"
		echo "which only works if the tool loop ran and the whole history is being resent."
		echo "round 2 has exactly one right answer here: of the six entries in"
		echo "notes/, $md_fixtures end in .md."
		echo
		printf '%s\n' \
			'{"user":"Hello! I am starting a new project. Give it a one-word codename and remember it."}' \
			'{"user":"Use your list_directory tool on the notes directory and tell me how many entries end in .md. Do not guess; call the tool."}' \
			'{"user":"What codename did you give my project?"}' |
			"$bin"
	else
		echo "three live rounds — round 3 asks it to recall what it said in round 1,"
		echo "which only works if the whole history is being resent."
		echo
		printf '%s\n' \
			'{"user":"Hello! I am starting a new project. Give it a codename and remember it."}' \
			'{"user":"What is the capital of France?"}' \
			'{"user":"What codename did you give my project?"}' |
			"$bin"
	fi
	;;

esac
