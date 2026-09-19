# Chapter 11 — Persistence

## Motivational Intro

Every agent you have built so far forgets everything the moment it exits. The
event log writes to disk, but the agent never reads it back. That is not a
limitation of the data structures — they already round-trip as JSON. It is
a missing feature: nobody told the agent to save its state, and nobody told
it how to resume.

This chapter closes that gap. You will build save and load, and in doing so
you will prove three invariants that have been implicit since Chapter 2:

1. The context is deterministically derived from the event log
2. A checkpoint plus the remaining events produces the same context as the full log
3. The LLM never needs the log — the context alone is sufficient to continue

These are not aspirations. They are testable properties. The grader verifies
all three.

## TL;DR

### SaveFile

A save file is `{context, log, config}` as JSON. The config captures what the
LLM saw (vendor, model, system prompt, tool declarations), not infrastructure
(API keys, base URLs). Save writes it. Load reads it. Both use the existing
JSON marshaling on Context, Log, and ToolDecl.

### Rebuild

`Rebuild(log) → context` replays the full event log from scratch using
`Apply`. The result must be byte-identical to the context built incrementally
during the conversation. This is the determinism invariant made testable.

### Checkpoint Resume

Save at event N. Continue to event N+M. The context at N+M must equal
`Rebuild(log[0..N+M])` and also `Load(checkpoint_N).Context + Apply(events[N+1..N+M])`.

### CLI Integration

`--save PATH` persists state after each turn. `--load PATH` resumes from a
save file. A loaded agent continues the conversation with its full history
visible to the LLM.

### Log Not Needed

Load a save file, discard the log. Send a prompt. The conversation continues.
The LLM sees the context (dialogue entries), not the log (raw events). The log
exists for auditing and rebuild — it is not in the critical path for
conversation.

## §11.1 Why Persistence Matters

The event log is the truth. The context is what the truth means right now.
Save both and you have a conversation you can resume, audit, or replay.

But the deeper reason is invariant verification. Every `Apply` call in every
chapter has been an implicit claim: "this reducer is deterministic, and the
context it produces is the only thing the LLM needs." Save and load make that
claim falsifiable.

## §11.2 The SaveFile

Three fields. The context is what the LLM sees. The log is the audit trail.
The config is what built the system prompt and tools. Together they are a
complete snapshot of an agent's state.

The config subset: vendor, model, system prompt, tool declarations. Not API
keys, not base URLs, not HTTP clients. The save file is portable — you can
load it on a different machine, with a different API key, against a different
endpoint.

## §11.3 Rebuild

A five-line function that proves the foundation:

```go
func Rebuild(log *Log) (*Context, error) {
    ctx := NewContext()
    for _, ev := range log.Events {
        if err := ctx.Apply(ev); err != nil {
            return nil, err
        }
    }
    return ctx, nil
}
```

If `json.Marshal(Rebuild(log))` differs from `json.Marshal(savedContext)`, the
reducer has a bug. This is the most valuable test in the chapter.

## §11.4 Checkpoint and Partial Replay

Save at event 5. Continue to event 10. Three things must be equal:
1. The live context at event 10
2. `Rebuild(log[0..10])`
3. `Load(checkpoint_5).Context` with `Apply(events[6..10])`

If any pair disagrees, the reducer depends on something beyond (state, event),
and that hidden dependency will corrupt every conversation that resumes from
a checkpoint.

## §11.5 The LLM Never Sees the Log

The renderer reads the Context. The Context contains Dialogue entries. The
Dialogue entries contain Parts. At no point does the renderer read the Log.

This means a loaded context with an empty log produces the exact same LLM
request as one with a full log. The grader tests this directly: load a save,
discard the log, send a prompt, verify the conversation continues normally.

## §11.6 CLI Wiring

`--save` and `--load` are flags, not modes. A loaded agent is
indistinguishable from one that got there by running — same tools, same system
prompt, same conversation history. The save path can be specified per-run or
set via `EN_SAVE_PATH`.

## §11.7 Config Preservation

The save file's config must contain the exact model name, vendor, system
prompt text, and tool declarations that were active at save time. This is what
makes resurrection faithful: you load the config, and the agent sees the
same system prompt and tools it saw before.

## §11.8 Looking Ahead

With persistence in place, two things become possible that were not before:
1. Memory cascade — compaction that survives across sessions
2. Agent resurrection — loading a saved agent with mock tools for interview

Both are later chapters. This chapter buys down the tech debt that makes them
safe.
