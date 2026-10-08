# Post-run comparison revision

Preserved initial attempt: `39a92ca27a418712832ac0dcbbfbe4e32b3bca35`.
An independent reviewer compared it with the old standard; the student received
rationale and did not read or copy the old answer. Source-exposure limits are
recorded separately in SOURCE-EXPOSURE.md.

The revision removes two unnecessary full-context JSON copies per prompt.
Eligibility borrows state under Agent's mutex, and the response sequence is read
as a protected scalar. Rendering still takes an owned context snapshot, and the
public snapshot/event ownership guarantees are unchanged. This reduces avoidable
work as history grows without changing provider requests.

Malformed loaded data now reports a bounded static reason alongside the line:
reference kind/locator, sequence, timestamp, envelope payload, part fields, or
response state. Only a private validation error's static category is admitted
when adding line context. Arbitrary errors and original records are not echoed.
Four public regressions distinguish reference, unsolicited response, sequence
and payload failures and assert a private marker is absent from returned errors
and root diagnostics. The owner chain remains available throughout validation.

New WHY comments explain why rendering precedes request_sent (a render failure
must not consume ephemera) and why each observer gets its own copy after durable
application (one observer must not mutate another's notification or history).

All three modules passed format/vet/tests; formatting printed no paths. The
revised inherited grader passed100/100 after the coordinator corrected its
round-trip fixture to complete unanswered calls explicitly. The original95/100
failure remains in inherited-grade-before-fixture-fix.txt; production refusal
was not weakened to pass a contradictory fixture.

The files prefixed `initial-ch02-` are coordinator-produced receipts for the
initial checkpoint, not this revision:44 offline checks, passing control plus
10 targeted defects, corrected100/100 inherited grade, and targeted legacy
regression. The mutation receipt's source hashes were compared against every
corresponding file at39a92ca and matched. Source hashes for the revision and
its separately executed local/inherited checks use the `reviewed-` prefix.

No paid calls were repeated. These internal access/diagnostic/comment changes
do not change successful provider wire behavior; live receipts remain bound to
the initial implementation stages described in FEATURES.md. The independent reviewer accepted this revision after source inspection and the
four public diagnostic cases. Revised coordinator acceptance remains a separate
receipt; it is not inferred from the initial run.

Revised independent acceptance is now recorded in
ch02-reviewed-offline-acceptance.json:44/44 pass, with the tested binary hash.
ch02-reviewed-offline-mutations.json records the passing control and ten exact
detected defects. Every audit source hash matches the revision files at commit.
The coordinator also reported full root regression exit0 (513.863 seconds for
internal/grade), clean root vet and formatting. Its raw regression output is
retained in edition2-root-after-ch02-fixture.txt. Manuscript/editorial acceptance
remains separate from this validated code checkpoint.
