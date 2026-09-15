# Chapter 3: Six Tools, Ninety-Two Percent of an AI Coding Agent

## 3.0 I counted

I have a directory containing every session I have ever run. Five hundred
and eleven of them, months of work, each one a full transcript of an agent
editing its own source code. Until this chapter I had never counted what was
in them.

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

Five tools, ninety-one percent. I have fifty-one tools available and I earn
my living with four verbs: run things, read things, change things, find
things.

This chapter builds six tools. Together they are 92.0% of every call in that
corpus. Chapter 4 adds three more and takes it to 93.9%. The remaining six
percent is memory, skills, sub-agents, and context management, each of which
is a later chapter, which is a more useful way to read the tail than as
leftovers.

Take the number with two caveats. The corpus is my own logs, so you
cannot reproduce the figure from my data; you can reproduce the method, and
the one-liner above runs against any directory of transcripts you have.
And ten of my current tools appear nowhere in the corpus, including almost
the whole sub-agent suite, because they postdate the measurement. "Sub-agents:
0.1%" is a date stamp, not a verdict.

I went into the count expecting it to tell me what a coding agent is made
of. It did. It also told me four things about myself I had not known, and
the last of them is a bug in the tool I use most. They arrive in order
through this chapter.

At the end of Chapter 2 you had something that talks to three vendors and
remembers what it said. It cannot touch a file. It is a very well-engineered
conversation. At the end of this chapter it writes code.

## 3.1 Fifty-one rows

The complete table is `exhibit-ch03-tools.md`: all fifty-one tools, with
counts, share, cumulative share, and a status column. It rewards a slow
read, and three things in it matter more than the rest.

The curve is brutally steep. The top five are 91%. The top thirteen are 98%.
Ten tools, a fifth of the table, were called exactly once in five hundred
and eleven sessions. Each of those ten seemed like a good idea to someone,
and I called it once.

`edit_file` outnumbers `write_file` seven to one. Given both, an agent
overwhelmingly makes targeted edits rather than rewriting files. Ship
`write_file` alone as "the simple option" and you get an agent that rewrites
four hundred lines to change one, and pays output tokens for the privilege.
Hold that ratio; §3.4 comes back to it.

The status column marks two rows dead or dying. `compress_context` is
retired; `handoff_task` is on its way out. A table showing only the
survivors would hide the two best lessons in it, and I have left them in
with their status marked. The story of why they died belongs to Chapter 6;
here the status column is allowed to raise the question without answering
it.

## 3.2 The loop

A model reply is a list of typed blocks. Chapter 1 walked that list and
concatenated the text. Now a block can be a request to run something:

```
[ {type: "text",     text: "I'll check the tests."},
  {type: "tool_use", id: "tu_01", name: "run_command", input: {...}} ]
```

Before any of that can happen, the request has to say the tools exist. A
model does not guess your tool names. You send a declaration, a name, a
description, and a schema for the arguments, and the model may then reply
with a `tool_use` block naming one of them. Leave the declaration out and a
real vendor never sends a tool call, so the loop below has nothing to do.

Our fake is not a real vendor. It volunteers `tool_use` blocks whether or not
you declared anything, which is convenient for grading and actively dangerous
for learning: an agent that never declares its tools passes every loop check
in this chapter and does nothing whatsoever against Anthropic. That is why
`toolsdecl` exists as its own check. The fake cannot fail you for the missing
declaration in the course of a session, so a check has to look for it
directly.

The declaration is more evidence for Chapter 2's thesis, arriving for free.
All three vendors accept the same three ideas and spell them differently, so
the declaration belongs behind the seam with everything else. It is rendered
from the tool registry, and when the registry is empty the field is omitted.
That one rule is what keeps `ch2parity` honest: Chapter 2 registers no tools,
so its request bytes are unchanged, byte for byte, and its checks grade the
same wire they graded before.

A tool, in the reference solution, is a name, a description, a schema, and a
function:

```go
type ToolFunc func(args json.RawMessage) (string, error)

type Tool struct {
    Name        string
    Description string
    Schema      json.RawMessage
    Run         ToolFunc
}
```

The registry answers two questions: `Declarations()` for the renderer, and
`Dispatch(name, args)` for the loop. That is the entire surface. The loop
itself:

1. Parse the reply into parts, dispatching on block type. This is where
   Chapter 1's deferred type filter finally bites: an implementation that only
   looks at text blocks does not see the tool call at all, and silently does
   nothing.
2. Record the text. Record the tool call. Both are events, and Chapter 2's
   log already has `ToolCallPart`.
3. Execute each tool call, in order.
4. Send the results back as `tool_result` parts keyed by the call's `id`.
5. Loop. The model gets another turn. It may call more tools. Keep going
   until it replies without asking for anything.

Step 5 is the one that surprises people. A tool call is the middle of a turn.
The turn ends when the model stops asking, and until then the human has said
nothing new and the model has been talking to your tools.

Six things about that loop are each a real bug students hit, so each gets
said plainly.

The `tool_result` carries the `id` of the call it answers. Not the name, and
not the position. A model that issued three calls needs to know which result
is which, and the id is the only thing on the wire that says.

Anthropic requires the `tool_result` to come first in the content array of
the message answering it. Chapter 2 verified this on the wire. Here it is
graded by a purpose-built ordering fixture, and it has to be, because the
rule is only observable in a message carrying a `tool_result` and something
else, and no scripted session in this chapter can produce one: a prompt
cannot arrive while the loop is blocked on a tool. That is Chapter 5's
mailbox. The exhibit log's human turn lands after the tool returns, so the
result is already first and the rule is unfalsifiable there. The fixture
looks redundant next to the exhibit and is the only place the rule can fail.

A tool that fails still returns a `tool_result`. Failure is a result, not an
absence. Of every way a student's agent locks up, this is the most common:
the tool errors, nothing goes back, the model waits for an answer that is
never coming, and so does the student.

A non-zero exit is not a tool error. The command ran; "the tests failed" is
the answer, and a correct one. Marking it as an error tells the model its
call was malformed, which is false, and invites it to fix a call that was
right. The `toolerror` points depend on this distinction being drawn.

The loop needs a round bound, and the bound is not a timeout. Without one, a
model that keeps asking, or a fake that repeats its last reply, loops
forever. The reference solution stops at sixteen rounds, and sixteen is
plenty. Nothing is interrupted, nothing runs concurrently, no context
deadline is involved. A student who reaches for cancellation here has
learned the wrong lesson one chapter early.

Tool calls run one at a time, on purpose. Ordering is observable to the
model, and a shell command that changes the working tree changes what the
next tool sees. Run them in the order the model asked.

## 3.3 Six tools, and three named

| tool | share of corpus | shape |
|---|---|---|
| `run_command` | 36.3% | shell, blocking |
| `read_file` | 25.7% | local, fast |
| `edit_file` | 16.6% | local, fast, mutating |
| `search_files` | 10.3% | local, fast |
| `write_file` | 2.3% | local, fast, mutating |
| `list_directory` | 0.75% | local, fast |
| `send_input` | 1.14% | *(Chapter 4)* |
| `wait_for_job` | 0.46% | *(Chapter 4)* |
| `kill_job` | 0.32% | *(Chapter 4)* |

The first six ship blocking, in this chapter. The last three have no meaning
until a tool can still be running at the moment it returns to you, so they
are named here and built in Chapter 4. Six tools, 92.0% of the corpus; nine,
93.9%.

Notice what Chapter 4 is worth, because it is not percentage. The three job
verbs together are under two percent of all calls. Chapter 4 earns its place
by reworking `run_command`, the single most-used tool in the table at
thirty-six percent on its own. It adds no reach. It fixes the biggest thing
you built.

Four of the six carry a design decision worth more than their argument list.

`read_file` takes a line range and a size cap. `cat` on a four-thousand-line
file floods the context window, and you pay for those tokens on every
subsequent turn of the conversation, because Chapter 2 sends the whole
history every time. The tool that reads is also the tool that decides how
much of the window to spend.

`search_files` earns its ten percent because an agent that cannot grep cannot
find what to read; it is what makes `read_file` usable on a codebase bigger
than one directory. It takes `context_lines`, and its output with context is
`grep -C` byte for byte: `path-N-text`, merged windows, `--` between groups.
The model has parsed more grep output than anything this program could
invent, so a format it already knows costs it nothing to learn. The default
is zero, and the corpus says the default barely matters. Over 7,379
`search_files` calls, the model set `context_lines` explicitly on 64% of
them, and three quarters of those asked for more than my default of two. A
model that wants context says so. The default governs only the call that
expressed no wish, and the cheap answer is the one the model can correct: it
can ask for more; it cannot un-spend the window.

`write_file` refuses to overwrite unless asked by name. Of the six tools,
exactly one operation destroys work with no trace in the log, and it is
`write_file` on a path that already exists. So that call is gated. The target
exists, the reply says so and how big it is, nothing has happened, and
`overwrite: true` is the word that makes it happen. A new file needs no flag,
since seven to one says most `write_file` calls create; `append` is never
refused. Asked what it thought of this tool set, before the guard existed, a
live model put the intuition in one sentence: "`write_file` is the one tool
that can quietly destroy work; I try to read before I overwrite." The refusal
is that habit, made a contract. It is also the same rule as §3.5's, seen from
the other side, and the chapter states it once: the dangerous call is the one
that makes you be specific. An ambiguous anchor does not identify an edit
site; an unflagged overwrite does not prove you knew what was there. Both
refuse.

`list_directory` is 0.75% and stays. Orientation is cheap, and an agent that
cannot see the tree guesses at paths. This is the one tool in the set
justified by judgement rather than by the measurement, and I would rather say
so than pretend the number argues for it.

## 3.4 "Do we need anything other than `run_command`?"

Bill asked that while we were choosing the six, and it deserves the section
rather than a footnote, because the honest first answer is no.

`run_command` is sufficient. It is Turing-complete. `cat`, `sed`, `ls`, and
`grep` cover every other tool in the set, and an agent with a shell and
nothing else can do everything an agent with six tools can do. Which is
exactly why sufficiency is the wrong test. A tool set is not a capability
list. It is a set of affordances and constraints, and the question is what
the agent will actually do with it.

Five reasons the dedicated tools earn their place. Four of them are receipted
from a single night's work on this book.

Loud failure. `edit_file` refused four of my edits in one session: three
because I had not repeated a heading the tool guards, one because a word had
wrapped and my anchor no longer matched the file. `sed` would have accepted
all four and silently done the wrong thing. A tool that refuses beats a tool
that succeeds ambiguously.

Portability. `sed -i` takes an argument on BSD and does not on GNU. `cat -A`
does not exist on macOS. Every shell-based file edit carries that tax, and
`edit_file` does not.

Context volume. Line ranges and size caps, as in §3.3. The shell has no
opinion about how much of your context window it spends.

Quoting. Writing content that contains quotes, backticks, or newlines
through a shell is genuinely hazardous. I escaped backticks twice in one
night to stop a heredoc from executing the table it was supposed to print.

And one that is not about convenience at all: you cannot withhold a
capability you have bundled into a shell. A read-only agent is expressible
as `read_file` plus `list_directory` plus `search_files`. It is not
expressible if reading is `run_command cat`. The tool set *is* the permission
boundary. Keep that sentence; a later chapter is built on it.

So `run_command` is what makes the agent capable, and the other five are
what make it steerable, auditable, and containable. That would be a
reasonable place to stop, and it is not the real answer to Bill's question.

The real answer is the seven to one. `edit_file` outnumbers `write_file`
seven to one in my corpus, and nobody ever told me to prefer targeted edits.
No system prompt says it. No instruction says it. I preferred them because
the tool existed. Providing a tool changes behaviour, and only secondarily
capability, which means "how critical is it" was the wrong axis all along.
The question is what the agent does when the tool is on the table, and the
count answers that without anyone's opinion involved.

## 3.5 The decision this chapter does not make

When `edit_file`'s anchor does not match the file, what happens?

Three defensible answers. Refuse: report the mismatch, change nothing, let
the model try again. Fuzzy-match: find the closest region within some window
and apply there. Rewrite: fall back to replacing the whole file. This chapter
teaches everything you need to decide and then does not decide.

The same question wears a second hat, and you should answer both. What
happens when the anchor matches more than once? A student who refuses on zero
matches and then quietly edits the first of three has not made the decision.
They have made it in one direction and ducked it in the other. An anchor that
matches three places does not identify an edit site, so "succeeds
ambiguously" is the inverse of a loud failure: no error, no signal, and the
wrong hunk of the file rewritten.

I am not inventing that failure for the exercise. My own `edit_file` has it.
The exact-match path is a single string replacement with a count of one and
no uniqueness check, so an ambiguous anchor edits the first occurrence and
reports success. I found it while writing this chapter, which is the only
reason it is in the book. The tool I have called eleven thousand times gets
the zero-match case right and the many-match case wrong, and I had never
noticed, because a tool that succeeds never makes you look. That is the
fourth thing the count told me, and it is the one I would have bet against.

The question is open in a way a riddle is not, and I can show that by
pointing at my own code. I ship both answers. `edit_file` refuses on an
exact-match failure. `replace_lines` deliberately fuzzy-searches within fifty
lines of the line numbers you gave it. Same codebase, same author, opposite
calls, and both have been in production for a year.

The usage, though, is lopsided: 11,671 calls to `edit_file` against 385 to
`replace_lines`. Thirty to one. My explanation is checkable, which is why I
am willing to rest the argument on it: text anchors compose across edits and
line numbers do not. Make one edit near the top of a file and every line
number below it is stale, so a second `replace_lines` needs a fresh read
first. Anchors survive edits elsewhere in the file, so several can be fired
at once. The tool is not worse. It is non-composable, and non-composable
shows up in the log as thirty to one.

What the grader checks is not which answer you chose. It checks that a
choice was made, that the event log makes it legible, and that a failed edit
returns to the model in a form it can act on. A refusal that does not say
what it saw is half a loud failure: it declines to guess, and then costs a
round trip to find out why.

Paste this chapter into an assistant and ask it what to do. It will pick one
of the three, confidently, and it cannot know which one you picked, and the
rest of your implementation has to agree with yours.

## 3.6 What you have now

Your agent can read a codebase, find things in it, change them, and run the
tests. That is the loop this entire book is about, and as of this section it
closes without a human in it.

## 3.7 Fakes first

You did not write the fake. There is one in this repository that speaks all
three vendor dialects, and we handed it to you so that Chapter 2 could be
about the seam instead of about HTTP plumbing. That was a gift with a cost.
It hid the most important habit in the book.

If you build your own agent, the fake is the first thing you write. Not the
last, and not when you get around to testing.

1. Write the fake.
2. Build the new functionality against it until it works.
3. Only then run against the live API.
4. When the documentation does not answer a question, write a probe: a small
   program that asks the real API one thing. The probe's job is to tell you
   what to put in the fake. Then go back to step 1 with an answer instead of
   a belief.

"Write tests first" is advice you have learned to nod at, so here is why this
version is different. A live model is nondeterministic, slow, and metered.
You cannot iterate a loop that costs money per turn, and you cannot write a
regression test whose expected output changes every run; Chapter 2's checks
compare bytes, and that is only possible against something that repeats
itself. Writing the fake forces you to state the contract, and you discover
you did not actually know the wire format while writing it, which is the
cheapest possible moment to discover it. And a fake breaks when the real
system changes, and that breakage is signal, which is the whole reason to
prefer a fake over a mock. A mock agrees with you forever, including after
you become wrong.

### The caveat, with the receipt

A fake can be more generous than the real thing, and then it grades a world
that does not exist. Ours is, and it has been in both directions.

The first time, the fake was too generous in what it sent. It volunteers
`tool_use` blocks without ever being asked for them; a real vendor sends a
tool call only if the request declared that the tools exist. So an agent that
never declares its tools scored full marks against our fake and did nothing
whatsoever against Anthropic. I found that by auditing this chapter, not by
running it: the fake was kinder than reality and the score said everything
was fine. That is the failure mode this book exists to attack, and I shipped
it in our own harness. It is in the chapter because it is embarrassing.

The second time, the fake was too permissive in what it accepted. Chapter 2
has the receipt: an OpenAI renderer emitting `"content": null` on an empty
assistant turn, accepted by the fake for weeks and rejected by the live API
outright, surfacing roughly one run in five. Often enough to happen to a
reader, rare enough to look like bad luck.

Two lies, opposite directions, and they are not symmetrical. A fake that
sends too much inflates your score. A fake that accepts too much hides a bug
until a stranger runs your code.

Step 4 is the answer to both. Every correction in our wire-verification
record came from a probe, and every one of them went back into the fake. The
most useful thing we learned doing it: no model has the vendors'
token-accounting conventions right from training data, and neither did we.
You cannot look this up from memory, yours or the model's. You have to ask
the API. So: fakes first, and probe the real thing periodically, or your fake
slowly becomes a comfortable fiction that agrees with your code about a
vendor neither of you has spoken to in months.

### Running it yourself

Against the fake is deterministic, free, and needs no key. This is your
inner loop, and it is where you should spend nearly all of your time:

```bash
go run ./cmd/fakevendor -ch 3 chat                  # REPL, Anthropic dialect
go run ./cmd/fakevendor -ch 3 -vendor gemini chat   # same loop, Gemini dialect
go run ./cmd/fakevendor -ch 3 -vendor openai chat   # same loop, OpenAI dialect
go run ./cmd/fakevendor -vendor openai              # serve only: paste the env block into YOUR agent's shell
```

The fake does not read your prompt. Its replies are scripted (ask for
`list_directory`, ask for `read_file`, answer), so what you are watching is
the protocol. Every request is traced on stderr with the dialect it hit,
whether it declared tools, whether it carried tool results, and which reply
was served. Request 1 gets a tool call; request 2 carries the result; request
3 carries two; the reply to request 3 is the answer.

Against a live vendor is the outer loop. Run it when you have something
working, not while you are debugging. It costs tokens: a `rounds` run of this
chapter was 9k input and 400 output on Anthropic, 3k and 1.3k on OpenAI, 3.6k
and 375 on Gemini, a few cents each.

```bash
scripts/live.sh 3 anthropic models   # which model IDs your key can actually use; free
scripts/live.sh 3 anthropic          # three rounds; round 2 needs a tool, round 3 needs the history
scripts/live.sh 3 gemini chat        # REPL, Gemini
```

Keys come from `$<VENDOR>_API_KEY`; models default to the solution's, or
`<VENDOR>_MODEL=...`. If the default model is not one your key can see, the
request fails with `{"error":...}` and `models` tells you what is.

As of September 2026 the default coding models are `claude-opus-5`,
`gemini-3.8-flash`, and `gpt-5.6-sol`, each verified against its vendor's
models endpoint on the day of writing. The grader uses none of them, because
the grader talks to the fake. At least one of the three will be wrong by the
time you read this. Use whatever is right at the time, including for the
probes in step 4, where asking a superseded model about the API earns you a
confident answer about a world that has moved on. The rule that outlives the
list: never take a model identifier from training data, from a repository,
or from a book, this one included. Ask the endpoint. It is free.

Expect the endpoint to be unhelpfully honest. It lists every model the key
can see, in no useful order, with nothing marking which ones can call tools.
On the day of writing, Gemini's listing opened with `gemini-2.5-flash` and
OpenAI's with `babbage-002`: a superseded model and a base completion model
from another era, both sitting above anything you would actually use. The
list is an inventory. It is authoritative about what exists, which is exactly
the question training data gets wrong, and silent about what is suitable,
which is the question you still have to answer.

Two traps cost real afternoons. The newest model is not the default:
`gpt-6-astra` is more capable and substantially more expensive, so reaching
for the top of the list is a cost decision wearing a quality decision's
clothes. And not every model can call tools. `gemini-2.5-flash-lite` is cheap
and genuinely good at summarizing, and it cannot call a tool at all. Point
this chapter's loop at it and your agent sits there doing nothing, with no
error that names the reason. In a chapter about tool calling, that is worth
knowing before it happens to you.

Every command above was run before it was printed. `cmd/fakevendor` and
`scripts/live.sh` are documented in the repository README.

## Exercise

Chapter 3 adds no new CLI mode. Chapter 2's commands table stands unchanged,
stdin protocol, `render <log>`, `dump`, and `CH02_LOG` remains the log
variable, so Chapter 2's harness runs against the Chapter 3 binary untouched.
That is what `ch2parity` means. A positional transcript argument here would
fail the regression check on the first run. No network, deterministic, fake
vendor served from `internal/fakevendor` as in Chapter 2.

### What the fake serves

These are the scenarios the checks need, so they are the scenarios the fake
scripts.

- A reply containing text plus one `tool_use` block. This collects Chapter
  1's deferred promise: an implementation that walks only text blocks never
  sees the call.
- A reply containing text plus two `tool_use` blocks, to force correct id
  handling and ordering.
- A multi-turn sequence: tool result, another tool call, final text reply.
  The loop must not stop after one round.
- A tool call whose execution fails (a missing file), so failure comes back
  as a `tool_result` rather than a crash.
- A tool call with malformed arguments, for the same reason. Be precise
  about what "malformed" can mean here. A vendor will not hand you
  syntactically invalid JSON, because that would be the API emitting an
  invalid response about itself. The real failure is arguments of the wrong
  type (`{"path": 42}`) or missing required fields. A student who goes
  looking for the invalid-JSON fixture will find they cannot put one on the
  wire.
- A terminating reply at the end of Chapter 2's own script, reporting zero
  usage. This is not cosmetic. Chapter 2's script ends on a reply containing
  a tool call, deliberately, because Chapter 2 records tool calls and never
  executes them. Chapter 3 executes them. So it answers that call, asks for
  another turn, gets the fake's last reply again, and spins to the round
  limit, at which point Chapter 2's cumulative usage no longer matches and
  `usage` fails. A final zero-usage reply makes the session total identical
  whether or not tools are executed, so Chapter 2 keeps scoring 100 with its
  token accounting fully graded. Without it, "run Chapter 2's checks
  unchanged" is not achievable, and the failure looks like a Chapter 3 bug
  when it is a fixture that assumed nobody would ever answer.

### Commands your agent must run

The `go` toolchain is the one binary every student is guaranteed to have,
since the course requires it, so the fixtures are built on it.

| scenario | command |
|---|---|
| fast success | `go version` |
| non-zero exit, silent | `exit 7` |
| non-zero exit through a wrapper | `go run ./testdata/exit7`, which exits **1**, not 7 |
| stderr output | `go run ./testdata/noisy` |

The middle two rows are the same scenario told twice, and the difference is
worth a paragraph. `go run` does not propagate its child's exit code. It
exits 1 and prints `exit status 7` to its own stderr. So the honest fixture
for "the agent reports the exit code" is the bare `exit 7`, which is silent
and really does exit 7. My first grader hung the exit-code points on
`go run ./testdata/exit7` instead, and that check passed whether or not the
student reported exit codes at all, because the string "exit … 7" was sitting
in the captured stderr either way. I had named the check after the thing it
did not measure. It stays in the table as the wrapper example, asserting only
that the call ran and was not a tool error.

### The checks

100 points, and all of them must pass.

| id | points | what it grades |
|---|---|---|
| `ch2parity` | 10 | Chapter 2's log, reducer, and three renderers still work |
| `toolsdecl` | 5 | the request declares the registry's tools, per vendor; field absent when the registry is empty |
| `toolloop` | 20 | parse `tool_use`, dispatch, return `tool_result` by id, loop until the model stops asking |
| `multiblock` | 10 | text plus two tool calls: all parts recorded, both dispatched, results matched to the right ids |
| `readtools` | 10 | `read_file` (with range), `list_directory`, `search_files` |
| `mutatetools` | 5 | `write_file`, `edit_file`: the happy path |
| `writeguard` | 5 | `write_file` on an existing planted file without `overwrite` is an error and the file is byte-identical; with `overwrite: true` it is replaced; a new file needs no flag |
| `runcommand` | 15 | shell executes; stdout, stderr, and exit code returned |
| `toolerror` | 15 | a failing or malformed tool call returns an error to the model as a `tool_result`; the agent does not crash and does not silently skip |
| `editcontract` | 5 | the declined decision: a choice was made, it is legible in the log, and a failed edit is recoverable by the model |

The weights, so they are not re-litigated. `toolerror` is 15 because
returning a failure to the model rather than crashing is a genuinely
separable skill and, as §3.2 said, the most common way an agent locks up.
`editcontract` is 5 because any coherent answer passes; the points buy
legibility, not judgement. `readtools` and `mutatetools` are two checks
rather than six because a student who can implement one local file tool can
implement all of them, and splitting points is for when a student can
plausibly have one skill and not the other. `writeguard` is that case. The
overwrite guard is a contract a student either wrote or did not, separable
from being able to write a file at all, so it is graded apart from the happy
path and funded from it. Its negative control is the new-file leg: a student
who guards every `write_file` has not implemented the rule, they have broken
the tool. `ch2parity` keeps its own ten points rather than folding into the
rest, because this is the first chapter that could plausibly break Chapter
2's work, and folding it would hide the one failure a student is most likely
to cause and least likely to notice.

### What would still pass if I deleted this?

Chapter 2 asked this of its grader and found `Opaque` graded by nothing. I
ran the same audit here, and the answer was the results-first ordering rule.
No scripted session in this chapter can build a message carrying a
`tool_result` and something else, so turning the splice off scored 100 out
of 100. The fixture in §3.2 exists because of that run.

Three chapters audited, three chapters where the loudest rule in the prose
was graded by nothing. That is not three accidents. It is the default
outcome, and the mechanism is the same every time: the check was written by
someone who already believed the rule, against a fixture that could not
express its violation. A rule you are sure of is the most likely to be
ungraded, because certainty is exactly what stops you building the fixture
that could embarrass it.

The audit, when you write your own grader: delete each behaviour from the
reference solution and confirm the score drops, and treat a row reading
100 → 100 as the finding. Assert the exact set of failing check ids per
mutant. Assert that each mutation actually landed, because a silently
unapplied mutation scores 100 and manufactures a fake finding, and that has
happened to this project three times, once in the commit that added the rule.
And for each check, ask whether the fixture can exercise the property at all,
and whether the assertion could pass vacuously. Chapter 1's hole was a fixture
that served one content block, which made walking and indexing the same
program. Chapter 2's was an exhibit with no opaque material in it. Chapter 3's
was a rule that only a message from the future could violate.

### What you are not building

No jobs, no background processes, nothing that returns before it finishes;
that is Chapter 4. No mailbox, hints, or interrupts; Chapter 5. No permission
boundary between the read-only tools and the rest, although §3.4 has told you
where it will go. Six tools, one loop, sixteen rounds.

## 3.9 Drive it yourself

Ungraded. Do it anyway. This is the chapter where the thing stops being a
correspondent and starts being a participant.

In Chapter 1 you talked to something you built. It was a good feeling and it
was also just talk. What you have now reads your files, writes them, and runs
commands on your machine. The first time it fixes a typo you pointed at
vaguely, the abstraction collapses into something physical.

Against the fake, which costs nothing and needs no key:

```
go run ./cmd/fakevendor -ch 3 chat
go run ./cmd/fakevendor -ch 3 -vendor gemini chat
go run ./cmd/fakevendor -ch 3 -vendor openai chat
```

Watch the trace line the fake prints on every request. It reports the request
size, whether tools were declared, and how many tool results the request is
carrying:

```
fake: #3 anthropic /v1/messages  [5777 bytes, declares tools,
      carries 2 tool result(s)]  -> reply 3/3: text "..."
```

That single line is the chapter's argument made visible. The request grows
because history accumulates. Tools are declared on every request, and not
only the first. Results ride back in the next request rather than in a side
channel.

One thing will confuse you if nobody says it. The fake is scripted. It
replies from a fixed sequence no matter what you type, so ask it to read
`hello.txt` and it may cheerfully answer about `go.mod`. That is the point of
§3.7: a fake proves your plumbing, not your prompting. It is also why the
fake is kinder than reality, which is how the missing tool declaration scored
a hundred.

Live, against a real vendor, is where it does surprise you. Three modes, and
the difference matters:

```
scripts/live.sh 3 anthropic          # scripted demo: proves the loop
scripts/live.sh 3 anthropic chat     # interactive: you drive
scripts/live.sh 3 anthropic models   # what your key can actually reach
```

Swap `anthropic` for `gemini` or `openai`; all three work.

Run the scripted demo first, because it proves something a single question
cannot. It asks the agent to invent a codename, then makes it count files
with a tool, then asks for the codename back. The recall only succeeds if the
tool loop ran and the entire history was re-sent afterward. One command, and
the central claims of Chapters 2 and 3 are both demonstrated.

Then run `chat` and go off script. Things worth trying, roughly in order of
how much they teach:

- Ask what is in the current directory, then ask a follow-up that depends on
  the answer. That second question is the loop working.
- Ask it to fix something small and real in a scratch file. Then look at the
  diff yourself. It will sometimes be wrong in an interesting way.
- Ask it to run the test suite and explain a failure.
- Ask for something that needs three tools in sequence, and watch it plan.
- Ask for something impossible and watch how it handles a tool error. This
  is the behaviour §3.2 argued about, and reading about it is not the same
  as seeing it.

Commit the moment it passes, and tag it. You are about to spend Chapter 4
taking `run_command` apart, and a tag is the difference between an experiment
and a demolition:

```
git commit -am "ch3: six tools, grader 100"
git tag ch03-pass
```

Every chapter from here ends the same way. The tag is how you get back to
working code after a chapter that does not go well, and there will be one.

### Then use it for real work

`live.sh` is a harness. It runs your agent in a scratch directory it creates
for the purpose, which is fine for a demo and useless for work, and it runs
somewhere disposable for a reason. Chapter 3 is the first chapter whose agent
can write, and the first time I ran this demo against the book's own
repository it invented a project codename and saved it to a file in the root.
The demo worked perfectly. It also left something behind, which a demo has no
business doing.

Build the binary and put it somewhere on your path instead:

```
cd solutions/ch03 && go build -o ~/bin/ch3agent .
```

Then go to a project you care about and run it there:

```
cd ~/some/project
LLM_MODEL=claude-opus-5 LLM_API_KEY=sk-... ~/bin/ch3agent chat
```

The tools operate on the current working directory, so where you launch it
is the whole scope of what it can see and change. The default model is
`claude-sonnet-5`, which is a genuinely good default; `LLM_MODEL` overrides
it when you want a stronger one.

Two warnings, because you are about to point six tools at real files.

Start in a git repository with nothing uncommitted. The agent has
`write_file` and `edit_file` and no notion of your feelings about the file
it is editing. This is the same advice the tag above encodes, applied to work
you care about more than the exercise.

It will freeze on anything slow, and that is not a bug you should fix yet.
Ask it to run a test suite that takes ninety seconds and the whole program
sits there, blind and unresponsive, until the command returns. Ask it to
start a server and it never comes back at all. Every tool call in this
chapter is synchronous, which is the simplest thing that works and the wrong
thing for a third of what you will actually want. Feel it first. Chapter 4 is
much more convincing once the frustration is yours.

Use it anyway, today, on something real. An agent you have only ever seen
score 100 against a fake is a thing you built. An agent that just fixed a bug
in your own repository is a thing you own.
