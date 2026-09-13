# Chapter 3 — seed notes (tools)

## 0. STRUCTURAL RULING 2026-09-13: ch3 splits into ch3 + ch4

**Bill:** *"run_command in its blocking form is trivial. Why not let the user
build the easy form of all 5 tools? They have an AI coding agent writing the
code. I feel this should be pretty fast, and the students will quickly learn why
blocking run_command is a bad idea. In fact, I did exactly this when building
StackAgent. Then ch 4 can be about doing tools right."*

| chapter | content | ends with |
|---|---|---|
| **ch3 — Tools** | the tool loop (parse `tool_use`, dispatch, return `tool_result`) and the **easy, blocking** implementation of all eight tools | a genuinely usable AI coding agent: reads, greps, edits, writes, runs the tests |
| **ch4 — Doing tools right** | the job model: handles, `wait_for_job`, `send_input`, `kill_job`, output to disk with stubs, the declined decision, `dlv` | an agent that can supervise work that hangs |
| **ch5 — Actors and hints** | *(was ch4)* mailbox, mid-turn steering | an agent that can hear you while it works |

Each chapter removes exactly one impossibility. Ch4 ends with the agent
supervising a long job; ch5 opens by observing it went deaf while doing so.

### Why this is better than pre-sorting the tools for them

The student implements all eight the obvious way and **hits the wall
themselves** — a command that hangs takes the whole agent down. That is a lesson
earned rather than announced. Bill lived it: StackAgent, July 2025, and ch1 §1.0
already tells the reader those 58,000 lines were thrown away. Ch4 can open with
*I built the blocking version first too.*

### Write-once: CHECKED, and it holds

`coderhapsody`'s `cr/docs/tool-calls-as-jobs-design.md` (Tier 1) settles it:

> Allocate a job handle at the single dispatch site [...] one funnel, already the
> single security enforcement point, so nothing bypasses it. [...] **No tool
> handler changes at all.** Tools stay `func(ToolUse) (string, error)`.

**The job model is a property of the dispatch site, not of the tools.** So of the
eight tools built in ch3, **seven need no change at all** for ch4's purposes.
Ch4 wraps the funnel and reworks exactly one — and reworking it is small, because
the blocking version was trivial to begin with.

**And which one is the point.** It is the tool that crosses a boundary you do not
control — the rule ch3 already taught. The chapter structure demonstrates the
thesis rather than asserting it: *exactly one of your eight was wrong, and you
could have predicted which.*

### P1: clarified, and it was never in danger

**Bill, 2026-09-13:** *"It is OK to rewrite small bits. We're not blowing up all
their work. run_command is trivial in ch 3."* And: *"We will wind up editing
prior code many times in this book. We're building on work, not blowing it up."*

P1's "additive" is about the **architecture**, not the files. Editing prior code
is ordinary engineering and this book will do it repeatedly. What ch1 did, and no
later chapter may do, is discard the *approach*. Ch4 turning fifteen trivial
lines of blocking `run_command` into supervised ones is building on the student's
work, not demolishing it. `course-policy.md` now says so explicitly.

### Forward promises: renumbered in `fb3af87`

All references were reclassified line by line, not by global substitution —
several "Chapter 4" mentions meant the *actors* chapter (now 5) and several
meant the *jobs* chapter (now 4). A blind `sed` would have manufactured false
promises. `chapter-04-actors-parking.md` was renamed to
`chapter-05-actors-parking.md`.

---

## 1. What chapter 3 has already been promised to do

**Status:** no outline yet. This file is the raw material from the 2026-09-13
session with Bill, written down before it rots.

After the ch3/ch4 split, chapter 3's inherited spec is **small**. Most of what
was promised to "chapter 3" was promised to the *jobs* material, which is now
chapter 4. What remains:

| # | promise | source | status |
|---|---|---|---|
| 1 | The response `type` filter becomes load-bearing. Chapter 3's wire **must** return a reply containing a non-text block. | `chapter-01-outline.md` §1.2, §1.8 | **discharged automatically** — a `tool_use` block *is* the non-text block |
| 2 | Chapter 3 is where a **deliberately-declined design decision** first appears (P6 anti-cheat). | `course-policy.md` P6 | **OPEN — the only real gap** |

Promise 1 is the happy case: the tool loop cannot be built without it, so no
special effort is needed to make the filter load-bearing. Chapter 1's forward
promise gets collected by the chapter simply doing its job.

Promise 2 is the gap, and the split created it. The declined decision I had
chosen — *what do you do when a supervised job never returns?* — belongs to the
job model, which is now chapter 4. Chapter 3 needs one of its own.

### Promises that moved to chapter 4

Recorded here so nobody looks for them in chapter 3 and concludes they were
dropped:

| promise | source |
|---|---|
| The event log gains **job events**. Note the noun: *job*, not *tool*. | `chapter-02-outline.md` |
| Chapter 4 must **time** a tool call; `ToolCalled` exists partly for this. | `chapter-02-outline.md` |
| Tool output arrives **by the megabyte** — where JSON-lines greppability pays. | `chapter-02-outline.md` |
| Chapter 4's tools are **not instant**, which is what makes chapter 5's mailbox earn its keep. | `chapter-05-actors-parking.md` |

The megabyte promise is a gift: chapter 2 built `RedactedPart` / `RedactSummary`
and nothing has yet *used* it. Megabyte tool output is its job — additive under
P1, and it collects on a debt.


---

## 2. The thesis — CORRECTED BY BILL, and the correction is the good part

**What I proposed (too broad, do not use):** "a tool call is a process you
start." Universal claim, all tools.

**Bill's correction, 2026-09-13, verbatim:**

> "There's not much reason to treat edit_file like a process, it works or fails
> quickly in every attempt the LLM ever made, and this is true of most of the
> locally executed tools that do not depend on anything out on the Internet.
> With MCP tools, we lost a lot of the control, and we have to assume any MCP
> tool we don't write ourselves could be flakey. They have to be handled like a
> process, just like run_command."

**The corrected thesis, which is a decision rule rather than a slogan:**

> The process model is earned by tools that **cross a boundary you do not
> control**. Local deterministic tools resolve or fail fast. `run_command` can
> hang. MCP tools are services someone else wrote, so you must assume they can
> wedge.

Why the narrow version is better:
- It is **true**, and Bill lived it; the broad version is retrospective tidying.
- It gives the reader a rule to apply on day one instead of a posture to adopt.
- It resists the over-engineering that "supervise everything" invites — which
  would have readers building job machinery around string replacement.
- It sets up chapter 5: MCP is where the process model stops being optional.

**Historical precision (Bill, 2026-09-13):** `run_command` has started *jobs*
since the StackAgent sprint in **July 2025** — `ai_callback_delay`, then check
status / wait / send input / kill, enough to drive `ed` interactively. Making
**all** tools manageable this way is new, "just in the last month or so."
Sub-agents were treated as processes rather than blocking calls from the start,
but were not introduced until **~December 2025 — UNVERIFIED, Bill would have to
check.** Do not print that date as-is.

**Write-once consequence.** Chapter 3 must ship both shapes: fast in-process
calls *and* supervised jobs. That is what CodeRhapsody actually is, and it keeps
chapter 4's mailbox purely additive, because the tools that need supervising are
exactly the slow ones.

---

## 3. The war story (CHAPTER 4 — it belongs to the job model) — RECEIPTED, two commits, both 2026-08-10

Repo `coderhapsody`, `internal/agent/watchdog.go`.

### 3a. `286bfd84` — "tool watchdog: a hung tool call must not be able to freeze the agent"

> Until now no tool call had any wall-clock bound. A handler that blocked
> forever took the whole agent with it: not crashed, not errored, just gone.
> The turn never completes, no observer fires, agents_status keeps reporting
> "processing", and a supervisor waiting on it waits forever.
>
> Hit live today by a `screenshot` call that never returned. The only recovery
> was killing CodeRhapsody -- which kills every agent in the tree, including
> sub-agents doing unrelated work. For one interactive session that costs an
> afternoon. For a fleet it is fatal, because the entire supervision model
> assumes turns end.

**The honesty beat, which is the best teaching in the whole commit:**

> Go cannot kill a goroutine. When the watchdog fires we STOP WAITING; the
> handler keeps running until it returns on its own or the process exits. This
> is containment, not cancellation, and the tool result says so in those words
> -- claiming otherwise would be exactly the "plausible default" anti-pattern we
> ban in compiler code.

Abandoned handlers are counted (`LeakedToolCalls`) so a human can decide to
restart. **What you buy is that the agent survives: a bounded, loud, recoverable
failure instead of an unbounded silent one.**

**MCP is explicitly the case inference cannot cover** — same point Bill made
tonight, written a month earlier:

> it cannot cover MCP tools at all, since those are third-party schemas we do
> not control -- and they are both the likeliest to wedge (a browser call on a
> page that never settles)

### 3b. `4bec4b96` — "watchdog: tools DECLARE their blocking contract; stop inferring it"

Same day. Two defects shipped in 3a: `send_secret` was missing from a
hand-written name→arg map and silently got the 3-minute default; and when the
deadline argument was omitted the budget fell back to 3 minutes even though the
tool's own default was longer. **Measured:** `join_agents` / `spawn_sub_agent`
(600s) and `send_message_to_parent` (300s) "would each have been abandoned
mid-wait while perfectly healthy."

**Bill's ruling:**

> conformance is the tool's job. The watchdog must not sniff schemas or parse
> English descriptions hunting for timeouts. Any tool that fails to declare its
> blocking behaviour and then trips the watchdog has a bug on its own side, and
> the loud abandonment message is the right way to learn that.

`ToolDefinition` gains `DeadlineArg` and `MaxBlockingTime`. Ten tools declare
contracts. The prose-parsing test is **deleted**; nothing at runtime reads a
description or scans a schema. This is "annotate the producer, don't detect from
the consumer."

**And the line that rhymes with this whole book's method:**

> A ruling written down and not wired up is invisible: the sentence reads like a
> reason and nothing tests it.

That is the same lesson as P9 and the chapter 1 grader audit, arrived at
independently a month earlier in a different subsystem. Worth saying out loud in
the chapter.

---

## 4. The declined decision (CHAPTER 4 — moved with the job model) — CANDIDATE

**Rejected candidate:** "run tool calls in parallel or sequentially?" Bill runs
them **sequentially on purpose**. It is not an open question in this book and
pretending otherwise would contradict the author.

**Proposed candidate: what do you do when a supervised job never returns?**

It is genuinely unsettled, by Bill's own account (2026-09-13): budgets and tool
call limits are "future work"; his instinct is to **pause** the agent that hit a
limit rather than kill it, let anything waiting on it wake and continue, and let
the parent resume it with more budget.

The knob has **two opposite failure modes and no correct setting**, and we have
receipts for both:
- too short → healthy work abandoned (`join_agents` at 600s, measured)
- too long → the freeze of 3a (a `screenshot` that never returned)

So the chapter can honestly say: we do not know the right answer, here are three
(kill / abandon-and-count / pause-and-ask-the-parent), pick one and make your
event log show which you chose. Nothing to paste, because the text declines to
answer, and the decision is real rather than a riddle.

---

## 5. The tool set — RULED 2026-09-13

**Bill's constraint, and it overrides tidiness:**

> "I feel like we should not move on from tools with a set that isn't usable for
> an AI coding agent to actually code."

This killed a proposed set of `read_file` + `fetch`. The reason it had to die is
chapter 4: ch4's whole payoff is steering an agent *mid-work*. Steering an agent
that can only fetch a URL is a demo. Steering one that is editing code and
running tests is the book. **A tool set that cannot code makes chapter 4 a toy.**

**The set: three tools, deliberately asymmetric.** **SUPERSEDED — see §7**,
which is authoritative for what the student implements. The reasoning below
(class asymmetry, portability, why `fetch` was cut) still holds; only the list
changed, after the §6 audit and Bill's "leave them a usable agent" ruling.

| tool | class | role in the chapter |
|---|---|---|
| `read_file` | fast, local | counterexample — building job machinery around this is waste |
| `write_file` | fast, local, **mutating** | completes the coding loop; the mutation matters in ch7 |
| `run_command` | **supervised job** | the chapter's spine, and Bill's own origin story |

Read the code, change the code, run the tests. That is the loop the book is
about, and roughly what every real coding agent ships as its core.

### Portability: the objection I raised and then dissolved

I initially cut `run_command` as ungradeable because `sleep` is absent on
Windows and shell quoting differs. That was treating a solvable problem as a
constraint — the classic move that makes a chapter teach the wrong thing for the
grader's convenience.

**Go is required for this course, so `go` is the one binary every student
provably has.** The grader scripts `go` invocations and gets determinism free.
For the awkward scenarios the exercise ships a tiny Go helper the agent runs:

- fast — returns immediately
- slow — dribbles output over several seconds (feeds ch4's mailbox)
- flood — emits a **megabyte** (discharges promise 4, exercises ch2's
  `RedactedPart`, forces output to disk with a stub in context)
- hang — never returns (forces the declined decision of §4)

All portable, all deterministic, no shell-quoting hazard.

### `fetch` is CUT, and cutting it improves chapter 7

`run_command` subsumes every scenario `fetch` was carrying. The original
argument for keeping it was that ch7 needs a bare GET as a command-and-control
channel. But the reveal is sharper without it:

> You never added a network tool. You added a shell, and a shell contains every
> tool.

The agent has had `curl` since the moment `run_command` existed. That is the
honest lesson and it lands harder than a tool planted for the purpose. It also
matches what is already written in `coderhapsody`'s `cr/docs/sandbox-design.md`:

> An agent's required boundary is a *function of the tools it holds* [...]
> `run_command` present → **Process or stronger**

The tool set *is* the threat model. Chapter 3 ships the tool that forces the
boundary; chapter 7 collects.

### What the student builds beyond the tools

The tool loop chapter 2 explicitly deferred: parse a `tool_use` block (where
ch1's deferred `type` filter finally bites), dispatch it, return a `tool_result`
in the correct vendor shape, loop until the model stops asking. Plus job events
in the log, a timing on every call, and one deliberate choice about the job that
never returns.

---

## 6. TOOL USAGE AUDIT — measured, 2026-09-13

Bill's framing: chapter 2 is complex enough that a student realistically drives
an AI coding agent to complete it. So the set chapter 3 teaches should be the set
**a real coding agent actually leans on**, not a set chosen for tidiness.

That is an empirical question, and the answer was sitting in the logs.

**Corpus:** 511 archived session histories from
`~/projects/coderhapsody/cr/histories` — CodeRhapsody working on its own
codebase, after Bill scrubbed Hewitt/Lyric-generated sessions on 2026-09-13.
**71,032 tool calls.**

```
grep -h '^### TOOL_CALL: ' *.md | sed 's/^### TOOL_CALL: //' \
  | sort | uniq -c | sort -rn
```

**THE CANONICAL TABLE IS `book/exhibit-ch03-tools.md`**, generated by
`scripts/tool-usage.sh`. It lists all 55 coding-relevant tools with family,
share and cumulative share. Do not retype it here; regenerate it.

**Two denominators, do not mix them:**
- **71,032** — every tool call in the corpus. Used for the robustness
  comparison below, because that comparison is about the corpus, not the set.
- **70,401** — coding-relevant calls only, after excluding social,
  presentations, mail, calendar, drive, browser automation and parse artifacts.
  **This is the denominator the book prints**, and the excluded share is **0.9%**
  — which is itself the finding: this agent spends essentially all its time
  coding.

Against the printed denominator: `run_command` 36.34%, `read_file` 25.70%,
`edit_file` 16.58%, `search_files` 10.29%, `write_file` 2.30%. **Top five
91.21%.** Four families cover **96.72%**: shell + jobs 38.37%, read + navigate
27.21%, mutate files 19.59%, search 11.55%.

### The distribution is ROBUST, and that is the stronger claim

The first run of this audit used a mixed corpus (514 files, 73,777 calls) that
still contained Hewitt/Lyric compiler sessions. Scrubbing them removed **2,745
calls, 3.7% of the corpus** — and moved no share by more than **0.3 points**:

| tool | mixed corpus | scrubbed | delta |
|---|---|---|---|
| `run_command` | 36.3% | 36.0% | −0.3 |
| `read_file` | 25.3% | 25.5% | +0.2 |
| `edit_file` | 16.4% | 16.4% | 0.0 |
| `search_files` | 10.0% | 10.2% | +0.2 |
| `write_file` | 2.3% | 2.3% | 0.0 |
| **top five** | 90.2% | **90.4%** | +0.2 |

So the shape is not an artifact of which sessions were included. A 3.7%
perturbation of the population leaves the ranking and the 90% concentration
intact. **Cite the robustness, not just the table** — a single distribution is a
number, but a distribution that survives having a chunk cut out of it is a
finding.

Roughly 110 further tools share the remaining ~10%: `refine_context`,
`send_input`, `find_files`, `list_directory`, `semantic_search`,
`replace_lines`, `wait_for_job`, `search_web`, and a long tail of browser,
memory, skill and sub-agent tools mostly in single or double digits. (Exact
counts for the tail were taken against the pre-scrub corpus and are not restated
here; re-run the one-liner if a specific figure is ever printed.)

### Three findings, in order of how much they change the chapter

**1. `search_files` was missing from my proposed set, and it is 10% of all
calls.** An agent without grep is blind: it cannot locate the thing to read
before reading it. The three-tool set I proposed would have shipped an agent that
cannot find its own work. **Caught only by measuring.**

**2. `edit_file` outnumbers `write_file` 7:1** (11,671 vs 1,616). Given both, a
coding agent overwhelmingly makes targeted edits instead of rewriting files.
Worth stating in the chapter, because the naive instinct is to ship `write_file`
alone as "simpler" — and the result is an agent that rewrites a 400-line file to
change one line, burning output tokens and clobbering concurrent edits.

**3. The supervision verbs are routine, not exotic — this is §2's thesis,
measured.** `send_input` 800 + `wait_for_job` 322 + `kill_job` 97 + `jobs` 80 =
**1,299 calls, comparable to `write_file`'s 1,616.** Interacting with a job
*while it runs* is a first-class activity in real usage, not a defensive corner
case. This is Bill's July 2025 `ed` story appearing in the aggregate, and it is
the strongest single argument that chapter 3 must ship the process model rather
than a blocking call.

### Revised tool set (supersedes §5's table)

| tool | class | share | why it is in |
|---|---|---|---|
| `run_command` | **supervised job** | 36.0% | the chapter's spine; carries slow/flood/hang |
| `read_file` | fast, local | 25.5% | the counterexample — do not wrap this in job machinery |
| `edit_file` | fast, local, mutating | 16.4% | how agents actually change code |
| `search_files` | fast, local | 10.2% | without it the agent cannot find anything |
| `write_file` | fast, local, mutating | 2.3% | file *creation*; trivial; keep or cut |

Four are non-negotiable. `write_file` is the only judgement call: 2.3% of calls,
but it is the sole way to create a new file, and it costs ~10 lines.

**This distribution is itself publishable material.** A measured answer to "what
is actually in an AI coding agent" — five tools, 90% of calls — is exactly the
kind of receipt §1.1's bare-metal argument trades on, and no framework
documentation will tell a reader this.

---

## 7. THE IMPLEMENTED SET — what the student ships by end of ch3

**Bill's requirement, 2026-09-13:** *"I'd like to leave the students with a
usable AI coding agent by the end of ch 3."* And: *"I think the exercise should
be to implement them all."* The 51-row exhibit
(`book/exhibit-ch03-tools.md`) is for the READER to study. This is what the
student BUILDS, and all of it is graded.

| # | tool | class | share | why it is in |
|---|---|---|---|---|
| 1 | `run_command` | **supervised job** | 36.3% | the crown jewel; starts a job, returns a handle |
| 2 | `send_input` | **job lifecycle** | 1.13% | write to a *running* job's stdin |
| 3 | `wait_for_job` | **job lifecycle** | 0.45% | block on a job until output or completion |
| 4 | `read_file` | fast, local | 25.7% | line ranges and limits; `cat` floods context |
| 5 | `edit_file` | fast, local, mutating | 16.6% | how agents actually change code (7:1 over write) |
| 6 | `write_file` | fast, local, mutating | 2.3% | file creation |
| 7 | `list_directory` | fast, local | 0.75% | orientation |
| 8 | `search_files` | fast, local | 10.2% | **RULED IN by Bill** — an agent that cannot grep cannot find what to read |

**Coverage: these eight tools are 93.6% of every coding call in the corpus**
(65,870 of 70,401). With `kill_job` it is 93.9%. That is the line worth printing:
**a student implements eight tools and gets ninety-four percent of what a real
AI coding agent actually does.** The remaining 6% is a long tail of memory,
skills, sub-agents and context management — every one of which is a later
chapter.

**One still needing Bill's word:**
- **`kill_job`** (225 calls merged) — not requested, but it *closes the
  lifecycle*: start → observe → interact → terminate. Without it, one of the
  three permitted answers to §4's declined decision ("kill it") cannot actually
  be implemented by the student. Recommend including it for that reason alone.

### THE PAYOFF: the student's agent drives `dlv`

Bill's framing: *"the crown jewel of course is run_command and to round it out,
we should add send_input, and wait_for_job as well so the user can watch their
AI coding agent use dlv."*

This is the chapter's closing demonstration, and it is chosen because **it
proves the process model is necessary rather than stylistic.**

A blocking tool call returns when the process exits. **A debugger never exits
until you tell it to quit.** So with blocking tool calls, driving `dlv` is not
slow and not awkward — it is *impossible*. There is no version of "just wait for
the command to finish" that produces a breakpoint. The chapter's thesis stops
being a matter of taste and becomes a capability boundary the student can stand
on either side of.

The sequence, which is how a real agent does it:

```
run_command  "PAGER=cat dlv debug main.go"   ai_callback_pattern="(dlv) "
send_input   "b main.go:42"                  ai_callback_pattern="(dlv) "
send_input   "c"                             ai_callback_pattern="(dlv) "
send_input   "p myVar"                       ai_callback_pattern="(dlv) "
send_input   "q"
```

Note `ai_callback_pattern`. The agent does not sleep a guessed number of
seconds; it **waits for the prompt string to appear**. That is worth teaching
explicitly: a fixed delay is a race condition with a comfortable name. Waiting
on a *signal in the output* is the difference between supervising a process and
hoping about one.

`dlv` is also portable enough to grade: it is `go install`-able, and Go is
already required.

**Why this ends the chapter.** The student began chapter 1 with a program that
could hold a conversation. They end chapter 3 watching their own agent set a
breakpoint, continue to it, and print a variable. Nothing about that reads as a
toy, and it is reachable in one chapter only because the tool set was chosen
from measurement rather than taste.

### "Do we need anything other than `run_command`?" — Bill's question, and the section it deserves

Bill raised it himself: *"I'm not entirely sure we need anything other than
run_command, though. It is more a matter that you actually use the other tools
than how critical they are."*

**He is right that it is sufficient.** `run_command` is Turing-complete. `cat`,
`sed`, `ls` and `grep` cover every other tool on the list. That is exactly why
"is it sufficient?" is the wrong test. A tool set is not a capability list; it
is a set of affordances and constraints. Five reasons the dedicated tools earn
their place, four of them receipted from the session that produced this file:

1. **Loud failure.** `edit_file` refused an edit three times in one session
   because a markdown heading was not repeated, and once because a word had
   wrapped mid-anchor. `sed` would have silently done the wrong thing. A tool
   that fails loudly beats one that succeeds ambiguously.
2. **Portability.** Same session: `sed -i ''` (BSD) vs `sed -i` (GNU), and
   `cat -A` unsupported on macOS. Every shell-based file operation carries that
   tax. `edit_file` does not.
3. **Context volume.** `read_file` has line ranges, per-line truncation and a
   text limit. `cat` floods the context window, and you pay for that flood on
   every subsequent turn.
4. **Quoting.** Content containing quotes, backticks or newlines is hazardous
   through a shell — backticks had to be escaped twice in this session to stop
   the shell executing a table that was being generated.
5. **You cannot withhold a capability you have bundled into a shell.** This is
   the one that matters, and it is chapter 7's spine. A read-only agent is
   expressible as `read_file` + `list_directory` + `search_files`. It is *not*
   expressible if reading is `run_command cat`. `coderhapsody`'s
   `cr/docs/workflow-design.md` already specifies a `readonly-agent` skill
   defined exactly this way, and `sandbox-design.md` states the rule directly:
   the required boundary is a function of the tools held.

**The framing for the chapter:**

> `run_command` is the tool that makes the agent capable. The others are what
> make it steerable, auditable, and containable.

And the empirical kicker: `edit_file` outnumbers `write_file` **7:1**. The model
did not have to prefer targeted edits — it preferred them *because the tool
existed*. **Providing a tool changes behaviour, not just capability.** That is
the answer to "how critical is it" — criticality is the wrong axis.

---

## 8. Open for Bill


**RULED, recorded here so they are not re-litigated:**

- *The corrected thesis* — the process model is earned by tools that cross a
  boundary you do not control. Confirmed by Bill: blocking `run_command` was
  trivial to build first, and he did exactly that during the StackAgent sprint.
- *Chapter 4's cold open* — the `screenshot` hang, with the MCP hang as the
  generalization. Specific failure first, then the class of failure.
- *Chapter 3's declined decision* — `edit_file`'s failure contract.
- *The tool set* — the eight above, plus a ruling still owed on `kill_job`.

**STILL OPEN:**

1. `kill_job` — in or out? Without it the student cannot implement the "kill it"
   branch of chapter 4's declined decision: they would be offered three answers
   and able to build two.
2. Verify the sub-agent introduction date (~Dec 2025?). Chapter 8's territory,
   unverified, not to be printed until checked.
3. The PTY trouble with `run_command` — Bill recalls "a lot of trouble" and that
   it is probably a real incident, but details need a history-log search. Not yet
   done. Memory fragment to check: PTY per **job** not per **session**, because
   shared mutable state destroys replay. Chapter 4 material if it pans out, and
   chapter 4 already has a receipted war story, so this is a bonus rather than a
   dependency.
