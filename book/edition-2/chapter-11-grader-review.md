# Chapter 11 independent checker preparation

Status: runnable foundation, **no Ensemble runtime acceptance**. Chapter 10 and
Chapter 11 implementation are pending. This report does not release either
chapter or substitute fixture runs for its live demonstrations.

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
