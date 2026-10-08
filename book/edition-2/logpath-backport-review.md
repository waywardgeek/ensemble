# Independent review of the Chapter 2 and 3 LogPath corrections

Accepted for the scoped revision checkpoints, 2026-10-07. No material finding
remains in these two patches. This review is independent of their coordinator
author, and makes no claim of a new cold-student implementation.

Reviewed identities:

| Chapter | Isolated correction | Accepted predecessor |
|---|---|---|
| 2 | `9fed66a8c8f2e25081cb3b9ee6e5eba2bef5cc17` | `ad0d80e33a3a2857e8e0887117d9b099f1a4786d` |
| 3 | `ea4bceac5f5b43774050559861845595a89a7eaa` | `1a61e1f2487cdc94ee65bd0e1593065cc1e95f45` |

The authority is the new Chapter 2 lifetime-history rule at §2.3: a public
configuration update cannot change the destination of an Agent's existing
writer. The reviewer reloaded the full coding skill before testing, inspected
both complete changed tests and production diffs, surrounding constructor and
configuration ownership, and each correction's receipts. Main was neither
edited nor used as an implementation source for this review.

Both patches compare the owned incoming LogPath with the stored path under
the existing configuration mutex before assignment. The existing operation
lock still excludes active requests. Returning from the guard leaves all
configuration fields intact; the deferred unlock covers both outcomes. The
patch does not reopen, rename or replace the writer. Its comment explains why
the apparently configurable field is fixed. Chapter 3 retains its separate
workspace, builtin and declaration checks.

The public regression covers a different destination and removal of the
destination, requiring rejection, full configuration equality and absence of
a second log. Its positive control changes the model while retaining the path,
observes that model in a real local HTTP request, and checks the accepted answer
in the original log. This is useful behavioral coverage rather than a test of
the guard's spelling. The retained original-production controls report exactly
the two rejection-subtest failures, plus their parent test, while the same-path
positive passes. These controls therefore establish the intended missing
behavior rather than an unrelated setup failure.

Independently rerun in both isolated source roots:

```sh
go test -race . -run '^TestConfigurationPreservesLogIdentity$' -count=1 -v
gofmt -l ensemble.go config_identity_test.go
```

Both tests passed, including all three subtests, and formatting output was
empty. The reviewer independently matched both changed-file SHA-256 values
against each `repair.json`. The coordinator's retained receipts, inspected
but not broadly rerun by this reviewer, show Chapter 2's three modules and
Chapter 3's four modules passing vet/tests, main race checks, inherited scores
of 100/100, offline counts 44/39 and human-interface counts 42/45. Those human
checks are local fixture checks, not new paid-provider demonstrations.

Retaining the earlier live receipts under their original identities is sound
for this correction. The changed outcome is a local configuration refusal;
the successful same-path request and persistence path remain intact. No new
provider capability, wire mapping or tool behavior is claimed. The review
does not reinterpret old live runs as runs of the revised executable, or
claim a comprehensive new review of these earlier chapters. Export manifests
and immutable revision tags remain the coordinator's checkpoint work.
