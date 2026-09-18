# Chapter 7: Streaming, or the Same Answer in Pieces

Ask the Chapter 6 agent something hard and the cursor stops. Twenty seconds,
or ninety. Behind the curtain a model is reasoning, writing, deciding to call
a tool, spelling out arguments one token at a time. You get none of it, then
all of it at once.

Every vendor has been willing to send that along as it is produced for years.
The agent never asked. Chapter 6 built an observer seam with a `PartDelta`
observation in it and nothing that produces one. This chapter fills it.

The result is not beautiful. Reasoning, reply, and half-finished tool
arguments land on one terminal in three colours with no layout. The mess is
the point: once you can watch the agent think, you will want somewhere better
to watch it. Chapter 8 builds that. This chapter earns it.

---

## TL;DR

Streaming changes how a response is **delivered**, not what it **is**.
Everything below follows from that one sentence.

### The seam

`Parse` changes signature. It gains no sibling.

```go
// internal/common/config.go
type Parser interface {
	Parse(resp *http.Response, cb StreamCallbacks) error
}
```

No `ParseStream`. Non-streaming is a stream of length one, so one method
covers both modes. `Render` keeps its signature,
`Render(*Context, Config) (*http.Request, error)`; streaming is one field
inside the request it already builds.

### The callbacks

`StreamCallbacks` lives in `internal/common`, the package that declares
`Parser`, because a method on a hub type cannot take a parameter from a spoke.

```go
// internal/common/delta.go
type StreamCallbacks struct {
	// One incremental chunk. partID identifies the PART, not the chunk.
	OnDelta func(partID uint64, kind DeltaKind, chunk string)

	// Each finalized event, in order, for the event log.
	OnEvent func(Event)

	// Each completed part, carrying THE SAME partID its deltas carried.
	OnPartFinal func(partID uint64, part Part)

	// Each raw wire frame: for SSE, one event's data bytes.
	OnFrame func(eventType string, data []byte)
}
```

| callback | wired by | to |
|---|---|---|
| `OnDelta` | actor | the observer |
| `OnPartFinal` | actor | the observer |
| `OnEvent` | engine | the event log |
| `OnFrame` | engine | `APILogf` |

A caller wanting the stream for itself, as chat mode does, supplies its own
`StreamCallbacks` and passes them to `AskWatching`.

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
{"part_id":3,"kind":"text","chunk":"Hello"}
```

`PartDelta` also carries an `agent` field, omitted when empty, filled in when
more than one agent is observed at once.

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
deliberate opposite of what the same table does for media. Section 7.7 is why.

### The reader

`internal/llm/sse.go`: 111 lines, one exported function, 14 tests. It reads an
`io.Reader` and yields `(eventType string, data []byte)` pairs. Every vendor
parser calls it.

### The request

| vendor | asks for streaming with | parser's oddity |
|---|---|---|
| Anthropic | `"stream": true` in the body | dispatch on `content_block_delta`, then on `text_delta` / `thinking_delta` / `input_json_delta` |
| OpenAI | `"stream": true` plus `"stream_options": {"include_usage": true}` | without `include_usage`, `usage` is absent and every token count reads zero |
| Gemini | `?alt=sse` on the URL | each frame is a whole response object with no block indices; continuation is inferred |

### What does not change

The event log. A `ResponseEnded` event still carries the complete response as
`ResponseData{Parts, From, Usage}`, the context still replays from finalized
events, and replay still equals live. Deltas are **ephemeral**: they reach
observers and are never recorded. Raw frames go to `api.log` as they arrive.

### The part id rule

> **Within one response**, a part id identifies exactly one part. Two parts
> never share an id, one part never uses two ids, and an id never crosses
> kinds.

Ids are the vendor's own block ids and restart with each response, so a turn
containing a tool call goes around the loop twice and reuses them. The key a
consumer needs is therefore (response, part id), not part id alone.

The parser supplies the id on every delta and reports the matching
`PartFinal`. The caller never computes one, because on real wire formats it
cannot. Section 7.4 has the receipt.

### The exercise

`ch07/main.go`: one agent, one observer, a report on the shape of the stream
including time to first delta. `CH07_NO_STREAM=1` runs the identical script
with streaming disabled.

```
make grade7
```

| check | points | what it tests |
|---|---|---|
| `stream-deltas` | 20 | Asking for the stream produces one: many deltas per turn, not one per part |
| `deltas-match-final` | 20 | Deltas grouped **by part id** concatenate exactly to that part; at least one part arrived in several deltas |
| `thinking-streamed` | 15 | Reasoning arrives as `DeltaThinking`, correlated by part id, never leaking into the reply |
| `tool-params-streamed` | 15 | A tool call's name and arguments arrive as `DeltaToolCall`, correlated by part id |
| `delivery-not-content` | 10 | Streaming off: same final text byte for byte, same parts in the same order, different delta count |
| `ch6-parity` | 20 | Every Chapter 6 check still passes |
| **total** | **100** | |

---

## 7.1 The idea in plain words

An HTTP response is a stream of bytes. `io.ReadAll` is a choice to ignore
that, convenient enough that it stops looking like a choice.

Non-streaming does not mean the vendor produces the answer atomically. The
model emits tokens one at a time either way; the server holds them until the
last one is written. The latency a user feels is the cost of thinking plus a
decision to say nothing while it happens.

Three things change. The **HTTP client** hands the open response to the parser
instead of reading the body. The **parser** takes a live stream and calls back
per chunk. The **engine** fires an observation per chunk instead of waiting in
silence.

One thing pointedly does not change: the record. At end of turn the event log
holds what it held before, and rebuilding context from it yields the same
bytes. Recording deltas would force every consumer of the log to reassemble
them and would break replay-equals-live, in exchange for a progress indicator.

Hence the sentence that makes the seam small. **Non-streaming is streaming
with a length of one.** A response in one piece is the same kind of thing as a
response in forty pieces. No code downstream of the observer asks whether
something was streamed; it receives deltas and the only question is how many.
Which buys one `Parse` instead of two, one path per vendor instead of two, and
a grader that runs the identical exercise both ways and demands identical
content.

---

## 7.2 Server-sent events, and the reader that survives them

Streaming from an LLM vendor is not WebSocket, not gRPC. It is an ordinary
HTTP response body framed in server-sent events, a text format older than all
of this.

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
| CRLF | leaves `\r` on the JSON, which unmarshals fine about half the time |
| `data: [DONE]` | tries to unmarshal a sentinel, erroring on a stream that completed |
| unterminated final event | drops the last chunk, which usually holds the stop reason |

None is hard. All share a shape: a framing imprecision that produces plausible
output and fails later, somewhere with no mention of SSE in the stack trace.

So the reader is one function, `internal/llm/sse.go`, 111 lines, 14 tests, one
per case above and then some. Vendor-specific work starts after framing is
resolved, with an event type and a slice of JSON in hand. Chapter 3's "fakes
first" pattern in different clothes: the framing is deterministic, ugly, and
testable without a network.

---

## 7.3 One `Parse`, not two

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

One method makes that drift unrepresentable:

```go
Parse(resp *http.Response, cb StreamCallbacks) error
```

On SSE, `Parse` reads frames and calls `OnDelta` per chunk. On a plain JSON
body it reads the body whole and calls `OnDelta` once per part with that
part's entire content. Both routes then run the same code to build events and
call `OnEvent` and `OnPartFinal`.

The non-streaming route could skip `OnDelta` entirely, since nothing is
incremental. It fires anyway, because "length one" has to be true in code and
not only in prose. Were it silent, every observer would need a second
rendering path fed from finalized events, and the seam would have bought
nothing.

The seam still has exactly two methods, `Render` and `Parse`, which is what
the vendor-independence claim has rested on since Chapter 2. Adding a third to
ship a feature would concede that the seam was shaped around the features that
existed when it was drawn.

---

## 7.4 Part ids: who is allowed to name a part

A stream delivers forty chunks belonging to a paragraph of reasoning, a reply,
and a tool call's arguments. A GUI appends each chunk into the right widget,
then replaces that widget's content with the finalized part. It needs to know
which chunks belong together. A part id is nothing more than that.

This chapter's brief originally minted a fresh id per chunk with an
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

The parser chose it, so the parser owns both ends: it supplies the id on every
`OnDelta` and reports `OnPartFinal` carrying the same id. Finals carry the
vendor's own block id, not a position in the finished list. The id is thereby
freed from meaning "where", and only ever means "the same part as that other
one", which is the sole property a consumer can rely on.

Inheriting the vendor's block ids has a consequence worth stating before
Chapter 8 trips over it: **the ids restart with each response**. A turn that
calls a tool goes around the loop at least twice, and part id 0 in the second
response is a different part from part id 0 in the first. Uniqueness holds
within a response, not within a turn. The grader for this chapter had to learn
the same lesson, and its harness now tracks response boundaries rather than
assuming ids are unique across a turn.

So a consumer keying widgets needs (response, part id). Chapter 8 will need a
response boundary that is observable from outside the agent, and the current
seam does not clearly offer one. That is an open problem rather than a solved
one, and it is named here so it arrives as a known cost rather than a
surprise.

`OnPartFinal` exists for this. It looks redundant beside `OnEvent`, since
finalized parts are inside the events, but the events do not carry the ids the
deltas used, and reconstructing that mapping from outside is exactly the guess
that fails.

Ordering matters too. `OnPartFinal` fires **after** the event carrying that
part is recorded, so an observer re-rendering on finalization reads a log that
already contains it. Reversed, every consumer gets an intermittent view of a
log one event behind.

`deltas-match-final` grades this the hard way. A check that concatenated every
text delta in a turn and compared against the final message would pass on a
submission whose ids were noise, including the fresh-id-per-chunk submission
this section rules out. So the check groups by id first, requires each group
to concatenate exactly to its own part, and requires at least one part to have
arrived in several deltas. Grouping is the check; concatenation is the easy
half.

---

## 7.5 Three vendors, three dialects

The framing is shared. The meaning is not, and the differences are listed in
the TL;DR table above. Two deserve expanding.

OpenAI's `stream_options.include_usage` is not decoration. Omit it and a
streamed response contains no `usage` object at all: not a zero, not an error,
simply absent, so the accounting built in Chapter 2 records zeros for every
streamed turn. The answers are correct, the tools run, the log is well formed,
and cost tracking reads zero forever. A silent success that is wrong, one JSON
field deep.

Gemini sends no incremental deltas. Every frame is a whole response object
with the new content inside and no block indices anywhere, so nothing says
"this text continues the previous frame" and continuation is inferred from
position and shape.

The seam does not make vendors identical. It makes their differences
**local**: all three oddities live inside one vendor's `Parse`, and none
appears in the engine, actor, observer, event log, or terminal. Locality is
the whole claim, and it is smaller than "vendors are interchangeable", which
was never true.

---

## 7.6 When a capability table pays for itself

Gemini's row has `StreamToolArgs` clear.

```go
"gemini-3.8-flash": {
	Media:  MediaImage | MediaAudio | MediaVideo | MediaDocument,
	Stream: StreamText | StreamThinking,
},
```

Streamed function parameters do not arrive incrementally there. The
instructive part is where that fact is **not** written: there is no
`if vendor == "gemini"` in the parser, no special case in the engine, no
apology in the actor. The table says the bit is off, `StreamingFor` returns a
mask without it, and tool arguments for that model arrive as a length-one
stream through the identical path. Behaviour degrades exactly as far as the
capability is missing and not one line further.

| model family | text | thinking | tool arguments |
|---|---|---|---|
| `claude-opus-5`, `claude-sonnet-5` | yes | yes | yes |
| `gpt-6-astra`, `gpt-5.6-sol` | yes | no | yes |
| `gemini-3.8-flash`, `gemini-3.1-pro-preview` | yes | yes | no |

Three families, three combinations, no two the same shape. A single
`Streaming bool` would force two of those rows to choose between losing text
streaming and inventing tool-argument chunks that never arrived. The bitmask
is the minimum structure the observed facts require.

The table will go stale; vendors ship these bits on their own schedule. The
fix is then a line of data rather than a code change, in a place a reader will
think to look, which is the point of putting it in a table.

---

## 7.7 Guessing about delivery is not guessing about content

Given a model name it has never seen, `StreamingFor` shrugs: no error, no
refusal, a non-streaming request, turn proceeds. Asked whether the same
unknown model accepts an image, the same table produces a loud refusal, which
Chapter 5 spent real effort keeping. Two lookups, two opposite answers on a
miss. An inconsistency like that is either a bug or a principle.

> A guess about **content** can corrupt the conversation. A guess about
> **delivery** cannot.

Send an image to a model that cannot see and the image is dropped. The request
succeeds, the model answers confidently about a picture it never received, and
nothing downstream can detect or recover it. Guess wrong about streaming and
the same response arrives, same text, same events, same replayed context, in a
different number of pieces. The worst case is a slower first token.

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

## 7.8 Three record-keepers

**The event log** is unchanged. Deltas fire at observers and are never
recorded; a `ResponseEnded` event carries the full text as before, and
`replay-is-live` still holds.

**`api.log`** needed work. Chapter 6's engine logged the response JSON after
reading the body, and under streaming there is no moment when the body exists
as a blob, because the parser consumes it as it arrives. Left alone, the
response half of the API log would have gone dark the day streaming shipped,
silently, which is the worst way for a debugging facility to fail. `OnFrame`
carries each raw frame out as it arrives and the engine writes it, so the
trace stays byte-complete in both modes.

The wiring enforces a Chapter 5 rule. Vendor parsers are stateless empty
structs with no `Host`, deliberately, so they cannot log; the engine has a
`Host`, so the engine wires `OnFrame` to `APILogf`. Logging stays a facility a
parent provides rather than one the seam grants itself.

**Observers** get deltas as they arrive, `PartFinal` on completion, and events
through the normal path. A widget streams chunks in for immediate reading,
then re-renders from the authoritative part. The fast path may be approximate
because the slow path corrects it, and the correction is safe only because the
part id ties them together.

---

## 7.9 The terminal, and the flush that makes it real

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

Then the line that decides whether any of this is visible:

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

The output is a mess: dim reasoning, plain reply, yellow half-formed JSON, all
at once, wrapping at terminal width. Development tooling, honest about it.

---

## 7.10 The exercise

`ch07/main.go` builds one agent, attaches one observer, and reports the
**shape** of the stream rather than its content: deltas per kind, parts they
group into, and time to first delta. Total turn latency barely moves, because
the model takes as long as it takes; what changes is how long the human stares
at nothing.

`CH07_NO_STREAM=1` runs the identical script with `DisableStreaming` on.

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
why replay still equals live, and why no observer asks whether a response was
streamed.

### The mutant that states the thesis

`no-stream-flag` removes the request field asking for streaming and nothing
else. Final text identical, parts identical, event log identical byte for
byte, and the grader says:

```
saw 2 text deltas, want at least 8
```

Nothing about the content changed. Only the shape of its arrival did.

Two predictions from that audit were wrong, and both are worth keeping. The
first `per-chunk-part-ids` mutant used a package-level counter; it failed its
intended check and also `ch6-parity`, by tripping Chapter 5's
`no-mutable-globals`. Two failures for one deleted behaviour means the audit
stops telling you which check does which job, so the mutant was rewritten to
derive the id from the chunk. Catching it required logging each failing
check's **details**, not just its id: *which* checks failed says a mutant is
wrong, *why* says whether it is wrong for the intended reason.

Second, `no-thinking-deltas` does not cascade into `deltas-match-final`,
contrary to prediction, because that check walks text parts and ignores the
reasoning stream. The division of labour is recorded as deliberate rather than
left looking accidental.

---

## 7.11 What this chapter does not build

**No re-rendering.** The terminal appends and never redraws a finalized part,
though `PartFinal` provides everything needed. Doing it well needs cursor
control, wrapping, and terminal width, which is Chapter 8's problem in
disguise.

**No acting on partial tool arguments.** Streaming tool arguments is for
**display only**; the arguments are incomplete JSON until the part finalizes.
Starting early, opening the file as soon as the path appears, is how an agent
runs a tool call with half its arguments. Deltas render, finalized parts
execute.

**No backpressure.** Observers are told, never asked, and `Observe` must not
block. A slow observer stalling the actor is a deaf agent by another route,
which Chapter 6 was about. An observer that cannot keep up drops or buffers on
its own time.

---

## Taking it for a spin

With a real API key:

```
$ agent chat
> write a short story about a lighthouse keeper who collects fog
```

Reasoning appears first, dim, on stderr. The story then arrives a few words at
a time. Ask for a tool and the arguments spell themselves out in yellow, the
first time in this book the agent's intentions are visible before it acts.

Then the same prompt without streaming:

```
$ CH07_NO_STREAM=1 go run ./ch07
```

In your own code the equivalent is one field: `DisableStreaming` on the config
handed to the agent. Same story, one lump, the wait returns.

Afterwards, page through `api.log` from the streamed run: every chunk the
vendor sent, in order. Complete in both modes precisely because `OnFrame`
exists, and the most useful debugging artifact this framework has.

---

## What Chapter 8 does with this

The terminal is now honest and ugly: three kinds of content distinguished by
colour, across two file descriptors, in a display that cannot redraw.

Everything needed to fix it exists. Chunks carry a kind, so a consumer can
route them. Chunks carry a part id, so a consumer can group them into one
widget. Parts finalize with the same id, so a consumer can replace streamed
approximations with the authoritative version. Observers attach and detach
freely, and the agent neither knows nor cares who is listening.

A GUI's requirements list, written as a seam before anyone wrote a line of the
GUI. Chapter 8 cashes it in.
