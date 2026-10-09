# Chapter 17 preparation evidence

October 8, 2026. Author-only research for new Chapter 17, mapped from old Chapter
18 under the current procedure/global map. This thread previously implemented
new Chapter 4 and subsequently authored later contracts. It has historical prose,
review and grader exposure; it is not a cold Chapter 17 student. No current student
source or old implementation was inspected for this preparation, and no runtime,
grader, provider inference or credential access occurred. Historical files remain
untouched. The scope is outline/evidence only.

## Actual source reads

- Full current voice.md, chapter-writing-procedure.md and architecture.md,
  reloaded after compaction. Truncated combined outputs were completed through
  smaller reads. Architecture's replaceable MCP transport and optional GUI
  WebSocket tunnel remain explicit obligations.
- Complete old chapter-18.md in three ranges; complete caching-design.md,
  brief-ch18-coder.md and chapter-18-coder-review.md.
- Complete chapter-18-coder-review-2026-10-01.md in successive ranges. Selected
  Gemini review sections: summary, probes A–C, findings 1/5/6/7 and the closing
  correction; selected ephemera review opener, lifetime discussion and author
  ruling/end-of-day section. Their other sections were only located by headings.
- Current global-review.md map, usage/provenance and material/caching forward
  lessons, and full API/caching/subscription research gate. Old Chapter 22 TL;DR
  owner/per-model rules and its complete cost section were read as forward evidence.
- Complete inherited internal/grade/ch18_checks.go: seven checks totaling 100.
  No harness, grader execution or old implementation read is implied.
- New Chapters 9–16 remain the preceding author/reviewer context. Focused current
  rereads include Chapter 16 ownership/record/replay/retirement and capability
  clauses, Chapter 15 physical-record rules and the architecture ledger; earlier
  full-draft author reads are recorded in those chapters' evidence. Chapter 10's
  latest accepted-byte/watermark clarifications remain binding.
- Complete local OpenAI Docs skill, followed by official-domain search and actual
  opened documentation. No dedicated callable docs-search tool was discovered;
  web search was the available fallback. Requested .md variants of three pages
  returned tool InternalError; HTML pages were then opened/read instead.

## Historical findings and claim limits

| Source | What it supports |
|---|---|
| caching-design.md, dated September 27 | Caching blocked intended daily use and needed visible diagnostics. Its claim that every unmarked request is cold is an inference the later implicit-caching work contradicts. Its initial Anthropic section order is superseded. |
| Old Chapter 18 opener | The invoice motivation and attributed spending account. No independent ledger was recovered here for its dollar total, projected monthly savings or universal hit-rate claim; omit those numbers from new factual prose. |
| `374609548d18ffe4187f6758e94d007a95931e30`, full message | Compact marker pattern versus indented actual input, misleading zero markers, and test fixtures repeating the faulty assumption. Old reported before/after percentages are dated claims, not reproduced measurements. |
| `40afc7a6573d20fe531c1074ba003aa1b112ef94`, full message | Meter using empty stored model rather than actual environment-selected model; independent wire checks, helper/foreground request-count distinction and process-lifetime usage scope. |
| `b40c71827256f05cfed983e75cebb6435f3414a0`, full message | Hard-coded messages field contradicted contents:appended. Its additional explanation blaming missing Gemini directives is stale; later review establishes the implicit path. |
| `ca9d807d9e344375037d0c75bb29185192920aab`, full message | Four-location replacement of rolling pair, single-provider correction, old numerical correction and explicit identification of unverified claims. |
| `828e23650f7b1eadf0a82196cddca2a93aa6b5b7`, full message | Historical Responses migration and Chat Completions retention; fixture identity mismatch across frozen solutions. It does not establish today's route or authorize credential reuse. |
| `744ce22338823906acbd6334bcff0964baeb27af`, full message | GUI view caps must prune correlation indexes, and last-response usage differs from cumulative totals. New Chapters 7/8 already teach corresponding ownership. |
| Old Chapter 22 cost section | Applying the active model's rate to mixed historical counts silently reprices them. The new design already owns producing-model usage on Engine. |

The October 1 review expressly reports no live measurement of the then-new
four-marker arrangement. Keep earlier measurements bound to their earlier request
shape. The Gemini review corrects a confident conclusion drawn from a fixed-body
probe that did not exercise growing history. Its later inference about permanent
rounding loss and universal two-cold-turn behavior is not a present API guarantee.
The human lesson is to measure the shape under discussion and retain the first
wrong conclusion with its correction.

The ephemera review documents a historical feature-removal decision and Bill's
reaction. That is insufficient evidence to impose a current model-name rule or
silently suppress one-shot guidance in this edition. New Chapters 14–16 deliberately
preserve opaque provenance and captured material under explicit compatibility
rules. Checkpoint is save-only; successful handoff retires recall; later compression
can retire older note bodies. Several old cache-stability explanations therefore
need fresh derivation from the new projection, even where their motivation survives.

## Current primary-source research

Accessed October 8, 2026. These pages were opened and their relevant body passages
read; search snippets alone do not establish the following findings. No reliable
publication date was visible in the retrieved passages, so access date is not
presented as publication date. Links describe documentation, not measurements.

The [OpenAI prompt-caching guide](https://developers.openai.com/api/docs/guides/prompt-caching)
distinguishes explicit and implicit modes, warns that explicit mode without
markers produces no writes, and explains why extending a message can hide the
old lookup boundary. Its current lookup passage names the first two and latest
50 explicit boundaries. The [Chat Completions create reference](https://developers.openai.com/api/reference/resources/chat/subresources/completions/methods/create)
documents content-block prompt_cache_breakpoint and prompt_cache_options, but
names a latest-80 lookup window. These are conflicting published descriptions;
this outline does not resolve them by guessing which backend is deployed.
Four writes and an internal lookup history are different limits. The eventual
contract should grade its own emitted directives, not an undocumented resolution
of this discrepancy.

The [Sign in with ChatGPT example](https://developers.openai.com/cookbook/articles/sign-in-with-chatgpt)
uses an application installation identity, browser authorization and saved
application connection, with user-approved plan access. The [models/inference
guide](https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference)
requires the selected account's model discovery and public /v1/responses endpoint,
store:false and stream:true, complete supplied history and a terminal completed
response. It explicitly distinguishes failure/incomplete/interruption from success.
This supports an authorized application route, not extraction of another client's
credentials. No account connection or entitlement was tested here.

The [preview limitations](https://developers.openai.com/siwc/token-sharing-open-source/preview-limitations)
also prohibit max_output_tokens and temperature on that flow, among other fields,
and constrain tool placement. This is a consequential incompatibility with simply
reusing inherited request settings or helper output caps. The separate [app-server
example](https://developers.openai.com/siwc/token-sharing-open-source/codex-app-server)
uses an authorized token with Responses, but adopting that loop would not prove
Ensemble's own transport and accounting. The coordinator was notified; the
outline leaves route/surface/auth scope for explicit decision. No copied shell
example will put a credential in argv during eventual evidence collection.

The [Claude prompt-caching guide](https://platform.claude.com/docs/en/build-with-claude/prompt-caching)
currently describes tools → system → messages, four breakpoint slots including
automatic placement, model-specific floors, separate write/read/input counters
and limited lookback. It supports visible reusable boundaries, not a universal
floor or a guarantee that any stable request will hit. Prices and model capability
rows require rechecking against the actual route chosen for the later run.

The [Generate Content caching guide](https://ai.google.dev/gemini-api/docs/generate-content/caching?hl=en)
labels this API surface separately from newer documentation and describes implicit
caching as default for its named model families without guaranteed savings.
Its table lists model-specific floors and its advice concerns repeated prefixes.
It does not promise the old review's exact warm-up count or permanent 4,096-token
rounding theory. Retain the inherited surface explicitly; do not accidentally
grade another API's usage field because a generic documentation URL moved.

## Proposed checks and unresolved scope

The old seven checks establish lens presence, stable prefix, markers, derived
cost, process-lifetime scope, read counters and provider normalization. They do
not by themselves cover new semantic snapshots, causal helper accounting, exact
browser integers, multiple Agents, retired-body absence, auth funding scope or
the supported OAuth route. A replacement acceptance matrix must retain applicable
legacy properties while independently testing these new interactions. No shared
grader changes have been made.

The outline's D1–D7 are proposals: bounded per-purpose diagnostic lifetime,
truthful comparison units, provider marker policy, strict profile/replay extension,
per-model/scoped cost, subscription surface and public/browser visibility. In
particular, Responses/auth integration may belong partly with new Chapter 18;
deferring it cannot be mislabeled as completion of Bill's cache-retest gate.
The old preference for Gemini implicit caching is recorded, while no current
universal efficiency claim or provider restriction is inferred from it.

No raw historical live transcripts were audited for this preparation. No new
performance, cache hit, billing, audio or model-result claim is made. The later
student must retain source/executable/route bindings, actual human interaction,
all attempts and raw usage under an agreed bounded matrix. The original OAuth
report remains an investigation target, not an established current defect.

## Preparation freeze checks

Source SHA-256 values at this preparation boundary:

| Read source | SHA-256 |
|---|---|
| voice.md | `17883353cf9653c0df46081c216b5b1623dfa640619e06fe89bff46fc29dfad1` |
| chapter-writing-procedure.md | `1131abf07ea86d246421122b20ee368cce30fb2cd684a6ca3ab279c220ec7d3c` |
| architecture.md | `9dada72e65c36fb649659f9076b4ca2bdfc8379fef919a3b575652dca0e49327` |
| global-review.md | `22394e0daa00ef2682535958eff6bb1002700af3263f2beeb46d3f1e922e8c85` |
| Old chapter-18.md | `2224c498d28cb1ba2aaec8b08eb3e21153f80f816842b4898929522e3f173d11` |
| caching-design.md | `34f29c2887023bf184ee40a5284d5626c95a8ff28308c0e833f074322d2a9336` |
| chapter-18-coder-review.md | `e479e75b19951f40ea96afbdf1c1e8cc09e780e6be3c91860ad4e522bbd019ed` |
| chapter-18-coder-review-2026-10-01.md | `ab07411e1108af58c98d608ceef2f3b894bc492b56ac19ec27bffccd34695730` |
| chapter-18-gemini-review-for-author.md | `864e349e09bd7a772b4f8aa0cef74ae2ae96d2e91469a3f81fbea24c716ec53b` |
| chapter-18-ephemera-review-2026-10-01.md | `b84512e5180f94327c4ff820603cacdc32c5dd28970cc7e9b3b541b8ea9d2f22` |
| brief-ch18-coder.md | `81a100aa994795582b3725ee554c7d9bc8ab4afe75a1ed37128d0c5ef46dd204` |
| ch18_checks.go | `00ffe9ddfb7745c71325efd0230d14ebb9232b49fb80a3a678a682bc689b9bb7` |

Retained prose lint passes hard rules for both owned preparation files; short
outline/evidence length and negation-density warnings are reviewed as preparation
format, not padded into a full chapter. Scoped whitespace checks pass. Manual
review retains the invoice/false-instrument stakes while separating historical
report, current documentation and unperformed measurement. No build was needed.
Root agreed that the documented route/settings conflict and lookup-window
disagreement should remain explicit design questions. They are not permission
to choose a different client's credentials or silently weaken inherited caps.

## Partial manuscript preparation, October 8, 2026

The resumed author read the entire current voice, procedure, architecture and
workflow, complete preparation outline/evidence and `/root/reviewer_ch17`'s
advisory. Read the coordinator decisions at `e9d78da` before drafting. This phase
has broad historical author/reviewer context and is unsuitable for a cold student.
The earlier read ledger above remains attributed to its preparation phase; this
entry does not claim another whole-textbook read or a new audit of old raw runs.

Read the full local OpenAI Docs skill, searched the requested official topic,
then opened the primary pages. No dedicated callable official docs search was
available. Targeted body passages were read after the broad page retrievals;
search snippets were not used alone. The sources accessed again were:

- [Chat Completions create](https://developers.openai.com/api/reference/resources/chat/subresources/completions/methods/create),
  text-content fields, tool messages, request controls and usage detail.
- [Prompt caching](https://developers.openai.com/api/docs/guides/prompt-caching),
  selected mode, lookup-window description and disjoint input arithmetic.
- [Messages caching](https://platform.claude.com/docs/en/build-with-claude/prompt-caching),
  selected manual placement, eligibility, duration, prefix order and usage.
- [Generate Content caching](https://ai.google.dev/gemini-api/docs/generate-content/caching?hl=en),
  implicit path and model-specific floors. No new remote-cache lifecycle selected.

The chapter cites the exact pages beside the supported claims and preserves
their conflicting lookup-window descriptions. This author did not reopen plan
authorization documentation in this phase: D6 remains attributed to the existing
research and coordinator gate, and no dependent route was chosen. No capability
probe, price query, credential access, inference or provider call occurred.

Focused predecessor rereads covered Chapter 15 profile/initializers and request
reconstruction/retirement, Chapter 16 profile/initializers and strict helper facts,
and Chapter 2 usage normalization. Reread the complete `3746095` and `40afc7a`
commit messages and the historical caching-design motivation/capture sections.
These support the preserved incidents, without authenticating their old reported
percentages. Literal byte-fixture arithmetic was independently checked from the
printed strings: A is 44 bytes, B 74; equal prefixes A/A 44, A/B 42, A/C 39.
The synthetic rate fixtures are hand-derived exact fractions: 656/1,000,000,
1312/1,000,000 and their sum 1968/1,000,000 USD. They are not live prices.

The draft is deliberately partial. Sections 17.5, 17.7 and 17.8 name unfinished
strict schemas and route decisions rather than allowing a student to infer them.
No implementation or grader was edited. No build was run while the reviewer
owned the compiler. No reusable lint executable was found at the checked local
temporary paths; scoped prose lint is pending. Scoped whitespace, literal fixture,
story/cut and manual voice checks cover only the three author-owned files.

At the coordinator's separate request, this author also read Chapter 9 §9.6 and
Chapter 10's strict/admitted-byte/standalone/limit rules and sent grouped judgments
about the retained standalone duplicate-skill-argument and argument-string replay
regressions. No Chapter 9/10 teaching was changed in this phase. Those runtime
failures and fixes belong to their own gate, not Chapter 17 evidence.

## Subscription integration research after the decision

October 8, 2026. Bill accepted explicitly selected subscription use without a
per-response output-token cap; the coordinator records the decision at `706682c`.
The author reloaded the full voice/procedure, architecture, current partial
chapter, validation decisions and preparation advisory. Focused predecessor reads
covered Chapter 15's compression config, grammar and bounds and Chapter 16's
plain judge, fixed cap, response-byte bound and accepted-usage distinction.

Inspected `53ee8e7` and the current manuscript before editing. Its warmer story
pass is intentional editorial input from the original first-edition author, as
Bill confirmed. This proposal preserves that input rather than restoring the
earlier partial manuscript or claiming authorship of those enhancements.

Read the complete OpenAI Docs skill. Official-domain search was followed by
opened primary HTML pages and focused body reads. The Responses create reference
exceeded the browser tool's response-size limit; its official Markdown endpoint
was then read directly through an unauthenticated, bounded-time HTTP retrieval.
The web tool also refused Markdown content type for two smaller pages; their
HTML bodies and, where necessary, directly retrieved official Markdown supplied
the actual text. No search snippet alone establishes the proposed wire contract.

| Primary source, accessed October 8 | Verified contribution and limit |
|---|---|
| [Registration/sign-in](https://developers.openai.com/siwc/token-sharing-open-source/sign-in) | Own public-client registration, loopback callback, state/nonce/PKCE, issued client ID, code exchange, identity validation and granted plan scope. Documentation is not proof of this application's entitlement. |
| [Accounts and sessions](https://developers.openai.com/siwc/token-sharing-open-source/profiles-and-sessions) | Distinct registration/identity records, replacement refresh tokens, serialized refresh and revocation. Does not dictate the proposed Ensemble owner hierarchy or exclusive mounted-store lease. |
| [Token reference](https://developers.openai.com/siwc/token-sharing-open-source/token-reference) | Token response fields and expiry/rotation. The retrieved Markdown names earliest_refresh_at without defining enough semantics for a new scheduling rule; the outline explicitly avoids inferring one. |
| [Models and inference](https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference) | Selected-account model discovery, public Responses route, store:false, stream:true and terminal completion. No discovered live model list or inference was obtained. |
| [Preview limitations](https://developers.openai.com/siwc/token-sharing-open-source/preview-limitations) | Full supplied history, developer/instructions route, namespace function placement and excluded request controls including output cap. These restrictions motivate explicit per-purpose changes. |
| [Errors/recovery](https://developers.openai.com/siwc/token-sharing-open-source/errors-and-recovery) | Missing grants and routing/usage restrictions require visible dispositions; no silent billing fallback. The outline's no automatic paid retry is a conservative application policy. |
| [Tools guide](https://developers.openai.com/api/docs/guides/tools) | Actual namespace declaration envelope and nested function schemas. The proposed local namespaces do not authorize hosted tools or tool search. |
| [Function calling](https://developers.openai.com/api/docs/guides/function-calling) | Function call identity, arguments string and call_id-linked output; accepted output is carried into continuation. Earlier first/last-key argument handling is not imported from SDK examples. |
| [Responses create reference](https://developers.openai.com/api/reference/resources/responses/methods/create) | Function namespace, input/output item fields, assistant phase, inclusive input/cached/write usage and reasoning output detail; exact item/capsule fixtures still need publication. |
| [Reasoning](https://developers.openai.com/api/docs/guides/reasoning) | Stateless encrypted reasoning continuation and preservation of returned items. No hidden reasoning text is inferred or exposed. |
| [Streaming Responses](https://developers.openai.com/api/docs/guides/streaming-responses) | Typed item/content/text/argument events and terminal response outcomes. Existing actor acceptance and local bounds must apply to the new adapter. |

The prior guide/reference disagreement about cache lookup windows remains
attributed in §17.4. This research does not reconcile it by choosing the larger
number, establish plan-route marker availability, or infer a cache hit from the
published schema. No live price table was selected.

The integration proposal in the outline separates Bill's cap preference from
coordinator choices still under review: adding a capped API-key Responses control,
Connections ownership, a mounted credential-store lease, per-turn funding capture
and an explicit replay-item capsule. Chapter §17.8 now reflects the resolved
preference while staying partial. No full contract/student release follows.

No code, build, credential access, browser login, inference or provider probe was
performed. No chapter outside the three owned Chapter 17 author files was changed.
Scope remains manuscript/outline/evidence; the coordinator owns the gate record.

Ran the existing `edition2-lintprose` executable against the three owned files:
all hard checks passed. The 4,337-word partial chapter retains soft negation and
long-person-gap warnings; outline/evidence length warnings do not call for prose
padding. Scoped `git diff --check` passed. These are partial-draft checks, not
independent full-contract proofreading or runtime validation.


## Full contract drafting after R1–R4 acceptance

October 8, 2026. The coordinator authorized completion from `a5423ed` after reading
its full integration review. The author reloaded all current voice.md,
chapter-writing-procedure.md and architecture.md, read the complete new review
append, and re-read the existing partial chapter in bounded ranges. Focused
predecessor reads covered Chapter 10 identity/initializers and its now-64-MiB
session record rule, Chapter 14 versioned anchors/protection, Chapter 15 helper
config/retirement and Chapter 16 identity/initializer/helper limits. This is broad
author context, never a fresh student evaluation.

Used the already loaded full OpenAI Docs skill, searched current official route
and stream topics, then actually opened registration, Responses streaming,
preview limitations and conversation-state pages. Direct bounded-time retrieval
of the official Responses streaming and sign-in Markdown supplied focused schema
sections. The identity page's discovery/issuer/JWKS section was likewise read.
No account data or authenticated request was accessed. Existing primary-source
citations above remain attributed to their earlier reads; this is not a claim to
have re-read every earlier page in full.

The full draft publishes the accepted R1–R4 requirements and their deterministic
fixtures. Routine choices made explicit include a v7 wrapper around exact base
identities; a frozen capability table in creation identity; one raw terminal
payload and indexed references; a protected Responses bundle with whole-unit
handoff retirement; text/function-only new Responses mapping; 256 refresh waiters;
private store bounds; safe public protocol; and a finite initial five-route matrix.
These are coordinator-author design choices within the accepted direction, not
new quotations or rulings from Bill. Original-author story input at `53ee8e7`
remains intact.

Independently parsed every printed JSON/JSON-lines literal with the standard JSON
parser. The SSE fixture E is 532 ASCII bytes including its two LF terminators;
colon + 7,657 x bytes + two LF + E is exactly 8,192 bytes, and one added x gives
8,193. This verifies the printed byte recipe only, not a parser implementation,
provider stream or successful real judge. API-key fixture insertion preserves all
other printed continuation bytes; fictional identities are clearly offline.

The existing external edition2-lintprose executable was used without compilation.
Hard checks pass; the manuscript's roughly 9,400 prose words, negation density and
long technical stretch remain soft review inputs. No story was fabricated to
satisfy a numeric warning. Scoped diff/fixture checks are separate from the pending
independent full-contract review. The author edited only Chapter 17 manuscript,
outline, evidence and, under this handoff's explicit assignment, validation.
No runtime, grader, credential, login, provider, build or push action occurred.
