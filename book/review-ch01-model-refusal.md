# Review: ch1 reference refuses when `ANTHROPIC_MODEL` is unset

Carried out per `book/brief-ch01-model-refusal.md`. One file changed:
`solutions/ch01/main.go`. The grader was not touched, no check was added, no
points moved.

## Diff summary

`solutions/ch01/main.go`, +8 / −3, in two places:

1. **Header doc comment.** `ANTHROPIC_MODEL optional, defaults to a current
   Claude model` became `required`, with two lines saying why there is
   deliberately no default.
2. **`newClient()`.** The `if model == "" { model = "claude-sonnet-5" }` fallback
   became an error return. The `curl /v1/models` hint that used to sit above the
   default was kept and moved into the refusal branch, so the reader is told
   where the real answer lives at the moment they hit the wall.

The `os.Getenv("ANTHROPIC_MODEL")` call site is byte-identical to before. That
was deliberate — see "Mutation rig" below.

With the variable set, behaviour is unchanged: no other line was touched, and
the control case still reaches the network (proof below).

## Line count

| | `wc -l solutions/ch01/main.go` |
|---|---|
| BEFORE | **273** |
| AFTER | **278** |

The chapter's "273 lines" needs to become **278**.

## The stderr line

Exactly, including the `ch01: ` prefix that `main()` already applied to the
sibling `ANTHROPIC_API_KEY` error:

```
ch01: ANTHROPIC_MODEL is not set; ask GET /v1/models which models exist
```

Verified in **both** modes (grader stdio and `./ch01 chat`): exit code `1`,
**0 bytes on stdout**, exactly **1 line** on stderr.

Verified it refuses *before any HTTP request*, not merely before a successful
one: with `ANTHROPIC_BASE_URL=http://127.0.0.1:9` (nothing listening) and the
variable unset, the program emits the line above. With the variable *set* and
the same dead port, it instead reports
`dial tcp 127.0.0.1:9: connect: connection refused`. The refusal therefore
happens strictly earlier than the first dial, which is the property the brief
asked for rather than the mechanism.

## Tripwires

**1. Grader — PASS.** `make grade-dir DIR=solutions/ch01` = **100/100**, same as
before the change. The request log shows `model="claude-fake-course-1"`,
i.e. the harness-supplied value, confirming the reference never wants a default.

**2. Mutation rig — PASS, same mutant set.** Both scripts under
`scripts/grader-audit/` were run before *and* after the change and compared
mechanically, not by eye.

- `ch01_promises.py` — 15 mutants. `results.json` compared field-by-field
  against a pre-change snapshot: **mutant set identical, 0 rows changed.**
- `ch01_reachability.py` — 9 mutants, all killed, **no `ANCHOR-FAIL`**, table
  identical before and after (baseline taken by restoring the original file,
  then restoring mine with a verified checksum).

Note for the record: **neither script has an exit code or a verdict** — both
always exit 0 and print a table plus `results.json`. "The rig passes" can only
mean "the table is unchanged", so that is what was checked, exactly.

The anchor risk, since it is not obvious: `ch01_promises.py` builds mutants with
a `sub1` helper that **asserts its regex matches exactly once**. The
`model-hardcoded` mutant anchors on the literal string
`os.Getenv("ANTHROPIC_MODEL")`. Refactoring that call into a helper (say
`env("ANTHROPIC_MODEL")`) would have turned that mutant into an `ANCHOR-FAIL`
and silently removed a promise from the audit. It was left verbatim for that
reason. `model-hardcoded` still scores 85 and is still killed: the mutant makes
the variable non-empty, so it flows past the new refusal and is caught on the
wire exactly as before.

**3. `grep -rn 'claude-' solutions/ch01/` — PASS, no match.** No model ID
remains in Go source. (`runChat`'s `claude> ` prompt does not match and is not a
model ID.) `gofmt -l` clean, `go vet` clean, builds clean.

**4. Harness env var — CONFIRMED.** `internal/grade/harness.go:104` sets
`"ANTHROPIC_MODEL=" + fakeanthropic.ExpectedModel` inside the
`cmd.Env = append(os.Environ(), ...)` block beginning at line 101. It is
appended *after* `os.Environ()`, so it wins over any ambient value in the
developer's shell — the reference cannot accidentally pass because of a stray
export. Every ch1 run goes through this one path, so the refusal is unreachable
under the harness and costs no points.

## Anything the brief got wrong

Nothing wrong, but three things worth the author's attention.

**1. This makes ch1 a deliberate exception to P10, which should probably say
so.** P10 (`course-policy.md:416-422`) states that the solution's built-in
default is deliberately not the pinned table, that **"the ch2 and ch3 solutions
default to `claude-sonnet-5`"**, and that this **"is not drift and should not be
'fixed'."** The ruling here is compatible with the letter of that — P10 names
ch2 and ch3 only, and I changed neither — but not obviously with its spirit,
since the rationale it gives for defaults is general. A reader who compares
`solutions/ch01` (refuses) with `solutions/ch02` (defaults to `claude-sonnet-5`)
sees an inconsistency with nothing explaining it. Recommend a sentence in P10
recording that ch1 is exempt *because* ch1 is the chapter whose subject is the
remembered-model-ID failure mode. I did not edit `course-policy.md`: P10 says
rulings are amended "with a date and a reason", which is the author's act, not
the coder's.

**2. The rig has a pre-existing survivor, unrelated to this change.** In
`ch01_promises.py`, `blocks-no-type-filter` scores **100** — it strips the
`block.Type == "text"` filter and the grader does not notice. By P9's own
standard that is a promise the chapter makes (L116, "concatenate the *text*
blocks") which no check enforces. It survives identically before and after, so
it is not a regression and was out of scope here, but it is a real gap and the
brief's "the rig still passes" would not have caught it either way.

**3. The ch01 solution egg is still unplanted.** `solutions/ch01` contains no
egg, while ch02–ch04 each carry one in `part.go`. This is already tracked in the
P7 placement ledger (`course-policy.md:240`, "solution egg not planted"), so it
is a known outstanding item rather than a finding — noted only because this
change touched the file and did not address it.
