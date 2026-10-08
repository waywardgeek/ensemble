# Chapter 7 source research and evidence plan

October 7, 2026. Author-only research; do not give this file or its historical
links to the student. Drafting does not authorize Chapter 7 code before accepted
Chapter 6. No Chapter 7 build, browser session, audio result or grade is claimed.

## Sources read

Read full current voice, chapter-writing procedure, workflow, architecture,
new Chapter 6, and first-edition Chapter 8. Read the complete old Chapter 8
assessment in `book/astra-chapter-reviews.md`, the relevant forward map in
`global-review.md`, old Chapter 13's typing-pause and renderer findings, old
Chapter 22's send/teardown incident, and course policy P11. Read inherited
`internal/grade/ch08_checks.go` and `ch08_harness.go` in full. Read the accepted
new Chapter 5 GUI stub and its module declaration for the actual public seam;
do not infer the evolving Chapter 6 implementation from old answers.

Some broad history/search outputs were truncated; claims below use the
specific passages and commit descriptions subsequently read, not an assertion
that all historical source was loaded.

## Historical receipts and decisions

- `f1336dd6f8e733cbe52e9166fb5159ca13914b76`: records old explicit pause-bypasses-
  mailbox/shared-gate and configuration rulings. Coordinator accepts a different
  current mechanism: owned pause registrations and a responsive actor, with an
  acknowledged admission boundary. No sibling gate injection and no actor
  condition-variable wait. Historical configuration does not replace the
  second edition's environment-held credential contract.
- `e84190d4134f804dbe95d4a43018930d171fcabb`: replaced flat message-count replay
  with a durable event window and current partials. Keep those distinct sources;
  repair ambiguous window policy, global part IDs, unsynchronized reads and
  snapshot/tail races rather than copying the earlier implementation.
- `03bd18964695fc20aae4681bac067c74ae3dbab0`: browser replay clears panes and treats
  event range as informational. New subscribe resets a generation explicitly;
  it never fetches a range recursively or resends a lost prompt automatically.
- `3d7b7d19ea14c774b0f2480868b3d43af1ac523e`: moved safe tool-card text insertion
  and accessible retained content into old Chapter 8's teaching. The new version
  uses accessible expanders, not a title tooltip as the only recovery path.
- `352b5904aa06371339c547f8941cf8a2b8733793`: fixes send-on-closed-channel after
  snapshotting recipients. The commit describes two insensitive regression
  attempts and a mutation-sensitive enlarged interleaving. Teach one connection
  lifetime and a deterministic sender/teardown barrier; no claim that random
  repeated races or `select/default` protects a closed send channel.
- Old Chapter 13's input-clear passage documents a GUI that paused on input and
  failed to release when submission assigned an empty value without an input
  event. Teach explicit reconciliation for both typing and programmatic submit.
  Do not copy the later driver's special flag as a requirement for this GUI.

## Current primary documentation

Checked October 7, 2026:

- [Gorilla WebSocket documentation](https://pkg.go.dev/github.com/gorilla/websocket):
  one concurrent reader and writer; origin validation is the server's duty.
  Library remains a student choice; selected module versions must be pinned.
- [MDN WebSocket server guide](https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API/Writing_WebSocket_servers):
  browsers send Origin; nonbrowser clients can forge it. Local origin checks
  protect the browser boundary but are not authentication of local processes.
- [SpeechSynthesis cancel](https://developer.mozilla.org/en-US/docs/Web/API/SpeechSynthesis/cancel)
  and [speech error reasons](https://developer.mozilla.org/en-US/docs/Web/API/SpeechSynthesisErrorEvent/error):
  cancellation clears pending utterances and interrupts current speech;
  cancellation/error callbacks need owned generation handling. These pages do
  not substantiate the old universal tab-switch workaround. Verify actual
  browser/version behavior during implementation and keep any workaround scoped.

## Known inherited grader limitations

The old stream check tests presence of names and a few fields. Reconnection is
only tested after a completed turn. Its pause fixture checks no second dispatch
within a window, but does not establish multi-client ownership or a responsive
actor while paused. The GUI log checker looks for timestamp/direction substrings,
not strict JSONL or browser receipt. Frontend rendering, speech, XSS prevention,
identity across response boundaries and slow-client recovery are outside those
checks. Do not weaken historical tests; add new contract-derived checks with
positive controls and exact mutation failures.

## Required future evidence

Before paid use, the student publishes its ownership plan and feature-to-action
live plan. Bind actual built GUI and CLI programs, source commit and launches.
Retain browser screenshots with accessible descriptions, terminal records,
actual request/log facts and workspace artifacts. A WebSocket write receipt is
not browser rendering or heard audio. Synthetic speech callbacks prove queue
logic only; real synthesis requires a real browser receipt with its limits.

All three vendor paths need actual browser use and human CLI parity; current
Gemini scope is discovered 3.0 Flash or newer. Local-only exposure, origin
rejection, missing voice, output overflow and concurrency barriers use controlled
fixtures. Snapshot reconstruction must perform no model calls. Verification
checks identities before reading/replaying outputs, retains raw artifacts and
proves each negative control reaches its intended guard. Keep initial answer,
student review, post-run comparison, guided revisions and distinguishing tests.


## Proposed public choices and review boundary

Coordinator accepted the pause/watch direction before drafting. The complete
proposed Chapter 7 now supplies owned pause causes, acknowledged admission,
atomic snapshot/tail, safe public fields, explicit 100-event window, bounded
queues and disconnect recovery. It also proposes a reusable public CLI client
for the combined server's optional terminal, rather than duplicating the CLI.
The 128 MiB encoded queue byte limit accommodates ordinary JSON escaping of a
Chapter 6 bounded response; overflow still explicitly invalidates the watch.
Core logical content bounds and per-connection encoded transport bounds differ.

This draft needs complete public-contract review, an independently published
new-checker invocation, and the accepted Chapter 6 predecessor before student
release. Source research and an outline are not those gates. Scoped prose lint
has no hard failures; soft density/person-gap warnings receive a reading pass,
not fabricated anecdotes or padding.


## Narrow coordinator review, October 7

Coordinator reviewed the complete draft and accepted shared CLI composition and
the fixed event window. Two missing contracts were taught before code: public
pause changes now have exact aggregate/count payloads and actor revisions,
including changes that keep paused true; browser projection explicitly descends
into all typed part lists without stripping similarly named tool-argument keys.
The snapshot watermark includes pause transitions. Active operation metadata
covers subscribe-after-begin-before-first-fragment without replaying a fabricated
begin event.

The projection fixture respects Chapter 2's restriction on result children:
text/blob/redacted only. It tests a bound signature on a valid nested result
text part and raw nested opaque bytes in a standalone accepted response part,
instead of grading an invalid opaque result child. These are fictional local
fixtures, not new model observations. No implementation or live claim added.
