# Second edition: Chapter 1 working outline

Status: student contract, body, and live demonstration validated by the
coordinator at revised checkpoint `75542c1`. Initial live receipts describe
`459e4ce`; code comparison, diagnostic revision, acceptance/audit, and
independent prose/receipt review are complete. Editorial approval is separate.
First-edition source is preserved. Working baseline: `ec41c6e`.

## Edition map

Bill's updated working map is Chapters 0–21: the previous plan through
Chapter 20 plus a new final comparison of the editions. Chapter 0 now has
the explicitly requested second-edition execution guide. Old chapters 1–4
keep their numbers; old 6–21 become new 5–20. Old chapters 5 and 22 disappear
as standalone repair chapters; their lessons enter where first relevant.
The final comparison is a separate purpose, described in
`ending-chapter-note.md`. Bill permits organization changes with a concrete
teaching benefit. Coder drift is corrected in code; it does not authorize
weakening these rules. Preserve first-edition artifacts unchanged.

## Voice plan

- Stake: the reader must be able to add a feature without duplicating
  state or reconstructing access to data the program already owns.
- Register: compiler-book mechanism, with one short sourced origin story.
  Open with the practical cost of a second conversation and unreachable
  diagnostics; show the missing connection before prescribing parent access.
- Bill moment: StackAgent bet and demonstration, recorded in chapter 1.
- Confession inventory: the first-block grader hole, documented in
  `835946f`; no invented story or freshly claimed live run.

## Story-preservation pass, voice v5

Source: first-edition `book/chapter-01.md` §1.0, reread October 7. Restore
the two-week proof-of-concept bet, request for a demonstration, working
StackAgent and deliberate restart as a short third-person sequence. The
earlier rewrite compressed this to biography and lost the choice to discard
code while retaining hard-won knowledge. The new opening connects that choice
to teaching ownership before construction. It attributes the outcome to the
recorded first-edition account, not a new benchmark of superiority to Windsurf.

Omit acquisition prices, exact demonstration/rewrite dates and the large
market-value thesis: they add external-verification work without explaining
this exercise. Omit invented dialogue or feelings. Trim the repeated generic
parent/logger motivation in §1.1 while retaining every rule, including
stateless-helper reachability. Independent story review is accepted;
the validated solution and its recorded runtime evidence remain unchanged.

Separate architectural clarification, October 7: Bill explicitly permits
private runtime structs when common interfaces expose their ownership chain.
§1.2 now distinguishes those implementations from shared data and interfaces;
this is not a request to relocate Engine/Registry/Jobs/Job or to weaken the
free-function rule for behavior on common values.

## Thesis

A conversation starts with explicit ownership, a star of package
dependencies, and parent interfaces that make owned data and logging
reachable. Teach those rules before the first code block and apply them
from the first implementation. Do not recreate architectural flaws for
later repair chapters to fix.

## Structure

1. Short motivation preserving the documented StackAgent origin.
2. 1.1 The idea in plain words: one owner for each fact, reachable through
   parent interfaces; debug logging applies to stateless helpers too.
3. 1.2 Packages follow responsibilities: `internal/common` declarations,
   `internal/llm` behavior, free functions rather than behavior migrating
   into common; standard-library interface methods are the narrow exception.
4. 1.3 Follow the owner: immediate-parent back-pointers, no globals,
   dependency bags, or closure bridges; distinguish shared package hub
   from runtime Ensemble; one owner with potentially many Agents.
5. 1.4 Events, requests, and the optional GUI: Observer for streaming
   events; parent-interface methods for actions/services; separate optional
   GUI module and public reusable GUI components. Starting in chapter 2,
   the clean data structures include Ensemble with CLI and browser-GUI
   client interfaces; the GUI can be stubbed while CLI use is live.
6. TL;DR: mandatory visible skill-loading instruction, then architecture,
   protocol, wire, history, and accounting contract. Instructions, checks,
   and review enforce the rules; file loading is no mechanical guarantee.
7. 1.5 The request carries the conversation: history and model discovery.
8. 1.6 A complete exchange has two messages: valid pairs, text blocks,
   and failure termination without fabricated output or retries.
9. 1.7 Read the token counts: owned accounting and independent
   Agents; fake tokens versus live usage and prices.
10. 1.8 Exercise, graded: inherited checks plus added configuration,
    failure, ownership, library, import, and parent-path acceptance.
11. 1.9 Taking it for a spin: exact October 7 CLI and external-consumer
    receipts, independently verified by the reviewer. Recorded CLI totals
    261 input/209 output; consumer proves independent Agents and captures
    a separately labeled local transport fault at Ensemble's logger.

## Grading decisions

The existing first-edition checks retain their current seven IDs and
weights. Preserve already-shipped exact model/key equality and nonempty
system validation when defining the new contract. The historical audit's
open items are resolved by `898f3b3`, not outstanding work. Passing those
checks alone cannot demonstrate the second-edition architecture.

Renewed autonomous authorization permits routine contract decisions:
negative configuration runs, failure responses with no retry, malformed
input, exact growth by two, nonempty answers and present nonnegative
usage. These are now chapter requirements alongside the ownership/import
rules. Root independently runs acceptance and deletion audits; final
retained reports determine validation. Post-run comparison additionally
improved the teaching of safe timeout/cancellation diagnostics and causes.

The earlier plan to defer parent chains and package topology until chapter
2 is superseded by Bill's instruction to teach the methodology in chapter
1 before any code. The first-edition sacrificial architecture is not the
second-edition construction plan.

## Implementable choices

Bill's later consolidation ruling supersedes separate student repositories:
develop in `solutions/edition-2/main/` under outer Git history, then export
exact validated versions to ordinary tracked `solutions/edition-2/chNN/`.
Each validated chapter gets a dedicated outer commit and immutable annotated
revision tag. Preserve the earlier repositories and their receipt identities.
The module root is the public library and `cmd/` holds the CLI; the
existing grader discovers that CLI when given the module root.

Coordinator choices for chapter 1: Ensemble owns logger; Agent owns
configuration/history; Engine owns transport/usage. These match Bill's
guidance but are not falsely attributed explicit rulings. Signatures and
names are student choices within the ownership contract.

Chapter 1 keeps the text-message slice; chapter 2 introduces neutral
events and two client surfaces. Package/ownership architecture survives.
Registry storage and public GUI signatures do not block a chapter that
implements neither feature.
