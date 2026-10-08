# Chapter 10 remaining boundary preparation

October 8, 2026. Independent preparation against frozen runtime
`8882a18cf98e9a4b70afccfbe980f6344630aae6`. No first-edition persistence
comparison or student implementation edit occurred. The reviewer read the full
mandatory coding skill, architecture ledger, current Chapter 10, public API and
semantic format, and relevant inherited Chapters 2, 4 and 9 rules. Existing
`20820c1` bounds and `beea2ca` / `401473c` fault/lifecycle coverage remains intact.

The new command was prepared before compiler release. Its subsequent runtime
results are recorded below:

```sh
python3 scripts/edition2/accept_ch10_remaining.py SOURCE_DIRECTORY \
  --source-commit 8882a18cf98e9a4b70afccfbe980f6344630aae6 \
  --receipt RECEIPT.json --audit
```

Six new groups distinguish these limited claims:

| Group | Positive and distinguishing refusal | Scope |
|---|---|---|
| Aggregate entries | One million typed entries split between two lists, then one extra | Owning Engine's semantic validator and admission counter; no serialized million-entry import |
| Aggregate parts | One million actual part slots split across entries, then one extra | Same component boundaries; shared immutable test backing avoids duplicate allocation without reducing counted slots |
| Seen requests | One million request IDs with matching distinct, ordered terminal turn facts; next admission refuses | Semantic validator and admission counter; no million real model turns |
| Streaming log bytes | Genuine initializer followed by real generated whitespace through exactly 1 GiB; +1 and a longer suffix refuse | Actual streaming log reader with reduction and at most fixed-buffer read-ahead; **not** physical regular-file proof |
| Public construction facts | Genuine initializer and imported anchor offered to exposed standalone/session Agents; refusal preserves bytes and subsequent ordinary append works | Public append authority, including standalone adoption |
| Store leaves | Genuine imported store first passes; each of four leaves becomes a symlink or directory; live and offline open refuse; restored store reopens | Regular/non-symlink policy and failed-construction reservation release |

The component fixture obtains its owner and base state from a genuine session
export. It derives admission indexes from actual collections, never supplies an
invented `Context.Index`, and never lowers a production bound. The remaining
tests use scratch stores only. Empty entry parts follow the published ordered
parts grammar; these are component positives rather than claims that a human
client can submit a million empty prompts.

Six deletion adapters remove the corresponding guards only after
all selected real positives pass. Aggregate entries/parts have more than one
check, so their deletion removes those checks together to avoid credit from a
different same-bound guard masking the intended defect. Each mutant must reach
the matching runtime assertion; compiler, source/hash and fixture failures earn
no detection credit. The initial preparation receipt claims no runtime result.

The runner verifies all 158 runtime/asset/API files against the immutable commit
before creating its overlay, then verifies identities afterward. It saves the
started command before execution and each result atomically. Three Python
orchestration controls pass: identity refusal preserves prior evidence,
interruption retains completed and started command records, and a failed
positive prevents all mutations. Formatting prints no Go filenames. These are
preparation results only; no compiler or large fixture ran while the coordinator
and retained-suite reviewer held the exclusive build slot.

## Explicit remaining coverage

These prepared groups do not close the chapter's deterministic matrix:

- The 256 MiB canonical semantic-state and 512 MiB checkpoint/origin physical
  file exact/+1 controls are not prepared in this subset. A later fixture should
  derive a valid small checkpoint first, use canonical-state byte arithmetic
  independently, and stream whitespace for physical size without constructing
  a second huge input buffer. Keep snapshot admission and existing-file refusal
  error classes separate.
- A valid physical 1 GiB log cannot fit in the observed 733 MiB free space.
  Streaming proves the actual reader's byte boundary only. Do not call it a
  regular-file acceptance result. All large physical fixtures must run one at
  a time after explicit resource release.
- Seen call IDs and activation totals have coupled semantic constraints. Calls
  retain their original parts, so exceeding the call limit can also violate the
  part limit. A million-activation valid parent additionally requires complete
  material and transition provenance within canonical-state, catalog, event and
  source bounds. A generic million-element JSON array is not such a parent.
  Do not invent unsupported watermarks, omit material/history, or credit an
  earlier unrelated refusal. Distinguish an unreachable combined exact state
  from a missing counter check before imposing a new teaching expectation. The
  reachability analysis below now settles these two counters for this codec.
- The streaming control bounds consumed input and read-ahead. It does not
  measure peak allocation, establish bounded event encoding, or prove every
  read/admission path. The long-overflow reader may report generic record-read
  context rather than the exact file-bound diagnostic; the +1 fixture requires
  the named 1 GiB diagnostic.
- Whole-group uint64 exhaustion remains separate. Existing student Skills tests
  already exercise a two-activation group at the last two IDs and rejection
  when only one remains, through an explicitly labeled runtime counter seam.
  This is not a valid large-watermark snapshot or independent deletion audit.
  Event-group and job allocator exhaustion also remain unclosed here.
- Missing origin, origin without log, checkpoint without log, full-origin with
  extra origin, and partial import failures still need a complete independent
  store-topology matrix. Store leaf type checks above do not substitute for it.
- Controlled Skills preflight versus terminal session file/count admission,
  durable limit consumption followed by append failure, noninherited lock
  descriptor during an actual surviving tool child, and root job allocator floor
  preservation require their own ordered runtime controls. Earlier ordinary
  append faults, process-death release, restored job display and request cursor
  tests do not establish these combinations.

No missing ownership rule was found in the teaching for these prepared controls.
The semantic codec's earlier watermark prose is corrected by its final Q5
clarification; the fixture follows the current chapter's exact represented
activation/job maxima. Live use, retained integration and historical comparison
remain coordinator-owned gates.

## Runtime and deletion results

The coordinator released the exclusive compiler slot after retained-suite work.
`ch10-remaining-first.json` preserves the initial fixture failure: formatting
and full main-module vet passed, but every new group stopped at construction
because the reviewer omitted a dummy API key. No runtime requirement or mutation
was tested by that failed setup. Adding a literal noncredential key and an
unreachable localhost endpoint repaired only the fixture configuration.

`ch10-remaining-config-corrected.json` passes all six actual groups, with the
source's 158 identities and all recorded checker hashes unchanged. The streaming
group consumes genuine generated bytes through the 1 GiB boundary; it creates no
large file. All six subsequent mutants reach the matching failing runtime
assertion. Their exact source edits, original/mutated hashes, commands and output
are retained in that receipt. This closes the six stated scopes, not the other
coverage listed above. The separately recorded complete main-module and public
race checks are in `ch10-remaining-module.json`.

## Why two nominal count limits cannot independently bind here

Every seen call retains its original tool-call Part in complete semantic state.
Redaction cannot erase that identity. Therefore `seen calls <= total parts` for
both unsettled live state and settled snapshots. In settled snapshots every call
also has a distinct tool-result Part, so `2 * seen calls <= total parts`. At the
one-million total-part limit, a settled snapshot has at most 500,000 calls. A
one-over call fixture cannot independently test the call guard while satisfying
the required part ceiling. This is a consequence of retained identities, not a
reason to invent an invalid positive or remove the declared ceiling.

For activation counts, the **current published codec** supplies a stronger size
bound. Every nonprimary activation has immutable material and a permanent skill
Entry, including an empty manual and a retired activation. Minimal canonical
wire elements use a one-character name, one-digit coordinates, empty bodies and
sets, and the shorter Boolean spelling. A material element is at least 212 bytes
and its skill Entry at least 98. This deliberately ignores array separators,
digit growth, transition history, the primary, and every other state field.
Consequently a million activations need at least
`999,999 * 310 = 309,999,690` bytes, above the 268,435,456-byte state ceiling.
The loose maximum from these two elements alone is 865,921 activations.

A live tail cannot bypass that snapshot bound by importing just below it and
growing to one million without saving. A valid catalog contains at most 256
definitions, so a transition activates at most 256 new records and at most 256
records remain active. Every changed transition records the complete retired
summary set. Starting from any valid origin, the last 523 activation-producing
transitions needed to reach a million each occur after 865,921 activations and
contain at least 865,665 retired summaries. The shortest summary is 27 bytes.
Those selected records alone therefore occupy at least
`523 * 865,665 * 27 = 12,224,055,465` bytes, beyond the 1 GiB log ceiling. A
full-origin store starts with no more than 256 initial activations and satisfies
the same conservative argument. There is no compaction or forgotten origin in
this chapter that would remove those required records.

The exact arithmetic is retained in `ch10-remaining-reachability.json`. It proves
that the activation count ceiling cannot bind first under this frozen codec's
other published invariants. It does **not** prove the same for every possible
alternative codec, establish that the 256 MiB/1 GiB guards themselves work, or
replace their pending exact boundary controls. No teaching contradiction was
found: the chapter explicitly accepts exact bounds only when all other rules
hold. The declaration remains meaningful without an impossible combined positive.
