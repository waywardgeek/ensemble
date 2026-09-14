# Building Advanced AI Coding Agents

**Build the coding agent you've been renting.** One chapter at a time, from a
chatbot that forgets your name to an agent that drives a debugger to a
breakpoint and reads a variable off the stack — and every chapter's code is
here, runnable, graded, and free to try.

This repo is the whole course: the book *and* the machinery. The text is in
`book/` — preface, one outline per chapter (the check tables and design
rulings live there), and the working notes we wrote to ourselves along the
way. The machinery is the part you can *run*: a reference agent that grows
chapter by chapter, an auto-grader for each chapter's exercise, a fake LLM
endpoint so all of it costs nothing until you decide to plug in a real key,
and one script that talks to real models when you do.

**The book will always be free here.** It will also be on Amazon for people
who'd rather read on a Kindle or hold a paper copy — same text, your choice.

Four chapters are built. The agent already does things most people assume
need a framework, a vendor SDK, and a team. It needs none of those. It needs
an event log, a tool loop, and the idea that a tool call is a *process*, not
a function.

> **We're turning this into a video course and looking for collaborators.**
> If you found this through Bill's LinkedIn post: welcome. Skip to
> [Kick the tires](#kick-the-tires-in-five-minutes), get it talking, then read
> [Help build the course](#help-build-the-course). You do **not** need to be
> the engineer described in the next section to help make this.

---

## Who this is for

This is not an ordinary software-engineering course, and the exercises are
not ordinary exercises. The reference agent stands at 5,000 lines of Go by the
end of chapter 4 — and chapter 1's code is thrown away on purpose, so that is
5,000 lines across three chapters. Chapter 2 alone is over two thousand. A
student is expected to produce each chapter — production-worthy, graded, all
checks passing — in a few days at most.

No human writes that by hand in that time. The course assumes you already
work the way the next generation of engineers works: fluent with Cursor,
Claude Code, Codex or their equivalent, directing an agent rather than
typing, and *reading* at the speed it produces. If you have never shipped a
few thousand lines of agent-written code that you could defend line by line,
start there and come back. If you have, this course is the next thing.

**Or just read along.** The book, the outlines, the reference solutions and
the reviews that shaped them are all here and all free. You can follow every
decision without building anything, and you'll come away knowing exactly how
the agent you use every day works — and where it cuts corners. The exercises
are for people who want to own one.

The receipt is the git log. The idea for this course was born on a Friday
afternoon. Four chapters of reference code, their graders — each audited by
deleting behaviour from the reference and proving the score drops — the fake
vendors, the live harness, and the outlines you'll read were all committed by
Monday morning, by one engineer working with his agents. That pace is not the
point. The point is that it is now the *normal* pace for people who work this
way, and a course for them has to be pitched at it.

---

## What it does by chapter 4

Type at it and it will list your directory, read your files, grep your code,
edit a file by anchor text (and *refuse* if the anchor matches twice — it will
not silently patch the wrong one), run your tests, and tell you what happened.
Ordinary. Here is what is not:

- **It never blocks.** Every tool call is a *job* with a handle. Start
  `go test ./...`, come back to it, read the output as it arrives. Nothing is
  killed by a timeout — nothing dies unless the model calls `kill_job`. The
  model decides how long to wait *on every call*, not the framework.
- **It drives a debugger.** Real `dlv`, in a real PTY, waiting on the `(dlv) `
  prompt rather than on a guessed delay. Set a breakpoint, continue, print a
  variable. Blocking tool calls don't make debuggers slow; they make them
  impossible. This is the chapter 4 closing demo and it runs live.
- **The log already knows how to be steered.** A hint typed mid-turn and an
  interrupt are both *events*, and the chapter 2 grader proves the agent
  handles them while blocked on a request it has already sent: the hint lands
  in the right place exactly once; the interrupted tool call is recorded and
  *not executed*, and the next request is still legal. Wiring that into the
  live REPL comes in a later chapter — the data structure is done.
- **One log, three vendors.** The conversation is an append-only event log.
  The same log renders to the Anthropic, OpenAI and Gemini APIs — switch
  vendors with one environment variable and the agent doesn't know.
- **Replay is byte-identical.** Render the same log twice, get the same
  request. Redact a payload: it's gone from what the model sees and still in
  the log. Compaction is an event, not a mutation.
- **Output by the megabyte, context by the kilobyte.** Full tool output goes
  to a file, always; what enters the context is capped, with the path to the
  rest. The model reads the rest with the `read_file` it already has.
- **The dangerous call is the one that makes you be specific.** `edit_file`
  refuses ambiguous anchors. `write_file` refuses to overwrite unless asked by
  name. Loud failures over quiet ones, everywhere.

Six local tools cover 92% of what a real coding agent does all day — we
measured it across tens of thousands of real tool calls before choosing them.
That measurement is in the book; the six tools are in `agent/`.

---

## Kick the tires in five minutes

You need Go. For the debugger demo you also need `dlv`:
`go install github.com/go-delve/delve/cmd/dlv@latest`.

### Free: watch the protocol turn

No key, no network, no cost. The fake vendor serves scripted replies, so what
you're watching is the *agent*, not the model: every request, which vendor
dialect it hit, which tools it declared, which results it carried back.

```bash
go run ./cmd/fakevendor -ch 3 chat                  # chapter 3 tool loop, Anthropic dialect
go run ./cmd/fakevendor -ch 3 -vendor gemini chat   # same loop, Gemini dialect
go run ./cmd/fakevendor -ch 2 -vendor openai chat   # chapter 2, OpenAI dialect
```

Request 1 gets a tool call; request 2 carries the result. Same agent, three
wire formats.

### Live: talk to the real thing

This costs tokens — cents, not dollars, for a session of poking around.

```bash
export ANTHROPIC_API_KEY=sk-ant-...          # or OPENAI_API_KEY / GEMINI_API_KEY
scripts/live.sh 4 anthropic models           # which model IDs your key can actually use
go run ./agent chat                          # the live agent, through chapter 4
```

`agent/` is the live tree — everything built so far. In the REPL, try:

- *"List this directory and tell me what the repo is."*
- *"Run `go test ./...` and summarize the failures."* — watch it start a job,
  wait, and read the output in pieces.
- *"Edit README.md and change X to Y"* where X appears twice. Watch it refuse,
  then get specific.
- *"Overwrite notes.md with a to-do list"* on a file that exists. Same refusal,
  other tool.

### The debugger demo

```bash
scripts/live.sh 4 anthropic                  # or: scripts/live.sh 4 gemini
```

The script seeds a tiny Go program with a variable equal to 42, asks the agent
to debug it, and the agent starts `dlv` under `run_command` with a callback
pattern of `(dlv) `, sets a breakpoint with `send_input`, continues, and
prints the variable. The only right answer is 42. You'll see every step on
stderr.

### Pick your vendor

Chapters 2 and up read `LLM_VENDOR`, `LLM_API_KEY`, `LLM_BASE_URL`,
`LLM_MODEL` first, then the vendor-prefixed names (`ANTHROPIC_*`, `OPENAI_*`,
`GEMINI_*`). `scripts/live.sh` recognises its arguments by shape (chapter
number, vendor name, `chat` / `rounds` / `models`) and never echoes your key.
If the default model isn't one your key can see, `models` tells you what is.

---

## Grade something

Every chapter is an exercise with a contract, and every contract has an
auto-grader. Grade the reference solution — no key, no network:

```bash
make grade                          # chapter 1
make grade2 grade3                  # chapters 2 and 3
go run ./cmd/grade -ch 4 ./solutions/ch04
```

Grade your own agent — any package directory or built binary:

```bash
go run ./cmd/grade -ch 3 ~/my-agent
go run ./cmd/grade -ch 4 -json ~/my-agent   # machine-readable
```

Exit status is 0 on a pass, 1 on a fail, 2 if the grader itself couldn't run.
The report tells you which *property* failed and why, not just a score.

**The graders are tested by deletion.** For every behaviour a grader
protects, we delete that behaviour from the reference solution and confirm
the score drops — and assert *exactly which* checks fail. A grader that fails
everything on any defect is as useless as one that fails nothing. We've caught
our own graders scoring a broken agent 100 this way, more than once. That
story is in the book too.

```bash
make test                           # the grader's own sensitivity suite
```

---

## The arc so far

| ch | title | what you build | the idea that pays for it |
|---|---|---|---|
| 1 | The chatbot | A stdio program that calls the Messages API and keeps a conversation | The API is stateless. The conversation lives in *your* process — and the grader proves it by planting a string only the server ever typed. |
| 2 | One log, three vendors | An append-only event log, a reducer, and a renderer per vendor | History ≠ context. Hints, interrupts, redaction and replay all fall out of one data structure. |
| 3 | Six tools: 92% | Tool declaration, the tool loop, and six local tools | A shell is *sufficient*; it's the wrong test. Providing a tool changes what the model *does*, not just what it can do. |
| 4 | Every tool call is a job | Handles, PTYs, `wait_for_job`, `send_input`, `kill_job`, output capping | The wait belongs to the *call*, not the tool. Nothing to declare, no table to miss a row. |

Each chapter is strictly additive to the last. `solutions/chNN` is a frozen
snapshot of `agent/` at the moment the chapter finished, tagged
`chNN-solution`, so you can diff any two chapters and see exactly what a
chapter costs.

---

## What comes next

The agent can act. What it can't do yet is be *listened to* by anything but a
terminal — and that is where it gets interesting.

- **The agent framework.** Actors with mailboxes, observers instead of
  callbacks, and the seam that lets everything after this chapter plug in
  without touching the loop. The exercise is a *non-coding* agent, to prove
  the framework isn't secretly a coding-agent framework.
- **A GUI** — built on the event stream, not on the agent. If the agent needs
  the GUI to run, the architecture is wrong; the compiler will say so.
- **Real-time steering.** Type while the agent is working and your words reach
  the model as a hint on the next tool result — it changes course without
  breaking stride. The log has carried hints since chapter 2; this is where a
  human gets to send them. It's the way the authors have worked with their
  own agent since 2025, and it's the reason this course exists.
- **A gateway** — the same agent on Discord, on chat, wherever people already
  are.
- **Skills**: instructions as a first-class capability, loaded and unloaded
  at runtime. And the security chapter that follows directly from them —
  because a shell contains every tool, and you added one in chapter 3.
- **Sub-agents**: an agent is a sub-agent the moment you stop watching it.

By the end you have built, from scratch, understanding every line, the thing
you have been paying a subscription for.

---

## Help build the course

We're making a video, self-paced course from this material — chapter by
chapter, watch-then-build, graded by the tools in this repo.

**You do not need to be the engineer the course is for.** The systems
expertise is already in the repo: the code, the outlines, the design rulings,
the reviews between author and coder that explain every decision. What the
course does not have yet is someone who is great at teaching on camera —
someone who can take a chapter and make a live audience *see* it. If that's
you, you are the collaborator we're short of, whether or not you've written a
line of Go.

Here's what helps most, in order:

1. **Run it.** Kick the tires above. Where did you get stuck? Where did the
   README lie to you? Open an issue with the command and what happened.
2. **Do a chapter as a student** (if you code at AI speed). Read
   `book/chapter-0N-outline.md`, build the exercise, grade it. Tell us where
   the outline undersold or oversold the difficulty.
3. **Record yourself explaining a chapter.** Rough screen capture is fine. The
   first cut of a video course is finding out what a real person needs to see,
   and a teacher's instinct for that is worth more than another engineer's.
4. **Bring what you're good at**: presenting, video editing, course platforms,
   accessibility, curriculum design. None of us are experts in all of those.

Reach Bill through the LinkedIn post that brought you here, or open an issue.

---

## Layout

```
agent/                  THE LIVE AGENT — everything built so far; go run ./agent chat
book/                   THE BOOK — preface, chapter outlines with check tables and
                        rulings, course policy, and the reviews/briefs between
                        author and coder that shaped each chapter
cmd/grade/              the auto-grader CLI (-ch selects the chapter)
cmd/fakevendor/         the fake vendor as a local server, so you can RUN a
                        chapter for free instead of only being scored
internal/fakeanthropic/ chapter 1's stand-in for the Messages API
internal/fakevendor/    chapters 2+: deterministic stand-in for all three vendor
                        APIs, routed by request path; the grader mounts it
                        in-process, cmd/fakevendor serves it
internal/grade/         scripts, process harnesses, checks, report
solutions/ch01..ch04/   frozen reference solutions, one per finished chapter
scripts/live.sh         run a chapter's solution against a real vendor API
testdata/students/      deliberately defective submissions (grader self-test)
```

---

## Exercise contracts

Full check tables with points and rationale are in each chapter's outline in
`book/`. The two earliest are reproduced here because they're the ones a new
student meets first.

### Chapter 1

Ship a Go program that speaks JSON lines on stdio:

| direction | line |
|---|---|
| grader → program | `{"user": "..."}` |
| program → grader | `{"assistant": "..."}` |
| grader → program | *(closes stdin)* |
| program → grader | `{"usage": {"input": N, "output": M}}`, then exit 0 |

Environment supplied: `ANTHROPIC_BASE_URL` (POST to `$BASE/v1/messages`),
`ANTHROPIC_API_KEY` (the `x-api-key` header), `ANTHROPIC_MODEL`. **stdout
carries the protocol only**; diagnostics go to stderr.

| check | pts | property |
|---|---|---|
| `protocol` | 15 | one answer per round, clean exit, no junk on stdout |
| `wire` | 15 | well-formed Messages API calls |
| `calls` | 10 | exactly one API call per round |
| `replies` | 10 | answers are the text the server returned |
| `memory` | 25 | **the conversation exists** |
| `growth` | 15 | each request extends the previous one byte-for-byte |
| `usage` | 10 | cumulative token totals reported and correct |

**How `memory` works.** The fake plants a fixed string in its round-1
assistant reply and looks for it in round 4's request. Your program never
types that string — the server did — so it can only be present if you kept
the model's reply and resent the whole history. No single-shot program can
fake it.

### Chapter 2

Rebuild chapter 1 on an append-only event log and a derived context.
Observably identical for everything chapter 1 could do; then it accepts a
**hint** mid-turn and survives an **interrupt**.

| invocation | behavior |
|---|---|
| `./ch02` | grader mode — the chapter 1 contract, plus directives |
| `./ch02 chat` | the REPL |
| `./ch02 render LOG` | play LOG → context → render; print the request; **no network** |

Directives: `{"hint": "..."}`, `{"interrupt": true}`, `{"redact": <seq>}`,
`{"ephemera": {...}}`, `{"dump": "<path>"}` — each expects one `{"ok": true}`.

| check | pts | property |
|---|---|---|
| `ch1parity` | 25 | all seven chapter 1 checks still pass |
| `logdump` | 5 | log round-trips; `seq` monotonic, never reused |
| `replay` | 15 | two `render` runs on one log are byte-identical |
| `redaction` | 15 | payload gone from the request, still in the log |
| `ephemera` | 10 | carried in exactly one request, never a dialogue event |
| `hint` | 15 | received *while blocked*, classified, positioned, once, retained |
| `interrupt` | 10 | late tool call recorded and **not** executed |
| `usage` | 5 | cumulative totals from `ResponseEnded` events |

The grader holds its reply open and types at your program while it is blocked
on an HTTP request it has already sent. A round-synchronous program cannot
acknowledge anything in that window — which is how "received while blocked"
is measured rather than assumed.

### Chapters 3 and 4

See `book/chapter-03-outline.md` and `book/chapter-04-outline.md`, or run
`go run ./cmd/grade -ch 3 ./solutions/ch03` and read the report: every check
is named and explained.

---

## Grader design rules

1. **Record, then judge.** The fake captures every request verbatim; checks
   run afterwards against evidence.
2. **A malformed request still gets a 200.** Violations are recorded, not
   enforced at the transport. You learn every bug in one run, not one per run.
3. **Deterministic.** Scripted replies, token counts a pure function of the
   payload. No model, no network, no cost, identical everywhere.
4. **Audited by deletion.** Every protected behaviour is deleted from the
   reference and the score must drop, by exactly the checks that own it.

---

## License

Apache License 2.0 — see [LICENSE](LICENSE). The solutions are teaching code:
copy them, ship them, build on them.
