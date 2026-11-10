# Chapter 21 workflow exercise checks

Run from the Ensemble repository root with an existing Python 3.8+ interpreter.
This grader uses only the standard library. It imports neither the supplied
implementation nor its tests. It starts the supplied server as a subprocess and
speaks newline-delimited MCP 2024-11-05 over stdin/stdout.

```sh
python3 scripts/workflow-grader/grade.py --server /absolute/path/to/skills/fiction-workflow/server.py
```

Exit zero and all seven PASS lines mean this **stdio workflow check** passed.
A failure is nonzero; setup errors are not successful negative tests. The candidate
must have its sibling `shared/ensemble_mcp.py` and its own packaged dependencies.
The grader's `contracts.json` is the fixed public Ensemble reverse-tool catalog;
it is not regenerated from the candidate. No packages, providers or network
connections are used. All book inputs and outputs are disposable temporary files;
the subprocess is joined before cleanup.

## Public fixture contract

The server takes one setup JSON path as its first argument. Version 1 setup has
`owner_directory`, `run_directory`, `baseline_root`, `chapter_count`, `outline`,
`voice`, `source_limit` and `result_limit`. Paths and the role baseline are supplied
by the harness; outline and voice paths are run-relative. Limits are bytes, not
model tokens. The server exposes argument-free `continue_fiction_book`, returning
MCP text containing JSON with `action` (`supervise` or `complete` at the boundaries
used here). Child calls use the ordinary Ensemble tool schemas and report forms.

For the deterministic fake writers, retain these task labels, or supply a small
candidate launch adapter that translates equivalent task instructions to them:

- `Workflow assignment_id: ID` on its own line; typed results echo that identity.
- `Load fiction-author.`, `Load fiction-editor.` or `Load fiction-reviewer`.
- `Chapter N of COUNT; review round R.` for writing turns.
- `Write chapter: PATH`, `Write continuity carryover: PATH`, and
  `Write feedback: PATH`, each on its own line when applicable.
- Author tasks include `Working outline: PATH` on its own line and
  `Editor-authorized earlier chapter numbers for THIS turn: JSON_ARRAY`.
- All writing inputs use `Read the complete input (N source bytes): JSON`,
  with complete `read_file` arguments. Earlier chapter and carryover paths must
  be included for requested permissions and for review after authorized repair.
- Review tasks include `with these arguments: JSON.` followed by newline for the
  complete `read_file` call, and `passage references to PATH. Finish the write`.

The fake's typed submission payload is `{"assignment_id": "ID"}` for Author
and Reviewer assignments. Editor submissions additionally contain
`"decision": "revise"`; the actual workflow must also accept `"approve"`.
The second, creative peer also submits Author `prior_chapter_requests: [1]`
while drafting chapter two, and Editor `prior_chapter_permissions: [1]` with
`revise` on its first attempt. It changes the working outline and, on the next
Author turn, repairs chapter one and its carryover. It checks complete Editor
inputs, original-outline preservation and the final revised text, and loses one
repair reply after saving it. The same 57-assignment bound applies; no extra
repair loop is authorized. These optional arrays must be supported; absent means
no request/permission. Local regressions additionally test denied/expired/late
permissions and unauthorized edits.

Declare compatible submission schemas; the fake does not invent additional
candidate-required fields. Files at the designated destinations carry the draft,
continuity, feedback and review content, not the submission payload.

Successful spawn reports contain two newline-separated JSON records: an
`agent_id`/`task_admitted: true` notice, then an `agent_id`/`task_completed: true`/
`submitted: true` record with `payload`. Continued turns return an `agent_id` and
`response`; retrieve their typed result with `get_submission`. Saved-result
reports contain `agent_id`, `historical: true` and `payload`. Historical status
is not current-task completion; validate the assignment identity and artifacts.
These are the deterministic fixture forms, not a replacement for handling native
running Job reports and subsequent observations.

These labels are a fixture interface, not a required internal Python architecture.
Review and working-chapter filenames are candidate choices. The fixture identifies
review inputs by their actual chapter content and validates final assembly against
the chapter destinations declared in the tasks, not a reference directory layout.
The fake writes the designated files, retains submissions independently, rejects
repeated assignment identities, and implements child creation, continued turns,
status and joined teardown. It never reads the candidate ledger. Every Editor
requests revision; no literary grade is manufactured.

The checks require refusal before work when tool contracts are absent; recovery
of a completed assignment after an ambiguous reply; exactly three author/editor
attempts for each chapter in each writing pass; all detail and global reviews
before that round's revisions; three outer rounds; final output assembled from
the last revisions rather than the last review input; joined children at completion;
and a second process reopening completion without redispatch. With two chapters
and no approvals this is exactly 57 assignments. Source read arguments must cover
the whole file and its numbered output. Actual model reading is tested separately.

## Other required checks and evidence boundaries

The chapter also requires the implementation's Python regression suite and the
native MCP/process tests. They establish cases this protocol fixture cannot:
real confinement, denied network/filesystem access, memory-only role transfer into
actual provider requests, complete single-call read transport, and a child question
waking an existing twenty-minute observation. See Chapter 21 for commands.

Neither this harness nor those local fake-provider tests establish literary
quality, a live provider's context capacity, automatic crash-free operation,
exactly-once external effects, or human acceptance. A real fiction run requires
separate authorization. Model instructions to read completely are not proof of
comprehension.

A bounded negative control was run on a disposable candidate copy: increasing
both chapter limits to four attempts reached the intended 73-versus-57 assertion.
The unchanged candidate passed; the reference implementation was not modified.
