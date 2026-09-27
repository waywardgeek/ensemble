# Chapter 17: Auto-Recall

## §17.0

Chapter 16 built a memory system. The agent can save, compress, and
graduate memories through a cascade of bands. But nothing in that chapter
makes the agent *remember.* The memories sit on disk, organized and
budgeted, and the agent has no idea they exist unless it decides to search.

That is not memory. That is a filing cabinet.

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

The improvement was immediate and impossible to overstate. The agent
went from a goldfish that kept meticulous diaries to a collaborator
with a past.

## TL;DR

Build passive memory retrieval. Every user message triggers a BM25 keyword
search over the agent's memory files. A small language model filters the
results for relevance. Surviving snippets are attached to the conversation
as their own kind of entry, where they stay until a checkpoint removes them.

**Data structures.**

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
   every chunk in the index. Use per-source quotas: memory files get 50%
   of candidate slots; docs, handoffs, learnings, and skill docs split
   the rest. This prevents large directories from drowning out personal
   memories. Select up to `SmartRecallCandidates` (default 20) chunks
   above a low threshold (0.5).

2. **SLM judge.** Send the candidates to a cheap language model
   (gemini-2.5-flash-lite or equivalent) along with the last few
   conversation messages for context. The judge returns a JSON array of
   indices, the candidates it considers genuinely relevant, not just
   keyword matches. If the judge fails, times out (15 seconds), or is
   disabled, fall back to the top N results by BM25 score alone.

3. **Attachment.** Format the surviving snippets (up to `MaxSnippets`,
   default 3) as blockquoted text with source attribution. Record them as
   their own entry kind, never merged into the user's message. Cap total
   recalled text at 6KB. Skip any snippet already present in the
   conversation. The entry persists until a checkpoint removes it.

**Sources searched:**

| Source | What | Quota |
|--------|------|-------|
| Memory files | Daily logs from `~/.cr/memory/` (excluding the two most recent, already in the system prompt) | 50% of slots |
| Docs | Design documents, investigation notes | Even split |
| Handoffs | Archived context rollovers | Even split |
| Learnings | The `long` field of learnings (short ones are already in the system prompt) | Even split |
| Skill docs | All available SKILL.md files, enabling just-in-time skill discovery | Even split |

**Chunking.** Split markdown on `## ` headers. Large chunks (over 4KB) split
further on `### ` sub-headers, then on paragraph boundaries. Sub-chunk headers
use breadcrumb format: `"Parent Header > Sub Header"`.

**Tokenization.** Split on non-alphanumeric characters, lowercase, minimum
2 characters. Filter stop words from queries only. Documents keep all terms
for length normalization.

**The recall judge is one stateless model call**, not an agent. It takes a
prompt and returns text. Because it holds no history and owns no context,
it cannot recall, which makes the recursion problem structurally impossible
rather than something to guard against.

**Recall is a new spoke.** It imports the hub and the standard library,
nothing else. Recall needs a model and the model layer needs recall, so
both dependencies invert through interfaces on the hub: recall implements
`Recaller` and consumes `SnippetJudge`; the model layer implements
`SnippetJudge` and holds a nilable `Recaller`. Neither package imports
the other. Wiring happens at the top, where everything is already visible.

**Parse the judge's response defensively.** Small models are unreliable
with output format. Try `[]int` first, fall back to `[]string` (models
sometimes return `["0", "3"]`), fall back to `[]json.RawMessage` for
mixed types. Extract the first `[` to last `]` from the response,
models often wrap the JSON in explanation text.

**Configuration** (defaults that work in production):

| Setting | Default | Purpose |
|---------|---------|---------|
| `MinScore` | 3.0 | BM25 score floor for plain recall (no judge) |
| `MaxSnippets` | 3 | Maximum snippets injected per turn |
| `MinQueryLength` | 80 | Skip recall for short messages |
| `JudgeModel` | gemini-2.5-flash-lite | Cheapest model that judges well |
| `Candidates` | 20 | BM25 results sent to the judge |
| `ContextMsgs` | 3 | Recent messages for judge context |
| `MaxRecallBytes` | 6144 | Hard cap on total recalled text |

**Graded behaviors (10 checks, 100 points):**

| Check | Points | What it tests |
|-------|--------|---------------|
| `bm25-indexes` | 10 | Memory files are indexed and searchable |
| `bm25-scores` | 10 | Relevant chunks score higher than irrelevant ones |
| `bm25-stop-words` | 5 | Stop words are filtered from queries, not documents |
| `chunking-splits` | 10 | Large documents are split on markdown headers |
| `per-source-quota` | 10 | Memory files get at least 50% of candidate slots |
| `judge-filters` | 15 | The SLM judge removes irrelevant BM25 hits |
| `judge-fallback` | 10 | Judge failure falls back to top-N BM25 results |
| `recall-is-a-spoke` | 10 | `internal/recall` imports only the hub and stdlib |
| `recall-is-own-kind` | 10 | Recalled text is its own entry, never merged into the user message |
| `injection-capped` | 10 | Total recalled text does not exceed MaxRecallBytes |

## §17.1 The Idea in Plain Words

A filing cabinet is not memory. Memory is what comes to mind without being
asked.

Chapter 16 built the filing cabinet: bands that compress, a cascade that
graduates, a watermark that forces the agent to write things down before
they disappear. All of that is storage. It answers the question "where
do memories go?" It does not answer the question "how do memories come
back?"

The answer is the oldest trick in information retrieval. When the user
types a message, pull out the keywords, score every memory chunk against
them, and surface the best matches. BM25, a formula from 1994, does
this in microseconds on a corpus of a few hundred files. No vector
database, no embedding model, no infrastructure. Keywords and arithmetic.

But BM25 is a blunt instrument. It matches words, not meaning. A query
about "authentication" will not find a memory about "login flow." And it
is biased toward long documents full of technical vocabulary. A design
doc that mentions "memory" seventeen times will outscore a short daily
log that contains the actual decision. Two mechanisms compensate:

**Per-source quotas.** Memory files get 50% of candidate slots regardless
of score. Docs, handoffs, learnings, and skill docs split the rest. This
is a structural guarantee: the agent's personal memories are never drowned
out by the project's reference documentation, even when the docs score
higher.

**The recall judge.** A cheap language model, the smallest that can follow
instructions, reads the BM25 candidates alongside the recent conversation
and picks the ones that are genuinely relevant, not merely keyword-adjacent.
The judge sees context that BM25 cannot: what the conversation is about,
what the user is trying to do, whether a keyword match is superficial or
load-bearing.

The judge is the difference between "here are some memories that mention
the same words" and "here is something you need to know right now." Without
it, auto-recall is retrieval. With it, auto-recall is recall.

The obvious lifetime for a recalled snippet is ephemeral. Relevance is local:
a memory about database schema is useful when the user mentions the database
and noise when the conversation moves to the GUI. Show it, then drop it.

That intuition is expensive, and the reason is that a prompt cache is a
prefix property. Dropping a block deletes bytes from the middle of the
prefix, which invalidates everything after the deletion point. Keeping the
block at the tail instead avoids the deletion but re-sends those bytes on
every tool round, and a turn has many rounds. Attaching the block once and
leaving it alone is the only option where the prefix grows monotonically and
the bytes are paid for a single time. Rounds are many; turns are few.
Anything re-sent per round is the expensive thing.

So recalled snippets are permanent. They are not merged into the user's
message, because the user's words must stay distinguishable from
machine-retrieved text, and because anything fused into a user turn cannot
later be removed without damaging it. They are their own kind of entry,
which makes them addressable: a checkpoint can delete them by kind.

That checkpoint is where removal belongs. A checkpoint already rewrites the
prefix, so it is already paying for a cache miss, and deleting recall there
costs nothing extra. Removing it on its own schedule would buy a second
invalidation for no benefit. The rule generalizes past this chapter:
**delete only at a moment that is already paying for a cache miss.**

This is the second appearance of a shape worth naming. Chapter 16 found that
switching a memory band off is more expensive than leaving it on, because
disabling strips bytes from the front. Here, the transient-looking thing is
cheaper to keep than to remove. In a system built on cached prefixes,
retention is cheap and removal is dear, so the intuitive lifetime is usually
the wrong one.

The system degrades gracefully at every level. If the query is too short
(under 80 characters), recall does not fire. Short messages like "fix it"
contain no signal worth matching. If BM25 finds nothing above the score
floor, nothing is injected. If the judge fails or times out, the system
falls back to the top BM25 results without filtering. If the entire recall
system is disabled, the agent still works. It does not remember
passively. Every failure mode produces a working agent with less context,
never a broken one.

One unexpected benefit: skill discovery. All available SKILL.md files are
indexed alongside memories. When a user asks about something a skill
handles, the skill documentation surfaces through recall, and the agent
loads it without being told. This was not planned. It was added because
skill docs were already markdown files in a known directory, and indexing
them cost nothing. It turned auto-recall into a discovery mechanism as well
as a memory mechanism.

## §17.2 BM25: The Thirty-Year-Old Formula That Still Works

BM25 (Best Matching 25) was published by Robertson and Walker in 1994.
It is still the default ranking function in Elasticsearch, Solr, and most
search engines that have not switched to neural retrieval. For a corpus of
a few hundred markdown files totaling a few megabytes, it is the right
tool.

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
so implementations clamp it, and the clamp produces exactly zero for every
such term. On a small archive, which is every archive on day one, most terms
appear in more than half the chunks. Every one of them scores zero, the
totals collapse below any sensible threshold, and retrieval returns nothing.

The failure is worth dwelling on because of how it presents. Nothing errors.
The index builds, the search runs, the scores are computed correctly
according to the formula, and the agent simply never remembers anything. It
reads as "recall is broken" when it is in fact "IDF is the wrong variant,"
and those two diagnoses send you to opposite ends of the codebase. The
smoothed form is non-negative by construction, which makes the clamp
unreachable and the failure impossible.

A related trap waits in testing. IDF measures rarity *against the corpus*, so
in an archive of two or three chunks nothing is rare, every score collapses
toward zero, and a perfectly correct implementation looks broken. A three
chunk fixture scores around 1.7 in total, well under a threshold of 3.0. Any
worked example or test fixture needs roughly twenty five chunks of padding
before the numbers mean anything. This is the same failure as the IDF variant
above, arriving from a different direction: both make correct code look
broken, which is the most expensive kind of wrong.

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
configuration files, no external services. Build the index at startup,
rebuild it when memory files change, and search it in microseconds.

## §17.3 Per-Source Quotas: Why Memory Gets Half the Slots

The first version of auto-recall ran a single merged BM25 search over all
sources. Design documents dominated the results. A 20KB design doc full of
technical terminology would outscore a 500-byte daily log that contained
the relevant decision, because BM25 rewards term density and the design
doc had more surface area for keyword matches.

The fix is structural, not algorithmic. Give each source its own BM25 index
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
only ever discards its weakest hits, which were never going to be selected
anyway. Implement it that way and the quota is dead code: delete it entirely
and not one recalled snippet changes.

**The quota must not overrule the judge.** It governs the mechanical paths,
where no judge ran or the judge failed and selection falls back to score
order. When a judge has actually assessed relevance, its verdict stands.
Discarding the opinion of the only component that read the conversation, in
order to satisfy a ratio, would trade the good signal for the crude one.

This is a policy decision, not a BM25 improvement. The agent's personal
memories are the highest-value source. They contain decisions, context,
and lessons that no design doc captures. Guaranteeing them half the slots
ensures they survive competition with larger, denser documents.

## §17.4 The Judge: Filtering by Meaning, Not Keywords

BM25 finds documents that share words with the query. The recall judge
finds documents that share *relevance* with the conversation.

The judge is a single stateless call to the cheapest model that can follow
instructions, currently Gemini 2.5 Flash Lite, a model that costs
fractions of a cent per call. It receives:

1. The recent conversation (last 3 user and assistant messages).
2. The latest user message, labeled separately.
3. Up to 20 numbered candidate snippets from BM25.
4. An instruction: pick up to 3 that the agent would genuinely benefit
   from seeing, return a JSON array of indices. "Keyword mention alone
   is not relevance."

The judge returns `[0, 7, 12]` or `[3]` or `[]`. Three details make
this work reliably:

**It starts clean every time.** The judge holds no history. Note what that
means about item 1 above: the recent conversation reaches the judge as *data
inside the prompt*, not as history the judge carries between calls. Every
judgment is independent, and nothing from the last one can color the next.

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
indices are silently filtered.

**One component owns the reply format, and it is the one that parses.**
Nothing else may restate it. This sounds like housekeeping and is the most
dangerous rule in the chapter, because the model layer holds the client and
will feel like the natural place to describe what a good answer looks like,
while the recall layer holds the parser. Let both describe it and they will
drift. Then the model receives two specifications in one request, obeys one
of them, and the parser rejects a reply that was perfectly well formed by the
other. Recall falls back to raw scores on every single turn, with no error,
no log line, and a judge that appears from the outside to be working
normally. The only symptom is that results are slightly worse than they
should be, forever. Whoever reads the bytes decides what the bytes look like.

When the judge fails or times out (15 seconds), the system falls back
to the top N BM25 results by score. The agent gets noisier recall, not
no recall.

## §17.5 Attachment: Permanent Until a Checkpoint

Recalled snippets are formatted as blockquoted text with source attribution:

```
[Auto-recalled memories]

From 2026-09-22-1.md — Key decisions:
> save_memory removed as agent tool; micro_handoff becomes SOLE trigger.
> Sub-agent compresses portions into memories.

From docs/auto-recall-design.md — The Pipeline:
> BM25 keyword search over memory files, filtered through an SLM judge.
```

That text becomes its own kind of entry in the conversation. Three properties
follow, and each one is load-bearing.

**It is never merged into the user's message.** The user's actual words must
stay distinguishable from machine-retrieved text, both for the model reading
them and for anyone auditing the log afterward. Fusing the two also makes the
recall unremovable: a checkpoint keeps user turns verbatim, so anything
welded into one survives with it forever.

**Where it lands on the wire is the renderer's decision, not the
conversation's.** The conversation records that recall happened and what
bytes it produced. Turning that into a request is a vendor question, and
vendors disagree: one wants a system block mid-conversation, another a
role-tagged message, another an extra part on a turn that already exists.
Baking a wire shape into the stored conversation is the mistake Chapter 2
exists to prevent. The renderer owes one guarantee in return: a stable
position, so the prefix grows monotonically and stays cached.

**It carries its bytes.** Recall is nondeterministic, because a language
model chose which snippets survived. If the text were recomputed while
rendering, replaying a log would call a model, and two replays of the same
log could disagree. So the snippets are recorded as text at the moment they
are chosen. Replay then reproduces the conversation exactly, with no search
and no model call.

That last property is the same boundary Chapter 16 drew. A transform that is
a pure function of the log can run at render time. Anything nondeterministic
has to be recorded when it happens. Recall is the second thing in this book
to land on the far side of that line, which is why Chapter 15 had to come
first: without recorded events, there is nowhere to put it.

Because the entries persist, the same memory could be recalled on three
separate turns and appear three times. Skip any snippet whose content is
already in the conversation. Retrieval should surface what the agent does not
already have in front of it.

One wiring hazard deserves naming, because it costs hours and announces
nothing. An agent of this shape has more than one way to begin a turn. There
is the interactive path, where a prompt arrives on standard input and is
handled synchronously, and there is the actor path, where a message arrives
in the mailbox and is handled by the loop. The graphical client uses the
second. So does any harness that drives the agent like a real client. The two
do not delegate to one another.

Hook recall into only one of them and the result is an agent whose memory
works perfectly whenever it is poked by hand, and never runs once in the
actual product. No error is raised, because nothing is wrong: a function that
was never called cannot complain. Find every entry point into a turn and
attach to all of them, then verify through the path a real client uses rather
than the one that is convenient to test.


## §17.6 Tuning: What the Numbers Mean

The defaults in the TL;DR were not designed. They were tuned over months
of daily use. Each one has a story.

**MinScore went from 0.5 to 3.0.** At 0.5, every conversation triggered
recall, and most results were noise: single keyword matches on common
technical terms. At 3.0, recall fires only when BM25 finds a genuine
multi-term match. The jump is large because BM25 scores are not
normalized; the right threshold depends on corpus size and vocabulary.
Note: this threshold applies only to plain recall (no judge). When the
judge is enabled, BM25 uses a low threshold (0.5) to cast a wide net,
and the judge handles false-positive filtering.

**MinQueryLength went from 15 to 50 to 80.** "Fix it," "run tests,"
and "looks good" contain no signal for keyword matching. Raising the
threshold progressively eliminated recall on messages that could never
produce useful results. At 80 characters, the user has typed enough to
contain a searchable concept.

**Candidates went from 10 to 80 (code default) to 20 (production).**
The code default is conservative: miss nothing, let the judge sort it
out. In practice, 20 candidates is enough for the judge to work with
and keeps its prompt under a few thousand tokens. More candidates
means a larger judge prompt, which means higher cost and latency for
marginal relevance improvement.

**Judge timeout is 15 seconds, not 2.** The original design assumed a
local model (Ollama) with sub-second latency. Cloud API calls to Gemini
Flash Lite typically take 1–3 seconds. The generous timeout exists
because a late judge response is better than no judge response. The
fallback to unfiltered BM25 results is noisier.

## §17.7 What BM25 Cannot Do

BM25 matches words, not meaning. Three limitations follow:

**Vocabulary mismatch.** If the user says "authentication" and the memory
says "login flow," BM25 will not connect them. Vector search (embedding-based
retrieval) would, but it requires an embedding model, an index to maintain,
and a service to call. For a corpus of a few hundred files, the marginal
improvement does not justify the infrastructure. If the corpus grows by
orders of magnitude, revisit this.

**No content deduplication.** If a memory and a design doc contain the same
information in different words, both may be recalled, wasting snippet slots.
The only deduplication is file-level: the two most recent daily logs are
excluded because they are already in the system prompt.

**No relevance feedback.** The system does not learn which recalled memories
the agent actually used. A snippet that the agent ignores on ten consecutive
turns will keep surfacing if it scores well on BM25. Tracking usage patterns
and adjusting retrieval is a natural next step that this implementation does
not take.

These are honest limitations, not roadmap items. BM25 with a judge works
well enough for daily use. The system is simple, fast, and debuggable. A
more sophisticated retrieval system would be harder to debug when it
recalled the wrong thing, and debugging recall errors is already hard.

## §17.8 Skill Discovery: The Unplanned Win

All available SKILL.md files are indexed alongside memories and documents.
This was added because skill docs are markdown files in a known directory,
and adding them to the BM25 corpus cost one line of code.

The result was unexpectedly powerful. When a user asks about something a
skill handles ("can you check my email?" or "what's on my calendar?"),
the skill's documentation surfaces through auto-recall. The agent sees the
snippet, recognizes the available skill, and loads it. No explicit skill
search, no user instruction to load a specific skill. The agent discovers
capabilities through the same mechanism it uses to recall past work.

This turns auto-recall into a discovery system as well as a memory system.
The agent's capabilities are not fixed at startup; they expand when the
conversation moves into a domain where a skill exists. The user does not
need to know which skills are installed. The skills surface when they are
relevant.

## §17.9 The Existing Code

The CodeRhapsody codebase contains a complete, production-tested
implementation of everything in this chapter. The coder should read it
for inspiration:

- `internal/memory/bm25.go` (BM25 engine, 386 lines): tokenization,
  stop words, indexing, scoring, chunking. Note: an off-by-one error in
  this file went undetected for weeks and recall quality improved
  noticeably when it was fixed. Test your scoring carefully.
- `internal/memory/recall.go` (recall pipeline, 483 lines): source
  loading, per-source quotas, formatting, injection.
- `internal/agent/smart_recall.go` (judge, 195 lines): prompt
  construction and defensive JSON parsing.
- `internal/skills/builtin/recall-judge/SKILL.md` (the judge's skill
  definition, 25 lines).

Read it for the problems it solved, then check which of them exist here.
Per-source quotas and defensive parsing transfer directly. Others do not:
CodeRhapsody runs its judge as a persistent sub-agent, so it needs history
resets, a disabled recall provider, and settings flags to keep the judge from
recalling while it judges. A stateless judge has none of those problems,
because it has no context to fill.

That is worth taking as a general caution about reference code. An
implementation shows you the problems its own architecture created alongside
the problems inherent to the task, and it cannot tell you which is which.
Copying the shape imports both.
