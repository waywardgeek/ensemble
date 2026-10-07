# Chapter 4: Stop Waiting, Keep the Job

A debugger waits for input. Chapter 3 waits for the debugger to exit.
Neither participant is doing anything wrong, and neither can make progress.
The model needs its turn back while the process is still alive.

Putting a timer around the wait solves only half the problem. The command
continues after the timer fires. If the program discards its result channel,
it has created work that nobody can inspect. A retry can now start a second
copy while the first is still changing files.

Keep the work instead. Give it a handle, retain its output, and let the model
look again, send input, or stop it. The job survives the end of the tool
call that started it.

> **Independently reviewed student contract.** No Chapter 4 implementation,
> successful grade, or new live demonstration is claimed.

## 4.1 A job needs an owner

Agent owns a Jobs service in `internal/jobs`, just as it owns its Registry
in `internal/tools`. Jobs owns each live Job. Ensemble owns an application-wide
handle allocator, starting at one and increasing for every newly allocated
job. These are explicit working choices for this edition. They provide
independent Agent access while preventing two Agents from assigning the
same handle to different work.

The CLI selects all ten tools. Public library callers still choose their
Agent's subset, and an omitted selection still enables none. Adding job
supervision does not silently enlarge a caller's chosen capability set.

The parents are interfaces declared in `internal/common`. Jobs reaches its
Agent, and Job reaches Jobs. A live Call belongs to Engine and follows
Engine's Agent parent to the Registry or Jobs service. Do not add separate
Agent, Jobs, logger, and configuration fields to Call merely because its
operations need those facilities. The chain already reaches them.

Common declares job snapshots, statuses, limits, and the interfaces. Jobs
implements lifecycle, waiting, output capture, and reporting. Tools implements
the tool operations; llm implements dispatch and model continuation. These
spokes import common and never one another. If a shared type needs behavior
in jobs, use a free function there rather than moving behavior into common
to keep method syntax. Helpers retain their owning context and logger path.

One worker owns completion. For a local function, that worker completes the
job after the function returns. For a process, its process/output lifecycle
completes the job after obtaining the exit status and draining output. The
dispatcher owns the wait for a report. Its return is not process completion.
Closing the output file at that point would cut off a reader still writing.

## TL;DR

Continue validated Chapter 3 history in `solutions/edition-2/ch04`. Read the
entire [coding skill](skills/ensemble-coding/SKILL.md), architecture, and this
contract before editing. A fresh student receives only the new curriculum
and preceding new snapshot; historical answers and research notes are excluded.

1. Add Agent-owned Jobs and the Ensemble-owned handle allocator. Every
   permitted ordinary tool execution becomes a job at the dispatch boundary,
   including local file tools. The Registry marks the four supervision tools
   below as operations that do not create jobs.
2. Persist `tool_called` with the running job before execution. Record normal
   completion as `job_ended` and killing as `job_killed`, through Agent's
   single append path. A tool result reports a snapshot; it need not wait for
   completion. Retain call/result IDs and batch order.
3. Add `wait_for_job`, `send_input`, `kill_job`, and `tool_limits`. A wait ends
   on completion, a new-output pattern, or its delay. A delay never kills work.
   Defaults are three seconds and 16384 report bytes, with no pattern.
4. Consume pending `tool_limits` on the next attempted call, whatever its
   name. Explicit call limits override pending ones. Report consumption on
   the consuming call, including invalid calls and another setter.
5. Write produced output to `cr/io/HANDLE` beneath the owning Agent's
   workspace. Reports contain bounded unseen output, with head/tail and an
   exact omission notice when capped. Preserve the file for later reads.
6. Run commands on a Unix PTY with input echo and merged stdout/stderr.
   This replaces Chapter 3's separate-stream result. Normalize CRLF, retain
   the actual exit status, and add a per-call `cwd` override. Never change
   process cwd or silently fall back after an invalid override.
7. Terminal states are monotone. Kill the process group, wake waiters, and
   prevent late completion from changing killed to done. For an arbitrary
   goroutine, report honestly that marking its job killed cannot stop it.
8. Close shuts down an Agent's jobs; application close shuts down every Agent.
   CLI EOF closes before reporting final usage. Errors still clean up managed
   processes even when their log is unwritable. Replay never resumes work.
9. Preserve public clients, optional GUI separation, model provenance,
   declarations on every continuation, and the sixteen-request turn bound.
   Retain Chapter 2's human chat and machine protocol as distinct clients.
   Add deterministic lifecycle checks and actual interactive-program runs
   on all three APIs before the independent comparison and revision gate.

**Yours.** Private names and synchronization choices, compatible report
wording, and the PTY library. The wire fields, limits, ownership, transitions,
and observable behavior below are the contract. No mailbox, model streaming,
turn interruption, process restart recovery, automatic retries, or remote-tool
connection is added.

From the course repository root:

```sh
make grade-dir CH=4 DIR=solutions/edition-2/ch04
```

## 4.2 Start once, report more than once

Resolve the attempted call's limits and Registry permission first. Unknown
or disabled tools return the Chapter 3 error result without starting a job.
A malformed limit also refuses execution. Such attempts still consume a
pending one-shot setting and record a matched call/result pair, without
inventing a job for work that was never admitted.

For a permitted ordinary tool, allocate a handle and create its output file
exclusively. Never truncate an existing artifact. Allocate before recording
`tool_called`, so that event names the job before the handler starts. If the
file cannot be created or the event cannot be persisted, do not execute.
A failed allocation may leave a gap in handles; reuse would make old records
ambiguous. In a fresh application with successful allocations, handles are
1, 2, 3, and so on across all its Agents.

Run the ordinary handler under the job lifecycle and wait for its first
report. A quick read can return its complete result immediately. A slow
command can return a running report while its worker continues. Append
one `tool_returned` for this call with the report and current job snapshot.
Subsequent `wait_for_job` calls have their own call IDs and results. They
do not append another result under the original call ID.

Dispatch accepted calls in their original batch order. One job can still
run after its initial report while the next call starts another. Once every
call in the batch has a result, continue the model conversation. A running
job report satisfies the original call's pairing; it does not claim the
job completed. The existing sixteen-request bound still counts model
requests, and its last accepted batch still receives all its results.

A local handler's missing file or bad ordinary argument produces a done
job with an error result. A shell that exits 7 produces a done job with
exit code 7 and a successful tool invocation, as before. A running report
is not an error. Killing is a separate terminal status, not a strange
successful exit code.

## 4.3 What the log records

Keep the version-1 log header and existing event envelope. Extend the known
event vocabulary explicitly. The new `tool.job` snapshot has these fields:

| Field | Contract |
|---|---|
| `handle` | Positive integer allocated by Ensemble |
| `status` | `running`, `done`, or `killed`; no zero/unknown state |
| `output` | Ref with `kind:3` and locator `cr/io/HANDLE` |
| `bytes` | Nonnegative count of normalized produced bytes in the artifact |
| `exit_code` | Present only for a normally reaped process; zero is a real value |
| `is_error` | Whether completed tool execution failed; absent means false |
| `reason` | On killed records: `kill_job` or `shutdown` |
| `cwd` | Absolute effective override for run_command; absent for the default workspace |

The artifact path resolves against the job's owning Agent workspace, even
when a particular command uses another cwd. Artifact access and supervision
are scoped to that Agent. An Agent must not wait on, send to, or kill another
Agent's handle merely because it guessed the integer. An application-wide
allocator is identity management, not a cross-Agent permission grant.

`job_ended` and `job_killed` carry a top-level `job` payload with the same
snapshot shape. The first records done and the second killed. Both retain
the original handle and locator. New live captures emit the terminal event
once; repeated waits only add the supervision call/result pair. The four
supervision operations have no `tool.job` of their own.

This compact literal fixture records a successful local job. Its produced
file contains exactly `ok\n`; a live capture also includes the corresponding
request event before the model response. The imported compact response
rule from Chapter 2 still applies.

```json
{"log_version":1}
{"seq":1,"type":"message_received","time":"2026-01-01T00:00:00Z","message":{"actor":"human","purpose":"dialogue","parts":[{"type":"text","text":"Read notes.txt."}]}}
{"seq":2,"type":"response_ended","time":"2026-01-01T00:00:01Z","response":{"from":{"vendor":"anthropic","model":"fixture-model","surface":"messages"},"parts":[{"type":"tool_call","call_id":"read-1","from":{"vendor":"anthropic","model":"fixture-model","surface":"messages"},"name":"read_file","args":{"path":"notes.txt"}}],"usage":{"input":1,"cache_write":0,"cache_read":0,"output":1}}}
{"seq":3,"type":"tool_called","time":"2026-01-01T00:00:02Z","tool":{"call_id":"read-1","name":"read_file","args":{"path":"notes.txt"},"job":{"handle":1,"status":"running","output":{"kind":3,"locator":"cr/io/1"},"bytes":0}}}
{"seq":4,"type":"job_ended","time":"2026-01-01T00:00:03Z","job":{"handle":1,"status":"done","output":{"kind":3,"locator":"cr/io/1"},"bytes":3}}
{"seq":5,"type":"tool_returned","time":"2026-01-01T00:00:04Z","tool":{"call_id":"read-1","parts":[{"type":"text","text":"ok\n"}],"job":{"handle":1,"status":"done","output":{"kind":3,"locator":"cr/io/1"},"bytes":3}}}
```

Validate job records before persistence and during replay. A terminal event
must refer to an existing running job. Reject a duplicate terminal event,
a changed locator, decreasing byte count, an exit code on running/killed,
or a killed-to-done transition. A tool result's job snapshot must belong to
its original ordinary call and agree with the known terminal state, if any.
Only job-creation records introduce handles. A local function needs no
invented exit code.

Commit a report's state and cursor consistently with terminal publication.
A completion racing the report must appear either before that report, which
then reflects done, or afterward, when a running report was still accurate.
Do not capture running, append job_ended, then publish the stale running
snapshot as if the job moved backward. Serialize that decision with event
append, without holding a worker lock while waiting for a long operation.

The reducer retains job facts for inspection and observer delivery. It
does not turn an imported running record into a goroutine or an attached
process. Offline dump/render must work after the output file was removed;
these operations reproduce recorded text and metadata without opening the
locator. A model later asking to read a missing artifact gets a normal tool
error. The reference is a recovery address, not a promise of immortal bytes.

Keep Agent's event append serialized independently of a long-running prompt
or wait. A worker completing while its dispatcher waits must be able to
append the terminal event. Holding the turn mutex across that append path
would turn the new job system into a deadlock. No observer sees an event
before persistence, and each receives its own owned snapshot in sequence.

Background events also change when a response's sequence becomes known.
A job can finish while the HTTP request is in flight, consuming the next
event sequence before the response arrives. Chapter 2's synthesized missing
call IDs must use the actual committed response sequence, not a number
predicted before HTTP. Finalize missing IDs on an owned response copy under
Agent's append serialization, after choosing that event's sequence and
before validating and writing it. Preserve supplied IDs and the specified
part index. Leave the parser's returned facts unchanged. Do not suppress
real-time job observations for an entire HTTP request to reserve a sequence.
Replay then sees the same finalized IDs that the model continuation used.

## 4.4 Ask when to look again

The default wait is deliberately short because it stops nothing. A model
can look after 0.2 seconds, wait longer on the next call, or wait for a
prompt such as `(dlv) `. It is choosing when to regain control, not how
long the operation is allowed to exist.

| Argument | Type and meaning |
|---|---|
| `ai_callback_delay` | Finite nonnegative number of seconds; default 3; zero polls without sleeping |
| `ai_callback_pattern` | Go regular expression; default absent; empty string explicitly clears it |
| `max_output_bytes` | Positive integer; default 16384; budget for retained report content |

Reject an invalid regex, wrong type, negative delay, duration overflow,
or nonpositive byte budget before executing the affected operation. Compile
the regex at this boundary. There is no separate upper byte-budget policy
in this chapter; the model's explicit positive budget remains observable.

These arguments belong directly on `run_command`, `wait_for_job`, and
`send_input`. Other ordinary tools keep their existing argument schemas.
The setter reaches those tools without making every external schema grow
three monitoring fields.

| Tool | Required arguments | Other arguments |
|---|---|---|
| `wait_for_job` | `handle`: positive integer | The three wait/report arguments |
| `send_input` | `handle`: positive integer, `input`: string | `append_newline`: boolean, default true; the three wait/report arguments |
| `kill_job` | `handle`: positive integer | None |
| `tool_limits` | At least one of the three limit fields | Only those fields |

Reject unknown fields and unavailable handles with actionable tool errors.
`send_input` appends one LF by default, as pressing Enter would. With
`append_newline:false`, it writes exactly the supplied bytes. The default
adds one LF even if the input already ends in a newline; select the false
option when supplying complete input yourself. Empty input is valid: it
sends a blank line by default, or only waits/reports when append_newline
is false. Sending to a finished job or a job
without an input-capable process is an error. Waiting on a finished job
returns immediately, including when a previous wait already consumed all
its output. A successful wait with no new bytes still reports its status.

Jobs owns the pending one-shot limit state for its Agent. To resolve each
attempt, take and clear that state, overlay it on the defaults, then overlay
the call's own legal limit arguments. A bad or unavailable attempted call
still consumed the pending state. A malformed model response never created
an accepted call and therefore consumes nothing.

A valid `tool_limits` operation stores only its supplied overrides for the
next call. A second setter consumes the old overrides before storing its
new ones. An invalid setter consumes any previous setting but stores nothing.
The consuming result begins with a line naming `tool_limits`, the consuming
tool, and the effective settings or the validation refusal. The setter's own
success message explains this next-call rule. No consumption means no note.

For example, a setter with delay 0.2 followed by `kill_job` spends the
setting on the kill operation. The next `run_command` gets the default
three seconds. A setter with delay 0.2 followed by a command explicitly
requesting delay 10 waits with 10. These are separate positive and negative
controls; a sticky setting can pass the first example's initial call.

## 4.5 Retain the output, spend fewer tokens

The job artifact holds the full produced result. The conversation holds a
bounded report of what the model has not yet consumed. Keeping these separate
lets a test suite produce a large failure log without inserting that entire
log into every subsequent model request.

For file tools, Chapter 3's selection limits still apply: a range read or
`max_bytes` controls what that tool produces. Its job retains all that result.
It does not secretly read the rest of a file merely because storage is now
on disk. A full-file read large enough to exceed the report budget must
explicitly request enough source bytes. For read_file artifacts, preserve
the selected file text itself; put optional presentation headers in reports.

The shell changes here. Spool its complete merged stream and final status
to disk. Its `max_output_bytes` now caps the inline report, replacing
Chapter 3's per-stream capture limit. Continue reading after the report
budget is reached. The artifact is the retained output; do not maintain a
second unbounded in-memory copy merely to serve reports.

Every job has one report cursor, initially zero. At a report, capture the
current end offset and consume the interval from the cursor to that offset.
If it fits, return those bytes. Otherwise retain the first ceiling-half and
last floor-half of the budget and insert a notice with the omitted byte
count, current total byte count, and locator. Metadata and the notice are
outside the content budget. A complete interval exactly at the limit is
not truncated. Advance the cursor to the captured end even when the middle
was represented by a notice; those bytes remain recoverable from the file.

Retained head and tail must contain complete UTF-8 code points for valid
text. Trim an incomplete code point from either cut boundary rather than
letting JSON introduce a replacement character. Count the actually omitted
bytes in the notice; unused budget is preferable to corrupting the output.
This extends Chapter 3's supported-text cap rule to a head-and-tail report.

With budget 4, the bytes `abcdefghij` produce `ab`, a notice that six of
ten bytes were omitted with the locator, and `ij`. The next report contains
no copy of those ten bytes. If the worker then appends `kl`, the following
report contains just `kl` plus necessary job metadata. Test the positive
content and the absence of already-consumed bytes separately.

A pattern is matched against new eligible output, including a match split
across reader chunks. `wait_for_job` starts at the current report cursor.
For `send_input`, the pattern starts at the output position captured just
before writing the new input, so an old prompt cannot answer a new keystroke.
Its report still includes every unconsumed byte. Completion or kill wakes
waiters regardless of the pattern; delay is measured for this wait only.

Reports must describe the state they observed. For a first report that
finishes normally within the delay, fits the content budget, and needs no
consumption/cwd note, bare tool text is sufficient. Other reports identify
the handle, state, locator, and total bytes. Running reports offer the next
operation. Killed reports say killed; their generated status metadata makes
no normal-completion claim and supplies no exit code. Retained child output
may itself contain words such as `done` or `exit_code:7`; preserve those
bytes and distinguish them from the program's metadata. A killed status can
be returned successfully by wait_for_job;
the observation itself is not a failed tool invocation.

The same distinction applies to a job whose handler failed: its original
execution result and `job.is_error` retain that failure. A valid later
wait_for_job returns `tool.is_error:false`, because observing the job
succeeded, while its report explicitly describes the job's failure and
retained error text. An invalid handle or failed supervision operation
still returns a tool error.

## 4.6 Give the process a terminal

Interactive programs can behave differently on pipes. The exercise uses
a Unix pseudo-terminal so a debugger receives a terminal, input is echoed,
and the model can recognize its prompt. A PTY merges stdout and stderr;
the Chapter 3 promise of separate streams no longer applies to run_command.
Keep this change explicit in both the schema description and user-facing
result documentation.

Start a fresh POSIX shell for each command, with its own process session/
group and controlling terminal. Use `TERM=dumb`, a 50-row by 200-column
terminal, and normalize CRLF to LF while preserving lone carriage returns.
Handle a CRLF pair split between reads. A normal process's artifact ends
with `exit_code: N` on its own line, after all its output, including the
actual nonzero status. Failure to start the PTY/process is a tool error;
there is no silent fallback to pipes.

The [creack/pty reference](https://pkg.go.dev/github.com/creack/pty), checked
October 7, 2026, documents starting a process with a controlling terminal
and a new session. Its sample code is explicitly illustrative. Choose and
test the actual reader-close behavior; do not assume copying the sample
provides the required lifecycle. Record the chosen dependency version in
the new module rather than copying an old solution's implementation.

Add optional `cwd` to run_command. Resolve a relative value against Agent's
captured workspace; accept an absolute directory. Reject an empty explicit
value, a missing directory, or a non-directory before starting the process.
Record the resolved override in the job and show it in the report. With no
override, leave `cwd` absent from the snapshot and use the Agent workspace.
The next call starts there again. A working directory is not confinement.

A persistent shell is still possible as explicit work: start a shell as a
job and send further commands to its handle. That shell's state belongs to
the job. Separate run_command calls never inherit its `cd` or environment
changes. Two simultaneous jobs must not share a hidden shell session.

## 4.7 Finish once, including when it is killed

A terminal transition wins once under synchronization. Normal completion
publishes done; a successful kill publishes killed. A late worker cannot
replace killed with done, append a fabricated successful exit status, or
write a late local-handler result into a closed artifact.

`kill_job` sends SIGKILL to the process group, not just the shell PID, and
wakes waiters. A kill racing with startup must either prevent launch or
kill the newly started child once it becomes known. A committed done job
stays done: killing it reports that it already finished and writes no
`job_killed` event. Killing an already-killed job is similarly idempotent.
An unavailable or other-Agent handle remains an error.

After killing a managed process, reap it and join its output reader before
closing the spool. Preserve output already captured. Terminal publication
must not race a writer still appending normal output. No grace-period
protocol is required in this chapter; the chosen operation is explicit kill.
Do not turn a wait delay into an implicit kill timer.

For an ordinary Go function, the runtime offers no general goroutine-kill
operation. Mark its job killed, wake waiters, and report that the function
may continue and may still have side effects. Discard its eventual result.
Shutdown must not wait forever trying to join an uninterruptible function.
There is no blanket recover around handlers: an invariant panic remains
a process failure rather than a made-up normal tool result.

Agent close refuses new operations, kills its running jobs, and records
`job_killed` with reason `shutdown`. Ensemble close closes every Agent.
Both operations are idempotent. CLI clean EOF performs this cleanup before
printing final usage; an error path also cleans up but retains its nonzero
exit and no-success-usage contract. Ending one model turn or returning a
library request error does not itself close the Agent, so its jobs remain
available until explicit close.

Persistence still constrains execution. A failed pre-dispatch record prevents
the handler from starting. If a result or terminal event cannot be recorded,
fault the Agent and surface the persistence failure without pretending the
physical work was rolled back. Wake blocked waiters with the failure. A spool
write failure likewise prevents a complete-output claim and faults further
operation. Best-effort process cleanup must still run; inability to append
a shutdown record is not a reason to leave a child process alive. Report
that the cleanup record could not be retained through the owned logger.

## 4.8 Checks that distinguish the lifecycle

The old killed-as-done mutation initially passed. A kill note contained the
word “killed” while the job's actual state said done. Commit `d08a92b`
records the fix: inspect structured status and forbid a generated
normal-completion claim on the killed report. Also plant child output
containing `done` and `exit_code:7`: preserving that output must pass while
generated killed metadata and the recorded status remain correct. A regex
over the complete report cannot distinguish the two sources of text.

Use fixtures that produce observable boundaries. Build the helper binaries
before measuring their runtime. A short program prints a marker and exits;
a slow one prints a marker, waits beyond the wake delay, then exits; a prompt
program waits for input and answers after a controlled delay. A blocker
stays alive until killed, and writes its PID so the check can verify its fate.
Use a child process in the group to distinguish group kill from shell-only
kill. Never rely on a compiler cache warming fast enough to fit a tool wake.

The prompt fixture should print the same ready marker before and after an
input. Delay its reply enough that a matcher reusing the old marker returns
observably too early. Commit `fcf9a2b` added this distinction and removed
parallel mutation runs that starved each other's subprocesses. Keep timing
margins explicit and run resource-sensitive audits without that contention.

The inherited grader retains these nine checks, totaling 100:

| Check | Points | Behavior |
|---|---:|---|
| `ch3parity` | 10 | Six original tools and conversation loop still work |
| `jobmodel` | 25 | Dispatch allocates every ordinary job; output and cwd are observable |
| `waitjob` | 10 | Delay returns running; completion and already-finished wait work |
| `sendinput` | 10 | Input reaches the process and only new output satisfies the pattern |
| `debugger` | 5 | Actual debugger reaches a breakpoint and reads a known value |
| `killjob` | 10 | Process stops, waiters wake, status remains killed |
| `bigoutput` | 15 | Full produced output remains on disk while reports are bounded |
| `toollimits` | 10 | Next-call consumption, precedence, and visible notes |
| `shutdown` | 5 | Exit cleans up remaining jobs and records its reason |

The new contract additionally requires:

| Property | Independent positive and failure controls |
|---|---|
| Ownership | Actual parent chains and logger access through jobs/tools/llm; no sibling imports or mutable global allocator |
| Agent isolation | Distinct application handles; shared-workspace artifacts do not collide; foreign handles refuse supervision |
| Durability | No execution after failed pre-call write; terminal-write/spool failure surfaces; cleanup still stops managed children |
| Completion | Late output survives normal completion; one terminal event; kill-vs-exit/start races; repeat wait/kill; no killed-to-done transition |
| Output | Exact cursor advancement including omitted bytes; head/tail recovery; exact-cap control; split pattern and CRLF boundaries |
| Settings | Pending state per Agent; consumption by unknown/invalid/inline calls; second setter; explicit override; absence of spurious note |
| Replay | New event validation, stable locator and byte count, no artifact reads or subprocess effects on dump/render |
| Response identity | A job completion during HTTP consumes a sequence; missing call IDs use the later actual response sequence; parser facts and supplied IDs remain unchanged |
| Clients | Background terminal events use the same public observation path; optional GUI stays outside the headless library |
| Real use | All introduced operations and recovery exercised through actual program/model paths |

Preserve first-edition checks as a regression floor. Reconcile the explicit
PTY change with the new Chapter 3 separate-stream tests; do not quietly
claim both incompatible representations. Keep positive controls for short
successful commands and byte limits, and confirm each deliberate mutation
actually changes behavior rather than merely failing to compile.

The debugger is a prerequisite, not an optional skipped check. Follow the
[Delve installation instructions](https://github.com/go-delve/delve/blob/master/Documentation/installation/README.md),
checked October 7, 2026, and put `dlv` on PATH. The documented installation
command is:

```sh
go install github.com/go-delve/delve/cmd/dlv@latest
```

Record the installed version used for evidence. A missing debugger or PTY
support is a clear failed prerequisite, not a successful empty demonstration.

## 4.9 Taking it for a spin

Build the new CLI and run `chat` in an actual terminal/PTY, using a scratch
workspace with a fresh explicit log. Type ordinary requests, wait for the
visible prompt, and inspect readable answers before choosing follow-ups.
The person talks to Ensemble through chat; the model operates a separate
PTY through its tools. Demonstrating the latter does not prove the former.
Keep credentials in the environment. Select a currently available tool-capable
model for each API, including any known resolved identity required for opaque
replay. Preserve requested/returned identities and usage as before.

Ask the model to start a slow command, inspect another file while it runs,
and wait for the result. Then start an interactive program, wait for its
prompt, send input using the default Enter behavior, and verify its reply.
Also exercise exact input with `append_newline:false`.
Start a persistent shell job, change its directory through input, and verify
that a separate command still uses the Agent workspace.

Drive a real debugger through the same program. Plant a small Go program
with a known variable, ask the model to start `dlv`, wait for its prompt,
set a breakpoint, continue, print the variable, and quit. Preserve the actual
tool records and output artifact. A model saying “42” without debugger output
showing the inspected value does not prove the demonstration happened.

Exercise explicit cwd and a missing-directory refusal, a capped result with
range recovery from its artifact, one-shot limit consumption, an explicit
kill, and a still-running job cleaned up on EOF. Repeat the feature checklist
through human chat on all three model paths. Exercise `/history`, a valid
`/redact` target, `/usage`, and clean EOF while a job is running. Inspect
files and process state independently of
the model's prose. Public-consumer runs cover independent Agents and their
ordered observations; fault fixtures cover races and persistence failures
that cannot be ordered reliably through a live model.

Retain the first student answer before comparative review. The independent
reviewer then reads the old standard for useful comments, simpler lifecycle
code, and missed contracts. The student receives findings and rationale,
without reading the old source, and revises from the new teaching. The author
incorporates the teaching findings. Keep the corrected ownership design and
recheck affected behavior before claiming the chapter complete.

[LIVE RECEIPTS PENDING: new Chapter 4 program, all-three-provider feature
runs, actual interactive debugger, output recovery, shutdown, public consumer,
measured usage, independent acceptance and comparative review.]
