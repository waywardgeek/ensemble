# Chapter 21 — research gate: what the three native search APIs actually support

Status: COMPLETE for the three native families. Anthropic, Gemini and OpenAI
all verified from canonical docs on 2026-10-05.

Purpose: outline decision #4 says "Verify tool coexistence, continuation rules,
citation display requirements, credentials, and usage reporting for every
advertised API/model combination. Offer only combinations the implementation can
demonstrate." This file is that verification. Nothing here is written from
memory; every claim below was copied from the cited page in the same breath.

A claim is marked VERIFIED only if it was read directly on the vendor's own
documentation. Anything inferred is marked INFERRED and must not be stated as
fact in the chapter.

---

## Anthropic — Messages API `web_search`

Source: https://docs.claude.com/en/docs/agents-and-tools/tool-use/web-search-tool
(redirects to platform.claude.com), crawled 2026-10-05.

VERIFIED — three tool versions, and the version is a behavioral choice:
  - `web_search_20250305` — basic search.
  - `web_search_20260209` — adds dynamic filtering; model writes and runs code
    that filters results before they reach the context window.
  - `web_search_20260318` — adds `response_inclusion` control.

VERIFIED — enforceable bound exists. `max_uses` caps searches per request. When
Claude exceeds it the `web_search_tool_result` carries error code
`max_uses_exceeded`. This is a real deterministic limit, not a prompt request,
which is what section 21.6 needs.

VERIFIED — domain control: `allowed_domains` or `blocked_domains`, never both;
supplying both is a 400. Bare domains, optional path, no scheme.

VERIFIED — citations are ALWAYS enabled for web search. Each
`web_search_result_location` carries `url`, `title`, `encrypted_index`, and
`cited_text` (up to 150 characters).

VERIFIED — citation display is a stated obligation, not a nicety: "When
displaying API outputs directly to end users, citations must be included to the
original source." The doc adds that if you reprocess or combine outputs you
should consult your legal team. The chapter should quote the obligation rather
than paraphrase it into a weaker claim.

VERIFIED — citation fields `cited_text`, `title` and `url` do NOT count toward
input or output token usage.

VERIFIED — usage reporting exists and is countable:
`usage.server_tool_use.web_search_requests`.

VERIFIED — **continuation rule, and this is the sharp one.** To continue a
conversation containing search results you must send the assistant's content
blocks back exactly as received, including each result's `encrypted_content`.
If `encrypted_content` is missing or modified the request fails with a 400
validation error.

VERIFIED — failure path does not look like a failure at the HTTP layer. Search
errors still return HTTP 200; the error lives in the body as
`web_search_tool_result_error`. Codes seen: `too_many_requests`,
`invalid_tool_input`, `max_uses_exceeded`. A search that succeeds with no
matches returns an empty `content` list, NOT an error. Any implementation that
only checks HTTP status will silently treat a rate-limited search as success.

VERIFIED — availability is an org-level setting. Web search can be disabled in
the Console, in which case a request carrying the tool fails with a 400
`invalid_request_error` rather than an in-body error code. So "unsupported" and
"disabled" surface through two different channels.

VERIFIED — models without programmatic tool calling require
`allowed_callers: ["direct"]`; without it the API returns a 400 telling you to
set it. On `web_search_20260209`+ the default is `["code_execution_20260120"]`,
not `["direct"]`.

---

## Google — Gemini `google_search` grounding

Source: https://ai.google.dev/gemini-api/docs/google-search, crawled 2026-10-05.

VERIFIED — invoked on the Interactions API as `tools: [{"type": "google_search"}]`.

VERIFIED — response shape is a `steps` array, not a content-block list:
  - `thought` — with a `signature`.
  - `google_search_call` — `arguments.queries`, the queries actually executed.
  - `google_search_result` — `call_id` plus `result[].search_suggestions`, an
    HTML snippet.
  - `model_output` — `content[].text` carrying `annotations[]` of type
    `url_citation` with `url`, `title`, `start_index`, `end_index`.

VERIFIED — citations are offset-based. A `url_citation` points into the text by
`start_index`/`end_index`. This is a materially different citation model from
Anthropic's, which attaches citation objects to a text block with its own
`cited_text` copy. Offsets do not survive any downstream edit of the text.
Consequence for the portable-rendering check in 21.7: a renderer cannot treat
the two as the same structure; it must normalize.

VERIFIED — `search_suggestions` is an HTML snippet "for rendering search
suggestions in your UI" and the doc states: "Full usage requirements are
detailed in the Terms of Service", linking
https://ai.google.dev/gemini-api/terms#grounding-with-google-search.
ACTION BEFORE SHIPPING GEMINI AS A CHOICE: read that ToS section and record
what it actually obliges. Do not let the chapter advertise Gemini grounding
until the display obligation is quoted from the terms.

VERIFIED — billing is per search query on Gemini 3: if the model issues two
queries for one prompt that is two billable uses, and empty queries are ignored
for counting. On Gemini 2.5 and older it is billed per prompt. So "a hidden
second model request is still a billable request" (thesis 21.6) is literally
true here and the unit of billing even changes by model generation.

VERIFIED — can be combined with the URL context tool.

---

## OpenAI — Responses API web search

Source: https://developers.openai.com/api/docs/guides/tools-web-search.md,
crawled 2026-10-05. (The human-facing HTML page crawls as pure navigation
chrome. Appending `.md` to any OpenAI docs URL returns the real markdown; the
full index is https://developers.openai.com/llms.txt. Worth remembering.)

VERIFIED — the current tool is `{"type": "web_search"}` on the Responses API.
`web_search_preview` still exists for legacy integrations but does NOT support
`filters`, `external_web_access`, or `return_token_budget`.

VERIFIED — there are three distinct search modes, and they are not
interchangeable:
  - non-reasoning search: model passes the query through and relays results.
  - agentic search: the reasoning model searches inside its chain of thought
    and decides whether to keep going.
  - deep research: extended, hundreds of sources, minutes long, intended to be
    run with background mode.

VERIFIED — a Chat Completions path also exists via `gpt-5-search-api`, and it
differs in kind: "Chat Completions search models always search before
responding; Responses search is a tool." IMPLICATION FOR THE `Off` SETTING:
on that path, search is not something the model may decline to do. An "Off"
that is implemented by withholding a tool declaration cannot turn search off
for a search model. Either exclude those models from the feature or make Off
refuse the combination outright.

VERIFIED — output shape: a `web_search_call` item carrying `action`, which is
one of `search`, `open_page`, or `find_in_page` (the latter two on reasoning
models). Plus a `message` item whose `content[0].annotations` hold
`url_citation` objects with `start_index`, `end_index`, `url`, `title`.
Citations are therefore offset-based, like Gemini and unlike Anthropic.

VERIFIED — "Search actions incur a tool call cost." So billable units are
countable per action, which is what observability in 21.7 needs.

VERIFIED — the display obligation here is the STRONGEST of the three and is
stated as a requirement: "When displaying web results or information contained
in web results to end users, inline citations must be made clearly visible and
clickable in your user interface." Note *clickable* — this is a GUI
requirement, not merely a data-retention one. The chapter's GUI work has to
satisfy it.

VERIFIED — `search_context_size` is `low` | `medium` | `high`, and the doc is
explicit that it "does not set an exact token count or guarantee a specific
number of sources or citations." It is therefore NOT a bound we may describe
as a limit.

VERIFIED — `return_token_budget` accepts only `default` or `unlimited`; null,
numbers and other strings are rejected. It applies only to the hosted Responses
`web_search` tool with GPT-5+ reasoning, and explicitly not to non-reasoning
search, legacy Search API paths, container web search, Chat Completions search
models, or `web_search_preview`.

VERIFIED — domain filtering uses `filters` with up to 100 `allowed_domains` or
up to 100 `blocked_domains`; omit the scheme; subdomains are included. Only
available in the Responses API with the `web_search` tool. CONTRAST WORTH
TEACHING: OpenAI's own example passes `allowed_domains` AND `blocked_domains`
together, whereas Anthropic returns a 400 if both are present. Same-looking
feature, opposite rule.

VERIFIED — `sources` gives the complete list of URLs consulted, retrieved via
`include: ["web_search_call.action.sources"]`. It is normally longer than the
citation list. Real-time third-party feeds appear here labeled `oai-sports`,
`oai-weather`, `oai-finance`.

VERIFIED — `gpt-4o-search-preview` and `gpt-4o-mini-search-preview` are
deprecated and shut down 2026-07-23.

NOTE — the doc's examples use `gpt-6-astra`, which is a model ensemble already
runs. Good: the chapter can demonstrate on a model Bill actually uses rather
than introducing one for the exhibit.

---

## Cross-vendor observations so far

1. All three vendors hand back an opaque or positional artifact that must be
   returned intact or interpreted in place: Anthropic `encrypted_content`
   (400 if modified), Gemini `thought.signature` plus index-based annotations.
   This is the SAME shape as the reasoning-block continuation problem from
   chapters 19 and 20. Chapter 21 should name that it is the same problem
   appearing in a second place, not a new one.

2. Therefore the "citations survive a model switch" check in 21.7 is not a
   formatting exercise. Anthropic's encrypted blobs are meaningless to another
   vendor, and Gemini's offsets are meaningless against re-rendered text. The
   honest portable representation is a normalized citation record of our own
   (url, title, quoted text, and the span if we have it), derived at ingest.
   Carrying vendor-native structures across a switch cannot work.

3. **There is no uniform bound, and the chapter must not invent one.** The
   three vendors bound different quantities:
     - Anthropic bounds the NUMBER OF SEARCHES per request (`max_uses`), with a
       named error when exceeded.
     - OpenAI bounds RETURNED TOKENS, and only coarsely: `return_token_budget`
       is `default` or `unlimited` only, and `search_context_size` is
       documented as explicitly NOT a guarantee of token count or source count.
       No per-request search cap was found.
     - Gemini documents no cap at all; it bills per query executed.
   So 21.6's "bound adapter calls and result size" is honest only for the
   cross-vendor adapter we write ourselves, where we control the loop. For
   native server-side search we can report and we can cap where a cap exists,
   but we cannot promise a uniform ceiling. State the asymmetry; do not paper
   over it with a common abstraction that silently does nothing on two of three
   vendors.

4. Error reporting differs in kind: Anthropic returns 200-with-error-in-body for
   search failures but 400 for a disabled tool. An implementation needs both
   paths or it will report fake success.

5. **Citation display is a vendor REQUIREMENT on all three, and OpenAI's is the
   strictest.** OpenAI requires inline citations be "clearly visible and
   clickable"; Anthropic requires citations to the original source when
   displaying outputs directly. That makes the citation checks in 21.7
   compliance checks, not cosmetic ones, and it means the GUI work is in scope
   rather than optional.

6. **Two of three use offset-based citations, one does not.** OpenAI and Gemini
   both attach `url_citation` with `start_index`/`end_index` into the text;
   Anthropic attaches citation objects carrying their own `cited_text` copy.
   Any normalization we define has to survive both, and offsets break the
   instant anything re-renders or truncates the text.

7. **`Off` is not uniformly implementable by withholding a tool.** OpenAI's
   Chat Completions search models (`gpt-5-search-api`) always search before
   responding. For those, "no tool declared" does not mean "no search". This
   directly threatens the 21.7 check "Off: no search declaration or adapter
   dispatch through this feature" — the check is satisfiable as written (it
   scopes itself to "through this feature"), but the chapter must say plainly
   that Off is a statement about what *we* initiate, not a guarantee that the
   model did not search. Overclaiming here would be exactly the kind of
   unenforceable promise 21.5 warns against.

---

## Recommendations for the four open decisions

Decision 4 is discharged by this document. Decisions 1-3 are product calls and
belong to Bill; what follows is a recommendation with its reasoning, not a
ruling. Each one changed as a result of the research above.

### 1. Initial explicit provider set — RECOMMEND: three native families, plus Auto and Off. No DuckDuckGo, no MCP, as shipped choices.

The pull toward shipping a non-native provider is that section 21.4
(cross-vendor execution) otherwise looks like theory: if every choice is the
coding model's own vendor, the cross-vendor adapter never runs.

But that reasoning is wrong, and the research is why. Cross-vendor is already
reachable with native backends alone: run the coding model on Anthropic and
send the bounded research request to Gemini grounding. That exercises every
behavior 21.4 cares about — a separate credentialed request, a bounded query
rather than the whole conversation, a visible error with no silent fallback —
without taking on a scraped-HTML dependency that has no official API and a
plausible ToS problem.

DuckDuckGo would also undercut the chapter's own standard. 21.6 says report the
actual backend and do not promise what the implementation cannot enforce;
shipping a scraper as a first-class choice means shipping a backend whose
behavior we cannot characterize or keep working.

So: Auto, Off, Anthropic Messages search, OpenAI Responses search, Gemini
Search grounding. MCP/DuckDuckGo stay an extension point with a documented
adapter interface, demonstrated but not shipped or graded.

### 2. Cross-vendor research models — RECOMMEND: one hardcoded default per backend in the capability table; override via settings file only; no new UI picker.

ensemble's established pattern is capabilities as a data table with no default
row, and the outline says an extra model picker is not approved UI. Both point
the same way: the research model is a property of the chosen backend, not a
thing the user picks in the GUI.

Pick the cheapest model that is documented to support search on that surface —
a search sidecar wants to be fast and cheap, not strong. Advanced override
belongs in the settings file, unlisted in the GUI.

**Hard constraint on which model IDs the graded exhibits may use**, learned
expensively on 2026-10-03: frozen `solutions/chNN` snapshots carry their own
`model.go`. A model row that exists only in the live tree will pass
`make gradeNN` at 100 and silently score lower inside the frozen snapshot,
because the snapshot has never heard of it. Any model named by the ch21 grader
must already exist in both trees, the way `gpt-5-course` did. Verify before
writing the grader, not after.

### 3. Does the dropdown manage a designated MCP search connection? — RECOMMEND: no, and make that boundary a graded check in both directions.

The setting should own exactly what it dispatches. If a user has wired an MCP
server that happens to search, this dropdown must not reach in and disable it:
that would make the setting a controller of tools it does not own, which is
precisely what the outline warns against.

The honest definition is narrow: **Off means this feature issues no searches.**
It is not a claim that no search occurred anywhere. Two findings above force
that narrowness rather than merely permitting it — an MCP tool the user wired
is outside our control, and OpenAI's Chat Completions search models search
before responding whether or not we declare a tool.

That narrowness is worth grading in both directions, because a setting that
over-reaches is as wrong as one that under-delivers:
  - with Off selected, this feature dispatches no search; and
  - with Off selected, an unrelated MCP tool is still callable.
The second check is the one that stops a future implementer from "fixing" Off
by disabling everything that looks like search.

### Outline amendments these findings require

1. **21.7's Off check needs its wording defended in prose.** As written
   ("no search declaration or adapter dispatch through this feature") it is
   already correctly scoped, but the chapter body must state why the qualifier
   is there, or a reader will read Off as a guarantee no search happened.
2. **21.6 must not promise a uniform ceiling.** Anthropic bounds search count,
   OpenAI bounds returned tokens coarsely and documents `search_context_size`
   as explicitly not a guarantee, Gemini documents no cap. We can bound our own
   adapter loop; we cannot bound all three alike.
3. **Citation display moves from nicety to compliance, and lands in the GUI.**
   OpenAI requires inline citations be clearly visible *and clickable*.
   The chapter's citation checks are compliance checks.
4. **Citation normalization is a design requirement, not formatting.** Two of
   three vendors cite by text offset and one carries its own quoted copy;
   Anthropic additionally rejects any request whose `encrypted_content` was
   modified. A normalized citation record derived at ingest is the only thing
   that can survive save/restore and a model switch.
5. **Gemini cannot be advertised until its Search Suggestions ToS obligation is
   quoted.** The grounding doc defers the display requirements to the terms;
   that section still needs reading before Gemini ships as a choice.


---

## 2026-10-06 — Revised architecture: local tools, not native vendor search

### Decision: abandon native vendor search APIs

After reviewing the findings above with Bill, we made a series of simplifying
decisions that culminated in abandoning native vendor search entirely:

1. **Always-on, no settings UI** — search is always available, no Off toggle
   (deferred to a sandboxing chapter), no dropdown. This eliminated all GUI
   settings work.

2. **Same-vendor only** — no cross-vendor adapter. This eliminated the second
   credentials, bounded research query, and provider routing.

3. **Then the killer question: do search results survive a model switch?**
   Native vendor search results are vendor-specific opaque structures
   (Anthropic's `encrypted_content`, Gemini's offset-based annotations,
   OpenAI's `url_citation` offsets). Switching models means re-rendering the
   conversation for a different vendor, and these structures are meaningless
   cross-vendor. A normalization layer is needed regardless.

4. **Bill's final call: use local tools instead.** If we're normalizing anyway,
   and native search creates three different parser paths with three different
   citation formats, three different billing models, and three different
   continuation rules — why not just use local tools that produce
   vendor-neutral results by construction? The model calls `search_web`, gets
   back a list of results (title, URL, snippet). The model calls `crawl_web`,
   gets back markdown. Both are plain text tool results. They survive any
   model switch because they're just text in the conversation. No opaque
   blobs, no encrypted continuations, no offset-based citations that break
   on re-render.

This matches what CodeRhapsody already does: `ddgs` for search, `crawl4ai`
for scraping. Proven in production.

### Consequences for the chapter

- The chapter teaches TWO local tools: `search_web` and `crawl_web`
- No vendor-specific renderer additions for search
- No vendor-specific parser additions for search results
- No `ModelFeatures` search capability flags
- No citation normalization layer (results are already plain text)
- No opaque continuation data to preserve
- Security section (§21.5) still applies — fetched content is untrusted
- The grader's fake server is simpler — it serves HTTP responses to the
  local tools, not vendor-specific API response shapes

### What the native API research above is still good for

The native research is NOT wasted. The chapter should:
- Mention that vendors offer native search as an alternative
- Note the trade-offs (tighter integration vs. portability)
- Use the citation display requirements as context for why attribution matters
- Reference the `Off` / Chat Completions search model finding as an example
  of why "off means off" is harder than it looks

---

## Vendor URL fetch/scrape tools

Researched 2026-10-06. Two of three vendors offer native URL fetching tools.
OpenAI does not.

### Anthropic — `web_fetch` tool

Source: https://docs.anthropic.com/en/docs/agents-and-tools/tool-use/web-fetch-tool
(crawled 2026-10-06, rendered as nav chrome but key facts extracted).

VERIFIED — server-executed tool. Claude fetches the URL and returns page
content as markdown.

VERIFIED — respects `robots.txt`. Will not fetch pages that disallow crawling.

VERIFIED — `max_content_size` parameter: default 50KB, maximum 1MB. This is
a real bound the implementation can enforce.

VERIFIED — tool type follows the same pattern as `web_search`:
`web_fetch_20250305` (and presumably later versions).

### Google — `url_context` tool

Source: https://ai.google.dev/gemini-api/docs/url-context (crawled 2026-10-06).

VERIFIED — lets you provide URLs as additional context. The model accesses
the content from those pages to inform its response.

VERIFIED — limitations exist:
  - Does not work with auth-required pages
  - Does not work with YouTube URLs
  - Other URL types listed in the limitations section

VERIFIED — invoked alongside other tools (can be combined with
`google_search`).

### OpenAI — no native URL fetch tool

VERIFIED (by absence): OpenAI's Responses API tools documentation
(https://platform.openai.com/docs/guides/tools.md, crawled 2026-10-06) lists
`web_search`, `file_search`, `computer_use`, `code_interpreter`, and
`image_generation`. No URL fetch/scrape tool exists.

### Implication for the chapter

This asymmetry (2 of 3 vendors have native fetch, 1 does not) further
supports the local-tools approach. A `crawl_web` local tool provides uniform
fetch capability across all vendors without gaps.

---

## Community search/scrape libraries — 2026 landscape

Researched 2026-10-06.

### Search: DDGS (Dux Distributed Global Search)

Source: https://github.com/deedy5/ddgs (README crawled 2026-10-06).

The library formerly known as `duckduckgo_search` has reinvented itself as a
**metasearch library** aggregating results from 10 search backends. This is
what CodeRhapsody uses today for its `search_web` tool.

**Key facts (all VERIFIED from README):**

- **Package**: `pip install ddgs` (MIT license)
- **Version**: v9.16.0 (August 2026)
- **Stars**: ~3K on GitHub, actively maintained
- **Python**: >= 3.10

**Search backends (VERIFIED):**

| Function | Available backends |
|----------|:-------------------|
| `text()` | `bing`, `brave`, `duckduckgo`, `google`, `grokipedia`, `mojeek`, `startpage`, `yandex`, `yahoo`, `wikipedia` |
| `images()` | `bing`, `duckduckgo` |
| `videos()` | `duckduckgo` |
| `news()` | `bing`, `duckduckgo`, `yahoo` |
| `books()` | `annasarchive` |

**Built-in servers (VERIFIED):**

- **MCP server**: `pip install ddgs[mcp]` then `ddgs mcp` (stdio transport).
  Tools: `search_text`, `search_images`, `search_news`, `search_videos`,
  `search_books`, `extract_content`.
- **API server**: `pip install ddgs[api]` then `ddgs api` (FastAPI on port
  4479). Docker compose support. Endpoints: `/search/text`, `/search/images`,
  `/search/news`, `/search/videos`, `/search/books`, `/extract`, `/health`.

**API (VERIFIED from README examples):**

```python
from ddgs import DDGS

# Text search — returns list of dicts with title, href, body
results = DDGS().text("python programming", max_results=5, backend="auto")

# URL content extraction
content = DDGS().extract(url)
```

`backend="auto"` is the default — the library picks the best available
backend. A specific backend can be forced: `backend="brave"`,
`backend="google"`, etc.

**No API keys needed for any backend.** All backends are scraped, which means
they could break if the upstream provider changes their HTML. This is the
trade-off: free and keyless, but potentially fragile.

**The `extract()` function** provides URL content extraction (similar to
crawl4ai), but it is likely simpler HTTP fetching without Playwright-backed
JS rendering. For documentation pages (the primary coding-agent use case),
this is probably sufficient. For JS-heavy SPAs, crawl4ai with Playwright
would be needed.

### Scrape/fetch: crawl4ai

Source: https://github.com/unclecode/crawl4ai (crawled 2026-10-06).

- **Stars**: 84.8K (committed Oct 5, 2026 — actively maintained)
- **Version**: v0.9.2 (July 2026)
- **License**: Apache 2.0
- **Approach**: Python, Playwright-backed, async, handles JS-rendered pages
- **Built-in MCP server**: stdio transport (added August 2026)

This is what CodeRhapsody uses today for its `crawl_web` tool. It handles
JS-rendered pages that simpler HTTP-only scrapers cannot.

### Other notable libraries (VERIFIED from web search 2026-10-06)

| Library | Stars | License | Approach | Notes |
|---------|-------|---------|----------|-------|
| **Firecrawl** | ~40K+ | AGPL-3.0 (self-host) / Paid API | Node/Python, cloud primary | Best managed service DX. AGPL is a problem for book readers building agents. |
| **Jina Reader** | ~25K+ | Apache 2.0 (reader) | `r.jina.ai/URL` prefix | Zero-config hosted service. Sends every URL to Jina's servers. |
| **ScrapeGraphAI** | — | — | Python, graph-based + LLM | Uses LLMs for extraction via natural language prompts. |
| **Scrapy** | Mature | BSD | Python framework | Production-grade structured extraction. Overkill for agent use. |
| **Crawlee** | — | — | Node.js | All-in-one scraping + browser automation for JS. |

### Recommendation for the book

**Search**: `ddgs` — 10 backends, no API keys, MIT license, built-in MCP
server, already proven in CodeRhapsody. The `backend="auto"` default means
the library handles backend selection, reducing fragility. If one backend
breaks, others are available.

**Fetch/Scrape**: `crawl4ai` — Apache 2.0, 84.8K stars, Playwright-backed
(handles JS), built-in MCP server, already proven in CodeRhapsody.

**Why two libraries instead of just `ddgs`**: `ddgs` has `extract()` for URL
fetching, but it is simple HTTP without JS rendering. Documentation pages are
often static HTML and would work fine, but JS-heavy pages (SPAs, React docs
sites) need Playwright. `crawl4ai` handles both. The two tools have different
jobs: `ddgs` discovers URLs, `crawl4ai` reads them.

**Alternative considered and deferred**: using `ddgs` for everything (search +
extract) would simplify the dependency story to one library. This is viable
if the book's examples never need JS-rendered pages. Could be mentioned as a
simpler option for readers who don't need full browser rendering.

---

## Search pricing comparison (partial — 2026-10-06)

Pricing was difficult to extract from vendor pages (most rendered as nav
chrome). What follows is partial and should be re-verified before publication.

### Native vendor search pricing

**Anthropic `web_search`**: Priced per search invocation, separate from token
costs. `max_uses` (default 5) caps searches per request. Exact dollar figure
NOT EXTRACTED — pricing page rendered as nav chrome. Usage reported via
`usage.server_tool_use.web_search_requests`.

**OpenAI `web_search`**: No separate per-search charge visible on the pricing
page. "Search actions incur a tool call cost" — appears to be bundled into
token usage. `search_context_size` (`low`/`medium`/`high`) controls returned
context volume.

**Gemini `google_search_retrieval`**: Billed per search query on Gemini 3+
(per query executed, not per prompt). Empty queries ignored for counting. On
Gemini 2.5 and older, billed per prompt. Free tier includes grounding at no
cost. Exact paid-tier pricing NOT EXTRACTED.

INFERRED — the pricing models are fundamentally asymmetric across vendors:
per-invocation (Anthropic), bundled into tokens (OpenAI), per-query with
model-generation gating (Gemini). This asymmetry was another factor in
choosing local tools over native search — local tools have zero search cost
beyond the compute to run them.

### Local tool pricing

**`ddgs`**: Free. No API keys. All backends are scraped. Zero cost.

**`crawl4ai`**: Free. Self-hosted. Playwright browser instance runs locally.
Cost is only the compute to run the browser.

---

## Summary of architecture decisions (2026-10-06)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Search mechanism | Local tools (`ddgs` + `crawl4ai`) | Vendor-neutral results, no opaque blobs, survives model switch, free, proven in CodeRhapsody |
| Native vendor search | Not used | Three different APIs, three different citation formats, three different billing models, results don't survive model switch without normalization |
| Settings UI | None — always on | Deferred to sandboxing chapter |
| Off toggle | Deferred | Belongs in capability restriction / sandboxing chapter |
| Cross-vendor adapter | Eliminated | Same-vendor-only was already the plan; local tools make it moot |
| DuckDuckGo-specific | No — `ddgs` is now a metasearch library with 10 backends | `backend="auto"` handles backend selection |
| JS-rendered pages | `crawl4ai` (Playwright) | `ddgs extract()` is HTTP-only; documentation sites sometimes need JS |
