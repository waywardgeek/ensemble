# Chapter 21: Forgetting on Purpose

Bill once took an agent experiment that worked remarkably well at home into his workplace, where it failed. He initially blamed the model. The important difference turned out to be a bug: memory worked at home and was completely broken at work. Fixing memory reversed the apparent comparison between models.

That historical account motivates a feature; it is not a controlled benchmark. The engineering problem is concrete enough without one. A new session can have the source code and still lack the reasons for it. A summary can preserve those reasons, but accepting an unreliable summary must not destroy the only usable account.

Memory therefore makes forgetting a replacement transaction. The old material remains active until a bounded worker's proposal has been validated and committed. History, working Context, human-readable files, and the model's judgment have separate roles throughout.

## Keep different rates of change apart

Context presents five bands from stable to recent: SOUL, MEMORY, the 64x archive, the 8x archive, and session memory. Conversation comes afterward. Compression flows in the opposite direction, from completed conversation into sessions and then through progressively coarser accounts.

The names describe desired resolution, not a promise that a model will compress every input by exactly eight. SOUL is outside automatic inbound graduation. Making identity available to the model does not authorize the compressor to rewrite it.

On disk, the bands use `SOUL.md`, `MEMORY.md`, `bucket-1/`, `bucket-0/`, and `memory/`. An Agent owns Memory, which reaches rooted files through that Agent's Sandbox. The worker does not choose a destination path. A library Agent without a home does not quietly acquire filesystem memory somewhere convenient.

Budgets here are UTF-8 bytes, not estimated tokens. The default conversation budget is 3,000 bytes, with 1,500 for each corpus band; high and low watermarks default to twice and once the resolved budget. These are trigger and writing policies, not hard promises about every result's length. In particular, the target byte budget does not turn an otherwise completed, nonblank candidate into a failed submission merely for exceeding that target.

Six immutable policies govern conversation and the five bands. Disabled bands leave disk intact while retiring their active projection. Re-enabling reads the current corpus. Policy saves precede publication, and acknowledgements describe what the selected Agent actually applied. A preference change cannot safely race a worker that captured a different source or destination policy.

## A checkpoint preserves the boundary

A handoff still removes old tool traffic and private replay from working context. It also leaves a readable continuation note. Earlier handoff notes now remain as completed-segment boundaries until a matching memory replacement commits.

When retained conversation exceeds its budget, selection takes a whole oldest completed prefix, stopping when the remainder fits or no further complete prefix is available. It never grabs the unfinished tail or pretends an unresolved tool call has a result. Selection itself changes nothing.

File graduation has a different condition: it requires at least two source files. Strict dated identities, deterministic ordering, and overlap checks establish which files are active. A directory enumeration is not a chronology, and a convenient Markdown filename is not evidence that a staged candidate was accepted.

The archive remains more detailed than the request. History preserves the canonical events; Context is a reduced working view. Memory renders as labeled user/data material before conversation. It cannot grant a tool, become a system instruction, or turn a remembered claim into current authority.

## Capture what the replacement means

An attempt records its task identity, exact selected source, prior destination bytes, fingerprint, and child trace path. Later unrelated dialogue may arrive without invalidating the capture. Editing a selected file or the destination does invalidate it.

Comparing a whole-History version would discard useful work whenever the user typed another sentence. Comparing only a task ID would accept a summary of bytes that no longer exist in that form. The capture identifies precisely what the candidate is allowed to replace.

The actor records the launch before starting its one owned compressor. Parent work can continue while the child thinks, but acceptance waits for a complete main-work boundary. The child returns immutable data through the mailbox; it never edits parent History directly.

Relevant policy changes cancel and join the attempt. Failure retains source and waits for a later checkpoint instead of immediately retrying forever. A successful commit may advance the next eligible file graduation. These transitions belong in the existing actor, not a second scheduler polling for signs of readiness.

## The compressor has a small job and a real end

An early direct implementation made one Engine call and expected a finished handoff. The model needed another turn, returned no final text, and the attempt failed. Unlimited calls would have hidden the missing completion rule rather than supplied one.

The current worker is an ordinary child with both owning parents, but without inherited skills, recursive memory, file access, shell access, or web tools. Its only tool is `submit`, accepting one required string named `memory`. The prompt supplies the source and existing destination as data. The model proposes text; the owner controls files and acceptance.

The actor permits at most three Engine sends and three tool batches. A batch containing several calls is refused before any runs. After accepted submission, the turn ends without asking for closing prose. Ordinary final text is not submission. Blank output is a terminal failed candidate; malformed shapes can consume the remaining bounded correction opportunities. Truncated or incomplete provider output cannot smuggle a plausible-looking submission through the terminal boundary.

The prompt asks for decisions, reasons, failures, uncertainty, and unfinished work. It asks for first-person continuation only where the source supports it, with attribution for somebody else's experience. No runtime check can prove that a fluent summary preserved the right judgment. That still requires reading the result and testing what later work can recover.

## Prepare, commit replay, finish disk work

Acceptance first rechecks source and destination, stages a hidden candidate and receipt, and archives canonical events for a conversation cut. After that I/O, it revalidates before publishing the replacement.

The actor then records task-correlated removal and matching population. Removal alone holds a proposed cut; it does not make the source disappear. Matching population applies the held removal and installs the actual replacement text in one Context transition. Skill entries inside a conversation interval survive; a generic destructive range cut would remove too much.

Only then does recovery promote the staged file and retire exactly the captured inputs. A successful terminal outcome is recorded as `Committed`, after which the receipt can be released. File finalization is an operation, not another success state. Retired files remain available on disk, excluded from active scans but available for later recall.

These steps cannot be one filesystem rename. A crash may leave a launch without a candidate, held removal without population, committed replay without file promotion, or a promoted destination with retirement unfinished. The receipt and replay state distinguish those cases. Startup resolves them before scanning the corpus.

Recovery completes the same accepted decision without asking another model what the summary should have been. Missing or conflicting required evidence causes refusal, not a guessed repair. Repeating finalization must neither overwrite unrelated edits nor apply an accepted transaction twice.

That is a bounded process-crash contract, not a promise of filesystem-wide atomicity against a hostile writer or every kind of power loss. The important invariant is that accepted replay, the live projection, and the corpus have a documented route back into agreement.

## Spending is not occupancy

The compressor's work costs money even when its result is unusable. After joining the child, the parent imports measured model, tier, billing, and usage rows exactly once. Unknown counts remain unknown. Replay of documentary usage does not charge the new process again.

Those rows must not replace the parent's latest direct-response observation or cache lens. Cumulative spending and the main conversation's current occupancy answer different questions.

Context pressure uses the latest actor-accepted main response and a known window for the selected model. Provider accounting differs: OpenAI cached input is included in its total input, while Anthropic exposes separate categories. A child's response or a session-wide token sum is not evidence about the main request's occupancy. Unknown limits remain unknown rather than being guessed from a model name.

At 90% occupancy, a request-only instruction asks for a handoff. At 95%, the runtime latches a checkpoint requirement: both declarations and dispatch admit only an already-authorized `micro_handoff`. It does not grant that tool merely because it would be useful. A later short or usage-less response does not silently clear the outstanding requirement.

If a compressor already owns the source needed to regain headroom, the actor waits for that exact attempt before another main request, while continuing to process controls. A matching committed replacement releases the wait. Failed or invalidated work produces an explicit headroom error, not another hopeful request into an already pressured context.

## Keep one owner for each fact

The memory cache boundary follows the most recent surviving handoff, or the final memory block when no handoff survives. That structural choice supports stable prefixes; it does not guarantee a provider cache hit.

The broader discipline is the same. Files supply corpus bytes through Memory. Accepted events supply replayable state. Engine renders the provided Context. Rendering does not rescan files, launch compressors, or decide which history should vanish.

The original source remains available long enough to reject a bad proposal, recover interrupted finalization, or retrieve a detail the summary omitted. The next chapter gives the Agent a habit of looking for those details before it repeats a mistake.

[Implementation chapter](../chapter-21.md)
