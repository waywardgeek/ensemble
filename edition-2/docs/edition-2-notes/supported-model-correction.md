# Supported-model correction, October 11

Coder: `edition2_grader_coder`, GPT-6 Astra/high. Independent review pending.
Mandatory handoff: `.agents/skills/ensemble-coding/SKILL.md`, reread in full;
current carryover and workflow also reloaded in full. Chapter15 stays paused.

Bill excludes Anthropic5.5 from current examples, defaults and future tests
because of his observed ephemeral-context incompatibility. This change applies
that support decision; it does not infer support from discovery or reproduce
provider behavior. No paid call, original-grader edit or frozen-export edit.

## Changes and limits

- Chapter5's runnable media map and revised grader default use the already
  declared `claude-sonnet-4-6`. Frozen Edition2 `ch05` does not declare Sonnet5.
  Legacy original6 instead declares Sonnet5 and not Sonnet4.6; its existing
  `-media-models` override selects Sonnet5, `gpt-6-astra`, `gemini-3.8-flash`.
  No common real Anthropic row exists in these two old tables. We did not add one.
- Only authoritative `solutions/edition-2/main/internal/engine/media.go` loses
  Haiku5.5, Opus5.5 and Sonnet5.5 capability/pricing rows. These now follow the
  existing unknown-model behavior: no declared media/streaming/window/pricing
  capability. Ordinary text remains permitted for unknown identifiers; this is
  removal of support claims, not a new model-name rejection policy.
- Active tests and the active speech harness select existing older rows.
  Distinct-model CLI/GUI controls use Opus5 then Sonnet5, so a no-op selection
  cannot pass by comparing identical names. Pricing/status controls use priced
  Sonnet5 then existing unpriced Opus5. Selection preserves previously measured
  cost; actual unpriced work makes total cost unknown. No price was invented.
- Sonnet5 is the only remaining priced Anthropic row. The two-priced Anthropic
  switch/judge checks therefore no longer establish mixed known-rate accounting.
  Root explicitly accepted that limit: historical frozen tests/evidence retain
  the old claim, while current supported OpenAI tier arithmetic remains exercised.
  Anthropic raw disjoint usage, exact Sonnet pricing, session/restore behavior,
  auxiliary usage and unknown cost remain checked.
- The dated captured empty-tool-argument SSE fixture keeps its original Haiku5.5
  model string and bytes. Its current request uses Sonnet5; replaying historical
  parser input does not claim a new Haiku run. All other historical reports,
  original chapters/graders, existing tags and frozen `chNN` exports are unchanged.

The removed media rows originated in student commits `8dc6dd82` (Haiku/Opus5.5)
and `7c940fc3` (Sonnet5.5). Chapter5's old runnable mapping/default came from
`e7f40fa9`. Removing those rows adds no production abstraction or exported symbol;
production shrinks by two lines. Test changes retain the real Agent, Engine,
HTTP parsing, GUI and browser paths with fakes only at external boundaries.

## Actual validation

Raw outputs and SHA-256 bindings are in
[`supported-model-correction`](../../solutions/evidence/supported-model-correction).

| Check | Actual result |
| --- | --- |
| Active core `go test ./... -count=1` | PASS; ensemble package31.235s |
| Active GUI `go test ./... -count=1` | PASS8.188s |
| Revised `TestChapter5Boundaries` | PASS4.181s |
| Focused media through actual binaries | Accepted10/10; legacy0/10 |
| Active pricing/status/media focused race checks | PASS1.113s |
| Core, GUI and revised grader vet | PASS |
| Changed JavaScript syntax and harness `--describe` | PASS |
| Active browser cache, speech and listener lifetime |3pass/2fail; see below |

The legacy media failure uses its own declared rows and reaches rendering:
Anthropic/OpenAI say blob rendering is unimplemented; Gemini refuses local-path
blobs in favor of URI. This is an actual focused0/10 result, not a new full grade.
Original scores and past enhanced scores remain intact. No inherited whole-book
sweep or paid freshness run was needed.

Both browser failures reproduce in an isolated unchanged archive of
`5d83c609ddf864f03ab7eedb800cb39e4e1bf034` with its original model rows/tests.
The actual GUI emits four replies where these earlier fixtures expect one,
consistent with the later fleet idle-correction path. Baseline cache fails its
session-token assertion (80/320/8 versus20/80/2); corrected Sonnet pricing exposes
the same extra work earlier ($0.0003 versus$0.0001). Both speech runs fail the
same whole-final assertion with four copies rather than one. These are recorded
failures, not clean browser passes. No fleet behavior or those older one-reply
expectations were repaired under this model correction. The actual listener
pending-sleep/EOF controls pass. Root approved keeping this unrelated limitation.

Existing media/replay/lifetime sensitivity evidence remains applicable because
those checks and runtime paths were not changed. The new model selection and
pricing expectations are exercised by the real boundary tests above; no new
framework or invented protocol was added.

## Chapter14 changed launch

Author's corrected launch explicitly sets `LLM_VENDOR=anthropic` and
`LLM_MODEL=claude-sonnet-5`. On exact manuscript SHA-256
`f786497ecda5dc700d8a2456fc86cea0239e2b88585b9f14876df4e574f7f802`,
the published setup/launch passed7 turns and16 requests through the existing
external HTTP fake, with six correlated successful reads, recorded cuts,
handoff/unload and normal exit0. All16 requests name Sonnet5. Exact launch and
request captures are in `samples.json` and `samples-requests.json`; unchanged
Go excerpts retain their earlier Chapter14 evidence without rerunning them.
The initial attempt stopped at its manuscript-hash guard after an author prose
edit; `ch14-samples.txt` preserves that setup failure. No provider call occurred.
