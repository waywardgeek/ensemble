#!/bin/bash
# Assemble the book into a single markdown file from its parts.
# Usage: scripts/assemble-book.sh > book/the-self-wielding-agent.md
set -euo pipefail

BOOK_DIR="$(cd "$(dirname "$0")/../book" && pwd)"
OUT=""

add() {
    if [ -n "$OUT" ]; then
        OUT="$OUT

---

"
    fi
    OUT="$OUT$(cat "$1")"
}

add "$BOOK_DIR/title.md"
add "$BOOK_DIR/preface.md"

for ch in "$BOOK_DIR"/chapter-[0-9][0-9].md; do
    [ -f "$ch" ] && add "$ch"
done

printf '%s\n' "$OUT"
