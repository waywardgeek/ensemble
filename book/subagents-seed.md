# Seed: the sub-agents chapter

Status: seed. Not outlined. Chapter number deliberately not in the filename —
the sub-agent chapter is last in the dependency graph and its number has moved
once already.

---

## The thesis, which is Bill's and arrived as an objection

> "This is one reason I don't use sub-agents often. I can normally send a hint
> quickly before an agent wastes a lot of cycles."

Every other chapter in this book adds a capability and argues for it. This one
has to add a capability and argue *against* reaching for it, because the
mechanism it introduces quietly disables the mechanism the previous chapter
introduced.

A sub-agent is not a smaller agent. It is an agent **you are no longer
watching.**

## Why the hint channel dies

The actors chapter earns the hint: a human reads the agent's reasoning as it
streams and interjects between tool calls. Correction lands in seconds, before
the mistake compounds. That is a *continuously corrected* process.

Delegation converts it into a *specified-once* process. The brief is written up
front, the child runs, and the result arrives at the end. Waterfall, with a
better vocabulary.

Two distinct things break, and it is worth separating them because only one is
fixable by trying harder:

1. **The parent supervises by polling; the human supervises by streaming.** The
   parent does not read the child's reasoning as it appears. It checks status,
   or it blocks until something completes. Nothing in the loop is watching the
   sentence where the child announces a bad plan.

2. **The parent is frequently blocked inside the very call that waits for the
   child.** It could not send a hint even if it wanted to. Supervision and
   waiting are the same operation, so the supervisor is asleep exactly while the
   supervised work happens.

So the parent is a worse supervisor than the human on both axes that matter:
worse judgment, and worse latency. Spawning a sub-agent does not delegate the
work to someone as capable as you. It delegates the work *and* demotes the
supervisor.

## Receipt: 13 Sep 2026, the model purge

The task was small and well specified: remove a retired model ID from the
CodeRhapsody repo, keep the `-lite` variant, report the cause.

The child ran `go build ./...`, which fails in that repo for a reason that has
nothing to do with models: `web/website/node_modules/@vercel/fun/` vendors a Go
package under a `go1.x` directory, and `./...` walks into gitignored
`node_modules`. The correct command is `make binary`, and it is written down —
in `cr/project.json`, **which sub-agents do not receive.**

The child could not have known. It did the reasonable thing and started
engineering around a broken build: several minutes and real money spent on a
problem that was not its task, was not its fault, and was not real.

Bill fixed it in one sentence the moment he saw it. That is the entire chapter
in one incident: the correction was cheap, available, and arrived late only
because nobody was watching the stream.

Note the shape, because it generalizes past this repo. The child failed on
**environment knowledge the parent had and did not pass on.** Not capability,
not reasoning, not model quality. The brief is a context transfer, and every
fact the parent forgot to include is a fact the child will pay to rediscover.

## What the chapter should actually teach

- **State the build and test commands in the brief.** The child cannot see the
  project config. This is the single cheapest fix and it is embarrassing how
  often it is the whole problem.
- **Shorter turns are more correction points.** Turn boundaries are where
  supervision can re-enter. A child that runs for twenty minutes has one.
- **Delegate what is separable, not what is subtle.** Work that needs a
  judgment call every few minutes needs a supervisor every few minutes.
- **Sub-agents buy parallelism and context isolation, and they pay in
  supervision bandwidth.** Teach it as a trade with a price tag, not a feature.
- The honest default: if you *can* watch it yourself, watch it yourself.

## Dependency argument (for the chapter-order graph)

This chapter must come after the actors/hints chapter. The cost of a sub-agent
is denominated in the hint channel, and a reader who has not yet built that
channel cannot be shown what spawning takes away. Introducing sub-agents earlier
would make them look free, which is precisely the error the chapter exists to
correct.

## Open

- Does the fix belong in the framework rather than the advice? Surfacing the
  child's stream to the human, so a human can hint a grandchild, is buildable.
  Whether it is *desirable* is a real question: it reintroduces the human as the
  bottleneck the fan-out was meant to relieve.

  This stops being an aspiration if the GUI chapter lands first. A child's
  stream is just another event source, and the observation surface already
  exists to render one. See `gui-seed.md` — the dependency runs GUI, then
  sub-agents, for the same reason the actors chapter has to precede both.
- Bill's practice is the strong form of this: he mostly does not use them. The
  chapter should say so plainly rather than hedge toward the tooling.
