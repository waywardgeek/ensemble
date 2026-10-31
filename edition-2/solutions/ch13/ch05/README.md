This separate Go module exercises Ensemble's public API. Its `shout` tool belongs
to one Agent and converts the supplied text to uppercase. The framework has no
knowledge of that tool.

Build here with `go build -o /tmp/ch05agent .`. Set `LLM_VENDOR`, `LLM_MODEL` to a
model available from that provider, and `LLM_API_KEY` in your environment. Run the
binary in a disposable directory:

```sh
/tmp/ch05agent 'Call shout with text hello framework, then report its result.'
```

The program prints the answer, usage and a timestamped tool diagnostic. Inspect
`events.jsonl` for the actual call/result; `cr/agent-1/io/` holds full job output.
The optional `LLM_BASE_URL` selects a different provider endpoint. Provider-prefixed
settings remain supported, with generic `LLM_*` settings taking precedence.

Applications create an Ensemble, create its Agent, and register tools on that
Agent before calling `Ask`. Registration rejects duplicate names (including
builtins), malformed JSON schemas, empty names and nil handlers. The handler
receives public `ToolContext` and JSON arguments. It can reach owned services via
`call.Engine().Agent().Ensemble()`. Register and ask synchronously; this example
does not introduce an actor or concurrent registration API. Call `Shutdown` when
the session ends. Builtin tools are optional and disabled in this small example.

`go test -race ./...` runs a faithful HTTP provider fake through the real tool
loop and checks that the custom declaration and result reach the provider.
