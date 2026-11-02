# Agent-owned and Ensemble-owned diagnostics

## Approved contract; implementation pending

Bill requests one diagnostic owner per Agent and a separate Ensemble owner for
host-wide activity. Fresh diagnostic logs on each host restart; durable History,
save.json, backups, event journals, memories and credentials are not diagnostic
logs and must not be truncated. API/GUI work belongs to its actual Agent, including
ordinary children. Preserve write-time credential masking and serialized writes.
Close owned writers after owned work has joined. No automatic installation or
restart: Bill will rebuild/restart after verification.

Current source opens GUI api.log/debug.log/gui.log in append mode. Observed live
api.log was10,562,990,415bytes and gui.log1,004,212,842bytes; both open. Those sizes
span restarts, not evidence of growth in one session. Live logs remain untouched.
Cleanup removed~2.54GiB obsolete builds/caches/test runtimes, preserving sources,
reports, current builds and MeasureX. Browser-local theme changes remain separate.

Review gates: startup restoration failure ordering; per-owner synchronization and
joined cleanup; real Agent/API/GUI routing; socket relay vs selected Agent; queued
frame ownership across selection; multiple peers and nested child isolation;
restart freshness without durable-state loss; shared redaction regressions.
Update Chapter27 crossover prose and acceptance requirements after verification.
Three-minute root goal enabled by explicit request; no new workers/provider probes.

## Core ownership checkpoint

Agent-owned debug/API/GUI writers now serialize writes and close after owned work
joins. Engine API traces and runtime debug messages from Agent, Engine, CacheLens,
Skills and Jobs route through the real Agent. Ensemble retains host diagnostics
and centralized write-time credential masking; no per-logger credential copies.
Partial-open cleanup attempts every writer and preserves the original failure.

Focused build and four diagnostic groups pass, including two independent Agents,
automatically configured ordinary child and grandchild destinations, restart
truncation with a durable sentinel preserved, bearer masking and streaming traces.
The expanded child fixture initially used the synchronous test registration after
becoming async; corrected to AsyncCase before the successful run.

Startup and GUI wiring remain pending. Persistence recovery messages still use the
host diagnostic path deliberately: moving them before an Agent writer exists would
lose recovery evidence. GUI MCP relay ownership differs from selected conversation
ownership, and send-time selection must not reattribute queued frames. Current live
logs are unchanged. These checks are not yet full-host restart acceptance.

## Host and GUI integration

GUI and ordinary CLI hosting now enable the Agent file policy. Ensemble opens its
own fresh ensemble.log; the root opens its own files after StartPersistence.
Children and peers apply the same policy after their durable-state admission.
Root-only CH06_API_LOG/CH07_GUI_LOG overrides do not propagate to descendants.
Library Agents remain opt-in, preserving intentionally nonpersistent examples.
Explicit GuiOptions.TracePath remains an aggregate embedding/test override; null
is normal per-owner routing. Outgoing conversation records are labelled queued,
MCP sends use the fixed relay Agent, and incoming commands use the actual captured
selection/relay even when client addressing is refused. Host-wide frames remain
on Ensemble. Current live logs and installed runtime remain untouched.

Real WebSocket routing checks pass for independent selections, falsely addressed
controls, queued output and fixed MCP relay identity across selection. Existing
public GUI groups, image admission/privacy and capture transport/privacy checks
passed. The image fixture now exercises normal owned logging; its final result
must be collected before claiming that updated path passed. Core tests include
fresh host logs, borrowed-writer survival and failed Agent restoration preserving
previous diagnostic evidence. GUI and CLI builds pass. Gateway adapter builds
using existing local libsodium packages after stale/missing restore manifests
were corrected offline; no network package download was used.

Initial shared run:334 passes/six failures. Five were fixtures still observing
Ensemble's borrowed writer after their diagnostics moved to Agent: MCP auth,
three cache-lens assertions, and a Job completion signal that timed out. Attach
those writers to their actual Agent without weakening assertions. The remaining
automatic-interrupt event-count failure passed unchanged on the next run; its
cause is not established. Second run:339 passes/one old Job-signal failure.
Final corrected shared run is in progress. Focused diagnostic4/0, cache-lens4/0
and HTTP-MCP3/0 pass. Chapter27 now teaches the contract; its first lint found four
existing em dashes, corrected narrowly. No final lint/shared acceptance yet.

## Final source verification

Final shared C# suite:340 passed, zero failed (job199; full output
/private/tmp/cadence-logging-shared-final.log). Owned GUI image/privacy and fixed
MCP relay tests passed (job201); five public GUI groups and seven capture groups
passed earlier in job173. Diagnostic4/0, cache-lens4/0, HTTP-MCP3/0 pass (job186).
GUI, CLI, Gateway adapter and Gateway Agent host compile with zero warnings/errors.
Chapter27 hard lint and whitespace checks pass; style advisories remain. Gateway
startup uses the same host-before-restore/Agent-after-restore policy; its executable
was built, not launched against a live Gateway. No installed runtime or open live
log has been replaced or truncated. The earlier intermittent automatic-interrupt
failure remains recorded, not assigned an invented cause by the final passing run.

The logging change is scoped independently from unfinished browser-local theme
edits. A build directly from the current working tree would include those edits;
a logging-only installation must build the committed source separately or wait
for the theme integration check. Bill owns rebuild/restart. Standalone DocumentChat
remains opt-out and its intentionally disabled transcript diagnostics stay disabled.
