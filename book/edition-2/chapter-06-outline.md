# Second edition: Chapter 6 outline

Status: the coordinator accepted the full contract and its clarifications.
The fresh student has implemented from validated Chapter 5 and retained initial
real runs. Independent comparative review, broader acceptance and final prose
review remain open in [the gate record](chapter-06-validation.md). New Chapter 6
maps to first-edition Chapter 7.

## Stake and voice plan

Stake: a reader watching a long reply needs useful output before completion,
but must still know whether the displayed answer finished, failed, or was
interrupted before any proposed tool was allowed to act.

Open with the difference between seeing a file path emerge and permitting a
write. Preserve the first edition's terminal-flush example: a parser can emit
every fragment correctly while a buffered terminal hides the entire feature.
Use a deterministic held-open response to make that failure observable. Avoid
invented wait times, new Bill anecdotes, or claims that two live generations
produce identical answers. Explain the mechanisms before their schemas.

## Story-preservation pass, voice v5

Source: first-edition `book/chapter-07.md` opener and §7.9, reread October 7.
Retain the reader's blank-terminal frustration and the sharp explanation that
a buffered writer can do its job correctly while hiding the entire streaming
feature. The mechanism needs a timing check at the actual terminal, not a new
claim about how long an unmeasured model took. The current opener adds the
reader's ability to recognize a proposed bad write before it becomes an effect.

Omit historical model-capability tables and exact waiting times as current
facts. No new live stream or Bill anecdote is invented. The existing parser,
actor and terminal contract already carries the human stake, so the editorial
change stays in §6.7's explanation. Code blocks and tables remain unchanged.
Independent story proofreading is accepted; it does not validate code.

## Teaching order

1. Extend Chapter 5's existing public submission, control and observation
   interfaces. There is still one actor owning turn decisions and history.
   Streaming supplies earlier observations, not a second orchestration loop.
2. Separate provisional display, successful response acceptance, and reliable
   request completion. A part that looks complete does not authorize tool
   execution while the model response remains unaccepted.
3. Establish observable identity before the first fragment. Distinguish Agent,
   request, model-operation/response identity, parser-local part identity,
   final part position, vendor call ID, and committed response sequence.
4. Read SSE framing in llm with access to the actual owner chain and logger.
   Keep framing rules separate from each API's completion and error signals.
5. Normalize streaming and plain responses through shared semantic assembly.
   A single parser entry point encourages this; equivalence fixtures prove it.
6. Publish precise incremental payloads for text, exposed reasoning, and tool
   name/argument fragments. Preserve opaque content and signatures without
   treating every opaque block as displayable reasoning.
7. Let the actor accept the completed owned response once, assign actual
   append-time sequence/call IDs, update Engine accounting once, and publish
   final observations after durable acceptance. Background job facts continue
   to arrive during HTTP.
8. Handle truncation, malformed payloads, explicit provider errors, interrupt,
   close, and stale worker fragments. Mark provisional output incomplete and
   settle the existing reliable request outcome without inventing a success.
9. Make human chat visibly incremental, keep controls usable during streaming,
   and preserve explicit machine-protocol compatibility. Do not print the
   completed answer twice. A final-only public consumer remains legitimate.
10. Demonstrate actual terminal interaction on all three supported API surfaces,
    plus deterministic framing, identity, timing and failure controls. Compare
    the first answer with the historical standard only in the reviewer phase.

## Initial design questions, with resolutions below

- **Identity:** allocate an operation identity before parser output and carry it
  through fragments, accepted finals and abort observations. Do not predict the
  response's durable sequence: a job can append while HTTP remains in flight.
  The final observation should map the operation/part identity to the actual
  response sequence and final position. Parser-local identities need not come
  from the vendor or equal finished-list indices.
- **Ownership:** parser assembly belongs in llm; common holds declarations and
  owner interfaces. Workers return owned facts through the existing actor path.
  Do not import the old callback bundle as a substitute for owner access or
  let a parser append directly to Agent history.
- **Atomic acceptance:** recommend retaining Chapter 5's response acceptance
  boundary. Incomplete responses leave no accepted response, execution of their
  proposed tools, or successful usage increment. Earlier accepted rounds and
  running jobs survive according to the existing interruption contract.
- **Display recovery:** preserve Chapter 5's explicit subscription-overflow
  termination and reliable request handle. Specify how human chat labels a
  partial answer after overflow or stream failure. Defer full reconnect/snapshot
  handoff to the GUI chapter, without promising the future GUI is already done.
- **Capabilities:** define default/disabled/unknown delivery behavior separately
  from content preservation. Verify current model and API support before fixing
  a table; historical model rows are not current defaults. Do not invent chunks
  for an API that only provides a complete tool call or no reasoning text.
- **Trace:** decide whether this chapter introduces an opt-in sanitized frame
  trace. Parsed event payloads are not an exact HTTP/SSE byte capture. Existing
  parent-chain diagnostics remain required regardless of that choice.
- **Protocol:** publish any opt-in observation records and exact stream-control
  setting before a checker depends on them. Preserve Chapter 5's request IDs,
  acceptance/completion records, interruption outcomes and legacy user format.

The coordinator explicitly accepted Engine-owned operations and default-on
delivery with an explicit disable switch. This is a working design, not a new
Bill ruling. The draft fixes the remaining choices as follows:

- `EN_DISABLE_STREAMING=1` disables delivery; absent/0 enables it. Historical
  `request.delivery` absence means plain, including request reconstruction.
  Stream choices deterministically restore stream flags, usage option and
  endpoint, rather than adopting today's default.
- Public `model_begin`, typed-channel `part_delta`, `part_final`, `model_end`
  carry Agent/request/operation identities. Final mappings add actual durable
  response sequence and part position. Plain delivery has no deltas.
- `protocol --observe` explicitly opts into new observation JSON; default
  protocol remains unchanged. Human chat flushes provisional text, labels
  proposals/thinking, and avoids duplicate final text.
- No production frame trace is added. Static owner-reachable diagnostics are
  required; controlled fixture bytes remain local test evidence.
- The shared SSE reader handles standard line endings and multiline data,
  bounded to 1 MiB per event and 16 MiB assembled content. Each adapter supplies
  its own terminal/usage rule. No EOF-based universal success or automatic retry.
- Token-limit completion is distinct from network truncation. Validated parts
  follow the plain semantics; malformed argument objects still fail the whole
  response. Raw stop reasons are retained and generation limits displayed.

Independent review should check that these choices are internally complete
before a student or grader depends on them.

## Checks to derive from the eventual contract

Keep inherited grader numbering: proposed command
`make grade-dir CH=7 DIR=solutions/edition-2/main`. New independent checks use
Chapter 6 names. The coordinator must reconcile inherited fixtures with the
published contract while retaining legacy coverage.

- A held-open fixture sends a fragment, then waits for the observer/PTY to
  expose it before releasing completion. Counting fragments after exit cannot
  establish incremental delivery.
- Interleave two tool calls and text, reuse local part IDs on the next response,
  and run two Agents. Compare each group's fragments with its own final typed
  content using explicit identities, never guessed boundaries.
- Compare equivalent scripted plain/stream responses for full ordered parts,
  call arguments and IDs, opaque/signature preservation, provenance and usage.
  Exclude transport-specific delivery facts deliberately. Independent paid
  generations are usability evidence, not byte-equality fixtures.
- Framing cases cover split reads, CRLF, comments, multiline data and an
  unfinished final frame. API fixtures separately cover completion, error and
  usage ordering. Publish exact bytes before grading their treatment.
- Truncate or interrupt after text/tool fragments; require incomplete display,
  no premature tool side effect, no accepted response/usage, correct request
  completion, and a later usable turn after intentional interruption.
- Append a background job completion between the first fragment and model end;
  require deterministic actual response-sequence identities and one usage owner.
- Overflow a display subscription while a reliable completion waiter remains;
  require explicit loss notification and correct completion. Never infer
  completion from the final progress message.

## Live evidence plan

Use real human chat in a PTY on Messages, Chat Completions and generateContent,
with discovered model identities and capability-specific expectations. Observe
early text, then issue a control or follow-up after seeing output; demonstrate
tool intent before the accepted call executes and successful continuation.
Repeat in disabled mode without promising the same generated text. Record
actual first-observation timing and total timing without a speedup guarantee.
Label local truncated-stream and framing tests separately. Preserve all prior
human commands, usage, tools and cleanup behavior.

## Initial live reconciliation, October 7 local / October 8 UTC

Runtime `aa5f86a4782c7479ea61b0e6abcbd4163968b574` remained unchanged during
nine sessions: streamed human chat, plain human chat and a public two-Agent
consumer on each API. §6.8 now follows actual terminal use, with a short
Messages tool/interruption excerpt and the retained timing difficulties:
Chat Completions finished before the first hint opportunity; Gemini refused an
idle hint, reached a generation limit, and completed before one interrupt.
A later observed fragment permitted actual interruption and recovery. The
reader's opportunity to act resolves the opening's silent-terminal problem;
no universal latency or task-completion claim follows from these runs.

The public consumer demonstrates independent identities, complete finals and
reliable completion despite stalled subscribers. Guaranteed overflow and
thinking deltas remain local-fixture scope. Gemini's limited public responses
are accepted partial output, not completed explanations. The prose links the
ledger instead of printing all nine run counters.

The author checked all human terminals, public completion/observation data,
12 event logs, six read artifacts and the retained 33-request reconstruction
receipts. Detailed read scope and the student teaching response are in the
corresponding evidence and feedback files. The inherited 0/100 grader result
remains recorded as an incompatible historical contract, not a passing gate.

Next action: independent receipt/prose review and comparative findings; amend
the teaching before any affected student correction. Keep this initial runtime
and evidence identity. Development remains in `solutions/edition-2/main/`;
`ch06/` becomes a frozen export only after validation. No new checker command
is invented while the coordinator completes that acceptance surface.
