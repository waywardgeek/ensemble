# Chapter 5: Two Seams and a Loop

The test is running in the wrong directory. The user types a correction. The application accepts the text, and nothing happens.

Background execution has solved only half the problem. A tool can run independently while the conversation loop remains parked waiting for its report. The input is present, but the owner that must interpret it is asleep.

The repair gives conversation state one owner that stays available. Transport and tool workers bring facts back to that owner instead of borrowing its history and changing it themselves.

## One owner orders the conversation

Each Agent has an actor: a loop with a mailbox and the authority to advance that Agent's conversation. Prompts, hints, control requests, transport outcomes, and tool reports enter through the mailbox. The actor applies their facts, records the resulting history, and decides what may run next.

This is not a second framework hidden inside the application. The C# implementation uses async operations to perform waits outside the conversation owner. Its lasting scheduling fields are declared with the Agent in Common. A worker does not acquire permission to edit History merely because it has an object reference capable of reaching it.

That distinction is more useful than counting threads. The application may need transport workers, process readers, tool handlers, and observers. The important count is the number of authorities ordering one conversation: one.

Admission order also needs a single meaning. Separate independently selected channels for prompts, hints, and completions could invent a new ordering after the facts had already arrived. The mailbox provides an order in which this Agent accepts work. It does not pretend to discover a universal chronology among simultaneous events elsewhere.

## Freeze before sending

To start a model request, the actor records the input, renders a frozen request, records its request boundary, and releases the HTTP operation. At most one model request is active for that Agent.

The transport receives the frozen request, not a live History object from which it can assemble a slightly later conversation. It returns immutable response facts without assigning their final place in the log. The actor supplies that order when it accepts the result.

A hint arriving after the request was frozen cannot retroactively change its bytes. It belongs in the next request. Recording that boundary makes the distinction reconstructible rather than dependent on when a background thread happened to read a mutable collection.

Tool execution has a similar admission prefix. The actor validates call identity and records the call before a handler is released. The subsequent wait for a report runs elsewhere and posts its outcome back. Long observation does not hold the mailbox hostage.

A report is recorded before the next model request consumes it. A directly invoked host tool has a real audit outcome too, but should not invent an assistant tool call that the model never made. Both cases preserve evidence; they have different authorship.

## Two public seams

An application needs to send work into the Agent and observe what comes out. Those are separate operations with different reliability requirements.

The inbound seam admits prompts and controls. A submitted request receives its own completion task. Concurrent submissions queue separate turns rather than racing to read a shared last-answer field. The actor settles each admitted request exactly once, and asynchronous continuations prevent an application callback from taking over the actor while the result is delivered.

The outbound seam carries progress through bounded consuming channels, one on the Agent and a separate merged channel on Ensemble. Two readers of the same channel compete for values; this is not broadcast subscription. A consumer may also discard nonmatching values while waiting for a particular observation. Later GUI fan-out needs one reader and a separate projection for its viewers. Observation can lose progress when a consumer is slow. Letting a browser stall the actor because its output queue is full would make presentation part of the application's correctness path.

The reliable completion task therefore does not depend on seeing every progress event, noticing an arbitrary Idle transition, or guessing that the last text chunk was final. A caller waiting for an answer and a window watching tokens are not doing the same job.

Cancellation follows that distinction. A caller can detach from its wait without cancelling the admitted turn. Ending the turn requires an explicit interrupt. Closing a tab should not silently become authority to destroy work that another consumer still cares about.

With durable logging enabled, recording precedes publication and completion. An answer cannot be reported as durably successful if the append that establishes it failed.

## Make the library usable without its front end

An earlier architecture exposed the cost of letting a framework import its own GUI. The supposed library depended on the application that was meant to consume it. A new front end could not simply borrow the agent; it inherited composition decisions it did not own.

Here, the application constructs the root, configuration, and Agents. A CLI translates terminal input into submissions and controls. A GUI does the same for browser messages. Neither owns a private copy of the conversational loop.

This separation matters before there are multiple user interfaces. It keeps stdin, window state, and application exit out of the provider and history code. It also gives tests and small external programs a real public seam instead of a privileged back door available only inside the repository.

Input capability belongs to the selected model and provider, not to what the front end can display. Successful download or decoding does not establish that a particular model accepts an image, audio, or video input. The early conservative defaults admit images and refuse other media unless configured support says otherwise. Unsupported input is refused before its history event is recorded. A browser's ability to play a video is a separate presentation fact.

## Hear the correction

A hint is a useful test of the whole design because receipt alone proves very little.

During HTTP, the actor retains the hint for the next request. During foreground tool observation, it ends that observation without killing the independent Job, obtains the actual running or completed report, and returns control to the model with the hint visible.

A response may have requested several tools. If a hint or control boundary prevents the undispatched tail from running, those calls still need exact-ID results saying they were not executed. Leaving them dangling would corrupt the next provider request. Executing them later after the model has changed its plan would be worse.

An idle hint does not necessarily start a fresh turn. It can remain available until an actual trigger arrives. That prevents a seemingly small control message from becoming an unrequested paid generation.

Pause is different again. It prevents new model and tool scheduling while the actor continues to receive input and process completions. Resume allows scheduling to continue. Freezing the mailbox would make pause unable to hear its own release.

Immediate interrupt cancels the active transport and foreground observation, then ends the current turn. It does not automatically kill independent Jobs. An after-tool interrupt preserves the already admitted response or dispatched-tool boundary and refuses subsequent effects. In both cases, late outcomes must be collected without letting them dispatch stale calls or settle a newer request.

These distinctions take more care than one global cancellation token, but they correspond to operations the user actually means: stop showing me this, stop this turn, stop starting new work, or terminate that process.

## Bound the loop without mutilating the answer

A tool-round limit bounds how much work one human turn can generate. It counts batches, not individual calls. After the last permitted batch, the model still gets the results and can answer. If it asks for another batch, the application refuses those effects and pairs the calls with explicit results.

Editing the model's returned content to conceal its extra calls would destroy the record and can invalidate signed content. The limit belongs at the execution boundary, where the application has authority to refuse an effect, not at a fictional rewrite of what the model said.

The same principle governs shutdown. Close admission, settle already admitted requests, cancel and join the transport and observations, stop owned Jobs, and only then detach the model. History and parent paths remain available during cleanup. Repeated shutdown joins the same lifetime rather than starting a competing cleanup operation.

A managed callback that ignores cancellation can still delay completion. The actor design improves ownership and responsiveness; it does not give .NET a safe arbitrary thread-abort operation.

The result is a library whose conversation can be watched, addressed, paused, and stopped without confusing those actions. With that separation established, streaming can finally be treated as what it is: a better way to observe an answer arriving.

*Implementation details: [codebook Chapter 5](../chapter-05.md).*
