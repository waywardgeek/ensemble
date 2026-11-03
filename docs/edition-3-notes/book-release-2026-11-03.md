# Current book artifacts rebuilt

Rebuilt the main Edition 3 codebook after the Chapter 26 external scopes reached
100/100 and the current speech/pause and diagnostic requirements were reconciled.
This release is not a new full-book implementation grade, companion-book rebuild,
KDP preview approval, or external upload.

- 13 publication regressions passed.
- EPUB: 34 navigation entries, navigation outside the reading spine, cover and
  styles embedded; 1,402 local links/resources resolve.
- Print interior: 520 pages, 7 by 10 inches, coverless; 31 contents references,
  black print text, unnumbered front matter, recto preface at physical page 5,
  10.5-point body and minimum 10-point code.
- Reading PDF: 521 pages including the cover; all 106 links and labels preserved.
- Five generated files copied into `book/` and byte-compared against build output.

Chrome's CLI wrote a PDF but exited 2 during teardown. Disabling default apps,
background networking, extensions and GPU did not fix it. No failed PDF was
accepted. The new optional Playwright backend uses explicit existing Node/module
paths, waits for fonts, and closes Chrome in finally. The complete build exited
zero with unchanged final PDF checks. Tests cover backend selection, nonzero exit,
and absent output. No dependencies were installed; normal browser profile untouched.

Evidence: `/private/tmp/cadence-book-release.yc9tFZ` contains original failed logs, successful build logs,
regression output and release.json. Rebuild with BOOK_OUTPUT_DIR set explicitly,
BOOK_NODE=/opt/homebrew/bin/node and BOOK_PLAYWRIGHT_MODULE pointing to the existing
edition-2/agent/gui/node_modules/playwright. The older default CLI backend remains
available; no error is silently ignored.

## Artifact receipt

```json
{
  "epub_local_links": 1402,
  "print_pages": 520,
  "reading_pages": 521,
  "files": {
    "the-art-of-building-ai-coding-agents.md": {
      "bytes": 1165406,
      "sha256": "56121ae419b41c1edd619681f6e0c8e9e5bb7242c5537ff56e12a7aa64ace75f"
    },
    "The Art of Building AI Coding Agents.epub": {
      "bytes": 17100685,
      "sha256": "1ea737e5ac7a314bdc4bfde514b05561b5b69c7663272845da9b78c56ca491f1"
    },
    "The Art of Building AI Coding Agents - Print Interior.pdf": {
      "bytes": 2441078,
      "sha256": "f7e76fb314ed7494d40fc1e9f776922f5f07f5c289c9358872fd72e097b3567e"
    },
    "The Art of Building AI Coding Agents.pdf": {
      "bytes": 19642916,
      "sha256": "d459e03385051c2c3b6fc7cb3f6c18a8c479bea51ed1e390c02a00092c57399d"
    },
    "interior-print.html": {
      "bytes": 1525617,
      "sha256": "1a72d987ab07b0eb8bb8ca19a3dcc256f60d1e4c09c61b1671ef862f0d4c2811"
    }
  }
}
```
