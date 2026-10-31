# Chapter 11 independent review

## Verdict

Accept. No verified material findings and no author revisions requested. Review followed the editor’s explicit confirmation that both manuscript and author carryover were stable; the final manuscript and carryover were then read in full.

## Basis

Independently read all of `book/companion/voice.md`, `book/edition-3/chapter-11.md` including exercises and ending, accepted `book/companion/chapter-10.md`, and `docs/companion-notes/chapter-10-author.md`. Recovered clipped chapter output with bounded reads. Final review covers `book/companion/chapter-11.md:1–103` and the full `docs/companion-notes/chapter-11-author.md`. No linked legacy documents were loaded.

The manuscript is a readable causal successor to Chapter 10, not a compressed assignment. Its opening and skill-lifecycle section distinguish persisted conversation facts from current executable capabilities. The central delayed-observation incident explains why callback return, completion, and turn ownership cannot be collapsed. Technical detail supports those decisions rather than becoming a replacement implementation checklist.

Material comparisons (source locations refer to `book/edition-3/chapter-11.md`; manuscript locations to `book/companion/chapter-11.md`):

- Historical wire and course extensions remain explicitly bounded: source 20–23; manuscript 7. Per-server grants and reverse-call trust are preserved without implying confinement: source 179–200; manuscript 15–21.
- Transport ownership, correlation, reverse-call deadlock, teardown, and job-context lifetimes survive in prose: source 39–106 and 156–177; manuscript 25–47. No second execution system or speculative abstraction is introduced.
- Hidden disclosure surfaces, reducer-mediated context, nonfatal failure, completed-result gating, and stale-turn rejection remain intact: source 110–153; manuscript 51–65. The 3.4-second peer and three-second callback are source facts, not invented measurements.
- Browser on-demand scheduling does not erase general automatic observations: source 5–9 and 207–236; manuscript 69–79. Snapshot size, truncation disclosure, control state, and observe/action/reobserve semantics remain explained.
- External server teardown remains distinct from retained instructions: source 264–281; manuscript 83–89. Exercise mechanisms survive without build commands, grading instructions, or code dumps; dedicated pipe service is not mistaken for model execution (source 302–305; manuscript 93).
- Historical provider and browser observations remain attributed and bounded, including the old automatic-browser path and local provider fake: source 450–468; manuscript 97–101. The correlated-result assertion failure is faithfully explained (source 470–476; manuscript 99). Earlier checks are not promoted into Edition 3 acceptance.

## Limits

This is source-grounded editorial review, not current implementation verification. No Git, code, builds, graders, browser, provider, runtime, or lint work was performed. Only this review note was written. No fix/recheck cycle was needed because the stable final candidate had no material findings.
