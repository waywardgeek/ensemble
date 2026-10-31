# Edition 2 publication and layout migration

## Hardcover front-cover preparation

A separate front-image asset for the 7 × 10 hardcover is now available at
`book/ensemble-edition-2-hardcover-front-7x10.png` (2800 × 4000 pixels, 400 DPI).
JPEG and PNG delivery copies are in
`/Users/bill/Downloads/AI Coding Agents - Hardcover Front/`.
The original artwork was preserved in full, reduced proportionally and inset
on an extended background to give the lettering additional clearance.

This is an image for **Cover Creator**, not a complete case-wrap PDF. The earlier
paperback bleed assets must not be substituted for a hardcover wrap. See the
[KDP print publishing guide](kdp-print-publishing.md) for upload steps,
provenance, safe-area caveats and official references.

The current interior was inspected, not rebuilt: **496 pages**, all
**504 × 720 points (7 × 10 inches)**. This is within the 75–550-page hardcover
range. Hardcover cover allowances do not require different interior margins.
No manuscript, build script, preface, existing cover or previous publication
artifact was changed for this task. Image dimensions, DPI and text detection
were checked; no visual/KDP preview, upload or physical proof was performed.
Historical page counts below retain their original artifact identity.

## Current release: Second Edition — Update 2

Update 2 is the first update after the original second-edition release (implicitly
version 1), and the final update before the third edition. Both front-matter
files say **Second Edition — Update 2, October 2026**. `book/note-from-bill.md`
identifies the release and explains that the third edition will result from
executing this book using Ensemble v2, built with Codex. It does not claim that
third-edition execution is complete.

This release rebuilt the current manuscript with `make book`, `make ebook`, and
`python3 book/build-print.py`. The preface, Chapters 0–26, epilogue, cover artwork,
and build scripts/styles were preserved byte-for-byte from the pre-build baseline.
Chapter 25 remains the deliberate crossover placeholder. Concurrent application
and third-edition work was left untouched.

Verification evidence: `solutions/evidence/publication-update2/verification.json`
(records source and artifact SHA-256 hashes).

- EPUB: 31 TOC entries, 32 NCX entries, one matching embedded cover, no navigation
  page in the reading spine, and all 1,896 internal XHTML links resolved.
- Print interior: 577 pages, all 504 × 720 points (7 × 10 inches); measured body
  text 12pt and minimum code text 10pt. Contents pagination remained stable.
- Reading PDF: 578 pages, with the cover followed by identical interior text;
  all 142 links preserved, including the 30 contents destinations shifted by
  one page. Every page remains 504 × 720 points.
- Text bounding boxes remain within the page; measured code spans are black.
  These mechanical checks do not substitute for visual layout review.

No new visual PDF review, Kindle/device preview, KDP print preview, physical
proof, or external URL reachability check was performed for this update. The
paperback wraparound cover remains separate from the reading PDF's front cover.
No commit, push, upload, or publication was performed.

The sections below retain the earlier publication history and review evidence;
they are not claims of a new independent or visual review of Update 2.

## October 14: rebuilt reordered edition

Bill authorized the remaining deployment and artifact build after stopping
Ensemble. Regenerated the assembled Markdown, EPUB, print HTML, print interior
and cover-prefixed reading PDF from reviewed source `274d3a49`. The interior is
577 pages and the reading PDF 578, at 7 × 10 inches with measured 12pt body and
minimum 10pt code. The preface, chapter sources, print styling and cover artwork
are unchanged. The chapter order now includes background memory at 22 and
Auto-Recall at 23.

Independent artifact review found that the old cover-concatenation recipe dropped
the PDF's named contents destinations. The print builder now creates the reading
PDF itself, resolving those links before adding the cover-page offset. All 30
contents targets and coordinates match the interior, and 112 other links remain
unchanged. The EPUB's 1,930 internal references resolve; its navigation stays out
of the reading spine. Representative rendered pages and the cover passed review.
The earlier broken reading PDF remains a local negative control.

The [release receipt](../solutions/evidence/background-compaction/deployment-publication.json)
records exact artifacts, checks, deployment and preserved inputs. These are local
publishing files; no upload, remote push or new physical print proof occurred.

## October 14: reordered manuscript inputs

The current chapter files and index now follow the approved dependency order:
old17–23 become15–21, old15/16 become22/23;0–14 and24–26 stay fixed.
`assemble-book.sh` consumes the sorted numbered files, and `book-toc.py`
derives titles and anchors from those inputs. No publication artifact is rebuilt
by this change. The reading-copy counts and hashes below retain their earlier
source identity, and Bill's preface remains unchanged. Revised cumulative
solution exports are a separate reviewed step; see the
[chapter map](edition-2-chapter-map.md) and
[renumbering record](edition-2-notes/codex-compaction-renumber.md).

## Print contrast update

At Bill's request, print code now uses black text on white backgrounds, including
comments and strings. Bold and italic syntax cues, spacing and wrapping remain.
The rebuilt PDFs retain 371 interior pages / 372 with the cover and the same 30
contents destinations. Bill's preface, all manuscript text and the EPUB are unchanged.

Implementation `4e21dfad` and independent review `8217c6d3` establish the bounded
print correction. See the [build and browser evidence](../solutions/evidence/publication-print-light/README.md)
and [independent rendered-page review](edition-2-notes/publication-review.md).
This is a local reading copy; no publication or physical proof is claimed.

## Latest reading copy: Bill's preface

At Bill's request, regenerate the book from his rewritten `book/preface.md`,
preserving it verbatim. Bill retains editorial ownership of the preface; future
author passes must leave it alone unless he explicitly requests an edit.
This supersedes the previous preface draft and its review, without relabeling
those historical records as review of Bill's new text.

The rebuilt interior has 371 pages; the cover-prefixed PDF has 372. All 30
contents page references agree with the new heading pages. EPUB chapter order,
navigation, cover and links pass the existing structural checks. The preface
is unchanged on disk, appears verbatim in the assembled Markdown, and its opening
and closing text are present in the EPUB and rendered PDF. Chapter text and the
11 pre-existing application/archive edits remain unchanged.

Build inputs, commands, artifact hashes and verification are retained in
[the rebuild record](../solutions/evidence/publication-preface-r2/build.json) and
[the checks](../solutions/evidence/publication-preface-r2/full-verification.json).
This is a mechanical regeneration for Bill's next reading, not a new editorial
review or a publication. No push or upload has occurred. The earlier independent
artifact review below describes its original version.

Bill authorized this sequence on October 12, 2026: finish the Gateway chapter,
archive Edition 1, promote Edition 2, and prepare its first publication. The
crossover has begun; Chapter 25 deliberately remains a placeholder. Producing
Edition 3 with Ensemble remains future work.

## Completed physical migration

Promotion commit `564e316f3f8aeac2ec26fe8edfbd8cdf42aa350b` follows accepted
Gateway review `63709086` and checkpoint `c35aab73`. It maps 8,386 tracked paths
and preserves their committed modes and blobs. The resulting tree has 9,020
tracked files, including separate original/shared grading support. All 8,386
moved working-file hashes and 443 protected local-file hashes match the pre-move
capture. Independent reviewer verification confirms those counts and identities.

The migration's per-file shared-index reconciliation collided with a concurrent
status refresh after the relocation commit was written. The coordinator rebuilt
only the changed index entries in one `update-index --index-info` operation,
verified the committed tree against the prepared entries and confirmed an empty
staged diff. No working bytes, commits or tags were reverted. The same 11 prior
modified files remain uncommitted at their new paths: ten active crossover files
and the historical first-edition chapter index. Their acceptance is not implied.

Current entry-point, link and publication checks follow below; physical
preservation alone does not establish that those commands work.

| Former path | Current location |
| --- | --- |
| `agent/` | `edition-1/agent/` |
| First-edition content in `book/` | `edition-1/book/` |
| First-edition `solutions/chNN/` | `edition-1/solutions/chNN/` |
| `solutions/edition-2/main/` | `agent/` |
| `book/edition-2/` | `book/` |
| Other contents of `solutions/edition-2/` | `solutions/` |
| Original `cmd/grade/` | `edition-1/cmd/grade/` |

Preserve the original grading module and its support under `edition-1/`, with
the original module path, so it can run from that directory. The current revised
graders still use `internal/grade` and the fake providers. Retain that shared
support at the top level as well; moving the old command does not justify a
rewrite of the grading library. Keep `cmd/grade-edition2` as the explicit current
command and make the top-level Makefile select it.

Preserve original build scripts, Makefile and README in the archive. Shared craft
guides, build assets and book tooling also remain available for the current book.
Archive old chapter-specific maintenance scripts; distinguish retired Edition 2
attempt scripts from the current exporter rather than silently making obsolete
checks look current.

The move preserves tracked blob identities for the original edition and frozen
solutions. Bill's dirty application files and local untracked material move with
their directories, without becoming accepted source or being staged. Historical
manifests, scores and receipts retain their source paths and revisions. This map
explains their new locations. There are no symlinks or new chapter numbers.

## Assignments and final checks

- Coordinator: physical moves, preservation check, current README and agent
  instructions, chapter paths, external launcher, final epilogue and assembly.
- Author: preface, front matter, author's note and narrow Chapter 0 corrections;
  preserve the documented voice and distinguish completed work from crossover.
- Coder: current grading defaults, affected test paths and publication tooling.
- Reviewer: independent Gateway acceptance, then relocated entry points,
  preservation and publication prose/artifact checks.

Verify relocated default grading and original archive grading, affected module
checks, runnable example paths and local links. Do not repeat paid demonstrations
for a path-only move. Build the assembled Markdown, EPUB, print interior and
cover-prefixed PDF. Preserve the established EPUB navigation behavior. Inspect
chapter coverage, edition metadata, page references, code layout and images.

The new epilogue is the coordinator's own first-person account from Codex, written
after the manuscript is complete. The archived epilogue's narrator running inside
Ensemble is a different contribution and stays intact. Local publication files
and commits are preparation; no external upload or push has occurred.

## Current preparation status

The external launcher now selects `agent/skills`; its existing binary and live
session were preserved. A normal session restart is needed to pick up that path.
Current README, chapter index and governing-document headers identify the promoted
layout. Historical literal paths and source hashes remain unchanged; relative
Markdown links in the notes now resolve to relocated current or archived artifacts.

Independent review accepts the new preface, front matter, author note and narrow
Chapter 0 corrections. The detailed [publication review](edition-2-notes/publication-review.md)
tracks the remaining tooling, chapter-path, epilogue and artifact gates.

## Uncommitted application work: provenance correction

The coordinator initially attributed the 11 pre-existing modified files to Bill's
crossover work without checking who wrote them. Bill reported that CodeRhapsody
had committed its own work and disclaimed these changes. Read-only inspection of
the separate live Ensemble session's event log established the source of the ten
application edits: Bill requested approximately 5× speech with Eloquence
calibration and ordinary chat completion for the human-facing root, reserving
`submit` for children. The session records the requests and corresponding edit
operations on October 12 UTC (October 11 locally). The existing untracked
`ch25-crossover.md` records the speed work and its validation limits.

These edits remain uncommitted, unchanged and outside the accepted snapshots;
publication preparation neither accepts nor discards them. The two added roadmap
items in `edition-1/book/chapters.md` also predate the move, but their writer has
not been established. No private conversation log was copied into the repository.

## Local reading copy, October 12

Bill will read the book before publication. No push, upload or copy to an external
publishing destination is authorized by this preparation. Crossover Chapter 25
remains reserved.

Artifact checkpoint `ca3f4764` contains assembled Markdown, EPUB, print HTML, the
370-page print interior and the 371-page cover-prefixed PDF. Every PDF page is
6 × 9 inches. All 30 printed contents references match the actual heading pages.
The EPUB includes all chapters in order, its cover and populated chapter-level
navigation, with no navigation page in the reading spine or duplicate printed
Contents. Repository links distinguish files from directories. The final artifact
hashes and measured checks are in [the verification record](../solutions/evidence/publication-tooling/full-verification.json).
The coordinator independently matched all five artifact hashes.

Current default grades for Chapters 1 and 26 and archived original Chapter 1 each
score 100/100 in new relocation runs. The real nested-module fakevendor loop,
maintained tooling build/vet, selected regressions and exporter tests pass. Current
CLI and GUI build to temporary binaries; this establishes buildability, not
acceptance of the separate uncommitted crossover changes. Full preface/chapter
prose lint exits zero with soft warnings retained. Reader-entry local links pass.
See [the tooling report](../solutions/evidence/publication-tooling/README.md) for
exact scope and the distinction between compilation and executed tests.

The epilogue is Astra's own first-person account from Codex. Independent review
resolved one attribution correction: the student raised the memory-trigger
contradiction, and Bill settled the intended behavior. The final text preserves
failed attempts, actual results and the unfinished crossover. Front matter and
all mechanically relocated chapter paths are independently accepted.

Independent rendered-page review accepts the local reading copy: cover, contents,
front matter, representative chapter openings, multi-page tables, long command/JSON
blocks, placeholder25, epilogue and author note are readable. All 370 interior
footers and all 30 contents references were checked. An extracted EPUB Gateway
page also renders readably in Chrome. The [independent review](edition-2-notes/publication-review.md)
records the precise scope. Kindle Previewer, KDP print preview, physical proof
and Bill's editorial read-through have not been performed.
These local artifacts are reading copies, not a published edition.

## October 13: correct print scaling after publication

Bill reports the book is published and requests a second-edition readability
fix. Comparing the assembled manuscripts showed 158,623 whitespace-separated
words in Edition 2 versus 129,538 in Edition 1. The 372-page second-edition
interior nevertheless printed its declared 11pt body at 7.33pt and its 9pt
code at 6pt. Long code tokens and table cells overflowed; Chrome scaled the
whole document down. The archived first edition also shrinks, but less.
Earlier checks of page dimensions and representative screenshots missed this.

The coordinator corrected print CSS to wrap long tokens and override pandoc's
inline-block code spans and hanging indent. At the printable width, the full
HTML now has no horizontal overflow. The PDF measures 10.99pt body (rounding
from 11pt) and minimum 9pt code. The print build now checks those actual sizes
with PyMuPDF; running that check against the previous PDF correctly rejects
its 7.33pt/6pt type. No manuscript or first-edition content was changed.

The rebuilt interior has 697 pages; the cover-prefixed PDF has 698. All pages
remain 6 x 9 inches, all 30 contents references match actual chapter pages,
and no extracted word extends outside its page. Visual inspection covered the
previously overflowing JSON example, code, tables and a chapter opening.
The preface, chapter sources, assembled Markdown, cover image and EPUB retain
their pre-fix hashes. These are local replacement PDFs; no upload or change to
the published listing was made. A physical print proof remains unperformed,
and a paperback cover's spine must match the new page count before reprinting.

## October 13: larger trim and type

Bill selected the recommended 7 × 10 inch format with 12pt body text and 10pt
code. The print stylesheet now uses those sizes, 11pt tables and larger section
headings. Mirrored inner margins are 0.875 inches, outer margins 0.5 inches,
and top/bottom margins 0.75 inches. Removing pandoc's browser body padding lets
the page margins determine the text width. The overflow fix remains in place;
the build's measured-type check now requires 12pt/10pt.

The rebuilt interior has 562 pages; the reading PDF has 563, including its
cover. Every page measures 504 × 720 points. The existing cover artwork is
centered proportionally on the larger reading-PDF page rather than stretched;
the source image is unchanged. All 30 printed contents references match actual
chapter pages. Actual type measures 12pt body and minimum 10pt code, the HTML
has no horizontal overflow at the printable width, and no extracted word falls
outside a PDF page. Representative table, code and chapter-opening pages were
visually inspected. Manuscript, preface, chapter sources and EPUB are unchanged.
No external publication action was taken; the physical paperback cover still
needs a separate trim/spine layout and a print proof.

## October 13: full artifact rebuild

At Bill's request, rebuilt assembled Markdown, EPUB, print HTML, print interior
and cover-prefixed PDF from the current sources, including Chapter 25's link
to the live crossover record. The PDFs remain 562/563 pages at 7 × 10 inches,
with measured 12pt body and minimum 10pt code; all 30 print contents references
match. EPUB contains Chapters 0–26 and all closing sections, its cover and code
stylesheet, with no navigation document in the spine or duplicate Contents
section. All 1,895 internal links/resource references resolve. Its navigation
has 31 TOC entries (30 book sections plus front matter); the build script's
33-link count also includes two landmark links. Rebuild only: no source prose
edits, external upload or paperback wraparound-cover generation.


## October 17: 10.5pt print body text

At Bill's request, reduced the current print body from 11pt to 10.5pt to
reduce the book's bulk. Code stays 10pt, tables 11pt, and line-height 1.4.
The 7 × 10 inch trim and mirrored margins are unchanged: 0.875 inch inside,
0.5 inch outside, and 0.75 inch top/bottom. The measured-font build check
now requires 10.5pt body and at least 10pt code, allowing rounding tolerance.

Rebuilt print HTML, print interior and cover-prefixed reading PDF. Interior
pagination fell from 519 to 496 pages; the reading PDF has 497 pages.
The build measured 10.5pt body and minimum 10pt code, matched all 30 contents
headings, and verified stable pagination across both passes. All PDF pages
measure 504 × 720 points. Extracted text stays inside even the stricter
501–700-page no-bleed minimum margins (0.75 inch inside; 0.25 inch other
edges), with a 0.5pt coordinate tolerance. This is a text-bound check, not
a visual inspection of graphics or KDP Preview.

Hashes confirm that manuscript sources, front matter, assembled Markdown,
EPUB and cover artwork are unchanged. No external upload was performed.
KDP Preview and a physical proof of this rebuild remain to be checked; the
paperback wraparound cover needs a spine sized for the new page count.


## Print quotation contrast and Preface-based numbering

The printed proof exposed gray quotations: the previous PDF contained 4,724
italic characters in #606060 inherited from Pandoc. Print blockquotes and their
descendants now explicitly use black while retaining italics and the existing
10.5pt size. Print links and pagination text are also black. EPUB/manuscript
content and styling were not changed by this print-only adjustment.

Replaced the two-render HTML TOC injection with a single Chrome render followed
by PyMuPDF stamping into a reserved TOC column. All 30 entries use resolved PDF
links checked against independently located chapter headings. Footers and PDF
labels start at 1 on the Preface, physical interior page 5. Earlier pages have
no footer numbers. The old blank page after Contents is replaced by an explicit
blank before the Preface, preserving recto start and all body-page positions.
Chrome did not enforce break-before:right in testing; the build verifies parity
and fails rather than silently accept an incorrect gutter/number arrangement.

The rebuilt interior remains 496 physical pages (last printed number 492); the
reading PDF remains 497. All 492 body pages have identical extracted text on
identical physical pages to the previous build. Verified every reading-PDF link
(144 total) and page label against the interior with the cover offset. Build
checks cover all text colors, 10.5pt body / minimum 10pt code, TOC references,
footers and page labels. Seven regression tests include a wrapped named-link
PDF fixture, gray-text rejection, and rejection of a verso Preface.

An initial TOC verification rectangle was too tall and captured the preceding
line; it was corrected. Removed optional whole-document font subsetting after
MuPDF warnings; existing Chrome fonts are left intact and Georgia pagination
font is embedded. This was programmatically verified, not a new physical proof
or a KDP Preview approval. Nothing was uploaded.


## Distinct print sidebar panels

The three explicit `Sidebar:` passages previously rendered as normal h3 sections.
Added a print-only Pandoc filter to wrap each heading and its body through the
next heading at the same or higher level. Their IDs/links remain intact; regular
"Sidebar tabs" sections are not matched. Panels use black borders with a heavier
left edge, inset padding, and 11pt bold headings; body and code retain 10.5pt and
10pt respectively. Panels may break across pages rather than force large gaps.
No manuscript source or EPUB was changed.

Rebuilt both PDFs: 496 interior / 497 reading pages, unchanged totals. Sidebar
headings occur on printed pages 16, 19 and 34. Verified bold black heading sizes
and enclosing black frame paths in the PDF. Full body text matches the previous
PDF in reading order after excluding footers and whitespace; the default PDF
object-order extraction differs around a split code block, so that comparison
must use sort=True. All 11 tests pass, including four new Pandoc grouping tests;
existing font, all-black-text, TOC, footer, label and reading-link checks pass.
