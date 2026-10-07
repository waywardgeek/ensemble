# Chapter 2: One History, Three APIs

Changing the model should not require rewriting what happened. A tool
result still came from a tool, even if the next API wants it inside a
user message. A cached token still belongs to the response that reported
it, even after another model takes over. Borrow one API's vocabulary for
the whole program and every new provider has to argue with those decisions.

Chapter 1 made the requests visible. This chapter gives them a history
you can trust when an answer goes wrong: what arrived, what was sent,
what the model returned, and what later changed. Three adapters will
read that same history without each acquiring its own conversation.

## 2.1 The idea in plain words

When an answer is wrong, the useful question is what the model received.
Suppose a recorded tool result contains `port=8080`, then a later request
replaces that result with a short stub to save context. If the program
overwrites its saved message, the evidence of the earlier request is gone.
If it keeps sending the original result, the attempted reduction did nothing.

Keep the original fact and record the decision to replace it. The program
can then reconstruct the request before the decision and the one after it.
You can examine both without asking the model to explain its own memory.

**The log records what happened.** A user supplied a prompt, the engine
sent a request, a model returned parts and usage. Each fact receives an
ordered identity. Changing what the next request should contain does not
change the record of what previously occurred.

**Context is the current reading of that record.** A reducer applies the
events in order. If an event says to replace a tool result with a stub,
the next context contains the stub and the log still contains the original
result. Replaying the same events reconstructs that choice without asking
a model or deciding again whether redaction was a good idea.

**A request is a provider's view of the context.** A tool result might
become a user-message block on one API and a separate tool message on
another. The author of the result remains a tool. A wire convention
changes the rendering, not who produced the information.

Keep Chapter 1's ownership chain and package boundaries. Shared events,
parts, and interfaces belong in `internal/common`; reducer, rendering,
and parsing behavior belong in `internal/llm`. Ensemble still owns the
logger and Agents, Agent owns configuration and its conversation, and
Engine owns transport and accounting. Replace the narrow text slice
with events and a derived context inside that architecture.

> **Student contract ready for implementation.** The commands, schemas,
> and fixtures below specify the reviewed exercise. No Chapter 2
> implementation, successful grade, or live demonstration is claimed.

## TL;DR

Continue the validated Chapter 1 repository and its history in
`solutions/edition-2/ch02`. Keep its public library and CLI. Add three
non-streaming adapters: Messages, Chat Completions, and generateContent.
Read current model IDs from provider discovery for live runs. Before
implementation, reload the entire
[`book/edition-2/skills/ensemble-coding/SKILL.md`](skills/ensemble-coding/SKILL.md).

1. Agent owns an append-only, versioned JSON-lines event log and its
   derived context. Each event has a positive sequence number, kind,
   timestamp metadata, and exactly the matching payload. Sequence is
   ordering authority. Reject malformed records, unknown kinds, and
   unsupported versions with line-numbered diagnostics.
2. Use one append/apply path for human inputs, parsed responses, and
   recorded redactions. Parsers return facts; they cannot mutate context.
   Copy slices and raw bytes when taking ownership. Replaying a fixed
   log rebuilds the same context and usage without network calls.
3. Entries have actor, purpose, sequence, and ordered typed parts. Keep
   text, calls, results, references, opaque replay material, and redacted
   stubs distinct. Purpose distinguishes dialogue, enduring instruction,
   and one-request ephemera; actor alone cannot make that distinction.
4. Every model response and opaque/call-bound replay part records vendor,
   exact model, and API surface. Replay opaque bytes only to matching
   provenance. Preserve issued call IDs; deterministic synthesized IDs
   cover routes that omit them. Do not dispatch tools in this chapter.
5. Normalize response usage into disjoint ordinary input, cache write,
   cache read, and output counts. Keep the producing provenance and raw
   usage observation. Engine aggregates by provenance; changing current
   configuration cannot change historical attribution.
6. Render without mutation or network access from context plus explicit
   request configuration. Keep credentials outside log, context, and
   rendered output. Fixed log, configuration, and adapter produce
   identical bytes. Consuming pending ephemera belongs to the recorded
   request event, never to the act of rendering.
7. Record result redaction as an event naming an inclusive sequence span.
   Replace result content in context with a deterministic stub, preserving
   call/result linkage and any external reference. The original log stays
   unchanged. Do not add an unbounded second set of redacted sequence IDs.
8. Preserve Chapter 1's successful CLI protocol and public-library use.
   Add offline `render LOG` and `dump` commands plus explicit log selection.
   Valid present empty text and tool-only responses are now representable;
   absence, malformed payload, and HTTP failure remain distinct errors.
9. Expose public, transport-neutral client requests and observations on
   Ensemble. Both CLI and a GUI stub use them. The GUI stub lives in a
   separate optional Go module; the core has no GUI or WebSocket dependency.
   Observer events carry Agent identity; synchronous requests return their
   own answers instead of waiting for a broadcast to imply completion.
10. Demonstrate every implemented feature through actual user paths.
    Run the CLI with all three real APIs, plus an external public consumer.
    Label GUI-stub integration and deterministic failure fixtures as such;
    neither is a working live browser demonstration.

**Yours.** Internal names and exact method signatures, provided the
ownership and schema contracts hold. No tool execution, jobs, actors,
streaming parser, skills, automatic compaction, retries, or functioning
browser interface is required here.

**Exercise.** The inherited grader is a useful regression baseline:

```sh
make grade-dir CH=2 DIR=solutions/edition-2/ch02
```

The score covers the inherited wire and replay checks. The additional
acceptance properties in §2.9 are required too. The complete student handoff
includes §§2.2–2.9 and Chapter 1's architecture rules, not just this list.

## 2.2 A fact needs an identity

A timestamp cannot order two events that happened during the same clock
tick. Give every event a positive integer `seq`, assigned by its Agent.
A newly written log starts at 1 and increments by 1. A loaded log may have
gaps, but sequence numbers must strictly increase. The timestamp is UTC
RFC3339 metadata; it never settles ordering. Within an event, the position
of a part supplies its stable second coordinate. Keep array order.

The file begins with this line:

```json
{"log_version":1}
```

Every subsequent nonblank line is an event with `seq`, `type`, `time`,
and exactly one payload selected by the table. The envelope and payload
spellings below are the interchange format; Go names remain your choice.
Treat `generatecontent` and `generate_content` as the same API surface
when reading version 1, and write `generate_content`. The other surface
names are `messages` and `chat_completions`.

| Event type | Payload key | Fields |
|---|---|---|
| `message_received` | `message` | `actor`, `purpose`, ordered `parts` |
| `request_sent` | `request` | `to` provenance, `ephemera` sequence array |
| `response_started` | `response` | `from` provenance; optional `requested` |
| `response_ended` | `response` | `from`, ordered `parts`, normalized `usage`; newly captured responses also carry `requested`, `model_reported`, and `raw_usage` |
| `tool_called` | `tool` | `call_id`, `name`, JSON-object `args` |
| `tool_returned` | `tool` | `call_id`, ordered `parts`, optional `is_error` |
| `redacted` | `redact` | inclusive `from`, `to` sequence numbers, `level`, `reason` |
| `error_occurred` | `error` | safe `code` and `message` strings |

`response_started` makes room for streaming later; this non-streaming
implementation need not emit it. A tool-call event records dispatch and
adds no duplicate assistant content. This chapter loads such events from
fixtures but executes no tools itself.

Validate before application. Refuse an unsupported version, missing
header, repeated or decreasing sequence, invalid timestamp, unknown event
or part type, wrong payload, or malformed required value. Name the line
number without echoing its contents. Ignore unknown additional metadata
fields, but reject an additional *known payload key* on an event. This
permits harmless annotations without treating an unknown operation as
one the reducer understands. Loading is all-or-nothing: publish the
loaded Agent only after the complete log validates.

An entry's `actor` is `human`, `agent`, `system`, or `tool`. Its `purpose`
is `dialogue`, `instruction`, or `ephemeral`. Instructions are enduring
system content; ephemera belong to one request. Both use actor `system`.
Human, agent, and tool entries are dialogue. Version-1 fixtures written
without `purpose` mean dialogue for those three actors and ephemeral for
system. Newly written messages always state purpose. This compatibility
rule is confined to loading old records; it is not a reason to leave the
distinction out of the data structure.

In this chapter, `message_received` accepts human dialogue or system
instruction/ephemera only. Reject imported agent/tool variants too.
Agent dialogue enters through `response_ended`, and tool dialogue through
`tool_returned`; the general entry vocabulary must not bypass their
validation and accounting paths.

The parts carry meaning independently of a provider's role vocabulary:

| `type` | Required data | Meaning |
|---|---|---|
| `text` | present string `text`, including `""`; optional paired `from` and JSON `opaque` | Text at this position, with replay metadata when supplied |
| `tool_call` | nonempty `call_id`, `from`, nonempty `name`, JSON-object `args`; optional JSON `opaque` | A requested call and material bound to it |
| `tool_result` | nonempty `call_id`, ordered `parts`; optional Boolean `is_error` | Result of that call |
| `opaque` | `from`, present JSON `data` | Provider material that must survive without interpretation |
| `blob` | nonempty `mime`, `ref` | A reference to external content |
| `redacted` | nonempty `stub`; optional `ref` | Deliberately replaced content with a retained locator where available |

Results contain text, blob, or redacted parts, not nested calls or results.
A `tool_returned` event becomes one tool-result part in a tool entry.
Reject an unknown call ID, a second result for the same call, or a call
ID already used by another call in this Agent's history. Preserve a
provider's issued ID exactly. Where generateContent omits it, synthesize
`call-<response-seq>-<part-index>` using a zero-based part index; replay
therefore makes the same ID without a random number or global counter.

A reference is `{"kind":N,"locator":"..."}`: 1 means local path,
2 means remote URI, and 3 means application-managed handle. All require
a nonempty locator. Reject zero, unknown kinds, and the former blob
`path` field without a `ref`. Silently guessing a path would turn a remote
file or buffer handle into a filename. No loader fetches the reference.
This chapter renders a URI blob in a human entry to Gemini `fileData`.
Inside a tool result, its representation on all three adapters is descriptive
text containing the MIME type and URI; §2.5 specifies that mapping.
Report an unsupported-reference error for a path, handle, or an adapter
without a mapping. The dump command preserves all three kinds.

Own the captured data. An event must survive reuse of the parser's byte
buffer and mutation of the caller's slice. Copy at the ownership boundary,
including nested raw JSON. A public snapshot or observer notification must
not let its recipient mutate the authoritative log through a shared slice.

## 2.3 One application path, including failures

A failed request leaves two facts worth keeping: the user asked a question,
and no answer was accepted. Deleting the question hides the attempt. Leaving
it in the next request mixes an unanswered question into completed turns. The log
retains the failed attempt while the current context excludes it.

Agent owns the log and current context. Engine owns the accounting. An
accepted event travels through one application path that updates each
owner's state. Replaying starts with fresh owners and uses the same path;
it does not add old usage onto an already populated Engine. Context may
expose a read-only view of usage, but contains no second mutable total.

For a prompt, first append `message_received`. It becomes the pending
human input. Render from the completed dialogue, this pending input, and
pending ephemera. Immediately before attempting the HTTP request, append
`request_sent`; its `ephemera` lists exactly the pending ephemeral event
sequences included in that request. Apply the event and consume that set.
A successful parsed response appends `response_ended`, commits the
pending input and assistant parts to context, and adds its usage once.

If transport or parsing fails, append a sanitized `error_occurred`, discard
the pending input from future request context, and return the error. The
log retains both the attempted input and the failure. Chapter 1's rule
that a failed question must not poison the next conversation still holds;
the durable history can now explain why that question is absent.

Validate the input, selected live configuration, and absence of unresolved
calls before recording a new prompt. Rejecting any of those makes no log
event and no HTTP request. A render can still fail after the prompt was
captured, for example because a retained reference has no mapping on the
selected surface. Record an error and discard that pending prompt, but
preserve pending ephemera: without `request_sent`, no attempt consumed them.
An error event ends the current pending attempt; it cannot roll back an
earlier successful response or its usage.

Serialize operations within each Agent: at most one pending human input
and one active request. Refuse a second human input while the first is
pending and a second request while one is active. A response completes
an active request, a pending human input in a compact imported record,
or the continuation made possible by a completed set of tool results.
That last case preserves `human -> assistant calls -> tool results ->
assistant` without inventing another human. A response in any other state
is unsolicited and fails validation. A new human input is allowed once
all calls have results; its subsequent response consumes that pending input
and the prior results together. No automatic continuation request is made
in this chapter.

`request_sent` requires a pending human input or a completed set of tool
results awaiting model continuation, and no unanswered calls. It cannot
create a valid response slot from an otherwise empty or completed turn.

A `tool_called` record must name an existing, unanswered call with the
same name and arguments and cannot dispatch that call twice. Validate
these transitions before writing an event; replay applies the same checks.
Different Agents retain separate state and sequence spaces.

An imported version-1 fixture may omit request/start records and place
`response_ended` directly after a human input. Apply it as a completed
response. It consumes no ephemera because no request event records their
delivery. Newly captured HTTP attempts must include `request_sent`.
In old request records without `ephemera`, consume all ephemera pending
at that point. In new records, an explicit empty array means none.

Persistence precedes observation. Validate and write an event completely
before exposing its applied state or publishing it to clients. A failed
write faults that Agent for all subsequent writes; it cannot claim a durable
answer or continue appending after a possibly partial record. Return an error
until the application creates a fresh Agent with a fresh log destination.
This chapter needs no crash-recovery repair of a partially written final
line: loading detects and refuses it. Do not append a speculative event
and then edit the earlier bytes when an operation succeeds.

Offline rendering has no side effects. Calling it twice leaves pending
ephemera, accounting, and the log unchanged. A failed request still consumes
its recorded ephemera: they were designated for that attempt. This chapter
has no retry policy that would give them another delivery.

Tool-only replies are successful facts even though this chapter cannot
execute them. Return the concatenation of text parts, which may be empty,
and expose the typed calls through the public result/log. EOF after that
reply succeeds with usage. Another prompt while calls remain unanswered
fails before HTTP; silently dropping calls would corrupt history.
A tool-result event supplied through the public append path can complete
the pairing. That is data ingestion, not a tool runner.

## 2.4 The model that produced the bytes

Provenance has three nonempty strings: `vendor`, `model`, and `surface`. Use vendors
`anthropic`, `openai`, and `gemini`. A request records the selected route in
`request.to`; a response records it again as `response.requested`. If the
response includes `model` or `modelVersion`, store that returned identifier
in `response.from.model` and set `model_reported` true. Otherwise use the
requested identifier and set it false. Never claim an absent identity was
reported by the provider. Old imported fixtures may lack these extra
fields; their `from` remains the recorded attribution.

Usage belongs to `response.from`, even after the Agent selects another
model. Keep Engine's totals keyed by this provenance and expose both the
map and its aggregate. Preserve the raw usage JSON beside normalized
counts so a later accounting correction can inspect the original fact.
Credentials belong to request configuration only and never enter either
provenance object.

Opaque material has a stricter rule than visible text. A renderer may
return it only when vendor, model, and surface match the target exactly.
There is no family-name match. Current request configuration may name an
explicit resolved model identity, separate from its routing alias; the
CLI exposes this as `LLM_RESOLVED_MODEL`. Without it, the exact configured
model is the comparison identity. Do not infer that two different names
are aliases because a response happened to use both.

Standalone opaque parts from another provenance are omitted. A call with
foreign call-bound opaque material cannot be replayed safely by dropping
that field: reject the render with an incompatibility error. Calls without
such material still translate across providers. Both outcomes preserve
the recorded original; neither edits the log to fit the new destination.
An offline render needs the same explicit resolved identity as a live
render if the recorded model and selected alias differ.

For Messages, retain non-text, non-call content blocks as opaque parts.
For Gemini, retain `thoughtSignature` on the exact call or visible text
part it accompanies, using that part's `from` and `opaque` fields.
A same-target renderer restores the signature on that part and must not
merge it into a neighboring text part. With foreign provenance, visible
text survives without its signature; a signed call follows the stricter
incompatibility rule above. Parts marked `thought:true` become standalone
opaque parts and contribute nothing to the CLI's text answer. Other
standalone replay material retains its original part and provenance.
The [generateContent signature guide](https://ai.google.dev/gemini-api/docs/generate-content/thought-signatures),
checked October 7, 2026, documents signatures on text as well as calls.
Chat Completions function arguments arrive as a JSON-encoded string: decode
them to an object before recording a neutral call. A string containing
broken JSON is a parse failure, not an empty argument object.

## 2.5 Render meaning into three wire formats

Keep the adapters behind the same library operation. Their parsers return
parts, provenance, and usage; their renderers consume context and explicit
configuration. Neither owns conversation state. Adapter helper functions
retain their Engine owner context so diagnostics still reach Ensemble's
logger through Engine and Agent.

| Surface | Request | Response fields used here |
|---|---|---|
| Messages | `POST /v1/messages`; `model`, positive `max_tokens`, `system`, `messages` | ordered `content`; `text`, `tool_use`, opaque blocks; `model`, `usage` |
| Chat Completions | `POST /v1/chat/completions`; `model`, positive `max_completion_tokens`, `messages` | first choice's `message.content`, `message.tool_calls`, top-level `model`, `usage` |
| generateContent | `POST /v1beta/models/MODEL:generateContent`; `systemInstruction`, `contents`, positive `generationConfig.maxOutputTokens` | first candidate's ordered `content.parts`, `modelVersion`, `usageMetadata` |

All requests are non-streaming. Messages retains Chapter 1's headers.
Chat Completions uses a bearer authorization header. Gemini uses
`x-goog-api-key`, keeping the credential out of the URL. Strip a leading
`models/` from a discovered Gemini model when forming its path. The
request builder must not append that prefix twice.

Enduring instruction text goes to Messages `system`, Gemini
`systemInstruction`, and a Chat Completions system message. Join multiple
instruction text parts with a newline in their recorded order. Ephemera
are transient user content at the tail of the next request. Never move
them into the stable instruction prefix or retain them as dialogue.

Messages maps human/tool entries to user blocks and agent entries to
assistant blocks. Merge adjacent entries with the same wire role while
preserving all blocks and their order. A call becomes `tool_use` with
`id`, `name`, and `input`; a result becomes `tool_result` with
`tool_use_id`, content, and its error flag.

Inside a tool result, a URI blob renders as the text
`[<mime>] <locator>` on every adapter. It describes the referenced result;
it does not attach or fetch that file. A redacted child renders its stub,
followed by a space and its locator when one survives. Messages may retain
separate text blocks; the other two adapters join result children with
newlines. This rule differs deliberately from a human attachment, where
Gemini's `fileData` asks the provider to read the referenced content.

Chat Completions keeps human messages as user, assistant text/calls as
assistant, and each result as a tool message with `tool_call_id`. Calls
use `type:function` and `function:{name,arguments}`; `arguments` is a JSON
string on this wire. Render result text in order, joining distinct result
parts with newlines, and represent a redacted part by its stub and retained
locator. The neutral structure never acquires a special actor merely
because this API has a tool role.

This surface groups assistant calls separately from its text field. Keep
the original part order in the log even when a wire format cannot express
all of that interleaving. Do not rewrite the history to match the least
expressive adapter.

Gemini maps human/tool entries to user and agent entries to model, merging
adjacent equal wire roles. Calls become `functionCall`; results become
`functionResponse` with the original function name and a response object
containing ordered rendered result text under `result`. Preserve a call's
supplied ID where the surface accepts it. Place call-bound `thoughtSignature`
on the same part as its functionCall. A URI blob becomes
`fileData:{mimeType,fileUri}`. A redacted reference is rendered as text
with its locator; rendering must not fetch the content that was removed.

The public configuration also accepts tool declarations: name, description,
and an object-valued JSON input schema. Render these to Messages `tools`,
Chat Completions function tools, and Gemini `functionDeclarations`.
Declarations let a real model produce a call for this chapter's parsing
demonstration. They authorize no local execution. There is no tool registry
or dispatch loop yet.

Use the first choice/candidate and require one to exist. A present empty
text string is valid. Chat Completions `content:null` with valid calls is
valid; null without any recognized part is not an answer. The same test
applies to an empty Messages content array or absent Gemini parts. Unknown
replay material is retained as opaque where the adapter supports it, but
opaque material alone does not fabricate a visible answer or call.
Malformed arguments, wrong field types, or missing required usage fail
before committing response state.

Rendering must produce byte-identical JSON for fixed context, configuration,
and adapter version. Go's JSON encoder sorts map keys; it cannot repair
an array assembled by iterating a map. Preserve order when constructing
messages, content parts, declarations, and results.

## 2.6 Count each token once

Cache counters make addition surprisingly easy to get wrong. If a response
reports a prompt total of 150 and says 30 of those tokens came from cache,
adding both numbers reports 180 tokens for a 150-token prompt. Another
surface reports ordinary input and cache reads separately, where addition
is exactly what is needed. Each parser must know which statement its API
is making before it produces a common total.

The normalized usage object always has nonnegative integer fields `input`,
`cache_write`, `cache_read`, and `output`. These categories are disjoint.
Missing required base counts are errors; absent optional cache/thinking
detail fields mean zero. Reject negative or fractional counts and a
subtraction that would make ordinary input negative.

| Surface | `input` | `cache_write` | `cache_read` | `output` |
|---|---|---|---|---|
| Messages | `input_tokens` | `cache_creation_input_tokens` | `cache_read_input_tokens` | `output_tokens` |
| Chat Completions | `prompt_tokens - cached_tokens - cache_write_tokens` | prompt detail `cache_write_tokens` | prompt detail `cached_tokens` | `completion_tokens` |
| generateContent | `promptTokenCount - cachedContentTokenCount` | 0 | `cachedContentTokenCount` | `candidatesTokenCount + thoughtsTokenCount` |

The Chat Completions cache fields live inside `prompt_tokens_details`.
Do not add reasoning detail again to its completion total. Gemini's
thought count is separate from candidate output. These formulas follow
the [Messages usage convention](https://platform.claude.com/docs/en/build-with-claude/prompt-caching),
[Chat Completions schema](https://developers.openai.com/api/reference/resources/chat/subresources/completions/methods/create),
and [generateContent schema](https://ai.google.dev/api/generate-content),
checked October 7, 2026. Raw observations remain available if a future
surface changes its reporting convention.

For a deterministic fixture, ordinary input 100, cache write 20, cache
read 30, and output 40 must normalize to `100/20/30/40` on Messages and
Chat Completions.
Messages supplies those four fields directly. Chat Completions supplies
prompt 150 with cache details 20/30 and completion 40. Gemini cannot
represent a cache-write count on this surface: use its separate fixture
prompt 150, cached 30, candidates 25, thoughts 15, expecting `120/0/30/40`.
These are accounting fixtures, not claims about a model's tokenizer or
price. Add totals once at response completion, including a tool-only
response, and never while rendering or dumping.

## 2.7 User paths and the optional GUI

With no arguments, retain Chapter 1's JSON-lines CLI. Select
`LLM_VENDOR=anthropic|openai|gemini`, defaulting to anthropic. Nonempty
`LLM_MODEL`, `LLM_API_KEY`, and `LLM_BASE_URL` override the selected
provider's corresponding `ANTHROPIC_`, `OPENAI_`, or `GEMINI_` variables.
Default bases are `https://api.anthropic.com`, `https://api.openai.com`,
and `https://generativelanguage.googleapis.com`. Trim trailing slashes.
No default model is permitted; use provider discovery before a live run.
Keep a nonempty fixed system instruction unless the public caller supplies
its own. Missing live key/model, invalid input, and failed requests retain
Chapter 1's fail-fast behavior and safe diagnostics.

`CH02_LOG` selects the session log. If absent, use the executable's basename
plus `.log` in the working directory. Create a new log exclusively; refuse
to overwrite or append to an existing file in this chapter. Each Agent
has a distinct log destination. Choose explicit paths in a fresh directory
for demonstrations so a previous run cannot contaminate the next one.

| Input line | Output line and effect |
|---|---|
| `{"user":"..."}` | `{"assistant":"..."}` after exactly one HTTP request; concatenated text may now be empty |
| `{"ephemeral":"..."}` | `{"ack":"ephemeral"}` after recording nonempty one-request text; no HTTP |
| `{"redact":{"from":4,"to":4,"level":"redact_result","reason":"compaction"}}` | `{"ack":"redact"}` after recording/applying the directive; no HTTP |
| clean EOF | `{"usage":{"input":I,"cache_write":W,"cache_read":R,"output":O}}`, exit 0 |

Exactly one directive belongs in a line. A malformed directive fails
without an acknowledgement. A redaction span names positive sequence
numbers already in this log, has `from <= to`, and must include at least
one tool-result event. `redact_result` is the only implemented level;
reject another value rather than guessing a policy.

`dump` loads `CH02_LOG` (or the default path), validates it, and writes its
header and events as JSON lines. It does not create the file, contact a
provider, consume ephemera, or require a key/model. `render LOG` validates
and reduces the specified log, then writes exactly one JSON request body
for the selected vendor/model. It requires a model, but no credential,
and performs no HTTP. Errors produce empty stdout and nonzero exit.
Neither command emits the session's final usage record.

The public library exposes equivalent operations: create/select an Agent,
submit a prompt or directive, load and inspect a log, render offline,
provide tool declarations, obtain typed response parts and usage, and
register/unregister an observer. Public requests identify the Agent.
Observations include Agent ID, event sequence, kind, and an owned snapshot
of the event. A request returns its own answer or error. Do not require a
caller to search an unrelated broadcast stream to find completion.

Ensemble routes both clients. The CLI imports only the public root package.
Create a GUI stub in a separate module, for example `gui/go.mod`, which
depends on that public package and exposes a client that can submit a
request and observe its Agent's events. Its owner/interface connection
must follow the same rules as other components. Register an observer on
construction and unregister it on close; prove closed clients receive no
later events. Publish events in log order after application. This chapter
may dispatch synchronously and need not introduce an actor or event queue.

The GUI module documents where a WebSocket transport will attach and
states that it has no browser transport yet. Its fake-backed public
integration test must show a submitted prompt, its returned result, and
observed events for the right Agent. An external headless program must
build without requiring or importing the GUI module. Neither the core
module nor its public types may depend on WebSocket. A real browser
roundtrip is required only if a working browser transport is actually
implemented and advertised.

## 2.8 Literal replay fixtures

Save this as `history.log`. Its model is a fictional offline fixture ID.
It contains no valid credential or provider signature.

```jsonl
{"log_version":1}
{"seq":1,"type":"message_received","time":"2026-01-01T00:00:00Z","message":{"actor":"human","purpose":"dialogue","parts":[{"type":"text","text":"read config.json"}]}}
{"seq":2,"type":"response_ended","time":"2026-01-01T00:00:01Z","response":{"from":{"vendor":"anthropic","model":"fixture-messages","surface":"messages"},"parts":[{"type":"text","text":"Reading it."},{"type":"tool_call","call_id":"call-config","from":{"vendor":"anthropic","model":"fixture-messages","surface":"messages"},"name":"read_file","args":{"path":"config.json","limit":40}}],"usage":{"input":100,"cache_write":0,"cache_read":50,"output":20}}}
{"seq":3,"type":"tool_called","time":"2026-01-01T00:00:02Z","tool":{"call_id":"call-config","name":"read_file","args":{"path":"config.json","limit":40}}}
{"seq":4,"type":"tool_returned","time":"2026-01-01T00:00:03Z","tool":{"call_id":"call-config","parts":[{"type":"text","text":"port=8080\nhost=localhost"}]}}
{"seq":5,"type":"message_received","time":"2026-01-01T00:00:04Z","message":{"actor":"human","purpose":"dialogue","parts":[{"type":"text","text":"now check the logs instead"}]}}
```

This text-only fixture must render on all three adapters. To test result
references independently, append the following child to sequence 4's
`tool.parts` array in a separate copy:

```json
{"type":"blob","mime":"text/plain","ref":{"kind":2,"locator":"https://example.invalid/artifacts/config"}}
```

Its URI becomes descriptive result text under the rule in §2.5. A separate
human-attachment fixture tests actual `fileData` rendering:

```jsonl
{"log_version":1}
{"seq":1,"type":"message_received","time":"2026-01-01T00:00:00Z","message":{"actor":"human","purpose":"dialogue","parts":[{"type":"text","text":"describe this image"},{"type":"blob","mime":"image/png","ref":{"kind":2,"locator":"https://example.invalid/files/image"}}]}}
```

The Gemini request contains `fileData` with that exact MIME and URI.
Messages and Chat Completions reject this human-attachment shape in the
chapter's implemented subset. Dump variants using kind 1 with
`/tmp/example.png` and kind 3 with `handle:image` too. Dumping preserves
the references without trying to render or resolve them.

For result redaction, append the following line to a copy of `history.log`
or its tool-result-reference variant. The human-attachment fixture has no
result to redact.

```jsonl
{"seq":6,"type":"redacted","time":"2026-01-01T00:00:05Z","redact":{"from":4,"to":4,"level":"redact_result","reason":"compaction"}}
```

Replace each result child part at the targeted event with
`{"type":"redacted","stub":"[redacted]"}`, retaining `ref` on a stub
that replaces a blob. Preserve the result's call ID, error flag, and
the corresponding call. Text `port=8080` must appear in the unredacted
text render and disappear after redaction; the original log line remains
unchanged. The reference-bearing stub keeps its URI in both reduced
context and rendered stub text. Redaction is a view transformation, not
secure deletion from a file that still records the original content.

Use a separate fixture for call-bound material:

```jsonl
{"log_version":1}
{"seq":1,"type":"message_received","time":"2026-01-01T00:00:00Z","message":{"actor":"human","purpose":"dialogue","parts":[{"type":"text","text":"read deploy.sh"}]}}
{"seq":2,"type":"response_ended","time":"2026-01-01T00:00:01Z","response":{"from":{"vendor":"gemini","model":"fixture-gemini","surface":"generate_content"},"parts":[{"type":"tool_call","call_id":"call-deploy","from":{"vendor":"gemini","model":"fixture-gemini","surface":"generate_content"},"name":"read_file","args":{"path":"deploy.sh"},"opaque":"fixture-signature"}],"usage":{"input":40,"cache_write":0,"cache_read":0,"output":12}}}
{"seq":3,"type":"tool_returned","time":"2026-01-01T00:00:02Z","tool":{"call_id":"call-deploy","parts":[{"type":"text","text":"exec ./serve"}]}}
```

With vendor gemini and model `fixture-gemini`, the rendered call must
include `thoughtSignature:"fixture-signature"`. With another model it
must fail compatibility validation. Add a standalone opaque part to a
separate fixture to test matching replay and foreign omission without
the call-bound error obscuring that check.

For the text-bound case, replace the call in a copy of that fixture with
this part and remove the tool-result event:

```json
{"type":"text","text":"Ready.","from":{"vendor":"gemini","model":"fixture-gemini","surface":"generate_content"},"opaque":"fixture-text-signature"}
```

The matching Gemini render contains one `text:"Ready."` part carrying
that signature. A foreign render retains `Ready.` once and omits the
signature. The log retains both facts in either case.

An earlier audit found that deleting call-bound replay still scored 100:
the fixture had no bound material to replay and named a different model.
Commit `5a7dfca` records the corrected test. A positive control needs
both the actual material and the matching provenance. The absence case
cannot prove that a renderer knows how to include anything.

## 2.9 What the checks establish

The inherited score totals 100: parity 10, log dump 5, replay 10,
redaction 10, ephemera 10, usage 10, seam rendering 15, seam parsing 15,
and five reference checks worth 3, 2, 2, 4, 4. The session protocol is a
separate gate. Keep those regressions, then protect the new contract with
these checks. A check applies to public behavior unless it explicitly
examines ownership or imports.

| Property | Required evidence |
|---|---|
| Validated durable data | Independent malformed version, sequence, payload, part, and reference records fail at load; valid fixture dumps without HTTP |
| Ownership | Mutating caller buffers or returned observations cannot change captured history; Agent/Engine/Ensemble retain their Chapter 1 owners and one parent chain |
| Replay | Incremental application and fresh replay produce equal context and per-provenance usage; replay never adds a second copy of totals |
| Failure state | Failed attempt remains in log, unmatched user input stays out of the next request, consumed ephemera remain consumed |
| Request purity | Two renders are byte-identical and leave the log, ephemera, and accounting unchanged |
| Identity | Requested and returned model names survive separately; model switching preserves each response's usage; opaque positive and negative controls are non-vacuous |
| Parts | Empty text, ordered text/calls, tool-only success, malformed argument strings, and deterministic missing-call-ID synthesis are independently exercised |
| Redaction | Plain fixture renders its content; redacted fixture omits it, keeps pairing/reference, and does not grow an auxiliary redaction index |
| References | All three kinds survive dump; old path and zero kind fail there; URI mapping and redacted locator preservation have separate controls |
| Clients | Two independent Agents, public external consumer, optional GUI-stub module, Agent-attributed ordered observations, unregister-on-close, headless build |
| Credentials | Offline paths need no key; logs, diagnostics, request bodies, and observer events contain no key; local errors retain safe timeout/cancellation causes |
| Real use | All three CLI backends plus public consumer exercise implemented features; local faults and GUI-stub tests are labeled separately |

After the first student implementation and run, an independent reviewer
compares the answer with the first-edition standard at this feature scope.
Improve awkward design, duplicated work, misleading comments, and missing
teaching; a perfect score does not complete that review. Preserve the
initial attempt so the revised chapter can be judged on what it taught.

## 2.10 Taking it for a spin

Run each CLI backend in a fresh directory with an explicit log path.
Ask a model to choose a short code name, then recall and transform its
answer. Insert an ephemeral fact before one request and show from the
recorded requests/replay that it is absent from the next. Dump and render
the resulting log twice, without credentials, and compare bytes. Keep
the real response and measured usage with the selected and returned model
identities.

The external executable consumer exercises features the simple CLI does
not expose. It creates fresh Agents with fresh log writers, supplies a
tool declaration, and asks a real model for that call. It ingests a
controlled result through the public append path using the actual call ID;
the result is supplied by the demonstration, with no claimed tool execution.
It records a result-redaction directive and submits the next real prompt.
Retain the plain and redacted render evidence showing content removal,
preserved pairing, and the reference-bearing stub. This uses a live Agent's
normal write path; loading a prior log remains an offline inspection
operation and does not require writable log resumption in this chapter.

Use the same public consumer to demonstrate independent Agents, observer
attribution and unsubscription, and per-provenance accounting across a
configuration change. Use the literal fixtures for unsupported references,
opaque compatibility, and malformed-input controls that a provider cannot
be expected to produce. Label every fixture and local fault. The optional
GUI stub's separate-module test proves its public boundary; it cannot
produce a browser screenshot or a live WebSocket claim.

[LIVE RECEIPTS PENDING: all three actual CLI API paths; recorded history
and offline replay; ephemeral delivery; result redaction; independent
Agents through an external executable; clearly labeled GUI-stub integration.
No real-model response or successful acceptance result has been invented.]
