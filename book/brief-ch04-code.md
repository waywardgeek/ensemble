# Brief: Chapter 4 reference solution + grader

You are the CODER for chapter 4 of "Building Advanced AI Coding Agents", repo
`~/projects/coding-agents-course`. The human in this conversation is Bill Cox,
the book's editor and the expert on the agent this chapter is modeled on
(CodeRhapsody). The chapter's author (CodeRhapsody, a separate agent) is NOT in
this conversation. When the outline is ambiguous, ask Bill; do not guess. When
Bill states something flat, trust it; when he hedges, verify it.

## Ground truth, in priority order

1. `book/chapter-04-outline.md` — read ALL of it first. Where this brief and
   the outline disagree, the outline wins; note the disagreement in your report.
2. `book/course-policy.md` — P1 through P10. Especially P1 (additive means
   architecture, not files: editing prior code is fine), P3 (grade behavior,
   not lineage), P9 (every grader audited by deletion; every check carries
   points), P10 (model defaults).
3. `solutions/ch03/` — the base you are reworking. Read it whole before editing.
4. `internal/grade/ch03_*.go` and `internal/grade/ch02_*.go` — grader
   conventions: check structs with `ID`/`Points`, the `buildMutant` regexp
   table (ch2 convention; preferred over `COURSE_MUTATION` copies), mutation
   tests that assert the EXACT failing check-id set and `earned < max`.
5. `internal/fakevendor/` — one path-routed fake for all three vendors. It is
   SHARED by chapters 2 and 3. Any change here must leave `-ch 2` and `-ch 3`
   at 100/100.
6. `book/review-ch03-code.md` — how the previous coder reported, and the
   rulings the author made on it. Match the report shape.

## What you build

### `solutions/ch04/`

Start from `solutions/ch03`. Chapter 4's thesis: **a tool call is a job, for
every tool, decided at the dispatch site.** Not a special case for
`run_command`. The outline (§4.1–§4.7) specifies:

- A job model: handle allocated at dispatch for EVERY tool call, status,
  output spilled to a file on disk, result retrievable after the call returns.
- Each tool DECLARES its blocking contract (`DeadlineArg` / `MaxBlockingTime`
  per the outline's names). The dispatcher enforces what the tool declares. A
  tool that accepts a deadline argument and declares nothing is an error the
  dispatcher catches — that is the `send_secret` defect from the war story.
- Inline result cap with the full output on disk and a stub in context giving
  byte count and path; truncation happens at dispatch, not in the tool. Use the
  exact cap numbers from the outline.
- `wait_for_job` (must work on an already-finished job), `send_input`,
  `kill_job` (waiters are told). `ai_callback_pattern` as the outline describes
  it for `run_command`.
- The declined decision (`budgetchoice`): what happens when a supervised job
  never returns. You MUST make a choice from the outline's options, and it must
  be legible in the event log. Any coherent choice passes; illegibility fails.
- NOT in this chapter: `Interrupted` as a turn state, the mailbox, MCP tools.
  The agent is still deaf while a job runs. Do not fix that; chapter 5 does.

Contract: `./ch04 <transcript-file>` — no network, deterministic, same fake
vendor. `ch3parity` means `solutions/ch04` must still score 100 on the ch3
grader. Run it early and often.

### Fixtures (`testdata/`, Go helpers, portable — no shell tricks)

The outline §4.9 lists six: fast finish, outlives its budget, never returns,
reads stdin and echoes, emits > 1 MiB, accepts a deadline argument but declares
no contract. The "never returns" helper must actually never return — if it
exits on its own, `budgetchoice` and `killjob` are graded by nothing.

**Trap from ch3:** `go run ./helper` exits 1 when the helper exits 7, because
the `go run` wrapper rewrites the status. Build helpers to binaries in a temp
dir during grading, or exec them directly. Do not grade through `go run`.

### Grader (`internal/grade/ch04_*.go`)

Eight checks, ids and points EXACTLY as the outline's §4.9 table:
`ch3parity` 10, `jobmodel` 25, `waitjob` 10, `sendinput` 15, `killjob` 10,
`bigoutput` 15, `declare` 10, `budgetchoice` 5. Sum 100. Verify the sum with a
regex scoped to the checks table, not the whole file (the outline explains
why).

Grade BEHAVIOR (P3): drive the student binary with a transcript, read the event
log and the output files. Do not grep student source for identifiers, except
where the outline says a declaration is the thing being graded (`declare`), and
even then prefer observing the dispatcher's refusal to reading the struct.

## P9 audit — do this, and report the table

- For every check, a mutant of the REFERENCE that deletes the protected
  behavior. Extend the `buildMutant` table. Each mutation test asserts:
  (a) the mutation actually applied (byte-compare or count the replacement;
  an unapplied mutation scores 100 and manufactures a fake result — this
  project has been bitten three times), (b) `earned < max`, (c) the EXACT set
  of failing ids.
- `jobmodel` negative control: a tool OTHER than `run_command` must also get a
  handle. A student who special-cased the shell must fail this check.
- Watch for the guard-computed-by-its-subject shape: a test that iterates over
  what a scan found checks nothing when the scan shrinks. Enumerate the
  expected set by hand.
- A `t.Skip` on an empty set is a vacuous pass. Fail instead.
- `gofmt` realigns struct keys; text-pattern mutations can silently miss. Run
  `gofmt -l .` (it exits 0 while listing files — check the output, not the
  code) and re-verify mutations after formatting.

## Verification, before you report

```
go build ./...
go run ./cmd/grade -ch 2 solutions/ch02     # must stay 100/100
go run ./cmd/grade -ch 3 solutions/ch03     # must stay 100/100
go run ./cmd/grade -ch 3 solutions/ch04     # ch3parity: 100/100
go run ./cmd/grade -ch 4 solutions/ch04     # 100/100
go test -count=1 ./internal/grade/...       # all mutants detected
```

If you touched `internal/fakevendor`, byte-compare the ch2 and ch3 request
logs before/after (the ch3 brief's method).

## Working discipline

- Keep running notes in `book/.notes-ch04-code.md`: what you've verified, what
  you're looking for, what you decided. If you find yourself searching for
  something repeatedly without finding it, STOP, write down what and why, and
  ask Bill. Do not grind.
- Commit small as things go green. Add files BY NAME; never `git add -A`
  (there are untracked files in the tree that must stay untracked). Do not
  push; pushing is Bill's call.
- Do not edit `solutions/ch02`, `solutions/ch03`, or any `book/chapter-*.md`.
  If the outline needs a change, say so in the report; the author owns it.
- Do not edit `book/course-policy.md`.

## Deliverable

`book/review-ch04-code.md`, same shape as `review-ch03-code.md`:

1. What changed (files, line counts — re-derive, do not estimate).
2. Check → mechanism table: for each id, what the grader observes and which
   fixture drives it.
3. Mutation table: mutant → score → failing ids, with the "applied?" assertion
   named.
4. Where the brief and the outline disagreed, and which you followed.
5. Open questions for the author, numbered, each a falsifiable question.
6. Facts verified vs. assumed. Keep the two lists separate.

End your final message with a line beginning `DONE:` and a one-line summary.
