# Chapter 2 student implementation

A standard-library Go library, JSON-lines CLI, and separate optional GUI client
stub. This repository derives from the reviewed Chapter 1 snapshot `75542c1`.
Historical Chapter 1 evidence remains in `evidence/`; Chapter 2 receipts belong
under `evidence/ch02/`.

Build the CLI with `go build -o ensemble ./cmd`. Select `LLM_VENDOR`, a discovered
`LLM_MODEL`, and an API key through the process environment. Use a fresh
`CH02_LOG` path: existing logs are never overwritten or resumed. Each input line
contains exactly one `user`, `ephemeral`, or `redact` directive.

`ensemble dump` reads `CH02_LOG`; `ensemble render LOG` renders the specified log
for the selected vendor/model. Neither operation needs credentials or contacts a
provider. `LLM_RESOLVED_MODEL` explicitly identifies an alias's exact target when
replaying bound material; returned model identity is not an automatic alias map.

Applications import `example.com/ensemble`, construct one Ensemble, then create
Agents with separate log paths. Prompt/directive requests name the Agent and
return their own result. Observers receive ordered owned event snapshots. The
public append path ingests supplied tool results; this chapter executes no tools.
Loading a log is read-only. A failed append permanently faults that Agent.

The independently built program in `examples/consumer` demonstrates tool calls,
controlled result ingestion, redaction, observers, independent Agents and model
switching. Its second discovered model is selected with `DEMO_SECOND_MODEL`.
The optional `gui` module demonstrates the same public client boundary using a
fake-backed integration test. It has no browser or WebSocket transport yet.

Checks run separately in all three Go modules: `go vet ./...` and
`go test ./... -count=1`. From the course root, the inherited grader command is
`make grade-dir CH=2 DIR=solutions/edition-2/ch02`. See `CHECKPOINT.md` and
`evidence/ch02/FEATURES.md` for actual results and remaining review gates.
