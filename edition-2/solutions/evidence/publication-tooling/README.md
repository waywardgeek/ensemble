# Layout migration: active tooling validation

These are new relocation checks, not rewritten historical chapter results.
The migration base is `564e316f3f8aeac2ec26fe8edfbd8cdf42aa350b`.
The mandatory `.agents/skills/ensemble-coding/SKILL.md` was read in full;
workflow/carryover remain active from the complete reload.

Current default Chapter 1 and Chapter 26 grades are 100/100; archived original
Chapter 1 also scores 100/100. Commands, exits and whole-command times are in
`commands.json`. The maintained tooling packages build and vet. Their test
packages compile, with the actual Chapter 1 header regression selected; this is
not a claim that every chapter suite ran again. All active literal snapshot paths
exist after relocation. No archived or frozen source was edited.

The fakevendor launcher now sets the child cwd to its selected Go module and
runs `go run .`; the actual Chapter 3 loop completes three local HTTP requests,
two tool results and normal EOF. No paid provider or personal credential is used.
The current exporter defaults to committed `agent/`, with explicit historical
`--source-prefix solutions/edition-2/main` support. Four tests exercise real Git
repositories, preserving committed bytes/modes/ignored receipts, refusing existing
outputs and checking historical-prefix provenance.

The root Makefile addresses maintained `cmd`/`internal` packages, avoiding the
mixed-package historical fixtures in `scripts/edition2` and the separate runtime
modules. Current grade targets use Edition 2 numbers. Historical mutation scripts
are no longer active make targets and remain preserved with their archived edition.
The scripts README labels retired acceptance tools honestly instead of making
their obsolete commands look current.

A tiny real publication built from an unrelated temporary directory exercises
Pandoc and Chrome: zero figure references plus the repository cover, both populated
EPUB navigation documents, chapter-level navigation, no nav in the spine and no
printed Contents. The link-only Lua filter sends local repository links to public
GitHub URLs, preserving external URLs and internal anchors; images remain bundled.
Print resolves inputs from its own script directory and maps its heading to a
stable page number. Tool errors and continued pagination changes now fail rather
than claiming a successful print build. The failed-convergence branch was inspected,
not simulated with mocks. This fixture is not the full publication build.

`source.json` binds the accepted tooling stage at `3e6a05b8`. The final follow-up
changes only directory links to GitHub `tree/main` rather than `blob/main`.
`final-source.json` binds final tooling, manuscript inputs and the five generated
artifacts. No application runtime, user process, original result or frozen snapshot
is modified by this work.

## Final local book build

All four actual build commands pass in `full-build.json`: assembled Markdown,
EPUB, Chrome print interior, and cover-prefixed PDF. These include the corrected
`c842bbda` epilogue attribution and directory-link filter. `full-build.py` retains
the invocation, including the existing Pillow cover format-conversion recipe.
The earlier successful build caught a concurrently corrected epilogue during
verification; `initial-*` preserves that attempt. `pre-directory-*` retains the
build before the directory-link correction. Neither is relabeled final.

`full-verification.py` and its JSON receipt verify all 30 complete source sections,
27 chapters in order (0 through 26), populated EPUB navigation and NCX, no
navigation document in the spine, no printed EPUB Contents, one embedded cover
and zero figure references. The generated title page is separate from the
manuscript's title heading: there are 31 content TOC entries, 32 NCX entries and
32 body H1 headings. Initial verifier count assumptions and their corrections are
retained in `verification-assumptions.txt`; these were validator assumptions, not
missing-book-content defects. All 61 local repository links become public GitHub
links; the epilogue directory uses `tree/main`. Source evidence is unchanged.

The print interior has 370 pages and the cover-prefixed PDF has 371. Every page
is 432 by 648 points (6 by 9 inches). Pagination converges and all 30 printed TOC
page numbers match the actual heading pages. The coder inspected rendered TOC
page 1 and Gateway code/table page 360; text and footers remain within the page.
Independent review additionally inspects representative actual rendered pages.
These checks are not a Kindle/KDP preview, exhaustive visual inspection of every
page, or approval to publish. Bill will read the local book before publication.
Nothing was pushed, uploaded or copied to an external publication destination.

`readme-builds.json` records successful core and optional GUI builds to temporary
outputs. This establishes current dirty-tree buildability, not acceptance of
Bill's unfinished crossover edits. The actual CLI parser accepts `chat` before
`--skills-dir` because it collects positional and option arguments across its
whole argument list. No inference was needed for that source check.
