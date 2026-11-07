---
name: goals
type: loadable
description: Keep an editable goal in request instructions and optionally continue on idle with failure backoff.
tools: goal_control
---
Ordinary authorized work should continue without this skill or reminders. Answer
an intervening question, then resume the task if the next step remains clear.
Use this skill only when the human requests goal reminders. Loading exposes the
control; it neither configures a file nor enables reminders. Write the agreed goal
with ordinary authorized file tools; include unfinished work, completion criteria
and boundaries, never credentials. Configure file and enabled:true explicitly.

Normal return to idle queues a reminder immediately, without an idle timer. Failed
model turns retry after 1, 10, 100, then 1000 seconds repeatedly; a successful turn
resets the backoff. retry_now:true bypasses one pending wait without enabling an
off goal or overriding pause. Legacy idle_seconds is validated but ignored.
Reminders give an idle Agent another chance
to assess actual state; they are not a watchdog, deadline or progress guarantee.
Inspect owned jobs, children and results, continue authorized work, and leave
healthy work alone. Never replay commands blindly or infer failure from quiet time.
When complete or genuinely needing the human, set enabled:false and explain why.
Off does not mean completed. Off retains goal instructions; file:"" clears them.

There is one configured goal file, not a goal stack. Do not claim that replacing
it saves a prior goal or that completing it automatically pops back to one. For
an authorized task switch, preserve the suspended task, completion criteria and
next step in this Agent's queue (see `task-queue`) before replacing its goal.
Returning to it is an explicit decision within the human's authorization. Native push/pop
behavior is a proposal requiring implementation, not an existing tool operation.

The file is read freshly for every normal request, including after tools and
handoff. A file error stops automatic reminders: repair and explicitly re-enable.
Standalone configuration is session-local. Registered ordinary children restore
their own saved goal settings on host restart under current skill authority, with
a fresh failure streak; submitted or paused children do not resume work. Children do
not inherit a parent's goal. Pause suspends reminders and retry waits; resume reevaluates idle eligibility.
Interrupt disables them. Unloading
clears configuration without deleting the file. No reminder expands authority.
