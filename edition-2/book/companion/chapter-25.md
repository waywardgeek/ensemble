# Chapter 25: Crossing Over to Ensemble

When the second-edition agent became the place where Bill and Cadence worked on the book, daily use exposed failures the construction tests had missed. The [current source chapter](../edition-3/chapter-25.md) draws this history from the [original crossover record](../../docs/edition-2-notes/ch25-crossover.md), not from fresh Edition 3 execution.

The speech-speed slider stopped at 2×, while the native voice rate was nonlinear. The fix changed both the control and saved-setting validation, using an approximate mapping above 2×. But the launcher still ran the old executable with its embedded assets. Later, restarting the server left an open browser running old JavaScript. Building, installing, restarting, and reloading affect different parts of the application; observing one doesn't establish the others.

A separate failure came from a 60-second total HTTP deadline that killed legitimate model responses. Removing it preserved explicit cancellation and connection protections. The distinction now introduced in Chapter 1 is between ending an observation and ending the work: a short job callback provides a progress interval, not a lifetime limit.

Automatic browser snapshots and speech-queue observations made every round noisy. Making those two tools on-demand addressed the cause without discarding general ephemeral context. Chapters 11–12 now place observation and confirmation under explicit control.

## Boundaries exposed by use

Bill also needed visible explanations before actions. Private or encrypted reasoning couldn't supply them. The implemented model capability requires ordinary, non-whitespace assistant text before the first tool call in the same batch. Rejected batches return corrective results without dispatching their tools. Chapters 3 and 6 own this check because dispatch and stream order become observable there. It enforces presence and ordering, not honest or useful rationale; that remains the agent's responsibility.

Application settings moved into one explicit, versioned configuration file, with credentials referenced separately. File mode isolates provider settings from inherited environment; legacy startup remains available for tests. Provider changes require restart, while ordinary GUI and memory settings apply immediately. A valid update must be persisted before it is accepted in memory, and secrets must stay out of flag defaults and help output. Chapter 21 owns this integration and GUI startup. Chapters 22–23 extend its strict schema and immediate-application contract to memory and recall; Chapters 16 and 20 own provider capabilities and reasoning choices, and inherited running configuration and role instructions, respectively.

Other repairs clarified lifecycle ownership: root conversation doesn't require worker submission; completed handoffs retire old opaque replay from active context; compaction runs as a restricted child; cleanup waits for that child's final save. Reducing working context is different from erasing archive facts.

That distinction mattered in a separate restart failure. A new response reused an old display ID and entered its card, although the messages remained in the archive. Chapters 7 and 10 clarify the identity obligation. They do not establish that the copied baseline or legacy saved collisions have been repaired.

These repairs belong where their mechanisms are introduced, not in a hidden crossover implementation pass. Chapter 26 continues from the exact accepted Chapter 24 code. The historical record supplies experience; Edition 3 still has to establish its own execution results.
