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

- [ ] **UNDER INVESTIGATION — `TestCh4DeletionAudit`.** Appeared in the full
      suite on 2026-09-20 *after* the `--tts-log` work (commit a4f47be), and was
      NOT in that morning's verified baseline set of six. But it **passes when
      run alone** (`go test ./internal/grade/ -run TestCh4DeletionAudit`, exit
      0). That asymmetry points at test interference or a parallelism flake
      rather than at the tts-log change, which touches `ws/`, `web/gui/` and
      `main.go` only. NOT YET PROVEN EITHER WAY.
      Next step: run the full `internal/grade` package twice to establish
      whether it is intermittent, then compare against a worktree at `a4f47be^`.
      Related history: ch4 had a real 13/120 flake earlier in the project,
      traced to parents holding stale child facts and since fixed.

---

## In flight

- [ ] **ch14 grader revision.** Design agreed and recorded in
      `book/proposal-ch14-revision.md`. The node-based checks are to be removed:
      they pin method names the TL;DR never promised, and `eval` cannot load ES
      modules, so a correct student using `export const TTS` scores zero.
      Replacement reads a speech log produced by a student-supplied harness.
      `--tts-log` shipped 2026-09-20 (a4f47be). Remaining: reference harness
      script, rewrite `internal/grade/ch14_*`, re-snapshot `solutions/ch14`,
      move the `ch14-solution` tag, update `book/review-ch14.md`.
