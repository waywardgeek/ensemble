#!/usr/bin/env bash
#
# tts-harness.sh - run this system once, with its speech channel recorded.
#
# Chapter 14 asks for a way to find out what an agent actually said, rather
# than what it drew on screen. This script is the door the grader knocks on.
# It runs the whole system for one prompt, in a mode where every utterance is
# written to a log, then exits.
#
# The grader supplies the conversation by pointing the agent at a fake model
# server, so a run is deterministic: no API key, no network, no charge.
#
# Contract
#   Environment
#     LLM_BASE_URL  where the model lives. Point the agent here, not at a
#                   real vendor.
#     TTS_LOG       file to write the speech log to, one JSON object per line.
#     WORKSPACE     the directory the agent should treat as its own, which is
#                   where the grader plants any file it asks the agent to read.
#     STREAMING     "on" (default) or "off". With it off, the system must
#                   still speak: a reply that arrives whole rather than in
#                   pieces is the case chapter 14 found silent. How you turn
#                   streaming off is your business; this harness sets the
#                   environment variable its own agent reads.
#   Arguments
#     "$1"          the prompt to send, as if a person had typed it.
#     --type-during-turn TEXT
#                   optional. Type TEXT into the input while the agent is
#                   still talking, then clear it. Used to show that speech
#                   pauses while a person is typing.
#     --describe    print what this harness supports, as JSON, and exit.
#   Exit status
#     0 if the run completed. Non-zero means the harness itself failed, which
#     is different from the system behaving badly: a run that completes but
#     says the wrong things still exits 0 and fails on the log contents.

set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/.." && pwd)"

# --------------------------------------------------------------- --describe
#
# The grader needs two things it cannot guess: which tool reads a file, and
# what that tool calls the path argument. Everything else has a default. A
# system that kept the names used in this book does not need this at all, but
# printing it costs nothing and makes the harness self-describing.

if [ "${1:-}" = "--describe" ]; then
  cat <<'JSON'
{
  "read_file_tool": "read_file",
  "read_file_arg": "path",
  "supports_type_during_turn": true,
  "supports_streaming_off": true,
  "log_format": "jsonl"
}
JSON
  exit 0
fi

prompt="${1:-}"
shift || true

type_during_turn=""
while [ $# -gt 0 ]; do
  case "$1" in
    --type-during-turn) type_during_turn="${2:-}"; shift 2 ;;
    *) shift ;;
  esac
done

: "${TTS_LOG:?TTS_LOG must be set}"
: "${LLM_BASE_URL:?LLM_BASE_URL must be set}"
workspace="${WORKSPACE:-$PWD}"

command -v node >/dev/null 2>&1 || {
  echo "tts-harness: node is required to run the GUI headlessly" >&2
  exit 1
}

# ------------------------------------------------------------------- build
#
# Built into a scratch directory rather than anywhere inside the tree, so a
# run leaves nothing behind and cannot collide with whatever the student
# already keeps there. Go's build cache means only the link step repeats.

tmp="$(mktemp -d)"
agent_pid=""

cleanup() {
  exec 9>&- 2>/dev/null || true          # closing stdin asks the agent to stop
  if [ -n "$agent_pid" ] && kill -0 "$agent_pid" 2>/dev/null; then
    for _ in $(seq 1 30); do
      kill -0 "$agent_pid" 2>/dev/null || break
      sleep 0.1
    done
    kill -9 "$agent_pid" 2>/dev/null || true
  fi
  rm -rf "$tmp"
}
trap cleanup EXIT

( cd "$root" && go build -o "$tmp/ensemble" ./cmd/ )

# -------------------------------------------------------------------- port
#
# Ask the operating system for a free port by binding one and letting go of
# it, which is more reliable than guessing a number and hoping.

port="$(node -e '
const net = require("net");
const srv = net.createServer();
srv.listen(0, "127.0.0.1", () => {
  const p = srv.address().port;
  srv.close(() => { process.stdout.write(String(p)); });
});
')"

if [ -z "$port" ]; then
  echo "tts-harness: could not find a free port" >&2
  exit 1
fi

# --------------------------------------------------------------- the agent
#
# The agent reads stdin and stops when it reaches the end of it, so stdin is
# held open with a pipe for as long as the run needs. Closing that pipe is how
# the agent is asked to stop.

mkfifo "$tmp/stdin"
# Read-write, because opening a pipe write-only blocks until a reader shows
# up, and the reader here is the agent that has not started yet.
exec 9<>"$tmp/stdin"

: > "$TTS_LOG"

# The contract speaks of STREAMING=on|off. Translating that into whatever
# your own system understands is exactly the part that is yours to write.
disable_streaming=""
if [ "${STREAMING:-on}" = "off" ]; then
  disable_streaming=1
fi

(
  cd "$workspace"
  LLM_BASE_URL="$LLM_BASE_URL" \
  EN_DISABLE_STREAMING="$disable_streaming" \
    "$tmp/ensemble" --port "$port" --tts-log "$TTS_LOG" \
    > "$tmp/agent.log" 2>&1 < "$tmp/stdin"
) &
agent_pid=$!

# Wait for the port to start accepting rather than sleeping a guessed amount.
ready=""
for _ in $(seq 1 100); do
  if curl -s -o /dev/null --max-time 1 "http://127.0.0.1:$port/" 2>/dev/null; then
    ready=1
    break
  fi
  sleep 0.1
done
if [ -z "$ready" ]; then
  echo "tts-harness: agent did not start listening on port $port" >&2
  sed -n '1,40p' "$tmp/agent.log" >&2 || true
  exit 1
fi

# ------------------------------------------------------------- the browser

browser_args=( --url "http://127.0.0.1:$port" --prompt "$prompt" )
if [ -n "$type_during_turn" ]; then
  browser_args+=( --type-during-turn "$type_during_turn" )
fi

if ! node "$here/tts-browser.js" "${browser_args[@]}"; then
  echo "tts-harness: the headless GUI run failed" >&2
  sed -n '1,40p' "$tmp/agent.log" >&2 || true
  exit 1
fi

exit 0
