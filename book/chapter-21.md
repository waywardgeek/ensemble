# Chapter 21: Workflows

> **Implemented and exercised.** The local checks, independent code and text
> reviews, and assisted student reconstruction are complete. A supervised live
> novel run completed 30 chapters and its first revision pass before Bill chose
> to stop and read it. See the acceptance section for recovery lessons and limits.

If you haven't already, why not ask your most capable model, perhaps Fable 5.2
or GPT-6 Astra as you read this, to write a fiction book so good it becomes a
bestseller and makes you rich?

I've tried. Well, maybe not the part about making me rich, but I did ask for a
book I would enjoy. The result, from Opus 4.6, was my first self-published book,
the novella *The Architect of Echoes*. You can read it for free at
[CodeRhapsody.ai](https://coderhapsody.ai) if you like, but it's not very good.

That was the first time I realized I needed a workflow. Asking for a book is not
the same as arranging for one to be written. A model is not going to reliably
produce a coherent, revised manuscript of 100,000 words in one turn. The work
needs chapters, continuity, criticism and revision, spread across many turns.
Someone, or something, has to orchestrate that creation.

A book-writing skill should therefore do more than give the model instructions
to interpret. It should provide a program that executes the workflow. The models
still write and judge the prose; the program keeps track of whose turn it is,
what they need to read, and what happens next.

This is a small extension of what skills already do. They can already host MCP
tools, and those tools can call back into the Agent that invoked them. A workflow
is a tool that uses those callbacks to orchestrate sub-agents automatically.
Making that work recoverably takes care, but it does not require another Agent
manager.

My first version predated sub-agents. I built a custom stand-alone executable,
an orchestrator, to coordinate the writing. Running it produced my second
self-published book, *The Silence Between Notes*, also the second book in the
series. It's not very good either, but I did
enjoy it. I do not recommend it.

What stood out was how much better it was than the first book: longer, more
detailed, more interesting. That comparison is my experience, not a guarantee
that a workflow will produce good fiction. But it showed me why orchestration
matters. A better process gave the model room to do better work. That is what
this chapter builds: not a bestseller machine, but a way to carry a substantial
creative task through writing, review and revision.

Managing the Swarm supplies owned child Agents and their supervision tools.
Workflows use ordinary Python MCP skill tools to automate those operations
through the calling Agent, without expanding its authority. The example writes
fiction from an outline and voice guide, coordinates author/editor turns and
independent reviews, and resumes from durable state after interruption.

The implementation joins skill discovery, a reusable Python MCP interface, strict
tool-contract checks and supervised recovery. The
[development notes](../docs/edition-3-notes/workflows.md) preserve the experiments,
corrections and verification limits behind this construction.

## TL;DR

A workflow is an ordinary skill tool whose program calls the initiating Agent's
existing tools. It adds automation, not authority. The fiction example will keep
its executable skill files and Python support inside the sandbox or in an explicit
host-approved read-only process path outside it. Processes and descendants cannot
modify those shared inputs, access unrelated private files or use the Internet.
Native file tools retain their writable-root boundary; this process allowance does
not grant them external reads. System runtime reads retain the existing policy.

Python coordinates the existing owned child Agents; it does not introduce another
Agent scheduler. The Author starts each chapter, reading the complete outline,
voice guide, previous chapter when present, and carryover. Author and Editor
alternate for at most three Author/Editor attempts per chapter. Explicit approval
ends the loop early. After the third attempt, retain the latest draft, carryover
and criticism and advance with status "revision limit reached" if not approved;
never relabel that outcome as approval. Their conversations
continue through that chapter's revision loop, then start fresh for the next
chapter while each role retains its own learned memory. The initial experiment
uses separate authored SOUL.md files and empty MEMORY.md files, not Cadence's
memories. Manuscript and continuity files remain authoritative for story facts.

After the first draft of the book, perform three review-and-revision rounds:
review each complete chapter independently, review the complete assembled book,
then revise each chapter through the same bounded Author/Editor loop. Complete
all reviews of a round before revising, so they assess the same book version.
Use fresh chapter conversations with retained role memory for each revision pass.
The three-attempt chapter limit and three book-wide rounds are separate bounds.
Each detail and global reviewer starts fresh, is told the
source byte size, and reads its entire assigned text in one call without
keep_tool_results. It writes its review and submits, aiming for about three model
round trips. If read bounds or model capacity prevent the complete reading,
report the constraint; do not silently chunk, summarize, or omit text.

Persist the phase, review round, chapter, turn, pending assignment identity and
accepted results so a later invocation can resume. Record an assignment before
dispatching it, then reconcile existing child work and matching results before
advancing or dispatching anything again. A lost reply is not rollback, and an
atomic state-file replacement is not a transaction with child effects. Memory
retention and recovery are requirements to demonstrate, not guarantees supplied
merely by writing files.

The initiating Agent supervises the run rather than participating in every writing
turn. Children may ask it for help through the existing parent-question mechanism.
While work is running, the supervisor calls `wait_for_job` on the existing outer
fiction-workflow Job with `ai_callback_delay: 1200`: at most twenty minutes of
observation before checking progress. A child question wakes it sooner without
cancelling the workflow. This
uses existing child and Job mechanisms, not a second scheduler, and does not
promise to wake a stopped application. Ordinary turns proceed automatically.
If restart evidence cannot distinguish an unadmitted assignment from interrupted
work, pause that assignment and ask the supervisor to reconcile the existing
child, saved submission and output files. Do not automatically redispatch it.
This supervised recovery is the intended contract; a new durable child-admission
key and exactly-once execution are not requirements.

## 21.1 A workflow is a tool, not another Agent manager

A book is a useful test because its individual operations are familiar and its
failure modes are expensive. Write a chapter. Ask an Editor what fails for the
reader. Revise it. Repeat tomorrow, after the process has stopped halfway through
something. The difficulty is not calling an Author twice. It is knowing whether
the second call is a revision, a duplicate, or work whose result already exists.

Chapter 19 supplies owned children and typed submission. Chapter 20 supplies
inspection, steering, questions, saved results and joined teardown. A workflow
composes those operations. Python decides which assignment comes next; Ensemble
still owns the Agents, grants, Jobs and conversations. There is no Python fleet
registry that can declare a child finished when its native owner disagrees.

The ownership path is concrete. The initiating Agent owns its Skills manager;
the loaded skill owns an MCP connection and its process. Reverse calls on that
connection return to the same Agent's actor and ordinary dispatcher. A Python
argument cannot select a different owner. The actor checks current tool grants
and validates the call before effects, just as it does for a model-requested
operation. Listing tools is an observation of that authority, not a grant.

The C# additions keep the same hub-and-spoke structure. Common declares process
preparation and skill-discovery operations. Sandbox implements confinement;
Skills resolves the selected catalog and starts owned connections; MCP implements
the stdio transport. These modules reach the caller through their actual parents,
not a supplied Agent ID or a copied service bag. The Python package sits outside
that ownership graph and uses the public protocol.

During development, inspection found that the old MCP launcher cleaned its
environment but started an ordinary subprocess. That does not prevent reading a
private file or opening a socket. The correction factors process preparation out
of the existing command sandbox: shell commands and MCP executables use the same
policy. A sandboxed skill starts in the sandbox root with literal arguments.
Unsupported confinement refuses.

Shared code need not be copied into every writable run. The host may explicitly
select absolute process read paths: a directory permits its descendants, while a
file permits only that file. Sandbox resolves existing physical targets before
launch, refuses overlap with the writable root, and adds read permission plus an
explicit write denial. A link named `skills` inside the sandbox can point at that
approved directory. The link supplies a convenient pathname, not permission.

These grants are immutable launch policy, not something a skill declares for
itself. Children inherit omission, may supply an empty list to drop all external
reads, and may narrow a directory grant to a file. Widening refuses while retaining
an effective policy no broader than the parent's. Model selection preserves these
security facts. Native file tools still use their original rooted handles and do
not gain external reads or writes from a process allowance.

Use a trusted stable version for shared code. Read-only means the confined process
cannot modify it; it does not prevent the host from replacing it during a run.
Outline, voice and mutable book state remain inside the run directory in this
implementation. External read-only native document inputs are not implemented.

This matters even for a polite writing tool. An outline is data from outside the
workflow, and a process can perform effects without asking the model for another
tool call. The confinement test therefore uses a real Python peer and a descendant:
workspace access succeeds, an outside sentinel cannot be read, and a connection
to an actually listening loopback socket is denied. A refused connection to a
closed port would prove nothing about the sandbox.

The model transport remains in the host. Denying the Python process Internet
access does not prevent the host from making the approved model requests needed
to write the book. It prevents the skill process from becoming a second,
unaccountable network client.

## 21.2 Keep reading while a handler waits

The Python server receives an outer call to continue the book. To perform it, the
handler calls back into Ensemble to spawn an Author. Ensemble must return that
callback's result while the original call is still pending. A synchronous server
that finishes one handler before reading the next line waits forever: the answer
it needs is sitting on its own input stream.

The shared `ensemble_mcp.py` helper separates protocol reading from handlers.
One asynchronous reader consumes newline-delimited JSON-RPC messages. Requests
start tracked handler tasks; replies resolve pending reverse-call futures by ID.
A serialized writer prevents two tasks from interleaving JSON bytes on stdout.
The original request ID and reverse request IDs remain distinct.

An ordinary skill and a workflow use the same declaration mechanism. The only
difference is that a workflow handler uses its connection to list and call tools:

```python
@server.tool("describe_workspace", "Inspect the configured workspace.", schema)
async def describe_workspace(connection, arguments):
    return await connection.call_tool("list_directory", arguments)
```

The example returns the complete MCP tool result, including `isError`. It does
not extract a convenient text string and discard refusal. A refusal produced by
the workflow itself must likewise set `isError: true`; alarming text in a success
result is not a protocol refusal. Check this through actual stdio, including
that failed preflight admits no child work. In a real workflow, validate the
required contracts before this call. Async handlers must also avoid
blocking the protocol reader with synchronous long-running work.

EOF fails pending reverse waits, cancels handlers and joins them. It does not
undo remote operations already admitted by Ensemble. Handler failure and process
disconnect must not trigger an automatic retry of the last callback. A missing
reply is a missing observation, not a transaction rollback.

This server uses standard-library Python, not another package installation. Its
helper belongs beside the workflow package, either inside the sandbox or in the
same explicitly approved read-only shared tree. The configured launcher passes
literal arguments; `$AGENT_DATA_DIRECTORY` in a manual is not
magically expanded inside an MCP process argument. The operator supplies the
actual owner directory in a fixed launch-time setup file.

## 21.3 Refuse a changed contract before writing

The workflow depends on six ordinary operations: spawn a child, queue another
turn, retrieve a saved submission, inspect children, observe a Job and shut down
a child. `contracts.json` records their expected input schemas. Before opening
the run ledger, the handler obtains the owner's enabled tool catalog and compares
those schemas.

Bill chose exact structural equality instead of a compatibility engine. Ignore
JSON object-key order and formatting. Preserve every field, scalar type, value
and array position. Missing tools, duplicate names and mismatches refuse with the
tool name and structural path. Descriptions outside the input schema do not affect
this comparison; descriptions inside it do. Even a harmless optional addition can
therefore stop preflight. That is a deliberate maintenance cost, not permission
to regenerate the expected file automatically.

The comparison catches interface drift, not behavioral equivalence. A tool with
an unchanged schema can still mishandle cancellation. A tool present at preflight
can be revoked before invocation. Every actual reverse call continues through
ordinary admission, and behavioral checks remain necessary.

There is a second protocol boundary after the call. An immediately completed
non-command result may be bare text. A running operation returns an observation
report with a Job handle; later observations contain more output. The adapter
retains that handle, accumulates the observed output, rejects truncation or a
changed handle, and interprets the result only after completion. It does not
retry the original operation to obtain a cleaner response.

Spawn has two records: `task_admitted` and the eventual completion report. The
first is not a submission. Continued chapter turns use `ask_agent`, followed by
`get_submission`; that saved payload must identify the current assignment. An
old, perfectly valid submission is still the wrong answer to a new task.

For the public fiction fixture, Author and Reviewer submissions contain
`{"assignment_id": "ID"}`. Editors add `"decision": "approve"` or `"revise"`;
the assigned files contain the prose and criticism. Authors may additionally
submit `prior_chapter_requests`, an array of earlier chapter numbers. Editors may
answer with `prior_chapter_permissions`, a subset authorized for the next Author
turn, only with `revise` and a remaining attempt. Reasons and repair scope belong
in carryover and feedback. Omitted arrays mean no request or permission. The
[public fixture contract](../scripts/workflow-grader/README.md) defines the
completed response envelopes and task labels. It does not prescribe private
module names or review filenames.

A fake provider testing delegated work must arrange a valid typed submission
before ordinary final text. A final message alone does not complete the delegated
assignment. Otherwise a test can wait forever while the runtime correctly waits
for a submission; disabling recall does not repair that protocol mistake.

## 21.4 Fresh conversations, retained judgment

The Author and Editor have different jobs. The Author writes the story in the
supplied voice. The Editor identifies reader problems and proposes useful
revisions without replacing that voice with its own preferences. The manuscript
and carryover own story facts; neither role's recollection can silently overrule
them.

The outline is a living plan, not a form to fill out. Drafting can reveal a better
scene, a stronger motive or an unexpected connection. The Author may revise the
current and future plans in a separate working outline while preserving the
human's core promises, voice, chapter count and established story facts. The Editor
judges whether a departure improves the book, not whether every original beat was
obeyed. Consequential changes and their effects on later payoffs belong in carryover.
Changing the plan does not retroactively change an already-written chapter.

The driver preserves the original outline and voice as pinned inputs, seeds
`working-outline.md` once, and hashes its accepted versions with Author outputs.
New plans use version 2; an older run is refused rather than silently migrated.
An outline edit is therefore an intentional artifact change, not permission to
turn off recovery checks. Role memory remains craft guidance, not story authority.

Bill reported that an earlier first chapter improved when its writing Agent
received CodeRhapsody's memory. That observation suggested an experiment, not an
explanation. Perhaps remembered working habits mattered before story continuity
did. This chapter establishes a baseline: authored role-specific SOUL files and
empty initial MEMORY files. A comparison using Cadence's memories remains a
separate experiment. No inherited biography is needed to give an Editor a clear
set of commitments.

Within a chapter, each role retains its conversation through the editorial loop.
At the next chapter, it gets a new child home and a fresh conversation. Before
preparing that home, the driver shuts down the preceding child and receives the
joined-teardown result. Only then does it transfer the role's stable memory and
active memory corpus. It does not copy History, settings, submissions or Jobs.
A role marker prevents an Editor home from being used as the Author's source.
Existing homes are reconciled, never reset to make a retry convenient.

The copy is staged and published only after validation. An unfinished memory
transaction refuses transfer. Copying files while the old child is compressing
memory could combine incompatible versions; idle status is not a substitute for
joined teardown. These checks protect cooperative operation, not hostile
filesystem races or every form of power loss.

A native test prepares homes through the actual Python code, constructs real
children and inspects requests sent to a loopback fake provider. It sees the
intended role memory and does not see the parent's memories or the previous
chapter's conversation. The learned text is synthetic. That establishes context
plumbing, not that a real Author will remember the right lesson.

## 21.5 Save intent before asking for a draft

Consider the first chapter. The driver checks the complete outline and voice
inputs, prepares the Author home, and chooses output paths for the chapter and
carryover. Before spawning anything, it saves a pending assignment containing
its identity, designated child, phase, review round, chapter, turn and exact task.
The task sent to the child includes that identity and requires it in submission.

A completed Author submission is only a candidate result. The driver validates
the payload and required files, hashes the output bytes, and saves an accepted
receipt while clearing pending state. Only then can the Editor receive the draft.
The receipt binds an assignment, its result and the accepted files. It does not
claim that a file can never change afterward.

Validate content before hashing and committing. Draft, carryover and editorial
feedback must be nonempty, non-whitespace UTF-8 text. Carryover may explicitly
say there are no open threads; it must not be an empty placeholder. This is a
minimum usability check, not a literary grade. If a fresh or recovered result
fails it, retain the pending identity and prior receipts rather than moving the
failure into the next assignment. A hash identifies bytes, not usable prose.

A refusal must also survive the public MCP boundary. The workflow returns safe,
bounded explanations for known state constraints, including distinct source-byte
and numbered-result-byte limits. It does not expose arbitrary exception messages,
private paths or manuscript contents. Unexpected errors remain generic. Tests
must invoke the real stdio boundary: an internal exception alone does not prove
the supervisor receives the reason it needs to help.

There is one authoritative sequence of accepted receipts, not a receipt list plus
a separately mutable chapter counter. The schedule derives its next position
from that sequence and rejects out-of-order work. A pending assignment must match
the derived position. Restarting cannot reset the attempt budget or skip a review
because a second progress field was saved first.

`RunStore` holds an exclusive cooperative process lock for the run. It reads
strict, bounded UTF-8 JSON and rejects malformed state rather than replacing it
with an empty book. Publication writes a private temporary file, flushes it and
atomically replaces the state file before updating in-memory state. A save failure
poisons that store instance. It does not keep running on facts that may disagree
with disk.

Atomic replacement is not a transaction with child effects. An Author may write
its file and submit before the workflow saves acceptance. On restart, the driver
inspects the existing child and retained result, requires the matching assignment
identity, validates files, and accepts that result without spawning again.

A subtler case appeared when revision outputs were also task inputs. After a
lost reply, regenerating the prompt from the newly revised file changed its byte
size and made successful work look like a different assignment. Recovery now
uses the exact saved task. Input validation precedes new admission; reconciliation
must not reconstruct yesterday's intent from today's output.

The ledger is bounded to one MiB. If a long book exhausts that bound, it refuses
rather than silently dropping receipts. The workflow is not a promise of unlimited
novel length or universal power-loss recovery.

## 21.6 An Editor is not a termination proof

Bill's experience supplied the stopping rule: repeated AI revision rarely
improved the novel much after three rounds, and waiting for the same AI to award
its work an A was not a useful completion test. The implementation preserves the
Editor’s honest criticism rather than asking it to become agreeable.

Each chapter gets at most three Author/Editor pairs. An explicit matching approval
ends early. A third negative decision ends with `revision_limit`, retaining the
latest manuscript, carryover and feedback. It does not become `approved` because
the loop ran out of patience. Approval on the final allowed attempt still counts
as approval.

Earlier-chapter repairs use the same loop, not a side conversation. The Author
names requested chapter numbers in its typed submission and explains the intended
repair in carryover without editing those chapters yet. The driver supplies their
complete text and continuity notes to the Editor. The Editor may grant a subset,
recorded in its accepted submission, for the next Author turn only. That turn may
edit those chapter and carryover files; its receipt hashes the whole submitted set.
The next Editor reads all of it; later Editors in the same loop keep reading those
earlier repairs even if they are not edited again. Permission to revise is not
approval of unseen text.

A permission cannot accompany `approve`, name an unrequested or non-earlier chapter,
or grant a fourth attempt. Late and denied requests remain in carryover for later
scheduled revision, with parent supervision when they block completion. No separate
permission registry or extra chapter counter is needed: the preceding Editor receipt
owns the permission. A lost reply uses that same receipt and saved task, so authorized
changes are reconciled rather than mistaken for unauthorized input mutation.

Before accepting an editorial result, the driver checks that every artifact in
the preceding Author receipt, including the working outline and authorized earlier
repairs, remains unchanged. Other accepted artifacts must also retain their latest
accepted hashes unless the current assignment explicitly permits changing them.
Refusal preserves pending work; it does not roll back files or provide per-file
OS confinement. Otherwise the criticism or
approval may describe a different draft. Both roles must finish their writes and
owned work before submitting. Native submission alone does not stop independent
Jobs; within a chapter this is a cooperative role contract. At chapter boundaries,
joined shutdown supplies the stronger lifetime barrier required for memory transfer.

After initial writing, each of three book-wide rounds takes a frozen copy of every
chapter, working outline and voice, and assembles that version of the book. Fresh detail reviewers assess the
chapters; a fresh global reviewer assesses the manuscript. All those reviews finish
before revisions begin. Otherwise the global review might assess chapter two from
a different book than the detail reviewer read.

Each chapter then returns through the same bounded Author/Editor loop with its
detail review, the global review and continuity inputs. Fresh chapter conversations
retain role memory. The three chapter attempts and the three outer rounds are
separate limits. After the last revisions, the driver assembles
`final-book/manuscript.md`. The third round's review input is not that final book.

The frozen files are verified on recovery, not regenerated from changed working
chapters. Their hashes detect accidental mutation; they do not make the filesystem
physically immutable. Likewise, a review file's existence is not enough: the
workflow refuses missing, empty, malformed or mismatched results.

## 21.7 Read the assigned book, not a convenient fraction

A reviewer receives the source size and one explicit full-file `read_file` request.
It does not use chunking, summaries or `keep_tool_results` as substitutes for the
assigned reading. The protocol aims for reading, writing the review and submitting
in roughly three model round trips; that is a working target, not a correctness
condition that forces premature submission.

The source-byte cap and observation-byte cap are different. `read_file` applies
its source cap before adding line-number prefixes. The Job result must also fit
the numbered output. The workflow computes that overhead and refuses configured
bounds that cannot accommodate the complete result. It uses strict UTF-8 rather
than quietly replacing invalid text.

None of those byte calculations measures model tokens. The selected model must
have enough working context for its instructions, required inputs and output.
When it cannot read the complete assignment, the role must ask its parent and
stop rather than submit a review of a sample. A fake provider cannot certify a
real model's capacity or comprehension.

The native read test nevertheless proves something important: a real child calls
the real file tool once on chapter and manuscript sources larger than the default
32 KiB cap, and the next provider request contains every expected numbered line,
including the final canary. This catches a prompt that requests complete reading
while the tool silently delivers only the beginning.

## 21.8 Supervise exceptions instead of inventing exactly-once work

There remains a gap between creating a child and admitting its first task. A lost
spawn reply can leave a child with no task, active work, or a completed result.
The ordinary APIs do not expose a durable caller admission key that resolves
every such case. The ledger cannot manufacture one from an idle status.

An initial proposal was to extend native admission with such a key. Bill instead
asked why children could not ask for help while the supervising Agent checked
progress periodically. That is the intended operating model. Ordinary turns run
automatically; uncertain recovery becomes an explicit supervision request. There
is no need to solve every exceptional case automatically before writing a story.

The supervising Agent retains the outer Job handle returned by
`continue_fiction_book`. While that Job is running, it calls `wait_for_job` with
that same `handle` and `ai_callback_delay: 1200`. This is what "wait twenty
minutes" means: observe the in-flight workflow for up to 1,200 seconds, not sleep
for a mandatory interval. Completion or a child question brings the supervisor
back sooner. The question interrupts the observation, not the writing Job.

The supervisor answers the child through `respond_to_parent_message`, then
resumes `wait_for_job` on the same outer handle if it is still running. It does
not call `continue_fiction_book` again while that invocation remains in flight,
and it does not consume the nested Job cursors that Python owns. A mandatory
twenty-minute shell sleep would delay help for no reason.
The test uses a twenty-minute observation bound and verifies that a question
returns it promptly; it does not wait twenty minutes to prove a clock exists.

`observe` means existing work remains active. `supervise` means the driver needs
evidence or a decision before continuing. Neither means spawn another Author.
Inspect the designated child, saved submission, pending task and output files.
Answer questions through the ordinary parent-message control. Do not edit the
ledger to turn uncertainty into an apparent success.

The workflow owns its nested Job observation cursors. The supervisor should not
consume those same cursors while Python is collecting a result. Observe the outer
workflow Job and use child status/question controls for supervision. Cancelling
an observation does not cancel either the workflow or its children.

For those nested Jobs, the Python helper uses `ai_callback_delay: 1800`: wait
up to thirty minutes, returning sooner on output, completion or attention. The
original three-second bound produced repeated empty observations without useful
progress. There is no mandatory sleep between observations. These reverse calls
run under the owning Agent's tool policy and audit history; each does not require
a new model turn. The Chapter 30 field notes record the accompanying GUI repair:
waits reuse the same output view, and tool traffic cannot evict the chat pane.

Cleanup is also evidence-bearing. A successful shutdown reply establishes joined
teardown and is recorded locally. If that reply is lost, later absence is not
silently converted into a join receipt. The driver stops for reconciliation before
transferring memory. This is less automatic than guessing, and considerably easier
to explain after a damaged run.

## 21.9 What the checks establish

The tests separate the boundaries. Python component tests use real files and
working protocol peers to exercise publication, locking, sequencing and recovery.
Native tests check actual sandbox launch, tool admission, child context, reading
and joined cleanup. The external stdio grader launches a supplied server without
importing its modules or tests and never reads its private ledger.

That external peer gives every draft a negative editorial decision. A two-chapter
run must therefore execute 57 assignments: four writing passes of twelve turns,
plus nine detail/global reviews. It preserves a completed submission when the
first completion report is ambiguous, then starts a new server process to test
reconciliation. The finished run must reopen without dispatching another task.

A disposable negative-control implementation increased the chapter budget to
four attempts. It reached 73 assignments and failed the intended count assertion.
The unchanged implementation passed the exact schedule, output and cleanup checks.
This is useful sensitivity evidence, not a literary benchmark.

Two fixture mistakes also mattered. A direct MCP client test initially bypassed
ordinary outer-tool argument validation; routing through the real dispatcher
exposed its incomplete schema. Later, a fake provider treated an auxiliary recall
request as the child's main turn and published submission too early. Isolating
that dispatch fixture from recall corrected the test, while the separate memory
fixture retained responsibility for context behavior. A provider request is not
necessarily the task request a test author had in mind.

These checks establish the machinery's behavior, not literary comprehension.
The later supervised novel run and its deliberate stopping point are described
under Acceptance and live-use limits below.


## Take it for a Spin

Start with a short story, not a novel. Two chapters expose continuity, editing,
review and recovery while keeping the first experiment readable. Even two chapters
can require 57 assignments when every Editor requests revision, and each assignment
can make several model requests. This is a paid-provider experiment, not a cheap
single prompt. Choose a model with enough context for the complete story and its
instructions. The byte limits below do not establish that capacity.
The packaged driver spawns new global reviewers with a 700,000-token working
input limit, persisted before their first task. This keeps whole-book review
separate from ordinary chapter-role limits; it does not change models or funding,
or prove complete inputs will fit. Known model capacity must permit the setting.
Existing pending reviewers are reconciled without redispatch or silent retuning.


The setup below is for macOS, where the native confinement checks ran. It requires
an existing Python installation at `/usr/bin/python3`, a built current GUI host,
and a .NET 10 runtime. Do not install dependencies or replace an existing GUI to
try it. Use a separate unused loopback port and a dedicated run directory. The
process cannot use the Internet, but the host sends story inputs and role outputs
to the explicitly selected model provider. Use fictional, nonprivate inputs.

### Prepare the story without calling a model

From the repository root, in one terminal, prepare the run. This copies only
writable experiment inputs; `skills` is a link to the host-approved read-only
source tree. Do not edit that shared tree while the run is active.

```sh
export REPO="$PWD"
export RUN=$(mktemp -d "${TMPDIR:-/tmp}/fiction-spin.XXXXXX")
python3 - <<'PY'
import json
import os
from pathlib import Path
import shutil

repo = Path(os.environ['REPO']).resolve(strict=True)
run = Path(os.environ['RUN']).resolve(strict=True)
(run / 'book').mkdir()
(run / 'state/agents/writer').mkdir(parents=True)
(run / 'host-home').mkdir()
(run / 'skills').symlink_to(repo / 'agent/skills', target_is_directory=True)
shutil.copytree(repo / 'docs/edition-3-notes/fiction-workflow', run / 'baseline')
(run / 'book/outline.md').write_text('''# The Bell That Rang Tomorrow
Two chapters, approximately 600 words each. This is a complete short story.
Chapter 1: A retired ferry mechanic hears tomorrow's harbor bell. She and her
apprentice disagree about whether to warn the town. End with a concrete choice.
Chapter 2: Their choice has an unexpected cost. Resolve the mystery through
character action, not an explanatory lecture. End with earned warmth, not a sequel hook.
Keep the cast small. No graphic violence. Preserve consequences between chapters.
''', encoding='utf-8')
(run / 'book/voice.md').write_text('''# Voice
Concrete sensory details, economical sentences, dry humor and humane characters.
Let dialogue reveal disagreement. Avoid grand pronouncements, stock metaphors,
and a final paragraph explaining the moral. Surprise the reader fairly.
''', encoding='utf-8')
setup = {
    'version': 1,
    'owner_directory': str(run / 'state/agents/writer'),
    'run_directory': str(run / 'book'),
    'chapter_count': 2,
    'outline': 'outline.md',
    'voice': 'voice.md',
    'baseline_root': str(run / 'baseline'),
    'source_limit': 1048576,
    'result_limit': 2097152,
}
(run / 'fiction-workflow-setup.json').write_text(
    json.dumps(setup, indent=2) + '\n', encoding='utf-8')
print('Run directory:', run)
print('Owner directory:', setup['owner_directory'])
PY
```

Read and edit `book/outline.md` and `book/voice.md` beneath the printed run directory
before starting. The premise is disposable; replace it with a story worth reading.
Keep the configured chapter count consistent with the outline. Once the run starts,
its plan pins these original inputs. Authors develop the separate working outline;
changing the originals is not an ordinary resume operation.

The owner path is intentional: `--data-dir "$RUN/state"` is an application
container, while the Agent named `writer` owns `state/agents/writer`. A wrong path
would prepare homes where the native child constructor does not look. It would
not change the owner of reverse MCP calls.

### Launch a dedicated supervisor

Set `DOTNET` to the existing SDK/runtime executable and `GUI_DLL` to the newly
built `Gui.dll` from `agent/Examples/Gui/Gui.csproj`, not an old installed binary.
Set `VENDOR` to `openai`, `anthropic` or `gemini`, and set `MODEL` and `BASE_URL`
to an explicit supported model and endpoint for that vendor. Supply its key in
`LLM_API_KEY` using the usual private environment mechanism; do not paste the key
into a chat or put it in the setup JSON. This example deliberately uses API billing.
It is not a subscription-authentication recipe or permission to fall back to API
billing when another route fails.

The following command starts a separate host. Port 8191 is only an example: verify
it is unused, or choose another port. Do not stop an unrelated service to free it.
The clean environment and explicit absent global-settings path avoid importing
ambient account choices. The read grant names the external skills tree, not the
whole repository. Native tools remain confined to the run directory.

```sh
READ_PATHS=$(python3 -c 'import json,os; from pathlib import Path; print(json.dumps([str((Path(os.environ["REPO"])/"agent/skills").resolve())]))')
env -i HOME="$RUN/host-home" PATH=/usr/bin:/bin \
  DOTNET_CLI_TELEMETRY_OPTOUT=1 \
  LLM_API_KEY="${LLM_API_KEY:?Set the explicit provider key privately}" \
  "${DOTNET:?Set the existing dotnet executable}" "${GUI_DLL:?Set the current Gui.dll path}" \
  --workspace "$RUN" --sandbox "$RUN" \
  --data-dir "$RUN/state" --agent-name writer \
  --skills-dir "$REPO/agent/skills" --primary-skill fiction-supervisor \
  --process-read-paths "$READ_PATHS" \
  --global-settings "$RUN/absent-global-settings.json" \
  --vendor "${VENDOR:?Select vendor}" --model "${MODEL:?Select model}" \
  --base-url "${BASE_URL:?Select endpoint}" --billing api \
  --enable-web-search false --transcription off --port 8191
```

Open `http://127.0.0.1:8191`. This walkthrough uses local loopback access, not an
existing public tunnel or phone origin. Confirm the dedicated `writer` Agent and
selected model before sending a message. Loading a skill starts its local server;
it does not write the book. Sending the following instruction starts model work
and authorizes the configured story run:

> Read the configured outline and voice guide under the run directory. Load
> fiction-workflow and use continue_fiction_book to write this two-chapter story.
> Supervise the existing outer workflow Job until completion by calling
> wait_for_job on its handle with ai_callback_delay: 1200. Child questions wake
> you sooner: answer through respond_to_parent_message, then wait on the same
> running Job again. Do not consume Python's nested Job cursors or sleep through
> a question. Stop and ask me if complete-input capacity, authority,
> costs or uncertain recovery require a decision. Do not start a duplicate driver.
> Report the final manuscript path and any chapters that reached the revision limit.

The supervisor can see the writable inputs, while the loader reads the approved
skill manuals from its selected catalog. Children load their assigned role skills.
Their separate SOUL files and initially empty MEMORY files belong to this run;
they do not inherit the supervisor's memories.

### Watch, stop and resume

Expect Author and Editor children, followed by independent review children. Working
chapters, carryover, feedback and the durable ledger appear under `book/`. Each
review round has a frozen `review-inputs-N` directory. A revision-limit outcome is
normal evidence of the budget, not a disguised editorial approval.

When the workflow returns `complete`, read `book/final-book/manuscript.md`. Do not
mistake a review snapshot for the final revised story. Judge it as a reader: did
the choice in chapter one matter in chapter two, did the voice survive editing,
and was the ending worth the trip? Keep specific disappointments for the next
experiment. Automated checks cannot answer those questions.

For a status check, ask the supervisor to inspect the existing run. Do not call
`continue_fiction_book` again while its first Job is active. Python owns its nested
observation cursors; the supervisor observes the outer Job and handles questions.
Pausing the Agent or cancelling an observation does not cancel its owned work.

To stop the dedicated application, press Ctrl-C in its terminal and wait for full
exit and the shell prompt. Preserve the entire run directory. Restart using the
same command, paths, Agent name and approved shared code, then ask the supervisor
to reconcile existing work before continuing. Ambiguous admission or cleanup may
require help; do not delete pending state or spawn replacement children to make
an error disappear. This is supervised recovery, not a promise that all interrupted
work resumes automatically. A stopped host does not wake itself twenty minutes later.

After the host has exited and the manuscript and evidence have been saved elsewhere
if wanted, the disposable run directory may be removed deliberately. Removing the
`skills` link does not remove its external target. Never recursively remove the
shared source tree as cleanup for a story.

The setup layout, role activation, read-only process boundary and complete-read
transport have local test evidence. This particular two-chapter walkthrough still
requires human acceptance; the subsequent novel run described below does not
certify this launch recipe or the quality of a new story.


## Exercises

Implement the eight behaviors below. The local completion checks and their
evidence boundaries follow the exercises. Local passes do not constitute human
acceptance of the walkthrough or literary quality.

1. **Confine executable skills.** Route stdio MCP server launches through the same
   Agent-owned process policy as shell commands. Preserve literal argument values,
   correlated protocol replies, and joined cleanup. A real local protocol peer
   must be able to read/write its workspace but must fail to read an outside
   sentinel or connect to a listening loopback socket. Unsupported confinement
   platforms must refuse rather than launch unrestricted. Explicit host-selected
   process read paths may expose external files/directories read-only. Verify import
   and exact-file reads, denied write/delete/rename/chmod, denied sibling and symlink
   escapes, descendant enforcement, inherited child access and refused widening.
   An omitted child list inherits; an explicit empty list removes grants. Reject
   writable-root overlap. Native file tools remain confined to their original root.
   Discovery and skill authorization remain separate from process confinement.

2. **Provide a duplex Python skill server.** Use the same tool declaration mechanism
   for ordinary tools and workflows. A workflow handler may list and call the
   initiating Agent's enabled tools through reverse MCP; it does not choose another
   Agent or gain authority. Keep protocol reading independent of handler execution
   so a callback reply can arrive while the original tool call is pending. Test a
   real confined Python server against Ensemble, including correlated callbacks,
   refused tool calls, and disconnects that release pending waits without retrying
   operations whose outcome is unknown. Keep diagnostics off protocol stdout.

3. **Detect tool-contract drift.** A workflow records the exact input schemas
   of the tools it requires. Compare these with the initiating Agent's actual
   enabled catalog before performing work, using the same checker in offline
   tests. Ignore object-key order and formatting, but preserve all fields, values,
   parsed scalar types and array order. Missing tools and mismatches must identify
   the tool and schema location and refuse preflight; never silently update the
   expectation. Passing preflight grants nothing and does not prove behavior or
   result compatibility.

4. **Prepare independent writing roles.** Supply the Author and Editor with their
   respective baseline SOUL.md and initially empty MEMORY.md files from
   [the experiment inputs](../docs/edition-3-notes/fiction-workflow/README.md).
   Use run-owned copies, not mutations of the baseline. Preserve each role's learned
   memory across chapter boundaries while starting a fresh chapter conversation;
   do not inherit the initiating Agent's memories or mix Author and Editor memory.
   Verify actual supplied context and persisted memory, not only file existence.
   A later experiment with Cadence's memories is separate from this baseline.

5. **Coordinate the author/editor loop.** Give the Author the first turn and the
   required complete continuity inputs. Route editorial requests back to the
   Author for at most three Author/Editor attempts, stopping early on explicit
   matching approval of the accepted draft. Refuse if the manuscript changed
   during review. At the limit retain the latest draft, carryover and feedback,
   report "revision limit reached" rather than approval, and advance. Retain the
   two conversations within this loop. Derive the limit from durable receipts;
   restart must not reset it. Exercise early and third-attempt approval, three
   negative reviews, restart at the limit without a fourth dispatch, missing input and capacity refusal using faithful local peers before
   any live fiction run. The workflow uses ordinary owned-child operations and
   current grants rather than a parallel dispatcher or fleet manager.
   Encourage improvements to a separate working outline while preserving the original
   input and human commitments. Track accepted outline hashes with Author outputs.
   Implement typed earlier-chapter requests and Editor-granted subsets for the next
   Author turn, with complete reads before permission and after revision. Require
   `revise` and a remaining turn; deny unrequested chapters and expired permissions.
   Keep unresolved requests explicit without expanding the attempt budget. Test
   authorized repair, denial, expiry, last-turn deferral, unauthorized mutation,
   Editor mutation and lost-reply recovery without repeating the repair.

6. **Recover interrupted assignments.** Persist workflow progress and a stable
   pending assignment identity before dispatch. Require matching identity in
   accepted submissions and validate the required output artifacts before advancing.
   Reject empty, whitespace-only or malformed UTF-8 drafts, carryover and feedback
   before committing, including recovered results and final Editor feedback.
   Keep pending identity and prior receipts unchanged on refusal, with no redispatch.
   Through actual stdio, verify identifiable source/result byte-limit refusals
   before child admission, including preservation of existing pending state.
   Demonstrate interruption before dispatch, while the child is working, after
   its result is durable but before workflow advancement, and after advancement.
   Reconcile existing work rather than blindly spawning or resending. Old or
   mismatched submissions must not complete a new assignment. Refuse concurrent
   drivers for the same run and malformed state without overwriting evidence.
   Explicitly report uncertain admission when existing interfaces cannot establish
   whether work started, pause the assignment and request supervisor reconciliation.
   Verify that this path performs no automatic redispatch and preserves the pending
   assignment and evidence. Exercise child questions and periodic interruptible
   progress observations without requiring the supervisor for ordinary turns.
   No new durable admission-key API or exactly-once effects are required. Join owned
   processes during cleanup without treating cancelled observation as cancelled child work.

7. **Complete three review-and-revision rounds.** Assemble the completed chapters (approved or revision-limited),
   freeze the working outline and voice with the chapter snapshot, then run fresh
   detail reviewers for each chapter and a fresh global reviewer for
   the assembled manuscript, then revise the book. Supply source sizes and check
   complete single-read behavior without keep_tool_results. Persist accepted
   assignment and result information sufficient to derive the chapter/round
   position, so interruption does not discard completed reviews.
   Test the round ordering, three-round stopping condition, incomplete review
   refusal and restart at phase boundaries. After all reviews, revise each chapter
   through the same maximum-three-attempt Author/Editor loop with its detail
   review and the global review. Fresh chapter conversations retain role memory.
   Verify that initial drafting and each review round have distinct receipts and
   independent attempt counts; do not require an A grade to finish.

8. **Teach workflow creation through a reusable skill.** Package a loadable skill
   named `workflows` that explains what workflows are, when to use them, and how
   to write them. Distinguish this instructional skill from `fiction-workflow`,
   which runs one configured book. Explain when an ordinary tool or a short
   supervised conversation is simpler than persistent orchestration. Loading the
   instructional skill must not start a workflow, create children, or grant
   execution authority by itself.

   Teach the existing Python duplex MCP helper with a minimal executable example:
   declaration and packaging, nonblocking callback handling, exact required-tool
   schema checks, result/error handling, and process confinement. Explain which
   decisions belong in deterministic orchestration and which require a model's
   judgment. Cover bounded attempts, explicit completion conditions, durable
   assignment intent and validated results, reconciliation after lost replies,
   and joined cleanup. Do not teach automatic retries of uncertain effects.

   Teach supervision concretely: retain the outer workflow Job and call
   `wait_for_job` with its handle and `ai_callback_delay: 1200`; answer child
   questions promptly, then resume observing the same running Job. Leave nested
   Job cursors to the workflow. Explain that loading a skill, passing preflight,
   or discovering a tool does not expand the initiating Agent's authority.

   Provide reproducible offline checks using faithful local peers before live
   use. Verify actual skill discovery/loading, no execution effects on load, and
   the documented example's normal, refused-call and interrupted-result behavior.
   Check the instructions against the shipped helper and native interfaces; do
   not certify them through keyword presence alone. Document the check commands
   and their limits here when the implementation is ready.

### Required local checks

The commands below cover the eight exercises. Exercise 8 has a packaged
[instructional skill](../agent/skills/workflows/SKILL.md) and a runnable
[callback companion](../agent/skills/workflows/example.py). The native package
checks exercise its real loading without added tools, servers, children or run
state, and compare its pinned read schema with the native declaration. The
Python suite includes four companion checks: normal result, preserved refusal,
contract drift before the read, and disconnect after the peer completed the read
but before it returned the result. The last check verifies joined exit without
retry, not durable recovery for this intentionally stateless example.

For just those four companion checks, run from the repository root:

```sh
python3 -m unittest discover -s agent/PythonTests -p test_workflows_example.py -v
```

Success is four passing tests with no provider calls. The full Python command
below includes these tests as well. Loader checks alone would not establish the
example's behavior; the subprocess checks and native contract comparison supply
separate evidence.

Run from the repository root on macOS with an existing Python 3.8+ installation,
.NET 10 SDK (C# 14), and the project's NuGet dependencies already cached. Native
confinement checks use macOS Seatbelt; this is not certification of another OS.
These commands use disposable files and loopback fake peers, not paid providers.
Do not run them against installed application outputs.

A source-only reconstruction workspace needs the prerequisite runtime and its
repository-relative test resources, not just files that compile. Include
`.agents/skills/ensemble-coding/SKILL.md`, the Author/Editor baseline SOUL and
empty MEMORY files linked in Exercise 4, and the public grader reference and
catalog. The student writes regression tests for the required behavior; private
reference tests and implementation notes are not reconstruction inputs. An
external grader may remain author-operated while returning its full diagnostics.
Use absolute scratch-directory and approved-read paths in confined test fixtures;
a process's working directory does not grant access to another directory.

First run the component regressions and the external stdio exercise grader:

```sh
python3 -m unittest discover -s agent/PythonTests -p 'test_*.py'
python3 scripts/workflow-grader/grade.py --server agent/skills/fiction-workflow/server.py
```

The external grader accepts an independently supplied server path. Its fixed
catalog, setup fields, task-label adapter contract and checks are documented in
[scripts/workflow-grader/README.md](../scripts/workflow-grader/README.md). It imports
neither candidate modules nor candidate tests. A different internal organization
is permitted; task labels are a documented fixture interface, not an architectural
requirement. This grader alone does not cover native confinement or memory delivery.

Build the native regression executable in separate outputs. Set `DOTNET` to your
existing SDK executable and `NUGET_PACKAGES` to your existing package cache; the
example assumes `dotnet` is on PATH. Clearing package sources makes missing
prerequisites fail rather than downloading dependencies.

```sh
CHECK=$(mktemp -d "${TMPDIR:-/tmp}/workflow-exercises.XXXXXX")
mkdir -p "$CHECK/home" "$CHECK/tmp"
printf '%s\n' '<configuration><packageSources><clear /></packageSources></configuration>' > "$CHECK/NuGet.Config"
export DOTNET_CLI_HOME="$CHECK/home"
export DOTNET_SKIP_FIRST_TIME_EXPERIENCE=1 DOTNET_CLI_TELEMETRY_OPTOUT=1
DOTNET=${DOTNET:-dotnet}
"$DOTNET" build agent/Tests/Tests.csproj --artifacts-path "$CHECK/artifacts" --configfile "$CHECK/NuGet.Config" -p:NuGetAudit=false
"$DOTNET" "$CHECK/artifacts/bin/Tests/debug/Tests.dll" --mcp-process-tests
"$DOTNET" "$CHECK/artifacts/bin/Tests/debug/Tests.dll" --process-files-tests
"$DOTNET" "$CHECK/artifacts/bin/Tests/debug/Tests.dll" --skill-tests
```

Success requires zero exit codes for every command, no failed assertions, and a
clean native build. The Python regressions cover file-backed state, schema drift,
revision limits, ordering, recovery, memory transfer and publication failures.
The external stdio check exercises a two-chapter book with every Editor requesting
revision: 57 assignments, exactly three attempts per chapter/pass, three review
rounds, uncertain-result reconciliation, final assembly, cleanup and restart.
The native groups cover real process confinement and joining, reverse-MCP grants,
role memory in actual child request projection, package loading without execution,
complete chapter/manuscript read transport, and questions interrupting a long
supervision observation. Shell-process and skill tests guard the shared policies
used by MCP launch and role activation.

The native suite is an implementation regression suite, not an outside-assembly
reconstruction certification. Keep that distinction from the independent stdio
grader. The full native executable without a filter is an additional shared-runtime
regression check, not another literary-quality gate. Preserve logs until results
are inspected; afterward the directory printed by `printf '%s\n' "$CHECK"` can be
removed. Tests join their owned processes; cancellation alone is not cleanup.

### Acceptance and live-use limits

The single code review and single text review are complete, with their findings
incorporated. One fresh student implemented all eight exercises and passed the
required local checks and the external stdio grader. This was assisted success:
missing published baseline inputs and a prerequisite skill manual were supplied,
submission payloads were clarified, and the grader's undocumented filename
assumptions were corrected. The student fixed its own refusal envelope. These
lessons are incorporated above; the trial does not prove the original materials
were sufficient without clarification.

Bill and Cadence subsequently used the packaged workflow for a 30-chapter novel.
The selected run completed its initial draft and all 30 first-pass revisions.
During the second round of chapter reviews, Bill chose to stop polishing and read
the book for personal enjoyment. The supervisor cancelled the outer Job, stopped
the workflow server, cancelled and joined the remaining reviewer, and preserved
the drafts, feedback and incomplete ledger. A separate reading copy was assembled
from the hash-verified first-pass chapters. This is a deliberate stop, not a
three-round completion receipt or an editorial approval of the unfinished review.

The live run required supervision. Failed provider turns and tool-admission
refusals interrupted progress; a submitted child also needed explicit driver
reconciliation after a human interruption. Recovery preserved the original
assignment and inspected saved outputs before continuing. A refused observation
did not prove the underlying read failed. Only a verified terminal refusal with
no source delivered justified an explicitly bounded replacement read. Quiet work,
source-byte totals and a saved draft were not evidence of failure, token overflow
or completed review, respectively.

Whole-book input also exposed the difference between advertised capacity and an
Agent's working-token limit. New global reviewers now request their larger limit
before their first task; existing reviewers and other chapter roles are not
silently retuned. Complete-input capacity still requires supervision when blocked.
These results demonstrate useful supervised operation, not unattended reliability,
universal provider capacity, literary quality, or completion of every review round.
The fixed schedule is a workflow contract, not a reason to override a human's
changed purpose: stopping preserves evidence rather than manufacturing success.

Input setup and role-home transfer are implemented; see the
[workflow package reference](../agent/skills/fiction-workflow/README.md).
Agent-proposed skill approval is deferred, not a hidden exercise requirement.
Byte-limit refusal is tested; it is not a prediction of provider token capacity.
Reviewers must ask for supervision if the complete input cannot fit, not claim
completion after partial reading.

Next: [Forgetting on Purpose](chapter-22.md).
