# Chapter 16 consequential-choice advisory

October 8, 2026. Independent review of outline/evidence freeze
`d4b17a3dd12f488a9eead1f7541dd3779f60ade7`, before any full manuscript or
student implementation. The complete two files were read and match the freeze.
The reviewer previously coded Chapter 8, prepared independent checkers and
reviewed later contracts; this is neither authorship nor a cold student attempt.
Full current voice/procedure and architecture remained loaded without compaction.
Chapter 14–15 full reviews remain applicable; this pass revisited their handoff,
anchor, pressure and retired-receipt passages and relevant Chapter 9 authority.

For the historical motivation this pass read old Chapter 17's opener/TL;DR,
targeted lifetime/judge/correction passages, and complete commit messages
`a91362b61fcdd3316bd15069f28add0151892f37` and
`81510ac3996b5ac1275f88e7b48d82e4bc92ef44`. No old implementation or grader
was read or executed. The preparation evidence's other historical and external
reads remain the author's reads, not independent reproductions by this reviewer.

The proposed direction fits the preceding contracts. Explicit captured corpus,
fresh v6 capability, default-disabled recall, a stateless bounded judge, durable
attributed selections and handoff retirement are preferable to silently reopening
retired memory or redefining checkpoint. The following consequential choices need
coordinator resolution and exact teaching before a checker or student handoff.

## D1–D2: Corpus authority and compatible creation

Support a separately selected read-only corpus, atomically captured by explicit
idle refresh. It is retained input authority, not a search permission over all
files owned by the application. Keep the proposed 8 MiB/256-file/4,096-chunk
limits as explicit engineering choices, subject to the complete emitted event
and snapshot bounds. Printed rules must count attribution, relative names,
headers and retained source bytes as well as chunk text; otherwise many small
chunks can evade a nominal text budget through metadata. Bound source count and
path/header lengths too. Do not retain both unbounded original files and a second
unbounded chunk archive behind the same advertised 8 MiB limit.

Specify deterministic enumeration, recursive inclusion/exclusion and symlink/
nonregular behavior, plus what bytes constitute the captured candidate if an
external writer changes a file during scanning. Chapter 15's captured-set promise
is sufficient; a cross-file filesystem transaction is unnecessary. Existing
empty roots may commit empty membership; missing/unreadable roots refuse the
whole refresh. Resume uses applied corpus even if its former roots are now
unavailable. Path selection for the next refresh must not become an implicit
startup scan or destroy that usable recorded corpus.

Prepare the bounded derived index for a candidate before publishing its revision,
or specify an equally precise readiness gate. Committing a corpus revision and
temporarily searching a stale/empty index would break the stated authority.
Index caches remain revision-bound derived data and need no new durable archive.
Use the existing idle operation exclusion with memory refresh/compression; this
avoids a second queue of concurrent corpus changes during a recall operation.

Support explicit fresh v6 extending v5, strict next policy version with disabled
read defaults, and the preceding exact old-session construction routes. Freeze
logical source membership/category/parser version; allow physical relocation.
Define enabled recall on an incapable older profile as a visible refusal before
state mutation. No default catalog, primary or tool change should be smuggled
through the profile. Current Engine route is a sufficient initial judge choice;
a separate model/credential router would add another authority without resolving
the chapter's core problem.

## D3: Record order, logical placement and retirement

“Before the prompt” needs two meanings separated. Human-turn admission can
already have recorded the prompt before an asynchronous judge completes. Prefer
normal prompt admission, followed by an accepted recall fact with an explicit
logical anchor placing its material immediately before that prompt. Define its
relation to pending hints, existing manuals and current memory, while preserving
their preceding order. Do not backdate sequence numbers or represent an admitted
prompt as though it had not existed when interrupted. A different timeline is
possible, but its cancellation/retained-prompt semantics must be printed.

Support retaining accepted recall across continuations and later turns, excluded
from Chapter 15 compression sources. Extend the v6 handoff change atomically to
retire the current recall identities and install the note/cut, even when no tool
pairs remain. Failed/canceled handoff retains all of them. Checkpoint, corpus
refresh and disabling future recall do not retire already accepted material.
Printed complete-set/counter/bounds rules must prevent a public append from
retiring arbitrary entries or producing half of the transition.

Choose the cumulative represented-recall byte/count ceiling and a visible full
disposition before the draft. There is a usable recovery path already: explicitly
enable context if needed and perform the zero-model idle handoff. Recall enable
must not silently enable context policy. The demonstration must teach that step,
since fresh capability alone leaves context maintenance disabled. Do not ask a
pressure-blocked model to perform the only available cleanup.

## D4: A reservation is local scheduling, not an answer guarantee

Support one judge attempt per human turn, captured effective turn allowance,
15-second absolute deadline, no retry and one remaining foreground-request slot.
Raw zero continues to mean 16. State that the reservation permits the next
foreground attempt; it cannot promise a completed answer, successful tool batch
or budget for later memory helpers. Those still consume the shared captured cap.
Record judge start before transport and accepted producing usage separately from
whether its verdict is usable or attached. A valid billed response with invalid
verdict must not become a zero-cost transport failure.

The pressure ordering needs a concrete algorithm. Before paying, apply inherited
local maintenance and measure the pending prompt/current projection. If it already
exceeds the Chapter 15 bound, terminate memory_pressure without judge HTTP.
Also decide how prospective attachment bytes fit: a mechanically bounded selection
must not knowingly make the next request impossible just to pay for a judge that
might shrink it. Prefer admission of only whole selections that fit the remaining
projection allowance, with a visible no-attachment/pressure disposition when none
fits. Recheck ordinary foreground pressure after selection. Keep observation
collection at its inherited later boundary and make no remote-token-size claim.

Only the admitted human prompt should supply the initial query under the simplest
contract. Including earlier conversation is a distinct retention/size/privacy
choice and must be named and bounded, not borrowed from the old judge example.
Specify exact operation/policy/Skills capture times and the outcome of a revision
change. With idle-only corpus refresh, no corpus mutation can race an active turn,
so a paid discard/recompute loop is unnecessary. A stale verdict may retain billed
usage but cannot attach outdated material or trigger a second paid attempt.

Pause needs separate transport and attachment rules: does a completed candidate
wait boundedly for resume, or expire/discard under the operation deadline?
Either can be taught consistently. Never attach during a paused admission boundary
or fall back after user interruption/close. Policy cancellation of recall should
not be confused with cancellation of the entire foreground turn. Preserve joined
worker cleanup and the inherited unfinished-operation resume refusal.

## D5: Make ranking and verdicts calculable

Support an explicitly printed smoothed BM25 formula, fixed tokenizer/stop list,
stable ties and a calculated small fixture. Per-source statistics make scores
depend on each source's own corpus, so a cross-source threshold is not an intrinsic
relevance measure. Either retain that choice with a worked cross-source example,
or use global statistics and let diversity supply source representation. Avoid
the old unexplained corpus-padding threshold rule. Define query term repetition,
empty/stopword-only query, Unicode/numeric tokens, zero-length chunks and threshold
equality, rather than leaving them to a library default.

Define whether quotas apply to logical roots or semantic categories. Categories
avoid gaining quota by splitting one directory into many selected roots. Candidate
pool diversity and final mechanical diversity are separate allocations; both need
exact rounding, redistribution and tie rules. A distinguishing fixture leaves a
lower-scoring eligible category outside the final top three without that rule.
No second quota pass may override a valid judge selection.

Support a single complete JSON array of unique integer indices and a valid []
empty verdict. Print zero- or one-based indexing, whether verdict order is kept,
and exact invalid/duplicate/out-of-range/oversize behavior. Integer strings,
prose and partial parses should not be silently salvaged under this proposed
strict grammar. The parser's owner supplies the sole format instruction.
Bound the complete judge request, reply and output-token limit, including quoted
query/candidate/attribution overhead. A 20-candidate count alone is not a byte cap.

## D6: Offered Skills can become historical suggestions

Restrict new skill candidates to current discoverable IDs/public descriptions;
never index hidden definitions or unactivated manual bodies. Capture Skills
revision and recheck eligibility before attachment. The applied external corpus
and this derived dynamic offer set should have distinct identities: an offer
change is not a filesystem refresh or a new archived corpus file.

D3 means a suggestion valid when attached can remain after that skill is no
longer offered. Prefer declaring it attributed historical/as-of data until the
normal handoff retires it. Ordinary load authority still checks the current
discoverable set and ceiling. “Hidden names absent” then means never-offered
definitions are not disclosed by retrieval; it cannot mean every previously
offered name disappears from history. Automatically retiring old suggestions on
each Skills change would be a second lifetime rule and needs its own deliberate
atomic transition instead. Root has received this consequence for decision.

## D7: Bounds, deduplication and reduced history

Print a scalar-safe whole-chunk rule and bound the complete attributed wrapper.
One simple option is a per-rendered-chunk limit small enough that three whole
chunks fit 6,144 bytes; otherwise specify selection/drop order under the total
cap. Define whether duplicate removal precedes quota filling and judge numbering.
Exact decoded whole-text equality is simpler to audit than substring matching;
state which current visible entries participate and what changed source versions
mean. Do not deduplicate against hidden memory, retired history or an unbounded
ever-seen set. A valid empty verdict must remain empty rather than refill slots.

Refresh replaces current applied corpus; old versions must not survive only
because an index, normalized candidate or judge request still references them.
Retain exact selected text while recall remains represented, even when its source
version has been refreshed away. After recall retirement retain bounded identity,
digest, size, usage and disposition metadata; drop duplicate body-bearing judge
prompts/raw outputs/candidate arrays. Full logs retain the original receipts.
Snapshot-only reconstruction returns history_unavailable when required bytes are
absent, and offline validation must not fetch old corpus files to make it succeed.

Specify the v6 initializer, corpus revision/content facts, operation/attempt/
disposition and attachment/retirement schema, including all exact uint64 domains,
no-op/exhaustion and allowed references. Pure rendering uses recorded selection,
without ranking, consuming, refreshing or paying. Strict public append requires
the owned admitted candidate; an internally consistent external snapshot is still
not authenticated evidence of an actual provider call.

## Story, evidence and next boundary

The historical opener supports an attributed account of repeated investigation,
not a measured productivity improvement. The two scoped commit messages support
the conflicting judge-format instructions and unreachable Actor/GUI path as
repository findings. Preserve that distinction and the observable fallback
status in the new explanation. The old checkpoint-removal and all-skill-doc rules
are precisely the historical choices this new contract needs to change.

The proposed all-provider spin separates selected bytes, delivered foreground
request and actual model use of the rejected-decision reason. Keep the marker
out of visible memory, newest note, prompt, observations and fresh answer-bearing
reads. Actual empty verdict, fallback and quota tests need deterministic controls
in addition to real judge receipts. No outline, old score or lint result supplies
those runtime facts. A bounded feature/action/call-budget matrix and runnable
independent command remain prerequisites to later paid work/student release.

Reviewed hashes:

- chapter-16-outline.md: `d01596cb7097ec1c759b11065dd627d93351650572f87d2323bd6ec34194d3b0`.
- chapter-16-evidence.md: `a7634dffdc324d208a4e1d38af430ae96a7e9e489317cb68672ce18d7bbd5000`.

Disposition: proposed direction supported, with the concrete choices above to
be settled before the full contract. This is advisory review only. No new code,
checker, runtime/build, provider call or historical reference execution occurred.
