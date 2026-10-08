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
