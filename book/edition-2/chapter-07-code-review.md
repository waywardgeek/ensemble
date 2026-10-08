# Chapter 7 independent code comparison

Status: the initial comparison and targeted revision review are complete.
All four findings are repaired at `9ba7855`; the retained full gate, revised
live acceptance and final proofread remain open. This is not final chapter
acceptance.

## Scope and independence

The reviewer is `/root/grader_ch05`, also the independent grader engineer for
Chapters 5–7 and proofreader of earlier second-edition chapters. That grader
exposure is disclosed; the reviewer did not implement the Chapter 7 student
runtime. The student received findings and rationale, not historical source or
grader implementation. Historical implementation was opened only after the
initial source, real runs and student teaching experience were frozen.

The compared second-edition source is
`da162e821369784b37cd6a6544610429af5f1cbe`, main tree
`61debeccdddc18779c24bbbb6bd7a3a18f8caf80`. Support revision
`ba902b7322a9cb5dd71c8352d45d6975663409c0` preserves its runtime;
`8cc87f2` freezes the initial live evidence and teaching experience. The initial
deterministic acceptance is `043e6de`, including the separately recorded
structural-check correction. The coordinator's independent live audit at
`950cead` accepts 44 requests across nine initial sessions, with its documented
limits. This review does not relabel those runs as executions of revised code.

The first-edition comparison is Chapter 8, `solutions/ch08`, last modified at
`34ef50bf44a154eb06536e8b2544d37e08b504ba`, tree
`f6aa4a1ae6216a267728f71a62d34f03e626d8e6`. Later teardown correction
`352b5904aa06371339c547f8941cf8a2b8733793` was read as corrective evidence.
Current architecture and the published new contract remain authoritative.

The source review covered the optional module's Server, Connector, trace,
projection and executable; core public watch/pause and actor integration;
browser Connector, Page, ArtifactScroll, Artifact and SpeechQueue; and the
public browser embedding. Historical comparison covered its WebSocket hub,
replay, pause gate, trace, agent facade, rendering and speech components.
Required ownership, behavior, comments, usability and teaching were assessed
alongside the independent acceptance checks.

## Consolidated findings

### R1. Preserve deliberate job termination across reconnect

The initial `internal/llm/watch.go` renderable filter includes `job_ended` but
omits `job_killed`. The live browser handles the latter, so a connected card
correctly changes from running to killed. Reconnection loses that outcome and
restores the retained earlier running report. This is a teaching omission:
the original explicit renderable-kind list told the student to omit it.

Two independent controls establish the effect. The public watch control
re-admits the preserved real Anthropic events without executing their tools,
then adds 60 ordinary system messages to distinguish the 100-event window and
omission count. The initial snapshot reports seven omitted events where the
corrected contract requires eight. The actual Server/WebSocket/Chrome control
shows a running card, delivers the preserved killed event, observes “Job
killed”, then reloads and observes “Report delivered · running”. These are
local replays of actual neutral events, with no fresh provider generation.

The author published both terminal kinds in §§7.3 and 7.6 before the filter
repair. Preserve the separate meaning of interruption: it does not itself kill
jobs. Acceptance requires the corrected public window and two actual browser
reconnections, with a filter-deletion control restoring the intended failure.
Initial receipts are [public watch](checkpoint-evidence/ch07-review-killed-initial.json)
and [actual reconnect](checkpoint-evidence/ch07-review-killed-browser-initial.json).
Two earlier fixture setup failures are retained separately and are not student
defects.

### R2. Dispose browser listeners when their Page closes

The initial `Page.close()` closes transport and speech but retains its DOM
listeners. Reusing the same layout creates two handlers. In the independent
Chrome control, the first Page successfully submits `POSITIVE`; after closing
and replacing it, the old handler consumes `REPLACEMENT`, clears the textarea
and reports “Not connected”. The replacement submits nothing.

A reusable public component needs an explicit disposal boundary. Remove its
owned DOM listeners and timers, reject pending requests, and fence callbacks
that can finish after closure. Apply this ownership to nested Artifact and
ArtifactScroll listeners too. Closed guards alone can suppress visible work
while retaining the old owner through live elements; the revised control
therefore also records actual listener removal. The author published this
lifecycle requirement before the grouped repair.

The [initial receipt](checkpoint-evidence/ch07-review-lifetime-original.json)
retains the successful first submission and failed replacement. Acceptance
requires a functional replacement, no retained owned listeners, and a targeted
listener-removal deletion that fails for retention even when closed guards
still prevent the duplicate action.

### R3. Give shared native speech one application owner

The initial local queues are independent objects, but every Page reaches the
same document speech API and calls native cancellation on reset. Opening or
resetting an idle second Page can cancel the first Page's active utterance.
The [native API contract](https://developer.mozilla.org/en-US/docs/Web/API/SpeechSynthesis/cancel)
states that cancellation removes all queued utterances and stops current
speech. Per-Page counters do not partition that service.

The coordinator and author clarified the missing owner: an explicit browser
application owns Pages and the native service; Pages retain their own logical
queues and pause causes. Ready heads enter one owned FIFO. Cancel removes the
requester's work and reaches native cancellation only when that requester owns
the active utterance. Other Pages continue. The proposed public constructor
names remain implementation choices. The actual parent chain, rather than a
global registry or injected sibling broker, is the architectural requirement.

The [initial controlled Chrome receipt](checkpoint-evidence/ch07-review-speech-initial.json)
shows both idle-Page cancellation and overlapping native admission. These are
controlled native-API calls, not claims about audible speech. Revised controls
cover idle reset, pending cancellation, active cancellation, stale callbacks,
FIFO ready heads, a closed queue and independent typing causes.

During revision review, the coordinator found another transition within this
finding: whole-application close disposed Pages before closing native-service
admission. Canceling Page A could start pending Page B while the root was
shutting down. The reviewer reproduced that failure in Chrome and added a
whole-application close check. Admission must end before child disposal can
advance that queue. The pre-fix result remains preserved separately. Required
real-browser/provider demonstrations of revised shared speech are separate
from these deterministic controls and remain the coordinator's live gate.

### R4. Explain rejection for messages larger than the boundary fixture

The student found this defect during its own comparative repair audit. The
initial Connector handles 65,537-byte input explicitly but sets the underlying
WebSocket read limit to that value. A larger or fragmented message can trigger
the library's earlier close with no explanatory reason. The existing §7.4
contract already requires an explanatory close for all oversized messages.

The independent [bound initial receipt](checkpoint-evidence/ch07-review-size-initial.json)
passes the 65,536-byte valid command and 65,537-byte refusal, then fails both
131,072-byte single-frame and fragmented commands for the intended missing
reason. No model request occurs. The repair must retain bounded reading while
owning its explanatory rejection across message sizes and fragments. No paid
call is needed to validate this transport refusal.

## Design and quality comparison

The new answer improves several important first-edition boundaries:

- The old WebSocket hub resides under core `internal/ws`; its constructor
  accepts a pause gate, send closure and event-log pointer. The new optional
  GUI module reaches Agent services through public owners. Its executable
  composes one Ensemble and reuses the public CLI instead of maintaining a
  browser-specific agent loop. A headless build does not need the GUI module.
- The old shared Boolean gate and condition variable cannot represent two
  independent typing clients and put scheduling outside the actor. New
  registrations and acknowledgements put tool admission, cause counts and
  control responsiveness at one ordered owner. The same public mechanism is
  usable without a browser.
- Historical replay reads a growing log and updates a client sequence while
  broadcast can proceed. It also advances delivery bookkeeping after a full
  queue discards a frame. The new atomic watch cut, revisioned tail and staged
  browser reset explicitly represent what the client has received. Overflow
  closes the generation instead of implying that a lost final was delivered.
- The old send-channel teardown permits a selected sender to race closure.
  The new Connector uses a canceled lifetime and owned queues, joins workers,
  bounds socket operations and releases its watch/pause registrations. The
  comment explaining why its queue is never closed should remain: the later
  historical crash is concrete evidence for that unusual choice.
- The new typed display projection preserves part order and empty text while
  withholding raw opaque payloads and signatures. It keeps literal tool
  arguments intact. Full identities, safe DOM construction, accessible full
  text expansion and reader-controlled following improve on historical raw
  markup and narrower part keys. Historical UI brevity does not justify losing
  these distinctions.

The owned watch projection adds memory and protocol machinery, but gives a
precise bounded recovery promise instead of relying on unread history or
silently dropped broadcasts. Materializing a snapshot at the watch boundary
and accumulating live deltas incrementally is justified by the Chapter 6
allocation lesson. The existing independent scaling control protects that
choice. Keep the comments explaining the atomic cut, actor admission and
completion barrier; they describe correctness conditions, not syntax.

The new trace distinguishes queue admission from successful write and reaches
the application's logger when its writer fails. This is more useful than the
old direction-only trace with silently ignored write errors. It does not
promise to force-stop an arbitrary externally supplied blocking writer.

The comparison supports retaining the new architecture. The four findings
above concern observable recovery, resource ownership and public lifecycle,
rather than copying the old implementation or reducing line count. The shared
native service also makes later multi-Agent layouts safer without putting a
browser dependency into the core.

## Validation and remaining work

The prior six browser groups pass both the frozen initial source and the
current revised public constructor shape. All seven historical browser
deletion controls still distinguish their intended behaviors on the initial
source. In the new implementation, both the native service and local queue
fence stale callbacks; deleting only the local guard correctly leaves behavior
protected. That failed mutation attempt is retained as
[redundant-fence evidence](checkpoint-evidence/ch07-review-redundant-fence-control.json).
The strengthened control deletes both protections, retains the same expected
behavioral failure, and records the extra mutation explicitly.

Final revised acceptance must bind the grouped committed source, execute the
affected public/transport/browser controls and distinguishing deletions,
retain required prior gates, and review the actual ownership chain. The
coordinator owns raw live-receipt verification; this reviewer will separately
close the implementation review and complete the final manuscript proofread.
Initial failed and successful attempts remain at their original identities.

## Revised implementation review

The grouped revision is `9ba7855b31a5eb134819602b35eb9acb277f6342`.
The student's main repair batch entered `49ae919` during a shared-index
collision with an author commit; the final root-close repair and source freeze
are `9ba7855`. Neither commit was rewritten. This review binds the resulting
immutable 101-file source set, independently of commit-message attribution.

All four findings are repaired. The filter includes both terminal job kinds.
The browser owns listener removal and closes child components; callbacks and
request settlements cannot reuse a closed Page. `BrowserApplication` creates
its Pages and native `SpeechService`; Page queues reach the service through
that parent. The service fences cancellation before invoking the native API,
serializes ready heads, and stops admission before application child disposal.
The transport reads at most 65,537 accumulated bytes and owns its explanatory
oversize close rather than relying on the library's earlier empty-reason close.
These paths and the updated public embedding were read directly after the
source freeze. No remaining material implementation finding was established.

The [immutable repair receipt](checkpoint-evidence/ch07-review-revisions-final.json)
passes all 12 groups, including both affected modules' vet/full tests with
independent fixtures. It contains the public killed-window positive and deletion,
actual Server/WebSocket/Chrome replay across two reloads, DOM disposal and six
speech ownership scenarios, six corresponding browser mutations, size positives
and explanatory-close deletion. The reused actual event fixture is bound to
`8cc87f2`; it executes no tools or model requests. The root-close pre-fix
failure is [retained separately](checkpoint-evidence/ch07-review-speech-root-close-before.json).
The initial/revised source distinction remains explicit.

Run the separate comparative correction supplement with:

```sh
python3 scripts/edition2/ch07-review-revisions.py SOURCE_COMMIT
```

The [legacy constructor check](checkpoint-evidence/ch07-review-browser-adapter-legacy.json)
and [seven legacy mutations](checkpoint-evidence/ch07-review-browser-adapter-legacy-mutations.json)
pass on the frozen initial source after the adapter change. The same seven
[revised mutation controls](checkpoint-evidence/ch07-review-browser-adapter-revised-mutations.json)
pass with the new application owner. The only compound deletion is explicitly
recorded: both independently sufficient stale-callback fences are removed to
recreate the protected failure.
