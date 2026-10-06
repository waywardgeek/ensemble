# Chapter 21 — Coder Brief

*For the coder (Opus 5). Read this, the chapter outline
(`book/chapter-21-outline.md`), and the research doc
(`book/chapter-21-research.md`) before writing code. Build the
reference solution and grader. Never edit `book/`.*

## What this chapter adds

Chapter 19 left the agent with three vendor renderers (Anthropic, OpenAI,
Gemini), OAuth credentials, the Responses API, and MCP tool support
(Chapter 12). Chapter 20 was prose-only (the crossover chapter). Chapter 21:

1. **`search_web` tool via ddgs MCP server.** The agent gains the ability
   to search the internet. The ddgs library is a Python metasearch
   aggregator with 10 backends (Google, Bing, Brave, DuckDuckGo, etc.).
   It ships a built-in MCP server (`ddgs mcp`, stdio transport). The agent
   starts this as a sidecar process, discovers its tools via the existing
   MCP infrastructure (Chapter 12), and registers them. No API key needed.
   Install: `pip install ddgs[mcp]`.

2. **`crawl_web` tool via crawl4ai MCP server.** The agent gains the ability
   to fetch and extract content from URLs. crawl4ai is a Playwright-backed
   Python library that handles JavaScript-rendered pages. It ships a
   built-in MCP server (stdio transport). Same pattern as ddgs: sidecar
   process, MCP discovery, tool registration. No API key needed.
   Install: `pip install crawl4ai`.

3. **MCP server configuration.** The agent needs a way to declare which MCP
   servers to start. This may be a `mcp_servers.json` in the skill
   directory, or a configuration in settings — check how the existing MCP
   server wiring works (Chapter 12's `internal/mcp/`) and extend it. The
   key requirement: the agent starts the Python sidecar processes and
   connects via stdio transport.

4. **Security boundary for retrieved content.** Web content is untrusted.
   The agent must enforce that retrieved content cannot:
   - Add or remove tools from the registry
   - Change tool permissions or capability flags
   - Disclose secrets or API keys
   - Widen the agent's tool access
   
   This is enforced at the dispatcher/registry level, not by prompting.
   The tool registry is the authority on what tools exist.

5. **Failure handling.** Search and fetch can fail (empty results, HTTP
   errors, timeouts, MCP server crashes). Every failure must produce a
   visible error the model can read, not silent swallowing. The model
   should report failure honestly, not hallucinate an answer.

## What this chapter does NOT add

- **No vendor-native search.** We do not use Anthropic's `web_search`,
  OpenAI's `web_search`, or Gemini's `google_search_retrieval`. Those are
  documented in the chapter prose for the reader's education, but the
  reference implementation uses local tools.
- **No settings UI.** Search is always on. No toggle, no dropdown, no
  settings page changes. Disabling search is deferred to a sandboxing
  chapter.
- **No citation normalization layer.** Search results are plain text tool
  results. They survive model switches and save/restore by construction.
  No opaque vendor blobs, no offset-based citations, no encrypted
  continuation data.

## Starting point

```
cp -r solutions/ch19 solutions/ch21
```

There is no ch20 in solutions (it was prose-only). The coder starts from
ch19's reference solution and adds the two MCP-based tools.

## Key architectural decisions

1. **MCP servers, not custom HTTP adapters.** Both ddgs and crawl4ai ship
   built-in MCP servers. Use them. The agent already has MCP infrastructure
   from Chapter 12 (client, bridge, tool registration). This chapter is the
   payoff for that investment — no new dispatcher logic needed.

2. **Tools are always available.** No settings gate, no capability flag.
   The MCP servers start when the agent starts. The tools are registered
   and available to the model on every turn.

3. **Results are plain text.** Search returns a list of results (title, URL,
   snippet). Crawl returns extracted markdown. Both are standard tool
   result strings. No special types, no opaque wrappers.

4. **Grader is implementation-neutral.** The grader tests behavioral
   contract, not implementation. A student who uses native vendor search,
   Firecrawl, or any other mechanism passes as long as the behavioral
   checks hold. The grader's local HTTP server provides known pages; the
   student's tools must be configurable to hit the local server instead
   of the real internet.

## Grader design

**Infrastructure:** A local HTTP server that serves controlled pages. The
grader starts this server and configures the agent to search/fetch from it
instead of the real internet. This means the search and crawl tools need
an endpoint override — an environment variable or configuration that points
them at `http://localhost:<port>` during testing.

**Behavioral checks (7 proposed):**

| Check | What to verify |
|---|---|
| Search available | Agent has a working search tool that returns results |
| Fetch available | Agent has a working fetch/crawl tool that returns page content |
| Results in conversation | Search/fetch results appear as tool results the model can read and use |
| Persistence | Search results survive save/restore with content intact |
| Untrusted content | A crafted page with malicious instructions cannot change tool permissions or grant new capabilities |
| Failure visibility | Empty results, HTTP errors, and timeouts produce visible error messages |
| No silent swallowing | A failed search/fetch does not result in the agent continuing as if it had succeeded |

**Mutation checks (4 proposed):**

| Mutant | Expected failure |
|---|---|
| Remove search tool registration | "Search available" fails |
| Remove fetch tool registration | "Fetch available" fails |
| Allow retrieved content to grant tool access | "Untrusted content" fails |
| Suppress error on fetch failure | "Failure visibility" fails |

**Fake server requirements:** The grader's HTTP server must serve:
- A page with known content (for verifying crawl extraction)
- Search-like results (the search tool's endpoint override must allow
  the grader to control what results come back)
- A page containing prompt injection instructions (for the security check)
- Error responses (404, 500, timeout) for failure path testing

## Files to read before coding

| File | Why |
|---|---|
| `book/chapter-21-outline.md` | The full chapter outline with section structure |
| `book/chapter-21-research.md` | Research on vendor APIs, ddgs, crawl4ai, pricing |
| `solutions/ch19/internal/mcp/` | The existing MCP client, bridge, and transport |
| `solutions/ch19/internal/mcp/bridge.go` | How MCP tools are bridged into the agent's registry |
| `solutions/ch19/cmd/main.go` | Where MCP servers are currently started (if any) |
| `solutions/ch19/skills/` | How skills declare MCP servers |
| `internal/grade/ch19_*.go` | The ch19 grader as a template for ch21's grader |
| `book/course-policy.md` | Grading policies P1–P9 |

## Environment requirements

The reference solution requires Python packages:
- `pip install ddgs[mcp]` — for the search MCP server
- `pip install crawl4ai` — for the crawl/fetch MCP server

The grader must NOT require internet access. All web interactions during
grading must go through the local HTTP fake server.

## Deliverables

1. **`solutions/ch21/`** — the reference solution, starting from `solutions/ch19`
2. **`internal/grade/ch21_*.go`** — the grader (7 checks, 4 mutations)
3. **`book/chapter-21-coder-review.md`** — a review for the author covering:
   - What was harder than expected and why
   - What architectural decisions you made that the outline didn't anticipate
   - Any factual errors or gaps in the outline or research doc
   - What the prose should emphasize based on implementation experience
   - What changed from the outline's proposed structure
   
   This review is critical. The author (Opus 4.6) writes the chapter prose
   next, and the implementation experience is the primary source of real
   teaching material. Don't skip it. Write it in plain English.

## What the coder must NOT do

- **Do not edit `book/`.** The author writes prose. The coder writes code.
  If you find an error in the outline or research, report it — do not fix it.
- **Do not add settings UI.** No new settings fields, no GUI changes for
  search. Always on.
- **Do not implement vendor-native search.** The chapter documents it but
  does not implement it. The reference solution uses local tools only.
- **Do not weaken existing tests.** Adding search must not break or degrade
  any existing grader check from earlier chapters.
- **Do not hardcode ddgs or crawl4ai.** The grader must accept any search
  implementation. The reference solution uses ddgs and crawl4ai, but the
  grader checks behavior, not library names.
- **Do not require internet access in the grader.** All grader checks must
  run against the local fake server.
