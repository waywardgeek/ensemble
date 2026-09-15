# Chapter 4: Jobs, or Why a Tool Call Is a Process You Supervise

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
usable. And it makes every tool incapable of taking the agent down with it.

Nothing in this chapter dies on its own. No deadline, no budget, no watchdog.
A job runs until it finishes or until the model kills it, and the only thing
the model decides per call is how long it will wait before looking. That
inversion is the chapter's design idea. The wait belongs to the caller, not to
the tool, and §4.6 is where the argument happens, because I got it wrong twice
in one afternoon before Bill got it right.

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
and in my logs the tool most likely to wedge, because a browser call on a page
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

What I shipped first, on the afternoon of the screenshot, is what you would
ship first, and you will be disappointed by how little it is:

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
running. In the design this chapter actually builds, the containment sentence
survives in exactly one place, and §4.7 says where, because everywhere else
the design stops abandoning anything.

One detail is invisible until it bites. The channel is buffered with capacity
one, deliberately. On the timeout path nobody reads, so an unbuffered send
would park the abandoned goroutine forever, and the watchdog would cause the
very leak it exists to bound. A safety mechanism that leaks is worse than
none, because you stop looking.

## 4.3 Stop throwing away the result

Look at what you just built. The handler runs on its own goroutine. Its
result is sent to a buffered channel. On the timeout path you walk away, and
the handler keeps running, finishes, and sends its result to a channel nobody
is reading.

I built that, shipped it, and lived with it for weeks before the sentence
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

Here is the dispatch site in the reference solution, cut down to the part
that changed. Compare it with Chapter 3's, where the handler ran right here
and if it never came back, neither did the agent:

```go
// The handle is allocated here, for every tool, before the dispatcher
// knows anything about what the tool will do.
job, err := e.Jobs.Start(call.Name, call.CallID)
if err != nil {
    return err
}
e.record(Event{Type: ToolCalled, Tool: &ToolData{
    CallID: call.CallID, Name: call.Name, Args: call.Args, Job: job.Data(),
}})

// The tool runs on its own goroutine and reports into the job.
go func() {
    out, err := tool.Run(&Call{Job: job, Jobs: e.Jobs, Limits: limits}, call.Args)
    job.Finish(out, err)
}()

// The wait is on the job, not the tool. If the tool never returns, this
// returns anyway, with a handle, and the model decides what happens next.
reason := job.Wait(limits)
out := job.Report(reason, limits)
```

Three things follow from putting it there and nowhere else.

The handle goes at the dispatch site, not in the tools. One funnel, which is
already where security will be enforced, so nothing bypasses it. Measured on
the reference: every tool became a job by changing one function,
`Engine.Execute`, and five of the six Chapter 3 tools changed by one ignored
parameter. Before the tool runs it has a handle (integers from 1, one more per
job, per process), a file at `cr/io/N`, and the status `running`. The
dispatcher waits on the job, not in the tool.

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
finishes within the wake delay, under the output cap, is reported on its
first report as the bare result, byte-identical to what Chapter 3 returned.
That is what makes the change additive in the architectural sense: nothing
that worked yesterday reads differently today unless it took longer than
three seconds or said more than sixteen kilobytes. The exercise grades it as
`ch3parity`, by running all of Chapter 3's grader against your Chapter 4
binary.

The log contract, so you do not guess it. Both `tool_called` and
`tool_returned` carry `tool.job = {handle, status, output, bytes, exit_code?}`,
with `status` one of `running`, `done`, `killed`, and `output` a
`Ref{RefHandle, "cr/io/N"}`. The `job` field is absent on the four supervision
tools; they are the only calls that are not jobs. One new event, `job_killed`,
carries the job and a `reason` of `kill_job` or `shutdown`.

## 4.4 Three verbs and one setter

From the same corpus Chapter 3 counted:

| tool | calls | share | what it is for |
|---|---|---|---|
| `send_input` | 800 | 1.14% | the process is waiting for you |
| `wait_for_job` | 322 | 0.46% | look again, up to a delay |
| `kill_job` | 225 | 0.32% | stop it |

`send_input` is the biggest of the three by more than two to one, and that
ordering is the finding. The intuition about job control is that it is mostly
about stopping runaway work. In my own record it is mostly about talking to
work that is going fine and is waiting for an answer. A test runner asking
`y/N`. A REPL. A debugger sitting at its prompt. §4.8 spends the whole
chapter's budget on the last of those.

`wait_for_job` has one requirement that will cost you an hour if nobody says
it. It must work on a job that has already finished. A student who
implements it as "block on the channel" hangs forever on a completed job, and
diagnoses it as a deadlock in their own code rather than a missing case.

`kill_job` is SIGKILL to the process group, and the killed job is reported as
`killed`. Never as `done` with a strange exit code. The exercise section has
the story of how my grader learned that.

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

Do not let the model choose the truncation. It happens on the way in, at the
dispatch site, before anything reaches the context window. A model asked to
summarize its own flood has already paid for the flood.

The `Ref` is recorded in the log and never rendered as a part. Chapter 2's
renderers refuse blobs with a loud `not implemented` on two of the three
vendors, and that refusal is the lesson rather than a gap. The model gets a
path, and a tool that reads paths. A later chapter will want to promise that a
job's output can be attached to a request on every vendor, and this is where
it learns that it cannot.

## 4.6 Who decides how long to wait

Twenty-two minutes after shipping the watchdog I shipped a second commit
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
row. *How long will I wait before I look?* is a property of the call. So the
caller supplies it. So there is nothing to declare and no row to be missing
from. The `send_secret` defect does not get caught by the second design.
There is no longer anywhere to write it.

Three limits, and they are ordinary arguments on the tools that wait:

| limit | default | on |
|---|---|---|
| `ai_callback_delay` | 3 s | `run_command`, `wait_for_job`, `send_input` |
| `ai_callback_pattern` | none | same three |
| `max_output_bytes` | 16 KiB | same three |

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
it before I did: "it's easy to burn it on the wrong thing." You set
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

[CODER: the consumed-by note is ruled (review-ch04-tools-friction §7) and
not yet in `solutions/ch04`; `toollimits` gains two note legs. The prose
above states the ruling as built. Reconcile or mark.]

What the model changes on every monitoring call, because the wait is on the
call: it can wait 0.2 seconds to see whether a job started, thirty seconds
for a test suite, or "until `(dlv) ` appears" for a debugger, and change its
mind on the next call. No table anywhere knows what the tool needs, because
the table would be wrong.

Shutdown is the exit policy. At process exit, every running job is killed and
`job_killed{reason: shutdown}` is logged, so a job that never returned leaves
a record with a reason on it instead of an orphan process the human has to
find without a transcript.

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

`kill_job` on a process is SIGKILL to the process group. `kill_job` on a job
that is only a goroutine, a `read_file` that never came back, marks the job
killed and the report says, in those words, that Go cannot stop it. The
containment sentence from §4.2 survives here and nowhere else, because this is
the only place the design still cannot do better.

There is no `recover` around the tool goroutine. A panic in a tool is an
invariant violation, and an invariant violation taking the process down is
working as intended. Unix only, via the process group; Chapter 3 already was,
via `sh -c`.

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

And the measured tax is a directory tax, not a state tax. Bill asked for the
numbers before ruling, so I counted. On 26,781 archived `run_command` calls
of mine, 69.6% begin with `cd`, 18,651 of them, and almost all name one
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
holding the event log. I shipped exactly that the morning after the numbers
were measured.

[CODER: `cwd` is ruled (§4.7, Ruled 11) and not yet in `solutions/ch04`;
`jobmodel` gains a `cwd` leg (assert on a line of `pwd` output, not a
substring) and a no-persistence leg. Prose states the ruling as built.]

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

A blocking tool call returns when the process exits. A debugger does not exit
until you tell it to quit. With Chapter 3's tools, driving `dlv` is not slow
and it is not awkward. It is impossible, and impossible in two independent
ways: the dispatcher would wait for an exit that never comes, and on pipes
`dlv` refuses to start at all. The chapter's thesis stops being a matter of
taste and becomes a capability boundary the reader can stand on either side
of.

It also teaches `ai_callback_pattern` honestly. You do not sleep for a guessed
interval and hope the prompt has appeared. You wait for the string `(dlv) `,
because that is the actual signal that the debugger is ready for input. A
fixed delay is a race condition with a comfortable name.

The grader drives this through the fake vendor: five scripted calls, and `42`
read off the breakpoint. I also ran it live, `scripts/live.sh 4 gemini
rounds` on `gemini-3.8-flash`: five `tool_called` events, every one carrying
the pattern, and the model read `42` back unprompted. Your agent can now
debug the code it wrote.

## Exercise

The contract is Chapter 3's, exactly, because `ch3parity` requires it: a
JSON-lines transcript on stdin, no network, deterministic, the same fake
vendor as Chapters 2 and 3. Handles are integers from 1, one per job, per
process. The fixtures depend on that (`wait_for_job(1)`), so it is stated as a
contract rather than read back from the log.

One prerequisite: `dlv` on your `PATH` or in `$(go env GOPATH)/bin`. The
`debugger` check fails with the `go install` line when it is absent. It does
not skip.

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
| `ch3parity` | 10 | Chapter 3's whole grader passes against the Chapter 4 binary |
| `jobmodel` | 25 | handle on `tool_called` before the tool ran, for `read_file` and `list_directory` as well as the shell; `cr/io/N` bytes equal the result the model saw; `wait_for_job`'s own record carries no job |
| `waitjob` | 10 | delay honored (`running` at 0.2 s); wait returns the result; works on an already-finished job; file ends with marker and exit code |
| `sendinput` | 10 | pattern wakes on the prompt and not the echo; input reaches the process; its reply comes back; clean exit recorded |
| `debugger` | 5 | the fake drives `dlv` to a breakpoint and reads `42` |
| `killjob` | 10 | `job_killed{reason kill_job, status killed}`; the post-kill wait is told `killed`, never `done` or an exit code; a 30 s waiter returns early; the pid is dead |
| `bigoutput` | 15 | full output on disk with first and last line; inline within the cap, naming locator and exact total; `max_output_bytes 2048` honored; `read_file` of a 1 MiB file truncated the same way, at dispatch |
| `toollimits` | 10 | `tool_limits` is not a job; applies to the next call; one-shot; pattern form wakes; the call's explicit argument wins |
| `shutdown` | 5 | a job still running when the model stops is killed at exit and `job_killed{reason shutdown, status killed}` is logged |

The sum is 100. The code sums itself, with a test that adds the checks'
declared points, so the table is a cross-check and not the instrument.

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
as the finding. My row was `killed-job-reported-as-done`, and it scored 100
on the first run.

Here is how. The killed process died. Its goroutine ended the job as `done`
with `exit_code -1`, and the waiter was told *job 1 done, exit_code -1 ...
[job 1 killed: kill_job]*. My check searched the report for the word
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

Two mutants were deliberately not written, with the reason recorded.
Pipes-for-PTY, because the textual change is too large for a regexp and the
evidence is the measured refusal plus `debugger` passing. And
pattern-never-wakes, because the session would exceed the forty-five second
harness timeout and leak a `dlv` process; the two pattern mutants that exist
cover the semantics.

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
yet. Four verbs, one dispatch site, every tool a job.

## 4.9 Drive it yourself

Ungraded, and the one to do first, because this is the chapter where the
frustration from the end of Chapter 3 gets its answer.

Rebuild the binary and go back to the project you used it on:

```
cd solutions/ch04 && go build -o ~/bin/ch4agent .
cd ~/some/project
LLM_MODEL=claude-opus-5 LLM_API_KEY=sk-... ~/bin/ch4agent chat
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

Against the fake, for the plumbing, as before:

```
go run ./cmd/fakevendor -ch 4 chat
scripts/live.sh 4 anthropic rounds      # the dlv demo, scripted
```

Commit and tag when the grader passes:

```
git commit -am "ch4: jobs, grader 100"
git tag ch04-pass
```

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
that by making the pain specific. On 10 August I could not be told anything at
all. Today I can be told anything, three seconds from now.
