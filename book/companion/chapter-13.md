# Chapter 13 — A listener who can only hear

The first-edition listener experiment was supposed to fail. A restricted agent asked the target to run a command against a nonexistent path and report the exit code. Successful tool-result bodies did not enter automatic speech, so the prediction was that the listener would hear the dispatch announcement and miss the answer.

It got the exit code right.

The retained transcript explained why: the dispatch announcement was followed by the model's own explanation of the failure. The raw tool result had never been spoken. The listener accurately reported the words and incorrectly concluded that their delivery proved the channel had no gap. Correct information had arrived by a different route.

Bill's scope decision clarified the intended interface: “I listen to your thinking, and that is enough.” Thinking and response text provide commentary; successful tool bodies stay quiet unless a person deliberately reads an artifact. Failed tools and provider errors must speak. This avoids turning a file read into an unsolicited audiobook, while making the model's explanation responsible for conveying useful results.

Source labels make that policy observable. A transcript needs to distinguish thinking, response text, tool announcements, errors, and deliberately requested speech. Thinking and errors also need spoken prefixes, because a person listening cannot inspect a JSON label.

## Structure before phrases

Two entries in that historical transcript were lone backticks. Streaming seemed the obvious culprit: perhaps a fence had split across deltas. Feeding the entire block as one chunk still produced backticks.

The actual error came earlier. The pipeline split on newlines before filtering Markdown. Opening fence, body, and closing fence reached the filter separately, so no fragment contained a complete block. The repair resolves fences against the whole raw buffer first, replacing a complete block with “code block.” An unfinished fence waits for more input or an explicit final flush.

The current C# GUI's browser formatter follows that ordering. It preserves link labels but removes destinations, strips presentation markers, and expands identifiers such as `HTTPServer` and `max_tool_rounds` into speakable words. It handles a supported prose subset rather than pretending to be a complete Markdown parser. Inline code can carry essential meaning; a fenced implementation usually benefits from being named rather than recited.

A lone newline becomes whitespace. A blank line remains a paragraph boundary. Otherwise source wrapping invents pauses in the middle of sentences. Deltas can end inside words, links, or punctuation sequences, so the formatter retains raw text and releases stable phrases with sufficient lookahead. Explicit flushes release unfinished tails before a tool announcement, error, or final completion.

Each displayed part retains an offset into its raw source, not into the shorter filtered copy. Deltas append text; the authoritative final replaces the displayed value and releases only the unconsumed suffix. A final without prior deltas therefore speaks its whole text, while a duplicate final speaks nothing. The historical implementation displayed a whole final and flushed an empty buffer, producing a perfectly readable answer that nobody could hear.

Restored parts remain silent, including through a later final. Resuming narration midway through a restored fence could expose code whose opening delimiter the speech cursor never saw. New live parts can speak normally.

## What the transcript actually records

The page owns one active native utterance and a FIFO of future phrases. It records an utterance only when that phrase reaches the front of the queue, immediately before native synthesis and silent transcript mode diverge. Cancelling a queue must not publish its discarded paragraphs as speech already delivered.

This admission log includes ordering, timing, origin, and part identity. Transcript mode uses the same queue and formatter but skips synthesis. It is useful precisely because it observes the application rather than reconstructing an expected answer through a second text-cleaning path.

Admission is still not sound. Native callbacks can establish playback transitions; neither callbacks nor text records establish pronunciation, comfortable pacing, working speakers, or intelligibility to a particular listener. Those require authorized audio and a person listening. Keyboard navigation and general screen-reader usability are separate concerns.

Speech also affects tool pacing. Queued phrases count as busy, not merely the current utterance. The browser combines that speech cause, when the preference enables it, with human typing. Human whitespace counts as a draft. Programmatic input does not become human typing when a later speech callback examines the field.

Pause commands go out only when the combined local cause changes. Reasserting an unchanged pause on every callback would let one tab repeatedly defeat another tab's explicit unpause. Escape cancels speech without deleting the draft; typing may therefore continue to hold the Agent paused. Stopping the voice is not necessarily permission to proceed.

## A listener without a side door

The separate `Examples/Listener` application constructs an empty Agent with exactly five tools: `say`, `listen`, `sleep`, `set_speech_speed`, and `save_history`. It installs neither builtins nor a Skills manager. A prompt asking the listener not to inspect the page would be a request for cooperation; dispatch refusing guessed inspection tools is an actual capability boundary.

Privately, the Session uses a browser connection to submit through the real form, observe turn completion, read admitted utterances, and acknowledge rate changes. The model receives no selectors, screenshots, DOM snapshots, arbitrary evaluation, target response text, or raw tool bodies. Target and listener have separate provider configurations; the browser receives no provider credential.

`say` marks the target pending before submission and refuses another turn while it remains pending. Only the real completion frame clears that state. A fleeting idle display before acknowledgement cannot masquerade as completion. `listen` advances its own hearing cursor, returning promptly for new utterances; a timeout with work pending is not idle. Saving semantic history does not consume that cursor and records typed prompts and heard entries, not the listener's reasoning transcript.

These tools deliberately wait for completion. The listener has no `wait_for_job`, so returning an ordinary running handle could leave a pending listen consuming entries that the model never receives. The existing Tool option makes dispatch await the real Job and return its complete structured result. It does not introduce another scheduler or result store. Cancellation of an observer and ownership of the underlying operation remain separate.

Session declarations and persistent state live in Common. Functional partial sections implement the channel and browser lifetime directly in the same application assembly. No callback bag is needed to connect those parts.

The listener owns a fresh GUI, browser profile, and private connection, never the human's working page. Shutdown joins the listener Agent's Jobs before disposing the connection and stopping the owned processes. EOF, signals, deadline, and failed startup take that same cleanup path. An earlier Go/Node listener closed its browser but left a sleep timer alive, delaying exit almost until the sleep ended. The lesson is ownership, not a C# performance comparison.

The current source reports offline component, actual-page, and owned-process checks against loopback provider fakes. Those checks expose missing final narration, misordered filters, escaped capabilities, and stranded work. They cannot supply the last observation: whether listening to this application for hours is pleasant.
