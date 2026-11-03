# Browser Markdown prototype

Bill approved a browser-side Markdown prototype and a later Dev on port 8089 trial.
Cadence on port 8088 and Measure X on port 8090 remain untouched. No staging or restart yet.

The isolated prototype established that reparsing accumulated Markdown is a
reasonable baseline: the desktop headless trial rendered 23,500 table-heavy
characters in about 13 ms. This is not phone performance evidence or a latency
bound. Native browser painting and real provider cadence still require human
acceptance on Dev.

## Candidate contract

- Pinned local markdown-it 14.1.0 browser bundle, MIT license and registry package
  integrity/extracted-file hashes recorded in vendor/README.md. No install
  scripts, runtime CDN or external image fetches.
- Raw HTML disabled. Only HTTP(S) links are clickable. Markdown image syntax
  becomes an escaped image-description placeholder, not a network request.
  No task-list or other plugins; do not claim every GitHub extension is supported.
- Original card.text remains authoritative for speech and History. Browser-only
  rendering coalesces deltas at 50 ms, creates cards in arrival order, and flushes
  finals immediately unless a selection intersects that card. Selection release
  applies deferred updates; replay discards pending work from the previous view.
- Prose headings/playback nodes persist. Table headers use scope=col and tables
  have labelled keyboard-scrollable containers. Code highlighting stays lexical;
  full code is retained behind the existing long-code preview with expansion
  preserved by block position. Previous table horizontal scroll is retained.
- Text-selection deferral includes cards lying between selection endpoints.
  A new stream can reclassify unfinished syntax; correctness of final rendering
  matters more than pretending partial Markdown has an immutable parse.

## Verification boundaries

Isolated prototype passed coalescing, every prefix of a Markdown sample, raw
HTML/unsafe-link refusal, zero remote image requests, selection deferral and
old scroll position. Actual pixels inspected for a synthetic standalone preview,
not the integrated Dev page. Production integration tests run the embedded page
and inject Markdown frames into its real handler; they are not model calls.
Existing real fake-provider streamed-turn test independently exercises actual
Agent/HTTP/WebSocket and speech completion. Final results recorded below once
collected. No live phone or Dev acceptance yet.

## Integrated candidate results

Final embedded-page checks passed: semantic tables, nested lists/emphasis,
original text unchanged, stable playback identity/focus, 100 deltas coalesced
into two paints, malicious HTML/URL cases inert with no remote requests,
selection deferral including intermediate cards, complete long-code expansion,
incomplete-prefix parsing, narrow-pane containment, old scroll position, and
pending-render cancellation on replay. Real fake-provider streamed-turn/speech
completion passed separately. Five public GUI groups, eight playback geometry
cases, four color/contrast cases and 32 composer layouts passed. GUI and GUI-test
builds clean. No core runtime change; shared C# suite not rerun for this browser
slice. Broader historical crossover-runner limitations remain unchanged.

Dev on port 8089 was confirmed running; installation directory exists at
`/Users/bill/ensemble-edition3-dev`. No files installed and no live host stopped.
Book requirements should be reconciled after the prototype's Dev review, before
calling this a fully accepted replacement. Markdown task-list plugins, arbitrary
HTML and automatic external images are intentionally not enabled.
