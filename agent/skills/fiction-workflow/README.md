# Fiction workflow implementation in progress

The Python book driver has a packaged loadable MCP skill and a configured stdio
entry point. Native loading of this complete package is verified without starting
book work, including role dependencies and joined unload. The
duplex helper lives in `../shared/ensemble_mcp.py`; discovery alone grants nothing.

The driver owns one `RunStore` context for a run directory inside its sandbox.
`prepare` saves an assignment ID, designated child, phase/round/chapter/turn and
exact task before dispatch. A pending assignment refuses another preparation.
The driver reconciles it using ordinary child tools; the ledger does not infer
remote admission from silence, idle status or a missing reply.

`accept` requires a matching submission and driver-selected output paths. The
driver validates the phase-specific submission. Roles must finish their file work
before submitting; joined teardown is enforced before cross-chapter memory transfer. The ledger hashes the current files and saves the assignment,
submission and hashes together while clearing pending state. Phase/round/chapter
progress is retained in these assignments; `book_plan.next_step` derives and
validates the complete schedule from those receipts. No automatic retry or exactly-once guarantee is provided.

One cooperative Unix `flock` excludes other drivers using this store. The lock
file remains after close to avoid inode replacement races. State is strict UTF-8
JSON, version 1, with duplicate/unknown fields and invalid positions refused. The
state byte limit is 1 MiB; reaching it refuses rather than discarding receipts.
A private temporary file is flushed and atomically replaced before live state
changes. Save failures poison that store instance; recovery requires a new one.
This is not a transaction with child files or a guarantee against power loss.

Run paths must be relative and resolve inside the run directory. This is an
accidental-escape check, not a hostile-filesystem race defense. OS confinement
remains required. The driver must not accept files while a child is writing them.
Hashes record observed bytes, not immutable archival copies.

## Local checks

From the repository root, with an existing Unix Python 3.8+ installation:

```sh
python3 -m unittest discover -s agent/PythonTests -v
```

Tests use temporary real files, an independent lock contender, and real Python
stdio pipes. They make no provider calls. These are implementation tests, not
independent public chapter graders or proof of a live fiction run. Native
child-dispatch and role-context checks are described below; native full-book
integration remains pending.

## Chapter-local role homes

`role_homes.prepare_home` seeds a new child directory with the chosen baseline
SOUL.md and initially empty MEMORY.md by default. For an explicitly authorized
imported-memory experiment, each role baseline may contain `imported-memory.txt`
with exactly `imported-memory-v1` followed by a newline. That opt-in copies its
stable files and active memory bands on the first chapter too; keep donor records
attributed, snapshot them before launch, and never copy credentials, settings or
History. Role copies are independent. An unfinished memory transaction or malformed
marker refuses preparation. On subsequent chapters it transfers that
role's stable files and active `memory`, `bucket-0` and `bucket-1` corpus files.
It excludes saved conversation, settings, submissions, Jobs and retired memories.
A role marker prevents accidental Author/Editor mixing. Existing homes refuse;
restart must reconcile them, not reset them. Files are staged before publication.
The caller must hold the run lock and shut down the previous child before copying.
An unfinished memory transaction refuses. This does not grant access to a child
home outside the workflow process's sandbox, prove hostile-filesystem race safety,
or promise power-loss durability. Native corpus loading validates entry names.

Six real-filesystem Python tests cover baseline separation, memory-only transfer,
existing-home preservation, wrong-role/unfinished-transaction refusal, invalid UTF-8
and links, and nonempty initial memory. A C# integration test invokes the Python
preparer, constructs real children and checks actual loopback-provider requests.
It demonstrates role-memory input and a fresh conversation after chapter transfer.
Learned text is synthetic; the fixture explicitly publishes submission through
existing public APIs, so it does not test model-driven submission or fiction quality.

## Assignment dispatch

`assignments.dispatch` saves intent before calling ordinary owned-child tools.
New children use `spawn_sub_agent`; the adapter validates both its admission and
completion records. Continued chapter conversations use `ask_agent` followed by
`get_submission`. Both paths require the current assignment identity in the
submission. The returned payload is only a validation candidate: no artifact is
accepted and no pending assignment cleared here. Lost replies, malformed output
and cancellation never trigger redispatch. Resume through read-only reconciliation.

New global (whole-book) reviewers request `working_context_tokens: 700000` at
spawn. Native child creation validates this against known serving/advertised
capacity and persists the exact-route policy before admitting the first task.
Chapter reviewers, Authors and Editors omit this option and retain ordinary
settings. A pending reviewer is reconciled, not recreated or silently retuned;
its supervisor must resolve any existing capacity block. The limit is input
tokens, not a claim that source bytes fit or that reading was complete. Model
selection and funding are unchanged.

Role-home preparation, exact tool-schema preflight and sole-driver ownership are
caller responsibilities. Subsequent turns retain the original child submission
schema. The adapter bounds complete output to one MiB in both bare and wrapped
Job reports, refuses truncation and retains the same observation handle. Do not
concurrently consume that Job's output cursor from the supervising Agent.

Dispatch tests use a real ledger and a small child-protocol fake. They prove
save-before-dispatch, matching-result checks, split spawn parsing and refusal to
repeat an assignment after a lost reply. Native dispatch evidence is separate.

## Bounded chapter turns

`chapter_turns.advance_chapter` executes at most one author/editor turn or reconciles
its pending result. It derives the next turn from accepted ledger receipts rather
than a separate counter. Each role is spawned once per chapter, then receives
subsequent turns in the same conversation. The Author writes the driver-selected
manuscript and carryover; the Editor writes feedback and explicitly submits
`revise` or `approve`. Approval ends early; otherwise the third Author/Editor
pair ends with `revision_limit`. Retain the latest draft and criticism without
claiming approval. Repeated calls after either terminal outcome dispatch nothing.
The limit comes from receipts, so restart cannot reset the budget.

`review_round=0` selects initial writing; values 1 through 3 select later revision
passes with independent receipts and the same attempt limit. The caller prepares
fresh chapter-local children for each pass, retaining role memory. This operation
does not schedule detail/global reviews or ensure they precede revisions; that
remains the book driver's responsibility.

Before accepting an editorial result, the driver checks that manuscript and
carryover still match the preceding accepted Author hashes. An altered draft,
missing artifact or malformed submission preserves pending intent. The roles must
finish all file work before submitting; native submission alone does not join
independent Jobs. This is a cooperative writing protocol, not protection against
an adversarial Agent racing file writes. The caller owns complete-input prompts,
role preparation, skill/tool preflight, lifetime cleanup and later chapter transfer.

A working fake writes actual files, requests a revision and
then approves; reopening after each turn demonstrates receipt-derived progression.
Lost-reply recovery does not readmit the task. These checks are not native workflow
integration, full-input capacity enforcement, literary evaluation or public graders.


A further C# integration runs the confined Python assignment adapter through real
reverse MCP, spawning and continuing one real child. It verifies two accepted
receipts and no pending assignment. The loopback fixture publishes submissions
through the public API and disables only its synthetic child's auxiliary recall;
it is not a test of model-driven submission. Pinned schema preflight matches the
actual reverse catalog. Further native checks load the complete package, deliver
complete chapter/manuscript reads beyond the default 32 KiB cap, and wake a
twenty-minute observation for a child question. With process-read policy and
read-only mapped-package checks, nine focused native groups pass. These also
check the instructional workflows skill and its pinned native contract; none
establishes model comprehension or fiction quality.

## Complete book driver and stable review inputs

`book_driver.bind_plan` pins chapter count, outline/voice paths and hashes, role
baseline and actual child-home locations, and explicit transport-byte limits.
It refuses changed run inputs. `advance_book` performs one scheduled assignment;
the entry point continues ordinary progress and surfaces observation or
supervision outcomes. It is not a new scheduler. Role homes are prepared only
for new conversations. Completed obsolete children are shut down through ordinary
joined shutdown before memory transfer, review snapshotting or final assembly.
A lost shutdown reply requires supervision; missing child status is not a join.

`review_inputs.freeze_round` publishes ordered complete chapter copies and an
assembled manuscript under `review-inputs-N/`. Recovery verifies the manifest
and hashes without replacing them from mutable working chapters. All detail and
global reviews in a round use this version. After the third revision and joined
cleanup, `final-book/manuscript.md` is assembled separately: the last review input
is not the final revised book. Publication uses flushed staged files and a
rename under the run lock, not physical immutability or power-loss guarantees.

Pending writing assignments reuse their saved original task. Revision outputs
may also have been inputs, so regenerating a prompt from their new sizes would
misidentify successful work as changed intent. Matching submissions still require
artifact validation; reusing intent does not skip result checks.

Historical local evidence after code-review incorporation: 74 Python tests passed,
including the instructional workflow companion and refusal/artifact validation
cases. Nine focused native groups and three external stdio grader checks pass.
The earlier 390-pass full-native run predates these later additions; the focused
results are not a new full-suite certification. The whole-book fake writes real files,
runs two chapters through all three rounds, reopens the ledger after every step,
transfers role memory after joined shutdown and verifies final output. Tests also
cover changed frozen inputs, failed snapshot publication, lost revision replies,
and uncertain cleanup preserving evidence. These are not model-driven full-read
proof, native full-book integration or literary judgment. The external stdio
exercise grader in `scripts/workflow-grader` separately drives a supplied server
without importing its modules/tests. It verifies the full three-round schedule
with critical Editors, lost-result recovery and final joined cleanup.


## Packaged launch and configured run

The launcher uses literal relative paths from the sandbox root. Prepare this
layout before loading the skill; no command here installs or starts a live run:

- `skills/fiction-workflow/`: this directory, including its role subdirectories.
- `skills/shared/ensemble_mcp.py`: the shared helper beside the package.

`skills` may instead be a symlink to a trusted external skill directory explicitly
listed in the host's `--process-read-paths` JSON array. Select that same directory
with `--skills-dir`. The process policy permits reads and denies mutation there;
the symlink alone grants nothing. Native file tools still refuse external reads
and writes. Use a stable shared release: the host can change these files, so this
is not a snapshot or a content-signature guarantee. Do not update shared code while
a run is active. Omission inherits process grants in children; an explicit empty
list drops them, and child widening refuses. Writable-root overlap is invalid.

The rest of the run layout stays writable and inside the sandbox:

- The normal file/tool skill dependencies authorized by the selected primary.
- `fiction-workflow-setup.json`: explicit configuration below.
- Existing run, role-baseline, and owning Agent data directories inside the sandbox.

Use the actual owning Agent directory, not a guessed HOME or workspace. The server
has no privileged owner-discovery API; setup is operator configuration. Reverse
calls still derive the real caller from the connection, never this file. A wrong
owner path cannot select another reverse-call Agent, but would prepare the wrong
homes and must be corrected before running. Do not treat this path as authorization.

```json
{
  "version": 1,
  "owner_directory": "/sandbox/agent",
  "run_directory": "/sandbox/book",
  "chapter_count": 2,
  "outline": "outline.md",
  "voice": "voice.md",
  "baseline_root": "/sandbox/role-baselines",
  "source_limit": 1048576,
  "result_limit": 2097152
}
```

Replace example paths and chapter count deliberately. Outline and voice paths are
relative to the run directory. Baselines have `author/` and `editor/`, each with
SOUL.md and initially empty MEMORY.md. Limits count bytes, not model tokens or a
capacity guarantee. Python requires the existing Unix facilities used by the
ledger; unsupported OS confinement must refuse, not fall back to unrestricted work.

`server.py` captures this setup at launch. Its sole tool, `continue_fiction_book`,
takes no arguments. Before filesystem mutation it checks the static
`contracts.json` against the actual reverse catalog. Never regenerate that file
automatically to accept API drift. A changed setup on later launch is checked
against the pinned run plan. Ordinary progress continues in the same invocation;
`observe`, `supervise`, and `complete` return control. Cancellation releases the
run lock without claiming to retract remote work. Do not invoke a second driver
while the first still runs; observe its existing Job instead.

The real stdio test initializes the server, lists its tool, supplies an empty
reverse catalog, verifies a correlated contract refusal before ledger creation,
and joins clean EOF shutdown. In-process tests complete a configured synthetic
book, reopen without redispatch, reject model-supplied owner arguments, and surface
a lost result. These do not replace native full-package or model-read checks.

## Creative authorship and Editor-authorized repairs

New run plans use version 2. Original outline/voice hashes stay pinned; the driver
seeds `working-outline.md` once. Each Author may improve that plan while respecting
the brief, voice, chapter count and established manuscript facts. The working
outline is an accepted Author artifact, not an unchecked mutable configuration.
Existing version-1 plans refuse; there is no silent migration or restart.

Authors may submit optional `prior_chapter_requests` integer arrays, with scope
and reasons in carryover. Editors receive complete requested chapters/carryovers
and may grant a subset through optional `prior_chapter_permissions`, only with
`revise` and a remaining Author turn. The permission belongs to that accepted
Editor receipt and expires after the next Author turn. Its chapter/carryover paths
become accepted outputs; the next Editor reads the entire set and working outline.
Approval covers those actual bytes. Denied or late requests remain documented for
later scheduled revision or parent supervision, never an automatic fourth attempt.

At assignment boundaries, latest accepted hashes protect all other artifacts.
Pending authorized outputs may have changed after a lost reply; reconciliation
uses the saved task and same permission, without redispatch. Violations preserve
pending state for supervision and do not roll back files. This cooperative check
is not an OS permission boundary. Frozen review rounds include their own outline
and voice copies; final-book also contains those context copies.

The creative-change Python tests cover permissions, expiry, last-turn refusal,
outline mutation and frozen context, Editor mutation, and lost-reply recovery.
The public stdio grader now runs both ordinary and creative peers, seven PASS
lines in total. Neither test proves literary quality or actual model compliance.

Latest creative-revision verification: 84 Python tests, nine native MCP/package
and process groups using the existing built runner with the updated package, and
seven external PASS lines across ordinary and creative stdio runs. No C# changes,
installation, live-provider writing or new student trial occurred in this edit.
