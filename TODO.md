# TODO — ensemble

Tech debt, deferred decisions, and things we do not want to forget.

**Why this file exists and not `// TODO` comments.** As of 2026-09-20 there are
**zero** TODO/FIXME/XXX comments in `agent/`. That is worth keeping on purpose.
The reference solution is teaching material, and an inline TODO there reads to a
student as either an exercise or as sloppiness. Debt is tracked here instead.

**Why the repo root and not `agent/`.** `solutions/chNN` are snapshots of
`agent/`, so a file placed there gets copied into every future snapshot and then
freezes. `solutions/ch09/TODO.md` would list items ch12 already fixed. Root also
covers debt that spans book prose, graders and repo hygiene, which `agent/`
cannot.

Mark every entry VERIFIED (observed directly, with a date) or ASSUMED
(inferred, not checked). The two read identically later if you let them blur.

---

## Open questions awaiting the author

- [ ] **Seventh TTS defect: bullets are spoken with their dashes.**
      `"- a\n- b"` is heard as "a dash b". `queueChunk` flattens lone newlines to
      spaces *before* `_speakable` runs its `^`-anchored `/gm` bullet regex, so
      every bullet after the first survives. Same shape as the shattered fence:
      a filter anchored to line structure running after line structure was
      destroyed. Chapter 14 documents six defects and draws its generalisation
      from one instance; this is arguably the better generalisation.
      Fixing it changes the reference solution, so it is the author's call.
      VERIFIED 2026-09-20 (found when a grader fixture failed against the
      reference).

- [ ] **`Ephemeral` conflates two different ideas.** Auto-called tools (must be
      fast and reliable, flag-gated) and auto-redacted tools (the model calls
      them normally, but results do not accumulate) both use
      `Ephemeral:"round"`. Bill: "leave as-is, I want to think on it."
      Raised ch12, 2026-09-19.

- [ ] **`callEphemeral` invokes `Tool.Run` directly**, bypassing the job system.
      Open item E2 in `book/review-ch12.md`, awaiting a ruling.

- [ ] **ch14 TL;DR names only `flush()`**, never the chunk entry point, so any
      grader touching the pipeline assumes a surface the contract never
      promised. Violates the house rule that a fresh coder scores 100/100 from
      the TL;DR alone. VERIFIED 2026-09-20.

- [ ] **`agent/verify_live.sh` is untracked.** Commit it or delete it. Flagged
      twice across sessions without a decision.

---

## Agent code

- [ ] **`runActorLoop` takes eleven positional parameters** (`agent/cmd/main.go`),
      mostly strings and bools, with `mcpPipe bool, guiDebug bool` adjacent.
      Two adjacent bools are one transposition away from a silent bug. Wants a
      config struct. Deliberately not refactored mid-task on 2026-09-20.
      VERIFIED 2026-09-20.

- [ ] **Flags are hand-parsed in a `switch`** rather than using a flag package,
      so every new flag needs two cases (`--x value` and `--x=value`) and
      `--help` can drift out of sync with what is actually parsed. `--port` and
      `--tts-log` are both absent from `--help` today. VERIFIED 2026-09-20.

---

## Repo hygiene

- [ ] **`agent/bin` is a stray 10MB executable FILE, not a directory.**
      Untracked, so it is a local build artifact. It collides with the
      conventional `bin/` directory: any script doing `mkdir -p agent/bin`
      fails with "File exists", which is how it was found. The ch14 harness
      now builds into a scratch directory instead, so nothing depends on it.
      Delete it, or rename whatever produces it. VERIFIED 2026-09-20.

- [ ] **`solutions/ch09` through `ch13` track roughly 49MB of committed
      binaries**, plus logs and a `settings.json` carrying runtime state
      including a personal `tts_speed`. `solutions/ch14` excludes all of these
      (580K, 64 files) and is the pattern to follow when re-snapshotting.
      VERIFIED 2026-09-20.

---

## Graders — pre-existing failures

Proven pre-existing on 2026-09-20 with `git worktree add /tmp/ens-base HEAD`,
running identical commands with the session's work absent. Not regressions.

- [ ] `grade2` scores **90/100**; `grade3` scores **85/100**. The ch14 coder
      brief's definition of done expects both to pass, so it was written
      against a cleaner tree than exists.
- [ ] **ch6 reference scores 0/100.**
- [ ] `go test ./...` fails: `TestCh3PointsSumTo100` ("got 10 checks, want 9"),
      `TestCh6ReferenceScores100`, `TestCh6DeletionAudit`,
      `TestCh7ReferenceScores100`, `TestCh7DeletionAudit`, `TestCh8Grade`.

- [ ] **`TestCh4DeletionAudit/kill-marks-but-does-not-kill` is FLAKY.**
      RESOLVED 2026-09-20: intermittent and pre-existing, not a regression.
      Two independent proofs.
      *Scope*: the audit grades `solutions/ch04` (`ch4ReferenceDir`,
      `internal/grade/ch04_grader_test.go:47`), a frozen snapshot. Nothing
      committed on 2026-09-20 touches that tree.
      *Behaviour*: three consecutive runs at HEAD gave exit 1, exit 1, exit 0,
      with 3, 2 and 0 failing lines. It passes sometimes.

      The underlying weakness is real and worth fixing. The mutant replaces the
      `syscall.Kill(-pid, SIGKILL)` call with a no-op, so the child should
      survive; the audit expects both `killjob` and `shutdown` to fail. Often
      only `shutdown` fails, meaning **the `killjob` check cannot reliably tell
      a real kill from a kill that only marks** — the child dies anyway, most
      likely reaped during process-group teardown when the agent exits.
      All 7 mutants run under `t.Parallel()`, so the suite puts a
      timing-sensitive kill test under self-inflicted load.
      Fix direction: have the `killjob` check observe liveness *before* the
      agent exits, rather than from a pid file read afterwards.

      Cautionary note on how this was nearly misdiagnosed: an earlier run of
      `go test ... -run TestCh4DeletionAudit -v | head -25` reported exit 0 and
      was recorded here as "passes when run alone". Both halves were wrong.
      `$?` after a pipeline is **`head`'s** exit code, not `go test`'s, and
      every subtest is `t.Parallel()` so `head` truncated the output before any
      result line was printed. Redirect to a file and check the exit code
      before the pipe.

---

## In flight

- [ ] **ch14 grader revision.** Design agreed and recorded in
      `book/proposal-ch14-revision.md`. The node-based checks are to be removed:
      they pin method names the TL;DR never promised, and `eval` cannot load ES
      modules, so a correct student using `export const TTS` scores zero.
      Replacement reads a speech log produced by a student-supplied harness.

      Done: `--tts-log` (a4f47be), gate cause in the log (fce3221), reference
      harness `agent/scripts/tts-harness.sh` + `tts-browser.js` (38ec1eb).
      The harness runs the real path end to end in about five seconds warm:
      agent, WebSocket, `gui.js`, `artifact-scroll.js`, `tts.js`, back over the
      socket, into the log.

      Remaining: rewrite `internal/grade/ch14_*` around plant-file /
      script-vendor / run-harness / read-log, delete `ch14_driver.js`, rewrite
      `scripts/ch14-mutations.sh` and re-run the P9 audit, re-snapshot
      `solutions/ch14`, move the `ch14-solution` tag, update
      `book/review-ch14.md`.

      Known gap for the grader author: the fake vendor's reply arrives in a
      few hundred milliseconds, which is too fast to type into. The
      pause-gate scenario needs a long scripted reply, or the check should
      rely on the `cause:"typing"` line that sending the prompt already
      produces.
