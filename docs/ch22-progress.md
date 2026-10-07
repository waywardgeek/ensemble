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
