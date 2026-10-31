Build the CLI in this module:

    go build -o ensemble-cli .

Set `LLM_VENDOR` (`anthropic`, `openai`, or `gemini`), `LLM_MODEL`, and
`LLM_API_KEY`, then run `ensemble-cli chat` from your chosen workspace. Model
names are explicit; select one available to your provider account. Messages
while the agent is busy are hints; `/interrupt` ends the current turn while
other jobs continue. EOF waits for the current turn and shuts down the Agent.

Known models stream by default. `EN_DISABLE_STREAMING=1` requests one-piece
responses. Unknown models remain usable for text with streaming off. Replies
go to stdout; reasoning and tool arguments go to stderr. Redirected streams
contain no ANSI formatting. The CLI writes wire frames to `api.log` and facts
to `ch02.jsonl` (override the latter with `CH07_AGENT_LOG`). Authentication
headers are excluded and the configured key is redacted from wire traces.

The no-argument CLI accepts the existing JSON-lines exercise protocol. The
separate [Chapter 7 application](ch07/README.md) demonstrates the public library
and reports delta counts, finalized parts and time to first delta.

Chapter 8's optional browser application lives in the separate `gui/` module;
see `gui/README.md`. The existing headless command remains independent of HTTP,
WebSocket and browser dependencies. Both applications use the same public Agent
API; the GUI executable also accepts terminal input for its existing Agent.

Chapter 10 accepts `EN_SKILLS_DIR` and `EN_PRIMARY_SKILL` in both clients.
Create one directory per skill with a `SKILL.md` containing the chapter's flat
frontmatter and Markdown body. For example, `base/SKILL.md` can contain:

```markdown
---
name: base
type: primary
tools: read_file
loadable-skills: code-tools
---
You are a coding assistant. Tools: $TOOLS. Available skills: $SKILLS.
```

A `code-tools/SKILL.md` with `type: loadable` and `tools: write_file edit_file`
can then be loaded through `load_skill` with `{"name":"code-tools"}`.
Run `EN_SKILLS_DIR=./skills EN_PRIMARY_SKILL=base ensemble-cli prompt "load the
code-tools skill"` for a single answer, or use `chat` for human conversation.
`EN_CUSTOM_VAR` supplies the example `$CUSTOM_VAR` application variable.

Embedders pass an optional `ensemble.AgentSpec` as the second argument to
`NewAgent`. Its `Variables` renderers run before the primary prompt is frozen;
`Agent.RegisterVar` affects subsequent loads. Dynamic bodies arrive through
ordinary tool results. `unload_skill` takes the same `name` argument and marks
pending removal; compaction and actual removal are not implemented in Chapter 10.
Without a skill specification the earlier unrestricted registration API remains
available. A skill can enable only installed handlers; declaring an unknown tool
never creates an executable implementation.

Chapter 12 adds MCP peers. `ensemble-cli --mcp-pipe` dedicates stdin/stdout to
newline JSON-RPC 2.0; it initializes and discovers the peer, accepts reverse
`tools/call`, and exits on disconnect. It does not start an implicit model turn.
The exercise uses protocol version `2024-11-05`, including its `$/cancelRequest`
notification. This is a historical exercise implementation, not a claim of
compatibility with every current MCP server.

A skill can add the following frontmatter, alongside an explicit `tools:` grant:

```yaml
tools: search_code
mcp_servers:
  - name: code-search
    transport: stdio
    command: python3
    args: ["scripts/search_server.py"]
```

Arguments use a JSON-style string array and paths resolve from the Agent's
workspace. Loading starts/discovers the server; only tools named by that skill
are admitted. Unloading stops its MCP servers immediately, while compiled tools
and instructions retain Chapter 10's pending-removal behavior. Subprocesses get
basic OS environment values, not the model client's provider keys.

For an external transport, embedders use `Agent.MCP().Connect`, with
`ensemble.NewMCPTransport` for streams or `ensemble.MCPPipe` for an in-process
peer. The optional GUI supplies its own WebSocket transport. Remote calls use
real jobs: callback delays return a handle, explicit kill cancels the RPC, and
Agent shutdown closes peers and joins its work. Discovered `round`/`turn`
ephemeral tools run automatically and remain absent from model declarations.

Chapter 13 supplies `skills/ensemble` and `skills/gui-debug`. Both commands
accept `--skills-dir PATH` (overriding `EN_SKILLS_DIR`) and `--gui-debug` to load
that optional skill before accepting input. With `--gui-debug`, the directory
otherwise defaults to `./skills`; selecting a directory defaults the primary
name to `ensemble` unless `EN_PRIMARY_SKILL` explicitly selects another one.
For the browser, build the separate GUI command and launch from this directory:

    ensemble-gui --gui-dir gui/web/gui --gui-debug --skills-dir ./skills

Open its browser before sending a prompt. The supplied skill declares a
WebSocket peer, so connection/discovery waits for the browser. An application
can instead supply an explicitly typed `gui-debug` skill with a stdio peer,
using the same four explicit tool grants. `type: loadable` remains required;
the original Chapter 13 sample omits it.

Embedders can select `AgentSpec.StartupSkills` alongside the primary. These use
the existing loadable-skill admission and MCP lifecycle. Their returned manuals
join the frozen startup prompt because no model-authored tool call exists yet;
ordinary `load_skill` calls still return instructions into conversation history.
Startup grants have dynamic provenance with sequence zero (before history).
Unloading immediately disconnects MCP tools and peers while instruction removal
remains pending. Programmatic browser input updates the control without taking
the human typing pause; actual typing and explicit Agent-wide unpause still work.

At the completed tool-batch boundary, `micro_handoff` also removes opaque
reasoning and private replay signatures from retained context. Visible dialogue
and the checkpoint note survive; the original event log keeps the provider facts.
