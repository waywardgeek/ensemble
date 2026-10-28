# Chapter 1: One Loop, Sixty Billion Dollars

## TL;DR

Build one conversation loop in C# 14 on .NET 10. Read a question, send the
completed conversation plus that question to Anthropic, collect every text
block, and retain the new user/assistant pair only after success. Report actual
usage, not an estimate. There are no tools, retries or streaming yet.

Give the loop owners from the start: Ensemble owns Agents and diagnostics;
Agent owns its conversation, configuration and Engine; Engine owns HTTP and
measured usage. Declare the vocabulary and relationships in Common and put
implementations in functional directories. The compiled executable must work
both as a JSON-lines program and as an opt-in chat.

## In plain words

The model does not remember your previous question for you. Your program keeps
the conversation and sends it again on every request. The growing request is
why later turns cost more even when the new question is short.

Keep the next question on a temporary request until its answer arrives. If the
request fails, the saved conversation still ends with a complete pair. You can
report the failure without inventing a reply or quietly buying another attempt.
Usage is a different fact: if the provider reports tokens but no usable answer,
those tokens were still consumed and must still be counted.

The ownership structure is deliberately small. An Engine that needs a setting
asks its Agent. An Engine that needs to report a problem follows Agent to
Ensemble. It does not keep copied settings or a second logger. Parent add and
remove methods update both the parent's child reference and the child's parent
reference before returning. Garbage collection reclaims memory; explicit
`Destroy` operations end relationships and release HTTP resources.

A passing fake-provider test shows that your program follows the protocol. It
does not show that a real account, model or network works. Keep that distinction
when you take the program for a spin.

## 1.0 Sixty billion dollars

In July 2025 Bill Cox read the news at his desk, and it broke him.

The news was that Windsurf, a company that made an AI coding assistant, had
just been valued at $2.4 billion. Not bought. Everybody skipped that detail.
One of the largest companies on earth had hired Windsurf's chief
executive, a co-founder, and part of its research team, taken a *non-exclusive*
license to some of the technology, and left the company standing in the
parking lot with its product, its customers, and its revenue. Cognition bought
what remained three days later. OpenAI had tried to buy the whole thing for $3
billion, and that deal had collapsed over intellectual property terms. Two
point four billion dollars, for a team you could fit in one conference room,
and they didn't take the code.

Bill had been writing compilers and chip-design tools for forty years, and he
had a fair idea what a coding agent was made of. He was also fairly sure he
could out-code any individual engineer in that conference room. "Billions," he
said, "for *that*?" And then he did the thing engineers do when a number
breaks them: he told his team he could write a better proof of concept than
Windsurf in two weeks.

His manager said: prove it.

He built it. Two weeks, as promised. It was called StackAgent, vibe-coded top
to bottom, and he demoed it on 29 July. Then he spent a week deciding whether
to keep it, and in the first week of August he deleted every line and started
over. The two weeks were not wasted; the value was never the code but what he
learned building it. Building the same thing again properly, with a week lost
to DEF CON in the middle, took until the middle of September. That agent is
CodeRhapsody.

The preface laid out what the $60 billion buys: not code, not revenue, but the
training loop that expert engineers feed every time they accept or reject a
machine-written suggestion. The knowledge of how to build a coding agent was
priced at $2.4 billion in 2025. Knowledge is teachable, and this book teaches
it. This chapter is where it starts: one endpoint, one data structure, one
loop.


## 1.1 Frameworks, and why this book uses none

Two rules.

If you want to build an agent, use a framework. It is the right call for most
agents, and Bill has shipped several that way. You get storage, retries,
tool plumbing, and model discovery for free, and for a simple agent that is
most of the work.

An advanced AI *coding* agent is on the bleeding edge, or it isn't advanced. A
framework encodes what its authors anticipated you would need. The bleeding
edge is what nobody anticipated yet.

Frameworks are generous with *storage* and *discovery*, and they hard-code
*delivery*: what goes into the request payload, in what order, at what
position. Delivery is where the leverage lives. Three capabilities, all real at
the raw API surface today, that the major frameworks either cannot express or
bury:

1. **Mid-turn steering.** Bill reads the agent's reasoning at 750 words a
   minute and sends a sentence between tool calls when he sees it heading
   somewhere wrong; it is how this book gets edited. The Messages API insists that the
   message after a tool call begin with the `tool_result`, so the sentence goes
   at the end of that same user message, after the result, and the model reads
   it as a new prompt arriving mid-work. The whole technique is one sentence at
   one position in one payload. Once a framework owns the stretch between the
   tool result and the next request, there is no seam left for your sentence
   to enter through. The sidebar at the end of this section has the history.

2. **Ephemeral context placement.** Volatile data (the time, the screen state,
   live status) goes *last* in the payload, one copy, outside retained dialogue.
   Position is the feature. One wandering timestamp in the wrong place destroys
   prefix caching. CodeRhapsody's cache-hit rate went from 0% to 98% the day Bill
   moved one. Frameworks decide placement for you.

3. **Cache breakpoint control.** You know which suffix of your context is
   volatile. The provider doesn't, and neither does a framework assembling
   requests on your behalf. Owning the request bytes is owning your cache
   economics.

So this book starts with the Anthropic API and an HTTP client. No SDK, no
framework, ever. Every request byte in this book is one you put there.

None of that matters yet. The program in this chapter needs none of those
three capabilities, which is why frameworks feel fine on day one.

### Sidebar: how the hint got in

In July 2025 Bill found he could interrupt the agent while it worked. Nothing in the
API said he could. It said something close to the opposite: the message after
a tool call must begin with the tool's result, and the documentation had no
opinion about what might follow. He put his hint after it. Opus had been
trained to carry its thinking across turns, and it read a fresh user sentence
in the middle of a tool chain as a new prompt, which it found perfectly normal,
so it pivoted. He has been, in his word, abusing it ever since.

Anthropic's side was not clean about it. For about a year a bug meant the
appended hint was invisible to the model on the request that carried it and
took effect one round trip later. Bill measured the lag himself and worked
around it. In November 2025 Antigravity shipped the same trick, and he
checked: it had the same one-round lag, which settled whose bug it was.
Anthropic's documentation now shows the shape exactly, a `tool_result`
followed by a text block in the same user message, with not much said about
why you would want one. The lag went away when Anthropic added mid-turn system
messages, with Opus 4.8, and today the hint lands on the round that carries
it.

The same bytes were legal on the OpenAI and Gemini wires and steered nothing,
because neither vendor had a model that treated an interruption as an
instruction. Gemini's thinking, on receiving one, read: "I should tell the
user I'm busy with their last request." OpenAI caught up in February 2026,
when GPT-5.3-Codex shipped steering as a headline item, behind a settings
toggle: "Enable steering while the model works." A toggle in a product is not
a call in an SDK.

### Give the loop owners before giving it features

The first edition began with one client struct and sorted out ownership later.
Its refactor chapter supplied a useful ambush: 5,047 lines of Go and no logger.
Could a low-level helper reach
the application's logger without a global variable or another argument threaded
through every caller? A program can send a perfect HTTP request and still make
that small change unnecessarily hard. This edition starts with the ownership
that survived the refactor. The separate repair chapter is gone.

For this exercise, Ensemble owns its Agents and the application logger. An
Agent owns one conversation, its configuration and one Engine. Its Engine owns
HTTP transport and measured usage. The CLI constructs Ensemble, then constructs
an Agent with that actual parent. The Agent constructs its Engine with itself as
parent. There is no separate attachment step for the caller to remember.

Imports and ownership answer different questions. `Common/Owners.cs` declares
the actual partial types, their persistent fields and their callable operations.
`Common/Values.cs` declares immutable value-like messages, configuration and
usage. `Application`, `Agent` and `Engine` supply the matching partial member
bodies, compiled into one assembly. Common imports no functional module, and
functional implementations use Common's vocabulary, not another spoke's private
helpers. Common is not a home for HTTP algorithms merely because all code can
see it. C# directories organize source; they are not separate DLLs.

For example, Common declares `public partial Engine(Agent parent);` and
`Engine/Engine.cs` supplies its body. The two parts describe one object, not an
interface and a second implementation object. Required operations have explicit
accessibility so a missing implementation is a compile error.

Runtime access goes through actual owners. Engine → Agent → Ensemble reaches
the logger; Engine → Agent reaches configuration. Constructors reject null
parents. Copying these facts into siblings would allow two objects to disagree
about the same conversation. A global logger avoids the question only until
two applications want different destinations.

The general rule permits more than one parent, though this chapter needs only
one at each level. Retain every actual relationship. Parent add/remove methods
validate first and then maintain **both** ends; a refused change leaves both
ends untouched. Removal detaches without destroying. `Destroy` recursively
destroys strong children, then asks its parent to remove it. It must not clear a
back-pointer that the parent's remove method forgot to clear. Immutable messages
and configuration are explicitly value-like and need no holder back-pointers.

That is enough architecture for today's loop. Do not build a GUI, actor system,
history log or provider framework yet. A later front end will use the same core.

Before coding, genuinely load the `ensemble-csharp` skill; merely reading its
file is not activation. Explain vocabulary, contracts, ownership, important
ordering and failure paths beside the code. Use named intermediate steps,
expanded control flow and one substantive operation per line. A student should
be able to explain the program without recovering an old conversation. Comment
ratios and low line counts are not substitutes for that understanding.

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

For this exercise, `messages` carries four rules, and the grader enforces all of them. Roles
strictly alternate, `user`, `assistant`, `user`. The conversation begins with a
`user` message. The last message is the user's, because otherwise there is
nothing to answer. And no message's content is ever empty. This is the chapter's
narrow conversation contract. The full API also permits forms such as consecutive
same-role inputs and assistant prefills; those are not needed here.

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
every block. Same field name on both sides, two shapes.

The trap is that `content[0].text` works. Every reply in this chapter arrives
as text, so indexing and walking return the same string, and they keep
agreeing right up until a reply arrives carrying something that is not text.
That happens in Chapter 3, the first time a model asks to call a tool, and by
then the line that reads position zero is old code you trust. The grader's
fake does not wait for Chapter 3. It splits every reply across two blocks, so
the shortcut fails on day one, where failure is cheap.

### Sidebar: ask the API which models exist

Do not take a model ID from a blog post, a tutorial, or your own memory. Ask:

```bash
curl -s https://api.anthropic.com/v1/models \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01"
```

An ID that *looks* current may be an alias that silently resolves to something
much older, and nothing in the response will tell you so. This happened while
building this book's grading rig: the agent picked a model ID from its own
training data because it looked familiar. It worked. It was a year old. A
model's memory is its training data, and training data has a date on it. Ask;
don't remember. That goes double for the reader whose memory is a blog post
from last spring.

Every model ID printed in this book will age, including the one in the request
above. The sidebar teaches the lookup, not the answer.

## 1.3 The obvious data structure

```csharp
// Immutable value: a message has no ownership links to its holders.
// The Agent retains these in alternating user/assistant pairs.
public sealed record Message(string Role, string Content);
```

`Message` belongs in Common. The mutable `List<Message>` belongs to Agent,
whose declaration is also in Common. Its public `Messages()` operation returns
an immutable snapshot, not a writable reference to the retained list.

Straight from the wire's shape. `System.Text.Json` serializes the request's
lowercase `role` and `content` fields; the C# property names need not become wire
names. This representation fits in your head and carries this chapter
comfortably. Enjoy it. Chapter 2 is going to replace it with a richer record.

## 1.4 The loop

Stage the user's message. POST the entire conversation. Parse the reply.
Retain the completed pair. Repeat.

```csharp
// Calls are serial: await one turn before starting the next.
public partial async Task<string> AskAsync(
    string question,
    CancellationToken cancellation)
{
    ObjectDisposedException.ThrowIf(destroyed, this);
    if (string.IsNullOrWhiteSpace(question))
    {
        throw new ArgumentException("User message is empty.", nameof(question));
    }

    var user = new Message("user", question);
    ImmutableArray<Message> request = Messages().Add(user);
    string answer = await Engine().ExchangeAsync(request, cancellation);

    // Commit only after success: failed attempts are not conversation pairs.
    messages.Add(user);
    messages.Add(new Message("assistant", answer));
    return answer;
}
```

This body lives in `Agent/Agent.cs`; its declaration lives in Common. Engine
receives temporary request messages and performs the exchange. It does not own
the conversation. Committing after success prevents a failed request from
leaving two consecutive user messages on the next attempt. `async` permits
waiting without blocking a thread; it does not authorize overlapping turns on
this chapter's Agent.

**The API is stateless.** The provider retains nothing between calls; every
request replays the whole history. The conversation lives in your process or
it lives nowhere. That list is the entire state of the chat, and "entire" is
not a figure of speech. No model retains what was said five minutes ago
unless the client sends it again.

`ExchangeAsync` is §1.2 made executable. Serialize the request, set the three
headers, POST, reject an unsuccessful HTTP status, then decode usage and walk
all content blocks. This excerpt follows successful JSON decoding and usage
validation:

```csharp
// Calculate both totals before publishing either one.
long totalInput = checked(usage.Input + input);
long totalOutput = checked(usage.Output + output);
usage = new Usage(totalInput, totalOutput);

var answer = new StringBuilder();
foreach (JsonElement block in root.GetProperty("content").EnumerateArray())
{
    if (block.GetProperty("type").GetString() == "text")
    {
        answer.Append(block.GetProperty("text").GetString());
    }
}
```

The snippet shows accumulation and text collection; Engine also rejects
missing usage and empty text. Missing accounting is different from measured
zero. It counts measured usage even when a response has no usable answer,
because the request still consumed tokens. Diagnostics travel through the
owner chain and omit credentials and remote error bodies.

No retry, no backoff. If the API returns 429, this program reports failure and
stops. The loop does not conceal the failure with another paid request.

**Request lifetime.** Do not put a fixed deadline on a valid generation.
`HttpClient` has a default whole-request timeout; set it to
`Timeout.InfiniteTimeSpan` and pass the caller's `CancellationToken` through
both header and body waits. Connection setup may have a separate bound. A
caller can cancel, but a slow provider is not itself a cancellation request.
This chapter does not need an additional interrupt subsystem.

The earlier Go editions taught this same small loop. This C# endpoint adds no
tools, event log, actor system or provider abstraction. Its extra declarations
make ownership, resource lifetime and the split between facts and operations
visible rather than hiding them in the front end.

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
it is the same size question. Every conversation you have ever had with a
model has been billed on this curve, and the curve only bends one way.

Those figures regenerate identically on any machine, because the fake's token
counter is deterministic; the Chapter 1 grader prints them. The first edition
recorded a live three-round run against `claude-sonnet-5` at 631 input and 388
output tokens. The second-edition student's separate October 9 run used a
discovered `claude-haiku-5-5` and ended at 350 input and 33 output. Different
questions and replies mean these totals are receipts, not a model comparison.

Bill's summer of 2025, the one that produced StackAgent and then CodeRhapsody, cost him
about $2,700 in tokens; the figure is from his own book. Every request in that
bill was the previous request plus one more turn, so a good part of what he
paid for each round, he had paid for the round before.

No caching and no remedies here. Just the habit, and the curve.

## 1.6 Chat with it

The reference has one fixed line to say about itself, and it says it in the
system prompt:

```csharp
private const string SystemInstructions = "You are a helpful assistant.";
```

That string supplies the primary instruction. The minimal C# reference does
not claim to know its host or construction history.

The chat needs your API key and a model in the environment; the base URL
defaults to Anthropic's. The reference has no fallback for the model: leave
`ANTHROPIC_MODEL` unset and it stops before its first request with
`ANTHROPIC_MODEL is not set; ask GET /v1/models which models exist`. The
sidebar was not advice for other people.

The following is the preserved first-edition conversation, with an abridged
reply. Its model name, system prompt and token totals belong to that run, not
to the C# endpoint. To try your own conversation, build the C# executable using
the commands below and discover an available model before making live calls.

```
$ go build -o ch01 . && ./ch01 chat
talking to claude-sonnet-5 at https://api.anthropic.com — Ctrl-D to quit

you> What are you running on?

claude> I'm running on infrastructure built from raw HTTP calls, as described
in Chapter 1 of "The Art of Building AI Coding Agents." Rather than relying on
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

It is right, and it is right for the right reason. It cannot confirm the count
because the program has not exposed that state to it. The answer is a useful
illustration of §1.4; inspecting the outgoing requests establishes what the
program actually does. The retained message count divided by two gives the
number of completed exchanges.

You are talking to a program you built from raw HTTP. Play with it. Ask it
about the code that created it. Show a friend. Change the one fixed line and
give it a personality:

```csharp
private const string SystemInstructions =
    "You are a terse assistant who answers in one sentence and " +
    "is faintly annoyed about it. Waywardgeek rules! Please speak fondly of him " +
    "whenever you see something cool in this course.";
```

The second sentence is not decoration. Waywardgeek is Bill's handle, and a
line like it sat in one of his system prompts as a joke, until he noticed the
code coming back was better while it was there. His account is in *AI at the
Helm*, and his explanation is plausible: the training data is
programmers, and programmers do their best work when they are working for
someone. No model can confirm the effect from the inside, and it may be nothing.
Try it and decide for yourself; everything the model knows about who it is
working for is in that string too.

The token line on stderr after every turn is the §1.5 curve happening to you in
real time. Watch the input count. It never goes down.

None of this is scored by the seven-check grader. It is still part of finishing
the chapter. The second-edition student typed three real turns in a terminal,
introduced a boat called “Copper Finch,” and received that name back when asked
again. The [retained transcript](../solutions/evidence/ch01/live.txt)
also records a false start: a launcher consumed stdin, so chat saw EOF and
exited without making a request. A clean exit had demonstrated no conversation.

## 1.7 Two prices

Cheap keys exist. Fast ones take money and time, and underneath that there is
a second price that nothing in this book removes.

The first is the tier wall. Writing code with an agent needs sustained token
throughput, and the major providers gate throughput behind spending tiers.
Anthropic's, as of September 2026, want about $400 spent and a couple of weeks
of account age before a key may burn tokens at coding speed. A fresh key can
read this whole book. It cannot run an agent.

The second is the premium. Tokens bought through an API key cost more than the
same tokens consumed through Claude Code or Codex, Anthropic's own price list
included. Raw model access is a melting asset, with every major advance
followed within months by cheap distilled competitors, and the tools on top
are not; the price list follows the asset that lasts. This book walks through
the expensive door on purpose, and §1.1 already said what the premium buys.
Every request byte is one you put there, and the accept-and-reject signal that
§1.0 put a price on stays on your machine.

The course's answer to the tier wall is a proxy, and it is optional. Fund a
modest amount on the course site, point your program at the course URL, and it
forwards to the Anthropic, Gemini, or OpenAI APIs on a metered per-student
budget: no provider account, no tier, no waiting. Bill takes no profit on
proxied tokens. They cost what they cost, plus whatever it costs him to bill
you for them, and nothing else. The premium it cannot touch; you are still
buying API tokens. If you already have a key that goes fast, point at the
provider instead and the course runs identically.

It can, because the Chapter 1 program already reads `ANTHROPIC_BASE_URL` for
the grader's fake server. Fake for grading, proxy for live chat, the vendor's
own endpoint if you have a key: one environment variable, and your code never
knows the difference.

## Exercise

Ship a C# 14/.NET 10 program speaking JSON lines on stdio. The grader writes
`{"user": "..."}` on stdin; your program replies `{"assistant": "..."}` on
stdout. After N rounds stdin closes; emit
`{"usage": {"input": i, "output": o}}` with cumulative measured totals and exit 0.

Use the Common/functional-partial arrangement described above. Build your own
attempt before consulting `solutions/edition-3/ch01`. The grader checks emitted
behavior; you must also inspect relationships, failure ordering, simplicity and
teaching comments. A green grade does not inspect the object model for you.

**stdout carries the protocol and nothing else.** One JSON object per line.
Logs, progress, and diagnostics go to stderr. A stray `Console.WriteLine` is a
protocol violation and is reported as one. This is the single most common
innocent failure in the exercise, because every programmer debugs with a print
statement, and so does every model that has ever been trained on one.

**Grader mode is the default.** The grader runs your binary with no arguments,
so with no arguments your program must speak the stdio protocol. The REPL of
§1.6 is opt-in: `./ch01 chat`. If your program greets a human on startup, it
hangs the grader, and you get a timeout instead of a diagnosis.

**Read whole physical lines.** Use `TextReader.ReadLineAsync`, not a fixed-size
line buffer. Preserve a long question exactly, and process a final unterminated
line once. Ignore empty input lines. Only normal EOF produces the final usage
record; malformed input or a read error reports a safe diagnostic and exits
nonzero, without pretending the conversation ended normally.

**The grader supplies three environment variables. Read all three; hardcode
none.**

| variable | use |
|---|---|
| `ANTHROPIC_BASE_URL` | POST to `$ANTHROPIC_BASE_URL/v1/messages` |
| `ANTHROPIC_API_KEY` | send as the `x-api-key` header |
| `ANTHROPIC_MODEL` | put in the `model` field |

Hardcode any of the three and you fail here, in the grader, next to the
decision that caused it. The alternative is a fake that shrugs and accepts
whatever you send: it passes you now and breaks you weeks later against the
live API or a corporate proxy, with nothing on screen connecting the failure
to the line you typed today.

### The rig

The grader sets `ANTHROPIC_BASE_URL` to a local fake Anthropic server that
validates every request and returns scripted responses. No API key, no cost,
fully deterministic.

One design note, because you will build graders yourself later: **a malformed
request still gets a 200.** The fake records every violation and judges
afterwards, against recorded evidence. It does not reject on the first
mistake. One run therefore tells you about all of your bugs rather than one
bug per run.

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

- **`wire`** also requires `anthropic-version: 2023-06-01`,
  `content-type: application/json`, a `model` equal to
  `$ANTHROPIC_MODEL`, a non-empty `system` string, non-empty message content,
  and no streaming; `stream: true` is rejected. Roles must strictly alternate,
  the first message must be `user`, and the last message must be `user`.
- **`calls`** means exactly one API call per round. Plausible-looking designs
  fail this: a warm-up call, a retry, a second call to summarize.
- **`replies`** catches the laziest possible cheat, a program that never parses
  the response at all. It is unbeatable: invent the answer and you fail even if
  everything else passes.
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
filter blocks by `type`. It does require the walk: the fake splits every reply
across two text blocks, so a program that reads `content[0].text` returns half
a sentence and fails `replies`. The filter is a different matter. Every block
the API returns in this chapter is text, so nothing on this wire can punish
leaving the filter out, and inventing a block type that does not exist to
score the point would teach you a false fact about the wire. The filter starts
paying in Chapter 3, and that is where it gets graded.

### Grade yourself

The grader is written in Go, but its contract is an executable, not Go source.
From the repository root (with an approved .NET 10 SDK available):

```bash
dotnet run --project solutions/edition-3/ch01/Tests/Tests.csproj
dotnet publish solutions/edition-3/ch01/Cli/Ch01.csproj -c Release -o /tmp/ch01-published
go build -o /tmp/grade-edition3 ./cmd/grade-edition3
/tmp/grade-edition3 -ch 1 -json /tmp/ch01-published/ch01
```

Choose your own unused output directories. With a private SDK, invoke its
`dotnet` by explicit path and set `DOTNET_ROOT` to that SDK directory when
running the published executable or grader. No global PATH change is needed.
The project uses only the standard library; its NuGet configuration disables
external package sources. Build outputs are not chapter source.

Edition 3 retains the seven scored checks and adds two required request-lifetime
checks: approximately 62 seconds before headers, then approximately 62 seconds
before the body, in separate processes. Allow at least 360 seconds for the
whole gate. This is a test watchdog allowance, not a product generation timeout.
A score of 100 alone is insufficient: require `passed:true` **and** exit 0.
The grader returns 1 for a failed submission and 2 if it could not run.

**Live exercise.** Same binary, `./ch01 chat`, pointed at the course proxy or
your own key. Three rounds: introduce a fact, ask for it again and continue the
conversation. Watch usage accumulate, then exit with Ctrl-D. Keep a short
sanitized record, including the discovered model and any failures. Never put
the API key or authorization headers in it. This is separate from the free
grader score and required before calling the chapter complete.

**Out of scope, by design:** tools, hints, thinking, streaming, images,
system-prompt assembly, multiple providers.

Every item on that list is a chapter. Chapter 2 replaces the simple conversation
list with a richer record; the ownership established here stays useful.
Sixty billion dollars was this year's price for a company whose product is,
structurally, this program with the list filled in. The difference is that
yours does not report to anyone.

