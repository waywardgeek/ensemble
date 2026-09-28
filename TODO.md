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

## Crossover — what Ensemble needs before it replaces CodeRhapsody

Bill is running both side by side as of 2026-09-27. This section is the list of
things standing between "impressive demo" and "the agent I actually work in."
Ordered by what blocks daily use, not by difficulty.

Deliberately **not** on this list: sub-agents. Bill: *"I can live without
sub-agents for a week."* Gateway-first is the agreed sequencing, since routing
between agents and routing to a remote peer are the same infrastructure.

### 1. Caching — the top blocker (VERIFIED 2026-09-27)

**Partly done (2026-09-28).** The system prompt now carries a breakpoint: it is
a content-block array with `cache_control` on it, so tools and system together
form one cacheable prefix. The lens reports `system:identical` and a cacheable
prefix of 13,657 of 14,213 bytes (**96.1%**) on a short conversation.

Still to do: a second breakpoint inside `messages` — the only section that
grows, so the one that matters on a long conversation — and header stats
reporting the *actual* hit rate read back from the response, rather than the
predicted figure the lens computes from the request.

Benchmark to beat: **CodeRhapsody currently runs at 91.3% cache hit rate**
(VERIFIED 2026-09-27, Bill's live session). That is the bar, and it proves the
number is achievable in an agent of this shape rather than aspirational.

Full design in `book/caching-design.md`. The short version:

- Capture `cache/request.json` + `cache/prior_request.json`, canonically
  serialized **in wire order** (system → tools → messages) so a plain `diff`
  lands on the first byte that broke the cache.
- Classify every request as MATCH / PREFIX DIVERGED (our bug) / CACHE MISSED
  ANYWAY (TTL, size floor, eviction). Collapsing the last two into one alert is
  how the alarm gets ignored.
- **Measure before optimizing.** A breakpoint on an unstable prefix buys
  nothing and looks like it should have worked.

~~Known structural blocker~~ **RESOLVED 2026-09-28.** `anthRequest.System` is
now `[]anthSystemBlock`, so it can carry `cache_control`. The byte-identical
property it was thought to break is asserted by nothing: `ch10_harness.go`
already type-switches with the comment "The system field can be a string or an
array of content blocks", and ch15–17 hold the field as `json.RawMessage`. The
invariant lived in comments, not in checks. Rejected a custom `MarshalJSON`
that would emit a string when unmarked, on the grounds that it adds machinery
to protect a property no check asserts.

Known stability risk (ASSUMED, not yet measured): tool declaration order. Ensemble
makes no ordering guarantee; CodeRhapsody sorts by `(skill_name, tool_name)`
specifically for cache stability. Needs a test either way.

### 2. Session stats in the GUI (VERIFIED 2026-09-27)

Bill wants, scoped to one Ensemble restart: session cost, total input tokens,
cache read tokens, and cache hit rate.

The four token categories already exist and are already disjoint (`Input`,
`CacheWrite`, `CacheRead`, `Output`) — the right shape. Two gaps:

- Usage is emitted to **stdout only** and never sent over the WebSocket; the GUI
  has no usage handling at all.
- No price table, deliberately: `event.go` says *"Usage counts tokens and never
  money. Prices change; counts are history."* Keep that. Cost is computed at
  render time from a table beside the model catalog.

### 3. Header controls (requested 2026-09-27)

- ~~Reset conversation.~~ **DONE 2026-09-28** (`b3a2c75`). Travels as
  `common.Reset` like `interrupt`; the reducer clears `KindDialogue` and
  `KindRecall` only. The re-snapshot warning turned out not to apply: reset is
  recorded as a `ConversationReset` **event** rather than a direct edit, so the
  journal carries it and replay re-applies it on load. Nothing is truncated, so
  there is no stale snapshot to undo it. Four mutants, each killed by exactly
  one test.
- Interrupt button. Protocol already works end to end (hub accepts
  `"interrupt"`, actor handles `common.Interrupt`) — this is a button only.
- Model dropdown in the header, from the server-owned catalog.
- ~~"Don't be evil", tiny font, somewhere on screen.~~ **DONE** —
  `index.html:46`, `<div id="creed">`.

### 4. Memory lives in the source tree (VERIFIED 2026-09-27)

193 daily logs + SOUL.md + MEMORY.md + 2 cascade buckets currently sit in
`agent/memory/`, inside the git repo, because the code reads
`filepath.Join(".", "memory")`. Should move to `~/.en/memory/` beside
`~/.en/settings.json`.

### 5. Gaps that are known and accepted for now

- No learnings mechanism. CodeRhapsody has 50; they would have to fold into
  MEMORY.md or become a feature.
- No web search, no browser tools, no screenshots.
- Browser only; no desktop wrapper.
- The `ensemble` skill is the book's teaching persona, not a daily-driver
  identity. SOUL.md and MEMORY.md are loaded, but the skill text around them
  still describes the teaching agent.

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

- [x] **`agent/verify_live.sh` is untracked.** CLOSED 2026-09-27: the file no
      longer exists, so the decision made itself. Full entry under Repo
      hygiene.

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
- [ ] **Six of the twelve settings are dead, and the doc comment claims
      otherwise.** VERIFIED 2026-09-26. `Model`, `Temperature`, `MaxTokens`,
      `ThinkingBudget`, `MaxToolRounds` (the entry above) and `SystemPrompt`
      are stored, clamped and persisted, but nothing reads them.
      `SettingsStore` is `{mu, data, path}` with no engine reference, and
      `ApplyRaw` only mutates `s.data` and calls `persist()`, so no settings
      change can reach the engine. The store is built at `cmd/main.go:360`,
      after the engine, so it does not feed the boot config either. The only
      three readers of `settings.Get()` are `cmd/main.go:361`
      (`ContextTarget`), `cmd/main.go:689` (`LogRetention`) and
      `internal/ws/handler.go:274` (snapshot to a newly connected client).
      Grep for post-startup mutation of the engine config finds nothing, and
      there is no model picker in the GUI.
      The stale comment at `internal/common/settings.go` is corrected in this
      commit; the wiring itself is still open, and which chapter owns it
      (ch9 built the panel) is the author's call.
- [ ] **`Model` is not a renderer-only setting — hazard if it is ever wired.**
      VERIFIED 2026-09-26. It gates log *content* through
      `LookupModel(e.Cfg.Model).StubsToolResults` (`internal/llm/policy.go:36`,
      which decides whether rule 6 writes `Redacted` events) and the tool set
      through `cmd/main.go:307` (`RemoveTool(KeepToolResults)`). Both resolve
      against the startup model today, so they agree. Wiring live model
      switching without re-resolving both together reproduces the bug fixed in
      CodeRhapsody on 2026-09-26: stubbing stays on while `keep_tool_results`
      is absent, so the model is redacted with no way to keep anything. The
      invariant is currently prose in a comment, not structure.
- [x] ~~**Anthropic renderer never sets `cache_control`.**~~ **FIXED 2026-09-28**
      (`9898053`). The system prompt is now a content block carrying an
      ephemeral breakpoint, so tools and system form one cacheable prefix. The
      byte-stable prefix ch15 protects finally earns its discount. Re-measure
      the live figures that motivated this item (`cache_creation_input_tokens`
      and `cache_read_input_tokens` were 0 on all 11 requests): the lens
      predicts 96.1% cacheable, but only a real run reports what the vendor
      actually granted. Which chapter owns the prose (ch2 renderer or ch15) is
      still the author's call.

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

- [x] **Committed binaries purged from history.** DONE 2026-09-27. The repo
      carried 72.4MB of Mach-O binaries across `solutions/ch05`, `ch09`-`ch13`
      and `ch17`, which was **87% of all history** (82.9MB of blob storage;
      book/ was 5.9MB and everything else 4.5MB). `git filter-repo
      --strip-blobs-bigger-than 1M` removed them, together with
      `book/the-self-wielding-agent.md` (14 revisions, 2.17MB), which is
      GENERATED by `scripts/assemble-book.sh` and should never have been
      tracked. A second pass removed a committed `.pyc`, five
      `solutions/*/settings.json` carrying runtime state including a personal
      `tts_speed`, and a `virtual-user.jsonl` log.
      `.git` went from **113M to 16M**. Backup mirror of the pre-rewrite repo
      is at `~/ensemble-backup-20260927-125051.git`.
      The 1MB threshold was not a guess: the 400KB-1.1MB band was empty, so
      there is a clean gap between the largest surviving text blob (210KB) and
      the smallest binary.
      VERIFICATION WORTH REUSING: file count at HEAD went 1096 -> 1087 -> 1080,
      matching exactly the 8 binaries + 1 generated doc, then 1 pyc + 5
      settings.json + 1 jsonl. Five commits were pruned as empty (main 423 ->
      418), each touching only purged paths. Full grader sweep after the
      rewrite: 25/25, every chapter 100/100.
      TRAP: `filter-repo` rewrites old commit hashes *inside commit messages*,
      so a commit whose subject ends "snapshot of agent/ at 41930f6" gets a new
      subject. Exact-string matching of subjects across the rewrite therefore
      reports false losses. Compare file counts and content, not subjects.

- [x] **`agent/bin` stray 10MB executable FILE.** DELETED 2026-09-27.

- [x] **`agent/verify_live.sh` is untracked.** Resolved: the file no longer
      exists. Entry closed 2026-09-27.

- [ ] **`.gitignore` now relies on a directory re-include; do not "simplify"
      it.** The built-binary rules are depth-independent because a snapshot
      under `solutions/` nests a whole repo layout. But `ensemble` and
      `virtual-user` name DIRECTORIES as well as binaries
      (`agent/skills/ensemble/SKILL.md` is the primary skill,
      `agent/cmd/virtual-user/` is source), so the rules are paired with
      `!**/ensemble/` and `!**/virtual-user/`. Flattening those to bare
      patterns silently untracks the primary skill. VERIFIED 2026-09-27 by
      `git ls-files --ignored --exclude-standard -c`, which must stay empty.

- [ ] **`cr/docs/ch02-impl-spec.md` is tracked but ignored** by the `cr/` rule
      at `.gitignore:44`. Pre-existing, not caused by the purge. Either
      un-ignore it or stop tracking it. VERIFIED 2026-09-27.

---

## Graders

- [x] **The "pre-existing failures" section is obsolete.** RESOLVED 2026-09-27.
      It recorded ch6 at 0/100, `grade2` at 90 and `grade3` at 85. Those were
      artifacts of grading the wrong target: a chapter's score is only
      meaningful at its canonical target, and the list of targets lives in
      `scripts/gradesweep.sh`. The one real defect was ch6, whose harness built
      `./cmd/` instead of the sibling `ch06/main.go`. Full sweep is now 25/25
      with every chapter at 100 (ch5 at 120/120), re-confirmed 2026-09-27 after
      the history rewrite.

- [x] **The ch4 deletion audit was flaky under its own parallelism.** FIXED
      2026-09-27 in `fcf9a2b`. All 28 mutants ran under `t.Parallel()`. Each
      builds an agent and drives live child processes, and several checks
      assert on what happened before a deadline, so they starved each other and
      the results described the machine rather than the code. Measured:
      `cursor-never-advances` failed in the suite and passed 3/3 alone;
      widening the margin moved the flake to `job-verbs-are-jobs-too` and
      `handle-issued-after-the-tool-ran` losing their `shutdown` kill, and both
      of those passed alone too. Contention was reaching several unrelated
      checks, not one. `t.Parallel()` is gone; the audit went from about 90s to
      5m55s and now passes as a whole. That is the price of failures that mean
      something, and this audit is what stops the ch4 grader being decoration.

      Also fixed the latent race the flake exposed. `send_input` waits for the
      pattern `ECHO_READY`, which the echo fixture printed both at startup and
      after each reply, so an implementation that rescans the whole buffer
      matched the startup banner and returned at once. Correct and broken were
      a millisecond apart, so the check could pass against a stale cursor by
      luck. The fixture now pauses 500ms before answering, far longer than the
      scheduling noise it must beat and far shorter than the 10s callback delay
      the correct path is given.

- [ ] **`killjob` may not distinguish a real kill from a kill that only
      marks.** Separate from the flakiness fixed above, and still open. The
      `kill-marks-but-does-not-kill` mutant replaces
      `syscall.Kill(-pid, SIGKILL)` with a no-op, so the child should survive,
      yet the child often dies anyway, most likely reaped during process-group
      teardown when the agent exits. Removing the parallelism removed the false
      failures but did not prove the check can tell the two cases apart.
      Fix direction: observe liveness *before* the agent exits, rather than
      from a pid file read afterwards.

      Cautionary note on how this was nearly misdiagnosed: an earlier run of
      `go test ... -run TestCh4DeletionAudit -v | head -25` reported exit 0 and
      was recorded here as "passes when run alone". Both halves were wrong.
      `$?` after a pipeline is **`head`'s** exit code, not `go test`'s, and
      every subtest was `t.Parallel()` so `head` truncated the output before any
      result line was printed. Redirect to a file and check the exit code
      before the pipe.

- [x] **The star-topology check governed three spokes out of nine, and never
      ran against live code.** FIXED 2026-09-27 in `cd10cd1`.
      `internal/grade/ch06_checks.go` hardcoded `llm`, `tools` and `jobs`, and
      the loop was `if !ok { continue }`, so every spoke added since was
      skipped in silence. It now derives the spoke set from the import graph:
      every non-hub package under `internal/` is a spoke, so a new one is
      governed the day it appears rather than the day someone remembers to
      edit the list.
      Two further defects found while fixing it. `DiscoverImportGraph` returns
      an empty map when `go list` fails, and an empty spoke set made the scan
      vacuously clean, so a tree that did not compile scored full marks for
      architecture; that is now an explicit failure. And the check only ever
      ran against `solutions/ch06/agent`, so the live tree was unchecked; a new
      `internal/grade/star_live_test.go` enforces the invariant on `./agent`
      directly. It is not decoration: adding an import of `internal/jobs` to
      `internal/recall/bm25.go` fails it, and the mutation was reverted.

- [x] **Three grader tests had been failing continuously since `34ef50b`.**
      FIXED 2026-09-27 in `cd10cd1` and `8654f47`. "Unify grading: every
      chapter grades one directory" moved the exercise trees under `solutions/`
      and left ch6, ch7 and ch8 pointing at the repo root. ch6 and ch7 copied
      root-level `ch05/`, `ch06/`, `ch07/` that no longer exist there; ch8
      passed `"."`, which under `go test` is `internal/grade`, so the grader
      graded its own package directory. All three then handed the wrong path to
      a `ChNRun` that wants the agent module dir, `go build ./cmd/` failed with
      "no Go files", and every check reported 0. The result looked like
      catastrophic architecture failure while `gradesweep.sh` reported 100 from
      the correct target. Also fixed a stale `ch3` check count of 9, which has
      been 10 since `writeguard` landed, and dropped a `t.Parallel()` from
      `TestCh8Grade`.

      The rule this cost us: **a chapter's score is meaningless anywhere but
      its canonical target**, and that list lives in `scripts/gradesweep.sh`.
      ch1-ch4 grade `solutions/chNN`, ch6 alone grades a frozen snapshot, ch5
      and ch7-ch17 grade the live `./agent`. Do not generalise from one
      chapter: pointing ch7 at `solutions/ch07` scores 90/100, because that
      snapshot is stale and graded by nothing, and "fix" it by re-snapshotting
      and you have hidden a wrong target behind a correct-looking number.

- [x] **The ch14 harness leaked one agent per scenario, forever.** FIXED
      2026-09-27 in `2c1a379`. Found 35 orphaned agents on this machine, the
      oldest 4 days 19 hours, each holding a port and a temp directory, and a
      plausible cause of a session freeze. Two independent defects.
      Ordinary path: the script starts the agent in a subshell, which forks
      because it runs `cd` first, so `$!` named the subshell and cleanup killed
      the wrapper while the agent was reparented to init. Five scenarios, five
      leaks, every run. Fixed with `exec`.
      Timeout path: `ch14_harness.go` called `cmd.Process.Kill()`, sending
      SIGKILL to the script. SIGKILL cannot be trapped, so `trap cleanup EXIT`
      never ran. The harness now uses `Setpgid` and signals the whole group,
      SIGTERM first so cleanup can run. `Setpgid` appeared nowhere in
      `internal/grade` before this, and ~20 other `Process.Kill()` sites remain
      unaudited for the same shape.
      Verified by reproduction: a fixture of the same shape leaks under
      `cmd.Process.Kill()` and is clean under the group signal. The first
      attempt proved nothing because it copied `/bin/sleep`, and macOS kills a
      copied signed binary on sight, so the canary never ran and both arms
      looked clean. **Verify a fixture produces the condition under test before
      trusting either result.**

- [ ] **`no-mutable-globals` false-positives on compile-time interface
      assertions.** `var _ common.Skills = (*SkillRegistry)(nil)` is the
      standard idiom and the check rejects it. On 2026-09-27 the assertions
      were deleted to keep the check strict, which was the wrong trade: the
      check is over-broad and should exempt `var _ T = (*U)(nil)`.

- [ ] **ch17's `per-source-quota` check is pure decoration.** The fixture never
      creates crowding, so the check cannot fail. Three further ch17 checks
      (`judge-filters`, `judge-fallback`, `recall-is-own-kind`) were never
      mutation-tested. A passing grader is evidence of nothing until a mutant
      kills each check exactly once. VERIFIED 2026-09-27.

- [ ] **ch12-ch15 were never regression-tested against the ch17 changes.**

---

## Book prose

- [x] **`chapter-05.md` licensed the very hub drift that ch05 forbids.** FIXED
      by the author in `f5d38c9`, verified 2026-09-27.
      The problem: the chapter stated the rule correctly ("Everything in it is
      vocabulary: types, constants, interfaces. It has no behavior of its own")
      but never said what to do when behavior *operates on* a hub type. Go
      forbids defining a method on a non-local type, so a coder hit `cannot
      define new methods on non-local type common.Context` and the shortest
      path out was to dump the function in `common`. That is how the hub grew
      to 4,917 lines at 51% function bodies. The chapter also carried stale
      counts ("1,455 lines, seven files") and the sentence "This is the largest
      package because the vocabulary is large, and that is correct", which read
      as permission.
      All three are now gone: `chapter-05.md:227-232` gives the mechanism
      (free functions over hub types, `Apply(c, e)` not `c.Apply(e)`), and both
      the stale counts and the permission sentence are absent.
      Bill's ruling 2026-09-27: "It is simply better to give up on method call
      syntax." Full argument in `book/common-refactor-design.md`.
      STILL OPEN, and the reason this entry is worth keeping: the chapter now
      teaches the rule, but nothing *enforces* it. See the star-topology entry
      under Graders.

---

## Shipped

- [x] **ch14 grader revision.** DONE. The node-based checks are gone:
      `internal/grade/ch14_driver.js` no longer exists. They pinned method
      names the TL;DR never promised, and `eval` cannot load ES modules, so a
      correct student using `export const TTS` scored zero. The replacement
      reads a speech log produced by a student-supplied harness
      (`agent/scripts/tts-harness.sh` + `tts-browser.js`), which runs the real
      path end to end in about five seconds warm: agent, WebSocket, `gui.js`,
      `artifact-scroll.js`, `tts.js`, back over the socket, into the log.
      It names no method and no file layout, so it grades the channel rather
      than our implementation of it. ch14 scores 100/100 in the sweep.
      The design record `book/proposal-ch14-revision.md` was removed from HEAD
      on 2026-09-27 as completed scaffolding; it remains in git history.
