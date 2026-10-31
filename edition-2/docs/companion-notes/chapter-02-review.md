# Chapter 2 independent review

## Verdict

Accept the current `book/companion/chapter-02.md`. No material correction required. This is an independent source-to-companion review, not adoption of the author's verdict.

The revised opening is accurate. “An interface was extracted from the Claude client, taking its existing methods as the requirements” preserves the source's actual failure: abstraction of a Claude-shaped client, not failure of an already-neutral conversation model. The subsequent paragraph explicitly retains the big-bang migration tail, including capabilities stranded in old clients and the reported broken audio attachments. “A cleaner seam had not completed the migration” makes the distinction between architectural improvement and delivered capability unmistakable without turning a historical audit into a present-status claim.

## Reasons for acceptance

The explanation follows ownership and information loss rather than merely listing rules. Chapter 1's narrow role/text history leads naturally into durable events, reduced context, and disposable requests. History, Agent, Engine, parser, and renderer have distinct responsibilities. The ordinary turn establishes the mechanism before the chapter introduces interleaved tool results, missing returns, and wire-only recovery. Requested tool calls remain separate from dispatch; a missing return is not invented execution evidence.

Consequential knowledge from the source exercises survives without their assignment machinery. In particular, ephemera is logged, replaceable, cleared by an empty update, rendered before send consumption, and consumed even on a failed exchange. Unsupported media is refused without unauthorized dereferencing. Redaction preserves references and durable facts; summary replacement folds a span rather than repeating prose per entry. The chapter distinguishes intended compaction policy from machinery already delivered.

The difficult seam details remain legible: result adjacency versus generic alternation, ordered multi-call grouping, empty versus absent text, exact JSON numbers, object-valued Gemini failures, actual producer identity versus routing aliases, call-bound signatures, and dependency-aware partial usage accounting. The first implementation's additional signature field and the later implementation's reuse are distinguished. These details give a successor TL reasons to question an apparently successful demo rather than merely instructions to trust an abstraction.

The historical stories retain causal substance and attribution. Dated signature probes establish reported HTTP acceptance, not backend use or current compatibility. Volume measurements motivate category-based compaction without becoming universal percentages. The prose also names the counterweight: tool results can contain indispensable evidence. Offline fixtures, live observations, stored signatures, parser capture, and internal Ref preservation are treated as different evidence, not interchangeable proof.

The third-person body is technically substantial but approachable. Concrete failures and restrained humor carry the difficult material; the chapter contains no TL;DR, assignments, grading recipes, contract tables, or code dumps. Its length is doing explanatory work. No shortening or stylistic churn is warranted for acceptance.

## Reading and evidence limits

Fully read the authorized smaller set: `book/companion/voice.md`; all 1,427 lines of `book/edition-3/chapter-02.md`, including the complete exercise, supplied log, checks, and reported live runs; accepted `book/companion/chapter-01.md`; `docs/companion-notes/chapter-01-author.md`; current `book/companion/chapter-02.md`; and `docs/companion-notes/chapter-02-author.md`. Recovered clipped tool output with bounded reads. Independently checked the revised opening against the source.

This review establishes textual fidelity, continuity, and editorial/technical adequacy within that set. It does not independently verify historical incidents, implementation behavior, present provider semantics, prices, or live API results. The author's targeted lint and whitespace checks were reported successful; they were not rerun or treated as technical proof. No implementation inspection, broad audits, builds, graders, live calls, manuscript edits, or Git writes were performed. Only this review note was written.
