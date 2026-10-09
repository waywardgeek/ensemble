# Chapter 10 final deterministic clearance

**Deterministic clearance: passed** for immutable runtime
`57d4aac3d26edcb78dc8d9fd27d7d0e1cce0ebf4`. No required deterministic failure or
unrun item remains in this finite integration pass. This does not grant chapter
acceptance: real-provider demonstrations, historical quality comparison, author
acceptance and the immutable chapter export/tag remain separate gates.

The [required-command evidence map](checkpoint-evidence/ch10-final-required-map-57d4aac.json)
hashes 19 selected receipts, verifies their available runtime source maps against
all 159 immutable runtime/API/asset identities, and checks the retained CLI/GUI
against the successful final source/build association. Receipts with narrower
maps keep that scope. Earlier failures and older-source deletion evidence retain
their original identities; none is relabeled as a new run.

## Required commands and results

`SOURCE` below is the verified small 57d4aac extraction. `CLI`/`GUI` are the
hash-checked binaries from the student's `retained-repair-build-watch` association.
Exact argv, hashes and outputs are in the linked receipts.

| Required command or group | Final result | Evidence disposition |
|---|---:|---|
| `accept_ch10_retained.py 57d4aac` | **70/70 joined rows** | [Full run](checkpoint-evidence/ch10-final-retained-57d4aac.json) was 68/70; [two affected reruns plus required builds](checkpoint-evidence/ch10-final-retained-parts-57d4aac.json) passed 4/4; [exact provenance join](checkpoint-evidence/ch10-final-retained-joined-57d4aac.json) selects each row explicitly |
| Build main CLI and separate GUI; complete delivered package discovery; vet/tests in every module; headless operation without GUI | Passed | Included in retained rows: 4,170 delivered files and all 12 modules, including the support consumer |
| `accept_ch10.py CLI` and `--self-test` | 93/93; oracle passed | Reused final-source CLI receipt; oracle rerun in [finite follow-up](checkpoint-evidence/ch10-final-affected-57d4aac.json) |
| `accept_ch10_public.py SOURCE` | 16 groups passed with race | Reused `retained-repair-watch-ch10-public.json`; actual source-map subset matches 57d4aac |
| `accept_ch10_clients.py CLI GUI --source-directory SOURCE` | 9/9 | Reused final-source client receipt and verified both binary hashes |
| `accept_ch10_arguments.py`, adapted retained management | 3 public argument/recovery groups; 138/138 management | Reused b2c8770 receipts; management changes only the duplicate-name expected field specified by dd1111e |
| `accept_ch10_faults.py SOURCE --source-commit 57d4aac` | 4 groups passed with race | Reused final source-bound post-cache receipt; original link ENOSPC failure remains |
| `accept_ch10_storage_bounds.py SOURCE --source-commit 57d4aac` | 5 groups passed with race | [Current-source rerun](checkpoint-evidence/ch10-final-storage-bounds-57d4aac.json); decimal/scalar, nesting, handler count/bytes and request uint64 controls |
| `accept_ch10_remaining.py SOURCE --run ...` | 5 selected groups passed | [Current-source rerun](checkpoint-evidence/ch10-final-aggregate-public-57d4aac.json): actual aggregate entries/parts/seen requests, public construction facts and store leaves; generated 1 GiB stream was not repeated |
| `accept_ch10_lifecycle.py SOURCE --source-commit 57d4aac` | 2 groups passed with race | [Current-source rerun](checkpoint-evidence/ch10-final-job-tail-57d4aac.json): real independent job tail and nonterminal invalid append |
| `accept_ch10_record_bounds.py CLI --source-directory SOURCE` | 38/38 | [Final bound executable rerun](checkpoint-evidence/ch10-final-record-bounds-57d4aac.json); closes the earlier intermediate executable's missing build association |
| Lifecycle edges, bounded admission/origin faults, canonical/physical large limits | Passed, reused | a944156, cefb823, 1befb73 and 9e2e94e receipts already bind 57d4aac; no repeated 256 MiB/512 MiB/1 GiB allocation |
| `make grade-dir CH=11 DIR=SOURCE` | 0/7; 0/100 diagnostic | Executed and preserved in finite follow-up. The inherited GUI/save.json startup does not match this edition's separate GUI/session API; all six persistence rows stop at `GUI server did not start`, plus inherited parity failure. The chapter explicitly calls this a diagnostic that does not cover its contract |

The final orchestration controls pass **13 retained-adapter tests plus 4 final
binding tests**. [Receipt](checkpoint-evidence/ch10-final-orchestration-57d4aac.json)
records source/binary mismatch refusal before evidence replacement and command
launch, and preservation of the unchanged historical checker.

## Why the two retained rows initially failed

The first full final-source run failed `retained-ch06-contract` and its dependent
mutation row. Public early identity and typed final mapping compared Parts using
reflect.DeepEqual. On each route, `Part.Parts` was nil in one public value and an
empty slice in its counterpart. This field is unused for text, opaque and tool-call
parts and omitted from their public JSON. The new snapshot clone explicitly
copies empty collections.

[Original diagnostic](checkpoint-evidence/ch10-final-stream-diagnostic-57d4aac.json)
keeps the original failing assertions and records exact field differences.
[The complete diagnostic experiment](checkpoint-evidence/ch10-final-stream-normalized-57d4aac.json)
reaches all later parts on all three routes: text, signatures, signed empty text,
tool calls, original argument strings and opaque data are byte-identical in public
JSON. The sole Go representation difference is the unused empty Parts slice.
Original, diagnostic and experiment fixture bytes are retained beside the receipts.

Chapter 6 §6.2 specifies owned neutral parts, order and final identity mapping;
it does not prescribe nilness for an omitted unused collection. The coordinator
approved a compatibility adapter after this diagnosis. Only the disposable
Chapter 10 copy changes, at exactly the two Part comparison sites. The helper
copies its values, normalizes Parts only when `Type != "tool_result"` and its
length is zero, then uses exact DeepEqual. Nonempty nested material, result parts,
raw/opaque/argument/text bytes, identities and all other fields remain exact.
No JSON roundtrip or general semantic-equality shortcut is used.

All **14 existing Chapter 6 groups and all 19 existing deletions** now pass their
controls. Each deletion first has its successful unchanged behavioral parent,
then its intended failing assertion. The first audit's seven blocked positive
parents receive no mutation credit. The adapter test reverses only its helper
and two substitutions and recovers the historical fixture byte-for-byte. The
original fixture, management checker and runtime were not edited. This is a test
compatibility correction, not a runtime repair or relaxed behavioral requirement.

The final retained result is a provenance join, not a claimed clean single run.
Both receipts have identical 4,170-file source maps. The only invoked fixture delta
is ch06_public_test.go.txt. Adapter orchestration and copied inventory-only final
review scripts are separately identified in the join. The management one-tuple
adaptation remains unchanged. No passing unrelated retained row was replaced.

## Published §10.10 matrix

| Required property | Covering final-source evidence |
|---|---|
| Construction compatibility | CLI 93, public selectors/configuration, clients and retained standalone routes |
| Ownership | Delivered package/module discovery, headless/GUI consumers, public mount reservations, fault lock lifetime and actual surviving-child descriptor control |
| Identity | Public owned resume/reservations, semantic mutations, public forged initializer/anchor refusal, request/job uint64 and root allocator-floor controls |
| Snapshot | Public owned export/import, prefix-free origin, subsequent save/reopen, semantic refusals, physical origin and origin-write faults |
| Tail/rebuild | Public three-route exact request/usage/Skills comparisons, argument recovery, actual independent job tail, honest pre-origin history refusal |
| Authority | Public Skills compatibility, plain System presence, CLI configuration canaries, retained policy/preferences controls |
| One-shot limits | Public literal attempts, setters, malformed/paused calls and ordering witnesses; lifecycle consumed-limit append-failure control |
| No resumed work | Public unfinished refusal/no startup effects, client historical jobs, root handle floor and represented uint64/collision controls |
| Capture/write | Public busy boundaries, checkpoint fault controls, independent job tail, close/worker/lock lifetime and origin faults |
| Lifecycle | Client EOF/quit/signals/GUI detachment, fault join/repeated close, surviving real child and lock release |
| Limits/corruption | Record 38, codec bounds 5, aggregate components, canonical 256 MiB, physical 512 MiB checkpoint/origin and 1 GiB log, bounded reads/preparation, Skills-versus-storage refusal |
| Clients | Public multi-Agent isolation, client checkpoint/watch/reconnect/history checks, retained browser/policy/preferences and preserved local support exercises |

Scope remains precise: count/admission and activation whole-group owner tests are
component controls, not fabricated million-turn or MaxUint64 imported histories.
The documented call/activation/event reachability analysis is specific to this
codec and its retained identities, provenance, cardinality and size constraints.
The repair adds optional argument preservation and stable collection copies; it
does not remove those constraints or shrink the prior conservative skill-material
and entry lower bounds. No new atomic event/job batch is inferred. Physical-file
boundary receipts exercise the stated public inspection paths, not every possible
live-open combination. Earlier 8882a18 deletion results remain earlier-source
proofs; affected positives were rerun here without reopening their mutation scope.

## Cleanup and remaining gates

The [cleanup audit](checkpoint-evidence/ch10-final-cleanup-57d4aac.json) found no
residual file larger than 64 MiB under this reviewer's test scratch prefixes.
The record runner removed its sequential 64 MiB payloads; full retained extractions,
diagnostic stores and temporary consumers were cleaned by their owning runners.
Prior small source adapters and retained binaries remain. No source, user file,
evidence, retained binary or module download was deleted. Observed free space
was 11,910,070,272 bytes; shared free-space changes are not attributed to this run.

Compiler ownership has been released. The deterministic residual list is empty.
The real-provider interface demonstrations, later historical quality comparison,
author/teaching review and final immutable chapter export/tag remain required for
chapter acceptance. No credentials, provider requests, historical solution copying
or student runtime edits occurred in this final integration pass.
