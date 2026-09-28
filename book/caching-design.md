# Caching: design notes

Written 2026-09-27, during the CodeRhapsody → Ensemble crossover. Bill named
caching as the single biggest thing standing between Ensemble and daily use.

---

## 1. Where we actually are

`grep -rn 'cache_control' internal/ --include=*.go` returns nothing. Every
"ephemeral" hit in the tree is about ephemeral *tools*, an unrelated concept.

So: **no breakpoints, no measurement, no stats.** Every request is a cold read
of the entire prefix — SOUL.md, MEMORY.md, the memory bands, the skill text, the
tool declarations, and the whole dialogue — at full input price, on every turn.

On Anthropic, a cache read costs a tenth of an input token. A stable prefix that
hits reliably is close to an order of magnitude off the dominant line item. That
is the size of the prize, and it is why this blocks the crossover rather than
merely annoying us.

---

## 2. The thesis: this failure is silent

Most defects in an agent announce themselves. A bad tool call errors. A broken
parse throws. A wedged job hangs where you can see it.

A cache miss does none of that. You get the right answer, slightly later, for
about ten times the money. Nothing is logged, nothing is red, no test fails. The
only channel that reports it is the invoice, thirty days later, aggregated across
every session so you cannot tell which change caused it.

**A failure mode with no natural signal needs a manufactured one.** That is the
argument for building an auditor rather than just placing the markers and
trusting them. Markers are three lines of JSON; knowing whether they *worked* is
the actual engineering.

---

## 3. Measure before optimizing — the ordering is the lesson

The tempting first move is to add `cache_control` breakpoints and watch the bill
drop. That is the wrong order, because a breakpoint on an unstable prefix buys
nothing at all and looks like it should have worked.

A cache hit requires the prefix to be **byte-identical** to the previous request.
Any of these silently destroys it, and every one is plausible in this codebase
today:

- tool declarations emitted in map-iteration order
- a memory band re-rendered with a different timestamp
- auto-recall attaching a different snippet set (chapter 17 puts recall *in* the
  prefix, permanently, by design)
- a skill loading mid-session and inserting text above the dialogue
- any `time.Now()` that reaches a rendered string

So phase one measures prefix stability with **zero** breakpoints. If the prefix
does not survive from one request to the next, that is the bug, and no amount of
cache configuration addresses it. Only once the prefix is provably stable do
markers mean anything.

This is a better chapter narrative than "here is how to save money": *you cannot
cache what you have not proved is stable.*

---

## 4. What to capture, and in what order

Write two files per request, canonically serialized:

```
cache/request.json         the request just sent
cache/prior_request.json   the one before it
```

**Serialized in the order the provider concatenates them**, which for Anthropic
is: `system`, then `tools`, then `messages`. Not struct-declaration order, not
alphabetical — wire order. The entire value of these files is that
`diff prior_request.json request.json` puts the cursor on the first byte that
broke the cache. If the file order does not match the wire order, the first diff
hunk is not the first divergence and the tool lies.

Two properties this needs:

- **Deterministic key order.** Go marshals struct fields in declaration order
  and sorts map keys, so we are mostly safe by construction — but this must be
  asserted by a test, not assumed, because a single `map[string]any` in a
  rendered payload reintroduces nondeterminism invisibly.
- **Deterministic tool order.** CodeRhapsody sorts tools by
  `(skill_name, tool_name)` specifically for prompt-cache stability. Ensemble
  has no such guarantee today. It needs one, and it needs a test.

---

## 5. The analysis, and the distinction that makes it trustworthy

For each request, compute:

| quantity | source |
|---|---|
| common prefix length (bytes) | byte compare against `prior_request.json` |
| divergence location | section + index + byte offset, e.g. `messages[4].content[0]` |
| predicted cache read | prefix length up to the last breakpoint that still matches |
| actual cache read | `usage.cache_read_input_tokens` from the response |

Then classify. **This is the part that determines whether the alarm gets
believed or ignored**, and it is why a single "cache hit rate" number is not
enough:

- **MATCH** — predicted and actual agree. Nothing to say.
- **PREFIX DIVERGED** — the common prefix ended before our breakpoint. *Our
  bug.* The diff localizes it. Actionable, and this is the case worth shouting
  about.
- **CACHE MISSED ANYWAY** — the prefix was identical through the breakpoint and
  the provider still charged us full price. *Not our bug.* Causes: the 5-minute
  TTL expired during an idle gap; the prefix was below the minimum cacheable
  size (1024 tokens for Opus and Sonnet); provider-side eviction.

Collapsing these two into one alert is how monitoring gets switched off. The
second case is normal and frequent — think, or read, for six minutes and every
subsequent first request legitimately misses. If that fires the same alarm as a
real nondeterminism bug, the alarm becomes noise inside a day.

---

## 6. The units trap

Prefix length is **bytes**. The provider reports **tokens**. There is no exact
conversion without the vendor's tokenizer, and shipping one is not worth it.

Two honest options:

1. **Self-calibrate.** Derive bytes-per-token from the previous response:
   request bytes over reported input tokens. Accurate enough to distinguish "we
   lost the cache" from "we kept it", which is the only question being asked.
2. **Refuse to convert.** Report divergence as bytes and a section path, report
   the provider's token counts as their own claim, and flag on the *ratio*
   rather than the absolute.

Prefer reporting both and alerting on the ratio. "Did we lose the cache" is
nearly binary; it does not need token-level precision, and pretending to that
precision would be the kind of number that later gets quoted as if measured.

---

## 7. The structural blocker, found while writing this

```go
System string `json:"system,omitempty"` // TOP-LEVEL, not a message
```

Anthropic's `cache_control` attaches to a **content block**, so caching the
system prompt requires `system` to become an array:

```json
"system": [{"type": "text", "text": "…", "cache_control": {"type": "ephemeral"}}]
```

That is a structural change to `anthRequest`, and it collides with a property the
surrounding comments defend explicitly and repeatedly: *"omitted when false so
that a non-streaming request is byte-identical to one written before streaming
existed."* Every feature so far has been added without disturbing the bytes of a
request that does not use it. Caching cannot be.

This is worth a section in the chapter rather than a quiet edit. The
compatibility discipline was correct, it was maintained for many chapters, and
here is the feature that ends it — with the mitigation being that the array form
is used only when caching is on.

---

## 8. Session stats

What Bill asked to see, scoped to one Ensemble restart:

- session cost
- total input tokens
- cache read tokens
- **cache hit rate** — `cache_read / (input + cache_read)`

The four token categories already exist and are already disjoint
(`Input`, `CacheWrite`, `CacheRead`, `Output`), which is exactly the shape this
needs. Two gaps:

1. Usage is printed to stdout and **never sent over the WebSocket**. The GUI has
   no usage handling at all.
2. There is no price table, deliberately: `event.go` states *"Usage counts tokens
   and never money. Prices change; counts are history."* That rule is right and
   should stand. Cost is therefore computed at **render** time from a price table
   living beside the model catalog, never stored in the log.

---

## 9. On making this a chapter

Bill suggested caching might combine with the crossover notes. I would keep them
separate.

Caching has a thesis of its own — *silent failures need manufactured signals* —
and it has the best exercise in the book available to it: plant a
nondeterministic serialization in the prefix and require the student's auditor to
localize it. That is gradeable without naming any of our identifiers, the
property the ch14 revision taught us to insist on.

The crossover log is a different genre. It is a migration diary, and its value is
that it is specific and dated. Folding it into a technical chapter would dilute
both. It belongs as an appendix, or as the seed of a "living with the thing you
built" chapter — which is where the sub-agent and gateway work will also land.

---

## 10. Build order

1. Canonical request capture + `prior_request.json`, in wire order. No markers.
2. The differ: prefix length, section path, first differing byte.
3. Prove the prefix is stable across two identical-input turns. **Expect this to
   fail the first time** and to name something real.
4. Deterministic tool ordering, with a test.
5. Breakpoints — system/tools boundary, and a sliding one near the dialogue tail.
6. Predicted-vs-actual classification with the three verdicts.
7. Usage over the WebSocket; session stats in the header; price table at render.
