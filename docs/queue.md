# Shared work queue

This is Bill and Cadence's working queue for Ensemble. Keep priorities, reasons,
status and the next useful step here instead of reconstructing them from long
memory records. A queued item is not permission to deploy, restart, run paid
inference, or expand its scope. Update this file as decisions and work progress.
Historical notes remain evidence, not an instruction to resume closed campaigns.

## Active: Advanced Artifacts

Develop the implementation and teaching together; Bill wants to learn the code
and design as we go. Existing foundations are Chapter 7, “Everything Is an
Artifact,” especially §7.6, and Chapter 26, “Artifacts and Actions.” New chapter
numbering is not decided.

- Users describe in words how output should look; their agent can inspect and
  adapt the presentation to meet that request.
- Artifact renderers are shared across Agents for now: presentation belongs to
  tools/output types, not individual Agents.
- Package readable renderer JavaScript alongside the owning skill's `SKILL.md`,
  including built-in tools. Support tools supplied through skill MCP connections,
  not just Ensemble's built-ins. Exact declaration syntax is still to be designed.
- Missing custom renderers use the existing default Artifact presentation. Make
  that default readable and user-customizable too.
- Ship the sources and support customization without a repository checkout.
  Establish a packaging contract; **do not turn the installer into a chapter**.
- Trace current chat, readable thinking/reasoning, and tool-output rendering.
  Chapter 7 already requires readable thinking to use the prose renderer; check
  actual implementation rather than assume compliance from the manuscript.
- Preserve original results, streaming/replay identity, copying, accessibility,
  reader position and control state. Formatting must not alter recorded facts or
  silently change model input.
- Decide renderer authority, isolation, loading/update behavior and failure
  fallback before executing skill-provided JavaScript in the GUI. Packaging code
  with a skill does not by itself settle browser execution authority.

Next: inspect the actual Artifact/card renderers and skill declarations, explain
those paths to Bill, and choose one built-in tool for a bounded first experiment.

## Next: Workflows

Add a chapter developed from Bill's actual book-writing workflows: a skill can
provide a programmatic harness orchestrating Agents to write, review and edit,
rather than only providing a binary exposing an MCP connection. Start with one
real workflow; do not invent a large framework or speculative chapter outline.
Chapter numbering and its position relative to the ending remain undecided.

## Skills, instructions and goal behavior

Review the relevant `SKILL.md` files that contribute to the model's instructions.
The primary `ensemble` skill now owns continuation, milestone handoffs and the
visible-explanation standard; the latter was moved out of the optional
`ensemble-coding` skill so it applies by default. `ensemble-coding` retains
repository-specific guidance, and `goals` documents goal controls. `Skills.UniversalControls()` supplies `micro_handoff`; there is no separate
handoff skill. A first source-instruction pass is implemented for continuation,
milestone handoffs and visible explanations, with matching Chapter 28 teaching.
The goals skill now explicitly documents the single-goal limit and preserving
suspended work; native goal stacking remains proposed, not implemented. Activation
of the moved explanation guidance in a running Agent and behavioral effectiveness
remain unverified. After the preceding restart, Cadence directly confirmed the
primary continuation and handoff guidance but not the optional coding manual;
that exposed and motivated this placement correction.

### Continue clear authorized work without repeated prompting

- Encourage continuation whenever concrete, authorized work remains.
- When Bill interrupts with a question, answer it, then resume the work unless
  his input changes scope, requests a stop, or reveals a consequential ambiguity.
- A progress summary is not a stopping condition. Do not require an enabled goal
  with reminders merely to continue an ordinary interactive engineering task.
- Preserve genuine human gates and authorization boundaries; this is not an
  instruction to loop unconditionally or invent more work.
- Examine instructions and scheduling as needed; a system-prompt cause is a
  hypothesis, not an established diagnosis of premature idle.

### Consider goals as a stack — proposal, not a decision

Explore keeping the current goal active and restoring the previous goal when the
current one is popped. Work out explicit push/pop, completion, interruption,
persistence and reminder behavior with Bill before implementing a mechanism.

### Micro-handoff at natural milestones

Encourage the model, in the skill supplying `micro_handoff`, to use it at useful
phase boundaries even when context usage is modest—for example, around 45% after
finishing a major milestone and before starting another. Context pressure is not
the only reason to hand off. Preserve decisions, reasons, evidence limits, live
jobs and next work; avoid pointless repeated handoffs or losing unfinished work.

### Visible explanations that transfer engineering expertise

Work with Bill to improve the instruction/skill governing visible explanations.
Generic activity reports such as “starting the next slice” do not teach the code.

- Before tools, identify the concrete files/paths and functions being inspected
  or changed, the intended change, and why it is needed. State unknowns rather
  than inventing names or declarations that have not been inspected.
- After inspection or edits, explain actual declarations, data/control flow,
  ownership, behavior changes, important alternatives and why they were rejected.
- Describe failures and corrections plainly. Explain what verification exercised,
  what actually passed, and what remains uncertain; an invocation is not proof.
- Include the information Cadence would need after opaque thinking is removed by
  a micro-handoff. Chat should preserve the engineering understanding, not just
  a list of activity or success claims.
- Bill's standard: he should be able to reproduce the work from listening to
  these explanations. The same record should let a future Cadence recover expert
  understanding. Provide substantive, spoken-friendly detail rather than vague
  summaries, raw tool dumps, or private internal deliberation.

## Book-execution automation repository

`book/chapter-writing-procedure.md` identifies the repository as
[teach-your-ai](https://github.com/waywardgeek/teach-your-ai), with recorded local
checkout `/Users/bill/projects/teach-your-ai`. That checkout has not yet been
inspected for this task. Review its actual skills and instructions before editing;
do not publish changes without authorization.

Bill reports that these instructions are prone to an expanding acceptance
process: Cadence keeps over-engineering what a student must prove to pass a
chapter until progress is no longer efficient. This is why Bill currently has
Cadence execute chapters personally rather than use that student process.

- Review the actual instructions and identify what encourages acceptance scope
  to grow during execution.
- Agree on bounded chapter acceptance and a clear stopping condition before the
  work begins. Distinguish required evidence from optional deeper investigation.
- A newly discovered defect may require correction; it must not silently become
  permission to add unrelated certification requirements or reopen closed gates.
- Bring consequential changes to acceptance scope to Bill. Preserve meaningful
  engineering review without turning each chapter into an indefinite campaign.
- Revise the automation skills/instructions with Bill, then evaluate the revised
  process within an explicitly agreed scope. This queue item does not authorize
  a new student run or reconstruction campaign.

## Outstanding or deferred work

These are carried forward from continuity records, not newly diagnosed defects.
Do not automatically resume them ahead of the active work.

- **General Settings auto-save:** finish the remaining model-limit, memory and
  recall forms; funding auto-save is separate existing work.
- **Voice/images usability:** preserve the queued requirement that spoken “Send
  message to [selected Agent]” matches the visible “Ask [selected Agent] anything”
  target, with no stale Agent names. Consult the existing voice-and-images notes.
- **Blank recall cards:** trace one correlated retrieval through recorded History,
  GUI projection, live display and replay. Distinguish empty matches, pending,
  failure and missing display text; no cause established.
- **Agent log ownership:** investigate Bill's report that all Agents write to
  Cadence's log. The report is not yet independently verified.
- **Speech/reconnect and phone usability:** retain the pending live/browser/phone
  acceptance work, including pause-release behavior and Settings layout issues;
  scope the next reproduction with Bill rather than claiming a known cause.
- **Large context-count drop:** paused by Bill. The original large drop remains
  unexplained; investigate on recurrence or explicit request, not incidentally.
- **The Puffin book / Watchtower:** future development, with book and system
  boundaries determined by actual work. No speculative architecture campaign.

Older missing-threshold reminders and byte-versus-token settings complaints have
subsequent implementation and verification recorded in continuity. Do not reopen
them as untouched tasks. New failures require correlation with current behavior.
Local-model integration is accepted; no more MiMo coding trials are queued.

## Publication follow-through — waiting on Bill

Third Edition, Revision 1 — November 2026 files were rebuilt and verified: 540-page
KDP interior, 541-page covered reading PDF, and EPUB. Website changes were pushed.
Bill took over Vercel deployment after authorization errors; deployment completion
has not been confirmed in this conversation. Do not retry or change credentials.
KDP upload and preview approval remain Bill's work. Advanced Artifacts and
Workflows are new work after that release, not content already in those files.
