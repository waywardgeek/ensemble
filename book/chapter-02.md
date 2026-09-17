# Chapter 2: One Log, Three Vendors

This chapter introduces the core data structures that every AI coding agent
in the world gets wrong. What follows is the most valuable chapter in the
book.

Chapter 1's `Conversation` is the vendor's request body wearing your type
name. Its roles are Anthropic's roles, its content is Anthropic's content,
and a second vendor changes every function that touches it. Do not adapt it.
Stop storing what you send.

Store what happened: an append-only log of events, each with a sequence
number. Derive the context, the state a request is built from, by folding
the log through one function. Put two functions at the vendor boundary: one
turns a context into an HTTP request, one turns an HTTP response into
events. Everything the vendor believes about the conversation is computed at
that boundary and thrown away. Everything you believe is in the log.

Three things follow, and the grader checks each. One log renders to
Anthropic, OpenAI, and Gemini requests, so switching vendors is an
environment variable. Rendering a log twice gives identical bytes, so a bug
report is a log file and a session survives a restart or a retired model.
Forgetting is an event appended to the log, never a deletion from it, so
anything compacted out of the context comes back by replaying without the
redaction.

**What you build.** The log, the parts, the context, and the seam. These are
the reference's types, field for field. Every enum starts at `iota + 1` so a
zero value means "never filled in" and is refused wherever it is read.

```go
type Seq uint64

type EventType uint8

const (
	MessageReceived EventType = iota + 1
	RequestSent
	ResponseStarted
	ResponseEnded
	ToolCalled
	ToolReturned
	Redacted
	ErrorOccurred
)

type Actor uint8

const (
	ActorHuman Actor = iota + 1
	ActorAgent
	ActorSystem
	ActorTool
)

// Event is one line of the log. Exactly one payload pointer is non-nil,
// selected by Type. Time is metadata; ordering comes from Seq, never Time.
type Event struct {
	Seq  Seq       `json:"seq"`
	Type EventType `json:"type"`
	Time time.Time `json:"time"`

	Message  *MessageData  `json:"message,omitempty"`
	Request  *RequestData  `json:"request,omitempty"`
	Response *ResponseData `json:"response,omitempty"`
	Tool     *ToolData     `json:"tool,omitempty"`
	Redact   *RedactData   `json:"redact,omitempty"`
	Error    *ErrorData    `json:"error,omitempty"`
}

type MessageData struct {
	Actor Actor    `json:"actor"`
	Parts PartList `json:"parts"`
}

type RequestData struct {
	To Provenance `json:"to"`
}

type ResponseData struct {
	Parts PartList   `json:"parts"`
	Usage Usage      `json:"usage"`
	From  Provenance `json:"from"`
}

// ToolData serves both ToolCalled (CallID, Name, Args) and ToolReturned
// (CallID, Parts, IsError).
type ToolData struct {
	CallID  string          `json:"call_id"`
	Name    string          `json:"name,omitempty"`
	Args    json.RawMessage `json:"args,omitempty"`
	Parts   PartList        `json:"parts,omitempty"`
	IsError bool            `json:"is_error,omitempty"`
}

// RedactData names a span of the log, From..To inclusive, that later
// renders must not carry at the given Level.
type RedactData struct {
	From        Seq       `json:"from"`
	To          Seq       `json:"to"`
	Level       Redaction `json:"level"`
	Replacement PartList  `json:"replacement,omitempty"` // RedactSummary only
	Reason      string    `json:"reason,omitempty"`
}

type ErrorData struct {
	Message string `json:"message"`
	Status  int    `json:"status,omitempty"`
}

type Redaction uint8

const (
	RedactResult   Redaction = iota + 1 // result content -> stub; the call survives
	RedactTool                          // call and result both go
	RedactDialogue                      // prose and reasoning go
	RedactSummary                       // span replaced by compressed prose
)

// Usage is four disjoint meters. Making them disjoint is the parser's job.
type Usage struct {
	Input      int `json:"input"`       // neither read from nor written to cache
	CacheWrite int `json:"cache_write"` // typically costs MORE than plain input
	CacheRead  int `json:"cache_read"`  // typically an order of magnitude LESS
	Output     int `json:"output"`
}
```

Content is a sealed set of parts, never a string. A blob is a locator, never
bytes.

```go
type Part interface{ isPart() }
type PartList []Part

type TextPart struct{ Text string }

type RefKind uint8

const (
	RefPath   RefKind = iota + 1 // a file on local disk
	RefURI                       // remote: vendor File API uri, gs://, https://
	RefHandle                    // framework-managed output; may be in memory
)

type Ref struct {
	Kind    RefKind `json:"kind"`
	Locator string  `json:"locator"`
}

type BlobPart struct {
	MIME string
	Ref  Ref
}

// OpaquePart is material the model issued and expects back verbatim: a
// thinking block with its signature. Carried, never interpreted.
type OpaquePart struct {
	From Provenance
	Data json.RawMessage
}

// RedactedPart is synthesized by the reducer when a Redacted event covers
// content. It is never written to the log.
type RedactedPart struct {
	Stub string
	Ref  Ref // zero when the superseded content had no locator
}

type ToolCallPart struct {
	CallID string // the id AS ISSUED, by the model named in From
	From   Provenance
	Name   string
	Args   json.RawMessage

	Opaque json.RawMessage // per-call replay material; returns only to its model
}

type ToolResultPart struct {
	CallID  string
	Parts   []Part
	IsError bool // a tool that ran and failed is CONTENT, not ErrorOccurred
}
```

Where a part came from is recorded when it is written, per model, and never
inferred later.

```go
type Vendor uint8
type Surface uint8

const (
	VendorAnthropic Vendor = iota + 1
	VendorGemini
	VendorOpenAI
)

const (
	SurfaceMessages        Surface = iota + 1 // Anthropic
	SurfaceChatCompletions                    // OpenAI
	SurfaceGenerateContent                    // Gemini
	SurfaceInteractions                       // Gemini's replacement surface
	SurfaceResponses                          // OpenAI's newer surface
)

type Provenance struct {
	Vendor  Vendor  `json:"vendor"`
	Model   string  `json:"model"` // OPEN set. Never switch on it.
	Surface Surface `json:"surface"`
}
```

The context is derived state. No field in it grows without bound.

```go
type TurnState uint8

const (
	Idle TurnState = iota + 1
	InputPending
	InFlight
	ToolsPending
)

type Entry struct {
	Seq   Seq      `json:"seq"`
	Actor Actor    `json:"actor"`
	Parts PartList `json:"parts"`
}

type Context struct {
	Turn     TurnState `json:"turn"`
	Dialogue []Entry   `json:"dialogue"`
	Ephemera PartList  `json:"ephemera"` // pending; delivered once, then cleared
	Usage    Usage     `json:"usage"`    // running totals, vendor-normalized
}

// Apply is total: a (state, event) pair it does not list is the identity.
func (c *Context) Apply(e Event) error
```

`Apply` in full: `MessageReceived` from `ActorSystem` appends the parts to
`Ephemera` and changes nothing else; from any other actor it appends an
`Entry` and moves `Idle` to `InputPending`. `RequestSent` clears `Ephemera`
and moves `InputPending` to `InFlight`. `ResponseEnded` appends an
`ActorAgent` entry, adds the usage, and moves to `ToolsPending` if any part
is a `ToolCallPart`, else back to `Idle`. `ToolReturned` appends an
`ActorTool` entry holding one `ToolResultPart` and moves to `InputPending`
when no call is outstanding. `Redacted` rewrites the covered entries in
place. `ErrorOccurred` moves `InFlight` or `ToolsPending` to `Idle`.
`ResponseStarted` and `ToolCalled` change nothing in the context.

The seam is two methods, with no vendor type in either signature.

```go
type Config struct {
	Vendor       Vendor
	Surface      Surface
	Model        string
	BaseURL      string
	APIKey       string
	SystemPrompt string
	MaxTokens    int

	AcceptsAudio bool
}

type Renderer interface {
	Render(*Context, Config) (*http.Request, error)
}

type Parser interface {
	Parse(status int, body []byte) ([]Event, error)
}

// SeamFor is the only function in the program that switches on Vendor.
func SeamFor(v Vendor) (Renderer, Parser, error)
```

`Parse` returns events rather than a message or a context, so there is one
door into the context. The system prompt is on `Config` and absent from the
log: it is rendered output. The engine that wires `main` to these types is
yours; nothing grades it.

**The log on disk.** JSON-lines: a header line, then one `Event` per line.
Enum values are written as names (`"human"`, `"anthropic"`, `"messages"`,
`"redact_result"`); `Ref.kind` is written as its integer; each part carries a
`"type"` discriminator. The grader compares names case-insensitively with
punctuation stripped, so `tool_called` and `ToolCalled` are one name. This
is the exact log the grader hands to `render`:

```json
{"log_version":1}
{"seq":1,"type":"message_received","time":"2026-01-01T00:00:00Z","message":{"actor":"human","parts":[{"type":"text","text":"read config.json"}]}}
{"seq":2,"type":"response_ended","time":"2026-01-01T00:00:01Z","response":{"parts":[{"type":"opaque","from":{"vendor":"anthropic","model":"claude-sonnet-5-course","surface":"messages"},"data":{"type":"thinking","thinking":"The user wants the config file.","signature":"sig-exhibit-1"}},{"type":"text","text":"I'll read it."},{"type":"tool_call","call_id":"toolu_exhibit_1","from":{"vendor":"anthropic","model":"claude-sonnet-5-course","surface":"messages"},"name":"read_file","args":{"path":"config.json","limit":40}}],"usage":{"input":100,"cache_write":0,"cache_read":50,"output":20},"from":{"vendor":"anthropic","model":"claude-sonnet-5-course","surface":"messages"}}}
{"seq":3,"type":"tool_called","time":"2026-01-01T00:00:02Z","tool":{"call_id":"toolu_exhibit_1","name":"read_file","args":{"path":"config.json","limit":40}}}
{"seq":4,"type":"tool_returned","time":"2026-01-01T00:00:03Z","tool":{"call_id":"toolu_exhibit_1","parts":[{"type":"text","text":"port=8080\nhost=localhost"}]}}
{"seq":5,"type":"message_received","time":"2026-01-01T00:00:04Z","message":{"actor":"human","parts":[{"type":"text","text":"now check the logs instead"}]}}
```

Three more spellings appear in grader logs: a blob, a tool call carrying
per-call material, and a redaction event.

```json
{"type":"blob","mime":"image/png","ref":{"kind":2,"locator":"https://generativelanguage.googleapis.com/v1beta/files/ch2exhibit"}}
{"type":"tool_call","call_id":"gemini_call_1","from":{"vendor":"gemini","model":"gemini-3.5-flash-course","surface":"generate_content"},"name":"read_file","args":{"path":"deploy.sh"},"opaque":"sig-bound-to-this-call"}
{"seq":6,"type":"redacted","time":"2026-01-01T00:00:05Z","redact":{"from":4,"to":4,"level":"redact_result","reason":"compaction"}}
```

Gemini's surface is spelled both `"generatecontent"` and
`"generate_content"` across the grader's logs; a loader that compares names
the way the grader does accepts both.

**Rules.** The grader runs the binary and reads its stdout, stderr, exit
status, log, and the fake vendor's record of every request. Each rule names
the check that fails it.

1. **Chapter 1 still passes, and the protocol grows.** All seven Chapter 1
   checks run unchanged against the new binary with only the `ANTHROPIC_*`
   variables set; Chapter 1's usage check reads `input` and `output` from
   the usage line and ignores the other two meters. Beyond them the binary
   reads `LLM_VENDOR` (`anthropic`, `openai`, or `gemini`; default
   `anthropic`), then `LLM_API_KEY`, `LLM_MODEL`, and `LLM_BASE_URL`, each
   falling back to the vendor's own spelling (`ANTHROPIC_MODEL`,
   `OPENAI_API_KEY`, `GEMINI_BASE_URL`, and so on), and `CH02_LOG`, the path
   of the log, defaulting to a name derived from the executable's. Grader
   mode starts a fresh log at that path; it does not append to a previous
   session's. Requests go to the vendor's own path under the base URL:
   `/v1/messages`, `/v1/chat/completions`,
   `/v1beta/models/{model}:generateContent`. Stdin takes two directives.
   `{"user": "..."}` is answered by one request and one `{"agent": "..."}`
   line carrying every text part of the reply; a tool call in the reply is
   recorded and not executed. `{"ephemeral": "..."}` is answered by
   `{"ack": "ephemeral"}` and no request. At EOF the last line is
   `{"usage": {"input": N, "cache_write": N, "cache_read": N, "output": N}}`.
   Every other stdout line, an error report included, fails the session.
   (`ch1parity`, `session`)
2. **The log re-reads.** `./ch02 dump` prints the log at `CH02_LOG` to
   stdout in ascending `seq`, and that output, saved and rendered in a fresh
   process, produces a request. All three `RefKind`s survive the round trip.
   A blob written with the pre-`Ref` `"path"` key, and a `Ref` whose `kind`
   is 0, are refused by the loader: exit non-zero, nothing on stdout, a
   diagnostic on stderr that names the field. Coercing either to `RefPath`
   fails. (`logdump`, `ref-roundtrip`, `ref-oldformat`, `ref-zerokind`)
3. **Rendering is a pure function of the log.** `LLM_VENDOR=v ./ch02 render
   LOG` prints one JSON request body on stdout, exit 0, nothing on stderr,
   with no network and no key. Two runs in two processes are byte-identical
   for every vendor. No clock in the renderer, no map iterated into output,
   no random id; when a call carries no vendor id, the renderer mints one
   from the entry's `Seq` and the call's position, since `Parse` runs before
   any `Seq` is assigned. (`replay`)
4. **One log, three shapes, and per-model material goes home.** Anthropic:
   top-level `system`; `tool_use` in an `assistant` message; `tool_result`
   with `tool_use_id` first in the next `user` message, the human's
   following text merged after it in the same message. OpenAI: the system
   prompt as a `system` message in the array; `tool_calls[].function.arguments`
   a JSON-encoded string; the result as role `tool` with `tool_call_id`.
   Gemini: `systemInstruction` hoisted out; `contents`, never `messages`;
   the agent's role is `model`; `functionCall.args` an object;
   `functionResponse` in a `user` turn with `name` resolved from the matching
   call and `response` a JSON object (the reference sends
   `{"output": "<the result's text>"}`). `Config.Surface` is the vendor's
   surface named above; nothing else is read from the environment for it.
   Opaque material renders only when its `Provenance` equals the configured
   vendor, model, and surface, all three, and is omitted for any other
   model. Anthropic gets an `OpaquePart`'s `Data` back verbatim as a content
   block of the agent's message; Gemini gets `ToolCallPart.Opaque` as
   `thoughtSignature`, a sibling key of `functionCall`; the OpenAI surface
   has no slot for either and gets neither. A blob's rendering depends on
   where it sits. In a human entry, a `RefURI` blob renders to Gemini as
   `fileData` with `mimeType` and `fileUri`, and Anthropic and OpenAI return
   an error rather than a request missing its attachment. Inside a tool
   result no vendor has an attachment slot, so a blob there flattens to
   text, `[<mime> at <locator>]`, on all three. (`seam-render`, `ref-render`)
5. **Three responses, one truth.** Parsing the same scripted turn from each
   vendor yields dialogues that agree on every text and on every tool call's
   name and decoded arguments. Only `Provenance`, the vendor's call id, and
   opaque material may differ, and `Provenance` must differ. The parser
   writes `Provenance` from the response's model field (`model` on
   Anthropic and OpenAI, `modelVersion` on Gemini) and its own surface
   rather than from config, and `dump` writes vendor and surface as names.
   The Chapter 2 fake serves each reply's text as one block on every vendor;
   whether your parser merges adjacent text blocks is yours. (`seam-parse`)
6. **Four meters, disjoint.** The parser normalizes each vendor's accounting
   into `Usage`; `Context.Usage` accumulates; the EOF line prints it. Nothing
   reads a total from the HTTP client. (`usage`)

   | meter | Anthropic `usage` | OpenAI `usage` | Gemini `usageMetadata` |
   |---|---|---|---|
   | `input` | `input_tokens` | `prompt_tokens` minus both cache counts | `promptTokenCount` minus `cachedContentTokenCount` |
   | `cache_write` | `cache_creation_input_tokens` | `prompt_tokens_details.cache_write_tokens` | 0; no token count is reported |
   | `cache_read` | `cache_read_input_tokens` | `prompt_tokens_details.cached_tokens` | `cachedContentTokenCount` |
   | `output` | `output_tokens` | `completion_tokens` | `candidatesTokenCount` plus `thoughtsTokenCount` |

   The grader's two-round session reports, per vendor, `{100, 0, 50, 30}`
   then `{12, 0, 200, 8}` in these four meters; the printed total must be
   `{112, 0, 250, 38}` on all three. A parser that adds OpenAI's or Gemini's
   prompt total to the cache count prints 362. A one-round cache-write probe
   expects `{7, 300, 0, 11}` on Anthropic and OpenAI and `{7, 0, 0, 11}` on
   Gemini.
7. **The context forgets on purpose, two ways.** Ephemera ride exactly one
   request, the first sent after they arrive, and never enter the dialogue;
   the marker text appears in request one of the ephemera session and in no
   other. On the wire they are human text placed last, after every dialogue
   entry: merged into the final human message when the dialogue ends on
   one, otherwise one more. A `Redacted` event at level `redact_result`
   over a `ToolReturned` replaces the result's parts with one
   `RedactedPart`, rendered as text; the stub's wording is yours, and the
   reference writes `[redacted: N bytes; full output retained]`. The tool
   call survives in every later render and the redacted bytes appear in
   none. When a redacted part carried a `Ref`, the stub carries it, and a
   vendor with a remote-file part renders it: after redaction the Gemini
   request still holds the locator in `fileData`. The other two carry the
   stub alone; the grader checks the locator on Gemini. (`ephemera`,
   `redaction`, `ref-redaction`)

Fourteen checks, all or nothing each, sum 100: `ch1parity` 10 and `session`
0 (rule 1), `logdump` 5, `ref-roundtrip` 3, `ref-oldformat` 2,
`ref-zerokind` 2 (rule 2), `replay` 10 (rule 3), `seam-render` 15 and
`ref-render` 4 (rule 4), `seam-parse` 15 (rule 5), `usage` 10 (rule 6),
`ephemera` 10, `redaction` 10, `ref-redaction` 4 (rule 7).

**Yours.** Structural, and the grader fails you for it: the log as the only
truth, a total reducer, the two-function seam, provenance per model at write
time, redaction as an event, disjoint usage, no unbounded field in
`Context`. Taste, which the reference chose and nothing checks:

- **What to forget first.** Four redaction levels ship; one is graded. Which
  category your agent compacts first, and at what fraction of the window, is
  the most personal decision in this chapter. The reference makes
  `RedactSummary` a fold, N entries to one, because the per-entry version
  grew the context it was meant to shrink.
- **A log from the future.** The reference refuses a `log_version` it does
  not know and treats a missing header as version 1. Decide what yours does
  when it meets its successor's log.
- **Media capability.** `Config.AcceptsAudio` makes an audio blob to a
  text-only model an error. A bitmask over media kinds is the generalization
  once your agent handles images.
- **A content hash on `Ref`.** Redaction is recoverable only while the bytes
  at the locator are unchanged. A hash beside the locator turns a silent
  substitution into a loud one. The reference carries none.
- **The engine.** The reference wraps the CLI in an `Engine` with `Say`,
  `Attach`, `Turn`, `Ask`, and `Save`, plus a package function `RenderOnly`
  that plays a log to a context and renders it with no network call.
  Nothing grades the shape between `main` and the seam.

**Exercise.** Rewrite Chapter 1's binary as `ch02` with three modes: no
arguments is grader mode, `dump` prints the log, `render LOG` prints one
vendor request with no network. `make grade-dir CH=2 DIR=path/to/yours`
must print `100/100`; `make grade2` runs the same grader on the reference.
The fake serves every request; cost is nothing.

## 2.1 The idea in plain words

The vendor's API is stateless. Every request carries the whole conversation
in the vendor's shape, and Chapter 1 stored the conversation in that shape,
so the program's memory was a copy of one vendor's request body. That holds
for one vendor and for as long as that vendor's shape stands still. The rest
of the chapter is one idea with five consequences, and the idea is the
bank's: keep the ledger, compute the balance.

**Keep the ledger, compute the balance.** A bank records every transaction
and computes your balance from the record; it never edits the balance by
hand. The event log is the ledger: a human typed, a request went out, a
response came back, a tool ran, something was forgotten. The context is the
balance: the conversation as the model should see it right now. Replay the
ledger and the balance comes back, on any machine, to the byte. Four
things you get for free once the ledger is the truth. A GUI draws the
transcript from the log, as a chat, a timeline, or a diff, without changing
what the model sees. A bug report is a log file, and replaying it
reproduces the request that went wrong. An audit is a read of the log: what
the agent did, in what order, and what it was told. And when the vendor
retires a model, the conversation survives it, because the log recorded what
happened and never recorded the vendor's shape.

**Save the context, and as much log as you need.** To continue a
conversation after a restart you need the context, or enough of the log to
rebuild it. An agent that does not care about history saves the context
alone and starts a fresh log each session. An agent with a GUI keeps as much
log as the GUI shows; an agent that must answer for what it did keeps all of
it. The reference saves the whole log and rebuilds the context from it on
every start, because that is the cheapest proof that the context is a
function of the log. How much log your agent keeps is a policy; the shapes
in the TL;DR do not change with it.

**One context, two translators per vendor.** Every vendor wants the same
conversation in a different shape: different role names, a different place
for the system prompt, a different container for a tool result. The old way
is a client per vendor, each holding the history in its own shape, which
means three copies of every bug and a rewrite of the product each time a
vendor changes. This chapter's way is one context and, per vendor, exactly
two functions: render the context into a request, parse the response into
events. Everything a vendor is peculiar about lives in those two functions,
and nothing else in the program knows which vendor it is talking to.
Switching vendors is one variable. Adding one is one renderer and one
parser, and §2.7 states how to check that claim against your own code.

> **Receipt.** CodeRhapsody's three vendor clients, each with its own
> history in its own shape, measured 31,364 lines with their tests. The
> rebuild on one context, with a renderer and a parser per vendor, measured
> 16,175 lines for all three vendors and everything around them, as of
> September 2026.

**The context is a superset of what any vendor can say.** For one context
to render to every vendor, it has to hold everything any of them can
produce or accept: text, tool calls and their results, files, and material
a model hands back that only it can read, such as a thinking signature.
Some of that material is bound to the exact model that produced it. A
signature valid for one model is a 400 on another, and a thinking block
one vendor requires back is a block another vendor silently drops. So every
such part records where it came from, vendor, model, and surface, at the
moment the parser writes it, and a renderer for any other model leaves the
part out. The vendor's peculiarity is carried, labeled, and never
interpreted.

**Turn state lives in the context and moves only on events.** Whose turn
is it, is a request in flight, is a tool call waiting for its result: the
answer decides what to do with the next line of input, and inferring it
from the shape of the history ("the last message is a tool result, so a
request must be pending") is fragile, because every edge case is a bug in a
different place. Here the state is one field, changed by the reducer on
specific events, replayed like everything else, and readable by anyone
holding the context. CodeRhapsody's earlier clients inferred turn
boundaries from the history, and that inference was a recurring source of
intermittent bugs; its current engine keeps turn state as a value that
events move, which is the design this chapter ships.

**Forgetting is an event.** Context windows fill and something has to go.
Delete it from the history and it is gone, along with any way to show the
user what was removed or to get it back when the deletion was wrong. Append
a redaction event instead and the context shrinks, the log grows by one
line, the GUI can show what went, and replaying without that line brings
it back. Forgetting becomes a decision your agent made, on the record, and
revisable.

## 2.2 Eight events, and why `ToolCalled` carries nothing

The taxonomy is frozen for the chapter at eight members and starts at
`iota + 1`. Each event carries exactly one payload, chosen by `Type`, and
the payload structs are small on purpose: `RequestData` is one
`Provenance`, `ErrorData` is a message and a status. An event is a fact
about one moment, and a fact about one moment does not need a field for
every other moment.

The model's tool calls travel in `ResponseEnded.Parts` as `ToolCallPart`s,
in order among the text and opaque parts of the same turn. `ToolCalled` is
the engine's event, recording that a call was dispatched, and `Apply`
changes nothing on it. The alternative, a `ToolCalled` event that carries
the call's content and a `ResponseEnded` that carries only text, loses the
order of a call relative to the text around it, and the model's turn was
the text `I'll read it.` followed by the call, in that order. A vendor that
renders the two in the wrong order is a vendor that gets a confused model.

`ToolResultPart.IsError` is content and `ErrorOccurred` is not. A tool that
ran and failed produced output the model must see, in the dialogue, with
the failure flagged. A request that got no response produced nothing the
model can see; it changed the turn state and nothing else.

`Apply` describes information flow. Everything needed to advance the
context is the context plus one event, and that claim is the chapter's
center: it is what makes the context recomputable from the log and the log
sufficient as a bug report. In Go the method has a pointer receiver and
mutates in place. A `Context` copied by value carries slices whose backing
arrays are shared with the original, and two contexts sharing one array
produce a bug that looks exactly like renderer non-determinism (§2.9) and is
found by neither the `replay` check nor a field-by-field diff.

Classification belongs to the reducer, never to the capture site. The same
line arriving on stdin is a new prompt when the turn is `Idle` and a hint to
be appended when it is `InFlight`; the engine that reads stdin does not know
which, and if it guesses it is wrong every time the human types fast. The
reducer holds `Turn` and decides. The `default` arm of `Apply` is the
identity, which is what lets a later chapter add an event type and replay
this chapter's logs unchanged.

Which events will your agent add? A tool started but not returned, a
hint that arrived mid-turn, a model switch. All of them are additive, and
the reducer's `default` arm is what makes each one safe to add.

## 2.3 Parts, never strings; locators, never bytes

`Part` is an interface with one unexported method, so the set of parts is
closed at compile time. A renderer's type switch over `Part` is exhaustive
in the only sense that matters: a new part added to the package is a
compile-time gap in every switch that does not handle it, in every
renderer, at once. An open behavioral interface would make the log
unreplayable, because the log would have to serialize behavior. A sealed
union of data is replayable, and it is the only kind of interface that
appears in the log.

`BlobPart` carried a `Path string` in the first build. `Ref` replaced it
because a path expresses one of the four ways Gemini accepts a file (a local
file the agent uploads) and none of the other three (a File API `uri`, a
`gs://` object, an `https://` reference). `RefKind` names what the locator
is; `RefHandle` exists for Chapter 4, where a job's output may live in
memory under a handle rather than at a path. The old `"path"` key is still
present in the wire struct for one reason: to be recognized and refused. A
loader that coerces `"path"` to `RefPath` silently rewrites every remote
reference in a real log to a local file that never existed.

Never inline a blob. A log with a megabyte of base64 on one line is a log
you cannot grep, and a log you cannot grep is a log you cannot debug. The
log holds the locator and the MIME type; the bytes stay where they are.

Two opaque slots exist because two different things are opaque. An
`OpaquePart` is bound to the turn: a thinking block with its signature,
issued alongside the text, replayed as a part of the agent's entry.
`ToolCallPart.Opaque` is bound to one call: Gemini 3.x issues a
`thoughtSignature` as a sibling key of `functionCall`, and replaying the
call without it is a 400. A context with only the first slot has nowhere to
put the second, and the agent breaks in the chapter where tools execute
rather than here, so the grader tests it here.

## 2.4 Provenance is per model, recorded at write time

`Provenance{Vendor, Model, Surface}` is recorded on every `ResponseData`
and every `ToolCallPart` and `OpaquePart` when the parser writes them. It
is never reconstructed later, because it cannot be: a signature is valid
for the exact model and surface that issued it, and the log that does not
say which one is a log whose signatures are all suspect.

`Model` is a string. New model IDs ship weekly; the set is open, so code
compares it and never switches on it. `Vendor` and `Surface` are enums,
because the renderer must handle every case and a missing case should be a
compile error. `Surface` has two members for surfaces that were announced
after the three this chapter renders to, so the enum already carries the
lesson that a vendor's surface is a thing that changes under you. The
predicate that decides whether replay material goes back is equality on
all three fields, and the `seam-render` check probes it from both sides:
Anthropic material must reach the Anthropic render and must not reach the
other two.

Wire-verified on 12 September 2026: the Gemini `generateContent` surface
returns an error when a replayed `functionCall` is missing its signature
or carries a corrupt one; the Anthropic Messages surface silently drops a
thinking block it did not issue. The loud failure is the good one. A
silently dropped block is a model reasoning from less than it had, and
nothing tells you. The chapter's fixtures make the failure loud on every
vendor by grading the presence and the absence of `sig-exhibit-1`.

Enum zero is invalid on purpose. A `Provenance` that was never populated
decodes to `Vendor(0)`, and `SeamFor` returns an error for it. Had
`VendorAnthropic` been zero, an empty provenance would mean Anthropic, and
every part written by a parser that forgot to fill the field in would
replay to the wrong model with no error anywhere.

## 2.5 No field grows without bound

The invariant on `Context` is that nothing in it grows except with the
dialogue. The first design had a `Redacted map[Seq]bool` beside the
dialogue, consulted at render time to skip entries. It grew forever, it was
a second source of truth about what the dialogue contained, and a renderer
that forgot to consult it rendered the redacted bytes. It became
`Entry.Seq` plus `RedactedPart` as content: the entry says what it is now,
and no side table says otherwise.

`RedactData` names a span, `From` to `To` inclusive, and a level. The
stub is synthesized by the reducer from the content it supersedes, the
byte count and the reason, and is never stored. So redaction is
recoverable by construction: replay the log without the `Redacted` event
and the content is back, and the `Ref` carried into the stub says where the
bytes still are in the meantime. A redaction that stored its stub would be
a second copy of a fact the log already holds.

Three of the four levels are per-entry filters: N entries in the span, N
entries out, each with fewer parts. `RedactSummary` is a fold: the span
becomes one entry, with `Seq` equal to the span's `From` so the dialogue
stays in order without a sort, `Actor` equal to `System` because a summary
of three speakers is none of their speech, and `Replacement` stored on the
event because it is the one level whose new content a model wrote and
nobody can recompute. Written as a fourth case in the per-entry loop, where
it fits so tidily, the fold copies the summary into every entry in the span
and compaction grows the context it was called to shrink. The reference's
first version did exactly that.

Compact by category, never by position. Dropping the oldest N entries
drops whatever happened to be old, including the task statement in entry
one. Dropping tool results first, then tool calls, then prose, keeps the
goal and loses the bulk, and the four levels exist in that order for that
reason.

> **Receipt.** A production coding agent SDK shipped a `compress_context`
> tool that let the model choose what to forget. In one of Bill's sessions
> the model deleted about eighty percent of the context starting from
> message one, including the task it was working on. The framework owns the
> forgetting policy; the model is the one component with no memory of why
> the context was there.

## 2.6 The system prompt is output; ephemera are delivered once

`SystemPrompt` lives on `Config`. It is rendered into every request, as a
parameter for Anthropic, a `system` message for OpenAI, `systemInstruction`
for Gemini, and it never enters the log. A later chapter changes the prompt
on every turn; the log does not change with it, and a log from a year ago
replays under today's prompt. Storing the prompt in history would mean
every session carried a copy of a string the program already has.

`Context.Ephemera` holds parts that are pending. A `MessageReceived` from
`ActorSystem` appends to it and changes nothing else, no entry and no turn
transition. The renderer places the pending parts last in the request,
after every dialogue entry, and `RequestSent` clears them. So an ephemeral
part is in exactly one request, the first sent after it arrived, and in no
later one, and it is never in the dialogue. The grader plants a timestamp
and counts the requests that carry it. Two is a failure, and the failure
message says why: a stale timestamp is a lie about the present, sent on
every turn until the session ends.

What does your agent inject once? The clock. The screen. A memory recall
that matched this prompt and no other. A hint the human typed while the
model was working. All of them are ephemera, and none of them belong in
the log as dialogue.

## 2.7 Two functions permitted to lie

`Render(*Context, Config) (*http.Request, error)` and
`Parse(status int, body []byte) ([]Event, error)`. No vendor type crosses
in either direction: a context goes in and an HTTP request comes out, a
status and a body go in and events come out. Every belief the vendor holds
about the conversation, who said what and in which role, is manufactured in
`Render` and discarded when the request is sent.

Exhibit A is one `ToolReturned` rendered three ways. Anthropic puts the
`tool_result` block in a `user` message: the human said it. OpenAI invents a
`tool` role. Gemini puts a `functionResponse` part inside a `user` turn,
with the tool's `name` repeated because the wire wants it and the context
never stored it, so the renderer resolves the name from the matching call.
Authorship is a rendering decision. The log records `ActorTool`, which is
true, and the renderer says whatever the vendor needs to hear.

Exhibit B is the human's next instruction after a tool result. The
Anthropic renderer merges it into the same `user` message as the
`tool_result`, result first. The first draft of the chapter said this was
required because the Messages API rejects consecutive `user` messages; on
the wire, 12 September 2026, it merges them. The real 400s are ordering: a
`tool_result` must follow its `tool_use` in the very next message, and a
`text` block before a `tool_result` in the same message is rejected. The
renderer merges because the rule that a result must come first is easier
to keep in one message than across two.

A `BlobPart` to a vendor that has no wire shape for it is an error from
`Render` rather than a dropped part. Anthropic and OpenAI have no verified
remote-file part in this chapter, so both return an error naming the MIME
type and the model; a request that quietly lost its attachment is a request
whose reply is about a file the model never saw. That rule covers a blob in
a human entry. Inside a tool result no vendor has an attachment slot, so
the shared `resultText` flattens a blob to its MIME type and locator as
text on all three: the model is told where the file is, which is the most
any of the three surfaces can carry there.

Three renderers over the same `Entry` would grow three type switches, and
three type switches over `Part` is how three near-identical clients start.
`classify` runs once per entry before any vendor code and splits the parts
into texts, calls, results, blobs, opaque material, and surviving locators.
It also does the one thing no renderer would think to do: a redaction of a
tool result leaves its `RedactedPart` nested inside the `ToolResultPart`
that survived, and the `Ref` it carried forward is invisible to a renderer
that looks only at top-level parts. `classify` lifts it out, and the Gemini
renderer, the one with a remote-file part, turns it into `fileData`. Without
that, `ref-redaction` passes a field-by-field diff of the context and the
locator is gone from the one request that could carry it.

> **Receipt.** Bill's own agent, CodeRhapsody, carried an
> `AIClientInterface` of nineteen methods whose signatures named vendor
> types. The seam was in the right place and the types crossing it were
> not, and every vendor change leaked through it. Its replacement, as of
> September 2026, is a four-method `Provider` whose request builder is
> required to be pure: same input, byte-identical output, no clock, no
> network. The tell is in the signature. If a vendor's type appears in it,
> the seam is decorative.

The bet, stated so it can lose. The second renderer costs real work. The
third should be nearly free, and if it is not the seam is wrong, which is
worth finding out in an hour rather than after thirty thousand lines. The
order is fixed to make the test honest: Anthropic first as the baseline,
then OpenAI, which diverges enough to force an abstraction (a `tool` role,
a flat message list, `tool_calls` as an array, arguments as a string), then
Gemini, the alien one (`contents`, `parts`, `role: "model"`,
`systemInstruction` hoisted out, `functionCall` and `functionResponse`).
Hardest last on purpose. Put the alien one second and "the third was nearly
free" is true because the third was easy, which proves nothing about the
seam.

Which vendor does your agent talk to today, and what would the third one
cost you? If the answer is "one renderer and one parser," the seam held.

## 2.8 Four meters, no money

`Usage` has four integers, and the parser's job is to make them disjoint
before they reach the context. Anthropic reports them disjoint already.
OpenAI reports `cached_tokens` as a subset of `prompt_tokens`, so the
parser subtracts. Gemini reports `cachedContentTokenCount` as a subset of
`promptTokenCount` and reports `thoughtsTokenCount` separately from
`candidatesTokenCount`, so the parser subtracts on one side and adds on
the other. Gemini reports no cache-write count at all; the cost exists and
is billed by storage duration, and the honest meter is zero rather than an
invented number. The grader's arithmetic is in the TL;DR: a parser that
sums naively prints 362 where 112 is right, and nothing crashes.

Record counts, never money. Prices change without notice and differ by
model, by tier, and by whether the tokens were cached. A price table
belongs in the layer that displays cost, applied to counts the log
recorded, and a log that recorded dollars is a log that is wrong the day
the price moves.

Totals live on `Context.Usage`, accumulated by `Apply` on `ResponseEnded`.
The Chapter 1 client kept them on the HTTP client because that was where
responses arrived; here responses arrive as events, and the context is
the thing that folds events. A GUI that shows the bill reads the context.
Nothing asks the transport.

## 2.9 Replay is a byte comparison

`replay` renders the exhibit log twice, in two processes, for each vendor,
and compares the bytes. Three things put non-determinism into a renderer,
and each is a mutant the grader's own tests run against the reference: a
clock read during rendering; a Go map iterated into output, which the
runtime randomizes on purpose; and a random id. The reference's part
serializer is a struct with ordered fields, never a map, and `synthID`, the
id it mints for a call whose issuing vendor supplied none, is a function of
`Seq` and position. When the vendor did supply an id the reference passes
it through unchanged. A target vendor rejects a missing correlation id and
accepts a foreign-looking one, and rewriting an id that already matches its
result gains nothing.

`Render` returns an `*http.Request`, and `render` never sends it. The body
is recovered through `GetBody`, which `http.NewRequest` populates for a
`strings.Reader`, so the bytes `render` prints are the bytes `chat` would
send. A side table from request to body would recover them too, and would
be a field that grows forever, which §2.5 deletes.

On the six-event exhibit log the three renders are 1,118 bytes for
Anthropic, 808 for OpenAI, and 1,111 for Gemini, and each figure is
identical on every run. The sizes differ because the vendors differ; the
stability is the property.

## 2.10 What you would do differently

Load-bearing, and the grader fails you for it: the log as the only truth,
`Apply` total over every event, the two-method seam with no vendor type in
it, provenance written by the parser from the response, redaction as an
event, four disjoint meters, and a `Context` with no field that grows
except with the dialogue. Also load-bearing and ungraded: the log stays
grep-able. No inlined blob, no base64, one event per line.

The reference's taste, with its reasons in a sentence each. Four redaction
levels when one is graded: the other three are the compaction gradient a
later chapter needs, and adding a level later would mean a log format
change. Refusing an unknown `log_version`: a log from a newer binary may
carry events this one would silently drop. `AcceptsAudio` as a boolean:
the reference handles one media asymmetry and a bitmask is the obvious
generalization when it handles two. `SeamFor` as a switch: three cases
read better than a registry, and a fourth vendor is where the registry
starts to pay. No content hash on `Ref`: the reference trusts the locator,
and a hash beside it is the first thing to add when redacted content is
being recovered across machines.

CodeRhapsody's current client, as of September 2026, makes the same
structural choices with different names. Its vendor packages supply four
functions and may contain no control flow, no product policy, and no
mutable state; its request builder is required to be pure, and its history
holds hint blocks that a render step strips or places. Its tool results
are compacted the way this chapter's `RedactResult` works: the bytes are
replaced by a short stub that names the file on disk where the full output
still is. Where it differs is taste: it keeps its whole framework under
`internal/`, which Go forbids any other module from importing, and the
seam this chapter ships is meant to be imported.

Elements that exist now for a chapter that needs them, verified against the
reference:

| element | used here | exists for |
|---|---|---|
| `ToolCallPart`, `ToolResultPart` | rendering grader-supplied logs | the chapter where tools execute |
| `RefHandle` | survives `dump` | jobs whose output lives under a handle |
| `OpaquePart`, `ToolCallPart.Opaque` | replayed to the issuing model | every chapter; never interpreted |
| `TurnState.ToolsPending` | reached at the end of every graded session | the tool loop |
| `RedactedPart` | fully exercised | the compaction gradient |
| `RedactTool`, `RedactDialogue`, `RedactSummary` | implemented, ungraded | the compaction gradient |
| `RedactData.From`, `To` as a span | spans of one | summaries over many entries |
| `SurfaceInteractions`, `SurfaceResponses` | named, never rendered | the day a surface is retired |

Your pain points decide the rest. If your current agent loses the thread
after compaction, the redaction levels and their order are where to start.
If it cannot switch models mid-session without a restart, provenance is
where to start. If its bill is a surprise, the four meters are.

## 2.11 Exercise

Brief for your agent: rewrite the Chapter 1 client as `ch02` in a directory
of its own inside the course module, standard library only. Store an
append-only event log at `CH02_LOG`, derive the context by folding it,
render to Anthropic, OpenAI, and Gemini through a two-method seam, and
parse each vendor's response into events with provenance and normalized
usage. Add `dump` and `render LOG`. Then run the grader.

```sh
make grade-dir CH=2 DIR=path/to/yours    # your client
make grade2                              # the reference
```

| check | points | passes when |
|---|---|---|
| `session` | 0 | every stdout line is `agent`, `ack`, or `usage`; the ephemeral directive is acknowledged |
| `ch1parity` | 10 | all seven Chapter 1 checks pass against the new binary |
| `logdump` | 5 | `dump` prints valid JSON-lines in ascending `seq`, and a fresh process renders it |
| `replay` | 10 | two renders of the exhibit log are byte-identical, all three vendors |
| `redaction` | 10 | the redacted result's bytes are absent, the call is present, and the un-redacted render carried the bytes |
| `ephemera` | 10 | the planted text is in request one and no other |
| `usage` | 10 | `{112, 0, 250, 38}` after two rounds and the cache-write probe's totals, all three vendors |
| `seam-render` | 15 | each vendor's request has the shape rule 4 lists; `sig-exhibit-1` reaches Anthropic only; Gemini replays `thoughtSignature` on its own call |
| `seam-parse` | 15 | the three dumped dialogues agree on every text, tool name, and decoded argument; provenance present and distinct |
| `ref-roundtrip` | 3 | all three `RefKind`s survive `dump` |
| `ref-oldformat` | 2 | a blob with the old `"path"` key is refused: non-zero exit, nothing on stdout, a diagnostic on stderr |
| `ref-zerokind` | 2 | a `Ref` with `kind` 0 is refused the same way |
| `ref-render` | 4 | a `RefURI` blob renders to Gemini as `fileData` with the locator in `fileUri` |
| `ref-redaction` | 4 | after redaction the locator is still in the Gemini request and the secret is gone; without redaction the secret is present |

Weights sum to 100. Every check is all or nothing. Exit status of the
grader: 0 pass, 1 fail, 2 could not run. Cost: nothing. The fake serves all
three vendors on one local port, routed by path.

The grader's own tests work by deletion, as in Chapter 1: twenty-two
mutants of the reference, each asserting the exact set of checks that
fails. The one worth telling: deleting the only use of
`ToolCallPart.Opaque` once scored 100 out of 100, because the exhibit log's
tool call carried no per-call material and its provenance was Anthropic,
so a correct Gemini renderer withheld it by design and the assertion passed
for nothing. The Gemini-authored replay log exists because that mutant was
run.

## 2.12 Drive it yourself

Export a real key and a model ID from the vendor's models endpoint, set
`LLM_VENDOR`, and run `./ch02 chat`. Then run `./ch02 dump` and read the
log. Two things to look for. The `model` string in each `response_ended`
event's `from` is the one the vendor returned, and it may differ from the
one you asked for; that difference is the reason provenance is read from
the response. And the four meters on each response are the vendor's
accounting after normalization; compare them to the vendor's dashboard at
the end of the day.

Then set `LLM_VENDOR` to a second vendor and run `./ch02 render` on the
same log. The request that comes out is the conversation you had with one
model, rendered for another, with the first model's opaque material
withheld. Your agent will outlive at least one of these three surfaces.
When that happens, count the files that change. If the count is anything
other than one renderer and one parser, the seam leaked.
