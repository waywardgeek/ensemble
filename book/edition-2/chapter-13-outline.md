# Chapter 13 outline: test the speech channel

Initial preparation, October 8, 2026, frozen at `bb6e04c`. New Chapter 13 maps to first-edition Chapter 14,
“The Channel Nobody Tested.” The original proposal below preserves the preparation chronology. The
[full contract draft](chapter-13.md) now records the coordinator choices below;
[validation](chapter-13-validation.md) remains incomplete. Chapter 12 has an accepted design at `2d4ea47`, with no
implementation or actual-run acceptance yet; see its
[gate record](chapter-12-validation.md). No Chapter 13 implementation, grader
or live result is claimed. Historical sources belong to the author/reviewer
handoff, outside the cold student's permitted inputs.

**Through-line stake:** a reader who depends on speech needs the failure and
the answer to reach that channel; a model that finds the answer elsewhere
cannot establish that the reader received either.

## Voice and story choices

Keep the original chapter's listener experiment, with its limitation attached.
The historical account asks a restricted observer to identify a failed command.
It gets the answer, but the assistant's narration supplied it: tool results
were intentionally silent. A correct answer can test the wrong route. Explain
what the observer was allowed to receive before presenting its conclusion.
Do not reproduce the old transcript as a new run or claim that the model heard
native audio. The existing account is a historical source, without a recovered
raw session binding for its quoted interaction.

Keep the fenced-code diagnosis. The account first suspects stream boundaries;
a single-chunk control reproduces the defect. Splitting at newlines had broken
the fence before the filter could recognize it. This gives the reader a useful
experiment and makes filter order matter. The source repair is recoverable;
its historical test count is not a new measurement.

Use the fire-alarm inspection analogy sparingly: checking the labels does not
exercise the alarm. Preserve Bill's documented dependence on spoken reasoning
as the human reason for testing this surface. Do not turn a listener profile
into a claim to simulate blindness, or claim this chapter establishes screen
reader compatibility. The proposed body stays in third person. It should
move from a failed channel to a controlled experiment, then explain ownership
and the exact public checks. Avoid repeating the accessibility motivation in
every section.

## Teaching order and the concrete fact in each section

1. **An answer from the wrong place.** The historical restricted observer
   answered correctly without establishing tool-result delivery. Distinguish
   the wanted information from the route being tested.
2. **Choose what may speak.** Tool results and replay stay silent under Chapter
   7. A later historical actor mislabeled a result as assistant text, so a
   correct speech filter still read it. Test provenance at the feed boundary.
3. **Filter before losing structure.** The fence failed even without streaming.
   Teach a bounded normalization grammar, chunk-invariant output and explicit
   terminal flushes, before prescribing individual substitutions.
4. **Record the shared output path.** A speech log must observe the same
   normalized request passed to the selected output service. Queued text,
   native admission, callbacks and audio recordings establish different facts.
5. **Give the listener only its channel.** Freeze a separate listener's rights
   before construction. A prompt telling a fully privileged observer to ignore
   the DOM cannot establish isolation.
6. **Exercise the whole application.** The historical direct-module grader
   missed an error that arrived on the wire but never reached speech. Drive
   the public GUI and actual browser through independently supplied fixtures.
7. **Taking it for a spin.** Eventually use real models through human CLI,
   browser and public embedding, preserving the distinction between recorder
   mode and native audio. This section remains unwritten until receipts exist.

## Scope carried forward

Chapter 7 already requires per-Page queues, sentence buffering, finalization
without duplicate speech, silent tool results/replay, captured part identity,
local cancellation and the combined typing/speaking pause predicate. Chapter
8 adds captured preferences and same-origin/storage-bucket native arbitration.
Chapter 12 adds actual MCP-over-WebSocket observation and control, bounded
telemetry, scoped mounts and human draft refusal. Preserve those contracts.

In particular, typing pauses subsequent Agent admissions; it does not acquire
a new meaning of automatically canceling another Page's speech. Escape with
nonempty input retains its Chapter 7 behavior. A recorder cannot bypass Actor
pause gates, create a second tool dispatch path or widen an Agent's tools.

New work proposed here is a precise text-normalization contract, observable
speech delivery and lifecycle records, an explicitly restricted listener
consumer, and a whole-application harness. Audible terminal-failure reporting
would be an explicit extension of the present automatic-speech selection;
ordinary tool-result text would remain silent. The full contract must distinguish
an infrastructure failure announcement from a failed command's result.

## Ownership and data proposal

| Owner | Proposed responsibility and route |
| --- | --- |
| BrowserApplication | Own the existing SpeechService and Page children; select native or recorder output explicitly at construction for a test application. No mutable module singleton. |
| Page | Own normalized-text buffering, part deduplication and its logical queue; retain mount/Agent identity, captured preference revision/rate and cancellation generation. Reach shared speech through its actual parent. |
| SpeechService | Preserve FIFO/native lease arbitration. Own a bounded speech journal and the output adapter it creates. An adapter receives its actual service parent and can reach diagnostics. |
| Speech journal | Record service-accepted owned facts with monotonically scoped record identities; index by Page/mount and runtime Agent identity. It describes speech, not authoritative conversation state. |
| Optional GUI server | Expose an explicitly selected listener logical endpoint through the public MCP transport seam. Authenticate scope through application binding, not caller-supplied arbitrary Agent IDs. |
| Separate listener Agent | Receive a frozen ceiling for hearing records and narrow input/submission actions for one selected target. Use ordinary public construction and completion APIs; no new general sub-agent machinery. |

Shared Go values and public interfaces belong in the appropriate common/public
seam; implementation behavior remains in its owning spoke or optional GUI
module. Browser classes retain equivalent actual-parent ownership. Do not
inject the recorder into every component as a sibling dependency or log speech
by re-rendering the conversation on the server.

A record should distinguish at least: normalized queue acceptance, native
submission after the lease is acquired, native start, terminal end/error/cancel,
and recorder completion. Preserve source kind and complete utterance identity.
No `spoken` Boolean should conflate these stages. A native end callback is not
proof of intelligibility; a WAV is a separate artifact with capture provenance.

Proposed retention is a bounded journal with sequence cursors and explicit
consumer-lag gaps, plus an opt-in bounded file recording for diagnostics. A
slow listener must receive an explicit gap rather than a fabricated contiguous
transcript. Page close fences callbacks and records the disposition of owned
work without clearing another Page's records. A service/root close joins its
writer and preserves inherited shutdown order. Full drafting must settle exact
record/byte limits, exhaustion behavior, file failure semantics and whether
closed-Page records remain available until ordinary journal eviction.

## Listener scope and evaluation

Keep Chapter 12's exact five-tool target endpoint unchanged. Prepare a separate
listener endpoint and discover its aliases before constructing the listener
Agent. The coordinator accepts this direction as a working design choice,
with explicit application-selected scope and caller-granted access. Do not
reproduce the historical post-construction registry removal.

The listener may consume the selected speech journal through a bounded cursor
and issue narrowly scoped input/submission requests. It must have no DOM,
artifact, full-conversation, file, shell, general GUI snapshot or automatic
context path to the target's answer. Its endpoint returns control outcomes,
not incidental target answer text. Mount and target identities can come from
the application; discovering them must not require a privileged GUI snapshot.
A guessed disabled alias fails at dispatch as well as disappearing from offers.

Use independently generated fixture information available only through the
allowed recorded speech for the hearing task. Verify the listener's actual
request bodies, grants and observations. Denial controls should leave it unable
to obtain the answer when speech delivery is removed. A separate fixture puts
a marker in a tool result and different prose in an assistant response, proving
both exclusion and inclusion without requiring the listener to reveal content
that the contract deliberately withholds.

Call the fast mode a transcript-only listener exercise. It demonstrates that
normalized text reached a substituted output adapter and that the listener used
that channel. It does not assert that a model perceived acoustic output or that
a blind person can use the application. A native-mode run independently tests
actual admission, callbacks and captured audio.

Do not use Chapter 12's programmatic `gui_input` to claim that a human typing
pause was exercised. That path deliberately differs from trusted human input.
The browser harness can produce actual input events for the inherited pause
case; the listener's narrow programmatic control keeps its own documented
semantics. Same-Agent pause constraints still apply. Use an explicitly scoped
independent observer for active-state inspection when necessary.

## Behavior candidates to settle before the full contract

- Specify a small, deterministic markdown-to-speech grammar: preserve words
  across deltas, recognize fenced blocks before sentence/line splitting,
  normalize a single newline to space, retain paragraph boundaries, strip
  selected markup and expand a bounded identifier grammar. Print literal
  fixtures for one chunk and multiple split patterns. Do not promise all
  Markdown or linguistic sentence segmentation.
- Define terminal flushes at accepted response end, tool boundary and failure,
  plus cancellation of provisional material. Streaming and unstreamed forms
  of the same accepted text must queue it once. Empty output must not keep a
  speaking cause. Preserve captured preference behavior when autoplay turns off.
- Bound the parser's pending text, including an unterminated fence or a word
  exceeding a normal chunk. Choose and publish visible overflow behavior;
  neither unbounded buffering nor arbitrary word chopping is an implicit rule.
- Specify which user-visible failures receive an automatic announcement and
  how retry, disconnect and duplicate terminal events avoid repeated messages.
  Preserve silent tool results even when their `is_error` field is true.
- Define journal cursor identity, per-entry and total-byte caps, exact lag/error
  behavior and whether the listener consumes acceptance or completion records.
  Proposal: it explicitly requests a phase, with recorder completion as the
  default fast-test channel. A canceled native utterance is never reported as
  fully heard merely because its text was queued.
- Keep journal export opt-in and selected-scope. It can contain sensitive
  assistant text. Do not dump raw credentials, opaque provider payloads,
  unpublished thought fields or every other Agent's transcript for convenience.
- Keep journal and native work out of resumable conversation/session state.
  A new listener's stable aliases/definitions use inherited identity rules;
  transient mount bindings require explicit rebind. Historical replay remains
  silent and never restarts an output adapter.

## Checks and actual-use plan

Use public behavior and a declared harness boundary rather than private
JavaScript method names. Keep simple pure normalization tests, but also drive
the actual browser-delivered modules over the real GUI transport. Honor the
student’s supported module format instead of evaluating source strings.

The independent checker should supply response text and tool fixtures, inspect
strict bounded records, and test declaration/dispatch isolation. Positive and
negative controls should distinguish normalization, provenance, unstreamed
finalization, terminal failure wiring, pause causes, cancel races, shared native
ownership, journal gaps and listener restrictions. A harness that exits early
without ever launching its child is not a passing process-cleanup test.

| Feature | Planned action and receipt |
| --- | --- |
| Human CLI parity | On each supported provider, submit and observe a real answer/tool turn through a PTY; preserve usage and request bindings. |
| Browser normalization | Ask for bounded prose with code/identifiers; retain original deltas, normalized entries and visible cards, including model noncompliance. |
| Restricted listener | Run a separate real-model public consumer with only the selected listener endpoint; retain its actual offers, requests, hearing cursors and answer. |
| Native output | Exercise actual browser synthesis and local cancel with captured audio, callbacks and identities; label unavailable voices or platform constraints. |
| Shared ownership | Two Pages and cooperating tabs retain their own queues; cancel/close one without ending another's native ownership. |
| Faults and limits | Deterministic endpoint failure, chunk splits, overflow, lag and late callbacks; label these local fixtures. |
| Shutdown | Close normally and at timeout; independently verify all owned browser/process children are gone, with a proven-running canary. |

All-provider coverage and bounded request budgets must be reviewed before any
paid launch. Preserve initial implementation and actual evidence before the
historical comparison. Node or fake-DOM checks remain useful local fixtures;
they cannot replace the actual browser or native-audio gates.

## Decisions requested from contract review

The owner direction and separate endpoint are agreed working choices. Exact
journal retention/export semantics, listener phase semantics, parser overflow
and failure-announcement selection remain proposals. Settle them before a
student handoff; no runtime or grader should guess them. The full chapter can
then publish literals and bounds without becoming a new general accessibility
framework or a speech-engine portability project.


## Full-draft decisions after outline review

The coordinator accepted the outline direction and supplied working choices;
these are not new Bill rulings. The full draft now fixes the following:

- Automatic normalization has an 8,192-byte pending budget and bounded active
  parsers. Overflow abandons the affected unqueued remainder visibly, preserves
  admitted speech and permits the next part to recover. Literal chunk-invariant
  fixtures specify the narrow grammar.
- SpeechService owns a 4,096-record/8 MiB ring. Public scoped cursor export has
  explicit gaps; the external harness owns bounded file output. No continuous
  browser-to-server diagnostic writer is added.
- Recorder_complete is the listener's fast channel. Native queued/submitted/
  start/end/canceled/error facts remain distinct, without a claim of hearing.
- Automatic failed-turn announcements use static safe text once per owned live
  request and only with autoplay. Model transport failure produces that same
  notice. GUI disconnect retains inherited cancel/status behavior; reconnect,
  tool failures, intentional cancel and synthesis failures stay silent.
- Explicit manual full-card speech is preserved. Journal entries may contain a
  scalar-safe prefix, exact byte/omission counts, digest and truncation flag.
  The digest covers the exact adapter text without journal normalization. A
  truncated required listener entry is an incomplete evaluation. An oversize
  manual positive must prove the adapter receives the whole text.
- listener_attach/listener_attached explicitly select the separate endpoint on
  the existing view. Endpoint-kind/view/mount identity lets target and listener
  coexist. Existing frame/detach protocols and shared capacity count both kinds;
  listener close cannot stop target watch or native speech.

The old unresolved-proposal section above remains an account of the initial
outline. Exact draft rules are now available for independent review, with no
student release or implementation claim. The historical story choices survive;
full native-audio and transcript-only demonstrations await actual receipts.
