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
- [ ] 6. Collapse composition roots (cliHost -> NewAgent)
- [ ] 7. Snapshot `solutions/ch22`
- [ ] 8. Grader `internal/grade/ch22_*` (7 checks / 100 pts)
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
