# Edition 3 — Chapter 6

This module implements the current Chapter 6, retaining Chapters 1–5 contracts.
It uses only the Go standard library; there is no provider SDK or framework.

## Chapter 6 — streaming

Engine has one `Parse(resp, callbacks)` path for JSON and SSE, using a shared SSE
reader and provider-local assembly. Deltas carry response-local part IDs (zero
is valid); complete parts retain native boundaries, usage, and source identity.
Unknown streaming capabilities use JSON; `DisableStreaming` and
`EN_DISABLE_STREAMING=1` opt out. The Agent alone records History and saves before
publishing finals. Partial calls cannot execute. Jobs reserve their stable handle,
then the Agent records/saves the actual call before releasing the handler.

The CLI drains a lossless observation queue: answer text goes to stdout, thinking
and tool progress to stderr, with no duplicate final answer. JSON `prompt` records
can receive observation records; legacy `user` records retain their plain-mode
contract. Root-owned `api.log` tracing omits headers/authenticated URLs and redacts
configured keys. `ch07/` is a separate public-library streaming shape example.
Ephemera updates replace the pending projection (empty clears it), and the next
`RequestSent` consumes it even when the HTTP request fails.

## Ownership rationale

The public `ensemble` package is a small composition root. It owns a logger and
its Agents, constructing both directions of the Agent/Engine relationship before
returning an Agent. `internal/common` declares shared values and interfaces only;
`internal/agent` and `internal/engine` never import each other.

Agent holds configuration and completed history. Engine holds the HTTP client
and measured usage, looking up configuration and diagnostics through its actual
Agent parent (and that parent's Ensemble). A failed request never commits an
unanswered user turn. Missing usage is not zero; measured usage from an empty
answer is still counted. History and usage accessors return copies, not mutable
internal state. Each Agent serializes conversation state in its actor; HTTP and
job work run outside that actor. Separate Agents have independent conversation and accounting.

The HTTP client has no total response or header deadline. Standard transport
connection/TLS protections remain, while the caller's context governs model
work, including reading the body. Redirects are rejected to avoid forwarding
credentials. Errors omit remote bodies, URLs, and raw transport diagnostics.

## Use

Set `ANTHROPIC_API_KEY` and `ANTHROPIC_MODEL`; discover the model with the
provider's `GET /v1/models` endpoint rather than guessing. `ANTHROPIC_BASE_URL`
defaults to `https://api.anthropic.com`, and can select a local HTTP fake.

```sh
go build -o /tmp/ch01 ./cmd/ch01
printf '%s\n' '{"user":"Hello"}' | /tmp/ch01
/tmp/ch01 chat
```

No arguments selects JSON-lines input/output: one `assistant` object per input,
then cumulative `usage` at EOF. Errors go to stderr and exit nonzero without
retrying. `chat` prints human prompts and per-turn usage; Ctrl-D ends it.
The root accepts an application logger (`ensemble.New(logger)`), then
`app.NewAgent(ensemble.Config{...})` creates a conversation for synchronous
`Ask(ctx, text)` calls. A nil logger discards diagnostics; no global logger is used.

## Checks

```sh
go test -race ./...
go vet ./...
```

Tests use real local HTTP servers with fake credentials. They cover wire shape,
parent/configuration/logger isolation, replay and failed-turn retention, usage,
sanitized errors, caller cancellation, redirects, and the CLI's line/EOF rules.
JSON input has no arbitrary physical-line size limit; a final unterminated line
is processed at EOF. The default JSON mode does not print chat prompts.


## Retained tool contracts

The same Agent owns an independent tool registry. Root composition wires
Agent, Engine, History, and Tools through `internal/common` interfaces; each
implementation keeps its state locally. Engine owns HTTP, Render, Parse, and all
provider tool envelopes. Tools do not retain conversation, and the GUI continues
to call the same Agent rather than providing a second implementation.

`Config.Tools == nil` enables the default catalogue, including fleet controls.
A nonempty slice selects names. A non-nil empty slice with no skill selection
is an application-only composition: it disables every tool, omits `tools` from
provider requests, and leaves the optional fleet catalogue uninstalled. This
allows an application such as the speech listener to register its own `submit`
operation and schema through the ordinary registry. Such a call is not fleet
completion. Installed core, compiled, and MCP names remain collision-protected;
this is not permission to override a disabled tool. Default and skill-selected
composition retain the fleet catalogue. CLI/GUI environment equivalents are `LLM_TOOLS=default`,
`LLM_TOOLS=none`, or a comma-separated list of built-in names. Registry declarations
and configuration are returned as copies, not mutable backing storage. Each
registry entry joins its declaration with its handler; declarations and dispatch
use that same per-Agent registered set.

The selected model's explanation policy is explicit:
`Config.RequireToolExplanation` / `LLM_REQUIRE_TOOL_EXPLANATION=true`. When enabled,
a whole batch lacking nonblank ordinary assistant text before its first call is
refused before effects. Each call gets a matching error result instructing a retry
with an explanation. Hidden reasoning, opaque blocks, stop reasons, and usage do
not satisfy this rule. Turn completion is based on returned calls. The default limit is 200 dispatched
batches (`Config.MaxToolBatches`); refused/invalid batches do not spend that budget.
A final request consumes the last dispatched batch's results before the limit
rejects any additional effects.

Tools use strict JSON argument types and reject unknown fields. Their schemas
are sent with each enabled declaration. `read_file` has optional one-based
inclusive `start_line`/`end_line` and a `max_bytes` cap. Omitted or zero start
means line 1, end means EOF, and `max_bytes` means no read cap. Negative bounds
or caps and an end before the start are errors. The tool selects the requested
line range first, then caps returned bytes; actual truncation adds an explicit
`[truncated at max_bytes=N]` marker outside the cap. Exact-size results do not
claim truncation. Reads use bounded fragments, including for very long lines.
`list_directory` is sorted and nonrecursive. `search_files` uses Go regular
expressions, walks regular text files without following symlinks, defaults
`context_lines` to zero, and merges overlapping/adjacent context ranges. Match
lines use `path:line:text`; context lines use `path-line-text`. Positive-context
groups have `--` separators, including across file boundaries; zero-context
searches emit only matching lines without separators.

`write_file` creates parents and refuses to replace existing files by default;
the refusal reports the current size and names the opt-in flags.
`overwrite:true` opts into replacement, `append:true` preserves existing bytes,
and using both is rejected. `edit_file` replaces one nonempty, exact, unique
`old_text` anchor; zero or multiple matches (including overlaps) leave the file
unchanged. Both refusal diagnostics identify the supplied path. Zero-match
refusals explicitly report 0 matches and suggest rereading
the file. Ambiguity refusals report the total count (including overlaps), list
up to eight one-based match-start line numbers, and say how many locations were
omitted if the list is truncated. Repeated line numbers represent distinct
matches on the same line; multiline anchors report their starting lines. No
file excerpts, fuzzy matching, or automatic match selection are used: reread
the current file and give a larger exact anchor on ambiguity. A unique anchor
can be deleted by supplying an empty `new_text`. `run_command` invokes `/bin/sh -c` and returns
JSON text with separate `stdout`, `stderr`, and `exit_code` fields. A nonzero
process exit is still a completed tool result; startup failure is a tool error.
It adds no timeout beyond caller context cancellation. The shell environment
removes `LLM_API_KEY`, `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, and `GEMINI_API_KEY`,
while retaining ordinary build variables. This removes those four credentials;
it is not a general secret detector or sandbox.

Relative paths and command `cwd` resolve against the registry's launch directory;
absolute paths and `..` are allowed. This is not a sandbox. No command changes the
process-wide directory. Ordinary calls return stable job handles with observed
output; callback limits do not kill jobs. Explicit kill and joined shutdown retain
their distinct terminal semantics. These tools are for trusted local use.
