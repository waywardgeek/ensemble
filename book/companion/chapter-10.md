# Chapter 10 — Persistence and Replay

A restart should preserve the conversation without resurrecting the old runtime. Those are different operations. Human messages, completed answers, exact tool arguments, signatures, usage, redactions, and display identities belong to memory. Jobs, sockets, pause state, credentials, and live ownership relationships belong to the process that ended.

Persistence joins a reduced conversation snapshot with an optional diagnostic event journal. On loading, the snapshot establishes the remembered state and only a newer journal tail extends it. Today’s configuration and manuals then establish what the new Agent can do.

## Snapshot and journal answer different questions

The save document carries documentary configuration, reduced Context, an optional event array, and an `as_of` watermark. The watermark names the latest History sequence represented by Context. It is not a wall-clock timestamp or a claim that the diagnostic array contains every earlier event.

This supports several useful forms of the same memory. A full snapshot with events can explain both the effective conversation and how it developed. A snapshot alone preserves the conversation without a large diagnostic log. An event-only document can reconstruct through the reducer. An older snapshot plus a newer event tail can advance the remembered state without reducing the older prefix twice.

Loading validates the whole candidate before replacing anything. Context must not contain entries beyond its watermark; event sequences must be positive and strictly increasing; supported part shapes, usage values, and relationships must decode. Explicit null is not silently treated as an omitted member.

Events at or below the snapshot watermark remain diagnostics. Only later ones enter reduction. The next event number follows the greatest valid sequence or watermark, even if retention has discarded the entire older journal. Sequence continuity belongs to the saved state, not to the current length of an array.

## Runtime authority is supplied anew

Saved configuration explains the old session. It does not select the new vendor, endpoint, model, credentials, tool handlers, or primary manual. The host chooses these at startup. Saving credentials would be both unsafe and the wrong ownership model; core persistence never reads environment variables to recover them.

Runtime skill entries in the old document are structurally validated but do not restore authority. The loader removes old runtime instruction entries, restores dialogue and other effective conversation state, then appends the freshly rendered current primary closure at continuous sequence numbers. Ordinary exchanges about previous skill loads remain part of the conversation, while old loadable grants do not silently become executable again.

This differs from offline replay with no runtime Skills manager. There, recorded rendered skill parts remain meaningful evidence of what the original Agent was taught. Runtime resume deliberately applies current policy. The same saved manual can be documentary evidence without being permission for a newly launched process.

Redactions and summaries survive as reduced state. A saved tool result that was replaced by a redaction does not grow its original bytes back during replay. Opaque replay content and exact numeric arguments likewise retain their representations rather than passing through a friendlier but lossy serializer.

## Load nothing until a whole candidate is usable

The primary save and its backup use the same decoder as real loading. A file that parses as JSON but violates the save format is corrupt for recovery purposes.

Startup tries the primary first, then the backup if the primary is missing or invalid. If neither file exists, it begins a new session. If either exists but neither is usable, startup fails with the existing files untouched. This is deliberately stricter than Chapter 8’s preference fallback: discarding unreadable font settings keeps an interface usable; silently discarding a conversation could convince the human that a different session had resumed.

Validation and reduction finish before live History is replaced. A bad late event cannot leave half a restored conversation behind. Successful recovery from the backup also does not rewrite the corrupt primary merely as a side effect of reading it.

Saving is careful about the backup for the same reason. The complete new save is written to a unique same-directory temporary file. An existing primary is decoded before it is copied into the backup path. If it is invalid, it must not overwrite a known-good backup. Only after preparation succeeds does replacement expose the new primary. Temporary files are cleaned up on failure, and failure remains visible.

Same-directory replacement prevents readers from seeing a partially written primary. It does not make several files one atomic transaction or promise recovery from every power-loss scenario.

## Retaining less does not mean remembering less

Diagnostic retention controls the event array, not Context. Keeping all events, the newest few, or none leaves conversation, usage, redaction state, and watermark intact. Captured in-memory diagnostics are trimmed only after the save succeeds. A failed replacement must not destroy the very evidence needed to investigate it.

A newly opened diagnostic log begins with the retained or bootstrap prefix. It cannot fabricate events already discarded. Once retention has removed earlier events, that truncated log alone is insufficient to recover the full conversation; the save document is authoritative.

An optional archive is a dialogue-only JSONL projection made from Context before trimming. It excludes runtime skill instructions and reflects current summaries and redactions. It is neither a second recovery authority nor a promise to preserve removed raw material forever. A later primary-save failure can leave this derivative archive newer than the save, because archive, primary, and backup do not share one multi-file transaction.

The distinction is practical: the snapshot says what the Agent remembers; diagnostics explain retained activity; the archive offers a convenient conversation projection. Choosing a smaller journal need not amputate memory, and a readable archive should not become an improvised restore format.

## Shutdown is part of saving

Library persistence is opt-in, avoiding accidental sharing of a working-directory save among embedded Agents. The CLI and GUI opt in by default and use `save.json` unless configured otherwise. Independent sessions need independent directories or explicit paths.

Configuration occurs once, before actor admission. Replacing a live Agent’s conversation concurrently is not authorized merely because a public load operation exists. Shutdown joins owned work and destroys Jobs while History remains attached, records completed shutdown, saves, and only then proceeds through logical destruction. The real ownership paths remain available for that final save. A save error cannot be renamed successful persistence.

The CLI drains accepted work at EOF; the GUI saves after its owned host lifetime ends. Neither can execute cleanup after a force kill. Recovery here covers what was saved, not work that existed only in a terminated process.

The GUI reconstructs finalized cards from Context and uses retained event facts for lifecycle and direct-tool diagnostics. A snapshot-only save therefore still produces a transcript. It preserves response and part identities, does not append synthetic history, and does not restart old jobs. A saved shutdown describes the previous runtime, not an instruction to disable the new one. Memory survives precisely because its representation does not pretend the old process survived with it.

Source: [Chapter 10](../chapter-10.md).
