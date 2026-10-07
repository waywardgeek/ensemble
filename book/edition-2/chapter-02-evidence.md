# Chapter 2 evidence and source reconciliation

Status: author derivation and proposed contract; no implementation tested.

## Reading

Read first-edition chapter 2's design/contract through §2.8 and exercise,
relevant first-edition grader fixture definitions and harness/checks,
the chapter-2 derived fact sheet's core declarations, selected git
history, the complete global review map, current architecture decisions,
the full current voice guide and rewritten chapter procedure. First-edition
implementation is evidence for authors only, not a student baseline.

## Historical receipts

- `5a7dfca`: deleting call-bound opaque replay initially scored 100. The
  old fixture lacked opaque call material and had foreign provenance, so
  correct withholding also made the test vacuous. The independent
  same-model Gemini fixture closed the gap without changing points.
- `8a85d5b` and `f2894bf`: Ref serialization/refusal/render/redaction checks
  and deletion audit. A loader test must not be satisfied by a later
  renderer rejecting data the loader incorrectly accepted.
- `ef155b1`: live use found collisions in default log paths, unclear CLI
  behavior, and demonstrations contaminating their own fixtures. New
  evidence must run in an isolated working directory with explicit logs.
- `29cd739`: corrected claim about Go map serialization. Standard JSON
  encoding sorts map keys; the nondeterminism risk includes iterating a
  map to construct ordered slices before encoding.
- Current grader constants allocate parity10 and Ref3/2/2/4/4. The old
  printed parity25 table and surrounding points commentary are stale.

## Later lessons absorbed now

Shared structures/interfaces and parent access follow new chapter 1.
Separate log facts, current context, request configuration, and credential
material. Record provenance at capture and retain per-model usage before
model switching can misattribute it. Keep actor identity distinct from
entry purpose so skills, redaction, and recall do not require changing the
meaning of an existing field. Recorded bytes cannot alias parser buffers.

The optional GUI is a separate module/public client stub in this chapter.
The reviewer recommends no live browser requirement for an explicitly
unimplemented transport, matching Bill's permission to stub the GUI.
If actual WebSocket behavior is implemented or advertised, it becomes a
feature requiring its own live user-path demonstration.

## Current official documentation, read 2026-10-07

Used the OpenAI Docs skill for OpenAI-specific verification, searched
official domains, and opened the relevant pages. These are contract
references, not proof the new implementation works.

- [Chat Completions request/response](https://developers.openai.com/api/reference/resources/chat/subresources/completions/methods/create):
  `max_completion_tokens` bounds visible and reasoning output;
  `max_tokens` is deprecated and incompatible with o-series. Current
  usage documents `cached_tokens` and `cache_write_tokens` inside prompt
  details. Do not assume the latter is a fictional legacy-fixture field.
- [OpenAI prompt caching](https://developers.openai.com/api/docs/guides/prompt-caching):
  current cache observability is available, but this chapter does not
  perform the later OAuth/caching investigation or make a cache-hit claim.
- [Gemini generateContent](https://ai.google.dev/api/generate-content):
  prompt counts include cached content; thoughts and candidate counts
  contribute separately to reported total generation tokens. No current
  deprecation claim is copied from the first edition.
- [Messages caching usage](https://platform.claude.com/docs/en/build-with-claude/prompt-caching):
  ordinary input, cache creation, and cache read are distinct reported
  input categories. No cached-input price ratio is asserted here.

No credential file has been read by the author. No provider call, timing,
current model selection, or live feature pass is claimed.

## Draft checks and unfinished work

Ran the prose linter on the initial chapter-2 contract: exit 0, no hard
failures. Soft warnings concern length and negation density in an
847-word working draft; no filler added.

The contract is not ready for a student until the schema, literal fixtures,
directives, log selection, normalization formulas, and adapter subsets are
printed. One identity detail remains to choose: record the requested route
separately from any actual model version returned in `model`/`modelVersion`,
and explicitly define which identity keys usage and opaque replay checks.
Current config must not overwrite either historical fact. Root has the
recommendation to use response provenance for usage and keep request
provenance separately.
