# Chapter 8: Build Your Dream GUI

A settings control makes a promise about the running system. If it says the font is larger, existing transcript text should grow. If it says the next turn permits three tool batches, the Agent should capture three when that turn begins. If saving fails, the page should not display a preference that only briefly existed in JavaScript.

The GUI keeps this promise by separating three lifetimes. The browser displays the last acknowledged snapshot. The server owns presentation preferences and their persistence. The Agent owns its next-turn tool budget, which an active turn captures and then keeps. None needs another scheduler.

## Comfort without a second application inside the page

The layout remains deliberately small: one transcript, a prompt area, wrapping controls, and a native expandable Settings panel. It uses ordinary document flow rather than fixed overlays or a collection of nested scrolling panes. Code can scroll horizontally; long messages remain in the transcript. Theme and font changes apply to content already present as well as to new messages.

Light, dark, and system themes share CSS variables, with system theme following the operating system’s preference. Native controls, associated labels, visible focus, and narrow-width behavior do useful work without requiring a custom widget framework. A geometry assertion can reveal overflow; it cannot establish that every reader will find the page comfortable.

Narration and per-artifact speech remain separate. Disabling automatic narration does not remove a card’s speaker button. Changing speech rate affects the next utterance constructed, not sound already playing. Turning off pause-while-speaking removes the speech-related pause cause without cancelling speech. Genuine typing remains an independent cause, so clearing one reason to pause need not clear them all.

## One acknowledged value

The server holds a value-like preferences record for theme, narration, speech rate, font size, and speech-related pausing. A value snapshot needs no parent links or destruction machinery. The tool budget is deliberately absent from that record: the server reads it from the actual Agent when constructing a complete wire or disk snapshot.

A browser sends a patch, not a new authoritative settings object. Under the host’s short settings lock, the server merges the patch with current values, validates it, and writes the complete candidate to a unique temporary file beside the destination. Only after replacing the destination does it update live preferences and Agent policy, then queue the same full snapshot to every connected client. Network sends remain outside the lock.

This order makes the acknowledgment meaningful. A write or replacement failure leaves the accepted in-memory values and prior good file unchanged, reports an error to the requester, and cleans up the owned temporary file. The page reverts its controls to the acknowledged snapshot rather than retaining an optimistic fiction. Two tabs converge on the server’s published state instead of maintaining rival local-storage copies.

The default file lives at `$HOME/.ensemble/settings.json`; an explicit application setting can select another path. Same-directory replacement avoids exposing half-written JSON. It does not promise power-loss durability or provide a backup archive. These are preferences, not credentials or conversation events.

## Reading a bad file is different from failing to save

Disk loading and incoming patches use the same validation policy. Invalid fields retain their prior values; loading starts from defaults. Types stay meaningful: a JSON Boolean is not interchangeable with a string or an integer. Valid numeric speech rates and font sizes are clamped to supported ranges; a tool budget must be a nonnegative integer. Unknown fields do not become accidental new settings.

A missing, unreadable, or malformed file falls back to defaults without rewriting or deleting the original. This is an intentional change from the older edition’s fail-startup policy. Broken presentation preferences should not lock the human out of the interface needed to recover. Keeping the file untouched preserves an opportunity to inspect it, although a later successful settings edit may replace it.

Read fallback and save failure therefore have different consequences. Startup can proceed with defaults when no usable preference snapshot can be loaded. An unsuccessful update cannot claim that a newly requested value was accepted. Both rules keep the page honest while avoiding unnecessary downtime.

## A budget belongs to a turn

The Agent’s runtime policy is synchronized through its ingress lock. The GUI can read or update the next-turn value without examining actor-private state. The constructor configuration seeds this policy; it does not remain a competing mutable settings object.

Zero means automatic, resolving to 200 tool batches when a turn begins. It stays zero on disk, in acknowledgments, and in the control. Only the captured effective limit becomes 200. A queued prompt captures the policy when it actually starts, so a settings change can affect queued work without changing a turn already in progress. A hint continuing that turn does not mint a fresh budget.

The earlier library constructor’s negative unlimited setting remains supported, but the new runtime setter and GUI accept only nonnegative values. Preserving an older API contract does not require exposing that convention as a new preference.

A batch is one model response containing tool calls, regardless of how many calls it contains. After the last admitted batch, the model still needs to consume the results and answer. The Engine adds a request-only instruction to finish without more tools. That instruction does not enter History or modify signed replay content. If the model nevertheless requests another batch, each call receives a refusal before any effect begins.

The budget thus has both an explanatory side and an enforcement side. The request tells the model how to finish gracefully; the actor prevents further execution. A settings panel that only changes prompt wording would offer a suggestion where it had promised a limit.

Source: [Chapter 8](../chapter-08.md).
