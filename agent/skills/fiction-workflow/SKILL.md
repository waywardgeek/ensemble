---
name: fiction-workflow
description: Continue a configured supplied-outline fiction book with bounded editorial rounds and supervised recovery.
type: loadable
tools: continue_fiction_book spawn_sub_agent ask_agent get_submission agents_status shutdown_agent wait_for_job
loadable-skills: fiction-author fiction-editor fiction-reviewer
mcp_servers:
  - name: fiction-workflow
    transport: stdio
    command: /usr/bin/python3
    args: ["-u", "skills/fiction-workflow/server.py", "fiction-workflow-setup.json"]
---

# Fiction workflow

This skill runs the configured book, not an arbitrary model-supplied script. Its
server starts when loaded; writing starts only when `continue_fiction_book` is
called. The host must first prepare the sandbox layout described in README.md.
The process arguments are literal relative paths, not shell or skill-variable
expansions. The `skills` path may be a host-prepared link to explicitly approved
read-only shared code; a link without process permission grants nothing. Native
file-tool access remains confined. Do not load this packaged launcher in an
unrelated working directory or edit shared code during an active run.

`continue_fiction_book` takes no arguments. It checks pinned native tool input
schemas before opening the run ledger. It continues ordinary accepted steps and
returns `complete`, `observe`, or `supervise`. Reopening the same configured run
reconciles saved assignments; do not delete pending state to force a retry.

The Author and Editor get at most three attempts per chapter. Approval may stop
earlier; reaching the limit is not approval. Three book-wide review/revision
rounds follow initial drafting. Complete output is `final-book/manuscript.md`
under the configured run directory, not a pre-revision review snapshot.

## Creative revision

The original outline and voice are pinned inputs. The driver seeds a separate
`working-outline.md`, which Authors may improve while preserving the brief and
established story facts. Editor-approved earlier-chapter repairs use typed requests
and permissions within the existing three-attempt loop; the next Editor reads all
changed artifacts. Permission is not approval. Frozen review rounds include their
own outline and voice copies. Out-of-scope changes refuse acceptance and preserve
pending work for supervision; validation is not filesystem confinement or rollback.
New plans use version 2. Older plans are not silently migrated or restarted.

## Supervision

Children are owned by the Agent making this call and can ask that Agent questions.
Answer through ordinary parent-message controls. Retain any running workflow Job
handle and observe that Job rather than calling the workflow again while it runs.
Use interruptible observations, with a progress check roughly every twenty minutes
when work remains healthy. A question warrants attention sooner; do not hide it
behind a mandatory sleep. Observation does not cancel writing or confer new grants.

`observe` means an existing assignment is active, not that it should be dispatched
again. `supervise` means inspect the named child, saved result, files and pending
assignment before deciding how to recover. A lost reply does not prove rollback.
The workflow does not restart a stopped application or resolve ambiguous admission
by guessing. Stop for missing authority or insufficient whole-input capacity.

Role memory belongs to its role, never the supervising Agent. Initial MEMORY.md
files are empty by default. A human-authorized imported-memory experiment may
use the explicit baseline marker documented in README.md; retain donor attribution
and independent role copies. Fresh chapter conversations retain only that role's transferred
memory after the old child has been joined. Reviewer tasks require a complete
single read of their assigned text; do not substitute chunks or keep_tool_results.
