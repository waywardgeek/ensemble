# Chapter 5 independent grader review

Independent grader work on 2026-10-07. The published Chapter 5 contract,
architecture, entire coding skill, and chapter-writing procedure were read before
writing probes. The seven scoring categories retain weights 10/25/15/15/15/10/10.
The checker tests the new contract rather than assigning points for first-edition
exercise filenames or observation spelling. Student code and historical graders
were not edited by this grader handoff.

Run from the outer repository:

```sh
PATH="$HOME/go/bin:$PATH" python3 scripts/edition2/accept_ch05.py solutions/edition-2/main
python3 scripts/edition2/audit_ch05_mutations.py solutions/edition-2/main
```

`--only`, `--skip-prior`, and `--no-race` are iteration controls, not full local
acceptance. Output records `full_local_scope`, per-category command results,
source file hashes, checker hashes, and score. Actual live-provider receipts,
semantic code review and manuscript acceptance remain separate gates.

## Baseline and compatibility findings

The unchanged `make grade-dir CH=6 DIR=solutions/edition-2/main` returns 10/100.
Its six failures mix incompatible fixture assumptions: a first-edition `ch06`
exercise and stdout observation schema, pipeline-label logging, a logger analysis
requiring concrete runtime structs in `common`, and refusal of unknown text-only
models. The current contract instead permits private runtime structs with common
interfaces, uses reliable completion handles, and explicitly preserves valid
unknown-model text requests. A missing fixture invocation is not evidence that
an actor is deaf. The inherited grader remains unchanged for historical trees;
its score is retained in `checkpoint-evidence/ch05-inherited-grader.txt`.

Two additional harness findings are preserved in the initial receipts:

- The process PATH contained literal `~/go/bin`. The shell found installed Delve,
  but Python's executable lookup did not expand the tilde. The documented command
  supplies the expanded path; no debugger prerequisite was removed.
- The Chapter 4 debugger fixture treated the next `You>` prompt as model
  completion. Chapter 5 must print that prompt while work is pending. A separate
  `accept_ch05_prior.py` adapter now follows the acknowledged request identity to
  its completed human answer. Every original debugger, job lifecycle, artifact,
  redaction, usage and wire assertion still runs; the original fixture is intact.

## Genuine student defect and correction

The initial independent unsupported-media probe received exactly
`unsupported reference` for a well-formed `video/mp4` URI rendered for a selected
OpenAI model. Section 5.8 requires model, media category/MIME and missing mapping,
without the private locator. This was reported as a student defect separately
from fixture mismatches. The coder corrected diagnostics and added student
regressions without reading checker implementation. The independent refusal
probe subsequently passed. No architecture or behavior requirement was waived.

## Independent coverage

| Category | Independently observed properties |
|---|---|
| Parity, 10 | All 53 Chapter 4 CLI/offline checks, including actual local Delve through a human PTY on all three API surfaces; sixteen model requests, paired final batch, no seventeenth HTTP operation, accepted-response usage |
| Responsiveness, 25 | Actually overlapping distinct prompts, FIFO, blocking and asynchronous callers, queued/active cancellation, reusable owned completions, 1,024 admitted queued requests, controls during blocked HTTP/report waits, busy setting refusal, interrupt/close settlement, preserved call pairing, late job fact, subsequent success, deterministic stale worker rejection, prepared-report cursor transactions and persistence-failure settlement; real CLI protocol and PTY controls on all three local APIs |
| Replay, 15 | Actual captured request bytes against exact prefix/configuration on all APIs, one-request hints/ephemera, consumed sequence identity, repeated pure replay, changed current configuration, removed-hint negative control, rejected invalid turn histories |
| Observers, 15 | Attributed durable response sequence/request/part position, text/call/opaque typed parts, callback/snapshot ownership, observable overflow and reliable completion despite a parked display consumer |
| Completion collection, 15 | Both-ready barrier with declared order, staggered arrivals, exhausted and empty collections, canceled wait preserving members, independent collections and individually rereadable handles |
| Loud refusal, 10 | Actual unsupported media render fails with selected model and MIME while withholding locator; text-only unknown models succeed in ordinary fixtures |
| Architecture, 10 | Every discovered spoke and nested executable, common behavior placement, actual runtime parent identities and logger reachability, common-declared interface links, actor access through owning Agent, Job→Jobs→Agent chain, new-spoke/global-state negative controls, headless dependency walk, all module builds and public author/editor/reviewer workflow with actual local model responses/file edits |

The external public test adapter uses published APIs. The narrow structural,
transport and persistence adapters name implementation fields only to reach the
property under test; spelling earns no points. A positive control renames actor
parent identifiers while preserving behavior. Source scans discover all packages
and modules instead of maintaining a historical hardcoded spoke list.

Fixtures park HTTP on explicit channels. The late-response transport deliberately
returns only after a later request is active. Collections complete both members
before the coalesced drain. Observer copy tests use a small response separately
from forced overflow, avoiding the assumption that a fast callback can never be
descheduled. An interrupted tool report uses the published `run_command`, not a
legacy production `think` tool.

The source scan catches imported sideways implementations and mutable package
session declarations; it is not a general proof that every helper's semantics or
all indirect mutations are correctly owned. Independent semantic review remains
required. The mutation record documents actual protected properties and score
losses; it is not a claim that every sentence of the chapter has a separate
mutation. Live timing/provider behavior is outside deterministic local coverage.

## Validation receipts

Initial failed and corrected runs are retained under
`checkpoint-evidence/ch05-independent-*.json`; the adapted prior-only result is
`ch05-parity-adapted.json`. The final complete invocation and deletion audit are
`ch05-independent-final.json` and `ch05-mutations-final.json`. The final collection
fixture places both Agents on the same Ensemble (avoiding an unintended
requirement to accept another application’s handles); its two affected deletion
controls are rerun in `ch05-mutations-collection-final.json`. These receipts bind
the copied source files and checker assets actually used; they do not relabel
older attempts as runs of a later source.

Legacy regression `go test ./internal/grade ./internal/fakevendor -count=1`
passed unchanged (grade 503.930s). This includes historical reference and deletion
controls. Root vet passed. The final local-check record contains the broader
root test command, formatting result, and script syntax checks.
