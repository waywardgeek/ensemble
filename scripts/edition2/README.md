# Second-edition acceptance tools

Run from the course repository root. These add evidence to the inherited
grader; they neither replace its score nor establish live usability.

The solution is a separate Go module, so build inside it:

```sh
(cd solutions/edition-2/ch01 && go build -o /tmp/ensemble-ed2-ch01 ./cmd)
python3 scripts/edition2/accept_ch01.py /tmp/ensemble-ed2-ch01
python3 scripts/edition2/audit_ch01.py solutions/edition-2/ch01
go run ./scripts/edition2/packagecheck solutions/edition-2/ch01
```

`accept_ch01.py` runs the actual executable against loopback fixtures with
fictional credentials. It checks 23 CLI properties, including a stalled
request. The 90-second harness limit is an infrastructure bound, not a
prescribed client timeout. Use `--skip-timeout` only for unrelated fast
mutations; the ordinary acceptance run must include the timeout.

`audit_ch01.py` applies eleven deliberate defects in disposable copies and
compares exact failing-check sets, with a passing control. It targets the new
reference snapshot's current source anchors; a missing or duplicate anchor
fails rather than silently skipping a mutant. The acceptor itself neither
imports student code nor depends on internal names. This is a partial audit
of the contract; the recorded public-API/parser tests and independent review
supply additional evidence. A build failure never counts as defect detection.

`packagecheck` parses production Go source and discovers all packages in the
core module. It checks implementation imports, executable access through the
public API, and behavior placed in `internal/common`. It skips separate
nested modules; check their public boundaries independently. Shared-type
methods for JSON/text/String/Error dispatch and their referenced helpers are
allowed after signature and symbol resolution with `go/types`. Resolve invalid
common imports before its behavior analysis can run. Their semantics still
require review. A clean report does not
prove interface back-pointers, actual service ownership, logger reachability,
absence of shared mutable state, or a single composition root. Review those
relationships and exercise them through the public library.

Both the student and reviewer must still follow the chapter procedure,
including live feature demonstrations and the post-run comparison with the
first-edition standard. Do not weaken old graders or edit old solutions for
these audits. Preserve source hashes and distinguish paid-model receipts
from local fault injection.

## Chapter 2 offline checks

```sh
(cd solutions/edition-2/ch02 && go build -o /tmp/ensemble-ed2-ch02 ./cmd)
python3 scripts/edition2/accept_ch02.py /tmp/ensemble-ed2-ch02
python3 scripts/edition2/audit_ch02.py solutions/edition-2/ch02
```

`accept_ch02.py` exercises the public offline executable with independent
contract fixtures. It checks versioned log validation, valid edge cases,
three wire projections, result redaction and pairing, deferred-input order,
and refusal to render unresolved calls or unsupported references. It checks
that commands leave the source log unchanged and make no requests to the
configured loopback trap. This is not a network sandbox or a live-provider
test. Rendering is checked as JSON, allowing harmless formatting differences.

`audit_ch02.py` freezes a source copy, then builds and checks a passing control
and ten deliberate defects. Each must produce exactly the expected failed
checks. It records source hashes and excludes repository metadata and evidence
from the disposable copies. This scoped audit does not establish every event
transition, public Go ownership, provider parsing, GUI behavior, or live
usability; those require the separate chapter acceptance and review evidence.
