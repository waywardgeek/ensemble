# Edition 2 carryover: what the student should know before starting

Read this alongside the original first-edition chapters and the mandatory
[coding skill](../.agents/skills/ensemble-coding/SKILL.md). Start with the
unchanged first-edition Chapter 1. No rewritten chapter is needed to carry
lessons forward, and the preserved second-edition Chapter 1 draft is not part
of this student's assignment.

**Coverage:** this initial document collects Bill's agreed architecture and
workflow corrections, with supporting passages from first-edition Chapters 5,
20 and 22. It is not yet the result of a complete first-edition reading. Before
the new student run starts, the coordinator must read the whole current edition,
complete this carryover and record the reviewed revision and coverage here.
The passages cited below were inspected at repository revision `01b4d7f`.

The original chapter still defines what to build. Carry forward a lesson when it
changes how to implement that exercise; do not implement a later chapter's
features early. Preserve the original graders. If a grader or chapter conflicts
with these agreed rules, record the conflict. Bill authorizes the reviewer to
accept an unreasonable exercise requirement when it independently agrees with
the coder's notes, record the exception and advance to the next chapter. Preserve
actual grader/live results separately from that decision. Ask Bill about
unresolved consequential conflicts; do not secretly weaken a check or reproduce
a known architectural flaw. The workflow describes this exception path.

## Apply from the first exercise

**Give facts owners and make those owners reachable.** Without a parent chain,
a parser that needs to log a failure ends up with a global logger, extra wiring
through unrelated callers, or no diagnostic. A child receives and stores an
interface to its actual parent; each parent interface exposes the next parent.
This lets new consumers reach existing configuration and logging without copies
or side channels. The root has no artificial parent. Source: Chapter 5,
[“The ambush” and “Kill the globals”](../book/chapter-05.md#59-the-ambush), and
Chapter 22, [“The chain, and why a flat root kills it”](../book/chapter-22.md#the-chain-and-why-a-flat-root-kills-it).
Bill clarified that the interfaces must represent the ownership chain, rather
than a bag of sibling dependencies.

**Put declarations together and behavior where it belongs.** Shared core values
and interfaces live in `internal/common`; implementation packages import common
instead of one another. Keep behavior in its responsible package, using free
functions where Go's receiver rule would otherwise push it into common. Private
runtime implementations behind common interfaces are allowed by Bill. Source:
Chapter 5, [§5.1](../book/chapter-05.md#51-the-idea-in-plain-words). Do not create
empty packages or generic services merely to anticipate later exercises.

**Use one real composition root.** The CLI enters through the public library
that constructs the ownership tree. A correct library constructor accomplishes
nothing if the executable bypasses it and wires a second, flat graph. Inspect
the actual constructor calls and paths to data, not just interface names. Source:
Chapter 22, [“The measurements that found it”](../book/chapter-22.md#the-measurements-that-found-it)
and [“The check that protected the vocabulary”](../book/chapter-22.md#the-check-that-protected-the-vocabulary).

**Use Bill's clarified ownership and client boundaries.** One Ensemble/Hub owns
potentially many Agents and the logger; Agent owns configuration, and Engine
owns usage. Tool visibility is per-Agent, whether registry storage is local or
root-owned. Agent streaming observations go through Observer; explicit parent
methods may provide services. GUI/WebSocket code belongs in a separate optional
Go module using public core interfaces. From Chapter 2, provide the CLI and the
small requested GUI stub over the same Agent; grow the GUI in its own chapter.
When MCP arrives, keep its transport replaceable, including a WebSocket tunnel.
These are Bill's instructions for this attempt, not claims that the old book
already specifies every boundary this way.

## Keep in mind; apply when the feature arrives

**A setting must change the operation, not only its readout.** Chapter 20
describes a model selector whose display changed while requests used the old
model, and vendor switches that failed to switch credentials. Follow the
setting to the request that consumes it; report the effective state. Source:
[§20.2](../book/chapter-20.md#202-the-bugs-that-succeed). This is guidance for
implementing settings and switching, not a requirement to add them in Chapter 1.

**Live, replay and different clients must describe the same Agent.** A saved
conversation is not sufficient if restart loses its visible content. Keep
identity consistent across live and replay paths, and keep authoritative turn
classification in the Agent rather than guessing from browser state. The CLI
and GUI should exercise the same runtime. Source:
[§20.3](../book/chapter-20.md#203-two-paths-one-screen). Bill separately clarified
that pause is Agent-wide: explicitly unpausing in any tab unpauses for everyone.
Do not introduce independent tab vetoes.

**Preserve the facts needed for later accounting.** When usage spans models,
keep counts attributable to the model that incurred them. Applying the current
model's price to all previous usage changes history's apparent cost. Source:
Chapter 22, [“Where the money went wrong”](../book/chapter-22.md#where-the-money-went-wrong).
Do not add pricing or model-switching features before their exercises.

## Lessons from the retired second-edition attempt

KISS is the yardstick. An agent-written contract can invent unnecessary work,
and tests can faithfully validate that invention. The coder implements the
original exercise, records difficulties for a future author, and may reject
provisional design suggestions for a simpler working approach. The reviewer
compares scope, active code, useful comments and meaningful tests against the
original solution. The coder does not read that solution.

Run the original grader and exercise the actual human interface with a real
model. Report failures and assistance honestly. Keep the author out of the loop
until the complete implementation succeeds and comparative reviews support its
improvements. Sources: [first-attempt history](edition-2-attempt-1.md), Bill's
restart instructions recorded in [AGENTS.md](../AGENTS.md), and the
[workflow and learning record](agentic-codebook-workflow.md).

This carryover explains the lessons; the skill supplies the operational coding
and review rules. A new discovery belongs here only with its source, rationale
and point of application. Separate established corrections from unresolved
questions. Neither this document nor its full-book preparation is permission
to rewrite the exercises, add acceptance machinery or copy an old implementation.
