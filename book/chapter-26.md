# Chapter 26: Artifacts and Actions

> **Implemented integration; verification limits retained.** The named public
> artifact grader verifies70/100: Actions25, snippets20, recorded time15 and
> interaction10. Work products30 remains UNGRADED. Broader implementation tests
> and explanatory teaching are recorded, but do not establish full independent
> certification or human usability acceptance. Further student/grader campaigns
> are deferred for closeout. Historical `-ch 26` still selects the multimedia
> grader, not this chapter.

## TL;DR: Specification Brief

Make the agent's work inspectable without creating another owner of that work.
The Actions pane projects recorded calls and observations. Code controls copy the
complete logical payload. Work-product controls offer a deliberate download from
an explicitly authorized, still-live Agent.

- Pair calls and results by recorded identity. Show Requested, Dispatched,
  Result returned and Error reported as distinct facts. A returned observation
  can describe a Job that was still running.
- Preserve raw details and the reader's position. Updates retain expanded
  controls, focus and selected text; explicit navigation may move the reader.
- Copy explicitly closed assistant code fences, including their complete hidden
  payload, with parser-normalized LF line endings. Clipboard refusal exposes a
  read-only manual fallback beside the control.
- Display recorded fact times locally, with exact timestamp detail. Display a
  fixed monotonic dispatch-to-observation duration separately from Job state.
  Missing historical values remain unknown through save, restart and replay.
- Offer workspace files only through the explicitly granted `offer_work_product`
  tool. A successful result carries an opaque identifier, basename and offered
  size. The human's download click reads current contents through the owning
  Agent's Sandbox again.
- Bound each work-product transfer to 64 MiB and each Agent lifetime to 64 offers.
  Refuse extra offers rather than evicting earlier ones. Agent destruction or
  restart expires the live grants; saved descriptions remain descriptions.
- Keep cancellation and ownership explicit. A transfer belongs to its Agent;
  shutdown cancels and joins it before destroying the Sandbox. Browser replay
  cancels obsolete requests and refuses stale completions.

## In plain words

An agent can say that it wrote a report. That sentence does not establish that
there is a file, that the browser may read it, or that a download contains all of
it. Likewise, a tool result can arrive while a background process is still
running. A useful interface tells these situations apart.

This chapter connects the owners built earlier. History remembers what happened.
Jobs own work that can outlive an observation. The Sandbox decides which files an
Agent may open. The browser presents those facts and the human's controls. None
of these responsibilities moves into a new artifact store.

Junien's [snippet-copy contribution](https://github.com/waywardgeek/ensemble/pull/7),
[timing contribution](https://github.com/waywardgeek/ensemble/pull/6) and
[work-product contribution](https://github.com/waywardgeek/ensemble/pull/9) supplied
requirements, tests and teaching that informed this development. Their Go code is
not the current C# implementation. We retain the useful distinctions while using
our existing C# owners and the bounded timing scope developed here.

## 26.1 Start with the existing owners

Read the construction points before joining them:

| Foundation | Responsibility used here |
| --- | --- |
| [Chapter 2](chapter-02.md) | Immutable recorded values and retained Context |
| [Chapter 3](chapter-03.md#preserve-typed-result-metadata) | Typed tool results rather than metadata parsed from prose |
| [Chapter 4](chapter-04.md#completed-structured-results-carry-delivery-metadata) | Jobs and observations, including completed structured results |
| [Chapter 7](chapter-07.md) | Browser projection, Markdown and stable controls |
| [Chapter 9](chapter-09.md#installed-discoverable-and-granted) | Explicit skill authority |
| [Chapter 10](chapter-10.md) | Save, restore and replay |
| [Chapter 18](chapter-18.md) | Agent-owned file authority |

Common declares the shared values and real ownership relationships. Matching
partial implementations perform their operations. The browser's card map is a
presentation index, not another History. A saved `WorkProductInfo` is an immutable
value, not a second collection of live grants.

Follow one successful offer through the system: record the call, execute the
admitted tool under its Agent, return typed metadata, record the observation,
project the result and wait for a human click. The next provider request receives
ordinary result content, not browser download authority. Replaying the recorded
result renders a description without executing the tool again.

## 26.2 A result is an observation

Consider a batch with two calls. The first has been dispatched and is waiting for
its foreground observation. The second is recorded but has not been dispatched.
Both exist, but only the first has crossed the execution boundary. Showing both
as running would claim an effect that has not happened.

A later result can say that Job 12 was running when observed. Keep that recorded
state even after another observation says Job 12 finished. The old card is a fact
about an earlier observation, not a live status widget. Display an available exit
code separately; a missing code is not zero, and a finished Job is not by itself
proof of successful work.

Use the projected call relationship for View call and View latest result. Tool
names, matching output text and Job handles are insufficient pairing keys. A call
can have more than one observation. Navigation must validate that both card
objects still belong to the current Agent and current rendered transcript. An old
button detached by replay must not target a replacement that reuses its ID.

Control tools expose a subtle integration rule. `wait_for_job` creates no wrapper
Job, but its result can describe an existing Job. Preserve that typed observation
when recording History. Testing only the returned method value would miss a
recording path that drops the metadata before the GUI sees it.

## 26.3 Preserve the thing the reader is using

Keep stable controls separate from changing payload presentation. Updating a
status label should not replace an unchanged detail summary, download button or
snippet control. Preserve actual nodes where their meaning is unchanged; a new
node that looks identical can still lose keyboard focus or textarea selection.

A reader may select raw output while another result arrives. Defer changes that
would replace selected content, and retain the pane position while the reader is
interacting. Explicit pair navigation may focus and scroll to its destination.
Ordinary output should not do that on the reader's behalf.

Exercise these distinctions on a narrow, genuinely overflowing pane. An empty
selection or a pane with no overflow cannot demonstrate preservation. Check
keyboard activation and touch-sized targets as well as mouse clicks. Copying and
navigation preserve an unsent draft and do not queue duplicate speech. Streamed
narration advances independently of visual rerendering; replay is silent rather
than a request to speak the conversation again.

## 26.4 Copy the payload, not the preview

The [Chapter 7 copy contract](chapter-07.md#copy-the-logical-code-not-the-preview)
uses the pinned Markdown parser's logical fence payload. It excludes opening and
closing fences and quote or list container prefixes. It preserves code indentation,
Unicode, blank lines, trailing spaces and the terminal newline, with the parser's
LF normalization. History retains the original Markdown.

A collapsed preview is presentation, not the copy source. Test a payload that
exceeds both the line and character preview limits. Explicit closure matters:
reaching the end of an assistant response does not close an unfinished fence.
Human messages and inline code do not acquire assistant-fence Copy controls.

Automatic copy succeeds only when the clipboard promise succeeds. If the API is
missing or denies access, show the exact payload in a read-only, selected fallback
beside the button. Preserve that fallback's actual focus and selection when later
prose streams in.

Capture the click-time payload and control identity. If the human switches Agents
before the clipboard promise completes, its resolution must not update the new
card or create a fallback there. The already-requested clipboard write cannot be
retracted merely by replacing the DOM. Fence the completion rather than claiming
navigation canceled a write that the browser may already have performed.

## 26.5 Keep recorded time distinct from elapsed work

A fact's recorded timestamp belongs to History. Retained entries preserve it
through filtering and persistence; snapshot-only projection must not substitute
the time of reload. Show local time for reading and exact timestamp detail for
inspection. Historical absence remains absence.

Observation duration has a different source. Measure admitted dispatch to the
returned foreground observation with a monotonic clock, then record the fixed
value. A zero-delay observation may return while its Job remains blocked. The
recorded duration stops there; it is not the eventual background runtime. Paths
without the defined dispatch boundary remain unknown rather than receiving an
invented zero.

Retain exact values through the application's real save operation and a separate
restore process. A hand-authored snapshot tests loading but cannot detect a save
operation that drops timing. Also compare provider requests with and without
these inspection fields: timing should not change the model's input. The public
grader exercises four renderer surfaces offline and an actual fake Anthropic
wire path; that is not four live-provider integrations or a full timing profiler.

## 26.6 A work-product description is not a file grant

An ordinary pathname in prose grants nothing. Job-output downloads serve completed
Job diagnostics under their own 16 MiB boundary. Work-product offers are separate:
they serve a deliberately offered workspace file under the 64 MiB policy. Neither
raises model image-input limits.

Grant `offer_work_product` through the work-products skill. Installation or
historical skill text alone grants no tool authority. The Agent validates a
regular readable file through its Sandbox, checks the offer size and capacity,
and creates an opaque identifier. Failed creation consumes no offer slot. The
successful typed result contains display metadata without the host pathname.

The live Agent owns the mapping from that identifier to the permitted path. A
request names an Agent and offer ID, never an arbitrary pathname. Require the
existing origin boundary and the download-specific request header, then look up
the offer under that owner. A globally unique token is not permission to read it
through another Agent. Same-origin checks are not a replacement for human login
in a different hosting design.

At click time, open through the Sandbox again and check the current length. A
replacement symlink that escapes the Sandbox, missing file, oversized growth or
destroyed owner must refuse.
The download contains current bytes, not a frozen copy of the original offer.
The offered size is descriptive, not a promise that the file is unchanged.

Saving and restoring the description does not recreate the live mapping. After
restart the old card may explain what was offered, but downloading requires a
new live offer. Reusing an Agent name does not revive the destroyed Agent's grants.

## 26.7 Own the transfer through cleanup

An admitted transfer owns its opened input stream and cancellation lifetime under
the Agent. The HTTP destination is borrowed. Keep the transfer registered until
input disposal and other cleanup complete, then remove its reverse membership.
Do not hold the Agent's gate across network I/O or cancellation callbacks.

Bound actual transferred bytes, not just a declared length. Refuse truncation or
growth rather than save a misleading successful download. This is a bounded
click-time transfer, not an immutable snapshot guarantee against same-size edits.

Browser Cancel and replacement of the conversation abort the obsolete request.
The server must observe that cancellation and join cleanup. Agent destruction
closes admission, cancels active copies and joins them before Sandbox teardown.
A test that sees a changed button label proves none of those server effects.
Use a controlled external response sink to demonstrate cancellation reaching the
copy and shutdown remaining pending until cleanup is released.

Only save the browser Blob after a complete allowed response and a final check
that the original card and Agent are still current. Render refusal visibly;
never convert an error response or partial transfer into an attachment.

## 26.8 Integration exercises

These are draft exercise contracts. The isolated pre-chapter reconstruction
baseline still needs to be assembled and audited. Do not hand a student the completed
runtime as a purported prerequisite or claim these exercises are student-verified.
Students must have the relevant foundation chapters, existing offline dependencies,
an explicit writable solution scope and access to the public grader and questions.

### Exercise A: Follow a call through observation and replay

Use a loopback fake provider to request two real tools. Hold the first callback
while the browser inspects the batch. Show the first call dispatched, the second
requested and no returned result yet. Release the callback and verify distinct
results, exact raw inputs and navigation in both directions. Add a real failure
and a running Job observation. Reload and repeat navigation without rerunning
callbacks or changing earlier observed states.

Record canonical counts as well as visible assertions. Explain which component
owns execution, which owns History and why a control result can carry Job metadata.

### Exercise B: Read while the answer changes

Return closed, unfinished, nested, empty and long fences through the actual
provider path. Compare clipboard writes with independently constructed expected
payloads. Refuse the clipboard API and verify the manual fallback. Hold real
stream deltas while the reader focuses a Copy button or selects fallback text;
release later prose and verify the same controls and selection survive.

Repeat while an unsent draft exists. Delay clipboard completion across Agent
selection. Check keyboard and touch controls, raw text selection, pane position
and speech noninterference. Use synthetic clipboard and speech boundaries; keep
native device acceptance a separate report.

### Exercise C: Restore facts without inventing measurements

Construct known timestamps and durations, an explicit zero and unknown values.
Save through public History, exit, then load in a separate process with no seed
fallback. Project from retained Context without a diagnostic event tail. Verify
exact values and browser absence for unknowns. Compare otherwise identical timed
and untimed provider requests while positively checking that tool payloads remain.

### Exercise D: Offer, change, download and revoke

Have a real tool create a binary with independently known random bytes. Offer it
through the granted tool, then change it before the human clicks. Verify the exact
new bytes and filename. Test failed and ungranted offers, wrong-owner and path
requests, symlink replacement, missing files, exact size boundaries, cancellation,
Agent destruction and restored descriptions without restored authority.

Explain the difference between an offer, an observation, a transfer and a saved
description. Include successful allowed operations beside refusals so a candidate
that refuses everything cannot pass. The deterministic running built-in offer
case remains an open public-grader gate; a delayed custom text callback does not
by itself verify that structured-result path.

## 26.9 Run and interpret the public grader

Build the candidate and grader using the repository's documented offline setup.
Supply absolute paths to existing tools and a candidate runtime directory containing
`Ensemble.dll` and `Ensemble.Gui.dll`:

```sh
/path/to/grade-edition3 --scope artifacts -executable-only -json \
  -dotnet /absolute/path/to/dotnet \
  -browser-node /absolute/path/to/node \
  -playwright-module /absolute/path/to/playwright \
  /absolute/path/to/candidate-runtime
```

The grader compiles its own outside-assembly consumer against the supplied
assemblies. It uses isolated workspaces, loopback providers, real browser controls
and joined hosts. It neither runs candidate internal tests for points nor needs
real provider credentials. Existing .NET 10, Node, Playwright and Chrome are
prerequisites; running the command is not permission to install them. The current
runner supports macOS and Linux hosts.

Exit 0 is reserved for a full pass, 1 reports a probe failure, 2 reports setup or
argument failure, and 3 reports unfinished coverage. A successful current run
reports 70/100 with work products UNGRADED. A failure can retain independently
supported row points while still exiting 1. Read row status and failure evidence,
not just the total.

Deliberately broken candidates must compile and fail the intended assertion.
Compilation errors and stranded test hosts do not demonstrate sensitivity.
Keep source implementation, public grading, reconstruction and human acceptance
as separate records. The [development evidence](../docs/edition-3-notes/artifacts-and-actions.md)
contains the individual runs, corrections and limits; the
[closeout plan](../docs/edition-3-notes/first-book-closeout-plan.md) tracks remaining gates.

## 26.10 Use it for real work

Ask the agent for a small report. Read a previous tool result while new output
arrives, copy a long snippet while preserving a draft, and deliberately download
the offered report. Check the error message after the file becomes unavailable.
Do this on the devices you actually use. Synthetic success does not establish
that a person can find a control, understand its state or comfortably read it.

This chapter ends with inspectable work, not a permanent content service, arbitrary
HTML previews or a comprehensive profiler. The next chapter adds local inference
after the non-local teaching and reconstruction gates are complete.

Next: [Running Local Models](chapter-27.md).
