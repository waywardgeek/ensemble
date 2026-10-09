# Chapter 11 independent ownership/API plan review

October 8 PDT / October 9 UTC 2026. Reviewed the complete 439-line student plan
at `55c9e7a44364123dd28ef7967a0a3663d74de925`. Its accepted predecessor is
`edition-2-ch10-r1`, peeled commit `d91861207f4c2e4b81bd6c239ae15b8f51a1d146`.
The proposed structure is suitable for the next implementation handoff after the
three narrow decisions below are published and the root parent seam is amended.
This is neither implementation release nor runtime acceptance.

## Read boundary

Freshly loaded the entire mandatory coding skill, architecture and Chapter 11.
Refreshed Chapter 4's Job ownership/admission/terminal clauses, Chapter 5's slow
work/interrupt/close rules, Chapter 9's frozen ceiling and Actor commit rules,
and Chapter 10's session identity/current-authority clauses. Its complete contract
was also read during the immediately preceding accepted-chapter review. Current
voice and chapter procedure remain loaded.

This reviewer has historical implementation/grader exposure from earlier assigned
work. This architectural advice is derived only from the new published contracts
and the frozen student plan, not from a historical MCP answer or a grader's private
assumptions. No mutable draft/transcript was used as the final plan. No student
source, runtime test, build, credential or provider operation was performed.

## Accepted structure and invariants

The plan preserves the star: common declares data/interfaces, behavior stays in
responsible spokes, and the public root constructs them. Ensemble owns MCP service;
service owns Connections; each transport receives its actual Connection parent.
An application endpoint resource does not replace that parent. External transport
construction and owned complete-message Send/Receive/Abandon values are sufficient
public seams for a separate consumer without internal imports or pipe assumptions.
No public alias-bypassing RPC dispatch API is proposed.

Agent owns frozen bindings and Registry/Skills authority. Connection discovery
cannot grant aliases or mutate a live ceiling. The service attachment index refers
to those frozen facts, while dispatch checks real connection state and compatibility.
The proposed atomic attach/ready boundary, all-attached-binding checks on reopen,
and generation fencing address construction/replacement races. Preserve those
conditions when making the declarations concrete; snapshot status is a derived
view, not a dispatch authority.

Protocol issue and settlement remain Connection-owned. The 64 permits cover
staged delivery and cancellation as well as pending replies. Cancellation before
issue allocates no ID; attempted handoff burns one contiguous ordinal. Cancellation
or response wins once, and a permit survives unfinished delivery. The watermark
and bounded pending state classify stale replies without an ever-growing canceled
ID archive. One bounded delivery worker, queue-inclusive deadlines and transport
close keep a stopped peer from parking Actor or an owner lock.

The planned Job extension correctly distinguishes cooperative MCP work from the
inherited arbitrary local Go-function limitation. Agent close cancels and joins
its remote work without closing a healthy shared transport; ordinary interrupt
still leaves admitted Jobs alive. Root close bars new admission and forces blocked
transport work to unblock while Agents retain their shutdown facts. Fault workers
signal a separate coordinator rather than joining themselves. The implementation
must preserve this ordering and never claim an advisory remote cancellation undid
an external effect or forcibly terminated an arbitrary local function.

The strict v2 identity/state extension is consistent with §11.8: nonempty frozen
bindings agree with installed handler definitions, semantic schemas remain JSON
values, and absent output schema differs from false. Empty normalized remote
descriptions are legal. V1 keeps its exact old fields; raw replay wrappers and
all inherited scalar/number/correspondence bounds remain intact. Live Connection,
generation, cursor, process, credentials and diagnostics are excluded from saved
state. Two-stage resume validates stored history and logical/local compatibility
before new remote I/O, then compares prepared definitions before publishing an
Agent. Inert historical validation confers no execution authority. Failure releases
construction resources while preserving stored history and shared service ownership.

## Three decisions for the coordinator and author

1. **Standalone offline status is a teaching ambiguity.** Chapter 11 §11.9 asks
   offline watch to use recorded bindings, but §11.8 records them only in session
   v2 identity. Publish the narrow distinction: standalone offline logs return
   `mcp:[]`; they still render recorded calls/results and captured requests, but
   do not reconstruct logical bindings from tool names or create a live connection.
   Session v2 offline views use their recorded bindings with closed state, null
   generation and mcp_unavailable. No new standalone initializer or status event
   is needed. This preserves explicit CH02_LOG behavior and the required safe-array
   shape without fabricating authority.
2. **Shared lossless JSON is a routine ownership choice.** An Ensemble-owned
   common-interface service can serve both MCP and SessionCodec without sideways
   imports. If the behavior is factored beyond persistence, put generic bounded
   value parsing/canonical-number work in a responsible neutral spoke. Keep session
   raw-argument exceptions, accepted-byte preparation and session errors with
   SessionCodec; keep MCP profiles, schemas, bounds and errors with MCP. Do not
   introduce a second mutable JSON authority or silently relax either caller's
   policy. The proposed public `MCPRoot` currently exposes only Logf: the revised
   declared root parent path must expose the JSON capability where the internal
   helpers require it. A richer common parent interface may embed the public
   diagnostic interface; a second injected JSON field, fake Agent, or unchecked
   type assertion is not a substitute for that owner path. Publish the amended
   seam in the student plan/API documentation before affected implementation.
3. **Exhaustion controls may use disclosed valid owner fixtures.** Owner-local
   `_test.go` construction is appropriate for a valid exhausted generation cursor
   or contiguous issued prefix. It need not add production setters or make billions
   of calls. Explain why the parent fixture is valid, preserve pending/permit and
   generation invariants, and assert overflow refusal before mutation or transport
   creation/send. The grader can review this disclosed mechanism without imposing
   private identifier spelling. These controls complement the actual public/client
   tests; the oracle foundation remains preparation only.

These decisions do not require a new Bill ruling or historical answer advice.
The method names and public interface spelling remain student choices where the
published behavior and owner routes agree. No additional broad design expansion
is requested.

## Next gate

Record the coordinator choices in the new teaching/direct feedback, obtain the
student's revised parent seam and acknowledgment, and publish the actual runtime
checker invocation against the documented public APIs before integration or
acceptance. The reviewer has not run the oracle or granted runtime coverage.
Independent custom-transport, lifecycle, authority, persistence, inherited behavior
and real-interface checks remain the complete §11.10 requirement. The later live
matrix still needs explicit bounded review and actual all-provider evidence.

Existing external prose lint and scoped staged whitespace checks pass this record.
Only this review file is changed; the student plan and implementation remain owned
by the student.
