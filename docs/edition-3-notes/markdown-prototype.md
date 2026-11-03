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

## Authorized Dev staging

Bill authorized Dev-only stop/stage/restart. Verified the exact 8089 process,
then preserved its running invocation, including transcription enabled through
an existing key-file path. The saved launcher intentionally remains unchanged
(and still defaults transcription off). No credential contents read or copied.

Two preflight refusals occurred before signalling or writing: the first compared
bundled skill manuals as runtime dependencies; the second required build-output
Linux binaries absent from this macOS-only installation. Corrected inventory to
all installed runtime files plus required macOS native library, excluding skills
while separately hash-checking their preservation. Then waited for full process
exit, atomically replaced six assemblies/symbols and verified all 15 installed
runtime files. Launcher/skills unchanged; no settings/data edits.

Receipt: `/private/tmp/cadence-markdown-dev-install/staging.json`. Dev restarted
with its previous invocation and isolated environment. Local8089 returned HTTP200
and exact source bytes for index,app,markdown,parser bundle and CSS.8088 and8090
listeners remain present; neither was signalled or staged. Public Cloudflare
cached assets and human phone/streaming acceptance remain unverified. No model
question sent. Candidate commit951c3ee2; the live Dev review is the next gate.

## Asset-version upgrade

Bill authorized eliminating repeated Cloudflare cache purges after a confirmed
stale-CSS incident. Each GUI mount now freezes the fixed public asset set and
computes a SHA-256 manifest fingerprint. HTML references versioned URL paths,
not query strings; dynamic and nested module imports inherit that directory.
HTML and legacy aliases are no-store; versioned static content is immutable.
Explicit disk overrides are startup snapshots, not live mutable files. Unknown
old versions never fall back to current content. Public-origin policy unchanged.

Focused actual HTTP tests passed root/mount parity, cache headers, changed-content
version changes, immutable bytes after disk edits, old-version refusal and nested
module access. Actual browser test confirmed all CSS/JS, capture modules and nested
voice modules use one version path; reload connects successfully. Five public GUI
groups and all Markdown integration cases passed. Chapter27 hard lint passed with
style advisories. Prior cached Cloudflare HTML still needs Bill's initial purge;
no claim that this diagnoses the newly reported size-related display error.

Asset-cache commit `b32ef1c7` was authorized and deployed to Dev only. Tested
GUI/core assemblies matched the final GUI build before staging. The prior Dev
process fully exited; all 15 runtime files verified after replacement. Launcher,
skills and running transcription options were preserved. Receipt:
`/private/tmp/cadence-asset-cache-dev-install/staging.json`. Live8089 verified
no-store HTML/legacy app.js and immutable versioned app/CSS/Markdown/capture/voice
assets with exact source bytes. The8088 and8090 processes remained unchanged.
No model request sent. Authenticated Cloudflare delivery remains unverified;
Bill plans the one-time purge when home.

## Human acceptance and main staging

Bill reported Dev loaded correctly, called the rendered HTML beautiful, and
confirmed tables and code blocks render incrementally: "it's fantastic."
This is live human acceptance of the Dev Markdown/streaming experience, not a
claim that every selection/security case was manually tested on every device.
Chapter7 now teaches the pinned parser, security policy, accumulated-Markdown
rendering, coalescing and selection/replay rules from the original construction
step. Chapter27's acceptance exercise matches. Both chapters pass hard lint
with style advisories; no historical endpoints or old grades were rewritten.

Bill authorized promotion and will restart main when home. Staged the exact
accepted Dev binaries to main, without another build: six application files
atomically replaced; all 15 runtime files match Dev and the tested build.
Launcher and installed skills preserved; credentials/settings/data untouched.
Main's8088 process remained running throughout; no restart performed.
Receipt: `/private/tmp/cadence-markdown-main-install/staging.json`.
Main runtime/browser acceptance awaits Bill's full exit/start/reload. Existing
Cloudflare entries may still need the one-time purge; versioning only takes
effect after the browser receives current HTML.

## Main acceptance and empty Thinking cleanup

After restarting main, Bill requested two rendering demonstrations and said they
looked great. This records main visual acceptance of Markdown/streaming. He then
approved hiding empty Thinking cards. The renderer now hides only thinking and
reasoning_summary cards with whitespace-only readable text; card identity/order
and backing data remain intact, and populated updates reveal the same card.

Negative-before test failed on the intended empty-card visibility assertions.
Eight geometry cases and four contrast cases pass after the fix. Actual embedded
page handler tests pass for empty opening, visible delta and empty final for both
kinds; all existing Markdown integration checks also pass. Chapter7 now teaches
this visibility rule. This cleanup is source-only, not installed/restarted.
