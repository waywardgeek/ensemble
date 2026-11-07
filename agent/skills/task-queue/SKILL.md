---
name: task-queue
description: Capture ideas in this Agent's own queue without derailing the current task.
type: dependency
---
# Per-Agent task queue

This Agent's data directory is: $AGENT_DATA_DIRECTORY
Keep its task queue in `queue.md` directly inside that directory. This is not the
shell HOME or current working directory, a project-wide backlog, or another
Agent's queue. If the directory is UNAVAILABLE, do not invent a path: ask for an
explicit persistent location before saving. A path describes storage; it does not
grant file access. This skill grants no tools. Use only already-authorized file
tools, or obtain an available file-tool skill through the normal skill controls.

When the human says “add XYZ to the queue” during other work:
- Read the existing queue, or create it if it is absent. Preserve other entries.
- Record the idea concisely, with the human's reason and important constraints.
  Mark it queued or proposed, not active or completed. Resolve only ambiguities
  that materially affect what is recorded; do not start a design investigation.
- Verify the write, give a brief confirmation, then resume the interrupted task.
  If recording fails or tools are unavailable, say it was not saved; do not imply
  that conversation alone is a durable queue entry.

Use a small Markdown document with readable headings and bullets. Do not impose
an elaborate ticket schema or expand a short idea into an acceptance campaign.
When the human chooses a task or work is actually completed, update its status
and preserve useful reasons. Avoid duplicates and never discard unfinished work
merely because the current task or workspace changes.

Queueing is capture, not a priority change, execution authorization, goal switch,
or request to enable reminders. Do not scan and execute the backlog after every
interruption. Consult it when establishing priorities or resuming suspended work.
The active goal may refer to a selected queue item, but the queue is not a goal
stack and completion does not automatically pop or launch another task. Keep
secrets and private transcripts out of entries; do not commit or publish this
Agent-specific file with a project unless the human explicitly requests it.
