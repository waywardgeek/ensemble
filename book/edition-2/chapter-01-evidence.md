# Chapter 1 second-edition evidence

Status: student contract and body drafted; implementation and validation
pending. No second-edition pass claimed by the author.

## Read before writing

`book/chapter-writing-procedure.md`, `book/voice.md` (v4 in full),
`book/course-policy.md`, `book/chapter-01.md`, `book/derive-ch01.md`,
`book/chapter-01-parking.md`, `book/brief-ch01-grader-audit.md`,
`book/review-ch01-grader-audit.md`, `book/brief-ch01-model-refusal.md`,
chapter 5 and chapter 22's architectural contract, current chapter-1
solution, fake, grader checks, harness, and mutation cases.

Untracked historical notes were read, not edited. Derivation reports
are evidence of earlier states, not authority over today's source.

## Git receipts

- `a3642d5`: initial chapter-1 grader, fake, and reference.
- `835946f`: deletion audit discovered first-block parsing and hardcoded
  API key both still scored 100. Fake now splits replies and checks key
  equality. It added mutants for those properties.
- `898f3b3`: exact model equality and nonempty system prompt enforced;
  `hardmodel` and `nosystem` added. Earlier report's open items are closed.
- `7eb3e24`: reference refuses an unset model rather than choosing an
  aging default. The original grader always supplies it, so absence
  remained untested.
- `33a4875`: Bill's ruling distinguishes exercise alternation from real
  Messages API acceptance and confirms the model-ID mistake was the
  coding agent's, not a new author anecdote.
- `1bf97c0`: StackAgent chronology and raw-HTTP motivation revised.
- `ec41c6e`: committed workspace baseline for this rewrite.

## Current-source findings

- Seven checks retain weights 15/15/10/10/25/15/10.
- `growth` compares decoded role and flattened text, not serialized
  JSON bytes. It only notes an increment other than two; new contract
  requires the exact two-message increment.
- There are 14 mutation table rows: one clean control and 13 defects.
  Older reports describing ten or twelve mutants are historical.
- Existing `Ask` appends the user before HTTP succeeds. Existing chat
  mode continues after failure, leaving an unmatched user message.
  Draft requires termination on request errors in both front ends.
- Missing key/model, malformed input, and HTTP/decode failure behavior
  are not covered by the original happy-path harness.
- Filtering text blocks stays good wire parsing practice. Do not invent
  a non-text block carrying `text` merely to make the filter measurable.

## Architecture ruling supersedes the first draft

Bill explicitly directed that the coding methodology move into chapter 1
before any code. The earlier author/global-review recommendation to keep
the sacrificial flat chapter and introduce package architecture later is
superseded. So is any assumption that corrected data structures must wait
for their old chapter numbers. The subsequent chapter-boundary decision
retains the small text-message shape in chapter 1 and introduces neutral
events in chapter 2; package architecture is never sacrificed.

Read the current `book/edition-2/architecture.md` and reread chapter 5's
star/free-function explanation and `book/voice.md` for this edit. The
chapter now teaches ownership, star imports, free functions, immediate
parent interfaces, logging reachability, runtime Ensemble, Observer events
versus action requests, and separate optional GUI modules before its first
code block. No implementation accompanied this edit.

Settled: shared core structures/interfaces in `internal/common`; behavior
in responsible packages such as `internal/llm`; no behavior migrating into
common just to gain method syntax. Standard-library interface methods are
the narrow exception. Code likely to need debug logging needs a route to
the logger even when stateless. Runtime Ensemble owns potentially many
Agents; it is distinct from common, the shared package in the import star.

Settled: Observer delivers streaming events and real-time GUI observations.
It does not forbid explicit parent-interface methods for sub-agent
creation, messaging, or shared-service access. An earlier observer-only
interpretation is superseded. GUI/WebSocket implementation belongs in an
optional separate Go module, with public reusable component interfaces.

At this clarification stage, open choices included parent methods,
logger owner, registry storage, public GUI seam, and coordinator name.
Per-agent tool visibility is settled; insisting that registry storage
must be per-agent would contradict Bill's allowance for a shared registry
with per-agent visibility.

Bill subsequently specified that clean data structures begin in chapter
2, with Ensemble serving two clients: the CLI and a browser GUI via
WebSocket. The GUI may be stubbed in chapter 2. This settles the two-client
requirement without authorizing a second composition root or GUI transport
dependencies in the agent module.

Bill also requires every "Taking it for a spin" section to follow a coder
run exercising all chapter features live with a real model, initially
through the CLI. Fake-grader success cannot stand in for that run. API
configuration is available in `~/.cr/settings.json`; credentials must not
be printed, copied into evidence, or committed. This author has not read
that configuration or made live API calls during the prose revision.

Chapter 5 itself contains contradictory historical text: the public-API
paragraph describes global registration; §5.10 replaces the global map
with an Agent-owned registry. Its earlier dependency examples also hand
collaborators directly to constructors. The second edition follows the
current ownership-chain ruling, not those stale construction examples.

P9/P11 still require fixtures that exercise the promise and deletions
that demonstrate its failure, rather than identifier searches. The
existing first-edition grader does not establish the new architecture.

## Renewed autonomous authorization and completed contract

Root instructed the author to complete a student contract under renewed
autonomous authorization. Routine choices are resolved: Ensemble is the
working name and owns logger; Agent owns configuration/history; Engine
owns transport/usage. These coordinator decisions match Bill's guidance;
they are not fabricated quotations or attributed explicit rulings.

Earlier held failure/configuration proposals are now requirements, along
with empty input, present/nonnegative usage, nonempty assembled answer,
and committing valid pairs. The inherited seven checks remain distinct
from additional pass/fail acceptance. Root implements checks independently;
no new pass is claimed here.

The handoff comprises architecture §§1.1–1.4, TL;DR, and acceptance §1.8,
or the entire chapter. Constructor names/signatures remain student choices
under the ownership contract. The request/response body teaches the same
requirements in detail.

Bill endorses `solutions/edition-2/chNN`, each a self-contained Git repo
whose history derives from the preceding new snapshot. The public library
lives at module root, the executable in `cmd/`. Root verified the existing
grader discovers `cmd/` when given the root, so commands now use
`DIR=solutions/edition-2/ch01`.

Bill's final numbering target is 21 chapters, 0–20: chapter 0 unchanged,
old 1–4 unchanged in number, old 6–21 shifted to new 5–20. Old 5 and old
22 become absorbed lessons, not new standalone chapters. Historical
chapter-5/22 references in this evidence file identify first-edition
sources. Chapter 1's reader-facing text makes no repair-chapter promise.

## External verification

Read official current API references on 2026-10-07:

- [Create a Message](https://platform.claude.com/docs/en/api/messages/create)
  supports stateless multi-turn conversation, request fields, and typed
  response content. It is not evidence that the course's strict
  alternation rules apply to every valid API call.
- [List Models](https://platform.claude.com/docs/en/api/models/list)
  supports `GET /v1/models` discovery. No current model ID is printed as
  a recommendation in the draft.
- [Go method declarations](https://go.dev/ref/spec#Method_declarations)
  confirms the receiver base type must be declared in the method's
  package. The chapter teaches free functions within that restriction;
  it does not claim cross-package methods are legal.

The first-edition private StackAgent story is retained as an attributed
record, not presented as newly externally verified reporting. Omitted
market valuations, pricing tiers, unavailable proxy promises, and live
token totals do not need fresh claims to teach this chapter.

## Prose validation

Author reloaded the full `book/voice.md` and rewritten
`book/chapter-writing-procedure.md` after the workflow change. Reviewer
findings resolved: model discovery now reads its credential inside a
Python helper rather than expanding it into curl's argv; finite timeout
and configurable logger writer are explicit in the TL;DR; live independent
Agents use a separate executable public consumer, not an internal test.

Ran `go run ./cmd/lintprose book/edition-2/chapter-01.md`: exit 0, no hard
failures, including after completing the contract and explanatory body.
Latest soft warnings: 3,560 prose words (below the suggested 4,000) and a
long mechanism passage without a named person. These are not gates; no
padding or invented story was added to satisfy them. No code/grader tests
were run by the author, and no implementation was changed.

## Validation pending

- Student implementation from the complete clean chapter contract.
- Existing seven-check grader and independent second-edition probes.
- Property deletion audit and final full-chapter prose review.
- Actual live-model feature exercise and a reproducible, secret-free
  transcript for "Taking it for a spin"; initially through the CLI.
