# Chapter 10 large-state boundary controls

Subsequent update: Bill freed disk space; the separately released
[physical 1 GiB log control](chapter-10-physical-log-review.md) now passes.
[Bounded admission and origin fault controls](chapter-10-admission-origin-review.md)
also passed afterward. The capacity observations and pending statements below
describe the preserved earlier stage.

The canonical-state, physical checkpoint and physical origin groups passed on
immutable `57d4aac3d26edcb78dc8d9fd27d7d0e1cce0ebf4`. All three exact boundaries
were accepted and their one-byte overruns refused. Two targeted constant
mutations then failed the intended overrun assertions; the file mutation was
detected independently for both checkpoint and origin. Constants were increased
by one only in disposable mutation overlays, never reduced to fit a fixture.

The [positive receipt](checkpoint-evidence/ch10-large-state-anchor-calibrated-57d4aac.json)
binds all 159 runtime/API files, checker bytes, compiled executable and exact
commands. Formatting was empty; overlay vet and full module tests passed before
execution. Each case ran separately. Actual checkpoint/origin inputs were
536,870,912 and 536,870,913 bytes, created with sequential writes, not sparse-file
length changes. The valid origin anchor was rehashed over the actual padded
origin. Source hashes remained unchanged after inspection.

The [mutation receipt](checkpoint-evidence/ch10-large-state-mutations-57d4aac.json)
records the exact altered source hashes and compiled executables. Canonical
inspection passed before its mutation reached `canonical state one-over limit
accepted`. Each physical exact-file control passed before its mutation reached
`physical state file one-over limit accepted`. No compilation, setup or source
identity failure was counted as mutation detection.

The [first attempt](checkpoint-evidence/ch10-large-state-first-57d4aac.json)
failed a fixture assumption before any large allocation: its real small import
measured zero anchor size delta, while preparation incorrectly required a positive
delta. Session anchors are not renderable window events; at these sequence widths
their metadata changes need no additional canonical bytes. Calibration now permits
that measured zero and still rejects negative or implausibly large deltas. The
original fixture bytes and failed receipt are retained.

An orchestration self-check subsequently exposed an empty positive checker map
being accepted by the mutation runner. The runner now requires the complete exact
checker set. Eight refusals start from a passing identity-gate parent, and the
actual complete positive receipt passes the corrected gate. The failed self-check
and original mutation-runner bytes remain retained. This gate-only correction
does not relabel or rerun the unchanged Go controls; the mutation receipt remains
bound to the exact driver version that ran it.

Each large process required at least 768 MiB free before starting, budgeting
roughly 512 MiB of fixture payload plus 256 MiB headroom after compilation.
The machine has 16 GiB RAM. Processes used `GOMEMLIMIT=1GiB` and `GOGC=25`; the
former is a soft GC target, not a claim of a 1 GiB memory ceiling. Observed peak
resident memory was approximately 1.6 GiB and `time -l` reported zero swaps.
The cases completed in roughly 4–20 seconds. No Go cache cleanup was necessary.
Only bounded runner inputs, binaries and receipts remain; large owned temporary
files were removed by the tests.

The final free-space observation was 847,114,240 bytes. A real 1 GiB log plus the
same 256 MiB headroom requires at least 1,342,177,280 bytes, so physical log-file
proof remains pending. The older 1 GiB streaming check does not fill that gap.
These checks also do not measure bounded allocation independently or prove live
opening at the physical state-file bound; their physical-file path is public
offline inspection. Controlled Skills preflight versus terminal storage and
origin-creation write faults remain separate gates. No provider, credentials,
student source change, historical comparison or full chapter acceptance is claimed.

## Preserved initial preparation

October 8, 2026. This is an uncompiled, unrun design and source preparation,
not acceptance evidence. Runtime target remains immutable `8882a18`.
`scripts/edition2/ch10-large-state-files_test.go` contains two prospective
groups, sharing the already validated real-owner helper. The student currently
owns the compiler. Disk last observed at 395 MiB free cannot accommodate these
fixtures. No physical 1 GiB log is proposed or allocated.

The canonical-state group starts with an actual 104-instruction session export.
Its oldest four instructions fall outside the retained 100-event window. It
extends those bodies with ASCII bytes to reach exactly 256 MiB of canonical
state. This is synthetic complete history derived from a real parent; it does
not claim that the earlier command created those large bodies. Each hypothetical
source event is separately sized below the unchanged 64 MiB event limit. Public
inspection must validate the complete candidate before it is useful as a parent.
One additional byte must fail canonical output admission.

A small genuine import first measures the semantic size change caused by its
anchor. The live large import reserves that exact overhead, so its subsequent
checkpoint lands at the canonical bound. An additional accepted instruction
then makes export and checkpoint exceed the bound; both must preserve the
previous checkpoint, origin and log. A further valid append must still work.
This distinguishes a recoverable snapshot refusal from terminal append failure.
It does not demand that an exact-size imported origin remains exactly the same
size after an anchor has been added.

The physical-file group starts separately from real closed full-origin and
imported stores. It incrementally writes actual whitespace after a small valid
checkpoint or origin to make a real 512 MiB file, then 512 MiB plus one byte.
For origin, it updates only the genuine anchor's SHA-256 to the streamed hash
of the padded origin, so a hash mismatch cannot masquerade as size detection.
Public offline inspection must accept the exact boundary, identify the overrun,
and preserve the bytes. This prospective group is physical-file inspection
coverage, not live-open evidence or an allocation-profile measurement.

Fixtures run serially and release each temporary directory before the next
group. Expected retained payload is roughly 512 MiB: one physical state file,
or a roughly 256 MiB origin and checkpoint together. Public decode/encode also
requires multiple large in-memory copies in this codec. Before execution, the
coordinator must release the compiler, establish disk headroom beyond that
payload and inspect memory pressure. A durable runner and targeted mutations
are still to be prepared. No constant reduction, sparse-file substitution or
source-identity refusal will stand in for these boundaries.
