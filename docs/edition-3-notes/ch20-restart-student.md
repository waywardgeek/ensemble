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


## Bill-directed production R7/10 — library-backed schema validation HOLD

**R7 source is stable for root and the retained reviewer.** This is the same student/post-closure correction, not a reset. Earlier notes and evidence above remain unchanged. Receipts: `/private/tmp/ch20-r7-schema.F6qzQC`.

### Change and ownership

`internal/common/submission.go` now delegates schema keyword-value checking and instance semantics to **github.com/santhosh-tekuri/jsonschema/v6 v6.0.3**. The old handwritten evaluator, type matcher, equality/hash implementation and meta-schema checker were removed. No parallel validator, global cache, new completion owner, or lifecycle/GUI redesign was added.

The existing owner retains bounded `UseNumber` JSON decoding, explicit object-root/subset admission, exact public raw/canonical byte/node/depth/numeric limits, byte-product checking, and default-versus-custom normalization. Compilation uses an in-memory URN, explicit Draft 2020-12, the library’s default **Go regexp** engine, and an empty URL-loader map. Unsupported keywords/resources remain refused. Original/canonical schema bytes are not rewritten by compiler adaptation; returned payload/schema bytes remain owned copies. Library structured errors supply the already tested `$.answer: violates type` prefix, with the complete library error retained as the cause.

Two dependency differences were confirmed in the isolated probe and handled as Cadence explicitly authorized:

- Draft 2020-12 library admission accepts empty/duplicate enums. A small **library-validated** array constraint (`minItems:1`, `uniqueItems:true`) enforces our stricter enum admission, including exact numeric/object/array equality. There is no custom equality routine.
- v6.0.3 converts length/item/property bounds to machine integers and can wrap huge legal JSON integers. On the private decoded compiler input only, nonnegative integer bounds above **1,048,577** are saturated to that ceiling. No bounded payload can reach that many characters/items/properties, so this preserves min/max semantics without overflow. Canonical/durable schemas keep their original exact numbers. Invalid bounds are still rejected by library meta-validation.

Runtime inspection shows the validator and `golang.org/x/text`; `regexp2` is **not** a runtime import. Its checksum entries come from the dependency’s tests. All dependencies were resolved offline from the existing cache.

Changed paths, all under `solutions/edition-3/main`:
- `internal/common/submission.go`
- new `internal/common/ch20_library_submission_test.go`
- `go.mod`, `go.sum`
- `gui/go.mod`, `gui/go.sum`
- `examples/ch13-speech/go.mod`, `examples/ch13-speech/go.sum`

No existing test assertion was weakened or edited. Durable-row writer/reader, fleet/actor completion, catalogue/collision/application-submit code, launcher and browser assets are unchanged. `.ch15-work/` and `.ch16-work/` were left untouched.

### Commands, actual results, and preserved evidence

`Q=/private/tmp/ch20-r5-validation.vTRrtJ/check`: existing `env -i`, owned canonical HOME/TMP, existing caches, controlled PATH, dummy keys, GOMAXPROCS=2, readonly serial modules, GOTOOLCHAIN=local, GOPROXY/GOSUMDB off. Go checks were sequential.

Root cwd: `solutions/edition-3/main`.

- `Q go test -race -vet=off -run '^TestCh20(SchemaContract|SubmissionBudgets|Round2SchemaBoundaries|AuthoritativeSubmissionRows|LibrarySubmission)' -count=1 -v ./internal/common` — **exit 0**, `schema-durable-race-01.txt`. Includes existing semantic/resource/whole-file row cases and new enum, oversized-count, Go-regexp, no-fetch and ownership cases.
- `Q go test -race -vet=off -run '^TestCh20(FleetActualAgentSubmit|CustomSubmitCorrectionAndPersistenceFailures|ApplicationSubmitOwnershipAndRecovery|ApplicationCatalogueBoundaries|AcceptedCompletionFactRecovery|CustomRowWriterAndRecovery|SubmissionAppendRefusesInvalidFile|RootSubmitQueueAndRecovery|Round2TaskToolAndComplete)$' -count=1 -timeout=90s -v .` — **exit 0**, `agent-submit-recovery-race-01.txt`. The `CustomRowWriterAndRecovery` selector did not match a test; it is not counted as evidence.
- Exact existing R5 row case: `Q go test -race -vet=off -run '^TestCh20CustomDurableRowAndRecovery$' -count=1 -timeout=60s -v .` — **exit 0**, `custom-durable-recovery-race-01.txt`.
- `Q go test -vet=off -run '^$' ./...` in root, then `gui`, then `examples/ch13-speech` — **all exit 0**, respectively `root-compile-01.txt`, `gui-compile-01.txt`, `listener-compile-01.txt`. These are compile-only, not behavior results.
- Offline `go mod tidy`: `root-dependencies-01.txt`, `gui-dependencies-01.txt`, `examples-ch13-speech-dependencies-01.txt`, all **exit 0**.
- `root-runtime-dependencies.txt`, `dependency-source-inspection.txt`, and `final-inspection.txt` record dependency membership, exact changed paths, absence of superseded engine symbols, file hashes, clean scoped diff checking and over 13 GiB free.

Probe evidence is retained rather than discarded: `library-probe-01.txt` **exit 1** was an isolated-module missing-`go.sum`-entry setup failure, as root directly verified from the saved output; the separate writable-copy/cache setup repair was not a product failure. `probe-cache-setup.txt` **exit 0** used cache-only dependency metadata. `library-probe-02.txt` **exit 0** demonstrates the dependency differences (enum acceptance and overflowing count semantics), not a product verdict. Focused product checks had no failures.

### Limits and handoff

The preserved tests cover default normalization, explicit custom payload/schema ownership and exact numbers, budget boundaries, custom correction feedback, application-local submit, genuine fleet/`Complete` terminal behavior, full/dialogue-only recovery, malformed later rows, append refusal and save failures. They do not constitute a full suite, grade, broad race campaign, or exhaustive schema conformance proof. The two bounded dependency adaptations are explicit review points; no unresolved tested contract mismatch remains.

No native/provider/browser/live campaign, manuscript, grader/private review access, snapshot, commit, install/restart, shared cleanup, or sub-agent was performed. Root owns the next review/release decision. No further source edits after this HOLD without authorization.
