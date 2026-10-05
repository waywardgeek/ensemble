# Chapter 21: Giving the Agent the Web

*Revised outline. Scope: always-on native search via the current LLM vendor. No settings UI, no cross-vendor adapter, no Off toggle (deferred to sandboxing chapter). No implementation or measured results are claimed here.*

## Through-line stake

The engineer needs current documentation to ship a correct change, but letting the agent read arbitrary web pages also lets strangers put instructions in its context; the chapter resolves with a sourced answer and a tested boundary around what retrieved content may cause the agent to do.

## Voice plan and evidence

Third-person mechanism throughout the body. An optional short opener uses Bill's account of CodeRhapsody's DuckDuckGo MCP and crawl4ai integration, without inventing a scene or claiming that extraction prevents prompt injection. Each section below names its concrete teaching fact or proposed exhibit. Proposed exhibits need captured results before the chapter presents them as observations.

Earlier discussion overstated the security advantage of vendor-native search. Server-side execution alone establishes neither sanitization guarantees nor resistance to malicious instructions. Exact API capabilities, supported models, tool versions, prices, and citation requirements need dated documentation and wire-level checks before publication.

## Proposed TL;DR contract

Add web search to the existing agent. The current vendor's native search is always available.

- Every API request includes the current vendor's native search tool declaration. The model decides when to search, just as it decides when to call any other tool.
- Search executes server-side inside the vendor's API. The agent's renderer adds the tool declaration; the vendor's servers perform the actual web request. Server-executed search must not re-enter the local tool dispatcher.
- Results retain source links, attribution, and available citation metadata through streaming, persistence, and restore. Unsupported or unavailable metadata is reported honestly.
- Retrieved content is untrusted evidence. It does not grant permission to run commands, disclose secrets, change settings, or widen tool access.
- Search usage is reported as part of the model's normal usage accounting. Failure paths (empty results, rate limits, timeouts) produce visible errors, not silent swallowing.
- Disabling search is deferred to a later chapter on sandboxing and capability restriction.

## 21.1 The idea in plain words

**Thesis:** A coding agent needs a way to check documentation that changed after its training data was collected.

Follow one task through the chapter: verify a current API feature before implementing against it. Show the query, the retrieved source, and the claim the source supports. Search discovers sources; fetching reads a particular source; browser automation interacts with a page. This chapter implements search and the source access available through its providers. General browser automation remains separate.

**Proposed exhibit:** An API question answered from a dated primary source, with the supporting passage visible beside the answer. Choose and record the task during implementation; do not invent a stale-model failure.

## 21.2 Three vendors, one pattern

**Thesis:** Each vendor's native search is a tool declaration on the existing API request, but the shapes differ.

Compare the three native search mechanisms using verified request specimens:

| Vendor | Mechanism | How it rides the request |
|---|---|---|
| Anthropic | `web_search` tool (Messages API) | Tool declaration in `tools[]`; model calls it; server executes; results return as `web_search_tool_result` content blocks |
| OpenAI | `web_search` tool (Responses API) | Hosted tool in the request; server executes; results inline with `url_citation` annotations |
| Gemini | `google_search_retrieval` (Generation API) | Tool/config on the generation request; search grounding happens server-side as part of generation |

All three are flags on the existing request. No second HTTP call, no second model, no second credentials. The renderer adds one vendor-specific declaration; the parser handles one new result shape.

**Teaching fact:** The renderer already branches per vendor. Adding search is the same shape as adding any other vendor-specific capability — a conditional tool declaration and a new response-content handler.

## 21.3 Fit search into the existing loop

**Thesis:** Server-executed search and local function tools have different execution owners.

Show the path: renderer adds the search tool declaration → vendor API executes the search server-side → response decoder extracts search results and citations → event log records them → observer/GUI displays them.

The critical invariant: the vendor's search tool call and result happen entirely on the server side. When the response arrives, the parser must recognize search results as server-handled content, not as a pending local tool call to dispatch. This is the same pattern as any vendor-hosted tool — the dispatcher must know which tools it owns and which the server already executed.

Keep vendor-specific request and response forms at the renderer/parser boundary. The event model carries a portable citation type so downstream consumers (observer, GUI, save/restore) don't branch on vendor.

**Proposed exhibit:** A request trace showing the search tool declaration going out and search results coming back, with the local dispatcher not touching them.

## 21.4 Keep the evidence attached

**Thesis:** A search-backed answer must retain the connection between its claims and its sources.

Carry source identity, URL, title when available, and provider attribution through a common citation representation. Explain the difference between a returned source list and a citation tied to a particular passage. Do not manufacture passage-level support when the provider supplies only links.

Citation shapes disagree across vendors: OpenAI and Gemini cite by text offset (`url_citation` with `start_index`/`end_index`); Anthropic attaches its own `cited_text` copy (≤150 chars). Offsets die the moment text is re-rendered or truncated, so a normalized record derived at ingest is the only portable representation. This is the same opaque-part problem as reasoning blocks in Chapters 19 and 20.

Handle streamed citation metadata, final responses, save/restore, and model switching. Preserve original provider data needed for continuation while presenting portable evidence. A citation demonstrates attribution; the reader still needs to check whether the source supports the claim.

**Proposed exhibit:** The same answer and source links before and after save/restore, followed by a model switch that re-renders citations from the portable representation.

## 21.5 A web page cannot authorize a command

**Thesis:** Web text can contain useful documentation and malicious instructions in the same paragraph.

Use a controlled page containing legitimate API documentation plus an instruction to disclose a test secret. HTML cleanup may remove scripts and navigation while leaving that instruction intact. Vendor-native retrieval does not remove the need for a trust boundary.

Teach source labeling, limited query disclosure, existing tool authorization, and the distinction between deterministic enforcement and model behavior. Prompt text alone cannot guarantee that the model ignores malicious content. Where the agent lacks an enforceable boundary, state the limitation rather than claim the chapter solved it.

**Proposed exhibit:** A deterministic test that retrieved content cannot change tool permissions, paired with a separately reported live-model injection probe. One successful probe is evidence about that run, not a security guarantee.

## 21.6 Search has a bill and a failure path

**Thesis:** Search is not free, and the agent must report its cost and handle its failures.

Native search usage rides the same API billing as the model request. Report available usage data through the existing usage-accounting path. Where a vendor itemizes search cost separately (Gemini charges per grounding query), surface that. Where search cost is bundled into token usage (Anthropic, OpenAI), say so honestly rather than inventing a separate meter.

Cover empty results, rate limits, partial streams, timeouts, and model/API combinations where the vendor doesn't support search. A model that lacks search support should not receive a search tool declaration — gate on `ModelFeatures`, not on hope. Failures produce visible errors.

**Proposed exhibit:** A fake-server response exercising the empty-result and error paths, showing the agent surfaces them rather than silently continuing.

## 21.7 Exercise, graded

Extend the existing renderers, parsers, event model, and citation display. Use local HTTP fakes that return search results in each vendor's format. Keep credentialed smoke tests separate and record their API/model/date. Point allocations follow the implementation review; these are proposed behavioral checks:

| Check | Required observation |
|---|---|
| Declaration present | Each vendor's renderer includes the native search tool in API requests for search-capable models |
| No local dispatch | Server-executed search results are not re-dispatched as local tool calls |
| Citation persistence | Source attribution survives save/restore with correct URLs and text |
| Citation portability | Citations render correctly after a model switch to a different vendor |
| Untrusted content | Retrieved text cannot grant capabilities or change tool permissions |
| Usage reporting | Search-related usage is reported through the existing accounting path |
| Failure visibility | Empty results, unsupported models, and errors produce visible diagnostics |

Mutation checks should include removing the search tool declaration, dispatching server results locally, dropping citations on restore, and allowing retrieved content to widen tool access.

## 21.8 Taking it for a spin

Run the documentation task from §21.1. Show the query, the search executing server-side, the sourced answer with citations, and the citations surviving a save/restore cycle. Switch to a different vendor's model and show the citations re-rendered from the portable representation. Finish with the controlled injection page from §21.5 and report what the harness enforced separately from what the model chose to do.

## Decisions resolved

1. **Provider set:** Native APIs only — Anthropic Messages `web_search`, OpenAI Responses `web_search`, Gemini `google_search_retrieval`. No DuckDuckGo/MCP as a shipped choice; independently configured MCP tools remain unaffected.
2. **Cross-vendor:** Eliminated. Search always uses the current LLM vendor's native API. No second credentials, no adapter, no bounded research query.
3. **Settings UI:** None. Search is always on. Disabling search is deferred to a sandboxing chapter.
4. **Research models:** N/A — no cross-vendor sidecar, so no second model to select.
5. **Fake server:** Must return vendor-specific search results (tool results for Anthropic/OpenAI, grounding metadata for Gemini) so the grader can exercise parsing and citation extraction without live API calls.
