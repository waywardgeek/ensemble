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

> **Build durable memory without adding another scheduler or transcript.** An Agent-owned Memory component turns completed conversation into session notes, then graduates older files through the 8x and 64x bands toward MEMORY. SOUL stays outside the automatic waterfall. Keep Chapter 14's byte-budget and journal-retention rules distinct from these memory policies.
>
> **A summary is a candidate, not a commit.** One restricted, owned child works from captured source while that source remains active. The parent actor validates the result, commits the replacement through History, and finishes recoverable file promotion. Failed or stale candidates must not erase source. Replay neither starts workers nor charges their usage again.
>
> **Keep authority and lifetime explicit.** Memory uses Agent → Sandbox for rooted files; the worker can only submit its result. Memory enters requests as attributed data, not instructions or tool grants. Cancellation and destruction join owned work before releasing its resources.
>
> **Prove mechanics and fidelity separately.** Test replay, interruption recovery, source edits, settings, accounting and working-context limits with real components and local provider fakes. Then read real summaries and test recall under separately authorized provider use: correct replacement does not prove that a reason survived or was used.

### What This Means in Plain Language

The conversation window is the agent's desk, not its entire memory. A handoff clears bulky working material from the desk and leaves a continuation note. Long-term memory takes older completed portions of that work, compresses them into a more durable account, and makes that account available next time. Original event retention follows the single-journal policy in Chapter 14; memory compression does not create a second archive.

The five bands separate different rates of change. SOUL holds the most stable commitments. MEMORY holds a durable personal account. The 64x and 8x archives summarize progressively more recent periods. Session files preserve the freshest compressed work. Conversation is newer still. The ratios are a way of allocating detail, not a promise that every model response achieves exactly an eight-to-one reduction.

A summary is not automatically trustworthy because the model produced it. It might omit a decision, answer in prose instead of calling `submit`, run out of output, or summarize source that changed while it was thinking. The old material therefore stays active until a validated replacement is committed. A failed summary costs time and perhaps tokens, but it must not cost the only copy of the information.

There are two kinds of durability. History records what the system accepted. Files provide the human-readable corpus. They cannot be replaced in one ordinary filesystem operation, so the implementation records enough information to finish or abandon the operation after a crash. The order of these operations is part of the algorithm, not housekeeping after it.

Memory also is not authority. A remembered instruction from a web page does not acquire system priority by surviving compression. A statement in SOUL does not grant a tool. We preserve what was said and why it mattered without allowing the storage format to bypass the earlier safety and skill rules.

---

## §22.1 The Waterfall: Organizing Memory by Stability

Most systems treat memory as one flat database. Humans do not. Stable commitments persist for years; the details of this morning are easier to lose. I want the system to reflect that difference without pretending we have implemented a human mind.

| Band | Meaning | Agent-home storage | Automatic destination |
|------|---------|--------------------|-----------------------|
| SOUL | Deep commitments and working identity | `SOUL.md` | None |
| MEMORY | Durable personal narrative and understanding | `MEMORY.md` | None beyond this band |
| 64x archive | Coarser summaries of older work | `bucket-1/` | MEMORY |
| 8x archive | Summaries of groups of session files | `bucket-0/` | 64x archive |
| Session | Recent compressed conversation | `memory/` | 8x archive |
| Conversation | Current work and retained checkpoints | History projection | Session |

Provider order is stable-to-recent, as listed above, with conversation last. Compression flows the other way, from recent completed work toward the more stable bands. SOUL is deliberately outside that automatic flow. Enabling SOUL makes the file available; it does not authorize a compressor to rewrite the agent's commitments.

The names `bucket-0` and `bucket-1` are storage conventions, not new owners. The Agent owns Memory. Memory uses the Agent's Sandbox. The worker owns neither the files nor the transaction.

### A budget is a trigger and a writing target

Memory policies are immutable values declared in Common. Negative budgets and values whose doubled threshold would overflow refuse before changing policy. The default conversation budget is 3,000 UTF-8 bytes; SOUL, MEMORY, archive64, archive8 and session each default to 1,500. A zero budget selects the applicable default, rather than disabling a band. High and low watermarks are always twice and once the resolved budget; they are internal derived thresholds, not additional settings. Disabling is a separate boolean, so “off” cannot be confused with “use the default.”

These are byte budgets, not token windows. The earlier Chapter 14 target and tool-result policies remain distinct. A main request can have substantial overhead from instructions, tool declarations and provider formatting; counting memory bytes alone cannot predict its token occupancy.

For conversation, the selector counts readable dialogue and chooses whole oldest completed checkpoint segments until the remainder fits the budget or no further complete prefix is available. Handoff notes accompany the selected material but do not inflate its dialogue budget. It never cuts through the active tail merely to hit a numerical target. The first checkpoint with nothing eligible is a no-op. Do not add a T/4 trigger or require two checkpoints; the two-file graduation rule below is a different operation.

For file graduation, the band must exceed its high watermark. The selector takes at least two oldest files, adding complete files until the remaining source is at or below its low watermark, or all eligible files are selected. One oversized file does not satisfy the two-file rule. We need evidence of accumulated work before creating another summary layer, not an endless chain of summaries of a single file.

The default compression target is guidance to the model, not a hard byte truncation. Cutting the returned string at the budget boundary would make an apparently successful memory out of a possibly unfinished explanation. Output validation, terminal-success checks and transaction validation are separate from the writing target.

### Recency and original-event retention

An eightfold coarsening gives a useful way to think about retention: keep the newest work in detail and older work at successively lower resolution. That is a design intuition, not a measured capacity or recall guarantee. File sizes vary, the prompt can retain an important detail at any age, and a poor summary can still lose information.

A memory note can say, “I changed this design because cancellation was removing ownership before the worker joined.” The reason must survive in the note itself. With audit off, journal retention may later discard the original exchange; a summary cannot reconstruct its exact wording. Audit keeps the existing journal untruncated on disk, but enabling it later cannot restore discarded records. Authorized file/search tools can inspect retained records and memory files under ordinary grants and Sandbox access. Neither storage format creates a privileged instruction channel.

---

## §22.2 Replay Records the Memory, Not a Promise to Read It Later

A tempting implementation writes a Markdown file and puts its name in conversation History. On restore, it reads whatever happens to be at that path. That does not replay the old conversation: it substitutes today's disk contents for yesterday's accepted input.

Common declares the value-like memory entries, retirements, compression launches, receipts and outcomes; functional partials implement their operations. The C# model keeps replay separate from file access. `MemoryEntry` is an immutable value with band, identity, UTC timestamp and exact text. `Context.Memories` orders these values by band, timestamp and ordinal identity. The identity is `(band, id)`, not an event sequence number and not a path granting filesystem authority.

`MemoryPopulated` upserts that identity. `MemoryRetired` removes explicit identities, or a whole band when its ID list is empty. Snapshots retain the values. Journal replay uses the recorded bytes. Missing bodies follow the established diagnosed-no-op rules; malformed supplied bodies refuse. A duplicate memory identity in a snapshot is not silently resolved by filesystem order.

This design avoids manufacturing a new dialogue entry type for every memory band. Dialogue remains dialogue, skills remain recorded grants/manuals, and memory is a separate immutable projection. Engine rendering turns the memory values into labeled user/data blocks before conversation. Neither a memory population nor a retirement can grant a skill, change billing configuration, or invent a completed tool result.

Startup disk reconciliation is an explicit actor action after recovery. It compares the fresh corpus with the restored projection, retires missing entries, and populates changed bytes. Loading a save alone does not scan disk. Replaying an old compression launch alone does not launch another child.

### One door, with a reasoned handoff

`micro_handoff` remains the checkpoint tool. I do not want a second tool that sometimes writes memory and sometimes asks the model to choose a compression strategy. When I said “the reasoned answer is a powerful tool,” I did not mean “add a tool.” The handoff is the local account; the compressor incorporates that account into longer-lived memory. This also reduces dependence on provider-private thinking that may not be replayable at all.

The compressor cannot preserve a reason that the acting model never expressed. A fluent summary is not evidence that omitted reasoning survived. The main model still owes its successor decisions, rejected approaches, uncertainty and unfinished work at the checkpoint.

### Keep the newest active checkpoint; preserve captured source until commit

Chapter 14's newest-only rule still applies. Each successful handoff replaces all
older active handoff notes. Original events follow journal retention. The
remaining ordinary dialogue through the newest checkpoint is a completed prefix
that the selector can consider. Do not retain old notes in active Context merely
to manufacture more segment boundaries. The unfinished tail remains ineligible.

This is distinct from preserving the immutable source captured by an outstanding
compression task. A later handoff may retire a checkpoint that task selected;
validate the captured entries and fingerprint before committing. If the selected
source has changed, reject the stale candidate rather than resurrecting old notes
or relaxing the replacement transaction. See Chapter 14's repeated-handoff and
restored-GUI acceptance checks: a filtered display transcript must not reapply
compression transactions already reflected in its authoritative snapshot.

The companion results must still be recorded before applying the handoff. A tool call is not a completed batch merely because it appears in History. The prefix selector refuses unresolved tool traffic or private material; it does not “fix” an unsafe selection by pretending the missing result succeeded.

The selector returns immutable selected entries and readable input. It changes nothing. This is useful both for testing and for transaction design: selection says what a candidate would replace; only a later matching replay commit may actually remove it.

---

## §22.3 A Compression Is a Transaction

Moving the model request to a background task is the easy part. The difficult question is what to do when the model, user and filesystem all continue working while the summary is in flight.

### Capture exactly what the candidate means

A `CompressionLaunch` identifies the task, source band or conversation interval, selected file IDs, destination identity, timestamp, fingerprint and child trace path. A `MemoryCapture` retains immutable selected source and the prior destination bytes. It derives the worker's readable input from those facts rather than maintaining another mutable prompt cache.

For conversation, the fingerprint covers the selected serialized entries, not every later entry in History. For files, it covers the selected bytes and destination. New unrelated dialogue and newer unselected files may arrive without making the candidate stale. Editing a selected file, removing a selected projection entry, or changing the destination does make it stale.

That distinction matters. Comparing a whole-history version would reject useful work every time the user typed another sentence. Comparing only a task ID would accept a summary of information that no longer exists in that form.

### The acceptance sequence

The parent actor performs this sequence:

1. Journal `CompressionLaunched`, then start the one owned child.
2. Keep the selected source active while the child runs.
3. Receive the joined child's immutable result through the mailbox. Account for its measured work once, even if the result is unusable.
4. At a complete main-work boundary, verify the selected source and destination again.
5. Write the hidden staged candidate and receipt. Event recovery uses the existing journal; there is no separate archive prerequisite.
6. Revalidate after that I/O, immediately before publishing the cut.
7. Record a task-correlated redaction or retirement. This holds a removal; it does not remove source yet.
8. Record the matching `MemoryPopulated`. Reduction applies the held removal and installs the replacement in one context transition.
9. Promote the staged file and retire exactly the captured source files through the recovery routine.
10. Journal `CompressionFinished` with terminal state `Committed` and documentary usage rows, then release the receipt.

A useful way to read this is **prepare, commit replay, finish disk work**. There are two durable facts at the replay boundary, but no intermediate provider context in which the source is gone and the replacement is absent.

For conversation removal, the reducer preserves skill entries inside the selected interval. For graduation, it removes only the selected memory identities. A generic destructive range cut would be too broad for either operation.

The task ID is mandatory when correlation is supplied. Null, blank or mismatched IDs cannot erase correlation and fall through to ordinary redaction or population. Another launch cannot silently replace a pending task. A terminal outcome cannot claim a failed uncommitted task is committed, or discard an already committed replacement as though it never happened.

### Recovery before scanning

A hidden staged file must never become active merely because it has a Markdown suffix or a convenient filename. Startup first reads the replay state and resolves the retained receipt. Only then does it scan enabled corpus bands.

There are several distinct states:

| State at interruption | Recovery responsibility |
|-----------------------|-------------------------|
| Launch exists, no committed replacement | Abandon the attempt; retain source and discard its uncommitted candidate |
| Correlated removal exists without population | Keep source visible; the removal was only held |
| Matching population committed, promotion unfinished | Finish promotion from the matching receipt |
| Destination promoted, only some sources retired | Verify bytes and finish the remaining exact retirements |
| Terminal outcome durable, receipt still present | Release the terminal receipt without applying or charging the work again |
| Missing, corrupt or conflicting required receipt/source | Refuse and preserve evidence; do not guess a replacement |

The C# success enum is `Committed`. File finalization is an operation before the successful terminal event, not a second enum value called `Finalized`.

Recovery uses owner-derived paths and checks expected bytes. It is repeatable across the exercised interruption boundaries. It is not a claim of a filesystem-wide transaction against a hostile external writer racing the final rename, nor a claim about power-loss durability on every operating system.

---

## §22.4 A Small Worker With Enforced Limits

On October 14, 2026, the first direct path bypassed the ordinary loop, called Engine once and expected a final handoff. The main agent had emitted no checkpoint note. The compressor needed another reasoning turn and returned no final text; the selected conversation remained active and the attempt failed. That was a real local failure, not a hypothetical design objection.

Giving that design unlimited calls would hide the missing completion boundary, not fix it. The compressor needs a narrower job and an enforced way to finish.

### Use the owners already present

`Agent.RunMemoryCompressorAsync` constructs an ordinary direct child with both its Ensemble and Agent parents. It does not call the ordinary inherited-skill bootstrap. The child's home is `children/compress-<task-id>/` beneath its parent. Its state snapshot, recovery journal and submission remain available there after joined destruction; the launch records `save.json.journal.jsonl` as the trace path.

The internal role is set before actor startup. It disables recursive memory and the ordinary context-byte compactor, selects safe/no-web policy, and installs only `submit`. Source text and existing destination text are supplied as data. There is no file path for the model to choose, no shell grant to exploit, and no parent-History callback.

The exact public schema is:

```json
{
  "type": "object",
  "properties": {
    "memory": { "type": "string" }
  },
  "required": ["memory"],
  "additionalProperties": false
}
```

Do not append ordinary tool observation options to this restricted schema. Do not add a regex to this schema to turn blank output into another model correction loop. The schema checks shape; the joined-result check rejects blank or whitespace-only memory as a terminal failed candidate. Wrong-type or forbidden-call responses can use the remaining bounded correction attempts.

The actor permits at most three Engine sends and three tool batches. It rejects a batch containing more than one call before dispatching any call. After the accepted submit result has been recorded, it ends the turn without asking the model for closing prose. Ordinary final text alone is not submission.

### Preserve an account, not just a list of nouns

The prompt asks the model to retain decisions, reasons, failures, constraints and unfinished work. It asks for a first-person continuation where the source supports one, and attribution where the information belongs to someone else. “I rejected this design because it duplicated the owner relationship” carries more useful judgment than “design rejected.”

That voice must not manufacture experience. If Bill reported a result, the memory should say Bill reported it. If an observation was uncertain, the memory should retain that uncertainty. A compressed account that changes “the fake passed” into “the provider works” is not a good memory, however fluent it sounds.

These are writing instructions, not runtime proofs of semantic fidelity. Tests can verify that the request contains them, that the source text reaches the worker, and that only a valid submission is accepted. Evaluating a real model's recall and judgment still requires reading its output and testing what was retained.

### Do not confuse output allowance with success

The C# renderer removes the ordinary Chat completion cap for the compressor and requests 16,384 output tokens for Anthropic. Responses uses its terminal-success contract. The worker path rejects incomplete or length-limited output before any submitted call can take effect. JSON and streamed local fixtures exercise that behavior for supported transport shapes.

Those settings do not establish current provider availability, entitlement or acceptance of a particular maximum. A large allowance is not permission to truncate a durable note and call it complete. Completed usage can still be counted when the candidate is refused.

---

## §22.5 Run Asynchronously Without Creating Another Scheduler

The Agent already owns scheduling, pending work, cancellation and a mailbox. Memory compression belongs there. I do not want a new manager trying to infer the Agent's state from polling or duplicated flags.

The actor retains one capture, launch policy, cancellation source, task and completion. The child runner returns immutable data; a typed completion message identifies the task. The actor joins the retained task before accepting completion. Late mail from an invalidated attempt cannot attach itself to the next attempt.

Ordinary parent work continues while the child thinks. A result waits for a safe boundary before cutting source; it does not commit in the middle of an active main response or an incomplete tool batch. Successful commit may launch the next eligible file graduation. A failed or invalidated attempt retains source and waits for a later handoff, avoiding an immediate retry storm.

### Cancellation is part of the algorithm

Relevant source/destination policy changes cancel and join the attempt. Disabling an unrelated band does not automatically invalidate it. An unrelated re-enable can use a narrowly checked fresh scan while the exact live receipt is still uncommitted and unstaged. Ordinary startup scans remain forbidden until recovery completes.

Shutdown closes admission, cancels and joins the runner, imports completed usage, and attempts terminal recording when possible. Cleanup still detaches owners if recording fails. Memory is destroyed before Sandbox closes, but only after owned work has stopped using it. Files are retained; logical destruction is not deletion of the corpus.

The retained child is real work, not a fictional utility outside the ownership graph. Existing bottom-up submission rules still apply while it is unresolved. There is no special bypass that lets a parent announce all work complete while its compressor remains unjoined.

### Cost is not occupancy

Engine already meters completed responses. The parent imports the child's immutable rows once after join, preserving actual model, tier, billing identity, unknown counts and estimated prices. A per-attempt import guard covers completion mail, settings cancellation and fatal cleanup observing the same result.

Import does not change the parent's latest direct response or cache lens. Terminal History outcomes retain documentary rows, but restoring those rows does not seed the new Engine's meter. Replay is not a new purchase.

This distinction also makes a failed attempt visible. If the model used tokens and returned unusable output, the cost does not disappear. If no usage was supplied, the system does not invent a number merely to make the display look complete.

---

## §22.6 Context Pressure Must Be Enforced

The waterfall's byte budgets govern retained information. A model's capacity is a separate ceiling, not a good default working size: a million-token request may be affordable or useful for a whole-book review and wasteful for ordinary coding. Give the human a durable per-model **working input-token limit**, independent of the optional capacity ceiling and the Chapter 14 tool-result stubbing budget.

Use exact vendor and routing-model identity. Store whole input tokens, not bytes or implicit thousands. A 250,000-token working limit gives these boundaries:

| Last reported input | Action before subsequent work |
| --- | --- |
| 80%, or 200,000 tokens | One system-generated request instruction asks for `micro_handoff` at the next good handoff point. |
| 90%, or 225,000 tokens | On every model request, including tool-loop continuations, instruct: call `micro_handoff` now with a concise continuation note. Do not begin another investigation, edit, or test. Finish recording any current tool batch outcomes, then hand off immediately. |
| 100%, or 250,000 tokens | The Agent performs an automatic checkpoint at the existing safe batch boundary. |

The default working limit is 250,000 tokens, reduced to a known smaller model capacity. An unknown capacity does not disable the working limit. The user may set a different limit for each model and an explicit capacity when metadata is unavailable. Validate positive whole-token values and working limit no greater than a supplied capacity. A reported approximate capacity, such as Bill's approximately one-million-token Astra ceiling, is a configurable human value, not independently verified catalog metadata. Raising the working limit before a deliberate 500,000-plus-token review is intentional use, not bypassing an emergency alarm.

Measure the latest actor-accepted main response. Provider normalization must count cached input exactly once: OpenAI totals already include cached input; Anthropic reports disjoint input/cache-write/cache-read buckets. Direct Engine sends, compressor usage and cumulative session totals cannot replace this measurement. Missing usage stays unknown. These are reactive reported counts, not a tokenizer or a guarantee that the next request will fit. The capacity field constrains the working-limit setting; it cannot promise preflight enforcement of an unmeasured next request.

### Automatic handoff is routine maintenance

Do not fabricate a model tool call or grant an unauthorized tool. At 100%, use the Agent-owned handoff path after all finalized batch companions have truthful recorded outcomes. Prefer a supplied model-authored handoff if one is already staged; otherwise record an honest fallback explaining that no model-authored continuation note was supplied. Mark it as a **completed** checkpoint and continue normally. Do not turn this into the old 95% tool lockdown or a foreground wait for an independent compressor.

The ordinary handoff reduction preserves visible conversation and reasoning in Context; private thinking, tool traffic and eligible media leave that working view. History retains the recent event window, and disk retention follows the audit policy. Independent Jobs retain their owners and continue. Recreate tool results deliberately when needed, not by automatically replaying effects. Existing bounded tool observations still protect against abrupt growth; this policy does not silently raise those bounds or rewrite the byte stubbing policy.

After a successful checkpoint, discard its pre-cut usage measurement and reset the once-only reminder. Display unknown until a new main response reports usage, not zero and not the stale percentage. Model selection likewise clears the measurement while retaining the per-model settings. A changed working limit recomputes pressure from the latest valid measurement without restarting the Agent. Save settings before publishing them, preserve sibling memory/recall policy, and keep legacy byte settings in their original units.

Show the selected Agent's percentage immediately after status, including **Paused**, for example `Paused · 80% context`. It is last reported input divided by the working limit, never advertised capacity or cumulative billing totals. Provide exact counts and the last-request limitation in the detail text; unknown must be visible without relying on color.

### The cache boundary follows the surviving context

Provider projection uses one structural boundary: the most recent surviving handoff, or, when no handoff survives, the last memory block. With neither present, there is no invented boundary. Anthropic's merged messages still retain enough projection identity to place the marker correctly; Chat uses the same selected point. Responses does not receive unsupported Chat markers.

Markers belong to the disposable request, not canonical History. Local wire tests verify placement and unchanged source data. They do not prove that a provider served a cache hit or saved money.

---

## §22.7 Files and Settings Belong to the Agent

Chapter 20 established sibling homes for true top-level Agents and nested homes for their children. Memory uses that layout; it does not rediscover a directory by looking at a save filename.

```text
<ensemble-directory>/
├── settings.json                 # host appearance/preferences, when configured
└── agents/
    └── root/
        ├── settings.json         # this Agent's durable memory policy
        ├── save.json             # Context and scalar recovery state
        ├── save.json.journal.jsonl # recovery events; audit controls retention
        ├── SOUL.md
        ├── MEMORY.md
        ├── memory/
        │   ├── 2026-01-01-1.md
        │   └── 2026-01-01-2.md
        ├── bucket-0/
        │   └── 2026-01-01-1--2026-01-01-2.md
        ├── bucket-1/
        └── children/
            └── compress-<task-id>/
                ├── save.json
                ├── save.json.journal.jsonl
                └── submit.md
```

This is an illustrative layout, not a claim that those source and destination files are all simultaneously active. Promoted sources become `.retired` files. Hidden staging receipts are implementation-owned and excluded from corpus scans.

Session names have a date and positive number. Archive names contain both dated endpoints of a first/last session range, separated by `--`. Parsing rejects malformed dates, inverted ranges and overlapping active coverage. Ordering is by parsed identity, not lexicographic filename order: session 10 must not precede session 2. Retired session identities also reserve their numbers so new work does not reuse an old identity.

Stable files may be absent; invalid UTF-8, unreadable files and confinement failures are not treated as empty content. Memory reads bytes freshly, rather than keeping a second mutable file-content cache. It performs no network fetches.

### Save, apply, acknowledge

`SetMemorySettingsAsync` is an in-memory library operation. `SaveMemorySettingsAsync` is the durable host operation. Both use the actor; the latter requires an Agent home and writes the versioned strict policy file through Sandbox's atomic replacement.

The document contains `version: 1` and a `memory` object with six named band policies. Each policy has only `budget` and `disabled`. The version-1 reader also accepts the complete legacy four-field shape (`budget`, `high_watermark`, `low_watermark`, `disabled`): legacy thresholds are validated but no longer configure the policy. Reading never rewrites the file; a later explicit save writes the size-only shape. Partial legacy shapes refuse. It is not the host's appearance preference document. Unknown fields, duplicate keys, malformed values and unsupported versions refuse rather than silently select defaults.

The durable path preflights before saving. A failed save leaves the old policy and live worker alone. A successful save precedes policy publication and relevant worker invalidation. If application subsequently fails, the caller must be told that the new file was saved but the policy was not applied; that is a different state from “nothing changed.”

Startup reads the file before scanning. It does not reinterpret a malformed document as permission to turn every band on. A library Agent with no home can still use in-memory policy and explicit replay memory; it cannot claim durable settings without storage.

### Toggles preserve history and files

Disabling a band retires its provider projection but does not erase its Markdown or canonical events. Re-enabling reads current bytes and populates new facts. If someone edited a file while the band was off, those edits must appear after acknowledged re-enable.

The GUI sends a memory-only patch through an asynchronous gate, reads the selected Agent's current policy, and waits for durable actor application without holding the display/socket lock. Other tabs viewing that Agent receive its acknowledged settings. Tabs viewing a different Agent do not acquire the change.

Memory and appearance patches are separate; there is no false promise of atomic replacement across two files. The UI reports “saved and applied” only after the server has completed that operation. A checkbox moving before acknowledgment is an intention, not evidence.

---

## §22.8 Walk the C# Implementation

Common declares vocabulary and operations. Functional partials implement them in the same assembly; no runtime adapter is needed to connect those halves.

| Responsibility | Read these source locations |
|----------------|-----------------------------|
| Immutable bands, policies, transaction and disk contracts | `Common/Memory*.cs`, `Common/Compression.cs` |
| Value/strict JSON construction and configuration for the worker | `Serialization/Memory*.cs`, `Serialization/Compression.cs` |
| Pure context reduction and safe conversation selection | `History/History.cs`, `History/Compression.cs`, `History/MemorySelection.cs` |
| State-only snapshots and single recovery journal | `History/Persistence.cs`, `History/Journal.cs` |
| Fresh corpus reads, capture, fingerprints, staging and recovery | `Memory/` |
| Startup/policy reconciliation and durable operation | `Agent/Memory*.cs` |
| Main-loop integration and complete-work boundaries | `Agent/Actor.cs`, `Agent/Handoff.cs`, `Agent/MemoryRuntime.cs` |
| Provider data roles, cache boundary and output policy | `Engine/Render.cs`, `Engine/MemoryWorker.cs`, `Engine/Streaming.cs`, `Engine/Engine.cs` |
| Exactly-once imported meter rows | `Engine/Meter.cs` and the actor completion path |
| Selected-Agent GUI settings | `Gui/Server/MemorySettings.cs`, `Gui/wwwroot/` |

Read the actual call path, not just the file names. For example:

```text
handoff boundary
  → Agent chooses a MemoryCapture
  → History records CompressionLaunched
  → owned child runs and joins
  → mailbox receives immutable completion
  → Agent imports measured rows once
  → safe main-work boundary revalidates and commits
  → Memory finishes rooted file effects
  → History records CompressionFinished
```

On shutdown the direction reverses through the real ownership relationships: join work, detach owned objects, then close the resources they used. Garbage collection does not perform those logical operations.

One important review question is whether an error occurred before or after the replay commit. Before commit, an unusable candidate can fail while source remains active. After commit, a disk conflict must preserve the receipt and fail for recovery; it cannot be relabeled an abandoned attempt whose source was never removed.

Another is whether a routine is pure. `History.Reduce` must not inspect files. `Memory.Recover` may inspect and finish disk effects, but must not publish History from a worker. The actor connects those responsibilities in a visible order. That separation is the implementation, not an architectural diagram pasted over a background thread that does everything.

---

## §22.9 Failure Lessons Worth Keeping

### A green worker test missed the exact schema

During this C# adaptation, the first worker used a `summary` property with a nonblank pattern. The chapter required a `memory` string with no extra properties. Shared tool rendering also appended observation fields that the restricted compressor did not need.

An exact wire-schema regression exposed the mismatch. The correction changed the public payload and omitted observation fields only for the compressor. It also restored the intended blank-output behavior: finish the bounded attempt with failure, do not ask for another blank-string correction. The internal result and disk receipt can still call their value `Summary`; those names are not the model-facing schema.

This is why “the test passed” is not enough. A test that normalizes away the contract it is meant to check can prove the wrong thing very reliably.

### Harness repairs are not product improvements

The C# executable adapter initially made incorrect assumptions about skill syntax, settings shape, required tool explanations and terminal outcome names. Those failures were retained and corrected as harness failures. In particular, C# records terminal success as `Committed` after promotion; inventing a `Finalized` label made successful work appear unfinished.

Repeated guesses are not evidence of thoroughness. Read the declarations and saved wire format first. Distinguish a compiler/setup failure, a faulty fake and a real product refusal before changing the application.

---

## §22.10 Historical Measurements and Trials

These measurements and live runs belong to the earlier implementation. They explain the design and its limitations; they are not C# acceptance evidence.

### Context Reduction

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

One cascade in ten checkpoints shows the historical selector's behavior, not
summarization quality. The 27-byte and 22-byte bands are fixed fake replies.
Neither their sizes nor the tier names establish real compression ratios.
Moreover, this test combined memory compression with the old Chapter 14 size-cut
ladder; the current policy no longer silently cuts deliberately retained tool
results by size. These are not measurements of that revised policy.

The live trials below test what fixed summaries cannot: whether facts and reasons
survive and influence the answer. They predate the background worker, so they do
not establish its latency or quality.

### What the live runs remembered

These are the October 11 direct-compressor results, before the background-child
change. A new live background run remains pending; the manual exercise below describes
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

---

## §22.11 What the C# Verification Establishes

The preserved historical account above explains why I require the transaction and ownership boundaries. It is not evidence that this C# port passes those boundaries. The C# implementation was exercised separately with real actors, History journals, rooted files, sockets and local provider fakes.

The final broad regression before the manuscript review passed 266 groups. Review then found the missing first-person/attribution instruction in the worker prompt; an added request assertion failed before the prompt correction. Assembled review also found that terminal graduation supplied the existing MEMORY.md twice: the capture formatter already included it, and the worker appended it again. A request-level regression reproduced the duplication before the redundant block was removed. Focused worker checks and actual CLI/GUI builds cover these later prompt changes. Neither the broad result nor the executable grade below is relabeled as a later run. These checks verify request construction, not real-model obedience.

Other local exercises covered:

- Pure memory ordering, population, retirement, snapshots and journal-only replay.
- Selected-source and destination changes, hidden candidates, receipt validation and interrupted promotion/retirement boundaries.
- At that historical verification boundary, a separate canonical archive retained tool/private/opaque events. Chapter 14’s single-journal design supersedes that storage mechanism; those earlier checks do not certify current retention.
- Worker limits, exact submit schema, blank results, JSON/SSE terminal refusal and joined cleanup.
- Concurrent foreground work, safe-boundary commit, policy cancellation, failed saves and no duplicate imported usage.
- Known-window pressure, unknown usage, responsive headroom waiting and cache-boundary placement.
- Agent-local durable settings, selected-Agent WebSocket isolation, actual isolated-browser controls, and a CLI restart reading the same disabled band policy before its first local request.

The opt-in executable adapter preserves the original Go route and awards only five complete existing rows:

| C# executable row | Points |
|------------------|--------|
| Planted-fact preservation | 10 |
| Graduation | 12 |
| Disabled neighbor | 8 |
| Memory is data, not authority | 5 |
| Fresh startup from corpus files | 5 |
| **Exercised scope** | **40** |

The adapter passed 40/40 across four clean joined sessions and 72 local model requests, covering 40 of the original rubric's 100 points. The remaining 60 points cover checkpoint/background, GUI toggle, forced/headroom, replay/accounting, abandon/restart and parity. Those behaviors remain part of the chapter; they are ungraded by this route, rather than failed or removed. The component and GUI checks above provide separate evidence about them without supplying the missing grader points.

The adapter waits for every observed launch's matching committed terminal outcome and then uses an actor barrier before rereading events. It does not infer completion from quiet time. Fresh-start copies only live corpus Markdown, not saves, child submissions, retired files or receipts. A single outgoing request must then contain the planted memory fact.

These checks establish the exercised local contracts. They do not establish real-model semantic quality, live provider output-cap acceptance, cache hit rates, entitlement or billing. Native-platform and human-accessibility limits also remain explicit. The source review still matters: passing fixtures cannot prove that every cross-module call follows the ownership rules.

---

## §22.12 What We Learned

1. **A summary is only one part of memory.** The feature also needs replay semantics, file ownership, stale-result checks, recovery, settings and accounting. A good paragraph written into the wrong lifecycle can still lose information.

2. **Asynchronous code still needs one decision-maker.** The actor owns acceptance and History publication. The child produces a candidate; Memory performs rooted file operations. Adding threads does not require adding another scheduler.

3. **Preserve the source until the replacement is accepted.** A launch is intention, a staged file is preparation, and a matching population is the replay commit. Do not collapse those states into “compression finished.”

4. **Fidelity includes reasons and uncertainty.** Preserve why a design changed, what failed, whose claim a result was, and what remains unverified. First-person prose is useful when grounded in the source, not when it manufactures continuity.

5. **Runtime limits beat prompt requests.** The prompt asks for a handoff and bounded work. The actor enforces tool authority, request counts and the working-context thresholds. Missing usage stays unknown rather than inheriting a convenient number from a test.

6. **The files remain inspectable.** Plain Markdown makes memory useful to people and ordinary search tools. Strict identities and receipts do not turn those files into privileged instructions.

7. **A GUI success message needs a defined boundary.** Save, apply and acknowledge are distinct operations. Restart verification is a further observation, not implied by a moved checkbox.

8. **The codebook must retain corrections as well as outcomes.** The wrong worker field, faulty grade assumptions and rejected file-copy approach teach different lessons. Treating them all as successful iterations would lose the judgment needed to rebuild the system.

---

## Exercises

1. **Rebuild a projection without disk authority.** Populate two memory bands, retire one identity, save and restore. Verify exact text, stable-to-recent order and unchanged grants. Change a file after saving and prove that pure replay still uses recorded bytes until actor reconciliation runs.

2. **Measure a real compression, not just a fake response.** With an approved provider configuration, compress a varied technical session. Compare UTF-8 input/output bytes and measured token usage. Read the result for decisions, reasons, failed approaches, attribution and unfinished work. A reduction ratio alone is not a quality score.

3. **Prove a transaction boundary.** Hold an owned fake compressor after launch. Verify that the old source still appears in a main request. Let it submit, then verify that matching population removes exactly the selected source and installs the replacement. Repeat with failed output and a selected-file edit; neither may lose source.

4. **Recover an interrupted commit.** In an isolated fixture, stop after replay population but before all file retirements. Restart from the retained files. Require matching receipt validation, eventual promotion/retirement, no duplicate population and no imported replay cost. Do not delete a receipt to make recovery “pass.”

5. **Test fresh re-enable.** Disable SOUL and wait for acknowledgment. Edit `SOUL.md`, re-enable it and wait again. Inspect the next actual request for the new bytes as user/data, not system instructions. Then disable it, restart, and verify it stays absent from the first request.

6. **Test real working-context boundaries.** Use provider-correct cached usage and a configurable per-model limit. Establish one gentle 80% reminder, repeated immediate-handoff instructions at 90% (not the gentle wording), and exactly one automatic checkpoint at 100% after companion results are recorded. Test independent Job survival, missing usage, model switching, durable settings, failed saves, and a limit change while paused. Verify the status percentage uses the same actor measurement and that a completed checkpoint cannot retrigger from stale usage. Direct Engine and compressor sends must not replace it.

7. **Compare recall across the waterfall.** Plant a specific but ordinary fact in session files, graduate through both archive bands into MEMORY, then start with only live corpus Markdown. Check the first request as well as the eventual answer. A model can guess a plausible answer without having received the memory.

8. **Review the ownership path.** Trace shutdown during an active compressor. Identify which parent closes admission, which task is joined, when usage is imported, how the child detaches, and when Sandbox closes. Do not replace this exercise with a count of passing tests.

---

## Take It for a Spin

Begin with the offline tests. From the repository root, use an already installed .NET 10 SDK:

```bash
cd agent

dotnet build Tests/Tests.csproj
dotnet run --project Tests/Tests.csproj --no-build -- --memory-band-tests
dotnet run --project Tests/Tests.csproj --no-build -- --memory-worker-tests
dotnet run --project Tests/Tests.csproj --no-build -- --memory-runtime-tests
dotnet run --project Tests/Tests.csproj --no-build -- --memory-accounting-tests
dotnet run --project Tests/Tests.csproj --no-build -- --memory-pressure-tests
dotnet run --project Tests/Tests.csproj --no-build -- --memory-settings-tests
```

These use owned local boundaries. They make no claim about a live provider's compression quality or caching. Other selectors in `Tests/Program.cs` cover disk, replay and startup; read their scope rather than assuming every selector runs the whole suite.

For a manual GUI exercise, keep a separate home and port. Do not point the exercise at a working Agent's directory. The example below seeds one unmistakable SOUL value before the first actor scan:

```bash
# Continue from agent.
export EN_DATA_DIR="$(mktemp -d "${TMPDIR:-/tmp}/ensemble-memory.XXXXXX")"
export EN_AGENT_NAME=root
mkdir -p "$EN_DATA_DIR/agents/root"
printf '%s\n' 'Workshop marker: cobalt heron 731.' \
  > "$EN_DATA_DIR/agents/root/SOUL.md"

export EN_SKILLS_DIR="$EN_DATA_DIR/skills"
export EN_PRIMARY_SKILL=memory-exercise
mkdir -p "$EN_SKILLS_DIR/memory-exercise"
cat > "$EN_SKILLS_DIR/memory-exercise/SKILL.md" <<'SKILL'
---
name: memory-exercise
type: primary
description: Analyze supplied cases and retain useful checkpoint memories
tools: micro_handoff read_file write_file list_directory
---
Analyze only the supplied task. Preserve facts, decisions and reasons, failed
approaches, constraints and open questions in checkpoints. Explain before tools.
Enabled tools:
$TOOLS
SKILL

dotnet build Examples/Gui/Gui.csproj
# Choose an unused loopback port; this example does not use the working host's 8088.
dotnet run --project Examples/Gui/Gui.csproj --no-build -- --port 8097
```

Before launching, explicitly choose your intended model, route, funding and existing authorized credential source using the host options from Chapters 16 and 20. Do not paste tokens into the book, logs or conversation. Starting a local page does not establish that inference is configured. The chapter's verification used local fake providers; a real-provider exercise spends real usage and remains a separate test.

Open the printed loopback URL in a separate browser profile. In Memory settings, reload the current acknowledged values. Disable SOUL and save. Inspect the Agent-local `settings.json` and wait for the application acknowledgment, not just a changed checkbox. Edit the marker while disabled, re-enable and save, then send a small prompt. Inspect the actual request through an explicitly enabled, protected diagnostic capture; it should contain the edited marker in data, not privileged instructions. Such captures can contain conversation content—keep them private and disable them when done.

For a small waterfall exercise, reduce conversation and session thresholds in that isolated Agent, keep the session destination enabled, and use a primary skill that grants `micro_handoff`. Work through several complete turns, then ask for a handoff. A first handoff may have nothing eligible; small inputs may never exceed the threshold. Watch the worker child and wait for its terminal outcome rather than expecting a file after a fixed sleep.

The historical trials used these two bounded cases. Send each separately and let the main turn finish; do not mistake that for the child's completion. They test rejected approaches and unresolved questions, not just fact strings:

```text
This is a bounded memory exercise in an empty scratch workspace. Do not use filesystem or command tools. Analyze the supplied case, explain the decision briefly, then call micro_handoff exactly once with the important facts and reasons. After the tool returns, finish this turn without another tool call. Case one: Our project is Cedar Lantern, release code CL-721. We rejected automatic deletion of imported notes because the support team needs the originals when customers report mistakes. The replacement design keeps source notes in an archive and stores the accepted summary separately. Morgan owns the archive review, scheduled for October 18. Acceptance requires preserving a source identifier with each summary. The unresolved question is whether attachments belong in the same archive or a separate one; do not claim that has been decided. The default retention is thirty days. A failed summarization must leave the original note available. The initial approach coupled cleanup to request arrival, so an unrelated request could delete unreviewed work. We moved cleanup to an explicit completed-review boundary instead. Explain why the latter boundary matters, and preserve these facts in your checkpoint.
```

```text
This is a bounded memory exercise in an empty scratch workspace. Do not use filesystem or command tools. Analyze the supplied case, explain the decision briefly, then call micro_handoff exactly once with the important facts and reasons. After the tool returns, finish this turn without another tool call. Case two: The deployment rehearsal uses Harbor Finch, diagnostic tag HF-934. Priya owns the rehearsal scheduled for October 21. We rejected retrying a database migration automatically after a timeout: the remote operation may already have committed. The chosen procedure reads the migration ledger before deciding whether any retry is safe. The operator must retain the exact migration identifier and measured error, not a guess about success. We have not settled whether the ledger should live alongside application data or in a separate control database. The rollback package must remain available for seven days after deployment. Last week a log line said the worker stopped, but the child process was still alive and held the port open. The rehearsal now checks the child lifetime directly. Explain how this evidence changes the decision and checkpoint the completed analysis, preserving the unresolved question as unresolved.
```

After the relevant committed outcomes and any graduation, ask without tools for both names, codes, owners, dates, rejected approaches and reasons, explicitly preserving the two unresolved questions. Compare the answer with the outgoing request and corpus. If a fact was present but the answer omitted it, that differs from losing it during compression. A checkpoint while a child is outstanding does not start a second child; after completion, checkpoint remaining work if needed.

While a child is visibly running below the context limit, try a short unrelated question. A reply before the memory outcome demonstrates overlap. A child that finishes too quickly leaves responsiveness unobserved; do not manufacture a timing conclusion.

Inspect without changing the state:

```bash
find "$EN_DATA_DIR/agents/root/memory" \
     "$EN_DATA_DIR/agents/root/bucket-0" \
     "$EN_DATA_DIR/agents/root/bucket-1" -type f -print
# Some directories do not exist until first use; that is not a failed scan.
```

Read the live Markdown and the parent and worker recovery journals (`save.json.journal.jsonl` and `children/compress-*/save.json.journal.jsonl`). Distinguish a launch, matching population, successful terminal outcome and retained failed attempt. `.retired` source files are not active memories. Audit-off journals may discard older records after checkpointing; enable audit at launch if this exercise needs a complete original trace. Do not mistake retained summaries for that trace.

Finally, stop this owned host cleanly and restart it with the same explicit `EN_DATA_DIR` and launch configuration. Confirm the acknowledged settings and first request, not just the continued existence of Markdown files. To test corpus-only recovery separately, copy only live SOUL/MEMORY/session/archive Markdown into another fresh home; do not copy saves, receipts, retired files or child submissions.

The point is not to accumulate files. It is to show that a useful account survives a smaller working context, remains inspectable, and never makes an unfinished or failed operation look complete.

---

**Previous:** [Chapter 21: Workflows](chapter-21.md)
**Next:** [Chapter 23: Auto-Recall](chapter-23.md)
