# Anthropic public Skills attempt: partial outcome

October 8, 2026. This narrow review concerns `live-anthropic-p`, source
`c0e3171fdc22834348f976f81fcaa7372bd13ed4`, Messages model
`claude-sonnet-4-6`. The process exited 1 after its original four-request cap.
The review does not change that outcome or authorize a retry. Chapter 9 initial
runs are still in progress; no historical solution comparison was performed.

The reviewer read the approved live plan, complete public consumer, terminal,
four captured requests, both event logs and the relevant raw response streams.
The independent audit validates the immutable binding before replay: 133 source
files, nine executable identities, support/catalog/browser bindings, complete
original file identities and the launch. It reconstructs the four requests in
the consumer's observed sequential order and compares exact HTTP bytes, removing
only the CLI output LF. All originals, including the failed launch, remain
unchanged. Results and hashes are in `ch09-review-anthropic-p-partial.json`.

```sh
python3 scripts/edition2/ch09-review-public-attempt.py \
  solutions/edition-2/main/evidence/ch09/live-anthropic-p \
  --binding solutions/edition-2/main/evidence/ch09/review-binding.json \
  --receipt /tmp/ch09-anthropic-p-review.json
```

This is an audit of a failed attempt, deliberately separate from the student's
success-only verifier. It never changes that verifier or a launch exit code.
No credentials, provider calls, Go builds or original-file writes are needed.

| Required public Skills evidence | Actual observation |
|---|---|
| Two Agents under a public application owner | Consumer constructs alpha and beta through public APIs; separate logs and prompts receive two model responses each. This live use is sequential, not a concurrency test. |
| Different installed ceilings | Alpha initialization includes write_file; beta does not. Alpha requests 001/002 declare it; beta requests 003/004 omit it. |
| Typed load and independent state | Alpha loads edit at revision 1. The consumer's public beta getter still reports revision 0 before beta's own operations. Beta edit is refused as unavailable in its ceiling, then beta inspect succeeds at revision 1. The terminal retains both owned inspections and contributor lists. |
| Per-Agent frozen material and literal scalar expansion | Recorded bodies and their digests are valid. Alpha carries literal alpha-$TOOLS, beta carries beta-$TOOLS; no opposite Agent marker appears in its requests. Each request contains its one dedicated activation-2 manual once. Repeating the same manual in the next request is retained history, not a second activation. |
| Actual model/tool use and separate data | Alpha reads its 41-byte note and beta its 40-byte note. Each returned result and cr/io artifact equals that Agent's original note bytes; files remain unchanged. |
| Paired management and state transitions | Beta's redundant load inspect returns unchanged at revision 1 without another transition. Its later unload commits revision 2, retires activation 2, and retains a paired successful result. Alpha's recorded state remains its own edit activation. |
| Request captures and accounting | All four chronological replays are byte-exact. Raw terminal SSE usage matches each durable response: alpha totals 2,475 input/122 output; beta 2,199 input/179 output; both cache counters are zero. |

Alpha completes the requested read/report task and prints the note and literal
project markers. Beta instead proposes an unnecessary load inspect alongside its
read, then unloads inspect on its second response despite the prompt saying not
to call another tool. The library retains the complete accepted final tool batch
and ends the turn with `round_limit` under the captured two-request policy. There
is no beta final report or successful consumer completion. Those absent outcomes
must remain explicit in the demonstration and any aggregate evidence summary.

The observed beta failure is provider task noncompliance followed by the intended
bound, not evidence of a Skills library failure. The chapter-required public
two-Agent calls, different ceilings, isolated state, actual read results and
literal material delivery are exercised. Returned-value mutation, races and
other deterministically controlled properties remain covered by their separate
local gates; this live attempt does not add such claims.

No additional paid call is needed merely to obtain successful final prose.
Retain this attempt as partial, teach the observed noncompliance and request
bound, and keep the standard success verifier strict. This recommendation does
not waive another scenario/provider feature, accept the entire live matrix or
close the later public-consumer usability and historical quality review.
