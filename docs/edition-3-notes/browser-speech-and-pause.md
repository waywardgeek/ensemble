# Browser-local speech and Agent-wide pause

## Approved contract and source implementation

Bill requested phone/laptop-independent TTS preferences, like browser-local theme.
He then explicitly chose one Agent-wide pause state: pause pauses, resume resumes,
regardless of the sending device. This supersedes connection-owned automatic holds.
No device registry, competing vetoes, or separate manual/automatic flags remain.

Narration enabled, speed, voice URI and speech-pausing use origin-local
`ensemble.speech.v1`. Same-origin tabs synchronize; profiles/devices are independent.
Fresh, removed or invalid preferences default to narration off. Do not seed enabled
from legacy server settings: unattended devices stay quiet. Storage failure falls
back to this page; controls work offline. Legacy server fields remain for old
clients/files only. Current controls never patch or apply those TTS fields.
Microphone transcription/funding/shortcut remain separate.

Typing OR enabled speech-pausing with current/queued audio creates a local activity
edge. Only changes send ordinary pause/unpause. Stop speech and Escape cancel local
audio and explicitly resume the selected Agent, preserving draft text and its edge
baseline. Unchanged activity cannot reassert an old pause. Normal completion sends
resume when local activity clears. Disconnect does not invent resume; reconnect
does not reassert stale activity. Saved children retain the single pause state.

Commands address only the tab's selected Agent, checked by the server. Selection
cancels outgoing speech and sends outgoing resume before changing target. Late
callbacks cannot affect the incoming Agent. Background tabs may still speak for
their selection when enabled; browser-local off prevents automatic narration.

Chapter7 teaches this where TTS and steering are built. Chapter8 distinguishes local
preferences from its retained legacy server schema. Chapter27 requires the matching
cross-device, selection, storage and lifecycle acceptance checks.

## Verification

- Clean Tests, GuiTests and standalone GUI builds, zero warnings/errors.
- Focused pause/model3 and child-restoration2 groups passed.
- Final shared suite345/345 and all five public GUI groups passed.
- Actual embedded GUI, real sockets and Agent with controlled native speech:
  Stop/Escape from another device, unchanged activity, preserved drafts,
  disconnect/reconnect, selected-Agent isolation and forged-target refusal.
- Local speech: reload, same-profile tabs, independent profiles, server values
  ignored, offline changes, removal, blocked storage and zero backend TTS writes.
- Speech lifecycle: queue, late callbacks, native liveness/startup grace, dialogs,
  silent replay and Agent-switch ordering. Real fake-provider turn completion:
  normal, missing-callback and disconnect modes all inspected actual actor state.
  Disconnect retains pause until explicit Stop, not automatic release.
- Chapters7/8/27 hard lint passed with style advisories. No paid model, native
  microphone, capture, installation, restart or human hearing acceptance here.

The old Thinking DOM test required the last character to be the end marker; the
full Markdown renderer adds trailing structural whitespace. Retaining the length
and marker checks with trimEnd passed without changing renderer behavior.
An initial completion-mode loop passed positional arguments that the fixture does
not consume, so all three runs were normal. Missing-callback and disconnect were
rerun using SPEECH_COMPLETION and passed. Those initial runs do not count as mode
coverage. Review also corrected event publication to follow pause-state mutation,
and limited audio cancellation to enabled-to-disabled, not arbitrary local edits.

Evidence: /private/tmp/cadence-local-speech-shared.log,
/private/tmp/cadence-local-speech-gui.log, and workspace pointers
/private/tmp/cadence-local-speech-final-latest and
/private/tmp/cadence-local-speech-extra-latest. Fixtures stopped and joined.

Source-only. After coordinated installation/restart, enable speech on the phone
and leave it off on the laptop; verify actual listening and cross-device resume.
