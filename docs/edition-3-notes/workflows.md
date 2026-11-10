# Workflows: fiction-writer — development notes

## Current direction (Bill, 2026-11-08)

Workflows belong in Ensemble and The Art of Building AI Coding Agents, superseding
the earlier Puffin-only placement. Bill confirmed Workflows as Chapter 21, after Managing the Swarm (20).
Former Chapters 21–29 move to 22–30. Historical endpoints and released PDFs/EPUB
remain unchanged. The new Chapter21 is an explicitly unfinished reservation.

Build and learn together before writing the full chapter. No live fiction run,
installation, publication, or new worker is authorized by this requirements note.

## Required example

An autonomous sandboxed fiction-writer workflow, with run_tools but no Internet
access. Inputs include the outline and a work directory containing outline.md,
voice.md, and a skills directory. Whether outline is supplied separately or only
through outline.md remains to clarify. Skills are external/loadable, not built
into Ensemble. Proposed skills are in /Users/bill/sandbox/skills; inspect and
adapt rather than activate them blindly. Skill discovery, security, loading and
proposal/approval details must be settled with Bill.

Provide a reusable Python library for MCP skill servers if one does not already
exist. Declaring a workflow tool should be as simple as declaring an ordinary
Python skill tool. Its distinguishing behavior is calling back into Ensemble to
execute the calling Agent's available tools. This is automation, not expanded
authority: the same grants and security settings must apply to every reverse call.
Do not create a second tool dispatcher or scheduler. Verify actual enforcement,
including workflow process/descendant no-network confinement, not just tool menus.

## Fiction stages

- Author starts each chapter. Reads the full outline, voice.md, prior chapter,
  and carryover.md written after the previous chapter (first chapter has no prior
  chapter). Author/editor details may live in skills or task instructions.
- Editor iterates with the author, requests changes, and explicitly approves the
  chapter before advancement.
- Detail reviewer reads one complete chapter in a single call, with its byte size
  supplied beforehand. No keep_tool_results. Writes review and calls submit,
  ideally within the requested three model round trips (write and submit may be
  in the same response). Do not silently chunk, summarize or omit source text.
- Global reviewer similarly reads the entire assembled book once. Python assembles
  the manuscript and provides its size. Bill said manuscript.py; likely intended
  manuscript.md, pending confirmation. Refuse/report capacity limits rather than
  silently turn full-book review into partial review.
- Agents may propose useful new skills; proposal is not automatic approval or a
  grant to execute their code.

Outer sequence: load durable state; write_book if not already written; then,
while completed review rounds < 3, detail-review every chapter, global-review,
and edit_book. Exact edit_book author/editor turn rules remain to settle.

## Interruption/restart requirement

state.json records review round, phase/turn, chapter and book-written status;
advance after each completed step so a fresh workflow invocation can resume a
run lasting a day. Detailed review chapter cursor and accepted result identity
must be recoverable, not only the outer loop counter. Design question: reconcile
an effect/result completed before the corresponding state update, rather than
blindly repeat work after a lost reply. Atomic state replacement alone does not
make child effects and workflow state one transaction. Concurrent invocation and
cleanup/child reuse policy remain to establish, not assumed solved.

## Schema compatibility — discuss before selecting mechanism

Bill identifies workflow callers silently breaking when Ensemble schemas change.
Cadence recommends a bounded contract check as chapter scope: check required tool
availability and expected schema compatibility before work; enforce live grants
again on calls. Exact schema comparison versus compatibility, input versus result
contracts, and CI coverage remain design decisions. Do not claim novelty or that
schema validation establishes behavioral compatibility.

## Initial inspection (not runtime verification)

- Chapter 5: ordinary C# author/editor/reviewer peers; 9: Skills; 11: reverse MCP;
  19: owned children/submissions/questions; 20: management/steering/status.
- McpClient.ReverseAsync resolves its actual Agent via Parent().Parent(), then
  DispatchReverseAsync enqueues ReverseTool on that Agent. Actor routes it to
  Jobs().DispatchAsync. Full security parity is not yet verified.
- DispatchReverseAsync cancellation stops observing its reply, not the admitted
  operation. A lost reply is not evidence of no effect.
- Proposed fiction-author/editor declare custom collaboration tools and impose
  novel-length/worldbuilding prerequisites. Reconcile with a supplied outline and
  the new workflow turn protocol rather than importing all requirements blindly.
- fiction-reader currently requires 500-line chunks, refine_context and
  submit_review; this conflicts with the new one-read/submit requirement.
- Current Python inventory includes a Moltbook-specific server; no conclusion yet
  about reusable Python MCP/reverse-call support. No package installed or executed.

Lesson: shared dispatch is necessary evidence of common enforcement, not sufficient
proof. Check actual admission and confinement, and test refusal through the real
reverse-call boundary.

## Imported authoring skills and discovery prerequisite

Bill directed this chapter to battle-test dev-driven-specs. Five skills are imported
byte-identically under agent/skills/dev-driven-specs with the upstream license;
the existing visible-reasoning is reused. Cadence's unnecessary duplicate and
primary-list additions were removed. The current Agent catalog is construction-time
and offered-list based, so imports are not yet active. Bill requests automatic
discovery of loadable files in the designated directory, including additions
while running, without primary-list edits. Discovery grants nothing; explicit
load and existing policy/confinement remain necessary. Refresh timing, malformed
files, collisions and source removal need implementation and bounded checks.
Crossing Over (Chapter30) records the encountered issue and pending correction;
fold demonstrated behavior into Chapter9 before the next book execution.

### Discovery implementation checkpoint

Source now discovers all loadable templates for non-sandboxed Agents, refreshing
at model-request and load admission boundaries. Sandboxed Agents retain their
construction templates and require reachability from the primary through depends
or loadable edges; intermediate activation is not required. Resolve checks the
same policy before dependency/MCP preparation. No separate allowlist was added.
Active/pending manuals and templates are pinned, complete candidate parse precedes
publication, duplicates/malformed files refuse, and request-only current discovery
supersedes frozen manual lists. No implicit credentials or server activation.
Clean build and eight focused skill groups passed, including real loopback request
advertisement after copying a file and actual load/refusal cases. Broader suite
first passed381/failed1 due to obsolete progressive startup order; corrected to
positive graph-order checks with invalid-selection refusals retained. Final clean
build and shared382/0 passed. No installation, restart, or live imported-skill
activation has occurred. Chapter30 records this for Chapter9 fold-back.

### Post-restart activation and MCP confinement finding

Bill reported restarting. Actual load_skill("chapter-writing") succeeded and
reported coding plus chapter-writing loaded; existing visible-reasoning instructions
were already present. code-tools subsequently loaded. This supersedes the earlier
not-yet-active checkpoint, without independently verifying installed binary hashes.

Inspection: Jobs/Jobs.cs DispatchAsync checks the caller's ToolEnabled and validates
arguments before execution, including control and skill-load paths. However,
skills/Connections.cs PrepareAsync launches stdio servers with McpProcessTransport;
Mcp/ProcessTransport.cs directly uses Process.Start with a reduced environment and
no OS confinement. Skills.Resolve and Configuration.AllowsSkill do not supply a
compensating process restriction. This is a source finding, not an executed escape
test. Existing run-command confinement is in Sandbox/Command.cs; inspect its contract
before choosing reuse. The required sandboxed/no-Internet fiction workflow cannot
rely on the current stdio launcher unchanged. Skill authorization is not executable
confinement. No MCP repair or live fiction run performed in this inspection.

### First confinement increment

Bill approved keeping workflow Python code, skills and support library inside the
work sandbox; no external skill-library read exceptions. Added working TL;DR and
first confinement exercise to Chapter21 before implementation. Sandbox.PrepareProcess
shares the unchanged command profile; PrepareCommand selects /bin/sh. MCP constructor
now requires caller Sandbox; loader uses it and selects sandbox root for startup
and dynamic load. Unrestricted MCP environment reduction retained; confined launch
environment never overwritten with host HOME/TMPDIR. No unrestricted overload.

Clean build; two focused MCP groups pass, including actual skill loading and real
peer/descendant outside-file + listening-loopback permission refusals. Existing four
filesystem sandbox groups pass. Job28 completed successfully: shell confinement, useful confined work and PTY
lifecycle compatibility all passed (three groups). No installation or fiction execution.
Registration edit initially refused a nonunique anchor before writing/building;
corrected with a unique branch. Direct transport test passed before moving the test
onto actual skill admission, which exposed the host-cwd assumption in the loader.

## Python duplex helper and strict preflight increment

Bill selected exact matching for KISS. Implemented shared/ensemble_mcp.py with
one asynchronous stdio reader, serialized writes, async handlers, correlated
reverse list/call requests, explicit task ownership and joined EOF cleanup.
Ordinary tools and workflows share one decorator; callbacks use the actual
owning Agent, not a new dispatcher. No pip dependencies or external read grants.
The helper is copied inside the sandbox with executable skills. Public API in
agent/skills/shared/python-mcp.md. No installed runtime change or fiction run.

check_contracts compares only expected tool names and complete parsed input
schemas: object order ignored, array order and scalar types retained. Missing,
duplicate or changed contracts refuse; schema location diagnostics exclude values.
No auto-update or automatic retry after uncertain replies. Six subprocess/contract
Python tests passed. Clean C# build, three --mcp-process-tests groups passed,
including real skill load and outer Agent.DispatchAsync -> Python -> reverse
write -> reply. Outside write and ungranted read refused, sentinel unchanged,
unload joined. This is focused verification, not a new full-suite result or an
independent chapter grader. Actual macOS sandbox only; no portability claim.

Meaningful failure: direct client invocation initially passed with an incomplete
fixture tool schema. The stronger outer Agent dispatch rejected undeclared tool
and arguments fields before Python ran. Corrected fixture declaration, preserved
validation, reran successfully. Lesson: testing a connected MCP client alone does
not establish ordinary outer tool admission. Integration preflight uses a schema
from the test Agent to prove interoperation, not as a production drift oracle;
separate Python negatives prove changed expectations refuse before callback.

Pending: fiction state/reconciliation and reviewer protocol; real workflow expected
schema artifact and outside-solution public graders; full chapter/reviews/trial.

## Author/editor continuity and initial memory experiment

Bill accepted keeping Author and Editor conversations through each chapter's
revision loop and starting fresh conversations for the next chapter. He then
selected a first experiment with separate Cadence-authored SOUL.md files and
initially empty MEMORY.md files for both roles, rather than inherited Cadence
memories. Prepared baseline files are in fiction-workflow/author and editor.
Intended role memory survives chapter conversation resets; child storage/recall
wiring is not yet implemented or tested. Detail/global reviewers start fresh.

Bill reports improved first-chapter writing when a prior Agent received
CodeRhapsody's memory. A later comparison using copies of Cadence's current
memory is worth considering; no memory copied or fiction run performed here.
Lesson: fresh conversation and fresh memory are separate experimental choices.
Story facts remain authoritative in manuscript/continuity files, not role recall.

## Assignment-ledger increment

Updated Chapter21 working TL;DR/exercises with fiction turns, separate role memory,
three review rounds and restart reconciliation; regenerated assembled Markdown.
Added agent/skills/fiction-workflow/run_state.py and implementation README. No MCP
skill registration or fiction driver yet. The context-managed RunStore holds a
cooperative Unix process lock, saves intent before dispatch, refuses replacement
of pending work, and accepts matching submissions with driver-required file hashes.
Atomic complete state retains accepted assignments/results and clears pending.
Strict bounded JSON rejects duplicate/unknown fields, invalid numeric positions
and nonfinite constants. Save failure poisons the instance; a new store must reload.
Hashing presumes quiescent producers, not malicious-writer race safety. No directory
fsync, remote transaction or exactly-once guarantee. OS confinement remains required.

Fourteen Python tests passed: six existing duplex/contract tests and eight real-file
ledger tests, including separate-process lock contention, intent/result recovery,
stale submission refusal, missing output refusal, snapshot isolation and a real
filesystem replacement obstruction. These are internal tests, not public graders,
live Agents, installed behavior or completed workflow evidence. No inference used.

Inspected CreateChild.cs, ChildSubmission.cs, ChildRestart.cs, Tools/Children.cs,
ChildStatus.cs and SubmissionRetrieval.cs. Children own their memory under their
own data homes; Corpus.cs reads SOUL.md/MEMORY.md there. Existing spawn inherits
configuration/templates and restores same-ID History. Fresh-chapter memory transfer
and role skill setup still need implementation and actual-context tests.

Critical remaining boundary: spawn creates child then SubmitAsync admits work;
its task_admitted report follows admission and can be lost. Public status has no
caller assignment key; historical retrieval has generation but not a durable
caller-key admission receipt. A ledger cannot distinguish not-admitted from an
admitted interrupted task solely from idle status. Need Bill's decision before
extending ordinary child admission or accepting a visible manual-reconciliation
stop in these ambiguous cases. Lesson: atomic local progress is not atomic remote
admission. No automatic replay has been added.

### Supervised recovery, not a new admission API

Bill selected child questions plus approximately twenty-minute supervisory checks.
The working TL;DR and exercise 6 now explicitly pause uncertain assignments for
supervisor reconciliation; no durable child-admission key is required. Existing
interruptible observations should deliver questions sooner. A stopped application
is not promised a scheduled wake. Ordinary author/editor turns remain automatic.

Added `fiction-workflow/child_calls.py` to consume ordinary reverse-MCP bare JSON
completion or retained Job observations without retrying the original operation.
It refuses truncation, changed handles, errors and non-object/ambiguous JSON.
Spawn's separate admission notice deliberately does not pass this single-object
parser; spawn-specific handling remains to implement. `recovery.py` reads actual
child status and historical submissions, returning observe, validate_result, or
supervise. It never dispatches, accepts artifacts or mutates pending intent.
A matching result remains subject to phase/content/artifact checks.

Verification: 23 Python unittest cases passed, including real durable ledger
files, a separate lock contender, MCP subprocess tests, and small working
read-only child/cursor fakes. The new adapter/recovery tests are not real C# child
integration or public chapter grading. No fiction run, provider request, core
admission change, installation or restart occurred. Source inspection showed
MCP carries text/isError, and Job completion may be bare text on its first report;
subsequent reports retain the handle and cursor. Status is not a shutdown barrier.


### Role-home preparation and actual child context

Implemented `role_homes.py`: first chapter seeds separate baseline identities and
empty MEMORY files; subsequent chapters transfer only the same role's stable and
active corpus files into a new home. No saved dialogue, settings, submission, Jobs
or retired memory is copied. The driver must hold its run lock and shut down the
old child first. Existing homes and unfinished memory transactions refuse rather
than discard recovery evidence. Staging is not hostile-filesystem race defense or
a directory-fsync guarantee. No core child-admission or memory API change.

Six additional filesystem tests passed (29 Python tests total). Added a real-child
C# integration: Python prepares homes, existing CreateChildAsync constructs owned
children, and loopback fake-provider requests contain role identity/learned memory
without parent memory or previous-chapter dialogue. First attempt timed out because
reusing top-level MemoryTurn returned final prose without child submission. Corrected
the fixture to explicitly publish a valid submission; no production behavior was
weakened. Learned memory is synthetic and submission is via the public API, not a
claim of model learning or tool-call reliability. Corrected build: zero warnings/
errors; four MCP/process/workflow groups pass. No fresh full-suite result, installed
changes, provider calls, or fiction run. Next: wire ordinary child dispatch and the
author/editor turn driver; these components are not yet a registered workflow.


### Assignment dispatch without new admission APIs

Added `assignments.py` using existing spawn for a new role and ask/get-submission
for subsequent chapter turns. Intent is saved first; both paths validate the
assignment identity and return only a candidate for phase/artifact validation.
Extended child_calls with explicit parsing of spawn's admission plus completion
records, retaining single-object validation for ordinary controls. No runtime
API changes, silent redispatch, or automatic artifact acceptance. Immediate bare
results now receive the same one-MiB bound as observed Job output.

37 Python tests pass, with real ledger files and a working child protocol fake.
New checks include a lost admission reply followed by reopened-ledger refusal of
redispatch, stale results for both paths, required initial schema, split spawn
records, malformed completion and bare-output bounds. These are not native child
or reverse-MCP integration results. No live fiction, installation or provider call.
Lesson: admission notices are not completion; matching submission is not artifact
acceptance. Next implement the initial chapter author/editor loop and integrate
these pieces before tackling the still-open later revision-phase rules.


### Initial author/editor turn sequence

Added chapter_turns.advance_chapter: one initial-draft turn or pending-result
reconciliation per call, derived from accepted ledger receipts. Author goes first;
Editor requests revision or explicitly approves. Spawn once per role/chapter,
then ask/get-submission for continued conversation. Driver supplies artifact paths.
Editorial acceptance checks that manuscript/carryover retain the preceding Author
hashes; changed drafts or missing files preserve pending intent. Approval stops
further dispatch for this chapter. No second authoritative turn counter.

41 Python tests pass plus whitespace checks. New working fake writes real files,
revises and approves, with store reopen between all turns. Lost admission reply
reconciles a saved result without another turn; missing carryover and changed draft
refuse acceptance. Native reverse-MCP dispatch integration and complete-reading/
capacity checks are still pending. Inspected ChildSubmission confirms submission
checks descendants but does not join independent Jobs: prompts require completed
file work; chapter-boundary shutdown remains the memory-transfer lifetime barrier.
No claim of adversarial artifact immutability or a completed fiction workflow.


### Native Python-to-child assignment integration

Added FictionNativeAssignmentsTests and registered it in the focused MCP group
and full suite. A confined actual Python server calls reverse spawn for a first
assignment and ask/get-submission for a second on the same real child. A loopback
fake provider and explicit public PublishSubmission isolate native dispatch,
assignment correlation and ledger receipts from model tool-use ability. The outer
call uses the MCP client directly; the existing workflow test separately covers
ordinary outer tool admission. Two receipts, no pending assignment, one child;
root destruction joins owned children and MCP processes.

Initial fixture failed before admission because tools metadata used commas rather
than space-separated names. Corrected fixture only. The next attempt treated an
auxiliary recall request as the main child turn and published submission too early;
diagnostic synthetic request showed recall preceding main dispatch. Seeded the
fixture child's settings before spawn using existing codecs with recall disabled.
No actual fiction-role policy changed. Removed verbose diagnostic afterward.
Final isolated build zero warnings/errors; all five focused MCP/workflow groups
pass (Job166). No new full-suite run, live provider, installed change or fiction run.
Lesson: provider request occurrence alone does not identify the main turn; isolate
or explicitly service auxiliary protocols in a boundary test.

### Bounded editorial attempts

Bill selected the same Author/Editor loop for book-wide revisions, capped at
three attempts. He reports that the same AI rarely awards its own novel an A and
that improvement is small after three rounds. This is his experience and our
stopping-policy rationale, not a measured universal quality result.

The working specification applies a maximum of three Author/Editor pairs to each
chapter, both initial writing and each later revision pass. Approval ends early;
three negative decisions retain the latest manuscript, carryover and feedback
and return `revision_limit`, never fabricated approval. The three book-wide
review rounds are a separate bound. All reviews precede that round's revisions.

`chapter_turns.py` derives the limit from accepted receipts and partitions initial
writing and revision rounds by phase/round/chapter. It does not own the full book
schedule or prepare the fresh role children. Four added real-ledger tests cover
negative-review exhaustion across reopen, third-attempt approval, lost final
reply reconciliation, and independent round counts/refusal of round four. All
45 Python tests passed; no new native C# test run, model call or literary-quality
claim accompanies this Python-only change. Full book orchestration is still pending.

### Book schedule and review assignments

After Bill authorized continuation until all exercises pass, added `book_plan.py`:
it derives the next phase from accepted receipts, with initial chapters followed
by exactly three rounds of all detail reviews, one global review and bounded
chapter revisions. Pending positions and accepted prefixes must match; skipped
reviews and work after the final revision refuse. There is no duplicate saved
cursor. Five tests reopen real ledgers, complete both early-approved and entirely
revision-limited books, and detect out-of-order or extra work.

Added `review_turn.py`: source fingerprint and exact full-read arguments are saved
in task intent; source and numbered-result byte budgets are checked separately.
Native read_file numbers LF-delimited source, so the observation cap includes
prefixes and headroom. Bytes are not tokens; these checks do not establish model
capacity. Missing/empty reports and changed sources refuse while retaining pending
intent. Five real-file/protocol-fake tests cover complete report acceptance,
repeated invocation, lost reply, changed text and size/empty refusal. All 55 Python
tests pass. Actual single-read reviewer behavior remains a native integration gate;
no self-report field is being treated as proof of reading.

New packaged role manuals under fiction-workflow/roles use existing tools and
supplied-outline requirements. They deliberately omit the old experimental skills'
mandatory novel length, extra worldbuilding prerequisites, custom protocol tools,
and chunked/refine_context review. These manuals are written, not yet verified
through catalog loading in the integrated workflow. No child or paid provider was
launched. Full driver, immutable review inputs/assembly, role lifetime wiring,
preflight contracts, capacity/full-read checks, public graders and chapter body
remain unfinished. Continue, do not call component passes exercise completion.

### Book driver and review snapshots

Implemented `book_driver.py` and `review_inputs.py`. Pin run inputs before
admission; derive the schedule from accepted receipts; prepare fresh role homes;
join obsolete children before role-memory transfer or book snapshotting. Unknown
shutdown outcomes request supervision instead of inferring cleanup from absence.
Round snapshots retain chapter order and hashes. Reopening verifies frozen bytes
rather than rereading changed working chapters. Final output is separately
assembled under `final-book/` after the third revision and joined cleanup.

Source review found that a revision task cannot be regenerated from its output
files after a lost reply: source sizes may have changed because the assignment
succeeded. Pending chapter tasks now explicitly reuse durable original intent.
A real-file test loses that reply and recovers without another admission.

Jobs 198 and 200: all 63 Python tests passed; Job 198 whitespace check passed.
Two-chapter fake run reopens every assignment, completes all three rounds, joins
all children and verifies final manuscript bytes differ from pre-revision input.
No native test rerun, live fiction, provider call, installation or release build.
Remaining: registered MCP entry point and pinned schemas, native complete-read
and supervision checks, independent public exercise graders. Chapter is not done.


### Configured workflow entry point and pinned contracts

Added source-package SKILL.md, server.py, and static contracts.json. MCP launch
arguments are literal; the reserved owner-directory variable expands in manuals,
not process arguments. Setup therefore supplies the real owner directory at
launch; the argument-free workflow tool cannot retarget it. This is configuration,
not an ownership proof or extra grant. Shared helper stays inside the sandbox.
Preflight runs before ledger mutation. The entry point continues ordinary progress
and returns observe/supervise/complete, preserving uncertain pending work.

Corrected two adapter assumptions before checks: normal chapter progress is
advanced, and recovery uses agents_status. Six pinned schemas include wait_for_job.
67 Python tests pass, including actual stdio handshake/list/preflight failure and
clean EOF join. Clean isolated build plus five native MCP groups pass with the
static contracts checked against actual reverse declarations before real child
admission. No provider, installation, or live fiction run. Full package loading,
actual reviewer read protocol, supervision wake integration and outside-solution
exercise graders remain pending.

### Native package, full reads, and supervision boundary

The complete distributed skill now has a real confined load/unload fixture:
copy workflow/shared/file-reading and baselines into a disposable sandbox, load
the literal launcher, load each role through its authorized dependency graph,
and unload/join. Loading creates neither a book ledger nor children. Corrected
server.py's obsolete beside-helper header to match the actual sibling shared path.

The native assignment fixture now asks the parent a question while Python is
waiting on the original spawn. A twenty-minute child-change observation with its
stuck timer disabled returns promptly for that question, the answer reaches the
child, and the same Python invocation completes both assignments. No timer or
scheduler was added. Source distinguishes spawn's two records from join's
pending-question output; no speculative parser relaxation was necessary.

A new native read fixture uses Python source_info's actual arguments for chapter
and manuscript inputs larger than 32 KiB. Each fresh child makes exactly one real
read_file call; its next loopback-provider request contains the exact complete
numbered text. The fake then explicitly publishes submission. This proves full
read transport, not real-model comprehension, literary review, token fit, or an
autonomous submission exchange. Recall is disabled only in the fixture so requests
cannot be confused with auxiliary recall. Internal Python bounds/refusal tests
remain separate evidence.

Clean isolated build, zero warnings/errors; seven focused MCP/workflow groups
passed. Logs: workflow-read-build.log and workflow-read-tests.log in the existing
isolated workspace. No providers, live fiction, installation or restart. Public
outside-solution exercise checks are still to establish; these are internal tests.

## Local exercise completion and independent stdio check

Added scripts/workflow-grader outside the implementation tree. It starts a
supplied server subprocess and uses its own frozen catalog and working child peer;
it imports neither candidate modules nor candidate tests and never reads the
candidate ledger. Published fixture task labels/setup are not hidden requirements.
All Editors request revision. Two chapters produce57 assignments with three
attempts in each of four writing passes and9 detail/global reviews. The peer
retains a completed submission when its first reply is ambiguous; a new server
process reconciles it, then finishes without duplicate dispatch. Exact ordering,
final output distinct from review input, joined cleanup and completed restart pass.
A disposable four-attempt mutant reached the intended73-versus57 assertion; no
reference implementation edits. This is not a native full-book or literary trial.

Final local evidence:67 Python tests pass; external stdio three PASS lines; clean
new-output .NET build zero warnings/errors; native7 MCP/workflow groups,3 process
groups,8 skill groups all pass. Full shared suite388 passed/0failed. Jobs261,264,
267,270 terminal; no background fixture remains. Clean-output build/focused logs
are /private/tmp/workflow-exercises.ecv1sp; full log is
/private/tmp/cadence-renderer-versions.gOOJ6z/workflow-full-tests.log.

Lesson: the outside-solution grader must drive a public boundary rather than
import the candidate's test helpers or private ledger. Passing that boundary is
still not OS-confinement evidence; native checks establish the separate runtime
properties. All currently specified local exercise checks pass. Teaching body,
Take it for a Spin, reviews, student trial and human acceptance remain distinct;
no live fiction/provider calls, installation, restart or publication occurred.

## Host-approved shared process reads

While writing Take it for a Spin, Bill asked to map shared code read/execute-only
instead of copying it into every run, then authorized implementation. Added
nullable immutable Configuration.ProcessReadPaths: omission inherits; explicit
empty removes grants. Sandbox canonicalizes existing absolute files/directories,
refuses writable-root overlap/duplicates, and retains resolved process policy.
Children may narrow but not widen; model/auxiliary configuration copies preserve
policy. GUI host accepts explicit --process-read-paths JSON (or host environment
EN_PROCESS_READ_PATHS). Skill/model arguments cannot self-authorize paths.

Scope is confined processes only. Native file tools stay rooted; external native
outline/voice reads are NOT implemented. The walkthrough keeps those run-local.
An in-root skills symlink plus approved external target allows Python imports;
link alone grants nothing. OS policy permits reads and explicitly denies writes
for selected targets. Host updates remain possible: shared code must be trusted
and stable during a run, not falsely described as a frozen signed snapshot.

Verified initial clean build and9 focused groups: actual shared package loads,
Python plus descendant and a real native child can import/read but cannot write,
delete, rename, chmod, create shared files, read an unrelated sibling, follow a
symlink escape or connect to a listening loopback socket. Workspace writes still
work; native file read/write remain refused. Added extra unrelated-grant and
directory-to-file narrowing cases; Job303 finished with clean GUI/Tests builds
and full native suite390 passed/0failed. Agent.DestroyAsync fixture compile error corrected to
actual joined Destroy(); old printed log was explicitly not treated as a new pass.

Chapter21 now contains the teaching body (21.1–21.9) and Take it for a Spin.
The walkthrough prepares a disposable two-chapter story, run-local role baselines,
correct container/agents/writer home and a mapped external skills tree. Its live
launch requires an explicit API vendor/model/endpoint/key, isolated environment,
separate port, web/transcription disabled and no ambient global settings import.
No live fiction run, installation or host restart occurred.

Added the minimal fiction-supervisor primary inside the package, with role-aware
instructions so inheriting children do their assignment rather than start a nested
book. The actual package-loading fixture now uses that primary rather than a
fixture-generated primary. Both copied/mapped modes pass among9 focused native
groups after a clean build (Job311). This adjustment follows the390-test full run;
no second full run is claimed. Python67 tests and the external stdio grader's3
checks pass. Executed the published setup Python block in a disposable directory:
correct owner home, mapped server and empty Author/Editor memories verified.
The live GUI/provider command is source-checked, not executed or human-accepted.

Chapter prose has no hard failures. Retained negation/person-gap advisories because
security refusals and recovery distinctions teach the contract; no invented story
or person insertions to meet a count. Assembled Markdown regenerated and verified:
Chapter21 exact inclusion once,31 headings00–30 in order, chapter local links valid.
PDF/EPUB releases unchanged. Independent reviewers/student/human acceptance remain
pending under the current no-delegation boundary. No staging or commit performed.

Lesson: shared read permission, skill authorization and writable role memory are
separate relationships. A symlink does not grant authority, an approved shared tree
is not immutable to the host, and a correct process setup must name the actual
Agent home rather than the application's container.

## General workflow authoring skill

Bill requested an instructional `workflows` skill and a real presence/load check.
Added the tool-free manual under agent/skills/workflows, available from ensemble
and fiction-supervisor. There is no existing `skills` skill; the later Skills9
requirement to create a skill-writing skill and expose workflows through it is
queued privately, not implemented here. The manual teaches selection, packaging,
duplex callbacks, pinned schemas, bounded work, recovery and outer-Job observation.
The package fixture now loads the exact manual in copied and shared-read layouts,
asserts unchanged tools/no MCP clients/no children/no ledger, and unloads it.
Isolated Tests build: zero warnings/errors; nine MCP/process/package groups pass.
This proves packaged loading, not installed activation or authoring-example
execution. Exercise8 still requires executable checks of the illustrative callback.
No provider, installation, restart, staging or commit occurred.

### Executable authoring companion verified

Added workflows/example.py and reviewed contracts.json. Removed the duplicate
manual code block in favor of the tested companion. Four real-stdio tests cover
normal read, preserved refusal, preflight drift refusal and lost completed-read
reply with joined exit/no retry. A native assertion compares the pinned schema
to actual read_file; it does not regenerate the pin. Full Python71 pass; clean
build0warnings/errors and nine native groups pass. This closes the local
Exercise8 example gap, not live provider, durable recovery of this stateless
example, independent review, or human acceptance. Published the focused command.

## One code review and incorporation

Authorized reviewer workflow-code-review submitted one consolidated read-only
report and was shut down. Two required findings accepted: expected domain
refusals were erased by generic MCP errors; writing artifacts were hashed and
committed before basic text validation. StateError now has a fixed safe-message
contract and the public handler maps expected constraints and file/UTF-8/JSON
failures without exposing arbitrary exception strings. Writing outputs are
stream-validated as nonblank UTF-8 before receipt acceptance; pending intent and
prior receipts survive refusal, including recovery and final Editor feedback.
All existing StateError constructions were checked to use literal diagnostics.
New artifact negative initially failed18 combinations; the initial fixture reused
tool name as artifact name, causing six fresh-author cases to return supervision
instead of validation. Corrected that fixture shadowing. Final Python74 pass,
including actual stdio source/result-limit refusals with existing pending state
and malformed recovered output without redispatch. No re-review requested.
Lessons: internal refusal and public explanation are separate contracts; hashes
identify bytes but do not establish usable content. Native/external checks next.

Post-incorporation external stdio grader: three PASS; native MCP/process/package:
nine PASS. Existing compiled native harness reads the corrected Python/manual
package; no C# changed in this incorporation. Python74 above is current. No live
provider or human walkthrough. Code review is complete; text review follows once.

## One text review and incorporation

Authorized reviewer workflow-text-review returned one consolidated read-only
report, then was shut down. No blocking teaching defect was found. Updated
chapter acceptance status and stale current counts in the linked package
reference; retained the distinction between current focused evidence and the
earlier full-suite run. Clarified Exercise7 to derive chapter/round position from
accepted assignment/result information, not require a second mutable cursor.
Bill's approved first-person opening remains intact. No re-review or new runtime
checks were required for these documentation changes. Student reconstruction and
the live human walkthrough remain pending; neither review establishes them.

## Prepared student boundary (not launched)

Source-only committed baseline75ae832e prepared at
/private/tmp/cadence-workflow-student.33wxsmy3/solution, with current Chapters0–21,
voice guide and public grader README/catalog. No Git history, target Python
implementation/tests, new shared-read source, repair notes or grader source.
Offline baseline Tests build passed with zero warnings/errors using existing
SDK/cache and cleared package sources. No baseline tests were run here.
Proposed scope is all eight exercises, including native confinement/shared-read
work; student supplies its own regression tests, author executes the external
grader against the student's server. Source-directory isolation alone is not an
OS access boundary. TRIAL.md records restrictions and pending scope approval.

Bill approved this boundary. Single student workflow-student launched, task
admitted as Job496. All eight exercises assigned; external grader remains
author-operated. No student result or human live walkthrough yet.

### Student clarification 1: supplied inputs and result contract

Student correctly reported missing published Author/Editor baseline files in the
trial workspace and missing typed result fields in the public grader reference.
Copied the four public identity/memory inputs byte-identically, verified empty
MEMORY files, and supplied the baseline README after correcting stale status.
Documented fake submission payloads and response envelopes in the public grader
README, then supplied that same reference to the student. No target source or
private grader code was provided. This is assisted clarification and a setup/
teaching correction, not a student defect or proof of original sufficiency.
Lesson: ship the exercise inputs and document result interfaces as well as task
labels. Grader behavior was not changed.

### Student external grader results and correction

Run1 failed missing-contract refusal because the candidate's result lacked isError.
Student corrected its envelope and added a stdio regression; no solution patch
was supplied. Run2 passed that check then crashed in grader review-path parsing.
The grader assumed reference reviews/rN-cN.md and chapter filenames despite the
public contract allowing different internals. Corrected the grader to identify
review scope by actual fake-written chapter bytes, track task-declared chapter
destinations, and compare final output with latest chapter bytes and the last
global-review source. Schedule/count/recovery/cleanup requirements unchanged.
Corrected grader passed reference and student (run3), all three checks, exit0.
No private grader source supplied to student. Native/local work still in progress.
Lesson: external protocol use alone does not prevent reference-shaped grading;
observable content and public labels must determine acceptance, not hidden paths.

Student setup correction2: existing SkillTests.NativeCodingGuidance requires
.agents/skills/ensemble-coding/SKILL.md outside the supplied agent tree. Supplied
exact committed75ae832e bytes and verified readback; no test was weakened.
Baseline build success had not established this repository-relative prerequisite.

## Student completed: assisted reconstruction

workflow-student submitted completion of all eight exercises, then shutdown
returned stopped:true. Final author-run external grader passed all three checks
on the final candidate, exit0. Inspected saved student logs: Tests build619 clean;
MCP620 five PASS; Python630 twenty PASS; GUI build log clean. Process checks in
Job571 passed, but that combined command exited1 for the missing skill-manual
setup; corrected skill rerun589 six PASS, exit0. No new full-runtime run claimed.
Student source remains isolated and was not imported. No second student/reviewer
or live fiction run.

Incorporated feedback into the original chapter: refusal isError envelopes;
typed delegate submission before final text in fake providers; published payload
shapes and layout independence; source-only prerequisite/test resources and
absolute scratch/grant paths. Existing third-attempt, frozen-input/recovery and
synthetic-memory limits retained rather than duplicated or expanded into gates.
Acceptance status now records assisted student success, not original unaided
sufficiency. Human live walkthrough and final acceptance remain pending.


## Creative authorship and Editor permission (2026-11-09)

Bill requested Authors improve the story rather than fill in a fixed plan, including
current/future outline revisions, and obtain Editor permission for earlier-chapter
repairs. Implemented personally; Bill explicitly waived a new student trial.
No additional reviewers, fiction roles, installs, restarts or live run were started.

The original outline/voice stay pinned. Plan version2 seeds working-outline.md;
Author receipts include it. Optional typed prior_chapter_requests and Editor
prior_chapter_permissions identify integer earlier chapters. Permission must be a
subset of the Author request, accompany revise, and have a remaining Author turn.
Scope/reasons stay in carryover and feedback. The next Author receipt hashes permitted
chapter/carryover repairs. Permissions expire after that turn; subsequent Editors
in the loop still read all earlier repairs before approval. No fourth attempt or
new permission registry. Late/denied requests are retained for later scheduled
revision or parent supervision. Current carryover is now an explicit revision input.

Latest accepted hashes protect other artifacts. Pending authorized outputs may have
changed after a lost reply; reconciliation reuses the saved task and preceding
permission without redispatch. Refusal preserves evidence, not rollback or OS file
confinement. Frozen review rounds and final output include outline/voice copies.
Old version1 plans refuse rather than silently migrating. No existing live book
needed migration. Role manuals, original Chapter21 teaching/exercise points, package
README, public grader contract and assembled manuscript are updated.

Evidence: initial positive creative tests failed at the old strict submission schema;
new nine focused cases passed, then source review found final Editor must continue
reading earlier repairs from prior attempts. Added a three-chapter regression and
fixed that reading set. Final Python84/0. Native existing built Tests runner with
current source package --mcp-process-tests:9/0; no fresh C# build/full-native claim.
External grader now independently runs ordinary and creative peers via stdio: seven
PASS lines, including earlier repair and lost-reply recovery; both runs retain57
assignments and joined cleanup. Final assembly31 ordered chapters/exact Chapter21
once; diff whitespace passed. Prose no hard failures, negative-wording/person-gap
advisories retained. Source/book changes remain uncommitted and releases untouched.

Lesson: a stable original plan is useful recovery evidence, not a requirement to
freeze creative discovery. Write permission expires, but approval must still cover
the changes made under it. These local tests establish protocol behavior, not
literary judgment or real-provider role compliance.

## Human acceptance and deliberate stop (2026-11-09T23:17:42.891540+00:00)

Bill called the live workflow a successful test and asked to stop the novel and
close Chapter 21. Chapter 21 is human-accepted and closed. The novel is unfinished
and deliberately discontinued, not completed or certified for literary quality.
The live driver reached chapter five's Author; whole-book review/final assembly
were not established by this trial. Bill waived a new student trial for the
creative-authorship extension.

Reminders confirmed disabled. Outer Job 5 marked killed; workflow skill tool grants
revoked and its external server reported stopped. The still-active chapter-five
Author was separately cancelled/joined with kill_agent returning stopped:true.
Drafts, original/working outlines, receipts and pending ledger are preserved;
do not resume or redispatch without new human authorization. No publication,
installation, restart, release regeneration, commit or push was performed.

Lesson: cancelling an outer workflow does not establish child shutdown. Stop the
server and separately cancel/join admitted children; preserve pending state rather
than falsely advancing it to completed.

## Optional imported-memory experiment after chapter closure

Bill authorized a separate comparison with copies of Cadence's active memories
for Author and Editor. Added explicit imported-memory.txt baseline opt-in while
retaining empty initial memory by default; active corpus copies exclude retired
entries and operational files. Eight role-home tests/full86 Python tests passed.
This is not a reopened chapter acceptance/student campaign. File-copy verification
does not prove every memory appears in a request or causes better fiction.
