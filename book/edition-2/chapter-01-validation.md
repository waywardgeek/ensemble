# Chapter 1 validation record

Date: 2026-10-07. Student initial run checkpoint: `459e4ce`; reviewed revision:
`75542c1c73388fa1ab618dbb8b1e252e3d80816a`, both in the independent
`solutions/edition-2/ch01/` Git repository. Bill's editorial approval is a
separate status and is not claimed here.

## Behavior and live evidence

The student reports clean formatting, `go vet ./...`, and
`go test ./... -count=1` in both the core and external consumer modules.
The inherited `make grade-dir CH=1 DIR=solutions/edition-2/ch01` scores 100/100.

The coordinator ran the revised executable through all 23 independent CLI
checks, including the actual 60-second stalled-response failure. All passed.
The revised deletion audit has a passing control and eleven compiled mutants;
all produced exactly the expected failing-check sets. No legacy grader was
modified. Commands and scope are in `scripts/edition2/README.md`.

The student also demonstrated that deleting each decoder's safe error
classification makes its corresponding public-API deadline test fail.
The positive control passed; mutants were disposable copies. See student
evidence for precise results and the reviewer report for the race concern
that prompted this additional evidence.

Sanitized live receipts, feature mapping, and source hashes are retained in
`solutions/edition-2/ch01/evidence/`. The actual Anthropic CLI and separate
public consumer ran successfully using discovered model `claude-sonnet-5-5`.
The final CLI totals are 261 input / 209 output; independent Agents retained
their own codes and totals. Logger failure evidence is explicitly a local
transport probe. The paid receipts are bound to the initial checkpoint;
the subsequent reviewed revision changes diagnostics, comments, and named
request construction. A new paid successful conversation was not represented
as having occurred on the revised hash.

## Architecture and quality

Independent source review checked actual constructors, imports, owner state,
and interface paths. It caught the implicit shared HTTP transport before final
live runs; the student fixed ownership and added a regression. Public tests
exercise independent histories/counters, failed-turn atomicity, history-copy
ownership, and parser access to the root logger.

The additional `packagecheck` report is clean. Its ten source-fixture
controls cover public CLI imports, a newly discovered sibling spoke,
common behavior/import violations, legitimate shared-type interface helpers,
and nested-module separation. Signature and shadowing counterexamples prevent
an interface-looking name from exempting unrelated behavior; equivalent type
aliases pass. A real-student identifier-rename control also passed, including
module tests and an external-consumer build. The checker is deliberately limited to those
source properties. It does not certify owner semantics or every possible
architectural violation; the source review and behavioral probes remain
required. This record does not claim an exhaustive mutation proof.

`chapter-01-code-review.md` compares the first-edition standard with the
student's initial answer and records the resulting revisions. Improvements
include reusable ownership, stricter response accounting, safe diagnostics
that preserve cancellation/deadline identity, clearer request construction,
and comments explaining the successful-exchange boundary. Teaching feedback
was incorporated into Chapter 1. Passing the grader alone did not end review.

## Legacy regression record

Before new independent tooling, the existing agent suite and Chapter 1
reference/mutation tests passed. The resumed full root baseline log shows all
packages passing; daemon recovery lost its process-exit receipt. That log is
saved as `checkpoint-evidence/legacy-root-completed.txt`, with that limitation.

After adding the package checker, root `go vet ./...` and a fresh full
root `go test ./... -count=1 -timeout=45m` passed, exit 0. The grade package
completed in 508.958 seconds. The checker's subsequent signature/symbol
hardening also passed its ten targeted controls and another full-root vet.
The complete suite log is `checkpoint-evidence/legacy-root-final.txt`.
Existing `agent/` and first-edition solutions were not edited.
