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
initial contract receipt awards no chapter points; HTTP/browser and the other
rubric rows still need independent drivers and sensitivity checks. The named
CLI scope remains unavailable until its driver and argument contract are tested.

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

The snippet, recorded-state and pair-navigation increments above are verified
within their stated presentation scopes. Next are recorded timing semantics and
explicit work-product offers.

Later slices cover paired state presentation, agreed timestamps/durations and
explicit work-product offers. Public grader points, broader sensitivity checks, substantive
chapter teaching, fresh student reconstruction and Bill's usability acceptance
remain pending. Students can ask Cadence questions and run public graders under
the [closeout plan](first-book-closeout-plan.md).

Foundation lint follow-up: Chapters3/4/10 retain exactly the pre-existing em-dash
lines compared with HEAD; Chapter18 passes hard checks. Advisories remain. The
printed lint exit1 was retained even though a subsequent log-print command exited0;
this is not recorded as a full manuscript lint pass.
