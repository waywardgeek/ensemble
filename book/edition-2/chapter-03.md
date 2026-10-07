# Chapter 3: Six Tools, One Turn

A conversation can describe a fix without changing a byte. This chapter
lets the model read the file, make the edit, and run a command to check it.
That makes the distinction between an answer and an action matter: a
confident sentence is no evidence that the file changed.

The second-edition run supplied a useful example. A model omitted the flag
that permitted an overwrite, then offered an explanation about parallel
execution. The call log showed no overwrite request. The file still held the
old lines. The coder's follow-up supplied the missing instruction, and the next
calls finally changed the bytes. §3.10 retains that exchange because the
difference between a proposed edit, an executed edit and a story about an
edit is the subject of this chapter.

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

Continue the exact validated Chapter 2 source in `solutions/edition-2/main/`.
The outer Ensemble repository owns its history; `ch03/` is the frozen export
produced after validation. Before editing, read the entire
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
7. Preserve Chapter 2's human `chat`, explicit machine `protocol`, terminal
   mode selection, `render LOG`, `dump`, and `CH02_LOG`.
   Emit one assistant line containing the final model response per completed
   user turn. Intermediate text and tool facts stay in history and observer
   events. Sum usage across all accepted responses, not just the final one.
8. If request sixteen still has calls, complete that batch, then report the
   round limit before request seventeen. Transport, parse, persistence,
   and round-limit errors end the CLI session without a fabricated final
   answer or success-usage line. Keep already accepted history and effects.
9. Demonstrate every tool and its important failure paths through the actual
   human chat in an actual terminal and the public library on the three
   real APIs. Separate deterministic fault
   fixtures from live evidence. After the initial build and run, an independent
   reviewer compares with the first edition and supplies findings; revise
   code and prose without giving the student the old answer.

**Yours.** Internal names, implementation structure within the architecture,
tool description wording, and presentation of ordinary tool output. The
argument names, semantic defaults, result/error distinctions, and lifecycle
rules below are the public contract. No background jobs, process containment,
streaming, mailbox, concurrent tool execution, or automatic retries yet.

**Exercise.** From the course repository root:

```sh
make grade-dir CH=3 DIR=solutions/edition-2/main
```

The inherited score is a regression baseline. The additional acceptance
table in §3.9 covers requirements that the original grader never protected.

## 3.3 A call is the middle of an answer

Chapter 2 stopped at a tool-only response and explained that execution was
not yet available. Replace that notice now. Its old one-response fixtures
also need a continuation: return the tool result, then a final model answer,
rather than repeating the same call ID as though it were another response.

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
file with the default full-file range succeeds with empty content.
`end_line:0` alone also means that full-file read and succeeds on an empty
file. An explicitly supplied `start_line`, even 1, or a positive `end_line`
names a line target and fails on an empty file. The tool must distinguish
that valid empty full-file result from an I/O error.

Truncation must be visible in the result, outside the retained-content
byte budget. Report it when content was omitted, not merely when a complete
result exactly fills its limit. State which limit was reached so the model
can request a narrower range or an explicit larger limit. Numbered lines or a short
header are acceptable; returning an entire file for a requested two-line
range is not. A read failure returns a tool error with the path and reason.

Byte caps apply to retained UTF-8 text, before JSON encoding. For valid
UTF-8 input, keep the longest complete prefix that fits the budget: if the
cut falls inside a code point, omit that incomplete final code point and
report the omitted content. Letting JSON replace a broken byte with a
replacement character both changes the text and can exceed the byte cap.
Apply this boundary rule to reads, listings, search output, and each captured
command stream. Preservation fixtures use valid UTF-8 text; this chapter
defines no transcoding format for arbitrary binary or invalid UTF-8 bytes.

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

Use the three-byte UTF-8 text `éX` to test a cap independently of ASCII
output. A one-byte cap retains an empty prefix and reports omission; a
two-byte cap retains exactly `é` and reports the missing `X`; a three-byte
cap retains `éX` with no truncation notice. Exercise both read_file and
run_command's stdout with these cases. A decoded replacement character is
a failure, even if the encoded JSON is syntactically valid.

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
model selection and `CH02_LOG`, and implement Chapter 2's human chat revision.
No additional tool-specific CLI mode is needed. In a
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

After the first run, an independent reviewer compares the new answer with
the first-edition standard. The student does not read the old source. The
reviewer looks for duplicated schema metadata, awkward dispatch, missing diagnostic
context, and comments that explain syntax instead of the invariant. Retain
the initial answer and the findings, revise code and teaching, rerun affected
checks, and obtain review of the revision. Passing is the start of that
comparison, not its conclusion.

## 3.10 Taking it for a spin

Use `chat` in an actual terminal/PTY and type ordinary requests. Ask for the
six-tool exercise below, inspect the readable result, and follow up at the
next visible prompt. Use `/history` to find a real `tool_returned` sequence,
then `/redact N N demonstration` using that displayed number. Ask a further
question and inspect the recorded request to verify the result became a stub
while its pairing survived. `/usage` reports the same accounting as protocol
mode. Human chat calls the same Agent and tool loop; it needs no jobs, actor,
or model streaming to be usable.

The October 7, 2026 human-terminal sessions are preserved in outer commit
`a347ce31511c4b124e486bb41ef98c07bd17cec5`. The coder used actual macOS PTYs,
read each answer, and sent the next request at the returned prompt. Messages
selected chat by detecting the terminal; Chat Completions and generateContent
used explicit `chat`. These are coder sessions. Independent review accepted
the integrated client and the retained demonstration.

Build `./cmd` from the Chapter 3 source, select a fresh `CH02_LOG`, and launch
the executable with `chat` from a scratch workspace containing the five-line
`notes.md` fixture in §3.7. Supply credentials through the environment and
choose a discovered tool-capable model. Set `LLM_RESOLVED_MODEL` when bound
replay material requires a known identity; automatic tool continuations obey
that rule too. The first typed request in every session was:

```text
Please exercise your six tools in this scratch workspace. Use list_directory, read_file on notes.md lines 3 through 4, and search_files for beta|delta in notes.md with context_lines 1. Use write_file to create created.txt with exactly first anchor\nsecond line\n, then edit_file to replace first anchor with first edited. Use run_command with cat created.txt; printf 'STDOUT-MARKER\n'; printf 'STDERR-MARKER\n' >&2; exit 7. Report concise observed results and invent a two-word code name for us to remember. Use the actual tools, not a simulated demonstration.
```

All three paths used the six tools. The ranged read returned lines 3–4,
search merged the overlapping context around lines 2 and 4, and the file
changed from `first anchor` to `first edited`. The shell result kept stdout,
stderr and exit status 7 separate. The tool invocation succeeded even though
the command chose a nonzero exit status. The models invented `Copper Lantern`,
`Scratch Check` and `SILVER FALCON`, respectively.

At the next prompt, `/history` exposed a real result at sequence 5 with its
call ID. The coder entered `/redact 5 5 demonstration`, then requested the
overwrite-refusal exercise. Messages and generateContent refused the first
unflagged write, then replaced the file with explicit permission and appended
`tail`. Chat Completions omitted the requested `overwrite:true` call. Its
answer offered a confused explanation about parallel execution; the actual
call records showed refusal, append and read, with no replacement request.
The file still contained `first edited`, `second line` and `tail`.

The coder checked the file and sent this correction:

```text
The log shows you never requested overwrite:true. Please now use write_file with overwrite:true and content replacement\n, then append tail\n, then read_file. Execute those actions sequentially and report the actual bytes. Also recall the exact code name you invented initially.
```

The model then requested the missing operation. The tool records and disk
agreed on `replacement\ntail\n`, and the answer recalled `Scratch Check`.
No code repair was needed for this model omission. The distinction matters:
the agent executed the calls it received; the model's account did not establish
that it had requested the right calls. The
[terminal transcript](../../solutions/edition-2/ch03/evidence/ch03/human-chat/openai-main/terminal.txt)
retains the failed plan and the corrective turn.

Further ordinary-text requests tested ambiguous and missing edit anchors,
an absent file and an invalid regular expression, then repaired one unique
anchor. Each failure returned to the model, and all final `anchors.txt` files
contained `first repaired\nsecond anchor\n`. Empty-file reads succeeded with
an omitted range or `end_line:0`, while explicit `start_line:1` returned an
ordinary tool error. Reading `éX` at byte caps 1, 2 and 3 retained empty text,
`é` and `éX`; only the first two reported truncation. A two-byte command cap
retained the complete `é`, and a silent `exit 7` reported the actual status.

Chat Completions described that complete two-byte character as an incomplete
sequence. The saved tool result contradicts the description. Inspecting the
result bytes is also how the reader distinguishes an empty successful read
from a model's paraphrase of its presentation.

Every session ended with the same literal-slash test. This excerpt is from
the Messages terminal:

```text
You> //help is literal text for this test. Do not use any tools. Reply with SLASH-OK on line one and our exact original two-word code name on line two.
Assistant:
SLASH-OK
Copper Lantern
```

The other two answers retained their own original names on the second line.
`/usage` and `/quit` displayed these cumulative counters, including every
intermediate tool round and the corrective turn:

| Surface and selected model | Human turns | Tool calls | Input | Cache write | Cache read | Output |
|---|---:|---:|---:|---:|---:|---:|
| Messages, `claude-sonnet-5-5` | 5 | 27 | 78489 | 0 | 0 | 3313 |
| Chat Completions, `gpt-4.1-mini-2025-04-14` | 6 | 29 | 10777 | 0 | 33664 | 1381 |
| generateContent, `models/gemini-3.8-flash` | 5 | 27 | 120234 | 0 | 0 | 4452 |

The first two returned their selected model names; generateContent returned
`gemini-3.8-flash`. These differing call plans and cache observations are dated
run receipts, not a controlled efficiency comparison. All main sessions exited
0. Separate actual-terminal EOF sessions exited with four zero counters.
Local commands, blank input and malformed command syntax returned readable
feedback without ending the conversation. Invalid UTF-8, oversized input and
provider-failure behavior remain separately labeled local controls.

The [human feature ledger](../../solutions/edition-2/ch03/evidence/ch03/human-chat/FEATURES.txt)
links the transcripts, source/binary binding, actual tool results and disk
checks. Offline reconstruction of each later request preserved sequence 5's
call ID while replacing its result with `[redacted]`. The one-request
`LILAC-614` directive appeared in exactly one reconstructed request. Repeated
rendering was byte-identical and left logs and files unchanged. These are
reconstructions of saved log prefixes, not intercepted HTTP request bodies.
Runtime terminal evidence is macOS; no Linux live run is claimed.

### Earlier machine-interface evidence

The earlier receipts remain useful for the protocol and public library.
They are preserved at their original source versions and are not relabeled
as the human sessions above.

The October 7, 2026 initial sessions ran all six tools through the JSON-lines CLI on
Messages, Chat Completions, and generateContent. The runner planted only
`notes.md` in each fresh workspace, then asked the model to list the directory,
read lines 3–4, search `beta|delta` with context, create a file, edit its
unique anchor, and run a verification command. The source and receipts are
preserved at student checkpoint `540fb4a`, following the initial code
checkpoint `590c4f4`.

The command emitted distinct stream markers and deliberately exited 7.
The tool returned a successful invocation with that actual status, stdout
containing the file and `STDOUT-MARKER`, and stderr containing
`STDERR-MARKER`. The model could report the failed command without losing
the rest of the turn.

The next prompt tried to replace `created.txt` without permission. On the
Messages run the result refused replacement of the existing 25-byte file.
The model then supplied `overwrite:true`, appended `tail\n`, and read the
result. Independent disk inspection found these exact bytes on all three
paths:

```text
replacement
tail
```

Another prompt created `first anchor\nsecond anchor\n` in `anchors.txt`.
The ambiguous edit of `anchor` failed with two matches, the edit of
`missing` failed with zero, and a read of an absent file failed. The model
received all three error results, selected `first anchor` as a unique
target, and completed the repair. Both the tool record and independent
disk inspection showed:

```text
first repaired
second anchor
```

Each initial CLI session completed four human turns and exited successfully.
Its measured cumulative usage was:

| Surface and selected model | Input | Cache write | Cache read | Output |
|---|---:|---:|---:|---:|
| Messages, `claude-sonnet-5-5` | 56911 | 0 | 0 | 1976 |
| Chat Completions, `gpt-4.1-mini-2025-04-14` | 4004 | 0 | 20608 | 775 |
| generateContent, `models/gemini-3.8-flash` | 50081 | 0 | 0 | 1469 |

The first two returned those same model identities; generateContent returned
`gemini-3.8-flash`. The differing call plans and cache observations make
these run receipts, not a controlled model-efficiency comparison. Usage
includes the intermediate tool rounds as well as the final visible answers.

### Repeat the task, including its refusals

Build `./cmd` from the new Chapter 3 module, run it in a scratch directory,
and select a fresh `CH02_LOG`. Supply credentials through the environment
and use a currently available tool-capable model. Configure a known resolved
identity with `LLM_RESOLVED_MODEL` where required for bound replay material.
Automatic continuation must follow the same provenance rule; it cannot
guess an alias after tools have changed files.

Plant the five-line fixture from §3.7. The actual
[input transcript](../../solutions/edition-2/ch03/evidence/ch03/live-anthropic/stdin.jsonl)
contains the four prompts used above; feed those lines to `protocol` and end
input. Inspect the resulting files yourself. Then dump and render the
recorded log twice without credentials. In all three original runs the
renders were byte-identical, the original log stayed unchanged, and the
scratch-file hashes stayed unchanged. Offline replay executed no edit.

The initial recall question asked for a supplied code name and status.
The later guided-revision runs used a stronger check: the model invented
its first answer, then had to recall that exact answer in the next turn.
Messages returned `Silent Seven` twice, Chat Completions `SilentEcho`
twice, and generateContent `Cobalt Falcon` twice.

Those revision runs also exercised the clarified empty-file rule through
actual calls. An omitted range and an `end_line:0`-only range both returned
valid empty text. An explicit `start_line:1` returned an ordinary tool error.
The turn continued and a silent `exit 7` still produced status 7 without
becoming a tool error. The receipts retain the guided revision's source
hashes rather than attributing these results to the earlier binary.

### Same filename, different Agents

The public program in
[`examples/tools-consumer`](../../solutions/edition-2/ch03/examples/tools-consumer/main.go)
created two workspaces containing a different `same.txt`. One Agent received
only read_file; the other received read_file and list_directory. Each real
model used its Agent's tool and returned the corresponding marker,
`NORTH-314` or `SOUTH-927`, on every API path. The shared filename never
made one Agent read the other's directory.

The consumer also checked Agent-attributed, increasing observer sequences,
including actual call/result events, and stopped receiving events after
unsubscription. A third Agent selected no tools and completed a real text
turn with `READY`. Unknown/disabled dispatch, exact round bounds, and
persistence faults are separately labeled deterministic controls. The
optional GUI is still a stub tested through its public module boundary;
these runs make no browser or WebSocket claim.

The [feature ledger](../../solutions/edition-2/ch03/evidence/ch03/FEATURES.json)
links initial CLI, consumer, offline, and guided-revision evidence. Comparative
review subsequently exposed the split-UTF-8 cap defect now described in
§§3.5 and 3.7. The earlier successful ASCII demonstrations could not prove the
multibyte case; explicit Unicode controls did. The
[validation record](chapter-03-validation.md) retains the acceptance counts,
deliberate defects, legacy checks and exact reviewed-source chronology.

Human integration also exposed an evidence-tool defect: a temporary pathname
was being trusted as the executable's identity. Review required the verifier
to check immutable source and executable hashes before writing reconstructed
requests. That evidence-only repair changed no production code and required
no repeated paid run. The accepted checkpoint is `edition-2-ch03-r1`; its
manifest binds source, frozen export and evidence. Editorial approval remains
separate from these checks.
