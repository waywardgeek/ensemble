# Chapter 4: Jobs

This chapter takes you from basic tools, which every coding agent has, to
tools done right, which as far as the author knows has only been done in
CodeRhapsody. By the end of the chapter your agent will interact with tools
the way a person does: drive a debugger, provide input on the fly,
and watch output arrive in real time. The LLM will not use this capability
often unprompted; Claude has launched `dlv` on its own perhaps twice. That is
because the tools used by the coding agents that trained Claude, Gemini, and
ChatGPT cannot do what yours will do by the end of this chapter. For the big
LLM providers, this is the most valuable chapter in the book: the capability
your agent gains here is the capability their models have never been trained
on.

Every tool call is a process you supervise, and the wait belongs to the
call, not the tool. `read_file` on an unmounted network share hangs
exactly the way `run_command` on a shell that wants input does. Once the
engine treats them the same, a whole class of bug becomes unwritable:
the one where a tool declared itself quick and was not.

Chapter 3 gave the agent six tools and a loop, but the loop blocked on
every tool call. That was fine for `edit_file`, where nothing is at
stake, and fatal for `screenshot`, where the screen itself may not
cooperate. Chapter 4 adds one idea: the tool call becomes a job. A
handle is allocated before the tool runs, output goes to disk under that
handle, and the engine waits three seconds before reporting whatever is
true at that moment. Nothing dies at three seconds. A job that is still
running is still running, and the model is told so, with the handle and
three verbs that act on it.

**What you build.** One dispatch site that turns every call into a job:

```go
type JobStatus int

const (
	StatusRunning JobStatus = iota + 1
	StatusDone
	StatusKilled
)

// WakeReason says why Execute stopped waiting. It picks the report
// shape: WokeDone on the first report is a bare Chapter 3 result;
// anything else is a job report carrying the handle.
type WakeReason int

const (
	WokeDone WakeReason = iota + 1
	WokeDelay
	WokePattern
)

// Limits are the three knobs a caller has. Pattern is compiled at the
// argument edge, so a bad regex fails the call that supplied it, not
// the job.
type Limits struct {
	Delay     time.Duration
	Pattern   *regexp.Regexp
	MaxOutput int
}

const (
	DefaultDelay     = 3 * time.Second
	DefaultMaxOutput = 16 * 1024
	IODir            = "cr/io"
)

// JobData is the log payload. Derived from Job, never stored primary.
type JobData struct {
	Handle   int       `json:"handle"`
	Status   JobStatus `json:"status"`
	Output   Ref       `json:"output"`
	Bytes    int       `json:"bytes"`
	ExitCode *int      `json:"exit_code,omitempty"`
	Reason   string    `json:"reason,omitempty"`
	Cwd      string    `json:"cwd,omitempty"`
}
```

The seam every tool is written against changes by one argument:

```go
type ToolFunc func(c *Call, args json.RawMessage) (string, error)

type Call struct {
	Job    *Job    // this tool's own job
	Jobs   *Jobs   // the store; the verbs that act on OTHER jobs
	Limits Limits  // resolved before the tool runs
	Events []Event // records for the dispatcher to append
}

type Tool struct {
	Name        string
	Description string
	Schema      json.RawMessage
	Run         ToolFunc
	NoJob       bool // supervision verbs and tool_limits run inline
}

const MaxToolRounds = 16

func (e *Engine) Execute(call ToolCallPart) error
func (e *Engine) Shutdown() error
```

`Execute` resolves limits, allocates a handle, records `ToolCalled`,
starts the tool on its own goroutine, waits by the limits, and reports
whatever is true. A running job's report names the handle, the byte
count, the path to the output on disk, and the three verbs that act on
it: `wait_for_job`, `send_input`, `kill_job`.

The events carry the job record on the `tool` payload, alongside
`call_id` and `name`. A `read_file` that finishes in time produces
two events:

```json
{"type":"tool_called","tool":{"call_id":"toolu_1","name":"read_file",
  "job":{"handle":1,"status":"running",
    "output":{"kind":"handle","locator":"cr/io/1"},"bytes":0}}}
{"type":"tool_returned","tool":{"call_id":"toolu_1","name":"read_file",
  "parts":[{"text":"line 3\nline 4\n"}],
  "job":{"handle":1,"status":"done",
    "output":{"kind":"handle","locator":"cr/io/1"},"bytes":14}}}
```

A killed job emits its own event type — no `tool` wrapper, because the
kill may happen long after the call returned:

```json
{"type":"job_killed","job":{"handle":1,"status":"killed",
  "reason":"shutdown"}}
```

Calls that are not jobs — `wait_for_job`, `send_input`, `kill_job`,
`tool_limits` — carry no `job` field at all.

Ten tools: Chapter 3's six plus `wait_for_job`, `send_input`,
`kill_job`, and `tool_limits`. `run_command` gains an optional `cwd`
(relative paths resolve against the workspace; a missing directory is
an argument error before anything starts) and runs under a PTY, because
a debugger needs a terminal.

**Rules.** Nine checks. The grader drives the binary through fake
vendors, reads the log via `dump`, reads `cr/io/<handle>` on disk, and
never reads source.

1. **Chapter 3 still passes.** Chapter 3's check bodies run unchanged
   against the Chapter 4 binary. A call that finishes in time returns
   exactly what it returned before. (`ch3parity`, 10)
2. **Every call is a job.** `run_command`, `read_file`,
   `list_directory`: each gets a handle at dispatch and its result is on
   disk at that handle after the call returns. Special-casing the shell
   scores zero; the read tools are the control. `cwd` is honored, is
   recorded in the log, and does not stick to the next call.
   (`jobmodel`, 25)
3. **`wait_for_job` blocks until done and returns at once on a finished
   job.** (`waitjob`, 10)
4. **`send_input` reaches a running process** and the response comes
   back on the next report. (`sendinput`, 10)
5. **The agent drives `dlv`:** breakpoint, continue, print a variable,
   quit. If `dlv` is not installed the check fails and prints the
   install line; it does not skip. (`debugger`, 5)
6. **`kill_job` stops the process and tells anything waiting on it.**
   The killed report never says done, finished, completed, or
   succeeded, and carries no exit code. (`killjob`, 10)
7. **Output is capped at dispatch, not inside `run_command`.** A 1.3 MB
   flood and a 1.1 MB `read_file` both land whole on disk; the model
   gets a stub naming the byte count and the path.
   `max_output_bytes: 2048` is honored. (`bigoutput`, 15)
8. **`tool_limits` binds exactly the next call, whatever it is.**
   Defaults otherwise; explicit arguments on the call win. The result
   of the call that consumed a pending limit names `tool_limits`; a
   bare call carries no such note. (`toollimits`, 10)
9. **Shutdown kills running jobs and the log says so.** (`shutdown`, 5)

Nine checks, sum 100: `ch3parity` 10, `jobmodel` 25, `waitjob` 10,
`sendinput` 10, `debugger` 5, `killjob` 10, `bigoutput` 15,
`toollimits` 10, `shutdown` 5.

**Yours.** Structure: one dispatch site, a handle for every call before
the tool runs, a wake at three seconds that ends nothing, output whole
on disk with a stub in context, `NoJob` verbs that act on other jobs,
limits consumed by the next call. Any implementation with these shapes
passes.

Taste, which the reference chose and nothing checks:

- **How to kill.** The reference sends `SIGKILL` to the process group
  with no `SIGTERM` grace. A killed `dlv` or database client never
  cleans up its terminal. The grace period costs one more state; add it.
- **A ceiling on `MaxOutput`.** The reference bounds it only below. A
  model that asks for a 500 MB inline return is refused by luck, not by
  a constant. Add the constant.
- **Which bytes survive truncation.** Head and tail, with the omitted
  middle marked and the path named. Head-only also passes; the tail is
  where the error message is.
- **The terminal.** PTY at 50x200, `TERM=dumb`, CRLF normalized. Any
  size passes; the environment variable is what stops cursor escapes
  landing in context.
- **A killed job's exit code.** The reference records none: the exit
  status of a process you killed is not news. The grader checks the
  report text, not the field.

**Exercise.** `make grade4`. Nine checks, 100 points. The agent drives
`dlv` to a breakpoint and reads a variable, which was impossible in
Chapter 3 for two independent reasons the body explains.

## 4.1 The idea in plain words

**The wait belongs to the call, not the tool.** Chapter 3's `Execute`
called a tool, blocked until it returned, and reported the result. That
is safe exactly when the tool finishes quickly, and the agent has no way
to know that in advance. A shell command that compiles a project takes
ninety seconds. A `screenshot` on a locked display never returns at all.
Both ran through the same `Execute`, and both froze the agent for the
duration. The fix is not a list of slow tools; the fix is to stop
pretending any tool is fast.

**A job is a handle and a goroutine.** `Execute` allocates a handle for
every tool call, runs the tool on its own goroutine, and waits three
seconds before looking at what happened. If the tool finished, the
report is the result. If it is still running, the report says so and
names the handle, the byte count so far, the path to the output on
disk, and the three verbs the model can use to act on it: wait longer,
send input, or kill it. The model is the supervisor.

**Three seconds is a wake, not a deadline.** Nothing dies at three
seconds. Nothing is cancelled. The engine wakes up, checks the state,
reports what it sees, and gives the model the information to decide.
Most tools finish before the first wake and the report is
indistinguishable from Chapter 3. The three-second default, the
pattern-match wake, and the output cap are all tuneable per call through
`tool_limits` or inline arguments, and the model can override every one.

**Output belongs on disk.** Every job's full output goes to a file at
`cr/io/<handle>`. The model gets a stub: the byte count and the path.
Chapter 3's `Ref` already knew how to say "the bytes are elsewhere."
Jobs apply it to time as well as size: a megabyte of test output lands
whole on disk, and the model reads as much as it needs.

**Supervision verbs are not jobs.** `wait_for_job`, `send_input`,
`kill_job`, and `tool_limits` operate on other jobs; they do not
become jobs themselves. They run inline on the dispatcher's goroutine,
because turning them into jobs would mean supervising the supervisor.
The distinction is one field on the `Tool` struct: `NoJob`.

## 4.2 Why the category is a trap

A tool that runs locally and deterministically looks like it returns
quickly, always. A tool that crosses a boundary the agent does not
control has no such property. Three examples, in ascending order of
how little control the agent has over the outcome: a shell command,
where the agent wrote the command but not the program it runs; a
network call, where neither the far end nor the path is under the
agent's control; and a third-party tool served over a protocol the
agent did not write, where the schema, the uptime, and the execution
are all someone else's.

Having named the category, the temptation is to use it: supervise the
boundary-crossing tools, leave `edit_file` alone. That is a list,
maintained by hand, and wrong the first time a "local" tool turns out
to have a boundary in it. `read_file` on a network mount that has gone
away does not return. `screenshot` on a locked display does not return.
Neither was ever on any list of slow tools, and neither needed to be:
the comment at the dispatch site states the operating rule. The cost
of uniformity is one function. The cost of the list was a shipped
defect.

A question worth answering before moving on: which of your agent's
tools have you assumed are quick?

> **Receipt.** On 10 August 2026 a coding agent froze on a
> `screenshot` call that never returned. The process had to be killed,
> and every sub-agent with it. The first fix was a watchdog with a
> per-tool timeout table; the second was a longer timeout for the slow
> tools. Bill ruled both wrong: the table was the bug. The wake at the
> dispatch site is what replaced it.

## 4.3 The handle comes first

`Execute`'s first act after resolving limits is to allocate a handle.
The handle exists before the tool runs, before the dispatcher knows
whether the tool is `run_command` or `read_file` or a name nobody
registered. That order is the structural commitment: the engine can
report on any call from the moment it starts, and a model that sees a
handle in the `ToolCalled` event can wait on it, send to it, or kill
it.

The alternative is to allocate the handle after the tool returns, as
a bookkeeping label for the result. The grader catches this: a call
with a handle on `ToolCalled` that appeared only on `ToolReturned`
scores zero, and so does a handle only for `run_command` while
`read_file` and `list_directory` get none.

`isError` at the dispatch site is true only for a finished job that
errored. A job that is still running is never an error, because
"still running" is not a failure; it is the reason jobs exist.

The tool runs on its own goroutine, and there is no `recover`. A
panic in a tool is an invariant violation, and an invariant violation
takes the process down loudly. `recover` converts a bug into a tool
error, and a tool error goes back to the model as content, and the
model tries again, and the bug is now a loop. The reference makes
the loud choice.

## 4.4 A wake is not a deadline

`WakeReason` has three values: `WokeDone`, `WokeDelay`, and
`WokePattern`. The report format depends on which one fired first.

If the tool finished within the delay, on the first report, and the
result fits: the report is a bare Chapter 3 result, byte for byte. A
model replaying a Chapter 3 transcript against a Chapter 4 engine
gets identical bytes. Any looser conjunction breaks one of two things:
Chapter 3's golden transcripts, if the report gains job vocabulary
that was not there before, or every `read_file` in the codebase, if
the absence of a handle in the report depends on a list of "quick"
tools rather than on what actually happened. The `ch3parity` check
is 10 points.

If the tool is still running or has produced more output than the
cap allows, the report carries the handle, the status, the byte
count on disk, and the path. A running report also names the three
verbs. `WokePattern` fires when the new output matches the pattern
the caller supplied, typically a prompt: `(dlv) ` for a debugger,
`ECHO_READY` for the test fixture.

`Limits` holds the three knobs: `Delay`, `Pattern`, and `MaxOutput`.
`Pattern` is a compiled `*regexp.Regexp`, not a string. The regex is
compiled at the argument edge, so a bad pattern is an argument error
on the call that supplied it, not a mid-flight surprise inside a job
that has been running for thirty seconds. `MaxOutput` is bounded only
below, by a positivity check; the reference has no upper ceiling, and
the outline names this as a taste decision worth reversing.

## 4.5 Three verbs and one setter

Four tools act on jobs rather than becoming jobs: `wait_for_job`,
`send_input`, `kill_job`, and `tool_limits`. All four are marked
`NoJob` on the `Tool` struct and run inline on the dispatcher's
goroutine.

`wait_for_job` takes a handle and waits by the same `Limits` the
dispatcher uses. On a finished job it returns at once; blocking on an
already-finished job is the single most common implementation error
in the exercise, and it surfaces as a deadlock the model cannot
diagnose. On a running job it waits until done, delay, or pattern,
whichever comes first.

`send_input` writes to the process's stdin. It is an error on a job
that has ended and on a job that has no process, and both errors say
which. The response the process writes goes into the job's output
buffer, and the next report carries the bytes the model has not yet
seen. A per-job cursor advances on every report so no byte is
reported twice.

`kill_job` sends `SIGKILL` to the process group and marks the job
killed. The killed report must not say done, finished, completed, or
succeeded, and must carry no exit code: the exit status of a process
you killed is not news. The grader checks both the text and the
absence of the field. If a waiter is parked on the job, the kill
wakes it; a 30-second wait on a killed job must return at once, not
block for the full delay. The `killjob` check's wall-clock bound
enforces this.

`tool_limits` sets the `Limits` for the next call, whatever it is,
and is consumed by that call. The limits are resolved in three
layers: defaults, then a pending `tool_limits`, then the call's own
arguments. Explicit arguments win. The result of the call that
consumed a pending limit names `tool_limits` in its text, so the
model learns where its setting went when the next call was not the
one it intended. A bare call with no pending limit carries no such
note. The `toollimits` check tests both sides: the note-or-no-note
negative control was a mid-build ruling promoted into a graded rule.

In the corpus, the three verbs together are under two percent of all
calls: `send_input` at 1.14% (800), `wait_for_job` at 0.46% (322),
`kill_job` at 0.32% (225). Under two percent, and the tools that
make the other ninety-two percent supervisable.

## 4.6 A megabyte of test output

The grader's `flood` fixture prints 20,000 lines, over 1.3 MB, to
stdout. The `bigoutput` check has three legs: the flood, a `read_file`
of a 1.1 MB planted file, and the flood again with
`max_output_bytes: 2048`. All three pass only when the full output
lands on disk, the model gets a stub that names the path and the byte
count, and the cap is applied at dispatch rather than inside
`run_command`.

The third leg is the negative control: a student who truncates inside
`run_command` has no truncation in `read_file`, and a student who
truncates in both has duplicated the logic in every tool when it
belongs in one place. `capText` in the reference does head-and-tail
truncation, keeping the first and last halves of the capped payload
with the omitted middle marked and the path named. Head-only also
passes. The tail is where the error message is.

Late results from a killed job are dropped rather than appended.
`Finish` only takes effect while the job is still running. Without
this guard, a killed job could gain bytes after its final report, and
the model would see a dead job whose byte count kept growing.

## 4.7 The terminal

`run_command` starts its process under a PTY via `creack/pty`. If the
PTY cannot be opened, that is an error, not a fallback to pipes. A
pipe fallback would mean `dlv` works until the day it does not, with
nothing in the architecture to say why; the pipe is the thing that
broke, and the error names it.

`TERM=dumb` in the child's environment stops programs from emitting
cursor escapes into captured output. The set value is what matters,
not the specific choice: without it, a `vim` session landing in the
agent's context would carry control sequences the model interprets as
text. CRLF normalization strips the `\r` a PTY produces, because the
model and the log and the grader all work in `\n`.

`cwd` is a per-call argument, resolved against the workspace. A
relative path becomes absolute; a missing directory is an argument
error before anything starts. The resolved path is recorded on
`JobData.Cwd`, omitted when empty. The next call without a `cwd`
argument runs in the workspace. There is no persistent shell, no
`set_cwd`, and no sticky working directory.

> **Receipt.** In one archived corpus, a majority of `run_command`
> calls began with `cd`. The exact proportion rests on a retired
> denominator and cannot be recomputed from the repository, but the
> shape is what matters: a model started in the wrong directory pays
> a tax on every call, and a sticky `cd` is state the model cannot
> see after compaction.

## 4.8 The agent drives a debugger

The `debugger` check is the chapter's closing demonstration, graded:
the fake drives the agent to start `dlv`, set a breakpoint at
`main.go:7`, continue to it, print `answer` (which equals 42), and
quit. Five calls: one `run_command` with
`ai_callback_pattern: "(dlv) "`, and four `send_input` calls each
waiting on the same prompt.

This was impossible in Chapter 3 for two independent reasons. First,
the loop blocked: a call that started `dlv` could not type into it
because `Execute` waited for the tool to finish, and `dlv` does not
finish until someone types `q`. Second, `dlv` refuses pipes: "Stdin
is not a terminal." The PTY fixes the second; jobs fix the first. The
combination is a capability Chapter 3 did not have.

The check fails, rather than skips, when `dlv` is not installed. A
check that skips is a check that never grades anyone, and a grader
that reports 100/100 with a footnote is a grader that trains the
reader to ignore footnotes. The failure message prints the
`go install` line.

## 4.9 Shutdown is the exit policy

`Shutdown` walks every running job and calls `Kill` with reason
`"shutdown"`. Each kill is recorded as a `JobKilled` event. The
`shutdown` check starts a process that never returns, lets the agent
exit, and reads the log for the event.

`Kill` on a job with no OS process marks the job killed and lets the
goroutine run on, because Go cannot cancel a goroutine. The status
changes, the waiters are notified, and the goroutine is now talking
to a dead job. This is documented, not defended: it is the cost of
goroutines, and it is real.

## 4.10 What this chapter cannot finish

Chapter 4 gives supervision by asking: start a job, and when you want
to know how it is doing, call `wait_for_job`. That is a complete,
shippable design and where the chapter stops.

The limit is worth naming now so it lands as a discovery rather than a
gap. While the agent is inside a tool call it cannot hear anything: not
a completion it did not ask about, and not you. The reason is
structural and survives making tools asynchronous: the tool runs on
its own goroutine, but the loop that dispatched it is parked waiting
for the answer, and that loop is the only thing that would notice an
inbound message. A supervised job fixes the freeze and leaves the
deafness. The agent can run a ninety-second test suite without dying
and still cannot be told "stop, wrong file" while it does.

Fixing that is not more job machinery. It is a different shape: one
inbound queue carrying prompts, interruptions, and job completions as
the same kind of thing, drained by a loop that never blocks on a tool.
That is the next chapter, and this one earns it by making the pain
specific.

## 4.11 What you would do differently

Load-bearing, and the grader fails for it: one dispatch site, a handle
for every call before the tool runs, a wake at three seconds that ends
nothing, output whole on disk with a stub in context, `NoJob` verbs
that act on other jobs, limits consumed by the next call, and the
consumed-by note.

The reference's taste, with its reasons. `SIGKILL` to the process
group with no `SIGTERM` grace: the grader checks that the process
stops, not how. A killed `dlv` never cleans up its terminal sockets;
a `SIGTERM` first, with a grace period, costs one more state and is
the thing to add. `MaxOutput` bounded only below: a model that asks
for 500 MB inline is refused by luck, not by a constant. Head-and-tail
truncation over head-only: the tail is where the error message is, and
head-only also passes `bigoutput`. PTY at 50x200, `TERM=dumb`, CRLF
normalized: any PTY size passes, and the environment variable matters
more than the dimensions. No `recover` around the tool goroutine: the
loud choice, untested by any check. No `list_jobs`: considered and
dropped, because no use was found.

No persistent shell. State that is invisible after compaction is state
that will be wrong after compaction, and a sticky `cd` is the example
that taught this. A long task that needs a shell starts `bash` as a
job, talks to it with `send_input`, and kills it when done. The shell
is a job like any other, and it has a handle.

CodeRhapsody, as of September 2026, uses the same dispatch site for all
53 tools. Its `tool_limits` is called `set_tool_watchdog` and takes a
reason string, so the model must say why this call needs longer than the
default. The constraint is a teaching device: a model that writes a
reason is less likely to raise the limit reflexively, and the reason
ends up in the transcript where a reviewer can read it.

## 4.12 The exercise, graded

Brief for the reader's agent: make every tool call a job in the
Chapter 3 binary. Add `wait_for_job`, `send_input`, `kill_job`, and
`tool_limits`. Put `run_command` under a PTY with `cwd`. Same
commands, same protocol, same log. `dlv` must be on the path.

```sh
make grade-dir CH=4 DIR=path/to/yours    # your client
make grade4                              # the reference
```

| check | points | passes when |
|---|---|---|
| `ch3parity` | 10 | all Chapter 3 checks pass against this binary |
| `jobmodel` | 25 | every call has a handle, a disk result, and `cwd` honored |
| `waitjob` | 10 | blocks on a live job, returns at once on a finished one |
| `sendinput` | 10 | input delivered, echo reported, cursor advances |
| `debugger` | 5 | `dlv` driven to `p answer`; absent `dlv` fails, not skips |
| `killjob` | 10 | process stops, kill recorded, report never says done |
| `bigoutput` | 15 | full output on disk, stub with path and byte count |
| `toollimits` | 10 | one-shot, defaults restored, explicit args win, note |
| `shutdown` | 5 | running job killed at exit, log records it |

Weights sum to 100. Every check is all or nothing. Exit status of the
grader: 0 pass, 1 fail, 2 could not run. Cost: nothing. The fake
serves all three vendors on one local port.

The grader's own tests work by deletion: 28 mutants of the reference,
each asserting the exact set of checks that fails. The mutant that
paid for itself: a killed job's goroutine setting the status to done
after the kill had already marked it killed. The kill scored as a kill,
the waiter reported done, and the grader scored 100 out of 100 because
its regex matched the word "killed" in the note. The `KillEvent.Status`
field exists because that mutant was run.

## 4.13 Drive it yourself

Export a real key and a model ID from the vendor's models endpoint, set
`LLM_VENDOR`, and run `./ch04`. Ask the agent to run a test suite, then
check its output with `wait_for_job`. Start a long compilation and kill
it. Run `./ch04 dump` and look at the handles: every tool call has one,
and every handle has a file at `cr/io/<handle>`.

Then try the debugger. `PAGER=cat dlv debug ./testdata/dbg`, a
breakpoint at `main.go:7`, continue, `p answer`, and quit. The agent
that could not hold a conversation with a running process in Chapter 3
does it in five calls. The difference is two things: a terminal and
jobs. Both had to be true.

The agent can now supervise long-running processes, interact with
programs that wait for input, and kill work that should not continue.
It is still deaf while doing so: a tool call occupies the loop, and
nothing else gets in. The next chapter fixes that by giving the
engine an inbound queue, so that a prompt, an interruption, and a job
completion enter by the same door.
