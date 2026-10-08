# Chapter 7 independent live-evidence review

Coordinator audit on October 8, 2026. Initial evidence is frozen at `8cc87f2`,
with runtime `da162e8` and support/binary binding `ba902b7`. This accepts the
initial receipts as observations of that implementation. It does not close
comparative findings or attribute these runs to a later repaired executable.

The independent local audit passes 24 controls. It replays all 44 captured
requests across nine actual CLI/browser/public-consumer sessions, first on the
original paths and then on valid copied paths. Source, executable, support,
browser dependency and last-session identity mutations each trigger their
intended refusal before derived writes. A changed final request body also fails
before output. All 386 original session files match the initial Git checkpoint
and remain unchanged after the audit. A credential-value scan finds none of
the three configured provider keys in those files.

| Provider | Prompts | HTTP requests | Verified scope |
|---|---:|---:|---|
| Anthropic | 9 | 15 | Combined browser/human CLI, plain CLI, two-Agent embedding |
| OpenAI | 10 | 14 | Same scope; initial no-tool response and explicit correction retained |
| Gemini 3.8 Flash | 9 | 15 | Same scope; initial user-card audio distinguished from subsequent answer audio |

Every session exited zero and stayed within its 12-prompt/32-request ceiling.
All logs retain contiguous durable sequence numbers; request counts agree with
launches, captured bodies, response files and independent reconstruction. The
browser records contain no action failure or page exception. Each provider's
scratch file contains the actual 13 bytes `BROWSER-CH07\n`; its logs retain an
interrupt, the surviving job's later deliberate kill, and the browser answer
after terminal EOF. Hint text appears in subsequent raw requests. In Gemini's
first such request, it is a text part in the same user content as the matching
`functionResponse`, with the preceding model call's signature retained. Model
compliance is a separate observation: the OpenAI run delivered the hint but
omitted its requested answer marker.

Four WAV files were independently reanalyzed, including Gemini's initial
user-card selection. Each hash and exact browser PID agrees with its capture
receipt. The first PCM second is at the -91 dB analysis floor. Generated-answer
speech intervals measure mean levels of -21.3, -19.7 and -20.7 dB respectively.
These establish produced audio from the isolated test browser, without claiming
transcription or that Bill listened. Capture startup is asynchronous; the
requested two-second wall-time lead does not establish two seconds of recorded
silence. The coder preserved that distinction.

The exact CLI, GUI, public-consumer and audio-capture executables are archived
outside the repository, with identities in
`checkpoint-evidence/ch07-executable-archive.json`. The complete audit is
`checkpoint-evidence/ch07-initial-live-audit.json`; its reproducible command is:

```sh
python3 scripts/edition2/ch07-review-evidence.py
```

No provider calls are made by this audit. The independent browser controls,
comparative code review and source-specific revised demonstrations remain
separate. In particular, initial live handling of `job_killed` does not prove
correct reconnect replay of that event, and single-panel audio does not prove
isolation between two speech queues sharing a document. Those review findings
must be repaired and checked before final chapter acceptance.

## Revised implementation receipts

The coordinator independently accepts the narrowed revised demonstrations at
evidence checkpoint `341f15d80a80bfecf7ccceaf7b7840de63c20f7f`, running source
`9ba7855b31a5eb134819602b35eb9acb277f6342`. The repaired browser application and
native speech service were exercised through the public two-Agent embedding
with actual models on all three providers. Each used two prompts and two HTTP
requests, within its separate two-prompt/four-request authorization. All three
sessions exited zero; no browser action failure or page exception was recorded.

The independent audit passes ten controls: original and copied-path replay,
specific source/source-set/last-launch identity refusals before writes, a changed
final request body, and three actual queued-cancel/active-handoff sequences.
All six original requests reconstruct exactly. All 145 retained session files
match their committed blobs and remain unchanged after auditing; none contains
the configured provider credential values. The student's separate 18-control
launch/preflight audit retains its own scope and source identity.

Each provider's first answer was spoken while the idle second view closed and
remounted on the same DOM. That replacement then submitted the second prompt.
Later the second answer was spoken while the idle first view closed/remounted.
The raw speech callbacks, view events and captured audio agree on this order.
A separate actual sequence starts A, queues B, cancels queued B, queues B again,
then cancels active A. Only A starts before active cancellation; exactly one B
starts afterward. A's applied pause observation retains typing while clearing
speech. Anthropic's earlier attempt happened after A naturally finished and
remains explicitly excluded from this stronger queue claim.

Six WAV files were independently reanalyzed using ffmpeg: all first PCM seconds
are at its -91 dB floor; seconds 2–6 measure means from -19.9 to -21.4 dB.
The student's float32 analysis reports digital silence at its own -200 dBFS
floor; these numerical floors are analysis conventions. Browser PIDs and WAV
hashes match capture receipts. This is produced audio, without transcription
or human-listening claims. The coordinator also inspected OpenAI screenshot 15
and Gemini screenshot 16: both show the remounted first view and independently
speaking second view. Card viewport clipping is distinct from missing content.

Gemini's two answers ended with `MAX_TOKENS` under the embedding example's
512-token budget. Raw visible token counts are 18/20, thinking counts 490/488,
and normalized output is 508 each. The short actual answers support these
browser/audio controls; they do not satisfy the requested 70-word paragraphs.
There was no retry. The first-edition comparison's killed-job repair is covered
by local replay of the retained real event through the revised public API and
actual browser transport, rather than a regenerated paid job.

One initial reviewer assertion incorrectly demanded immediate applied pause
state in Gemini's body-text capture. A completed click does not imply the
asynchronous acknowledgement has arrived; the later screenshot/receipt timestamp
does not date the preceding text read. The corrected check requires the exact
applied wire observation before the next action, while retaining typing in the
visible capture. The failed reviewer assumption and correction are preserved in
`checkpoint-evidence/ch07-revised-audit-timing-correction.json`; no source or
original receipt changed and no model call was repeated.

Reproduce this audit without provider calls:

```sh
python3 scripts/edition2/ch07-review-revised-evidence.py 341f15d80a80bfecf7ccceaf7b7840de63c20f7f
```

The result is `checkpoint-evidence/ch07-revised-live-audit.json`. Exact revised
CLI/GUI/consumer/capture executables are archived outside the repository with
identities in `checkpoint-evidence/ch07-revision-executable-archive.json`.
Initial 44-request evidence continues to cover its original unaffected scope;
the six revised requests do not pretend to repeat every initial demonstration.
Final manuscript review and export/tag remain separate acceptance steps.
