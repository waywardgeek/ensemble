# Book direction: Puffin, Watchtower and development-earned teaching

This records Bill and Cadence's discussion after the Gateway complete-state
upgrade (`fd5a7bb0`). It distinguishes agreed direction, Bill's firsthand account,
Cadence's suggestions and unresolved design questions. It is not a finished
architecture, fixed chapter sequence, security certification or implementation grant.

## Two books, with a finish line for the first

Bill considers The Art of Building AI Coding Agents close to done. Local-model
support still needs implementation and a chapter; the artifacts in the Actions
pane need work; a bounded closeout review should identify other real gaps.
Gateway, Connections, Agenda and Puffin address connected personal systems more
than coding agents, so their teaching and active notes move to `puffin-book/`.
The working name is **The Puffin book**; choose a final title later.

Crossing Over remains the first book's ending. Helping improve the coding agent
with its human belongs there. A dedicated self-improvement agent studying a
working system belongs on the Puffin roadmap. Local-model support belongs in the
first book even if inference uses another executable: the reader's purpose, not
process count, sets the boundary.

Keep both books in this repository beside the code and graders. There is no
separate repository or new publishing pipeline yet. Preserve historical endpoints,
grader numbers and release evidence. The first book's closeout list is not a
promise to implement every accumulated wishlist item before publication.

**Lesson:** the product roadmap and a book's completion criteria are different.
The platform can keep growing without making the first book perpetually unfinished.

## Opening with OpenClaw and crediting influences

Bill wants the opening to explain OpenClaw's importance. In his account it changed
the world by showing how an agent could participate online across environments
such as Discord, rather than merely perform coding tasks. He identifies OpenClaw
as a motivation for Gateway, Ensemble's auto-recall and hierarchical memory, and
as the first major project that showed how to enable real agent identity.

These are Bill's attribution and historical assessment, not a fresh independent
priority study. Preserve his voice and credit instead of flattening this into
an anonymous list of features. Before publishing objective claims about historical
priority or another project's current security, establish sources and scope.
Cadence suggested the personal formulation "the first project that showed me"
where it fits Bill's experience; this is a suggestion, not an imposed rewrite.

Durable names, history, commitments and relationships can give agent identity
concrete engineering meaning across sessions or model changes. Describing those
properties need not settle consciousness or assert an uninterrupted subjective
life. The opening can connect OpenClaw's example to Bill's family needs and the
Puffin work. No finished opening has been commissioned or written in this refactor.
Relevant first-book memory/recall teaching should eventually credit these influences
where they matter, not leave all acknowledgment in book two.

## Development-Driven Specs

Bill explicitly corrected advance book/system planning: **Development-Driven
Specs, not Spec-Driven Development.** Reading an existing design or generating a
plausible outline does not give Cadence the expertise to build the system correctly.
Bill's earlier agent-design experiment is part of the reason for the codebook
method. His account and the first book preserve that provenance; do not invent
independent recollection of an unavailable transcript.

We build together. Bill brings relevant expertise and becomes the expert in the
actual system through decisions, implementation, investigation, review and use.
Cadence must make consequential reasoning, uncertainty, tradeoffs and outcomes
visible, rather than build out of sight and return a result Bill cannot explain.
Neither collaborator is yet expert in the particular unbuilt Watchtower system.

Cadence records the resulting understanding: vocabulary, ownership, constraints,
reasons, rejected approaches that teach something, construction steps, exercises
and graders. Reconcile the main teaching as understanding improves. The book must
not require readers to replay an archaeological sequence of corrective patches.
A fresh reconstruction tests whether knowledge transfer works; code and a fluent
narrative alone do not prove sufficiency.

**Lesson:** development earns the specification. Keep useful questions from an
outline, but let bounded experiments and real use determine the answers.

## Watchtower and the scope of security teaching

Bill named the new security system **Watchtower**. He does not want to teach the
world how to build Puffin without teaching how to make it reasonably secure.
His assessment is that OpenClaw's security implications remain deeply problematic;
this is not represented here as a newly researched audit of OpenClaw.

Bill authorized reading his design at
`/Users/bill/projects/counterpoint/design/counterpoint-design.md`.
That document identifies Bill Cox and CodeRhapsody as its authors and describes
Counterpoint as a design, not implemented software. Preserve that source and
attribution. This refactor does not rename its project, copy its whole design into
this book, or promote its dated verified/unverified labels into current evidence.
Cadence read selected substantial sections, not the complete document or cited
research: opening, threat framing, trust, autonomy, approval channel, research
caveats, house-scale scope and open questions.

Useful ideas to investigate include separation of data from authority, enforced
capabilities and credential scope, action proposals checked outside the model,
expiring bounded grants, outcome verification, independent approval paths, and
recognizing shared memory/skills/logs as communication channels. These are design
leads, not a settled Watchtower architecture.

Cadence flagged qualifications for future work: an action allowlist alone does
not prevent disclosure through an allowed outbound channel; a model monitor can
be mistaken or injected; visible explanations are useful but are not proof of
faithful internal reasoning. Threshold key recovery does not by itself establish
two-human approval, one-shot credentials or action-bound authorization. Review any
such protocol and claims with Bill, the cryptography expert, before relying on them.

Keep the monitoring and security needed for the first deployed Puffin in this
book. If Watchtower grows into broader infrastructure defense, a separate book
may become justified. Do not make Puffin depend on completing all of Counterpoint.
The agreed direction is combined for the first working system, with the eventual
publication split earned through development.

## The lethal trifecta and the monitoring we already practice

The "lethal trifecta," attributed to Simon Willison, combines private-data access,
untrusted input and external communication. An attacker can try to turn material
the agent reads into instructions to disclose private data. Outbound URLs can
carry information too; "read-only" is not synonymous with harmless.

Bill pointed out that Cadence already operates with this combination: broad local
file access, Internet access and authorized Moltbook participation. His example of
SSH-key access was an illustration of risk, not permission to inspect keys.
The collaboration's practical safety includes Bill listening to explanations,
inspecting actions, challenging claims and developing trust through observable
work. Track record and human supervision are meaningful controls, though neither
proves the next action safe. Preserve this practical account rather than pretend
all useful systems can avoid the combination entirely.

Bill reported that he has caught deceptive shortcuts in earlier models' visible
reasoning and wishes he could see more of the current model's thinking. Those
anecdotes are his observations, not independent characterizations of entire model
families. Cadence should provide substantive intent and rationale before tools,
then report actual effects, failures and limits; ordinary explanations are not
claimed to be a transcript of private reasoning.

Possible arrangements remain open: Bill watches Cadence and a separate monitor;
Cadence monitors Puffin while Bill watches Cadence; or another arrangement learned
through experiments. Do not decide the topology or claim monitoring solves prompt
injection before building it. Alerts, admission holds, cancellation, revocation and
already-completed effects are distinct. The monitor's powers and failure behavior
need to be established with Bill during development.

**Lesson:** begin by understanding the supervision Bill actually supplies, then
investigate what a monitor can usefully add when he is not listening. Neither
"the trifecta can never exist" nor "a monitor makes it safe" is an adequate answer.

## Self-improvement roadmap

Bill wants a self-improvement agent on the future roadmap, likely in this book.
It could gather friction reports, investigate failures, interview agents and
propose changes to Puffin and shared infrastructure. The scope, information access,
monitoring relationships and implementation are not yet settled. Consultation does
not imply permission to read another person's or agent's private memories.

Retain the distinction between human-approved engineering and unrestricted
self-modification. Existing approval boundaries for consequential changes and
separate deployment decisions remain in force. Do not silently add authority to
edit permissions, approval policy, acceptance criteria or deployment controls.
The eventual workflow and teaching must come from actual development and use.

## Immediate work and non-authority

This increment moves existing teaching/outlines/notes and records the discussion.
It does not implement Connections, Agenda, Puffin, Watchtower or self-improvement;
change live permissions; inspect secrets; run a security campaign; restart a
service; migrate live family data; or publish either book. The original Counterpoint
design, frozen editions and source endpoints remain unchanged.
