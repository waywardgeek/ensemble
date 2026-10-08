# Chapter 7 reviewed browser ownership demonstration

On October 8, 2026 (Pacific time), the student drove the public two-Agent
consumer in actual Chrome 154.0.8037.93. This revision uses immutable runtime
and support `9ba7855b31a5eb134819602b35eb9acb277f6342`, with all 94 source files,
six support files, 74 browser dependencies and eight executable hashes in
[initial-binding.json](initial-binding.json). The word “initial” in that filename
means this revision's launch binding; it does not relabel the earlier Chapter 7
runtime or its 44-request receipts. The coordinator archived exact executables.

Before paid use, 18 local adapter controls passed, including a valid real browser
launch, specific identity refusals before writes, the 4-HTTP ceiling, and two
actual browser prompts with third-prompt refusal and remount. The immutable
independent gate passed 33/33 and comparative repair controls passed 12/12.
The final whole-application-close repair prevents queued speech from starting
while children are being disposed. Its controlled stale-callback check remains
separate from real native timing.

| Provider/model | Prompts | HTTP | Browser PID | Captured files |
|---|---:|---:|---:|---|
| Anthropic `claude-sonnet-4-6` |2|2|9992|[audio3](consumer-anthropic-r1/audio-3.wav), [audio19](consumer-anthropic-r1/audio-19.wav)|
| OpenAI `gpt-4.1-mini-2025-04-14` |2|2|11054|[audio3](consumer-openai-r1/audio-3.wav), [audio15](consumer-openai-r1/audio-15.wav)|
| Gemini `models/gemini-3.8-flash` |2|2|12483|[audio4](consumer-gemini-r1/audio-4.wav), [audio16](consumer-gemini-r1/audio-16.wav)|

All sessions exited 0. The authorized cap was two prompts and at most four model
HTTP requests per provider, separate from the initial demonstration. No automatic
retry or extra prompt occurred. Each provider's first Agent answered about a
morning walk; its second answered about an evening garden. The first card was
spoken while the idle second view closed and reconnected; that replacement then
accepted the second prompt. The second card was later captured while the idle
first view closed and reconnected. Both Pages used one application/native service.

Each run then started A, queued B, canceled queued B, queued B again and canceled
active A. Native callback records show only A before active cancellation, then
one B start. A's unsent correction retained typing while speech cleared. The
precise action spans are A12–17, O8–13 and G9–14 in each raw browser JSONL.
Anthropic's first attempted queue sequence (7–10) occurred after A had naturally
finished, so it is retained but does not prove queued-peer cancellation. The
explicit second selection (12–17) supplies that distinguishing real sequence;
no model regeneration was needed.

All six raw request bodies exactly match source-bound independent replay:
[Anthropic/OpenAI](reconstructed-anthropic-openai/receipts.json) and
[Gemini](reconstructed-gemini/receipts.json). Original raw logs, response bodies,
terminal transcripts, WebSocket frames, UI actions, screenshots and WAVs remain
separate from derived audits. `derived-audit.json` in each session binds the raw
browser/launch/WAV hashes, response text/stop reasons, action timing and measured
audio. Before writing those audits, the historical94-file source set, support,
executables, complete browser dependencies and each launch identity were checked.

The WAVs are 48 kHz stereo IEEE float32. Each measured first PCM second is digital
silence (reported as −200 dBFS floor); seconds 2–6 have RMS between −19.93 and −21.42 dBFS.
ScreenCaptureKit selected the isolated test browser's exact PID, application name
and bundle identity. The requested two-second lead is wall time; capture startup
is asynchronous. These samples establish produced audio, not transcription or
human listening. The idle peer was closed/reopened after the native start and
before capture completion, with no intervening speech settlement callback.

Gemini spent most of the example's 512-token output budget on thinking. Both
answers ended with `MAX_TOKENS`: raw candidate token counts 18/20 and thinking
counts 490/488, with normalized output 508 each. Their visible text has 17/19
whitespace-delimited words instead of the requested 70-word paragraphs. The accepted short answers and empty text part
remain in the raw receipts; there is no claim of length compliance. They still
supply distinct actual provider text for the revised remount/queue/audio path.
No retry was made to conceal this budget outcome.

The student inspected the OpenAI browser 12 screenshot: both Agent panels are
visible; A shows typing 1/speaking 0 with the unsent correction, while B shows
speaking 1. Each ArtifactScroll crops its own viewport, so a full-page screenshot
does not display every retained card; the accompanying body-text and raw events
supply content outside that viewport.

Initial Chapter 7 all-feature human CLI/browser/public embedding receipts remain
unchanged. Killed-job reconnect is proved using the retained real initial event
through the revised public snapshot and actual browser transport; no paid job
was regenerated for that projection-only correction.
