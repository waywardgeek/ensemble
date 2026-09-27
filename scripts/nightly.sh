#!/bin/bash
# Nightly verification. Runs the expensive checks unattended and leaves a
# timestamped report to read in the morning.
#
#   bash scripts/nightly.sh
#
# Why this exists. There are two kinds of check here and they were being run
# as if they were one. Interactive checks answer "did I just break it" and a
# human waits for them, so they must be fast, and anything over about three
# minutes is broken by definition. Batch checks answer "is everything still
# true" and nobody waits for them, so they may take as long as they honestly
# need. Trying to fit the batch checks into the interactive budget produced
# the worst of both: a sweep too slow to use in a loop and too rushed to be
# thorough.
#
# Interactive loop, during the day:
#   ONLY=14 bash scripts/gradesweep.sh      # one chapter, seconds
#   go test ./internal/grade/ -run TestCh14 # one chapter's tests
#
# Batch, overnight: this script.
#
# Install with launchd (macOS), 2am daily:
#   see scripts/nightly.plist
cd "$(dirname "$0")/.."
REPO="$(pwd)"

STAMP="$(date +%Y-%m-%d-%H%M)"
REPORT_DIR="${NIGHTLY_DIR:-$REPO/.nightly}"
mkdir -p "$REPORT_DIR"
REPORT="$REPORT_DIR/$STAMP.md"
LATEST="$REPORT_DIR/latest.md"

start_epoch=$(date +%s)

{
  echo "# Nightly verification - $(date '+%Y-%m-%d %H:%M:%S')"
  echo
  echo "Commit: \`$(git rev-parse --short HEAD)\` on \`$(git rev-parse --abbrev-ref HEAD)\`"
  dirty=$(git status --porcelain | grep -v '^??' | wc -l | tr -d ' ')
  echo "Uncommitted tracked changes: $dirty"
  echo
} > "$REPORT"

section() { echo "" >> "$REPORT"; echo "## $1" >> "$REPORT"; echo "" >> "$REPORT"; }

# --- build and vet -----------------------------------------------------
section "Build and vet"
{
  echo '```'
  if (cd agent && go build ./... 2>&1); then echo "agent build: OK"; else echo "agent build: FAIL"; fi
  if go build ./... 2>&1; then echo "root build: OK"; else echo "root build: FAIL"; fi
  if (cd agent && go vet ./... 2>&1 | head -20); then :; fi
  echo "vet: done"
  unformatted=$(gofmt -l . 2>/dev/null | grep -v '^solutions/' | tr '\n' ' ')
  echo "gofmt offenders (excluding solutions/): ${unformatted:-none}"
  echo '```'
} >> "$REPORT"

# --- grader sweep ------------------------------------------------------
section "Grader sweep"
sweep_start=$(date +%s)
DEADLINE="${SWEEP_DEADLINE:-300}" bash scripts/gradesweep.sh "$REPORT_DIR/$STAMP-sweep.txt" > /dev/null 2>&1
sweep_rc=$?
sweep_elapsed=$(( $(date +%s) - sweep_start ))
{
  echo '```'
  cat "$REPORT_DIR/$STAMP-sweep.txt"
  echo '```'
  echo
  echo "Sweep took ${sweep_elapsed}s, exit $sweep_rc."
} >> "$REPORT"

# Anything not at full marks, called out so it is not buried in 25 lines.
section "Chapters not at full marks"
{
  echo '```'
  grep -E '[0-9]+/[0-9]+' "$REPORT_DIR/$STAMP-sweep.txt" \
    | grep -vE '(100/100|120/120)' || echo "(none)"
  grep -E 'DEADLINE-EXCEEDED|MISSING|NOSCORE' "$REPORT_DIR/$STAMP-sweep.txt" || true
  echo '```'
} >> "$REPORT"

# --- full test suite ---------------------------------------------------
# -p 1 is mandatory. Graders that overlap produce false scores: a chapter
# once measured 11/100 in a concurrent sweep and 100/100 by itself.
section "Test suite (go test ./... -p 1)"
test_start=$(date +%s)
go test ${NIGHTLY_TEST_PKGS:-./...} ${NIGHTLY_TEST_ARGS:-} -count=1 -p 1 -timeout 30m > "$REPORT_DIR/$STAMP-tests.txt" 2>&1
test_rc=$?
test_elapsed=$(( $(date +%s) - test_start ))
{
  echo '```'
  grep -E '^(ok|FAIL|---)' "$REPORT_DIR/$STAMP-tests.txt" | head -40
  echo '```'
  echo
  echo "Tests took ${test_elapsed}s, exit $test_rc."
  echo
  echo "### Failures"
  echo '```'
  grep -E '^(    )?--- FAIL' "$REPORT_DIR/$STAMP-tests.txt" || echo "(none)"
  echo '```'
} >> "$REPORT"

# --- slowest tests -----------------------------------------------------
# Kept because slow tests are how the interactive loop gets ruined, and they
# are invisible unless something records them.
section "Slowest tests"
{
  echo '```'
  grep -E '^(    )?--- (PASS|FAIL):' "$REPORT_DIR/$STAMP-tests.txt" \
    | sed 's/^ *//; s/--- //' \
    | awk '{gsub(/[()s]/,"",$3); if ($3+0 > 5) print $3, $2}' \
    | sort -rn | head -15 || echo "(none over 5s)"
  echo '```'
} >> "$REPORT"

# --- leaked processes --------------------------------------------------
section "Leaked agents"
{
  echo '```'
  leaked=$(pgrep -f '/T/tmp\..*/ensemble' 2>/dev/null | wc -l | tr -d ' ')
  echo "leaked agent processes: $leaked"
  [ "$leaked" != "0" ] && pgrep -fl '/T/tmp\..*/ensemble' | head -10
  echo '```'
} >> "$REPORT"

# --- verdict -----------------------------------------------------------
total=$(( $(date +%s) - start_epoch ))
{
  echo
  echo "## Verdict"
  echo
  if [ "$sweep_rc" -eq 0 ] && [ "$test_rc" -eq 0 ]; then
    echo "**GREEN.** Sweep and tests both clean."
  else
    echo "**RED.** sweep exit $sweep_rc, tests exit $test_rc. See sections above."
  fi
  echo
  echo "Total wall time: $((total / 60))m $((total % 60))s."
} >> "$REPORT"

cp "$REPORT" "$LATEST"
echo "report: $REPORT"
[ "$sweep_rc" -eq 0 ] && [ "$test_rc" -eq 0 ]
