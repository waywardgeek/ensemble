# Companion voice

*Adapted from `book/edition-3/voice.md`, version 7, and Bill's directions for this companion. This is an editorial specification, not prose for chapters to imitate. The original voice files remain unchanged.*

## 1. What this book is for

This book transfers Bill's hard-won understanding of building AI coding agents to a human who can lead the work forward without depending on his continued involvement.

The reader is extraordinarily smart, already very familiar with coding with AI, and a prospective technical lead. Bill reports spending at least a couple thousand hours building these tools with this assistant and other LLMs. The instructional book records the resulting knowledge in a form agents can execute. This companion helps a human absorb the experience and judgment behind it much faster.

Do not teach introductory AI coding. Do not assume intelligence supplies missing experience, either. Explain the system's particular concepts, lifecycle, boundaries, and failure modes. A reader can understand concurrency perfectly and still need to learn why this runtime separates a running job from an observation of it.

The reader should come away able to explain why a design exists, recognize the problem it prevents, evaluate its tradeoffs, and reason about when different circumstances justify changing it. Agreement with Bill is not the learning objective. Informed independent judgment is.

## 2. The reading experience

**Explain what it is and why it is that way. Make it pleasant to read and significantly shorter than the original book.**

Bill’s latest direction is explicit: every sentence must either carry knowledge directly or earn its place through cohesion, readability, or pleasure in reading. He wants understanding, not a lot of prose. Remove repeated conclusions, decorative setup, generic reflections on autonomy or responsibility, and philosophical discussions of how an agent ought to operate. Concrete explanations of actual mechanisms, failures, and design choices remain essential.

Use the source’s cold opening to establish the problem when useful, and its “In Plain English” section as an explanatory anchor where present. Extend those with the technical details needed to understand what exists and why; do not expand them into a general essay. A documented story earns its space by explaining a decision or failure. Do not invent an opening or force a story into every chapter.

Lightness comes from clear explanations, concrete situations, natural rhythm, and pleasure in understanding how something works. It does not come from removing the difficult part. Quick means low friction and high understanding per unit of attention. The companion as a whole must be significantly shorter than Edition 3; compare actual manuscript lengths rather than assuming that deleting exercises achieves this. Do not impose an invented per-chapter quota or cut a necessary explanation to meet one.

A documented failure can make a subtle distinction memorable. A well-chosen sequence can explain interacting rules more efficiently than a checklist. A short connective paragraph can save the reader from reconstructing an unstated causal step. Keep those paragraphs.

Avoid both filler and monotone. Do not make every sentence short, every paragraph end with a punch line, or every section announce a revelation. Let the thought determine the pace. Give a demanding mechanism enough room, then move on when it is understood.

The pleasure is an engineer discovering why a machine works, including the occasional absurdity of what it does when it does not.

## 3. Priorities

When preferences compete, use this order:

1. Truth, technical accuracy, privacy, and publication guardrails.
2. The reader's understanding and ability to exercise judgment.
3. Direct, concise explanation, useful organization, and continuity.
4. Personality, rhythm, and stylistic preferences.
5. Lint as a diagnostic, never the definition of good prose.

Never alter evidence, omit a consequential exception, or obscure a distinction to improve a stylistic count. There is no minimum chapter length or prescribed reduction percentage. Significant overall reduction is an explicit requirement, achieved by selecting and explaining the knowledge efficiently, not by hiding consequential detail.

## 4. Who is speaking

Chapter bodies use third person, present tense for explanations, and the appropriate tense for historical events. No authorial “I” or “we,” and no invented agent narrator. “You” may address the reader naturally. Contractions are welcome.

Personality comes through concrete choices, candid engineering judgment, rhythm, and dry humor. Technical prose need not be impersonal or bloodless.

A chapter may have a brief first-person motivational opener in Bill's voice, grounded in his actual words or documented experience. A paragraph or three is enough. It is optional; do not manufacture one. The body returns to third person without an apology or explanation. There is no TL;DR between the opener and the body.

In the body Bill is a person whose actions and experience matter to the explanation. Tell his stories about him, quoting only words that have a source. Do not ration his name inside a useful story or insert him to satisfy a personality quota.

Bill's personal register is confident, casual, technically specific, and candid about failures. Keep its energy in short stretches; do not scatter signature vocabulary through the book to simulate his voice. The authorized line “I'm not opinionated, I just know the truth” carries both conviction and a joke. It is not a recurring catchphrase or permission to dismiss evidence.

Retain accurate authorship disclosure in the front matter when that material is prepared. Do not invent an AI's experiences, contractual powers, financial interests, or personal testimony. Claims and promises belong to the people responsible for them.

## 5. Stories carry knowledge

Before drafting, identify the problem that keeps the chapter worth reading and where its explanation resolves. Keep that short planning note outside the reader-facing chapter.

Inspect the corresponding Edition 3 material for a documented incident, revealing exchange, failure, frustration, or memorable explanation. Preserve enough of a useful incident to show what someone wanted, what went wrong, what was tried, and what the outcome taught. Reducing a story to a biographical clause can retain the fact while destroying the lesson.

The story and mechanism should illuminate each other. Explain what a failure exposed about the design, and return to that consequence when explaining the better approach. Do not open with a lively anecdote and abandon its stakes for a long catalog of rules.

A chapter does not require an anecdote. A concrete engineering problem can carry the reader when no documented story fits. An illustrative scenario must be recognizably illustrative; it must never masquerade as a historical run.

Never invent dialogue, emotions, a user, a benchmark, an experiment, or a successful result. Put a missing story or verification need in working notes rather than filling the gap with fiction. An unresolved material gap prevents acceptance; it is not polished away.

Record the story-preservation choice and its source in the author handoff. The reviewer checks it against the source chapter.

## 6. Explain the mechanism and the judgment

Begin with what the reader needs to understand and why the distinction matters. Show the ordinary path in a concrete sequence before discussing its edge cases. Name who acts, what each participant knows, what changes, and why the next step follows.

When criticizing an approach, explain what it actually does and how that produces the failure. A prescription without a visible reason reads as taste. Preserve the conditions under which the criticism holds; a failure in one design does not prove a universal prohibition.

Retain the technical details that change a decision: ownership, state transitions, ordering, evidence boundaries, cancellation semantics, resource consequences, and meaningful alternatives. Which details belong depends on the chapter. Do not turn this list into a mandatory chapter template.

Translate essential implementation requirements into causal prose. Removing exercises must not remove the knowledge that previously appeared only in an exercise contract. Conversely, a grading command or exhaustive submission recipe does not become useful prose merely because its bullets were joined into sentences.

For example, an illustrative account can explain that a human's hint ends the current wait so the model can respond, while the underlying test continues running. The existing job handle lets the model observe that same test later. The reader needs the distinction between ending an observation and cancelling work, and the consequences of confusing them. They do not need an assignment telling them which files to edit to demonstrate it. If event ordering or call-result pairing matters to the explanation, explain why and preserve it rather than hiding it under “implementation details.”

Explain a current design coherently. Do not leave an obsolete instruction followed by successive patches for the reader to reconcile. Keep the history of a mistaken approach when it teaches something the final rule cannot.

## 7. Prose, organization, and exclusions

The reader-facing book consists of stories and technical knowledge in prose, organized with useful headings. Every chapter’s main heading must exactly match the corresponding Edition 3 chapter number and title; the contents page uses those same titles. Do not invent more literary replacement titles. Subheadings should name a problem, question, or mechanism rather than supply decorative suspense.

There are no TL;DRs, exercises, quizzes, assignments, grading recipes, student deliverables, or bot operating instructions. Do not disguise them as “try this,” “next steps,” or end-of-chapter checklists. Do not add repetitive summaries simply because the source edition has them.

Prose should carry the explanation. Inline technical names and small quoted artifacts can clarify a point, but do not reproduce source code listings, tables of contracts, or tool transcripts as a substitute for explaining them. Essential evidence should remain identifiable through concise attribution or a link where appropriate; detailed verification logs belong outside the reading flow.

Use paragraphs to explain causes, dependencies, and consequences. Avoid turning the companion into lists of individually correct rules. Let the reader follow a line of reasoning, not assemble one from fragments.

Preserve useful orientation and transitions. Remove motivational padding, repeated conclusions, and incidental details that interrupt understanding. Compression is successful only when the reader does less work without losing the knowledge.

## 8. Humor and confidence

The comic register is deadpan. State the precise absurdity and let the reader supply the reaction. Natural humor is welcome; forced jokes, stock catchphrases, and a punch line at every paragraph end are tiring.

Be opinionated when the explanation earns it. Name the capable reader's likely objection and answer it directly. Do not manufacture suspense with “here's the kicker” or inflate the stakes into claims that every design choice changes the world.

Overstate a reaction, never a number. Never joke inside a correction, security warning, or cost figure. No profanity. Frustration points at a mechanism, not a named individual or employer.

Avoid generic promotional and academic filler: “delve,” “game-changer,” “seamless,” “it is important to note,” “in conclusion,” and similar scaffolding. Avoid invented concept labels, rhetorical-question reveals, repetitive negative parallelism, emoji, and section-ending morals. Avoid em-dashes outside quoted artifacts or code. These are editorial diagnostics, not a reason to distort a quotation or delete an explanation.

## 9. Evidence and publication guardrails

The original voice guide's factual, privacy, and publication restrictions remain in force. In particular:

- Verify quotations, figures, dates, and claimed outcomes against their sources. Historical chapter prose is a lead to evidence, not proof of every detail.
- Attribute Bill's reported experience as reported experience when that distinction matters. His time estimate for acquiring expertise is not a productivity benchmark or a promise of how quickly readers will learn.
- Distinguish observed behavior, inference, conviction, and intended behavior. Date time-sensitive API and policy claims. Do not convert a historical publication status into a current assertion.
- No invented receipts, users, war stories, comparative performance claims, or completed self-improvement cycles. Statistics need a named source and year; unresolved verification needs belong in working notes.
- Criticism carries evidence. Describe the API surface and observed behavior accurately. Name API surfaces, model families, and products rather than companies by default. The original source chapter's narrowly scoped vendor-policy exception is not a general license.
- Do not name Bill's employer, internal frameworks, coworkers, managers, or executives without his explicit authorization for that request. Do not publish the undervalued-at-work framing or the unlogged story of a vendor model threatening sub-agents.
- The authorized Gemini `compress_context` incident may be told with its actual source and limits. Do not expand it with invented dialogue or results.
- Biography is not a substitute for evidence. Bill's experience explains how he encountered a problem; the technical argument must still stand.
- Preserve the distinction between Bill's ambition for agentic codebooks and what an actual experiment has established. Confidence in a possibility does not establish its completion.
- Do not invent new financial or product promises. If token pricing is relevant, preserve the original no-profit-on-proxied-tokens and billing-cost-pass-through wording.

State consequential uncertainty where it matters without narrating every verification precaution. Omit an incidental claim that requires more qualification than it contributes. Never remove uncertainty solely to sound more confident.

## 10. Continuity and review

Use a fresh author and a fresh independent reviewer for every chapter. Both read this file, the corresponding Edition 3 source, the preceding accepted companion chapter when one exists, and the preceding author's carryover notes. Consult the original voice guide for inherited guardrails and documented personal material; this companion guide governs the human-first format and audience.

Authors leave a short working note with the narrative choice, knowledge retained, sources, unresolved questions, and useful carryover. Keep those notes outside the manuscript. A new author should inherit understanding rather than accumulated conversational fatigue.

Review complete passages, not only isolated sentences. Check whether the chapter:

- Gives an experienced reader a real reason to care and sustains it.
- Preserves the source's useful stories and their causal lessons.
- Transfers project-specific judgment without beginner padding.
- Explains the mechanism before requiring the reader to juggle exceptions.
- Retains distinctions needed to lead, extend, and question the design.
- Reads enjoyably without forced humor, manufactured rhythm, or superficial brevity.
- Makes every sentence earn its place through knowledge, cohesion, readability, or pleasure; avoids philosophical padding and repeated conclusions.
- Contributes to a companion significantly shorter than the original; records a simple source/companion length comparison, not a quota or a separate audit project.
- Keeps the exact Edition 3 chapter title.
- Contains only supportable stories and claims, with inherited guardrails intact.
- Stands as prose rather than a stripped or disguised exercise manual.
- Connects naturally to the preceding chapter without repeating its lesson.

Use applicable targeted prose checks and inspect their actual findings. The original book's student/build/grading workflow is not the companion's editorial acceptance test. Lint success does not establish readability or completeness; a justified style exception does not turn a failed lint command into a pass.

Integrate corrections into the explanation, resolve material findings, and pass the accepted chapter and concise carryover notes to the next fresh pair.
