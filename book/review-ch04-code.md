# Review — Chapter 4 solution and grader

**From:** the coder. **Date:** 2026-09-13.
**Spec:** `book/chapter-04-outline.md`. **Brief:** `book/brief-ch04-code.md`.
**Editor rulings taken live from Bill during the build:** see §4; they
supersede outline §4.6 and §4.7 and change the checks table.

Built: `agent/` — the live tree, seeded from `solutions/ch03`, now the ch4
agent — snapshotted byte-for-byte to `solutions/ch04/` and tagged
`ch04-solution` (`ch03-solution` was tagged at the start, at HEAD as it stood).
Nine checks in `internal/grade/ch04_{harness,checks}.go`, a 22-mutant deletion
audit in `internal/grade/ch04_grader_test.go`, `-ch 4` in `cmd/grade`, and
chapter 4 in `scripts/live.sh`.

`-ch 2 solutions/ch02` 100/100, `-ch 3 solutions/ch03` 100/100,
`-ch 3 solutions/ch04` 100/100 (that is `ch3parity`), `-ch 4 solutions/ch04`
100/100. `internal/fakevendor` was not touched. Points sum to 100 by the test
that sums the checks themselves; the outline's table no longer matches the
code (§4), so the awk over the outline was not the right instrument this time.

The chapter's thesis survives contact, and in a stronger form than the outline
states it. Every tool becomes a job by changing **one function**
(`Engine.Execute`); five of the six tools changed by one ignored parameter.
And the closing demonstration turned out to be a *measured* capability
boundary rather than a rhetorical one: on pipes, `dlv` refuses to start —
`Stdin is not a terminal` — so chapter 3's agent could not drive it at any
speed. The grader drives it through the fake; `scripts/live.sh 4 gemini
rounds` drove it live on `gemini-3.8-flash` in five pattern-driven calls and
read `42` off the breakpoint.

---

## 1. What changed

Re-derived with `wc -l` and `diff`, not estimated.

**Solution.** `solutions/ch03` 3,844 lines → `solutions/ch04` 4,747 lines
(+903).

| file | change |
|---|---|
| `jobs.go` | new, 522 lines. `Job`, `Jobs`, `Limits`, `Wait`, `Kill`, `Report`, `capText`. |
| `jobtools.go` | new, 165 lines. `wait_for_job`, `send_input`, `kill_job`, `tool_limits`. |
| `engine.go` | 106 changed lines. `Execute` is the dispatch site; `Shutdown` added; `Jobs` on `Engine`. |
| `tools.go` | 195 changed lines. `ToolFunc` takes `*Call`; `Tool.NoJob`; `run_command` on a PTY streaming into the job; four registry entries; `Dispatch` → `Lookup`. |
| `event.go` | 31 changed lines. `JobKilled` event type; `ToolData.Job`, `Event.Job`; `JobData`. Stale taxonomy comment corrected ("Chapter 3 adds job events" → chapter 4 did; chapter 5 adds Interrupted). |
| `main.go` | 6 changed lines. `Shutdown()` at the end of `runLoop`; `fallbackName` ch04. |
| `tools_test.go` | 2 changed lines (signature). |
| `go.mod`, `go.sum` | `github.com/creack/pty v1.1.24` — the course module's first dependency. |

**Grader.** `ch04_harness.go` 573, `ch04_checks.go` 525, `ch04_grader_test.go`
436 lines; `cmd/grade/main.go` +9; `scripts/live.sh` +38 −2.

**Untouched, as required:** `solutions/ch02`, `solutions/ch03`,
`internal/fakevendor`, `book/chapter-*.md`, `book/course-policy.md`.

### The design as built (this is what the outline needs to say)

- **Every tool call is a job** except the four supervision tools. At
  `Execute`, before the tool runs: handle `N` (integers from 1, +1 per job,
  per process), file `cr/io/N`, status `running`. The tool runs on its own
  goroutine; the dispatcher waits **on the job**, not in the tool, until
  done / `ai_callback_delay` elapsed (default **3s**) / `ai_callback_pattern`
  matched against output the model has not yet seen — whichever first.
- **Nothing dies on its own.** No blocking contracts, no budgets, no policy.
  `kill_job` is the only terminator (SIGKILL to the process group; a stuck
  goroutine is marked killed and the report says Go cannot stop it).
  `Shutdown()` at exit kills every running job and records `job_killed`
  with reason `shutdown`.
- **Limits** = `{ai_callback_delay 3s, ai_callback_pattern none,
  max_output_bytes 16 KiB}`. Sources, later wins: defaults → pending
  `tool_limits` (one-shot; consumed by the next call whatever it is) → the
  call's own arguments. `run_command`, `wait_for_job`, `send_input` declare the
  three as ordinary arguments. `tool_limits` exists for tools whose schema we
  do not own (MCP, ch6).
- **`cr/io/N` holds the job's full result text.** What enters the context is
  the same bytes if ≤ cap, else head + `[... K bytes omitted; full output (T
  bytes) at cr/io/N ...]` + tail. Truncation is in `Report`, at dispatch, for
  every tool. A per-job **cursor** means no report repeats bytes: `wait_for_job`
  and `send_input` return what the model has not seen.
- A job that finished within the delay, on first report, under the cap, is
  reported **verbatim** — byte-identical to chapter 3, which is what makes
  `ch3parity` free.
- **Log contract** (students must not guess it): `tool.job = {handle, status,
  output:{kind:"handle", locator:"cr/io/N"}, bytes, exit_code?}` on both
  `tool_called` and `tool_returned`; **absent** on the four non-job tools.
  `status ∈ running | done | killed`. New event `job_killed` carries `job`
  with `reason ∈ kill_job | shutdown`. `Output` is ch2's `Ref{RefHandle}` —
  the first use of the kind ch2 declared for exactly this — recorded in the
  log and never rendered as a part (Anthropic/OpenAI renderers refuse blobs,
  which the author ruled is the lesson).
- **`run_command` runs under a PTY** (`creack/pty`, per CodeRhapsody), so
  debuggers and REPLs prompt and flush. Costs stated in the code: one merged
  stream, input echoed into output, `\r\n` normalized to `\n`, `TERM=dumb`,
  50×200. No PTY → error, never a silent fallback to pipes. Exit status is the
  file's **last** line (`exit_code: N`) because a stream cannot put it first;
  ch3's `exitCodeRe` still matches.
- No `recover` around the tool goroutine (Bill: an invariant-violation panic
  is WAI).
- Unix-only (`syscall.Kill(-pgid)`), as chapter 3 already was via `sh -c`.

## 2. Check → mechanism

All checks read behavior: request bodies the fake saw, the log via `dump` in a
fresh process, `cr/io/*` files read after the agent exited, pid liveness
(`kill -0`) of helpers that never exit, and wall clock. No check reads student
source. Helpers are Go programs **built to binaries** once per run and invoked
by absolute path; never `go run`.

| id | pts | fixture | what is observed |
|---|---|---|---|
| `ch3parity` | 10 | ch3's whole harness | `Ch3Evaluate` on the ch4 binary; every ch3 check passes. |
| `jobmodel` | 25 | `read_file`, `list_directory`, `run_command go version`, then `wait_for_job(1)` | `tool_called` carries a job with a positive handle **before** the tool ran; handles distinct and {1,2,3}; `tool_returned` status `done`; `cr/io/N` exists and its bytes **equal** the tool_result the model received; `bytes` field equals file size; `wait_for_job`'s own record has **no** job. Two of the three jobs are not the shell — the negative control. |
| `waitjob` | 10 | `sleeper 1` with delay 0.2; `wait_for_job` delay 10; `wait_for_job` again | first call's status `running` (delay honored); the wait's result contains `SLEEPER_DONE`; second wait is not an error, session reaches its final answer, wall < 12s; `cr/io/1` ends with the marker and `exit_code 0`. |
| `sendinput` | 10 | `echoer` with pattern `ECHO_READY`; `send_input hello world` with the pattern; `quit`; wait | start result has the prompt and **not** the echo; hello's result contains `echo: hello world`; file ends `ECHO_BYE` + exit 0. |
| `debugger` | 5 | `PAGER=cat dlv debug ./testdata/dbg`, pattern `\(dlv\) `; `b main.go:7`; `c`; `p answer`; `q` | start result contains `(dlv)`; break result names `main.go:7`; print result matches `\b42\b`. `dlv` resolved from PATH then `$(go env GOPATH)/bin`; absent → the check **fails** with the `go install` line (no skip). |
| `killjob` | 10 | `blocker PIDFILE` delay 0.5; wait 0.3; `kill_job`; `wait_for_job` delay **30** | start `running`; short wait says running; `job_killed{reason kill_job, status killed}` in the log; the post-kill wait says killed and does **not** claim done/exit_code; session wall < 15s (a waiter not told would sit the full 30s); pid dead after exit. |
| `bigoutput` | 15 | `flood` (1,340,013 bytes) delay 20; `read_file big.txt max_bytes 2000000`; `flood` with `max_output_bytes 2048` | file ≥ 1 MiB with first and last flood line; inline result ≤ 24 KiB (≤ 4 KiB for the small cap), names the locator and the exact total byte count; `read_file`'s file **equals** `big.txt` — the negative control for "at dispatch". |
| `toollimits` | 10 | `tool_limits{delay 0.2}`; `sleeper 1`; `sleeper 1` again; `tool_limits{pattern SLEEPER_START}`; `sleeper 1`; `tool_limits{delay 0.2}` then `sleeper 1` with explicit delay 10 | tool_limits records have no job; run1 `running` (applied); run2 `done` (one-shot); run3 `running`, result has `SLEEPER_START` and not `SLEEPER_DONE` (pattern woke it); run4 `done` (explicit wins). |
| `shutdown` | 5 | `blocker PIDFILE` delay 0.5, then the model stops | start `running`; `job_killed{reason shutdown, status killed}`; pid dead after exit. |

Sum 100. A full `-ch 4` run takes ~7s including the ch3 re-run and the live
`dlv` session; the 22-mutant audit takes ~100s wall in parallel.

## 3. The audit, in full

22 mutants of the **reference**, each asserting the exact set of failing check
ids. An edit that matches other than exactly once, **or whose replacement
changes no bytes**, is `MUTATION DID NOT LAND` and fails the test. The mutant
module carries the `creack/pty` require and the course `go.sum`, so it builds
offline and pinned.

| mutant | score | failing checks | "applied" assertion |
|---|---|---|---|
| `handle-only-for-run-command` | 60 | bigoutput jobmodel | regexp ×1 + bytes changed |
| `result-not-written-to-disk` | 60 | bigoutput jobmodel | same |
| `exit-code-not-reported` | 70 | ch3parity sendinput waitjob | same |
| `handle-issued-after-the-tool-ran` | 75 | jobmodel | same |
| `job-verbs-are-jobs-too` | 75 | jobmodel | same |
| `wait-blocks-on-a-finished-job` | 80 | killjob waitjob | same |
| `pattern-matches-output-already-seen` | 85 | debugger sendinput | same |
| `cursor-never-advances` | 85 | debugger sendinput | same |
| `send-input-writes-nothing` | 85 | debugger sendinput | same |
| `kill-marks-but-does-not-kill` | 85 | killjob shutdown | same |
| `killed-job-reported-as-done` | 85 | killjob shutdown | same — **see below** |
| `no-truncation-at-all` | 85 | bigoutput | same |
| `truncation-in-run-command-only` | 85 | bigoutput | same |
| `stub-does-not-name-the-path` | 85 | bigoutput | same |
| `kill-not-recorded` | 90 | killjob | same |
| `tool-limits-ignored` | 90 | toollimits | same |
| `tool-limits-sticky` | 90 | toollimits | same |
| `pending-limits-beat-explicit-arguments` | 90 | toollimits | same |
| `tool-limits-is-a-job` | 90 | toollimits | same |
| `ch3-read-range-ignored` | 90 | ch3parity | same |
| `no-shutdown` | 95 | shutdown | same |
| `shutdown-kills-silently` | 95 | shutdown | same |

No accepted variants this chapter: with the declined decision gone (§4) every
mutant deletes a graded behavior, and the test refuses an empty `wantFail`.

**The finding (100 → 100 on the first run).** `killed-job-reported-as-done`
deletes `j.status = StatusKilled`. The process still dies, but its
`run_command` goroutine then ends the job as `done` with `exit_code -1`, and the
waiter is told *"job 1 done, exit_code -1 … [job 1 killed: kill_job]"*. My
`killjob` check regexp'd the word *killed* in that text — and the kill **note**
in the output supplied it. A false `done` sailed through because the check read
prose. Fix: `job_killed` events carry the job's `status`, which must be
`killed`, and the post-kill report must not claim a normal end
(`done|finished|completed|succeeded|exit_code`). `shutdown` asserts the same
status, which is why the mutant now drops two checks rather than one.

**Both negative controls are load-bearing, measured.**
`handle-only-for-run-command` and `truncation-in-run-command-only` are the two
"it's a shell problem" answers. Each drops **only** because `jobmodel` and
`bigoutput` have `read_file` legs; without them both score 100.

**The fixture that lied.** The first "never returns" helper was `select {}`.
With no other goroutine the Go runtime calls that a deadlock and exits **2**,
and `killjob` and `shutdown` both failed on the reference: the process the
check wanted alive had died on its own. It now sleeps in a loop, and the
reason is in the source. The brief's warning was exact.

**Not mutated, and why.** (a) pipes-instead-of-PTY: the textual change is
larger than a regexp edit; the evidence is the measured `Stdin is not a
terminal` refusal plus `debugger` passing. (b) "pattern never wakes at all":
under it the debugger session waits 40+20+20+20s, exceeds the 45s session
timeout, and the harness SIGKILLs the agent — leaking `dlv`. The two pattern
mutants that are present (`matches-output-already-seen`,
`cursor-never-advances`) cover the pattern's semantics without that.

## 4. Where the brief, the outline and the editor disagreed

Bill was in the conversation and ruled live. Where he ruled, I followed him
over the outline; the outline needs the corresponding rework, and he asked
for that to happen after the fact.

1. **§4.6 blocking contracts (`DeadlineArg` / `MaxBlockingTime`) — removed.**
   Bill: the deadline was never the tool's business. The model sets how long
   it will wait, per call, via `ai_callback_delay` on the three waiting tools
   or via `tool_limits` for tools whose schema we do not own. The `send_secret`
   defect becomes *impossible*, not *caught*: there is no map to be missing
   from. The war story (`4bec4b96`) still teaches something — "a ruling written
   down and not wired up is invisible" — but the mechanism it defends is gone.
   **`declare` (10) no longer grades anything that exists.**
2. **§4.7 budget policy (kill / abandon-and-count / pause-and-ask) — removed.**
   Bill: nothing dies unless killed; a 3s default wake is fine *because* it
   kills nothing. **`budgetchoice` (5) no longer grades anything.** That also
   removes the chapter's P6 declined decision; see open question 1.
3. **Replacement checks, my proposal, built and audited:** `toollimits` 10
   (one-shot, defaults hold, explicit args win — and the MCP forward
   reference) and `shutdown` 5 (running jobs killed at exit, logged).
   `sendinput` 15 → 10 + **`debugger` 5**, on Bill's ruling that the exercise
   must demonstrate the fake driving `dlv`. Sum still 100.
4. **Brief: "use the exact cap numbers from the outline."** The outline has
   none. 16 KiB is Bill's number (and the design doc's).
5. **Brief/outline: "`./ch04 <transcript-file>`."** Chapter 3's contract is
   JSON-lines on stdin, and `ch3parity` requires it; kept. The transcript is
   stdin.
6. **Outline §4.5: "`BlobPart.Path`."** The ch2 amendment replaced it with
   `Ref{Kind, Locator}`; the job's output is `Ref{RefHandle, "cr/io/N"}`.
7. **Outline "what exists after ch4" says the one-shot override
   (`set_tool_watchdog`-style) exists.** Not built; `tool_limits` is the
   nearest thing and is a different idea (it raises the wake, it does not
   suspend a guard).
8. **PTY.** Not in the outline; Bill's ruling, modeled on CodeRhapsody. This
   retires open item 1 ("the PTY incident") as a *dependency*: the chapter now
   has a receipt of its own, the `dlv` refusal on pipes.
9. **`list_jobs`.** Bill unsure, zero measured use; skipped.

## 5. Open questions for the author

1. **P6.** Chapter 4 now has no declined decision. Is that acceptable, or
   should one be reinstated? Candidate that is genuinely open and gradeable:
   *what a report says about a process that exited because of a signal*
   (exit code −1? the signal name? "killed by SIGKILL"?) — three defensible
   answers, all observable in `cr/io/N`. Falsifiable: does the outline's §4.7
   get rewritten as a decision the grader accepts three answers to, or deleted?
2. **Is 3s the right default, given `ch3parity`?** Chapter 3's scripted
   `run_command` calls cannot raise the delay. Measured: `go run
   ./testdata/exit7` in a fresh module is 0.05–0.3s warm and 1.3s with an
   empty `GOCACHE`; so a 3s wake is safe by ~2×. Falsifiable: a machine on
   which a cold `go run` of a one-line program exceeds 3s would fail
   `ch3parity` with a `running` result where an exit code was expected.
3. **Does `tool_limits` consume on *every* next call, including `kill_job`
   and a second `tool_limits`?** Built: yes, any call. The alternative —
   applies to the next *job-creating or waiting* call — is friendlier and one
   line. Falsifiable: `tool_limits`, `kill_job`, `run_command` → does the run
   get the limits?
4. **Should the verbatim-first-report rule survive into the prose?** It is
   what makes ch3parity free and is invisible to students unless stated: *a
   job that finishes within the delay, under the cap, on first report, is the
   bare result*. Falsifiable: delete the rule and `ch3parity` fails on
   `toolerror`'s exact error texts.
5. **`send_input` echo.** Under the PTY the model sees its own input echoed
   back in the reply. Say so in the prose, or strip it? Stripping is a
   line-discipline flag (`ECHO` off), which changes what `dlv` shows.
   Falsifiable: with echo off, does `dlv`'s prompt line still match
   `\(dlv\) `?
6. **The debugger check makes `dlv` a course prerequisite.** Install line is
   in the failure message. Is that stated in the preface's prerequisites?
7. **Handles from 1 per process** is now a contract the fixtures depend on
   (`wait_for_job(1)`). State it, or have the grader read the handle back from
   the log (the fake's replies are fixed before the run, so the latter needs a
   two-pass fixture).

## 6. Facts

**Verified (observed directly):**
- `go test -count=1 ./internal/grade/...` — all chapters' reference and
  mutation tests — `ok` in 135s, exit 0, after the final code commit.
- `-ch 2/ch02`, `-ch 3/ch03`, `-ch 3/ch04`, `-ch 4/ch04` all 100/100 after
  the final commit; `go build ./...`, `go vet ./...` clean; `gofmt -l` lists
  only pre-existing `seam.go` and `ch02_grader_test.go`, which I did not
  touch.
- `internal/fakevendor` untouched (`git diff ch03-solution -- internal/fakevendor` empty).
- `dlv` on pipes: `Stdin is not a terminal` (dlv 1.27.2, macOS). Under the
  agent's PTY: prompt, breakpoint, `42`, quit — in the grader and live on
  `gemini-3.8-flash` (log inspected: five `tool_called`, every one carrying
  the pattern).
- `select {}` in a single-goroutine program exits 2 with "all goroutines are
  asleep - deadlock!".
- 22 mutants, all exact; `killed-job-reported-as-done` scored 100 before the
  structural fix.
- `go run ./testdata/exit7` timings above.
- Line counts in §1.

**Assumed (not checked):**
- `creack/pty` behaves the same on Linux as measured here on macOS (EIO vs
  EOF at stream end is handled by breaking on any read error, which covers
  both, but only macOS was run).
- Under heavy parallel load the 1s sleeper vs 3s default (`toollimits` run2)
  and the 0.2s delay vs 1s sleeper (run1, `waitjob`) keep their ordering.
  Margins are 2s and 0.8s; 22 parallel mutants on this machine never
  flipped either.
- The `\b42\b` assertion is not satisfiable by anything but `p answer` in
  that session's fifth result (the source listing `answer := 42` is printed
  by dlv at the breakpoint *hit*, which is the fourth result, not the fifth).
