---
name: code-tools
description: Edit files and run independent background jobs.
type: loadable
tools: write_file edit_file run_command wait_for_job send_input kill_job tool_limits
---
Use small, checked edits and test the changed behavior. Shell state does not persist
between commands. Running jobs are not failures: retain their handles and observe
with wait_for_job. Observation delays never cancel work. send_input sends a line;
kill_job or shutdown cancels work. tool_limits applies once to the next call.
Full output remains at output_path even when an inline report is truncated.

Enabled tools:
$TOOLS
