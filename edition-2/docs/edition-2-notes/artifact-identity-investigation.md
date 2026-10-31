# Artifact identity investigation: restored history and live streaming

Status: **reproduced; not fixed**. Read-only application investigation on 2026-10-14. Diagnostic fixtures and this note were written; no application code, installed binary, active GUI, or runtime settings were modified. No real provider requests were made by the probe. Codex is working in parallel on compaction/schema/actor lifecycle; coordinate before editing the shared actor code.

## Bill's report and the demonstrated fault

Bill reported streaming text landing in the wrong artifact and recalled the Edition 1 bug. The current interface already carries `part_id`/`ArtifactID`, but uniqueness ends at the runtime lifetime rather than spanning restored history.

`agent/internal/agent/actor.go` constructs `runtime{owner: a}` with a zero `partID`. `displayID` maps parser-local IDs through a per-response map to incrementing runtime IDs. The counter is not restored by `agent/internal/agent/save.go`'s `Load`, which recovers History. Finalized parts retain their old `ArtifactID` in storage and are replayed by `agent/gui/gui.go`.

`agent/gui/web/gui/artifact-scroll.js` keys cards by `String(part_id)`. A delta with a reused ID appends to the existing card; a final replaces that card's text. Resetting the display on reconnect doesn't solve this: replay repopulates the old IDs before new live output arrives.

A structured scan of the actual main snapshot/journal found 470 distinct finalized IDs, of which 385 appeared on multiple finalized parts. Some reuse crossed kinds (old text versus later opaque/thinking content). A later scan found zero duplicate IDs within a single response. Counts describe the recorded scan, not an assertion about every future response. No conversation contents were printed. This establishes a real collision path, not proof that every reported visual anomaly has this cause.

## Isolated reproduction

Evidence is retained privately under:

`/Users/bill/projects/cadence-memories/session/diagnostics/artifact-id-restart-x1xhemna/`

- `probe.go`: real public Ensemble agent construction, Load/Ask/Shutdown/Save, a real streamed response parser and observers, and a local deterministic HTTP provider. Only the provider boundary is fake. Two separate runtimes share saved history. No app source was patched.
- `result.json`: first reply `FIRST` has delta/final ID1, final event sequence3; restored runtime replies `SECOND` with delta/final ID1, final event sequence6. Two local provider requests, no real provider use.
- `renderer-probe.mjs` and `renderer-result.json`: real shipped artifact renderer in an isolated Chrome page served from a local static server. No application shell or live WebSocket connection; it does not take over the connected GUI's browser tools.
- `investigation-metadata.json`: inspected source commit and served-renderer checksum comparison.

The renderer was fed the old finalized part, then the restored runtime's actual delta and final. It showed one card throughout:

1. Replay: `FIRST`.
2. New delta: `FIRSTSECOND` in the old card.
3. New final: only `SECOND`; the first card's displayed content was replaced.

Assertions checked both renderer state and actual DOM text. The served localhost GUI renderer was byte-for-byte equal to the inspected source. The initial browser attempt lacked bundled Chromium; the completed probe used the already-installed Chrome, as the project's existing component tests do. No browser installation or live GUI actions were performed.

The Go probe can be repeated from the `agent/` module with a fresh temporary output directory. The browser probe accepts a static asset-server URL, that run's `result.json`, and an output JSON path. Its current assertions deliberately demonstrate the unfixed bug; they are not acceptance assertions for a future fix.

## Edition 1 precedent

The read-only `artifact-history-review` worker submitted, was joined, and was shut down. Its findings were source/history inspection, not execution of the old test suite.

- `edition-1/book/crossover-changelog.md`, section **Part IDs must be globally unique, not per-response**, records the same visual symptom and Bill's rule: anything talking to an artifact needs a unique ID. Clearing the map every turn was explicitly rejected as hiding the identity defect.
- `72196ddc2bab1219a17730814186784aa5bc5cf3` added a shared local-to-agent-lifetime mapper. Delta and final helpers must share it; independently allocating IDs for them breaks correlation.
- `a1eebb4e` addressed replayed finals without distinct IDs; see `edition-1/agent/gui/replay_test.go`, `TestResponsesGetDistinctPartIDs`.
- `03bd18964695fc20aae4681bac067c74ae3dbab0` added replay/reset behavior and a real-browser restart regression in `edition-1/agent/gui/gui_replay_browser_test.go`.
- `eb15005c8bfe8b747fc1cb9c63cfbb6678a20fd3` ensured Responses finals reuse their corresponding delta IDs; see `TestCh19ResponsesFinalsReuseDeltaIDs`.

## Required identity contract and proposed next work

1. Each artifact needs a stable producer-owned address, unique among everything that can coexist in the view, including restored history. A DOM-only UUID doesn't fix wire routing.
2. Every delta and final for one artifact must use the same address. Do not allocate an ID per chunk or independently on finalization.
3. Handle restart/reload explicitly. A properly persisted/restored allocator or a durable scoped identity can work. Audit public interfaces and coordinate the choice with Codex; this investigation does not choose or implement an API change.
4. Existing saved history already contains collisions. Merely moving the next-ID counter above the old maximum prevents future reuse but does not disambiguate old replayed parts. Plan a deterministic legacy presentation-key strategy without casually rewriting historical events or provider payloads.
5. Preserve agent-selection/reconnect ordering, cancellation filtering, partial handling, tool call/result correlation, and speech behavior. The current server serializes selection with replay; no cross-agent routing defect was established here.
6. Do not clear history on every turn as a workaround. The underlying historical parts remain separate records; the reproduced overwrite happens in the renderer's address map.

Minimum acceptance test: save response A; restore a new runtime; replay A into a fresh/reset scroll; stream B and pause before finalization. Assert A is unchanged and B has its own card. Then finalize B and assert it updates that exact card without duplication. Add same-runtime, repeated-restart, agent-switch/reconnect, interrupted-part, mixed-kind, and legacy-collision replay cases. Use the real actor/persistence and renderer; fake the network provider boundary, not the identity allocator or history reducer.

No fix, successful full GUI regression run, or live corrected behavior is claimed by this note.
