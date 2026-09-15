# Chapter 4 — Jobs: Containment, Not Cancellation

*Status: outline, second pass — reworked after the build. The solution exists
(`solutions/ch04`, tag `ch04-solution`) and the grader is 100/100 with a
22-mutant audit; see `book/review-ch04-code.md`. Bill ruled live during the
build and his rulings supersede the first pass of §4.6 and §4.7 (blocking
contracts and budget policy are gone). War story is receipted — commits
`286bfd84` and `4bec4b96`, both 10 August 2026, `internal/agent/watchdog.go`.
Architecture is receipted — `cr/docs/tool-calls-as-jobs-design.md`. Shell
state across calls is RULED (§4.7, with corpus numbers); the reference still
needs the `cwd` argument and the `jobmodel` legs that grade it — coder brief
pending.*

---

## Voice plan (voice.md §9.1)

**Through-line stake:** on 10 August 2026 the narrator froze on a
`screenshot` call and Bill had to kill the process, taking every sub-agent
with it; the narrator shipped a fix in twenty-two minutes, shipped a fix to
the fix twenty-two minutes after that, and Bill later ruled both wrong. The
question that stays open the whole chapter is whether an agent can be made
unable to take itself down without anyone maintaining a list of which tools
are dangerous. Resolves at §4.7 (the containment sentence survives in exactly
one place) and §4.8 (the agent drives a debugger, which yesterday was
impossible). Bill enters at §4.1 ("even `read_file` can hang") and rules at
§4.6 and §4.7.

**Wild fact per section:** §4.0 the call that froze the agent was a
screenshot; §4.1 `read_file` on a departed NFS mount does not return; §4.2 Go
cannot kill a goroutine, and the channel is buffered so the watchdog does not
cause the leak it exists to bound; §4.3 the result of an abandoned call still
arrives and the first fix threw it away, and every tool became a job by
changing one function; §4.4 `send_input` outnumbers `kill_job` three and a
half to one; §4.5 `flood` is 1,340,013 bytes and you pay for them on every
later turn; §4.6 fix and fix-to-the-fix at 14:28 and 14:50, `send_secret`
missing from the map; §4.7 `Stdin is not a terminal`, and 69.6% of my
`run_command` calls begin with `cd`, 7,979 of them into the same directory;
§4.8 driving `dlv` was impossible in two independent ways; Exercise: a killed
job scored 100 because the kill note contained the word "killed", and the
never-returning fixture died of the deadlock detector.

**Confession budget (3):** §4.2/§4.3 the watchdog threw away a live result;
§4.6 the declaration was the wrong repair; Exercise
`killed-job-reported-as-done` scored 100. The classification trap in §4.1 is
stated as a fact about the instinct, not a confession. The `select {}`
fixture is a fact.

---

## What this chapter is

Chapter 3 built six tools in their simplest honest form, and all six share one
shape: you call the tool, you block, you get a result. That shape is not a
property of those six tools. It is a property of how chapter 3 *dispatched*
them, and it is wrong for all of them.

The proof is the cold open below. The call that froze the agent was a
`screenshot`. Not a shell command, not a network fetch. A tool whose whole job is
to grab the framebuffer and return, which on paper cannot be slow. If the defect
lived in `run_command`, that hang was impossible.

So this chapter does not rework one tool. It changes what a tool call **is**: not
a function you call and wait for, but a **job you start and supervise**. Every
tool call gets a handle, an output file, and a status. The three verbs it
introduces (`wait_for_job`, `send_input`, `kill_job`) are the supervision API for
all of them, and one setter (`tool_limits`) reaches the tools whose arguments
you do not own.

**Chapter 4 still does not add reach.** Those verbs are together under two
percent of all tool calls, and your agent cannot touch one thing it could not
touch before. It earns its place twice over: it makes the biggest tool you
already built (`run_command`, thirty-six percent on its own) genuinely usable,
and it makes *every* tool incapable of taking the agent down with it.

**Nothing in this chapter dies on its own.** No deadline, no budget, no
watchdog. A job runs until it finishes or until the model kills it, and the
only thing the model decides per call is *how long it will wait before
looking*. That inversion — the wait is the caller's, not the tool's — is the
chapter's design idea, and §4.6 is where it is argued.

The chapter's central distinction is which tools can report progress and which
cannot. A subprocess emits bytes over time, so you can watch it and match a
pattern against what it says. A `read_file` produces everything at the instant it
finishes or nothing at all. Both are jobs. Only one can be *watched*, and a
design that pretends otherwise ships a parameter that silently does nothing.

And it ends by doing something your chapter 3 agent could not do at all, not
slowly but *at all* — and that is measured, not asserted (§4.8).

### What this chapter deliberately cannot finish

Chapter 4 gives you supervision by **asking**. You start a job, and when you want
to know how it is doing, you call `wait_for_job`. That is a complete, shippable
design and it is where this chapter stops.

It is not the whole story, and the limit is worth naming now so that it lands as
a discovery rather than a gap. While your agent is inside a tool call it cannot
hear anything. Not a completion it did not ask about, and not *you*. The reason
is structural and it survives making tools asynchronous: the tool runs on its own
goroutine, but the loop that dispatched it is parked waiting for the answer, and
that same loop is the only thing that would notice an inbound message.

So a supervised job fixes the freeze and leaves the deafness. Your agent can run
a ninety second test suite without dying, and still cannot be told "stop, wrong
file" while it does. Fixing that is not more job machinery. It is a different
shape: one inbound queue carrying prompts, interruptions and job completions as
the same kind of thing, drained by a loop that never blocks on a tool.

That is the actors chapter. This chapter earns it by making the pain specific.

---

## §4.0 Cold open — the call that never came back

10 August 2026. A `screenshot` call did not return.

Not crashed. Not errored. From the outside the agent looked busy, because by
every measure available to it, it was:

> A handler that blocked forever took the whole agent with it: not crashed, not
> errored, just gone. The turn never completes, no observer fires,
> `agents_status` keeps reporting "processing", and a supervisor waiting on it
> waits forever.
>
> — commit `286bfd84`

The only recovery was killing the process. That kills every agent in the tree,
including sub-agents doing unrelated work that was going fine.

> For one interactive session that costs an afternoon. For a fleet it is fatal,
> because the entire supervision model assumes turns end.

Until that afternoon, no tool call in the system had any wall-clock bound at
all. Not because anyone decided against one. Because a tool call looks like a
function call, and nobody puts a timeout on a function call.

**Voice note:** the cold open is the incident, told flat. No "little did I
know." The comedy in this chapter is in the fix being embarrassing, not in the
failure being dramatic.

---

## §4.1 Why this is not a rare event — and why you must not classify

The instinct is to file that under bad luck. It is not. It is the predictable
consequence of a category error, and the category has a name.

A tool that runs locally and deterministically — read a file, edit a file, list
a directory — *looks* like it either returns or fails, fast, always. A tool that
crosses a boundary you do not control has no such property. Three examples, in
ascending order of how little control you have:

1. **A shell command.** You wrote the command, but not the program it runs.
2. **A network call.** You control neither the far end nor the path to it.
3. **An MCP tool.** Someone else's service, someone else's schema, someone
   else's uptime — and the tool most likely to wedge:

> it cannot cover MCP tools at all, since those are third-party schemas we do
> not control — and they are both the likeliest to wedge (a browser call on a
> page that never settles)
>
> — commit `286bfd84`

Here is the trap, and the first pass of this outline walked straight into it:
having named the category, you want to *use* it. Supervise the boundary-crossing
tools; leave `edit_file` alone, it has never once failed to return. That is a
list, and you will maintain it by hand, and it will be wrong the first time a
tool you filed under "local" turns out to have a boundary in it. `read_file` on
an NFS mount that has gone away does not return. A `screenshot` on a display
that has been locked does not return. The cold open *is* a tool that was on the
wrong side of that list.

**The chapter's operating rule:** you cannot tell from a tool's name which side
of the boundary it is on, so do not try. Every tool call is a job. The cost of
uniformity is stated in §4.3 and it is one function. The cost of the list is
§4.6, and it was a shipped defect.

*(Forward note, not to be stated in the prose: MCP is chapter 6. This section
plants the category without naming the chapter.)*

---

## §4.2 The fix that does not fix it

What I shipped first — and what the student would ship first — is four lines,
and the student will be disappointed by how few:

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

**The agent survives.** That is the whole win, and it is a large one.

Now the part most books would skip, and this chapter must not:

> **Go cannot kill a goroutine.** When the watchdog fires we STOP WAITING; the
> handler keeps running until it returns on its own or the process exits. This
> is containment, not cancellation, and the tool result says so in those words.
>
> — commit `286bfd84`

The timeout does not stop the work. It stops *waiting* for the work. The hung
`screenshot` call is still hung; it now has no one listening.

**The chapter's title is this sentence**, and the discipline it demands is that
the report says so *in those words*. A message claiming the call was cancelled
would be a lie that reads like a success. In the design this chapter actually
builds, the sentence survives in exactly one place — `kill_job` on a job that is
a goroutine and not a process (§4.7) — because everywhere else the design stops
*abandoning* anything. Which is §4.3.

**One detail worth a paragraph**, because it is the kind of thing that is
invisible until it bites: the channel is buffered with capacity one, deliberately.

> on the timeout path nobody reads, so an unbuffered send would park the
> abandoned goroutine forever — the watchdog causing the very leak it exists to
> bound.

A safety mechanism that leaks is worse than none, because you stop looking.

---

## §4.3 You are throwing away a live result

Here is the turn, and it is the best idea in the chapter because it costs
almost nothing.

Look at what you just built. The handler runs on its own goroutine. Its result
is sent to a buffered channel. On the timeout path you walk away — **and the
handler keeps running, finishes, and sends its result to a channel nobody is
reading.**

> **Every tool call is already asynchronous.** There is no "make tools async"
> project. It is done. […] the result of an abandoned call still arrives — we
> simply throw it away.
>
> — `tool-calls-as-jobs-design.md`

So the job system is not a concurrency project. It is one sentence:

> Tier 1 is not "add concurrency". It is: **stop discarding the result.**
> Register the pending call, keep the channel, hand the model a handle.

And once you have the handle, the timer stops being a deadline. It is a
**wake**: the moment the dispatcher hands control back so the model can look.
Nothing was abandoned, nothing was killed, and a default of three seconds is
safe *because* it kills nothing. The four-line `select` is still the mechanism;
the comment `// abandon` becomes `return handle`.

Three consequences the prose should draw out:

1. **The handle goes at the dispatch site**, not in the tools. One funnel, which
   is already where security is enforced, so nothing bypasses it. Measured on
   the reference: every tool became a job by changing **one function**
   (`Engine.Execute`); five of the six chapter 3 tools changed by one ignored
   parameter. Before the tool runs: handle `N` (integers from 1, one more per
   job, per process), file `cr/io/N`, status `running`. The dispatcher waits
   **on the job**, not in the tool.

2. **The scary error message evaporates.** Before, an abandonment had to return
   an essay explaining that the handler was still running, might still land side
   effects, and must not be naively retried. That warning existed *because there
   was no way to re-attach*. With a handle it becomes: *still running, handle 47,
   call `wait_for_job(47)`.* **Give the reader a way back and the warning is not
   needed.** That is worth stating as a general principle about error messages.

3. **Every tool's output is now on disk**, which fixes an asymmetry the student
   has been living with since chapter 2 without noticing: a redacted
   `run_command` result could cite a file path, and every other tool's stub could
   only say "re-run it."

**The verbatim-first-report rule**, stated in the prose because it is a
contract and it is invisible otherwise: *a job that finishes within the wake
delay, under the output cap, is reported on its first report as the bare
result* — byte-identical to what chapter 3 returned. This is what makes
`ch3parity` free, and it is what makes the change additive in the architectural
sense: nothing that worked yesterday reads differently today unless it took
longer than three seconds or said more than sixteen kilobytes.

**Log contract** (students must not guess it): `tool.job = {handle, status,
output, bytes, exit_code?}` on both `tool_called` and `tool_returned`; `status ∈
running | done | killed`; `output` is chapter 2's `Ref{RefHandle, "cr/io/N"}` —
the first use of the kind chapter 2 declared for exactly this. The job field is
**absent** on the four supervision tools; they are the only calls that are not
jobs. New event `job_killed`, carrying the job and `reason ∈ kill_job |
shutdown`.

---

## §4.4 Three verbs and one setter

| tool | calls | share | what it is for |
|---|---|---|---|
| `send_input` | 800 | 1.14% | the process is waiting for you |
| `wait_for_job` | 322 | 0.46% | look again, up to a delay |
| `kill_job` | 225 | 0.32% | stop it |

**`send_input` is the biggest of the three, by more than two to one**, and that
ordering is the finding. The common intuition is that job control is mostly
about *stopping* runaway work. In practice it is mostly about *talking to*
work that is going fine and is waiting for an answer.

That is the capability the chapter is really buying, and §4.8 spends it.

`wait_for_job` must work on a job that has *already finished*. A student who
implements it as "block on the channel" will hang forever on a completed job,
and will diagnose it as a deadlock in their own code rather than a missing case.

`kill_job` is SIGKILL to the process group, and the killed job is reported as
`killed` — never as `done` with a strange exit code. The grader learned this the
hard way (§4.9).

The fourth tool is not a verb and has no corpus count: **`tool_limits`** sets the
wait for the *next* call, for tools whose arguments you do not own. §4.6.

`list_jobs` is not built. Zero measured use.

---

## §4.5 Output by the megabyte

Chapter 2 declared `Ref{Kind, Locator}` and has not used `RefHandle` since. Here
is where that debt is collected.

`go test ./...` on a real repository produces more text than you want in a
context window, and you pay for those tokens on every subsequent turn of the
conversation, not just the one that ran the command. The rule:

- The full result text goes to `cr/io/<handle>`, always, for every tool.
- What enters the context is the same bytes if they fit under the cap
  (**16 KiB**), else head + `[... K bytes omitted; full output (T bytes) at
  cr/io/N ...]` + tail.
- The path is not decoration. It is the recovery route — the model can read a
  range of it with the `read_file` it already has.
- **A report never repeats bytes.** Each job carries a cursor; `wait_for_job`
  and `send_input` return what the model has not yet seen. `ai_callback_pattern`
  is matched against unseen output only — a prompt the model already saw does
  not wake it twice.

**Do not let the model choose the truncation.** The truncation happens on the
way in, at the dispatch site, before anything reaches the context window. A
model asked to summarize its own flood has already paid for the flood.

The `Ref` is recorded in the log and **never rendered as a part**. Chapter 2's
Anthropic and OpenAI renderers refuse blobs with a loud `not implemented`, and
that refusal is the lesson, not a gap: the model gets a path and a tool that
reads paths. (Chapter 5's seam must not promise all-vendor blob rendering; this
is where that constraint is earned.)

---

## §4.6 Who decides how long to wait

Twenty-two minutes after shipping the watchdog, I shipped a second commit fixing
it. Both are dated 10 August 2026 and the timestamps are 14:28 and 14:50.

The watchdog needed to know how long each tool may legitimately block. The first
version *inferred* it: a hand-maintained map of tool names, plus a scan of the
tool's description text hunting for documented timeouts. `send_secret` accepted a
deadline argument and was missing from the map, so it silently got the default.
Tools whose own default wait was longer than the watchdog's would have been
abandoned mid-wait while perfectly healthy. The second commit fixed both by
making every tool *declare* its blocking contract, and this outline's first pass
taught that declaration as the ruling.

Bill's ruling during the build is that the declaration was the wrong repair,
because the map was the wrong question. **How long may this tool block?** is a
property of the tool, so it needs a table, so the table can be missing a row.
**How long will I wait before I look?** is a property of the *call*, so the
caller supplies it, so there is nothing to declare and no row to be missing
from. The `send_secret` defect does not get *caught*; it becomes *impossible*.

Three limits, and they are ordinary arguments on the tools that wait:

| limit | default | on |
|---|---|---|
| `ai_callback_delay` | 3 s | `run_command`, `wait_for_job`, `send_input` |
| `ai_callback_pattern` | none | same three |
| `max_output_bytes` | 16 KiB | same three |

**Precedence, later wins:** defaults → a pending `tool_limits` → the call's own
arguments. `tool_limits` exists because most tools have nowhere to put these
three arguments: the five local tools you own but never gave them to, and
every tool whose schema you do not own at all (an MCP server's, chapter 6). So
the model sets them the call *before*. It is **one-shot** — consumed by the very
next call, whatever that call is, including a `kill_job` or another
`tool_limits`. The friendlier rule ("applies to the next job-creating or
waiting call") is one line of code and a category in the model's head; a limit
set several calls ago and still pending is the persistent escape hatch the
first pass warned about, wearing a friendlier name. *The next call* is the rule
a model can hold.

**The footgun in a one-shot setting is a silent misfire**, and a live model
named it unprompted: "it's easy to burn it on the wrong thing." Set
`max_output_bytes: 200000` for the `read_file` you are about to make, glance at
a directory first, and the limits went to `list_directory`; the symptom is a
truncated read one call later with no cause in sight. The rule stays — a target
argument is a declaration the setting must match, which is the shape this
section just spent a page removing — but the misfire must not be silent. The
consuming call's result begins with one line:

    [tool_limits consumed by this list_directory call: ai_callback_delay 3s,
     ai_callback_pattern none, max_output_bytes 200000]

on every tool including the four that are not jobs (a second `tool_limits`
consumes the first and says so), and the `tool_limits` reply itself says the
next call, whichever tool that is, will report the consumption. A burn is now an
error you read in the result it caused, one round trip from the correction.
This is the book's rule about failures applied to the book's own mechanism, and
it pre-empts the objection a careful reader raises on their own.

The one-shot, capped, reasoned override from the first pass
(`set_tool_watchdog`) is **not built** and is not the same idea. It suspended a
guard. `tool_limits` raises a wake. There is no guard to suspend.

**What the model changes on every monitoring call**, because the wait is on the
call: it can wait 0.2 s to see whether a job started, 30 s for a test suite, or
"until `(dlv) ` appears" for a debugger — and change its mind on the next call.
No table anywhere knows what the tool needs, because the table would be wrong.

**Shutdown is the exit policy.** At process exit, every running job is killed and
`job_killed{reason: shutdown}` is logged. A job that never returned is not a
leak the process carries to the grave; it is a recorded kill.

*(The three mutation-test failures inside `4bec4b96` — the guard computed by
the thing it guards, the vacuous `t.Skip`, the mutation that did not land — are
parked for the testing chapter. The sentence worth keeping here is only: **a
ruling written down and not wired up is invisible**. It is the reason the
verbatim-first-report rule is in the prose *and* in `ch3parity`.)*

---

## §4.7 The terminal

`run_command` runs under a pseudo-terminal. This is not in the first pass; it is
Bill's ruling, modeled on CodeRhapsody's `command_executor.go`, and the chapter
has a receipt of its own for why:

> `Stdin is not a terminal`
>
> — `dlv`, on pipes, refusing to start

That is the closing demonstration failing before it begins. Programs that
prompt — debuggers, REPLs, anything that asks `y/N` — check whether they are
talking to a terminal and either refuse or stop flushing. A PTY is not a nicety
for `send_input`; it is the capability, measured.

**Costs, stated in the code and in the prose:** one merged stream (stdout and
stderr are the same bytes), the model's input is **echoed** back in the output
(that is the terminal's line discipline, and it is how the model sees its own
keystroke land; turning echo off changes what `dlv` shows, so it stays), `\r\n`
normalized to `\n`, `TERM=dumb`, 50×200. The exit status is the file's **last**
line (`exit_code: N`), because a stream cannot put it first; chapter 3's regex
still matches. No PTY available → an error. **Never a silent fallback to
pipes**, which would ship an agent whose debugger works on one machine and
refuses on another with no message saying why.

`kill_job` on a process is SIGKILL to the process group. `kill_job` on a job that
is only a goroutine — a `read_file` that never came back — marks the job killed
and the report says, in those words, that Go cannot stop it. That is the
containment sentence from §4.2, and it is the only place it survives, because it
is the only place the design still cannot do better.

No `recover` around the tool goroutine. A panic in a tool is an invariant
violation, and an invariant violation taking the process down is working as
intended.

Unix only, via the process group. Chapter 3 already was, via `sh -c`.

### Does shell state persist between calls? — RULED: no, and here is the receipt

The first pass declined "what happens when a job outruns its budget." There is
no budget, so there is nothing to decline; the coder's candidate replacement
(how a report spells a signal death) is three spellings of one fact with nothing
downstream depending on the answer, and is rejected. The question that *is*
real, that students will hit, and that shipping agents answer differently:

**Does shell state persist between `run_command` calls?** Claude Code's Bash
tool says yes — one shell, `cd` sticks — and its docs issue #45478 is the cost
made visible: a `cd` outside the approved directories is silently reverted and
a "Shell cwd was reset" notice appended, because the shell now holds state the
security boundary has to police. Editor-style agents ship named terminal
panes whose state persists. The reference says no. This chapter **rules**
rather than declines, because the job model makes it not a close call:

1. **A persistent shell is a job.** `run_command "bash --norc"` with pattern
   `\$ `, then `send_input "cd lyric"`, `send_input "make"`. The model gets a
   shell whose state sticks — *and it has a handle.* Every `send_input` to that
   handle is a logged event in order, so the log still reproduces the session.
   Claude Code's problem was never state; it was **implicit** state with no
   handle to attribute it to. Those named terminal panes are handles with a
   GUI. Isolation is the primitive; persistence is composed on top of it,
   explicitly. You can build a persistent shell out of isolated jobs. You
   cannot build isolation out of a persistent shell.
2. **Overlapping jobs cannot share one shell.** Two `run_command`s in flight
   at once — the point of this chapter — is not a thing one bash does.
3. **The measured tax is a *directory* tax, not a state tax.** On 26,781
   archived `run_command` calls, **69.6% begin with `cd`** (18,651), and
   almost all name one directory: the project root (`cd ~/projects/forge`
   7,979 times; `lyric` 1,908; `coderhapsody` 1,240). Environment setup —
   `source …/activate`, `export`, `nvm use` — is **0.12%** (33 calls). That is
   not a model navigating. That is a model started in the wrong directory and
   correcting for it on every call, forever.

So the remedy for the 69.6% is not persistence. It is two things, neither of
which is state:

- **`cwd` as an argument on `run_command`**, for this call only. Relative
  paths resolve against the workspace; a missing directory is an error, never a
  silent fall back to the workspace. The effective directory is recorded on
  the job and printed in the report when it is not the default — chapter 2's
  rule, record and never infer.
- **The default cwd is a launch-time setting**: the agent's workspace, which
  is not the same directory as the one holding the event log. (CodeRhapsody
  shipped exactly this the morning after the numbers were measured:
  `./coderhapsody [workspace]` with `--data-dir` anchored to the launch
  directory, commit `48839be8`.)

**Why no `set_cwd` tool.** It is the `tool_limits` argument again, with a worse
failure mode. A sticky default set at call 40 and compacted away by call 300
is state the model can no longer see but still acts on — and a wrong wake
costs seconds, a wrong directory runs `rm -rf build` in the wrong tree. To make
it safe you would have to promote sticky settings to the never-dropped category
in the reducer. **One-shot settings never need to survive compaction; sticky
ones always do.** That sentence is the general rule, and it is why both of
this chapter's knobs are per-call.

**Grader:** `jobmodel`'s fixture gains a `cwd` leg (a `run_command pwd` with
`cwd: testdata` must print the subdirectory as a line of *output* — a
substring test is satisfied by the report's own header, a mutant of the
reference passed exactly that test in CodeRhapsody) and a no-persistence leg
(the next call without `cwd` prints the workspace). No new check id; the
points stay at 25. Chapter 4 has no P6 declined decision; P6 is a norm, not a
per-chapter quota.

---

## §4.8 Your agent drives a debugger

The closing demonstration, and it is not a flourish. It is the proof.

```
run_command "PAGER=cat dlv debug ./testdata/dbg"   ai_callback_pattern="\(dlv\) "
send_input  "b main.go:7"
send_input  "c"
send_input  "p answer"
send_input  "q"
```

**Why this is the right ending:** a blocking tool call returns when the process
exits. A debugger does not exit until you tell it to quit. With chapter 3's
tools, driving `dlv` is not slow and it is not awkward — it is **impossible**,
and impossible in two independent ways: the dispatcher would wait for an exit
that never comes, and on pipes `dlv` refuses to start at all. The chapter's
thesis stops being a matter of taste and becomes a capability boundary the
reader can stand on either side of.

It also teaches `ai_callback_pattern` honestly. You do not sleep for a guessed
interval and hope the prompt has appeared. You wait for the string `(dlv) `,
because that is the actual signal that the debugger is ready for input. **A
fixed delay is a race condition with a comfortable name.**

**Receipt:** the grader drives this through the fake vendor (five scripted
calls, `42` read off the breakpoint). `scripts/live.sh 4 gemini rounds` drove it
live on `gemini-3.8-flash`: five `tool_called` events, every one carrying the
pattern, and the model read `42` unprompted. The student's agent can now debug
the code it wrote. That is the note to end on.

---

## §4.9 The exercise

**Contract:** JSON-lines transcript on stdin, exactly chapter 3's contract
(`ch3parity` requires it) — no network, deterministic, same fake vendor as
chapters 2 and 3. Handles are integers from 1, one per job, per process; the
fixtures depend on it (`wait_for_job(1)`), so it is stated as a contract rather
than read back from the log.

**Prerequisite:** `dlv` on `PATH` or in `$(go env GOPATH)/bin`. The `debugger`
check fails with the `go install` line when it is absent; it does not skip. (In
the preface's "What you need".)

**What the fixture provides**, all Go helpers built to binaries once per run,
never `go run`:

- A command that **finishes fast** (`go version`).
- A command that **outlives the wake** (`sleeper`).
- A command that **never returns** (`blocker`) — the `screenshot` incident,
  reproducible. *Not* `select {}`: with no other goroutine the Go runtime calls
  that a deadlock and exits 2, and the first version of this fixture died on
  its own while the check waited for it to be alive. It sleeps in a loop, and
  the reason is in its source.
- A command that **prompts and echoes** (`echoer`), so `send_input` has
  something real to talk to.
- A command that emits **more than a megabyte** (`flood`, 1,340,013 bytes).
- A one-line program with a breakpoint to hit (`testdata/dbg`, `answer := 42`).

### Checks — ratified as built

| id | points | what it grades |
|---|---|---|
| `ch3parity` | 10 | chapter 3's whole grader passes against the chapter 4 binary |
| `jobmodel` | 25 | handle on `tool_called` **before** the tool ran, for `read_file` and `list_directory` as well as the shell; `cr/io/N` bytes **equal** the result the model saw; `wait_for_job`'s own record carries no job; `cwd` honoured as a line of `pwd` output and **not** persisted to the next call *(legs pending — see §4.7)* |
| `waitjob` | 10 | delay honored (`running` at 0.2 s); wait returns the result; **works on an already-finished job**; file ends with marker and exit code |
| `sendinput` | 10 | pattern wakes on the prompt and not the echo; input reaches the process; its reply comes back; clean exit recorded |
| `debugger` | 5 | the fake drives `dlv` to a breakpoint and reads `42` |
| `killjob` | 10 | `job_killed{reason kill_job, status killed}`; the post-kill wait is told **killed**, never `done`/`exit_code`; a 30 s waiter returns early; the pid is dead |
| `bigoutput` | 15 | full output on disk with first and last line; inline ≤ cap, names locator and exact total; `max_output_bytes 2048` honored; `read_file` of a 1 MiB file truncated the same way — **at dispatch** |
| `toollimits` | 10 | `tool_limits` is not a job; applies to the next call; one-shot; pattern form wakes; the call's explicit argument wins; **the consuming call's result names `tool_limits`** on both a job tool and a non-job tool (a second `tool_limits`), and the call after it does not |
| `shutdown` | 5 | a job still running when the model stops is killed at exit and `job_killed{reason shutdown, status killed}` is logged |

**Sum: 100.** The code sums itself (a test adds the checks' declared points);
the awk over this table is a cross-check, not the instrument.

**Weighting rationale:** `jobmodel` is 25 because it is the chapter, and because
the common wrong answer — special-casing `run_command` instead of changing the
dispatch site — passes a naive test and fails the moment an MCP tool wedges.
Both negative controls (`jobmodel`'s and `bigoutput`'s `read_file` legs) are
**measured** load-bearing: delete either and the "it's a shell problem" student
scores 100. `toollimits` is 10 because it grades three properties and the
forward reference to chapter 6 in one fixture. `debugger` is 5, not more,
because `sendinput` already grades the mechanism; the five points buy the proof
that the mechanism reaches a real program. `shutdown` is 5 because it is one
behavior, and because without it the never-returning job is a leak the grader
would otherwise have to hunt with `kill -0`.

### Grader audit (P9)

- Delete each protected behaviour from the reference and confirm the score
  drops. **A row reading 100 → 100 is the finding.** This chapter's:
  `killed-job-reported-as-done` scored 100 on the first run. The killed process
  died, its goroutine ended the job as `done` with `exit_code -1`, and the
  waiter was told *"job 1 done, exit_code -1 … [job 1 killed: kill_job]"*. The
  check regexp'd the word *killed* — and the kill **note** supplied it. **A text
  regex can be satisfied by the very message that documents the failure.** Fix:
  grade structure — `job_killed` carries `status`, which must be `killed`; the
  post-kill report must not claim `done|finished|completed|succeeded|exit_code`.
- **Assert every mutation actually landed — and changed bytes.** This project
  has now been bitten three times; the chapter 4 rig refuses a regexp that
  matches other than once *or* whose replacement is a no-op.
- A mutant of a module **with dependencies** needs the dependency: the mutant
  `go.mod` carries the `creack/pty` require and the course `go.sum`, or the
  mutant fails to build and the failure is misread as detection.
- Watch for the guard-computed-by-the-thing-it-guards shape. From `4bec4b96`:
  a schema-scan test iterated over *what the scan found*, so shrinking the scan
  shrank the set under test. **A guard derived from its subject checks nothing.**
- A `t.Skip` on an empty scan is *"a vacuous pass with better manners."* Treat a
  skipped check as a failed check. The `debugger` check fails, not skips, when
  `dlv` is absent.
- Two mutants deliberately not written, with the reason recorded: pipes-for-PTY
  (the textual change is too large for a regexp; the evidence is the measured
  refusal plus `debugger` passing) and pattern-never-wakes (the session would
  exceed the 45 s harness timeout and leak `dlv`; the two pattern mutants that
  exist cover the semantics).

**Environmental assumptions, stated so a failure elsewhere is diagnosable:**
`ch3parity` holds because chapter 3's fixtures finish inside the 3 s wake
(`go run ./testdata/exit7` measured 0.05–0.3 s warm, 1.3 s cold). A machine
where a cold `go run` of a one-line program exceeds 3 s would fail `ch3parity`
with a `running` report where an exit code was expected — the fixture assumed
blocking; the wake is behaving. Cheap mitigation for the harness (recommended,
not required): run `exit7` once before grading to warm the build cache.
`creack/pty` measured on macOS only; Linux assumed.

---

## What exists now for a later chapter

| thing | state after ch4 | collected in |
|---|---|---|
| `Interrupted` as a turn state | not built | Ch5 |
| the mailbox | not built | Ch5 — the agent is still deaf while a job runs |
| jobs outliving a turn | not a question yet — one transcript is one process; `Shutdown` at exit | Ch5, when turns and processes come apart |
| `Ref{RefHandle}` recorded, never rendered | the reason ch5's seam cannot promise blob rendering everywhere | Ch5 |
| MCP tools | named as the worst case; `tool_limits` is built for them | Ch6 |
| the shell | still contains every tool | Ch8 |

---

## Ruled

1. **Blocking contracts (`DeadlineArg`/`MaxBlockingTime`) and budget policy —
   gone** (Bill, live). Nothing dies unless killed. The wait belongs to the
   call. `declare` and `budgetchoice` retired; `toollimits`, `shutdown`,
   `debugger` replace them.
2. **Every tool call is a job** except the four supervision tools. "Even
   `read_file` can hang if an NFS mount is unmounted."
3. **Three seconds** is the default wake, because it kills nothing.
4. **`tool_limits` is consumed by the next call, whatever it is.** The
   category-shaped alternative is rejected (§4.6).
5. **Verbatim-first-report rule** goes in the prose as a contract (§4.3).
6. **PTY echo stays** and is stated as a cost (§4.7).
7. **`dlv` is a prerequisite**, added to the preface.
8. **Handles from 1 per process** is a stated contract, not read back from the
   log.
9. **Signal-death spelling** rejected as a P6 candidate — representational.
10. **`list_jobs` not built.**
11. **Shell state does not persist between calls** (Bill, after the corpus
    numbers). `cwd` is a per-call argument on `run_command`; the default is
    the launch-time workspace; no `set_cwd` — one-shot settings never need to
    survive compaction, sticky ones always do. A persistent shell is a job the
    model opens explicitly. No P6 declined decision in this chapter.

## Open

1. **Coder brief:** add `cwd` to the reference's `run_command` (per call,
   relative to workspace, error on missing dir, recorded on the job and in the
   report) and the two `jobmodel` legs; mirror in `agent/`, re-snapshot
   `solutions/ch04`, retag; `-ch 3 solutions/ch04` must stay 100. Mutant:
   `cwd-ignored` must fail `jobmodel`, and the leg's assertion must be on a
   line of output, not a substring.
2. **Title.** "Containment, Not Cancellation" now names a sentence that
   survives in one place (`kill_job` on a goroutine). The chapter's spine is
   "stop discarding the result" / "the wait belongs to the call." Keep the
   title for the war story it honors, or retitle? Bill.
3. **The PTY incident** from Bill's memory ("a lot of trouble") — no longer a
   dependency; the chapter has its own receipt (`dlv` on pipes). Search
   deferred indefinitely unless it turns up on its own.
