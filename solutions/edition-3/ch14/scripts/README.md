# Executable speech harness

`tts-harness.sh PROMPT [--type-during-turn TEXT]` launches the real C# GUI and an
owned browser in transcript mode. It submits through the actual form and writes
the page's admission log, not a reconstructed answer. `--describe` prints the
read-file and streaming contract without launching anything.

Build `Examples/Gui/Gui.csproj` first. Supply existing executables explicitly:

- `TTS_NODE`: Node executable.
- `TTS_DOTNET`: .NET 10 executable.
- `TTS_GUI`: absolute built `Gui.dll` path.
- `TTS_CHROME`: existing Chrome executable.
- `TTS_PUPPETEER`: absolute ES module exporting `puppeteer` (or the default driver).

The harness installs nothing. The verified host is macOS/Chrome using the already
available driver; other platforms are not established by that evidence. Node and
the driver are test dependencies, not library or GUI dependencies.

The caller supplies `LLM_BASE_URL` for its local Anthropic-shaped fixture,
`WORKSPACE` as a fresh owned directory, `TTS_LOG` as the output JSONL path, and
`STREAMING=on|off`. Provider credentials are dummy, and Chrome receives only the
selected runtime environment. The known model capability is explicit. This is
not a launcher for paid-provider testing or the human's connected browser.

From the repository root, the scoped grader runs five independent launches:

```sh
go run ./cmd/grade-edition2 -ch 13 -executable-only -json solutions/edition-3/main
```

It retains all nine speech checks (95 points), including the zero-point harness
check. It excludes inherited Go/Chapter 12 checks. The harness requires normal
GUI shutdown and writes its stdout/stderr under the fixture workspace. Grading
owns and removes that workspace. The chapter's portable formatter regression is
`node Tests/SpeechTests.mjs`, run from the source root.

A transcript entry means admission to the production speech queue. It does not
prove a native voice spoke, that anyone heard it, or that the voice is usable.
