# Chapter 1 independent review

## Verdict

**Accept. No material findings or required revisions.**

The chapter carries Chapter 0's inheritance-of-judgment premise into concrete architecture. The late logger supplies a grounded opening rather than a manufactured anecdote; the payload-placement and clean-EOF stories illuminate distinct decisions. The prose is accessible to an AI-coding-fluent successor lead without becoming a beginner tutorial, checklist, or implementation assignment. Its light touches (the shared `content` field's “small act of misdirection”; a program proving it can stop rather than converse) support rather than interrupt the explanation.

## Reasons for acceptance

- Ownership is explained causally, not merely named. Runtime parent relationships are distinguished from imports and service dependencies; `common`, implementation behavior, configuration, usage, the composition root, and deferred Observer/GUI responsibilities have intelligible homes. The pin example preserves the source's warning against forcing every relationship into a single-parent model.
- Request control retains both motivations: timely steering through precise block placement and prefix stability through volatile-context placement. Framework criticism is qualified rather than universal. Historical hint timing is explicitly not a current vendor guarantee.
- The successful-turn commitment rule explains why history must not acquire failed user turns. Complete user **and assistant** history, synchronous ordering, wire statelessness, narrow message invariants, request fields, explicit model choice, typed response blocks, stop reasons, and one-call scope survive the transformation out of exercise form.
- “History describes successful dialogue; accounting describes measured expenditure” correctly separates conversational acceptance from measured usage, including a reply with no usable text. Missing usage is not fabricated as zero. Timeout policy preserves cancellation and connection/TLS protection without equating slow generation with a stuck request.
- Exercise-buried interface knowledge is retained: stdout is protocol-only; front ends share Agent; large lines must not inherit Scanner's accidental limit; data returned with EOF is processed once; non-EOF read failure must not masquerade as normal completion with terminal usage. Controlled history/block/accounting checks, actual chat, and model self-report answer different questions.
- Quantitative reasoning is appropriately bounded. Roughly quadratic cumulative context is conditional on similarly sized turns and unpruned history, not a dollar estimate. Historical percentages, prices, and transient rankings are omitted. Reported stories stay attributed; deferred capabilities remain future boundaries, not demonstrated implementations. The explicit distinction between stateless requests and provider retention avoids an unsupported privacy assurance; diagnostic guidance avoids credentials and remote error bodies.

## Reading and actual check scope

Fully read, in the requested sequence: `book/companion/voice.md`; all 712 lines of `book/edition-3/chapter-01.md`, including exercises and preserved interaction; accepted `book/companion/chapter-00.md`; `docs/companion-notes/chapter-00-author.md`; current `book/companion/chapter-01.md`; and `docs/companion-notes/chapter-01-author.md`. Recovered the clipped middle of the current chapter before deciding. Compared actual source coverage and attribution, not merely the author's coverage summary.

The source showed no working-tree diff during review. Reviewed draft SHA-256: `befceee88e7329992e278602236d3503d5d0571994a3683f654216295e094536`. Reused the author note's reported final lint exit 0/all categories OK for this unchanged draft; did not independently rerun lint or treat it as an editorial acceptance criterion.

This was source-grounded editorial and technical-reasoning review, not independent reproduction of historical incidents or verification of current implementation, API behavior, pricing, or live service results. No broad builds, graders, live calls, legacy documents, or student pipeline were used. Only this review note was written; no manuscript edits, staging, commits, or resets.
