#!/usr/bin/env bash
# Chapter 14 mutation audit (P9).
#
# A check that cannot fail is not a check. Each mutant below deletes exactly ONE
# protected behaviour from the reference solution, then runs the real grader
# against the mutated tree. Every mutant must produce a specific, predicted
# failure; a mutant that nothing catches is a stale mutant, not a passing grader.
#
# This is not part of the normal test run. It is run when the ch14 grader or the
# speech pipeline changes, to confirm the checks still bite.
#
# Usage:  bash scripts/ch14-mutations.sh        (from the repo root)
# Output: ch14-mutation-results.txt, plus a live log on stdout.
#
# Mutations are applied IN PLACE and reverted with git checkout, including on
# interrupt. The tree must be clean for the three files below before starting.

set -u

REPO="$(pwd)"
OUT="${REPO}/ch14-mutation-results.txt"
TTS=agent/web/gui/tts.js
ART=agent/web/gui/artifact-scroll.js
GUI=agent/web/gui/gui.js

restore() { git checkout -- "$TTS" "$ART" "$GUI" 2>/dev/null; }
trap restore EXIT INT TERM

if ! git diff --quiet -- "$TTS" "$ART" "$GUI"; then
  echo "refusing to run: $TTS, $ART or $GUI has uncommitted changes" >&2
  exit 1
fi

: > "$OUT"

# run_mutant <name> <predicted failing check> <file> <perl expression>
run_mutant() {
  local name="$1" expect="$2" file="$3" expr="$4"

  perl -0777 -pi -e "$expr" "$file"
  if ! grep -q "MUTANT" "$file"; then
    { echo "── ${name}"; echo "   VERDICT : *** PATCH DID NOT APPLY — pattern is stale ***"; echo; } | tee -a "$OUT"
    restore; return
  fi

  local raw failed score
  raw="$(go run ./cmd/grade -ch 14 ./agent 2>&1)"
  score="$(printf '%s\n' "$raw" | grep -Eo 'score: [0-9]+/[0-9]+' | tail -1 | sed 's/score: //')"
  # The report prints each check's TITLE, not its id. Every failing check emits a
  # detail line beginning "<id>: ", and a passing check emits no details at all,
  # so collecting those prefixes names the failing set exactly.
  failed="$(printf '%s\n' "$raw" | grep -Eo '(tts-[a-z-]+|ch13-parity):' | sed 's/:$//' | sort -u | tr '\n' ' ')"
  failed="${failed% }"

  {
    echo "── ${name}"
    echo "   file    : ${file}"
    echo "   predicted failure : ${expect}"
    echo "   score   : ${score}"
    echo "   failed  : ${failed:-<none>}"
    if [ -z "$failed" ]; then
      echo "   VERDICT : *** STALE MUTANT — no check caught this deletion ***"
    elif [ "$failed" = "$expect" ]; then
      echo "   VERDICT : caught, exactly the predicted check"
    else
      echo "   VERDICT : caught (${failed}); wider than predicted, see note"
    fi
    echo
  } | tee -a "$OUT"

  restore
}

{ echo "Chapter 14 mutation audit"; echo "run: $(date -u '+%Y-%m-%dT%H:%M:%SZ')"; echo; } | tee -a "$OUT"

baseline="$(go run ./cmd/grade -ch 14 ./agent 2>&1 | grep -Eo '[0-9]+/[0-9]+' | tail -1)"
{ echo "── BASELINE (unmutated reference)"; echo "   score   : ${baseline}"; echo; } | tee -a "$OUT"

# 1. The buffer goes. Fragments enqueue raw, so a word split across chunks is
#    spoken in pieces. Everything downstream of the buffer suffers too, which is
#    why the failing set is wider than the one predicted check.
run_mutant "M1 buffer removed" "tts-buffers-fragments" "$TTS" \
  's/(this\._buffer \+= text;)/this._enqueue(text); this._processQueue(); return; \/\/ MUTANT/'

# 2. Fence resolution stops running against the whole buffer. This is the real
#    pre-fix bug (a959042^): a fence shattered by the splitter matches no fence
#    pattern, and the listener hears backticks.
run_mutant "M2 fence resolution neutered" "tts-filters-markup" "$TTS" \
  's/(  _resolveFences\(s\) \{\n)/$1    return [s, ""]; \/\/ MUTANT\n/'

# 3. A lone newline becomes a phrase boundary again, so a wrapped sentence is
#    read as two utterances with a pause in the middle.
run_mutant "M3 newline treated as a boundary" "tts-boundaries" "$TTS" \
  "s/\.replace\(\/\\\\n\/g, ' '\)/.replace(\/\\\\n\/g, '\\\\u0001') \/* MUTANT *\//"

# 4. flush() stops speaking the remainder, so a response ending in a colon is
#    never heard at all.
run_mutant "M4 flush drops the remainder" "tts-boundaries" "$TTS" \
  "s/(  flush\(\) \{\n)/\$1    this._buffer = ''; return; \/\/ MUTANT\n/"

# 5. Every capital gets a space in front of it. This passes camelCase and
#    HTTPServer, and spells ALL CAPS out letter by letter.
run_mutant "M5 ordinary-word guard removed" "tts-expands-identifiers" "$TTS" \
  "s/(  _expandIdentifier\(tok\) \{\n)/\$1    return tok.replace(\/_\/g, ' ').split(\/(?=[A-Z])\/).join(' ').replace(\/\\\\s+\/g, ' ').trim(); \/\/ MUTANT\n/"

# 6. The accumulated guard always passes: a part that already streamed is spoken
#    a second time when it finalises. The doubling direction.
run_mutant "M6 doubling (guard always true)" "tts-speaks-unstreamed" "$ART" \
  's/!this\.accumulated\.has\(id\)/true \/* MUTANT *\//'

# 7. The guard never passes: a part that arrives whole is never spoken. The
#    silence direction, which is the defect that shipped.
run_mutant "M7 silence (guard always false)" "tts-speaks-unstreamed" "$ART" \
  's/!this\.accumulated\.has\(id\)/false \/* MUTANT *\//'

# 8. The gate loses its edge comparison and sends on every call.
run_mutant "M8 gate edge comparison removed" "tts-gate-both-causes" "$GUI" \
  's/if \(blocked === gatePaused\) return;/\/* MUTANT *\//'

echo "results written to $OUT"
