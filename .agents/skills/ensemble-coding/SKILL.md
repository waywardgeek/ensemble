---
name: ensemble-coding
description: Current Edition 3 C# architecture, ownership, KISS, teaching and verification rules for Ensemble engineering.
---

# Ensemble — current Edition 3 work

## Canonical layout

- `agent/`: active C# 14 / .NET 10 application.
- `book/`: current Edition 3 manuscript, Chapters 00–26, and companion.
- `solutions/chNN/`: preserved C# chapter endpoints. Their historical numbers
  differ from some current manuscript numbers; see `book/README.md`.
- `edition-1/` and `edition-2/`: frozen archives. Do not rewrite their sources,
  books, manifests or evidence to accommodate current changes.
- Root `cmd/` and `internal/`: Go grading/support tools, not the runtime.

Read the top of `docs/edition-3-csharp-progress.md` for the current boundary and
`docs/edition-3-csharp-procedure.md` for the implementation method. Historical
phase instructions and source paths are records, not authority to resume work.
The promotion and known verification limits are in `docs/edition-3-promotion.md`.

## Working with Bill

Explain what you will do and why before tools. Work personally unless Bill
explicitly authorizes delegation. Inspect actual declarations and errors; do not
claim a tool invocation proves its effect. Use small checked edits, focused
behavior tests, and broader regressions for shared changes. Preserve unrelated
work, private credentials, and running applications. No automatic installations,
provider campaigns, host restarts, pushes or publication. Scoped local commits
are permitted for authorized work; Bill owns pushes.

Normal work is interactive collaboration, not an autonomous chapter campaign.
New student exercises require an explicitly assigned isolated worktree and scope.
A candidate is not permission to replace Bill's installed application.

## C# architecture and review

Common owns shared vocabulary, persistent declarations and real relationships.
Functional directories implement matching partial members in the same assembly.
Use strong parents and explicit ownership rather than copied service bags or
reconstructed identities. Constructors establish both sides; refusal changes
neither. Destruction joins owned work, removes reverse memberships, recursively
destroys strong children and detaches weak links without cascading.

Agent owns scheduling and History; Engine consumes frozen requests and returns
immutable results. Record calls before effects and results before further model
requests. Observation cancellation is not work cancellation. Hints cap ordinary
observations at three seconds from observation start and preserve finalized batch
calls; explicit Interrupt remains distinct. Skills grant authority; historical
text and discovery do not. Passing tests do not waive architecture or review.

Keep teaching code and prose readable, with reasons for nontrivial decisions.
Update the executable book and relevant graders when changing its contract.
Do not rewrite historical endpoints or imply ungraded obligations were certified.

## Build and verification

Use the root README/Makefile. `make DOTNET=/path/to/dotnet` builds C# and writes
the local root launcher; it does not launch or install it. `make cli` builds the
JSON CLI. `make test` runs current C# and shared V3 grading checks.
`make test-legacy-go` retains broader historical tests with known obsolete layout
and pricing expectations; it is not V3 runtime acceptance. Use isolated homes,
ports, fake providers and explicit SDKs. Do not download dependencies silently.
Source, build, installation, host restart and browser refresh are separate facts.
