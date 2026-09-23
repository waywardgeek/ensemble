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

# Front matter in Chicago order: title page, copyright page, dedication,
# contents, preface.
add "$BOOK_DIR/title.md"
add "$BOOK_DIR/copyright.md"
add "$BOOK_DIR/dedication.md"
# book-toc.py replaces this marker with a linked table of contents.
OUT="$OUT

---

<!-- toc -->"
add "$BOOK_DIR/preface.md"

for ch in "$BOOK_DIR"/chapter-[0-9][0-9].md; do
    [ -f "$ch" ] && add "$ch"
done

printf '%s\n' "$OUT" | python3 "$(dirname "$0")/book-toc.py"
