# Chapter 18: Listen before it acts

A useful explanation arrives while there is still something to change. Imagine
hearing an Agent describe a plan to replace the wrong file. Hearing that sentence
before the tool runs gives the listener a chance to interrupt. Hearing it in the
final answer gives the listener a repair job.

Bill's need to listen while working helped drive Ensemble's steering interface.
This chapter connects one more source to that interface: the summaries a Responses
model explicitly exposes. Those summaries are model-generated explanations. They
are neither the model's complete internal reasoning nor a reliable promise about
its next action. The file still tells the truth about whether an edit happened.

The first edition found a particularly unhelpful way to deliver an explanation.
Its streamed parts and final parts received different IDs. The GUI treated the
final as another answer, displayed it again and spoke it again. The October 2
repair, recorded in commit `eb15005`, reused the live identity at finalization.
A listener had already paid attention once. The second reading added no insight.

Keep that lesson at the adapter boundary. Several summaries can belong to one
reasoning item; each needs a stable place in the display, while the original item
remains intact for replay.

**Draft contract for independent review.** No Chapter 18 implementation or live
demonstration is claimed. The [gate record](chapter-18-validation.md) separates
contract acceptance, the required predecessor and checker, and later actual use.

## TL;DR

Read the complete [coding skill](skills/ensemble-coding/SKILL.md), architecture
and this chapter before implementation. Extend the accepted Chapter 17 source in
`solutions/edition-2/main/`; the coordinator supplies that exact predecessor.

1. Add an explicit creation-only summary capability, producing the strict v8
   wrapper in §18.2. Older versions keep their exact codecs and request bytes.
   Agent owns an Actor-applied, revisioned `off|auto` selection, initially off.
   Capture it for each whole foreground turn. Helpers always capture off.
2. On Responses, auto adds exactly `reasoning.summary:"auto"`; off omits that
   member. Preserve any already-supported effort value. Other routes refuse auto
   locally. Model/account support remains conditional; no retry, model substitution
   or funding fallback repairs a refusal or an empty summary.
3. Engine maps each reasoning item and summary index through its existing owned
   operation. Actor alone publishes thinking observations and accepts finals.
   Keep cancellable fragment backpressure, control fairness and terminal usage
   settlement. A summary-done event never authorizes a tool.
4. Retain one raw capsule and its existing refs. The safe public item view contains
   an indexed summary array; deltas use the item's part ID plus summary index.
   Finals, snapshots, speech and cancellation use that same identity. No copied
   summary becomes durable dialogue, compressor input or a second raw archive.
5. Show exposed text promptly in human chat and the existing optional GUI. Reuse
   thinking display, speech selection, normalization, recorder, native ownership
   and pause causes. Finalization cannot repeat already queued words. Off changes
   future opt-in, not accepted history; reconnect never automatically speaks it.
6. Keep hints for the next actual request. Interrupt races with current response
   acceptance and tool admission; report which side won. Prove both outcomes with
   deterministic barriers, and preserve the actual outcome in live receipts.
7. Pass §18.7's distinguishing checks, including multi-summary identity, strict
   version compatibility, safe projection, exact replay, canceled usage and
   application-level speech. Then execute §18.8's bounded real-user matrix.

Language-level method spelling is the student's choice. Wire names, versioned
fields, identities, limits and transitions below are the contract. Do not ask a
reader to infer them from an old solution.

From `solutions/edition-2/main/`:

```sh
go build -o /tmp/ensemble-ch18-cli ./cmd
go vet ./...
go test ./... -count=1
```

Build the optional GUI from `solutions/edition-2/main/gui/` with `go build -o
/tmp/ensemble-ch18-gui ./cmd/ensemble-gui`; check every affected module. Retain
`make grade-dir CH=19 DIR=solutions/edition-2/main` from the repository root as a
historical diagnostic. Its old summary wire spelling is not this contract.
The independent new checker invocation must be published before student handoff.

## 18.1 Ask for something the provider exposes

The Responses reasoning guide documents summary opt-in and a `summary` array on
a reasoning item. The available modes depend on the selected model. This chapter
uses only auto, leaving the choice of supported summary detail to that model.
These are the documented API semantics checked on October 8, 2026, not evidence
that a particular account has produced text today. [Reasoning guide](https://developers.openai.com/api/docs/guides/reasoning).

There is a tempting shortcut: turn on every field containing the word reasoning
and call whatever arrives a summary. That would expose material the public view
has no permission to display. Accept summary text only from the reasoning item's
`summary` entries of type `summary_text` and the corresponding summary events.
Other reasoning content, encrypted content, signatures and unknown fields remain
opaque. A summary never supplies a call, tool permission or accepted answer text.

The plan route keeps Chapter 17's explicit selection and account-associated model
discovery. Its documentation requires streamed, nonstored Responses. The absence
of a summary prohibition is not an availability guarantee. Establish documented
model support and actual route access before the bounded demonstration; neither
an old model table nor a successful text-only response establishes summary output.
[Plan models and inference](https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference).

The first-edition coder reported an empty summary on a trivial probe and summaries
on later tasks. The underlying probe streams have not been recovered for this
rewrite. The useful lesson survives without their reported counts: an empty
array is an observation about one response. It is a poor model-capability table.

Keep separate outcomes for a local unsupported route, a provider request failure,
a valid accepted response with no nonempty summary, and actual exposed text.
Show a safe refusal through existing request-error handling. Do not guess that a
generic provider error means summaries are unavailable, and do not echo a remote
body to explain it. A successful answer with no summary remains successful.

## 18.2 Capture the choice without rewriting yesterday

Suppose a reader turns summaries off while a response is arriving. Changing the
parser halfway through that response would leave half a sentence in the browser
and a different account in the log. Apply the change to the next admitted turn.
The current turn and every automatic continuation keep their original selection.

Agent owns this setting. Actor validates and applies public requests, records the
change, then acknowledges it. Engine receives the captured option through its
operation's owner chain. Connections owns credentials and leases; it acquires no
summary settings. Browser preferences still control which sources the reader
sees or hears. Those preferences do not change a provider request already sent.

### Select a new capability explicitly

Add creation-only `--summary-profile`, requiring `--request-profile`. Without it,
all v1–v7 construction, resume, validation and rendering stay unchanged. With it,
construct version 8. An old store cannot be upgraded in place; an incompatible
resume refuses `session_incompatible`. No additional Registry handler is installed.

The exact v8 creation identity is `{requests:V7Identity,summaries:{version:1}}`.
V7Identity is Chapter 17's complete strict identity, including its base version,
base identity and immutable cache table. This wrapping grants no new base skills,
handoff, memory or recall capability. Session origin/checkpoint version and
state_version become 8; high-watermark keys remain those of the wrapped profile.

Construction emits `session_initialized` with version 8 and that identity, then
the wrapped base initializers, then Chapter 17's `request_profile_initialized`,
whose version and identity remain exactly its inner v7 values. Finally emit
construction-only `summaries_initialized` with payload `summaries` equal to
`{version:1,revision:0,mode:"off"}`. The inner request initializer is valid here
only as part of this complete v8 construction. A standalone log omits the session
fact and uses the same initializer order. Duplicate, missing, public-appended or
out-of-order initializers refuse. An anchor instead binds the complete v8 origin
and never repeats construction facts. No model request is possible during partial
construction.

Here is the complete minimal session initializer fixture. Dates and identities
are synthetic; this fixture makes no provider request.

```jsonl
{"log_version":1}
{"seq":1,"type":"session_initialized","time":"2026-01-01T00:00:00Z","session":{"version":8,"session_id":"11111111111111111111111111111111","identity":{"requests":{"base_version":1,"base":{"mode":"plain","system":"BASE","skills":null,"handlers":[]},"request_profile":{"version":1,"capabilities":null}},"summaries":{"version":1}}}}
{"seq":2,"type":"request_profile_initialized","time":"2026-01-01T00:00:01Z","requests":{"version":1,"base_version":1,"identity":{"base_version":1,"base":{"mode":"plain","system":"BASE","skills":null,"handlers":[]},"request_profile":{"version":1,"capabilities":null}},"profile":{"revision":0,"surface":"inherited","funding":"api_key","connection":null,"binding":null,"generation":null,"cache_policy":"off"}}}
{"seq":3,"type":"summaries_initialized","time":"2026-01-01T00:00:02Z","summaries":{"version":1,"revision":0,"mode":"off"}}
```

The v8 semantic state has exactly `requests` and `summaries`. Requests is the
complete wrapped v7 semantic state. Summaries is exactly the initialized object
above with its latest revision and mode. Within that wrapped state, the only new
execution substitution is the version-2 capture below, wherever v7 retains a
captured execution. The raw capsule/ref codec, compact helper accounting,
retirement rules and base codecs otherwise remain unchanged. Validate the entire
candidate before installing any owner state. Keep existing record/item limits
and the 256 MiB semantic-state bound.

### Apply and capture

A public update takes an expected uint64 revision and exactly one mode, off or
auto. A no-op with the current revision returns the current state without a
record or increment. A stale expected revision refuses `config_conflict`; an
exhausted counter refuses before mutation. Auto requires the current request
profile's Responses surface. Otherwise refuse `summary_route_unsupported` without
changing either setting. Conversely, switching that profile to inherited while
auto is selected refuses the same code. Turn summaries off first. This avoids
an undisclosed transaction between independently revisioned settings.

An applied update appends `summaries_changed` with exactly
`{base,revision,mode}` under `summaries`; revision is base plus one. Replay checks
the current base and validates the combined resulting settings. It cannot infer
mode from the current executable's defaults. Creation and anchor replay also
validate that auto has a Responses profile.

Every v8 `turn_started.turn.execution`, foreground `request_sent.execution` and
helper execution uses Chapter 17's exact capture with version changed to 2 and
one added member `summaries:{version:1,mode:"off"|"auto"}`. A foreground request
must match its admitted turn. Compression and judge captures always use off;
all their other purpose-specific settings and limits stay unchanged. Reject
version 1 captures in v8 and version 2 captures in older profiles. Do not fill
missing fields on load.

```json
{"version":2,"route":"openai-responses","requested_model":"fixture-responses","resolved_model":null,"funding":"chatgpt_plan","binding":"0123456789abcdef0123456789abcdef","generation":1,"delivery":"stream","cap":{"mode":"absent","tokens":null},"cache":{"version":1,"policy":"off","mechanism":"unchanged","capability":null},"summaries":{"version":1,"mode":"auto"}}
```

For exact wire tests, Chapter 17's uncapped, no-tools request for user text `Go`
and system `BASE` becomes this under auto:

```json
{"model":"fixture-responses","store":false,"stream":true,"reasoning":{"summary":"auto"},"input":[{"type":"message","role":"developer","content":[{"type":"input_text","text":"BASE"}]},{"type":"message","role":"user","content":[{"type":"input_text","text":"Go"}]}]}
```

Off deletes the entire reasoning member in this fixture. Insert reasoning after
the inherited optional `max_output_tokens` and before input. If an inherited
supported reasoning object already exists, preserve its members and add summary
last; off omits only summary, removing an otherwise empty object. No new effort
option is taught here. API-key Responses keeps its positive remote cap; selected
plan Responses still omits it. Local deadlines and request counts are not remote
token or charge ceilings.

Diagnostics compare the actual prepared bodies, including this member. Full-log,
checkpoint and snapshot-tail reconstruction use the captured selection, even if
the reader has since turned it off. Helpers must remain byte-identical to their
Chapter 17 requests. No helper summary is published or spoken, including text a
provider happens to return without opt-in.

## 18.3 One item, several explanations

The public display needs more detail than the replay index. Chapter 17 gives a
reasoning item one opaque ref. Replacing that ref with three independent text
parts would give the model a different conversation on its next request.
Keep the ref and derive a narrow display view.

For v8 foreground Responses only, a valid reasoning item projects to exactly
`{type:"opaque",placeholder:true,summaries:[{index,text},...]}`. Index is its
zero-based summary-array position; text is the exact exposed string, including
empty strings. Entries appear in index order. Empty summary arrays remain empty.
Only known `summary_text` entries qualify; a present malformed summary array or
known entry refuses before acceptance. Bounded unknown entry types remain opaque
and occupy their position; represent those entries as `{index,unavailable:true}`,
with no copied field from their raw object. No text may be inferred from an unknown
type merely because it has a field named text.

Ordinary unknown items keep Chapter 7's exact opaque placeholder without a
summaries member. Existing text/call positions and recursive result projection
are unchanged. Raw reasoning content, encrypted bytes and signatures never appear
in this view. Off controls request opt-in only: any valid accepted safe summaries
in a v8 foreground response still project, including earlier history. Older
profiles retain their old projections; they do not silently acquire this feature.

Use one positive part_id for the entire item, allocated in the operation's existing
namespace. Its thinking `part_delta` adds exactly `summary_index`, a nonnegative
integer. The complete live text identity is Agent, request, operation, part_id,
summary_index. Other deltas retain their old shape. Never concatenate different
indexes into one hidden thinking string or coalesce unsent fragments across indexes.
Each item still receives exactly one `part_final`, with its existing part_index
and the safe item view as `part`. Internal replay continues to derive private
opaque data from its capsule; this public substitution cannot be submitted as a
durable neutral part.

This is an explicit extension of Chapters 6 and 7's opaque final projection.
It changes neither the ref count nor the number of accepted responses. The same
item ID may appear in another operation or another Agent without sharing state.

### A fixture that can be heard twice if identity is wrong

Use this terminal response object, in the ordinary `response.completed` SSE
envelope with a valid increasing sequence_number. It has three refs, in order:
`{item:0,content:null,kind:"opaque"}`, `{item:1,content:null,kind:"opaque"}` and
`{item:2,content:0,kind:"text"}`. The encrypted strings are leak canaries.

```json
{"id":"resp-S","object":"response","status":"completed","model":"fixture-responses","error":null,"incomplete_details":null,"output":[{"id":"r-A","type":"reasoning","summary":[{"type":"summary_text","text":"Check café."},{"type":"summary_text","text":""}],"encrypted_content":"SECRET-A"},{"id":"r-B","type":"reasoning","summary":[{"type":"summary_text","text":"Read first."}],"encrypted_content":"SECRET-B"},{"id":"m-A","type":"message","status":"completed","role":"assistant","content":[{"type":"output_text","text":"Ready.","annotations":[]}]}],"usage":{"input_tokens":2,"output_tokens":3,"total_tokens":5}}
```

In a controlled operation with Agent `a1`, request `q1`, operation `op1`, allocate
part IDs 1, 2 and 3 in item order. Before completion, the following public delta
is legitimate; no response sequence has been predicted:

```json
{"kind":"part_delta","agent_id":"a1","request_id":"q1","operation_id":"op1","part_id":1,"channel":"thinking","summary_index":0,"text":"Check café."}
```

Accept the response at sequence 11. Its finals are exactly:

```jsonl
{"kind":"part_final","agent_id":"a1","request_id":"q1","operation_id":"op1","part_id":1,"response_seq":11,"part_index":0,"part":{"type":"opaque","placeholder":true,"summaries":[{"index":0,"text":"Check café."},{"index":1,"text":""}]}}
{"kind":"part_final","agent_id":"a1","request_id":"q1","operation_id":"op1","part_id":2,"response_seq":11,"part_index":1,"part":{"type":"opaque","placeholder":true,"summaries":[{"index":0,"text":"Read first."}]}}
{"kind":"part_final","agent_id":"a1","request_id":"q1","operation_id":"op1","part_id":3,"response_seq":11,"part_index":2,"part":{"type":"text","text":"Ready."}}
```

Here are the literal summary events for that operation. Each JSON line becomes
one `data: <JSON>` line followed by a blank line. A summary event can establish
the item binding before an output-item event; terminal output must confirm that
the item is reasoning. This fixture needs no invented item-added envelope.

```jsonl
{"type":"response.reasoning_summary_part.added","sequence_number":0,"item_id":"r-A","output_index":0,"summary_index":0,"part":{"type":"summary_text","text":"Check "}}
{"type":"response.reasoning_summary_text.delta","sequence_number":1,"item_id":"r-A","output_index":0,"summary_index":0,"delta":"caf"}
{"type":"response.reasoning_summary_text.delta","sequence_number":2,"item_id":"r-A","output_index":0,"summary_index":0,"delta":"é."}
{"type":"response.reasoning_summary_text.done","sequence_number":3,"item_id":"r-A","output_index":0,"summary_index":0,"text":"Check café."}
{"type":"response.reasoning_summary_part.done","sequence_number":4,"item_id":"r-A","output_index":0,"summary_index":0,"part":{"type":"summary_text","text":"Check café."}}
{"type":"response.reasoning_summary_part.added","sequence_number":5,"item_id":"r-A","output_index":0,"summary_index":1,"part":{"type":"summary_text","text":""}}
{"type":"response.reasoning_summary_part.done","sequence_number":6,"item_id":"r-A","output_index":0,"summary_index":1,"part":{"type":"summary_text","text":""}}
{"type":"response.reasoning_summary_part.added","sequence_number":7,"item_id":"r-B","output_index":1,"summary_index":0,"part":{"type":"summary_text","text":""}}
{"type":"response.reasoning_summary_text.done","sequence_number":8,"item_id":"r-B","output_index":1,"summary_index":0,"text":"Read first."}
{"type":"response.reasoning_summary_part.done","sequence_number":9,"item_id":"r-B","output_index":1,"summary_index":0,"part":{"type":"summary_text","text":"Read first."}}
```

Follow with output-item-done events numbered 10 and 11 whose item fields are the
exact r-A and r-B objects above, then response.completed numbered 12 containing
the printed terminal object. The answer is supplied only by that terminal in this
fixture. Also run the same terminal response without prior part events. Both
cases accept the same capsule and final projection; only the first proves
incremental delivery.

For independent one-fact negatives, change event 3 text to `Check tea.`, event 2
item_id to `r-B`, or add `status:"incomplete"` to event 4. Each variant retains
all other positive bytes and refuses without a capsule or executed call. A separate
no-completion variant removes only the terminal event. Preserve any provisional
text already displayed and mark its operation rejected; do not claim it was
never seen.

This fixture's public text cannot contain SECRET-A or SECRET-B. Its private next
request must still contain the original item spans with both strings. Changing a
public view copy must alter neither replay nor another observer's view.

## 18.4 Append fragments; compare snapshots

The streaming reference names separate summary-part and summary-text events. Their
item and summary indexes supply addresses; done events supply snapshots. A done
summary can precede a still-unfinished model response. The parser must keep those
completion levels separate. [Summary event schema](https://developers.openai.com/api/reference/resources/responses/streaming-events).

Validate known fields before use. Per operation, output_index binds one item_id;
per reasoning item, summary indexes form a zero-based contiguous array when
complete. At most 1,024 summary entries across the response are allowed, including
unknown or empty entries. Index 1,024 or a count above 1,024 refuses before creating
a slot. Existing overall item/byte limits apply as well. A complete snapshot can
introduce previously unseen entries in order; an incremental event cannot jump
over a missing index. Index reuse across distinct items is ordinary.

| Input | Required assembly and observation behavior |
|---|---|
| part-added | Create that slot once. A nonempty known text is its initial prefix and is published as actual thinking text; an empty value creates no delta. |
| text-delta | Require an existing known, unsealed slot; append its nonempty delta and publish promptly. Empty delta changes no text. |
| text-done or part-done | A complete snapshot. If any nonempty prefix/delta was observed, require exact equality with its accumulated text. If no text was observed, it may supply the complete text. Seal the slot. |
| repeated done snapshot | Require exact equality with the sealed value; append and publish nothing again. |
| item-done | Validate its summary array against all observed slots; it may introduce unseen slots/text. No observed slot may disappear or change. Seal the item. |
| response.completed | Apply the same complete-array agreement and Chapter 17's terminal validation. It may supply unseen items/slots, then Actor accepts or refuses the whole response. |

A complete snapshot encountered before terminal completion may publish newly
supplied text once, as one fragment. Text discovered only at terminal acceptance
appears through finalization; do not manufacture deltas after completion. A
nonempty prefix is evidence: `Check ` followed by a done snapshot `Read first.`
is a disagreement, not an invitation to replace what the user already heard.
This exact-agreement rule deliberately refuses a stream with missing deltas after
an observed prefix. It preserves Chapter 17's existing terminal contract.

Text-done may precede part-done. Identical snapshots agree; a new delta after
either done, duplicate part-added, ID rebinding or a changing completed snapshot
refuses safely. An item introduced by done alone needs no invented added event.
A part-done status of incomplete refuses the operation even if a later envelope
claims completion. Other malformed known status fields refuse. No guessed text
repairs a missing terminal response. Provider sequence numbers retain Chapter
17's strict ordering.

Only `summary_text` contributes thinking deltas. Unknown summary entries have no
delta interpretation. Unknown event types still consume framing bounds and grant
no authority. Do not reinterpret `response.reasoning_text` as the documented
summary stream.

Engine's operation owns assembly, the item map and pending fragments. Keep the
1 MiB physical SSE-event bound, 16 MiB assembled-content bound, 1 MiB pending
fragment storage and at most one ready notice. Charge retained opaque data in
its actual escaped representation. Safe summary views are derived from that
authority and cannot become an unbounded parallel transcript. Actor drains at
most 64 KiB per notice, then yields to queued controls. Every capacity/final-drain
wait observes the operation's existing deadline and cancellation.

Count each logical accumulated string once in the assembly budget; do not charge
an identical done snapshot as appended content. A temporarily owned bounded frame
and final raw payload still obey their individual and inherited peak-memory
constraints. Build checked offsets/views where possible and copy only at public
ownership boundaries. Rebuilding all preceding summary text for each delta would
make listening progressively more expensive as the explanation grows.

The terminal response remains the sole acceptance and usage boundary. Summary-only
output without an ordinary text position or call still fails the Chapter 17
foreground rule. Preserve its distinction between valid canceled terminal usage
and rejected effects; an interrupted summary creates no capsule, answer or call.
Once-only usage settlement cannot be inferred from how much speech was heard.

## 18.5 Recover the display without replaying the voice

Extend the existing watch, rather than adding a summary subscription beside it.
Actor publishes the same revision-ordered observations to CLI, public consumers
and GUI. Copied current-operation state contains the item part IDs and indexed
summary prefixes, so a newly joined client can recover what is known so far.
An active operation before its first summary remains an active operation with no
summary part; it needs no fabricated beginning sentence.

For a v8 summary item, `snapshot_partial` has the existing identity fields,
`channels:{}` and an added `summaries` array using §18.3's exact entries. Other
partials keep their earlier form. Here is a mid-operation snapshot record:

```json
{"type":"snapshot_partial","generation":"g1","agent_id":"a1","request_id":"q1","operation_id":"op1","part_id":1,"channels":{},"summaries":[{"index":0,"text":"Check café."},{"index":1,"text":""}]}
```

It is staged between the ordinary snapshot_begin/end at one atomic watermark.
The subsequent live tail uses the same part and index identities. Summary
publication consumes the ordinary watch revision; a snapshot plus tail cannot
lose or repeat a prefix at their join. The existing queue bounds and explicit
overflow/resubscribe behavior apply. An overflowing viewer cannot stall Actor,
reliable completion or another viewer.

For accepted history, the browser projection derives ordered parts from the
capsule refs, using the same item view printed in §18.3. Its `snapshot_event`
contains these projected response parts and no `responses.raw`, private refs or
raw usage payload. Retain the ordinary safe provenance/counter fields. There is
no durable part_id to invent: accepted cards key by Agent, response sequence,
part_index and summary index. Live finals explicitly connect operation-local
part_id to that accepted position. A snapshot replaces a generation atomically,
so a reconnect can create these cards without reviving old speech ownership.

A v8 watch's safe state adds `summaries` with exactly revision, mode, surface and
support from §18.6's current status, captured at the same watermark. Older watches
retain their old state shape. The ordinary durable observations for
summaries_changed and request_profile_changed let an attached client update that
status in order; reconnect obtains a fresh coherent copy. They carry no credential
or inferred model-availability field. A setting changed by another client must be
visible without waiting for the next model response.

Render summary entries in item/index order, preserving empty and unavailable
positions. Label the known text “Reasoning summary”; an unavailable entry has
no speakable text. Manual card speech uses only known safe summary text, joining
separate entries with LF in index order. It must never read the opaque placeholder
or substitute decrypted material.

Automatic speech uses the existing thinking source and the reader's existing
selection. Extend its private source identity by summary_index and reset the
incremental normalizer at each index. Queue complete utterances in actual arrival
order; display ordering does not reorder speech already admitted. A final supplies
only the unconsumed suffix for that exact source. Empty text queues nothing.
The speech journal retains its strict v1 record shape: part_id names the whole
item, while its distinct utterance IDs identify admitted sentences. It does not
claim to encode summary_index. Tests correlate the application source tuple with
the actual recorder text and utterance lifecycle, rather than treating two records
with the same item part_id as duplicates.

For the §18.3 fixture, the normalizer should offer `Check café.`, `Read first.`
and `Ready.` exactly once each when both thinking and answers are selected. Final
arrival adds none of the already offered text. A terminal-only variant offers
the same sentences from its finals. Chapter 13's code filtering, bounds, overflow
and failure announcements still apply. Raw tool results remain silent.

On interrupt, cancel this operation's queued/current provisional speech through
its Page owner, fence later callbacks and preserve other Pages' work. A sentence
already spoken cannot be retracted. When response acceptance wins the race,
its accepted cards remain; interruption cannot rewrite an executed tool into a
prevented one. Close and mount replacement keep their inherited lifetime rules.
An old callback cannot clear a replacement Page's speaking cause.

Off does not erase accepted cards. Whole-bundle retirement does: derive the next
public history from the retired capsule state and remove its summary views with
the rest of that bundle. Do not retain an extra summary history in a server cache,
session snapshot or diagnostic prepared-body export. Explicit user exports retain
their Chapter 17 ownership and retention meaning; hashes cannot restore removed
raw text. The speech journal remains its independent bounded transient diagnostic,
and cannot be used as a model-history source. Existing journal retention does not
grant a second durable display archive.

## 18.6 Keep the controls honest

Human chat adds `/summaries`, `/summaries off` and `/summaries auto`. The first
shows current mode, revision, selected surface and its conditional support. The
others apply the same public update, reporting its acknowledged state. A client
reads the current revision and submits it; a conflict is shown without an automatic
overwrite. The optional GUI adds the corresponding off/auto control beside the
existing execution settings, plus a separate explanation that thinking display
and speech remain reader preferences.

Use the existing machine/client command envelopes. The protocol command is
`{kind:"summaries"}` for inspection or
`{kind:"summaries",mode:"auto",expected_revision:"0"}` for an update. The GUI
uses its ordinary correlated envelope with type `summaries` and the same fields.
Revision is a canonical decimal uint64 string on these public wires. Success
returns exactly the following status, inside the GUI's ordinary correlated reply
when applicable:

```json
{"kind":"summaries","revision":"1","mode":"auto","surface":"responses","support":"model_dependent"}
```

Surface is inherited/responses. Support is unsupported_route for inherited,
model_dependent for Responses. The latter reports an adapter path, never proven
account access. Pre-v8 Agents return `summary_capability_required` for this new
command. Bad fields, extra fields and bad revision tokens use the existing safe
invalid-command response; failed updates preserve all state and make no HTTP call.
Preserve every old default protocol reply.

The launcher applies `--summary-profile` at creation/resume, then the ordinary
Chapter 17 route command selects Responses before `/summaries auto`. Resume loads
the durable mode; the creation flag supplies capability identity rather than
silently resetting that mutable setting. The public constructor and GUI launcher
provide the equivalent selection. A change of model on the same surface keeps
auto but makes no new promise of model support; a remote refusal follows the
ordinary safe failure path without silently disabling the selected setting.

For v8 foreground Responses, `model_end` additionally carries
`summaries:{mode,outcome}`. Mode is the captured value; outcome is present when
an accepted response contains any nonempty known summary, empty when auto was
requested but none exists, not_requested when off and none exists, or unavailable
when the operation was rejected/canceled. Existing rejection code/message identify
the actual failure. Nonempty provisional text followed by failure is unavailable,
not an accepted summary. Accepted unsolicited text under off is present. These
outcomes are derived observations, never independent durable facts or usage.

Show “No reasoning summary returned” for empty; show the existing safe error for
unavailable. Preserve a visibly provisional label until acceptance. The terminal
must print exposed fragments before completion while still accepting controls.
No raw SSE or JSON command is needed for a human demonstration.

Chapter 5's `/hint` persists a hint for the next actual request. If the current
response proposes a call, a hint does not reach backward into its already-sent
body. Interrupt is the control that can prevent unaccepted work from proceeding.
Its acknowledgement identifies the request; the log identifies whether response
acceptance or tool admission had already happened. A slow human may lose that
race. Report the loss and use the next turn instead of calling it successful
prevention.

These are separate deterministic controls: hold response completion until an
interrupt is admitted and prove no proposed call executes; then allow acceptance
and admission first and prove the earlier effect remains. For hints, hold a tool
result long enough to admit the hint and prove it appears in the next request
exactly once. None of these tests requires a new remote steering API.

## 18.7 Distinguish the implementation from its convenient substitute

Run literal fixture variants before paid demonstrations. Build positive controls
with complete valid captures, timestamps, initializers and real public consumers;
an earlier malformed fixture cannot prove a later lifecycle check. Keep the
existing base-profile and Chapter 17 request/capsule tests intact.

| Check group | Positive control and distinguishing failure |
|---|---|
| Version and capture | Complete v8 over each permitted base; old v7 unchanged. Reject v8 fields in old records, mixed execution versions, partial initialization and stale updates. Toggle during a turn and reconstruct all continuations from its old capture. |
| Exact request | Compare printed auto/off bytes and both funding routes. Preserve effort, helpers and caps. Reject auto on inherited and inherited switch while auto, without HTTP or state mutation. |
| Timeliness | A server withholds terminal completion until the actual public subscriber sees a nonempty summary; buffer-until-final must fail. Repeat through actual human CLI and GUI handlers. |
| Identity | Two summaries in one item, another item reusing index 0, reused provider IDs in another operation and another Agent; one final per ref and once-only speech. |
| Assembly | Nonempty added plus split deltas; done-only and terminal-only positives. Change one done character, remove an observed slot, duplicate added, rebind ID, add a post-done delta or mark incomplete: refuse with no accepted effects. |
| Bounds | Split café across every byte boundary; split JSON escapes and CRLF. Exact 1 MiB SSE framing and next byte; exact 1,024 summaries and one extra; inherited assembled/queue bounds, cancellation while full and responsive controls. |
| Privacy and persistence | SECRET-A/B absent from every public/browser/speech output and present in exact private replay; copied view mutation isolated. Full-log/checkpoint/tail equality; authorized whole-bundle retirement removes safe views and raw payload together. |
| Lifecycle | Interrupt before acceptance versus after admission; canceled terminal usage once; close, overflow, resubscribe and mount replacement. No late text, duplicated settlement or effect revival. |
| Speech and availability | Actual application recorder/native admission; final-only versus streamed text once; no historical autoplay. Empty-success, local route refusal and provider error remain distinct. |

Summary bounds are per response; one entry cannot evade them by cycling IDs.
Unknown summary entries consume slots and opaque bytes. For the count boundary,
use 1,024 empty known entries in one reasoning item plus one ordinary empty text
position. This accepts without inventing speech. Add a 1,025th entry for the
negative. Use distinct increasing event sequence numbers so that the intended
count check is reached.

The inherited maximum frame can be tested with one comment-only event followed
by the small terminal fixture. Include its blank delimiter in the byte count and
exclude only the permitted leading BOM, as Chapter 6 requires. An encrypted
canary test must include successful public delivery; an empty observer proves
neither safety nor useful summaries.

The independent reviewer compares the initial implementation with the historical
standard after initial actual use. The student receives the rationale and revises
from this chapter, without reading the old answer. Preserve the initial teaching
review even if these checks pass. A source-identity failure in an evidence tool
must be corrected before any derived replay can be trusted.

## 18.8 Taking it for a spin

**Demonstration pending.** The following is the finite exercise plan, not a
transcript. Fill the reader-facing account from actual terminal, public-consumer
and browser receipts after implementation. No model has run for this chapter.

Use a scratch workspace with `choice.txt` containing exactly `BLUE\n` and a
different sentinel file whose hash is recorded before and after. Supply only the
ordinary read tool needed by these tasks. Keep maintenance disabled so compression
or judge calls cannot consume an undisclosed allowance. They have separate local
request-equality coverage here; Chapter 17 supplies their live feature evidence.

Before launch, select a documented summary-capable model available through each
intended Responses funding route. Retain the dated discovery association and
requested/returned identities. If access or support is absent, stop that route
and record the missing live gate. Do not obtain another client's credentials or
change funding automatically. Reuse Chapter 17's owned Connections and explicit
registration; new authorization work is not this chapter's exercise.

The provisional ceiling is 30 actual model starts: 12 per Responses funding route,
then two each for Messages, Chat Completions and generateContent. A start consumes
its slot at durable model-request admission, including an attempt subsequently
refused, interrupted or failed. Continuations and any unexpectedly admitted helper
count too. Unused slots expire; no process restart or nicer prompt resets them.

Use this exact allocation on each Responses route:

| Slots | User path and action | Required observation |
|---|---|---|
| 1–2 | Human chat, summaries auto: ask for a comparison of reading the small file versus guessing its contents, then have it read choice.txt and answer from the result. Admit `/hint End the answer with CHECKED.` after observing activity and before the continuation. | Actual exposed summary timing, local tool result, next-request hint bytes and answer; if the hint arrived late, preserve that result. |
| 3 | Same human chat: request a substantial explanation of a safe inspection plan without tools; send `/interrupt` after an observed summary fragment. | Actual interrupt acknowledgement and acceptance race outcome, with provisional speech/display handling. |
| 4 | Same chat: request the exact answer `READY-EIGHTEEN` with no tool. | Usable next turn and correct completion after the interruption attempt. |
| 5–6 | Public executable: one concurrent prompt on each of two Agents with distinct summary selections and identifying text. A uses auto; B uses off. | Independent ownership, captures, observers, accounting and no cross-Agent text or cancellation. |
| 7–8 | Actual GUI, recorder selected visibly: request the same one-tool file inspection with summaries auto and answer it. | First exposed summary reaches the real Page before model end; final speech has no duplicate text; tool continuation and safe cards. |
| 9 | Same mounted GUI: request a substantial tool-free explanation, then interrupt after an observed summary fragment. | Recorder cancellation for that source, other Page unaffected, and preserved race result. |
| 10 | Reconnect that Page, use manual summary-card speech, then set summaries off and request `GUI-READY`. | Snapshot text preserved without autoplay, explicit manual safe text, off request bytes and usable next turn. |
| 11 | Fresh native-output application: summaries auto, one tool-free comparison prompt. | Actual native submission, exact journal text and a named observer/audio capture of output; callback alone is not hearing. |
| 12 | Same native application: turn summaries off and request a short answer, then quit. | Independent speech selection, owned pause release and clean close, with no retained speech or workers. |

Slots 1–2 and 7–8 reserve two starts total for each task, including its continuation;
if the model asks for more, stop at the ceiling and report that task incomplete.
Other rows reserve one start each except the two independent public Agents.
Give tool-free rows no visible tools and a one-response limit. Use the inherited
turn request-count policy for each bounded task. The total allocation is 12 per
route even if a tool-free response finishes before the human can interrupt it.

Configure a 120-second operation deadline, covering lease wait, HTTP reads,
fragment pressure and final drain, and an external 300-second ceiling per
two-start task. One-start tasks have a 180-second external ceiling. Keep the
existing local 1 MiB frame, 16 MiB assembly and speech bounds. API-key requests
use an explicitly recorded positive cap of 2,048 output tokens; selected plan
requests have no remote output-token cap. External termination must close owned
work and retain the failure; these times do not promise a remote spending limit.
Local native-speech start/lease timers remain the Chapter 8 values rather than
being extended to obtain a favorable demonstration.

For each inherited adapter's two starts, use real human chat with summaries off:
one ordinary short answer and one usable follow-up. Attempt `/summaries auto`
between them and observe its local unsupported-route refusal without consuming a
model start. Preserve their earlier thinking behavior; this chapter does not
rename that output a Responses summary.

No model-based listener is allocated. Use the application recorder for exact text
and a person or captured audio for the stated native observation. If a listener
model becomes necessary, revise the finite allocation before launch rather than
hiding its requests outside the count. Required nonempty summaries must be
observed during generation on both intended Responses routes; empty valid
responses leave this feature unverified. Local barriers prove interruption and
hint semantics when human timing loses, but cannot replace a claimed live success.

Retain sanitized exact prepared requests, raw streams, durable logs, terminal
input/output, browser screenshots with text descriptions, recorder/native records
and workspace hashes. Bind source, binaries and launch identities before replay
or derived writes. Compare actual ordered requests with offline reconstructions;
label each accordingly. Keep failure attempts and their consumed slots visible.

The eventual account should follow the listener's experience: which explanation
arrived, what control was used, whether it took effect, and whether the next turn
worked. Link the detailed ledger rather than printing it in place of that story.
The historical duplicate-answer repair earns its place when the same accepted
explanation reaches the listener once.
