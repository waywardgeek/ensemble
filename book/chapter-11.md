# Chapter 11: Persistence

Every agent you have built so far forgets everything the moment it exits. The event log writes to disk, but the agent never reads it back. This chapter closes that gap and, in doing so, proves three invariants that have been implicit since Chapter 2:

1. The context is deterministically derived from the event log.
2. A checkpoint plus the remaining events produces the same context as the full log.
3. The LLM never needs the log. The context alone is sufficient to continue.

These are not aspirations. They are testable properties. The grader verifies all three.

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
