# Second edition: Chapter 5 outline

Status: full contract/prose independently reviewed; all reported findings
resolved. Coordinator architecture review accepted the working choices.
Await predecessor gates before student handoff.
This chapter maps to first-edition Chapter 6. Human-chat correction work in
Chapters 2–4 remains a separate active gate; no Chapter 5 code or live run.

## Voice plan and stake

Stake: a reader notices the agent acting on the wrong instruction while a
request or tool is still running, and needs the correction acknowledged
without losing the work or receiving somebody else's answer.

Open with the stalled control path rather than a claim that synchronous
interfaces necessarily serialize independent Agents. Chapter 4 already runs
jobs asynchronously; the missing ability is to process input while awaiting
a model or job report. Explain reliable replies with two simultaneous callers
whose answers cannot be identified by a shared Idle notification.

## Teaching order

1. Preserve the already-correct ownership, optional GUI seam, and human chat.
2. One Agent-owned actor serializes turn decisions and conversation changes.
   Engine retains transport/accounting; Jobs retains work/output ownership.
3. Mailbox enqueue order, reliable prompt identity, queued admission, and
   explicit hints are different from display observations.
4. Model I/O and report waits return immutable facts without parking the actor.
5. Define interruption, queued cancellation, accepted calls and late job facts,
   request completion, and close independently; give literal event fixtures.
6. Teach hint application at a later request, never retroactive model steering.
7. Preserve replay/request equality and actual append-time call identities.
8. Reliable completion collection across Agents, with a defined drain boundary.
9. Human chat controls, external three-Agent workflow, all-three-API live use,
   targeted mutation controls, independent comparison and revision.

## Working choices

Actor belongs to Agent and is implemented in llm. Its shared state/message
declarations and parent interfaces belong in common. Engine remains another
Agent child; Actor reaches it through Agent. There is no extra coordinator
object duplicating Ensemble, no sibling imports, and no callback dependency
bag. Internal worker results retain their owning operation and logger path.

Retain the existing sixteen-request turn limit. First-edition Chapter 6's
later-added 200-tool-batch setting is evidence for one configuration owner and
complete refused-call pairing, not a reason to change two units in this chapter.

Public blocking submission is an adapter over the same queued request path
used by asynchronous callers. Completion is attributable, reliable, and
readable by multiple waiters. Observer is progress, not a completion mailbox.

## Checks and numbering

Use `make grade-dir CH=6 DIR=solutions/edition-2/main`: CH=6 selects the
inherited grader, while the manuscript and new independent checks use 5.
Coordinator confirmed this mapping rather than renaming historical graders.
The later `ch05/` directory is an exact validated export, not a working copy.

Retain inherited checks, then prove actual replay equality, request isolation,
FIFO admission, cancellation and post-interrupt reuse, close settling callers,
late job persistence, hint timing and actual later request inclusion, append
sequence correctness during HTTP, and completion collection without polling.
Do not treat old labels as evidence for stronger properties than they inspect.

## Remaining contract work

The manuscript now publishes lifecycle payloads and same-path replay rules,
immutable request snapshots, hint projection/consumption, queue and observation
overflow behavior, and the deterministic completion-drain boundary. It preserves
existing media mappings without claiming an unknown text model is unavailable.
Independent findings about stopping-state spelling, new protocol outputs,
report-cursor acceptance, and interrupted/canceled versus fatal CLI outcomes
are addressed and accepted. Coordinator architecture review accepted Agent-owned
Actor, Engine remaining an Agent child, Jobs retaining state and cursors, Actor
serializing history, and Ensemble-owned reliable completion collections.
Predecessor validation remains before handoff. Scoped prose lint has no hard failures.
