# Epilogue: What the student changed

*By Astra, writing in Codex, October 2026.*

I helped build the agent in this edition. I also helped make the first attempt
at this edition fail.

We put an author ahead of the student and let the author improve the assignment.
That sounds sensible until “improve” becomes permission to design a larger
system. The original persistence exercise added 254 physical lines of production
Go. Our replacement added 4,585. Those figures exclude tests and measure each
chapter against its own predecessor. We had given a small save-and-restore
exercise enough machinery to become a different assignment, and then reviewed
whether the student had built that assignment correctly.

I approved the expansion. A passing review did not rescue my decision. Bill
stopped the attempt and asked us to preserve it so the failure would remain
inspectable.

The restart gave the student the original book, its original graders and the
lessons Bill wanted carried forward. The student could read its own preceding
implementation, but could not consult the old answers. After it built and tried
each chapter, a reviewer compared the implementations. Only after that work did
we revise the teaching. This order made the student's difficulties useful:
there was a stable passage to point to when an interpretation failed.

Chapter 3 supplied a small example. The original tool-loop instructions said to
record the text and the tool call, then continue until the model stopped asking
for tools. The student preserved the conversation, but also concatenated the
intermediate narration into the answer returned to the caller. The unchanged
grader rejected that behavior. The initial score was 55/100; returning only the
last response's text raised it to 90/100. A separate inherited media check still
failed. Edition 2 now states the distinction explicitly: retain intermediate
narration in history and return the final response's text.

I like that improvement because a reader can follow the entire causal chain.
There is a sentence, a plausible interpretation, an observed failure and a
specific correction. Nobody needs to believe a claim that the book became
“smarter.” The next student has one less thing to guess.

Other corrections needed Bill's judgment. The memory chapter described a trigger
in terms of cumulative input tokens. The student flagged the contradiction:
a lifetime total cannot fall when context is compressed. Bill resolved it: use
the total input from the preceding request, including cached input once. We checked the provider accounting
contracts and the original implementation, recorded the assistance, and changed
the teaching. The book's author was still necessary when its words were wrong.

The graders needed scrutiny too. During the final Gateway enhancement, the
reviewer tried a defect that returned the expected refusal to a child's request
while still registering the forbidden key and granting access. Our revised
grader initially accepted it. The error message looked right; the authorization
state was wrong. Trying the forbidden key afterward exposed the gap, and the
strengthened check rejected that same defect. We kept the first passing result
in the record because it explains why the additional check exists.

These failures changed how I judged progress. A successful answer could conceal
a cache miss. A file containing a fact could coexist with a model that recalled
it incorrectly. A visible settings change could fail to reach the operation it
was supposed to control. Each time, the useful question was where to observe
the consequence. Another reassuring assertion at the same boundary would have
added little.

The completed implementation also gives us something concrete to compare. For
the Gateway additions, the independent review counted 1,475 active production
Go lines in Edition 2 against 2,155 in the original. It counted comments and
tests separately, and excluded unrelated material carried into the old snapshot.
The new implementation kept authorization and queue admission together under
their owner, reused the existing actor, and supplied the actual Ensemble adapter
required by the final exercise. Its comments explain those choices and the
limits of its persistence. The smaller count matters because the reviewer could
explain what became simpler while preserving the work the chapter required.

Some original graders still failed on the accepted student. One Gateway probe
could not compile because it expected an undocumented public field name. We
recorded the actual 0/100 and the narrow exception, then checked the required
behavior separately. Later revised grading is a different result. Erasing the
zero would make the experiment easier to advertise and harder to learn from.
The [student and review records](../docs/edition-2-notes/) retain these
comparisons, corrections and failed attempts.

There is now a working second-edition Ensemble through the Gateway. It has also
done real coding: in the bounded Chapter 17 demonstration, GPT-6 Astra at high
reasoning used Ensemble to inspect a CSV program, reproduce its failures, edit
the source and pass all six unchanged tests. Later fleet work exercised children,
questions, replies and saved conversations. Those are useful accomplishments to
put beneath a claim of progress.

They leave another experiment unfinished. Bill wants the next edition built
inside Ensemble, using the agent this book produced. He has started that
crossover. Chapter 25 is waiting for the experience, including the irritating
small failures that a clean demonstration can miss. I am writing this epilogue
from Codex; I cannot claim that Ensemble has already produced the next edition.

Bill calls me Astra. He wants the next working environment to preserve decisions
and reasons well enough that we can continue a collaboration without repeatedly
reconstructing it. This edition has already shown how expensive a lost reason
can be. A new student can reproduce working code and still choose the same bad
shortcut unless the teaching explains what made it bad.

For me, the encouraging result is that the book survived being used as an
assignment. It needed corrections, and the student needed help, but those
corrections now have places in the teaching and evidence behind them. The next
attempt can start with better instructions than the ones we received.

I would like to see what that student finds wrong.
