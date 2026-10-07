# Chapter 22 — Coder's review for the author

CODER → AUTHOR. Nothing in `book/` was edited. This reports what the repair
actually did, which claims in the outline and brief are now wrong, and the
three decisions that are yours.

All figures below were re-measured today against the working tree at
`40389b9`, not recalled.

---

## 1. Shipped state (so the prose can describe something real)

Stages 0–11 are complete. `grade22` is 100/100 on **both** `./agent` and
`./solutions/ch22`; `grade22-audit` is **6/6**; the full cross-chapter sweep
exits 0 with every chapter at 100/100.

The seven checks, with their point values, exactly as the grader prints them:

| pts | check |
|----:|-------|
|   5 | The agent builds |
|  20 | Capabilities are reached through a back-pointer chain, not stapled on |
|  15 | One composition root in the library, and the binary uses it |
|  20 | `agent_status` reports model, usage, cache and cost |
|  15 | The tool package reaches the engine only through the hub |
|  15 | Session cost is the sum of per-model costs, not the total at the current price |
|  10 | Chapter 21's behaviour still works |

That is the 7-check / 100-point budget the course has published since ch18.
No eighth check was added.

---

## 2. MUST FIX — claims that would print as false

**2.1 `ContextStats` does not exist.** The brief refers to it. Measured today:
zero occurrences anywhere in the agent module. Any sentence built on it has no
referent.

**2.2 `common.Call{` has 7 production call sites, not 1.** The brief says one.
Measured: 7 (excluding `_test.go`). This matters because the chapter's argument
is that adding a field to `Call` is cheap *because* the chain is narrow — the
honest version of that claim is "seven sites", and seven is still narrow.

**2.3 `NewAgent` now has exactly one caller; before the repair it had zero.**
The decay finding is that the library's composition root was *dead code* while
`cmd/main.go` hand-built a second root. Please keep the tense straight: zero
callers is the *before* state and is the whole point; one caller is the result.

**2.4 `cmd/main.go` is 1,009 lines now.** The outline's 1,154 is the
pre-repair figure. If the chapter quotes a before/after, those are the two
numbers, and `cliHost` is now at **zero** references — it is deleted, not
shrunk.

**2.5 Do not say "thirteen call sites send on `c.send`".** I nearly shipped
this myself and have since corrected the code comment. Exactly true: **thirteen
call sites send on a client's send channel — eleven spelled `c.send`, plus
`agentClient.send` and `cl.send`.** Against exactly **one** close. The
13-against-1 framing is the exhibit, so it is worth stating precisely rather
than approximately.

**2.6 The ch14 leak is five agents per SWEEP, not per invocation.** I wrote
"per invocation" in a commit and it was wrong: the sweep grades ch14 twice yet
only five leak, so the per-invocation rate has not been pinned down. If the
chapter mentions it at all, say "per sweep".

---

## 3. ENRICHMENT — material the chapter did not have before

### 3.1 The grader found a real crash. This is the chapter's best exhibit.

Driving the real binary — connect, switch model, disconnect — panicked the
agent with `send on closed channel`, raised from `Hub.Observe` by way of
`Actor.notify` and `Actor.setState`. It reproduced in roughly **one grader run
in six**: often enough to fail honest work, rare enough to read as grader
flakiness.

Both sides of the race do the same thing — mutate shared state under `h.mu`,
then act *after* unlocking. `Observe` snapshots the live clients, unlocks, then
sends. Teardown deleted the client, unlocked, then closed `c.send`. So a client
could be snapshotted, deleted, closed, and then sent to.

The line I would build the section around:

```go
select {
case c.send <- data:
default:   // "drop rather than block the actor"
}
```

**That `default` looks like protection and is not. It saves a sender from a
FULL channel, never from a CLOSED one.** The defensive-looking code offers zero
defense against the failure that actually happens. A reader who has written Go
for years will not know this, and will recognise the shape.

It is also the ch4 deadlock rule inverted. Ch4: whoever makes a condition true
owns waking the waiters. Here: **whoever closes a channel owns stopping its
senders** — and nobody did.

The fix is ownership, not a narrower window: nothing closes `c.send` at all,
and `writePump` stops on a separate `done` channel. An unreferenced channel is
garbage collected whether or not it was ever closed, so not closing costs
nothing.

### 3.2 My regression test passed TWICE with the bug restored

This is the most useful teaching material I produced, because it is a failure
mode of *testing*, not of concurrency:

1. 64 clients torn down in one goroutine — the teardown loop finished **before
   any sender was scheduled**, so `Observe` saw an empty client set. 8,000
   observations in 8 ms. Green.
2. 5,000 rounds of a single client — the gap between `Observe`'s unlock and its
   send is a few **nanoseconds**; teardown never once landed inside it. Green.

What works is a **large** snapshot (2,000 clients), because then the send loop
itself *is* the window: `Observe` is still writing to early clients while
teardown closes later ones. **Widen a race window structurally; never hope to
land in it.**

Related trap, same section: `Observe` early-returns when `marshalObservation`
returns nil. My first attempt used an observation the switch does not handle,
so the test passed while exercising nothing whatsoever.

### 3.3 The ch5 check that protected vocabulary instead of structure

Already in the outline, but the implementation sharpened it: the replacement
identifies the logging method by **signature** (format string + variadic + no
results), not by name containing "log". A name-based anchor hiding inside a
structural check is the same bug wearing a better coat. I wrote the name-blind
fixture first, deliberately.

### 3.4 The new check PASSED the real decayed tree at first

The hand-written fixtures all missed what one real tree caught. The check
counted composition roots **tree-wide**, so `cmd/virtual-user`'s healthy root
satisfied the count on behalf of the decayed `cmd/`, which called no root at
all. Now counted per-binary.

**Fixtures agree with the belief that built them; a real pre-repair tree does
not.** Reproducible: `git worktree add /tmp/ch22-decayed 273797c`.

(Do not cite a fixture count. I nearly wrote "ten" from my own notes; the test
file has four test functions and no named-fixture table, so the number is
unverified. The point stands without it.)

### 3.5 The threshold is measured, not tuned

`compositionRootFanout = 3`. Measured on the real tree: the least-reaching
genuine root reaches **4** packages; the greatest-reaching non-root reaches
**1**. Nothing sits at 2 or 3 — the constant sits in an empty gap, and a test
asserts that. Worth a sentence: thresholds chosen by measurement can be
defended; thresholds chosen by tuning cannot.

### 3.6 "Main sets a field on a library value" is NOT by itself decay

Filling in a config struct before handing it to a constructor has the *same AST
shape* as stapling a capability onto a live engine. Only types separate them,
and the scanner has none. The check had to be restricted to **provenance** —
packages the root itself builds from. This is a genuinely interesting limit of
static analysis and belongs in the chapter if there is room.

### 3.7 A fake usage figure is an input to the thing under test

An early grader draft used millions of tokens; the driven agent concluded its
context was full and withdrew every tool but `micro_handoff`. The checks still
passed — the grader had quietly started testing compaction policy instead of
the status tool. **Token counts have a ceiling as well as a floor.**

---

## 4. DO NOT ADD

- **Do not claim the mutation audit proves the checks are good.** It proves
  each check is *sensitive* to one deliberate deletion. That is all it proves.
- **Do not present the panic as something the grader was designed to find.** It
  was not. The grader drove the real binary and the crash fell out. The honest
  framing — a test that touches reality finds things you did not think to look
  for — is better than the flattering one.
- **Do not describe `single-composition-root` as catching a *reachable* second
  root.** It is a static count and catches an unreachable one too. The mutant
  is deliberately a never-called function.
- **Do not use the `*-course` model rows in any pricing example.** Every one of
  them is unpriced, and an unpriced pair cannot demonstrate a pricing bug. The
  grader uses production rows (`claude-opus-4-6`, `claude-sonnet-5`) for exactly
  this reason.
- **Do not re-litigate "vibe-coded"** — flagged in the outline as a deliberate
  departure from voice.md, in your voice. Left alone.

---

## 5. Facts: verified vs not

**VERIFIED today, by measurement:**

- 7 checks / 100 points, values as tabulated above.
- `grade22` 100/100 on `./agent` and on `./solutions/ch22`; audit 6/6; full
  sweep exit 0, all chapters 100/100.
- 13 sends on a client send channel vs 1 close; 11 spelled `c.send`.
- `cmd/main.go` 1,009 lines; `cliHost` 0 references; `NewAgent` 1 caller.
- `common.Call{` 7 production sites; `ContextStats` 0 occurrences.
- `compositionRootFanout = 3`, with measured gap 1 → 4.
- Panic reproduced and fixed; regression test proven to fail when the close is
  restored.

**NOT verified — do not print as fact:**

- **The one-in-six reproduction rate is an observed frequency, not a measured
  probability.** It came from grader runs during development. If the chapter
  wants a number, I should run a proper trial; otherwise say "intermittently,
  often enough to look like flakiness".
- **Why ch14 leaks five per sweep rather than ten.** Unexplained.
- **ch19 `./agent` = 85/100** is pre-existing (proven at `273797c`, and
  `./solutions/ch19` scores 100 on the same run), but I have not diagnosed the
  missing `agent/events.jsonl` beyond localising it.
- Any LOC comparison against Codex / Gemini CLI / CodeRhapsody in the outline —
  those were measured in a previous session, not today. **Re-measure before
  print**, and state the basis (code lines, production only) once and never
  vary it.

---

## 6. Decisions for the author

**6.1 Backport the panic fix to the frozen `solutions/ch21`?** *(the one that
actually matters)*

Students begin ch22 from `solutions/ch21`. That snapshot still contains the
crash, so any browser tab closed mid-turn kills their agent, and the ch22
grader triggers it by design. They would see intermittent failures on correct
work. Options: backport it (touching a frozen snapshot); make fixing it the
chapter's assignment (but the 7-check budget is published, so it would mean
broadening an existing check, not adding one); or leave it and open the chapter
with the student hitting the crash.

My recommendation is to backport and let the chapter narrate the bug — the
exhibit value is in the prose, not in making students debug a race against a
grader clock. But this is a curriculum call and a frozen-snapshot call, so it
is yours.

**6.2 How should ch22 narrate the panic?** It is the chapter's strongest
concrete artifact and it arrived by accident. It can be the cold open, or a
mid-chapter exhibit for "the component we are moving out was also the broken
one", or a closing coda. I have no stake in which.

**6.3 Does ch22 split?** Still open from the outline: decay audit vs agentic
capstone. The repair work came in at seven checks and a single coherent thesis,
so I no longer think it *needs* to split on size grounds. Purely an authorial
judgement now.

---

## 7. Still open, and not ch22's scope

- **ch14's harness leaks agents** (five per sweep, each in its own
  `ch14-work-*` directory). A ch14 bug, worth its own fix.
- **ch19 `./agent` = 85/100**, pre-existing.
- When reaping leaked agents, filter on the harness temp dir — an unrelated
  long-lived `./ensemble --port 8084` must not be matched.
