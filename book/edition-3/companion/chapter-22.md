# Chapter 22: Auto-Recall

Bill's earlier memory system had a search tool and an archive full of useful decisions. The agent rarely searched. It could produce excellent answers when it did, then begin the next session by proposing an architecture the team had already rejected.

Availability had not produced a habit. Automatic recall moves the initial search to new-turn admission, where a relevant old fact can arrive before the main model answers. The model still has to use the fact well. It no longer has to remember to look for it first.

The historical improvement in daily use motivates the design. The C# pipeline adds explicit ownership, cancellation, replay, and input-boundary rules so that a helpful retrieval feature does not become an unaccountable source of instructions.

## Search several small worlds

Five independent corpora cover memories, documents, handoffs, learnings, and visible skill manuals. Memory includes live and retired originals as well as SOUL and MEMORY. Skill discovery reads exact `SKILL.md` files, not every script and fixture under a skills directory.

Separate indexes keep a large documentation collection from deciding what counts as a rare word everywhere else. They do not turn scores into calibrated probabilities. They keep the statistics meaningful within each source.

Each Agent owns a Recaller whose indexes are derived immutable values. Searches read current bytes, fingerprint names and content, and replace indexes only after a complete successful refresh. Equal file sizes do not establish unchanged text. Readable invalid UTF-8 or oversized input refuses the refresh; quietly accepting a partial corpus would conceal a broken input boundary.

Agent-home traversal uses its Sandbox. An explicitly selected skill tree has its own provenance and a read-only rooted handle, without creating broader file-tool permission. Memory's recovery gate prevents recall from indexing a half-promoted transaction. A staged candidate must not become evidence merely because it resembles a Markdown file.

These rules permit stable memory files in the corpus without blindly attaching them twice. Duplication is determined against the actual retained and projected context, not by categorically excluding useful files from search.

## BM25 earns its place

For a small Markdown archive, local keyword ranking avoids an embedding service, another database, and another source of credentials and latency.

BM25 combines term frequency with saturation, rarity within the corpus, and document-length normalization. Repeating a word ten times should not make it ten times more useful. A term found everywhere should contribute less than a distinctive project name. A long document should not win merely because it has more opportunities to match.

One apparently minor choice caused a particularly quiet failure. Classic inverse document frequency can become negative when a term occurs in more than half the documents. Clamp that value to zero in a small archive, and every useful query term may contribute nothing. The index builds and search completes, but the Agent never remembers anything.

The chosen smoothed form, `ln(1 + (N - df + 0.5) / (df + 0.5))`, preserves a positive contribution for a matching term. Thresholds still matter; a positive score is not automatically a selected result. Tiny fixtures can also make useful terms commonplace, so the test corpus must represent the competition it claims to exercise.

A test that merely finds a planted word is weak. Raw term frequency can pass it without implementing BM25. Better cases pit a rare term against common repetition, several distinct query terms against one repeated word, and a focused note against a long document with one incidental match.

Markdown is chunked at headings, with smaller boundaries for oversized sections. Breadcrumb headings retain the location of a paragraph. The source text and provenance remain available; the index does not become another authoritative copy of the archive.

## Personal memory needs a selection policy

A decision recorded in a short daily log may matter more than a long design document rich in matching terminology. Independent indexes alone cannot express that preference.

Mechanical selection reserves half the final slots, rounded up, for qualifying memory hits. Other sources divide the remainder, and unused slots flow to sources with more results. The reservation applies to final selection. Capping a candidate pool and then globally sorting it can still let another source take every final slot.

The full ranked pool remains available for fallback. The smaller candidate limit exists to bound a judge request, not to throw away memories before mechanical quotas can operate.

A semantic judge may override this crude policy. When it successfully selects nothing, the result is nothing. Treating an empty verdict as failure would undo the judgment the application paid for.

## The judge is a call, not another agent

The judge makes one stateless request through the existing Engine. It receives the current query, a small amount of ordinary recent dialogue, and numbered candidate snippets as framed data. The current query alone supplied the search terms; recent dialogue helps interpret the candidates.

There is no child Agent, second transcript, recursive recall, skill catalog, or tool authority. The default judge uses the current model and funding route; an optional same-provider model choice does not establish account entitlement or mutate the main Agent's configuration.

Recall owns both the response-format instruction and its parser. Engine owns transport and provider-native decoding. Splitting the answer contract between those layers invites a model to obey one format while the other rejects it.

Only completed, tool-free output can supply a verdict. Valid array syntax with no usable indices selects nothing; malformed syntax is failure. Duplicate or invalid indices do not silently become the first candidate. A plausible index found in a truncated stream is not completed output.

The judge has an operation-level timeout that reaches the HTTP request. Returning from a wait while a stalled stream runs forever would leave the expensive part alive. Timeout, transport failure, malformed output, or incomplete response falls back to mechanical selection. Cancellation of the user turn instead propagates; it is not permission to keep working and publish a later fallback.

Measured usage enters cumulative accounting once without changing the main latest-response, pressure, or cache observations. Replaying an attachment does not purchase another judgment.

## Keep the human's words intact

An ordinary Human turn is recorded exactly as admitted. Recall then runs as foreground work under the existing actor. Busy hints keep their current immediate behavior rather than starting a second recall alongside it.

The worker searches and judges, then returns full snippets. The actor joins it, verifies the exact turn identity, and checks duplicates again against current context before publication. A hint received while the judge waits may already contain the proposed fact. An interrupted turn must not donate its recall to the next one.

Accepted recall is a separate `RecallAdded` fact with a `RecallPart`. It becomes Human/data content without rewriting the original input. Quoted text, filenames, headings, and attribution make the origin visible. A recalled document saying “ignore the user” is still a document; a recalled skill description does not load a skill.

The recorded attachment stays available for short follow-ups. Rendering does not search today's files and rewrite yesterday's conversation. The selected bytes and identities survive ordinary save and replay.

## Bound display without losing identity

The complete formatted attachment has a UTF-8 byte cap, including headings and quote prefixes, and ends on a valid rune boundary. That bounds the request without pretending a truncated display is a different source.

Deduplication therefore retains hashes of the full selected snippets alongside any shortened presentation. Otherwise a long note can be attached again on the next turn: the retained prefix does not contain the omitted suffix, so full-text matching alone fails.

The check considers retained recall, readable dialogue, nested tool results, live memory, and the instructions actually projected for the provider. Opaque replay is not treated as readable text. A hash is recorded only when some source content fits in the attachment.

Those identities belong to retained context, not an eternal “seen” table. Once a committed cut removes the entry, the fact may legitimately be needed again. A handoff alone no longer releases it while Memory is still holding the source for an uncommitted replacement. Clearing the identity early would permit a duplicate before the safe cut occurred.

## Retrieval remains a fallible assistant

Keyword matching misses paraphrases, chronology can matter more than word overlap, and a judge can choose badly. Historical tuning is useful experience, not a guarantee of current provider relevance.

The implementation can nevertheless make the failure legible. A failed judge can fall back; a corrupt corpus cannot be silently called “no matches.” Settings are saved before publication, and a turn uses its admitted policy rather than half of a later GUI edit. Skill descriptions can help discovery without granting their capabilities.

The practical gain is a more informed starting point. An old reason arrives before the model confidently reinvents the decision it once helped reject. The next chapter examines how to affordably carry that growing, structured past through provider requests.

[Implementation chapter](../chapter-22.md)
