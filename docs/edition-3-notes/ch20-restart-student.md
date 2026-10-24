# Chapter 20 — dispatched join / human hint test-only HOLD

**Production remains frozen at R6/10; no R7 correction round.** Added one focused test and this note only. The browser HOLD is preserved at `/private/tmp/e3-ch20-student.2X1OuS/r6-browser-hold-before-hint-join.md`; all evidence at `/private/tmp/ch20-r6-reuse.cBgKMg` and the earlier R5 stop remains untouched.

## Test and exact ordering

Added `solutions/edition-3/main/ch20_dispatched_join_hint_test.go`, `TestCh20HumanHintDuringDispatchedJoin`. It reuses `ch20Rig`, `ch20Reply`, the actual Agent actors/engines, the actual `join_agents` handler, the existing job owner and normal tool-based child submit. No production hooks, fake job table, sleep barrier, or new framework.

The child reaches a channel-held loopback provider response. The parent model requests `tool_limits` (60-second observation) followed by `join_agents` for that child. Before posting `ensemble.Hint`, the test observes `ToolDispatched`, finds the canonical dispatched release with its real job handle, verifies that owned job is running, and rejects an already-returned join observation. The test context is 20 seconds, with a 40-second Go test deadline.

Observed single-run proof:

1. **Release sequence 10, turn 1, handle 1**, `join_agents` dispatched and running; child provider still blocked.
2. Ordinary human hint is admitted at **sequence 11**. `HintObserved` identifies sender `human` and the exact hint text.
3. Join observation returns at **sequence 12**, still **handle 1 / running / non-error / dispatched**, not killed or treated as skipped-before-dispatch. The parent’s next real model request includes the hint, and its ordinary turn completes while the same join job and child remain running.
4. Only then does the fixture release the child’s provider response. The child performs genuine validated `submit`; retained outcome is `submitted` with the expected structured payload.
5. `parent.Jobs().Get(1)` returns the same job object. Its actual terminal `Complete` report is `done`, non-error, and contains:

```json
{"outcomes":[{"agent_id":"controlled-child","parent_id":"root","state":"submitted","submission":{"count":1,"result":"child finished after hint"}}]}
```

There were **2 parent and 1 child loopback requests**. The test body completed in **0.31 seconds**. Cleanup releases the controlled response and shuts down the existing owned fixture on all paths.

## Commands and receipts

Root module cwd: `solutions/edition-3/main`. `Q=/private/tmp/e3-ch20-student.2X1OuS/check` is the existing isolated `env -i` wrapper: owned canonical HOME/TMP, existing caches, controlled Go/Delve PATH, dummy keys, GOMAXPROCS=2, readonly serial modules, local toolchain and proxy/sumdb off.

- **One race run:** `Q go test -race -vet=off -run '^TestCh20HumanHintDuringDispatchedJoin$' -count=1 -timeout=40s -v .` — **exit 0**. Receipt `hint-join-race-01.txt` records all three ordering milestones and the actual terminal job output.
- **Affected package compile only:** `Q go test -vet=off -run '^$' .` — **exit 0**, no behavior tests run. Receipt `hint-join-compile-01.txt`.
- `hint-join-test-diff.txt`: complete added-test diff (no-index diff status 1 means a new file, not a failure).
- `hint-join-final-inspection.txt`: results, clean scoped whitespace check, preserved browser evidence and free-space check (over 13 GiB).
- Test SHA256: `c44f8e065cea961afe90e39b86d6320f9a94e04b34d40b59287e6f9dbbd32fd7`.

All receipt filenames above are under `/private/tmp/e3-ch20-student.2X1OuS`. No behavioral or race failure occurred in this release.

## Limits / closure

This proves the precise **local** ordering missing from the earlier evidence: dispatched/running real join → ordinary human hint → nonterminal observation wake → same-handle eventual completion. It does not retroactively prove that timing in the prior successful live run, and it does not claim coverage of every scheduler interleaving.

No production, browser, real-provider, GUI, native/device, broad suite/gate/grade, installation/restart/8088, commit/snapshot, manuscript or coding-worker action occurred. Return stable test-only HOLD for root/retained-reviewer diff check; no further edits or runs without authorization.
