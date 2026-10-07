# Chapter 0: The Perpetual Machine

## 0.1 Why this book exists

I gave Astra a simple prompt: read all of CodeRhapsody's documentation and
source code, then design a better next-generation version. This is a task
I could do in an afternoon.

Astra failed. Not partially. I would not have given a passing grade to a
student studying how to build AI coding agents.

I am the guru of this codebase, and this book encodes what I know: the art
of building AI coding agents. Not the theory, the art. The kind you earn
by spending thousands of hours steering an agent through real code,
listening to its reasoning at five times speaking speed, catching its
mistakes before they land, and remembering what it cannot.

## 0.2 What has to survive the session

That account comes from the first edition. Its useful result is a question
for the next builder: which decisions can be recovered from source code,
and which depend on knowing why that code exists? A model can read an
interface without knowing which expensive mistake made the interface
necessary. A human joining the project has the same problem.

The book records the reasons alongside the requirements. A parser needs
a route to its application's logger before a malformed reply arrives.
An event log must retain the model that produced a token count before a
user switches models. These rules have consequences the builder can
inspect and tests can challenge. They become part of the working method,
instead of a conversation someone hopes the next session will remember.

The first edition also compared competing agents by source-line counts,
large-file counts, and mock usage. Those are dated measurements, with
different inclusion rules and product scopes; they are not current
benchmarks for this edition. Source size alone cannot establish design
quality. The useful comparison asks where state lives, whether behavior
is duplicated, how failures become visible, and what a change forces the
next engineer to understand.

## 0.3 What this book builds

Ensemble is an agent for an engineer who steers. Its user needs to see what
is happening, interrupt work when a premise is wrong, and carry decisions
forward. Useful autonomy fits inside that relationship: the agent can
continue understood work while the human settles a question that changes
the design. An unanswered question does not authorize a guess about an
architectural requirement.

The book is also a way to rebuild the agent. Each chapter supplies teaching,
an exercise contract, executable checks, and evidence from an actual run.
A new implementation can challenge the teaching: a missing ownership rule
or an impossible user path is a defect in the chapter as well as in the
answer. Correct both, then run the relevant checks again.

A language model does not receive a guarantee that this will work on its
first attempt. Neither does a human reader. The repeatable part is the
correction process: make the requirement visible, test it independently,
observe the real program, and preserve what the attempt taught.

## 0.4 Start as a reader

You do not need to operate a publishing team to work through the book.
Read Chapter 1, implement its exercise, and use its checks. A coding
assistant can be the student while you review its choices, or you can
write the code and ask an assistant to review it. If the same assistant
has already studied the answer, describe the work as a guided build;
it is no longer an independent test of the chapter's teaching.

The construction order is deliberate: learn the rules, build the small
program correctly, then extend it. The book explains a bad design's cost
without asking you to build it first. Chapter 1 uses a narrow conversation
representation; Chapter 2 replaces that representation with a durable
event model while retaining the ownership and package boundaries.

Before any coding task, including tests, refactors, and grader changes,
read the entire repository skill:

[`book/edition-2/skills/ensemble-coding/SKILL.md`](skills/ensemble-coding/SKILL.md)

Read it again after context compaction or a change of coding task. Every
delegation must name that path and require the full read before editing.
The chapter contract and [architecture decisions](architecture.md) travel
with it. Automatic skill discovery and a remembered summary are insufficient
for this workflow. The skill supplies instructions; it does not prevent a
model from violating them. Checks and independent review must still inspect
the result.

Chapter 1 explains the method before its first code example. Shared data
and interfaces live in `internal/common`. Packages containing behavior
import that common vocabulary and do not import sibling implementations.
Model work belongs in `internal/llm`; use free functions there when Go's
method rules would otherwise draw behavior into `common`. Keep the narrow
methods required for standard-library interface dispatch with their types.

Each child receives an interface back-pointer to its actual owner. Engine
reaches Agent configuration through Agent and Ensemble's logger through
Agent's parent. A logger global or a growing bag of injected sibling
services bypasses that design. Even a stateless helper can need diagnostics;
changing a method into a function must not cut its route to the logger.

One runtime Ensemble can own many Agents. This object is different from
the shared package at the center of the import star. Starting in Chapter 2,
the CLI and a GUI stub use Ensemble's public client interfaces; GUI and
WebSocket code belong in a separate optional Go module. Observations
report events, while explicit parent methods request actions and services.
The core must work without a GUI dependency.

## 0.5 Choose the workers for the job

Independent roles become especially useful when testing whether the book
can teach a fresh student, or when extending the book itself. Give each
worker a specific output and an explicit boundary:

| Role | Work and boundary |
|---|---|
| Coordinator | Sequences the work, relays questions to the human, resolves source conflicts, and maintains progress. It may also run independent grading. |
| Author | Reads the chapter's history and later lessons, writes the teaching and executable contract, and incorporates findings. It does not implement the student's answer. |
| Student coder | Builds from the supplied contract and preceding new chapter. It reports gaps, runs the actual program, and supplies sanitized evidence. It does not inspect the old answer or grader internals to guess requirements. |
| Grader engineer | Derives behavioral and structural checks from the published contract, preserves regression coverage, and tests the checks by deleting protected behavior. |
| Global reviewer | Reads the textbook, records dependencies and late discoveries, and identifies lessons that belong in the present chapter. |
| Code reviewer | After the student's initial build and run, compares it with the first-edition standard and returns concrete design, clarity, comment, and behavior findings. |
| Proofreader | Independently checks the revised chapter's voice, explanations, contracts, and evidence, then checks the author's corrections. |

These are responsibilities, not a demand for seven simultaneous agents.
For example, a host offering three worker slots besides the coordinator
can run an author, a student, and a reviewer. The reviewer first records
the whole-book map, later performs code comparison, and then proofreads
the author's revision. Record which role each review fulfilled. Other
hosts can schedule the phases differently; slot limits and resumption
behavior are properties of the host, not assumptions of the book.

Ask the global reviewer a question with a concrete consequence: which later
chapter needed information that this data structure currently discards?
In this edition, model provenance and ownership rules move to their first
use because later repairs exposed the cost of omitting them. The reviewer
should cite the relevant chapter and passage, not merely recommend adding
every future feature now.

Reading the whole book does not guarantee that its entire contents remain
in a worker's active context. Save a read ledger, dependency map, decisions,
and source pointers. Reload the source for a detailed review after
compaction. A durable map helps find evidence; it is not a substitute for
the evidence it names.

When authoring or revising the book, both author and proofreader read all
of [`book/voice.md`](../voice.md) and
[`book/chapter-writing-procedure.md`](../chapter-writing-procedure.md).
The former governs how the explanation reads; the latter governs the
work required before a chapter can be called validated. An ordinary reader
can follow the chapter without managing all the author's research files.

## 0.6 Give the student enough, then let it build

A cold build tests the teaching only if the student receives the teaching
and stays away from the answer. Supply Chapter 1's architecture rules,
the current and earlier second-edition contracts, all formal schemas and fixtures
the chapter says are required, the coding skill, and the preceding
second-edition solution. Chapter 1 has no preceding implementation.
Do not cut the handoff down to a TL;DR that refers to omitted definitions.

Start the student in a fresh context without inheriting the coordinator's
conversation. That conversation may contain the very old-book passages and
answer code being withheld. Use a new student context for each chapter and
record its actual reading. The student can keep its context while repairing
that chapter and receiving the later comparative review.

This is a reusable handoff. Replace the angle-bracketed fields with actual
paths and the chapter number:

```text
Implement Chapter <N> in solutions/edition-2/main/.

Before editing, read the entire
book/edition-2/skills/ensemble-coding/SKILL.md.
Reload it after context compaction or a change of coding task.
Read book/edition-2/architecture.md, Chapter 1's architecture sections,
and the current chapter contract, including its schemas and fixtures.

For Chapter 1, start a fresh Go module within the outer Ensemble repository.
For later chapters, the coordinator prepares main at the accepted preceding
second-edition source version <previous-commit-or-tag>.
Do not initialize a nested Git repository or edit frozen chNN exports.
Do not inspect future main history to infer the answer.
Do not read first-edition chapters or solutions, future solutions, grader
implementation, or author/reviewer research notes. Historical links in the
coding skill are research sources for those other roles, not prerequisites
for this student. Do not edit agent/ or first-edition solutions.

Report missing requirements to the coordinator before affected code.
Continue work that does not depend on the answer. Preserve unrelated work.

Run the chapter's build, tests, vet, and grader. Exercise every implemented
feature through the real user interface with the supported real models.
Keep credentials out of commands, logs, prompts, evidence, and commits.
Return exact commands/results, a feature checklist, sanitized live receipts,
contract gaps, and the initial student checkpoint for independent review.
```

The coordinator gives questions a route to the human. A worker reports the
ambiguity, the source passages in conflict, the implementation affected,
and a recommendation. The coordinator asks once and relays the answer to
the workers it affects. Workers can continue independent work while waiting;
elapsed time is not an answer. A platform with direct worker-to-user
questions may use that facility, but the shared decision still needs to
reach the other workers and the recorded contract.

The isolation is an instruction boundary in a shared repository. It is
not a claim that tools make the old solution unreadable. Record accidental
answer exposure instead of presenting the resulting build as cold.

## 0.7 The loop has more than a score

The second edition teaches the contract before the student implements it.
That contract is provisional until the build and real use challenge it.
The chapter cycle has these gates:

1. Derive the teaching from the existing chapter, history, audits, and
   later lessons. Resolve ownership before affected code begins.
2. Publish a concrete contract with failure behavior, user paths, fixtures,
   and a mapping from requirements to checks. Let the student build from it.
3. Run builds, tests, vet, the chapter grader, and applicable earlier
   regressions. Repair missing teaching alongside missing behavior.
4. Audit advertised properties by deletion, keeping a passing control.
   A mutation must fail the intended checks for the intended reason.
5. Take the actual program for a spin with real models and record the
   user inputs, outputs, measured usage, and limitations.
6. Preserve the first student checkpoint. An independent reviewer compares
   it with the corresponding first-edition standard, then the coder and
   author revise from the findings. Review those revisions too.
7. Proofread the complete chapter against the contract and receipts.
   Resolve material findings, checkpoint the validated answer, and continue.

A grader starts local fake services, runs the submitted binary, and
inspects what it did. Fakes let it request the same edge case repeatedly.
Mutation audits ask a different question: would a missing behavior be
noticed? If deleting multi-block text assembly still passes because the
server always sends one block, the test has not protected that behavior.
Chapter 1 carries the historical audit that exposed exactly this gap.

Keep positive and negative controls independent. A redaction check needs
to see the original content before it can prove that redaction removed it.
Check-level coverage cannot establish every promise inside a broad check.
When a shared grader changes, record its old baseline and retain passing
legacy checks. A frozen reference solution may have a frozen model table;
changing the main fixture catalog does not update that snapshot.

From the course repository root, Chapter 1's grader accepts the student's
module directory:

```sh
make grade-dir CH=1 DIR=solutions/edition-2/main
```

Inside each affected Go module, run its tests and vet; a test at the course
root does not automatically test a nested solution or GUI module. Confirm
that `gofmt -l` prints no changed Go files. Record the actual commands and
results, including failures. Do not turn an unrun checklist into a report
of successful validation.

The independent code comparison happens after the initial student build
and run. The reviewer reads the old standard; the student receives
findings and reasons, not code to copy. Compare ownership, unnecessary
machinery, duplicated work, comments, diagnostics, and public usability at
the same feature scope. The older solution may explain an invariant well
and still violate the new architecture. Retain the explanation and correct
the architecture.

Retain the initial answer, the criticism, the revision, and the checks that
distinguish them. That record captures the expert judgment behind the change
and lets another reader test it. Skills and chapter instructions guide the
current session; this workflow does not update a model's trained weights.

A score of 100 sets a floor for the behavior actually checked. It cannot
prove that the code is clear, that every chapter promise is covered, or
that another model would produce the same answer. Record specific
improvements and justified tradeoffs instead of treating a newer model
or a passing score as evidence of better engineering.

## 0.8 Actually use it

Every chapter's “Taking it for a spin” comes from using its real interface
with a real model backend. Initially that interface is the CLI. Exercise
each supported provider as its adapter enters the book: Chapter 1 uses
the Messages API, and Chapter 2 adds Chat Completions and generateContent.
The multi-provider chapter needs all three paths to work.

Starting in Chapter 2, the CLI also needs a human chat mode. The reader
launches it, types ordinary text, reads the answer, and follows up in the
same conversation. The chapter explains how to do that without constructing
JSON. The coder must use that same mode through an interactive terminal,
observing replies before sending follow-ups, with real models on all three
provider paths. A PTY lets the coder exercise the terminal interface directly.
Piped JSON sessions test the machine protocol and remain separate evidence.
Neither those sessions nor a direct debugger run outside Ensemble proves
that a person can use Ensemble interactively. Record actual coder runs;
do not attribute them to a reader or to Bill.

Before a run, map every chapter feature to an action a user can perform
and an observable result. Use an external executable consumer when the
feature belongs to the public library rather than the CLI. A direct call
to an internal helper is not the same demonstration. Label the Chapter 2
GUI as a stub; its public integration check does not establish a working
browser or WebSocket transport.

Discover an available model through the provider's current documented
interface. The model printed in an old transcript is a dated observation,
not a safe default. Keep fake fixture IDs separate from live selections.
Bound the demonstration, record actual usage, and avoid an unattended retry
loop that converts a setup error into repeated paid requests.

Use your own authorized credential source. Supply the required variables
through a trusted credential manager or the child process environment;
do not paste keys into a command, prompt, transcript, or repository file.
Read only the values needed for the selected provider and keep retained
request evidence free of authorization material. A chapter's invocation
can state which variable was supplied without recording its value.

Keep the source checkpoint, build command, date, selected and returned model
identities, mode, inputs, outputs, and measured counters with the evidence.
Save both successes and limitations. Deterministic local faults can test a
timeout or broken payload that a real provider does not reliably produce,
but label them as local faults. They cannot replace the required live
feature demonstration.

When a provider is unavailable, record the blocker and leave the affected
validation incomplete. A plausible transcript is no receipt. The author
writes the demonstration after the run, and the proofreader checks its
numbers and claims against the retained output.

## 0.9 Leave a continuation point

The reference workflow develops the new agent in `solutions/edition-2/main/`,
tracked by the outer Ensemble Git repository. Chapter 1 starts from scratch;
later chapters extend the accepted preceding second-edition version in that
same tree. Readers working elsewhere can keep their own main tree. Preserve
the first-edition implementations and existing `agent/` tree as evidence.

After validation, export that exact source version to
`solutions/edition-2/chNN/`. These are self-contained Go modules and frozen
ordinary tracked directories, without nested Git repositories. The student
works in main; the coordinator produces the export with source and file hashes.
Earlier standalone student repositories belong to the historical record:
preserve their histories and unfinished work before consolidating them.

Commit the source before exporting it. From the repository root, use two
output paths that do not yet exist:

```sh
python3 scripts/edition2/export_snapshot.py SOURCE_COMMIT /tmp/ensemble-chNN-export --manifest /tmp/ensemble-chNN-manifest.json
```

This exports committed `solutions/edition-2/main/` and records its source and
file identities. It does not validate, replace a snapshot, stage files or tag.
After the chapter gates pass, the coordinator checks and promotes the export
and manifest into the chapter checkpoint.

Keep an initial student checkpoint before comparative review and a validated
checkpoint after corrections. Each validated chapter receives a dedicated
outer commit and immutable annotated tag such as `edition-2-ch03-r1`, binding
the manuscript, matching main source, chapter export and validation manifest.
A correction gets a new revision tag; never move an old tag or relabel its
receipts. An interrupted attempt can have an unfinished commit too, provided
its status says what remains. Commit only intended files; a shared workspace
may contain someone else's work. A local commit or tag does not publish it.

These are the reference workflow's durable records:

| Path | What it answers |
|---|---|
| `book/edition-2/chapter-NN.md` | What does the student have to learn and implement? |
| `book/edition-2/chapter-NN-outline.md` | What is the chapter's stake, teaching order, and intended scope? |
| `book/edition-2/chapter-NN-evidence.md` | Which sources, runs, findings, and limitations support the prose? |
| `book/edition-2/architecture.md` | Which ownership and dependency decisions are current? |
| `book/edition-2/skills/ensemble-coding/SKILL.md` | Which coding rules must a worker load? |
| `book/chapter-writing-procedure.md` and `book/edition-2/workflow.md` | Which work and review gates are required? |
| `book/edition-2/global-review.md` | Which later lessons and dependencies bear on this chapter? |
| `book/edition-2/progress.md` | What is complete, what is blocked, and what is the next action? |
| `solutions/edition-2/main/` | Where does the current student implement and integrate the new agent? |
| `solutions/edition-2/chNN/` | Which frozen source export and evidence belong to this chapter? |
| Outer commit, `edition-2-chNN-rN` tag and validation manifest | Which exact source, manuscript, export and checks define the validated revision? |

A checkpoint should let a fresh worker continue without reconstructing
the entire conversation. Record accepted decisions separately from tentative
ideas, the exact source revision and commands, unresolved questions, failed
checks, and the next executable action. Do not rely on a host preserving
every worker or keeping a whole book in context across restarts.

## 0.10 The crossing

The agent being built can eventually become the tool used to build its
next exercise. The first edition calls that **self-wielding**, by analogy
with a self-hosting compiler. It is a useful milestone when the actual
workflow demonstrates it, not a property granted by naming the loop.

The loop remains the same after the crossing: teach a capability, implement
it, test it, use it, compare and revise, then preserve the evidence. The
agent can help find a missing requirement or write a better explanation.
The resulting claim still needs a check or a receipt.

## 0.11 Your Ensemble

The right agent depends on how you work. Customize the tool set, the
memory system, the observer, and the interface. Keep the chapter contracts
and their checks together so an intentional change can be distinguished
from a regression. If a feature changes the architecture, update the
teaching where that decision first matters instead of accumulating a
repair chapter at the end.

The book was designed to be forked. Its useful inheritance is more than a
working snapshot: it includes the reasons for the design and a way to test
the next version. Start with Chapter 1 and give the first conversation a
place to live.
