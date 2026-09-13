# Brief: enforce `system` and `model` in the ch1 fake

**Role:** coder. **Author rulings are in `bc54c4f`** (chapter-01-outline.md).
Source of the work: F4 and the model-hardcode item in
`book/review-ch01-grader-audit.md`, both of which correctly stopped at
"author's call." The calls are made. This is the implementation.

## What changed in the chapter

The disclosed wire contract (§1.8, the `wire` bullet) now reads:

> requires `content-type: application/json`, a `model` equal to
> `$ANTHROPIC_MODEL`, a non-empty `system` string, non-empty message content,
> and **no streaming**

So two new rules in the fake's wire validation:

1. **`system` must be present and non-empty.** It was in the §1.2 anatomy all
   along and absent from the graded contract. Deleting it scored 100/100.
2. **`model` must equal the expected model**, not merely be non-empty. Use the
   same shape as the `ExpectedAPIKey` fix: one exported const, harness sets the
   env from it, one source of truth. `ExpectedModel` is the obvious name.

Both are `wire` rules. **Do not add a check and do not re-divide the 100
points** — folding vs splitting is the author's call, and these are wire
conformance, which `wire` already owns. If you think one of them deserves its
own points, escalate the argument rather than acting on it (precedent: the ch2
`seam-render` weighting discussion).

## Why this is allowed to fail a previously-passing student

It can. A student who omitted `system` or hardcoded the model scored 100 and
now will not. That is acceptable **only** because the chapter asked for both
before this change: `system` is in the §1.2 anatomy, and the config table has
always said "read all three, hardcode none." We are collecting on a promise,
not adding a requirement. If you find a case where the chapter does *not*
already ask for the behavior, stop — that is an outline change and it is the
author's.

## Required rigor (P9)

- **Two permanent mutants**, in the style of `firstblock` / `hardkey`:
  `nosystem` (omit the system field) and `hardmodel` (hardcode the model
  string). Suite goes 12 → 14.
- Each mutant asserts the **exact set** of failing check ids.
- **Every mutation must assert its anchor matched exactly once.** A silently
  unapplied mutation scores 100 and manufactures a fake finding, which is the
  precise failure this whole audit exists to catch.
- Negative control: a student sending a correct `system` and the right model
  must still score 100. Confirm the reference solution does.

## The regression that will bite you

**The fake is SHARED with ch2.** `-ch 2` must still score 100/100 after this.
Ch2's solution sends requests through three vendor renderers; the Anthropic one
must be sending a `system` field and the expected model, and if it isn't, the
fix is in ch2's renderer or the harness env, *not* in relaxing the new rule.
Run both graders before you commit:

```
go run ./cmd/grade -ch 1 solutions/ch01
go run ./cmd/grade -ch 2 solutions/ch02
```

Also: `ch02.log` and `internal/grade/ch02.log` are tracked and churn on every
grader run. `git checkout --` them before committing. `book/tic-apply-report.md`
is the author's untracked WIP. Never `git add -A`.

## Not in scope

The **`type` filter (F3) is now a ruled deliberate gap** and must stay ungraded.
The chapter explains why in §1.8: no real Anthropic block carries a `text`
field, so catching a missing filter would require inventing a block type that
does not exist on the wire. If you find an honest way to grade it that does not
teach a falsehood about the API, say so and stop; do not implement it.

## Report

Short. Per-rule: mutant name, failing set, before/after score. Plus the ch2
regression result. If either rule turns out to be unenforceable for a reason
the ruling did not anticipate, say so plainly rather than working around it.
