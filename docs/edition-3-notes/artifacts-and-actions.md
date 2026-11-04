# Artifacts and Actions: development evidence

## Current boundary: representative presentation baseline

Navigation is committed as `9d057020`; Chapter26 is a development placeholder.
No artifact runtime upgrade or student reconstruction is claimed yet.

`agent/GuiTests/actions-baseline.cjs` runs against the real embedded GUI served by
`GuiTests --crossover-host`. It injects synthetic frames through the production
page handler for a command, exact edit, search, refusal, running-Job observation
and report path. These cases verify presentation, not execution or persistence.
The fixture's own host, Agents, Jobs and loopback provider are real components.
No model request or external provider is needed for the new browser test.

The baseline verifies:

- One call and result card for each supplied identity, with the expected tool name.
- Complete received text remains in expanded detail, including search results.
- Exact edit detail distinguishes old and new text.
- A report pathname alone creates no download control.
- Expansion survives rerender and the mobile Tools pane contains its content.

An important existing limitation is explicit: `tool_finished` marks an observation
as finished even when its text reports a running Job. The later presentation
upgrade must distinguish observation completion from ongoing execution using real
facts, not infer execution state from prose. This baseline records current behavior;
it is not the future acceptance rubric and may change with that implementation.

## Reproduction and results

Build with the explicit .NET10 SDK, isolated CLI home/temp and existing offline
assets. Build `agent/GuiTests/GuiTests.csproj` with `--no-restore` and the existing
isolated artifacts directory. This build passed with zero warnings and errors.
Run its `GuiTests.dll --crossover-host` from a fresh temporary working directory;
this fixture currently reserves loopback ports18202/18203. Check availability
first. Wait for READY, then run:

```sh
/opt/homebrew/bin/node agent/GuiTests/actions-baseline.cjs
```

The script uses the repository's existing Playwright dependency and installed
Chrome. It installs nothing. Its default URL is the fixture's `/assistant/` mount;
`GUI_FIXTURE_URL` can override the URL but does not change the fixture's ports.
Create `.crossover/runtime/stop` in that temporary working directory and join the
host process. Browser cleanup closes its owned Chromium instance in `finally`.

Successful local evidence: `/private/tmp/cadence-actions-baseline.Qq1MRc`,
`browser.log` and `host.log`; browser and joined host exit0. The host reported
STOPPED for the fixture, provider and owned Agents. No live application changed.

The first attempt, `/private/tmp/cadence-actions-baseline.Yq95dj`, passed the six
content cases but timed out clicking an invisible detail control. At mobile width,
the browser had not selected Tools. The test now clicks the real Tools tab before
interacting. This was a fixture setup error, not a product expansion defect; the
failed run also joined its host cleanup.

## Next work and acceptance limits

Implement exact copying for complete fenced snippets using the pinned parser.
Rendered code text alone does not establish that a streamed fence is closed.
Specify source/newline handling and test clipboard denial, literal hostile text,
focus/selection and ongoing streaming before claiming this feature complete.

Later slices cover paired state presentation, agreed timestamps/durations and
explicit work-product offers. Public grader points, sensitivity tests, substantive
chapter teaching, fresh student reconstruction and Bill's usability acceptance
remain pending. Students can ask Cadence questions and run public graders under
the [closeout plan](first-book-closeout-plan.md).
