# Chapter 3 comparative code review

Reviewer: `review_restart_ch01`, October 9, 2026. **Accepted with the inherited
media exception.** Initial review; zero correction rounds used. No new blocking
finding. This accepts the implementation, not a passing original-grader result.

Read the entire `.agents/skills/ensemble-coding/SKILL.md`, current AGENTS,
carryover/workflow, original Chapter 3 and the completed
[student review](ch03.md). Chapter SHA-256 is
`b83b9e989bb213bc3ac36f092ed80297fb996e5d64d9da1adb7af74bcbd4f3d8`.
Compared current `solutions/edition-2/main/` with frozen Edition 2 Chapter 2,
and original `solutions/ch03/` with its own `solutions/ch02/` predecessor.

## Behavior, ownership and scope

The six blocking tools and sequential sixteen-round loop satisfy the current
exercise. Agent records the complete response before dispatch, preserves call
identities, records failures as results, and returns the final response text.
The student's corrections to intermediate narration, overlapping edit anchors
and unexpected calls with an empty registry are genuine fixes. The Gemini fake
now terminates the conversation without weakening its existing exact model,
signature and large-number assertions.

Ownership is real: the public root constructs and retains Agents; Engine and
History retain their owning Agent; Agent owns its registry and configuration.
Tools receive that actual Agent through a narrow common interface, preserving
access to the ownership chain without a synthetic service container. Import
inspection confirms implementation spokes depend on common, not one another;
common contains vocabulary/interfaces and JSON dispatch, not tool behavior.
The unchanged separate GUI module uses the public Agent and passes its shared
conversation test. No mocks, mutable global registry or new dependency appears.

Compared with the original, an ordered per-Agent registry removes global
capability state and sorting machinery. Dispatch operates on the current reply
rather than repeatedly scanning all history. Direct typed argument decoding,
atomic new-file refusal and precise unique-anchor handling are clear improvements.
Search implements the chapter's context-window contract, which the matching
original tool does not implement. Growth is attributable to the exercise and
Bill's credential rule; there are no premature jobs, scheduling or permissions
frameworks. Reading whole files and collecting command output in memory are
documented synchronous Chapter 3 limitations, not grounds for early redesign.

## Comparison

Counts classify every physical Go line as blank, leading `//` comment-only, or
active, including braces/imports and literal-schema lines. There are no block
comments. Tests are separate; no files moved between categories. Original
compiled binaries are excluded. The student's two unchanged module manifests
are support files; neither tree adds generated code or another source language.

| Go category | Active | Comments | Physical | Chapter delta: active / comments / physical |
| --- | ---: | ---: | ---: | ---: |
| Original production | 2,373 | 848 | 3,536 | +521 / +197 / +774 |
| Student core production | 1,815 | 300 | 2,223 | +395 / +51 / +458 |
| Student GUI production | 8 | 3 | 13 | 0 / 0 / 0 |
| Original tests | 234 | 85 | 345 | +46 / +8 / +57 |
| Student core tests | 693 | 20 | 739 | +226 / +4 / +239 |
| Student GUI tests | 29 | 2 | 34 | 0 / 0 / 0 |

The student adds 24% fewer active production lines. Single-line JSON schemas
also affect this metric; size alone does not establish improvement. Control flow
remains ordinary readable Go. Combined production comments are 14.3% of nonblank
lines versus 26.3% in the original, below Bill's approximate 20% target. The
new explanations nevertheless teach ordering, ownership, failure semantics,
overwrite atomicity, overlapping anchors, output bounds and credential lifetime
beside the relevant code. I accept their usefulness without requesting filler.

The larger test addition protects real integration behavior: sequential disk
effects, exact grep groups, capped ranges, append/refusal, call correlation,
empty registries and subprocess environment filtering. It exercises the real
Agent, Engine, History, filesystem and shell with small HTTP boundary fakes.

## Verification and exception

Independent reviewer runs: empty `gofmt -l`; core and GUI vet and uncached tests
pass. Original `make grade-dir CH=3 DIR=solutions/edition-2/main` reproduces
**90/100 FAIL**: all nine new checks pass; only `ch2parity` fails with inherited
`ref-render` and `ref-redaction`. The coordinator's original reference baseline
is 100/100. The student's initial 55/100 and subsequent corrections remain
recorded in its review.

Read all [35 mutation records](../../solutions/edition-2/evidence/ch03/mutations.txt).
They identify actual defects and intended failures, preserving the baseline
parity failure. Independently restored three defects in disposable copies:
credential inheritance, empty-registry call skipping and unmerged adjacent search
windows. Each failed its intended assertion, without compilation failure. The
credential check runs the actual shell with four dummy keys and an ordinary
environment positive control. Scratch copies were removed. Unchanged earlier
behavior reuses prior mutation evidence; this is not exhaustive proof.

Audited all four live JSON logs against the [live record](../../solutions/edition-2/evidence/ch03/live.txt):
ordered calls/results, actual failing-then-passing tests, targeted edits,
missing-file failures and subsequent codename recall are present for all three
providers. Usage sums match 16,271/863, 7,375/343 and 8,467/563 input/output;
the extra search/write/read session matches 5,975/476. Initial model refusals
remain visible. The student's PTY/build and inspected-diff observations support
human-interface use; I did not repeat paid calls. The two later narrow fixes
affect paths outside those enabled-tool demonstrations and have deterministic
coverage.

I accept the student's exception explanation: these are exactly the positive
media-rendering prerequisites waived in [Chapter 2 review](ch02-code-review.md),
not new Chapter 3 failures. Ref preservation, redaction and Anthropic ordering
remain required and pass their regressions. Keep the actual failed parity score
visible in checkpoints and later results; do not turn it into an earlier-feature
requirement or silently label it passing. Reassess the media exception when the
original Chapter 4 introduces blob capability. No additional exception is granted.
