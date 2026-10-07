# Chapter 3 human client review

Independent review, October 7, 2026. Code, evidence-verifier revision and
manuscript reconciliation accepted. Coordinator checkpoint/export is recorded separately in the validation
manifest; Bill's editorial approval is not inferred.

## Sources and independence

The human checker was prepared from the new Chapter 2 and Chapter 3 contracts
before opening this student's implementation. Its initial version is preserved
at outer commit `10ab549`. The student froze implementation and actual terminal
runs at `a347ce31511c4b124e486bb41ef98c07bd17cec5`. The recorded executable has
SHA256 `3d49d9037077929b7380fbbae020e968c79075ac94eb95c939535aa606523990`.
That executable passed the first independent black-box run before source review.

The reviewer reloaded the full coding skill and architecture, full voice guide
and chapter procedure, Chapter 2's human contract, and the complete current
Chapter 3. The implementation review read the client integration diff, current
chat/client entry and terminal owner path, and first-edition `solutions/ch03/main.go`
(last changed at `ef155b1cccecf90a1f6fa791e61cc800489d6f1c`). The broader original
tool implementation comparison remains in `chapter-03-code-review.md`.
This ledger does not claim the whole textbook is resident in current context.

The student used a fresh, new-only context and the approved preceding source.
This is a guided integration of an already reviewed human client, not a fresh
whole-chapter cold regeneration. No first-edition answer snippets were supplied
to the student. The reviewer performed no paid calls.

## Independent terminal checks

`accept_chat.py --chapter 3` passes **45/45** cases on the initial frozen
executable. Thirty-nine shared cases retain the Chapter 2 interface promises.
Six new cases exercise a complete tool turn with successful history-directed
redaction, and a failed continuation that retains completed tool effects, on
each of the three fake API surfaces. The Chapter 2-only unexecuted-tool notice
is deliberately excluded because Chapter 3 executes calls before returning.

The checker drives actual local PTYs and waits for the displayed prompt. It
inspects recorded calls/results, disk effects, outgoing pairing and signatures,
final-only display, cumulative usage and the next request's redacted result.
The seven-call fixture includes an ordinary failed read before later successful
calls to all six tools. Fixtures and projection checks are independent of the
student serializer. Fake APIs prove deterministic behavior, not paid usability.

Seven focused checker controls pass. Before the Chapter 3 implementation was
opened, the Chapter 2 executable still passed all 42 Chapter 2 cases and failed
exactly the six new Chapter 3 cases in the expanded checker. Those preservation
and absence receipts are retained in `checkpoint-evidence/`.

The updated `audit_chat.py --chapter 3` builds serial disposable copies. Its
passing control and seven deliberate defects all produce their exact expected
failure sets on the three API surfaces:

| Defect | Failing cases per provider |
|---|---|
| Disable terminal default selection | default PTY |
| Retain the extra slash | explicit/default PTY |
| Accept invalid UTF-8 | invalid UTF-8 input |
| JSON-escape answers | explicit/default PTY and completed tool turn |
| Omit prompt flush | explicit/default PTY and both tool scenarios |
| Hide history call IDs | completed tool turn/redaction |
| Retain original result during redaction | completed tool turn/redaction |

All source hashes remained unchanged. The audit is scoped sensitivity evidence,
not exhaustive coverage of every client or tool promise. Chapter 2 remains the
audit's default mode; its original mutation set and expected cases are unchanged.

## Comparison and revision

The first edition already offered explicit human chat. The initial second-edition
machine interface therefore lost a real user path despite passing its graders.
The repaired client restores that path and improves exact input handling,
terminal default selection, discoverable history, local directives, readable
multiline output and checked prompt flushing. Its public submission service
keeps the CLI from constructing or duplicating internal model/tool orchestration.

The Chapter 3 integration appropriately removes only the stale Chapter 2
tool-only notice and changes the inherited fixture to require actual continuation,
paired ordinary-error results, accumulated usage and final-only display. A
second CLI tool loop would have been a regression; none was introduced. The
owner interface still reaches terminal helpers and the root logger. No new
production-code revision was warranted by this comparison. The old interactive
error-continuation policy differs from the new explicitly taught fail-fast
policy; that is a contract choice, not an unnoticed regression.

One evidence defect required repair: `verify-receipts.py` used a hardcoded
temporary executable and could rewrite derived requests/receipts without checking
its identity. The reviewer independently confirmed the coordinator's finding.
Revision `339a2a61e7107b921bdc7acbd704b2eb7da17931` changes only evidence tooling
and records, preserving all production and raw live receipts.

The revised verifier requires an explicit executable, repository, immutable
`a347ce3...:solutions/edition-2/main` source and new output directory. Before
executing the binary or writing output it compares the executable hash, historical
Git source hashes, initial binding, launch metadata, raw logs and terminal bytes.
Reconstructions go outside the original evidence directory. This works after
main advances to later chapters and does not relabel reconstructed requests as
intercepted HTTP.

The reviewer read the revised verifier and its isolated controls, then reran
them. Wrong executable, wrong source and altered launch metadata all fail before
output creation; the wrong executable is never run. The original binary and
source reproduce every original derived byte. Evidence remains byte-identical.
The skill and Chapter 0 now teach the binding requirement where it first matters.
No paid rerun was needed for this evidence-only correction.

## Actual terminal and prose review

The reviewer inspected all three complete main terminal transcripts, EOF
transcripts, feature ledger, source/binary bindings, usage records, tool facts
and reconstructed-request verification. Each main session shows all six tools,
write guard and edit refusals followed by recovery, separate streams and actual
exit 7, empty-file distinctions, Unicode caps, history-directed redaction,
one-request ephemera, literal slash input, remembered model-invented names,
usage and clean quit. EOF sessions finish with four zero counts.

| API | Human turns | Tool calls | Input | Cache write | Cache read | Output |
|---|---:|---:|---:|---:|---:|---:|
| Messages | 5 | 27 | 78489 | 0 | 0 | 3313 |
| Chat Completions | 6 | 29 | 10777 | 0 | 33664 | 1381 |
| generateContent | 5 | 27 | 120234 | 0 | 0 | 4452 |

The verifier independently recomputes these totals from accepted responses.
The chapter retains the omitted OpenAI overwrite call and its corrective human
follow-up, and corrects the model's mistaken description of the complete `é`
from the actual tool bytes. These details teach why inspecting effects matters.
The quoted Messages slash/recall excerpt and the final counters agree with the
terminal records. Provider plans differ, so these are not efficiency benchmarks.

The revised spin explains a usable human invocation and distinguishes original
machine/public-consumer evidence, later human sessions and offline reconstruction.
No Bill participation, Linux terminal run or live GUI is invented. Malformed
encoding, oversize input and provider failures remain local controls. The
optional GUI is still a separately built stub. No material prose or contract
finding remains; final status/link reconciliation does not alter these receipts.

## Receipts

Coordinator preservation inputs, copied without changing their bytes:

- `/tmp/ensemble-ed2-ch03-human-independent-initial.json`: initial 45 cases.
- `/tmp/ensemble-ed2-ch03-human-mutations.json`: passing control and seven defects,
  exact failure sets, source/checker/audit hashes.
- `/tmp/ensemble-ed2-ch03-human-verifier-independent.json`: independent rerun of
  the repaired verifier controls.

These reports were copied byte-identically to `independent-initial.json`,
`independent-mutations.json` and `independent-verifier.json` under the accepted
Chapter 3 export's `evidence/ch03/human-chat/`. Reviewed source commit
`8494fdb0e5d6c096445bfac039458b83dd225332` includes them; the coordinator's
`solutions/edition-2/manifests/ch03-r1.json` binds that exact export and review.
The raw paid receipts retain their initial source and executable identities.
