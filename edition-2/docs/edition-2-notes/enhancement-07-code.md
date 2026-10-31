# Edition 2 Chapter 7 grader review submission

> Historical paths: the `edition-2/ch04`–`ch09` paths recorded below use
> original student numbering. Those exports are now preserved under
> `solutions/edition-2/history/`; see the [solution mapping](../../solutions/README.md).

Coder: `enhancement_coder_resumed`, October 10, 2026. Initial formal submission;
independent review pending. Mandatory handoff: `.agents/skills/ensemble-coding/SKILL.md`.

Read the entire skill, carryover, enhancement plan, original Chapter 8, student
notes and comparative review; inspected both actual GUI entrypoints and original
harness/checks. Coordinator and reviewer approved the narrow plan before edits.
No original grader, runtime, frozen export or public protocol changed. No mocks,
paid repeats, whole-root suites or future-chapter work were introduced.

## Source and historical provenance

Original Chapter 8 revision is `994bf218b31033b7dbf2740dd0187d509d2927dc`, SHA-256
`207ae885a7a9d6210757b5f3b4d9773a98206a66e62710e5ce0be0a6ce180f56`.
Accepted endpoint: `solutions/edition-2/ch08`, tag `edition-2-ch08-r3`, unchanged
from reviewed progress `2af66bb`. The historical 66-file export manifest is
`4aff5286052ed7aa49f5e3a56d093563c07c376a221c81a3014e60989c990ebb`.
Actual legacy module is `solutions/ch08` with `cmd/main.go`; a nested
`solutions/ch08/agent` does not exist. The initial handoff's path assumption was
corrected from the files before grading.

Retain [student notes](ch08.md), [comparative review](ch08-code-review.md) and the
[original output](../../solutions/evidence/ch08/grader-gui.txt).
Original accepted **55/100 FAIL** is the actual earlier GUI binary, before its
socket-lifetime, terminal-display and replay-status corrections. Its missing
`state_changed.from` was a real defect fixed in R1, not a grader exception.
Its inherited CLI cascade ran the GUI server as finite core exercises for roughly
47 minutes. The coordinator explicitly directed reuse of that historical result,
not another run of the known wrong-entrypoint cascade. Historical legacy original
**100/100 PASS** also remains actual recorded history, not a new execution.

## Contract and coverage

The command now supports `-ch 7 [CORE]`, defaulting to
`solutions/edition-2/ch08`. `-gui DIR` defaults to `CORE/gui`; it builds `cmd/`
when present, otherwise the module root. Existing `--port` and `--gui-dir` launch
that actual application with a disposable workspace and the established provider
environment. `-gui-model` defaults to declared local Anthropic fixture
`claude-sonnet-5`; this is neither a model-discovery claim nor a mandated table
spelling. Earlier consumer/workflow/media/stream flags apply to core parity.
The grader never translates a GUI server into a core protocol adapter.

| Check | Points | Preserved and corrected proof |
| --- | ---: | --- |
| websocket-streams | 25 | Original six types and state/tool fields; actual text delta/final content and identities, recognizable kinds, final tool arguments, correlated dispatch/result/error, actual provider tool result; optional module resolves this core through public imports. |
| event-replay | 20 | Original completed-turn replay, strengthened to actual completed final/tool content and matching artifact/call identities, with no stale partials. |
| gui-log | 15 | Original two directions; every nonempty record parses as RFC3339Nano timestamp, direction and message JSON. |
| pause-holds-tools | 20 | Original three dispatches/results and turn completion; three real commands, gated first completion, no second execution while paused, explicit other-tab unpause and actual successful results. |
| ch6-parity | 20 | Every corrected Chapter 6 check on the real core entrypoint. |

The original `Ch8Evaluate` predicates remain in use. Its private launcher cannot
be reused without the recursive wrong-binary cascade, so a narrow separate driver
collects the evidence. GUI public-dependency failure attaches to websocket25, but
actual wire failures or successful wire checks remain independently legible in
details. The shared dependency helper adds a target-package parameter internally;
existing consumer/workflow defaults still inspect `.`. GUI inspection targets the
actual `./cmd` executable when present. A clean library root cannot hide private
imports in its command. There is no AST framework or required API spelling.

The first wire fixture uses real `run_command` with `printf GUI_TOOL_RESULT`.
The next captured provider request must carry its nonerror correlated result;
a fake final answer alone is insufficient. Embedded neutral finals and historical
flattened finals are both understood. JSON tool arguments compare semantically.
Numeric or namespaced-string artifact IDs are allowed; missing IDs and repeated
final identity across responses do not silently become valid cards. Observable
opaque identity is recognized, without claiming that flattened legacy opaque
flags prove retained provider bytes.

Pause uses a shell waiting on a scratch release file, followed by two marker
commands. After dispatch, the first WebSocket sends pause followed by an ordinary
Ping. Both implementations synchronously process client messages in their reader;
Pong therefore establishes that pause handling returned before the first command
is released. No new application acknowledgement or accepted-only `pause_changed`
requirement was imposed. The first real result arrives while paused. A bounded
200ms negative interval checks both dispatches and the second command's file;
unpause from the other connection then permits all three real outputs and the
next provider request. This tests the existing Agent-wide rule, not per-tab vetoes.

Completed replay has no mandated terminal envelope. The driver collects a bounded
500ms idle transcript from this otherwise finished Agent, then compares all its
actual earlier finals/tool facts. Causal active-partial replay, atomic history/live
handoff, fixed-window bounds, paused interrupt, disconnect races, actual ownership,
browser rendering and native speech remain required source/integration proof.
This numeric score does not claim those properties from output type labels.

## Actual results and the log correction

| Implementation | Original historical result reused | Initial revised full execution | Final revised composed result |
| --- | --- | --- | --- |
| Accepted Chapter 8 endpoint | 55/100 FAIL, earlier binary | 100/100 PASS | 100/100 PASS |
| Legacy Chapter 8 | 100/100 PASS | 20/100 FAIL | 35/100 FAIL |

The initial strict trace parser incorrectly counted whitespace-only separator
lines as malformed messages. Gorilla `WriteJSON` retains a trailing newline;
the legacy logger appends another. Those blank separators contain no message and
do not invalidate the actual timestamp/direction/JSON records. Coordinator and
reviewer agreed to ignore empty separators while strictly validating every
nonempty record. Initial results are retained, not relabeled.

Final results are explicitly **composed**: actual rerun final GUI components are
accepted80/80 and legacy35/80, combined with unchanged actual full-run corrected
parity20/20 and0/20 respectively. Accepted received69 live messages and9 replay
messages. Legacy passes logging15 and genuine pause20. It fails websocket25 for
reused final artifact identity and the integrated-module boundary, and replay20
for absent completed-part identity/content correspondence. These are real observed
GUI differences, not just the structural boundary penalty.

Legacy parity is reported separately. Its Chapter 8 core now honors the disable
flag; do not import the earlier Chapter 7 failure as a current observation. Its
absent-model fallback still fails, and its export lacks the separate ch05/ch06/ch07
applications needed by a complete merged-chapter submission. The raw inherited
consumer build falls through to a non-executable library archive and reports
permission denied; that is missing-exercise setup, not evidence that its real
custom tool ran and failed. Other inherited tool/parity failures and default media
fixture mismatches remain in the actual report. No old-answer application was
invented, no earlier snapshot was substituted, and no adapter forced a pass.

## Validation and published examples

Final isolated affected checks passed:

- `go test ./internal/gradeedition2 ./cmd/grade-edition2 -count=1` (8.263 seconds;
  command package has no tests).
- `go vet ./internal/gradeedition2 ./cmd/grade-edition2`.
- Accepted optional GUI `go test -race ./... -count=1` (2.845 seconds), including
  `TestStreamReplayPause`, `TestPausedInterruptAndReplayBatch`, `TestReplayStatus`
  and `TestReplayWindow` against real core, HTTP, shell tools and WebSockets.
- Formatting, whitespace and final shared/isolated grader hash equality. Later
  edits before submission were explanatory comments only, not behavior changes.

These source tests retain causal active-partial replay, cleanup during tools,
identity across turns, global pause, paused interruption/continued use, more history
than the live queue, fixed window and connection lifetime proof. Their accepted
source mutation receipts remain in `solutions/edition-2/evidence/ch08/`:
`mutations.txt`, `mutations-corrected.txt`, `window-mutation.txt` and
`replay-status-mutation.txt`. Preserve the original ineffective cleanup mutation
and corrected timing of its assertion; this pass does not claim the numeric
completed-turn check replaces that stronger causal source test.

Final manuscript SHA-256:
`067f449d07535d60d952059fadb0c54e1c17f07a92504e55f22fdc5d27c4c180`.
All twelve published JSON records parse. The Go observation vocabulary matches
the public source. The exact own-module `go build ... ./cmd` and published
`--port 8088 --gui-dir` launch ran against a faithful provider fake; only the
binary destination was redirected into private scratch and the explicit asset
placeholder filled. The six client examples were submitted as independent
commands, followed by another prompt. Actual reply, interruption and subsequent
successful reply were observed; `published-samples.json` retains messages,
requests and trace, and `published-samples.txt` records PASS (0.245 seconds).
Server envelopes are illustrative vocabulary, not a fabricated full conversation.

The first sample assertion wrongly took the interrupt's `turn_ended` as the next
successful request completion. It failed with “Agent not usable after controls.”
Source and the corrected receipt show the expected `turn interrupted` event,
then `Still usable.`. Only the scratch assertion changed to correlate the desired
reply; no manuscript or runtime correction was needed. The initial output remains
`published-samples-initial.txt`.

Existing browser/native speech/live evidence is reused without paid repeats:
`browser.json`, `speech.json`, `browser-mutations.txt`, `projection.json`,
`projection-mutation.json` and `live-summary.json`. They cover safe/reachable
rendering, readable rather than opaque speech, genuine native interruption and
three-provider tool turns including required Gemini3.8. Native start/end proves
observed API behavior, not subjective listening or audible intelligibility.

## Targeted mutation results

Nineteen plausible defects were detected. `mutations-final.json` contains each
actual check result; corresponding `.patch` and `.txt` files retain source and
output. Only intended behavioral failures count.

| Mutation | Observed sensitivity |
| --- | --- |
| Missing delta ID | websocket identity/content check |
| Unknown delta kind | websocket required-kind check |
| Corrupted final text | delta/final and expected text mismatch |
| Missing final opaque kind | recognizable finalized-kind check |
| Corrupted final tool args | actual finalized name/arguments check |
| Missing dispatch input | correlated actual dispatch contract |
| Corrupted result text | websocket result and pause real-result checks |
| Missing result error field | websocket and pause error/result contract |
| Corrupted replay text | completed-content replay check |
| Corrupted replay ID | completed artifact correspondence |
| Corrupted replay tool result | completed call-result correspondence |
| Injected stale replay partial | replay rejects stale data; initial subscription also fails text correlation |
| Invalid timestamp retaining year-like shape | full timestamp parser, not the old substring heuristic |
| Invalid JSON retaining an opening brace | full message parser, not the old substring heuristic |
| Omitted outbound trace | required actual direction |
| Bypassed Agent pause | three starts observed before other-tab unpause |
| Ignored other-tab unpause | bounded completion deadline with working baseline and held real gate |
| GUI resolves another core | actual resolved-submission check; wire behavior still passes |
| Only command imports private core | actual executable graph check; build and wire behavior pass |

The command-only private import control uses a permitted related module prefix
so Go's internal rule allows compilation; the grader then rejects that actual
private dependency. This distinguishes meaningful graph sensitivity from a
compiler rejection and from checking only the clean library root.

The initial invalid-JSON mutation had a missing closing parenthesis and failed
compilation. It is preserved as `log-invalid-json-setup-failure.patch/.txt` and
in `mutations-initial.json`, and does **not** count as a kill. The corrected
executable mutation fails gui-log alone. No behavioral survivor remains in this
changed-claim set; unchanged earlier-source survivors remain historical evidence.

## Size, comments and handoff

Count nonblank source lines excluding comment-only lines; mixed code/comment
lines count as active. New production `ch07.go`/`ch07_driver.go` contains
**609 active /37 comment-only /665 physical** lines. The real integration test is
**16/2/20**. Original checks/harness remain **687/77/867**, unchanged and reused
where callable. Shared command/helper edits add23/delete10 physical lines. This
is additional versioned grading code, not a claimed deletion from the original.
The private original drivers and recursive launcher cannot supply separate clients;
the new driver owns the real process/socket lifecycle and richer observations.

Comment share is about6% of new nonblank production, below Bill's approximate20%
teaching target. Comments explain the separate launchers, public executable graph,
reader cancellation, synchronous Ping/Pong ordering, real-result capture,
replay collection/limitations, identity/content comparison and tool cleanup.
Exported Chapter7/test symbols are documented. Reviewer should assess whether the
fixture/lifecycle code is understandable; a percentage alone does not justify
padding or acceptance. No new generic transport or mutation framework ships.

Raw evidence:
`/var/folders/p_/11vk6y114xq7k9g5r_pc3yv80000gn/T/ensemble-enhancement07-m77yoi34`
(pointer `/tmp/ensemble-enhancement07-current`). Its `validation/` worktree uses
committed **54939b6** with only our five submitted grader/command files overlaid;
`validation-overlay.json` records final hashes. Final composed JSON files and
actual final component transcripts sit beside full initial runs, mutations and
sample receipts. Disposable mutant source was removed; stable validation remains
for independent review. Prior Chapter6 validation was removed after acceptance,
while its small raw evidence remains. Coder harness is idle and ready to release.
No Chapter8 work or commit is performed by this submission.
