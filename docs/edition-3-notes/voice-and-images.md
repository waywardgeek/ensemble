# Chapter 26 implementation notebook — Voice and Images

This is the working engineering/teaching notebook, not the finished chapter or a
passing grade. Cadence maintains it while implementing, then uses the verified
results to improve `book/chapter-26.md`. Historical Chapter 26 Gateway endpoints
and graders retain their meanings. Do not overwrite them for this chapter.

## Latest boundary:75 external points; native host-screen test awaits Bill

Image-files20, attachments25 and voice30 have passing public external grades.
Reports, negative controls and exact scope limits are recorded in
[ch26-grader-completion.md](ch26-grader-completion.md). Chapter26 is not complete:
host-screen25 remains ungraded. Native driver/schema/pixel/Job/replay controls
pass with synthetic noncapturing candidates; no native enumeration, prompt or
capture occurred during this grading work.

The guarded host-screen CLI and static four-band fixture generator are now ready
for review. Missing native opt-in, wrong scope flags, malformed/ambiguous manifests,
degenerate colors and mismatched prepared inventory refuse before capture.
Pattern files are prepared but unopened at
/private/tmp/cadence-host-driver/pattern.W1c0az/. The2048x1152 expectation comes from
the previous display setup, not a fresh inventory. Current dimensions and color
profile remain native-test facts to verify. The HTML needs a human full-screen
click on the server's primary display; no requesting-phone capture is substituted.

Next requires Bill at the host: confirm a fully synthetic primary display, then
explicitly authorize inventory plus two captures (default0/explicit0), delayed-Job
collection and replay without recapture. Secondary displays, installs, restarts,
paid calls and background capture are outside that permission. Reminders are
being disabled at this human-dependent boundary, retaining the unfinished goal.

Bill has since restarted and accepted streamed Markdown on main; older staged/
restart-pending notes below are historical. Empty-Thinking cleanup9c557a2d and
subsequent public voice seam/grading changes remain source-only for live hosts.
No live service was changed by these external grading runs.

## Image-context repair staging and chapter incident note — 2026-11-02

Bill requested an incident passage in Chapter 26 and authorized staging for his
restart. Section 26.3 now explains Test's reported 33-image/~15-MiB incident,
why restarting did not recover it, the missing visible error, external Clear
tools recovery, and the required refusal/checkpoint/retry regression. Chapter 26
hard lint passes; negation/person-gap advisories remain. Full-book lint fails on
other chapter findings (including Chapter 27 em-dashes); no unrelated prose was
rewritten. Initial lint command could not find Go; rerun used existing
/usr/local/go/bin/go with downloads disabled.

GUI build passed with zero warnings/errors. Before staging, all 15 installed
runtime hashes and launcher matched the retention-fix receipt. Six runtime files
were atomically replaced; all 15 match the new build, including core and GUI
assemblies matched to their tested copies. Launcher, installed skills, settings,
credentials and data were preserved. No process was stopped or restarted.
Receipt: /private/tmp/cadence-image-pressure-install/staging.json. Bill owns the
restart; browser refresh is separate. Transcription remains paused and reminders
off. Earlier source-only statements below describe the pre-staging boundary.

## Image-context deadlock repair — 2026-11-02 (source only)

Bill approved checkpoint-before-overflow: refuse a new image delivery rather than
silently retire retained images or raise the model-request limits. Test's incident
report identified 33 retained tool images (~15 MiB), repeated local pre-HTTP
failures, and recovery through Clear tools. I independently traced the renderer,
actor, and GUI: image pressure was not part of token pressure, and local rendering
failure emitted TurnEnded without the ErrorOccurred event used by narration.

Implemented actor-owned image admission using the renderer's projected context
and shared count/byte accounting without allocating image payloads again. Queued
submissions reserve allowance. Human overflow refuses atomically; tool delivery
returns an explicit error with no new pixels, while the Job retains original
output. Automatic snapshot batches are checked too. Existing active/canonical
images are not silently removed. Model-request limits remain 32 images/32 MiB.
Raw invalid/unresolved library media retains its existing render-time refusal
contract; the hard render guard remains authoritative for changed files.

A sealed fixed-text image-budget exception produces a durable ErrorOccurred for
local render overflow. GUI admission preserves that safe guidance and the image
sequence rather than replacing it with a generic control error. Other exception
details stay sanitized. Clear tools remains external recovery for old over-limit
tool context; it does not remove human attachments or canonical History.

Original image allowance is now 64 MiB, including the GUI combined-source limit
and its base64 envelope. The 24-million-pixel, 2048-edge, 8-MiB prepared-image,
four-image/16-MiB interactive input bounds remain unchanged. Chapter 26 and the
inspect_image description are updated. Browser native capture keeps its own
smaller source guard; this is not a promise to accept larger decoded rasters.

Evidence: /private/tmp/cadence-image-pressure/. Eight focused image groups passed,
including real inspect_image refusal -> handoff -> successful retry using a fake
provider; queued/paused input refusal and legacy pre-HTTP visible error. A real
incompressible >32-MiB PNG prepared within output bounds. Five GUI regression
groups and the headless browser picker/draft/source-limit checks passed. Core
shared run and final-source shared rerun both passed 337/0, exit zero. Builds had
zero warnings/errors. All test jobs were collected; fixture cleanup joined.
Reminders are off at Bill's deployment decision boundary, not because the
Measure X reconstruction or remaining Chapter 26 acceptance is complete. No paid calls,
installation, restarts, live reference edits or worker delegation.

Corrections retained: initial queue enumeration mishandled default immutable
arrays; initial real-tool fixture used a PNG rejected by the decoder; broader
regressions found raw file exceptions escaping the new admission boundary; GUI
regression exposed sanitization of the actionable refusal. Each was corrected
and its affected regression rerun. Full native audible narration was not tested.

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
- Corrected screenshot requirement: CodeRhapsody-style host `list_displays` and
  `screenshot`, capturing a server display or explicit rectangle and returning PNG
  pixels. Browser-tab capture is separate, not acceptance of this requirement.
  Explicit host-screen authority and OS permission are required; no background
  monitoring, privilege elevation, or browser picker. Host implementation is pending.
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


## Paste-test installation — staged, awaiting Bill's restart

Bill authorized staging so he can test image pasting on his laptop. All 13 files
from the prior settings-dialog installation receipt still matched before staging.
The GUI host build passed with zero warnings/errors. Nine files were replaced or
added by verified temporary-file rename, preserving existing modes. All 15 files
in the Mac runtime set match the build, including SkiaSharp.dll and the macOS
native library. Other platforms' native assets were deliberately not installed.
A separate reflection/dependency-resolver smoke test decoded a synthetic 1×1 PNG
using the actual installed managed/native files; it exited zero. Launcher and
installed skills were verified unchanged. Settings, credentials, live process,
network/authentication configuration and port 8088 were not altered.

Receipt: `/private/tmp/cadence-image-install/staging.json`; build/smoke evidence
in that directory. This supersedes the settings-dialog receipt for subsequent
installation comparisons. No restart, browser reload or human paste acceptance
has yet been observed. Installed skill manuals were preserved, so this staging
does not claim new inspect_image grants. After Bill restarts, re-establish the
three-minute goal reminders if needed: their configuration is session-local.


## Laptop paste accepted; voice WAV admission implemented

After staging, Bill pasted an image from his laptop. Actual image pixels arrived
with his message and were described; Bill confirmed, “Perfect! It works!” This is
native laptop paste/send/model-understanding acceptance, not phone clipboard,
voice or screenshot acceptance. No copy of his image is added to this notebook.
The prior installation receipt remains the baseline for later comparisons.

In this session code-tools, goals and ensemble-coding activated successfully.
The agreed 180-second reminders were explicitly restored and reported enabled,
available and error-free. I incorrectly described that update beforehand as a
read-only query; it establishes current configuration, not its previous state or
a new timer firing. The coding manual is now genuinely loaded, superseding the
older live-catalog refusal. No additional restart was performed by me.

Implemented PcmVoiceAudio as an immutable value: declarations/state in
Common/VoiceAudio.cs, factory implementation in Serialization/VoiceAudio.cs.
It accepts only the canonical 44-byte RIFF/WAVE PCM header with mono, 16 kHz,
16-bit little-endian samples. Nonempty sample data must be aligned and both
container lengths must match the actual bytes. No compression, extended headers,
additional chunks, trailing bytes or browser-supplied duration are accepted.
The maximum is 960,000 samples / 60 seconds / 1,920,044 bytes including the header
(about 1.83 MiB, not an exact 1.9 MiB). The byte ceiling is checked before copying;
the bounded owned snapshot is then validated, preventing caller mutation between
validation and retained admission. Sample count and duration derive from that
snapshot instead of duplicated mutable metadata. The unchecked constructor is
private. No dependency or I/O owner was added.

Reason: server-verifiable duration is needed before paid transcription. Accepting
compressed MediaRecorder bytes plus a client duration would not establish that
bound without a suitable server decoder. Canonical PCM is larger on the wire but
needs no extra codec dependency and has a simple auditable duration calculation.
This validator is not a general WAV reader, speech detector or transcription
quality check; structurally valid digital silence is deliberately accepted.
The eventual network handler must call it before paying for transcription; that
handler, browser recording and release-to-send lifecycle are NOT wired yet.

Tests build WAV independently of product constants. Three focused groups passed:
one sample and the exact minute boundary, over-limit/empty refusal, all shorter
prefixes, all 352 single-bit header mutations, malformed/extra/odd data, hostile
32-bit lengths/rates, and immutable-copy/private-constructor checks. Build had
zero warnings/errors; final shared suite **324 passed, zero failed, exit zero**.
Evidence: /private/tmp/cadence-voice-implementation/build-wave.log,
wave-focused.log and wave-shared.log. No microphone, audio provider request,
installation, restart, new funding route or push occurred for this slice.

## Browser PCM recorder core — verified synthetic Web Audio slice

Added `voice-pcm.js`, `voice-worklet.js`, and `voice-recorder.js`. This is not yet
hold-to-talk UI or transcription. The recorder accepts an already-authorized,
one-track audio stream; it never calls getUserMedia, selects an Agent, or sends
HTTP/WebSocket requests. Each instance is one capture and owns its track,
AudioContext, nodes, port, and preparation/finish/cleanup promises.

Reasons and contract:
- Use a 16 kHz mono AudioContext and verify the actual worklet sample rate. Convert
  samples to explicitly little-endian PCM16 with bounded storage, then produce the
  same canonical WAV accepted by PcmVoiceAudio. This avoids trusting compressed
  container timestamps or adding an audio decoder just to bound paid duration.
- The default sample ceiling is 960,000 (60 seconds); a caller may choose a lower
  ceiling, never a higher one. Reaching it cancels; it does not auto-send a partial
  recording. The worklet's output is silent, not microphone monitoring.
- Finish is idempotent internally; cancellation rejects pending finish and makes
  later completion inert. Tracks stop immediately; cleanup joins owned promises
  and AudioContext closure. Native addModule is not claimed abortable.
- Review caught a mutable-container ownership edge. The recorder now takes a
  private MediaStream membership snapshot with the original track object, so a
  caller cannot remove the track to evade cleanup or add an unrelated track that
  cleanup would stop. Recheck liveness after asynchronous preparation as well.
- Unknown worklet controls, ended tracks, interrupted contexts, malformed sample
  messages, and module-load failures refuse rather than return partial audio.

Verification: `node agent/GuiTests/voice-browser.cjs` runs an owned loopback server
and isolated installed Chrome, with generated oscillator audio and muted hardware
output. The microphone API is forbidden by the fixture. Twelve browser groups
passed, including finish/cancel races, preparation gates, private track ownership,
source/context loss, limits, copied WAV bytes, and complete context cleanup. An
external C# probe references the built library and calls public
PcmVoiceAudio.FromWave on the actual browser-produced bytes. Its final receipt is
`/private/tmp/cadence-voice-browser/probe-result.json`.
The admitted native fixture had 13356 bytes, 6656 samples, and 0.416 seconds; this is a generated waveform, not recorded speech.

Two browser-response-only mutants were rejected by their intended assertions:
a mislabeled 8 kHz WAV header and copying captured samples to audio output. The
latter is checked with a real analyser while Chrome remains hardware-muted.
Neither mutation changed repository files. Modules are tested under a mount
prefix; the real GUI consumer also checks their embedded and mounted URLs and
JavaScript content types. GUI build had zero warnings/errors; all five GUI
regression groups passed, exit zero. Final native job and GUI jobs exited zero;
owned browsers, HTTP servers, tracks and contexts were closed. Evidence is in
`/private/tmp/cadence-voice-browser/` (`positive-final.log`, `probe-result.json`,
`mutant-*-final.log`, `gui-build-final.log`, `gui-regression-final.log`).

Fixture correction: the first run incorrectly used page.evaluate(clean), referring
to a browser function as a Node variable. The preceding waveform assertions had
passed; corrected it to execute the helper inside the page. Preserve this as a
fixture mistake, not an audio failure. Added checks after the ownership review
were rerun on final source. Core code did not change; carry the existing 324/0
shared C# result rather than imply it was rerun for these assets.

Limits and next integration: no native microphone permission, phone browser,
real speech, transcription, keyboard/button gesture, Agent/socket routing, or
pause/TTS interaction is certified by this slice. These modules are not connected
to the composer and have not been installed. The hold controller must own late
permission grants and a capture lease for the original Agent/socket. Proposed
server flow is transcription followed by an automatic browser validity check and
commit, not direct send from an HTTP completion callback. It must retain the
release-to-send-once contract without an editable draft or extra Send button,
redact audio from traces, cancel/join owned HTTP work, and keep transcription
funding separate. No paid call, microphone access, live restart, or push occurred.

## Separately funded transcription owner — initial focused verification

The optional root-owned VoiceTranscriber now compiles and has three passing local
HTTP-handler test groups, alongside the three existing PCM admission groups.
The client requires an explicit API key, uses fixed gpt-transcribe multipart JSON
transcription, allows one active request and never retries or submits chat. The
root owns cancellation, joining, transport disposal and both relationship links.
A task admission barrier makes the request joinable before HTTP starts. A held,
deliberately cancellation-ignoring transport demonstrated that destruction waits
for completion and refuses its late transcript rather than returning success.

Tests inspect endpoint, Bearer payer, all four multipart fields and exact WAV
bytes; verify physical duration remains separate from reported usage; reject
missing/empty/duplicate/oversized text and malformed usage; prove recovery and no
401/503 retry; check diagnostic key masking, refused attachment preserving the
existing owner and caller-owned transport, and successful root detach/disposal.
Evidence: /private/tmp/cadence-voice-transcription/build.log (initial build),
build-tests.log and focused.log (six groups, zero failures, exit zero).

This is initial fixture evidence, not complete acceptance. Caller cancellation
recovery, deadline behavior, streamed response ceilings and disposal-failure
cascades still need targeted coverage. No GUI funding configuration, microphone,
audio provider request or automatic send is connected. No live install/restart.
The next controller must retain the original Agent/socket lease and revalidate
an automatic commit; transcription success alone cannot authorize a chat send.
The initial test-file lookup used the wrong filename; corrected to the actual
VoiceAudioTests.cs. A clipped inspection is not counted as complete review.


## Handoff releases media from model context

Bill reported an uploaded image surviving handoff and clarified the intended
boundary: GUI history may keep displaying old media; future model requests must
not keep its payload. The shared reducer now excludes BlobPart along with tool
traffic, preserving accompanying text/part identities and dropping empty entries.
No GUI or log-retention behavior changed. Source media files are not deleted.

Five focused memory-selection/handoff groups pass. The extended replay fixture
covers image/audio/video/PDF, mixed and media-only messages, original canonical
events, snapshot-only restoration and event replay, and provider rendering for
OpenAI, Anthropic and Gemini without reopening deliberately absent media paths.
Evidence: /private/tmp/cadence-handoff-media-{build,tests,shared}.log. Shared suite
is still pending at this entry; no deployment is claimed.

My first test incorrectly checked canonical events after an explicit RetainLog:0
save. The save trims its retained in-memory diagnostics as well as its snapshot;
that is existing retention, not handoff deleting canonical media. Moving the
preservation assertion before save and separately testing event replay corrected
the fixture. Bill clarified that retained GUI history is allowed. I must not
conflate displayed history, retained diagnostics and active model context.

Final verification: shared suite 327/0, exit zero; GUI build zero warnings/errors.
Dev staging verified all 15 runtime files, six replacements, unchanged launcher;
receipt /private/tmp/cadence-handoff-dev/staging.json. Dev was launched using its
existing launcher, separately from working port 8088. No claim of authenticated
phone/tunnel or human handoff acceptance follows from local HTTP readiness.


## Explicit image attachment fallback — browser component implemented

Bill reported Android copy/paste did not attach an image. Existing paste accepts
only file items exposed by the clipboard event; the exact native failure is not
yet diagnosed. Added a labeled Attach images button and native multiple-file
picker without camera capture or new clipboard permission. It shares ImageComposer
file admission, readers, previews, limits and addressed acknowledgment-based send.
A picker lease captures the connection and Agent; selection changes (including
away-and-back), disconnect, clear and page teardown invalidate it. Same-file
reselection resets the input. Accompanying composer text remains unchanged.

`agent/GuiTests/image-picker.cjs` runs real Chrome with synthetic PNG bytes and a
phone-sized viewport, checking preview/remove, exact bytes plus text, send once,
same-file reselection, stale switch/disconnect/connection replacement and whole
selection count/type/empty refusal. All three groups pass; GUI build has zero
warnings/errors. Logs: /private/tmp/cadence-image-picker.log and
/private/tmp/cadence-picker-build.log. Initial notice-count assertion included a
prior selection notice; corrected by measuring the delta. Initial shell ended
with cat and masked Node failure; rerun used set -e and exited zero. No product
failure was hidden as passing evidence.

This is component-level browser evidence, not native Android picker/clipboard,
full-page layout or live server picker acceptance. No dev staging/restart yet for
this slice. GUI regressions are running separately; collect their actual result.


### Picker verification and working-instance staging

Bill subsequently confirmed Android paste works in dev: do not describe native
Android paste as unsupported. The explicit picker remains an optional alternative.
All five GUI regression groups passed. Full-page Chrome now passes picker plus
text with real server acknowledgment, paste, phone-size geometry and lifecycle
cases (/private/tmp/cadence-picker-full/browser-final.log). The first full-page
picker run timed out at admission using a hard-coded PNG. Replacing it with a
browser-canvas PNG and reporting explicit refusals passed; the precise original
PNG defect has not been established. Fixtures stopped/joined with exit zero.

At Bill's explicit request, rebuilt and staged the working Cadence installation,
not restarted. Build zero warnings/errors; all 15 runtime hashes verified and
launcher unchanged. Receipt: /private/tmp/cadence-picker-install/staging.json.
Preflight initially guessed receipt key files instead of verified and safely
stopped before build/replacement; corrected preflight passed. No settings,
credentials or installed skills changed. Bill will restart and test phone paste.

### Dev recovery after attachment layout rebuild

Bill confirmed Android paste works in updated dev, and reported the separate attachment grid item squeezed the composer. Grouped attachment and Send in `compose-actions`; eight static real-Chrome geometry checks passed, GUI build clean. Full phone layout still requires Bill's assessment.

Dev restart then failed with InvalidDataException. Read-only decoder checks found both snapshots and all individual journal records valid; journal sequence 32 was missing between 31 and 33. Primary snapshot at 40 retained contiguous events 1–40, and all 39 existing journal events matched their snapshot counterparts. Missing event was `runtime_configured`; the omission's cause remains unknown. No loader validation was weakened.

With Bill's explicit repair permission, preserved the original journal privately alongside dev state as `save.json.journal.before-sequence32-repair.jsonl`, inserted only exact snapshot event 32, and preserved every other journal byte. Decoder then accepted all 40 records. Unchanged dev launcher restarted successfully; local HTTP 200 served grouped controls, and new runtime-configured event 41 was present in the contiguous journal. Working Cadence instance was untouched. This proves recovery and this startup's append, not resolution of the original omission. Evidence: /private/tmp/cadence-dev-decode/ and /private/tmp/cadence-picker-dev/ (no conversation payload copied there).

### Cancellation recovery and restart-boundary review

Added test-only `CancellationRecovery`: caller cancellation and the internal deadline each propagate to the fake transport, preserve OperationCanceledException, make exactly one HTTP attempt, leave the client owned/reusable, and permit a following successful request. Joined cleanup disposes the handler and detaches. Focused voice suite 7/0, exit zero; build and test evidence in /private/tmp/cadence-voice-transcription/{build-cancellation.log,cancellation-tests.log}. No production transcription change or paid request. Prior shared 327/0 remains earlier evidence, not a new full run. Streamed-size and disposal-failure tests remain unfinished.

Restart procedure review found my dev restart script waited for listener closure rather than full old-process exit. Host code finishes serving before Agent shutdown, so listener closure is not a joined persistence barrier. Overlapping writers are a plausible explanation for the missing runtime-configured event, not an established cause. Future dev restarts must wait for the identified process to exit before starting a replacement; do not infer exit from a closed port. Current dev startup appended event 41 correctly. No new dev restart or working-instance change during this review.

### Stream bounds and cleanup privacy

Added forward-only, unknown-length response tests: over-limit streams consume exactly the 65,537-byte sentinel allowance then refuse; oversized declared lengths refuse before reading; exact 65,536-byte valid JSON succeeds; all response streams dispose; refusals do not retry. Added root teardown failure coverage ensuring the transcriber detaches and the joined task retains failure. Initial nine focused groups passed.

An additional synthetic-key-bearing disposal negative control exposed a real defect: cleanup aggregation retained the key after owner detachment removed it from diagnostic redaction. The negative run failed its intended assertion (8 passed/1 failed), evidence cleanup-negative.log. Cleanup now discards only secret-bearing exception objects before key clearing, retaining a static failure and aggregate status rather than retaining private inner exceptions. No actual credential was used. Final focused and shared verification launched as job 50; results pending at this entry. Evidence under /private/tmp/cadence-voice-transcription/. No installation, restart, microphone use or paid call.


### Hold ownership — first controller slice

Added one-shot `VoiceHold`, not yet wired into the composer. It owns the permission
request and recorder, returns audio only after a normal release, and does not pick
an Agent or send a message. Release before permission/preparation completes cancels;
late streams are stopped instead of recording after the gesture. Permission cannot
be forcibly aborted: cancellation joins its eventual result, with immediate
recorder cancellation when available. Caller/UI lifecycle wiring remains pending.

First gated Chrome test failed in two seconds because cancellation read its mutable
release promise after awaiting preparation, creating a self-wait on early release.
Corrected by snapshotting the finish operation that already exists when cancellation
begins. Nine groups now pass: early permission/preparation release, idempotent cancel,
normal exactly-once finish, late finish suppression, permission refusal, observable
cleanup failure, device cancellation, and real PCM recorder/WAV/resource cleanup
using generated audio. Evidence `/private/tmp/cadence-voice-hold/{before-fix,after-fix,native-final}.log`.
Owned Chrome/server cleanup joined. No actual microphone, server lease, paid
transcription, phone gesture, UI integration, installation or restart is claimed.

Collected prior transcription verification: nine focused groups and shared **330/0**,
exit zero (prior job 50). Main composer rebuild installed and Bill restarted;
Bill reports much better layout, phone buttons stack vertically but are acceptable
for now. The hold-controller module is newer source-only work.


### Explicit host transcription payer

Added GUI host `--transcription off|openai-api-key`, mirrored by `EN_TRANSCRIPTION`,
default off. Only explicit enabled selection reads `EN_TRANSCRIPTION_API_KEY`; no
model key or subscription fallback. No key flag, no enabled-by-key-presence, and
explicit `off` overrides environment. The optional root-owned VoiceTranscriber
uses its existing fixed endpoint/model and lifetime. Dedicated key excluded from
skill/MCP startup variables; arbitrary host-process environment access remains a
separate trust boundary. No live launchers/keys were changed.

WebSocket connection metadata contains only enabled/payer/model; microphone UI and
socket audio lease remain unimplemented. Real-host executable test
`agent/GuiTests/voice-funding.cjs` passed five cases (default off with ambient speech
key, explicit enable, explicit off override, missing dedicated key despite model
key, unknown payer). It used synthetic keys, fresh isolated homes, literal loopback
provider, zero provider requests, and full process exit joins between cases.
All five existing GUI regression groups passed; both builds clean, whitespace check
passed. Evidence `/private/tmp/cadence-voice-hold/{gui-funding-build,gui-tests-build,funding-tests,funding-gui-regression}.log`.
No microphone, paid request, installation or restart. This is payer configuration,
not a completed voice interaction. Shared core 330/0 evidence remains applicable;
no new full core run claimed for this host-only change.


### Reuse the existing OpenAI key through explicit file selection

Bill asked whether existing API keys could be used. A bounded presence-only check
confirmed the existing settings has a nonempty top-level `openai_api_key`; no value
was printed, copied into evidence, or changed. The dev launcher currently uses an
isolated environment and Codex funding, not that API settings file. Neither launcher
was changed in this slice.

Added `--transcription-key-file` / `EN_TRANSCRIPTION_KEY_FILE` to the explicit
`openai-api-key` route. The composition root reads at most 1 MiB through one opened
file generation, validates bounded strict JSON, refuses duplicate root fields or
missing/nonstring/empty OpenAI key, and never writes the selected file. This is
explicit host authority, not an Agent Sandbox read or automatic credential discovery.
A selected file wins over the dedicated environment key, with no fallback on failure;
when off, it is not read. It is a startup snapshot, not live rotation/refresh.

GUI build clean. Thirteen executable real-host funding cases passed, including a
valid selected file with an invalid environment key (proves actual file selection),
environment-selected file, disabled missing file, malformed/duplicate/missing/wrong-
type/oversized input with a valid fallback key present (all refuse). Synthetic key
files unchanged; zero provider calls and full process joins. Evidence
`/private/tmp/cadence-voice-hold/key-file-build.log` and `key-file-tests.log`.
Existing five GUI regression groups remain valid from the preceding metadata change;
no new core-suite run claimed. No installation, restart, paid call or microphone.
The actual existing API key has not yet been configured into either live host.


### Connection-scoped voice lease foundation (not yet socket-wired)

Added internal GUI `VoiceLease`, with declarations in Common and implementation in
Server. A future socket operation owns/disposes it; it borrows the actual target
Agent and retains the capture sequence, cancellation lifetime, transcription task,
actor-admission task and one idempotent destruction task. It does not reconstruct
target identity from a string, mutate model funding or submit on transcription
completion. Start requires the explicit root transcriber; duplicate audio refuses.
Commit requires the same Agent reference/sequence, successful transcription, a live
lease and an unused commit. It marks consumption before PostAsync; cancellation
cannot retract an admitted message, and failed/uncertain commits are not retried.
Disposal invalidates immediately, cancels and joins transcription/admission before
releasing its CTS; request errors remain on the request task.

A separate `agent/VoiceLeaseTests` internal executable uses explicit friend access to
the existing test-only HTTP transport seam and the internal GUI lease. This does
not widen public runtime APIs or change the public GUI-consumer test contract.
Project restore used existing local packages only. Six grouped checks pass: paused
real Agent PostAsync admission, not just enqueue; readiness and duplicate audio/
commit refusal; actual target and sequence validation; cancellation after ready;
connection cancellation before provider work; uncooperative HTTP joined by disposal
with late result inert; repeated disposal; provider failure without retry/commit;
root transport cleanup. Groups contain multiple assertions. No real provider/mic.
Evidence `/private/tmp/cadence-voice-hold/lease-build-final.log` and
`lease-tests-final.log`. Corrected two fixture failures: delegate name hid inherited
Send under warnings-as-errors, and HTTP 401 correctly throws HttpRequestException,
not the fixture's initial InvalidOperationException expectation. Neither required a
production-policy relaxation.

Initial public GUI regression ran from an older output directory and is not evidence
for this new assembly. The freshly rebuilt run passed all five GUI groups and exited
zero; evidence `lease-gui-build.log` and `lease-gui-regression-final.log`.
Next remains receive-loop protocol, selection/disconnect invalidation and joined
socket teardown, then browser wiring and real end-to-end synthetic fixture. The
lease object alone is not WebSocket/gesture/paid-provider acceptance. No installation
or live restart. Existing core330/0 evidence carried forward; only internal test
friend access changed in core, no core behavior change in this slice.

## Voice WebSocket lease protocol — verified local slice

Replaced the socket peer tuple with a connection-scoped owner retaining the existing queue, actual selected Agent and cursor, plus one VoiceLease and joined notification task. Central Select cancels the current lease, including automatic tree recovery and away/back selection. The receive loop admits addressed `voice_begin`, `voice_audio`, `voice_cancel`, and `voice_commit` with per-connection monotonic begin sequences; explicit funding must already be enabled. Audio is canonical bounded PCM WAV only, base64 bounded before decoding and validated before calling the paid client. Ordinary controls cannot carry audio. Failed or cancelled audio cannot commit, and a ready result alone does not send. Only the matching connection/Agent/sequence commit invokes actual Agent.PostAsync once; its acknowledgment means actor admission, not model completion. A commit already admitted cannot be undone by a later cancel. Ready/failure/acceptance frames carry metadata, not duplicate transcript text. Raw audio is excluded from GUI trace projection.

Transcription notification runs without blocking receive, so cancel and selection remain responsive during HTTP. Another begin refuses while old work is still closing rather than queueing more paid requests. Disconnect removes the peer, cancels its socket lifetime and joins browser/sender/voice work through nested finally cleanup. A noncooperative transport may delay teardown until it returns; cancellation is not a claim that the transport was killed.

Evidence: /private/tmp/cadence-voice-hold/protocol-build-final.log, protocol-tests-final.log, protocol-gui-build.log, protocol-gui-tests.log. All ten combined internal-lease and real WebSocket groups passed; freshly rebuilt public GUI regression suite passed all five groups, final job exit zero. Wire cases exercise ready/commit separation, exact-once/monotonic refusal, wrong Agent/sequence/connection, malformed/oversize/wrong-MIME WAV, audio smuggled into ordinary controls, cancellation during held HTTP, no late ready, away/back selection invalidation, disconnect cancellation, joined host teardown and trace absence of the synthetic WAV. All listeners/HTTP transports joined/disposed. Existing shared core 330/0 evidence stands; no fresh full core suite claimed for this GUI slice.

Limits: endpoint and owner are source-only, not installed. Browser microphone/hold/commit wiring, pause ownership, native phone/keyboard and real transcription acceptance remain unfinished. Tests use real Agent/HTTP-client/GUI/WebSocket paths with a fake transcription handler and synthetic WAV; no microphone, paid request, credential change, live restart or deployment occurred. This completes the server-protocol slice, not voice interaction or the chapter. Next requested interaction: scoped actual-pixel page capture, reusing the established image path.

## Connected-page capture core — real pixels and scoped ownership

Implemented source-only `page-capture.js` as a one-shot capture owner, not yet wired into the composer or model-tool surface. It invokes getDisplayMedia synchronously from the caller's gesture, requests video/no audio and current-tab hints, and requires capture-handle registration/read support. A fresh handle exposed only to the current origin must match BOTH handle and origin on the selected browser track. Browser-tab metadata alone cannot prove this is the connected page. Another tab, window, monitor, wrong origin or missing handle refuses before frame extraction. Android/unsupported browsers receive explicit refusal and a manual screenshot/attachment fallback, not DOM reconstruction.

The owner snapshots original track membership, uses its own muted video/stream, registers a frame callback before play (a static tab may deliver only one frame), bounds capture at 24 million pixels before canvas allocation and PNG output at 32 MiB, and returns a real File with dimensions. Existing server image preparation remains responsible for resizing/delivery on later integration. A capture represents the visible tab content, not a stitched full-scroll screenshot. Cancellation stops owned tracks immediately and joins eventual native picker/encoder completion; it does not claim those native APIs are abortable. Handle changes/device end/cancellation make late results inert. Temporary video/canvas resources are detached/released on success and failure. Pixel/byte limits are not a global browser CPU/memory isolation guarantee.

Evidence: /private/tmp/cadence-page-capture/. Installed Chrome capability probe made no capture request. Ten real-Chrome synthetic groups in synthetic-final.log verify actual red/blue video-to-PNG pixels, wrong identity/origin/surface refusal, unsupported/denied admission, duplicate picker/cancel/late permission, pre-frame playback cancellation, encoding-time cancel/identity/device loss, private original track membership, and pixel-metadata/failed-encoding/PNG-byte limits. Resource-boundary negatives use injected metadata/encoder results, not a claim of decoding a 25-million-pixel native stream. Browser-only identity-check and late-result-check mutants both failed their intended assertions (mutant-identity.log and mutant-late-result.log). No production source was overwritten for those mutants.

One bounded NATIVE getDisplayMedia attempt used a fresh isolated Chrome profile, an owned synthetic red/blue page and Chrome's this-tab auto-selection flag; it did not capture the live conversation, desktop, another window or microphone. Native capture returned a 640×360 PNG (9,040 bytes), verified handle/origin, red and blue sampled pixels and busy:false after cleanup; native.log, exit zero. This proves the installed desktop Chrome capture path in the isolated fixture, not Android, another browser, authenticated-tunnel acceptance, manual picker usability or screenshot delivery to a model. Owned browsers/listeners joined. Earlier five public GUI regression groups passed, build zero warnings/errors; the final rebuilt run adding explicit root and mounted-path HTTP checks for page-capture.js also passed all five groups, exit zero (gui-build-final.log and gui-tests-final.log).

Corrections/limits: initial inspection used obsolete asset/project paths; corrected to agent/Gui/wwwroot, agent/Gui/Ensemble.Gui.csproj and agent/Examples/Gui/Gui.csproj. Whitespace checks passed independently; missing grep files did not establish inspection or build success. No paid provider request, real key use, installation, live restart or UI capture control occurred. Next is integrating capture with the existing addressed image draft/send path without further squeezing the phone composer, followed by human desktop capture assessment and explicit phone fallback. Voice UI integration, live transcription checks, executable exercises/graders and the full chapter body remain unfinished.

## Connected-page capture integrated with addressed image drafts

The header camera control sits beside Settings, not in the composer, so attachment and Send retain their existing behavior and input width. The module loads before the control is enabled; the click invokes capture before awaiting, preserving native transient activation. A second click cancels visibly. Capture retains the original socket and Agent, and Agent selection (including away/back), disconnect, shutdown state and page departure invalidate admission before cleanup. Pending capture blocks both the Send button and form submission; accompanying text is retained. Successful capture enters the existing removable image draft and normal addressed image_prompt path, with the existing server preparation, acknowledgments and uncertainty rules. Unsupported capture explains manual screenshot attachment. Native picker/encoding cancellation remains a join of eventual completion, not a claim that native APIs are abortable.

Evidence: /private/tmp/cadence-page-capture/ui/. The final browser-verified.log has seven full-page groups passing: real PNG color/dimension checks, caption plus server admission, explicit cancel and premature form-submit refusal, switch-away/back, disconnect/reconnect, controlled shutdown-frame cancellation, unsupported fallback and phone/desktop composer-width invariance. These are isolated Chrome/loopback checks, not native phone or human layout acceptance. images-regression.log records four existing image-picker/paste groups passing; gui-final.log records all five freshly rebuilt public GUI groups passing. Builds and whitespace checks passed. Previous capture-core ten-group/two-mutant/native evidence remains applicable; this slice changed UI integration, not capture-core policy.

The browser-only selection mutant removes both Agent-change cancellation hooks from the served test copy, never production files. On a fresh host it failed the exact assertion 'Agent change must invalidate old capture even after selecting back'; mutant-selection-fresh.log and mutant-host.log. Earlier mutant attempts failed setup by redundantly changing/waiting for a model on a busy reused host, so those are NOT successful negative-control evidence. A fresh host and conditional model selection corrected the fixture without relaxing server policy. Initial pixel equality was also too strict: browser video conversion produced red [255,1,0] and blue [1,0,255]. The corrected test permits only small low-channel variation, while checking dominant colors, alpha and dimensions.

One bounded native FULL GUI test used a separate fresh host/profile and Chrome's this-tab auto-accept flag, with synthetic red/blue markers on the owned GUI page. The actual camera click produced a 960x640 PNG (133472 bytes), with expected sampled pixels, retained caption, exactly one addressed image submission, real server preparation/admission acknowledgment and cleared draft. native-ui-verified.log records exit zero. This establishes native user-activation/capture-to-admission integration on installed desktop Chrome, not a paid model turn, manual picker/tunnel usability or Android display capture. No live conversation, desktop, other window or microphone was captured. Both final fixture hosts, providers, Agents and browsers joined; root job 140 exited zero.

No live installation or restart occurred during these checks. Dev staging for human assessment is a separate next operation. Voice hold UI, live transcription, external exercises/graders and full chapter prose remain unfinished; reminders remain at 180 seconds.

### Dev staging after capture integration

The GUI application build passed with zero warnings/errors. Verified all 15 dev runtime files and the launcher against the prior dev receipt before replacement; preflighted every build output. Confirmed the actual dev listener identity, requested graceful termination and waited for the whole old process to exit before staging or starting another. Six files changed; all 15 runtime files matched the new build afterward. Launcher, authentication, skills, settings and credentials were unchanged. Transcription remains off under that unchanged launcher. Dev restarted successfully, returned local HTTP 200, exposed the camera control and served app.js/page-capture.js byte-equal to source. Main port 8088 was not restarted or modified. New authoritative dev receipt: /private/tmp/cadence-page-capture/dev-install/staging.json; build.log, launch.json and readiness.json retain the separate build/stage/launch/readiness evidence. No live screenshot or provider request was made to verify staging. Human desktop capture-picker/tunnel assessment remains pending; it does not block remaining voice UI implementation.

## Agent-requested capture — transport groundwork (not yet advertised)

Bill observed that Dev could not see a capture tool. The staged camera is a human-operated attachment path, not an Agent tool; a restart does not add a missing tool declaration. Work therefore moved to the existing browser MCP bridge before returning to voice UI integration.

Added a capture-aware McpTunnel constructor bound to the actual requesting Agent, one pending gui_capture RPC identity, read-free tool-image capability checks before browser dispatch, explicit cancellation invalidation, duplicate/stale media refusal, and bounded PNG reply preparation through Agent.PrepareImage. Only a matching pending request may deliver capture media in this mode. The transport checks PNG MIME/signature, encoded/source bounds, and native preparation limits, then emits actual prepared PNG bytes with dimensions as typed MCP content. The server describes these as browser-supplied pixels; surface verification must still come from the trusted browser capture owner and its fresh handle/origin check. Ordinary string-ID RPC traffic remains valid. A failed outbound write releases its reservation; the reply channel now correctly declares multiple producers because local refusals and socket replies both write it.

Evidence: /private/tmp/cadence-capture-tool/test-build-final.log (zero warnings/errors), tunnel-final.log (five capture transport groups), gui-final.log (five public GUI groups); verification job 179 exited zero and is joined. The outside-assembly tests use real Agents, native synthetic red/blue PNG preparation (4096x64 to 2048x32), McpContent decoding and public media validation. They cover metadata-only/missing/invalid payload refusals, duplicate/stale images, concurrency, explicit cancellation, subsequent recovery, unsupported-model refusal before browser dispatch, ordinary RPC compatibility, disposal and cancellation under actual output backpressure. No model turn, paid request, native screenshot capture or live deployment occurred in this slice.

Failures and corrections: the first test registration guessed args.Contains rather than the existing list-pattern dispatch; fail-fast was enabled too late and a subsequent build ran without the test registration. It is not test evidence. Compilation caught fixture Count-versus-Length and synchronous root.Destroy mistakes. The real negative test then exposed an escaping KeyNotFoundException for missing capture data; the transport now returns a paired bounded tool error. Review also added a JSON-number guard before TryGetInt64 so ordinary string IDs do not throw. Original logs remain beside the passing logs.

Remaining before Dev can discover/use gui_capture: bind the capture-aware tunnel in the host; permit only bounded matching capture replies through the frame-size guard; redact image bytes in MCP traces; invalidate on selection/connection/authority changes; implement asynchronous browser request handling and dedicated trusted human confirmation (programmatic gui_click cannot approve); publish the tool/manual only with that working path; test actual model-turn History/media delivery and native consent. The new mode is not wired into the live host and is intentionally not advertised yet. Existing two-argument callers retain ordinary transport behavior. Use canonical lowercase agent/gui paths when eventually staging Git changes; new case-insensitive-working-tree paths need care. Both live instances remain unchanged. Three-minute reminder configuration was not changed.

## Agent-requested capture — host admission, selection and trace privacy

Connected the host to the capture-aware tunnel using the actual relay Agent. The connection state borrows that tunnel while the socket scope owns disposal; removal detaches the borrowed link. Selection updates invalidate any pending capture and return a paired tool refusal, including selection away/back or reselecting the same Agent. New capture requests refuse while the displayed Agent is not the requester. The selection replay is queued before this transition and will drive browser-side consent/capture invalidation in the next slice; no independent cancellation-notification task was added.

Only a matching pending capture RPC may exceed the ordinary 1 MiB control ceiling, within the existing bounded image envelope. Wrong-target and stale replies refuse; matching capture content still passes bounded PNG/native preparation. Trace projection now recognizes MCP mimeType and treats string-valued data fields as opaque, including malformed image blocks. This is media-field redaction, not a promise to recognize arbitrary secrets in arbitrary text.

Evidence: /private/tmp/cadence-capture-tool/host-build-final.log (zero warnings/errors), host-tests-third.log (five transport groups plus two real HTTP/WebSocket groups), host-gui-final.log (five existing public GUI groups). Job 204 exited zero and is joined. Tests verify actual noisy PNG bytes above 1 MiB, pixel equality after preparation, wrong-target/duplicate/stale refusal, selection invalidation, unchanged ordinary-message ceiling, malformed-payload trace privacy and recovery. Direct MCP requests use an explicitly granting primary fixture skill and fake browser discovery; this is not production advertisement, trusted-click/native capture or model-turn History acceptance. No provider request, installation or live restart occurred. Disk remained above 1 GiB.

Fixture failures: the first host test assumed browser registration implied a claimed client and dereferenced null; the corrected fixture declares/discovers its scoped capture tool and asserts the grant. The second awaited initialization before servicing the handshake replies needed to complete it; its bounded run eventually failed and exited without forceful termination. The passing fixture starts initialization, services real initialize/tools-list RPCs, then awaits readiness. Neither failure justified weakening production skill authority. Earlier failure logs remain beside the passing logs.

Next: asynchronous browser dispatch with a dedicated trusted human confirmation (gui_click cannot approve), explicit cancel/unload/disconnect and selection invalidation, actual pixels delivered to the requesting RPC, and only then production tool/manual advertisement and model-turn/native acceptance. Dev still has only the human camera until that path is complete and staged. Voice UI, remaining native/provider checks, exercises/graders and chapter body remain unfinished.

## Agent-requested capture — trusted consent and asynchronous dispatch (source only)

Added `capture-consent.js`, a page-owned dialog controller using the shared actual PageCapture owner. A trusted click is required to start capture; programmatic `gui_click` cannot approve it. The operation binds the actual socket, Agent ID and RPC ID, retains ownership through capture/encoding cleanup, refuses competing manual capture or another open dialog, and returns typed PNG content or an explicit tool error. Labels use the displayed Agent name; routing keeps the original ID. Decline, exact-RPC cancellation, session close, connection loss and late results are tested. No human draft insertion or extra Send step belongs to this tool result.

`app.js` now preloads consent separately from manual capture, rechecks selection/connection after asynchronous loading, and cancels consent through existing selection/disconnect/shutdown/page-hide hooks. Manual camera start also requires a trusted event; its existing `onclick = capturePage` correctly forwards the event. MCP capture replies are asynchronous, leaving normal DOM/speech request handling responsive. Replies remain on the receiving socket and are never replayed onto a replacement. Tunnel disposal attempts an ordered browser-session-close notification; a closed host queue means socket teardown owns frontend cancellation. No claim of forcibly aborting a native browser picker or canvas work is made.

Verification: seven real-Chrome component/dispatcher groups passed with synthetic capture and actual generated PNG bytes, including `gui_click` refusal, exact-request cancellation, uncooperative cleanup, session closure and replacement-socket suppression. Seven C# capture/real-WebSocket groups and five public GUI groups passed after rebuilding; asset delivery includes the consent module at root and mounted paths. Build and whitespace checks passed. All jobs collected through 234; fixtures and browsers joined. Evidence: `/private/tmp/cadence-capture-consent/{dispatcher-final.log,build-final.log,host-final.log,gui-final.log}`.

Corrections retained: the initial component fixture omitted an HTML content type, leaving the expected dialog absent; fixed in the fixture. Checked application edits refused mismatched import/listener anchors without writing app.js. I incorrectly inferred that the listener dropped the event; actual `onclick = capturePage` forwards it, so only my overly narrow assertion was wrong. An asset edit also refused because the module name appeared in separate root and mount checks; both exact sites were then updated. These are fixture/edit mistakes, not product failures.

Boundary: production browserTools still does NOT advertise gui_capture, and the production skill manual is unchanged. The application integration has syntax/build and component/server checks but still needs a full-page Agent-request flow, then catalogue/skill publication and model-turn/native verification before staging. Neither live instance was updated or restarted. No paid provider request, microphone use or live-page capture occurred. Voice UI, chapter exercises/graders and full chapter prose remain unfinished.

## Handoff continuation regression and Android acceptance

Bill reported Android image paste now works here after reinstalling the Chrome
web app. This is human Android acceptance after reinstall, not a diagnosis of
the installation/clipboard defect. Goal reminders remain off at his request.

The repeated handoff promises exposed a missing continuation contract. Inspection
found that the renderer projected HandoffPart as bare note text after the reducer
removed the successful tool receipt while retaining future-tense narration. A
bounded real-Agent/loopback-provider regression reproduces repeated handoffs when
that completion signal is absent. Before the repair it failed the intended
continuation assertion, not compilation or setup. The renderer now labels typed
checkpoints as completed and says to continue from the note rather than repeat
the handoff merely to perform it. Canonical note bytes and reduction policy are
unchanged; later legitimate handoffs remain allowed.

Evidence: /private/tmp/cadence-handoff-loop/{build.log,before.log,build-after.log,
after.log,selection.log}. Focused continuation 1/0 and related selection/replay/
interruption 5/0 passed. Replay test checks completion projection across provider
renderers. The prior missing-label version is the negative control. Chapter 14
now specifies this bounded test and distinguishes it from historical grades.
This is deterministic wire-contract evidence, not live-model causation or a
guarantee against all repetition. Shared regression running; see shared.log and
subsequent result before claiming acceptance. No install/restart/provider calls.

The previously outstanding capture fixture build log model-build-fourth.log
reports build success, zero warnings/errors. Native replay remains unrun.

Shared first run: 329 passed, 2 failed. MemoryCacheTests and CacheMarkerTests
looked up rendered checkpoints by exact bare-note text; both now require the
completion prefix and exact note suffix, retaining counts, ordering and canonical
state assertions. No cache implementation change was needed. Final rebuilt shared
run pending in shared-final.log.

Final rebuilt shared suite: 331 passed, zero failed; build zero warnings/errors.
Evidence: build-final.log and shared-final.log in the same handoff-loop folder.
No installation or restart; live behavior of the repair remains untested.

## Large tool-result retention correction — source only

Removed the automatic T/8 result and T/16 call/stub budget ladder and its
presentation watermark. Explicit keep decisions are no longer overridden by a
later size cut. Opt-in once-only stubbing, tool output limits, handoff, and replay
of already-recorded redactions remain distinct and unchanged. Chapter 14's
TL;DR, contract, code example and regression requirements now teach this rule;
Chapter 21's historical additive-ladder claim is labelled historical.

The strengthened real-Agent fake-provider tests failed against the old code
(including explicit keep of a 120 KB result) and pass after removal. Coverage
includes later requests, small companion results, opt-in stubbing without keep,
and zero-log save/restore. Focused 2/0; full shared suite 331/0, exit zero;
GUI-test build zero warnings/errors. Evidence: `/private/tmp/cadence-tool-retention/`
(`before.log`, `after.log`, `build-after.log`, `shared.log`, `gui-build.log`).
An initial post-edit compile failed on four leftover budget API test references;
those were corrected before the passing build/tests. Historical Chapter 14
grader receipts are not new acceptance of this changed contract; frozen graders
and endpoints were not rewritten.

Removed the screenshot fixture's 8 MiB budget workaround; native replay under
the normal default remains unrun. No installation, restart, paid requests or
reminder changes. The separate request to expose only one size per memory band
and remove configurable high/low watermarks is still pending.

## Speech pause after normal turn completion — 2026-11-01

Bill reported actual `paused` status while away and manually resumed, then proposed
that speech ending after normal model-turn completion might strand the pause.
I tested that boundary with `agent/GuiTests/speech-after-turn.cjs` and a
content-free `/fixture/activity` endpoint in the existing isolated crossover host.
The browser controls only native speech; the Agent, streamed fake-provider turn,
WebSocket pause controls, and activity/paused status are real. No live host or
private conversation was inspected, paused, resumed, or restarted.

The test holds speech across a real final answer and checks the actor is both
Idle and Paused before finishing speech. With fresh hosts for each scenario:

- Normal native completion: PASS; actor ends Idle, not Paused.
- Native speech stops without its callback: PASS; existing watchdog releases
  the automatic pause after the turn is idle.
- Socket disconnect/reconnect while that speech holds the pause: FAIL at the
  intended assertion; speech is cleared but actor remains Idle and Paused.

This reproduces the previously suspected disconnect ownership leak, not the
normal idle transition itself and not proof of the live incident's cause. Source
`resetAgentActivity()` clears browser pause intent during disconnect without a
server-owned automatic-pause lease to release. Do not fix by blindly unpausing on
reconnect: manual pauses and other owners must remain authoritative. Production
code is unchanged; the disconnect scenario deliberately remains a failing
regression pending an ownership-safe repair.

Evidence: `/private/tmp/cadence-speech-after-turn/`, normal-result.json and
results.json, scenario browser/host logs and fixture pointers. Normal job60
exit0; exceptional runs job62 exit1 for the reproduced disconnect assertion;
all three hosts joined exit0. GuiTests build succeeded with zero warnings/errors
(build.log). No shared-suite rerun or native-phone speech acceptance claimed.

Setup mistakes are retained under setup-first through setup-fifth: I initially
selected invalid/non-media model IDs and briefly masked a timeout with a wrong
JavaScript diagnostic name. None count as product failures. The final test selects
the capability-table model `claude-3-haiku-20240307` explicitly; recommendation
menu entries are not capability facts. A port preflight also refused a reuse
delay; no listener was present and only the owned runner's reuse check changed.

### Deferred at Bill's request

Bill is mobile and asked to defer the speech-pause repair and resume prior work.
He will manually unpause when needed. The required repair remains explicit
automatic-pause ownership with disconnect cleanup that never clears a manual
hold; keep the failing disconnect regression. No automatic-resume workaround,
live pause changes, restart, or reminder enablement is authorized by this deferral.
Next pending prior item: one size setting per memory band, with the default
high/low thresholds derived rather than independently configurable.

## One size per memory band — completed in source

Following Bill's instruction to resume prior work while mobile, I completed the
pending settings simplification, not the deferred speech repair. MemoryBandSettings
now declares budget and disabled only; the runtime derives graduation thresholds
at twice/once the effective budget. Zero preserves existing defaults. Budgets
above long.MaxValue/2 refuse before they can overflow the derived threshold.
The GUI shows six size inputs (including conversation) and enable controls,
without high/low fields. Chapters 21 and 22 describe the corrected contract.

Compatibility is read-only: complete legacy version-1 four-field bands are
validated, including zero-as-default watermark pair validation, then their old
watermarks are discarded. Partial legacy fields, bad types/ranges and unknown
fields refuse. Loading does not rewrite disk. An explicit later save emits only
budget and disabled, preserving sibling recall policy. No installed settings,
credentials, launchers or skills were touched.

A new serialization assertion failed against the old implementation before the
repair (before.log). Final settings persistence/refusal/codec groups pass 3/0;
shared-final.log reports 331 passed/0 failed and job82 exit0. Tests and GuiTests
builds report zero warnings/errors. The new memory-size-browser.cjs runs in the
existing isolated CrossoverFixture with local fake provider: six numeric fields,
no watermark inputs, acknowledged save of all six sizes/disabled flags, explicit
reload and browser refresh passed. The browser and fully joined host both exited
zero (results.json); no provider inference was needed. That browser run preceded
the final legacy-zero validation tightening; it tested the unchanged new GUI and
new-shape save path. The final C# suite covers the later codec change. Evidence
is /private/tmp/cadence-memory-one-size/. The deferred speech test remains failing
by design; it is not part of the shared suite's passing total.

No deployment, restart, commit or reminder enablement. Continue Chapter 26's
native screenshot replay under normal defaults, then voice session/UI work.

## Native screenshot replay after the retention correction

On the resumed Chapter 26 goal I checked children (none), fixture processes
(none), ports (free), current source, progress and saved results before running.
The fixture has no SetContextByteBudget override. Rebuilt current GuiTests
with existing SDK and --no-restore: zero warnings/errors. A fresh owned host
/private/tmp/cadence-capture-replay.ww8a6bl1 used synthetic credentials and only
loopback fake Responses on 18203, isolated GUI on 18202/assistant. Native Chrome
captured only its synthetic tab after a trusted Share click; automated this-tab
picker acceptance belongs to the fixture, not human picker acceptance.

Result: 960x640 PNG, 74,146 bytes, synthetic red/blue markers checked in actual
provider-bound media. Exactly four requests, one capture, matching call/result,
canonical History bytes and identical replay bytes on the later human turn.
Tracks ended, no attachment draft mutation, no browser errors. Browser exit0,
host exit0, STOPPED receipt after host/provider/Agent teardown. Outer job94 exit0.
Evidence: /private/tmp/cadence-capture-replay/{build.log,browser.log,host.log,
result.json,fixture.txt}. Prior failed replay is retained as historical evidence
of the removed budget policy; do not reintroduce a fixture-only enlarged budget.

No live provider, paid inference, private page/desktop/microphone capture,
installation, restart, skills/settings change or pause repair. Screenshot
human/tunnel checks remain pending. Next work is voice browser-session/UI and
its exact routing/cancellation/commit semantics, then remaining acceptance,
executable exercises/graders and the full accurate chapter. Reminders remain
at the authorized 180 seconds; no reconfiguration needed for this checkpoint.

## Voice browser-session coordinator — source-only component slice

Added `agent/gui/wwwroot/voice-session.js`: one gesture owns one hold and borrows the original socket/Agent/sequence. Begin, release/audio, ready and automatic commit are correlated; duplicate/stale results cannot redirect or repeat a send. Cancellation invalidates synchronously before joining hold cleanup. Disconnect or cancellation after commit is uncertain, never a retry or a claim of retraction. No transcript draft, additional Send, alternate funding, or pause manipulation is introduced.

`agent/GuiTests/voice-session.cjs` passes 12 Chrome groups covering exact audio/envelope, empty-root address, delayed begin, wrong connection/Agent/sequence/error, stale selection, late finish, disconnect, uncertain admission, failure/limits, real VoiceHold early permission release and late-track cleanup, joined teardown and cleanup failure. The first run exposed a real reentrant state-notification cancellation race: the following commit could still send. The corrected implementation rechecks current binding and terminal state immediately before effects and publishes cleanup before notifying observers; that negative run is retained. Existing VoiceHold nine groups and rebuilt five public GUI groups (including root/mounted asset delivery) pass. Build zero warnings/errors; all commands exited zero after correction and fixtures joined. Evidence: `/private/tmp/cadence-voice-session/{browser-first.log,browser-final.log,hold.log,build.log,gui.log}`. Existing shared core 331/0 evidence is carried, not newly rerun for this JS-only slice.

Still UNWIRED: composer microphone/shortcut, app message dispatch/lifecycle, full-page real lease protocol, native microphone and paid transcription acceptance. Controlled sockets/holds are not server or human acceptance. Next integrate through app lifecycle with visible hold/cancel/error/uncertain status and original binding. Neither live instance, installed skill, funding configuration nor pause state changed; no installation, restart, paid call or microphone access.

## Composer controls below the input — Bill’s voice-layout direction

Bill requested a separate control row beneath the chat input as the microphone is added. Source markup now labels the action group “Message controls”; CSS gives previews, textarea and wrapping actions their own full-width rows, attachment left and Send right. No nonfunctional microphone button was added. `agent/GuiTests/composer-layout.cjs` uses actual markup/styles (without app startup) for 12 geometry combinations: 1440/390/320px widths, 16/28px text, and with/without a fixture-only 44px third control. All pass touch-size, full-width input, below-input placement, containment and nonoverlap checks. This is geometry evidence, not native phone acceptance.

Rebuilt full-page image-browser regression passed picker/paste, caption/media admission, preview/remove, exact-once send, stale selection, refusal, reconnect without resend and cleanup. Five public GUI groups also pass; build clean, whitespace check clean, all fixture resources joined and host exit zero. Evidence `/private/tmp/cadence-voice-ui/{layout.log,build.log,browser-browser.log,browser-host.log,gui.log,results.json}`. Initial geometry run failed before Chrome because I guessed an absent Playwright path; corrected to the already-installed dependency, with no installation. No production behavior failure found in this slice. Existing unrelated changes preserved.

Source only: neither live instance installed/restarted. VoiceSession remains unwired; next integrate microphone hold and protocol/lifecycle routing into this row, preserving original Agent/socket and uncertain-commit refusal. Pause repair remains deferred.

## Voice controls and page wiring — source-only

Bill requested continued integration after the separate composer-controls row.
Added VoiceControl, which owns the held gesture and session cleanup, borrows the
current socket/Agent binding, and starts only on trusted pointer or button-key
(Space/Enter) events. The microphone, cancellation control and non-color status
are below the input. Global configurable Right Alt remains unfinished. Funding
must be explicitly advertised as enabled/openai-api-key; nothing enables a payer.
Existing text/image drafts block voice; voice blocks draft Send/attachment/camera.
Selection, disconnect, closed Agent, blur, hidden page and pointer/key loss
invalidate capture. Matching messages go to the original session. An uncertain
commit blocks reuse until explicit dismissal, never retry. No pause ownership
logic changed; its repair remains deferred.

Evidence /private/tmp/cadence-voice-control/: eight Chrome control groups and
12 geometry combinations pass; rebuilt public GUI five groups and full-page
image regression pass. Clean build, zero warnings/errors. New voice-page.cjs
passes five full-page groups: native synthetic-stream PCM through the real
recorder/hold/session/control, one addressed automatic commit, draft preservation,
blur cancellation, uncertain commit/no retry, and real funding-off advertisement
refusal. Voice replies and funding in the positive browser cases are explicitly
simulated; this is NOT server transcription/actor-admission end-to-end evidence.
No real microphone access or paid provider was used. Test oscillator contexts,
tracks, browsers and isolated hosts joined; final browser and host exit zero.
Existing core 331/0 evidence is carried, not rerun for this browser-only slice.

Corrections: asset-list edit initially inserted a string into a tuple, causing a
compiler error; fixed before the passing build. The first full-page attempt used
a nonexistent speech-setting selector and failed before recording; corrected to
actual #narrate, retaining the failed log. No product defect is inferred from
those setup failures. Earlier wrong inspection paths and dependency path were
corrected without installations. No live deployment/restart, payer configuration,
credentials/settings changes, or pause repair. Next: actual browser-to-server
lease/transcription/actor admission with a bounded fake-funded host; configurable
global hold shortcut; native phone/provider acceptance when coordinated with Bill.

## Real browser-to-server voice admission — verified, synthetic providers

The new `VoiceLeaseTests --browser-host` and `GuiTests/voice-server-page.cjs` exercise real page controls, native browser PCM recording from a synthetic audio stream, real WebSocket begin/audio/ready/commit, the lease, actual VoiceTranscriber multipart request/parser, actor admission, a model request, and canonical History. Only external provider HTTP boundaries are fake. The fixture enforces two transcription requests and one model request, uses distinct synthetic payer keys, validates canonical browser WAV and multipart fields, and does not save audio payloads.

Four browser groups pass: once-only full-path admission and model/History delivery; duplicate commit refused without new provider work; blur before release yields no upload; cancellation while an uncooperative transcription response is held makes late success inert. Tracks end. The final History count is taken after GUI/socket-lease work joins and before Agent destruction; immutable counters are checked after root teardown. Final browser/host exits are zero: two audio requests/completions, one model request, one admitted transcript, no provider failure. Ten existing lease/wire regression groups also pass, clean build, diff whitespace clean.

Evidence: `/private/tmp/cadence-voice-end-to-end/` (final build, regression log, browser/host logs, fixture path and results JSON). Initial fixture wrongly required an explicit `stream:false` model field; omission is valid non-streaming and that assertion was corrected. A subsequent stronger final check incorrectly read destroyed History; corrected by snapshotting after serving joins but before destruction. Both failures and the intervening successful run remain in subdirectories. These were fixture defects, not failures of production voice behavior.

No live microphone, paid provider, installation, restart, credential access or pause repair. This closes the earlier simulated-WebSocket-reply gap but does not establish real-provider/native-phone acceptance. Bill is available to test the microphone remotely on Dev once ready. Next: configurable global hold shortcut, followed by verified live transcription contract/funding and scoped Dev staging/testing; main remains untouched.

## Configurable voice hold shortcut — source verified

Added a GUI `voice_shortcut` preference with a strict three-value choice: `AltRight`
(default), `F8`, or `off`. It uses the existing save-before-publication path; legacy
settings without the field retain the default. Invalid values refuse the whole patch,
and failed saves preserve the old shortcut. Speech & Display exposes the choice.
This is a shortcut in the focused Ensemble page, not an OS-wide background listener.

VoiceControl uses trusted key events, ignores repeats, refuses AltGr/composition and
modifier chords, and declines Right Alt on Apple platforms rather than hijacking
Option. A different key, focus loss, composition, setting change, stale binding,
or hidden tab cancels without sending. Open dialogs block capture. Button-local
Space/Enter and pointer/touch gestures remain separate. No pause ownership changes.

Evidence `/private/tmp/cadence-voice-shortcut/`: 18 Chrome shortcut groups, eight
existing control groups, twelve session groups, five rebuilt public GUI groups,
and five real fake-funded end-to-end page groups pass. Builds have zero warnings or
errors, both final jobs exit zero, and browser/host/provider/Agent cleanup is joined.
The end-to-end test saves F8 through the actual form, verifies dialog refusal and
reload, then records real synthetic PCM with repeated F8 keydown and releases to
exactly one transcript/model request/History entry. Duplicate commit, blur and held
cancelled transcription still refuse; exactly two fake audio requests complete.
Public GUI checks include persistence across host restart and failed-save rollback.

AltGr/platform flags in component tests are controlled simulations, not native
Windows/Linux keyboard acceptance. An initial asset-check edit accidentally made a
three-item tuple; compiler refusal was corrected before test success. No live mic,
paid provider, installation, restart, settings modification in either live instance,
or pause repair. Next: verify live transcription model/API/pricing and explicit payer,
then announced Dev-only staging/restart and Bill's remote microphone check.

### Dev voice session staged and ready for human microphone acceptance

Bill identified the previously authorized credential source as `/Users/bill/.en/settings.json`. A bounded presence-only check confirmed a nonempty OpenAI field; no key value was printed, copied into a fixture, or changed. Primary OpenAI guide/model pages were checked for the fixed transcription contract and published estimated $0.0045/min rate; this is not an actual bill or entitlement proof. No paid/audio request was made during preparation.

Preflight matched all 15 Dev runtime hashes and its launcher against the prior receipt. GUI build passed with zero warnings/errors. The exact old Dev process fully exited after graceful signal before replacement. Six runtime files changed; all 15 installed hashes verified. Launcher, installed skills and settings fingerprints unchanged. A private supplementary launch script in `/private/tmp/cadence-voice-dev-install/` adds explicit `--transcription openai-api-key --transcription-key-file /Users/bill/.en/settings.json` while retaining Codex chat funding and the authenticated public origin. The saved Dev launcher is unchanged and will launch with transcription off if used later.

New Dev listener, local HTTP 200, exact served voice-control/session/app assets, and a metadata-only WebSocket advertisement (enabled, openai-api-key, gpt-transcribe) verified. Connection closed without prompt/audio. Main8088 listener unchanged throughout. Evidence: `/private/tmp/cadence-voice-dev-install/{preflight,stopped,staging,launch,readiness,funding}.json` and build.log. No native microphone/provider acceptance yet. Next: Bill refreshes Dev and performs one short hold/release phrase; no extra Send, automatic retries, or background listening. Reminders remain off pending that human test.

## Host screenshot scope correction — Bill's clarified requirement

Bill clarified that screenshots mean the computer running the C# server, like
CodeRhapsody, not Chrome-tab capture. I should have clarified that distinction
before choosing the browser route. Inspected reference:
`/Users/bill/projects/coderhapsody/internal/tools/screenshot_ops.go`, tool schema
and isolation classification. It lists displays and captures a display or screen
rectangle, encodes PNG, and delivers image content directly. Preserve the behavior,
not advice to elevate privileges. No desktop capture or permission change occurred
during this inspection. Chapter 26 opening and active goal are corrected; older
browser evidence remains valid only for that separate feature. Next implement
and test host capture with existing media preparation/placement guards, bounded
resources, explicit authority, cancellation and joined cleanup. No new dependency
installation or live-app replacement is implied. Native tests must be announced
and limited to an agreed synthetic region, not incidental private desktop content.

Bill also accepted the live Dev voice test: two supplied test phrases reached
canonical Human messages and completed model replies; current Dev state was idle
and unpaused. His follow-up attributed the perceived absence to waiting for the
reply. This is live voice acceptance, not evidence of a display defect.

## Host screenshot backend — source-only, native acceptance pending

Bill clarified that screenshots mean CodeRhapsody-style server capture, not a
browser picker. The chapter opening and goal now carry that requirement. I read
CodeRhapsody's tools and its installed kbinani Darwin backend: the newer path uses
ScreenCaptureKit, so simply copying an older CGWindowList capture call is not a
sound macOS 26 plan. The first C# backend uses CoreGraphics display metadata and
permission preflight plus the fixed built-in `/usr/sbin/screencapture` executable.
No downloaded dependency, shell command chosen by a model, permission request,
elevation, browser picker, or native screen acquisition was used in this slice.

`Common/HostScreens.cs` declares immutable display/rectangle/selection values and
partial backend operations; `Agent/HostScreens.cs` implements them. Source bounds
use conservative backing-pixel scaling and the existing image pixel ceiling.
Negative display origins are retained. Rectangles must currently fit exactly one
display; cross-display regions explicitly refuse, a parity limitation rather than
an accepted claim of full CodeRhapsody behavior. Native coordinate/scaling,
display enumeration, and TCC behavior remain untested.

Capture uses private owned scratch, a fixed argument list, an empty child
environment, a 15-second deadline, bounded PNG reads, and no persistent raw image.
Cancellation kills and joins the owned process/tree and both output readers;
scratch removal follows joining. Review tightened cleanup so kill errors cannot
skip joining or hide the original wait failure. Prepared image delivery will use
the existing Agent pipeline; no tool names or skills are advertised yet, and
backend permission preflight is not the missing model authority gate.

Two new focused groups use synthetic display inventories and owned fake
executables only. They cover exact rectangle arguments, scaling/overflow/source
bounds, inherited-environment stripping, real synthetic PNG preparation, process
and malformed/oversize-output failure, cancellation/deadline, pre-cancel refusal,
and scratch cleanup. Both pass. Rebuild: zero warnings/errors. Shared suite:
333 passed, zero failed, exit zero. Evidence:
`/private/tmp/cadence-host-screens/{build-final.log,focused-final.log,shared.log}`.
All owned fixture processes joined. No native screenshot, live install/restart,
paid request, microphone use, worker, or push. Main and Dev remain untouched.

Next: wire `list_displays`/`screenshot` with explicit host-screen grants, safe-mode
and capability refusal before effects, and real Job/image-result retention tests.
Then announce and agree a synthetic region before testing the native backend.
Only after implementation/acceptance come exercises, external graders, and full
chapter prose. Pause repair stays deferred.

## Host-screen tool wiring — source only; native acceptance waits for Bill

Added `list_displays` and `screenshot` (CodeRhapsody-compatible `display`, or all four rectangle fields) to the built-in registry and Job dispatch. Screenshot results retain actual prepared PNG parts through the existing structured-result path. `host-screen` is an optional loadable skill offered by the source primary skills; neither installing the registry nor ordinary code/browser access grants capture authority. Both safe mode and legacy unskilled contexts refuse host tools, including declaration rendering. Capability validation precedes enumeration/capture; authority and cancellation are checked again before returning prepared pixels. Bounds and native output scale must match; unsupported models, OS/platform failures and deadlines remain explicit. No new channel or browser picker.

Four focused groups pass: selection/source limits, strict argument parsing, grant/unload/safe-mode and synthetic pixel rendering, and fake-process cancellation/cleanup. No native enumeration, OS prompt or actual screen capture ran. Native successful acquisition and delayed-Job image acceptance remain pending; synthetic preparation and renderer tests do not certify those paths. Current backend is macOS-only and refuses cross-display rectangles. It preflights permission and refuses rather than requesting OS permission automatically.

Evidence: `/private/tmp/cadence-host-screen-tools/`. Retained corrections: non-record Configuration fixture construction; incorrect positional test filter (subset 118/1, not a full suite); old legacy declaration count excluded new explicitly gated tools; missing `type: loadable` in the new skill and `type: primary` in its fixture. Final focused build is clean and four groups pass. Full shared suite passed **335/0**, exit zero; job 270 is joined (`shared-final.log`). No installation, restart, skill changes in live homes, paid requests, or private capture. Bill is away; native acceptance requires his participation and an agreed synthetic region.

## Chapter 26 exercises and first external scope — image-file delivery

Bill authorized drafting exercises and external graders for already-verified paths
while native host capture awaits his return. Read the current procedure and the
complete book/voice.md; the older worker author-role workflow does not override
personal work. Chapter 26 now has four exercise contracts with a proposed
20/25/30/25 allocation. Full narrative prose remains unfinished. The book index
now identifies the exercise draft and limited grade rather than no implementation.

Added Edition 3 `-ch 26 -executable-only -media-scope image-files`, without changing
preserved Gateway graders. Strict flag/scope/executable checks precede candidate
launch. The external Go grader drives the candidate's public JSON CLI against a
loopback fake Anthropic provider. It creates a private random four-band 4096x64 PNG,
checks one matching 2048x32 PNG tool result, checks all band-interior pixels within
resampling tolerance, preserves the original byte-for-byte, deletes it, then checks
identical retained pixels in the next human turn. Three provider requests, bounded
diagnostics/deadline, isolated HOME/workspace/skill and synthetic key, and joined
CLI teardown. This deliberately narrow scope does not certify all image safety,
provider, delayed-Job, UI, voice, or host-screen requirements. Title and chapter
explicitly say 20 points available, 80 ungraded; it is not a complete chapter grade.

Parser negative controls reject missing/metadata-only/corrupt/mispaired/duplicate
images and wrong pixels, including a single corrupted interior pixel. External
negative candidate programs exercise the public CLI/provider protocol but return
metadata only or lose replay; both are rejected. The wrong-color mutation initially
used XOR 127, which can change a channel by only one and fall inside tolerance;
changed it to XOR 128 and repeated the negative test twenty times successfully.

Fixture corrections retained: guessed Go helper APIs and module prefix corrected;
CLI prompt needs kind, not type (the first replacement missed map[string]any);
early exit detection and bounded tool-error diagnostics added. The first real CLI
run exposed stale --no-restore CLI assets lacking SkiaSharp, not a project/source
packaging defect. Regenerated assets in a separate grader directory using already
cached packages, an empty package feed and NuGetAudit=false, then rebuilt cleanly.
No dependency download or runtime source change was needed.

Final external scope: 20/20, exit zero, against that rebuilt C# CLI. Current Go
command/grader/fake-provider package tests pass with GOTOOLCHAIN=local, GOPROXY=off,
GOSUMDB=off and the existing module cache. Three companion publication tests and
four non-PDF book tests pass; broader book discovery has one import error because
the default Python lacks PyMuPDF. No installation attempted. Assembly to a private
file and git diff --check pass. No published artifact replaced, no live requests,
native capture, permission prompt, installation/restart, commit or push. Existing
core 335/0 evidence is carried, not rerun for these grader/document changes.

Evidence: /private/tmp/cadence-media-grader/ (image-grade-tightened.json,
go-current-regression.log, pixel-negative-repeat.log, cli-restore-offline.log,
cli-build-final.log, book-checks.log, book-root-checks.log, assembled-book.md).
Next: implement the addressed attachment and voice external scopes with meaningful
negative candidates. Native screenshot success/model-turn/delayed-Job acceptance
still waits for Bill and an agreed synthetic region. Pause repair remains deferred.

## Full Chapter 26 narrative draft at Bill's request

Bill explicitly changed the earlier sequence and asked for full prose now,
while native host acceptance and the remaining external grades are pending.
`book/chapter-26.md` now contains a full narrative draft: fourteen sections on
ownership, bounded preparation, provider rendering, attachments, host capture,
PCM recording, explicit transcription funding, ready/commit admission,
uncertainty/cancellation, keyboard controls and evidence. Four existing exercise
contracts are preserved. The screenshot-surface correction and live Dev voice
test are attributed to the supported Bill/Cadence record, in third person as
required by `book/voice.md`. The index and current authoring procedure reflect
this chapter-specific sequence change; historical procedure records remain.

Read-back corrected an invented `PreparedImage.ToBlobPart()` call to the actual
public `.Image` property, removed obsolete future-narrative wording, and kept
pending native acquisition/model-turn/delayed-Job acceptance, the 20/20 scoped
image grade versus 80 ungraded points, and the deferred speech-pause bug explicit.
A few checked text anchors refused before writes; subsequent edits used the
actual source wording. No implementation test or grader result was advanced on
the strength of prose edits.

Evidence: `/private/tmp/cadence-media-prose/`. Prose lint exits zero: hard checks
pass, with advisory warnings for negative framing and a long impersonal
exercise/grading span. These are recorded, not represented as an all-green style
review. Three companion publication and four print-sidebar tests pass; assembly,
local-link and 00–27 ordering checks pass; whitespace check passes. The chapter
is about 6,700 words including specification and exercises. Full PDF-test coverage
remains unavailable in the default Python because PyMuPDF is absent, as recorded
in the previous grader slice; no dependency installation or published artifact
replacement occurred. No live host, setting, credential or reminder configuration
changed. Next: public-interface attachment and voice graders; native host testing
still requires Bill at home, OS permission and an agreed synthetic region.

### PDF-test dependency resolved with Bill's permission

Bill explicitly authorized installing PyMuPDF. Created an isolated environment
at `/private/tmp/cadence-media-prose/pdf-venv`, leaving system Python unchanged.
Installed the published binary PyMuPDF 2.0 package and its declared dependencies
from PyPI. Exact resolved versions are retained in
`/private/tmp/cadence-media-prose/pdf-requirements-resolved.txt`; installation log
is `pdf-install.log` in that directory. All eight print/PDF regression tests
passed, including native PDF stamping/link/color checks with no skips. The full
book test discovery and companion publication suite were then rerun in the same
environment; logs are `book-tests-with-pdf.log` and
`companion-with-pdf-env.log`. The earlier missing-module limitation is resolved
for this isolated environment, not retroactively erased from earlier evidence.
These checks use fixture PDFs. No full-book PDF/EPUB was regenerated or published,
no live runtime or credential changed, and no reminder setting changed.

## Native screenshot permission preflight; capture deferred when Bill left

Bill approved opening a synthetic red/blue window and capturing only its content
rectangle. The opt-in .NET test probe returned screen-recording permission
`true` and one 1920×1080 logical/physical display. This establishes permission
for that probe process, not a live installed tool grant or successful capture.
No permission request, screen capture, or private desktop inspection occurred.

An owned AppKit synthetic-window fixture compiled after explicitly selecting the
installed macOS 14 deployment target; the default Swift target could not load its
macOS 26 standard library. No compiler/package installation was performed. An
opt-in capture harness was also built, restricted to the fixture's 640×360
rectangle, with dimension and red/blue pixel checks before saving its PNG.
Neither the window nor that capture mode was run. Bill then left and requested
continued work; native capture remains deferred while he is away. Continue the
remaining external grader scopes without capturing or prompting remotely.

Evidence: `/private/tmp/cadence-host-native/permission.json`, `build.log`,
`build-capture.log`, `window-build.log`, and `window-build-second.log`.
No main/Dev runtime, settings, launcher, or installed skill changed. Reminders
remain off; this note does not reconfigure them.

## Attachment grader payload checks prepared while native capture waits

Extracted the existing prepared-PNG pixel validator for reuse without changing
its dimensions, byte checks, or randomized interior-pixel comparisons. Added an
attachment-specific provider-request check that requires the supplied caption and
one actual image in the same human message. Its positive and fourteen negative
checks cover text/metadata-only output, wrong roles/captions, split messages,
duplicate messages/captions/images, missing or malformed bytes, wrong MIME, bad
JSON, and corrupted pixels. The code comments explicitly distinguish this payload
oracle from the pending candidate admission/acknowledgement/lifecycle grade.

`go test ./internal/gradeedition3 ./cmd/grade-edition3` passed with dependency
networking disabled. The rebuilt external image-file grade also remains 20/20,
exit zero with its CLI joined. Evidence is in
`/private/tmp/cadence-media-grader/attachment-oracle-tests.log` and
`image-grade-after-attachment-oracle.json`. No additional chapter points were
awarded: eighty remain ungraded. Next is the public GUI candidate driver and
addressed admission/stale/duplicate-send checks, then voice grading.

No native window, screen capture, microphone, paid provider, installation, restart,
or reminder change occurred. The permission-only native preflight and the compiled
opt-in window/capture fixtures remain ready for a later agreed human test.

## Bill's queued GUI follow-up — after Measure X delegation

Bill requests these issues be retained in the queue while Chapter 26 work resumes:

1. Settings gear has wrapped onto the next header line. Inspect actual responsive geometry; consider giving that header region more width, but preserve phone layout and accessible controls rather than assuming width alone is the fix.
2. Model selector no longer opens, although it previously did. First reproduce and trace DOM/event/disabled-state behavior. Check availability of the three vendor API credentials only through scoped, non-disclosing configuration checks; never print or copy values. Bill expects all model choices to remain visible, with only available vendors enabled (even if only OpenAI is configured). Credential presence is distinct from model entitlement and from safe cross-provider switching: running Engine provider/funding ownership must be respected, with no silent API-key fallback from Codex funding. Determine the actual failure before choosing a repair.

These are reported symptoms and desired behavior, not independently verified findings. No credential inspection or live GUI action has yet been performed for this queue entry. No install/restart or paid provider request is implied.

Measure X reconstruction is now delegated to child MeasureX, which confirmed its own 180-second reminders enabled and the 54-reviewed-draft boundary. Main collaboration returns to Chapter 26 host screenshot acceptance and remaining graders, then scheduler/Gateway/multi-user design. Native screenshot testing still needs Bill's current agreement to a synthetic window/rectangle; old preflight permission is not a new capture result.

## Native host capture — HiDPI correction and first successful pixels

Bill explicitly authorized the synthetic-window capture. The original native
probe failed its fixed 640x360 expectation; its window was joined and no PNG saved.
A single documented-syntax diagnostic also returned 1280x720 and was deleted
without pixel inspection. Argument spacing was not established as a defect.
Metadata-only comparison established the cause: CGDisplayPixelsWide/High reported
1920x1080, but CGDisplayModeGetPixelWidth/Height reported 3840x2160 for the same
1920x1080 logical mode. Source now copies the current display mode, reads backing
pixel dimensions, and releases it in finally. This corrects pre-capture source
pixel accounting as well as fixture expectations. Native probe samples relative
to backing dimensions instead of assuming 1:1 scaling. Its scale-ceiling-based
exact-size assertion is a synthetic fixture expectation, not universal fractional
scale acceptance.

Clean test build, four focused host-screen groups passed (exit0), whitespace clean.
One corrected real HostScreens capture of the synthetic640x360 logical rectangle
returned1280x720 PNG,25108 bytes. Probe verified red/blue sample pixels; Cadence
inspected actual saved pixels and saw red left, blue right. Window joined exit0.
Evidence /private/tmp/cadence-host-native/scaling-acceptance-7zmyjuja/;
build-scaling.log and scaling-focused.log in the parent directory. No private
screen sweep, live service change, paid request, installation or restart.

This proves native backend capture and actual-file pixel delivery, NOT native
model-requested screenshot tool/delayed-Job acceptance. Those paths, remaining
external grades, broader final regression and prose reconciliation remain next.

## Native screenshot model-turn, delayed Job, and replay acceptance

An opt-in fixture in ActorTests.HostScreenNative.cs now drives a real Agent against
a loopback fake Responses provider with an explicit host-screen skill. It is NOT
registered in the ordinary suite. It accepts only the approved640x360 synthetic
window bounds file. Under Bill's existing controlled-test authorization, one native
screenshot call with zero-delay observation returned a running Job; wait_for_job
collected its structured image. Four local requests verified one1280x720 prepared
PNG (6824 bytes), red/blue pixels, actual inline image bytes in the paired tool-output
wire, and identical canonical bytes replayed on the next human turn with no second
capture. Cadence inspected the delivered PNG. Agent/provider and window cleanup
joined; native fixture exited0. No paid request, private screen sweep, live install
or restart. The fixture's exact1280x720 assertion describes this agreed HiDPI test,
not a generic fractional-scale acceptance claim.

Evidence /private/tmp/cadence-host-native/model-turn-nhpqw5f8/ (turn.log,
fixture/result.json and delivered.png). Build-model-turn.log clean. Shared suite
reported337 passed/0 failed and exited successfully; its combined shell job112
then exited2 solely because git diff --check found a trailing blank notebook line.
That whitespace was corrected. Chapter26 acceptance statements now reflect this
native scope and explain the HiDPI lesson; external scope remains20/100 with80
ungraded. Next: remaining attachment/voice/host external graders, review/prose
reconciliation, and separately queued GUI issues. No more native capture needed
merely to repeat this successful evidence.

## Queued debugging improvement — per-Agent logging

Bill reports that all Agents appear to write into Cadence's log rather than logs in their own Agent directories, complicating diagnosis. Investigate actual logger ownership and routing, then provide separate logger instances/destinations per Agent, including appropriate child paths. This is a queued request only, not an independently verified all-Agent finding or an implemented fix. Preserve diagnostic masking and distinguish shared host events from Agent-owned events. No logging changes are part of the screenshot staging.

## Queued lifecycle improvement — unfinished children across host restart

Bill requests investigation and design of restoring unfinished delegated Agents after a host restart. Current observed incident: after Bill restarted the staged host, agents_status returned no live children; Cadence/children/MeasureX and its transcription files/goal survived. Saved worker checkpoint reached page84 (83 full-page reviewed drafts plus incomplete page59; page85 next), but no live membership or timer remained. The GUI absence reflected missing live ownership, not loss of saved page work.

Design explicit restoration of parent/child membership, GUI visibility, unfinished-work status, and goal/reminder lifecycle. Distinguish documentary saved state from current authority; do not silently reactivate stale grants, submitted work, or arbitrary scheduled effects. Test restart recovery and failed restoration without duplicate live children or duplicate work. Bill explicitly authorizes restoring MeasureX now and re-enabling its180-second reminders; this is not authorization to implement or deploy a general restart mechanism in this slice.

## Screenshot simplification — display-only public contract

Bill removed rectangle selection to avoid window-measurement distractions during
debugging. Public screenshot now accepts only optional nonnegative integer display;
omission selects0. Strict callers may explicitly send0. Null is invalid. No public
x/y/width/height selectors remain. Stale rectangle requests refuse before capture,
never silently expand into whole-display capture. list_displays remains available
when another display needs identifying. Internal rectangle geometry remains for
whole-display bounds and backend tests, not a second public selection API.

Updated schema/parser, loaded-skill model declaration regression, default/explicit/
secondary/refusal tests, host-screen skill, Chapter26/README and grader-side
screenshotSelectionSchema oracle with positive/negative tests. Existing validator
expects scalar type strings: an initial nullable-array draft exposed that mismatch;
removed the unnecessary nullable alternative instead of expanding shared validation.
An initial schema test rendered an empty context without skill grants; corrected
to the actual loaded History. Final focused4/0, shared337/0, clean build; both Go
grader packages passed offline using existing /Users/bill/go/pkg/mod. First offline
run used redirected home's missing cache and failed setup; no dependency downloads.
Chapter26 hard prose lint passed with style advisories. Evidence:
/private/tmp/cadence-display-only/ (build-third.log, focused-third.log, shared.log,
graders-cached.log, prose.log). No native capture, installation or restart here.

Native model fixture now explicitly opts into --host-screen-native-display-turn
with a whole synthetic display; the retired rectangle command is no longer wired.
It has compiled, not run. Earlier native rectangle-path evidence remains historical,
not acceptance of changed whole-display selection. Full host external driver is
still pending: the tested schema oracle alone awards no points. External total
remains20 image-file points with80 ungraded. Future display-only native acceptance
must use a currently agreed whole display. Next deployment requires both runtime
and host-screen skill wording, preserving other installed skills and configuration.

## Queued UI cleanup — remove persistent voice-disabled footer

Bill requests removal of the chat-pane footer line, “Voice is off: separate transcription funding is not enabled.” The disabled microphone icon already communicates unavailability; do not replace this line with another persistent funding notice. Keep useful actionable error feedback and accessible microphone labeling. This is queued only, not implemented.

Bill also authorizes scoped local Git commits when reviewed changes are ready. Preserve unrelated changes and MeasureX-owned reconstruction work; commit readiness is separate from staging, restart, and publication. No push is authorized.

## Queued settings ownership — browser-local appearance

Bill requests browser-local settings rather than sharing all preferences through the server across devices. Light/dark theme specifically belongs to browser state, not backend settings. Sharing between tabs in the same browser is acceptable. Use origin-scoped localStorage as the proposed persistence mechanism, with storage-event synchronization across tabs; handle unavailable storage gracefully. Review other preferences individually rather than moving all settings indiscriminately. Define migration/default behavior for the existing server theme without continuing to publish browser theme changes to the backend. Verify independent browser contexts/devices, same-origin tab synchronization, and reload persistence. This is queued only, not implemented.

## Queued speech-pause follow-up — intermittent unpause with multiple clients

Bill reports that automatic unpause with speech enabled still appears intermittently broken and suspects simultaneous connections from multiple devices. In this reported exchange, he initially expected Enter would leave Cadence paused, then immediately corrected the observation: “Nope you unpaused directly.” Thus this attempt succeeded; it is not a new failing reproduction or proof of multi-device causation. Investigate with isolated concurrent-browser/client tests, including Enter submission, speech-owned pause release, disconnect/reconnect and manual pause preservation. Keep this hypothesis distinct from the earlier reproduced speech-owned pause leak on browser disconnect. Release only the appropriate automatic pause ownership; do not introduce blind unpause that overrides manual holds. No live state manipulation or repair performed for this queue entry.

## Ordinary-child restart restoration — source verification

Bill approved automatic reconstruction of ordinary delegated children, their own enabled goals and reminder intervals, preserved deliberate pause/submission, and fresh reminder-based continuation rather than replay of interrupted effects. He authorized autonomous implementation, Chapter20 Managing the Swarm prose/grader updates and scoped commits while away. No runtime restart was authorized.

Implemented Common ChildRestartState/Agent declarations and Agent-owned child.json registration via existing Sandbox helpers. Ordinary CreateChildAsync registers; direct transient/compressor construction does not. RestoreChildrenAsync validates bounded records, reconstructs real ownership/current policy and own History recursively, and rolls back newly constructed children on failure. CLI, GUI composition root and restored peers invoke it; reusable GUI mount remains borrowed. GUI can wait for each child's own browser relay; headless defaults still refuse missing browser startup. Existing live children are not duplicated. Explicit shutdown/kill retires registration but retains artifacts; ordinary host shutdown retains registration.

Goal, pause, model, generation and submission state save before publication where applicable. Shutdown stops runtime timers without persisting goal-off. Interrupted calls/processes/Jobs are not replayed. Current goals authority and fresh file validation gate restored reminders. Missing authority/file disables with diagnostic. A new interval, not missed-timer catch-up, starts after restoration. Submitted children stay off; paused children stay paused. Submission artifact and registration are separate writes, not a multi-file transaction: generation and normalized payload SHA256 bind acknowledged completion; mismatch refuses rather than inventing acceptance. Historical get_submission stays independent. Standalone goals remain session-local and parent goals are not inherited.

Verification: two focused real-component groups cover nested ownership, History, explicit pause plus enabled180s goal, submitted state, old-result/new-work refusal, explicit removal, no helper restoration, repeated restore, malformed/mismatched/duplicate metadata, failed registration write preserving live goal settings, unavailable skill authority, and submission checksum mismatch. A real-Agent fake-Anthropic test observes a fresh reminder interval, retained goal/conversation, two HTTP requests and one durable submit disabling reminders. Final shared run339/0 (job287, exit0); final narrow exception-filter amendment rebuilt and both focused groups passed (job299). Tests, CLI and GUI builds clean. No paid requests or native screenshots.

The active historical-number Chapter21 C# executable adapter now tests actual CLI restart with the same isolated home: no second spawn, restored idle/submitted state, retained History on fresh work, exactly two requested exchanges. Existing70-point scope retained,70/70 passed; no extra GUI/persistent-peer points invented. Negative wrapper removed only its own fixture registrations and correctly failed the composition row with “No owned child has that ID.” Relevant TestChapter21 Go tests pass; Edition3 packages pass. An overly broad full gradeedition2 package run failed known archived-layout/old-contract tests; frozen archives remain unchanged. Chapters20/25/26 hard prose lint passes with style advisories; existing em-dash punctuation in touched20/25 was reconciled. Evidence /private/tmp/cadence-child-restart/.

Corrections: initial fixture installed tools twice after SkillSpec construction; fixed setup. Shared338/1 found the old peer test explicitly forbidding descendants; updated to require real restored ownership while preserving launch-choice checks. Missing-skills refusal on a paused child initially waited for a non-control load_skill; added current owner/authority precheck and demonstrated prompt refusal. Initial broad search included binary artifacts; subsequent reads used actual source paths. None of these observations is live restart acceptance.

Live transition pending: MeasureX predates registration. It needs explicit migration from current known live state immediately before restart, not indiscriminate discovery of historical child directories. At latest child report it reached126 (114 full-page drafts,12 incomplete,140 pending); its reminders were re-enabled180s after another explicitly partial framework-prompted submission. Latest tool status also reported an HTTP transport failure, not lost files. Current source/runtime bundle is not staged; host-screen and goals installed manuals need matching updates when staging. Preserve launcher/settings/credentials/data; Bill owns restart. All transcription and unrelated untracked notes remain outside the source commit.

## Live restart/screenshot acceptance and next UI work

Bill reported restart of staged73c42d77. agents_status showed MeasureX automatically restored, idle/unpaused/unsubmitted/no failure, without spawn. Later its new parent question1 reported a fresh automatic reminder and asked whether the staging hold was released. Parent explicitly released it to resume133–266/full audit. This verifies automatic live membership and a real reminder, not complete reconstruction.

Bill explicitly authorized one whole-primary-display screenshot. Live screenshot(display:0) succeeded directly, no native-command workaround or window measurement: bounds1920x1080 logical,3840x2160 pixels prepared2048x1152,493944 PNG bytes. Actual pixels inspected; temporary capture removed. GUI, child tree, wrapped gear and voice-disabled footer visible. This supplies live display-only acceptance; omitted display was covered by source tests, not a separate live invocation. No further capture authorization inferred.

Collaboration correction: Bill says narration is critical to how he learns and contributes to the dyad; if anything, more substantive explanation, not less. My visual-clutter criticism must not become a policy of reducing engineering narration. Preserve intent, reasons, discoveries, alternatives and corrections in chat. Improve visual presentation separately; do not hide Actions or narrated work.

Queued Actions-pane proposals, for discussion before significant redesign: emphasize action name, actual state and concise result summary; keep each call/result visibly paired; reduce repetitive wrapper labels and duplicated detail affordances; make raw arguments/results accessible on demand without losing full output; review speech controls and keyboard/screen-reader labels; preserve stable scrolling and the always-visible Actions pane. Header grouping and clearer visual hierarchy are proposals beyond the already-approved gear-wrap repair. Do not remove user-visible evidence or auto-collapse useful content without discussion.

Bill authorizes a short autonomous slice for simple settled tasks with no significant design/learning tradeoffs. Priority Chapter26 closure review first, then gear wrapping and removal of redundant voice-disabled footer, with bounded tests. Do not mislabel remaining80 ungraded media points completed. Browser-local theme remains queued; cross-vendor funding/model behavior, per-Agent logging and concurrent-client speech-pause need deliberate review/pairing. No restart, installation, paid calls, new private capture or push in this slice. Root goal to be enabled180s; worker goal separate.

## Settled UI fixes — source-only completion

Reproduced gear wrapping in a new failing geometry assertion: seven toolbar controls were placed in an explicitly six-column grid. Replaced fixed-count columns with horizontal implicit columns, retaining36px header targets and existing local horizontal scrolling. The regression includes an eighth synthetic control and32 desktop/phone,18/28px text,dark/light combinations. Lesson: derive the row from actual controls rather than hard-code their count.

Removed the visible idle funding-off footer using Bootstrap visually-hidden on the existing status element, preserving the microphone's aria-describedby connection and actual explanation. Active/terminal voice states and unsupported-browser feedback remain visible; no funding, transcription, speech-pause or ownership behavior changed. Nine control groups cover funding/untrusted gestures/pointer/keyboard/draft/selection/blur/uncertainty/failure. The full page confirms the unfunded description occupies no footer row while successful feedback remains visible; five synthetic page checks and five public GUI groups pass, with browser/host joined exit0. New regressions failed before production edits. GUI test-host build clean; Chapter26 hard prose lint passes with advisories. No main/Dev installation or restart, native microphone, provider call or extra display capture. Evidence /private/tmp/cadence-ui-polish/.

Disk reserve: free space approached1.18GiB. Removed only Cadence's disposable /private/tmp/cadence-child-restart/go-cache (289336KiB); retained source, package cache, test logs/receipts and all MeasureX work. No running owned Go job used that completed-test cache.

Chapter26 now records actual live display-only screenshot acceptance and separates it from earlier synthetic rectangle-path evidence. Remaining80 external points are still ungraded; this is not whole-chapter completion. Actions-pane work remains discussion-only, with substantive narration explicitly preserved.

## Remaining grader design boundary after the short autonomous UI slice

Source inspection confirms the image-attachment scope currently contains only attachmentPNG payload validation, not a browser/candidate lifecycle driver. Exercise26.2 additionally requires real preview/draft, addressed WebSocket admission and negative controls. The voice scope requires real WAV/transcription/lease/admission; Common/VoiceTranscription.cs exposes public(parent,apiKey), while injectable HTTP handler is internal and GUI startup uses the public fixed-route constructor. A complete external fake-provider voice grade therefore needs a deliberate candidate/test-host contract, not a silently added production endpoint override or a paid test mislabeled synthetic. Host25-point external scope likewise needs controlled native-display setup/permission; schema-only checks and the prior one-shot live screenshot do not award it. Discuss these testability/driver choices before a larger autonomous grader implementation. They are open design work, not proof the feature or grading is infeasible.

Short authorized straightforward slice completed in local commit294247d2: gear/idle footer fixes, tests and accurate live acceptance prose. No runtime installation or restart. Remaining queued model/funding, per-Agent logging, concurrent-client pause, browser-theme migration/default semantics and Actions-pane redesign need pairing or explicit choices. Parent reminders are being disabled at this decision boundary; worker reminders remain independent. Original request for more substantive narration is preserved. Do not repeat completed UI checks merely for freshness or call Chapter26 closed with80 points ungraded.

## Queued appearance change — swap human and Agent message backgrounds

Bill confirms the rebuilt gear-wrap and idle voice-footer fixes work in his live GUI. He requests one additional queued change: swap the existing human-message and Agent-message background colors, rather than choose a new palette. Agent output occupies more of the conversation, and he prefers the dominant screen area to use the background currently behind his own messages. Bill identifies himself as a protanope; refer to message roles/current appearances rather than requiring color names. Preserve explicit speaker labels and other non-color distinctions; check text/link/code contrast and both themes when implementing. This is queued only, not yet implemented or installed.

## Approved queue continuation — message-background swap

Bill approved background swap → browser-local theme → model-dropdown diagnosis →
per-Agent logging → speech/unpause repair. Substantive narration remains essential.
Swapped the exact existing human/Agent fills in both themes, not panel surfaces or
speaker names. Browser regression failed before the change, then caught inherited
Bootstrap link contrast on the newly colored human bubble. Applied the existing
accessible message-link colors to both roles. Four role/theme checks now pass
(exact fills, body/link contrast >=4.5, visible identity); all32 layout cases pass.
Tests use real CSS in isolated headless Chrome, no live capture/provider. Source
only; no installation/restart. Lesson: a background swap also requires auditing
role-specific foreground styles. Bill already confirmed previous gear/footer fixes
working after his own rebuild/restart. Remaining media80 points still ungraded.


## Browser-local theme ready for coordinated rebuild

Theme now uses origin-local storage (ensemble.theme.v1), shared by same-origin
tabs but independent across browser profiles/devices. A browser with no valid
local choice seeds once from the legacy server theme. Later settings frames do
not override it; header/dialog controls send no theme patches. System follows
OS changes. Blocked storage retains a page-session choice; offline theme controls
remain usable. Storage events from sessionStorage are ignored. Other settings
retain server save/acknowledgement semantics. UI labels theme as this-browser.

Verified isolated component tests plus the real embedded crossover page: migration,
header/dialog controls, acknowledged settings refresh, reload, OS changes, tab sync,
profile isolation, storage-area filtering, disconnect/reconnect and no backend
theme writes. Existing settings-dialog shared-save/reload, keyboard, six-tab layout
and enlarged-text checks pass. Four role-color/contrast cases and32 composer layout
cases pass. Isolated fixture/provider/Agents joined, exit0; no live user browser,
provider prompt, microphone, capture, installation or restart. Evidence tool jobs
301/313/315/316 and /private/tmp/cadence-theme-integration/. Initial Chapter8 lint
found two existing em dashes; corrected without changing historical endpoints.
Chapters8 and27 now explicitly teach local theme and legacy field compatibility.
Final build/lint/public-GUI result is collected separately before commit.

Final theme verification job321: GUI build clean, five public GUI groups pass,
Chapters8/27 hard lint pass with style advisories, whitespace clean. No installation
or restart. Normal working-tree rebuild no longer includes an unfinished theme.

## Live rebuild/restart acceptance, 2026-11-02

Bill rebuilt/restarted and confirmed “The colors look right.” Six installed
application assemblies/symbols match the promotion build. Live loopback8088
PID97113 serves byte-equal theme.js/app.js/index.html (HTTP200). Separate host,
Cadence and MeasureX diagnostic files are open in that process. Cadence API/gui
logs measured2,574,159/1,816,574bytes, replacing the earlier multi-GB accumulation;
free space15,802,468KiB (~15.1GiB). Metadata only, no private log content inspected.
MeasureX automatically restored idle/unpaused/unsubmitted/no failure; no spawn,
interrupt or timer mutation. Fresh post-restart worker reminder not yet observed.
Human color acceptance is established; manual cross-device theme independence is
not claimed (isolated browser tests cover it). No restart/install performed by me.

## Model picker: idle pause is not active work

Live DOM inspection found the model picker disabled while Cadence was paused;
all seven curated choices were present. Browser renderControls disabled it for
both busy and paused states, whereas Agent.SelectModel permits an idle pause
and independently rejects active/queued work and overlapping Engine sends.
Removed only the blanket paused browser gate. Disconnected, selecting, closed,
unacknowledged and busy states remain disabled; provider/API/funding restrictions
and server acknowledgement remain unchanged. This establishes the observed
paused-state cause, not every possible historical dropdown incident.

New model-controls.cjs failed before the fix specifically on paused, then passed
eight state cases. Existing header-models.cjs passed12 layout/catalogue/restriction
cases. C# model tests passed2 groups after adding idle-paused selection before
the existing paused-with-queued-work refusal. No paid requests, live model changes,
credential reads or restart. Source-only; live picker acceptance remains pending.

Bill also confirmed cross-device theme independence after the prior rebuild.
Measure X site outage was repaired using its unchanged existing application and
tunnel launchers: local8090 and public HTTPS returned200, no model prompt sent.
Bill suspects unplugging/sleep; cause unproved, no power settings changed.

## Automatic pause ownership and one Stop speech control

Bill replaced the earlier Pause/Resume GUI design: Escape and one header Stop
speech button cancel current/queued TTS, preserve the draft, and release only
speech. Genuine typing (including whitespace) and speech OR together per browser;
clearing one cannot erase the other. Programmatic tool drafts remain nonhuman.
Agent mailbox now combines explicit runtime pause with transient actual connection
owners. GUI validates addressing, retains the actual held Agent, awaits release
on selection/disconnect, and never accepts a client-provided owner identity.
Child restart saves only explicit pause; automatic holds never survive restart.
Runtime pause/unpause API remains separate for CLI/library callers.

Verified: core ownership/model3groups and child restart2groups, shared341/0;
updated real speech regression including native-callback recovery/dialog focus/
queue/stale selection/manual runtime compatibility; dedicated actual-page Stop
speech/Escape/draft/two-tab/disconnect cases; three real streamed-turn modes
normal/missing-callback/disconnect all report actual Idle+unpaused. Twelve header
layouts,32 composer layouts and5public GUI groups pass. Chapters7/27 hard lint
pass; Chapter8 follow-up prose lint and final GUI build collected separately.
No paid calls, microphone/capture, installation or live restart.

Failures corrected: test subscribe was a no-op, so use get_settings ack; global
Escape preventDefault blocked native dialog closing/focus return, removed that
prevention; reused disposable fixture hit nonwritable prior Job output, retained
evidence and used fresh workspace. Two initially launched browser checks had
possible overlap; reran sequentially, passing. Broad crossover-browser runner
updated for current pause controls but stops before them on an unrelated old
composer DOM expectation (prompt/send versus current action/voice wrapper).
This runner is NOT claimed fully passing; dedicated real-page regression is the
current pause acceptance. Existing endpoints/archives remain unchanged.

Lesson: automatic pause is a transient owned relationship, not a persistent
boolean to clear blindly at reconnect. A global shortcut must also preserve
native dialog cancellation/focus behavior.

Final follow-up: Chapter8 hard lint passes with style advisories; final GUI build
clean, whitespace clean, disposable fixture/provider/Agents joined with exit0.

## Stop speech placement and deployment mismatch

Bill restarted/reloaded but still saw Pause/Unpause. Fresh localhost8088 HTML
contains those old controls; served app.js lacks automatic_pause. All six installed
application assemblies/symbols differ from the latest promotion build (comparison
reported different-or-missing). Therefore browser reload alone cannot install the
new implementation. No installation or restart performed by Cadence.

Bill requested the cancellation control beside the microphone for accessibility.
Moved the existing Stop speech button into a nonwrapping composer voice-controls
group immediately after mic, with type=button to prevent accidental submission.
Cancel voice recording remains separate. Updated Chapters7/27 construction guidance,
layout regression and actual-page selectors. Pause ownership code is unchanged.

Placement verification:32 composer cases include actual mic/stop adjacency,
44px targets and zero form submissions on Stop speech click;12 header cases
pass. Chapters7/27 hard lint pass with advisories; final GUI build clean. This
placement slice does not rerun the unchanged pause-ownership suite. Not installed.

### Playback beside the speaker label

Bill identified the floated playback control as a contributor to excess space around short messages and approved trying it beside the speaker label before Markdown work. Text cards now append playback to the existing heading like other cards; obsolete text-body float/clear rules are removed. The 44px target and original card callback remain. This frees the prose body from the control's minimum height, but does not establish that total card height is smaller: the heading now owns that height. Human visual acceptance remains pending.

Verified eight isolated real-render-function/real-CSS cases (320/1200px,18/28px,both roles): heading placement,44px target,independent short prose,no horizontal overflow,correct playback callback. Markdown is stubbed in this focused geometry fixture, not claimed tested. Four existing color/contrast cases pass. GUI build clean,zero warnings/errors. No installation or restart. Markdown implementation unchanged.

### Compact name row with overflowing playback target

Bill's screenshot showed the prior placement merely transferred excess height to the name row. At his suggestion, playback is now absolutely positioned centered on the name, retaining its44px target without sizing the row. A48px prose gutter avoids invisible hit-target overlap with links; minimum44px card height separates adjacent empty-card targets. Empty thinking suppression is not part of this change. Eight focused real-render/CSS cases verify compact name and whole short-card height, prose/target separation, adjacent empty-card target separation and playback callback at320/1200px and18/28px. Four contrast cases pass; GUI build zero warnings/errors. Visual acceptance pending; no install/restart. Lesson: measure whole-card geometry and clickable rectangles, not only prose height or visible glyphs.

### Full-width bubbles beneath transparent playback

Bill accepted the compact name layout visually, then rejected the prose gutter. Removed the48px message margin; the name still reserves icon space. Playback background/border are explicitly transparent at rest, with hover/focus behavior retained. Its44px target can intentionally overlap prose: transparent paint does not make clicks pass through, a tradeoff explained in chat. Eight updated geometry/playback cases and four contrast cases pass; GUI build clean. No install/restart; visual acceptance of this adjustment pending.

## Queued accessibility request: composer spoken label matches visible prompt

Bill reports hearing "Send message to Cadence" for the chat input while its
visible prompt says "Ask Cadence anything". He requests the spoken accessible
name match the visible wording exactly: "Ask Cadence anything".

Inspect the actual composer label/aria-label/placeholder ownership before editing.
Apply the selected Agent's name consistently to both visible and accessible text;
verify initial selection and Agent switching do not retain a stale name. Check the
browser accessibility name, then obtain Bill's screen-reader acceptance. This is
queued, not implemented or live-verified. No restart or deployment follows from
this note.
