# Chapter 7: Everything Is an Artifact

Text, thinking, tool arguments, and results compete poorly for one terminal cursor. The browser gives each a stable place without acquiring authority over the Agent. Closing a tab must not cancel a turn or kill a job. A button sends the same public input that a non-GUI caller would send; the existing actor still schedules work and records what happened.

The GUI is an optional .NET 10 application using ASP.NET Core. Its composition root constructs the runtime and host. Its own Common declarations and server implementations remain separate from the core library, which knows nothing about ASP.NET. Connection queues and partial display text live for the hosting operation; they are derived state, not another transcript owner.

## Reconnection needs facts, not memories of notifications

The browser has two content sources. Finalized responses, human input, tool admission and results, and terminal facts come from History. Partial text for an unfinished response comes from the host’s accumulated observations. Treating both as a recent-notification list would make a lost notification indistinguishable from something that never happened.

History supplies an atomic snapshot of events and reduced context. Its append boundary validates, sequences, writes, and flushes before publishing the new state. The following observation may be dropped; the recorded event remains available. The server can therefore repair finalized content after a quiet interval or reconnect without inventing replacement deltas.

Connection registration, replay construction, and queueing share a short handoff lock. A new client gets replay before becoming eligible for subsequent frames. Network sends happen outside that lock, with a separate bounded output queue and writer for each socket. A browser that falls too far behind is disconnected rather than allowed to block every other client or the actor. Its recovery path is the ordinary replay path.

Ordering needs more care than “take the latest snapshot.” By the time an old delta reaches the host, History may already contain the final response. Publishing that newest snapshot first would erase genuine live progress. The live path projects through the observation’s recorded sequence; full catch-up happens on idle polls and connection establishment. Conversely, a reconnect that already received the final must not revive the part when an old queued delta arrives.

This restores a view of the same live Agent. Restoring an application after process restart requires the persistence work of a later chapter.

## The card has one identity

Every displayed message is an artifact: a stable card whose rendering varies by role and kind. A single card map handles human text, assistant text, thinking, planned calls, and results. Separate widgets do not require separate conversation models.

For model parts, Chapter 6’s parser identities now survive into recorded response metadata. The response identity is the actual request sequence; part IDs follow canonical part order while retaining the parser’s original values. The browser key combines Agent, response, and part. Human messages, results, and older records without parser metadata use a distinct deterministic event-and-position namespace.

That separation avoids collisions without a global display counter. A final replaces a partial card; replay reconstructs the same card; a repeated final is harmless. Tool-call IDs keep their provider correlation meaning, and opaque signatures are preserved rather than rewritten to suit a widget.

The page and its static assets come from the loopback GUI host. Its `/ws` connection carries replay, live display, and controls. Vendor SSE never reaches the browser: Engine has already translated it into the shared vocabulary. A subscribe request is idempotent, not permission to create another Agent or replay execution. The host’s history frontier is internal bookkeeping, not a browser-supplied authority over which facts exist.

## A plan is not a started job

Tool cards need three distinct moments. The model requested a call; the application admitted execution; an observation returned. `ToolCalled` establishes the first. `ToolStarted` records the second, before effects are allowed to begin. If that recording fails, execution cannot start.

A refused call therefore has a request and a paired failure result, but never appears dispatched. An admitted command that later cannot start is different: admission happened, and its outcome failed. Keeping those facts separate makes the UI useful for understanding both policy refusals and operational errors.

Even “finished” needs a precise scope. A foreground tool observation may return a running-job handle while the underlying job continues. The result card retains that handle and its locator rather than claiming the process exited. Likewise, turn completion comes from an explicit terminal fact, not a momentary Idle state between tool rounds. The answer is derived from the recorded response rather than stored in a second authoritative answer field.

## Compact presentation, complete content

Collapsing thinking or tool details saves space without discarding their contents. Scrollable code and readable search results preserve the rest of the evidence. A successful exact-replacement edit can show the applied replacement from its admitted arguments, but that is not a whole-file diff.

Markdown rendering constructs safe DOM nodes. Provider and tool text does not become executable HTML, and terminal escapes do not become browser controls. A source file containing HTML should appear as source, even when its contents would make an excellent attack on a careless transcript renderer.

The full backing text also feeds each artifact’s speaker button. Reading a card means reading the whole human message, assistant answer, thought, tool call, or result—not just its collapsed preview. Opaque signatures are never speech. Selecting another card replaces the previous speech; selecting the active card again stops it.

Automatic narration is a separate opt-in behavior. It reads new assistant text and thinking as complete clauses arrive, flushes the last remainder, and announces concise tool-call summaries. It does not automatically read tool results, human input, or replayed history. Reconnection should not trigger a recital of everything the human already heard.

## Speech and typing steer scheduling

The speech queue counts as busy before audio actually starts. Otherwise the gap between enqueueing and the browser beginning playback becomes a scheduling race. Both speech completion and speech failure settle the current item; callbacks from cancelled items are ignored by identity. The historical Chrome visibility workaround—an empty, silent utterance when the page becomes visible—stays outside that queue so it cannot manufacture a speech pause.

Typing or queued speech creates a local pause cause. Its false-to-true transition sends pause; its true-to-false transition sends unpause. The actor stops starting new model requests and tools while continuing to accept controls and worker completions. An in-flight request may finish and be recorded, with its planned tools waiting for unpause.

These controls are edge-triggered because pause is Agent-wide. Another tab or caller may explicitly unpause while this tab still contains text. Unchanged local typing must not reassert a permanent veto on every state update. A later local transition can issue another request, but the server’s effective state remains authoritative.

Loopback binding and origin checks protect this local control surface from casual cross-origin access. They do not sandbox the Agent’s tools. The browser makes work easier to inspect and steer; it does not reduce the consequences of the commands the Agent is authorized to run.

Source: [Chapter 7](../chapter-07.md).
