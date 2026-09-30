# Chapter 18: The Invisible Invoice

Bill's model spend on this book reached $2,475 through chapter 17. A
few hundred of those dollars were wasted on a caching bug that produced
correct answers, on time, every time, while quietly charging ten times
what it should have for every turn it touched.

Every request the agent sends carries a prefix: the system prompt, the
tool declarations, the memory bands, the skill text. That prefix is
nearly identical from one turn to the next. Every major provider offers a
cache: send the same prefix twice and the second read costs a fraction of
the first. With working markers and a stable prefix, the agent now
sustains cache hit rates above ninety percent. The difference between that
and a cold prefix, at Opus rates and a million-token context, is the
difference between four hundred dollars a month and three thousand.

The bug that wasted those hundreds left no trace in the agent's behavior.
No test failed. No log line turned red. The only signal was the invoice,
thirty days later, aggregated past the point where anyone could trace the
spike to the turn that caused it. A cache miss is the only failure in
this program that produces a correct answer.

This chapter builds the instrument that detects the failure, the markers
that prevent it, and the meter that reports the cost in real time so the
invoice holds no surprises.

---

## TL;DR

### Data structures

```go
Price: a field on ModelFeatures, of type Pricing. Every paid model has
one. An unpriced model reports a dash, never "$0.00".

  type Pricing struct {
      CacheWrite  float64   // typically higher than Input
      CacheRead   float64   // typically 10x cheaper than Input
      Input       float64
      Output      float64
  }

Usage: token counts, never money. Prices change; counts are history.
Recorded per turn in the event log, accumulated per session in
UsageCounter.

  type Usage struct {
      Input       int
      CacheWrite  int
      CacheRead   int
      Output      int
  }

UsageCounter: process-lifetime accumulator. Not restored from save.json.
Session cost means cost since the process started, not cost since the
agent was born.

CacheLens: an interface with two methods:
  ObserveRequest(name string, body []byte)
  ObserveUsage(u Usage)
Called at Engine.Turn, not Render. Render has other callers (the judge,
the compressor) that build different conversations. Comparing a judge
prompt against a conversation prompt fires a false alarm.

  NopCacheLens: the default. Does nothing.
  Lens: the real implementation (internal/cachelens/).
```

### Breakpoints

A `cache_control` marker on a content block tells the provider: everything
before this is a cacheable prefix. Anthropic allows four markers per
request, and OpenAI now allows four cache writes under nearly the same
scheme. Place three:

1. **System prompt.** One marker at the end of the system block. Stable
   by construction.

2. **Rolling pair in the message history.** An anchor at the end of the
   previous exchange, fixed for the turn so it scores reads while tool
   rounds advance. A rolling marker at the current stable end of history.
   On each new turn the anchor advances to where the rolling marker was.

3. **The compaction bound.** One marker at the newest `micro_handoff`
   entry. Compaction rewrites everything above it and nothing below it,
   so that entry is the highest point in the conversation that survives
   compaction untouched. The rolling pair chases the tail and is lost on
   every compaction; this marker banks the part that does not move.

That is four markers for three jobs, which is the entire budget. There is
no room for a fifth idea.

The rolling design works because the marker is not part of the cache
key. Moving a marker does not invalidate what it previously banked.

Two constraints:

- **Exclude ephemera.** Ephemera are appended last in Render and cleared
  after delivery. A marker placed after them caches bytes guaranteed to
  disappear, which guarantees a miss. Capture the stable boundary before
  the ephemera append.
- **Strip markers before comparing.** The rolling marker moves between
  turns, so a block carrying `cache_control` in one request carries none
  in the next. Compare raw bytes and every healthy turn is classified as
  edited. The comparison strips markers first: a breakpoint is an
  instruction to the provider, not conversation content.

### Wire format

Anthropic: explicit `cache_control` markers on content blocks. System
must be an array of content blocks rather than a string. Up to four
markers per request.

OpenAI: explicit `prompt_cache_breakpoint` markers on content blocks,
enabled by a request-level `prompt_cache_options: {mode: "explicit"}`.
Up to four cache writes per request. GPT-5.6 and later only; earlier
models cache implicitly and ignore the fields.

The footgun: explicit mode disables implicit caching. A request that
declares explicit mode but places no markers caches nothing at all,
which is strictly worse than staying silent. On Anthropic, a missing
marker costs a cache the agent never had. On OpenAI, it destroys the
implicit cache the agent already had. The opt-in must follow evidence
that a marker landed, never intent.

Gemini: implicit caching, on by default for 2.5+ models. No markers and
nothing to enable. The floor is 4,096 tokens: prompts below it report
zero cached, indistinguishable from a broken cache. The system prompt
caches after one sighting; conversation messages need two successive
requests, so a growing history sees two cold turns at startup. Caching
commits in blocks of roughly 4,096 tokens. Implicit is the recommended
approach: it cannot be aimed at the wrong byte and degrades to "no
discount" rather than to a misplaced marker reporting success.

### The meter

The GUI header displays four numbers: session cost, last-response input
tokens, last-response cache read tokens, and session cache hit rate. The
input and cache-read figures show the most recent response so the
operator can see how large the current prompt is and how much of it was
cached. Session totals go to the tooltip. Cost is computed from Usage
and Price at render time and held only in memory. An unpriced model shows
token counts and a dash for cost.

The meter reads from UsageCounter (process lifetime) rather than
Context.Usage (restored from save.json). Session cost means cost since
the developer sat down.

### Rules

1. Build the lens before the markers. A breakpoint on an unstable prefix
   buys nothing and looks like it worked.
2. Prices live on ModelFeatures. Usage lives in the event log. Cost is
   derived at render time from both.
3. Token counts only in the event log. Prices change; counts are history.
4. The section expected to grow (messages) is exempt from the stability
   alarm. Every other section must be byte-identical.
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

Build from `solutions/ch17`. Wire a `CacheLens` to the engine, place the
breakpoints, build the meter, and verify that the prefix is stable and the
markers hit.

```
make grade18        # must be 100/100
```

### Grader checks

```
cache-lens-exists          (15 pts)  a cache lens observes consecutive
                                     requests; a broken prefix can be found
prefix-is-stable           (20 pts)  tools and system are byte-identical
                                     across turns
request-has-breakpoints    (15 pts)  system and message history both carry
                                     cache_control, and the history one
                                     advances
cost-is-computed-not-stored(15 pts)  event log holds counts only; cost is
                                     derived at display time
usage-is-session-scoped    (10 pts)  the meter's tokens come from the last
                                     response, not the lifetime restored from disk
cache-read-rate            (10 pts)  cache reads are zero cold, nonzero warm
all-vendors-report-usage   (15 pts)  every vendor parser extracts all four
                                     usage categories
```

---

## 18.1 The idea in plain words

A cache miss is the only defect in a coding agent that produces a correct
answer. The tool call works. The edit lands. The test passes. Nothing in
the agent's behavior changes. The failure is pure cost, and it is silent.

Most bugs announce themselves. A bad tool call returns an error. A broken
parse throws an exception. A wedged job hangs where the operator can see
it. A cache miss does none of these things. The correct output arrives,
slightly later, for roughly ten times the price. No log line turns red.
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
invalidated anyway. The lens verifies that the markers hit.

**The meter** displays cost in real time. An invoice is not a feedback
loop. A number that updates on every turn is. The meter does not prevent
waste, but it makes waste visible at the moment it happens, to the person
who can fix it.

The economic argument is straightforward. At the rates Bill pays for
Opus, a million-token request with a cold prefix costs roughly five
dollars. The same request with a warm prefix costs about sixty-seven
cents. A developer running two hundred turns a day pays the difference
two hundred times. Every vendor offers the discount. The engineering is
placing the markers, proving they hit, and reporting the result.

## 18.2 The lens

The lens works by rotating two files: the current request and the prior
one. Each is split into named sections (tools, system, messages) and
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

Because the rolling marker moves, a content block that carried
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

The engineer who built the first strip used a constant holding the exact
marker bytes the renderer emits, with a comment justifying it: "we
control the emitter." The emitter is controlled. But the lens compares
what the canonicalizer produced, and the canonicalizer marshals with
`json.MarshalIndent`. The marker arrives with spaces the constant does
not have.

Nothing matched. The strip did nothing, for every request Bill's agent
had ever sent. Nothing crashed. The symptom was a diagnostic reporting a
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

Measured on one real two-turn run, before and after the fix:

```
before:  messages:edited     66.6%,  0 breakpoints
after:   messages:appended   73.3%,  3 breakpoints
```

The lesson is general. When the instrument and the code share a belief,
the instrument certifies the code against the belief, not against the
artifact. Testing against the artifact, built the way the artifact is
built, is the only way to find the gap.

## 18.3 Where the markers go

The system marker is straightforward. The system prompt does not change
between turns, so a marker at the end of the system block banks the
entire prompt, skills, memory, and all. One marker, one stable prefix.

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

### The rolling pair

The implementation places two markers in the history: an anchor and a
rolling boundary.

The anchor sits at the end of the previous exchange, fixed for the
duration of the current turn. Tool rounds may add new content after it,
but the anchor does not move until the next turn starts. Because it is
fixed, the provider can serve a cache read for the entire conversation
prefix, including every prior exchange, on every tool round within the
current turn.

The rolling marker sits at the current end of stable history, after the
latest user message and before the assistant's in-progress response. On
the next turn this position becomes the new anchor.

Anthropic allows four markers per request. OpenAI allows four cache
writes. The system marker plus the two history markers use three of the
four, which is the budget for the compaction bound below.

The property that makes the design work is that a marker is not part of
the cache key. Moving the rolling marker does not invalidate the bytes
it previously banked. Measured across three live vendors, the cache read
count climbs by exactly the size of the previous exchange on each turn.
That 21-token climb is the marker working.

### The compaction bound

The rolling pair chases the tail of the conversation. A `micro_handoff`
rewrites everything below it: tool calls are stripped, recall entries
are removed, and the affected region settles into the redaction ladder's
terminal state. The rolling pair's anchor, which sat somewhere in the
middle of that region, now points at bytes that no longer exist. The
pair rebuilds from the new tail and the old prefix is gone, along with
the cache it banked.

The newest `micro_handoff` entry is the highest point in the
conversation that survives compaction untouched. Everything below it is
in the terminal state of the redaction ladder, applied once in the
reducer and frozen into the projection. No later cut can rewrite that
region, and no future compaction can move the boundary downward. A
marker on that entry banks everything compaction cannot touch.

This is a different answer from an agent that keeps demonstration pairs
after compaction. Keeping the three most recent tool-call pairs past the
handoff creates a moving boundary, because what counts as "most recent"
is recomputed on every render. Pairs kept before an older handoff are
dropped once a newer handoff pushes the boundary forward. The marker
must then track the first surviving pair rather than the handoff itself.
An agent whose compaction strip is uniform avoids the moving boundary
entirely: the handoff is the bound, and that is where the marker goes.

A memory-band compaction, which rewrites the memory section of the
system prompt, takes a full cache miss. That is accepted. Band
compactions are rarer than handoffs, and a marker cannot survive a
rewritten system prompt regardless of where it sits in the history.

Four markers for four jobs: system, compaction bound, stable end, and
anchor. That is the entire Anthropic budget and the entire OpenAI
budget. There is no room for a fifth idea.

### Ephemera

Ephemera are content appended last in `Render` and documented as
"delivered once, then cleared." They may merge into the final user
message. A marker placed after them caches bytes guaranteed to disappear
on the next turn, which guarantees a cache miss.

The stable boundary must be captured before the ephemera append. This is
a single line of code and a paragraph of explanation, but omitting it
produces a cache that can never be read, and the failure is invisible
because the agent still works correctly.

### The provider minimum

Every vendor has a floor. Anthropic will not cache a prefix below a
minimum size. Gemini's is 4,096 tokens for 3.x Flash and 2,048 for 2.5
and newer. A prompt below the floor reports zero cached tokens, which is
exactly what a broken cache reports.

A zero can mean the cache failed, or it can mean the request was never
eligible. An instrument that cannot tell those apart will eventually
report the second as the first. The live check script that ships with
this chapter originally used prompts of a few hundred tokens. Every
Gemini turn reported zero. The cache was working. The requests were under
the floor.

The practical consequence: a stability claim measured on a request with
nothing in it proves nothing about a request with something in it. The
grader's fixture declares five tools, taking the request to 6,631 bytes
and a 93.9% cacheable prefix. Without them the fixture was 1,716 bytes
and nothing was ever cached.

## 18.4 What the wire shows

The chapter's measured numbers come from a live check against the
real vendor APIs, cache read tokens per turn:

| vendor | model | cache\_read by turn | what it needs |
|---|---|---|---|
| Anthropic | `claude-sonnet-5` | 0, 3627, 3648 | explicit `cache_control` markers |
| OpenAI | `gpt-5.6-sol` | 0, 7359 | explicit `prompt_cache_breakpoint`; `mode: "explicit"` |
| Gemini | `gemini-3.8-flash` | 0, 0, 16349, 24531 | implicit; two cold turns, then ~4096-token blocks |

Three findings in that table.

Anthropic climbs by exactly 21 tokens per turn, which is the size of the
previous exchange. That is the rolling marker banking the conversation
one exchange at a time. It is the direct evidence that a moving marker
does not invalidate the cache, measured rather than assumed. The
assumption held for two sessions before anyone thought to check.

OpenAI now supports the same explicit scheme, under different field
names: `prompt_cache_breakpoint` on a content block and
`prompt_cache_options: {mode: "explicit"}` on the request. It was
documented only for the Responses API and never mentioned for Chat
Completions; it works on both, measured. The end-to-end measurement,
using the renderer's own output with four breakpoints placed at system,
compaction bound, stable end, and anchor, reads 7,359 of 7,362 prompt
tokens from cache on the second turn. That is 99.96 percent of the
prefix, which confirms the markers land and the prefix survives.

The measurement that matters is the one that almost did not get made.
The first probe sent a request with the marker removed from a position
that had been primed, and got zero reads. The diagnosis that came to
mind was "removing the marker broke the prefix." But the probe had
also changed the content encoding of that message from an array of
parts back to a plain string, and two variables changed at once cannot
diagnose anything. A second probe changed only the encoding and got
7,813 of 7,816 tokens cached: string and array forms are
token-identical. The zero meant "no breakpoint at the primed position
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

Gemini needs a fourth column. It caches implicitly, detecting the
repeated prefix itself, so there are no markers to place. What it asks
in exchange is patience, and every part of that patience looks like a
bug the first time.

The two leading zeroes are the warm-up. Gemini caches the system prompt
and tool declarations after one sighting, but a conversation message
must be stable across two successive requests. A growing history takes a
turn longer. An agent sees two cold turns at startup. The actionable
consequence: put large stable content in the system prompt rather than
in the first user message, because the head caches a full turn earlier
than the history does.

The 16,349 and 24,531 are four blocks and six blocks of roughly 4,096
tokens each. Caching commits at that granularity, so a turn that appends
a sentence adds nothing to the cached count. The achievable rate depends
on how the stable region divides into blocks: a prefix of 21,660 tokens
with nothing moving cached 16,359 of them, or 75.5 percent, because the
remainder never completed a fifth block. Against the eligible portion of
a growing conversation the rate is above ninety percent. The raw number
looks lower because the last partial block is charged in full forever.

Gemini's old explicit API required declaring a whole prefix with no
incremental growth and rarely exceeded seventy percent. Its current
implicit mechanism cannot be aimed at the wrong byte, needs no marker
budget, and degrades to "no discount" rather than to a misplaced marker
the log reports as success. Explicit markers would buy back one of those
two cold turns. Over a real session that rounds to nothing.

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

Two of the four defects found during implementation were invisible to
the fake and caught within minutes of going live.

The usage meter priced a paid model as free. It looked up prices by the
model named in the settings store, which is empty when the model comes
from the environment. An empty name matches nothing, so `claude-sonnet-5`
reported `priced:false, cost:0`. The instrument built to make spending
visible was the thing concealing it. It now prices the model actually in
force, taken from the engine config at the composition root.

And `gpt-5.6-sol` refused every request carrying `reasoning_effort`
alongside function tools, returning a 400. An agent always has tools, so
every turn failed. The subtle half: the field must be sent as `"none"`,
not omitted. Omitting it means "your default," and the default is a
reasoning effort, so the request comes back refused with the identical
error while the field appears nowhere in the body. The error names a
field the developer can grep for and not find.

A fake vendor cannot catch either of these. Neither is a protocol error.
Both are defects in how a real provider interprets a well-formed request.
The argument for shipping a live check alongside the grader is empirical.
It found half the bugs.

## 18.5 The meter

The GUI header shows four numbers: session cost, last-response input
tokens, last-response cache read tokens, and session cache hit rate.
Input and cache-read show the most recent response so the operator can
see how large the current prompt is and how much of it hit cache, rather
than a running sum that only grows. Session totals go to the tooltip on
hover. Cost is computed from Usage and Price at render time and held only
in memory.

The event log holds token counts in four categories: Input, CacheWrite,
CacheRead, Output. Prices live on ModelFeatures. The meter multiplies
counts by prices each time it renders, which means a price correction
retroactively fixes every displayed cost without replaying the log.
Storing computed cost would freeze the wrong price into the permanent
record.

### Session scope

Cost and hit rate read from UsageCounter, a process-lifetime accumulator
that resets on restart. Input and cache-read tokens read from the most
recent response stored in the same counter. Context.Usage, restored from
save.json, accumulates across every run the agent has ever had. The
distinction matters: a developer who opens the GUI and sees yesterday's
total plus today's spend has no feedback loop. "Session cost" means the
cost since the developer sat down.

### Units

Usage counts tokens. Prices are dollars per million tokens. The two meet
only at render time, and the result is denominated in dollars. The event
log never holds a dollar amount.

The lens predicts a cacheable fraction from the request, measured in
bytes. The provider reports a cache hit rate in the response, measured
in tokens. These cannot be added or subtracted because the units differ
and the tokenizer is not a linear function of byte count. Comparing them
as ratios is legitimate: the prediction is a fraction, the grant is a
fraction, and a large gap between the two is a finding while a small gap
is confirmation. On the live Anthropic run, the prediction was 96.8% and
the grant was 99.4%.

## 18.6 Exercise, graded

Build from `solutions/ch17`. Wire a `CacheLens` to `Engine.Turn`. Place
breakpoints on the system prompt and in the message history. Build the
meter. Verify prefix stability and cache reads.

```
make grade18        # must be 100/100
```

| check | pts | what it tests |
|---|---|---|
| `cache-lens-exists` | 15 | a cache lens observes consecutive requests, so a broken prefix can be found |
| `prefix-is-stable` | 20 | tools and system are byte-identical across turns |
| `request-has-breakpoints` | 15 | system and message history both carry `cache_control`, and the history one advances |
| `cost-is-computed-not-stored` | 15 | event log holds counts only; cost is derived at display time |
| `usage-is-session-scoped` | 10 | the meter's tokens come from this response, not the lifetime total restored from disk |
| `cache-read-rate` | 10 | cache reads are zero on a cold turn and nonzero on a warm one |
| `all-vendors-report-usage` | 15 | every vendor parser extracts all four usage categories from its response |
