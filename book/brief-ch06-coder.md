# Chapter 6 Coder Brief — Two Seams and a Loop

**Role**: Coder (Opus 5). Build the reference solution and grader for
Chapter 6.

**Read before starting**:
- `book/chapter-06-outline.md` — the full outline (read every word)
- `book/chapter-06-seam-draft.md` — Go type drafts for the seam
- `book/voice.md` — rulings that bind
- `agent/` — the current live code (your starting point = ch05-solution)
- `internal/grade/ch05_*` — ch5 grader (structural model for ch6 grader)
- `internal/fakevendor/` — the shared fake vendor

**Starting point**: Chapter 5's star topology is complete. The student has:
- `internal/common/` as the hub (all shared types + interfaces)
- `internal/llm/` (engine + vendors), `internal/jobs/`, `internal/tools/`
- `agent.go` public facade with type aliases
- `agent/cmd/main.go` CLI
- Host parent-interface chain, no mutable globals
- 120/120 on the ch5 grader

---

## Step 0: Understand the architecture

Read ALL files in `agent/` before touching anything. The star topology
(internal/common is the hub, implementation packages are spokes) must be
preserved. Everything you add in this chapter is either:
- A new interface/type in `internal/common/` (the hub)
- New implementation in an existing spoke or a new spoke
- An update to the public facade in `agent.go`

Chapter 5's rules remain in force:
- Implementation packages import only `internal/common/`
- `internal/common/` imports only stdlib + the logger package from ch5
- The public facade re-exports via type aliases
- No mutable package-level variables

---

## Step 1: Add Observer types to internal/common/

The outbound seam. Add to `internal/common/`:

### Observer interface
```go
type Observer interface {
    Observe(Observation)
}
```
Observe MUST NOT BLOCK. A slow observer that blocks re-creates the deafness.

### Observation types (sealed union)
```go
type Observation interface{ isObservation() }

type PartDelta struct {
    Agent  AgentID `json:"agent,omitempty"`
    PartID uint64  `json:"part_id"`
    Chunk  string  `json:"chunk"`
}

type PartFinal struct {
    Agent  AgentID `json:"agent,omitempty"`
    Seq    int     `json:"seq"`
    PartID uint64  `json:"part_id"`
    Part   Part    `json:"part"`
}

type StateChanged struct {
    Agent AgentID   `json:"agent,omitempty"`
    From  TurnState `json:"from"`
    To    TurnState `json:"to"`
}
```

### AgentID
```go
type AgentID string
```
Empty means "the one agent". `omitempty` everywhere so single-agent logs
are unchanged.

---

## Step 2: Add Mailbox types to internal/common/

The inbound seam. Types go in `internal/common/`; implementation can go
in the engine or a new spoke.

```go
type Inbound interface{ isInbound() }

type UserMessage struct{ Text string }
type Hint struct{ Text string }
type ToolCompleted struct {
    CallID string
    Result string
    IsError bool
}
type Interrupt struct{}
```

The key principle: ONE queue carries all of these. A prompt, a hint, a
tool finishing, and an interrupt are the same kind of event to the loop.

---

## Step 3: Add Ref and ModelFeatures to internal/common/

### Ref type
```go
type RefKind uint8
const (
    RefPath   RefKind = iota + 1
    RefURI
    RefHandle
)

type Ref struct {
    Kind    RefKind `json:"kind"`
    Locator string  `json:"locator"`
}
```

Amend `BlobPart` to use `Ref` instead of `Path string`. Amend
`RedactedPart` to carry a `Ref` for recoverability.

### ModelFeatures
```go
type Media uint8
const (
    MediaImage    Media = 1 << iota
    MediaAudio
    MediaVideo
    MediaDocument
)

type ModelFeatures struct {
    Media Media
}

// No default row. Unknown model → loud refusal.
func LookupModel(model string) (ModelFeatures, bool)
```

Update the renderers to check ModelFeatures before rendering a BlobPart.
If the model doesn't support the media type, return an error naming the
model and the media type. Do NOT silently drop the part.

---

## Step 4: Build the actor loop

This is the chapter's core. Replace the synchronous `Ask` in the engine
with an actor loop that drains a mailbox.

### The current Ask (synchronous, deaf)
```go
func (e *Engine) Ask(text string) (string, error) {
    e.Say(text)
    for round := 0; ; round++ {
        reply := e.Turn()
        calls := e.pendingCalls()
        if len(calls) == 0 { break }
        for _, call := range calls {
            e.Execute(call)  // BLOCKS here — agent is deaf
        }
    }
    return reply, nil
}
```

### The new actor loop (non-blocking, hears hints)
The engine gets a mailbox (a mutex-guarded queue with a signal channel).
When a tool finishes, its completion is posted to the mailbox instead of
being awaited in a select. The actor loop drains the mailbox and processes
messages in order:

- **UserMessage**: start a new turn (render, send, record)
- **Hint**: attach to the current turn's context (pending hint)
- **ToolCompleted**: record the result, check if more calls pending
- **Interrupt**: set the interrupted state, stop processing

The actor loop notifies observers on every state change and content event.

### Keep Ask as a convenience
The synchronous `Ask(text) (string, error)` becomes:
1. Post a UserMessage to the mailbox
2. Wait for the turn to end (state → Idle or Interrupted)
3. Return the last agent text

This is the "blocking send = write to mailbox + wait for turn-end" pattern
from the outline.

### The Wait primitive
```go
func (a *Agent) Wait(ctx context.Context, pred func(Observation) bool) (Observation, error)
```
Blocks until an observation satisfies `pred`, or `ctx` is done. All waiting
operations derive from this (blocking send, wait-for-agent, join, wake-any).

---

## Step 5: Add hint/interrupt support to cmd/main.go

Update `cmd/main.go` to read stdin on its own goroutine. Support a JSON
protocol:

```json
{"kind":"prompt","text":"..."}
{"kind":"hint","text":"..."}
{"kind":"interrupt"}
```

Keep backward compatibility with the ch1–5 format:
```json
{"user":"..."}
```

Also add the observation stream on stdout: when an observer fires, emit
JSON lines for state changes and content events. This is the door the
grader walks through.

---

## Step 6: Build the multi-agent Framework

A `Framework` type (or equivalent) that manages multiple agents. It lives
in the engine or a new implementation package.

Key operations:
- Create agents with different configs, tools, and prompts
- Attach observers to individual agents or all agents
- Wait for state changes across multiple agents (wake-once semantics)

The exercise needs this to run three agents. The framework should be
minimal — just enough to support the exercise.

---

## Step 7: Build the exercise binary

One binary, `./ch06`, runs a three-agent author/editor/reviewer workflow.

The author writes a draft, the editor improves it, the reviewer approves it.
Each agent has its own tools and prompt. The workflow is hardcoded Go code
that passes messages between agents.

Contract:
- Inbound events arrive as JSON lines on stdin
- Observations leave as JSON lines on stdout
- Event log written to the path in `CH06_LOG`
- Vendor: the fake from ch2–5, unchanged

**IMPORTANT**: The reviewer is a THIRD actor, not a second editor pass.
Three agents prove the cost was paid once.

The exercise binary should demonstrate:
1. The observer seam firing (state changes + content on stdout)
2. The agent hearing hints mid-tool (not-deaf)
3. Multi-agent coordination (wake-once)
4. Loud media refusal (if triggered by the grader)

---

## Step 8: Build the grader

In `internal/grade/ch06_*`. Seven checks, 100 points:

| Check | Pts | What it tests |
|---|---|---|
| `ch5-parity` | 10 | Re-run ch5 grader against student tree |
| `not-deaf` | 25 | Fake requests a slow tool; grader writes hint to stdin while it runs; observation stream must show hint BEFORE tool completion |
| `replay-is-live` | 15 | Byte-compare context from log replay vs live run |
| `observer-fires` | 15 | Observation stream contains both state transitions AND content events (PartFinal) |
| `wake-once` | 15 | Two agents finish same turn; exactly one wakeup on observation stream |
| `loud-refusal` | 10 | Present unsupported media; renderer refuses naming model+media |
| `hub-clean` | 10 | `go list -deps` on internal/common/ shows only stdlib + first-party leaves |

### Implementing not-deaf (the hardest check)

The fake vendor must script a sequence that includes a slow tool call:
1. Fake responds with a tool call for a tool that sleeps (e.g. 3 seconds)
2. While the tool sleeps, the grader writes `{"kind":"hint","text":"redirect"}` to stdin
3. The grader reads stdout observations and checks that a `Hint`-related
   observation appears BEFORE the `ToolCompleted` observation
4. A student whose engine blocks on tool calls will show the hint AFTER
   the tool completion — the hint was queued but not observed until the
   tool finished

### Implementing wake-once

The workflow runs two agents that both make a turn. The fake serves both
quickly. The grader checks that the parent observes both completions in a
single wakeup — not two separate ones.

---

## Step 9: Mutation tests

Write mutation tests per P9. Each mutant drops a specific behavior and
asserts an exact set of failing checks.

Key mutants:
- Remove mailbox (tools block inline) → not-deaf fails
- Remove observer notifications → observer-fires fails
- Add polling loop instead of wake → wake-once fails
- Remove ModelFeatures check → loud-refusal fails
- Remove replay discipline → replay-is-live fails

---

## Step 10: Snapshot and tag

```bash
cp -r agent/ solutions/ch06/
# Verify
make grade6    # must be 100/100
make grade5    # regression check
make grade     # ch1 regression
git tag ch06-solution
```

---

## Easter egg

Every chapter solution carries a `waywardgeest` easter egg in a doc comment.
The Observer or Mailbox doc comment would be a natural home.

---

## Constraints

- All work in `agent/`, not `solutions/ch06/`
- `make grade` through `make grade5` must pass after every commit
- Never `git add -A`; never stage `book/preface.md` or Bill's uncommitted
  book files
- Commit as CodeRhapsody
- Push is Bill's
- The fake vendor is SHARED — regression-check earlier graders after
  any change to it
