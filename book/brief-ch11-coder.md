# Chapter 11 Coder Brief: Persistence

## Chapter thesis

The context is deterministic. Save the log, rebuild the context. Save both,
resume from the checkpoint. The LLM never sees the log; it sees the context.
These are testable invariants, not aspirations.

## What the student builds

### 1. SaveFile type (internal/common/save.go)

```go
type SaveFile struct {
    Context *Context         `json:"context"`
    Log     *Log             `json:"log"`
    Config  SaveConfig       `json:"config"`
}

type SaveConfig struct {
    Vendor       Vendor       `json:"vendor"`
    Model        string       `json:"model"`
    SystemPrompt string       `json:"system_prompt"`
    Tools        []ToolDecl   `json:"tools"`
}
```

The SaveConfig is a SUBSET of Config. It captures what the LLM saw, not
infrastructure (BaseURL, APIKey, HTTP client, Cwd). A resurrected agent
uses the saved system prompt and tools; a resumed agent rebuilds from code.

### 2. Save and Load functions (internal/common/save.go)

```go
func Save(path string, ctx *Context, log *Log, cfg Config) error
func Load(path string) (*SaveFile, error)
```

Save writes JSON. Load reads it. Both use the existing JSON marshal/unmarshal
on Context, Log, ToolDecl, and PartList (all already implemented).

### 3. Rebuild function (internal/common/context.go or save.go)

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

Replays the full log from scratch. The result must be identical to the
context that was built incrementally during the conversation.

### 4. Engine integration

Add to the engine:
- `Save(path string) error` — saves current state
- `LoadAndResume(path string) error` — loads a save file, sets engine's
  Log/Ctx, optionally rebuilds Config from code or uses saved Config

### 5. CLI integration

Add `--save PATH` and `--load PATH` flags to `cmd/main.go`.
- On `--load`: read the save file, set the engine's context and log,
  continue the conversation from where it left off
- On `--save`: after each turn completes, save the current state

### 6. Built-in `save` and `dump` tools

- `save` tool: saves the current state to a path (or default path)
- `dump` tool: writes the event log to stdout or a file (debugging aid)

These already exist partially (`dump` writes the log). Extend or add `save`.

## Grader checks (7 checks, 100 points)

### deterministic-rebuild (25 points)
Drive a multi-turn conversation (prompt → response → tool call → result →
prompt → response). Save the context and log. Rebuild a fresh context from
the log alone. Compare JSON-marshaled contexts byte-for-byte.

### checkpoint-resume (20 points)
Drive conversation for N events. Save. Drive M more events. Save again.
Load the first save, replay only events N+1..N+M. Compare the result to
the second save's context.

### save-load-roundtrip (15 points)
Save the full state. Load it. Re-save. Compare the two save files
byte-for-byte.

### resume-continues (15 points)
Drive a conversation, save. Start a NEW binary with `--load`, send a
prompt. Verify the fakevendor receives a request whose conversation history
includes the prior turns (check message count in the request body).

### config-preserved (10 points)
Save. Load. Verify the SaveConfig contains the correct model, vendor,
system prompt, and tool names.

### log-not-needed (10 points)
Load a save file. Delete the log field (set to empty). Send a prompt.
Verify the conversation continues (the LLM only sees the context, not
the log).

### ch10-parity (5 points)
All ch10 checks still pass.

## Fixture data

The fakevendor provides scripted responses. The grader drives a multi-turn
conversation, exercises save/load, and verifies the invariants.

## Environment variables

- `EN_SAVE_PATH` — default save file path (optional)

## What NOT to build

- Memory cascade (later chapter)
- Interview/resurrection mode with mock tools (later chapter)
- Compaction during save/load (the reducer already handles Redacted events)

## Files to read before coding

- `internal/common/context.go` — the reducer (Apply)
- `internal/common/event.go` — event types and JSON marshaling
- `internal/common/event_log.go` — the Log type
- `internal/common/config.go` — Config, ToolDecl, Renderer
- `internal/common/part.go` — Part types and PartList JSON round-trip
- `internal/llm/engine.go` — the Engine struct and Turn method
- `cmd/main.go` — CLI arg parsing, runActorLoop
- `internal/grade/ch10_checks.go` — grader pattern
- `internal/grade/ch10_harness.go` — test harness pattern

## Success criteria

`make grade11` passes 100/100. All previous grades unaffected.
