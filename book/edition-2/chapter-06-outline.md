# Second edition: Chapter 6 research outline

Status: author research and proposed teaching order only. This is not a student
contract or implementation handoff. Chapter 4 and Chapter 5 must establish their
validated predecessors first. No new streaming implementation or live run is
claimed. New Chapter 6 maps to first-edition Chapter 7.

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

## Proposed teaching order

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

## Proposed boundaries to settle in the full contract

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

These are design proposals for the full contract, not new Bill rulings. No
architectural ambiguity should reach the student as an implementation exercise.

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

Next action: consult the global reviewer, verify current official wire rules,
and turn these proposals into exact schemas/fixtures after predecessor work
has settled. Development will extend the accepted Chapter 5 source in
`solutions/edition-2/main/`; `ch06/` will be a frozen validated export.
