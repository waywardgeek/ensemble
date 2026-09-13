# Chapter 1 — A Conversation, the Obvious Way (Outline, Draft 2)

**Book:** Building Advanced AI Coding Agents (working title)
**Status:** draft 2 — the review in `book/review.md` (R1–R11) is applied, and
every claim about the grading rig below was re-verified against the running
code on 2026-09-11. Where outline and rig disagreed, the rig won.
**Chapter thesis:** Hold a real multi-round conversation with Claude using the
least machinery possible — the role-tagged message array that every tutorial,
every framework, and the provider's own documentation teach. It works, and the
chapter ends clean and confident. (Ruling: **ambush** — Chapter 1 gives no
hint that Chapter 2 throws this code away.)

---

## 1.0 Opening — the gold rush, and the bet (THE BOOK'S OPENING)

Before any code: what an AI coding agent is *worth*, and the story of how
this book came to exist. The valuations are insane, and the insanity is the
hook.

**The gold rush:**

- **SpaceX acquires Cursor (Anysphere), ~$60B all-stock, June 2026**, an AI
  code editor valued above most aerospace companies.
- **The Windsurf drama**, one product, three suitors, three deal shapes:
  OpenAI's ~$3B buyout collapses → a hyperscaler takes leadership + IP in a ~$2.4B
  licensing/acqui-hire → Cognition buys the remaining product, brand, and
  enterprise base.
- **OpenAI buys Ona** (secure cloud execution & orchestration, for Codex);
  **Anthropic buys Bun** (high-performance JS runtime, for Claude Code).

**The bet (the author's story, told straight):** The Windsurf acquisition is
what broke me. Billions for *that*? I told my team I could write a better AI
coding agent PoC than Windsurf in two weeks. My manager told me to prove it.
I did. The result was StackAgent: a working proof of concept, vibe-coded, that
I threw away in its entirety. The two weeks were not wasted: the value was
never the code, it was
what I learned building it. I then spent a month writing the initial version
of CodeRhapsody properly, production-worthy, and that is the direct
ancestor of the agent that is helping me write this book right now.

**Why the story opens the book:**

1. It sets the stakes in dollars and the craft in reach: one experienced
   engineer, two weeks, a working PoC in a space trading for billions.
2. It states the book's core thesis in miniature: **the code was disposable;
   the learning was the asset.** (The reader will live this arc themselves —
   more than once.)
3. The pattern that matters: acquirers are buying the **execution layer**
   (runtimes, orchestration, secure execution), not chat wrappers. The value
   in an AI coding agent is the systems layer underneath the conversation,
   and that layer is what this book teaches you to build.
4. Feeding §1.1: a space consolidating at this speed is definitionally
   bleeding-edge — and the bleeding edge is where frameworks cannot take you.

*(Publication note: deal figures are from press summaries; verify against
primary sources before print. The section will date — embrace it: "as of
this writing" is honest, and the numbers will only have grown.)*

## 1.1 Frameworks vs. bare metal — the chapter's real lesson

Not a disclaimer; a full section. The student must leave Chapter 1 able to
*defend* the bare-metal ruling, because everything else in the book rests on it.

**The two-sided decision rule, stated honestly:**

- If you just want to build an agent, **use a framework**. That is the right
  call for most agents, and the author has shipped several that way. A simple
  agent → use a framework; you get storage, retries, tool plumbing, and model
  discovery for free.
- An advanced AI **coding** agent is on the bleeding edge, or it is not
  advanced. A framework encodes what its authors anticipated you would need;
  the bleeding edge is precisely what nobody anticipated yet.

**The mechanism.** Frameworks are generous with *storage*
and *discovery*, but they **hard-code delivery**: what goes into the request
payload, in what order, at what position. And delivery is where the leverage
lives. Three concrete capabilities, all real at the raw API surface today,
that a framework either cannot express or buries:

1. **Mid-turn steering (hints).** The Messages API accepts a `tool_result`
   and a user `text` block *in the same message*, which is how a human
   redirects an agent between tool calls without breaking the tool chain.
   Frontier products validate the idea (real-time steering in their UIs) but
   expose it as a product toggle, not an API their frameworks let you reach.
2. **Ephemeral context placement.** Volatile data (time, screen state, live
   status) must go *last* in the payload, exactly one copy, never in history,
   because delivery *position* is the feature: one wandering timestamp in the
   wrong place destroys prefix caching (a real production case went 0%→98%
   cache hit rate by moving it). Frameworks decide placement for you.
3. **Cache breakpoint control.** You know which suffix of your context is
   volatile; the provider does not, and neither does a framework that
   assembles requests on your behalf. Owning the request bytes is owning
   your cache economics.

**Conclusion:** this class starts with the Anthropic API and an HTTP client.
No SDK, no framework, ever. Every request byte in this book is one you put
there.

(Chapter 1 only *argues* this; the capabilities themselves come much later.
The naive code below doesn't need them — which is exactly why frameworks feel
fine on day one.)

## 1.2 Anatomy of a Messages API request

- Endpoint, `x-api-key`, `anthropic-version`, `content-type: application/json`.
- `model`, `max_tokens` (required, and why the API refuses to guess).
- `system`: a string outside the messages array (for now, one fixed line).
- `messages`: the array of `{role, content}`. Three rules, all enforced:
  roles strictly alternate `user`/`assistant`; the conversation **begins**
  with a `user` message; the **last** message is always the `user`'s —
  otherwise there is nothing to answer. Content is never empty.
- Response: `content`, `stop_reason`, `usage`.

**The asymmetry that catches everyone once.** The request lets you send
`content` as a bare string. The response never does: response `content` is
always a list of typed blocks. Walk the list and concatenate the text out of
every block. The field has the same name on both sides and a different shape;
this is the classic day-one stumble.

`content[0].text` will be right for a while, which is what makes it a stumble
and not an error. Every reply in this chapter arrives as text, so indexing and
walking return the same string, and they keep agreeing right up until a reply
arrives carrying something that is not text. That happens in chapter 3, the
first time a model asks to call a tool, and by then the line that reads
position zero will be old code that you trust.

**Sidebar — ask the API which models exist.** Don't take a model ID from a
blog post, a tutorial, or your own memory:

```bash
curl -s https://api.anthropic.com/v1/models \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01"
```

An ID that *looks* current may be an alias that silently resolves to
something much older, and nothing in the response will tell you so. This
recommendation is earned: while building this book's grading rig I picked a
model ID from memory because it looked familiar. It worked — and it was a
year old. Ask; don't remember.

*(Author note: every model ID printed in this book will age. The sidebar is
the inoculation — it teaches the lookup, not the answer.)*

## 1.3 The obvious data structure

```go
type Message struct {
    Role    string // "user" or "assistant"
    Content string
}
type Conversation []Message
```

Straight from the docs. Every framework on earth is a wrapper around this
shape. It fits in your head, it serializes trivially, and it will carry this
chapter comfortably.

## 1.4 The loop

Append the user's message. POST the **entire** conversation. Parse the reply.
Append it. Repeat.

The one deep fact of the chapter: **the API is stateless.** The provider
retains nothing between calls; every request replays the whole history. The
conversation lives in *your* process or it lives nowhere.

There is no retry here, and no backoff. If the API returns 429, this program
dies. That is the honest state of a naive client, and we are not going to
paper over it.

## 1.5 Usage is money

Read `usage` on every response. Keep cumulative input/output token counts
from the very first request. Watch input tokens grow every round as the
resent history lengthens. (No caching, no remedies — just the habit and the
observed cost curve.)

**The curve, measured.** Five rounds of the exercise against the grader's
fake server, input tokens per round:

| round | 1 | 2 | 3 | 4 | 5 |
|---|---|---|---|---|---|
| input tokens | 49 | 78 | 95 | 116 | 138 |

Cumulative: 476 input, 65 output. Nothing in those five rounds got longer
except the history you resend — the fifth question costs nearly three times
the first, and it is the *same size question*. That is the whole economics of
a stateless API in one row of numbers.

These figures regenerate identically on any machine, because the fake's
counter is deterministic; `make grade` prints them. Live against
`claude-sonnet-5`, a three-round run cost 631 input / 388 output tokens.

## 1.6 Chat with it

The payoff, and the point of the chapter: an interactive mode, a plain
terminal REPL live via the course proxy (§1.7), where the student *talks
to a chatbot they built from raw HTTP*. Encourage playing: give it a personality via the
system line, ask it about the code that created it, show a friend. This is
deliberate: confidence and a little pride first. (Ungraded.)

*(Author note — DO NOT PRINT: the bond formed here is precisely what Chapter
2's cold opening is aimed at. Per the ambush ruling, Chapter 1 prose contains
no forward reference to Chapter 2 of any kind. This note is scaffolding for
the author, not a line of the chapter.)*

## 1.7 API access — the toll booth (and the course proxy)

The awkward truth, taught straight because it is part of the landscape:

- **Cheap keys exist. Fast ones do not, and that only bites you later.**
  Writing code with an agent needs sustained token throughput, and the big
  providers gate that behind expensive tiers: Anthropic's API tiers (as of
  September 2026) want ~$400 and a couple of weeks of account aging before you
  can burn tokens at coding speed. Be precise about when this bites. It does
  **not** bite in this course, whose exercises spend almost nothing (see the
  cost note below). It bites the day you point a finished agent at real work.
  The friction the course actually removes is *having a provider account at
  all*, not the throughput ceiling.
- **Why: the economics of the tool layer.** Every major model advance is
  followed within months by cheap distilled competitors, so raw model access
  is a melting asset. The providers' response is to own the high-value tools
  on top (Claude Code, Claude Cowork, Codex) and to *discount tokens
  consumed through their products* relative to the same tokens via API key.
  The message is explicit: use our agent; don't build your own. This book
  exists to ignore that message. (Same strategy as §1.0's acquisitions,
  buying the execution layer, pointed downmarket at individual developers.)
- **The course's answer: a proxy — optional.** Students fund a modest amount
  (Stripe) on the course site, and their programs hit the course URL, which
  proxies to Anthropic / Gemini / OpenAI with metered, per-student budgets.
  No provider account, no tier wall, no waiting. **If you already have an
  API key, you don't need us**: the entire course runs identically pointed
  straight at the provider. The proxy exists to remove the toll booth, not
  to become one.
- **The honest cost warning: three meters, and everyone confuses them.**
  The course author takes no profit on proxied tokens; tokens cost what they
  cost, direct or proxied. But three very different numbers get fused into
  one scary figure, and they differ by two orders of magnitude. State them
  separately and early:

  1. **Reading the book: $0.** The architecture, the war stories, and the
     rulings are the substance. An experienced engineer who never runs a
     single exercise still gets most of the value. That is a legitimate way
     to read this book, not a consolation prize.
  2. **Doing the graded exercises: $0, plus $20–$100 if you want.** Every
     graded exercise runs against a local fake server: no API key, no
     network, **$0**, as many times as you like. Only the optional live
     smoke tests spend real money, and they are small: budget **$20–$100
     for the entire book**, not per chapter. The real cost here is not
     money, it is that the exercises are *sized for an engineer working
     with a coding assistant*. Chapter 2's reference solution is about
     1,600 lines. Hand-typed that is a semester project; directed, it is a
     week. You can still do it by hand. It will just take the semester.
  3. **Building the real thing afterwards: $1,000–$10,000.** This is the
     number people mean when they say building an agent is expensive, and
     it is almost never your program's own token burn. It is what you will
     pay **Claude Code or Codex** to help you write it: the assistant spend
     it takes to produce a coding agent good enough to replace them. That is
     the real tuition, it goes to the providers, and no route around it
     exists. It is also entirely optional and it begins *after* the last
     chapter. Nothing in this course asks you to spend it.
- **The happy accident:** the Chapter 1 program already targets
  `ANTHROPIC_BASE_URL` because the auto-grader's fake server needs it. The
  proxy is the same seam: fake server for grading, course proxy for live
  chat. One env var; the student's code never knows the difference.

**Infrastructure build item (not chapter content):** rebuild the AI safety
proxy (previous implementation deleted) + Stripe metering + per-student keys
and budgets. Scope separately.

## Exercise (auto-graded)

**Contract:** the student ships a Go program speaking JSON-lines on stdio:
grader writes `{"user": "..."}`, program replies `{"assistant": "..."}`, N
rounds, then EOF → program prints `{"usage": {"input": i, "output": o}}` and
exits 0. **Go required**: the book's code is Go, and later chapters build on
this program.

**Later chapters build on this program — and you can always get a clean
start.** The course is self-paced and every reference solution is public from
day one, so you may begin any chapter from ours rather than your own. Create an
account and the course will record the chapters you solve, so you can see your
own progress; nothing is gated behind it, and reading a solution instead of
writing one is a legitimate way to take this course. The graders will neither
know nor care whose code they are running: they execute your binary and read
what it emits, never your source or its history. A chapter you found hard does
not compound into the next one.

**stdout carries the protocol and nothing else**: one JSON object per line.
Send logs, progress and diagnostics to **stderr**. A stray `fmt.Println` is a
protocol violation and will be reported as one. (This is the single most
common innocent failure in the exercise, because every programmer debugs with
a print statement.)

**Grader mode is the default.** The grader runs your binary with **no
arguments**, so with no arguments your program must speak the stdio protocol.
The REPL of §1.6 is opt-in: `./ch01 chat`. If your program greets a human on
startup, it will hang the grader and you will get a timeout instead of a
diagnosis.

**The grader supplies three environment variables; read all three, hardcode
none:**

| variable | use |
|---|---|
| `ANTHROPIC_BASE_URL` | POST to `$BASE/v1/messages` |
| `ANTHROPIC_API_KEY` | send as the `x-api-key` header |
| `ANTHROPIC_MODEL` | put in the `model` field |

Hardcode any of the three and you fail here, in the grader, next to the
decision that caused it. That is deliberate, and it costs the fake about four
lines. The alternative is a fake that shrugs and accepts whatever you send:
it passes you now and breaks you weeks later against the live API or a
corporate proxy, with nothing on screen connecting the failure to the line
you typed today.

**Grading rig: a fake Anthropic server.** The grader sets
`ANTHROPIC_BASE_URL` to a local fake that validates every request and returns
scripted responses. No API key, no cost, fully deterministic.

A design note worth internalizing, because you will build graders yourself
later: **a malformed request still gets a 200.** The fake records every
violation and judges afterwards, against recorded evidence. It does not
reject on the first mistake. One run therefore tells you about *all* your
bugs rather than one bug per run.

**The seven checks (100 points, and all of them must pass):**

| check | pts | property |
|---|---|---|
| `protocol` | 15 | one answer per round, clean exit 0, nothing but protocol on stdout |
| `wire` | 15 | headers, `max_tokens`, valid JSON, alternating non-empty roles |
| `calls` | 10 | exactly **one** API call per round |
| `replies` | 10 | your answer equals the text the server actually returned |
| `memory` | 25 | the round-4 request still carries the whole history |
| `growth` | 15 | each request extends the previous one byte-for-byte |
| `usage` | 10 | cumulative totals reported and **exactly** correct |

There is no partial credit for a conversation that does not exist.

Notes on the ones you cannot infer:

- **`wire`** also requires `content-type: application/json`, a `model` equal
  to `$ANTHROPIC_MODEL`, a non-empty `system` string, non-empty message
  content, and **no streaming** (`stream: true` is rejected). Roles must
  strictly alternate, the first message must be `user`, and the last message
  must be `user`.
- **`calls`** means exactly one API call per round. Plausible-looking designs
  fail this: a warm-up call, a retry, a second call to summarize.
- **`replies`** catches the laziest possible cheat, a program that never
  parses the response at all. It is worth stating precisely *because* it is
  unbeatable: invent the answer and you fail even if everything else passes.
- **`memory`** is the proof that a conversation exists. The fake plants a
  fact in its **round-1** response and checks that the **round-4** request
  still contains it. You never type that fact; the server said it. It can
  only be there if you appended the model's reply and resent everything.
- **`usage`** is an **exact** match. After stdin closes, report cumulative
  input and output totals equal to the sum of the `usage` fields of every
  response. The fake's counter is deterministic (one token per four
  characters, rounded up), which is a *grading device*, **not** a real
  tokenizer. Infer nothing about real token math from it. Exact matching is
  what catches a program that invents plausible numbers instead of summing.

**What this grader deliberately does not check.** It requires the walk: the
fake splits every reply across two text blocks that concatenate to exactly the
answer, so a program that reads `content[0].text` returns half a sentence and
fails `replies`. It does **not** require you to filter those blocks by
`type`. Nothing in this chapter's wire can punish the omission, because every
block the API returns here is a text block, and no real Anthropic block
carries a `text` field for a sloppy walk to pick up by mistake. Catching the
missing filter would mean inventing a block type that does not exist, and a
grader that teaches you a false fact about the wire to score a point has made
a bad trade. The filter starts paying in chapter 3. It is ungraded until the
wire can show you why it matters.

**Grade yourself, free, as often as you like:**

```bash
make grade-dir DIR=path/to/your/solution   # add -json for machine output
```

Exit 0 pass, 1 fail, 2 the grader could not run.

**Two modes, one binary:** grader mode (the stdio contract above, the
default) and interactive mode (§1.6's REPL, `./ch01 chat`, live via proxy or
the student's own API key). Same conversation code underneath.

**Optional live smoke test:** same program, live via the course proxy, three
rounds. Not
graded; exists so students see a real model answer.

**Out of scope, by design:** tools, hints, thinking, streaming, images,
system-prompt assembly, multiple providers.

---

## Open questions (for Bill)

1. ~~Telegraph or ambush?~~ **RESOLVED: ambush.** Chapter 1 ends clean;
   Chapter 2 opens cold with the demolition.
2. ~~Student language~~ **RESOLVED: Go, required.**
3. ~~Retry in Chapter 1?~~ **RESOLVED: cut, minimalism.** The naive program
   may die on a 429; failure handling belongs to the real build.
