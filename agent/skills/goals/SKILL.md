---
name: goals
type: loadable
description: Keep an editable goal in SYSTEM and optionally resume work after an idle delay
tools: goal_control
---
Use this skill only when the human asks to load/use goal reminders. Loading it
exposes the control; it does not create a goal or activate reminders.

Write the agreed goal in a file using the ordinary Agent-owned file tools. Include
what is unfinished, what counts as done, and relevant authorization boundaries.
Keep it concise and edit it when the goal changes. Do not put credentials in it.

Call goal_control with file and enabled:true to configure and enable reminders.
The default idle_seconds is 480 (eight minutes); a positive value changes it. With
no fields, the tool reports current state. The configured file is read into SYSTEM
for every request, including subsequent requests in the same turn. It is not a
conversation summary and survives handoff/compaction independently of dialogue.

A reminder is an automatic opportunity to assess current state, not permission
to repeat a command or infer that a quiet child/job failed. Check actual work,
continue authorized unfinished steps, and leave existing healthy work alone.
If the goal is complete or you genuinely need the human, call goal_control with
enabled:false and explain why. Reminders off is not a runtime completion verdict.

Disabling retains the goal in SYSTEM. Set file:"" to clear it. The off switch
works even when the file is unavailable. A file error stops automatic reminders;
repair it and explicitly re-enable if appropriate. Unloading this skill clears
configuration and reminders but does not delete the file. Configuration is local
to this Agent/process; children and restarted sessions are not automatically opted in.

Pause suspends reminders; resume starts a fresh delay. Explicit Interrupt disables
them. The timer only wakes an idle Agent and cannot diagnose an active stuck
request. It adds no provider, deployment, retry, device or account permission.
