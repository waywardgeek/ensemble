# Second edition: Chapter 5 outline

Status: contract, architecture and revised runtime independently accepted.
Actual all-three-API human CLI and public-consumer demonstrations are reconciled
below. Final manuscript proofreading and the coordinator's frozen export/tag
are tracked in [the validation record](chapter-05-validation.md); Bill's
editorial approval remains separate.
This chapter maps to first-edition Chapter 6.

## Voice plan and stake

Stake: a reader notices the agent acting on the wrong instruction while a
request or tool is still running, and needs the correction acknowledged
without losing the work or receiving somebody else's answer.

Open with the stalled control path rather than a claim that synchronous
interfaces necessarily serialize independent Agents. Chapter 4 already runs
jobs asynchronously; the missing ability is to process input while awaiting
a model or job report. Explain reliable replies with two simultaneous callers
whose answers cannot be identified by a shared Idle notification.

## Story-preservation pass, voice v5

Source: first-edition `book/chapter-06.md` opener and actor discussion, reread
October 7. Retain the desire to correct work while it is still happening and
the memorable two-loops/two-owners explanation. The opening now states the
human consequence: an instruction accepted only afterward becomes a review
comment. The blocking API stays familiar by entering the same actor.

Omit the unsupported exact weeks spent separating the historical GUI, the
old prediction that the GUI chapter comes immediately next, and the claim
that synchronous APIs necessarily serialize independent Agents. The new
architecture already separates GUI ownership and allows background jobs;
the missing capability is admission during an HTTP/report wait. The later spin section adds the observed blocked-input defect and its repair,
plus actual hint delivery, interruption and collaborative editing. These come
from retained receipts, not an invented scene. Normative contracts and fixtures stay
unchanged by the editorial pass.

## Teaching order

1. Preserve the already-correct ownership, optional GUI seam, and human chat.
2. One Agent-owned actor serializes turn decisions and conversation changes.
   Engine retains transport/accounting; Jobs retains work/output ownership.
3. Mailbox enqueue order, reliable prompt identity, queued admission, and
   explicit hints are different from display observations.
4. Model I/O, report waits and process-input I/O return owned facts without parking the actor.
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

The inherited 10/100 diagnostic reflects incompatible historical expectations,
not acceptance of the new contract. The manuscript now publishes
`scripts/edition2/accept_ch05.py` and `scripts/edition2/audit_ch05_mutations.py`
as the separate new-contract checks; the old grader remains unchanged.
Prove actual replay equality, request isolation,
FIFO admission, cancellation and post-interrupt reuse, close settling callers,
late job persistence, hint timing and actual later request inclusion, append
sequence correctness during HTTP, and completion collection without polling.
Do not treat old labels as evidence for stronger properties than they inspect.

## Contract review and observed revisions

The manuscript now publishes lifecycle payloads and same-path replay rules,
immutable request snapshots, hint projection/consumption, queue and observation
overflow behavior, and the deterministic completion-drain boundary. It preserves
existing media mappings without claiming an unknown text model is unavailable.
Independent findings about stopping-state spelling, new protocol outputs,
report-cursor acceptance, and interrupted/canceled versus fatal CLI outcomes
are addressed and accepted. Coordinator architecture review accepted Agent-owned
Actor, Engine remaining an Agent child, Jobs retaining state and cursors, Actor
serializing history, and Ensemble-owned reliable completion collections.
The initial student and revised runtime have completed this handoff. Comparative
review found blocking process input and retained observer state after shutdown;
§§5.4 and 5.6 teach the corrected ownership and lifetime explicitly.

## Demonstration and checkpoint

The actual spin retains eight applicable runs at `8aa40c3` and six affected-path
runs at `959c663`, including all four Gemini modes on discovered 3.8 Flash and
Messages/Chat Completions input reruns. The 14 logical runs contain 78 captured
requests with exact replay. Human PTYs were driven by the coder, not Bill.
Public workflow and completion-collection runs use the external executable.
The ordinary live writes do not establish the deterministic full-buffer bound.

Older Gemini refusals, an invented prompt argument, the original blocked-write
implementation and masked verifier controls remain evidence. The corrected
independent evidence audit has a passing control and 13 intended identity
refusals. Reader links target the forthcoming frozen `ch05` export. Source
`185ba76` includes the `469730f` evidence freeze and documentation cleanup;
runtime stays `959c663`. Coordinator export/tag follows final proofreading.
