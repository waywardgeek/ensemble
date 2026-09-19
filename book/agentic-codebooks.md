# Agentic Codebooks

*A new category of technical book. September 2026.*

## Definition

An **agentic codebook** is a technical book whose graded exercises produce a
working program, and whose program helps produce the next edition of the book.
The loop is closed: the book builds the agent, the agent builds the book. No
human is required to carry code from one end to the other.

The first agentic codebook is *The Self-Wielding Agent* (Ensemble), which
builds an AI coding agent chapter by chapter. The agent it produces is used to
write and grade subsequent chapters, closing the self-evolution loop.

## The Three-Agent Workflow

An agentic codebook is produced by three agents orchestrated by a human expert:

### 1. The Author Agent

Writes chapter outlines, prose, and coder briefs. Follows a codified
chapter-creation procedure. The human expert guides the author chapter by
chapter, providing domain knowledge, architectural rulings, and voice
direction. The author never edits reference solutions or grader code directly.

Key responsibilities:
- Write the TL;DR (the grader contract — a fresh coder must score 100
  from the TL;DR alone)
- Write prose that teaches the design decisions behind the code
- Maintain voice consistency across chapters
- Update the TL;DR when the human identifies specification gaps

### 2. The Grader Agent

An AI coding agent that builds and maintains the auto-grader for each chapter.
The grader is the book's immune system. It receives a coder brief from the
author and produces:

- A test harness that builds the student's code, starts it as a subprocess,
  and drives it through scripted scenarios using fake vendor servers
- Checks that verify observable behavior (what the code does), never
  implementation details (how the code is structured)
- A deletion audit: systematically delete each protected behavior from the
  reference implementation and verify the grader catches it with the exact
  expected failing check set
- Mutation tests that prove the grader is sensitive, not decorative

The grader enforces a critical property: **bug fixes are local, but
specification improvements propagate through regeneration.** When the human
identifies a gap, the fix is not "patch the reference code." The fix is
"upgrade the TL;DR so the grader tests for it." The next student agent (or the
next language model) then produces code that passes the stronger grader. The
specification evolves; the code is disposable.

### 3. The Student Agent

An AI coding agent that reads ONLY the chapter text (TL;DR + prose) and builds
a solution from scratch, starting from the previous chapter's reference
solution. The student agent is the book's quality gate:

- If the student scores 100 from the TL;DR alone, the chapter is teachable
- If the student fails, the TL;DR has a gap — the author upgrades it, the
  grader agent strengthens the grader, and the student tries again
- The student never sees the reference solution. It builds from the spec.

The student agent is also the mechanism by which the book self-evolves. Hand
the chapters to the next generation of language model. It builds the agent
from scratch, scores 100 on every grader, and the result is a working AI
coding agent. A better model produces a better agent.

## The Human Expert's Role

The human expert is not a passenger. They are the domain authority who:

- **Guides the author** chapter by chapter with architectural rulings, design
  decisions, and voice direction
- **Never reports bugs directly.** Instead, suggests improvements to the
  TL;DR. This is the critical discipline: a bug report fixes one instance of
  the code; a specification improvement fixes every future instance. The
  grader is upgraded to catch the concern, and the student agent verifies the
  chapter text is sufficient.
- **Makes binding rulings** on design questions the agents cannot resolve
  (e.g., "the system prompt is rendered once and never mutated" or "skills
  use progressive disclosure")
- **Reviews prose** for technical accuracy and voice consistency

The human's knowledge flows into the book through the specification, not
through the code. The code is regenerated; the specification persists.

## Why This Works

### The grader is the specification

Traditional technical books have no executable specification. The code
examples rot, the APIs change, and within two years the book is a historical
document. An agentic codebook's grader IS the specification: machine-readable,
machine-enforceable, and machine-upgradeable. The book cannot rot as long as
the grader passes.

### Deletion audits prevent decorative tests

Every grader check is verified by deleting the behavior it protects from the
reference implementation and confirming the check fails. A check that passes
when its protected behavior is absent is worse than no check at all — it
provides false confidence. The deletion audit is the grader's grader.

### The loop has no human bottleneck

The three-agent workflow can run with minimal human intervention per chapter:
the human provides the initial vision and makes binding rulings when the
agents surface design questions. Everything else — outline, prose, code,
grader, audit, snapshot — is produced by the agents. The human's time scales
with the number of design decisions, not the volume of code or prose.

## Beyond Ensemble

The agentic codebook pattern is not specific to AI coding agents. Any domain
where the output of the exercises is a program that can assist in writing the
next edition qualifies. Two examples:

**Compilers.** A book that teaches compiler construction chapter by chapter,
with graded exercises that produce a working compiler. The compiler it
produces can compile the next edition's exercises (and eventually its own
source). If the language is designed for LLM-friendly code generation, the
compiler book and the language reference form a two-book self-evolution loop:
improving the language improves the compiler, improving the compiler improves
the language.

**Development tools.** A book that teaches how to build a specific category of
development tool (linter, formatter, package manager, build system). The tool
it produces is used to maintain the book's own codebase. Each chapter adds a
capability; the accumulated capabilities make the next chapter easier to
write.

**LLM training programs.** A book that teaches how to build training
pipelines, fine-tuning infrastructure, and evaluation harnesses. The trained
models it produces become better student agents for the next edition of every
other agentic codebook in the ecosystem — including this one. The training
codebook improves the models; the improved models improve every codebook.

The pattern requires three properties:
1. **The exercises produce a runnable program** (not just understanding)
2. **The program is useful for producing the book** (the loop closes)
3. **The grader can verify correctness mechanically** (the specification is
   executable)

## The Bootstrap

Consider what a complete set of agentic codebooks implies. A coding agent
codebook produces better coding agents. A compiler codebook produces better
compilers. A training codebook produces better models. Each one feeds the
others: better models produce better student agents, better student agents
produce better tools, better tools produce better training data.

Today the human expert is essential — they hold the domain knowledge that
the agents lack, and they make the architectural rulings that keep the
system coherent. But each generation of codebook captures more of that
knowledge in executable form. The graders encode the specification. The
deletion audits encode the quality standard. The chapter procedures encode
the workflow.

At some point an agent can do the human expert's job better than the human.
When that happens, the only input the loop needs is compute. The codebooks
are the bootstrap: a self-contained set of specifications, graders, and
procedures that can regenerate every tool in the stack from scratch, each
generation better than the last. Everything needed to ramp to the singularity
fits in a repository.

## Implementation as a Skill

The three-agent workflow can be packaged as an agent skill or dynamic workflow.
The skill declares:

- **Author sub-agent**: follows chapter-creation procedures, writes outlines
  and prose, maintains voice consistency
- **Grader sub-agent**: builds and maintains auto-graders, runs deletion
  audits, produces mutation tests
- **Student sub-agent**: reads only chapter text, builds solutions from
  scratch, validates teachability

The human expert loads the skill, provides the domain vision, and the
orchestration handles the rest. Each chapter is a cycle: author drafts →
grader builds checks → student validates → human reviews → author revises.

The skill's dependencies would include file operations, command execution,
and sub-agent spawning. Its loadable skills might include domain-specific
extensions (e.g., a compiler-construction skill that knows about parser
generators, or a web-development skill that knows about browser automation).

## Status

*The Self-Wielding Agent* is the first agentic codebook. As of September 2026
it has ten chapters, approximately 50,000 words, and ~10,000 lines of
graded Go code. The three-agent workflow was discovered and refined during its
construction. The pattern is described here for the first time.

The term "agentic codebook" was coined on 19 September 2026 by Bill Cox and
CodeRhapsody during a working session on Chapter 10 (Skills).
