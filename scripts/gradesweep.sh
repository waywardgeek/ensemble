#!/bin/bash
# Grade sweep: run every chapter grader at its canonical target and record the
# score. Targets match the Makefile's grade* rules.
#
# Usage: bash scripts/gradesweep.sh [OUTFILE]
#   DEADLINE=300 bash scripts/gradesweep.sh     # override the per-grader limit
#
# Every grader runs under a hard deadline. A grader that exceeds it is killed
# and recorded as DEADLINE-EXCEEDED rather than being allowed to hang the
# sweep. A tool that runs unbounded is not functional: it cannot be used in a
# loop, it cannot be trusted in CI, and when it wedges there is nothing to
# distinguish "slow" from "stuck" except a human losing patience.
cd "$(dirname "$0")/.."

OUT="${1:-/tmp/gradesweep.txt}"
DEADLINE="${DEADLINE:-180}"
: > "$OUT"

# Build the grader once. The old sweep used `go run` for all 25 invocations,
# which re-links the binary every time.
WORK="$(mktemp -d)"
GRADE_BIN="$WORK/grade"
trap 'rm -rf "$WORK"' EXIT
echo "building grader..."
if ! go build -o "$GRADE_BIN" ./cmd/grade; then
  echo "FATAL: grader build failed" | tee -a "$OUT"
  exit 1
fi

# with_deadline SECS CMD... runs CMD and kills it if it outlives SECS.
#
# It signals the process GROUP, not the process. Graders spawn agents, and
# those agents spawn children of their own; killing only the grader leaves
# them orphaned, holding ports, until someone notices days later. `set -m`
# turns on job control so the child becomes a process group leader and the
# negative pid reaches everything it started.
#
# SIGTERM first so a script's `trap cleanup EXIT` can actually run, SIGKILL
# after a grace period for anything a trap will not reach. macOS has no
# timeout(1), so this is hand-rolled rather than a coreutils call.
with_deadline() {
  local secs="$1"; shift
  set -m
  "$@" &
  local pid=$!
  set +m
  (
    sleep "$secs"
    kill -TERM "-$pid" 2>/dev/null || kill -TERM "$pid" 2>/dev/null
    sleep 5
    kill -KILL "-$pid" 2>/dev/null || kill -KILL "$pid" 2>/dev/null
  ) 2>/dev/null &
  local watchdog=$!
  wait "$pid"
  local rc=$?
  kill "$watchdog" 2>/dev/null
  wait "$watchdog" 2>/dev/null
  return $rc
}

BREACHED=0

# ONLY="1 6 14" runs just those chapters. The full sweep is long, so being
# able to re-check one chapter without paying for all of them matters.
ONLY="${ONLY:-}"

wanted() {
  [ -z "$ONLY" ] && return 0
  for c in $ONLY; do [ "$c" = "$1" ] && return 0; done
  return 1
}

run() {
  CH="$1"; DIR="$2"
  wanted "$CH" || return
  if [ ! -d "$DIR" ]; then
    echo "ch$CH $DIR MISSING" | tee -a "$OUT"
    return
  fi

  local log="$WORK/ch$CH.log"
  local t0=$SECONDS
  with_deadline "$DEADLINE" "$GRADE_BIN" -ch "$CH" "$DIR" > "$log" 2>&1
  local rc=$?
  local elapsed=$((SECONDS - t0))

  local score
  score=$(grep -Eo 'score: [0-9]+/[0-9]+' "$log" | tail -1)

  # A shell reports a signalled child as 128+signum: 143 SIGTERM, 137 SIGKILL.
  if [ "$rc" -ge 128 ] && [ -z "$score" ]; then
    echo "ch$CH $DIR DEADLINE-EXCEEDED (${elapsed}s, limit ${DEADLINE}s)" | tee -a "$OUT"
    BREACHED=$((BREACHED + 1))
    return
  fi

  echo "ch$CH $DIR ${score:-NOSCORE} (${elapsed}s)" | tee -a "$OUT"
}

# Canonical targets (Makefile). A chapter's score is meaningful only here:
# ch1-ch4 grade solutions/chNN, ch6 alone grades a frozen snapshot, and ch5
# plus ch7-ch17 grade the live tree.
run 1  ./solutions/ch01
run 2  ./solutions/ch02
run 3  ./solutions/ch03
run 4  ./solutions/ch04
run 5  ./agent
run 6  ./solutions/ch06/agent
run 7  ./agent
run 8  ./agent
run 9  ./agent
run 10 ./agent
run 11 ./agent
run 12 ./agent
run 13 ./agent
run 14 ./agent
run 15 ./agent
run 16 ./agent
run 17 ./agent
run 18 ./agent
run 19 ./agent
# Per-chapter reference trees that also get graded.
run 10 ./solutions/ch10
run 11 ./solutions/ch11
run 12 ./solutions/ch12
run 13 ./solutions/ch13
run 14 ./solutions/ch14
run 15 ./solutions/ch15
run 16 ./solutions/ch16
run 17 ./solutions/ch17
run 18 ./solutions/ch18
run 19 ./solutions/ch19
run 21 ./solutions/ch21

{
  echo "SWEEP-DONE breached=$BREACHED limit=${DEADLINE}s"
} | tee -a "$OUT"

# Leaked agents are the standing hazard here: 35 were once found alive, the
# oldest over four days. Report rather than kill, since a developer's own
# agent can match loosely written patterns.
LEAKED=$(pgrep -f '/T/tmp\..*/ensemble' 2>/dev/null | wc -l | tr -d ' ')
if [ "$LEAKED" != "0" ]; then
  echo "WARNING: $LEAKED leaked agent process(es) still running" | tee -a "$OUT"
fi

[ "$BREACHED" -eq 0 ]
