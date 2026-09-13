# Design note: one linear solutions repo, tagged per chapter

Status: proposal, awaiting Bill's ruling. Raised by Bill 13 Sep 2026.
Implementation is a sub-agent task, to be done AFTER ch3 lands and BEFORE ch4
code exists.

---

## The problem, measured rather than asserted

`solutions/ch03` is a physical copy of `solutions/ch02`: eleven files, five
byte-identical, six edited (+225/−6), two new. The copy already rotted. At the
time of writing, `solutions/ch03/main.go` still introduces itself as **ch02** in
its header comment, its usage string, and its interactive banner.

That is chapter three of ten. A bug found in Chapter 1's code today must be
fixed in nine places, by hand, with nine chances to miss one. And P2 already
promises students they may rebase from our solution at any chapter, which means
these copies are a *maintained baseline* we have committed to keeping correct.

## The proposal

One linear history where the agent is built commit by commit, with a tag at each
chapter boundary. Chapter snapshots under `solutions/chNN` become **exports of
those tags** rather than hand-maintained copies. Fix Chapter 1 once, rebase the
chapter commits forward, re-export.

This also makes `git diff ch02..ch03` literally "what Chapter 3 added" — a
teaching artifact we want to print anyway, that currently has to be reconstructed
by hand.

## Recommended shape

**Not submodules.** Students must be able to clone one repo and see working
code. Submodules produce detached heads, empty directories, and support mail.

**Recommendation: an orphan branch plus a worktree, in the existing repo.**

    git worktree add ../solutions-linear solutions-linear

Authors develop linearly in that worktree. One remote, one clone, no second
project to keep in sync. A separate repo also works and is slightly cleaner
conceptually; it costs a second remote and a cross-repo export. Bill's call.

**Export, and commit the exports.**

    git archive ch03 | tar -x -C solutions/ch03

The exports stay committed. A student clones the course repo and reads the code
without running a generator, which is non-negotiable.

**The guardrail that makes "generated but committed" safe:** a `make
check-solutions` target that re-exports every tag to a temp dir and diffs
against what is committed. Without it the snapshots silently drift from the
linear history and we have the original problem plus a script. With it, drift is
a failing build.

## Three things that will bite, stated up front

**1. Rebasing rewrites tags.** Fixing Chapter 1 and rebasing forward means every
downstream tag points at a new commit. Tags are *regenerated*, not appended.
Anyone who cloned at `ch03` gets different bytes after a fix. For a book this is
correct behavior, but it must be published as such, and pushing means force
pushing a branch.

**2. The linear history deliberately contains wrong code.** Chapter 1 teaches a
naive format specifically so Chapter 2 can demolish it. A well-meaning "fix
forward" that sanitizes Chapter 1 destroys the pedagogy. The rule: propagate
*defects*, never *improvements*. If a change would make an early chapter better
engineering, it almost certainly belongs in the chapter that teaches that lesson.

**3. `ch01 -> ch02` is a rewrite, not a delta.** Chapter 1 is the only
sacrificial chapter, so Chapter 2 replaces most of it. The linear history has one
genuine discontinuity there, and `git diff ch01..ch02` will not be a tidy
exhibit. Everything from ch02 onward is additive, and that is exactly the range
where fix-once-propagate pays.

## A simplification worth taking while we are in here

**Decided 13 Sep 2026 (Bill), and it is free: the default log name becomes the
executable's own name plus `.log`.**

Today both solutions hardcode `envOr("CH02_LOG", "ch02.log")`, so a binary built
from `solutions/ch03` writes `ch02.log`, and Bill's own `ch3agent` did too the
first time he used it for real work. Deriving the default from `os.Args[0]`
instead means `ch02` writes `ch02.log`, `ch03` writes `ch03.log`, `ch3agent`
writes `ch3agent.log`, and a future binary named `agent` writes `agent.log`. The
chapter number stops being a constant in the source and becomes a consequence of
what you named the thing.

**Why it is free.** Every grader harness sets `CH02_LOG` explicitly:
`internal/grade/ch02_harness.go` at three call sites and
`internal/grade/ch03_harness.go` at one. The hardcoded default is reached only
when a human runs the binary by hand, so changing it cannot move a graded byte.
`ch2parity` is unaffected.

**What is NOT free, and should not be bundled with it:** renaming the
environment variable itself. `CH02_LOG` is chapter-numbered and ugly, but a
student who passed Chapter 2 wrote code that reads `CH02_LOG`. If the harness
starts setting `AGENT_LOG`, that student's passing solution fails for a reason
the chapter never asked about — which is the line drawn earlier: strengthening a
grader may fail a prior-100 student only where the chapter already made the
promise. Keep `CH02_LOG` for now and fold the rename into this restructure,
where the tags are being regenerated anyway and a coordinated change is cheap.

---

The copy/paste rot exists *because we pretend each chapter is a different
program.* It is one program that grows. If the linear repo builds a binary
called `agent`, the "ch02 in a ch03 banner" class of bug cannot occur again:
there is no chapter name embedded in the source to go stale.

Open question, because it touches graders: today the CLI contract and the
`CH02_LOG` environment variable carry chapter numbers, and `ch2parity` grades
Chapter 2's commands unchanged. Either the export renames on the way out, or we
standardize on a stable name and accept a one-time grader change. My lean is the
stable name, because it removes a whole category of defect permanently, but it is
a grader-visible change and therefore not mine to take unilaterally.

## Student-facing half, which we should adopt because we are asking them to

Bill's point: students copy code forward too, and should be using source control
with a tag at each chapter that passes. Ch3 §3.9 now ends on:

    git commit -am "ch3: six tools, grader 100"
    git tag ch03-pass

Proposal: **P11** — every chapter ends by telling the student to commit and tag
the passing state, and every chapter has a "drive it yourself" section before it.
The tag is what makes the next chapter safe to attempt, and Chapter 4 immediately
takes `run_command` apart, so Chapter 3 is where the habit has to start.

## Sequencing

Do this after ch3's prose is final and before ch4 code exists. The conversion
cost scales with the number of chapters; it is three now and will be four soon.
The work is mechanical and well specified, which makes it a good sub-agent task —
with the build commands stated in the brief, because sub-agents do not see
`cr/project.json`.
