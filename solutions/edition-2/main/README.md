# Chapter 5 student implementation

A Go library, human and JSON-lines CLI clients, and separate
optional GUI client stub. This is the canonical second-edition source in the
outer repository. Earlier student checkpoints and receipts remain preserved;
Chapter 5 evidence is in `evidence/ch05/`. The
[chapter validation record](../../../book/edition-2/chapter-05-validation.md)
distinguishes implementation, live review and checkpoint acceptance.

Build the CLI with `go build -o ensemble ./cmd`. Select `LLM_VENDOR`, a discovered
`LLM_MODEL`, and an API key through the process environment. Use a fresh
`CH02_LOG` path: existing logs are never overwritten or resumed. Run
`ensemble chat` and type an ordinary single-line request at `You> `. With no
arguments, actual terminal stdin and stdout select chat; redirected input selects
the machine protocol. Explicit `ensemble protocol` takes one JSON `user`,
`ephemeral`, or `redact` directive per line.

Human commands are `/help`, `/usage`, `/history`, `/ephemeral TEXT`,
`/redact FROM TO REASON`, `/hint TEXT`, `/interrupt`, and `/quit`. History shows `tool_returned` sequences
and call IDs so a person can select a real redaction target. EOF also exits with
four usage counts. Prefix a literal leading slash with another slash. Human
input is UTF-8, at most 1 MiB per line excluding LF or CRLF.

The CLI enables read, list, search, write, edit, `run_command`, `wait_for_job`,
`send_input`, `kill_job`, and `tool_limits`. Managed commands use a PTY and can
continue between reports. Use a scratch working directory: this workspace is
not a sandbox.
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

Checks run separately in all six Go modules: the root, `gui`, and the four
directories under `examples`. Run `go vet ./...` and `go test ./... -count=1`
in each, plus `go test -race ./... -count=1` in the root module. From the course
repository root, run:

```sh
PATH="$HOME/go/bin:$PATH" python3 scripts/edition2/accept_ch05.py solutions/edition-2/main
python3 scripts/edition2/audit_ch05_mutations.py solutions/edition-2/main
```

The historical Chapter 6 grader retains its original fixtures and is recorded
separately. Current live evidence combines the original Anthropic/OpenAI paths
with revised input checks and Gemini 3.8 Flash demonstrations; see
`evidence/ch05/gemini38-addendum-plan.json` for each run's source identity.
Earlier evidence stays attributed to its original scope.

## Chapter 5: responsive turns

`chat` accepts ordinary prompts while a turn runs and prints each request ID.
Use `/hint TEXT` for one-request guidance to the active turn, `/interrupt` to
end that turn while preserving its jobs, and `/quit` for full cleanup. EOF lets
all admitted prompts finish before cleanup. `protocol` preserves legacy user
records and also accepts explicit `kind: prompt`, `kind: hint`, and
`kind: interrupt` records with attributable reliable completions.

Embedding clients use `Agent.Submit` for a reusable request handle, `Wait` to
read its completion, and `Cancel` to cancel only that request. `Agent.Prompt`
is a blocking wrapper over the same actor. `Ensemble.Collect` creates an
independent collection that drains all currently ready handles in declared
order; cancellation of a collection wait does not cancel any request.

The optional GUI module exposes these public seams and remains a transport
stub. Display subscriptions have their own bounded queues and report
`overflow` through `SubscriptionStatus`; reliable request completion does not
rely on display delivery. Model responses are nonstreaming.

`examples/workflow` builds a public author/editor/reviewer consumer. Its
`collection` mode demonstrates two Agents, queued cancellation, reusable
handles and independent collections. Set `ENSEMBLE_RUN_DIRECTORY` to a fresh
directory and use the ordinary `LLM_*` environment configuration.

For an event log with captured request configuration,
`ensemble replay LOG SEQ` reconstructs that exact request's bytes from the
prefix before `request_sent`. This path performs no model or tool effect.
