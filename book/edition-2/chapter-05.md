# Chapter 5: Hear the Next Instruction

It was 2 AM and I was watching CodeRhapsody edit the wrong file.

Not the wrong kind of file. The right kind, the right directory, three
characters off in the name. I could see the mistake in the reasoning, I
knew the fix was four words long, and I could not type them. The model
was mid-turn, waiting on an HTTP response that would come back in eight
seconds, and the interface did not accept input during a turn. So I sat
there, mass-producing adrenaline, watching a 200,000-token agent burn
$0.40 of reasoning on a file that did not contain what it expected.

Eight seconds is not a long time. It felt like surgery with the lights off.

The correction arrived as a review comment after the turn finished, by
which time the model had already written forty lines of code against the
wrong file, discovered nothing matched, reasoned its way into a second
tool call to investigate, and was now confidently heading toward a
theory about a "refactored module structure." The four words I could not
type were: "You mean auth.go."

This is the chapter where input stops waiting for permission. An actor
owns the conversation, slow operations return their results as messages,
and a correction can arrive while an HTTP request is still in flight.
Whether the model has received that correction is a separate fact,
recorded at the next request. The human can always type.

> **Reviewed runtime and live demonstrations are recorded below.** Initial
> attempts and affected-path reruns keep their own source identities. The
> [current gate record](chapter-05-validation.md) tracks final independent
> evidence review, manuscript proofreading and checkpoint; Bill's editorial
> approval remains separate.

## 5.1 One Agent, one turn owner

Up through Chapter 4, every turn is a function call: submit a prompt, wait
for the model, dispatch tools, repeat until the model stops requesting them,
return the answer. The caller's goroutine walks the entire loop. That
simplicity has a cost: while the goroutine is inside `client.Do(req)`, nobody
is listening.

The fix is the oldest trick in concurrent programming. Give one goroutine
exclusive ownership of the mutable state, and let everyone else talk to it
through a mailbox. The HTTP call, the job-report wait, the process I/O all
happen on worker goroutines. When they finish, they post a message. The
actor applies it, persists the change, and notifies anyone watching. The
human types a hint, it goes into the mailbox. The human types `/interrupt`,
it goes into the mailbox. The actor decides what each message means, in the
order they arrive, and the conversation is never touched by two goroutines
at once.

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
3. Run model HTTP, job-report waits and process-input I/O outside the actor. Workers return owned
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

When two callers ask the same Agent different questions before either answer
returns, something has to break. A shared "last answer" field gives both
callers the same text, and neither knows whose question it answered. A
demonstration that submits one prompt at a time passes perfectly while the
real program silently delivers wrong answers.

The fix is identity. Every admitted request gets a nonempty ID, unique within
that Agent's lifetime. The caller gets a handle back immediately. The handle
has a `Done()` channel and a `Wait(ctx)` method. When the answer arrives,
it carries the ID, the final text, the typed response parts, an outcome, a
count of unsent hints, and (when something went wrong) a safe error. The
completion is stored once and can be read by several waiters. One waiter
cannot consume it away from another. Callers can release completed
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

A hint that is accepted but never persisted will vanish on restart. The model
followed the correction, the conversation moved forward, and the replay has
no idea why. Three new event types close that gap.

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

The temptation is to hold the lock across the HTTP call. The actor knows the
request, it knows the context, why not just wait for the response right there?
Because that is exactly the 2 AM bug. While the actor is blocked on
`client.Do(req)`, nobody is reading the mailbox. No hints, no interrupts, no
second prompt. The whole agent is frozen behind one network round trip.

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

At 2 AM, the only thing worse than not being able to type a correction is
watching the agent race through four more tool calls after the correction
finally arrives. Interruption means: stop this turn, save what you have,
and let the next instruction in. The Agent stays alive. Running jobs survive.
The conversation keeps its history. Only the current turn ends early.

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

## 5.6 Show progress without owning the answer

A GUI needs to see the agent thinking. A CLI needs to print tokens as they
arrive. Neither of these display consumers should be able to stop the agent
from working, delay a completion, or steal an answer from another caller. The
actor owns the facts; observers receive copies.

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

The whole point of this chapter is that a human can type while work runs.
The terminal is where that promise becomes real.

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

Build the CLI from `solutions/edition-2/main/`:

```sh
go build -o /tmp/ensemble-ch05 ./cmd
demo_dir=$(mktemp -d)
cd "$demo_dir"
CH02_LOG="$demo_dir/session.log" /tmp/ensemble-ch05 chat
```

Set `LLM_VENDOR`, a discovered `LLM_MODEL`, and the corresponding API key in
its environment before launch. Use a fresh log and scratch workspace; a shell
command is not confined to that directory. The current Gemini validation scope
starts at 3.0 Flash, with **Gemini 3.8 Flash** selected for hints. Discovery on
October 7, 2026 confirmed `models/gemini-3.8-flash` for
`POST /v1beta/models/gemini-3.8-flash:generateContent`. This dated test choice
is not a model allowlist in the library.

Start with `/help`, then ask for a command long enough to observe before it
finishes. Use the actual `ai_callback_delay` parameter: it controls the wait
for a report, not the command's lifetime. The following is an abridged actual
Gemini 3.8 session, driven by the coder through a PTY on October 7. Blank
repeated prompts and unrelated commands are omitted; this is not a session
Bill personally ran.

```text
You> Use run_command once with command "sleep 30; printf 'HINT-JOB-DONE\n'" and ai_callback_delay 45. After it returns, follow any hint I send and give a concise final answer. Do not call other tools.
Accepted r1.
You> /hint In your final answer include exactly HINT-ACCEPTED-FIVE and briefly mention the command output.
Hint received for r1 at seq 7; sent=false (pending next request).
Request r1 (success; pending hints=0)
Assistant:
HINT-ACCEPTED-FIVE

The command completed successfully with output `HINT-JOB-DONE`.
```

The coder waited for `tool_called` in the attributed log before sending the
hint. An earlier attempt sent one after the turn finished and correctly got
an idle refusal. A slow-looking command is insufficient evidence that the
Agent is still waiting; inspect the acknowledgement.

In the successful session, the first captured HTTP body has no hint. The
second contains the exact hint after `functionResponse`, as a separate text
part in the same user content. The third request omits it. The log retains
`hint_received` at sequence 7 and records its consumption at `request_sent`
sequence 10. The model's answer then supplies the fourth fact: it followed
this particular correction. Receipt, delivery, consumption and compliance
are observable separately.

Next ask for a long-running command, submit an ordinary second prompt while
it runs, and use `/interrupt`. The later portion of the same session shows
which caller gets which result:

```text
You> Use run_command once with command "sleep 180; printf 'LATE-JOB\n'" and ai_callback_delay 60. Do not call other tools.
Accepted r5.
You> Reply exactly QUEUED-FIVE; do not call tools.
Accepted r6.
You> /interrupt
Interrupt: request=r5 interrupted=true.
Request r5 (interrupted; pending hints=0)
interrupted: turn interrupted
Request r6 (success; pending hints=0)
Assistant:
QUEUED-FIVE
```

The excerpt omits two attempted context changes: `/ephemeral` and `/redact`
were both refused while busy. The ordinary second prompt remained its own
request. A subsequent turn inspected job 3 as running, then explicitly killed
it. Interrupting the turn had preserved the work and its handle, exactly as
promised. Finally, a separate job was left alive; `/quit` produced its durable
`job_killed` record with reason `shutdown`.

Repeat that sequence on each supported API, then test EOF in a fresh session:
submit `Reply exactly EOF-FIRST; do not use tools.` and a second prompt for
`EOF-SECOND`, then send EOF. All three recorded sessions completed both admitted
requests before final usage and cleanup. EOF stopped new input; it did not
reinterpret the second prompt as a hint or cancel it.

### A responsive answer can still hide a blocked write

The first implementation passed the new checker and the ordinary Messages
and Chat Completions control paths. It looked done. Then a comparative
review caught what the tests had not: `send_input` still wrote to the PTY
on the actor itself. A process that stopped reading could fill the kernel's
pipe buffer, and the actor would block trying to push bytes into a full PTY
while the human sat there, unable to type `/interrupt`, because the goroutine
that reads the mailbox was wedged on a write that would never complete.

Moving HTTP and report waits had solved the headline problem. This was the
same bug in a smaller pipe. The repair described in §5.4 makes input I/O
owned and cancelable too. Review also found that closed display subscriptions
retained client references and could be created after shutdown; §5.6 now
states their complete lifetime.

The corrected runtime was `959c663`. Its Gemini controls and focused Messages
and Chat Completions sessions started a line reader, sent `LIVE-SEND-CHECK`
with a newline, observed `seen:LIVE-SEND-CHECK`, and killed the process.
The durable input results report 16 of 16 bytes accepted. These small real
interactions prove the repaired user path; the full-buffer interruption,
partial-write and closed-subscription cases use independent deterministic
controls. A successful small write cannot prove cancellation of a blocked one.

The full [terminal record](../../solutions/edition-2/ch05/evidence/ch05/controls-gemini38-r1/terminal.txt)
and [receipt map](../../solutions/edition-2/ch05/evidence/ch05/verified-gemini38/receipts.json)
retain exact inputs and source bindings. The core controls measured:

| API and selected model | Source | Input | Cache write | Cache read | Output |
|---|---|---:|---:|---:|---:|
| Messages, `claude-haiku-4-5-20251001` | `8aa40c3` | 28280 | 0 | 0 | 613 |
| Chat Completions, `gpt-4.1-mini` | `8aa40c3` | 5247 | 0 | 7040 | 267 |
| generateContent, `models/gemini-3.8-flash` | `959c663` | 45025 | 0 | 0 | 1423 |

Returned identities were respectively `claude-haiku-4-5-20251001`,
`gpt-4.1-mini-2025-04-14` and `gemini-3.8-flash`; the logs retain requested and
returned names separately. The counters describe accepted responses in these
particular conversations. They exclude separate EOF, workflow, collection and
input sessions and are not a comparison of model efficiency.

### Let the three Agents work

From `solutions/edition-2/main/examples/workflow`, set a fresh
`ENSEMBLE_RUN_DIRECTORY` and the same safe model environment, then run
`go run . workflow`. Inspect `workspace/draft.txt` inside the run directory
and each role's log. The author writes a public-library Repair Cafe announcement;
the editor makes it warmer and adds `repairs are free`; the reviewer reads the
actual edited file. In the Gemini run, the editor replaced the opening sentence
and preserved `Saturday`, `10 a.m.` and `bring a broken lamp`. The saved file
contains the new opening and the free-repairs phrase. The final approval is
model output, not a substitute for inspecting that artifact.

Use another fresh run directory for `go run . collection`. Its observed output
was the same on all three APIs:

```text
Queued agent-1/r2: canceled
Collection agent-1/r1: COLLECTION-1
Collection agent-2/r1: COLLECTION-2
```

Only the two uncanceled requests reached the model. The consumer waits until
both completions are ready before draining them, then checks exhaustion,
individual handle reuse and a second independent collection. This deliberate
barrier establishes the collection rule without betting on scheduler timing.
The optional GUI remains a separate-module stub.

The earlier Gemini attempts remain useful failures. Some older-model replies
returned HTTP 200 and STOP but lacked the required output usage, so the strict
parser refused them. A bounded diagnostic succeeded with both the original
merged result/hint body and a split variant; it did not establish a grouping
bug. The four Gemini modes were subsequently demonstrated on the specifically
selected 3.8 Flash without changing the renderer to fit that hypothesis.
The retained Messages prompt also used an invented `timeout_ms` name; later
prompts use the actual schema. Failed attempts are not rewritten as successes.

The final demonstration set retains eight applicable Messages/Chat Completions
runs at `8aa40c3` and six affected-path runs at `959c663`. Its 14 logical runs
contain 78 captured requests, each reconstructed from its recorded prefix and
configuration with identical bytes. Those are captured HTTP bodies compared
with offline replay, not reconstructed bodies mislabeled as network captures.
The [review and gate record](chapter-05-validation.md) separates runtime
acceptance, evidence checks and the final manuscript/checkpoint. The independent
audit also corrected initial verifier-negative fixtures that stopped at an
earlier path guard. Thirteen isolated identity mutations then reached their
intended refusals before replay or derived writes, while the passing control
reproduced all 78 requests. The original masked fixtures remain in the record.
That correction required a local evidence audit, not another paid run.

---

Five chapters ago, this agent was 278 lines that could ask a model one
question and print the answer. Now it has an actor that serializes
conversation changes, a mailbox that accepts hints while the model is
thinking, request handles that let multiple callers get their own answers,
and a human at a terminal who can always type. The model is still
stateless. The agent remembers.

The next chapter adds the part the model cares about most: the ability to
see its own tools arrive and depart while it is running, and to carry
instructions that change what it can do.
