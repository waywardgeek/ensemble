# Trusted Chapter 1 grading on macOS

The ordinary Agent command sandbox denies networking, including loopback. Do not
run a student's executable unrestricted to work around that boundary. This
host-side launcher lets the **unchanged** Chapter 1 grader own its HTTP fake,
while the candidate retains filesystem confinement and gets outbound TCP access
only to the fake's assigned host-local port. It grants no listening, other-port,
UDP or external-host access. Ordinary Agent tools and saved child policy do not
change. This is not a new model-facing grading tool; parent-mediated grading must
be disclosed in trial evidence.

`Sandbox.PrepareGradedProcess` reuses the normal process profile and environment
allowlist. The CLI requires the grader's public fake key, a nonblank model, and a
literal `http://127.0.0.1:PORT` or `http://[::1]:PORT` origin with no credentials,
path, query or fragment. It never accepts a real provider key. SDK access is
read-only; the candidate is resolved inside the writable workspace. The process
is replaced with `sandbox-exec`, then the candidate: PID, stdio and exit status
remain the grader's, rather than belonging to a detached child behind a wrapper.
For an IPv4 fake URL the launcher sets `DOTNET_SYSTEM_NET_DISABLEIPV6=1`:
macOS denied IPv4-mapped IPv6 sockets under the otherwise passing exact-port
rule. This is a disclosed runtime-environment adaptation, not a student-source
repair or broader network permission. An IPv6 literal does not get that setting.

## Prepare and run

Use an existing .NET 10 SDK and cached dependencies; do not download implicitly.
Build `agent/Examples/GradedCandidate/GradedCandidate.csproj` into a **trusted
output directory outside the student workspace**. Build `cmd/grade-edition3`
with the existing Go toolchain. The root README describes offline build setup.
For example, after building both and publishing the candidate in its workspace:

```sh
python3 scripts/confined-grading/prepare-wrapper.py \
  /absolute/sdk/dotnet /trusted/build/GradedCandidate.dll \
  /absolute/student/workspace /absolute/student/workspace/bin/ch01 \
  /absolute/sdk /trusted/new-candidate-wrapper
/trusted/grade-edition3 -ch 1 -json /trusted/new-candidate-wrapper
```

The helper requires absolute paths, refuses trusted components inside the
workspace, and exclusively creates the wrapper. Keep the wrapper and the entire
launcher build host-owned and outside student write authority. Its environment
is cleared before starting the .NET launcher, retaining only the grader's three
fake-provider variables. Candidate caches/output must stay inside the workspace;
no host cache or credential directory is imported. The working directory is the
workspace, not the grader's temporary directory.

Require both `passed:true` and exit 0. Allow at least 360 seconds for the two
62-second lifetime probes and the rest of the gate. A fake-provider grade does
not complete the separate real-provider exercise or inspect source architecture.
Never give the candidate access to reference solutions or mutable grader code.

## Verify the launcher

```sh
python3 scripts/confined-grading/test-launcher.py \
  /absolute/sdk/dotnet /trusted/build/GradedCandidate.dll
```

The test uses real disposable HTTP servers, a candidate and its descendant. It
checks the assigned port, denial of another local port and nonlocal address,
listening/UDP denial, outside reads/writes and symlink escapes, a read-only import,
secret-environment removal, workspace writes, stdin, exact PID and exit 17.
Malformed origins, a non-grader key and an outside candidate path must refuse
before execution. The native process suite also checks preparation and ordinary
network denial. Run the unchanged grader against a known fixture separately;
launcher tests alone do not certify grader compatibility.

## Why not general loopback permission?

On this host, native numeric-address predicates were rejected. The sandbox's
`localhost` selector also matched this host's non-loopback interface. Requiring a
remote peer in an inbound rule prevented `listen`, because that operation has no
peer yet. The grading design therefore grants **no inbound rule** and scopes the
outbound exception to one trusted grader port. macOS still enforces host-local
port scope, not an exact interface or private network namespace: another local
interface using that same port is not distinguished. Do not describe this as a
general-purpose network isolation mechanism or resource/CPU quota. Process-group
cleanup and the grader's watchdog remain the host runner's responsibility.

The initial reference validation passed the unchanged Chapter 1 grader at
100/100 with exit 0, including both 62-second delays. The native suite passed
400 tests, and the real launcher checks passed against the final build. These
are harness/reference results, not a student grade or live-provider acceptance.
