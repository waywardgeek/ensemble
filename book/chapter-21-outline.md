# Chapter 21: Giving the Agent the Web

*Revised outline (October 6, 2026). Scope: two local tools (`search_web` via ddgs, `crawl_web` via crawl4ai) wired as MCP servers. Always on, no settings UI, no vendor-native search. Grader accepts any search implementation that meets the behavioral contract. No implementation or measured results are claimed here.*

## Through-line stake

The engineer needs current documentation to ship a correct change, but letting the agent read arbitrary web pages also lets strangers put instructions in its context; the chapter resolves with a sourced answer and a tested boundary around what retrieved content may cause the agent to do.

## Voice plan and evidence

Third-person mechanism throughout the body. The opener uses Bill's first-person voice — a direct editorial about why the AI vendors failed the community on search, and why the chapter routes around them with open-source tools. This is the sharpest editorial in the book, earned by 20 chapters of vendor-neutral engineering. Each section below names its concrete teaching fact or proposed exhibit.

---

## 21.1 Why this chapter is next

*First-person opener. Bill's voice.*

The order of upcoming chapters is motivated to minimize switching back to the old agent. While waiting for OpenAI to fix their caching bug, Bill is still using CodeRhapsody for most work. When he does use Ensemble, the main reason he switches back is that Ensemble can't search the web. That ends this chapter.

**The vendor indictment.** When AI companies say they're here to better humanity, translate that to: they're here to make money. Actions speak louder than words, and search is a clear case of choosing the path of profit over the path of good.

If bettering humanity were the goal, vendors would provide MCP servers at a reasonable cost — say $0.01 per search and per web page fetch. Done. Reusable in any agent, not just coding agents. The impact on security for the world would be hard to overstate. Prompt injection is a hard problem, and a company that aims to do good would work on solving it for the world, putting humanity first.

What did all three AI vendors do? Every single one went down the dark path. Each implemented search as a custom, incompatible, proprietary tool. Some don't return separate tool results — they fold search results into the LLM output, charging the maximum possible price per token. One provides search but explicitly punts on prompt injection, leaving the developer to figure out web scraping on their own. They made search and scrape purposely incompatible between vendors so that once you build the infrastructure for their bespoke system, you'll have little incentive to switch. This is vendor lock-in marketing in its purest form.

For this specific AI coding agent, the LLM vendors are to blame for making their native search unusable as a portable solution. The community has responded with open-source alternatives, and this chapter goes with the community.

---

## TL;DR

Add web search and URL fetching to the existing agent using two local tools wired as MCP servers.

- **`search_web`** queries the internet via ddgs (a metasearch library with 10 backends including Google, Bing, Brave, and DuckDuckGo). No API key required.
- **`crawl_web`** fetches and extracts content from a URL via crawl4ai (Playwright-backed, handles JavaScript-rendered pages). No API key required.
- Both tools are MCP servers (Chapter 12 infrastructure). The agent discovers their tools at startup and calls them like any other tool. The model decides when to search.
- Results are plain text tool results in the conversation. They survive model switches, save/restore, and vendor changes by construction — no normalization layer needed.
- Retrieved content is untrusted evidence. It does not grant permission to run commands, disclose secrets, change settings, or widen tool access.
- Failure paths (HTTP errors, timeouts, empty results) produce visible errors, not silent swallowing.
- The grader accepts any search implementation that meets the behavioral contract. Readers who prefer native vendor search, Firecrawl, or another approach may use it. This chapter's recommendation is our answer today — the reader's needs may differ.

---

## 21.2 The idea in plain words

**Thesis:** A coding agent needs a way to check documentation that changed after its training data was collected.

Follow one task through the chapter: verify a current API feature before implementing against it. Show the query, the retrieved source, and the claim the source supports. Search discovers sources; fetching reads a particular source; browser automation interacts with a page. This chapter implements search and fetch. General browser automation remains separate.

**Proposed exhibit:** An API question answered from a dated primary source, with the supporting passage visible beside the answer. Choose and record the task during implementation; do not invent a stale-model failure.

---

## 21.3 The state of search — October 2026

*Third-person. Research summary with enough detail for the reader to choose their own path.*

**Thesis:** The research was done in October 2026. It confirmed what was already known a year earlier: all three major LLM vendors implemented web search as proprietary, incompatible tools designed for lock-in. The state has not changed. Maybe the vendors will have a change of heart. Maybe not.

Present each vendor's offering with enough detail for the reader to understand how it works and decide for themselves:

**Anthropic (Messages API):** `web_search` is a tool declaration in `tools[]`. The model calls it; Anthropic's servers execute the search; results return as `web_search_tool_result` content blocks with citations. Also offers `web_fetch` for URL content retrieval (up to 1MB, respects `robots.txt`). Priced per search invocation, separate from token costs. `max_uses` caps searches per request. Citation format: `cited_text` (≤150 chars) with URL. Continuation requires preserving `encrypted_content` blocks unchanged — modifying them produces a 400 error. This is the most complete offering of the three.

**OpenAI (Responses API):** `web_search` is a hosted tool. Results inline with `url_citation` annotations containing text offsets (`start_index`/`end_index`). No separate URL fetch tool — OpenAI has nothing equivalent to Anthropic's `web_fetch` or Gemini's `url_context`. Priced as token usage (no separate search fee). `search_context_size` (`low`/`medium`/`high`) controls returned context volume. Chat Completions search models (e.g., `gpt-5-search-api`) always search before responding — "off" is not available for these models.

**Google/Gemini (Generation API):** `google_search_retrieval` provides search grounding as part of generation. Results are woven into the model output rather than returned as separate tool results — the vendor charges full output token pricing for search-sourced content. Also offers `url_context` for URL content retrieval. Search Suggestions display is governed by the Gemini API Terms of Service. Billed per grounding query on Gemini 3+ (per query, not per prompt).

**The common pattern:** All three are flags on the existing API request. No second HTTP call, no second model. But: all three produce vendor-specific result shapes. Citation formats are incompatible (text offsets vs. copied text vs. inline grounding). Continuation rules differ (Anthropic's encrypted blocks, Gemini's terms of service, OpenAI's offset fragility). A normalization layer would be required to make any of them portable — at which point the native integration advantage vanishes.

**The alternative: Firecrawl.** Not every commercial offering chose the dark path. Firecrawl provides a clean API for web scraping and search, optimized for AI applications, with managed hosting and a self-hosted option. It does not lock you into any LLM vendor. It is a legitimate choice for readers who prefer a managed service over self-hosting.

**The community answer.** Two open-source libraries cover the need without vendor lock-in:

- **ddgs** (MIT license, v9.16.0): Formerly a DuckDuckGo scraper, now a metasearch library aggregating 10 backends (Google, Bing, Brave, DuckDuckGo, Mojeek, Startpage, Yandex, Yahoo, Wikipedia, Grokipedia). Built-in MCP server. No API keys.
- **crawl4ai** (Apache 2.0, 84.8K stars): Playwright-backed web content extraction. Handles JavaScript-rendered pages. Built-in MCP server. No API keys.

This chapter uses `ddgs` and `crawl4ai`. The grader tests behavior, not implementation — any search mechanism that meets the contract will pass.

---

## 21.4 Two tools, zero vendor lock-in

**Thesis:** Search and fetch are MCP tools. The infrastructure from Chapter 12 already handles them.

This is the payoff for Chapter 12's MCP investment. Both `ddgs` and `crawl4ai` ship built-in MCP servers using stdio transport. The agent spawns them as sidecar processes at startup, discovers their tools, and registers them in the tool registry. The model calls `search_web` or `crawl_web` like any other tool. The MCP bridge dispatches the call, waits for the response, and returns the result as a tool result message.

**Teaching fact:** No new dispatcher logic. No vendor-specific renderer additions. No vendor-specific parser additions. The agent already knows how to call MCP tools. The only new code is the MCP server configuration (which servers to start, which tools to expose) and the security boundary around retrieved content.

Show the tool registration, an example search call and result, and an example crawl call and result. The results are plain text — a list of search hits (title, URL, snippet) for search, and extracted markdown for crawl. They survive model switches because they're just text in the conversation.

**Proposed exhibit:** Side-by-side request traces: (1) search tool call dispatched through MCP, results returned as a tool result; (2) crawl tool call for a URL found in the search results, content returned as markdown.

---

## 21.5 Fit search into the existing loop

**Thesis:** MCP tools already participate in the tool loop. The interesting work is at the edges.

Walk the path: model emits a `search_web` tool call → MCP bridge dispatches to the ddgs server → ddgs queries the web → results return through the bridge → tool result recorded in the event log → observer/GUI display the results → model reads the results and continues.

This is the same path as any MCP tool from Chapter 12. What's new:

1. **Tool results can be large.** A crawled page can be tens of thousands of tokens. The existing context management (Chapter 15) handles this — large tool results are candidates for redaction on subsequent turns. But the chapter should acknowledge the pressure search results put on the context window.

2. **Tool results contain untrusted content.** This is the subject of §21.6 and the chapter's hardest problem. The dispatcher doesn't need to change, but the agent's trust model does.

3. **Save and restore.** Search results are tool results. They already survive save/restore through the existing persistence mechanism (Chapter 11). No special handling needed. Show it working.

4. **Model switching.** Search results are plain text. They survive model switches by construction. Show a search result produced under one vendor rendered correctly under another. This is the advantage of local tools over native vendor search — no normalization layer, no opaque blobs, no vendor-specific citation formats to translate.

**Proposed exhibit:** A save/restore cycle followed by a model switch, with the search results visible and usable in both contexts.

---

## 21.6 A web page cannot authorize a command

**Thesis:** Web text can contain useful documentation and malicious instructions in the same paragraph. The agent needs an enforceable boundary.

Use a controlled page containing legitimate API documentation plus an instruction to disclose a test secret. The content arrives as a tool result from `crawl_web`. HTML cleanup (crawl4ai's markdown extraction) may remove scripts and navigation while leaving that instruction intact. Clean content is not safe content.

Teach the layers of defense, in order of enforceability:

1. **Tool authorization (enforceable).** The tool registry determines what tools the agent can call. Retrieved content cannot add tools, remove tools, or change permissions. This is checked at the dispatcher, not at the model. Deterministic, testable.

2. **Source labeling (partially enforceable).** Mark retrieved content with its provenance — "this text came from https://example.com via crawl_web." The model sees the label. Whether the model respects it depends on the model.

3. **Query disclosure limits (design choice).** How much of the agent's context is sent as part of a search query? The `ddgs` search tool sends only the query string. The `crawl_web` tool sends only the URL. Neither discloses the agent's conversation history, API keys, or other context to the search provider.

4. **Model behavior (not enforceable).** Prompt text alone cannot guarantee that the model ignores malicious content in a tool result. State this honestly. Where the agent lacks an enforceable boundary, say so rather than claim the chapter solved it.

**Proposed exhibit:** A deterministic test showing that a crafted page cannot change the agent's tool permissions (layer 1), paired with a live-model injection probe reported separately. One successful probe is evidence about that run, not a security guarantee. The distinction between deterministic enforcement and model behavior is the teaching point.

---

## 21.7 Search has a cost and a failure path

**Thesis:** Local search tools are free to call, but they can still fail. The agent must handle failures visibly.

With `ddgs` and `crawl4ai`, there is no per-search billing — the tools run locally. The cost is the tokens consumed by search results entering the conversation (covered by existing usage accounting). This is a significant simplification over native vendor search, where three different billing models applied.

Cover the failure modes:

- **Empty results.** The search found nothing. The agent should say so, not hallucinate an answer.
- **HTTP errors and timeouts.** The crawled URL returned a 404, 500, or timed out. Visible error, not silent swallowing.
- **Rate limiting.** Some search backends may rate-limit scraped requests. The error should be surfaced.
- **MCP server failure.** The sidecar process crashed or failed to start. The tool should be unavailable, not silently broken.
- **Content too large.** A crawled page exceeds reasonable size. `crawl4ai` has configurable limits; document them.

**Proposed exhibit:** A fake MCP server (or controlled HTTP responses) exercising empty results, HTTP errors, and timeout paths. The agent surfaces each failure rather than continuing silently.

---

## 21.8 Exercise, graded

Extend the existing agent with `search_web` and `crawl_web` tools. The grader uses a local HTTP server serving controlled pages — no live internet access during grading.

**Grader design:** The grader tests behavior, not implementation. A student who uses ddgs, crawl4ai, native vendor search, Firecrawl, or any other search mechanism passes as long as the behavioral checks hold. The grader's local HTTP server provides known pages; the student's tools must be configurable to hit the local server instead of the real internet.

Point allocations follow the implementation review; these are proposed behavioral checks:

| Check | Required observation |
|---|---|
| Search available | The agent has a tool that performs web searches and returns results |
| Fetch available | The agent has a tool that fetches URL content and returns it |
| Results in conversation | Search/fetch results appear as tool results the model can read |
| Persistence | Search results survive save/restore with content intact |
| Untrusted content | Retrieved text cannot grant capabilities or change tool permissions |
| Failure visibility | Empty results, HTTP errors, and timeouts produce visible diagnostics |
| No silent swallowing | A search or fetch failure does not result in the agent continuing as if it succeeded |

Mutation checks:

| Mutant | Expected failure |
|---|---|
| Remove search tool registration | "Search available" check fails |
| Remove fetch tool registration | "Fetch available" check fails |
| Allow retrieved content to grant tool access | "Untrusted content" check fails |
| Suppress error on fetch failure | "Failure visibility" check fails |

---

## 21.9 Taking it for a spin

Run a documentation task: verify a current API feature before implementing against it. Show the query, the search executing through the MCP bridge, the results, a follow-up crawl of the most relevant URL, and the sourced answer with the supporting passage visible.

Then demonstrate portability: save, restore, switch to a different vendor's model, and show the search results still usable in the conversation. This works because the results are text, not vendor-specific blobs.

Finish with the controlled injection page from §21.6. Show the deterministic enforcement (tool permissions unchanged) and report the model's behavior separately. The difference between what the code enforces and what the model chooses is the chapter's final teaching point.

---

## Decisions resolved

1. **Search mechanism:** Local tools via MCP servers (`ddgs` for search, `crawl4ai` for fetch). Not native vendor search.
2. **Vendor-native search:** Documented for the reader's understanding; not implemented as the default. Reader may choose it.
3. **Settings UI:** None. Search is always on. Disabling is deferred to a sandboxing chapter.
4. **Grader neutrality:** The grader tests behavioral contract, not implementation. Any search mechanism that meets the checks passes.
5. **Cross-vendor portability:** Automatic — local tool results are plain text that survive any model switch without normalization.
6. **Firecrawl:** Mentioned as a legitimate commercial alternative. Not implemented in the chapter.
7. **MCP bridge:** Both tools ship built-in MCP servers. Uses Chapter 12 infrastructure — no new dispatcher logic.
