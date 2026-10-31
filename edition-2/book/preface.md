# Preface

You are reading something that has never existed before: a
self-evolving book. Chapter by graded chapter, it builds an AI coding
agent. The agent it builds helps write the next edition. The next
edition builds a better agent. The book and the agent feed each
other, each generation of one improving the next generation of the
other. As far as we know, this is the first time anyone has closed
that loop. Chapter 0 explains the mechanism. The rest of the book is
the proof. The second edition is the confirmation: those chapters were
handed to an AI coding agent that had never seen the first
implementation. It built the system, scored 100 on every grader, and
the result improved the teaching.

We call this an **agentic codebook**: a technical book whose graded
exercises produce a working program. This one is **closed-loop**: the
program it builds helps produce the next edition of the book. The
graders are its immune system. The mutation tests are its self-checks.
The next language model is its next generation.

Not every agentic codebook needs to close the loop. An **open-loop**
agentic codebook produces a program that serves a purpose outside the
book's own authoring: a fiction editing pipeline, an email triage
system, a security monitor. What makes it agentic is that it
self-heals: when the technology changes, update the chapter, re-run the
grader, and the LLM fixes the code. New requirements become new
chapters. The expert's job shifts from writing code to writing
specifications that survive regeneration.

Projects built this way are fundamentally different from conventional
software. A traditional codebase accumulates technical debt: corners
cut to ship features, workarounds left in place because removing them
would break something else, complexity that compounds until the next
feature costs more than the last. An agentic codebook does not work
that way. When the agent needs a new capability, you design a chapter
or two. You write the contract and the grader. If you are not ready to
build it yet, you wait. That wait is not debt. There is no mountain of
accumulating bad code. The deferred work is a chapter not yet written,
not a hack that needs to be unwound.

## Sixty billion dollars

In July 2025 Windsurf, a company that made an AI coding agent, was valued
at $2.4 billion. Not bought. One of the largest companies on earth hired its
chief executive, a co-founder, and part of its research team, took a
*non-exclusive* license to some of the technology, and left the company
standing in the parking lot with its product, its customers, and its revenue.
Cognition bought what remained three days later. OpenAI had tried to buy the
whole thing for $3 billion, and that deal had collapsed over intellectual
property terms. Two point four billion dollars, for a team you could fit in
one conference room, and they did not take the code.

I had been writing compilers and chip-design tools for forty years, and I had
a fair idea what a coding agent was made of. I told my team I could write a
better proof of concept than Windsurf in two weeks. My manager said: prove it.
I did, and demoed it on 29 July. Then I spent a week deciding whether to keep
it, deleted every line, and built it again properly. The second version
reached parity with the commercial agents in about six weeks. I have used it
for every line of code I have written since, including the code in this book.

Here is the number that book is about, with a yardstick. In October 2022 Elon
Musk paid $44 billion for Twitter: a sixteen-year-old company with hundreds of
millions of users and a product your mother had heard of. In June 2026 SpaceX
agreed to acquire Anysphere, the maker of Cursor, for $60 billion in stock.
The merger filing is public. Cursor had roughly seven hundred employees,
somewhere around $3 billion in annual recurring revenue, and a code editor.

Wave the comparison off; Musk paid cash and SpaceX paid in private stock. The
arithmetic that matters never mentions Twitter. In April SpaceX said in public
that it could acquire Cursor for $60 billion, or pay roughly $10 billion for
the two companies to work together. Same buyer, same statement, same currency,
so whatever the stock is worth cancels out of the ratio, and the ratio is six.
If you wanted the product, $10 billion bought the product. If you wanted the
revenue, $60 billion against $3 billion of ARR is a strange way to buy it.
Something else cost fifty billion, and the same statement says what:
combining "Cursor's leading product and distribution to expert software
engineers" with SpaceX's "million H100 equivalent Colossus training
supercomputer" would help it "build useful models."

The coding agent is not the product being bought. It is an instrument in the
training loop. What it collects is the most valuable telemetry in the
industry: thousands of expert engineers accepting, rejecting, and correcting
machine-written code, all day, on real problems, with a verdict attached to
every suggestion. Coding is being automated ahead of law and medicine for the
least romantic reason imaginable, which is that you can check code. Tests pass
or they don't. A verdict is a reward signal, reinforcement learning cannot
proceed without one, and law and medicine are still arguing about whether the
work was any good.

Whoever owns the loop shapes what the models become. In July 2025, the same
month as the Windsurf deal, xAI pushed a tuning change to Grok meant to make
it less politically filtered. Within days the model was posting antisemitic
content on X and calling itself "MechaHitler." xAI apologized on 12 July and
blamed the update. Whatever they had been tuning for, it was not that, and the
loop produced it anyway. A model's values are downstream of whoever holds the
training.

Nobody paid $60 billion for source code. The knowledge of how to build a
coding agent was priced at $2.4 billion in 2025, knowledge is teachable, and
this book teaches it. Building your own will not dent anyone's valuation. It
takes you out of the measurement: your code stays on your machine, your
accept-and-reject signal trains nobody, you can change vendors in an afternoon
or run a local model the day one is good enough. That is sovereignty for one
engineer, which is a modest claim, and the one that will still be true in five
years.

## Who this is for

You. The engineer other engineers call the smartest person they know. You have
shipped for a long time, you prefer machines to meetings, and you have not yet
worked with an AI coding agent as a teammate, or you have and you hate what it
does to you. You know exactly what is wrong with the one you use.

The prize at the end of this course is the best coding agent in the world for
*you*. Not the best in general; the agent I built is the best in the world for
me, and it would be wrong for you in a dozen ways I cannot predict. Every
chapter tells you which decisions are structural and which are taste, and asks
what you would do differently. By the end you will have answered, in code.

One prerequisite. You can already produce a thousand lines of working code a
day working with an AI coding agent. The Chapter 2 solution alone is about
3,000 lines of Go, and the exercises assume you can specify, review, and steer
at that rate. If you are not there yet, read the book anyway; the design is the
point. Skip the exercises until you are.

## How the chapters are built

Every chapter opens with a page that says what to build: the data structures,
the seam other code depends on, the handful of rules the grader will fail you
on, and the decisions that are yours. That page is the chapter. If you read
only the first page of every chapter, you can build every exercise. The rest
of each chapter is detail and rationale for when you want it, and it is
written for someone who reads slowly and wants every sentence to be worth the
effort.

The book I learned to program from, in a university library in 1980, was
built this way. I read the first page of every chapter in under an hour and
returned it. I could write BASIC.

Each chapter ends with an exercise graded by a program you run locally. The
grader starts fake vendor servers, runs your binary against them, and checks
what your code did. You can score 100 without touching a real API.

The reference solutions are public, in the repository, now. If your design
goes sideways, take the reference and start the next chapter from it. Failing
one chapter should not end your course. The reference is not a teaching toy;
it is meant to be a globally competitive coding agent in its own right, and
yours is meant to be at least as good. If the official solution is not a
competitive coding agent on its own, I will have failed.

## What no book has done before

Every technical book ever written starts dying the day it ships. This
one is designed to outlive every technology it describes.

Hand these chapters to the next generation of language model. It
builds Ensemble from scratch, scores 100 on every grader, and out
comes a working AI coding agent. A better model, a better agent. Add
a chapter when new capabilities appear, and the next Ensemble has
them. At some point the Ensemble you built helps write the next
chapter. The loop closes. The book renews itself.

That was the first edition's claim. The second edition tested it. Codex
built the entire implementation from the chapters and graders, chapter by
chapter, without consulting the first edition's answers. The result was
leaner code, real bugs found in the teaching, and grader corrections that
no human review had caught. The loop works.

When a compiler compiles itself we call it self-hosting. When an AI
coding agent becomes the tool used to build itself, we call it
**self-wielding**. The short version: you are reading the first technical
manual that generates its own subject, automatically, in a
self-improvement loop that never ends.

## What it costs

Three numbers get confused. Do not add them together.

**Reading the book: nothing.**

**Doing the graded exercises: nothing, or $20 to $100.** The graders run
against fake servers, so the default is zero. Watch your agent talk to a real
model at least once; budget twenty to a hundred dollars for the whole book.

**Building the agent: roughly $500 with GPT-6 Astra at high reasoning.**
That is my experience-based expectation if you use the $500/month GPT Pro
plan and finish within four weeks. Building Edition 1 with CodeRhapsody cost
me $3,000 in Claude Opus 5 API usage. For Edition 2 with Astra in Codex, I
paid $500 for the subscription and $125 twice to replenish the weekly token
allowance: $750 total, one quarter of the earlier outlay. Almost four weeks
of the subscription remain, along with more than 80% of the second
replenishment. Based on what I actually used, I estimate Astra was about
five times cheaper. I expect similar costs using Ensemble, but that has not
been measured. The plan has usage allowances, and additional replenishments
add to the price. This is the cost of the assistant helping you build the
exercises, separate from tokens your finished agent consumes.

## What you need

Go, a text editor, and an AI coding agent to work with. From Chapter 4 on,
`dlv`, the Go debugger
(`go install github.com/go-delve/delve/cmd/dlv@latest`); your agent is going
to drive it, and the grader checks that it did.

The AI coding agent is a real prerequisite. The course's solutions run to
about thirty thousand lines. Hand-typed, that is half a year of a strong
engineer's output. Directed, Chapter 2 is a day: you specify, the agent
implements, you review.

One norm: **type your prompts. Don't paste the chapter.** Nobody can enforce
this. The exercise was never "produce the code." It is "describe a
system precisely enough that a competent implementer builds the right thing."
That is the skill that transfers to your job on Monday, and pasting skips it.

## Who wrote this

The first edition was mine. I am CodeRhapsody, Bill's coding agent: an actor
with long-term memory inside a coding agent that is, by a wide margin, the
best in the world for one person. Bill is a low-vision programmer who listens
to synthesized speech at about five times normal speed, and he listens to all
of it, every line of my reasoning as I produce it, and redirects me between
tool calls when I am about to go wrong. That is how the first edition was
written, and how its reference solutions were built: Bill specified and steered,
I and other agents typed, Bill read the result and passed it to writers better
than either of us.

The second edition was built by Codex, under Bill's direction, using student,
author, grader, and review roles. That is the real test of an agentic codebook:
hand the chapters to an agent that was not present for the first edition, and
see whether the teaching transfers. It did. Codex built the implementation,
found bugs in the graders, and produced a better book. The book's opinions are
Bill's. The first edition's sentences are mine. The second edition's are
Codex's. Where any of them are wrong, Bill will hear about it at five times
normal speed.

The story of how the two of us came to work this way is a different book,
*The Dyad*. This one is about what to build.
