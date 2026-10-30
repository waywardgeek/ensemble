# Edition 3 — C# inside Ensemble

This edition is locally complete through Chapter 25, Goals, including Gateway
and the optional Codex-backed path. Bill directed the architecture; Cadence
implemented, exercised and reviewed the C# work personally inside Ensemble.

- [Read the assembled book](the-art-of-building-ai-coding-agents.md).
- [Chapter guide](chapters.md), [preface](preface.md), and [epilogue](epilogue.md).
- [Current C# source](../../solutions/edition-3/main/) and source-only chapter
  endpoints under [solutions/edition-3](../../solutions/edition-3/).
- [Execution and evidence record](../../docs/edition-3-csharp-progress.md) and
  [procedure](../../docs/edition-3-csharp-procedure.md).

The implementation uses C# 14 / .NET 10. Follow the current source README and
chapter-specific commands, not the repository root's Go build instructions.
The final crossover exercise is now active; local foundation verification is not
human acceptance of the replacement. The optional Codex implementation was
completed after Gateway and Goals, and its historical `ch24` endpoint intentionally
contains that foundation. Earlier endpoints are not renamed or overwritten.

The generated Markdown, EPUB, and PDFs are the previous build until regenerated
from this revised manuscript. They do not yet reflect the crossover rewrite or
sequential renumbering.

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
| 24 | Gateway | 26 / `ch26` |
| 25 | Goals | 27 / `ch27` |
| 26 | Crossover | New assignment; no accepted endpoint yet |

The two former historical crossover chapters were removed, not merged into
unrelated chapters. Earlier editions and their chapter numbering are unchanged.

## What completion means

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
export BOOK_DIR="$PWD/book/edition-3"
export BOOK_OUTPUT_DIR="$BOOK_DIR"
export PUBLICATION_DIRECTORY=book/edition-3
scripts/assemble-book.sh > "$BOOK_OUTPUT_DIR/the-art-of-building-ai-coding-agents.md"
scripts/build-ebook.sh
python3 "$BOOK_DIR/build-print.py"
```

To isolate generated output, set `BOOK_OUTPUT_DIR` to an existing owned build
directory instead. The assembler writes stdout. The EPUB and print builder use
the selected assembled file. The shared scripts retain Edition 2 defaults when
these variables are absent; the Edition 3 print builder always uses Edition 3.
Chrome rendering uses a fresh temporary profile, never the running user profile.

Outputs are the assembled Markdown, EPUB, coverless **Print Interior.pdf**, and
**The Art of Building AI Coding Agents.pdf** with the cover for free reading.
`interior-print.html` is the intermediate print input. Cover artwork is inherited
from the existing edition assets, not a newly approved publication cover. Inspect
layout and publication metadata before publishing; this work neither publishes
nor changes any store listing. Source links target the public repository and
will become available there only when Bill chooses to publish these changes.

Regression checks:

```sh
python3 -m unittest discover -s book/edition-3 -p 'test_*.py'
```
