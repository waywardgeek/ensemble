# Chapter 26: Voice and Images

> **Full narrative draft; acceptance and grading remain open.** Voice and image
> paths have implementation evidence and live acceptance as described below.
> Host-screen capture, model-turn delivery, delayed Jobs, and replay passed an
> agreed macOS synthetic-window test with a local fake provider before selection
> was simplified to whole displays. Updated whole-display native acceptance remains
> open. The image-files, attachment and voice external scopes pass, totaling 75 points;
> 25 points remain ungraded for host screenshots. Speech-pause cleanup
> now uses connection-owned holds as taught in Chapter 7. The narrative was drafted at the human's
> request; it is not a claim of complete implementation acceptance or grading.

## TL;DR: Specification Brief

Extend the working GUI with push-to-talk input, pasted images, screenshots of the
computer running the C# server, and Agent inspection of image files. Build on the media
parts and capability checks from Chapter 5 and the file authority from Chapter 18;
do not create a second conversation or media pipeline.

- **Hold to talk; release to send.** Provide a press-and-hold microphone button
  usable on a phone and a configurable keyboard shortcut. Right Alt is the
  initial choice where the browser permits it without interfering with AltGr,
  input methods, or reserved shortcuts. Press starts an explicitly authorized
  recording; release stops capture, transcribes it, and sends the resulting text
  once through the normal message path. While the Agent is working, that message
  follows the existing hint contract. Display recording, transcription, and
  failure states, and provide cancellation. Empty or failed transcription sends
  nothing. Lost focus, pointer cancellation, connection loss, or an agent switch
  cancels uncommitted work rather than sending partial speech later. Cancellation
  cannot retract a message already admitted. Late
  callbacks must not submit it or redirect it to a newly selected Agent.
- **Paste images into chat.** Accept image content from the clipboard, show a
  preview with removal controls, and send it with the accompanying text when the
  user presses Send. Preserve their order and the intended recipient. A preview
  is not delivery to the Agent. Validate format, size, and image dimensions before
  accepting the attachment; refusals leave the user's remaining draft intact.
- **Resize oversized images safely.** Use a maintained image library to decode
  and proportionally downsize images that exceed delivery dimensions or byte
  limits, without enlarging small images or overwriting original files. Bound
  source bytes, decoded pixels, frame count, and output bytes; reject unsafe or
  malformed input before expensive processing. Preserve readable detail where
  possible, report the resulting dimensions and any reduction, and refuse clearly
  if a useful bounded image cannot be produced. Deliver the actual resized bytes,
  not only resize metadata. Test this shared behavior for pasted images,
  screenshots, and image-file inspection; provider-specific limits remain explicit.
- **Capture the server computer.** Implement the C# equivalent of CodeRhapsody's
  host-side `list_displays` and `screenshot` tools. Enumerate displays and their
  bounds; capture one whole selected display (display zero when omitted) on the
  machine running Ensemble. Return actual bounded PNG pixels through the existing
  image-tool path, with capture and output dimensions, not just a filename or DOM
  description. A request from a phone captures the server computer, not the phone.
  This operation does not require a connected Chrome page or a browser sharing
  picker. Require explicit host-screen tool authority and respect OS screen-
  recording permission; refuse denied access, invalid bounds, unavailable displays,
  and unsupported platforms clearly. Do not elevate privileges or bypass OS
  permission. Capture only on request, not as background screen monitoring.
- **Inspect an image file.** Provide an image-reading operation for an explicit
  path, using the Agent's existing Sandbox and supported media representation.
  Enforce file authority, format, and resource limits; report missing, forbidden,
  or unsupported input honestly. The operation must deliver the image to the
  selected model, rather than treat a path, browser-only preview, or textual
  metadata as proof that the image was seen.
- **Preserve the existing owners.** The browser owns microphone capture and
  attachment composition; the server owns native screenshot capture and its
  resource cleanup. The Agent owns tool authority, admission, and History; the
  Engine renders supported media for
  its provider. Keep media references and provenance in the established parts,
  not base64 copies in ordinary transcript text. A model recommendation is not a
  media capability declaration. Unsupported models or placements refuse clearly;
  they must not silently discard an image, switch vendors, or change funding.
- **Make permission and cleanup observable.** Stop microphone capture on release
  or cancellation and show when it is active. Captured speech, pasted images,
  screenshots, and image-file contents are private input, not material to collect
  speculatively. Preserve secure-context requirements and the existing
  authenticated-origin policy; do not weaken them to obtain capture access.
  Transcription needs an explicitly selected service and funding policy; it must
  not silently charge an API key when subscription funding cannot
  perform the operation. Playback/capture cleanup must not leave an automatic
  pause behind or undo a manual pause. Replays must not restart capture or send
  another message.

Before implementation, settle the transcription route, supported capture
mechanism, resource limits, and authorized live tests with the human. Use local
fixtures for deterministic input, cancellation, routing, and refusal checks.
Those checks do not establish real microphone behavior, actual image delivery to
a vendor, or phone usability; each needs evidence of its own.

## In Plain Words

Bill was using Ensemble from his phone while away from his desk. Text input
already worked, but the phone made the missing inputs obvious. Describing a
picture takes work that the picture could do directly. Typing a message occupies
hands that could instead hold a microphone button for a moment.

The voice interaction is deliberately simple: hold, speak, release. Releasing
means send, not open another confirmation step. That convenience makes the
failure behavior important. A recording belongs to the Agent selected when it
began. If the person switches agents or loses the connection before it can be
sent, the old recording must not suddenly appear in another conversation. A late
transcription result is not fresh permission to send it.

Once transcribed, the speech is an ordinary message. It does not need a special
scheduler or a second chat history. If the Agent is busy, the existing hint
behavior lets it hear the person without throwing away prepared work. A visible
cancel action gives the person a way to abandon the recording instead of sending
an accidental fragment.

Pasting an image is similarly familiar: paste, check the preview, add words if
needed, and send. But the preview and the model's input are different things.
Displaying the picture in the browser proves only that the browser can display
it. The Agent needs the actual image through the media path supported by its
model. Otherwise, an apparently successful attachment would invite an answer
based on a filename or a guess.

Screenshots and image files meet at that same image-delivery path. The screenshot
tool captures a whole display on the computer running the C# server, just
as CodeRhapsody does. The image-file tool reads a named file under the existing
file policy. Neither needs a second conversation history. File-read permission
alone is not permission to capture the host screen.

This distinction matters when the person is using a phone: the phone can ask the
Agent to inspect the server computer's screen without sharing a Chrome tab or
opening a browser picker. The operating system may still require screen-recording
permission. If permission or platform support is absent, the tool must say so. A
structured page snapshot is not a screenshot, and the separately implemented
browser-tab capture is not a substitute for this host-side operation.

The important boundary is not whether the interface has a microphone or camera
symbol. It is whether the person's intended input reaches the intended Agent,
with permission, once, and with a clear failure when it cannot. That boundary is
what lets the same interaction remain trustworthy on a desktop or a phone.

## 26.1 Four inputs, one conversation

The new inputs should arrive at the conversation the person already understands.
An image pasted beside a caption belongs to that message. A spoken sentence,
after transcription, is a normal human message. An image inspected by a tool
belongs to that tool call and its result. None of these paths needs another chat
history or a private queue that competes with the Agent's mailbox.

Their acquisition paths differ:

```text
clipboard or file picker → browser draft → server preparation → AgentInput
microphone → bounded WAV → transcription → original Agent's mailbox
Sandbox image file → preparation → typed tool result
whole host display → native capture → preparation → typed tool result
```

After admission, the existing owners take over. The Agent records History and
schedules work. The Engine renders a frozen request for its provider. The GUI
projects that history for the person. A busy Agent receives a transcribed message
through the same hint behavior as typed input; voice does not interrupt a model
request by inventing a different scheduling rule.

A preview proves that the browser decoded something. A filename proves that a
name was returned. Neither proves that the model received pixels. The useful
assertion is further downstream: the actual provider request contains the
expected image bytes in the correct human message or paired tool result.

That distinction shapes both implementation and tests. Keep images as typed
media parts. A sentence such as “Screenshot saved to capture.png” is useful
metadata, but it cannot stand in for the screenshot. Likewise, audio recording
success is only the start of a voice operation. The intended Agent must eventually
receive the transcript, or the operation must fail without sending it elsewhere.

## 26.2 Put ownership where the lifetime lives

The browser owns the live microphone track, its audio context and worklet, and
the buffers used to make a recording. Its image composer owns pending file
readers and preview URLs. Selection changes can therefore invalidate unfinished
work immediately and release the resources those operations created.

On the server, an Agent keeps its existing Sandbox for image-file access. Host
capture has a different authority boundary: the operating system's displays are
not files inside that Sandbox. The native backend owns its short-lived helper
process and private scratch directory. It returns bytes into the same preparation
pipeline rather than exposing a new general-purpose filesystem API.

Voice transcription is an optional child of the Ensemble application. One
application-owned client holds the separately selected payer and joins its HTTP
work during destruction. A socket's voice lease borrows the actual Agent selected
when recording began. It does not look up a fresh recipient after transcription
finishes. The connection owns the lease and its notification work and joins them
when the connection ends.

These relationships matter most when something changes. A different selected
Agent does not inherit an old recording. A replacement socket does not inherit
permission to resend an old commit. Destroying a resource owner cannot simply
forget a task because its cancellation token was signalled.

Use immutable values at the boundaries. Prepared images own their output bytes;
validated voice audio owns a snapshot of its canonical WAV. Once admitted, these
values do not depend on a browser array that can still change or an original file
that might be removed. The provider renderer consumes the frozen request, not
whatever happens to be on disk later.

## 26.3 Prepare an image once, with bounded work

Image preparation is shared by attachments, image-file inspection, and host
screenshots. This avoids three independent interpretations of orientation,
resizing, and acceptable output.

The source policy admits static PNG, JPEG, WebP, and GIF images. It bounds source
bytes before unbounded reading, checks decoded dimensions before full raster
processing, refuses animation, applies EXIF orientation, and strips metadata from
the delivered derivative. Resizing preserves aspect ratio and transparency, never
enlarges a small image, and never overwrites the original. The derivative is PNG.

Bill asked for large images to be resized rather than merely rejected. The
derivative needs to remain useful without changing the source. The current
application limits serve different purposes:

- A source image is at most **64 MiB** and **24 million decoded pixels**.
- A prepared image has a longest edge of at most **2,048 pixels** and contains at
  most **8 MiB** of PNG bytes.
- One interactive input admits at most **four images**, each within that prepared
  limit, with **16 MiB** of prepared image data in total.
- The GUI separately bounds its combined source-image bytes at **64 MiB** before
  preparation. A source-upload allowance is not a prepared-message allowance.
- Each provider render has a fresh allowance of **32 image occurrences** and
  **32 MiB** of image payloads, shared across human, tool, and ephemeral images.

The last allowance is a request-admission check, not permission to delete older
images from History. Repeated appearances consume it too. It excludes other
request overhead, such as base64 expansion and JSON, and does not establish a
provider's token capacity or account entitlement.

### When an image limit made a working Agent look idle

During page-by-page review of a scanned document, Cadence stopped responding to
messages and automatic reminders. Restarting restored the same failure. Bill
asked another Agent, Test, to investigate. Test reported 33 retained tool-result
images totaling about 15 MiB: the count limit had been crossed, even though the
byte allowance had not. Request construction failed locally before any provider
request was sent. Each failed turn returned Cadence to idle with the same images
still in active context.

The guard protected the request, but its placement trapped the recovery. Asking
Cadence to call `micro_handoff` required another model request, which the same
guard refused. Token-pressure handling could not substitute for image accounting.
The GUI compounded the problem: it narrated explicit error events, while this
local failure recorded only turn completion with an error reason. Bill heard
silence instead of an explanation.

Test used the external **Clear tools** control to remove the image-bearing tool
context without deleting canonical History or project files. Bill confirmed that
Cadence responded again, then requested a proper handoff. The permanent lesson
is to preserve a usable request *before* admitting the next image, and to report
local failures through the same visible error path as provider failures. Raising
the count limit would merely postpone the next trap.

A regression must exercise the whole sequence: fill the image allowance, request
one more image, receive an explicit refusal while the previous images remain,
checkpoint, then retry successfully. Also test queued human attachments, byte
exhaustion, and an already-over-limit saved context. A renderer test that merely
expects the limit exception does not establish that the Agent can recover.

The Agent checks cumulative image pressure before accepting new human images or
publishing image-bearing tool results. Queued inputs reserve their allowance too.
A human refusal admits neither the caption nor the images and preserves the GUI
draft. A refused tool delivery records an explicit error instead of claiming the
model received pixels; the Job retains its output. Automatic snapshots follow the
same rule. Existing images remain available so the model can write a deliberate
`micro_handoff`, then retry the refused image. No automatic image retirement or
silent tool-result trimming is authorized by this safeguard.

A previously saved over-limit context can still fail rendering. That refusal is
recorded as a visible, narratable error before turn completion, without a provider
request. For retained tool images, the external **Clear tools** control can restore
headroom without requiring inference or deleting canonical History. It does not
clear human attachments; those require an explicit handoff through host control.

Preparation first chooses a bounded size. If the PNG remains too large, smaller
attempts are sampled from the original decoded raster rather than repeatedly
resampling a derivative. When a useful bounded output cannot be produced, the
whole operation refuses. Report original and delivered dimensions so the model
and person can reason about the detail that remains.

The implementation uses pinned SkiaSharp packages. Choosing that dependency
required testing the actual Debug and Release paths, not merely finding a
library with a convenient API. An earlier library candidate encountered a
Release-build licensing requirement and was replaced. The general lesson is to
check the package version, distribution conditions, and native assets you will
actually ship. A successful restore alone does not verify those assets on every
operating system.

These limits bound accepted input and allocation; they are not isolation from
bugs in native decoders, a process-wide concurrency budget, or a guarantee of
readable small print. Keep those claims separate.

## 26.4 Inspecting a file must deliver the file's pixels

`inspect_image` starts with ordinary tool authority and a supported image-result
surface. Check the selected model's capability before reading the path. An
unsupported model should receive a clear refusal, not cause private file access
followed by silent omission.

Then use the Agent's existing Sandbox to read and prepare the file. Do not add an
absolute-path fallback because a confined read failed. In the public API, the
preparation and access to the pixels are deliberately small:

```csharp
PreparedImage image = agent.PrepareImage(path);
BlobPart pixels = image.Image;
```

The tool result contains those pixels together with dimensions and reduction
metadata. It remains paired with the original tool-call ID. Errors such as a
missing file, forbidden path, malformed image, or unsupported format are explicit;
they do not become an apparently successful metadata-only result.

Acquisition is performed once. Rendering a later turn should use the retained
prepared bytes, not reopen the source path. Deleting the original after successful
inspection is therefore a useful test: the first request and a later request must
still contain identical prepared pixels.

Tools retain their ordinary Job behavior. A foreground observation can return a
running Job before image preparation finishes. The Job must retain a structured
terminal result, including media, for later observation. A filename in an output
log is not enough. Ending an observation does not cancel that Job, and an image
must not be flattened into a JSON string merely because it arrived after the
first observation window.

Retention policy is also distinct from output allowance. A person may deliberately
ask for a large result and keep it for discussion. Chapter 14's explicit retention
and handoff rules govern that decision; the former size-based ladder must not
silently override a kept result. A deliberate handoff removes media from derived
active context while retaining accompanying text according to that reducer's
contract. It does not prove that the original GUI history or every retained log
entry was deleted.

## 26.5 Render the media the provider actually supports

The shared value is a typed image part. The wire encoding is provider-specific.
Responses represents image-bearing function output as typed content alongside
text; text-only results retain their simpler representation. Anthropic places
image content inside the appropriate tool result. Gemini's supported
function-response image form requires its own parts and references, along with
preserved call identity and thought-signature information where applicable.

Do not infer those contracts from model names. Human-input images and images in a
function response are different capabilities. The implementation has an explicit
function-response-image capability for the verified `gemini-3-flash-preview`
path. That is not a promise about every newer Gemini model or an older fallback.
Unsupported nesting or placement must refuse rather than disappear.

Rendering must preserve ordering, call pairing, and the existing shared image
budget. It must not mutate canonical History to satisfy a provider. Nor may it
switch vendors, choose another model, or silently change the payer because the
current selection lacks a capability.

Test the actual outgoing request. A local fake provider can assert that a specific
tool-call ID is answered by one image-bearing result and decode its pixels. A
bounded real-provider probe establishes something additional: that the selected
API and model accept the representation. Neither result implies entitlement or
support for a different model, streaming mode, or platform.

## 26.6 An attachment draft belongs to one recipient

The browser's image composer retains the text and images that the person is
preparing. It provides previews, removal, and an explicit file-picker fallback to
clipboard paste. File reads are bounded and cancellable; preview URLs are revoked
when no longer owned. Late file-reader completion cannot attach an image to a
newly selected Agent.

Sending captures the original socket, Agent address, and a monotonic per-socket
sequence. The server prepares the images once, admits the text and prepared
images together, and returns a matching acknowledgement with dimensions. This
acknowledgement means actor admission, not completed inference.

On a clear refusal, keep the draft available for correction. On success, clear
only the matching draft. Ignore acknowledgements for another socket, Agent, or
sequence. If the connection disappears after sending but before acknowledgement,
the outcome is uncertain. Automatically resending could duplicate a message the
Agent already accepted. Preserve that uncertainty and require an explicit human
decision instead of treating reconnect as permission to retry.

Keep layout work part of the interaction contract. The composer now gives the
text input its own full-width row, with attachment, microphone, and Send controls
below it. A phone-sized viewport and enlarged text must not squeeze the writing
area to make room for another button. Accessible names and visible states matter
more than a color change alone.

Laptop image delivery was accepted in live use. Android paste was later accepted
after reinstalling Chrome's web app. That establishes the successful interaction;
it does not establish the underlying cause of the earlier failure. Retain the
file picker as a useful explicit attachment path without declaring all Android
clipboard behavior either broken or solved by that one observation.

## 26.7 A screenshot of which computer?

The screenshot requirement exposed a specification mistake. The first
implementation captured the connected browser page with Chrome's sharing API.
It produced actual pixels and passed isolated native-tab tests, but it captured
the wrong surface for the requested feature.

Bill clarified that he meant CodeRhapsody's host-side capture: ask the server to
capture its primary display, or select another from its display list. Cadence
had implemented a different feature. The clarification changed the remaining
work even though the browser tests passed. A person
using a phone can request a screenshot of the server's screen. There is no need
to negotiate a Chrome tab picker on the phone.

The two features remain distinct. Browser capture has browser gesture and picker
requirements. Host capture has explicit tool authority and OS screen-recording
permission. Success in either path does not establish acceptance of the other.
The additional Ensemble confirmation dialog in the browser implementation was a
design choice, not a separate requirement imposed by the human.

The host interface follows the reference tool shape:

```json
{"name":"list_displays","arguments":{}}
{"name":"screenshot","arguments":{"display":0}}
{"name":"screenshot","arguments":{}}
```

These are example tool requests, not a script to run against an arbitrary private
screen. Agree on the whole display before a native test. The only selection
argument is optional `display`, a nonnegative zero-based integer. Omitting it
selects display zero. There is no window measurement or rectangle selection.
Old rectangle arguments refuse rather than silently expanding a region request
to a whole-display capture. Strict interfaces that require a value can send zero.

This simpler contract followed a live debugging failure: the model-facing surface
supplied display and rectangle fields together, while the backend required one or
the other. Repeated retries did not resolve that mismatch. Bill removed rectangle
selection because window measurement was distracting from the debugging task.
Test the advertised schema as well as the parser: an omitted selector and explicit
zero must choose the same display, and no rectangle fields should be advertised.

The source `host-screen` skill grants these tools separately from `code-tools`
and browser access. Safe mode and unskilled legacy registries refuse them, and
an unsupported image model is rejected before native screenshot acquisition.
Recheck authority and cancellation before returning the prepared pixels. Installing
a tool definition alone is not a grant to see the desktop.

The current backend is **macOS-only**. It obtains bounded display metadata,
preflights screen-recording permission, and runs the fixed built-in
`/usr/sbin/screencapture` executable with a fixed argument structure. There is no
model-selected shell command. The helper receives an empty environment and writes
to private owned scratch. Its deadline is 15 seconds. Cancellation kills and joins
the owned process and output readers before scratch removal; cleanup errors must
not hide the original failure.

The backend checks source-pixel bounds using conservative display scaling, then
passes returned bytes through shared image preparation. Output dimensions must
agree with the selected display and supported scale. Each request captures one
whole display, not a composite of several. Other operating systems refuse explicitly. These are current
implementation limits, not claims of full cross-platform parity with CodeRhapsody.

Permission preflight currently refuses denied access; it does **not** open an OS
permission prompt automatically. Before the display-only simplification, an agreed
macOS synthetic-window test verified the native backend and tool-delivery path. It
exercised native capture through the real Agent tool, a running Job followed by
collection, actual pixels in the local fake-provider request, and identical replay
on the next human turn. It made one capture and four local provider requests.
This is not a paid-provider interpretation test, other-platform acceptance, or
proof of every permission-change race. Display-only native model-turn acceptance
requires a newly agreed whole-display fixture; old rectangle-based evidence does
not certify that changed selection contract. The earlier browser replay test remains
separate from host-screen acceptance.

That native test caught a HiDPI assumption the synthetic metadata tests could not.
A 640×360 logical rectangle produced 1280×720 pixels. `CGDisplayPixelsWide/High`
reported 1920×1080, while the current display mode's pixel dimensions were
3840×2160. Read backing dimensions from the copied display mode and release that
native object in `finally`; do not assume logical coordinates equal raster pixels.
This matters before capture, because the source-pixel safety bound must include
scaling. Verify the selected display's actual pixels as well as its dimensions;
simply relaxing a failed dimension assertion is not acceptance.

## 26.8 Make the recording format small enough to reason about

Bill specified hold, talk, and release to send, without an editable transcript or
another Send press. That interaction does not require every audio container a
browser or provider can produce. The current recorder makes one canonical representation: mono,
16 kHz, 16-bit PCM in a 44-byte-header WAV. At 60 seconds that is 960,000 samples,
or **1,920,044 bytes including the header**.

The browser recorder owns its original audio tracks, audio context, worklet,
message ports, and sample storage. Keeping the original track membership is
important: a mutable stream object must not make cleanup forget an original
track or stop an unrelated track added later. The worklet provides bounded sample
processing without playing the microphone back through the speakers.

Release stops capture and finishes one WAV. Cancel, device loss, processing
failure, or exceeding the duration limit produces no partial send. An asynchronous
permission result or worklet load may arrive after cancellation. Recheck liveness
before using it, stop late tracks, and join cleanup. Cancelling an operation does
not mean the browser's native permission prompt or module load can be forcibly
aborted.

The server constructs `PcmVoiceAudio` only from a validated copied snapshot. It
checks the exact canonical header, matched lengths, sample alignment, nonempty
data, and the duration ceiling before paying for transcription. Duration derives
from sample count, not a browser clock or an untrusted container field. This is a
format validator, not a general WAV decoder or a speech detector; valid samples
can still be silence.

The hold controller sits above the recorder. Releasing before permission and
recording are ready cancels the gesture instead of turning a late permission
grant into background listening. It owns repeated releases and delayed finish
callbacks. The page-level control adds pointer loss, blur, visibility, selection,
and socket cancellation, plus the visible Cancel action.

## 26.9 Keep the transcription payer explicit

Transcription is a separate service operation from the Agent's chat inference.
A running application with subscription-funded chat must not silently use an API
key because that subscription cannot transcribe audio. The GUI host defaults to
transcription off.

To enable the implemented route, append explicit options to the authorized GUI
host launch:

```text
--transcription openai-api-key
--transcription-key-file /path/to/private/settings.json
```

The selected file must contain a nonempty `openai_api_key`. Alternatively, the
explicitly enabled route can use `EN_TRANSCRIPTION_API_KEY`. Do not put the secret
itself in command arguments, a fixture, or a transcript. An explicitly selected
bad key file fails closed; it does not fall back to another credential source.
Disabled transcription does not read the file. Advertising availability exposes
payer and model metadata, not the key.

The current client uploads the canonical WAV to OpenAI's transcription endpoint
using `gpt-transcribe` and a non-streaming JSON response. Its fixed policy includes
one active request per application-owned transcriber, no automatic retry, a
45-second maximum request timeout, a 65,536-byte response bound, and a 32,768-byte
transcript bound. Missing usage remains unknown; supplied malformed usage refuses.
Token usage and duration usage are distinct, and neither is fabricated from the
recording's duration.

The single active request is a deliberate concurrency limit: simultaneous voice
requests can receive a busy refusal rather than building an invisible paid queue.
Cancellation invalidates delivery even if a transport returns late, and destruction
joins that work before disposing its owner. Provider errors and cleanup failures
must not expose the key. Raw audio does not belong in ordinary chat history or
unredacted frame traces.

Verify current [speech-to-text documentation](https://developers.openai.com/api/docs/guides/speech-to-text)
and [model pricing](https://developers.openai.com/api/docs/models/gpt-transcribe)
before a live test. Published rates are estimates for planning, not a measured
bill. The accepted Dev test used an explicitly selected OpenAI key for
transcription while retaining its existing chat funding.

## 26.10 Separate readiness from commitment

Bill's release-to-send requirement needs an admission protocol as well as a
correct recorder and transcriber.

A transcription result can be correct and still belong to an operation that the
person has abandoned. The voice protocol therefore separates acquisition from
commit:

```text
browser                          connection-owned server lease
voice_begin  ------------------> bind original Agent and sequence
             <------------------ voice_begun
voice_audio  ------------------> validate WAV, transcribe once
             <------------------ voice_ready
voice_commit ------------------> consume once, admit to original Agent
             <------------------ voice_accepted
```

Every request carries the original Agent address and sequence; the browser also
keeps the original socket. An empty-string address is valid for an unnamed root,
so check its type rather than rejecting it for being falsey. The server lease
borrows the real Agent selected at begin. Comparing that relationship at commit
prevents a delayed result from following the GUI's latest selection.

`voice_ready` does not carry a draft transcript for another Send button. It tells
the still-current browser session that transcription completed. The session
rechecks its binding and automatically sends `voice_commit`. This preserves the
required hold–talk–release interaction while giving cancellation a final boundary.

The server consumes the lease before posting the ordinary message:

```csharp
consumed = true;
commit = Target.PostAsync(
    new AgentInput(InputKind.Message, transcription.Result.Text));
```

In the actual implementation this occurs under the lease gate, after checking the
selected Agent, sequence, cancellation, prior consumption, and successful
transcription. The acknowledgement reports admission. It does not report model
completion, guarantee an audible response, or retract a message if the browser
subsequently disconnects.

After commit may have been sent, an error or lost acknowledgement is uncertain,
not proof of non-delivery. Do not retry that capture on a replacement socket.
The person can inspect the original conversation. A new intentional recording is
a new operation; an automatic replay of the old one is not.

## 26.11 Cancellation must win before the next effect

A state notification is a callback boundary. Tests exposed a race in which a
listener cancelled the session during a state change, but execution continued to
the next send. Setting a variable earlier in the method was insufficient.

The coordinator now rechecks terminal state and current binding at the send
boundary. It also publishes its cleanup promise before notifying observers, so a
reentrant cancellation can join the work that is actually being cleaned up. These
rules are useful beyond voice: publishing an event can run code that invalidates
the assumptions of the caller.

Cleanup has two sides. Browser cancellation stops capture and makes late callbacks
inert. Server lease disposal invalidates transcription/commit and joins the tasks
it owns. A failed or cancelled request stays failed on its own task; disposal does
not turn it into success. A consumed admission remains consumed, even when its
outcome is uncertain to the browser.

The microphone control supports pointer/touch holding and Space or Enter while
the button is focused. The persisted page-level shortcut offers Right Alt, F8,
or Off. Right Alt is refused for Apple Option-key use and AltGr/composition or
modifier combinations; it is not a portable system-wide hotkey. Repeats cannot
start another recording. Changing the preference, losing focus, or forming another
key chord cancels instead of submitting a fragment. Native Windows and Linux
keyboard behavior still needs platform acceptance beyond the simulated guards.

An existing text or image draft is not silently bundled into a voice send. Resolve
that draft first. Keep visible recording, transcription, cancellation, failure,
and uncertain-commit states; a person listening quickly still needs to know
whether the recording has stopped and whether a message may already have arrived.

## 26.12 Test the boundary you intend to claim

Use deterministic fixtures to expose races, then use bounded native checks for
the device and provider behavior those fixtures cannot establish.

For images, decode the outgoing PNG and inspect pixels, dimensions, and pairing.
A randomized image prevents the fake provider from returning a memorized color
answer. Verify that preparation leaves the source unchanged, then delete it and
verify replay. Negative candidates should return plausible metadata without
pixels or lose the retained payload; the grader must reject both.

For voice, test the recorder, hold controller, browser session, and server lease
separately, then combine the real components. Synthetic browser audio can exercise
the actual worklet and canonical WAV production. A held fake transcription
response lets the test cancel first, release the late response, and prove there
is no later admission. Duplicate commits, wrong sequences, wrong sockets, and
selection changes should fail without redirecting speech.

The integrated voice fixture uses real WebSocket leases, the transcription
multipart/parser path, Agent admission, a local fake model request, and canonical
History. It bounds the number of requests and joins the browser, host, provider,
and Agents. In a later live Dev test, Bill released the microphone and saw the transcript,
but initially thought the Agent had not received it. Cadence checked the durable
events: the Agent had processed the test messages and completed replies. Bill
then accepted the interaction as successful. That established live hold/release
transcription and response, not a display bug or a guarantee about every future
recording.

Fixtures also need review. A wrong model ID can refuse media before the tested
path begins. Omitting a non-streaming field is not equivalent to requesting
streaming. Reading History after destroying its owner is an invalid assertion.
Those were fixture corrections, not product failures or passing feature evidence.

The external CLI grade initially found a missing SkiaSharp assembly because its
old `--no-restore` assets predated the image dependency. An offline asset refresh
from already-cached packages fixed that run without a runtime source change.
Build state, source state, installed state, and live acceptance are separate facts.
Do not hide a stale-manifest workaround inside a claim that the product was fixed.

## 26.13 What is verified, and what remains open

At this draft boundary, image-file inspection has local provider-wire and replay
tests plus a bounded live test for the exact supported Gemini path. Desktop and
Android image attachment have human acceptance. Voice has component and integrated
fake-provider tests, and the live Dev interaction has been accepted. Those facts
do not establish every supported platform, every provider model, or every failure
race.

Host-screen tools are implemented in source with grant, safe-mode, argument,
preparation, renderer, and fake-process cleanup tests. Native macOS acquisition,
model-turn/delayed-Job delivery, and subsequent replay passed with an agreed
synthetic window and a local fake provider. After simplifying the public tool to
one display selector, Bill restarted the staged runtime and authorized a whole
primary-display capture. The live `screenshot(display: 0)` call delivered actual
3840-by-2160 pixels prepared to 2048-by-1152, with no window measurement, picker,
or command-line workaround. Cadence inspected the returned GUI pixels. Omitted
selection is source-tested; it was not a separate live invocation.

The shared implementation suite passed 339 groups with the child-restoration
changes; those groups span the application and are not 339 independent multimedia
acceptance tests. Installation, Bill's reported restart, and the successful live
capture are separate observations. Neither the earlier synthetic-window test nor
this earlier authorized display capture awarded the later external host-screen grade.

The **20-point image-files**, **25-point attachment**, **30-point voice**, and
**25-point host-screen** external scopes now pass, totaling **100/100**. The native
host run used the agreed synthetic primary display, both selector forms, distinct
delayed Jobs and full-payload replay on a new turn without recapture. This completes
the external point allocation, not every platform or human-use acceptance claim.

Speech output now follows Chapter7's browser-local preferences and one Agent-wide
pause state. A tab addresses only its selected Agent. Stop speech and Escape
cancel local current/queued speech, preserve typed drafts, and explicitly resume
from any device. Unchanged local activity cannot veto that resume. Disconnect
does not invent resume, and reconnect does not reassert stale activity. This repair is separate evidence,
not a substitute for the separate voice-input external grade.

This draft describes the implementation before the remaining acceptance and
grading gates close. Update this status from evidence as the work finishes.
A written chapter, an installed button, a passing fake-provider test, and a live
human interaction each establish something different.

## 26.14 Review and operate the feature deliberately

Before native capture, agree on a synthetic display and arrange OS permission with
the person at the host. Do not sweep a private desktop to discover whether access
works. Confirm actual returned pixels and dimensions, then exercise a model turn
and delayed Job. Record platform refusals and the cross-display limitation rather
than substituting browser capture for this test.

Before a live voice check, verify the explicit payer and refresh the tested page.
Use one short phrase first. A permission prompt may cancel the initial hold;
allow it, then begin a new deliberate hold. Release should stop capture and send
without another confirmation. Verify the original Agent's response as well as
the transcript appearing in chat.

Keep installation separate from testing. Stage only authorized runtime files,
preserve launchers and credentials, and wait for the old process to exit fully
before replacing a running host. A closed listener alone does not prove that
History persistence and Agent destruction finished. Refresh browser assets after
a verified rebuild; neither a source change nor a build proves the person is
using the new page.

These interactions extend the existing Agent rather than replace it. Pixels enter
as typed media; accepted speech enters as ordinary text. Ownership, bounded work,
explicit authority, and durable admission give the new controls their meaning.
The following exercises test those relationships, not just whether a button can
be made to light up.

## Exercises: implementation contract

These exercises extend the same Agent, History, Sandbox, and GUI. They do not
ask for a parallel media service or a second conversation store. The proposed
chapter allocation is 100 points. The image-files, attachments and voice commands
below award separate 20-point, 25-point and 30-point scopes. None awards the
host-screen scope by itself or certifies the entire chapter.
The implementation's focused tests remain evidence, not external grades.

### Exercise 26.1: Image-file delivery and replay (20 points)

Implement `inspect_image` as a skill-authorized tool. Read the selected file through
its Agent's Sandbox, prepare a bounded PNG without changing the original, and
return actual image content in the matching tool result. Capability validation
must precede file access. Keep errors explicit: a filename or a description of
an image is not image delivery.

The initial executable scope checks a narrower, concrete delivery contract:

- Inspect a randomized 4096-by-64 PNG and prepare a 2048-by-32 PNG with the
  expected interior pixels. Keep the tool-call ID and image placement intact.
- Preserve the source byte-for-byte during inspection. Then retain the prepared
  pixels after the source is deleted, so a later human turn can discuss them
  without another file read or capture.
- Deliver exactly one matching image result. Metadata-only content, absent or
  corrupt bytes, wrong dimensions, wrong call pairing, duplicate images, and
  missing replay must not earn credit.

Use a real candidate CLI and a local fake provider. The grader creates a private
workspace and a synthetic skill grant; it supplies no live vendor credentials.
It observes the actual provider requests and joins the candidate process. Its
fake provider does not prove real-provider availability or model understanding.

Run the **Edition 3** command against an executable wrapper for the candidate's
JSON CLI:

```sh
go run ./cmd/grade-edition3 -ch 26 -executable-only \
  -media-scope image-files -json /absolute/path/to/candidate-cli
```

Do not substitute `./cmd/grade`: its Chapter 26 is the preserved Gateway grade.
The new command reports a maximum of **20**, not 100. A 20/20 result certifies only
this delivery-and-replay scope. Sandbox escapes, resource ceilings, unsupported
models, cancellation, and cross-provider correctness still require their focused
implementation tests and review; this scope does not certify them by implication.

### Exercise 26.2: Addressed image attachments (25 points)

Accept pasted images and explicit file selection into a removable preview draft.
Keep accompanying text and image order, then send the whole message to the Agent
and connection selected for that draft. A preview or client-side send is not
actor admission: use the matching server acknowledgement before clearing it.

Test whole-message refusal, late file-reader completion after selection changes,
duplicate acknowledgements, and reconnect after an uncertain send. Refusal must
preserve the usable draft. Uncertainty must not trigger automatic retransmission.
The server must prepare and validate the real image bytes; a client MIME string
is not sufficient validation. Check usable controls on a narrow phone viewport,
but record native clipboard acceptance separately from viewport geometry.

The external grader must observe the real addressed WebSocket path and the
resulting model input. Its negative controls must reject preview-only delivery,
lost accompanying text, a changed recipient, and duplicate admission. This
25-point scope now has a passing executable driver. Run it against an ordinary
GUI executable wrapper that forwards its arguments:

```sh
go run ./cmd/grade-edition3 -ch 26 -executable-only \
  -media-scope attachments -browser-node /absolute/path/to/node \
  -playwright-module /absolute/path/to/playwright \
  -json /absolute/path/to/candidate-gui
```

The grader requires an already installed Playwright package and Chrome. It checks
browser availability before launching the candidate; setup failure exits without
a candidate score. It downloads nothing. The candidate accepts the ordinary GUI
port, workspace, data-directory, provider/model/base-URL, funding and no-save
options. The grader supplies a fresh key-free home and synthetic credential, not
a running GUI or internal test host.

Eight scenarios cover picker, synthetic paste/removal, separate recipient
histories, late readers, refusal recovery, uncertain reconnect, acknowledgement
ordering, and repeated wire envelopes. Randomized prepared pixels and captions
must reach the loopback provider exactly once. Preview-only, lost-caption,
misdirected and duplicate-send mutants fail named gates. A phone-sized viewport
checks layout, not native clipboard support. The command reports **25/25**, not
100/100; image-file replay remains its separate scope.

### Exercise 26.3: Hold, release, and send once (30 points)

Connect the microphone button and configurable page-level shortcut to bounded
recording and explicitly funded transcription. Release sends once through the
original Agent/socket lease, without an editable transcript or another Send step.
Transcription readiness is not permission to redirect or duplicate admission.
Keep the transcription payer separate from the running Agent's chat funding.

Use synthetic audio and a bounded fake transcription provider first. Observe the
real WAV, transcription request, lease, actor admission, and model request. Test
permission arriving after release, repeat keydowns, duplicate replies, blur,
Agent switching, disconnect, cancellation during an uncooperative response, and
cleanup failures. Stop owned tracks and join owned work. A result that arrives
after cancellation cannot become a late message. After a commit whose outcome is
unknown, report uncertainty rather than retrying.

The external grader must reject extra confirmation requirements, duplicate or
misdirected sends, continued capture after cancellation, and implicit API-key
fallback. Fake audio does not certify a native microphone or a paid API. Keep
human phone acceptance and provider evidence separate. This 30-point scope now
has a passing executable driver:

```sh
go run ./cmd/grade-edition3 -ch 26 -executable-only -media-scope voice \
  -dotnet /absolute/path/to/dotnet -browser-node /absolute/path/to/node \
  -playwright-module /absolute/path/to/playwright \
  -json /absolute/path/to/candidate-runtime-directory
```

Unlike the executable wrappers above, the voice candidate is a directory containing
`Ensemble.dll` and `Ensemble.Gui.dll`. The grader compiles its own public-API
consumer against those assemblies with an installed .NET 10 SDK and empty package
sources. Missing browser/SDK setup refuses before scoring; failure to satisfy the
public consumer API is a candidate failure. No package installation or paid call
is part of this command.

The public `VoiceTranscriber(parent, key, HttpMessageHandler)` overload allows a
trusted host to supply synthetic HTTP responses without changing the production
URL or 45-second deadline. Handler ownership transfers only after successful
construction; refused construction leaves it with the caller. Destruction joins
active requests before disposing an owned handler. Use synthetic credentials in
a grader transport, never a real payer key.

Eight modes cover normal hold/release and duplicate commit, explicit funding off,
permission after release, disconnect and Agent switching during capture or held
transcription, and provider failure followed by a fresh deliberate recording.
The oracle independently checks WAV bytes, exactly one human transcript in model
input, canonical admission across both Agents, ended tracks and joined cleanup.
Four mutants expose extra confirmation, duplicate commit, wrong recipient and
actual track leakage. The command reports **30/30** for this synthetic scope,
not native microphone reliability, real-provider behavior or the host-screen grade.

### Exercise 26.4: Host screenshots (25 points; native acceptance pending)

Implement `list_displays` and `screenshot` on the computer running the C# server.
Support one optional display index, defaulting to display zero, and return prepared PNG pixels
through the existing typed tool-result path. A browser tab, DOM reconstruction,
file path, or text claiming success does not satisfy this exercise.

Require the explicit host-screen grant and OS permission. Refuse safe mode,
ungranted calls, unsupported platforms, invalid selection, and unsupported model
image placement before their corresponding effects. Cancellation must join the
owned capture process; temporary output must not leak. Preserve structured image
results through delayed Jobs as well as immediate results.

Test refusal and cleanup with synthetic helpers first. Native success must use
an announced, agreed synthetic whole display, verify the actual pixels, and
exercise both a model turn and delayed-Job delivery. Test omitted selection and
explicit zero, another available display, invalid indices, and refusal of removed
rectangle arguments before capture. Inspect the advertised schema: only optional
`display` selects capture; observation controls do not select a region. Record
platform limits honestly; one display is not cross-display compositing. Do not trigger a permission prompt while the human is away,
elevate privilege, or introduce background monitoring.

The external grader must reject metadata-only capture, a missing grant,
capability checks performed after capture, a wrong display, mandatory selection,
advertised rectangle arguments, missing delayed
image payloads, and abandoned native work. The driver now passes synthetic wire
candidates and the agreed real native run passes **25/25**. This result is distinct
from the reference implementation's earlier synthetic-window and live-display
acceptance; those earlier observations did not silently award this grade.

Prepare a four-band fixture without opening or capturing any display. For example,
a 3840-by-2160 backing display prepares to 2048-by-1152:

```sh
go run ./cmd/host-screen-pattern -width 2048 -height 1152 ./screen-fixture
```

The command writes `display.html` and a strict `pattern.json`, refusing to overwrite
existing files. Open the HTML on the **server computer's primary display**, not
on a requesting phone. Its button enters full-screen only through a human click
and hides the cursor while full-screen.
Wait until browser/OS overlays disappear and confirm that the whole display is
synthetic. Prepared dimensions must match the current inventory; a mismatch
refuses before capture. A manifest is an expectation, not proof that the screen
currently contains the pattern.

Only after fresh human authorization and confirmation of that setup, run:

```sh
go run ./cmd/grade-edition3 -ch 26 -executable-only -media-scope host-screen \
  -allow-native-host-screen -host-pattern ./screen-fixture/pattern.json \
  -json /absolute/path/to/candidate-cli
```

The flag is an explicit operator attestation, not an OS-permission bypass. The
runner loads the grant, validates inventory, makes two primary-display screenshot
requests (omitted selector and explicit zero), collects their distinct running
Jobs, and checks both payloads on a new human turn without recapture. It refuses
without the opt-in; it never substitutes an unapproved primary display for an
approved secondary one. The current external scope uses one primary display.
Other-display selection and platform refusals remain separately attributed.

The pixel oracle checks every pixel within broad, predetermined band interiors:
inset each band by5% horizontally and the frame by5% vertically, with a minimum
two-pixel inset. Roughly81% of a large frame is checked, using the same three-level
RGBA tolerance. Exact dimensions and byte-identical full-image replay remain
required. Native testing exposed a darkened first row and a tiny corner indicator;
the screenshot faithfully includes such display details. Do not mistake an HTML
fixture's ideal colors for a guarantee that OS/browser edges contain nothing else.
Negative controls alter the first checked pixel in every band and must fail.

For diagnosis, `-host-failure-image /absolute/new/path.png` retains only the first
mismatching bounded PNG with private permissions. It is opt-in and refuses existing
paths; normal grading retains no images. Inspect actual failed pixels before
changing acceptance criteria, and report changes rather than quietly widening them.

The fake provider also distinguishes tool-free recall-judge requests by their
system instruction and answers them with an empty selection without advancing the
capture script. A longer prompt exposed this helper traffic before the main turn.
HTTP request order alone is not the Agent's tool sequence. Retain a long-prompt
schema-only regression that never enumerates or captures a display.


Synthetic controls reject early tool advertisement, rectangle/mandatory-selector
schemas, metadata-only output, missing or changed Jobs, bad pixels and lost replay.
Focused backend tests and review must still establish pre-effect capability and
permission checks and native cancellation: a returned error alone cannot prove
that no OS side effect occurred. The native test verifies the selected colors on
this display, not universal color-profile fidelity. Do not widen color tolerances
merely to obtain a passing grade.

## Grading boundary

All four executable external scopes pass, totaling100/100: image-files20,
attachments25, voice30 and host-screen25. The host scope has native macOS
primary-display evidence; the others retain their documented fake-provider,
synthetic audio and browser boundaries. Architecture review, other-platform
acceptance and native microphone/phone reliability are not additional claims
implied by that score. Building, staging, restarting and human acceptance remain
separate facts. Preserve the evidence and limits beside the executable contracts.


## Takeaways

- Prepare and retain actual media bytes at admission; replay must not depend on
  a source file or device still being available.
- Bind asynchronous work to its original owner and recipient. A new GUI
  selection cannot redirect an old recording or attachment.
- Separate capture, transcription readiness, actor admission, and model response.
  Each acknowledgement establishes only its own step.
- Preserve uncertainty after a potentially delivered send. Retrying can duplicate
  an action that already succeeded.
- Test the claimed surface: browser capture is not host capture, fake audio is
  not a native microphone, and a scoped grade is not chapter completion.

[Chapter 27: Crossing Over to Ensemble](chapter-27.md) turns to using the built
Agent as the working collaborator. The media paths here belong to that same
Agent and its existing ownership rules.
