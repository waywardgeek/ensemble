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
