# Bill-owned preface rebuild

The coordinator regenerated existing publication artifacts at Bill's request,
without editing his preface or any chapter. `build.json` records exact source
hashes, commands and artifact hashes. The unchanged build scripts generated the
Markdown, EPUB, print interior and cover-prefixed PDF. No provider calls ran.

Reused `publication-tooling/full-verification.py` with this output directory and
actual page counts (371 interior, 372 with cover). All 30 contents references,
all chapter ordering/navigation assertions, image/link assertions and page-size
checks pass. The first temporary verifier adaptation replaced the old counts
sequentially and accidentally changed both to 372; its page-count assertion
failed. Simultaneous numeric substitution corrected that adaptation; no artifact
or build source correction was needed. This was a verifier error, not a failed
book build.

The exact preface text is in assembled Markdown. Its opening and closing text
are also found in the generated EPUB's Preface section and PDF pages 4–7. The
input hash remained unchanged throughout. Existing application/archive edits
remain untouched. This mechanical check does not recertify editorial claims or
repeat the earlier independent visual review. Bill will read before publication.
