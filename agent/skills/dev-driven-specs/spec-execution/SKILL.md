---
name: spec-execution
description: Execute one chapter, a chapter range, or an entire development-driven specification autonomously.
type: loadable
depends: coding visible-reasoning
---

# Execute the specification

Before starting, explicitly load [coding](../coding/SKILL.md) and
[visible-reasoning](../visible-reasoning/SKILL.md) as active instructions unless
already loaded. Do not assume the `depends` metadata activates them. If your
agent cannot load skills, ask the human to add their bodies to persistent
instructions. Reading the files once is not equivalent to keeping them active.

You are both builder and coordinator. Execute the assigned scope yourself, keeping
the coding and visible-reasoning dependencies active. Retain your work and understanding across chapters;
do not launch a new student or orchestration process for each one.

## Establish the scope

Agree on the specification version, assigned chapter or range, starting code,
separate solution workspace, grader commands, resources and authority. The scope
may be the whole book. Preserve the previous implementation and unrelated work.
Before executing the book again, read its **crossover chapter**. Incorporate every
recorded fix's lesson into the chapters that should have prevented the problem,
including their exercises and graders where appropriate. Record where each lesson
was incorporated so none is lost. This is correction from observed use, not permission
to invent requirements or redesign the specification before executing it.

A fresh execution tests the teaching without the author's private context. Do not
consult the target reference, earlier target solutions, private repair discussions,
grader source or private fixtures. Use accepted prerequisites and your own earlier
chapter work. Published crossover lessons are part of the specification, not private
repair hints or permission to inspect the previous solution. A separate directory
alone is not an access sandbox.

## Execute each assigned chapter

Read the chapter in full and inspect the existing implementation. Build the
increment, run relevant tests and the external grader, and exercise the actual
application within authority. Extend the real earlier system, not a stub or a
parallel fixture-shaped implementation. Inspect correctness, simplicity, ownership
and integration, not just a PASS label.

For a chapter trial during writing, do not launch reviewers: the author has already
completed the code and chapter reviews. Ask the author questions as needed, pass
the required graders, and return your results and chapter feedback. The human decides
whether to advance; do not begin the next chapter on your own.

During a separately authorized autonomous execution, use a separate-context
[code reviewer](../codebook-code-reviewer/SKILL.md) only when useful and authorized.
Call it once and incorporate feedback yourself, without reviewer iteration. Keep
checks and notes proportionate and reuse valid evidence for unchanged work.

Investigate failures and correct their actual cause: implementation, teaching,
grader or setup. Report suspected private grader defects rather than changing its
source. The author may correct graders to match the agreed requirements; rerun them
when corrected. Ask about ambiguity instead of inventing requirements. Return useful
clarifications to the author and identify assisted results honestly. Your repaired
code does not prove the original teaching sufficient, but that is not permission
to rerun a passing student: discuss any proposed rerun with the human first. Repeated failure calls for diagnosis,
not blind retries or weakened grading.

During a whole-specification run, when the chapter's requirements are met, continue
to the next chapter within the agreed scope. Routine completion and progress reports are not reasons to wait for
human approval. Ask when consequential ambiguity, a scope/design change, missing
authority or a resource limit needs the human. New attempts never reset budgets
or permissions. Report blocked and unperformed checks rather than counting them
as successful.

## Preserve continuity and finish

Before context is lost, retain the current chapter, accepted work, reasons,
verification limits, live work and next authorized step. Resume the same run;
a handoff does not justify restarting or repeating unchanged checks.

At the end of the assigned scope, test the integrated result and its intended use.
Return the implementation, corrections earned by execution, and a concise account
of actual results and unresolved limits, including human-only checks. Keep useful
lessons in the specification, not only private notes. Stop at the agreed boundary.
Installation, publication and replacement of a running application require separate
authority; successful execution does not grant it.

## Crossover: make the result usable and preserve every fix

Completing the book produces a candidate, not a guarantee of usability. The human
tries the resulting system and may find issues that need fixing before adopting it.
Work with the human on those fixes within the granted authority.

**Every fix made during crossover MUST be written down as prose in the crossover
chapter.** Explain what the human encountered, why it happened or what remains
uncertain, what changed and why, and how the correction was checked. A commit,
patch, private note or passing test alone does not satisfy this obligation.
Create the crossover chapter if the specification does not yet have one; write
from actual findings, not predicted problems.

Keep this record current as fixes are made, not as a reconstruction from memory
at the end. It supplies the next execution's teaching corrections. Before that run,
fold the lessons into their original construction chapters so the next agent can
avoid the same mistakes. Do not leave the next builder to reverse-engineer a patch
list. Human adoption and any installation or deployment remain separate decisions.
