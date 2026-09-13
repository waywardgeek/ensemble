# Chapter 3 — Eight Tools: Ninety-Four Percent of an AI Coding Agent

*Status: outline, first pass. Design decisions are settled (see
`chapter-03-seed.md` for the reasoning and receipts behind each). This document
is the spec the prose and the grader are both written from.*

---

## What this chapter is

At the end of chapter 2 the student has something that talks to three vendors
and remembers what it said. It cannot touch a file. It is a very well-engineered
conversation.

At the end of chapter 3 it writes code.

That is the whole arc, and it is the most satisfying chapter in the book to
finish. Chapter 2 was the hard one — a log format, a reducer, three renderers,
three parsers. This one is fast, and it should feel fast. The student implements
eight tools in their simplest honest form, and the thing comes alive.

**What it deliberately does not do:** supervise anything. Every tool here blocks
until it returns. That is the correct thing to build first, it is what I built
first, and it is enough to write real code with.

---

## §3.0 Cold open — I counted

I have a directory containing every session my agent has ever run. Five hundred
and eleven of them, months of work, each a full transcript of an agent editing
its own source code.

I had never counted what was in them.

```
grep -h '^### TOOL_CALL: ' *.md | sed 's/^### TOOL_CALL: //' | sort | uniq -c | sort -rn
```

Seventy thousand four hundred and one tool calls. The top five:

| tool | share |
|---|---|
| `run_command` | 36.3% |
| `read_file` | 25.7% |
| `edit_file` | 16.6% |
| `search_files` | 10.3% |
| `write_file` | 2.3% |

**Five tools. Ninety-one percent.**

The agent has fifty-one tools available. It earns its living with four verbs:
run things, read things, change things, find things.

This chapter builds eight tools. Together they are **93.6%** of every tool call
in that corpus. The remaining six and a half percent is memory, skills,
sub-agents and context management — and each of those is a later chapter, which
is a more useful way to read the tail than as leftovers.

**Voice note:** the cold open is a receipt, not a boast. State the method, state
the number, move on. No "you might be surprised to learn."

**Honesty note for the prose:** this corpus is my own agent's logs, so a reader
cannot reproduce my number from my data. Print the one-liner, which they can run
against their own. First-party measurement, labelled as such. Not reproducible
in data, fully reproducible in method.

**Caveat that must survive into print:** ten of the agent's current tools appear
nowhere in the corpus, including almost the whole sub-agent suite. They postdate
the measurement. So "sub-agents: 0.1%" is a date stamp, not a verdict.

---

## §3.1 The exhibit — the full table

The complete measured table lives in `exhibit-ch03-tools.md` and goes in the
chapter as a reader-facing exhibit: all fifty-one tools, with counts, share,
cumulative share and a status column.

Three things the reader should take from it, called out in the prose:

1. **The curve is brutally steep.** Top 5 is 90%. Top 13 is 97%. Twenty-eight
   tools were called exactly once.

2. **`edit_file` outnumbers `write_file` seven to one.** Given both, an agent
   overwhelmingly makes targeted edits rather than rewriting files. Ship
   `write_file` alone as "the simple option" and you get an agent that rewrites
   four hundred lines to change one, paying output tokens for the privilege.

3. **Two rows are dead or dying, and they are the most instructive rows in the
   table.** `compress_context` is retired. `handoff_task` is on its way out. A
   table showing only the tools that survived would hide the two best lessons in
   it. (The story of *why* they died is chapter 6's, not this one's — here, just
   let the status column raise the question.)

---

## §3.2 The tool loop

The mechanism, and it is smaller than the student expects.

A model reply is a list of typed blocks. Chapter 1 walked that list and
concatenated the text. Now a block can be a request to run something:

```
[ {type: "text",     text: "I'll check the tests."},
  {type: "tool_use", id: "tu_01", name: "run_command", input: {...}} ]
```

The loop:

1. Parse the reply into parts. **Dispatch on block type** — this is where
   chapter 1's deferred type filter finally bites. An implementation that only
   looks at text blocks does not see the tool call at all and silently does
   nothing.
2. Record the text. Record the tool call. Both are events; chapter 2's log
   already has `ToolCallPart`.
3. Execute each tool call, in order.
4. Send the results back as `tool_result` parts keyed by the call's `id`.
5. **Loop.** The model gets another turn. It may call more tools. Keep going
   until it replies without asking for anything.

Step 5 is the part that surprises people. A tool call is not the end of a turn,
it is the middle of one. The turn ends when the model stops asking.

**Things the chapter must say plainly, because each is a real bug students hit:**

- The `tool_result` must carry the `id` of the call it answers. Not the name,
  not the position — the id. A model that issued three calls needs to know which
  result is which.
- Anthropic requires the `tool_result` to come **first** in the content array of
  the message answering it. (Verified on the wire; see the ch2 verification
  record.)
- A tool that fails still returns a `tool_result`. Failure is a *result*, not an
  absence. This is the single most common way a student's agent locks up.
- Sequential execution is a deliberate choice here, not an oversight. I run tool
  calls one at a time on purpose. Say so, and say why: ordering is observable to
  the model, and a shell command that changes the working tree changes what the
  next tool sees.

---

## §3.3 The eight tools

| tool | share of corpus | shape |
|---|---|---|
| `run_command` | 36.3% | shell, blocking |
| `read_file` | 25.7% | local, fast |
| `edit_file` | 16.6% | local, fast, mutating |
| `search_files` | 10.3% | local, fast |
| `write_file` | 2.3% | local, fast, mutating |
| `list_directory` | 0.75% | local, fast |
| `send_input` | 1.1% | *(chapter 4)* |
| `wait_for_job` | 0.46% | *(chapter 4)* |

The first six ship blocking, in this chapter. `send_input` and `wait_for_job`
have no meaning until a tool can still be running when it returns to you, so
they are named here and built in chapter 4.

**Six tools, 92.0% of the corpus. Eight, 93.6%.**

Design notes the prose should carry:

- **`read_file` takes a line range and a size cap.** Not a nicety. `cat` on a
  four-thousand-line file floods the context window and the student pays for
  those tokens on every subsequent turn of the conversation. The tool that reads
  is also the tool that decides how much of the window to spend.
- **`search_files` earns its 10%** because an agent that cannot grep cannot find
  what to read. It is the tool that makes `read_file` usable on a codebase
  bigger than one directory.
- **`list_directory` is only 0.75%** and stays anyway, because orientation is
  cheap and an agent that cannot see the tree guesses at paths. This is the one
  tool in the set justified by judgement rather than by the measurement, and the
  chapter should admit that rather than pretend the number argues for it.

---

## §3.4 "Do we need anything other than `run_command`?"

Bill's question, and it deserves the section rather than a footnote.

**`run_command` is sufficient.** It is Turing-complete. `cat`, `sed`, `ls`,
`grep` cover every other tool in the set. Which is exactly why sufficiency is
the wrong test. A tool set is not a capability list, it is a set of affordances
and constraints.

Five reasons the dedicated tools earn their place. Four are receipted from a
single night's work on this book:

1. **Loud failure.** `edit_file` refused four of my edits in one session — three
   because I had not repeated a heading the tool guards, one because a word had
   wrapped and my anchor no longer matched the file. `sed` would have accepted
   all four and silently done the wrong thing. A tool that refuses beats a tool
   that succeeds ambiguously.

2. **Portability.** `sed -i` takes an argument on BSD and does not on GNU.
   `cat -A` does not exist on macOS. Every shell-based file edit carries that
   tax. `edit_file` does not.

3. **Context volume.** Line ranges and size caps, as above. The shell has no
   opinion about how much of your context window it spends.

4. **Quoting.** Writing content that contains quotes, backticks or newlines
   through a shell is genuinely hazardous. I escaped backticks twice in one
   night to stop a heredoc executing the table it was supposed to print.

5. **You cannot withhold a capability you have bundled into a shell.** A
   read-only agent is expressible as `read_file` + `list_directory` +
   `search_files`. It is not expressible if reading is `run_command cat`. The
   tool set *is* the permission boundary.

**The framing that ties it together:** `run_command` is what makes the agent
capable. The other tools are what make it steerable, auditable and containable.

**The kicker, and it is the chapter's best empirical point:** `edit_file`
outnumbers `write_file` seven to one. The model was never told to prefer
targeted edits. It preferred them *because the tool existed*. Providing a tool
changes behaviour, not just capability — which is the real answer to "how
critical is it." Criticality was the wrong axis.

*(Reason 5 is chapter 8's spine and should be planted here without being
explained. The student adds a shell in chapter 3 and discovers in chapter 8 that
they never added a network tool — they added a shell, and a shell contains every
tool.)*

---

## §3.5 The decision this chapter does not make

**When `edit_file`'s anchor does not match the file, what happens?**

Three defensible answers:

- **Refuse.** Report the mismatch, change nothing, let the model try again.
- **Fuzzy-match.** Find the closest region within some window and apply there.
- **Rewrite.** Fall back to replacing the whole file.

The chapter teaches everything needed to decide and then does not decide.

**Why this is a genuine open question and not a riddle with a hidden answer:**
my agent ships *both* answers. `edit_file` refuses on an exact-match failure.
`replace_lines` deliberately fuzzy-searches within fifty lines of the line
numbers you gave it. Same codebase, same author, opposite calls.

And the usage numbers are lopsided: `edit_file` 11,671 calls against
`replace_lines` 385. **Thirty to one.** My explanation — and it is checkable,
which is why I will rest the argument on it — is that text anchors compose
across edits and line numbers do not. Make one edit near the top of a file and
every line number below it is stale, so a second `replace_lines` needs a fresh
read first. Anchors survive edits elsewhere in the file, so several can be fired
at once. The tool is not worse. It is non-composable, and that shows up as
thirty to one.

**What the grader checks:** not which answer. That a choice was made, that the
event log makes it legible, and that a failed edit returns to the model in a
form it can act on. A refusal that does not say what it saw is only half a loud
failure — it declines to guess, and then costs a round trip to find out why.

*(P6: this is chapter 3's ungradeable-by-pasting decision. Paste the chapter
into an assistant and there is nothing in the text to copy out. The assistant
will pick one confidently, and it cannot know which one you picked, and the rest
of your implementation has to agree with it.)*

---

## §3.6 What you have now

Short section. The student's agent can read a codebase, find things in it,
change them, and run the tests. That is the loop this entire book is about, and
it now closes without a human in it.

End on the capability, not on a preview of chapter 4. (Ambush ruling: do not
telegraph that one of these tools is about to be reworked. The wall is more
instructive when the student hits it themselves — which they will, the first
time they ask their agent to run something that takes a while.)

---

## §3.7 The exercise

**Contract:** `./ch03 <transcript-file>` — no network, deterministic, fake
vendor served from `internal/fakevendor` as in chapter 2.

**What the fake must serve**, since these are the scenarios the checks need:

- A reply containing **text plus one `tool_use` block**. (Collects chapter 1's
  deferred type-filter promise: an implementation that walks only text blocks
  never sees the call.)
- A reply containing **text plus two `tool_use` blocks**, to force correct id
  handling and ordering.
- A **multi-turn** sequence: tool result → another tool call → final text reply.
  The loop must not stop after one round.
- A tool call whose **execution fails** (missing file), to force failure to come
  back as a `tool_result` rather than a crash.
- A tool call with **malformed arguments**, same reason.

**Commands the student's agent must be able to run** (the `go` toolchain is the
one binary every student is guaranteed to have, since the course requires it):

| scenario | command |
|---|---|
| fast success | `go version` |
| non-zero exit | `go run ./testdata/exit7` |
| stderr output | `go run ./testdata/noisy` |

### Checks

| id | points | what it grades |
|---|---|---|
| `ch2parity` | 10 | chapter 2's log, reducer and three renderers still work |
| `toolloop` | 25 | parse `tool_use`, dispatch, return `tool_result` by id, loop until the model stops asking |
| `multiblock` | 10 | text + two tool calls: all parts recorded, both dispatched, results matched to the right ids |
| `localtools` | 20 | `read_file` (with range), `write_file`, `edit_file`, `list_directory`, `search_files` |
| `runcommand` | 15 | shell executes, stdout/stderr/exit code returned |
| `toolerror` | 15 | a failing or malformed tool call returns an error **to the model** as a `tool_result`; the agent does not crash and does not silently skip |
| `editcontract` | 5 | the declined decision: a choice was made, it is legible in the log, and a failed edit is recoverable by the model |

**Sum: 100.** (Verify with awk over the table, not by eye. The character class
must include the hyphen or hyphenated ids drop out of the sum silently —
this bit us in chapter 2.)

**Weighting rationale**, so it is not re-litigated: `toolerror` is 15 because
returning a failure to the model rather than crashing is a genuinely separable
skill and the most common way a student's agent locks up. `editcontract` is 5
because any coherent answer passes — the points buy legibility, not judgement.
`localtools` is one check rather than five because a student who can implement
one local file tool can implement all of them; splitting points is for when a
student can plausibly have one skill and not the other.

### Grader audit (P9)

Mandatory before this grader ships:

- Delete each behaviour above from the reference solution and confirm the score
  drops. **A row reading 100 → 100 is the finding.**
- Assert the exact set of failing check ids per mutant.
- **Assert that each mutation actually landed.** A silently-unapplied mutation
  scores 100 and manufactures a fake finding — this has now bitten this project
  three times, including once in the commit that added the rule.
- Negative control for anything asserting absence.
- For each check, ask the two vacuity questions: *can the fixture exercise this
  property at all?* and *does the assertion pass vacuously?* Chapter 1's hole
  was a fixture that served one content block, which made walking and indexing
  the same program. Chapter 2's was an exhibit with no opaque material in it.

---

## What exists now for a later chapter

| thing | state after ch3 | collected in |
|---|---|---|
| `run_command`, blocking | works, returns when the process exits | Ch4, which makes it a job |
| `send_input`, `wait_for_job` | named, not built | Ch4 |
| `BlobPart.Path` | still not exercised | Ch4, when output arrives by the megabyte |
| the shell itself | added innocently | Ch8, where it turns out to contain every tool |
| tool arguments from the model | trusted | Ch8 |

---

## Open

1. **`kill_job` — in or out?** Not in the eight. Without it the student cannot
   implement the "kill it" branch of chapter 4's declined decision: three
   answers offered, two buildable. Recommend adding it in chapter 4 rather than
   here, since it has no meaning without a job to kill.
2. Confirm the `ch2parity` check is worth 10 points rather than folded into the
   others. It is a regression guard, and the argument for keeping it visible is
   that chapter 3 is the first chapter that could plausibly break chapter 2's
   work.
