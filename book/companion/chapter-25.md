# Chapter 25 — Goals

Bill would give Ensemble a job meant to occupy most of a day, then find it idle an hour later. The model had finished a response; the job was not finished. Resuming work required a human poke. After enough context compression, that poke might also need to reconstruct why the Agent had been doing the work in the first place.

Goals separates those two needs. A file keeps the agreed objective present in each ordinary request. An optional idle reminder gives the Agent another opportunity to inspect its state. The first preserves direction; the second supplies a turn.

## Keep the objective outside the conversation

A useful goal file records what remains, what counts as done, and where authority ends. It is closer to a short work agreement than a command to repeat. “Continue the migration, verify the changed paths, and stop before deployment” lets the Agent adapt to evidence without converting persistence into permission to deploy.

The `goals` skill grants `goal_control`. Loading the skill neither creates the file nor starts reminders. Ordinary file tools write and update the goal; the control selects the file, enables reminders, and can bypass one pending failure wait with `retry_now`. Calling it without fields reports status without postponing a pending deadline.

Selecting a file validates it through the Agent’s existing Sandbox before publishing configuration. The file must be nonblank, strictly decoded UTF-8, regular, confined, and no larger than one MiB. A bad selection leaves the previous configuration unchanged. File access uses the same authority boundary as the Agent’s other file operations rather than a privileged path around it.

The path is configuration; the text stays in the file. Every ordinary provider request rereads it, including subsequent requests within the same tool loop. If the Agent updates its own goal, the next request sees the update. Skills, History, and the GUI do not maintain competing copies of the objective.

## Reminders off still means instructions on

The configured goal is projected into system-side instructions even when idle reminders are disabled. On Anthropic it follows existing system blocks; chat-completions-style adapters add a system message. Responses uses the top-level `instructions` field. This is request rendering, not another History event that compression must preserve.

Auxiliary compressor, judge, and embedding requests do not receive the parent goal. They have bounded jobs of their own, not an invitation to pursue the surrounding task. Synthetic reminder text is also marked so AutoRecall does not search on it as though it were a new human question.

Disabling reminders retains the file and its instructions. Clearing the file removes configuration and injection. Unloading the skill also clears configuration, but neither operation deletes the human’s file. Standalone configuration is session-local. Registered children restore their own configuration under current authority, with a fresh failure streak; no child inherits its parent’s goal.

If a configured file becomes unreadable, Goals disables reminders and reports a stable error. Human work may continue with an explanation; an automatic wake cannot infer without the objective that justified waking it. The read carries the configuration epoch, so a late failure for an old file cannot disable a newer goal. Off and Clear remain available even when the file is missing. Repairing the file does not secretly restart automation—the user or Agent must explicitly enable reminders again.

## Idle queues mail; failures back off

Normal return to idle immediately queues one generation-tagged wake, without a timer. Failed model turns instead wait 1, 10, 100, then 1,000 seconds between subsequent attempts. Success resets the failure streak. Only this backoff uses a timer, and its callback merely posts wake mail. Configuration and lifecycle changes invalidate old wake generations.

At the actor mailbox, the wake must still be current, enabled, granted, and idle. Active foreground work, admitted turns, controls, startup work, and direct Engine sends prevent it. A pending human turn wins over a timer that has not yet been admitted. Once the wake is admitted, it is an ordinary serialized Agent turn rather than a second loop beside the human’s work.

Independent Jobs may remain alive while the actor is idle. The reminder therefore tells the model to inspect current state, continue already authorized unfinished work, leave healthy background work alone, and disable reminders when done or blocked. An expired interval says nothing about whether a quiet child or process has failed.

The automatic turn supplies the reminder, while the goal arrives separately through the normal instructions projection. If the model does another piece of work and remains enabled, becoming idle again queues continuation. Status queries do not create duplicate idle transitions. This can continue a long task without requiring a scheduler service, a goal database, or a second model conversation.

It is still an opportunity, not a progress detector. An idle transition cannot diagnose an active stuck request, recognize task completion reliably, or expand provider, deployment, device, account, or retry authorization. Those decisions remain part of the Agent’s work under the human’s original boundaries.

Retries continue from recorded state; they do not replay the failed turn’s tool effects. A delivered tool error normally remains model-handled, while unavailable goal authority, an unreadable file, or failed History persistence requires repair. Legacy `idle_seconds` values remain validated for compatibility but no longer set scheduling policy.

## Stopping has precise meanings

Pause suspends reminders and retries without clearing the goal or enabled state. Resume reevaluates eligibility, immediately for ordinary idle continuation or with the current failure delay. Retry now bypasses one wait but cannot enable an off goal or override pause. Explicit Interrupt disables reminders even if no request is active; otherwise stopping a turn could merely postpone its automatic replacement.

Shutdown and actor failure disable reminders and invalidate pending wakes. Unload and destruction invalidate queued wake mail too. A delegated Agent’s successful submission also disarms reminders and invalidates pending wakes. Its retained goal may remain visible as instructions, but finishing a delegated task cannot be followed by a surprise automatic turn that reopens it.

That last boundary matters in a swarm. Submission is an observable completion transition for the parent, not merely a quiet period on the child’s timer. Parent coordination would become unreliable if a completed child could wake itself again from an earlier reminder schedule.

## Show what the selected Agent accepted

The GUI’s Goals controls follow the selected Agent. Apply, Retry now, Reminders off, Clear, and Reload wait for actor acknowledgment; a changed checkbox is not evidence that configuration succeeded. The browser displays unloaded, unconfigured, working, paused, off, file-error, or waiting state as appropriate, including the safe failure explanation, failure count, and next retry time when known.

Status refreshes preserve unsaved draft edits rather than overwriting them, and Pause/Resume remains separate from disabling or clearing. A reconnect or conversation switch clears stale display state until the new snapshot arrives. The browser does not edit the goal file, maintain its own reminder timer, or persist this session-local configuration in appearance or memory preferences.

Goals closes a modest but recurring gap: the distance between finishing a response and finishing an authorized job. The objective stays current through compression, and idle time can produce a new chance to act. When the work is complete or genuinely blocked, disabling reminders is the control action; explaining that state to the human is still the Agent’s responsibility.

[Implementation chapter](../chapter-24.md)
