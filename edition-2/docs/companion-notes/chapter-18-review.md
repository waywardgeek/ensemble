# Chapter 18 independent review

## Disposition

Accepted after focused recheck of the stable corrected release. All three material findings are resolved; no material issues remain from this review.

## Scope

The initial review read the stable-release manuscript and author note in full against the complete current `book/edition-3/chapter-18.md`, including exercises and ending, with the full voice guide and accepted Chapter 17/manuscript note as context. Clipped reads were recovered. The recheck inspected only the corrected manuscript passages, updated author note, and relevant current source sections; completed checks elsewhere were reused. No implementation, linked legacy guidance, build, grader, runtime, Git, or student pipeline was inspected. Only this review note was edited.

## Resolved findings

1. **Authentication and secret redaction.** “A connection is more than a POST,” paragraph beginning “Credentials belong on this transport side,” now correctly describes `auth-env`, connection-time Bearer resolution, and a logged anonymous attempt when unset, with acceptance left to the server. It explicitly requires configured-token redaction from reflected diagnostics and successful RPC content. Confirmed against “The web-search skill” and the transport’s non-2xx/redaction discussion.

2. **Deferred definitions and cache markers.** “Grant the capability, not the page’s wishes,” paragraph beginning “Catalogue presence is not itself a grant,” now distinguishes catalogue presence, inline `tool_reference` disclosure, callable authorization, and cache eligibility. It accurately attributes the October 22 HTTP 400 to combining `defer_loading: true` with `cache_control`, gives the last-non-deferred breakpoint/omit remedy, and preserves the catalogue, order, defer flags, paired references, and system/history markers. Confirmed against “Deferred definitions and cache markers.”

3. **Community-tool history.** “Web access belongs to the agent” removes the unsupported model-discoverer and progressive-setup actions. It retains the historical measurements and correctly limits the `crawl4ai` finding to the tested pip distribution, distinguishing Docker’s HTTP/SSE MCP endpoint. Confirmed against the source’s first-edition investigation account.

## Retained review conclusions

The H1 exactly matches the current source. The chapter has a clear human WHAT/WHY progression and carries Chapter 17’s distinction between selection, execution, and observed evidence forward. It covers POST reply framing and queue lifetime, session identity, skill/dispatch gates, prompt-injection limits, correlated multi-block results, the three error boundaries, wait-versus-job lifetime, actual-owner status, model-bucket accounting and unknown cost, snapshot timing, same-provider turn-boundary selection, and the observer shutdown race. The web and browser records remain bounded rather than promises of current availability. No cosmetic quota changes requested.
