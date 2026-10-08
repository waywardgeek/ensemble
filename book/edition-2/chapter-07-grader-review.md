# Chapter 7 independent acceptance plan

Contract-first milestone, 2026-10-07. This reviewer reloaded the full coding
skill, architecture ledger and Chapter 7 before writing the checker. The current
contract includes the incremental watch-projection paragraph at `2f105ce`.
No Chapter 7 student code exists or has been inspected. The role previously
reviewed Chapters 5 and 6, including their historical comparisons; it has not
opened the historical GUI answer for this new acceptance plan. Student coders
receive public requirements and results, never checker implementation.

The complete contract remains required. These are separate distinguishing groups:

| Group | Independent evidence and controls |
|---|---|
| Atomic watch | Hold before/after the actor snapshot cut at partial/final/tool transitions; snapshot plus strictly later revisions has neither a hole nor duplicate; include joining after begin before first delta |
| Window and ownership | More than 100 renderable events, omitted prefix and earlier-call result; mutate returned buffers; no HTTP or credential/config leakage; generation reset discards partial snapshots |
| Incremental projection | Fixed-size fragment doubling shows append-oriented cost; materialize owned snapshot at watch cut; preserve full identity, channel and bound accounting |
| Pause/admission | Two public registrations and two sockets with independent typing/speaking causes; pause observation precedes applied ack; exact barrier excludes later tool admissions; existing work and controls continue |
| Capacity/lifetime | Saturate watch and outgoing socket queues separately by count and encoded bytes; valid exact positives; close generation without dropping finals; other client and reliable handle work; deterministic selected-recipient/teardown race |
| Wire/projection | Exact Host/Origin/JSON/UTF-8/message/command-ID limits; correlated refusals; one serialized writer; recursive typed signature stripping with argument keys preserved; trace stages match actual queue/write/failure boundaries |
| Browser cards | Actual DOM for partial/final identity mapping, window reset, running versus terminal jobs, long-text expansion, accessible controls, focus and scroll preservation; malicious payload must actually reach every tested renderer |
| Speech | Mock queue A/B, cancel, then C and stale callbacks; typing survives speech transitions; replay silent and finals not repeated; separate real synthesis/audio evidence with browser/platform identity |
| Public composition | Two Agents with independent watches/registrations; reuse public Go and browser components from different layout; core builds with GUI module absent; inspect actual owners/imports/logger chains |
| Prior behavior | Retain Chapter 6 deterministic gate and CLI/protocol/plain/replay coverage; new pause observations do not alter old protocol output; initial live scope retains original source identities |

Every deletion or negative starts with a passing fixture and must reach its
intended assertion. Malformed setup or a prior rejection is not evidence for a
later guard. Examples include valid oversized JSON for the byte limit, a sender
known to have selected the departing recipient, and an XSS payload confirmed
present in the rendered card before asserting inert behavior. Public Go adapter
names will follow the student's reviewed ownership plan; the checker will not
invent method spellings or architectural requirements.

## Initial executable command

```sh
python3 scripts/edition2/accept_ch07.py /absolute/path/to/ensemble-gui
python3 -m unittest discover -s scripts/edition2 -p test_accept_ch07.py -v
```

Python `websocket-client` is required by this local checker. It launches the
published `--port 0` command in an isolated temporary workspace with explicit
fake model configuration. All model requests are counted by a local rejecting
endpoint; this milestone expects none. It never reads real credentials.

The initial executable contains 15 checks: empty snapshot wire shape;
two-client pause/count/ack/disconnect; a correlated idle hint refusal followed
by usable pause controls; four Host/Origin refusals; an exactly 65,536-byte valid
subscribe command; six malformed/over-limit cases; and absence of model HTTP.
(The listed totals are verified from the generator before publication.) It
uses ordinary WebSocket client framing under [RFC 6455](https://www.rfc-editor.org/rfc/rfc6455.html);
it neither specifies the student's server library nor adds one to the core.

Six assertion-control tests pass. They include positive empty snapshots and
pause changes, then targeted generation/watermark, null-range, credential,
cause-loss and acknowledgement-order failures. A deliberately absent server
(`/usr/bin/true`) fails at startup as expected, with a retained receipt. That
negative proves only the launch guard. There is no claim of a passing GUI
runtime or end-to-end positive transport fixture at this milestone. Initial
checker integration can reveal adapter/test defects; preserve those findings
separately from student defects.

This is a partial command, not a complete Chapter 7 score. Atomic public watch,
window seeding, admission barriers, both capacity queues, deterministic teardown,
projection, actual browser and speech tests remain to implement against the
published seams. Required real all-provider browser/CLI/public demonstrations
and independent historical comparison follow the initial source and student
experience freeze. Historical graders and earlier independent checks are
unchanged.

## Public API integration milestone

The initial five public groups now run with:

```sh
python3 scripts/edition2/accept_ch07_public.py solutions/edition-2/main
```

The fixture derives from the published watch/pause contract and adapts only the
student's public `Watch`, `WatchSnapshot`, `PauseRegistration` and `PauseState`
spellings. It reuses this reviewer's Chapter 6 local HTTP fixture helpers. No
historical GUI implementation was read. A disposable copy of the current student
worktree passes all five groups under the race detector, with exact source and
checker hashes in [the initial receipt](checkpoint-evidence/ch07-public-initial-pass.json).
This is integration evidence for unfinished source, not an immutable Chapter 7
acceptance checkpoint. The initial fixture compile failure used `Accepted`
instead of the established `ControlAck.Interrupted`; its [failed receipt](checkpoint-evidence/ch07-public-adapter-failure.json)
is a checker adapter mistake, not a student defect.

The groups establish independent registration causes, no-change revisions,
idempotent close and refusal afterward, two-Agent isolation; joining after begin
before any fragment and after `Hel` on all three adapters, then receiving `lo.`
and one accepted final; shallow and nested returned-snapshot ownership; exactly
100 retained renderable events with omission/range metadata and lifetime usage;
256 queued watch observations accepted with overflow at the next item, a healthy
peer and reliable completion continuing; and a complete model-proposed write
held by two independent registrations, released only after the last cause or
interrupted without its file effect. All network requests are local fixtures.

Coordinator review clarified the initial wire checker: actor publication must
precede the pause API return, but asynchronous socket delivery need only carry
the eventual matching observation revision/counts. The checker no longer
requires an unpublished observation-before-ack socket order. Its positive
control accepts either transport interleaving; the targeted negative changes
the observation revision and still fails. This corrects a private assumption,
not a weakening of the actor's published ordering requirement. Direct actor
publication-order inspection/control remains part of the unfinished audit.

The remaining plan above still applies, including encoded-byte capacity,
registration retention after idle close, invalidation of an unrecoverable
projection, deterministic selected-recipient teardown, socket saturation,
recursive typed projection, browser cards/accessibility, speech, immutable
mutation controls and retained Chapter 6 behavior. The new public CLI module
`example.com/ensemble/cli` is a permitted client of the public root, not an
implementation spoke. Existing CLI recovery assertions must follow that moved
public client rather than enforce their former `cmd` package location.

## Expanded deterministic coverage before the student freeze

The public runner now has seven groups, adding actual two-Agent watch identity
isolation and an idle-child lifetime check. The latter keeps Agent alive and
uses a finalizer plus a nonpublishing mailbox boundary: closing a watch must
release the child even if no later observation arrives. It reproduced retention
through the unused slice tail, then passed after the coder cleared removed
slots. [Expanded public receipt](checkpoint-evidence/ch07-review-public-expanded.json),
[initial retention failure](checkpoint-evidence/ch07-review-idle-watch-failure.json)
and [repair](checkpoint-evidence/ch07-review-idle-watch-repaired.json) keep their
own copied-source hashes.

```sh
python3 scripts/edition2/accept_ch07_components.py solutions/edition-2/main
node scripts/edition2/accept_ch07_browser.cjs solutions/edition-2/main
python3 scripts/edition2/audit_ch07_mutations.py solutions/edition-2/main
python3 scripts/edition2/audit_ch07_browser.py solutions/edition-2/main
python3 scripts/edition2/ch07-review-scaling.py solutions/edition-2/main
```

Five component groups cover real encoded-payload watch capacity (exact 128 MiB,
drain and refill, aggregate plus one, and a single oversized item), a controlled
publication-before-pause-ack barrier, exact projection capacity followed by
invalidation and restored availability after model end, a selected watch
recipient closed before its send, and recursive GUI projection. The large-byte
fixtures omit race instrumentation; ordering and GUI projection use it. The
explicit `is_error:false` fixture initially exposed an omitted display field;
the coder fixed only the projection. The passing component and mutation receipts
bind that repair. The first transient diagnostic retained command output but
had no immutable source map, as its record states.

The browser command uses installed Chrome through the external pinned Playwright
tool directory; `CH07_BROWSER_TOOLS` can select that directory. It copies and
hashes static assets before loading them. Its six groups use actual DOM,
keyboard, focus and scroll behavior with controlled WebSocket and speech
objects. They cover partial/final card identity, two-Agent/operation separation,
window replacement and earlier-call labels, hostile text reaching preview and
expanded renderers, the required Markdown subset, long-card keyboard expansion,
job report/terminal distinction, cancellation and stale speech callbacks, replay
silence and final deduplication, incomplete snapshot abandonment, uncertain
prompt submission, and the actual Page's typing/speaking reconciliation. The
Page case found that a rejected pause promise could hide prompt acceptance
uncertainty and leave an unhandled rejection. The coder now owns both promises
immediately and gives the prompt result priority; the same case passes.
These are real-browser component checks, not paid model or audible speech runs.

The wire checker now has 23 cases. Published clarification `1bba1bc` distinguishes
transport-malformed closure from correctable command errors with usable IDs;
the latter must return correlated errors and remain usable without HTTP. New
controls include missing/empty/non-string IDs, exact 4,096 accepted command
capacity and refusal of the next command. The 65,536-byte positive and valid
65,537-byte oversized JSON remain distinct. Inspection of the peer's raw close
frame avoids `websocket-client` masking that frame with a failed automatic
close reply. That exposed an actual empty close reason from the earlier binary;
the latest independently run binary passes all 23. Both binary hashes remain
in their respective receipts; the earlier failure is not relabeled.

Nine implementation deletion controls pass, each starting from its relevant
positive and requiring the precise failing leaf set and diagnostic. Seven
browser deletion controls likewise require the exact failing group set. They
remove watch retention release, count/byte bounds, projection invalidation,
pause publication/admission, late-send closure, opaque stripping, false-flag
preservation, safe text rendering, stale-callback fencing, replay silence,
uncertainty priority, atomic snapshot staging, full identity and provisional
replacement. The fixed-fragment projection benchmark doubles 64 KiB to 128 KiB:
its median cumulative allocation ratio is approximately 1.91; a deliberate
repeated-copy mutation raises that to 3.81. This is a local allocation control,
not a universal timing or peak-memory claim.

The new prior-CLI adapter preserves every assertion and exact deletion-failure
set from Chapter 6 while selecting either the historical `cmd` layout or the
new public `cli` layout. It passes the unchanged `75bd14d` positive and both
negative controls. The historical checker itself remains untouched. The new
immutable gate is being assembled at `accept_ch07_gate.py SOURCE_COMMIT`, with
all required embedded assets retained and hashed. It includes the coordinator's
separate actual Connector lifetime probe. No complete immutable Chapter 7 gate,
live demonstration, historical comparison or final acceptance is claimed yet.

## Immutable deterministic gate and retained-check correction

Source `da162e821369784b37cd6a6544610429af5f1cbe` is accepted for the
33-group deterministic gate by composition of two preserved receipts:

- [Original immutable gate](checkpoint-evidence/ch07-review-full-gate-initial.json):
  32 groups passed; the retained Chapter 5 structural category failed, leaving
  that retained score at 90/100. All 98 source files, including embedded browser
  assets, were extracted from the committed source and bound in the receipt.
- [Correction supplement](checkpoint-evidence/ch07-review-retained-supplement.json):
  the identical 98-file source passes the retained seven categories at 100/100.
  The original failure remains a failure; this is a targeted replacement result,
  not a rewritten green full-run receipt.

The old source heuristic classified a literal `fmt.Errorf` sentinel and a
`go:embed` filesystem as mutable session state. The new Chapter 7 adapter
recognizes immutable declaration forms and checks production files for direct
writes or address escape, including other files and importing packages.
Fourteen targeted controls include renamed/aliased values, a local shadow,
missing embed directive, dynamic formatting, assignment and address escape.
Original channel/counter rejection remains protected. Both original and new
source checkers pass the accepted Chapter 6 runtime `75bd14d`. The historical
checker and its assertion fixtures are unchanged; the separate adapter retains
all behavioral assertions and the seven category weights.

Reproduce the focused correction with:

```sh
python3 scripts/edition2/ch07-review-retained.py book/edition-2/checkpoint-evidence/ch07-review-full-gate-initial.json
```

For a fresh complete run, `python3 scripts/edition2/accept_ch07_gate.py SOURCE_COMMIT`
now selects that adapter and also records nested checker source hashes. The
accepted original groups include wire, public watch/pause, component bounds,
actual Chrome DOM/speech-double/Connector/Page behavior, implementation and
browser deletion controls, append-cost scaling, actual socket lifetime controls,
retained Chapter 6 behavior and CLI overflow, all eight module vet/test pairs,
and the headless build with the optional GUI absent. These deterministic results
do not establish real audio, live provider use, historical comparative quality
or final manuscript acceptance. Those remain separate gates.

## Draft manuscript review while live work proceeds

The [bound draft proofread](checkpoint-evidence/ch07-review-draft-proofread.json)
reads the full chapter and author records at `71da5b5`, full current voice and
procedure, and the corresponding first-edition Chapter 8 prose. It preserves
the old implementation boundary: no historical solution code was opened before
the initial live/student-experience freeze. The reviewer authored independent
graders and inspected the new implementation but did not author student code.

The chapter retains a concrete human reason to hear and steer the work, the
radio/receiver explanation, Artifact lifecycle and the documented teardown
failure. The finite window, safe snapshot, per-client causes and cancellation
generation give those stories precise consequences. The new draft removes the
unsupported Eloquence and diagnosis-duration claims and keeps the actual spin
pending. Its 750-wpm motivation is supported by the dated GUI design document.
The current opener already limits what graders can establish; an overlapping
read initially reported its previous sentence and that finding was withdrawn.

One current-instruction correction remains with the author: lead the TL;DR and
outline with the complete immutable gate, distinguish the current 23 wire
checks from the initial 15, and retain those initial counts as historical
preparation. Hard prose lint passes. The two soft density warnings were read
in context; they do not justify deleting contract requirements or adding an
invented anecdote. This is draft acceptance subject to that instruction update,
actual-use reconciliation and final proofread, not final chapter acceptance.

The adapter's [affected-module checks](checkpoint-evidence/ch07-review-adapter-module-checks.json)
pass: outer-module `go vet ./...`, full `go test ./... -count=1` (including the
unchanged historical grader suite), clean formatting and Python compilation.
The legacy grader suite completed in 511.997 seconds; no legacy implementation
or fixture was edited.
