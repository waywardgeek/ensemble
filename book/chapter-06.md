# Chapter 6: Streaming, or the Same Answer in Pieces

Ask the Chapter 5 Agent something difficult and the cursor stops. Behind the
HTTP connection, the provider may be producing text or tool arguments already.
Our client waits for the completed body before showing any of it.

This chapter removes that unnecessary wait. It does **not** turn half a tool
call into permission to execute it. The display can be provisional; the
conversation and effects cannot.

## TL;DR

**Streaming changes delivery, not the authoritative answer.**

- Common declares `Chunk`, `ParsedResponse`, and the Engine's parsing operation.
  Engine implements SSE framing and vendor translation. The Agent's existing
  mailbox remains the sole conversation owner.
- `Engine.ParseAsync(HttpResponseMessage, Action<Chunk>?, CancellationToken)`
  accepts either plain JSON or SSE. Both paths use the existing completed-body
  decoder for canonical parts, provenance, exact arguments, and usage.
- A chunk carries a **response-local part ID**, kind, and display fragments.
  The parser returns the matching IDs with the completed events. The actor
  never guesses final identity from array position.
- Deltas enter the mailbox, then the bounded observation queues. They are not
  recorded in History. Final events are appended **before** their parts are
  published to observers.
- A declared model streams by default unless `Configuration.DisableStreaming`
  is true. Unknown models use plain delivery. Plain delivery publishes finals
  but no synthetic deltas.
- Partial arguments are display-only. Missing terminators, malformed frames,
  provider errors, and interrupted requests cannot authorize tool effects.
- The root optionally writes sanitized request/frame payloads to a caller-owned
  trace writer. Engine reaches it through Engine → Agent → Ensemble.
- `Examples/Observer` measures the delivered stream. `Examples/Workflow` tags
  real chunks with their Agent, response ID, and part ID. Both are separate
  assemblies using the public library.

The runnable endpoint is `solutions/ch06`. Working development uses
`agent`. These C# sources compile with C# 14 and .NET 10.

---

## 6.1 The idea in plain words

An HTTP response body is already a stream of bytes. Reading the whole body
before parsing is a client decision, not an inherent property of HTTP.

Three responsibilities change:

1. **Transport** hands an open response to the parser after headers arrive.
2. **Parsing** recognizes complete SSE frames and emits transient chunks as
   their data arrives.
3. **The actor** receives those chunks through the same mailbox as controls
   and completions, then publishes optional progress.

One responsibility deliberately does not change: **History owns finalized
conversation facts**. Replaying a saved response does not require stitching
hundreds of display fragments back together. The recorded `ResponseEnded`
event contains the completed parts and usage, just as it did in Chapter 5.

The HTTP worker cannot append History, start tools, or finish the user's turn.
It sends chunks and then an immutable completion back to the actor. Pause,
hint, interrupt, and shutdown therefore continue to use the existing owner
rather than a parallel streaming control loop.

**Upgrading from Edition 2:** the old “non-streaming is streaming with a length
of one” rule no longer applies to observer notifications. A non-streaming
response emits **zero deltas**, not one synthetic delta per part. Final parts
are authoritative in both modes: a display must be able to create a part from
its final notification even if no delta preceded it, and must not append the
final text to an already displayed partial copy. Streaming changes how soon
content can be shown, not what belongs in the conversation. Compare final
content and usage across modes; do not require a positive plain-mode delta count.

A slow observer still cannot stall that owner. Observation queues remain
bounded and lossy. A reliable request task, not a quiet screen or a final-looking
chunk, tells the application whether its request completed.

---

## 6.2 The shared vocabulary

Read `Common/Streaming.cs` and `Common/Actor.cs` before their implementations.
The essential declarations are:

```csharp
// Value-like parser output, not an independently owned runtime service.
public enum ChunkKind { Reply, Thinking, ToolCall }

public sealed record Chunk(
    int PartID,
    ChunkKind Kind,
    string Text = "",
    string ToolName = "",
    string ArgsText = "");

public sealed record ParsedResponse(
    ImmutableArray<Event> Events,
    ImmutableArray<int> PartIDs,
    bool ToolBeforeText = false);
```

The Engine declares its required partial operation in Common:

```csharp
public partial Task<ParsedResponse> ParseAsync(
    HttpResponseMessage response,
    Action<Chunk>? onChunk = null,
    CancellationToken cancellation = default);
```

This callback carries transient parser output for this invocation. It is not
a constructor service bag, a second owner, or an alternative route to logging.
The actor's HTTP worker uses it only to post a delta message to the mailbox.
Final events and identities return together in `ParsedResponse`.

`Observation` carries the resulting `Delta`, or a finalized `Part` with its
`PartID`, plus `ResponseID`. At the root, `AgentObservation` also retains
the actual originating Agent. A string label in JSON is a display of that
relationship, not a substitute for it.

The accumulation state is per parse and value-like. Its declaration lives in
Common; `Engine/Streaming.cs` supplies the algorithms. It is not another
persistent transcript. Once the complete native response is decoded, those
scratch dictionaries and JSON objects can be reclaimed normally.

---

## 6.3 One transport entry, shared final construction

The trap is two independent semantic decoders: one for JSON and one for SSE.
They inevitably acquire different rules for arguments, opaque signatures,
usage, and errors.

Our single HTTP entry is `ParseAsync`. It inspects the response content type:

- **Plain JSON:** read the completed body and call the existing `Parse(status,
  body)` decoder. Derive final identities and explanation order, but emit no
  transient chunks.
- **SSE:** frame and accumulate the provider's native response, emitting chunks
  along the way. At terminal success, pass that native response through the
  **same** completed-body decoder and return its events with the accumulated
  identity mapping.

Retaining the completed-body decoder is intentional. It is the shared semantic
operation, not a competing transport implementation or a new `ParseStream`
service. Canonical JSON numbers, provenance, signatures, and nullable usage
still have one decoding home.

A plain response could manufacture one delta per part. This C# implementation
does not: an event called a delta should mean something arrived incrementally.
Consumers already need final parts to correct lossy previews, so final-only
plain delivery does not require a second conversation model. A plain run has
zero deltas and the same authoritative parts.

Do not infer success from partial visible text. The parser requires terminal
success and complete arguments. Expected protocol failures become a fixed,
credential-safe `ErrorOccurred`, not a partial `ResponseEnded`. Explicit
cancellation propagates to the actor. Previously displayed chunks may remain
on a terminal, but cannot enter History as a successful response or cause a
tool to run.

---

## 6.4 Server-sent events, and the reader that survives them

SSE is ordinary HTTP with text framing:

```text
event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" world"}}

```

A blank line ends a frame. Each `data:` line contributes to its payload.
`StreamReader.ReadLineAsync` handles the line endings; the small shared loop
handles the framing before vendor interpretation begins.

| Case | Required behavior |
|---|---|
| Multiple `data:` lines | Join with newlines; do not silently discard later lines. |
| Comment beginning `:` | Ignore it, including keepalive comments. |
| CRLF | Recognize the same boundaries as LF. |
| `[DONE]` | Treat it as OpenAI's terminator, not a JSON object. |
| Last frame without a blank line | Process its pending data at EOF. |
| EOF without required protocol completion | Refuse success, even if text was visible. |
| Missing/null required members | Return a safe protocol error, not an accidental null-reference fault. |

A frame boundary is not a part boundary. One part may span many frames; one
frame may contain several parts. Shared framing does not make the vendors'
content semantics identical.

Tests use real `HttpResponseMessage` bodies and a held loopback HTTP response.
They include comments, multiline data, CRLF, missing terminators, invalid JSON,
missing members, and incomplete arguments. No provider account is needed to
establish these deterministic rules.

---

## 6.5 Part IDs: who is allowed to name a part

> Within one response, a part ID identifies exactly one part. A part never
> changes IDs, and one ID never crosses kinds.

A fresh ID for every chunk tells a consumer that forty things happened, not
which chunks belong to the same paragraph. A final array position is no
better: it may not exist when the first chunk arrives.

OpenAI makes this concrete. A tool fragment can arrive before the text that
will precede it in the final decoded response. Sorting final parts and then
numbering them cannot recover the identity used earlier.

**The code that chose an ID must repeat it on finalization.**

- Anthropic keeps its nonnegative block indices, including gaps.
- OpenAI maps wire tool indices to parser-local IDs and allocates text/thinking
  IDs on first appearance. Final decoder order may differ from arrival order;
  `PartIDs` explicitly maps the finished parts back to their original IDs.
- Gemini infers continuation from content kind, explicit part boundaries, and
  signature closure. It has no Anthropic-style block index to borrow.

Zero is valid. IDs may restart on the next response. The display key is
`(response_id, part_id)`, or `(agent, response_id, part_id)` for several Agents.
The actor uses the recorded `RequestSent` sequence as `ResponseID`; it does not
ask the CLI to infer response boundaries from a state transition it might miss.

Part IDs themselves remain transient parser identities, not a new persisted
widget schema. Later replay or reconnect code must not assume a fresh parser
run will reproduce the same display mapping without an explicit contract.

The actor first appends each final event, then publishes the corresponding
`PartFinal` observations using those stored parts and parser IDs. An observer
that reads History on finalization therefore sees the event already present.
Human input and Job results have no parser part ID and must not masquerade as
model finals.

Tests group deltas by identity before comparing them with finals. Concatenating
all text in a turn would let meaningless IDs pass. Tool fragments need a
slightly different assertion: decode the completed argument JSON and compare
values, because canonical formatting may differ from the incoming fragments.
The large integer `9007199254740993` must not become a rounded floating-point
value on the way through that comparison.

---

## 6.6 Three vendors, three local translations

| Vendor | Request selection | Stream interpretation |
|---|---|---|
| Anthropic | `stream: true` | Message start, block start/delta/stop, usage updates, message stop. |
| OpenAI chat completions | `stream: true`, `stream_options: {include_usage: true}` | Choice deltas, tool-index mapping, finish reason, optional usage-only frame, `[DONE]`. |
| Gemini | `:streamGenerateContent?alt=sse` | Response-shaped frames; infer continuation locally and retain signatures. |

### Anthropic

Block start creates identity. Text and thinking deltas append visible content;
input-JSON deltas append raw argument text. Block stop decodes the complete
argument value. A stream cannot succeed while a block remains open. Signatures
and opaque blocks survive into the shared final decoder.

### OpenAI

Only choice zero belongs to this single-answer API. Tool indices identify
calls even when several calls are interleaved. Names can also arrive in pieces;
show them, but do not dispatch by a partially received name.

Usage can arrive after the finish reason in a frame with no choices. Allow
that accounting frame, but reject new choice content after completion.
Requesting `include_usage` does not make missing counts measured zeros: the
existing nullable usage rules still apply.

### Gemini

A first same-kind entry can continue the current part. Additional explicit
entries start distinct parts; function calls are complete parts; a signature
closes the part it signs. Empty signed text is meaningful replay material and
must not vanish merely because it paints no characters.

Readable thought content and ordinary text remain different kinds. Preserve
signatures on thought, text, and tool-call parts according to the existing
provenance codec. Do not invent a uniform wrapper shape just for a streaming
test; replay semantics, not a convenient fixture assertion, govern the value.

### Arrival order is policy evidence

The strict explanation policy asks whether substantive assistant text preceded
the first tool fragment. Later text cannot retroactively explain an earlier
call. `ToolBeforeText` records this decision during parsing, before final parts
are rearranged by the shared decoder.

The actor uses that fact together with its existing explanation check. A
refusal still produces a paired result for the complete call and no effect.
Neither an explanation nor a complete-looking argument fragment permits
execution before terminal success.

---

## 6.7 Capabilities: conservative delivery, strict content

`ModelFeatures.Stream` is a Boolean permission to request streaming. It does
not promise incremental text, thought, and arguments from every model. The
adapters emit only what actually arrived: a complete Gemini function object
may produce one display chunk while text arrives in many pieces.

This makes a per-kind capability bitmask unnecessary for this implementation.
A future requirement to negotiate individual stream kinds could justify one;
the current HTTP APIs and display consumers do not need it.

`Configuration.DisableStreaming` defaults to false. A caller can opt out
without changing the selected model. Effective selection requires both a known
stream-capable row and no explicit opt-out. Stream capability is declared
separately from media capability; adding an image type must not silently change
HTTP delivery.

The small offline-rendering table includes `claude-3-haiku-20240307`, `gpt-4o`,
`gpt-4o-mini`, and `gemini-2.5-flash`. These declarations are not current model
availability checks or proof of real-provider acceptance.

An unknown model falls back to plain delivery. An unsupported media attachment
still fails before HTTP. That difference is deliberate: choosing slower
delivery preserves content; silently dropping an image would corrupt it.
A new wire format needs parser work and evidence, not merely another table row.

---

## 6.8 Three records, three purposes

**History** retains complete events. Deltas never become extra dialogue or
extra token accounting. Successful responses preserve the established replay
contract; interrupted or malformed attempts retain their actual terminal facts.

**API trace** retains sanitized payloads for diagnosis. The root accepts an
optional `TextWriter` and serializes writes. The caller owns that writer's
lifetime and closes it after the root has stopped its children. Engine reaches
it through its real parents; no copied logger is injected into the parser.

Requests and received frame payloads are traced as they arrive. The configured
key and its JSON-escaped representation are redacted before writing. Headers
and endpoint URLs are not traced. This is not a byte-for-byte HTTP capture:
SSE comments and framing are not response payloads. It is also not a general
secret detector—prompts, file content, opaque signatures, and other private
material may remain. Keep traces outside version control and review them before
sharing.

**Observers** receive provisional chunks and stored-part final notifications.
They can update a display, but neither progress loss nor a display exception
changes the reliable request reply. Counts from a bounded observer describe
received observations, not an authoritative accounting ledger.

---

## 6.9 The terminal, and the flush that makes streaming real

The JSON CLI emits one flushed record per delivered observation:

```json
{"observation":"part_delta","response_id":2,"part_id":0,"kind":"text","chunk":"Hello"}
```

A model `part_final` includes the same identity and a typed `part`. State and
turn observations remain distinct. Human input uses `input_part`; Job results
use `tool_result`. A successful posted turn produces a completed `assistant`
record; a blocking request already has its own reliable answer and must not
print it twice.

Use `kind: prompt` for live JSON output. Legacy `user` requests retain their
reliable reply-only compatibility behavior. `-no-stream` or
`EN_DISABLE_STREAMING=1` forces plain delivery. `CH06_API_LOG` selects the trace;
`CH02_LOG` retains its earlier name for the canonical CLI event log.

In interactive `chat`, streamed text, thinking, and tool fragments are a
**stderr preview**. Thinking is dim and tool arguments yellow only when stderr
is a terminal; redirected output has no ANSI codes. The reliable completed
answer is written once to stdout. Prompts and usage go to stderr too.

This differs deliberately from the older Go terminal's streamed-answer stdout.
A lossy preview cannot guarantee a complete redirected answer. Rather than add
an unbounded queue or a second reliable streaming protocol, this example keeps
stdout reliable and makes the preview explicitly provisional. Both file
descriptors are visible in an ordinary terminal. The GUI chapter can replace a
preview from authoritative finalized content instead of merely appending it.

Each preview or JSON record is flushed immediately. Otherwise every stage may
be logically correct while the user still sees nothing until EOF. The gated
exercise holds the provider's remaining frames until the application's first
meaningful chunk appears. That proves visibility **before completion**, unlike
counting stdout reads after the provider has already finished.

Hint acknowledgments are separate records, not strings recognized by a
`hint:` prefix. A model may legitimately begin its answer with that text.
Display content is never a protocol discriminator.

On EOF, the JSON CLI first flushes accepted turns, then stops work, closes the
Agent, and drains queued display records. It captures usage before destruction.
Cancelling the display reader immediately after the reliable answer would lose
already-published finals and create a misleadingly incomplete transcript.

---

## 6.10 Two separate applications

### Aggregate observer

`Examples/Observer` builds one Agent, installs its tools, starts one observation
reader, and awaits `SendAsync`. Local counters measure:

- chunks and UTF-8 display bytes per kind;
- finalized model parts, keyed by response and part ID;
- state transitions and model-request steps;
- time to first delivered delta and total reliable-turn duration.

After completion it destroys the root and drains the observation task, then
prints one `stream_stats` record. The report also includes the reliable answer,
measured History usage, and effective streaming selection. Plain mode reports
zero deltas, `has_delta: false`, and `first_delta_ms: 0`; that zero means no
delta was delivered, not zero-latency network service.

`CH07_NO_STREAM=1` disables streaming in this example. `CH07_LOG` selects the
aggregate JSONL file; `CH07_AGENT_LOG`, when supplied, selects its event log.
The historical environment names remain for compatibility. Both example
applications require an explicitly configured loopback endpoint in this
permitted offline exercise.

### Three-role workflow

`Examples/Workflow` retains Chapter 5's one manuscript and real named Agents.
Reliable author, editor, and reviewer tasks drive the pipeline. Its root
observer only displays activity; an Idle transition is not permission to
advance the workflow.

Deltas come directly from `Observation.Delta`; finals use the parser's IDs.
The example no longer invents a whole-part delta or increments a local counter
on finalization. Every output record identifies the actual Agent. A
`hint_accepted` record is not a fabricated model chunk. `CH06_LOG` duplicates
output records to a file, and `CH06_HISTORY_DIR` optionally stores each Agent's
canonical events. `EN_DISABLE_STREAMING=1` selects plain delivery.

The local exercise holds the author's first streamed tool-name chunk until
that chunk is displayed. It then sends a busy hint during a real Job, verifies
the untouched batch tail is paired, and lets the reliable three-role pipeline
finish at EOF. Finals and assembled text/tool arguments are checked by
`(agent, response_id, part_id)`, including noncontiguous IDs.

---

## Taking it for a spin

Build from the endpoint directory with the approved SDK and an owned CLI home:

```sh
dotnet build Cli/Ch04.csproj
dotnet build Examples/Observer/Observer.csproj
dotnet build Examples/Workflow/Workflow.csproj
dotnet run --project Tests
```

The CLI project retains the name `Ch04.csproj` to preserve existing executable
references; it compiles this endpoint's current library. Run from an owned
scratch directory so job files and traces stay outside the source tree.
Configure `LLM_VENDOR`, `LLM_MODEL`, `LLM_BASE_URL`, and `LLM_API_KEY` for the
owned fake. Never put a real credential in a command argument or transcript.

```sh
printf '%s\n' '{"kind":"prompt","text":"Describe the lighthouse."}' |
    dotnet /absolute/endpoint/Cli/bin/Debug/net10.0/ch04.dll

printf '%s\n' '{"kind":"prompt","text":"Describe the lighthouse."}' |
    dotnet /absolute/endpoint/Cli/bin/Debug/net10.0/ch04.dll -no-stream

dotnet /absolute/endpoint/Cli/bin/Debug/net10.0/ch04.dll chat

dotnet /absolute/endpoint/Examples/Observer/bin/Debug/net10.0/Observer.dll \
    'What does the configuration say?'
CH07_NO_STREAM=1 dotnet /absolute/endpoint/Examples/Observer/bin/Debug/net10.0/Observer.dll \
    'What does the configuration say?'
```

Use identical scripted responses for exact on/off comparisons. Two independent
live generations need not choose the same words or chunk boundaries. Real
provider exercises require separate authorization; this endpoint's offline
work does not claim them completed.

### Scoped executable grading

The existing grader has language-independent streaming fixtures. The C# route
accepts a compiled CLI launcher and the separate observer launcher:

```sh
grade-edition2 -ch 6 -executable-only -json \
    -stream-model claude-3-haiku-20240307 \
    -stream-exercise /absolute/observer /absolute/cli
```

| Check | Points | Meaning |
|---|---:|---|
| `stream-deltas` | 20 | Genuine multiple chunks, not just completed parts. |
| `deltas-match-final` | 20 | Response-scoped text identity and reconstruction. |
| `thinking-streamed` | 15 | Thinking chunks remain separate from reply text. |
| `tool-params-streamed` | 15 | Tool names and argument fragments correlate with complete calls. |
| `delivery-not-content` | 10 | On/off finals and usage agree; unknown models fall back; external aggregate reports real measurements. |
| **Scoped total** | **80** | Excludes inherited Chapter 5 and Go consumer/topology checks. |

The full Go route remains intact. The scoped grade is not a claim of full Go
consumer parity. Source review, all-vendor parser fixtures, gated delivery,
interruption tests, and the three-role application are separate evidence.

The first C# scoped grade misclassified opaque finals because its adapter read
a generic observation `text` label before the typed `part`. Correcting that
adapter preserved the actual typed-part contract. A green number still cannot
replace inspection of parser IDs, ownership, cleanup, or real output timing.

---

## Lessons retained from earlier editions

Earlier demonstrations and mutation tests are historical, not C# integration
receipts. Their useful lessons remain:

- A no-stream request can return identical content while failing every meaningful
  incremental-delivery assertion. Content equality alone does not test streaming.
- A fresh ID per chunk can look plausible in logs. Grouping before comparing is
  what exposes the mistake.
- A test waiting on file modification time can mislabel a busy hint as a new
  turn. Reliable completion, not timestamps, defines the boundary.
- A working provider model outside an explicitly authorized selection still does
  not satisfy the exercise. Successful output does not expand permission.
- Chunk counts and OS read counts are not token counts. A few rapid reads may
  establish incremental writes without proving perceptibly paced typing.

---

## What this chapter does not build

**No execution from fragments.** Only a successful complete response can
supply a tool batch. Interrupted streams cannot dispatch a late call.

**No durable delta transcript.** History retains canonical facts, not a second
conversation assembled from previews.

**No lossless progress guarantee.** Reliable answers and canonical History
remain available when a bounded display drops optional observations.

**No polished redraw.** The terminal appends previews. Replacing a partial
widget with its authoritative final, handling reconnects, and giving several
Agents usable screen space belong to Chapter 7.

The vocabulary is ready for that next step: actual origins, response-local
identity, typed chunks, and stored-before-published finals. The GUI will project
those facts; it must not become another owner of the conversation.
