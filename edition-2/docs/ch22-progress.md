# Ch22 coder progress log

Working tree: `~/projects/ensemble`. Library module is `agent/`
(`github.com/waywardgeek/ensemble/agent`); root module holds
`internal/grade`, `internal/fakevendor`, `solutions/`, `scripts/`, `book/`.

Brief path translation: brief says `internal/llm/actor.go`, means
`agent/internal/llm/actor.go`. Brief says `cmd/main.go`, means
`agent/cmd/main.go`. Brief says `internal/grade/ch05_harness.go`, means
ROOT-level `internal/grade/ch05_harness.go`.

## Facts verified in the live tree (2026-10-06)

- `agent/cmd/main.go` = 1,154 lines. VERIFIED.
- `NewAgent` (agent/agent.go:205) has ZERO callers. VERIFIED by grep across
  both modules excluding solutions/.
- `NewBareAgent` used only by `agent/cmd/virtual-user/main.go:111`. VERIFIED.
- `cliHost` declared agent/cmd/main.go:265, comment "the composition root for
  the server". VERIFIED.
- Closure staples VERIFIED: `hub.Model` (cmd/main.go:675), `eng.Target`,
  `eng.ToolRoundLimit`, `eng.Bands` (cmd/main.go:432-441).
- Cost bug VERIFIED at `agent/internal/ws/handler.go:438`:
  `CostUSD(session, f.Price)` prices whole session at current model rate.
- Dispatch site VERIFIED `agent/internal/llm/actor.go:441`: builds
  `&common.Call{Host: a.host, Jobs: a.eng.Jobs, Limits: limits}` — reaches
  into `a.eng` for `.Jobs` and discards the rest.

## NEW finding for the author (not in the outline)

`common.Host` ALREADY carries `RecordUsage`/`SessionUsage`/`LastUsage`, and
the comment above them in `agent/internal/common/interfaces.go` justifies it
explicitly: usage lives on Host *because Host is the thing every object can
reach*. That is the decay mechanism stating its own motive in a source
comment. It is a better exhibit than the `hub.Model` closure because it is
not a staple bolted on from outside — it is the back-pointer interface
accreting an unrelated capability from the inside.

## Constructor inventory (back-pointer rule)

HONOR the rule (take `common.Host`): `NewEngine`, `NewActor`, `NewFramework`,
`NewJobs`. These are exactly the ch5/ch6 spine.

VIOLATE the rule: `ws.NewHub` (6 params), `cachelens.New(dir, report func)`,
`llm.NewJudge(*Engine)` (concrete, not interface), `recall.New`,
`settings.NewSettingsStore(path)`, `store.NewStore(root)`.

## Stage status

- [x] 0. Process fix: SKILL.md Architecture Invariants + course-policy P11
      -- commit 0803991
- [x] 1. Fix ch5 `detectLogf` (P9 violation; blocked the rename)
      -- commit 1e5c456, new `internal/grade/astscan.go` + 6 sharpness tests
- [x] 2. Rename `common.Host` -> `common.Agent`
      -- commit 7a38c42, 15 files, 66/66 insert/delete, pure rename
- [x] 3. `common.Engine` interface + `Call.Engine`        -- commit 80dd708
- [x] 4. Per-model usage on Engine; fix cost bug          -- commit 80dd708
- [x] 5. `agent_status` tool                             -- commit fa9b6e4
- [x] 6a. NewAgent builds a complete agent from a data directory
      -- commit 70acadd, new AgentSpec + attachState + 2 tests
- [x] 6b. Collapse composition roots; cliHost deleted
      -- commit 6077d4b, main.go -214 net lines
- [x] 7. Snapshot `solutions/ch22`                       -- commit df86628
- [~] 8. Grader `internal/grade/ch22_*` (7 checks / 100 pts)
      - [x] 8a `back-pointer-chain` + astscan ResultTypes -- commit 22104ff
      - [x] 8b skeleton + `agent-builds` + `reaches-through-the-chain`,
            wired into cmd/grade as case 22            -- commit 77f5275
      - [x] 8c `ch21-parity`                            -- this commit
      - [ ] 8d `single-composition-root`   (needs call tracking in astscan)
      - [ ] 8e `agent-status-tool`         (needs fake-vendor harness)
      - [ ] 8f `per-model-cost`            (needs fake-vendor harness)
- [ ] 9. `scripts/ch22-mutants.sh`
- [ ] 10. Makefile `grade22`/`grade22-audit` + gradesweep entry

SCOPE CHANGE (Bill, mid-session): stage 6 does NOT restructure `ws.NewHub`.
See `docs/ensemble-topology-decisions.md`. Today's Hub becomes a GUIServer
and leaves the agent library, so wiring it into `common.Agent` now would have
the grader enforce a shape about to be inverted. Only its usage source was
fixed (one line, `cmd/main.go`). Collapsing `cliHost` into `NewAgent` stays
in scope and is more justified under the new design, not less: the flat root
is the embryonic Ensemble.

## Verified measurements (re-measure before print)

- ch5 scores 120/120 BOTH before and after the rename, evidence line moving
  from `common.Host` to `common.Agent` on its own. The old grep-based check
  would have failed on a tree whose structure improved.
- Agent module: 12 packages ok, 0 FAIL, ~6s. `go test ./internal/grade/`
  takes ~505s -- budget for it.
- `go list -deps ./internal/tools/` reports exactly two packages from this
  module: `common` and `tools`. No `llm`, no `ws`. The anti-cheat property
  for `reaches-through-the-chain` is true today and checkable by toolchain.


## EXHIBIT for the author: the decay stating its own motive

Removed from `agent/internal/common/interfaces.go` in stage 3/4. This is the
comment that sat above `RecordUsage`/`SessionUsage` on the hub's back-pointer
interface, verbatim:

> // RecordUsage adds one request's token counts to this run's totals, and
> // SessionUsage reads them back.
> //
> // These live on Agent rather than being handed around because Agent is
> // already the thing every object can reach through its parent chain, and
> // its lifespan is already exactly one run of the program — which is what a
> // session is. Anything that spends tokens can therefore report them
> // without being wired to a reporter, and anything that displays them can
> // read them without being wired to a producer.
> //
> // Embed UsageCounter to satisfy both.

(The comment said `Host` before stage 2's rename; `Agent` above is the
post-rename text.)

Why this is a better exhibit than the `hub.Model` closure: the closure is a
staple applied from OUTSIDE, at the wiring site, by someone who knew they
were working around something. This is the back-pointer interface accreting
an unrelated capability from the INSIDE, in a comment that is pleased with
itself. Every sentence is true. "Reachable" has quietly become the criterion
for where a capability lives, which is exactly how the parent interface stops
meaning anything — and it is the same root cause as the closures, wearing a
better argument.

## Stage 6 findings (measured, for the author)

The two composition roots were never duplicates. The library root was a
STRICT SUBSET: `cmd/main.go` stapled nine capabilities onto the engine that
`NewAgent` set none of -- Cache, Ctx, Journal, Target, ToolRoundLimit,
Memory, Bands, Recall, and the hub's Model closure. Chapters 15 through 18
existed only inside `main()`. `NewAgent` had not merely gone unused; it had
stopped being able to build an agent that could run.

The nine do not become nine constructor parameters. They collapse to ONE,
because each was reading the same hardcoded `"."`. That is what makes the
full collapse correct rather than arity growth.

THREE constructors, not two. `NewAgent` 0 callers, `NewWSHub` 0 callers, but
`NewBareAgent` has FOUR live callers (`cmd/virtual-user/main.go:111` plus
three tests) and is documented in chapter 13. The secondary binary uses the
library correctly while the main binary hand-wires. The flat root is not in
the simple case; it is in the important one. `NewBareAgent` also proves
single-phase construction was always sufficient -- virtual-user builds an
agent, then registers tools and connects MCP afterwards.

`runActorLoop` had THIRTEEN parameters. That is what a composition root looks
like when there is no object to hang anything on.

A dissolved constraint, preserved. The flat root's startup sequence read as a
necessary ordering: discover skills, sync model-gated tools, freeze
declarations, then build the engine. It stopped being necessary when chapter
10 made skills loadable at runtime -- `LoadSkill` re-derives the system
prompt AND the declarations, and `main.go` already re-synced declarations at
four later points. The root preserved an ordering that had not been required
for six chapters. Bill spotted this from the chapter content alone; it killed
a two-phase constructor split that was about to be built.

Per-agent identity as a process global. The defining skill was selected by
`envOr("EN_PRIMARY_SKILL", "ensemble")`, so two agents in one process could
not have had different primary skills either. Same disease as the data
directory, different axis.

`cliHost` was a pass-through wrapper around `Logger` plus a usage counter
nothing had written since stage 3 -- but it also created `api.log` and
`debug.log` in the working directory, so deleting it naively would have
silently stopped two logs.

## Pre-existing failure, NOT caused by this work

ch19 grades 85/100; one check fails on a missing `agent/events.jsonl`.
Verified pre-existing by `git worktree` at `273797c`, the commit BEFORE this
session's first, where it fails identically. Raise with the author
separately. (Memory also records ch8-ch12 at 80-90 pre-existing.)

## The thinnest possible exhibit: *Logger satisfies common.Agent

`common.Agent` is {Logf, APILogf, Debugf}. `*Logger` has all three, so a
LOGGER satisfies the AGENT interface. This session relied on that fact
without noticing: when `cliHost` was deleted, its replacement in
`cmd/tools_test.go` is literally `host := agent.DefaultLogger()`.

That is the measurement proving the interface never grew past the capability
it was named for. Five types satisfy it today:

    logger.go:49                    *Logger          (the base)
    agent.go:534                    *Agent
    internal/jobs/jobs.go:52        *Jobs            (delegates to its host)
    internal/llm/compact_test.go:81 testHost         (fake)
    internal/tools/statustools_test.go:23 fakeAgent  (fake)

DECIDED NOT TO FIX IN CH22: adding `Settings()` to `common.Agent` would make
the three remaining settings closures die, which is the thesis. But it forces
a meaningless `Settings()` onto `*Logger` and onto both test fakes, and it
would stop a logger being usable as a parent -- a change worth making
deliberately, with its own argument, not as a late addition to this chapter.

The closures are no longer staples in any case: they are created by the agent
over its OWN settings store inside `attachState`, not handed in from outside.
The remaining distance is one interface method, and naming that distance is
probably better teaching than silently closing it.

## Grader state (end of this session)

`go run ./cmd/grade -ch 22 ./agent` reports **50/100**, with four checks
implemented and passing and three reporting GRADER INCOMPLETE in those words.
That wording is deliberate: an incomplete grader must never be mistakable for
a failing tree.

    agent-builds               5/5   PASS
    back-pointer-chain        20/20  PASS   (framework-blind; 7 sharpness tests)
    reaches-through-the-chain 15/15  PASS
    ch21-parity               10/10  PASS
    single-composition-root    0/15  GRADER INCOMPLETE
    agent-status-tool          0/20  GRADER INCOMPLETE
    per-model-cost             0/15  GRADER INCOMPLETE

### What each remaining check needs

**single-composition-root (15).** Needs call-site tracking, which
`astscan.go` does NOT have today (one `*ast.CallExpr` at astscan.go:361, used
only by the chapter 5 walked-clause check). The property to assert, which is
clean and framework-blind: *the main package must not call any constructor
that the library's root constructor also calls.* Before this chapter's fix
`main` called `llm.NewEngine`, `jobs.NewJobs`, `tools.NewRegistry` and
`skills.NewSkillRegistry`, every one of which `NewAgent` also calls; after it,
`main` calls `agent.NewAgent` alone. Add a per-function set of `pkg.Func` call
targets to the scanner and the check is about twenty lines.

CAUTION: `NewBareAgent` is a SECOND legitimate library constructor with four
live callers (`agent/cmd/virtual-user/main.go:111` plus three tests), and it
is documented in chapter 13. The check must forbid a second flat ROOT, not a
second constructor.

**agent-status-tool (20) and per-model-cost (15).** Both need a fake-vendor
harness driving the real binary. `internal/grade/ch21_run.go` is the nearest
template. Give every scenario its OWN work directory -- chapter 21 lost time
to a shared workdir letting `save.json` resume the previous scenario.

For `per-model-cost`, assert BOTH that the reported figure equals
`sum(tokens_N * price_N)` AND that it differs from `total * current_price`;
and first assert those two formulas actually produce different strings for the
inputs chosen, or the test proves nothing. Use a tolerance: `CostByModel` sums
over a map, so float addition order varies between runs.

Unit-level versions of both already exist and pass, in
`agent/internal/common/usage_test.go` and
`agent/internal/tools/statustools_test.go`. The grader versions must drive the
binary rather than re-test the library.

## Stage 8d-8f COMPLETE — grader is 100/100 (7 of 7 checks)

Commits: `e714350` (8d single-composition-root), `6f62793` (8e/8f behavioural).

### 8d single-composition-root (15 pts) — `internal/grade/ch22_root.go`

Two name-blind clauses:
1. ONE ROOT PER BINARY. A main package may call at most one composition
   root, and -- if its own reach into the library is >= the threshold --
   must call at least one.
2. NO REBUILDING. A binary must not take an object from a package the root
   already builds from and then assign its fields.

A composition root is found with NO name matching: a function reaching >= 3
of the tree's own packages, directly or through helpers in its own package
(`treeFanout`, transitive within the package). `astscan.go` gained
`imports`/`pkgNames` maps, `callsIn`, `findFunc`, `treeFanout`,
`varsFromTreeCalls`, `fieldAssignsOn`, `line`.

MEASURED separation (TestCompositionRootFanout, asserts it survives):
least-reaching root = 4; greatest-reaching of the 29 other constructors the
binaries call = 1. Nothing sits at 2 or 3 — the threshold is in an empty gap.
Roots found: `agent.NewAgent`, `agent.NewBareAgent` (different binaries).

TWO FINDINGS THAT CHANGED THE DESIGN:
- "main sets a field on a library value" is NOT a decay signal by itself.
  Filling in a config struct before handing it to a constructor has the
  SAME SHAPE as stapling a capability onto a live engine. Only types
  separate them and this scanner has none. Both shapes are in the reference
  tree (`cfg.SystemPrompt` in virtual-user is legitimate; `hub.Model =
  func(){}` is the GUIServer closure Bill ruled out of scope). Hence clause
  2 is restricted by PROVENANCE to packages the root itself builds from,
  excluding the root's own package.
- THE CHECK PASSED THE REAL DECAYED TREE at first. Ten fixtures missed what
  one real tree caught. The draft counted roots tree-wide, and
  cmd/virtual-user's healthy `NewBareAgent` call satisfied the count on
  behalf of the decayed `cmd/` — the broken binary was not the one being
  counted. Now per-binary. Permanent fixture: "a healthy second binary
  masks a decayed first". Verify with:
      git worktree add /tmp/ch22-decayed 273797c
  The decayed tree must report: "cmd reaches 12 of the library's packages
  ... but calls no composition root".

### 8e/8f behavioural (20 + 15 pts) — `ch22_harness.go`, `ch22_status.go`

One driven session serves both. Builds the binary, runs 4 turns against
`fakevendor`, switches model mid-session over the GUI websocket, recovers
the `agent_status` output from the transcript the agent sent BACK to the
vendor (a tool result rides in the next request — no instrumentation).

- WS is CLIENT-PULL: a fresh connection is sent NOTHING until it sends
  `{"type":"subscribe"}`. Then `{"type":"update_settings","settings":{"model":...}}`.
  Endpoint `/ws`; the envelope field is `"type"`, NOT `"kind"` (stdin uses
  `"kind"`). There is no stdin verb for the model and settings.json is read
  once at startup, so the socket is the only seam.
- THE SWITCH IS ASYNCHRONOUS and a stdin prompt overtakes it: the turn
  after the switch is sometimes still billed to the old model. Fixed by
  reading each turn's model OFF THE WIRE (`ch22RequestModels`) instead of
  assuming. Removes the race; cost is still predicted independently.
- TOKEN COUNTS HAVE A CEILING: an earlier draft used millions and the
  driven agent decided its context was full and withdrew every tool but
  micro_handoff. Checks still passed — the grader had started testing
  compaction policy instead. Now 10k/2k/1k.
- Models are PRODUCTION rows `claude-opus-4-6` (5/25, read 0.5) and
  `claude-sonnet-5` (3/15, read 0.3): same vendor (wire shape unchanged)
  and both PRICED. Every `*-course` row is unpriced, and an unpriced pair
  cannot demonstrate a pricing bug. Prices are hardcoded — internal/grade is
  in the ROOT module and cannot import the agent module's internal packages.
- Cost check asserts correct AND not-flat-rate, but only after confirming
  the two formulas disagree for the session actually driven; otherwise it
  reports the GRADER incomplete rather than awarding points.
- Shipped-SKILL.md check is separate (`ch22StatusToolDeclared`) because the
  grader supplies its own skill file and is therefore blind to ch21's
  dead-on-arrival fault.

### Remaining: stages 9, 10, 11

## Stage 9 IN PROGRESS — `scripts/ch22-mutants.sh` written, NOT committed

Script exists and runs. Harness (guard, run_mutant, JSON check-ID extraction,
compile gate, landed-verification) is copied from `scripts/ch21-mutants.sh`
and WORKS. Six mutants; `agent-builds` deliberately has none (a build failure
is not a deletion of a taught behavior) — rationale is in the script header.

### Mutant status, measured — ALL SIX KILL EXACTLY THEIR CHECK

Last full run after the ws fix: **6 PASS, 0 FAIL**. Every mutant kills the one
check it names and nothing else.

| # | target check | status |
|---|---|---|
| 1 | back-pointer-chain | PASSES |
| 2 | reaches-through-the-chain | PASSES — fixed (import placement) |
| 3 | ch21-parity | PASSES — fixed (retargeted) |
| 4 | single-composition-root | PASSES — fixed (`spec` not in scope) |
| 5 | agent-status-tool | PASSES — was the panic, now fixed |
| 6 | per-model-cost | PASSES — fixed (unused range var) |

What each fix turned out to be:

- **M2**: `statustools.go` already has an `import (...)` block, so inserting a
  new import after `package tools` pushed the real block below a declaration →
  "imports must appear before other declarations". Fixed by inserting the
  import INSIDE the existing block. Verified: `llm.Engine` is a struct and
  there is **no import cycle** — the mutant builds clean.
- **M3**: the old mutant edited `agent/skills/web-search/SKILL.md`, which
  ch21's grader never reads (it supplies its own skill file for determinism).
  The mutation landed on a file no check looks at, so it survived. Retargeted
  to `case "url":` in `agent/cmd/main.go` — the comment on that line calls it
  "the whole host-side cost" of ch21. Retiring the label sends URL servers to
  the default unsupported-transport branch: the transport is still fully
  implemented and simply never selected, which is ch21's own lesson restored
  as a fault.
- **M4**: `spec` never existed — the real call passes an inline
  `agent.AgentSpec{...}` literal at `agent/cmd/main.go:287`. Fixed with a
  never-called `func mutantSecondRoot()`. **Unreachable is deliberate**: the
  check is static and counts root call sites per main package regardless of
  reachability, while a *live* second `NewAgent` would also build a second
  agent at run time and redden the behavioral checks. A mutant that kills more
  than it names is a defect by this script's own doctrine.
- **M6**: pinning the lookup to a literal leaves the loop's `model` variable
  unused, which Go refuses to compile. Fixed by also rewriting the range to
  `for _, u := range byModel` — two textual edits, one deleted behavior.

### M5 is NOT check coupling — it is the panic, proven

I twice reasoned my way to "the `agent-status-tool` and `per-model-cost`
checks are coupled." They are not. Evidence, from the run above:

- m5's JSON reports both checks failing with **"agent produced no assistant
  reply"** — the crash signature, not a parse or assertion failure.
- `m5.err` is the **only** one of the six stderr captures containing
  `panic: send on closed channel` (exactly 1 occurrence; all others 0), and
  it is 643 bytes larger than its siblings — that delta is the stack trace.

The script already redirects each grader's stderr to `$OUT/mN.err`. **Read
those files before theorising about a mutant result.** The evidence was
sitting on disk through three wrong diagnoses.

Consequence: the audit cannot be trusted to a clean 6/6 until the panic is
fixed. Every driven-grader invocation carries ~17% chance of a spurious
double-failure, and a re-run that happens to come up clean is not a fix.


## THE REAL BLOCKER — the driven session is ~17% FLAKY

MEASURED: 1 of 6 clean grader runs fails; 3 of 6 in another sample. When it
fails, BOTH `agent-status-tool` and `per-model-cost` fail together with:

    "the session did not run: agent produced no assistant reply"

So `out.replies` is EMPTY — turn ONE never got a reply. This is a startup /
timing race in `ch22DriveSession`, NOT a check defect and NOT the model switch.

**A grader that fails honest work 1-in-6 cannot ship. Fix this before stage 10.**

Ruled out by measurement, do NOT re-investigate:
- NOT the save.json-shared-workdir bug: `ch22_harness.go:186` uses
  `os.MkdirTemp` per run with `defer os.RemoveAll`.
- NOT a model-switch race: `ch22SwitchModel` ALREADY waits for the agent to
  echo the new model back over the websocket before returning
  (`ch22_harness.go` ~334-343). I added a bounded drive-until-switch loop to
  "fix" this and it was both unnecessary and HARMFUL — see below. Reverted.
- NOT check coupling between agent-status-tool and per-model-cost. The single
  observation suggesting it was this same flake.

**DO NOT drive extra turns.** `ch22_harness.go:107`
`turns := []ch22Usage{ch22TurnOne, ch22TurnTwo, ch22TurnThree}` means the
EXPECTED figures assume exactly three turns, and the fake vendor's `replies`
list (~line 177) is finite. Extra prompts exhaust it and produce the very
"no assistant reply" symptom being chased.

Next diagnostic step: the harness does not appear to capture the child's
stderr (`grep -n Stderr internal/grade/ch22_harness.go` — see output above).
Capture it to a buffer and print it when no reply arrives; that will say
whether the process exited, the fake vendor 500'd on a port race, or the
first prompt was written before the actor began reading stdin.
`ch22WaitPort` returning only proves the LISTENER is up, which is weaker than
"the actor is ready to accept a prompt".

## Stage 10 DONE — plumbing wired, FULL SWEEP GREEN

Committed `9199ea6`: Makefile `grade22` / `grade22-audit` (the latter marked
BATCH in a comment — six mutants, each a full driven-binary grader run), plus
`run 22 ./solutions/ch22` in `scripts/gradesweep.sh`.

Verified before wiring: `git diff HEAD:agent HEAD:solutions/ch22` is EMPTY, so
the stage-7 snapshot has zero tracked drift and the new sweep line targets a
tree that really passes. `solutions/ch22` grades 100/100 on 7 of 7.

**Full cross-chapter sweep: exit 0, every chapter 100/100.** This was the real
gate for this chapter: stage 1 replaced ch5's fake `detectLogf` with the
structural `detectParentChain`, and via the parity cascade that new check runs
against every later snapshot. Nothing broke. The `common.Host`→`common.Agent`
rename across 15 files likewise propagated cleanly.

One non-100, NOT ours: `ch19 ./agent` = 85/100. `ch19 ./solutions/ch19` scores
100 on the same run, which localises it to the missing `agent/events.jsonl`
already proven pre-existing at `273797c`.

### Defect found by the sweep: ch14's harness leaks agents

Every sweep ends with a `WARNING: N leaked agent process(es) still running`,
and every leaked process is ch14's — each carries a
`--tts-log .../ch14-work-<random>/tts.log` flag, and each sits in its OWN
`ch14-work-*` directory. So ch14 leaves one agent behind per scenario, and
they survive the grader's exit.

Measured: **5 leaked per sweep** (one sweep → 5; three sweeps → 15).
Deliberately not stated per ch14 invocation: the sweep grades ch14 twice
(`./agent` and `./solutions/ch14`) yet only five leak, so the per-invocation
rate is NOT simply five and has not been pinned down.

Reaped by hand each time. Beware when cleaning up: an unrelated long-lived
`./ensemble --port 8084` was running throughout, and must not be matched.
Filter on the harness's temp dir, not on the binary name.

Not fixed here — a ch14 harness bug, outside ch22's scope, worth raising on
its own.

## The panic is FIXED (`352b590`), and the fix is ch22 material

Diagnosis: both sides of the race mutate shared state under `h.mu` and then
act after releasing it. `Observe` snapshots the live clients, unlocks, and
only then sends. Teardown deleted the client, unlocked, and only then closed
`c.send`. So a client could be snapshotted, deleted, closed, then sent to.

The count is the exhibit: **thirteen call sites send on `c.send`; exactly one
closed it; and that one was not a sender.** The `select`/`default` on most of
those sends looks like protection and is not — `default` saves a sender from
a FULL channel, never from a CLOSED one. The defensive-looking code gives no
defense against the failure that actually occurs.

Fixed by ownership rather than by narrowing the window: nothing closes
`c.send` at all, and `writePump` stops on a new `done` channel (draining
what is already buffered first). An unreferenced channel is garbage collected
whether or not it was closed. `sendReplay` selects on `done` too, so a replay
no longer burns its full 5s timeout on a socket that has gone.

Teardown was extracted from the tail of `ServeWS` into `Hub.removeClient` so
the test drives the real path. It also drops MCP agent registrations —
behavior a hand-copied teardown in the first test had already silently lost.

### The regression test lied twice before it worked

`TestObserveDuringTeardownDoesNotPanic` is proven sensitive: restoring
`close(c.send)` panics it. Two earlier versions **passed with the bug
restored**, and both are recorded in the test's own comments:

1. 64 clients torn down in one goroutine — the teardown loop finished before
   the senders were scheduled, so `Observe` saw an empty set. 8000
   observations in 8ms, green.
2. 5000 rounds of a single client — the gap between `Observe`'s unlock and
   its send is a few nanoseconds, and teardown never once landed inside it.

What works is a LARGE snapshot: the send loop itself becomes the window,
because `Observe` is still working through early clients while teardown is
closing later ones. **Widen a race window structurally; do not hope for it.**

Also worth keeping: `Observe` early-returns when `marshalObservation` yields
nil, so an observation the switch does not handle never reaches the send loop
at all. A test using one would pass while exercising nothing.

## Remaining: nothing blocking — ch22's grader work is complete

Stages 0–11 are done. Final state: `grade22` = 100/100 on 7 checks for both
`./agent` and `./solutions/ch22`; `grade22-audit` = 6/6; snapshot re-taken at
`3a6ef2b` with zero drift; full sweep green.

Still open, and NOT ch22's scope:

- **ch14's harness leaks five agents per run** (see above). Its own bug.
- **ch19 `./agent` = 85/100**, pre-existing at `273797c`, missing
  `agent/events.jsonl`. `./solutions/ch19` scores 100.
- **Two authorial calls on the panic**, which are Bill's, not mine:
  whether to backport the fix to the frozen `solutions/ch21` (students start
  ch22 from it and would otherwise meet a ~1-run-in-6 crash through no fault
  of their own), and how ch22 narrates the bug. The fix itself landing in
  `./agent` was required under all three options considered, which is why it
  did not wait.


## !!! CORRECTION — the "flake" is a REAL PRODUCTION BUG, not a grader defect

Captured on run 4 of 6 by NOT discarding the child's stderr (`cmd.Stderr =
os.Stderr` at `ch22_harness.go:206`, so `2>/dev/null` was hiding it — that
redirect is why three earlier diagnoses were wrong):

    panic: send on closed channel

    ws.(*Hub).Observe(...)            agent/internal/ws/handler.go:211
    llm.(*Actor).notify(...)          agent/internal/llm/actor.go:124
    llm.(*Actor).setState(...)        agent/internal/llm/actor.go:110
    llm.(*Actor).runTurnLoop(...)     agent/internal/llm/actor.go:228
    llm.(*Actor).handleUserMessage    agent/internal/llm/actor.go:221
    llm.(*Actor).handleRequest        agent/internal/llm/actor_runtime.go:147

**Diagnosis.** The grader connects a websocket client, switches the model,
and disconnects. The Hub closes that client's channel on disconnect, but the
Actor goes on delivering observations to it, and `Hub.Observe` sends on the
closed channel. The whole process panics, so turn one never produces a reply
— which surfaced as the misleading "agent produced no assistant reply".

**This is a race in shipped agent code, not in the grader.** It needs a real
owner: whoever makes the channel unusable must stop the sender first, or
`Observe` must not own the lifetime of a channel it does not close. The
pattern is the ch4 deadlock rule inverted — there the rule was that whoever
makes a condition true owns waking the waiters; here, whoever closes a
channel owns stopping its senders.

**Do NOT "fix" this by having the grader keep its websocket open.** That
hides a crash any GUI client triggers by closing a browser tab mid-turn. The
grader found a genuine defect; that is the grader working.

**This also means the mutation audit's earlier results are suspect.** Mutant
5's "extra kill" of per-model-cost was almost certainly this panic, not check
coupling. Re-run `scripts/ch22-mutants.sh` after the panic is fixed and
re-read every mutant's kill set from scratch.

**Raise with Bill before fixing**: ch22's thesis is architectural decay, and a
crash on client disconnect in the component that is being moved out of the
library (`ws.Hub` → GUIServer) may belong in the chapter as an exhibit rather
than being silently repaired.
