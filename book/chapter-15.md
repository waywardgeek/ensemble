# Chapter 15: Keep the Words

Every long session with a coding agent ends the same way. The answers
get vaguer. The agent re-reads a file it read an hour ago, then forgets
a decision made before lunch. Eventually the human gives up on the
conversation, copies out the parts worth keeping, pastes them into a
fresh chat, and starts again. That manual reset is how most people
manage context today, including the people who build these agents:
Bill did it several times on the day this chapter was designed. The
reset works because a human decides what to keep. This chapter moves
that decision inside the agent, makes it continuous instead of
catastrophic, and records every cut as an event, so nothing leaves the
window without a record of what removed it.

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
   declarations in the dialog. A model without it folds every delta
   into the startup set and re-declares, paying the cache miss.
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
`claude-sonnet-5-course`, whose row sets neither. Add both rows to your
model table.

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
