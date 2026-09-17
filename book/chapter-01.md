# Chapter 1: A Conversation, the Obvious Way

If you already know how to write a basic chatbot that talks directly to the
Anthropic API, skip to Chapter 2, where we begin to blow your mind.

The Anthropic Messages API keeps nothing between calls. Every request carries
the entire conversation, and the reply is a function of that request alone.
So a chat client has exactly one piece of state, a slice of messages, and
"memory" is what happens when nobody deletes from it.

The reference is 285 lines of Go: two types, one HTTP call, two front ends,
and the standard library. No SDK and no framework, here or anywhere in the
book; §1.1 argues why. The short form: a framework decides what goes into
the request payload and where, and the payload is where every later chapter's
leverage lives.

**What you build.** A history, a wire format, a client, and one function that
ties them together.

```go
// Message is one turn. Role is "user" or "assistant"; the API requires that
// they strictly alternate.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Conversation is the entire history. This is the whole state of the chat —
// which is the deep fact of this chapter: the API is stateless, so the
// conversation lives in this slice or it lives nowhere.
type Conversation []Message
```

```go
type request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
}

type response struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}
```

`request.Messages` has the same type as the history. The request body is the
conversation plus three scalars: the model ID, `max_tokens` (the reference
sends 1024), and one constant system prompt.

```go
// Client holds the connection details and the running bill.
type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client

	// Usage is money. Accumulated from the first request onward.
	InputTokens  int
	OutputTokens int
}
```

```go
// Ask is the loop of the whole chapter: append the question, send everything,
// append the answer, hand it back.
func (c *Client) Ask(conv *Conversation, question string) (string, error) {
	*conv = append(*conv, Message{Role: "user", Content: question})
	reply, err := c.Send(*conv)
	if err != nil {
		return "", err
	}
	*conv = append(*conv, Message{Role: "assistant", Content: reply})
	return reply, nil
}
```

`Send` marshals a `request` from the whole slice, POSTs it, fails on any
non-200 status with the status and body in the error, adds the response's
`usage` to the client's totals, and returns the concatenation of every
`content` block whose `type` is `text`. Nothing else is in it: no retry, no
backoff, no streaming.

**Rules.** The grader fails you on each.

1. **Wire.** POST `$ANTHROPIC_BASE_URL/v1/messages` with headers
   `x-api-key`, `anthropic-version: 2023-06-01`, and
   `content-type: application/json`. The body has a non-empty `system`, a
   positive `max_tokens`, and no `stream`. Messages strictly alternate
   `user` and `assistant`, none is empty, and the first and last are `user`.
2. **Three variables, hardcode none.** `ANTHROPIC_API_KEY`;
   `ANTHROPIC_BASE_URL`, defaulting to `https://api.anthropic.com`;
   `ANTHROPIC_MODEL`, with no default. The grader supplies its own values
   for all three and checks that they arrived on the wire. A model ID
   remembered from training data fails.
3. **Send everything, every time, append-only.** Each request's messages are
   the previous request's messages, byte for byte, plus the new turns. Never
   fewer, never rewritten.
4. **The reply goes into the history** as an `assistant` message. The
   request for round four still carries the server's reply from round one.
5. **Walk every content block.** The answer is every `text` block
   concatenated in order with nothing between them, never `content[0].text`
   alone. The grader's fake splits every reply into two blocks.
6. **One API call per round.** No warm-up call, no retry, no second call.
7. **Protocol and bill.** With no arguments the binary is in grader mode: it
   reads one JSON object `{"user": "..."}` per line on stdin, and for each
   one writes exactly one line `{"agent": "..."}` on stdout. The exchange
   is lockstep: the grader writes a line, waits for your reply line, then
   writes the next, so write and flush each reply before reading again.
   Everything else goes to stderr. When stdin closes it writes
   `{"usage": {"input": N, "output": M}}`, where N and M are the sums of
   every response's `usage.input_tokens` and `usage.output_tokens`, and
   exits 0.

Seven checks, all or nothing each, sum 100: `wire` 15 (rules 1 and 2),
`growth` 15 (rule 3), `memory` 25 (rule 4), `replies` 10 (rule 5), `calls`
10 (rule 6), `protocol` 15 and `usage` 10 (rule 7).

**Yours.** The seven rules are structural; the grader enforces them and
nothing else. Also structural and ungraded: the HTTP client is the only
dependency, because Chapters 2 onward edit request bytes an SDK does not let
you reach. Taste, yours to change: where the running token totals live (the
reference hangs them on the `Client`; a session or the conversation itself
are both defensible); whether an unset `ANTHROPIC_MODEL` is fatal (the
reference refuses to start and names `GET /v1/models`); the 120 s HTTP
timeout; the 8 MB line buffer; the `type == "text"` filter, which changes
nothing on today's wire; and the entire interactive mode (`./ch01 chat`),
which the grader never runs. There is no retry in the reference. A 429 kills
it. That is the honest state of a naive client, and it is the first pain
point to fix in yours.

**Exercise.** Build the two-mode client in one directory inside this module,
then `make grade-dir CH=1 DIR=path/to/yours`. It must print `100/100`. The
grader compiles your directory, runs the binary against a fake server, and
never reads your source. `make grade` runs the same grader on the reference.

## 1.1 Frameworks, and why not

The rule has two sides. If you want an agent, use a framework. Storage,
retries, tool plumbing, model discovery, and a year of other people's bug
fixes arrive on day one, and for most agents that is the right trade. If you
want an advanced coding agent, you are on the bleeding edge, or it is not
advanced. The bleeding edge is what nobody has anticipated yet, and a
framework is a record of what its authors anticipated.

The mechanism is specific. Frameworks are generous with storage and
discovery and hard-code delivery: what goes into the request payload, in
what order, at what position. Delivery is where the leverage lives. Three
capabilities, all real at the raw Messages API surface as of September 2026,
that a framework either cannot express or buries:

1. **Mid-turn steering.** The API accepts a `tool_result` block and a user
   `text` block in the same message. That is how a human redirects an agent
   between tool calls without breaking the chain. Products expose it as a
   feature; their SDKs do not let you place the block.
2. **Ephemeral placement.** Volatile data (the clock, screen state, live
   status) goes last in the payload, one copy, never in history, because
   position is the feature. One timestamp that moves destroys prefix
   caching. A production agent went from a 0% to a 98% cache hit rate by
   moving it.
3. **Cache breakpoints.** You know which suffix of your context is volatile;
   the provider does not, and neither does a framework that assembles
   requests for you. Owning the bytes is owning the cache economics.

The ruling for the course: the Messages API and `net/http`. Every request
byte in this book is one you put there. The Chapter 1 client needs none of
the three capabilities, which is why frameworks feel fine on day one and
why this chapter is where the habit has to start.

Which of the three does the product you use today expose, and which does
its API let you reach?

> **Receipt.** Bill has shipped several agents on frameworks and would do so
> again for anything that is not a coding agent. The rule is two-sided
> because he has stood on both sides of it.

## 1.2 The request is the conversation

There is no client-side model of "a chat" beyond `[]Message`. The request
type has four fields, and the fourth is the history itself. The other three
are scalars: the model ID, `max_tokens` (1024 in the reference; a cap on the
reply, and the API refuses a request without one), and `system`, one
constant string that rides outside the messages array.

The API is stateless. It does not know that the request it received a second
ago came from the same program. So the whole chat is a JSON array you can
print. This is the body of the grader's round-two request, which is the
round-one request with two messages appended:

```json
{
  "model": "<the value of ANTHROPIC_MODEL>",
  "max_tokens": 1024,
  "system": "You are a chat agent built from raw HTTP calls in Chapter 1 of Building Advanced AI Coding Agents. Answer briefly.",
  "messages": [
    {"role": "user", "content": "Hello! I'm starting a new project. Give it a codename and remember it."},
    {"role": "assistant", "content": "Welcome aboard. Your project's codename is TANAGER-4417, and I'll remember it."},
    {"role": "user", "content": "Good. What is the capital of France?"}
  ]
}
```

A wrapper struct around the slice buys nothing here. The slice is the
design.

## 1.3 Two roles, and who enforces the alternation

`Role` is a plain string with two legal values. No Go enum, because nothing
in this program switches on it; the client writes `"user"` before a send and
`"assistant"` after, and the only reader is the API.

Three wire rules apply to the array: roles strictly alternate, no message
has empty content, and the first and last messages are `user`. The grader's
fake enforces all three and names the offending index in its violation
record (`messages[2] and [3] are both "user"`, `messages[1] has empty
content`). The live API, checked on the wire on 12 September 2026, is more
lenient than the fake: it merges consecutive same-role messages into one,
and it accepts a trailing `assistant` message as a prefill for the model to
continue. Your fake is stricter than the vendor on purpose. A vendor that
quietly repairs a malformed array hides the bug that produced it, and the
bug that produces two `user` turns in a row is usually the bug that also
lost a reply.

## 1.4 Reading a reply: walk the blocks

The answer is the concatenation of every `text` block in `content`, joined
with nothing between them. `content` is an array of typed blocks, and
reading `content[0].text` is the day-one stumble: it compiles, it runs, and
it returns the first block, which is the whole answer until the day it is
not.

The grader's fake makes the walk load-bearing. It cuts every reply at the
first space past the first third of the string and serves two `text` blocks
that concatenate to the original. A client that reads only the first block
returns half a sentence, and the `replies` check compares your stdout line
to the full scripted reply.

The reference filters on `block.Type == "text"`. On today's wire a plain
text request gets back only text blocks, so the filter changes nothing and
the grader does not test it. It is there because `content` is typed, and a
reader who copies the loop into a program that does more than chat will be
glad the filter was already in place.

> **Receipt.** The first version of the grader's fake served every reply as
> one block. A reference solution with the block walk deleted, reading
> `content[0].text`, scored 100 out of 100. Nobody found this by reading the
> grader; it was found by deleting the walk from the reference and asking
> what still passed. The fake now serves two blocks, and that mutant fails
> `replies` and `memory` both.

## 1.5 Where memory lives

Memory is the slice. `Ask` appends the question, sends the whole slice,
appends the reply as an `assistant` turn, and returns. Nothing else in the
program remembers anything.

The `memory` check, 25 points, rests on a planted string. In round one the
fake's scripted reply contains a codename, `TANAGER-4417`. The program never
types that string; only the server does. In round four the grader asks for
the codename, then finds the recorded request whose last message is round
four's question and requires two things of it: seven messages (four `user`,
three `assistant`), and the codename somewhere in the body. A program that
starts a fresh conversation each round, or that forwards only the user's
turns, cannot have the codename in that request, because it never held it.

The `growth` check, 15 points, walks the five recorded requests in order
and compares each one to its predecessor, position by position, on role and
text. Every request must begin with the previous request, unchanged, and be
longer.

Forty of the hundred points test whether an array grew append-only. This
section has no "yours." The array is the whole design, and the decision
to never delete from it is the one this chapter makes without qualification.

## 1.6 What it costs

Every response carries `usage.input_tokens` and `usage.output_tokens`. The
reference adds both to the `Client` on every `Send` and prints the totals
when stdin closes. In interactive mode it prints them to stderr after every
reply, next to the turn count.

The fake's tokenizer is one token per four characters, applied to the system
prompt and to each message's text separately, each rounded up with a minimum
of one, so the numbers reproduce on any machine. Across the grader's five
rounds the input count per request is 47, 76, 93, 114, and 136 tokens;
totals are 466 in and 65 out. The fifth question costs 2.9 times the first,
and the fifth question is six words long. Every request pays again for every
earlier turn, because every request carries every earlier turn.

Where the totals live is taste. The reference hangs them on the HTTP client
because that is where the responses arrive. A session object or the
conversation itself are both defensible homes. The curve is not taste. It is
the first design pressure a chat client feels, and this chapter states it
and leaves it there.

## 1.7 Three variables, hardcode none

The client reads `ANTHROPIC_API_KEY`, `ANTHROPIC_BASE_URL`, and
`ANTHROPIC_MODEL`. The first has no default. The second defaults to
`https://api.anthropic.com`, with a trailing slash trimmed. The third has no
default, and its absence is fatal: the reference exits 1 with nothing on
stdout and one line on stderr that names `GET /v1/models` as the place the
answer lives.

The grader's harness exports its own values for all three, plus
`ANTHROPIC_API_URL` and `LLM_MODEL` carrying the same URL and model for
clients that spell the variables that way. The fake then requires the key
header and the `model` field to equal what the harness exported. A client
that reads the variables and passes them through cannot fail this. A client
with a model ID typed into the source fails `wire` on every request.

No default model ID exists in the reference because any ID written there
is a soft hardcode: correct the day it is typed, wrong some later day, and
invisible to a grader that always sets the variable. The endpoint that
knows what exists today is one request away:

```sh
curl -s https://api.anthropic.com/v1/models \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01"
```

> **Receipt.** While building this chapter's grader, a coding agent
> hardcoded a model ID from its training data into a fixture. The ID had
> already been retired. Training data has a date on it; the endpoint does
> not.

## 1.8 Two modes, one core

`Ask(*Conversation, string)` is the only function both front ends call.
Grader mode and chat mode differ in framing and in nothing else.

Grader mode reads stdin with a `bufio.Scanner` whose buffer is raised from
the 64 KB default to 8 MB, skips blank lines, and treats a line that does
not parse as `{"user": ...}` as fatal. Replies go out through a
`json.Encoder` on stdout, which supplies the trailing newline and the
escaping, so a reply containing a quote or a newline is still one line.
After the scanner reports end of input the program encodes the usage object
and returns. Any error, including a non-200 status, is printed to stderr
with a `ch01: ` prefix and the process exits 1.

Chat mode is the same loop with a prompt. It prints `you> `, reads a line,
calls `Ask`, prints the reply, and prints the turn count and running totals
to stderr. It is not graded. Make it yours: the reference's prompt strings,
its choice to print errors and continue rather than exit, and its lack of
any history command are all placeholders for what you want from a client
you talk to directly.

## 1.9 Graded by observation only

The grader compiles your directory with `go build`, starts a fake Messages
API server on a local port, exports the environment, runs your binary in
grader mode, and writes the five scripted lines to its stdin one at a time,
waiting up to 20 seconds for each reply line. After the fifth reply it
closes stdin and waits up to 15 seconds for exit. It then scores seven
checks from four sources: your stdout lines, your stderr, your exit status,
and the fake's record of every HTTP request it received. It never reads a
line of your source.

The fake is forgiving at the transport and strict in the record. Every
request gets a 200 and a scripted reply, even a malformed one, so a client
with three bugs shows all three in one run instead of dying on the first.
Each violation is appended to that request's record, and the `wire` check
reads the records.

The grader's own tests work by deletion. Thirteen mutants of the reference,
each with one defect (the block walk removed, the reply never appended,
usage fabricated, a second API call per round), are graded, and each test
asserts the exact set of check IDs that fail, not merely that the score
dropped. A fourteenth case, the unmodified reference, asserts the empty
set. A check that no mutant can make fail is a check that measures nothing,
and the two-block fake in §1.4 exists because that test found one.

This is how to test a program that talks to a model without a model. The
scripted fake, the recorded requests, and the deletion audit are the parts
of this chapter that survive unchanged in your agent's own test suite.

## 1.10 What you would do differently

The load-bearing walls are the seven rules. Each one is a check the grader
runs, and a client that breaks any of them is a client that lost a reply,
paid for a call it did not need, or lied about its bill. Also load-bearing,
and ungraded: no SDK. The grader cannot tell an SDK from `net/http`, but
every later chapter edits request bytes an SDK does not expose, and a
client built on one has to be rebuilt.

Everything else is the reference's taste, and the reference's reasons are
short. Token totals on the `Client`: that is where responses arrive. Fatal
on a missing model ID: the alternative is a default that rots. A 120 s HTTP
timeout: long enough for a slow reply on a large context, short enough that
a hung connection is noticed. An 8 MB scanner buffer: a pasted file should
not crash a chat client. The `type == "text"` filter: `content` is typed.
No retry on 429: a naive client should fail loudly on the wall it will hit
first, so that the fix is visible work rather than a hidden loop.

CodeRhapsody's production client shows where two of these choices go once
a client grows up. It does retry, but the decision is made a layer above
the HTTP call: its Anthropic provider classifies each failed response into a
disposition (retry, fatal, downgrade capabilities) and the transport
re-sends on retry. The policy exists; the HTTP client still does not own
it. On model IDs it is less strict than the reference. A capability table
in its source names the models it knows, and one stale entry in that table
produced 404s in September 2026. The reference's refusal to name any model
is the version of that lesson with no table to go stale.

Your pain points decide the rest. If your current agent loses your place
when you paste a large file, the buffer size is where to start. If it
silently swallows rate limits, the 429 path is where to start. Both are
yours.

## 1.11 Exercise

Brief for your agent: build the client described on the first page of this
chapter as a single Go `package main` in a directory of its own inside the
course module, standard library only, both modes. Then run the grader.

```sh
make grade-dir CH=1 DIR=path/to/yours    # your client
make grade                               # the reference
```

| check | points | passes when |
|---|---|---|
| `protocol` | 15 | one `{"agent": ...}` line per input line, a `{"usage": ...}` line after EOF, nothing else on stdout, exit 0 |
| `wire` | 15 | every request is a well-formed Messages API call with the exported key and model, non-empty `system`, positive `max_tokens`, no `stream`, alternating non-empty roles, `user` first and last |
| `calls` | 10 | exactly one request per round |
| `replies` | 10 | each stdout reply equals the scripted reply, after trimming |
| `memory` | 25 | round four's request has seven messages and contains `TANAGER-4417` |
| `growth` | 15 | each request begins with the previous request's messages, unchanged, and is longer |
| `usage` | 10 | reported totals equal the fake's totals exactly, both nonzero |

Weights sum to 100. Every check is all or nothing. Exit status of the
grader: 0 pass, 1 fail, 2 could not run (build failed, binary did not
start, or a round timed out). Cost: nothing. The fake serves every
request.

To talk to it live, export a real key and a model ID from `GET /v1/models`
and run `./ch01 chat`. Live token counts depend on the model's tokenizer
and the day's system prompt; none are printed here.

## 1.12 API access: the toll booth

Cheap keys exist. Fast ones do not. Coding with an agent needs sustained
throughput, and the major providers gate throughput behind account tiers.
In 2025 the Anthropic API sold Tier 4, the first tier with limits high
enough to run a coding agent, for a $400 prepayment; as of 15 September
2026 the published ladder is Evaluation, Start ($500 per month), Build
($1,000), and Scale ($200,000), with promotion "automatically based on
usage history," and new organizations start below the standard limits
"while account history is established." The number changes. The wall does
not. It does not bite in this course, where every graded request hits a
local fake. It bites the day a finished agent meets real work.

> **Receipt.** Bill, on the 2025 ladder: "You have to reach Tier 4 or the
> rate limits are so low you can't code."

The economics behind the wall are the economics of the tool layer. Raw
model access is a melting asset; every advance is followed within months by
cheaper distilled competitors. The providers' answer is to own the tools on
top, and to price the same tokens lower through their agent than through
your key. The message is: use ours, do not build your own. This book
exists to ignore that message, and the toll is the price of ignoring it.

The course proxy is optional. Fund a modest amount on the course site and
point `ANTHROPIC_BASE_URL` at the course URL; it proxies to the vendors with
per-student budgets, no provider account, no tier wall, no waiting. Cost is
pass-through, with no profit on tokens, as the preface states. With your
own key you do not need it; the course runs identically pointed at the
vendor.

The Chapter 1 client already reads `ANTHROPIC_BASE_URL`, because the
grader's fake needs it. Fake for grading, proxy for live chat, vendor for
production: one variable, and the code never knows which it is talking to.
