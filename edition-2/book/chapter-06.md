# Chapter 6: Streaming, or the Same Answer in Pieces

Ask the Chapter 5 agent something hard and the cursor stops. Twenty seconds,
or ninety. Behind the curtain a model is reasoning, writing, deciding to call
a tool, spelling out arguments one token at a time. You get none of it, then
all of it at once.

Every vendor has been willing to send that along as it is produced for years.
The agent never asked. Chapter 5 built an observer seam with `PartDelta`
observations, but a whole response still arrived before its length-one deltas.
This chapter makes those observations arrive while the response is being read.

The result is not beautiful. Reasoning, reply, and half-finished tool
arguments land on one terminal in three colours with no layout. The mess is
the point: once you can watch the agent think, you will want somewhere better
to watch it. Chapter 7 builds that. This chapter earns it.

---

## TL;DR

Streaming changes how a response is **delivered**, not what it **is**.
Everything below follows from that one sentence.

### The seam

The first-edition incidents and measurements below remain historical. The
interfaces shown here follow the accepted Ensemble implementation.

`Parse` changes signature. It gains no sibling.

```go
// internal/common/stream.go
type Parser interface {
	Parse(resp *http.Response, cb StreamCallbacks) error
}
```

No `ParseStream`. Non-streaming is a stream of length one, so one method
covers both modes. `Render` still renders the context through its owning
Engine; configuration comes from the Agent parent. The accepted signature
is `Render(*Context) ([]byte, error)`. Transport builds the URL and headers
separately. Streaming changes those existing request decisions.

### The notifications

`StreamCallbacks` lives in `internal/common`, alongside `Parser`. These are
response notifications, not copied configuration, sibling services or substitutes
for parent interfaces. They run synchronously; their consumers queue slow work.
The parser remains attached to its Engine and can reach its Agent and Ensemble.

```go
// internal/common/stream.go
type StreamCallbacks struct {
	// One incremental chunk. partID identifies the PART, not the chunk.
	OnDelta func(partID uint64, kind DeltaKind, chunk string)

	// Each finalized event, in order, for the event log.
	OnEvent func(Event)

	// Each completed part, carrying THE SAME partID its deltas carried.
	OnPartFinal func(partID uint64, part Part)

	// Each frame payload: for SSE, one event's data bytes.
	OnFrame func(eventType string, data []byte)
}
```

| callback | wired by | to |
|---|---|---|
| `OnDelta` | Agent actor | mailbox message, then observer |
| `OnPartFinal` | Agent actor | parser ID retained until stored-part publication |
| `OnEvent` | Engine | finalized facts returned to the actor for recording |
| `OnFrame` | Engine | sanitized root-owned API trace |

A caller wanting request-scoped notifications can pass them to `AskWatching`;
it uses the same actor as Ask. The accepted CLI attaches an ordinary observer
and queues its display work. Neither path starts a second orchestration loop.

### The taxonomy

```go
type DeltaKind uint8

const (
	DeltaThinking DeltaKind = iota + 1 // wire: "thinking"
	DeltaText                          // wire: "text"
	DeltaToolCall                      // wire: "tool_call"
)
```

An enum, not a string, because every consumer switches on every case: the
terminal picks an ANSI colour, a GUI picks a CSS class. Constants start at
`iota + 1` so the zero value is invalid, and `MarshalJSON` refuses an unset
kind rather than emitting a plausible default. Serialized:

```json
{"observation":"part_delta","part_id":3,"kind":"text","chunk":"Hello"}
```

`PartDelta` retains an optional `agent` label for clients combining streams.
The actual Ensemble routing still supplies the source Agent in its observation
envelope. A display label does not replace that ownership relationship.

### The capability

```go
type Stream uint8

const (
	StreamText Stream = 1 << iota
	StreamThinking
	StreamToolArgs
)

const StreamAll = StreamText | StreamThinking | StreamToolArgs
```

A bitmask on `ModelFeatures`, not a bool, because vendors ship this capability
in pieces. `Config` gains one field, `DisableStreaming bool`. Effective
behaviour is the AND of caller and table:

```go
func StreamingFor(cfg Config) Stream {
	if cfg.DisableStreaming {
		return 0
	}
	features, ok := LookupModel(cfg.Model)
	if !ok {
		return 0
	}
	return features.Stream
}
```

An unknown model gets no streaming, and that is **not** an error, which is the
deliberate opposite of what the same table does for media. Section 6.7 is why.

### The reader

One shared reader takes an `io.Reader` and yields
`(eventType string, data []byte)` pairs. The original `internal/llm/sse.go` measured 111 lines with 14 tests.
Those are historical measurements, not target counts; the accepted implementation
keeps its smaller reader with translation in `internal/engine/sse.go`.

### The request

| vendor | asks for streaming with | parser's oddity |
|---|---|---|
| Anthropic | `"stream": true` in the body | dispatch on `content_block_delta`, then on `text_delta` / `thinking_delta` / `input_json_delta` |
| OpenAI | `"stream": true` plus `"stream_options": {"include_usage": true}` | without `include_usage`, final usage is absent; absence is not measured zero |
| Gemini | `:streamGenerateContent?alt=sse` endpoint | each frame is a whole response object with no block indices; continuation is inferred |

### What does not change

The event log. A `ResponseEnded` event still carries the complete response as
`ResponseData{Parts, From, Usage}`, the context still replays from finalized
events, and replay still equals live. Deltas are **ephemeral**: they reach
observers and are never recorded. Sanitized frame payloads go to the root-owned `api.log` trace as they arrive.

### The part id rule

> **Within one response**, a part id identifies exactly one part. Two parts
> never share an id, one part never uses two ids, and an id never crosses
> kinds.

The parser chooses IDs from vendor indices or its local continuation mapping.
They can restart with each response, so a turn containing a tool call can
go around the loop twice and reuse them. The key a
consumer needs is therefore (response, part id), not part id alone.

The parser supplies the id on every delta and reports the matching
final identity; the actor publishes the stored `PartFinal`. The caller never
reconstructs an ID from final position, because real wire formats can defeat
that guess. Section 6.4 has the receipt.

### The exercise

`ch07/main.go`: one agent, one observer, a report on the shape of the stream
including time to first delta. `CH07_NO_STREAM=1` runs the identical script
with streaming disabled.

```
go run ./cmd/grade-edition2 -ch 6 path/to/core
```

The local streaming fixture uses a declared model, by default the historical
`claude-sonnet-5` row shared by the references. `-stream-model MODEL` selects
another implemented streaming-capable Anthropic row; it is a grader input,
not provider discovery. A separate deliberately unlisted model tests plain
fallback without teaching a special capability for a fake.

`-stream-exercise DIR` selects the separately built shape module; otherwise
the grader finds `ch07` inside or beside the core directory. Earlier consumer,
workflow and media-model flags still apply to inherited parity. The core's
incremental JSON observations and the external application's aggregate report
serve different checks; neither substitutes for the other.


| check | points | what it tests |
|---|---|---|
| `stream-deltas` | 20 | Asking for the stream produces one: many deltas per turn, not one per part |
| `deltas-match-final` | 20 | Text deltas grouped **by response and part id** concatenate exactly to that part; at least one part arrived in several deltas |
| `thinking-streamed` | 15 | Reasoning arrives as `DeltaThinking`, correlated by part id, never leaking into the reply |
| `tool-params-streamed` | 15 | A tool call's name and arguments arrive as `DeltaToolCall`, correlated by part id |
| `delivery-not-content` | 10 | Streaming off: same final content and usage, different delta count; unknown-model fallback and the separate shape report remain usable |
| `ch5-parity` | 20 | Every Chapter 5 check still passes |
| **total** | **100** | |

---

## 6.1 The idea in plain words

An HTTP response is a stream of bytes. `io.ReadAll` is a choice to ignore
that, convenient enough that it stops looking like a choice.

Non-streaming does not mean the vendor produces the answer atomically. The
model emits tokens one at a time either way; the server holds them until the
last one is written. The latency a user feels is the cost of thinking plus a
decision to say nothing while it happens.

Three things change. The **HTTP client** hands the open response to the parser
instead of reading the body. The **parser** takes a live stream and calls back
per chunk. The **Agent actor** receives those chunks through its mailbox and
fires observations instead of buffering the whole response behind a blank terminal.
Each worker message carries its turn identity, so stale progress after an
interrupt cannot contaminate the new turn.

One thing pointedly does not change: the record. At end of turn the event log
holds what it held before, and rebuilding context from it yields the same
bytes. Recording deltas as additional conversation content would duplicate finalized
parts or force replay to assemble them. Keeping display fragments out of the
durable log preserves the existing replay contract.

Hence the sentence that makes the seam small. **Non-streaming is streaming
with a length of one.** A response in one piece is the same kind of thing as a
response in forty pieces. No code downstream of the observer asks whether
something was streamed; it receives deltas and the only question is how many.
Which buys one `Parse` instead of two, one path per vendor instead of two, and
a grader that runs the identical exercise both ways and demands identical
content byte for byte.

---

## 6.2 Server-sent events, and the reader that survives them

Streaming from an LLM vendor uses plain HTTP. The response body is framed in
server-sent events, a text format older than all of this.

```
event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" world"}}

```

Lines of `field: value`; a blank line ends an event. Those rules carry every
streamed token you have ever watched appear in a chat window.

Each vendor parser could scan for `data: ` itself in six lines, and those six
lines pass every happy-path test. Then:

| case | what a naive scanner does |
|---|---|
| multi-line `data:` field | keeps the first line only, truncating content with no error |
| comment line, starting `:` | treats a keepalive as an event with an empty payload to unmarshal |
| CRLF | fails to recognize the separator because the line still contains `\r` |
| `data: [DONE]` | tries to unmarshal a sentinel, erroring on a stream that completed |
| unterminated final event | drops the last chunk, which usually holds the stop reason |

None is hard. All share a shape: a framing imprecision that produces plausible
output and fails later, somewhere with no mention of SSE in the stack trace.

The first edition collected these cases in one reader with 14 tests.
Keep the framing function small; line and test counts are consequences. Vendor-specific work starts after framing is
resolved, with an event type and a slice of JSON in hand. The same "fakes
first" pattern from Chapter 3 in different clothes: the framing is
deterministic, ugly, and testable without touching the network or spending a
token.

---

## 6.3 One `Parse`, not two

The obvious design keeps `Parse(status int, body []byte)` and adds
`ParseStream`. Nothing existing breaks and each method is simpler than a
method doing both.

Three vendors times two methods is six implementations. Each vendor's pair
must agree exactly about what a tool call is, what a thinking block is, which
stop reasons map to which events, and how usage is counted. Nothing in the
type system enforces that, and nothing in the test suite does unless somebody
keeps writing a test that runs both and compares. Fixes land in the streaming
path because that is the path in production, and the other drifts until
somebody disables streaming to debug something and finds the framework behaves
differently in the mode they picked to have fewer variables.

One entry point keeps the modes together; shared final construction prevents
the semantic drift:

```go
Parse(resp *http.Response, cb StreamCallbacks) error
```

On SSE, `Parse` reads frames and calls `OnDelta` per chunk. On a plain JSON
body it reads the body whole and calls `OnDelta` once per part with that
part's entire content. Both routes then run the same semantic decoder to
build events and call `OnEvent` and `OnPartFinal`. The accepted stream readers
accumulate provider messages before using that decoder, preserving one place
for canonical arguments, provenance, opaque replay material and usage.

The non-streaming route could skip `OnDelta` entirely, since nothing is
incremental. It fires anyway, because "length one" has to be true in code and
not only in prose. Were it silent, every observer would need a second
rendering path fed from finalized events, and the seam would have bought
nothing except a longer interface.

The seam still has exactly two methods, `Render` and `Parse`, which is what
the vendor-independence claim has rested on since Chapter 2. Adding a third to
ship a feature would concede that the seam was shaped around the features that
existed when it was drawn, which is the thing a seam exists to deny.

---

## 6.4 Part ids: who is allowed to name a part

A stream delivers forty chunks belonging to a paragraph of reasoning, a reply,
and a tool call's arguments. A GUI appends each chunk into the right widget,
then replaces that widget's content with the finalized part. It needs to know
which chunks belong together. A part id is nothing more than that.

The first-edition brief originally minted a fresh id per chunk with an
`atomic.AddUint64`. Forty unique ids tell a consumer that forty things
happened and nothing about which were the same thing. The field would exist,
be populated, look reasonable in a log, and carry no information.

So ids are per part. The question is who assigns them.

The instinct is the caller, leaving the parser a pure translation layer, with
the obvious formula being the part's position in the finished response. That
formula is wrong on the wire, and OpenAI is the proof: it streams a tool
call's arguments **before** it is knowable whether a text part will occupy
index zero. When the first tool-argument chunk arrives the response is
unfinished, positions are undefined, and any id the caller computes is a guess
a later chunk can falsify.

> The code that **chose** an id is the only code that can be trusted to
> **repeat** it.

Anthropic can retain block indices. The accepted OpenAI parser reserves zero
for text and maps tool index i to i+1, avoiding a collision when a tool arrives
first. Gemini must infer continuation locally. Zero is a valid ID; insisting
that all IDs are positive tests a different contract.

The parser chose it, so the parser owns both ends: it supplies the id on every
`OnDelta` and reports `OnPartFinal` carrying the same id. Finals carry the
parser-chosen ID, not a position recomputed from the finished list. The id is thereby
freed from meaning "where", and only ever means "the same part as that other
one", which is the sole property a consumer can rely on.

These parser-local IDs have a consequence worth stating before
Chapter 7 trips over it: **the ids may restart with each response**. A turn that
calls a tool goes around the loop at least twice; if part id 0 is reused in
the second response, it names a different part from part id 0 in the first. Uniqueness holds
within a response, not within a turn. The grader for this chapter had to learn
the same lesson, and its harness now tracks response boundaries rather than
assuming ids are unique across a turn.

So a consumer keying widgets needs (response, part id). The current
`state_changed` transition to `in_flight` lets the grader separate live
responses. That transient boundary does not itself give a reconnecting GUI a
persistent response identity. Keep that distinction in view when Chapter 7
reaches for the seam.

`OnPartFinal` exists for this. It looks redundant beside `OnEvent`, since
finalized parts are inside the events, but the events do not carry the ids the
deltas used, and reconstructing that mapping from outside is the same guess
that fails on the wire.

Ordering matters too. Parse reports a completed event before its part finals.
The HTTP worker collects those facts and IDs; it does not mutate History. The
actor records the response, then publishes `PartFinal` using the stored parts
and the parser's IDs. A request watcher receives its final notification there
too. An observer re-rendering on finalization reads a log that already contains it. Reversed, every consumer gets an intermittent view of a
log one event behind.

`deltas-match-final` grades this the hard way. A check that concatenated every
text delta in a turn and compared against the final message would pass on a
submission whose ids were noise, including the fresh-id-per-chunk submission
this section rules out. So the check groups by id first, requires each group
to concatenate exactly to its own part, and requires at least one part to have
arrived in several deltas. Grouping by id is the real check; concatenation
afterward is arithmetic.

Text and readable thinking fragments reconstruct their corresponding content.
Tool arguments need one additional distinction: the transient JSON fragments
are displayed as received, while the complete tool part retains canonical
JSON. Whitespace can differ without losing an argument. Compare the decoded
complete argument value when testing that boundary, and never execute fragments.

---

## 6.5 Three vendors, three dialects

The framing is shared. The meaning is not, and the differences are listed in
the TL;DR table above. Two deserve a closer look.

OpenAI's `stream_options.include_usage` is not decoration. Omit it and a
streamed response contains no `usage` object at all, just an absent field, so
a careless accumulator can record zeros for every
streamed turn. The answers are correct, the tools run, the log is well formed,
and careless cost tracking reads zero forever. Missing accounting must still
be treated as missing, rather than converted into measured zero. A silent success that is wrong, one JSON
field deep.

Gemini sends response-shaped frames containing new content, rather than
Anthropic-style named delta events. Text can still arrive incrementally. There
are no equivalent block indices, so continuation is inferred from position
and the shape of what arrived before. Preserve the thought flag and trailing
signatures, including signatures on ordinary text; a signature-only frame
must not disappear just because it has no new visible words.

The seam does not make vendors identical. It makes their differences
**local**: all three oddities live inside one vendor's `Parse`, and none
escapes the translation implementation into the actor, observer, event log
or terminal. Locality is
the whole claim, and it is smaller than "vendors are interchangeable", which
was never true.

---

## 6.6 When a capability table pays for itself

Gemini's row has `StreamToolArgs` clear.

```go
"gemini-3.8-flash": {
	Media:  MediaImage | MediaAudio | MediaVideo | MediaDocument,
	Stream: StreamText | StreamThinking,
},
```

In the retained implementation and observations, function parameters arrive
as complete objects there. The parser handles that dialect locally; the actor
and terminal need no vendor-specific workaround. The table says the bit is off, `StreamingFor` returns a
mask without it, and the parser emits the complete tool object as one
display chunk. It does not manufacture fragments that the wire never sent. Behaviour degrades exactly as far as the
capability is missing and stops there.

The following are historical capability rows retained in this implementation,
not a current provider-availability or crossover-readiness claim:

| model family | text | thinking | tool arguments |
|---|---|---|---|
| `claude-opus-5`, `claude-sonnet-5` | yes | yes | yes |
| `gpt-6-astra`, `gpt-5.6-sol` | yes | no | yes |
| `gemini-3.8-flash`, `gemini-3.1-pro-preview` | yes | yes | no |

Three families, three combinations, no two the same shape. A single
`Streaming bool` would force two of those rows to choose between losing text
streaming and inventing tool-argument chunks that never arrived. The bitmask
is the minimum structure the observed facts require.

The table will go stale; vendors ship these bits on their own schedule. An
updated row can enable a capability the translator already supports. A new
wire shape still needs parser work and evidence; changing a bit cannot create
a decoder. The table gives the reader a clear place to start.

---

## 6.7 Guessing about delivery is not guessing about content

Given a model name it has never seen, `StreamingFor` shrugs: no error, no
refusal, a non-streaming request, turn proceeds. Asked whether the same
unknown model accepts an image, the same table produces a loud refusal, which
Chapter 5 spent real effort keeping. Two lookups, two opposite answers on a
miss. An inconsistency like that is either a bug or a principle.

> A guess about **content** can corrupt the conversation. A guess about
> **delivery** cannot.

Silently drop an image because a model cannot see and the request
may still succeed. The model answers confidently about a picture it never
received, and nothing downstream can recover the missing input. Fall back to plain delivery and
the same scripted response arrives, same text and parts, in fewer pieces.
The cost of that conservative choice is a slower first token, not a missing
image. Separate live generations can differ; deterministic parity compares
the same scripted content in both delivery modes.

The practical half matters as much. Model names are an **open set**: this
framework's own tests already use `gpt-5-2025-08-07`, `fake-model`, and
several `-fake` suffixes no honest table would claim to know. A framework that
refuses to send any request until its table has heard of your model is one
whose first patch in every deployment removes that check. Refusing to guess is
right; refusing to work is not. The line is drawn per capability, which is why
the two lookups disagree.

### The negative boolean

```go
DisableStreaming bool
```

`Stream bool` is the better name and the wrong field, because in Go the zero
value picks your default. `Stream bool` makes a hand-built `Config{}` mean
streaming **off**, the opposite of the intended default, recoverable only
inside a constructor other people are free not to call. In a framework whose
premise is that other people construct these structs, a default holding only
when someone uses the front door is not a default. When the zero value and the
good name point in opposite directions, the zero value wins: the compiler
enforces it and the name does not.

---

## 6.8 Three record-keepers

**The event log** is unchanged. Deltas fire at observers and are never
recorded; a `ResponseEnded` event carries the full text as before, and
`replay-is-live` still holds.

**`api.log`** records wire payloads separately. In the first-edition story,
the preceding engine logged the response JSON after
reading the body, and under streaming there is no moment when the body exists
as a blob, because the parser consumes it as it arrives. Left alone, the
response half of the API log would have gone dark the day streaming shipped,
silently, which is the worst way for a debugging facility to fail. `OnFrame`
carries each raw frame out as it arrives and the engine writes it, so the
trace retains payloads in order in both modes. This is not a byte-for-byte
HTTP transcript: SSE framing/comments and authentication headers are not the
payloads delivered by `OnFrame`.

The wiring follows ownership. Parser methods belong to Engine and reach its
Agent parent; Engine reaches `Agent.Ensemble()` for logging. Ensemble owns and
serializes the trace writer. Notification callbacks carry parser output, not
logger services smuggled around this chain. The accepted Engine redacts a
reflected configured key before logging a frame or forwarding `OnFrame`, and
keeps authentication headers and credential-bearing URLs out of the trace.

**Observers** get deltas as they arrive, `PartFinal` on completion, and events
through the normal path. A widget streams chunks in for immediate reading,
then re-renders from the authoritative part. The fast path may be approximate
because the slow path corrects it, and the correction is safe because the
part id links each streamed approximation to the authoritative version that
replaces it.

---

## 6.9 The terminal, and the flush that makes it real

| content | stream | style |
|---|---|---|
| reply text | stdout | plain |
| thinking | stderr | dim |
| tool call arguments | stderr | yellow |

So `agent chat < prompts.txt > answer.txt` yields exactly what the assistant
said and nothing else, while a human sees all three interleaved. One binary
serves the pipeline and the person with no flag, because the stream split
already encodes the difference between output and commentary. Colour is
chosen by `isTerminal(stderr)`, so redirected output is never salted with
escape codes.

With a buffered writer, the line that decides whether any of this is visible:

```go
case common.DeltaText:
	clear()
	fmt.Fprint(out, chunk)
	// Flush per chunk. Without it the buffer holds the whole
	// answer and releases it in one lump at the end, which looks
	// exactly like streaming having no effect.
	out.Flush()
```

A buffered writer does its job perfectly and destroys the feature. Every delta
delivered correctly, every chunk written correctly, and the user sees a blank
terminal then the whole answer at once, indistinguishable from never having
implemented streaming. No error, no warning, no failing test unless somebody
wrote one about timing. A pipeline can be correct at every stage and useless
end to end when the property that matters, here "arrives incrementally",
belongs to no single stage.

The accepted CLI instead writes each chunk directly to `os.Stdout`, an
unbuffered file. Either approach must deliver bytes before the response ends.
Its observer queues display work without dropping reply chunks or waiting on
a slow terminal; after the Agent joins producers, the display drains before
its summary. Finals add separators, not a second copy of the streamed reply.

The initial student streaming CLI had another quiet routing defect. It treated
any chunk beginning `hint:` as a human-hint acknowledgment and moved it to
stderr. A real parser/actor/CLI test sent `hint:early` and then `late`; stdout
contained only `late`. The final implementation marks actual acknowledgments
with `PartDelta.Hint` and routes by origin, leaving assistant prose untouched.
A display prefix is not a protocol discriminator.

The corrected test withholds the rest of the HTTP response until stdout
contains `hint:early`, then requires the complete `hint:earlylate` reply.
That proves flushing while work remains, rather than merely observing several
reads after the provider has finished. Restoring the prefix heuristic fails
this same test; removing the actor's origin marker fails the acknowledgment
check. The observation vocabulary still has the same three content kinds.

The output is a mess: dim reasoning, plain reply, yellow half-formed JSON, all
at once, wrapping at terminal width. Development tooling, honest about it.

---

## 6.10 The exercise

`ch07/main.go` is a separate public-library module. It builds one agent,
attaches one observer, and reports the
**shape** of the stream rather than its content: deltas per kind, parts they
group into, and time to first delta. Total turn latency barely moves, because
the model takes as long as it takes; what changes is how long the human stares
at a blank terminal before the first word appears.

`CH07_NO_STREAM=1` runs the identical script with `DisableStreaming` on.


The reference application is in
[`solutions/ch06/ch07/`](../solutions/ch06/ch07/).
Keep the historical `ch07` directory and environment names. The core's
no-argument JSON protocol receives this prompt, followed by EOF:

```json
{"kind":"prompt","text":"what does the configuration say?"}
```

Its delta records use `observation:"part_delta"`, `part_id`, `kind` and
`chunk`. Finals use `observation:"part_final"` and the matching ID. The
existing exercise envelope exposes top-level `text` or `tool` names; the
neutral `part` retains full content, including tool `args`. The grader also
accepts complete arguments in the envelope. Tool-result finals are distinct
from model-response finals and must not be folded into the streamed response.
A final `assistant` record carries the completed answer.

`EN_DISABLE_STREAMING=1` controls the core CLI, while `CH07_NO_STREAM=1`
controls the separate exercise. Do not silently depend on one name while
testing the other. The original grader launched only the core, despite the
chapter asking for an external exercise; the revised grader builds and runs
both through the actual public library.

The external observer can count in memory and print a summary after reliable
request completion. It need not print individual deltas. Its JSON record uses
`observation:"stream_stats"`, with per-kind `deltas`, finalized `parts`,
`first_delta_ms`, and the final `assistant` text (or a separate answer record).
The accepted example also reports `elapsed_ms`, effective `streaming` and
`usage`. The grader checks all three delta counts, a nonnegative
`first_delta_ms`, and answer delivery. Review the observer's timing and finalized
part counting in source; the summary alone does not prove those measurements
or that a human terminal flushed before completion. The gated CLI regression
described above supplies the flush evidence.

The original deterministic fixture produced this shape, not a promise about
how many chunks a live provider will choose:

| measurement | streaming on | streaming off |
|---|---|---|
| text deltas | 36, across 2 parts | 2 |
| thinking deltas | 27, none leaking into the reply | 1 per part |
| tool call deltas | 19 | 1 per part |
| final reply text | identical | identical |
| finalized parts | identical | identical |

Thirty-six pieces or two pieces. Same answer. `delivery-not-content` grades
this rather than leaving it as a remark, because the claim carries the
chapter: it is why one `Parse` suffices, why the event log needed no changes,
why replay still equals live, and why no observer needs to ask whether a
response arrived in one piece or forty.

### The mutant that states the thesis

In the first-edition audit, `no-stream-flag` removed the request field asking for streaming and nothing
else. Final text identical, parts identical, event log identical byte for
byte, and the grader says:

```
saw 2 text deltas, want at least 8
```

Nothing about the content changed. Only the shape of its arrival did.

Two predictions from that audit were wrong, and both are worth keeping. The
first `per-chunk-part-ids` mutant used a package-level counter; it failed its
intended check and also historical `ch6-parity`, by tripping original Chapter 5's
`no-mutable-globals`. Two failures for one deleted behaviour means the audit
stops telling you which check does which job, so the mutant was rewritten to
derive the id from the chunk. Catching it required logging each failing
check's **details**, not only its id: *which* checks failed says a mutant is
wrong, *why* says whether it is wrong for the intended reason.

Second, `no-thinking-deltas` does not cascade into `deltas-match-final`,
contrary to prediction, because that check walks text parts and ignores the
reasoning stream. Recording that division of labour here keeps it from
looking accidental when a reader sees the two checks side by side.

---

## 6.11 What this chapter does not build

**No re-rendering.** The terminal appends and never redraws a finalized part,
though `PartFinal` provides everything needed. Doing it well needs cursor
control, wrapping, and terminal width, which is Chapter 7's problem wearing a
different hat.

**No acting on partial tool arguments.** Streaming tool arguments is for
**display only**; the arguments are incomplete JSON until the part finalizes.
Starting early, opening the file as soon as the path appears, is how an agent
runs a tool call with half its arguments. Deltas render, finalized parts
execute.

**No observer backpressure.** Observers are told, never asked, and `Observe` must not
block. A slow observer stalling the actor is a deaf agent by another route,
which Chapter 5 was about. An observer that cannot keep up drops optional progress or buffers
without holding up the actor. A terminal promising the full reply must preserve
its bytes; lossy observation queues are not reliable request completion.

---

## Taking it for a spin

Build your own CLI from the core module root, and run it from a scratch
workspace with the provider environment configured as in Chapter 0. Discover
an available streaming-capable model; keep the key in the intended child's
environment without printing it or putting it in a command argument.

```sh
go build -o /tmp/ch6agent .
```

```sh
/tmp/ch6agent chat
```

Type an ordinary prompt:

```
write a short story about a lighthouse keeper who collects fog
```

For a model exposing readable reasoning, it appears dim on stderr. The story
arrives on stdout as the provider delivers it. Ask for a tool and the name and
argument fragments appear on stderr, yellow on a terminal. The tool still waits
for complete arguments. Provider behavior determines whether those arguments
arrive incrementally or as a single object.

The core's explicit opt-out is `EN_DISABLE_STREAMING=1`:

```sh
EN_DISABLE_STREAMING=1 /tmp/ch6agent chat
```

The separate shape exercise uses `CH07_NO_STREAM=1`. Build it from its own
module root and run the same prompt in separate scratch directories:

```sh
go build -o /tmp/ch6shape .
```

```sh
/tmp/ch6shape "Return only the numbers 1 through 20 separated by spaces."
CH07_NO_STREAM=1 /tmp/ch6shape "Return only the numbers 1 through 20 separated by spaces."
```

In your own code the equivalent is one field: `DisableStreaming` on the
Agent's config. The delivered parts should remain intact in either mode.
Two independent live generations need not choose the same words; the fake
makes equivalence exact by returning the same scripted content both ways.

The retained October 9 live shape pair used `gpt-4o-mini` and did return the
same numbered text in both modes. It reported 40 text deltas versus one,
one finalized text part each, and first-delta times of 1,059 ms versus
1,149 ms. Total times were 1,350 ms and 1,150 ms. Each request used 733 input
and 40 output tokens. Streaming improved the first visible delivery in that
small pair; it did not make the total run faster. The exact reports remain in
[enabled](../solutions/evidence/ch07/shape-enabled.json) and
[disabled](../solutions/evidence/ch07/shape-disabled.json) evidence.

The human CLI demonstrations were less tidy, and more useful:

| Historical run | Visible story delivery | What the retained record establishes |
| --- | --- | --- |
| `claude-sonnet-4-6` | 62 stdout reads, first 2.6112 s, last 5.5389 s | Separate reasoning, streamed arguments, successful `echo STREAM_PROBE`; two human turns and one busy hint. |
| `gpt-4o-mini` | 61 stdout reads, first 1.8427 s, last 2.5454 s | Three turns; the model chose nonexistent shell `print`, received exit 127, and accurately recalled the failure. |
| Initial `gemini-2.5-flash` | Four stdout reads, first 1.6210 s, last 1.6541 s | Three turns, thought summaries, whole tool arguments and successful output recall; this model selection failed Bill's instruction. |

These are OS read counts, not token counts. The
[initial evidence summary](../solutions/evidence/ch07/live-summary.json)
records 5,991/261, 3,419/160 and 5,826/248 input/output tokens respectively,
with zero recorded cache counts. The raw event and API logs remain alongside it.

The Anthropic launcher initially labeled its last input a third turn because
it watched file modification times. A log rewrite changed the timestamp more
than once; the launcher sent that input while the second turn was active.
The durable log proved a hint instead. That useful hint was kept, along with
the mistaken capture label and its correction. A changed file is not a reliable
request-completion signal.

The Gemini choice needed a different correction. Bill had required Gemini
3.0 Flash or newer and specifically requested 3.8 Flash for hints. The student
selected 2.5 from discovery; the handoff and first review missed the restriction.
A working older model did not satisfy it. One bounded run on the actual
`gemini-3.8-flash` CLI supplied the missing evidence; no production change or
other-provider repeat was needed.

That run made a real command create `waiting.marker`, wait for `release.marker`,
then print `RELEASED`. After seeing the waiting marker at 5.1767 s, the launcher
sent a busy hint changing a red-lighthouse paragraph to blue and requiring
`HINT_38_APPLIED`. It observed the actual stderr acknowledgment at 5.1780 s,
then released the command at 5.1787 s. The next provider request contained both
the successful tool result and the hint. Its answer began `HINT_38_APPLIED` and
described a blue lighthouse. Tool and CLI both exited zero.

The [corrected Gemini record](../solutions/evidence/ch07/gemini38-summary.json)
contains two requests and eleven response frames, including six nonempty reply
chunks, two thinking chunks and one tool-call frame. Three nonempty stdout
reads spanned only 1.1 ms, so this run proves correct incremental writes and
hint delivery, not perceptibly paced typing. The gated CLI test supplies the
stronger before-completion flush evidence. Usage was 1,241/181 followed by
1,486/755 input/output, including 130 and 649 thinking tokens in those output
counts. No cache counts were recorded.

Afterwards, page through `api.log` from a streamed run: frame payloads in order,
including terminators and opaque signatures, with the configured key redacted
and authentication headers excluded. The accepted
CLI writes it in the working directory; `CH07_AGENT_LOG` selects its durable
event log. The public shape application uses `api.log` and `ch07-agent.jsonl`.
Those are different records serving different questions, and keeping both modes
visible is the purpose of `OnFrame`.

Original student grading remained **0/100 FAIL**. Its five streaming checks
selected unknown `fake-model` while demanding streaming; the implementation
correctly chose plain delivery. Inherited parity failed the documented earlier
exceptions. The original legacy reference scored **90/100 FAIL**, with its
supposedly disabled run still streaming. Those results remain historical;
[the chapter record](../docs/edition-2-notes/ch07.md) separates fixture mistakes,
real CLI corrections and the approved-model live evidence.

---

## What Chapter 7 does with this

The terminal is now honest and ugly: three kinds of content distinguished by
colour, across two file descriptors, in a display that cannot redraw.

The display vocabulary is in place, with response identity still to resolve. Chunks carry a kind, so a consumer can
route them. Chunks carry a part id, so a consumer can group them into one
widget. Parts finalize with the same id, so a consumer can replace streamed
approximations with the authoritative version. Clients attach observers without importing GUI behavior into the core.
Persistent widget identity and connection lifetime belong to the GUI chapter.

A GUI's requirements list, written as a seam before anyone wrote a line of the
GUI. Chapter 7 cashes it in.
