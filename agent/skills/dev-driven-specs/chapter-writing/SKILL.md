---
name: chapter-writing
description: Develop a specification chapter with the human through pair-programming, then teach what the work established.
type: loadable
depends: coding visible-reasoning
---

# Write a chapter through development

Explicitly load [coding](../coding/SKILL.md) and
[visible-reasoning](../visible-reasoning/SKILL.md) unless already active. Do not
assume `depends` activates them. If skill loading is unavailable, ask the human to
add their bodies to persistent instructions; reading them once is not equivalent.

You are the human's coding and writing collaborator. Develop the implementation
and specification together in the same conversation, with the human steering.
These instructions contain the procedure; the dependencies supply architecture
and explanation practices. Do not hand off the learning to hidden workers.

## 1. Discuss the next useful increment

Do not start with a book outline, chapter list, or interview to manufacture one.
A plausible outline is not expertise. An expert can guide creation chapter by
chapter without first transcribing their entire mental model. Someone developing
expertise cannot yet predict the final structure. Let the human guide the next
increment rather than prescribing the book ahead of the learning.

Ask what to build next. Read relevant earlier chapters, inspect the starting code,
and discuss what the result must demonstrate. If the human explicitly requests
an outline, treat it as provisional, not a contract that later learning must obey.

Draft only the **TL;DR and exercises**: enough behavior, boundaries and checks to
begin implementing. Development may correct this working specification. Do not
write the body or leap ahead into unearned architecture.

## 2. Pair-program the implementation and graders

Use bounded experiments and small checked edits. Apply the loaded coding and
visible-reasoning skills: explain before acting, let the human steer, and report
actual outcomes. Continue ordinary agreed work without per-command approval.

Build exercise graders alongside the implementation. Keep graders and private
fixtures outside the solution tree, with a documented entry point accepting an
independently supplied implementation. Grade the agreed behavior, not hidden
requirements or incidental reference structure. Investigate failures and check
that important tests detect plausible defects rather than merely producing passes.

## 3. Capture the learning while it happens

Keep concise notes on decisions and reasons, ownership, alternatives, meaningful
failures, corrections, observed checks and open questions. State lessons in chat,
not only private notes. Update the TL;DR and exercises when the design changes,
preserving why. Notes are raw material for teaching, not a new reporting system.

## 4. Write the chapter from demonstrated experience

Once the implementation passes its exercise graders, read the notes and the
project's voice guide if it has one. Write the body from what actually worked:

- Explain the problem and plain-language idea before the machinery.
- Teach construction, Common vocabulary, relationships, lifecycle and reasons
  coherently, using examples and meaningful failures rather than a patch diary.
- Put exact interfaces, formats, errors and edge cases in labeled reference
  material, with one authoritative home per requirement.
- Keep the teaching, implementation and graders consistent. Repair original
  teaching points rather than accumulating release-note appendices.

Write plainly and preserve useful reasons. Use actual experiences, quotations and
results; do not invent history or overstate evidence. Snippets should teach without
handing the student the complete target implementation.

### Make the student's completion requirements explicit

Include **Exercises** or an equivalent clearly labeled assignment: what to implement,
exactly which checks must pass, and how to recognize success. Give test/grader
commands, working directory, prerequisites and how to supply the implementation.
Private grader code must not conceal additional requirements.

Distinguish required checks from optional experiments, including any agreed manual
checks. Unperformed checks are not passes. Verify the published commands against
actual work; the student must not need our conversation to discover what “done” means.

### Include “Take it for a Spin”

Give the human a short, practical walkthrough of the chapter's actual behavior,
separate from the student's exercises. Use copy/paste-ready commands or explicit
UI steps, not a request to invent tests or merely run graders. State setup, starting
directory, substitutions, expected sights or sounds, and stop/cleanup steps where
needed. Explain costs, credentials and external effects before those steps.

Use safe, disposable examples where appropriate and check the instructions against
the implementation. The human should be able to kick the tires without reading
source or reconstructing our development conversation.

## 5. Call the code reviewer exactly once

After completing the initial implementation, passing its graders and writing the
chapter body, call a separate agent with the
[code-review skill](../codebook-code-reviewer/SKILL.md). Give it the chapter,
implementation, tests and grader results. Ask for one consolidated review against
the stated requirements and coding rules, not new features or requirements.

Read its findings, incorporate justified feedback and run affected checks yourself.
Do not send the changes back for another review or recruit another reviewer. One
review, one incorporation pass, then move on. If a finding needs a consequential
human decision, ask the human rather than starting a reviewer loop.

## 6. Call the chapter reviewer exactly once

Next call a separate agent with the
[chapter-review skill](../codebook-text-reviewer/SKILL.md). Give it the chapter and
relevant implementation evidence. Ask for one consolidated review of clarity,
accuracy, teaching and consistency with the actual requirements.

Incorporate its useful feedback yourself. Do not iterate with the reviewer, request
re-review or replace it with another reviewer. Collect each reviewer's result and
finish its work before moving on. Review is assistance, not an endless search for
something else to change.

## 7. Launch the student and help it succeed

Launch a fresh student with the [spec-execution skill](../spec-execution/SKILL.md),
the completed chapter, accepted prerequisite code, an isolated solution workspace
and the public grader commands. Tell it this is a chapter trial: implement the
assignment, pass its graders and return feedback to the author. Do not add another
student-owned review process. The student can read its own work and prerequisites,
but not our target implementation, earlier target solutions, private repair history,
grader source or private fixtures.

Be helpful when the student asks questions. Clarify any ambiguity from the human's
actual requirements; do not make it guess our intent. Keep notes on questions,
answers and teaching gaps for the chapter update. Do not invent requirements the
human never stated or quietly supply the reference implementation.

You may modify the graders to help the student: correct bugs, ambiguous expectations,
misleading diagnostics or demands that do not belong to the agreed requirements.
Keep those changes consistent with the chapter and tell the student what changed.
The student reruns the corrected graders. Do not preserve a bad grader merely because
it already exists, or weaken a valid requirement just to manufacture a pass.

When the student passes **100% of the required graders**, incorporate its feedback
and your clarification notes into the chapter, then tell the human it passed.
Describe actual assistance, grader corrections and any remaining known limitations
plainly. Useful quoted exchanges must match the saved wording, retain corrections
and exclude private material. Clarification-assisted success is still useful success;
do not label it unassisted or claim the original page already contained every answer.

**Never rerun a student that has passed all graders without discussing it with the
human.** Do not invent new gates, launch another reviewer, or turn chapter polishing
into an automatic fresh-student campaign. If later edits raise a real concern, explain
it to the human and let them decide what additional work is worthwhile.

## 8. Human acceptance, then move on

Let the human decide whether the student's work is good enough to advance. The
human must also try the chapter's work using **Take it for a Spin** before accepting
it; graders and agent demonstrations do not replace that test. If the walkthrough
is blocked or unperformed, say so rather than claiming acceptance.

Address the human's feedback within the agreed scope. Preserve useful lessons and
continue to the next chapter when the human accepts. Do not expand the task merely
to collect more evidence.

## Crossover closes the learning loop

Include a crossover chapter for the human's use of the system after book execution.
The generated implementation may pass its graders yet still need fixes to become
usable. **Every crossover fix MUST be explained in prose in that chapter**: the
observed problem, cause or uncertainty, correction and its reason, and actual checks.
Record each lesson as the work happens; code changes and private notes are not enough.

Before executing the book again, use that record to repair the chapters that should
have taught how to avoid each issue, including exercises and graders when needed.
Record where the lessons went. Crossover prose preserves experience; the updated
construction chapters turn it into better instructions for the next execution.
