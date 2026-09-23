# Chapters

One line per chapter that exists or is planned. Titles are copied from each
chapter file's first heading. Candidates at the bottom are unnumbered on
purpose: no chapter count is fixed until a chapter is built.

## Written

- 0. The Perpetual Machine
- 1. One Loop, Sixty Billion Dollars
- 2. One Log, Three Vendors
- 3. Six Tools, Ninety-Two Percent of an AI Coding Agent
- 4. Jobs, or Why a Tool Call Is a Process You Supervise
- 5. The Big Refactor
- 6. Two Seams and a Loop
- 7. Streaming, or the Same Answer in Pieces
- 8. Everything Is an Artifact
- 9. Build Your Dream GUI
- 10. Skills
- 11. Persistence (TL;DR rewritten on the snapshot-anchor contract; body pending)
- 12. MCP -- The Extension Protocol
- 13. The Agent Sees Itself
- 14. The Channel Nobody Tested
- 15. The World's Best Context Engineering, Before Breakfast

## Planned

- 16. Memory (bands, save_memory, graduation chain; outline: chapter-16-outline.md)
- 17. Goal stack (needs a design session)
- Auto-recall (split from ch16: BM25 + vectors + relevance judge; number TBD,
  before or after the goal stack)

## Candidates

- **Sandboxing** (decided: one chapter). Owed by the ch15 design, Part V:
  identity-band write protection, the identity-write refusal check, the
  compressor sandbox profile. Also the capability rule that sub-agents need
  (a child may narrow, never widen), secrets that never enter context, and
  the ch3 writeguard grown into a policy.
- **Artifacts, deep dive.** Ch8 introduced the artifact; this goes further:
  renderers, addressable bytes, the content-addressed store that ch15 Q3
  needs, and sandboxed untrusted renderers.
- **Connections.** Connectors to outside systems (deferred from the GUI
  design): auth, credential flow, and what the agent may touch.
- **Gateway.** Channels (chat, webhooks, voice) and a scheduler: the agent
  acting with no human at the keyboard. Must follow sandboxing.
- **Workflows.** AI Native ch9-10: a skill is advice, a workflow is a
  program; then workflows the agent writes for itself. Also where the ch5
  three-agent exercise and the wake-once check were deferred.
- **Sub-agents.** Spawn, sync/async, structured submit that survives
  shutdown, join, blocking child-to-parent messages. "Submitted" is not
  "succeeded".
- **Self-wielding capstone.** The agent writes and grades a new chapter
  itself.

## AI Native chapters already covered here

- AI Native ch1 (goldfish): ensemble ch11 opener.
- AI Native ch4 (actors): ensemble ch6.
- AI Native ch5 (talking to an agent while it works): ensemble ch5-6 hints.
- AI Native ch8 (memory): ensemble ch16, planned.
- AI Native ch16-17 (the applied recipe, shipping in a quarter): no
  counterpart; this book's recipe is the agent itself.
