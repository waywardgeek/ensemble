# Chapter 11 independent checker preparation

Status: runnable foundation and partial CLI runtime command, **no Chapter 11
runtime acceptance**. Chapter 10 is accepted at `edition-2-ch10-r1`;
Chapter 11 implementation is pending. The original foundation report below
retains its earlier preparation context. Neither fixture runs nor the new
command substitute for the chapter's complete acceptance and live demonstrations.

## Role and exposure

The coordinator reassigned the former Chapter 8 student / initial Chapter 9
checker engineer to independent Chapter 11 checker preparation. This engineer
will not be the Chapter 11 student. Chapter 9 checker ownership was already
handed to the independent reviewer. The Chapter 8 post-build teaching confirmation
was completed separately at `446d7f2`.

For this assignment I read the entire repository coding skill (and reloaded it
after compaction), architecture, Chapter 11 contract, Chapter 11 review and
validation documents, workflow, and chapter-writing procedure. Relevant new
predecessor reads were Chapter 10 §§10.1–10.3 and §§10.7–10.8: session identity,
lossless canonical numbers and persistence limits. The basis is the contract
accepted at `0c82feb`, drafted at `a5fc9ea` and clarified through `39bc526`.
No first-edition implementation, historical answer, Chapter 12 teaching, or
current student runtime source was read for this preparation. Earlier Chapter 8
implementation exposure remains part of this engineer's history. No implementation
API spelling has been inferred from that exposure.

## Runnable commands

From the repository root, using Python 3.9 or newer:

```sh
python3 scripts/edition2/accept_ch11.py --self-test
python3 scripts/edition2/audit_ch11_foundation.py
python3 scripts/edition2/accept_ch11.py --emit /tmp/ch11-new-fixtures
python3 scripts/edition2/accept_ch11.py --peer
```

The export destination must not exist. It receives four request/notification
files and three response files, all complete JSON messages without a delimiter.
The peer reads LF-delimited requests from stdin and writes LF-delimited replies.
It performs no notebook writes: its `notes.append` result echoes text and counts
characters solely as a deterministic protocol fixture. The fixed protocol version
and metadata implement the taught profile, not a claim about all MCP revisions.
`--receipt PATH` retains structured results from either non-peer command; the
audit also accepts it. Every receipt expressly states `runtime_acceptance: false`.

The command intentionally takes no student executable or guessed constructor.
The future public transport consumer must call the student's documented public
seam. Passing this foundation alone says nothing about that client's behavior.

## Prepared coverage and limits

| Contract area | Exercised preparation | Runtime evidence still required |
|---|---|---|
| Complete message / stdio framing | Real subprocess fixture peer agrees byte-for-byte with direct complete-message fixture replies; LF exclusion, escaped newline, empty line, partial EOF and clean EOF controls | Ensemble stdio and memory adapters, copying/byte ownership, blocked send/receive and joined close, public third-party adapter |
| Envelopes and bounds | Duplicate/trailing/UTF-8/surrogate and unknown-member refusals; result/error shape; input-required classification; exact/+1 message 8 MiB, nesting 64, nodes 100,000; signed int32 mathematical error codes | Actual connection fault/publication effects, safe error projection, incoming/outgoing allocation bounds, full notification grammar |
| Correlation | Out-of-order pending identities, issued stale identity, uint64 maximum and precise invalid/future/noncanonical ID controls | One reader, pending-map races, issued watermark mutation, retired generation callbacks, ID exhaustion and no wrap |
| Discovery | Candidate unpublished before final page; duplicate/repeated/empty cursor refusal; exact/+1 64 pages, 1,024 definitions, 4 KiB cursor, 64 KiB description, 256 KiB canonical schema | Actual transactional ready publication, failed preparation cleanup, 16 MiB accumulated bound, schema compilation before publish |
| Frozen selected definitions | Semantic numeric equivalence, adjacent unsafe-number distinction, description/input/output changes, tolerance for unselected extras | Public install/binding ownership, missing selected tool, alias/connection/remote-name identity, reconnect retry, durable v2 session identity |
| Request and cancellation literals | Discover/list/call metadata, exact safe cancellation notification, missing metadata refusal | No automatic downgrade, pending cancellation race winner, per-operation delivery, no unauthorized RPC, real lifecycle/permit retention |

These are independent oracle vectors and a fixture peer, **not positive tests of
student implementation**. Direct `Peer.accept` is not an Ensemble in-memory
adapter. Correlation classification does not prove synchronization; recognizing
a reverse request does not prove that a client refrains from executing it. The
descriptor helper normalizes the taught retained fields but is not a JSON Schema
compiler. The bounded canonical-number corpus uses Python Decimal; its passing
large-exponent example does not certify every legal JSON number lexeme or the
entire Chapter 10 canonicalizer. Python integer fixture values are normalized
through the same coefficient/exponent rule as parsed wire numbers.

## Actual controls and receipts

- `checkpoint-evidence/ch11-foundation-controls-initial.json`: original 64/64
  local preparation run, retained with its original checker hash. It predates
  the integer-fixture normalization correction and five additional controls;
  do not relabel it as the final checker version.
- `checkpoint-evidence/ch11-foundation-controls-final.json`: 69/69 final controls,
  binding the checker and Chapter 11 contract bytes. The checker verifies these
  hashes stayed unchanged before writing its result.
- `checkpoint-evidence/ch11-foundation-oracle-deletions.json`: the 69-check
  positive baseline, then nine separate in-memory predicate deletions. Each
  produces its specific expected `negative-accepted` failure. These audit the
  oracle's own safeguards, not the eventual student implementation. The
  subprocess framing positive still uses the unchanged fixture peer.
- `checkpoint-evidence/ch11-foundation-export-control.json`: seven exported
  files verified by hash, then actual subprocess output compared with the three
  exported responses. Both scripts passed Python AST parsing. The disposable
  fixture directory was removed; file identities remain in the receipt.

No initial failure was encountered. Inspection found that a native integer
fixture could canonicalize differently from an equivalent parsed Decimal;
the helper was corrected and the exact equivalence control was added before
publication. Initial receipt bytes remain unchanged. No Go module was affected,
so no Go build/vet/test run is claimed. No credentials, provider calls, historical
grader edits, main-source edits, or student mutation audit occurred.

## Continuation for the implementing reviewer

1. Start from accepted Chapter 10. Obtain the student's public constructor,
   message transport and configuration signatures after ownership-plan review;
   compile an external consumer without internal imports. Use this peer and the
   literal fixtures against both actual adapters, then a separately implemented
   public transport. Bind positive runs to full source and binary identities.
2. Exercise 64 permits across waiting, staged delivery and cancellation;
   immediate excess refusal; one-second delivery stall, absolute deadlines,
   killed-waiter/sibling isolation, all reply/cancel race winners, generation
   fencing, root-close joins and stopped-reading peers. Demonstrate owned process
   group termination and bounded stderr without treating peer tests as clients.
3. Cover the complete schema profile, local refs/cycles, numeric precision,
   physical/expanded/validation-step budgets, 16 MiB discovery accumulation,
   result/artifact byte identity and bounds. Invalid mixed results must fail as
   a whole. Ensure denied grants, pre-dispatch durable/artifact failures and
   invalid arguments produce zero remote calls.
4. Cover public two-Agent authority, frozen selected definitions, explicit
   compatible replacement, no resend, v1/v2 persistence identity and resume
   validation before prepare, and replay without external effects. Cover CLI
   creation-only config/absolute executable/environment restrictions and safe
   status projection. Actual GUI tunnel/control belongs to Chapter 12; do not
   invent it as a Chapter 11 implementation requirement.
5. Perform structural parent/logger/import review, exact generation/ID boundary
   controls, retained chapter gates and student-behavior deletion audits. The
   eventual student still owes its separately approved real-provider human CLI
   demonstrations. This preparation authorizes no paid run and claims no live
   coverage.

No unresolved contract ambiguity was needed to encode this subset. Unimplemented
checker areas above remain gaps, rather than unpublished assumptions or waived
promises.

## Independent preparation review and peer repair

After `ee89051`, the independent reviewer reproduced the 69 controls and nine
oracle deletions, then found additional fixture defects: a valid discover
request with an extra envelope member was accepted; cancellation accepted
`rpc-03` and null request IDs; and null call arguments raised an uncaught Python
AttributeError. Its original probes remain in
`checkpoint-evidence/ch11-review-foundation-initial.json`. These were checker
fixture findings, not an Ensemble runtime failure or a teaching gap. The earlier
"no initial failure" statement describes the original preparation run only.

After reloading the full coding skill, the narrow repair requires the exact
request/notification envelope keys, string method and object params; validates
the cancellation ID before recording it; and checks call arguments are an object
before accessing text. No additional schema or runtime promise is claimed.

All original 69 assertions and nine deletion controls remain. Seven new cases
each exercise a valid parent, the precise direct refusal, and an actual subprocess
refusal with exit 2, empty stdout and exactly the safe error code on stderr. They
cover the reported defects plus missing cancellation ID, null cancellation params
and array call arguments. The revised run is 90/90, with all nine original oracle
deletions still producing their intended failures. Receipts are
`checkpoint-evidence/ch11-foundation-peer-repaired.json` and
`checkpoint-evidence/ch11-foundation-peer-repaired-deletions.json`; original
receipts retain their original identities. This remains preparation-only coverage.

## CLI runtime integration before implementation

A different independent grader engineer continued this work after Chapter 10
acceptance. This engineer read the entire coding skill, architecture and current
Chapter 11 contract and the retained foundation. Its prior authorized Chapter 10
historical review included first-edition persistence code; no old Chapter 12
runtime or future Chapter 11 student implementation supplied this checker.
Foundation `c80af0a` and its independent review `d13346c` remain unchanged.

The new [student-facing command and build receipt contract](chapter-11-runtime-command.md)
is suitable for the author/coordinator handoff. The runtime command accepts an
explicit CLI, source directory, immutable revision, existing Chapter 10-style
build association and new receipt destination. It checks the complete required
source map and all supplied entries against both Git and the chosen source tree,
plus executable identity, before any launch/receipt write and after execution.
It does not build or inspect private MCP interfaces.

`scripts/edition2/ch11_effect_peer.py` wraps the retained literal protocol peer
with a real bounded notebook append and a separate request/response receipt.
The original peer still does no file effect. Discovery alone cannot satisfy the
new successful-call assertions. The CLI must produce the intended notebook
bytes once, ordinary Job artifact containing exact canonical result bytes,
selected alias declaration and recorded Anthropic continuation. The peer also
supplies ignored metadata to verify its exclusion. The valid configuration uses
a relative executable/cwd, a selected allowlisted canary and an unused connection
whose command does not exist. No real credential is read or inherited.

After that real runtime parent passes, three cases check invalid arguments
with zero `tools/call`, whole mixed text/image refusal with a safe failure
artifact, and missing referenced configuration before any process/model start.
The real effect can occur before the mixed-result refusal; the test does not
pretend refusal can undo an external edit. Inputs, commands, stdout/stderr,
model requests, peer records, notebook and artifacts are retained in compact
receipts. Each workspace is removed. Owned peer cleanup checks explicit process
identity; forced cleanup is recorded and cannot count as normal joined close.

### Executed preparation and predecessor baseline

- `checkpoint-evidence/ch11-runtime-preparation-commands.json` preserves exact
  commands, exit/output and final checker hashes. Four Python test methods pass,
  including actual subprocess notebook writes and mixed replies. Canned complete
  captures and targeted predicate deletions exercise the checker assertions;
  these are **not** runtime positive parents or student behavior mutations.
- `checkpoint-evidence/ch11-runtime-ch10-baseline.json` retains the initial
  accepted-Chapter-10 attempt: one vacuous metadata-absence assertion passed
  while five positives failed. Inspection corrected that predicate to require
  actual peer metadata and a returned result before awarding exclusion.
- `checkpoint-evidence/ch11-runtime-ch10-baseline-final.json` binds runtime
  `70d86f7419c82fcf7cb8a394d54e472feccd2eed`, the accepted `ch10` source export,
  the complete 164-entry original build-input map and CLI SHA256
  `60c924755420831a8a3a73bdd080f2c247087a09f05784fd6aba25decd618671`.
  The Chapter 10 CLI exits 1 with its usage message before launching the peer or
  making any HTTP request: **0 pass / 6 fail / 3 blocked**. This establishes the
  missing-feature baseline, not a regression in accepted Chapter 10.
- From that successful binding path, three precise changes independently test
  wrong binary hash, wrong source hash at a valid path, and a missing required
  source-map entry. Each reaches its intended identity refusal before writing
  the output receipt. No runtime mutation credit is claimed.
- `checkpoint-evidence/ch11-runtime-foundation-retained.json` retains **90/90**;
  `checkpoint-evidence/ch11-runtime-foundation-deletions-retained.json` retains
  all nine intended oracle deletions. Foundation script bytes are unchanged.

No Go build, provider call, runtime/source edit or large payload was needed.
Only owned small temporary peer/workspace/binding-control directories were
removed. The original accepted binaries, source, failed attempts and evidence
remain. No passing Chapter 11 runtime parent exists yet; independent integration
review and eventual source-bound positive/deletion runs remain gates.

### Full §11.10 remaining map

| Required property | Initial runtime subset | Still required |
|---|---|---|
| Transport independence | Real CLI stdio invocation prepared | Actual stdio/memory same-suite proof, external public adapter, ownership/copies and intended pipe-assumption mutation |
| Protocol and bounds | Literal discovery/list/call and metadata | Pagination, exact/+1 limits, malformed/unknown/stale IDs, reverse-execution refusal and allocation bounds |
| Authority | Selected alias and local invalid-argument no-send | Distinct two-Agent grants, forced hidden calls, Skills unload/admitted lifetime, durable/artifact-failure no-send |
| Schema/results | Exact artifact/continuation and mixed-content whole refusal | Complete schema profile, local refs, precision, budgets, isError, result/message limits |
| Lifecycle | Bounded invocation and owned cleanup observations | Out-of-order/racing replies, permits/cancel memory, stalled delivery, EOF, generations, shared isolation, joined close, owner-state uint64 seams |
| Persistence | None | v1/v2 identity and pre-prepare validation, snapshot/tail equivalence, offline no-effects and no restored remote workers |
| Existing behavior | One local Anthropic continuation and ordinary Job artifact | All three renderers, report limits, policy/hints/interrupt/Skills, optional module boundary and retained gates |
| Public usability | Published machine CLI command/config subset | Human CLI, GUI safe watch/cards, public headless/custom adapter, immutable snapshots and reviewed real-model feature matrix |

The remaining checks require the student's documented public seams and semantic
codec where appropriate. This table is a finite integration boundary, not a
replacement acceptance matrix or a waiver of any printed row. Race checks,
delivered-module validation, architectural review, student-behavior deletion
controls, initial live experience and subsequent historical quality comparison
remain required. No speculative new feature is demanded.
