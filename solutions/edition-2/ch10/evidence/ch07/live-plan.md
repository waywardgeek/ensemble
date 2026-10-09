# Chapter 7 live feature/action plan — awaiting coordinator completeness check

Freeze source commit, every required source/support hash, GUI/CLI/embedding
binary SHA-256 and model discovery before paid launches. Identity adapters must
pass valid historical-source fixtures and isolated changed binary, source-map,
source hash and support hash negatives before any derived writes. Reuse allowed
Chapter 6 evidence machinery, preserving original transcripts/wire logs and
separately reconstructed requests. No retry loop and no secret argv/logging.
Read only the needed provider keys from authorized settings into process env.

## Bounded sessions and provider coverage

For Anthropic Messages, OpenAI Chat Completions and Gemini generateContent,
discover the exact accessible model IDs; Gemini target is
`models/gemini-3.8-flash` with the taught explicit model mapping. Run the same
matrix for each provider. Stop a failing attempt, retain it, diagnose locally,
and request coordinator review of a material scope change before another paid
attempt. Initial ceiling is 12 admitted prompts and 32 total HTTP model requests per
provider across interfaces (including every tool continuation); the local
recording proxy refuses request 33 before forwarding it. There are no retries;
normal planned work is lower. Tool-turn model continuations are recorded
separately from user prompts. Each session uses a scratch workspace and fresh log.

| Feature | Human interface action | Observable receipt |
|---|---|---|
| Human CLI retained and shared Agent | In an actual PTY, launch GUI `--terminal`, type a short ordinary question, inspect reply, then `/history` and a follow-up | Sanitized terminal shows streaming text/completion and durable sequences; browser snapshot of the same request IDs/answer |
| Streaming/plain CLI parity | Launch standalone public CLI in a second PTY with `EN_DISABLE_STREAMING=1`; ask a bounded comparable question, inspect reply, then `/usage`, `/quit` | Actual human terminal input/output, producing model and disjoint usage; request wire records prove stream versus plain delivery |
| Browser prompt and incremental acceptance | Open browser tab A; type and click Send prompt for a bounded explanation | Recorded real page text/screenshots, incremental observations, finalized answer card without duplication, accepted/completion IDs |
| Atomic late join and reconnect | Open tab B during generation; reload B during work and after completion | Original browser frames and page text converge to one final answer; active operation metadata/partial snapshot when timing permits; exact cuts separately proven locally |
| Tools and report cards | Ask browser Agent to write/read a small scratch file and run a short delayed command | Actual file bytes, accepted proposal/call/result cards, request wire call IDs, running-report versus later terminal job status, screenshot including connection state |
| Independent typing/speaking pause | While work runs, type correction in tab A; queue card speech in B, then finish/cancel B speech while A retains text | Pause acknowledgements/count changes, surviving typing cause, UI states, exact tool-boundary exclusion also locally tested |
| Hint receipt and later delivery | Submit A correction with Send hint during active work | Correlated received/sent=false ack, input clears/reconciles, later reconstructed provider request includes hint; record actual model behavior even if it ignores hint |
| Interrupt while paused | With A holding input pause, type `/interrupt` in combined PTY | Active request settles interrupted, unfinished tool calls pair safely, GUI outcome and terminal output agree |
| Existing job survival/supervision | Inspect delayed job after interrupt; deliberately ask to kill its known handle | Running-job report/final job fact and actual process cleanup; pause does not claim to suspend already admitted work |
| Post-interrupt reuse and disconnect release | Clear correction, submit short follow-up, then close a typing tab | Agent remains usable, surviving tab observes pause count release, browser disconnect does not cancel admitted prompt |
| Terminal EOF detach versus quit | Send terminal EOF after its admitted prompts settle, then another browser prompt; separate `/quit`/SIGTERM shutdown | Browser remains usable after EOF; application shutdown closes clients and process without retained job/worker |
| Opt-in actual speech | Click Auto speech on, ask/on-demand speak visible card, cancel midway, type/send correction while queued | Browser/platform version, synthesis start/end/error events, captured actual produced audio or explicit exact blocker. Mock queue tests do not satisfy this row |
| Safe accessible retained cards | In controlled actual browser fixtures, render malicious tool name/args/result, ANSI text, long answer; keyboard expand and Speak; scroll up while tokens arrive, then Return to latest | No script/navigation/image fetch; full retained text, omitted count, expanded state, focus remains input, user scroll preserved; browser screenshot/text |
| Public two-Agent reuse | Launch independent `examples/browser-consumer` with two Agents, custom layout reusing Page/Connector/ArtifactScroll; submit distinct browser prompts and independent pauses | Two distinct histories/watch generations, same local part IDs cannot mix cards/speech, isolated aggregate counts and completed responses; public Go startup snapshot/pause receipts |

For speech, inspect real browser voice availability and exercise synthesis through
actual user clicks. Capture system audio only through an available authorized
local mechanism; do not relabel an onstart event as audible output. If synthesis
or audio capture is unavailable, preserve the observed versioned failure and
leave the required gate open. No workaround without an observed failure/test.

## Deterministic checks separate from paid evidence

Local fixtures prove the properties paid timing cannot reliably distinguish:
held `Hel`/`lo.` snapshot cuts before/after acceptance and tool-final boundaries;
active-operation snapshot before first delta; repeated local part IDs across
responses/Agents; exactly 100 events plus omission and earlier-call result;
owned-copy mutation and no config/credential projection; independent watch
count/byte and socket saturation with another healthy viewer/completion;
recipient selected then disconnected before enqueue; idle subscriber cleanup;
pause acknowledgement excluding later tool admission; interruption/close while
held; valid error/refusal recovery and exact message/ID limits; trace received /
queued / written accuracy and trace write failure; recursive typed opaque
projection retaining empty text, order, IDs, false flags and user argument keys;
partial snapshot abandonment and prompt acceptance unknown without resend;
speech A/B cancel, C then stale A callbacks; replay silence and final cursor
continuity. Browser fixtures must render payload positives before negative
assertions. Headless core builds with optional module removed. Earlier chapter
checks run on the committed candidate and every affected Go module gets format,
vet and tests before a source freeze.
