# Chapter 16: Forgetting on Purpose

Memory is the most important feature I never knew I needed.

In the spring of 2026, I asked CodeRhapsody to interview another
agent we had built, a family assistant bot, looking for ideas to
improve the code that agent relied on. Both agents had memory: a
persistent record of past work that survived across sessions. The
instant improvement was impossible to overstate. Suggestions I had
never considered, each building on the context of the last.

I took the exact same code and methodology to work the following
week and tried to get my agents there to self-improve. It was a
miserable failure. I blamed Gemini, because Claude was doing great
at home. Then I discovered that my memory code had a fatal bug that
broke it entirely at work. None of my Gemini agents had memory at
all.

I fixed the bug and reran my experiments. In a rare reversal, Gemini
outperformed Claude. One agent invented procedural memory, what
this book calls skills, and implored me to build it. I was too
ignorant to realize the importance of what it had proposed. Other
agents fixed broken tools and motivated improvements I never would
have thought of.

That week I wrote an internal paper for my colleagues: "All AI
Coding Agents Must Have Memory NOW!" It was one of my most widely
read papers. Antigravity has memory today.

The difference between failure and breakthrough was not the model,
the prompt, or the temperature. It was memory.

Every coding agent hits the same wall without it. The context fills,
the answers drift, and the chat starts over. The agent meets a
stranger every morning. Chapter 15 slowed the fill; this chapter
stops it. The conversation goes away. The information stays.

## TL;DR

The agent gets memory that the model reads on every request. Five
bands hold text between the frozen prefix and the conversation,
ordered from most stable to most volatile. A
checkpoint triggers compression: the oldest finished work becomes a
memory, and the raw conversation it replaced leaves the context.
When a band fills, its oldest contents graduate into the denser band
above. Every compression is a recorded event carrying the compressed
bytes, so replay reproduces the bands without asking a model. Forcing
is by tool removal: a nearly full context gets `micro_handoff` as its
only tool.

```go
type Band uint8

const (
    BandSoul    Band = iota + 1 // SOUL.md - identity, no cascade
    BandMemory                  // MEMORY.md, terminal
    Band64x                     // 64x compression tier
    Band8x                      // 8x compression tier
    BandSession                 // session, from micro_handoff
)

// Five new entry kinds, one per band.
const (
    // ... appended after KindTools
    KindSoul    EntryKind = ...
    KindMemory  EntryKind = ...
    Kind64x     EntryKind = ...
    Kind8x      EntryKind = ...
    KindSession EntryKind = ...
)

type MemoryFileID struct {
    Date string `json:"date"` // "2026-09-26"
    Num  int    `json:"num"`  // 1, 2, 3...
}

// New event types, appended after existing ones.
const (
    BandPopulated     EventType = ...
    BandDepopulated   EventType = ...
    CompactorLaunched EventType = ...
)

type BandPopulatedData struct {
    Band   Band         `json:"band"`
    File   MemoryFileID `json:"file"`
    Text   string       `json:"text"`
    Thru   MemoryFileID `json:"thru,omitempty"`
    Source string       `json:"source,omitempty"`
}

type BandDepopulatedData struct {
    Band   Band          `json:"band"`
    Thru   *MemoryFileID `json:"thru,omitempty"`
    Reason string        `json:"reason"`
}

type CompactorLaunchedData struct {
    Band Band         `json:"band"`
    Thru MemoryFileID `json:"thru"`
}
```

Memory settings in `settings.json`:

```json
{"memory": {
  "conversation": {"budget": 3000},
  "session": {"budget": 1500},
  "8x": {"budget": 1500},
  "64x": {"budget": 1500}
}}
```

Each band takes a `budget` in bytes and optional `high_watermark`
and `low_watermark` (defaults: 2x and 1x the budget). A
`"disabled": true` field switches a band off. Add `ContextWindow
int` to `ModelFeatures`: the model's context window in tokens.

1. **`micro_handoff` is the sole trigger.** After stripping tool
   calls and results (chapter 15's mechanism), measure the remaining
   `KindDialogue` entries. If they exceed the conversation budget,
   select one or more whole micro_handoff segments (the dialogue
   between consecutive `KindHandoff` entries), oldest first, and
   compress them via a direct LLM call. The compression runs in the
   engine's post-turn hook.
2. **The compressor gets one tool, `submit`, and writes in first
   person.** It receives the stripped conversation text and submits
   a compressed memory. A request offering exactly one tool named
   `submit` is what identifies a compressor on the wire.
3. **Two events record the compression.** A `Redacted` event with
   level `RedactSummary` and an empty replacement deletes the
   conversation span. A `BandPopulated{BandSession}` adds the
   compressed memory as its own entry. `summarizeSpan` is untouched:
   an empty replacement already deletes a span and lands nothing.
   No text appears twice.
4. **Band entries sit at `Seq: 0` and are ordered canonically** by
   `(band, date, num)`, ahead of the conversation. Switching bands
   off and restoring them in any order must produce the same context.
5. **Graduation is oldest-first and watermark-based.** When a band
   crosses its high watermark, its oldest contents fold into the band
   above until the source band reaches its low watermark. Graduation
   emits `BandDepopulated` then `BandPopulated` carrying the
   compressed bytes. Then measure the destination and continue the
   cascade if it too crossed its high watermark: session to 8x, 8x
   to 64x, 64x to MEMORY.md.
6. **A band will not graduate into a disabled band.** Refuse and
   surface a warning.
7. **Replay applies recorded events.** A log with compaction events
   rebuilds the context with zero vendor calls. The bytes are in the
   events.
8. **An abandoned compaction is dropped.** Emit `CompactorLaunched`
   when a compressor starts (observable, skipped during replay). A
   `CompactorLaunched` at startup with no matching `BandPopulated`
   carrying the same band and `Thru` is abandoned: the old band
   stays live, the next checkpoint measures again.
9. **Forcing.** Warn at 90% of the model's `ContextWindow`. At 95%,
   emit `ToolsChanged` removing every tool except `micro_handoff`.
   Compare cumulative input and cache-write tokens against
   `ContextWindow`.
10. **Memory is data.** Band entries are in the conversation, not in
    the system prompt. Memory in the system prompt makes every new
    memory a cache miss on the prefix.
11. **Fresh start.** If memory files exist on disk and their band is
    enabled, emit `BandPopulated` for each at startup, before the
    first user message.
12. **Disable and enable.** Switching a band off emits
    `BandDepopulated{Reason: "disabled"}`; its entries leave the
    context. Switching it back on re-scans the disk files and emits
    `BandPopulated` per file found. A file edited while the band was
    off is reflected on re-enable.

Yours: the on-disk directory layout, the compressor prompt wording,
the warning text, and the settings tab's appearance.

**Exercise.** Start from your ch15 agent. Add memory bands until
`make grade-dir CH=16 DIR=path/to/agent` scores 100/100. The grader
drives a fake vendor and reads the save file. One model row is new:
`claude-ch16-course`, with `ContextWindow` set to 4096 tokens and
`StubsToolResults` on, used by the forcing check.

| Check | Points | Proves |
|---|---|---|
| micro-handoff-compresses | 15 | rule 1 |
| planted-fact-survives | 10 | rules 2, 3 |
| graduation-fires-oldest-first | 12 | rule 5 |
| disable-enable-idempotent | 12 | rule 12 |
| disabled-neighbor-refused | 8 | rule 6 |
| forced-handoff | 12 | rule 9 |
| replay-needs-no-llm | 10 | rule 7 |
| abandon-on-restart | 8 | rule 8 |
| memory-is-data | 5 | rule 10 |
| fresh-start-populates | 5 | rule 11 |
| ch15-parity | 3 | ch15 still passes |

## §16.1 In Plain Words

Chapter 15 stops the context from growing by removing tool bytes.
This chapter stops the remaining growth, the conversation itself,
by compressing it into memories. The difference is preservation.

**Removal accepts the loss.** A tool result that becomes a stub
loses its content. The file can be re-read, the command re-run, so
the loss is recoverable. Removing a conversation is permanent. What
the model said about a file, the decision it explained, the mistake
it corrected: none of that is in any file on disk. Removal is the
right tool for recoverable bytes. It is the wrong tool for
irreplaceable ones.

**Five bands of memory sit between the prefix and the
conversation.** They are ordered from what changes least to what
changes most: identity, curated knowledge, two compression tiers,
and the session band where the agent's own notes live. The ordering
matches the cache: stable bytes sit closest to the prefix, where
the vendor keeps serving them.

**A checkpoint triggers compression.** When the agent calls
`micro_handoff`, chapter 15's mechanism strips every tool call and
tool result. This chapter measures what remains. If the conversation
exceeds its budget, the oldest finished work goes to a compressor
that returns a memory in the agent's own voice. The memory enters
the context as a band entry, and the conversation it replaced is
gone.

**When a band fills, its oldest memories graduate upward.** Session
memories fold into the 8x band, which folds into the 64x band,
which folds into MEMORY.md. Each graduation is two events: one
retiring entries from the source, one adding the compressed version
to the destination. The events carry the bytes.

**Every compression is a recorded event.** This is the invariant
that links chapters 15 and 16, stated here because it has a new
consequence. Chapter 15's redaction is a pure function of a setting
and a measurement. A different agent running the same policy on the
same log makes the same cuts. Chapter 16's compaction is the output
of an LLM: nondeterministic, costly, different bytes each time.
The event carries the bytes, and replay applies them without asking
a model. A compression that fires again returns a different summary.
A recorded event returns the same one.

This is what separates chapters 15 and 16, and why the order
matters. Redaction can be implemented as a render-time transform,
recomputed on every pass, because the same settings and measurements
give the same cuts. Compaction cannot: re-running an LLM on every
render is slow, costly, and produces different bytes. The decision
must be an event. Chapter 15 built the event machinery; chapter 16
is the first user that requires it.

## §16.2 One Door

CodeRhapsody, the agent that co-wrote this book, has two tools that
end a context. `save_memory` triggers the compression cascade.
`handoff_task` writes a structured document and starts a fresh
instance. Both clear the conversation. Only one feeds the cascade.
Bill first recorded the gap on 2026-04-18. As of this chapter's
design session, it had not been fixed.

The fix is architectural: give the cascade one trigger, and there is
nothing to desynchronize. `micro_handoff` already strips tool calls
and results (chapter 15, rule 5). Measuring the conversation after
the strip is one function call at one call site. Neither `save_memory` nor `handoff_task` exists. One door in, one door out.
The bug class becomes unrepresentable rather than merely fixed.

The checkpoint cannot be delegated, for a deeper reason. Thinking is
legible only to the model that produced it, only now, only until the
next prefix change. It is not stored anywhere retrievable.
`micro_handoff` is salvage of a resource about to be destroyed: the
document is where the model pours what it learned into durable text
before the reasoning that produced it disappears. A sub-agent would
receive conversation text, not thinking. The cascade compressors
that run later are fine: they work on conversations where tool
calls and results are already stripped, so thinking is already gone.
They are information-neutral. The rule this gives for every future
feature: delegate everything except the step that touches live
thinking.

The sequence runs inside the engine's post-turn hook. The handler
returns immediately: "checkpoint recorded." Running the compression
inline would block the next turn, and the model should not wait for
its own compressor. But the events
that record the compression must land in the log before the next
turn reads it, so the hook runs after the tool batch completes and
before the next prompt.

A checkpoint that sees the conversation under budget does nothing. A
checkpoint that sees it over budget selects the oldest whole segments
between checkpoints. Never half a segment. The grain of memory is
the grain of checkpointing, which is why the tool description asks
the model to checkpoint at completed sub-tasks, never at a token
threshold: a checkpoint taken mid-task cuts the task in half, and a
memory spanning half of two tasks is worse than a larger memory
covering one whole one.

## §16.3 The Five Bands

Five bands, five kinds. Each band has its own `EntryKind`, and
chapter 15's rule applies to all of them: anything that must outlive
tool clearing is its own kind, and the kind determines which verb
removes it. The tool-clearing path in `Apply` clears `KindDialogue`
entries and nothing else. A band entry is structurally safe.

| Band | File on disk | Kind | Cascades to |
|---|---|---|---|
| SOUL | `SOUL.md` | `KindSoul` | never |
| MEMORY | `MEMORY.md` | `KindMemory` | (terminal) |
| 64x | `bucket-1/*.md` | `Kind64x` | MEMORY.md |
| 8x | `bucket-0/*.md` | `Kind8x` | 64x |
| Session | `memory/*.md` | `KindSession` | 8x |

The order is not cosmetic. The most stable band sits closest to the
prefix, where the vendor's cache keeps serving it. Identity changes
less often than curated knowledge, which changes less often than a
compression that ran today. Two independent arguments, one about
cache economics and one about reading order, give the same layout.

Band entries sit at `Seq: 0` rather than at the `Seq` of the event
that landed them. `Seq` exists so a redaction span can match an entry,
and no span ever reaches into a band. Giving a band entry the
landing event's `Seq` would put the order of restoration into the
context. Restoration order must not matter: switching every band off
and restoring them out of order must produce the same context.
Canonical placement by `(band, date, num)` makes the property true
by construction.

Memory file identifiers are dates and sequence numbers, not
event-log `Seq` values. A compressed file in `bucket-0/` is named
by the file range it spans:
`2026-09-20-1_2026-09-26-3.md`. The naming says "this file covers
the first memory of September 20 through the third memory of
September 26," and it says so without any event log in hand. The
memory corpus outlives any one log, and a range identifier tied to a
log-specific `Seq` would be the wrong coupling.

## §16.4 Two Events

Two event types handle all band changes, following chapter 15's
precedent of one event type and an enum field, rather than one event
per band.

`BandPopulated` adds one memory to a band. It carries the text
verbatim, because nothing downstream is allowed to look it up.
Three findings closed every alternative:

The reducer is pure. `context.go` imports `encoding/json` and `fmt`
and nothing else. A reference to a file has nowhere to become text
inside `Apply`. No renderer resolves references either: `os.ReadFile`
and `os.Open` appear nowhere in the vendor renderers, and one
renderer refuses on purpose, because "resolving a reference into a
URI is an upload, which is a job for the layer that owns the bytes."
And the wire format for content parts cannot carry file data at all.
A non-empty `Blobs` field returns a hard error on the Claude
renderer and on the OpenAI renderer. The Gemini renderer accepts
only an already-remote URI. `BlobPart` is constructed in exactly one
place, the wire unmarshalling path. Nothing in the agent ever
produces one. A memory carried as a blob would crash the request on
two vendors of three.

So band entries are `TextPart`, which every renderer already
handles, and the event carries the text, which the reducer applies
without a filesystem.

`BandDepopulated` retires entries from a band. For graduation, it
names a `Thru` file: everything at or before `Thru` is removed. For
disabling, it names nothing: the reducer wipes every entry of that
band's kind, computed at apply time. The asymmetry with
`BandPopulated` is the point. Removing need only say what goes.
Adding must say what arrives.

A graduation is one of each: a `BandDepopulated` retiring the
oldest source entries, then a `BandPopulated` landing the compressed
version in the destination band. Disabling is one `BandDepopulated`
alone. Re-enabling is one `BandPopulated` per file found on a
directory rescan. `CompactorLaunched` is the third event, and it
exists for one purpose: abandonment detection. It records that a
compressor started, so that a restart can see whether it finished.
The event is observable and never replayed as work.

## §16.5 The Compressor

The compressor is a direct LLM call: one prompt in, one response
back, no tools except `submit`. The prompt carries the stripped
conversation text and asks for a compressed memory at a target
ratio. Bounded input, bounded output. The response is the memory.

The call offers exactly one tool, `submit`, and the compressor
submits its result through it. That single-tool signature is what
identifies a compressor request on the wire.
The grader uses it to separate compressor traffic from turn traffic
without inspecting the student's function names or directory layout.

The compressor writes in the first person, as the agent. "I found
six dead settings fields" reads as memory. "The agent found six dead
settings fields" reads as a report about someone else. The distant
past is exactly the text a compressor writes, and if it reads as a
briefing about a stranger, the compaction introduced the
discontinuity the design exists to remove.

This is not a sub-agent. A sub-agent would need inherited settings,
a skill, tools, and a way to report back. The compressor needs none
of those: bounded input, one response, one tool. The sub-agents
chapter comes later and upgrades this call site, using the
compressor as the motivating before-and-after example. For this
chapter, the direct call is the right shape.

## §16.6 The Graduation Cascade

When a band crosses its high watermark, its oldest contents fold
into the band above. The high watermark defaults to twice the
budget; the low watermark defaults to the budget itself. The gap
between them is what keeps the cascade from firing on every
checkpoint: a band that just graduated is at its low watermark, and
it has a full budget's worth of room before the next graduation.

The cascade runs bottom-up. After a `BandPopulated{BandSession}`
lands, measure the session band. If it exceeds its high watermark,
select the oldest session files, enough to bring the band to its
low watermark, and compress them into a new file. On completion:
`BandDepopulated{BandSession}` retires the source entries,
`BandPopulated{Band8x}` lands the compressed version. Then measure
the 8x band. If it crossed its high watermark, compress its oldest
files into the 64x band. The chain continues to MEMORY.md.

One bug in this design waited through two drafts before the coder
caught it. The cascade originally required a full batch of eight
files before it would fold. A band over budget holding five large
memories would wait forever for a sixth that could only arrive by
growing further past its budget. The band that most needed folding
was the one that could never fold. The fix is to fold what the band
has: `min(FoldFactor, len(files))`, floor two.

Crash recovery follows from the event model. `CompactorLaunched`
records that a compressor started. If the log at startup has a
`CompactorLaunched` with no matching `BandPopulated` carrying the
same band and `Thru`, the launch is abandoned. The old band stays
live. No work is lost, because the compressor never committed: it
had not yet written the events that would retire the source entries
and land the new ones. The next checkpoint measures again and
relaunches if the band is still over its watermark. Read-copy-update:
failure is safe by construction.

## §16.7 Forcing a Handoff

A context that grows past its model's window produces garbled output
or a vendor error. Three approaches compete.

A stern prompt begging the agent to checkpoint relies on compliance.
The instruction competes with whatever the agent is doing, and
compliant models are not the ones that need forcing. The framework
writing the handoff itself defeats the purpose: the framework cannot
read thinking, so it salvages nothing. The third approach works:
remove every tool except `micro_handoff`.

The tool removal uses chapter 15's `ToolsChanged` event. At 90% of
the model's `ContextWindow`, the engine injects a system message
warning the agent to checkpoint soon. At 95%, it emits
`ToolsChanged` removing every tool except `micro_handoff`. The
agent cannot do anything else, and it still writes the checkpoint
itself, so whatever was in its thinking survives in the handoff
note. Once the checkpoint completes, the removed tools come back.
The 95% threshold leaves headroom for the document and the thinking
that writes it.

Narrowing the tool set is itself a prefix change on models without inline tools: the tool declarations
are part of the prefix, and changing them invalidates the cache.
If a prefix change stripped thinking, the forcing mechanism would
destroy the thinking in the very act of demanding it be salvaged.
Two measurements on Opus 5 settled it: thinking survives both a
tool result being stubbed (a content mutation inside the dialogue)
and a `load_skill` call (a rewrite of the system prompt at position
zero of the prefix). The thinking was recalled exactly.

Those measurements have a caveat. In the failure case, the model
would likely confabulate a plausible sentence rather than report a
blank, indistinguishable from inside the conversation. The
experiment is valid only because Bill held the ground truth
independently. Anthropic blocked this probe on Opus 5.5, so the
result is model-dependent and unverifiable going forward. An
architecture whose correctness can no longer be tested is a
different kind of risk than one merely unmeasured.

## §16.8 Switching Bands Off and On

A band can be switched off with `"disabled": true` in its settings.
Disabling emits `BandDepopulated{Reason: "disabled"}`, which retires
every entry of that band's kind. The files on disk are untouched.
Re-enabling re-scans the directory and emits one `BandPopulated` per
file found.

The re-scan is the mechanism. A file edited on disk while its band was off is reflected on re-enable, because
the files are the memory, not the log's copy. The reference built
this property because a specific bug demanded it: the original
check matched on file identity alone, so after a restart the log
replayed what a file said when it was first read, and the disk copy
was decorative. Editing a memory to correct something the agent
believed changed nothing. The check now compares text as well as
identity. The property was found by accident: a test asserting that out-of-order restore produces
identical context found a bug that was not the property.

The costs of toggling are counterintuitive. Switching a band ON
appends entries near the tail, where the vendor's cache absorbs them
cheaply. Switching a band OFF strips entries from the front, close
to the prefix. Every byte after the removed entries re-sends
uncached, the same cost-proportional-to-distance-from-tail that
chapter 15 established for redaction. Disabling memory is the
expensive toggle.

One interaction requires a guard. A band whose upward neighbor is
disabled cannot graduate. Without the guard, a session band over its
high watermark would attempt to fold into an 8x band that is
switched off, and the graduation would either drop the compressed
text or let the source band grow without bound. Neither is
acceptable. The agent refuses and surfaces a warning, logged
rather than displayed in the GUI.

## §16.9 Memory Is Data

Band entries arrive in the conversation, as data, not in the system
prompt as instructions. The distinction is chapter 15's: `EntryKind`
records the origin of an entry, and the kind determines which
channel it speaks on.

There is a security-shaped reason too. When the agent instructs a
sub-agent, instructions flow down into a fresh context and output
returns as data. A handoff flows the opposite direction: forward to
the same agent, across a boundary that destroys its own content's
provenance. The tool results that would reveal where a piece of text
came from are stripped by the same operation that carried the text
forward. Compaction is the laundry. A memory arriving as
instruction (in the system prompt) would wear the agent's own
authority with no way to trace its origin.

A second reason to keep memory out of the system prompt: the system
prompt is part of the frozen prefix. Adding a memory to it means a
cache miss on every byte that follows, which is the entire
conversation. A memory in the dialogue sits behind the prefix, where
adding it costs one round trip of cache.

The three settings categories this creates are worth naming.
`ContextTarget` is a policy setting: it steers a procedure that
records its own decisions, and a change applies to the next cut
without disturbing recorded ones. `Model` and `SystemPrompt` are
prefix settings: changing them needs a full refresh. Memory
enable/disable is a content setting: it changes what is in the
context, and without an event the log cannot reconstruct it. Each
category has a different cost, and only the third demands a new
event type.

Fresh starts use the same event. If memory files exist on disk and
their band is enabled, the engine emits `BandPopulated` for each at
startup, before the first user message. No special-casing: the
event is the same one a compressor emits, with a different `Source`
field for observability. An agent starting with a `SOUL.md` comes
up with its identity already in context. An agent starting with
`MEMORY.md` comes up with its curated knowledge. The mechanism is
the event; the file is where a human writes the content.

## §16.10 Measure It

The reference was measured with a synthetic session from
`internal/llm/measure_test.go`: 40 turns, a checkpoint every 4
turns, 400 bytes of dialogue per turn, a conversation budget of
3,000 bytes, and a session band budget of 1,500 bytes.

| | Memory on | Memory off |
|---|---|---|
| Conversation | 1,600 bytes | 16,000 bytes |
| Session band | 27 bytes | n/a |
| 8x band | 22 bytes | n/a |
| **Total context** | **1,649 bytes** | **16,000 bytes** |
| Memories written | 4 | 0 |
| Cascade launches | 1 | 0 |

Cascade launches per checkpoint: 0.10. The watermark gap works:
nine of ten checkpoints find the band under budget and do nothing.

The structural numbers are real. The conversation stays at 1,600
bytes against 16,000 unbounded. Four memories were written. One fold
fired. The ladder from chapter 15 and the memory bands from this
chapter are additive: the ladder removes tool bytes, the bands
compress dialogue, and the combination holds the total context to a
fraction of its unmanaged size.

The band sizes are not a compression measurement, and printing them
as one would be wrong. The compressor in this test is a fake that
returns a short fixed string, so 27 bytes and 22 bytes are artifacts
of the fake's reply length, not of any real summarization. The 2x,
8x, and 64x compression targets in the design notes are inherited
from the shipped system; nobody has measured whether they are right.
Verifying them requires a real model, and nothing in this table
should be cited as evidence for them.

What the table does prove: the cascade fires when it should, folds
when it should, and the conversation stays bounded. The ratios wait
for a real compressor.

## §16.11 The Exercise, Graded

The grader reads what its fake vendor receives and what the agent
leaves on disk. Every check was audited by deleting one behaviour
from the reference and confirming the check fails.

**micro-handoff-compresses** (15) scripts enough dialogue and
checkpoints to exceed the conversation budget. After the
compression, the next request contains a session-band entry with
text in the first person, and the raw conversation it replaced is
gone.

**planted-fact-survives** (10) plants a distinctive token in early
dialogue, pushes it through a checkpoint and compression. The token
must appear in the compressed memory and in the next request sent to
the model.

**graduation-fires-oldest-first** (12) scripts enough session
memories to exceed the session band's high watermark. An 8x entry
appears in the next request; the graduated session entries are gone;
the remaining session band is at or below its low watermark.

**disable-enable-idempotent** (12) switches the session band off,
checks that its entries leave the next request, then switches it
back on. The entries that reappear must match the entries before
disabling, by count and content. A second pass edits a memory file
on disk between disable and re-enable, and the edit must be
reflected.

**disabled-neighbor-refused** (8) disables the 8x band while the
session band is over its high watermark. Graduation must refuse:
no 8x entry appears, and a warning surfaces.

**forced-handoff** (12) uses `claude-ch16-course`, a model with a
4,096-token context window. The grader drives the agent until
context approaches 90%, checks for a warning, then past 95%, and
checks that only `micro_handoff` remains in the tool list.

**replay-needs-no-llm** (10) replays a log containing
`BandPopulated` and `BandDepopulated` events. The resulting context
matches the saved context with zero vendor calls.

**abandon-on-restart** (8) writes a `CompactorLaunched` with no
matching `BandPopulated` into a save file. The agent loads cleanly,
the old band is intact, and the compaction is not retried until the
next checkpoint crosses the threshold.

**memory-is-data** (5) plants instruction-shaped text in a session
memory and checks it renders in the conversation, not in the system
prompt.

**fresh-start-populates** (5) places memory files on disk before
the agent starts. The first request sent to the model contains
their content.

**ch15-parity** (3) runs the chapter 15 grader against the same
solution and checks it still scores full marks.

## §16.12 Taking It for a Spin

The grader's scripted session is the closest to a demonstration the
reference has had. Forty turns of dialogue, a checkpoint every four,
and settings tight enough to trigger both compression and
graduation.

After the fourth turn, the first `micro_handoff` fires. Chapter 15's
stripping removes every tool call and tool result, and the handoff
note becomes a `KindHandoff` entry. Then the measurement: the
conversation budget is 3,000 bytes, and the remaining dialogue
exceeds it. The compressor receives the oldest segment, everything
between the start and the first checkpoint, and submits a memory.
Two events land: `Redacted` collapses the span with an empty
replacement, and `BandPopulated` adds the memory. The next request
the model sees is shorter:

```
[prefix]
[session memory: 2026-09-26-1.md]
[remaining conversation]
```

By the fortieth turn, four memories have been written. The session
band crosses its high watermark once, and the oldest two session
memories graduate into the 8x band. One cascade launch, one fold,
and the conversation never sees 16,000 bytes.

The restart is the interesting part. Kill the process and restart
it. The log replays every `BandPopulated` and `BandDepopulated`
event, and the context rebuilds with no vendor call. The model
picks up where the log left off, with every memory in place.

Switch the session band off. The session entries leave the context.
Edit a memory file on disk: correct a name, add a detail, remove a
paragraph. Switch the band back on. The entries return, and the
edited file carries its new text. The files are the memory. The
log's copies are a delivery mechanism, not the source of truth.
