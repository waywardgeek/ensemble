# Ch11 Review — Persistence

## Chapter thesis

The event log is the truth; the context is what the truth means right now.
Save and load make the reducer's correctness claim falsifiable.

That thesis survived the rewrite. What changed is the contract around it:
persistence is no longer something the user opts into with a flag. The agent
saves itself on exit and loads itself on start, from `save.json`, with no
flag at all. The save carries an anchor (`as_of`) so a load installs the
snapshot and replays only the tail above it, instead of rebuilding the whole
context from event zero every time.

## Decisions

1. `SaveFile` = `{Config, AsOf, Context, Log}` — **four** fields. `AsOf` is the
   Seq of the last event already folded into `Context`.
2. `SaveConfig` captures vendor, model, system prompt, tools — never API keys.
   On load it is a record, not a restore: the running agent's own config wins.
3. One replay loop survives, `SaveFile.Restore`. Non-null context: install it
   and apply events with `seq > as_of`. Null context: apply the whole log to a
   fresh context. The old exported `Rebuild(events)` is gone; `Log.Replay`
   delegates to `Restore` so there is exactly one place replay can be wrong.
4. `--load` is gone. `--save PATH` changes *where*, never *whether*: the same
   path is read at start and written at exit. Default `save.json`.
5. A save file that exists but does not parse is **fatal** — exit non-zero,
   name the file, leave the bytes alone. Starting fresh over it would
   overwrite the user's history ten seconds later at exit.
6. Writes go through a temp file and a rename, at mode 0644.

## What the grader tests (7 checks, 100 points)

| Check | Pts | Proves |
|---|---|---|
| save-shape | 15 | top-level `config`/`as_of`/`context`/`log`; config fields present; log Seq strictly increasing; `as_of` == last log Seq |
| default-load | 20 | a second start in the same directory, no flags, sees the first conversation |
| replay-equals-snapshot | 20 | the save as written and the same save with `context: null` render byte-identical vendor requests (rule 8) |
| tail-applied-once | 15 | an old snapshot spliced onto a newer log yields the same context as having run straight through — no skipped tail, no doubled tail |
| log-not-needed | 10 | `"log": []` with a non-null context is a complete save |
| bad-save-refused | 5 | an unparseable save file is fatal and its bytes are untouched |
| ch10-parity | 15 | every Chapter 10 check still passes |

Two properties of the check set are worth stating, because they are what the
old grader got wrong.

**Nothing is graded on a verdict the student printed.** The old
`deterministic-context 20` ran the student's own `verify` subcommand and
awarded 20 points if it printed `MATCH`. A solution that printed `MATCH`
unconditionally scored those 20 points. Every check here reads behaviour from
outside: the top-level fields of the save file on disk, and the bodies of the
requests the fake vendor received.

**Nothing parses an event.** Event shape is the student's design — the TL;DR
fixes the four top-level keys and nothing below them. Where a check needs an
old snapshot spliced onto a newer log, it swaps top-level JSON fields only.

## Deletion audit (8 mutations, all caught, none survived)

Each mutant deletes exactly one behaviour from `solutions/ch11`, then the
whole ch11 grader is run. Reverted with `git checkout` after each.

| # | Behaviour deleted | Failing checks | Score |
|---|---|---|---|
| M1 | `as_of` never recorded (anchor always 0) | save-shape, replay-equals-snapshot, tail-applied-once | 50 |
| M2 | save file never loaded at start | default-load, tail-applied-once, log-not-needed, bad-save-refused | 50 |
| M3 | tail above the anchor never applied | tail-applied-once | 85 |
| M4 | anchor ignored: whole log replayed over the snapshot | replay-equals-snapshot, tail-applied-once | 65 |
| M5 | unparseable save tolerated instead of fatal | bad-save-refused | 95 |
| M6 | snapshot ignored: context always rebuilt from the log | log-not-needed | 90 |
| M7 | context never written to the save file | log-not-needed | 90 |
| M8 | `config.system_prompt` never written | save-shape | 85 |

Coverage, read down the other axis — every point-bearing ch11 check is killed
by at least one mutant:

- save-shape — M1, M8
- default-load — M2
- replay-equals-snapshot — M1, M4
- tail-applied-once — M1, M2, M3, M4
- log-not-needed — M2, M6, M7
- bad-save-refused — M2, M5

`ch10-parity` is not mutated here: it re-runs the Chapter 10 suite, which owns
its own audit.

M3 is the one that matters most. Skipping the tail is the single most likely
student bug — it looks correct on every save where `as_of` happens to equal
the last Seq, which is every save the agent writes itself. Only a spliced
fixture catches it, which is why the harness builds one.

Two non-obvious results. M2 (never load at all) also fails `bad-save-refused`,
because an agent that never reads the file never notices it is garbage. And
M7 (never write the context) is caught only by `log-not-needed` — `save-shape`
deliberately does **not** require `context` to be non-null, because the TL;DR
explicitly blesses a log with no snapshot as a complete save.

## Problems found in the spec

The brief says to report TL;DR rules that are wrong rather than silently
deviate. Three things, in descending order of severity. The first two are the
chapter body contradicting its own TL;DR; the third is a genuine internal
contradiction between two TL;DR rules.

**1. §11.2 declares a different struct from the TL;DR.** The TL;DR gives four
fields with an anchor. §11.2 still says:

> Three fields:
> ```go
> type SaveFile struct {
>     Context *Context   `json:"context"`
>     Log     *Log       `json:"log"`
>     Config  SaveConfig `json:"config"`
> }
> ```

No `AsOf`, and `Log *Log` rather than `[]Event`. A reader who codes §11.2
cannot pass `save-shape`, which requires `as_of`. §11.3 then presents
`Rebuild(events)` — full replay from scratch — as "the most valuable test in
the chapter", which is precisely the behaviour mutant M4 deletes.

**2. §11.6 still documents the removed flag.**

> ```
> --save PATH    Persist state to PATH after the conversation
> --load PATH    Resume from a previously saved file
> ```

TL;DR rule 1 says the opposite: "Neither needs a flag… The flag changes where,
never whether." `--load` no longer exists in the reference.

I did not edit `chapter-11.md` — the brief limits me to this file. Sections
11.2, 11.3 and 11.6 need a prose pass to match the TL;DR.

**3. Rule 4 and rule 8 contradict each other.** Rule 4:

> **The log is not needed.** `"log": []` with a non-null `context` is a
> complete save.

Rule 8:

> **Replay is deterministic.** Loading a save as written, and loading the same
> save with `context` set to null, must produce byte-identical vendor requests
> for the next prompt.

Take a save that rule 4 declares complete: non-null context, empty log. Set
its context to null, as rule 8 instructs, and you have an empty context and an
empty log — a fresh agent. Its next request cannot be byte-identical to one
carrying the whole conversation. Rule 8 is false for exactly the saves rule 4
blesses.

The rules are each individually right about something real; rule 8 is just
missing a precondition. It holds when the log is complete, which is true of
every save the agent writes itself, and false for the hand-trimmed save rule 4
permits. Suggested wording: *"For a save whose log is complete, loading it as
written and loading it with `context` set to null must produce byte-identical
vendor requests."*

I implemented rule 8 with that precondition — `replay-equals-snapshot` runs
against a save the agent actually wrote, which always has a complete log. So
the grader is correct under either wording, but a student reading rule 8
literally and testing it on a trimmed save will think they have a bug.

**Ungraded rule.** Rule 6's "write the file and exit within 10 seconds" is not
measured by any check. It is a real requirement and a student could violate it
while scoring 100. Worth either a check or a softer phrasing.

## Collateral found: runs were sharing a working directory

Default load made a latent bug reachable. The grader launched the agent with
no private working directory in five places: ch5 and ch6 used the student's
own solution directory, and ch1, ch7 and ch12 used the grader's. With no
`--load` flag to opt in, the agent now finds `save.json` wherever it starts —
so consecutive grading runs began loading each other's conversations, and
`save.json` was left behind inside `solutions/ch11` and the repo root.

That is a grader bug that predates this chapter; default load only exposed it.
Every launch now gets its own directory (`freshRunDir`). After the fix a full
sweep leaves no `save.json` anywhere in the tree.

This is also a warning for the student-facing text: the chapter should say
that the agent writes to the directory it is *started in*, because that is now
load-bearing behaviour rather than a flag.

## Verification

- `make grade-dir CH=11 DIR=./solutions/ch11` → **100/100**, all 7 checks pass.
- ch11 grader also scores 100/100 on `solutions/ch12`, `solutions/ch13`,
  `solutions/ch14` and `./agent` after the forward port.
- Full sweep of every chapter grader after the change is **byte-identical** to
  the sweep before it. No regressions.
- Pre-existing failures, unchanged and untouched: ch5 on `./agent` NOSCORE,
  ch6 45/100, ch14 on `./solutions/ch14` 5/100. `go test ./... -count=1`
  fails 19 tests under six top-level cases — `TestCh3PointsSumTo100`,
  `TestCh6ReferenceScores100`, `TestCh6DeletionAudit`,
  `TestCh7ReferenceScores100`, `TestCh7DeletionAudit`, `TestCh8Grade`. The
  failure set is **identical, name for name, at my parent commit** (`7a0ddd9`,
  checked in a clean worktree), so none of it is fallout from this work. It
  does mean the repo's test suite was already red before this chapter, which
  someone should look at — ch6 and ch7 are the two chapters whose harnesses I
  touched, and it would be easy to blame the wrong change later.

## Voice

Dense, code-forward. 1015 prose words, 183 lines. Three invariants stated as
testable properties, not aspirations. (The previous revision of this review
said 751 words; that predates the TL;DR rewrite.)

## Open

- Sections 11.2, 11.3 and 11.6 still describe the pre-rewrite contract and
  need a prose pass — see "Problems found in the spec" above.
- Rule 8 needs its missing precondition, or rule 4 needs to stop blessing the
  log-less save.
- Rule 6's 10-second exit budget is unmeasured.
- Interview/resurrection mode (mock tools, exact context replay) — future chapter.
- Memory cascade depends on persistence working correctly — next chapter.
