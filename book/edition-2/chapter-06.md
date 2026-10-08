# Chapter 6: Show the answer while it arrives

A coding agent can be busy for a long time while its terminal says nothing.
The reader cannot tell whether it is choosing a file, waiting for a model, or
stuck. Chapter 5 made interruption possible. This chapter supplies something
worth interrupting: the answer as it arrives, including the name and arguments
of a proposed tool call.

That last word matters. A path appearing on screen does not authorize a write.
The rest of the response could contain malformed arguments, a provider error,
or nothing at all because the connection broke. Let the reader see the work
early. Keep the decision to act at the boundary already established by the
actor: one complete, validated, durably accepted response.

**Draft status:** the initial implementation and real user runs remain preserved.
Comparative review led to the repairs taught below; revised validation and live
demonstrations are in progress. See the [Chapter 6 gate record](chapter-06-validation.md).

## TL;DR

Extend the accepted Chapter 5 source in `solutions/edition-2/main/`. Read the
entire `book/edition-2/skills/ensemble-coding/SKILL.md`, the architecture ledger,
and this chapter before editing. Do not inspect first-edition answers or
author research notes. `ch06/` will be a frozen export after validation.

1. Add streaming delivery to Messages, Chat Completions, and generateContent.
   An Agent configuration switch disables it; the default is enabled for these
   adapters. Expose `EN_DISABLE_STREAMING=1` in the CLI, with absent or `0`
   meaning enabled and other values a configuration error. Snapshot the choice
   per model operation and record `request.delivery` as `stream` or `plain`.
   Missing delivery in historical events means plain. No model-name allowlist,
   automatic fallback request, retry, or new API surface is required.
2. Keep one actor owning durable history and turn decisions. Engine owns each
   model operation and its parser state. Common declares shared facts and
   parent interfaces; parsing and assembly remain in llm, with logger access
   through the actual owner chain. Workers send owned facts to the actor.
3. Allocate an operation ID before HTTP and a positive local part ID before
   publishing that part. Every transient observation identifies Agent, request,
   operation, and part where applicable. Never predict a durable response
   sequence. Accepted final observations map these identities to the actual
   response sequence and zero-based part position.
4. Publish append-only fragments for visible text, explicitly exposed thinking
   text, tool name, and JSON argument text. Publish complete typed finals only
   after acceptance. Opaque bytes and signatures remain replay material;
   their presence does not make them printable reasoning. Complete calls can
   arrive without argument fragments. No tool executes from a fragment.
5. Read SSE incrementally with the framing and API completion rules below.
   Bound one frame to 1 MiB and assembled response data to 16 MiB. Preserve
   UTF-8 across reads. A finished response needs valid parts and required usage.
   Normalize plain and streaming results through the same semantic rules.
6. An incomplete, malformed, interrupted, or failed operation contributes no
   accepted response, proposed tool effect, or usage increment. Earlier accepted
   rounds and running jobs survive under Chapter 5. Mark provisional display
   incomplete, settle the reliable request handle, and ignore late facts by ID.
7. Human `chat` flushes fragments before completion, labels tool proposals and
   exposed thinking separately, keeps controls available, and avoids repeating
   the answer at completion. Plain mode shows the completed answer once. Keep
   default machine `protocol` unchanged; `protocol --observe` adds the exact
   observation records below. The optional GUI module remains a stub.
8. Prove delivery with a server barrier, full typed-content equivalence with
   controlled fixtures, identity isolation, framing failures, cancellation,
   background-job sequence interleaving, and observer overflow. Then actually
   use human chat in a PTY with all three real APIs, including tools, interrupt,
   and plain mode. Preserve the initial answer for independent comparison.

Commit the implementation, then run the independent deterministic gate from
the repository root, replacing `SOURCE_COMMIT` with that immutable commit:

```sh
python3 scripts/edition2/accept_ch06_gate.py SOURCE_COMMIT
```

It checks an exported copy of the committed source. Actual user runs, receipt
review and independent code/prose review remain separate gates. The historical
diagnostic is `make grade-dir CH=7 DIR=solutions/edition-2/main`; seven is the
first-edition number. Its old observation and structural contract produces
0/100 on this implementation and does not grade the new exercise correctly.
Method names and internal channel layout are student choices. The wire forms,
ownership, ordering, failure boundaries, and observations below are requirements.

## 6.1 Three different meanings of finished

A parser can finish a JSON object while the model is still producing the next
one. A model response can finish while the Agent still has tools to run. The
human request finishes only when its reliable handle settles. Combining those
three boundaries makes a terminal appear responsive at the price of doing
work the model never successfully requested.

Keep the Chapter 5 loop. Engine creates an owned model operation for each HTTP
attempt, including automatic continuations. Its interface back-pointer reaches
Engine, then Agent, then Ensemble. The parser uses that context for diagnostics
and to deliver facts through the existing owner interfaces. An extra bundle of
logging and publishing callbacks would conceal the same ownership that the
earlier chapters worked to make visible.

The operation owns incremental assembly. The actor owns whether the operation
is still current and whether its result can enter history. A parser may report
text while it reads; it cannot append `response_ended`, update conversation
state, settle the human request, or dispatch a tool. Shared payload declarations
belong in common, and free functions in llm perform the parsing and assembly.

The actor accepts a completed owned response by the existing append path.
Persistence succeeds before usage is committed and accepted finals are
published. Engine remains the accounting authority. A failed append faults the
Agent under the existing persistence rule, without advertising a successful
answer merely because some text already reached the terminal.

No delta becomes a durable conversation event. Saving every fragment would
mix display transport with replay and force the renderer to reconstruct an
answer from network accidents. Keep one accepted response, with its complete
ordered parts, requested and returned identities, raw usage, and normalized
usage. The existing `response_started` vocabulary does not become a second
mandatory durable event; the transient boundary below has its own name.

## 6.2 Give the first fragment an identity

Suppose a command finishes while the model is writing its answer. Chapter 4
requires the job's terminal fact to enter history promptly. The sequence number
that looked available when HTTP began may now belong to `job_ended`. Reserving
that number for a future response would either break log order or hold the job
fact hostage until the model finished.

Allocate a nonempty operation ID unique within the Agent's lifetime before
starting HTTP. A request may have several operations; an operation belongs to
exactly one request. IDs need not be counters or UUIDs. Public identity always
includes Agent ID, so two Agents may choose the same local spelling safely.
Assign each logical part a positive integer ID within its operation. IDs are
shared across all part kinds, never reused during that operation, and remain
stable from first observation through acceptance or abort.

Vendor block indices and tool IDs are inputs to this mapping. They are not the
public identity contract. A Messages text block and a tool block cannot both
be public part 1 in the same operation. A later response can start its local
part numbering at 1 because the operation ID changed.

At acceptance, the actor supplies the actual `response_ended` sequence. Missing
Gemini call IDs become `call-<response seq>-<part position>` at this point, as
Chapter 4 specifies. Work on an owned copy; parser-returned facts do not mutate
after delivery. The final typed call contains that ID, even though its earlier
proposal could not know it. A client correlates the two through operation and
part identity, not by guessing a call ID from text.

Use the following public observation semantics. These fields also define the
opt-in protocol representation; language-level spelling can vary.

| Kind | Fields in addition to `kind`, `agent_id`, `request_id`, `operation_id` |
|---|---|
| `model_begin` | `delivery`: `stream` or `plain` |
| `part_delta` | positive `part_id`; `channel`: `text`, `thinking`, `tool_name`, or `tool_args`; nonempty string `text` |
| `part_final` | `part_id`, positive `response_seq`, zero-based `part_index`, owned neutral `part` |
| `model_end` | `accepted`: boolean; accepted case has `response_seq`; rejected case has safe `code` and `message` |

Publish exactly one begin before any observation of an operation, and exactly
one end after its accepted finals or its abort. Empty fragments are omitted.
Begin follows successful rendering and the persisted `request_sent`; an
earlier configuration/rendering refusal does not invent an HTTP operation.
Text and thinking fragments append independently; tool name and argument
fragments append to separate strings for that part. A part cannot change
kind. An opaque final may have no deltas, and a complete tool call may produce
one name fragment and one complete argument fragment. Do not manufacture many
fragments after completion to imitate streaming.

For an accepted response, publish one final for every ordered part, including
empty text and opaque material. Finals arrive in final part order, carry the
same local identities as any preceding fragments, and precede accepted end.
Plain mode emits begin, finals, and end, with no part deltas. Both modes retain
Chapter 5's durable event and state observations and reliable request handles.

Only the actor publishes these public observations. Engine's operation keeps
at most 1 MiB of pending decoded fragment text and at most one outstanding
fragment-ready notice in the actor's mailbox. Split a larger individual
fragment at complete UTF-8 boundaries as needed, retaining its identity and
channel. When that store fills, the producer waits for capacity or cancellation.
It holds no actor, mailbox or owner lock while waiting. Coalescing adjacent
unsent fragments of the same part/channel is allowed, without delaying the
first available fragment until HTTP ends.

The actor drains at most 64 KiB of fragment text per notice, then places any
remaining-work notice at the mailbox's tail. This bounded service lets already
queued controls and job facts run between drains. Chapter 5's admission mailbox
remains growable and nonblocking for producers; it carries readiness notices
rather than an unbounded copy of every fragment. No fragment is silently lost
while its operation remains current. The producer enqueues its completed
response only after the actor has consumed its pending fragments, preserving
their order before acceptance and finals. That wait is cancellable too.

Interrupt and close cancel a producer waiting for space, discard its now-stale
pending fragments, and publish the rejected operation end through the actor.
Late worker facts cannot answer another request. Accepted finals use the
ordinary actor acceptance path and cannot be dropped by the fragment bound;
public subscriber overflow remains the separately reported display failure
from Chapter 5. Equivalent storage mechanisms must preserve these bounds,
cancellation, ordering and control responsiveness.

## 6.3 Read the envelope before interpreting its contents

An SSE event is a small text envelope. It has no universal model-finished
instruction. In particular, `[DONE]` is a Chat Completions payload convention,
not part of SSE itself. The shared reader should return an event name and joined
data to an API decoder; it should not contain a switch on three model families.

The SSE standard specifies UTF-8, line endings that may be LF, CRLF, or CR,
colon-prefixed comments, and repeated `data` lines joined with LF. A blank line
dispatches the event. EOF alone does not dispatch an unfinished event.
These framing rules were checked on October 7, 2026 against the
[WHATWG specification](https://html.spec.whatwg.org/multipage/server-sent-events.html).

Handle arbitrary read boundaries, including a split CRLF pair, a split UTF-8
code point, and a JSON escape split between reads. Ignore one leading UTF-8
BOM, comments, `id`, and `retry`; this HTTP client does not reconnect. Split a
field at its first colon and remove at most one following space. Missing event
name means `message`. Preserve all data-line content apart from the framing
rules. Reject invalid UTF-8 data instead of replacing it with invented glyphs.

The event limit is exactly 1,048,576 physical UTF-8 wire bytes. Count every
field name, colon, space, comment, ignored field and line terminator, including
the terminating blank line. LF and CR each count as one byte; CRLF counts as
two even when split between reads. Exclude only the optional three-byte BOM at
the beginning of the response. Reset the counter after every blank-line event
boundary, including a comment-only or otherwise ignored event. Exactly the
limit is allowed; the next byte fails before dispatch. Do not reset at CR
until its possible following LF has been counted with the same event.

Count accumulated part strings, opaque payloads, and argument bytes against
16 MiB, without charging repeated usage snapshots as new response content.
Charge opaque data in its retained representation: a control character inside
retained JSON may occupy six escaped bytes even though its decoded string has
only one. A visible-text counter cannot stand in for that opaque payload size.
Fail safely on overflow. A reader may hold one bounded frame, the assembled
response and the bounded pending fragments above; it cannot buffer the entire
HTTP body before delivering the first delta. Keep the configured request
timeout and cancellation active during every read and capacity wait.
Give the whole operation one deadline derived from that configured timeout,
including the final wait for pending fragments to drain. An HTTP client's
timer can stop its transport while a parser remains parked on a separate
channel. The operation's waits must observe the same deadline.

That memory ceiling does not make assembly cheap. Copying the entire answer
for every small fragment makes each arriving word pay for all the words before
it. Doubling the answer can quadruple the copying. A terminal intended to show
progress can spend its time rebuilding text the reader has already seen.

Accumulate each part in owned, append-oriented storage and maintain response
size counters as content changes. Decode each incoming wire object once;
materialize the completed parts and run shared normalization at the completion
boundary. Avoid repeatedly parsing, serializing or scanning the growing answer
to append a fragment or measure its size. Storage choices remain open, but work
and allocations for fixed-size fragments should grow approximately with total
content size. Verify this with a local size-doubling benchmark, without a
machine-specific timing cutoff. Preserve the same exact bounds, UTF-8 handling,
part identities, coalescing boundaries, signatures and opaque material.

A streaming request requires an SSE success body. Accept media-type parameters
on `text/event-stream`; a success response containing ordinary JSON instead is
a safe delivery error, without a second request. Preserve existing HTTP-error
handling before decoding. A plain request continues to expect the ordinary
JSON envelope. Content-type differences do not authorize guessing a different
API from its payload.

This literal framing fixture uses LF after each printed line, including the
last blank line:

```text
: keepalive
event: message
data: {"a":
data: 1}

```

It produces one data string, `{"a":\n1}`, which is valid JSON. Repeat with
CRLF, lone CR, a leading BOM, and every possible split between input bytes.
Remove the final blank line and close the input: the reader must report
unfinished data to the adapter, not hand it a completed object. A comment-only
stream has no content response. These are local controls, not observations of
how any paid API happened to split its packets.

For an exact-limit framing control, these byte expressions describe the
complete event. They are framing fixtures, not model response payloads:

```text
LF:   b'data: "' + b'x' * (1_048_576 - 10) + b'"\n\n'
CRLF: b'data: "' + b'x' * (1_048_576 - 12) + b'"\r\n\r\n'
```

Each dispatches once. Add one `x` and it fails; prepend the leading BOM and
the exact-limit version still passes. Two exact-limit events concatenated
both pass because the counter resets. Also count ignored material: `:` plus
1,048,573 `x` bytes plus two LF bytes is one exact-limit comment event, ignored
after accounting. One more `x` must fail even though it has no data field.

Use static diagnostic reasons such as `stream ended before message_stop` or
`tool arguments are not a JSON object`. Include safe Agent/request identity
when useful. Do not log raw frames, authorization headers, URLs, prompts,
arguments, or provider error text. This chapter adds no production frame-trace
facility. A controlled fake server can retain its own published fixture bytes
without turning user conversations into debug logs.

## 6.4 Assemble what each API actually sends

Streaming changes delivery, while the earlier neutral parts and provenance
rules still govern retained content. Reuse semantic normalization for plain
and streamed responses. Separate decoders may recognize their different wire
envelopes; they must meet at shared validation of parts, arguments, signatures,
returned identity, and usage. Calling two unrelated decoders through one
function does not establish equivalence.

Keep the existing API surfaces. The logical provenance surface remains
`messages`, `chat_completions`, or `generate_content`; switching delivery does
not make an otherwise compatible signature foreign. Record requested delivery
in `request_sent.request.delivery`. Newly written events always carry it;
historical events without it remain valid. Validate any present value on
append and replay. Replay never emits provisional deltas or performs HTTP.

Historical request reconstruction uses the recorded choice, including the
plain meaning of an absent field, rather than today's Agent default. For
stream delivery it restores the adapter's stream flag, Chat Completions usage
option, or generateContent endpoint selection. These are deterministic
consequences of that recorded delivery and adapter version, not a second
mutable request renderer. Compare a pre-streaming log and a new plain log
under today's default-on configuration: their reconstructed requests remain
plain. A new stream record reconstructs the corresponding streaming options.

The CLI switch is delivery policy, not a claim about model capabilities. A model
that supports text streaming may expose no thinking text, and a complete tool
call may arrive in one event. Keep current discovery and explicit resolved-model
configuration. A provider's refusal is a normal safe request error. Do not
spend a second request automatically by retrying without streaming.

### Messages

Send `stream:true`. Track block indices through start, delta and stop events;
require their starts before deltas and reject reuse or changes of kind. Keep
final order by block index, allowing interleaved updates. Accumulate text,
tool argument JSON strings, exposed thinking text, and signature fragments
in their respective blocks. Preserve completed non-text blocks as opaque
material. Unknown standalone event types may be ignored for compatibility;
an unknown delta affecting an open block is a safe unsupported-content error
when it cannot be retained correctly.

Nonempty text or thinking supplied at block start belongs to the same part
and is observed before later deltas. Never drop that initial content. Changes
to a closed block fail, as do duplicate starts or stops.

The documented lifecycle ends with `message_stop`; ping events may appear
between content events, in-stream errors can follow HTTP 200, and usage in
`message_delta` is cumulative. Thinking and signature fragments are distinct
delta types. These are wire facts from the current
[Messages streaming guide](https://platform.claude.com/docs/en/build-with-claude/streaming),
checked October 7, 2026.

Initialize usage from `message_start.message.usage`, then replace fields
explicitly updated by subsequent usage snapshots. Omission does not erase an
earlier count. The final output count replaces an initial zero. Require all
started blocks to stop, a nonempty stop reason, and `message_stop`. Decode each
completed tool argument string to an object; an empty delta sequence may use
the complete object supplied at block start. If partial JSON arrives, it
replaces the initial placeholder object. Never combine two argument objects
by guessing what the model intended.

Publish a tool name when known. Publish argument deltas as supplied; when
there are none, publish the complete start object once at block stop. This
avoids showing `{}` as an argument prefix and later replacing it silently.

Retain the merged final usage object as `raw_usage`; it describes the accepted
response without retaining every cumulative intermediate snapshot.

### Chat Completions

Send `stream:true` and `stream_options:{"include_usage":true}`. Use choice index
0; ignore other choices without merging their content or usage into it. Text
deltas append to its text part. Tool-call deltas use their index to identify a
call; ID, function name, and argument strings may arrive separately. Retain the
call ID once supplied, reject a conflicting replacement, and concatenate name
and argument fragments in order. Final order is visible text when present,
then calls by tool index, matching the plain adapter.

The API documents a final usage chunk with an empty choices array when usage
is requested, followed by `[DONE]`. Usage can be absent after an interrupted
stream. Those facts are described in the current
[Chat Completions streaming reference](https://developers.openai.com/api/reference/resources/chat/subresources/completions/streaming-events)
and [request reference](https://developers.openai.com/api/reference/resources/chat/subresources/completions/methods/create),
checked October 7, 2026.

Require choice 0 to have a finish reason, the requested usage object to arrive,
and `[DONE]` to follow. Continue after content finishes so the usage-only chunk
is not lost. Decode each argument string as an object. Content null with valid
calls remains valid; a stream containing only role metadata is not an answer.
Preserve recognized refusal material as opaque rather than presenting it as
ordinary text. It does not waive the earlier visible-text-or-call requirement.
This chapter does not invent a generic reasoning field absent from the chosen
API's documented schema.

Retention must survive the next turn. For matching provenance, restore the
recognized opaque `{"refusal":"..."}` payload as the original assistant
message's string `refusal` field, alongside its visible content and calls.
Reject malformed or unsupported opaque shapes and conflicting duplicate
refusal material safely; keep the earlier foreign-target omission rule.
Exercise a subsequent human turn and a tool continuation in both delivery
modes. Accepting a response and then refusing to render that same recognized
content would strand an otherwise usable conversation.

Retain the final usage object as `raw_usage`. Across any adapter's operation,
nonempty returned model identities must agree. Missing identity follows the
existing requested/returned fallback rule; a conflicting later identity is
a parse error, not a reason to move earlier fragments to another model.

### generateContent

Use `:streamGenerateContent?alt=sse` for streamed delivery and retain
`:generateContent` for plain delivery. Read candidate index 0, with an omitted
index meaning 0 when there is only one candidate. Preserve ordered parts from
successive payloads. A functionCall object supplies complete arguments; there
is no requirement to manufacture partial JSON. Use a supplied call ID or leave
it for the actor's append-time rule.

Classify a complete Gemini part before publishing fragments from it. For a
text-bearing part, the recognized top-level fields are `text`, `thought` and
`thoughtSignature`. Validate their types first: text is a string, thought is a
Boolean when present, and a signature is a string when present. A malformed
known field is an error, even if an unknown field accompanies it. If additional
unrecognized fields remain, retain the entire original JSON part as one
provenance-bound opaque part. Do not discard the extra fields, split out a
second text part, or emit visible/thinking deltas from text whose surrounding
meaning the adapter does not understand. An accepted opaque final still carries
its original part position and identity.

Use that same classification in plain and streaming normalization. Matching
provenance restores the whole original part; a foreign target omits it under
Chapter 2's standalone-opaque rule. This conservative choice does not invent a
schema for a future field. A recognized signed text part continues to use the
existing text-plus-signature representation, and recognized `thought:true`
material keeps its existing non-answer semantics. `thought:false` has the same
ordinary-text meaning as an absent thought flag and permits coalescing when
no signature or unknown field is present. `thought:true` plus an unknown field
falls under the whole-object opaque rule and emits no synthetic thinking delta.
This clarification concerns text-bearing parts; it adds no new function-call
classification or dispatch permission. Opaque material alone still
cannot satisfy the required visible-text-or-call answer boundary.

Coalesce adjacent unsigned text-only parts into one text run, including across
frames. Do the same in plain normalization. Any call, signature, opaque field,
or change between ordinary and thought text ends that run. Contiguous unsigned
thought-text-only parts may form one opaque thought part. Preserve a signed
part at its exact position, including empty signed text. Do not move its
signature onto an earlier text run. The
[signature guide](https://ai.google.dev/gemini-api/docs/generate-content/thought-signatures)
explicitly documents an empty text part carrying a signature during streaming;
the whole response must be read, not merely the first visible answer.

Treat each usageMetadata object as the latest complete usage snapshot, never
as an increment. Require the final available snapshot to satisfy Chapter 2's
base-count rules. Read through orderly EOF after candidate 0 reports a
nonempty finishReason; allow later metadata-only frames. EOF before that reason,
a pending unfinished frame, an explicit error object, or a read error fails
the operation. The [generateContent reference](https://ai.google.dev/api/generate-content)
defines an empty finishReason as generation still in progress. The selected
surface is currently labeled legacy in its guide; migrating to another API is
outside this exercise.

### Preserve meaning at the stop boundary

A provider's complete token-limit stop differs from a broken stream. Retain
the raw stop reason on each newly captured response as optional string
`response.stop_reason`, including plain responses when supplied. Historical
events may omit it. An accepted, valid text answer stopped by `max_tokens`,
`length`, or `MAX_TOKENS` is displayed with a clear generation-limit notice;
the request can still complete successfully in the existing sense of returning
an accepted answer. It does not claim the requested task was fully accomplished.

Validate every complete part before accepting any of the response's calls.
An unfinished argument object remains invalid even when the provider supplied
a token-limit stop. A fully framed response with valid parts follows the
same acceptance rules as its plain equivalent; this chapter does not invent
a new finish-reason allowlist for calls. A malformed-call error is still an
error. Record a present stop reason without making its absence in an older
plain fixture a new failure. Transport completeness and semantic validation
are separate checks, and both must succeed before an effect.

Provisional text is not a usage receipt. Normalize the final usage once and
commit it once with acceptance. If required usage never arrives, fail rather
than publishing zero cost. Discarding an incomplete operation's accounting
does not assert that the provider charged nothing; retained totals continue
to mean usage from accepted responses, as in Chapter 5.

## 6.5 Small fixtures with decisive failures

The following are synthetic local streams. `fixture-*` names are explicit
offline identities, not live model recommendations. Each shown event ends in
a blank line; repeat any fixture with one-byte reads. These bodies assume a
successful HTTP response with `Content-Type: text/event-stream`.

The Messages fixture yields text `Hello.`, input 10 and output 2:

```text
event: message_start
data: {"type":"message_start","message":{"model":"fixture-messages","content":[],"usage":{"input_tokens":10,"output_tokens":0}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hel"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"lo."}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":2}}

event: message_stop
data: {"type":"message_stop"}

```

Its plain counterpart has that model, one text block `Hello.`, stop_reason
`end_turn`, and usage `{ "input_tokens":10, "output_tokens":2 }`. Hold the
server after `Hel` until the public observer and terminal have exposed it;
only then release `lo.` and completion. An implementation that buffers the
whole response cannot pass that handshake.

The Chat Completions fixture proves that content can finish before accounting:

```text
data: {"id":"fixture-response","model":"fixture-chat","choices":[{"index":0,"delta":{"role":"assistant","content":"Hel"},"finish_reason":null}]}

data: {"id":"fixture-response","model":"fixture-chat","choices":[{"index":0,"delta":{"content":"lo."},"finish_reason":"stop"}]}

data: {"id":"fixture-response","model":"fixture-chat","choices":[],"usage":{"prompt_tokens":10,"completion_tokens":2}}

data: [DONE]

```

Its plain counterpart has choice 0, assistant content `Hello.`, finish_reason
`stop`, and the same usage. Removing the usage event must fail rather than
increment output by zero. Removing `[DONE]` must fail even though a finish
reason was received. Other choices may contain distracting text; none belongs
to choice 0's final answer.

The generateContent fixture preserves a signature after the visible text:

```text
data: {"modelVersion":"fixture-gemini","candidates":[{"index":0,"content":{"role":"model","parts":[{"text":"Hel"}]}}]}

data: {"modelVersion":"fixture-gemini","candidates":[{"index":0,"content":{"role":"model","parts":[{"text":"lo."},{"text":"","thoughtSignature":"fixture-signature"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":2}}

```

The plain counterpart has parts `[{"text":"Hello."},{"text":"",`
`"thoughtSignature":"fixture-signature"}]`, the same identity, STOP and usage.
Both normalize to an unsigned `Hello.` part followed by a distinct signed empty
part. A renderer to matching provenance must retain that second part. Withhold
the second frame and close: no accepted answer or usage survives.

Add a Gemini plain/stream pair with these ordered parts and otherwise valid
identity, STOP and usage:

```json
[{"text":"A"},{"text":"hidden","future_payload":{"tag":"retain-me"}},{"text":"B"}]
```

The neutral response contains text `A`, one opaque part holding the entire
middle object, then text `B`. Its visible answer is `AB`; the unknown part
emits no text or thinking delta. The two text parts must not coalesce across
it. Matching replay restores the middle object with `future_payload` intact;
foreign rendering omits that object and preserves `A` and `B`. Removing both
recognized text parts leaves an opaque-only response and must fail acceptance.
Changing the middle `text` to a number must also fail, rather than conceal a
malformed known field inside opaque storage. These fixtures use a fictional
field to test preservation, not a claimed current provider feature.

Add a two-call fixture using the actual `read_file` and `list_directory`
declarations. For Messages use block indices 0 and 1, with IDs `read-a` and
`list-b`; for Chat Completions use tool indices 0 and 1 with those IDs. Interleave
argument fragments so read-a receives `{"path":` then `"notes.txt"}`, while
list-b receives `{"path":"` then `."}`. Preserve each exact resulting object
and name. generateContent supplies those two complete functionCall objects
in ordered parts. A complete call has normal call-capable finish and the same
required usage as the text fixtures.

Repeat a response with local part IDs reused in a new operation, then run two
Agents with overlapping local IDs. Compare fragments by their full identity.
Joining all argument fragments into one buffer and finding both names proves
nothing about which file the first call will read.

For Messages thinking, include an opaque thinking block whose deltas expose
`Checking.` and whose signature fragments form `signed-value`. Compare it with
the same complete plain block. For Gemini include `thought:true` text and a
signed call. The final retained parts must agree with their supplied plain
counterparts; thinking text must not enter the visible answer string. None
of these fictional signatures is evidence that a real provider accepts them.

## 6.6 Failure after something has appeared

The reader has seen half an answer when the socket closes. Clearing the screen
would hide useful context; silently leaving the text under a successful answer
heading would lie about its status. Keep the provisional text and label it
incomplete with the safe reason. The reliable handle still determines whether
the human request succeeded, was interrupted, or failed.

On parser error, timeout, cancellation, provider error, or missing terminal
signal, deliver one rejected model end and apply Chapter 5's request-failure
or interruption path. No accepted response event or usage increment is added
for that operation. Earlier accepted tool rounds, their effects, and running
jobs remain. A later turn may be admitted to an unfaulted public Agent; the
CLI retains its existing fatal-error policy for ordinary request failures and
remains alive after intentional interruption or cancellation.

An actor that accepts interruption first rejects subsequent facts from that
operation, including an apparently successful late result. If it has already
accepted the response, interruption cannot erase it; existing Chapter 5 rules
settle the remaining turn. Close cancels and joins the managed model reader.
No automatic reconnection, continuation from partial text, or second paid
request is part of recovery here.

Observers remain optional progress consumers. A full subscriber queue closes
with its explicit overflow reason, without blocking the actor or consuming a
completion. The CLI marks its display incomplete and uses the reliable handle
to show a clearly labeled complete final answer when available. Reprinting in
that recovery case is deliberate; normal operation never prints the answer
twice. A client that chooses finals only still receives complete typed parts.

## 6.7 Make streaming visible to a human

Use the existing `chat` command and actual terminal detection. Prefix an
operation's output with its request identity, then write and flush each
available visible-text fragment. Print exposed thinking under a separate
`thinking` label and tool names/argument fragments under a `proposed tool`
label. Keep signatures and other opaque fields off the terminal. Input,
acknowledgements, and output share serialized writes so a hint cannot split
another record halfway through its bytes.

The display need not repaint a perfect editor line. It must remain readable
when `/hint`, `/interrupt`, `/usage`, or an ordinary queued prompt arrives while
text is appearing. A normal accepted end closes the provisional display; the
request completion adds its final status without copying text already shown.
Plain mode displays the accepted text once. The same final typed response and
usage reach public consumers in both modes.

For machine inspection, `protocol --observe` opts into additional JSON lines:

```json
{"observation":{"kind":"model_begin","agent_id":"a1","request_id":"r1","operation_id":"m1","delivery":"stream"}}
{"observation":{"kind":"part_delta","agent_id":"a1","request_id":"r1","operation_id":"m1","part_id":1,"channel":"text","text":"Hel"}}
{"observation":{"kind":"part_final","agent_id":"a1","request_id":"r1","operation_id":"m1","part_id":1,"response_seq":8,"part_index":0,"part":{"type":"text","text":"Hello."}}}
{"observation":{"kind":"model_end","agent_id":"a1","request_id":"r1","operation_id":"m1","accepted":true,"response_seq":8}}
```

The omitted `lo.` delta is only an abbreviation in this shape example; the
actual text fixture emits all its text. The flag exposes these four new kinds
and preserves Chapter 5's acceptance, acknowledgements and completion records.
It need not dump every existing durable observation. A request's accepted
record precedes its model begin, and model end precedes its completion. Legacy
`{"user":"..."}` has no accepted record and still receives its one assistant
record in this explicitly expanded mode. Without the flag its exact old output
remains unchanged. Unknown CLI flags fail locally before any request.

That observation ordering applies while the subscription remains intact.
On overflow, observe mode emits
`{"observation_gap":{"agent_id":"a1","reason":"overflow"}}` once for that
closed subscription and continues delivering reliable completions. It cannot
promise the missing finals or end. Serialize this loss notice before a
completion that must recover the missing display. An optional new subscription
can observe later operations; it cannot retrospectively fill the gap. Waiting
to print ordered progress must never make library completion depend on a
subscriber's acknowledgement.

A parser emitting fragments is only half the feature. A buffered writer can
do its job perfectly and hide every word until the answer is finished. The
parser and writer both look correct; the person staring at the terminal gets
the old silent wait.
The held-open fixture must inspect the actual human PTY as well as a public
observer. A line-count check after process exit cannot detect that failure.

## 6.8 Taking it for a spin

Start with a file whose answer is easy to check. Build from
`solutions/edition-2/main`, then launch human chat in a scratch directory:

```sh
go build -o /tmp/ensemble-ch06 ./cmd
demo_dir=$(mktemp -d)
cd "$demo_dir"
printf 'CHAPTER-SIX-FILE-MARKER\nport=8080\n' > notes.txt
CH02_LOG="$demo_dir/session.log" /tmp/ensemble-ch06 chat
```

Use the earlier chapters' environment-held credentials and discovered model
configuration. Keep `LLM_RESOLVED_MODEL` accurate when provenance requires it.
No streaming switch is needed for the default. Ask for a bounded explanation,
watch actual words arrive, and type `/hint Keep the final answer concise.`
while that request is active. A hint acknowledged as pending reaches the next
request; it cannot edit the request already on the wire.

The student coder drove these actual PTYs on October 7, 2026 local time
(October 8 UTC). The initial runtime is `aa5f86a`; no runtime edit was needed
during the nine live sessions. Bill did not run these demonstrations. The
selected Messages model was `claude-haiku-4-5-20251001`, which returned the
same identity. Chat Completions selected `gpt-4.1-mini` and returned
`gpt-4.1-mini-2025-04-14`. generateContent selected
`models/gemini-3.8-flash` and returned `gemini-3.8-flash`. These are dated run
identities, not a permanent recommendation table.

The first prompt requested 30 short numbered paragraphs. Messages remained
active long enough to receive the hint. Chat Completions finished before the
coder's next observation call; a second, longer bounded prompt provided the
opportunity. With Gemini, the first hint arrived while idle and received:

```text
Command error: hint requires nonempty text and an active turn.
```

A subsequent request for 150 sentences ended with `MAX_TOKENS` after only a
short visible answer. Chat printed `Generation limit reached; accepted answer
may be incomplete.` The request had produced a valid response. It had not
produced 150 sentences. Asking for a bounded sequence of integers then made
early output easy to observe, and an active hint succeeded. All three logs
retain hint receipt separately from consumption in the next model request.

Next, enter this ordinary human prompt:

```text
Use read_file to read notes.txt. Report its exact marker and port, then stop. Do not modify files.
```

The Messages terminal showed the proposed call before its continuation. This
excerpt omits intervening acknowledgement and completion lines:

```text
[r2 proposed tool part 1] read_file
[r2 proposed tool part 1] {"path": "notes.txt"}
Request r2 / m4 (stream)
[r2 Assistant part 1] **Exact marker:** `CHAPTER-SIX-FILE-MARKER`

**Port:** `8080`
Request r2 (success; pending hints=0)
```

Each API retained the accepted call and matching result, then returned the
marker and port. The six reads across streamed and plain sessions left six
tool artifacts containing exactly the two source lines. The artifact provides
a better check than a model saying that it read the file. A local fixture
supplies the stronger boundary test: interrupt or truncate a complete-looking
proposed write and require that no file effect occurs.

For interruption, ask for another bounded long answer. After seeing a fragment,
enter `/interrupt`, then ask `Reply exactly RECOVERED-SIX. Do not use tools.`
The Messages run included:

```text
Interrupt: request=r3 interrupted=true.
You>
Incomplete display: interrupted: turn interrupted

Request r3 (interrupted; pending hints=0)
interrupted: turn interrupted
You> Reply exactly RECOVERED-SIX. Do not use tools.
Accepted r4.
You>
Request r4 / m7 (stream)
[r4 Assistant part 1] RECOVERED-SIX
Request r4 (success; pending hints=0)
```

The retained Gemini attempts show why the acknowledgement matters. One
interrupt arrived after completion and reported `interrupted=false`. The next
attempt used a bounded integer stream; the coder read its first visible
fragment and immediately interrupted. That request ended as interrupted, and
the following turn returned `RECOVERED-SIX`. All three APIs eventually
exercised actual interruption and recovery. Their interrupted operations have
no accepted response or usage increment; that accounting rule does not claim
the provider charged nothing for discarded output.

Restart with a separate log and plain delivery:

```sh
CH02_LOG="$demo_dir/plain.log" EN_DISABLE_STREAMING=1 /tmp/ensemble-ch06 chat
```

Enter `Reply exactly PLAIN-SIX. Do not use tools.`, then the same file prompt.
All three plain sessions returned `PLAIN-SIX` and the marker/port answer once,
without provisional fragments. Independent paid generations can use different
words and token counts. The paired local fixtures test exact normalized
content equivalence; these sessions test whether a person can use both modes.

The streamed human sessions retained these accepted-response totals. Cache
columns stay separate from input; output follows the normalization already
taught in Chapter 2.

| API | Input | Cache write | Cache read | Output |
|---|---:|---:|---:|---:|
| Messages | 13,133 | 0 | 0 | 960 |
| Chat Completions | 1,930 | 0 | 10,112 | 1,746 |
| generateContent | 52,233 | 0 | 0 | 10,730 |

The public example in `examples/stream-consumer` adds a different user path.
From that module, build its executable, give `ENSEMBLE_RUN_DIRECTORY` a fresh
directory, and use the same model environment. It creates two Agents with
ordinary, finals-only and deliberately stalled subscribers. On each real API,
both reliable completions arrived before the stalled callbacks were released.
Agent identity kept their otherwise repeated `r1`, `m1` and part numbers
separate; finals-only typed parts matched their corresponding completions.

Those small public runs did not overflow a subscriber. None of the selected
models exposed a supported thinking delta. Local fixtures cover guaranteed
overflow and exposed thinking separately; a signature is not visible reasoning.
Gemini's two public requests each reached the example's 512-token limit and
retained signed empty text parts. Their partial explanations were accepted
responses, with `MAX_TOKENS` recorded, rather than completed explanatory tasks.
The optional GUI remains a stub.

The [run summary](../../solutions/edition-2/main/evidence/ch06/live-summary.json)
links the nine sessions, original terminals and logs. The retained verifier
receipts reconstruct all 33 captured requests semantically from their logged
prefixes and configuration, including hint inclusion and stream options.
That comparison does not assert identical JSON serialization or identical
answers from separate generations. See the [evidence record](chapter-06-evidence.md)
for source binding, read scope and the still-open independent review gates.

## 6.9 What the checks must establish

| Contract | Required distinguishing check |
|---|---|
| Early output | A server barrier waits for the first fragment at both public observer and actual human PTY before sending completion; a buffer-until-end mutant fails |
| Correlation | Two interleaved calls, text, repeated local IDs in later operations, and two Agents retain separate full-identity groups |
| Accepted content | Paired plain/stream fixtures compare all ordered typed parts, exact arguments, signatures, requested/returned identities and normalized usage; exclude delivery and transient IDs deliberately |
| Framing | Every byte split, UTF-8, CRLF/CR, comments, multiline data, exact frame limit and one-byte overflow, unfinished frame, explicit error and missing API terminal |
| Accounting | Cumulative updates replace prior values; usage-only final is read; absent required usage fails; accepted response increments exactly once |
| Storage and assembly | Exact retained-content limit and one-byte overflow include escaped opaque JSON; fixed-fragment size-doubling measurements detect repeated rebuilding of accumulated content |
| Effects | Pause after complete-looking call arguments, then truncate or interrupt; require no tool effect, accepted response or usage for that operation |
| Concurrency | A job terminal event consumes a sequence during HTTP; final call ID and observation mapping use the actual later response sequence |
| Lifecycle | Interrupt, timeout, close and stale facts settle the right handle once; intentional interruption leaves the next turn usable |
| Display failure | Overflow closes one subscription with an explicit reason while reliable completion and another subscriber work; CLI labels partial display and recovers the complete result |
| Compatibility | Default protocol remains exact, observe mode has correlated complete JSON lines, plain mode emits no deltas, replay performs no streaming work, optional GUI stays optional |
| Architecture | Discover every spoke and executable; verify the actual Engine-operation owner chain and parser diagnostic reachability, with actor-only persistence and no callback dependency bag |

Keep the initial student source and runs before the independent reviewer reads
the first-edition standard. That reviewer compares design, clarity, diagnostics,
comments, behavior and tests, then supplies reasons for revisions. The student
does not open the old answer. The author incorporates missing teaching, and
affected deterministic and live checks are repeated before review closes.

Record difficulties immediately in `evidence/ch06/student-review.md` under
main and notify the author. Preserve the initial account before comparative
feedback, then append what clarification and actual use changed. The author
records dispositions in `chapter-06-student-feedback.md`; the student checks
that material questions were resolved. A passing build still needs this
teaching review.

The initial implementation and live receipts remain preserved before comparison.
The inherited 0/100 result retains its original scope. Independent review
identified refusal replay, repeated assembly work, timeout coverage and example
observer ownership defects; the revised source addresses them, including a
second correction for escaped opaque-byte accounting. The complete deterministic
gate, affected revised demonstrations and final proofreading must close in the
linked gate record before this chapter receives a validated checkpoint.
