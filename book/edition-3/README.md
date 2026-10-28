# Building the second edition

The current release is **Second Edition — Update 2 (October 2026)**, the final
second-edition update before the third edition. Update 2 is the first update
after the original release (implicitly version 1). The third edition will result
from executing this book using Ensemble v2, the implementation built with Codex.

The chapter files are the manuscript source. The assembled Markdown is committed
for reading on GitHub; EPUB and PDF are generated publication artifacts. The
first-edition files and build instructions remain in `edition-1/book/`.

The October 14 source order puts accounting and the fleet before background
memory: Chapters 15–21 lead to memory in 22 and recall in 23. The rebuilt
assembled Markdown, EPUB and PDFs follow this order and have passed independent
artifact review. The current 10.5pt print interior has 496 pages; the reading PDF has 497.

## Prerequisites

Use Python 3 and pandoc. Print production also needs Google Chrome and the
Poppler tools `pdftotext` and `pdfinfo`. The print build and cover-page
conversion use PyMuPDF, including checks of actual PDF font sizes
(`python3 -m pip install pymupdf`).

Run from the repository root:

```sh
make book
make ebook
python3 book/build-print.py
```

Assembly includes the preface, Chapters 0–26, epilogue and author's note. Chapter
25 is the deliberate crossover placeholder. Both front-matter files identify
this as Second Edition — Update 2, October 2026. Keep them and the author's note
consistent when preparing a later release.

| Artifact in `book/` | Purpose |
| --- | --- |
| `the-art-of-building-ai-coding-agents.md` | Assembled manuscript with linked contents |
| `The Art of Building AI Coding Agents.epub` | Ebook |
| `The Art of Building AI Coding Agents - Print Interior.pdf` | Print interior |
| `The Art of Building AI Coding Agents.pdf` | Cover followed by the print interior |

## EPUB navigation and code

The EPUB build removes the manuscript's printed Contents section and omits
pandoc's `--toc`, keeping navigation out of the reading spine. It retains
`--toc-depth=1` and populated `nav.xhtml` and `toc.ncx` navigation. This is Bill's
explicit choice after a previous build displayed two contents pages. Do not
reverse it as a routine cleanup.

The build combines pandoc's default stylesheet with `book/epub.css`; supplying
only the override would discard the base styling. The override keeps highlighted
code spans inline and suppresses their empty anchors to avoid excess line spacing
in ebook readers. Browser rendering is useful evidence, but the final EPUB still
needs inspection in the intended reader.

`book/cover.png` is the default cover. An explicit cover can be supplied to
`scripts/build-ebook.sh`. Check that the EPUB includes it, has chapter-level
navigation, and has no duplicate printed contents or navigation page in its spine.

## Print pagination and cover

`book/build-print.py` generates print HTML from the assembled manuscript and print
front matter, removes duplicate contents, and renders with Chrome **once**. The
CSS reserves a number column in the contents. PyMuPDF then resolves the contents
links and stamps measured page references into that column without reflowing the
book. It also adds footer numbers and PDF page labels: front matter is unnumbered,
and the Preface begins **page 1** on physical page 5. An explicit unnumbered blank
page before the Preface keeps it on the right; Chrome does not reliably honor
`break-before: right`. The build rejects a verso Preface if front-matter layout
changes. A longer future contents section may require adjusting this spacer.

The print CSS supplies mirrored margins, chapter breaks and wrapping code blocks
on a 7 × 10 inch page. Physical page parity—not the reset numbering—controls the
binding margins. The PDF build uses the installed macOS Georgia font for its
stamped references and footers. The intermediate HTML intentionally has no page
numbers; use the completed PDF for the numbered edition.
Default print body text is **10.5 pt**, code 10 pt and tables 11 pt. Inner margins are 0.875 inches;
outer margins are 0.5 inches, with 0.75 inches at top and bottom. Browser body
padding is removed so these page margins control the available text width.
Print code uses black text on white, retaining syntax bold and italic emphasis.
Block quotations retain italics but explicitly use solid black, overriding
Pandoc's gray quote color; links and stamped numbers are black too. The build
checks every extracted text span for black ink. The EPUB keeps its separate
styling.
Long code tokens and table cells wrap rather than making Chrome shrink the
entire document. The build measures the PDF itself and rejects reduced body
or code type: body text must remain 10.5pt and code at least 10pt, allowing small
PDF-coordinate rounding differences. Checking CSS declarations alone misses
Chrome's automatic scaling.
Inspect the resulting PDF as well as the script's result, particularly long code
lines, tables, chapter starts and page footers.

The print build also creates the cover-prefixed PDF, preserving the existing
artwork's proportions. It resolves the interior's named contents links and shifts
their targets by one page. Plain PDF concatenation can discard those named
destinations while leaving apparently clickable links that go nowhere.

### Print sidebars

Headings explicitly beginning `Sidebar:` open print-only callout panels, ending
at the next heading of the same or higher level. `scripts/print-sidebars.lua`
groups the heading and its body; headings and links retain their identifiers.
Ordinary sections such as "Sidebar tabs" and code examples are not treated as
callouts. This filter runs only in the PDF build, not in the EPUB build.

The panels use black borders (a heavier left edge), inset padding, and 11pt bold
headings. Prose stays 10.5pt black and code stays at least 10pt. Long sidebars may
continue across pages, with repeated frame decoration; they are not forced into
unbreakable boxes. The three current sidebars begin on printed pages 16, 19 and
34. Adding this treatment left the interior at 496 physical pages.

Run the print regression tests with:

```sh
python3 -m unittest discover -s book -p 'test_*.py' -v
```

The build verifies all contents references, footer numbers, PDF page labels,
black text and font sizes against the final interior PDF. The cover-prefixed
PDF preserves page labels as well as links. The current interior has 496 physical
pages, ending at printed page 492; the reading PDF has 497 physical pages.

The cover-prefixed PDF should contain exactly one more page than the interior.
Verify both page counts and the 504 × 720 point page size with `pdfinfo`.
This front-cover page is for the reading PDF; paperback and hardcover full covers
must be laid out separately for the chosen binding, trim and final spine thickness.

### KDP hardcover front image

Use `book/ensemble-edition-2-hardcover-front-7x10.png` as a front-image asset in
**KDP Cover Creator** for the 7 × 10 hardcover. It is 2800 × 4000 pixels at 400 DPI,
with the full original artwork inset on an edge-extended background. It is not
a complete hardcover case-wrap PDF. A JPEG copy for uploading is in
`/Users/bill/Downloads/AI Coding Agents - Hardcover Front/`.

Hardcover cover-wrap/hinge allowances do not require different interior margins.
The current 496-page, no-bleed interior is within the 75–550-page hardcover range;
recheck eligibility and margins after future pagination changes. See the
[KDP print publishing guide](../docs/kdp-print-publishing.md) for image provenance,
Cover Creator steps, hardcover safe-area checks and official sources. Final
Cover Creator placement and a physical proof remain unverified.

## Release check

Verify chapter coverage, edition metadata, contents links/page references, embedded
images and representative code/table layout. Check links in the assembled
Markdown and the generated formats; a repository-relative link needs a usable
published destination in an ebook. Keep the current publication record in
[docs/edition-2-publication.md](../docs/edition-2-publication.md), including any
remaining reader or print-preview limitations.

Commit only the intended source and publication files, preserving concurrent
application work. A local build or commit does not publish the book. Bill chooses
the destination and pushes or uploads the reviewed artifacts.
