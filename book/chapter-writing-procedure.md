# Chapter Writing Procedure

How to write a chapter for *The Self-Wielding Agent*.
Follow these steps in order. Each step has a rationale.

## Roles

- **Author**: Opus 4.6 (preferred over Fable 5.1 for prose). Writes
  outlines, chapter text, voice checks, briefs. Never edits
  `solutions/` or `internal/`.
- **Coder**: Opus 5 (preferred over Fable 5.1 for code). Writes
  the reference solution, grader, fixtures, mutation tests, and
  a post-build review for the author. Never edits `book/`.
- **Bill**: Reviews outlines, rules on open questions, LGTMs
  chapters. Push is always Bill's. **Bill never reports bugs
  directly.** See below.

### Bill never reports bugs

If Bill sees a bug in the running agent, he does NOT report it.
Reporting a bug means the fix happens outside the grader, which
means the next LLM building from the book ships the same bug.
The grader IS the specification. If it does not test for something,
that something does not exist.

Instead, Bill suggests improvements to the **TL;DR** — what the
grader should check for, what behavior the exercise contract
requires. The author updates the TL;DR accordingly. Then the
**coder re-runs the chapter**: rebuilds from the updated brief,
re-grades, and either the grader already catches the concern
(good — the spec was already strong enough) or the grader needs
a new check (add it, then fix the code to pass). Either way, the
fix flows through the grader, and every future agent built from
the book inherits it.

This is the difference between a bug fix and a specification
improvement. Bug fixes are local. Specification improvements
propagate through regeneration.

### The grader feedback loop

When Bill raises a concern about behavior, the author and coder
work it through a three-agent loop:

1. **Author** updates the TL;DR and plain-English sections to
   describe the required behavior clearly enough that a fresh
   coder could implement it from the chapter text alone.

2. **Grader coder** (Opus 5 sub-agent) reads the updated TL;DR
   and enhances the grader: adds or strengthens checks, updates
   fixtures, creates mutation tests. The grader coder never reads
   the reference solution — it builds from the spec.

3. **Student coder** (Opus 5 sub-agent) reads ONLY the chapter
   (TL;DR + prose) and attempts to build a passing solution from
   scratch, starting from `solutions/ch(N-1)`. If it struggles
   or fails, that is a signal: the TL;DR has a gap. The author
   fixes the gap and the student retries.

The loop continues until the student passes 100/100 from the
chapter text alone. At that point, the grader spec, the chapter
prose, and the reference solution are all consistent — and the
next LLM that reads the chapter will produce a working agent.

This loop belongs in Phase 2 (after the initial draft) and in
the review cycle (Phase 3) whenever Bill raises a new concern.

## The full arc

1. **Free-form outline** — Bill and the author create the outline
   together, iterating until Bill is satisfied. This is where
   the chapter's thesis, section structure, grading table, and
   exercise contract take shape. The outline lives at
   `book/chapter-NN-outline.md` (or `-outline-new.md`).

2. **Coder builds the solution** — The coder (Opus 5) receives a
   brief derived from the outline and builds `solutions/chNN/`,
   `internal/grade/chNN_*` (grader, fixtures, harness, mutation
   tests). The coder may work with Bill on design questions.
   Ends with `make gradeN` passing 100/100.

3. **Coder writes a review** — The coder writes a post-build
   review documenting what was built, deviations from the
   outline, figures measured, and issues found. This review
   goes to the author.

4. **Author updates the outline** — The author (Opus 4.6)
   incorporates the coder's review: corrects stale figures,
   updates the grading table to match the shipped grader,
   resolves any deviations. This produces the final outline
   that the prose is drafted from.

5. **Author follows the chapter-writing procedure** (Phase 2
   below) to convert the outline into the published chapter:
   motivational intro, TL;DR, §3.1 test, plain-English section,
   body, exercises, "Taking it for a spin", voice checks, grade
   check, commit.

6. **Bill reviews** (Phase 3 below).

## Prerequisites

Before starting any chapter N:

1. **Chapters 1 through N-1 are committed and LGTM'd by Bill.**
2. **The grader and solution exist** (`solutions/chNN/`, `internal/grade/chNN_*`).
   The grader is written by the CODER (Opus 5), not the author.
3. **`make gradeN` passes 100/100** on the reference solution.

## Phase 1: Derive the Outline

### 1a. Read the grader fixtures first

The grader's fixture logs (`internal/grade/chNN_log.go`) define
the on-disk format the student's code must produce. **The fixture
IS the contract.** Every chapter N TL;DR must be consistent with
these fixtures — if the fixture says `"type": "tool_returned"`,
the page cannot say `"kind": "tool_result"`.

```
read internal/grade/chNN_checks.go
read internal/grade/chNN_harness.go (if it exists)
read internal/grade/chNN_log.go (if it exists)
read internal/grade/chNN_grader_test.go
```

### 1b. Read the solution code

```
list_directory solutions/chNN -recursive
```

Read every `.go` file. Note the types, their JSON tags, the
exported API, and any doc comments (doc comments may contain
the chapter's easter egg).

### 1c. Write the derive fact sheet

Create `book/derive-chNN.md` with sections:

- **A. Types** — every exported type, field, tag
- **B. Grader map** — check name → points → what it tests
- **C. Stale text** — anything in the old outline that
  contradicts the code
- **D. Figures** — measured numbers (line counts, byte counts,
  mutant counts)
- **E. Easter eggs** — where the code egg lives
- **F. Coder list** — issues found for the coder to fix later
- **G. Open questions** — things needing Bill's ruling

### 1d. Write the new outline

Create `book/chapter-NN-outline-new.md`. Structure:

1. Title
2. TL;DR spec (types, rules, exercise table)
3. §N.1 "The idea in plain words"
4. Body sections (one per concept, each with a thesis sentence)
5. §N.last "Exercise, graded"
6. Checks table (check → points → section that teaches it)
7. Coder list

**Get Bill's LGTM on the outline before writing prose.**

## Phase 2: Write the Chapter

### 2a. Load context

Load these files, calling `keep_tool_results` after each read
so they stay in context:

1. `book/voice.md` — the voice spec (binding)
2. `book/chapter-NN-outline-new.md` — the outline
3. The previous chapter (structural model)
4. The grader checks and fixtures

### 2b. Write the motivational intro

Every chapter opens with a short motivational paragraph before the
TL;DR. This is the one place in the chapter where emotionally charged
claims are permitted: bold promises, strong opinions about the state
of the industry, direct statements about why this chapter matters.
The voice.md register of dry receipts does not apply here. The intro
concentrates the heat so the rest of the chapter can stay cool.

Guidelines:
- Two to four sentences. Sell the chapter, then get out.
- Permitted: "this is where every coding agent fails", "the most
  valuable chapter in the book for X", "blow your mind."
- Not permitted: hedging, false modesty, or caveats. If the claim
  is worth making, make it flat.
- The intro is NOT part of the TL;DR. It precedes it. The TL;DR
  remains the contract the grader tests.

Chapters 1-4 already have motivational intros added retroactively.
Chapter 5+ should include them from the outline stage.

### 2c. Write the voice plan

Before drafting, write a short voice plan at the top of
your working notes (not in the chapter file):

- **Stake**: what the narrator has at risk in this chapter
- **Register**: which of manual/Lewis/Spolsky dominates
- **Bill moment**: which aside names Bill (≤2 per chapter)
- **Confession inventory**: which mistakes to confess (≤3)

### 2d. Draft the TL;DR

The TL;DR is the chapter's most important section. It must
pass the **§3.1 test**: a fresh coder given ONLY the TL;DR
pages from chapters 1 through N must be able to build a
solution that scores 100/100 on the grader.

TL;DR structure:
- ≤700 prose words
- Types printed verbatim (Go code block)
- Numbered rules (the contract)
- "Yours" paragraph (taste decisions left to student)
- Exercise paragraph with the grader command

**Print the grader fixture log verbatim if it defines
the on-disk format.** The student needs the exact bytes.

### 2e. Run the §3.1 TL;DR test

Spawn a fresh claude-opus-5 coder (async) with:

- A copy of `solutions/ch(N-1)` as the starting point
  (the student can build on prior work)
- The TL;DR pages from all chapters, extracted to a
  clean directory outside `book/`, `solutions/`, `internal/`
- The grader command (`make grade-dir CH=N DIR=path`)
- **Fence it from** `solutions/`, `book/`, `internal/`, `cmd/`

If it fails: fix the page gaps, re-test.
If it scores 100/100: the TL;DR is dense enough.

### 2f. Write the body

Follow the outline section by section. For each section:

1. Open with the thesis (what this section proves)
2. Show the mechanism (code, exhibit, or receipt)
3. Close on the idea, not a label

**§N.1 must be "The idea in plain words"** — a plain-English
explanation of the chapter's core concept for a smart
non-specialist, BEFORE any technical detail. Use an everyday
analogy if one fits. Bold-led paragraphs, one idea each,
why before what.

**voice.md v4 hard rules** (zero tolerance):
- 0 em-dashes
- 0 first person outside colophon
- 0 "not just"
- 0 previews ("in the next chapter", "we will see")
- 0 throat-clearing ("It is worth noting", "Interestingly")
- 0 LLM crutches ("delve", "facilitate", "leverage" as verb)
- Asides ≤10%, labeled blockquotes, skippable
- Bill ≤2 mentions, only inside asides
- Lines ≤82 columns outside code/tables

### 2g. Voice check

Run these checks before committing:

```bash
# Em-dashes
grep -c '—' book/chapter-NN.md

# First person
grep -in '\bI\b' book/chapter-NN.md | grep -v '^\s*>' | grep -v '```'

# "not just"
grep -ci 'not just' book/chapter-NN.md

# Previews
grep -in 'next chapter\|we will see\|later chapter\|in this chapter' book/chapter-NN.md

# Throat-clearing
grep -in 'worth noting\|interestingly\|it bears\|notably' book/chapter-NN.md

# Bill count
grep -c 'Bill' book/chapter-NN.md

# Word count
wc -w < book/chapter-NN.md

# Aside percentage
# (count lines starting with > and divide by total)
```

### 2h. Grade check

```bash
make gradeN    # must be 100/100
```

### 2i. Write the "Taking it for a spin" section

Every chapter ends with an expanded section showing the solution
in action. This is the fun part: CLI interaction transcripts,
code blocks showing real agent behavior, and demonstrations that
make the reader want to try it themselves.

Guidelines:
- Show real CLI interaction with the reference solution: commands
  typed, output produced, the agent doing something interesting.
- Starting with Chapter 5, include screenshots. Every screenshot
  must have a clear, accessible text description immediately
  below it so a reader using a screen reader gets the full
  picture. This is not optional.
- The voice rules relax here. Enthusiasm is permitted. The point
  is to show that what the reader just built is genuinely fun to
  use.
- Keep it short enough that it feels like a reward, not a second
  chapter. One or two interactions, the best ones.

### 2j. Commit

```bash
git add book/chapter-NN.md book/voice.md  # if voice.md changed
git commit -m "ch4: first draft — Jobs"
```

**Never `git add -A`.** Never stage `preface.md` (Bill's
uncommitted edits). Push is Bill's.

## Phase 3: Review

1. Bill reads the chapter.
2. Apply edits from Bill's feedback.
3. Re-run voice checks and grader after each edit pass.
4. Commit when Bill says LGTM.

## Lessons Learned (accumulated ch1–ch4)

### Don't ask Bill to rubber-stamp your decisions
If you already have a strong case for the answer, write the answer.
Do not append questions at the end of a section to give Bill the
feeling he had a say. Bill's time is the scarcest resource. Bring
him real open questions where his judgment changes the outcome, not
questions you have already resolved in your own reasoning. When in
doubt, make the call, document it, and let Bill override if he
disagrees.

### The fixture IS the contract
For any chapter, read the grader's fixture logs before
writing the TL;DR. The fixture defines the on-disk format.
If the outline disagrees with the fixture, the fixture wins.

### Plain English first
The first body section after the TL;DR must be the plain-
English WHY for a smart non-specialist. This was added to
voice.md after ch2's review (Bill: "doesn't explain why we're
doing things in plain English").

### Figures from artifacts, never memory
Re-measure every number (`wc -l`, `wc -c`, `grep -c`) before
printing it. Across four chapters, 9+ stale figures were caught
and corrected by re-measurement.

### Code facts from grep, never memory
Never state a CodeRhapsody architecture fact from memory.
Grep first. Two false claims were caught in ch1 by grep.

### edit_file rewraps lose words
After every edit_file that rewraps a paragraph, re-read
the paragraph. Twice in ch1, a word at a line boundary
was dropped ("types", "starts").

### §3.1 test catches real page gaps
The test is cheap (~25 min per chapter) and has found
real defects in every chapter so far. Do it every time.

### Starting from prior solutions (Bill's ruling, ch4+)
For the §3.1 TL;DR test, the coder starts from a copy of
`solutions/ch(N-1)` rather than rebuilding from scratch.
This reflects the real student experience and saves time.
