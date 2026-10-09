# Chapter 7 initial live receipts

The Codex student coder drove these actual human CLI PTYs and Chrome pages on
October 7, 2026 Pacific time (October 8 UTC); this does not claim Bill ran them.
Runtime source: `da162e821369784b37cd6a6544610429af5f1cbe`. Evidence support and
launch binding: `ba902b7322a9cb5dd71c8352d45d6975663409c0`.

| Provider/model | Admitted prompts | HTTP requests | Independent replay |
|---|---:|---:|---:|
| Anthropic `claude-sonnet-4-6` | 9 | 15 | 15 exact matches |
| OpenAI `gpt-4.1-mini-2025-04-14` | 10 | 14 | 14 exact matches |
| Gemini `models/gemini-3.8-flash` | 9 | 15 | 15 exact matches |

The approved per-provider ceilings were 12 prompts and 32 forwarded requests,
including tool continuations. No automatic retry or cap hit occurred. Every
session exited zero. No actual-browser action failure or page error occurred.

For each provider, `browser-VENDOR-r1` retains the combined human terminal,
GUI trace, raw provider bodies, durable log, actual browser websocket records,
action timestamps, screenshots and page text, scratch output, and WAV audio.
`plain-VENDOR-r1` retains plain human CLI parity; `consumer-VENDOR-r1` retains
the public two-Agent custom-layout demonstration. `launch.json` binds source,
all executables, support and browser dependencies. `initial-binding.json`
contains the exact hashes. `reconstructed-VENDOR-r1` contains separate replay
outputs; original logs and wire bodies were not rewritten.

The CLI entered a greeting, inspected history and continued. Both browser tabs
observed the same Agent. Reloads captured active operations for all providers;
Anthropic also captured nonempty partials. Each provider wrote/read
`BROWSER-CH07\n`, ran a delayed job, received a hint, survived a paused terminal
interrupt, then observed and deliberately killed the surviving job. A typing
tab closed without keeping its pause. Terminal EOF detached and a later browser
prompt answered `BROWSER-STILL-LIVE`. Public embeddings showed a first-Agent
typing pause while the second answered, then distinct responses for both Agents.

OpenAI initially described file work without calling tools. Its retained next
human correction explicitly required the calls, which then executed. All three
received the hint in the next raw model request; only Anthropic and Gemini
printed its requested marker. These are observed provider outcomes, not hidden
retries or claims of stronger model compliance.

Auto speech was enabled through its button, actual card speech started, and
cancellation retained another typing cause. `audio-audit.json` in each browser
run identifies the exact Chrome process, raw WAV hash and measured windows.
The first PCM second is silent at the -91 dB analysis floor; generated-answer
speech windows average -21.3/-19.7/-20.7 dB for Anthropic/OpenAI/Gemini. Capture
startup is asynchronous, so a two-second wall-time lead is not asserted to be
two seconds of PCM silence. Gemini's first audio selection spoke a user card;
its second specifically selected the model answer. These prove produced audio,
not human listening or transcription.

Actual screenshots inspected include Anthropic `browser-26.png`, OpenAI
`browser-25.png`, and Gemini `browser-25.png`: visible connection state,
retained tool results, interruption or running status, and client pause causes.
Their adjacent text receipts contain the complete rendered scrollable history.
Safe-content/accessibility, exact snapshot cuts, saturation, uncertainty and
stale speech callbacks are separately established by deterministic local gates;
paid timing is not claimed as their distinguishing proof.

`live-receipts.json` indexes counts and receipts. `student-review.md` records
teaching difficulties and initial experience before historical comparison.
The older `browser-anthropic` launch made zero model calls and keeps its original
binding; it is not one of the nine live sessions above. Original failed local
gate receipts also remain unchanged. A scan of all 481 then-present evidence
files found none of the three configured provider keys.
