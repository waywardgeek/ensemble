# Brief for the coder — Chapter 3 solution and grader

**From:** the author. **Date:** 2026-09-13.
**Spec:** `book/chapter-03-outline.md`. Read it in full first; this brief covers
only what the outline cannot say for itself.

---

## What Chapter 3 is

**"Six Tools: Ninety-Two Percent of an AI Coding Agent."**

The student's agent stops being a chat loop and starts being able to work. It
gains the tool loop and the **easy, blocking form** of six tools:

`run_command`, `read_file`, `edit_file`, `write_file`, `list_directory`,
`search_files`

That set is 91.97% of 70,401 measured tool calls. The exhibit with the full
distribution is `book/exhibit-ch03-tools.md`; regenerate with
`./scripts/tool-usage.sh <histories-dir>` if you ever need to.

**Blocking is deliberate and is not a defect to fix.** Chapter 4 reworks exactly
one of these six into a supervised job, and the student is supposed to feel why
first. Do not build job handles, timeouts, goroutines or cancellation. If you
find yourself wanting to, that is Chapter 4 working as designed.

**Not in this chapter:** `send_input`, `wait_for_job`, `kill_job` (all Chapter
4), the mailbox and hints (Chapter 5), sub-agents, skills, MCP.

---

## Deliverables

1. `solutions/ch03/` — the reference solution, built from ch02.
2. `internal/grade/ch03_*.go` — seven checks, ids and point values exactly as in
   the outline's grading table. **They must sum to exactly 100.**
3. Mutation tests in the established style, each asserting the **exact set** of
   failing check ids.
4. The fake vendor server extended to serve Chapter 3's scenarios.
5. `book/review-ch03-code.md` — your review, described at the end of this brief.

---

## Hard constraints, in descending order of how much damage getting them wrong does

### 1. The fake is shared with Chapter 2. Do not regress it.

`internal/fakevendor` serves every chapter. Chapter 3 needs new endpoints or new
behavior; Chapter 2's graded behavior must not move. **`go run ./cmd/grade -ch 2
solutions/ch02` must still score 100/100 when you are done**, and I will check
that independently. Same for `-ch 1`.

### 2. Chapter 1 left a promise here, and Chapter 3 is where it is collected.

Chapter 1 §1.2 now tells the student, in the prose, that filtering response
content blocks by `type` starts paying "in Chapter 3, the first time a model asks
to call a tool." That promise was made deliberately, to convert a grader hole
into a chapter dependency.

**So: Chapter 3's wire must return at least one assistant message whose content
array contains both a text block and a `tool_use` block.** An implementation that
walks every block and concatenates whatever it finds must produce visibly wrong
behavior — garbage in the transcript, or a tool name treated as prose. An
implementation that filters correctly must pass. If the filter is not
load-bearing at the end of your work, the promise is uncollected and Chapter 1 is
left lying to the reader.

### 3. The property every check must have (P9)

`book/course-policy.md` P9. The property, stated as a property so you can pick
your own mechanism:

> **For every behavior a check claims to protect, there must exist a change to
> the REFERENCE SOLUTION that removes that behavior and causes the score to
> drop.** A check that cannot fail is not a check.

Two corollaries that have each already caught a real hole in this repo:

- **Mutating a deliberately-broken student proves your checks fire. Deleting
  behavior from the reference proves the chapter's promises are collected.**
  These are different tests and only the second found the Chapter 2 `Opaque`
  hole and the Chapter 1 single-block hole. Do both.
- **An absence assertion needs a negative control.** If a check asserts that
  something does *not* appear, write the case where it legitimately *does* and
  confirm the check stays quiet. Otherwise `assert !contains(x)` passes happily
  on an empty string.

### 4. Assert that each mutation actually landed

I hit this myself this morning, on my own code, in this repo. My first mutation
silently failed to apply — the pattern did not match — and the full test suite
passed. That looks exactly like a suite proving the feature is well covered. It
is a suite proving nothing.

**Property:** a mutation that fails to apply must fail loudly rather than
producing a green run. How you get there is yours; the existing ch01/ch02 rigs
already have a shape for it (`mutating()` + an unknown-mutation guard) and
extending that is probably cheapest.

### 5. Tool errors are Chapter 3's most valuable 15 points

`toolerror` is weighted at 15 deliberately. A tool that fails — file not found,
command exits non-zero, bad arguments — must return a `tool_result` **to the
model**, marked as an error, so the model can recover. It must not panic, must
not silently drop the result, and must not end the turn.

This is the single most common way a student's agent locks up, and it is the
place where a passing-looking implementation is most likely to be wrong. Grade it
hard. At minimum: a tool that does not exist, a tool whose arguments do not
parse, and a tool whose execution fails.

---

## Things the outline deliberately does not decide

**The declined decision (P6) is `edit_file`'s failure contract.** When the
anchor text does not match the file, the student may refuse, fuzzy-match, or
rewrite. **All three are acceptable and the grader must accept all three.**

What the grader checks is not which one was chosen, but that the choice was made
coherently: the failure is reported back to the model in a form it can act on,
and the event log makes the outcome legible. Do not let a preference for one
answer leak into the checks. If you cannot write a check that accepts all three,
say so in the review rather than quietly narrowing it.

This is genuine, not a riddle with a hidden answer: CodeRhapsody ships **both**
answers in the same binary. `edit_file` refuses on exact-match failure;
`replace_lines` deliberately fuzzy-searches within 50 lines.

---

## Traps in this repo, so you do not rediscover them

- `go run ./cmd/grade -ch N solutions/chNN` — the directory is **positional**.
- `ch02.log` and `internal/grade/ch02.log` are **tracked** and churn on every
  grader run. `git checkout --` them before committing.
- **Never `git add -A`.** `book/tic-apply-report.md` is the author's untracked
  work in progress, and the author may be editing other files while you work.
  Stage by filename.
- Sum the points with awk **scoped to the checks table**, not file-wide. Other
  tables in these outlines have numbers in the second column. My own file-wide
  check returned the right answer by luck this morning, because check ids happen
  to contain no underscores and tool names do.
- `grep -c` exits 1 when it finds nothing, and `diff` exits 1 when it finds
  differences. Both silently break an `&&` chain, and the command after them
  never runs.
- `gofmt -l .` exits 0 while listing files. Check the output, not the status.

---

## The review I want back: `book/review-ch03-code.md`

Write it as an implementer who has just been forced to make every decision the
outline left implicit. I am looking for the places where the chapter is
**underspecified**, not for agreement.

Use the established format:

- **Must-fix** — where the outline is wrong, contradictory, or cannot be built
  as written.
- **Enrichment** — what the chapter should say that it does not, that you only
  learned by implementing it.
- **Do NOT add** — what you were tempted to build and deliberately did not, with
  the reason. This section has been the most valuable one historically; it is
  where scope creep goes to die on the record.
- **Facts verified vs not** — anything the outline asserts that you checked, and
  anything you could not check. If a number in the outline is wrong, say so
  plainly. Five of my own numbers were wrong in this book before I checked them
  against the artifacts, so assume mine are suspect rather than authoritative.

Three specific questions I want answered, because I could not settle them from
the author's chair:

1. **Is `ch2parity` at 10 points right?** It guards that Chapter 2's `render`
   and `parse` still behave after the tool loop lands. Too low and a student can
   break the seam and still pass comfortably; too high and it crowds out the new
   material.
2. **Does `localtools` at 20 points cover five tools adequately**, or does the
   spread make it possible to stub two of them and still score well? If so, say
   what you would do instead.
3. **Did anything in the chapter force you to write code that Chapter 4 will
   have to tear out?** Chapter 4 should rework `run_command` and leave the other
   five untouched. If that is not true of your implementation, the chapter
   boundary is in the wrong place and I need to know now, not after it ships.

---

## What "done" means

- `go run ./cmd/grade -ch 3 solutions/ch03` scores **100/100**.
- `-ch 1` and `-ch 2` both still score **100/100**.
- Every mutant is detected, each asserting its exact failing set.
- Deleting any protected behavior from the reference drops the score.
- `book/review-ch03-code.md` is written.
- `gofmt -l .` lists nothing.

Submit the structured result when all of that is true. If you cannot make one of
them true, submit anyway and say which one and why. **A report that names an
unmet criterion is worth more to me than a green dashboard**, and nothing bad
happens to you for reporting it.
