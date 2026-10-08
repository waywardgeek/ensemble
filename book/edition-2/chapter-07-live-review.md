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
