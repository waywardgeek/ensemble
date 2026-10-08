# Chapter 16 outline: Recall before asking again

Through-line stake: a reader wants the Agent to use an earlier project decision
without having to remember to request a search, while keeping the source,
retrieval cost and limits of that decision visible.

Preparation, October 8, 2026. New Chapter 16 maps to first-edition Chapter 17,
“Auto-Recall,” under the current workflow. This is an author outline for design
review, not a student contract or an implementation claim. Chapter 15's grouped
contract is `a80138d`, independently closed at `738fe0f`; its implementation and
actual demonstrations remain pending. The decisions below require review before
a full Chapter 16 draft. No runtime, grader or provider work accompanies this
outline. [Research and source limits](chapter-16-evidence.md) are author-only.

## Human story and teaching order

Keep the historical account of an Agent that could search its saved knowledge
but seldom did. In old §17.0, Bill describes repeatedly seeing decisions
re-derived, fixed bugs investigated again and rejected architectures proposed
again. The knowledge existed; choosing to look for it was the missing action.
Attribute the account rather than inventing a new conversation or claiming a
measured intelligence improvement. The filing-cabinet comparison can survive in
a short opener, without the old performance superlatives.

The stronger diagnostic incident belongs beside the judge contract. One
component requested comma-separated numbers, another parsed only a JSON array.
The model could follow an instruction and still make every request take the
fallback path. Commits `a91362b` and `81510ac` document that failure. Give the
reply grammar one owner, make fallback observable, and distinguish a valid
empty selection from an invalid answer. A judge that was called has not
necessarily judged anything the foreground request used.

Proposed teaching sequence, with a concrete consequence in each section:

1. The question arrives before the reader remembers which old note answers it.
   Introduce passive retrieval as bounded attributed material, separate from
   instructions, tool rights and current conversation.
2. Select the corpus explicitly. Explain why Chapter 15's retired sources and
   hidden memory are not an archive that this feature may silently reopen.
3. Build an owned, bounded index from captured text. A changed file affects a
   later refresh, not replay of yesterday's question.
4. Rank chunks using a printed BM25 variant. Show a tiny corpus and a repeated
   generic word so the score formula and threshold can be understood together.
5. Apply diversity to the final mechanical selection. Use a fixture where the
   less numerous source loses without the quota, rather than merely displaying
   a quota configuration.
6. Let an optional stateless judge select numbered candidates. Show the format
   mismatch incident, exact accepted grammar, paid budget and visible fallback.
7. Record the selected bytes once on the shared Actor path. Explain continued
   tool rounds, handoff retirement, interruption and restart before cache talk.
8. Exercise human CLI, browser and public consumers. A private marker absent
   from the prompt must reach the actual foreground request through recall.

A worked project question should concern a previously rejected configuration
choice. The corpus contains the reason for rejecting it and an irrelevant note
with more repeated query terms. Keep the answer out of the prompt, visible
memory, working note, automatic observations and other sources. Future receipts
must distinguish selection, delivery and the model's use of the reason.

## Scope and inherited boundaries

Retain passive lexical retrieval, bounded Markdown chunks, optional model
filtering, source attribution, duplicate suppression and recorded selected text.
No embedding database or autonomous search Agent is needed. The judge has no
conversation, tools, Skills grants, memory refresh or MCP collection loop.

The old chapter's phrase “until a checkpoint removes them” cannot survive:
Chapter 10 checkpointing saves state without editing it. A recall lifetime must
instead name a new, explicit context transition. Likewise, the old archive of
daily files is not supplied by Chapter 15, whose exports are derived artifacts
and whose reduced snapshots deliberately lack retired bodies.

The new chapter must preserve Chapter 9's discoverable set and fixed ceiling,
Chapter 10's accepted-record bytes and exact-number rules, Chapter 12's captured
observations at the request tail, Chapter 14's logical hint/manual anchors and
Chapter 15's visible-memory and pressure rules. No retrieval result can change
any of those authorities. Browser status and controls remain public consumers in
the optional GUI module. Bill's required MCP-over-WebSocket GUI visibility and
replaceable core transport remain intact; recall neither hard-codes that
transport nor grants its judge a GUI endpoint.

## Proposed ownership and data plan

| Work or state | Proposed owner and actual parent |
|---|---|
| Creation corpus selection, logical source scopes and profile | Agent creation configuration; physical roots remain caller-selected |
| Applied corpus revision and accepted recalled entries | Existing Actor/reducer authority, using shared common values |
| Bounded corpus refresh and derived index | Agent-owned Recall spoke; index references one accepted corpus revision |
| One turn's retrieval/judge operation | Actor-owned operation, with its actual parent interface and cancellation scope |
| Chunker, ranker and judge parser | Recall behavior; operation children reach owners and logger through common interfaces |
| HTTP and usage, including judge attempts | Existing Engine reached through Agent; current credentials never enter corpus/history |
| Durable admission | Actor through Agent to EventLog, which owns the sole append descriptor; SessionStore owns lock/checkpoint work |
| Browser display, drafts and command completion | Optional GUI using public state/commands; no second recall authority |

A derived index is not a second copy of the conversation. Index token counts,
postings and ranking caches may be rebuilt from the currently applied corpus.
The accepted corpus and selected recall bodies have separate purposes: the first
is an explicitly retained input dataset; the second records what a request
actually received. A corpus refresh must not silently rewrite earlier entries.

Use owned immutable candidate values across worker completion. Actor checks the
turn, corpus/policy/Skills generation, current pause and remaining budget before
admitting a judge or accepting a result. It remains responsive to hints and
interrupts. No package globals, sibling-service injection or detached callback
bags are needed. Concurrent refresh, compression and recall must have a printed
admission rule rather than each starting another unbounded worker.

## Consequential decisions before the full draft

These are author proposals for coordinator/reviewer resolution, not Bill rulings.
They identify the choices that change behavior rather than private API spelling.

### D1. What corpus is recall allowed to search?

Prefer explicit read-only external collections with stable logical source IDs,
selected at creation, then atomically captured by a public refresh. Categories
can distinguish archive notes and project documents without pretending that
old daily logs, handoffs and learnings exist in the new implementation. The
application may deliberately select a separately maintained archive; it does
not obtain old EventLog bodies, generated-memory exports or private directories
by default. Refuse overlap with the live store/derived export tree.

Preserve the applied captured corpus in reduced state so resume can reproduce
future ranking without silently scanning today's files. External changes require
a later explicit refresh. This is a declared retained corpus, not an incidental
archive of retired conversation. A fresh corpus starts empty. Propose a complete
refresh bound of 8 MiB decoded text, at most 4,096 chunks and fixed per-file/read
limits, with independent inherited encoded record/state caps checked before
commit. Exact bounds and symlink/recursive-directory rules must be printed in
the draft, including empty/missing/partial-read behavior.

Proposed source/read contract for that review:

| Boundary | Candidate rule |
|---|---|
| Source identity | Frozen logical source ID/category and relative file name; a digest binds exact captured file bytes. Absolute host paths never become recall labels or session identity. |
| File/version | UTF-8 Markdown only; changed accepted bytes advance the corpus revision. Chunk identity includes source/file identity, captured digest and deterministic chunk ordinal. No mtime-only equality. |
| Read lifetime | Explicit idle refresh owns one cancelable scan with a 30-second total deadline. Proposed maxima: 256 files, 2 MiB per file, 8 MiB total retained text, 4,096 chunks. All limits apply before unbounded allocation. |
| Partial failure | Missing selected root, unreadable/invalid file, symlink or bound failure refuses the complete candidate; previous applied corpus/index remain usable. An existing empty root is a valid empty selection. |
| Commit | Actor accepts one complete bounded corpus revision before publishing/rebuilding its derived index. No partial per-file refresh; encoded-event and snapshot growth checks precede mutation. |
| Freshness/cache | Index is a derived cache of that revision, with no TTL, watcher or turn-time filesystem read. Refresh unchanged bytes is a no-op. A user sees the applied revision and explicitly chooses when to replace it. |
| Resume/replay | Restore accepted corpus, then rebuild the index from it without external I/O. Physical root selection is used only for a later explicit refresh. Historical requests use their recorded selection rather than today's index. |

The draft must make recursive enumeration/opened-file race handling precise. A
same-content corpus at a different host path is compatible; that does not grant
access to additional logical sources. New source membership requires a fresh
creation profile under this proposal. Root's interim review supports a separate
explicit corpus authority while leaving the final format and mechanism open.

An alternative is an explicitly volatile index rebuilt from caller inputs on
resume, but it makes the next turn depend on an unrecorded external change.
Avoid that route unless its weaker restart promise is intentional. Do not index
hidden memory or retired sources merely because the files still exist.

### D2. How do creation identity and policy evolve?

Prefer an explicit fresh v6 recall-capable profile extending v5, retaining exact
v1–v5 compatible construction/resume routes. No silent default tool installation,
primary-manual edit, version conversion or old-catalog migration. Stable corpus
scopes, source categories and parsing version belong in creation identity;
physical directories and credentials do not. Freeze the selected source set.

Propose recall initially disabled, with explicit enable and separate judge opt-in
in a strict next policy version. A user who turns recall on sees whether the
mechanical or paid path is selected. Reading an older policy supplies disabled
defaults without writing or advancing its revision. An enabled incompatible
policy cannot silently widen an older session. Decide whether a separate current
judge route is needed; prefer the current foreground Engine route initially,
without hard-coded historical “cheap model” IDs or credential fallback.

### D3. Where do recalled entries live and retire?

Prefer their own attributed durable kind, inserted at the admitted human-turn
boundary before its prompt. Preserve complete call/results and inherited H/S/P
anchors. They remain represented across tool continuations and later turns until
a successful handoff retires the current recall set. The new handoff extension
must retire it atomically with the note/cut in the v6 transition, including when
there are no tool pairs. A refused or interrupted handoff retires nothing.
Checkpoint, refresh and policy disable do not remove already accepted entries.

Recall is excluded from Chapter 15's ordinary-dialogue compression sources, so
it cannot be laundered into memory through a segment membership accident. Its
bytes count toward pressure. Propose a bounded cumulative represented-recall
budget with a visible no-attachment disposition when full; no automatic cut or
extra model call. Clearing through a handoff makes later recall possible again.

A per-request-only capture would simplify retirement but changes the historical
feature's cross-round lifetime. This is a real preference choice. Do not justify
persistence with an unmeasured provider cache-hit or billing claim.

### D4. When does work start, and what does it cost?

Run once per admitted human turn through Actor, shared by CLI, browser and public
calls. Hints, tool continuations, reconstruction and compression helpers do not
start new recall. Apply existing pause/admission gates; terminal memory pressure
or an exhausted request allowance must not first spend on a judge.

Propose at most one judge attempt with a 15-second absolute deadline and no
unattended retry. It consumes one of the turn's captured effective model-request
allowance, where raw zero means 16. Reserve one remaining request for the
foreground answer: with only one slot, use mechanical selection without a paid
judge. This proposed reservation is a scheduling rule, not free helper usage.
Account accepted usage once by actual producing route; unknown usage is not zero.

For the reader asking about the rejected configuration, a judge that consumes the
last allowed request would leave the useful note selected but no answer. The
proposed reservation makes that tradeoff visible without increasing the budget.

An ordinary remote/format failure may fall back mechanically with safe status.
Interrupt/close cancels the operation and fences late completions; it must not
attach a fallback after cancellation. Pause or changed eligible-source generation
needs an exact disposition, preferably discard/recompute without paying twice
for one turn. A settings edit cannot reset the captured turn budget. Unfinished
paid operations retain the inherited fail-closed live-resume boundary.

### D5. Ranking, quotas and judge grammar

Propose the smoothed BM25 IDF variant, k1=1.2 and b=0.75, separate statistics per
source, deterministic stable-ID tie breaking and a fixed printed tokenizer/stop
list. [Lucene 10.2.2's documentation](https://lucene.apache.org/core/10_2_2/core/org/apache/lucene/search/similarities/BM25Similarity.html)
supports that variant and those defaults; it does not establish useful thresholds
for this corpus. The draft must calculate a small fixture rather than import the
old claim that a certain number of filler chunks always makes a query work.

Start from the historical 20 judge candidates and three final snippets. Decide
thresholds alongside literal score fixtures. Apply a defined diversity allocation
to the final mechanical selection and redistribute unused slots deterministically.
A valid judge selection stands without a second quota pass. The candidate set
must also be bounded and genuinely allow multiple sources to reach the judge;
state that separately from final-result diversity.

Prefer one strict complete JSON array of unique integer candidate indices, in
range, with at most three members; [] is a successful empty verdict. Reject prose,
unknown structure and partial parses visibly, then fall back. The old parser
later allowed some number-list variants; retaining that tolerance would need an
explicit finite grammar and controls, never digit scraping. The component that
parses the verdict owns its only format instruction. Archive text is quoted data
inside the judge request and cannot add tools or alter that protocol.

### D6. What does skill discovery expose?

Index only currently discoverable loadable offers from the frozen Chapter 9
catalog, using their ID and public description. Do not expose hidden definitions
or load inactive bodies as if they were accepted skill material. Active manuals
already present in the request are not new recall. Loading remains an ordinary
explicit model management call or typed public operation, under the existing
ceiling and atomic graph rules. Revocation changes eligibility before admission.

This narrows the old “all skill docs” corpus deliberately. If full offered-skill
manual text is desired, settle its distinct data status before draft; it cannot
replace Chapter 9's dedicated material event or confer authority by retrieval.

### D7. Exact attachments and bounded history

Propose at most 6,144 UTF-8 bytes for one complete rendered recall block, including
attribution/wrappers, admitting whole bounded chunks in order. Oversize chunks
need a printed scalar-safe chunking rule before selection, rather than misleading
post-selection truncation. Dedupe exact text against currently represented recall,
visible memory and other explicitly defined text entries; avoid an unbounded
set containing every text ever seen. Changed source versions and partial matches
need literal tests.

Record source ID/version, chunk identity, selected text, mechanical/judged/fallback
disposition, query identity and producing attempt metadata. Corpus, operation and
material transitions need strict fields and reference validation, construction
admission and uint64 overflow checks. Public append cannot forge a pending
selection or fabricate a source refresh. Independent offline validation verifies
structure and references without repeating retrieval or HTTP.

Full logs retain real receipts. Reduced snapshots retain current applied corpus
and represented recall, plus bounded metadata for retired selections. They must
not retain retired text again through old judge prompts/responses or normalized
candidates. Exact reconstruction returns history_unavailable when required bytes
are absent. Pure rendering performs no search, refresh, judge call or consumption.

## Public behavior and acceptance preparation

The draft must print exact typed operations for refresh, state inspection,
policy changes and current ranking preview, if a preview is included. CLI and
browser commands must use those operations; proposed human spellings are
`/recall`, `/recall refresh`, `/recall on|off` and a distinct judge toggle. These
are new commands, not claims about the predecessor. A preview is zero-model and
read-only. Safe watch state exposes revisions, bounded counts, enabled/judge
state, current operation and last disposition, excluding corpus text and paths.

Before handoff, specify all exact event/policy/v6 shapes, source identity rules,
whole-refresh errors, tokenizer/chunker, quota arithmetic, attachment wrappers,
three-provider literal render/judge fixtures, browser acknowledgments and
completion ordering. Keep inherited ordinary WebSocket message limits and owned
DOM teardown. No recalled material autoplays as if it were an assistant answer.

The independent checker must cover the old ten behavioral/structural promises
plus new bounds, corruption, creation compatibility, Actor parity, usage,
interruption, pause, stale generations, pressure, Skills visibility, two-Agent
isolation and snapshot-only imports. Mutation controls must isolate ranking,
indexing, diversity, judging, cap and replay. The historical 8 KiB cap assertion
is insufficient for an exact 6,144-byte new boundary.

## Actual-spin plan, all outcomes pending

Before any paid run, create the normal feature/action/receipt matrix for all
three supported provider paths and every public/client seam. Freeze the initial
implementation and teaching review before comparative feedback.

| Path/action | Evidence required |
|---|---|
| Human CLI, each provider | Explicit corpus refresh and enable; real question, observed answer, actual request containing only the eligible selected markers, source/model/usage binding |
| Judge enabled, each provider | Real judge request and complete verdict, selected foreground bytes, separate judge/foreground usage; no promise that every model selects the intended candidate |
| Mechanical and negative controls | Judge disabled/no-budget path, actual empty corpus; deterministic malformed/error/timeout and exact caps labeled local faults |
| Continued work | Tool round keeps the same entry without another recall operation; subsequent handoff retires it; checkpoint alone preserves it |
| Public two-Agent consumer | Distinct corpora/rights, independent completion/partial outcome, refresh isolation and zero-call reconstruction with external roots/endpoints unavailable |
| Optional browser | Same Actor behavior, safe status and controls, pause/cancel/remount, actual screenshot with alt text; GUI tunnel and shared socket lifetimes remain unchanged |
| Skills and hostile notes | Hidden catalog names absent; offered skill suggestion cannot grant a tool; a planted instruction in retrieved data cannot itself mutate admission rights |
| Restart and retention | Current corpus and represented recall restored without scanning; changed external file only after explicit refresh; snapshot history limits shown honestly |

The original account is motivation, not a receipt for these runs. No new speed,
price, cache-hit, semantic-retention or “model remembered” result is asserted.

## Full-draft decision reconciliation

Coordinator decisions `81e2065`, after independent advisory `0320911`, release
chapter-16.md for full contract review. These supersede the corresponding proposals
above without rewriting preparation chronology. All execution remains pending;
current gates live in [Chapter 16 validation](chapter-16-validation.md).

- D1 chooses frozen external sources in archive/project categories, explicit
  bounded idle refresh and complete ready-index preparation before atomic
  publication. The draft defines source/version/read accounting and captured-set
  limitations; no external scan occurs on resume.
- D2 selects explicit fresh v6 extending v5, strict policy v4 with disabled defaults,
  current foreground Engine route for the judge and unchanged old-profile routes.
- D3 preserves normal prompt admission, then records recall at a later event
  sequence with a logical before-prompt anchor. Accepted material persists until
  the v6 handoff extension atomically retires all of it. Limits are 64 attachments
  and 64 KiB of represented blocks; checkpoint remains save-only.
- D4 selects one 15-second operation/at-most-one judge, captured effective turn
  allowance and foreground reservation. Current pressure is measured with the
  actual canonical neutral-projection delta; pause expiry/stale/canceled work
  settles without another paid attempt or late attachment.
- D5 replaces the outline's per-source statistics with global BM25, positive
  scores, one-per-category reserve then ranked filling and pool of 20/final three.
  It settles the previously unspecified index base with strict one-based verdicts.
  Query is only the human prompt,
  bounded to 8,192 bytes. Request/reply/output bounds are 256 KiB/8 KiB/256 tokens.
- D6 retains only current offered IDs/descriptions for new skill candidates;
  accepted suggestions remain historical/as-of after offer revocation.
- D7 specifies 1,024-byte whole chunks, 6,144-byte blocks, exact-text dedupe and
  explicit omitted-item reasons. Reduced snapshots drop retired corpus/helper
  body copies while preserving usage/identity metadata and history_unavailable.

The draft's public/browser commands, strict event/identity/policy shapes and
three-provider literals are new teaching before implementation. No tool, old
solution, current student source or grader was edited. The historical format
mismatch and unused-search story remain attributed; the actual spin is a reader
plan awaiting source-bound evidence rather than a fabricated successful run.

Independent full review `5138e53` identified a missing physical record class for
standalone recall facts. The grouped clarification gives all six new recall
events 64 MiB including actual framing LF, after valid profile initialization,
with bounded original reads and exact prepared writes. Unrelated event/header
limits remain unchanged. The review also corrected this reconciliation's prior
claim of a zero-based proposal: frozen outline `d4b17a3` left that base unspecified.
