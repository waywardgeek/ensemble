# Chapter 26: Voice and Images

> **Opening draft.** This chapter currently contains only the TL;DR and the idea
> in plain words. The new interaction paths have not been implemented or graded.
> Working code comes next, then exercises and graders, passing results, and the
> full chapter body. Earlier media support is a foundation, not evidence that
> these GUI features already work.

## TL;DR — Specification Brief

Extend the working GUI with push-to-talk input, pasted images, screenshots of the
connected Ensemble page, and Agent inspection of image files. Build on the media
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
  cancels the pending operation rather than sending partial speech later. Late
  callbacks must not submit it or redirect it to a newly selected Agent.
- **Paste images into chat.** Accept image content from the clipboard, show a
  preview with removal controls, and send it with the accompanying text when the
  user presses Send. Preserve their order and the intended recipient. A preview
  is not delivery to the Agent. Validate format, size, and image dimensions before
  accepting the attachment; refusals leave the user's remaining draft intact.
- **Capture the connected page.** Provide an Agent-callable screenshot operation
  for the connected Ensemble page, with explicit capture authority and a truthful
  result. The result must make actual image pixels available to the Agent, not
  merely return a filename or a description of page elements. Other windows,
  tabs, and the desktop are outside this initial scope and require a separate
  permission decision. There is no background screen monitoring.
- **Inspect an image file.** Provide an image-reading operation for an explicit
  path, using the Agent's existing Sandbox and supported media representation.
  Enforce file authority, format, and resource limits; report missing, forbidden,
  or unsupported input honestly. The operation must deliver the image to the
  selected model, rather than treat a path, browser-only preview, or textual
  metadata as proof that the image was seen.
- **Preserve the existing owners.** The browser owns capture and composition;
  the Agent owns admission and History; the Engine renders supported media for
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

A person can already type to Ensemble and listen to its replies. That is useful
at a desk, but a phone makes the missing inputs obvious. Describing a picture in
text takes work that the picture could do directly. Typing a message occupies
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

Screenshots and image files meet at that same path. One obtains pixels from an
authorized page capture; the other reads a named file under the existing file
policy. Neither needs its own history or permission system. A structured page
snapshot can help find controls, but it is not a screenshot and cannot establish
what those controls look like. Conversely, permission to capture the Ensemble
page does not grant access to everything else on the screen.

The important boundary is not whether the interface has a microphone or camera
symbol. It is whether the person's intended input reaches the intended Agent,
with permission, once, and with a clear failure when it cannot. That boundary is
what lets the same interaction remain trustworthy on a desktop or a phone.
