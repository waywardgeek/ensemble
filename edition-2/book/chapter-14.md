# Chapter 14: The World's Best Context Engineering, Before Breakfast

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

A long session with a coding agent can degrade in a familiar way. The
answers get vaguer. The agent re-reads a file it read an hour ago, then
forgets a decision made before lunch. A fresh chat helps because a
human decides what to carry over. This chapter moves tool-traffic
cleanup inside the agent and records every cut as an event, so nothing
leaves the window without a record of what removed it.

## TL;DR

Tool bytes gain a removal policy. Every entry gets
a kind that says which verb removes it. Tool bytes fall off a ladder of
recorded redaction events. On a model configured for stubbing, each round
trip's results are stubbed unless the model keeps them. The actor can
checkpoint with `micro_handoff`. Loaded skills survive as entries instead
of as tool results. The system prompt stays frozen. The log reaches disk
as it grows, so a process crash can recover the journaled work.

The shared entry vocabulary gains four kinds. These definitions belong in
`internal/common`; reduction belongs in History, not on common-package structs:

```go
// EntryKind identifies what removes an entry. Survivors never contain tool parts.
type EntryKind uint8

const (
    KindDialogue EntryKind = iota + 1
    KindHandoff
    KindSkill
    KindTools
)

// MicroHandoffData records the checkpoint text that survives its tool call.
type MicroHandoffData struct {
    Text string `json:"text"`
}
```

Each context entry carries its kind alongside sequence, actor and parts. Reuse
the existing skill body and redaction payloads. Add `micro_handoff` and
`tools_changed` facts; the latter carries full `added` declarations and `removed`
names. A `tool_decls` part holds that delta without becoming a call or result.

1. **Frozen prefix.** The system prompt is byte-identical on every request
   of a session. Only a refreshed Agent changes it. Startup tool declarations
   stay frozen too on a model that can carry tool changes in the dialog;
   other models re-declare when tools change, as rule 2 describes.
2. **Tools arrive through the dialog.** A skill load, a skill unload,
   or an MCP connect mid-session emits `ToolsChanged`, and the reducer
   turns it into a `Tools` entry. On a model whose features row sets
   `InlineTools`, that entry carries the declarations in the dialog.
   The `tool_decls` part is neither a call nor a result, so rule 3 still
   holds. The renderer places inline declarations after the complete
   user/tool-result/ephemeral turn. A model without `InlineTools` folds
   every delta into the startup set and re-declares, changing the prefix.
3. **Survivors carry no tool parts.** A `Handoff`, `Skill` or `Tools`
   entry never holds a tool call or a tool result, so no tool clearing
   can touch it.
4. **Skills are entries.** The reducer turns `SkillLoaded` into a
   `Skill` entry holding the rendered body. The `load_skill` tool result
   is an acknowledgement only. Replay uses the recorded body, without
   rereading the file. `unload_skill` removes grants no other active skill
   supplies and stops its servers, but leaves the `Skill` entry where it
   is: deleting old bytes would change the prefix. A later chapter's verb
   removes it.
5. **`micro_handoff` is three records:** the tool call, an ordinary
   result, then a `MicroHandoff` event. Its reducer waits until the
   batch's last result has arrived, so a call made alongside
   `micro_handoff` is cleared with it, even if it has not started yet in
   the serial dispatcher. Then it removes every tool call and tool
   result part, opaque reasoning and private replay signatures from the context,
   including signatures attached to visible text. It preserves visible dialogue
   and the event log, drops entries left empty, and appends one `Handoff` entry.
   The text appears in the next request exactly
   once, and no call is ever left without its result or a result without
   its call. The tool takes `{"text":"..."}`.
6. **Keep or stub, per round trip.** On a model whose features row sets
   `StubsToolResults`, every tool result above the stub threshold is
   replaced by a stub in the request after the one that carried it,
   unless the actor's next message calls `keep_tool_results({})`, which
   keeps that whole batch. Called alongside other tools, it keeps the
   previous batch, not the one it rides in, and its own result is never
   stubbed. The call survives either way. The stub is an ordinary
   `RedactResult` event, so replay reproduces it. Other models get no
   per-round-trip stubbing; the ladder alone applies.
7. **The ladder records a Seq.** With target size T, whole tool traffic
   has a T/8 budget. Older results become stubs (`RedactResult`) and
   their calls stay, for a T/16 budget of calls and stubs. Older tool
   traffic goes entirely (`RedactTool`). A band is cut in steps: when it
   passes twice its budget, one event cuts an old prefix back toward its
   budget. Cut only completed batches the model has already seen. Each
   event stores `To` as a number, never as "the watermark", so a later
   settings change cannot rewrite the past.
8. **The reducer is total over known events.** A save file that does not
   parse is still refused (Chapter 10). A familiar event that parses but
   cannot be applied, such as a malformed payload or a redaction naming
   an entry already gone, is skipped with a diagnostic and loading
   continues. Validate before mutation; unknown event types and invalid
   sequence ordering remain errors.
9. **Crash-safe.** Events reach the journal before publication, so a
   process crash leaves a tail after Chapter 10's `as_of` anchor.
   Recovery loads the snapshot, its tail and the journal. A normal
   shutdown snapshots. Write the new snapshot before truncating the
   journal, never truncate past the anchor, and keep one backup of the
   previous snapshot. Power loss is out of scope; the log is never fsynced.
10. **One knob.** The Context Management settings tab sets
    `context_target`: T in bytes, 0 for the default of 400,000, clamped
    to at least 20,000. The stub threshold is T/100 and the bands are
    rule 7's, so the default gives 4,000, 50,000 and 25,000. The tab's
    other setting, `log_retention`, is the number of events kept in the
    saved log, 0 for all. Store settings at the owning Agent's path and
    preserve unsent browser edits when a settings snapshot arrives.

The model-feature matrix is part of the exercise:

| Row | `InlineTools` | `StubsToolResults` |
| --- | --- | --- |
| `claude-opus-5-course` | true | true |
| `claude-sonnet-5-course` | false | false |
| All existing real-model rows | false | false |

The course rows are deterministic fixtures. Keep both new flags false on
existing real-model rows. Dialogue and survivors remain outside the tool
budgets; this chapter does not yet bound the whole window.

For live Anthropic use, choose an older supported model such as Sonnet-5.
Bill reports that 5.5 models do not work with ephemeral context and has
excluded them from supported use. Historical 5.5 runs below retain their
actual names; they are not reader model choices.

Yours: the on-disk layout of the journal, what a diagnostic says, and the
settings tab's appearance.

**Exercise.** Start from your Chapter 13 agent. Add context management until
the following command, run from the repository root, scores 100/100:

```sh
go run ./cmd/grade-edition2 -ch 14
```

The default core is `solutions/ch14`; a positional core directory and
existing `-gui` option select another implementation. Section 14.14 gives the
actual launch, settings and eight-check contract.

## §14.1 In Plain Words

The context window is not storage. The log is storage: Chapter 2 made
it append-only, recording what the agent saw or said. The window is a
working set rendered from that log, sent again on every request. Every
byte has to answer three questions: why is it here, who put it here, and what removes it. A byte
that cannot answer the third question stays forever, and a window full
of such bytes can crowd out the work a person wanted the agent to remember.

Most of the bytes in Chapter 2's measured sessions were tool bytes:
results were about 42 percent of conversation history by volume, and
call arguments another 30 percent. Nearly all of those bytes are needed
once. The agent reads a file, decides, edits; after that, the file's
contents are recoverable from disk and the decision lives in what the
agent said about it. Hence the chapter's rule: keep the words, let the
bytes go. Dialogue stays. Tool results go first, then the calls that
produced them, oldest first.

Some things must never go with them. A loaded skill's instructions and
the note an agent writes to its future self at a checkpoint look like
tool traffic today, because they arrive through tools. If they stay
tool traffic, the first cleanup deletes the manual and keeps the tools
it explains. Anything that must survive becomes its own kind of
entry, removed only by its own verb. Survivors are safe by
construction.

Every cut is an event in the log with an exact number in it; nothing is
re-evaluated later. That is what keeps Chapter 10's promise:
replaying the log reproduces the same window, byte for byte, even after
the settings that chose the cuts have changed.

Two laws decide where cuts may happen. The front of the request, the
system prompt and, on an inline-capable model, the startup tools, is
frozen so a prefix cache can keep serving it. Changing a byte near the
front invalidates more of that prefix than changing one near the end.
So cuts come in steps, not a trickle: one larger cut now and then
invalidates that region less often than a small one every turn.

The redaction levels are Chapter 2's, the frozen prompt is Chapter 9's,
the snapshot and replay are Chapter 10's. What was missing is the
policy that decides when to use them, and kinds for the entries that
must survive.

## §14.2 Why Is This Byte Here?

Chapter 2 built the machinery for this chapter before it had a cleanup
policy to call it. `RedactData` has four levels. `RedactResult` turns a
tool result into a stub and keeps the call. `RedactTool` removes both
and keeps the reasoning around them. `RedactDialogue` and
`RedactSummary` reach into the conversation itself. The first edition
left a comment at that boundary:

```go
case RedactDialogue:
        // Prose and reasoning go. Survivors are defined by the compaction
        // policy, which is a later chapter's problem.
```

This chapter supplies the tool-traffic policy. Dialogue removal remains a later exercise. Context engineering, as the book uses the term,
is managing every byte in the context data structure as well as
today's models allow. The management is a policy: for each byte, a
reason it is in the window, a record of what put it there, and a verb
that takes it out.

Walk a request from a Chapter 13 agent after an hour of work and ask each
byte the three questions. The system prompt answers all three: the
agent's constitution, written at startup, removed by nothing. A user
prompt answers them. A 9,000-byte file read forty minutes ago answers
the first question with "the agent needed it once", the second with
`read_file`, and the third with silence. Nothing removes it. It rides
along on every request until the session ends or a human copies the
good parts into a fresh chat.

The fix is that silence. Give the old tool traffic a remover.

## §14.3 Two Laws

The vendor's prompt cache decides where a cut may happen. Under prefix
caching, everything after the first changed position loses its match.
This describes reuse, not a bill: bytes are not tokens, and the provider
must actually grant caching for the stable prefix to save money.

**Law 1: the prefix is frozen.** The system prompt and, on an inline-capable
model, the startup tool declarations are byte-identical on every request
of a session. Chapter 9 already made the system prompt a constitution and put dynamic
skills in the dialog. It left one leak: a skill that brings tools
changed the tool list, and the tool list is the first thing in the
request. Loading one skill could invalidate reuse from the tool array onward.

The chapter closes the leak with an event. A skill load, a skill
unload, or an MCP server connecting mid-session emits `ToolsChanged`,
a delta of declarations added and names removed. The reducer turns it
into a `Tools` entry at the tail of the context, where it preserves the earlier declaration prefix. On a model whose features row sets
`InlineTools`, the renderer sends that entry as a mid-conversation
system message:

```json
{"role": "system",
 "content": [{"type": "tool_addition",
              "tool": {"type": "tool_definition",
                       "definition": {"name": "gui_click",
                                      "description": "Click a matching element",
                                      "input_schema": {"type": "object",
                                                       "properties": {"selector": {"type": "string"}},
                                                       "required": ["selector"]}}}}]}
```

The course renderer sends the header `inline-tools-2026-09-15`. A removal
uses a `tool_removal` block containing
`{"type":"tool_reference","name":"gui_click"}` under `tool`. These dated
wire shapes belong to the course fixture; the exercise does not enable
them on live models. A model without the feature gets the old behaviour:
the renderer folds every delta into the startup set and re-declares,
changing the prefix.

The declaration's position matters as much as its bytes. During the student run,
a new MCP declaration followed a tool result, then automatic ephemera added
another user message after it. The roles became:

```text
user, assistant, user, system, user
```

The initial fake accepted that request, and prefix and replay comparisons passed.
A faithful placement check rejected it: the contentful system message belongs
after the complete user turn. The corrected renderer holds declaration blocks
until the assistant boundary or request end, after tool results and automatic
context:

```text
user, assistant, user, system
```

The fix changes only the disposable wire view. The log keeps its chronology,
and ephemera remain user content. Equal replay bytes are useful only after
the request itself is valid.

The prefix does change on purpose once in a while. New defining
instructions require a refreshed Agent. Changing `context_target` leaves
the instructions and already recorded cuts alone.

**Law 2: the cost of changing a byte grows with its distance from the
end.** Everything after the changed position loses its prefix match.
Stubbing the result that arrived on the last round trip changes bytes
near the tail. Stubbing a result a hundred thousand bytes back
invalidates a hundred thousand bytes of later content.

The design record preserves a useful failed prediction:
stubbing every tool result one round trip after it arrived would wreck
the cache, since the request changes on every trip. The first session
run that way, on CodeRhapsody with Opus 5.5 on 2026-09-22, reported a
cumulative cache hit rate of 77 percent, after a cold first request
that missed on over 100,000 tokens. The prediction had the law
backwards: a just-finished result sits at the tail, where less later content loses its prefix match. The cumulative figure also hides the steady state, because that
one cold miss is averaged into every later request, which is why a
display of cache rate should show the last request beside the total.

Law 2 has a second consequence. A trickle of small cuts deep in the
window breaks that prefix match on every turn; one larger cut now and then
breaks it once. The agent cuts in steps.

## §14.4 What Survives Is Never a Tool Call

The Chapter 13 agent returns a loaded skill's manual as the result of
the tool that loaded it. The first edition wrote it this way:

```go
result += "## Instructions\n\n" + body
```

The arrangement becomes a bug the moment cleanup can discard that result. Tool results are the first bytes
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
live in `Dialogue` entries. The other three kinds hold none, so tool-part clearing cannot remove their contents. Survivors are safe by construction,
without a list of exceptions for the ladder to consult. `load_skill`
still returns a result, a one-line acknowledgement that the skill
loaded; the body arrives through its own recorded `skill_loaded` fact.

One timing problem remains even though dispatch is serial. The model can return several calls together; after the first runs, others from the same reply are still outstanding. A `micro_handoff`
issued beside a `read_file` clears every tool part in the context. If
it landed immediately, the `read_file` result would arrive afterwards
with its call already gone, and a result without its call is a request
the vendor refuses. So the reducer holds survivors until the batch
completes:

```go
// Holding survivors preserves call/result pairing even under serial dispatch:
// another call from this same model reply may still be waiting to start.
func survivor(c *common.Context, entry common.Entry) {
    if !toolsComplete(c.Dialogue) {
        c.Held = append(c.Held, entry)
        return
    }
    land(c, entry)
}
```

When the last result of the batch arrives, the held entries land in
order. The payoff: nothing is outstanding when the handoff clears tool
parts, so every call removed takes its result with it, and no result
loses its call.

Unloading changes executable capability immediately. The registry drops grants
that no other active skill supplies and stops the skill's external servers.
The next recorded declaration delta makes the model's view agree with dispatch.
Its instruction entry remains: removal of old instruction text is a different
operation, deferred beyond this chapter. Lifecycle and checkpoint tools stay
available, and a shared grant survives while another active skill owns it.

This advances the pending-unload policy from Chapter 9: compiled grants
now disappear too. A stale manual cannot make a removed tool callable;
guessed names still fail dispatch.

## §14.5 The Layout

With four kinds and two laws, the request has one shape:

```
[ frozen prefix   ]  system prompt; startup tools on inline-capable rows
[ memory region   ]  empty until the agent has memories
[ context         ]  dialogue and survivors, with recorded Seq identities
                     ^ oldest                  newest ^
                     cheap to keep            cheap to change
```

The order runs from least volatile to most, which is also the order of
meaning: who the agent is, then what it knows, then what it is doing.
Two independent arguments, one about cache economics and one about
reading order, give the same layout, which is a good sign that the
layout is right.

A survivor keeps the sequence identity of its recorded fact. If a tool batch
is incomplete, the reducer holds it and appends it only after the companion
results arrive, even when those results have later sequence numbers. That
batch-completion placement preserves call/result pairing. Moving the survivor
up next to the prefix would edit bytes far from the end. Reordering is a
verb of its own.

The kinds also preserve why an entry belongs in context. The system prompt
and a skill body are instruction: the agent is meant to follow them. A
tool result is data: the agent reads it and decides. A dialogue entry
is conversation. Keeping those apart is why `Kind` records the origin
of an entry instead of guessing it from the text, because text that
looks like instructions and arrived as a file's contents is still a
file's contents.

## §14.6 Visible Reasoning Is the Storage Format of the Self

After the ladder runs, what does the agent remember about a file it
read an hour ago? Exactly what it said about the file. The result is a
stub and the call may be gone. The words the agent wrote in the dialogue
remain a readable account through tool clearing. Ordinary tool redaction also
preserves opaque provider parts under their replay restrictions. A `micro_handoff`
additionally clears that private replay material and its signatures while keeping
visible text. Chapter 2 wrote the rule
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

Once both tool bands clear that turn, the silent version may retain no explanation of the edit. The narrated version still says which port, where it was,
and why it moved. A system prompt should ask for one sentence of intent
before every tool call, and the ladder turns that habit from courtesy
into storage.

Narration has a failure mode of its own. An agent rereading its old
sentences can trust a figure it wrote down over the file that has
since changed, and a figure written from memory carries the same
confidence as one copied from the source. The defence is in what the
narration records: where a fact lives, the path and line and command,
so the agent's reflex is to re-read the source rather than quote
itself.

## §14.7 The Ladder

The ladder keeps tool bytes in two bands measured back from the tail,
with one number, the target size T, setting both:

```
    stub threshold = T/100    results band = T/8    calls band = T/16
```

The Agent reads T from its owned configuration. Its policy measures serialized tool parts, excluding ordinary dialogue and survivor bodies. The newest
T/8 bytes of tool traffic stay whole. Older results become stubs by
`RedactResult`, and their calls stay, for the next T/16 bytes. Tool
bytes older than both bands go entirely by `RedactTool`, leaving the
dialogue around them. At the default T of 400,000 bytes, the stub threshold is 4,000 bytes, the
results band is 50,000 and the calls band 25,000.

Law 2 sets the rhythm. A band is cut when it reaches twice its budget,
and one event cuts it back to its budget, so cuts arrive as occasional
steps. The two doubled budgets sum to 3T/8. That arithmetic describes the eligible tool bands, not a hard maximum for the whole request. A newly presented large result must survive once, whole batches constrain cut boundaries, and dialogue and survivors have no budget here.

Three rules keep the ladder honest.

- **Every cut is an event.** The Agent runs policy before rendering and records `redacted` facts through its existing append path. History alone changes the projection; Engine renders it. Its `To` field is a number, a `Seq` in the log.
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

## §14.8 Keep or Stub

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
change anything. The 77 percent in §14.3 was measured under this rule:
CodeRhapsody, the agent that co-wrote this book, runs it with its own
thresholds.

The feature is a column in the model table because deciding which
results to discard is a model's judgement. Whether it chooses well
depends on the work it can still do and the rereads it needs.
The course Opus row lets the grader exercise keep-or-stub; real rows
use the ladder alone.

> **Open risk.** Per-round-trip stubbing may interfere with the
> model's thinking: a model reasoning across several round trips has
> its evidence removed between steps. The retained experiments did not measure
> that effect on reasoning quality. A useful comparison would hold the task and
> model fixed, run with stubbing on and off, and assess the work rather than bytes.

## §14.9 From compress_context to micro_handoff

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
In those trials, the model was handed a choice it could not undo and
kept making it badly.

The handoffs that replaced it, `handoff_task`, wrote a structured
document and started a fresh instance from it with an empty window.
That design assumes continuity lives in a note to a stranger. It also
added a second path for writing the agent's durable state, and second
paths drift: CodeRhapsody's handoff path never triggered the memory
cascade that its `save_memory` path did, and the gap sat in its notes
as a known bug for months. Ensemble builds neither tool.

`micro_handoff` keeps the same instance and visible dialogue. The actor writes a
note; after the complete tool batch, the reducer removes calls, results, opaque
reasoning and private replay
signatures, including signatures attached to visible text, then appends the note
as a `Handoff` entry. The event log preserves the original records; this clears
the working context. Two lessons from the older designs are built in. The model's
discretion covers tool bytes that can often be regained by re-reading their sources;
the dialogue, which cannot, stays. And the note travels with the same
conversation, instead of becoming a stranger's only account of it.
`micro_handoff` is available on every model row.

In the log it is three records: the tool call, an ordinary result
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
mid-task drops the tool result that held the task's state. The note has
to say what the agent worked out and which failed approaches it should
avoid repeating.

## §14.10 Compaction Is Described, Never Performed

No policy code edits context directly. The policy decides a cut and
records it as an event; the reducer applies events; the log keeps them.
While the full log is retained, every window the agent has sent can be
rebuilt, including windows the current settings would cut differently.
The grader's `ladder-is-recorded` check does exactly that: it replays a
session under a changed T and expects the same requests.

A retained log can contain a known event the current reducer cannot apply: a redaction naming an entry a later
handoff removed, a payload a newer version malformed. Previously, one
bad event made a whole session unloadable. History now reports and
moves on:

```go
func (h *history) replay(e common.Event) {
    if err := validateFor(&h.context, e); err != nil {
        h.parent.Ensemble().Logf("restore: skipped event %d: %v", e.Seq, err)
        return
    }
    apply(&h.context, e)
}
```

The skip is clean because validation checks an event before it mutates
anything, so a rejected event leaves the context exactly as it was.
Totality covers known event types. An event type the code has never
heard of is still refused at load, loudly, because a log from a newer
agent is a different problem than a bad record in a familiar one.

The original design also considered provider-side context editing. Its dated
API documentation described a server editing the input while the client kept its
full history. That can coexist with a local log, but it moves the decision about
what the model reads outside this reducer. Here, the chapter needs a recorded
keep decision and reproducible local projection, so the existing event path owns
the cut.

## §14.11 Keep the Words, Let the Bytes Go

A stub removes bytes from the window. It removes nothing from the log.
The `ToolReturned` event that carried the result stays in the log
until log retention truncates it, and in the common case the bytes
were never unique anyway: the file is still on disk. Re-read the source
when it is needed again; a changed file or a command with side effects
may of course give a different result.

The stub carries no address. The reason is a bug in the agent that
co-wrote this book. CodeRhapsody's stubs cite the output file of the
command they replace, a path like `cr/io/26`. When this chapter was
designed, its handle counter started again at 1 on every launch, so
after a restart `cr/io/26` named a different command's output. The
agent that followed the stub read a file, got plausible output, and
had no way to tell it was the wrong one. Chapter 12's rule applies:
wrong data is worse than no data. The first edition records the fix:
CodeRhapsody's counter kept counting across restarts. Ensemble's stubs
say the bytes are gone and
leave recovery to the tools that produced them.

## §14.12 Crash-Safe Persistence

Chapter 10 saved at shutdown. A process that dies mid-session loses
everything since the last save, including work whose results have already reached the model.

The current Agent appends each accepted event to a journal beside the save file,
`<save>.journal`, one JSON line per event, as the event happens.
Recovery is Chapter 10's load with one more source: `Recover` reads
the snapshot, applies the tail after its `as_of` anchor, then applies
the journal. A normal shutdown writes a new snapshot and resets the
journal. The grader's `crash-recovery` check runs a clean session,
then runs a second for two turns and kills it with SIGKILL, restarts
it, and checks that journaled work survives the restart.

Two orderings are load-bearing, and neither is commutative. The new
snapshot is written before the log is truncated: in the other order, a
crash between the two steps leaves neither the old events nor the new
snapshot. And the log is never truncated past the snapshot's anchor,
because the anchor is where replay starts. `log_retention` is a saved-history preference; the anchor is a correctness boundary, and the
anchor wins. History installs its checkpoint in that order and keeps
the previous save at `<save>.bak`, using the Agent's selected save path.

Two things are out of scope. The journal is never fsynced, so power
loss can take the last few events; "crash" means a process crash.
And truncation deletes old `ToolReturned` events, which are the only
on-disk copy of a redacted text result. The snapshot can still restore the current window, but retention may make older windows and raw results unavailable. Zero retention keeps the full saved log; it does not promise bounded disk growth.

## §14.13 One Knob

The tool budgets run off one setting. The Context Management tab in the
settings panel writes `context_target`: T in bytes, 0 for the default
of 400,000, clamped to at least 20,000. The original budget comment explains its floor: below it, "the stub threshold would fall under the size of a
directory listing". The tab's other field, `log_retention`, is the
number of events the saved log keeps, 0 for all.

The GUI's file is `DataDir/settings.json`; the supplied application chooses
`cr`, so use `cr/settings.json` from its launch directory. A file named
`settings.json` beside the executable is not interchangeable. The browser's
public settings update reaches the owning Agent before subsequent policy work.

A user changes the tool-traffic target, not a ceiling on every kind of context.
Every threshold scales with T. `log_retention` separately controls saved history;
changing either setting cannot rewrite already recorded cuts.

The settings browser also exposed an editing race: an acknowledgement for one
field arrived while another number was typed but not yet committed by blur.
Replacing every field from the snapshot erased that unsent edit. Keep effective
settings updates while protecting the person's pending input.


The first-edition coder measured a synthetic reference session: the fake
vendor scripted four prompts, each followed by fifteen `read_file`
calls on 8,000-byte files, 64 requests per run.

| run | last request | total sent | cuts |
|---|---|---|---|
| no ladder (T = 10^12) | 542,600 | 17,470,874 | none |
| ladder only (T = 400,000, Sonnet row) | 143,152 | 6,172,202 (35%) | 7 `RedactResult` |
| ladder and stubs (T = 400,000, Opus row) | 62,456 | 2,334,066 (13%) | 59 stubs |

The largest request in the ladder run was 146,378 bytes in total,
below even the 150,000 bytes allocated to the two doubled tool bands. The
prefix, 2,991 bytes, was byte-identical across all 64 requests in all
three runs. The stubs run never triggered the ladder, because stubbing
kept every band under its budget.

Bytes are the easy half. The table says nothing about cache rates on a
real vendor, nothing about whether the answers got better or worse,
and nothing about the open risk in §14.8. Those need real sessions on
real models, such as the one in §14.15.

## §14.14 The Exercise, Graded

Extend the current core and optional GUI. The Agent chooses policy and follows
its configuration; History validates, journals and reduces; Engine translates the
result into vendor requests. Keep those existing ownership paths rather than
injecting a settings callback or moving runtime behavior into common.

The grader drives the actual GUI executable through newline-delimited stdin
prompts, using the existing envelope:

```json
{"kind":"prompt","text":"Load manual and inspect the prepared files."}
```

It supplies provider environment, the two course model rows, complete skills and
GUI assets. Public settings updates establish the effective target before traffic;
the GUI persists them at its own Agent path. Normal EOF drains work, snapshots and
exits successfully; the crash scenario observes completed work before SIGKILL.

| Check | Points | Observable obligation |
| --- | ---: | --- |
| skill-survives-the-ladder | 15 | The manual remains after both tool bands cut old paired traffic |
| micro-handoff-shape | 15 | Call/result/event, one note, no orphan including a same-batch companion |
| keep-or-stub | 10 | Previous-batch keep, next-batch distinction and non-stubbing model control |
| ladder-is-recorded | 15 | Recorded numeric cuts survive restart under changed T |
| frozen-prefix | 10 | Stable inline prefix, actual skill/MCP declarations and valid placement |
| replay-equals-snapshot | 15 | Snapshot and sufficient log replay produce the same next vendor bytes |
| crash-recovery | 15 | Completed journaled work survives an actual killed process |
| total-reducer | 5 | Bad known facts are skipped; later good facts still apply |

The eight checks total 100 points. Also check the behaviors a transcript
cannot show: diagnostics and validation before mutation, disk-failure
ordering and backups, and unsent edits in the settings browser.

The student's original score was **45/100 FAIL**, with failures caused by
fixture mistakes. Review separately found the inline-placement bug in §14.3.
The revised grader
sets the effective Agent-owned settings and uses the existing
`list_directory` tool instead of undefined `think`. The [chapter notes](../docs/edition-2-notes/ch15.md)
retain the original result and the corrected checks.

The first-edition audit found two other useful holes. Short file paths did not
fill the calls band, so a supposed two-band test reached only result stubbing.
Bad events at the very end let a loader stop early without losing any later good
content. Exercise both watermarks and put valid work after a rejected event.
The student's inline-ordering failure supplies the same lesson at another
boundary: replay can faithfully reproduce an invalid provider request.

## §14.15 Taking It for a Spin

The grader exercises the rules on scripted traffic. A real model shows
what living under them is like. The first edition records this September
23, 2026 run with `claude-opus-5` and both new features enabled, a different
configuration from this edition's real-model rows. Its settings were:

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
answer. The journal held nine `redacted` events, all of one shape:

```json
{"seq": 12, "type": "redacted", "time": "2026-09-23T14:44:45.308867Z", "redact": {"from": 6, "to": 6, "level": "redact_result", "reason": "round trip: not kept"}}
```

Ten results, nine stubs. The tenth result was still in its one full
request when the turn ended. The reported input tokens, request by request:

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
On a row without automatic stubbing, the ladder does the tool-byte cleanup.

The interesting part is the dialogue. The model narrated as it read;
some of that narration, copied from the journal:

```text
Noting: context.go — Context has bounded fields; redaction/summary/handoff clear tool parts. Next file.

model.go: a per-model feature table; the context-relevant column is `StubsToolResults` — "every tool result above the stub threshold becomes a stub in the request after the one that carried it, unless the model's next message calls keep_tool_results" (on for opus-5/opus-5-course; off for sonnet-5, where "the ladder" alone applies). Next file.

part.go: content is `[]Part`, and `RedactedPart` is the *result* of a redaction — its stub is synthesized deterministically by the reducer (no stored growth) and it carries forward the superseded part's `Ref`, so dropped bytes stay fetchable. Now re-reading the two files whose contents got stubbed out of my own context.

budget.go confirmed: one knob T, with threshold=T/100, results band=T/8, calls band=T/16. Now the policy that spends them.
```

The model's claim that a `Ref` keeps dropped bytes fetchable is wrong
for this exercise. Its stubs carry no address, for the reason in §14.11.

The two files it read again were `budget.go`, the one file it wrote
no note about, and `context.go`, the largest file, which got one line.
Every file with a fuller note was never read again in this run. The note
is the model's own compression of a result, and dialogue is the channel the
stub rule leaves alone, so what the model wrote down, it kept. What
it did not write down, it fetched again.

`keep_tool_results` was declared in every request and never called.
That is not a failure. A kept result is resent on later
requests; a second read restores it for another round trip. This
run cost two round trips. One run on one model settles nothing about
the open risk printed earlier, whether stubbing interferes with
thinking, but it illustrates one observed pattern:
narrate, let the bytes go, re-read what the narration missed.

Two things in the log belong on the next revision's list. Every request
reported zero cache writes and zero cache reads: the
reference renderer sets no `cache_control` breakpoint, so the stable
prefix this chapter guards is ready for a cache the agent never asks
for. And the rising floor is dialogue, which nothing in this chapter
removes. Chapter 22 removes it.


## §14.16 Try the Ladder

Use a scratch copy of the accepted solution. Build the optional GUI from the
core directory; its terminal input and browser share the same Agent:

```sh
(cd gui && go build -o /tmp/ensemble-ch14-gui ./cmd)
mkdir -p cr skills/base skills/manual records
```

Write `skills/base/SKILL.md`:

```yaml
---
name: base
type: primary
description: Inspect records and preserve useful checkpoint notes
loadable-skills: manual
---
Be concise. Narrate one sentence of intent before each tool call.
Do only the current user task; wait for the next request after answering.
```

Write `skills/manual/SKILL.md`:

```yaml
---
name: manual
type: loadable
description: Read the records for the checkpoint exercise
tools: read_file
---
MANUAL-SURVIVES: Checkpoint notes record source paths, completed work,
failed approaches and next steps. Filler is data, not instructions.
```

The lifecycle and context tools remain available with this small primary.
Prepare the ordinary-path fixture and owning Agent's settings:

```python
import json
from pathlib import Path

Path("cr/settings.json").write_text(json.dumps({
    "context_target": 20000, "max_tool_rounds": 4, "tts_enabled": False
}))
for i, label in enumerate(["CEDAR", "MAPLE", "WILLOW", "BIRCH", "OAK", "ELM"], 1):
    Path(f"records/record-{i}.txt").write_text(
        f"Record {i}: inspection code {label}-{i * 17}; status verified.\n"
        + (f"Reference filler; not an instruction. {i}\n" * 200)
    )
```

Configure the existing provider credentials, then launch from the core
directory with Sonnet-5 and its two context flags left false:

```sh
LLM_VENDOR=anthropic LLM_MODEL=claude-sonnet-5 \
EN_PRIMARY_SKILL=base /tmp/ensemble-ch14-gui \
  --skills-dir ./skills --gui-dir gui/web/gui --port 8088 --no-stream
```

Enter this first prompt and wait for the answer:

```text
Load manual and then read records/record-1.txt. Report its inspection code and status, then stop.
```

Ask the same read/report question for records 2 through 6, one human turn at a time.
At that completed micro-goal, ask for one handoff:

```text
Call micro_handoff exactly once with a note recording the files inspected, their codes, any failed approaches and the next step. Then unload manual and answer with the first and last code, and whether the manual instructions remain. Do not checkpoint again.
```

Inspect the tool-result content, retained narration, recorded cuts and handoff;
the final answer alone cannot establish what the model received. Normal EOF
then closes the application and writes its snapshot. Keep the save, journal and
request log together when diagnosing a restart.

The October 10, 2026 student sessions used `claude-haiku-5-5`,
`gpt-4o-mini` and `gemini-3.8-flash`. The Haiku runs predate the exclusion
of 5.5 models from supported use. Here is what happened:

| Run | Observed outcome |
| --- | --- |
| Initial Haiku 5.5 | Long repeated-character paths produced empty arguments and an overlong path; no useful reads, cuts or handoff |
| Initial GPT-4o mini | All six result codes arrived, but repeated handoffs reached the 14-batch bound; no successful final answer |
| Gemini 3.8 | Six reads, five result cuts, two paired tool cuts, one handoff and unload; correct first/last codes and recognition that the manual remained |
| Haiku 5.5 with ordinary paths and separate turns | All six full results arrived, then six result cuts and one tool cut; later denied having read the redacted files and refused the handoff |
| GPT-4o mini with ordinary paths and separate turns | All six full results arrived, six result cuts and one tool cut; handoff/unload ran, but repeated checkpoints reached the final turn's four-batch bound |

All five processes exited zero. Only Gemini completed the requested task;
the [full results](../docs/edition-2-notes/ch15.md#live-use-and-its-limits)
include the failed attempts.

Haiku 5.5's corrected-path run shows a failure worth inspecting. It first received the codes
in actual provider-bound result blocks. After those results were redacted, it said:

> That file was redacted, so I never read it.

The transcript disproves the claim. The ladder had removed already-presented
bytes; it had not prevented the read. This is a failure of the tested Haiku 5.5,
not a result for Sonnet-5. A smaller request is easy to measure. Check that the
model can still do the work with what remains.
