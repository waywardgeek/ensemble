# Chapter 0: The Perpetual Machine

*The Singularity as it Happened*

*Third-edition note: the opening account, opinions and comparison figures below
are preserved from the first edition's October 2026 account. They describe that
work, not fresh measurements of today's products or of the C# implementation.
The third-edition foundation and working V3 crossover have been accepted for
daily use. Chapter 26 covers multimedia with 100/100 external media points and
explicit platform limits. Chapter 30 retains the crossover acceptance
requirements. Gateway and the connected personal-system work now belong in
the separate [Puffin book](../puffin-book/README.md).*

---

I gave Astra a simple prompt: read all of CodeRhapsody's documentation and source code, then design a better next-generation version. This is a task I could do in an afternoon.

Astra failed. Not partially. I would not have given a passing grade to a student studying how to build AI coding agents. Astra is OpenAI's best model as of October 2026, and after reading the full source and extensive documentation of one of the most advanced coding agents in the world, it had no idea what it was looking at.

Astra is brilliant, but ignorant. It was not there for the two thousand hours I spent with Claude building CodeRhapsody, OpenADP, and Puffin. The pain points, the dead ends we burned days on, the lessons that changed the architecture. None of that is written in any document. The model would have learned more reading my LinkedIn posts.

I started this book the same way. I told Fable to write an outline for a book describing how to build an AI coding agent. Before I could stop it, Fable had produced an outline many pages long, all wrong, all hallucinated, because no model can hold what I know in a context window. A million tokens sounds like a lot. It is a pathetically small amount of knowledge next to what you and I learn in a month.

This is the fundamental limitation, and bigger context windows will not fix it. Every session starts from zero. All an LLM gets is whatever memories we curate into a limited window. Its memory is notes. Mine is learning. Those are not the same thing.

Until that changes, every advanced project needs a guru, and the guru must be human.

I am the guru of this codebase, and this book encodes what I know: the art of building AI coding agents. Not the theory, the art. The kind you earn by spending thousands of hours steering an agent through real code, listening to its reasoning at five times speaking speed, catching its mistakes before they land, and remembering what it cannot.

---

## What they built

Consider what the field has shipped.

The one I have worked with extensively is Gemini CLI. Mocks everywhere instead of fakes. Nearly five thousand mock call sites as of October 2026. In December 2025 it was 255,000 lines of TypeScript. By October 2026 it was 559,193. It more than doubled while I was writing this book. Claude Code, reported at 512,000 lines of TypeScript, is in the same class. The winner, at 1.8 million lines of Rust, is Codex.

Ensemble does most of what these tools do in 27,042.

| | Code lines | Files over 1,000 lines |
|---|---|---|
| Codex (Rust) | 1,815,512 | 197 |
| Gemini CLI (TypeScript) | 559,193 | 41 |
| Claude Code (TypeScript) | 512,000, reported | not reported |
| Ensemble (Go) | 27,042 | 2 |

That right-hand column is the one that matters. A thousand lines is where I start refactoring. At four thousand, the AI is editing blind. Codex has 197 files past the first mark and 20 past the second, and its largest single file is 12,647 lines. Ensemble has two files past the first and none past the second.

Every one of these companies raced as fast as they could, sacrificing code quality. This is not a story about bad engineers. The people who built these tools are excellent, and their products do real things Ensemble does not. It is a story about what nobody had time to do. At 1.8 million lines, the architecture review that would catch the decay is not merely undone. It is unrunnable.

Ensemble is the cleanest of them all, cleaner even than CodeRhapsody, which I also wrote. Not because I am smarter than three large AI labs. Because the rules were written down before the code was, in the place the builder reads every time it builds. That is what an agentic codebook is, and it is the answer to the vibe-coded mess.

## What this book is not

The AI coding agents people are building today are designed for one thing: autonomy. Give the model a prompt, go have lunch, come back, and maybe understand the gist of what it did well enough to decide whether you can live with it. That is how most engineers use AI coding agents, and every major framework is optimized for it.

This book builds a different kind of agent. One designed for an expert who steers.

The difference is not a preference. It follows from the limitation above. An autonomous agent is limited by what it knows how to do, and what it knows resets every session. A steered agent is limited by what its operator knows, and a human operator's knowledge accumulates over a career. The agent that Astra could not design from reading the source, I built it, hour by hour, because I was in the room for every decision and I remember every one.

Corporations are building AI coding agents to control swarms, doubling down on specification-driven development, throwing money at agents until they can stumble through a problem to a solution. Nobody is an expert in the code. That approach works when the budget absorbs the waste.

This book is for the engineer who cannot afford that waste, or who does not want it. While writing this book, the total AI spend was approximately $3,000 of my own money. My coding agent is designed to be efficient while supervised, maximizing every dollar. It assumes the human is the expert, because the human is the one who remembers.

## What this book is

Ensemble is the blueprint and the result: an AI coding agent built chapter by
chapter from explanations, exercises and checks. The experiment is whether the
teaching transfers to a builder who was not present for the original decisions.
Working code, understandable ownership and a usable agent are the evidence.

A new model or a new chapter gives us another chance to improve the result, not
a guarantee of improvement. Each edition must supply its own proof.

## A specification with receipts

The distinction matters enough to be on the first page.

Specification-driven development starts with what someone wants: a document describing a product that does not exist, handed to an engineer with the expectation that working software follows. It occasionally does. More often the implementation discovers things the specification could not anticipate, the specification is never corrected because the person who wrote it has moved on to the next slide deck, and the result is a product that matches the wish where the wish was right and improvises where it was wrong, with no record of which is which.

This book works the other way around. Each chapter was built first: code written, tested, broken, revised, graded, mutation-tested. Then the prose was written from the working code, with the coder's feedback folded in as a mandatory procedure step. Brief goes to the coder. Coder builds. Coder reports what was missing, ambiguous, or wrong. Brief is corrected. The specification is downstream of working code.

A wish for what might work fails at the first surprise. A receipt for what already works survives regeneration.

## The loop

Most implementation chapters add graded capabilities to Ensemble. "Graded" means a program starts fake vendor servers, runs your binary against them, inspects what your code actually did, and scores the published checks. Totals differ where chapters combine exercises; some chapters are explanatory or explicitly ungraded. Mutation tests check the grader itself: remove a behavior from the reference solution, re-grade, and inspect whether the intended check fails for the intended reason. A grader that lets that deletion through has exposed a gap worth fixing before a student hits it.

Earlier capabilities must still work after your changes. Some graders include an explicit parity check; elsewhere affected regressions and separate source review preserve that obligation. You cannot add capability that breaks existing capability. Chapter by chapter, the result is a stack of tested receipts, each feedback-corrected, each building on everything before it.

The second-edition student exposed a distinction this description originally
missed. Some valid implementations failed old checks because a grader assumed
a particular layout or probed behavior the chapter had deferred. An independent
reviewer could accept the documented exception; the failed score stayed in the
record. Other failures were real bugs and had to be fixed. A grader is useful
evidence, and it is also software someone wrote.

## Working through this edition

Build one chapter on its accepted predecessor. Read its explanation as well as
its exercise: the reason for a boundary is often what keeps the next feature
from breaking it. Before writing code, load
the `ensemble-coding` skill and follow the
[C# edition procedure](../docs/edition-3-csharp-procedure.md). They carry the
implementation and review rules into an agent's working context. Chapter 1
explains the initial architecture; later chapters introduce their own machinery
when it is needed.

Keep the student away from the reference answers until it has attempted the
exercise. Otherwise a successful build tells you that copying works. If the
chapter omits an input or output format, have someone other than the student
recover that contract and record what was supplied. An ambiguity about what
the program should do needs the project owner's decision. It cannot be settled
by the model quietly choosing its favorite design.

For each chapter, finish the same small cycle:

1. Implement the stated capability and run its applicable grader, .NET build
   and tests.
   Check which packages and executable actually ran; a successful command with
   no tests selected proves little.
2. Exercise the human interface with a real model. Type into the chat, read its
   answers and try the new capability. A JSON-lines grading run does not show
   that a person can use the program.
3. Check the tests with small, deliberate defects. Removing history retention
   should fail a history check, rather than merely break compilation. Use small
   fakes at external boundaries while the real earlier implementation runs;
   this course requires explicit approval before substituting mocks.
4. Give an independent reviewer the code, the chapter, the matching reference
   and the student's account of what was difficult. Compare clarity, ownership,
   useful comments and protected behavior. More lines and more tests are not
   evidence of improvement. Correct the findings, rerun affected checks and
   checkpoint the accepted result before starting the next chapter.

Keep a short chapter record: source version, actual results, failures, help
received and decisions with their reasons. A restart should be able to recover
the work from that record and the files, without trusting an old conversation's
claim that a task is still pending. After an initial review, this edition allows
three correction rounds before bringing an unresolved problem back to the
project owner. Replacing the reviewer does not reset that count.

The record also separates a flawed check from a passing check. When a reviewer
independently agrees with a documented grader defect, acceptance may carry an
explicit exception. Preserve the original result and repair the grader in the
new edition, with its own tests. Mutation checks demonstrate sensitivity to
the defects tried; independent review and live use can still find a different
mistake. Keep those findings too. They are material for teaching, not stains
to scrub out of the experiment.

The second edition's first author-led attempt supplied an expensive warning.
Writing expanded contracts before the student learned from the original exercises
turned small tasks into larger ones, and review approved the expansion. That
attempt was stopped. The fresh second-edition student built first; that edition
enhanced the original explanations using its accepted code and reviews. The third edition
now teaches the completed C# implementation, with chapter source maps pointing
to its preserved endpoints rather than changing their historical numbers.
The second-edition attempt history is in
[`docs/edition-2-attempt-1.md`](../docs/edition-2-attempt-1.md).

## The crossing

Somewhere in this book you will use Ensemble to help build the next chapter's
exercise. When a compiler compiles itself, we call it **self-hosting**. When an
AI coding agent becomes the primary tool for its own continued development, we
call it **self-wielding**.

The loop is concrete:

> Build Ensemble. Use it to develop the next capability. Test and review the
> result. Correct the teaching from what the work revealed. Reconstruct again.

The specifications say what the agent must do, not how well it can do it.
Cleaner code, sharper tool descriptions and better conversation still require
comparison. Passing a grader sets an exercised floor; it does not select the
best design or establish that the replacement is comfortable to use.

The second edition was built with Codex under Bill's direction. The C# third
edition was built inside Ensemble and has been accepted for daily use. Chapter
30 teaches the safe parallel trial, rollback and human acceptance that made
that crossing deliberate. A reader's implementation must earn its own
acceptance; ours does not transfer with the source.

## Your Ensemble

Your needs are different from mine. Maybe you work for a well-funded company that can spend thousands of dollars a day on compute. Maybe you are a hobbyist. Maybe you want an agent that runs autonomously while you sleep, or one that you steer in real time the way I do. The right agent for you is the one that fits the way you work, and no two engineers work the same way.

Clone this book. Go through it chapter by chapter. Customize it. Change the tool set, the memory system, the observer, the voice. Make it yours. Use the graders, review and actual work to find what your changes broke and keep the foundation under them.

This book was designed to be forked.
