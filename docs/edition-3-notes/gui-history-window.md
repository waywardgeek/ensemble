# GUI history window and Agent switching

## November 10: workflow output identity, wait duration and separate pane limits

A read-only 16-second socket sample found two Python drivers returning separate
MCP observations about every three seconds for Cadence Jobs 32 and 48. Both came
from spawn_sub_agent, remained running and repeatedly reported 91 bytes without
new output. Each ToolObserved event had a new sequence and therefore a new card.
The 100-event display cap did not prevent churn and let Tools displace Chat.

The first source attempt filtered identical background reports. Bill clarified
that wait_for_job must move the existing output to the bottom. That attempt was
removed: suppressing duplicate events did not repair continuing output identity.
The existing browser session code only admitted run_command origins. Background
MCP results also lacked the assistant call card from which it derived tool names.

The final candidate carries recorded operation names independently of call cards.
Running Jobs and waits can establish continuing output views for any recorded
origin, including spawn_sub_agent and ask_agent. Successful plain observations
live inside the same view's collapsed history; wait/input moves that existing DOM
node. Errors, rich evidence, missing identity and unavailable readback preserve
ordinary result presentation. Origin validation, Agent separation, bounded Job
readback and observation cursors remain intact. No audit event is discarded.

ChatHistoryEventLimit and ToolHistoryEventLimit default independently to 100.
The legacy HistoryEventLimit initializer configures both. Projection counts an
event once in each pane to which it contributes; mixed prose/tool responses can
retain their parts independently. State/lifecycle notices have a bounded tail
while final Agent state still reduces all facts. Display-window eviction never
changes saved History or the model's working context.

The fiction helper now supplies `ai_callback_delay: 1800` for subsequent Job waits.
This is a maximum observation duration, not a sleep: output, completion, attention
and the existing hint behavior can wake it earlier. The human supervisor's
separate twenty-minute observation policy is unchanged. Existing Python processes
have not reloaded this edit, and no installed workflow package was replaced.

Verification uses `/tmp/ensemble-poll-card-fix/bin`, empty test workspaces, fake
providers and owned loopback ports. Builds have zero warnings/errors. Projection
checks cover background names without call cards, asymmetric pane limits, mixed
responses, bounded lifecycle replay and compatibility. Socket checks send 150 tool
results without evicting 100 chat events, then 125 chat messages without evicting
100 tool events. The broader GUI consumer passes. Real-browser tests cover 150
workflow waits reusing one output node, grouping, movement, completion, fallback,
eviction and Agent isolation. Existing command tests pass at 1280/390 widths,
including focus, selection, background stability, raw observations and retries;
the parent/child history-window regression passes. Final browser hosts join cleanly.
An earlier browser fixture hit its timeout after old selectors treated newly
collapsed observations as visible; the tests now inspect those disclosed records
explicitly. This was test adaptation to the requested presentation change.

The Job window test uses a 1800-second observation and completes on Job completion
within its 45-second test deadline. The Python adapter's nine checks pass, and all
86 Python tests pass after correcting an existing recovery assertion: searching
for bare c2 matched random run IDs, not just chapter 2. Deterministic run IDs c2…
and ab… reproduced the false failure/pass; the assertion now matches -c2-.
Initial test dependency metadata pointed at an obsolete cache; refreshing solely
from existing local NuGet packages resolved SkiaSharp without a download.

Chapter 30 records these lessons and required next-execution checks. Chapters 7
and 29 explain the resulting GUI behavior. No live installation, restart, model
request or workflow mutation occurred. Browser refresh alone cannot apply the
server projection or reload already-running Python code.

## Earlier installed history-window repair

Bill reported an unresponsive V3 browser and tool cards crossing between parent
and child views on November9. The running Cadence trace contained one replay with
31,488 frames and 9,349 card identities. The prior projection serialized the full
retained Context, and the browser had no live display eviction. Each observation
or 200ms polling wake also rebuilt the projection. Its phase reducer repeatedly
copied a growing temporary dialogue.

A separate browser ownership defect was reproduced: replay cleared pending
renders, then reset Job sessions while the old card map was still populated.
The reset callback scheduled those old results again. A delayed render appended
them to the newly selected Agent's tool pane without checking card ownership.
The new browser test fails on the previous app.js at the parent-result assertion
and passes on the repaired page.

The candidate changes only the GUI assembly:

- `GuiOptions.HistoryEventLimit` defaults to 100 displayed history events.
  Projection selects that window before serializing payloads. An event may have
  several parts; this is not a byte limit or deletion from saved History.
- Live `display_window` frames retire cards, pending renders, fetches and Job
  views that leave the selected Agent's window. Other Agents' frames cannot
  change that window.
- Replay clears card ownership before Job-session callbacks. Both immediate and
  delayed rendering require the same card object in the current Agent's map.
- Unchanged durable snapshots reuse the projection. Capture is monotonic rather
  than rewinding for delayed observations; completed responses reject late
  deltas. Phase reduction retains only the current pending tool batch. Observer
  exit releases its cached snapshot and previews.

Evidence is in `/private/tmp/ensemble-gui-history-repair`. The 10,000-entry
snapshot test, live parent/child socket test, GUI suite, browser ownership/window
test, and existing Job-session, compact-artifact and Markdown browser tests pass.
The browser tests use synthetic observations, not provider inference. Histories
remain intact after display eviction. Existing redaction/handoff tests pass.
One existing Job-session test mutated a card's Agent in place and expected a
rerender; it now verifies that the foreign identity has no session association
and cannot repaint the current view.

The GUI-only candidate in `installed-compatible/bin/Debug/net10.0` was compiled
against Cadence's installed Ensemble assembly. Its existing GuiOptions constructor
and two-argument Project entry point remain compatible. The same GUI/socket and
browser tests pass using those installed dependencies. The installed host executable
also passes three isolated startup/recovery cycles with the candidate GUI and
synthetic credentials. No installation, host restart, model call or live History
change was performed during this repair. Installation remains Bill's decision.

Bill subsequently approved installation and restart. The previous GUI assembly
and symbols are retained in Cadence's
`backup-before-history-window-20261109-230310`. The old process stopped gracefully;
only Ensemble.Gui.dll and its symbols were replaced and byte-verified. Installed
Ensemble/SkiaSharp dependencies matched the tested candidate; host and engine
binaries were not replaced. The restarted process serves a 96,919-byte replay
with exactly100 display events and100 cards (dispatch/state updates are counted
separately). The installed-page browser check observed parent100, child25,
parent100 cards, with no foreign/orphaned nodes or JavaScript errors. Page refresh
loads the newly versioned browser assets. No inference was submitted by verification.
