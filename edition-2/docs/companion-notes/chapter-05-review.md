# Chapter 05 independent review

## Disposition

**Accept the corrected manuscript**, “Let the Loop Listen.” The three verified source-fidelity issues below are corrected in the final manuscript and acknowledged accurately in the final author note. No material finding remains open.

## Reads and authority

Fully read the approved context: `book/companion/voice.md`; current `book/edition-3/chapter-05.md`, including all exercises, deferred work, and final demonstrations/evidence discussion; accepted `book/companion/chapter-04.md`; `docs/companion-notes/chapter-04-author.md`; the complete Chapter 05 companion and its author note. Recovered clipped output. The final sequential source recovery covered lines 1–120, 121–300, 301–480, 481–780, 781–1030, and 1031–1243. Rechecked the final corrected manuscript in consecutive reads through EOF and read the revised author note in full.

Current Chapter 05 source is authority for exact facts and policy. Chapter 04 supplies continuity, not permission to invent Chapter 05 requirements. No linked legacy process/voice/grading/execution documents or evidence archives were opened.

## Material findings and corrections

1. **Historical cause at the opening.** The original narrator anecdote attributed deafness to stdin executing the tool inline. The source (§5.4) instead says the tool already runs on its own goroutine while the conversation loop parks awaiting its report and cannot drain inbound events. The corrected opening now gives that actual mechanism without inventing an incident.
2. **Observation fields.** The original prose gave `StateChanged` a sequence number. The source declarations give it only `From` and `To`; `PartFinal` has `Seq`. The corrected observation paragraph preserves this distinction and keeps delta/finalizer identity separate from durable sequence.
3. **Part identity prescription.** The original persisted-history reseeding requirement for future part IDs was not established by this source. It has been removed; the useful explanation of identity correlating deltas with their finalizer remains.

All three were reported to the editor and checked after correction. A wider set of reviewer allegations about historical anecdotes and round-limit policy was erroneous and explicitly withdrawn: the current source supports the stale-child history in its concurrency discussion, hint and PTY-echo failures in its responsiveness discussion, default 200/zero semantics in its round-limit policy, and CLI metadata, word-count, media, grouped-completion-review, and Gemini-key stories in the final evidence discussion. They are not outstanding findings. A speculative human-default/stdout CLI requirement was also discarded because the approved source did not establish it.

## Assessment

The chapter reads as a successor TL explaining consequences, not an exercise checklist. The causal spine holds: mailbox ownership, actual message delivery, truthful batch closure, independent lifetimes, and observation without control. The jokes are brief and useful; they do not substitute for mechanisms.

Technical coverage includes the easily missed exercise essentials: reliable admission and private request replies; save-before-reply; hint versus interrupt versus kill versus shutdown; HTTP-final hint continuation; running-job wakeup and same-handle resumption; completed/dispatched/undispatched batch accounting; literal-next-call limits; timer retirement and output-condition distinctions; nonblocking/lossy observers and isolated snapshots; round-limit policy; legacy/live routing; synchronized workflow artifacts, separate histories, interruption and EOF; media capability refusal and actual attachment delivery. It correctly separates progress from completion and leaves streaming, URI/handle rendering, spawning, and grouped completion outside the implemented claim.

Historical demonstrations remain historical, feature-specific evidence. The manuscript does not turn receipts into current live claims or imply every feature ran on every vendor. Its retained quantitative policy is source-supported.

## Real checks and limits

This was a textual, source-grounded review, not execution verification. No code inspection, build, tests, graders, provider calls, Git operations, or archive investigation. Wrote only this review note. No current runtime behavior or independently reproduced historical result is claimed. Available for a bounded recheck if the manuscript changes materially.
