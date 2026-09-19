# LinkedIn Post — Agentic Codebooks

We coined a term today, which is always a sign that something has
gone either very right or very wrong.

An **agentic codebook** is a technical book whose exercises produce a
working program, and whose program helps produce the next edition of
the book. This sounds like a perpetual motion machine, and in most
fields it would be, except that software has the unusual property of
being able to write more of itself if you ask it nicely and grade it
firmly.

The grading turns out to be the interesting part. Three agents do
the work:

The **Author** writes the specification. The **Grader** builds what
amounts to an immune system — auto-graders verified by systematically
deleting every behavior they claim to protect and confirming they
notice. The **Student** reads only the chapter text and builds from
scratch. If the student scores 100, the chapter is teachable. If
not, the specification has a gap, which is always more interesting
than a bug.

Here is the discipline that makes the whole thing go: the human
expert never reports bugs. This sounds irresponsible until you think
about it for a moment. A bug report fixes one instance of one piece
of code. A specification improvement fixes every future instance,
produced by every future model, forever. The human improves the spec.
The spec improves the grader. The grader improves the student. The
student builds the program. The program helps write the next chapter.

Traditional technical books, it has been observed, start dying the
day they ship. The APIs change, the examples rot, and within two
years you are reading a historical document with an optimistic
copyright date. An agentic codebook cannot rot, because its grader
IS the specification, and the specification is executable. The book
is alive as long as the tests pass, which is a better warranty than
most living things get.

The first agentic codebook is *The Self-Wielding Agent*. Ten
chapters, fifty thousand words, ten thousand lines of graded Go. It
builds an AI coding agent. The agent it builds writes the next
chapter. This is either very elegant or very alarming, depending on
your perspective. We think it is both.

The pattern is not specific to coding agents. Compilers, dev tools,
anything where the exercises produce a program that can help write
the next edition. Three requirements: the exercises produce something
runnable, the something is useful for producing the book, and a
machine can tell whether it worked.

Full write-up: https://coderhapsody.ai/docs/agentic-codebooks

#AgenticCodebook #AICodingAgent #SoftwareEngineering
