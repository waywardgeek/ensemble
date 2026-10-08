# Chapter 6 zero-byte argument maintenance

This is disclosed maintenance by the independent grader/reviewer, who has seen
later second-edition source and checks. It is not another cold student attempt.
The patch was derived here from Chapter 6's Messages assembly contract, clarified
at `305b1b00d1565a73ee9312aaafef7c652e677772`, and the unchanged observed SSE body.
No later implementation or test file was copied backward. No provider call or
credential access was made for this correction.

The isolated branch `edition-2/ch06-empty-args-r2` starts at accepted export
source `c3fa7583c5e4c3dcf026b3d5a0a93da000dd8307`, main tree
`91f6da877ab76fcda6887bc3ec523332a499fbd0`. The original annotated
`edition-2-ch06-r1` checkpoint remains unchanged. Its complete tracked evidence
remains in this branch; sparse checkout omitted it from the working directory.
Production changes are limited to five lines in `internal/llm/stream.go`.
Runtime and regression source are frozen at
`788c5e9a6928fb125c9e606336ad3a67632b6cc6`.

The provider fixture is exactly 2,668 bytes, SHA-256
`e6f7b7446b3dab2fe5ef6466111a6790c31ac52d45eccc8296a416ea9a4971ef`.
It was originally recorded during Chapter 9 Anthropic G, response 002, produced
by runtime `c0e3171fdc22834348f976f81fcaa7372bd13ed4` and preserved in `75a7255`.
Git diff whitespace checking reports the fixture's final blank line. That line
is the SSE terminal frame delimiter and part of the recorded hash; it is
intentionally preserved. Changed Go files have no formatting differences.

Replaying those bytes locally establishes parser behavior; it is not a new
live Chapter 6 request and does not relabel the original failed operation.

The start input remains available until an argument delta supplies actual
bytes. Zero-byte fragments do not allocate a replacement or publish empty
display fragments. Whitespace is still replacement content and fails object
validation. A wrongly typed fragment retains its parser error. Existing block,
message terminal, usage, and whole-response acceptance gates remain in force.

## Checks and retained failure

Run commands from `solutions/edition-2/main` in this isolated checkout unless
an absolute checker path is shown. The source-bound summary is `review.json`.

* `go test ./internal/llm -run '^TestMessagesZeroByteArguments$' -count=1 -v`
  failed on the original production source in exactly the recorded, one-empty,
  and two-empty cases (`before.txt`). Ten controls already passed. The identical
  13 cases pass after correction (`after.txt`). They retain replacement rather
  than merge, invalid/missing starts, whitespace, malformed/non-object JSON,
  wrongly typed deltas, terminal requirements, exact usage, and one complete
  argument display. One-byte reads exercise the unchanged original body too.
* `go vet ./...`, `go test ./... -count=1`, and `go test -race ./... -count=1`
  are retained in `vet.txt`, `tests.txt`, and `race.txt`. Changed Go files have
  empty `gofmt -l` output.
* `python3 /Users/bill/projects/ensemble/scripts/edition2/accept_ch06_gate.py
  788c5e9a6928fb125c9e606336ad3a67632b6cc6` produced `full-gate.json`:
  22 of 23 command groups passed, including all seven modules, six CLI barriers,
  61 wire cases, all 14 independent contract groups, prior Chapter 5 assertions,
  CLI overflow controls, and assembly observations. The mutation group recorded
  18 passing deletions and one **adapter failure**, not a runtime failure:
  the old globally unique `if text == "" {` anchor now occurs twice.
* `python3 evidence/ch06/empty-arguments-r2/thinking-mutation.py
  /Users/bill/projects/ensemble/scripts/edition2 .` supplies only an
  `emit`-qualified anchor. The original positive, thinking-suppression mutation,
  and intended refusal are unchanged; both controls pass. Its separate receipt
  is `thinking-mutation.json`. The original failed gate is not relabeled green.
* `python3 /Users/bill/projects/ensemble/scripts/edition2/ch09-review-empty-stream.py
  /Users/bill/projects/ensemble-edition-2-revisions/executables/ch06-empty-args-r2-cli
  --receipt evidence/ch06/empty-arguments-r2/cli-barriers.json` passes all 14
  local CLI cases. This existing independent checker exercises only Chapter 6
  functionality. It includes exact recorded bytes, valid replacements, malformed
  and incomplete operations, provider errors, and a held `message_stop` barrier.
  Before that barrier there is no accepted part, completion, response, or file
  effect; afterward the valid call executes once. Invalid operations have no
  tool effects. The executable was retained from the full gate's single build.

These checks preserve original assertions and original evidence. They are
scoped validation, not an exhaustive proof or a new all-provider demonstration.
The original live sessions retain their old identities. Root review and any
export/tag decisions remain separate; this branch has not changed current main
or any frozen chapter export.

## Proposed forward propagation, pending root review

Accepted Chapter 7 r2 export source is
`9ec94ef551b08de6fbd714bf95efee901db49400`, main tree
`944b187319224bc2d17b833ae1c72bfd93643c2e`; Chapter 8 r1 source is
`446d7f27fcef052d6bb780901c786842dc5286c1`, main tree
`58d3fbf4ec6493839985c6968a13c1e6a7bdad83`.
Their `stream.go` files are byte-identical to accepted Chapter 6's file
(SHA-256 `949ae85628d55fe3c6d6514601ec967b205b7dd0f8b2ed2fb991ba97e719f1ea`).

After review, create separate sparse worktrees at those exact sources. Apply
the five-line production correction and this Chapter 6 regression/fixture,
preserving each stage's existing source and receipts. Run the affected core
vet/tests/race, the 14 local CLI cases, complete delivered-module discovery,
and each stage's applicable retained streaming/client checks; bind any reused
unchanged evidence explicitly. Browser settings and speech behavior are not
changed by the parser correction. No Chapter 7/8 implementation, export or tag
has been made as part of this proposal. Root determines final checkpoint scope.
