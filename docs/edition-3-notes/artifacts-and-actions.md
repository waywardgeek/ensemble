# Artifacts and Actions: development evidence

## Current boundary: logical snippet copying, source candidate

Bill approved copying logical code payload, excluding fences and container
prefixes while preserving code whitespace and terminal newline. Parser LF
normalization is accepted; this is not raw Markdown-byte export. The current
browser implementation adds Copy code to explicitly closed assistant-text
fences only. It uses markdown-it tokens, not a second parser or DOM preview text.
Incomplete fences remain noninteractive even when the response is final.

The pinned parser's source span contains one opening line, all body lines and a
closing line only if consumed. The closure check counts an unterminated last body
line as well. Eighteen focused cases cover empty/unfinished fences, EOF without
newline, nested quotes/lists, indentation, mismatched fence types and lengths,
trailing material and CRLF. This depends on the pinned parser's mapping contract;
recheck it on parser upgrades.

Unchanged snippets retain their DOM controls across later prose updates. Copy
captures the logical payload, including when the code is collapsed. Clipboard
success is reported after completion; denied or missing access exposes a labelled
read-only field immediately below the controls. That field retains selection and
focus through streaming. A detached control's completion cannot update a new card.
This is transient browser state, not another persistent content store.

### Verification and limits

Final evidence: `/private/tmp/cadence-snippet-final.zQISff`, including source SHA256
receipt, browser logs, mutant scripts/logs and joined host log. Explicit offline
GuiTests build passed with zero warnings/errors. The three actual embedded-page
suites passed: `markdown-copy.cjs`, `markdown-page.cjs`, and
`actions-baseline.cjs`. The new suite checks exact Unicode/whitespace/LF payloads,
empty and hostile-looking code, nested containers, full long-code copying,
nonassistant refusal, manual fallback and its placement, retained keyboard focus,
late completion and narrow-pane containment. Existing Markdown selection,
coalescing, safety, expansion, replay and scroll regressions still pass.

Only the clipboard permission/transport boundary is replaced with deterministic
success, denial or delay. The tests neither read nor overwrite the operating-system
clipboard. They do not prove native clipboard behavior on every browser or Bill's
phone. No model request, installation, restart or live deployment was performed.
No public grader points or student reconstruction are claimed.

Sensitivity checks served two deliberately changed local `markdown.js` assets
through the browser test's opt-in `GUI_MARKDOWN_SOURCE` override. The unchanged
page and assertions rejected unfinished-fence copying and preview-only copying
at their intended assertions. Default runs use the real embedded script. Final
positive/negative runs and the host all joined normally; host exit0.

The first preview mutant passed because the fixture exceeded the preview's line
limit but not its 1800-character limit. Its mutation did not affect that sample.
The fixture now exceeds both, with an explicit size assertion. Evidence is retained
at `/private/tmp/cadence-snippet-mutants.5OtELb`; corrected sensitivity passed at
`/private/tmp/cadence-snippet-mutants.z2gDVs` and again in the final run. Review also
moved the manual field from below the entire expanded block to below its controls;
the final run includes that placement assertion.

### Teaching and student gates

Chapter7 section7.6 teaches the payload, closure and interaction contracts at the
original rendering construction point, credits Junien's PR7, and adds verification
obligations. Its historical numeric grader does not cover these added obligations.
Chapter26 remains in development and points back to that foundation.

- Chapter7: snippet teaching updated; further Actions upgrades, named public
  grader coverage and final reconstruction baseline/assignment remain pending.
- Chapter26: integrated teaching, exercises and grader remain unfinished; no
  student assignment or completion claim yet.

Fresh students will receive honest prerequisites, question access and public
grader commands when the substantive chapter upgrade is ready. Their candidate
will be independently graded and reviewed; these browser tests are not a substitute.


## Earlier boundary: representative presentation baseline

Navigation is committed as `9d057020`; Chapter26 is a development placeholder.
At that baseline, no artifact runtime upgrade or student reconstruction was claimed.

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

The snippet source increment above is verified within its stated scope. Next
resolve Actions state presentation over real facts, including the distinction
between a completed observation and a still-running Job.

Later slices cover paired state presentation, agreed timestamps/durations and
explicit work-product offers. Public grader points, broader sensitivity checks, substantive
chapter teaching, fresh student reconstruction and Bill's usability acceptance
remain pending. Students can ask Cadence questions and run public graders under
the [closeout plan](first-book-closeout-plan.md).
