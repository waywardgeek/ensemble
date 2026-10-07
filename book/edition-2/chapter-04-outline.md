# Second edition: Chapter 4 outline

Status: full contract/prose in `chapter-04.md` independently accepted after
the fixture and lifecycle clarifications recorded in `chapter-04-review.md`.
Ready for the coordinator's student handoff after Chapter 3 validation.
No new Chapter 4 run or grade is claimed.

## Voice plan and stake

Stake: the reader starts a command that needs input or outlives one model
request, and must regain control without losing its output or mistaking a
stopped wait for stopped work. Resolve through an actual interactive process,
its recorded handle, incremental output, and verified shutdown.

Open with the capability failure: a debugger waits for input while a
synchronous dispatcher waits for its exit. Explain the two waits before
introducing jobs. The first edition's screenshot incident is historical
source material; do not retell an AI narrator's first-person account as a
new personal memory. Omit the retired corpus percentages and provider/model
defaults. No unconditional claim that a job wrapper can prevent every crash
or kill an arbitrary Go goroutine.

## Thesis and teaching order

The job owns continuing work. A tool call's result can report a still-running
job; ending that report does not finish the job. Waiting, observing new output,
and killing are distinct operations with different records.

1. A synchronous call cannot drive an interactive program; a timer alone
   abandons its result. Start correctly with retained job ownership.
2. Locate Jobs, Job, Call, Registry and parent interfaces within the existing
   star. Explain who creates, reports, finishes, and closes each object.
3. Publish the log additions and three-state lifecycle before implementation:
   running, done, killed; completed success is distinct from missing exit code.
4. Put job creation at the dispatch boundary for every ordinary tool.
   Supervision operations are explicitly marked exceptions in the registry.
5. Explain wake delay, new-output pattern, and completion; none is an implicit
   cancellation deadline. Retain unread output and report a handle.
6. Teach one-shot tool_limits with next-call consumption, explicit-argument
   precedence, and a visible consuming-call note on every dispatch path.
7. Separate full produced output on disk from bounded reports in context;
   retain head and tail, exact omitted/total byte counts, locator and cursor.
8. Add an actual PTY, interactive input, process groups and per-call cwd.
   State explicitly which Chapter 3 shell-output conventions change.
9. Explain shutdown, killed goroutines, late completion, log-write failure,
   and why replay can never relaunch recorded work.
10. Deterministic controls, real debugger/process demonstrations through all
    three model paths, independent standard comparison and revision.

## Proposed scope

Continue validated new Chapter 3 history in `solutions/edition-2/ch04`.
Keep the public library, CLI, observer attribution, and separate optional
GUI module. No actor mailbox, turn interruption, streaming model transport,
cross-process recovery, remote tool connection, or GUI transport is added.

Add `wait_for_job`, `send_input`, `kill_job`, and `tool_limits`; these four
supervision operations do not themselves create jobs. Every other enabled
tool is supervised, including read/list operations. Capability visibility
and dispatch permission still share one Registry authority.

Preserve the historical model-selected wake policy: three seconds and
16 KiB as defaults; pattern absent. Pending one-shot settings are consumed
by the next attempted call, including a supervision operation, unknown tool,
or another setter. Explicit arguments win over pending settings. The result
names the consumed setting; a call consuming nothing has no such note.
Validate pattern/ranges before starting effects. Publish precise treatment
of invalid setters and invalid consuming calls in the full contract.

Jobs can run concurrently after serial dispatch has returned their initial
reports. Tool-call/result ordering remains the model batch's order. The
sixteen-request turn bound does not kill jobs; Agent/application shutdown
is the explicit cleanup boundary. A library request error and closing the
Agent must not be silently treated as the same operation.

## Owner proposal to resolve before code

Working proposal, not a claimed Bill ruling: Agent owns a Jobs service in
`internal/jobs`, with its actual Agent interface parent. Jobs owns live jobs;
Engine owns a live Call if introduced. Engine/tools reach Jobs through their
owner interfaces, never sibling imports or a Call dependency bag. Shared
job snapshots, limits, statuses and interfaces live in common; job behavior
stays in jobs. Tool declarations/execution stay in tools.

Each job has one completion authority. For an ordinary function, its worker
records completion. For a subprocess, the process/output lifecycle owns
completion after draining output and obtaining status. The dispatcher must
never finish a process job merely because its start function returned.

Handle and artifact allocation must support multiple Agents without collisions
or cross-Agent access. A constructed Ensemble-owned allocator is a possible
way to retain the first-edition process-wide increasing handle contract
without a package global. Keep workspace resolution on Agent. Choose the
artifact-root/relative-locator policy explicitly and test two Agents sharing
a workspace as well as Agents with separate workspaces.

## Compatibility points that need explicit teaching

- PTY output merges stdout/stderr and echoes input. This deliberately replaces
  Chapter 3's separate-stream shell result; do not claim byte-identical shell
  output across that change. Keep actual exit status and short-result content.
- Chapter 3 tool selection limits remain meaningful. A job's artifact holds
  the full **produced result**, not necessarily every byte of a file when the
  caller requested a range or `max_bytes`. The inherited large-read fixture
  explicitly requests two million bytes. Its report is separately capped.
- A shell now spools full merged output to disk; `max_output_bytes` becomes
  the report budget, replacing the previous per-stream capture budget. Teach
  this change before the student implements it.
- A short completed result with no consumption note can remain bare tool text.
  Running/capped/subsequent reports need status/handle/locator information.
  Define the cursor to advance over the entire reported interval, including
  bytes represented by a truncation stub, so omitted bytes are not repeatedly
  offered as new output. The artifact remains the recovery route.
- `cwd` is a per-call override resolved against Agent's workspace. Missing
  directories fail. Nothing changes process cwd or the next call's default.
- Retain serial event append and observer order even though job workers run
  concurrently. Background workers do not bypass Agent's durable append path.
- A failed log must not prevent process cleanup. Report inability to persist
  cleanup truthfully; do not leave a child alive just to avoid an unlogged kill.
- Killing a goroutine-only job marks it killed and prevents later results
  from resurrecting it, while reporting that the underlying goroutine may
  still run. A process kill must stop the process group and wake waiters.
- Keep job references as metadata/recovery locators; do not silently expand
  Chapter 2's unsupported handle-blob vendor attachment mapping.

## Inherited checks and additional controls

The current inherited score has nine checks: ch3parity 10, jobmodel 25,
waitjob 10, sendinput 10, debugger 5, killjob 10, bigoutput 15,
toollimits 10, shutdown 5. Preserve those checks and earlier regressions.
Verify actual current fixture behavior before publishing final wire bytes.

Add explicit ownership/import checks for the new spoke and all clients;
two-Agent handle visibility/artifact isolation; pre-dispatch persistence
failure; post-effect persistence failure and cleanup; concurrent completion
through one writer; no early close before late output; finished-job wait;
stale-pattern and cursor controls; killed-state immutability and absent exit
code; output-limit edge cases; no live effect during replay. Reconcile PTY
output changes with Chapter 3's stronger second-edition checks openly.

Run timing-sensitive deletion controls serially or under an explicit resource
bound. Preserve positive controls and exact failing-check sets. A mutation
that fails to compile or a missing debugger that gets skipped is not evidence
of detection. Build fixture programs before measuring their runtime.

## Required actual demonstrations

For each introduced API, run the real CLI through a slow job, repeated wait,
input/prompt response, kill and shutdown, capped output with file recovery,
one-shot limits, and per-call cwd. Drive a real interactive debugger to a
known value. Public-consumer evidence covers independent Agents and observer
ownership. Deterministic fault controls remain labeled separately. No saved
first-edition transcript is a new run.

## Next action

The full draft chooses Agent-owned Jobs and an Ensemble-owned handle allocator,
as the coordinator authorized. It adds job_ended alongside job_killed so
background completion is durable and observable. It also finalizes synthesized
missing call IDs using the actual response sequence under the append gate;
background completion during HTTP cannot steal a predicted identity.

Independent contract findings are resolved. Handoff includes the full chapter,
mandatory skill and validated predecessor, not this author-only research
outline. Actual student implementation and demonstrations remain the next gate.
