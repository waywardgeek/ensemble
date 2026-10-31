# Chapter 7: Everything Is an Artifact

Chapter 6 gets the words onto the terminal before the answer is finished. That is a considerable improvement over staring at nothing. It is less helpful when the thing worth noticing has already scrolled away, the tool arguments wrap across six lines, and the human is trying to compose a correction while the next tool starts.

The GUI has a harder job than making that terminal attractive. It must give the human a usable view of ongoing work, recover that view after a disconnect, and leave enough room to intervene. A display that announces an execution which never happened is misleading. A pause button that prevents interruption is worse than no pause button. Speech that continues reading yesterday’s history is merely thorough in the wrong direction.

The [source chapter](../edition-3/chapter-07.md) builds one reusable scroll component, not the whole workbench. Its central choice is to keep the Agent unaware of the GUI. Observation and control cross public interfaces; layout and sockets stay outside. The interesting work lies in making that separation hold when the browser is slow, absent, speaking, reconnecting, or wrong.

## One Agent, Several Receivers

The source’s radio analogy is useful: a station does not stop transmitting when somebody unplugs a receiver. A browser subscribes to observations from the existing Agent. A second browser subscribes to the same Agent. Closing either tab changes neither the conversation nor the running work.

Unlike a radio, the browser also gets a recent recording and can send input back. The recording is durable History, not an accumulation of whatever notifications a particular connection happened to catch. The return path is the same public `Post` entrance the terminal uses. A message arriving while the Agent is busy becomes a hint under the actor’s existing rules; the browser does not choose a turn for it or start another model loop.

The application constructs one Ensemble and one Agent for terminal and browser alike. The GUI and WebSocket server live in an optional Go module using the public library. Core acquires no WebSocket dependency, rendering package, or display configuration. Configuration remains with Agent, Engine reaches it through its actual parent, and the transport owns sockets. There is no need to make a web server the Agent’s parent merely because it knows where a button is.

That boundary has operational consequences. The accepted server listens on localhost, serves the page and `/ws`, and takes its asset location separately from the tool workspace. Its public flags override inherited provider environment defaults; credentials remain in the intended process environment, not command arguments or logs. The working directory supplies the default workspace and log location. Terminal EOF detaches that input rather than shutting down the browser. These details preserve one running application with several entrances, rather than letting each entrance quietly invent its own lifecycle.

## A Requested Call Is Not a Started Tool

Streaming already shows the model composing a tool call. The GUI also needs to distinguish that request from execution, and execution from a returned observation of the job.

`ToolDispatched` supplies the call ID, tool name, and raw arguments at the execution boundary. `ToolFinished` supplies the call ID, bounded result text, and error flag when the tool wait returns. That result may contain a handle for a job still running. The full output remains recoverable, and later real job reports must remain visible even if the human has interrupted the turn. “Finished” describes the wait reported here, not a universal claim that the process is dead.

`TurnEnded` gives displays the final reply, excluding intermediate tool narration, or a safe diagnostic when the turn fails or is interrupted. It is progress, not the reliable completion used by `Ask`. A missing display notification cannot be allowed to redefine whether a request completed. Ensemble identifies the source Agent when broadcasting these observations; tools do not need a copied Agent identity or a GUI parent.

The source records a revealing Edition 3 defect: the GUI announced tool starts for a batch the core had refused. It did so live and again on replay. The requested calls existed, and refusal results existed, so the display inferred the missing action between them. It had produced a convincing account of an execution that never occurred.

The repair belongs at admission, not in a more persuasive card label. Saving and publishing the dispatch fact happen inside the same boundary that releases the handler. A refused call remains visible as a request with its refusal result, but receives no dispatch announcement. Cancellation after publication can stop admitted work; publication must not describe a reservation that was never released.

This is why the three vocabularies stay distinct. Stored events such as `tool_called` and `tool_returned` are historical records. Mailbox reports are internal signals. Observations are broadcasts. Similar names and related content do not make them interchangeable evidence. In particular, the existence of a stored request is not proof that its handler ran.

## Pause Buys the Next Boundary

The actor dispatches tools serially: start one, wait for its report, then consider the next. A continuing job can outlive that report, so this is not a promise that only one operating-system process exists at a time. It is a promise of a dispatch boundary between calls.

That boundary is what makes pause useful. Starting a whole batch in parallel would leave a pause arriving just afterward with nothing left to prevent. Serial dispatch gives the human a place to catch the next call. An engineer considering parallelization needs to see that dependency, not just a loop that appears insufficiently ambitious.

Pause is live Agent control, not another conversation event awaiting reduction. The Agent owns the gate; the browser reaches its public `Pause` and `Unpause` methods through the server. There is no second GUI gate and no Engine sibling handed to the transport.

The accepted implementation uses a mutex and a wake channel. When the gate is closed, a small boundary waiter waits for unpause or cancellation and reports back to the actor. The actor continues consuming its mailbox. Blocking the actor itself would strand the interrupt intended to rescue the human from the pause.

The two wake reasons mean different things. Unpause permits the next dispatch. Cancellation ends the current turn, leaving the Agent usable for another prompt and other jobs free to report their actual outcomes. The gate deliberately bypasses mailbox FIFO order because a pause waiting behind pending reports may arrive too late to hold the next start.

None of this retracts an admission already made. A running tool continues producing output while paused. Pause is neither process suspension nor retroactive cancellation. Its power is narrower and easier to reason about: no new tool starts past a closed boundary.

## Reconnecting Without Reconstructing the Conversation

A fresh browser and a reconnecting browser use the same subscription path. Neither supplies a cursor or sequence number. The server sends the recent durable history, then accumulated content for active parts, then live delivery. The browser receives identities and renderable payloads, not provider SSE frames to interpret for itself.

The history window is fixed: the last hour or the last 100 renderable events, whichever reaches farther back. A busy hour can contain more than 100 events; a quiet Agent can supply older context to reach 100. An event producing several rendered parts still counts as one event. Counting socket messages or unique cards instead would silently change the policy.

The subscription also reports current turn and pause state, marked as replay. The turn snapshot uses `state_changed` with `from` and `to` both set to the current state. Equal endpoints say “this is where things stand,” without inventing a transition. The initial subscription snapshot omitted `from`; the source records the corrected subscription check and a restored-defect mutation that detects the omission. Likewise, an empty delta still carries `chunk` with an empty string. Empty content does not make a public field optional.

The hub keeps two tiers of state. Durable events supply finalized parts, user messages, calls, and results. An in-flight map accumulates the chunks of parts still streaming. Live delivery sends only the new chunk; subscription sends the prefix accumulated so far. Finalization removes that prefix, because History now supplies the authoritative content.

There is an internal frontier even though there is no browser cursor. On each observation the hub checks for new history since its captured event-log `Seq`, projects the renderable facts, and combines them with progress. The event log supplies order. A connection does not get to claim which durable facts the server has already captured.

Chapter 6’s parser IDs are too short-lived for this job. They distinguish parts inside one model response; a GUI card can survive several responses and human turns. The actor therefore assigns a conversation-scoped display ID, broadcasts it through the root observer, and stores it on the final part. Replay updates the card that received the live chunks. Together, `agent_id` and `part_id` identify the logical part, even when several events update its dispatch status or job metadata.

This identity is for display, not a new interpretation of provider call IDs or signed content. Those retain their original meanings. Nor does writing IDs to the log implement process restoration. When save/load arrives later, new parts after restart must avoid colliding with old cards; this chapter recovers a browser view while the Agent continues running.

The awkward case is not a reconnect after everything is quiet. It is a reconnect exactly as a part finishes. Without coordination, the subscriber can receive completed history and an obsolete partial for the same content, or miss the final between two snapshots.

Append-only history does not solve that race. Appending can replace a Go slice’s backing array while another goroutine reads it. The history owner snapshots the requested suffix under its lock. The GUI then coordinates the captured history frontier, partial map, and subscriber admission under one handoff lock. A captured final removes its partial before publication. Active partials replay in assigned creation-identity order rather than Go map iteration order. Interruption clears abandoned partials that will never finalize, without deleting completed facts or late real job output.

The source’s testing history shows why the timing matters. Removing final-part cleanup initially passed: the test reconnected after `TurnEnded`, whose separate cleanup had already emptied the map. Reconnecting while tools were still running exposed the stale partial. Both tests reconnected successfully. Only one visited the interval in which the defect existed.

## A Slow Browser Does Not Own the Actor

The observer callback captures history and updates partials, then offers messages to clients without waiting for network I/O. Each connection has a buffered channel and a writer goroutine. A full live channel drops that client’s progress update rather than blocking the Agent.

This is a different promise from Chapter 6’s terminal queue, which preserves reply bytes. A browser can miss live progress and recover durable content through subscription. The choice protects the actor from a slow receiver; it does not guarantee every receiver sees every intermediate state.

That bargain cannot apply indiscriminately to replay. Dropping half the reconnect history would break the recovery path offered in exchange for lossy progress. The accepted server admits the complete replay as one queued batch while adding the subscriber. The connection writer sends the batch outside the lock, so history admission is coordinated without turning the handoff lock into a network lock.

A retained test used 270 replay events, more than the live queue’s 256-batch capacity. A truncation mutation initially produced a timeout, which did not establish the cause. After the test preserved terminal status, the failure became specific: 249 facts arrived instead of 270. The useful number is the missing content, not merely the fact that a test went red.

The writer also owns WebSocket control replies. A Ping requires a Pong control frame echoing its payload; an application JSON message named `pong` is a different thing. Network sends and outgoing trace writes stay with the writer, incoming traces with the reader. Neither belongs in the observation callback.

Disconnect uses a separate lifetime signal. Closing a data channel that a broadcaster can still reach is unsafe even if the broadcaster uses nonblocking sends. Shutdown joins writers before closing their logs. These are modest connection-lifetime rules, but violating them would let an optional receiver crash or delay the work it was meant to observe.

The optional transport adds `gui.log` alongside the event, API, and diagnostic logs. Each line contains an RFC3339 timestamp, a direction marker, and one compact JSON payload; the entire line is not itself a JSON object. Both directions matter. Seeing the server’s output without the browser’s command leaves a different debugging hole from seeing the command without what the server sent back. General diagnostics still belong to Ensemble’s logger; this is the transport’s account of its own wire.

## One Artifact, Several Ways to Lie

`ArtifactScroll` manages a list of Artifacts: thinking, chat, tool arguments, tool results, and user messages. An Artifact can accumulate content and then finalize; user messages and tool results need no invented streaming phase. Renderers handle markdown, JSON, or ANSI as appropriate, and CSS supplies differences such as dim thinking or compact tools. The component owns identity, updating, and finalization once rather than teaching a separate panel each version of those rules.

A delta finds or creates the part’s card and appends through its renderer. A final replaces the provisional view with the authoritative rendering. Dispatch and finish observations give tool cards their execution state. The abstraction is useful because it shares a lifecycle, not because every artifact contains the same kind of text.

The first-edition implementation broke two basic display rules until its later audit. It built card content by interpolating values into `innerHTML`, and it capped displayed values without keeping the rest reachable.

A tool’s name or arguments can contain filenames, URLs, and file contents influenced by outside input. Interpolating those strings as HTML moves untrusted markup into the page. The card shell must use `textContent` for such values rather than treating them as markup. Choosing a markdown renderer does not grant permission to assemble the surrounding card from raw HTML interpolation. The security boundary reaches the human through the renderer, without requiring the model to obey a malicious instruction first.

Similarly, a shortened display must leave the full value available through a title or expander. An ellipsis without an affordance is not compact presentation; it has made the omitted content inaccessible. That becomes especially consequential when the hidden portion is the error needed to understand a tool result.

A later live Gemini page exposed a subtler failure. A legitimate empty text part displayed its own metadata. The renderer tested whether the text was truthy, then fell back to JSON. The empty string lost, and the card filled the vacancy with information that was never intended as prose.

The same fallback would expose opaque signatures and redacted thinking. Because automatic speech and the speaker button also used the card’s text, the error travelled from the screen into both speech paths. Correct execution underneath did not make the page correct.

The repair is an explicit projection of readable content. Empty text stays empty. Readable Anthropic thinking and Gemini thought text can become prose; signatures and redacted bodies remain opaque replay material. Their bytes belong in History without being promoted into something to display or read aloud.

The source records a focused browser check using the retained Gemini part plus readable and opaque controls from an external peer. It examined the card, automatic-speech input, and speaker-button input separately. Restoring the fallback broke all three paths for empty or opaque cases while readable controls still produced native speech events. The saved part was enough to reproduce the bug. Another paid generation would have added randomness, not understanding.

## Speech Is Also Flow Control

Automatic speech reads thinking and chat in full, abbreviates tool dispatches, and leaves tool results silent. The speaker on each card has a different purpose: it reads the full accessible text on demand. The Edition 3 page once displayed a long tool result but spoke only its short generated caption. Reusing the automatic summary for the speaker button had erased the distinction between a notification and a request to hear the content.

Network chunks are not phrases. The speech queue accumulates unfinished phrases by artifact identity, queues speakable portions, and flushes the remainder on finalization. A final must not read the entire answer again. Replayed history remains silent until the human requests it.

Speech also controls the tool gate. A tab’s local reason to pause is queued or active speech, or user typing. Pause starts when an utterance is queued, not when audio finally begins; otherwise a tool can start while the browser is still waiting to speak its explanation. When the queue empties, the tab can unpause unless input still contains text. Sending the input reevaluates that condition. Escape with an empty input cancels speech and reevaluates it too. Typing must work as a pause cause even with automatic speech switched off.

This gives the speech queue an operational responsibility. If it gets stuck, tool starts can remain paused indefinitely. The source’s Chrome behavior makes `onerror` as important as `onend`: cancellation can produce an `interrupted` error as ordinary queue behavior. Both callbacks must finish the current utterance and advance. The page keeps a reference to the active utterance and ignores an old cancellation callback after a newer utterance has taken its place.

The original Chrome work also found the speech service asleep after tab switches. Its recorded workaround is a zero-volume empty utterance on `visibilitychange`. The accepted page keeps that wakeup outside the audible queue, so a compatibility nudge does not acquire the pause gate. This is a recorded browser behavior and workaround, not a claim that every browser version needs the same treatment.

Multiple tabs introduce a more interesting policy than simply OR-ing every browser’s desires forever. Pause is Agent-wide. An explicit **Unpause everyone** releases the one gate even if another tab still contains unsent text or queued speech. That other tab’s unchanged local cause cannot veto the decision.

Automatic pause commands therefore follow changes in the combined local cause, not every utterance boundary. Repeatedly asserting pause while the cause remains true would undo another tab’s explicit unpause. Reconnection must not reassert stale local pause state either. The server owns effective Agent state; each tab contributes transitions, not a permanent lease on the gate.

The policy trades unanimous consent for a usable override. Another tab’s genuine new transition can affect the gate, but its old condition cannot keep winning by repetition. That distinction matters more than whether the button looks reassuring.

## What the Browser Evidence Establishes

A WebSocket test can establish message fields and replay content. It cannot establish safe DOM construction, the text supplied to speech, or whether cancellation advances a native speech queue. The source keeps those checks separate rather than treating transport success as proof of a usable interface.

Its retained browser work covered two tabs, connection during an active partial, completed-history reload, safe markup, reachable long content, typing with speech disabled, cross-tab unpause, speaker clicks, Escape, and visibility changes. The [historical browser receipt](../../solutions/evidence/ch08/browser.json) records native speech events. Those establish API activity, not that a person heard intelligible audio.

The speech work did not pass cleanly at first. One multi-tab record included a native `canceled` error. A later repeat waited for speech that never completed: reloading the second tab had restored its checked Read aloud default and accidentally enabled overlapping speech during the long response. Explicitly disabling voice in that tab after reload produced the clean automatic-speech result. Reloading had changed the experiment’s conditions, not merely its appearance.

A separate [interruption record](../../solutions/evidence/ch08/speech.json) exercised a real `speechSynthesis.cancel()`: the first utterance started, cancellation occurred, and the second started and ended with pause released. Restoring the defective error handler stranded the queue. Listeners had to be attached before playback so a fast start could not escape observation. This demonstrates queue recovery in the native browser; listening remains a separate judgment.

The source’s [live summary](../../solutions/evidence/ch08/live-summary.json) records historical Anthropic, OpenAI, and required Gemini 3.8 turns that ran `run_command` and returned their markers successfully. The Anthropic run also delivered a terminal busy hint to the browser on the same Agent. Gemini’s successful tool run was the one that revealed the empty-text metadata card. The later [projection correction](../../solutions/evidence/ch08/projection.json) repairs that display defect; successful tool execution had never proved the rendering right.

These are retained source results, not new executions for this companion or a current provider-availability list. Their value is the distinction they expose: a system can finish the requested work, show the wrong thing, and leave the human unable to steer the next action.

The chapter leaves one scroll pane and one input field, with a reusable unit for both streaming and completed content. The next chapter arranges those units into a workbench with separate tool views, an agent tree, and workspace controls. The layout can grow because the harder agreements are already in place: the Agent owns the work, History owns the facts, and a browser can come and go without becoming either.
