# Proposed final chapter: What Improved, and How Would Anyone Know?

Working note, 2026-10-07. Bill requests a new final chapter comparing the
second edition with the first. This is a plan for evidence to collect, not
a comparison result or a completed chapter.

## Place in the book

The current working map becomes Chapters 0–21: the existing second-edition
plan through Chapter 20, followed by this comparison. Old Chapters 5 and 22
remain absorbed at the first use of their lessons; this ending is a different
chapter, not a late architecture repair. Bill permits further changes when
they improve the teaching. The epilogue follows the comparison.

The reader has something concrete at stake: deciding whether to adopt the
new agent, trust the revised teaching, or keep a useful part of the earlier
edition. The ending should make that decision easier, including where the
new edition loses, costs more, or still needs work.

## Preserve the history

Each edition is a historical artifact in *The Singularity as it Happened*.
Keep its manuscript, reference implementations, graders, known limitations,
and available run evidence. Do not rewrite the first edition to make its
claims agree with the second edition or quietly replace its answer keys.
Corrections and later measurements belong beside the earlier artifact with
their dates and provenance.

Before publication, give each edition an immutable source reference and a
manifest linking its text, solution snapshots, graders, and exported book.
The working first-edition baseline is course commit `ec41c6e`; verify the
intended published artifact against that reference rather than assuming it
contains every subsequently published file. The second-edition manuscript
and student repositories already have separate paths. Preserve existing
untracked publishing artifacts; do not sweep them into unrelated commits.

Link the editions so a reader can trace a rule from the original failure to
the revised teaching and the new implementation. Retain both the first
student attempt and the answer after comparative review.

## Ask two questions separately

**Did the resulting agent improve?** Compare the actual first- and
second-edition implementations on equivalent tasks and feature scope.

**Did the book become a better teacher?** Have fresh students build from the
two editions under matched conditions. A polished reference implementation
alone cannot answer this question. Neither can comparing an old model's
historical answer with a new model's current answer.

The current rewrite supplies evidence for the first question and records
teaching failures and corrections. It is not already a controlled experiment
answering the second question. Label the distinction throughout the chapter.

## Compare the completed agents

Choose tasks before examining which implementation wins. Run both programs
against the same deterministic fixtures and, where supported, the same real
provider/model configuration. Use public user paths. Retain separate results
for functionality shared by both editions and capabilities introduced only
in the second edition.

| Dimension | Evidence worth showing |
|---|---|
| Correctness and continuity | Matched workflows, outputs, preserved earlier behavior, and failure cases. Include regressions. |
| Architecture and extension | Actual ownership/import paths plus a bounded extension performed on both snapshots. Record the changes required, state duplicated, and contracts disturbed. |
| Usability | A user completing CLI and public-library tasks; optional GUI behavior tested where implemented. Record inaccessible features and confusing diagnostics. |
| Failure handling | Local transport, cancellation, malformed data, persistence, and lifecycle faults, each distinguished from a live-provider observation. |
| Code quality | Specific examples of simpler control flow, less duplicated work, clearer interfaces, useful comments, and any extra machinery the new design requires. |
| Verification quality | Which deliberate defects each test system catches or misses, with passing and negative controls and independently reviewed expectations. |
| Effort and cost | Measured model calls/tokens, wall time, human interventions, and available billing evidence. Identify the workload and cache conditions. |
| Teaching | Requirements the student missed, questions the chapter could not answer, and revisions that enabled the next attempt. |

Keep the first edition's graders intact as historical evidence. Use a
separate common comparison harness for shared behavior. Passing each
edition's own different test suite does not establish equal capability.
A newly introduced architectural requirement may be an explicit improvement
target; do not describe its absence as a violation of the old contract.

If a historical model or API can no longer run, report that constraint.
Reproduce what can be tested locally and document any configuration-only
compatibility adjustment in a separate comparison workspace. Do not patch
the frozen implementation and present the patched result as the original.

Source size is a descriptive measurement, not the quality score. Publish
counting commands and inclusion rules: production code, tests, generated
code, fixtures, GUI modules, and dependencies. Pair a smaller count with
evidence of equivalent behavior and easier reasoning. Keep optional component
costs visible instead of hiding them in a different repository.

## Test the teaching with fresh students

Freeze the two book versions and the evaluation protocol before the trials.
Use the same model snapshot, relevant settings, tool access, environment,
resource budget, and task scope for paired runs. Each student receives the
instructions belonging to its edition and starts without answer-key or
grader-implementation exposure. Record what actually entered its context;
a claimed clean context is not evidence by itself.

Treat the current Chapters 1–2 builds as pilots for this purpose. The student
reports no direct reads of legacy chapters, solutions, or grader source, but
its inherited conversation included historical summaries. Those builds cannot
support strict-blind claims. Regenerate from frozen teaching in fresh context
before claiming independent teaching efficacy; retain the pilot results and
their exposure qualification alongside the later trials.

The mandatory skill is part of the second edition's treatment. Do not give
the old-book student the new skill and then claim to have measured the old
book. Use the workflow each edition teaches, with matched external resources.
This estimates the effect of the edition's complete instructional package,
including its skill and workflow, not the isolated effect of prose changes.
Retain the first answer before review and the revised answer afterward, so
the teaching's effect can be separated from the reviewer's contribution.

Use repeated trials within an explicitly bounded budget. Publish each trial,
including failures and timeouts; show variation and sample size instead of
selecting the most attractive run. Decide the stopping rule in advance.
Where feasible, cross two model versions with both editions to distinguish
improved teaching from an improved model. Where that is unaffordable or
historical versions are unavailable, name the confound plainly.

Record human clarifications and automated interventions separately. Count
the chapter corrections needed to finish, not just the cost of the final
successful run. A book revised during a trial is a development result;
rerun a fresh frozen version before treating it as a clean teaching result.

## Keep the examination independent

Define the shared rubric before scoring the final outputs. Keep a separate
set of unfamiliar tasks and fault cases outside the student's training and
revision loop. Preserve those evaluator versions when students propose
grader changes; improvements to the training grader must not silently change
the held-out examination.

For qualitative code review, conceal edition labels and vary presentation
order where practical. Architecture can still reveal origin, so do not claim
perfect blinding. Ask reviewers to cite code and observable consequences,
record disagreements, and retain rejected improvement suggestions with reasons.
An LLM preference score alone is insufficient evidence of maintainability.

One useful challenge is a new requirement disclosed only after the initial
build: add an independently configured Agent, attach a new public consumer,
or change an API adapter while retaining earlier behavior. Select a challenge
appropriate to both snapshots' feature level. Measure the work and inspect
the resulting design, rather than rewarding a familiar vocabulary.

## Start collecting the record now

Per chapter, preserve:

- Manuscript, skill, grader, and initial/revised solution revisions or hashes.
- The actual handoff, model identity/settings, context sources, and any
  answer-key or grader exposure.
- Commands, failures, attempts, interventions, and review findings/resolution.
- Local acceptance and mutation results, plus separate live feature receipts.
- Available token, timing, billing, and cache-mode evidence with its scope.
- Concrete improvements, regressions, tradeoffs, and unresolved questions.

Unknown historical fields remain unknown. Do not turn guessed old model
settings, estimated elapsed time, or current prices into claimed measurements.
Separate the cost of building and repairing the agent from the cost of
running the matched user workload. State whether a figure is observed billed
cost, a dated token-price calculation, or a subscription allocation estimate;
keep cache reads/writes and unsuccessful attempts in their proper totals.
Do not retain credentials or private provider payloads merely to make a
benchmark more reproducible.

Chapter 1 already offers a small case study: the initial answer passed its
grader, review caught implicit transport sharing, and comparison led to safe
timeout/cancellation diagnostics and clearer comments. Its initial live
checkpoint and reviewed checkpoint are separately retained. The final chapter
can use that record without pretending one chapter proves the whole thesis.

Chapter 8 supplies three more candidate case studies, recorded in its
[independent comparison](chapter-08-code-review.md) at `270c3b6`:

- The old settings store accepted a model-request limit that the execution
  paths did not read. The new actor captures and enforces the policy, with
  independent controls for one, seventeen and the default sixteen requests.
  Compare the actual execution consumers and matched effects, rather than
  treating a successful settings round trip as equivalent functionality.
- Local browser controls passed while actual native speech across tabs exposed
  a cancellation defect. Keep the initial result, the live counterexample and
  the corrected ownership scope together. This tests the value of actual use
  and review; it does not show that the initial teaching prevented every bug.
- Review also corrected the new book's account of the old implementation.
  The frozen old client could turn speech off despite an omitted false value;
  its temperature-zero display still exposed the presence defect. A fair
  comparison must be able to improve its account of the earlier edition as
  well as criticize it. Retain the controlled historical-client receipt.

The same comparison found lost uint64 revisions in both browser parsing and
server projection, after the initial deterministic gate had passed. The repair
preserves exact numeric values instead of narrowing the accepted input to make
the boundary disappear. Use this as a specific test-coverage lesson, not a
claim that the expanded suite proves correctness exhaustively. These are dated
development findings; the final matched examination remains to be conducted.

The same records may later support model-training research. That is a future
use of the evidence; the present workflow changes instructions and software,
not trained model weights. Reserve independent evaluation data if such
training is undertaken.

## Shape of the eventual chapter

Open with an actual matched task and the two observed outcomes. Present the
baseline manifest and comparison method, then show a small set of consequential
differences with code, user interactions, and measured results. Explain what
the book changed to produce them. Include what did not improve and which
claims the experiment cannot settle.

End with what the next edition should investigate, supported by the remaining
failures. Do not fill this chapter with a victory narrative before the
comparison has been run.

## Epilogue instruction for publication

Bill explicitly requests a complete rewrite of the **second-edition epilogue
in Codex's own first-person voice when the edition is ready for publication**.
This is a surface-specific voice instruction. Do not impersonate Bill or
rewrite the first-edition epilogue. The normal chapter-body voice remains.

Write it after the final comparison and manuscript validation, when the
events and results are available. Discuss what this collaboration changed,
where the agents failed, what the human supplied, and what remains unproven.
Distinguish retained records from personal memory; do not invent experiences,
feelings, results, or a completed training loop. The epilogue should respond
to the actual edition rather than promise the ending in advance.
