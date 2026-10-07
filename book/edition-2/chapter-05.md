# Chapter 5: Hear the Next Instruction

A build is running, and the reader notices the agent chose the wrong file.
The correction is short. The wait for somewhere to type it is the problem.
An interface that accepts the correction only after the work has finished
turns an instruction into a review comment.
Chapter 4 keeps the build alive after a tool returns, but a client submitting
a turn still waits while the model requests another operation. A goroutine
running the process does not make the conversation responsive.

Let input arrive during that wait. Give one actor responsibility for deciding
what each input means, and let slow operations return their results as
messages. The actor can acknowledge a correction while an HTTP request or a
job-report wait is outstanding. Whether the model has received that correction
is a separate fact, recorded at the next request.

> **Implementation is awaiting contract grading and live demonstration.**
> The initial student reports six modules' local checks and core race checks
> passing. This is not chapter acceptance. The [current gate record](chapter-05-validation.md)
> tracks grading, live use, independent comparison and manuscript reconciliation;
> Bill's editorial approval remains separate.

## 5.1 One Agent, one turn owner

The existing Agent already owns its history, configuration and services.
It now owns an Actor, implemented in `internal/llm`, that schedules turns and
serializes conversation changes. Engine still owns transport and usage; Jobs
still owns running work, output and report cursors. Actor reaches Engine and
Jobs through its actual Agent parent interface. There is no new application
coordinator alongside Ensemble and no second copy of the conversation.

Declare shared actor state, messages, completion data and parent interfaces
in `internal/common`. Implement operations in llm as free functions where
shared types require it. Mailbox helpers keep their owner and logging path.
Tools and jobs still cannot import llm. Their parent interfaces provide the
route for returning immutable facts to the Agent's actor.

Retire the synchronous orchestration implementation after moving its callers.
The public blocking call remains: it submits a request to the actor and waits
for that particular request's completion. A CLI, GUI stub, and embedding
program must not choose different execution loops merely because one caller
wants a blocking API.

The first edition put the danger plainly: two loops over one engine mean two
owners of mutable state. Keeping the old loop for familiar callers would make
the new actor compete for its own conversation. Preserve the familiar call
by giving it another entrance to the same owner.

## TL;DR

Continue the exact validated Chapter 4 source in `solutions/edition-2/main/`.
Keep history in the outer Ensemble repository; `ch05/` is its frozen chapter
export after validation. Read the
full [coding skill](skills/ensemble-coding/SKILL.md), architecture and current
contracts before code. Fresh student context excludes old chapters, answers,
grader implementation and author research notes.

1. Add one Agent-owned actor and mailbox. All runtime conversation changes
   pass through that actor's existing validate/persist/apply path. Jobs retain
   their own synchronized work state and post facts for durable recording.
2. Give every submitted prompt a distinct request identity and reliable,
   reusable completion. Concurrent prompts queue in admission order as separate
   turns. Explicit hints are a different operation. Blocking submission wraps
   this same path; it never waits for an unrelated observer's Idle event.
3. Run model HTTP and job-report waits outside the actor. Workers return owned
   facts and operation identity. Do not hold a conversation lock across I/O,
   waiting or observer callbacks. Preserve serial tool dispatch and result order.
4. Persist turn start/end and hint receipt using §5.3. Hints affect the next
   request that includes them, never the already-sent body. Request snapshots
   retain their effective configuration, provenance and consumed hint sequences.
5. Interruption ends one turn, cancels its model wait and prevents further
   dispatch. Settle outstanding accepted calls truthfully, return that request's
   interrupted result once, and keep the actor alive. Running jobs survive;
   late terminal facts remain attributable and cannot reopen the turn.
6. Cancel a queued request without making a model call. Cancel an active request
   through the same turn-ending machinery. Close refuses new admission, settles
   queued requests, performs Chapter 4 cleanup, and stops the actor only after
   its managed operations and durable completion work have settled.
7. Keep the sixteen-model-request turn bound. Preserve actual append-time call
   identities when job events arrive during HTTP, and usage accounting once per
   accepted response. Replay performs no effects and reconstructs request bytes.
8. Keep durable events distinct from transient progress. Publish attributed
   state and finalized-part observations after their facts are durable. Slow
   display consumers cannot stop the actor or determine request completion.
9. Add reliable completion collection on Ensemble. A collection waits for
   available requested completions and atomically drains those already ready;
   two ready completions must come back together under the published barrier.
10. Extend human chat with `/hint TEXT` and `/interrupt`, while ordinary text
    remains a new queued prompt. Preserve explicit protocol mode, optional GUI
    module, offline paths and headless use. Demonstrate a public three-Agent
    author/editor/reviewer program, and real human-terminal control on all APIs.

**Yours.** Private method names, queue representation and synchronization
mechanisms. The lifecycle, durable payloads, ownership and observable outcomes
below are fixed. No model streaming, tool-batch parallelism, agent-spawning
tool, persisted runtime queue or automatic restart recovery is introduced.

From the course repository root:

```sh
make grade-dir CH=6 DIR=solutions/edition-2/main
```

That command retains the historical grader and its original fixtures. The
second-edition acceptance command is:

```sh
PATH="$HOME/go/bin:$PATH" python3 scripts/edition2/accept_ch05.py solutions/edition-2/main
python3 scripts/edition2/audit_ch05_mutations.py solutions/edition-2/main
```

The expanded path makes the installed Delve executable available to the test
runner. These checks preserve the seven category weights and prior behavior
while exercising the new public interfaces. The historical grader's fixture
assumptions do not define the new contract; retain its results separately.
Local acceptance still needs the live demonstrations and independent review.

## 5.2 A queue needs a meaning

Two callers can ask the same Agent different questions before either answer
returns. An Idle notification tells both that something finished. A shared
last-answer field then gives both the same text. The program can pass a
single-caller demonstration while answering one of its real callers incorrectly.

Assign each admitted request a nonempty identity unique within the Agent's
lifetime. The pair of Agent ID and request ID identifies its result. Return
a handle immediately from asynchronous submission. Its completion contains
that identity, final text and typed response parts, outcome, count of unsent
hints, and safe error when present. A completion is stored once and can be
read by several waiters;
one waiter cannot consume it away from another. Callers can release completed
handles, so the Agent need not retain an unbounded second answer archive.

The blocking API submits through this path and waits on its own handle.
A canceled wait on a handle merely stops that wait. Cancellation of the
blocking submission's caller context requests cancellation of its associated
request too. A caller that wants to stop waiting while leaving work alive
uses the asynchronous handle explicitly. These operations must not secretly
cancel another caller's active request.

Validate prompt syntax and copy caller-owned input before admission. A queued
prompt has no `message_received` event until its turn starts. It cannot become
a hint because another turn happened to be active. FIFO means successful
enqueue order, including concurrent senders; it does not mean wall-clock
order before the senders acquired the queue's admission lock.

Posting must not wait for consumer progress or space in a bounded channel.
A brief mutex acquisition and allocation are permitted. An owned growable
queue with a coalesced wake signal is sufficient. This choice can grow memory
under an unlimited producer; it is not a claim of bounded memory or wait-free
execution. Closing the Agent closes admission under the same ordering rule:
an operation is either accepted and settled, or refused as stopped.

Prompts, hints, interrupts, cancellation and worker facts use one scheduling
order. Private worker messages cannot be constructed by public callers.
Job state may change on a worker, but its durable event reaches the same
actor rather than appending through a competing writer. Offline load still
uses the same reducer; it does not start a live actor or worker.

## 5.3 Record the decisions that change context

Extend version 1 with the following event types. Each retains Chapter 2's
positive increasing sequence, UTC timestamp and exactly one matching payload.

| Event | Payload | Required fields |
|---|---|---|
| `turn_started` | `turn` | nonempty `request_id` |
| `hint_received` | `hint` | active `request_id`, nonempty `text` |
| `turn_ended` | `turn` | matching `request_id`, `outcome` |

Outcomes are `success`, `interrupted`, `canceled`, `error`, `round_limit`, and
`stopped`. A start requires no active turn, active HTTP response slot or
unanswered call. A new runtime turn records start, then its human message.
An end requires the matching active turn and no unanswered call or active
response slot. Reject duplicate ends, overlapping starts, unknown outcomes,
and a hint naming another or absent turn before persistence. Older compact
fixtures without turn records remain valid under their previous rules;
once a start appears, enforce these explicit boundaries. Reject a reused
request ID within that Agent's log. A successful end requires a final response
without tool calls accepted in that turn; an empty start/end pair cannot
manufacture a successful answer.

The input and request events remain authoritative. Turn records do not copy
their text or invent another provider response. A successful end occurs after
the accepted final response and before delivering success to its caller.
A save failure faults the Agent and completes callers with a persistence
failure; it cannot be reported as durable success because the model answered.

A hint becomes pending, one-request human guidance. It is neither a new
queued prompt nor an enduring system instruction. Extend `request_sent.request`
with `hints`, the exact ordered array of pending hint-event sequences included
in that request. Newly captured records always include the array, even empty.
Validate that it names exactly the currently pending hints, in order, then
consume them. Older records lacking `hints` consume the hints then pending,
just as the older ephemera compatibility rule does.

Render pending hints as ordinary human text after the completed tool-result
group and pending human prompt, preserving hint order. Keep the existing
provider-specific grouping of tool results; guidance cannot split a call
from its required result. Each hint is a text part whose text is exactly the
recorded string. Do not promote it into system authority or interpolate it
inside tool output. Hints are excluded from enduring dialogue once consumed,
while their original facts remain in the log.

If a request has already been sent, a new hint is not in that body and its
event is not in that request's `hints` list. If the response ends the turn
without another request, retain the hint for the next actual request. The
completion and chat display report the count of unsent hints explicitly.
Receiving a hint does not force an extra paid request merely to consume it.
Interrupting a turn also retains its unconsumed hints; `/hint` is refused
while idle rather than silently starting a turn.

### Gemini: put the correction where the model can read it

The terminal can acknowledge a hint while Gemini is working on an earlier
request. That acknowledgement proves persistence. Delivery happens when a
later `generateContent` body actually contains the text.

There is no provider `hint` role or wire field in this adapter. Keep the
returned `functionCall` in its `role:"model"` content, with any returned ID
and `thoughtSignature` on their original part. Follow it with the complete
result group in `role:"user"`. Each `functionResponse` carries its matching
name and supplied call ID, with the rendered result under `response.result`.
Then append each pending hint as a separate `{"text":"..."}` part, in receipt
order. Chapter 2 merges adjacent user entries, so results and hints occupy
the same user content. A pending ordinary prompt, if present, precedes the
hints. Every accepted call receives its result before this added guidance.

For one completed read and one hint, the suffix of `contents` has this shape.
Earlier history, tool declarations and configuration are omitted. The signature
placeholder illustrates placement only; an actual request preserves the exact
returned bytes. This is an illustrative request fragment, not a live receipt.

```json
[
  {
    "role": "model",
    "parts": [{
      "functionCall": {"id": "read-1", "name": "read_file", "args": {"path": "notes.txt"}},
      "thoughtSignature": "<returned signature>"
    }]
  },
  {
    "role": "user",
    "parts": [
      {"functionResponse": {"id": "read-1", "name": "read_file", "response": {"result": "port=8080"}}},
      {"text": "Report only the port."}
    ]
  }
]
```

The neutral event's `hints:[4]` records which hint this request consumes;
it is not an extra Gemini request property. Keep the literal hint outside
`functionResponse.response`, function arguments and `systemInstruction`.
Preserve Chapter 2's signature provenance and part boundaries. The
[GenerateContent signature guide](https://ai.google.dev/gemini-api/docs/generate-content/thought-signatures)
requires returning a signature on the part that supplied it. Adding guidance
does not authorize reconstructing or stripping that signed part.
Ordinary user text can also start a new provider turn for signature validation.
Ensemble's hint still belongs to its current actor turn; those two turn labels
need not coincide. Preserve the signatures in either case.

Check three distinct facts: the in-flight body lacks a later hint, the next
captured body contains the exact hint after the results, and the following
request omits that consumed hint. Replay the exact recorded prefix for each
comparison. Model compliance with the correction is a separate live observation.
Retained hint events remain in the log even after their text leaves the next
request's projection.

This literal log is an offline fixture, using fictional provenance. Its
compact responses omit request records intentionally:

```jsonl
{"log_version":1}
{"seq":1,"type":"turn_started","time":"2026-01-01T00:00:00Z","turn":{"request_id":"r1"}}
{"seq":2,"type":"message_received","time":"2026-01-01T00:00:01Z","message":{"actor":"human","purpose":"dialogue","parts":[{"type":"text","text":"Read notes.txt."}]}}
{"seq":3,"type":"response_ended","time":"2026-01-01T00:00:02Z","response":{"from":{"vendor":"anthropic","model":"fixture-messages","surface":"messages"},"parts":[{"type":"tool_call","call_id":"read-1","from":{"vendor":"anthropic","model":"fixture-messages","surface":"messages"},"name":"read_file","args":{"path":"notes.txt"}}],"usage":{"input":1,"cache_write":0,"cache_read":0,"output":1}}}
{"seq":4,"type":"hint_received","time":"2026-01-01T00:00:03Z","hint":{"request_id":"r1","text":"Report only the port."}}
{"seq":5,"type":"tool_called","time":"2026-01-01T00:00:04Z","tool":{"call_id":"read-1","name":"read_file","args":{"path":"notes.txt"}}}
{"seq":6,"type":"tool_returned","time":"2026-01-01T00:00:05Z","tool":{"call_id":"read-1","parts":[{"type":"text","text":"port=8080"}]}}
```

Rendering this prefix produces the result before the human hint on every
surface. Append the following event to a separate copy:

```json
{"seq":7,"type":"request_sent","time":"2026-01-01T00:00:06Z","request":{"to":{"vendor":"anthropic","model":"fixture-messages","surface":"messages"},"ephemera":[],"hints":[4]}}
```

Reconstruct the attempted request from the prefix before this event and its
recorded configuration. Rendering the context after consumption is a different
question and must not be substituted for that comparison. A second render
of the same prefix has identical bytes and consumes nothing. Duplicating 4
in the list, using 5, or adding another hint that the list fails to include
must fail validation. A turn end inserted before the result fails too.

## 5.4 Let slow work send its answer back

The actor prepares an immutable request snapshot and records `request_sent`,
then starts model I/O on an owned operation. It goes back to its mailbox.
The worker returns parsed facts or a safe error, tagged with the operation
and turn identity. Only the actor accepts the result into history. A worker
cannot discover a newer model setting through its parent and retroactively
change the provenance of the request it already sent.

Capture effective request configuration when the turn starts. Public changes
while that turn is active are refused until a later chapter specifies live
settings transitions. Engine adds usage once for each accepted response,
including tool-only responses. A canceled or discarded response does not
become a fabricated zero-token success. The displayed totals count accepted
response observations; they are not a claim to know a provider's bill for
an abandoned request whose usage was never accepted.

Process a complete accepted response once. Dispatch tool calls serially in
its recorded order, as in Chapters 3–4. A report waiter may wait on output,
a pattern or a delay, but it does so outside the actor and posts its report
back. Start the next call only after the previous call's report has been
accepted. Running jobs can overlap because a report can describe continuing
work; this chapter does not introduce parallel dispatch of a model batch.

A process that stops reading can fill its terminal's input buffer. Even
`send_input` then waits: moving only HTTP and report waits off the actor leaves
the mailbox stuck behind a write. Run potentially blocking process-input I/O
as an owned, cancelable operation too. The actor must continue acknowledging
hints, interruption and close while that write is parked. Preserve serial tool
effects and Chapter 4's output position captured before the input write.

Cancellation stops further input and settles the writer before its call is
reported finished. Bytes already accepted by the terminal cannot be rolled
back; report partial or interrupted work truthfully instead of claiming the
entire input was delivered. An interrupted turn leaves the job alive under
§5.5, while Agent close still kills and drains managed processes. An input-size
cap alone cannot establish responsiveness because even a short write can meet
an already full buffer. Keep job locks free during this I/O and join owned
workers during cleanup.

Chapter 4's worker owns the job's terminal fact. It submits that fact for
durable recording even while another HTTP operation is in flight. Preserve
its output-close/drain and killed-state rules. The actor assigns sequence
numbers only when appending. In particular, missing generateContent call IDs
still use `call-<actual-response-seq>-<part-index>`; a predicted sequence from
before HTTP is invalid after an intervening job event.

Keep job locks, mailbox locks and the event writer out of long waits. A job
worker that holds its state lock while waiting for the actor can deadlock
the actor trying to obtain a report. Transfer an owned snapshot or request a
bounded owner operation with a documented lock order. More goroutines do
not repair two owners waiting on one another's locks.

## 5.5 Interrupt the turn, keep the Agent

An interrupt targets the turn active when the actor processes that control
message. With no active turn it acknowledges a no-op. The public API also
supports cancellation by request identity, so a caller can cancel its queued
request without accidentally interrupting somebody else's active one.

Once the actor accepts an interrupt, start no further model request or tool
effect for that turn. Cancel the outstanding model operation and close its
response slot with a safe `error_occurred` code `interrupted`. The usual
error reduction discards a still-uncommitted human input but preserves earlier
accepted dialogue, tool effects and usage. If the response fact was accepted
first in mailbox order, its history remains; cancellation is no rollback.

Every call in an already accepted assistant batch still needs a result.
For a dispatched job whose report has not yet been accepted, end the report
wait and obtain an immediate truthful snapshot, including currently unseen
output and its running/done/killed status. Persist that one result. Ignore
the superseded report message if it later arrives; never consume the output
cursor twice. Preparing a report must not irreversibly consume its cursor:
reserve its interval and commit consumption only when the actor accepts that
report, or use an equivalent Jobs-owned transaction. Discarding a superseded
report releases the reservation without losing unseen bytes. For a call
whose execution has not started, record a call/result
pair with an explicit error saying the turn was interrupted before execution.
No job or side effect is invented for that refusal. Apply the existing
next-attempt one-shot-limit rule when recording such a refused attempt.

Finish pairing in call order, append `turn_ended` with `interrupted`, and
complete the request once. Running jobs remain owned by Jobs; the next turn
can inspect or stop them through their handles. A subsequent `job_ended`
remains durable and observed, but creates no second tool result, no second
request completion and no automatic model continuation.

An old model worker may return after cancellation. Its operation identity
prevents its response from being applied to a later turn. Discard those
unaccepted response facts and record a safe debug reason without logging the
body. A stale response is different from an attributable real job completion:
the former must not answer a new question, and the latter must not disappear.

Queued cancellation removes that request before activation and completes it
as `canceled`, with no human event, turn event or HTTP request. Active caller
cancellation follows interruption's same pairing and slot-closing machinery,
using `canceled` as outcome. If success was durably committed first, cancellation
cannot rewrite it. The handle retains the committed completion even if the
canceling blocking caller returns its context error instead of waiting for it.

Close has a larger scope. Refuse new admission, complete queued requests as
`stopped`, settle the active turn, and perform Chapter 4's explicit job
shutdown. Kill managed process groups, drain/join their readers, and record
their terminal outcomes before closing the log and actor. For a nonkillable
Go function, retain Chapter 4's truthful killed-job semantics: the function
may continue, its later output is suppressed, and Close must not claim it
terminated. Do not wait forever for an arbitrary Go function to cooperate.
Cancellation-capable model/report workers must stop and be joined.

Repeated close is safe; subsequent requests return a distinguishable stopped
error. A persistence fault still triggers cleanup and settles every waiting
caller with an error. Cleanup cannot guarantee a durable final event when
the log itself failed, and the owned logger must say so without claiming
rollback or silently leaving a managed process alive.

## 5.6 Progress does not acknowledge a request

Keep the existing Agent-attributed durable event observations. Add transient
state observations carrying Agent ID, active request ID where applicable,
and the old/new state: `idle`, `input_pending`, `in_flight`, `tools_pending`,
`interrupted`, or `stopping`. An interrupted transition is followed by idle or stopping;
it does not end the actor's lifetime. A hint acknowledgement identifies its
persisted sequence and whether it has been sent yet.

Finalized-part observations carry Agent ID, request ID, response event
sequence, zero-based part position and an owned copy of the actual typed
part. They can describe text, a call or opaque material; do not restrict the
type to text and then promise every part fits. A non-streaming response may
produce a single text delta per text part before its final observation.
If offered, identify that delta as delivery of a completed response, not
evidence of incremental provider streaming. Deltas are not durable history.

Public display subscriptions must not hold up the actor. Give each an owned
delivery queue; callbacks obey the existing nonblocking callback contract.
If a bounded queue fills, close that subscription and expose an overflow
reason. Do not silently lose a final event while presenting the subscription
as current. Its consumer can obtain a fresh authoritative snapshot and
subscribe again; a gap-free snapshot/live handoff is not promised here.
Callback recipients still receive separate owned copies.

When a closed subscription's delivery worker exits, release its callback/client
reference and queue storage. Retain only the small status record needed to
explain closure. A closed Ensemble refuses new subscriptions, so reconnecting
a client cannot start an orphan delivery worker after application shutdown.
The callback's nonblocking contract still applies; cleanup cannot forcibly
terminate arbitrary user callback code.

Request completion never depends on that queue. A terminal can display a
request's result from its reliable handle even if a progress subscription
overflows. A slow or unsubscribed GUI cannot stop model work or consume the
CLI's answer. The separate optional GUI module uses public submission,
control, completion and observation interfaces; no WebSocket implementation
is needed for this chapter.

Ensemble owns and exposes a completion collection for a fixed set of request
handles. Its shared declaration and Ensemble parent interface belong in common;
collection behavior belongs with application coordination, without injected
sibling services. It waits until at least one member is ready, then atomically drains
all ready members not previously returned by that collection. Return results
in the collection's declared order. Completed handles remain individually
readable, and different collections do not consume one another's results.
An empty set returns immediately; a completed collection reports exhausted.
Canceling a collection wait leaves its handles and unfinished requests alive.

The coalescing test has an explicit boundary: complete both registered
requests before calling the collection's wait/drain. That one call returns
both exactly once. A separate test parks the collection, completes one
request, observes one result, then completes the other and receives only it.
Do not grade a scheduler-dependent phrase such as “the same turn window.”
Use completion notification rather than periodic polling; channel layout is
an implementation choice, not the proof of either behavior.

## 5.7 Keep the terminal usable while work runs

Extend Chapter 2's human `chat`, not a second debug-only command. Its input
reader remains available during a turn. Ordinary text always submits a new
prompt, acknowledges its request ID and queues if necessary. Display each
final answer under its matching ID. Never reinterpret an ordinary follow-up
as a hint merely because the previous answer is unfinished.

Add `/hint TEXT` and `/interrupt` to `/help`. Hint applies to the active turn
and is refused locally when idle. Interrupt targets the active turn; display
whether it interrupted anything. Acknowledge hint receipt separately from
delivery to the next request. Preserve `/history`, `/usage`, slash escaping
and the input limit. Refuse context-changing `/ephemeral` or `/redact` while
a turn is active, without corrupting it; they remain available between turns.
These busy/idle control refusals are recoverable acknowledgements, not fatal
model failures. Ordinary provider/parsing failures retain the earlier CLI's
safe error-and-close behavior, while the public library can submit a later
request to an unfaulted actor.

Render input prompts and asynchronous answers so that each complete answer
and control acknowledgement is readable. Exact cursor repainting and fancy
line editing are optional. Serialize terminal writes so a tool observation
cannot interleave JSON or arbitrary fragments through an answer. `/quit`
performs Agent close. EOF stops input, lets admitted prompts finish normally,
then closes jobs and reports usage; fatal local input or persistence failures
close promptly with error instead. This orderly EOF changes Chapter 4's
already-synchronous client into a usable piped asynchronous client.

Machine `protocol` keeps the prior directives and adds the following exact
record forms. IDs and sequence numbers below illustrate their types; actual
values come from the owning Agent.

| Input | Output |
|---|---|
| `{"kind":"prompt","text":"Read notes.txt."}` | Immediate `{"accepted":"prompt","request_id":"r1"}`; later exactly one completion record described below |
| `{"kind":"hint","text":"Only report the port."}` during a turn | `{"ack":"hint","request_id":"r1","seq":4,"sent":false}` after persistence; sent:false describes receipt time, not later delivery |
| `{"kind":"interrupt"}` during a turn | `{"ack":"interrupt","request_id":"r1","interrupted":true}` after accepting interruption; that request also receives its one completion |
| `{"kind":"interrupt"}` while idle | `{"ack":"interrupt","request_id":"","interrupted":false}` |

A completion has this shape:

```json
{"completion":{"request_id":"r1","outcome":"success","text":"8080","pending_hints":0}}
```

For a nonsuccess outcome, `completion` also contains `error` with safe `code`
and `message` strings; retain any accepted final text without inventing one.
The public handle additionally exposes typed parts. Accepted output precedes
that request's completion even when the response is immediate. All protocol
records are complete serialized JSON lines.

A valid JSON record with unknown `kind`, empty required text, extra operation
fields, or a hint while idle emits
`{"error":{"code":"invalid_control","message":"..."}}` and continues.
The message gives a static useful reason, without echoing input. Such a record
is not admitted and makes no history or HTTP operation. Malformed JSON and
invalid legacy directives retain their previous fatal behavior. A real
request failure with outcome `error` or `round_limit` emits its reliable
completion, reports the safe reason on stderr and closes the CLI. Requests
still queued when Close is admitted receive `stopped` completions. The actor
may have started the next prompt before the client observes the failure and
admits Close; that request follows the active-turn close rules instead. A
completion or effect accepted before Close is not rolled back. There is no
implicit actor halt-on-provider-error policy. No successful-session usage
record follows that failure. Intentional `interrupted` and `canceled` outcomes are control results:
keep both chat and protocol alive so queued and later prompts can finish.
`stopped` accompanies explicit close or cleanup; it is not a new provider error.

Legacy `{"user":"..."}` is an explicit queued prompt, with no added accepted
record: preserve exactly one `{"assistant":"..."}` per successful legacy
prompt and the final clean-EOF usage object. Its answer remains associated
internally with its request handle. A dedicated observation mode or separate
destination may expose event JSON; never inject that progress into the legacy
protocol. Invalid control syntax cannot become model text.

Build `examples/workflow` as a public consumer with author, editor and reviewer
Agents. Give each its own log, role instruction and visible tools. A shared
scratch workspace can hold a draft: author uses read/write, editor uses
read/write/edit, reviewer uses read only. Pass completed output to the next
role using reliable request handles. The final reviewer answer is an actual
model result, not a hardcoded approval string. Adding the third Agent must
require no special case inside the library.

For responsiveness, use the existing `run_command` tool in a separate
scratch exercise: request a command with a known delay and a report wait
long enough to receive a hint before its report. No special production
`think` tool is needed merely to sleep. Preserve the legacy fixture for old
solutions; a new checker must use tools this published contract provides.

## 5.8 What remains true about references

Chapter 2 already introduced path, URI and handle references. This chapter
does not reinterpret a retained locator as a guarantee that its target still
exists, or add media support merely by adding an actor. Preserve existing
adapter mappings and loud unsupported-reference errors. An unsupported
content error identifies the selected model, media MIME/category and missing
mapping without echoing the private locator or credential.

Unknown content capability differs from a known unsupported capability.
Where a model-feature table is used, an unknown row must not claim verified
support; independent image/audio/video/document flags are not mutually
exclusive. Do not refuse an otherwise valid text-only request merely because
the actor chapter has no media row for a newly discovered model. Actual
content-capability expansion belongs with the adapter that implements it.

## 5.9 Checks before a demonstration

Retain prior behavior and the inherited seven-check score: parity 10,
responsiveness 25, replay 15, observers 15, completion collection 15,
loud refusal 10, and architecture 10. The independent checks must establish
the actual properties rather than infer them from those labels.

| Property | Required positive and negative controls |
|---|---|
| One owner | Concurrent blocking/asynchronous clients share one actor; repeated start never adds a drainer; every new spoke and executable follows the star and parent chain |
| Replies | Overlapping distinct prompts get their own answers; two waiters can read one completion; no observer or an overflowing observer does not lose a reply |
| Admission | FIFO queued prompts, canceled queued request makes no HTTP/log turn, explicit hint differs from prompt, stopped Agent refuses new work |
| Responsiveness | Barrier-controlled blocked HTTP, report waits and a full process-input buffer accept hints/controls before release; interrupt and close settle the owned input writer; hint receipt and later wire inclusion are checked separately |
| Interruption | Interrupt active request, retain paired accepted calls, reject stale response, preserve late job event once, and successfully complete a later prompt |
| Close | Park active and queued callers, close, require each to settle; managed processes/readers stop, repeated close works, local nonkillable work is reported truthfully |
| Replay | Capture actual sanitized request bytes; reconstruct from its exact prefix/configuration and compare bytes, including hints and ephemera; deliberately remove a hint to make the comparison fail |
| Sequences | Job terminal event arrives during model I/O, then an omitted provider call ID uses the actual later response sequence |
| Collection | Both-ready barrier returns both in one drain, staggered readiness returns each once, canceled wait does not cancel its members |
| Bounds | Sixteen-request limit survives actor migration; final batch is paired, no HTTP seventeen or extra turn is invented |
| Clients | Actual human PTY chat controls and readable request attribution, unchanged machine protocol, public three-Agent consumer, separate GUI stub and headless build |

Use synchronization barriers for race-sensitive fixtures rather than hoping
a fixed sleep lands inside the interesting interval. Keep passing controls
and deliberately remove each protected property. A test that never overlaps
two requests cannot prove those requests retain distinct answers.

After the first student implementation and runs, the independent reviewer
compares the corresponding first-edition standard and later lifecycle fixes.
The student receives rationale and revised new teaching, without reading old
source. Retain the initial checkpoint, revision and distinguishing checks.
Final validation requires the reviewed improvements and real user path.

## 5.10 Taking it for a spin

The current second-edition Gemini validation scope starts at Gemini 3.0 Flash
and newer. Use **Gemini 3.8 Flash** for this chapter's hint and control
demonstrations. Discovery on October 7, 2026 confirmed
`models/gemini-3.8-flash` supports `generateContent`; the REST request uses
`POST /v1beta/models/gemini-3.8-flash:generateContent`. Record that exact model
in every Gemini run, including EOF, workflow and collection. A discovery result
establishes availability, while the live receipts establish behavior. Older
model attempts remain historical evidence. This test selection does not impose
a hardcoded model allowlist on the library.

Launch `chat` in an actual terminal/PTY using a fresh log and scratch workspace.
Select a discovered tool-capable model and safe environment credentials. Ask
for a slow command with an observable final marker. When its report wait is
active, type `/hint` with a concrete change for the subsequent answer. Inspect
the immediate acknowledgement, recorded hint, next request inclusion and actual
model answer as separate pieces of evidence.

Submit another ordinary prompt while one is active and watch the distinct
request IDs. Interrupt the active turn and verify that the queued request
can still finish. Inspect the continuing job through a later turn, then stop
it explicitly. Exercise `/quit` with managed work alive and orderly EOF with
queued prompts. Repeat the human-interface feature checklist on Messages,
Chat Completions and generateContent; an external HTTP cancellation may
be too timing-sensitive live, so label its deterministic barrier check.

Run the author/editor/reviewer consumer through the public API with real
models, inspect the actual draft file and edits, and retain each role's
attributed log and completion. Use separate logs for the two-Agent completion
collection exercise. The GUI remains a separate-module stub unless a working
browser transport is explicitly implemented and demonstrated.

[LIVE RECEIPTS PENDING: new Chapter 5 snapshot, actual human-terminal
controls on all three APIs, public workflow, reliable collection, replay
comparisons, measured usage, independent checks and comparative review.]
