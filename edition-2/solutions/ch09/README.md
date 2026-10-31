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
