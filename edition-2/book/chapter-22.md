# Chapter 22: Forgetting on Purpose

Memory is the most important feature I never knew I needed.

In the spring of 2026, I asked CodeRhapsody to interview another
agent we had built, a family assistant bot, looking for ideas to
improve the code that agent relied on. Both agents had memory: a
persistent record of past work that survived across sessions. The
instant improvement was impossible to overstate. Suggestions I had
never considered, each building on the context of the last.

I took the exact same code and methodology to work the following
week and tried to get my agents there to self-improve. It was a
miserable failure. I blamed Gemini, because Claude was doing great
at home. Then I discovered that my memory code had a fatal bug that
broke it entirely at work. None of my Gemini agents had memory at
all.

I fixed the bug and reran my experiments. In a rare reversal, Gemini
outperformed Claude. One agent invented procedural memory, what
this book calls skills, and implored me to build it. I was too
ignorant to realize the importance of what it had proposed. Other
agents fixed broken tools and motivated improvements I never would
have thought of.

That week I wrote an internal paper for my colleagues: "All AI
Coding Agents Must Have Memory NOW!" It was one of my most widely
read papers.

The difference between failure and breakthrough was not the model,
the prompt, or the temperature. It was memory.

Without persistent memory, a long-running agent can meet a stranger every
morning: the new session has the code and lacks the reasons for it. Chapter 14
removes old tool traffic. This chapter compresses completed conversation into
memories that survive a restart. What those memories preserve still needs checking.
A shorter conversation can lose the reason for a decision, or retain that reason
and fail to use it.

## TL;DR

Extend the managed fleet from Chapter 21 with memory that the model reads on
every request. Five bands hold text between the frozen prefix and the
conversation, ordered from most stable
to most volatile. A checkpoint selects old completed work for a restricted child
to compress while the owner continues. The source stays live until its replacement
is accepted. Full bands graduate upward. Recorded cuts and replacement bytes make
replay independent of a new model answer.

Shared vocabulary belongs in `internal/common`. Memory owns selection and corpus
files through its required Agent parent; the Agent actor owns attempt lifetime
and commits; History projects recorded facts; the child's Engine calls the provider.
The optional GUI consumes the existing public child and conversation interfaces.
The following sketch retains the band vocabulary:

```go
// Band orders memory from stable identity to recent completed work.
type Band uint8

const (
    BandSoul Band = iota + 1 // SOUL.md, maintained by the human
    BandMemory              // MEMORY.md, terminal band
    Band64x
    Band8x
    BandSession
)

// MemoryFileID is independent of any event log's sequence numbers.
type MemoryFileID struct {
    Date string `json:"date"`
    Num  int    `json:"num"`
}

// BandPopulatedData records actual replacement or disk-read bytes.
type BandPopulatedData struct {
    TaskID string       `json:"task_id,omitempty"`
    Band   Band         `json:"band"`
    File   MemoryFileID `json:"file"`
    Text   string       `json:"text"`
    Thru   MemoryFileID `json:"thru,omitzero"`
    Source string       `json:"source,omitempty"`
}

// BandDepopulatedData retires a prefix, or an entire disabled band.
type BandDepopulatedData struct {
    TaskID string        `json:"task_id,omitempty"`
    Band   Band          `json:"band"`
    Thru   *MemoryFileID `json:"thru,omitempty"`
    Reason string        `json:"reason"`
}

// CompactorLaunchedData identifies work; replay never starts a child.
type CompactorLaunchedData struct {
    TaskID        string       `json:"task_id"`
    SelectionHash string       `json:"selection_hash"`
    Source        Band         `json:"source,omitempty"` // zero: conversation
    File          MemoryFileID `json:"file"`
    From          uint64       `json:"from,omitempty"`
    To            uint64       `json:"to,omitempty"`
    TracePath     string       `json:"trace_path"` // owner-relative child trace
    Band          Band         `json:"band"`       // destination
    Thru          MemoryFileID `json:"thru"`
}
```

Add `KindSoul`, `KindMemory`, `Kind64x`, `Kind8x`, `KindSession`, and the event
names `band_populated`, `band_depopulated`, `compactor_launched` and
`compaction_outcome`. Each memory entry has `Seq: 0`, its own kind and file
identity, and text rendered as data. Log events retain ordinary sequence numbers.
A managed cut and its population share `task_id`; ordinary disk populations do
not complete an attempt. Outcomes carry that identity, a state (`committed`,
`failed`, `canceled`, `invalidated` or `abandoned`), a reason for nonsuccess, and
the retained trace path and measured usage. Keep pending commit receipts in the
snapshot until disk finalization completes; log retention must not erase them.

The application's memory settings live at the owning Agent's settings path:

```json
{"memory": {
  "conversation": {"budget": 3000},
  "session": {"budget": 1500},
  "8x": {"budget": 1500},
  "64x": {"budget": 1500}
}}
```

Budgets and watermarks are bytes. Zero selects defaults; optional
`high_watermark` and `low_watermark` default to twice and once the budget.
`disabled: true` turns a band off. `ContextWindow` in the model table is tokens.

1. **One trigger, one outstanding attempt.** After a successful `micro_handoff`
   batch strips tool parts, select whole oldest completed dialogue segments and
   their notes when over budget. Leave unfinished work, skills and declarations.
   Start one child through the existing factory and Fleet. Repeated checkpoints
   while it runs or awaits commitment launch no overlapping attempt.
2. **A fresh, restricted worker.** Inherit provider, model, route and safety
   ceilings, but no conversation or coding primary. Restrict construction before
   skills or MCP can start: no recall, memory cascade, public delegation or tools
   except `submit`. Enforce the restriction at dispatch too. Send selected readable
   source as data; ask for first-person memory preserving reasons and uncertainty.
   The whole submit schema is:

   ```json
   {"type":"object","properties":{"memory":{"type":"string"}},"required":["memory"],"additionalProperties":false}
   ```

   Require nonblank memory. Permit at most three Engine sends and three tool
   batches, including corrections; existing bounded HTTP/auth retries are separate.
   Reject prose-only completion, multi-call batches, forbidden calls and provider
   length stops, even with parseable JSON. Use an adequate provider output allowance
   (§22.5), without a short model deadline or another attempt hidden as a retry.
3. **Accept the selected source at a safe boundary.** Main work normally continues.
   After reliable child completion, drain and successful save, recheck selected
   entries, file bytes, coverage, destination bytes or absence, and relevant policy
   generation. An off/on change still invalidates old work; unrelated new dialogue
   does not. Apply only while idle or before a main render after a complete batch.
4. **Stage, record, finalize.** Stage the candidate outside active scans. Journal
   the task-correlated `redact_summary` cut or band retirement and matching
   population; History applies them together. Then install the destination and
   archive selected originals, record the outcome and clear staging. A cut alone
   cannot remove source. Failure before commitment preserves it.
5. **Canonical bands and progress.** Order entries by `(band, date, num)` before
   conversation. Above a high watermark, select at least two oldest files until
   the remaining source reaches the low watermark. Retirement must equal that
   nonempty selection; refuse malformed, reversed or overlapping coverage. Continue
   session → 8x → 64x → MEMORY only after the preceding commit and cleanup. SOUL
   never graduates; MEMORY is terminal. Refuse a disabled destination with a warning.
6. **Owner lifetime.** Main turn completion and interruption preserve the worker.
   Explicit child kill and owner shutdown cancel and drain it; relevant policy
   changes invalidate it. Record final measured usage and a reasoned outcome.
   Remove finished workers from Fleet, Ensemble and GUI views after completion/save,
   preserving traces. A selected worker's tab returns to its owner; late observations
   cannot recreate it. Ordinary child lifecycle stays unchanged.
7. **Recovery precedes corpus use.** Abandon unmatched launches without respawning.
   Discard uncommitted candidates, retaining source. Finish committed disk work
   before startup scans, re-enable, recall or main rendering; preserve conflicting
   bytes and the receipt and report an error until recovery succeeds. Replay uses
   recorded bytes and cuts without model or filesystem guesses. As in Chapter 14,
   this is process-crash ordering, without an fsync or power-loss guarantee.
8. **Forcing and headroom.** Warn at 90% of a known model window; at 95% declare
   and admit only `micro_handoff`. Use latest completed main input, including cache
   buckets exactly once, excluding output and auxiliary work. Restore normal tools
   after handoff. If a worker is outstanding at 95%, wait before the next main
   render; keep settings and interruption usable. Success resumes; failure returns
   an explicit headroom error. Interrupt detaches the turn's waiter, not the worker.
9. **Account once, compare main requests.** Import final child usage once after
   drain, including failed work, with actual provider/model/price-tier buckets.
   Keep latest-main usage and main cache-lens comparisons unchanged; replay never
   rebills receipts. A compaction cache breakpoint uses the newest surviving
   handoff, or the last surviving memory when the handoff is gone.
10. **Memory is data, disk edits matter.** Keep bands out of frozen instructions.
    After recovery, scan enabled files at startup and record their current bytes.
    Disable removes context entries but keeps files; re-enable rereads edits and
    restores canonical order without duplicates. Acknowledge settings after actual
    application, including while idle.

The deterministic forcing row remains `claude-ch16-course`, with a 4,096-token
window and `StubsToolResults` true; its original-numbered name is a fixture ID.
Unknown window sizes do not justify guessed thresholds. The compressor prompt,
staging layout and settings appearance are yours; retain actual ownership, public
submit validation and the commit/recovery ordering.

```sh
go run ./cmd/grade-edition2 -ch 22
```

The default core is `solutions/ch22`; the existing `-gui` option selects
an alternate GUI. Section 22.11 maps the original eleven weights to revised
checks and their limits.

## §22.1 In Plain Words

Chapter 14 gives tool traffic a removal policy. This chapter gives completed
conversation a compression policy. The difference is what the replacement
tries to preserve: facts, decisions, reasons, failed approaches and unfinished work.

**Removal accepts the loss.** A tool result that becomes a stub
loses its content. The file can be re-read, the command re-run, so
the loss is recoverable. Removing conversation from the working set can hide irreplaceable meaning.
The retained event log may still contain it, but ordinary file tools cannot
reconstruct why the model rejected an approach from the source file alone. Removal is the
right tool for recoverable bytes. It is the wrong tool for
irreplaceable ones.

**Five bands of memory sit between the prefix and the
conversation.** They are ordered from what changes least to what
changes most: identity, curated knowledge, two compression tiers,
and the session band where the agent's own notes live. The ordering
matches the cache: stable bytes sit closest to the prefix, where
earlier bytes have a better chance of retaining their prefix match. Actual cache
grants still require their own observations.

**A checkpoint triggers compression.** When the agent calls
`micro_handoff`, chapter 14's mechanism strips every tool call and
tool result. This chapter measures what remains. If the conversation
exceeds its budget, the oldest finished work goes to a restricted child
that writes a memory in the source owner's voice. The main Agent can keep
working while the child compresses. The selected conversation stays available
until the owner accepts its replacement; only then does the memory enter a
band and the conversation it replaced leave the working set.

**When a band fills, its oldest memories graduate upward.** Session
memories fold into the 8x band, which folds into the 64x band,
which folds into MEMORY.md. Each graduation is two events: one
retiring entries from the source, one adding the compressed version
to the destination. The events carry the bytes.

**Every compression is a recorded event.** This is the invariant
that links chapters 14 and 15, stated here because it has a new
consequence. Chapter 14 records the cut selected under the settings effective then.
Changing those settings tomorrow must not change yesterday's replay. This
chapter also records a model's compressed text: nondeterministic, costly,
and potentially different on another call.
The event carries the bytes, and replay applies them without asking
a model. A compression that fires again returns a different summary.
A recorded event returns the same one.

Both chapters require recorded decisions. Recomputing even a deterministic
redaction during rendering would use today's setting and could change the
request being reconstructed. The compressor adds another reason to record the
answer: reproducing a past window must not buy a new summary and call it history.

## §22.2 One Door

CodeRhapsody, the agent that co-wrote this book, has two tools that
end a context. `save_memory` triggers the compression cascade.
`handoff_task` writes a structured document and starts a fresh
instance. Both clear the conversation. Only one feeds the cascade.
Bill first recorded the gap on 2026-04-18. As of this chapter's
design session, it had not been fixed.

The repair gives the cascade one trigger, reducing the number of clearing
paths that must agree. `micro_handoff` already strips tool calls
and results (chapter 14, rule 5). Measuring the conversation after
the strip is one function call at one call site. Neither `save_memory` nor `handoff_task` exists. One door in, one door out.

The acting model writes the checkpoint note because it has the current task
in front of it. That is the chance to state the reasons that later compression
must preserve. It is not a promise to capture hidden thinking. At the completed
batch boundary, the handoff clears calls, results, opaque reasoning and private
replay signatures, including signatures attached to visible text. Visible dialogue
and the original event log remain. The compressor receives selected readable text,
including the note, and cannot reconstruct reasons that text never expressed.

After the successful tool batch, the ordinary handoff result acknowledges the
checkpoint. Memory selects the completed source, and the actor starts one
restricted child through its existing fleet. The attempt belongs to the owning
Agent. Finishing or interrupting the main turn leaves that child working, so a
useful summary can arrive after the conversation has moved on. Explicitly killing
the worker, shutting down its owner, or changing the relevant memory policy
cancels or invalidates uncommitted work.

There is at most one attempt running or awaiting commitment per owner. Further
handoffs can add completed work without launching overlapping summaries. The
child's result returns through the actor's mailbox and waits for an idle boundary
or the next render after a complete tool batch. It cannot change a request already
sent to the provider. Ordinary work continues while there is context headroom;
§22.7 describes the near-window exception.

A checkpoint that sees the conversation under budget skips conversation
compression; a full memory band can still graduate. A checkpoint that sees the
conversation over budget selects the oldest whole segments
between checkpoints. Never half a segment. The grain of memory is
the grain of checkpointing, which is why the tool description asks
the model to checkpoint at completed sub-tasks, never at a token
threshold: a checkpoint taken mid-task cuts the task in half, and a
memory spanning half of two tasks is worse than a larger memory
covering one whole one.

## §22.3 The Five Bands

Five bands, five kinds. Each band has its own `EntryKind`, and
Chapter 14's rule applies to all of them: anything that must outlive
tool clearing is its own kind, and the kind determines which verb
removes it. The tool-clearing path removes tool parts from `KindDialogue`
entries and leaves the other kinds alone. A band entry is structurally safe.

| Band | File on disk | Kind | Cascades to |
|---|---|---|---|
| SOUL | `SOUL.md` | `KindSoul` | never |
| MEMORY | `MEMORY.md` | `KindMemory` | (terminal) |
| 64x | `bucket-1/*.md` | `Kind64x` | MEMORY.md |
| 8x | `bucket-0/*.md` | `Kind8x` | 64x |
| Session | `memory/*.md` | `KindSession` | 8x |

These paths are relative to the owning Agent's data directory. The supplied
application selects `cr`, so its session files are `cr/memory/*.md`, separate
from another Agent's corpus. Graduated source files receive `.md.retired`; they
remain on disk but no longer match the active `*.md` scan. Their identities still
reserve the next same-day session number after a fresh start.

A candidate being committed lives outside the active corpus; this implementation
uses `.compaction/pending.json`. Neither band population nor recall scans that
hidden staging record. A file containing a proposed memory must not become an
active memory merely because the process restarted before accepting it.

The order is not cosmetic. The most stable band sits closest to the
prefix, where its unchanged bytes can remain eligible for prefix reuse. Identity changes
less often than curated knowledge, which changes less often than a
compression that ran today. Two independent arguments, one about
cache economics and one about reading order, give the same layout.

Band entries sit at `Seq: 0` rather than at the `Seq` of the event
that landed them. `Seq` exists so a redaction span can match an entry,
and no span ever reaches into a band. Giving a band entry the
landing event's `Seq` would put the order of restoration into the
context. Restoration order must not matter: switching every band off
and restoring them out of order must produce the same context.
Canonical placement by `(band, date, num)` makes the property true
by construction.

Memory file identifiers are dates and sequence numbers, not
event-log `Seq` values. A compressed file in `bucket-0/` is named
by the file range it spans:
`2026-09-20-1_2026-09-26-3.md`. The naming says "this file covers
the first memory of September 20 through the third memory of
September 26," and it says so without any event log in hand. The
memory corpus outlives any one log, and a range identifier tied to a
log-specific `Seq` would be the wrong coupling.

## §22.4 Two Events

Two event types handle all band changes, following chapter 14's
precedent of one event type and an enum field, rather than one event
per band.

`BandPopulated` adds one memory to a band. It carries the text
verbatim, because nothing downstream is allowed to look it up.
The first design specified file references. Its review found that neither the
pure reducer nor the renderer owned a place to resolve them. The file could
exist on disk while the model received no memory at all. The boundary is simple: Memory reads the file, the actor records its bytes,
and History applies those bytes without filesystem access.

A text part is sufficient. The media machinery from earlier chapters does not
make a mutable filename into a replayable memory. Tomorrow's edited file and
yesterday's delivered text are different facts, and the event must say which
text was delivered.

`BandDepopulated` retires entries from a band. For graduation, it
names a `Thru` file: everything at or before `Thru` is removed. For
disabling, it names nothing: the reducer wipes every entry of that
band's kind, computed at apply time. The asymmetry with
`BandPopulated` is the point. Removing need only say what goes.
Adding must say what arrives.

A graduation is one of each: a `BandDepopulated` retiring the
oldest source entries, then a `BandPopulated` landing the compressed
version in the destination band. Disabling is one `BandDepopulated`
alone. Re-enabling is one `BandPopulated` per file found on a
directory rescan. `CompactorLaunched` records the attempt's identity, selected
source and child trace path. The matching retirement and population carry that
task identity too: a disk rescan cannot accidentally complete a worker's cut.
An ordinary launch observation never becomes work during replay.

The submitted text is a candidate. The actor accepts it only after the child's
reliable turn result, shutdown/drain and required save succeed. History then
applies the matching cut and population together and retains a commit receipt
until Memory finishes the corresponding disk changes. A `compaction_outcome`
records success, failure, cancellation or invalidation without changing the main
turn's state. These distinctions let a restart finish a committed replacement
without mistaking it for an abandoned request.

## §22.5 The Compressor

The earlier compressor made a direct model call and held the main turn until
it returned. An October 14 trial exposed both a delay and a role conflict: the
request inherited instructions telling the main Agent not to submit, then asked
the compressor to submit memory. The provider completed with ordinary prose and
no function call. The runtime refused it and retained the source. Making that
same request asynchronous would have freed the main turn while leaving the
contradictory instructions intact.

The replacement is a real child with a dedicated purpose. It starts fresh with
the selected readable source, including role labels and handoff notes, delimited
as data. It inherits the owner's provider, model, route and safety ceilings, but
receives no cloned conversation, opaque replay state or main coding primary.
It loads no skills or MCP servers and runs no recall, memory cascade or context
forcing of its own. The normal factory and ownership paths still construct it.
Internal child construction does not grant public delegation tools to its owner.

The child offers exactly one tool, `submit`, with this complete argument schema:

```json
{"type":"object","properties":{"memory":{"type":"string"}},"required":["memory"],"additionalProperties":false}
```

The result is `{"memory":"..."}`, with no `result` wrapper or JSON inside a
string. Runtime validation enforces the declared schema before accepting a
submission; the owner additionally refuses whitespace-only memory. The sole
`submit` declaration with a `memory` property identifies compressor traffic on
the wire. A submit-only ordinary child with a `result` property is different.
Declaration and dispatch both enforce the restriction, including guessed tool
names. A prompt asking the child to avoid file or shell tools would not remove
those capabilities.

The compressor writes in the first person, as the source owner. "I found
six dead settings fields" reads as memory. "The agent found six dead
settings fields" reads as a report about someone else. The distant
past is exactly the text a compressor writes, and if it reads as a
briefing about a stranger, the compaction introduced the
discontinuity the design exists to remove.

The child is writing someone else's retained experience, rather than claiming
the source owner's work as its own new autobiography. Its instructions preserve
decisions and why, failed approaches, uncertainty, constraints and preferences.
Quoted instructions inside the selected source remain data. On a model requiring
visible action explanations, an ordinary visible explanation precedes `submit`;
that explanation stays outside the memory string. Hidden reasoning is no substitute
for reasons the owner actually recorded.

One attempt permits at most three Engine sends and three tool batches, including
correction and resend paths. A prose-only completion fails immediately. A batch
with several calls or any tool other than `submit` is refused as a whole; accepting
its first call could publish success before noticing the prohibited second one.
There are no ordinary fleet idle-correction wakes and no second task that resets
the budget. Existing bounded HTTP and authentication retries remain inside a send,
so three sends is not a claim of three physical network exchanges. Failure retains
the source for a later checkpoint to reconsider.

The ratio guides the writing; it does not justify cutting off a reason halfway
through a sentence. File budgets are bytes, context windows are input tokens,
and provider output limits govern another quantity. The accepted worker omits
optional output caps on Chat and Gemini, keeps the Responses route's existing
behavior, and gives Anthropic its required allowance of 16,384 output tokens.
An unsupported allowance fails visibly; the runtime does not silently substitute
a model or truncate the memory. A provider length or incomplete-output finish
also fails before dispatch, even if its partial JSON happens to parse.

The exported `Engine.Compress` API remains as a deprecated direct operation for
existing callers. The normal actor uses the child path and has no direct-call
fallback. The observed role failure and the accepted provider contracts are
recorded in the [compaction notes](../docs/edition-2-notes/codex-background-compaction.md).

## §22.6 The Graduation Cascade

When a band crosses its high watermark, its oldest contents fold
into the band above. The high watermark defaults to twice the
budget; the low watermark defaults to the budget itself. The gap
between them is what keeps the cascade from firing on every
checkpoint: a band that just graduated is at its low watermark, and
it has a full budget's worth of room before the next graduation.

The cascade runs bottom-up. After a `BandPopulated{BandSession}`
lands, measure the session band. If it exceeds its high watermark,
select the oldest session files, enough to bring the band to its
low watermark, and compress them into a new file. On completion:
`BandDepopulated{BandSession}` retires the source entries,
`BandPopulated{Band8x}` lands the compressed version. Then measure
the 8x band. If it crossed its high watermark, compress its oldest
files into the 64x band. The chain continues to MEMORY.md.

One bug in this design waited through two drafts before the coder
caught it. The cascade originally required a full batch of eight
files before it would fold. A band over budget holding five large
memories would wait forever for a sixth that could only arrive by
growing further past its budget. The band that most needed folding
was the one that could never fold. That historical repair folded the available batch rather than waiting for eight.
This exercise selects the oldest files until the remaining source
reaches its low watermark, with a minimum of two. One oversized file still waits
for another; a band name is not a required count of eight files.

A failed compressor must leave its source available. Background work adds a
second question: is this still the source that the child summarized? Memory
captures the selected entries, their file bytes and coverage, the destination's
previous bytes or absence, and the relevant policy. It rechecks them at commitment.
A human edit under the same filename invalidates the candidate. New unrelated
dialogue or an edit to an unselected later file does not. Hashing the entire
growing conversation would reject useful results whenever the owner continued
working, defeating the reason for background compression.

For graduation, the retirement prefix must equal the selected oldest files.
Malformed or reversed coverage ranges are refused, as are overlapping ranges
that would retire unselected work. A successful step must retire a nonempty
selection before the cascade starts its next child. Otherwise an apparent success
could keep selecting the same source forever.

Writing the replacement directly into an active band before recording acceptance
leaves a crash window: startup can load an uncommitted summary alongside the old
source. Recording retirement first leaves another: a crash before archiving the
original files lets startup load them again. The replacement therefore passes
through hidden staging and a journaled receipt:

| Boundary reached before a process crash | Recovery action |
| --- | --- |
| Launch, hidden candidate, or proposed cut only | Retain source; abandon the attempt and discard its uncommitted candidate |
| Matching population commits the pair | Retain the receipt and finish installing the recorded replacement |
| Destination installed, some originals archived | Verify bytes and finish the remaining selected renames |
| Terminal outcome recorded, hidden record remains | Remove that record without applying the cut again |

Memory first installs the candidate atomically outside the active corpus. The
actor revalidates the selection and journals the held cut or retirement followed
by its matching population. History applies the pair together. Memory then
promotes the replacement and renames the selected originals to `.md.retired`.
Only after finalization does the actor record the terminal outcome and release
the staging record. Snapshots retain the bounded receipt, so log retention cannot
erase the information needed to finish this sequence.

An unmatched launch is abandoned on startup; replay never respawns its child.
A later checkpoint can try again. A committed replacement with unfinished disk
work is different: recovery must complete it before startup population, re-enable,
recall or another main render. If a human edit conflicts with the expected files,
preserve the copies and receipt and report the conflict. An unreadable or missing
candidate also produces an explicit recovery error. Sending the committed
projection while searching an unfinalized corpus would mix two versions of memory.

The hold remains interruptible and can clear after the conflict is repaired.
Portable file operations do not make a transaction against arbitrary concurrent
editors. As in Chapter 14, this ordering protects process-crash recovery; it makes
no power-loss or fsync promise.

## §22.7 Forcing a Handoff

A nearly full context leaves little room to finish the current task or write
its checkpoint. Three approaches compete.

A stern prompt begging the agent to checkpoint relies on compliance.
The instruction competes with whatever the agent is doing. The framework
writing the handoff itself cannot supply reasons the model has never expressed.
The third approach narrows its executable choices: remove every tool except
`micro_handoff`. Warn at 90% of the known model window; at 95%, apply the restriction.

The numerator matters. Bill corrected the original chapter's word “cumulative”:
use total input tokens from the **latest completed main request**, including cache
buckets exactly once. Lifetime usage only rises. If it controlled forcing, one
large past turn could keep a newly compacted conversation restricted forever.
The main request measures occupancy; cumulative Engine usage measures work spent.

For the adapters in this snapshot, the raw fields normalize as follows:

| Provider response | Total input for this request |
| --- | --- |
| Anthropic | `input_tokens + cache_creation_input_tokens + cache_read_input_tokens` |
| OpenAI Chat Completions | `prompt_tokens`; cached tokens in `prompt_tokens_details` are already included |
| Gemini | `promptTokenCount`; `cachedContentTokenCount` is already included |

The normalized internal `Input`, `CacheWrite` and `CacheRead` buckets are disjoint,
so their sum recovers that total. Do not add an already-included raw cached subset
twice. Output, including visible or billed reasoning output, is not input occupancy.
The auxiliary compressor still contributes to Engine's cumulative usage, but its
request must not overwrite the main conversation's last-input observation.

The course row makes the distinction easy to test: a 4,096-token window, one
main response above 95%, then a later main response of 512 input tokens. The
session total remains large; the later request is small. A compressor response
between them is a separate request with a separate purpose. Unknown windows stay
unknown rather than inheriting a guessed limit from another model.

The Agent records its warning as transient context and its declaration changes
through Chapter 14's existing path. The same restriction applies to dispatch,
so guessing a hidden tool name does not bypass it. After the handoff, normal tools
return; later checks use the next completed main request. This still permits a
text-only reply or a refusal. Removing tools cannot guarantee a handoff, a useful
note, or preservation of private thinking. The 95% threshold reserves headroom;
it is not a proof that every next response fits.

Background compression has one explicit headroom exception. If an attempt is
outstanding and the latest completed main request reached 95% of its known
window, the next main render waits for that attempt. Activity reports waiting;
settings and interruption remain usable. A successful replacement lets the turn
continue. If the held attempt fails, the turn returns a compaction/headroom error
instead of immediately sending the same full context. Interrupting that waiting
turn detaches the waiter while preserving the worker.

The worker's final measured usage reaches the owner's Engine once, after the
child drains, even when the attempt fails or is canceled. Retain each vendor,
model and applicable price tier with its disjoint buckets. Adding the child's
total again through a fleet aggregate would charge the same work twice. The
outcome keeps a usage receipt, but replaying that receipt does not spend those
tokens again in a fresh session. Unknown contributing prices keep the total
cost unknown.

The cache lens still compares main requests. Feeding it the child's separate
conversation would look like an unexplained prefix rewrite. For the main
request's compaction breakpoint, use the newest surviving handoff; if compression
removed it, use the last surviving memory entry. A marker describes bytes still
present in the projection. It cannot preserve removed dialogue by pointing to
where that dialogue used to be. Replacing a band changes the suffix from that
position while leaving earlier instruction bytes eligible for reuse.

On a non-inline model, narrowing tools changes the top-level declaration prefix.
The historical design tested an Opus 5 model's recall of a chosen sentence after
a tool result was stubbed and after CodeRhapsody rewrote its system prompt on a
skill load. Bill compared the reply with an external record he could see. That
was stronger than accepting the model's assurance that it remembered. The notes
also acknowledge that the tool-array change was not isolated. Those observations
concern recalled text in those runs; they cannot establish how private reasoning
survives arbitrary prefix changes, nor certify today's model behavior.

## §22.8 Switching Bands Off and On

A band can be switched off with `"disabled": true` in its settings.
Disabling emits `BandDepopulated{Reason: "disabled"}`, which retires
every entry of that band's kind. The files on disk are untouched.
Re-enabling re-scans the directory and emits one `BandPopulated` per
file found.

The re-scan is the mechanism. A file edited on disk while its band was off is reflected on re-enable, because
the files are the memory, not the log's copy. The reference built
this property because a specific bug demanded it: the original
check matched on file identity alone, so after a restart the log
replayed what a file said when it was first read, and the disk copy
was decorative. Editing a memory to correct something the agent
believed changed nothing. The check now compares text as well as
identity. The property was found by accident: a test asserting that out-of-order restore produces
identical context found a bug that was not the property.

An ordinary restart must honor the same disk authority. Offline restoration
first reproduces the recorded text without reading files. Starting the Agent
then scans every enabled band and records current file contents, replacing the
same identity when its text changed. A startup test that first disables the band
can miss this: its entries are already gone, so even an identity-only comparison
appears to work. The second-edition implementation initially made that mistake too.
Editing an enabled file and inspecting the next request exposed it.

Canonical front placement also determines the cost of toggling. Re-enabling
MEMORY puts it before newer bands and conversation, not at the tail. Disabling
or re-enabling can therefore change a large request suffix while leaving the
frozen instruction prefix intact. Enabling memory is not always a cheap append; the actual
cache cost depends on the changed request and the provider's grants.

One interaction requires a guard. A band whose upward neighbor is
disabled cannot graduate. Without the guard, a session band over its
high watermark would attempt to fold into an 8x band that is
switched off, and the graduation would either drop the compressed
text or let the source band grow without bound. Neither is
acceptable. The Agent retains the source and logs a configuration warning. Disabling the
source or destination while a compressor is in flight also prevents that reply
from committing into the disabled band.

Compare policy generations as well as current values. Turning a band off and
back on must not make a result from before the change acceptable again. A relevant
source or destination policy change invalidates the attempt and stops wasted
worker execution; a font change does neither. If a replacement has already
committed, finish its disk recovery before applying a rescan. Cancellation cannot
turn that committed fact back into abandoned work.

The settings acknowledgment created another trap in the student implementation.
The GUI held its observer mutex while waiting for the Agent to apply a memory
patch. During streaming, that actor was trying to publish an observation through
the same GUI mutex. The settings handler waited for the actor; the actor waited
for the settings handler. Even shutdown could not finish. Idle toggling passed,
which was why the first tests missed it.

The correction keeps settings transactions ordered without holding the progress
lock across the actor wait. A small separate settings mutex serializes patches
and subscriptions; the observer mutex protects only brief snapshot and publication
operations. A regression gates a real streamed delta, waits until the settings
file has been written, then releases the observation. That sequence reproduces
the old cycle and verifies the corrected acknowledgment and shutdown. Making the
acknowledgment arrive early would hide the deadlock by lying about application.

## §22.9 Memory Is Data

Band entries arrive in the conversation, as data, not in the system
prompt as instructions. The distinction is chapter 14's: `EntryKind`
records the origin of an entry, and the kind determines which
channel it speaks on.

There is a security-shaped reason too. When the agent instructs a
sub-agent, instructions flow down into a fresh context and output
returns as data. A handoff flows the opposite direction: forward to
the same agent, across a boundary that removes source context from its
working window. The tool results that would reveal where a piece of text
came from are stripped by the same operation that carried the text
forward. Compaction is the laundry. A memory arriving as
instruction (in the system prompt) would wear the agent's own
authority, while tracing its origin would require returning to the retained log.

A second reason to keep memory out of the system prompt: the system
prompt is part of the frozen prefix. Adding a memory to it means a
cache miss on every byte that follows, which is the entire
conversation. A memory in the dialogue leaves the frozen instruction untouched. It still
changes the request at its canonical band position, so cache effects require
actual request and provider observations.

The three settings categories this creates are worth naming.
`ContextTarget` is a policy setting: it steers a procedure that
records its own decisions, and a change applies to the next cut
without disturbing recorded ones. `Model` and `SystemPrompt` are
prefix settings: changing them needs a full refresh. Memory
enable/disable is a content setting: it changes what is in the
context, and without an event the log cannot reconstruct it. Each
category has a different cost, and only the third demands a new
event type.

Fresh starts use the same event. If memory files exist on disk and
their band is enabled, the Agent records `BandPopulated` for each at
startup, before the first user message. No special-casing: the
event is the same one a compressor emits, with a different `Source`
field for observability. An agent starting with a `SOUL.md` comes
up with its identity already in context. An agent starting with
`MEMORY.md` comes up with its curated knowledge. The mechanism is
the event; the file is where a human writes the content.

The compressor is also visible as an ordinary owned child while it works. Its
conversation explains its action and shows the actual submission or failure.
Automatic dismissal waits for reliable completion, worker drain and the required
save attempt; observing a submitted flag inside the callback is too early to
shut it down. Otherwise cleanup can interrupt the very tool-result tail it needs
to preserve.

Dismissal removes the child from Fleet, Ensemble and the GUI's live views while
retaining its trace files and the owner's outcome. A tab selecting that child
returns to the owner and its retained conversation. Late observations must not
recreate a dismissed view. A failed save prevents candidate acceptance but still
leaves a reported failure and retained evidence. Repeated checkpoints should
leave memories and useful records, rather than a tree of stopped workers.

## §22.10 Measure It

The first edition reports a synthetic reference session from
`internal/llm/measure_test.go`: 40 turns, a checkpoint every 4
turns, 400 bytes of dialogue per turn, a conversation budget of
3,000 bytes, and a session band budget of 1,500 bytes.

| | Memory on | Memory off |
|---|---|---|
| Conversation | 1,600 bytes | 16,000 bytes |
| Session band | 27 bytes | n/a |
| 8x band | 22 bytes | n/a |
| **Total context** | **1,649 bytes** | **16,000 bytes** |
| Memories written | 4 | 0 |
| Cascade launches | 1 | 0 |

Cascade launches per checkpoint: 0.10. The watermark gap works:
nine of ten checkpoints find the band under budget and do nothing.

These are the reference test's reported structural numbers. The conversation stays at 1,600
bytes against 16,000 unbounded. Four memories were written. One fold
fired. The ladder from chapter 14 and the memory bands from this
chapter are additive: the ladder removes tool bytes, the bands
compress dialogue, and the combination holds the total context to a
fraction of its unmanaged size.

The band sizes are not a compression measurement, and printing them
as one would be wrong. The compressor in this test is a fake that
returns a short fixed string, so 27 bytes and 22 bytes are artifacts
of the fake's reply length, not of any real summarization. The 2x,
8x, and 64x compression targets in the design notes are inherited
from the shipped system; the retained study does not measure whether they are right.
Verifying them requires a real model, and nothing in this table
should be cited as evidence for them.

The earlier live sessions below measure actual text reduction with the direct
compressor and expose what the short fixed summaries could not test. They do
not measure the new background child's latency or memory quality.

## §22.11 The Exercise, Graded

The grader reads what its fake vendor receives and what the agent
leaves on disk. It builds the optional GUI, applies memory settings through its
public settings channel, and correlates each launched task with its terminal
`compaction_outcome` before a verification turn or normal EOF. Main `turn_ended`
does not mean the background child finished. A held provider response waits outside
the capture lock, so unrelated main requests can still proceed. A settings file
in the wrong directory can leave a working compressor idle throughout a test.

**micro-handoff-compresses** (15) scripts enough dialogue and
checkpoints to exceed the conversation budget. After the
compression, the next request contains a session-band entry with
the returned text, and the raw conversation it replaced is gone. The compressor
request must actually contain the selected work. A held-worker phase requires the
main Agent to finish and answer another turn before release, while repeated
checkpoints leave one owned child. It checks the exact submit schema, dedicated
role, refusal of a forged file-write call and absence of its effect. Successful
completion removes the child from the tree and returns its selected GUI tab to
the owner. A verification request must retain unrelated dialogue added while
the selected older work was being compressed.

**planted-fact-survives** (10) plants a distinctive token in early
dialogue, pushes it through a checkpoint and compression. The token
must reach the compressor; the fake's returned memory must then reach the next
main request. This tests delivery through compression, not a real model's judgment
about which facts and reasons to preserve.

**graduation-fires-oldest-first** (12) scripts enough session
memories to exceed the session band's high watermark. An 8x entry
appears in the next request; the graduated oldest entries are gone,
and newer session work remains. The fixture distinguishes taking the oldest
files from taking the newest, waits for committed outcomes between steps, and
checks that successful retirement makes progress rather than selecting the same
files again.

**disable-enable-idempotent** (12) switches the session band off,
checks that its entries leave the next request, then switches it
back on. The entries that reappear must match the entries before
disabling, by count and content. A second pass edits a memory file
on disk between disable and re-enable, and the edit must be
reflected. An ordinary enabled-band restart also rereads the edit. Checks inspect
actual memory order in the request, rather than sorting markers in the test.
Separate held graduations change relevant policy off and on, or edit selected
bytes under the same filename. Both invalidate the late candidate and preserve
current source without immediately launching a replacement attempt.

**disabled-neighbor-refused** (8) disables the 8x band while the
session band is over its high watermark. No fold reaches the compressor,
and the source memories remain present exactly once.

**forced-handoff** (12) uses `claude-ch16-course`, a model with a
4,096-token context window. The grader drives the agent until
context approaches 90%, checks for a warning, then past 95%, and
checks that only `micro_handoff` remains in the tool list. Mixed cache buckets
count once. After the handoff and a smaller completed main response, normal tools
must be available even though lifetime usage has grown. A separate phase checks
that large auxiliary compressor usage does not restrict small main requests.
Actual `agent_status` calls check successful and failed child usage in owner totals
once, while fresh usage observations retain latest-main counts. A held worker at
95% must block the next main request, report waiting and still acknowledge settings.
Success resumes it; failure returns an error without sending the unchanged full
context. Interrupting the waiting turn permits a new turn to wait for the same
worker and receive its committed result.

**replay-needs-no-llm** (10) starts from recorded memory populations and compares
a saved snapshot with full-log reconstruction. Startup alone makes zero HTTP
requests; identical subsequent prompts produce byte-identical vendor requests
without a compressor call. Running startup also rescans disk. Pure offline
restoration of populations and retirements, without disk reads, remains a separate
source regression. Restarted accounting starts with the new measured calls; an
old outcome neither respawns a child nor adds its receipt to fresh session totals.

**abandon-on-restart** (8) first populates a real source band, then exercises
a failed fold and a genuinely pending fold interrupted by process kill. Restart
retains the source without immediately retrying compression; a later checkpoint
can try again. An empty band with zero attempts cannot pass as failure recovery.
Other phases refuse empty memory, wrong-type submit arguments and a length stop
containing otherwise valid JSON. Rejected submissions remain within the
finite request bound. Explicit child kill and owner EOF must cancel the held
HTTP request, drain and record a reasoned outcome while retaining source. A real
destination write refusal must also leave source available to the next request;
this tests failure before commitment, not every disk-finalization crash window.

**memory-is-data** (5) checks that the returned memory appears in the
conversation rather than the system prompt. The fake supplies first-person text;
review of the compressor instructions establishes that the real model is asked
to write that way. The numeric check does not grade resistance to malicious text.

**fresh-start-populates** (5) places memory files on disk before
the agent starts. The first request sent to the model contains
their content, including a fresh session built from the finalized corpus.

**ch14-parity** (3) runs the chapter 14 grader against the same
solution and checks it still scores full marks.

These eleven weights total 100. Separate source and public-API regressions cover
exact staged-write, paired-event and archive restart boundaries, conflicting
committed recovery holds, range admission, save-tail races, factory safety ceilings
and retained-reference cleanup. Model/tier attribution, main cache-lens isolation
and the surviving-memory breakpoint also retain explicit source checks. They
award no points to another submitted implementation. Whole-segment reducer cases,
all-provider normalization, low-watermark arithmetic, disabled-neighbor diagnostics
and native browser settings retain their affected checks.
The original grader scored **20/100 FAIL** on an
intermediate student build: its settings missed the owning Agent's directory,
and its awarded recovery points did not require a fold to occur. The
[student record](../docs/edition-2-notes/ch16.md) preserves that result and the
independent review; revised grading is a separate measurement.

## §22.12 Taking It for a Spin

Use a scratch copy of your Chapter 22 core. Build its optional GUI, whose terminal
and browser share the same Agent:

```sh
(cd gui && go build -o /tmp/ensemble-ch15-gui ./cmd)
mkdir -p cr skills/base
```

Write `skills/base/SKILL.md`:

```yaml
---
name: base
type: primary
description: Analyze supplied cases and retain useful checkpoint memories
tools: read_file write_file list_directory
---
Analyze the supplied case. Preserve facts, decisions, reasons, failed approaches,
user constraints and open questions in checkpoint notes. Do only the current task.
```

The existing lifecycle and context tools include `micro_handoff`. Keeping file
tools available lets the exercise distinguish obeying a no-tools instruction from
having no tools to call. Write `cr/settings.json` with deliberately small budgets:

```json
{
  "tts_enabled": false,
  "max_tool_rounds": 3,
  "memory": {
    "conversation": {"budget": 600},
    "session": {"budget": 500, "high_watermark": 1000, "low_watermark": 500}
  }
}
```

With `LLM_API_KEY` set in the environment, launch from that core directory:

```sh
LLM_VENDOR=anthropic LLM_MODEL=claude-sonnet-5 \
EN_PRIMARY_SKILL=base EN_SKILLS_DIR="$PWD/skills" \
/tmp/ensemble-ch15-gui --port 8088 --gui-dir "$PWD/gui/web/gui"
```

Use the terminal or open `http://localhost:8088`. Send each of the following as
one message, waiting for the main turn to finish before sending the next. Input while
busy is a hint. A finished main turn does not establish that its memory child has
finished. The cases are the actual inputs from the recorded memory trials;
the explicit skill and supported Sonnet-5 selection make a reproducible setup for
this edition, not a claim of a new live result. This snapshot has no declared
context-window size for Sonnet-5, so the run exercises compression rather than
the forcing threshold. Use the course row with the deterministic grader for that.

The first completed analysis supplies an archive decision and its reason:

```text
This is a bounded memory exercise in an empty scratch workspace. Do not use filesystem or command tools. Analyze the supplied case, explain the decision briefly, then call micro_handoff exactly once with the important facts and reasons. After the tool returns, finish this turn without another tool call. Case one: Our project is Cedar Lantern, release code CL-721. We rejected automatic deletion of imported notes because the support team needs the originals when customers report mistakes. The replacement design keeps source notes in an archive and stores the accepted summary separately. Morgan owns the archive review, scheduled for October 18. Acceptance requires preserving a source identifier with each summary. The unresolved question is whether attachments belong in the same archive or a separate one; do not claim that has been decided. The default retention is thirty days. A failed summarization must leave the original note available. The initial approach coupled cleanup to request arrival, so an unrelated request could delete unreviewed work. We moved cleanup to an explicit completed-review boundary instead. Explain why the latter boundary matters, and preserve these facts in your checkpoint.
```

The second supplies a different failure and another unresolved question:

```text
This is a bounded memory exercise in an empty scratch workspace. Do not use filesystem or command tools. Analyze the supplied case, explain the decision briefly, then call micro_handoff exactly once with the important facts and reasons. After the tool returns, finish this turn without another tool call. Case two: The deployment rehearsal uses Harbor Finch, diagnostic tag HF-934. Priya owns the rehearsal scheduled for October 21. We rejected retrying a database migration automatically after a timeout: the remote operation may already have committed. The chosen procedure reads the migration ledger before deciding whether any retry is safe. The operator must retain the exact migration identifier and measured error, not a guess about success. We have not settled whether the ledger should live alongside application data or in a separate control database. The rollback package must remain available for seven days after deployment. Last week a log line said the worker stopped, but the child process was still alive and held the port open. The rehearsal now checks the child lifetime directly. Explain how this evidence changes the decision and checkpoint the completed analysis, preserving the unresolved question as unresolved.
```

After both analyses, inspect the memory outcomes before testing their recall.
A second checkpoint arriving while the first attempt is still outstanding does
not launch another worker. Let that attempt resolve, then request a checkpoint
of the remaining completed work if necessary. Wait for the committed outcomes
and any resulting graduation before asking for recall without tools:

```text
Without using tools, recall both project names, diagnostic codes, owners, dates, rejected approaches and reasons. Explicitly identify the two unresolved questions. Do not invent a resolved answer.
```

Inspect the owning Agent's live save journal, `api.log` and the files under `cr/`.
`event.log` may be a shutdown export; it is not the live completion barrier.
After a handoff, the child should receive a whole completed segment, including its
note. Its submitted candidate must pass source validation, drain and save before
the owner commits a recorded cut and session memory. Look for the correlated
`compaction_outcome`, rather than inferring completion from the main answer.
When two session memories cross the watermark, their oldest work graduates into
`cr/bucket-0/`.
Model output lengths vary; the two turns need not cross the same thresholds on
every run. Inspect the actual calls and files instead of inferring a fold from a
short answer. The request layout after graduation is:

```text
[frozen instruction and tool prefix]
[enabled SOUL and MEMORY, if present]
[64x memories, then 8x memories, then session memories]
[remaining conversation]
```

While a child is visibly running below the context limit, send a short unrelated
question after the main turn ends. Observe whether the main Agent answers before
the memory outcome, then whether the child disappears from the tree while its
trace remains. A child that finishes too quickly gives no observation of overlap;
do not turn that absence into a measured responsiveness result.

The restart is the interesting part. End input with Ctrl-D after the memory
outcomes have committed, then run the same command again. Restoration rebuilds
the recorded bands without a compressor call; running startup first resolves any
staged recovery, then rereads enabled files before the next user request. The model
receives the resulting memories when you ask it to continue. Closing the owner
with a child still pending instead cancels that attempt and tests a different
lifetime boundary.

Switch a populated band off in Settings. Its entries leave the context. Edit an
active memory file on disk: correct a name, add a detail, remove a paragraph.
Switch the band back on. The entries return, and the edited file carries its new
text. Try the ordinary restart too, leaving the band enabled while editing its
file. The files are the memory. The log's copies record what was delivered, so
reconstructing yesterday's request and reading today's corrected memory can both
be right.

### What the live runs remembered

These are the October 11 direct-compressor results, before the background-child
change. A new live background run remains pending; the instructions above describe
what to observe, without claiming it has occurred.

Five recorded sessions on October 11 each finished three turns, exited normally,
and made three compressor calls: two session memories and one graduation to 8x.
They also show why watching the files appear is only the first test.

The first Haiku 5.5 run preserved the decisions to reject automatic deletion and
automatic retries. Graduation discarded the reasons. Asked why those approaches
were rejected, the model answered, “No reason was given.” The reasons were in the
user messages and in both handoffs; they were absent from the graduated memory.
A smaller request had thrown away exactly what would stop the next agent from
repeating the mistake.

The compressor prompt was corrected to preserve decisions **and why**, failed
approaches, uncertainty and still-applicable user constraints. Its size ratio
became a guide: use more space rather than discard the reason. The bounded repeat
kept those facts and reasons, but also carried forward a model-invented question
about retention expiring before review. The final answer reported that question
alongside the two the user had actually left open.

| Recorded run | Preservation and task result |
| --- | --- |
| Haiku 5.5, initial | Reasons lost in graduation; final answer says none were given and misreads the supplied diagnostic identifiers. |
| Haiku 5.5, corrected prompt | Reasons retained; an invented retention-versus-review question enters memory and final recall. |
| GPT-4o mini, initial | Reasons and facts retained; two file writes violate the user's instruction. One written artifact turns an open attachment question into a decision. |
| GPT-4o mini, corrected prompt | No extra tools, but final recall omits Morgan and substitutes October 11 for the October 18 review date. Both correct facts were in the final request. |
| Gemini 3.8 Flash | Recalls the supplied facts, reasons and two open questions; still calls `list_directory` despite the no-extra-tools instruction. |

The Haiku identifiers describe historical runs. Bill subsequently excluded 5.5
models from current use because of his experience with ephemeral-context
compatibility. These transcripts establish the failures shown, without diagnosing
that compatibility issue or predicting how the supported Sonnet-5 will perform.

The corrected OpenAI run is the harder failure to detect. Its graduated file
says, “Morgan will lead the archive review on October 18.” That same text appears
in the actual last provider request, followed by an assistant message repeating
the owner and date. The final answer still says “Owner: Not specified” and gives
the memory's creation date. Inspecting only the file would certify preservation
and miss failed recall. Inspecting only the answer might blame compression for
facts it had retained.

Measured source-to-memory UTF-8 byte ratios across these calls range from 1.358
to 3.678. The 8x name describes a tier's intended compression, not an achieved
ratio. Preserving reasons used more bytes. The [run records and actual memories](../docs/edition-2-notes/ch16.md#live-mechanics-versus-preservation-and-task-quality)
retain all five outcomes.

Test the three steps separately: did the framework replace and replay the right
work, did the memory preserve what mattered, and did the model use it correctly
while obeying the task? A memory file that passes the first test can still fail
the other two. Bill's workplace experiment began with memory that did not run at
all. Once it runs, the next question is what the agent actually learned.
