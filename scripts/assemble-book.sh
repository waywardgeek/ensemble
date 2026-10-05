#!/bin/bash
# Assemble the book into a single markdown file from its parts.
# Usage: scripts/assemble-book.sh > book/the-art-of-building-ai-coding-agents.md
#
# The output IS COMMITTED ON PURPOSE. Do not treat it as a disposable build
# artifact and do not purge it to save repository space: README.md links to
# it so the partial book is readable directly on the GitHub repo, and it is
# published as the book is written.
#
# This has already gone wrong once. A git filter-repo pass aimed at large
# binaries swept the assembled book up with them, judging files by size
# rather than by value. Recovery was luck: an untracked working-tree copy.
# Note that filter-repo rewrites every commit, so a purged path leaves NO
# deletion commit behind and "git log -- <path>" returns empty. That empty
# history is the signature of a purge, not proof the file was generated
# scratch output. Being reproducible does not make it disposable.
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

add "$BOOK_DIR/epilogue.md"
add "$BOOK_DIR/note-from-bill.md"

printf '%s\n' "$OUT" | python3 "$(dirname "$0")/book-toc.py"
