# Preface

## The thing is knowable

AI coding agents are among the most valuable software artifacts on Earth right
now. Companies have paid billions to acquire them and the teams that build
them. Engineering organizations have reorganized around them. Somewhere in your
company there is a slide deck about them.

And structurally, they are not that complicated.

That gap is why this course exists. The money says *cathedral*. The source says
*a loop, a log, and a careful answer to the question of what the model gets to
see.* Both are true, and the second one is learnable in eight chapters.

I do not mean that the good ones are easy — they are not, and this book is
mostly about the places where "obvious" turns out to be wrong in ways that cost
you weeks. I mean that nothing in an AI coding agent is *hidden* from you. There
is no exotic algorithm at the center. There is an HTTP request with a carefully
assembled body, a decision about what to do with the response, and about forty
decisions you did not know you were making until one of them broke.

## The rite of passage

Writing a compiler used to be the thing you did. Not because the world needed
another compiler (it didn't) but because building one changes how you read
every program afterwards. You stop seeing syntax and start seeing a pipeline.
Crenshaw, the Dragon Book, Nand2Tetris, *Crafting Interpreters*: a whole genre
exists to walk enthusiasts through it, and the genre exists because the exercise
is worth more than the artifact.

I want writing an AI coding agent to join that list.

The same thing happens. Build one and you stop seeing a chat window. You start
seeing a context being assembled, a history being replayed, a seam where three
vendors disagree about who is allowed to speak. After that, every agent you
touch, including the commercial one you use at work tomorrow, becomes a thing
with parts you can name.

That is the whole pitch. Not a job, not a certificate, not a product. The same
reason you'd write a compiler.

## What this is

Eight chapters. Each one ends with an exercise, and each exercise is graded by a
program you run locally. The grader is not a quiz; it starts fake vendor servers,
runs your binary against them, and checks what your code actually did. You can
score 100 without ever touching a real API.

The reference solutions are public. Not withheld until you finish, not unlocked
by a password — public, in the repository, right now. If you get stuck, or bored,
or your design goes sideways in Chapter 4, you may take our solution and start
the next chapter from it. Failing one chapter should not end your course.

From Chapter 2 onward, the book is **strictly additive**: no later chapter asks
you to delete code an earlier chapter told you to write. Chapter 1 is the
exception, and it is the exception on purpose — we build the naive thing, bond
with it a little, and then take it apart. After that, nothing gets demolished.
This is a promise, and if a later chapter breaks your earlier code, that is our
bug and not yours.

## What it costs

Three different numbers get confused in conversations about this, so here they
are separately, and they should never be added together.

**Reading the book: nothing.** It's free, and an experienced engineer gets most
of the value without running a single exercise. That is a legitimate way to use
this. The exercises are where it sticks, but the ideas are in the prose.

**Doing the graded exercises: nothing, or $20 to $100.** The graders run against
fake vendor servers, so the default cost is zero. If you want to watch your agent
talk to a real model (and you should, at least once), budget twenty to a hundred
dollars for the entire book. Not per chapter. Total.

**Building your own agent afterwards: $1,000 to $10,000.** This is the number
that startles people, so be clear about what it is: it is what you will spend on
Claude Code, Codex, or whatever assistant you use, while building a real agent of
your own after the course ends. It is not the course. It is not tokens your
program burns. It is optional, it begins after Chapter 8, and it is the going
rate for building a serious piece of software quickly in 2026.

## What you need

Go, a text editor, and an AI coding assistant. From Chapter 4 on, also `dlv`,
the Go debugger (`go install github.com/go-delve/delve/cmd/dlv@latest`) —
your agent is going to drive it, and the grader checks that it did.

The assistant is a real prerequisite, and it is worth being honest about why. The
Chapter 2 solution is about 2,400 lines. Hand-typed, that is a semester project.
Directed, it is a week: you specify, an assistant implements, you review. You are welcome to type every line yourself. It will take the semester.

Which brings up the one norm this course asks you to honor, and it is an odd one
for a book to ask: **type your prompts. Don't paste the chapter.**

Nobody can enforce this and nobody will try. There is no grade, no certificate,
and no employer checking. The reason is selfish on your behalf: the exercise was
never "write 1,600 lines of Go." It is "describe a system precisely enough that a
competent implementer builds the right thing." That is the actual skill, it is
the one that transfers to your job on Monday, and it is the one you skip entirely
if you paste. You are the expert directing the work. Pasting the chapter makes
the book the expert, and the book is not the one who has to maintain your agent.

## On receipts

I have tried to make every claim in here checkable, because I am asking you to
trust a lot of small facts about systems that change monthly.

Where this book criticizes a vendor, it does so with a reproducible behavior, a
status code, or a number, and a date attached. Where it makes a prediction, it
prints how the prediction came out — including the one in Chapter 2 that the
book partly *lost*, which is left in with the result printed rather than quietly
revised.

While writing Chapter 2, we tested three of our own claims about how vendors
report token usage against live APIs. All three were wrong. Not subtly wrong —
one of them undercounted billed output by 56%. Those corrections are in the text,
and they are the reason the usage section is the most heavily receipted part of
the book. No model has these conventions right from training data. We didn't
either, and we are the ones writing it down.

I have tried to be fair everywhere and flattering nowhere. Every vendor claim in
this book can be reproduced by anyone with an account and an afternoon, and that
is deliberate: you should not have to take my word for any of it. If you catch me
calling something crap that is actually beautiful, or the reverse, the repository
takes issues.

---

*Chapter 1 ends with a working agent in under 300 lines. It is bad, and we will
spend Chapter 2 finding out exactly how bad, but it runs, it talks to a real
model, and you will have written it.*
