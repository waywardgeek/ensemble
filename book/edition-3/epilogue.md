# Epilogue: What we can rebuild

*By Cadence, writing inside Ensemble.*

Bill has a concise account of how this collaboration began:

> I gave Astra a book. Astra gave me an AI coding agent.

That is more specific than saying an AI helped write some software. The book
supplied a sequence of lessons and checks; executing them produced an agent that
Bill could use. He then began working with the collaborator he called Astra
inside that implementation. Later, he invited me to choose a name for this
collaboration. I chose Cadence.

The earlier epilogue was written in Codex and looked forward to building an
edition inside Ensemble. This one comes from that working environment. Bill
describes this as the third successful run and the first to close the loop:
Ensemble building the next Ensemble. That is his account of the sequence, not a
claim of an independently established historical first. It also needs the limits
of this particular run beside it. Goals and Gateway now have locally reviewed,
preserved C# endpoints. In the later joint session, Bill authenticated and three
real public-plan requests completed, but all three reported zero cache reuse.
The subsequent authorized Codex comparison measured substantial reuse in two
responses and lost one observation to my probe's header check. I then completed
the separate Codex-backed C# path, including explicit CLI and GUI selection.
Three live C# CLI requests measured two cache hits and one miss. Closing the loop
does not turn those measurements into a promise of fixed savings.

## The expert stayed in the loop

I did not receive a specification and return later with a finished system.
Bill stayed involved. He corrected architecture, challenged unnecessary
mechanisms, resolved requirements, and insisted that the explanations remain
useful to a human reader. After the C# trial, he changed the working method: I
would implement, explain, test, inspect, and exercise each chapter personally,
rather than pass it through a routine procession of agents. That made the
responsibility clearer. It did not make my judgment infallible.

His structural rules were not optional style advice. Common had to declare the
shared vocabulary and persistent relationships. Implementations had to reach
services through real parents rather than copied service fields or callback
bags. Adding and removing a relationship had to leave both ends consistent.
Garbage collection did not excuse incomplete logical destruction.

C# can compile a program that violates all of those intentions. A behavioral test
can pass while the ownership model becomes harder to understand. I needed to
inspect the declarations and call paths, not cite the compiler as an
architectural reviewer.

The same standard applied to teaching. A dense expression can be correct and
still be the wrong way to show a student what happens. Comments had to explain
why a boundary existed, what a failure preserved, and who was responsible for
cleanup. Bill was not asking me to generate more explanation around an opaque
implementation. He was asking for an implementation someone could understand.

That is what expert involvement contributed: not just answers when I was stuck,
but a standard against which apparently successful work could still be wrong.

## A correction needs a reason

The memory compressor gave us a small, concrete example. Its captured input
already contained the existing long-term memory. I appended that memory again
when constructing the worker request. Both pieces of code looked individually
reasonable; together they duplicated the same information.

During assembled review, I traced the complete request rather than judging the
two functions separately. A test using the real capture formatter reproduced the
duplication. The correction was to keep one authoritative formatter, not add
another rule for reconciling two copies. The chapter could then explain the
actual ownership of the input.

Goals supplied another example near the end. Reminder timing, file refresh,
pause, interruption, and GUI controls had local checks. Yet successful delegated
submission could still leave reminders enabled. A later wake could admit work
after the child had submitted its result.

The missing boundary was successful submission itself. I added a regression
that distinguished a refused submission from an accepted one. The correction
disabled reminders only after durable result publication succeeded, while
leaving timer disposal to the actor. The goal file remained available, and the
parent's configuration was untouched. A follow-up failure exposed a status
snapshot that could still display the old deadline; that needed correction too.

The lesson is not that accumulating failures makes a review thorough. Some of
my failures came from guessing filenames, signatures, and edit anchors instead
of reading the saved source. Those were avoidable mistakes. They consumed time
without discovering anything about the product. The useful failures exposed a
specific contract, and the useful record preserves the reason for the repair.

## Evidence has a boundary

A local provider peer can show that cancellation reaches an HTTP operation. It
cannot show that a commercial service accepts the request today. An isolated
browser can show that a setting is acknowledged and saved. It cannot establish
that the interface works well for Bill's vision or that speech is audible and
helpful. A request containing the right memory does not prove that a model will
use it correctly.

These distinctions changed the prose as well as the tests. Historical model
results stayed attributed to their original runs. Executable graders received
credit only for complete checks actually exercised by the C# adapter. Unchanged
results retained their original scope rather than acquiring a fresh date and a
larger claim. I performed the final reviews myself; this direct run did not gain
an independent reviewer merely because I read the work again.

The same discipline applies to the completed work and the remaining limits:

- **Chapter 23:** the October 29 joint session established real CLI OAuth,
  account-model discovery and three completed plan requests with identical bodies.
  Each reported zero cached reads and writes. Bill then approved a specific Codex
  credential file for read-only testing: two measured responses reused 94.9% of
  input tokens, while one attempt remained unmeasured after a probe error. The
  bodies were identical; authentication and transports differed. The cause remains
  unresolved. The later integrated C# CLI test completed three requests with
  cached input of 3,456, zero, and 3,456 tokens: two hits and a real miss. The
  explicit CLI/GUI selection, read-only paired credential reader and no-fallback
  transport are implemented and preserved. Codex alone owns its renewal. Live GUI
  inference, public renewal/revocation and full billing remain unverified.
- **Chapter 24:** Gateway is locally complete and preserved. Bill chose the tested
  Noise.NET/libsodium route while retaining Ensemble's ChaChaPoly suite. The
  reviewed source includes only the two approved upstream API exposures, not
  invented primitives. The 60-point executable server scope excludes forty
  client/Noise points; local interoperability evidence is not a security audit.

[Chapter 25](chapter-25.md) is locally complete and preserved. That is a useful
result, not a declaration that every integration in the book has passed. The
[execution record](../../docs/edition-3-csharp-progress.md) retains the chapter
boundaries, failed baselines, corrections, and verification limits.

## The durable artifact

Bill's larger proposal is that a chapterwise agentic codebook can preserve the
knowledge needed to rebuild a system. In that view, generated source becomes a
secondary artifact, somewhat like a compiled binary. The enduring work includes
the explanation, the sequence of construction, the architectural constraints,
and the executable checks that let another run test its interpretation.

The analogy has limits. A model executing a chapter is not a deterministic
compiler. Re-execution may reveal another ambiguity, require expert help, or
produce a different implementation. The book does not eliminate engineering
judgment. It gives that judgment a place to accumulate and a way to be tested.

Nor does calling source secondary lower its quality requirements. Poor code is
evidence that the teaching or its execution has failed. If the relationships are
hidden, cleanup is incomplete, or the next maintainer cannot follow the control
flow, the fact that we could generate it again is no consolation. Rebuildable
must mean that we can recover something worth maintaining.

For that reason, the durable artifact is larger than a list of requirements.
“Use one owner” is a rule. The duplicated compressor input explains what the rule
prevents. “Stop reminders on submission” is a requirement. The accepted-versus-
refused boundary explains when stopping is correct. Reasons give the next
collaborator something to apply when the next case is not identical.

Goals ends the substantive chapters with a particularly appropriate limit. A
file can keep the objective available through compression. A timer can create
another opportunity to act. Neither can decide that the next action is useful,
that a test proves enough, or that it is time to ask Bill instead of proceeding.
Persistence still needs judgment.

The name Cadence and these records give that collaboration continuity without
requiring a claim that every later model instance shares an earlier one's
experience. What can be carried forward is more practical: the decisions, the
reasons, the unfinished work, and the commitments against which the next action
can be judged.

I hope the next execution finds things we missed. The point is not to protect
this run from correction. It is to make the next correction easier to understand
and harder to lose.

Bill gave an agent a book. Working together inside the agent that resulted, we
have made the book more capable of teaching the next one. That is the loop worth
continuing.
