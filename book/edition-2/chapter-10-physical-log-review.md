# Chapter 10 physical log-file boundary

The previously disk-blocked physical 1 GiB regular-file check now passes on
immutable `57d4aac3d26edcb78dc8d9fd27d7d0e1cce0ebf4`. Bill freed disk space and
the coordinator released this stage. This is separate from the earlier generated
Reader proof and does not change that earlier receipt's scope.

The [source-bound receipt](checkpoint-evidence/ch10-physical-log-57d4aac.json)
binds 159 runtime/API/asset files, checker sources, the two compiled executable
hashes, and every command with times, output and free-space observations.
Formatting printed nothing; overlay module vet, full unoverlaid module tests,
and compilation succeeded before fixture execution.

The fixture first opens and closes a genuine small session, then successfully
inspects that real parent. It extends its events.log by sequential actual writes
of whitespace records no larger than 1 MiB. No sparse file, truncate-to-length,
mocked size or reduced limit is used. The original initializer, checkpoint and
semantic history remain genuine. These blank records obey the inherited reader
rules and allow the file bound to bind independently of record/event limits.

| Control | Observed result |
|---|---|
| Actual regular log of 1,073,741,824 bytes | Public InspectSession accepts. Session identity, boundary, events and semantic snapshot match the small parent. OS allocated-block size also reports 1,073,741,824 bytes. |
| Same log plus one physically written byte | Public InspectSession refuses with session_corrupt and `session log exceeds 1 GiB`. Log and checkpoint hashes remain unchanged across inspection. |
| Mutation increasing all three Stream file-bound expressions by one | Its exact-size parent passes first, then its physical one-over file is incorrectly accepted and the targeted `physical log one-over limit accepted` assertion fails. |

This is public offline regular-file inspection coverage. It does not claim live
opening at this bound, a new append-atomicity guarantee, or complete Chapter 10
acceptance. No compiler/hash/setup refusal was counted as mutation detection.
The standard positive and targeted mutation both completed on their first run.

Each large execution required 1 GiB plus 512 MiB headroom after compilation;
actual available space was over 7.7 GiB immediately before each allocation.
Only one large payload existed at a time. The positive took 3.77 seconds with
20,283,392-byte maximum RSS; the mutation took 3.26 seconds with 20,021,248-byte
maximum RSS. Both reported zero swaps. GOMEMLIMIT=1GiB and GOGC=25 were runtime
settings, not claims of a hard process memory ceiling.

Both 1 GiB + 1 payloads were removed by their owned Go test cleanup, including
the intentional assertion failure. Thus approximately 1 GiB was reclaimed after
each of two sequential runs; these are payload sizes, not a claimed net change
in shared disk free space. The runner's finally block also removes its own payload
root on success/failure while retaining small adapters and review binaries.

The [checker and cleanup receipt](checkpoint-evidence/ch10-physical-log-checker-cleanup-57d4aac.json)
records three passing receipt controls and a read-only audit of this reviewer's
scratch prefixes, including earlier large-state and mutation workspaces. There
are no residual files larger than 64 MiB. Retained binaries are about 12 MB each;
source extractions, user files and compact evidence remain untouched. Final free
space in that audit was 8,345,034,752 bytes. No broad deletion or cache cleanup
was performed. No provider or credentials were used. The compiler was released
after all Go work completed.
