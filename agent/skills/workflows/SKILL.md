---
name: workflows
description: Decide when to use executable workflows and build, test, supervise, and recover them through ordinary MCP skill tools.
type: loadable
---

# Writing workflows

A workflow is an ordinary skill tool whose program orchestrates calls through the
Agent that invoked it. It automates that Agent's existing authority; it is not a
new Agent, fleet registry, scheduler, or source of permission. This skill teaches
the method. Loading it starts no server, creates no children, and grants no tools.
Obtain the appropriate coding tools and human authorization separately.

## Choose the smallest useful mechanism

Use a workflow when a substantial task repeats a known sequence, needs bounded
review/revision loops, coordinates multiple owned children, or must resume from
durable progress. Use an ordinary tool for one calculation or operation. Keep a
short exploratory task conversational when its steps are not yet understood.
Do not build a persistent engine merely to replace two straightforward calls.

Agree on inputs, outputs, costs, authority, completion criteria and failure
behavior before implementing. Put ordering, attempt counts, validation and state
transitions in code. Give models the work requiring judgment: writing, reviewing,
interpreting evidence and asking for help. An executable carries the process;
it does not prove that a model's judgment is correct.

## Package an ordinary MCP tool

Use the shipped standard-library helper `../shared/ensemble_mcp.py` and its
[API reference](../shared/python-mcp.md). Keep the helper and executable code in
the writable sandbox or in an explicit host-approved process read-only path.
A symlink grants no access. Host-selected shared code remains mutable by the host;
do not change it during a run. Native file tools remain confined to their root.

Declare the executable tool in the package's `tools` metadata and the stdio server
in `mcp_servers`. Command arguments are literal: no shell interpolation or skill
variable expansion. Preserve the relative sibling layout used by imports. Use
an explicit Python interpreter already available on the host; do not install one
or widen confinement without permission. A launcher such as
`/usr/bin/python3 -u skills/my-workflow/server.py` assumes that exact run layout.
Loading an executable skill can start its MCP server, so server initialization
must not begin the user's work. Begin work only in the declared tool handler.

The helper's `Server` owns the protocol reader, handler tasks, reverse futures and
serialized writes. Register async handlers before `run()`. Never block the event
loop with a sleep or synchronous long operation. Keep stdout for protocol JSON;
use stderr for safe diagnostics. Do not log secrets or private request contents.

The runnable companion [example.py](example.py) implements one read-only callback.
Its reviewed [contracts.json](contracts.json) pins the exact `read_file` input
schema, including observation fields. The native package test compares that file
against the real tool declaration; it never regenerates expectations to pass.
The example imports `../shared/ensemble_mcp.py`, registers argument-free
`inspect_brief`, checks the catalog, calls `read_file` on the fixed run-relative
`brief.md`, and returns the full result unchanged. Read the short companion before
adapting it. This is not a durable multi-step workflow and does not parse a native
Job report into completed file contents.

The instructional skill deliberately has no MCP launcher. The offline test runs
`example.py` as a real stdio subprocess in a disposable sibling package layout;
running it alone in a terminal will wait for protocol input, not write a book.
A production executable skill needs its own reviewed launcher, tool grant, inputs
and human authorization as described above.

The callback targets the actual connection's owning Agent, not a model-supplied
Agent identifier. One protocol reader must keep processing replies while the
handler awaits callbacks, or the exchange deadlocks. Preserve full MCP results,
including `isError`; RPC failure is not an empty successful result.

## Pin contracts and interpret results honestly

Before effects, compare required tools with their currently enabled declarations
using `check_contracts`. Pin reviewed input schemas in the package. Ignore JSON
object-key order, but retain all fields, array ordering and scalar types. Refuse
missing tools, duplicate names or mismatches with the diagnostic tool/path.
Never refresh expectations automatically to conceal drift. Preflight grants
nothing and establishes neither behavior nor result-format compatibility.

Native calls may complete immediately or return a running Job. Keep the returned
handle and observe that Job; do not repeat its operation to obtain a result.
Truncated output is not a complete record. Validate correlation, errors, output
bounds and typed submissions before advancing. Consult the fiction package's
adapter for the actual bare-result and subsequent Job-report cases rather than
inventing a universal JSON envelope or extracting approval from arbitrary prose.

## Bound work and preserve recovery evidence

Use existing owned-child operations. Define assignment identity, required inputs,
result schema and output destinations before dispatch. Persist pending intent
before admission; accept only a matching submission with validated output files.
Keep one run lock, atomic state replacement and durable accepted receipts. A hash
identifies accepted bytes; it does not make a file immutable.

Lost replies, disconnects and cancelled observations do not roll back admitted
work. Reconcile child status and saved submissions before deciding what happened.
An idle or missing child is not proof of success, and a historical submission is
not proof of current-task completion. If evidence cannot resolve admission or
cleanup, stop and ask the supervisor. Never redispatch an uncertain effect merely
to get a fresh reply. Atomic ledger replacement is not an exactly-once transaction
with a child or external service.

Give review loops explicit attempt budgets and completion states. Reaching a
budget is not approval. Keep small editing units and separate assembled review
snapshots from mutable sources. Preserve the original task in pending state so
recovery does not rebuild it from already-mutated output. Use manuscript files
for story facts, role-specific memory for learned practice, and fresh reviewer
conversations for independent review.

Finish file writes and owned work before submitting. Submission alone does not
join independent Jobs. At lifetime boundaries use joined shutdown before copying
role memory or destroying resources. Close admissions, cancel as appropriate,
and join work; requesting cancellation is not proof of completion.

## Supervise the running workflow

The human-facing supervisor retains the outer workflow Job. Call `wait_for_job`
with that handle and `ai_callback_delay: 1200` for up to twenty minutes of healthy
observation. A child question brings the supervisor back sooner without cancelling
the workflow. Answer with `respond_to_parent_message`, then observe the same Job
again if it is still running. Do not start a second driver or consume nested Job
cursors owned by the Python workflow. Do not substitute a shell sleep.

The Python driver handles ordinary progress; the supervisor handles questions,
authority, costs and ambiguous recovery. A stopped application cannot wake itself
through an observation timer. Restart uses the same saved run and reconciliation,
not a new assignment disguised as a retry.

## Test before live use

Use disposable inputs, real subprocess pipes and small faithful protocol peers.
Test normal callbacks, refused calls, contract drift before effects, and a lost
reply after the peer has retained the result. Verify that recovery does not repeat
the effect. Test concurrent-run refusal, bounded attempts, output validation and
joined cleanup where those behaviors are implemented. Keep fake literary
judgments distinct from claims about actual story quality.

Check the real loader: discovery, authorized activation, expected tools and
owned server lifetime. For this instructional skill, loading must leave tools,
MCP clients, children and run state unchanged. Test both copied and host-approved
shared layouts where supported. Passing a fake does not establish live provider
compatibility, entitlement or model comprehension.

From the repository root, existing helper and fiction checks include:

```sh
python3 -m unittest discover -s agent/PythonTests -v
```

The built native Tests runner's `--mcp-process-tests` checks actual macOS process
confinement and package loading without paid providers. See Chapter 21's required
local checks for the isolated .NET build command and independent fiction grader.
To test only the shipped companion, from the repository root:

```sh
python3 -m unittest discover -s agent/PythonTests -p test_workflows_example.py -v
```

Success is four passing tests: a complete normal result, preserved tool refusal,
contract drift before the callback, and a completed read whose reply is lost.
The last case verifies joined exit without a second call; it does not claim
crash recovery. The tests use actual subprocess pipes, disposable input files and
a local protocol peer, with no paid model or network access. Python 3.8+ and the
source checkout are prerequisites. Temporary workspaces are cleaned and server
exit is joined. The nine native MCP/package groups also verify inert loading and
the pinned read schema in copied and read-only mapped layouts on macOS.
These checks do not certify every new workflow. Obtain separate approval for
live calls, credentials, installation or deployment.
