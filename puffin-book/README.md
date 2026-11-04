# The Puffin book

Working name agreed with Bill. A final title and chapter numbering are deliberately
undecided. This book grows beside [The Art of Building AI Coding Agents](../book/README.md)
in the same repository, so implementation, graders and teaching can evolve together.

## How we work

**Development-Driven Specs, not Spec-Driven Development.** Bill brings expertise
and judgment; Bill and Cadence build, investigate, test and use the system together.
Bill becomes the expert in the actual system. Cadence records the decisions,
ownership, reasons, failures and verification in an agentic codebook that transfers
that expertise. Reconstruction tests whether the teaching is sufficient; a polished
outline does not establish that it is.

The material below is an inventory, not a fixed table of contents. We do not yet
know the finished architecture of Puffin or Watchtower. Existing outlines preserve
useful thinking and open questions, not authority to implement a whole platform.

## Material so far

- [Gateway](gateway.md): implemented teaching moved from the first book, including
  complete-state persistence and its scoped external grader. Source candidate,
  not permission for live installation or migration.
- [Connections](connections.md): provisional subject outline.
- [Agenda](agenda.md): provisional subject outline.
- [Puffin](puffin.md): provisional integration outline.
- [Book direction and discussion decisions](notes/book-direction.md): OpenClaw,
  identity, Watchtower, monitoring, self-improvement and the first-book boundary.
- [Detailed existing outline](notes/puffin-chapter-outlines.md).
- [Scheduling and Puffin decisions and investigation](notes/scheduling-and-puffin.md).
- [Read-only xyzzy inventory](notes/puffin-xyzzy-inventory.md): dated evidence,
  not a current machine-health report or deployment authorization.
- [Gateway complete-state evidence](notes/gateway-complete-state.md).

## Boundary with the coding-agent book

The first book retains the coding collaborator, local-model work still to be
implemented/taught, Actions-pane/artifact improvements, and Crossing Over as its
ending. Gateway through Puffin belongs here because it concerns connected personal
AI systems, not because every component must be a separate executable. A local
model server can be another executable and still belong in the first book.

Self-improvement is a future subject here. The security system is named
**Watchtower**. Include the security and monitoring needed for the Puffin we
actually deploy; whether a broader Watchtower system earns a separate book is a
later decision grounded in development. Counterpoint remains an attributed design
source, not this book's specification.

## Source, evidence and publication

Active implementation remains under `agent/`, including `agent/Gateway/`.
Graders remain in their current locations. Historical endpoints, grader flags,
archived editions and scores are unchanged. Gateway's former chapter number does
not assign a number in this book or change the historical multimedia grader.

No second publishing pipeline or release artifacts have been created. The first
book's previously assembled Markdown/PDF/EPUB files remain prior-release outputs.
Moving prose is not deployment, proof of security, or completion of the planned
Connections, Agenda, Puffin or Watchtower work.
