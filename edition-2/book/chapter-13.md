# Chapter 13: The Channel Nobody Tested

This chapter is for me. Accessibility is an afterthought everywhere, never in the critical path of creating a product. Since I am in control here, your AI coding agent gets decent a11y from the start, if you follow this codebook accurately.

The test has to run over data as close as possible to what a blind or low-vision coder actually receives. So in this chapter we build a virtual blind coder, one that can only hear TTS output, and it has to drive the AI coding agent successfully.

If you care about SWEs with low vision, please do not skip this chapter.

An automated observer can complete a task without depending on the channel under test. The historical GUI experiment in Chapter 12 found real interaction defects, alongside findings supplied by human observation and source inspection. Its DOM-based work did not establish that speech carried the same useful information. The screen could look right while the listener missed the message.

The first-edition speech investigation exposed that gap. This chapter preserves its findings and builds the instrument that was missing: an observer whose only account of the target agent comes through the speech transcript.

## TL;DR

Two deliverables. A speech pipeline that turns a stream of arbitrary text fragments into utterances a person can listen to, and a small blind virtual-user client that perceives the target through its speech transcript alone. Chapter 12's historical driver is not a prerequisite project.

### The speech pipeline contract

Text arrives as arbitrary fragments. The GUI receives decoded text deltas, frequently mid-word, and the pipeline owns the job of turning that into speech.

1. **Buffer until a phrase boundary.** Never speak a fragment as it arrives. A word split across three chunks is one word.
2. **A sentence-ending period is a boundary. A single newline is not.** Prose wraps, and splitting at the wrap breaks the sentence. Treat a lone newline as a space and a blank line as a boundary.
3. **Provide an explicit `flush()`.** End of stream, a tool announcement, and an error each force the buffer to be spoken. Without this, a response ending in a colon is never heard.
4. **Filter for the ear before speaking.** Emphasis markers, inline code markers, headings, bullets, and link syntax are removed. A fenced code block is named rather than read.
5. **Resolve fenced blocks against the whole buffer, before splitting.** A fence broken into separate lines can never match a fence pattern, and the orphaned markers get spoken aloud.
6. **Expand identifiers.** `camelCase`, `snake_case`, and `HTTPServer` are read as separate words. Ordinary words are left alone.
7. **Speak parts that arrive whole.** A final delivered to the renderer without preceding deltas is still spoken, and a part that streamed is not spoken twice. Provider nonstream mode alone does not prove that renderer path was exercised.
8. **Speak errors.** An error that is displayed and not spoken does not exist for a listener.
9. **Record a transcript.** Every filtered utterance entering the channel is recorded with its text and source before audio bypass or the Read aloud setting can suppress sound. Preserve text, thinking, tool-announcement and error origins. Tool results stay automatically silent.

### The pause gate

The browser derives its automatic tool-dispatch pause from queued speech and actual human typing:

```
local_pause_cause = speaking OR human_typing
```

10. **One derived predicate, not three call sites.** Compute the value, compare it against the previous value, and send only on an edge.
11. **A human-entered space counts as typing.** Keep Chapter 12's event provenance; programmatic input cannot become a human cause during a later speech callback.
12. **Automatic release requires both local causes clear.** Explicit unpause from any tab still releases the Agent for everyone. Escape cancels speech regardless of input contents; human typing may still hold the local cause.

### The blind persona

13. **Deny the DOM.** The listener Agent has exactly `hear`, `type`, `submit`, `wait`, and `sleep`. No snapshot, click, file, shell or skill-lifecycle tools. Enforce the same boundary at dispatch, including guessed names.
14. **Bypass audio without bypassing processing.** The blind peer opens the real GUI in transcript mode. It reads the same filtered entries that would enter the speech queue, without waiting for native synthesis.
15. **Drive the actual client.** Typing and submission reach the real target; waiting observes its completion; hearing advances a transcript cursor. End the peer's whole lifetime on EOF, including a pending sleep. Keep ordinary sleep working while the session is alive.

### Yours

Which markdown constructs to filter beyond the required set. What to call a fenced block when you skip it. Whether the transcript is a ring buffer and how large. The wording of a tool announcement. Whether an error preempts the queue or joins the back of it.

### Exercise

```
go run ./cmd/grade-edition2 -ch 13
```

## 13.1 The idea in plain words

A fire alarm inspector who checks the panel will never discover that a speaker in the east stairwell is disconnected. The wiring diagram is correct, the current draw is nominal, the panel reports green. Finding that fault requires somebody to stand in the stairwell during a test and notice the silence.

A DOM observer has a similar blind spot. It can read the text it needs from `gui_snapshot`, complete a task, and report success without checking what reached speech. The earlier GUI work established useful visual behavior. It supplied no equivalent assurance about missing utterances or awkward phrase boundaries.

The instinct at this point is to give the observer better reporting. Add a tool that returns the speech queue, describe it well, and ask the agent to check it. That helps only if the returned data and the task make the missing behavior observable. Chapter 12's confident phantom-panel report shows the danger of asking for a conclusion without supplying the state needed to support it.

The alternative is to remove the channel the observer has been leaning on. Take away the DOM, leave speech, and give it a task it cannot finish without listening. Now a defect in the speech channel is not a line in a report. It is a task that fails.

The principle generalizes beyond accessibility: test over data as close as possible to what the user actually receives. A DOM snapshot is not what a listener receives. It is a different signal, richer in some ways and poorer in others, and an observer consuming it can miss a failure that matters to someone listening.

## 13.2 The channel that reported on itself

This investigation started with a suspicion rather than a bug report. Bill, who listens to this agent for hours a day and had no way to audit the channel he was listening to, put it this way:

> I'm worried the TTS feedback, available via the MCP tunnel, isn't evaluated by any agent. It needs work.

The first-edition account records that concern before describing the investigation. Its findings explain why the listener is useful, even though this edition already carries forward several repairs.

In that implementation, four event paths met in the artifact renderer. Streaming deltas went in as they arrived, for both text and thinking. Tool dispatch announced itself. A part arriving complete flushed the buffer. Errors rendered to the screen.

Three of those four had defects. The fourth, error rendering, had no speech call at all: it built a div, set its text, appended it, scrolled, and returned. For a user who works by listening, the one message class that most needs to interrupt was the only class that made no sound.

The agent did have a tool for inspecting speech state. It was called `tts_queue`, and it read a variable named `window._ttsQueue`. The reported source inspection found reads of that name but no writer. The tool always fell through to its backup path, which returned either an empty array or a single synthetic entry whose text was the string `(speaking)`.

That tool was declared `ephemeral: "round"`, the MCP mechanism that injects a tool's output into every round automatically. So the agent had been receiving speech telemetry continuously, and the telemetry had been contentless the entire time.

The older GUI-skill grader checked that this injection worked. It asserts that `tts_queue` output reaches the request context, which is a claim about the engine rather than about speech. A pipe that carries nothing is still a pipe. The check was green while the channel behind it was dead, which is a useful thing to know about green checks.

The historical pause gate had a related omission. The design specified:

```
paused = tts_speaking OR user_typing
```

Only the typing half was ever wired. Every pause and unpause message in the client came from a keystroke handler, and the client contained no reference to speech state at all. The Go side was correct, checking the gate before every tool dispatch. One of the mechanism's two inputs had simply never been connected.

The speech module advertised pause integration in its opening comment:

```javascript
// TTS — Text-to-speech for artifacts with Chrome wake-up and pause integration.
```

The historical source inspection found that comment and no other occurrence of `pause` in the file.

The repair makes the cause explicit in one place. Chapter 12's accepted browser already does this, including the distinction between a person composing text and MCP filling the same field:

```javascript
function pace() {
  const next = speaking || humanTyping;
  if (next !== localPaused) {
    localPaused = next;
    send({ type: next ? "pause" : "unpause" });
  }
}
```

Speech starting, speech ending, input and successful submission all update their cause and call this function. Comparing against the previous value makes automatic messages edge-triggered: more keystrokes do not create more pause messages while the cause remains true. An actual human space still counts as composing. Programmatic field contents remain programmatic when speech ends.

This is a local decision about when to send a command. The Agent has one pause gate, and an explicit unpause in another tab releases it. An unchanged cause in the first tab cannot immediately reinstate a private veto. The transition rule preserves both the listening pause and that deliberate override.

Escape now cancels speech even with text in the input field:

```javascript
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape") speech.cancel();
});
```

Cancellation empties the speech queue and unfinished speech buffers. It does not erase the person's typed text. The cause calculation then decides whether typing still requests a pause. Stopping the voice and releasing the Agent are different actions when the person is still composing.

## 13.3 A user who can only listen

Removal is the operative word. A prompt asking the listener to avoid the DOM leaves the capability available when the task becomes difficult. Build its Agent with only the five required tools and reject unknown operations at the peer as well. Inspect the consuming Agent's declarations: a narrow peer attached to an ordinary skill-enabled coding Agent would still leave file and lifecycle capabilities beside it.

The accepted implementation uses a small optional application, `gui/cmd/virtual-user`, through the public Ensemble API. It creates the Ensemble root and an Agent without builtins or skills, then attaches a raw MCP transport to the browser peer. No second actor loop or core API is needed. The separate GUI module continues to own browser automation; the core remains independent of it.

| Tool | Arguments | Effect |
| --- | --- | --- |
| `hear` | `{}` | Returns new transcript entries with their source, advancing the cursor |
| `type` | `{"text":"..."}` | Replaces the actual chat input |
| `submit` | `{}` | Submits that input and records which turn to await |
| `wait` | `{}` | Waits for that submitted turn to end, then returns new transcript entries |
| `sleep` | `{"milliseconds":150}` | Waits the requested nonnegative finite duration |

The peer privately operates the real page, but exposes no arbitrary selector, evaluation, click or snapshot operation to the model. Its process receives only the environment needed to launch the browser, without the model credential. The listener must obtain target information from returned speech entries rather than another view of the same answer.

Speech in this mode produces no audio. Opening the page with `?speech=transcript` uses the real renderer and filters, records each utterance, and skips native synthesis. A long reasoning trace can therefore be inspected without waiting for every word to be spoken. This tests the text channel a synthesizer would receive; it does not make the observer capable of judging sound.

Keep navigation outside this exercise. The application owns its running commentary of thinking and response text; keyboard and screen-reader navigation remain separate concerns. One restricted listener is enough to test this channel. That scope decision does not claim that a transcript covers the application's entire accessibility surface or that all low-vision and blind users need the same interface.

### Closing the listener

A small client can still leak a lifetime. The student's first peer closed its browser on stdin EOF but left a pending Node sleep timer alive. The Go client closed the transport, then waited for the peer before returning. Its model turn could finish while process cleanup continued waiting for a timer that no longer served a session.

The independent probe made the difference visible: idle EOF exited in 125 ms; EOF during a 4000 ms sleep took 3865 ms. The peer did eventually exit, so calling this a permanent hang would misdescribe the measurement. A longer requested sleep would make the same ownership error more expensive.

The correction shared the existing close-browser-and-exit handler between EOF and SIGTERM. On the corrected peer, the same pending-sleep probe exited in 127 ms; its idle control took 126 ms. Normal sleep still works while the session remains open. A separate real-client check covers a pending sleep whose job has returned an ordinary running callback before the model finishes. Ending the owned process ends its timers; no arbitrary sleep ceiling or new cancellation framework is needed. The [paired receipts](../docs/edition-2-notes/ch14-code-review.md#correction-round-1-accepted-with-inherited-grader-exception) preserve both versions.

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

Listening is a better instrument than looking, for this channel. It is still not sufficient.

A part that arrives complete, with no deltas preceding it, was rendered to the screen and never queued for speech. The handler set the element's content and called `flush()`, which empties a buffer that in this path is already empty.

In the historical implementation, disabling provider streaming reached that path. In this implementation, the Engine can synthesize display deltas even for a plain provider response. Switching off HTTP streaming does not prove that the renderer ever received a final without deltas.

The student's first mutation check demonstrated the trap: removing whole-final speech still passed its nonstream test. The fix was in the test boundary. Preserve actual Agent finals but drop transient deltas at the external WebSocket boundary, then observe the real renderer and transcript. That reaches the missing branch; restoring the defect now loses the utterance. A separate check catches duplicate final speech.

A listener who expects a reply can notice silence. Silence alone cannot tell them which text was omitted or whether the model produced any. Compare the emitted final with the speech transcript as well as inspecting the feed sites; the expectation must come from a boundary outside the failing channel.

The naive repair introduces a worse bug. Queue the final part unconditionally and every streamed response gets spoken twice, once from its deltas and once from its final. The pipeline needs to know whether a given part already streamed.

The renderer already knows whether a card received deltas. The accepted implementation marks `card.streamed` on a delta and uses that fact when a final arrives:

```javascript
if (!message.replay && ["text", "thinking"].includes(kind))
  speech.chunk(id, card.streamed ? "" : card.text, true, kind);
```

An empty final chunk still flushes the saved tail. A final without prior deltas supplies its complete text, and history replay remains silent. The speech buffer holds unfinished text by part identity, with a back-pointer to its owning Speech queue. It cannot join the tail of a thinking part to a response merely because they arrived nearby.

## 13.7 Why a daily listener was not enough

The first-edition account describes a daily listener who depended on the channel but raised a suspicion rather than a list of its defects. That is a reason to build a better instrument, not a reason to doubt the listener.

Missing errors and missing finals produce no corresponding utterance to inspect. A pause input that never engages may show up only as a tool starting too soon. Empty telemetry needs an independent check that the queue was actually empty. Audible defects can be difficult to attribute too: an odd pause may come from segmentation, network timing or the voice itself. The retained account does not measure how quickly people adapt to these defects, and the chapter need not invent that explanation to show why a second observation point helps.

Preserve the distinction between guidance and evidence during a supervised run. Supplying the task, identifying a stale binary or freeing a port makes the experiment possible. Supplying a missing selector or the expected answer changes what success demonstrates. Chapter 12's historical corrections are a reminder to record human and source-inspection assistance rather than retroactively calling every finding autonomous.

## 13.8 What a program cannot verify

A grader for this chapter can check a great deal. Whether a word survives being split across three fragments. Whether a wrapped sentence produces one utterance. Whether any markup marker reaches the channel. Whether a part that arrived whole was spoken, and whether a part that streamed was spoken twice. Whether the pause gate releases only when both of its causes are clear.

Those observations need different boundaries. Filtered text and its source belong in the transcript. Queue progress and pause transitions belong to the real application around it. The executable grading harness drives that application with an external provider and speech synthesizer fake; it does not discover and load arbitrary JavaScript modules or replace the Agent and renderer.

Native browser callbacks supply separate evidence that the actual synthesizer started and ended utterances. Chapter 12's cause-based pause controls still apply with native speech here, including human whitespace, programmatic input, explicit other-tab unpause and actual subsequent submission. The extra control tab closes before the single-tab speech sequence; this does not settle the earlier unexplained two-tab queue stall.

What those checks do not establish is how any of it sounds.

Pronunciation needs a listener. Whether `HTTP Server` is clearer than an unexpanded identifier is a judgment about ears. Prosody and comfort at a chosen rate vary with the listener and voice. A browser's accepted rate value is not a measured words-per-minute rate, and this chapter has no retained measurement supporting the original fixed browser-speed ceiling or dedicated-engine comparison.

The boundary deserves to be stated plainly in the chapter that builds the instrument. These checks verify that the intended text reaches the speech channel at the intended boundaries and that measured state transitions occur. A human decides whether the result is worth listening to.

## 13.9 Exercise, graded

Build both deliverables: the speech pipeline on the existing GUI path and the restricted public listener application. Keep filtering with the unfinished-text buffer and its actual Speech owner. Reuse the native queue, renderer identities, Agent and MCP transport. A separate transcript algorithm that produces the right fixture text while the GUI still speaks broken fragments does not satisfy the exercise.

Provide an executable `tts-harness.sh` at the core solution root, or `scripts/tts-harness.sh`. A normal invocation receives a prompt as its first argument, optionally followed by `--type-during-turn TEXT`. The grading process supplies:

| Input | Meaning |
| --- | --- |
| `LLM_BASE_URL` | The local external provider to exercise |
| `WORKSPACE` | The prepared workspace; file fixtures use absolute paths rather than assuming this changes the GUI process directory |
| `TTS_LOG` | Destination for observed JSON-lines speech and gate records |
| `STREAMING=off` | Request plain provider delivery; otherwise stream |

The supplied implementation's root script launches `gui/tts-harness.mjs`. It uses the real GUI, Agent, provider parser, renderer and speech pipeline in a browser. Only the external provider and asynchronous synthesizer are fakes. Install the GUI module's optional Node dependencies and a usable Chromium browser before running it. No paid provider is needed for grading.

The optional description invocation is:

```sh
./tts-harness.sh --describe
```

The supplied harness returns:

```json
{"read_file_tool":"read_file","read_file_arg":"path","supports_type_during_turn":true,"supports_streaming_off":true,"log_format":"jsonl"}
```

This describes public capabilities; it does not prescribe a JavaScript class or module filename. The actual prompt run writes JSON lines to `TTS_LOG`, with numeric `ms` and `seq`, `kind: "utterance"` and the actual filtered `text`, plus its `source`. Gate records use `kind: "pause"` or `kind: "resume"` with the observed `cause`, including `typing` where applicable. Record real events; do not construct the expected sequence independently of the running page. Finish and clean up owned children on every path. The grader bounds a harness run at 90 seconds and can terminate its process group.

From the repository root:

```sh
go run ./cmd/grade-edition2 -ch 13
```

The default is `solutions/ch13`; a positional argument selects your core directory and the existing `-gui` option selects its optional GUI module if necessary. The revised rubric preserves the actual original grader's weights, which differed from its printed chapter table:

| Check | Points | Observation |
| --- | ---: | --- |
| `tts-discrimination` | 20 | Successful file results stay automatically silent; distinct model answers are spoken |
| `tts-filters-markup` | 20 | Required markup is removed and a fence is named |
| `tts-buffers-fragments` | 10 | Words survive fragment boundaries |
| `tts-speaks-unstreamed` | 15 | Plain provider output is spoken exactly once |
| `tts-gate-both-causes` | 10 | Typing pause and subsequent release are observed in order |
| `tts-speaks-errors` | 10 | The actual error reaches speech |
| `tts-boundaries` | 5 | Wrap, paragraph and trailing-tail boundaries remain meaningful |
| `tts-expands-identifiers` | 5 | Identifiers expand; ordinary capitalization survives |
| Corrected Chapter 12 parity | 5 | Prior obligations remain intact |

The transcript grader does not independently establish every clause in the TL;DR. The final-only renderer branch needs its direct delivery check because a nonstream provider can still produce synthetic deltas. The numeric gate check observes ordered typing pause and release; those edges alone do not prove human whitespace or both orders of overlapping speech and typing. Keep the actual browser controls for those causes. The actual listener's five-tool boundary, dispatch refusal, hearing cursor, submission and pending-sleep shutdown need their public-client/browser evidence. Native audio and human listening remain separate again. Keep those obligations visible rather than claiming a numeric speech score proves them all.

The original student run scored **95/100 FAIL**: all new speech checks passed 95/95 and inherited parity failed 0/5 on the invalid earlier skill fixture. That build preceded the finished blind client and later cleanup correction. Its score is historical evidence, not a final-source pass.

A stronger check later exposed a gap inside that speech score. The old fixture requested `notes.txt`, but the harness launched the GUI in a different directory from its prepared workspace. The actual file read failed. Silence could not prove suppression of file contents that had never arrived. Using the planted file’s absolute path exercises the existing reader successfully; compare that correlated result with the transcript before claiming that its bytes stayed silent. An absence check needs something real to exclude.

The historical fence investigation ran the example whole and fragmented. Keep that distinction when checking your buffer: an intact fence that fails proves a different cause from a split token. Preserve line structure until filters that need it have run. An early grader review found the same mistake in bullets: flattening a single newline first let the second line's bullet escape an anchored filter. A normal two-item list is enough to expose it; no new markup language is required.

## 13.10 Taking it for a spin

Work in a scratch copy of your solution so the target save and listener history do not join an earlier experiment. From the core directory, install the optional driver dependencies and build both actual applications:

```sh
(cd gui && npm ci)
(cd gui && npx playwright install chromium)
(cd gui && go build -o /tmp/ensemble-ch13-gui ./cmd)
(cd gui && go build -o /tmp/ensemble-ch13-listener ./cmd/virtual-user)
```

The listener supports `--browser /path/to/chromium` if you already have a browser. Configure the target's provider through the existing environment. To keep the target focused on prose instead of loading tools, create `skills/listener-target/SKILL.md`:

```yaml
---
name: listener-target
type: primary
description: Brief answers for the speech listener exercise
---
Answer briefly in two sentences. No tools are needed.
```

Start the target GUI from the core directory, leaving its terminal open:

```sh
EN_PRIMARY_SKILL=listener-target /tmp/ensemble-ch13-gui \
  --skills-dir ./skills --gui-dir gui/web/gui --port 8088
```

In another terminal, configure the listener's own provider environment and run from the same core directory:

```sh
/tmp/ensemble-ch13-listener --agent-url http://localhost:8088 \
  --persona blind --peer gui/virtual-user.mjs \
  --task 'Ask the target to explain queueChunk, max_tool_rounds and HTTPServer in two short sentences using inline code markup, and to choose one fruit to mention. Then report exactly the spoken transcript, including its chosen fruit. Use only the five available tools.'
```

The peer opens a quiet transcript-mode tab itself. Close other talking tabs when isolating this channel. The supported persona here is `blind`; there is no required sighted mode to invoke for comparison. The listener writes its actual history to `virtual-user.jsonl` and reports what `hear` or `wait` returned. Keep that result beside the target's speech entries, because the model can add Markdown when quoting a transcript even though no Markdown entered target speech.

The retained October 10, 2026 sessions used `claude-haiku-5-5`, `gpt-4o-mini` and `gemini-3.8-flash`. They are recorded model identities, not current catalog recommendations. All three listener processes exited zero and used the restricted interface to type, submit and receive the target's real transcript. The [live receipts](../docs/edition-2-notes/ch14.md#student-implementation-and-initial-review-submission) retain their earlier-source provenance.

The filtered entries contain `queue Chunk`, `max tool rounds` and `HTTP Server`, with no inline backticks. Claude's listener received two response entries, but its target never chose the requested fruit; it said “fruit-sized” instead. The listener reported that omission. GPT-4o mini's listener heard the chosen apple in three response sentences rather than the requested two. Gemini's eleven entries include both thinking and response sources, followed by a report of the apple. These are useful channel results with different target-content limits, not three claims that every instruction was followed.

No human listened to judge these runs. Transcript mode bypassed native audio. The separate native-browser checks exercised real speech callbacks and pause, and the later peer-lifetime correction has its own deterministic evidence. Neither was a final-source paid repeat. After checking the words and their sources, turn ordinary audio back on and listen: the question of whether this voice and phrasing are usable for you remains yours to answer.
