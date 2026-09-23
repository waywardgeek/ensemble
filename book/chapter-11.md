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
6. **Save at exit.** When stdin closes, write the file and exit. `as_of`
   is the Seq of the last event folded into the saved context. The
   saved log is the loaded log plus every event created since, oldest
   first, Seq strictly increasing.
7. **Config is a record, not a restore.** On load the running agent's
   own model, vendor, prompt and tools win. The context is
   vendor-independent (ch2); the save must not pin a model.
8. **Replay is deterministic.** For a save whose log is complete (every
   save the agent writes itself), loading it as written and loading it
   with `context` set to null must produce byte-identical vendor
   requests for the next prompt. A trimmed log (rule 4) has nothing to
   replay, so the rule cannot apply to it.

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

## §11.1 In Plain Words

The event log is the truth. The context is what the truth means right
now. Chapter 2 split them on purpose: events are appended and never
edited, and the context is whatever `Apply` makes of them. Ten
chapters later the split has carried streaming, jobs, artifacts, and
skills without once being tested for the one thing it was built for.

That thing is this: the context must be a pure function of the events.
No clock, no map iteration order, no field set by the engine behind the
reducer's back. Every chapter since has assumed it. Nothing has checked
it. An agent that runs start to finish in one process can violate it
forever and never notice, because the live context is the only copy
anyone ever looks at.

Saving creates a second copy. Once a snapshot sits on disk next to the
log that produced it, the claim becomes falsifiable: rebuild from the
log, compare with the snapshot, and any difference is a reducer bug
with a name. Persistence is the feature a user sees. Verification is
the reason it belongs this early in the book.

## §11.2 The Save File

Four fields:

```go
type SaveFile struct {
    Config  SaveConfig `json:"config"`
    AsOf    Seq        `json:"as_of"`
    Context *Context   `json:"context"`
    Log     []Event    `json:"log"`
}
```

`Context` is the snapshot: exactly what the renderer reads. `Log` is
the audit trail. `AsOf` joins them. It is the Seq of the last event
already folded into the snapshot, and it is the field everything else
in the chapter hangs on. Without it, a loader holding a snapshot and a
log cannot tell which events the snapshot already contains, so it has
two choices and both are wrong: apply the whole log and duplicate every
turn, or apply none of it and lose whatever came after the snapshot.

`Config` records vendor, model, system prompt, and tool declarations.
It is a record only. The running agent's own configuration
wins on load, so a save made with one model resumes under whatever
model the user starts with today. A save that pinned its model would
turn every model retirement into a pile of unloadable files. API keys
and base URLs never enter the file at all; it is portable across
machines and endpoints.

## §11.3 Snapshot Plus Tail

Loading has one loop:

```go
func (sf *SaveFile) Restore() (*Context, error) {
    ctx, after := sf.Context, sf.AsOf
    if ctx == nil {
        ctx, after = NewContext(), 0
    }
    for _, e := range sf.Log {
        if e.Seq <= after {
            continue
        }
        if err := ctx.Apply(e); err != nil {
            return nil, fmt.Errorf("restore: event %d: %w", e.Seq, err)
        }
    }
    return ctx, nil
}
```

With a snapshot, install it and apply only the tail: events whose Seq
is strictly greater than `AsOf`. Without one, start from an empty
context and apply everything. The same loop serves both, which leaves
one replay path to get right instead of two.

The tail looks unnecessary. A save the agent writes itself always has
`AsOf` equal to the last Seq in its log, so the tail is empty and a
loader that skips it passes every test built from its own output. The
tail matters the moment a snapshot and a log come from different
moments: a snapshot kept from an earlier save, with a newer log written
after it. That is the shape a crash leaves behind, once the log is
written event by event and the snapshot only now and then. The grader
builds that shape on purpose, splicing an old snapshot onto a newer
log, because it is the only fixture that tells a correct loader from
one that ignores its anchor.

The `<=` is the other half. An event at or below the anchor is already
inside the snapshot. Applying it again gives the conversation a
duplicate turn, and the model answers a question it has already
answered.

## §11.4 Replay Equals Snapshot

The chapter's central claim fits in one sentence. For a save whose log
is complete, loading it as written and loading it with `context` set
to null must produce byte-identical vendor requests for the next
prompt.

The first load takes the snapshot path. The second takes the rebuild
path, replaying every event from nothing. If the reducer is a pure
function of the events, the two paths land on the same context and the
renderer turns that context into the same bytes. If they differ, the
reducer is reading something besides its input, and that dependency
will corrupt every conversation that resumes from disk.

The comparison is made on vendor requests.
A request is the only thing the model ever sees, and it is a format
every student's agent already produces, so the grader can compare two
of them without knowing anything about how a particular solution
stores its context. Two contexts that differ in some field the renderer
never reads are the same conversation. Two requests that differ by one
byte are not.

The precondition is real. A save with a trimmed log (§11.5) has
nothing to rebuild from, so nulling its context leaves an empty agent.
Every save the agent writes itself carries its whole history, which is
where the claim applies and where the grader tests it.

## §11.5 The LLM Never Sees the Log

The renderer reads the context. The context holds dialogue entries,
the entries hold parts, and at no point does the renderer consult the
log. A save with `"log": []` and a non-null context is therefore
complete: it loads, it resumes, and the next request carries the full
conversation.

This is what makes the log trimmable later without touching behavior.
It also creates one small trap. Numbering must continue after the
loaded events, and a save with an empty log knows its anchor and
nothing else. The next Seq is one past the larger of `AsOf` and the
last event in the log; neither alone is enough.

## §11.6 Default Load, and Refusing a Bad File

The first design had a `--load` flag. Bill's review of it was one
line: "Let's load by default without a flag." The agent now loads
`./save.json` at startup if it exists and writes it when stdin closes.
`--save PATH` names a different file for both directions and leaves
loading on.
The file sits beside `settings.json`, in the directory the agent runs
from, so a project directory remembers its own conversation.

Loading by default makes one failure mode dangerous. If `save.json`
exists but cannot be read, the agent exits with an error and leaves the
file untouched. Starting fresh instead looks friendlier and is worse:
the fresh session saves on exit and overwrites the history the user
came back for. A refusal costs one confusing startup. A silent reset
costs the conversation.

## §11.7 Taking It for a Spin

Run the agent in an empty directory, ask it to remember a word, and
close stdin. `save.json` appears. Run it again in the same directory
and ask for the word. The second process never saw the first one's
turn; it answers from a context rebuilt out of a file.

Then break things on purpose. Delete `"log"` down to `[]` and the agent
still remembers, because the log was never on the rendering path. Set
`"context"` to `null` and it still remembers, because the log alone
rebuilds the same context. Truncate the file to half its bytes and the
agent refuses to start, which is the correct answer.

The reference solution also ships a `verify` subcommand that rebuilds
from the log and prints `MATCH` or `MISMATCH` against the saved
snapshot. It is a debugging aid outside the contract, and nothing
grades it. The grader never trusts an agent's opinion of its own
determinism; it compares the requests.
