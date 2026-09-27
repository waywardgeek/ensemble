# Brief: Chapter 17 Coder (Auto-Recall)

Read `book/chapter-17.md` first. It is the contract. This brief carries what
the chapter does not: architecture constraints earlier chapters enforce,
decisions already ruled, and the traps that cost real points last time.

Reference material: `book/auto-recall-design.md` is an audit of the
CodeRhapsody implementation with file:line citations. Read it for
inspiration, not as a spec. CodeRhapsody has infrastructure this codebase
does not, and copying its shape will break graded architecture here.

---

## 1. The single most important thing

**Run the full sweep before you declare done.** Not this chapter's grader.
All of them.

Chapter 16 scored 100/100 on its own grader for most of a day while silently
breaking chapters 5, 6, 7 and 8. Twenty points lost across three chapters,
plus ten in chapter 5. The cause was two architecture violations that ch16's
own checks could not see, because every one of them was behavioural: they
drove the agent through a fake vendor and read the save file. The agent
behaved perfectly. Behaviour was never what broke.

A chapter's own grader cannot see what earlier chapters protect. The
cross-chapter sweep is the only instrument that can.

---

## 2. Architecture: recall is a new spoke (RULED)

This codebase is a star. The hub is `agent/internal/common`. The spokes are
`internal/llm`, `internal/tools`, `internal/jobs`, `internal/mcp`,
`internal/ws`.

**A spoke may import the hub and the standard library. Nothing else.**

This is graded by chapter 6 (`hub-clean`, `ch5-parity`), and that grader
contains a mutation named `spoke-imports-spoke` whose expected failure set is
exactly those two checks. It has been waiting since chapter 6 was written.
Chapter 16 walked into it.

VERIFIED today: `internal/llm` imports only stdlib and `common`.

### The ruling

Recall is its own spoke: `agent/internal/recall`.

Recall needs a model (for the judge). The model layer needs recall (to attach
memories). Neither may import the other. **Both dependencies invert through
interfaces on the hub.**

```go
// agent/internal/common/recall.go

// Chunk is one searchable piece of a memory file or document.
type Chunk struct {
    Filename string
    Header   string
    Content  string
}

// Recaller is what the turn loop uses to get memories for a user message.
// Implemented by internal/recall. A nil Recaller means recall is off, and
// the agent must still work.
type Recaller interface {
    Recall(query string, convo []string) PartList
}

// SnippetJudge is what recall uses to filter candidates by relevance.
// Implemented by internal/llm, which owns the clients. A nil SnippetJudge
// means no judge: fall back to top-N by BM25 score.
type SnippetJudge interface {
    Pick(prompt string) (string, error)
}
```

| Package | May import | Implements | Consumes |
|---|---|---|---|
| `internal/recall` (new) | `common` + stdlib | `common.Recaller` | `common.SnippetJudge` |
| `internal/llm` (unchanged) | `common` + stdlib | `common.SnippetJudge` | `common.Recaller` |
| `common` (hub) | stdlib | adds the three types above | |

Wiring happens at the top, in `agentkit`/`main.go`, where everything is
already visible. That is the only place both concrete packages appear.

A callback added solely to break a cycle is a red flag. An interface on the
hub is the pattern this codebase already teaches.

---

## 3. The judge is NOT a sub-agent

This codebase has no sub-agent infrastructure. VERIFIED: the only `spawn`
references are the chapter 13 MCP fake-server harness. Sub-agents are an
unbuilt candidate in `book/chapters.md`.

The CodeRhapsody implementation makes the judge a persistent sub-agent with
history reset, a nil ephemeral provider, and a `SKILL.md`. **Do not copy
that.** All of that machinery exists to stop an *agent* from recursing.

Here the judge is one stateless call behind `common.SnippetJudge`. It takes a
prompt, returns text. It holds no history, owns no context, and has no
ephemeral provider, so **recursion is structurally impossible rather than
defended against.** Three of CodeRhapsody's "critical details" simply
evaporate. That is the correct outcome, not a shortcut.

Keep from CodeRhapsody: the prompt shape and the defensive parsing.

---

## 4. Recall is NOT ephemeral (RULED)

Do not use `Ephemera` for recall, and do not add a turn-scoped variant. Both
are wrong, and the reason is cache economics.

VERIFIED at `internal/common/context.go`: `Ephemera PartList` is cleared by
`case RequestSent: c.Ephemera = nil`. `RequestSent` fires on **every** API
request, including every tool round, so `Ephemera` is **per-request**. Its
comment gives the rationale: "a stale timestamp is not stale data, it is a
lie." That is correct for time. It is wrong for memories.

VERIFIED in CodeRhapsody (`internal/agent/project_state.go:797`): the
ephemeral provider fires only when `content != ""`, and the result is
attached to the user message object. Recall there is computed once per user
message and rides with it for the whole turn. Turn-scoped, never per-round.

Now the cost argument, which settles it. A prompt cache is a **prefix**
property. Three placements, one turn of ten rounds, a 6KB recall block:

| Placement | What the cache does | Cost per turn |
|---|---|---|
| Inline, deleted next turn | Deletes bytes from mid-prefix; invalidates **everything after** the deletion point | Catastrophic |
| Tail, turn-scoped | No mid-prefix deletion, but the block sits past the divergence point and is re-processed **every round** | ~6KB x 10 = 60KB uncached |
| Inline, kept | Prefix grows monotonically; block is cached after the first round | ~6KB once |

Rounds are many; turns are few. Anything re-sent per round is the expensive
thing. Anything added once per turn is cheap.

**So recall lands in `Dialogue` as its own entry kind, and it is PERMANENT.**
Add `KindRecall`, the way chapter 15 added `KindTools` and chapter 16 added
its band kind. Do not call it ephemera anywhere in the code or the prose:
it persists until `micro_handoff` removes it, which is what permanent means.
"Ephemera" would name the intuition rather than the behaviour.

**Never merge recall into the user's message.** In the context it is a
separate entry, never folded into the user's `Parts`. Merging would cost
three things at once:

- The user's actual words would stop being distinguishable from
  machine-retrieved text.
- `micro_handoff` could not remove recall without touching a user turn, and
  user turns are kept verbatim by definition.
- The log would lose provenance: nothing would record who said what.

**Placement on the wire is the renderer's job, not the context's.** The
context records *that* recall happened and *what bytes* it produced. How that
becomes a request is a vendor decision, and vendors differ: one wants a
mid-conversation system block, another a role-tagged message, another an
extra part on an existing turn. Baking a wire shape into the context would
repeat the mistake chapter 2 exists to prevent.

The renderer owes exactly one guarantee: **a stable position**, so the prefix
grows monotonically and the cache behaviour above holds.

Because it is a distinct kind, it is **addressable for deletion by kind**.

It is therefore subject to chapter 15's `curate()` ladder exactly like every
other entry. Stale recall is cut by machinery that already exists, under
pressure, oldest first. That satisfies the binding rule that no context field
may grow without bound, without inventing a second lifetime concept to do it.

This makes the chapter's current §17.5 ("Ephemeral, Not Persistent") wrong.
It is being rewritten. Build what this brief says.

**Dedupe.** Because blocks persist, the same memory recalled on turns 3, 7
and 12 would appear three times. Before attaching a snippet, skip it if its
content is already present in context. Cheap, and it makes keeping strictly
better than clearing.

**`micro_handoff` is the remover.** Recall blocks accumulate, so something
must eventually clear them, and the removal belongs to `micro_handoff`:

- It already strips tool calls and tool results below the watermark. A recall
  block is the same category of thing, an input to thinking rather than the
  thinking itself, and by the time a checkpoint is written whatever mattered
  about it is in the checkpoint document.
- It already rewrites the prefix, so it is **already paying for a cache
  miss**. Deleting recall there is free. Deleting it on its own schedule
  would buy a second invalidation for nothing.

That is the general rule this chapter should leave the reader with: **delete
only at a moment that is already paying for a cache miss.** Chapter 15's
ladder cuts, chapter 16's memory graduation, and `micro_handoff` are all such
moments. Recall removal rides along with one of them; it never creates its
own.

It also keeps chapter 16's single-writer rule intact: every non-conversation
entry is removed only by its own verb. Recall does not delete itself.

Do **not** rename or repurpose `Ephemera`. It is graded, shipped chapter 15
behaviour, and time still belongs in it.

---

## 5. Recall must be an EVENT carrying bytes

This is the trap most likely to cost you a chapter.

The reducer is a pure function over events. The judge is an LLM, so recall is
**nondeterministic**: the same query can produce different snippets. If recall
were computed while rendering, then replaying the log would re-run BM25 and
call a model, and two replays could disagree.

Chapter 16 already grades this with `replay-needs-no-llm`.

So recall attaches as an event that carries its bytes, exactly as
`BandPopulated` does:

```go
// Emitted after recall runs, before the request is sent.
type RecallAttached struct {
    Parts PartList // the formatted snippets, as TEXT
}
// Reducer: append to Dialogue as ordinary content (see §4).
```

Replay then reproduces context byte-for-byte with **zero** model calls and
zero BM25 work. No `Ref`, no path to dereference, no recomputation. The
reducer has nowhere to turn a reference into text, and it must not acquire
one.

This is the render-time versus event-time boundary from §16.1, appearing for
the second time. Chapter 15 had to come first precisely so this would be
expressible.

---

## 6. No mutable globals (10 points, chapter 5)

Stop-word sets and any compiled regular expressions are **fields on the
index struct, compiled in the constructor**. Not package-level vars.

Idiomatic Go would write `var stopWords = map[string]bool{...}` at package
scope. This architecture rejects it, and chapter 5 grades the rejection.
Chapter 16 lost the points here with two `regexp.MustCompile` package vars.

---

## 7. Build order

1. `common`: add `Chunk`, `Recaller`, `SnippetJudge`, `RecallAttached` + its
   reducer arm. Build. Nothing else compiles yet and
   that is fine.
2. `internal/recall`: BM25 (tokenizer, stop words, index, scorer, chunker),
   then per-source quotas, then formatting. Pure functions, easy to test,
   no model needed.
3. `internal/llm`: implement `SnippetJudge`; hold a nilable `Recaller`; emit
   `RecallAttached` before the request.
4. Wire in `agentkit`/`main.go`.
5. Grader.
6. **Full sweep.**

After step 1 and after step 3, run `go vet ./...` and confirm the spoke
imports are still clean. Do not wait until the end to discover a cycle.

---

## 8. Grading the judge without a real model

The judge is a model call, so the grader drives it with the fake vendor.
Route the judge model to a scripted responder that returns a chosen index
array.

**Counting gotcha, learned the hard way in ch16:** an `expect []int` that
counts *turn* requests will not see judge calls at all. Count router deltas
instead. This is exactly why a ch16 mutation survived its first audit.

Test the fallback by making the judge return garbage, time out, and be nil.
All three must degrade to top-N BM25, never to a crash and never to an empty
turn.

---

## 9. Grader checks

The chapter's TL;DR lists ten checks. One of them is now wrong and must be
replaced.

`judge-no-recursion` (10 pts) is **dead**. With a stateless judge, recursion
cannot happen, so the check cannot fail and cannot teach. A check that no
mutant can kill is decoration. Chapter policy P9 forbids it.

Replace it with:

| Check | Points | What it proves |
|---|---|---|
| `recall-is-a-spoke` | 10 | `internal/recall` imports only `common` + stdlib |

This is worth grading in-chapter despite chapter 6 covering star topology
globally, because chapter 17 is the chapter that *adds* a spoke, and adding a
spoke wrongly is the precise mistake that cost twenty points in chapter 16.
Update the TL;DR table when you make the change.

Two more the chapter promises and you must therefore grade honestly:
`injection-attached` must prove the recalled text reaches the model in the
request; `injection-capped` must prove the 6KB ceiling holds when snippets
would exceed it. Rename `injection-ephemeral` to `injection-attached` in the
TL;DR table: with recall landing in `Dialogue`, "ephemeral" is no longer what
the check proves.

---

## 10. Mutation audit (P9)

Every check needs a mutant that kills exactly it. Write the mutant by
**deleting one behaviour from the reference**, not by breaking syntax.

Known-equivalent traps to avoid wasting time on: if a fold or a cut always
consumes the whole collection, then "oldest" and "newest" name the same
slice and no mutant can distinguish them. Document an equivalent mutant
rather than tuning parameters for hours trying to kill it.

If a mutant's scenario fails *every* path, nothing downstream ever runs and
the mutant proves nothing. Check that the scenario still produces output
before concluding the check is strong.

---

## 11. Do not

- Do not add a `Ref` that the emitter or reducer must dereference. A stub
  that *tells the model* where bytes live is fine; a stub the code must
  resolve is not.
- Do not rename `Ephemera`.
- Do not make `internal/llm` import `internal/recall`.
- Do not give the judge memory, history, or an ephemeral provider.
- Do not run two graders at once. Chapter 16 scored 11/100 inside a
  concurrent sweep and 100/100 alone. Graders build binaries, bind ports,
  and race for CPU. **Any number measured alongside other work is void.**
- Do not `git add -A`. Stage exact paths.

---

## 12. Defaults

From the chapter's TL;DR, which are production-tuned values, not guesses:

| Setting | Default |
|---|---|
| `MinScore` (no judge) | 3.0 |
| BM25 threshold (with judge) | 0.5 |
| `MaxSnippets` | 3 |
| `MinQueryLength` | 80 |
| `Candidates` | 20 |
| `ContextMsgs` | 3 |
| `MaxRecallBytes` | 6144 |
| Judge timeout | 15s |
| BM25 `k1` / `b` | 1.2 / 0.75 |

Memory files get 50% of candidate slots. Unused quota flows to sources with
overflow.

---

## 13. One known bug worth inheriting

CodeRhapsody's BM25 carried an off-by-one error for weeks. Recall quality
improved noticeably when it was fixed, and nothing failed loudly while it was
broken, because a ranking function that is slightly wrong still returns
plausible results.

Test scoring against hand-computed values on a tiny fixture. Do not test it
by reading the output and deciding it looks reasonable.
