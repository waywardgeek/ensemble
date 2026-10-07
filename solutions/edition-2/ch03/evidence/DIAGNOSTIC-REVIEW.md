# Post-run diagnostic review revision

The independent reviewer compared initial passing snapshot `459e4ce` with the
first-edition standard. The student received the review rationale without reading
or copying that implementation. Passing behavior and ownership were stronger,
but blanket error messages hid whether a request was canceled or timed out.

The revision classifies request and response-read failures using known safe
context sentinels and network timeout interfaces. It wraps only
`context.Canceled` and `context.DeadlineExceeded`, preserving `errors.Is`
without retaining raw transport errors, credential-bearing URLs, or arbitrary
provider strings. Both decoder failure sites use the same parent-aware helper.
The existing generic transport diagnostic remains for other connection failures.

Named request literal fields make future edits safer. A concise WHY comment
explains why validated conversation pairs and usage are committed together.
No response parsing contract, successful request behavior, retries, or owner
relationships changed. Existing safety rationale was retained centrally.

Six public-library test cases exercise cancellation and deadlines before headers,
while awaiting a body, and after a complete JSON object while awaiting EOF.
They check safe diagnostic text, `errors.Is`, absence of URL/key markers, and
unchanged history/usage. The initial fixture hung in server cleanup because it
assumed server-side cancellation detection. Test-owned release signaling now
bounds cleanup independently of the behavior under test.

Final checks after that fixture correction:

- Main module: `go vet ./...`; `go test ./... -count=1 -timeout=20s` pass
  (public library 0.317 seconds).
- Consumer module: `go vet ./...`; `go test ./... -count=1` pass.
- Scoped `gofmt -l` prints nothing; `git diff --check` passes.
- Inherited Chapter 1 grader: 100/100.

The reviewer asked whether timing could mask parser-site coverage. Two disposable
copies removed the classification independently at each decoder site. With the
passing working tree unchanged, the first-decode mutant failed exactly
`body/deadline`; the trailing-decode mutant failed exactly `after-json/deadline`.
Each reported lost `context.DeadlineExceeded` identity and exited 1. The raw
results are retained in `diagnostic-mutations.json`.

The reviewer accepted the revision based on source inspection and those deletion
results. No paid calls were repeated for this local error-path change. Initial
live receipts remain bound to snapshot `459e4ce`; updated source hashes are in
`reviewed-source-sha256.json`. Coordinator-owned broader audits remain separate.

The coordinator subsequently ran the complete revised CLI acceptance: 23/23,
including the 60-second timeout. Its expanded mutation audit passed the control
and eleven mutants with exact expected failures. Retained reports:
`independent-cli-acceptance-reviewed.json` and
`independent-mutations-reviewed.json`. The student read these results, not the
independent grader implementation. Additional persistent architecture checks are
coordinator-owned and do not alter this solution.
