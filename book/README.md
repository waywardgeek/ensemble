# Edition 3 — C# inside Ensemble

This edition includes the completed Chapter 25, Voice and Images, and the
current Chapter 30 crossover requirements, alongside Gateway, Goals, and the
optional Codex-backed path. Bill directed the architecture; Cadence
implemented, exercised and reviewed the C# work personally inside Ensemble.

- [Read the assembled book](the-art-of-building-ai-coding-agents.md).
- [Chapter guide](chapters.md), [preface](preface.md), and [epilogue](epilogue.md).
- [Current C# source](../agent) and source-only chapter
  endpoints under [solutions](../solutions).
- [Execution and evidence record](../docs/edition-3-csharp-progress.md) and
  [procedure](../docs/edition-3-csharp-procedure.md).

The implementation uses C# 14 / .NET 10. Follow the current source README and
chapter-specific commands; the repository root now builds the C# application.
Bill accepted the working V3 crossover and authorized this top-level promotion.
Historical foundation checks alone were not that human acceptance. The optional Codex implementation was
completed after Gateway and Goals, and its historical `ch24` endpoint intentionally
contains that foundation. Earlier endpoints are not renamed or overwritten.

The last release artifacts (assembled Markdown, EPUB, print interior, and covered
reading PDF) were rebuilt
from the current manuscript after the former Chapter 26 passed all 100 external media points.
The release checks cover EPUB navigation and local resources, print typography,
contents destinations, page labels, and reading-copy links. They are not a new
whole-book execution grade or KDP preview approval.
**Working manuscript:** the current 00–30 ordering and design outlines are newer
than those release artifacts. The PDFs/EPUB have not been rebuilt for this
renumbering. Connections, Agenda, and Puffin are not completed chapters.

The companion publication is available under [companion/](companion/README.md).

## Sequential manuscript numbers and historical endpoints

Only Edition 3 manuscript numbers changed. Do not renumber immutable source
endpoints, archived evidence, or executable grader flags to match the book.

| Current chapter | Topic | Original chapter / preserved endpoint number |
|---|---|---|
| 00–16 | Earlier foundation | Unchanged |
| 17 | Web and status | 18 / `ch18` |
| 18 | Sandbox | 19 / `ch19` |
| 19 | Ensemble | 20 / `ch20` |
| 20 | Managing the swarm | 21 / `ch21` |
| 21 | Memory | 22 / `ch22` |
| 22 | Recall | 23 / `ch23` |
| 23 | Codex/cache | 24 / `ch24` |
| 24 | Goals | 27 / `ch27` |
| 25 | [Voice and Images](chapter-25.md) | External media grader remains `-ch 26`; scopes pass 100/100 with documented limits. |
| 26 | Gateway | 26 / `ch26`; [complete-state persistence evidence](../docs/edition-3-notes/gateway-complete-state.md) |
| 27 | Connections | Design-stage outline; no implemented endpoint or grade |
| 28 | Agenda | Design-stage outline; no implemented endpoint or grade |
| 29 | Puffin | Design-stage outline; no implemented endpoint or grade |
| 30 | Crossover | Active source in `agent/`; no numbered snapshot |

The two former historical crossover chapters were removed, not merged into
unrelated chapters. Earlier editions and their chapter numbering are unchanged.

## What completion means

Chapter 25 contains its TL;DR, In Plain Words, implementation contracts, and
externally graded exercises. Its four media scopes pass 100/100. The native
host-screen scope verifies default and explicit primary-display captures through
delayed Jobs and byte-identical new-turn replay without recapture. This does not
establish all-platform, native-microphone, or every-phone reliability. These
results are distinct from the historical `-ch 26` Gateway grader.

Each implemented chapter has reviewed teaching source and scoped verification.
Some original grader rows do not transfer to C#; ungraded rows remain explicitly
ungraded. Local fakes, browser exercises and interoperability are not universal
provider, platform, security-audit or human-accessibility acceptance.

The joint live exercise verified public OAuth login, discovery and inference,
with zero cache reuse in three public requests. The integrated C# Codex path then
completed three identical-body requests with cached input of 3,456, zero and
3,456 tokens out of 3,641 each. This is two hits and one miss, not guaranteed
savings or a billing measurement. No further live exercise is needed for this
book boundary. See Chapter 23 for the separate earlier probe and its unknown
attempt. Live renewal/revocation, GUI inference and audible/human accessibility
remain outside these results.

## Build locally

Existing tools: Pandoc, Python 3 with PyMuPDF, Poppler (`pdftotext` and `pdfinfo`),
and Chrome at the path configured in `build-print.py`. No installation is part
of this recipe. From the repository root:

```sh
export BOOK_DIR="$PWD/book"
export BOOK_OUTPUT_DIR="$BOOK_DIR"
export PUBLICATION_DIRECTORY=book
scripts/assemble-book.sh > "$BOOK_OUTPUT_DIR/the-art-of-building-ai-coding-agents.md"
scripts/build-ebook.sh
python3 "$BOOK_DIR/build-print.py"
```

To isolate generated output, set `BOOK_OUTPUT_DIR` to an existing owned build
directory instead. The assembler writes stdout. The EPUB and print builder use
the selected assembled file. The shared scripts now default to the promoted Edition 3 book when
these variables are absent. Archived Edition 2 scripts retain their old defaults.
Chrome rendering uses a fresh temporary profile, never the running user profile.
If Chrome's command-line PDF renderer fails during teardown, use the optional
controlled-lifecycle backend with an existing Playwright installation:

```sh
export BOOK_NODE=/absolute/path/to/node
export BOOK_PLAYWRIGHT_MODULE=/absolute/path/to/node_modules/playwright
python3 "$BOOK_DIR/build-print.py"
```

This backend installs nothing, uses an isolated browser, waits for fonts, prints,
and closes Chrome explicitly. A failed renderer still refuses the build. Both
backends run the same final PDF checks. This release used the controlled backend
because the command-line renderer wrote a PDF but exited nonzero on teardown.


Outputs are the assembled Markdown, EPUB, coverless **Print Interior.pdf**, and
**The Art of Building AI Coding Agents.pdf** with the cover for free reading.
`interior-print.html` is the intermediate print input. Existing cover artwork is
reused as requested. The current coverless interior is **520 pages at 7 × 10
inches**; the covered reading PDF is **521 pages**. The reading PDF is for the
website, not the KDP interior upload. A print cover wrap is separate and must be
sized for the final interior and selected paper.

The build checks contents destinations, printed references, page labels, black
print text, embedded fonts, readable body/code type, and EPUB navigation and
local resources. Review the files in KDP's preview tools before approving a
publication. No upload or store-listing change is performed by these scripts.
Source links target the public repository and become available there when Bill
publishes the corresponding source changes.

Regression checks:

```sh
python3 -m unittest discover -s book -p 'test_*.py'
```
