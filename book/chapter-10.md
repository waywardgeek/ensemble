# Chapter 10 — Persistence and Replay

## TL;DR

Give an Agent a durable conversation without giving yesterday’s configuration authority over today’s runtime. Save one reduced Context, its `as_of` event sequence, a credential-free description of effective configuration, and two handoff sequence markers. Validate a complete candidate before replacing live state or a save file. On restart, load the snapshot, fold only the newer tail, and number new events above the recovered boundary.

The C# endpoint is `solutions/ch10`. It uses the existing Agent → History relationship. There is no second transcript, persistence service bag, or replacement actor loop.

## 10.1 In plain language

Think of a snapshot as a notebook whose contents have already incorporated events through page number N. The snapshot needs no second copy of those events. Reading already-folded events into the notebook again would duplicate the conversation.

An event newer than N is different: it has not yet been incorporated and must be applied once. If the snapshot is absent, start with an empty notebook and replay the complete event log instead. If the log is absent, the snapshot must still stand on its own.

The notebook records what happened. It does not choose which provider, credentials, model, skills, or tools you are allowed to use now. Those come from the current application’s explicit runtime configuration.

A backup is another candidate, not permission to ignore validation. A corrupt primary must never displace a known-good backup. If neither candidate is usable, stop without rewriting either one.

## 10.2 Ownership and the save contract

Common declares the shared vocabulary; functional modules provide the bodies:

- `Common/Persistence.cs` declares immutable `PersistenceOptions` and `SaveDocument`, plus the operations on Agent, History, Engine, and JsonCodec.
- `History/Persistence.cs` owns capture, replay, candidate selection, and replacement.
- `Serialization/Persistence.cs` owns the neutral JSON format and validation.
- `Engine/Persistence.cs` projects effective configuration without credentials or endpoints.
- `Agent/Persistence.cs` performs startup loading and arms automatic saving. The actor saves after completed handoffs and joined shutdown.
- CLI and GUI composition roots choose paths and policies. Core code does not read environment variables or standard input.

`PersistenceOptions` is value-like. It contains a snapshot path and an audit flag, not a service, owner, or mutable conversation. It needs no parent back-pointer.

A new save has these top-level fields:

```json
{
  "context": {},
  "as_of": 42,
  "config": {},
  "latest_handoff": 40,
  "previous_handoff": 20
}
```

The empty objects above are schematic, not valid complete Context or configuration values.

`context` contains dialogue entries, the current ephemera slot, reducer phase, and measured usage. Each entry carries its sequence, actor, parts, and any finalized display identities. Text, tool calls, tool results, exact JSON numbers, blob locators, opaque provenance, signatures, redaction stubs, and skill instruction values use the existing neutral part codec. Saving does not resolve media files, fetch URIs, or convert the conversation into provider wire JSON.

`as_of` is the event boundary already incorporated into Context. History owns this fact explicitly. It cannot infer it from the last retained diagnostic event: a valid snapshot-only save has no such event.

`config` documents the effective vendor, model, system prompt, and normalized tool declarations. Engine derives it through the existing request projection, using only the captured skill instruction entries. Keys, authorization headers, provider endpoints, and unrelated runtime settings are not configuration fields in the save. Conversation text is still private user data; excluding credentials from configuration does not make a session file public.

`latest_handoff` and `previous_handoff` are sequence markers, not event payloads. Zero means no known boundary. They preserve the working-history cutoff across restart. New snapshots have no `log` field, and `SaveDocument` has no event collection.

## 10.3 Capture and replay

`History.Snapshot()` captures working Events, Context, sequence, and handoff markers under one short lock for observers. Saving captures only Context and scalar metadata; it does not copy the working event collection.

The loader validates each candidate locally before publishing anything. New snapshots are self-contained. For compatibility, the decoder accepts an old embedded `log`: it checks event shapes and increasing sequences, skips facts already folded into Context, and reduces only a contiguous newer tail. A legacy log-only document must begin at sequence one. Decoding returns the same state-only `SaveDocument`, not a retained legacy event array.

Recovery then streams the separate journal, folding only events newer than the selected snapshot and retaining the recent working window for GUI facts. Context, sequence, markers, and working events are published together only after validation succeeds. A failed load leaves live state and disk bytes unchanged. `Agent.StartPersistence` enforces startup before actor admission; Chapter 14 details journal recovery and corruption handling.


Each retained Entry carries its optional recorded time and observation duration.
Serialize the time with its offset and the duration as exact nonnegative ticks.
Missing fields in legacy saves mean unknown, not restart time or zero. Preserve
them through filtering, handoff and snapshot-only replay; a replacement summary
uses its own recorded fact time. This does not extend diagnostic retention:
discarded direct-tool or lifecycle diagnostics are still discarded, not copied
into a second permanent timing log.

A snapshot without any journal must still restore conversation and future numbering correctly; it cannot restore discarded lifecycle or direct-tool facts that are absent from Context.

## 10.4 Safe replacement and recovery

History serializes appends and saves through its save gate, distinct from the short state lock. Readers can still observe the last committed state during disk I/O.

Saving follows this order:

1. Normalize the save path and capture Context, sequence, and handoff markers.
2. Derive credential-free configuration and validate the state-only candidate.
3. Validate the old primary before rotating it into backup. Re-encode its recovered state, rather than copying legacy embedded-log bytes. Never replace a valid backup with corrupt primary data.
4. Write the new primary through a same-directory temporary, flush, and replace the destination.
5. Unless audit is enabled, trim the journal only where both the validated backup and preceding-handoff boundary permit it.

Temporary files are cleaned up on failure. Replacement avoids publishing partially written JSON, but is not a directory-fsync or hardware power-loss guarantee. Chapter 14 explains journal-before-publication ordering and backup-anchored recovery.

Startup tries primary and backup independently with their required journal suffixes. Missing snapshots permit journal-only recovery only from sequence one, or a fresh session when no history exists. Invalid candidates never authorize overwriting recovery data with an empty conversation.


### Persist descriptions, not live file grants

Work-product metadata on a retained successful result survives snapshot and event
roundtrips as ordinary immutable data. Do not serialize the Agent's live offer
collection, file streams or active transfers. Loading a description must never
reopen its path, register its ID, or create download authority. A new Agent with
the same display name is still a new lifetime. Chapter 27 tests retained offer
metadata alongside explicit refusal by the replacement live owner.

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

## 10.6 One working window and one journal

History retains working events from the second-most-recent handoff, including that handoff itself. Before two handoffs are known, it conservatively retains the available prefix. This is an event-history window, not a fixed-byte ceiling: a long interval between handoffs can still be large. Context remains self-contained and is never reconstructed from missing old payloads.

`PersistenceOptions.Audit` defaults to false. Normal saves may truncate the single recovery journal only after both the working-history boundary and validated backup permit it. Audit simply disables that truncation. It allocates no additional event collection and creates no second log or archive. Replay validates records incrementally and retains the same bounded working window in either mode; perpetual audit stays disk-only.

Enabling audit later preserves only the history still available. It cannot manufacture an earlier missing prefix from reduced Context. Memory compression does not create or require a full-event archive. Explicit offline log import remains available, but is not an automatic writer or an additional persistence authority.


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
        new PersistenceOptions(savePath));

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

Both applications default to `save.json` in their Agent data directory. Use a distinct directory or explicit path for each independent session.

| CLI/GUI option | Environment fallback | Meaning |
|---|---|---|
| `--save PATH` | `EN_SAVE_PATH` | Primary save; backup is `PATH.bak` |
| `--audit` | None | Keep the single journal without truncation; default off |
| `--no-save` | `EN_DISABLE_SAVE=1` | Disable automatic load/save |

Explicit options override the corresponding environment defaults. The library reads none of these variables. Obsolete `--event-log`, `--save-log`, and `--archive-log` requests or unusable save candidates fail startup rather than silently selecting another session.

The CLI drains accepted work at EOF and joins shutdown before exiting. The GUI saves after its owned host lifetime ends and shutdown completes. A force kill cannot run that code; this chapter does not promise to recover unsaved work after `SIGKILL`.

GUI replay uses reduced Context for finalized transcript content, plus retained event facts for lifecycle and direct-tool diagnostics. It can therefore restore a snapshot-only save. Context entries retain actual response/part IDs; legacy entries still use deterministic replay IDs. Projection does not append synthetic history or restart old jobs.

Chat and Tools each display the newest 100 contributing events by default,
configured independently by `GuiOptions.ChatHistoryEventLimit` and
`ToolHistoryEventLimit`. The compatible `HistoryEventLimit` initializer sets
both. These are display windows, not deletion from History or changes to model
context. Choose each window before serializing payloads; a mixed event can leave
one pane while remaining in the other. Live updates retire expired cards,
job-output views, pending renders and polling timers without cancelling Jobs.

Live capture uses one current atomic snapshot and caches its projection until
the durable sequence or Context changes. A delayed notification cannot rewind
that snapshot; deltas for already finalized responses are ignored. Reconnect
must not duplicate cards or revive a saved shutdown state. On Agent selection,
clear card ownership before resetting job views, and reject delayed renders whose
card object no longer belongs to the current selection. Matching an Agent name
alone is insufficient when replay has replaced that Agent's previous cards.

## 10.9 Exercises and verification

Use an owned directory, loopback fake provider, dummy key, and isolated browser profile. Do not point exercises at another running Ensemble session.

1. Send several turns, exit normally, restart without replay flags, and inspect the actual next provider request for each prior human turn exactly once.
2. Change vendor, model, primary skill, variable values, and allowed tools. Check that saved conversation survives while current configuration governs the next request and save.
3. Compare snapshot-plus-log, snapshot-only, log-only, and older-snapshot/newer-tail representations. Require equal effective conversation and continuous event numbering. Include exact retained recorded times and observation durations; remove timing fields to test legacy unknown values, and reject malformed times or negative durations.
4. Save opaque values, exact-number arguments, media locators, usage, redacted results, summaries, and display identities. Test the real codec and reducer, not a substitute persistence model.
5. Corrupt primary alone, then both candidates. Isolate invalid cases so a valid backup cannot hide decoder rejection. Verify file and live-state nonmutation on refusal.
6. Block the journal append or snapshot replacement path in separate fixtures. Require a visible failure and no false success. A failed append must poison further writes without publishing its event; failed snapshot replacement must not trim the recovery journal. Verify a corrupt primary is not rotated over the good backup.
7. Exercise audit on and off. Require the same recent resident event window, an untruncated audit journal, and backup-safe trimming with audit off. Confirm snapshot-only restart restores Context without inventing missing historical events.
8. Restart the real GUI, inspect visible restored cards and their IDs, submit a new prompt, and refresh. Confirm the saved shutdown does not disable the new session.

The reusable executable Chapter 10 grade covers six CLI persistence checks for 85 points. Its C# `-executable-only` route excludes inherited Go checks and the old stdin-driven GUI probe. The GUI is a WebSocket host; its restart behavior is exercised separately through the actual host and an isolated browser, not credited as another stdin client.

Passing local fakes establishes the persistence and protocol behavior exercised here. It does not establish real-provider integration, audible accessibility, or recovery of events never saved. Those limits belong in the progress record alongside implementation and offline evidence.
