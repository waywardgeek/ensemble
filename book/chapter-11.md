# Chapter 11: Persistence

Every agent in this book so far is a goldfish. It reasons, calls tools,
drives its own GUI and loads skills on demand, and the moment the
process exits it forgets the user's name, the project, the afternoon of
work and every decision made along the way. The next start meets a
stranger. This chapter ends that. Nothing else in the book changes more
about what the agent is: a program that runs becomes a colleague that
stays.

## TL;DR

The agent saves itself on exit and loads itself on start. Neither needs
a flag. The save file is one JSON object: the configuration that shaped
the wire, the ch2 context as a snapshot, the Seq that snapshot was taken
at, and the event log. A load installs the snapshot and replays only the
events after its anchor. A snapshot with no log is a complete save. So
is a log with no snapshot.

```go
// SaveFile is the whole agent on disk.
type SaveFile struct {
    Config  SaveConfig `json:"config"`
    AsOf    Seq        `json:"as_of"`   // last event folded into Context
    Context *Context   `json:"context"` // ch2 Context as-is; null = rebuild
    Log     []Event    `json:"log"`     // ch2 events, oldest first
}

// SaveConfig records what shaped the wire. Never the API key.
type SaveConfig struct {
    Model        string     `json:"model"`
    Vendor       string     `json:"vendor"` // "anthropic", "gemini", "openai"
    SystemPrompt string     `json:"system_prompt"`
    Tools        []ToolDecl `json:"tools"`
}

// ToolDecl gains JSON tags so tools[].name reads cleanly on disk.
type ToolDecl struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Schema      json.RawMessage `json:"schema"`
}
```

1. **Default location.** The save file is `save.json` in the working
   directory, beside `settings.json`. `--save PATH` names a different
   file. The flag changes where, never whether: the same path is loaded
   at start and written at exit.
2. **Load at start.** A missing file means a fresh start. A file that
   exists but does not parse as a `SaveFile` is a fatal error: exit
   non-zero, name the file, leave its bytes untouched. Starting fresh
   over a save that failed to load would overwrite the user's history
   at exit.
3. **Snapshot plus tail.** If `context` is non-null, install it, then
   apply in order every log event with `seq > as_of`. If `context` is
   null, apply every log event to a fresh context. Events at or below
   `as_of` are already inside the snapshot; applying one twice is a bug.
4. **The log is not needed.** `"log": []` with a non-null `context` is a
   complete save. The vendor sees the context, never the log.
5. **Numbering continues.** The first new event gets the Seq one past
   the larger of `as_of` and the last log event's Seq.
6. **Save at exit.** When stdin closes, write the file and exit within
   10 seconds. `as_of` is the Seq of the last event folded into the
   saved context. The saved log is the loaded log plus every event
   created since, oldest first, Seq strictly increasing.
7. **Config is a record, not a restore.** On load the running agent's
   own model, vendor, prompt and tools win. The context is
   vendor-independent (ch2); the save must not pin a model.
8. **Replay is deterministic.** Loading a save as written, and loading
   the same save with `context` set to null, must produce byte-identical
   vendor requests for the next prompt.

Yours: indentation, whether to write through a temporary file and rename
(recommended; a crash mid-write otherwise destroys the only copy), a
`verify` subcommand for debugging, and what to print on load.

**Exercise.** Start from your ch10 agent. Add save and load until
`make grade-dir CH=11 DIR=path/to/agent` scores 100/100.

| Check | Points | Proves |
|---|---|---|
| save-shape | 15 | config fields present, `as_of` equals the last log Seq, log Seq strictly increasing |
| default-load | 20 | a second start in the same directory, no flags, sends the first session's prompts to the vendor |
| replay-equals-snapshot | 20 | rule 8 |
| tail-applied-once | 15 | turn-2 snapshot spliced onto the turn-3 log yields the same next request as the turn-3 save |
| log-not-needed | 10 | rules 4 and 5 |
| bad-save-refused | 5 | rule 2 |
| ch10-parity | 15 | chapter 10 still passes |

## §11.1 The Why

The event log is the truth. The context is what the truth means right now. Save both and you have a conversation you can resume, audit, or replay from any point.

But the deeper reason is verification. Every `Apply` call in every chapter has been an implicit claim: this reducer is deterministic, and the context it produces is the only thing the LLM needs. Save and load make that claim falsifiable. If the claim is wrong, you will find out now, not three chapters from now when memory compaction silently corrupts a conversation.

## §11.2 The SaveFile

Three fields:

```go
type SaveFile struct {
    Context *Context   `json:"context"`
    Log     *Log       `json:"log"`
    Config  SaveConfig `json:"config"`
}
```

The context is what the LLM sees. The log is the audit trail. The config is what built the system prompt and tools. Together they are a complete snapshot of an agent's state.

The config subset captures what matters for resurrection:

```go
type SaveConfig struct {
    Vendor       string     `json:"vendor"`
    Model        string     `json:"model"`
    SystemPrompt string     `json:"system_prompt"`
    Tools        []ToolDecl `json:"tools"`
}
```

Vendor, model, system prompt, tool declarations. Not API keys, not base URLs, not HTTP clients. The save file is portable. Load it on a different machine, with a different API key, against a different endpoint.

## §11.3 Rebuild

A function that proves the foundation:

```go
func Rebuild(events []Event) (*Context, error) {
    ctx := NewContext()
    for _, ev := range events {
        if err := ctx.Apply(ev); err != nil {
            return nil, err
        }
    }
    return ctx, nil
}
```

Replay the full event log from scratch. Marshal both contexts. Compare byte for byte. If the bytes differ, the reducer has a bug. This is the most valuable test in the chapter because it validates every `Apply` call you have written since Chapter 2, retroactively, in one comparison.

## §11.4 Checkpoint and Partial Replay

Save at event 5. Continue to event 10. Three things must be equal:

1. The live context at event 10.
2. `Rebuild(log.Events[:10])`.
3. `Load(checkpoint_5).Context` with `Apply(events[5:10])`.

If any pair disagrees, the reducer depends on something beyond (state, event). That hidden dependency will corrupt every conversation that resumes from a checkpoint. The grader tests this by saving mid-conversation, continuing, then comparing all three values.

## §11.5 The LLM Never Sees the Log

The renderer reads the Context. The Context contains Dialogue entries. The Dialogue entries contain Parts. At no point does the renderer read the Log.

This means a loaded context with an empty log produces the exact same LLM request as one with a full log. The grader tests this directly: load a save file, send a new prompt, verify the vendor receives a well-formed request with the full conversation history. The log is for auditing and rebuild. It is not in the critical path.

## §11.6 CLI Integration

Two flags:

```
--save PATH    Persist state to PATH after the conversation
--load PATH    Resume from a previously saved file
```

A loaded agent is indistinguishable from one that got there by running. Same tools, same system prompt, same conversation history. The only difference is startup time: load skips the events and starts from the result.

The verify command tests determinism from the command line:

```
./agent verify SAVE_FILE
```

It loads the save file, rebuilds the context from the log, and compares. If they match, it prints OK. If they differ, it prints the byte offset of the first difference. The grader calls this command to verify the invariant without importing your internal packages.

## §11.7 Looking Ahead

With persistence in place, two capabilities become possible that were not before:

1. Memory cascade: compaction that survives across sessions, compressing old memories while preserving the conversation.
2. Agent resurrection: loading a saved agent with mock tools for interview, seeing exactly what it saw, asking what it was thinking.

Both are later chapters. This one buys down the tech debt that makes them safe.
