# Chapter 15: The World's Best Context Engineering, Before Breakfast

This may be the most important chapter in the book.

If you've ever started a new chat because your agent seemed confused,
well, that's yesterday's context engineering. I did it myself, several
times, on the day I designed this chapter. As of September 2026, what
follows is tomorrow's. The book is evergreen, so I hope it always has
modern advice.

The design had been in my head for a couple of months. On a Tuesday
morning I typed it out for about twenty minutes while CodeRhapsody
turned it into a design record and checked it against the code. Then I
took a shower. The first commit landed at 7:42, before breakfast, and
frankly it felt good to get it out.

This is how long-running actors should work: what the actor knows,
organized by section at several levels of detail, with the most
critical parts kept longest, instead of some lame summary of the first
half of the context.

Every long session with a coding agent degrades the same way. The
answers get vaguer. The agent re-reads a file it read an hour ago, then
forgets a decision made before lunch. A fresh chat fixes it because a
human decides what to carry over. This chapter moves that decision
inside the agent, makes it continuous instead of catastrophic, and
records every cut as an event, so nothing leaves the window without a
record of what removed it.

## TL;DR

The context stops growing without bound. Every entry gets a kind that
says which verb removes it. Tool bytes fall off a ladder of recorded
redaction events. Each round trip's results are stubbed unless the
model keeps them. The actor can checkpoint with `micro_handoff`. Loaded
skills survive as entries instead of as tool results. The system
prompt and fixed tools never change mid-session. The log reaches disk
as it grows, so a process crash loses nothing.

```go
// EntryKind says why an entry is in the context and which verb
// removes it. Anything that must outlive tool clearing is its own kind.
type EntryKind uint8

const (
    KindDialogue EntryKind = iota + 1 // prompts, hints, output, tool parts
    KindHandoff                       // from MicroHandoff
    KindSkill                         // from SkillLoaded
    KindTools                         // from ToolsChanged
)

type Entry struct {
    Seq   Seq       `json:"seq"`
    Actor Actor     `json:"actor"`
    Kind  EntryKind `json:"kind"` // new
    Parts PartList  `json:"parts"`
}

// Appended to the EventType list after SkillLoaded. A number once
// assigned is never reused.
const (
    // ...
    SkillLoaded
    MicroHandoff // new
    ToolsChanged // new
)

type MicroHandoffData struct {
    Text string `json:"text"`
}

// A delta against the declarations in force just before it.
type ToolsChangedData struct {
    Added   []ToolDecl `json:"added,omitempty"`   // full declarations, ch10's shape
    Removed []string   `json:"removed,omitempty"` // names only
}
```

Reused unchanged: ch2's `RedactData{From, To, Level, Replacement,
Reason}` with levels `RedactResult` (the result becomes a stub, the
call survives) and `RedactTool` (both go), and ch10's `SkillData{Name,
Body}`, which already carries the skill's body.

1. **Frozen prefix.** The system prompt and the startup tool
   declarations are byte-identical on every request of a session. Only
   a full refresh changes them, and, on a model that cannot carry tools
   in the dialog, a tool change (rule 2).
2. **Tools arrive through the dialog.** A skill load, a skill unload,
   or an MCP connect mid-session emits `ToolsChanged`, and the reducer
   turns it into a `Tools` entry. On a model whose features row sets
   `InlineTools` (today, the Anthropic API), that entry carries the
   declarations in the dialog. They ride in a new part,
   `ToolDeclPart{Added []ToolDecl, Removed []string}`, JSON type
   `"tool_decls"`: a declaration is not a call or a result, so rule 3
   still holds. A model without `InlineTools` folds every delta into
   the startup set and re-declares, paying the cache miss.
3. **Survivors carry no tool parts.** A `Handoff`, `Skill` or `Tools`
   entry never holds a tool call or a tool result, so no tool clearing
   can touch it.
4. **Skills are entries.** The reducer turns `SkillLoaded` into a
   `Skill` entry holding the body. The `load_skill` tool result is an
   acknowledgement only. `unload_skill` removes the skill's tools
   through `ToolsChanged` and leaves the `Skill` entry where it is:
   deleting old bytes would miss the cache, and ch10's unload is lazy.
   A later chapter's verb removes it.
5. **`micro_handoff` is three records:** the tool call, an ordinary
   result, then a `MicroHandoff` event. Its reducer waits until the
   batch's last result has arrived, so a call made alongside
   `micro_handoff` is cleared with it. Then it removes every tool call
   and tool result part from the context, drops entries left empty,
   and appends one `Handoff` entry. The text appears in the next
   request exactly once, and no call is ever left without its result or
   a result without its call.
6. **Keep or stub, per round trip.** On a model whose features row sets
   `StubsToolResults`, every tool result above the stub threshold is
   replaced by a stub in the request after the one that carried it,
   unless the actor's next message calls `keep_tool_results`, which
   keeps that whole batch. The keep governs the batch before the message
   that calls it: called alongside other tools, it keeps the previous
   batch, not the one it rides in, and its own result is never stubbed.
   The call survives either way. The stub is an ordinary `RedactResult`
   event, so replay reproduces it. Other models get no per-round-trip
   stubbing; the ladder alone applies.
7. **The ladder records a Seq.** With target size T, the newest T/8
   bytes of tool traffic (the results band) stay whole. Older than that,
   results become stubs (`RedactResult`) and their calls stay, for T/16
   bytes of calls and stubs (the calls band). Tool bytes older than both
   bands go entirely (`RedactTool`). A band is cut in steps: when it
   passes twice its budget, one event cuts it back to its budget. Each
   event stores `To` as a number, never as "the watermark", so a later
   settings change cannot rewrite the past.
8. **The reducer is total.** A save file that does not parse is still
   refused (ch11 rule 2). An event that parses but cannot be applied,
   such as a malformed payload or a redaction naming an entry already
   gone, is skipped with a diagnostic in the agent's log, and loading
   continues.
9. **Crash-safe.** Events reach disk as they happen, so a process crash
   leaves a tail after ch11's `as_of` anchor, and recovery is ch11's
   load: snapshot plus tail. A normal shutdown snapshots. Power loss is
   out of scope; the log is never fsynced. Ungraded, but
   do it anyway: write the new snapshot before truncating the log, never
   truncate past the anchor, and keep one backup of the previous
   snapshot.
10. **One knob.** The Context Management settings tab sets
    `context_target` in `settings.json`: T in bytes, 0 for the default
    of 400,000, clamped to at least 20,000. The stub threshold is T/100
    and the bands are rule 7's, so the default gives 4,000, 50,000 and
    25,000. The tab's other setting, `log_retention`, is the number of
    events kept in the saved log, 0 for all.

Yours: the on-disk layout of the log, what a diagnostic says, and the
settings tab's appearance.

**Exercise.** Start from your ch14 agent. Add context management until
`make grade-dir CH=15 DIR=path/to/agent` scores 100/100. The grader
reads only the requests its fake vendor receives and the files on disk.
It launches the agent as `claude-opus-5-course`, whose features row
sets `StubsToolResults` and `InlineTools`, and as
`claude-sonnet-5-course`, whose row sets neither. Both flags are new
`bool` fields on `ModelFeatures`, false for every existing row. The
sonnet row already exists from Chapter 14 and needs no change; add the
opus row.

| Check | Points | Proves |
|---|---|---|
| skill-survives-the-ladder | 15 | rules 3, 4, 7 |
| micro-handoff-shape | 15 | rule 5 |
| keep-or-stub | 10 | rule 6 |
| ladder-is-recorded | 15 | rule 7: replay under changed settings reproduces the cuts |
| frozen-prefix | 10 | rules 1, 2 |
| replay-equals-snapshot | 15 | rules 8, 9 |
| crash-recovery | 15 | rule 9: kill mid-session, restart, nothing lost |
| total-reducer | 5 | rule 8 |

## §15.1 In Plain Words

The context window is not storage. The log is storage: Chapter 2 made
it append-only, and it keeps every byte the agent ever saw or said. The
window is a working set rendered from that log, and every byte in it is
paid for again on every request. So every byte has to answer three
questions: why is it here, who put it here, and what removes it. A byte
that cannot answer the third question stays forever, and a window full
of such bytes is how a session degrades until a human resets it by hand.

Most of the bytes are tool bytes. Chapter 2 printed a measurement from
real coding sessions: tool results were about 42 percent of
conversation history by volume, and tool-call arguments another 30
percent. Nearly all of those bytes are needed once. The agent reads a
file, decides, edits; after that, the file's
contents are recoverable from disk and the decision lives in what the
agent said about it. Hence the chapter's rule: keep the words, let the
bytes go. Dialogue stays. Tool results go first, then the calls that
produced them, oldest first.

Some things must never go with them. A loaded skill's instructions and
the note an agent writes to its future self at a checkpoint look like
tool traffic today, because they arrive through tools. If they stay
tool traffic, the first cleanup deletes the manual and keeps the tools
it explains. So anything that must survive becomes its own kind of
entry, removed only by its own verb. Survivors are safe by
construction.

Every cut is an event in the log with an exact number in it; nothing is
re-evaluated later. That is what keeps Chapter 11's promise:
replaying the log reproduces the same window, byte for byte, even after
the settings that chose the cuts have changed.

Two laws decide where cuts may happen. The front of the request, the
system prompt and the startup tools, is frozen, so the vendor's cache
can keep serving it. The rest is append-only, and the cost of changing
a byte grows with its distance from the end, because everything after
it must be re-sent uncached. So cuts come in steps, not a trickle: one
larger cut now and then costs less than a small one every turn.

None of this is a new data structure. The redaction levels are
Chapter 2's, the compaction event is Chapter 2's, the frozen prompt is
Chapter 10's, the snapshot and replay are Chapter 11's. What was
missing is the policy that decides when to use them.

## §15.2 Why Is This Byte Here?

Chapter 2 built the machinery for this chapter and then nothing called
it. Its `RedactData` event has four levels. `RedactResult` turns a tool
result into a stub and keeps the call. `RedactTool` removes both and
keeps the reasoning around them. `RedactDialogue` and `RedactSummary`
reach into the conversation itself. The reducer applies all four, the
tests cover all four, and until now no production code in the agent
ever emitted a single one. The dialogue level even says where the
missing piece lives:

```go
case RedactDialogue:
        // Prose and reasoning go. Survivors are defined by the compaction
        // policy, which is a later chapter's problem.
```

This is that chapter. Context engineering, as the book uses the term,
is managing every byte in the context data structure as well as
today's models allow. The management is a policy: for each byte, a
reason it is in the window, a record of what put it there, and a verb
that takes it out.

Walk a request from a ch14 agent after an hour of work and ask each
byte the three questions. The system prompt answers all three: the
agent's constitution, written at startup, removed by nothing. A user
prompt answers them. A 9,000-byte file read forty minutes ago answers
the first question with "the agent needed it once", the second with
`read_file`, and the third with silence. Nothing removes it. It rides
along on every request until the session ends or a human copies the
good parts into a fresh chat.

The fix is that silence. Once every byte has a remover, the window
stops growing.

## §15.3 Two Laws

The vendor's prompt cache decides where a cut may happen, and it
works by prefix. The Anthropic API, as of September 2026, orders a
request as tools, then system prompt, then messages, and serves from
cache the longest prefix it has seen before. Everything after the first
changed byte is re-read at full price.

**Law 1: the prefix is frozen.** The system prompt and the startup tool
declarations are byte-identical on every request of a session. Chapter
10 already made the system prompt a constitution and put dynamic
skills in the dialog. It left one leak: a skill that brings tools
changed the tool list, and the tool list is the first thing in the
request. Loading one skill invalidated the whole cached conversation.

The chapter closes the leak with an event. A skill load, a skill
unload, or an MCP server connecting mid-session emits `ToolsChanged`,
a delta of declarations added and names removed. The reducer turns it
into a `Tools` entry at the tail of the context, where the change
costs one round trip of cache. On a model whose features row sets
`InlineTools`, the renderer sends that entry as a mid-conversation
system message:

```json
{"role": "system",
 "content": [{"type": "tool_addition",
              "tool": {"type": "tool_definition",
                       "definition": {"name": "gui_click",
                                      "description": "...",
                                      "input_schema": {"...": "..."}}}}]}
```

with the beta header
`mid-conversation-tool-changes-2026-07-01,inline-tools-2026-09-15`. A
removal is the same message with `tool_removal` blocks carrying names.
A model without the feature gets the old behaviour: the renderer folds
every delta into the startup set, re-declares, and pays the miss. The
cost lands on the model that cannot avoid it, and only there.

The prefix does change on purpose once in a while. A settings change
that rewrites the system prompt is a full refresh, and the full refresh
is an accepted miss, paid once, by a human who asked for it.

**Law 2: the cost of changing a byte grows with its distance from the
end.** Everything after the changed byte goes out uncached, so the
price of an edit is the number of bytes behind it. Stubbing the result
that arrived on the last round trip costs about one round trip.
Stubbing a result a hundred thousand bytes back costs a hundred
thousand bytes, every time.

The law was measured before it was believed. The design predicted that
stubbing every tool result one round trip after it arrived would wreck
the cache, since the request changes on every trip. The first session
run that way, on CodeRhapsody with Opus 5.5 on 2026-09-22, measured a
cumulative cache hit rate of 77 percent, after a cold first request
that missed on over 100,000 tokens. The prediction had the law
backwards: a just-finished result sits at the tail, where changes are
cheap. The cumulative figure also hides the steady state, because that
one cold miss is averaged into every later request, which is why a
display of cache rate should show the last request beside the total.

Law 2 has a second consequence. A trickle of small cuts deep in the
window misses the cache on every turn; one larger cut now and then
misses once. So the agent cuts in steps.

## §15.4 What Survives Is Never a Tool Call

The ch14 agent returns a loaded skill's manual as the result of the
tool that loaded it:

```go
result += "## Instructions\n\n" + body
```

That line, at `internal/tools/tools.go:1016` in `solutions/ch14`, is a
bug the moment any cleanup exists. Tool results are the first bytes
to go. The first time the ladder stubs old results, the manual becomes
a one-line stub and the tools it explains stay declared. The agent
keeps calling `gui_click` with no memory of the rules for calling it.
A checkpoint note passed as a tool-call argument has the same defect
one level later, because `RedactTool` deletes calls.

The cure is a naming rule, printed in the code as the comment on
`EntryKind`: anything that must outlive tool clearing is its own kind.

| Kind | Created by | Holds |
|---|---|---|
| `KindDialogue` | prompts, hints, model output | text, reasoning, tool calls and results |
| `KindHandoff` | `MicroHandoff` | the actor's checkpoint note |
| `KindSkill` | `SkillLoaded` | a skill's body |
| `KindTools` | `ToolsChanged` | a declaration delta |

Tool clearing only ever touches tool parts, and tool parts only ever
live in `Dialogue` entries. The other three kinds hold none, so no
redaction level can reach them. Survivors are safe by construction,
without a list of exceptions for the ladder to consult. `load_skill`
still returns a result, a one-line acknowledgement that the skill
loaded; the body arrives through the event Chapter 10 already had,
since `SkillData` always carried it.

One timing problem remains. Tools run in parallel, and a survivor
event can arrive while calls are still outstanding. A `micro_handoff`
issued beside a `read_file` clears every tool part in the context. If
it landed immediately, the `read_file` result would arrive afterwards
with its call already gone, and a result without its call is a request
the vendor refuses. So the reducer holds survivors until the batch
completes:

```go
// survivor lands a non-dialogue entry now, or holds it until the current
// batch of tool calls completes.
func (c *Context) survivor(e Entry) {
        if c.outstandingCalls() > 0 {
                c.Held = append(c.Held, e)
                return
        }
        c.land(e)
}
```

When the last result of the batch arrives, the held entries land in
order. The comment on the handoff branch of `land` states the payoff:
"Nothing is outstanding when this runs, so every call removed takes
its result with it, and no result loses its call."

Unloading a skill removes its tools through `ToolsChanged` and leaves
its `Skill` entry where it is. Deleting the body would edit old bytes,
which Law 2 prices high, and Chapter 10 already decided that unload is
lazy:

```go
LoadPendingUnload LoadState = "pending-unload"  // Marked for removal at compaction.
```

The stale manual costs its bytes and nothing else, because the tools
it describes are gone.

## §15.5 The Layout

With four kinds and two laws, the request has one shape:

```
[ frozen prefix   ]  startup tools, system prompt      never changes
[ memory region   ]  empty until the agent has memories
[ context         ]  dialogue and survivors, in Seq order
                     ^ oldest                  newest ^
                     cheap to keep            cheap to change
```

The order runs from least volatile to most, which is also the order of
meaning: who the agent is, then what it knows, then what it is doing.
Two independent arguments, one about cache economics and one about
reading order, give the same layout, which is a good sign that the
layout is right.

Survivors land in `Seq` order among the dialogue, at the tail where
they were created. Moving them up next to the prefix would read better
and would edit bytes far from the end. Reordering is a verb of its
own, and until one exists, the context is append-only everywhere
except for the ladder's cuts.

The kinds also say which channel an entry speaks on. The system prompt
and a skill body are instruction: the agent is meant to follow them. A
tool result is data: the agent reads it and decides. A dialogue entry
is conversation. Keeping those apart is why `Kind` records the origin
of an entry instead of guessing it from the text, because text that
looks like instructions and arrived as a file's contents is still a
file's contents.

## §15.6 Visible Reasoning Is the Storage Format of the Self

After the ladder runs, what does the agent remember about a file it
read an hour ago? Exactly what it said about the file. The result is a
stub, the call may be gone, and the model's private thinking was never
durable: vendors return it on their own terms, and a checkpoint drops
it. The words the agent wrote in the dialogue are the only record that
survives every level up to `RedactDialogue`. Chapter 2 wrote the rule
into the definition of `RedactTool`: remove calls and results
entirely, keep visible reasoning.

So an agent that narrates keeps its past, and an agent that works in
silence loses it. Compare two turns that make the same edit:

```
(silent)   read_file config.go   edit_file config.go

(narrated) The port is set in config.go; reading it.   read_file config.go
           Port is 8092 at line 40, hardcoded. Moving it to settings.
           edit_file config.go
```

Once the ladder reaches that turn, the silent version says an edit
happened. The narrated version still says which port, where it was,
and why it moved. A system prompt should ask for one sentence of
intent before every tool call, and the ladder turns that habit from
courtesy into storage.

Narration has a failure mode of its own. An agent rereading its old
sentences can trust a figure it wrote down over the file that has
since changed, and a figure written from memory carries the same
confidence as one copied from the source. The defence is in what the
narration records: where a fact lives, the path and line and command,
so the agent's reflex is to re-read the source rather than quote
itself.

## §15.7 The Ladder

The ladder keeps tool bytes in two bands measured back from the tail,
with one number, the target size T, setting both:

```
    stub threshold = T/100    results band = T/8    calls band = T/16
```

That line is the comment in `internal/common/budget.go`. The newest
T/8 bytes of tool traffic stay whole. Older results become stubs by
`RedactResult`, and their calls stay, for the next T/16 bytes. Tool
bytes older than both bands go entirely by `RedactTool`, leaving the
dialogue around them. At the default T of 400,000 bytes, about 100,000
tokens at four bytes a token, the stub threshold is 4,000 bytes, the
results band is 50,000 and the calls band 25,000.

Law 2 sets the rhythm. A band is cut when it reaches twice its budget,
and one event cuts it back to its budget, so cuts arrive as occasional
steps. The same comment bounds the worst case: "at its worst the tool
bytes in the window are 2(T/8 + T/16) = 3T/8, and the remaining
five-eighths are the prefix, the survivors and the dialogue".

Three rules keep the ladder honest.

- **Every cut is an event.** The policy runs in `Engine.Turn` before
  each render and emits `RedactData` through the ordinary `Record`
  path. Its `To` field is a number, a `Seq` in the log.
  The watermark is computed once, at the moment of the cut, and never
  again. Replay applies the recorded number, so lowering T tomorrow
  cannot reach back and cut yesterday's session differently.
- **Nothing is cut before the model has seen it.** A result goes out
  whole at least once, however large. The ladder only reaches entries
  already sent.
- **No call loses its result.** The calls-band cut snaps back to a
  point where every call removed takes its result with it.

The ladder applies to every model, because it needs no judgement from
the model. It is also blunt: a 40,000-byte file read that the agent
will never look at again stays whole until the band passes it.

## §15.8 Keep or Stub

A capable model can do better than the ladder, because it knows which
results it is finished with. On a model whose features row sets
`StubsToolResults`, every tool result above the stub threshold becomes
a stub in the request after the one that carried it, unless the
actor's next message calls `keep_tool_results`. The keep takes no
arguments and keeps the whole batch.

The keep governs the batch that has already arrived, the one the
actor is looking at. Called alongside other tools, it keeps the
previous batch, never the one it rides in, since those results do not
exist yet when the actor decides. Its own result is never stubbed. The
call survives either way, so the agent can see what it asked for and
re-run it. The stub is an ordinary `RedactResult` event, so replay
reproduces it without asking the model again.

Law 2 prices this well. The stub replaces bytes that arrived one round
trip ago, at the tail, which is the cheapest place in the window to
change anything. The 77 percent in §15.3 was measured under this rule:
CodeRhapsody, the agent that co-wrote this book, runs it with its own
thresholds.

The feature is a column in the model table because the judgement is
uneven across models. Asked to curate its own context, Opus 5 does it
well, Opus 4.6 acceptably, and Sonnet 5 not well enough to be given
the job. `claude-opus-5` sets `StubsToolResults`; the Sonnet rows do
not. A capability that works on one model is a claim about
that model.

> **Open risk.** Per-round-trip stubbing may interfere with the
> model's thinking: a model reasoning across several round trips has
> its evidence removed between steps. Nobody has measured it. The
> experiment that would settle it is one task, one model, stubbing on
> and off, with the outcomes compared by quality rather than by bytes.

## §15.9 From compress_context to micro_handoff

The ladder removes bytes the agent is done with. It gives the agent no
way to say what it needs in order to keep going. A checkpoint does,
and the book's own lineage tried two designs for it before this one.

CodeRhapsody's first design was a tool called `compress_context`: the
model chose a range of messages and replaced them with its own summary.
Bill described how that went on 2026-09-13:

> "compress_context was the old system that you (Claude) did pretty
> well, picking a range of messages to summarize, but freaking Gemini
> almost always deleted 80% of messages, starting with message 1, with
> a terrible summary, lobotomizing the LLM, so we switched to handoffs
> instead."

The failure has a shape. The cut ran by position, from message 1, and
message 1 is where the user stated the goal. The summary was written
by the model that was about to lose the originals, so nothing could
check it, and summarized dialogue is in no file the agent can re-read.
The model was handed a choice it could not undo, and one model family
made it badly almost every time.

The handoffs that replaced it, `handoff_task`, wrote a structured
document and started a fresh instance from it with an empty window.
That design assumes continuity lives in a note to a stranger. It also
added a second path for writing the agent's durable state, and second
paths drift: CodeRhapsody's handoff path never triggered the memory
cascade that its `save_memory` path did, and the gap sat in its notes
as a known bug for months. Ensemble builds neither tool.

`micro_handoff` keeps the same instance and the same dialogue. The
actor writes a note, and the reducer removes every tool call and tool
result in the context and appends the note as a `Handoff` entry. Two
lessons from the older designs are built in. The model's discretion
covers only what is recoverable: tool bytes can be regained by
re-running tools, and the dialogue, which cannot, stays. And the
judgement is trusted per model, as the `StubsToolResults` column
already is, never assumed to transfer from one model to the next.

On the wire it is three records: the tool call, an ordinary result
that acknowledges it, and a `MicroHandoff` event carrying the text.
The event is what the reducer acts on. The result is an
acknowledgement for the same reason `load_skill`'s is: a tool result
is tool traffic, and the note must outlive tool traffic. Replay sees
the event and makes the same cut.

What goes in the note decides whether the checkpoint works. The
version CodeRhapsody uses asks for fields, and its tool description is
blunt about which one matters. Of `tried_and_failed`, it says: "Highest
value per byte in the document: it is the only field that prevents
repeating a mistake, and errors are the part of a record most likely
to be discarded as noise." It also says when to call the tool: "at a
completed micro-goal, never at a token threshold." A checkpoint taken
mid-task drops the tool result that held the task's state. And a
checkpoint loses the model's thinking permanently, so the note has to
say what the thinking had worked out.

## §15.10 Compaction Is Described, Never Performed

No code in the agent edits the context directly. The policy decides a
cut and records it as an event; the reducer applies events; the log
keeps them. Every window the agent has ever sent can be rebuilt from
the log, including windows the current settings would cut differently.
The grader's `ladder-is-recorded` check does exactly that: it replays a
session under a changed T and expects the same requests.

A log that is replayed for months will eventually hold an event the
current reducer cannot apply: a redaction naming an entry a later
handoff removed, a payload a newer version malformed. The ch14 loader
gave up at the first one:

```go
if err := ctx.Apply(e); err != nil {
        return nil, fmt.Errorf("restore: event %d: %w", e.Seq, err)
}
```

One bad event made a whole session unloadable. The ch15 loader
reports and moves on:

```go
if err := ctx.Apply(e); err != nil && diag != nil {
        diag(fmt.Errorf("restore: skipped event %d: %w", e.Seq, err))
}
```

The skip is clean because `Apply` checks an event before it mutates
anything, so a rejected event leaves the context exactly as it was.
Totality covers known event types. An event type the code has never
heard of is still refused at load, loudly, because a log from a newer
agent is a different problem than a bad record in a familiar one.

> **Aside: why not the vendor's context editing.** The Anthropic API
> offers server-side context editing (beta header
> `context-management-2025-06-27`) that clears old tool uses and
> thinking blocks before the model reads the request. It is stateless
> and observable, and its documentation says so plainly: "Your client
> application maintains the full, unmodified conversation history.
> **You do not need to sync your client state with the edited
> version.**" So it is compatible with the log in a way the stateful
> conversation APIs of Chapter 2 were not. It still loses on three
> counts. Its policy is a declarative trigger that clears by position
> and by tool name, which cannot express a keep chosen by the actor
> each round trip. It can only clear, never summarize or move anything
> into memory. And replay would become a claim about someone else's
> deployment: the log would say what was sent, and the vendor would
> decide what was read.

## §15.11 Keep the Words, Let the Bytes Go

A stub removes bytes from the window. It removes nothing from the log.
The `ToolReturned` event that carried the result stays in the log
until log retention truncates it, and in the common case the bytes
were never unique anyway: the file is still on disk, and the command
can be run again.

So the stub carries no address. The reason is a bug in the agent that
co-wrote this book. CodeRhapsody's stubs cite the output file of the
command they replace, a path like `cr/io/26`. When this chapter was
designed, its handle counter started again at 1 on every launch, so
after a restart `cr/io/26` named a different command's output. The
agent that followed the stub read a file, got plausible output, and
had no way to tell it was the wrong one. Chapter 14's rule applies:
wrong data is worse than no data. CodeRhapsody's counter now keeps
counting across restarts. Ensemble's stubs say the bytes are gone and
leave recovery to the tools that produced them.

## §15.12 Crash-Safe Persistence

Chapter 11 saved at shutdown. A process that dies mid-session loses
everything since the last save, and a coding agent that runs builds
and debuggers dies more often than a text editor does.

The ch15 agent appends every event to a journal beside the save file,
`<save>.journal`, one JSON line per event, as the event happens.
Recovery is Chapter 11's load with one more source: `Recover` reads
the snapshot, applies the tail after its `as_of` anchor, then applies
the journal. A normal shutdown writes a new snapshot and resets the
journal. The grader's `crash-recovery` check runs a clean session,
then runs a second for two turns and kills it with SIGKILL, restarts
it, and expects nothing lost.

Two orderings are load-bearing, and neither is commutative. The new
snapshot is written before the log is truncated: in the other order, a
crash between the two steps leaves neither the old events nor the new
snapshot. And the log is never truncated past the snapshot's anchor,
because the anchor is where replay starts. `log_retention` is a
display preference; the anchor is a correctness boundary, and the
anchor wins. `SaveRetaining` does it in that order and copies the
previous save to `<save>.bak` first.

Two things are out of scope. The journal is never fsynced, so power
loss can take the last few events; "crash" means a process crash.
And truncation deletes old `ToolReturned` events, which are the only
on-disk copy of a redacted text result. The bytes were already out of
the window, and a log that keeps every byte forever is the unbounded
growth this chapter exists to stop.

## §15.13 One Knob

All of it runs off one setting. The Context Management tab in the
settings panel writes `context_target`: T in bytes, 0 for the default
of 400,000, clamped to at least 20,000. The floor has a reason in its
comment: below it, "the stub threshold would fall under the size of a
directory listing". The tab's other field, `log_retention`, is the
number of events the saved log keeps, 0 for all.

A user who wants more context pays for more context by raising T.
Every threshold scales with it, so there is nothing else to tune.

The coder measured the reference with a synthetic session: the fake
vendor scripted four prompts, each followed by fifteen `read_file`
calls on 8,000-byte files, 64 requests per run.

| run | last request | total sent | cuts |
|---|---|---|---|
| no ladder (T = 10^12) | 542,600 | 17,470,874 | none |
| ladder only (T = 400,000, Sonnet row) | 143,152 | 6,172,202 (35%) | 7 `RedactResult` |
| ladder and stubs (T = 400,000, Opus row) | 62,456 | 2,334,066 (13%) | 59 stubs |

The largest request in the ladder run was 146,378 bytes in total,
below even the 150,000 bytes that 3T/8 allows tool bytes alone. The
prefix, 2,991 bytes, was byte-identical across all 64 requests in all
three runs. The stubs run never triggered the ladder, because stubbing
kept every band under its budget.

Bytes are the easy half. The table says nothing about cache rates on a
real vendor, nothing about whether the answers got better or worse,
and nothing about the open risk in §15.8. Those need real sessions on
real models, and none has been run against this reference yet.

## §15.14 The Exercise, Graded

The grader reads only what its fake vendor receives and what the agent
leaves on disk. Every check was audited by deleting one behaviour from
the reference and confirming the check fails.

- **skill-survives-the-ladder** (15) loads a skill, runs tool traffic
  until the ladder cuts past the load, and looks for the skill body in
  the next request. The ch14 bug, body inside the tool result, scores
  85 overall and fails here.
- **micro-handoff-shape** (15) checks the three records, the text
  appearing exactly once, and no orphaned call or result, including a
  call made alongside the handoff. Clearing calls but keeping results
  fails it.
- **keep-or-stub** (10) runs the Opus row, keeping one batch and not
  the next, then runs the Sonnet row. Stubbing that ignores the keep
  fails, and so does stubbing on a model that did not ask for it.
- **ladder-is-recorded** (15) replays a saved session under a changed
  T. An agent that recomputes cuts from current settings instead of
  recording them fails this and two other checks, 55 points in all.
- **frozen-prefix** (10) compares the prefix across every request,
  through a skill load and an MCP connect. Re-declaring the startup
  tools fails it, and so does a skill load that emits no
  `ToolsChanged`.
- **replay-equals-snapshot** (15) rebuilds the context from the log and
  compares vendor requests with the saved snapshot's.
- **crash-recovery** (15) is §15.12's SIGKILL test.
- **total-reducer** (5) plants unappliable events in the middle of a
  log, with the snapshot nulled so replay must cross them, and expects
  the session to load with every good event after them applied.

The audit found two holes in the grader's first version, both of which
had let a broken agent score 100. Its test files had short paths, so
the calls band never filled and `RedactTool` never fired; the fixture
now reads long real paths and asserts both watermarks. Its bad events
sat at the end of the log, so an agent that stopped at the first bad
event passed; they now sit before the last turn.

## §15.15 Taking It for a Spin

The grader proves the rules hold on scripted traffic. A real model
shows what living under them is like. The run below drove the
reference agent with `claude-opus-5`, a row with both
`StubsToolResults` and `InlineTools` on, and a `settings.json` of one
line:

```json
{"context_target":20000}
```

Twenty thousand is the floor the settings clamp allows. It puts the
stub threshold at 200 bytes, the results band at 2,500 and the calls
band at 1,250. The workspace held seven files copied from
`agent/internal/common`, 60,865 bytes in all, three times the target.
The prompt:

```text
Read each .go file in src/ one at a time with read_file (budget.go,
context.go, save.go, journal.go, model.go, event.go, part.go). Then
answer in three sentences: how does this agent keep its context from
growing without bound?
```

The turn took eleven requests and ten tool calls: a directory
listing, seven reads, two reads of files already read, and the
answer. The journal holds nine `redacted` events, all of one shape:

```json
{"seq": 12, "type": "redacted", "time": "2026-09-23T14:44:45.308867Z", "redact": {"from": 6, "to": 6, "level": "redact_result", "reason": "round trip: not kept"}}
```

Ten results, nine stubs. The tenth result was still in its one full
request when the turn ended. The vendor's reported input tokens,
request by request:

| Request | Carried in full        | Input tokens |
|--------:|------------------------|-------------:|
| 1       | the prompt             | 3,787        |
| 2       | directory listing      | 3,955        |
| 3       | budget.go, 1,683 B     | 4,560        |
| 4       | context.go, 18,855 B   | 10,889       |
| 5       | save.go, 6,282 B       | 6,943        |
| 6       | journal.go, 4,522 B    | 6,347        |
| 7       | model.go, 7,347 B      | 8,475        |
| 8       | event.go, 10,742 B     | 10,109       |
| 9       | part.go, 11,434 B      | 10,655       |
| 10      | budget.go, again       | 7,112        |
| 11      | context.go, again      | 13,522       |

Each file is a tooth: it rises for one request and falls on the next.
The floor under the teeth rises too, because every call, every stub
and every sentence the model wrote stays. The first request, before
any file, was already 3,787 tokens of system prompt and tool
declarations. The target caps neither of those. It sets three numbers,
and all three govern tool bytes.

The ladder never fired. It never cuts a result that has not yet gone
out whole once, and the stub rule had already removed every older
result, so no request held anything the ladder was allowed to cut.
Ten calls with short paths never reached the calls band's trigger.
On a stubbing model under this workload the ladder is the backstop.
On `claude-sonnet-5`, where the stub column is off, it does all the
work.

The interesting part is the dialogue. The model narrated as it read,
and some of its narration, copied from the journal:

```text
Noting: context.go — Context has bounded fields; redaction/summary/handoff clear tool parts. Next file.

model.go: a per-model feature table; the context-relevant column is `StubsToolResults` — "every tool result above the stub threshold becomes a stub in the request after the one that carried it, unless the model's next message calls keep_tool_results" (on for opus-5/opus-5-course; off for sonnet-5, where "the ladder" alone applies). Next file.

part.go: content is `[]Part`, and `RedactedPart` is the *result* of a redaction — its stub is synthesized deterministically by the reducer (no stored growth) and it carries forward the superseded part's `Ref`, so dropped bytes stay fetchable. Now re-reading the two files whose contents got stubbed out of my own context.

budget.go confirmed: one knob T, with threshold=T/100, results band=T/8, calls band=T/16. Now the policy that spends them.
```

The two files it read again were `budget.go`, the one file it wrote
no note about, and `context.go`, the largest file, which got one line.
Every file with a real note was never read again. The note is the
model's own compression of a result, and dialogue is the channel the
stub rule leaves alone, so what the model wrote down, it kept. What
it did not write down, it fetched again.

`keep_tool_results` was declared in every request and never called.
That is not a failure. A kept result is paid for on every later
request; a second read is paid for once, plus one round trip. This
run cost two round trips. One run on one model settles nothing about
the open risk printed earlier, whether stubbing interferes with
thinking, but it does show the shape a capable model falls into:
narrate, let the bytes go, re-read what the narration missed.

Two things in the log belong on the next revision's list. Every
request reported zero cache writes and zero cache reads: the
reference renderer sets no `cache_control` breakpoint, so the stable
prefix this chapter guards is ready for a cache the agent never asks
for. And the rising floor is dialogue, which nothing in this chapter
removes. Chapter 16 removes it.
