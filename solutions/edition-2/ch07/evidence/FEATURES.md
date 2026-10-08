# Chapter 1 implementation receipts

Date: 2026-10-07. Toolchain: Go 1.25.4, darwin/arm64. Source hashes in
`source-sha256.json` identify the live-run code, tests, and consumer. Previous WIP
commit: `4d8ee165956d8acbb761bd60bc33ce7dcf8f1ea2`.

The post-run diagnostic review revision has its own
`reviewed-source-sha256.json`. Existing live evidence remains bound to the
preserved passing snapshot `459e4ce`; local checks validate the subsequent
error-classification-only change without repeating paid calls. See
`DIAGNOSTIC-REVIEW.md` for the reviewer findings and their resolution.

## Actual live runs

Discovery returned `claude-sonnet-5-5` at 17:01:14 UTC; no further pages.
This ID was selected from that response, never guessed or built in.

The student prepared the program and credential-safe helper; the coordinator
executed final live commands because the student worker temporarily retained
old sandbox permissions. The student inspected the receipts. These runs used
the real executable and separate public-library consumer, not internal calls.

Build commands, from this module and `examples/consumer` respectively:

```sh
go build -o /tmp/ensemble-ed2-ch01 ./cmd
go build -o /tmp/ensemble-ed2-consumer/demo .
```

Executed helper commands:

```sh
python3 /tmp/ensemble-ed2-live.py discover
python3 /tmp/ensemble-ed2-live.py cli claude-sonnet-5-5
python3 /tmp/ensemble-ed2-live.py consumer claude-sonnet-5-5
```

The temporary helper selected the authorized Anthropic credential in memory,
passed it through child environments, and recorded sanitized outputs. Its
base URL was `https://api.anthropic.com///`, exercising normalization. The
durable CLI and consumer run with `ANTHROPIC_API_KEY`, `ANTHROPIC_MODEL`, and
optional `ANTHROPIC_BASE_URL` in their environment. Consumer source and complete
live inputs/outputs are retained here; the helper itself remains temporary.

Final CLI at 17:06:50 UTC: invented `Silent Harbor`, recalled `Silent Harbor`,
then reversed it as `robraH tneliS`. EOF reported input=261/output=209. Exit 0,
stderr empty. Earlier output=222 preceded the transport correction; the retained
final receipt is authoritative.

Final consumer at 17:08:33 UTC: Agent A recalled `CORAL-271` with usage144/118;
Agent B recalled `HERON-839` with usage144/28. Assertions verified separate
histories and unchanged usage across another failed Agent. Output included
`Independent=true`; exit0, stderr empty. A deliberately local connection failure
with a placeholder key reached the root capture writer as
`ensemble: model transport failed`. This was not a live-provider failure.

## Feature checklist

| Feature | Action and observable evidence |
|---|---|
| Model discovery | Provider response in `live.json`, including selected ID |
| Configured raw HTTP | Real CLI accepted chosen model/key and trailing-slash base |
| JSON-lines interface | Three questions then EOF; exact protocol output in `live.json` |
| Complete conversation | Assistant-invented fact recalled and correctly reversed |
| Usage accumulation | CLI 261/209; separate consumer totals 144/118 and 144/28 |
| Independent Agents | External consumer performs two real conversations and asserts history/usage separation |
| Parent logging | Consumer's local transport failure reaches Ensemble's capture writer |
| Missing configuration / invalid input | CLI probes fail cleanly: `cli-negative.json` and independent acceptance |
| Empty input / blank lines | Local probes emit zero usage |
| Multiple text blocks / exact prefix | Inherited grader, public-library tests, independent acceptance |
| Invalid responses / no retry | Local fault tests and independent 23-check acceptance report |
| Failed turns do not commit | `TestIndependentAgentsAndFailureAtomicity` resumes after rejection |
| Finite timeout | Independent stalled-server test: 60.01 seconds |
| Star / owned state | Actual imports, constructors, fields and logger path reviewed |
| Owned transport | Regression rejects nil, global, or shared Engine transports |

Chapter 1 has one vendor and no GUI, tools, streaming, persistence, or optional
chat mode. No such features are claimed. Deterministic faults are distinct from
paid-provider receipts. Broader mutation audits and editorial review belong to
the coordinator and are not claimed complete by this student evidence.

Final local checks: scoped `gofmt -l` empty; `go vet ./...` and
`go test ./... -count=1` pass in both modules (consumer has no test files).
Course-root `make grade-dir CH=1 DIR=solutions/edition-2/ch01` passes 100/100,
five requests, fake usage406/65. Independent CLI acceptance: 23/23.
