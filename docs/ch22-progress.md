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

### Mutant status, measured

| # | target check | status |
|---|---|---|
| 1 | back-pointer-chain | **PASSES** — kills exactly its check |
| 2 | reaches-through-the-chain | does not compile — FIXABLE, see below |
| 3 | ch21-parity | SURVIVED — wrong mutant, see below |
| 4 | single-composition-root | does not compile — FIXABLE, see below |
| 5 | agent-status-tool | kills its check; the extra kill was FLAKE, not coupling |
| 6 | per-model-cost | does not compile — FIXABLE, see below |

Fixes, all diagnosed, none yet applied:

- **M2**: `statustools.go` already has an `import (...)` block, so inserting a
  new import after `package tools` puts a declaration before it →
  "imports must appear before other declarations". Insert the import INSIDE
  the existing block and append `var _ = mutantllm.Engine{}` at END of file.
  `llm.Engine` IS a struct, and there is no import cycle.
- **M4**: anchor `host := a` is right, but `spec` is not in scope there. The
  real call is `agent.NewAgent(cfg, agent.AgentSpec{DataDir: ".", SkillDir:
  skillDir, Skills: []string{primaryName}, LogPath: logPath, SavePath:
  savePath})` at `agent/cmd/main.go:287`. Use
  `_, _ = agent.NewAgent(cfg, agent.AgentSpec{DataDir: "."})` — a composite
  literal with a subset of fields compiles.
- **M6**: substituting `LookupModel("claude-opus-4-6")` leaves the range
  variable `model` unused → compile error. Append `_ = model` on the next line.
  The mutation itself is the right one: it is the historical flat-rate bug
  exactly, and changes no number in the report, only which price sheet is
  consulted.
- **M3**: ch21-parity does NOT read the shipped `agent/skills/web-search/SKILL.md`
  (ch21's grader supplies its own skill file — that is documented and
  deliberate), so renaming `firecrawl_search` there changes nothing. Pick a
  mutant in ch21's CODE path instead: break the Streamable HTTP transport in
  `agent/internal/mcp/http.go` (e.g. the `case "url"` dispatch ch21 added).

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

## Remaining: finish stage 9, then 10, 11

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
