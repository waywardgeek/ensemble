# Runtime speech/selection pause observation — 2026-10-16

## Finding and limits

Bill observed apparent worker stalls and proposed that switching agent views with read-aloud enabled leaves the old agent paused; its inherited speech speed also appears wrong. Read-only inspection supports a **source-level pause-routing defect**. This is NOT a Chapter1 student defect or a new Chapter1 contract. No application edit, deployment, restart, live reproduction, forced unpause or settings change was performed by the orchestrator during diagnosis.

Current repository code (`agent/gui/web/gui/app.js`, `agent/gui/tree.go`, `agent/gui/gui.go`) shows:

1. Speech enqueue sets `speaking=true`; `pace()` sends unaddressed `pause` on a local cause transition.
2. Agent selection sends `select_agent` without first cancelling the old agent's speech.
3. `selectConversation` queues the ordered new replay and sets `s.selected[conn]` to the new view under the same server lock.
4. Browser receives `agent_selected`, calls `speech.cancel()`, which may transition `pace()` to unaddressed `unpause`.
5. Server dispatches pause/unpause through `s.selection(conn)`: now the NEW view. The previous agent can remain paused. Human typing can also keep the local cause true; no claim of universal cancellation-triggered unpause.
6. The control labelled **Unpause everyone** sends that same unaddressed command; inspected handler unpauses selected agent only, not all agents.
7. New GUI views initialize `TTSEnabled:true, TTSSpeed:1`. Saved main speech speed at observation was3.51057095078834; saved student/author/text-review child speeds were1. All four saved TTS flags were false by observation time. Those current files do NOT prove when TTS was enabled during the earlier stall.

Read-only HTTP GET `http://127.0.0.1:8088/app.js` returned200, `text/javascript`,21706bytes; SHA256 **7910f2f15778d4d00f2f56d0839df70c7ddd52003c75292191931a393bbcc7b0**, byte-identical to inspected source. The live backend binary was NOT rehashed/source-matched in this diagnostic; server ordering above is verified current repository source. No controlled browser-switch reproduction or captured WebSocket pause trace established the earlier stall's exact cause.

## Observed delayed dispatch

Student final response event266 at18:38:12 contained visible explanatory text followed by complete submit call. No tool dispatch until267 at18:45:28; return269 and saved turn_ended270 followed immediately. Earlier write_file response256 ended18:30:43 and dispatched after human Please continue at18:37:26. During hold status displayed running/thinking. Runtime source has pre-tool pause wait without separate activity label, consistent with this symptom.

Parent first incorrectly described silence as still-finalizing, then overclaimed a broken lifecycle before checking pause. Those claims were corrected. Submission ultimately executed normally. This later source inspection establishes a specific routing defect, not a failed submission protocol and not definitive historical causation.

## Disposition

Keep as separate runtime issue for an authorized fix/test. Preserve Chapter1 acceptance and continue the bounded book loop. Do not silently alter speaker defaults, globally unpause unrelated agents, add Chapter1 machinery, or restart the running actor. A future fix should cover selection-time speech release with correct agent identity, inherited/browser preference policy, and truthful scope of the Unpause everyone control; precise policy must be confirmed rather than invented. Isolated regression testing is needed; do not attach test browsers to the live session.
