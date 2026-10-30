# One-shot recovery support (candidate, not authorization to switch)

This controller is separate from the GUI. It uses Python 3's standard library,
macOS `ps` and `lsof`; it neither installs software nor reads credentials. Keep
its explicit plan, status and logs in a private trial directory. Do not point it
at the working application until Bill authorizes the actual switch and supplies
the known-good original launch arguments. Do not copy a private launcher as an
answer key. Begin with a parallel candidate on an unused port and separate data.

A plan is JSON with `port`, `original_pid`, `original_argv`, `candidate_argv`,
`cwd`, and `state` (a new directory). Both commands are argument arrays, not shell
strings. `seconds` defaults to 1800; `startup_seconds` and `shutdown_seconds`
default to 20 and 10. These bound candidate/recovery observation, not inference.
Never put credential contents in the plan. Use only explicitly approved paths.

For an authorized macOS trial, start `recover.py PLAN` from a separate temporary
per-user launchd job, **not** from a GUI Tool Job or a shell backgrounded under
that GUI. Use `ProgramArguments` with absolute Python, controller and plan paths,
`RunAtLoad: false`, `KeepAlive: false`, `AbandonProcessGroup: true`, a private
`WorkingDirectory`, and private `StandardOutPath`/`StandardErrorPath`. Do not place
the plist in `~/Library/LaunchAgents` or install a login/global service. Bootstrap
the private plist into `gui/$(id -u)`, then explicitly `launchctl kickstart` its
unique label. An accepted launchctl command is **not** proof it ran.

1. Observe `state/status.json` reporting **armed**, with the controller PID.
   No original-process signal has been sent. Touch `state/cancel` to cancel now.
2. Only after human approval, touch `state/switch`. The controller rechecks PID
   start/executable identity and sole listener ownership, then sends TERM once.
   It never force-kills the original. A slow original shutdown keeps the
   controller watching, so a late exit cannot leave an abandoned empty port.
3. Touch `state/restore` for early restore, or let the timer expire. Candidate
   crash also triggers restoration. Candidate TERM is bounded; KILL is restricted
   to the new process group created by this controller. An unexpected owner is
   refused, not killed. Check status and private logs on every refusal.
4. **restored** means a listener in the newly started original process group was
   observed. It does not establish remote connectivity, inference, voice quality
   or restored conversation continuity. The restored process uses a new session
   and redirected files, so controller exit does not own its lifetime. The
   temporary launchd definition must use `AbandonProcessGroup` for this reason.
5. After observing recovery, boot out only this temporary launchd job and remove
   its definition when the logs are no longer needed. Never stop the restored
   application as cleanup of this controller.

The executable lifecycle test uses real disposable loopback servers, independent
process sessions and retained private logs. It tests timer expiry, early restore,
crash, launch refusal, ignored TERM, changed-owner refusal, cancellation before switch, slow
original shutdown, launching-process exit, and restored survival after controller exit. It does not
claim to verify launchd service behavior, OS/power/network survival, or an actual
working-server switch. Those require an authorized human trial.
