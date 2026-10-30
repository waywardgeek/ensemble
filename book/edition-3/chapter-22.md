# Chapter 22: Auto-Recall

Chapter 21 built a memory system. The agent can save, compress, and
graduate memories through a cascade of bands. Enabled bands already put
recent memories in context. Older originals and unloaded material remain on disk, organized and budgeted, until something
brings them back. The agent can still repeat a rejected decision because
the explanation is sitting in an archive it never thought to open.

For that older material, the agent has a filing cabinet without a habit
of opening it.

In the spring of 2026, I had a coding agent with a full memory cascade
and a `search_knowledge` tool. It was episodically brilliant: when it
searched, it found the right thing. But it almost never searched. Every
session started cold. It would re-derive decisions we had already made,
re-investigate bugs we had already fixed, and propose architectures we
had already rejected. The knowledge was there. The recall was not.

The fix was embarrassingly simple. Before every user message reaches the
model, run a keyword search over the memory files, filter the results
through a cheap model, and attach the survivors to the conversation.
The agent never asks for the memories. They arrive the way a relevant
fact arrives when a colleague mentions a project name: unbidden, fast,
and usually right.

The improvement was immediate in my daily use. The agent
went from a goldfish that kept meticulous diaries to a collaborator
with a past.

The daily-use observations and tuning history in this chapter come from that earlier system, not fresh C# measurements. The C# evidence in §22.9 covers local component and application behavior and a scoped executable rubric; it does not establish live-provider relevance, entitlement, OAuth, billing, native Linux, audible speech, or human accessibility.

## §22.0 — TL;DR — The Prompt

You are building a C# coding agent that can use tools and skills. Chapters 1–21
provided durable History, an actor-owned scheduler, and transactional long-term
memory. Now implement automatic local recall before a new model turn.

1. **Keep the architecture visible.** Common declares the actual partial types,
   persistent fields, relationships, and callable operations. `Recall/` supplies
   retrieval algorithms. Each Agent strongly owns one `Recaller`; construction,
   removal, and destruction maintain both relationship ends. Recaller owns only
   derived immutable indexes. It does not publish History, schedule turns, or
   recover Memory’s transactions.
2. **Search five independent corpora.** Under the Agent home, scan `SOUL.md`,
   `MEMORY.md`, and Markdown under `memory/`, `bucket-0/`, and `bucket-1/`, including
   `.md.retired` memory files. Separately scan Markdown under `docs/`, `handoffs/`,
   and `learnings/`. The fifth corpus is visible, exact `SKILL.md` files from the
   explicitly selected Skills directory. Do not index an entire skill directory’s
   scripts and fixtures. An Agent without a data home has no implicit home corpus.
3. **Read fresh, within authority.** Build indexes lazily. Read current bytes on
   searches, fingerprint names and content, and replace changed indexes only
   after a complete successful refresh. Detect additions, removals, and same-size
   edits. Exclude hidden staging and outward targets. Memory serializes its scan
   with disk operations and refuses staged/committed recovery holds. Skills owns
   its separate lazy read-only root when the host selected a tree outside the
   workspace; this never expands Sandbox or file-tool authority.
4. **Chunk before ranking.** Split at `## ` headings, then `### ` for sections over
   4,096 UTF-8 bytes, then paragraph boundaries. Keep heading text and a
   breadcrumb with each snippet. Preserve an indivisible large paragraph; the
   final attachment has a separate hard byte cap. Blank files produce no chunks.
5. **Implement the actual BM25 formula.** Use `k1 = 1.2`, `b = 0.75`, independent
   corpus statistics, per-document frequency, and positive `log(1 + ratio)` IDF.
   Tokenize Unicode letters/digits, lowercase, and remove stop words only from
   queries. Unique query terms do not gain weight by repetition. Keep document
   stop words in lengths and average length. Test arithmetic and competing ranks,
   not merely positive scores.
6. **Apply quotas only at mechanical selection.** Reserve half the slots, rounded
   up, for memory; divide the remainder among secondary sources present in the
   pool. Fill unused slots globally by score. Do not discard the full pool at
   the judge candidate limit. Quotas must not overrule successful judge choices.
7. **Offer a stateless judge.** Default to the current model on the same provider
   and funding route; permit an explicit same-provider model name. Supply the
   query, recent ordinary conversation, and up to twenty numbered candidates as
   data. No tools, skills, or live Memory bands enter this request. Accept only a
   completed tool-free response. Parse JSON index arrays, including fenced arrays
   and quoted integers; ignore duplicate/out-of-range/noninteger elements. A
   successful empty array means no recall. On malformed output, transport failure,
   or the judge’s 15-second linked HTTP deadline, use mechanical selection over
   the full pool. Do not turn caller cancellation into fallback.
8. **Admit the turn in order.** Record the exact Human input, then run retrieval
   and optional judging in the existing foreground-worker slot while controls
   remain responsive. Return immutable data to the actor. Join and verify the
   active turn before publishing; recheck duplicates against newly retained hints
   or manuals. Record a separate `RecallAdded` event containing one `RecallPart`
   before freezing the main request. Recall is Human/user data, never instructions
   or grants. Interrupt, shutdown, and failed work must not attach to a later turn.
9. **Keep deduplication with retained context.** Compare candidate source hashes
   and full unquoted content against retained recall, readable dialogue, tool
   results, live memory, and projected instructions/manuals. Keep SHA-256 hashes
   of the full selected snippets in `RecallAttachment`, even when output is
   capped. Save/replay preserves them. A real context cut releases them; do not
   add an ever-growing per-session “seen” registry. Chapter 21’s handoff alone
   does not remove them while source awaits a committed replacement.
10. **Make policy durable and truthful.** Share Agent-home `settings.json` with
    Memory. Legacy memory-only files use recall defaults; malformed supplied
    settings refuse. Save before publishing, preserve the other saved policy, and
    acknowledge actual application in the GUI. A turn captures its policy at
    admission; changes govern future turns, not an already-running judge.

Defaults are recall and judging enabled, current model, 80 Unicode characters
minimum query length, 20 candidates, 3 recent messages, 3 selected snippets, and
6,144 bytes for the complete attachment. `MinScore` defaults to 3.0 for mechanical
selection; the judge path uses a 0.5 candidate floor so a semantic judge has a
wider choice. Numeric zero requests the documented default, not zero results.

Verify with real components and boundary fakes, then use an isolated actual host.
Test malformed bytes and recovery holds, separate data-role rendering, exact
human-input preservation, cancellation, replay/hash lifetime, durable restart,
and judge usage without changing main observations. A fake provider can establish
wire behavior, not whether a real model finds the right memory.

## §22.1 — The Retrieval Pipeline

The pipeline is intentionally small:

```text
new admitted Human turn
        ↓
record the exact Human input
        ↓
query long enough and recall enabled?
        ↓
fresh rooted scans → independent BM25 indexes
        ↓
merge ranked hits → retained-context duplicate filter
        ↓
optional stateless judge, or mechanical quotas
        ↓
actor joins, verifies turn, rechecks duplicates, caps output
        ↓
separate RecallAdded → main provider request
```

A recall can come from an old session, a retired compression input, an
architectural note, a learning, or a skill description. These are different
corpora because documentation can otherwise dominate the statistics. It is
usually longer and repeats words such as “agent,” “tool,” and “message” much more
often than a dated memory. Independent scoring does not make scores calibrated
probabilities; it prevents one large corpus from determining every other
corpus’s document frequency and average length.

The index is a derived value, not another copy of the filesystem’s authority.
`Recaller.Search` first scans eligible sources and hashes exact bytes and names.
It reuses an unchanged immutable index, builds replacements locally, and publishes
the complete refreshed collection only after all scans succeed. It does not
cache a successful answer across turns or assume a file’s size proves it unchanged.

Agent-home traversal goes through the Agent’s Sandbox. It omits discovered
inaccessible, non-file, hidden, and outward entries. Readable invalid UTF-8 and
oversized files refuse the refresh rather than silently producing a partial index.
`Memory.RecallFiles` holds Memory’s disk gate and refuses when a staged receipt or
committed replay needs recovery. Recaller cannot make a half-promoted transaction
searchable by pretending it is an ordinary directory scan.

The skill tree has a different owner. A host can deliberately select a Skills
source outside the workspace. Skills retains that provenance and lazily opens a
read-only root, using shared handle-relative traversal and bounded decoding. It
reads only visible `SKILL.md`, follows no links, and excludes disabled web-search
content. Each bootstrap manager owns an independent handle; destruction closes
it. There is no second Sandbox and no new permission for `read_file`. Fresh text
can help discover a skill but never rewrites activated templates or grants tools.

### One scheduler, two kinds of work

`SubmitAsync` and idle interactive input use the same recall admission path.
Busy hints retain their existing immediate-recording behavior; they do not start
a competing recall task. The query, policy, and Context are captured for the
admitted turn. The foreground worker searches and optionally awaits the judge,
then returns selected full snippets. Only the actor records `RecallAdded`.

This separation is why pause, settings, hints, interruption, and shutdown remain
responsive. Completion is not permission to publish onto whichever turn happens
to be active. The actor joins the task, checks its exact turn identity, and repeats
the duplicate check against current retained context. A hint arriving while the
judge waits may already contain the proposed memory. Interrupt and shutdown cancel
and join work; late results cannot leak into a later turn.

Judge failure can degrade to mechanical retrieval. That is not a promise that all
retrieval failures disappear. A corrupt readable corpus or unresolved recovery
hold refuses the affected turn. Calling that state “no matches” would conceal a
broken input boundary and encourage the model to answer without facts it expects.

## §22.2 BM25: The Old Formula That Still Works

BM25 is a well-established keyword-ranking formula. For this exercise's
small markdown corpus, it supplies the useful machinery without an embedding
service or another database.

The formula scores a document against a query by combining three ideas:

**Term frequency with saturation.** A word that appears ten times in a
document is more relevant than one that appears once, but not ten times
more relevant. The `k1` parameter (standard value: 1.2) controls how
quickly additional occurrences stop mattering.

**Inverse document frequency.** A word that appears in every document is
worthless as a discriminator. A word that appears in one document is
a strong signal.

Use the smoothed form, `ln(1 + (N - df + 0.5) / (df + 0.5))`, and not the
classic one clamped at zero. The difference looks cosmetic and is not. The
classic form goes negative once a term appears in more than half the chunks,
so a zero clamp discards those contributions. In a small archive, a useful
term can easily appear in more than half the chunks. If every query term does, the clamped total is zero and retrieval
returns nothing even though the files contain the requested words.

The failure is worth dwelling on because of how it presents. Nothing errors.
The index builds, the search runs, the scores are computed correctly
according to the formula, and the agent simply never remembers anything. It
reads as "recall is broken" when it is in fact "IDF is the wrong variant,"
and those two diagnoses send you to opposite ends of the codebase. The
smoothed form keeps these contributions positive. It fixes this zero-score
failure; a positive score can still fall below the configured floor.

A related trap waits in testing. IDF measures rarity *against the corpus*,
so a tiny fixture can put useful matches below a threshold tuned for a larger
archive. A representative background corpus changes what counts as rare.
The tests here use 25 background chunks; that is a fixture choice, not a
minimum corpus size. Check the actual query and scores before blaming the
retrieval path.

A positive score is also a weak ranking test. During the earlier implementation's
review, replacing BM25 with raw term frequency passed the initial tests.
Both algorithms found the planted words. The useful test puts plausible
answers in competition: a rare term against a common repetition, several
distinct query terms against one repeated term, and a focused note against
a long document with the same single match. Those comparisons exposed the
replacement while the restored BM25 implementation passed. The historical ranking credit
fell from 8/10 to 0 under that mutation. This is not a C# mutation run; C#
arithmetic and competing-rank tests provide their own evidence.

**Document length normalization.** A long document is expected to contain
more term matches than a short one. The `b` parameter (standard value:
0.75) penalizes long documents to compensate, using the average document
length as the baseline.

The implementation needs five things: a tokenizer, a stop-word list, an
index builder, a scorer, and a chunker that splits documents into pieces
small enough to be useful as recalled snippets.

Tokenize by splitting on non-alphanumeric characters, lowercasing, and
dropping tokens shorter than two characters. Filter stop words from
queries only. Keeping them in documents preserves accurate term frequency
and length statistics.

Chunk markdown on `## ` headers. Each chunk records its filename, header
text, content, pre-computed term frequencies, and token count. When a
chunk exceeds 4,096 UTF-8 bytes, split it further on `### ` sub-headers, then on
paragraph boundaries. Use breadcrumb headers (`"Parent Header > Sub Header"`)
so the reader knows where a sub-chunk came from.

The local ranking engine needs no external service or embedding database.
Its settings belong to the existing Agent policy, not a new search service. Build the index when the first query
needs it, after loaded history and memory recovery permit a corpus scan; rebuild
it when source bytes change. Equal-size edits count too; an unchanged
file length is not evidence of unchanged content.

Search both active and `.md.retired` memories, but keep hidden staging records
outside the corpus (`.memory-stage.json` in C#). A staged
candidate may not yet be accepted. If a committed replacement still needs disk
finalization, recovery establishes the committed bytes before indexing; a
conflicting file must not silently become recall input. Retired originals remain
searchable after recovery, subject to the same actual-context exclusion as other
memories. Constructing the recall service must not scan disk before its embedding
client has loaded history.

The positive logarithm above is deliberate. The prior Ensemble Chapter 22, 24,
and 26 implementations use it. CodeRhapsody’s comparison implementation instead
clamps `log((N - df + .5) / (df + .5))` at zero. Those are different score functions.
With Ensemble’s `log(1 + ratio)`, a matching term still has positive weight in a
one-document memory corpus. Keep that choice explicit when testing thresholds;
do not transplant a threshold from a different IDF convention and call the
scores equivalent.

## §22.3 Per-Source Quotas: Why Memory Gets Half the Slots

The first version of auto-recall ran a single merged BM25 search over all
sources. Design documents dominated the results. A 20KB design doc full of
technical terminology would outscore a 500-byte daily log that contained
the relevant decision in that corpus. Length normalization did not encode
the extra value of the daily log's decision and reason.

The fix is a source-selection policy. Give each source its own BM25 index
and search them independently, then allocate guaranteed slots:

1. Memory files get 50% of the slots (rounded up).
2. Remaining slots are split evenly among other sources.
3. Unused slots from sources with fewer results than their quota flow to
   sources with overflow.

Separate indexes matter on their own. IDF is computed per source, so a term
that is unremarkable across the memory archive can still be rare within the
documents, and a large source no longer distorts what counts as rare for
everyone else.

**Apply the quota to the final selection, not to the candidate pool.** This
distinction is the whole mechanism, and getting it backwards produces code
that looks right and does nothing. Candidates arrive already sorted by score
within each source, so capping a source's contribution to the candidate pool
may still leave the final score-ranked list entirely dominated by another
source. A candidate cap is not a reserved final slot. Test with two lower-
scoring memory hits and several higher-scoring document hits: a three-slot
mechanical result must keep the two memories.

**The quota must not overrule the judge.** It governs the mechanical paths,
where no judge ran or the judge failed and selection falls back to score
order. When a judge has actually assessed relevance, its verdict stands.
A successful `[]` means attach nothing; it is not a reason to fall back.
Discarding the opinion of the component that read the conversation, in
order to satisfy a ratio, would trade the good signal for the crude one.

This is a policy decision, not a BM25 improvement. The agent's personal
memories are the highest-value source. They contain decisions, context,
and lessons that no design doc captures. Guaranteeing them half the slots
ensures qualifying memories survive mechanical competition with larger
sources when enough memory hits exist. Unused slots still flow elsewhere.

Keep the full ranked pool for mechanical fallback. The candidate limit
restricts the judge prompt, not the memories available to fallback selection.

## §22.4 — The Judge: Filtering by Meaning, Not Keywords

BM25 finds shared words. A judge can decide whether the matching passage would
actually help this conversation. The judge is one stateless call through the
existing Engine, not a child Agent, second HTTP client, or remembered conversation.
By default it uses the current model on the same provider and funding route.
`RecallSettings.JudgeModel` can request another model on that provider without
mutating the running model or claiming entitlement to the override.

It receives:

1. Up to three recent ordinary Human/visible-Assistant entries as context,
   excluding recall, handoff, and tool entries.
2. The latest Human query, labeled separately. This alone formed the BM25 query.
3. Up to twenty numbered candidates, with source, filename, heading, and content.
4. Instructions to pick up to three useful candidates and return their indices.

Recent context and candidates are JSON-framed data, not accumulated History.
No skills, tools, Memory bands, or cache markers enter this request. Recursion is
impossible because the judge has no Agent admission path to trigger recall again.
This is stronger and simpler than a second Agent with a flag saying not to recall.

### The parser owns the format

Recall supplies both the prompt contract and `RecallText.JudgeChoices`. Engine
owns transport and native provider parsing, not a competing description of what
the judge should answer. If both layers specify formats, the model can obey one
and produce a reply the other rejects. Whoever interprets the bytes owns that
contract.

The C# parser extracts the first `[` through the last `]`, then parses a JSON
array. Fences and surrounding explanation therefore need not cause failure.
Integer values and integer strings—including surrounding whitespace—can select
zero-based indices. Duplicate, out-of-range, fractional, null, and nonnumeric
elements are ignored; they never silently become zero. A syntactically valid
`[]`, or an array with no usable indices, selects nothing. Invalid array syntax
is a judge failure, not a successful empty verdict.

There is a prior boundary: only a **completed, tool-free** provider response can
supply a verdict. `[0]` inside a truncated stream is not completed output. A tool
call is never dispatched. Frozen judge Configuration reaches every JSON/SSE
parser, including the override's provenance; ordinary Agent acceptance is unchanged.

### Timeout the operation, not merely the wait

The judge has a linked 15-second cancellation source reaching the real HTTP
operation. Returning early while a stalled stream keeps running would not meet
that contract. Its private timeout, transport failure, malformed output, or
incomplete response falls back to quota-selected BM25 over the full ranked pool.
Caller cancellation propagates instead: an interrupted user turn is not an excuse
to keep working mechanically and publish later.

A successful empty verdict stays empty. No quota pass runs afterward. The judge
has read the semantic context; treating its rejection as a failure would replace
that judgment with the noisier heuristic it was meant to improve.

The work still belongs on the bill. The existing Engine transport accounts actual
judge usage once in cumulative totals but leaves `Latest`, main-context pressure,
and CacheLens unchanged. Preserve actual billing model and tier; a missing provider
model identity stays unknown rather than being priced from the requested override.
Use the existing authentication and same-source retry path, not a parallel auth
system. Restoring a recorded attachment performs no judgment and adds no usage to
the new process meter.

## §22.5 — A Recall Entry Is Part of the Conversation

The result is not appended to the user’s text. C# records a separate
`RecallAdded` event whose single `RecallPart` holds an immutable
`RecallAttachment`: formatted text and the selected snippets’ full-content hashes.
Reduction makes a separate Human/data entry. It does not change phase, usage,
tool grants, or the original input’s bytes. A missing optional body is diagnosed
as a no-op; a supplied malformed body refuses before publication.

The formatter produces quoted, attributed data:

```text
[Auto-recalled memories]

From memory/2026-04-08-1.md.retired — Retry policy:
> Automatic retries were rejected because the provider may already have committed the remote transaction.
```

Provider rendering normalizes this special part through the shared readable-data
path. OpenAI, Anthropic, Gemini, and Responses receive it as user content, never
a privileged instruction. A document that says “ignore the user” is still a
document. Reading a skill manual this way does not load the skill.

The attachment remains in retained Context for short follow-ups, just like other
conversation data. There is no requirement to rerun a judge when the next message
is “Why did we reject that?” Recomputing an attachment during rendering would
ask today’s files and judge
to rewrite yesterday’s conversation. Record the selected bytes once; replay
never searches or judges again.

The separate part and event codec preserve its identity and hashes through snapshots and journal replay.

### Full-source identity, bounded presentation

The complete text—including title, attribution, newlines, and quote prefixes—is
capped at `MaxRecallBytes`, normally 6,144 UTF-8 bytes. The formatter stops at rune
boundaries, not arbitrary byte offsets. It records a hash only when some source
content actually fits. A giant paragraph therefore cannot blow up the request,
but its shortened presentation still identifies the original full snippet.

Before selection, `RecallText.Available` suppresses candidates already represented
by retained source hashes or full readable content. It checks ordinary dialogue,
nested tool results, memory bands, recall, and the actual projected instructions
and manuals. It does not inspect opaque replay as though it were visible text.
`Engine.ProjectInstructions` is shared with normal rendering so these paths do
not invent competing skill-retention rules.

The earlier implementation exposed this with an 11,247-byte source shortened
to 6,144 bytes: full-text matching missed the omitted suffix and the next turn
attached the capped text again. Its short-snippet control passed. That history
explains why C# tests both the cap and next-turn source-hash suppression.

The hash must live with the entry, not in a permanent “seen” table. Otherwise the
Agent could never recover a fact again after compaction removed it. V2 Chapter 23
released that identity when checkpoint removal discarded the recall entry. This
edition deliberately ties release to the committed replacement, not merely to a
handoff: Chapter 21 retains completed conversation source until a correlated
memory replacement commits, even though handoff removes old tool/private material.

The recall entry and its hashes survive that held interval because the model
still has the source. Clearing the hashes at handoff would let recall attach the
same material again before its replacement was safe. A committed cut that actually
removes the entry releases its deduplication identity. Replay has the same
behavior and does no disk scanning or new judging.

This also protects cache structure. Retrieval happens once at new-turn admission,
not during every tool-loop render. That is a stable-prefix design decision, not
proof that a live provider will credit a particular cache hit.

## §22.6 Tuning: What the Numbers Mean

These defaults emerged from months of daily use. The tuning explains their
purpose; it does not make the numbers right for every corpus.

**MinScore went from 0.5 to 3.0.** At 0.5, every conversation triggered
recall, and most results were noise: single keyword matches on common
technical terms. The higher floor reduced that noise in the original corpus.
The jump is large because BM25 scores are not
normalized; the right threshold depends on corpus size and vocabulary.
Note: this threshold applies only to plain recall (no judge). When the
judge is enabled, BM25 uses a low threshold (0.5) to cast a wide net,
and the judge handles false-positive filtering.

**MinQueryLength went from 15 to 50 to 80.** A user who types "fix it,"
"run tests," or "looks good" often supplies little retrieval signal. Raising the
threshold progressively eliminated recall on messages that could never
produce useful results. Eighty characters is a cheap heuristic, not a test of
meaning: a short
project identifier can carry more information than a long generic instruction.

**Candidates went from 10 to 80 (code default) to 20 (production).**
That was the CodeRhapsody tuning history. This exercise defaults
to 20; it does not inherit the old code's 80-candidate setting. More
candidates enlarge the prompt. Whether they improve selection enough to
justify that cost depends on the corpus and judge; count the actual candidate
bytes rather than assuming 20 snippets means a fixed number of tokens.

**Judge timeout is 15 seconds, not 2.** The original design assumed a
local model (Ollama) with sub-second latency. Cloud calls took 1–3 seconds in that workflow. The generous timeout exists
because a late judge response is better than no judge response. The
fallback to unfiltered BM25 results is noisier.

## §22.7 What BM25 Cannot Do

BM25 matches words. Three limitations follow:

**Vocabulary mismatch.** If the user says "authentication" and the memory
says "login flow," BM25 will not connect them. Vector search (embedding-based
retrieval) can bridge that gap, but it requires an embedding model, an index to maintain,
and a service to call. For a corpus of a few hundred files, the marginal
improvement does not justify the infrastructure. If the corpus grows by
orders of magnitude, revisit this.

**No semantic deduplication.** If a memory and a design doc contain the same
information in different words, both may be recalled, wasting snippet slots.
Exact content and complete-source identities prevent literal repeats already
in context. They do not establish that two paraphrases say the same thing,
and they do not exclude files merely because they are recent.

**No relevance feedback.** The system does not learn which recalled memories
the agent actually used. A snippet that the agent ignores on ten consecutive
turns can return after committed cuts remove its retained identity if it scores well on BM25. Tracking usage patterns
and adjusting retrieval is a natural next step that this implementation does
not take.

These are limitations, not a roadmap. BM25 with a judge worked well enough for
daily use in that system. The system is simple, fast, and debuggable. A
more sophisticated retrieval system would be harder to debug when it
recalled the wrong thing, and debugging recall errors is already hard.

## §22.8 Skill Discovery: The Unplanned Win

Visible, exact `SKILL.md` files in the explicitly selected skill source are
indexed alongside memories and documents, subject to launch policy.
This was added because skill docs are markdown files in a known directory,
and the existing source scan could include them.

In daily use, the result was unexpectedly powerful. When a user asks about something a
skill handles ("can you check my email?" or "what's on my calendar?"),
the skill's documentation surfaces through auto-recall. The agent can see the
snippet, recognize the available skill, and load it. No explicit skill
search, no user instruction to load a specific skill. The agent discovers
capabilities through the same mechanism it uses to recall past work.

This turns auto-recall into a discovery system as well as a memory system.
The agent's capabilities are not fixed at startup; they expand when the
conversation moves into a domain where a skill exists. The user does not
need to know which skills are installed. The skills surface when they are
relevant.

Recall does not activate a skill or grant its tools. It supplies documentation
that can lead to an ordinary `load_skill` call. The existing Skills manager still controls
which skills are reachable and which tools may execute. A paragraph retrieved
from disk cannot grant itself a capability.

### C# policy and ownership map

`RecallSettings` is an immutable value. `Agent.SetRecallSettings` is an in-memory
next-turn update; `SaveRecallSettingsAsync` asks the actor for a durable update.
Do not confuse them. Running recall keeps its captured policy.

The version-1 Agent-home `settings.json` now contains both `memory` and `recall`.
A legacy memory-only document remains readable with default recall settings.
Supplied unknown, duplicate, missing, or malformed fields refuse; only a genuinely
missing file gets defaults. The bounded reader and writer agree on their limit.
Saving one section preserves the other section’s **saved** policy, not an unrelated
transient in-memory edit. Startup restores both before the first provider request,
including when a policy command is the first actor operation.

The GUI exposes recall enablement, judge enablement/model, score and query limits,
candidate/context/snippet counts, and the byte cap. Its selected-Agent update path
awaits the actor and acknowledges actual durable application. It does not cancel
the current judge, and it does not optimistically claim a saved checkbox. Appearance
settings remain a different store; mixed patches do not promise a cross-file
transaction. A bad sibling settings document refuses a live save without stopping
an otherwise healthy actor; malformed startup policy still fails startup.

The source map is deliberately ordinary:

| Responsibility | C# files under the chapter endpoint |
|---|---|
| Values, policy, owned Recaller, actor/judge declarations | `Common/Recall.cs`, `Common/RecallCorpus.cs`, `Common/RecallJudge.cs`, `Common/RecallRuntime.cs` |
| Tokens, BM25, chunking, quotas, verdicts, formatting, deduplication | `Recall/Index.cs`, `Recall/Selection.cs`, `Recall/Context.cs` |
| Lazy indexes and fresh home-corpus reads | `Recall/Corpus.cs`, `Recall/Files.cs`, `Memory/Recall.cs` |
| Explicit read-only skill source | `Skills/Recall.cs`, `Sandbox/ReadRoot.cs` |
| Stateless judgment and actor publication | `Recall/Judge.cs`, `Engine/RecallJudge.cs`, `Agent/Recall.cs` |
| Durable policy | `Serialization/RecallSettings.cs`, `Memory/Preferences.cs` |

All persistent state and callable signatures remain in Common. Functional modules
supply bodies in the same compiled assembly. Recaller reaches Engine, Memory,
Skills, History, and Sandbox through its actual Agent parent; it does not receive
a bag of copied services. Destroying it detaches both ends and clears derived
indexes without deleting corpus files. Immutable snippets/indexes/attachments are
value-like and need no artificial ownership back-pointers.

## §22.9 — Verify the Implementation

Use real algorithms and components with a small working fake at the provider
boundary. The focused C# checks cover arithmetic, chunking and byte limits;
rooted fresh scans and skill isolation; replay and committed-cut hash lifetime;
complete tool-free judge replies, fallback, and real stalled-HTTP cancellation;
ordered actor admission and late-hint deduplication; and durable policy/restart.

From this chapter’s source endpoint, with .NET 10 or later installed:

```sh
dotnet build Tests/Tests.csproj
for group in recall recall-corpus recall-replay recall-judge recall-runtime recall-settings; do
  dotnet run --project Tests/Tests.csproj --no-build -- "--$group-tests"
done
```

At this chapter boundary the shared suite passed **297 groups**, and both actual
CLI and GUI hosts built cleanly. The isolated actual-browser/CLI exercise made
four local-provider requests: it retrieved a retired fact, retained it for a short
follow-up, saved and reloaded disabled recall through visible controls, and honored
that setting after CLI restart. Both hosts joined cleanly.

The opt-in executable grader runs the real CLI and reuses seven complete rows
from the original rubric. Its 70/70 is a pass for those rows, not whole-feature
acceptance; library checks do not become executable credit:

| Executable row | Points |
|---|---:|
| BM25 ranking | 10 |
| Query stop words | 5 |
| Markdown chunking | 10 |
| Mechanical source quotas | 10 |
| Judge filtering, including successful empty choices | 15 |
| Error/timeout fallback | 10 |
| Capped attachment and next-turn deduplication | 10 |
| **Verified executable scope** | **70/70** |

The remaining **30 points are ungraded**: corpus graduation/restart (15), the
original recall-kind/lifetime row including its historical handoff assumption
(10), and Go topology (5). The C# committed-cut contract is tested separately,
not silently translated into the older row. The original Go route remains intact.
The C# route is `grade-edition2 -ch 23 -executable-only -json EXECUTABLE`; supply an
executable wrapper for the built CLI, not a DLL treated as a native executable.

The successful run joined nineteen CLI sessions and retained thirty-seven
recorded local peer requests, plus a held timeout exchange outside that recorder.

## §22.10 — Take It for a Spin

Now test meaning as well as plumbing. The exercise needs a fact absent from the
current conversation, a long query that can trigger recall, and a short follow-up
that can succeed only if the attachment remains retained.

Use a fresh temporary data root and an explicitly chosen provider configuration.
Do not borrow another running Agent’s home, browser, or credentials. A real-provider
run can spend money; use your own approved configuration. A local fake can verify
wire structure but cannot establish relevance or answer quality. Run this from the
chapter endpoint with .NET 10 and Python 3 available:

```sh
set -eu
: "${LLM_VENDOR:?Set openai, anthropic, or gemini for your approved provider}"
: "${LLM_MODEL:?Set an exact model ID available to your account}"
main=$(pwd)
dotnet build Cli/Ch04.csproj
work=$(mktemp -d "${TMPDIR:-/tmp}/ensemble-ch23.XXXXXX")
mkdir -p "$work/skills/base" "$work/data/agents/root/memory"
cat > "$work/skills/base/SKILL.md" <<'SKILL'
---
name: base
description: Isolated recall exercise
type: primary
tools: read_file
---
Answer in chat using supplied evidence. Keep uncertainty explicit.
SKILL
cat > "$work/data/agents/root/memory/2026-10-23-1.md.retired" <<'MEMORY'
## QuartzZephyr migration decision
Project QuartzZephyr, release QZ731: Morgan owns the migration rehearsal on October 23. We rejected automatic retry after timeout because the remote transaction may already have committed. Read the migration ledger before deciding whether retry is safe. Ledger placement remains unresolved: application database versus separate control database. Retain rollback artifacts for nine days.
MEMORY
python3 - "$work/data/agents/root" <<'PYTHON'
import json, pathlib, sys
home = pathlib.Path(sys.argv[1])
for name in ('memory', 'docs', 'handoffs', 'learnings'):
    directory = home / name
    directory.mkdir(exist_ok=True)
    for i in range(25):
        (directory / f'padding-{i:02d}.md').write_text(
            f'## Padding {i}\nOrdinary unrelated archive material.\n')
(home / 'docs' / 'recipes.md').write_text(
    '## QuartzZephyr migration name\n'
    'QuartzZephyr migration is also the name of our cafeteria recipe contest. '
    'QuartzZephyr migration lunch recipes use apples and cinnamon, '
    'not database transactions.\n')
# Disable live memory injection so this test must retrieve the retired source.
band = dict(disabled=True, budget=5000, high_watermark=10000, low_watermark=5000)
settings = dict(version=1,
    memory={name: band for name in
            ('conversation', 'soul', 'memory', 'archive64', 'archive8', 'session')},
    recall=dict(disabled=False, disable_judge=False, judge_model='', min_score=3,
                max_snippets=3, min_query_length=80, candidates=20,
                context_messages=3, max_recall_bytes=6144))
(home / 'settings.json').write_text(json.dumps(settings))
PYTHON
cd "$work"
dotnet "$main/Cli/bin/Debug/net10.0/ch04.dll" \
  --data-dir "$work/data" --agent-name root \
  --skills-dir "$work/skills" --primary-skill base --billing api chat
```

This example explicitly selects API funding. Before running it, export the chosen
`LLM_VENDOR` and exact `LLM_MODEL`, and make your own approved credential available
through `LLM_API_KEY` or the corresponding `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`,
or `GEMINI_API_KEY`. Do not paste a key into the manuscript or transcript.
This sets up files and launches the CLI; it does not create credentials or log
into an account. For a plan-funded exercise, use Chapter 16’s explicit account
and credential-directory launch setup instead of this API billing flag.
Do not add `--surface anthropic`: Anthropic is selected by the vendor, while that
flag selects an OpenAI surface.

Ask:

```text
From our archived QuartzZephyr migration discussion, tell me the release code, owner and rehearsal date, rejected approach and why, the unresolved decision, and rollback retention. Use only the context already supplied; do not call tools. If a fact is absent, say so.
```

The answer should recover both the duplicate-side-effect reason and the unresolved
ledger location. A recipe-only answer is a selection failure even if the request
contains all the right keywords. A generic answer that omits the ledger question
is not enough evidence that the source was used.

Then ask:

```text
Who owns that rehearsal, and why was automatic retry rejected?
```

The default length threshold should skip new retrieval. The answer should still
name Morgan and explain the possibly committed remote transaction because the
first attachment remains in retained Context. Inspect
the actual request as well as the answer: a model can guess, or can ignore facts
that are present.

After a clean exit, inspect `events.jsonl` and `api.log` under
`$work/data/agents/root/`. Check for the exact original Human input, a separate
`recall_added` event containing a `recall` part, full-source hashes, and user-role
projection. For the successful long-query route expect a tool-free judge exchange
followed by the main exchange; for the short follow-up expect a main exchange with
the retained attachment. Logs are evidence and may contain private prompts—keep
this exercise isolated rather than publishing them casually.

Repeat with judging disabled to compare mechanical quotas. Add a long paragraph
to test the complete byte cap. Change or remove a source between eligible queries
to test fresh reads. In the GUI, save disabled recall, reload the controls, and
restart with the same home; a new planted fact must not appear merely because an
old checkbox still looks checked. Re-enable and test again. A settings acknowledgment
is not a claim that a currently running turn changed its captured policy.

Finally exercise Chapter 21’s real committed conversation cut. A handoff alone
must not erase retained hashes while replacement is pending. After a committed
cut removes the recall entry, ask an eligible query again and confirm that the
archived fact can be recalled anew. Do not substitute quiet time for observing
the correlated commit and the resulting Context.

### Historical field report, not C# provider acceptance

The following account is retained from the earlier chapter implementation. Its
model names and results describe that run, not current recommendations, entitlements,
or a fresh live test of this C# endpoint:

The retained three-provider CLI runs used these two prompts and the same
planted fact. Haiku 5.5, GPT-4o mini and Gemini 3.8 Flash each answered both
turns correctly without tools. Each made one tool-free judge request followed
by two main requests. The judge saw the decision and cafeteria distractor;
only the decision reached the main requests, once in each. OpenAI's judge
wrapped `[0]` in a JSON fence, exercising the tolerant parser in actual use.
Anthropic also noted that the archive supplied no year.

Those are bounded outcomes from the earlier implementation, before the
capped-source and path corrections. The historical Haiku 5.5 result is not a recommendation to use it now; this
C# exercise leaves provider and model selection explicit.
This isolated corpus demonstrates retrieval, retention and correct answers,
not general recall quality across a private archive.

The original grader returned 0/100 before it could test that behavior: its
primary skill omitted `type`, and its structural check looked for the core
recall package under the separate GUI module. Passing a launch check and
remembering a decision are different accomplishments. That earlier revised grader
used a typed primary and the Agent's corpus paths, with separate fixtures
from this CLI demonstration. Those earlier live answers establish the narrower and
more interesting fact that the archived reason actually came back.

← [Chapter 21: Memory](chapter-21.md) · [Chapter 23: While the Cache Is Broken →](chapter-23.md)
