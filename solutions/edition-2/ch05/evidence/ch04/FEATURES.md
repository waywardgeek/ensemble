# Chapter 4 live feature/action/result checklist (established before runs)

Each provider uses the real `chat` client in a PTY. The coder reads each response before choosing a follow-up. The model uses Ensemble tools; external direct debugger commands do not count.

| Feature | Human action to request | Observable result to retain |
|---|---|---|
| Background work/local jobs | Start a slow command with short callback; read notes while it runs; wait | Running snapshot, intervening file job, later output and one terminal event |
| Interactive PTY | Start interactive.py, wait for READY>, send alpha with default Enter and beta plus LF with append_newline:false | Echo, delayed replies, identical fresh prompts, process remains running |
| Shell isolation/cwd | Start persistent sh; cd subdir by input; separate pwd; explicit cwd; missing cwd then recovery | Shell-local directory changes, separate workspace cwd, override snapshot, refusal |
| Bounded reporting/recovery | Produce 100 numbered lines with small byte budget; read omitted line range from artifact | Head/tail omission count and full artifact; middle recovered via read_file |
| One-shot limits | Setter consumed by an ordinary tool; setter consumed by kill; explicit command override; subsequent default call | Consumption notes, precedence, and absence of sticky note in next result |
| Debugger | Build known-variable program, run dlv, break main.main, continue, step as needed, print value, quit | Actual dlv artifact with breakpoint and value 42, tool call records |
| Explicit kill and EOF cleanup | Kill one blocker; leave another running until clean EOF | Killed snapshots, reasons kill_job/shutdown, independently dead PIDs |
| Human client controls | /history, valid /redact, /usage, EOF | Readable output, durable redaction, final measured usage after cleanup |
| Provenance/replay | Inspect event identities/usage; dump/render after run | Requested/returned provider identities; deterministic offline facts with no new work |

Local fault/race tests and public-library consumers supplement these paid runs. GUI is an optional tested stub, not a live browser demonstration.
