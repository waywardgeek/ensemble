# Chapter 5: Two Seams and a Loop

The test is still running. You type, “Wait, wrong directory.” The application accepts the line. Nothing happens.

Chapter 4 made the work survivable: the tool has a job handle, its output is retained, and the end of a wait does not kill it. But the conversation loop can still be parked waiting for that job’s report. The goroutine that ought to hear you is the goroutine taking a nap. Moving execution off the foreground was necessary; it did not move *observation* off the foreground.

The [source chapter](../edition-3/chapter-05.md) pinpoints the blockage: the tool already runs on its own goroutine, but the conversation loop parks in a `select` awaiting its report. That same loop must drain inbound events. Hints can reach the channel and still sit unread because the drainer is parked.

The repair is not a shorter timeout. It is a change of ownership. The Agent owns a conversation loop that stays available for incoming facts. HTTP and job observations happen elsewhere and return through the same entrance as human input. The application can watch that conversation without borrowing its steering wheel.

## One Owner, Several Messengers

The tempting design is to let each source of activity do a little conversation management. The stdin reader can append a hint. The HTTP worker can append a response. The tool worker can update context when it finishes. The GUI can inspect the current turn and decide what the next line means.

Now everyone has a small, reasonable responsibility and a slightly different idea of the present.

The actor design makes one goroutine responsible for the Agent’s conversation state. Its mailbox accepts prompts, hints, interrupts, request submissions, and private worker outcomes. The actor drains those messages, records and reduces events, decides what happens next, and notifies observers. Workers bring facts back; they do not append directly to live History or mutate the Engine’s context behind its back.

The queue need not be exotic. The source uses a mutex-protected slice plus a buffered wake channel with room for one signal. Posting appends under the mutex and attempts to signal. If a signal is already pending, another is unnecessary: the queued messages, not the channel’s contents, are the work. Draining takes the accumulated slice and clears it.

That distinction buys a simple, useful guarantee. A producer need not wait for the actor to become a receiver. This is not lock-free—enqueueing briefly takes the mutex—but it avoids turning a busy actor into a blocked stdin reader. Nor does a full wake channel mean a message may be dropped. The queue has already retained it. Reliable admission matters particularly for requests: a caller cannot receive a reply to a request the actor never got.

One FIFO gives the owner one order to process. Separate channels for hints and completions would let a `select` choose arbitrarily when both were ready. That would introduce a new scheduling decision after admission, precisely where the application wants to preserve which fact arrived first. The queue does not discover a universal order among concurrent events; it preserves the order in which this Agent accepts them.

The actor must also stay out of the HTTP call. Otherwise the first slow provider request recreates the deafness with a different logo. It records the request boundary before releasing a transport worker, then returns to its mailbox. That boundary tells later reduction which input could have influenced the request and which input arrived too late to be in its bytes.

Tools require another distinction. The execution worker runs the handler. A separate report worker observes the job until completion, a pattern, or the callback delay. Combining them would make a blocked handler prevent the running-job report that was supposed to rescue the caller. A hint can wake the report worker without stopping the execution worker.

This is not a promise to use exactly two goroutines. Processes need readers; transport needs workers; jobs synchronize their own output and lifecycle. The useful count is the number of *owners*. One owner orders the conversation. The job subsystem owns job state. Messages cross those boundaries instead of shared mutable guesses.

The source’s earlier concurrency history makes that point concrete: the recurring fault was a parent reading stale child facts while the child’s actor wrote them. Adding more careful-looking reads did not establish ownership. Moving the read onto the owning goroutine, or communicating through the mailbox, did.

## A Hint Has to Reach the Model

A mailbox acknowledgment is not the finish line. “We received your message” can be perfectly true while the model remains oblivious for an hour.

Suppose the model starts a test and chooses a callback delay of 3,600 seconds. A human then says, “Check the first failure before changing anything.” The actor can now record that hint promptly. It must also wake the foreground job observation promptly. Otherwise the queue is responsive and the conversation is not.

The wake ends the observation, not the test and not the turn. The outstanding tool call gets one correlated return under its original call ID: the existing handle, bounded unseen output, and the state actually observed. If the test is still running, the return says so. If completion won the race, the return reports the real completed outcome. “Observation interrupted; job still running” is an entirely coherent result. “The test failed because somebody spoke” is not.

The next model request includes both that return and the incoming hint. The call is paired, the new instruction is visible, and the model can decide what to do while the original job continues. It need not wait for the test to finish before answering a question or issuing fresh calls.

This applies equally to a later `wait_for_job`. Waking it does not create replacement work. A subsequent wait observes the same handle using the new call’s limits. Available unseen output, or a retained finished result, can return immediately. The job still owns its output file and cursor; ending a wait neither discards bytes nor entitles the next wait to repeat bytes already consumed.

An addressed question is more demanding than a notification that somebody wants attention. Its sender, recipient, and request or turn correlation must survive delivery along with the actual question. Without that information, the model may be awake but unable to answer the right caller. Such a message is reliable actor input, not a progress cue that an observer may lose.

Bill reported the operational version: student agents were calling a parent stuck in `join_agents`; the calls timed out without answers. The parent was waiting for the students to finish, while the students needed the parent to proceed. A longer timeout only gives the stalemate more comfortable seating.

The remedy is the same observation boundary. Wake the parent’s foreground wait, preserve the original join job, deliver the question, and later resume observation of that same join handle. Bill’s report motivates the design; it is not a captured wire trace establishing the incident’s exact timing. This chapter demonstrates the seam with an application-registered join-like handler, not with a new child-agent framework. Later supervision tools can inherit the mechanism without being invented here.

There is a second place where a hint can disappear: an HTTP response that contains only final assistant text. The hint arrived too late to alter the request already sent, but that does not make it somebody else’s problem. Preserve the assistant reply, apply the pending hint, and continue the *same* turn with another request. The source records an initial implementation that instead returned the final reply, leaving the admitted hint waiting for a later `Ask`.

That implementation had received the message. It had not delivered the message in the sense the human needed. Responsiveness ends at useful model input, not at a cheerful acknowledgment.

## The Rest of the Batch Is Still Real

A model response can ask for several tools. Dispatch remains serial within that batch. A wakeup may therefore find one call finished, one job running, and several calls not yet dispatched.

Sending the hint straight to the provider at that point would leave dangling calls. Finishing every queued call first could hide the hint behind another long wait. Replaying the whole batch would duplicate effects. The right answer is less magical: account for each call according to what actually happened.

Completed calls keep their real results. The dispatched job gets the truthful observation result and its existing handle. Every undispatched call gets a correlated error saying it was interrupted and not executed. Those handlers do not run, have no effects, and allocate no job IDs. The original assistant calls and signatures remain intact; the framework supplies honest results rather than editing the model’s past to make the present convenient.

Only then can the next request carry the incoming message. The discarded tail does not replay automatically. If the model still wants one of those operations, it can ask again with the new information in view.

The dispatch boundary decides the close race. Once a call has started, it belongs to the dispatched case. It cannot be called “not executed” merely because its handler has not returned. That distinction is especially important for effects: a write can happen before the result arrives.

Chapter 4’s literal-next-call rule for `tool_limits` survives this interruption. Pending limits are consumed by the next call even if that call receives a not-executed error, then reset. An undispatched `tool_limits` cannot arm new settings. Without those rules, a setting intended for the discarded tail could quietly govern fresh calls after the hint. The user would have corrected the plan while an invisible piece of the old plan remained armed.

This is one reason not to implement wakeup as a convenient return from whichever nested function happens to notice it. The conversation owner has bookkeeping to finish across the entire response batch. Early control return is useful only if the history remains a valid account of the work.

The timer needs a finish line too. A message and completion can arrive together; both facts must survive, but each outstanding observation call still receives only one return. Retire the old timer when that observation ends. Otherwise it can wake later and manufacture a ghost result during a new wait.

Output notifications have a narrower meaning: recheck the observation’s conditions. They do not necessarily mean return immediately. The source records a wakeup implementation that returned on every new byte and made `send_input` report a PTY’s echo before the process exited. Seeing your own input echoed is not seeing the command finish. Unless an actual hint or question wakes it, `send_input` retains its pattern, completion, or callback-delay conditions. The immediate read of available bytes on a resumed `wait_for_job` is a different operation, not a universal rule for all observers.

Concurrency has a talent for making two reasonable sentences sound equivalent until a terminal repeats your command back to you.

## Four Different Ways to Stop

A hint, an interrupt, a job kill, and shutdown all disturb waiting. They should not share one vague cancellation policy.

A hint changes the pending conversation context and may end the foreground observation. It does not change the turn state or end the active turn. Explicit live input is classified by the Agent using its current reduced state: idle input starts a turn; input during an in-flight or tools-pending turn becomes a hint. The stdin reader must not classify from a snapshot that may already be stale.

An interrupt ends the active turn and cancels its in-flight model request. It does not end the actor. The actor keeps draining messages, records later results, and can accept another turn. `Interrupted` is a state rather than a side flag so that a late result cannot accidentally resume the old turn’s dispatch.

The unresolved calls still need honest pairing. A wait was interrupted; its job may remain alive. Undispatched calls were not executed. A later real result must not become a duplicate vendor return, nor finish whichever newer turn happens to be active. The accepted implementation retains that later fact as `job_report`, with the bytes in the job’s output file. Old work is still evidence even after it has lost the right to advance the conversation.

An explicit job kill has different semantics. It stops managed processes through the job machinery and makes the killed lifecycle terminal. A Go-only handler may not be physically cancellable; the kill can return without joining it and suppress its late output. That is not the policy for an ordinary interrupted turn, where the actual later report remains valuable.

Shutdown closes admission, cancels in-flight HTTP, stops managed processes, joins workers, records outcomes, and saves. A non-cancellable Go handler may delay shutdown because its actual outcome still has to be collected. Pretending it stopped would make shutdown quicker by making shutdown false.

The actor has to remain alive long enough for joined workers and admitted kill calls to deliver terminal facts. A shell handler returning after launch is not the process finishing; Chapter 4 already taught that distinction. It matters again at the final save. Repeated shutdown is safe, and later requests receive a distinguishable stopped error rather than slipping into an abandoned queue.

The historical human-CLI demonstration shows the useful separation. The source records a hint acknowledged during a running shell job, an interrupt, and a new turn answering while the old job remained alive. Its real late output arrived afterward. That is evidence of an interface usable before old work finished, not evidence that interruption had killed the work.

The same record preserves a metadata correction: the initial implementation emitted `job_finished` with status `running` when the shell handler returned from launch. The late output was genuine; the lifecycle label was wrong. Correcting the terminal-state guard did not require pretending the original demonstration had measured something different.

## Synchronous Does Not Mean a Second Loop

An embedding application may still want the simplest possible call: ask a question, receive its answer. The public API need not become a tangle of channels just because the implementation becomes an actor.

`Submit` posts a separate request and returns its private, one-element buffered reply channel. `Ask` waits on that request’s reply, caller cancellation, or actor shutdown. The actor keeps working independently. The buffer lets it deliver a completed result even when the caller has stopped waiting.

“Private” is the important word. Waiting for a shared idle state and then reading the Agent’s last text would not identify which request completed. Two callers could receive the same answer, or one could receive the other’s error. Reliable request completion carries that request’s text and error. Shared progress cannot impersonate it.

Blocking `Ask` and `Submit` deliberately queue separate requests rather than turning into live hints. This preserves a useful distinction between “answer this next” and “adjust what you are doing now.” The legacy stdin form with a `user` key and no `kind` goes through that request/reply adapter. Explicit live prompt and hint forms use current turn state. Compatibility is preserved by routing old behavior through the new owner, not by keeping the old orchestration loop hidden nearby.

When persistence is configured, the actor saves before replying. A failed save is a failed completion, not a durable success with a logging footnote. An empty `LogPath` leaves persistence to the embedding caller; it does not silently promise a saved conversation. Observers can be notified of the outcome without their delivery becoming a prerequisite for the caller’s reply.

The Agent starts its actor lazily and idempotently. Repeated calls use the same owned mailbox consumer. Engine retains one-request rendering and transport; History retains reduction and persistence; Agent orders the multi-round conversation and runtime context changes. The old `Engine.Ask` orchestration and sibling loops must go after callers migrate. Two implementations that happen to agree today are two policies waiting to disagree tomorrow.

The tool-round limit belongs to this same owner. A positive N permits N dispatched tool batches in one human turn. Zero selects the named default of 200, not infinity. The effective setting is read once at turn start. After the last allowed batch, the model still gets to consume its results and finish; only a request for another batch triggers explicit not-executed results for every call and a distinguishable limit error. No dangling calls, no rewriting signed assistant content, and no permanent ban on a later turn.

A synchronous facade is perfectly compatible with responsive internals. The duplication to remove is not the waiting caller. It is the second authority over what a turn means.

## The Witness Is Not the Driver

The outbound seam has almost the opposite reliability needs from the mailbox.

Before this change, a caller can ask and inspect the finished History. It cannot conveniently watch state changes, tool activity, or output taking shape. An observer receives typed observations as the actor commits facts. It is told what happened; it does not decide the next transition.

The boundary is deliberately after reduction: append the event, update state, then notify. If the actor announces a tool call before committing it, an observer can see a future that History does not yet contain. A crash between the two would leave an observation without its supposed fact. Post-reduction notification keeps the account in the right direction: observations report stored facts rather than promising impending ones.

The public vocabulary is small. `StateChanged` carries the old and new states. `PartDelta` carries a part identity and incremental text. `PartFinal` carries the complete part and its committed sequence. The core seals the observation interface so the framework owns this vocabulary and consumers can handle its defined variants. Application-level workflow markers need not become new core variants merely because an application wants to print them.

A delta has no committed-event sequence number. It is a transient preview of an open part, not independently replayable history. The finalizer ties that part ID to the committed part.

The vocabulary is ready for streaming, but the provider seam here still returns a complete response. The implementation emits a whole-part delta followed by a finalizer. A type called `PartDelta` is not proof that live tokens are arriving.

No observation gives the recipient a mutable pointer into live History. That prohibition includes more than the top-level struct. A value copy can still share a slice backing array or a nested metadata map. If an observer edits its snapshot and changes the next provider request, the witness has acquired a steering wheel by aliasing. Immutable or deeply isolated views preserve the single-owner argument across nested data, not just at the function signature.

There is an unavoidable performance bargain. Registered observer methods run synchronously on the actor goroutine, so they must return promptly. A logger that needs disk work or a UI that needs its own event loop should copy the observation and enqueue it privately. A buffered-channel adapter uses a nonblocking send and drops when full. It must not block the actor to keep every progress cue.

That means a slow display can miss progress. The alternative is a slow display freezing conversation execution. The design chooses an authoritative event log and a reliable request reply over an infallible progress feed. Observation channels are lossy cues; readers use history and independent snapshots for facts that must survive loss. A callback registered as an observer is legitimate behavior at an explicit boundary, not a function-pointer escape hatch for a broken package dependency.

“Notify me” and “make my receipt a condition of your progress” are very different requests. Only the first belongs in this seam.

## Watching Is Not Collecting

The same distinction governs `Wait` and `WaitAny`.

An Agent’s `Wait` consumes observations until one satisfies a predicate or its caller’s context ends. It is not a replaying subscription that delivers every event to every waiter. Ensemble’s `WaitAny` consumes a matching observation from a merged progress stream, with the source Agent attached. Each Agent keeps its own mailbox and conversation owner; Ensemble supplies the identity envelope through the existing parent relationship.

That is enough to watch several Agents without opening their private histories and spelunking for clues. It is not enough to acknowledge a particular submitted request. A shared idle transition may concern different work, and a bounded progress stream may drop precisely the cue a coordinator hoped to use. Decisions that require completion must retain the relevant request replies reliably, then publish progress about the result.

The source’s review history includes a useful scope correction. An earlier “one wakeup” description led a reviewer to demand grouped completion, and the student initially agreed. The actual demonstration concerned observations from distinct Agents, not a measured wake count or an API that collects a group of results. Bill withdrew the added requirement. An attractive abstraction is not automatically a missing feature.

The concrete application here has an author, editor, and reviewer. Go constructs the peer Agents with different prompts and application-specific tools. The author stores a draft and can count its words; the editor reads and replaces that stored text; the reviewer records an acceptance or rejection with notes. These are tool declarations and handlers supplied by the application, not manuscript-shaped builtins added to Ensemble.

The application advances sequentially using reliable request completions. The reviewer receives the actual edited text, not merely the editor’s assertion that editing occurred. The source’s historical workflow shows that difference nicely: the author’s first stored draft missed the requested word count, the count tool exposed the miss, and the author revised the stored text before the editor worked on it. The tools make the shared artifact inspectable rather than relying on eloquent claims about it.

Even this sequential workflow needs synchronized draft storage. A tool worker can outlive the turn that launched it; “we run the stages in order” is not a lock on application data. If a role is interrupted, its request returns an error instead of advancing with an unfinished draft. EOF waits for the submitted workflow before shutdown.

Tagged observations let a display distinguish the Agents. Completion records and stage markers belong to the application protocol, while each Agent keeps its own durable history and sequence space. Adding the reviewer should require application wiring, not a patch to the framework. That is the payoff of exposing composition rather than implementing one favored collaboration.

It is also the limit. These Agents do not invent their team or choose the workflow. A different collaboration still needs different Go glue. The framework is composable; the application remains hardcoded. Sub-agent spawning and its supervision contract come later.

## An Attachment Must Actually Arrive

Media capability may seem like an odd companion to an actor loop. It shares the same underlying concern: accepted input is not necessarily delivered input.

A swallowed hint leaves the model answering an old question. A silently dropped attachment leaves it answering without material the programmer believes it received. In both cases, improving the prompt can become a very elaborate attempt to repair the wrong boundary.

The model catalog therefore declares image, audio, video, and document support separately. Image support implies none of the others. An unsupported attachment fails locally with a diagnostic naming the model and media type, rather than disappearing from the request.

An unknown model has no optimistic default row. Text alone remains allowed; rendering a blob requires known support. That distinction avoids converting “we do not know this model’s media capabilities” into the much broader and unsupported claim “this model cannot accept text.”

A blob carries a `Ref`: a local path, a URI, or a handle. If the blob is superseded, the redacted part preserves its reference so the address survives. The address is not a promise that an external file or job output will remain available forever.

Nor is retaining a locator equivalent to rendering an attachment. Local-path rendering reads the actual file bytes into the vendor’s attachment structure. URI and handle rendering remain deferred even though the reference vocabulary can represent them. This is a useful kind of future readiness: the log can retain identity without claiming an implementation the renderer does not have.

The source’s media demonstration sent a local red image through a public-API client and received `red`, then separately observed local refusal of unsupported audio. Positive audio, video, and PDF paths were checked at the rendering boundary, not presented as additional live model demonstrations. The distinction matters: a correctly shaped request and a model’s observed consumption are related evidence, not interchangeable evidence.

One repaired check sharpens the engineering lesson. Looking for attachment bytes somewhere in the request missed a wrong Gemini `mimeType` key. Decoding the actual attachment fields caught the error. The vendor consumes a particular structure, not a bag of reassuring substrings. Whether the input is a hint or an image, delivery has to be established at the boundary that actually consumes it.

## Available Is a Property of the Whole Path

The difficult failures here often look locally correct.

The stdin reader accepted the hint. The actor recorded it. The tool ran on a worker. The timer had a legitimate deadline. The observer emitted an acknowledgment. Every sentence can be true while the model still cannot act on what the human said.

The useful question follows the whole path: was the message admitted reliably, ordered by the conversation owner, paired with an honest observation return, and included in the next request without waiting for unrelated work to finish? At the other end, did the caller receive its own completion after the configured save, rather than infer it from a lossy progress cue?

That also explains what convincing concurrency evidence must distinguish. A long-delay case can show that a hint wakes observation promptly, but it cannot show that an obsolete short timer stays retired during a later wait. A passing interrupt case cannot establish hint wakeup, because interrupt ends the turn. A hint acknowledgment cannot establish delivery into the next provider request. Completion-before-message, message-before-completion, and the close race all need to preserve the message and the real outcome without duplicating the tool return.

These are questions about ownership and causal order, not reasons to let an hour elapse or to commission another broad live campaign. The source’s historical provider demonstrations exercised different feature slices; they were not every feature on every provider. Their value is in the particular boundary each makes visible.

Chapter 4 gave work a lifetime independent of an individual wait. Chapter 5 gives the conversation an owner independent of the work it observes. The human can speak while tools run; the model can hear without killing them; an application can watch without driving; and a synchronous caller can get its own answer without owning the loop.
