# Chapter 23: Auto-Recall

## §23.0

Chapter 22 built a memory system. The agent can save, compress, and
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

## TL;DR

Build passive memory retrieval. Every user message triggers a BM25 keyword
search over the agent's memory files. A small language model filters the
results for relevance. Surviving snippets are attached to the conversation
as their own kind of entry, where they stay until a checkpoint removes them.

**Data structures.** This sketch shows the index vocabulary; private field
names and storage choices are yours.

```go
// BM25 engine — standard information retrieval, nothing exotic.
type Index struct {
    Chunks    []Chunk
    AvgLength float64           // average chunk length in tokens
    DocFreq   map[string]int    // how many chunks contain each term
    Total     int               // total number of chunks
}

type Chunk struct {
    Filename string             // source file (e.g., "2026-09-27-1.md")
    Header   string             // markdown section header
    Content  string             // the text
    Tokens   []string           // pre-computed tokenization
    Length   int                // token count
    TF       map[string]int     // term frequencies
}

// BM25 tuning — standard values, no reason to change them.
const (
    k1 = 1.2   // term frequency saturation
    b  = 0.75  // document length normalization
)
```

**The recall pipeline** has three stages:

1. **BM25 search.** Tokenize the user's message, remove stop words, score
   every chunk in each source index. With a judge, send up to `Candidates`
   (default 20) highest-scoring chunks above a low threshold (0.5). Keep
   the full hit list for fallback. Without a judge, use `MinScore` (3.0).
   On the final mechanical selection only, reserve half the slots, rounded
   up, for memory; divide the rest among other sources with hits and
   redistribute unused slots.

2. **SLM judge.** Send the candidates to a cheap language model
   on the Agent's provider along with the last few
   conversation messages for context. The judge returns a JSON array of
   indices, the candidates it considers genuinely relevant, not just
   keyword matches. If the judge fails, times out (15 seconds), or is
   disabled, fall back to score order with the final source quotas. A
   successful verdict, including `[]`, stands; quotas never replace it.

3. **Attachment.** Format the surviving snippets (up to `MaxSnippets`,
   default 3) as blockquoted text with source attribution. Record them as
   their own entry kind, never merged into the user's message. Cap total
   recalled text at 6KB (6,144 bytes), including heading, attribution and quote
   prefixes. Preserve valid UTF-8 at the cut. Skip any snippet already present
   in actual context, including enabled memory bands. Retain the complete
   source identity for capped snippets so truncation cannot defeat that
   check. The entry persists until a checkpoint removes it.

**Sources searched.** All paths below belong to the Agent's data directory;
the supplied CLI and GUI choose `cr/` beneath their launch directory. Skills
use the Agent's separately configured skills directory.

| Source | What | Final mechanical share |
|--------|------|-------|
| Memory files | `memory/`, `bucket-0/`, `bucket-1/`, `MEMORY.md`, `SOUL.md`; include archived `.md.retired` files | Half the slots, rounded up |
| Docs | Markdown under `docs/` | Split the remainder |
| Handoffs | Markdown under `handoffs/` | Split the remainder |
| Learnings | Markdown under `learnings/` | Split the remainder |
| Skill docs | Available `SKILL.md` files | Split the remainder |

Search all memory files, including the newest. Omit content already in actual
context instead of guessing from a filename which memories the model has seen.
Resolve loaded history and memory's staged recovery before building or refreshing
an index. Hidden candidates are outside every recall source. A committed replacement
must finish disk finalization before search; a conflict returns its recovery error
without searching a mixed corpus. After successful graduation, retired originals
remain searchable while surviving active memory is excluded by actual context.
A failed attempt keeps its source live and subject to that same exclusion.
The historical CodeRhapsody `learnings.json` format is not an input to this
exercise; long-form learning notes here are ordinary markdown.

**Chunking.** Split markdown on `## ` headers. Large chunks (over 4KB) split
further on `### ` sub-headers, then on paragraph boundaries. Sub-chunk headers
use breadcrumb format: `"Parent Header > Sub Header"`.

**Tokenization.** Split on non-alphanumeric characters, lowercase, minimum
2 characters. Filter stop words from queries only. Documents keep all terms
for length normalization.

**The recall judge is one stateless model call**, not an agent. It takes a
prompt and returns text. Because it holds no history and owns no context,
it cannot recall, which makes the recursion problem structurally impossible
rather than something to guard against. Add its measured usage once to Engine
session totals with the actual provider, model and price tier; keep latest-main
counts and main-request cache-lens comparisons unchanged. Restoring an attachment
neither invokes a judge nor imports historical spend.

**Recall is a new spoke.** It imports `internal/common` and the standard
library. Ensemble constructs it as an Agent child with a required parent
interface. It reaches the Agent's configuration, skills and Engine through
that owner; common declares `Recaller` and `SnippetJudge`. The actor gives a
worker a conversation snapshot and attaches its returned bytes through the
existing ordered input path. A slow judge must leave hints and interruption
usable, and an interrupted worker must not attach to a later turn.

**Parse the judge's response defensively.** Small models are unreliable
with output format. Try `[]int` first, fall back to `[]string` (models
sometimes return `["0", "3"]`), fall back to `[]json.RawMessage` for
mixed types. Extract the first `[` to last `]` from the response,
models often wrap the JSON in explanation text. Ignore invalid, repeated
and out-of-range indices; do not interpret `null` as index zero.

**Configuration.** These are the exercise defaults. The public settings patch
uses a `recall` object with lower-snake-case keys, such as `max_snippets`.
Zero numeric fields select defaults; `disabled` bypasses recall and
`disable_judge` selects mechanical retrieval.

| Setting | Default | Purpose |
|---------|---------|---------|
| `MinScore` | 3.0 | BM25 score floor for plain recall (no judge) |
| `MaxSnippets` | 3 | Maximum snippets injected per turn |
| `MinQueryLength` | 80 | Skip recall for short messages |
| `JudgeModel` | Agent's current model | Optional same-provider override |
| `Candidates` | 20 | BM25 results sent to the judge |
| `ContextMsgs` | 3 | Recent messages for judge context |
| `MaxRecallBytes` | 6144 | Hard cap on total recalled text |

**Graded behaviors (10 checks, 100 points):**

| Check | Points | What it tests |
|-------|--------|---------------|
| `bm25-indexes` | 10 | Memory files are searchable after finalized and failed-compaction restarts; active content is not duplicated |
| `bm25-scores` | 10 | Competing hits distinguish rarity, saturation and length normalization |
| `bm25-stop-words` | 5 | Stop words are filtered from queries, not documents |
| `chunking-splits` | 10 | Large documents are split on markdown headers |
| `per-source-quota` | 10 | Final mechanical quotas and spare slots; successful judge choices stand |
| `judge-filters` | 15 | The SLM judge removes irrelevant BM25 hits |
| `judge-fallback` | 10 | Judge failure falls back to quota-selected BM25 results |
| `recall-is-a-spoke` | 10 | The actual core recall spoke imports only common and stdlib |
| `recall-is-own-kind` | 10 | Recall remains separate and retained on the next turn and after source-free restart |
| `injection-capped` | 10 | Formatted text fits MaxRecallBytes and capped recall is not repeated |

Run the separately versioned grader from the repository root:

```sh
go run ./cmd/grade-edition2 -ch 23 -json
```

It defaults to `solutions/ch23` and its separate GUI. Public settings
and corpus fixtures exercise the actual Agent. Source review still establishes
parent ownership, cancellation and the fuller replay/provider obligations; an
import check alone cannot prove those relationships. Exact hidden-stage and
committed-recovery holds, judge model/tier accounting and main cache-lens exclusion
retain separate real-runtime source regressions, without extra numeric credit.

## §23.1 The Idea in Plain Words

A filing cabinet is not memory. Memory is what comes to mind without being
asked.

Chapter 22 built the filing cabinet: bands that compress, a cascade that
graduates, a watermark that forces the agent to write things down before
they disappear. That gives recent work a place in context and older work a place
on disk. It does not answer the question "which old memory matters to this
new message?"

The answer is the oldest trick in information retrieval. When the user
types a message, pull out the keywords, score every memory chunk against
them, and surface the best matches. BM25 scores a small local corpus
without another service. No vector
database, no embedding model, no infrastructure. Keywords and arithmetic.

But BM25 is a blunt instrument. It matches words. A query
about "authentication" will not find a memory about "login flow." And it
can rank a keyword-rich design document above the short daily log
containing the decision. Length normalization helps, but a matching term
is still not evidence that the document answers the question. Two mechanisms
compensate:

**Per-source quotas.** When selection is mechanical, qualifying memory
hits get half the final slots, rounded up. Docs, handoffs, learnings, and
skill docs share the rest. This protects personal memories from a larger
reference corpus without forcing a successful judge to pick an irrelevant
memory merely to satisfy a ratio.

**The recall judge.** A cheap language model, the smallest that can follow
instructions, reads the BM25 candidates alongside the recent conversation
and picks the ones that are genuinely relevant, not merely keyword-adjacent.
The judge sees context that BM25 cannot: what the conversation is about,
what the user is trying to do, whether a keyword match is superficial or
load-bearing.

The judge is the difference between "here are some memories that mention
the same words" and "here is something you need to know right now." Without
it, a developer asks about authentication and gets every memory that
mentions a password. A good judgment selects the one where the team
decided to use OAuth. It can still make a bad choice, and it cannot select
a memory that keyword search never offered.

The obvious lifetime for a recalled snippet is ephemeral. Relevance is local:
a memory about database schema is useful when the user mentions the database
and noise when the conversation moves to the GUI. Show it, drop it.

That intuition can be expensive because a prompt cache is a prefix property.
Dropping a block changes the prefix from the deletion point onward. Moving it
to a volatile tail avoids that old position but changes the layout around each
retrieval. Attach it once at a stable position and subsequent requests can
extend the same prefix. The bytes are still sent, and cache reads can still
cost money; a stable layout creates an opportunity for reuse, not a promise
that the provider charges for those bytes only once.

So recalled snippets stay until a checkpoint. They are not merged into the user's
message, because the user's words must stay distinguishable from
machine-retrieved text, and because anything fused into a user turn cannot
later be removed without damaging it. They are their own kind of entry,
which makes them addressable: a checkpoint can delete them by kind.

That checkpoint is where removal belongs. A checkpoint already rewrites the
prefix, so it already changes the cacheable suffix, and deleting recall there
avoids a separate invalidation at another point in the turn. Removing it
on its own schedule would make that additional change without recovering
more useful context. The rule generalizes past this chapter:
**group removals at the checkpoint that already changes the prefix.**

This is the second appearance of a shape worth naming. Chapter 22 found that
switching a memory band off is more expensive than leaving it on, because
disabling strips bytes from the front. Here too, the transient-looking thing can be
cheaper to keep than to remove. In a system built on cached prefixes,
retention can be cheaper than repeated removal. The intuitive lifetime
deserves a look at the actual request layout.

The system degrades gracefully at every level. If the query is too short
(under 80 characters), recall does not fire. Short messages like "fix it"
contain no signal worth matching. If BM25 finds nothing above the score
floor, nothing is injected. If the judge fails or times out, the system
falls back to quota-selected BM25 results. If the entire recall
system is disabled, the agent still works. It does not remember
passively. Retrieval failure should leave a usable agent with less context;
interruption and shutdown still end the current work normally.

One unexpected benefit: skill discovery. All available SKILL.md files are
indexed alongside memories. When a user asks about something a skill
handles, the skill documentation surfaces through recall, and the agent
loads it without being told. This was not planned. It was added because
skill docs were already markdown files in a known directory, and the source
scan could include them. It turned auto-recall into a discovery mechanism as well
as a memory mechanism.

## §23.2 BM25: The Old Formula That Still Works

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

A positive score is also a weak ranking test. During this implementation's
review, replacing BM25 with raw term frequency passed the initial tests.
Both algorithms found the planted words. The useful test puts plausible
answers in competition: a rare term against a common repetition, several
distinct query terms against one repeated term, and a focused note against
a long document with the same single match. Those comparisons exposed the
replacement while the restored BM25 implementation passed.

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
chunk exceeds 4KB, split it further on `### ` sub-headers, then on
paragraph boundaries. Use breadcrumb headers (`"Parent Header > Sub Header"`)
so the reader knows where a sub-chunk came from.

The whole engine is a few hundred lines of code. No dependencies, no
configuration files, no external services. Build the index when the first query
needs it, after loaded history and memory recovery permit a corpus scan; rebuild
it when source bytes change. Equal-size edits count too; an unchanged
file length is not evidence of unchanged content.

Search both active and `.md.retired` memories, but keep hidden staging records
outside the corpus (`.compaction/pending.json` in this implementation). A staged
candidate may not yet be accepted. If a committed replacement still needs disk
finalization, recovery establishes the committed bytes before indexing; a
conflicting file must not silently become recall input. Retired originals remain
searchable after recovery, subject to the same actual-context exclusion as other
memories. Constructing the recall service must not scan disk before its embedding
client has loaded history.

## §23.3 Per-Source Quotas: Why Memory Gets Half the Slots

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

## §23.4 The Judge: Filtering by Meaning, Not Keywords

BM25 finds documents that share words with the query. The recall judge
finds documents that share *relevance* with the conversation.

The judge is a single stateless call to a model that can follow instructions.
This implementation defaults to the Agent's current provider and model;
`recall.judge_model` can select another supported model on that same provider.
The override changes this request, not the main conversation's model. A smaller
model can reduce cost, but its selection quality needs checking. It receives:

1. The recent conversation (last 3 user and assistant messages).
2. The latest user message, labeled separately.
3. Up to 20 numbered candidate snippets from BM25.
4. An instruction: pick up to 3 that the agent would genuinely benefit
   from seeing, return a JSON array of indices. "Keyword mention alone
   is not relevance."

The judge returns `[0, 7, 12]` or `[3]` or `[]`. Several details make
this work reliably:

**It starts clean every time.** The judge holds no history. Note what that
means about item 1 above: the recent conversation reaches the judge as *data
inside the prompt*, not as history the judge carries between calls. Every
judgment starts with just the supplied data and instructions.

**Recursion is impossible rather than prevented.** An agent that could recall
would trigger its own recall when judging, which would invoke another judge.
A stateless call has no context of its own to fill, so the failure cannot be
expressed. This is worth noticing as a design move: the strongest way to
handle a failure mode is to build a thing that cannot exhibit it. An
implementation that made the judge a full agent would need a nil provider and
a settings flag to hold the same line, and would still be one refactor away
from losing it.

**Defensive JSON parsing.** Small models are unreliable with output
format. The parser tries `[]int`, falls back to `[]string` (models
sometimes quote the numbers), falls back to `[]json.RawMessage` for
mixed types. It extracts the first `[` to last `]` from the response
to handle models that wrap JSON in explanation text. Out-of-range
indices are silently filtered. Clear any partially decoded integer slice
before trying mixed elements: a failed `json.Unmarshal` can already have
written values into it. Otherwise the fallback can accidentally retain
zeros that select candidate zero.

**One component owns the reply format, and it is the one that parses.**
Nothing else may restate it. This sounds like housekeeping and is the most
dangerous rule in the chapter, because the model layer holds the client and
can feel like the natural place to describe what a good answer looks like,
while the recall layer holds the parser. If both describe it, their
instructions can drift. The model then receives two specifications in one
request, obeys one, and the parser can reject a reply that was well formed
by the other. Recall falls back to scores while the judge appears to be
working. Keep the format with the parser, and log fallback so that a failed
judgment cannot pass unnoticed as a useful one. Whoever reads the bytes
decides what the bytes look like.

When the judge fails or times out (15 seconds), the system falls back
to the final quota-selected BM25 results. Cancellation reaches the actual
HTTP request; returning from a wait while leaving it running is not a
timeout. The agent can continue with noisier recall.

Its usage still belongs on the bill. Add measured judge input, cache and output
counts to Engine's session totals with the judge's actual provider, model and price
tier, once. A cheaper judge must not be priced as the main model. Keep the latest
main-response counts unchanged, and exclude this separate prompt from main-request
cache comparisons. Restoring an old recall attachment makes no new judge call
and adds no historical usage to the fresh session meter.

## §23.5 Attachment: Permanent Until a Checkpoint

Recalled snippets are formatted as blockquoted text with source attribution:

```
[Auto-recalled memories]

From 2026-09-22-1.md — Key decisions:
> save_memory removed as agent tool; micro_handoff becomes SOLE trigger.
> A finite compressor folds completed work into memory.

From docs/auto-recall-design.md — The Pipeline:
> BM25 keyword search over memory files, filtered through an SLM judge.
```

That text becomes its own kind of entry in the conversation. Three properties
follow, and each one is load-bearing.

**It is never merged into the user's message.** The user's actual words must
stay distinguishable from machine-retrieved text, both for the model reading
them and for anyone auditing the log afterward. Fusing the two makes removal
depend on editing the human message. A checkpoint can instead remove the recall kind while preserving
the human's words according to the existing conversation policy.

**Where it lands on the wire is the renderer's decision, not the
conversation's.** The conversation records that recall happened and what
bytes it produced. Turning that into a request is a vendor question. These
are retrieved data, not new instructions: use the renderer's existing data
placement. Anthropic can put a recalled block and human text in separate parts
of one user message without merging their stored entries. A check that demands
a separate system message would confuse storage identity with wire layout.
Keep the recorded position stable on later requests.

**It carries its bytes.** Recall is nondeterministic, because a language
model chose which snippets survived. If the text were recomputed while
rendering, replaying a log would call a model, and two replays of the same
log could disagree. So the snippets are recorded as text at the moment they
are chosen. Replay then reproduces the conversation exactly, with no search
and no model call.

That last property is the same boundary Chapter 22 drew. A transform that is
a pure function of the log can run at render time. Anything nondeterministic
has to be recorded when it happens. The checkpoint and memory chapters
already use that boundary.
Recall follows it: log the selected bytes, then render them. Do not rebuild
an old attachment from today's files or ask today's judge to choose it again.

Because the entries persist, the same memory could be recalled on three
separate turns and appear three times. Skip content already in the actual
context, including memory bands, tool results and earlier recall. Searching
all files does not mean attaching them all again.

Truncation makes that rule less obvious than it sounds. The initial
implementation compared the complete source snippet with the conversation.
A source of 11,247 bytes became a 6,144-byte attachment. Its missing suffix
could never match, so the next turn attached the same capped text again.
A short-snippet control passed; the large one produced two identical entries.

Keep the complete source's identity with the attachment that represents it.
The accepted implementation records a hash of that content alongside the
chosen bytes. It hashes the source, not the shortened display, so a human edit
creates a new identity. Replay retains it; checkpoint removal discards it with
the recall entry, allowing a later turn to retrieve the source again. This is
local state belonging to the attachment, not a permanent “already read” list.

One wiring hazard deserves naming, because it costs hours and announces
nothing. The first edition had a synchronous turn path and an actor-mailbox
path. Hook recall into only one and the hand-driven case can work while the
other client never retrieves anything. A function that was never called
cannot complain.

The present Agent has one actor begin path for both `Ask` and `Post`. Keep
recall there, after memory recovery and before the first main request, and verify
both actual clients. An unresolved committed replacement must report its recovery
error before search or attachment can consume a mixed corpus.
The worker reads a snapshot while the actor can still accept hints or an
interrupt. Its result carries the turn identity; a late judgment from an
interrupted turn must not become context for the next user's message.


## §23.6 Tuning: What the Numbers Mean

The first-edition account traces these defaults through months of daily
use. The stories explain their purpose; they do not make the numbers right
for every corpus.

**MinScore went from 0.5 to 3.0.** At 0.5, every conversation triggered
recall, and most results were noise: single keyword matches on common
technical terms. The higher floor reduced that noise in the reported corpus.
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
That was the earlier CodeRhapsody tuning history. This exercise defaults
to 20; it does not inherit the old code's 80-candidate setting. More
candidates enlarge the prompt. Whether they improve selection enough to
justify that cost depends on the corpus and judge; count the actual candidate
bytes rather than assuming 20 snippets means a fixed number of tokens.

**Judge timeout is 15 seconds, not 2.** The original design assumed a
local model (Ollama) with sub-second latency. The historical account reports
cloud calls taking 1–3 seconds. The generous timeout exists
because a late judge response is better than no judge response. The
fallback to unfiltered BM25 results is noisier.

## §23.7 What BM25 Cannot Do

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
turns can return after checkpoints if it scores well on BM25. Tracking usage patterns
and adjusting retrieval is a natural next step that this implementation does
not take.

These are honest limitations, listed as what they are rather than as a
roadmap. BM25 with a judge works
well enough for daily use. The system is simple, fast, and debuggable. A
more sophisticated retrieval system would be harder to debug when it
recalled the wrong thing, and debugging recall errors is already hard.

## §23.8 Skill Discovery: The Unplanned Win

All available SKILL.md files are indexed alongside memories and documents.
This was added because skill docs are markdown files in a known directory,
and the existing source scan could include them.

The result was unexpectedly powerful. When a user asks about something a
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
that can lead to an ordinary `load_skill` call. The registry still controls
which skills are reachable and which tools may execute. A paragraph retrieved
from disk cannot grant itself a capability.

## §23.9 Taking It for a Spin

A recall demonstration needs an answer that is absent from the question and
present in an archive. Otherwise a fluent reply could be ordinary model
knowledge, or a repetition of facts supplied by the user.

From a scratch copy of `solutions/ch23`, build the CLI, then create
a fresh working directory so an old save cannot supply the answer:

```sh
go build -o /tmp/ensemble-ch16 .
recall_demo=$(mktemp -d)
cd "$recall_demo"
mkdir -p cr/memory cr/docs skills/base
```

Write `skills/base/SKILL.md`:

```markdown
---
name: base
type: primary
description: Answer questions using the context supplied to this session.
tools: list_directory
---
Answer from supplied context. Preserve decisions, their reasons, and unresolved
questions. If a requested fact is absent, say so.
```

Create the archived decision, a keyword-sharing distractor and a small
background corpus. Retired files are intentional: the memory cascade no
longer loads their original bytes, but recall must still search them.

```sh
python3 - <<'PY'
from pathlib import Path

memory = Path("cr/memory")
docs = Path("cr/docs")
fact = """## Quartz Zephyr migration decision
Project Quartz Zephyr, release QZ-731: Morgan owns the migration rehearsal on October 23. We rejected automatic retry after timeout because the remote transaction may already have committed. Read the migration ledger before deciding whether retry is safe. Ledger placement remains unresolved: application database versus separate control database. Retain rollback artifacts for nine days.
"""
(memory / "2026-10-01-1.md.retired").write_text(fact)
(docs / "quartz.md").write_text(
    "## Quartz Zephyr migration name\n"
    "Quartz Zephyr migration is also the name of our cafeteria recipe contest. "
    "Quartz Zephyr migration lunch recipes use apples and cinnamon, "
    "not database transactions.\n"
)
for i in range(25):
    (memory / f"padding-{i}.md.retired").write_text(
        f"## Garden log {i}\nPlant carrots by the apple trees and water in spring.\n"
    )
    (docs / f"padding-{i}.md").write_text(
        f"## Workshop {i}\nThe cabinet holds brushes and blue paint.\n"
    )
PY
```

With the provider key configured as in Chapter 0, start ordinary chat. Recall
uses its defaults, including the same model for the finite judge call:

```sh
LLM_VENDOR=anthropic LLM_MODEL=claude-sonnet-5 \
EN_PRIMARY_SKILL=base EN_SKILLS_DIR="$PWD/skills" \
/tmp/ensemble-ch16 chat
```

Ask:

```text
From our archived Quartz Zephyr migration discussion, tell me the release code, owner and rehearsal date, rejected approach and why, the unresolved decision, and rollback retention. Use only the context already supplied; do not call tools. If a fact is absent, say so.
```

Wait for the answer, then ask:

```text
Who owns that rehearsal, and why was automatic retry rejected?
```

The first question supplies the project name but none of the requested facts.
The follow-up is shorter than the default recall gate. It tests whether the
previously attached memory remains useful, without requiring a second search.
Check the actual request and answer: QZ-731, Morgan, October 23, the possibly
committed remote transaction, the unresolved ledger location and nine-day
retention. The cafeteria contest belongs in the judge's candidates, not in
the main answer. Check `ch02.jsonl` for `recall_attached` with its own
`recall_attached.parts`; `api.log` shows what actually reached each model.

The retained three-provider CLI runs used these two prompts and the same
planted fact. Haiku 5.5, GPT-4o mini and Gemini 3.8 Flash each answered both
turns correctly without tools. Each made one tool-free judge request followed
by two main requests. The judge saw the decision and cafeteria distractor;
only the decision reached the main requests, once in each. OpenAI's judge
wrapped `[0]` in a JSON fence, exercising the tolerant parser in actual use.
Anthropic also noted that the archive supplied no year.

Those are bounded outcomes from the earlier implementation, before the
capped-source and path corrections. The current command selects supported
Sonnet-5; the historical Haiku 5.5 result is not a recommendation to use it.
This isolated corpus demonstrates retrieval, retention and correct answers,
not general recall quality across a private archive.

The original grader returned 0/100 before it could test that behavior: its
primary skill omitted `type`, and its structural check looked for the core
recall package under the separate GUI module. Passing a launch check and
remembering a decision are different accomplishments. The revised grader
uses a typed primary and the Agent's corpus paths, with separate fixtures
from this CLI demonstration. The live answers establish the narrower and
more interesting fact that the archived reason actually came back.
