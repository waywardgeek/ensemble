# Chapter 26 implementation notebook — Voice and Images

This is the working engineering/teaching notebook, not the finished chapter or a
passing grade. Cadence maintains it while implementing, then uses the verified
results to improve `book/chapter-26.md`. Historical Chapter 26 Gateway endpoints
and graders retain their meanings. Do not overwrite them for this chapter.

## Current boundary

The opening draft is committed at `bff61c65`; implementation has not started.
The initial source audit has established the gaps below. No microphone capture,
pixel capture, transcription request, or live multimodal inference has been made
in this audit. Publication tests for the opening are not implementation evidence.

The required sequence is opening sections -> working implementation/tests ->
exercises and executable external graders -> passing results -> full chapter.
Bill explicitly asked for notes throughout implementation so the chapter teaches
what was learned, not just the final successful commands.

## Settled requirements and reasons

- Voice is **hold, talk, release, transcribe, send once**. Use an on-screen held
  microphone button and a configurable shortcut, initially Right Alt where safe.
  My earlier editable-draft proposal was superseded by Bill's explicit preference;
  do not reintroduce a required second Send action. Repeats and late callbacks
  cannot send twice. Cancel/failure/focus loss/selection changes/disconnect must
  not send stale speech or redirect it to another Agent. Respect AltGr and
  reserved shortcuts rather than assume all Right Alt events are available.
- Paste actual images with accompanying text; show previews and allow removal.
  Normal Send admits the text and images together for the intended Agent.
- Initial screenshot scope is the connected Ensemble page. Other windows and
  desktop require separate explicit permission. DOM text, a capture request, or
  a filename is not captured pixels delivered to a model.
- Image-file inspection must use the Agent's existing Sandbox and deliver actual
  media to a supported model. A path or metadata-only tool reply does not count.
- Pixel 9 with Android Chrome is the primary phone target. Desktop Chrome is the
  initial shortcut target. Bill wants Claude and Gemini support as well as OpenAI,
  where the actual model/API supports the needed placement. Availability in a
  curated menu is not proof of media capability or permission to switch funding.
- Bill approved separately metered transcription and live multimedia tests across
  supported platforms/vendors. Keep test batches bounded and announced; verify
  current API/pricing facts first. Use synthetic or explicitly provided content,
  keep credentials out of files/logs/History, and coordinate native phone checks.
  Permission is not background listening, arbitrary file/window access, or an
  automatic fallback from Astra's current funding to an API key.

## Verified source observations

These describe the inspected source, not a provider limitation or live acceptance.

1. `agent/Common/Actor.cs` declares `AgentInput(InputKind Kind, string Text = "")`.
   `agent/gui/Server/Host.cs` handles `prompt` by posting that text-only input.
   The ingress path must carry real image content without bypassing Agent
   scheduling/History or losing busy-turn/hint behavior.
2. `agent/Common/Media.cs` and `agent/Engine/Media.cs` provide the existing media
   vocabulary and capability checks. Astra is currently recorded with no media
   support in this implementation. That does not establish the provider's actual
   capabilities. Recommendations and media support are separate facts.
3. `agent/Engine/Responses.cs` explicitly refuses top-level `BlobPart` and media
   in tool results. It currently renders function-call outputs as text. Image
   paste and image-returning tools therefore need real Responses wire support;
   adding a GUI button or changing a capability bit alone is insufficient.
4. `Engine.ReadMedia` already reaches path bytes through
   `Engine -> Agent -> Sandbox`. Preserve this path. Inline data takes a separate
   no-filesystem branch. The standalone `BlobRef.Data()` API is not a substitute
   for the Agent's confinement when inspecting a supplied path.
5. The existing adapters render images for some other model/surface combinations.
   `CheckMedia` currently permits media inside tool results only for Anthropic.
   Cross-vendor input and tool-result placement must be investigated separately;
   do not silently discard media when a placement is unsupported.
6. `agent/gui/wwwroot/mcp.js` offers DOM snapshot/click/input and speech-output
   operations, not speech input or screenshot pixels. Its tool replies are
   currently text content. A future capture path must return real media through
   the MCP/Agent pipeline and remain bound to the correct browser/Agent owner.

The first audit guessed a nonexistent GUI input filename, then located the actual
Host handler. One broad search was clipped; the relevant prompt block was read
separately. These are inspection corrections, not product defects or test passes.

## Open design and evidence obligations

- Verify the official Responses image-input and tool-result contracts, then the
  corresponding Claude/Gemini contracts and exact model capabilities. Preserve
  frozen History and provenance; no second scheduler or copied Sandbox/service.
- Choose and verify the transcription service, funding path, browser audio
  formats, resource limits, cancellation and exact-once admission behavior.
  Live-test permission is settled. The candidate and published rate below still
  require a configured funding path and live acceptance.
- Establish a genuine connected-page pixel-capture mechanism. Do not relabel a
  DOM reconstruction as native pixels. Check supported browser/platform limits
  and the role of visible user permission before promising universal support.
- Bound image count/bytes/dimensions/formats and audio duration/bytes before
  expensive work. Keep invalid input explicit and preserve accompanying text.
- Account for the known automatic-pause ownership defect on disconnect/AFK.
  Earlier speech-watchdog tests do not close that separate server-ownership gap.
  Do not solve it by unconditionally undoing a manual pause.
- Test real public components with isolated fake devices/providers, meaningful
  negative controls, and later authorized live checks. Record separately what
  each proves: bytes/placement, model interpretation, native capture, phone
  usability, cleanup, and actual provider usage/cost are different evidence.

## Public documentation check — initial audit

Official sources fetched successfully with TLS verification enabled:

- https://developers.openai.com/api/docs/guides/speech-to-text
- https://developers.openai.com/api/docs/guides/images-vision
- https://developers.openai.com/api/docs/pricing

The speech guide recommends `gpt-transcribe` for completed recordings/bounded
requests through `/v1/audio/transcriptions`. Its published upload ceiling is
25 MB, with mp3, mp4, mpeg, mpga, m4a, wav, and webm inputs. It distinguishes this
from Realtime transcription for an ongoing stream. Hold/release is a bounded
recording, so a completed-file request is the initial candidate rather than a
persistent realtime connection; this still needs format/cancellation/live checks.
Our application should impose its own smaller deliberate limits, not blindly
adopt the provider's maximum.

The price table lists `gpt-transcribe` at **$0.0045/minute**. This is a published
rate, not an actual bill, account entitlement, or measured test charge. The guide
marks `whisper-1`, the older 4o transcription variants, and the translation endpoint
as deprecated with a February 26, 2027 shutdown. Therefore do not choose an older
model merely from remembered names/prices. The candidate is identified; final
route configuration and live acceptance are still pending.

The image guide includes a Responses request with `input_text` and `input_image`,
where the latter carries `image_url` and `detail`. This supports investigating
real typed Responses image input; it does not establish image tool-output shape,
Codex-route acceptance, every curated model's capability, or an existing Ensemble
implementation. Those checks remain separate.

Evidence: `/private/tmp/cadence-media-api-5gdn6lg6/` retains the public HTML,
extracted text, and the first fetch receipt. The first Python fetch failed SSL
certificate validation for all three pages and retrieved no documents. Retrying
once with `/usr/bin/curl` returned HTTP 200 for all three, with TLS verification
still enabled. Do not disable certificate checks to compensate for that Python
installation's trust-store problem. Only relevant excerpts were reviewed; this
was not a complete read of every example in the three long pages.

No credentials, microphone recording, screenshots, or paid API calls were used
for this documentation check. No product code changed.

## How to maintain this notebook

For each meaningful implementation step, append or revise a concise entry with:

- the observation/problem and the decision, including why;
- alternatives rejected and the tradeoff, not only the eventual patch;
- actual source/ownership paths and wire contracts changed;
- test command/evidence location and result, including failures and corrections;
- limitations or human/provider checks still pending;
- the teaching/exercise/grader change implied by the result.

Keep one current boundary rather than accumulating competing status summaries.
Retain important failed approaches and superseded decisions with their reasons.
No credentials, raw private audio/images, or unnecessary transcript copies belong
here. Do not claim a feature works because it was proposed, built, staged, or
accepted by a synthetic fixture. The full chapter follows verified implementation
and passing external grades, not the other way around.

## First implementation slice — Responses human images

- The official images guide uses `gpt-6-astra` with Responses `input_image`;
  implemented that input placement, not arbitrary media capability. The model
  row now advertises images, and preflight still rejects unknown models, other
  modalities, non-human image entries, and image tool results before reading files.
- Adjacent text and images remain ordered in one message content array. Tool and
  native opaque items end that group. Rendering uses Engine → Agent → Sandbox;
  it does not mutate canonical parts, expose file locators, or introduce a second
  filesystem owner. Inline and rooted payload bytes are asserted in tests.
- Ten Responses groups and the full shared suite (315 passed, zero failed) passed.
  Evidence: `/private/tmp/cadence-multimedia-implementation/`, especially
  `responses-images.log` and `shared-images.log`. This is local projection,
  Sandbox, and existing fake-provider coverage—not live vision acceptance.
- The first edit attempt guessed obsolete type/declaration names and failed before
  changing any source. Its shell lacked fail-fast handling and then ran the old
  tests. Those passes were not patch evidence. I read the actual declarations,
  rebuilt the checked patch, and ran the new focused and full tests successfully.
  This was my workflow error, not a product failure.
- Still needed: bounded image validation/decoding, typed GUI message admission,
  tool-result images and image-file inspection, paste/preview ownership, capture,
  transcription, live provider/platform checks, exercises and external graders.
  No application binaries were installed or restarted for this slice.

## Capture and tool-output contract findings

- MDN's current browser compatibility data lists `getDisplayMedia` as unsupported
  on Chrome Android and Safari iOS, supported on desktop Chrome. Its capture-handle
  API is likewise desktop-Chrome-only among those primary targets. This is
  documentation evidence, not a native phone check.
- Proposed initial capture scope: real connected-page pixels on desktop Chrome,
  explicit refusal on unsupported browsers; a phone OS screenshot can instead be
  supplied as an image. This platform limitation still needs Bill's assessment.
  DOM reconstruction is not an acceptable silent substitute for actual capture.
- OpenAI's function-calling guide and official Python SDK describe
  `function_call_output.output` as text or typed content. The exact list schema is
  being checked before enabling the currently refused tool-image path. Sources:
  <https://developers.openai.com/api/docs/guides/function-calling>, and
  <https://github.com/openai/openai-python/tree/main/src/openai/types/responses>.
- The broad Responses reference exceeded the bounded download cap; I did not
  treat that incomplete fetch as reviewed. Smaller official SDK schema files
  provide the specific contract without an unbounded reference-page download.
- Public source receipts are under `/private/tmp/cadence-media-api-5gdn6lg6/`;
  none of these research requests used credentials or made model calls.

## Second implementation slice — typed Responses tool images

- Verified the official SDK list elements: `input_text` with `text`, and
  `input_image` with `image_url` and optional `detail`. Image-bearing function
  outputs now use that ordered array; text-only outputs retain their prior string
  contract. A shared formatter handles human and tool image blocks.
- The media preflight permits Responses tool images, but still rejects other
  modalities, unsupported models and remote/mismatched references before reads.
  Actual file reads retain the Agent's Sandbox. Redaction markers remain readable
  text and never cause their private retained locators to be opened or sent.
- Ten focused Responses groups passed, with new assertions for ordered mixed
  tool output, actual inline/rooted bytes, call pairing, unchanged canonical parts,
  redaction privacy and outward-path/symlink refusal. The previous negative image
  case is now an unsupported audio case, not a weakened missing-payload check.
- Evidence: `build-tool-images.log`, `responses-tool-images.log`, and the successful
  shared run `shared-tool-images.log` (315 passed, zero failed; exit zero) under the implementation evidence directory.
  No provider call, image-file tool, GUI ingress, installation or restart yet.
- Next: bounded byte reads and image format validation. `OpenRegularRead` already
  provides a Sandbox-owned stream; use it rather than checking length after an
  unbounded `ReadFile`. Keep global resource policy explicit and test malformed,
  missing, oversized and format-mismatched payloads before enabling user tools.

## Third implementation slice — bounded image bytes

- Introduced an explicit 8 MiB per-image application limit, shared by Responses,
  Chat Completions, Anthropic and Gemini rendering. This is conservative local
  resource policy, not a claim about provider upload limits. Other modalities
  retain their existing behavior.
- Path images use the existing Agent Sandbox `OpenRegularRead`, reject empty or
  oversized files before allocation, read exactly the bounded length, and reject
  extra trailing bytes. Disposal is lexical on success/failure. This catches
  observed truncation/growth, not same-size concurrent rewrites or a file changing
  after the final read; it is not an immutable filesystem snapshot.
- Inline images bound the encoded length and decoded ceiling before copying or
  decoding. Empty, malformed base64, missing files and oversized encoded/decoded
  payloads now refuse explicitly. Standalone `BlobRef.Data()` is unchanged;
  runtime rendering uses the confined bounded path, not that standalone helper.
- Ten Responses and two cross-provider media groups passed before the final
  decoded-ceiling tightening; the final version passed all 315 shared groups
  with exit zero (including Responses and cross-provider media groups).
  Logs: `build-bounds.log`, `responses-bounds.log`, `media-bounds.log`, and
  `shared-bounds.log` in the implementation evidence directory.
- The first focused test failed because I incorrectly assumed InvalidDataException
  was covered by an IOException catch filter. The implementation correctly refused
  empty input. Correcting the fixture's exception list produced the focused pass.
- Remaining: image content/format and pixel-dimension validation, aggregate request
  limits, supported-format policy, GUI admission and image-file tool. A byte limit
  alone does not establish image validity or prevent decompression bombs. No new
  dependency was added, no provider contacted, and no live application changed.

## Bounded resizing foundation — implementation and dependency decision

Bill requested image resizing and its inclusion in the Chapter 26 TL;DR. Added
that specification without advancing the full chapter ahead of acceptance.

Implemented `Agent.PrepareImage(path)` and `Agent.PrepareImage(bytes)` with an
immutable `PreparedImage` delivery value. Declarations/policy stay in Common;
bodies stay in Agent. The path overload uses the owning Agent's Sandbox stream.
No original is overwritten. Preparation produces actual inline PNG bytes for
admission into History; it is not a provider-render-time transformation.

Policy defaults (application limits, not advertised provider limits): source
1 byte–32 MiB, at most 24,000,000 pixels, each source edge at most 32,768;
output long edge at most 2,048, at most 8 MiB. PNG/JPEG/WebP/static GIF are
accepted after actual codec decoding; incomplete decode refuses. Animation is
explicitly refused rather than silently selecting a frame. EXIF orientation is
applied before metadata is discarded. PNG output preserves alpha and avoids
JPEG artifacts in screenshots. Oversized encoded output reduces dimensions by
0.75, always sampling from the original decoded image, for at most eight encode
attempts. Dimensions/byte counts and `Resized` are returned for caller reporting.
The UI must show this transformation; returning those facts is not yet UI work.

### Library choice and actual package inspection

Pinned SkiaSharp and SkiaSharp.NativeAssets.Linux.NoDependencies 4.154.0.
SkiaSharp's package LICENSE.txt is MIT; its native dependencies/notices still
ship with upstream packages. macOS/Windows native assets are transitive; Linux
is explicit. Active agent/NuGet.Config now explicitly permits nuget.org; frozen
endpoints were not changed. No global install or credential use occurred.

Initially considered ImageSharp because it is managed. Its public license offers
Apache terms for open-source/source-available use, but the inspected 4.1.3
package includes a mandatory Release-build license-key check. I rejected adding
registration/key provisioning to the executable book, not the library's decoding
quality. I did not bypass the check or select an old package merely to evade it.
Skia's native packaging is the deliberate tradeoff. Public license/package/API
receipts and logs: `/private/tmp/cadence-image-library/`.

### Evidence and corrections

`--image-tests` passes against the actual native decoder on this Mac. It checks
real dimensions and decoded pixels, alpha, no enlargement, 4096×2048 → 2048×1024,
all eight EXIF transforms, source preservation, outside/rooted symlink refusal,
empty/malformed/truncated input, 32 MiB source bound, 25-million-pixel rejection,
two-frame GIF refusal, and a deterministic noisy image above 8 MiB that must be
reduced further to meet the output-byte limit. Responses, Anthropic and Gemini
wire projections retain the exact derivative bytes and canonical History after
the source test file is deleted. No provider request was sent.

Failures/corrections: first restore failed with the existing cleared sources;
explicit source restore succeeded. I initially misreported the combined job's
failure as restore failure before reading its final log; it was then compilation.
An Engine type/member collision prompted moving the delivery policy to Agent
with an Engine compatibility constant. Test registration incorrectly used `Case`
instead of the actual `CheckCase`; corrected. Skia reports FrameCount zero for
static images, not one; the initial guard rejected a real PNG. Read package docs,
corrected the guard and added actual animation/orientation controls. No failed
run is counted as passing evidence.

### Remaining boundaries

This is an admission foundation, not an installed GUI feature or a live Gemini
acceptance result. Existing arbitrary low-level BlobPart construction does not
automatically call preparation. GUI/image-file tools must call it once and retain
its returned bytes; aggregate request limits still need implementation. Pixel
bounds prevent allocating the declared giant raster, but the native decoder is
not a process sandbox or a hard CPU/total-memory quota. Concurrent preparation
and metadata parsing are not globally resource-budgeted. Native Linux/Windows
execution has not been tested. Color-profile fidelity and human legibility of
resized screenshots are not established by these synthetic tests. Changes are
source-only; live 8088, installed binaries and credentials are untouched.

Final verification for this slice: Debug and Release builds completed with zero
warnings/errors. Focused real-image tests passed in both configurations. The
shared Debug suite completed **316 passed, zero failed, exit zero**. Logs are
`build.log`, `tests.log`, `release-build.log`, `release-tests.log`, and `shared.log`
in the evidence directory above. These are local Mac/native/wire checks, not
Linux/Windows, browser capture, provider inference, or usability acceptance.


### Image-file tool admission — structured inspect_image

Added `inspect_image` to code-tools, using the existing Agent/Tool/Job/Sandbox
ownership path. Engine.ValidateMedia exposes the existing read-free capability
and placement validator; the tool checks a nested image result before opening
its actual source path. It prepares once and returns dimensions, source/output
byte counts, and the actual PNG BlobPart. Structured Job retention preserves the
payload through delayed `wait_for_job` observation rather than flattening it into
bounded text output. Source files are not overwritten or reopened for projection.
Unsupported model/surface and invalid decode errors are explicit error results.

Current support is Responses and Anthropic image tool results for models already
in the capability table. Gemini's tool-result image placement remains explicitly
unsupported, even though its human-image projection accepts prepared derivatives.
Do not imply curated menu membership establishes image capability.

Focused tests use real Agent, SkillManager, Job, Sandbox, decoder and renderer:
skill denial/load/revocation, delayed observation, actual resized pixels and
transformation report, original preservation, source deletion before paired wire
projection, malformed image and outside-path refusal, unsupported capability
before missing-file access. Both image groups passed. No paid requests occurred.

Fixture corrections: nonexistent DestroyAsync was replaced by actual Destroy;
direct control calls record ToolObserved events, not model dialogue, so projection
uses an explicit paired Context containing the actual returned result (not claimed
model-turn History acceptance). An unlisted Anthropic model correctly refused;
fixture now uses the existing haiku capability row. One-byte diagnostic truncation
obscured that error; the test now uses a bounded 4096-byte report. These were
fixture mistakes, not runtime/media failures. Shared regression is pending below.

The first shared run finished 316 passed / one failed: the exact built-in tool
inventory did not yet include inspect_image. Updated that name set explicitly;
no relaxed count assertion. Also corrected the new fixture's source path to the
tracked lowercase agent/skills directory (Mac case-insensitivity had hidden it).

Final verification: build zero warnings/errors; focused image 2/2 and tool 1/1
passed; final shared Debug suite **317 passed, zero failed, exit zero**. Evidence:
`/private/tmp/cadence-image-library/inspection-{build,image,tool,shared}-final.log`
(build file is `inspection-build-final.log`). Production changes reviewed for
read-free capability checks, direct ownership, structured retention and cleanup.
No install/restart/push or live provider requests. End-to-end model-turn acceptance,
GUI attachment admission, aggregate limits and Gemini image tool results remain
unfinished. Native decoder cancellation is checked before and after, not a
preemptible CPU quota. Full source/low-level BlobPart validation is not automatic.


## Interactive image admission — preparing the GUI paste path

Bill authorized continued implementation with synthetic fixtures while away;
no native clipboard/phone participation is required for these local checks.
AgentInput now carries an immutable image array alongside its existing text.
Both idle human input and busy hints record the same text-plus-image parts in
canonical History; frozen in-flight requests remain unchanged. Controls such as
Pause, Resume and Ephemeral refuse attachments rather than silently discard them.

Admission permits at most four prepared inline PNGs, 8 MiB each and 16 MiB combined.
It checks payload shape/base64 and model support before History publication;
unsupported or malformed inputs use the actor's existing admission-error contract
instead of escaping its loop. It does not raster-decode again: host admission
must call PrepareImage first. This does not make arbitrary low-level BlobPart
construction a validated raster or establish a cumulative History/request quota.

A gated local HTTP fixture uses real Agent/Engine/History objects: it observes
one derivative in the first frozen request, admits a second as a busy hint, then
observes both exact payloads in the next request, with two canonical images and
retained accompanying text. Negative controls cover control-kind misuse, count,
per-image and combined byte limits, path references, malformed base64, unsupported
models, no History publication on refusal and continued actor usability.
Focused final image tests pass 3/3; final shared Debug suite passes **318/0**, exit zero.
Logs: /private/tmp/cadence-image-library/input-{build,focused,shared}-final.log.

GUI inspection: controls currently have a 1 MiB WebSocket ceiling and inbound
raw-JSON tracing. Do not simply raise the ordinary-control limit or log pasted
media. The attachment wire needs explicit count/source/combined bounds, payload-
free tracing, fixed selected-Agent routing, server preparation, and browser
preview/removal ownership. GUI and native clipboard acceptance are not established
by this actor fixture. Work stays on isolated fixtures, not the live server.


## GUI attachment transport — isolated server implementation

Added an image_prompt WebSocket envelope with explicit agent_id, positive safe-
integer sequence (strictly increasing per socket), text and one to four images
with mime and unspaced base64 data. Source formats: PNG/JPEG/WebP/GIF, decoded by
PrepareImage and normalized to PNG once. Sources are capped at 32 MiB combined,
text at 32 KiB, and prepared delivery at the existing 16 MiB aggregate input cap.
A bounded larger receive buffer supports this envelope; ordinary controls still
refuse above 1 MiB. This changes the maximum socket buffering budget and does not
claim a global multi-connection memory or decoder-concurrency quota.

The server reserves sequences even on refusal, requires the current selected
Agent address, and acknowledges only after PostAsync accepts the immutable input.
image_accepted includes per-image original/output dimensions, sizes and resized
status, not pixel payloads. Errors carry image_sequence when parsed. This is
admission, not model completion. Sequence scope is a live socket; clients must
never automatically replay an uncertain send after reconnect. Image fields on
ordinary controls refuse rather than silently disappear. No browser composer
integration has been implemented yet.

Trace projection replaces any frame containing image fields, MIME markers or
image data URIs with a type/byte-count/redaction summary, covering inbound images
and nested outbound History/tool payloads. Ordinary diagnostic frames retain
existing behavior. This is not a promise to detect arbitrary private content
hidden in ordinary text.

A public GUI-consumer WebSocket fixture keeps the Agent paused (no provider
requests), sends a real noisy PNG exceeding the ordinary 1 MiB control limit,
checks retained pixel equality/text, duplicate and wrong-Agent refusal, missing
address/data, malformed raster, misplaced image fields, ordinary-control size
refusal, later valid admission, preparation metadata and trace privacy.
Initial fixture compilation failed because I guessed an internal StartAsync API,
used uppercase Surface instead of the real parameter, and treated Context as an
entry sequence. Public History.Events and the existing actor construction contract
corrected the fixture; no product visibility was expanded. The next focused run
passed. Final focused image check and all five default GUI regression groups passed,
exit zero, including selection, peer controls, image controls, public-origin
security and external consumer lifetimes. The added valid-first/bad-second-image
case verifies atomic message refusal. Logs:
/private/tmp/cadence-image-library/gui-images-*-final.log. Core shared code did not
change in this slice; its prior 318/0 evidence remains distinct. No clipboard,
live provider, installation, restart or push occurred.

## Browser paste composer — implementation and verification

Added an embedded `images.js` component owning attachment entries, FileReaders,
object URLs and one pending send. Paste preserves accompanying plain text, shows
numbered/removable previews, and bounds four PNG/JPEG/WebP/GIF sources to 32 MiB.
The browser does not declare pixels valid: server preparation remains authoritative
and rejects animation/invalid data. Native browser preview decoding has its own
resource behavior; server pixel ceilings are not a browser-wide memory guarantee.

The existing form/selection/socket code remains authoritative. Reading and pending
admission disable Send. A matching socket, sequence and Agent acknowledgment clears
the draft; refusal retains it. Selection discards attachments, aborts pending reads
and makes stale callbacks inert. Pending selection-change notices explicitly warn
that a previous send may already have been admitted. Disconnect discards unsent
attachments while retaining text; an uncertain sent draft remains blocked across
reconnect and requires explicit discard. There is no automatic retransmission.
Preview URLs are revoked on removal, clear and page exit. Existing media cards,
speech opt-in and manual/automatic pause ownership are preserved, not replaced.

Evidence directory: `/private/tmp/cadence-image-browser/`. Real headless installed
Chrome uses synthetic ClipboardEvents/DataTransfers, not the system clipboard.
`agent/GuiTests/image-browser.cjs` drives an isolated real crossover host and checks
preview/removal/caption, actual server preparation acknowledgment, duplicate form
submission, phone-width geometry and real socket reconnect without resend. Controlled
component cases cover refusal, stale reader callbacks, pending-send guard and wrong
acknowledgment identity. These are not native Pixel 9 clipboard or real-provider tests.

Failures and corrections: bundled Playwright Chromium was absent; used installed
Chrome without downloading. First implementation constructed the component before
`notice` initialization: real startup defect, corrected by moving construction.
The initial test omitted finite waits; its owned process was cancelled and bounded
waits/page-error reporting added. Reusing the fixture directory failed exclusive
Job-file creation before startup; preserved evidence and used a fresh directory.
A wrong `#model` fixture selector became actual `#model-choice`. A duplicate-send
component mutant initially survived because the form independently disabled Send;
added a direct component guard assertion. The final duplicate and selection mutants
failed their intended assertions (2 versus 1 sends; retained versus cleared draft).
Neither mutation touched repository or live application code. Existing speech-pause
regression passed with controlled native speech; this does not establish audible or
native-device acceptance, nor repair the separately known disconnect/pause leak.

Final positive browser/default-GUI results and owned fixture exit are recorded in
the progress checkpoint after collection. Core remains at the earlier 318/0 result;
this slice changes browser assets only. No installation, live provider request,
restart of the working application or push. Cumulative request budgets, image-tool
model-turn acceptance, Gemini tool media, voice/capture and external graders remain.

Final collection: `browser-final.log` passes all three browser groups (including
wrong acknowledgment identity checks); `gui-final.log` passes five groups;
`speech.log` passes 24 assertions. All corresponding processes exited zero.
The fresh owned host at fixture-fresh.G71VIf was stopped using its documented
stop file; host process exited zero and joined its provider/child work. Source
syntax and whitespace checks passed. No live host or provider was used.

## Cumulative image request budget — locally verified

Added a request-local Engine budget: at most 32 transmitted images and 32 MiB
of image payload reservations across human History, tool results and ephemera.
Repeated images count each time they are transmitted. All four image adapters
share the same bounded reader; reservation precedes the next decoded allocation.
Inline base64 uses its decoded-size upper bound (noncanonical whitespace can
therefore be conservatively overcounted). The budget is not stored on Engine,
not shared with the recall judge, and not reused across failed or concurrent
renders. Canonical History is never rewritten and images are not silently dropped.
These are application safeguards, not provider limits, total wire/token/cost
limits, or a global allocator/concurrency quota. Long conversations can hit them;
a dedicated GUI context-reduction/recovery action is not implemented.

Four focused image groups pass, including exact count/byte boundaries on all
four adapters, inline/path bytes, tool results, ephemera, unchanged History and
parallel/fresh render allowances. Large payload fixtures test byte accounting
only, not raster validity. Final shared regression: 319 passed, zero failed,
process exit zero. No installation/restart/push or live-provider test occurred.
Failures corrected: initial build used a nonexistent isolated TMPDIR; compiler
identified recall-judge direct renderer call sites omitted in my first edit; new
fixture incorrectly used IDisposable rather than Ensemble's explicit Destroy.
Final focused evidence: /private/tmp/cadence-image-library/budget-*-final.log.

Gemini investigation: public python-genai types.py now defines FunctionResponse
parts containing FunctionResponsePart.inline_data, with MIME/raw bytes and an
optional matching call ID. Saved public SDK and official guide under
/private/tmp/cadence-gemini-media/. SDK syntax alone does not establish model
support or live acceptance; Gemini tool-image refusal remains unchanged.

Official function-calling guide restriction: its multimodal example says Gemini 3
series, but currently demonstrates the Interactions API (`function_result` content
blocks), not Ensemble's GenerateContent wire. Do not transpose those fields into
GenerateContent. Verify its own REST/SDK conversion and model support next.


## Model-issued image inspection: verified local acceptance

Added a three-request real Agent/HTTP fixture for Responses and Anthropic. The fake
model asks for `inspect_image`; the next request must contain the correctly paired
image bytes and text. The fixture checks that the source was not overwritten, deletes
it, asks for a separately granted caption-file read, and verifies the third request
still carries the derivative and the caption nonce. Canonical History has exactly
one inspection call/result with the actual PNG; decoded dimensions are 2048 × 16
from a 4096 × 32 source, and the synthetic magenta pixels and resize report survive.
A one-byte ordinary output cap cannot flatten the structured image result.

Fixture corrections: an HTTP disconnect initially masked the fake provider's precise
assertion; the fixture now surfaces its owned server task's error. The minimal primary
skill omitted `read_file`, which code-tools deliberately does not grant. Granting the
follow-up explicitly fixed that refusal; an upfront advertised-tool assertion now
catches this setup error before tool execution. No runtime policy was weakened.
Five focused image groups and final shared regression passed **320/0, exit zero**.
Evidence: `/private/tmp/cadence-image-library/model-turn-*-final.log`; earlier failed
logs retain the diagnostic and authority mistakes. This is real actor/History/HTTP
integration with fake providers, not paid-provider acceptance. No install/restart/push.

### Gemini contract and Bill's target clarification

Bill explicitly uses **Gemini 3 Flash and newer**. Gemini 2.5 fixtures are compatibility
or refusal evidence, not his multimedia acceptance target. Verify exact newer model
IDs and capability support; do not enable an arbitrary model-name prefix or alter the
curated menu/funding as a side effect.

Public sources are retained under `/private/tmp/cadence-gemini-media/`. The v1beta
GenerateContent discovery schema defines `FunctionResponse.parts`, `FunctionResponsePart`
with `inlineData`, and blob `data`, `mimeType`, and optional `displayName`. Its response
JSON can reference named blobs through `$ref`, and call IDs remain explicit. The current
Python SDK's Developer API `_Part_to_mldev` passes `function_response` through; it does
not contain a rejection for that media field. Thus the REST wire location is verified,
not merely inferred from the Interactions SDK. The current function-calling and Gemini 3
guides nevertheless demonstrate Interactions, including `gemini-3-flash-preview`; they
are not a GenerateContent live acceptance result or proof every newer model is available.
No credential use or paid requests occurred. Keep Gemini tool media refused until the
specific supported model path is implemented and verified. Existing Gemini human-image
wire tests remain separate evidence.

## Gemini 3 Flash function-response images — implementation and live contract probe

Bill explicitly uses Gemini 3 Flash and newer; 2.5 is retained only as legacy
coverage, not his acceptance target. The implementation now has an exact
`gemini-3-flash-preview` capability row, including its published/discovered
1,048,576 input-token window and a separate `FunctionResponseImages` capability.
No model-name prefix implies support; no curated menu, running provider, funding,
or installation was changed. Other newer IDs still require verification.

### Source and account evidence

Public v1beta discovery confirms `FunctionResponse.parts` with
`FunctionResponsePart.inlineData`, including MIME, base64 data, and displayName;
response JSON may refer to a named blob with a single-key `$ref` object. Current
Python SDK Developer API conversion passes functionResponse through. The current
guide's examples use Interactions, so they were not copied into GenerateContent.
Public sources and extracts are under `/private/tmp/cadence-gemini-media/`.

The standard pricing page lists Gemini 3 Flash Preview input text/image/video at
$0.50/million tokens and output including thinking at $3.00/million tokens. A
single authenticated metadata GET for `gemini-3-flash-preview` succeeded and
advertised GenerateContent. Only the authorized Gemini key source was used;
credentials were never printed, placed in argv, copied into fixtures, or logged.
The key was supplied through a private pipe. No funding fallback occurred.

An announced, non-retrying two-request GenerateContent contract probe then used a
64×32 synthetic red/blue PNG as a function result. Both requests returned 200;
the second request preserved the first returned model content and call ID and
provided real bytes in functionResponse.parts. The final answer correctly named
red on the left and blue on the right. Sanitized receipt:
`tool-image-probe-result.json`. Prompt tokens were 75 + 1,351 = 1,426; candidate
plus thought tokens were (10 + 166) + (19 + 84) = 279. Standard-rate arithmetic
estimates $0.00155, not a verified bill or entitlement. The second request reported
1,081 image tokens: this probe does not show that shrinking an image necessarily
reduces image-token charges. There were no retries or private media.

This was a real API wire-contract check, **not a live Ensemble model turn** and
not streaming/phone/browser acceptance. Those boundaries remain separate.

### Runtime and local tests

Image-bearing Gemini results now contain ordered text and `$ref` content plus
named inline image parts. Names are response-local and never expose source paths.
ReadMedia remains Engine→Agent→Sandbox and shares the request budget. Error
results retain their error wrapper; text-only object/string projection is unchanged.
Redaction markers do not expose private retained-output locators. Nested image
results and other unsupported mixed structures explicitly refuse rather than
flattening pixels into metadata. Call IDs, function names, and thought signatures
remain paired through replay.

The real Agent/HTTP model-turn fixture now covers Gemini 3 Flash as well as
Responses/Anthropic: actual inspect_image dispatch, resized pixels, deletion of
the original, a real follow-up read_file, retained History and signature replay.
New focused cases cover mixed/error ordering, unique image references, redaction
privacy, text-only compatibility, unsupported nested/modality/model cases, and
Gemini tool-result request budgeting. Six focused groups passed, exit zero;
build-final.log and focused-final.log are in the evidence directory above.

An edit script initially stopped on an incorrect CheckCase registration anchor;
the actual existing test uses AsyncCase. Earlier file edits were inspected, the
single missing registration was completed, and only then were builds/tests run.
An inspection grep also referenced a nonexistent README; the actual skill manual
was inspected instead. Neither was a product defect or passing stale-code test.
Final shared regression passed **321 groups, zero failures, exit zero**
(`shared-final.log`). Final build and six focused groups also exited zero. No
live Ensemble model-turn, streaming, phone, installation, restart, or push is
claimed. The legacy 2.5 refusal remains a compatibility case, not Bill's target.

Preparation for voice: the current public OpenAI SDK's audio model schema includes
`gpt-transcribe`, and transcription-create parameters support file/model plus
non-streaming responses. Source files are under `/private/tmp/cadence-voice-contract/`.
No audio request or microphone access occurred. Its deprecated-model shutdown
comment should be reconciled with the previously collected guide before teaching
an exact date. Voice ownership, explicit separate funding, capture cancellation,
and a server-verifiable duration/byte bound remain implementation work.


## Live Ensemble Gemini image-tool turn — bounded, non-streaming

Bill reaffirmed permission for useful paid checks. A fresh isolated executable
referencing the real Ensemble project passed a fake-provider dry run and then one
announced live batch against exact `gemini-3-flash-preview`. No production source
changed. The real Agent, Engine, Sandbox, skill grants, tool, native preparation,
History and HTTP authentication ran. The injected test transport added only the
1,024-output-token generation cap, bounded request/response bodies, enforced the
exact endpoint and at most two requests, checked tool calls before dispatch, and
collected sanitized usage. This is instrumented non-streaming acceptance, not an
unmodified-production or streaming campaign. Recall/judging and all memory bands
were explicitly disabled; the tool-round limit was one. Universal skill controls
remained advertised, but the transport refused their invocation; only one
`inspect_image` on the owned synthetic file could execute.

The live process exited zero: exactly two HTTP-200 requests; one canonical
inspection; actual red/blue pixels correctly identified; original 4096×64 resized
to 2048×32; original function ID and thought signature preserved. The original
file was deleted at the second-request boundary, and retained History re-rendered
successfully after completion without reopening it. Both application teardown and
transport disposal were verified. The authorized Gemini key remained in memory;
it was neither printed nor copied into fixtures. A bounded scan of the owned
fixture files found no credential value. No user media, microphone, phone,
installation, restart, funding/menu change, or push was involved.

Measured input: 752 + 2069 = 2821 tokens. Output/thinking:
27 + 132 + 24 + 50 = 233 tokens. At the previously verified standard prices,
$0.50/M input and $3/M output/thinking, the estimate is $0.0021095, not a bill.
The second request reported 1068 image tokens. Resizing is not a guarantee of
proportional token savings. This batch is separate from the earlier two-request
raw API contract probe ($0.00155 estimate).

Harness corrections: restore took about a minute, not a stall; an invented
`Context.Transcript()` call was corrected to the actual `Dialogue` collection.
The first dry run refused before network because its declaration allowlist omitted
universal `load_skill`, `unload_skill`, and `micro_handoff` controls. Reading the
real skill rule corrected the allowlist while retaining strict response-side
execution gating. One mismatched edit anchor stopped unchanged. Fake-only error
detail made these setup failures diagnosable without logging live response bodies.
Dry-run failures all joined cleanup and made zero upstream requests. These were
harness defects, not runtime image defects.

Evidence: `/private/tmp/cadence-gemini-live/`, especially `dry-3.log`, `live-1.log`,
the corresponding owned run receipts, and the local harness source/build logs.
The earlier shared 321/0 result remains applicable: only evidence notes changed.
Remaining acceptance includes streaming, other exact newer IDs, other live
providers, native phone clipboard and usability, voice and capture, then graders.
