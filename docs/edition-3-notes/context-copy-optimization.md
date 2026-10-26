# Context-copy optimization branch review

This independent feature starts from `origin/main` at `123d6017`. It ports the
approved narrow History reads and request-preparation snapshot reuse from
`37e9b0b3`, without timing instrumentation or other local features. JSON copying
remains unchanged for full snapshots and independently owned single entries.
Chapter 2 and the combined book carry the same bounded explanation.

## Validation

- Focused race checks for narrow reads, context, memory, compaction, hints and
  publication pass. Core and GUI vet pass; the GUI short suite passes with the
  installed Playwright dependency and `CHROME_BIN` configured.
- The real local-model Firefox run executes one tool, displays its final answer,
  and preserves it across Agent selection, reload and restart, with zero
  JavaScript errors. No timing UI is introduced. Temporary receipt:
  `/tmp/ensemble-live-context-optimization-mqGMUc/summary.json`.
- The synthetic roughly 1MiB real-actor/local-HTTP benchmark has median turn
  costs of 279.06ms before and 60.74ms after, with 13642 versus 2718 allocations.
  From `agent/`, reproduce with `go test ./ensemble -run '^$' -bench '^BenchmarkContextSnapshotTurn$' -benchmem -benchtime=5x -count=3`.
  For the before comparison, use that same benchmark fixture on unchanged
  `123d6017`; the upstream tree does not already contain the fixture.
  These measurements are not a promise about real provider latency.
- Goodall independently accepts the code with the test exception below and
  accepts the prose. Paired book additions match, and the preface and frozen
  source trees are unchanged.

## Broad-suite failures retained

The whole core suite is not green. The CLI human-escalation check also fails on
the unchanged upstream baseline. `TestKillWakesWaiterAndKillsGroup` fails at
`agent/ensemble/jobs_test.go:215`, where `kill(pid, 0)` treats PID existence as
proof that a child is running. Diagnostics found a dead zombie (`Z`); independent
syscall tracing showed both shell and child terminated by SIGKILL. The orphan
was subsequently reaped by its systemd adopter. The traced
test passed. Baseline repetitions passed while optimized repetitions failed,
so the failure frequency is not claimed unchanged or baseline-reproduced.

Independent review accepts this assertion/reaping limitation, not a live child
surviving termination. No process-control code or existing assertion is changed
by this feature. Retain the failed broad-suite results when presenting the PR;
focused and GUI passes do not establish an entirely passing core suite.
