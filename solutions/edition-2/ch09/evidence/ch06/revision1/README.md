# Revised live evidence

Runtime source: `75bd14d5d2424778ebf45cb9025e9dbf314f956a`.
The first repair `3ccaed6` remains preserved; no paid call used it. Original
Chapter 6 runs retain runtime `aa5f86a` and their original helpers/binding.

The student coder, not Bill, drove seven actual PTY sessions on October 7, 2026
local time (October 8 UTC). Root approved the affected-path plan before these
20 requests. Each `launch.json` retains exact executable paths, hashes, selected
model, delivery, dates, exit status and raw request count. `binding.json` binds
all 73 historical Go/module files, both revised executables, Python interpreter,
PTY recorder and the three launch/replay helpers. Helpers passed nine local
controls before calls; original helpers remain unchanged.

| Directory | Model | Actual requests | Result |
|---|---|---:|---|
| stream-anthropic | claude-haiku-4-5-20251001 | 4 | File read/continuation, active interruption, recovery |
| stream-openai | gpt-4.1-mini | 4 | File read/continuation, active interruption, recovery |
| stream-gemini | models/gemini-3.8-flash | 4 | File read/continuation, active interruption, recovery |
| plain-gemini | models/gemini-3.8-flash | 2 | File read/continuation, completed answer displayed once |
| public-anthropic | claude-haiku-4-5-20251001 | 2 | Two Agents, matching typed finals, completion before slow release |
| public-openai | gpt-4.1-mini | 2 | Two Agents, matching typed finals, completion before slow release |
| public-gemini | models/gemini-3.8-flash | 2 | Same boundaries; both accepted partial MAX_TOKENS answers |

Each human streamed session used these inputs, waiting for each response and
sending `/interrupt` only after observing an actual integer fragment:

```text
Use read_file to read notes.txt. Report its exact marker and port, then stop. Do not modify files.
Print every integer from 1 through 4000, one integer per line. Do not use tools. Begin immediately.
/interrupt
Reply exactly RECOVERED-SIX-REVISION. Do not use tools.
/quit
```

The plain session used the first prompt and `/quit`. The public consumer's
bounded prompt is retained in its event logs and source. All three actual
interruptions reported `interrupted=true`; each interrupted operation added no
accepted response or usage, and each recovery succeeded. Four exact read_file
artifacts retain `CHAPTER-SIX-FILE-MARKER` and `port=8080`.

The public Gemini outputs reached the example's 512-token limit and retained
signed empty text. They did not complete the requested 80-word explanations.
No exposed thinking or subscriber overflow occurred in these runs. Deterministic
fixtures provide those guarantees separately; slow callbacks alone do not prove
overflow. The optional GUI remains a stub.

`verify-receipts.py` reconstructed all 20 requests semantically after identity
preflight, retaining results under `reconstructed/`. `summarize.py` checked raw
log chronology, file results, interruption boundaries, terminal usage totals,
public Agent identities and final-part equality. `receipts-manifest.json` hashes
75 original files, including 10 event logs and four tool artifacts. Raw receipts
remain separate from reconstructed output. No retries or runtime edits occurred
within this revised demonstration.

For example, from the repository root the human launcher was:

```sh
python3 solutions/edition-2/main/evidence/ch06/revision1/terminal-run.py anthropic claude-haiku-4-5-20251001 chat stream-anthropic
```

Other rows substitute the provider/model and directory; public rows use
`consumer`, and the plain row adds `--plain`. Existing directories are never
overwritten. The launcher reads authorized credentials into memory/environment;
no key is placed in argv or the repository.
