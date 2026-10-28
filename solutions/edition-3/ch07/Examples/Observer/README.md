# Aggregate streaming observer

A separate .NET 10 assembly using only the public Ensemble library. Its local
counters measure delivered observations; they are not another History or a
lossless token-accounting system. Reliable `SendAsync` completion governs the
run. Root destruction closes the observation queue, which is drained before
printing one `stream_stats` record.

Build with `dotnet build Examples/Observer/Observer.csproj` from the endpoint.
Run the compiled `Observer.dll` from an owned scratch directory. Set an owned
loopback `LLM_BASE_URL`, plus `LLM_VENDOR`, `LLM_MODEL` and dummy `LLM_API_KEY`.
A command-line string is the prompt. No implicit public-provider endpoint is used.

- `CH07_NO_STREAM=1` forces plain delivery.
- `CH07_LOG` selects the aggregate JSONL file (default `stream.jsonl`).
- `CH07_AGENT_LOG` optionally selects the canonical event log.

The historical environment names remain compatible. Plain mode has zero deltas,
`has_delta: false`, and a zero first-delta duration meaning **not observed**,
not instantaneous delivery. Final parts, the reliable answer and measured usage
are still reported. Display bytes are UTF-8 bytes, not model tokens.
