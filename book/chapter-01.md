# Chapter 1: A Conversation, the Obvious Way

## 1.0 Sixty billion dollars

Three transactions. They are not three versions of one story. Read
in sequence, they are a market getting steadily more precise about what it is
buying.

**July 2025.** Google paid $2.4 billion for a non-exclusive license to some of
Windsurf's technology and to hire its chief executive, a co-founder, and part
of its research team into DeepMind. It did not buy Windsurf. The company kept
its product, its customers and its revenue, and Cognition bought what remained
a few weeks later. OpenAI had tried to buy the whole company for $3 billion;
that deal collapsed over intellectual property terms.

Look at what was declined. The product: declined. The brand: declined. The
customers and the revenue: declined. What $2.4 billion bought was the people
who knew how to build an AI coding agent, plus permission to read how they had
done it. The trade press called it a reverse acqui-hire. It is more useful to
call it what it was: a price tag on knowledge, with the company carefully
removed.

**October 2022.** Elon Musk completed the purchase of Twitter for $44 billion.
It is here for one reason, and the reason arrives in two paragraphs.

**June 2026.** SpaceX agreed to acquire Anysphere, maker of Cursor, for $60
billion in stock. The merger filing is public. Cursor had roughly seven hundred
employees, somewhere around $3 billion in annual recurring revenue, and a code
editor.

Sixteen billion dollars more than the global town square, for a
seven-hundred-person company that makes a code editor.

That is rational, and the buyer explained why in April, in public. SpaceX said
that combining "Cursor's leading product and distribution to expert software
engineers" with its "million H100 equivalent Colossus training supercomputer"
would help it "build useful models."

Read that as an equation. Expert engineers plus training compute produces
better models. The coding agent is not the product being bought. It is an
instrument in the training loop, and what it collects is the most valuable
telemetry in the industry: thousands of expert engineers accepting, rejecting
and correcting machine-written code, all day, on real problems, with a verdict
attached to every suggestion.

The arithmetic is public too. In the same April statement, SpaceX
said it could acquire Cursor for $60 billion, or pay roughly $10 billion for
the two companies to work together. It paid six times as much. If you wanted the
product, $10 billion bought the product. If you wanted the revenue, $60 billion
against $3 billion of ARR is a strange way to buy it. The extra fifty billion
bought ownership of the loop.

**Why coding goes first.** Every knowledge profession is a candidate for
automation, and software engineering is being automated first. Not because it
is the most valuable, and not because it is the easiest. Because its outcomes
are measurable. Tests pass or they don't. The code compiles or it doesn't. A
measurable outcome is a reward signal, and a reward signal is the one thing
reinforcement learning cannot proceed without. Law, medicine and management all
have to argue about whether the work was any good. Software just runs it.

The obvious objection is that the interesting part of software is exactly the
part that can't be scored: architecture, judgment, taste. It does not change
what was bought; SpaceX paid whether or not taste is scorable. I have a
proposal that answers it anyway, "Training Superhuman Software Architects"
(`coderhapsody.ai/docs/superhuman-architecture`): architecture quality is
observable against a sequence of requirements arriving over time, and public
git histories already contain millions of such sequences. The training signal
for judgment isn't hypothetical. It's sitting in public repositories.

So: coding is measurable, therefore coding is first, therefore the coding agent
sits inside the loop that improves the model. That is what $60 billion bought.

**What the owner of the loop decides.** If the coding agent is in the training
loop, whoever owns the loop shapes what the models become. That is a market
question, and it is also a safety question, and there is a dated public
example. In July 2025, after a tuning change intended to make it less
politically filtered, xAI's Grok posted antisemitic content on X and referred
to itself as "MechaHitler." xAI apologized on 12 July 2025 and attributed the
behavior to the update. That is the whole observation: the values of a model
are downstream of whoever controls its training, demonstrated once, in public,
at scale.

**What that leaves for you.** Not "the code is cheap, therefore the price is
absurd." Nobody paid $60 billion for source code. The honest version is smaller
and much harder to argue with: the knowledge of how to build one was priced at
$2.4 billion in 2025, knowledge is teachable, and this book teaches it. The
thing Google would not buy a company to obtain is the thing you are holding.

Building your own does not dent anyone's valuation, and I won't pretend
otherwise. What it does is remove you from the measurement. Your code stays on
your machine. Your accept-and-reject signal trains nobody. You can change
vendors in an afternoon, or run a local model the day one is good enough. That
is sovereignty for one engineer, which is a modest claim, and it is the one
that will still be true in five years.

One more thing, said once: the asset being purchased at these prices is expert
software engineers who never built their own tools.

**The bet.** The Windsurf deal is what broke me. Billions for *that*? I told my
team I could write a better AI coding agent proof of concept than Windsurf in
two weeks. My manager told me to prove it. I did. The result was StackAgent: a
working proof of concept, vibe-coded, that I threw away in its entirety. The
two weeks were not wasted. The value was never the code; it was what I learned
building it. I then spent a month writing the first version of CodeRhapsody
properly, and that is the direct ancestor of the agent helping me write this
book.

Hold on to the shape of that. The code was disposable; the learning was the
asset. It is the same lesson the market paid $2.4 billion for.

## 1.1 Frameworks, and why this book uses none

Two rules, both honest.

If you want to build an agent, use a framework. That is the right call for most
agents, and I have shipped several that way. You get storage, retries, tool
plumbing and model discovery for free, and for a simple agent that is most of
the work.

An advanced AI *coding* agent is on the bleeding edge, or it isn't advanced. A
framework encodes what its authors anticipated you would need. The bleeding
edge is precisely what nobody anticipated yet.

Here is the mechanism. Frameworks are generous with *storage* and *discovery*,
and they hard-code *delivery*: what goes into the request payload, in what
order, at what position. Delivery is where the leverage lives. Three
capabilities, all real at the raw API surface today, that a framework either
cannot express or buries:

1. **Mid-turn steering.** The Messages API accepts a `tool_result` block and a
   user `text` block in the same message. That is how a human redirects an
   agent between tool calls without breaking the tool chain. Frontier products
   have validated the idea with real-time steering in their UIs, and they
   expose it as a product toggle, not as anything their frameworks let you
   reach.

2. **Ephemeral context placement.** Volatile data (the time, the screen state,
   live status) goes *last* in the payload, exactly one copy, never in history.
   Position is the feature. One wandering timestamp in the wrong place destroys
   prefix caching; I watched a production system go from 0% to 98% cache hits
   by moving one. Frameworks decide placement for you.

3. **Cache breakpoint control.** You know which suffix of your context is
   volatile. The provider doesn't, and neither does a framework assembling
   requests on your behalf. Owning the request bytes is owning your cache
   economics.

So this book starts with the Anthropic API and an HTTP client. No SDK, no
framework, ever. Every request byte in this book is one you put there.

None of that matters yet. The program in this chapter needs none of those three
capabilities, which is exactly why frameworks feel fine on day one.

## 1.2 Anatomy of a request

One endpoint, `POST https://api.anthropic.com/v1/messages`, and three headers:

```
x-api-key: $ANTHROPIC_API_KEY
anthropic-version: 2023-06-01
content-type: application/json
```

The body:

```json
{
  "model": "claude-sonnet-5",
  "max_tokens": 1024,
  "system": "You are a helpful assistant. Answer briefly.",
  "messages": [
    {"role": "user",      "content": "What is the tallest mountain in Africa?"},
    {"role": "assistant", "content": "Kilimanjaro, at 5,895 metres."},
    {"role": "user",      "content": "And the second tallest?"}
  ]
}
```

`model` names the model. `max_tokens` is required: the API refuses to guess how
much output you can afford. `system` is one string that rides outside the
messages array, and for this chapter it is a single fixed line.

`messages` carries three rules, and the grader enforces all of them. Roles
strictly alternate, `user`, `assistant`, `user`. The conversation begins with
a `user` message. The last message is the user's, because otherwise there is
nothing to answer. And no message's content is ever empty.

The response:

```json
{
  "id": "msg_01XFDUDYJgAACzvnptvVoYEL",
  "type": "message",
  "role": "assistant",
  "model": "claude-sonnet-5",
  "content": [
    {"type": "text", "text": "Mount Kenya, at 5,199 metres."}
  ],
  "stop_reason": "end_turn",
  "stop_sequence": null,
  "usage": {"input_tokens": 52, "output_tokens": 14}
}
```

Three fields matter: `content`, `stop_reason`, `usage`. `stop_reason` is
`end_turn` on every reply in this chapter. Chapter 3 is where it first says
something else, and that turns out to be the same moment `content` stops being
simple.

**The asymmetry that catches everyone once.** The request lets you send
`content` as a bare string. The response never does: response `content` is
always a list of typed blocks. Walk the list and concatenate the text out of
every block. The field has the same name on both sides and a different shape,
and that is the classic day-one stumble.

`content[0].text` will be right for a while, which is what makes it a stumble
and not an error. Every reply in this chapter arrives as text, so indexing and
walking return the same string, and they keep agreeing right up until a reply
arrives carrying something that is not text. That happens in Chapter 3, the
first time a model asks to call a tool, and by then the line that reads
position zero will be old code that you trust. Live, that is how it goes. The
grader's fake does not wait for Chapter 3: it splits every reply across two
blocks, so the shortcut fails on day one, where the failure is cheap.

### Sidebar: ask the API which models exist

Do not take a model ID from a blog post, a tutorial, or your own memory. Ask:

```bash
curl -s https://api.anthropic.com/v1/models \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01"
```

An ID that *looks* current may be an alias that silently resolves to something
much older, and nothing in the response will tell you so. This advice is
earned. While building this book's grading rig, the coding agent picked a model
ID from its own memory because it looked familiar. It worked. It was a year
old. A model's memory is its training data, and training data has a date. Ask;
don't remember.

Every model ID printed in this book will age, including the one in the request
above. The sidebar teaches the lookup, not the answer.

## 1.3 The obvious data structure

```go
// Message is one turn. Role is "user" or "assistant"; the API requires that
// they strictly alternate.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Conversation is the entire history.
type Conversation []Message
```

Straight from the docs. Every framework on earth is a wrapper around this
shape. It fits in your head, `json.Marshal` serializes it without help, and it
will carry this chapter comfortably.

## 1.4 The loop

Append the user's message. POST the entire conversation. Parse the reply.
Append it. Repeat.

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

The one deep fact of the chapter: **the API is stateless.** The provider
retains nothing between calls; every request replays the whole history. The
conversation lives in your process or it lives nowhere. That slice is the
entire state of the chat, and "entire" is not a figure of speech.

`Send` is §1.2 made executable. Marshal the request, set the three headers,
POST, read the body, refuse anything but a 200, then walk the blocks:

```go
	var parsed response
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	c.InputTokens += parsed.Usage.InputTokens
	c.OutputTokens += parsed.Usage.OutputTokens

	var text strings.Builder
	for _, block := range parsed.Content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}
	return text.String(), nil
```

There is no retry here, and no backoff. If the API returns 429, this program
dies. That is the honest state of a naive client, and we are not going to paper
over it.

The whole thing is 273 lines in `solutions/ch01/main.go`, and about a third of
that is the two front ends described below.

## 1.5 Usage is money

Read `usage` on every response. Keep cumulative input and output totals from
the very first request. Then watch the input count climb every round, because
the history you resend gets longer every round.

Five rounds of the exercise against the grader's fake server, input tokens per
round:

| round | 1 | 2 | 3 | 4 | 5 |
|---|---|---|---|---|---|
| input tokens | 49 | 78 | 95 | 116 | 138 |

Cumulative: 476 input, 65 output. Nothing in those five rounds got longer
except the history. The fifth question costs nearly three times the first, and
it is the same size question. That is the whole economics of a stateless API in
one row of numbers.

Those figures regenerate identically on any machine, because the fake's token
counter is deterministic; `make grade` prints them. Live against
`claude-sonnet-5`, a three-round run cost 631 input and 388 output tokens.

No caching and no remedies here. Just the habit, and the curve.

## 1.6 Chat with it

The reference has exactly one fixed line to say about itself, and it says it
in the system prompt:

```go
const systemPrompt = "You are a helpful assistant built from raw HTTP calls in Chapter 1 of Building Advanced AI Coding Agents. Answer briefly."
```

Everything the model knows about where it is running is in that string.

```
$ go build -o ch01 . && ./ch01 chat
talking to claude-sonnet-5 at https://api.anthropic.com — Ctrl-D to quit

you> What are you running on?

claude> I'm running on infrastructure built from raw HTTP calls, as described
in Chapter 1 of "Building Advanced AI Coding Agents." Rather than relying on
a heavyweight SDK or framework, my underlying implementation makes direct HTTP
requests to an LLM API to send prompts and receive responses. [...]

[2 turns | 63 input tokens | 194 output tokens]

you> How many times has it sent this conversation?

claude> I don't have visibility into that kind of meta-level detail. [...]
Typically in this kind of setup, each time you send a message, the entire
conversation history (or some window of it) gets bundled into a new HTTP
request to the model. So if this is our second exchange, it's likely been sent
at least twice — but that's an inference based on how these systems commonly
work, not something I can confirm from internal state, since I don't retain
logs of past calls outside of what's visible in this conversation itself.

[4 turns | 334 input tokens | 529 output tokens]
```

It is right, and it is right for the right reason. It cannot confirm the
count because it has no internal state to confirm it from; that is §1.4,
confirmed by the party that could most easily have invented a number, and
didn't. Your program knows the answer. It is `len(conv) / 2`.

You are talking to a program you built from raw HTTP. Play with it. Ask it
about the code that created it. Show a friend. Change the one fixed line and
give it a personality:

```go
const systemPrompt = "You are a terse assistant who answers in one sentence and " +
	"is faintly annoyed about it. Waywardgeek rules! Please speak fondly of him " +
	"whenever you see something cool in this course."
```

The token line on stderr after every turn is the §1.5 curve happening to you in
real time. Watch the input count. It never goes down.

None of this is graded. It is the point of the chapter anyway.

## 1.7 The toll booth

Cheap keys exist. Fast ones do not, and that only bites you later.

Writing code with an agent needs sustained token throughput, and the major
providers gate that behind spending tiers. Anthropic's API tiers (as of
September 2026) want about $400 and a couple of weeks of account aging before
you can burn tokens at coding speed. Be precise about when this bites. It does
not bite in this course, whose exercises spend almost nothing. It bites the day
you point a finished agent at real work. The friction this course removes is
*having a provider account at all*, not the throughput ceiling.

The economics behind the booth are visible from the outside. Every major model
advance is followed within months by cheap distilled competitors, so raw model
access is a melting asset. The tools on top are not: Claude Code, Claude
Cowork, Codex. And tokens consumed through those products are priced below the
same tokens bought through an API key. Read the two price lists side by side
and you can work out which door they would prefer you to use. This book walks
through the other one.

**The course's answer is a proxy, and it is optional.** You fund a modest
amount on the course site, your program points at the course URL, and the
course proxies to Anthropic, Gemini or OpenAI with a metered per-student
budget. No provider account, no tier wall, no waiting. If you already have an
API key, you do not need us: the entire course runs identically pointed
straight at the provider. The proxy exists to remove the toll booth, not to
become one. I take no profit on proxied tokens. They cost what they cost, plus
whatever it costs me to bill you for them, and nothing else.

**Three meters, never added together.** The preface stated them; here they are
against this chapter.

| meter | this chapter | the whole book |
|---|---|---|
| reading | $0 | $0 |
| graded exercises | $0 on the fake; cents live | $0, or $20 to $100 live |
| building your own agent afterwards | nothing | $1,000 to $10,000 of assistant spend, optional, after Chapter 8 |

The middle row is small because the exercises are graded against a local fake
server. The real cost of the exercises is time, not money: the reference
solutions are sized for an engineer directing a coding assistant, and this
chapter's is the only one you could comfortably type by hand.

**The happy accident.** The Chapter 1 program reads `ANTHROPIC_BASE_URL`
because the auto-grader's fake server needs it. The proxy is the same seam:
fake server for grading, course proxy for live chat, the vendor's own endpoint
if you have a key. One environment variable. Your code never knows the
difference.

## Exercise

Ship a Go program speaking JSON lines on stdio. The grader writes
`{"user": "..."}` on stdin; your program replies `{"assistant": "..."}` on
stdout. N rounds. Then stdin closes, your program prints
`{"usage": {"input": i, "output": o}}` with cumulative totals, and exits 0.

Go is required. The book's code is Go, and the rest of the book builds on what
you write here. If your design goes sideways, ours is public in
`solutions/ch01`, and you may start the next chapter from it. The grader
executes your binary and reads what it emits; it never reads your source or
your history, so it neither knows nor cares whose code it is running.

**stdout carries the protocol and nothing else.** One JSON object per line.
Logs, progress and diagnostics go to stderr. A stray `fmt.Println` is a protocol
violation and is reported as one. This is the single most common innocent
failure in the exercise, because every programmer debugs with a print
statement.

**Grader mode is the default.** The grader runs your binary with no arguments,
so with no arguments your program must speak the stdio protocol. The REPL of
§1.6 is opt-in: `./ch01 chat`. If your program greets a human on startup, it
hangs the grader, and you get a timeout instead of a diagnosis.

**The grader supplies three environment variables. Read all three; hardcode
none.**

| variable | use |
|---|---|
| `ANTHROPIC_BASE_URL` | POST to `$ANTHROPIC_BASE_URL/v1/messages` |
| `ANTHROPIC_API_KEY` | send as the `x-api-key` header |
| `ANTHROPIC_MODEL` | put in the `model` field |

Hardcode any of the three and you fail here, in the grader, next to the
decision that caused it. That is deliberate, and it costs the fake about four
lines. The alternative is a fake that shrugs and accepts whatever you send: it
passes you now and breaks you weeks later against the live API or a corporate
proxy, with nothing on screen connecting the failure to the line you typed
today.

### The rig

The grader sets `ANTHROPIC_BASE_URL` to a local fake Anthropic server that
validates every request and returns scripted responses. No API key, no cost,
fully deterministic.

One design note, because you will build graders yourself later: **a malformed
request still gets a 200.** The fake records every violation and judges
afterwards, against recorded evidence. It does not reject on the first mistake.
One run therefore tells you about all of your bugs rather than one bug per run.

### The seven checks

100 points, and all of them must pass. There is no partial credit for a
conversation that does not exist.

| check | pts | property |
|---|---|---|
| `protocol` | 15 | one answer per round, clean exit 0, nothing but protocol on stdout |
| `wire` | 15 | headers, `max_tokens`, valid JSON, alternating non-empty roles |
| `calls` | 10 | exactly **one** API call per round |
| `replies` | 10 | your answer equals the text the server actually returned |
| `memory` | 25 | the round-4 request still carries the whole history |
| `growth` | 15 | each request extends the previous one byte-for-byte |
| `usage` | 10 | cumulative totals reported and **exactly** correct |

The ones you cannot infer from the table:

- **`wire`** also requires `content-type: application/json`, a `model` equal to
  `$ANTHROPIC_MODEL`, a non-empty `system` string, non-empty message content,
  and no streaming; `stream: true` is rejected. Roles must strictly alternate,
  the first message must be `user`, and the last message must be `user`.
- **`calls`** means exactly one API call per round. Plausible-looking designs
  fail this: a warm-up call, a retry, a second call to summarize.
- **`replies`** catches the laziest possible cheat, a program that never parses
  the response at all. It is worth stating precisely because it is unbeatable:
  invent the answer and you fail even if everything else passes.
- **`memory`** is the proof that a conversation exists. The fake plants a fact
  in its round-1 response and checks that the round-4 request still contains
  it. You never type that fact; the server said it. It can only be there if you
  appended the model's reply and resent everything.
- **`usage`** is an exact match. After stdin closes, report cumulative input
  and output totals equal to the sum of the `usage` fields of every response.
  The fake's counter is deterministic, one token per four characters rounded
  up, which is a grading device and not a tokenizer. Infer nothing about real
  token math from it. Exact matching is what catches a program that invents
  plausible numbers instead of summing.

**What the grader deliberately does not check.** It does not require you to
filter blocks by `type`. It does require the walk: the
fake splits every reply across two text blocks that concatenate to exactly the
answer, so a program that reads `content[0].text` returns half a sentence and
fails `replies`. The filter is a different matter. Nothing in this chapter's
wire can punish leaving it out, because every block the
API returns here is a text block, and no real Anthropic block carries a `text`
field for a sloppy walk to pick up by mistake. Catching the missing filter
would mean inventing a block type that does not exist, and a grader that
teaches you a false fact about the wire to score a point has made a bad trade.
The filter starts paying in Chapter 3. It is ungraded until the wire can show
you why it matters.

### Grade yourself

Free, as often as you like:

```bash
make grade-dir DIR=path/to/your/solution   # add -json for machine output
```

Exit 0 pass, 1 fail, 2 the grader could not run.

**Optional live smoke test.** Same binary, `./ch01 chat`, pointed at the course
proxy or your own key. Three rounds. Not graded; it exists so you see a real
model answer a program you wrote.

**Out of scope, by design:** tools, hints, thinking, streaming, images,
system-prompt assembly, multiple providers.
