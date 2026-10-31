# Edition 2 Chapter 13 / original Chapter 14 review

Independent reviewer: `enhancement_reviewer_resumed`, October 10, 2026. Initial
submission `4a7816ae1dd203f1fc7ea4904965f58ad2a9cb92` extends accepted original13
`2e804b1` (now Edition 2 `ch12`). Read the entire mandatory
`.agents/skills/ensemble-coding/SKILL.md`, governing carryover/workflow, original
297-line Chapter 14, and completed [student submission](ch14.md). The original
chapter revision/hash and supplied harness contract are recorded there. Compared
both editions against their own predecessor: current `main` versus accepted
original13, and original `solutions/ch14` versus `solutions/ch13`.

## Initial decision: one lifetime correction required

The new blind peer does not finish promptly on EOF during `sleep`. Its EOF
handler closes the browser but leaves the awaited Node timer alive. The virtual
client closes the transport and then waits for that peer before Agent shutdown,
so even its 90-second request context cannot bound this final process wait.
No arbitrary sleep cap or new timeout framework is required; the existing owner
shutdown must actually terminate its owned peer when the session ends.

Independently reproduced through the actual GUI and peer, without a provider
request: initialized idle peer exits **125 ms** after EOF; a peer in a **4000 ms**
sleep exits **3865 ms** after EOF, essentially the remaining timer duration.
Both exit zero, and the positive control rules out general launch/teardown delay.
The [result](../../solutions/evidence/ch14/reviewer-peer-eof.json)
includes revision and source hashes; the adjacent `.mjs` preserves the probe.
The latter observation is not a permanent hang claim; source establishes that
longer requested sleeps retain the same unbounded final wait.

This is the sole confirmed blocker. Correction round 1 is released for that
small lifetime fix, a real pending-sleep shutdown regression and a normal-sleep
positive control. Preserve actual sleep while the session is alive. No paid
repeat, original-grader repeat, core change or generic driver machinery is needed.
All reviewer browser/direct processes are stopped; the harness is released.

## Scope, architecture and teaching comparison

The two delivered pieces match the chapter: speech processing and a blind
listener that must use it. `SpeechBuffer` retains its actual Speech parent;
filtering, queue ownership and transcript remain in the optional GUI module.
The blind peer exposes only hear/type/submit/wait/sleep and rejects guessed DOM
operations at dispatch. The actual public Go client constructs an Ensemble root
and Agent with no builtins or skill tools. This additional small entrypoint is
justified: using the ordinary skill-enabled CLI would violate “nothing else.”
The provider credentials do not enter its browser subprocess. No core API,
callback service bundle, transport rewrite or future fleet feature was added.

The speech module separates unfinished per-part buffers from the existing native
queue, rather than creating a second speech pipeline. Whole finals reuse the
renderer's streamed-part distinction; error/tool boundaries flush real tails;
automatic tool results stay silent. Transcript recording precedes audio bypass,
and entries retain their source. Existing human-input provenance and Agent-wide
explicit-unpause behavior remain in place. Browser observation is not a claim
of subjective intelligibility, pronunciation or listening comfort.

Counts below are net changes in tracked files against each edition's predecessor.
Active lines exclude blanks and comment-only lines; comment counts include
whole-line `//`, block comments and support-script `#` lines; physical counts
include blanks. No generated logs, installed dependencies or lockfile expansion
is counted as application code. Nested earlier examples are unchanged. The
speech extraction is accounted for net across app.js and speech.js, not counted
as an all-new queue. HTML changes in the old solution are separate markup.

| Chapter delta | Active | Comment-only | Physical |
| --- | ---: | ---: | ---: |
| Student production Go (virtual client) | +95 | +4 | +103 |
| Student production JavaScript | +230 | +20 | +256 |
| Student JavaScript browser test | +503 | +7 | +510 |
| Student executable harness support | +148 | +5 | +154 |
| Original production Go | +413 | +171 | +634 |
| Original production JavaScript | +364 | +143 | +545 |
| Original Go tests | +168 | +17 | +206 |
| Original executable/script support | +336 | +138 | +532 |

The original delta also repairs earlier settings, PTY and renderer behavior and
adds WebSocket routing/logging; the student already has those earlier obligations
and does not need to reproduce those changes. Smaller counts alone therefore do
not establish improvement. The new client remains direct flag/config/process
wiring; the longer browser test protects actual required boundaries with external
provider/synthesis fakes. No mocks replace the Agent, jobs, speech transformation
or browser runtime. Comments are below the approximate teaching target, but the
important fence ordering, owner, provenance, transcript and capability reasons
are explained. New Go symbols are private; no undocumented exported Go member was
introduced. No filler or arbitrary line quota is requested.

## Grader, tests and actual demonstrations

Original grading completed at **95/100 FAIL**: all new checks **95/95**, inherited
original13 parity **0/5**. Its recorded build predates completion of the blind
client and later support changes. I independently retain the earlier narrow
exception: the original13 fixture omits its skill kind and its parity machinery
also depends on implementation names. Current speech/persona behavior is not
waived. The grade does not independently prove the blind capability restriction,
all renderer branches, or audio quality, and the book's table differs from the
actual original check weights. No failing score is relabeled a pass.

Verified all 102 submitted manifest hashes and read affected GUI test/race/vet,
formatting and final real-browser receipts. Reviewed source and the retained
targeted assertion failures. The initial whole-final mutant genuinely survived:
plain provider responses still produce synthesized display deltas. The repaired
test preserves real Agent finals while dropping transient deltas at the external
WebSocket boundary, reaching the previously untested renderer branch. It then
catches omitted final speech; duplicate speech is caught separately. Persona
registry, dispatch, real typing/waiting, transcript cursor and consuming-client
capability mutations fail their intended checks. Added paragraph/tool/error-tail
mutations protect their distinct flush boundaries. Initial setup failures and
the mutation-cleanup timeout remain recorded; a timeout after the desired
assertion was not counted as a clean kill. That cleanup work did not cover the
pending-sleep case found above.

The retained native Chrome run passes the seven inherited speech/typing/provenance
controls, including actual next MCP submission. Its second control tab closes
before single-tab speech, so it does not settle the earlier unexplained concurrent
native-queue observation. The deterministic synthesis fake checks transcript and
gate behavior; native callbacks supply separate real-browser evidence. No human
listening judgment is claimed.

Read the three live summaries and their source-aware observation method. Real
listeners use the five permitted capabilities and report actual returned
transcript entries, not matching text in their own call arguments. Gemini 3.8
hears both thinking and response sources. Independently correlating each raw
`tool_returned` with its `tool_called` identity reproduces the summaries' exact
2 Anthropic, 3 OpenAI and 11 Gemini transcript entries. Anthropic's target never chooses the
requested fruit and the listener honestly reports its absence; OpenAI's target
uses three sentences rather than two. These are retained task limits, not reasons
for another paid attempt or evidence that speech filtering failed. Live binaries
precede the documented cleanup/formatting edits; no final-source paid result is
claimed. Acceptance awaits the sole lifetime correction, within the existing
three-round maximum; coordinator owns the eventual top-level `ch13` export.

## Correction round 1: accepted with inherited grader exception

Read the full revised student response, source and regression at `d01e846`;
independently verified all **103** submitted file hashes and modes. The production
fix only shares the peer's close-browser-and-exit handler between EOF and SIGTERM.
It ends the owner's lifetime, including its timer, without shortening a live
session's requested sleep or changing core, Go production or speech processing.
The nearby explanation now states the actual timer-lifetime reason. Production
JavaScript grows by one active and two comment-only lines against the initial
submission; the focused new test is 254 active / 7 comment-only / 263 physical
lines, separate from production. Its direct peer and actual-client controls are
justified by the two observed cleanup boundaries, without a testing framework.

Repeated the original independent real-GUI/peer probe on corrected source:
idle EOF exits in **126 ms**; EOF during a **4000 ms** sleep exits in **127 ms**,
both zero, compared with the initial pending-sleep **3865 ms**. The
[round-1 receipt](../../solutions/evidence/ch14/reviewer-round1-peer-eof.json)
preserves exact revision/hashes and observations; its adjacent script retains
the invocation. Reviewer processes and scratch are cleaned up.

Reviewed the new real-browser/public-client regression and raw baseline/mutation
results. Normal 150 ms sleep, idle wait, idle EOF and ordinary client completion
pass. A real 60-second MCP sleep yields the existing job running callback; the
fake provider then ends the actual model turn, and the real client exits within
the observed post-answer bound. Restoring the old EOF handler fails both the
pending-sleep peer bound and that actual-client exit bound, with normal controls
passing first and no unrelated cleanup failure. The final restored baseline
passes both subchecks (three Node records including their parent); syntax and
whitespace checks pass. These are meaningful lifetime failures, not arbitrary
timing substitutes for unrelated behavior. No further broad or paid run is needed.

**Accepted with the inherited original13 grader exception after one correction
round.** The sole current-scope blocker is fixed. Original **95/100 FAIL**, its
earlier build provenance, the initial mutation survivor/cleanup failures, live
target-content limits and absence of a human sound-quality judgment remain
unchanged. No current speech or persona requirement is waived. Coordinator owns
the direct top-level Edition 2 `ch13` export and any subsequent chapter release.
