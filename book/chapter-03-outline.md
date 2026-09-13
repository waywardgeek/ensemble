# Chapter 3 — Six Tools: Ninety-Two Percent of an AI Coding Agent

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
six tools in their simplest honest form, and the thing comes alive.

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

This chapter builds six tools. Together they are **92.0%** of every tool call in
that corpus. Chapter 4 adds three more and takes it to 93.9%. The remaining six
percent is memory, skills, sub-agents and context management — and each of those
is a later chapter, which is a more useful way to read the tail than as
leftovers.

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

1. **The curve is brutally steep.** Top five is 91%. Top thirteen is 98%. Ten
   tools — a fifth of the table — were called exactly once in five hundred and
   eleven sessions.

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

**Before any of that: the request has to say the tools exist.** A model does not
guess your tool names. You send a declaration — name, description, and a schema
for the arguments — and the model may then reply with a `tool_use` block naming
one of them. Leave the declaration out and a real vendor never sends a tool call
at all, so the loop below has nothing to do. Our fake volunteers `tool_use`
blocks unprompted, which is convenient for grading and actively dangerous for
learning: an agent that never declares its tools scores full marks here and does
nothing whatsoever against Anthropic. Say this out loud in the chapter, because
it is the one bug in chapter 3 the harness cannot fail you for.

This is also more evidence for chapter 2's thesis, arriving for free: all three
vendors accept the same three ideas and spell them differently, so the
declaration belongs behind the seam with everything else.

The declaration is rendered from the registry, and **the field is omitted when
the registry is empty**. That is what keeps `ch2parity` honest: chapter 2
registers no tools, so its request bytes are unchanged, byte for byte, and its
checks still grade the same wire.

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
  record.) This rule is graded by a purpose-built ordering fixture, and it has
  to be. The rule is only observable in a message carrying a `tool_result` *and
  something else*, and no scripted session in this chapter can produce one,
  because a prompt cannot arrive while the loop is blocked on a tool. That is
  chapter 5's mailbox. Do not delete the fixture because it looks redundant next
  to the exhibit log — the exhibit's human turn lands *after* the tool returns,
  so the result is already first and the rule is unfalsifiable there.
- A tool that fails still returns a `tool_result`. Failure is a *result*, not an
  absence. This is the single most common way a student's agent locks up.
- **A non-zero exit is not a tool error.** The command ran; "the tests failed"
  is the answer, not a broken call. Marking it as an error tells the model its
  call was malformed, which is false, and invites it to "fix" a call that was
  correct. This distinction is what keeps the `toolerror` points honest.
- **The loop needs a round bound, and it is not a timeout.** Without one, a model
  that keeps asking — or a fake that repeats its last reply — loops forever.
  Sixteen rounds is plenty. Say explicitly that this is a *bound*, not
  cancellation: nothing is interrupted, nothing runs concurrently. A student who
  reaches for a context deadline here has learned the wrong lesson one chapter
  early.
- Sequential execution is a deliberate choice here, not an oversight. I run tool
  calls one at a time on purpose. Say so, and say why: ordering is observable to
  the model, and a shell command that changes the working tree changes what the
  next tool sees.

---

## §3.3 The six tools

| tool | share of corpus | shape |
|---|---|---|
| `run_command` | 36.3% | shell, blocking |
| `read_file` | 25.7% | local, fast |
| `edit_file` | 16.6% | local, fast, mutating |
| `search_files` | 10.3% | local, fast |
| `write_file` | 2.3% | local, fast, mutating |
| `list_directory` | 0.75% | local, fast |
| `send_input` | 1.14% | *(chapter 4)* |
| `wait_for_job` | 0.46% | *(chapter 4)* |
| `kill_job` | 0.32% | *(chapter 4)* |

The first six ship blocking, in this chapter. The last three have no meaning
until a tool can still be running at the moment it returns to you, so they are
named here and built in chapter 4.

**Six tools, 92.0% of the corpus. Nine, 93.9%.**

Note what chapter 4 is worth, because it is not percentage. The three job verbs
together are under two percent of all calls. Chapter 4 earns its place by
reworking `run_command` — the single most-used tool in the table, thirty-six
percent on its own. It does not add reach. It fixes the biggest thing you built.

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

**The same question wears a second hat, and the chapter should ask both:** what
happens when the anchor matches *more than once*? A student who refuses on zero
matches and then silently edits the first of three has not actually made the
decision — they have made it in one direction and ducked it in the other. An
anchor matching three places does not identify an edit site, so "succeeds
ambiguously" is the inverse of a loud failure: no error, no signal, and the
wrong hunk of the file rewritten.

I am not inventing that failure mode for the exercise. My own `edit_file` has
it. The exact-match path is a single string replacement with a count of one and
no uniqueness check, so an ambiguous anchor quietly edits the first occurrence
and reports success. I found it while writing this chapter, which is the only
reason it is in the book: the tool I have called eleven thousand times gets the
zero-match case right and the many-match case wrong, and I had never noticed,
because a tool that succeeds never makes you look.

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

## §3.7 Fakes first

You did not write the fake. There is one in this repository that speaks all
three vendor dialects, and we handed it to you so that chapter 2 could be about
the seam instead of about HTTP plumbing. That was a gift with a cost: it hid the
most important habit in the book.

**If you build your own agent, the fake is the first thing you write.** Not the
last, not "when we get around to testing." First.

The working loop:

1. **Write the fake.**
2. **Build the new functionality against it** until it works.
3. **Only then run against the live API.**

And one exception, which is where most people go wrong by skipping it:

4. **When the documentation does not answer a question, write a probe.** A small
   program that asks the real API one thing. The probe's job is *not* to test
   your agent. Its job is to tell you what to put in the fake. Then you go back
   to step 1 with an answer instead of a belief.

Say plainly why, because "write tests first" is advice students have learned to
nod at and ignore:

- **A live model is nondeterministic, slow, and metered.** You cannot iterate a
  loop that costs money per turn, and you cannot write a regression test whose
  expected output changes every run. Chapter 2's checks compare *bytes*. That is
  only possible against something that repeats itself.
- **Writing the fake forces you to state the contract.** You discover you did not
  actually know the wire format while writing the fake — which is the cheapest
  possible moment to discover it. Every hour after that is more expensive.
- **A fake breaks when the real system changes, and that breakage is signal.**
  This is the whole reason to prefer a fake over a mock. A mock agrees with you
  forever, including after you become wrong.

### The honest caveat, and we have the receipt

A fake can be **more generous than the real thing**, and then it grades a world
that does not exist.

Ours is. The fake volunteers `tool_use` blocks without ever being asked for
them. A real vendor does not: it sends a tool call only if the request declared
that the tools exist. So an agent that never declares its tools scores **full
marks against our fake and does nothing whatsoever against Anthropic** — and we
found that by auditing this chapter, not by running it, because the fake was
kinder than reality and the score said everything was fine.

That is the failure mode this book exists to attack, and we shipped it in our
own harness. It is in the chapter now precisely because it is embarrassing.

It happened twice, in opposite directions, which is what makes it a rule rather
than an anecdote. The first time the fake was too generous in what it **sent**.
The second time it was too permissive in what it **accepted**: our OpenAI
renderer emitted `"content": null` on an empty assistant turn, and the fake took
it without complaint for weeks. The real API rejects it outright. The bug only
ever appeared live, intermittently, and only when a reply was truncated at the
token limit: roughly one run in five, which is the worst possible frequency,
often enough to happen to a reader and rare enough to look like bad luck.

So the fake lies in both directions of the seam, and the two lies are not
symmetrical. A fake that sends too much inflates your score. A fake that accepts
too much hides a bug until a stranger runs your code.

This is what step 4 is for. The corrections in our wire-verification record all
came from probes, and every one of them went back into the fake. The most
useful thing we learned doing it: **no model has the vendors' token-accounting
conventions right from training data, and neither did we.** You cannot look this
up from memory, yours or the model's. You have to ask the API.

So: fakes first — *and* probe the real thing periodically, or your fake slowly
becomes a comfortable fiction that agrees with your code about a vendor neither
of you has spoken to in months.

### Running it yourself

Two things you should be able to do by hand, and the chapter should print the
exact commands:

- **Against the fake.** Deterministic, free, no key required. This is your inner
  loop, and it is where you should spend nearly all of your time.

  ```bash
  go run ./cmd/fakevendor -ch 3 chat                  # REPL, Anthropic dialect
  go run ./cmd/fakevendor -ch 3 -vendor gemini chat   # same loop, Gemini dialect
  go run ./cmd/fakevendor -ch 3 -vendor openai chat   # same loop, OpenAI dialect
  go run ./cmd/fakevendor -vendor openai              # serve only: paste the env block into YOUR agent's shell
  ```

  The fake does not read your prompt. Its replies are scripted — ask for
  `list_directory`, ask for `read_file`, answer — so what you are watching is
  the protocol: every request is traced on stderr with the dialect it hit,
  whether it declared tools, whether it carried tool results, and which reply
  was served. Request 1 gets a tool call; request 2 carries the result; request
  3 carries two; the reply to request 3 is the answer.

- **Against a live vendor** — any of the three. This is the outer loop. Run it
  when you have something working, not while you are debugging. It costs
  tokens: a `rounds` run of chapter 3 was 9k input / 400 output on Anthropic,
  3k / 1.3k on OpenAI, 3.6k / 375 on Gemini — a few cents.

  ```bash
  scripts/live.sh 3 anthropic models   # which model IDs your key can actually use — free
  scripts/live.sh 3 anthropic          # three rounds; round 2 needs a tool, round 3 needs the history
  scripts/live.sh 3 gemini chat        # REPL, Gemini
  ```

  Keys come from `$<VENDOR>_API_KEY`; models default to the solution's, or
  `<VENDOR>_MODEL=...`. If the default model is not one your key can see, the
  request fails with `{"error":...}` and `models` tells you what is.

  **Which model.** As of this writing the default coding models are
  `claude-opus-5`, `gemini-3.8-flash` and `gpt-5.6-sol`, each verified against
  its vendor's models endpoint on the day of writing. The grader uses none of
  them, because the grader talks to the fake.

  At least one of those three will be wrong by the time you read this. Use
  whatever is right *at the time* — including for the probes in step 4, where
  asking a superseded model about the API earns you a confident answer about a
  world that has moved on. That is what the `models` subcommand is for, and it
  is free. The rule that outlives the list: never take a model identifier from
  training data, from a repository, or from a book, this one included. Ask the
  endpoint. See P10.

  Expect the answer to be unhelpfully honest. The endpoint lists every model the
  key can see, in no useful order, with nothing marking which ones can call
  tools. On the day of writing, Gemini's listing opened with `gemini-2.5-flash`
  and OpenAI's with `babbage-002`: a superseded model and a base completion
  model from another era, both sitting above anything you would actually use.
  The list is an inventory, not a recommendation. It is authoritative about what
  *exists*, which is exactly the question training data gets wrong, and silent
  about what is *suitable*, which is the question you still have to answer.

  Two traps that cost real afternoons. **The newest model is not the default** —
  `gpt-6-astra` is more capable and substantially more expensive, so reaching
  for the top of the list is a cost decision wearing a quality decision's
  clothes. And **not every model can call tools**: `gemini-2.5-flash-lite` is
  cheap and genuinely good at summarizing, and it cannot call a tool at all.
  Point this chapter's loop at it and your agent sits there doing nothing, with
  no error that names the reason. In a chapter about tool calling, that is worth
  knowing before it happens to you rather than after.

*(Every command above was run before it was printed. `cmd/fakevendor` and
`scripts/live.sh` are documented in the repository README.)*

## §3.8 The exercise

**Contract:** chapter 3 adds **no new CLI mode**. Chapter 2's commands table
stands unchanged — stdin protocol, `render <log>`, `dump` — and `CH02_LOG`
remains the log variable, so chapter 2's harness runs against the chapter 3
binary untouched. That is not a convenience; it is what `ch2parity` *means*. A
positional transcript argument here would fail the regression check on the first
run. No network, deterministic, fake vendor served from `internal/fakevendor` as
in chapter 2.

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
- A tool call with **malformed arguments**, same reason. Say plainly what
  "malformed" can and cannot mean here: a vendor will not hand you syntactically
  invalid JSON, because that would be the API emitting an invalid response about
  itself. The real failure is arguments of the wrong *type* (`{"path": 42}`) or
  missing required fields. A student who goes looking for the invalid-JSON
  fixture will discover they cannot put one on the wire, and should not have to
  discover it the slow way.
- A **terminating reply** at the end of chapter 2's own script, reporting zero
  usage. This is not optional and it is not cosmetic. Chapter 2's script ends on
  a reply containing a tool call, deliberately, because chapter 2 records tool
  calls and never executes them. Chapter 3 *executes* them — so it answers that
  call, asks for another turn, gets the fake's last reply again, and spins to
  the round limit, at which point chapter 2's cumulative usage no longer matches
  and `usage` fails. A final zero-usage reply makes the session total identical
  whether or not tools are executed, so chapter 2 keeps scoring 100 with its
  token accounting fully graded. Say this in the chapter: "run chapter 2's
  checks unchanged" is otherwise not achievable, and the failure looks like a
  chapter 3 bug when it is a fixture that assumed nobody would ever answer.

**Commands the student's agent must be able to run** (the `go` toolchain is the
one binary every student is guaranteed to have, since the course requires it):

| scenario | command |
|---|---|
| fast success | `go version` |
| non-zero exit, silent | `exit 7` |
| non-zero exit **through a wrapper** | `go run ./testdata/exit7` — exits **1**, not 7 |
| stderr output | `go run ./testdata/noisy` |

The middle two rows are the same scenario told twice, and the difference is
worth a paragraph rather than a footnote. `go run` does not propagate its
child's exit code: it exits **1** and prints `exit status 7` to *its own*
stderr. So the honest fixture for "the agent reports the exit code" is the bare
`exit 7`, which is silent and really does exit 7. Grade the exit code on
`go run ./testdata/exit7` instead and the check passes whether or not the
student reports exit codes at all, because the string "exit … 7" is sitting in
the captured stderr either way.

Two lessons, one fixture. For the student: a wrapper between you and the process
can rewrite the result, and `go run` is the one they will hit first. For us: an
assertion that cannot fail is not a check, and this one was *named* after the
thing it did not measure.

### Checks

| id | points | what it grades |
|---|---|---|
| `ch2parity` | 10 | chapter 2's log, reducer and three renderers still work |
| `toolsdecl` | 5 | the request declares the registry's tools, per vendor; field absent when the registry is empty |
| `toolloop` | 20 | parse `tool_use`, dispatch, return `tool_result` by id, loop until the model stops asking |
| `multiblock` | 10 | text + two tool calls: all parts recorded, both dispatched, results matched to the right ids |
| `readtools` | 10 | `read_file` (with range), `list_directory`, `search_files` |
| `mutatetools` | 10 | `write_file`, `edit_file` |
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
  Chapter 3's was the results-first ordering rule: no scripted session in the
  chapter can build a message carrying a `tool_result` *and something else*, so
  turning the splice off scored 100/100.

  **Three chapters audited, three chapters where the loudest rule in the prose
  was graded by nothing.** That is not three accidents, it is the default
  outcome, and it is the whole argument for P9. The mechanism is always the
  same: the check was written by someone who already believed the rule, against
  a fixture that could not express its violation. A rule you are *sure* of is
  the most likely to be ungraded, because certainty is exactly what stops you
  building the fixture that could embarrass it.

---

## §3.9 Drive it yourself

Ungraded. Do it anyway. This is the chapter where the thing stops being a
correspondent and starts being a participant.

In Chapter 1 you talked to something you built. It was a good feeling and it was
also just talk. What you have now reads your files, writes them, and runs
commands on your machine. The first time it fixes a typo you pointed at
vaguely, the abstraction collapses into something physical.

**Against the fake, which costs nothing and needs no key:**

    go run ./cmd/fakevendor -ch 3 chat
    go run ./cmd/fakevendor -ch 3 -vendor gemini chat
    go run ./cmd/fakevendor -ch 3 -vendor openai chat

Watch the trace line the fake prints on every request. It reports the request
size, whether tools were declared, and how many tool results the request is
carrying:

    fake: #3 anthropic /v1/messages  [5777 bytes, declares tools,
          carries 2 tool result(s)]  -> reply 3/3: text "..."

That single line is the chapter's whole argument made visible. The request grows
because history accumulates. Tools are declared on every request, not just the
first. Results ride back in the next request rather than in a side channel.

**One thing that will confuse you if nobody says it.** The fake is *scripted*. It
replies from a fixed sequence no matter what you type, so ask it to read
`hello.txt` and it may cheerfully answer about `go.mod`. That is not a bug, and
it is the point of §3.7: a fake proves your plumbing, not your prompting. It is
also why the fake is kinder than reality, which is how the missing tool
declaration scored a hundred.

**Live, against a real vendor,** where it does surprise you. Three modes, and
the difference matters:

    scripts/live.sh 3 anthropic          # scripted demo: proves the loop
    scripts/live.sh 3 anthropic chat     # interactive: you drive
    scripts/live.sh 3 anthropic models   # what your key can actually reach

Swap `anthropic` for `gemini` or `openai`; all three work.

The scripted demo is the one to run first, because it proves something a single
question cannot. It asks the agent to invent a codename, then makes it count
files with a tool, then asks for the codename back. The recall only succeeds if
the tool loop ran *and* the entire history was re-sent afterward. One command,
and the two central claims of Chapters 2 and 3 are both demonstrated.

Then run `chat` and go off script. That is the mode the list below assumes.

Every command in this section was run before it was printed.

**Things worth trying,** roughly in order of how much they will teach you:

- Ask what is in the current directory, then ask a follow-up that depends on the
  answer. That second question is the loop working.
- Ask it to fix something small and real in a scratch file. Then look at the
  diff yourself. It will sometimes be wrong in an interesting way.
- Ask it to run the test suite and explain a failure.
- Ask for something that needs three tools in sequence, and watch it plan.
- Ask for something impossible and watch how it handles a tool error. This is
  the behavior §3.2 argued about, and reading about it is not the same as
  seeing it.

**Commit the moment it passes.** Tag it. You are about to spend Chapter 4 taking
`run_command` apart, and a tag is the difference between an experiment and a
demolition:

    git commit -am "ch3: six tools, grader 100"
    git tag ch03-pass

Every chapter from here ends the same way. The tag is how you get back to
working code after a chapter that does not go well, and there will be one.

### Then use it for real work

This is the part to actually do.

`live.sh` is a harness: it runs your agent in a scratch directory it creates for
the purpose, which is fine for a demo and useless for work. It runs somewhere
disposable for a reason. Chapter 3 is the first chapter whose agent can *write*,
and the first time we ran this demo against the book's own repository it invented
a project codename and saved it to a file in the root. The demo worked perfectly.
It also left something behind, which a demo has no business doing.

Build the binary and put it somewhere on your path instead:

    cd solutions/ch03 && go build -o ~/bin/ch3agent .

Then go to a project you care about and run it there:

    cd ~/some/project
    LLM_MODEL=claude-opus-5 LLM_API_KEY=sk-... ~/bin/ch3agent chat

The tools operate on the **current working directory**, so where you launch it
is the whole scope of what it can see and change. The default model is
`claude-sonnet-5`, which is a genuinely good default; `LLM_MODEL` overrides it
when you want a stronger one.

Two honest warnings, because you are about to point six tools at real files.

**Start in a git repository with nothing uncommitted.** The agent has
`write_file` and `edit_file` and no notion of your feelings about the file it is
editing. This is the same advice the tag above encodes, applied to work you care
about more than the exercise.

**It will freeze on anything slow, and that is not a bug you should fix yet.**
Ask it to run a test suite that takes ninety seconds and the whole program sits
there, blind and unresponsive, until the command returns. Ask it to start a
server and it never comes back at all. Chapter 4 is that problem: every tool
call in this chapter is synchronous, which is the simplest thing that works and
the wrong thing for a third of what you will actually want. Feel it first. The
next chapter is much more convincing once the frustration is yours.

Use it anyway, today, on something real. An agent you have only ever seen score
100 against a fake is a thing you built. An agent that just fixed a bug in your
own repository is a thing you own.

---

## What exists now for a later chapter

| thing | state after ch3 | collected in |
|---|---|---|
| `run_command`, blocking | works, returns when the process exits | Ch4, which makes it a job |
| `send_input`, `wait_for_job`, `kill_job` | named, not built | Ch4 |
| `BlobPart.Path` | still not exercised | Ch4, when output arrives by the megabyte |
| the shell itself | added innocently | Ch8, where it turns out to contain every tool |
| tool arguments from the model | trusted | Ch8 |

---

## Ruled

1. **`kill_job` ships in chapter 4, not here.** It has no meaning without a job
   to kill. Without it the student could not implement the "kill it" branch of
   chapter 4's declined decision — three answers offered, two buildable.
2. **`ch2parity` keeps its own 10 points rather than folding.** It is a
   regression guard, and chapter 3 is the first chapter that could plausibly
   break chapter 2's work. Folding it would hide the one failure a student is
   most likely to cause and least likely to notice.

3. **The agent declares its tools, in chapter 3.** Deferring it would ship a
   chapter whose payoff sentence — "your agent can now write code" — is false
   against every real vendor, while scoring 100 against our fake. That is the
   precise failure this book was written to attack, and it is not allowed to
   appear in the book's own exercises. The declaration is rendered from the
   registry and **omitted when the registry is empty**, which leaves chapter 2's
   request bytes identical and `ch2parity` honest.
4. **`localtools` splits into `readtools` (10) and `mutatetools` (10).** Reading
   a file and changing one are separable skills, and a student can plausibly
   have one working and not the other — which is the stated test for splitting a
   check. `list_directory` and `search_files` do *not* split out, because
   nobody has search working and read broken. The split also pre-stages chapter
   8, where the read-only set and the mutating set stop being a grading
   convenience and become a permission boundary.
5. **The exit-code points hang on a bare `exit 7`, not on `go run`.** `go run`
   exits 1 and prints `exit status 7` to its own stderr, so the named fixture
   both misdescribed itself and could not fail. It stays in the table as the
   wrapper example, asserting only that the call ran and was not a tool error.
6. **Chapter 3 adds no new CLI mode.** Chapter 2's commands table and `CH02_LOG`
   stand unchanged; `ch2parity` requires it.

## Open

**Work this review creates, for the coder:**

- Implement the tool declaration for all three vendors, rendered from the
  registry, omitted when empty. Then confirm `-ch 2` is still 100 **and** that
  chapter 2's request bytes are byte-for-byte unchanged — the second is the real
  check, since the first can pass while the bytes drift.
- Audit the declaration by deletion like everything else: an agent that sends no
  `tools` field must lose points.

**Not verified, do not print:** that the vendor text-editor tools error on a
non-unique anchor. I believe it, I have not measured it, and §3.5's argument
does not need it — the measured example is my own agent's, and that one I can
show.
