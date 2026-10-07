# Chapter 2 validation

Date: 2026-10-07. Implementation, independent comparison/revision, and live
demonstrations are validated. Bill's editorial approval is a separate status.
Initial student checkpoint: `39a92ca27a418712832ac0dcbbfbe4e32b3bca35`.
Reviewed checkpoint: `cc1bec45c3327c87728a4040f762155d8e860a0b`, clean, in
`solutions/edition-2/ch02/`. Both histories are preserved.

## Evidence and scope

- Core, external public consumer, and optional GUI modules pass formatting,
  `go vet ./...`, and `go test ./... -count=1`.
- Inherited Chapter 2 grader: 100/100 after the fixture correction below.
- Independent offline CLI acceptance: 44/44. The deletion audit passes a
  control and ten deliberately broken variants, each with the exact expected
  failing checks. Source and binary hashes bind these reports to the tested
  artifacts. Nine independent checker tests cover valid alternate JSON forms
  and malformed projections, without relying on the student's serializer.
- Scoped package analysis passes. The separate architecture review inspected
  actual parent/logger paths and public copy boundaries; the package checker
  alone does not prove those semantics.
- All three real CLI paths and all three real public-consumer paths succeeded.
  The live ledger records selected/reported model identity, inputs, outputs,
  usage, exact source states, controlled result ingestion, redaction, separate
  Agents, observer delivery/unsubscription, and model changes. The GUI remains
  a fake-backed optional-module stub; no browser/WebSocket run is claimed.
- Code comparison and revisions are accepted. The new implementation gains
  explicit owners, durable append-before-observe behavior, safer strict parsing,
  producing-model accounting, and exact signature provenance. Review removed
  unnecessary internal history copies, added safe diagnostic reasons, and
  explained subtle ordering/copy boundaries. No speed benchmark is claimed.
- Prose was independently reviewed against the feature ledger; the coordinator
  also checked the actual live receipts. Scoped prose lint has no hard failure.
  Its person-gap and negation warnings remain editorial considerations.

Reports: `chapter-02-architecture-review.md`, `chapter-02-code-review.md`,
`chapter-02-evidence.md`, and `checkpoint-evidence/ch02-*.json` / `.txt`.
The student retains its own full `evidence/ch02/` ledger and receipts.

## The inherited fixture correction

The initial final-contract run scored 95/100 because the old roundtrip phase
attempted to render a dumped history with unanswered calls. The new contract
permits loading that history but correctly refuses a request until results
exist. The old reference still passed before the harness edit; its baseline
is recorded in `checkpoint-evidence/ch02-initial-check-manifest.json`.

The corrected harness preserves every original dump byte and appends explicitly
labeled supplied fixture results only for outstanding call IDs. It executes
nothing and leaves already completed histories unchanged. No assertion or score
was removed. New controls cover exact-prefix preservation, correct result ID and
sequence, idempotence, and refusal to repair a malformed dump. Independent review
found no coverage weakening within this fixture's purpose.

The legacy Chapter 2 reference and mutation suite passed after the change.
The complete root `go test ./... -count=1 -timeout=45m` then passed with exit 0,
including `internal/grade` in 513.863 seconds; root vet and changed-file gofmt
checks also passed. Receipt: `checkpoint-evidence/legacy-root-after-ch02-fixture.txt`.
First-edition solutions and existing `agent/` implementation were not edited.

## Limits that remain visible

Live receipts precede the final diagnostics/internal-copy revision. Those
changes passed relevant local gates and review; unchanged paid conversations
were not repeated. The earlier Gemini declaration correction did receive a
fresh successful live run, and its original failure is retained. The observed
OpenAI model restriction is a dated surface/configuration result, not a blanket
claim about that model's capabilities.

The audits are partial, not proof of every possible event or architecture defect.
Offline ephemera evidence reconstructs requests from actual live log prefixes;
it is not intercepted HTTP. Sanitized provider signatures are replay material,
not API credentials. Credential scans were performed without printing keys.

The coder reports no direct reads of first-edition chapters, answers, existing
agent code, or grader source. It did inherit historical summaries, and compacted
history prevents a complete raw-trace audit. Chapters 1–2 are guided student
builds with no known answer-key exposure, not certified strictly blind trials.
See the student's `evidence/ch02/SOURCE-EXPOSURE.md`. Subsequent cold chapters use
fresh contexts without coordinator-history inheritance.
