# Chapter 1 student checkpoint — 2026-10-07

Current status: student implementation and live checks pass; independent CLI
acceptance passes 23/23. Coordinator-owned broader audits and manuscript review
are separate. This is not a claim of Bill's editorial approval.

Post-run comparison review has now accepted the diagnostic revision. Initial
passing snapshot `459e4ce` remains in history. Cancellation and timeout errors
retain useful safe identities; body-read failures are classified at both decode
sites. The revision also names request fields and explains atomic pair/usage
commit. Local checks and two targeted deletion tests pass their expectations;
see `evidence/DIAGNOSTIC-REVIEW.md`. No paid calls were repeated for this change.
Revised independent acceptance also passed all 23 checks, including timeout;
the coordinator's expanded audit passed its control and eleven mutants with
exact expected failures. Both revised reports are saved in `evidence/`.

Final scoped formatting returned no paths. `go vet ./...` and
`go test ./... -count=1` passed in this module and in `examples/consumer`.
The inherited chapter grader passed 100/100. An independent 23-check CLI report
includes a measured 60-second timeout. All reports are in `evidence/`.

Real Anthropic model discovery, CLI conversation and separate public-library
consumer succeeded. Final model `claude-sonnet-5-5`; CLI usage 261/209; separate
Agents 144/118 and 144/28. The consumer also verified local failure logging via
the owner chain. No paid calls were repeated after final success.

The reviewer found that a nil Client.Transport used the global default. Each
Engine now creates its own transport, protected by a regression test. Final
live receipts use the corrected code. Source hashes identify the tested files.
The student read no first-edition implementation or grader implementation.

See `evidence/FEATURES.md` for feature mapping, exact commands, and limits.
No credentials are stored here; an in-memory scan against the actual credential
confirmed its absence. There are no running student sessions or approvals.

Next: coordinator completes broader audits/review. Begin Chapter 2 only from
its finalized contract, cloning this new history. Reload the complete coding
skill before new work. Chapter 2 needs clean core structures, CLI/browser-WS
client boundaries with a separate optional GUI module, and three-vendor live
demonstrations once its adapters exist.

## Historical interrupted WIP record (superseded above)

Status: unfinished, not validated. User requested stopping for a Codex restart.

Fresh implementation, with no first-edition code copied or consulted:

- `go.mod`: independent standard-library-only module, `example.com/ensemble`.
- `ensemble.go`: public library and composition root; Ensemble owns logger,
  Agent owns configuration/history, Engine created through the Agent.
- `internal/common/types.go`: shared data and parent interfaces.
- `internal/llm/engine.go`: request construction, parsing, transport, usage;
  logging follows Engine -> Agent -> Ensemble. Failed turns do not commit.
- `cmd/main.go`: JSON-lines CLI through public library, no optional chat mode.
- `ensemble_test.go`: newly written public-API tests for separate Agents,
  failed-turn atomicity, invalid responses, and parent-chain logging.
  THESE TESTS HAVE NOT BEEN FORMATTED OR RUN YET.

Commands actually completed before the test file was added:

- `git init`: succeeded in this directory.
- `gofmt -w ensemble.go internal/common/types.go internal/llm/engine.go cmd/main.go`.
- `go build -o /tmp/ensemble-ed2-ch01 ./cmd`: succeeded.
- `go vet ./...`: succeeded before tests existed.
- `go test ./... -count=1`: succeeded, reporting no test files at that time.
- `gofmt -l ensemble.go internal/common/types.go internal/llm/engine.go cmd/main.go`:
  empty output.
- From course root, `make grade-dir CH=1 DIR=solutions/edition-2/ch01`:
  initial sandbox attempt failed on Go build-cache access; approved escalated
  rerun PASSED 100/100, all seven inherited Chapter 1 checks, five requests,
  cumulative input=406, output=65. This does not prove edition-2 acceptance.

Live validation: NOT STARTED. No discovery API request or paid model call.
Programmatic settings schema inspection identified `directClaudeAPIKey` as a
string field; no value was printed or copied. No credentials are in this tree.

Outside this repo: compiled executable `/tmp/ensemble-ed2-ch01`. No new helper
scripts or evidence files. No known running coder process, tool session, or
pending approval. The grader tool call completed before interruption.

Next action after restart: freshly read root `AGENTS.md`, the entire
`book/edition-2/skills/ensemble-coding/SKILL.md`, current architecture/chapter
contract, and newly rewritten `book/chapter-writing-procedure.md` (the latter
has not yet been reread). Then format the new test file and run scoped
formatting, vet, tests, and edition-2 acceptance with the coordinator.

Remaining work: independent structural review and acceptance/mutation evidence;
real Anthropic model discovery using credentials only in memory/environment;
bounded CLI multi-turn live demonstration and external public-library consumer
for independent Agents/logger; local negative CLI probes; sanitized receipts;
rerun all required checks and commit a validated snapshot only after passing.
Coordinate grader enhancements with root; student must not read grader internals.

This checkpoint commit is explicitly WIP and intentionally precedes the
unrun test formatting/validation. It is a restart snapshot, not completion.
