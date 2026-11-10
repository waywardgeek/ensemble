# GUI history window and Agent switching

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
