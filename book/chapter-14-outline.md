# Chapter 14 outline — Accessibility: The Channel Nobody Tested

*Status: DRAFT for Bill's LGTM. Written by the author (Opus 4.6) from the
code Bill and Opus 5 built on 2026-09-20. Per `chapter-writing-procedure.md`
Phase 1, prose does not start until this is LGTM'd and the grader exists.*

## Title candidates

1. **The Channel Nobody Tested** (preferred: names the finding)
2. Testing What You Cannot See
3. The Observer That Could Not Fail

## Through-line stake (voice.md §4)

A user who cannot see the screen has exactly one channel, and that channel was
broken in ways no one could detect, because the only automated observer the
book had built looked at the DOM. It resolves when a speech-only observer finds
a defect by ear that no DOM observer could ever have seen.

## Per-section wild fact (voice.md §7.3)

| section | the one fact performed |
|---|---|
| 14.1 | An observer only finds bugs in the channel it needs to succeed. |
| 14.2 | `tts_queue` read a variable nothing ever wrote, and was injected every round. |
| 14.3 | Deleting `tts.js` entirely would have left every prior test passing. |
| 14.4 | The blind run SUCCEEDED, for the wrong reason. |
| 14.5 | A fence in a single chunk leaked backticks. Streaming was not the cause. |
| 14.6 | Turn streaming off and the agent goes silent while the screen looks fine. |
| 14.7 | An agent can verify the pipeline and can never hear it. |

## Voice plan

- **Register**: Spolsky for mechanism, one Lewis beat at §14.4.
- **Bill moments**: 2. The domain knowledge about phrase boundaries, and the
  ruling that tool results are not spoken.
- **Confessions**: 2 of 3. (a) predicted the blind run would fail, it passed;
  (b) assumed the fence bug was caused by streaming, disproved it by test.
- **Motivational opener**: yes, Bill's voice, §1.2. Content is his, and the
  material in voice.md §10.1 covers it: no macular vision, 80 wpm, listens at
  speed. Needs his sign-off or a `[STORY SLOT]`.

---

## TL;DR spec (draft — this is the grader contract)

### What you are building

A speech channel that a program can verify, and an observer that can only
perceive speech.

Two deliverables:

1. **A speech pipeline** that turns a stream of arbitrary text fragments into
   utterances a person can actually listen to.
2. **A blind persona** for the virtual user of chapter 13, which perceives the
   agent through speech alone and has no access to the DOM.

### The speech pipeline contract

The pipeline receives text as **arbitrary fragments**. A streaming model emits
deltas on byte boundaries, frequently mid-word. The pipeline owns the job of
turning that into speech.

Rules:

1. **Buffer until a phrase boundary.** Never speak a fragment as it arrives. A
   word split across three chunks is one word.
2. **A sentence-ending period is a boundary. A newline is not.** Prose wraps.
   Splitting at the wrap breaks the sentence. Treat a single newline as a
   space; treat a blank line as a boundary.
3. **Provide an explicit `flush()`.** End of stream, a completed tool
   announcement, and an error each force whatever is buffered to be spoken.
   Without this, text ending in a colon is never heard.
4. **Filter for the ear before speaking.** Markdown markers are not words.
   Emphasis, inline code, headings, bullets and link syntax are removed. A
   fenced code block is *named*, not read.
5. **Resolve fenced blocks against the whole buffer, before splitting.** A
   fence broken into lines can never be matched by a fence pattern, and the
   orphaned markers get spoken.
6. **Expand identifiers.** `camelCase`, `snake_case` and `HTTPServer` are read
   as separate words. Ordinary words are left alone.
7. **Speak parts that arrive whole.** A part delivered without deltas must
   still be spoken, and a part that streamed must not be spoken twice.
8. **Speak errors.** An error that is displayed and not spoken is invisible to
   a listener.
9. **Record a transcript.** Every utterance entering the channel is recorded
   with its text and its source, whether or not audio is produced.

### The pause gate

The agent pauses tool dispatch while the user is reading or typing. The
predicate is:

```
paused = speaking OR input_non_empty
```

Rules:

10. **One derived predicate, not three call sites.** Compute the value, compare
    it to the previous value, and send only on an edge.
11. **A space counts as input.** `trim()` is wrong here.
12. **Unpause requires both causes clear**, regardless of which one changed.

### The blind persona

13. **Deny the DOM.** The blind persona has no snapshot tool and no click tool.
    If it can see, it will use sight and the speech channel goes untested.
14. **Bypass audio.** A test that waits for real speech runs in real time.
    Record to the transcript and return immediately.
15. **It keeps**: hear, type, submit, wait, sleep. Nothing else.

### Yours

Which markdown constructs to filter beyond the required set. What to name a
fenced block when you skip it. Whether the transcript is a ring buffer or
unbounded, and how large. The wording of the tool announcement. Whether errors
preempt the queue or append to it.

### Exercise

`make grade14`

---

## Proposed grader (for the coder)

**Mechanism.** The speech pipeline is plain JavaScript with one browser
dependency. The grader loads the student's speech module under `node` with a
stubbed `speechSynthesis`, feeds a fixed sequence of fragments, and asserts on
the transcript. Deterministic, no API key, no browser, no flake.

**Prerequisite to rule on:** this requires `node` on the grading machine. The
alternative is a CLI probe mode on the agent binary, which matches chapter 2's
`render LOG` contract but cannot reach code that lives in the browser. Bill's
call. Recommendation: node, with a clear prerequisite error if absent.

| check | pts | property | fixture |
|---|---|---|---|
| `tts-buffers-fragments` | 20 | no utterance breaks a word; reassembled transcript equals input | `"Hel"`, `"lo wor"`, `"ld. Bye."` → one utterance `"Hello world."`, then `"Bye."` |
| `tts-boundaries` | 15 | newline is a space, blank line is a boundary, trailing colon flushes | wrapped sentence → 1 utterance; paragraph break → 2; `"Next:"` + `flush()` → spoken |
| `tts-filters-markup` | 20 | no `*`, `` ` ``, `#`, `_` reaches the channel; fence is named | run BOTH one-chunk and split-across-deltas |
| `tts-expands-identifiers` | 10 | `camelCase` → two words; plain words untouched | the untouched case is the mutation guard |
| `tts-speaks-unstreamed` | 15 | whole part is spoken; streamed part is not doubled | drive both paths |
| `tts-gate-both-causes` | 10 | unpause only on both clear; space counts as input | edge assertions |
| `ch13-parity` | 10 | prior chapter still passes | |

Total 100.

**Why the weights.** `tts-filters-markup` and `tts-speaks-unstreamed` carry the
most points because those two defects survived a passing grader suite, a
sighted automated observer, and a human listening daily.

**Mutation tests required (P9).** Each check must fail when its behavior is
deleted from the reference. Specifically: remove the buffer and
`tts-buffers-fragments` must fail; make fence resolution per-line and
`tts-filters-markup` must fail; delete the `accumulated` guard and
`tts-speaks-unstreamed` must fail on the doubling side, not only the silence
side.

---

## Body sections

### 14.1 The idea in plain words

**Thesis:** an automated observer only finds bugs in the channel it depends on
to finish its work.

Chapter 13's virtual user completed every task by reading the DOM. The speech
module could have been deleted outright and every one of those runs would still
have passed. Feedback an observer does not *need* is feedback it ignores, and
an observer asked to evaluate a channel it does not depend on will produce a
confident answer in either direction.

The fix is not more reporting. It is to make the untested channel the only way
to succeed.

Everyday framing: a fire alarm inspector who checks the wiring by looking at
the panel will never find that the speaker is disconnected. Someone has to
stand in the building and listen.

### 14.2 What the channel actually was

**Thesis:** the speech channel had four feed sites, and one tool that claimed to
report on it was reading a variable nothing ever wrote.

- `tts_queue` read `window._ttsQueue`. Two references existed in the tree, both
  of them reads. It always fell through to a contentless `"(speaking)"`.
- It was declared `ephemeral: "round"`, so the engine injected it **every
  round**. The agent had been receiving useless speech data continuously.
- The pause gate specified `paused = tts_speaking OR user_typing`. Only the
  typing half was ever wired. Line 1 of the speech module had claimed "pause
  integration" since the day it was written.

Receipt to print: a documentation comment describing a feature the file does
not contain.

### 14.3 A user who can only listen

**Thesis:** channel isolation is the instrument; denying capability is how you
build it.

The blind persona is the chapter-13 driver with the DOM tools removed at
startup. Not discouraged in a prompt. Removed from the registry, so the
capability does not exist.

Scope ruling, Bill's: hover-to-speak belongs to the operating system's screen
reader, not to this GUI. So the GUI owns the streaming speech channel and
nothing else, a low-vision persona would mostly be testing the screen reader,
and for this stack low-vision and blind collapse into one instrument. There is
no second persona worth building. State this plainly; it is the kind of scope
decision that saves a reader a week.

### 14.4 The run that succeeded for the wrong reason

**Thesis:** from inside a channel, narration is indistinguishable from a working
channel.

The prediction was written down before the run: the blind user would fail to
report an exit code, because tool results are never spoken. It succeeded.

The transcript, in order:

```
1. "run command"                                      <- tool dispatch
2. "The command failed with exit code 1:"             <- the model's prose
3. "`"
4. "ls: /nonexistent-path-xyz: No such file or..."    <- the model's prose
5. "`"
6. "This is expected since..."                        <- the model's prose
```

No utterance from a tool result appears. The listener heard the exit code
because the model chose to narrate it. The observer did not fabricate; it
reported its transcript accurately and then concluded "no gap found", which was
wrong.

Bill's ruling, and it closes the gap rather than leaving one: tool results are
not spoken. Thinking and chat text are, and that is the channel to verify.

Items 3 and 5 are the next section.

### 14.5 A bug found by ear

**Thesis:** the observer heard punctuation, and the cause was not the one the
author assumed.

Lone backticks reached the channel. The obvious explanation is that streaming
split the fence across chunks. That explanation is wrong, and a test disproved
it: a fence arriving in a **single** chunk leaked too.

The real cause: the splitter broke on newlines before the filter ran, so the
fence was already in pieces and the fence pattern could never match. Fences
have to be resolved against the whole buffer first.

Bill's domain knowledge, which the code did not have: broken words are
mispronounced, and every non-boundary seam is heard as an unnatural pause. A
listener perceives the seams a reader never sees.

### 14.6 The bug nobody could hear

**Thesis:** listening is not sufficient either.

A part that arrives complete, with no deltas, was displayed and never queued.
Reachable in one setting: disable streaming, or use a model without the
capability, and the entire response arrives as one final part. The screen looks
perfectly healthy and the agent is silent.

No listener can report text that was never sent. This one was found by
enumerating the feed sites, which is the method when the symptom is absence.

The naive fix speaks every streamed part twice. The correct test for "did this
part stream" already existed in the accumulator, which deltas populate and
nothing clears.

### 14.7 What a program cannot verify

**Thesis:** name the boundary, because a grader that claims more than it checks
is worse than no grader.

Verifiable: completeness, ordering, duplication, segmentation, timing, and
whether markup reached the channel.

Not verifiable by any test in this chapter: pronunciation, prosody, whether a
rate is intelligible, whether a voice is tolerable for eight hours. The
observer has no ears.

Print the ceiling honestly: browser speech tops out near 2 to 3 times normal
rate, against roughly 750 words per minute on a dedicated engine. That is an
API limit, not an application defect, and the chapter should say which is
which.

### 14.8 Exercise, graded

Standard section. Contract, `make grade14`, the checks table.

### 14.9 Taking it for a spin

The second blind run, against real model output, with the transcript showing
`queueChunk` heard as "queue Chunk" and `max_tool_rounds` as "max tool rounds".
Per procedure §2i, include a screenshot with a full text description.

---

## Coder list (for Opus 5)

1. Build `internal/grade/ch14_*` per the table above, including the mutation
   tests. Grader must not read the reference solution.
2. Decide the node prerequisite question if Bill has not ruled.
3. Snapshot `solutions/ch14`, tag `ch14-solution`.
4. Confirm `make grade14` is 100/100 and prior chapters unaffected.
5. Write the post-build review for the author (procedure Phase 1, step 3).

## Open questions for Bill

1. **Title.** Preference above is "The Channel Nobody Tested".
2. **Node as a grader prerequisite.** Recommended, with a clear error if absent.
3. **The motivational opener.** This chapter's subject is his daily experience,
   so the opener should be his voice. Needs his text or a `[STORY SLOT]`.
4. **Process note.** Whether the chapter says out loud that this one was built
   spec-and-code intermingled rather than design-first. It is a real lesson and
   the book has printed process failures before, but it is his call.
