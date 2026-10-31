---
name: goals
type: loadable
description: Keep an editable goal in request instructions and optionally resume after an idle delay.
tools: goal_control
---
Use this skill only when the human requests goal reminders. Loading exposes the
control; it neither configures a file nor enables reminders. Write the agreed goal
with ordinary authorized file tools; include unfinished work, completion criteria
and boundaries, never credentials. Configure file and enabled:true explicitly.

The default idle delay is 480 seconds. Reminders give an idle Agent another chance
to assess actual state; they are not a watchdog, deadline or progress guarantee.
Inspect owned jobs, children and results, continue authorized work, and leave
healthy work alone. Never replay commands blindly or infer failure from quiet time.
When complete or genuinely needing the human, set enabled:false and explain why.
Off does not mean completed. Off retains goal instructions; file:"" clears them.

The file is read freshly for every normal request, including after tools and
handoff. A file error stops automatic reminders: repair and explicitly re-enable.
Configuration is session-local, not restored or inherited by children. Pause
suspends reminders; resume starts a fresh delay. Interrupt disables them. Unloading
clears configuration without deleting the file. No reminder expands authority.
