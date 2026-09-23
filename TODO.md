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

- [ ] **Env vars: Bill's ruling vs. the graders.** Ruling 2026-09-20: "stop
      using environment variables except for the API key. I don't want to have
      to deal with them."
      Done so far: `EN_PRIMARY_SKILL` now defaults to `ensemble`, so a bare run
      needs no env vars at all. `EN_SKILLS_DIR` already has a `--skills-dir`
      flag.
      CONSTRAINT FOUND, which is why the rest is not done yet. `LLM_BASE_URL`
      is read by TWELVE grader harnesses (ch02, ch03, ch05-ch12, ch14) plus
      `cmd/fakevendor` — it is the mechanism by which every grader points the
      agent at a fake vendor, and it belongs to the same class as the API key
      rather than being behaviour config. `EN_PRIMARY_SKILL`, `EN_SKILLS_DIR`
      and `EN_CUSTOM_VAR` are set by the ch10, ch11 and ch13 harnesses (8 call
      sites), and ch10 has a check that explicitly exercises `EN_CUSTOM_VAR`.
      Converting those to flags ALSO breaks grading of the frozen
      `solutions/ch10`-`ch14` snapshots, which only understand env vars.
      PROPOSAL: keep `LLM_BASE_URL` and the API key as endpoint/credential
      config; leave the three skill vars as overrides that only graders set,
      now that the defaults make them unnecessary for a human. Still readable
      by env var, never required. Say the word if you want them gone entirely
      and I will migrate the harnesses and re-snapshot the solutions.
      Remaining behaviour vars that could become flags cheaply, none of which a
      human needs: `CH02_LOG`, `LLM_VENDOR`, `EN_DISABLE_STREAMING`.

- [ ] **Seventh TTS defect: bullets are spoken with their dashes.**
      `"- a\n- b"` is heard as "a dash b". `queueChunk` flattens lone newlines to
      spaces *before* `_speakable` runs its `^`-anchored `/gm` bullet regex, so
      every bullet after the first survives. Same shape as the shattered fence:
      a filter anchored to line structure running after line structure was
      destroyed. Chapter 14 documents six defects and draws its generalisation
      from one instance; this is arguably the better generalisation.
      Fixing it changes the reference solution, so it is the author's call.

- [x] **Eighth TTS defect: tool results are spoken aloud.** FIXED 2026-09-20 in
      `bc7610d`. Found by the new ch14 grader, which failed the reference on
      first contact: plant a file containing `ZEBRAFISH`, script the model to
      read it and then say a sentence containing `PELICAN`, and the speech log
      recorded the file's contents read out byte by byte between the two.
      Root cause was provenance, not audio. On `ToolCompleted` the actor
      announced the result twice: correctly as `ToolFinished`, and again
      wrapped in a `TextPart`. There is no `tool_result` DeltaKind, so every
      consumer downstream of the second event believed the model had spoken
      those bytes. The tool card renders from `ToolFinished`, so the duplicate
      was visually redundant and deleting it lost nothing.
      Bill's ruling, confirmed 2026-09-20: tool results are not to be spoken.

- [x] **Ninth TTS defect: a failed model endpoint is never spoken.** FIXED
      2026-09-20 in `4dbb925`. Point `LLM_BASE_URL` at a server returning 500
      and a listener heard the typing pause, then silence forever.
      CORRECTION to the original diagnosis recorded here, which named the wrong
      mechanism. It was NOT that "the producer never fires". No `ErrorOccurred`
      event is emitted for an API failure at all; the reason travels as the
      `error` FIELD on `turn_ended` (`handler.go:693` sets `m["error"]`, and
      `TurnEnded.Err` carries it). The message did arrive. The client's
      `turn_ended` handler called `_scrollToBottom()` and dropped the field, so
      a failed turn was neither displayed nor spoken, while `speakError` sat two
      hundred lines away with a comment explaining why errors above all must be
      voiced. Fixed by routing a `turn_ended` carrying an error through the
      existing `_handleError` path.
      Same shape as the unwired pause gate that chapter 14 opens with: both
      ends built, the wire never connected.
      METHOD NOTE worth keeping: the wrong diagnosis came from grepping for
      emitters instead of watching the wire. What settled it was running the
      agent against a 500 endpoint and reading its own stdout, which showed
      `{"error":"anthropic: ...","observation":"turn_ended"}` in one line.
      A narrow grep for `ErrorOccurred{` also returns zero emitters and invites
      a second wrong conclusion; the real form is `common.Event{Type: ...}` at
      `agent/internal/llm/seam.go:276`, `engine.go`, and `actor.go`.

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

- [ ] **`max_tool_rounds` is a dead setting (ch9).** VERIFIED 2026-09-23.
      Added in `8fd790f` (ch9); `agent/internal/common/settings.go` stores and
      clamps it, and nothing else reads `.MaxToolRounds`. Both dispatchers stop
      at `const MaxToolRounds = 16` (`agent/internal/llm/engine.go:224`, checked
      at `engine.go:275` and `actor.go:243`). Found by the ch15 measurement:
      60 reads in one prompt stopped at 16 rounds with `max_tool_rounds: 100`.
      Not ch15's contract. Fix belongs to ch9's settings path; the ch15
      grader's longest turn is 13 rounds, so fixing it moves no ch15 score.
- [ ] **Anthropic renderer never sets `cache_control`.** VERIFIED 2026-09-23.
      `grep -rn cache_control agent --include=*.go` finds nothing. A real
      `claude-opus-5` run (ch15 §15.15, `context_target` 20000, 11 requests)
      reported `cache_creation_input_tokens` 0 and `cache_read_input_tokens` 0
      on every request. The byte-stable prefix ch15 protects earns no discount
      on this vendor until a breakpoint is sent. Which chapter owns it (ch2
      renderer or ch15) is the author's call.

- [x] **A bare `./ensemble --port 8084` had only two tools.** FIXED 2026-09-20
      in `6335237`. Bill hit this live: the agent truthfully reported that it
      could not run a shell command, because its entire toolset was
      `load_skill` and `unload_skill`.
      Two independent causes. `IsToolEnabled` enforced "only tools a loaded
      skill declares" over a set that was usually empty, and an empty set made
      the loop never run, so every tool fell through to the default deny. A
      filter over an empty set is vacuously true and catastrophic. Separately
      the primary skill defaulted to `""`, so nothing was ever loaded unless
      `EN_PRIMARY_SKILL` was set.
      Measured before and after by capturing the declarations sent to the
      model: 2 tools before, 12 including `run_command` after.
      TRAP, hit and recorded: making a missing primary skill FATAL broke the
      ch7/ch8/ch9 parity scenarios, which run the agent with no skills
      directory at all and must keep unfiltered behaviour. It warns instead.
      Hard-coding the name as a `const` also broke ch10/ch11 (which load
      `base`) — it has to be a default, not a constant.

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
