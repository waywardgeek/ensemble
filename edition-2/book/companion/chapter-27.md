# Chapter 27: Goals

Bill kept having to say “please continue.” In the [source account](../edition-3/chapter-27.md#271-why-please-continue-became-a-feature), an agent would finish a progress report while its authorized assignment still had work left. Stronger instructions could explain what remained, but could not run themselves after the turn ended.

The same account records a dangerous misdiagnosis: the assistant twice treated quiet observations as evidence that a student had stopped, while Bill reported it was still running. Those are different problems. An idle Agent needs another opportunity to act; a quiet but active request may need nothing at all. Goals addresses the first without pretending to diagnose the second.

The feature combines an editable goal file with an optional idle reminder. The file supplies direction on each model request. The reminder can start a new turn after the Agent becomes idle. As with the gateway, new work enters through the existing actor rather than a second execution loop. The current source describes an implemented, locally tested mechanism, not a deployed scheduler or proof that real models reliably continue unfinished work.

## Three separate choices

Availability, configuration and automatic waking are separate states. Loading the `goals` skill exposes `goal_control`; it neither selects a file nor enables reminders. Before loading, the control is absent from both tool declarations and dispatch grants, including for legacy Agents without a primary skill. Compiling a control into the runtime doesn't make it callable everywhere.

A configured file supplies the continuing assignment outside conversation History. It can describe unfinished work, completion conditions and authorization boundaries, and ordinary file tools can revise it. Enabling reminders adds a timer. Disabling them leaves the file configured and its contents present in subsequent requests. Clearing the file selection removes that contribution and stops reminders. Unloading the skill withdraws the control and clears configuration, but does not delete the file; loading it again doesn't restore the old settings.

These distinctions let a person stop automatic activity without discarding direction. They also prevent the runtime from treating “reminders off” as “assignment complete.” The model might need permission, or the human might simply want a pause. The explanation belongs in the conversation, not in an inference from the enabled flag.

The control reports status when called without fields and merges only the fields supplied in an update. Invalid updates leave the previous configuration intact. The initial delay is eight minutes; an override must be positive, finite and representable as a duration. The fixed interval replaced proposals for durable scheduling and escalating delays. The simpler design has no backoff counter or recovery scheduler to maintain.

Only the file is durable. Its selection, reminder setting and timer belong to this Agent in this process. Children, compaction workers and restarted Agents do not inherit an enabled reminder lifetime. A surviving file therefore cannot silently restore yesterday's authority. Nor does configuring it add filesystem access, provider permission, deployment authority or permission to repeat an external effect.

## Direction belongs in the request, not the transcript

The Engine assembles its usual instructions, obtains the current goal contribution from the Agent, and passes the combined request-local text to the provider renderer. Primary instructions and loaded skill manuals remain in place. The goal isn't appended to canonical dialogue, substituted for the primary skill or stored in shared mutable prompt state.

The existing filesystem owner reads the file once for each request render, including successive requests within a tool-using turn. This timing matters when a model edits its own goal: the next request must see the edit, not the version read at turn start. Reading independently in multiple renderer branches would introduce the opposite problem, allowing one request to depend on several filesystem observations. Already-sent bytes remain unchanged, while handoff and compaction leave later renders free to return to the configured file rather than a summary's recollection of it.

The contribution appears exactly once in the provider's instruction lane: SYSTEM text for Anthropic, a system message for Chat Completions, developer input for Responses, or `systemInstruction` for Gemini. Finding the text somewhere in a packet is insufficient; a tool result or user message isn't equivalent placement.

The [Codex Responses correction](../edition-3/chapter-27.md#274-read-once-for-each-request) shows why both placement and duplication matter. That route already supplied base instructions through developer input and a required top-level `instructions` field. Passing the combined text to both duplicated the new goal. The correction retained the inherited base-instruction behavior but added the goal only to developer input. A whole-request count caught what checking a single lane had missed.

The file remains direction, not execution evidence. A note saying that a test is probably running cannot establish the job's state. That distinction becomes consequential when the previous action spent money, sent a message or changed external state: another opportunity to assess work must not become automatic repetition of its last command.

## The actor admits the reminder

The actor already owns turn admission, History, cancellation and completion. It also owns goal configuration, timer generation and the next due time. The timer callback only enqueues a generation-tagged wake message. It doesn't call the Engine, inspect model output or replay a tool invocation.

A callback is a candidate, not permission to start. When the actor receives it, current state determines eligibility. A stale generation cannot revive a cleared or replaced configuration, and stopping a timer isn't enough by itself: its callback may already be queued. Generation checks also prevent an old callback from consuming a newly enabled timer.

Idle means no active turn and no queued ordinary input. Ordinary input takes priority and invalidates the idle timer. An active request remains busy however long it produces no output. Once a turn ends, a fresh full delay begins. Background jobs or children may still be running during that interval because their lifetimes aren't the Agent's turn lifetime.

At an eligible expiry, the actor checks that the goal remains readable and admits a labelled automatic reminder through the normal turn path. The reminder asks the model to inspect actual progress, jobs, children and saved results, then continue authorized work or disable reminders and explain why it needs the human. History identifies it as system-originated, not a newly authored human instruction, and it doesn't duplicate the goal text.

A failed provider request can end a turn and allow another full idle delay. The later request is a fresh assessment with current History and goal contents, not a transport retry of captured bytes. A request that never leaves the active state gets no such wake. The timer is deliberately unable to repair every stalled connection.

Human controls dominate queued callbacks. Pause suspends reminders; resume begins a fresh delay. Interrupt disables them rather than leaving a wake poised to undo the stop. Clearing, unloading, shutdown and terminal submission invalidate further waking. None of this cancels a healthy background job merely because the Agent has become idle.

## Errors must not manufacture progress

Goal reads use the Agent's existing sandbox. A rejected outside path or symlink leaves the previous configuration intact rather than creating a convenient second filesystem authority. If a selected goal becomes missing, unreadable or empty, automatic waking stops without sending a request, reminders become disabled, and status exposes the error. Repairing the file doesn't automatically re-enable them.

Ordinary human requests remain possible. They receive a labelled unavailable-goal notice in SYSTEM so the Agent can help repair or clear the configuration. Silently omitting the goal would conceal lost direction; serving cached contents as current would conceal the file failure. The off switch works even when the file cannot be read, since stopping reminders must not depend on the resource that broke them.

Two [local regressions](../edition-3/chapter-27.md#276-keep-failures-visible) exposed misleading lifecycle assumptions. In one test, an owned directory replaced the journal file, causing durable append to fail before a reminder turn could start. The initial implementation nevertheless published “working” and left no next timer. It had confused attempting startup with admitting a turn. The correction checks whether a turn actually exists before publishing working or installing the automatic-turn marker; synchronous refusal returns to the normal idle policy.

A broader GUI test found a deadlock during rejected child construction. The parent held its child-construction lock while waiting for discard to finish. The new goal publisher tried to notify that parent, overlooking the existing discard path's suppression of normal publication. The fix made goal status respect that boundary rather than redesigning fleet locking. Adding an observer had introduced a dependency into a lifecycle that deliberately excluded it.

## Show the state the core actually owns

The GUI observes a snapshot for the selected Agent: file, reminder setting, delay, phase, due time and error. It distinguishes no goal, working, waiting, paused, reminders off and error. Working with reminders disabled says so explicitly, since a human-started turn can still use the configured goal. None of these labels asserts completion.

Initial connection, reconnect and Agent selection replay the core's status through the existing WebSocket path. Selection clears the previous Agent's display while the new snapshot is pending; frames for another Agent are ignored. A disconnected browser says it is disconnected instead of presenting an old snapshot as current. There is no browser-owned wake timer for reconnect to restart.

The [source's accessibility decision](../edition-3/chapter-27.md#277-give-the-human-a-truthful-indicator) follows Bill's reliance on speech and difficulty inspecting small visual distinctions. Text labels carry meaning that a colored dot would not, with file, delay, due time and error also available in the title and accessible label. The feature adds no automatic speech. Conversation and Actions remain available for the separate questions of why the Agent acted and what actually ran.

## What the local evidence establishes

The [implementation record](../edition-3/chapter-27.md#278-exercise-the-boundaries) describes tests using real Agents, skill loading, tool dispatch, rendering and owned files against local synthetic HTTP responses. They exercised same-turn goal edits, instruction placement, ordinary-input priority, stale callbacks, human stops, file repair, sandbox refusal and independent background work. Short delays tested transitions; the eight-minute default was asserted separately, not observed through an eight-minute wait.

An isolated headless-browser run used the real GUI with synthetic model responses. It observed reconnect preserving the core due time, child selection showing no inherited goal, Interrupt turning reminders off, foreign-frame filtering and disconnected state after the owned server stopped. This establishes browser text, accessible attributes and measured layout, not human visual or hearing assessment; speech handling was transcript-only.

The source reports passing focused core race tests, the full GUI race suite and affected static and compilation checks after the corrections. These are bounded local results, without a live provider, real credentials or a running user session. They establish neither deployment nor an installed or enabled goal in this collaboration. Synthetic responses can expose a duplicate instruction or a stale wake; they cannot establish that a real model will correctly decide whether to continue, wait, stop or ask for permission.
