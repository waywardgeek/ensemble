# Chapter 3: Six Tools, One Turn

A conversation can describe a fix without changing a byte. This chapter
lets the model read the file, make the edit, and run a command to check it.
That makes the distinction between an answer and an action matter: a
confident sentence is no evidence that the file changed.

The useful unit is a completed turn. A model can ask for a file, inspect
the result, ask for another file, and only then answer the human. The
program must keep that exchange moving without losing a result, confusing
two calls to the same tool, or forgetting what already happened when the
next request fails.

## 3.1 Why six tools

Bill's [September 2026 tool-use exhibit](../exhibit-ch03-tools.md) records
a steep concentration of work in running commands, reading, editing,
searching, and writing files. It is a historical account of his own archived
sessions, not a measurement of every engineer or today's models. The
chapter uses it to choose a small working set, then tests whether each
tool is useful through the user interface.

A shell can perform all these operations. Dedicated tools still give the
model better ways to express what it wants. A range read limits the text
entering the conversation. An edit names an exact old fragment and refuses
an ambiguous match. A write distinguishes creating a file from replacing
one. Their arguments and failures are easier to inspect than a shell
command containing several layers of quoting.

Separate tools also let one Agent read without granting it a shell or a
writer. Hiding a tool's declaration is insufficient if dispatch will still
execute its name. The same Agent-visible set must govern both what the
model is offered and what the program will run.

The six tools are `read_file`, `list_directory`, `search_files`,
`write_file`, `edit_file`, and `run_command`. All run synchronously in
this chapter. Jobs arrive in Chapter 4; they extend the work recorded here
with a lifecycle that can outlast one dispatch.

## 3.2 Put the tools under their Agent

An Agent owns its Registry. This chapter chooses per-Agent storage because
the first requirement is independent capability sets, and no shared registry
is needed to provide them. A later implementation may share storage under
Ensemble while preserving per-Agent visibility. The permission rule remains
the same in either arrangement.

The Registry is an `internal/tools` service created for its owning Agent.
Its stored parent is an Agent interface declared in `internal/common`.
Engine reaches the service through its own Agent parent; `internal/llm`
does not import `internal/tools`. Shared declarations, call/result values,
and interfaces stay in common, while tool execution and argument handling
stay in tools. Use free functions over shared data where the package
boundary requires them.

The Registry reaches workspace configuration through Agent and the logger
through Agent's Ensemble parent. A file helper needs that route too: a
short function can still fail. Do not turn the old standalone tool
function into a closure capturing a separate logger and working directory.
If the implementation introduces a live Call object, Engine owns it; its
parent chain must describe that relationship. Plain recorded call data is
still a value, not another service to wire.

Declare and dispatch from one authoritative set. Each entry supplies a
name, a description written for the model, a JSON argument schema, and
the implementation it permits. Build declarations in stable name order
and take owned snapshots for the request. An empty Registry emits no
`tools` field. A guessed or disabled name produces an error result even
if the implementation exists elsewhere in the process.

The public library lets its caller choose the built-in subset when creating
an Agent. An omitted selection keeps a headless Agent without tools; the
CLI explicitly installs all six. Changing a tool set during an active turn
is unsupported, and no dynamic registration API is required here. Chapter
2's declaration type remains the renderer's input, now derived from the
Registry for live requests. Offline rendering can still accept an explicit
declaration snapshot. It must never dispatch a tool.

The Chapter 2 public declaration input does not become a second authority
for live tools. Reject a live configuration that supplies declarations
inconsistent with its selected built-ins before making a request. Arbitrary
declarations remain useful to offline renderers; a live Agent derives the
names and schemas it advertises from the Registry it will actually use.

## TL;DR

Continue the validated Chapter 2 repository and history in
`solutions/edition-2/ch03`. Before editing, read the entire
[`book/edition-2/skills/ensemble-coding/SKILL.md`](skills/ensemble-coding/SKILL.md),
the current architecture decisions, and this chapter's detailed contract.

1. Add the Agent-owned Registry and six tools specified below. Declarations
   and dispatch use the same per-Agent set. Preserve star imports, owner
   back-pointers, logger access, public clients, and the optional GUI module.
2. A user turn may make up to sixteen model requests, including its first
   request. Persist every accepted response and usage once. Execute all
   calls in their original order, append their results, then ask the model
   again until a response contains no calls.
3. Persist `tool_called` before attempting a tool and `tool_returned` after
   it finishes. Preserve call IDs. Ordinary tool errors still return a
   matching result, and later calls in the batch still run. Failure to
   persist faults the Agent; it cannot retry or claim rollback of a side
   effect that already happened.
4. Declare the visible tools on every model request. Render results and
   error status on all three surfaces; preserve bound signatures. Complete
   a batch before its continuation request, with results preceding deferred
   human text under Chapter 2's published projection rule.
5. Capture an absolute workspace on Agent at construction. Resolve relative
   tool paths against it and accept absolute paths. Never change the process
   working directory. This path rule is not a filesystem sandbox.
6. Use exact unique-anchor edits and explicit overwrite permission. Bound
   read, search, listing, and captured command output; report truncation.
   A command's nonzero exit status is a result of execution, not a malformed
   tool call. Report stdout, stderr, and the actual process exit status.
7. Preserve the JSON-lines CLI, `render LOG`, `dump`, and `CH02_LOG`.
   Emit one assistant line containing the final model response per completed
   user turn. Intermediate text and tool facts stay in history and observer
   events. Sum usage across all accepted responses, not just the final one.
8. If request sixteen still has calls, complete that batch, then report the
   round limit before request seventeen. Transport, parse, persistence,
   and round-limit errors end the CLI session without a fabricated final
   answer or success-usage line. Keep already accepted history and effects.
9. Demonstrate every tool and its important failure paths through the actual
   CLI/public library on the three real APIs. Separate deterministic fault
   fixtures from live evidence. Compare the initial answer with the first
   edition only after the initial build and run, then revise code and prose.

**Yours.** Internal names, implementation structure within the architecture,
tool description wording, and presentation of ordinary tool output. The
argument names, semantic defaults, result/error distinctions, and lifecycle
rules below are the public contract. No background jobs, process containment,
streaming, mailbox, concurrent tool execution, or automatic retries yet.

**Exercise.** From the course repository root:

```sh
make grade-dir CH=3 DIR=solutions/edition-2/ch03
```

The inherited score is a regression baseline. The additional acceptance
table in §3.9 covers requirements that the original grader never protected.

## 3.3 A call is the middle of an answer

Suppose the model asks to read two regions of the same file. Both calls
have the name `read_file`; the name cannot tell their results apart.
Their IDs can. Keep each issued ID attached to its arguments, dispatch
record, result, and outgoing provider representation.

After a valid response has been appended and applied, inspect its ordered
parts. Preserve text and opaque material as well as calls. For each call,
append `tool_called` with that ID, name, and arguments. Once that write
succeeds, ask the Agent's Registry to execute it. Append one
`tool_returned`, including its result parts and `is_error` flag, through
the same Agent event path used everywhere else.

An unavailable tool, missing file, bad argument type, or refused edit is
an ordinary tool failure. Return an error result naming the operation and
explaining what the model can correct. Do not omit the result or end the
batch merely because its first call failed. A later read in that batch
must still run, and the next model request must carry both results.

Tool argument validation and response parsing are different boundaries.
Valid JSON such as `{"path":42}` reaches the tool and produces a correctable
argument error. A broken response envelope is an API parse failure. Under
Chapter 2's contract, a Chat Completions argument string that does not
decode to a JSON object also fails response parsing; a JSON string can
contain malformed argument text while the outer response remains valid
JSON. Do not call that impossible merely because other surfaces send an
argument object directly.

Call the model again only after the whole accepted batch has results.
That continuation introduces no new human message. It uses the same
configuration, declarations, full projected history, and usage accounting
as an ordinary request. The Engine owns the turn's request count; replay
does not restart the tool loop or execute any recorded action.

A response without calls completes the turn. Its text, including a valid
empty text part, becomes the CLI answer. Earlier narration remains in the
log and public observations, where a client can display it without adding
extra assistant lines to the CLI protocol. A tool's output is data for the
model; it cannot register another tool or increase the Agent's permissions.

## 3.4 Three ways a turn can fail

The tool may fail while the conversation continues. A failed file lookup
answers the model's question with an error. The next response can repair
the path. This is why `is_error` belongs on the result, rather than being
an event that aborts the entire turn.

The model request may fail after tools have already changed files. Retain
the accepted responses, tool records, effects, and their usage. Append a
safe `error_occurred` when the log remains writable, return the request
error, and end that CLI session. Chapter 1's atomic conversation pair does
not authorize deleting an already accepted tool exchange or pretending a
successful edit was undone. Chapter 2 already distinguishes a pending
unaccepted prompt from committed history; keep that distinction here.

The log may fail at a side-effect boundary. If `tool_called` cannot be
persisted, do not execute the tool. If the operation ran but its result
cannot be persisted, report that completion could not be recorded and
fault the Agent for future writes. The file may already have changed.
Neither a retry nor a rollback is implied. Subsequent operations require
a fresh Agent/log, and the application must inspect the actual effect
before deciding what to do next.

The round bound is separate from all three failures. Count model requests
from one through sixteen for each human turn. If response sixteen has
calls, record and execute the complete batch, preserving its pairs, then
append an error with code `round_limit` and stop before another model
request. If it has no calls, finish normally. A round bound does not
interrupt a running shell command, and this chapter makes no promise of
process cancellation or a background handle.

## 3.5 The six contracts

All argument payloads are JSON objects. Require the stated types; reject
missing required fields, unknown fields, invalid regular expressions or
glob syntax, negative ranges, and nonpositive explicit limits. Report the
tool name and the offending field. Use the Registry's declaration metadata
when explaining arguments so a second hand-maintained argument summary
cannot forget a newly added flag.

Schemas use object properties, required fields, primitive types,
`additionalProperties:false`, and clear descriptions. Declare defaults
and refusal behavior in the descriptions; the model reads them before making
the call. Empty `content` and `new_text`
are valid strings. Empty paths, commands, patterns, and `old_text` are not.
Numbers naming lines or byte limits are integers.

| Tool | Required arguments | Optional arguments and defaults |
|---|---|---|
| `read_file` | `path`: string | `start_line`: 1; `end_line`: 0 means through EOF; `max_bytes`: 65536 |
| `list_directory` | none | `path`: `.`; `max_entries`: 200; `max_bytes`: 65536 |
| `search_files` | `pattern`: string | `path`: `.`; `file_pattern`: no filter; `context_lines`: 0; `max_matches`: 200; `max_bytes`: 65536 |
| `write_file` | `path`: string, `content`: string | `append`: false; `overwrite`: false |
| `edit_file` | `path`: string, `old_text`: string, `new_text`: string | none |
| `run_command` | `command`: string | `max_output_bytes`: 65536 per stream |

The workspace is captured on Agent when it is created, defaulting to the
application's current directory at that moment. A relative argument resolves
against that directory. An absolute argument names that absolute path.
`run_command` starts its shell there using the child process's directory
setting. Two Agents with different workspaces must not change one another's
path interpretation. These tools do not confine access: a shell command or
an absolute path can reach outside the workspace. Use a scratch directory
for demonstrations; a workspace setting is not a security boundary.

### Read enough to act

`read_file` returns the selected one-based inclusive line range, preserving
the selected text and line breaks up to its byte cap. A trailing newline
ends a line instead of creating an extra empty line. An omitted end reads
through EOF; explicit end zero has the same meaning. Reject start below
one, a positive end before start, or a start beyond EOF. An end beyond EOF
stops at EOF. Reading an empty
file with the default full-file range succeeds with empty content; an
explicit range into an empty file fails. The tool must distinguish that
valid empty result from an I/O error.

Truncation must be visible in the result, outside the retained-content
byte budget. Report it when content was omitted, not merely when a complete
result exactly fills its limit. State which limit was reached so the model
can request a narrower range or an explicit larger limit. Numbered lines or a short
header are acceptable; returning an entire file for a requested two-line
range is not. A read failure returns a tool error with the path and reason.

`list_directory` is one level deep, sorted by entry name. Return names and
distinguish directories, for example with a trailing slash. Respect the
entry and byte limits and say when the listing was truncated. A missing
path or a path that is not a directory is a tool error.

`search_files` recursively searches regular text files with Go's regular
expression syntax. A file path searches that file; a directory path searches
below it. Skip `.git` directories and do not descend through symbolic links.
An optional `file_pattern` uses `filepath.Match` against each basename.
For this exercise, skip a file as binary if its first 8192 bytes contain a
NUL byte; search other selected regular files as text. This explicit rule
keeps classification reproducible without adding a format detector.
Traverse paths and emit line numbers in stable order. Unreadable selected
files are reported as errors rather than silently claiming a complete search.

With no context, a selected match is `path:line:text`. With context, a
neighboring line is `path-line-text`; merge overlapping or adjacent windows
within one file and print `--` between separated groups. Select at most
`max_matches` matching lines, then form their context windows. Within those
windows, selected matches keep the colon marker and other rows use the
hyphen marker. Report truncation only when a further match or output byte
was omitted; a complete result exactly at either limit is not truncated.
An empty search result is successful and says no matches;
an invalid pattern is an error the model can fix.

### Make the edit identify its target

`write_file` creates missing parent directories and a new file without
requiring `overwrite`. When `append` is true, append to an existing file
or create a new one; `overwrite` is irrelevant to that mode. Otherwise,
replacing an existing file requires `overwrite:true`. A refusal identifies
the target and its size and leaves its bytes unchanged. Use an exclusive
create when replacement is not authorized so a file appearing between
an existence check and creation cannot be silently overwritten.

`edit_file` reads an existing file and counts exact, non-overlapping matches
of the nonempty `old_text`. Exactly one match permits replacement. Zero or
multiple matches return an error with the observed match count, change
nothing, and tell the model to read the file and choose a unique anchor.
There is no fuzzy fallback and no implicit full-file rewrite. Empty
`new_text` deletes the unique fragment; replacing a fragment with itself
is a successful no-op and should be reported as such.

Validation and guard refusals leave target bytes unchanged. An actual I/O
failure is reported as an error; it must not claim that a partially completed
operation was rolled back unless the implementation actually provides that
guarantee. Successful write/edit reports describe the operation performed.
The acceptance checks inspect the file itself, because the tool's own
success message is not independent evidence of a write.

### Let the command report failure

`run_command` executes through a POSIX shell in the Agent's workspace and
waits for it to finish. The exercise requires an environment with a POSIX
shell. Capture stdout and stderr separately and return both with the actual
shell-process exit status. The presentation is yours, but a reader and the
grader must be able to identify the streams and status without guessing.

Exit status 7 is a successful tool invocation whose command failed. Set
`is_error:false` and report 7. Failure to start the shell or decode its
arguments is a tool error. A command that the shell cannot find still
produces the shell's exit status and diagnostic output; the shell did run.

Retain at most `max_output_bytes` from each stream and mark any truncation.
Continue draining after the cap so a child cannot block because its output
pipe filled. A capture limit is not permission to stop the process or
manufacture success. Long-running commands still block this chapter's
turn; Chapter 4 adds supervised jobs. No demonstration needs to start an
unbounded server merely to prove that limitation.

## 3.6 Send the results where the model can see them

Every continuation request carries the Agent's visible declarations.
Use the shapes from Chapter 2: Messages tool objects with `input_schema`,
Chat Completions function tools with `parameters`, and Gemini
`functionDeclarations` with `parametersJsonSchema`. The argument schema
remains the same owned document
at the neutral boundary. Do not insert an empty tools field for an empty
Registry, and do not send declarations only on the first request.

Results are rendered from recorded data. Messages uses `tool_result` with
`tool_use_id`, content, and `is_error`. Its answering user message begins
with the results before other text, as specified by the
[tool-call handling documentation](https://platform.claude.com/docs/en/agents-and-tools/tool-use/handle-tool-calls),
checked October 7, 2026. Preserve the order of results within the batch.
Chapter 2's mixed-content fixture makes this observable without introducing
concurrent input into the live CLI.

Chat Completions uses tool messages with `tool_call_id`. It has no parallel
`is_error` flag on that message, so an error result's content must explicitly
say that the tool failed and carry its actionable description. Gemini uses
`functionResponse` with the call ID and original function name. Successful
text goes under `response.result`; a failure uses `response.error`. The
[generateContent schema](https://ai.google.dev/api/generate-content) defines
that response as an object. Keep the neutral `is_error` fact in the log
regardless of how a surface spells it.

The model's call and any bound replay material stay together. Copying the
call's name and arguments while dropping its signature recreates a bug
Chapter 2 already prevented. Preserve requested and returned model identity,
apply the same exact-provenance compatibility rules, and keep usage from
every accepted response under the Engine that received it.

## 3.7 Fixtures that can expose the mistake

Plant this independent `notes.md` fixture for reading and searching. Do
not make a read test depend on `write_file` working first:

```text
alpha line one
beta line two
gamma line three
delta line four
epsilon line five
```

A model response with two calls to `read_file`, one for line 1 and one for
line 5, must produce two ID-matched results containing the corresponding
text and excluding the other. A range from 3 through 4 contains gamma and
delta and excludes alpha and epsilon. With `context_lines:1`, searching
for gamma produces these lines:

```text
notes.md-2-beta line two
notes.md:3:gamma line three
notes.md-4-delta line four
```

Searching for `beta|delta` with context one yields a single merged window;
searching for `alpha|epsilon` with context one yields two separated windows.
Match the path spelling to the requested search root; a leading `./` is
acceptable. Independently test the absence of a match, a bad regex, the
basename filter, and a cap small enough to require a truncation notice.

For edits, plant a file containing `first anchor\nsecond anchor\n` and
try `old_text:"anchor"`. Refuse its two matches and verify unchanged
bytes. Then replace `old_text:"first anchor"` with `new_text:"first edit"`
and verify the other line remains. A missing anchor is a separate failure
control. New-file write, refused replacement, explicit replacement, and
append each need their own disk observation.

For exit status, run the silent command `exit 7`. There is no stderr text
from which a broken implementation can accidentally obtain the right
number. Pair it with a success such as `go version` and a command that
emits distinct stdout/stderr markers. A wrapper such as `go run` can
print a child's exit status while returning a different status itself;
that output is evidence about the wrapper, not a substitute for reading
the process status.

Use the deferred-human fixture printed in Chapter 2 to test result-first
rendering. Its call, later human input, and result must all survive replay.
Also verify that loading, dumping, and offline rendering a log containing
`tool_called` never executes it. A planted sentinel file can expose an
accidental side effect during replay.

## 3.8 Fakes need witnesses too

The original tool loop passed its fake without declaring any tools. The
fake volunteered calls, so the program could exercise dispatch without
ever telling a real model which functions existed. Commit `6f4b4c1`
records the added declaration checks and their deletion controls. A
fixture that supplies a call is not evidence that a model can discover it.

Build deterministic behavior against the fake, then probe a real API when
a wire question remains. Turn the observed answer into a repeatable
fixture. Keep the directions separate: a fake can send an unrealistically
helpful response, or accept a request that the real API refuses. Either
can leave the tests green while the user cannot finish the task.

The same discipline applies to local tools. The first edition's overwrite
guard needed its own check; a successful new-file write said nothing about
an unauthorized replacement. Commit `554d1cb` records that correction.
Deleting the guard must lose its intended check, while refusing every
write must also fail the valid new-file control.

Do not treat a stale mutation as a successful audit. Confirm the changed
behavior actually landed before interpreting the score, and compare the
exact expected failure set with the observed one. The chapter's tests
must be able to distinguish a wrong result from an unrelated feature
that failed to create their input.

## 3.9 Exercise and acceptance

Keep the previous command and environment contracts, including explicit
model selection and `CH02_LOG`. No additional CLI mode is needed. In a
public consumer, the same Ensemble request service runs the loop, and
the same observer receives its persisted tool facts with Agent ID and
sequence. Tools have no GUI imports or direct GUI calls. The separate
optional GUI stub continues to build through public interfaces.

The inherited checks total 100 points:

| Check | Points | Protected behavior |
|---|---:|---|
| `ch2parity` | 10 | Prior log, reducer, renderers, accounting, and CLI checks remain passing |
| `toolsdecl` | 5 | All six CLI tools are declared correctly on every request |
| `toolloop` | 20 | Calls return ID-matched results and continuation reaches a final answer |
| `multiblock` | 10 | Two calls to the same tool both run with their own arguments and IDs |
| `readtools` | 10 | Ranged read, directory listing, and search operate on independent fixtures |
| `mutatetools` | 5 | Writes and edits change actual disk bytes |
| `writeguard` | 5 | New file succeeds, unflagged replacement refuses, explicit replacement succeeds |
| `runcommand` | 15 | Shell runs and reports stdout, stderr, and actual exit status |
| `toolerror` | 15 | Ordinary tool failures return to the model and the turn continues |
| `editcontract` | 5 | Edit report agrees with the observed file state |

The old `editcontract` accepted several coherent edit policies. This
edition specifies exact unique-anchor replacement, so its stronger
requirement needs additional checks. Preserve the old regression coverage
instead of claiming the broad old check already proves the narrower rule.

| Additional property | Evidence required |
|---|---|
| Ownership and reachability | Actual tools-spoke imports and Agent parent; helper logger access; no mutable global Registry or duplicate workspace owner |
| Capability selection | Two Agents with different visible sets; no declarations for empty set; guessed/disabled call returns error without side effect |
| Workspace isolation | Two Agents resolve the same relative name in separate directories; one never changes process cwd for the other |
| Durable execution | Pre-call append failure prevents execution; post-effect result failure faults Agent and does not repeat the action |
| Loop lifecycle | All calls in a batch run after ordinary errors; final batch completes at the sixteen-request limit; no request seventeen; mid-loop API failure retains accepted history and usage |
| Edit and write edges | Missing/ambiguous/empty anchor refusal, unique replacement, valid deletion/no-op, empty file, append, and overwrite controls |
| Bounded content | Read ranges, listing/search limits, merged context, invalid arguments, command-stream draining and visible truncation |
| Rendering and replay | All three error/result surfaces, bound signatures, result-first mixed content, deterministic replay with zero side effects |
| Public clients | CLI final-answer protocol, observer attribution/order, headless consumer and optional GUI stub remain usable |
| Real use | All six tools and recovery paths exercised through the actual program on all three provider paths, with disk and request/log evidence |

After the first run, compare the new answer with the first-edition standard.
Look for duplicated schema metadata, awkward dispatch, missing diagnostic
context, and comments that explain syntax instead of the invariant. Retain
the initial answer and the findings, revise code and teaching, rerun affected
checks, and obtain review of the revision. Passing is the start of that
comparison, not its conclusion.

## 3.10 Taking it for a spin

Build the new CLI and run it in a fresh scratch workspace with a fresh
explicit log path. Supply credentials through the environment and select
an actually available tool-capable model using current provider information.
If replay material requires a resolved model identity, configure its known
value through Chapter 2's public configuration or `LLM_RESOLVED_MODEL`
before the turn. Automatic continuation obeys the same exact-provenance
rule; it must not guess an alias after tools have already changed files.
Repeat the feature checklist on Messages, Chat Completions, and
generateContent; record selected and returned model IDs without assuming
the old book's defaults are still available.

Ask the model to list the directory, create a small text file, read a range,
search for a word with context, replace a unique fragment, and run a short
command that verifies the final bytes. Inspect the file independently.
Then ask a follow-up that depends on an earlier answer so a tool session
cannot conceal lost conversation history.

Exercise the overwrite refusal and a failed edit, then let the model recover
with explicit overwrite permission or a corrected unique anchor. The record
must show the failing result reaching the model and the later successful
operation. The model may choose a different plan on each run; the retained
calls and actual disk state determine which features were demonstrated.

Use a public executable consumer for independently configured Agents,
capability subsets, and observer delivery. Use deterministic fixtures for
unknown tools, persistence faults, the exact round bound, and output limits
that a live model will not reliably trigger. Label those controls separately
from the live program. The GUI remains a stub unless an actual browser
transport has been implemented and demonstrated.

[LIVE RECEIPTS PENDING: new Chapter 3 executable, all three provider paths,
six tools, failures and recovery, public consumer, measured usage, and
independent post-run comparison. No new run or acceptance result is claimed.]
