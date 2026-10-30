# Chapter 14 — The world's best context engineering, before breakfast

Bill had the design in his head for a couple of months. One Tuesday morning he spent about twenty minutes describing it while CodeRhapsody checked it against the code. He took a shower; the first commit landed at 7:42, before breakfast. He also described starting several new chats to recover from deteriorating context.

The useful idea beneath the title is modest: the context window is a working set, not the archive. A file result may answer one question and then occupy every subsequent request. The user's goal and the reasons for an edit have a different lifetime.

The current design removes tool traffic without asking a model to summarize away the conversation. Agent chooses cuts at safe actor boundaries, History records and applies them, and Engine renders the resulting Context. Workers neither trim History nor decide when the next model request starts. In C#, those responsibilities use Common declarations and direct partial implementations, not a new reduction service.

## Tool traffic ages in two stages

The default byte target is 400,000. It derives a whole-result budget of one eighth and a call-and-stub budget of one sixteenth. These are serialized tool bytes, not tokens and not a hard cap on the conversation. Dialogue and manuals can keep growing.

When a band exceeds twice its budget, the actor selects an old prefix to bring it back toward one budget. First, full results become stubs while their calls and identities remain. Later, calls and stubs disappear together. The second stage cannot jump past a still-whole batch. Larger, less frequent cuts avoid continually rewriting the front of the request, although that alone proves no provider cache benefit.

Complete call/result batches are the unit of selection. Incomplete, orphaned, or ambiguous exchanges stop safe prefix selection rather than inviting the reducer to invent a missing result. A result becomes eligible only after a successful response has presented it to the model. Restoration conservatively presents retained results again before making new automatic cuts.

Every cut records its exact sequence bounds. Replay applies that fact, not today's policy. Raising the target after restart cannot resurrect bytes already removed, and lowering it cannot retroactively change an earlier decision.

An optional, explicitly enabled policy treats large results more aggressively. After a successful request presents a result larger than one hundredth of the target, the consuming response can call `keep_tool_results` to veto immediate stubbing. That decision happens once, after all companion tool outcomes are durable. A later response or restart must not reconsider a kept batch. Keeping defeats this immediate policy, not eventual budget pressure. A similarly named custom tool does not acquire the control's authority.

Stubs make a deliberately limited promise: bytes were removed. An earlier implementation cited output paths whose numeric handles were reused after restart. A plausible path then returned plausible output from the wrong command. An addressless stub is less convenient and considerably less misleading. Rereading a file may recover useful current information, but not necessarily its original contents; rerunning a command may have effects. Historical identity belongs in preserved events.

## Keep the explanation and the manual

Once tool bytes leave context, useful visible narration becomes valuable working memory. A short explanation of why a file is being inspected and what the inspection established preserves more than an unexplained read/edit sequence. It need not expose private reasoning or narrate every keystroke. Nor does its survival make it permanently true: changing sources still need to be checked.

Skill manuals already belong once in provider instructions. Their load results are expendable traffic; their instructions are not. Unloading immediately revokes grants that no other active skill supplies, including deferred discovery references and automatic snapshots. Dispatch refuses a guessed unavailable name, and owned server connections close. The pending manual nevertheless remains in SYSTEM through tool clearing and handoff. Removing permission and retiring instructions are different operations.

The earlier positional `compress_context` experiment gave the model much broader destructive choice. Bill's account described repeated loss of most of the conversation, including the goal near its beginning. A later whole-task handoff started a fresh instance from a document, making another durable-state path responsible for continuity.

`micro_handoff` is narrower. It records an ordinary successful tool result, waits for the batch's other outcomes, and then records a distinct handoff fact. Interrupted undispatched calls receive their exact-ID refusals before the cut. Independent Jobs continue; forgetting an observation does not cancel its work.

Applying the handoff removes tool traffic, opaque reasoning, and private replay signatures, including signatures attached to visible text. It retains visible dialogue, system entries, and manuals. Older active handoff notes disappear, and one new marked note takes their place. Its identity is structural, not a magic phrase at the start of prose. Engine projects it as ordinary system text only when building a provider request.

A useful note identifies what succeeded, what failed, where the current truth lives, and what remains. Failed approaches often prevent more wasted work than a catalog of successes. Original notes and removed private parts can remain in the archive even though they leave the working view.

Manual clearing uses the same actor boundary. It removes tool traffic after completed batch outcomes without resuming a paused turn or waiting for every independent Job. The GUI cannot accomplish that by deleting cards locally. A recorded cut triggers silent replay of the reduced Context so old cards and late partials cannot reappear, and refresh does not speak the conversation again.

## Record before publishing

A shutdown snapshot cannot preserve work completed before a process is killed. History therefore owns a recovery journal as well as snapshots. Writers serialize through its existing save gate: compute the next fact, write and flush it, then publish the committed Context and sequence. Slow disk work does not hold the short state lock that readers need to capture the last committed view.

If append fails, nothing new becomes live, and subsequent append/save attempts refuse. Reusing a possibly half-written sequence would manufacture ambiguity. Cleanup must still join work and detach the Agent from its parent while preserving the original recording failure.

Recovery validates candidates independently: primary plus a contiguous journal suffix, otherwise backup plus its suffix. A journal beginning at sequence one can recover a session with no snapshot. An incomplete final record may be ignored with a diagnostic; malformed complete records, gaps, and duplicates are errors. A failed candidate never becomes a partly installed conversation.

The easily missed retention boundary is the backup. If the backup ends at 100, the primary at 150, and the journal at 160, deleting everything through 150 destroys recovery from the backup when the primary becomes corrupt. Saving validates the new snapshot and old primary before rotation, preserves a good backup, replaces the primary, and only then prunes records at or before the validated backup boundary, retaining every newer record. Without a valid backup, the journal remains intact. This is process-crash recovery, not a claim of hardware power-loss durability.

Reduction itself distinguishes an inapplicable known fact from corrupt input. A recognized cut with unusable bounds can be diagnosed and skipped while later facts apply. Unknown kinds, unsupported versions, malformed identities, and broken sequences still refuse decoding. A blanket catch would hide invariant failures rather than make replay robust.

## An archive cannot be reconstructed from what remains

The optional archive contains original events, including tool payloads and the cuts that removed them from context. Before saving it, History merges available archive, journal, and diagnostic events and requires a contiguous sequence with identical overlaps. Missing early facts cannot be manufactured from the reduced snapshot. Turning archiving on late may therefore fail honestly rather than produce a file misleadingly called complete.

The live byte target has one authority: Agent. The GUI reads that value and publishes a settings change only after validation and saving succeed. There is a consequential interface difference: library and CLI zero disable automatic reduction, while GUI nonpositive values select the default. Sharing a parser merely to conceal that difference would not simplify the semantics.

The source's current checks use local and loopback paths; the breakfast implementation and earlier model experiments are historical, not fresh C# performance results. The design's durable contribution is more concrete than a memory claim: the system can explain which bytes left, why instructions survived, and how the same working state returns after a crash.
