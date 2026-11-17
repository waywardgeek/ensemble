# Shell-first Flint student

Bill requested a smaller tool surface after reporting Flint's complaint about too
many tools. Terminus2 and mini-SWE-agent evaluation interfaces use terminal/shell
actions; their use for evaluation does not prove Flint's training interface.

The student primary drops file-reading and explicit write_file/edit_file grants.
Its eight effective tools are run_command, wait_for_job, agent_status,
send_message_to_parent, submit, load_skill, unload_skill and micro_handoff.
The last three are universal runtime controls. Coding stays active, and loading
ensemble-coding adds instructions without widening tools. The read-only reviewer
is unchanged. No runtime code or sandbox authority changes.

File inspection, search and changes now use shell commands. Each call has a fresh
shell and sources the supplied environment recipe. Unlike Terminus's persistent
tmux session, this is still Ensemble's run_command/Job protocol, not a replica of
that harness. Retain and observe running jobs. Shell redirection can overwrite
files: the smaller vocabulary trades explicit file-tool guards for deliberate
target inspection and checked replacement scripts. Do not relax confinement.

The old skill was backed up and the tested copy installed with Bill's approval.
SHA256: cde27337e6387c586276a7e3f2ceeb2c8b728bf244b2d2f11cf7917731a9e299.
Both real-loader modes passed; adding file-reading back to a disposable copy was
rejected. These checks establish the tool surface, not improved Flint performance.
No fresh student was launched and no running student's active skills/settings
were modified in this update. Current Chapter2 trial retains its admitted setup.

Inspection also found a stale “No Chapter2 work” sentence in the chapter-generic
primary supplied to that trial. It conflicted with its explicit Chapter2 task.
The new text says “No work beyond the assigned chapter.” The conflict is a setup
confound; do not attribute possible effects to model capability alone. It was not
silently corrected in the active trial.
