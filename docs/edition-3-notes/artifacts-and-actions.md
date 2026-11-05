## Named public artifact runner: partial coverage

`cmd/grade-edition3 --scope artifacts` now runs grader-owned public consumers
against an explicit runtime assembly directory. It preserves historical `-ch26`
media grading and refuses an explicit `-ch` or unrelated flags in the named scope
before launching tools. The runtime directory must contain Ensemble.dll and
Ensemble.Gui.dll. Supply existing absolute SDK, Node and Playwright paths:

```sh
grade-edition3 --scope artifacts -executable-only -json \
  -dotnet /absolute/path/to/dotnet \
  -browser-node /absolute/path/to/node \
  -playwright-module /absolute/path/to/playwright \
  /absolute/path/to/candidate-runtime
```

The current runner uses macOS/Linux with an existing Chrome installation. No
package installation or real provider credentials are used. The runner creates
isolated homes/workspaces, clears NuGet package sources, builds only grader-owned
consumer sources against candidate assemblies, bounds output, and owns subprocess
groups. Browser hosts stop through their stop file and must join successfully;
terminal receipts independently require three requests, two tool returns, one
offer and no oracle failure. Both named and unnamed viewer modes run.

The report distinguishes scored rubric rows from partial probes. Currently the
core probe and two browser probes can pass, but all five rubric rows remain
explicitly UNGRADED and award zero points. Exit0 is reserved for a full pass,
exit1 reports a failed probe, exit2 is argument/environment setup failure, and
exit3 means coverage is unfinished. A passing probe is not a complete row or
chapter certification. Historical report formats and exit behavior are unchanged.

Verification: both Go packages pass; the opt-in external candidate runner test
passes. The actual CLI check in `/private/tmp/cadence-artifact-cli.xeYGxV` confirms
three passing probes, five ungraded rows,0/100 and exit3. Argument tests use a
sentinel executable and require invalid mixed-scope/missing-candidate invocations
to refuse without launching it. Receipt tests reject missing terminal facts and
bare success without required core checks. Independent behavioral coverage and
negative candidates for the remaining rows are still required before scoring.

## Independent browser consumer and discovered identity defect

The public grader's work-product browser driver now uses only the public page,
HTTP routes and grader-owned inspection endpoints. It never calls page internals
or injects tool frames. A loopback provider requests real write_file and
offer_work_product calls, checks the actual randomized file bytes and recorded
result before each next request, and finishes after exactly three requests. The
browser changes the file through the isolated grader host, verifies the downloaded
hash and filename, checks click-only transfer, HTTP guarding and replacement expiry.
This is a partial contract receipt, not the complete artifact rubric or points.

Two fixture errors were corrected: GUI settings initially lived outside the
viewer Sandbox, which correctly refused startup; the driver initially selected
an always-hidden compatibility select rather than the visible Agent tree. The
latter run issued zero model requests. Cleanup now attempts host and Agent teardown
even when serving faults. Neither failure was counted against candidate behavior.

The next run passed the browser assertions but exposed a real GUI observer fault
on replacing the selected Agent. AddressOf converted a destroyed Agent's missing
address to the empty mount-local alias, and queued diagnostics passed that empty
value to the global address parser. Evidence:
`/private/tmp/cadence-artifact-browser.eaKNxN/host.log`. The repair reserves the
empty alias for the actual mounted source, treats address loss by another selected
Agent as disposal, and resolves valid empty aliases directly to the source for
tracing. Chapter7 teaches the distinction at the construction point.

The repaired candidate passes named and unnamed viewer modes with clean joined
exit in `/private/tmp/cadence-artifact-public.w1jTM8`. Each result records three
provider requests, two tool returns and one successful offer; real changed-file
hash checks pass. Diagnostic routing and five existing public GUI groups pass
job24. No full runtime rerun or universal GUI-race claim is implied. This fix is
source-only; Bill's running main instance was not rebuilt or restarted by Cadence.

## Live usability report and public-grader work

Bill reports rebuilding and restarting the main instance himself. A fresh
list_directory call was then issued, and Bill said it seemed to work. This is
live acceptance of that fresh call/result presentation, not independent installed
binary identity verification or acceptance of snippet copying and downloads.
Cadence made no installation or restart for this check.

The public artifact grader is now under construction. Its first grader-owned
outside-assembly consumer is in internal/gradeedition3/testdata/artifacts-consumer.
It references candidate assemblies, creates independent randomized files and
skills, and uses public APIs. It does not invoke candidate test suites. This
initial core-only receipt awards no chapter points. Subsequent browser and CLI
work is described above; complete rubric coverage and sensitivity remain open.

## Work products: integrated source candidate

Bill approved explicit granted offers, human-click transfer without a second
approval dialog, click-time contents rather than immutable storage, expiry on
Agent destruction/restart, 64 MiB per transfer and 64 offers per Agent lifetime.
Additional offers refuse at the cap; existing offers are not silently evicted.
Image-input and existing completed-Job-output limits remain separate.

The Agent owns immutable live offer records behind its lifetime gate. IDs are
opaque; retained WorkProductInfo exposes ID, filename and size when offered, not
a host path. Both offering and opening use the same Sandbox regular-file read.
Failed validation consumes no slot. Deserializing a result restores metadata only,
never the live collection. Typed metadata follows successful ToolResultPart/Event,
codec, numbering, Context, structured Job completion and GUI projection paths.
Error results cannot offer files; result redaction drops metadata with the payload.

The new work-products skill explicitly grants offer_work_product. Built-in
installation and primary-skill discovery do not grant it; legacy unskilled model
declarations exclude it. Unload prevents new offers. The packaged full/ensemble
primary manifests list it as loadable; installed skill files are unchanged.
Provider text describes the offer without transmitting its opaque ID. A focused
Anthropic rendering assertion checks actual JSON properties and ID absence.

CopyWorkProductAsync owns an in-flight transfer with an explicit Agent parent,
opened file, linked cancellation and joined cleanup. The destination is borrowed.
Agent destruction closes admission, cancels transfers and joins cleanup before
Sandbox teardown. Copying and cancellation callbacks run outside the ownership
gate. The HTTP route accepts only Agent identity and offer ID, checks the existing
origin/custom-header boundary, and serves no-store, nosniff, download-only binary
content. Failed partial transfers abort rather than report a truncated success.
This is not a new GUI login mechanism or a file snapshot resistant to concurrent
same-size modifications.

The browser renders controls only from successful typed metadata. No transfer
occurs on render. It captures the card's owner/offer, checks both declared and
streamed 64 MiB bounds, supports cancellation, retains controls through rerender,
and checks current card identity before initiating a browser download. Replay
aborts discarded cards' requests. Expired offers show refusal, not a fresh grant.

Verification: core --work-product-tests passes3 groups for ownership/click-time
bytes/caps/replacement symlink, real skill and delayed Job metadata/replay, and
barrier-controlled cancellation/joined Agent teardown. The output stream remains
borrowed; a blocked copy does not block another. GUI --work-product-tests passes
the real loopback HTTP route, including exact64 MiB transfer, oversized growth,
missing file, foreign owner, same-name replacement, origin/header and path-query
refusals. The additional slow-response test described below covers actual HTTP abort.

Final browser/shared evidence: `/private/tmp/cadence-work-products-final.01eDyF`.
Seven actual-page suites pass, including exact real-route post-offer file hash,
click-only transfer, original filename, cancellation, expired metadata, exact
64 MiB browser success, overflow with and without Content-Length, and stale-card
completion refusal. Truncated-bytes and stale-completion mutants fail intended
assertions. Fixture/provider/owned Agents join with exit0. Shared suite passes
350/0 from the repository. Earlier browser evidence is
`/private/tmp/cadence-work-products-browser.3mqruk`.

Corrections retained: the first provider oracle matched the legitimate tool name
with a broad substring; it now checks JSON property names plus opaque-ID absence.
The first shared run passed348/1 because an exact ToolTests built-in inventory
omitted the new name. A second legacy declaration expectation also needed to
exclude the explicitly gated tool. Both exact expectations were corrected without
changing production authority. The corrected pre-transfer shared run passed349/0
in `/private/tmp/cadence-work-products.jVxCLY/shared-final.log`; it is earlier than
the final350/0. Browser request tracking was moved before navigation so an
accidental replay-time download cannot evade the click-only test.

Targeted follow-up: job266 passes real client-abort and mount-stop tests using a
barrier-controlled response sink at the test host's external I/O boundary. The
real socket, route, Agent and Sandbox remain in use. Cancellation reaches the
copy, and both request and mount remain unfinished until cleanup is released.
Job268 in `/private/tmp/cadence-work-product-cancel.jSP234` passes the actual replay
handler test; omitting replay cancellation fails the named assertion. The fixture
joins with exit0. Only these two test files changed after the earlier350/0 run;
production source still matches that receipt. No broad rerun was needed merely
for freshness. Final ownership review covered admission, both membership links,
borrowed response streams, cancellation outside the gate and joined teardown.

Remaining gates: public artifact grader and its sensitivity,
coherent Chapter26 exercises/integration, fresh student reconstruction for the
substantively changed foundations, and Bill's usability acceptance. No installed
application, service, credential, release output or publication was changed.

## Recorded time and observation duration: source increment verified

Bill approved finalized fact recorded time, local display with exact timestamp,
and monotonic dispatch-to-result observation duration, not background Job runtime.
No streaming spans, whole-turn profiler or live timer is included.

History-assigned UTC time now survives in immutable Entry values, their snapshot
codec, filtering and handoff. A replacement summary carries its replacement fact's
time. Historical absent time stays null. The GUI reconstructs finalized facts
from these entries, so snapshot-only replay does not depend on retained diagnostic
Events for dialogue timestamps. Existing diagnostic retention remains unchanged;
direct-tool observations removed with that tail are not made permanent elsewhere.

An internal immutable ToolObservation returns the existing result plus nullable
elapsed time. Jobs captures a monotonic clock after successful ToolStarted append
and stops it when the foreground observation has returned. Agent records the
value on ToolReturned/ToolObserved; retained tool entries preserve it. Nullable
nonnegative ticks roundtrip exactly, while GUI frames expose milliseconds. There
is no shared timer registry, new Job lifetime or provider payload field. Refusals
and skill activation without that explicit dispatch boundary remain untimed.

Focused core tests pass: exact timestamp/offset and duration persistence, legacy
omission, malformed/negative values, filtering, replacement summaries, handoff,
and actual History append. A real barrier-held custom tool proves a running Job
can outlive its recorded observation duration; later completion does not rewrite
that result. The outside-assembly GUI projection test passes with an empty
diagnostic tail, known original times and legacy unknown values.

Browser evidence: `/private/tmp/cadence-timing-check.raf8eH`. Timing, pairing,
state, baseline, copy and Markdown suites pass; timestamp-now and duration-zero
mutants fail their assertions, as do the earlier pairing mutants. The fixture,
provider and owned Agents exit0 after joined stop. Native acceptance, public
grader points and fresh chapter reconstruction are not established.

Failures and corrections: the first focused handoff fixture supplied TextPart
instead of required HandoffPart; pure reduction correctly skipped it. The first
shared run after duration edits launched from the browser fixture directory,
causing thirteen missing-repository-fixture failures. A fourteenth failure was
an overstrong new test: a zero-delay refused call's first report need not yet
contain its worker's error. The corrected test checks no ToolStarted/no effect
and absent duration, rather than the first report's IsError. The corrected shared
run passes347/0 with exit0 in `shared-repo.log`; earlier `shared.log` is retained. A prior
recorded-time-only shared run passed346/0 and does not certify duration changes.

Chapters2/4/7/10/26 receive construction-point teaching and verification criteria.
Their student/public-grader gates remain open. The final test-registration edit
only corrects indentation; runtime and GUI source match the tested candidate.
Hard lint failures in Chapters2/4/10
are pre-existing em-dash lines, compared unchanged with HEAD; Chapter7/26 hard
checks pass with advisories. No installation, restart or publication occurred.

## Pair navigation and stable tool controls

The browser now keeps heading, details, preview and output controls instead of
replacing an entire tool card on status updates. Raw text changes only when its
value changes; selection continues to defer rendering. Current card relationships
provide View call and View latest result. Activation checks object identity,
selected Agent and connected nodes, refusing stale controls after replacement.
Pair controls have 44-pixel targets. Focus or selection inside a pane prevents
incoming output from pulling the reader to the bottom.

`agent/GuiTests/actions-pairing.cjs` uses the actual embedded page and synthetic
wire facts. It checks retained summary and pair-button focus, expansion, output
control identity and feedback, selected-text deferral, focused-reader scroll,
latest and earlier result navigation, missing/wrong-kind counterparts and detached
controls with reused IDs. The pending-download case sets the control state
synthetically; it does not claim new transport cancellation coverage.

Final evidence: `/private/tmp/cadence-pairing-check.3dACzV`. Pairing, state,
baseline, snippet-copy and Markdown suites pass. Whole-card replacement and
stale-navigation mutants fail the intended assertions. Explicit SDK build has
zero warnings/errors; browser and joined host exit0. Source hashes are retained
in that directory. No installation, public-grade points, student reconstruction
or human usability acceptance is implied.

Earlier evidence `/private/tmp/cadence-actions-pairing.AfThSs` passed the initial
five suites. The expanded selection test in
`/private/tmp/cadence-actions-pairing-final.wNr2wX` initially selected text in closed
details and obtained an empty selection. Opening details corrected the test setup;
no product fix was inferred from that failure. Both hosts joined cleanup.
Review also caught that requiring a result node to be connected before retaining
the call button would remove focus during the result's first insertion. Rendering
keeps the valid current-object link; activation still requires connected nodes.

Lesson: retaining an expanded value is not retaining the actual control and its
interaction state. Tests must exercise the reader's visible state, not a selection
inside hidden detail.

# Artifacts and Actions: development evidence

## Current boundary: observation labels and recorded Job state

The browser now labels existing machine states Requested, Dispatched/awaiting
result, Result returned and Error reported. Tooltips use the same wording. Raw
`planned`/`started`/`finished`/`error` values and speech decisions are unchanged.
A tool result's existing typed Job metadata supplies a separate handle and
state-when-observed line, including its optional exit code. Running/done/killed
map to running/finished/killed. Unknown values stay unavailable; missing exit code
is not zero. No result prose is parsed into Job facts.

This uses the existing JobInfo/codec/projection path, whose source was inspected.
No new poller, status registry, persistent field, Job behavior or provider request
was added. An old observation stays unchanged after a newer result for the same
Job; its label intentionally does not promise present execution state. Error
reported is not a claim that the underlying process failed rather than the
observation. Exact raw detail remains available.

Final evidence: `/private/tmp/cadence-actions-state-final.F5B8mq`, including source
SHA256 receipt, logs and two changed candidate scripts. Explicit offline GuiTests
build passed with zero warnings/errors. `actions-state.cjs`,
`actions-baseline.cjs`, `markdown-copy.cjs` and `markdown-page.cjs` passed against
the actual embedded page. Label and recorded-running-to-finished mutants failed
the intended assertions through opt-in local `GUI_ACTIONS_SOURCE` asset overrides.
The fixture and owned Agents/provider joined with exit0. No live service changed.

The new state tests inject synthetic wire frames into the production page handler;
they establish presentation and retained-card behavior, not a new end-to-end Job
lifecycle or persistence acceptance result. They cover requested/dispatched/error,
recorded running/done/killed, nonzero/missing exit codes, malformed handles,
unknown status, contradictory prose, later observations, tooltips and mobile
containment. Earlier pre-tooltip evidence is retained at
`/private/tmp/cadence-actions-state.crF020`, not substituted for the final run.

Chapter7's original tool-state teaching now distinguishes readable labels from
machine states and explains recorded rather than live Job status. Pair navigation,
focus preservation for tool-card updates, public artifact grading and fresh
student reconstruction remain pending. This is a bounded presentation increment,
not completion of the Actions chapter.


## Earlier boundary: logical snippet copying, source candidate

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

The snippet, recorded-state, pair-navigation, timing and work-product source
increments above are verified within their stated scopes. The named public
artifact grader remains partial. Public grader points, remaining HTTP/browser
coverage and sensitivity checks, substantive chapter integration, fresh student
reconstruction and broader human usability acceptance remain pending. Students can ask Cadence questions and run public graders under
the [closeout plan](first-book-closeout-plan.md).

Foundation lint follow-up: Chapters3/4/10 retain exactly the pre-existing em-dash
lines compared with HEAD; Chapter18 passes hard checks. Advisories remain. The
printed lint exit1 was retained even though a subsequent log-print command exited0;
this is not recorded as a full manuscript lint pass.

## Independent core grader: limits, cancellation and restored authority

The grader now owns literal expectations of 64 MiB and 64 offers. Deriving a
boundary from a candidate constant could make the oracle repeat the same defect.
Its public consumer checks exact-size copy length and SHA256, oversized growth,
missing files, outside-root offering, replacement symlink escape, failed offers
consuming no slot, 64 unique grants, refusal of a 65th grant and retention of the
first grant. Failed result/event values cannot carry offer metadata.

A grader-owned blocking destination controls only external output I/O. Real
candidate copies demonstrate independent concurrent transfer, caller cancellation,
Agent destruction waiting for cleanup, and borrowed destination ownership. A real
History.LoadSave restores the description into a replacement Agent but cannot
restore permission to open it. Codec roundtrip alone would not test that loader
boundary.

Positive external runner results: job46 passed in 7.08 seconds; after adding the
actual restore checks, job49 passed in 6.47 seconds, including core plus named and
unnamed browser hosts. Both Go packages passed in job54. These are scoped results,
not full chapter certification.

Negative evidence is retained at
`/private/tmp/cadence-artifact-mutants.8IGYER`. All four deliberately broken core
candidates compiled successfully and failed the intended behavioral assertion:

- 16 MiB limit: `exact-64mib-copy-hash`.
- 65-offer cap: `offer-cap-refuses-65th`.
- Click-time Sandbox bypass: `replacement-symlink-escape-refused`.
- Recreating a grant from restored metadata: `restored-description-cannot-grant-read`.

The first three ran in job48 before the restore assertions were added. Job53 built
the current consumer and the fourth mutant; 32 earlier checks passed before the
intended restored-authority failure. Only grader-created files were used. Separate
temporary output directories preserved the good runtime and working source;
package sources were disabled. No application installation or restart occurred.

All five rubric rows remain UNGRADED, zero points. Public HTTP/browser size,
cancellation and replay coverage is still incomplete; internal application tests
do not fill those public grading gaps. Actions, snippets, recorded-time and
interaction drivers also remain unfinished.

## Public browser bounds, replay and HTTP cancellation

The external host now supplies a grader-owned random 64 MiB file after the real
model/tool offer. The public browser downloads it through the candidate route and
checks length and SHA256. Growing the same file to 64 MiB plus one byte must refuse
both direct HTTP access and browser saving. The refusal carries no attachment
header; the browser shows a failure beside the offer. The oracle owns its numeric
boundary, rather than reading a candidate limit.

A page reload replays the retained offer without another provider request, tool
effect or automatic transfer. An explicit click still reads from its live owner.
For cancellation, grader-owned middleware slows only the external response stream
and exposes barrier status. Browser Cancel download and selection of the other
Agent each reach the real server copy's cancellation token. The request remains
pending until cleanup is released, then joins without saving partial bytes. No
application globals, injected wire frames or candidate test assembly are used.
The barrier borrows the actual response stream; host cleanup releases it even
when a browser assertion fails. These fixture endpoints are outside the candidate.

Browser receipts now require every named check. Unit checks remove each required
name in turn and verify rejection; bare passed:true is insufficient. All scoring
rows remain UNGRADED and award zero points.

Job67 initially failed because the new oracle guessed HTTP413. The actual route
intentionally groups expired, unavailable and oversized offers as HTTP404 before
response bytes start. The corrected check requires that refusal, no attachment
header and no browser download. This changed the grader, not the product. Job70
passed the corrected bounds; job73 passed reload and receipt checks; job76 passed
the added real cancellation paths. Final job81 passed both Go packages and the
full external runner in 8.16 seconds, including named and unnamed viewer modes
and joined host cleanup. These are public-boundary results, not a full rubric.

Mutation evidence lives at `/private/tmp/cadence-artifact-browser-mutants.YLDmIg`.
The 16 MiB browser-cap candidate built and was rejected in both viewer modes at
`exact-64mib-browser-hash`. The no-replay-abort mutant also built and failed in
both modes because the HTTP transfer never reached cancellation after selection
changed. Job79 completed successfully as a sensitivity campaign; both mutants'
hosts reported STOPPED without joined-shutdown errors. These intentional candidate
failures are preserved in the result JSON, distinct from setup failures.
Both are isolated embedded-asset builds; good runtime, installed applications and
live services remain unchanged. Real failed/running-offer coverage and additional
sensitivity checks remain before awarding the work-product row. Actions, snippets,
recorded-time and interaction rows still require independent drivers.

## Real failed offers in the public provider/browser path

The external host now requests write_file, a successful offer, then an offer for
a nonexistent grader-owned file. Before sending the final answer it requires the
failed ToolReturned fact to be recorded with IsError and no WorkProduct metadata,
only one successful offer description overall, and no created missing file. The
browser requires the real error card to have no work-product controls. Terminal
receipts now require four provider requests, three returns, one successful offer
and one failed offer. Unit tests remove or corrupt each receipt field independently.
The browser monitors the grader-owned provider failure status while awaiting the
answer so negative candidates report the actual oracle failure instead of only a
missing-answer timeout.

Job88 initially rejected the good candidate because the assertion matched every
Download button on an error card. A failed Job legitimately has Download complete
output for diagnostics; that is not a work-product grant. The corrected assertion
checks work-product controls specifically and preserves the diagnostic feature.
Job91 then passed. Final job93 passed both Go packages and the external runner in
7.02 seconds; named and unnamed hosts joined cleanly.

Job92 built an isolated core mutant that caught a file-open IOException and
incorrectly returned successful offer metadata. Both browser modes rejected it at
`failed-offer-must-not-create-download`; the core probe still passed and hosts
reported STOPPED without joined-shutdown errors. Evidence is retained in
`/private/tmp/cadence-artifact-failed-offer.RJhMZp`, including build log and JSON
report. No installed application or good candidate binary changed.

Running-offer coverage remains a distinct gap: the current core loop checks any
running observations it encounters, but scheduling does not guarantee that it
encounters one. This is not claimed as deterministic running-state certification.
All five rubric rows remain UNGRADED with zero points; next work covers that gap
and the remaining public Actions/snippet/time/interaction contracts.

## Independent running Actions probe and control-observation repair

The public artifact runner now includes a fourth partial probe, using a separate
outside-assembly Actions host. A public custom callback remains behind a barrier
while its real zero-delay observation is recorded. The fake provider waits for
browser release, then observes the same terminal Job through `wait_for_job`.
The browser checks actual call/result navigation, returned-versus-running labels,
retained historical running state, pathname output without work-product controls,
and reload without callback or provider reexecution. No injected application
frames, internal card maps, candidate test assembly or real provider is used.
This tests ordinary Job observations, not deterministic delayed built-in offer
metadata; that narrower public seam remains unavailable.

The probe exposed a real recording defect: Agent discarded `result.Job` for all
control calls. `wait_for_job` supplied typed metadata, but canonical History lost
it before GUI projection. Controls allocate no wrapper Jobs; their observations
can still describe existing Jobs. Recording now preserves the result metadata.
Focused Job assertions compare both running and terminal returned observations
with their recorded History facts. Chapters4 and7 clarify the distinction.

Evidence in `/private/tmp/cadence-actions-public.iTpa8D` includes clean GUI/test
build logs, retained pre-fix core assembly, runtime diff, shared350/0, Go package
results and final external four-probe pass (8.99seconds). Job124, before the
recording repair, failed specifically with terminal status missing and error
false; host shutdown completed. Job138 passes after the repair, including the
actual provider-to-History-to-browser path. No installation or restart occurred.
The good candidate runtime in temporary build outputs now contains the repair.

Earlier setup job116 failed before readiness because the primary skill named a
custom tool before its constructor registered it. The fixture now registers only
its two required public tools directly, independent of the separate skill-based
download host. Job120 reached the running pair but failed the terminal oracle.
Waiting on public terminal state before requesting the final observation removed
a timing ambiguity, but job124 proved the actual problem was missing metadata,
not scheduling. The oracle was retained rather than replaced with prose parsing.

Both Go packages pass. Receipt omission tests require every Actions browser check
and terminal count. Chapter7 passes hard lint with advisories; Chapter4 fails only
four pre-existing em-dash lines verified identical to HEAD. All five rubric rows
remain ungraded with zero points. Requested/dispatched transitions, additional
Actions interaction and negative candidates, snippets and timing public coverage,
chapter reconstruction and human acceptance remain separate unfinished gates.

### Public Actions sensitivity checks

The public Actions driver now locates the actual first result before asserting
its recorded running label. A label-based locator would turn a wrong label into
a timeout instead of identifying the violated observation contract.

Two isolated GUI candidates built successfully from commit9a944dfd. Changing the
running label to finished failed `real-running-observation must display running,
not finished`. Focusing the source card instead of its counterpart failed
`View call focuses a real call card`. The three work-product probes passed in
both runs; each Actions host reported STOPPED without a joined-shutdown error.
Evidence: `/private/tmp/cadence-actions-mutants.pp1uvZ`, including build logs and
JSON reports. The good runtime was unchanged. Both Go packages passed with the
external candidate enabled after the assertion change (job156).

These checks establish sensitivity for the exercised running label and real pair
navigation. Requested/dispatched transitions, broader interaction behavior,
snippet and timing public drivers remain open. All five rows remain UNGRADED;
there is no new chapter score, student acceptance, installation or restart.

### Public focused-reader continuity

The real Actions scenario now expands the running result's raw payload and
focuses its actual summary node before releasing the callback. After the later
terminal observation arrives, the same node must remain connected and focused,
its details expanded, and its complete raw text unchanged. A required named
browser receipt prevents silently omitting this check.

Initial job162 failed because a broad result locator matched both observations
after completion. The retained-node focus assertion passed; the subsequent
expansion query was ambiguous. Binding that query to the recorded result ID
corrected the fixture. Job165 passes both Go packages with the external candidate
enabled. An isolated focus-stealing GUI built successfully and failed exactly
`later-observation preserves the actual focused reader control`; its host stopped
cleanly. Evidence is `result-steal-reader-focus.json` and the corresponding build
log in `/private/tmp/cadence-actions-mutants.pp1uvZ`. This negative run used the
preceding locator version but failed before that query, at the unchanged focus
assertion. No product change was needed.

This is focused-node, expansion and payload continuity evidence, not complete
scroll, selection, streaming or interaction-row certification. Rows remain
UNGRADED. Installed applications and live services remain untouched.

### Public canonical timing in the real Actions page

The independent Actions host now exposes synthetic canonical ToolReturned timing
facts through its grader-owned status endpoint. The browser compares displayed
result timestamps, exact details and observation durations with those facts,
before and after completion and after page reload. A Honolulu browser timezone
checks local presentation separately from the retained exact timestamp. Later
completion must leave the earlier observation unchanged. Required named receipts
cover this live/replay boundary; no application globals or injected frames are
used.

Job170 passes both Go packages with the external candidate enabled. Job171 built
two isolated GUI candidates: replacing recorded time with the browser clock
failed `display time equals canonical recorded fact, not browser arrival`;
replacing duration with zero failed `display duration equals recorded observation`.
Both Actions hosts stopped cleanly. Logs and JSON reports are under
`/private/tmp/cadence-actions-mutants.pp1uvZ`, with suffixes timing-now and
duration-zero. The full three-file diff was reviewed in job172.

These checks cover real measured observations and ordinary replay. The conditional
unknown-duration assertion is not evidence that this fixture exercised an unknown
value. Explicit legacy unknown values and snapshot-only restoration remain open,
as do the other unfinished rubric cases. All rows remain UNGRADED and no runtime
or installed application changed.

### Public snapshot-only timing probe

A separate fifth probe, timing-core, constructs grader-owned historical results
with a fixed offset timestamp and tick-precise duration, an explicit zero, and
missing timing. It exercises the Event codec, writes a real SaveDocument, calls
History.LoadSave, checks the restored Context, then projects with an empty event
tail. Known values and offsets must survive; measured zero and unknown remain
distinct. The runner requires all ten named receipts; unit tests reject each
omission and incomplete contract. No model request is made.

Job179 passes both Go packages with all five external probes. Full helper and
runner diff review completed in job181. The isolated dropped-timing projection
candidate built successfully; codec and loader checks passed before it failed
`timing-known-snapshot-only-projection`. The invented-defaults candidate also built
successfully and failed exactly `timing-unknown-snapshot-only-projection`, after
known and zero projections passed. Job180 completed both negative runs. Saved
browser failure reports retain STOPPED with no joined-shutdown error. Evidence
remains under `/private/tmp/cadence-actions-mutants.pp1uvZ`.

This probe tests the public snapshot-only projection, not legacy unknown values
rendered in a browser. It does not complete the timing rubric or any other row;
all five rows remain UNGRADED. No runtime source, installed binaries, services or
publication outputs changed.

### Public legacy timing in the browser

The Actions host now loads an actual snapshot into a separate Agent before the
GUI mount. Its retained results distinguish missing timing from a known timestamp
and measured zero duration. The public browser selects that Agent, checks absent
unknown timing and exact known/zero display, then reloads and explicitly selects
it again. Selection and replay leave provider and callback counts unchanged.
The runner requires the named legacy-browser receipt; all rubric rows remain
ungraded with zero earned points.

Job 189 failed after reload because the fixture assumed peer selection survived
reload. Initial selection passed. The driver now explicitly selects the peer
again rather than imposing an unrelated selection-persistence contract. Job 191
passed both Go packages with all five external probes enabled. Full three-file
review and whitespace checks passed; no production code changed.

Job 194 built an isolated GUI that invents a timestamp and zero duration for
missing tool-result values. It failed specifically at `legacy browser must not
invent recorded time`; the other four probes passed and the Actions host stopped
cleanly. Evidence is `/private/tmp/cadence-legacy-timing-mutant.bf6XpA`.
This combined mutant stops at the timestamp assertion, so it is not separate
browser sensitivity evidence for the missing-duration assertion. The earlier
projection-level unknown-duration check remains distinct evidence.

### Public provider-delivered snippet copying

The buffered Actions provider now delivers real Markdown through the model,
History and GUI path. The public driver uses visible controls and replaces only
the browser clipboard boundary, without reading or writing the system clipboard.
Independent expected strings check Unicode, indentation, tabs, trailing spaces,
blank lines, terminal LF, CRLF normalization in quoted code, list nesting, literal
hostile markup, an empty fence and a long collapsed preview. An unfinished fence
has no Copy control. The fenced human prompt must be present and have no
assistant-only Copy control. Rendering performs no clipboard write. Deliberate
clicks copy full payloads; permission denial exposes the exact read-only selected
manual fallback. Copy leaves the prompt empty and provider/tool counts unchanged.
The runner requires the new named receipt. No application globals or injected
protocol frames are used.

Jobs 202 and final 206 passed both Go packages with all five external probes.
Review added positive human-card presence and collapsed-preview assertions, so
absence or an expanded fixture cannot accidentally satisfy those checks. Job 205
built two isolated embedded-GUI mutants. Preview-only copying failed precisely
at the long payload assertion; unfinished-fence copying failed the expected six
controls assertion with seven. Both hosts stopped cleanly and the other probes
passed. Evidence is `/private/tmp/cadence-public-snippet-mutants.ZVoAs4`.
The negative driver predates only the two added presence/expansion assertions;
its decisive payload and closure assertions are unchanged.

All rows remain ungraded. These buffered checks do not certify streaming focus
or selection, missing-clipboard-API behavior, delayed completion replacement,
native clipboard permissions, or human usability. No runtime source, installed
application, settings, credentials or historical endpoints changed.

### Public streaming snippet continuity

A separate Agent now receives a real Anthropic SSE response from the grader.
Three external provider barriers make the browser observe an unfinished fence,
its closure, and later prose updates without injecting renderer frames. Before
closure there is no Copy control. After closure, copying returns the exact
payload; an expanded snippet, its original focused Copy button and its success
feedback survive the next delta. Removing the clipboard API gives an exact
read-only manual payload whose actual node, focus and complete selection survive
the final delta. The receipt requires one streaming request and completion,
no streaming oracle failure, and the unchanged buffered Actions counts. Teardown
releases all provider barriers before joining the host.

Job 212 passed both Go packages and all five external probes. Job 215 reviewed
the full four-file diff and whitespace checks passed. An initially broad edit
also changed the download receipt validator; inspection caught and corrected it
before tests, leaving that contract unchanged. Job 214 built an isolated GUI
with unchanged-snippet reuse disabled. It failed specifically at the original
focused Copy node assertion, with the other probes passing and clean host stop.
Evidence is `/private/tmp/cadence-streaming-snippet-mutant.WKz3KN`. This negative
control fails before the manual-selection assertion, so separate selection-loss
sensitivity remains open. All rubric rows remain ungraded. Delayed clipboard
completion after replacement, full prose selection/scroll coverage and native
human acceptance remain separate gates. No production code or installed files
changed.

### Public delayed clipboard completion across replay

The independent Actions browser now holds the clipboard promise after a real
click, selects another Agent through the tree, then returns to the original
streamed answer. It proves the original snippet is detached and the replayed
snippet exists with fresh feedback before resolving or rejecting the promise.
Both outcomes preserve replacement feedback, create no replacement manual
fallback, leave obsolete feedback untouched, and issue exactly one write of the
click-time payload. Agent replay issues no additional provider request or tool
execution. Only the clipboard API boundary is substituted; there is no native
clipboard access or injected application frame. The runner requires the named
`delayed-clipboard-replay-fenced` receipt.

Job 220 passed both Go packages with all five external probes enabled (exit 0).
Job 222 built an isolated GUI with both detached-snippet completion guards
removed. It failed specifically at `detached clipboard completion does not update
obsolete feedback`: the obsolete status changed from Copying to Copied. Other
four probes passed and the Actions host stopped cleanly. Evidence is retained in
`/private/tmp/cadence-clipboard-mutant.Od2EX9`. The mutant stops at delayed success;
this is not separate denial-path sensitivity. The good runtime and installed
applications were unchanged. Full diff review and whitespace checks passed.

The lesson is that asynchronous clipboard completion belongs to the original
control, not the conversation currently displayed. Public rubric rows remain
ungraded. Separate manual-selection-loss sensitivity and remaining Actions and
interaction coverage are still open.

### Separate manual-selection sensitivity

Job 228 built an isolated GUI that retains snippet nodes and restores focus but
then collapses a focused textarea selection to position zero. The public Actions
probe passed its earlier Copy-button focus check and failed specifically at
`real streaming preserves actual manual selection and focus`. The other four
probes passed and the Actions host stopped cleanly. Build and JSON receipts are
in `/private/tmp/cadence-selection-mutant.EZEM6H`. The unchanged good candidate
passed this same assertion in job 220; no repeat broad run was needed for this
negative-only evidence. No runtime or grader source changed for this experiment.

Retaining focus does not establish selection preservation. This separate failure
closes the sensitivity limit recorded for the earlier no-reuse mutant, which
stopped at button focus before reaching the manual textarea assertion. Rubric
rows remain ungraded pending the explicit coverage review.
