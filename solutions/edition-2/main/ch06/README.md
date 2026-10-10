# Chapter 6 exercise

From this directory, run `go run .` with `LLM_VENDOR`, `LLM_MODEL`, and
`LLM_API_KEY` configured. `LLM_BASE_URL` is optional. Discover a supported model
from your provider rather than assuming the book's historical names are current.

The program reads one topic and runs author → editor → reviewer:

```json
{"kind":"prompt","text":"Write about the ocean at dawn"}
{"kind":"hint","agent":"author","text":"Make it vivid"}
```

Hints target the named Agent. `{"kind":"interrupt","agent":"author"}` ends
its turn; other jobs continue. An interrupted pipeline returns its error instead
of handing an incomplete draft to the next role. EOF waits for the workflow.

Observations on stdout carry `observation` and `agent`. Request completion emits
`turn_ended`; completed stages emit `pipeline` markers. `CH06_LOG` selects the
observation log. Appending `.author`, `.editor`, or `.reviewer` gives the separate
Agent event logs. Each Agent has only its role's tools. `think` is an exercise
fixture tool, not an Ensemble builtin.

For human chat, run `go run . chat` from the parent directory. Type ordinary
text; messages while busy become hints. `/interrupt` or Ctrl-C ends the turn,
then another message starts a new turn. Ctrl-D flushes submitted work and shuts
down, stopping managed processes and joining any remaining Go handlers.

Embedding clients use `Ask` or `Submit` for reliable request completion. `Post`
and `Wait`/`WaitAny` expose live interaction and bounded progress; progress is
not a request acknowledgement. Install tools before the first request. Observers
must return promptly and queue slow output themselves. `History.Load`/`Append`
are offline operations; submit live facts through `Agent.Record`. `LogPath`
selects saving before request completion. Empty `LogPath` leaves persistence to
the embedding client.

Media uses `Part{Type:"blob", MIME:..., Ref:Ref{Kind:RefPath, Locator:...}}`.
The renderer reads the local file into its vendor request. `LookupModel` has
explicit capability entries; unknown models still accept text but cannot render
media. URI and handle attachment rendering remain deferred. Job output handles
continue to identify recoverable bytes without becoming attachments.
