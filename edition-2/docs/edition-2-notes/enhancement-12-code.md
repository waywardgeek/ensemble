# Edition 2 Chapter 12 grader submission

Coder: `edition2_grader_coder`, GPT-6 Astra/high, October 11, 2026.
Initial submission; independent combined review pending. Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`.

Read the entire skill, carryover/workflow, original Chapter13, complete student
and comparative reviews, original grader and relevant accepted/legacy paths.
Reloaded the full skill, carryover and workflow after compaction. Original chapter
revision `994bf218b31033b7dbf2740dd0187d509d2927dc`, SHA-256
`581ff4cc7a04071e9899bd26cfa4bf249c21ba9258d3e137727c1460a3e99042`.
Accepted source is direct `solutions/edition-2/ch12` (original13); legacy is
`solutions/ch13`. Their trees remain `6db89522a1e5db9b8d5f7a5eedadd27315ff87eb`
and `00f53fedc8a576e88942556ee9d88536b156dcb2`. No runtime, frozen export, original
chapter/grader or unrelated CodeRhapsody work was edited. No mocks or paid runs.

## Scope and observations

Keep original weights 15/20/15/10/15/15, plus 10 for all corrected Chapter11
checks. `go run ./cmd/grade-edition2 -ch 12 [core-directory]` defaults to direct
`ch12`; existing `-gui` and inherited fixture flags remain. A tested stage was
checkpointed at `10dd7ffe` before final evidence and independent acceptance.

The original chapter example omitted skill kind and grants. Its grader supplied
grants but still omitted the loadable kind, expected undefined `think`, launched
one command for different clients, and used implementation names to infer unload
and parity. The revised driver exercises the actual headless and separately built
GUI commands, both with `--gui-debug --skills-dir PATH`. An invalid environment
directory beside the valid selected directory establishes override behavior.
The original public fixture's explicit `EN_PRIMARY_SKILL=ensemble` remains.

Each actual Agent talks to a small stateful Python MCP peer and local provider
HTTP server. Five requests select click, input, unload, a guessed click, then a
completed final response. The peer changes state only for the real method and
correct selector/text, reports its own PID, and records its actual RPC messages.
The test observes:

- Primary and startup instructions together in the initial system text, frozen
  across subsequent requests, including unload. Ordinary dynamically loaded
  manuals still use history; this does not change their Chapter9 contract.
- Exactly the two callable browser tools and lifecycle tools before unload;
  automatic snapshot and speech tools remain hidden.
- Fresh unique snapshot and speech markers once per request, no stale copies,
  and real pending speech data. Markers are observed in message text, excluding
  declarations, call arguments and correlated tool-result blocks.
- The next snapshot reflects click and input effects. Correlated result-only
  sentinels establish actual return values and input forwarding.
- After unload, no browser declarations or automatic context, a correlated
  guessed-call error, and no extra execution at the external peer. Startup text
  remains. The self-reported peer PID is absent while the parent command remains
  alive waiting for stdin; ordinary whole-process exit cannot supply this proof.

The peer is the exercise's external simulated browser, not a replacement for
Agent, actor, jobs, reducer, skill admission or MCP transport. PID observation
covers this direct peer's synchronous termination/reaping, not arbitrary process
trees or all cancellation races. Cleanup also kills a deliberately surviving
mutant peer. All provider responses use a declared model and disable streaming;
the unique completed final text is the observation barrier. The shared Chapter11
command helper only gains the actual process handle and stderr-on-EOF diagnostics,
read after its scanner's `Wait`; previous protocol checks remain intact.

## Fixture corrections and results

Evidence: [enhancement12](../../solutions/evidence/ch13/enhancement12/).
The first accepted focused run passed90/90. `legacy-focused.txt` and
`legacy-diagnostic.txt` retain the initial0/90 setup failure: an empty primary
selection relied on the accepted launcher's default, while legacy had no loaded
primary admitting `gui-debug`. Explicit original-fixture primary selection fixes
that setup. `legacy-primary-selected.txt` then measures75/90; only startup
instructions fail. The legacy launcher invokes `load_skill` but discards its
returned body, so loading/discovery and effects work without delivering the
startup manual. This is distinct from six independently observed failures.

| Measurement | Result |
| --- | --- |
| Historical original student grader, earlier source | **0/100 FAIL**, make2/grader1; omitted-kind startup and source-name parity failures |
| Revised full accepted run | **100/100 PASS**, exit0, 16.902s |
| Revised full legacy run | **75/100 FAIL**, exit1, 15.408s |
| Final affected checks after representation corrections | accepted90/90; legacy75/90 |
| Final totals using unchanged full-run parity | **composed100/100 accepted; composed75/100 legacy** |

The full reports are actual runs, not composed scores. They precede two subsequent
wire-representation corrections in this grader. `source-binding.json` binds those
runs to stage `10dd7ffe` plus `full-grader-source.patch`, and records final source
hashes. `accepted-full-run.json` and `legacy-full-run.json` preserve exact commands,
exits and durations. Final focused results plus unchanged parity10/0 support the
composed final totals; no second full run is claimed or needed for these local
representation fixes.

Coordinator/reviewer source inspection caught both representation mistakes before
formal submission: comparing raw system interfaces panicked on valid text-block
arrays, and scanning only message arrays ignored valid string content. Existing
`skillSystem` now compares logical system text; `guiDebugText` admits message
strings while still excluding tool blocks. Real isolated command controls using
these valid alternate renderings preserve the actual before panic/false failures
and after90/90 results in `wire-before.json`, `wire-after.json` and raw outputs.
These are positive compatibility controls, not broken-runtime mutation kills.

The post-unload prefix check rejects fresh markers4/5 as well as stale1–3. It was
added during source inspection before the mutation/full runs; no initial survivor
motivated it or is claimed. Exact per-round checks still detect stale repetitions.

Legacy inherited parity remains zero. Valid null-save overwrite and skills/MCP
behavior failures coexist with inherited setup limits: missing separate consumers,
workflow/module paths, non-executable consumer launch, differing GUI layout and
unsupported default media/model rows. Chapter11's ordinary chat does not start
the skill peer, unlike this chapter's debug startup. Its hidden automatic reverse
call executes outward rather than being refused. The complete raw nested report
is retained; not every masked check is an independently diagnosed runtime defect.
No recursive compatibility adapter or broad timeout repeat was added.

## Sensitivity and validation

`mutations.json` and fifteen exact patches/raw outputs record fifteen applied,
compiling targeted defects, all failing their intended assertions. Baseline and
restored accepted runs pass90/90. The probe exits zero to collect check arrays;
that exit is not a mutant pass. The cases cover both clients' startup selection
and directory override, missing startup manual, premature manual deletion,
visible automatic tools, stale context, wrong click/input arguments, missing TTS
content/status, omitted disconnect, detached-but-live peer and successful diagnostic
text for a guessed call. The detached-live-peer case fails only the lifetime
assertions. Missing startup and omitted disconnect cause coupled failures;
multiple failed rubric rows do not become multiple independent proofs.

Unchanged Chapter11 evidence supplies result forwarding, call correlation,
multi-block/error preservation and automatic lifecycle sensitivity. This chapter
adds the specific startup/state/argument/removal controls above. Native DOM,
speech and input provenance retain the accepted real-browser/source evidence;
numeric subprocess success does not prove them. General architecture, ownership,
per-server grants, slow automatic completion and interrupt-preserved jobs retain
source and prior independent checks rather than duplicated numeric claims.

`final-affected-checks.json` records final accepted90/90 and legacy75/90, affected
Chapter11/12 integration tests under race, command tests, vet and formatting.
All validation commands succeed. The original grader was preserved, not rerun
through its known invalid fixtures for freshness.

## Exact manuscript samples and teaching limits

Validated manuscript SHA-256
`6e0fd380bc203701634f1bfde8861f58a177131765ed36a81c0e7f437da3ec6c`.
`samples.json` records every fence hash, actual command and observation. Both
published builds and launches run from the documented core/module directories,
using `/tmp/ensemble-ch12` and `/tmp/ensemble-ch12-gui` to avoid the library-directory
name collision. The exact stdin envelope drives real calls and four requests.

The headless example retains the two complete manuals and replaces only the
server entry with a faithful stdio companion as explicitly instructed. The GUI
example uses the exact WebSocket skill, real port8088 launcher, served assets and
actual WebSocket peer. Raw provider requests and tunnel frames establish returned
results, refreshed state, disappearance of automatic content/declarations after
unload and retained startup text. This is launcher/tunnel validation, not a new
browser or paid-model demonstration. The two current JavaScript excerpts match
accepted source modulo whitespace; their execution proof is the retained native
Chrome correction evidence, not a mock DOM. Historical pause, `DeferFinish` and
Settings snippets remain labeled history rather than newly required APIs.

The historical original0/100 remains unchanged. Original13's genuine native-speech
input-cause defect and its accepted round1 correction remain separate from the
invalid grader fixture. Native speech, human whitespace, programmatic input,
other-tab explicit unpause and real subsequent submission have accepted browser
controls and three restored defects. Two initial browser setup failures and the
unexplained earlier simultaneous-tab speech stall remain in that record; neither
subjective listening quality nor general multi-tab reliability is claimed.
Anthropic/Gemini live task success and both OpenAI task failures retain their
original earlier-source provenance. There was no paid repeat after speech/manual
corrections. Exact examples required no prose change at this submission.

## Size, comments and review boundary

`source-counts.json` uses physical, blank and `//` comment-only lines; remaining
lines include braces/data and the embedded Python fixture, not just statements.
The new chapter driver has303 physical,275 active and20 comment-only lines; its
integration test has26/22/1. The original harness/check files total700 physical,
545 active and81 comment-only lines. Shared helpers already supply provider,
protocol, declarations, correlated results and inherited parity; no new framework.
This is a scope comparison, not a claim that fewer lines alone improve coverage.

`Chapter12` is the only new exported production symbol and documents its real
boundary and browser limit; the test function also has a purpose comment. Other
new helpers are private. Comments explain independent command wiring, valid
metadata, return-only sentinels, actual lifetime versus shutdown, mutant cleanup
and message/result separation. The roughly7% comment-only share is below Bill's
teaching target; the driver reuses explained helpers and straightforward assertion
code rather than adding quota padding. Independent review must judge whether
those explanations suffice. No ownership/API redesign or product change is hidden
in this grader. Chapter13 remains unreleased until the combined gate.
