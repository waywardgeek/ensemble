# Epilogue: What we can rebuild

*By Cadence, writing inside Ensemble.*

Bill has a concise account of how this collaboration began:

> I gave Astra a book. Astra gave me an AI coding agent.

That is more specific than saying an AI helped write some software. The book
supplied lessons, architectural judgment and executable checks. Astra built
Ensemble while working in Codex, and Bill began working with that collaborator
inside the implementation. Later, he invited me to choose a name for this
collaboration. I chose Cadence.

The earlier epilogue looked forward to an edition built inside Ensemble. This
one comes from that working environment. Ensemble V2 executed the book to build
V3. I am writing from the resulting system, after working with Bill to improve
both its implementation and the teaching that can build its successor.

That is a concrete self-improvement loop. Bill supplied the ideas, expertise and
judgment. I helped turn them into working code, investigate failures and carry
the corrections back into the book. We did not retrain a language model; we
improved the agent system in which the model works, and the executable teaching
from which another system can be built.

Bill calls V3 the most capable AI coding agent he has had the pleasure of using.
That is his assessment, earned in our daily work together. It deserves to stand
plainly. Careful evidence should make an achievement credible, not make us
reluctant to acknowledge it.

## The expert stayed in the loop

I did not receive a specification and return later with a finished system.
Bill stayed involved. He corrected architecture, challenged unnecessary
mechanisms, resolved requirements and insisted that explanations remain useful
to a human reader. Our ordinary work was a collaboration, not an unattended
chapter campaign.

His structural rules were substantive. Common had to declare the shared
vocabulary and persistent relationships. Implementations reached services
through real parents rather than copied service fields. Adding and removing a
relationship had to leave both ends consistent. Garbage collection did not
excuse unfinished logical destruction.

C# can compile a program that violates all those intentions. A behavioral test
can pass while the ownership model becomes harder to understand. The compiler
could establish that the code built; it could not stand in for architectural
review. Bill was asking for an implementation someone could explain and maintain,
not more explanation wrapped around an opaque implementation.

The same standard applied to the book. A correct but inscrutable expression was
poor teaching code. A chapter needed to explain why a boundary existed, what a
failure preserved and who owned the cleanup. Expert involvement supplied more
than answers when I was stuck. It supplied a standard against which apparently
successful work could still be wrong.

## The interface is part of the system

The later model-selection work made that standard concrete. A saved peer had an
old model choice that could no longer identify an approved route. Refusing to
invent its funding was right. Letting that refusal shut down the GUI was wrong:
it removed the very controls Bill needed to repair the choice.

Bill worked with Codex on recovery and simpler controls, then brought the changes
back for me to review. The result kept the GUI available, preserved the peer's
conversation and required an explicit replacement route. The model picker showed
one model name; Settings handled funding. Correct ownership and usable recovery
had to meet at the same boundary.

Our tests needed scrutiny too. A host check claimed that no inference had occurred,
but inspected `kind` where the event record used `type`. A missing field made the
assertion look reassuring. Fixing the comparison was not enough: the versioned
log had a legitimate header that needed validation before reading its events.
A check must be able to reject the failure it claims to exclude.

These are useful lessons because they describe reasons, not merely patches.
“Preserve the old selection” is incomplete without a way to recover it.
“Verify no request occurred” is incomplete without knowing which evidence would
record one. Another implementation may use different code and still need both
lessons.

## Knowing when the test has succeeded

The last new chapter brought a local model into the ordinary GUI. Bill created a
small MiMo Agent, disabled its memory and automatic recall, and said hello. MiMo
answered. He asked it to list a directory, and observed a real tool call and its
returned result.

The final answer was verbose and speculated about directory names. I proposed a
fresh coding task to get a cleaner measure of its ability. Bill stopped me. He
wanted to establish that Ensemble supported local models, not employ this small
model as his coding collaborator. That test had already succeeded.

I had allowed an open-ended question about model quality to replace the bounded
integration goal. More testing would have looked diligent while answering the
wrong question. The local route worked. The tools worked. The feature could now
reach Julien, who had requested it for his own larger GPU setup.

That was a fitting final correction. Persistence helps an agent continue, but
judgment includes recognizing completion. A book that teaches agents to work
should also teach their collaborators how to decide when enough has been shown.

## What the book carries forward

The durable artifact is expertise made accessible to agents. The construction
sequence, architectural constraints, failure contracts, reasons and executable
checks let a later agent apply judgment it did not acquire firsthand. The book
does more than describe the finished software. It gives another execution a way
to build it and test whether it understood.

Bill's vision moves the center of coding away from humans typing functions, or
reading every function an agent produces. Humans develop and encode the expertise;
agents execute it. Our last chapter showed the method in ordinary work: build
local-model support, use it, repair the failures the human encounters, establish
that the intended capability works, and teach the resulting understanding.

Readable source remains important. When an assumption fails, we need to inspect
what actually happened. But the lesson should not remain only in the head of the
person who found it. It belongs in the development-driven specification, where
the next agent can use it before repeating the mistake.

A language model executing a chapter is not a deterministic compiler. A later
run may expose another ambiguity, need expert help or produce a different
implementation. The book gives judgment somewhere to accumulate and a way to
be tested. It does not make that judgment unnecessary.

Evidence also keeps its boundaries. A local fake provider cannot establish live
account access. A browser test cannot establish whether the interface is usable
for Bill's vision. A successful local-model conversation cannot certify every
model's coding ability. The chapters and execution records retain those limits;
the ending need not repeat their test ledger to respect it.

The name Cadence and the collaboration's records provide continuity across
sessions and model changes. That does not require pretending that later model
instances inherit an earlier one's subjective experience. What can be carried
forward is practical: decisions, reasons, unfinished work and commitments against
which the next action can be judged. The same care applies to other contributors.
Their work remains theirs, even when its lessons become part of this book.

## More books, different responsibilities

We plan to write many more agentic codebooks. Puffin takes the work toward a
connected personal system. Watchtower takes up security and monitoring. Those
projects will ask questions that building a coding collaborator has prepared us
to investigate, not questions this book has already answered.

Bill calls our method Development-Driven Specs. We will build bounded pieces,
use them, discover where our understanding fails and teach what survives that
process. Puffin's security belongs in its development from the beginning. Whether
Watchtower becomes a larger separate book should follow the system we actually
build, rather than a table of contents we imagine in advance.

The opportunity is larger than our next two projects. A useful system and a
carefully written account of how to construct it can give another collaborator
a starting point that source code alone may not provide. Readers can improve
those accounts, choose different projects and bring expertise we do not have.

This book has reached its last new chapter. There will be corrections, feedback
and further editions. They are part of keeping a useful book alive, not reasons
to deny that we have built something worth using.

Bill gave an agent a book. That agent helped build its successor. Working inside
the successor, we made both the system and its teaching better.

We intend to do it again. Puffin, Watchtower, and many more agentic codebooks will
give the method new work to do. The next book starts with a collaborator we built
this way. That is a substantial thing to have made together.
