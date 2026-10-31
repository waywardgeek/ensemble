# Chapter 9 comparative code review

Reviewer: Codex independent code reviewer, October 9, 2026.

**Final verdict, October 10, 2026: accepted with the documented inherited grader
exceptions. One correction round used (maximum three).** R1 is resolved; no
implementation, scope or test finding remains open. Actual original grading is
**80/100 FAIL**, with all four new settings checks passing. The initial review
and its pending-result statements are preserved below as dated history.

Read the entire mandatory `.agents/skills/ensemble-coding/SKILL.md`, carryover,
workflow, original Chapter 9 and completed student review, including its
evidence-only mutation clarification. Original chapter revision is
`2b2fedbc8b50bc688c31ab823712553e589f8a3b`, SHA-256
`3693cdbb54b6ce66698f635586b17973b9f4a707f8a51c3898b17e2d3ee3349b`.
Reviewed student source at `89d17e0` against accepted predecessor `46fd955`
(`edition-2-ch08-r3`), and matching original `solutions/ch09` against `ch08`.
Every current file in `review-source.json` matches its recorded hash.

## Required correction

**R1 — accepted speech speed differs from the visible control and actual speech.**
`gui/settings.go` accepts and restores an in-range `tts_speed` of `1.75`.
`gui/web/gui/app.js:347` assigns that value to the range input while line 348
displays the original value. However, `index.html:25` gives the range input
`step="0.1"`; Chrome sanitizes its value to `1.8`. `app.js:49` reads that
sanitized control for the native utterance. The saved/echoed setting and label
therefore say `1.75`, while speech uses approximately `1.8`.

Independently reproduced with the actual compiled GUI command, a saved settings
file, native headed Chrome, the real page/WebSocket/Agent, and a small external
HTTP provider fake. A wrapper recorded the utterance and forwarded it to the
real native speech API. The observed result was:

```json
{"label":"1.75","slider":"1.8","nativeRates":[1.7999999523162842]}
```

This is a step-quantization mismatch, not the harmless float32 representation
of `0.6` already documented by the student. Keep the accepted/saved value,
displayed control and effective speech rate consistent for supported values;
add a focused browser regression and meaningful restored-defect check. This
requires no precision framework, new setting or paid provider repeat. A first
bare headless Chrome probe timed out; it was a failed reviewer setup attempt,
not product evidence. The subsequent real-application reproduction above worked.
All reviewer scratch, browser processes and fake-server workspaces were removed.

## Ownership, scope and teaching quality

The corrected ownership split follows Bill's rules. The optional GUI module owns
browser preferences and the persisted envelope. The Agent owns the runtime tool
budget, exposed through its public interface; the actor reads its synchronized
configuration once at turn start. GUI snapshots obtain the effective budget from
that owner. Saving it for startup does not introduce a competing runtime policy.
CLI and browser still enter one Agent and actor; Engine, history, tools and jobs
retain their actual-parent access. Core imports no GUI/WebSocket implementation,
and implementation spokes still depend on common rather than one another.

The new mechanism is small: a value snapshot, decode/clamp helper shared by disk
and wire, replacement file write, existing server lock and one public setter.
Decode occurs on a copy; failed decoding or persistence cannot publish a partial
update. Subscribe, apply and broadcast share ordering. Zero/false are serialized
in full snapshots, while omitted patch fields retain their values. The setter
rejects negative limits, zero uses the existing named 200 default, and changing
the limit leaves the active turn's snapshot intact. The coordinator's earlier
boundary correction and the student's discarded core preference placement are
properly attributed in the student review.

Two panes reuse the same artifact renderer, including tool-part identity and
replay. The tree is recursive but contains only the required single node; the
active indicator now uses the actual reducer's state vocabulary. Native controls
and selected attributes expose state. Pointer capture and the small shared
keyboard resize path are reasonable current-interface choices. Theme changes,
system preference changes, font changes and TTS enable/disable have real
consumers. No sub-agent management, model selector, API-key panel, conversation
persistence or configuration service hierarchy was added. Existing Agent-wide
unpause and interrupt/job lifetime semantics remain intact.

The original adds a larger shared-core settings store and GUI-facing exports,
with repeated per-field merge paths and extra model/request settings beyond
this chapter's four visible controls. Its new budget field is stored without
replacing the actual hard-coded execution limit; its snapshot tags omit false
and zero, and load/write errors and clamping are not handled as the revised
chapter requires. The student materially improves those current obligations,
despite the original's 100/100 grader result. The separate GUI module and public
runtime-budget interface are clearer than embedding UI behavior in common.

New exported production symbols and changed interfaces have useful documentation.
Comments explain patch versus snapshot, ownership, turn lifetime, persistence
before publication and subscribe ordering. Core production comments are 19.5%
of nonblank lines; GUI Go is 14.8%, and browser JavaScript 9.6%. These local ratios
do not justify padding: the non-obvious choices are explained and inherited
teaching comments remain. The CSS is more compact than the original, so its line
reduction alone is not credited as simpler design.

## Counts and chapter growth

Count physical lines as blank, comment-only or active. Active includes imports,
braces and literals; mixed code/comment lines remain active. Count leading `//`,
block comments and HTML comments separately. Original inline JavaScript is counted
as JavaScript, leaving its script-tag lines in HTML. Exclude binaries, runtime
logs/data, module manifests, README prose and evidence from production. No
generated source was introduced. Values are **active / comments / physical**.

| Category | Chapter 9 | Delta from its own Chapter 8 |
| --- | ---: | ---: |
| Original core Go production, including transport | 5,862 / 1,872 / 8,535 | +198 / +32 / +251 |
| Original Go tests | 769 / 160 / 1,012 | unchanged |
| Original browser JavaScript | 480 / 30 / 589 | +233 / +18 / +290 |
| Original CSS | 360 / 15 / 435 | +244 / +13 / +297 |
| Original HTML | 93 / 8 / 108 | +65 / +8 / +79 |
| Student core Go production | 4,065 / 982 / 5,359 | +15 / +3 / +19 |
| Student optional GUI Go production, including command | 508 / 88 / 616 | +103 / +19 / +127 |
| Student ch05 example production | 74 / 13 / 92 | unchanged |
| Student ch06 example production | 267 / 20 / 295 | unchanged |
| Student ch07 example production | 74 / 4 / 82 | unchanged |
| Student core Go tests | 2,159 / 177 / 2,409 | +10 / +2 / +13 |
| Student GUI Go tests | 598 / 29 / 641 | +275 / +13 / +293 |
| Student ch05/ch06 example tests | 204 / 25 / 234 | unchanged |
| Student browser JavaScript | 413 / 44 / 461 | +93 / +6 / +100 |
| Student CSS | 165 / 0 / 165 | +26 / +0 / +26 |
| Student HTML | 52 / 0 / 52 | +23 / +0 / +23 |

Combined student Go production is **4,988 / 1,107 / 6,444**, growth
**118 / 22 / 146**; tests **2,961 / 231 / 3,284**, growth **285 / 15 / 306**.
Original browser orchestration moves from inline HTML into `gui.js`; its whole
new file is not wholly new behavior. Other module/example locations are stable.
The original adds a separate `verify_live.sh` support script, **59 / 1 / 68**.
The student's two retained browser/live `.mjs` launchers are evidence support,
**148 / 4 / 152**, not shipping code or a production-size advantage. Their compact
formatting is not a measure of test quality.

The larger student test increment protects actual persistence and execution
requirements the original added no tests for. The real-command test is justified:
testing only a settings struct could not detect an entrypoint ignoring a saved
budget. The production increase is proportionate and contains no speculative
framework; no bloat or additional scope correction is requested.

## Verification and evidence assessment

Independently ran GUI `go test -race -count=1 ./...` (PASS, 3.383s) and vet;
core focused budget/limit/setter race tests passed uncached (28.239s). Reuse the
student's successful full core and separate ch05/ch06/ch07 tests/vet, and the
unchanged inherited provider/live evidence. The GUI entry test builds its actual
command and exercises real stdin, WebSocket, actor, tools and shell execution;
only the external HTTP provider is fake. No mocks were introduced.

Read complete Go/browser mutation output and the later exact edit mapping.
Seventeen Go and eight browser mutants fail their intended checks, with no
compile error, incidental race or cleanup panic presented as a kill. Missing
broadcast/snapshot mutations use bounded waits for exactly the removed message.
Independently restored the mid-turn-budget defect in an isolated copy: the real
entrypoint test cleanly failed with **18 batches, wanted 2**, while the baseline
passes. This protects the required active-turn snapshot, not merely persistence.

The broad load-validation mutant fails because the real setter rejects the
unclamped negative budget; the separate loaded-UI corruption mutant fails the
earlier valid restart assertion. Accept the student's clarified limitation:
these are not independent proofs of every invalid-disk clamp branch. The correct
test does assert invalid loaded UI values and the same helper serves both entry
paths. No extra matrix is needed. The native speech-rate mutant proves ignoring
the selected speed is detected, but does not cover R1's accepted non-step value.

Inspected the headed-browser script, all twelve recorded checks, failed attempts,
mutation failures and screenshot. The CSS state-name fix is a real correction;
the subsequent `0.6` equality failures are test precision errors, honestly retained.
The browser uses the actual Go application and WebSockets. Native speech wrappers
forward to the real API and retain queued/start/end observations. They establish
API use and effective rate for the tested value, not subjective intelligibility.

The separate live script discovers `gpt-4o-mini`, restarts saved preferences,
drives the actual GUI, observes `CH09_LIVE_OK` in Actions, and supplies terminal
input visible in Chat during the same turn. Its native speaker records start/end
for the final answer. Credentials are read programmatically and passed to the
intended child, not printed or retained in the selected evidence. Unchanged
Anthropic/Gemini 3.8 coverage is appropriately reused rather than repeating a
paid vendor matrix for preferences. No reviewer paid call was made.

## Grading provenance and remaining gates

Original frozen reference: **100/100 PASS**, coordinator baseline. The only
student original grading invocation is the still-running GUI-directory command;
partial output has no final score. No second grader was started for review.

`grader-source.json` and `review-source.json` differ in exactly four files:
the added core setter test, GUI README, GUI constructor documentation and CSS
state selector correction. Go production behavior at grading launch is unchanged;
do not label that earlier input an exact final-source run. Browser and final Go
checks cover the later changes separately.

The inherited optional-GUI/single-directory CLI mismatch and earlier core
streaming/model/logger exceptions remain recorded history, not a blanket waiver
of new Chapter 9 checks. Chapter 8's earlier replay-status defect was corrected,
not waived. Once this grader completes, inspect the actual Chapter 9 settings
checks and nested outcomes individually before deciding any exception. Preserve
every actual failure and distinguish final verification from the grader's build.

Remaining gates: resolve R1 with bounded affected evidence; read the final student
review and original grader result; then issue the acceptance/exception decision.

## Correction round 1 reassessment

Reread the entire mandatory coding skill, updated student review, exact source
diff, focused native-speed script and complete baseline/mutation results, plus
the broader browser rerun. **R1 is resolved.** Current source matches every hash
in `review-source-r1.json`; its only difference from the initially reviewed
source manifest is `gui/web/gui/index.html`.

The input now uses `step="any"`, with a useful comment explaining why saved
rates between tenths must survive. This removes the unwanted browser
quantization while retaining the server's existing range, setting ownership and
native speech consumer. It is the appropriately small correction.

`speed-check.mjs` uses the real command, server, WebSocket settings path,
restart, browser controls and native speech. Only the external provider is
fake. The wrapper forwards the actual native call and records callbacks. The
baseline saves and restores **1.75**, with slider, label and native rate all
**1.75**; the second accepted value **1.2345** reaches native rate
**1.2345000505447388**. Both utterances start and end. The test's `0.00001`
tolerance covers representation noise without accepting the original defect.

The isolated mutation restores exactly `step="0.1"`. It captures the complete
behavior before asserting: saved/restored and displayed values remain correct,
but controls/native rates become **1.8 / 1.7999999523162842** and
**1.2 / 1.2000000476837158**. Both utterances still start/end, and the consistency
check fails. This is a clean intended kill, not a speech-service or startup
failure. The unchanged initial independent reproduction grounds the defect;
the corrected source and native evidence suffice without another browser run.

The twelve broader browser checks pass again. Replacing the generic slider's
ArrowRight assumption with native End, and checking the accepted value 2, is
appropriate now that the control has no fixed tenths step. The new focused
regression supplies the non-step coverage that the original 0.6 case lacked;
the old results remain as history. No paid call, Go behavior change, repeated
full mutation sweep or duplicate original grader was needed.

Final HTML counts are **52 active / 1 comment / 53 physical**, chapter delta
**+23 / +1 / +24**. All Go, JavaScript and CSS production/test counts above are
unchanged. The focused evidence launcher adds **57 / 2 / 59** support lines;
the three current `.mjs` launchers total **205 / 6 / 211**, separately from
shipping source. No further implementation, scope, teaching-comment or test
finding remains open.

**Sole remaining gate:** collect the existing original grader's actual result,
read the student's final explanation, and assess each new or inherited failure
before issuing final acceptance. Its earlier build predates this HTML correction
as well as the previously recorded four-file differences; the final native
browser results are separate verification, not a final-source original score.

## Final outcome and exception decision — October 10

Read the complete original output, `grader-result.json` and the student's final
per-failure explanation. Verified the raw output SHA-256 against the collection
record and all **69** final source hashes against `review-source-r1.json`.
The sole grading command completed with **80/100 FAIL**, grader exit 1 and make
exit 2. The recorded approximately 46m59s is an output-file timestamp estimate,
not separately instrumented process duration.

| Original check | Actual result |
| --- | --- |
| Settings roundtrip | 25/25 PASS |
| Settings on connect | 20/20 PASS |
| Settings broadcast | 20/20 PASS |
| Settings restart persistence | 15/15 PASS |
| Chapter 8 parity | 0/20 FAIL |

The nested report names only Chapter 7 parity under Chapter 8. Five stream check
labels report 30-second exercise timeouts; they are not five independently
established invocations. Nested Chapter 6 additionally reports Chapter 5
`star-topology`, `loud-refusal`, and `hub-clean`. No other Chapter 8 failure or
unreported nested numeric score is inferred.

I independently accept the student's inherited exception request, with these
precise grounds:

- **CLI lifetime/protocol mismatch.** The earlier harness launches the submitted
  GUI command as a finite protocol CLI, closes stdin and waits for exit. The
  reviewed optional application deliberately leaves its browser service alive
  when terminal input ends. That accounts for the timeout path; it is not
  evidence of a new parser hang or successful streaming. Existing deterministic
  core streaming tests and declared-model live evidence remain the independent
  behavioral proof. The earlier `fake-model`/plain-fallback conflict is retained
  separately, not substituted for this run's actual timeout result.
- **Wrong module for topology inspection.** The original helper uses
  `go list ./...` in the submitted directory; it does not enumerate the separate
  core dependency's package graph. Its hub checks derive the expected internal
  package from that submission's module base. A public optional GUI client
  therefore has neither the inferred hub nor its own `internal/common` node.
  This is narrower than saying the core is absent from all transitive imports.
  The actual core star and parent relationships were independently reviewed;
  duplicating core packages inside the GUI would undermine Bill's boundary.
- **Inherited refusal probe.** The original probe sends only `{"user":"test"}`
  to an unknown model, with no unsupported attachment. It also waits for this
  same long-lived GUI command to exit before evaluating output. Its failed
  result is not a new media-refusal regression. The existing accepted policy
  permits unknown-model text/plain delivery and rejects unsupported media when
  actually requested; demanding a blanket unknown-model refusal remains the
  documented earlier contract conflict.

No Chapter 9 settings obligation is waived. R1's speed mismatch was corrected
and mutation-checked, not accepted as an exception. The original build predates
the documented CSS/test/documentation and HTML changes; final native-browser
verification and unchanged Go behavior justify acceptance without repeating the
same inherited timeout cascade. **No exact-final-source original-grade run is
claimed.** The original reference's 100/100 PASS and the student's actual
80/100 FAIL remain distinct from this review decision.

Carry these bounded exceptions and their actual failed results into later parity
reviews. They do not waive new GUI/browser behavior, core architecture, actual
streaming, or unsupported-media handling. The later author/grader phase should
allow the correct core and optional-client entrypoints/module paths and reconcile
the unknown-model fixture, preserving valid behavior checks. No current product
feature or grader rewrite is required. All acceptance gates are now satisfied;
the coordinator may checkpoint and advance the student.
