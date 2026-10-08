# Chapter 8 source research

Author/reviewer-only material, October 7, 2026. Do not give these historical
sources to the fresh student. This is research for new Chapter 8 (old Chapter 9),
not a ready contract or a claim of implementation/live validation.

## Reads

Read the full current voice and chapter procedure in this author phase; read
all of old `book/chapter-09.md`, the full old CH9 checks and harness, the current
architecture ledger, current Chapter 7 handoff/owner sections and the global
map's settings/UI, GUI-testing and later integration lessons. Read the exact
old Chapter 13 settings-validation/zero-broadcast account and old Chapter 22's
settings closure/dead-value passage. Relevant old CH9 source history was read
with `git log` and exact diffs; no student source was edited.

## Historical leads and corrections

- `ca775b6` identifies the old chapter's self-use motivational addition. Keep
  the reader's practical reason to customize an interface; do not import an
  unmeasured productivity comparison or claim the reader already used it.
- `3d7b7d19ea14c774b0f2480868b3d43af1ac523e` adds explicit obligations after
  defects survived the original teaching: input/load validation, patch versus
  snapshot, and observable control state. The old phrasing that sparse patches
  need `omitempty` because zero is “unset” is itself insufficient. New teaching
  must track field presence separately so explicit zero/false remains writable.
- Old Chapter 13 records temperature -5 accepted, then a corrected zero missing
  from the broadcast while the browser still displayed -5. It records speech
  false omitted for the same reason. Its later measurement paragraph says zero
  was absent after a repair described as removing omission; that internal
  contradiction is a reason not to reproduce the measurement as proof. Use the
  mechanism and source-backed incident, with new independent fixtures later.
- `2b2fedbc8b50bc688c31ab823712553e589f8a3b` adds the fourth obligation: prove a
  saved setting affects execution. The old GUI saved max_tool_rounds=200 while
  two loops still used16. Preserve the missing-consumer lesson, not old loop
  architecture or default/tool-batch semantics. New Chapter 5 already owns turn
  decisions and snapshots configuration at turn start.
- The global map and old Chapter 22 identify settings closures bridging an
  unreachable owner and settings values with no consumer. GUI persistence must
  not become a second Agent configuration authority.

## Old checker coverage observed from source

CH9 checks compare sent theme/tts_speed values in the acknowledgement, a
nonempty current-settings object on subscribe, dark theme in a second client,
and dark theme after restart. They also invoke old CH8 parity. The harness does
not demonstrate explicit false/zero, wrong-type atomicity, failed writes,
versioned disk load, stale revisions, semantic browser control state, actual
speech behavior or an execution-setting consumer. No new deletion run was made
by the author, and source inspection is not a measured checker result.

Its executable layout, `--gui-dir`, handshake and old fake identity differ from
new Chapter 7. Retain its historical diagnostic identity; the independent new
checker must derive from published new teaching rather than force the student
to restore old wiring.

## Decisions to publish next

The outline proposes Server-owned GUI preferences and Agent-owned execution
settings as distinct domains, with actual owner interfaces and public commands.
The coordinator accepted that scope before the complete student contract:
persisted display defaults are shared, actual page speech/input/pause is locally
owned, and Agent changes remain actor-ordered. The full draft must define when
a preference update applies in each tab and preserve active-turn snapshots. Shared complete
snapshots must be safe public projections, never serialized credentials or
creation-only log identity. Missing values, explicit values, invalid values and
failed writes need different documented outcomes.

No new external browser/API claim has been made in this research pass. Current
platform behavior needed by the eventual browser contract will be checked
against official sources when drafting. Existing browser tooling availability
is coordinator evidence; it does not prove the new interface or audible speech.


## Full draft contract, following coordinator decisions

The coordinator accepted shared future-autoplay semantics before drafting:
future enqueue uses each page's currently applied preference revision and rate;
already queued/current utterances keep their captures; local Cancel affects
only the owning page. No settings update impersonates a different registration.
The new chapter publishes exact multi-tab tests for those distinctions.

The draft includes strict patch/file validation, versioned full snapshots,
separate persisted revisions and update commands, stale/busy/write-failure
outcomes, bounded subscribe ordering, creation-only paths and one writer per
domain. Agent policy keeps the existing default 16/model-request count and
captures its effective value at turn activation. Its new optional historical
turn payload is explicit; old absent fields retain the old policy meaning.
These are proposed detailed contract choices for coordinator review, not claimed
implementation behavior. The initial outline's open decisions are now answered
in the draft; no Chapter 8 solution or passing checker exists.

The author read Chapter 5's activation/configuration rule, Chapter 3's exact
sixteen-request/final-batch rule, Chapter 7's pause/watch/wire/card/speech rules
and the current workflow during this drafting pass. The new chapter deliberately
extends only execution policy during an active turn; it does not silently lift
other configuration restrictions. No current browser-vendor claim or new paid
model observation was introduced. Scoping and local examples rely on the
published predecessor contracts rather than an old implementation copied forward.

## October 8 author reconciliation

This later pass supersedes the initial pending implementation summary above
without rewriting its chronology. The reassigned author previously served as
Chapter 4 student and Chapter 7/9/10 author, and authored no Chapter 8 runtime.
Read the complete current voice, chapter procedure, architecture, Chapter 8,
student review and independent code review, then checked the concrete receipts
below. No student code, grader or validation ledger was edited.

### Historical correction and external specifications

The initial source research treated the old Chapter 13 speech-false account as
a lead. Independent comparison's whole frozen Chapter 9 client control now
narrows the claim. At old source `fd93089b8f53bdde42f9846c9aa36d5d0e6a9051`,
`solutions/ch09/web/gui/gui.js` coerces an absent `tts_enabled` to false. The
[retained control](checkpoint-evidence/ch08-review-historical-settings.json)
shows true→false and an unchecked checkbox. Its temperature control remains -5
when zero is omitted. That local DOM/socket control is not an actual historical
browser/model session. The new opener keeps the supported display defect and
explicitly corrects the unsupported speech outcome.

Verified current primary sources on October 8, 2026:

- [Web Locks](https://www.w3.org/TR/web-locks/): cooperating contexts within an
  origin/storage bucket, exclusive named ownership and promise-bounded lease.
  Chapter 8's service lifetime, cancellation and unsupported behavior remain
  the application contract, not a claim of cross-profile native coordination.
- [ECMAScript InternalizeJSONProperty](https://tc39.es/ecma262/multipage/structured-data.html#sec-internalizejsonproperty):
  a primitive reviver context can carry its original source token. The chapter
  requires a verified lossless facility and a visible refusal if unavailable;
  it does not assert universal browser support or change numeric wire types.

R1 and T1 in [the comparison](chapter-08-code-review.md) now have explicit
teaching in §§8.3–8.4 and §8.7. The student response is recorded separately in
[chapter-08-student-feedback.md](chapter-08-student-feedback.md).

### Actual-run identities and checked claims

Initial student runtime is `bd5c05a62069266c6bc5bd3b1a859bc76c85f4f3`;
initial implementation, experience and runs are preserved at
`5f9684b07fc02081ca91fa48cceeb676f27d8051`. Native repair runtime is
`cd9de3e4ec6be4a560144bf46caf9ddd54302dc2`. Exact-counter runtime is
`a06d4f3c848685d81306166279df7c977d71bddd`. Final support `52c9561` changes only
the neutral replay comparator; final evidence checkpoint is `7f517d8`.
The launch manifests and binding files under
[main/evidence/ch08](../../solutions/edition-2/main/evidence/ch08/) retain
complete runtime, executable and support identities. Original log files were
explicitly committed unchanged at `0aad851`; their prior omission was an
ignored-file evidence defect, not another run.

| Runtime scope | Real model HTTP requests | Exact reconstruction receipt |
|---|---:|---|
| bd5c05a: all three CLI/headless runs and initial browser work | 21 | `reconstructed-bd5c05a/receipts.json` |
| cd9de3e: all three native browser tasks and public embedding | 24 | `reconstructed-cd9de3e/receipts.json` |
| cd9de3e: Gemini corrective interruption | 1 | `reconstructed-cd9de3e-gemini-interrupt/receipts.json` |
| a06d4f3: approved Anthropic text-only interruption correction | 1 | `reconstructed-a06d4f3-interrupt/receipts.json` |

Per-provider totals in `final-live-summary.json` are Anthropic 11 prompts/18
HTTP, OpenAI 9/14 and Gemini 10/15; one discovery per provider is separate.
The selected dated models are claude-sonnet-4-6, gpt-4.1-mini-2025-04-14 and
models/gemini-3.8-flash. Do not silently assign all 47 requests to final runtime.

Author reads included all three actual PTY terminal logs, all three headless
consumer terminal results, the native browser action logs and selected full
text, the interruption before/after text, three restart-stage persisted files,
replay comparisons, request-count receipts, audio analyses and exact launch
identities. In `anthropic-browser-native/requests/006.json`, the retained tool
history contains the first write of marker plus port=9090 and the later write
of marker alone. Final workspace files agree; OpenAI and Gemini keep the port.
The hint's temporary scope is observable in those bodies, not inferred from
an answer claiming success.

All three headless consumers show Agent 1 limit 1/request count 1/round_limit,
Agent 2 limit 2/request count 2/success, then independently restored policies.
The CLI first-success usage in §8.8 is read from each retained `/usage` result;
its later limit-one outcome preserves the paired file read without inventing
a second-request answer.

The final Anthropic consumer displays provisional `The`, then incomplete text
and an interrupted outcome. The no-tools refusals and exhausted initial prompt
cap remain; the one additional correction was explicitly authorized. Gemini's
late interrupt and subsequent paused-proposal correction remain distinct.

The OpenAI/Gemini endpoint-disabled replay supplements establish speech overlap
that their paid runs did not. Public Append rewrites only top-level admission
times; exact comparisons cover all other fields plus count/order/sequence.
The original overstrict failed assertions remain, and support-only 52c9561
adds a passing helper control plus intended refusals. Both supplements use
actual Chrome/native synthesis, not a fake speech callback; they do not produce
new model responses. Three fresh settings restart processes make zero model
calls and retain positive, explicit false/zero and confirmation files separately.

### Images, limits and checks

The author inspected the actual pixels of `openai-browser-native/browser-11.png`
and `settings-restart-confirm/browser-5.png`. The first visibly shows a paused,
accepted-but-unstarted read proposal; active1/next2 is in its adjacent full text,
below the captured viewport. The second visibly shows light theme, unchecked
autoplay and empty history; policy zero/default16 is below the viewport and
established by full text and persisted JSON. Captions and alt text keep those
limits. WAV analysis and native callbacks establish playback evidence without
claiming intelligibility, human listening or transcription.

The full reproducible checker command is now
`python3 scripts/edition2/accept_ch08_gate.py SOURCE_COMMIT`.
[Independent deterministic reconciliation](checkpoint-evidence/ch08-review-repair-reconciled.json)
accepts 52 current groups by combining 12 distinct affected checks on a06d4f3
with 40 still-applicable earlier cd9de3e groups; it is not a new complete
52-group run of a06d4f3. The historical CH9 diagnostic retains its incompatible
old-layout result. Neither check count substitutes for the final live audit,
student confirmation, independent prose review or coordinator checkpoint.

Author checks: scoped `go run ./cmd/lintprose book/edition-2/chapter-08.md`
passes all hard limits at 5,641 prose words. The remaining person-gap heuristic
was read against the actual sections: the wire and lifecycle contract repeatedly
names the reader's draft, competing edit, usable controls and local pause, then
returns to their observed tasks in §8.8. The cut pass removed a redundant spin
conclusion; paragraph endings were read in sequence. Literal schema fixtures
are unchanged. New command/transcript blocks are deliberate reconciliation
additions, with their source receipts above; no runtime grade was claimed for
this prose-only edit. `git diff --check` is clean.


Independent live audit accepted at `c84f46b`; see
[chapter-08-live-review.md](chapter-08-live-review.md). It checks all 47 model
requests and five WAVs with their separate paid/replay scopes. This closes live
evidence review, not the still-separate manuscript/student checkpoint gates.
The proofreader corrected the CLI build to name an output, removed adjacent
no-change repetition, clarified the layout opener and changed the budget's
ending from permission to actual turn completion. Outline reconciliation now
labels the initial plan historical and corrects its speech-false inference.

The student's response also prevented an author overclaim: browser exact
conversion covers settings/conflict/watch-envelope counters, while the server
preserves all projected numeric tokens. The current browser does not promise
lossless interpretation of every nested event/part number. Unsupported source
context refuses an unsafe counter when encountered; safe-counter-only browsers
can operate. Section 8.4 now states those exact supported boundaries. New skill
activation identities will receive their own explicit requirement in Chapter 9.
