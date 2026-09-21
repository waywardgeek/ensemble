# Chapter 14: The Channel Nobody Tested

This chapter is for me. The world generally supports accessibility as a kind of afterthought, never in the critical path of creating a product. Since I am in control here, your AI coding agent is going to have decent a11y from the start, if you follow this codebook accurately.

A key insight is that testing needs to be done over data as close as possible to what a blind or low vision coder experiences. So in this chapter we build a virtual blind coder, one that can only hear the output of TTS, and it has to drive the AI coding agent successfully.

If you care about SWEs with low vision, please do not skip this chapter.

An automated observer only finds bugs in the channel it needs to succeed. Chapter 13 built an observer that reads the DOM, and it found real defects: an XSS hole in a tool card, truncation nobody could reach, controls whose state lived only in a CSS class. Every one of those bugs was visible. The observer was looking, so the observer found them.

The same agent had a speech channel running the entire time. That channel was broken in six separate ways, and the observer reported nothing, because the observer never had to listen to finish its work. This chapter builds the observer that does.

## TL;DR

Two deliverables. A speech pipeline that turns a stream of arbitrary text fragments into utterances a person can listen to, and a blind persona for the Chapter 13 virtual user that perceives the agent through speech alone.

### The speech pipeline contract

Text arrives as arbitrary fragments. A streaming model emits deltas on byte boundaries, frequently mid-word, and the pipeline owns the job of turning that into speech.

1. **Buffer until a phrase boundary.** Never speak a fragment as it arrives. A word split across three chunks is one word.
2. **A sentence-ending period is a boundary. A single newline is not.** Prose wraps, and splitting at the wrap breaks the sentence. Treat a lone newline as a space and a blank line as a boundary.
3. **Provide an explicit `flush()`.** End of stream, a tool announcement, and an error each force the buffer to be spoken. Without this, a response ending in a colon is never heard.
4. **Filter for the ear before speaking.** Emphasis markers, inline code markers, headings, bullets, and link syntax are removed. A fenced code block is named rather than read.
5. **Resolve fenced blocks against the whole buffer, before splitting.** A fence broken into separate lines can never match a fence pattern, and the orphaned markers get spoken aloud.
6. **Expand identifiers.** `camelCase`, `snake_case`, and `HTTPServer` are read as separate words. Ordinary words are left alone.
7. **Speak parts that arrive whole.** A part delivered without deltas is still spoken, and a part that streamed is not spoken twice.
8. **Speak errors.** An error that is displayed and not spoken does not exist for a listener.
9. **Record a transcript.** Every utterance entering the channel is recorded with its text and its source, whether or not audio is produced.

### The pause gate

The agent pauses tool dispatch while the user is reading or typing:

```
paused = speaking OR input_non_empty
```

10. **One derived predicate, not three call sites.** Compute the value, compare it against the previous value, and send only on an edge.
11. **A space counts as input.** Trimming the field is wrong here.
12. **Unpausing requires both causes clear**, whichever one changed.

### The blind persona

13. **Deny the DOM.** The blind persona has no snapshot tool and no click tool. Remove them from the registry rather than discouraging them in a prompt.
14. **Bypass audio.** A test that waits for real speech runs in real time. Record to the transcript and return immediately.
15. **It keeps** the ability to hear, type, submit, wait, and sleep. Nothing else.

### Yours

Which markdown constructs to filter beyond the required set. What to call a fenced block when you skip it. Whether the transcript is a ring buffer and how large. The wording of a tool announcement. Whether an error preempts the queue or joins the back of it.

### Exercise

```
make grade14
```

## 14.1 The idea in plain words

A fire alarm inspector who checks the panel will never discover that a speaker in the east stairwell is disconnected. The wiring diagram is correct, the current draw is nominal, the panel reports green. Finding that fault requires somebody to stand in the stairwell during a test and notice the silence.

Automated observers have the same blind spot, for the same reason. Chapter 13's virtual user completed every task by calling `gui_snapshot` and reading the DOM that came back. Speech was never part of finishing the job. The speech module could have been deleted from the tree entirely and every one of those runs would still have passed, because nothing the observer needed came through that channel.

The instinct at this point is to give the observer better reporting. Add a tool that returns the speech queue, describe it well, and ask the agent to check it. That instinct is wrong, and Chapter 13 explains why: an agent asked to evaluate something it does not depend on will produce a confident answer in either direction. It has no way to be wrong that it can feel.

The alternative is to remove the channel the observer has been leaning on. Take away the DOM, leave speech, and give it a task it cannot finish without listening. Now a defect in the speech channel is not a line in a report. It is a task that fails.

This is the principle worth carrying out of the chapter, and it generalizes past accessibility: test over data as close as possible to what the user actually receives. A DOM snapshot is not what a listener receives. It is a different signal, richer in some ways and poorer in others, and an observer consuming it will faithfully report on a system no human is using.

## 14.2 The channel that reported on itself

This investigation started with a suspicion rather than a bug report. Bill, who listens to this agent for hours a day and had no way to audit the channel he was listening to, put it this way:

> I'm worried the TTS feedback, available via the MCP tunnel, isn't evaluated by any agent. It needs work.

That is a claim about the instrument rather than about the artifact, and it was exactly right. Before building the listener, it helps to see how right.

Four call sites fed it, all in the artifact renderer. Streaming deltas went in as they arrived, for both text and thinking. Tool dispatch announced itself. A part arriving complete flushed the buffer. Errors rendered to the screen.

Three of those four had defects. The fourth, error rendering, had no speech call at all: it built a div, set its text, appended it, scrolled, and returned. For a user who works by listening, the one message class that most needs to interrupt was the only class that made no sound.

The agent did have a tool for inspecting speech state. It was called `tts_queue`, and it read a variable named `window._ttsQueue`. A search of the source tree found exactly two references to that name, and both of them were reads inside the tool itself. Nothing in the codebase ever assigned it. The tool always fell through to its backup path, which returned either an empty array or a single synthetic entry whose text was the string `(speaking)`.

That tool was declared `ephemeral: "round"`, the Chapter 12 mechanism that injects a tool's output into every round automatically. So the agent had been receiving speech telemetry continuously, and the telemetry had been contentless the entire time.

Chapter 13's grader checks that this injection works, and it passes. It asserts that `tts_queue` output reaches the request context, which is a claim about the engine rather than about speech. A pipe that carries nothing is still a pipe. The check was green while the channel behind it was dead, which is a useful thing to know about green checks.

The pause gate has the same shape. The design document specifies it precisely:

```
paused = tts_speaking OR user_typing
```

Only the typing half was ever wired. Every pause and unpause message in the client came from a keystroke handler, and the client contained no reference to speech state at all. The Go side was correct, checking the gate before every tool dispatch. One of the mechanism's two inputs had simply never been connected.

The speech module had been advertising the missing feature since the day it was written. Line 1 reads:

```javascript
// TTS — Text-to-speech for artifacts with Chrome wake-up and pause integration.
```

Searching that file for the word `pause` returns that comment and nothing else.

The repair is smaller than the diagnosis, and it is worth showing because the original shape is the one most people write first. Three handlers each sent pause and unpause on their own, and each knew about only one of the two causes. A handler that noticed the input field had emptied sent unpause without any idea whether speech was still playing.

Deriving the value in one place removes the entire class of mistake:

```javascript
function updateGate() {
  const blocked = userTyping || TTS.speaking;
  if (blocked === gatePaused) return;
  gatePaused = blocked;
  // WebSocket readiness guard elided
  ws.send(JSON.stringify({type: blocked ? 'pause' : 'unpause'}));
}
```

Every event calls that same function. Speech starting, speech ending, a keystroke, a submission. It computes the predicate, compares the answer against the last value it sent, and sends only when the answer changed. Checking both causes before releasing the gate stops being a rule anyone has to remember, because both terms are sitting in the expression.

Two details earn their place. Comparing against `gatePaused` makes the messages edge-triggered, so forty keystrokes produce one pause rather than forty. And `userTyping` is set from the field's length rather than from a trimmed copy of it, because a user who has typed a single space is composing.

Wiring the missing half then costs one line, and its comment says what the line is for:

```javascript
// The speaking half of the gate. This subscription is the whole reason TTS state
// is exported: without it the agent runs tools while it is still talking.
TTS.onStateChange = updateGate;
```

One related ruling came out of the same session. The escape key now always cancels speech:

```javascript
if (e.key === 'Escape') {
  TTS.cancel();          // Escape always silences speech, whatever is typed.
```

It had been cancelling only when the input field was empty, which is what happens when cancellation is implemented as a side effect of clearing the box instead of as a command in its own right. A listener who wants the talking to stop wants it to stop.

## 14.3 A user who can only listen

The blind persona is the Chapter 13 driver with its eyes removed.

Removal is the operative word. A prompt instructing an agent to avoid a tool is a suggestion it will follow until the task gets hard. The persona deletes the DOM tools from the registry after the MCP bridge connects, so the capability does not exist:

```go
if *persona == "blind" {
    for _, name := range []string{"gui_snapshot", "gui_click"} {
        a.RemoveTool(name)
    }
    // logging elided
}
```

`RemoveTool` came from Chapter 13, where it served skill unloading. It works here unchanged, which is the payoff for having put tool removal in the registry rather than in the skill system.

What remains is hearing, typing, submitting, waiting, and sleeping. The persona can queue a prompt, send it, wait for the agent to go idle, and read back everything that entered the speech channel while it waited.

Speech in this mode does not produce audio. A test that waits for real utterances runs at the speed of talking, and a long reasoning trace is ten minutes of it. The pipeline records to the transcript and returns immediately, so a run finishes in seconds and the machine stays quiet. The transcript is the artifact under test, and the transcript is complete whether or not a speaker was involved.

One scope decision saves a week of work here, and it came from Bill. Hover-to-speak, arrow-key navigation, and element announcement belong to the operating system's screen reader. Chrome cooperates with JAWS, NVDA, and VoiceOver well enough that a developer who cannot see the screen already has a working way to move around a page, and building a second navigation model on top of that would duplicate the screen reader and do it worse. What this application owns is the self-speaking layer: the running commentary of thinking and response text that the agent produces while it works. A user turns that on, listens while the agent is talking, and navigates with their screen reader when it goes quiet.

That narrowing has a consequence worth stating. For this stack, a low-vision persona and a blind persona collapse into the same instrument, because the part under test is the same part. There is no second persona worth building, and the entire testable accessibility surface of the application is one channel carrying two kinds of text.

## 14.4 The run that succeeded for the wrong reason

The first task given to the blind persona was chosen to fail.

Tool results never reach the speech channel. Only dispatch is announced, so a listener hears that a command is running and then hears nothing about what it did. The task was to run a command against a path that does not exist and report the exit code, and the prediction, written down before the run, was that the listener would come back empty.

It came back with the correct exit code.

The transcript explains how. Six utterances entered the channel, in this order:

```
1. "run command"
2. "The command failed with exit code 1:"
3. "`"
4. "ls: /nonexistent-path-xyz: No such file or directory"
5. "`"
6. "This is expected since..."
```

Utterance 1 is the dispatch announcement. Utterances 2 through 6 are the model's own prose, describing what happened. No tool result appears anywhere in the transcript. The listener learned the exit code because the model chose to mention it.

The observer reported its transcript accurately and drew the wrong conclusion from it, writing that no gap had been found. That conclusion was reasonable given what it could perceive, which is the whole problem. From inside a channel, narration is indistinguishable from a working channel. A listener receiving the right information cannot tell whether the system delivered it or the model happened to be chatty that turn.

Accessibility resting on a model's prose habits is discoverability by luck. A terser response, a different system prompt, a model tuned to skip the summary, and the same task yields silence with no warning and no error.

Bill's ruling closed the gap rather than leaving it open. Tool results are not spoken, by design: "I listen to your thinking, and that is enough." Thinking and response text are both fed to the channel, so the channel a listener depends on is fully wired, and the exit code arriving through prose is the system working as specified.

That ruling also redirected the instrument. A blind persona that hunts for unspoken tool results is testing a decision rather than a defect. The right question is whether thinking and response text arrive completely, in order, and intelligibly.

Utterances 3 and 5 say they do not.

## 14.5 A bug found by ear

Two of the six utterances were a single backtick. The listener was hearing punctuation read aloud.

The obvious explanation is streaming. Deltas arrive on arbitrary boundaries, a fenced code block gets split across two chunks, and each half is filtered separately, so neither half contains a complete fence and the markers survive. That explanation is clean, mechanical, and wrong.

A test disproved it. Feeding the entire fenced block as one chunk, with no split anywhere near it, still produced spoken backticks.

The real cause sat one layer earlier. Filtering ran per phrase, and phrases were produced by splitting the buffer on newlines. A fenced block contains newlines by construction, so by the time the filter saw anything, the fence had already been cut into separate lines. A pattern that needs an opening marker and a closing marker to match will never match a line containing exactly one of them. The filter examined three fragments, found no fences, and passed all three through.

The fix resolves fences against the whole buffer before any splitting happens. The filter then sees a complete block and replaces it with a name, and the listener hears "code block" where a wall of syntax used to be.

The same test run surfaced a second problem with splitting on newlines, and it came from Bill rather than from the code. Prose wraps. A sentence broken across two source lines is one sentence, and splitting at the wrap produces two utterances with an unnatural pause between them. Speech engines work a phrase at a time, and every boundary the pipeline invents is a pause the listener hears.

So a lone newline became whitespace and a blank line stayed a boundary. A wrapped sentence is now spoken as one utterance, and paragraphs still separate.

That category of defect is worth dwelling on, because it is invisible to every other observer in this book. A DOM snapshot shows text that is present and correct. A screenshot shows a page that renders properly. A grader asserting on rendered content passes. The content is fine. The *segmentation* of the content is wrong, and segmentation only exists in the channel where text becomes time.

## 14.6 The bug nobody could hear

Listening is a better instrument than looking, for this channel. It is still not sufficient.

A part that arrives complete, with no deltas preceding it, was rendered to the screen and never queued for speech. The handler set the element's content and called `flush()`, which empties a buffer that in this path is already empty.

Reaching that code requires one setting. Turn streaming off, or use a model without the streaming capability, and every response arrives as a single final part. The screen fills normally. The agent says nothing at all.

No listener can report this. A person hearing silence cannot distinguish "the system failed to speak" from "the system had nothing to say," and neither can an agent. The failure produces no signal in the channel, which is precisely what makes it a failure.

It was found by enumerating the four feed sites and asking, at each one, what reaches speech and under what conditions. When the symptom is absence, reading the code is the instrument.

The naive repair introduces a worse bug. Queue the final part unconditionally and every streamed response gets spoken twice, once from its deltas and once from its final. The pipeline needs to know whether a given part already streamed.

That information already existed. The renderer keeps an accumulator of streamed text, keyed by part, populated only by deltas and never cleared. Its membership test answers the question exactly:

```javascript
if (!this.accumulated.has(id)) TTS.queueChunk(msg.text);
TTS.flush();
```

One line of new logic, and the state it consults was already being maintained for another purpose.

## 14.7 Why none of it was reported

The speech channel in this application has a daily listener. Bill depends on it, works through it for hours at a stretch, and had filed no bug report about any of the six defects in this chapter.

That is worth understanding rather than apologizing for, because it is the strongest argument here for building the instrument at all.

Four of the six are undetectable from inside the channel by construction. An error that renders to the screen and never reaches speech produces silence, and silence is what a turn with no errors also produces. A response that arrived whole during a streaming-disabled session sounds exactly like a quiet turn. A pause gate with one input wired makes no sound whatsoever, and its only symptom is a tool call that ran slightly earlier than it should have. Telemetry reporting an empty queue looks identical to a queue that is genuinely empty.

The other two are audible and get absorbed. A listener hearing a stray backtick stops noticing it within a day. Broken words at chunk seams sound like the synthesizer, and every synthesizer mangles something, while an unnatural pause mid-sentence reads as network lag. People are extraordinary at filtering noise out of a channel they depend on, which is a useful adaptation and a poor property in a bug reporter.

What the daily listener produced instead was the suspicion quoted in §14.2, which is a claim about the instrument rather than a list of defects. Someone who works inside a system can often tell that a region of it is under-observed while being the worst available witness to what is wrong inside that region.

The supervision protocol during the repair followed the same division of labor. Direction was given freely: which task to run next, which binary was stale, which stale server was still holding the port while a rebuilt one served nothing. Findings were withheld entirely. A supervisor who volunteers the answer cannot distinguish an instrument that works from one that agrees.

## 14.8 What a program cannot verify

A grader for this chapter can check a great deal. Whether a word survives being split across three fragments. Whether a wrapped sentence produces one utterance. Whether any markup marker reaches the channel. Whether a part that arrived whole was spoken, and whether a part that streamed was spoken twice. Whether the pause gate releases only when both of its causes are clear.

All of those are properties of a transcript, and a transcript is a pure function of the fragments that went in. The grader loads the speech module with a stubbed speech synthesizer, feeds a fixed sequence, and asserts. No browser, no model, no API key, no flake.

What no test in this chapter can check is how any of it sounds.

Pronunciation is outside it. Whether `HTTPServer` read as three words is clearer than `HTTPServer` read as one is a judgment about ears. Prosody is outside it. Whether a rate of speech is intelligible for eight hours is outside it, and it varies by listener, by voice, and by fatigue.

The boundary deserves to be stated plainly in the chapter that builds the instrument, because a grader claiming more than it checks is worse than no grader. This one verifies that the right text reaches the speech channel at the right boundaries. A human decides whether the result is worth listening to.

There is a hardware fact on the far side of that line. Browser speech synthesis tops out around two to three times normal rate. A dedicated engine runs comfortably at roughly 750 words per minute, which is where an experienced listener actually works. That gap is an API limitation rather than an application defect, and knowing which is which determines whether the next hour goes into the code or into replacing the synthesizer.

## 14.9 Exercise, graded

The exercise is the speech pipeline and the blind persona, verified against the contract in the TL;DR.

| check | points | property |
|---|---|---|
| `tts-buffers-fragments` | 20 | No utterance breaks a word. Text reassembles. |
| `tts-filters-markup` | 20 | No markup marker reaches the channel. A fence is named. |
| `tts-boundaries` | 15 | Newline is a space, blank line is a boundary, trailing text flushes. |
| `tts-speaks-unstreamed` | 15 | A whole part is spoken. A streamed part is not doubled. |
| `tts-expands-identifiers` | 10 | Identifiers split into words. Ordinary words are untouched. |
| `tts-gate-both-causes` | 10 | Unpause only when both causes clear. A space counts. |
| `ch13-parity` | 10 | Chapter 13 still passes. |

Two of those carry twenty points for a reason that is visible in this chapter's history. `tts-filters-markup` and `tts-speaks-unstreamed` cover the two defects that survived a full grader suite, an automated observer, and a human listening to the output every working day.

The fence check runs its fixture twice, once as a single chunk and once split across deltas. The single-chunk case is the one that disproved the streaming theory, and a grader that only tests the split case would pass an implementation that still speaks backticks.

The identifier check asserts that ordinary words are left alone, which is the mutation guard. An implementation that inserts a space before every capital letter passes the positive case and fails this one.

## 14.10 Taking it for a spin

Rebuild, refresh the browser tab so the client picks up the new speech module, and run the driver with the blind persona against a task that produces identifiers in prose:

```
./virtual-user --agent-url ws://localhost:8084/ws --persona blind \
  --task "Ask the agent to explain how queueChunk handles max_tool_rounds,
          then report exactly what you heard."
```

The report comes back as a transcript rather than a description of a screen. `queueChunk` arrives as "queue Chunk" and `max_tool_rounds` as "max tool rounds". No backticks, no asterisks, no fragments cut mid-word, and nothing displayed that failed to arrive.

The same command with `--persona sighted` produces a report about the DOM and says nothing about any of this, which is the chapter in one comparison.
