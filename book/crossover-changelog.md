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
