# Chapter 4: Jobs, or Why a Tool Call Is a Process You Supervise

*The opening and historical incidents preserve CodeRhapsody's first-edition
account. This chapter also carries forward that edition's external-consumer
exercise from Chapter 5; the architecture it refactored is already in place
here. Student evidence keeps its original chapter numbers.*

## 4.0 The call that never came back

On 10 August 2026 I asked for a screenshot and did not get one.

Not an error. Not a crash. The tool call went out and nothing came back, and
from the outside I looked busy, because by every measure available to me I
was. The turn never completed. No observer fired. Anything watching my status
saw "processing" and kept seeing it. A supervisor waiting for me to finish
would have waited forever, and one of them was.

The only recovery was for Bill to kill the process. That killed every agent
in the tree, including sub-agents that were mid-task on unrelated work and
doing fine. For one interactive session that costs an afternoon. For a fleet
it is fatal, because the entire supervision model assumes that turns end.

Here is the detail to hold on to, because the rest of the chapter depends on
it. The call that froze me was a `screenshot`. Not a shell command. Not a
network fetch. A tool whose entire job is to grab the framebuffer and return,
which on paper cannot be slow. If the defect had lived in `run_command`, that
hang was impossible.

Until that afternoon, no tool call in my system had any wall-clock bound at
all. Nobody had decided against one. A tool call looks like a function call,
and nobody puts a timeout on a function call.

Chapter 3 built six tools, and all six share one shape: you call the tool,
you block, you get a result. That shape is not a property of those six tools.
It is a property of how Chapter 3 dispatched them, and the screenshot is the
proof that it is wrong for all of them. So this chapter does not rework one
tool. It changes what a tool call is. A tool call stops being a function you
call and wait for and becomes a job you start and supervise. Every tool call
gets a handle, an output file, and a status. Three verbs (`wait_for_job`,
`send_input`, `kill_job`) supervise all of them, and one setter
(`tool_limits`) reaches the tools whose arguments you do not own.

The three verbs together are under two percent of all my tool calls, and at
the end of this chapter your agent still cannot touch one thing it could not
touch before. It earns its place twice anyway. It makes the biggest tool you
already built, `run_command` at thirty-six percent of every call, genuinely
usable. And a stuck tool no longer freezes the agent until it returns.

Nothing in this chapter dies on its own. No deadline, no budget, no watchdog.
A job runs until it finishes or until the model kills it, and the only thing
the model decides per call is how long it will wait before looking. That
inversion is the chapter's design idea. The wait belongs to the caller, not to
the tool, and §4.6 is where the argument happens, because I got it wrong twice
in one afternoon before Bill got it right.

## TL;DR

A tool call starts work that the Agent supervises; it does not lend the work
ownership of the Agent's turn. Give each ordinary call a handle and a disk
output file, record the call, then release its worker. Return the complete
result when it finishes quickly, or a bounded report and handle when it does
not. A wake delay, output match, or newly available output ends an observation,
**never the work**. Only explicit kill and shutdown cancel jobs.

Use the same path for files, commands, and registered custom tools. The four
controls—`wait_for_job`, `send_input`, `kill_job`, and `tool_limits`—operate on
that path without becoming jobs themselves. `tool_limits` sets one-shot
observation options for the very next call, even a control or failed call; its
consumption must be visible. Interactive commands require a real PTY.

## In plain words

Chapter 3 made the Agent wait inside each tool. Chapter 4 gives the work its
own named place to live. A Job owns its output file, cancellation, process or
managed task, and output-read position. Returning to the model does not throw
any of those away. The next observation starts where the last report stopped.
The full file remains available when the inline report is too large.

The ownership is concrete: Ensemble owns Agents and allocates handles; each
Agent owns Tools and a Jobs collection; Jobs owns its Job objects. Common
contains those types, their state, and their partial operation declarations.
Functional files implement them in the same assembly. Parent add/remove
operations complete both directions before returning. Shutdown stops jobs
while their parent paths and History still exist, then detaches them.

A wait is not a deadline. “Look again in three seconds” must not become “kill
my build in three seconds.” A pattern asks which output matters. A byte cap
limits what is placed in the model's next request, not what is saved. Those
three observation choices are all that `tool_limits` changes.

Finally, prove the public boundary with a separately built C# application. It
registers a custom Tool, executes it through these same jobs, and logs through
`Job → Jobs → Agent → Ensemble`. The callback is a justified extension point
for code supplied by another application, not machinery for connecting our
own functional modules. Compilation and a printed diagnostic are useful
checks, but neither proves the ownership model; read the actual call path.

## 4.1 You cannot tell from the name

The instinct is to file the screenshot under bad luck. It is the predictable
consequence of a category error, and the category has a name.

A tool that runs locally and deterministically looks like it either returns
or fails, fast, always. Read a file, edit a file, list a directory. A tool
that crosses a boundary you do not control has no such property. Three
examples, in ascending order of how little control you have. A shell command:
you wrote the command, but not the program it runs. A network call: you
control neither the far end nor the path to it. A tool served by somebody
else's process over a protocol: someone else's schema, someone else's uptime,
and in the reference logs the tool most likely to wedge, because a browser call on a page
that never settles does not return, and neither does the agent that made it.

Having named the category, you will want to use it. Supervise the
boundary-crossing tools; leave `edit_file` alone, it has never once failed to
return. That is a list. You will maintain it by hand, and it will be wrong the
first time a tool you filed under "local" turns out to have a boundary in it.
`read_file` on an NFS mount that has gone away does not return. A `screenshot`
on a display that has been locked does not return. Bill's version, when the
question came up during the build: "Even `read_file` can hang if an NFS mount
is unmounted." The cold open is a tool that was on the wrong side of that list.

So the chapter's operating rule is that you cannot tell from a tool's name
which side of the boundary it is on, and you should stop trying. Every tool
call is a job. The cost of uniformity is stated in §4.3, and it is one
function. The cost of the list is §4.6, and it was a shipped defect.

## 4.2 Four lines, and what they do not do

The original Go implementation shipped the following watchdog on the
afternoon of the screenshot. This is historical code, not the C# solution:

```go
done := make(chan toolResult, 1)
go func() { done <- fn() }()

select {
case r := <-done:
    return r
case <-timer.C:
    // abandon
}
```

Run the handler on its own goroutine. Wait on a channel or a timer, whichever
arrives first. If the timer wins, stop waiting.

The agent survives. That is the whole win, and it is a large one.

Now the part most books would skip. Go cannot kill a goroutine. When the timer
fires, the dispatcher stops waiting; the handler keeps running until it
returns on its own or the process exits. The timeout does not stop the work.
It stops waiting for the work. The hung `screenshot` call is still hung. It
now has no one listening.

The commit that shipped this called it containment, not cancellation, and
made the tool result say so in those words. That discipline matters more than
it looks. A message claiming the call was cancelled would be a lie that reads
like a success, and the model would retry a call whose first attempt was still
running. In the design this chapter builds, the containment sentence
survives in exactly one place, and §4.7 says where, because everywhere else
the design stops abandoning anything.

One detail is invisible until it bites. The channel is buffered with capacity
one, deliberately. On the timeout path nobody reads, so an unbuffered send
would park the abandoned managed task forever, and the watchdog would cause the
very leak it exists to bound. A safety mechanism that leaks is worse than
none, because you stop looking.

## 4.3 Stop throwing away the result

Look at what you just built. The handler runs on its own goroutine. Its
result is sent to a buffered channel. On the timeout path you walk away, and
the handler keeps running, finishes, and sends its result to a channel nobody
is reading.

That code shipped and ran for weeks before the sentence
arrived that this chapter is built on: every tool call is already
asynchronous. There is no "make tools async" project to plan. It is done. The
result of an abandoned call still arrives; the watchdog simply throws it
away.

So the job system is not a concurrency project. It is: stop discarding the
result. Register the pending call, keep the channel, hand the model a handle.

Once the model has the handle, the timer stops being a deadline. It is a wake,
the moment the dispatcher hands control back so the model can look. Nothing
was abandoned, nothing was killed, and a default of three seconds is safe
precisely because it kills nothing. The four-line `select` is still the
mechanism. The comment `// abandon` becomes `return handle`.

Here is the dispatch site, cut down to the ordinary-job path and using the
student's existing ownership. Compare it with Chapter 3's, where the handler
ran right here and if it never came back, neither did the agent:

```csharp
// Outline of Jobs.DispatchAsync. The real method also consumes pending
// observation settings and validates arguments before releasing the worker.
var job = new Job(this, Parent().Parent().AllocateJobHandle(), call);
try
{
    Parent().History().Append(new Event(EventKind.ToolCalled,
        callID: call.CallID, name: call.Name, args: call.Args,
        job: job.Snapshot()));
}
catch
{
    job.Destroy(); // No worker was started; no effect can escape the failed log.
    throw;
}
job.Start(tool, sanitizedCall, validationError);
return await job.ObserveAsync(call.CallID, delay, pattern, maxBytes,
    callerCancellation, wakeOnOutput: false);
```

Three things follow from putting it there and nowhere else.

In C#, `Task.WhenAny` can similarly end a wait without ending the task.
There is no safe general operation to forcibly abort arbitrary managed code.
We therefore retain its Job, pass a cancellation token, and tell the truth
when a handler does not cooperate. PTY processes are different: their owned
process group can be terminated and drained.

The handle goes at the dispatch site, not in the tools. One funnel, which is
already where security will be enforced, so nothing bypasses it. Measured on
the first-edition reference: every tool became a job by changing one function,
`Engine.Execute`, and five of the six Chapter 3 tools changed by one ignored
parameter. Before the tool runs it has a handle (integers from 1, one more per
job, per process), a file at `cr/io/N`, and the status `running`. The
dispatcher waits on the job, not in the tool.

In this edition Agent already owns dispatch. It now owns the Jobs collection;
each Job retains that collection as its parent. Ensemble allocates handles so
Agents within one Ensemble cannot accidentally reuse them, while each Agent's
data directory owns its output files. Shared partial declarations and value-like records remain in
Common, job behavior in Jobs, and tool behavior in Tools. The CLI uses `cr` as
its data directory; an embedded application supplies or receives a separate
directory per Agent. The specimen path `cr/io/N` is the CLI contract, not a
global filename for every Agent an application creates.

The scary error message evaporates. Before the handle, an abandonment had to
return an essay: the handler is still running, it may still land side
effects, do not naively retry. That warning existed because there was no way
back. With a handle it becomes *still running, handle 47, call
`wait_for_job(47)`*. Give the reader of an error a way back and the warning
is no longer needed, and that is true of error messages generally, not only
this one.

Every tool's output is now on disk. That fixes an asymmetry you have been
living with since Chapter 2 without noticing it: a redacted `run_command`
result could cite a file path, and every other tool's stub could only say
"re-run it." Chapter 2 declared `Ref{Kind, Locator}` with a `RefHandle` kind
and has not used it since. This is where that debt is collected, and the
output field on every job record is the first `RefHandle` in the log.

One rule is a contract and invisible unless stated, so here it is. A job that
finishes within the wake delay, under the output cap and without a cwd override,
is reported on its
first report as the bare result, byte-identical to what Chapter 3 returned.
That is what makes the change additive in the architectural sense: nothing
that worked yesterday reads differently today unless it took longer than
three seconds or said more than sixteen kilobytes. The exercise grades it as
`ch3parity`, by running all of Chapter 3's grader against your Chapter 4
binary.

The PTY shell format is the intentional exception explained in §4.7: merged
output replaces the old separate stdout/stderr sections while preserving the
exit-status information. A cwd override or a consumed one-shot setting also
needs its stated diagnostic. Do not hide those facts to imitate the old bytes.

The log contract, so you do not guess it. Both `tool_called` and
`tool_returned` carry `tool.job = {handle, status, output, bytes, exit_code?}`,
with `status` one of `running`, `done`, `killed`, and `output` a
`Ref{RefHandle, "cr/io/N"}`. The `job` field is absent on the four supervision
tools; they are the only calls that are not jobs. One new event, `job_killed`,
carries the job and a `reason` of `kill_job` or `shutdown`.

On disk the killed event uses a top-level `job` payload, with the reason inside
that object. This is an illustrative event, not a live receipt:

```json
{"seq":9,"type":"job_killed","time":"2026-01-01T00:00:09Z","job":{"handle":1,"status":"killed","reason":"kill_job","output":{"kind":3,"locator":"cr/io/1"},"bytes":0}}
```

## 4.4 Three verbs and one setter

From the same corpus Chapter 3 counted:

| tool | calls | share | what it is for |
|---|---|---|---|
| `send_input` | 800 | 1.14% | the process is waiting for you |
| `wait_for_job` | 322 | 0.46% | look again, up to a delay |
| `kill_job` | 225 | 0.32% | stop it |

`send_input` is the biggest of the three by more than two to one, and that
ordering is the finding. The intuition about job control is that it is mostly
about stopping runaway work. In the reference record it is mostly about talking to
work that is going fine and is waiting for an answer. A test runner asking
`y/N`. A REPL. A debugger sitting at its prompt. §4.8 spends the whole
chapter's budget on the last of those.

`wait_for_job` has one requirement that will cost you an hour if nobody says
it. It must work on a job that has already finished. A student who
implements it as "block on the channel" hangs forever on a completed job, and
diagnoses it as a deadlock in their own code rather than a missing case.

Completion, new output and kill must also wake a waiter that is already
blocked. Checking terminal state at entry covers a different ordering. A late
managed handler can finish after its job is killed; that completion must not turn
the retained state back into `done` or reopen the output file.

`kill_job` is SIGKILL to the process group, and the killed job is reported as
`killed`. Never as `done` with a strange exit code. The exercise section has
the story of how the grader learned that.

The fourth tool is not a verb and has no row in the table, because it did not
exist when the corpus was recorded. `tool_limits` sets the wait for the next
call, for tools whose arguments you do not own, and §4.6 is about it.

`list_jobs` is not built. Zero measured use.

## 4.5 A megabyte of test output

`go test ./...` on a real repository produces more text than you want in a
context window, and the cost is worse than it looks, because you pay for
those tokens on every subsequent turn of the conversation, not only the one
that ran the command. The exercise's `flood` fixture emits 1,340,013 bytes.
Left alone, that is a megabyte and a third in every request until the
session ends.

The rule has four parts. The full result text goes to `cr/io/<handle>`,
always, for every tool. What enters the context is the same bytes if they fit
under the cap, sixteen kibibytes by default, and otherwise a head, a line
saying how many bytes were omitted and the total size and the path, and a
tail. The path is the recovery route rather than decoration: the model can
read a range of it with the `read_file` it already has. And a report never
repeats bytes. Each job carries a cursor; `wait_for_job` and `send_input`
return what the model has not yet seen, and `ai_callback_pattern` is matched
against unseen output only, so a prompt the model already saw does not wake
it twice.

The cap bounds returned payload bytes; the status and omission marker add their
own text. Keep the complete bytes on disk. Chapter 3's explicit `read_file`
range and `max_bytes` remain available, but its implicit read cap moves to the
dispatch layer. Otherwise the tool would discard most of a large file before
the job could retain it for recovery.

Do not let the model choose the truncation. It happens on the way in, at the
dispatch site, before anything reaches the context window. A model asked to
summarize its own flood has already paid for the flood.

The `Ref` is recorded in the log and never rendered as an attachment part.
Chapter 2 preserved references without implementing media rendering, and this
chapter retains that boundary. The model gets a
path, and a tool that reads paths. A later chapter will want to promise that a
job's output can be attached to a request on every vendor, and this is where
it learns that it cannot.

## 4.6 Who decides how long to wait

Twenty-two minutes after shipping the watchdog a second commit shipped
fixing it. Both are dated 10 August 2026. The timestamps are 14:28 and 14:50.

The watchdog needed to know how long each tool may legitimately block. The
first version inferred it: a hand-maintained map of tool names, plus a scan
of each tool's description text hunting for documented timeouts.
`send_secret` accepted a deadline argument and was missing from the map, so it
silently got the default. Tools whose own default wait was longer than the
watchdog's would have been abandoned mid-wait while perfectly healthy. The
second commit fixed both by making every tool declare its blocking contract,
and the first version of this chapter's outline taught that declaration as
the answer.

Bill ruled, during the build, that the declaration was the wrong repair,
because the map had been the wrong question. *How long may this tool block?*
is a property of the tool. So it needs a table. So the table can be missing a
row. *How long to wait before looking?* is a property of the call. So the
caller supplies it. So there is nothing to declare and no row to be missing
from. The `send_secret` defect does not get caught by the second design.
There is no longer anywhere to write it.

Three limits, and they are ordinary arguments on the tools that wait:

| limit | default | on |
|---|---|---|
| `ai_callback_delay` | 3 s | `run_command`, `wait_for_job`, `send_input` |
| `ai_callback_pattern` | none | same three |
| `max_output_bytes` | 16 KiB | same three |

On the wire, `ai_callback_delay` is a JSON number of seconds, including
fractional values such as `0.2`; zero means look now. The pattern is a regex
string and `max_output_bytes` an integer. Negative delay, a nonpositive byte
cap or an invalid regex returns an ordinary tool failure.

| tool | arguments beyond the limits above |
| --- | --- |
| `run_command` | required `command` string; optional `cwd` string for this call |
| `wait_for_job` | required integer `handle` |
| `send_input` | required integer `handle` and string `input` |
| `kill_job` | required integer `handle` |
| `tool_limits` | the limit fields to set for the next call |

`send_input` submits a line, adding the terminal newline if it is missing.
For example, `{"handle":1,"input":"p answer","ai_callback_delay":10}`
submits the debugger command and waits up to ten seconds. It does not require
the model to put a newline inside every argument. No extra raw-input mode is
part of this exercise.

Precedence, later wins: the defaults, then a pending `tool_limits`, then the
call's own arguments. `tool_limits` exists because most tools have nowhere to
put these three arguments. The five local tools you own but never gave them
to, and every tool whose schema you do not own at all. So the model sets them
on the call before. It is one-shot, consumed by the very next call whatever
that call is, including a `kill_job` or another `tool_limits`. The friendlier
rule, "applies to the next job-creating or waiting call," is one line of code
and one more category the model has to hold in its head, and a limit set
several calls ago and still pending is a persistent escape hatch wearing a
friendlier name. *The next call* is a rule a model can keep.

The footgun in a one-shot setting is a silent misfire, and a live model named
it first: "it's easy to burn it on the wrong thing." You set
`max_output_bytes: 200000` for the `read_file` you are about to make, glance
at a directory first, and the limits went to `list_directory`. The symptom is
a truncated read one call later with no cause in sight. The fix is not a
target argument on `tool_limits`, because a target is a declaration the
setting has to match, which is the shape this section just spent a page
removing. The fix is that the misfire is loud. The consuming call's result
begins with one line:

```
[tool_limits consumed by this list_directory call: ai_callback_delay 3s,
 ai_callback_pattern none, max_output_bytes 200000]
```

It appears on every tool, including the four that are not jobs, and the
`tool_limits` reply itself says that the next call, whichever tool that is,
will report the consumption. A burn is now an error you read in the result it
caused, one round trip from the correction.

What the model changes on every monitoring call, because the wait is on the
call: it can wait 0.2 seconds to see whether a job started, thirty seconds
for a test suite, or "until `(dlv) ` appears" for a debugger, and change its
mind on the next call. No table anywhere knows what the tool needs, because
the table would be wrong.

Shutdown is the exit policy. At process exit, every running job is killed and
`job_killed{reason: shutdown}` is logged, so a job that never returned leaves
a record with a reason on it instead of an orphan process the human has to
find without a transcript.

The CLI calls Shutdown before its final log save, including error exits from
its input loop. Returning a final model answer alone does not end the CLI
session; EOF does. A library application has the same responsibility when it
finishes using its Agent.

## 4.7 The terminal

`run_command` runs under a pseudo-terminal, and the chapter has a receipt of
its own for why:

```
Stdin is not a terminal
```

That is `dlv`, on pipes, refusing to start. It is the closing demonstration
failing before it begins. Programs that prompt (debuggers, REPLs, anything
that asks `y/N`) check whether they are talking to a terminal and either
refuse or stop flushing. A PTY is not a nicety for `send_input`. Without one,
the closing demonstration does not start.

The costs, stated in the code and here. One merged stream: stdout and stderr
are the same bytes. The model's input is echoed back in the output, because
that is the terminal's line discipline; it is how the model sees its own
keystroke land, and turning echo off changes what `dlv` shows, so it stays.
`\r\n` is normalized to `\n`. `TERM=dumb`, fifty rows by two hundred columns.
The exit status is the file's last line, `exit_code: N`, because a stream
cannot put it first, and Chapter 3's regex still matches. If no PTY is
available the tool returns an error. It never silently falls back to pipes,
which would ship an agent whose debugger works on one machine and refuses on
another with no message saying why.

`kill_job` on a process is SIGKILL to the process group. For a managed handler,
.NET supplies cooperative cancellation, not a safe way to abort arbitrary code.
Mark the job killed, report that limitation explicitly, and ignore late output
and completion. Do not block shutdown forever waiting for an uncooperative
callback or pretend that its code was forcibly terminated.

Ordinary argument, JSON, I/O, permission and process-start failures become paired
tool errors. Unexpected invariant failures must not be disguised as successful
reports or routine model feedback. In C#, the Task retains the exception;
observation and completed-job cleanup propagate it to the host. Unlike a Go
panic, a faulted Task does not automatically terminate the whole process, so the
host owns that policy. Cleanup must still detach the job and close its resources.
The PTY implementation here is macOS-specific and never falls back to pipes.

### Does shell state persist between calls?

Students hit this one, and shipping agents answer it differently. One widely
used terminal agent's shell tool says yes: one shell, `cd` sticks, and its
issue tracker shows the cost, a `cd` outside the approved directories
silently reverted with a "Shell cwd was reset" notice appended, because the
shell now holds state the security boundary has to police. Editor-style
agents ship named terminal panes whose state persists. The reference says no,
and the job model makes it not a close call.

A persistent shell is a job. `run_command "bash --norc"` with pattern `\$ `,
then `send_input "cd lyric"`, `send_input "make"`. The model gets a shell
whose state sticks, and it has a handle. Every `send_input` to that handle is
a logged event in order, so the log still reproduces the session. The
terminal agent's problem was never state; it was implicit state with no
handle to attribute it to. Those named panes are handles with a GUI.
Isolation is the primitive and persistence is composed on top of it,
explicitly. You can build a persistent shell out of isolated jobs. You cannot
build isolation out of a persistent shell.

Overlapping jobs cannot share one shell. Two `run_command`s in flight at
once, which is the point of this chapter, is not a thing one bash does.

And the measured tax is a directory tax, not a state tax. The numbers were
requested before the ruling, and they are worth seeing. On 26,781 archived
`run_command` calls in the reference agent's logs, 69.6% begin with `cd`,
18,651 of them, and almost all name one
directory: `cd ~/projects/forge` 7,979 times, `lyric` 1,908, `coderhapsody`
1,240. Environment setup (`source .../activate`, `export`, `nvm use`) is
0.12%, thirty-three calls. That is not a model navigating. That is a model
started in the wrong directory and correcting for it on every call, forever.

So the remedy for the 69.6% is two things, neither of which is state. `cwd`
as an argument on `run_command`, for this call only: relative paths resolve
against the workspace, a missing directory is an error and never a silent
fall back to the workspace, and the effective directory is recorded on the
job and printed in the report when it is not the default, which is Chapter
2's rule, record and never infer. And the default `cwd` is a launch-time
setting, the agent's workspace, which is not the same directory as the one
holding the event log. That shipped the morning after the numbers
were measured.

The serialized `tool.job.cwd` is the optional override, not a repetition of the
default absolute directory. Leave it absent or empty when the call uses the
default, while still executing in the correct directory. The student's first
run recorded the default everywhere and scored 65/100; the coordinator's first
erratum had also omitted this convention. Correcting the representation raised
the original score to 90/100, with only inherited media parity still failing.

The grader launches the CLI in its scratch workspace and supplies no new cwd
flag. The accepted library uses Agent-owned Workspace configuration, separately
from its DataDir; its CLI offers `ENSEMBLE_WORKSPACE` as a launch-time choice.
That name is an implementation choice. Preserve the distinction between where
commands run and where logs/output live without inventing sticky shell state.

Why no `set_cwd` tool. It is the `tool_limits` argument again, with a worse
failure mode. A sticky default set at call 40 and compacted away by call 300
is state the model can no longer see but still acts on, and a wrong wake
costs seconds while a wrong directory runs `rm -rf build` in the wrong tree.
To make it safe you would have to promote sticky settings to the
never-dropped category in Chapter 2's reducer. One-shot settings never need
to survive compaction; sticky ones always do, and that single sentence is why
both of this chapter's knobs are per-call.

## 4.8 Your agent drives a debugger

The chapter closes on a demonstration, chosen because it is the one thing the
Chapter 3 agent could not do at any speed:

```
run_command "PAGER=cat dlv debug ./testdata/dbg"   ai_callback_pattern="\(dlv\) "
send_input  "b main.go:7"
send_input  "c"
send_input  "p answer"
send_input  "q"
```

Those are tool actions, not shell commands to paste into a terminal. For a
debuggable Go package in the workspace, the first argument object is:

```json
{"command":"PAGER=cat dlv debug .","ai_callback_delay":10,"ai_callback_pattern":"\\(dlv\\) "}
```

JSON escapes each backslash; the decoded regex contains one before each
parenthesis. Use the returned handle for subsequent `send_input` calls.

A blocking tool call returns when the process exits. A debugger does not exit
until you tell it to quit. With Chapter 3's tools, driving `dlv` is not slow
and it is not awkward. It is impossible, and impossible in two independent
ways: the dispatcher would wait for an exit that never comes, and on pipes
`dlv` refuses to start at all. The chapter's thesis stops being a matter of
taste and becomes a capability boundary the reader can stand on either side
of.

The C# worker awaits the process and its readers; the dispatcher waits only for
an observation. Returning a report does not finish the worker. From startup to
drain, Job owns the process and readers, including the eventual exit status.

Finishing a worker closes its output writer, while reports reopen the retained
file for reading. A dispatcher that declares completion as soon as it returns a
handle cuts off a reader that is still writing. The apparent output-capture bug
is really a lifetime bug. Test both the running state and the eventual bytes;
an empty successful result proves neither.

It also teaches `ai_callback_pattern` honestly. You do not sleep for a guessed
interval and hope the prompt has appeared. You wait for the string `(dlv) `,
because that is the actual signal that the debugger is ready for input. A
fixed delay is a race condition with a comfortable name.

The grader drives this through the fake vendor: five scripted calls, and `42`
read off the breakpoint. The second-edition student also drove a real Gemini
session through its own human chat interface: launch, breakpoint, continue,
inspect, quit. The retained result was `42`, followed by exit code 0.

That demonstration produced a useful mistake in the evidence reading. The
student saw escaped text in a serialized display and concluded that the
pattern had two backslashes before each parenthesis. It then said the calls
had waited their full ten seconds and paid for a repeat. The original log
disproved both claims: its decoded pattern was `\(dlv\) `, and the first four
calls returned in about 0.270, 0.005, 0.064 and 0.002 seconds. Quit correctly
waited for exit without a pattern. The repeat worked too; it was unnecessary.
Inspect the decoded argument and measured intervals before diagnosing escaping
from its printed representation. Your agent can now debug the code it wrote.

## A tool from outside the framework

The first edition put a refactor here. Its agent could read files and drive a
debugger, but another program could not use the machinery without entering its
implementation. This edition established the assembly and ownership boundaries
before the first request. The useful proof from that chapter still belongs
here: a separately built program imports the framework, registers a tool the
framework has never seen, and runs an Agent that calls it.

The package names and public API spelling can differ. The external program
must use the public composition root, with public types for configuration,
declarations and handler context. It must not import an internal package or
assemble a second Engine/History/job loop beside the one the CLI uses. The
consumer is a separate assembly with an ordinary project reference, without
`InternalsVisibleTo` or reflection. Its compilation exposes missing public
types that tests inside the framework can accidentally reach.

Register the tool on its Agent during synchronous setup, before asking it to
work. Keep declaration and dispatch on the same per-Agent registry. A global
registration function would let one consumer change another Agent's tools.
Preserve existing builtins when adding a tool; reject invalid registration
without altering the existing set, and own the declaration's schema bytes so
the caller cannot change them after setup. No concurrent registration service
or extra dispatch route is needed.

The C# example consumer defines `shout`, taking `{"text":"hello framework"}`
and returning `HELLO FRAMEWORK`. Its handler receives the actual public Job,
validates its argument and uppercases the supplied text. There is no special
case for that name in the core. The declaration goes to the model, the returned
call follows the ordinary job path, the handler's output is saved, and the
same output reaches the model in the correlated result. A final answer alone
cannot establish any of those steps.

You can use arithmetic, a lookup or another small operation instead. The point
is that the framework executes code it did not contain, through machinery it
already did contain. Build the application separately and inspect the actual
call, successful result, job metadata and output file. An unknown-tool error
also contains a tool-result marker; accepting that marker would certify a
handler that never ran. The first-edition grader made exactly that mistake.

### The logger ambush

How do you debug your agent?

The engine calls tools, the tools call the shell, the shell runs builds that
fail and tests that flap. When something goes wrong, the model's next response
is its interpretation of the evidence, not the evidence itself. The first
edition deliberately reached its refactor with 5,047 lines of Go and no logger,
then asked the student to add one. That was the ambush Chapter 1 used to explain
why ownership mattered before it seemed necessary.

A logger is a facility that every piece of code in the system needs access to.
In a flat package, the tempting solution is a global variable. With multiple
Agents in one application, that hides which application owns the writer and
how a caller can replace it. Passing a new logger argument to every tool would
instead duplicate access that the existing parent relationships should provide.

The root is Ensemble. It owns one logger; Agent retains its Ensemble, Engine
retains its Agent, and a tool call retains its actual Agent. The custom tool
reaches the root's narrow logging interface through that chain:

```csharp
job.Parent().Parent().Parent().Log("shout called");
```

The original refactor's per-Agent logger and global-registration examples do
not describe this ownership. The accepted implementation uses the standard
writer with synchronized writes and timestamps, behind `Log`. No caller needs
the concrete logger or a copied service bundle. The external consumer's
diagnostic is useful proof: a leaf the framework did not know about can reach
the application's existing facility.

Inspect imports and constructor relationships as well as that output. A
compiler rejects an import cycle, but an acyclic sibling import can still
violate the star. Likewise, finding `Logf` in a struct does not prove that the
running tool follows its actual parents. Common holds the actual shared types, persistent fields, and mandatory
partial declarations. Bodies belong in functional directories. The C# compiler
joins the parts; it does not police directory-level dependencies or the location
of every field. Those remain source-review obligations.
The live diagnostic proves that a message reached the writer, not the entire
architecture that carried it.

## Exercise

The core CLI contract remains Chapter 3's because `ch3parity` requires it: a
JSON-lines transcript on stdin, no network, deterministic, the same fake
vendor as Chapters 2 and 3. Handles are integers from 1, one per job, per
Ensemble. In the grader's single-Ensemble process, the first is 1. The fixtures depend on that (`wait_for_job(1)`), so it is stated as a
contract rather than read back from the log.

Keep `chat`, `render LOG`, `dump` and `CH02_LOG`; no new core CLI mode is
required. The separate consumer is a new application of the public library,
not a replacement for that CLI.

The submission has two working programs: the core CLI with supervised jobs,
and the external application that proves somebody else can use that machinery.
For the external application:

1. Create a separate C# executable project referencing the public library.
   Build it independently, without special internal access or another tool loop.
   `Examples/Consumer` demonstrates this boundary.
2. Read the established provider environment (`LLM_VENDOR`, `LLM_MODEL`,
   `LLM_API_KEY`, `LLM_BASE_URL`) and construct the Ensemble and its Agent
   through the public API.
3. During setup, register one custom tool on that Agent, keeping its builtins.
   Give the tool a name, a nonempty description, an object argument schema and
   a handler that computes a result from its arguments. Have that handler log
   a timestamped diagnostic through the actual parent chain to Ensemble's
   logger.
4. Ask a prompt intended to trigger the tool and complete the exchange. Verify
   that its declaration reaches the provider, its handler executes through the
   existing job machinery, and its successful output reaches the next request
   under the call's id. Inspect the saved output and job metadata too; a final
   answer or an unknown-tool error is insufficient proof.
5. Shut down through the public lifecycle so the application's outstanding jobs
   are cleaned up. Run the ordinary CLI checks as well as the separate consumer
   check below.

Use the public names your library exposes. This carries forward the original
Chapter 5 exercise without its global registration example or a new refactor.
The topology and ownership requirements apply to both programs; the source
review obligations below remain part of completing the exercise.

The retained debugger fixture needs an installed Go toolchain and `dlv`.
Those are programs exercised by our C# shell tool, not the solution language.
Check availability first; a missing prerequisite is not a product failure or
permission to install software. This endpoint uses macOS's installed `script`
for its real PTY, with fixed geometry and a fresh process group; other operating
systems need their own tested launcher. Pipes are not a portability fallback.

### What the fixture provides

All Go helpers are built to binaries once per run, never `go run`, so their
timing is theirs and not the compiler's. A command that finishes fast
(`go version`). A command that outlives the wake (`sleeper`). A command that
never returns (`blocker`), which is the screenshot incident, reproducible; it
sleeps in a loop, and the reason is in its source, because with no other
goroutine the Go runtime calls `select {}` a deadlock and exits 2, and the
first version of this fixture died on its own while the check waited for it
to be alive. A command that prompts and echoes (`echoer`), so `send_input`
has something real to talk to. A command that emits more than a megabyte
(`flood`, 1,340,013 bytes). And a one-line program with a breakpoint to hit
(`testdata/dbg`, `answer := 42`).

### Checks

| id | points | what it grades |
|---|---|---|
| `ch3parity` | 10 | Chapter 3's corrected grader passes fully against the Chapter 4 binary |
| `jobmodel` | 25 | handle on `tool_called` before the tool ran, for `read_file` and `list_directory` as well as the shell; `cr/io/N` bytes equal the result the model saw; `wait_for_job`'s own record carries no job |
| `waitjob` | 10 | delay honored (`running` at 0.2 s); wait returns the result; works on an already-finished job; file ends with marker and exit code |
| `sendinput` | 10 | pattern wakes on the prompt and not the echo; input reaches the process; its reply comes back; clean exit recorded |
| `debugger` | 5 | the fake drives `dlv` to a breakpoint and reads `42` |
| `killjob` | 10 | `job_killed{reason kill_job, status killed}`; the post-kill wait is told `killed`, never `done` or an exit code; a 30 s waiter returns early; the pid is dead |
| `bigoutput` | 15 | full output on disk with first and last line; inline within the cap, naming locator and exact total; `max_output_bytes 2048` honored; `read_file` of a 1 MiB file truncated the same way, at dispatch |
| `toollimits` | 10 | `tool_limits` is not a job; applies to the next call; one-shot; pattern form wakes; the call's explicit argument wins |
| `shutdown` | 5 | a job still running when the model stops is killed at exit and `job_killed{reason shutdown, status killed}` is logged |

These retained jobs checks sum to 100. The merged chapter separately grades
the public boundary for 60 points:

| id | points | what it grades |
| --- | ---: | --- |
| `external-consumer` | 25 | a separately built application imports the submitted public library, declares its custom tool and returns its successful, correlated output to the vendor |
| `star-topology` | 25 | implementation imports respect the common hub; the consumer depends on the submitted public package |
| `tool-diagnostic` | 10 | the executed custom tool emits the expected timestamped diagnostic |

The total is 160, with the 100 jobs points kept visible rather than silently
rescaling them. An import failure or a custom tool that returns an unknown-tool
error is a failure of the external proof, not evidence of execution.

### Running the C# endpoint checks

The inherited 100-point jobs harness accepts our executable. The additional
60-point consumer adapter builds Go modules and inspects Go imports; it cannot
grade a C# assembly by being given a different path. Keep that setup limitation
visible instead of inventing a 160/160 result. Independently build and exercise
the C# consumer through a local protocol fake and inspect the partial declarations
and actual parent access paths. A failed overall grade is still a failed grade,
with its executable and language-specific portions reported separately.

From the repository root, using the approved .NET 10 SDK and controlled offline
environment:

```sh
dotnet run --project solutions/edition-3/main/Tests/Tests.csproj
dotnet publish solutions/edition-3/main/Cli/Ch04.csproj -c Release -o /tmp/ch04-cli
go run ./cmd/grade-edition2 -ch 4 -json /tmp/ch04-cli/ch04
dotnet build solutions/edition-3/main/Examples/Consumer/Consumer.csproj
```

The component suite tests real tools and PTYs with fake HTTP at the provider
boundary. The consumer exercise additionally verifies successful, correlated
custom output—not merely the presence of a result marker—and the timestamped
root diagnostic. No real-provider acceptance follows from either fake.

The numeric result has limits. Independent source review still checks actual
parent relationships, Ensemble-owned logging, behavior in the correct package
and absence of mutable global application state. The original logger check
rejected a working multi-hop parent path because it expected a direct field
selector. Its globals check depended on variable names. Neither can establish
those architectural properties. Keep the source findings alongside the runtime
results; do not label a printed log line proof of ownership.

The weights. `jobmodel` is 25 because it is the chapter, and because the
common wrong answer, special-casing `run_command` instead of changing the
dispatch site, passes a naive test and fails the moment a tool you do not own
wedges. Both negative controls, the `read_file` legs in `jobmodel` and
`bigoutput`, are measured load-bearing: delete either and the "it's a shell
problem" student scores 100. `toollimits` is 10 because it grades three
properties and a forward reference in one fixture. `debugger` is 5 rather
than more because `sendinput` already grades the mechanism; the five points
buy the proof that the mechanism reaches a real program. `shutdown` is 5
because it is one behavior, and because without it the never-returning job is
a leak the grader would otherwise have to hunt with `kill -0`.

### What would still pass if I deleted this?

Same audit as the last three chapters: delete each protected behavior from
the reference and confirm the score drops, and treat a row reading 100 → 100
as the finding. One row was `killed-job-reported-as-done`, and it scored 100
on the first run.

Here is how. The killed process died. Its goroutine ended the job as `done`
with `exit_code -1`, and the waiter was told *job 1 done, exit_code -1 ...
[job 1 killed: kill_job]*. The check searched the report for the word
"killed", and the kill note supplied it. A text regex can be satisfied by the
very message that documents the failure. The fix is to grade structure:
`job_killed` carries a `status`, which must be `killed`, and the post-kill
report must not claim `done`, `finished`, `completed`, `succeeded`, or an
`exit_code`.

Three more findings from this chapter's audit, each of them a shape you will
meet in your own grader. Assert every mutation actually landed and changed
bytes; this project has now been bitten three times, and the Chapter 4 rig
refuses a regexp that matches other than once or whose replacement is a
no-op. A mutant of a module with dependencies needs the dependency: the
mutant `go.mod` carries the `creack/pty` require and the course `go.sum`, or
the mutant fails to build and the failure is misread as detection. And treat
a skipped check as a failed check. A `t.Skip` on an empty fixture is a
vacuous pass with better manners, which is why `debugger` fails rather than
skips when `dlv` is absent.

The first-edition audit deliberately omitted two mutants, with the reason recorded.
Pipes-for-PTY, because the textual change is too large for a regexp and the
evidence is the measured refusal plus `debugger` passing. And
pattern-never-wakes, because the session would exceed the forty-five second
harness timeout and leak a `dlv` process; the two pattern mutants that exist
cover the semantics.

The second-edition review found another timing trap: starting a waiter
goroutine does not prove it is waiting when Kill happens. Removing the
notification still passed the initial test 20 times. Releasing a blocked file
read likewise does not prove its late completion has reached Finish before
the assertion. The corrected tests establish those orderings around the real
job code. Their first repair then raced while cleaning up a failing test;
a buffered result channel made both successful and defective paths finish
cleanly. Check the complete failure output. The intended assertion does not
make a race warning beside it disappear.

One environmental assumption, stated so a failure elsewhere is diagnosable.
`ch3parity` holds because Chapter 3's fixtures finish inside the three-second
wake; `go run ./testdata/exit7` measured 0.05 to 0.3 seconds warm and 1.3
cold. A machine where a cold `go run` of a one-line program exceeds three
seconds fails `ch3parity` with a `running` report where an exit code was
expected. The fixture assumed blocking; the wake is behaving. Run `exit7`
once before grading to warm the build cache. The PTY library is measured on
macOS; Linux is assumed.

### What you are not building

No mailbox, no interrupts, no `Interrupted` turn state. Jobs do not outlive
the process, because one transcript is one process and `Shutdown` runs at
exit. A tool served by somebody else's process is named in §4.1 as the worst
case and `tool_limits` is built for it, but you are not connecting to one
yet. The custom consumer uses the same dispatch site; it does not add another
scheduler. Four supervision tools and one ordinary-job path.

## 4.9 Drive it yourself

Ungraded, and the one to do first, because this is the chapter where the
frustration from the end of Chapter 3 gets its answer.

Rebuild your own binary from its module directory. Configure the provider,
discovered model, endpoint and selected key as in Chapter 3, keeping the key
out of command arguments and tool subprocesses. Then go back to a disposable
project:

```
dotnet publish solutions/edition-3/main/Cli/Ch04.csproj -c Release -o /tmp/ch04-cli
cd /path/to/scratch/project
CH02_LOG=session.jsonl /tmp/ch04-cli/ch04 chat
```

Ask it to run the test suite that took ninety seconds. Three seconds in, it
comes back: *job 3 still running after 3s. no new output; 0 bytes total at
cr/io/3*, and a line telling it what it can do next. Watch what the model does
with that. Most of them wait again with a longer delay. Some of them go read
a file while they wait, which is a thing your Chapter 3 agent could not have
conceived of. Then ask it to start the development server, the request that
never came back last chapter, and watch it start the server, get a handle,
and carry on with the conversation while the server runs.

Then do the demonstration from §4.8 by hand, in chat. Ask it to debug
something small with `dlv`. Watch the `(dlv) ` pattern do the waiting. If you
have never seen a language model sit at a debugger prompt, set a breakpoint,
and read a variable back, it is worth the fifteen minutes on its own.

The second-edition student drove those paths with all three providers on
October 9. Each started a delayed job, waited for its end, answered a `NAME>`
PTY prompt with `Ada`, observed `hello Ada`, and killed a running sleep job.
The retained [student record](../../docs/edition-2-notes/ch04.md) links the
actual event logs and distinguishes model summaries from inspected output:

| live model | plain input | cache read | output |
| --- | ---: | ---: | ---: |
| `claude-haiku-5-5` | 29,113 | 0 | 985 |
| `gpt-5.4-mini` | 7,215 | 6,272 | 593 |
| `gemini-3.8-flash` | 30,236 | 0 | 1,437 |

No cache writes were recorded. OpenAI also set a 256-byte cap: the job stored
564 bytes and the report named 308 omitted bytes. The Gemini session includes
the successful debugger sequence from §4.8. These totals belong to the actual
sessions, not a cost forecast or new runs of the enhanced chapter.

Shutdown needs its own observation. In the first attempt, a `sleep 60` job
expired before EOF, so it demonstrated no shutdown kill. A bounded repeat
checked that a `sleep 300` process existed immediately before EOF and was gone
immediately after; the log recorded `killed` with reason `shutdown`. The repeat
used 4,444 input and 139 output tokens. The unnecessary debugger repeat used
11,578 input and 728 output; both remain in the record rather than disappearing
from the account because their lessons were inconvenient.

For free plumbing checks, use the revised grader and Chapter 3's
serve-only fake with your own binary. The original convenience launchers build
first-edition answers; they do not validate yours. The simple fake script
demonstrates ordinary calls, while the jobs grader supplies the wait, input,
kill, output and debugger scenarios.

### Take the library outside

Build the separately compiled C# example from `Examples/Consumer`. Its local
exercise reads the established environment and refuses a non-loopback endpoint,
so accidentally inherited provider settings cannot turn a plumbing check into
a paid request. Registering `shout` preserves the built-ins. A local fake should
request `shout`, `read_file`, and `run_command`, then verify all three results,
job files, and the timestamped `shout called` diagnostic before sending its final
reply. The application has no private access to the library.

The earlier Edition 2 external-consumer provider runs are historical evidence,
not C# acceptance. Real-provider exercises require current permission and budget.
Record blocked portions as incomplete; do not replace model judgment with a
scripted fake and describe the two as equivalent.

Then try what the original external-use exercise proposed: a small application
with a tool that knows something about your project. A tool that queries your
database, checks CI status or reads a monitoring dashboard. The framework does
not know those operations and does not need to. These are ideas for using the
library, not additional tools required by this exercise.

Preserve a runnable chapter endpoint and record actual results and limitations.
Commit or tag only with the repository owner's authorization. Keep
the actual scores and accepted exceptions; original Chapter 4 scored 90/100
FAIL, and original Chapter 5 scored 110/120 FAIL with nested Chapter 4 still
90/100. Those accepted student checkpoints retain their history separately
from the revised grader's results.

Now the thing this chapter cannot fix, and you should feel it before the next
chapter names it. Start that ninety-second test suite again, and while the
job is running, type something. Anything. "Stop, wrong directory." Nothing
happens until the wake returns. Your agent can run a job without dying, and
it still cannot hear you while it does, because the loop that dispatched the
tool is the only thing that would notice your message, and it is parked
waiting for the answer. A supervised job fixes the freeze and leaves the
deafness.

Fixing that is not more job machinery. It is a different shape: one inbound
queue carrying prompts, interruptions, and job completions as the same kind of
thing, drained by a loop that never blocks on a tool. This chapter has earned
that by making the pain specific. On 10 August the agent could not be told
anything at all. After this chapter, it can be told anything, three seconds
from now.
