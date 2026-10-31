# Chapter 10 — Persistence and Replay

## TL;DR

Give an Agent a durable conversation without giving yesterday’s configuration authority over today’s runtime. Save one reduced Context, its `as_of` event sequence, a credential-free description of effective configuration, and an optional diagnostic log. Validate a complete candidate before replacing live state or a save file. On restart, load the snapshot, fold only the newer tail, and number new events above the recovered boundary.

The C# endpoint is `solutions/ch10`. It uses the existing Agent → History relationship. There is no second transcript, persistence service bag, or replacement actor loop.

## 10.1 In plain language

Think of a snapshot as a notebook whose contents have already incorporated events through page number N. Keeping copies of those old events beside the notebook is useful for diagnosis, but reading them into the notebook again would duplicate the conversation.

An event newer than N is different: it has not yet been incorporated and must be applied once. If the snapshot is absent, start with an empty notebook and replay the complete event log instead. If the log is absent, the snapshot must still stand on its own.

The notebook records what happened. It does not choose which provider, credentials, model, skills, or tools you are allowed to use now. Those come from the current application’s explicit runtime configuration.

A backup is another candidate, not permission to ignore validation. A corrupt primary must never displace a known-good backup. If neither candidate is usable, stop without rewriting either one.

## 10.2 Ownership and the save contract

Common declares the shared vocabulary; functional modules provide the bodies:

- `Common/Persistence.cs` declares immutable `PersistenceOptions` and `SaveDocument`, plus the operations on Agent, History, Engine, and JsonCodec.
- `History/Persistence.cs` owns capture, replay, candidate selection, replacement, and diagnostic retention.
- `Serialization/Persistence.cs` owns the neutral JSON format and validation.
- `Engine/Persistence.cs` projects effective configuration without credentials or endpoints.
- `Agent/Persistence.cs` performs startup loading and arms automatic saving. The existing actor’s joined shutdown performs the save.
- CLI and GUI composition roots choose paths and policies. Core code does not read environment variables or standard input.

`PersistenceOptions` is value-like. It contains paths and a retention count, not a service, owner, or mutable conversation. It needs no parent back-pointer.

A save has exactly these four top-level fields:

```json
{
  "context": {},
  "as_of": 42,
  "config": {},
  "log": []
}
```

The empty objects above are schematic, not valid complete Context or configuration values.

`context` contains dialogue entries, the current ephemera slot, reducer phase, and measured usage. Each entry carries its sequence, actor, parts, and any finalized display identities. Text, tool calls, tool results, exact JSON numbers, blob locators, opaque provenance, signatures, redaction stubs, and skill instruction values use the existing neutral part codec. Saving does not resolve media files, fetch URIs, or convert the conversation into provider wire JSON.

`as_of` is the event boundary already incorporated into Context. History owns this fact explicitly. It cannot infer it from the last retained diagnostic event: a valid snapshot-only save has no such event.

`config` documents the effective vendor, model, system prompt, and normalized tool declarations. Engine derives it through the existing request projection, using only the captured skill instruction entries. Keys, authorization headers, provider endpoints, and unrelated runtime settings are not configuration fields in the save. Conversation text is still private user data; excluding credentials from configuration does not make a session file public.

`log` contains canonical events retained for diagnosis and, when needed, newer tail replay. It is not a second mutable conversation authority.

## 10.3 Capture and replay

`History.Snapshot()` obtains Events, Context, and Sequence under the same short lock. All three describe one committed instant. Readers never copy an unsynchronized list, and a save does not hold that lock across filesystem I/O.

The loader validates a candidate into local values before publishing anything:

1. Decode the document and required fields.
2. Validate Context entries, usage, phases, display identities, and event shapes.
3. Require strictly increasing positive event sequences.
4. If Context exists, start with it and skip events at or below `as_of`.
5. Fold each newer event in sequence. A gap in the required tail is an error, not an invitation to manufacture missing state.
6. If Context is `null`, fold the complete log from empty Context instead. The complete log must start at event 1.
7. Set the new numbering boundary to the larger of the saved `as_of` and the replayed sequence.
8. Publish the recovered Context, retained events, and sequence together.

A failed load leaves the prior live state and file bytes unchanged. Loading is a startup operation, before opening an output event log or accepting actor work. `Agent.StartPersistence` enforces that startup boundary. The lower-level History operations are also available for explicit import/testing by a caller that owns the startup sequence.

There are three useful equivalent representations:

| Representation | Context | Log |
|---|---|---|
| Snapshot plus diagnostics | Complete through N | May include events ≤ N |
| Snapshot plus tail | Complete through N | Includes required events > N |
| Log-only | `null` | Complete from event 1 |

A snapshot-only save is the first row with an empty log. It must restore both conversation and future numbering correctly.

## 10.4 Safe replacement and recovery

History serializes save operations with a save lock distinct from its short state lock. An append can continue while a captured candidate is being written. A later append is not part of that candidate and must not be removed during its diagnostic trimming.

Saving follows this order:

1. Normalize and validate the save/archive paths. The archive cannot be the primary or its `.bak` file.
2. Capture one coherent snapshot and derive its credential-free configuration.
3. Build the selected diagnostic suffix and validate the serialized candidate by decoding and replaying it.
4. If enabled, write the dialogue archive from that captured Context.
5. Read and validate the previous primary before rotating it into `save.json.bak`. Never promote corrupt primary bytes over a good backup.
6. Write the new primary through a same-directory temporary file, flush its contents, and replace the destination.
7. Only after successful publication, trim the captured diagnostic prefix in live History. Preserve every event appended after the capture.

Temporary files are cleaned up on failure. Same-directory replacement avoids publishing a partially written JSON document. This is not a claim that every filesystem or power failure preserves all metadata; the chapter does not implement a write-ahead journal or directory-fsync recovery protocol.

Startup tries the primary and backup independently. Missing files allow a new session only when neither candidate exists. Parse, replay, access, or shape failures are reported safely; corrupt file contents are not copied into diagnostics. Successful backup recovery reports its saved sequence. It does not claim to recover events that were never durably saved.

If both candidates fail, construction/startup is refused and saving is not armed. Do not “repair” an unreadable save by silently overwriting it with an empty conversation.

## 10.5 Reapply the current runtime

Saved configuration is documentary. The application still supplies the current Configuration, skill catalog, primary skill, variable values, and tool handlers.

`Agent.StartPersistence` captures the freshly constructed runtime’s skill instruction values before loading. After a successful restore, it records `RuntimeConfigured`:

- remove saved skill instruction/grant entries;
- append the current runtime’s initial skill values;
- reset the reducer phase to Idle;
- preserve ordinary historical system entries and all other reduced conversation.

Skill source and pending-unload queries use the reduced skill entries. A pending
flag is part of the immutable skill value, so trimming diagnostic events cannot
erase it during the current session. Runtime reconfiguration replaces those old
values rather than reviving their flags. Yesterday’s dynamically loaded skill does not become today’s permission merely because its tool result remains in the conversation. The new primary is rendered from today’s inputs, then remains frozen for this Agent as Chapter 9 requires.

`RuntimeConfigured` is a durable explanation, not an out-of-band mutation. Log-only replay reaches the same runtime state. The GUI also treats it as a new live-session boundary, clearing old pause/shutdown projection flags without deleting the transcript.

Hints retain their recorded actor. In particular, a human hint is not a skill instruction or a system message. Anthropic’s broad user projection can conceal a role error that OpenAI exposes, so cross-vendor restoration tests must check actual roles, not only whole-request text.

## 10.6 Retention and the optional archive

`RetainLog` has three meanings:

- negative: retain all available diagnostic events;
- zero: retain none;
- positive N: retain the newest N diagnostic events.

A newly opened diagnostic event log writes the retained/bootstrap prefix, not
a fabricated complete event history. After retention, use the save document—not
a standalone truncated event log—to recover the full conversation.

This controls diagnostics, not conversation. Context, usage, redaction state, and `as_of` are always saved. Retention applies to both the saved diagnostic array and the corresponding captured prefix in memory. It never renumbers events or recreates previously discarded diagnostics.

An optional `ArchivePath` receives a dialogue-only JSONL projection from Context before trimming. It excludes runtime skill instruction entries and is not a recovery authority. It is neither the primary event log nor Chapter 14’s full archive/journal design. Rewriting that projection preserves the current reduced conversation, including summaries and redactions; it does not promise to retain removed raw material forever.

The archive, primary, and backup are not one multi-file atomic transaction. If a later primary write fails, the archive may already be newer. Live diagnostics are still untrimmed, and primary recovery follows the validated primary/backup contract. Do not load the derivative archive as a substitute save.

## 10.7 Public C# usage

The library defaults to no automatic filesystem persistence until its caller configures a policy. This prevents unrelated embedded Agents from accidentally sharing the process working directory. The CLI and GUI opt in by default.

```csharp
var root = new Ensemble.Ensemble(Console.Error);
try
{
    var configuration = new Configuration(
        Vendor.Anthropic,
        model,
        baseUrl,
        apiKey);

    var agent = new Agent(root, configuration);
    agent.InstallTools();
    _ = new Skills(agent, new SkillSpec(skillsDirectory, primarySkill));

    bool resumed = agent.StartPersistence(
        new PersistenceOptions(savePath, RetainLog: 200));

    // This writes the restored/bootstrap diagnostic prefix before new facts.
    agent.History().StartLog(eventLogPath);

    // Accept messages through the same actor API as earlier chapters.
    await agent.FlushAsync();
    await agent.ShutdownAsync();
}
finally
{
    root.Destroy();
}
```

Configure persistence once, before actor admission. A second configuration or configuration after actor startup is refused. Explicit `History.Save` and `History.LoadSave` support callers that manage their own lifecycle, but they do not authorize replacing an active Agent’s conversation concurrently.

Shutdown first joins the owned work and destroys Jobs while History remains attached. It records completed shutdown, then saves. Logical destruction follows. A save error remains a failure; it cannot be relabeled as successful shutdown persistence. The destruction-in-progress guard prevents repeated teardown while keeping real parent paths usable until saving is finished.

`Examples/Persistence` is a separate assembly demonstrating snapshot-only save/load, runtime model change, continuous sequence numbers, and automatic shutdown saving. It takes a new output directory and performs no network request.

## 10.8 CLI and GUI behavior

Both applications default to `save.json` in their working directory. Use a distinct directory or explicit path for each independent session.

| CLI/GUI option | Environment fallback | Meaning |
|---|---|---|
| `--save PATH` | `EN_SAVE_PATH` | Primary save; backup is `PATH.bak` |
| `--save-log N` | `EN_SAVE_LOG` | Diagnostic retention; default `-1` |
| `--archive-log PATH` | `EN_ARCHIVE_LOG` | Optional dialogue projection |
| `--no-save` | `EN_DISABLE_SAVE=1` | Disable automatic load/save |

Explicit options override the corresponding environment defaults. The library reads none of these variables. Invalid numeric values or unusable save candidates fail startup rather than silently selecting another session.

The CLI drains accepted work at EOF and joins shutdown before exiting. The GUI saves after its owned host lifetime ends and shutdown completes. A force kill cannot run that code; this chapter does not promise to recover unsaved work after `SIGKILL`.

GUI replay uses reduced Context for finalized transcript content, plus retained event facts for lifecycle and direct-tool diagnostics. It can therefore restore a snapshot-only save. Context entries retain actual response/part IDs; legacy entries still use deterministic replay IDs. Projection does not append synthetic history or restart old jobs.

Live capture bounds both events and Context entries by the same observed sequence. Otherwise a current snapshot could publish a later final before older queued deltas. Reconnect uses a complete consistent snapshot and must not create duplicate cards or revive a saved shutdown state.

## 10.9 Exercises and verification

Use an owned directory, loopback fake provider, dummy key, and isolated browser profile. Do not point exercises at another running Ensemble session.

1. Send several turns, exit normally, restart without replay flags, and inspect the actual next provider request for each prior human turn exactly once.
2. Change vendor, model, primary skill, variable values, and allowed tools. Check that saved conversation survives while current configuration governs the next request and save.
3. Compare snapshot-plus-log, snapshot-only, log-only, and older-snapshot/newer-tail representations. Require equal effective conversation and continuous event numbering.
4. Save opaque values, exact-number arguments, media locators, usage, redacted results, summaries, and display identities. Test the real codec and reducer, not a substitute persistence model.
5. Corrupt primary alone, then both candidates. Isolate invalid cases so a valid backup cannot hide decoder rejection. Verify file and live-state nonmutation on refusal.
6. Block an archive or replacement path. Require a visible failure, no false success, and no diagnostic trimming. Verify a corrupt primary is not rotated over the good backup.
7. Exercise all retention modes and confirm snapshot-only restart works without the optional archive.
8. Restart the real GUI, inspect visible restored cards and their IDs, submit a new prompt, and refresh. Confirm the saved shutdown does not disable the new session.

The reusable executable Chapter 10 grade covers six CLI persistence checks for 85 points. Its C# `-executable-only` route excludes inherited Go checks and the old stdin-driven GUI probe. The GUI is a WebSocket host; its restart behavior is exercised separately through the actual host and an isolated browser, not credited as another stdin client.

Passing local fakes establishes the persistence and protocol behavior exercised here. It does not establish real-provider integration, audible accessibility, or recovery of events never saved. Those limits belong in the progress record alongside implementation and offline evidence.
