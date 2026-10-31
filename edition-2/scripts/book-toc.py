#!/usr/bin/env python3
"""Replace the <!-- toc --> marker in an assembled book with a linked table
of contents whose anchors match the ids GitHub gives headings.

Usage: scripts/book-toc.py < assembled.md > book.md

GitHub's anchor rule (github-slugger): take the heading's rendered text,
lowercase it, delete every character that is not a letter, digit, mark,
underscore, hyphen or space, and turn each space into a hyphen. A repeated
slug gets -1, -2, ... in document order, counted over EVERY heading in the
file, not only the ones listed here. So every heading is slugged, in order,
including the Contents heading this script adds; headings inside fenced
code blocks are not headings and are skipped.

The contents list shows level-1 headings (preface, chapters) with their
level-2 sections nested beneath. The book title, which precedes the marker,
is left out.
"""
import re
import sys

MARKER = "<!-- toc -->"
TOC_TITLE = "Contents"

ATX = re.compile(r"^ {0,3}(#{1,6})[ \t]+(.*?)(?:[ \t]+#+)?[ \t]*$")
FENCE = re.compile(r"^ {0,3}(`{3,}|~{3,})")
LINK = re.compile(r"!?\[([^\]]*)\]\([^)]*\)")
SETEXT = re.compile(r"^ {0,3}(=+|-+)[ \t]*$")


def rendered_text(heading):
    """Approximate the text GitHub renders for a heading: link text instead
    of the link, no inline-code backticks. Emphasis markers and other
    punctuation are deleted by slug() anyway."""
    return LINK.sub(r"\1", heading).replace("`", "")


def slug(text):
    s = rendered_text(text).strip().lower()
    s = re.sub(r"[^\w\- ]", "", s)
    return s.replace(" ", "-")


class Slugger:
    """github-slugger's de-duplication, step for step."""

    def __init__(self):
        self.seen = {}

    def __call__(self, text):
        base = slug(text)
        result = base
        while result in self.seen:
            self.seen[base] += 1
            result = f"{base}-{self.seen[base]}"
        self.seen[result] = 0
        return result


def headings(lines):
    """Yield (index, level, text) for every heading outside a code fence.
    Setext headings (a line of === or --- under paragraph text) are refused
    loudly: they would take an anchor this script did not count."""
    fence = None
    prev_text = False
    for i, line in enumerate(lines):
        m = FENCE.match(line)
        if fence:
            if m and m.group(1)[0] == fence[0] and len(m.group(1)) >= len(fence):
                fence = None
            prev_text = False
            continue
        if m:
            fence = m.group(1)
            prev_text = False
            continue
        h = ATX.match(line)
        if h:
            yield i, len(h.group(1)), h.group(2)
            prev_text = False
            continue
        if prev_text and SETEXT.match(line):
            sys.exit(f"book-toc: setext heading at line {i + 1}; use # instead")
        prev_text = bool(line.strip()) and not line.lstrip().startswith(("-", "*", ">", "|", "<"))


def link_text(text):
    return text.replace("[", r"\[").replace("]", r"\]")


def main():
    lines = sys.stdin.read().split("\n")
    marks = [i for i, l in enumerate(lines) if l.strip() == MARKER]
    if len(marks) != 1:
        sys.exit(f"book-toc: expected exactly one {MARKER} line, found {len(marks)}")
    at = marks[0]
    lines[at] = f"# {TOC_TITLE}"

    slugger = Slugger()
    entries = []
    for i, level, text in headings(lines):
        anchor = slugger(text)
        if i > at and level <= 2:
            entries.append((level, text, anchor))

    toc = [f"# {TOC_TITLE}", ""]
    for level, text, anchor in entries:
        indent = "  " * (level - 1)
        toc.append(f"{indent}- [{link_text(text)}](#{anchor})")
    lines[at:at + 1] = toc
    sys.stdout.write("\n".join(lines))


if __name__ == "__main__":
    main()
