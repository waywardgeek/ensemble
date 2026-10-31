# Chapter 10: Persistence

Every agent in this book so far is a goldfish. It reasons, calls tools,
streams into its GUI and loads skills on demand, and the moment the
process exits it forgets the user's name, the project, the afternoon of
work and every decision made along the way. The next start meets a
stranger. This chapter ends that. Nothing else in the book changes more
about what the agent is: a program that runs becomes one that
remembers.

## TL;DR

The agent saves itself on exit and loads itself on start. Neither needs
a flag. The save file is one JSON object: the configuration that shaped
the wire, the Chapter 2 context as a snapshot, the Seq that snapshot was taken
at, and the event log. A load installs the snapshot and replays only the
events after its anchor. A snapshot with no log is a complete save. So
is a complete log with no snapshot.

```go
// SaveFile joins the conversation snapshot to the facts that produced it.
type SaveFile struct {
    // Config records request settings without restoring them later.
    Config SaveConfig `json:"config"`
    // AsOf is the last event folded into Context, not the log length.
    AsOf uint64 `json:"as_of"`
    // Context is the Chapter 2 projection; null requests a full rebuild.
    Context *Context `json:"context"`
    // Log retains the available recorded facts, oldest first.
    Log []Event `json:"log"`
}

// SaveConfig records what shaped the wire. Never the API key.
type SaveConfig struct {
    // Model identifies the configured request model.
    Model string `json:"model"`
    // Vendor uses its existing disk spelling: anthropic, gemini or openai.
    Vendor Vendor `json:"vendor"`
    // SystemPrompt is the effective instruction used by the renderer.
    SystemPrompt string `json:"system_prompt"`
    // Tools contains declarations, never executable handlers.
    Tools []ToolDeclaration `json:"tools"`
}

// ToolDeclaration is the existing tool vocabulary with explicit disk names.
type ToolDeclaration struct {
    // Name identifies the callable tool.
    Name string `json:"name"`
    // Description tells the model when to use it.
    Description string `json:"description"`
    // Schema describes its JSON arguments.
    Schema json.RawMessage `json:"schema"`
}
```

The Go names follow the accepted implementation's common vocabulary; preserve
the JSON fields without requiring these exact internal type names.

1. **Default location.** The single-Agent commands select `save.json` in their
   working directory. The library defaults to `save.json` under the owning
   Agent's data directory. `--save PATH` names a different
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
6. **Save at exit.** When stdin closes, finish admitted work, shut down the
   Agent and join its managed jobs, then write the file and exit. `as_of`
   is the Seq of the last event folded into the saved context. The
   saved log is the loaded log plus every event created since, oldest
   first, Seq strictly increasing.
7. **Config is a record, not a restore.** On load the running agent's
   own model, vendor, prompt and tools win. The context is
   vendor-independent (ch2); the save must not pin a model.
8. **Replay is deterministic.** For a save whose log is complete, loading it
   as written and loading it with `context` set to null must produce byte-identical vendor
   requests for the same next prompt under the same runtime configuration.
   A trimmed log (rule 4) cannot rebuild the missing history, so the rule cannot
   apply to it, including a later save descended from that trimmed input.

Yours: indentation, whether to write through a temporary file and rename
(recommended; a crash mid-write otherwise destroys the only copy), a
`verify` subcommand for debugging, and what to print on load.

**Exercise.** Extend your Chapter 9 agent with this save/load lifecycle in
both CLI and optional GUI. Keep one public Agent implementation behind them.
From the repository root:

```sh
go run ./cmd/grade-edition2 -ch 10 path/to/core
```

The existing `-gui DIR`, `-gui-model` and earlier fixture options remain
available. The GUI stays a separate module; the grader uses the appropriate
command for each phase rather than applying headless checks to a GUI server.

| Check | Points | Proves |
|---|---|---|
| save-shape | 15 | recorded effective config and declaration fields, correct anchor and increasing log sequences |
| default-load | 20 | restart retains conversation; default and override use the same path for load and save; runtime configuration wins |
| replay-equals-snapshot | 20 | rule 8 |
| tail-applied-once | 15 | turn-2 snapshot spliced onto the turn-3 log yields the same next request as the turn-3 save |
| log-not-needed | 10 | rules 4 and 5 |
| bad-save-refused | 5 | rule 2 |
| ch9-parity | 15 | Chapter 9 still passes through the appropriate clients |

The persistence phases exercise both actual commands through their existing
stdin input and EOF lifecycle. In the GUI command, redirected prompt lines
use the same `{"kind":"prompt","text":"remember this"}` envelope as the
headless protocol; ordinary terminal text still works. Closing that stdin now
flushes admitted work and ends the application, rather than merely detaching
the terminal view. Browser prompts continue to use the same Agent.

Keep source tests for independent Agent directories, rich neutral context
and intentionally empty primary instructions. The numeric grader's chosen
fixtures cannot prove every ownership or renderer property.

## §10.1 In Plain Words

The event log is the truth. The context is what the truth means right
now. Chapter 2 split them on purpose: events are appended and never
edited, and the context is whatever the reducer makes of them. The
split has carried streaming, jobs, artifacts and skills. Persistence now
tests that relationship across a process boundary.

The context must be a pure function of the events.
No clock, no map iteration order, no field set by the engine behind the
reducer's back. The earlier replay checks depend on that rule too.
An agent that runs start to finish in one process can violate it
forever and never notice, because the live context is the only copy
anyone ever looks at.

Saving creates a second copy. Once a snapshot sits on disk next to the
log that produced it, the claim becomes falsifiable: rebuild from the
log and compare the rendered requests under the same configuration. A difference
exposes something the restore, reducer or renderer failed to preserve.
Persistence is the feature a user sees. Verification is
the reason it belongs this early in the book.

## §10.2 The Save File

Four fields: `config`, `as_of`, `context`, `log`. Use the existing Context
and Event types; a second persistence-specific conversation model would need
a second set of rules for the same history.

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
and base URLs are excluded from this metadata, so the save is portable across
machines and endpoints without carrying transport credentials.

There is a small distinction between configured and effective instructions.
The student's first save code copied `Config.SystemPrompt`. An ordinary Agent
without a skill left that field empty, while Engine supplied a nonempty default
to the actual request. The conversation resumed correctly and the metadata lied.
The grader's supplied primary skill hid the problem because that path had
already filled in the configured text.

The correction asks Engine for the effective instruction through its existing
owner interface. Saving and all three renderers now use the same decision.
An intentionally empty primary body must remain empty; blindly filling every
empty string with the default would fix one case by breaking the other.
The regression compares saved `system_prompt` with the captured HTTP field,
then checks an empty primary separately. Restoring either defect fails its
corresponding assertion. No new paid run was needed to verify this correction.

The save does not reconstruct a running job, a handler closure or the skills
registry's loaded/pending state. The new Agent supplies its current tools and
primary skill; its restored conversation may contain earlier tool results and
manuals. Serialized declarations describe what was available when saved. They
do not install executable capability on restart.

## §10.3 Snapshot Plus Tail

Loading has one loop. This excerpt belongs in the history implementation,
where `validate` and `apply` are the existing event validator and reducer;
`SaveFile` itself stays shared vocabulary in common:

```go
func restoreContext(sf common.SaveFile) (*common.Context, error) {
    ctx, after := sf.Context, sf.AsOf
    if ctx == nil {
        ctx, after = &common.Context{Turn: "idle"}, 0
    }
    for _, e := range sf.Log {
        if err := validate(e); err != nil {
            return nil, fmt.Errorf("restore: event %d: %w", e.Seq, err)
        }
        if e.Seq <= after {
            continue
        }
        apply(ctx, e)
    }
    return ctx, nil
}
```

With a snapshot, install it and apply only the tail: events whose Seq
is strictly greater than `AsOf`. Without one, start from an empty
context and apply everything. The same loop serves both, which leaves
one replay path to get right instead of two.

Decode into a fresh save value and validate its configuration, snapshot state
and increasing log sequence numbers before installing the restored history.
The excerpt shows the fold, not that surrounding validation and installation.
An invalid later event must not leave the live Agent half-restored. The accepted
loader stages the candidate context and log, then replaces the owned state only
after all checks succeed.

The tail looks unnecessary. A normal save from a complete history has
`AsOf` equal to the last Seq in its log, so the tail is empty and a
loader that skips it passes every test built from its own output. The
tail matters the moment a snapshot and a log come from different
moments: a snapshot kept from an earlier save, with a newer log written
after it. That is the shape a crash leaves behind, once the log is
written event by event and the snapshot only now and then. The grader
builds that shape on purpose, splicing an old snapshot onto a newer
log to distinguish a correct loader from one that ignores its anchor.
This chapter still saves on normal exit. Journaling events during execution
and recovering from a process crash arrive in Chapter 14; the tail fixture
does not claim that machinery exists yet.

The `<=` is the other half. An event at or below the anchor is already
inside the snapshot. Applying it again gives the conversation a
duplicate turn, and the model answers a question it has already
answered.

Take the snapshot and its anchor together. If one is copied before an append
and the other afterward, the file can claim to contain a fact its context has
never seen. History already owns those values; its existing lock can protect
the pair without a new persistence coordinator. Load before submitting work;
save after shutdown has joined the jobs that can still append final facts.

## §10.4 Replay Equals Snapshot

The chapter's central claim fits in one sentence. For a save whose log
is complete, loading it as written and loading it with `context` set
to null must produce byte-identical vendor requests for the next
prompt under the same runtime configuration.

The first load takes the snapshot path. The second takes the rebuild
path, replaying every event from nothing. If the reducer is a pure
function of the events, the two paths land on the same context and the
renderer turns that context into the same bytes. If they differ, inspect the
restoration and reduction first, then the renderer's ordering and configuration.
Both paths must use the same model, declarations and settings for the comparison
to isolate persistence.

The grader compares vendor requests.
A request is the only thing the model ever sees, and it is a format
every student's agent already produces, so the grader can compare two
of them without knowing anything about how a particular solution
stores its context. Two contexts that differ in some field the renderer
never reads are the same conversation. Two requests that differ by one
byte are not.

Byte-identical requests do not require byte-identical save files. Indentation
and object-key formatting on disk are yours; the renderer already determines
the bytes compared here. A canonical disk format, checksums or an origin
certificate would solve a different problem and add nothing to this test.

The precondition is real. A save with a trimmed log (§10.5) has
nothing to rebuild from, so nulling its context leaves an empty agent.
A run that starts fresh and retains its log can write a complete history.
After loading a deliberately trimmed save, however, the agent can retain only
the events it was given and the ones it creates next. A new save cannot recover
discarded facts. Keep the snapshot in that case; do not call nulling it a
replay-equivalence test.

## §10.5 The LLM Never Sees the Log

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

For example, a snapshot anchored at 7 with an empty log must assign 8 to the
next event. The resulting save has anchor 8 and one retained event, not an
eight-event log. Sequence identity and array length measure different things.

## §10.6 Default Load, and Refusing a Bad File

The first design had a `--load` flag. Bill's review of it was one
line: "Let's load by default without a flag." The agent now loads
`./save.json` at startup if it exists and writes it when stdin closes.
`--save PATH` names a different file for both directions and leaves
loading on.
That is the command's choice, so a project directory remembers its own
conversation. The library itself resolves its default under the owning Agent's
data directory; two Agents with different directories must not share one
process-wide save. An explicit path belongs to Agent configuration too. In the
current GUI command, preferences remain at `cr/settings.json`; the save does
not need to move beside them or take over their persistence.

Loading by default makes one failure mode dangerous. If `save.json`
exists but cannot be read, the agent exits with an error and leaves the
file untouched. Starting fresh instead looks friendlier and is worse:
the fresh session saves on exit and overwrites the history the user
came back for. A refusal costs one confusing startup. A silent reset
costs the conversation.

Install save-on-exit only after load succeeds. An unconditional deferred save
placed above `Load` would overwrite the bad file while returning the right
error. Check the bytes after refusal and verify that no HTTP request was sent;
an error message alone leaves the destructive half of that bug invisible.

For a successful exit, the accepted implementation writes a temporary file in
the destination directory, closes it, then renames it over the selected path.
Encoding, close and rename failures propagate to the caller. This avoids
truncating the old save before the replacement is complete; it makes no promise
of power-loss durability. A database or write-ahead journal is unnecessary for
the normal-exit exercise.

## §10.7 Taking It for a Spin

Run the agent in an empty directory, ask it to remember a word, and
close stdin with Ctrl-D after the answer. `save.json` appears. Run it again in
the same directory and ask for the word. The second process never saw the first one's
turn; it answers from a context restored out of a file. Build from the core
module, then launch from your scratch directory with provider configuration
already in the child environment:

```sh
go build -o /tmp/ensemble-ch10 .
```

```sh
/tmp/ensemble-ch10 chat
```

For the override case, use `/tmp/ensemble-ch10 --save chosen.json chat`
on both starts. It should create and resume `chosen.json` without also creating
the default `save.json`. Test the optional GUI command's stdin EOF too; a
browser that appears restored does not establish that the application saved
when its terminal closed.

Then break copies on purpose. Preserve the complete first save, and make each
variant independently from it. Delete `"log"` down to `[]` and the agent
still remembers, because the log was never on the rendering path. In another
copy, set `"context"` to `null` and it still remembers, because the full log
rebuilds the same context. Do not make that second edit to the already trimmed
copy. Truncate a third copy to half its bytes and the agent refuses to start,
which is the correct answer.

The first-edition reference also shipped a `verify` subcommand that rebuilt
from the log and printed `MATCH` or `MISMATCH` against the saved
snapshot. It remains an optional debugging aid outside the contract; the
accepted Edition 2 solution does not need it. The grader never trusts an agent's
opinion of its own determinism; it compares the requests.

The retained October 10, 2026 sessions used `claude-haiku-5-5`, `gpt-4o-mini`
and `gemini-3.8-flash`, reusing the same day's model discovery from the Skills
exercise. Each first process received:

> Remember the secret word cobalt-juniper. Reply briefly; do not use tools.

Each second process was asked for the word without supplying it again, and each
answered `cobalt-juniper`. The Claude session also resumed independently from
the snapshot-only and log-only variants, with the same answer. Those eight
model-response processes exited normally without a timeout. The ninth process,
given the truncated save, exited 1 before any HTTP request and left the damaged
bytes unchanged. The [Claude](../solutions/evidence/ch11/anthropic-live.json),
[OpenAI](../solutions/evidence/ch11/openai-live.json) and
[Gemini](../solutions/evidence/ch11/gemini-live.json) receipts retain
the actual inputs, replies and captured conversation payloads. These are
historical model choices, not an account inventory for your next run.

Remembering the word demonstrates useful restoration, but it cannot establish
byte identity. A model might answer correctly despite duplicated history or
a dropped tool result. The deterministic source checks use all three real
renderers with recorded tool identities, opaque material, a redaction and pending
ephemera. Snapshot, full replay, older-snapshot-plus-tail and snapshot-only
variants produce the same rendered request. Appending afterward verifies the
next sequence and the next save's anchor as separate obligations.

The original full grader was interrupted and unscored during inherited parity.
Its separately invoked Persistence phases scored **70/85** on the initial
submission. The shape check stopped at the undefined `think` declaration, before
checking later sequence and anchor predicates; those predicates cannot be called
original-grader passes. Independent final-source inspection and focused tests
verified the valid shape obligations after the effective-prompt correction.
The [original review](../docs/edition-2-notes/ch11-code-review.md) preserves
that distinction and the live receipts' earlier source provenance. The
[Edition 2 grader report](../docs/edition-2-notes/enhancement-10-code.md)
records revised results separately.
