# Ensemble: Edition 2

This is the active application. Chapter numbers below use Edition 2 unless
explicitly labeled historical; accepted chapter snapshots live in `../solutions/`.

Build the CLI in this module:

    go build -o ensemble-cli .

Set `LLM_VENDOR` (`anthropic`, `openai`, or `gemini`), `LLM_MODEL`, and
`LLM_API_KEY`, then run `ensemble-cli chat` from your chosen workspace. Set
`LLM_WOL_MAC` only for a local provider, using its six-octet MAC address.
Ensemble checks the configured endpoint with TCP and ICMP before broadcasting a
magic packet, then waits up to 90 seconds for its TCP port. Model
names are explicit; select one supported by this application and available to
your provider account. For Anthropic, use supported older models such as
Sonnet-5; 5.5 models are excluded from current use. Messages
while the agent is busy are hints; `/interrupt` ends the current turn while
other jobs continue. Use `/model NAME` between turns to select another model
on the same provider and billing route. The optional GUI exposes the same
session control in Settings. Launch configuration wins on restart. Ask for
`agent_status` to inspect the current model, last response, session tokens,
cache hit rate and per-model cost; unavailable prices are reported as unknown.
EOF waits for the current turn and shuts down the Agent.

Known models stream by default. `EN_DISABLE_STREAMING=1` requests one-piece
responses. Unknown models remain usable for text with streaming off. Replies
go to stdout; reasoning and tool arguments go to stderr. Redirected streams
contain no ANSI formatting. The CLI writes wire frames to `api.log` and facts
to `ch02.jsonl` (override the latter with `CH07_AGENT_LOG`). Authentication
headers are excluded and the configured key is redacted from wire traces.

The no-argument CLI accepts the existing JSON-lines exercise protocol. The
separate [Chapter 6 application](ch07/README.md) demonstrates the public library
and reports delta counts, finalized parts and time to first delta.

Chapter 7's optional browser application lives in the separate `gui/` module;
see `gui/README.md`. The existing headless command remains independent of HTTP,
WebSocket and browser dependencies. Both applications use the same public Agent
API; the GUI executable also accepts terminal input for its existing Agent.

Chapter 9 introduces `EN_SKILLS_DIR` and `EN_PRIMARY_SKILL` in both clients.
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
pending instruction removal. Its tool grants stop immediately; shared grants
from another loaded skill remain available.
Without a skill specification the earlier unrestricted registration API remains
available. A skill can enable only installed handlers; declaring an unknown tool
never creates an executable implementation.

Chapter 11 introduces MCP peers. `ensemble-cli --mcp-pipe` dedicates stdin/stdout to
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
are admitted. Unloading stops its MCP servers and tool grants immediately,
while instructions retain the pending-removal behavior. Subprocesses get
basic OS environment values, not the model client's provider keys.

For an external transport, embedders use `Agent.MCP().Connect`, with
`ensemble.NewMCPTransport` for streams or `ensemble.MCPPipe` for an in-process
peer. The optional GUI supplies its own WebSocket transport. Remote calls use
real jobs: callback delays return a handle, explicit kill cancels the RPC, and
Agent shutdown closes peers and joins its work. Discovered `round`/`turn`
ephemeral tools run automatically and remain absent from model declarations.

Chapter 12 supplies `skills/ensemble` and `skills/gui-debug`. Both commands
accept `--skills-dir PATH` (overriding `EN_SKILLS_DIR`) and `--gui-debug` to load
that optional skill before accepting input. With `--gui-debug`, the directory
otherwise defaults to `./skills`; selecting a directory defaults the primary
name to `ensemble` unless `EN_PRIMARY_SKILL` explicitly selects another one.
For the browser, build the separate GUI command. Its assets are embedded; this
skill-enabled example uses paths relative to this directory:

    ensemble-gui --gui-debug --skills-dir ./skills

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

Memory uses the owning Agent's data directory: `SOUL.md`, `MEMORY.md`, session
files in `memory/`, and graduated files in `bucket-0/` and `bucket-1/`.
The optional GUI's Context Management tab controls each band. Budgets and
watermarks are bytes; zero selects defaults (conversation 3,000, bands 1,500,
high twice the budget, low the budget). Disabling a band preserves its files;
re-enabling rereads edits. Graduated source files remain beside the originals
with a `.retired` suffix and are not active memories.

`micro_handoff` checkpoints completed work and is the only compression trigger.
At the completed-batch boundary it clears tool calls/results, opaque reasoning and
private replay signatures, including those attached to visible text. Visible
dialogue and the event log remain.
The compressor makes one bounded request using the current provider and one
`submit` tool. Failed or interrupted calls preserve the old source. Recorded
memory bytes and cuts replay without another model call. A compression target
is a prompt guideline, not a guarantee that the model preserves every fact;
review important memories on disk and correct them when necessary.

Auto-recall searches the Agent's memory files (including retired originals),
`docs/`, `handoffs/`, `learnings/` markdown, and its configured skill directory.
These directories are relative to `Config.DataDir` (`cr/` in the commands).
It skips text already present in context, including the newest memories when
already loaded by a band. Markdown learnings are the supported public corpus
format; this application does not read another application's private JSON files.

Each user turn of at least 80 characters searches per-source BM25 indexes. The
stateless judge defaults to the current provider/model; set
`EN_RECALL_JUDGE_MODEL` to a cheaper compatible model on that same provider.
`EN_RECALL_DISABLE_JUDGE=1` selects mechanical recall, and
`EN_RECALL_DISABLED=1` bypasses recall. Embedders can set `Config.Recall` limits;
zero selects the chapter defaults (3 snippets, 20 candidates, 3 recent messages,
6,144 attachment bytes). Judge failure falls back to final source quotas:
memories receive half the slots rounded up, and unused slots are redistributed.
A successful judge's selections, including none, stand unchanged.

Recall is recorded under `recall_attached.parts`, separately from the user's
words. Renderers retain its position; offline replay never searches or judges.
A `micro_handoff` removes recalled context at the existing checkpoint boundary.
Each attachment also retains hashes of the source snippets it represents, so a
byte-capped snippet is not injected repeatedly. These identities leave with the
attachment; edited source content can be recalled again.

Chapter 16 introduces subscription authentication to both executables:

    ensemble-cli auth login openai-chatgpt
    ensemble-cli auth status
    ensemble-cli auth models
    ensemble-cli auth logout

A successful login explicitly selects `chatgpt` billing for later OpenAI
launches, including when an API key is also configured. Expiry, sign-out and
quota errors never switch to that key. Set `OPENAI_BILLING_MODE=api` to explicitly
select metered billing or `chatgpt` to require the subscription. Automatic
`chatgpt_then_api` fallback is deferred. Other providers continue using their own
API keys. Identity-only login is retained and displayed, but inference requires
the actual `chatgpt.tokens.use.direct` grant.

Select an account-available model with `LLM_MODEL`. `gpt-6-astra` and `gpt-6.1-sol`
use Responses; the other tested OpenAI rows retain Chat Completions. For Bill's
crossover configuration, set `LLM_VENDOR=openai`, `LLM_MODEL=gpt-6-astra`,
`OPENAI_REASONING_EFFORT=high` and `OPENAI_BILLING_MODE=chatgpt`.
Subscription inference requires a supported Responses row.
`OPENAI_REASONING_EFFORT` configures effort and
`OPENAI_REASONING_SUMMARY` configures summary detail (`auto` by default). Responses
always streams internally because terminal completion establishes success.
Public summaries are distinct from assistant text and private reasoning bytes.

Credentials live in the platform's user configuration directory under `ensemble`,
or the explicit `EN_HOME` directory. The public OAuth flow never borrows Codex or
CodeRhapsody profiles. The optional Chapter 24 route below explicitly selects
a read-only Codex credential file instead. Atomic owner-only files and a cross-process lock protect
rotating refresh tokens. `auth login --print-url` exposes a browser handoff for
headless use; optional identity-token hints are omitted so it cannot print a
retained token. `EN_BROWSER_CMD` selects a browser executable and arguments;
it is executed without a shell. `OPENAI_OIDC_ISSUER` and `LLM_BASE_URL` support
isolated external peers for protocol tests. Production OAuth uses OpenAI's public
issuer and `/v1` API endpoints.

The observed ChatGPT route rejects `prompt_cache_options`, so subscription
requests omit that option and its paired explicit markers. API-key Responses
requests retain explicit caching on supported models. Cache diagnostics remain
active on both routes; an answered request does not establish a cache grant.

Chapter 18 introduces Streamable HTTP MCP peers with `transport: url`, `url` and
`auth-env` fields. The named environment variable is resolved at skill load as
a Bearer token; an absent value logs an anonymous-connection diagnostic. JSON
and SSE POST replies, optional session headers and job cancellation use the
same MCP manager and dispatch path. Non-2xx response bodies retain their reason,
with the configured credential redacted if a server reflects it.

Run `ensemble-cli chat --skills-dir /absolute/path/to/skills` and ask to load
`web-search`. The shipped primary permits that skill, which declares only
`firecrawl_search` and `firecrawl_scrape` at `https://mcp.firecrawl.dev/v2/mcp`.
Set `FIRECRAWL_API_KEY` in the launching environment; never paste it into chat
or a skill. Discovery does not grant additional tools. Unload removes the web
capabilities immediately. Search/page text is untrusted evidence; check source
URLs and scrape metadata before citing it. Provider tools, schemas, access and
quotas can change; this is not a promise of URL-only provider interchangeability.

Chapter 19 introduces per-Agent confinement. Run `ensemble-cli chat --sandbox /absolute/work`
(and optionally `--safe-mode` or `--no-web-search`). File-tool relative paths use
that root. Commands and descendants use macOS Seatbelt with no network; ordinary
system tools remain readable so Python and Git work. Other platforms refuse
confined commands. File tools use Go's rooted filesystem operations, including
state-file writes, so checking a path and later following an escaping symlink is
not an allowed shortcut. CLI saves, history, trace and Agent data live under the
selected root. Library callers must choose data/log/save paths inside their root.

Safe mode removes `run_command`, `send_input`, `wait_for_job`, `kill_job`, and
`tool_limits`, while ordinary file tools remain usable. Disabling web search
withholds its skill and MCP connection. An omitted library AgentSpec retains the
prior full-access defaults; an explicit spec's `EnableWebSearch` must be true to
permit that skill. `AgentSpec.Child` derives restrictions, and `Clamp` returns an
unconditionally narrowed spec plus a diagnostic for attempted widening. The fleet
derives these restrictions when spawning children. The binary's credential provider
and model HTTP client remain outside the command sandbox. The tested boundary is
file tools, retained state and macOS command descendants, not a claim of immunity
to arbitrary compromised native code or all forms of prompt injection.


Chapter 20 makes browser assets reusable and deployment independent of the source
tree. Build with `cd gui && go build -o /tmp/ensemble-gui ./cmd`; the binary serves
its embedded GUI from any working directory. `--gui-dir /absolute/path/to/web/gui`
explicitly selects editable disk assets during development.

External applications import the optional `ensemble-gui-stub` module and the
core's `ensemble/ensemble` public package. Construct the actual Ensemble and Agent,
then `gui.New(root, agent, log)`. Public `Event`, `MessageData`, `Part` and `Human`
already allow a consumer to plant history before attaching the GUI. `Handler("")`
provides embedded HTTP assets and the unchanged `/ws` protocol. The separate GUI
module keeps WebSocket dependencies outside both headless executables and MCP.

`gui.Assets()` returns a read-only `fs.FS` rooted at the browser filenames. Mount
only `artifact-scroll.js` and `renderers.js` to reuse `new ArtifactScroll(element)`
without serving the application shell. Feed it the existing display messages via
`receive(message)` and call `reset()` before a fresh replay. A speech service is
optional; importing the component neither queries application controls nor opens
a socket. The full application passes its existing speech owner explicitly.

`server.ServeMCP(listener)` exposes a JSONL browser relay on a caller-supplied
listener, normally loopback. It preserves request IDs, including concurrent
clients reusing an ID, independently of the Agent's own MCP session. With no
browser it returns a correlated error. A connected but silent browser waits until
that browser, the TCP client or the Server closes; no timeout is promised. The
application chooses access to this listener; `Server.Close` closes its listeners
and clients. The externally built test in `gui/testdata/reuse` demonstrates these
public boundaries with real core objects and fake external peers.

Chapter 20 introduces a recursive fleet in the CLI. Ask it to delegate work with
`spawn_sub_agent`, inspect `agents_status`, collect `join_agents`, and finish
with `submit`. Each child owns a separate actor and `cr/children/ID` state
directory; shutdown preserves its save for a later child with the same ID.
Children share the application's working directory for task files. The human
is an addressable recipient, not a managed child.

A child asks its parent with `send_message_to_parent`; the parent answers with
`respond_to_parent_message`. A root question appears on stderr. Enter
`/reply ANSWER`, or `/reply MESSAGE_ID ANSWER` when several questions are pending.
Ordinary text entered during a tool wait steers the next request and releases
that wait while the actual job continues; its handle still retrieves the result.
Three automatic reminders follow a successful turn without submission, then
the fleet reports escalation. Failed turns are reported without automatic retry.

Library applications enable this protocol by installing `SetAgentFactory`
before starting Agents and construct every child with `NewChildAgent(parent,
cfg, spec)`. Configure `SetUserSink` to receive human questions and escalation,
and answer through the root's `Fleet().Reply`. Without a factory, the prior
single-Agent turn protocol remains and fleet tools are hidden. The optional
GUI also provides the Chapter 21 fleet console described below.

## Fleet console (Chapter 21)

The shipped CLI and optional GUI call `Ensemble.Start` to restore named peers
from `ensemble.json`. The GUI's Agents sidebar creates peers with a name, model
and defining skill, selects their conversations, and deletes completed peers.
Peers retain their own data under `agents/NAME`; task children remain scoped to
the spawning Agent and are not registered as persistent peers. Tree keys follow
actual parent paths, so repeated names in different branches stay distinct.

Activity (thinking, tool call, waiting, idle) is separate from lifecycle state.
The fleet tools include scoped status, change/output waiting, live messages,
synchronous questions and durable submissions. `shutdown_agent` waits for a
child's current/queued turns; `kill_agent` cancels its lifetime immediately.
`interrupt_agent` ends only a turn (`now` or `after_tool`), leaving the Agent
and independent jobs available. Output cursors address finalized history events;
streamed partials are a GUI display concern.

Build the optional browser app in `gui` with `go build -o /tmp/ensemble-gui ./cmd`.
Run the browser regression with `GUI_BINARY=/tmp/ensemble-gui node fleet.test.mjs`
from that module. It uses a local provider fake and real Chrome, not paid inference.

## Optional Chapter 24: Codex endpoint

Use this temporary route only if the public OAuth caching problem affects your
work. It has no separate chapter grader. Sign in with Codex, then launch either
the CLI or GUI with:

```sh
export LLM_VENDOR=openai
export LLM_MODEL=gpt-6-astra
export OPENAI_REASONING_EFFORT=high
export OPENAI_BILLING_MODE=codex
unset LLM_BASE_URL OPENAI_BASE_URL
./ensemble-cli chat
```

The selected file is `EN_CODEX_AUTH_FILE`, else `CODEX_HOME/auth.json`, else
`~/.codex/auth.json`. File-backed ChatGPT sign-in must provide
`tokens.access_token` and `tokens.account_id`; keychain-only or API-key Codex
sign-ins are not read by this adapter. Embedders set `Config.BillingMode` to
`CredentialCodex` and explicitly supply `Config.CodexAuthFile`.

The reader obtains both fields together on each request. Codex owns login and
renewal; Ensemble does not change that credential file. On rejection, renew the
sign-in in Codex and retry. No API-key fallback or authentication retry occurs.
The route is `https://chatgpt.com/backend-api/codex/responses`; redirects and
stale public endpoint overrides are rejected. Loopback endpoints support tests.
The existing Responses tool/summary/replay pipeline and measured usage remain
in use, with automatic caching and a stable Agent/model cache key, also sent as
the Codex `session-id` cache-affinity header.

To return to Ensemble's public OAuth registration, set
`OPENAI_BILLING_MODE=chatgpt` and leave the base URL unset. `auth login`, `status`,
`models` and `logout` continue to manage Ensemble's own public OAuth profile;
they do not manage Codex's sign-in. Select `api` only for intentional metered use.
