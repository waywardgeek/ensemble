# Chapter 7 code review — streaming

Coder's record for the author. Everything here is measured at `fd92f1a`
(tag `ch07-solution`), eight commits ahead of `origin/main`. Push is yours.

The gate: eight grader invocations pass, `go vet ./...` is clean, `gofmt`
reports only two pre-existing offenders I did not touch, and
`go test -race -count=2 ./...` is clean in `agent/`.

| grader | target | score |
|---|---|---|
| ch1 | `solutions/ch01` | 100/100 |
| ch2 | `solutions/ch02` | 100/100 |
| ch3 | `solutions/ch03` | 100/100 |
| ch4 | `solutions/ch04` | 100/100 |
| ch5 | `.` | 120/120 |
| ch6 | `.` | 100/100 |
| ch7 | `.` | 100/100 |
| ch7 | `solutions/ch07` | 100/100 |

---

## 1. What changed

**The seam.** `Parse` was re-signatured rather than joined by a sibling:

```go
Parse(resp *http.Response, cb StreamCallbacks) error
```

One method, both modes. `StreamCallbacks` and `DeltaKind` live in
`internal/common` beside `Parser`, per Chapter 5's rule that interfaces live in
the hub. The seam still has exactly two methods, which is what `seam.go`'s own
package comment stakes the chapter on.

**The reader.** `internal/llm/sse.go`, 111 lines, with 14 tests covering the
framing cases that actually bite: multi-line `data:` fields, comment lines,
CRLF, `[DONE]`, and an unterminated final event.

**Capability.** A `Stream` bitmask on `ModelFeatures`
(`StreamText | StreamThinking | StreamToolArgs`) plus a `DisableStreaming`
field on config. Effective behaviour is the AND of the two. Gemini's row has
`StreamToolArgs` clear, because its tool arguments still do not arrive
incrementally; that is expressed as a missing capability bit rather than a
special case in the parser.

**Three vendors.** Anthropic dispatches on `content_block_delta` subtypes.
OpenAI needed `stream_options.include_usage`, without which a streamed
response omits `usage` entirely and the token accounting Chapter 2 worked for
silently reads zero. Gemini is the alien one: every SSE frame is a whole
response object with no block indices, so part continuation is inferred.

**Logging.** `send` logs each SSE frame to `api.log` as it arrives, so the
wire trace stays byte-complete in both modes. The parser is a stateless empty
struct with no `Host`, so the frame hook is carried in `StreamCallbacks` and
wired to `APILogf` by the engine: logging remains a parent-provided facility
rather than something the seam owns. After assembly the event is still
recorded to the event log and `PartFinal` still fires, which is what lets a
GUI stream chunks into a widget and then re-render that widget from the
authoritative event.

**The actor.** `notifyContent` is gone. The parser's callbacks supersede it,
and leaving both would have meant two code paths minting ids for the same
content. `AskWatching` is a new method rather than a parameter on `Ask`,
because `Ask` is what every earlier chapter and both earlier exercises call,
and a new parameter would have made this chapter reach backwards into code
with nothing to do with streaming.

**Chat mode** streams reply text to stdout, with reasoning dimmed and tool
calls yellow on stderr, so piping still yields exactly what the assistant
said and nothing else. Colour is chosen by `isTerminal(stderr)`, not a flag.

**The exercise.** `ch07/main.go`, 325 lines: one agent, one observer, and a
report about the shape of the stream. It measures time to first delta, which
is the number streaming exists to move.

---

## 2. Check → mechanism

| check | points | what would have to be true to fake it |
|---|---|---|
| `stream-deltas` | 25 | Emit at least 8 text deltas across a turn. Non-streaming produces 2, so this cannot pass without asking for the stream. |
| `deltas-match-final` | 20 | For every text final, a delta group with the **same** `PartID` must concatenate to it exactly, and at least one part must have arrived in more than one delta. |
| `thinking-streamed` | 15 | At least 2 thinking deltas under `DeltaThinking`, with no reasoning text leaking into the reply. |
| `tool-params-streamed` | 15 | At least 3 `DeltaToolCall` deltas reassembling to the call's name and arguments. |
| `ch6-parity` | 25 | Every Chapter 6 check still passes against the same tree. |

`deltas-match-final` is the one that carries the chapter. It groups **by
part**. A grader that concatenated every text delta in a turn and compared the
result to the last assistant message would pass while the ids were pure noise,
and would keep passing under a submission that minted a fresh id per chunk,
which is exactly the bug this chapter removes.

---

## 3. The audit, in full

Five mutants. Each deletes one behaviour from the reference and must fail
exactly the checks that should notice.

| mutant | score | fails |
|---|---|---|
| `no-stream-flag` | 25/100 | `deltas-match-final`, `stream-deltas`, `thinking-streamed`, `tool-params-streamed` |
| `per-chunk-part-ids` | 80/100 | `deltas-match-final` |
| `no-thinking-deltas` | 85/100 | `thinking-streamed` |
| `no-tool-arg-deltas` | 85/100 | `tool-params-streamed` |
| `no-model-validation` | 75/100 | `ch6-parity` |

`no-stream-flag` is the chapter's thesis as a receipt. Remove `stream: true`
and the final text is identical, the parts are identical, and the event log is
identical; the grader reports *"saw 2 text deltas, want at least 8"*. Nothing
about the content changed. Only the shape of its arrival did.

Two of my five predictions were wrong, and both corrections are worth having
in the chapter:

**`per-chunk-part-ids` first used a package-level counter.** That version
failed two checks, not one: `ch6-parity` broke because it cascaded into
Chapter 5's `no-mutable-globals`. A better story but a worse test, since only
one of the two failures was about streaming. The mutation now derives the id
from the chunk, needs no global, and is isolated to `deltas-match-final`.

I found that only because the audit logs each failing check's `Details`, not
just its id. *Which* checks failed tells you a mutant is wrong; *why* tells you
whether it is wrong for the reason you intended. That logging is now permanent.

**`no-thinking-deltas` does not cascade into `deltas-match-final`**, because
that check walks the text parts and is indifferent to the reasoning stream. I
predicted a cascade and there was none. Recorded with the reason, so the
division of labour between the two checks is deliberate rather than apparent.

Every pattern must match exactly once or the test fails hard. A pattern
matching zero times yields a mutant identical to the reference, which then
scores 100 and reports a *passing* audit for a behaviour nobody deleted. That
has already happened once on this book.

I also re-ran Chapter 6's audit after the refactor, because removing
`notifyContent` touched actor code that three ch6 mutants target. All seven
still land; the ch6 audit is intact.

---

## 4. Where the brief and the code disagreed

Four of these you ruled on directly; the last two were mine.

**`PartID` per part, not per delta.** The brief's actor snippet minted a fresh
id per chunk via `atomic.AddUint64`. That would have made correlation
impossible in principle, and a GUI could never append chunks into one widget.
The parser now owns **both** ends: it supplies the id on every delta and
reports the matching `PartFinal`. This turned out to be necessary rather than
merely tidy, because OpenAI streams a tool call's arguments before it is
knowable whether a text part will occupy index 0 — the caller cannot compute
the id, so the caller must not be the one choosing it.

**`StreamCallbacks` in `common`.** The brief put it in `llm`; `Parser` is
declared in `common`, so a method on it cannot take a type from a spoke.

**One `Parse`, not a second method.** You asked whether deleting `Parse` would
cost non-streaming support. It does not: non-streaming is a stream of length
one, which is already the rule written into `observer.go`. Two methods would
have meant six vendor implementations and two chances per vendor to disagree
about what a tool call means.

**Per-frame `api.log`.** Cleaner than the tee I proposed, and it means the log
shows the same JSON blobs whether or not we streamed.

**`DisableStreaming`, a negative boolean.** You asked for a field defaulting
to streaming on. Go's zero value for `bool` is `false`, so `Stream bool` would
default a hand-built `Config{}` to streaming *off* — the opposite of the
default you asked for, recoverable only inside a constructor, in a framework
where other people construct configs. The negative name reads worse and is
right. Flag it if you would rather pay the other cost.

**`StreamingFor` never fails on an unknown model.** I first made it return an
error, on the "no default row" principle. It broke a test using model `gpt-5`,
and the model set is provably open: the tests and graders already use
`gpt-5-2025-08-07`, `fake-model`, and several `*-fake` variants. Hard-failing
every request on an unfamiliar model name would brick the framework for any
model released after this chapter. So an unknown model streams nothing rather
than erroring, while `Media` stays a loud refusal. The asymmetry is the point,
and I would like it stated in the chapter: guessing about **content** corrupts
a request, guessing about **delivery** only changes the chunk count.

---

## 5. Open questions for the author

**1. A sixth check for the equivalence claim?** The brief's table is five
checks summing to 100, and I implemented it as specified. But the chapter's
central claim — that streaming changes delivery and not content — is currently
evidence in a `notef` rather than a graded property:

> streaming disabled produced 2 text deltas (length-one stream), streaming produced 36

The exercise already runs both ways under `CH07_NO_STREAM=1`, so a check
asserting *same final text, same part count, different delta count* would cost
almost nothing. I did not re-weight your table on my own judgement. Say the
word and it is a Go unit test or a sixth check.

**2. `deltas-match-final` validates text correlation only.** A submission
could correlate text ids perfectly and scatter ids across thinking and tool
parts unnoticed, since `thinking-streamed` and `tool-params-streamed` check
that those kinds *arrive*, not that their ids correlate. Extending the check
is easy but means re-auditing; it is a real if minor gap and I would rather
you chose.

**3. Gemini's `StreamToolArgs` bit is clear.** That matches what you reported:
streamed function parameters still do not work there. Worth a sentence in the
chapter, because it is the best example in the book of a capability table
earning its keep — the parser has no Gemini special case, the row simply
says the bit is off and tool arguments arrive as a length-one stream.

**4. 18MB of tracked binaries, not mine.** `solutions/ch05/bin` and
`solutions/ch05/agent/bin` are 9MB each, committed before this chapter. No
grader needs them; the graders build their own. They sit in a frozen snapshot,
so removing them is your call, but `e083e02` suggests it is what you wanted.

**5. The snapshots disagree with each other.** `solutions/ch06` contains the
agent module flattened with no `ch06/` at all, so `grade -ch 6 solutions/ch06`
cannot score 100 and only `grade -ch 6 .` can. `solutions/ch05` has both the
correct shape *and* a flattened duplicate. That inconsistency is why the
Makefile points `grade5`/`grade6`/`grade7` at the repo root while
`grade`/`grade2`/`grade3` point into `solutions/`. I shaped `solutions/ch07`
correctly and left the older two alone rather than reshaping frozen artifacts
as a side effect of this chapter.

**6. One parity gap in the fake vendor.** `Reply.Thinking` is rendered only by
the SSE path, so with `DisableStreaming` set, a scripted reply's reasoning
block does not appear. It affects the fake only, never a real vendor, and no
check depends on it. Small and worth closing if you want the two paths exactly
symmetric.

---

## 6. Facts, with provenance

Measured at `fd92f1a`. Re-derive rather than trust these if the tree moves.

| figure | value | how |
|---|---|---|
| commits ahead of `origin/main` | 8 | `git log --oneline origin/main..HEAD` |
| `agent/` change | 19 files, +1744 / −241 | `git diff --stat origin/main..HEAD -- agent/` |
| text deltas, streaming on | 36 across 2 parts | grade7 `stream-deltas` |
| text deltas, streaming off | 2 | grade7, `CH07_NO_STREAM=1` |
| thinking deltas | 27, none leaking into reply text | grade7 `thinking-streamed` |
| tool_call deltas | 19 | grade7 `tool-params-streamed` |
| SSE reader | 111 lines, 14 tests | `wc -l agent/internal/llm/sse.go` |
| new code this chapter | 1,698 lines across 7 new files | `wc -l` on the added files |

New files: `agent/internal/common/delta.go` (148),
`agent/internal/llm/sse.go` (111), `agent/internal/llm/sse_test.go`,
`internal/fakevendor/stream.go` (279), `internal/grade/ch07_checks.go` (282),
`internal/grade/ch07_harness.go` (266),
`internal/grade/ch07_grader_test.go` (287), `ch07/main.go` (325).

Two operational notes for whoever runs this next. The graders **build
binaries into the submission tree** on every run, so deleting artifacts by
hand does not keep them gone; the `.gitignore` rules are now depth-independent
because a repo-root-shaped snapshot nests a second copy of the whole layout
and `solutions/ch07/ch05/bin` is a real path. And `internal/grade`'s audit
subtests use `t.Parallel()`, so the parent test reports `0.00s` and a
`^`-anchored grep will hide every subtest result — which briefly convinced me
the ch6 audit had gone dead when it was perfectly healthy.
