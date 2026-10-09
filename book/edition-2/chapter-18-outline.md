# Chapter 18 outline: Listen before it acts

Through-line stake: a listener needs enough provider-exposed explanation to notice
a wrong direction while the Agent can still be steered, without hearing the same
summary twice or mistaking a provisional explanation for an accepted action.

October 8, 2026. **Research and proposed scope only.** The coordinator selected
this focused direction after reviewing the overlap with Chapter 17. This is not
a complete student contract, implementation release or live capability claim.
The [evidence record](chapter-18-evidence.md) distinguishes historical accounts,
current documentation and still-required measurements. No current student needs
to read this author research.

## Why this chapter still earns its place

The old Chapter 19 combined credentials, Responses, reasoning summaries and billing.
New Chapter 17 already teaches own-application authorization, Connections ownership,
refresh/cancellation, explicit funding selection, Responses replay and usage. Moving
that material back here would break its required subscription-cache investigation.
Repeating it would leave the reader building the same feature twice.

Keep the remaining integration lesson: opt in to Responses summaries, carry their
identity through the existing stream, expose only their public summary text, and
let the existing listening/steering interface use it while work continues. Chapters
6, 7 and 13 provide the machinery; this chapter proves the new adapter actually
reaches it. The following real-use chapter remains a broader daily-work evaluation,
not a second summary implementation or a delayed repair of known defects.

The coordinator chose this direction over merging the residue into that following
chapter. The current map remains new 18 from old 19, followed by new 19/old 20,
new 20/old 21 and the required ending comparison. The publication-time first-person
Codex epilogue remains after the comparison. No first-edition file or historical
evidence changes with this redistribution.

## Story preservation and voice

Retain the accessibility purpose of old Chapter 19's reasoning-summary section:
a completed paragraph arrives too late to help someone redirect work already done.
Use the established biography from voice.md without importing a new listening-speed
number. This is a concrete reader problem, not a promise to reveal the model's
complete internal reasoning. Summaries are provider-exposed output; encrypted replay
material remains opaque.

The first-edition coder reported a trivial probe with no summary, followed by
nontrivial tasks that did produce summaries. Preserve that as a dated reported
near-miss if used; the raw probe artifacts have not been recovered in this research,
so do not publish its delta counts as fresh measurements. A zero-summary response
cannot by itself establish that an entire model or funding route lacks support.

A stronger source-backed mechanism comes from `eb15005`: streamed Responses parts
were finalized with new IDs, so observers displayed and spoke answers twice. The
commit records both the failure and its correction. This explains why the new
exercise must join deltas and finals by full identity, even though earlier chapters
already teach that rule. Tell the historical incident briefly; do not make the
student recreate it.

Preserve the old migration opener in its historical edition. Its blanket vendor
claims, price multiplier and synthetic-looking sign-in transcript do not become
new Chapter 18 evidence. No factual dispute requires erasing the original author's
voice; retain useful stakes and remeasure current claims where they matter.

## Material already taught and the remaining delta

| Subject | Existing teaching to reuse | Proposed Chapter 18 work |
|---|---|---|
| Funding and credentials | Ch17 Connections/leases, route selection, own-app login, no fallback | No new auth store, login flow or automatic metered fallback |
| Raw Responses history | Ch17 one authoritative capsule, exact replay, protected retirement | Derive public summary views from that capsule without duplicating replay text |
| Stream ownership | Ch6 Engine operation, Actor acceptance, provisional IDs, bounded delivery | Map summary item/index lifecycle onto existing thinking observations |
| Human CLI | Ch5 controls; Ch6 separate thinking label and terminal lifecycle | Show summaries immediately and keep hint/interrupt available |
| Browser and speech | Ch7/8 Page queues, shared speech owner, pause causes and settings | Exercise the existing path with actual Responses summaries and safe final views |
| Spoken-output evidence | Ch13 journal/recorder/native paths and listener provenance | Prove once-only speech, interruption and no replay autoplay for this source |
| Usage and limits | Ch17 purpose accounting, uncapped selected plan versus capped API-key | Preserve the funding/cap distinction; no summary-specific usage counter |

## Proposed teaching order

1. Open with the listener noticing a wrong approach before a file changes. Explain
   the duplicate-final historical failure and what its ID mismatch did to speech.
2. Distinguish exposed summary text, visible answer, provisional tool proposal and
   encrypted replay data. The summary reports a model-generated explanation, not
   an authenticated plan or permission to execute.
3. Select summaries explicitly for foreground Responses requests. Initial proposal:
   off/auto only, default off; omit the field when off and emit
   `reasoning:{summary:"auto"}` when on. Leave reasoning effort unchanged. The
   current documentation supports the API shape, but support/availability remains
   model and account dependent. Preserve unsupported-route refusal and a successful
   response containing no summaries as different outcomes.
4. Capture that selection before rendering, across automatic continuations and
   exact replay. Helpers keep summaries off and their present result-size limits.
   Publish a versioned compatibility extension before code; do not silently add
   fields to old strict v7 records or change old request bytes.
5. Follow item_id/output_index/summary_index through added, delta and done events,
   terminal agreement and accepted finalization. Reuse the thinking channel and
   operation ownership; avoid a second transport or observer bus. Publish the exact
   mapping for multiple summaries inside one reasoning item before implementation.
6. Project accepted summary text for public display without exposing encrypted
   content, raw capsules, credentials or bound signatures. Keep original ordered
   items for replay and whole-bundle retirement. No copied summary becomes a new
   conversation message or an unsigned compressor source.
7. Take that path through human chat, two-Agent public embedding and the optional
   GUI's existing speech/steering controls. End with an observed correction or
   interruption followed by a successful next turn, using actual receipts.

## Ownership and compatibility constraints for the eventual contract

Agent owns the selected execution option; Actor applies changes and captures the
turn. Engine owns the request/parser operation, its IDs, bounds and producing usage.
Shared declarations remain common; behavior stays in llm. Public display is an
owned projection of accepted facts and provisional state, with existing watch
watermarks and overflow recovery. BrowserApplication/SpeechService and Page retain
their present ownership, pause registration and close/reconnect responsibilities.
Connections has no summary state. No sibling imports, injected callback bags or
new globals are justified by this adapter feature.

The existing GUI suppresses raw opaque final content. The new contract must specify
the narrow public summary view and its identity correspondence explicitly; a
student must not solve visibility by passing opaque payloads to the browser.
Determine whether one safe final view with indexed summaries suffices, or whether
additional projection entries are required. Preserve Ch17's one raw authority.
This is the remaining representation detail to resolve during full drafting.

Cancellation invalidates provisional summaries under the existing operation,
cancels their queued speech and leaves unrelated Page work alone. A done summary
is not a completed response. No partial call executes; no accepted usage is invented.
A late frame cannot revive a canceled request. Final snapshots may supply absent
text but must not replay already delivered words. Reconnect may recover display
state, but must never automatically read historical material aloud.

## Finite acceptance plan to turn into literal fixtures

| Property | Required distinguishing control |
|---|---|
| Selection | Off preserves exact Ch17 request bytes; on has exactly the captured summary option; changing settings cannot rewrite an active turn |
| Timeliness | Server withholds terminal/answer until the public consumer receives the first summary fragment; buffering to completion fails |
| Identity | Two reasoning items with repeated summary_index, two operations reusing item IDs, and two Agents; no merge or duplicate final |
| Terminal agreement | Nonempty added part, split deltas, done-only text and terminal snapshot; mismatch/incomplete response cannot become accepted output |
| Safe projection | Public/browser output contains only exposed summary text; an encrypted canary stays absent while exact private replay remains intact |
| Lifecycle | Interrupt during summary, blocked/overflowing observer, close and reconnect; later turn progresses and no summary gains tool authority |
| Speech | Real application handler reaches recorder/native queue, finalization queues no duplicate words, canceled provisional work stops, replay stays silent |
| Persistence | Request reconstruction uses captured selection; summary display does not duplicate capsule payload or survive whole-bundle retirement as a hidden archive |
| Capabilities | Unsupported selected route fails visibly with no retry/funding change; empty-summary success stays an honest success with no fabricated explanation |

Inherited `make grade-dir CH=19 DIR=solutions/edition-2/main` can be diagnostic only:
its OAuth/funding/cache expectations are mostly Chapter 17 concerns, and its summary
check assumes an old GUI wire format. Publish an independent new acceptance command
before handoff. Preserve old coverage; do not force the new student to imitate old
message spelling or inferred timing to satisfy it.

## Proposed actual-use budget and remaining gate

Before paid work, allocate at most 12 model starts on each Responses funding route
across human CLI, public two-Agent and actual GUI/listening exercises. Add at most
two starts on each inherited API-key adapter for parity: 30 starts total. Count
failures, continuations, canceled requests and any unexpectedly admitted helper
inside that total. No timer, process restart or client change resets it. This is
an outline budget for review, not authorization to launch.

Required observations on each intended Responses route: a nonempty exposed summary
arrives during generation; a real local tool turn continues; the human sends a hint
or interrupt after observing output; the next turn works; the public consumer
keeps two Agents independent; browser speech follows the same source once. Use
recorder receipts for exact text and actual native-speech observation for usability.
Keep all three earlier adapters working; do not claim they expose this new summary
schema. A route producing no summaries within the finite matrix leaves that live
feature unverified. Report it rather than retrying until an attractive transcript
appears, silently changing funding, or treating a fake frame as provider evidence.

Root next reviews this direction and finite scope. Full draft must then resolve
capture/version grammar, summary-to-final public identity and exact unavailable
replies with primary sources and literal fixtures. No Chapter 18 code, auth or
provider request is released by this outline.
