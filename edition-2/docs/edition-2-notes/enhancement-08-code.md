# Edition 2 Chapter 8 grader review

> Historical paths: the `edition-2/ch04`–`ch09` paths recorded below use
> original student numbering. Those exports are now preserved under
> `solutions/edition-2/history/`; see the [solution mapping](../../solutions/README.md).

Coder: `enhancement_coder_resumed`, October 10, 2026. Accepted at initial formal
combined review, zero correction rounds; see [review](enhancement-08-review.md).
Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`.

## Scope and provenance

This is the final chapter of Bill's bounded pass: original Chapter 9 becomes
Edition 2 Chapter 8. I read the full original, carryover, skill, enhancement plan,
student notes, comparative review, original checks/launcher and accepted/legacy
settings paths. Original chapter revision is
`2b2fedbc8b50bc688c31ab823712553e589f8a3b`, SHA-256
`3693cdbb54b6ce66698f635586b17973b9f4a707f8a51c3898b17e2d3ee3349b`.
Accepted endpoint is `solutions/edition-2/ch09` (`edition-2-ch09-r2`) with its
separate `gui` module; legacy is the actual `solutions/ch09` module and `cmd`
entry point. No original grader, first-edition chapter, runtime or frozen export
changed. No mock, paid call, later feature or new settings API was added.

Coordinator and reviewer approved the narrow plan before edits. The new
`internal/gradeedition2/ch08.go` reuses the four original `Ch9Evaluate` predicates
and 25/20/20/15 weights, plus existing real GUI process/WebSocket helpers. Its
fifth check requires every corrected Chapter 7 check to pass for 20 points.
Separate GUI/core entry points and public-dependency proof carry forward. The
known wrong-directory original CLI cascade is not repeated.

Invoke `go run ./cmd/grade-edition2 -ch 8 path/to/core`. No submission argument
selects `solutions/edition-2/ch09`; `-gui DIR` overrides `gui` below that core.
Existing consumer/workflow/media/stream-exercise/model flags retain their
meaning. No new flag was needed. Historical declared GUI/streaming model fixture
`claude-sonnet-5` remains overridable; it is not a current availability claim.

## What the checks establish

The exact ordered patches are:

1. `theme: light`, `tts_enabled: true`, `tts_speed: 1.5`, `font_size: 18`.
2. `theme: dark`, `tts_speed: 2`.
3. `tts_enabled: false`, `max_tool_rounds: 2`.
4. `max_tool_rounds: 0`.
5. `theme: system`.

The first two preserve original light/speed-1.5 round trip and dark/speed-2
broadcast expectations. The others exercise omissions, explicit false and
meaningful zero through the existing budget field. These are valid inputs,
not prescribed defaults or clamp endpoints. Initial subscriptions require all
four visible preferences with their JSON types. Each sender echo and second
client receipt is compared to known fixture values accumulated independently
of replies. Extra unrelated settings are allowed.

A fresh subscriber must receive `system, false, 2, 18, 0`. Then clients close,
the real server stops, and a newly launched process reads the same disposable
working directory; its subscription must return the same five values.
Nondefault speed/font and false distinguish persistence from startup defaults.
The original theme-only restart check could pass after saving default dark.
Ordered receipts associate each patch with its own broadcast, without adding
an acknowledgement or transport protocol.

Numeric checks establish wire snapshots and restoration. Effective budget
execution, entry/load validation, write-failure publication, browser semantics
and native speech remain required source/integration obligations below.

## Actual results

| Submission and grader | Actual result | Meaning |
| --- | --- | --- |
| Accepted student, historical original | **80/100 FAIL** | Settings earned 80; inherited parity failed. |
| Legacy, historical original | **100/100 PASS** | Preserved historical weaker-check result. |
| Accepted frozen endpoint, revised full run | **100/100 PASS** | Strengthened settings and corrected Chapter 7 parity pass. |
| Legacy endpoint, revised full run | **0/100 FAIL** | All four settings checks and corrected inherited parity fail. |

Revised scores are complete runs, not composed scores. Raw reports are
`accepted-revised08.json` and `legacy-revised08.json`; both stderr files are
empty. Legacy settings failures show missing `tts_enabled: false` in sender,
other-client, fresh-subscription and restarted snapshots; initial snapshots
also lack all four correctly typed visible preferences. The comparison stops
at the first difference per snapshot, so these messages alone do not separately
prove every omitted field, including zero.

Legacy parity separately reports reused final artifact identity, lost replayed
content/identity, single-module GUI/public-core boundary, unknown-model fallback,
earlier tool checks, media and external-example obligations. Its default
workflow location has no separate module; the selected legacy consumer path
does not produce an executable consumer, and the streaming exercise lacks the
separate-module boundary. These are missing/inapplicable submission setup, not
proof that an unexecuted external workflow is broken. Default media identifiers
also mismatch parts of its historical table; Gemini reports its actual
unresolved local-path refusal. Exact nested details remain in the raw report.
No invented adjacent app, compatibility adapter or forced passing score was
supplied.

The historical accepted original run took approximately 46m59s, estimated from
output timestamps, while inherited CLI exercises targeted the continuing GUI
server. Its build predates CSS/test/documentation changes and the later HTML
speech fix. Reusing the actual result does not claim a post-R1 original run.
See [student notes](ch09.md) and [comparative review](ch09-code-review.md).

## Validation and targeted defects

The accepted settings baseline passes 80/80. Eleven isolated mutations of actual
accepted GUI source were built and exercised through the new driver:

| Defect | Intended observed failures |
| --- | --- |
| `omitempty` on false TTS snapshot | All four: false disappears; restart can restore true default. |
| `omitempty` on zero budget snapshot | All four: explicit zero absent. |
| Force TTS true during apply | All four: false ignored. |
| Change zero budget to two | All four: explicit zero becomes two. |
| Decode sparse patches over empty settings | All four: omitted preferences reset. |
| Corrupt broadcast font to 19 | Round trip/broadcast fail; connect/persistence pass. |
| Send update only to sender | Only broadcast fails on bounded waits for missing second-client receipt. |
| Omit font from connection snapshots | Connect/persistence fail; update checks pass. |
| Serialize connection font as string | Connect/persistence fail on type/value; update checks pass. |
| Persist only theme | Only persistence fails; memory echoes/fresh subscription pass. |
| Restore budget two instead of saved zero | Only persistence fails; same-process checks pass. |

All eleven applied, compiled and failed intended behavioral assertions. No
survivors or mutation setup failures. The missing-broadcast deadline tests that
liveness property against a working baseline, not an incidental hang.
`mutations.json` and each named `.patch`/`.txt` retain edits/results. Restored
scratch mutant source was removed. Unchanged inherited behavior reuses the
accepted [Chapter 7 evidence](enhancement-07-code.md).

Final stable-overlay checks passed:

- `go test ./internal/gradeedition2 ./cmd/grade-edition2 -count=1` (9.012s).
- `go vet ./internal/gradeedition2 ./cmd/grade-edition2`.
- In accepted optional GUI module: `go test -race ./... -run '^Test(SettingsWireAndDisk|SavedBudgetEntry)$' -count=1` (1.618s).
- `gofmt -l` and scoped `git diff --check` produced no findings.

Initial focused build failed because the new grader lacked the required `time`
import. Restored before all passing results; `initial-settings-test.txt` retains
this setup mistake, not a runtime defect or mutant kill. No broad root suite or
known wrong-directory original cascade was rerun.

Required retained source proof is substantive: `TestSettingsWireAndDisk` uses
real files, malformed input, write failure and publication. `TestSavedBudgetEntry`
launches the actual app, saves two-batch and eighteen-batch limits, crosses a
restart, tests zero's default, and updates during a held response to prove the
active turn keeps its original budget. Existing seventeen Go and eight browser
mutation records remain under `solutions/edition-2/evidence/ch09`. Preserve their
load-validation limitation: broad bypass reaches negative-budget rejection; a
separate Go source mutant forces wrong loaded UI values and fails the earlier
valid-save restart assertion. It does not isolate the later invalid-disk clamps
or independently prove each clamp assertion's sensitivity. No exact local range
is imposed here.

## Published examples and retained interface evidence

Final manuscript SHA-256:
`a2214a2f7a44763a4689c01386bac8f0c48a1bdd858dc017481adb5de9a46ed7`.
Both JSON fences were extracted directly and parsed. After setting the stated
system theme, speed 2, font 18, true TTS and budget 2, the exact published sparse
request bytes were sent to the real accepted WebSocket server. Sender and second
client returned the exact published settings object: system, false, 2, 18, 0.
`sample-wire.txt` records both. The two Go fences compiled together; sparse
false/zero preservation, decode failure and explicitly local clamps passed
(`sample-go.txt`). No manuscript fix was needed.

Source comparison confirms tool routing recognizes call deltas/finals and
execution observations, current/changed snapshots share application, CSS uses
actual `in_flight`/`tools_pending` states, and the speech slider uses the shown
`step="any"`. No new launch specimen: Chapter 7's explicitly inherited own-module
build/launch already has execution evidence. These schematic comparisons are
not claims of a new browser run.

Student history matters: copying `thinking`/`tool_use` into CSS and test hid the
real state-name mismatch. Corrected computed-animation and restored-selector
mutation evidence remain. Original independent review then found saved 1.75
shown correctly while a 0.1 slider step yielded native speech near 1.8. R1's
`any` fix and restored defect cover 1.75 and 1.2345. Earlier exact-equality
float-precision failure remains separate, as does an earlier timeout whose
location did not establish that cause. See `browser-r1.json`, `speed-r1.json`
and `speed-r1-mutation.json` in accepted evidence.

Retained live `gpt-4o-mini` browser use followed actual discovery, restored
preferences, observed real `CH09_LIVE_OK` command output, shared an Agent with
a terminal hint and received native speech start/end events. Unchanged
Anthropic/Gemini 3.8 evidence carries forward; no duplicate paid matrix was
needed. Native events/rates establish API behavior, not human listening quality.

## Size and handoff

Counts exclude blank and `//`-only lines from active Go; no block comments occur
in these files. New production is **216 active / 15 comment / 238 physical**;
real integration test **16 / 2 / 20**. Command diff: 13 physical insertions and
11 deletions, chiefly extending dispatch and flag descriptions. Original
Chapter 9 checks/launcher total **438 / 68 / 587**, preserved unchanged. Reusing
the evaluator and Chapter 7 helpers avoids copying another private launcher or
introducing a framework. This is grader growth, not a runtime-size claim.

New production comments are about 6.5% of nonblank lines, below the roughly 20%
teaching target. They explain fixture/default distinctions, known-field
comparison, real boundary, original scenarios and why restart must replace the
process. The sole new exported function documents its contract. Much remaining
code is sequential resource/error handling; review should judge clarity rather
than reward a percentage or request padding.

Raw evidence:
`/var/folders/p_/11vk6y114xq7k9g5r_pc3yv80000gn/T/ensemble-enhancement08-e73gc59d`.
Detached `validation/` is committed base **e6ae895**, with only three final
grader/command files overlaid to isolate concurrent first-edition work.
`validation-overlay.json` matches shared source:

| File | SHA-256 |
| --- | --- |
| `cmd/grade-edition2/main.go` | `7b3e168aa57708b1e2d1b153008c8e225e2729832755f9013a6a8974870a0614` |
| `internal/gradeedition2/ch08.go` | `126df145ca90815fd17ed2fec5879c29c9d248aeeb1a908937343a6b78e95c34` |
| `internal/gradeedition2/ch08_test.go` | `abe64d2c64844f609729d4f8a35c7e29afc9e6b39eb74fb28225d44156cded42` |

After independent acceptance and release, the Chapter 8 validation worktree,
its scratch-only probe/sample tests, temporary sample Go program and compiled
harness binaries were removed. Raw JSON/text/patch receipts remain at the
directory above, including the exact JSON sample bytes and final hash manifest.
Prior Chapter 7 validation was likewise removed after acceptance. No process
using the Chapter 8 evidence directory remained active at cleanup; reviewer
also reported its processes idle and scratch removed. No further runs or coder
commits are planned. No chapter beyond this bounded pass is authorized here.
