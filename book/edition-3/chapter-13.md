# Chapter 13 — A listener who can only hear

Chapter 12 gave the Agent tools for looking at and operating its browser. This chapter asks a different question: what reaches a user who relies on the application's speech? A correct screen is not an answer. The speech channel has its own ordering, filtering, interruption and failure behavior.

The first-edition investigation used a restricted listener to expose gaps in that channel. We retain its useful discoveries below, but build the current endpoint in C# 14/.NET 10. The listener is an instrument, not a substitute for a person listening to the result.

## TL;DR: what you build

1. **One speech pipeline.** Thinking and response text enter a per-part raw buffer. Resolve Markdown fences before splitting phrases, strip presentation syntax, expand identifiers, preserve paragraph boundaries, and explicitly flush unfinished tails. Successful tool bodies stay automatically silent; dispatch names and failures are spoken separately. Whole finals without preceding deltas must enter the same path exactly once.
2. **One admission log.** Record each utterance when it reaches the front of the speech queue, before choosing native playback or silent transcript mode. Include source labels. Record typing/speech pause transitions too. Admission is not proof of sound or completed playback.
3. **A genuinely restricted consumer.** `Examples/Listener` constructs an Agent with exactly five custom tools: `say`, `listen`, `sleep`, `set_speech_speed`, and `save_history`. It installs no builtins or Skills manager. Guessed file, browser-inspection and evaluation tools must fail ordinary dispatch.
4. **Complete lifetimes.** The listener owns a fresh target GUI, browser profile and private browser connection. Its Session joins Agent Jobs before closing that connection and stopping the owned processes. EOF, interruption, deadline and failed startup must not strand a pending sleep or browser.
5. **Layered verification.** Check the formatter, actual page, executable speech harness, restricted listener and cleanup separately. None establishes real-provider compatibility or audible accessibility by itself.

The existing C# solution remains one core assembly. `Examples/Gui` and `Examples/Listener` are separate applications referring to it. The GUI owns speech presentation; the Agent still owns scheduling and durable conversation.

## 13.1 Build the speech channel before judging it

`Examples/Gui/wwwroot/speech.js` is a small pure formatter. `speechPhrases(raw, offset, flush)` returns completed phrases and a new raw-source offset. The page retains that offset on the displayed part. It does not replace displayed text with a spoken copy or run independent regular expressions on each delta.

A delta is not a sentence. It may stop inside a word, Markdown link, identifier or fence marker. Keep the raw text until there is enough structure to release a phrase. At ordinary sentence punctuation, retain the necessary lookahead rather than guessing what the next delta contains. At an explicit flush, release the final tail even without punctuation.

Fences come first. Their delimiters and bodies span lines, so removing them after splitting on newlines is too late. Recognize a complete fence as one source span and replace it with “code block”; a partial fence remains buffered until completion or final flush. Remove link destinations while preserving the label, strip inline backticks and heading/emphasis/bullet markers, and turn a lone newline into whitespace. A blank line remains a paragraph boundary. These transformations describe the supported prose subset, not a complete Markdown parser.

Identifier expansion is similarly modest:

| Source | Spoken words |
| --- | --- |
| `queueChunk` | queue Chunk |
| `max_tool_rounds` | max tool rounds |
| `HTTPServer` | HTTP Server |
| `ALL CAPS` | ALL CAPS |
| `Hello ordinary World.` | Hello ordinary World. |

Do not confuse removing code notation with removing its content. Inline identifiers can carry meaning; a fenced implementation is usually better named than read token by token. Manual artifact reading still uses the same formatter.

### Release points and whole finals

`narrate(card, flush)` uses the part's raw cursor. Model deltas append text; a final replaces the displayed text with the authoritative complete value and flushes only its unconsumed suffix. If no delta arrived, the cursor is zero and the entire final is eligible. A duplicate final finds no remaining suffix.

Before a tool dispatch or error announcement, `flushNarration()` releases pending prose tails. Tool calls announce the tool name, not its arguments. Successful tool-result bodies stay silent unless the person explicitly asks to read the artifact. Failed tools enter the labeled error path. Provider failures arrive as durable `error_occurred` lifecycle frames; control refusals arrive as `error` frames. Both must speak, while ordinary lifecycle notices and terminal turn completion must not duplicate those errors.

Replay is silent. A restored or muted partial part remains suppressed through its final. Starting midway through a restored fence would otherwise expose code whose opening delimiter was never observed by the speech cursor. New live parts can narrate normally.

## 13.2 Admission, playback and the two pause causes

The page owns one current native utterance and a FIFO of future phrases. Queued speech already counts as busy. Only when `startSpeech()` removes the next phrase does it append an utterance to `window._ttsLog`. This is the last shared point before native playback and `?speech=transcript` diverge.

Each entry has increasing `seq`, millisecond timestamp `ms`, and `kind`. Utterances include `text`, `source`, and `part_id`; pause and resume entries include `cause`. Sources distinguish `thinking`, response `text`, `tool-announcement`, `error`, and deliberately requested speech. Thinking and errors also receive spoken prefixes so the distinction is available to a native listener, not merely a JSON reader.

In transcript mode the same queue admits and records phrases but makes no native synthesis call. In ordinary mode the current acknowledged rate and selected voice are copied onto each new `SpeechSynthesisUtterance`. Completion and error callbacks advance the same FIFO. Stale callbacks cannot finish a newer utterance. Cancelling clears current and future speech; it does not append all cancelled future paragraphs as if they had been heard.

Typing and speech are separate causes of one transition-based pause decision:

```javascript
const speaking = activeSpeech !== null || speechQueue.length > 0;
const cause = humanTyping ||
    (settings?.pause_while_speaking !== false && speaking);
```

Human whitespace is still a draft. Programmatic input is not human typing, even when a later speech callback recomputes the cause. Send pause/unpause only when the combined cause changes. This avoids floods of pause commands and respects another tab's explicit override rather than continuously reasserting a private veto.

Escape cancels speech even when the prompt contains a draft. It leaves that draft intact, so the typing cause can continue to hold the Agent paused. Stopping the voice and releasing the Agent are different operations.

## 13.3 A user who can only listen

A prompt that says “do not inspect the DOM” is not a capability boundary. The separate C# listener creates an empty library Agent, then constructs only these custom Tools:

| Tool | Required arguments | Result |
| --- | --- | --- |
| `say` | `text` | Submits one target turn through the actual form; refuses while a previous turn is pending |
| `listen` | `timeout_seconds` | Returns newly admitted utterances with sources and explicit `pending`/status fields |
| `sleep` | `seconds` | Waits the nonnegative finite duration, cancellably |
| `set_speech_speed` | `rate` | Updates the target and waits for acknowledgement; valid range is 0.25–4 |
| `save_history` | `path` | Writes typed submissions and heard entries as semantic JSONL |

`Common/Session.cs` declares Session state, resource handles and operations. `Channel/Session.cs` implements the five tools, one semantic history and its independent hearing cursor. `Browser/Protocol.cs` owns correlated DevTools commands; `Browser/Lifetime.cs` launches and closes the resources. `Program.cs` composes the session and owns its overall deadline. All partial Session sections compile into the same application assembly; they are not separate services connected by callback bags.

The Session's private browser connection sets the actual prompt, observes the explicit `turn_ended` boundary, reads admitted `_ttsLog` utterances, and acknowledges rate changes. It exposes no selector, DOM snapshot, screenshot, arbitrary evaluation, target response text or raw tool-result body to the model. Chrome receives no provider credential. The target and listener have separately supplied provider configurations.

Before submitting, `say` marks the target pending. Only its real completion frame clears that flag. A momentary idle display before acknowledgement therefore cannot masquerade as a completed turn. `listen` returns promptly for new speech; a timeout with `pending:true` is not idle. Subsequent calls advance the hearing cursor, while `save_history` does not consume it. Saved entries have only `seq`, `kind`, `source`, and `text`: typed prompts and source-labeled heard speech, not the listener Agent's reasoning or ordinary tool transcript.

Here, “heard” means admitted to the speech channel. It does not mean a microphone captured sound. Keyboard navigation and general screen-reader usability remain separate concerns.

### Why these tools wait for completion

Ordinary custom tools use Jobs and may return a timed running report. This listener deliberately has no `wait_for_job`. Returning a running handle for `listen` could consume its hearing cursor later without delivering those entries to the model. Returning a running handle for a long `sleep` would likewise leave no supported observation operation.

The custom Tool constructor therefore has an optional `waitForCompletion` parameter, defaulting to false. The listener opts in. Dispatch still reserves and starts a real Job, but awaits that worker before ordinary observation. The existing structured-result path returns the complete value once, without the byte cap truncating a semantic listen result. There is no second result store or scheduler. Cancellation detaches a waiting observer; Job ownership and shutdown still govern the actual operation. Existing coding tools keep their previous behavior.

### Closing the session

The listener uses a new or empty caller-selected directory, a fresh browser profile, and owned loopback ports. It never attaches to the human's page or running Ensemble. Task mode finishes its reliable Agent turn; stdin mode treats EOF as session shutdown. SIGINT, SIGTERM, deadline and startup failure all enter the same explicit cleanup.

First shut down the listener Agent and join its Jobs, including pending sleeps. Then destroy the core root, dispose the browser connection, stop the owned browser, and stop the owned GUI. A short graceful process wait has a kill-tree fallback. Output streams are drained into the session's logs. Long sleeps are sliced only for the platform's delay range, not silently capped by product policy.

An earlier Go/Node listener closed its browser on EOF but retained a sleep timer. Its idle exit took 125 ms, while EOF during a four-second sleep took 3865 ms. The corrected historical peer exited in 127 ms with the sleep pending. Those measurements are lessons about ownership, not C# performance claims. The current endpoint has its own process-tree cleanup tests.

## 13.4 The run that succeeded for the wrong reason

The first-edition investigation gave the blind persona a task expected to fail.

Tool results do not enter the automatic speech channel. Only dispatch is announced, so a listener hears that a command is running and then hears nothing about what it did. The task was to run a command against a path that does not exist and report the exit code, and the prediction, written down before the run, was that the listener would come back empty.

It came back with the correct exit code.

Its retained account prints the explanation: six transcript entries, in this order. These are historical entries, not output produced by the corrected pipeline:

```
1. "run command"
2. "The command failed with exit code 1:"
3. "`"
4. "ls: /nonexistent-path-xyz: No such file or directory"
5. "`"
6. "This is expected since..."
```

Utterance 1 is the dispatch announcement. Utterances 2 through 6 are the model's own prose, describing what happened. No tool result appears anywhere in the transcript. The listener learned the exit code because the model chose to mention it.

The observer reported its transcript accurately and drew the wrong conclusion from it, writing that no gap had been found. That conclusion was reasonable given what it could perceive, which is the whole problem. Without a source label, the words alone cannot show whether the exit code came from tool-result delivery or model narration. A listener receiving the right information cannot tell whether the system delivered it or the model happened to be chatty that turn.

Accessibility resting on a model's prose habits is discoverability by luck. A terser response, a different system prompt, a model tuned to skip the summary, and the same task can yield silence with no warning and no error.

The scope ruling made the intended channel explicit. Tool results are not spoken automatically, by design: "I listen to your thinking, and that is enough." Thinking and response text both feed the channel; the exit code arriving through prose is permitted behavior, not proof that raw tool output was spoken. The GUI still lets a person deliberately read an artifact through its speaker button.

That ruling also redirected the instrument. A blind persona that hunts for unspoken tool results is testing a decision rather than a defect. The transcript checks whether thinking and response text arrive completely and in order, with announcements and errors labeled separately. A person must still judge intelligibility.

Utterances 3 and 5 expose another defect: punctuation leaked into the channel.

## 13.5 A bug found by ear

Two of the historical entries were a single backtick. Punctuation had reached the synthesizer input instead of being removed by the speech filter.

The obvious explanation is streaming. Deltas arrive on arbitrary boundaries, a fenced code block gets split across two chunks, and each half is filtered separately, so neither half contains a complete fence and the markers survive. That explanation is clean, mechanical, and wrong.

A test disproved it. Feeding the entire fenced block as one chunk, with no split anywhere near it, still produced spoken backticks.

The real cause sat one layer earlier. Filtering ran per phrase, and phrases were produced by splitting the buffer on newlines. A fenced block contains newlines by construction, so by the time the filter saw anything, the fence had already been cut into separate lines. A pattern that needs an opening marker and a closing marker to match will never match a line containing exactly one of them. The filter examined three fragments, found no fences, and passed all three through.

The fix resolves fences against the whole buffer before any splitting happens. The filter then sees a complete block and replaces it with a name, and the listener hears "code block" where a wall of syntax used to be.

The account also records the listener's objection to splitting on newlines. Prose wraps. A sentence broken across two source lines is one sentence, and splitting at the wrap produces two utterances with an unnatural pause between them. Speech engines work a phrase at a time, and every boundary the pipeline invents is a pause the listener hears.

So a lone newline became whitespace and a blank line stayed a boundary. A wrapped sentence is now spoken as one utterance, and paragraphs still separate.

A visual-content check does not establish correct speech segmentation. A DOM snapshot shows text that is present and correct. A screenshot shows a page that renders properly. A grader asserting on rendered content passes. The content is fine. The *segmentation* of the content is wrong, and segmentation determines where the voice can pause.

## 13.6 The bug nobody could hear

A complete part can be rendered without any preceding deltas. The historical handler displayed that text and called `flush()` on an empty buffer. It produced no speech at all.

The C# final handler feeds the authoritative text through the same raw-cursor path as deltas, so zero prior deltas is an ordinary case rather than a separate formatter. Test the actual page with a whole `part_final` and no earlier `part_delta`, then repeat the final and reconnect. Require one utterance, no duplicate and silent replay.

Disabling provider streaming is useful but insufficient evidence for this branch: the Engine may synthesize display deltas from a plain response. A final-only page check and a nonstream executable check answer different questions. Keep both.

## 13.7 Using the instrument carefully

The historical investigation's most useful surprise was not another speech filter. The listener received an exit code through the model's narration and concluded the channel had no gap. That is why source labels matter. Correct words can still support an incorrect causal conclusion.

Tool bodies are intentionally not automatic speech. A person can deliberately read an artifact with its speaker button; the model may also explain the result in its own prose. A listener who hears that explanation has not proved raw tool-result delivery. Conversely, absence of a secret string proves nothing if the tool never read the planted file successfully. Correlate the successful tool result before asserting that its contents stayed silent.

Likewise, a transcript demonstrates text admission, not voice quality. Fake native callbacks can verify queue and pause wiring. Neither establishes pronunciation, audible pacing, interruption comfort or accessibility for a particular person. Those remain human exercises with authorized native audio.

## 13.8 Exercise: build and verify the C# endpoint

### Source map and builds

Use the approved .NET 10 SDK and existing package cache. From the solution directory:

```sh
"$DOTNET" build Ensemble.csproj
"$DOTNET" build Examples/Gui/Gui.csproj
"$DOTNET" build Examples/Listener/Listener.csproj
"$DOTNET" run --project Tests/Tests.csproj
node Tests/SpeechTests.mjs
```

The portable formatter test fragments input at every split and also one character at a time. It checks whole and fragmented fences, wrapped prose, paragraph boundaries, identifiers and final tails. It does not launch a browser or claim sound. The listener is a separate assembly consuming the public core API, not code linked into the core to obtain private access.

### Executable speech grading

The public `scripts/tts-harness.sh` entry point launches the real GUI, submits the prompt through the actual form, and writes the page's admitted-speech JSONL. It runs in transcript mode. It does not reconstruct the provider answer or replace the formatter under test.

Supply existing paths through `TTS_NODE`, `TTS_DOTNET`, `TTS_GUI`, `TTS_CHROME`, and `TTS_PUPPETEER`; see `scripts/README.md`. No installation is performed. The caller supplies a fresh `WORKSPACE`, `TTS_LOG`, its loopback `LLM_BASE_URL`, and `STREAMING=on|off`. `--describe` reports the real `read_file`/`path` contract without launching. `--type-during-turn TEXT` generates a human input event during a response and later clears it.

From the repository root, with those explicit paths set:

```sh
go run ./cmd/grade-edition2 -ch 13 -executable-only -json solutions/edition-3/main
```

The command name is retained from the shared grader. This scoped adapter runs the nine Chapter 13 checks, including its zero-point launch check, over five fresh GUI launches. It excludes inherited Go/Chapter 12 grading rather than pretending C# satisfies that topology.

| Check | Points | Observation |
| --- | ---: | --- |
| `tts-discrimination` | 20 | Successful file results stay automatically silent; distinct model answers are spoken |
| `tts-filters-markup` | 20 | Markup is removed and a fence is named |
| `tts-buffers-fragments` | 10 | Words survive fragment boundaries |
| `tts-speaks-unstreamed` | 15 | Plain provider output is spoken once |
| `tts-gate-both-causes` | 10 | Typing pause and subsequent release are ordered |
| `tts-speaks-errors` | 10 | Actual provider failure reaches speech |
| `tts-boundaries` | 5 | Wrap, paragraph and trailing-tail boundaries remain meaningful |
| `tts-expands-identifiers` | 5 | Identifiers expand while ordinary capitalization survives |

The gate's name is broader than its numeric observation. It does not by itself prove both overlap orders, human whitespace, programmatic-input discrimination, final-only delivery, listener restriction or lifetime cleanup. Those need their focused checks. The prior chapter's unchanged evidence remains valid within its original scope; it is not newly earned through this score.

### Restricted-listener exercise

Configure `LLM_VENDOR`, `LLM_MODEL`, `LLM_BASE_URL`, and `LLM_API_KEY` for the target. Configure the four corresponding `LISTENER_*` values separately. The listener has no credential fallback. For offline verification, use working loopback provider fakes with dummy keys, not paid services.

From the solution directory, with a new session path and an existing browser:

```sh
"$DOTNET" Examples/Listener/bin/Debug/net10.0/Listener.dll \
  --dotnet "$DOTNET" \
  --gui "$PWD/Examples/Gui/bin/Debug/net10.0/Gui.dll" \
  --browser "$CHROME" \
  --directory "$SESSION_DIR" \
  --timeout 120 \
  --task 'Ask the target to explain HTTPServer, queueChunk and max_tool_rounds. Listen until pending is false, report the admitted words and sources, and save semantic history to speech.jsonl.'
```

Inspect the listener's actual model declarations and deliberately attempt an undeclared tool. Check pending versus idle, repeated listens, source labels, rate acknowledgement, a sleep longer than the ordinary observation delay, and a saved history containing only typed/heard entries. End separate sessions through EOF, SIGINT, SIGTERM, deadline and failed startup while a real sleep is pending; verify their owned descendants exit. Never use another person's browser/profile to obtain that evidence.

## 13.9 What this endpoint establishes

The C# run has 71 passing component groups, a separate restricted listener, focused real-page exercises with working socket/native-speech fakes, real owned GUI/browser/listener exchanges against loopback provider fakes, and the scoped executable speech grade of 95/95. Process-lifetime checks cover pending-sleep cancellation and owned cleanup. The first grade was 65/95: its model fixture accidentally disabled streaming, and it exposed a real missing narration path for provider lifecycle errors. Correcting both and adding the lifecycle regression preceded the final passing grade.

The historical first-edition and later Go listener sessions are not C# integration evidence. No real provider, native audible quality, human listening or full accessibility claim follows from this offline boundary. Turn on ordinary audio only with appropriate authority and listen: whether the voice and phrasing are usable remains a human question.
