# Chapter 16 coordinator decisions

October 8, 2026. The coordinator read complete outline/research `d4b17a3` and
independent advisory `0320911`. These are working engineering choices under
Bill's autonomous second-edition authorization, not new Bill rulings or evidence
of an implementation. They release a full draft, followed by independent review.
The author may flag a contradiction before writing an affected contract.

## D1. Capture a deliberately selected corpus

Accept explicit read-only external collections, frozen logical membership and
atomic idle refresh. They are an additional declared input dataset, not access
to retired conversation, hidden memory, the live store or derived exports.
Resume restores the accepted corpus without scanning current paths. A relocated
physical root affects a later explicit refresh, never historical reconstruction.

Use at most 16 logical external sources, 256 Markdown files, 2 MiB per file,
8 MiB aggregate retained text/attribution and 4,096 chunks. Define the counted
representation and bound every identifier, relative name and header in the
draft. Retain each current source body only once; index postings and chunks
reference that accepted representation rather than maintaining a second archive.
Existing event/canonical-state limits apply independently after escaping.

External source categories are `archive` and `project`; dynamic offered Skills
form the separate `skills` category. More directories do not buy more diversity
slots. Require deterministic recursive Markdown enumeration, bounded reads and
explicit symlink/nonregular refusal. Print exact inclusion rules, file identity
checks and the captured-set limitation; do not promise a cross-file filesystem
transaction. A missing/unreadable root refuses the entire refresh. A valid empty
root can replace the corpus with empty content. Unchanged bytes are a no-op.

Prepare the complete bounded derived index before atomically publishing the new
corpus revision and index together. One idle operation excludes memory refresh,
compression and active recall. Keep controls/cancellation responsive and joined.
Do not add an independent refresh queue or turn-time watcher. The proposed
30-second total refresh deadline is accepted.

## D2. Capability and current policy remain distinct

Accept an explicit fresh v6 profile extending v5, exact v1–v5 routes and disabled
defaults in the next strict policy version. Logical source IDs/categories/parser
version belong to creation identity; physical paths and credentials do not.
Enabling recall on an incapable profile refuses before mutation. Recall enable
and judge opt-in are separate. Neither silently enables context maintenance.
Initially the judge uses the captured foreground Engine route. No second model
router or historical recommended model ID is needed.

## D3. Admit the prompt, then anchor recalled material

Preserve normal human-turn admission and its recorded prompt. Asynchronous recall
accepts a later durable fact whose logical anchor places its material immediately
before that prompt. Explain durable sequence versus render placement with a
literal event/request example. Preserve earlier hints/manuals/memory ordering and
the later observation-collection boundary; never backdate a fact.

Accepted recall survives tool continuations and later turns until a successful
v6 handoff atomically retires the complete current recall set with the note/cut,
including when no tool pairs remain. Failed or canceled handoff retires nothing.
Checkpoint, corpus refresh and disabling future recall do not remove it. Exclude
recall from ordinary-dialogue compression membership; count it toward pressure.

Bound represented recall to 64 KiB of complete rendered blocks and 64 attachment
records. Exhaustion gives a visible no-attachment disposition, with no automatic
cut or paid recovery. Teach enabling context explicitly when necessary, then
using the existing zero-model public handoff. No blocked model is required to
perform the only cleanup operation.

## D4. One judge attempt with an honest shared budget

Run once per admitted human turn. Use only its human prompt as the query;
bound the scalar-safe query prefix to 8,192 UTF-8 bytes and report truncation.
Hints, continuations, rendering and maintenance helpers do not trigger recall.

At most one judge attempt, with a 15-second total operation deadline, consumes
one captured effective request slot. Raw zero still means 16. Reserve one slot
for the next foreground attempt; if only one remains, select mechanically.
This does not guarantee an answer, future tool rounds or later helper budget.
Record attempt admission before transport and accepted producing usage once,
including valid billed responses whose verdict is unusable. Never manufacture
zero usage for an unknown result or reset the cap after settings changes.

Apply inherited local maintenance/pressure checks before paying for the judge.
If the admitted prompt/current projection is already terminally over its bound,
finish memory_pressure without judge HTTP. Admit only whole prospective recalled
chunks that fit the remaining known projection allowance. If none fit, attach
nothing and report that reason. Recheck ordinary foreground pressure afterward;
later observations still have their own bound and do not support an answer
guarantee. Print the ordering relative to existing memory maintenance precisely.

Capture recall policy with the turn: later policy edits affect future turns,
not its budget or selected mode. Idle refresh prevents corpus changes during
that operation. Capture/recheck dynamic Skills eligibility; if it changes,
discard the stale selection without a second paid attempt. Retain billed usage.

For pause, a completed owned candidate may wait only within the same operation
deadline. Expiry discards it with a visible no-attachment disposition; no material
attaches while paused. Interruption/close cancels and joins work, fences late
completion and produces no mechanical fallback afterward. Ordinary remote or
format failures may fall back only while the turn is still eligible and active.
Teach these dispositions separately from cancellation of the entire turn.

## D5. Specify one calculable ranking and verdict

Use global corpus statistics for the printed smoothed BM25 variant, k1=1.2 and
b=0.75; diversity supplies source representation. Do not use independent source
statistics and an unexplained cross-source threshold. Specify a fixed Unicode
letter/digit tokenizer, simple lowercase mapping without normalization, an exact
small stop list and unique query terms. Zero/stopword-only queries select nothing;
only positive scores qualify. Print the complete formula and calculated fixtures.

The candidate pool is at most 20, the final selection at most three. For each
selection, reserve one best eligible chunk from each nonempty category, then fill
remaining slots by global score. Ties use the printed stable identity order.
After membership selection, order by score/stable identity for numbering. Explain
candidate diversity and final mechanical diversity separately. Deduplicate before
quota filling; directories within a category do not receive independent quotas.

Judge grammar is exactly one complete JSON array of unique one-based integer
indices into those candidates, at most three members. Preserve its order; [] is
a successful empty verdict and is never refilled. Reject prose, strings,
duplicates, fractions, range errors and partial parses visibly. No quota pass
overrides a valid verdict. The parser owner supplies the sole format instruction.

Bound the complete encoded judge request to 256 KiB, reply to 8 KiB and requested
output to 256 tokens. Quote query/corpus as data with no tools or observations.
If the bounded request cannot be encoded, report local size fallback before
transport; never send an oversized request merely because candidate count fits.
Print exact provider requests and fallback semantics rather than relying on a
library default or an observed model's tolerance.

## D6. Suggestions do not confer Skills authority

Only currently discoverable IDs/public descriptions enter new skill candidates.
Never index hidden definitions or unactivated manual bodies. Recheck eligibility
before attachment. Previously eligible accepted suggestions remain attributed
historical/as-of data until ordinary handoff retirement; they do not disappear
from history merely because a later offer changes. Actual loading always checks
current discovery and grants. Test never-offered secrecy separately from this
historical retention rule. Dynamic offers are not external corpus revisions.

## D7. Keep selected evidence; retire duplicate bodies

Chunk at scalar boundaries with at most 1,024 UTF-8 body bytes; print deterministic
Markdown/chunk boundaries and stable IDs. At most three whole chunks may enter
one complete rendered block of 6,144 UTF-8 bytes including every attribution and
wrapper. If the remaining byte/pressure allowance cannot fit a whole selected
chunk, skip it with a visible bounded disposition; never silently cut its text.
Exact decoded whole-text equality governs deduplication against named currently
visible text domains, never hidden or retired data or an ever-seen global set.

Refresh replaces the old applied corpus/index. Keep selected text while still
represented even if its source revision has been replaced. Reduced snapshots
retain only current corpus, represented recall and bounded identity/hash/size/
usage/disposition metadata for retired work. Retired candidate lists, judge
prompts, raw outputs or normalized data must not smuggle old bodies back into
state. Full logs retain actual receipts. Missing historical inputs give the
inherited history_unavailable response; rendering never searches or pays.

Before full-draft review, print exact v6/events/state/policy grammars, candidate
authorization and complete-set validation, counters/overflow, error codes,
ownership/lifetimes, public/CLI/browser controls and source-bound acceptance
fixtures. These decisions authorize drafting, not checker/student release or a
claim that recall improves task quality. Actual all-provider demonstrations and
the new/old comparison remain required later.
