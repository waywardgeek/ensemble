# Chapter 18 preparation evidence

October 8, 2026. Author research and full-contract preparation for new Chapter 18,
originally mapped from old Chapter 19. Scope frozen at `3c91992` and independently
reviewed at `d61388e` is accepted by the coordinator. The new manuscript is now
reviewed at `2fc8af0`, with its three corrections published for closure and
[release gates](chapter-18-validation.md) still open. No runtime, grader,
credential, browser authorization or provider
inference work occurred. The initial research chronology below is retained.

## Reads and source boundaries

Read full current voice.md and chapter-writing-procedure.md for this task. Current
architecture and the complete Chapter 17 contract/closure are retained from the
preceding author task; its closure at `b06d780` accepts `b53c3c8` for preparation,
with checker/predecessor release gates still open. Updated Chapter 17's own
validation record accordingly, without claiming a student or live result.

Read all of first-edition chapter-19.md and chapter-19-coder-review.md, and the full
coder brief. The outline's combined output was truncated; focused follow-up reads
covered its reasoning-summary section, billing section and resolved choices.
No claim of a fresh full-outline read is made. Focused reads of the historical
openai_auth_design.md covered §§21–23, §42.10 and §42.16, rather than all of that
large design document. It is historical evidence, not a source of new authority.

Read the global map's old19/old20/old21 lessons, the ending chapter's placement and
publication-time epilogue instruction, new Chapters 6/7/13's stream/display/speech
clauses, and old Chapter 20's opener/TL;DR/first failure mechanism. Read the complete
old ch19_checks.go table and focused reasoning-summary check/meter sections from
ch19_run.go/ch19_harness.go. No grader or historical implementation ran, and no
withheld Chapter 10 comparative finding or current student implementation was read.

## Historical chronology and its limits

| Source | What it establishes for this outline |
|---|---|
| `36e5323` | First-edition Chapter 19 entered the repository on October 2; current chapter remains a historical artifact |
| `d50e9ca`, coder review | Reported summary probe near-miss, eventual observations and teaching corrections; sign-in itself had not yet been tested live in that report |
| `acfc2c6`, inspected stat/subject | Old implementation introduced summary-specific vocabulary/model data; this does not require a second new-edition channel |
| `3bfb494`, complete commit message | A later route-specific marker rejection contradicted the earlier plan-caching assumption; metered and plan tests needed separate controls |
| `ec5f086`, complete commit message | Historical plan streams needed different content-type/item completion handling and corrected call provenance; a reported live GUI call followed |
| `eb15005`, complete commit message | Deltas/finals used different IDs, causing duplicate display/speech; the fix reused identity. A separate absent-callback bug affected the recall judge |
| `e160591`, complete commit message | Later editorial repair corrected the old chapter's contradictory tool-array/cache-boundary sentence |

The coder review reports 314/186/4 summary deltas under different settings and a
first trivial probe with none. These are historical reported measurements; this
research did not recover their raw probe streams or reproduce them. Use the
qualitative near-miss with attribution, or locate stronger artifacts before
publishing counts. The existing voice source supports the accessibility stake;
there is no need to repeat a separately unverified listening-speed figure.

The old auth design's later §42.10 recommends mapping summaries into existing
DeltaThinking, whereas the early chapter/grader adds a separate reasoning-summary
kind. The new edition already has a thinking channel and speech source. Reusing
that owned path is the proposed design; old vocabulary does not decide it.

The historical plan-stream note at `ec5f086` describes terminal output missing
items present in done events. This is a source to investigate during Chapter 17's
actual route validation, not permission to weaken its now-reviewed terminal
agreement rule or to claim today's endpoint behaves the same way. Preserve the
old note and record any newly observed mismatch before a changed contract/code.

## Current primary documentation lookup

Used the already-read full OpenAI Docs skill: searched official summary/stream
topics, opened the matching pages, and read the focused schema sections. Retrieved
the complete short plan limitations and model/inference Markdown directly after
the general web reader included extensive navigation. Reads were unauthenticated.

- The [reasoning guide](https://developers.openai.com/api/docs/guides/reasoning)
  specifies explicit summary opt-in and model-dependent supported modes. It places
  exposed text in the reasoning item's summary array and distinguishes it from
  raw reasoning. This supports proposing off/auto without promising that every
  response contains a summary.
- The [Responses streaming reference](https://developers.openai.com/api/reference/resources/responses/streaming-events)
  defines summary part/text added, delta and done events with item_id,
  output_index, summary_index and sequence_number. Part-done can mark incomplete.
  Those fields require operation-scoped mapping and terminal agreement; merely
  counting deltas cannot prove timely delivery or once-only finalization.
- The [plan inference guide](https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference)
  specifies account-specific discovery and public streamed Responses with store
  false. The [preview restrictions](https://developers.openai.com/siwc/token-sharing-open-source/preview-limitations)
  do not list reasoning.summary among forbidden parameters. That is compatible
  with using the documented Responses option; it does not establish that a named
  account/model will expose summaries. No access or entitlement was checked.

The create-reference schema for summary/Direct/input forms was retrieved in the
preceding Chapter 17 task. This research does not infer summary availability from
a model name, streaming support, a historical table, or an absent restriction.
Before the full live plan, establish the selected model's current documented
support and discover its availability on each funding route. The eventual contract
must distinguish local unsupported selection, provider refusal, successful empty
summary, and actual observed summary delivery. None authorizes a paid fallback.

## Scope disposition

The coordinator accepted keeping a focused Chapter 18 direction, preserving the
following daily-use chapter and the final comparison. Credentials, PKCE, refresh,
funding selection, capsule replay and cache investigation stay in Chapter 17.
No automatic metered fallback, new credential store, remote model-steering API,
provider-hosted tool, or raw-reasoning extraction is proposed here.

The proposed addition is explicitly selected foreground summary display/listening
through existing owners. Helpers remain unchanged. A new versioned capture and a
safe derived summary view must be printed before implementation; the current
opaque final placeholder cannot simply expose its raw payload to make speech work.
The outline names that remaining representation work rather than privately deciding
an interface in advance of review.

The old broad vendor-policy/security claims and savings multiplier are not reused
as current claims. Their historical text remains intact. The original-author
stories in Chapters 9–17 are untouched. The new ending comparison and the
publication-time Codex first-person epilogue remain required, with no result
invented before the edition is complete.

## Initial scope verification (before full draft)

No executable chapter fixture or new grader command is published at this scope.
The acceptance table and 30-start ceiling are proposed review inputs, not evidence
of passing checks or authorized launches. The inherited seven-category grader
cannot establish the new stream/public/lifecycle guarantees on its own.

The existing external prose-lint executable passed every hard rule on these
preparation files and the Chapter 17 status update; scoped whitespace checks
passed. Short-file/density warnings remain appropriate for research notes, without
padding an outline into a manuscript. Next is coordinator review of the focused scope,
followed by the full student contract and independent review. No build, auth,
provider call, code change, push or earlier-edition edit is part of this task.

## Complete draft: source refresh and grouped decisions

Reloaded the entire current voice.md, chapter-writing-procedure.md and
architecture.md, and read the complete scope review, outline and evidence.
Focused predecessor reads covered Chapter17's exact v7 identity/semantic codecs,
capture/profile changes, capsule refs, terminal acceptance and helper limits;
Chapter6's public observation/fragment/byte bounds; Chapter7's snapshot and safe
projection wire; and Chapter13's speech source identity and strict journal records.
Reopened the complete `eb15005` commit message to check the historical incident.
No withheld Chapter10 comparative finding or student runtime was read.

Reloaded the entire OpenAI Docs skill and refreshed a concise official search,
then actually opened the reasoning guide, Responses streaming schema and plan
model/inference page. Focused page reads verified summary opt-in/model dependence,
the four indexed summary event shapes including part-done incomplete, and plan
stream/store requirements. The exact URLs above remain the supporting sources.
This does not establish account entitlement or measured summary availability.
The pages also mention other capabilities outside this task; none was imported
as a new remote-steering or effort feature.

The coordinator selected off as request-only behavior and preserved existing
owners. The author chose the explicit v8 wrapper and one-item indexed public
projection as the concrete proposal now under review. The public extension is
printed instead of assuming that Chapter7's exact opaque placeholder could quietly
gain raw content. It retains one Chapter17 capsule/ref and one final part, with
summary_index only on the new thinking-delta subtype. Public snapshots and final
views derive text from that same authority. Helpers remain summary-off and retain
their caps/received-byte boundaries.

The new transcript in §18.3 is a synthetic literal fixture, not a provider trace.
Its two reasoning items contain three summary slots, including an empty one;
another message supplies the ordinary accepted text position. Engine assembly
uses exact agreement when an observed prefix exists; done-only and terminal-only
sources are separately specified. The 1,024-entry bound is an explicit local
design choice, not a provider limit. S1–S4 dispositions are mapped in the outline.

No new speech journal version is introduced. The existing item part_id and distinct
utterance identities remain true to their original meaning; the chapter explicitly
states that the journal does not encode summary_index. Private normalization uses
the complete indexed source. Reviewer should check that the application-level
recorder controls distinguish subsummary finalization without treating every
same-part utterance as a duplicate.

## Full-draft verification and remaining work

Parsed all 23 JSON/JSONL values in the manuscript with the standard JSON parser.
Checked the two-item summary fixture, matching final views, empty-slot preservation,
auto-request field placement and 12+12+2+2+2 start allocation. The literal stream's
ten summary events have increasing sequence numbers 0–9; recipe item-done events
use 10–11 and terminal completion 12. No code/compiler or provider call was needed
for these author fixture checks.

The existing external linter is
`/Users/bill/projects/ensemble-edition-2-revisions/executables/edition2-lintprose`.
It passes the manuscript's hard rules; the soft person-gap warning was read as
an editorial prompt. The reader's concrete controls, duplicate speech consequence
and documented historical repair carry the technical stretch; no manufactured
scene or extra Bill mention was added to satisfy a counter. Scoped whitespace and
literal verification accompany the author freeze. No existing code/table/artifact
was rewritten by this new chapter.

The demonstration section remains a plan with exact actions and observables.
Native callback completion alone cannot establish hearing. Empty summaries,
provider refusals, late hints and lost interrupt races remain possible outcomes,
not successful demonstrations prewritten by the author. The 30 starts include
failures/continuations and have no hidden listener-model allowance. Actual account
access, selected identities and feature evidence remain required after the
coordinator releases implementation and launch.

Next: grouped independent contract review, coordinator acceptance, accepted
Chapter17 baseline and new checker invocation, then a fresh student. No runtime,
live, comparative or publication checkpoint gate is closed by these prose checks.

## Grouped full-review correction

Read the complete `2fc8af0` review and reloaded full voice.md and the chapter-writing
procedure. Focused predecessor reads checked Chapter7's error/watch envelopes and
Chapter17's public revision/status conventions. The coordinator accepted the
three finite corrections without expanding the chapter's feature scope.

The new GUI literals distinguish settings revision strings from inherited watch
watermarks. Inspection and no-op preserve both; one applied setting publishes
summary_status_changed before its correlated acknowledgment. Profile changes also
publish current safe status without incrementing the summary setting revision.
The fixed 30-start allocation now keeps read handlers visible within an existing
session and records unexpected tool outcomes honestly. Canary controls apply to
safe observation/display/speech audiences, preserving deliberately requested raw
exports and exact replay.

These are contract clarifications, not observed provider behavior or passing
implementation tests. The original 23-value fixture check above remains the initial
draft record; the correction adds five GUI JSONL examples. Reviewer closure,
accepted predecessor and the new checker command remain required before handoff.

The correction check parsed all 28 JSON/JSONL values and compared the GUI literal
status/revision fields: inspection at watch 40/settings 0, applied observation
before acknowledgment at watch 42/settings 1, and unchanged acknowledgment at the
same revisions. Scoped whitespace and existing-linter hard checks pass. The
manuscript is 5,317 prose words; its person-gap warning remains an editorial input,
not a reason to invent a scene. No compiler or runtime was exercised.
