# Independent GUI Digest authentication feature

October 26, 2026: selectively port reviewed Digest authentication from
`15084c1e` into `feature/gui-digest-auth`, prepared as one independent feature
commit based on `origin/main` at `123d6017`. The original main-auth commit
`15084c1e` remains unchanged.

## Scope and teaching

The feature keeps upstream's hardcoded `127.0.0.1` launcher. It excludes the
separate `--host`/LAN launch feature and CPU profiling, including their startup
policy and profile-output collision claims. Timestamp, theme, rename,
Wake-on-LAN and snippet additions are also excluded. Existing upstream prose
about ordinary preferences remains intact.

Authentication is optional until local initialization. Existing credentials
protect HTTP assets and WebSocket upgrades, including loopback. The API uses
`ensemble auth-init`, `ensemble auth-reset`, setup `--file`, launch `--gui-auth-file`, and
`Server.ConfigureDigestAuth(path)` before serving for custom embedders. Default
storage is `cr/gui-auth.json` or the configured root data directory. The README
explains restart, owner-only storage, password-equivalent HA1, SHA-256/qop=auth,
replay protection and bounded per-source backoff. Digest assumes an attacker
who can connect but cannot intercept or modify traffic; it supplies no TLS.

Chapter 8 gains only a short current optional note before “What is not graded
and why.” Its historical narrative, exercise and results are preserved. Only
assembled Markdown is regenerated; Bill's preface remains verbatim.

## Evidence and remaining checks

The record in `15084c1e:docs/edition-3-notes/gui-digest-auth.md` reports independent
review, passing GUI short tests/vet/focused race tests, native Firefox Digest
and host tests, an RFC known-answer check, and targeted guard and persistence
mutations. These are reused main-auth evidence, not runs of this independent
feature. Its host, profiling and unrelated browser-feature results do not
validate this branch. The recorded real-model attempt failed with
`provider transport failed`; it is not successful live-model evidence.
That live-provider attempt was not rerun on this branch.

The coordinator reports passing branch GUI `go test ./... -short`,
`go vet ./...` and focused Digest/CLI tests under `-race`.

The initial native Firefox run passed six of seven subtests. Authentication,
restart, restored first reply and a new authenticated second chat worked; the
final assertion that the first reply remained visible after that second chat
failed. Independent native Firefox checks on pristine `123d6017` confirmed the
inherited behavior: without restart the first reply survives the second turn
(count 1); after process restart, history replay and a new turn, it disappears
(count 0), while the second reply is present (count 1) and provider calls total 2.

The auth-specific test removes only that inherited post-restart retention
assertion. It still verifies the restored first reply before the second chat,
the new authenticated chat, unchanged credentials and unauthenticated HTTP 401
responses. No unrelated runtime fix is included. The adjusted native Firefox
rerun, `node --test digest-auth.test.mjs` with native Firefox 157.0, passes
8/8 tests with zero skipped. The coordinator confirms the final GUI short
suite, vet and focused Digest/CLI race checks also pass. Volta's bounded
independent review accepts the port with no blockers, including feature
independence, preserved safeguards, documentation and the baseline-backed test
adjustment. At that port checkpoint, main was unchanged at `15084c1e`.

Documentation prose lint passes with soft style warnings. Assembly matches
`scripts/assemble-book.sh` byte for byte; its only change from the baseline is
the optional Chapter 8 note. Removing that insertion restores the historical
chapter exactly, and both preface source and assembled inclusion remain
byte-identical. Owned-path whitespace checks pass.

## CLI simplification, October 26

At the user's request, setup is now `ensemble auth-init` and rotation is
`ensemble auth-reset`. The old `gui-auth` syntax is rejected before startup.
Initialization still generates once and refuses an existing file; reset rotates
an existing valid file. Options, `gui-auth.json` and `--gui-auth-file` are unchanged.
The coordinator reports passing full GUI short tests, vet and focused Digest/CLI
race checks in both trees. Native Firefox passes 15/15 on main (auth and host)
and 8/8 on the independent feature (auth), with zero skips. Two focused Go
overlay mutations fail their intended checks: incorrect `Args[2:]` dispatch
fails explicit-command setup; removing the legacy-token guard fails the explicit
usage-before-provider assertion. Unmutated tests pass and scratch mutants were
removed. Darwin independently accepts the command rename in both trees with no
blockers; feature independence and the existing credential contract are preserved.
