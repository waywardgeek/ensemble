# Chapter 13: The Channel Nobody Tested

Bill listens to the coding agent for hours a day. In the first-edition investigation, that dependence produced a suspicion rather than a tidy bug report:

> I'm worried the TTS feedback, available via the MCP tunnel, isn't evaluated by any agent. It needs work.

Chapter 12's GUI observer could read a snapshot, complete a task, and never notice that the same information had failed to reach speech. Giving it another observation tool would help only if the tool exposed useful evidence and the task actually depended on that evidence. Otherwise the observer could keep succeeding by looking.

The [source chapter](../edition-3/chapter-13.md) takes the more revealing route: remove the DOM. A listener that can obtain the target's answer only through speech has a reason to notice silence. That does not make it a complete accessibility evaluator, or even a judge of how a voice sounds. It makes a neglected channel part of the task's critical path.

## Telemetry With Nothing to Say

The historical investigation found four paths meeting in the artifact renderer. Text and thinking deltas entered as they arrived. Tool dispatch announced itself. A completed part flushed the buffer. Errors rendered on screen.

The error path never called speech. It created a div, set its text, appended it, scrolled, and returned. The listener could miss precisely the message that explained why progress had stopped.

There was already a tool for inspecting speech: `tts_queue`. According to the source's account of the inspection, it read `window._ttsQueue`, but nothing wrote that variable. Its fallback returned an empty array or one synthetic entry containing `(speaking)`. The tool was marked `ephemeral: "round"`, so this contentless telemetry reached the model automatically on every round. The older GUI-skill check confirmed that injection worked. It had established reliable delivery of almost no information.

That is the historical mechanism, not the current four-tool GUI design from Chapter 12, where speech-queue inspection is on demand. The useful distinction survives either delivery policy: confirming that a result enters context does not establish that the result describes the channel.

Pause integration had a similar shape. The intended gate combined speech and human typing. Only typing was wired. The Go side checked pause before tool dispatch, but the browser never supplied the speech cause. Its speech module's opening comment advertised “Chrome wake-up and pause integration”; the reported inspection found no other occurrence of `pause` in the file. The comment had completed the integration before the code did.

These were not defects a daily listener could necessarily enumerate. Missing errors produce no bad utterance to examine. A missing pause may appear only as a tool starting too soon. An odd break in a sentence might come from segmentation, network timing, or the voice. Dependence on a channel makes its failures consequential; it does not give the user an independent record of what should have arrived.

## The Correct Answer From the Wrong Path

The first blind-persona experiment was expected to fail. It asked the target to run a command against a nonexistent path and report the exit code. Tool results did not enter automatic speech, so the prediction was that the listener would hear the dispatch announcement and lack the answer.

It reported the correct exit code.

The retained transcript explained why. Its six entries began with “run command,” followed by “The command failed with exit code 1:”, a lone backtick, the missing-path error, another lone backtick, and the start of an explanation. The first entry announced dispatch. Everything afterward was the model's prose about the result. The raw tool result had never been spoken.

The listener accurately reported what it received and concluded that no gap had been found. That conclusion exceeded the evidence. Words containing an exit code cannot, by themselves, establish which route delivered it. A talkative model had bridged a gap in the experiment's prediction.

Bill's scope decision matters here. The intended interface does **not** automatically read successful tool-result bodies aloud: “I listen to your thinking, and that is enough.” Thinking and response text carry the running commentary. A person can deliberately invoke an artifact's speaker button when its contents are wanted. Failed tools are different: the current dispatch path feeds their failures into error speech. Suppressing a successful file's contents must not also suppress the explanation that the file could not be read.

This changes what the listener is there to establish. Hearing the exit code in model prose is permitted behavior, not evidence that raw output was delivered. The question is whether thinking and response text arrive completely and in order, alongside identifiable announcements and errors. Source labels distinguish those paths for a program. Audible labels distinguish them for a person; the browser described in the source prefixes thinking and error segments with “Thinking:” and “Error:” before both recording and playback.

The policy has a tradeoff worth keeping visible. It avoids turning every file read into an unsolicited audiobook. It also means the listener's understanding of a successful tool result depends on what the model chooses to explain. That is the chosen commentary interface, not a promise to vocalize every byte available to the target. A listener that hunts for intentionally unspoken successful results is testing the wrong requirement.

The two isolated backticks, however, were nobody's intended interface.

## Preserve Structure Until It Has Done Its Job

Streaming was the obvious suspect. Perhaps a fence had arrived across several deltas, leaving the filter with pieces that could not match. The historical investigation fed the entire fenced block as one chunk. Backticks still reached speech.

That counterexample moved the fault upstream. The old pipeline split the buffer on newlines, then filtered each resulting phrase. A fenced block contains newlines. By the time the fence pattern ran, its opening marker, body, and closing marker had become separate inputs. No amount of improving the pattern would reunite them.

The repair resolves fenced blocks against the whole buffer before splitting it. A completed block becomes a spoken name such as “code block,” rather than a recital of its syntax. An unfinished fence stays buffered while its structure is unresolved. Chunk boundaries are transport accidents; they cannot decide which syntax exists.

Newlines needed a second distinction. A single newline can wrap a sentence without ending it. Treating every wrap as a phrase boundary adds a pause the author did not intend. A lone newline therefore becomes whitespace, while a blank line remains a boundary. Sentence-ending periods also let complete phrases leave the buffer. End of stream, a tool announcement, or an error forces a flush, so a final tail ending in a colon is not left waiting forever for punctuation that will never arrive.

Filtering removes emphasis markers, inline-code markers, heading and bullet prefixes, and link syntax before speech. The order matters beyond fences. An early review found that flattening a newline before removing bullets allowed the second item's prefix to escape a line-anchored pattern. A perfectly ordinary two-item list exposed the mistake. Structure should disappear only after every operation that needs it has finished.

Cleanup can destroy paragraph boundaries too. The current `gui/speech_text.go` described in the source uses horizontal whitespace, `[ \t]`, in prefix-removal patterns rather than `\s`. The broader class can consume blank lines. Without that distinction, a whole final could merge paragraphs that happened to stay separate during streaming. Identical text should not acquire different phrasing because the provider delivered it in a different number of pieces.

Identifiers get their own treatment: `camelCase`, `snake_case`, and `HTTPServer` expand into words, without chopping ordinary words or ordinary capitalization. That improves the synthesizer's input, but intelligibility still depends on the voice and listener. Expanding `HTTPServer` is a deterministic formatting decision; deciding whether its pronunciation is comfortable is not.

The buffer consequently has more responsibility than collecting characters until a timer expires. Deltas can arrive mid-word. The formatter must retain unfinished sentences, identifiers, links, and fences until a stable prefix can be emitted. Speaking each incoming fragment sacrifices meaning to network timing. Buffering forever preserves meaning for an audience that never hears it. Phrase boundaries and explicit flushes balance those two failures.

## A Final Is Not Necessarily a New Utterance

Another historical path displayed a complete part without speaking it. No deltas had preceded the final. The handler set the card's contents and called `flush()`, which dutifully emptied an already empty buffer.

Queuing every final would repair that silence and make streamed responses speak twice. The current design instead gives each part its own `speechText` formatter, with an offset into the **raw source text**. Filtering may remove many characters, so an offset into the cleaned text would lose its relationship to the next update.

For a delta, the formatter exposes only a stable prefix beyond the consumed offset. For a final, it exposes the remaining raw suffix. If no deltas arrived, that suffix is the whole part. After streaming, it is just the unsaid tail. A repeated final finds nothing new. Keeping state per part also prevents the unfinished tail of thinking from being joined to a response merely because the two arrived close together.

The source places this work in the GUI server. `gui/hub.go` feeds the formatters; `gui/speech.go` schedules segments with their source and part identity. The browser claims a speech task and admits its segments for playback. Replaying history reconstructs displayed cards without feeding old text into speech again. This is the current division of ownership, not the earlier browser formatter with its `card.streamed` guard.

That move changes where a useful final-only check must enter. In the historical implementation, turning provider streaming off reached the complete-part path. The current Engine can synthesize display deltas from a plain provider response. Disabling HTTP streaming therefore does not establish that the formatter received a final alone.

The previous edition caught this with a mutation: removing whole-final speech still passed the nonstream check. Suppressing deltas at the external WebSocket boundary then reached the browser-owned formatter. With formatting now in the GUI server, that boundary is too late. A meaningful check must deliver a complete part without preceding deltas to the actual formatter owner and observe the resulting speech.

This is also where a listener needs an expectation from outside the channel. It can notice that no reply arrived, but silence does not reveal which text was omitted, or whether the target produced text at all. Comparing the emitted final with the transcript distinguishes missing speech from an absent answer. The restricted listener and the independent comparison answer different questions; neither needs to pretend to be the other.

## The Transcript Begins Where Speech Begins

A transcript made from all queued text would look wonderfully complete even after cancellation discarded half the queue. It would describe what the application intended to say, not what entered the listener's channel.

The browser therefore records each filtered segment at the shared admission point, as that segment begins, before deciding whether native synthesis will play it. Transcript mode takes the same path and skips synthesis. It does not run a second formatter that happens to produce the expected fixture text while the GUI continues speaking broken fragments. Segments still waiting in the queue are not reported as heard; canceled future paragraphs cannot turn up later in the transcript as delivered speech.

This makes quiet operation useful without overstating it. The transcript is speech-channel evidence, not acoustic hearing. A record can carry the filtered text, its origin, and ordering; it cannot tell whether the speakers worked or a particular voice made an acronym intelligible. Native playback and transcript mode share admission, not the entire experience after admission.

The restriction on the listener is equally concrete. The accepted implementation described in the source is the separate public-API consumer at `examples/ch13-speech`, invoked by `tts-harness.sh`. It creates an Agent without builtins or skills and registers exactly `hear`, `type`, `submit`, `wait`, and `sleep`. The older Node virtual-user program is history, not an additional current component.

`hear` returns new speech-channel entries and advances a cursor. `type` appends to the actual composer through browser input events. `submit` sends the owned draft and remembers which target turn belongs to the session. `wait` observes that turn's outcome, distinguishing success from failure or interruption; it does not return the answer. The listener still has to hear it. `sleep` provides a cancellable delay rather than another route to target state. Waits have a bounded timeout, with a 120-second default, and sleeps remain subject to session cancellation.

Privately, the helper uses Chrome's page-target protocol to drive the real client. Publicly, the listener gets no selector, click, evaluation, snapshot, shell, file, or skill-lifecycle tool. The consuming Agent's declarations matter as much as the peer's advertised surface: attaching a narrow peer to an ordinary coding Agent would leave other ways to peek. Dispatch also rejects unsupported operations and arguments, including guessed names. Asking a capable model to avoid the DOM is cooperation; removing the capability is confinement.

The helper settles its owned target before reporting success. The listener finishing its own model turn is not evidence that the target succeeded. Likewise, its final report needs to agree with both the observed speech and the actual target outcome. The report is an interpretation of evidence, not a replacement for it.

## Pause Tracks Causes, Not Field Contents

Speech adds a second reason to hold automatic tool dispatch. A person may be listening to the current explanation, composing a correction, or doing both. For the local pacing policy, release follows only when neither cause remains.

Chapter 12's programmatic-input failure makes the representation important. The browser must remember human typing as a cause rather than repeatedly infer it from a nonempty field. MCP text does not become human text when an utterance ends. A genuinely entered space does count as composing, even though trimming the field would make it look empty.

The current browser described in the source centralizes the calculation in `updatePause(agent)`, with per-Agent typing and speech ownership and the Pause while speaking preference. Speech start, speech end, input, and submission update their causes through that calculation. It compares the derived value with the previous local value and sends a pause or unpause only on an edge. Extra keystrokes while already typing do not send more pauses.

Local causes govern automatic commands, not permanent ownership of the Agent's gate. An explicit unpause from another tab releases the Agent globally. An unchanged local cause in the first tab must not immediately reassert pause and veto that decision. This preserves the difference between automatic pacing and a person's deliberate override.

Escape cancels speech even when the composer contains text. It clears speech work without deleting the draft. Human typing may still hold the local cause, so stopping the voice and releasing the Agent are not synonymous. A check that sees only a typing pause and a later resume cannot establish this interaction. Speech ending first and typing ending first exercise different transitions.

Submission adds an ordering requirement. The browser clears the old typing hold **before** sending the new prompt. A late unpause can advance the continuation epoch after the prompt starts and invalidate its fresh speech. The stale cause must be released first; a delayed acknowledgement must also avoid clearing newer human typing. Otherwise the apparent speech defect is really a control message arriving in the wrong generation of work.

## The Listener Has a Lifetime Too

The earlier Node peer closed its browser on stdin EOF but left a pending sleep timer alive. The Go client closed the transport and waited for the peer. Its model turn could be finished while cleanup continued waiting for a timer whose session no longer existed.

The source reports the revealing comparison: idle EOF exited in 125 milliseconds, while EOF during a 4000-millisecond sleep took 3865 milliseconds. That was delayed cleanup, not a permanent hang. Sharing the existing close-browser-and-exit handler between EOF and SIGTERM brought the corresponding pending-sleep probe to 127 milliseconds, with a 126-millisecond idle control. Normal sleep still worked while the session remained open. The fix ended the process that owned the timer; it did not need a new cancellation framework or an arbitrary sleep limit.

The current Go helper gives interactive stdin its own watcher. EOF cancels the context used by the listener, pending `wait`, and pending `sleep`, even while one is busy. Cleanup releases its draft only if it is unchanged and interrupts only a matching unfinished target turn. In normal launch mode the helper also closes the browser and GUI it owns. Attach mode disconnects without closing the caller's processes. An ordinary one-shot invocation has no interactive stdin owner, so closed stdin does not cancel valid work.

Those ownership distinctions also shape attachment. The helper connects to the matching GUI page's debugger WebSocket, not the HTTP discovery endpoint or the browser-wide socket, and validates the expected GUI scheme, host, and path. An isolated save and browser profile keep an observation session away from someone's live work. The helper submits the required initial target prompt before starting the listener; the listener's goal is a separate instruction, not a substitute for starting that turn.

This remains browser-driving work in a small public consumer, outside the core. The GUI is itself a separate public consumer. Neither needs to absorb the listener's orchestration merely because the listener exercises its behavior.

## Evidence at the Boundary That Can Fail

The source's automated application checks use an external provider and asynchronous synthesizer fake while retaining the real Agent, provider parser, renderer, and speech path. Observed utterance records carry text and source with sequence and timing; gate records describe actual pause and resume events and their causes. Constructing an expected log independently of the page would prove the log generator. It would leave the application unexamined.

Even an absence assertion needs a successful event behind it. The old file-suppression fixture requested `notes.txt`, but the harness launched the GUI in a different directory from the prepared workspace. The file read failed. Silence could not establish that successful file contents stayed out of speech, because those contents had never arrived. Using the planted file's absolute path and checking the correlated successful result supplied something real to exclude. The model's distinct answer could then be compared with the transcript without confusing it with the file's bytes.

The historical speech checks passed before the blind client and cleanup correction were finished; the stronger file check later exposed a gap even in that passing speech result. Earlier client and command-lifetime behavior still needs separate compatibility evidence. Neither a successful speech score nor a listener process exiting zero establishes the whole system.

Native browser callbacks add evidence that the actual synthesizer started and ended utterances. The source carries Chapter 12's controls for whitespace, programmatic input, explicit other-tab unpause, and a subsequent real submission into this chapter. The earlier native check closed its extra control tab before the single-tab speech sequence; it did not resolve the unexplained two-tab queue stall. A quiet transcript check cannot fill that gap.

The source's three historical listener sessions illustrate the distinction between channel success and target-content success. All three used the restricted tools to type, submit, and receive real target transcripts. One target never chose the requested fruit, and its listener reported the omission; another supplied an apple in three sentences rather than the requested two; the third delivered thinking and response entries before the listener reported the apple. These were useful channel observations, not universal instruction following. No human listened to those runs. Separate native-callback checks and later deterministic cleanup evidence were not final-source repeats of the listener sessions.

The remaining human judgment is not a ceremonial final approval. Pronunciation, paragraph pauses, rate, and fatigue determine whether the channel is worth using. A browser accepting a rate value does not measure words per minute. The source retains no measurement supporting its older fixed browser-speed ceiling or dedicated-engine comparison, so neither belongs in the design argument.

The listener tests dependence on speech; the transcript exposes delivered text and origin; independent target evidence exposes omissions; native callbacks expose playback transitions. A person still supplies the ears. Keyboard and screen-reader navigation remain separate accessibility concerns, and different blind and low-vision users need not want the same interface. Keeping this instrument narrow makes its findings useful without asking it to certify an experience it cannot have.
