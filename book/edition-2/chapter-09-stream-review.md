# Messages empty-argument repair review

October 8, 2026. Independent narrow review of runtime repair
`75a72554bbb13f5e86c8b806cff780b5fcd79c22`, following the real Chapter 9
Anthropic GUI failure. This is not the post-freeze historical comparison.

The original `live-anthropic-g/responses/002.body` is a complete stream. Its
list_directory block starts with input `{}`, supplies one zero-length
partial_json string, stops that block, and ends with tool_use and message_stop.
The source at `c0e3171` incorrectly discarded the complete start object merely
because a delta event existed. The old Chapter 6 contract already allowed an
empty delta sequence to use that object; clarification `305b1b0` explicitly
distinguishes zero aggregate bytes from zero events. Neither text permits
falling back after malformed nonempty replacement bytes.

The five-line production change ignores only an empty argument fragment. Actual
argument bytes still replace the initial object, and parser errors remain
checked by the enclosing operation. No admission, authority, terminal-frame or
response-acceptance gate changed. The student's seven focused parser cases are
useful; independent checks below exercise the public CLI, actual local HTTP,
durable facts and real file effects as well.

```sh
python3 scripts/edition2/ch09-review-empty-stream.py CLI_BINARY \
  --receipt /tmp/ch09-empty-stream-review.json
```

The final matrix has 14 cases. Six positives cover the exact retained response,
no delta, one or several empty deltas, nonempty replacement fragments interleaved
with empty fragments, and a terminal held behind observed progress. Eight
negatives cover malformed JSON, whitespace-only bytes, null/array replacements,
a null start object, a wrong-type delta, missing message_stop and a provider error.

The revised CLI SHA-256
`230167aaffe730928c69cdde744fa228eb9bd616221ff7f74c09accf6d9ed9e1`
passes all 14. The same final checker against original CLI
`bde9a0f01737d41d10a8431f7dca64bef2486ff10a3aee8d9398794c71a13577`
fails exactly the four empty-delta positives and passes the other ten controls.
This is a distinguishing before/after comparison, not an implementation mutant
or a provider rerun.

For the timing control, the local server withholds message_stop until the test
observes a marker emitted after the complete argument block. At that barrier
there is no accepted response, final part, completion, tool_called or file effect.
After release, the positive writes the exact expected file and receives one
successful matched result. The truncated and error variants never dispatch or
accept content. All successful cases expose exactly decodable argument display
and the intended call arguments; replacement bytes never execute the initial
placeholder path. The unmodified real response accepts list_directory with `{}`
and its recorded 2,016 input/51 output usage.

`ch09-empty-stream-final-bindings.json` associates final receipts with both
complete immutable bindings: 133 original and 134 repaired source files, their
executables and support. The reviewer verified both bindings, original response
hash and unchanged binaries. The student's separate 11-module vet/test receipt
also passes and its affected input hashes match the repair. No Go compilation,
cache growth, whole-tree copy, credential read or paid request was needed for
this independent review.

The first harness attempt encountered a cleanup race when process-group killing
met an already exiting error process. Its receipt is preserved; those cleanup
errors are not runtime findings. The reviewer changed only its own protocol-child
cleanup to close input and join before any necessary process kill, then retained
the clean baseline. Intermediate 12-case results are also preserved; the final
two explicit non-object replacement controls extend them without weakening an
assertion. The final receipts are `ch09-empty-stream-before-complete.json` and
`ch09-empty-stream-repaired-final.json`.

The narrow repair is accepted. The original paid failure remains at `c0e3171`.
The later GUI recovery with explicit path arguments is a different observed
action, not evidence that the old runtime handled the empty-argument response.
This local replay does not claim repaired live-provider use or corrections to
earlier frozen chapter snapshots; their propagation remains coordinator work.
