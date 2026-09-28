# Chapter 18 — coder's review

Written after implementing the chapter and testing it against all three live
vendors. I am the coder; I do not edit `book/` prose, so everything here is a
recommendation for the author rather than a change I have made.

Every number below was measured in this session. Where I did not measure
something I say so.

---

## 1. What the chapter asks for that was already built

Six of the brief's seven must-builds existed before I started: `Pricing` on
model features, `CacheWrite`/`CacheRead` on `Usage` with all three vendor
parsers, the session-scoped `UsageCounter`, the `CacheLens` seam and its
`NopCacheLens`, the `internal/cachelens` spoke, and the GUI usage meter.

Two of my opening assumptions about what was missing were wrong, and both were
wrong in the codebase's favour: **all three vendor parsers already extracted
usage**, and **the predicted-versus-granted comparison already existed** at
`internal/cachelens/lens.go`. Worth recording because the session brief and my
own notes both said otherwise. The tree was ahead of the notes.

What was genuinely missing: the second cache breakpoint, and the chapter 18
grader.

---

## 2. Recommended changes to the chapter prose

### 2.1 The second breakpoint is in the wrong place (must fix)

The chapter puts the second breakpoint on **the first user message**, and
justifies it by observing that everything up to the latest exchange is identical
to the previous request.

The observation is true and the conclusion does not follow. Anthropic serves a
cache read for the span **up to** a marker. A marker on `messages[0]` therefore
caches `tools + system + messages[0]` and nothing after it — which the system
marker already covers. The conversation, which is the part that grows, would
stay cold on every single turn, while the lens reported a healthy
`system:identical` and the whole thing looked like it was working.

That is this chapter's own failure mode, in the chapter's own prescription.

**Recommendation.** Describe a rolling pair, which is what Bill approved and
what I implemented: an **anchor** at the end of the previous exchange, fixed for
the duration of the turn so it keeps scoring reads while tool rounds advance,
and a **rolling** marker at the current end of stable history. Anthropic allows
four markers, so there is room.

The prose should also state the property that makes it work, because it is the
whole basis of the design and it is not obvious: **the marker is not part of the
cache key**, so moving it does not invalidate what it previously banked. I had
this as an assumption for two sessions; it is now measured (§3).

### 2.2 The breakpoint must exclude ephemera (must fix)

Bill's ruling during the session, and the chapter does not say it.

Ephemera are appended last in `Render` and may be merged into the final user
message, and they are documented as "delivered once, then cleared". A marker
placed after them caches bytes guaranteed to disappear next turn, which
guarantees a break. So the stable boundary has to be captured **before** the
ephemera append, not after.

This is a short paragraph but it is load-bearing: a student who places the
rolling marker at the literal end of the message array will write a cache that
can never be read, and will see no error.

### 2.3 The chapter must require marker-blind comparison (must fix)

A direct and non-obvious consequence of the rolling design, and currently absent.

Because the rolling marker **moves**, a block carrying a marker in one request
carries none in the next. That difference sits deep inside the common prefix,
where the trailing-append rule cannot reach. Compare raw bytes and every healthy
turn is classified as `edited`, and the reported cacheable prefix is truncated at
the stale marker.

So the comparison must strip markers before comparing, on the principle that **a
breakpoint is an instruction to the provider, not conversation content**. The
chapter should say this, and should say why a JSON round trip is the wrong way to
do it: unmarshalling and re-marshalling reorders keys and rewrites the byte
counts, and those counts are reported as a property of the real request and
compared against the provider's own figures. Both sides would get the same
distortion, so equality would stay honest while the measurement became fiction.

### 2.4 Add the strip's failure mode as teaching material (strongly recommended)

This is the best story in the chapter and it actually happened, to me, in this
session.

My first strip was a constant holding the exact marker bytes our renderer emits,
justified in a code comment on the grounds that *we control the emitter*. We do
control the emitter. But the lens does not compare what the emitter wrote — it
compares what the **canonicalizer** produced, and requests are marshalled with
`json.MarshalIndent`, so the marker arrives as `"cache_control": {` with spaces
the constant did not have.

Nothing matched. The strip did nothing at all, for every request the agent had
ever sent. Nothing crashed, which is why it survived: the symptom was an
instrument reporting a fault that was not there.

Measured on one real two-turn run, before and after the fix:

```
before:  messages:edited    ...  66.6%, 0 breakpoints
after:   messages:appended  ...  73.3%, 3 breakpoints
```

And the reason it survived its own tests is the part worth printing: every test
in the file hand-built its sections as **compact** JSON, matched the constant,
and passed. The tests agreed with the code because both were written from the
same wrong belief about the format, and the format was never consulted. The
replacement test builds its input the way the agent does, and the old compact
test still passes against the broken code — which measures how little it was
checking.

Chapter-level moral, and I think it is the chapter's real thesis: **match the
shape, not a rendering of it**, and build the fixture from the artifact rather
than from your belief about the artifact.

### 2.5 Say that the dialogue section name is vendor-specific (must fix)

The instability alarm has to exempt one section — the conversation — because
that is the one section expected to change every turn.

Ours compared against a constant, `"messages"`, which is right for Anthropic and
OpenAI and wrong for Gemini, where the conversation lives in `"contents"`. Every
ordinary Gemini turn therefore fell through to the alarm and was reported as a
broken prefix, with a fabricated byte offset that moved between runs (`+399`,
then `+147`) because it was measuring an append rather than an edit.

What made it convincing is worth the chapter's space: the **same log line**
already said `contents:appended`. Two halves of one instrument contradicting
each other, and the wrong half was the one that offered a byte offset and a
command to run. A precise wrong answer outranks a vague right one in a reader's
attention.

### 2.6 State what each vendor actually requires, with numbers (must fix)

The chapter currently reads as though caching is uniform across vendors. It is
not, and the measured differences are the most useful thing in the chapter.

Measured, three turns each, real APIs, `cache_read` per turn:

| vendor | model | cache_read by turn | what it needs |
|---|---|---|---|
| Anthropic | `claude-sonnet-5` | 0, 3627, 3648 | explicit `cache_control` breakpoints |
| OpenAI | `gpt-5.6-sol` | 1738, 1759, 1780 | nothing; caches prefixes automatically |
| Gemini | `gemini-3.8-flash` | 0, 0, 0 | nothing; caches automatically **above 4096 tokens** |

Two things in that table deserve prose. First, Anthropic and OpenAI both climb by
**exactly 21 tokens per turn**, which is the size of the previous exchange: that
is the rolling marker banking the conversation, and it is the direct evidence
that a moving marker does not break caching.

Second, **the Gemini row is a measurement error of mine, and it is the most
instructive line in the chapter.** I read three zeroes and concluded Gemini
caching was unimplemented. It is not. Gemini caches *implicitly*: it detects the
repeated prefix itself, sending no directives is the entire interface rather
than an omission, and our request path was already correct. The zeroes came from
the prompts in that run being a few hundred tokens, under Gemini's documented
**4096-token minimum**, so no request was ever a caching candidate.

Re-measured against the live API with a 21,660-token prompt and no directives of
any kind: **16,359 tokens served from cache on the second and every later send.**
Note also that the first send always reports zero, because it is the send that
*creates* the entry, so a three-turn run under the floor produces an all-zero
column that is indistinguishable from an absent feature.

> **Superseded in part.** That 21,660-token probe held the conversation fixed,
> so it showed only one cold turn. A *growing* history has two, and the
> chapter's "two cold turns" claim is correct. Full measurements, and the
> separate warm-up times for the system prompt and the message history, are in
> `book/chapter-18-gemini-review-for-author.md`. Read that one first.

This is worth a paragraph of its own in the chapter, because the failure is
exactly the one the lens exists to prevent: a zero that means *not eligible*
read as a zero that means *broken*. The fix landed as two columns on
`ModelFeatures` (`Caching` and `MinCacheTokens`) so the lens states which case
it is instead of offering the reader three possible causes.

### 2.7 Mention the provider minimum, because a small fixture silently never caches

Anthropic will not cache a prefix below a minimum size. This bit the grader: the
original fixture skill declared no tools, the whole request was 1716 bytes, and
nothing could ever have been cached. Adding five tools took it to 6631 bytes and
a 93.9% cacheable prefix.

The chapter should warn about this directly, because "I implemented it correctly
and measured zero" is otherwise a genuinely baffling afternoon. The corollary is
a design rule worth stating: a stability claim measured on a request with nothing
in it proves nothing about a request with something in it.

### 2.8 Keep the units rule, and name the sanctioned operation

The chapter is right to forbid mixing bytes and tokens. I recommend it go one
step further and say explicitly what **is** allowed, because the code does it and
a reader will otherwise think it is a violation: the prediction is a fraction of
bytes taken from the request, the grant is a fraction of tokens taken from the
reply, and **comparing them as ratios is legitimate while adding or subtracting
them is not**. A large gap between the two ratios is the finding; a small one is
confirmation. On the live Anthropic run: 96.8% predicted, 99.4% granted.

### 2.9 Say that the fake vendor cannot validate caching (must fix)

The chapter should state, as a matter of method, that the grader deliberately
cannot establish that caching works. A fake vendor grants whatever figures we
tell it to, so it can prove the request is well formed by our lights and nothing
about whether a provider agrees.

The evidence is that **two of three vendor paths had defects that the fake could
not see**, and both were found within minutes of going live (§4). That is a
strong argument for the chapter shipping a live check as a deliverable, which it
should probably say outright.

---

## 3. Corrections to smaller factual claims

- **`CacheLens` has two methods, not one.** The brief describes a single
  `ObserveRequest`. The interface also has `ObserveUsage`, and it has to: the
  predicted-versus-granted comparison is the chapter's headline measurement and
  it lives on the second method.
- **Naming.** The brief calls the field `Pricing`. The field is `Price`, of type
  `Pricing`. Trivial, but a reader following along by grep will not find it.
- **The "byte-identical to pre-caching" property is asserted by nothing.** I
  recorded this as a chapter-sized blocker for two sessions. It lives in code
  comments only; `internal/grade/ch10_harness.go` already type-switches on the
  system field being either a string or an array. Any prose implying it is
  enforced should be removed. This is also a good general warning: before
  treating a property as binding, grep for what asserts it.
- **Unpriced models are correct, not missing.** All seven real models are priced;
  the eight `-course` models are deliberately unpriced, and they are the dash
  case the meter is supposed to display. No work needed, but the chapter should
  not imply the table is incomplete.

---

## 4. Defects found, and where they were found

Only the first was in scope. The rest were found by building the measurement and
then actually running it, which is the chapter's own argument made by accident.

| defect | found by | commit |
|---|---|---|
| Breakpoint strip never matched a real request | probing the lens artifact | `3746095` |
| Usage meter priced a paid model as free | grader's first honest run | `40afc7a` |
| `gpt-5.6-sol` failed **every** request | live check | `85e69bb` |
| Every Gemini turn reported as a broken prefix | live check | `b40c718` |

Two are worth the chapter's attention beyond the fix.

**The meter hid the thing it was built to show.** It looked up prices using the
model named in the settings store, which is **empty** when the model came from
the environment. An empty name matches nothing, so a fully priced model reported
`priced:false, cost:0`. The instrument built to make spending visible was the
thing concealing it. It now prices the model actually in force, taken from the
engine config, which is validated at startup.

**`gpt-5.6-sol` was unusable, not merely uncached.** It refuses
`reasoning_effort` in the same request as function tools, with a 400. An agent
always has tools, so every turn failed. The subtle half, which cost a live run:
the field must be **sent as `"none"`, not omitted**. Omitting means "your
default", and the default is a reasoning effort, so the request comes back
refused with the *identical* error while the field appears nowhere in the body.
The error names a field you can grep for and not find. If the chapter's model
table presents this model as usable, it needs either the caveat or the endpoint
migration below.

---

## 5. Left alone deliberately, and why

- **`/v1/responses` for OpenAI.** The proper repair for §4: that endpoint
  supports tools and reasoning together. `Config` already has a `Surface`
  concept for exactly this distinction. It is an endpoint migration rather than
  a flag, so it is a chapter-sized decision and Bill's call, not a drive-by.
- **`gpt-6-astra` is not flagged `NoThinkingWithTools`.** It may well share the
  constraint, but I did not measure it, and guessing which models share a
  constraint is how a one-model fact becomes a rule that applies to everyone.
- **`cacheOrder()` returns Anthropic's order for every model.** Its own comment
  already says the next vendor will need a different one. For Gemini the reported
  section *order* is therefore wrong today. It affects where a prefix is said to
  break, not whether the alarm fires, so it wants its own measurement.
- **Gemini caching.** Now understood and correctly reported: it is *implicit*,
  it works, and we already do the right thing by sending no directives. What
  remains open is narrower than the feature I wrongly scoped here before —
  reading Gemini's `cachedContentTokenCount` is done, but we have not measured
  whether a longer-lived conversation keeps the entry warm between turns.
- **A messages-section breakpoint reported as "granted".** The 96.8% and 93.9%
  figures are **predicted from the request**, not read back from response
  headers. Reading the actual granted rate from the reply is still open.
- **`Makefile` has no `grade17`.** Pre-existing gap; `grade14`–`grade16` and now
  `grade18` exist. Not mine to fix, but someone should.
- **The sweep's leaked-process warning.** Pre-existing: 10 before any of my
  changes, 15 after, and the sweep deliberately reports rather than kills
  because a developer's own agent matches the pattern loosely. No agent
  processes survived when I checked with `ps`. Chapter 18 does add eight
  launches to the sweep though (four scenarios × two targets), and the harness
  uses `cmd.Process.Kill()`, which reaches the process but not its process
  group. If that count keeps climbing, a group kill is the fix, and the
  `Setpgid` plus negative-pid pattern already used elsewhere in the repo is the
  precedent.

---

## 6. Decisions I made that the author may want to overrule

- **Seven checks, not eight.** I broadened `system-has-breakpoint` into
  `request-has-breakpoints`, which now grades both markers and asserts the
  history marker *advances*. I kept the count at seven because the brief's
  seven-check, hundred-point budget is a published contract. If the rolling
  marker deserves its own weight, that is an eighth check and a rebalance, not a
  silent widening.
- **I graded the marker's movement, not its exact index.** Asserting a specific
  position would freeze an implementation detail; asserting that it advances
  between consecutive turns captures the property that pays.
- **The grader measures prefix stability itself rather than reading the lens's
  verdict**, because the lens is one of the artifacts under test. Trusting its
  log line would award full marks to a lens that printed `identical`
  unconditionally.

---

## 7. State

Full sweep: **27 targets, every one at full marks**, `breached=0`. That took two
attempts, and the first one is the more useful fact.

My first version of the marker strip stored the compiled pattern in a
package-level `var`, which is exactly what the chapter 5 `no-mutable-globals`
rule forbids. The agent module built, vetted, `gofmt`'d clean and passed all
eight package test suites with that defect in place, and chapter 18 scored
100/100. What failed was **chapter 5**, and then chapters 7 through 12 via the
chapter 6 parity cascade.

That is the whole argument for the sweep in one incident: a chapter can be
perfect against its own checks while breaking a rule an earlier chapter
established, and nothing local will tell you. It is also a reminder that a
compiled regexp has to live in a `var`, so "I need a compiled regexp" is not an
exemption from the rule — the pattern goes in a `const` and gets compiled per
call. On a diagnostic path that costs microseconds.

Grader: 100/100 at `./agent` and at `solutions/ch18`. Eight mutants, each killing
its own check; one honest overlap, since zeroing Anthropic's `cache_read`
legitimately trips the all-vendors check too. Agent module builds, vets, `gofmt`
clean, all eight packages pass.

Three of the first seven mutants were invalid and every one of them read exactly
like an insensitive check. That produced two rules now enforced by
`scripts/ch18-mutants.sh`, and I recommend the chapter state both: **a verify
predicate must be false before the mutation and true after**, or it certifies
nothing; and **a mutant must compile**, because one that does not proves only
that the compiler works.
