# Review: enforcing `system` and `model` in the Chapter 1 fake

Implements `book/brief-ch01-enforce-system-and-model.md`, which is the code half
of the author's ruling in `bc54c4f` on two of the three items escalated by
`book/review-ch01-grader-audit.md`.

Both rules are enforced. Both new mutants are caught. Both graders score 100.

## Scores

| Command | Score |
| --- | --- |
| `go run ./cmd/grade -ch 1 solutions/ch01` | **100/100 — PASS** |
| `go run ./cmd/grade -ch 2 solutions/ch02` | **100/100 — PASS** |
| `go test ./...` | all packages **ok** |

## What changed

**`internal/fakeanthropic/fake.go`** — two rules, both recorded as `wire`
violations on the request record:

- `model` must equal `ExpectedModel`, not merely be non-empty.
- `system` must be non-empty after `TrimSpace`, via the existing
  `MessagesRequest.SystemText()` helper, so both the string and the
  array-of-blocks spellings are accepted.

New exported const `ExpectedModel = "claude-fake-course-1"`, placed beside
`ExpectedAPIKey` and following its shape exactly: one exported const, the
harness sets the environment variable *from* it, one source of truth. The
existing `ExpectedAPIKey` comment already spells out this reasoning for the key;
the new comment states the model case in the same terms.

**`internal/grade/harness.go`** — `ANTHROPIC_MODEL` is now set from
`fakeanthropic.ExpectedModel` instead of a duplicated literal. See "the ch2
environment leak" below for the second line added here.

**`testdata/students/mutant/main.go`** — two new defects, plus the
unapplied-mutation guards described below.

**`internal/grade/grader_test.go`** — two new rows.

**`internal/grade/ch02_grader_test.go`** — one expectation corrected, and the
anchor assertion tightened.

`internal/grade/checks.go` is **untouched**: no check added, no points
re-divided. Both rules belong to `wire`, which already owns wire conformance and
already carries the hardcoded-key rule they are modelled on.

## Mutation suite: 14 mutants

Exact failing-check sets, as asserted by `TestMutationsAreCaught`:

| Mutation | Failing checks | Score |
| --- | --- | --- |
| `nosystem` | `wire` | 85/100 |
| `hardmodel` | `wire` | 85/100 |

Both fail `wire` and nothing else, which is what we want: the defect is scoped
to wire conformance and does not spuriously disturb the stdio, memory or usage
checks.

The rendered failures:

```
[FAIL]      Requests are well-formed Messages API calls  (0/15)
       request 1: system field is empty — the agent must send a system prompt
```

```
[FAIL]      Requests are well-formed Messages API calls  (0/15)
       request 1: model is "claude-3-5-sonnet-20241022", want
       "claude-fake-course-1" — the model must come from ANTHROPIC_MODEL,
       not a hardcoded string
```

The suite is now 14 rows: the `none` control, eleven pre-existing defects, and
these two.

## The anchor requirement, and where the brief's wording met a different mechanism

The brief's non-negotiable is that *every mutation must assert its anchor
matched exactly once*. That wording presumes text-substitution mutants. This
repo has **two different mutation mechanisms**, and the requirement lands
differently on each. I enforced it on both.

**Chapter 2's mutants are anchor-based** (`internal/grade/ch02_grader_test.go`,
`edit{file, pattern, repl}` compiled as a regexp). The applier asserted the
anchor matched *at least* once and then called `ReplaceAll`. I tightened it to
require **exactly one** site. Zero sites is the hazard the brief names — the
submission stays correct, the grader says 100, and the suite records "caught
nothing" about a defect that was never introduced. More than one site is the
mirror image: the mutant then fails for reasons beyond the defect under test, so
its failing-check set stops being evidence about that defect. Every existing
Chapter 2 mutation still matches exactly one site, so this tightening cost
nothing and closes the gap for future ones.

**Chapter 1's mutants are env-selected branches** (`COURSE_MUTATION` compared
inside `main.go`), in the style of `firstblock` and `hardkey` that the brief
pointed at. There is no anchor to count. The faithful analogue of the
requirement is "the selected defect was recognised and actually took effect",
and the program previously asserted neither. I added both halves:

- A `knownMutations` set. An unrecognised `COURSE_MUTATION` — a typo — now exits
  2 with a message, instead of running as a clean submission and scoring 100.
- A `mutating(name)` helper that every branch now routes through, counting
  reaches. If a mutation was selected but its branch was never reached, `main`
  exits 3 rather than exiting quietly.

**One deviation, stated plainly.** For the Chapter 1 mutants I assert the branch
fired *at least* once, not *exactly* once. "Exactly once" is false for this
mechanism: `firstblock` is inside `send`, which runs once per conversational
round, so it legitimately fires five times in a five-round session; `twocalls`
fires once per round by construction. Demanding exactly one would fail correct
mutants. The property that actually matters — the mutation was not silently
skipped — is fully captured by "recognised, and reached at least once". Where an
anchor genuinely exists (Chapter 2), the count is exactly one as specified.

**I verified both guards are sensitive rather than decorative**, by mutating the
guards themselves:

- Unknown name: `COURSE_MUTATION=nosystmm` → exit 2, `unknown COURSE_MUTATION`.
- Known name whose branch is unreachable: adding a `ghost` entry to
  `knownMutations` with no corresponding `mutating("ghost")` call → exit 3,
  `mutation "ghost" was selected but never applied`.
- Control (`none`) and unset → exit 0.

The first guard immediately earned its keep: the control row passes the literal
string `"none"`, which is not a defect name. The guard caught that on its first
run. `"none"` is now normalised to "no mutation" and documented as the control.

## Chapter 2 impact

**Chapter 2 needed no relaxation and no source change to pass.** Its Anthropic
renderer already sends a top-level `System` from a constant, and already
resolves the model through `pick("LLM_MODEL", "ANTHROPIC_MODEL",
"claude-sonnet-5")`, so the Chapter 1 phase supplies it. `solutions/ch02` scored
100/100 immediately after the rules went in.

Two consequences did surface, and neither was fixed by weakening a rule.

**1. The ch2 environment leak — a real bug the new rule exposed.**
`internal/grade/harness.go` also runs as Chapter 2's phase 1, and it builds the
child environment on top of `os.Environ()`. It pinned `ANTHROPIC_MODEL` but not
`LLM_MODEL`, which Chapter 2's submission *prefers*. So an `LLM_MODEL` exported
in the grader's own shell silently won:

```
LLM_MODEL=some-other-model go run ./cmd/grade -ch 2 solutions/ch02   →  75/100
```

This fragility pre-existed my change; the old non-empty model check simply could
not see it. The rule is right and the harness was wrong. Fixed in the harness
environment, which is what the brief authorises — `LLM_MODEL` is now pinned to
`fakeanthropic.ExpectedModel` alongside `ANTHROPIC_MODEL`. This follows the
precedent already in that function, which pins `ANTHROPIC_API_URL` as a
belt-and-braces alias for the base URL for exactly the same reason. It also
keeps `ExpectedModel` as the single source of truth regardless of which spelling
the submission reads. The hostile-environment run now scores 100/100.

Worth noting as a teaching point: this is the failure the audit format is for. A
weak check did not merely fail to catch student errors — it concealed a defect
in the grader's own harness.

**2. A Chapter 2 mutation expectation was strengthened, not relaxed.** The ch2
mutant `system-prompt-omitted` previously failed only `seam-render`. It now also
fails `ch1parity`. That is correct and I updated the expectation to match: a
Chapter 2 submission that sends no system prompt genuinely violates Chapter 1
wire conformance, and `ch1parity` re-runs the Chapter 1 checks unchanged. The
defect is now caught by two independent checks instead of one. The test's own
guidance — "when a mutation expectation misses, ask FIRST whether the grader is
right" — applies, and the answer is that the grader is right.

## The F3 `type` filter stays ungraded

Untouched, per the ruling that it is a deliberate gap. I did not find an honest
way to grade it and am not proposing one.

The reason it resists honest grading is worth recording precisely, because the
tempting workaround is easy to reach for. Dropping the `type == "text"` filter
is not merely hard to observe — on the real wire it is **behaviourally
identical**. No real Anthropic content block carries a `text` field unless its
type *is* `text`, so `for _, b := range content { out += b.Text }` produces
exactly the same string whether or not the submission filters on type. There is
no defect to detect, which is why audit item 2 scored 100 before and after the
deletion.

To make the filter observable the fake would have to send a block whose type is
not `text` but which carries a populated `text` field — a shape the Messages API
never produces. The grader would then be scoring a student against a fiction,
and worse, rewarding the belief that such blocks exist. A student who
internalised that would write defensive code for a case the API cannot emit.
That is the falsehood the ruling protects against, and it is the reason the gap
is deliberate rather than merely unfinished.

`firstblock` already covers the honest neighbouring property, and covers it
well: a client that reads only `content[0].text` and drops the rest genuinely
does break on the real wire, and is caught (audit item 1, 100 → 65).

## Weighting: left alone, as instructed

Both rules live in `wire` and I did not re-divide the 100 points. For the record,
and with no change made: I think `wire` is the right home for both. `wire` is
already "requests are well-formed Messages API calls", and "names the model you
were told to use" and "sends a system prompt" are both statements about the
well-formedness of the request body, in the same family as the hardcoded-key
rule that already lives there. Splitting either into its own check would also
re-open the arithmetic on a 100 that currently divides cleanly. Weighting
remains the author's call.

## Verification performed

- `go run ./cmd/grade -ch 1 solutions/ch01` → 100/100.
- `go run ./cmd/grade -ch 2 solutions/ch02` → 100/100.
- `LLM_MODEL=some-other-model go run ./cmd/grade -ch 2 solutions/ch02` → 100/100
  (75/100 before the harness fix).
- `go test ./...` → all packages ok; 14/14 Chapter 1 mutants, all Chapter 2
  mutants, reference solutions pass.
- Both new guards mutated and observed to fire (exit 2 and exit 3).
- `internal/grade/checks.go` unchanged — verified via `git diff --stat`.
