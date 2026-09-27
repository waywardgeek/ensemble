#!/bin/bash
# Run the ch17 grader against the CURRENT (possibly mutated) agent tree and
# record which checks failed, by ID.
#
# Usage: ./mutate.sh <mutant-name>
#
# Never run two graders at once: a concurrent sweep produced 11/100 for a
# grader that scores 100/100 alone. Any number measured alongside other work
# is void. This script is deliberately serial and deliberately explicit.
set -u
cd /Users/bill/projects/ensemble || exit 1
NAME="$1"
DIR=docs/ch17-mutation-audit
OUT="$DIR/$NAME.txt"

# Compile the agent first. A mutant that does not build looks devastatingly
# effective and proves nothing at all.
if ! (cd agent && go build ./... 2>&1); then
  echo "BUILD FAILED -- mutant is invalid, not effective"
  exit 2
fi

go run ./cmd/grade -ch 17 ./agent >"$OUT" 2>&1

IDS=(bm25-indexes bm25-scores bm25-stop-words chunking-splits per-source-quota \
     judge-filters judge-fallback recall-is-a-spoke recall-is-own-kind injection-capped)

echo "=== $NAME ==="
grep -E '^\s*\[(PASS|FAIL)\]' "$OUT" | awk -v ids="${IDS[*]}" '
  BEGIN { split(ids, a, " ") }
  { st = ($0 ~ /PASS/) ? "pass" : "FAIL"; printf "%-22s %s\n", a[++n], st;
    if (st == "FAIL") f = f " " a[n] }
  END { print "---"; print "failed:" (f == "" ? " NONE" : f) }'
grep -E 'TOTAL|/100' "$OUT" | tail -2
