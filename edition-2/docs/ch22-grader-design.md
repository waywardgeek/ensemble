# ch22 grader design

Checks and point values are fixed by `book/ch22-coder-brief.md` (7 checks,
100 points). This document records only what took judgement: how the thesis
check is made framework-blind, and the measured facts about the reference
tree that every check must pass.

## The thesis check: `back-pointer-chain` (20 pts)

**It must not become `detectLogf`.** Chapter 5's original check grepped for
the literal strings "Logf" and "Host" and reported a structure it had never
inspected; it was replaced in commit 1e5c456. A check that names
`common.Agent` would fail a student who named the type `common.Parent`, and
would pass a tree that merely contained the right words. Neither is
acceptable.

### Structural clauses (no type name appears in any of them)

Reuse `internal/grade/astscan.go`, written for the chapter 5 repair:

    scanTree(dir)                      -> *astScan
    s.dispatchStructFor(iface)         -> the struct beside iface that
                                          carries it and is passed to
                                          another package

Three clauses:

1. **FIRST LINK.** The dispatch struct has at least one field whose type is
   an interface declared in the same package. (Embedded counts.)

2. **SECOND LINK.** At least one interface reachable from clause 1 has a
   method whose RESULT is another interface declared in that same package.
   This is what makes it a *chain* rather than a single back-pointer: the
   parent can hand you its own parent, or a further capability, without the
   dispatch struct growing a field.

3. **NO CLOSURES.** The dispatch struct has no func-typed fields. This is
   the "none takes a closure where an interface exists" half of the brief's
   wording, and it is the distinction the whole chapter turns on: a closure
   answers exactly one question forever, a back-pointer answers the next one
   too.

### Why the reference tree passes (measured, agent/internal/common)

    type Call struct {
        Agent                 // embedded interface      -> clause 1
        Engine Engine         // interface-typed field   -> clause 1
        Job    JobHandle
        Jobs   JobManager
        Limits Limits
        Events []Event
        DeferFinish bool
    }                         // NO func-typed fields    -> clause 3

    type Engine interface {
        Agent() Agent         // interface result        -> clause 2
        Model() string
        Pricing() Pricing
        Usage() UsageSource   // interface result        -> clause 2
    }

Two independent witnesses for clause 2, which is worth keeping: a student
who exposes only one of them still passes.

### Sharpness requirement

Write the name-blind fixture FIRST. The chapter 5 repair shipped a bug
caught only this way: the first draft identified the logging method by
`strings.Contains(name, "log")`, which is a vocabulary anchor hiding inside
a structural check. The fixture that renamed `Logf` to `Printf` caught it,
and the fix was to match on SIGNATURE instead (format string, variadic, zero
results). Expect the same class of mistake here and test for it.

## `reaches-through-the-chain` (15 pts)

Anti-cheat for `agent-status-tool`: a student could report the model from a
package-level global. Assert the tools package does not import the engine
package.

Do this with the toolchain, not with grep. Measured on the reference tree:

    go list -deps ./internal/tools/   ->  exactly two packages from this
                                          module: common and tools

Grep over import blocks would miss a transitive import; `go list -deps` does
not.

## `per-model-cost` (15 pts)

The one check with an observable wrong number. Drive two models with
different prices and different token counts, then assert:

    reported == sum(tokens_model_N * price_model_N)
    reported != total_tokens * price_of_current_model

Assert BOTH. The second clause is what makes it a bug test rather than an
arithmetic test, and it must be shown to differ: if the two formulas happen
to round to the same string for the chosen inputs, the test proves nothing.
Pick prices an order of magnitude apart.

Reference models with real prices in the tree: `gemini-3.8-flash` (cheap),
`gpt-6-astra` (~10x), `gpt-ch19-course` (deliberately unpriced -- an
unpriced model must be reported as a LOWER BOUND, not billed at zero).

Float equality flakes here: `CostByModel` sums over a MAP, so addition order
varies per run and float addition is not associative. Use a tolerance and
say why in the comment.

## `agent-status-tool` (20 pts)

Drive a conversation through the fake vendor, switch models mid-session,
call the tool, verify the five reported values against known inputs.

Assert the FRAME, not the prose inside it. Chapter 21 shipped a check on the
literal text "404" that passed agents which had dropped `is_error`; it was
rewritten to assert the flag. Here that means asserting the tool returns
values that match computed expectations, not that its output contains a
particular word.

The grader supplies its own skill file, which makes it deterministic and
structurally blind: chapter 21's live probe caught `ensemble/SKILL.md`
omitting `web-search`, which would have shipped the feature dead on arrival
with every check green. `agent_status` is in the builtin table AND in
`agent/skills/ensemble/SKILL.md` (added in commit fa9b6e4) -- verify both.

## `single-composition-root` (15 pts)

One composition root in the library; the binary uses it. The reference tree
has exactly one production call to the library constructor after commit
6077d4b.

Beware: `NewBareAgent` is a SECOND legitimate library constructor with four
live callers (`cmd/virtual-user/main.go:111` plus three tests), documented in
chapter 13. The check must not treat its existence as a failure. What the
check forbids is the binary hand-building the parts instead of calling a
constructor -- i.e. a second *flat root*, not a second constructor.

## `ch21-parity` (10 pts) and `agent-builds` (5 pts)

Parity runs chapter 21's checks. Note the known O(n^2) parity cascade
recorded in memory: making parity cheap may weaken it, so leave it alone.

`agent-builds` has no mutant; a build failure is not a valid mutation.
Record that in the mutant script header rather than inventing one.

## Harness notes

- `go test ./internal/grade/` takes ~505s. Budget for it; it has not hung.
- Grade with `go run ./cmd/grade -ch 22 ./agent` from the repo root.
- Chapter 20 is prose-only and has no solutions directory; chapter 19 is the
  nearest grader template, and chapter 21 is the nearest *behavioural* one
  (`ch21_run.go` launches the real binary against a fake vendor).
- Each scenario needs its OWN work directory. Chapter 21 lost time to a
  shared workdir letting `save.json` resume the previous scenario.

## Pre-existing failure, do not chase

ch19 grades 85/100 on a missing `agent/events.jsonl`. Verified pre-existing
by `git worktree` at `273797c`, the commit before this session's first,
where it fails identically. Memory also records ch8-ch12 at 80-90
pre-existing.
