Run from this directory with `LLM_VENDOR`, `LLM_MODEL`, and `LLM_API_KEY` set:

    go run .
    CH07_NO_STREAM=1 go run .

The default prompt exercises a tool and reply. An argument replaces the prompt.
The report counts deltas by kind and finalized parts, including time to first
delta. `api.log` records sanitized wire frames; `ch07-agent.jsonl` records facts.
Use a scratch working directory when trying tools. Unknown models use one-piece
responses; `EN_DISABLE_STREAMING=1` also disables streaming in the core CLI.
