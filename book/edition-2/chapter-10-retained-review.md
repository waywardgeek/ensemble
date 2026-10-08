# Chapter 10 retained behavior integration

October 8, 2026. Coordinator-owned independent integration, against immutable
student source `8882a18cf98e9a4b70afccfbe980f6344630aae6`. This is a subset
preparation result, not complete retained behavior or Chapter 10 acceptance.
The coordinator has read current implementation and reviewer fixtures; that
exposure is separate from the student's preserved new-only exercise.

The earlier broad run on `122b04a` exhausted disk during final GUI-directory
restoration and emitted no final JSON. Its external error/empty result remain
preserved. No pass count can be recovered from that run. Correction `c8353a2`
adds optional progress callbacks and an atomic per-command receipt wrapper.
Nine Python orchestration controls pass, including default command/result
equivalence, late failure/interruption preservation, source refusal before any
write, subset labeling and explicitly coordinated cache maintenance. These
controls do not establish Ensemble behavior.

The first source-bound compatibility subset completed with three passing rows
and two failures. Both executables built and the retained queue/projection group
passed. The physical-reader fixture panicked through an old nil embedded
`Codec` method. The older stale-response fixture panicked through its nil
embedded `AppendFailure` method. Its controlled owner also lacked the explicit
no-session response needed by the new close path. Original receipts are retained
in `checkpoint-evidence/ch10-retained-compatibility-first.json`.

Adapter `be9ce12` edits only a disposable checker bundle. The physical-reader
fixture becomes an external composition test that constructs the real persistence
codec with its actual fixture parent; it does not stub validation or add a
sideways production import. Calls use the same actual event-log reader and all
physical-bound assertions remain. The controlled stale-response owner explicitly
reports no append failure and no mounted session. Its original behavioral
assertions and scoring are unchanged. The historical fixture files and student
source remain byte-for-byte untouched; exact adaptations and hashes accompany
the receipt. Optional adapter selection leaves the preceding gate's default
commands/results unchanged.

The adapted run passes both builds, complete physical-reader controls, the
retained Chapter 5 assertions and main-module vet. The nine orchestration
controls also pass after integration. Its receipt is
`checkpoint-evidence/ch10-retained-compatibility-adapted.json`; selected checks
and the incomplete full-gate status are explicit. All fixtures use local inputs,
with no credentials or provider requests.

Next: run the complete retained gate with per-command receipts after the
serialized fault-deletion stage. Investigate any further fixture drift separately
from actual runtime regressions. Complete delivered-tree discovery, all retained
groups and appropriate architecture inspection remain required; this subset
does not close them.
