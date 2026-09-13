# Chapter 4 — Jobs: Containment, Not Cancellation

*Status: outline, first pass. War story is receipted — commits `286bfd84` and
`4bec4b96`, both 10 August 2026, `internal/agent/watchdog.go`. Architecture is
receipted — `cr/docs/tool-calls-as-jobs-design.md`. This document is the spec the
prose and the grader are both written from.*

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
tool gets a handle, an output file, a status, and a deadline. The three verbs it
introduces (`wait_for_job`, `send_input`, `kill_job`) are the supervision API for
all of them.

**Chapter 4 still does not add reach.** Those three verbs are together under two
percent of all tool calls, and your agent cannot touch one thing it could not
touch before. It earns its place twice over: it makes the biggest tool you
already built (`run_command`, thirty-six percent on its own) genuinely usable,
and it makes *every* tool incapable of taking the agent down with it.

The chapter's central distinction is which tools can report progress and which
cannot. A subprocess emits bytes over time, so you can watch it and match a
pattern against what it says. A `read_file` produces everything at the instant it
finishes or nothing at all. Both can be jobs. Only one can be *watched*, and a
design that pretends otherwise ships a parameter that silently does nothing.

And it ends by doing something your chapter 3 agent could not do at all, not
slowly but *at all*.

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

## §4.1 Why this is not a rare event

The instinct is to file that under bad luck. It is not. It is the predictable
consequence of a category error, and the category has a name.

A tool that runs locally and deterministically — read a file, edit a file, list
a directory — either returns or fails, fast, always. A tool that crosses a
boundary you do not control has no such property. Three examples, in ascending
order of how little control you have:

1. **A shell command.** You wrote the command, but not the program it runs.
2. **A network call.** You control neither the far end nor the path to it.
3. **An MCP tool.** Someone else's service, someone else's schema, someone
   else's uptime — and the tool most likely to wedge:

> it cannot cover MCP tools at all, since those are third-party schemas we do
> not control — and they are both the likeliest to wedge (a browser call on a
> page that never settles)
>
> — commit `286bfd84`

This is the chapter's operating rule and it should be stated once, plainly:
**the process model is earned by tools that cross a boundary you do not
control.** Do not supervise `edit_file`. It has never once failed to return.

*(Forward note, not to be stated in the prose: MCP is chapter 6. This section
plants the category without naming the chapter.)*

---

## §4.2 The fix that does not fix it

The mechanism is four lines and the student will be disappointed by how few:

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
`screenshot` call is still hung; it now has no one listening. Abandoned handlers
are counted, so a human can decide whether to restart the process.

**The chapter's title is this sentence**, and the discipline it demands is that
the tool result says so *in those words*. An error message claiming the call was
cancelled would be a lie that reads like a success. What you buy is a bounded,
loud, recoverable failure in place of an unbounded silent one. That is not the
same as fixing it, and the message must not pretend otherwise.

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

Three consequences the prose should draw out:

1. **The handle goes at the dispatch site**, not in the tools. One funnel, which
   is already where security is enforced, so nothing bypasses it. The design doc
   is blunt about the cost: *"No tool handler changes at all. Tools stay
   `func(ToolUse) (string, error)`."* Five of your six tools are not touched.

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

---

## §4.4 The three verbs

| tool | calls | share | what it is for |
|---|---|---|---|
| `send_input` | 800 | 1.14% | the process is waiting for you |
| `wait_for_job` | 322 | 0.46% | block until it is done |
| `kill_job` | 225 | 0.32% | stop it |

**`send_input` is the biggest of the three, by more than two to one**, and that
ordering is the finding. The common intuition is that job control is mostly
about *stopping* runaway work. In practice it is mostly about *talking to*
work that is going fine and is waiting for an answer.

That is the capability the chapter is really buying, and §4.8 spends it.

`wait_for_job` must work on a job that has *already finished*. A student who
implements it as "block on the channel" will hang forever on a completed job,
and will diagnose it as a deadlock in their own code rather than a missing case.

---

## §4.5 Output by the megabyte

Chapter 2 built `BlobPart.Path` and has not used it since. Here is where that
debt is collected.

`go test ./...` on a real repository produces more text than you want in a
context window, and you pay for those tokens on every subsequent turn of the
conversation, not just the one that ran the command. The rule:

- The full output goes to `cr/io/<handle>`, always, for every tool.
- What enters the context is a stub: the first N lines, the last N lines, the
  byte count, and the path.
- The path is not decoration. It is the recovery route — the model can read a
  range of it with the `read_file` it already has.

**Do not let the model choose the truncation.** The truncation happens on the
way in, at the dispatch site, before anything reaches the context window. A
model asked to summarize its own flood has already paid for the flood.

---

## §4.6 Tools declare their blocking contract

Twenty-two minutes after shipping the watchdog, I shipped a second commit fixing
it. Both are dated 10 August 2026 and the timestamps are 14:28 and 14:50.

The watchdog needed to know how long each tool may legitimately block. The first
version *inferred* it: a hand-maintained map of tool names, plus a scan of the
tool's description text hunting for documented timeouts. Two defects shipped:

> 1. `send_secret` declares `ai_callback_delay` and was MISSING from the
>    hand-written name→arg map, so it silently got the 3-minute default.
> 2. When the deadline argument was OMITTED, the budget fell back to the
>    3-minute default even though the tool's own default is longer. Measured:
>    `join_agents` / `spawn_sub_agent` (600s) and `send_message_to_parent` (300s)
>    would each have been abandoned mid-wait while perfectly healthy.

The second is the one that stings, and the commit says why:

> Defect 2 is the worse one because my own comment in the previous commit
> SPECIFIED the fix […] and the code did not do it. **A ruling written down and
> not wired up is invisible: the sentence reads like a reason and nothing tests
> it.**

**The ruling:** conformance is the tool's job. The watchdog must not sniff
schemas or parse English prose looking for timeouts. A tool declares two things
— the argument by which a caller sets its deadline, and how long it blocks when
that argument is absent. A tool that fails to declare and then trips the
watchdog has a bug on its own side, and the loud abandonment message is the
correct way to find out.

This is **annotate the producer, don't detect from the consumer**, and it
generalizes well beyond watchdogs. The prose-parsing test was deleted outright.
Nothing at runtime reads a description or scans a schema.

**Budget precedence**, in order:

1. an explicit one-shot override
2. the tool's own declared deadline argument, plus slack
3. a static entry, for the handful of tools that are slow by nature and take no
   deadline argument
4. a default — three minutes

**Why the override is one-shot, capped, and requires a written reason.** An
escape hatch that persists is an escape hatch that is always open: one decision
twenty turns ago silently disables the guard for the rest of the session. The
reason is required and enforced rather than documented, and it lands in the
transcript where a human can audit whether overrides are deliberate. Without
those three disciplines the override is `--no-verify` with a longer name.

**Why slack exists**, which is subtler and worth the paragraph: the supervisor
call `join_agents` has its own 600-second deadline and is *built* to return a
partial result on timeout. That is a success path. Without slack, the watchdog
and the tool's own deadline fire at the same moment and which one wins is a coin
flip — **nondeterminism manufactured by the safety net.**

---

## §4.7 The decision this chapter does not make

**A job outruns its budget. What happens?**

- **Kill it.** Reclaim the resources, report the failure, move on.
- **Abandon it and count it.** Stop waiting, leave it running, keep a tally a
  human can inspect. (This is what I shipped, and it is a consequence of "Go
  cannot kill a goroutine" rather than a free choice.)
- **Pause it and ask.** Suspend the job, wake anything waiting on it, and let
  the supervisor decide whether to grant more budget and resume.

I do not currently set budgets or call limits on sub-agents. It is future work.
My instinct is the third: pause rather than kill, let waiters wake and continue,
and let the parent top the budget up and resume. But I have not built it, so I
am not going to teach it as the answer.

The chapter declines to pick. The grader accepts any of the three, and checks
only that the choice is legible in the event log and that whatever is waiting on
the job is told what happened rather than left blocked.

*(P6: this is chapter 4's ungradeable-by-pasting decision, and it is genuinely
open — the author has an instinct and no implementation.)*

---

## §4.8 Your agent drives a debugger

The closing demonstration, and it is not a flourish. It is the proof.

```
run_command "PAGER=cat dlv debug main.go"   ai_callback_pattern="(dlv) "
send_input  "b main.go:42"
send_input  "c"
send_input  "p myVar"
send_input  "q"
```

**Why this is the right ending:** a blocking tool call returns when the process
exits. A debugger does not exit until you tell it to quit. With chapter 3's
tools, driving `dlv` is not slow and it is not awkward — it is **impossible**.
The chapter's thesis stops being a matter of taste and becomes a capability
boundary the reader can stand on either side of.

It also teaches `ai_callback_pattern` honestly. You do not sleep for a guessed
interval and hope the prompt has appeared. You wait for the string `(dlv) `,
because that is the actual signal that the debugger is ready for input. **A
fixed delay is a race condition with a comfortable name.**

The student's agent can now debug the code it wrote. That is the note to end on.

---

## §4.9 The exercise

**Contract:** `./ch04 <transcript-file>` — no network, deterministic, same fake
vendor as chapters 2 and 3.

**What the fixture must provide:**

- A command that **finishes fast** (`go version`).
- A command that **outlives its budget** (a Go helper that sleeps past the
  deadline). Portable, no shell tricks.
- A command that **never returns** (a Go helper that blocks forever) — this is
  the `screenshot` incident, reproducible.
- A command that **reads from stdin** and echoes a response, so `send_input` has
  something real to talk to.
- A command that emits **more than a megabyte** to stdout.
- A tool that **accepts a deadline argument but declares no contract** — the
  `send_secret` defect, as a fixture.

### Checks

| id | points | what it grades |
|---|---|---|
| `ch3parity` | 10 | chapter 3's six tools and the tool loop still work |
| `jobmodel` | 25 | handle allocated at the dispatch site for **every** tool; result retrievable after the call returns |
| `waitjob` | 10 | `wait_for_job` blocks until done **and works on an already-finished job** |
| `sendinput` | 15 | input reaches a running process; its response comes back |
| `killjob` | 10 | `kill_job` stops the process and anything waiting on it is told |
| `bigoutput` | 15 | full output on disk, stub in context with byte count and path, truncation happens at dispatch |
| `declare` | 10 | tools declare `DeadlineArg` / `MaxBlockingTime`; a tool that accepts a deadline argument and declares nothing is caught |
| `budgetchoice` | 5 | the declined decision: a choice was made, it is legible in the log, waiters are not left blocked |

**Sum: 100.** Verify by scoping awk to the checks table itself, not to the whole
file — this chapter has other tables whose second column is a number, and a
file-wide regex silently sums those too. (It happens to give the right answer
here only because the check ids contain no underscores and the tool names do.
That is luck, not a method.)

**Weighting rationale:** `jobmodel` is 25 because it is the chapter, and because
the common wrong answer — special-casing `run_command` instead of changing the
dispatch site — passes a naive test and fails the moment an MCP tool wedges.
`declare` is 10 rather than 5 because the inferring version is the one a student
will write first and it fails silently. `budgetchoice` is 5 because any coherent
answer passes; the points buy legibility.

### Grader audit (P9)

- Delete each protected behaviour from the reference and confirm the score
  drops. **A row reading 100 → 100 is the finding.**
- **Assert every mutation actually landed.** This project has now been bitten
  three times, twice inside the very commits that added the rule: once because
  `gofmt` realigned struct keys so a text match found nothing, once because a
  `perl` pattern missed. A mutation that does not apply scores 100 and
  manufactures a fake result.
- Watch for the guard-computed-by-the-thing-it-guards shape. From `4bec4b96`:
  a schema-scan test iterated over *what the scan found*, so shrinking the scan
  shrank the set under test. **A guard derived from its subject checks nothing.**
- A `t.Skip` on an empty scan is *"a vacuous pass with better manners."* Treat a
  skipped check as a failed check.
- `jobmodel` needs a negative control: a tool other than `run_command` must also
  get a handle, or a student who special-cased the shell scores full marks.

---

## What exists now for a later chapter

| thing | state after ch4 | collected in |
|---|---|---|
| `Interrupted` as a turn state | not built | Ch5 |
| the mailbox | not built | Ch5 — the agent is still deaf while a job runs |
| MCP tools | named as the worst case, not implemented | Ch6 |
| `set_tool_watchdog`-style overrides | one-shot, capped, reasoned | Ch7, as an audit surface |
| the shell | still contains every tool | Ch8 |

---

## Open

1. **The PTY incident.** Bill recalls "a lot of trouble" with the PTY for
   `run_command` and believes it is a real incident; details need a history-log
   search that has not been done. Memory fragment to check: PTY per **job** not
   per **session**, because shared mutable state destroys replay. This chapter
   already has a fully receipted war story, so this is a bonus rather than a
   dependency — and two war stories competing for one lesson is worse than one.
2. **Does §4.6 belong in this chapter or in the testing chapter?** The blocking
   contract is a tools lesson; the three mutation failures inside `4bec4b96` are
   a testing lesson, and they are excellent. Current plan: keep the ruling here,
   park the mutation stories for the testing chapter rather than spending them
   twice.
