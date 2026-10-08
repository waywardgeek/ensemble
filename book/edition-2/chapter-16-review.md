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

## Full draft review: f097e24

October 8, 2026. Independently read the complete manuscript, outline and evidence
at `f097e24a0b10f8293c68fb7758a9a8d90e7a0153`, and the complete coordinator
decisions at `81e2065`. The three author files and decision file still match
those commits. This follows the advisory above; it does not replace its chronology
or claim a new cold student attempt. Earlier Chapter 8 implementation and later
grader/reviewer roles remain disclosed. No Chapter 16 runtime exists to review.

The entire current `book/voice.md` and chapter-writing procedure were reloaded;
an initially truncated combined tool result was followed by separate complete
reads. The architecture ledger was read completely during the immediately
preceding task. Focused predecessor reads covered Chapter 14 §14.6, Chapter 15
§§15.2–15.3 and §§15.5–15.8, and Chapter 9's description/discoverability rules.
Those supplement the earlier complete Chapter 14–15 contract reviews. Historical
story verification revisited old Chapter 17's opener/TL;DR and the full commit
message `a91362b`; no old implementation or historical grader was opened or run.
The author's other historical reads are not attributed to this reviewer.

### R1: Give recall events an explicit physical record class

Material teaching gap in §16.7, read together with §16.1's explicit standalone
capability and §16.2's 8 MiB corpus. The draft requires inherited encoded-record
checks but never assigns the new `recall_*` records a standalone physical limit.
Chapter 15 grants the 64 MiB exception specifically to new memory events while
retaining the smaller bounds for unrelated standalone records. It does not
already classify a new recall corpus record. A student or checker therefore
has no printed common answer for reading a valid multi-megabyte standalone
corpus event, including escaped text.

Print the intended class explicitly: all new recall events are bounded to
64 MiB including their actual framing LF in recall-capable sessions and explicitly
capable standalone histories. Preserve existing bounds for unrelated events and
headers. Bound original physical reads before allocation and exact prepared
write bytes before append; a remarshal is not the size of an imported record.
Retain the lower corpus/helper/block limits as independent restrictions, and
distinguish controlled candidate refusal from terminal storage failure. Include
the classification and exact/+1/LF/escaping controls in the required check table.
Root independently agrees this is material and supports that scoped direction;
it remains a proposed clarification until the author publishes it.

### R2: Describe the indexing proposal accurately

Small evidence correction, not a runtime design issue. The full-draft reconciliation
in the outline and evidence says a preceding zero-based proposal was superseded.
The frozen `d4b17a3` outline's D5 leaves the index base unspecified. It requires
unique in-range integer indices but does not choose zero-based indexing. Say that
the full contract settles the previously unspecified base at one, unless a
different dated source for that particular proposal is supplied. The printed
one-based grammar and literal fixtures themselves are consistent.

### Retained decisions and integration assessment

| Decision | Independent contract assessment |
|---|---|
| D1 corpus authority | Explicit frozen logical roots, no construction/resume scan, bounded enumeration, captured-file limitations and ready index before publication preserve one applied authority. Retired memory and exports remain excluded. R1 supplies the missing physical record classification. |
| D2 capability/policy | Fresh v6 and policy v4 preserve old strict routes and disabled defaults. Logical membership is immutable; physical relocation and current settings remain separate. No handler/catalog expansion is implied. |
| D3 placement/lifetime | Recorded prompt precedes the later recall fact while the literal render anchor places the attachment before it. Complete-set v6 handoff retirement happens atomically, including empty tool sets; checkpoint, refresh and disable preserve accepted material. |
| D4 operation/cost | Actual parent paths, one joined worker, 15-second total deadline, one attempt, captured effective turn cap and foreground reservation are explicit. Current pressure stays current. Paused/stale/canceled outcomes cannot attach or retry after their boundary; accepted billed usage remains visible. |
| D5 ranking/verdict | Global post-filter statistics, unique query terms, two separate diversity selections, stable ties and strict one-based verdicts are calculable. Empty success differs from fallback, no fit and cancellation. Literal helper requests omit authority, tools and observations. |
| D6 Skills | Only currently discoverable IDs/descriptions enter new candidates; revision recheck prevents stale attachment. Historical suggestions survive offer revocation as attributed data without granting current load authority. |
| D7 retention/bounds | Whole chunks, complete wrappers, cumulative limits and actual neutral-projection delta are independent. Handoff prevents an attachment's prompt from becoming a compressible closed segment while that attachment still remains represented. Reduced state drops retired corpus/helper body copies and preserves the inherited history_unavailable boundary. |

The public, CLI and browser paths share Actor authority and owned safe state.
The exact refresh command/ack, policy patch, watch-before-ack ordering and Page
remount cleanup are printed. The separate optional GUI/MCP/native-speech
lifetimes remain intact. No second recall authority, automatic observation
access for the judge or new model-facing recall tool is introduced.

The spin is appropriately pending. It teaches explicit refresh/enable/judge
selection and the context-enable step required for zero-model handoff recovery.
It distinguishes selected bytes, delivered request and answer quality; its later
paid feature matrix still must allocate concrete actions and call caps. No
successful runtime, relevance improvement, cache effect or actual provider
demonstration follows from this prose review.

### Voice, facts and literal checks

The filing-cabinet incident retains the wasted investigation and the reader's
reason to build recall. It is correctly attributed to the old account, while
the unsupported improvement claim is declined. The judge-format story matches
the cited commit message and explains why parser ownership and visible fallback
matter. Concrete port-decision, refresh and delivery examples carry the mechanism
without inventing a new observed result. The dense contract earns its place;
the long-span and negation warnings are not a reason to add filler or remove
necessary distinctions. No additional voice blocker was found.

Independent local checks on the unchanged manuscript:

- All nine JSON blocks parse. The three provider helper fixtures have identical
  system/user text, and the inner user text has the specified canonical ordering.
- The selected-text SHA-256 matches. All three attachment fixtures render the
  exact same prescribed block, 221 UTF-8 bytes for the illustrated identities.
- Direct BM25 calculation yields `0.5981864372218454` and
  `0.42081720292932145`, with zero for the unrelated chunk.
- The versioned [Lucene 10.2.2 primary documentation](https://lucene.apache.org/core/10_2_2/core/org/apache/lucene/search/similarities/BM25Similarity.html)
  confirms the cited defaults and smoothed IDF definition; it supplies no
  application-quality claim.
- The retained `edition2-lintprose` executable reports 6,044 words and no hard
  failure. Its 18 negation forms and long technical-span warnings were reviewed
  manually. No Go build was performed.

The checked manuscript SHA-256 is
`a138e632340651b25608a81c748768bcc0e9d605904c18e1d3b52e6b9f268361`.
Full-draft disposition: D1–D7 direction supported; close R1 and the small R2
record correction before contract acceptance. This review makes no implementation,
checker-coverage or paid-run acceptance claim. The separate pending Chapter 10
bounds fixtures remain uncommitted and unexecuted under coordinated build scheduling.

## Grouped correction closure: 8adb4b5

October 8, 2026. Read the complete author delta from `f097e24` to
`8adb4b5038516ed50818778cad23a80da647c70b`, including manuscript, outline and
evidence, and revisited its affected initialization/record/disposition passages.
Full voice/procedure remain loaded from the full review without compaction.
The manuscript matches the correction freeze.

R1 is closed. Section 16.7 gives all six recall types a 64 MiB physical limit
including an actual LF. It separates v6 session identity from standalone's
prescribed construction-only empty recall initializer after memory initialization;
an arbitrary interior name cannot select the class. The bounded recognition/read
rule preserves original bytes, and write admission measures the exact prepared
record. Lower decoded limits, controlled recall_limit and terminal storage
semantics stay distinct. Unrelated/header limits remain inherited. The new
required-check row covers classification, exact/+1, LF and escaping; these are
future tests, not runtime results established by this closure.

R2 is closed. Both reconciliation records now correctly say the previous outline
left the index base unspecified. The evidence explicitly attributes the earlier
zero-based statement to author reconciliation error. It does not rewrite that
error into a historical design decision.

Every fenced literal remains byte-identical to the initial full draft; all nine
JSON fixtures parse. The retained hard linter passes at 6,228 words. No new
contradiction or voice blocker was found in the scoped change. D1–D7, the printed
literal fixtures and pending-spin distinction remain intact. The full contract
and prose review is accepted on this freeze, subject to the separately pending
checker/student/runtime/live/comparative gates. No code, provider call or Go
build was needed for this closure.
