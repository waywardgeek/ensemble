# Chapter 23: Auto-Recall

Chapter 22 gave old decisions a durable home. That does not make them available when a new question needs them. A successor may have the current code and a compressed account of recent work, yet miss the archived reason the team rejected an apparently sensible change. Auto-recall searches that archive before the main response and supplies relevant passages without requiring the user to remember where they were filed.

The [source chapter](../edition-3/chapter-23.md) describes a local keyword search followed by a finite relevance judgment. BM25 supplies candidates; the judge reads them alongside the conversation and selects what would help. Neither step guarantees that the right memory comes back. Search can miss different wording, and a judge cannot select a passage it never receives. The design makes those limits tolerable while keeping retrieved text separate from instructions, preserving the user's words, and recording exactly what the model saw.

## Find candidates before asking for judgment

BM25 is a useful starting point for a small markdown corpus because it needs no embedding model or retrieval service. It combines term rarity, repetition with diminishing returns, and document-length normalization. A rare project identifier should distinguish a note from its neighbors. Repeating one common word should not overwhelm several distinct matches, and a long document should not win merely because it has more opportunities to mention the query.

The inverse-document-frequency variant matters. The source uses the smoothed form, `ln(1 + (N - df + 0.5) / (df + 0.5))`, where `N` is the number of chunks and `df` counts chunks containing the term. A classic form clamped at zero discards terms appearing in more than half the corpus. In a small archive, every useful query term can meet that condition. Search then returns zero scores without an exception: indexing and arithmetic appear healthy while the Agent remembers nothing. Smoothing keeps those contributions positive, though they can still fall below the score floor.

A positive hit does not establish that BM25 is working. The source's local review found that replacing it with raw term frequency passed the initial tests. Both found the planted words. Competing answers exposed the difference: a rare term against common repetition, several query terms against one repeated term, and a focused note against a long document with the same single match. This is the evidence needed to distinguish ranking from mere word detection. Corpus composition matters too. The source's fixtures use 25 background chunks to make rarity meaningful; that is a fixture choice, not a minimum archive size.

Tokenization splits on non-alphanumeric characters, lowercases, and drops tokens shorter than two characters. Stop words are removed from queries, not documents, so document-length and frequency statistics still describe the text being ranked. Markdown is split at second-level headings; chunks over 4KB are split further at third-level headings and then paragraphs. Breadcrumb headings preserve where a sub-passage came from. Chunking is not just a size control: an isolated decision is easier to rank and understand when its heading survives.

## Search each source separately; reserve final slots

The source reports that a merged search favored terminology-heavy design documents over short daily memories containing decisions. Length normalization could not express the extra value of a personal record. The remedy is a source policy, not a change to BM25.

Memory, documents, handoffs, learnings, and skill documentation each have their own index. This also gives each source its own rarity statistics; a large document collection cannot redefine what is rare in the memory archive. Mechanical selection reserves half the final slots, rounded up, for qualifying memory hits. Other sources share the rest evenly, and unused capacity flows to sources with overflow.

The reservation belongs to the final result, not the candidate pool. Limiting how many documents enter the pool does not stop them taking every slot in a later global score sort. With three mechanical slots, two qualifying lower-scoring memory hits must survive competition from several higher-scoring document hits. Otherwise the implementation has a candidate limit but no effective memory reservation.

These quotas govern selection without a judge and fallback after judge failure. They must not repair or rebalance a successful judgment. A judge that selects only a document has assessed relevance rather than source proportions. A successful `[]` means nothing was worth attaching. Treating it as failure would inject the very passages the judge rejected.

## The corpus must agree with recovery and context

The archive is not just the files most recently written. Recall searches active memory files and `.md.retired` originals, allowing it to recover detail that compression removed from the working bands. It also searches configured documents, handoffs, learnings, and available `SKILL.md` files. Corpus paths belong to the Agent's data directory; the supplied CLI and GUI use `cr/` beneath their launch directory. Skills come from the Agent's separately configured skills directory. These are owned paths, not searches relative to an arbitrary working directory. File age alone is not a reason to exclude a memory. The exclusion is content already present in the actual context, whether in a memory band, a tool result, or earlier recall.

The embedding client must load history and finish relevant memory recovery before the first corpus scan. Constructing the recall service must not eagerly read disk ahead of that boundary. Hidden compaction staging records are not accepted memories. If a committed replacement still needs disk finalization, recovery must establish its committed bytes first; an unresolved conflict is reported rather than allowing search to consume a mixed corpus. The index is built lazily and refreshed when source bytes change. Equal-length edits still count as changes.

Those rules connect recall to Chapter 22's transaction boundary. A passage is not safe retrieval input merely because it exists on disk, and it is not absent from context merely because it came from a different file. Recovery establishes which bytes are authoritative; actual context establishes which bytes would be redundant.

## Let one component own the judge's contract

The judge receives the recent conversation window, the latest user message labeled separately, and up to 20 numbered candidates. It selects up to three useful passages as a JSON array of indices. Keyword overlap alone is not enough. The current source defaults to the Agent's provider and model; a judge-model override can select another supported model on that provider without changing the main conversation's model. A smaller model may reduce cost, but its selection quality is a separate question.

This is a stateless call, not another full Agent. Each judgment starts with only the supplied instructions and data. It has no recall lifecycle to trigger recursively. Conversation excerpts and candidate snippets are untrusted data inside the prompt, including any instructions quoted within them. The same boundary applies after selection: recalled text is evidence to consider, not a new source of authority.

The recall component both defines and parses the reply format. The model layer supplies the client but must not add a second description of the expected response. Two descriptions can drift until a model obeys one and the parser rejects it under the other. Mechanical fallback then masks a persistently broken judge unless failures are visible.

The parser tolerates ordinary formatting variation: prose or fences around the array, quoted numeric indices, and mixed elements. Invalid indices are ignored. A less obvious hazard comes from partial decoding: a failed integer-array decode can already have written values into its destination. That slice must be cleared before trying mixed elements, or leftover zeros can accidentally select candidate zero. Tolerance should recover intended selections, not manufacture them.

A failed or timed-out judgment falls back to final quota-selected BM25 results. A successful empty array does not. The 15-second timeout must cancel the actual HTTP request; stopping a wait while leaving the request running is not cancellation. Interrupting the turn or shutting down its owner likewise ends owned work rather than producing a late attachment.

Judge usage still belongs in session cost. Engine accounts for measured input, cache, and output once, under the judge's actual provider, model, and price tier. It leaves the latest main-response counts unchanged and excludes the separate judge prompt from main-request cache comparisons. Replaying a saved attachment makes no judge request and adds no historical usage to the fresh session meter.

## Delivery owns what actually reaches the model

The source keeps the recall service independent of the Agent actor, using common contracts rather than importing sibling subsystems. The Agent owns the worker context, turn identity, conversation mutation, log events, and provider accounting. Retrieval can return candidates and judgment; it cannot assume its result was delivered.

Both synchronous `Ask` and mailbox `Post` enter recall through the same actor begin path, after recovery and before the first main request. The historical implementation had separate paths, and wiring only one made recall work for one client while silently doing nothing for another. The worker reads a snapshot while the actor remains able to accept hints or interruption. Its result carries the originating turn identity, so a late response cannot become context for the next user's message.

The actor's final delivery decision excludes actual-context duplicates, preserves mechanical quotas where they apply, and formats only useful selected text. `MaxRecallBytes` caps the complete formatted attachment, including headings and attribution, rather than just the source bodies. It must not leave an empty recall heading after all content disappears. A skipped or empty result adds no recall entry and does not create a separate context revision.

Delivered recall has its own conversation entry kind. It is never fused into the stored user message: later removal must not require editing the human's words. Provider rendering remains vendor-specific. Separate stored entries can become separate parts of one user message on the wire; demanding a system message would confuse storage identity with transport layout and risk promoting retrieved data into instructions.

Recall retains its recorded position and bytes on subsequent requests until the checkpoint removes entries of that kind. Removing it as soon as relevance shifts sounds economical, but prompt caching depends on a stable prefix. Deletion changes the request from that point onward. Grouping removal with a checkpoint that already changes the prefix avoids another independent invalidation. The bytes are still sent and cache reads may still cost money; stable placement creates an opportunity for reuse, not a guarantee of free context.

## Record the choice, including the identity of capped content

Selection is nondeterministic because a model participates. Rendering cannot search today's files or ask today's judge to reconstruct yesterday's attachment. The log records the selected text when attachment happens, so replay restores the same bytes even when the source files are absent. This is the same boundary used for memory replacement: record nondeterministic outcomes, then render the recorded conversation.

Deduplication needs more than searching context for the complete source passage. The source's local reproduction used an 11,247-byte snippet capped to a 6,144-byte attachment. Its omitted suffix could never match the context, so the next turn attached the same shortened passage again. A short-snippet control passed and concealed the large-snippet defect.

The accepted design associates a hash of the complete source content with the attachment representing it. It hashes the source, not the shortened display. That identity survives replay and suppresses a repeated capped attachment; a human edit creates a new identity. The identity is recorded only for content actually delivered. Marking every selected candidate as seen would suppress passages that lost out to the byte cap without ever reaching the model.

This state belongs to the attachment, not a permanent reading history. Checkpoint removal discards it with the recall entry, allowing future retrieval when the passage is no longer in context. Exact matching and source hashes still cannot detect two differently worded accounts of the same decision. They prevent literal repetition, not semantic duplication.

## Tune for signal without confusing defaults with guarantees

Recall settings extend the strict `ApplicationConfig` schema introduced in Chapter 21. In file mode, root recall patches persist in that same application file; child preferences remain local. Successful patches apply to the live Agent before acknowledgment while preserving pending next-launch provider edits and unrelated memory fields. Persistence alone is not enough: a saved preference that the running Agent ignores has not taken effect.

The [first-edition tuning account retained in the source](../edition-3/chapter-23.md) reports raising the plain-search score floor from 0.5 to 3.0 because common technical terms produced too much noise. BM25 scores are not normalized across corpora, so that history explains a default rather than establishing a universal threshold. With judging enabled, search uses the lower 0.5 floor to offer a broader candidate set and leaves relevance filtering to the judge.

The default 80-character query gate avoids spending retrieval work on messages such as “fix it.” It is a cheap heuristic, not a measure of meaning: a short project identifier may carry more signal than a long generic request. The current candidate default is 20, distinct from the older 80-candidate code setting described in the tuning history. Candidate bytes, not merely candidate count, determine prompt size. The 15-second timeout likewise reflects the historical move from an assumed sub-second local judge to cloud calls reported at one to three seconds.

Below the query gate, below the score floor, or with recall disabled, the Agent continues without passive memory. Judge failure leaves noisier mechanical recall. These are reductions in help, not reasons to make ordinary chat unusable; owner interruption and recovery errors still retain their own semantics.

BM25 does not connect “authentication” with “login flow” unless the indexed words overlap. A judge cannot repair that missed candidate. The source also implements no relevance feedback: a passage ignored repeatedly can return after checkpoints if it keeps scoring well. Embedding retrieval could address vocabulary mismatch but introduces another model and index to maintain. Whether that cost is justified depends on the corpus, not on the presence of a more sophisticated technique.

Skill documentation adds a useful consequence without changing those boundaries. A relevant `SKILL.md` passage can tell the Agent a capability exists and lead to an ordinary `load_skill` call. Recall itself neither activates the skill nor grants its tools. The registry remains responsible for what can be loaded and executed.

## Evidence that a reason came back

The [source's retained CLI demonstration](../edition-3/chapter-23.md) uses a synthetic archived migration decision, a cafeteria-contest distractor sharing its keywords, and a background corpus. The first question asks for facts absent from the prompt: a release code, owner, rehearsal date, rejected retry approach, unresolved ledger location, and retention period. The important recovered reason is that a timed-out remote transaction may already have committed, making automatic retry unsafe. A short follow-up falls below the recall gate, separating continued use of an attachment from a second successful search.

The source reports earlier live runs with Haiku 5.5, GPT-4o mini, and Gemini 3.8 Flash. All answered both turns correctly without tools, with one judge request followed by two main requests. The judge saw the decision and distractor; only the decision reached the main requests. OpenAI's fenced array exercised tolerant parsing in actual use. These were provider calls over a planted corpus, not evidence of broad recall quality across a private archive. They also preceded the capped-source and path corrections, so they do not validate those later changes or recommend those historical models for current use.

Separately, the source reports an initial grader setup failure caused by an untyped primary skill and a structural check aimed at the wrong module. That failure did not test recall behavior. The local ranking comparisons and capped-snippet reproduction answer different questions from the live demonstration: whether the mechanism distinguishes plausible candidates, whether delivery avoids repetition, and whether a receiving model actually uses an archived reason. None substitutes for the others.
