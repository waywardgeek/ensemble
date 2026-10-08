# Chapter 13 preparation evidence

Author research, October 8, 2026. This ledger supports the
[outline](chapter-13-outline.md), not a completed exercise. New Chapter 13 maps
to old Chapter 14 under the current workflow and global-review map. Chapter 12
contract revision `2d4ea471587a5e458de92ae2d00e58042477515f` is accepted for
checker preparation; its implementation and live use remain pending.

The author thread previously implemented new Chapter 4 and subsequently served
as author for Chapters 7–12. It has historical source and grader exposure and
is not a cold Chapter 13 student or an independent reviewer of its own prose.
No old implementation has been copied into the new source. No runtime, grader,
legacy solution or frozen export was edited for this preparation. No grader,
model call or browser/audio experiment was run here.

## Source reads and authority

Read the complete current `book/voice.md`, `book/chapter-writing-procedure.md`
and `book/edition-2/architecture.md`, reloading voice/procedure after compaction.
Read the complete `book/edition-2/global-review.md`, including its October 8
correction of old Chapters 12/13 and its forward-lesson table. The procedure's
old Chapters 6–21 to new Chapters 5–20 mapping places old 14 at new 13; no
numbering change is proposed. The current workflow was searched for related
mapping terms. Bill/CodeRhapsody's new sandboxing work remains a pending future
lesson source, untouched here; it supplies no completed chapter-count claim.

Read the complete old `book/chapter-14.md` and `book/review-ch14.md`, and the
complete historical grader files `internal/grade/ch14_checks.go`,
`ch14_harness.go` and `ch14_run.go`. Read the full proposal recovered with
`git show caca911:book/proposal-ch14-revision.md`. It is historical evidence,
not a newly issued user instruction.

Read the current new Chapter 7 speech/ownership passage, searched Chapter 8
speech controls and consulted the accepted Chapter 12 design and gate record.
Chapter 12's full contract was read and revised in the immediately preceding
author task. These inherited contracts govern the new proposal; the old
chapter's differing registry, global-state and Escape examples do not.

Read commit metadata for the historical repairs below and the complete source
diffs for `d45f552` and `cf9af4f`. Read the complete `08c8780` documentation diff.
Targeted searches of old Chapters 15–17 and 20–22 located replay, provenance,
credential and lifecycle lessons; the global review supplies their whole-book
interpretation. Those later chapters were not reread in full for this task.
Distinguish this targeted reading from the earlier global reviewer's full-book
read.

## Historical chronology and limits

| Source | What it supports | Limit on a new claim |
| --- | --- | --- |
| Old Chapter 14, introduced at `3603183`, voice pass `994bf21` | Human dependence on speech, restricted-listener experiment, fence diagnosis and initial direct-module grader description | Its prose is a historical account; no raw listener session was recovered here. Its grader description predates the later log-based revision. |
| `116b23b08844ea8fcff51d0f67c8be2bda24ce6a` | Commit account of combined pause causes, error announcement, queue telemetry and cancel-generation fixes | Historical reported checks, not rerun results. New Chapters 7/8 already own most of this behavior. |
| `f6e32d25741c06918ae1ce7f118db9560662be09` | Historical listener removed DOM tools and added transcript/bypass paths | Post-construction removal conflicts with the new frozen ceiling. A transcript entry at enqueue does not establish native playback. |
| `064f1b188485f0449129599a2546c041e676d37b` | Commit account of fence filtering before splitting, single-chunk reproduction and newline treatment | Retain the diagnostic experiment, not an invented new test count or successful live rerun. |
| `2966081c52b54950c229078f8d0ba52e0a9443fc` | Historical unstreamed-final omission and repair | New Chapter 7 already requires final-only delivery without duplicate streamed text. |
| `caca911f3ba3a2b67c7163b7b82b4edca0c9c9e6` | Full revised-grader proposal records replacing private-module invocation with student harness and speech log | It explicitly withdraws the persona grading promise while retaining the persona as teaching/live work. Do not silently delete that deliverable. |
| `92352d5e2975bfa07a572314896a8e55c11b16e9` | Commit records the grader transition and direct `eval`/ES-module failure | Behavioral public checks should not pin unpublished internal method names. Historical scores remain dated reports. |
| `d45f552cb9615d07d3f000d1f249fa55748a96f1` | Source diff removes a second tool-result notification mislabeled as assistant `TextPart` | This is a provenance defect upstream of the speech filter. The diff proves the repair, not new audio measurements. |
| `cf9af4fd69a24fd1a355e243985262fce19793e7` | Source diff connects `turn_ended.error` to the GUI error path | A helper already existed; the missing event-to-helper call caused the silent failure described by the commit. |
| `46f6cf39f1e499a1a8f3e862a67c0a5b74be00d3`, `24f6e84bd72cf7a2b4e4e52525100c621264d5e7` | Metadata describes post-filter speech logging and a reference harness with real GUI modules, small stub DOM and real WebSocket | That harness was not an actual browser. Its reported headless voice failure is a dated environment observation, not a universal browser limitation. |
| `22307f32095334fbe5f8aafb39d1581104fe9372` | Metadata describes shipping a default speech log | The new edition must explicitly choose diagnostic retention and sensitive-content scope rather than inherit an unrestricted default. |
| `2c1a3793843e1a24bb0b58f20b29195dbe48c214` | Metadata describes leaked harness processes and a false cleanup control whose canary had never remained alive | Require a proven-running child and normal/fault cleanup. Historical leak counts and durations were not independently measured here. |
| `08c87802b8dfa1d27f2a25edf31f1b18217c39c0` | Documentation adds a two-utterance speech transcript: tool intent, then assistant narration of a file's content | This is a retained documentation receipt, without a raw source-bound audio/session artifact recovered here. It illustrates intended provenance but does not prove hearing. |

The initial `book/review-ch14.md` belongs to the earlier direct-module phase.
It reports passing mutants while acknowledging ungraded listener rules,
error speech and transcript behavior. Its bullet/newline finding is a useful
fixture lead, not proof that the same bug remains in current historical code.
Do not relabel either historical review as an audit of the second edition.

## What the inherited grader actually observes

The current historical checker has a zero-point harness precheck and nine
weighted checks totaling 100: discrimination 20, filtering 20, fragment
buffering 10, unstreamed speech 15, gate 10, errors 10, boundaries 5,
identifiers 5 and predecessor parity 5. This differs from the seven-check
node-era table still printed in the old chapter. These are source counts,
not results from a grader run.

The harness accepts a declared read-tool name, launches against a local fake
vendor, supplies planted files and reads JSON-lines speech records. This is a
useful public-boundary pattern. It does not certify that a real browser or
native speaker was used. The recorded prototype can substitute an output
recorder to make deterministic tests possible.

Several assertions need stronger distinguishing controls in the new checker:

- Discrimination uses a tool-result marker and a different assistant marker.
  The source comments' “same string twice” description is imprecise. Preserve
  both positive and negative provenance checks with independently seeded data.
- The buffering check limits utterance count and checks a substring; it does
  not by itself prove exact chunk-invariant text or every word boundary.
- Unstreamed speech checks marker presence, which alone cannot reject duplicate
  playback. The new check must count identity-bound delivery.
- The error scenario accepts any utterance. An unrelated opening announcement
  can satisfy it unless the failure itself has an independently required value.
- The gate check finds a typing pause and later resume in the log. It does not
  independently establish combined causes or actual tool-admission behavior.
- The log reader skips malformed/unrecognized lines. New evidence readers
  should use the published strict format and reject missing/truncated evidence
  rather than silently reducing it to a smaller apparently successful trace.
- The listener's absent DOM/answer privileges need a separate grant/dispatch
  and actual-request audit. A speech log cannot establish that isolation.

Preserve the historical grader and old baselines. An independent new checker
must be derived from the eventual new contract; private implementation names
and historical assertions do not become hidden student requirements. A future
shared grader change requires its own retained legacy checks and negative
controls. None has been performed for this outline.

## Whole-book dependency decisions

| Earlier or later feature | Consequence for Chapter 13 preparation |
| --- | --- |
| New 2/5/6: event provenance, Actor and streaming | Speech consumers preserve event/part identity; they do not synthesize assistant text from arbitrary result bytes or mutate conversation state. |
| New 7/8: speech, pause and preferences | Extend the existing Page queue and application-owned service. Keep native lease ownership, local cancel, captured rates and unsafe-counter protections. |
| New 9: frozen catalog and grants | Listener rights are selected before construction. No “blind” prompt compensates for a broad installed tool ceiling or an automatic DOM source. |
| New 10: persistence | Speech journals are diagnostic output, outside the single conversation writer. Session replay restores no queued utterance or native lease. Journal exports need their own explicit lifecycle. |
| New 11/12: public MCP and real GUI tunnel | Use a distinct listener logical endpoint and explicit scope. Keep the existing five-tool endpoint, physical socket lifetime, framing limits and request-effect deduplication unchanged. |
| Old 15/16: redaction and compaction | Changing context retention must not replay old speech or silently claim to erase already recorded audio/transcripts. Future retention policy must distinguish those stores. |
| Old 17: recall | Recalled content and newly spoken narration retain provenance; replay of a recalled entry is not new automatic speech. |
| Old 18/19: caching and credentials | Speech diagnostics cannot receive credential headers, opaque replay signatures or other Agents' route secrets. They are not provider usage or billing evidence. |
| Old 20/21: live integration and default skills | Exercise delivered modules, actual default wiring and real user controls. A helper, parsed config or synthetic fixture alone cannot establish the running feature. |
| Old 5/22: architecture repairs | Put recorder/output ownership and logger reachability on actual parents from introduction; no global queue, callback bag or injected sibling dependency. |

These are forward constraints and proposed checks, not early implementations
of later chapters. No redaction, recall, credential or sandboxing feature is
added to Chapter 13 merely because it supplies a relevant lesson.

## Coordinator review and open contract work

The coordinator agrees with separate listener ownership, explicit selected
scope and creation-time grants, keeping Chapter 12's endpoint unchanged. It
asks the outline to specify recorder ownership, per-Page identity within shared
speech arbitration, bounded retention/consumer lag and the distinctions among
queued text, native admission, callbacks and captured audio. The outline now
contains that proposal and leaves exact bounds and failure behavior for review.
These are working coordinator directions, not new decisions attributed to Bill.

Before full drafting, settle parser overflow, journal retention/export failure,
listener delivery-phase semantics and terminal-error announcement selection.
Then print fixtures and a feature/action/result matrix. The actual spin must
follow actual implementation and retained all-provider receipts. No historical
transcript or fake speech engine will be relabeled as a new native-audio run.

Concurrent Chapter 9 work remains higher priority. The coordinator reports
that `d8c7738`'s whole-record rule is implemented and that an initial offline
import mistakenly re-encoded a raw-valid record before checking its size. The
published read-original/write-emitted distinction already governs; await the
student's report before adding a teaching disposition. Its later spin commands
also need to use the same executable path as its TL;DR build. No Chapter 9
source or manuscript was edited during this preparation.
