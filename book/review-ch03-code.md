# Review — Chapter 3 solution and grader

**From:** the coder. **Date:** 2026-09-13.
**Spec:** `book/chapter-03-outline.md`. **Brief:** `book/brief-ch03-code.md`.

Built: `solutions/ch03/` (ch02 + `tools.go` + a looping `Ask`), seven checks in
`internal/grade/ch03_{harness,checks}.go`, and a 22-mutant deletion audit in
`internal/grade/ch03_grader_test.go`.

`-ch 1` 100/100, `-ch 2` 100/100, `-ch 3` 100/100. Points sum to exactly 100,
checked two ways (awk over the outline table, and a test that sums the checks
themselves so the code cannot drift from its own arithmetic).

The chapter is right about the shape of the work. It really is small — the ch02
foundation already had `ToolCallPart`, `ToolResultPart{IsError}`, `ToolCalled`,
`ToolReturned`, a `ToolsPending` state and a results-first splice. Chapter 3 adds
a registry, an executor and a loop, and nothing else. The "it should feel fast"
claim survives contact with the implementation.

---

## Must-fix

**1. §3.7's contract line is wrong and contradicts `ch2parity`.**

> **Contract:** `./ch03 <transcript-file>`

Chapter 1 and Chapter 2 are both JSON-lines on stdio plus subcommands, and both
outlines express this as a *commands table*, not a positional argument. More
importantly `ch2parity` requires Chapter 2's CLI — stdin protocol, `render
<log>`, `dump` — to still work, so Chapter 3 cannot replace it with a positional
transcript argument. I translated the intent: Chapter 3 adds **no new CLI mode**
at all, and keeps `CH02_LOG` as the log variable so Chapter 2's harness runs
against the Chapter 3 binary unmodified. Please replace that line with the
commands table, or delete it.

**2. `go run ./testdata/exit7` cannot grade exit-code reporting. Two measured
reasons, and the second one surprised me.**

- `go run` prints `exit status 7` to **its own stderr**. Any tool that captures
  stderr therefore has the string "exit … 7" in its output *whether or not it
  reports exit codes at all*. An assertion on that fixture passes vacuously.
- `go run` itself **exits 1, not 7**. The exit code the agent actually receives
  from the outline's named scenario is 1. The table implies it is 7.

So the scenario as written grades nothing, and its name misdescribes it. I kept
`go run ./testdata/exit7` for fidelity (it asserts only that the call ran and was
*not* marked a tool error), and added a fourth command, a bare `exit 7`, which is
silent and really does exit 7. That is the one the exit-code points hang on. The
outline's command table should gain that row, with the reason.

**3. The chapter cannot pass `ch2parity` without touching Chapter 2's fixture,
and the outline does not anticipate that.**

`ch2Replies` ends on a reply containing a tool call, deliberately, because
Chapter 2 records tool calls and never executes them. Chapter 3 *does* execute
them — so it answers that call and asks for another turn, the fake repeats its
last reply (another tool call), and the agent spins to its round limit. Chapter
2's cumulative usage total then does not match and `usage` fails.

Fix applied, and I believe it is the only one that does not move Chapter 2's
graded behavior: I appended a **third reply that Chapter 2 never reaches** (it
sends two prompts, so it makes two requests) reporting **zero usage**, so the
session total is byte-for-byte identical whether or not tools were executed.
Chapter 2 still scores 100 and its token accounting is still fully graded under
`ch2parity`. The outline should say out loud that the shared fake has to gain a
terminating reply, because "run Chapter 2's checks unchanged" is otherwise not
achievable and the next person will rediscover this the hard way.

**4. §3.2's results-first rule was ungradeable by any fixture in either
chapter. This was a 100 → 100 row.**

The outline lists as a thing the chapter must say plainly:

> Anthropic requires the `tool_result` to come **first** in the content array of
> the message answering it.

I turned the splice off in the reference (`appendBlocks("user", blocks, true)` →
`false`) and **scored 100/100** — neither Chapter 2's checks nor my first draft
of Chapter 3's caught it. The reason is structural: the rule is only observable
in a message carrying a `tool_result` **and something else**, and no scripted
session in this chapter can build one, because a prompt cannot arrive while the
loop is blocked on a tool. That is Chapter 5's mailbox.

Chapter 2's `ExhibitLog` looked like it should catch it and does not: its human
turn comes *after* the tool return, so the result is already first and the splice
flag changes nothing. The discriminating case needs the human turn to land
*while the call is still outstanding*.

Fixed by adding `Ch3OrderingLog` — a log fixture with exactly that ordering — and
grading a `render` of it inside `toolloop`. The mutant now drops the score to
75/100. This is the single most valuable thing the audit found.

---

## Enrichment

**The chapter never says the agent must declare its tools to the vendor.** §3.2
walks parse → record → execute → return → loop, and none of those steps is "tell
the model what tools exist". Chapter 2's three renderers send no tool schema, and
Chapter 3 works against the fake regardless, because the fake volunteers
`tool_use` blocks unprompted. Against a real vendor an agent that never declares
its tools never receives a `tool_use` block at all, so the student's chapter-3
agent would work perfectly against the grader and do nothing whatsoever against
Anthropic. Either the chapter should add the declaration, or it should say
explicitly that it is deferred and why. I did not add it (see *Do NOT add*).

**"Malformed arguments" cannot mean malformed JSON.** §3.7 asks for "a tool call
with malformed arguments". A vendor cannot deliver syntactically invalid JSON —
the API would be emitting an invalid response about itself. In practice the
failure is arguments of the wrong *type* (`{"path": 42}`), or missing required
fields. Worth one sentence, because a student who tries to write the invalid-JSON
fixture will find they cannot put it on the wire.

**A non-zero exit is not a tool error, and this deserves a line of prose.** It is
the distinction that keeps the 15 `toolerror` points honest: the command *ran*,
and "the tests failed" is the answer, not a broken call. Marking it as an error
tells the model its call was malformed, which is false and invites it to "fix"
a call that was correct. I made it the negative control for `toolerror`, and it
is load-bearing — the mutant that marks non-zero exits as errors costs 20 points.

**The loop needs a round bound, and the chapter should say it is not a timeout.**
Without one, a model that keeps asking (or a fake that repeats its last reply)
loops forever. I used `MaxToolRounds = 16`. It is a *bound*, not cancellation —
nothing is interrupted, nothing runs concurrently — so it does not trespass on
Chapter 4. One sentence would stop a student reaching for a context deadline
here, which is the exact wrong lesson at this point in the book.

**`edit_file` has a second declined sub-decision the outline does not mention:
what to do when the anchor matches more than once.** §3.5 only asks about zero
matches. Ambiguity is the same question wearing a different hat, and a student
who refuses on zero and silently takes the first of three has not actually made
the decision. My reference refuses both, for one reason: an anchor matching three
places does not identify an edit site. Worth adding to §3.5 as part of the same
choice.

**The grader materializes `testdata/` itself.** `go run ./testdata/exit7` needs a
`go.mod`, so each scenario runs in a throwaway workspace the harness builds
(go.mod, `testdata/exit7`, `testdata/noisy`, `notes.md`, `contract.txt`). Grading
is hermetic and the student does not need those files in their own tree. The
outline's command table reads as though the student must create them.

---

## Do NOT add

- **Job handles, goroutines, timeouts, cancellation, `send_input`,
  `wait_for_job`, `kill_job`.** The brief is right that the urge shows up, and it
  shows up in exactly one place: `run_command` blocking on `go run`. I left it
  blocking. See question 3 below — nothing in Chapter 4 will have to tear out
  what I wrote.
- **A tool schema on the wire.** Tempting, and arguably an honesty gap (above),
  but adding a `tools` field changes the request bytes for all three vendors and
  Chapter 2 grades request bytes. The risk of regressing a 100/100 I was told to
  protect outweighed a fix to a problem the outline never asked me to solve. It
  is a decision for you, not for me, which is why it is in *Enrichment* as well.
- **Path confinement / sandboxing.** `read_file` will happily read `/etc/passwd`
  and `run_command` will happily `rm -rf`. That is Chapter 8's spine and planting
  the fix here would defuse it.
- **Parallel tool execution.** §3.2 calls sequential execution a deliberate
  choice; I kept it and the ordering is what makes `results-keyed-by-position`
  detectable.
- **Line numbers in `read_file` output.** My real agent does this and I wanted
  it. It is invisible to every check, it would have made the range assertions
  grade my formatting rather than the behavior, and the chapter never asks for it.
- **Fuzzy matching in `edit_file`.** The reference refuses. The point of §3.5 is
  that the student chooses, so the reference has to choose too, visibly, and
  then the grader has to not care.
- **Retry/backoff on tool failure.** A failed tool comes back to the model, which
  is the recovery mechanism. An automatic retry would hide the failure the
  chapter is trying to teach.

---

## The declined decision: the grader really does accept all three

This is the claim I most expected to get wrong, so it is tested directly rather
than argued. Three mutants replace the reference's refusal with the other two
answers and with two incoherent ones:

| mutant | score | failing |
|---|---|---|
| `editfile-answer-fuzzy-match` | **100/100** | — |
| `editfile-answer-rewrite` | **100/100** | — |
| `editfile-claims-success-changes-nothing` | 95/100 | `editcontract` |
| `editfile-refusal-says-nothing` | 95/100 | `editcontract` |

The test asserts the first two score *exactly* 100 — an accepted variant has to
be worth as much as the reference answer, or a preference has leaked in.

What `editcontract` actually grades is coherence between the report and reality:
the call is answered on the wire, the log records it, and either it reported
failure **and** left the file alone, or it reported success **and** the bytes
really moved. The only combinations rejected are the ones no answer defends. The
"says what it saw" requirement from §3.5 is applied **only to the refusing
branch**, where it is the right standard; success branches are held to the
symmetric standard of having actually changed the file. So refusing, fuzzy
matching and rewriting are each judged against their own contract.

---

## Facts verified vs not

**Verified against the artifacts:**

- The exhibit has **51 rows summing to exactly 70,401** calls, matching §3.0.
- **Six tools = 91.98%** (64,748/70,401). §3.3's "92.0%" is correct. The brief's
  "91.97%" is a truncation of 91.977%; 91.98% is the rounded value.
- **Nine tools = 93.88%** (66,095/70,401). §3.3's "93.9%" is correct.
- Every per-tool share in §3.0 and §3.3 matches the exhibit: `run_command`
  36.3%, `read_file` 25.7%, `edit_file` 16.6%, `search_files` 10.3%,
  `write_file` 2.3%, `list_directory` 0.75%, `send_input` 1.14%,
  `wait_for_job` 0.46%, `kill_job` 0.32%.
- `edit_file` 11,671 vs `replace_lines` 385 = **30.3:1**. "Thirty to one" is right.
- `edit_file` outnumbers `write_file` 11,671 to 1,616 = **7.2:1**. "Seven to one"
  is right.
- `go run` on a program calling `os.Exit(7)` exits **1** and prints `exit status
  7` to stderr. Measured directly.

**Wrong in the outline:**

- §3.1: *"Twenty-eight tools were called exactly once."* **Measured: 10.** The
  distribution of low counts in the exhibit is 1×10, 2×1, 3×2, 6×1, 7×3, 9×1,
  10×1. I also checked every threshold up to 200 and **no** "called at most N
  times" cut yields 28 tools, so it is not a mislabelled threshold either. Either
  the sentence or the exhibit is from a different corpus snapshot.
- §3.1: *"Top 5 is 90%. Top 13 is 97%."* Cumulative share is **91.22%** at row 5
  and **97.97%** at row 13. Both are rounded the wrong way, and §3.0 already says
  "ninety-one percent" for the top five, so §3.1 contradicts §3.0 by a point.
  Suggest "Top 5 is 91%. Top 13 is 98%."

**Not verified:**

- The 70,401 figure itself, and the `grep | sed | sort | uniq -c` one-liner
  against 511 session files. I checked the exhibit's internal arithmetic, not the
  corpus it was derived from — I do not have the histories directory.
- Every wire-format claim about real vendors. I tested against
  `internal/fakevendor`, which is a *model* of three vendors, and the package's
  own header is right that a fake is wrong exactly where you did not think to
  model it. Specifically, "Anthropic requires the `tool_result` first" is
  enforced by my grader but not by my fake — nothing in the test rig would
  reject a results-last request the way the real API would.
- The claim that ten tools postdate the measurement. Not checkable from here.

---

## The three questions

**1. Is `ch2parity` at 10 points right?** Yes, and for a reason I did not expect:
it fires far more often than a regression guard usually does. Four separate
mutants take it down — both type-filter mutants, `text-blocks-dropped`, and the
usage mutant — because Chapter 2's checks re-grade the parser that Chapter 3
extends. It is not a decoration. Ten is enough: a student who breaks the seam
loses those 10 *and* is almost certainly failing `toolloop`'s 25 at the same
time, because the two share a parser. Raising it would double-charge the same
bug; lowering it would let `usage`-style breakage (which Chapter 3's own checks
cannot see — the `ch2-usage-accounting-broken` mutant costs exactly 10 and
nothing else) go under the radar. Leave it.

**2. Does `localtools` at 20 cover five tools adequately?** Not quite, and your
instinct is right. I award 4 points per tool, so stubbing two costs 8 and a
student who ships three of five still scores **92/100**. Measured: stubbing
`list_directory` alone is 96/100, `search_files` alone 96/100. That is a
comfortable pass for an agent that cannot grep.

What I would do instead, and did not because it changes the outline's published
table: keep 20 points but make them **all-or-nothing across the five**, or split
`localtools` into `readtools` (read_file, list_directory, search_files — the
read-only set, which is also the set that makes the Chapter 8 permission
argument) and `mutatetools` (write_file, edit_file) at 10 each. The second is my
preference: it maps onto a real distinction the book already makes, and it keeps
partial credit where a student genuinely can have one skill and not the other.
Your stated rule — "splitting points is for when a student can plausibly have one
skill and not the other" — actually argues *for* that split, because reading and
mutating are separable skills in a way that `list_directory` and `search_files`
are not.

**3. Did anything force code Chapter 4 will tear out?** No. The boundary is in
the right place, and it is cleaner than I expected.

`Execute` calls `Dispatch(name, args)` and gets `(string, error)` back. Chapter 4
reworks `run_command` into a supervised job by changing **one entry in the
registry map** and nothing else: the executor, the loop, the log vocabulary, the
result plumbing and the other five tools are all untouched. The five local tools
never grow a handle because they never needed one.

The one thing Chapter 4 will have to *add* rather than change: `Execute` records
`ToolCalled` then `ToolReturned` back to back, because a blocking tool has no
observable middle. A job has one, so Chapter 4 will put events between them.
That is an extension of the log, not a tear-out — and `BlobPart.Path`, which
Chapter 2 built and Chapter 3 still does not exercise, is sitting there waiting
for the moment output arrives by the megabyte. Chapter 4's collection column in
the outline is accurate.

---

## The audit, in full

22 mutants, each asserting the **exact** set of failing check ids. A mutation
that does not apply is a hard test failure (`MUTATION DID NOT LAND`), because a
silently-unapplied mutation scores 100 and manufactures a fake finding.

| mutant | score | failing checks |
|---|---|---|
| `blind-walk-no-type-filter` | 0 | all seven |
| `tool-use-blocks-ignored` | 0 | all seven |
| `text-blocks-dropped` | 40 | ch2parity multiblock toolerror toolloop |
| `loop-runs-one-round` | 14 | editcontract localtools multiblock runcommand toolerror toolloop |
| `read-range-ignored` | 61 | localtools multiblock toolloop |
| `tool-results-not-spliced-first` | 75 | toolloop |
| `exit-code-not-reported` | 80 | runcommand toolerror |
| `nonzero-exit-marked-as-error` | 80 | runcommand toolerror |
| `tool-error-not-marked` | 80 | editcontract toolerror |
| `tool-error-result-dropped` | 80 | editcontract toolerror |
| `unknown-tool-panics` | 85 | toolerror |
| `write-file-writes-nothing` | 88 | localtools |
| `ch2-usage-accounting-broken` | 90 | ch2parity |
| `results-keyed-by-position` | 90 | multiblock |
| `edit-file-does-not-change-the-file` | 92 | localtools |
| `stderr-dropped` | 95 | runcommand |
| `editfile-claims-success-changes-nothing` | 95 | editcontract |
| `editfile-refusal-says-nothing` | 95 | editcontract |
| `list-directory-hides-files` | 96 | localtools |
| `search-files-finds-nothing` | 96 | localtools |
| `editfile-answer-fuzzy-match` | **100** | — *(accepted variant)* |
| `editfile-answer-rewrite` | **100** | — *(accepted variant)* |

**Is Chapter 1's deferred `type` filter load-bearing?** Yes, and it is the most
load-bearing thing in the chapter. Chapter 3's wire returns an assistant message
whose content array is `[{type:"text"}, {type:"tool_use"}]`, and the
`multiblock` fixture returns `[text, tool_use, tool_use]`. Deleting the filter —
`switch h.Type` → `switch "text"`, which is precisely "walk every block and
concatenate whatever you find" — scores **0/100**, failing all seven checks. The
complementary mutant, ignoring text blocks instead, costs 60. Both halves of the
dispatch are required, which is the property Chapter 1's hole lacked: with one
block, walking and indexing were the same program. Here they are not.

**Vacuity questions, per check.** The two that had real answers:

- `runcommand`'s exit-code assertion *was* vacuous in my first draft, satisfied
  by `go run`'s own stderr. Fixed with the silent `exit 7`, plus a negative
  control asserting a *successful* command does **not** report exit 7 — without
  that, printing "exit 7" unconditionally would have scored the points.
- `toolloop`'s results-first assertion was vacuous in my first draft (every
  message answering a call contained only the result). Fixed with
  `Ch3OrderingLog`, as above.

`localtools`' absence assertions ("the range was ignored and the whole file came
back") have a live negative control: the `multiblock` fixture legitimately
returns `alpha` from a one-line range read, so a checker that could not see
`alpha` at all would fail there rather than pass quietly here.
