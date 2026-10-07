# Chapter 3 validation

Date: 2026-10-07. Status: the reopened human-chat implementation, actual terminal
evidence and independent review now pass. The exact reviewed source export and manifest are bound by checkpoint
`edition-2-ch03-r1`. Bill's editorial approval is separate. The earlier tool implementation
was accepted at student checkpoint `7cbbd8e2e0101e40dee63fe18d05f48b791c55ff`.
Its original nested history was preserved during consolidation; development now
uses outer-tracked `solutions/edition-2/main/`, with `ch03/` a frozen export.

## Human interface correction

The integrated human client and all-three-provider actual PTY runs were frozen
in outer commit `a347ce31511c4b124e486bb41ef98c07bd17cec5`. Its executable SHA256
is `3d49d9037077929b7380fbbae020e968c79075ac94eb95c939535aa606523990`.
The fresh student extended the reviewed client and Chapter 3 tool implementation
using new-only source instructions. This is a guided integration, not a fresh
whole-chapter blind trial.

Independent black-box acceptance passes 45/45; seven checker controls pass.
A passing disposable control plus seven compiled defects each produce their
exact expected failure sets, with original production hashes unchanged.
The checker was prepared at `10ab549` before inspecting this implementation.
It retains Chapter 2's 42-case mode and proves absence with the older no-tools
binary failing exactly the six newly introduced Chapter 3 cases.

The student recorded clean formatting, vet and tests in all four modules and
inherited grade 100. No shared grader or legacy implementation changed for this
human-client integration. The earlier full-root regression below remains the
legacy gate; it is not relabeled as a new run.

Post-run comparison with the first-edition CLI found the small client integration
appropriate. Evidence-only revision `339a2a61e7107b921bdc7acbd704b2eb7da17931`
fixes the verifier's missing immutable source/executable binding before execution
or derived writes. Independent negative and positive controls pass. Production
and paid raw receipts remain those of the initial frozen executable.

The reviewer checked actual all-three-provider terminal transcripts, tool facts,
disk effects, history-selected redaction, reconstructed subsequent requests,
ephemera, accounting, commands and EOF. The author's updated spin matches those
receipts and preserves observed model mistakes. See `chapter-03-human-review.md`
for precise comparison, commands, scope and receipt paths. Reviewed source and imported independent receipts are committed at
`8494fdb0e5d6c096445bfac039458b83dd225332`. The 491-file frozen export is
`solutions/edition-2/ch03/`; `solutions/edition-2/manifests/ch03-r1.json`
binds its source tree, contract, skill and review evidence. All four modules
also passed vet/tests from the exported location, with clean formatting.
The snapshot exporter matched every committed blob and mode.

## Student attempt and revisions

The student started with a fresh context (`fork_turns="none"`) and an explicit
new-only handoff. Its source-read ledger records the new chapters, mandatory
skill, architecture, and preceding new implementation. Historical links in
the skill were excluded. The shared filesystem is an instruction boundary,
not a sandbox or a proof of model-training independence.

Initial implementation `590c4f4` and initial run checkpoint `540fb4a` contain
identical production code, derived from Chapter 2's `cc1bec45`. Preserve both.
Guided corrections and post-run comparison feedback follow that initial trial;
they must not be counted as unaided initial success.

The reviewed binary SHA256 is
`0cc92c69d1f6dd2316af2a0d3f59f5f336f72384ae2c2e65f7d009fe8ff81e9d`.
The student's `evidence/ch03/reviewed-source-hashes.json` binds its source.

## Checks and their scope

- Core, optional GUI, and both public-consumer modules pass formatting,
  `go vet ./...`, and `go test ./... -count=1`. The formatting output is empty.
- Inherited grader: 100/100. Initial fixture incompatibilities are explained
  below; neither their correction nor the final score proves the new contract
  in full.
- Independent acceptance: 39/39 scenario groups, thirteen on each fake API.
  The earlier 36-group checker produced 33 passes on the frozen initial binary.
  With Unicode cases added after the new teaching, that same initial binary
  produced 33/39. Its failures were the empty-file EOF spelling and split-rune
  caps on all three surfaces. Keep the checker versions and hashes distinct.
- Eight independent checker controls pass. A passing source control and five
  compiled mutations produce exactly the expected failures: read and command
  Unicode caps, continuation after an ordinary tool error, the request bound,
  and completion of the final accepted batch. The mutation run uses Messages;
  ordinary acceptance covers all three surfaces. This is a partial audit.
- Scoped package analysis passes. Independent source review separately checked
  actual owners, interface parent chains, logger access, and the public boundary.
  A package-check pass alone does not establish those properties.
- Focused student tests cover durable side-effect boundaries, accepted history
  after continuation failure, strict arguments, bounded output, and replay.
  Fault injection is deterministic local evidence, not a claimed live API
  failure.

CLI acceptance inspects requests, recorded facts, and actual disk effects.
It does not prove public Agent isolation or persistence-fault handling.
Its command markers and labels do not establish every label-to-stream
association; inherited grading, focused tests, and review are separate evidence.
Listing/search checks bound fixture content without prescribing presentation
headers. Keep these limits with the reported counts.

## Real use and improvements

All three actual CLI and public-consumer paths succeeded. Each initial CLI
session exercised all six tools, guard refusals and recovery, separate command
streams with exit 7, and multiple human turns. Independent file inspection
confirmed effects. Consumers exercised Agent workspace and capability
isolation, empty declarations, attributed ordered observations and unsubscribe.
The GUI remains an optional-module stub; no browser run is claimed.

The guided empty-file clarification received fresh live runs on all three
providers, including explicit EOF, explicit invalid start, silent exit 7, and
recall of a model-invented answer. Offline rendering of original live logs was
repeated without credentials and left logs and disk effects unchanged. The
feature ledger preserves exact commands, model identities, usage and source
states. Unchanged paid ASCII demonstrations were not repeated for scalar
workspace lookup or Unicode boundary fixes; deterministic Unicode controls
establish that correction.

The independent first-edition comparison accepted four revisions: an interface
capture parent, the clarified empty-file rule, an Agent-owned scalar workspace
query replacing whole-configuration copies, and complete UTF-8 prefixes at
byte caps. The initial architecture violation and new teaching gaps remain
visible. The historical answer shared the Unicode defect; correcting it is
an improvement over both initial answers. No performance benchmark is claimed.
See `chapter-03-code-review.md` for compared snapshots and concrete tradeoffs.

## Shared grader compatibility and legacy gate

The unchanged initial binary moved from 65 to 100 after two fixture repairs.
Gemini declaration inspection now accepts the documented JSON Schema field
alongside the older Schema form, rejecting conflicting fields and malformed
schemas. Fake live requests receive an explicit known resolved model identity
for signed continuations, while retaining the historical requested alias.
Offline fixtures clear inherited resolved identity rather than guessing it.

The legacy Chapter 3 reference passed before changes. Targeted Chapter 2/3
reference and mutation tests and new fixture controls passed afterward.
Independent review found no scoped coverage weakening. Root vet and changed-Go
formatting checks pass. First-edition solutions and `agent/` were not edited.

The first full-root run failed when the linker exhausted disk space. Its log
is retained as `checkpoint-evidence/legacy-root-after-ch03-disk-failure.txt`.
Clearing the disposable Go build cache recovered space. The complete rerun
passed with exit 0, including `internal/grade` in 508.374 seconds; its log is
`checkpoint-evidence/legacy-root-after-ch03-fixtures.txt`.

Independent receipts are in `checkpoint-evidence/ch03-independent-*.json`.
The student's `evidence/ch03/` retains live, local, source-read, initial-grade,
and revised-grade receipts. Independent manuscript proof is recorded in
`chapter-03-review.md`; the reviewer recomputed usage and inspected recorded
effects and render pairs. The final student checkpoint retains the reviewed
production unchanged. Bill then identified the missing human chat mode: the
old real-model JSON sessions establish the machine interface, not interactive
human usability. The later correction described above carries Chapter 2's new
chat contract forward and closes the implementation/live-review gate with actual
terminal evidence. These later receipts do not change what the earlier runs prove.
