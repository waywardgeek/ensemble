# Chapter 0 learning revision: author handoff

## Choice and exact changes

Added one six-paragraph section, **What survives the session**, between the accepted opening's failed knowledge-transfer stories and **An expert who stays in the loop**. The existing failure supplies the narrative connection; no new incident, dialogue, or result was invented. The section explains ordinary inference's fixed trained weights, useful in-context adaptation and persistent external notes, the limits of retrieving unrecorded reasons, and the separate possibility of fine-tuning or online training. Human sleep consolidation is an analogy about persistence, explicitly not an equivalence of biological and model-training mechanisms.

The explanation connects broad model knowledge to missing accumulated project experience, then places responsibility on the human lead to preserve judgments for both future sessions and human successors. It distinguishes an agentic codebook's executable knowledge from a lead's need for rationale, tradeoffs, and grounds for changing a design. Bill's intended “first example” is presented as an intention for this companion, not a claim of historical priority or a permanent human monopoly on judgment.

Moved the agentic-codebook definition into this section and removed its duplicate sentence from **When the teaching becomes the bug**. No other accepted chapter prose changed. The approved preface, original author/review records, source chapter, and unrelated work were left untouched.

## Sources

Read fully: `book/companion/voice.md`, accepted `book/companion/chapter-00.md`, `book/edition-3/chapter-00.md` (recovering clipped output), approved `book/companion/preface.md`, and `docs/companion-notes/chapter-00-author.md`. The authorized smaller reading set replaced the legacy-document requirement.

Bill's new explicit explanation in the revision request is the primary source for this addition, especially the sleep analogy and the intended translation from executable knowledge to human education. Edition 3 already supplies the failed handoff and a memory/learning contrast; the revision does not pretend its new framing or sleep explanation came from that chapter. The existing relative source link remains unchanged and valid. Historical stories retain their prior attribution rather than becoming independently verified experiments.

## Checks and carryover

Reread the saved chapter, recovering the clipped unchanged passage. `go run ./cmd/lintprose -v book/companion/chapter-00.md` exited 0 with every rule OK. `git --no-pager diff --check -- book/companion/chapter-00.md` exited 0. Python checks confirmed no trailing whitespace, a final newline, and a resolving existing relative source link. The first Git-check command was blocked by its terminal pager; it was replaced by the successful non-paged run and its waiting processes were stopped. No product tests, builds, graders, or live-agent tests were run for this prose-only change.

Independent editorial review remains for the parent. Preserve the distinction between useful contextual/external memory and automatic durable changes to trained parameters; neither “notes are worthless” nor “models can never learn this” is the claim. Later authors can build on this explanation without repeating the sleep analogy or the preface's motivation. The existing transition to initial architecture remains intact. No Git writes were performed.
