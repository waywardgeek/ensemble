# Auto-Recall: Passive Memory Retrieval in CodeRhapsody

*Reference document for the Ensemble book chapter on auto-recall.*
*Audited from source code on 2026-09-27. Every claim cites file:line.*

---

## What Auto-Recall Is

Auto-recall is a passive retrieval system that surfaces relevant memories
before the LLM starts thinking. When a user message arrives, it runs BM25
keyword search over the agent's memory files, design documents, skill
documentation, and learnings — then optionally filters those results through
a small language model (the "recall judge") to eliminate false positives.
The surviving snippets are injected into the conversation as ephemeral
context that the agent sees but that is never persisted to history.

The agent experiences recalled memories as if they "came to mind" — no
tool call needed, no decision to search. This fills the gap between
always-loaded context (MEMORY.md, the two most recent daily logs,
learnings injected into the system prompt) and manual search (the
`search_knowledge` tool the agent can call explicitly).

---

## The Pipeline

```
User message arrives
        │
        ▼
┌─────────────────────┐
│  Gate: too short?    │──── skip if len(query) < MinQueryLength (default 80)
└─────────────────────┘
        │
        ▼
┌─────────────────────┐
│  BM25 search         │  Per-source quotas: memory 50%, rest split evenly
│  across all sources   │  Low threshold (0.5) to cast a wide net
│  (up to Candidates)  │  Default: 20 candidates (Bill's setting)
└─────────────────────┘
        │
        ▼
┌─────────────────────┐
│  SLM Judge           │  Persistent sub-agent (gemini-2.5-flash-lite)
│  (if enabled)        │  Sees conversation context + numbered snippets
│  15s timeout         │  Returns JSON array of approved indices
└─────────────────────┘
        │
        ▼
┌─────────────────────┐
│  Format & inject     │  Up to MaxSnippets (default 3), capped at 6KB
│  as ephemeral text   │  Injected as "[Auto-recalled memories]" block
└─────────────────────┘
```

If the judge is disabled, unavailable, or times out, the pipeline falls
back to injecting the top N BM25 results by score directly.

---

## Where It Hooks In

Auto-recall runs inside the **ephemeral provider** — a callback registered
during agent construction that fires on every user message.

**`internal/agent/agent.go:800`** — The ephemeral provider closure:
```go
agent.ephemeralProvider = func(userText string) *common.EphemeralContent {
```

The provider does three things in order:
1. Injects current time (line 803)
2. Runs any custom ephemeral provider (e.g., sender identity for Puffin) (line 808)
3. Runs auto-recall if the `tool-memory` skill is active (line 817)

The results are concatenated and returned as `EphemeralContent`. The AI
client injects this text into the API request alongside the user message
but does not persist it to conversation history. On the next turn, new
recall results replace the old ones.

**Gating**: Auto-recall only fires when `common.HasActiveSkill(agent, "tool-memory")`
returns true (`agent.go:817`). Sub-agents spawned without memory skills
(like the recall judge itself) never trigger recall.

---

## BM25 Implementation

The BM25 engine is built from scratch in `internal/memory/bm25.go` (386 lines).
It was **not** extracted from `semantic_search` as originally planned — instead,
a purpose-built implementation was written with standard tuning parameters
(k1=1.2, b=0.75, `bm25.go:14-15`).

### Tokenization

- Splits on non-alphanumeric characters, lowercased (`bm25.go:86-92`)
- Minimum token length: 2 characters (`bm25.go:90`)
- **Stop words filtered from queries only** (`bm25.go:105-114`) — document
  indexing keeps all terms. This is a deliberate asymmetry: stop words in
  documents contribute to term frequency normalization but don't inflate
  query-side IDF.
- 55 common English stop words (`bm25.go:59-80`)
- Fallback: if stop-word removal empties the query, falls back to unfiltered
  tokenization (`bm25.go:321-325`)

### Chunking

Markdown documents are split by `## ` headers (`bm25.go:120-160`). Each chunk
records its filename, header text, content, pre-computed term frequencies, and
token length.

Large documents (design docs, handoffs) use `ChunkMarkdownWithMaxSize` with a
4KB limit (`recall.go:122`). Oversized chunks are further split on `### `
sub-headers, then on paragraph boundaries (`bm25.go:164-207`). Sub-chunk headers
use breadcrumb format: `"Parent Header > Sub Header"` (`bm25.go:189`).

### History chunking

A separate `ChunkHistory` function (`bm25.go:214-280`) handles v2 history format,
splitting on `### USER`, `### ASSISTANT`, etc. block markers. `TOOL_RESULT` and
`TOOL_ERROR` blocks are skipped entirely — they're noise for retrieval. Session
boundaries (`## Session ...`) are tracked and prepended to chunk headers for
context.

### Scoring

Standard BM25 with IDF floored at 0 to prevent negative scores from very
common terms (`bm25.go:336`). Results are sorted by score descending using
insertion sort (justified by small result sets, `bm25.go:342-347`). Snippets
are truncated to 800 characters at word boundaries (`bm25.go:366`).

---

## Memory Sources Searched

### 1. Memory files (`~/.cr/memory/*.md`)

Daily memory logs. The two most recent are **excluded** (they're already
injected into the system prompt). Summary files (`summary-*.md`) are also
excluded. (`recall.go:75-108`)

Memory gets **50% of candidate slots** in the per-source quota system
(`recall.go:226`). This is the most important source — it's the agent's
episodic memory.

### 2. Docs directory (`{dataDir}/docs/*.md`)

Project-specific design documents, sprint plans, investigation notes.
Included when the directory exists and `SmartRecallSearchDocs` is true.
Files are chunked with 4KB max chunk size. Filenames are prefixed with
the directory basename for disambiguation (e.g., `docs/auto-recall-design.md`).
(`recall.go:111-127`)

### 3. Handoffs directory (`{dataDir}/handoffs/*.md`)

Archived handoff documents from previous context rollovers. Same chunking
rules as docs. (`agent.go:831-832`)

### 4. Learnings (long form) (`learnings.json`)

The `long` field of each learning (global and skill-scoped) is loaded as a
searchable chunk. Short learnings (already in the system prompt) are not
duplicated. (`recall.go:130-153`)

### 5. Skill docs (`[SKILL DOC: name]`)

All available SKILL.md files are chunked and added as `ExtraChunks` — this
enables just-in-time skill discovery. When the user asks about something a
skill handles, the skill doc can surface through recall and prompt the agent
to `load_skill`. (`agent.go:859-866`)

This was **not in any design doc** — it was added organically and is one of
the system's more powerful features.

---

## Per-Source Quota System

The `searchPerSource` function (`recall.go:170-248`) runs independent BM25
searches per source and allocates guaranteed slots to prevent large sources
(like `docs/` with many files) from drowning out memory hits.

**Allocation algorithm:**
1. Memory gets 50% of total candidate slots (rounded up)
2. Remaining slots are split evenly among extra sources (docs, handoffs,
   learnings, skill docs)
3. If only memory exists, it gets 100% of slots
4. **Unused slot redistribution**: After the first pass, any unused slots
   (from sources with fewer results than their quota) flow to other sources
   that have overflow results (`recall.go:240-260`)

This was a significant divergence from both design docs, which proposed
either a simple merged search (auto-recall design) or a `searchAllSources`
function (smart recall plan). The per-source guarantee was added to solve
a real problem: docs/ files, being large and keyword-dense, would otherwise
consume all candidate slots.

---

## The Recall Judge

### Architecture

The judge is a **persistent sub-agent** (`smart_recall.go:32-48`), not a sync
spawn as originally planned. It's created lazily on first use and reused across
turns, with conversation history reset before each evaluation
(`smart_recall.go:55`). This avoids the overhead of spawning a new agent on
every user message.

**Critical detail**: The judge's `EphemeralProvider` is set to nil
(`smart_recall.go:44-48`). Without this, the judge would trigger its own
auto-recall, which would spawn another judge, creating infinite recursion.

### Skill

The recall-judge skill (`internal/skills/builtin/recall-judge/SKILL.md`) is
deliberately minimal — 25 lines, no tools, no dependencies:

```
You pick the most useful memory snippets for a conversation.
...
You MUST respond with ONLY a JSON array of integer indices, most useful first.
```

Its `settings.json` disables both `smartRecallEnabled` and `autoRecallEnabled`
to prevent the recursive recall problem from the settings side as well (belt
and suspenders with the nil EphemeralProvider).

### Prompt Construction

`buildJudgePrompt` (`smart_recall.go:68-86`) sends:

1. **Task instruction**: Pick up to N snippets, return JSON array. Emphasizes
   that keyword mention alone is not relevance.
2. **Conversation context**: Recent user and assistant messages (default 3 of
   each), plus the latest user message labeled as "USER (latest)".
3. **Numbered candidates**: Each candidate shows index, filename, section
   header, and content truncated to 500 characters.
4. **Answer prompt**: Explicit instruction for JSON format with examples.

### Response Parsing

`parseJudgeResponse` (`smart_recall.go:92-141`) is deliberately over-tolerant
to handle SLM quirks:

1. Extracts the first `[` to last `]` from the response (SLMs often add
   explanation text around the JSON)
2. Tries parsing as `[]int`
3. Falls back to `[]string` (SLMs sometimes return `["0", "3"]` instead
   of `[0, 3]`)
4. Falls back to `[]json.RawMessage` for mixed types
5. Each fallback attempts `fmt.Sscanf` to parse string values as integers

Out-of-range indices are silently filtered. Duplicates are deduplicated
using a `seen` map (`recall.go:388-394`).

### Model

The default judge model is `gemini-2.5-flash-lite` (`smart_recall.go:19`).
The original plan specified `ollama/gemma3:1b` — but the system evolved to use
Google's lightweight model instead, which doesn't require a local Ollama
installation. Bill's production setting confirms this: `"smartRecallJudgeModel": "gemini-2.5-flash-lite"`.

### Timeout

15 seconds (`smart_recall.go:16`), not 2 seconds as originally planned. The
2-second timeout was unrealistic for a cloud API call. On timeout, the system
falls back to top-N BM25 results.

---

## Configuration

All settings live in `internal/common/settings.go:262-278` as the
`MemorySettings` struct.

### Production Defaults (Bill's Settings)

| Setting | Code Default | Bill's Setting | Purpose |
|---------|-------------|----------------|---------|
| `autoRecallEnabled` | `true` | `true` | Master switch |
| `autoRecallMinScore` | `3.0` | `3` | BM25 score floor for plain recall (high = strict) |
| `autoRecallMaxSnippets` | `3` | `3` | Maximum snippets injected |
| `autoRecallMinQueryLength` | `50` | `80` | Minimum user message length to trigger |
| `smartRecallEnabled` | `false` | `true` | Enable the SLM judge pipeline |
| `smartRecallJudgeModel` | `"gemini-2.5-flash-lite"` | `"gemini-2.5-flash-lite"` | Judge model |
| `smartRecallCandidates` | `80` | `20` | BM25 candidates sent to judge |
| `smartRecallContextMsgs` | `3` | `3` | Recent messages for judge context |
| `smartRecallSearchDocs` | `true` | `true` | Include docs/ in search |
| `smartRecallSearchHandoffs` | `true` | `true` | Include handoffs/ in search |

*Source: `settings.go:363-377` for code defaults.*

**Notable divergences between code defaults and Bill's production settings:**
- `smartRecallEnabled`: Code defaults to `false` (the comment says "Off until
  Ollama is detected" — a remnant of the Ollama-first plan). Bill enables it.
- `smartRecallCandidates`: Code defaults to `80`, Bill uses `20`. The wide
  net (80) was probably the initial "let's not miss anything" value; 20
  candidates is enough for the judge to work with and keeps the prompt smaller.
- `autoRecallMinQueryLength`: Code defaults to `50`, Bill uses `80`. Short
  messages like "fix it" or "run the tests" don't benefit from recall.

### Hard-Coded Constants

| Constant | Value | Location | Purpose |
|----------|-------|----------|---------|
| `MaxRecallBytes` | 6144 (6KB) | `recall.go:13` | Hard cap on total recalled text |
| `k1` | 1.2 | `bm25.go:14` | BM25 term frequency saturation |
| `b` | 0.75 | `bm25.go:15` | BM25 document length normalization |
| `smartRecallTimeout` | 15s | `smart_recall.go:16` | Judge response deadline |
| Snippet truncation | 800 chars | `bm25.go:366` | Max length per snippet in results |
| Judge snippet truncation | 500 chars | `smart_recall.go:79` | Max length per candidate in judge prompt |
| Large chunk max | 4096 bytes | `recall.go:122` | Max chunk size for docs/handoffs |
| Judge min BM25 score | 0.5 | `recall.go:366` | Low threshold for wide candidate net |

---

## Injection Format

Recalled memories are injected as a single text block with this format:

```
[Auto-recalled memories]

From 2026-03-22-1.md — What We Built:
> Built Lobby Platform MVP — social gaming platform, humans + AI agents...
> Architecture: Platform is a relay, not a host.

From docs/smart-recall-design.md — Phase 3:
> The SLM judge filters BM25 candidates using conversation context...
```

Each snippet is blockquoted with `> ` prefix. The header shows filename and
section. The total block is capped at `MaxRecallBytes` (6KB) — snippets
that would push past this limit are silently dropped (`recall.go:307-309`,
`recall.go:420-422`).

This text arrives as part of the `EphemeralContent` returned by the ephemeral
provider. The AI client concatenates it with other ephemeral content (current
time, custom provider output) and injects it alongside the user message.
It is **not** part of the system prompt and is **not** persisted to history.

---

## Relationship to search_knowledge

The `search_knowledge` tool (`internal/memory/search_knowledge.go`,
`internal/memory/search.go`) shares the same BM25 infrastructure but is
an **active** tool the agent calls explicitly. It searches the same memory
files, docs, and history, and returns results the agent can act on.

Auto-recall is passive — it fires automatically on every user message
without the agent deciding to search. The two complement each other:
auto-recall handles "memories that should come to mind," while
`search_knowledge` handles "I need to look this up."

---

## What Was Planned but Never Built

1. **First message bonus** (auto-recall design doc): "On the first real user
   message of a session, could recall more aggressively (5 snippets instead
   of 3)." Never implemented — all messages use the same `MaxSnippets`.

2. **Extract search core from semantic_search** (auto-recall design, Phase 1):
   "Refactor `pkg/tools/semantic_search.go` to expose a `SearchFiles` function."
   Instead, a completely new BM25 engine was built from scratch in `bm25.go`.
   The semantic_search tool and auto-recall share no code.

3. **Ollama as default judge provider** (smart recall plan): The entire Phase 2
   of the implementation plan was about adding an Ollama provider. The system
   shifted to using Gemini Flash Lite instead, which is simpler (no local model
   management) and more reliable.

4. **Vector/hybrid search**: The original design noted "Full hybrid (vector +
   BM25): Over-engineering for our corpus size. Skip." It was never revisited.
   Pure BM25 has proven sufficient for the corpus size (dozens to low hundreds
   of markdown files).

5. **Deduplication against system prompt content** (auto-recall design):
   "Before injecting a snippet, check that its content doesn't substantially
   overlap with text already in MEMORY.md." The only deduplication implemented
   is file-level exclusion of the two most recent daily logs. No content-level
   overlap detection exists.

6. **Sync sub-agent for judge** (smart recall plan): The plan specified spawning
   a sync sub-agent per evaluation. The implementation uses a persistent agent
   with history reset instead — more efficient for repeated evaluations.

---

## What Was Built Differently Than Planned

1. **BM25 from scratch vs. extracted from semantic_search**: The plan said
   reuse; the implementation built new. This was the right call — semantic_search
   is a tool handler with ripgrep integration, while auto-recall needed a
   pure Go BM25 library for server-side use.

2. **Per-source quotas**: Neither design doc mentioned source-aware allocation.
   The implementation added `searchPerSource` with 50% memory guarantee after
   discovering that docs/ files dominated candidate lists.

3. **MaxRecallBytes = 6KB vs. 3KB planned**: The original design said "Hard cap:
   3KB of recalled memory text per message." Production uses 6KB. (The original
   was likely too conservative.)

4. **MinQueryLength = 50/80 vs. 15 planned**: The original design said "Skip
   conditions: Message shorter than 15 characters." Code defaults to 50; Bill
   uses 80. Short messages rarely benefit from recall.

5. **15s judge timeout vs. 2s planned**: The plan's 2-second timeout was based
   on Ollama (local inference). With a cloud model, 15 seconds is realistic.

6. **Persistent judge agent vs. sync spawn**: Lower overhead per evaluation,
   and the history reset pattern avoids context accumulation.

7. **Skill doc injection as ExtraChunks**: Not in any design doc. Enables
   just-in-time skill discovery through recall.

8. **Learnings as a searchable source**: `LoadChunksFromLearnings` was not
   planned. It makes the `long` field of learnings discoverable through recall.

9. **Judge prompt includes maxSnippets**: The judge function signature includes
   `maxSnippets` (`func(context string, candidates []Chunk, maxSnippets int)`),
   allowing the prompt to tell the judge exactly how many to pick. The original
   plan's signature didn't pass this.

10. **Robust JSON parsing with 3-level fallback**: The plan mentioned basic JSON
    parsing. The implementation handles `[]int`, `[]string`, and `[]json.RawMessage`
    because small models are unreliable with output format.

---

## What Was Added Without Being in Any Design Doc

1. **Stop word filtering** on queries (`bm25.go:59-114`): Improves BM25
   precision by removing common English words from the query while keeping
   them in the document index.

2. **ChunkHistory** for v2 history format (`bm25.go:214-280`): Enables
   searching conversation history, filtering out noisy TOOL_RESULT blocks.

3. **Recursive EphemeralProvider disabling** (`smart_recall.go:44-48`):
   Prevents infinite recursion when the judge agent would otherwise trigger
   its own auto-recall.

4. **The entire `search_knowledge` tool**: Shares BM25 infrastructure to
   provide active search as a complement to passive recall.

5. **Time injection**: The ephemeral provider always injects current time
   (`agent.go:803-805`), giving agents temporal awareness for relative time
   references in memories.

---

## Tuning History and Lessons Learned

### MinRecallScore: 0.5 → 3.0

The original threshold was set low to avoid missing relevant results. In
practice, BM25 scores below 3.0 are almost always noise — single keyword
matches on common terms. Raising to 3.0 dramatically reduced false positives
without losing relevant results. (For the judge pipeline, the BM25 threshold
is kept at 0.5 because the judge handles false-positive filtering.)

### MinQueryLength: 15 → 50 → 80

Short messages like "fix it," "run tests," or "looks good" don't contain
enough signal for keyword matching. The threshold was raised progressively
as we observed recall triggering on messages that couldn't produce useful
results.

### SmartRecallCandidates: 10 → 80 (code) → 20 (production)

The original plan proposed 10. The code default was widened to 80 to ensure
the judge sees everything possibly relevant. In practice, Bill uses 20 — a
balance between coverage and judge prompt size. With 20 candidates, the
judge prompt stays well under the model's context window.

### Persistent judge vs. sync spawn

The first implementation likely spawned a fresh agent per evaluation. The
persistent agent pattern with history reset was adopted for efficiency:
agent construction is expensive (loading skills, settings, creating AI
clients), while history reset is cheap.

### Why BM25 works at this scale

With a corpus of dozens to low hundreds of markdown files totaling a few
megabytes, BM25's simplicity is an advantage. Building and searching the
index takes milliseconds. Vector search would require an embedding model
(another API call per message), an index to maintain, and more complexity
— all for marginal relevance improvement on a corpus this size.

The system will need to evolve if the corpus grows by orders of magnitude,
but that hasn't happened yet.

---

## Honest Assessment

### What Works Well

- **Passive surfacing is transformative.** Before auto-recall, the agent
  had to remember to search. Now relevant memories appear automatically.
  This is the single highest-value feature in the memory system.

- **The judge eliminates noise.** Without the judge, BM25 keyword matches
  are often superficial ("the user mentioned 'memory,' here's every memory
  about memory"). The SLM judge understands conversational context and
  filters to genuinely useful snippets.

- **Skill discovery through recall** is an unexpected win. When a user asks
  about something a skill handles, the skill doc surfaces through recall,
  prompting the agent to load it. This replaced an explicit skill-search step.

- **Graceful degradation.** If the judge fails, times out, or is disabled,
  the system falls back to plain BM25. If BM25 finds nothing relevant,
  nothing is injected. The system never blocks the conversation.

- **Per-source quotas solve a real problem.** Without them, a large docs/
  directory would consume all candidate slots, pushing out personal memories
  that are often more relevant.

### Known Limitations

- **No semantic understanding.** BM25 is keyword matching. If the user asks
  "how do we handle authentication?" but the memory uses "login flow," the
  system won't make the connection. This is the fundamental limitation of
  BM25 vs. vector search.

- **No content-level deduplication.** If a memory and a doc contain the same
  information in different words, both may be recalled, wasting snippet slots.
  The only deduplication is file-level exclusion of recent daily logs.

- **Judge adds latency.** The judge call takes 1-3 seconds typically (cloud
  API to Gemini Flash Lite). This delays every response. The 15-second timeout
  is generous; most calls complete much faster.

- **Judge cost.** Every user message triggers a judge call (if the BM25 stage
  finds candidates). At high message volumes, this adds up. The cost is small
  per call (Flash Lite is cheap), but it's non-zero.

- **No relevance feedback loop.** The system doesn't learn which recalled
  memories the agent actually used. There's no mechanism to improve retrieval
  quality over time based on usage patterns.

- **Query is latest message only.** The BM25 query uses only the latest user
  message (`recall.go:357`), not the conversation context. This means if the
  conversation has been about topic X for several turns but the latest message
  is a short follow-up, the query may not capture enough signal. The judge
  gets conversation context, but BM25 doesn't.

- **Fixed snippet budget.** MaxRecallBytes (6KB) and MaxSnippets (3) don't
  adapt to conversation length or context window usage. A fresh conversation
  could afford more recall; a long conversation with a full context window
  probably needs less.

---

## File Reference

| File | Lines | Purpose |
|------|-------|---------|
| `internal/memory/bm25.go` | 386 | BM25 engine: tokenization, indexing, search, chunking |
| `internal/memory/recall.go` | 483 | Recall pipeline: source loading, per-source search, SmartRecall, formatting |
| `internal/agent/smart_recall.go` | 195 | Judge agent: creation, prompt building, response parsing, settings bridge |
| `internal/agent/agent.go` | ~800-910 | Ephemeral provider: call site that wires everything together |
| `internal/common/settings.go` | 262-278 | MemorySettings struct with all configurable parameters |
| `internal/common/settings.go` | 363-377 | Code defaults for all memory settings |
| `internal/skills/builtin/recall-judge/SKILL.md` | 25 | Judge skill: minimal instructions for JSON-array output |
| `internal/skills/builtin/recall-judge/settings.json` | 11 | Judge settings: disables recall on the judge itself |
| `internal/memory/search_knowledge.go` | — | SearchKnowledge tool using shared BM25 infra |
| `internal/memory/search.go` | — | Search tool handler |
