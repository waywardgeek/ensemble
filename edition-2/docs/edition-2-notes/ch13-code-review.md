# Chapter 13 independent code review

Reviewer: `enhancement_reviewer_resumed`, October 10, 2026. Initial formal review
of `a24260b851c85ae68438aaec7db8a48bc3233c5c`, extending accepted Chapter 12
`cab2281`. Read the entire mandatory `.agents/skills/ensemble-coding/SKILL.md`,
carryover, workflow, original 315-line Chapter 13 and completed student submission
in [ch13.md](ch13.md). Chapter source revision/hash are recorded there. Compared
the student's changes with its own predecessor and original `solutions/ch13`
with `solutions/ch12`, including implementation, comments, tests and skills.

## Initial decision: correction required

One confirmed current-scope blocker: `gui/web/gui/app.js` ignores the marked MCP
input event, but `pace()` subsequently derives typing from `input.value` whenever
speech starts or finishes. Programmatically populated text becomes a typing pause
again. Speech ends without unpausing, so subsequent model-driven submission can
be stranded behind the Agent gate. The speech-disabled passing browser test did
not exercise this path.

The independent [native-speech probe](../../solutions/evidence/ch13/reviewer-speech-probe.json)
ran the actual built GUI, Agent, MCP browser tunnel and Chrome with only the HTTP
provider replaced by a small fake. The observation wrapper calls native
`speechSynthesis.speak`; it does not fake completion. With empty input, actual
speech ended and pause cleared. After actual `gui_input` filled `MCP submitted`,
subsequent native speech ended but the Agent remained `paused`; human clearing
the field released it. Native start/end events and wire trace are retained beside
the [probe source](../../solutions/evidence/ch13/reviewer-speech-probe.mjs).
This directly establishes stranded pause; the blocked next dispatch follows the
existing gate, rather than a separately measured submit in this initial probe.
No subjective audibility claim is made. The direct binary and browser were shut
down before handing the harness back to the coder.

Correction round 1 is released for that cause distinction and its focused
regression. Preserve genuine human typing, speech pacing, submission and Bill's
Agent-wide explicit-unpause rule. No Chapter 14 feature, new protocol, paid repeat
or broad sweep is required. There are no other confirmed blockers.

## Scope, ownership and startup interpretation

The new flags on both actual clients reach the same public composition root.
`StartupSkills` adds construction-time selection to the existing registry and
MCP lifecycle. It introduces no second dispatcher, callback service bag, mutable
global or GUI import into core/common. Existing actual-parent access and the
optional public GUI module remain intact. The shipped manual now names actual
`list_directory` and `search_files` handlers; the earlier unused grant spellings
were corrected during initial submission, with the live-source difference
explicitly retained. Adding aliases for those mistakes is unnecessary.

I accept the student's startup-manual interpretation: construction assembles
the prompt before it freezes, with application-origin sequence zero and no
fabricated model call. This differs from literal model-authored `load_skill`,
but preserves the purpose of Chapter 13's startup convenience and Chapter 10's
post-construction immutability. Ordinary dynamic loading still puts its manual
in history; executable MCP capabilities disconnect on unload while instructions
remain pending removal. The clarified result accurately separates those facts.
Future compaction is not implemented or required here.

## Original grader and evidence limits

Actual original grading is **0/100 FAIL**, grader exit 1 / make exit 2, from the
recorded headless build before final GUI changes. It completed; this is not an
interrupted or passing result. Independent inspection supports both proposed
exceptions: the supplied `gui-debug` fixture omits the kind required by the
accepted skill contract, and the parity check requires specific implementation
method/file/callback names rather than the real public MCP behavior. Chapter 10
defines the kinds without an absent-kind default. Neither a permissive default
nor dummy names/callback wiring should be added to satisfy these checks.

The fixture rejection masks later runtime checks. Moreover, original
`skill-unload` relies on source names rather than actually invoking unload. Those
points are not runtime evidence even for an implementation passing that checker.
These narrow grader exceptions are justified independently; they do not waive
startup, fresh context, interaction, actual TTS observation or unload behavior,
and do not excuse the speech/input defect above. The original remains unchanged.

Read the actual test/vet/race receipts and mutation failures. Three new core
mutations fail for their intended startup/manual/directory reasons. Two browser
defects now fail the intended checks. Preserve the initial missing-startup
survivor: its marker matched old call arguments; the corrected assertion reads
actual snapshot text/value. Reuse unchanged Chapter 12 behavior/mutation evidence
with its stated limits. Independently ran `TestGUIDebugStartup` on the submitted
source: PASS, 0.282 seconds. It crosses real CLI, registry, MCP, jobs, history and
rendering with external stdio/HTTP fakes; no mocks. Broader submitted core/GUI
test/vet and selected race receipts pass; there is no reason to duplicate them.

Reviewed retained live captures as well as the student's corrected assessment.
Anthropic and Gemini 3.8 changed the real/effective font size, observed refreshed
state and unloaded; final requests remove executable browser declarations and
automatic context. Anthropic's prose incorrectly claimed tools remained, which
motivated the truthful unload-result wording. Both OpenAI attempts are **task
failures**: first a wrong selector without repair, then a successful guided GUI
change/unload followed by an erroneous forbidden file read and false narration.
The latter still demonstrates the GUI capability but is not autonomous task
success. No third paid attempt is needed to conceal that limitation. Live runs
precede the documented small manual/message corrections; no final-source live
rerun is claimed.

## Comparative size and teaching assessment

Counts are net changes against each edition's own predecessor, from tracked
Go/JavaScript files, excluding blank and comment-only lines from active counts.
Comment-only `//` and block-comment lines are counted separately; physical totals
include blanks. No source moves occurred. Nested example modules are unchanged;
evidence scripts/generated logs are excluded. The two new skill data files total
23 physical lines; README prose is separate from implementation.

| Initial chapter delta | Active | Comment-only | Physical |
| --- | ---: | ---: | ---: |
| Student production Go | +43 | +4 | +47 |
| Student Go tests | +105 | +8 | +114 |
| Student JavaScript | +2 | +2 | +4 |
| Original production Go | +129 | +24 | +170 |
| Original Go tests / JavaScript | 0 | 0 | 0 |

The smaller production delta principally reuses Chapter 12's already working
owner-based lifecycle rather than duplicating connection/dispatch callbacks.
The added integration test is justified by actual startup and unload boundaries;
its length does not count as production bloat. StartupSkills' useful exported
field documentation and adjacent sequence/freeze comments explain the important
choice. The low incremental Go comment share (4 of 47 nonblank added lines) is
mostly mundane flag parsing, not unexplained architecture; inherited ownership
documentation remains. No padding or arbitrary quota is requested. The current
speech correction is incomplete, so neither these comparisons nor passing checks
constitute acceptance yet. Preserve this initial decision when appending round 1.

## Correction round 1: accepted with the recorded grader exceptions

Reviewed the full revised student response and stable correction
`95fcd2c10559554f6281a89a63b050bd14edd39b`. All 93 manifest file hashes match the
submitted tree. The only product change is `gui/web/gui/app.js`: retain the human
typing cause from input provenance, combine that with speaking, and reset it on
submission. Its adjacent comment explains why field contents cannot substitute
for provenance. This is the small current-scope fix requested, without a new
speech subsystem. Compared with the initial submission it adds three active and
two comment-only JavaScript lines; production Go and other boundaries are unchanged.

Independently rebuilt the actual GUI and repeated the original native-speech
interleaving. The [round-1 result](../../solutions/evidence/ch13/reviewer-round1-speech-probe.json)
now records empty pause with `MCP submitted` still in the field after actual native
speech completion; the original retained result records `paused` in the same
situation. Empty-input speech also releases correctly. The retained script and
result identify the exact correction revision. One reused cleanup-check label
still says “releases stranded pause”; the artifact annotation clarifies that
pause was already empty before clearing on corrected source. No observation was
rewritten to manufacture a pass. Direct GUI/browser processes were shut down and
the reviewer's scratch tree removed.

Read the coder's expanded native regression and final restored-source result:
all seven controls pass, including human whitespace preserving pause through
speech, actual other-tab explicit unpause, subsequent MCP submit, and final speech
completion with empty pause/input. Three restored defects fail cleanly for the
intended reason: field-derived typing strands programmatic input after speech;
ignored human typing leaves genuine input unpaused; missing submit reset leaves
an empty submitted field paused. The submitted GUI tests/vet and JavaScript syntax
check pass. These focused results and source inspection resolve the sole blocker.

The two setup failures remain recorded. Closing the second control tab after its
explicit-unpause observation isolates the intended single-tab speech regression;
it does not establish why two simultaneous native speech queues failed to drain
within the earlier bound or prove general multi-tab speech reliability. No wider
speech claim or speculative fix is required for this chapter.

**Accepted with exceptions after one correction round.** Original grading remains
**0/100 FAIL** on its recorded earlier source; both narrowly justified exceptions
above stand. The initial source failure, corrected-source proof, masked grader
coverage, and both OpenAI task failures remain distinct. No original grader or
paid demonstration was repeated, and no current runtime behavior is waived.
Coordinator owns the accepted checkpoint/export and subsequent chapter release.
