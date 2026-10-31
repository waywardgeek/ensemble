# Chapter 15: The Invisible Invoice

Bill reported $2,475 in model spend on the first edition through its recall
chapter. He attributed a few hundred dollars to a caching bug that produced
correct answers, on time, while quietly losing the discount on repeated input.
The answers gave him no reason to look for it.

Every request the agent sends carries a prefix: tool declarations, the system
prompt, then retained memory, skills and conversation. Much of that prefix is
identical from one turn to the next. A provider cache can make those repeated
tokens cheaper to read. Lose the cache and the same useful answer arrives with
a different bill.

The bug left no obvious trace in the agent's behavior. No test failed. No log
line turned red. An aggregated invoice is a poor place to discover which turn
first lost its prefix. Correct output can conceal a cost defect indefinitely.

This chapter builds the instrument that detects prefix changes, the markers
that identify reusable boundaries, and the meter that reports the cost while
the operator can still do something about it.

---

## TL;DR

### Data structures

Prices live on `ModelFeatures`; usage records counts, never money. An unknown
price displays a dash. It must not turn an unpriced paid model into a free one.
These shared types use disjoint token categories:

```go
// Pricing holds full USD rates per million tokens, not surcharges.
type Pricing struct {
    Input      float64
    CacheWrite float64
    CacheRead  float64
    Output     float64
}

// Usage retains measured counts independently of today's prices.
type Usage struct {
    Input      int // uncached input only
    CacheWrite int
    CacheRead  int
    Output     int
}
```

Engine owns the process-lifetime accumulator and retains the latest
main-response counts separately. Neither is restored from `save.json`.
Keep each request's model and applicable price tier with its counts.
Chapters 22 and 23 later add compressor and judge usage to these totals,
priced by the model that incurred it.

The cache lens belongs to Engine. Its common interface exposes that required
parent and two observations:

```go
// CacheLens observes main requests and compares their predictions with usage.
type CacheLens interface {
    Engine() Engine
    ObserveRequest(name string, body []byte) float64
    ObserveUsage(usage Usage, predicted float64)
}
```

The returned prediction belongs to that request's eventual response. An
interrupted worker can finish after a new turn begins. Serialize comparison and
file rotation, but do not hold a lens lock across HTTP. Observe actual main
sends rather than every `Render` call. Chapters 22 and 23 later introduce
compressor and judge conversations; those must stay out of this comparison
while their costs still count.

### Breakpoints

A `cache_control` marker on a content block tells the provider: everything
before this is a cacheable prefix. Anthropic allows four markers per
request, and OpenAI allows four cache writes under nearly the same scheme.
Use these four positions when the corresponding content exists, in cache order:

1. **Tool array.** One marker on the last tool declaration. Tools sit at
   the front of the cache order, ahead of the system prompt. This is the
   only marker that survives a change to the system prompt.

2. **System prompt.** One marker at the end of the system block. Stable
   by construction.

3. **The compaction bound.** One marker at the newest surviving
   `micro_handoff`. Chapter 22 later adds the last surviving memory entry
   as the fallback when compression removes that handoff.
   Mark the projection that actually remains;
   never resurrect old dialogue to manufacture a boundary.

4. **Stable end of history.** One rolling marker at the last eligible
   retained block before ephemera. It advances on each turn. Moving
   a marker does not invalidate what it previously banked.

That is up to four markers for four jobs, the Anthropic budget.
OpenAI can write three of these: its breakpoints attach only to message content
blocks, so the tool-array marker is not expressible there. The system
marker already covers tools, but cannot keep them independently when the
prompt changes. There is no room for a fifth idea.


Two constraints:

- **Exclude ephemera.** Ephemera are appended last in Render and cleared
  after delivery. A marker placed after them caches bytes guaranteed to
  disappear. That boundary cannot match after the content is removed,
  although an earlier prefix can still hit. Capture the stable boundary
  before the ephemera append, even when retained and ephemeral parts later
  share a wire message.
- **Strip markers before comparing.** The stable-end marker moves
  between turns, so a block carrying `cache_control` in one request
  carries none in the next. Compare raw bytes and every healthy turn is
  classified as edited. The comparison strips markers first: a breakpoint
  is an instruction to the provider, not conversation content.

### Wire format

Anthropic: explicit `cache_control` markers on content blocks. System
must be an array of content blocks rather than a string. Up to four
markers per request.

OpenAI: on this chapter's supported GPT-5.6 Sol Chat Completions route,
`prompt_cache_breakpoint` markers on content blocks, enabled by a request-level
`prompt_cache_options: {mode: "explicit"}`. Up to four cache writes per request.
Select explicit support from the model table; older rows retain implicit caching. Breakpoints
attach only to message content blocks (text, image, audio, file), not to
the tool array. The agent therefore writes three markers rather than four;
the system marker already covers tools, but cannot keep them independently
when the prompt changes.

The footgun: explicit mode disables implicit caching. A request that
declares explicit mode but places no markers caches nothing at all,
which is strictly worse than staying silent. On Anthropic, a missing
marker costs a cache the agent never had. On OpenAI, it destroys the
implicit cache the agent already had. The opt-in must follow evidence
that a marker landed, never intent.

Gemini: implicit caching, with no markers to place. The selected model's
minimum prefix size matters; Gemini 3.8 Flash documents 4,096 tokens. A small
prompt reporting zero reads has not established a caching defect. Warm-up and
the amount granted are observations to measure with the actual growing
conversation, not fixed numbers of cold turns to enforce.

The provider contracts above were checked on October 11, 2026 against the
[Claude caching guide](https://platform.claude.com/docs/en/build-with-claude/prompt-caching),
[OpenAI caching guide](https://developers.openai.com/api/docs/guides/prompt-caching)
and [Gemini caching guide](https://ai.google.dev/gemini-api/docs/caching).
A documented field and a well-formed request still need a real grant.

### The meter

The GUI header displays four numbers: session cost, last-response input
tokens, last-response cache read tokens, and session cache hit rate. The
input and cache-read figures show the most recent response so the
operator can see how large the current prompt is and how much of it was
cached. Session totals go to the tooltip. Cost is computed from Usage
and Price at render time and held only in memory. An unpriced model shows
token counts and a dash for cost.

The meter reads Engine's current-process accounting rather than the totals
restored with conversation history. Latest input means all input to the last
main response, including its cached buckets exactly once. Session cost and
hit rate include auxiliary calls when Chapters 22 and 23 add them.
Session cost means cost since the developer sat down.

### Rules

1. Build the lens before the markers. A breakpoint on an unstable prefix
   buys nothing and looks like it worked.
2. Prices live on ModelFeatures. Usage lives in the event log. Cost is
   derived at render time from both.
3. Token counts only in the event log. Prices change; counts are history.
4. Recognize the conversation section, including Gemini's `contents`, as
   the section expected to grow. Report unexpected stable-section changes;
   a deliberate checkpoint or model change needs its own interpretation.
5. A growing JSON array is an append rather than an edit. Without this
   distinction, every turn fires a false alarm.
6. The meter prices the model in force rather than the model named in
   settings. The two can differ when the model comes from the environment.
7. Compare predicted and granted cache rates as ratios. A prediction is
   bytes from the request; a grant is tokens from the response. Adding
   them is a type error.

### Yours

How the lens reports its verdict: log file, structured events, or both.
Whether to display the per-section stability status in the GUI. How much
detail the cache miss alarm carries. The format of the rotated request
files, as long as plain `diff` puts the cursor on the first broken byte.

### Exercise

Extend `solutions/ch14`. Wire a `CacheLens` to the main send boundary,
place supported breakpoints, build the meter, and verify stable prefixes and
actual cache reads. The grader runs from the repository root, defaulting to
`solutions/ch15` and its separate GUI:

```
go run ./cmd/grade-edition2 -ch 15 -json
```

### Grader checks

```
cache-lens-exists          (15 pts)  a cache lens observes consecutive
                                     requests; the actual requests are inspectable
prefix-is-stable           (20 pts)  tools and system are byte-identical
                                     across turns
request-has-breakpoints    (15 pts)  last tool, system and stable history
                                     blocks carry legal cache_control;
                                     history advances within four markers
cost-is-computed-not-stored(15 pts)  event log holds counts only; cost is
                                     derived at display time
usage-is-session-scoped    (10 pts)  the meter's tokens come from the last
                                     response, not the lifetime restored from disk
cache-read-rate            (10 pts)  cache reads are zero cold, nonzero warm
all-vendors-report-usage   (15 pts)  vendor parsers normalize the supplied
                                     usage categories without double counting
```

---

## §15.1 The idea in plain words

A cache miss can produce a completely correct answer. The tool call works. The edit lands. The test passes. Nothing in
the agent's behavior changes. The failure is pure cost, and it is silent.

Most bugs announce themselves. A bad tool call returns an error. A broken
parse throws an exception. A wedged job hangs where the operator can see
it. A cache miss does none of these things. The correct output arrives,
possibly later, at the uncached input rate. No log line turns red.
No test fails. The only reporting channel is the invoice, and it arrives
thirty days late, aggregated across every session, past the point where
anyone can attribute a cost spike to the change that caused it.

This chapter builds three things, in an order that matters.

**The lens** compares consecutive requests, section by section, and
reports whether the prefix is stable. A breakpoint placed on an unstable
prefix buys nothing while looking like it worked, because the agent still
produces correct output. The lens catches the nothing.

**The markers** tell the provider where the cacheable prefix ends. Place
them wrong and every turn is a cold read that looks like it should have
been warm. Place them on something that changes and the cache is
invalidated anyway. The lens locates stable bytes; the response usage says whether the provider
actually served them from cache.

**The meter** displays cost in real time. An invoice is not a feedback
loop. A number that updates on every turn is. The meter does not prevent
waste, but it makes waste visible at the moment it happens, to the person
who can fix it.

The economic argument is straightforward. At the October 11, 2026
[Sonnet-5 rates](https://platform.claude.com/docs/en/models/sonnet-5/overview),
a million ordinary input tokens cost $2, a five-minute cache write costs $2.50,
and a cache read costs $0.20. Output has its own $10 rate. Those are alternative
input buckets: do not charge the full input rate again on a cache write.
Two hundred repeated turns make a small per-turn difference consequential.
The saving depends on which tokens actually receive the discount; it is not
a tenfold reduction in every request's total cost.

## §15.2 The lens

The lens works by rotating two files: the current request and the prior
one. In the supplied clients they are `cr/cache-current.txt` and
`cr/cache-prior.txt`, under the owning Agent's directory. Each is split into named sections (tools, system, messages) and
re-emitted in cache concatenation order: the order the provider lays them
out in context. For Anthropic, that is tools, then system, then messages,
regardless of the struct's JSON field order. The ordering matters because
a prefix is a contiguous run of bytes from the start, and two sections in
the wrong order produce the wrong prefix boundary.

Each section gets one of three verdicts: identical, appended, or changed.
"Appended" means a JSON array grew at the tail, as the messages array does
on every turn. Without this distinction, every turn reports a changed
prefix, because the closing bracket of the array has moved. A changed
non-messages section is the alarm. An appended messages section is the
normal case.

The lens writes its sections to disk so that plain `diff` puts the cursor
on the first byte that broke the cache. Reading the raw file is the fast
diagnostic. Structured verdicts are the automation.

### Stripping the markers

Because the stable-end marker moves, a content block that carried
`cache_control` in one request carries none in the next. That difference
sits deep inside the prefix, where the trailing-append rule cannot reach.
Comparing raw bytes classifies every healthy turn as edited, and the
reported cacheable prefix is truncated at the stale marker position.

The comparison strips markers before diffing. The principle: a breakpoint
is an instruction to the provider rather than conversation content.
Removing it from the comparison is the same kind of move as ignoring
whitespace in a semantic diff.

The wrong way to strip is a JSON round trip. Unmarshalling and
re-marshalling reorders keys and rewrites byte counts, and those counts
are reported as a property of the real request and compared against the
provider's own figures. Both sides of the comparison get the same
distortion, so equality stays honest while the measurement becomes
fiction. The lens strips by pattern, in place, preserving every other
byte.

### Match the shape, not a rendering of it

The first-edition coder recorded a strip built from a constant holding the exact
marker bytes the renderer emits, with a comment justifying it: "we
control the emitter." The emitter is controlled. But the lens compares
what the canonicalizer produced, and the canonicalizer marshals with
`json.MarshalIndent`. The marker arrives with spaces the constant does
not have.

Nothing matched the indented marker. The strip left it in place. Nothing crashed. The symptom was a diagnostic reporting a
prefix fault that was not there, on a path that produces no error when it
fails to strip, because a strip that misses everything simply leaves the
markers in place and the comparison fires.

The reason it survived its tests: every test hand-built sections as
compact JSON, matched the constant, and passed. Test and code shared one
wrong belief about the format, and the format itself was never consulted.
The replacement test builds its input the way the agent does, with
`MarshalIndent`, and it kills the defect. The old compact test still
passes against the broken code, which measures exactly how much it was
checking.

The coder's retained review reports this two-turn comparison:

```
before:  messages:edited     66.6%,  0 breakpoints
after:   messages:appended   73.3%,  3 breakpoints
```

The lesson is general. When the instrument and the code share a belief,
the instrument certifies the code against the belief, not against the
artifact. Testing against the artifact, built the way the artifact is
built, exposes the gap.

### The late reply belongs to its own request

There is another way for the instrument to lie while every answer remains
correct. The new implementation initially assumed the actor serialized lens
observations. It does serialize turn decisions. An interrupted HTTP worker,
however, can still be finishing when the next turn begins.

Independent review held an old cold response, interrupted its turn, and let a
new request complete. The lens predicted 0.0000 for the cold request and 0.8554
for the new one. When the old response was released, it reported the new
request's 0.8554 prediction. Locking a shared “latest prediction” would prevent
a race while preserving the wrong attribution.

Keep the prediction with the request that produced it. The corrected lens
returns a ratio to Engine, and that exchange's response callback carries it
back with the granted usage. A short lock protects comparison and artifact
rotation; no lock spans the network wait. The same held-response probe then
reports 0.0000 for the old response and 0.8554 for the new one, with both actual
usages counted. Interrupting the conversation does not make an already incurred
charge disappear.

## §15.3 Where the markers go

The system marker is straightforward. The system prompt does not change
between turns, so a marker at the end of the system block banks the
startup instructions. Dynamically loaded skill bodies remain conversation
data; moving them into the system prompt would undo Chapter 9. Chapter 22
applies the same rule to mutable memory. One marker, one stable prefix.

The message-history marker is harder, because the conversation grows.
Everything before the current exchange is identical to the previous
request, and everything after it is new. A marker that banks the identical
part saves money on every warm turn.

### Why not the first message

An obvious placement is the first user message. Everything before it is
stable, so a marker there should cache the prefix.

The error is in how the provider serves the cache. A read covers the span
up to a marker. A marker on the first message caches tools, system, and
the first message. The system marker already caches tools and system. The
first-message marker adds one message to the cached span and leaves the
rest of the conversation cold on every turn, while the conversation is
the part that grows and the part that costs.

### The tool array

Tools sit at the very front of the cache order, ahead of the system
prompt and the messages. A marker on the last tool declaration closes a
prefix containing all of them.

The tool-array marker is the only one that survives a change to the
system prompt. Without it, editing a single word of the prompt discards
the tool declarations along with it, and tool schemas are not small. With
it, the declarations stay cached while the prompt rebuilds. The system
marker covers tools under normal operation, but the tool-array marker
is the one that matters under maintenance.

### The stable end

One marker sits on the last eligible retained block before ephemera,
even when the two share a wire message.
It advances on each turn as the conversation grows. Because a marker is
not part of the cache key, moving it does not invalidate the bytes it
previously banked. The first-edition Anthropic run recorded a 21-token rise between its two warm
requests, matching that exchange's size. It demonstrates a moving marker with
successful cache reads in that run; it does not promise the same increment for
every response or provider.

A start-of-turn anchor is unnecessary. The rolling marker from the
previous turn already wrote a cache entry at exactly that prefix, and
both vendors walk backward past a bounded number of positions to find a
prior write. The anchor would buy a second copy of something already in
hand, using one of only four slots.

For this layout, an extra start-of-turn breakpoint duplicates the rolling
boundary. If adding one appears to help, inspect the actual prefix before
spending another slot. A rewritten prefix needs its cause fixed; an extra
marker can make that mistake harder to see.

### The compaction bound

A `micro_handoff` rewrites the conversation above it: tool calls are
stripped, opaque reasoning and private replay signatures are removed, and the
affected region settles
into the redaction ladder's terminal state. The stable-end marker, which
sat somewhere in the middle of that region, now points at bytes that no
longer exist. It rebuilds from the new tail and the old prefix is gone,
along with the cache it banked.

The newest surviving `micro_handoff` gives the renderer a useful boundary
between completed work and the continuing exchange. The cuts were recorded and
applied by the reducer, not recomputed to suit today's render. Chapter 22
extends this boundary: if a memory fold removes that handoff, the projection
uses the last surviving memory entry instead. The cache marker describes retained
bytes; it cannot
preserve a dialogue entry that the memory policy has legitimately replaced.

This is a different answer from an agent that keeps demonstration pairs
after compaction. Keeping the three most recent tool-call pairs past the
handoff creates a moving boundary, because what counts as "most recent"
is recomputed on every render. Pairs kept before an older handoff are
dropped once a newer handoff pushes the boundary forward. The marker
must then track the first surviving pair rather than the handoff itself.
An agent whose compaction strip is uniform avoids the moving boundary
entirely: the handoff is the bound, and that is where the marker goes.

Chapter 22's memory-band compaction will rewrite data near the front of the
conversation.
That invalidates the old suffix from the changed point, while earlier tools
and system instructions can still be reused. Accept that cost when replacing
old memory; do not turn memory into frozen instructions to avoid it.

### The vendor asymmetry

Anthropic takes `cache_control` on a tool definition, so it gets all
four markers. OpenAI cannot. Its breakpoints attach to message content
blocks, and the tools array is not a content block. OpenAI writes three
of the four: system, compaction bound, and stable end.

Nothing is lost except independence. A breakpoint includes all prompt
content before it, and tools precede the system prompt in the cache
order, so the system marker already covers the tool declarations. The
system marker simply cannot keep them when the prompt changes. Under
routine operation this makes no difference. Under maintenance, Anthropic
holds its tool cache through a prompt edit while OpenAI does not.

The exact provider lookback is not a client-side algorithm to reproduce.
Earlier documentation even disagreed about its count. What matters here is
that a read and a write are different observations, and a cache write can have
its own full rate. This chapter is about an invisible invoice; write tokens
belong in the accounting too.

That is up to four markers on Anthropic, three on the supported OpenAI route.
An empty tool array or absent compaction bound does not need a placeholder
merely to consume a slot.

### Ephemera

Ephemera are content appended last in `Render` and documented as
"delivered once, then cleared." They may merge into the final user
message. A marker placed after them includes bytes that disappear on the
next turn. That exact boundary cannot match afterward, even if an earlier
part of the prefix still receives a discount.

The stable boundary must be captured before the ephemera append. This is
a single line of code and a paragraph of explanation, but omitting it
leaves the rolling boundary beyond transient bytes. Correct answers alone
will not reveal that placement error.

### The provider minimum

Cache eligibility has a floor, and the floor belongs to a model and route.
The October 11 documentation lists 1,024 tokens for Sonnet-5 and GPT-5.6 Sol,
and 4,096 for Gemini 3.8 Flash. These are token counts, not request-file byte
lengths. Check the selected model's current contract before diagnosing a zero.

A zero can mean the cache failed, or it can mean the request was never
eligible. An instrument that cannot tell those apart will eventually
report the second as the first. The original live script used prompts of a
few hundred tokens and got three Gemini zeroes. Its author first called
caching unimplemented, then corrected the diagnosis after an eligible run.

The old grader fixture grew from 1,716 to 6,631 bytes when five tool
declarations were added. That made its prefix test less trivial, but byte
length alone does not establish a provider's token minimum. A fake can award
cache reads below any real floor; only a real response supplies the grant.

## §15.4 What the wire shows

The first-edition coder reported these live cache-read counts. They are
historical measurements under its earlier marker arrangement, not a new run
of the final implementation:

| vendor | model | cache\_read by turn | what it needs |
|---|---|---|---|
| Anthropic | `claude-sonnet-5` | 0, 3627, 3648 | explicit `cache_control` markers |
| OpenAI | `gpt-5.6-sol` | 0, 7359 | explicit `prompt_cache_breakpoint`; `mode: "explicit"` |
| Gemini | `gemini-3.8-flash` | 0, 0, 16349, 24531 | implicit; history grew by roughly 8,400 tokens per turn |

Three findings in that table.

Anthropic's two warm counts differ by 21 tokens, the reported size of the
previous exchange. That is the stable-end marker banking the conversation
one exchange at a time. It is the direct evidence that a moving marker
does not invalidate the cache, measured rather than assumed. The
assumption held for two sessions before anyone thought to check.

The original OpenAI experiment used the explicit scheme, under different field
names: `prompt_cache_breakpoint` on a content block and
`prompt_cache_options: {mode: "explicit"}` on the request. It was
then documented only for the Responses API in the coder's reading; the
experiment exercised Chat Completions. Its [current reference](https://developers.openai.com/api/reference/resources/chat/subresources/completions/methods/create) also documents
the content marker and request opt-in. The end-to-end measurement,
using that version's renderer output, reads 7,359 of 7,362 prompt
tokens from cache on the second turn. That is 99.96 percent of the
prefix, which confirms the markers land and the prefix survives.

The measurement that matters is the one that almost did not get made.
The first probe sent a request with the marker removed from a position
that had been primed, and got zero reads. The diagnosis that came to
mind was "removing the marker broke the prefix." But the probe had
also changed the content encoding of that message from an array of
parts back to a plain string, and two variables changed at once cannot
diagnose anything. A second probe changed only the encoding and got
7,813 of 7,816 tokens cached: the encoding change preserved that tested prefix. The zero meant "no breakpoint at the primed position
to look up," not "the format change invalidated the cache." This
chapter is about instruments, and the first instrument an instrument
builder needs is the controlled experiment.

The asymmetry between the two explicit vendors is the opt-in. On
Anthropic, a missing marker costs a cache the agent never had.
On OpenAI, explicit mode disables implicit caching, so a request
that declares explicit mode and places no markers caches nothing at
all, which is strictly worse than staying silent. The implementation
must gate the opt-in on evidence: send `mode: "explicit"` only when
at least one marker actually landed.

Gemini needs the shape of the experiment alongside its numbers. It caches
implicitly, detecting the repeated prefix itself, so there are no markers to
place. The original coder tried three shapes: a fixed large user message, a
large stable system reference with a small growing history, and a history
growing by roughly 8,400 tokens each turn.

The first two hit on turn two. The third produced the two leading zeroes in
the table. An early review had tried to disprove the history claim using the
fixed-message probe. Bill's prompting led to the growing-history experiment.
Both measurements were real; the first answered a different question.

The later implementation's browser run used a large stable reference and
four short questions. Gemini warmed on the second request there too, with
15,725, 15,801 and 15,910 cache-read tokens. Those changing values do not
establish a universal 4,096-token block rule. Nor does the earlier four-turn
run prove that a partial block stays uncached forever.

Put genuinely stable content early and keep nearby requests comparable, as
the [Gemini guide](https://ai.google.dev/gemini-api/docs/caching) recommends.
That is not permission to promote editable memories or retrieved documents
into system instructions. This chapter uses implicit caching for Gemini;
it needs no marker budget or separately managed cache object. Measure the
conversation shape the application actually sends before turning a warm-up
observation into a rule.

### The section name

The instability alarm exempts the conversation section, because the
conversation is the one section expected to grow every turn. The exemption
compares against a section name. For Anthropic and OpenAI that name is
`messages`. For Gemini it is `contents`.

A hardcoded constant gets the wrong answer for one of the three vendors.
The lens detects the name from the body's shape rather than from a vendor
table, because the conversation section is the one containing an array
of objects with `role` fields regardless of what the vendor calls it.

The failure mode before the fix: every ordinary Gemini turn was reported
as a broken prefix, with a byte offset that moved between runs. The same
log line already said `contents:appended`, which was the correct
diagnosis. Two halves of one instrument contradicting each other, and the
half offering a precise offset was the wrong one. A precise wrong answer
outranks a vague right one in a reader's attention, and that is exactly
what happened.

### What the fake cannot see

The grader runs against a fake vendor. A fake vendor grants whatever
figures it is told to, so it can prove the request is well-formed and the
lens observes it. It cannot prove the provider agrees.

The first implementation's fake and live runs found different defects.
Neither replaced the other.

The usage meter failed to find a paid model's price. It looked up prices by the
model named in the settings store, which is empty when the model comes
from the environment. An empty name matches nothing, so `claude-sonnet-5`
reported `priced:false, cost:0`. The instrument built to make spending
visible was the thing concealing it. It now prices the model actually in
force, taken from the engine config at the composition root.

The original live `gpt-5.6-sol` Chat Completions run refused requests carrying `reasoning_effort`
alongside function tools, returning a 400. That exercise exposed tools, so every attempted turn failed. The subtle half: the field must be sent as `"none"`,
not omitted. Omitting it means "your default," and the default is a
reasoning effort, so the request comes back refused with the identical
error while the field appears nowhere in the body. The error names a
field the developer can grep for and not find.

The price lookup is ordinary application behavior and a good fake-driven
fixture can catch it. In fact, the retained coder review attributes that
finding to the grader, not the live run. The tool/reasoning restriction needs
a faithful provider contract before a fake can reproduce it. Real use found
the missing fact; a focused regression can preserve it. Keep both kinds of
checks, and be precise about which one found the defect.

## §15.5 The meter

The GUI header shows four numbers: session cost, last-response input
tokens, last-response cache read tokens, and session cache hit rate.
Input and cache-read show the most recent response so the operator can
see how large the current prompt is and how much of it hit cache, rather
than a running sum that only grows. Session totals go to the tooltip on
hover. Cost is computed from Usage and Price at render time and held only
in memory.

The event log holds token counts in four categories: Input, CacheWrite,
CacheRead, Output. Prices live on ModelFeatures. Engine's meter multiplies each model/tier's
counts by its corresponding rates when the GUI asks for a view. A price
correction can change the derived cost without rewriting the event log.
Storing computed cost would freeze the wrong price into the permanent
record.

### Session scope

Cost and hit rate read from Engine's process-lifetime accounting, which
resets on restart. Input and cache-read display the latest main response.
Chapters 22 and 23 will add compressor and judge usage to session totals
and cost without replacing that main-response display. Restored conversation
usage is a different
history and does not seed the current process's meter. The
distinction matters: a developer who opens the GUI and sees yesterday's
total plus today's spend has no feedback loop. "Session cost" means the
cost since the developer sat down.

### Units

Usage counts tokens. Prices are dollars per million tokens. The two meet
only at render time, and the result is denominated in dollars. The event
log never holds a dollar amount. Normalize vendor buckets before applying
prices: Anthropic reports disjoint ordinary/write/read input; OpenAI's prompt
total already includes cached reads and writes; Gemini's prompt total includes
cache reads, and its charged output includes reported thinking. Subtract the
included cached buckets when forming ordinary input, then count each once.
For streams, retain the final usage rather than summing cumulative updates.

The lens predicts a cacheable fraction from the request, measured in
bytes. The provider reports a cache hit rate in the response, measured
in tokens. These cannot be added or subtracted because the units differ
and the tokenizer is not a linear function of byte count. Comparing them
as ratios is legitimate: the prediction is a fraction, the grant is a
fraction, and a large gap is a reason to investigate. A small gap is consistent with
reuse, but is not a tokenization identity or a billing audit. The original Anthropic report paired a 96.8% byte prediction with a
99.4% token grant. The comparison is useful because both remain visible.

## §15.6 Exercise, graded

Extend `solutions/ch14`. Observe actual main sends with an
Engine-owned cache lens; exclude auxiliary requests from that comparison.
Place supported markers at tools, system, the surviving compaction boundary
and the stable end before ephemera. Build the four-number GUI meter from
process-scoped accounting. Keep model/tier identity with incurred usage,
price unknowns honestly and preserve history across a restart that resets
the current process's spend.

Run from the repository root:

```sh
go run ./cmd/grade-edition2 -ch 15 -json
```

The grader builds `solutions/ch15/gui` separately from the core.
`-gui` can select another GUI module; `-cache-current` and `-cache-prior`
select its diagnostic artifact paths relative to the scratch launch directory.
The defaults are `cr/cache-current.txt` and `cr/cache-prior.txt`.
`-cache-prices` supplies the four per-million rates, in ordinary-input,
cache-write, cache-read and output order; its Sonnet-5 default is
`2,2.5,0.2,10`. A historical implementation with a different dated tariff
can be checked against its own rates. These are grader options, not new
application flags.

The public GUI uses `--port`, `--gui-dir`, provider environment settings and
a typed primary skill. Its WebSocket `/ws` accepts a `subscribe` message;
ordinary browser prompts or stdin `{"kind":"prompt","text":"..."}` enter
the same Agent. A `usage` frame reports latest disjoint `input`, `cache_write`,
`cache_read`, `output`; nested `total` contains session counts, and
`hit_rate`, `cost_usd`, `priced` describe the disposable meter. The header
adds the three input buckets; the tooltip exposes the session totals. The
GUI command saves to launch-directory `save.json` by default; its Agent's
lens files still belong under `cr/`.

| check | pts | what it tests |
|---|---|---|
| `cache-lens-exists` | 15 | a cache lens observes consecutive requests, so the actual requests are inspectable |
| `prefix-is-stable` | 20 | tools and system are byte-identical across turns |
| `request-has-breakpoints` | 15 | the last tool declaration, system block and advancing stable history block carry legal `cache_control`, within the four-marker budget |
| `cost-is-computed-not-stored` | 15 | event log holds counts only; cost is derived at display time |
| `usage-is-session-scoped` | 10 | the meter's tokens come from this response, not the lifetime total restored from disk |
| `cache-read-rate` | 10 | cache reads are zero on a cold turn and nonzero on a warm one |
| `all-vendors-report-usage` | 15 | vendor parsers normalize supplied usage without double counting |

The numeric checks drive the real GUI and compare its requests, counts and
artifacts with an external provider fake. Actual browser display, interrupted
worker attribution and full provider marker rules also retain their focused
source/integration checks. Auxiliary exclusion and surviving memory boundaries
join those obligations in Chapters 22 and 23.
The numeric marker fixture has no handoff or ephemeral tail.
A fake's warm reply cannot establish a real cache grant.

## §15.7 Taking It for a Spin

Use a fresh directory and a stable reference large enough to make caching
eligible. The reference belongs in the primary here because it is fixed for
the demonstration, not because arbitrary retrieved text belongs in instructions.
From a scratch copy of `solutions/ch15`:

```sh
cache_source="$PWD"
(cd gui && go build -o /tmp/ensemble-ch17-gui ./cmd)
cache_demo=$(mktemp -d)
cd "$cache_demo"
mkdir -p skills/base
python3 - <<'PY'
from pathlib import Path
header = """---
name: base
type: primary
description: Answer questions about the fixed archive reference.
tools: list_directory read_file
---
Answer the user's archive question in one short sentence, without tools.
The archive project code is COPPER-FERRY-18.
"""
records = "".join(
    f"Archive record {i}: color copper; preserve the original source and record the reason for each change.\n"
    for i in range(1, 701)
)
Path("skills/base/SKILL.md").write_text(header + records)
PY
```

With the provider key configured as in Chapter 0, start the GUI in that
terminal and leave stdin open while using the browser:

```sh
LLM_VENDOR=anthropic LLM_MODEL=claude-sonnet-5 \
EN_PRIMARY_SKILL=base EN_SKILLS_DIR="$PWD/skills" \
/tmp/ensemble-ch17-gui --port 8088 --gui-dir "$cache_source/gui/web/gui"
```

Open `http://localhost:8088` and send these in order, waiting for each reply:

```text
What is the archive project code? Reply in one short sentence.
What is the rule for archive record 37? Reply in one short sentence.
What color identifies archive record 402? Reply in one short sentence.
Remind me of the archive project code, in five words or fewer.
```

The answers should identify COPPER-FERRY-18, preserving the source and reason,
and copper. Watch the header, then inspect its session tooltip. Compare the
last response's usage in `api.log` with the `usage` frame in `gui.log`.
`debug.log` records predicted and granted ratios; plain `diff` on the two
`cr/cache-*.txt` files shows the retained prefix and appended conversation.
A good answer with zero cache reads is still a zero-cache result. Check the
actual input size, eligible route and prefix before deciding why.

Stop with Ctrl-D in the launch terminal after the last turn. Restart the
same command in the same directory. Conversation history should return while
the new process's meter starts at zero; no new response has been charged yet.
That is a separate check from seeing a warm second turn.

The retained October 11 browser sessions made four turns each on the earlier
implementation `7436fcb`. All twelve answers matched the supplied reference.
Their raw provider usage and visible meters were independently compared:

| Historical model | Cache-read tokens by turn | Final session USD | Final session hit rate |
| --- | --- | ---: | ---: |
| `claude-haiku-5-5` | 0, 21924, 21968, 22058 | 0.003481475 | 74.8879% |
| `gpt-5.6-sol` | 0, 16349, 16349, 16349 | 0.1040918 | 74.6439% |
| `gemini-3.8-flash` | 0, 15725, 15801, 15910 | 0.0226167 | 66.2209% |

The dollar values use the dated model prices captured for those runs. The
session percentages include the cold first request; they are not the last
request's hit rate. Haiku 5.5 remains the historical identity, not a current
choice: the executable example uses supported Sonnet-5. Those captures contain
no restarted process or handoff, and precede the lens-overlap correction.
Restart and overlap have their separate real-runtime checks.

The unchanged original grader returned 0/100 because its primary fixture
omitted the required skill type and never started the GUI. That did not
measure seven broken cache features. Conversely, twelve correct answers did
not establish that the instrument could attribute an interrupted reply. The
useful result is the combination: inspect the actual request, preserve the
provider's counts, and make the displayed cost answer to both.
