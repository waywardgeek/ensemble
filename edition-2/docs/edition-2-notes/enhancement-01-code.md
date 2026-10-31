# Edition 2 Chapter 1: grader enhancement

Contributor: `enhancement_coder`, October 10, 2026. Submitted for independent
review. Mandatory reviewer handoff: `.agents/skills/ensemble-coding/SKILL.md`.

## Change and provenance

The new entry point is `cmd/grade-edition2`; corrections live in
`internal/gradeedition2`. Chapter 1 reuses the original build, run, fake provider
and all seven checks. Its only corrected verdict is the existing 15-point
`wire` check: every recorded `anthropic-version` must equal `2023-06-01`, the
value already specified by original §1.2. Missing or wrong versions fail. No
new check, point total, wire protocol or runtime feature was introduced.

Original chapter revision: `e1605916114230570b348ab5961c193fa1cebf80`.
Latest shared revision affecting the reused harness/check/fake files:
`c24b8e84da112ef66f729c0cdc60413ee6fb547d`. Accepted student source remains the
unchanged `solutions/edition-2/ch01/` export at `edition-2-ch01-r1`.
The [student review](ch01.md) and [independent code review](ch01-code-review.md)
already reproduced this specific version-header survivor. Original graders,
fake servers, accepted source, frozen exports and historical results are untouched.

This is an additive correction using captured HTTP headers, not a new grading
framework. On a correct request every original check is returned unchanged;
the regression test compares the complete original check list. On an invalid
version, it preserves other failures and replaces the obsolete wire-success
note with an exact request-level diagnostic. The command returns through its
cleanup before exiting, including when grading fails.

## Actual checks

Run from the repository root:

```sh
go run ./cmd/grade-edition2 -ch 1 ./solutions/edition-2/ch01
go run ./cmd/grade-edition2 -ch 1 ./solutions/ch01
go test ./internal/gradeedition2 ./cmd/grade-edition2 -count=1
go vet ./internal/gradeedition2 ./cmd/grade-edition2
gofmt -l cmd/grade-edition2 internal/gradeedition2
```

Both corrected-grader runs passed **100/100**, all seven checks; five real local
HTTP requests, exact totals **476 input / 65 output**. The second verifies legacy
compatibility without copying its implementation. Focused tests and vet passed;
formatting listed no files. From `solutions/edition-2/ch01`, uncached
`go test ./... -count=1` and `go vet ./...` also passed. Initial sandbox attempts
could not access Go's cache or bind the fake server; approved reruns supplied
the results above. Those setup failures were not application failures.

For machine output, add `-json` before the directory. With no directory, the
command grades the accepted Edition 2 Chapter 1 export. Only Chapter 1 is
implemented at this handoff; later chapters are added after their own gates.
The executable uses exit 0 for pass, 1 for a failed grade and 2 for invocation
or setup failure. `go run` itself wraps a nonzero program exit and prints it.

## Targeted mutation and retained evidence

Copied the accepted source to `/tmp/ensemble-enhance-ch01-hloOxj` and changed only
the Engine header assignment from `2023-06-01` to `wrong-version`. The mutation
compiled, answered all five rounds and reached the real HTTP boundary.

| Command against the scratch copy | Observed result |
| --- | --- |
| `make grade-dir CH=1 DIR=/tmp/ensemble-enhance-ch01-hloOxj` | Original grader **100/100 PASS**; confirms the historical gap remains. |
| `go run ./cmd/grade-edition2 -ch 1 /tmp/ensemble-enhance-ch01-hloOxj` | Corrected grader **85/100 FAIL**; only `wire` fails, reporting wrong version for requests 1–5. |
| Same corrected command with `-json` | Same score and five exact-version diagnostics; all six unrelated checks pass. |

The first corrected report retained the original fake's “all well-formed”
success note alongside the failure. Removed that misleading note when the new
check first changes a passing wire verdict to failing; final tests and JSON
mutation verification passed with the corrected diagnostic. Regression tests
also exercise missing-version evidence and preserve the six unrelated checks.
Scratch source was removed after verification. The authoritative export never
changed; `git diff --quiet` verified the export and original grader paths.

Unchanged behavior reuses the [original mutation record](../../solutions/evidence/ch01/mutations.txt)
and [live chat transcript](../../solutions/evidence/ch01/live.txt),
already independently reviewed. No new paid live run was needed for an offline
grader correction. No mocks were introduced. Architecture, comment assessment,
and production/test counts remain those of the historical code review; this
pass changes grading code only. Frozen Chapter 1 predates the later requirement
for documentation of every exported field and method; its missing documentation
is historical, not a claim that the newer rule was satisfied then.
