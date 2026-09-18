# Chapter 0: The Blueprint

*The Singularity as it Happened*

---

When a compiler written in its own language can compile itself to a
fixed point, stage N's output identical to stage N+1, we call it
self-hosting. The milestone matters because it proves the tool is
sufficient for its own domain.

When an AI coding agent is capable enough that it becomes the primary
tool for its own development, we call it **self-wielding**. This book
builds that agent.

The thesis is not the agent. It is what happens after.

## The loop

Each chapter adds one graded capability to Ensemble. "Graded" means a
program runs your code against fake servers and scores it, 100 out of
100 or not. Mutation tests verify the grader is not decorative: delete
a behavior and exactly the right checks fail.

Every chapter includes a parity check: all previous graders must still
pass. You cannot add capability that breaks existing capability. The
result is a stack of tested specifications, each a receipt for working
code rather than a wish for code that might work.

That stack is a blueprint. Not the kind a VP writes on a whiteboard.
The kind that was built, tested, torn apart by the coder who built it,
revised, and tested again. The procedure has a feedback loop baked in:
brief goes to the coder, the coder builds, the coder's feedback comes
back, the brief is corrected. The specification is downstream of
working code, not upstream of hoped-for code.

## The generation

Hand this book to the next generation of language model. It builds
Ensemble from Chapter 1, scores 100 on every grader, and produces a
working agent. A better model produces a better agent: cleaner code,
sharper tool descriptions, tighter error handling, while hitting the
same graded floor.

When a new capability appears in the field, write a chapter. The
chapter comes with a grader, mutation tests, and a parity check
against everything before it. Ensemble gains the capability. The book
grows. The next build incorporates it.

No version of Ensemble is final. Each is a phenotype expressed from
the same genome in the environment of whatever model reads it. The
graders are the immune system: they reject any build that loses a
capability a previous chapter established. The prose is the teaching.
Together they define a living standard for what an AI coding agent is.

## The crossing

Somewhere in this book you will use Ensemble to help build the next
chapter's exercise. The agent you have been building becomes the tool
you build with. That is the crossing. After it, every chapter you add
is written with the tool the chapter extends.

The loop closes. The blueprint renews itself. The agent is
self-wielding.
