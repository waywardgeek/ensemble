#!/bin/bash
# Grade sweep: run every chapter grader at its canonical target and
# record the score. Targets match the Makefile's grade* rules.
# Usage: bash scripts/gradesweep.sh OUTFILE
cd "$(dirname "$0")/.."
OUT="${1:-/tmp/gradesweep.txt}"
: > "$OUT"
run() {
  CH="$1"; DIR="$2"
  if [ ! -d "$DIR" ]; then echo "ch$CH $DIR MISSING" >> "$OUT"; return; fi
  S=$(go run ./cmd/grade -ch "$CH" "$DIR" 2>/dev/null | grep -Eo 'score: [0-9]+/100' | tail -1)
  echo "ch$CH $DIR ${S:-NOSCORE}" >> "$OUT"
  echo "ch$CH $DIR ${S:-NOSCORE}"
}
# Canonical targets (Makefile).
run 1  ./solutions/ch01
run 2  ./solutions/ch02
run 3  ./solutions/ch03
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
# Per-chapter reference trees that also get graded.
run 10 ./solutions/ch10
run 11 ./solutions/ch11
run 12 ./solutions/ch12
run 13 ./solutions/ch13
run 14 ./solutions/ch14
echo "SWEEP-DONE" >> "$OUT"
