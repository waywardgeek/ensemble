# Chapter 7 source research and evidence plan

October 7, 2026. Author-only research; do not give this file or its historical
links to the student. The dated research and preparation entries below precede
the implementation. Current gate status belongs to the
[validation record](chapter-07-validation.md); the reconciliation entry records
subsequent evidence without relabeling earlier plans as completed runs.

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

The complete public-contract review is accepted. The initial partial checker
invocation is now published; the accepted Chapter 6 predecessor remains required
before student release. Source research and an outline do not replace those gates. Scoped prose lint
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


## Published initial checker

The coordinator accepted the complete contract in `chapter-07-review.md`.
Checker `a5a07ba` publishes
`python3 scripts/edition2/accept_ch07.py GUI_BINARY`, now printed in the TL;DR.
Its 15 initial local checks cover transport, snapshot, pause, origin and malformed
commands. This author update does not run the checker or claim a student pass.
Full §7.9 coverage remains required, including public, concurrency and browser
checks being developed as the interfaces emerge. Chapter 7 implementation and
live evidence still await the accepted Chapter 6 baseline and student handoff.

## Temporary author reconciliation after the initial source freeze

The coordinator reassigned the available `/root/coder_ch04` thread to Chapter 7
prose while the original author could not resume. This author implemented
Chapter 4 and has that source exposure; it did not author Chapter 7 code and
is not the independent reviewer of that code. Its edit scope is this chapter,
outline, evidence and student-feedback record.

Read the complete current `book/voice.md` and chapter-writing procedure,
architecture, complete Chapter 7, its outline/evidence/review/feedback/validation
records and the student's `main/evidence/ch07/student-review.md`. Read
`book/gui-design.md`'s opening and chapter arc to check the personal workflow
and GUI scope. No Chapter 7 implementation or grader source was read for this
prose pass.

The initial source is preserved at `da162e8`; evidence support repair `ba902b7`
retains the same runtime. Local results and speech feasibility are described
in the coordinator's validation record. Actual paid browser/PTY/public runs
and historical comparison are still pending at this entry. No new successful
spin, audio transcription, human listening or final acceptance is inferred.

Revision `0f05359`'s personal motivation and examples remain. The introduction
now assigns separate panes, draggable dividers and saved preferences to
Chapter 8, while Chapter 7 supplies a single scroll of cards and transient
speech/input controls. The closing paragraph promises the last-100-event
window and an omission count. `book/gui-design.md` supports 750-wpm TTS and
macular dystrophy, but the consulted evidence does not substantiate the new
Eloquence attribution, reading-speed comparison or crash-diagnosis duration;
those claims are narrowed. The documented send/close incident remains intact.
The explicit speech scope restored at `4a2e95c` is retained unchanged.

Scoped check: `go run ./cmd/lintprose book/edition-2/chapter-07.md` passes every
hard rule (5,623 prose words). Soft warnings remain for 20 negation forms and a
2,033-word person gap. The reading pass retained technical refusals that define
the wire and ownership boundaries; reconnect paragraphs now name the reader's
action and uncertain prompt acceptance directly. Paragraph endings vary between
consequences, state rules and the next operation. No anecdote or padding was
added to satisfy a count. Fenced code and table rows are byte-identical to the
pre-edit chapter. Runtime and grader files were neither edited nor rerun by
this author; their independent gates remain in the validation record.


## Initial actual-use author reconciliation, October 8

Initial implementation and student experience are frozen at `8cc87f2`.
Runtime is `da162e821369784b37cd6a6544610429af5f1cbe`; support/launch binding is
`ba902b7322a9cb5dd71c8352d45d6975663409c0`. This appended account supersedes
pending prose as a current description; it does not relabel any older receipt.
The coordinator's validation record remains the gate authority.

Actual additional reads for this pass: complete current voice and writing
procedure after compaction; current Chapter 7, outline, feedback and validation;
review's final editorial findings; Chapter 4's normal/kill lifecycle teaching;
student `live-results.md`, `live-receipts.json`, initial experience in
`student-review.md`; all three combined `terminal.txt` transcripts; their GUI
prompt/hint trace records and sanitized launch metadata; all audio audits;
selected reconstructed replay receipts. Inspected all three selected actual
screenshots and their adjacent page text, and read the scratch file bytes.
No runtime or grader edit was made by this author. Prior Chapter 4 coding
exposure and absence of Chapter 7 implementation authorship remain disclosed.

The student drove nine actual sessions on October 7 Pacific / October 8 UTC:
combined browser/terminal, standalone plain terminal and public two-Agent
embedding on each provider. `live-receipts.json` records 9/10/9 admitted prompts
and 15/14/15 HTTP requests for Messages/Chat Completions/generateContent;
44 original request bodies match the separate replay derivatives. The browser
subsets have 12/11/12 raw request and response files. The models were
`claude-sonnet-4-6`, `gpt-4.1-mini-2025-04-14` and
`models/gemini-3.8-flash`. Those are dated observed identities, not undiscovered
future defaults. Chapter 7's actual spin now uses these receipts.

The following counters are quoted from the combined terminal's EOF output,
**before** the later browser-only prompt. They are not whole-matrix totals:

| API | Input | Cache write | Cache read | Output |
|---|---:|---:|---:|---:|
| Messages | 35,358 | 0 | 0 | 928 |
| Chat Completions | 2,747 | 0 | 14,592 | 513 |
| generateContent | 40,055 | 0 | 0 | 3,022 |

All three scratch files contain exactly `BROWSER-CH07\n` (13 bytes), SHA-256
`914b491dcca0f5bd16dd194247ba237f418cf09e18e37288c9cd6c29db28f824`.
The author recomputed those hashes. The Chat Completions initial intention-only
answer and subsequent explicit correction remain in print; the earlier no-tools
instruction is a possible explanation, not proven causation. All three next
model requests contain the hint marker; only Messages and generateContent
printed it. The live page and tool records distinguish the interrupted turn,
still-running process, and deliberate later kill.

Actual image references, under `main/evidence/ch07/`, are:

| Image | Visible evidence | SHA-256 |
|---|---|---|
| `browser-anthropic-r1/browser-26.png` | Connected, refused pending call, interrupted r4, typing cause held | `ccbf6da3434819d17600d81fc5cf2b30fc5bcc2d1fcc8d3b82c2b8ead35a711a` |
| `browser-openai-r1/browser-25.png` | Connected, refused pending call, interrupted r5, typing cause held | `b125e072bbf8c220bc3165d9cd9e53ec70872d8a3ee44edb043ceda38709132d` |
| `browser-gemini-r1/browser-25.png` | Running job3/17bytes, supervision call, typing1/speaking1 | `90120dd5c7416aab83483be92467fa2c9311c9fadfe06d24830c4cfc92e08c0b` |

The author recomputed the three image hashes. Section 7.8 embeds the first and
third with descriptions of their visible viewport. Adjacent text receipts retain
offscreen cards; neither screenshot is a complete-history claim.

Chrome 154.0.8037.93 on macOS 26.3.1(a) produced actual speech audio. The
per-run `audio-audit.json` binds each WAV to its Chrome PID and records a silent
first PCM second at the -91 dB analysis floor. The answer windows at 2–6 seconds
have means -21.3/-19.7/-20.7 dB. They establish produced audio, without a
transcription or human-listening claim. The first generateContent selection
spoke a user card and is retained separately from its deliberate model-answer
selection. Asynchronous capture startup explains why the requested wall-time
lead is not claimed as two silent PCM seconds. No Bill use is implied.

Local controls remain separate for exact watch cuts, hostile content, capacity,
uncertain submission and stale speech callbacks. The TL;DR/outline now publish
`python3 scripts/edition2/accept_ch07_gate.py SOURCE_COMMIT`, with the original
32/33 plus corrected retained structural group explained. The wire checker now
has 23 checks. The initial 15-check account above remains a dated entry, as do
all failed local receipts. The author has not rerun runtime gates or made paid
requests in this reconciliation.

## Post-comparison contract repairs published before code

The student's lifecycle note exposed an actual teaching omission, confirmed by
root and reviewer: live cards handled `job_killed`, but the snapshot filter
followed the printed list that omitted it. Sections 7.3 and 7.6 now include
both inherited terminal facts. The coder read the clarification and confirmed
it resolves the difficulty; a narrow filter repair and reconnect regression
remain separate from the initial frozen use.

Root also accepted two explicit ownership corrections from browser review.
Sections 7.6/7.9 require Page, Scroll and Artifact to remove owned DOM handlers
on idempotent close and cancel pending reconnect callbacks, so replacement on
the same root cannot trigger old actions. Sections 7.1/7.7/7.9 teach an explicit
application/document root owning the shared native speech service; child Pages
retain their own state and reach the service through their parent. FIFO requests
submit one native utterance. Cancellation removes only the owning Page's work,
uses native cancel only for its active utterance, and preserves peers and stale
callback fences. These are coordinator working choices under the architecture,
not new Bill rulings. The coder must record the affected ownership plan before
implementation; initial receipts are not evidence that these repairs already
worked.


Author checks for this scoped pass: `go run ./cmd/lintprose
book/edition-2/chapter-07.md` passes all hard rules at 6,388 prose words.
The 19 negation forms and 2,059-word person gap received a reading pass;
the dense stretch defines snapshot/wire boundaries, while the spin now returns
to actual actions and consequences. The cut pass removed redundant attribution
and listening disclaimers. Paragraph-ending review found varied mechanisms,
examples and consequences rather than repeated short conclusions. Existing wire
and projection code fixtures remain byte-identical; deliberate artifact changes
are the current checker command, actual spin blocks/table, and two added
acceptance-table controls. All six local manuscript links resolve. Three file
hashes, three screenshot hashes and all four audited WAV hashes were recomputed
read-only. `git diff --check` is clean. No runtime test result is claimed from
these prose checks; independent proofreading and revised live reconciliation
remain open. The student confirmed all initial-live feedback dispositions.


## Narrow revised-review prose follow-up

The independent reviewer reports targeted R1–R4 checks passing at `9ba7855`
and independently reproduced the coordinator's shutdown ordering finding: end native speech admission
before disposing Pages, otherwise canceling one can start a peer's queued work
during application shutdown. Section 7.7 now states that order. The TL;DR adds
`python3 scripts/edition2/ch07-review-revisions.py SOURCE_COMMIT` as a supplement
to the full retained gate. The author did not execute those runtime checks.
A grammar repair changes “without claiming transcription, human listening” to
“without claiming transcription or human listening.” Initial source/live identity
remains `da162e8`/`ba902b7`; final revised evidence and prose reconciliation are
still pending at this entry.


## Revised actual-use reconciliation, receipt checkpoint `341f15d`

Revised runtime/support is `9ba7855b31a5eb134819602b35eb9acb277f6342`.
The student's committed receipts and final teaching review are frozen at
`341f15d80a80bfecf7ccceaf7b7840de63c20f7f`. The author read the complete revised
`live-results.md` and `live-receipts.json`, all three per-run derived audits,
sanitized launch metadata and terminal records, raw browser action/prompt/queue
records, accepted response records and request budget fields. The author compared
all three raw trace/launch/audit files with their frozen Git bytes. It visually
inspected the actual OpenAI `browser-12.png` and checked adjacent rendered text.
The current voice/procedure remained in context; no runtime or grader was edited.

The revised public consumer ran on October 8 Pacific, in Chrome 154.0.8037.93.
Each API used two prompts/two HTTP requests and exited 0, within the authorized
two-prompt/four-HTTP ceiling. The source-bound replay receipts reconstruct all
six original bodies. These are new public browser sessions, not new standalone
human CLI sessions or a relabeled rerun of the original 44-request matrix.
Initial all-feature CLI/browser/public receipts retain their previous identity.

| API/model | Input, first/second | Output, first/second | Stop reasons |
|---|---|---|---|
| Messages `claude-sonnet-4-6` |66 /66|96 /103|end_turn /end_turn|
| Chat Completions `gpt-4.1-mini-2025-04-14` |67 /68|77 /92|stop /stop|
| generateContent `models/gemini-3.8-flash` |56 /56|508 /508|MAX_TOKENS /MAX_TOKENS|

All revised cache-write/read counters are zero. The generateContent returned
identity is `gemini-3.8-flash`; requested identity retains `models/`. Both raw
request generationConfig objects cap output at 512. Raw candidate tokens 18/20
plus thinking 490/488 give normalized output 508 each. Visible text has 17/19
whitespace-delimited words, not the requested 70; the accepted empty text parts
and opaque replay material remain in the raw logs. No retry or successful length
compliance is claimed. These are distinct real answers for the browser ownership
exercise despite the example's tight model budget.

All three runs closed/remounted the idle second Page while first-Agent speech
was active, submitted the second prompt through its replacement, and later did
the reverse during second-Agent speech. Exact Chrome PIDs are 9992/11054/12483.
The revised queue evidence is action spans 12–17, 8–13 and 9–14 respectively:
A starts, B queues, B is canceled before starting, B is requeued, active A is
canceled, then B starts once. A keeps its unsent correction and typing cause.
The Messages first attempted span 7–10 followed A's natural completion and is
explicitly excluded as proof of overlapping queued cancellation. Re-selecting
retained cards supplied the intended sequence with no new model generation.

Receipt timing matters: generateContent action 13's immediate body-text snapshot
still says typing 1/speaking 1. The subsequent raw pause_changed for agent-1 at
11:37:37.756 UTC says typing 1/speaking 0 before action 14. Root identified this
as asynchronous observation timing. Section 7.8 uses the OpenAI 12 screenshot,
which actually shows the settled 1/0 state, and makes no contrary claim about
that Gemini text snapshot. The screenshot SHA-256 is
`82bb47d06edd9f151e27496726e9d83463a68bd293d1afd20a24b91a398b48a2`.
Its alt text describes the visible causes and unsent correction; each Scroll's
viewport crops some cards even in the full-page image.

Six WAV files are 48 kHz stereo IEEE float32, with exact-browser-PID capture.
The author recomputed their hashes and matched all per-run audit records, plus
all three raw trace and launch hashes. The recorded first PCM second is digital
silence; the derived analysis reports a -200 dBFS floor, unlike the earlier
ffmpeg analysis's -91 dB floor. Speech-window RMS values are -19.93/-21.06,
-21.08/-21.42 and -20.15/-20.57 dBFS. Different analysis floors are not a measured
change in the browser's silence. The student's first Python wave inspection
rejected float32; explicit RIFF fmt/data parsing supplied the reported metrics.
The chapter claims produced audio, without transcription or human listening.

The public snapshot/killed-job browser regression reuses the actual initial
job event as a local replay control. The larger/fragmented message repair,
whole-root shutdown and stale native callbacks remain controlled tests. The
independent reviewer reports final 33/33 full-gate groups and 12/12 correction
groups at `17aa661`; these are kept separate from the new live sessions.

The author corrected the earlier discovery attribution in feedback: the
coordinator first found the root-close ordering defect and independent review
reproduced it in Chrome. All revised student teaching feedback has a disposition;
the student already confirms the contract repairs resolve the comparative
ambiguities. Root's independent immutable live audit is accepted at `ffbad61`; the grader's
final proofread remains separate. The author read the complete revised section
of `chapter-07-live-review.md`: ten distinguishing controls, six exact replays,
145 unchanged committed session files and six independently reanalyzed WAVs.
Its ffmpeg floor is -91 dB; the student's -200 dBFS floor is a different
analysis convention. Initial audit `950cead` retains its original scope.


Final scoped author checks: `go run ./cmd/lintprose
book/edition-2/chapter-07.md` passes all hard rules at 6,898 prose words.
The 19 negation forms and 2,113-word person-gap warning were read in context;
the revised spin stays with actual typing, hearing/capture and view replacement,
without inventing a human story for the detector. A paragraph-ending pass retains
varied instructions, consequences and receipt limits. All nine local manuscript
links resolve, all previous fenced code fixtures remain unchanged, and
`git diff --check` is clean. The new screenshot is actual captured UI. These
checks supplement the separately accepted runtime/live audits; they do not claim
independent proofreading by the author.


Root's final receipt reading narrows the no-settlement claim to the idle-view
close/remount interval. Section 7.8 states that interval explicitly and retains
the generateContent utterance that ended normally later in its recording; it
never promises an entire recording without a settlement callback.
