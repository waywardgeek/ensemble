# Chapter 14: The World's Best Context Engineering, Before Breakfast

Bill had the design in his head for a couple of months. One Tuesday morning he spent about twenty minutes describing it while CodeRhapsody turned it into a design record and checked it against the code. He took a shower; the first commit landed at 7:42, before breakfast. The ambition was to stop treating a long-running agent's memory as a transcript that eventually needs an indiscriminate summary.

The familiar symptom is an agent that starts rereading files and losing decisions. A new chat helps because a human chooses what to carry forward. The [source chapter](../edition-3/chapter-14.md) moves part of that judgment into the runtime: identify expendable tool traffic, remove it at explicit boundaries, and preserve the instructions and decisions needed to continue.

The event log records history. The context window is a working set rendered from that history. Removing bytes from the working set must not erase the record of their removal.

## Give Every Entry a Removal Rule

The source reports that tool traffic dominated the measured sessions from Chapter 2: file bodies, command output, searches, stack traces. A file can be useful for one turn and wasteful on later turns. The initial target is therefore tool traffic, not dialogue. This choice limits disruption, but it also limits the guarantee: the chapter does not bound the whole context window. Dialogue and retained instructions can still grow.

Each entry gains a kind identifying what can remove it. Dialogue contains ordinary conversation and tool traffic. Handoff contains a checkpoint written by the actor. Skill contains a loaded skill's rendered instructions. Tools contains changes to tool declarations. The last three are survivors: they never contain tool calls or results, so clearing tool parts cannot accidentally remove them.

This classification separates meaning from transport. A skill arrives through a tool call, but its instructions are not expendable merely because the load operation returned a result. Similarly, the checkpoint must outlive the call that created it. Shared vocabulary belongs in `internal/common`; History owns reduction of events into active context. Putting removal behavior on the common structs would give shared data a responsibility that belongs to the history owner.

Bill's earlier `compress_context` experience explains the caution. The model chose a message range and replaced it with its own summary. Bill reports that Gemini repeatedly cut from message one, where the user had stated the goal, and produced a poor summary. The model lost dialogue that existed in no file it could reread. The irreversible choice was not merely which output to forget, but which account of the task would remain.

The replacement, `handoff_task`, wrote a structured document and started a fresh instance with an empty window. That made the note solely responsible for continuity. It also added another path for durable state: CodeRhapsody's handoff path did not trigger the memory cascade used by `save_memory`, a gap that remained in its notes for months. Ensemble builds neither tool. Its micro-handoff stays in the same conversation and uses the existing event path.

This design gives the actor a narrower job: decide what a completed piece of work has established. The runtime handles tool-byte removal through explicit rules. Visible dialogue, including the user's goal, remains intact.

## Record the Cut, Not Just the Policy

Context management uses a target size, T, measured in bytes. Zero selects the default of 400,000; other targets are clamped to at least 20,000. T is a policy input rather than a hard cap on total context. It determines a result-stub threshold of T/100, a whole-tool-traffic budget of T/8, and a calls-and-stubs budget of T/16. At the default those are 4,000, 50,000, and 25,000 bytes.

Tool traffic ages in two stages. A `RedactResult` event replaces older result bodies with stubs but leaves their calls. A `RedactTool` event removes older calls and results together. A band can grow to twice its budget before a single event cuts an old prefix back toward the budget. That hysteresis avoids a stream of tiny cuts around one threshold and leaves a recent stretch of traffic intact. Under a prefix cache, changing an earlier byte invalidates the suffix after it. Cutting in steps preserves the useful prefix longer than shaving a few old bytes on every request.

Only completed batches already seen by the model are eligible. A result that has arrived but has not yet been included in a request is not expendable under this rule. Nor may a cut divide an unfinished call/result exchange. Batch completion and exposure to the model matter more than the age of an individual part.

The policy measures serialized tool parts, excluding dialogue and survivor bodies. Each event records its selected sequence boundary as a number. Replay applies that number; it does not recalculate the cut from today's settings. Otherwise changing T could change the apparent past: a restarted Agent might remove different content from the content removed during the original run. The retained event says what was cut, while settings explain how the live controller chose it.

The two removal stages preserve different amounts of continuity. A stub still tells the model that a call occurred and that its full output has been removed. Complete tool removal spends less context but loses that local trace. Neither cut deletes the underlying events. Their later availability depends on archive retention; a dialogue-only archive cannot reconstruct discarded raw results.

## Addressless Stubs

On a model configured with `StubsToolResults`, results above the threshold become stubs in the request after the one that first carried them. The model gets one request with the full result and can use its next response to call `keep_tool_results`. That keeps the previous batch together. If the keep call travels alongside another file read, it does not keep that new read's result; its own acknowledgement is never stubbed. Keeping prevents this immediate per-round-trip replacement, not the later budget-based removal policy.

Other models do not receive automatic per-round-trip stubbing. The budget ladder still applies. In the source's feature matrix, the course Opus fixture enables both stubbing and inline declarations, the course Sonnet fixture enables neither, and both flags remain false on existing real-model rows. A deterministic fixture is not evidence that a real provider supports its protocol.

Ensemble's stubs deliberately carry no address. CodeRhapsody supplied the counterexample: its stubs cited output files such as `cr/io/26`, but its handle counter restarted at one on every launch. After a restart, that path could name another command's output. The actor followed a plausible address, received plausible bytes, and had no indication they were wrong.

The first edition records CodeRhapsody's fix: keep the counter increasing across restarts. Ensemble makes the narrower promise that the bytes are gone from context and leaves recovery to the tools that produced them. A file can be read again, but it may have changed; rerunning a command can also have side effects. Neither operation guarantees retrieval of the original result. The original `ToolReturned` event preserves those bytes only while archive retention keeps it.

## Checkpoint at the End of a Batch

`micro_handoff` lets the actor supply a short working note at a completed step: what succeeded, what failed, where the evidence lives, and what comes next. The note is the actor's selection of useful state, not a second model's summary of an arbitrary half-window.

The operation consists of three records: its tool call, an ordinary paired result, and a `MicroHandoff` event. The event's reduction waits until the batch's last result arrives. This matters even in a serial dispatcher. A tool beside the handoff may not have started when the handoff call returns; clearing immediately would allow that later result to enter context after its call had disappeared. Siblings before or after the handoff must complete before the batch is cleared.

The committed event removes every tool call and result part, opaque reasoning, and private replay signatures, including signatures attached to visible text. Visible dialogue remains. Empty entries are dropped. Earlier Handoff entries leave the active context, and the new Handoff entry is appended, so only the latest checkpoint remains active and its text appears exactly once in the next request. Skill and Tools entries survive. Archived events and their sequence boundaries do not change.

Ordinary redaction is less aggressive. It preserves non-tool parts, including provider-private state; a handoff deliberately removes that hidden state while retaining visible conversation. Confusing these operations either leaves supposedly cleared private context behind or strips state that an ordinary tool cut was not authorized to remove.

The handoff event must be appended successfully before it replaces committed context. If that append fails, the earlier handoff and existing tool and reasoning state remain. The ordinary tool result is not, by itself, a durable checkpoint. Conversely, a later snapshot failure does not undo a handoff event already committed to the journal. This failure case is important because the happy path can look correct even when mutation precedes persistence: only the interrupted write reveals that recovery has no event explaining the in-memory cut.

Keeping every handoff active would rebuild the growth problem in another entry kind. The new note therefore carries unfinished work, decisions, and still-relevant commitments, not an accumulated history of completed tasks. Replacing older active notes does not itself remove their archived events. The checkpoint belongs at a completed micro-goal, not a token threshold: clearing midway through a task can discard the result that still holds its unfinished state. Failed approaches deserve space in the note because otherwise the actor may repeat them.

## Instructions Survive Their Load Calls

A `SkillLoaded` event records the rendered body and reduces to a Skill entry. Subsequent requests compose that body into SYSTEM using Chapter 9's instruction path, including after tool clearing and restart. Replay uses the recorded body rather than reopening a skill file whose contents may have changed.

The paired load result is not a second copy of the instructions. It is an acknowledgement on ordinary paths, or reference-only content on Anthropic's supported deferred-tool-search path. Mixing acknowledgement prose or a manual into that reference-only result changes the required wire shape. The instruction body remains owned by the Skill entry.

Unloading has a deliberately narrower effect than erasing instructions. It removes grants that no other active skill supplies and stops the skill's servers, but leaves its Skill entry contributing to SYSTEM. Ordinary clearing and this chapter's handoff do not remove it. A later verb owns that removal. A successor changing unload behavior must account for both capability and instruction lifetime; they are not the same state.

The initial primary instructions remain stable. Loading a skill changes SYSTEM and may invalidate a provider cache prefix. The design does not promise an unchanging cached prompt, nor clear history to conceal that change. Preserving instructions is more important than presenting a misleading cache claim.

Tool availability changes have their own event, `ToolsChanged`, carrying complete added declarations and removed names. Its `tool_decls` part is neither a call nor a result. It therefore survives tool clearing without violating the rule that survivors contain no tool traffic.

Only the synthetic `InlineTools` course protocol freezes startup declarations and renders later changes inline in dialogue. Real models keep supported Chapter 9 disclosure paths: updated declarations or Anthropic's full deferred catalogue with references in genuinely paired successful load results. Remembering an inline declaration as history does not make it a valid provider request.

## Clear Context Without Breaking the Next Request

A result can be correctly retained or removed and still leave an invalid wire message. The source's ordering failure makes that distinction concrete.

The first student reducer emitted a Tools entry after each individual result. The renderer treated that entry as a fresh user turn and placed declarations between two results from one call batch. An Anthropic-like ordering guard rejected the request. The required placement is after the complete user, tool-result, and ephemeral turn, not merely after the result that happened to trigger a declaration change. Serial tool execution does not remove the problem: the model issued one batch, and the renderer must preserve that request-level structure.

Calls and results form a unit even when their entries are separate. Clearing both sides of an exchange preserves that unit; retaining a result after removing its call does not. A declaration entry can survive cleanup without belonging between the parts of a tool exchange. The reducer's entry kinds and the renderer's message ordering solve different problems.

Replay and provider validity answer different questions. Replay can faithfully reproduce an invalid transcript. Inspecting the next request independently can establish pairing, declaration placement, and absence of hidden reasoning after handoff. An event-count assertion cannot establish those properties, and a test that uses the renderer's own helper to construct its expected request can reproduce the same mistake on both sides.

## Keep Malformed Payloads From Partly Applying

History reduction is total over recognized events within a structurally valid stream. A known event with an unusable payload, such as a skill load without a body or a redaction naming absent entries, is diagnosed and skipped before mutation. Later valid events can still apply. A useful diagnostic identifies the skipped event and reason so the missing effect can be investigated. Skipping advances the replay position without changing context.

This deliberately changes Chapter 10's treatment of malformed known payloads. It does not relax JSON decoding, version, anchor, or sequence validation. Unknown event types still refuse the whole recovery candidate atomically, without installing an Agent or overwriting either recovery file. An implementation that cannot interpret an event cannot safely pretend that skipping it preserves meaning.

The distinction prevents two opposite failures. Panicking on every inapplicable known payload makes one bad operation prevent recovery of later valid work. Mutating partially and then announcing that an event was skipped leaves state with an effect that neither the reducer nor replay can explain. Applicability must be checked before any change.

## Recovery Starts at the Backup's Boundary

A snapshot taken only at orderly shutdown cannot recover work completed after the last save when the process crashes. Chapter 14 requires complete events to reach an independent journal before publication or acknowledgement. Recovery can then reconstruct journaled progress after a process crash instead of depending on the last clean exit.

The journal's retention boundary is the validated backup snapshot, not the newest primary. In the source's example, the backup represents sequence 100, the primary 150, and the journal reaches 160. Keeping only events after 150 would leave no route from the backup to the newer work if the primary later proved corrupt. Recovery from that backup still needs 101–160.

Saving first encodes and validates the new state, then preserves Chapter 10's validated backup rotation and atomic primary replacement, and only afterward trims the journal. If no usable backup boundary is established, the required recovery tail must remain. A nominal backup filename is not enough; its validity determines whether it can justify deletion.

Recovery selects a valid primary or backup independently. It applies the selected snapshot's embedded tail, then applies only journal sequences beyond what the recovered state already covers. Overlap is expected and must not duplicate events. New sequence allocation continues beyond the recovered tail. Sequence boundaries make the merged history exact rather than a guess based on file age.

Append, save and rotation, and trimming must serialize at the actual event writer. A lock around an unrelated caller does not protect a concurrent append from being lost during replacement or trim. Likewise, failed append or save operations must preserve the previously committed recovery and diagnostic bytes. Retry must neither duplicate an acknowledged event nor skip an unacknowledged one.

The promise is process-crash recovery. The source does not claim universal power-loss durability across operating systems, filesystems, and devices. A forced process exit before orderly save tests whether the journal carries recent work; it does not prove that every storage layer has made those bytes immune to power failure.

Diagnostic retention is separate from canonical recovery. Chapter 10's `log_mode` selects the archive policy and `diagnostic_max_records` limits diagnostic records; neither supplies a last-N cap on events needed to recover from the backup. After a successful handoff recovery save, the existing diagnostic writer performs its usual retention operation, the same one used at startup and exit. Its default is 200 whole records, zero means unlimited, and negative values are invalid; continued writes must still work afterward. Failed saves preserve both kinds of bytes. Creating a second truncation path would make retention depend on which lifecycle action happened to run.

## One Policy Setting, Correct Ownership

The Context Management tab exposes `context_target`, not independent sliders for every derived threshold. One setting keeps the result threshold and the two budgets in proportion. It also avoids asking a person to predict the interaction of several removal policies before seeing a session run.

Settings belong to the owning Agent's data directory. The GUI uses `DataDir/settings.json`; the supplied application chooses `cr`, making its path `cr/settings.json` relative to its launch directory. A similarly named file beside the executable is not interchangeable. The browser's public settings update reaches the owning Agent before subsequent policy work.

The browser must preserve unsent edits when a settings snapshot arrives. A server update is information about persisted state, not permission to overwrite a person's draft. The source's boundary checks tie editing to real requests: changing T must affect the next request without rewriting previously recorded redaction boundaries. The visible field, persisted value, live controller, and replayed past each have different responsibilities.

The architecture follows those owners. The Agent chooses policy from its configuration; History validates, journals, and reduces; Engine translates the result into vendor requests. Shared types do not acquire runtime behavior, and a new settings callback should not bypass the owning Agent. The GUI remains a separate public consumer, and the core retains the Chapter 1 client's behavior without taking a GUI dependency. A screen that can display the new setting does not establish that cleanup or recovery works, and a correct core reducer does not establish that the browser preserves an unsent edit.

## Smaller Requests, Unsettled Model Behavior

A synthetic reference session in the first edition sent about 35 percent as many bytes with the ladder as without it, and about 13 percent with both the ladder and stubs. That measured the effect on scripted traffic, not real-provider cache behavior or answer quality. The stubbing run never triggered the ladder because it kept both bands below their budgets.

A historical `claude-opus-5` run illustrates the model's part in the policy. It read seven files, reread two, and never called the available `keep_tool_results`. The two rereads were the file it had not narrated and the largest file, which received only a short note. Files with fuller notes were not reread in that run. Dialogue preserved what the model wrote down; another tool round trip restored what its notes missed.

That observation supports a possible working pattern, not a general claim that one exposure is enough. The source explicitly leaves open whether stubbing interferes with model thinking. This run also reported no cache reads or writes: the reference renderer set no `cache_control` breakpoint. Preserving a stable prefix did not mean the application requested a cache. The run enabled features that this edition leaves false on real-model rows, so it is not a current deployment recipe.

The later multi-model observations show why correct cleanup is not enough. In a corrected-path Haiku 5.5 session, all six full file results arrived before redaction. The model subsequently claimed, "That file was redacted, so I never read it," and refused the handoff. The transcript disproved that account: cleanup removed already-presented bytes, not the original read. GPT-4o mini completed reads but repeatedly checkpointed until it reached the turn's batch limit. Of the five historical processes, all exited zero and only the Gemini session completed the requested task. Process success did not establish task success.

Those historical 5.5 names remain historical. Bill reports that 5.5 models do not work with ephemeral context and excludes them from supported use. The source names `claude-sonnet-4-6` as a supported live choice, without claiming that support guarantees successful completion.

Its bounded observation of the accepted Edition 3 build used that Sonnet model. Six human turns completed with full file results, six result cuts, and one tool cut. After handoff and unload, the next captured request retained dialogue, SYSTEM manuals, and the handoff note but no tool traffic. The unloaded manual remained pending while `read_file` was no longer available. The observation guard then stopped on an unexpected or repeated handoff before the seventh final answer. The receipt remains failed: it establishes neither a completed post-handoff turn nor normal EOF nor live recovery. Passing local Chapter 14 and separate CLI Chapter 1 checks on the same held source did not turn this paid observation into a successful run.

The automated checks have their own limits. An early two-band test used paths too short to fill the calls band and therefore tested only result stubbing. Placing malformed events at the very end let a loader stop early without losing later valid work. Meaningful evidence must show both cuts actually occurring and valid events applying after a skipped event. Independent next-request inspection, disk-failure ordering, and browser edit behavior cover failures that a final model answer cannot reveal.

Dialogue still grows, even when tool cleanup works. The historical reading run's baseline request size rose as narration, calls, and stubs accumulated. Chapter 22 addresses dialogue removal. Until then, T remains a tool-traffic policy rather than a bound on the entire window, and reconstruction of earlier windows remains conditional on retaining enough canonical history.
