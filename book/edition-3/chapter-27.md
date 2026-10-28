# Chapter 27: Goals

> **Development status.** Implemented and locally tested in the current runtime under Bill’s explicit self-coding and self-review exception. This is not a deployment or an executed student Chapter 27.

## TL;DR

Give an Agent an editable goal file and an optional idle reminder. These are different mechanisms. The goal supplies direction on every model request. The timer supplies another opportunity to act after a turn has ended.

- Put the configured goal in SYSTEM on every request. Keep the primary instructions and loaded skill manuals in place. Read the file afresh. Editing the file changes the next request without rewriting History or asking a summary to remember the goal.
- Make the feature explicitly opt-in through a loadable `goals` skill. Loading it exposes `goal_control`; it neither configures a goal nor starts a timer. The control is absent from default declarations and dispatch grants.
- Let `goal_control` select a file, turn reminders on or off, and set `idle_seconds`. With no fields it reports status. The default delay is eight minutes. Turning reminders off keeps the goal in SYSTEM; clearing the file selection removes it. Disabling must work even if the file has disappeared.
- The Agent's existing filesystem owner reads the file. Do not bypass its sandbox. A missing or empty goal is an explicit error, not proof of completion. Stop automatic reminders on that error. Ordinary human requests receive a labelled SYSTEM notice so the agent can help repair or clear the configuration, rather than silently working without its goal.
- Start one timer with the same delay each time the Agent becomes idle: no active turn and no queued ordinary input. Background jobs or children may still be running. On expiry, post a message to the normal actor event loop; do not call the Engine or replay commands from a timer callback.
- Have the automatic message ask the Agent to inspect the goal and actual progress, continue unfinished authorized work, or turn reminders off and explain why it needs the human. Label the message as automatic. Do not invent a new human instruction or duplicate the goal text in conversation.
- A normal response or a failed request may return to idle and begin another full delay. An active request is not idle merely because it is quiet. Pause suspends reminders; resuming begins a fresh delay. Explicit Interrupt disables them. Clearing, unloading the skill, or shutting down invalidates queued timer callbacks.
- Keep configuration local to this Agent and this process. Children and restarted Agents do not silently inherit enabled reminders. The goal file survives, but the first version is not a durable scheduler.
- Show a text-labelled goal indicator in the GUI for the selected Agent. Distinguish no goal, working, waiting, paused, reminders off, and error. Replay the current state on reconnect. The GUI observes the core timer; it does not own another one. Off does not mean completed.

This is not a task tree, a network watchdog, or automatic command retry. It helps an idle Agent resume useful work. It cannot rescue an Agent whose request remains active forever, guarantee good judgment, or authorize effects that the human has not permitted.

## 27.1 Why “please continue” became a feature

Bill kept having to say “please continue.” The agent would finish a response while an authorized assignment still had work left. A useful progress report had become an accidental stopping point. Stronger instructions helped describe the desired behavior, but instructions cannot run themselves after a turn ends.

We also discussed requests that might have failed or stopped making progress. Those cases need care. During the preceding chapter work, I twice treated quiet observations as evidence that a student had stopped, while Bill reported that it was still running. That was my judgment error. A timer does not make the earlier diagnosis correct.

This feature handles an Agent that has actually become idle. It gives that Agent another opportunity to inspect the work. A request that remains active, however quiet its output, stays active. Detecting and repairing a stuck request would be a different feature with different risks.

The design began with more machinery: durable scheduling state, recovery behavior, and eventually a proposed escalating reminder interval. We reduced it to a file, an explicit control, and one timer owned by the existing actor. Bill chose a fixed eight-minute interval. There is no backoff counter, task hierarchy, or second scheduler to explain.

## 27.2 Put the goal somewhere editable

A goal file carries the continuing assignment outside the conversation. Ordinary file tools can update it, and conversation compaction does not need to preserve a special immortal message. The Agent reads the configured file when it renders each request.

A useful goal describes the work and its boundaries. For example:

```markdown
Finish the local parser correction and its focused regression test.
Inspect the current diff and any running job before starting another command.
Do not install, restart, publish, or change unrelated files.
When the correction is verified, turn reminders off and explain the result.
If a consequential decision needs Bill, turn reminders off and ask him.
```

That file supplies direction, not evidence that a command ran or a test passed. If it says “the test is probably still running,” the Agent still needs to inspect the actual job. A stale note is particularly dangerous when the last action had an external effect: repeating a request can spend money, send a second message, or overwrite useful work.

The existing Sandbox owns these reads. Configuring a goal does not create another filesystem authority, bypass a rooted path check, or grant a missing tool. The path has the same meaning as other Agent-directed file paths. A rejected outside path or symlink leaves the previous configuration intact.

Only the file is durable in this first implementation. The selected path, enabled flag, and timer are per-Agent session state. A child does not inherit them, and restarting the process does not silently restore them. That is a useful safety boundary: finding yesterday's goal file is not enough to resume yesterday's authority.

## 27.3 Make activation explicit

The `goals` skill exposes `goal_control`. Listing the skill as available does not load it; loading it does not select a file or enable a timer. Legacy Agents do not gain a callable control just because its implementation was compiled into the program.

After creating a goal file through the ordinary file tools, an Agent can load the skill and configure it:

```json
{"name":"goals"}
```

That is a `load_skill` call. The separate `goal_control` call is:

```json
{"file":"GOAL.md","enabled":true}
```

Omitting `idle_seconds` uses the current delay, initially 480 seconds. An explicit override must be positive, finite, and representable as a duration. Short overrides make isolated tests practical; they do not change the eight-minute default.

Calling `goal_control` with `{}` reports status without changing it. Updates merge the fields supplied by the caller, and an invalid update changes nothing. The two stopping operations have deliberately different meanings:

```json
{"enabled":false}
```

This stops reminders but leaves the goal in subsequent requests. It works even if the file has become unreadable, because a broken file must not disable the off switch.

```json
{"file":""}
```

This clears the selected goal and stops reminders. Unloading the `goals` skill also clears the runtime configuration and timer, while leaving the file alone. Loading it again does not reactivate the old configuration.

None of these operations means the assignment is complete. A person may turn reminders off to think, leave, or resolve a decision. The model may turn them off because it needs permission. The explanation belongs in the conversation; the runtime should not invent a completion judgment from a boolean.

## 27.4 Read once for each request

The Engine assembles its ordinary SYSTEM instructions, asks the Agent for the current goal contribution, and passes the resulting request-local string to the chosen renderer. It does not store the goal in History, replace the primary skill, or mutate shared prompt state.

The read happens on every Render, including the next request within a tool-using turn. If a model calls `write_file` to change its goal and then continues, that next request sees the new bytes. Reading only at turn start would miss precisely the edit the model just made. Reading separately inside several renderer branches would make one logical request depend on several filesystem observations.

Each provider still uses its established instruction lane. Anthropic receives SYSTEM text, Chat Completions receives a system message, Responses receives developer input, and Gemini receives `systemInstruction`. Tests inspect those fields, not merely whether the goal text occurs somewhere in the request. A tool result containing the text would not satisfy the contract.

Source review found an extra wrinkle in the existing Codex Responses route. It already supplies base instructions through both developer input and the required top-level `instructions` field. Passing the combined goal string to both duplicated the goal. The correction preserved that inherited base-instruction behavior and put the new goal contribution only in developer input. A whole-request count now catches the duplicate that a single-lane assertion missed.

The file is read afresh, but an already sent request is unchanged. Editing a file cannot retroactively alter bytes on the wire. Handoff and compaction likewise do not copy or rewrite the goal; subsequent rendering returns to the configured source.

## 27.5 Let the actor decide when to wake

The actor already owns turn admission, History, cancellation, tool dispatch, and completion. It also owns goal configuration, timer generation, and the next due time. The timer callback has one job: enqueue a generation-tagged wake message.

When that message reaches the actor, it is only a candidate wake. The actor checks its current state before doing anything with it. An old generation cannot revive a cleared configuration. Pause cannot be bypassed by a callback that was already queued when the timer was stopped.

Ordinary input gets priority. If an ordinary turn is pending, the actor takes that turn and invalidates the idle timer. If a turn is active, the Agent is busy even when no output is arriving. Once the turn finishes, a fresh full delay begins. A background Job may still be running during that idle interval; the reminder does not cancel it or assume it needs replacing.

On an eligible expiry, the actor checks that the goal can still be read and starts a labelled automatic reminder through the normal turn path. The reminder asks the model to inspect the goal, jobs, children, and saved results before deciding what to do. It is recorded as a system-originated reminder, not as a newly authored human request, and it does not replay an old tool call.

The fixed delay applies again after each completed reminder. A failed provider request can also return the Agent to idle and start a new full interval. This is a fresh assessment opportunity, with current History and current goal text. It is not a transport retry of the old body.

Human controls remain stronger than the timer. Pause suspends reminders; unpause begins a full new delay. Interrupt disables them. Clear, unload, shutdown, and terminal submission invalidate further waking. Compaction workers cannot acquire an independent reminder lifetime.

## 27.6 Keep failures visible

A missing or empty goal must not silently disappear from the instructions. Automatic waking stops without sending a request, reminders become disabled, and the status exposes the error. Repairing the file does not re-enable them.

An ordinary human request is still useful in this state. Its SYSTEM instructions contain a clearly labelled unavailable-goal notice so the Agent can help repair the file or clear the configuration. It does not receive stale goal contents presented as current. This separates “we cannot safely wake automatically” from “the human cannot ask for help.”

The startup path also matters. In a regression test, an owned directory replaced the Agent's journal file, making durable append fail before a reminder could start. The initial implementation nevertheless published “working” and left no next timer. It had inferred an active turn from the attempt to begin one.

The fix asks the existing runtime what actually happened: if a turn exists, publish working; if startup refused synchronously, return to the normal idle policy. The automatic-turn marker is installed only when the turn exists. This keeps error handling inside the same owner rather than adding a separate retry loop.

A broader GUI test exposed a different lifecycle mistake. Rejected child construction already had a discard path that suppressed normal publication. The new goal status publisher overlooked it and tried to notify the parent while that parent held its child-construction lock and waited for discard to finish. The resulting deadlock was a production defect, not a slow test. Goal publication now respects the existing discard boundary; fleet locking did not need a redesign.

## 27.7 Give the human a truthful indicator

The GUI displays a snapshot from the selected Agent. The core supplies the configured file, reminder setting, delay, phase, next due time, and error. Initial connection, reconnect, and agent selection replay that same status through the existing WebSocket path.

The indicator uses text: “No goal,” “Goal: working,” “Goal: waiting,” “Goal: paused,” “Goal: reminders off,” or “Goal: error.” Working with reminders disabled says so explicitly. File, delay, due time, and error are also available in the title and accessible label.

Selection clears the previous Agent's display before the new replay arrives. Frames for a different Agent are ignored. A disconnected browser says it is disconnected instead of continuing to present an old snapshot as current. The browser does not own a wake timer and cannot restart one by reconnecting.

Bill relies heavily on speech and cannot reliably inspect small visual distinctions. A colored dot would give him too little information. We kept meaningful text without adding automatic speech or another narrator, and kept the Conversation and Actions panes available for the separate questions “why?” and “what actually ran?”

## 27.8 Exercise the boundaries

The implementation tests use real Agents, their skill registry, actual rendering and tool dispatch, owned files, and local synthetic HTTP responses. One tool loop configures a goal and edits its file before the next request. Other tests distinguish an HTTP error from a still-active request, place ordinary input behind a queued callback, and verify that Interrupt prevents the stale wake from starting work.

The tests also check disabled reminders retaining SYSTEM text, failed reads and explicit repair, unload, shutdown, invalid updates, rooted path refusal, a live background Job, handoff, and compaction-worker isolation. Provider tests inspect both instruction placement and absence from canonical dialogue. Short test delays establish transitions; the default of 480 seconds is asserted separately rather than pretending an eight-minute wait occurred.

The GUI has a real WebSocket replay/selection test and a test of its production text renderer. An isolated headless-browser run used the real GUI and three synthetic model responses to load the skill, configure a goal, and end the turn. Reconnect retained the same core due time, selecting a child showed no inherited goal, and Interrupt changed the indicator to reminders off. Injected foreign-agent and selection frames exercised filtering and loading uncertainty. Terminating that owned server established the disconnected display.

These observations establish browser text, accessible attributes, and measured layout, not a human visual assessment or hearing result. The browser used transcript-only speech handling. No live provider, real credential, or running user session was involved.

Focused core race tests, the full GUI Go race suite, affected vet checks, existing Engine and skill tests, and compilation checks passed after the corrections. The failed journal, duplicated Codex goal, and rejected-child timeout remain part of the implementation record. A passing final run does not make those earlier failures disappear.

This is implemented and locally tested source in the current Ensemble runtime. It is not an installation, an enabled goal in this collaboration, a completed student execution of Chapter 27, or evidence that a real model will always make the right continuation decision. The runtime supplies a controlled opportunity to reassess work. The goal file and the model still need sound engineering judgment.

## Exercises

Use the existing library, actor, Engine, skills and GUI. Build the mechanism, not another Agent loop. Use a local provider fixture for timing and wire assertions; a fake-provider pass is not evidence about a real provider. Keep the actual goal, wake message, model request, and tool execution distinguishable in the evidence.

### 1. Make opt-in observable

Start with an ordinary Agent that has not loaded `goals`. Verify that `goal_control` is neither declared nor callable and that becoming idle produces no automatic request. Check the legacy/no-primary-skill case too; registering a compiled tool must not accidentally advertise it everywhere.

Load the skill through the real loader. Its instructions and control should become available, but no timer should start. Create a goal file through the existing file interface, then configure the file and enable a short test delay. Verify the returned file, delay and enabled state. Invalid settings must not partly replace a valid configuration.

Disable reminders without supplying another file. Repeat while the configured file is unavailable. The off switch must not depend on successfully reading the goal it is meant to stop.

### 2. Inspect the actual SYSTEM wire

Write a distinctive goal, make a real Agent request to the local fixture, and inspect the rendered request. Find the goal exactly once in the provider's SYSTEM/developer instruction lane. Verify that the primary instructions and skill manuals are still present. A text search across the whole packet is not enough: a goal misplaced in a tool result or user message fails this exercise.

Edit the file and make the next request. Include a second request within an ongoing tool-using turn so the test cannot pass by loading the file only at turn start. Verify that the new bytes replace the old goal contribution and that canonical conversation has not acquired duplicate copies. Turning reminders off must not remove the configured goal from SYSTEM.

Exercise the supported provider render paths with local requests, including the Responses path where selected. Establish lane placement, not provider obedience. A handoff or context rebuild must not erase the configured file source or move its contents into a conversation summary. Do not modify accepted handoff behavior to achieve this.

### 3. Let idle time wake the existing actor

Finish a turn and wait for the configured idle delay. Observe one automatic message and one ordinary new turn through the existing event loop. The reminder should ask for a current-state check; it must not claim that a child is finished, a command failed, or a prior action needs replay.

Start another ordinary turn before the timer expires. Hold its provider response with a controlled latch for longer than the idle delay. No reminder may start a concurrent request. Release the turn and verify that a fresh full idle delay begins. Queue ordinary input and verify that the idle reminder does not jump ahead of it.

Leave an ordinary background job or child running while the Agent itself becomes idle. A reminder may wake the Agent to inspect that work; it must not cancel, duplicate, or restart it. Waiting for a reminder is not an execution timeout for the existing work.

### 4. Recover an opportunity, not an old command

Have the fixture end one model exchange with an explicit transport/inference error. Once the turn has actually ended, the enabled idle timer should supply a later opportunity to assess the current goal. Verify the full delay and the distinct new request. Do not resend captured request bytes, fabricate missing tool results, or repeat an earlier side effect.

Contrast that with a deliberately latched active request. The goal timer must leave it alone. This proves the limit of the feature: it can wake an idle actor, not diagnose or repair every connection failure.

### 5. Make human control win

Test disabling just before expiry, and disabling after a timer callback has already been queued. Neither stale callback may create a new turn. Re-enable and verify that an old callback cannot consume the new timer's generation.

Pause before expiry. Reminders stay suspended until the Agent is unpaused, then a fresh delay begins. Interrupt and verify that reminders are off rather than poised to undo the stop. Test shutdown with a timer pending and with a callback queued. There must be no post-shutdown model request or reminder-owned work keeping shutdown alive.

Unload `goals`. Verify immediate control withdrawal, cleared goal configuration and cancelled reminder, while the file remains untouched. Reloading exposes the control but does not silently reactivate it. Verify isolation between two Agents and that a child or compaction worker does not inherit its parent's reminder lifetime.

### 6. Keep file failures honest

After valid configuration, remove the file, make it unreadable, or empty it. The next read must expose the problem. An automatic wake must not send an empty-goal request or enter a rapid retry loop. Status must show the error, not completion or an armed timer that no longer exists.

An ordinary human request should receive a clearly labelled unavailable-goal notice in SYSTEM, preserving the ability to repair or clear the configuration. Do not claim that stale cached goal text is current. Restore the file and verify fresh reads; re-enabling reminders is explicit. A refused outside-workspace path must retain the existing filesystem boundary rather than introduce a convenient host-file bypass.

### 7. Read the GUI state, including after reconnect

Drive the real status-update path and inspect the selected Agent's indicator. Check no goal, configured with reminders off, working, idle waiting, paused, and file error. Preserve the distinction between working on a configured goal and having reminders enabled. Do not label reminders off as completed.

Reconnect while waiting without causing another core transition. The current file, delay and due-time information must still appear. Switch between two Agents with different goal states; one Agent's frame must not overwrite the other's indicator. Use text labels and accessible descriptions, not color alone, and keep Conversation and Actions available.

Run deterministic tests against the production GUI code, then inspect an isolated browser instance for readability and layout. State separately what the automated tests, browser inspection, and human observations establish. Do not alter or restart the user's running GUI merely to obtain a screenshot, and do not add automatic speech or device operations to this exercise.

### 8. Review what the passing tests cannot decide

Trace the single owner of configuration, timer generation, automatic-message admission and public status. Confirm that callbacks only enqueue work, blocking file/model work is not hidden under shared locks, and the GUI does not start a second timer. Inspect request-local goal placement and the opt-in grant path directly.

Then try the intended human workflow: load the skill on request, write a useful goal, enable reminders, let the Agent return to idle, and inspect its resumed work. Keep paid-provider or deployment authority separate from local implementation tests. Report the configured delay, what actually woke, what remained unfinished, and how the reminder was stopped. A plausible final answer is not proof that every requested action occurred.

Only after these results and code review should the chapter body explain the implementation and the experience. Preserve useful failures and limits; do not turn a successful demonstration into a promise that goals prevent all accidental stops.
