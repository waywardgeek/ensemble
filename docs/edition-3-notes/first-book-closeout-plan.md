# First-book closeout: Artifacts, Actions, and Local Models

Status: Bill approved closing the adopted PR-integration work and proceeding
personally to Local Models. Further student campaigns and expanded artifact/GUI
grading are deferred; their unverified coverage remains unverified. The existing
artifact score remains70/100, not a newly awarded full pass. No more students
without a new explicit request. Finish Local Models on the agreed real target,
then perform a bounded manuscript/publication check and human usability acceptance.
The earlier detailed student and exhaustive grading gates below are retained as
historical planning, not mandatory steps or current launch authority.

Prepared by Cadence after reviewing all nine open GitHub PRs from madbrain76
(Junien), their implementation and chapter changes, and selected substantive
tests. No PR was executed, merged, commented on, or closed during that review.
This plan authorizes neither deployment nor model/runtime installation.

## 1. Purpose and working method

Finish The Art of Building AI Coding Agents before beginning Puffin development.
There are two new chapter subjects, not nine independent feature obligations.
Build bounded slices with Bill, observe behavior, correct the design, and then
write the executable teaching that the implementation has earned. This is
Development-Driven Specs. The detailed contracts below are proposed starting
points; observations can change them, with the reasons recorded visibly.

For remaining development, Cadence inspects, implements, tests and writes the
teaching personally. Use focused behavioral checks and broader regressions only
where a shared contract changes; do not expand the acceptance program merely to
fill an uncertified row. Preserve completed evidence and record actual limits.
Scoped local commits are permitted. Do not push, replace installed applications,
restart services, download models, install dependencies, or run paid providers
without separate authority.

Protect unrelated work, existing Measure X ownership, private credentials, and
live services on 8088/8089/8090. Use isolated homes, ports and browser profiles.
Keep at least 1 GiB free before heavy work. Do not rerun broad suites simply for
freshness; run them when a shared contract changes and after integration.

## 2. Scope and contribution attribution

Primary inputs from https://github.com/waywardgeek/ensemble/pulls:

- PR 9: explicitly offered workspace work products. Adopt the use case and
  security/replay tests, not the old Go implementation wholesale.
- PR 7: exact snippet copying, complete-block safety, focus preservation and
  clipboard fallback. Extend our pinned Markdown parser rather than replace it.
- PR 6: recorded timestamps and honest timing vocabulary. Separate modest useful
  timing from an always-on detailed profiler. Its explicit-streaming proposal
  belongs in Local Models, not in the timestamp implementation.
- PR 3: sleeping local inference machines. Consider Wake-on-LAN only after the
  basic local-provider path works; it is not required for first acceptance.
- PR 5: avoid expensive whole-context copies to answer narrow questions. Current
  C# History.View returns immutable Context rather than JSON-deep-copying it.
  Profile actual remaining costs before introducing caches or new read APIs.
- PR 1: system theme is already implemented and taught, with browser-local
  ownership. No port or new exercise required.
- PR 2: display-label renaming remains a useful separate enhancement. It is not
  part of this closeout; if later adopted, do not rename routing/storage IDs.
- PRs 4 and 8: remote-container access and GUI authentication deserve separate
  hosting/security decisions. No direct LAN listener or Digest implementation
  is included here. Current provider-account OAuth is not human GUI login, and
  origin validation is not authentication.

Credit Junien's specific use cases and contributions in the two new chapters
and implementation evidence. His benchmark/test reports remain attributed,
not relabeled as our measurements. All nine PRs include chapter edits.

Explicit non-goals: permanent content-addressed artifact storage, arbitrary
HTML/application previews, untrusted renderer plugins, a general filesystem
browser, a new scheduler, Puffin/Watchtower, automatic provider fallback,
whole-profiler instrumentation, and retroactive grading of historical endpoints.

## 3. Manuscript placement and navigation

Proposed final order:

- Chapters 00–25 retain their subjects and filenames.
- Chapter 26: Artifacts and Actions.
- Chapter 27: Running Local Models.
- Chapter 28: Crossing Over to Ensemble (moved from current chapter-26.md).

Before moving files, build an inventory of current links to Crossing Over and
rewrite by resolved target, not a global replacement of the number 26. Preserve
solutions/chNN, edition-1, edition-2, historical evidence and grader flags.
Create clearly labeled in-development chapter files, not completed-looking prose.
Do not copy outlines over Crossing Over. Update book/README.md, book/chapters.md,
root navigation, relevant companion links, and procedure/progress headers.

Repair the current Voice and Images internal 26.x labels to 25.x, including
current exercise labels and local anchors. Keep the documented external media
command -ch 26 and all historical score records unchanged. Check other moved
chapters for the same internal-heading defect without rewriting historical text.

The assembler discovers chapter-NN.md files. Verify the resulting 29 chapters
00–28 are unique, ordered and end in Crossing Over. Publication outputs remain
explicitly stale until the publication phase; do not claim a rename rebuilt PDFs.

## 4. Foundational chapter edits

Each enduring contract has one authoritative teaching home; later chapters link
back instead of duplicating complete specifications.

### Chapter 2: History and provider boundaries

Clarify immutable Context views versus copying the event collection in a full
History snapshot. Teach any new persisted artifact metadata at its declaration
and reducer/serialization construction points. Display metadata must never leak
into provider requests. Missing old metadata is unknown, not fabricated.
Introduce only the shared vocabulary actually required by the finished slices.

### Chapters 3 and 4: tools and Jobs

Teach a typed work-product offer when that contract is implemented. A tool call,
a successful execution, an observation timeout and a running Job are distinct.
A running report is not a successful offer; delayed success must project once.
Preserve call/result identity and original Job ownership across waits. File
transfer and tool observation are different operations with different lifetimes.

### Chapters 5 and 6: configuration and streaming

Add the explicit streaming policy at the real configuration construction point,
then teach its use by both request rendering and transport. Buffered and streaming
responses must retain the same finalized content and tool semantics. Unknown
model names do not imply unsupported text, free inference, or guessed media
capabilities. Keep explicit billing/account selection independent of endpoint.

### Chapters 7 and 8: artifact GUI foundations

Teach stable identity, canonical retained payload versus bounded display preview,
safe Markdown, non-color state labels, exact copy semantics, and stable reading
position. Retain the current browser-profile preferences and speech behavior.
Chapter 26 adds integrated Actions behavior without replacing these foundations.

### Chapter 10: persistence

Update the versioned save/journal and restore contracts for metadata we actually
add. Old records without timing/offers remain readable without invented data.
Malformed supplied metadata refuses appropriately; replay cannot execute tools,
wake machines, transfer files or regenerate external effects.

### Chapters 15 and 18: accounting and Sandbox

Local cost/prices remain unknown unless measured/configured, never silently zero.
Token usage reported by a server is distinct from estimates and missing usage.
Workspace downloads reuse the one Agent Sandbox. Teach the actual handle-based
file restrictions at Sandbox construction, with no independent GUI path checker
or metadata-check/open race masquerading as confinement.

### Chapter 25 and the new Chapter 28

Preserve image/audio and attachment contracts while integrating new controls.
Crossover gains the final acceptance checklist for Actions and local-model use;
it does not become a release-note appendix. Existing acceptance gaps remain
explicit unless this work actually exercises and closes them.

## 5. Artifact implementation slices

### A0. Establish the baseline and settle the visible design

Inspect agent/gui/wwwroot/app.js (card state, render, fullOutputLinks),
markdown.js, style.css, Server/Projection.cs, Server/JobOutput.cs,
Common/Protocol.cs, and existing browser tests. Record what already works:
paired calls/results, bounded previews, complete retained detail, Job-output
read/download, Markdown and media attachments. Do not rebuild those mechanisms.

Use a deterministic fixture showing a command, edit, search, running Job,
failed tool, and binary report. Agree with Bill on compact default summaries and
expanded detail before polishing the implementation. The default pane must help
answer: what was requested, what is happening, what finished, and what failed?
Raw arguments/results remain reachable. No inferred completion from a filename,
spinner, quiet output, elapsed time or disappearance of a Job.

### A1. Exact snippet copying

Modify markdown.js and the existing card integration; add a small browser-only
helper only if it makes ownership/readability clearer. Do not add a second parser.

Proposed behavior: Copy is available for complete fenced code in assistant text.
It copies the payload only, not prose, fences, syntax-highlight markup or a
truncated preview. Define terminal-newline handling explicitly from parser/source
boundaries and test it; do not let incidental DOM normalization decide. Preserve
Unicode, blank lines, indentation and longer nested fences. An unfinished streamed
block stays noninteractive until complete. Existing speech reads original text.

Use the clipboard API on explicit user click. If unavailable/denied, offer a
manual selection fallback with an honest status. A legacy copy-command fallback
is optional, not required for correctness. Never read the user's clipboard in
production. Do not execute or submit copied text. Preserve selection/focus and
open detail state while later prose streams. On conversation switch, an admitted
copy owns the click-time bytes and cannot copy a replacement card's contents.

### A2. Actions presentation and stable state

Improve summaries over existing card state rather than create another History.
Show queued/running/completed/failed/interrupted or observation-ended labels only
where supported by actual facts. Keep call/result associations and Agent/Job
identity explicit. A failed observation must not be mislabeled a failed execution.
Keep raw complete received payload separate from the full retained Job output.
Use accessible text, keyboard focus and adequate touch targets, not color alone.

Preserve reader scroll/selection while updates arrive; follow the bottom only
when the reader was already following. Expansion, copy and download controls
must not trigger speech accidentally or be intercepted by the heading target.
Check narrow phone widths, long commands and selected text during streaming.

### A3. Recorded timestamps and bounded timing

First implement recorded timestamps on meaningful human/response/tool facts.
Store UTC; render local time with an accessible full date/time, not tooltip-only
information. Replay uses recorded time, never browser arrival time.

Current Server/Projection.DisplayFacts recreates dialogue facts from Context
entries; these entries currently lack original timestamps. Inspect retention and
snapshot requirements before selecting the smallest immutable metadata change.
Ensure timestamps survive snapshot-only restore as well as journal replay.
Do not create an authoritative side table keyed by copied IDs.

Use a monotonic clock for elapsed durations, UTC for dates. Proposed initial
metrics are total admitted-turn duration and foreground tool-observation duration,
with labels that distinguish continued background execution. Exclude granular
model/tool/other phase accounting initially. Decide with Bill whether per-part
first-to-last observed fragment span adds enough value; if included, label it
as observed delivery span, not model compute or time-to-first-token. Missing or
single-fragment measurements are unknown, not an invented duration.

Define start/end boundaries before implementation, including queueing, pause,
interruption and final persistence. Record once at the chosen completion boundary;
late workers cannot revise a closed turn. Do not call a metric user-visible total
latency if it excludes persistence or final delivery. Tests enforce the written
boundary rather than a name borrowed from PR 6.

### A4. Explicit work-product offers

Start with one report generated in the owned test workspace and offered through
the real Agent/tool/History/GUI path. Suggested tool name: offer_download.
Introduce a typed, immutable offer result with provenance, rather than parsing
success prose or trusting a model-supplied HTTP path. Tool authority comes from
normal skill grants. Add it to an appropriate explicit skill, not every Agent
or child automatically. Public declaration, result encoding and replay must
be specified together before external grading.

Proposed minimum semantics for Bill's approval:

- An explicit successful offer creates a Download control; ordinary path text
  does not. A user click is required before transfer. Asking the model to offer
  only on request is guidance, not a claim of deterministic consent enforcement.
- Offer identifies a workspace-relative regular file through its real Agent.
  Even an unconfined Agent may offer only within its workspace for this feature.
  Existing Sandbox policy may narrow that further.
- Start with click-time contents, clearly labeled, not a frozen copy or permanent
  publication. Replacement inside the permitted workspace can change bytes.
  If immutable delivery is required, stop and design that separately rather than
  quietly introduce a content store.
- Live GUI mount owns any transient random download capabilities. No arbitrary
  HTTP filesystem path; no tokens in diagnostic logs or provider requests.
  Apply the existing mount/origin/access boundary to downloads. Possession of a
  link is not represented as a new person identity or universal authentication.
- Offer, file, Agent, capability and mount lifetimes are distinct. Revalidate
  live Agent ownership and file access at transfer. Deleted/recreated Agent IDs
  must not revive old capabilities. Retained offers may rebuild new links after
  restart; old links stop working. Replay never reads file bytes automatically.
- Bound active offers and transfer resources, join transfers on mount shutdown,
  cancel on disconnect/abort, and specify expiry/revocation before release.
  Suggested starting policy: links live for the current mount and are removed
  with their Agent; select a concrete offer-count/size limit after fixture tests.
  No unbounded map is an acceptable final implementation.
- Stream binary bytes with attachment disposition, no-store and nosniff; render
  filenames as text. Do not load arbitrary-sized files into browser memory.
  Decide size limit and Range/HEAD support explicitly; test accepted behavior.

Implementation homes: Common declarations under agent/Common; functional tool
implementation under agent/Tools; regular-file/workspace confinement under
agent/Sandbox; serialization/replay under agent/History and existing codec
implementations; GUI projection/routes under agent/gui/Server with matching
Common declarations. Proposed new filenames are not claims existing APIs exist.
Keep one authoritative owner and no duplicate filesystem/security service.

## 6. Local-model implementation slices

### L0. Choose one real target and capture its capabilities

First target recommendation: Ollama, through its OpenAI-compatible Chat
Completions endpoint, using an explicitly selected installed tool-capable model.
This is a compatibility surface, not a claim that Ollama is OpenAI or that every
OpenAI extension works. Ask Bill for the machine, endpoint, installed model tag,
context capacity and permission to make local inference requests. Inspect those
facts before prescribing hardware or downloading anything. Record exact runtime
and model versions. No network discovery/port scanning.

Begin with buffered text, then streaming text, then a real harmless file tool
roundtrip. Inspect actual wire behavior, including tool arguments, finish reasons,
usage and streamed call IDs. A server's model listing is not proof of tool quality.
Do not add a native Ollama protocol adapter unless the compatible path proves
insufficient for the agreed use case. No speculative second backend.

### L1. One explicit endpoint and capability policy

Use existing Configuration/Engine/Agent ownership. Preserve explicit BaseUrl,
model, Surface and Billing. Local configuration must not load unrelated cloud
credentials or fall back to a cloud provider. Missing local credentials can be
valid; absence must not trigger credential discovery. Fail unavailable endpoints
honestly and keep existing caller cancellation.

Propose a tri-state streaming choice: model default, explicitly enabled,
explicitly disabled. Reconcile existing DisableStreaming compatibility rather
than maintain contradictory mutable flags. One effective-policy calculation must
feed request rendering, endpoint selection and response handling, including
auxiliary recall requests. Check Configuration.WithModel and persistence paths;
configuration changes apply at the existing admitted-request boundary.

Do not infer media support from a name prefix or from successful text. Do not
hardcode the user's local model into the cloud catalog. Unsupported extra fields
(stream_options, reasoning controls, usage expectations) must be resolved from
actual target behavior, not dropped silently after arbitrary retries. A failed
request never becomes permission to switch endpoint, model, billing or protocol.
Model choice, declared context budget and capabilities are explicit and validated.
No claim of reliable overflow prevention from a guessed token-per-byte ratio.

Likely touched code: agent/Common/Values.cs and Media.cs; the actual Configuration
implementation; Engine/{Engine,Render,Streaming,RecallJudge,Media,Persistence}.cs;
CLI and GUI launch/configuration paths; model-choice presentation and tests.
Inspect constructors, copy operations and restore compatibility before edits.
Do not use a GUI settings change to disguise an application restart requirement.

### L2. Bounded real workflow and operational guidance

Exercise a harmless coding task in an isolated workspace with a real local
model: read a small fixture, propose/make a bounded edit, run a local check, and
report the actual result. Compare streaming/buffered output and cancellation.
Record latency and any available usage, but no invented energy cost or universal
quality ranking. Test how the chosen model behaves with invalid tool arguments
and a modest conversation; separate model mistakes from runtime protocol bugs.

Teach setup, exact launch recipe, model/context constraints, memory requirements
observed on that target, refusal/error diagnosis and cleanup. Explain that local
inference does not make web/MCP/tools private or remove prompt-injection risks.
No service startup on boot or deployment supervisor is included.

### L3. Optional Wake-on-LAN, only if needed and approved

If Bill's target sleeps and waking it is wanted, add an explicit per-endpoint
wake policy. Specify approved MAC, local IPv4 destination/broadcast, timeout and
cancellation. Do not guess that every unreachable host is asleep. Avoid a new
background wake service. Decide whether concurrent callers share one wake attempt
only if the observed workload requires it; do not prebuild a wake scheduler.

A packet sent is not a machine awake; TCP open is not a model ready. Validate the
broadcast's local-subnet scope. Use a synthetic UDP receiver and delayed local
HTTP server for routine tests; actual waking requires separate machine authority.
Do not attach wake traffic to replay, offline rendering, model listing or restore.
Keep this optional section ungraded or separately scored until independently
verified, without blocking the core local-model chapter.

## 7. Exercises and public graders

Do not reuse numeric -ch 26: it already names the retained media grader. Add
named scopes to cmd/grade-edition3, proposed --scope artifacts and --scope
local-models, mutually exclusive with explicit --ch. The artifact scope now exposes partial
public probes with reviewed Actions, snippet, timing and interaction scoring (70 points)
and an explicit UNGRADED work-product row. A successful partial run remains exit3, not full chapter acceptance.
The local-model scope and remaining flags below are still planned. Preserve all existing invocations.
Implement the new drivers in internal/gradeedition3 with descriptive filenames
(artifacts*.go and local_models*.go), reusing existing browser/process helpers
where their contracts actually match. Extend CLI validation and driver tests.

Public fixture launch: explicit candidate executable, isolated data/workspace,
loopback fake-provider endpoint, model/surface, random GUI port and stop protocol.
Specify that contract in the exercise and validate flags before effects. Outside-
assembly consumers reference candidate assemblies only where a public library
contract needs examination. Do not reach through reflection/private members or
compile the grader against repository implementation internals.

### Artifacts: proposed 100-point allocation

1. Actions identity and state, 25: real successful/failed/delayed tool calls;
   exactly one paired presentation per call; observation versus execution labels;
   wrong-Agent/foreign-Job rejection; raw detail retained; reconnect consistency.
2. Exact snippets, 20: payload-only copying; Unicode/whitespace/nested fences;
   incomplete/preview refusal; hostile markup remains text; blocked clipboard
   fallback; focus and selection during streaming; no prompt/tool side effect.
3. Recorded time, 15: persisted timestamps through replay/snapshot/restart;
   legacy unknowns; explicitly defined elapsed boundaries; late completion cannot
   change finalized measurements; no timing fields in provider requests.
4. Work products, 30: a real tool-created binary with random known bytes;
   successful offer and exact download; no early offer on running/failed work;
   traversal/symlink escape refusal; changed/missing file; deletion/replacement;
   HTTP guarding, token redaction, transfer limits/cancellation and replay policy.
5. Interaction stability, 10: keyboard/touch controls; narrow viewport; expanded
   detail/scroll/selection preserved; no misleading color-only states; no speech
   duplication or interference with drafts. Human usability remains separate.

Each row has an external oracle: actual wire exchange, downloaded bytes, visible
DOM behavior, retained History, independent on-disk JSON, or joined process result.
Internal tests are supporting evidence, not a replacement for public points.

### Local Models: proposed 100-point deterministic allocation

1. Explicit routing/funding, 25: random unknown model at fake local endpoint;
   exact selected model/path/surface; hostile cloud environment cannot divert it;
   optional-key behavior; unavailable server has no fallback request.
2. Buffered/streaming equivalence, 25: real renderer/parser path with fragmented
   SSE, Unicode, tool deltas and final content; requested streaming policy agrees
   with transport; malformed/incomplete streams are not successful answers.
3. Tool conversation/replay, 25: fake model requests a real confined file action;
   unique call/result pairing, actual file effect, following answer, and retained
   conversation after restart; replay causes no fresh tool/network effects.
4. Cancellation/failure/lifecycle, 15: stalled connection, interrupt, invalid
   arguments, server error, clean next turn, joined owned work and no silent retry
   of effects. Clear distinction between canceling observation and execution.
5. Configuration/accounting compatibility, 10: policy roundtrip and old defaults;
   explicit budgets/capabilities; unsupported media refusal before request; absent
   usage/pricing remains unknown. Offline render/dump requires no server.

These points certify a protocol/runtime contract, not local-model intelligence.
A separate real-target acceptance report is mandatory for claiming the documented
Ollama/model combination works. It names versions, exact exercised capabilities,
failures and limits. No paid/native-network action is implicit in the grader.

### Grader sensitivity and failure reporting

For each major guarantee, build a scoped broken candidate and require the grader
to fail the intended behavior, not setup. Minimum mutants: wrong call pairing;
partial snippet copy; timestamp replaced with reload time; file route bypassing
workspace enforcement; failed offer accepted; local route silently redirected;
streaming final lost; duplicate tool execution; cancellation ignored. Add mutants
when a test oracle proves weak, not a campaign of unrelated changes.

Report setup failures, skipped environments, UNGRADED scopes and implementation
failures separately. Fixed synthetic inputs must not make a constant hardcoded
answer pass: randomize file names/content/model tags and use grader-owned values.
Always join children/process groups and preserve bounded failure evidence.

## 8. Test and evidence organization

Add focused core tests under agent/Tests for offer encoding/confinement, immutable
metadata, restart/replay and streaming policy. Add GUI tests under agent/GuiTests
for actual Actions/snippet/download pages, using the existing fixtures and real
server sockets. New filenames may be ArtifactOfferTests.cs, ArtifactTimingTests.cs,
LocalProviderTests.cs, artifact-actions.cjs, artifact-copy.cjs and work-products.cjs;
choose the actual split for readability after inspecting neighboring declarations.

Adjacent regression targets: Markdown and theme/profile behavior; media/image
admission and replay; Job output download/cancellation; pause/hints/interrupt;
History/persistence; public-origin refusal; streaming tool order; provider funding.
Run the shared C# suite after shared History/configuration changes. Preserve known
unrelated failures with their scope rather than weakening assertions.

Maintain an evidence note for each new chapter with candidate commit/hash,
commands, test counts, public grader JSON, sensitivity failures, real-model target
and human acceptance. Distinguish source, build, installed application, served
assets, browser refresh and Bill's observed behavior. No stale checkpoint results
are presented as new tests.

## 8A. Historical student reconstruction gate (deferred)

Superseded by the current status above. No new student launch or repeated
reconstruction is required for this closeout; preserve the recorded results.

Bill requested this gate when each chapter's substantive changes and upgrades
are complete. Navigation-only edits do not require a full student exercise.
Track every substantively changed chapter in a matrix: prerequisite baseline,
teaching revision, assigned exercises, public grader revision, student result,
clarifications, rerun evidence and remaining limits. A new capstone pass does not
automatically certify its revised prerequisite chapters. The
[reconstruction matrix](artifact-reconstruction-matrix.md) records the changed
chapters, candidate starting points and unprepared baseline/acceptance gates.

Cadence owns implementation and teaching. Launch one fresh student at a time,
with an explicitly assigned isolated worktree, writable scope, resource bounds,
existing offline dependencies and allowed commands. The student must load the
coding discipline, inspect the assigned chapter and implement its exercises.
No descendants, live services, credentials, installations, paid-provider tests,
real screen captures, laptop/network changes, or edits to the main worktree.
Preserve the independently owned Measure X child.

Prepare a genuine pre-chapter baseline containing only declared prerequisites.
Frozen endpoints are references, not automatically sufficient baselines. Inspect
the actual dependency graph; do not give the student completed HEAD with the
answer already present, a partial deletion that leaks the solution, or later
chapters that silently supply missing requirements. Export the allowed material
into an isolated student repository/worktree if the full repository would expose
solutions. Keep the source revision and exact baseline receipt. If constructing
an honest baseline exposes a circular dependency, stop, explain it to Bill, and
repair the teaching order rather than invent a new contract to satisfy the test.

The assignment specifies public behavior and exercises, not Cadence's intended
implementation. The student may ask Cadence questions throughout the exercise
using send_message_to_parent; Cadence answers through the owned child's pending
question channel and retains the running join handle. The student may run all
assigned public graders themselves. Supply exact commands, candidate interfaces,
existing offline dependencies and permitted isolated test resources in the
assignment. Grader access must not require editing the grader or obtaining live
credentials. Keep grader sources/tests outside the student's writable solution
scope; questions about grader defects are welcome and investigated by Cadence.
Record each answer and any extra
help; do not call an assisted implementation an unassisted reconstruction. Ask for
an explicit submission containing changed files/commit, commands and outcomes,
architecture reasoning, ambiguous or missing instructions, failed approaches,
proposed clarifications and limits. Idle is not submitted. Retain child/job
handles, answer pending questions, collect the submission, and join teardown.

Run the public acceptance grader against the student's candidate independently
of its own assertions, and review ownership/readability. Reconstructing behavior
and architecture is the criterion, not byte-identical source. Do not automatically
merge or deploy the student's implementation. Integrate useful corrections into
the original chapter construction points; retain attributed Q&A when useful,
without private operational chatter. Retest the changed teaching with a fresh
student when the gap was substantive. Minor wording/link repairs need focused
checks, not an automatic whole-book campaign. Record any untested clarification.

Finish non-local artifact/prerequisite student gates before beginning Local
Models. Chapters whose final upgrade necessarily depends on the local-model
slice receive their final student gate after that upgrade. The last development
slice is Local Models; its student exercise uses synthetic local servers, while
actual Ollama/model acceptance remains a separately authorized human gate.

## 9. Execution order and commit boundaries

1. Record Bill's closeout and student-test authority. Resolve the section 11
   choices at their relevant slices; save the plan and PR attribution.
2. Navigation-only move/stubs and label repair; link/assembly checks. Existing
   release artifacts remain labeled prior outputs.
3. A0/A1: concrete Actions fixture and snippet copying; focused tests and the
   first public grader rows; Bill evaluates interaction before broader polish.
4. A2: Actions summaries/states/reading stability; update Chapter 7 construction
   teaching and Chapter 26 narrative/exercises in the same coherent increment.
5. A3: timestamps and agreed durations; History/codec compatibility, restart
   tests, grader rows and original Chapter 2/10 teaching.
6. A4: one real report offer end-to-end, then refusal/lifetime/resource tests;
   public download grade and mutants; Sandbox/tool teaching; Bill acceptance.
7. Finish Chapter 26's narrative, complete public artifact grade, architecture
   review, scoped adjacent regressions, and fresh student gates for completed
   substantive chapter upgrades. Resolve non-local closeout issues before the
   laptop slice. Commit the verified source candidate.
8. Last development slice, L0/L1: target preflight and compatibility experiment, then minimum config/
   streaming corrections and deterministic grader. No speculative backend work.
9. L2: authorized real-model trial and failure investigation; write Chapter 27
   from the observed workflow. The target is Bill's Alienware laptop, reportedly
   an RTX 4070 Laptop GPU with 8 GB VRAM, with an older Ollama launched inside
   WSL. Hardware, versions and installed models remain unverified. Inspect before
   requesting an update. Add L3 only if its use is agreed and necessary.
10. Finish local-model public grade and real-target report; integrate crossover
    requirements and complete fresh student gates for the final local-model and
    dependent chapter upgrades. Inspect first-book closeout gaps. Changes beyond these two
    chapters are named and agreed, not silently accumulated.
11. Publication and final acceptance below. Deployment remains Bill's decision.

If protocol behavior blocks local testing, report the exact blocker and revise
that slice. If a proposed timing display is not useful, simplify it with Bill
rather than expanding instrumentation to justify the original proposal.

## 10. Publication and completion gates

After both chapters and exercised prerequisites are coherent:

- Verify chapter headings, cross-references, anchors, exercise contracts and
  current-vs-historical grader mappings. No implied universal full-book grade.
- Run hard prose lint; review advisories for readability rather than blindly
  rewrite technical terms. Preserve author voice and Bill's preface.
- Run the publication regression suite. Assemble in an isolated output directory;
  validate exactly 29 chapters and Crossing Over last.
- Build EPUB and PDFs using existing offline tools. Check local resources,
  links, embedded fonts, navigation, labels, clipping, code/tables and print versus
  covered reading-copy distinction. Recalculate page counts; do not repeat 520/521
  as current after adding chapters. No automatic upload or publication.
- Review the final exercise-to-grader matrix: implemented/graded/human accepted/
  platform limits are separate columns. Document any retained historical gaps.
- Bill accepts the Actions workflow and the actual local-model workflow. A clean
  synthetic suite is necessary but cannot replace those judgments.
- Only then describe the first-book closeout as complete, within its stated
  verification limits. Update progress/procedure top, guides and release evidence.

## 11. Choices to confirm before their implementation slice

Recommended defaults, not hidden decisions:

1. Approve proposed chapter order 26 Artifacts and Actions, 27 Running Local
   Models, 28 Crossing Over; historical graders/endpoints unchanged.
2. Work products use click-time bytes initially, not immutable artifact storage.
   Confirm any requirement to freeze the exact offered report before implementing.
3. Use user-click transfer and explicit tool grants; decide whether an additional
   deterministic per-offer human approval is required. Do not mislabel model
   guidance as enforced consent.
4. Start with timestamps and modest total/observation durations, not PR 6's full
   profiler. Agree on useful visible metrics with the fixture.
5. Select the actual Ollama host and installed model, its context limit and local
   inference permission. No installation/download or Wake-on-LAN assumed.
6. Confirm concrete offer-count, transfer-size and capability lifetime policy
   after the first bounded download experiment, before its public acceptance.

The plan can begin with navigation and the synthetic artifact fixture without
waiting for a local model machine. Consequential choices are resolved at their
slice, not by prematurely designing a complete platform.
