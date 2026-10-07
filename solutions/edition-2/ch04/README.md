# Chapter 3 student implementation

A standard-library Go library, human and JSON-lines CLI clients, and separate
optional GUI client stub. This is the canonical second-edition source in the
outer repository. Earlier student checkpoints and receipts remain preserved;
Chapter 3 human terminal evidence is in `evidence/ch03/human-chat/`.

Build the CLI with `go build -o ensemble ./cmd`. Select `LLM_VENDOR`, a discovered
`LLM_MODEL`, and an API key through the process environment. Use a fresh
`CH02_LOG` path: existing logs are never overwritten or resumed. Run
`ensemble chat` and type an ordinary single-line request at `You> `. With no
arguments, actual terminal stdin and stdout select chat; redirected input selects
the machine protocol. Explicit `ensemble protocol` takes one JSON `user`,
`ephemeral`, or `redact` directive per line.

Human commands are `/help`, `/usage`, `/history`, `/ephemeral TEXT`,
`/redact FROM TO REASON`, and `/quit`. History shows `tool_returned` sequences
and call IDs so a person can select a real redaction target. EOF also exits with
four usage counts. Prefix a literal leading slash with another slash. Human
input is UTF-8, at most 1 MiB per line excluding LF or CRLF.

The CLI enables all six synchronous tools: read, list, search, write, edit, and
run_command. Use a scratch working directory: this workspace is not a sandbox.
The same public submission operation runs up to sixteen model requests per
turn, records tool results, and returns the final answer. Ordinary tool errors
can be corrected by the model; provider or persistence failures end the session.

`ensemble dump` reads `CH02_LOG`; `ensemble render LOG` renders the specified log
for the selected vendor/model. Neither operation needs credentials or contacts a
provider. `LLM_RESOLVED_MODEL` explicitly identifies an alias's exact target when
replaying bound material; returned model identity is not an automatic alias map.

Applications import `example.com/ensemble`, construct one Ensemble, then create
Agents with separate log paths. Prompt/directive requests name the Agent and
return their own result. Observers receive ordered owned event snapshots. The
public append path ingests supplied tool results. Agents explicitly select their
built-in tools; an omitted set exposes none. Each Agent owns its absolute
workspace and Registry; declaration and dispatch use the same visible set.
Loading a log is read-only. A failed append permanently faults that Agent.

The independently built program in `examples/tools-consumer` demonstrates
workspace isolation, per-Agent tools and observations. `examples/consumer`
retains the earlier controlled-result ingestion example and its independent
Agents and model switching; its second model uses `DEMO_SECOND_MODEL`.
The optional `gui` module demonstrates the same public client boundary using a
fake-backed integration test. It has no browser or WebSocket transport yet.

Checks run separately in all four Go modules: `go vet ./...` and
`go test ./... -count=1`. From the course root, the inherited grader command is
`make grade-dir CH=3 DIR=solutions/edition-2/main`. See
`evidence/ch03/human-chat/FEATURES.txt` for the human integration's actual results
and review status. Earlier machine-interface receipts remain separately labeled.
