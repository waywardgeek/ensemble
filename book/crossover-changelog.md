# Ensemble Crossover Changelog

Running record of changes made while migrating from CodeRhapsody to Ensemble as
the daily-driver agent. Written for possible use in the book — each entry says
what changed, why, and what it cost.

Started 2026-09-27.

---

## 1. API keys move out of environment variables

**Problem.** Ensemble read every credential from the environment
(`ANTHROPIC_API_KEY`, `LLM_API_KEY`, …). That spreads secrets across shell
profiles, launch agents and process listings.

**Change.** Added `~/.en/settings.json` (mode 0600) holding the three vendor
keys. `pick()` and `envOr()` in `cmd/main.go` now resolve in this order:

1. environment variable — highest priority, so graders keep working
2. `~/.en/settings.json`
3. compiled default

**Why that order.** The graders set env vars to point at a fake vendor. If the
file won the tie, a developer with real keys on disk would silently grade
against the live API. Environment wins so the test harness can always override.

**Cost.** One JSON read at startup, cached for the process lifetime.

---

## 2. Part IDs must be globally unique, not per-response

**Symptom.** In the GUI, the second assistant response rendered *inside* the
first one's bubble. The third rendered correctly on its own.

**Root cause.** Every vendor parser minted part IDs from a counter local to one
API response. `partID(blockIndex)` returned `blockIndex + 1`, so the first text
block of *every* turn was `part_id: 1`. The client keyed its artifact map by
`part_id` and never cleared it, so turn 2's deltas found turn 1's DOM node and
appended to it.

**Rejected fix.** Clearing the client's map at each turn boundary. It works,
but it makes the client responsible for a property the server should guarantee,
and it needs a reliable "new turn" signal on the client — which does not exist
before the first delta arrives. Bill's ruling: *"anything talking to an artifact
needs a unique ID"* — the address belongs to the producer.

**Change.** Added `AllocPartID func() uint64` to `common.StreamCallbacks`. The
actor supplies an allocator backed by its existing `partSeq` atomic counter,
which is monotonic for the life of the agent. All three vendor parsers and the
two shared helpers (`emitLengthOneDeltas`, `emitFinals`) now allocate from it.

**Subtlety that bit once.** `emitLengthOneDeltas` and `emitFinals` are always
called as a pair and must agree on the ID for a given block. Giving each its own
mapper allocated two disjoint ranges and broke correlation. The mapper is now
created by the caller and passed to both.

**Non-streaming path.** Claude's non-stream parser correlates deltas and finals
by Anthropic's block index, so it needs a block-index → allocated-ID map rather
than a bare counter. That is `partIDMapper`.

---

## 3. Favicon

A bright cyan "E" on `#0d0d0d`, as an inline SVG at `web/gui/favicon.svg`. Bill
runs many tabs and could not find Ensemble among them. No image-generation step
needed; an SVG favicon renders at any size and stays legible at 16×16.

---

## 4. Settings: four tabs, real controls

**Problem.** Every setting was a text input in one flat list, including ones
with a fixed set of legal values.

**Change.** Four tabs — AI, Appearance, Accessibility, Memory — and controls
that match the data:

| Setting | Was | Now |
|---|---|---|
| Model | free text | `<select>` grouped by vendor, human-readable names |
| Thinking | raw token count | Low / Medium / High / Max |
| Context target | number on Memory tab | slider on AI tab, 0 → model's context window |
| Theme | free text | `<select>` dark / light / system |
| Font size | free text | slider 8–72 px, live preview |
| TTS speed | free text | slider 0.1–10×, live readout |
| Temperature | free text | **removed** |
| Max tokens | free text | **removed** — comes from model features |
| System prompt | textarea | **removed** — it is generated |

**Why remove rather than hide.** A control that does nothing is worse than no
control: it invites the user to set it and then silently ignores them. Max
tokens belongs to the model, not the user, and the system prompt is assembled
from skills.

---

## 5. Model catalog is server-owned

The GUI used to have no idea what models exist. Added to `internal/common/model.go`:

- `ModelList()` — user-facing models in a deliberate display order
- `ModelListJSON()` — the wire form, with `display_name`, `vendor`,
  `context_window`, `max_output`
- `displayName()` / `vendor()` — presentation, kept next to the table it describes

`current_settings` now carries `models`, so the dropdown is built from the same
table that drives request construction. Ordering is explicit in `ModelList()`
rather than emerging from map iteration.

**Also corrected the table itself.** Opus 4.6 was missing. Opus 5 and Sonnet 5
were recorded at a 200 000-token context window; both are 1 000 000. Max output
was 16 384; it is 128 000.

---

## 6. settings.json is sectioned

**Problem.** Twelve keys at the top level, in no order, mixing model
configuration with font size.

**Change.** `common.Settings` is now four structs — `AI`, `Appearance`,
`Accessibility`, `Memory` — matching the tabs. Patches from the GUI are nested
the same way (`{"appearance": {"font_size": 18}}`).

**Migration.** `NewSettingsStore` tries the nested shape first; if `ai.model` is
empty it re-reads the bytes as the old flat shape and rewrites the file. One
path, run once, no flag.

---

## 7. Font size had no effect

**Two bugs, both invisible from the code.**

1. Every `font-size` in `style.css` was an absolute `px`, so setting
   `documentElement.style.fontSize` changed nothing downstream. All 21
   declarations are now `rem`.
2. `if (app.font_size)` skipped the value zero, which is exactly what a
   never-set font size deserializes to. Zero now means "use 16".

The slider also applies on `input` rather than only on `change`, so the page
resizes while dragging instead of when released.

---

## 8. Ctrl-C did not save the conversation

**Symptom.** `save.json.journal` existed; `save.json` never did.

**Root cause.** There was no signal handling at all. In server mode the main
goroutine blocks on `for in.Scan()` over stdin, and the save runs *after* that
loop returns. SIGINT killed the process where it stood, so the save was
unreachable by construction — not a race, simply never run.

**Rejected fix.** A signal handler that saves and calls `os.Exit(0)`. Exit skips
deferred functions, and the defers here are load-bearing: `journal.Close()`,
`hub.Close()`, `srv.Close()`, `cancel()`.

**Change.** stdin moved to its own goroutine that closes `stdinDone`; the main
goroutine now selects on `stdinDone` versus `SIGINT`/`SIGTERM`. Either way it
falls out of the select and leaves through the existing exit path, so the save
and every defer still run. A blocking `Scan` cannot be cancelled, so on a signal
that goroutine is simply abandoned — which is fine, the process is leaving.

**Verified end to end**, not by reading. Before: no `save.json`, journal 17 861
bytes. Send SIGINT. After: `interrupt received, saving conversation...`, clean
exit (not killed), `save.json` 49 555 bytes of valid JSON.

**Fixture note.** The first attempt backgrounded the process, which gave it an
immediately-closed stdin: it exited through the normal path before any signal
arrived, and the run proved nothing. Holding stdin open with a long-lived writer
is what makes the fixture actually produce the condition under test.

---

## 9. The GUI opened blank on a restored conversation

Bill asked for history to be replayed into the GUI at startup. The protocol to
do it already existed — `subscribe` → `event_range` → `fetch`, with
`renderEvent` handling `ResponseEnded`, `ToolCalled`, `ToolReturned` and
`MessageReceived`. The client already spoke all of it. Nothing needed building;
two bugs needed removing.

### 9a. The hub believed its log was empty

`Hub.logLen` — the hub's notion of how much of the log is renderable — was only
ever advanced inside `Observe`. It began at zero. On a restored session the log
is already full but no observation has arrived yet, so `logLen` was still zero
and the replay loop ran `for i := 0; i < 0`. The hub was holding the entire
conversation and reporting that none of it existed.

`NewHub` now seeds `logLen` from the log it is handed.

The ordering around it was already correct and worth recording, because it is
the part that looks suspicious and is not: `eng.Log.Events = sf.Log` assigns the
*field*, so the restore mutates the same `Log` the hub is later given a pointer
to. Had it replaced the pointer, seeding `logLen` would not have been enough.

### 9b. Replay dropped silently once the buffer filled

Every send used `select { case c.send <- msg: default: }`. Dropping is right for
a live frame, which a newer frame supersedes. It is wrong for a replayed frame,
which is never re-sent: the drop is a permanent hole in the history the user is
reading. A long conversation renders to far more messages than the 256-slot
client buffer. Replay now waits for room, with a five-second bound so a dead
client cannot pin the goroutine.

### The test I nearly shipped as decoration

Both fixes have regression tests in `internal/ws/replay_test.go`, and both tests
passed the moment I wrote them. That is worth nothing on its own, so each was
audited by deleting the behavior it protects:

| Mutant | Test that failed | Message |
|---|---|---|
| remove the `logLen` seeding | `ReplaysRestoredLog` | replayed **0** of 2 |
| restore the dropping send | `DoesNotDropWhenBufferOverflows` | replayed **255** of 400 |

The second mutant **survived the first version of its test**. The test had a
collector goroutine running before `subscribe`, so the buffer never filled and
the dropping branch was never reached — it passed cleanly against the bug it
claimed to catch. The fix is to let `subscribe` run ahead for a moment and only
then start reading. Both mutants had to compile; one that breaks the build
demonstrates nothing.

A detail that confirms the diagnosis rather than merely fitting it: the mutant
delivers 255 of 400, not 256. The `event_range` frame sent before the replay
loop occupies the first slot.

## 10. The composer grows with the message

**Problem.** The chat input was `<input type="text">`, which is a single line by
definition. Long messages scrolled sideways through a one-line slot.

**Change.** It is now a `<textarea rows="1">` that grows to fit its content up
to a cap, then scrolls.

**Why the cap lives in CSS rather than JS.** The obvious implementation puts a
`MAX_HEIGHT = 200` constant next to the resize code. That silently breaks the
font-size slider: at 32px text the box still stops at 200 physical pixels, so it
holds half as many lines as it did at 16px. The cap is `max-height: 11rem` in
the stylesheet and the resize function reads it back with
`getComputedStyle`. One number, expressed in the same units as everything else
it has to agree with.

**The detail that makes shrinking work.** `autoGrow` sets `height: auto` before
reading `scrollHeight`. Without that first line the element's own height floors
the measurement, so the box grows when text is added and then never comes back
down when it is deleted.

**Two clear sites, not one.** Assigning `.value` programmatically does not fire
an `input` event, so both places that empty the box — Enter-to-send and
Escape-to-cancel — have to shrink it by hand. Fixing only the send path leaves a
tall empty box behind every Escape.

**Enter semantics had to be restated.** An `<input>` cannot contain a newline,
so `Enter` unambiguously meant send. A `<textarea>` inserts one, so the handler
now takes plain Enter as send with `preventDefault()`, and leaves Shift+Enter to
the browser as a newline.

`font-family: inherit` was already on the rule, which happens to cover the other
common trap here: a bare textarea renders in monospace.

## 11. Prompt-cache lens: prediction vs actual

**Problem.** Ensemble sets no cache breakpoints at all (`grep -rn cache_control internal/`
returns nothing; every "ephemeral" hit is about ephemeral *tools*, a different concept).
Every request is a full cold read. At 1M context and Opus prices that is the single
largest cost item, and CodeRhapsody runs at 91.3% cache hit rate for comparison.

But markers were not built first. Measurement was.

**Why measurement comes first.** A cache miss is the only failure in this program that
produces a *correct answer*. Nothing reports it: no error, no exception, no failing
test. It surfaces on the invoice thirty days later, aggregated past the point where you
could tell which change caused it. Every other class of bug announces itself; this one
does not, so the signal has to be manufactured. And a breakpoint placed on an *unstable*
prefix buys nothing while looking like it should have worked — so proving stability is a
prerequisite, not a follow-up.

**Built.** `internal/common/cachelens.go` (the seam: `CacheLens` interface + `NopCacheLens`)
and `internal/cachelens/` (the spoke):

- `canonical.go` — splits a request into named sections and re-emits them in **cache
  concatenation order**, which for Anthropic is tools → system → messages. Note this is
  *not* the struct's JSON field order (system precedes tools there), so raw bytes cannot
  be diffed directly; the canonicalizer has to reorder. Sorting by "how it lays out in
  context" was load-bearing, not cosmetic.
- `diff.go` — per-section comparison with a verdict per section.
- `lens.go` — rotates `request.json` / `prior_request.json` so `diff` puts the cursor on
  the first byte that broke the cache.

**Hooked at the call site, not inside Render.** `Render` has three other callers: the
relevance judge, the memory compressor, and `RenderOnly` (a CLI debug helper). The judge
and compressor build *entirely different conversations*. Had the hook gone inside
`Render`, a judge call landing between two turns would become `prior_request` and the
diff would compare a judge prompt against a conversation. Hooking `Engine.Turn` excludes
them by construction rather than by filtering. `common.BodyOf` reads from `req.GetBody`,
which returns a fresh reader, so the live body is not consumed.

**Three verdicts, not one number.** MATCH / PREFIX DIVERGED (ours, and the diff localizes
it) / CACHE MISSED ANYWAY (TTL expiry, the size floor, eviction — normal and frequent).
Collapsing these into one hit-rate figure would make the instrument unactionable, and
firing the same alarm for the third case would make it noise within a day.

**Units.** Prefix is bytes; providers report tokens. There is no exact conversion without
the vendor's tokenizer, so the report gives both and alerts on the *ratio*. "Did we lose
the cache" is nearly binary and does not need token precision. An estimate dressed as a
token count is a number that gets requoted later as though it had been measured.

### The bug the lens found in itself

First live run, inside the ch09 and ch14 graders:

```
cachelens: edited: common prefix 13565 of 14159 bytes, first change in messages at +150
cachelens: PREFIX DIVERGED in messages at +304
```

That alert is a **false positive**, and it exposed a real defect in the classifier.
Appending a message to a JSON array turns `[{"a":1}]` into `[{"a":1},{"b":2}]`, so the
two strings first differ *at the closing bracket* — prior has `]` where current has `,`.
Compared naively that reads as an edit. And it happens on **every single turn**, so the
tool would have cried wolf constantly and been muted inside a day. A false alarm that
always fires is worse than no alarm: it also trains you to ignore the real one.

Fixed by (a) comparing per-section rather than over one flat blob, since a cache
breakpoint sits at a section boundary and the provider's question is "which sections are
untouched, from the front", not "how many bytes match"; and (b) treating a grown JSON
container as an append — trim the prior's trailing framing and check whether what remains
is a prefix of the current. The alarm now fires only when a section *above the dialogue*
changes, which is the only case that is both a bug and ours.

A second defect surfaced from a test: `CacheableBytes` summed raw payloads while
`CurrentBytes` used the canonical form's length, which includes the `name=` framing the
tool itself adds. Two numbers that get divided must be measured with the same ruler, or
the percentage reports the formatter rather than the request.

**Mutation-audited.** Both fixes have tests that die to a compiling mutant, each killing
exactly one test: removing the framing detection kills
`TestGrowingDialogueIsAnAppendNotAnEdit`; removing the cacheable-run gate kills
`TestCacheableRunStopsAtFirstChangedSection`.

**Still to do:** the breakpoints themselves. `anthRequest.System` is a plain `string`, and
`cache_control` attaches to a *content block*, so system must become an array of blocks.
That breaks the byte-identical-to-a-pre-caching-request property the surrounding comments
maintain deliberately, which is a chapter-sized decision rather than an edit.

## 12. Settings: sectioned file, flat wire (a restructure, reverted)

The request was for `settings.json` to be grouped into sections matching the dialog tabs.
The first implementation sectioned `common.Settings` itself — and that was wrong, because
that struct **is the published WebSocket contract**. ch09's grader sends flat
`{"theme":"light"}` and requires flat keys back in the `settings_changed` broadcast.
Sectioning the type dropped ch09 from 100 to 20.

The distinction that resolves it: the *file* is a human-editable artifact, the *wire* is a
published protocol, and the *tabs* are presentation. Three consumers, and only one of them
was asking for sections. The type stays flat; grouping belongs to whoever owns the file.

Reverted the struct, the store, the test file, and the GUI's patch builder to flat. The
GUI comment now records why, so the next person does not redo it. Sectioning the on-disk
file remains to be done, in the settings spoke where the file-format knowledge belongs.

**Also fixed, same investigation:** `var homeSettings = loadHomeSettings()` from change 1
was a package-level mutable global, which fails ch5's `no-mutable-globals` check and
cascades through ch6→ch7→ch8→ch9. Replaced with an on-demand `homeSetting(key)` that
re-reads the file. Not caching it is deliberate: the file holds three keys and is read
only while assembling startup config, so the cost is microseconds once, and that is a
good price for keeping a process-wide mutable global out of the tree.

## 13. The settings dialog never saved anything

`sendPatch` was called by every settings control and **was never defined** — a
`ReferenceError` on every change event. ch14 fell to 5/100 and the failing checks carried
*empty* messages, because the harness could not enable TTS through the settings channel.

The reason this went unnoticed is worth recording: the dialog *looked* correct. Theme and
font size both apply locally on the way past, so the visible half worked while the durable
half never ran. Every setting appeared to take effect and none was persisted. A control
that shows its effect immediately can hide a broken write indefinitely.

Found by reading the harness log rather than by inspection. Four wrong hypotheses came
first — stub-DOM `getComputedStyle`, a missing `agentState` global, the textarea `input`
handler, the CDN pins — and the log named the real cause on the first line. The bisect
(reverting `web/gui` → 100/100) had correctly identified the *file* immediately; what took
time was guessing at the line instead of reading the error.

**ch09 and ch14 both restored to 100/100.** Agent module: 7 packages, 0 failures.

---

## 14. Session meter: cost, tokens, cache hit rate

**Why.** Caching was the crossover blocker, and caching cannot be adopted on
faith. Ensemble sets zero cache breakpoints today, which means every request
is a full cold read; at a million tokens of context and Opus rates that is the
single largest cost in the system. But a breakpoint placed on an unstable
prefix buys nothing *while looking exactly like it should have worked*. So the
meter comes before the markers: first make the money visible, then move it.

The reference point is a measured one. CodeRhapsody runs at **91.3%** cache
hit rate. That is the bar Ensemble is being measured against, and it is
recorded in `TODO.md` so it does not decay into "high".

**What shows.** Four numbers in the status bar, refreshed on every turn end
and again whenever a browser connects:

| Field | Meaning |
| --- | --- |
| `$0.0834` | session cost, computed from the vetted price table |
| `in 187K` | total input tokens, all three input categories summed |
| `cache 184K` | the subset of those served from cache |
| `91.3%` | hit rate, `CacheRead / (Input + CacheWrite + CacheRead)` |

Hover gives exact unrounded figures for all four categories.

### Where the accumulator lives, and why not in the event log

The obvious place to read a session total is `Context.Usage`, which the
reducer already maintains. That number is wrong for this purpose, and the
reason is worth recording: `SaveFile` persists the whole `Context` and
`sf.Restore` restores it, so `Context.Usage` is a **lifetime** tally spanning
every run the save file has ever seen. Displaying it as "session cost" would
overstate the figure by however many previous sessions existed, and would do
so in the direction that makes caching look like it is not working.

Bill's framing settled it: the lifespan of the accumulator should be the
lifespan of the agent. So the counter is process-scoped, held on the object
whose life *is* the process.

Rejected: capturing `Context.Usage` at startup and reporting
current-minus-baseline. It depends on construction order relative to the save
restore, and if the agent were ever built before the restore it would silently
report a lifetime total as a session total. A process-scoped counter has the
right lifespan by construction rather than by careful sequencing.

### An interface, not a callback

The first implementation was going to be `Engine.OnUsage func(common.Usage)`,
because the engine records usage and the thing that wants to display it lives
elsewhere. Bill stopped it, and the standing rule is worth quoting: before
writing any callback, ask, because a callback is almost always a symptom of
Go import-cycle pain, and the cure is an interface in `internal/common`.

He was right, and the fix needed *no new wiring at all*. `common.Host` is
already documented as "the root interface every object can reach through its
parent chain". Every object already holds one. So `Host` gained two methods:

```go
RecordUsage(Usage)
SessionUsage() Usage
```

and `common.UsageCounter` (a mutex and a tally) is embedded by the
implementors that should accumulate. The engine calls
`e.Host.RecordUsage(...)` on `ResponseEnded`, one line next to the existing
cache-lens hook.

Two implementors deliberately do *not* embed the counter:

- `jobs.Jobs` **delegates** to its parent host. Embedding would have compiled
  and then silently swallowed every count into a struct nobody reads. A
  spoke's job is to pass the tally up to the root, not to keep its own.
- `testHost` gets explicit no-op methods because it is a **value type**, and
  embedding a struct containing a mutex would copy a lock.

The hub reads through `common.UsageSource`, a one-method interface, rather
than through `Host`. The hub needs to read the tally and must never add to
it, and the narrow interface says so in the type system. It is nilable and
nil for the CLI and for graders, which is the correct behaviour: no browser,
no meter.

### Prices: explicit per model, never a multiplier

Pricing was copied from CodeRhapsody's vetted table, which carries four
per-million figures matching `Usage`'s four disjoint categories exactly.

The temptation is to store one input price and derive the rest, since
Anthropic's public rates are a tidy 1.25x to write a cache entry and 0.1x to
read one. **That derivation is wrong for most rows in the table.** One
Anthropic model reads at 0.05x. Gemini charges no write premium whatsoever
(`CacheWrite == Input`). Several models support no caching at all. A computed
multiplier would produce confident, wrong numbers for the majority of the
catalog, so all four rates are stored explicitly.

Zero means **unpriced, not free**. The course and fake models used by graders
have no price sheet, and an unpriced model must not be able to masquerade as a
free one, so the wire frame carries a separate `priced` boolean and the GUI
renders `$—` rather than `$0.00`. Both arms were verified against a live
server: `claude-opus-5` returns `priced: true`, `claude-opus-5-course`
returns `priced: false`.

`TestEveryOfferedModelIsPriced` walks `ModelList()` — exactly the set the GUI
dropdown offers — and fails if any of them lacks a sheet. That is the guard
that matters when someone adds a model six months from now: you cannot offer
a model in the dropdown and then show a confident wrong price for it.

### Output is not in the hit-rate denominator

`CacheHitRate` divides cache reads by the sum of the three *input*
categories. Output is excluded because no cache could ever have supplied an
output token, and including it would make the rate sag whenever the model
simply said more. A verbose turn would read as a caching regression.

This is pinned by a test that computes the rate for two tallies identical
except for output length and requires the same answer. Under the mutant that
adds output to the denominator the two diverge to 0.79 and 0.07 — the whole
meter collapsing purely because the model talked more.

An empty session returns 0 rather than NaN, because the alternative renders
as `NaN%` in the status bar the moment the page first connects.

### Cost is computed server-side

The event log stores counts and never money, on the standing rule that prices
change while counts are history. The *display* frame is a different thing: it
carries dollars, computed in Go from the one price table, so that the table
has exactly one home and is not duplicated in JavaScript.

`CostUSD` returns zero for an unpriced sheet and does not try to distinguish
"free" from "unknown", because zero is also a perfectly valid cost for a
session that has sent nothing. That judgment belongs to the renderer, which
has somewhere to put a dash.

### Presentation details that are not cosmetic

- **Tabular monospace figures.** The meter updates every turn. In a
  proportional font every digit change shifts the whole row sideways.
- **Compact units** (`187K`, `1.24M`). A working session runs to millions of
  tokens; the raw digits are unreadable and wide enough to reflow the bar.
- **Four decimals under a dollar.** With two, the first several turns of every
  session all read `$0.00` and the meter looks broken rather than cheap.

### "Don't be evil"

Sits at the foot of the sidebar in small, quiet type, with a border above it.
It was first placed in the status bar next to the meter; Bill moved it to the
sidebar, which is better — a creed does not belong in a row of live counters.

Placing it surfaced a latent layout bug: `#sidebar-content` already carried
`flex: 1`, but `#sidebar` was never made a flex column, so that declaration
had been a no-op. Adding `display: flex; flex-direction: column` completed a
design that was already written down and never turned on.

### Verified

- `go build ./...`, `go vet ./...`, `gofmt -l` clean; 9 test packages pass.
- Three mutants, each compiling and each killing exactly its own test:
  output in the hit-rate denominator; cache read charged at the input rate; a
  price removed from an offered model.
- `node --check` on all five GUI scripts, and the script-tag count confirmed,
  after the earlier incident where an edit silently ate four of them.
- Live WebSocket probe against a real server on an isolated port, for both
  the priced and unpriced paths.
- ch09 **100/100** (settings wire, including the ch8 parity cascade) and
  ch14 **100/100** (GUI, all ten checks) — the two graders most sensitive to
  this work, each run alone.

### Still open

Breakpoints themselves. `anthRequest.System` is a plain string, but
`cache_control` attaches to a content *block*, so `System` must become an
array of blocks. That change breaks the byte-identical-to-a-pre-caching-request
property the surrounding comments maintain deliberately, which makes it a
design decision rather than an edit.

### Gotcha for the next person

`go run ./cmd/grade -ch 09` fails with `exit status 2` and prints usage. Go's
flag parser rejects `09` as an invalid octal literal. Use `-ch 9`.

Worse, wrapped in a watchdog subshell the failure reports **exit 0**, because
the exit status belongs to the subshell rather than to the grader. A grader
invocation that never ran looks exactly like a grader invocation that passed.

## 15. The system prompt carries a cache breakpoint

Ensemble sent zero cache breakpoints. Every request re-read the entire prefix at
full price, while the same workload in a comparable agent runs at a 91.3% cache
hit rate. The lens from entry 11 could measure the problem but not fix it.

The apparent blocker, recorded in entry 14 as "chapter-sized": `cache_control`
attaches to a content *block*, but `anthRequest.System` was a plain `string`.
Converting it to an array changes the JSON shape of every request ever sent,
including chapters that predate caching, which seemed to break the
byte-identity property the surrounding comments maintain on purpose.

That framing was wrong twice over, and both corrections came from looking at
artifacts instead of reasoning from the struct.

First, the property was never asserted by anything. The graders that read the
system field already tolerate both shapes: `ch10_harness.go` type-switches with
the comment "The system field can be a string or an array of content blocks",
and ch15 through ch17 hold it as `json.RawMessage`. The invariant lived in
comments, not in checks.

Second, a production request settled the design in one read. The shape is not
elaborate: one block, the whole prompt, one marker.

```json
"system": [{"type": "text", "text": "…", "cache_control": {"type": "ephemeral"}}]
```

No sectioning, no per-chunk markers. The system prompt is the largest stable
span in the request, so the single breakpoint at its end covers tools and system
together and pays for itself immediately.

### Why a separate block type

`anthSystemBlock` is deliberately not `anthBlock`. That type's `MarshalJSON`
returns its `Raw` bytes verbatim for opaque replay material, so a
`cache_control` set on a replayed block would be silently dropped: the marker
would be present in Go, absent on the wire, and the only symptom would be a bill.
A separate three-field struct cannot fail that way.

### Two small decisions

The breakpoint is unconditional. A prefix below the vendor's minimum cacheable
length is processed uncached rather than rejected, so there is no size to check
and no setting to get wrong.

An empty prompt returns `nil`, not an empty slice, so `omitempty` drops the key
entirely and a promptless request keeps exactly the bytes it always had.

### Rejected

A custom `MarshalJSON` emitting a bare string when no marker is set and an array
when one is. It would preserve byte-identity exactly, but it adds a marshaller
to protect a property that no check asserts and no reader depends on. The
simpler shape is the one the vendor documents.

### Verified

Three tests in `internal/llm/caching_test.go` lock the wire shape. Both mutants
were killed by exactly the tests that should kill them: dropping the marker
failed the two breakpoint tests and left the omitempty test passing, and
returning an empty slice instead of `nil` failed only the omitempty test. Build,
vet, gofmt and all agent packages clean; ch10 and ch14 each 100/100, run alone.

### Still open

A second breakpoint inside `messages`, which would extend the cached prefix
across the conversation as it grows. The production request uses one there too,
on the first user message. That one needs measurement first: the lens already
computes predicted-versus-actual hit rate, so the question of whether a moving
message breakpoint earns its slot should be answered with a number rather than
an argument.

## 16. A reset button that clears the screen, not the agent

The GUI had no way to start a fresh conversation. Restarting the process was
the only reset, which also threw away the running job table and every
connected client.

### The event, not the edit

The obvious implementation clears the context and returns. It works, and then
it silently undoes itself on the next restart, because the context is a
projection of the event log: a mutation the log does not record is a mutation
that replay does not reproduce. So the reset is an event, `ConversationReset`,
recorded through `Engine.Record` like every other removal. Durability then
comes for free rather than as a second step to remember, since the journal
replays on load and re-applies the reset.

This is the same rule the rest of the context obeys, and worth restating
because the shortcut is so inviting: whoever changes the context owns adding an
event that reproduces the change.

### The safety property

Every kind of entry lives in one slice. Memory bands, the soul and memory
documents, loaded skills, the tool roster, checkpoints and the conversation are
all `c.Dialogue`, told apart only by `Kind`. The one-line implementation,
setting that slice to nil, would delete the agent's memory in order to clear
its screen, and it would look like it worked.

So the reducer filters by kind, removing `KindDialogue` and `KindRecall` and
nothing else. Recall goes with the dialogue for the reason it also goes at a
checkpoint: an auto-recalled memory is an input to a conversation, and once the
conversation is gone it is answering a question nobody asked.

### Two details that would have been bugs

`ArtifactScroll.clear()` empties three correlation maps as well as the DOM.
Those maps key streaming deltas to the elements they belong to, so an entry
outliving its element would route the next turn's text into a node no longer on
the page: text vanishing, no error anywhere.

The reset reaches the screen as an ordinary rendered event rather than as a
reply to the button press. A reconnecting client rebuilds its page by replaying
the log, so anything delivered only as a reply would leave a reconnected client
showing a transcript the agent no longer has.

### Verified

Four tests, four mutants, each killed by exactly the test that should kill it:
clearing the whole slice, dropping the turn-state reset, unregistering the wire
name, and dropping the event from the renderable set. ch9 and ch14 both 100/100.

One mutant initially appeared to survive, which would have meant a decorative
test. It was the mutation that was wrong, not the test: `perl` without `/g`
rewrote the first of three identical lines in the file rather than the intended
one. An invalid mutant and an insensitive test look identical in the output, so
a surviving mutant is worth confirming before it is believed.

### A measurement, from entry 15

The ch9 grader run reports the lens output, and after the system breakpoint
landed it reads:

```
tools:identical system:identical messages:appended
cacheable prefix 13657 of 14213 bytes (96.1%)
```

`system:identical` is the breakpoint doing its job: the largest span in the
request is byte-stable across turns. At 96.1% already cacheable on a short
conversation, the second breakpoint inside `messages` has little left to buy
here, though a long conversation would shift the balance since that is the only
section that grows. Still open, and still a question for a measurement rather
than an argument.


---

## 17. GUI reconnect: assistant text and thinking vanish on restart

**Symptom.** After restarting the server, the GUI showed Bill's chat messages
(dark bubbles) and tool calls (right panel), but every model response and
thinking block was gone — as if the agent had never spoken.

**Root cause.** `partToWireMsg` produced `part_final` messages with no
`part_id` field. The GUI's `_handleFinal` uses `msg.part_id` as a key in its
`artifacts` Map. With `part_id === undefined`, every replayed `part_final`
overwrote a single DOM element — the first one created, positioned after the
first user message and scrolled off the top. 63 agent responses collapsed into
one invisible div.

User messages survived because `_handleUserMessage` always creates a new
element without consulting the Map. Tool calls survived because
`_handleToolDispatched` also always creates a new element.

**Why it took tracing, not guessing.** The save file was correct: 392 events,
63 `ResponseEnded`, all with proper `TextPart` content. The `PartList` JSON
marshaling round-tripped every part type correctly. The subscribe handler sent
events. The GUI received them. The rendering code handled `TextPart`. Every
layer worked in isolation. The bug was a missing field at the boundary between
the server's wire format and the client's DOM correlation.

**Fix (a1eebb4).**

1. `partToWireMsg` now takes a `partIndex` parameter and generates a unique
   `part_id` from the event seq and part position (e.g. `r42.0`, `r42.1`).
   Each response gets its own div.

2. `OpaquePart` with `type: "thinking"` in its JSON data now renders as
   `kind: "thinking"` on reconnect, extracting the thinking text. Previously
   `partToWireMsg` returned `nil` for `OpaquePart`, so thinking was visible
   during live streaming but gone after restart.

3. `artifact-scroll.js` applies the `.thinking` CSS class when
   `msg.kind === 'thinking'` in `_handleFinal`, matching the live-streaming
   path that does the same in `_handleDelta`.

**Tests.** `TestResponsesGetDistinctPartIDs` (multiple responses get distinct
IDs) and `TestThinkingRendersOnReconnect` (opaque thinking parts render with
correct kind and text).

**Lesson for the chapter.** The live streaming path and the replay path are
different code paths that must produce equivalent output. A field that is
present in `marshalObservation` (the live path) but absent in `partToWireMsg`
(the replay path) creates a class of bug that is invisible during development
— it only appears after a restart, which is the one time the user most needs
the transcript to be intact. The fix is not "add the field" but "share the
correlation key": the `part_id` is the identity of a piece of content, and
identity must survive both paths.

## 18. The reset that cleared the screen for everyone except the operator

Reset worked, and the operator could not tell. Clicking it cleared the screen,
and the old conversation stayed on screen anyway. A browser refresh removed it.

That asymmetry is the entire diagnosis, and it points the opposite way from
intuition. Refresh is the path that goes all the way back to disk and rebuilds
from scratch, so a bug that *survives* a refresh is a storage bug. A bug that a
refresh *cures* is the reverse: the durable record is right and the live screen
is the thing that is wrong.

### The fix for one path was the gap in the other

Section 16 made a deliberate choice, and it was the correct one:

> The reset reaches the screen as an ordinary rendered event rather than as a
> reply to the button press.

That is exactly right for a reconnecting client, and it is exactly why a
connected one saw nothing. The GUI is a projection of the event log fed by two
different channels. Replay reads the log; a live screen never replays anything
and learns about the agent only through observations. Delivering the reset as a
rendered event served the first channel completely and the second not at all.

`Engine.Record` appends to the log, writes the journal and applies the reducer.
It does not notify the observer, so `handleReset` produced *zero* observations.
Measured, not inferred: attaching a capture observer and calling the handler
returns an empty list.

Replay, meanwhile, was provably fine. Driving `Hub.subscribe` over a log of
`[message, message, ConversationReset, message]` yields frames in that order,
so a replaying client renders the stale transcript and then clears it. The
self-healing refresh was the system working as designed.

The fix is a `ConversationCleared` observation notified from `handleReset`,
rendered into the frame `renderEvent` already produces, so both channels drive
the client down one path. The sealed observation union turned the missing
`isObservation` into a compile error rather than a silent no-op.

**For the book:** this is a sibling of section 17's lesson, inverted. There the
live and replay paths both existed and disagreed. Here the replay path was
correct and the live path did not exist at all, which is harder to see, because
every test of the durable behavior passes. The question to ask of any event is
not "is it recorded" but "which of the two channels carries it, and who is
listening to the other one."

A note on scope that is itself a design argument: the event cap in section 19
deliberately does not window replay. Had it done so, the reset frame could fall
outside the window, and the refresh that currently cures this bug would stop
curing it.

## 19. A view limit is not a retention limit

The GUI rendered every event it had ever seen. The new Appearance setting caps
that at 100, discarding oldest first.

The useful distinction is that three different limits want to live here and
only two of them are the same question:

- **`MaxEvents`** bounds rendering work. It is a view limit. It loses nothing,
  because the log still holds every event and a refresh replays them all.
- **`LogRetention`** bounds the file on disk, and already existed. It trims on a
  clean save, keeping the newest events *after* the snapshot is written, so the
  snapshot always exists before the old events are gone.
- The **context window** is a third thing entirely, and the ladder owns it.

Conflating the first two is the tempting error, and it would be lossy: a view
preference would start deleting history. Keeping them separate costs one honest
consequence worth stating in the GUI rather than hiding, which is that after a
restart the screen can show no more than `min(MaxEvents, LogRetention)`.

Unset resolves to the default on the *server*, so the broadcast always carries a
concrete number and the client keeps no second copy of the default to drift
from. Zero is not honored literally here, unlike `temperature` or
`tts_enabled`, because a screen showing no events is not a choice anyone wants;
effectively unlimited is spelled as the cap.

**The bug that was avoided:** the trim prunes the three correlation maps along
with the DOM, for the same reason `clear()` empties them. An entry outliving its
element routes the next delta into a node no longer on the page, so text
vanishes with no error anywhere. The hazard is sharper when trimming than when
clearing, because trimming happens *while a turn is streaming*. Discarding only
the oldest is what makes it safe: a part still streaming is the newest, so it is
never the one dropped.

## 20. The model selector that selected nothing

The GUI offered a model dropdown. Choosing a different model changed nothing at
all: every request went to the startup default.

The dropdown wrote `settings.Model`. `cmd/main.go` set `cfg.Model` from
environment variables. Nothing carried one to the other. It was a known-dead
setting, and not the only inert one in that struct.

### The readout confirmed the lie

What makes this worth a section is not the dead wire but the instrument.
`effectiveModel` consulted the settings store *first*, with a comment explaining
that switching the model in the GUI is what the operator just did. So the GUI
displayed the selected model while the engine dialed the old one, and the
display was the thing that was wrong.

A readout fed from *intent* always agrees with the operator. Only a readout fed
from the *fact* can disagree, and disagreeing is the entire job. It now reports
`Cfg.Model`, the model actually dialed.

### Switching needs no rebuild, and that is a finding

The intuition is that changing models means rebuilding the client. Here it does
not, and the reason is a design property worth naming: the renderer and
`curate()` both resolve features through `LookupModel` on every call rather than
caching them at construction. Derive-per-use costs a map lookup and buys live
reconfiguration for free.

Two things do not follow automatically, and both would fail quietly:

1. **Vendor and Surface** are separate config fields. Move the model alone and
   the next request is rendered in one vendor's dialect and posted to another
   vendor's endpoint.
2. **The model-gated tool set.** `keep_tool_results` is withheld on models that
   do not stub tool results. That gate was resolved once at startup, which was
   correct right up until the operator switched. One function now decides, for
   both startup and switch, so they cannot drift.

Restoring a withheld builtin re-registers it with `Initial` provenance, because
that is what it is. The ordinary `RegisterTool` stamps `Dynamic` and would have
quietly falsified the provenance record.

### Vendor is a property of the model

The vendor was derived two ways: the engine dialed whatever `Cfg.Vendor` said,
while the GUI labelled models by string-prefixing the ID. Two mappings that
could disagree, and neither able to answer the question a switch actually asks,
which is *which API do we dial for this model*.

It now lives in the model table, with the display label derived from the same
field. The zero value is deliberately invalid rather than defaulting to a
vendor, so a row that forgets to declare one is detectable. A test asserts every
row declares a vendor, and it earned its keep within a minute: it caught five
rows a grep had reported clean, because `gofmt` aligns the map and the pattern
required exactly one space after the colon.

`fake-model` is driven against several vendors by environment prefix and so
genuinely has no single answer. It leaves the field unset and `VendorFor`
refuses, which callers must treat as a refusal. An unknown model is refused
rather than guessed at: a prefix table would cheerfully resolve
`claude-not-a-real-model` to Anthropic and dial the wrong API.

**For the book:** the switch travels as a mailbox message, for the same reason
`Reset` does. The actor owns the config, so applying a switch from the WebSocket
goroutine could land midway through rendering a request. Applied between turns,
"switch at any time" becomes safe including mid-turn, when it simply takes
effect at the end of that turn. The question "must we rebuild on a model
switch?" turns out to be a question about where features are resolved, and the
answer was already written into the architecture.

## 21. The breakpoint that paid for a prefix twice

Four cache breakpoints is all either vendor allows, so the question of where
they go is a question about which four boundaries in a request are worth
money. Reading the renderer to check the answer turned up one slot spent on a
prefix that was already paid for, and one boundary left unmarked.

The rule they settled into, in cache order:

1. the end of the tool array,
2. the system prompt,
3. the compaction bound, meaning the end of the newest micro_handoff,
4. the most recent message carrying no ephemeral part.

There is deliberately none at the start of a turn.

### Why the start of a turn is the wrong place

The removed marker sat at the end of the previous exchange, one message behind
the newest prompt. The argument for it was that it reads a prefix an earlier
request already paid to write, and that it stays put through a long tool loop
while the rolling marker advances.

Both halves are true and neither pays. The rolling marker from the *previous
round* already wrote an entry at that very prefix, and both vendors look
backward past a bounded number of positions to find a prior write. The slot was
buying a second copy of something already in hand.

Bill's framing is the sharper one, and it turns a tuning question into a
correctness question: a start-of-turn breakpoint should never be *needed*, and
if it ever were, that is a bug to go fix rather than a cost to cache around.
The only way such a marker earns its slot is if the prefix below the newest
prompt changes between rounds. That is prefix instability, and the cure is to
stop rewriting the prefix.

Ensemble cannot drift that way, for a structural reason: `land()`
strips every tool call and tool result at reduce time, once, and the result is
frozen into the projection. CodeRhapsody, by contrast, keeps the three most
recent tool pairs across a handoff and counts them back from a boundary that
moves, which is exactly the instability that would tempt someone into pinning
the start of a turn. Same rule, different answer, because the two agents
compact differently.

### What the freed slot bought

Tools sit at the very front of the cache order, ahead of the system prompt and
the messages. A marker there is the only one that survives an edit to the
system prompt: without it, changing a single word of the prompt discards the
tool declarations too, and tool schemas are not small.

It goes on the *last* declaration, never an earlier one. A breakpoint includes
every byte before it, so marking the last tool closes a prefix containing all
of them; marking any earlier one leaves the remainder outside.

### The vendors disagree, and only one of them can be obeyed

Anthropic takes `cache_control` on a tool definition, so it gets all four.

OpenAI cannot. Its breakpoints attach to message *content blocks*, and Chat
Completions supports them on `text`, `image_url`, `input_audio` and `file`.
The tools array is not a content block, so the first breakpoint is not
expressible there at all. Nothing is lost except independence: a breakpoint
includes all prompt content before it, and tools precede the system prompt, so
OpenAI's system marker already covers the tools. It simply cannot keep them
when the prompt changes. OpenAI therefore writes three of its permitted four.

Three further details of OpenAI's scheme are worth having on the record,
because they bear on whether the budget is really full:

- **Earlier turns' breakpoints are read-only.** They can match the cache but
  are not written again, so the compaction bound costs a write slot once and is
  free to read thereafter.
- **Reads consider the latest fifty breakpoints**, far more than are written.
- **On GPT-5.6 and later, cache writes can be charged.** The usage meter counts
  tokens and applies prices at render time, so this is a pricing-table question
  rather than a metering one, but it is no longer safe to assume a write is
  free.

### A test that asserts position, not text

The guard against the start-of-turn marker coming back checks *where* markers
are, not what they sit on. The first version compared block text and would have
passed a restored anchor straight through, because a marker can land on a block
whose wire text is empty, a tool result being the obvious case. The rewritten
version asserts that in an uncompacted conversation no marker appears before
the final message.

Restoring the anchor as a mutant is caught by five tests. Dropping the tool
marker, and moving it from the last declaration to the first, are each caught
by one.

## 22. An agent that could not see its own GUI

Bill's observation started it: a student who passes every chapter ends up
where we were on the morning of the switch. The agent has a GUI, the GUI has an
MCP server built into it (`gui_snapshot`, `gui_click`, `gui_input` and
more), and
nothing can use it. The agent cannot look at its own screen, and no outside
agent can connect to fix it. "We'll need you to see the Ensemble GUI and fix
Ensemble until Ensemble can do it itself."

**What was actually wired.** Less than the docs said. The gui-debug skill told
the model it could inspect and drive the GUI, and that has been false in the
shipping binary since chapter 12: `NewMCPWSTransport` had no caller. The
chapter 13 grader never noticed, because it writes a temporary SKILL.md that
launches a stdio fake MCP server, so it tested the MCP client and never the
path from the agent to the browser.

**Change.** Two doors, one per direction.

- `--mcp-port N` (1d2cbfa) relays the GUI's MCP server onto a loopback TCP
  port, one JSON-RPC message per line, with each connection tagged as its own
  source so replies route back to the right caller. `cmd/mcp-connect` is a
  stdio-to-TCP pipe, so any MCP client that spawns a stdio server, including
  CodeRhapsody, can attach. `--mcp-port` without `--port` is refused, since
  there is no GUI to relay.
- `view_gui` (8e0c60b) is a builtin tool that asks the connected GUI for a
  snapshot over the existing WebSocket (`Hub.CallGUI`, sharing a reply router
  with the relay). It reports a missing GUI or a timeout as errors, not as an
  empty screen. The gui-debug SKILL.md was rewritten to claim only what is
  true: the agent can look; it cannot click or type from inside.

**Verified live.** Headless Chrome on the GUI, the agent on the ChatGPT plan
route with gpt-6.1-sol, prompt "Look at your own GUI and list the buttons you
can see, with their selectors." It called `view_gui` and answered with
`#hamburger` (expanded), `#reset-btn`, `#settings-btn` and the Chats/Artifacts
tabs, which is what the snapshot contained. Getting there took the fixes in
entries 23 and 24.

**Not built.** Clicking and typing from inside. `view_gui` only looks; an
external agent on `--mcp-port` can drive the GUI with `gui_click` and
`gui_input`.

**Still open.** The snapshot showed an artifact reading `{}{}{}` for a tool
call whose arguments were `{}`, which suggests argument deltas are rendered
once per delta plus once at the end. Cosmetic, not yet investigated.

## 23. The underscore that hid a tool

`view_gui` was registered, unit-tested, and never reached a request.

The registry keys tools by `common.NormalizeName(name)`, which strips every
non-alphanumeric, so `view_gui` is stored under `viewgui`. `Declarations()`
then asked the skill registry `IsToolEnabled(key)`. The skill lists
`view_gui`, the question was about `viewgui`, and the answer was no. Every
non-builtin tool with an underscore in its name was silently filtered out of
every request, whatever the loaded skills said. Builtins survived for a
structural reason, measured on the pre-fix tree: `NewRegistry` stores them
under their raw name, while `Register`, `RegisterInitial` and dynamic
registration store the normalized one. One map, two key conventions; for
builtins the key happens to be the name, so asking about the key worked.

**Change** (e171c42). Ask about `r.tools[n].Name`, the name a skill author
writes. `TestSkillFilterUsesTheDeclaredName` fails on the old code.

**The fix broke something the chapter could not see.** The MCP connect path in
`load_skill` records each MCP tool as enabled, and it recorded the normalized
key, under a comment that said why: "IsToolEnabled compares against normalized
map keys." The old convention was consistent; e171c42 moved one end of it. After
that, `gui_click` from a skill-connected MCP server no longer reached the dialog.
ch19 stayed at 100. The cross-chapter sweep caught it: ch15 `frozen-prefix`
fell to 90 and ch16 to 97 through its ch15 parity check. Bisected in a worktree
(3bfb494 ch15 100, e171c42 ch15 90), fixed in f35c94e by recording the declared
name. `IsToolEnabled` has two callers, and every producer of enabled names now
supplies the declared name.

**Still open.** The registry map still has two key conventions (raw for
builtins, normalized for everything else). The filter no longer mixes them, but
the map invites the next comparison of a key to a name.

**For the book:** the tests passed because they tested the tool, not the path
a tool takes from registration through the skill filter into a request. The
live run caught in one prompt what a green suite had hidden. A normalized key
is an index, not an identity; any question about identity has to be asked in
the name the other party used.

## 24. The plan route is not the documented route

Chapter 19 was built from OpenAI's documentation and scored 100/100 against a
fake built from the same documentation. On the real ChatGPT-plan route of
`POST /v1/responses`, every turn failed. The route differs from the metered
(API-key) route in four ways the docs do not mention. They were found live,
one at a time, because each masked the next:

1. **No explicit caching.** `prompt_cache_options` and a bare
   `prompt_cache_breakpoint` both return 400 "not supported on this model".
   The message blames the model; the metered route accepts both for the same
   model. Implicit caching is absent too: an identical repeat reported zero
   cached tokens. Fixed in 3bfb494 by adding both fields to the plan route's
   forbidden list, which the renderer consults before attaching breakpoints.
2. **No Content-Type header on the stream.** The metered route sends
   `text/event-stream; charset=utf-8`. `isSSE` trusted the header alone and
   handed the stream to the JSON decoder: "invalid character 'e'". Now the
   header decides when present, and when absent the first non-blank byte does
   (`{` means JSON).
3. **`response.completed` carries `output: []`.** Items arrive only as
   `response.output_item.done` events. The parser read parts only from
   `completed.output`, so the model's text and tool calls were dropped and the
   turn ended successfully with nothing in it. Now done items are collected by
   `output_index` and used when `completed.output` is empty.
4. **A consequence of 3, and a bug on both routes.** Once tool calls survived,
   the turn ended in "refusing to marshal invalid vendor 0". `respAssemble`
   built `ToolCallPart` without its `From` provenance. The journal's refusal
   was correct and stayed; the producer was fixed. The metered fixture test
   now asserts the provenance, so this one was always catchable.

Fixes 2–4 are ec5f086. Each has a test that fails when the fix is removed.

**A misdiagnosis worth recording.** The first 400 looked transient, because
replaying the exact request body from Python succeeded. The replay used the
API key; Ensemble used the plan token. The difference was the credential, not
the body. Replay with the credential the agent used, or the experiment
measures a different route.

**A second finding from the same probes.** The plan token gets 401 on
`/v1/chat/completions`. Bill's instance had been answering on gpt-5.6-sol,
which is on the Chat Completions surface, so that traffic cannot have been on
his plan. Reading `attachCredentials` in `cmd/auth.go` explains it: plan
credentials attach only when the vendor is OpenAI at startup, and the default
`LLM_VENDOR` is anthropic, so a later switch to an OpenAI model ran on the API
key from settings.json. That is inferred from the code and the 401, not
observed on his bill.

**Still open.** Only gpt-6.1-sol (and the course model) are on the Responses
surface; Bill wants all OpenAI traffic there, per the chapter 19 ruling. The
cache lens reported "PREFIX DIVERGED" in `input` on the second round of a
plan-route turn although `diff prior_request.json request.json` shows a pure
append. Probably the lens's grown-array rule not recognising the Responses
`input` shape; not yet checked.

## 25. A fake built from the docs agrees with code built from the docs

The ch19 fake Responses server sent the documented stream shape to everyone,
so it could not see any of entry 24. The grader and the agent shared one
belief about the wire, and agreement between them proved only that.

**Change** (d68fef8). The fake picks the route from the credential, as the
vendor does: any bearer other than the metered API key gets the plan route as
measured, with no Content-Type and an empty `completed.output`. One Go trap:
leaving the header unset is not the same as sending none, because net/http
sniffs and sends `text/plain; charset=utf-8`. Setting
`w.Header()["Content-Type"] = nil` suppresses it, and a test fails without
that line.

`responses-format` gains the converse of its old rule "claim success only on
`response.completed`": a turn the vendor completed must end without an error,
and the scripted tool call must come back as `function_call` plus
`function_call_output` in a later request. Before this, an agent that failed
every plan turn still passed it.

**Verified.** `./agent` 100/100. The frozen `solutions/ch19` fell from 90 to
60 and names the cause ("turn 1 ended in error although the vendor completed
it"). Three new audit mutants, one per defect, each kill `responses-format`
(the Content-Type mutant also kills `reasoning-summaries`, since an unparsed
stream delivers no summaries). The audit is 11/11.

**Rejected.** A `PlanRoute` option on the fake. The route is a property of the
credential, and a flag the test sets would let a scenario claim one route
while sending the other's token.

## 26. Every answer drawn twice and spoken twice

**Symptom.** Live on the plan route, the GUI showed each assistant answer as
two artifacts, and a snapshot listed a `{}{}{}{}` artifact that never
resolved. The JSON-lines log showed why: deltas and finals for the same part
carried different ids (tool call: deltas `1`, final `2`; text: deltas `3`,
final `4`).

**Cause.** The stream parser keyed delta ids by item (`text:<item>`,
`call:<item>`, `sum:<item>:<summary_index>`), but finalized through a
separate mapper keyed by part index, so every final allocated a fresh id.
Observers key parts by id. The GUI draws a final under an unknown id as a new
element, and because that id has no streamed text, the "arrived whole" rule
in `_handleFinal` speaks it. For a listener every Responses answer was read
aloud twice.

**Change** (eb15005). `respAssemble` returns, parallel to the parts, a key per
part in the delta path's naming, and the stream path finalizes under
`live[key]`. A part that never streamed gets a fresh id, as the whole-document
path would give it. A reasoning item finalizes under its first summary's id;
its later summary parts stay complete as streamed, since `summary_index` part
boundaries are a graded property.

**Second defect, same file.** `partFor` called `cb.AllocPartID()` directly.
The recall judge parses with no callbacks set, so the first recall on the
Responses route dereferenced nil and killed the process (seen live, on the
first prompt that matched memory snippets). It now allocates through the
mapper, which already had the nil fallback for these callers. The Chat
Completions and Gemini parsers nil-check already.

**Third, GUI** (37ee29d). Encrypted reasoning that streamed no summary
finalizes under an id with no element, and `_handleFinal` created an empty
artifact for it each turn. An opaque final with no element now draws nothing.

**Verified.** Two tests on the verbatim live fixture: finals reuse delta ids
(fails on all three kinds without the fix) and a zero-callbacks parse (panics
without it). Live: delta and final ids match for tool call, text and
reasoning; a recall-triggering prompt completes. ch7, ch14, ch19 100/100.

**Measured, not yet applied.** gpt-5.6-sol and gpt-6-astra on the plan route
with the Responses surface: both accept thinking together with tools (so
`NoThinkingWithTools` is a Chat Completions fact only, as the vendor's own 400
says), both stream reasoning summaries, both complete a tool round-trip. The
table rows were not flipped because they double as fixtures: gpt-6-astra is
chapter 18's grader model against a Chat Completions fake, and gpt-5.6-sol is
the fixture for the `NoThinkingWithTools` and Chat Completions cache tests.
Flipping them first needs those tests moved to course rows.

---

## 27. A terminal that could not reproduce a single GUI bug

Every bug in the three entries above was found in the GUI, and every one of
them had to be found in the GUI, because the terminal ran different code.

`ensemble chat` called `runLoop`, which drove a bare engine: no actor, no
skills, no auto-recall, no save file. The default mode called `runActorLoop`,
which drove all of it. Two dispatchers, and only the one behind a browser
exercised the real agent. So the cheapest front end could not reproduce
anything the expensive one reported, and a GUI bug report could not be
narrowed to the GUI without opening a browser and clicking.

`chat` now calls `runActorLoop` with a text front end attached. The CLI and
the GUI are two Observers of one agent. That makes the diagnosis mechanical:
a bug that reproduces in both is in the agent, a bug that reproduces in only
one is in that front end. Passing `--port` as well runs both against a single
agent, which is the rig the next three entries were debugged on.

The front end is an Observer and nothing else. It never calls back into the
agent, which is what lets it be swapped for a browser without the agent
noticing, and it is why adding it required no change to the engine.

**The terminal takes prose, not protocol.** The JSON-lines protocol is
unchanged and still available to programs. A human types a sentence. Slash
commands carry only the two things prose cannot express, both of which exist
because turns are asynchronous: `/hint` steers a turn that is already running,
`/interrupt` stops it.

**Reply text goes to stdout, commentary to stderr.** This rule was inherited
from the deleted `chatStream`, and it was nearly lost in the rewrite: the
first version printed everything to stdout. Piping the binary yields exactly
what the assistant said and nothing else, which is what makes the CLI usable
in a shell pipeline. Usage counts moved to stderr for the same reason, after
a measured run put a JSON line in the middle of the reply text.

**The duplicate-draw bug was waiting here too.** The observer prints a final
part only for a part that never streamed. The Responses surface reuses delta
ids for finals, so a renderer that draws both shows every answer twice, which
is entry 26 in a different front end. Writing a second renderer against the
same seam is the cheapest test of whether a seam is honest.

`runLoop` and `chatStream` are deleted. One dispatcher remains.

**Measured live on the plan route**, gpt-6.1-sol: a prompt that writes a file
streams the reply, prints `-> write_file` and `<- ok wrote 4 bytes`, creates
the file, and leaves stdout holding only the sentence the assistant wrote.

---

## 28. A picker that showed the wrong model

Two settings are changed far more often than the rest, and both were behind
the settings panel. The top bar had a `#model-name` span that nothing ever
populated, so the bar did not even report which model was running.

The bar now carries a model picker and a speech toggle. Neither adds state.
Both register in `settingFields` like every other control, and the server
broadcasts the entire settings object after any update, so the panel and the
bar re-sync from one source and cannot drift apart. The picker is filled from
the server-sent catalog rather than a hardcoded list, so it cannot offer a
model the agent does not have.

Putting the picker where it is always visible exposed a bug that the hidden
panel had made harmless. An unset model means "follow the startup default".
That is the right thing to store, and the wrong thing to show. `applySettings`
skipped an empty model, so the select kept its first option and displayed
`Claude Opus 4.6` while the agent was running `gpt-6.1-sol`.

The fix did not need inventing. `effectiveModel` already existed for the usage
meter, and its comment already records this failure:

> the meter cheerfully confirmed a model that every request ignored, which is
> a worse failure than showing nothing at all.

A wrong model in a picker is worse again, because a picker looks like a
control rather than a readout. `settingsForDisplay` fills an empty model from
`effectiveModel` before either settings message leaves the hub.

**It fills an empty model only, and that restriction is the interesting half.**
A model switch is an actor mailbox message, applied asynchronously, so for a
moment after the operator picks a model the live model is still the old one.
Filling unconditionally would broadcast that stale value and snap the picker
back to the previous model immediately after the change. Both halves are
pinned by tests, and each was killed by its own mutant: removing the fill
kills the first, making it unconditional kills the second.

**The same bug survives a restart, in the opposite direction.** Startup wires
`ContextTarget`, `Bands` and `LogRetention` from the settings store, and never
the model: the engine takes its model from the flag or the environment. So a
model chosen in the GUI is written to `settings.json`, ignored by the engine
on the next start, and then displayed to the next client that connects. The
store would have been believed over the wire.

A freshly connected client has nothing pending, so `settingsAtLoad` shows the
model in force unconditionally, while `settingsForDisplay` protects the
operator's pending choice during a broadcast. Two rules, because a fresh load
and the instant after a change are genuinely different moments. Three tests,
three mutants, each killing exactly one.

**On sizing.** The controls are not shrunk to fit the bar. The person most
likely to reach for a speech toggle is the person least able to hit a small
target, and a control that cannot be seen is not really there.

The speech toggle drives the existing `tts_enabled` setting rather than a new
accessibility flag. Speech is the accessibility feature this codebase has, and
a second flag that only turned on speech would be two names for one thing.

**Verified live** through the GUI's own MCP relay, with the agent watching its
own interface: the toggle click round-trips to the server and persists to
`settings.json`, and the picker renders the catalog grouped by vendor.

**The speech path was verified against the chapter 14 contract**, not merely
switched on. With the toggle enabled, one turn that reads a file produced
exactly two utterances:

```
seq 1  read file: /tmp/ens-gui/hello.txt, line 1
seq 2  The file contains the single word "pong."
```

The tool's INTENT is spoken and its RESULT is not, while the same word spoken
as assistant prose is. That is the rule chapter 14 exists to enforce: a string
must be silent as a tool result and audible as prose. A toggle that turned
speech on while breaking that rule would be worse than no toggle, because the
reader who depends on speech is the reader least able to see that the output
has gone wrong.

---

## 29. The bug the picker armed

Entry 28 added a model picker to the top bar. Entry 29 is about what that
picker made reachable, which was already there and had been for some time.

Credentials for the ChatGPT plan were attached once, at startup, by a host
function that opened with this:

```go
if cfg.Vendor != common.VendorOpenAI {
    return
}
```

Read on its own that is defensible. The provider is an OpenAI credential, so
installing it while the agent is pointed at Anthropic would present a ChatGPT
bearer token to the wrong host. The guard is doing real work.

The trouble is *when* it runs. `attachCredentials` is called exactly once, from
the composition root. The vendor it consults is the vendor in force at startup.
And since chapter 11 the model has been changeable at runtime, which means the
vendor has been changeable at runtime, which means the question this guard
answers had an expiry date on it.

**What the switch already knew.** The runtime switch is not naive. It had been
fixed once before, and it carries a comment explaining why:

> The endpoint moves with the model. The renderer follows Vendor, so an
> endpoint left behind renders one vendor's dialect and posts it to another
> vendor's host, carrying that vendor's key. Resolve and refuse before mutating
> anything: a half applied switch is worse than none.

So the switch moves `Model`, `Vendor`, `Surface`, `BaseURL` and `APIKey`. Five
fields, carefully, with a refusal path. It moves everything the earlier bug
taught it to move, and the credential provider was not on that list, because
when that fix was written the provider was not yet a thing that could be wrong.

The result is a switch that is scrupulous about *where* the request goes and
silent about *who pays for it*. Flip the picker from Claude to GPT and the
dialect changes, the host changes, the key is refilled from the endpoint table,
and the plan credential never arrives. The request is rendered correctly, sent
correctly, answered correctly, and billed to the metered API key.

**This is the failure mode worth naming.** Nothing breaks. There is no error, no
retry, no degraded reply. The only component that disagrees is the invoice, and
it disagrees a month later. A bug that announces itself is a bug you fix on
Tuesday; this class of bug is found by someone reading a statement and saying
*that number is wrong*.

**The fix is a lookup, not a guard.** The guard was in the right place
logically and the wrong place temporally, so it moved from attachment time to
use time. The host still does discovery, because reading the operator's home
directory is host policy and a library that went looking for credentials on its
own would be making that decision for every embedder. But it now registers the
provider *against the vendor it belongs to* instead of installing it only when
that vendor happens to be first:

```go
eng.CredsByVendor[common.VendorOpenAI] = p
if cfg.Vendor == common.VendorOpenAI {
    eng.Creds = p
}
```

and the switch gained one line beside the endpoint it already moved:

```go
a.eng.Creds = a.eng.CredsByVendor[v]
```

A vendor absent from the map means "no provider", so a switch to Anthropic
correctly reverts to its metered key rather than inheriting the previous
vendor's bearer token. Absence is modelled, not defaulted.

This is the same shape endpoints already use. Endpoints are resolved for every
vendor up front so that a switch is a lookup rather than a rediscovery. Doing
the same for credentials is not a new idea in this codebase, it is the
application of an existing one to the field that got missed.

**On testing the observable.** The test asserts the bearer token actually
presented and the route actually selected, after switching to OpenAI and again
after switching back. It deliberately does not assert that `eng.Creds` is
non-nil. A provider that is attached but never consulted would satisfy that
weaker check and still bill the wrong account, and this codebase has already
shipped one bug of exactly that shape: a test that checked `Vendor` while the
request still went to the previous host. The proxy passed; the behavior did
not. Assert the thing the user is charged for.

Deleting the single carry line fails the test with the metered key in the
message, which is the only evidence that the test is load-bearing rather than
decorative.

**The general point for a crossover chapter.** The picker did not introduce
this defect. It removed the last thing protecting us from it, which was the
inconvenience of changing vendors. A feature that makes an action easy is also
a feature that makes every latent bug on that path reachable, and the honest
way to ship one is to go looking for what it just armed.

---

## 30. Moving two models to Responses, and the mark it cost

A table row is a claim about the world. Two of ours had stopped being true.

`gpt-6-astra` and `gpt-5.6-sol` were both marked Chat Completions. Neither was
there because anyone had measured them there and concluded it was right. They
were there because nine test files and a grader named them, and a row that
nine tests depend on is a row nobody edits casually.

Both had been run against the live ChatGPT-plan route carrying thinking, tools
and streamed reasoning summaries. The table said otherwise, and the table is
what the code reads, so `cfg.Surface` came from `SurfaceForModel` and there is
no runtime override. Using either model on the plan route was not a setting. It
required changing the row.

**The ruling, and the part worth keeping.** The first instinct was to delete
Chat Completions outright: it is the old API, OpenAI is steering everyone to
Responses, why teach it. The reason not to is that Chat Completions is no
longer really OpenAI's API. It is the de-facto dialect of everything that
speaks "OpenAI-compatible" — Ollama, vLLM, LM Studio, and most of the local
inference ecosystem. Deleting it would cost a reader the ability to point this
agent at a model on their own machine, which has nothing to do with OpenAI's
roadmap. So: keep the surface, move the models. New models use Responses.

For the same reason `DefaultSurface(VendorOpenAI)` stays Chat Completions. A
model earns the new surface by naming it, which is what keeps the chapter 19
claim — surface is per-model, not per-vendor — a fact rather than a slogan.

**The best argument for the migration came from the vendor.** The old row
carried a flag, `NoThinkingWithTools`, pinning a live-only failure:

> Function tools with reasoning_effort are not supported for gpt-5.6-sol in
> /v1/chat/completions. To use function tools, use /v1/responses or set
> reasoning_effort to 'none'.

An agent always has tools, so on Chat Completions every turn failed. OpenAI's
own error text names `/v1/responses` as the remedy. The migration is not a
preference dressed up as engineering; it is the documented fix.

That flag could not simply be deleted along with the row, though. The
restriction is real for Chat Completions, and deleting the row's flag would
have quietly discarded the regression test for a bug that broke everything. It
moved to a `gpt-5.6-course` fixture that stays on the old surface. The
production row records today's measurement; the fixture holds yesterday's
scar.

**A guard fired, correctly.** Adding `StreamReasoningSummary` to two rows broke
a test called `TestOnlyMeasuredRowsClaimReasoningSummary`, whose comment reads:
*The table records measurements. A row claims summary streaming only if someone
watched it stream a summary.* The honest repair was to extend the measured set,
because we had watched exactly that — not to loosen the assertion. A guard that
fails when you change the thing it guards is not an obstacle. It is the only
part of the change that was free.

**Now the mistake.** The ch18 grader named `gpt-6-astra`, so freeing that row
meant pointing the grader somewhere else. A new `gpt-ch18-course` row was added
and the grader aimed at it. `make grade18` returned 100/100. Everything built,
every unit test passed, both modules were clean.

The cross-chapter sweep then reported this:

```
ch18 ./agent            score: 100/100
ch18 ./solutions/ch18   score:  85/100
```

`solutions/ch18` is a frozen snapshot with its own copy of the model table, and
that copy has never heard of a row added today. The grader asked a frozen
student tree for a model it did not have, and it lost fifteen marks without
erroring.

Two things about that are worth more than the fix. First, **the per-chapter
grader passed.** It grades the live tree, which did have the row, so the one
command most likely to be run after touching chapter 18 returned full marks on
a change that broke chapter 18. Second, **nothing failed.** No exception, no
missing-model error; just a lower number in a column, in a log, thirteen
minutes later.

The fix was to stop inventing a row and use `gpt-5-course`, which already
existed in both trees with an identical profile for chapter 18's purposes —
Chat Completions, implicit caching, same minimum and same window. One new
fixture instead of two, and the grader names something its frozen snapshots
actually contain.

**The rule this leaves behind.** A grader model must exist in every frozen
snapshot that grader grades. Frozen trees cannot be given new rows, so a grader
may only name models that were already there. The compiler cannot check this,
the per-chapter grader does not notice it, and the failure is a silent score
drop. The sweep is not belt-and-braces; for this class of change it is the only
instrument that works.

Which also settles the sequencing argument. Fixtures move first, the suite is
proven green, and only then does the production row change. Doing it the other
way round means every failure afterwards has two candidate causes, and the
sweep takes thirteen minutes per guess.

**Verified live, which is the only verification that counted here.** With
`LLM_MODEL=gpt-5.6-sol` on the ChatGPT-plan route, one CLI turn asked for a
file to be written and described back. It exited 0, the tool ran, the file
contained `pong`, and stdout held exactly one sentence. Tools and a reasoning
budget in the same request — the pair that drew the 400 quoted above — now
work, because the request goes to `/v1/responses`.

This is worth stating plainly because the fake vendor could never have
established it. A fake accepts whatever we send, so the old request was well
formed by our lights and refused by theirs. The grader scored 100/100 against
Chat Completions for as long as that row was wrong. Only the live wire settles
a vendor question; everything else only settles whether we are consistent with
ourselves.

The usual plan-route caveats still show in the lens: zero breakpoints and zero
cache reads, because that route declines explicit caching, and a spurious
PREFIX DIVERGED line that means nothing where there is no cache to diverge
from. Both are known and neither is new.

---

## 31. Why ./agent no longer matches solutions/ch19

A note for whoever diffs those two trees next and wonders which one is broken.

Neither is. `solutions/ch19` is a frozen snapshot of what chapter 19 teaches,
and it is correct as of the end of that chapter. `./agent` is the live tree and
has moved on. As of this entry nineteen tracked files differ, and every one of
them is chapter 20 material:

| commit    | files | what it was |
|-----------|-------|-------------|
| `cd61f8b` | `cmd/main.go`, `cmd/textui.go` | `chat` rebuilt as a text Observer over the actor loop |
| `3ab0f51` | `web/gui/gui.js`, `index.html`, `style.css` | model picker and speech toggle in the top bar |
| `2061b0b` | `internal/ws/handler.go`, `settings_display_test.go` | show a fresh client the model actually in force |
| `29faa0e` | `cmd/auth.go`, `internal/llm/actor.go`, `engine.go` | carry the credential provider across a vendor switch |
| `828e236` | `internal/common/model.go`, `reasoning_summary_test.go`, and six `internal/llm` test files | move two models to the Responses surface |
| `f107a6e` | `internal/ws/replay_test.go` | gofmt |

The snapshot was deliberately **not** refreshed. Freezing this work into
chapter 19 would misdate it: chapter 19 is about leaving one vendor for
another, and none of the above is that. The CLI rebuild, the top bar, the
billing fix and the surface migration are all consequences of the crossover,
which is the next chapter's subject.

**The rule being followed.** A snapshot is refreshed when the work belongs to
the chapter it snapshots, and only then. Earlier sessions did refresh ch19
after plan-route fixes, correctly, because those fixes *were* chapter 19
material. "Is it committed?" and "does it belong to this chapter?" are
different questions, and only the second decides whether a snapshot moves.

**What the drift is good for.** It is a free, continuously maintained diff of
exactly what the next chapter has to explain. Run it and the chapter's scope
prints itself:

```
for f in $(git ls-files agent | sed 's|^agent/||'); do
  if [ -f "solutions/ch19/$f" ]; then
    cmp -s "agent/$f" "solutions/ch19/$f" || echo "DIFF $f"
  else
    echo "NEW  $f"
  fi
done | sort
```

A caution that already bit once. The count is a moving measurement, not a fact
to quote later: it was seven files at the start of the session that produced
entries 29 and 30, and nineteen by the end. Any statement of the form "the N
files" goes stale the moment the next commit lands, so re-run the loop rather
than trusting a number written down — including the nineteen above.

## 32. Logging which credential is in use

There was no way to tell, from a running session, whether a turn was paid for by
a ChatGPT plan token or by a metered API key. The configuration said which one
*should* be used. Nothing said which one *was*.

That gap is how the bug in entry 29 survived. A model switch re-derived the
route and the endpoint but left the credential behind, so the agent spoke the
plan dialect while billing the metered key. Every request succeeded. The only
artifact that disagreed was the invoice, and an invoice arrives weeks after the
session it describes.

So the engine now records the credential in force when it builds a request:

```
[03:50:59.861] credential: vendor=anthropic using api_key (static, from configuration)
[04:28:11.204] credential: vendor=openai using chatgpt_oauth
```

Three decisions are worth keeping.

**It logs on change, not every turn.** A line printed every turn is a line
nobody reads. The fault worth seeing is a credential that changes when nothing
asked it to, and that is visible only against a remembered previous value.

**The redaction lives in one function.** `DescribeCredential` names the kind and
has no path that can emit a token. Putting the rule at each log site instead
would mean a rule enforced in five places and broken in the sixth: a bearer
token leaks through exactly the one line nobody reviewed.

**A missing provider reports as the static key, not as "none" or blank.**
Falling back to a billed key while expecting a plan credential is the commonest
mistake here. It must not read like the correct case, and "none" would.

The line earned its place within a minute of existing. Running with
`LLM_MODEL=gpt-5.6-sol`, it printed `vendor=anthropic using api_key` -- an
OpenAI model id being sent to Anthropic. `cmd/main.go` resolves the vendor
first and then uses `LLM_MODEL` to pick a model *within* that vendor, so naming
a model from another vendor does not move the vendor with it. The request
failed silently and the session simply looked quiet. The correct invocation
sets both: `LLM_VENDOR=openai LLM_MODEL=gpt-5.6-sol`.

## 33. A model that will not say what it is doing

Some models reason at length internally and then act in silence. For an agent
that edits files, that is not a style difference. It is unusable.

The supervisor here reads narration as it streams, at around 750 words per
minute, through a screen reader. The narration is not commentary on the work;
it *is* the review, and it happens live. When the agent explains each edit as it
makes it, the supervisor follows the change as it happens, stops a wrong
assumption before it becomes ten wrong edits, and by the end already understands
the change well enough not to review it at all.

When the agent works silently, none of that exists. There is no way to tell what
was touched or what broke. The only safe move left is to throw the whole run
away, correct work included. That is not hypothetical: it happened, and the
reset took the supervisor's own uncommitted work with it.

It left a second mark worth recording. `git reset --hard` does not remove
untracked files, so three half-finished files the model had created in silence
survived the reset -- and they did not compile. The tree was broken by work that
the reset was supposed to have erased, from a session nobody could see.

### The capability, not the instruction

The flag is a fact about the model, on `ModelFeatures`:

```go
RequiresVisibleReasoning bool
```

It belongs there for the same reason every other row in that table does: it
records something measured about a specific model. A model that narrates
willingly pays nothing. An unknown model is not enforced against either --
enforcing on a zero-value row would turn "I have never heard of this model" into
a behaviour change, which is how a table lookup quietly becomes policy.

A reasoning *summary* does not satisfy this, which is why the flag is separate
from `StreamReasoningSummary`. A summary is produced after the fact, is not
always shown, and does not say what the model is about to do to the files on
disk.

### Refuse the batch, do not complain afterwards

Enforcement sits in the actor loop, between gathering the pending calls and
dispatching them. If the model produced no text this round, every call in the
batch comes back as an error and nothing executes.

Refusing *before* execution is the whole point. A tool that has already run
cannot be un-run by a complaint, and effects nobody watched being decided are
exactly what makes a silent run unreviewable. Every call in the batch is
answered, because a vendor requires a result for each call it made: refusing
some and ignoring the rest would trade this fault for a malformed request.

The error text is long on purpose. It is read by a model that has just been
stopped, and it has to do three things -- say plainly that nothing ran, explain
the rule well enough that the model cooperates instead of working around it, and
give an unambiguous next step. A terse refusal gets optimised against.

It also asks for something beyond narration: that the model report what
surprised it. If a tool did not behave as expected, it must say what it expected,
what it got instead, and what it concluded from the difference. The agent is
still being built, and it improves from exactly those reports. A surprise the
model absorbs silently is a bug nobody learns about; a surprise it describes is
a fix. The model is the only party that can see the tools from the inside.

### What the tests do and do not prove

Eight tests cover the decision: whitespace is not narration, a narrated call
passes, other models and unknown models are untouched, every call in a batch is
answered, and the message still contains the clauses that make the agent
improvable.

They do not prove the wiring. Deleting the call site in the actor loop leaves
every one of them green, because they call the function directly -- the same
hole that let three earlier commits ship without firing. The wiring was proved
instead by forcing the condition true and running the real model: the save file
then shows `read_file` answered with `is_error=true` and the refusal text, and
the dispatch line that appears in an unforced run is absent.

That is evidence, but it is evidence that expires. An automated test would have
to drive the actor loop against a fake vendor, and until one exists this
enforcement is one careless edit away from silently doing nothing.

## 34. The client that guessed what only the sequencer knows

Killing the agent mid-turn is not an edge case. It is what you do when a model
starts editing files without saying why, and it is going to happen constantly
until the agent is good enough to trust. The process dies, the save file keeps
a turn state of in flight, and the next launch restores a session that believes
a request is still running when nothing is.

That much is a small bug. What made it a wedge was the GUI.

### The guess

The browser client chose the message type itself:

    if (agentState === 'idle') send {type:'prompt'} else send {type:'hint'}

A hint is delivered to a turn that is already running. If no turn is running it
has nothing to attach to and is dropped. So with a stale in-flight state, every
message typed into the session was classified as a hint to a turn that did not
exist, and vanished. The window was responsive. The connection was open. The
agent answered nothing, forever, and nothing in the interface suggested why.

Chapter 2 states the rule this breaks: a hint and a prompt are the same event,
distinguished only by turn state, and classification belongs in the reducer. We
taught that and then shipped a client that guessed.

The guess cannot be made correct by fixing the state bug, either. The state a
client branches on is a snapshot of something that may have changed before the
message arrives -- the turn can end in the time it takes to send. A client that
classifies is racing the server no matter how accurate its copy of the state is.
The only place that knows whether a turn is running is the place where the
message is sequenced.

### The fix is a deletion plus a move

The GUI now always sends a prompt. The actor classifies, and it does so without
consulting any state variable at all, because its own position already carries
the answer. It drains its mailbox in two places: the main loop, reached when it
is idle, and again between rounds inside a running turn. Arriving at the second
drain is what makes a message a hint. The classification is structural.

The second half matters more than it looks. That mid-turn case previously
re-posted a user message to run as its own turn later. Removing the client's
guess without changing it would have silently converted every mid-tool-call
hint into a queued follow-up turn -- destroying real-time steering while every
test stayed green, because the message was still delivered. It would just have
arrived as a different kind of thing, one turn too late.

### The button

An interrupt control sits beside reset. The server already accepted an
interrupt message and the actor already handled it; there was simply no way to
send one, so a capability that existed in two layers was unreachable from the
only interface anyone uses.

It is worth being precise about what it can and cannot do. The actor drains its
mailbox on the same goroutine that runs a turn, and nothing in the engine holds
a cancel function -- only a timeout. So an interrupt sent during a live request
is not an abort; it waits in the mailbox until the response arrives. What it
does recover is the stale case, where the actor is idle and the message lands
at once. True mid-flight cancellation is a different piece of work, and calling
this button an abort would be a lie told by a tooltip.

The speech toggle moved into the same group, as an icon showing whether speech
is on rather than a checkbox captioned with a word. The hit area stayed the size
it was: the person most likely to reach for a speech control is the person least
able to hit a small target.

### Proof without a browser

The command-line chat client is a second renderer over the same actor loop, so
it tests actor behaviour without any browser rig at all. Running it against a
copy of a real stuck save -- 58 entries, restored state in flight, no dangling
tool calls -- a prompt starts a turn and the reply comes back.

That is the measurement that identifies the culprit. The state was never what
blocked the session, because the same state does not block the text client. The
only difference between the two clients was that one of them was guessing.

One caution learned by doing it: the save file was a copy, but the tools were
not. The agent resumed the old conversation and began rebuilding files that had
just been deleted, against the live working tree. A copied save is not a
sandbox.
