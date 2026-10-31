# Edition 2 solution renumbering: grader validation

Coder: `enhancement_coder_resumed`, October 10, 2026. Validation complete after
the coordinator's filesystem-ready signal. Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`.

Bill authorized aligning the current solution directories with Edition 2 chapter
numbers while preserving the original-numbered snapshots separately. Coordinator
owns archive/moves and prose links. This coder changes only executable grader
defaults and five existing test fixture paths: Chapter N now selects
`solutions/edition-2/chNN`, including Chapters 4–8. The previous Chapter 4 special
case and later `chapter+1` offset are removed. Nested consumer/workflow/stream
module names `ch05`, `ch06` and `ch07` remain unchanged, as do their relative
module replacements, runtime source, grader predicates, scores and tests.

Searches of `cmd`, `internal`, `scripts` and Makefile found no additional direct
current executable references requiring moved-root updates. Historical
`scripts/edition2` evidence/main references and first-edition paths remain intact.
No new tests, mocks, live calls or original grader invocation are planned.

## Actual validation

Commands ran sequentially from the shared repository after the move; the reviewer
held all harness execution. No isolated checkout or runtime adaptation was needed.

| Command or check | Actual result |
| --- | --- |
| `go test ./internal/gradeedition2 ./cmd/grade-edition2 -count=1` | PASS; grader package 9.451s, command package has no tests. |
| `go vet ./internal/gradeedition2 ./cmd/grade-edition2` | PASS, no diagnostics. |
| `go list -m -json ensemble` in every nested module under current Chapters 4–8 | All 17 resolve to their own enclosing relocated core directory. |
| `go run ./cmd/grade-edition2 -ch 8 -json` | **100/100 PASS**, exit 0; 14.645s elapsed including launch/build. No submission or application-path override. |
| `gofmt -l` on the six changed Go files and scoped `git diff --check` | Clean, no output. |

The complete default Chapter 8 run earned 25/25 round trip, 20/20 connection,
20/20 broadcast, 15/15 persistence and 20/20 corrected Chapter 7 parity. This
exercises the inherited chain against the renamed final snapshot and discovers
the preserved nested consumer/workflow/stream modules without special mappings.
It is a new relocation check, not a rewrite of historical chapter scores.
No behavioral predicate or test was added, so existing targeted mutation evidence
remains applicable; no mutation or paid-provider sweep was repeated.

The entire executable diff is five removed default-offset lines plus one fixture
path substitution in each of five existing tests. Nested suffixes and relative
module replacements remain unchanged. Archive/file preservation is independently
checked by the coordinator and reviewer against pre-move revision `c2df11b`;
their manifest is `solutions/edition-2/history/mapping.json`.

Raw command statuses/timings, test output, full grade JSON, module-resolution
records and six changed-file SHA-256 values are retained at
`/var/folders/p_/11vk6y114xq7k9g5r_pc3yv80000gn/T/ensemble-renumbering-oni2ry87`.
All coder execution is complete; the harness is released for independent review.
No source mutation, original grader run, commit or further chapter work was done.
