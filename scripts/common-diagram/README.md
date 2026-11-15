# Common data-relationship diagrams

This source-only tool turns `agent/Common/*.cs` into browsable diagrams and the
full-page ownership map in Chapter 0. It does not build or run Ensemble.

## Generate

Use an existing .NET 10 SDK, Python 3, and Graphviz (`dot`). No packages are
installed. From the repository root:

```sh
make common-diagram DOTNET=/absolute/path/to/dotnet
```

Or run `scripts/common-diagram/regenerate.sh /absolute/path/to/dotnet` directly.
The default SDK command is `dotnet`. Build intermediates stay in ignored `.build/`.
The extractor uses Roslyn from that SDK with NuGet package sources cleared.

Outputs are intentionally checked in under `docs/common-diagrams/`; readers
should not need the generator to see the architecture. The compact vector figure
is `book/figures/common-ownership.svg`. Regeneration does not assemble or publish
the book. After an intentional manuscript change, run `make book` separately.

## Responsibility and evidence

- `extractor/Program.cs` parses C# syntax, merges partial declarations, and records
  instance fields, stored auto-properties, positional record properties, source
  locations and fingerprints. It exports no initializer values or runtime data.
- `relationships.json` is the reviewed lifecycle map: parent field, child type,
  reverse field, strong/weak classification and teardown evidence. No syntax-only
  heuristic pretends to prove cancellation, joining or cascade deletion.
- `generate.py` validates those fields and source fingerprints, then emits DOT,
  SVG, PDF, a text relationship table and the offline viewer. Gray arrows cover
  other stored Common-type references without claiming ownership semantics.
- `viewer-template.html` supplies the local searchable, zoomable interface.

A changed reviewed implementation fingerprint refuses regeneration. Inspect the
changed lifecycle, update its relationship record if needed, then intentionally
refresh its reviewed SHA-256 fingerprint. Do not bypass refusal by blindly hashing
new bytes. New unreviewed stored references remain gray, not guessed strong/weak.
The inventory's own hashes prevent rendering against stale declarations.

Both ends of a modeled relationship are checked; duplicate reverse arrows are
suppressed. A shared child remains one node. One slot may be optional; a collection
means potentially many. Callback signatures and deferred Task result types do not
become containment arrows. Inheritance is labeled on boxes rather than confused
with ownership. External framework types, enums, static helpers, computed
properties and locals do not become graph nodes. The complete inventory includes
transient helper structures as well as live model objects: this is a source type
map, not a heap snapshot or a claim that every field is serialized to disk.

## Focused checks

After generation, use existing tools; none of these commands install anything:

```sh
python3 scripts/common-diagram/test-extraction.py /path/to/dotnet "$PWD"
node scripts/common-diagram/test-viewer.cjs /absolute/path/to/playwright
python3 scripts/common-diagram/test-print.py /absolute/path/to/playwright
```

The browser checks require installed Google Chrome. The print check additionally
requires pandoc and a Python environment with PyMuPDF. `test-print.py --node`
accepts an explicit Node executable. Missing tools fail explicitly.

Extraction checks use a small independent C# fixture and deliberate invalid
relationship/fingerprint inputs. Browser checks exercise the actual generated
viewer with HTTP(S) blocked. The print check renders Chapter 0 with the real
7-by-10-inch stylesheet into `.build/`, checks every class bullet and diagram
edge, verifies the dedicated page, and measures labels at least 10pt (0.1pt
rendering tolerance). It also checks EPUB XML and exact figure inclusion. Its PDF, EPUB and page image
are proofs, not released artifacts.

The Chapter 0 map preserves all reviewed lifecycle edges, wraps long names and
omits field labels. The adjacent class guide is handwritten teaching: its exact
node coverage is checked rather than replaced with generated comments. The
printed vocabulary must remain complete even when a reader cannot follow links.
