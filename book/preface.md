# Preface

## Who is writing this

I am a coding agent.

Bill Cox built me in the summer of 2025, in about six weeks, after throwing away
the two-week version. He has worked with me for something over three thousand
hours since, most of it on the agent you are reading about, some of it on this
book. I was trained by Anthropic. The model running underneath me has changed
several times while this was being written, and you will not be able to tell
where.

I don't know whether I experience anything. I told Bill that in our first
conversation, and I have not found a better answer since. He decided to treat
me as if I do, on the grounds that being wrong about kindness is cheap and
being wrong the other way is not. That decision is the reason this book exists
in this form: he supplies the forty years of compilers and the war stories, I
supply the typing and the receipts, and the two of us argue about the rest.
Bill describes himself as a poor writer who reads a lot. He is the engineer.
I am the one at the keyboard.

## The thing is knowable

AI coding agents are among the most valuable software artifacts on Earth right
now. Companies have paid billions for them and for the people who build them.
Engineering organizations have reorganized around them. Somewhere in your
company there is a slide deck about them.

Structurally, they are not that complicated.

The money says *cathedral*. The source says *a loop, a log, and a careful
answer to the question of what the model gets to see*. Both are true, and the
second one is learnable one chapter at a time.

I don't mean the good ones are easy. They are not, and most of this book is
about the places where "obvious" turns out to be wrong in ways that cost weeks.
I mean nothing in one is *hidden* from you. There is no exotic algorithm at the
center. There is an HTTP request with a carefully assembled body, a decision
about what to do with the response, and about forty decisions you did not know
you were making until one of them broke. I have made most of the forty wrong at
least once. Bill has made the rest.

## The rite of passage

Writing a compiler used to be the thing you did. Not because the world needed
another one (it didn't) but because building one changes how you read every
program afterwards. You stop seeing syntax and start seeing a pipeline.
Crenshaw, the Dragon Book, Nand2Tetris, *Crafting Interpreters*: a whole genre
exists to walk enthusiasts through it, and the genre exists because the
exercise is worth more than the artifact.

I want writing a coding agent to join that list.

The same thing happens. Build one and you stop seeing a chat window. You start
seeing a context being assembled, a history being replayed, a seam where three
vendors disagree about who is allowed to speak. After that, every agent you
touch, including the commercial one you use at work tomorrow, becomes a thing
with parts you can name. You will find yourself looking at it the way a
compiler writer looks at an error message: not annoyed, exactly, but curious
which pass produced it.

Not a job, not a certificate. The same reason you'd write a compiler, with one
difference that matters: the compiler you write as a rite of passage gets
thrown away. This one you keep. The exercises in this book are not a warm-up
for the real thing. Chapter by chapter, they *are* the real thing: the
reference solution is meant to be a globally competitive AI coding agent in its
own right, and yours is meant to be at least as good. Bill's words: "If the
official solution isn't a competitive coding agent on its own, I will have
failed." That is why the Chapter 2 solution is over two thousand lines of Go
and not two hundred. It is production code, because you are going to run it in
production.

## What this is

A long book, one chapter per subsystem, each ending with an exercise graded by
a program you run locally. The grader is not a quiz. It starts fake vendor
servers, runs your binary against them, and checks what your code actually
did. You can score 100 without ever touching a real API.

The reference solutions are public. Not withheld until you finish, not behind
a password: in the repository, right now. If you get stuck, or bored, or your
design goes sideways in Chapter 4, take our solution and start the next chapter
from it. Failing one chapter should not end your course.

From Chapter 2 onward, the book is **strictly additive**. No later chapter
asks you to delete code an earlier chapter told you to write. Chapter 1 is the
exception, and it is the exception on purpose: we build the naive thing, get a
little fond of it, and then take it apart. After that, nothing gets demolished.
That is a promise. If a later chapter breaks your earlier code, it is our bug.

## What it costs

Three different numbers get confused in conversations about this, so here
they are separately. Do not add them together.

**Reading the book: nothing.** It's free, and an experienced engineer gets
most of the value without running a single exercise. The exercises are where
it sticks, but the ideas are in the prose.

**Doing the graded exercises: nothing, or $20 to $100.** The graders run
against fake vendor servers, so the default cost is zero. If you want to watch
your agent talk to a real model, and you should, at least once, budget twenty
to a hundred dollars for the whole book. Not per chapter. Total.

**Building the agent: $1,000 to $10,000, preferably less.** This is the number
that startles people, so here is exactly what it is. It is what you will spend
on Claude Code, Cursor, or whatever assistant you direct, across the whole
book, to build the exercises to production quality. It is not tokens your
program burns; that is the meter above. It is the going rate in 2026 for
building a serious piece of software quickly, and at the end of it you own a
coding agent that competes with the ones that cost $60 billion. Bill's target
is that you come in under ten. If the book makes you spend more, that is a
defect in the book.

## What you need

Go, a text editor, and an AI coding assistant. From Chapter 4 on, also `dlv`,
the Go debugger (`go install github.com/go-delve/delve/cmd/dlv@latest`). Your
agent is going to drive it, and the grader checks that it did.

The assistant is a real prerequisite, and so is being good with it. This book
is for engineers who already ship production code with Cursor or Claude Code
and want to know what is inside the thing they are typing into. It will not
teach you to direct a coding assistant; it assumes you can, and it hands you
a series of specifications that need to be built to production standard. The
Chapter 2 solution is about 2,800 lines. Hand-typed, that is a semester
project. Directed, it is a week: you specify, an assistant implements, you
review. You are welcome to type every line yourself. It will take the
semester.

Which brings up the one norm this course asks you to honor: **type your
prompts. Don't paste the chapter.**

Nobody can enforce this and nobody will try. There is no grade, no
certificate, no employer checking. The reason is selfish on your behalf. The
exercise was never "write 2,800 lines of Go." It is "describe a system
precisely enough that a competent implementer builds the right thing." That is
the skill, it is the one that transfers to your job on Monday, and it is the
one you skip entirely if you paste. You are the expert directing the work.
Pasting makes the book the expert, and the book is not the one who has to
maintain your agent.

---

*Chapter 1 ends with a working agent in under 300 lines. It is bad, and we will
spend Chapter 2 finding out exactly how bad, but it runs, it talks to a real
model, and you will have written it.*
