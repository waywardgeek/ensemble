# Chapter 10: Persistence

Chapter 9 gives an Agent new instructions and tools without replacing its identity. None of that helps when the process exits. A useful afternoon of work can still end with the next launch meeting a stranger.

The first persistence design made remembering optional: it had a `--load` flag. Bill’s review, quoted in the [source chapter](../edition-3/chapter-10.md#106-default-load-and-refusing-a-bad-file), was one line: “Let's load by default without a flag.”

That changes more than the command line. If loading is an explicit experiment, a bad file is an experiment that failed. If loading is the normal start of work, a bad file is the user’s conversation in trouble. Starting fresh and saving cheerfully on exit can destroy the very thing the user returned to recover.

The design in this chapter makes saving and loading ordinary lifecycle operations. It also puts the earlier event architecture under a useful strain. The Agent has a snapshot of what the conversation means and a history of how it got there. Across a restart, those two accounts must agree.

## What Actually Survives

The save is one JSON object. Its recovery vocabulary is small: `context` holds the snapshot, `as_of` identifies its event boundary, and `log` holds the available canonical events. `config` records the request configuration at the time of saving. Optional `log_mode` and `dialogue_log` distinguish a full archive from a derived dialogue archive; older four-field saves mean full mode.

The context is what the renderer reads: dialogue, its structured parts, and the other state already defined by the event model. The log is the sequence of facts from which the reducer derives that state. A snapshot is not a second interpretation of the conversation invented for persistence. It is the renderer’s existing context, saved.

That distinction limits what “the Agent remembers” means. A saved tool declaration does not become a handler closure on restart. A recorded job does not resume as a running process. The save does not reconstruct the skills registry’s loaded or pending-unload state. The new Agent supplies its current tools and primary skill, while the restored conversation can still contain previous tool results and manuals. Persistence carries conversation state, not the whole machine that once produced it.

The recorded configuration contains vendor, model, effective system prompt, and tool declarations. API keys and base URLs are excluded from that metadata. More importantly, the metadata does not configure the new Agent. Today’s model, prompt, tools, and settings win. Pinning restoration to yesterday’s model would make a model retirement a conversation-retirement service.

The source records a quieter mistake in the earlier Edition 2 implementation. Its first save code copied the configured system prompt. For an ordinary Agent without a skill, that field was empty, but Engine supplied a nonempty default to the actual request. The conversation resumed correctly; its metadata described instructions the model had never received. A test using a primary skill missed the discrepancy because that path had already filled in the configured text.

The correction asks Engine for the effective instruction through its existing owner interface, sharing the decision with all three renderers. Even that has an exception worth keeping: an intentionally empty primary skill must stay empty. Replacing every empty string with a default would repair the no-skill case by breaking the empty-skill case. The source’s focused checks compared saved instructions with the captured HTTP field and checked the empty primary separately. The useful lesson is not to normalize more aggressively. It is to ask the component that owns the decision.

## The Anchor Makes Two Copies Useful

Suppose a snapshot contains everything through event 7, while the available log runs through event 10. The loader installs the snapshot, then applies 8, 9, and 10. It neither replays the first seven events nor ignores the final three.

`as_of` makes that possible. It is the sequence number of the last event folded into the snapshot, not the number of retained log entries. Without it, a loader has no principled way to choose between applying everything twice and omitting everything newer than the snapshot.

With a non-null context, restoration applies only canonical events whose sequence is strictly greater than the anchor. With a null context, restoration begins from an empty context and applies a complete canonical history. The reducer is the same in both paths. Snapshot-only recovery is also legitimate: a non-null context and an empty log contain everything the renderer needs to continue.

A normal full-history save usually has an anchor equal to its last logged sequence, so it has no recovery tail. This makes a broken loader surprisingly easy to admire. A loader that always ignores the tail can pass round trips using every file its own save routine produces. The revealing case pairs an older snapshot with a newer log. The source deliberately tests that combination. It exercises the shape needed for later crash recovery without claiming that this chapter already writes events durably during execution. Here, saving happens at normal exit; Chapter 14 adds journaling.

The snapshot and anchor must also be captured together. Copy the context just before an event arrives and the anchor just afterward, and the save claims to contain a fact it has never seen. History already owns both values, so its existing lock can protect the pair. A new persistence coordinator would add another owner to a problem that already has one.

Restoration has a similar ownership boundary. Decode into a fresh candidate, validate the configuration, snapshot state, anchor, events, and strictly increasing sequence numbers, and stage the resulting context and log before installing them. Even an event below the snapshot boundary must be valid, although it will not be applied again. A malformed late event must not leave a live Agent with the first half of a restored conversation.

Validation cannot simply demand the tidiest possible transcript. The public `Tools().Dispatch` path can produce tool events outside a model turn without a native call ID. That absence is legitimate history and must remain absent. Tool calls inside a model turn still need their supplied correlation IDs. Inventing IDs to make an archive look uniform would alter the facts it is supposed to preserve.

Loading is local reconstruction. It does not ask a provider to explain the old conversation, nor should merely opening a restored GUI view make a provider request.

## Two Roads to the Same Request

Persistence makes the Chapter 2 promise falsifiable: context is a pure function of events. No clock, unstable map traversal, or engine mutation behind the reducer’s back may supply a fact needed to reconstruct it.

A process can violate that promise for its entire lifetime without noticing. Its live context is the only copy anyone consults. A save creates a second account. For a complete canonical history, loading the snapshot and rebuilding from the log should produce byte-identical vendor requests for the same next prompt under the same runtime configuration.

The request is a useful comparison boundary because it is what the model receives. It avoids requiring an outside observer to know the Agent’s internal representation. It also catches failures that a plausible answer conceals. A model can remember a word despite duplicated turns, lost tool results, or missing opaque provider material. Correct recall is welcome; it is not proof of faithful reconstruction.

The source’s historical Claude, OpenAI, and Gemini sessions each remembered `cobalt-juniper` in a second process without being given the word again. Claude also remembered it from independent snapshot-only and log-only variants. Those are reported restoration observations, not current provider coverage or proof of request equality. The source’s deterministic checks did different work: they exercised all three renderers with tool identities, opaque material, a redaction, and pending ephemera, comparing snapshot, full replay, older-snapshot-plus-tail, and snapshot-only paths.

There are two kinds of equality here. Rendered requests must match byte for byte under the fixed comparison configuration. Comparing a reconstructed context with its saved snapshot should compare decoded values, not JSON indentation or whitespace inside encoded tool arguments. Nil and empty dialogue or ephemera collections can be normalized for that comparison. Numbers must retain their exact values; rounding a large integer could hide a changed tool argument. Representation tolerance does not excuse missing fields, different facts, bad anchors, or disordered sequences.

The complete-history condition is not small print. Once a log prefix has been discarded, a later save cannot manufacture it. A snapshot-only save can resume correctly, acquire new events, and save again, but its resulting log is still not a complete account from the beginning. Nulling that snapshot is not a replay test. It is deleting the only surviving representation of earlier state. The loader must refuse a rebuild whose originals are unavailable.

The source also preserves a useful false alarm from the fresh Edition 3 work. Default-load checks failed while snapshot, tail, and refusal checks passed. The fixture supplied `CURRENT_RUNTIME_PROMPT` with a terminal newline; its exact expectation had none. Correcting the fixture input resolved the mismatch with production code unchanged. Trimming prompts in production would have made the test happy by changing a different behavior. Restoring conversation and selecting current runtime instructions remain separate operations, and a failing comparison is evidence to investigate, not permission to blur that boundary.

## Identity Cannot Restart at One

An archive can lose log entries without losing sequence identity. If its snapshot is anchored at 7 and its log is empty, the next event is 8. If it has a tail ending at 10, the next is 11. Allocation continues one past the larger of the anchor and the last retained event sequence. The length of the array is irrelevant.

The browser has a related problem with different ownership. A provider tool-call ID correlates a call and result. A transcript card address identifies something the user can select. A parser’s turn-local counter is not automatically a safe address for a card that survives process restart.

The source describes the fresh implementation having to preserve old card addresses while allocating new ones beyond the restored conversation. Otherwise, a new stream can append to an old card or an old control can select a new result. Remembering the text while moving its buttons to somebody else’s result is only partial success.

Preventing future collisions is not enough when a legacy archive already contains colliding parser-local IDs. Those old artifacts still need distinct presentations, deterministically across replay and restart. The display must repair its addressing without rewriting provider payloads or call/result correlations. The archive records what happened; the GUI decides how to address its presentation. Neither needs a new public identity framework merely to respect the other’s ownership.

This is also a boundary on evidence. Backend replay equality cannot establish browser selection behavior. The source reports separate DOM assertions against a local fake provider across a GUI restart. Those assertions checked old and new card identities; screenshots were not a human visual or audio review, and the fake provider was not a live-provider test.

## Retain the Conversation, Not Every Diagnostic Forever

The model never reads the canonical log directly. It reads the context. Already-folded events can therefore be removed from a save without changing the next request, provided the snapshot remains. What disappears is the ability to reconstruct or diagnose discarded history, not necessarily the ability to continue the conversation.

The source offers two saved-history modes. `full`, the default, retains all available canonical history. “Available” matters after a trimmed save: changing the setting cannot recover a missing prefix. `dialogue_only` keeps an ordered, derived archive of visible user and assistant text, with source sequence numbers for provenance. It excludes tool calls and results, opaque reasoning, private replay payloads, and API or GUI traces. For an assistant event containing several kinds of material, only its visible dialogue text enters that archive.

A dialogue record is not a filtered Event wearing its old name. Its source sequence points to where the text came from; it is not a newly allocated event sequence. The reducer never consumes `dialogue_log`. Preserving prior available dialogue records on subsequent saves keeps the archive useful, but switching back to full mode cannot restore the originals from that projection.

Dialogue-only mode still requires a non-null snapshot. It does not compact that snapshot, which may retain tool or provider material needed for current state. Nor may the policy filter the canonical recovery tail above `as_of`. That tail stays complete and separate in `log` until a newer snapshot folds it. A dialogue-only archive is consequently neither complete replay history nor a promise that all non-dialogue material has disappeared from the save.

Diagnostics have another purpose again. API and GUI traces explain operation; they are not the conversation’s recovery state. Bill reports that his Gemini CLI fork ran out of disk space, though the source does not establish that fork’s storage internals or whether it retained context snapshots. The report supplies a reason to care about disk growth, not a diagnosis of another system’s persistence design.

The diagnostic policy retains the newest configured number of complete records in order. The `diagnostic_max_records` setting defaults to 200 per managed diagnostic log at each retention boundary. Zero means unlimited; negative limits are invalid. Counting lines would split multiline records, and counting bytes would be a different policy. These boundaries limit retained records after cleanup, not disk usage continuously between cleanups.

History policy belongs to the Agent; diagnostic cleanup belongs to the writer. Both settings use the existing settings ownership and persistence from Chapter 8, not provider credentials, browser display preferences, or extra fields in the recorded request configuration. An explicit zero must survive settings updates and restart. Unknown history modes are errors. The mode recorded in a save describes that file; current runtime settings govern the next save.

Cleanup comes after recovery is secure. At startup, load or recover successfully before trimming, establishing a valid recovery save first if none exists. At normal exit, finish accepted work and save successfully before trimming. A failed load or save leaves diagnostics intact, precisely when those diagnostics may be useful. Constructing a dialogue-only replacement likewise must not destroy prior recovery state or trim live history if the save fails.

The writer must coordinate retention with appends. Replacing a diagnostic file requires reopening it or resetting the writer’s offset before writes resume. External truncation can leave a non-append descriptor positioned beyond the new end, creating a sparse gap. Renaming or unlinking a pathname does not close an open inode or guarantee that its space is reclaimed. Cleanup failure should be reported without destroying the prior diagnostic file or invalidating the recovery save.

This policy does not authorize a general disk sweep. Saves, recovery journals, memories, and worker state are not diagnostics. Shared rebuildable caches need separate authorization. Chapter 14 will add retention after a successful `micro_handoff` save; no compactor or handoff tool is introduced here.

## Default Recovery Must Refuse Destructive Help

For the single-Agent commands, the ordinary location is `save.json` in the working directory. `--save PATH` changes the location for both loading and saving, not whether loading occurs. The library’s default belongs under the owning Agent’s data directory so two Agents do not accidentally share a process-wide conversation.

Path resolution carries the same ownership distinction. A relative `DataDir` follows its Agent’s `BaseDir`, rather than whichever directory happened to launch the process. An explicit `SavePath` remains the caller’s choice and must not be silently rebased. Preferences keep their separate persistence; conversation saving does not take it over.

Startup first validates the selected primary completely. If it is missing or invalid, startup independently validates its `.bak` sibling. Only two absent candidates mean a fresh conversation. If neither existing candidate is valid, startup refuses, names the failures, installs no restored state, makes no provider request, and leaves both files untouched.

The refusal is deliberately less friendly than a silent reset. A fresh session would eventually overwrite the damaged conversation with a perfectly valid account of having forgotten it.

A backup recovery must report the boundary it actually restored. Its snapshot, anchor, and log travel together. A loader cannot attach an arbitrary tail from the damaged primary and call the result recovered history. Here the log lives inside the save, so newer bytes lost with that file are lost to this mechanism. Chapter 14’s separate journal will change what evidence is available after the backup anchor; it is not already present because the loader understands tails.

The timing of save registration matters too. Save-on-exit belongs after successful loading. An unconditional deferred save installed before `Load` can return exactly the right error and then overwrite the bad file on the way out. Checking the error alone would miss the destructive half of the behavior.

Successful saving first encodes and validates the entire new candidate without changing either file. If the current primary validates, its bytes become the backup through a temporary file and rename before the new primary is installed the same way. If recovery used a good backup because the primary was absent or corrupt, that backup stays intact while the new primary is installed. Corruption must never be promoted over the recovery copy.

The existing save owner serializes this operation for one Agent. Encoding, closing, or replacement failures propagate to the caller and leave a valid recovery candidate intact. They do not trigger diagnostic cleanup. Atomic replacement prevents a partially written JSON file from becoming the primary; it does not promise universal power-loss durability. Normal-exit persistence does not need a database or a write-ahead journal smuggled in under the word “atomic.”

## EOF Is the Beginning of Shutdown

Closing stdin ends admission of input, not necessarily the work already admitted. A final turn can still be running, and a quiet observer stream does not prove that it has finished. The application waits for reliable completion, shuts down the Agent, joins managed jobs that can append final facts, and only then saves.

Both CLI and GUI use that lifecycle behind the same public Agent. In the GUI command, closing stdin ends the application after admitted work is flushed; it is no longer merely detaching a terminal view while the server carries on. Browser prompts and terminal input still reach the same Agent.

The source’s CLI emits final usage before its deferred shutdown save. That usage record therefore cannot acknowledge persistence. Successful process exit is the observable boundary before reopening the file. The same distinction runs through the chapter: finishing an answer, saving a snapshot, installing a replacement, and recovering it later are different events.

The source reports a live tool-bearing turn carried through EOF and relaunch, followed by an answer from the restored conversation. An initial paid attempt failed at the transport boundary; bounded TLS recovery completed the two launches without establishing the original failure’s cause. That observation supports the tested model’s restart path, not every provider or every failure mode.
