# Publishing the companion on KDP

The publication title is **The Art of Building AI Coding Agents: A Human Companion**,
by Bill Cox. The series is **The Singularity as it Happened**, **Book 6**.
The approved manuscript is not rewritten by the build.

## Build

From the repository root:

```sh
python3 book/companion/build-publication.py
```

This assembles the preface, Chapters 0–27, and epilogue in that exact order. It
excludes the voice guide, README, restart notes, and editorial records. Outputs
are isolated in `book/companion/build/`; the original book and Edition 3 artifacts
are never overwritten.

The workflow reuses the maintained `book/build-print.py`, `book/print-header.html`,
`book/epub.css`, and the repository's contents/link helpers. Like the original
print workflow, it requires **macOS**, Google Chrome in `/Applications`, Georgia
and Menlo fonts, **Pandoc** (tested with 3.9.0.2), **Poppler** (`pdftotext` and
`pdfinfo`), and **PyMuPDF** available to `python3`.

Individual formats and an optional ebook cover:

```sh
python3 book/companion/build-publication.py --format markdown
python3 book/companion/build-publication.py --format print
python3 book/companion/build-publication.py --format epub --cover /path/to/companion-cover.png
```

The ebook cover is optional for building, not a substitute for choosing the final
KDP cover. The current committed EPUB has **no embedded cover**, pending the new
companion artwork. Do not reuse the original edition's cover without updating its
subtitle and series numbering. `--cover` embeds a front-cover image into the EPUB;
it does **not** produce a paperback wrap or change the interior PDF.

## Files to upload

- **Kindle manuscript:** `build/The Art of Building AI Coding Agents - A Human Companion.epub`
- **Print manuscript:** `build/The Art of Building AI Coding Agents - A Human Companion - Print Interior.pdf`
- **Reading/debug copy:** `build/companion.md` (not a KDP upload).

The PDF is a **6 × 9-inch, no-bleed, black-text interior**, currently **357 pages**.
It has a title page, copyright page, a 12-point dedication to Caitlin Cox,
numbered contents references, and unnumbered front matter. The Preface starts on physical page 7 and printed page 1. Body type
is 11.5-point Georgia; inline code is 11.5-point Menlo. Inside margins are 0.875 in,
outer margins 0.5 in, and top/bottom margins 0.75 in.

The EPUB is reflowable. Its navigation contains the 30 manuscript sections without
a duplicate in-spine contents page. Repository source/evidence links become public
GitHub links; local contents anchors stay inside the book. The build checks EPUB
navigation and internal links, printed font sizes, all PDF contents destinations,
page references, black text, and page labels. It fails rather than silently
omitting a missing chapter.

The PDF and EPUB are committed as requested. Intermediate Markdown/HTML are ignored.
Rebuild and commit the binaries together with any publication-source changes.

## Before publishing

1. Enter the title, author, and **Book 6** series details consistently in KDP.
2. Upload the PDF as a **6 × 9, no-bleed** interior. Select the desired paper type.
3. Use KDP's cover calculator/template with the **actual final PDF page count**
   and selected paper type to assemble the full back/spine/front cover. KDP may
   append a blank page to an odd-length interior; use its processed page count. A front-cover
   PNG alone is not the paperback cover upload. Leave barcode placement to the
   KDP template; do not invent an ISBN or spine width.
4. Inspect the PDF in KDP's Print Previewer and the EPUB in Kindle Previewer,
   including long chapter titles, links, and navigation. Local checks are not
   KDP acceptance. Order a physical proof before approving print publication.

This script builds files only. It does not log into KDP, upload, or publish.

## Regression checks

```sh
python3 -m unittest discover -s book/companion -p 'test_*.py'
```
