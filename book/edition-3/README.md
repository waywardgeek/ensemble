# Edition 3 — C# inside Ensemble

This edition is locally complete through Chapter 27, Goals, including Gateway
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
Chapter 25 is a historical interlude without a runtime endpoint. Chapter 24 was
completed after Gateway and Goals; its preserved endpoint intentionally contains
that foundation. Earlier endpoints are not overwritten by later finalization.

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
book boundary. See Chapter 24 for the separate earlier probe and its unknown
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
