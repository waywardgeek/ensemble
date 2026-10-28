# Chapter 14 — The world's best context engineering, before breakfast

A long session with a coding agent can degrade in a familiar way. The answers get vaguer. The agent reads a file it read an hour ago, then forgets a decision made before lunch. A fresh chat can help because a person decides what to carry over. This chapter moves tool-traffic cleanup inside the Agent and records every cut, so replay can explain what left the working context and reproduce the result.

The title comes from the author's account of the original design: after carrying it around for months, he described it to CodeRhapsody one morning, and the first commit landed at 7:42, before breakfast. That is the book's historical account, not a claim about the time taken by this C# implementation or a benchmark against every other system.

The important idea is less grand and more useful: **the context window is a working set, not the archive**. Keep the decisions and their reasons. Remove tool bytes deliberately. Do not let a cleanup erase the manual that explains the tools still available.

## TL;DR: what you build

1. **One owner for each decision.** Agent selects reductions at safe actor boundaries; History applies recorded facts; Engine renders the resulting Context. A worker never trims History or decides when another model request may start.
2. **One target, two bands.** The default target is 400,000 bytes. Results use T/8, calls and stubs T/16. Each band crosses twice its budget before cutting an old prefix back toward one budget. Complete call/result batches are the unit of selection.
3. **An explicit keep-or-stub option.** For opted-in configurations, a large result is considered once after a successful request presents it. The model may call `keep_tool_results` in the consuming response. Ordinary models do not acquire this policy from a guessed name.
4. **A real checkpoint kind.** `micro_handoff` records an ordinary result and then `HandoffRecorded`. The reducer removes tool traffic and private replay material, replaces older active handoff notes, and preserves visible dialogue and manuals.
5. **Immediate revocation, retained instructions.** Pending skill unload removes authority, including deferred references and automatic snapshots. Its manual remains in SYSTEM. Retiring that manual is not this chapter's job.
6. **Recoverable facts.** Serialize writers, append journal records before publication, validate recovery candidates independently, and retain journal events beyond the validated backup—not merely beyond the current primary.
7. **A full event archive.** An optional archive contains original facts, including tool payloads and cuts. It is not reconstructed dialogue and does not silently invent missing history.
8. **Honest verification.** Component, separate-consumer, CLI, GUI, browser and executable-grader checks establish different things. Passing a loopback provider does not establish a real provider, audible speech or human accessibility.

The endpoint is still C# 14/.NET 10. Shared vocabulary and persistent state are declared in `Common`; functional partial implementations supply the bodies. The GUI remains a separate application, not another scheduler or transcript owner.

## In plain words

Keep the user's words and useful explanations. Once a tool result has done its
job, replace its bulk with a stub, and later remove the paired call and stub.
At a deliberate checkpoint, keep one new note and clear tool traffic without
summarizing away the conversation. Record every cut so replay knows exactly what
happened. Preserve original facts separately, and make crash recovery depend on
validated files rather than guesses about what was probably saved.

## 14.1 Why is this byte here?

Ask three questions of every byte sent again with a request:

- Why is it here?
- What put it here?
- What can remove it?

A user prompt has a reason to remain. A useful explanation of an edit does too. A large file result often answered a question once and is now occupying the window without another decision to support. The first edition's measurements reported about 42 percent result bytes and another 30 percent call arguments in the measured histories. Those figures motivate examining tool traffic; they are not measurements of this C# endpoint.

A removed file result can often be obtained again by reading the file. That is not the same as recovering the original result: the file may have changed, and repeating a command may have effects. Preserve original facts in the archive when their historical identity matters. A stub should say that bytes were removed, not fabricate a trustworthy locator.

The book records a particularly instructive earlier bug: a stub referred to an output path whose numeric handle was reused after restart. The next reader received plausible output from the wrong command. Wrong data was worse than no data. This endpoint does not turn a result stub into an unverified promise that some path still names its original bytes.

Tool clearing is not dialogue summarization. The automatic policy in this chapter never asks a model to replace the user's goal or a discussion with a compressed paraphrase. Existing reduction primitives can express other operations, but the policy here selects result and tool reductions only.

## 14.2 Keep the relationships simple

The real ownership paths remain:

```text
Ensemble → Agent → History
                 → Engine
                 → Jobs → Job
                 → Skills / MCP owners, when configured
```

History owns committed Context, retained diagnostic events, sequence numbering and persistence coordination. Agent owns the target, the actor's temporary decisions and request admission. Immutable Context, entries and parts are value-like snapshots; they do not acquire holder back-pointers simply because several readers use them.

Common declares the new operations and values, including:

```csharp
public const int DefaultContextByteBudget = 400_000;
public partial int ContextByteBudget();
public partial void SetContextByteBudget(int bytes);
public partial (int ResultsBytes, int CallsBytes) ToolCompactionBudgets();
```

These declarations belong to the existing partial Agent. Their implementation belongs in `Agent/Compaction.cs`. There is no settings service copied into every object and no separate reduction scheduler.

Constructors establish the existing parent relationships. Explicit destruction joins owned work and removes both ends. History remains attached while terminal facts and shutdown persistence need it. A recording failure does not permit cleanup to abandon an Agent in its parent's collection; cleanup must still complete, while preserving the original failure.

## 14.3 Instructions are not disposable tool results

Chapter 9 already put loaded manuals in provider instruction context. Do not move them into dialogue merely because a historical implementation did so. Primary and loaded `SkillPart` bodies appear once in the provider's instruction field; ordinary transcript messages do not duplicate them.

A skill's recorded grants describe what was activated. Its pending marker changes what is authorized now. Derived Context queries compute active and revoked names from the immutable request snapshot:

- A pending-only grant is unavailable immediately.
- Another active skill may still grant the same name.
- Dispatch refuses an unavailable tool even if a caller guesses its name.
- Ordinary declarations, deferred catalogs and emitted discovery references follow the same rule.
- Historical discovery metadata stays in canonical events; request projection filters it without rewriting the record.

Automatic MCP operations remain outside callable model tools. A pending skill cannot continue collecting automatic snapshots. Unloading a server-backed skill also closes its owned connection/tools under the existing ownership rules.

Both tool clearing and handoff retain pending manuals in SYSTEM. **Removing permission and retiring an instruction body are different operations.** This chapter implements the former, not a new manual-retirement scheme.

The historical chapter discussed frozen inline declarations. This C# edition uses the Chapter 9 provider declaration mechanism instead. Stable primary instructions and infrequent cuts remain useful, but this endpoint makes no claim to implement that old inline protocol or to have measured a provider's cache benefit.

## 14.4 The two-band ladder

Let T be the Agent's context-byte target. The default is 400,000:

| Use | Derived amount at the default |
| --- | ---: |
| Whole-result band, T/8 | 50,000 bytes |
| Call-and-stub band, T/16 | 25,000 bytes |
| Optional immediate-stub threshold, T/100 | 4,000 bytes |

A positive target has a one-byte minimum for each ladder band. The public setter refuses negative values without mutation; zero disables automatic byte-budget reduction. These are byte measures of serialized tool material, not token counts, model context capacities or a hard cap on the entire conversation. Dialogue and manuals may grow beyond T.

Each immutable entry caches its tool-byte measure. This is a derived value of immutable content, not another mutable accounting authority. Selection examines complete batches, using response-local call/result pairing. Repeated IDs in later completed batches remain legal. An incomplete, orphaned or ambiguous exchange stops safe prefix selection; the algorithm does not invent a result to make progress.

The result band measures eligible batches with whole results. Above twice its budget, select the oldest prefix sufficient to bring the measured band down toward one budget. Record a `Redacted` event with `Redaction.Result` and fixed sequence bounds. Applying it replaces result bodies with stubs while preserving their calls and identities. Re-stubbing an existing stub preserves its metadata.

Then consider the call/stub band. It may remove only a stub-only prefix, never jump over a still-whole batch. Above twice that band's budget, record `Redaction.Tool` over complete old batches. Calls and results disappear together; surrounding visible text remains.

Why twice the budget? A cut invalidates a region of the request. Cutting larger chunks less often avoids constantly shaving its front. This is hysteresis, not a guarantee about cache hits on an untested provider.

The actor advances its presentation watermark only after a successful response. A new result therefore reaches a successful request before automatic clearing can select it. After restoration, the first request conservatively presents retained results again. Recorded cuts, however, are never reconsidered: raising T later cannot resurrect them.

## 14.5 Keep or stub, once

Some sessions benefit from more aggressive treatment of a large result after its first use. `Configuration(stubsToolResults: true)` enables that explicit policy. The hosts expose `--stub-tool-results`, with `EN_STUB_TOOL_RESULTS=1` as their environment alternative. The default is false; no model-name prefix silently enables it.

The concrete `keep_tool_results` control is installed and granted for this option. It takes no policy arguments. A successful model-dispatched call vetoes the immediate cuts staged for that consuming response. A same-named custom tool is not the control. Invalid arguments, a failed call or an undispatched interrupted call do not count as keeping.

Selection occurs against the Context that was presented, before recording the consuming response. Results larger than T/100 are candidates. The actor holds only temporary proposed cuts while that response's tool batch completes. Successful keep discards those proposals; otherwise the actor records the cuts after all companion outcomes are durable.

Every completed assistant response has a canonical entry, even an empty response. That entry closes the prior batch's once-only decision boundary. A later text-only turn or snapshot-only restart must not reconsider a batch that was kept. This uses the existing conversation structure rather than a persistent batch registry.

Keeping is not unlimited retention. The ordinary ladder may still stub or remove an oversized kept batch when its band requires it. Small results, ordinary-model sessions and a zero target retain their defined behavior.

## 14.6 Explain the work so useful reasons survive

Once a file result disappears, what remains about the edit? Whatever useful explanation the Agent recorded, plus any facts still available in their original sources.

Compare:

```text
read_file config.cs
edit_file config.cs
```

with:

```text
The port appears to be fixed in config.cs; I'll check it before changing settings.
read_file config.cs
The listener uses a literal port here. I'll move that choice to the existing settings owner.
edit_file config.cs
```

The second version preserves intent and a reason. It does not need private chain-of-thought or a verbose narration of every keystroke. A concise explanation before an action lets a person supervise it now and helps a future continuation understand it later.

Narration is evidence of what the Agent said, not proof the statement remains true. Record paths, relevant operations, observed outcomes and uncertainties. Re-read a changing source instead of treating an old sentence as a current fact. This is why the chapter keeps words without pretending that words are infallible memory.

## 14.7 A handoff is a fact, not a destructive summary

The book's earlier experiments tried positional `compress_context`: a model chose messages to replace with a summary. Bill's September 2026 account described repeated loss of most of the conversation, including the goal near its beginning. Later whole-task handoffs started a fresh instance from a document, introducing a second path for durable state. Those are historical lessons, not implementations to reproduce here.

`micro_handoff` is smaller and more constrained. The model supplies a nonempty `text` note through the normal tool protocol. Admission validates it; the ordinary successful result is recorded before the note can take effect. If the response also contains another tool, its outcome must be recorded before the cut. Interruption pairs undispatched tails with their exact-ID refusals first. Independent Jobs are not cancelled merely because their observations no longer belong in context.

The actor then records `HandoffRecorded` carrying a value-like `HandoffPart`. History applies that fact by:

1. Removing tool calls and results from working Context.
2. Removing opaque reasoning and private replay signatures, including signatures attached to otherwise visible text.
3. Retaining the visible text, ordinary dialogue, system entries and skill manuals.
4. Removing older active handoff parts and appending the newest note.

Do not recognize a handoff by a magic prose prefix. The marked part makes replacement precise. Engine converts it to ordinary system text only in the disposable provider projection. Canonical Context and snapshots retain its identity. Original notes and private parts remain in the event archive; clearing a working view does not rewrite history.

The note should say what succeeded, what failed, where the source of truth lives, and what comes next. Failed approaches are often the highest-value bytes because they prevent repetition. Record uncertainty too. A note is a continuation aid, not proof of subjective continuity or a substitute for checking current state.

## 14.8 Manual clearing follows the same actor boundary

The public control is `AgentInput(InputKind.ClearTools)`. The CLI accepts `{"kind":"clear_tools"}`; the GUI provides **Clear tool traffic**. Neither creates a new worker that edits History.

The actor stages one pending flag. After current batch outcomes and any staged result cuts/handoff are recorded, it records a tool-only redaction over the actual committed sequence range. Pause blocks a new model request, not this completed-batch control. An idle paused Agent can therefore clear immediately. A repeated clear with no tool traffic creates no synthetic cut.

For a caller that needs to observe completion:

```csharp
await agent.PostAsync(new AgentInput(InputKind.ClearTools));
await agent.FlushAsync(cancellationToken);
```

`PostAsync` admits the control. `FlushAsync` observes accepted foreground work; it does not resume a paused turn or wait for all independent Jobs. Manual clear preserves dialogue, pending manuals and the newest handoff. It does not terminate a command simply because that command's earlier observation is cleared.

## 14.9 Total reduction is not permissive decoding

A recorded cut contains its decision: event kind, level and exact sequence range. History applies that event. It does not ask today's Agent policy to select the range again.

Known but inapplicable facts must not stop all later replay. An inverted or out-of-range reduction, or a known event missing the payload needed to apply it, is a safe no-op on Context. History reports a safe diagnostic with sequence, kind and reason, and later valid facts still apply. Sequence advancement includes the skipped durable fact.

That is different from accepting corrupt data. Unknown event kinds, unsupported versions, invalid supplied values, malformed tool identities and broken journal sequences remain errors. Do not implement totality as a blanket catch that hides invariant faults. Separate strict decoding from explicit applicability checks; keep public `Reduce` pure and put diagnostics on the real History owner.

Snapshot/tail loading stages a complete candidate before publication. A bad candidate cannot leave half a restored conversation live. Replay and incremental append use the same reduction behavior, so a full log and its snapshot produce the same request under the same current runtime configuration.

## 14.10 The recovery journal, primary and backup

Chapter 10 saved on joined shutdown. A killed process cannot perform that shutdown. Chapter 14 adds a journal to the same History owner, not a second event store or scheduler.

`Agent.StartPersistence` restores first, activates `<save-path>.journal.jsonl`, then records current runtime configuration. It runs before accepting input. Direct `History.Save` remains a snapshot operation; it does not silently enable continuous journaling. Explicitly disabled persistence remains available.

All writers serialize through History's existing save gate. Append computes its next event and reduced state under a short state lock, writes and flushes the journal and optional diagnostic writer outside that state lock, then publishes committed events/Context/sequence. Readers can still capture the last committed snapshot during slow disk work. Another writer cannot race past it.

If append fails, publish nothing live and poison further append/save attempts. A partially written sequence must not be reused under the pretense that nothing happened. Shutdown may then report the original recording fault, but explicit destruction still joins work and detaches owners.

Recovery validates candidates independently. Prefer a valid primary plus its contiguous journal suffix; otherwise try a valid backup plus its suffix. With neither snapshot present, a journal beginning at sequence one can recover the initial session. Do not turn corrupt candidates into an empty conversation or use a valid primary to conceal an invalid backup-retention assumption.

Only complete newline-terminated journal records count. An incomplete final suffix, even a partial UTF-8 character, may be ignored with a diagnostic. Malformed complete records, duplicate numbers and gaps refuse. Loading does not repair files. Activating a writer after successful recovery rewrites the validated complete prefix before new appends.

Saving follows a careful order:

1. Capture a consistent snapshot without holding the state lock across disk I/O.
2. Render and validate the candidate; prepare any requested archive.
3. Validate the old primary before rotating it into backup. Never overwrite a good backup with corrupt primary bytes.
4. Replace the primary with the new validated snapshot.
5. Retain journal records newer than the **validated backup**. Without a valid backup, retain them all.
6. Only then trim the live diagnostic tail as configured.

The journal is not trimmed merely to the latest primary boundary: that would destroy the facts a backup needs after primary corruption. No `fsync` or hardware power-loss guarantee is claimed here. The executable crash exercise tests process kill and recovery, including a torn final suffix.

## 14.11 A full archive must really be full

An optional `ArchivePath` is an event-complete JSON-lines log. A dialogue-only reconstruction cannot replay cuts, tool outcomes, skill grants or lifecycle facts, so it is not an adequate substitute.

Before changing snapshots or trimming, merge the existing archive, available journal and retained diagnostic events. Require a contiguous sequence from one through the captured boundary. Overlapping sequence identities must serialize identically; missing earlier facts, conflicting overlaps or future events refuse before snapshot/backup changes. Strict UTF-8 prevents silently changing original bytes through replacement characters.

If no source still contains an early event, enabling an archive later cannot manufacture it from reduced Context. Report the missing history. Keep archive, save, backup, journal and open diagnostic paths distinct. Writing an archive successfully before a later snapshot failure is acceptable; a retry must validate and reuse it safely.

The public `History.Load` log path can replay this archive. It recovers the reduced working view because the original payloads and the subsequent cuts are both there. Diagnostic retention zero therefore need not mean lost history—provided journaling/archive saving actually preserved it. A snapshot alone remains a different artifact.

## 14.12 One setting, one authority

The library Agent owns the live byte target. The GUI's settings snapshot reads it rather than keeping a second mutable copy in presentation preferences. Validate a proposed patch and save it before publishing either preferences or policy. A failed save must leave the old file, live policy and both clients' acknowledged values unchanged.

The GUI field is `context_target_bytes`. Its convention is intentionally different from the library setter: zero or a negative value selects the 400,000-byte default, positive values round to thousands, and a rounded zero also selects the default. Large values are bounded safely. A wrong-typed optional field is ignored like the existing preference fields. The control shows acknowledged values, retains a proposed edit while saving, and restores acknowledgement on failure.

The CLI takes `--context-budget-bytes N` or `EN_CONTEXT_BUDGET_BYTES`; an explicit flag wins. Here zero disables automatic reduction, and negative values refuse. Document this difference instead of hiding it behind a shared parser with contradictory semantics.

Live reductions also change already-rendered cards. Incremental frames newer than a cursor cannot remove an older tool card. The GUI therefore uses its existing silent replay envelope when a recorded cut crosses the peer's view. Snapshot Context supplies the transcript; diagnostics cannot resurrect a cleared model dispatch card. The completed-response frontier retires both stored and late streamed partials at `ResponseEnded`, not only at the end of the entire turn.

A replay is presentation recovery, not another speech event. The newest handoff is readable; refresh and a zero-log restart must project the same reduced transcript without automatically sending a model request.

## 14.13 Build and exercise the endpoint

From a preserved source endpoint, use an approved .NET 10 SDK:

```sh
dotnet build Ensemble.csproj
dotnet run --project Tests/Tests.csproj
dotnet build Cli/Ch04.csproj
dotnet build Examples/Gui/Gui.csproj
```

The CLI project name remains `Ch04.csproj`; renaming it is not part of context engineering. Keep credentials, logs, profiles, captures and build products outside source endpoints. Use a fresh owned workspace and local ports, not another person's running application.

An ordinary interactive CLI launch can opt in explicitly:

```sh
dotnet Cli/bin/Debug/net10.0/ch04.dll \
  --context-budget-bytes 400000 --stub-tool-results \
  --save /owned/session/save.json --save-log 0 \
  --archive-log /owned/session/archive.jsonl
```

Supply provider configuration through the already documented host boundary; do not put credentials in conversation or the archive. For offline replay:

```sh
dotnet Cli/bin/Debug/net10.0/ch04.dll dump /owned/session/archive.jsonl
dotnet Cli/bin/Debug/net10.0/ch04.dll render /owned/session/archive.jsonl
```

These commands inspect recorded facts and render a request; they do not establish provider success. CLI JSON-line user input queues work. If an exercise needs a clear after a particular turn, wait for that turn's durable completion rather than sending several lines and assuming each blocks.

The executable-scoped grader is invoked from the book repository with `-ch 14 -executable-only -gui EXECUTABLE -json`. Its executable argument is a launcher for the built C# GUI, not a static-assets directory. The launcher forwards flags; the grader supplies owned paths, dummy credentials and loopback peers. The default Go grading route remains separate.

Its eight checks cover instruction survival, handoff pairing and sibling order, keep/stub behavior, recorded cuts across a changed target, current declarations and actual MCP unload refusal, snapshot versus full-log replay, anchored SIGKILL/backup recovery, and total reduction. Current startup skills are explicit in restart fixtures; saved configuration is not secretly reactivated. Historical inline-tool placement is not credited to this C# path.

Then use the application, not just the report. Observe a full result before a later stub, keep a result, clear while paused, issue two handoffs, refresh, and restart with zero diagnostic retention. Change settings from a second client and deliberately fail an owned settings write. Repeat the cut projection with SSE so a disappearing final cannot leave an old partial tool card. Speech wiring and visual/browser assertions are still not proof of native hearing or accessibility.

## 14.14 What this endpoint establishes

The final component suite has 89 passing groups. A separate assembly consumer uses public APIs for policy updates, real host-control handoffs, paused clear, full archive replay, zero-log restoration and complete owner detachment. Its seeded historical tool exchanges are replay fixtures, not claims of executed effects. CLI and plain/SSE browser exercises separately use real actor, Job, persistence and host paths with working loopback boundary fakes.

The scoped executable grade passes all eight checks for 100/100. The first attempt scored 70/100 because its adapter assumed restored manuals and a Go event label; correcting the fixture made current startup selection explicit and checked the actual C# encoding. Those earlier failures remain failures, not retroactive successes. The progress record retains the detailed receipts and process exits.

No live-provider result is credited at this boundary. Live testing is authorized, but the test session has no approved credential source available. Native/audible speech and human-accessibility verification also remain incomplete. These limitations do not erase the offline evidence, and the offline evidence does not erase them.

The lesson is not that smaller context proves better memory. It is that every retained byte has a reason, every cut is an inspectable fact, and recovery can reproduce the working state without guessing what the Agent must have meant.
