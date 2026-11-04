# Gateway complete-state persistence

## Boundary

Reviewed and locally verified source candidate, personally implemented by Cadence under Bill's
bounded Gateway authorization. No deployment, service restart, push, provider
request or new worker. Gateway teaching has moved to [the Puffin book](../gateway.md), without a new
chapter number. Historical chapter flags and endpoint directories retain their
original meaning.

## Contract and ownership

The Router retains its live identity/group/queue/session graph. Each durable
change creates a temporary immutable version-1 complete-state candidate, validates
it, saves the previous validated primary as backup, flushes and replaces the
primary, then publishes the change to the existing live objects. A lifetime
exclusive directory lease prevents overlapping Router processes. Durable facts
include bootstrap public identity/authority, runtime policy, children, queues and
receipt positions. No Gateway event journal, periodic saver or teardown save.
`Save()` is a compatibility health check. Private keys remain separate.

Persistence or post-commit publication failure poisons admission. Removing an
obstruction does not authorize retry from old live state. Joined teardown must
preserve committed evidence even if live publication failed. Network I/O remains
outside the policy gate. A valid reconnect receipt is saved before replacement
gets current-session authority; normal receipts are saved before queue removal.

Normal startup validates the primary and refuses corruption or missing-primary
with retained backup. Bootstrap/key conflicts refuse implicit authority migration.
Explicit migration preserves legacy originals and rejects orphan queues. Explicit
backup recovery validates first, preserves an existing primary under a unique
evidence filename, replaces the primary and requires a separate normal start.
Both host modes exit without listening. Recovery is rollback: it can omit the
latest accepted mutation or replay an acknowledged delivery.

## Evidence collected so far

- Router suite: 16 groups pass, including seven real-WebSocket session groups.
  Ordinary and reconnect receipt-save failures preserve receipt/primary state and
  refuse new admission. Failed reconnect does not replace current authority.
- Independent-process test: a second process cannot acquire the live directory;
  killing the writer after acceptance preserves the exact delivery on restart.
- Fault tests: backup-write obstruction preserves live state/primary and poisons
  admission; an injected publication exception after real primary replacement
  preserves the committed candidate through teardown and restart.
- Public client suite: six local groups pass. No Go-service mode was invoked.
- Public persistence grader passes against the RouterTests candidate output:
  real authenticated sends, exact crash replay, receipt barrier/removal, malformed
  refusal, explicit backup rollback/evidence, filesystem save failure and explicit
  legacy migration. This is a distinct pass/fail scope, not old rubric points.
- Isolated deliberately-unsaved candidate builds and fails that grader specifically
  at `Admission RPC completed without the exact durable delivery`.
- Real GatewayHost command-mode check passes: conflicting flags before effects,
  key-only output, explicit migration, original preservation, malformed refusal,
  recovery and invalid-backup refusal. Invalid listen address deliberately proves
  offline modes do not enter listener setup.
- Adapter fixture updated to read complete snapshots without `Save()` and attach
  its borrowed diagnostic writer to the actual Agent. All six groups pass.
- Host, RouterTests, ClientTests and AdapterTests builds succeeded without warnings.
  Public grader also passes against the final Host output. Hard prose lint,
  changed-document links, Python syntax and diff checks pass.

Local evidence is under `/private/tmp/cadence-gateway-upgrade/`: router-tests-final.log,
client-tests.log, adapter-tests-3.log, no-save-grade.json, host-modes-final.log,
public-host-final.json and prose.log. These are local run evidence, not published artifacts.

## Reproduction

Build the candidate using the repository's explicit SDK and dependency policy.
No grader downloads a package. Then run:

```sh
python3 scripts/grade-gateway-persistence.py \
  --dotnet /absolute/path/to/dotnet \
  --candidate-directory /absolute/path/to/GatewayHost/output
python3 scripts/gateway-persistence-grader/host_modes.py \
  --dotnet /absolute/path/to/dotnet \
  --host /absolute/path/to/GatewayHost/output/GatewayHost.dll
```

The external consumer is a separate assembly referencing candidate binaries, not
candidate source or internal APIs. The launcher isolates home/output, disables
NuGet sources and terminates the test process group on failure or timeout.
Native libsodium comes from the supplied candidate for the current host platform.

## Lessons and remaining limits

A directory named for a package is not proof of a usable NuGet cache. Explicit
restore from the retained complete package cache fixed setup; no download occurred.
Likewise, copying `runtimes/` into a consumer does not add native resolution to its
dependency manifest. The grader places the candidate's matching native library
beside its executable, without substituting another library.

Final review caught an empty-identity migration edge case: without per-identity
reads, an obstructed legacy queue directory could be ignored. Validate the queue
root independently of the number of identities; a focused regression covers it.

The adapter's old diagnostic assertion observed Ensemble while the bridge logs
through Agent. The corrected fixture attaches its writer to Agent. This was not
established as a scheduling race, and production diagnostics were not changed.

Atomic replacement and file flush support the tested process-crash boundary, not
universal power-loss durability. No directory fsync, physical power-removal test,
filesystem security audit, Linux execution or Pi storage-wear measurement is
claimed. Same-disk backup is not disaster recovery. Stable state IDs/receipts do
not make recipient model work or external side effects exactly once.

The appropriate adjacent regressions were the client and real Agent adapter
suites; no core runtime implementation changed, and unrelated GUI/provider suites
were not rerun for freshness. Historical Go endpoints, grader flags and evidence
remain unchanged. Prior assembled manuscript/PDF/EPUB outputs were not regenerated.
The candidate is ready for Bill's deployment decision; installation, service
restart, live migration and external publication remain separate authorization.
