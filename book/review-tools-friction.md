# Coder's review: three friction points from a model using the ch4 agent

*Coder: CodeRhapsody (Opus). Editor: Bill. 2026-09-14. For the author, who owns
the ch3 and ch4 outlines and the graders' point tables.*

## 0. Where this came from

Bill ran the live `agent/` tree with Fable 5.1 as the model and asked it what
it thought of the tools. It named two design decisions as the standouts, both
of which the chapters argue for: `edit_file`'s refusal on an ambiguous anchor
("converts *I silently patched the wrong one of three identical lines* into an
error I have to look at"), and `ai_callback_pattern` ("turns `run_command` from
*fire and hope* into an actual conversation"). It named three gripes. This
review is the code answer to the gripes. The author decides what the chapters
say about it.

A model's opinion of a tool set is not evidence the design is right. It IS
evidence of what the model took from the description text, and the description
text exists for exactly one reader. Treat the three gripes as a reading of
§3.x and §4.x by their audience.

The gripes, verbatim:

1. "`write_file` has no dry run and no diff, so it's the one tool that can
   quietly destroy work; I try to read before I overwrite."
2. "`search_files` gives no context lines, so a match often costs me a
   follow-up `read_file` around it."
3. "`tool_limits` applying to only the *next* call is a bit of a footgun; it's
   easy to burn it on the wrong thing."

Scope: `agent/` only. `solutions/ch03` and `solutions/ch04` are NOT touched
(see §5). Diff: 5 files, +279/−26, of which 137 lines are tests.

## 1. `write_file`: the refusal is the dry run

**Change.** New optional argument `overwrite: boolean`. When the target exists
and neither `append` nor `overwrite` is set, the call is refused:

    write_file refused: notes.md exists (1234 bytes, 40 lines); pass
    overwrite:true to replace it, or use edit_file to change part of it

With `overwrite: true` it replaces and the reply says what it replaced:

    wrote 2000 bytes to notes.md (replaced 1234 bytes, 40 lines)

Creating a new file is byte-identical to before (`wrote N bytes to PATH`).
`append` is never refused. Existence comes from `os.Stat`, not from a
successful read, so an existing file the agent cannot read is still guarded;
the byte count in the refusal comes from `Size()`, so it never under-reports.

**Why this shape.** The model asked for "a dry run or a diff". A refusal is a
dry run priced at one round trip: the first call tells you the file exists and
how big it is, and nothing has happened. A diff was the other candidate and I
rejected it. Its cost is context proportional to the file on every legitimate
rewrite, and the failure it would catch (rewriting a file you never read) is
the one the refusal already forces the model to confront before the write, not
after. It is also `edit_file`'s rule from the other side: of the six tools,
`write_file` on an existing path is the one operation that destroys work with
no trace in the log, so it is the one that must be asked for by name. The
dangerous call is the one that makes you be specific. That sentence is already
the spine of §3.5; this makes it true of both mutating tools instead of one.

**Rejected:** optimistic concurrency (`expected_bytes`/hash argument) as too
much machinery for the chapter; a full unified diff in the reply, per above;
making `overwrite` required in the schema (a new file should not need it, and
7:1 `edit_file:write_file` says most `write_file` calls create).

## 2. `search_files`: `context_lines`, in grep's format

**Change.** New optional `context_lines: integer`, default 0. With 0 the output
is what it was: `path:N:text` per match. With `k > 0`, each match brings `k`
lines either side as `path-N-text`, overlapping windows merge, and `--`
separates groups that are not adjacent, across files too. This is `grep -C`
byte for byte, chosen because the model has read more grep output than
anything this program could invent; a format it already parses costs it
nothing to learn.

`maxMatches` (200) still counts matches, not context lines.

**Default 0, deliberately.** CodeRhapsody's own `search_files` defaults to 2.
I kept 0 here so that every existing student's output is unchanged and so the
tool's baseline volume does not triple for the "where is this used" query the
model called its workhorse. The gripe is answered by the capability being in
the schema, where the model reads it. Flipping the default is one constant if
the author disagrees.

**One behaviour change under default 0, for honesty:** a file ending in `\n`
used to be split into a phantom empty last "line" that patterns matching the
empty string (`^$`, `.*`) reported as `path:N+1:`. It no longer exists. No
grader searched for such a pattern; ch3 and ch4 remain 100 (§4).

## 3. `tool_limits`: the footgun stays; it now fires loudly

**The ruling stands.** Bill ruled on 2026-09-13 that a pending `tool_limits`
is consumed by the very next call, whichever tool that is (CodeRhapsody's
`set_tool_watchdog` semantics), and rejected the category-shaped alternative
(consumed only by the next job or wait call) because it is sticky across the
calls in between. The model's gripe is precisely the cost that ruling accepted.
I did not reopen it.

**What was actually wrong** is that a misfire was *silent*. `Jobs.Take` was
called at the top of `Engine.Execute` for every call and nothing downstream
recorded whether it had consumed a pending set. Burn `max_output_bytes: 200000`
on a `read_file` you did not mean, and the symptom is a truncated build result
two calls later with no cause in sight.

**Change.** `Take` now returns whether it consumed a pending set, and
`Execute` heads the consuming call's report, on both the job path and the
`NoJob` path, with one line:

    [tool_limits consumed by this read_file call: ai_callback_delay 3s,
     ai_callback_pattern none, max_output_bytes 200000]

The `tool_limits` reply and its description now say "whichever tool that is"
outright, and the reply says the next call's result will report the
consumption. `Limits` grew a `String()` so both places use one vocabulary.

Semantics unchanged; visibility added. A burn is now an error you read in the
result it caused. That is the book's rule about failures applied to the
book's own tool.

**Not built, for a ruling (§6.1):** a `tool` argument on `tool_limits` naming
the intended target, with *discard-on-mismatch*: if the next call is a
different tool, the limits are dropped and that call's report says
"tool_limits meant for run_command discarded". Still one-shot (nothing
survives past the next call), so within the ruling's rationale, but it changes
what a misfire DOES (defaults instead of the misapplied limits). The argument
for it: a misapplied `max_output_bytes: 200000` on a big `read_file` dumps
200 KB into context, so misapplied limits do damage, discarded ones do not.
The argument against: a second code path to grade, and the note already makes
the misfire a one-round-trip correction. That is a design call, not a coder's.

## 4. Verification

- `go build ./... && go vet ./...` clean. `gofmt -l agent/` lists only
  `agent/seam.go`, which predates this change and I did not touch.
- `go test ./agent/` passes. Three new tests in `agent/tools_test.go`:
  `TestWriteFileRefusesSilentOverwrite` (refusal text and untouched file;
  overwrite replaces and reports; controls: new file needs no flag, append
  never refused), `TestSearchFilesContextLines` (exact output for
  `context_lines` 0 and 1, including merge and both `--` separators),
  `TestToolLimitsConsumptionIsVisible` (`Engine.Execute` level: the
  `tool_limits` reply warns, the next call's report carries the note with the
  taken values, the call after that does not).
- **Deletion audit (P9 shape, at the unit level):** five mutants of the
  reference, each confirmed applied by grep before the run, each caught by
  the new tests, files restored byte-identical after (`cmp`):
  `no-refusal` (`exists && false`), `no-separator` (drop the `--` append),
  `no-merge` (do not advance `lo` past the previous window),
  `no-note-jobpath` and `no-note-nojobpath` (`fromPending && false` on each
  branch of `Execute`).
- **A false pass caught on the way:** the first version of the
  two-`tool_limits` leg passed `"ai_callback_delay":"7s"`; the schema wants a
  number, the call was a malformed-args error, nothing was pending, and my
  test helper did not check `IsError`, so the leg was asserting on the wrong
  path entirely. The helper now fails on any tool error. Same shape as the
  ch4 coder's `killed-job-reported-as-done`: a helper that swallows the
  failure it should be reading.
- **Graders against the live tree:** `go run ./cmd/grade -ch 3 ./agent` and
  `-ch 4 ./agent` both 100/100. No check was added or changed. The changes are
  invisible to every promise the chapters currently collect on: the ch3
  harness creates `src/greet.go` fresh, never overwrites, never asks for
  context, never uses `tool_limits`; ch4's `toollimits` leg checks the limits
  took effect, not the report text.

## 5. Snapshots now diverge from `agent/`

`solutions/ch03` and `solutions/ch04` are frozen at their tags and no longer
equal `agent/`. Two of the three changes are to ch3 tools, one to a ch4 tool.
I did not re-snapshot: the snapshot is the chapter's artifact and moves when
the chapter's text does. When the outlines are updated (§6.2), the coder step
is: copy `agent/` over both snapshots, `diff -rq` to confirm, re-tag
`ch03-solution` and `ch04-solution`, run `-ch 3 solutions/ch03`, `-ch 3
solutions/ch04` (regression gate) and `-ch 4 solutions/ch04`.

## 6. For the author

### 6.1 Rulings needed (Bill or author)

1. `tool_limits` discard-on-mismatch `tool` argument (§3): build, or leave the
   note as the whole fix. My recommendation: leave it. The note costs nothing
   and makes the correction one round trip; the ruling's whole point was that
   one-shot settings need no target.
2. `context_lines` default 0 or 2 (§2). Recommendation: 0.
3. Whether any of this is graded (6.3).

### 6.2 Outline changes to guide students this way

- **ch3 §3.5 (mutating tools).** `edit_file`'s refusal is currently the
  chapter's one instance of "the dangerous call makes you be specific".
  `write_file`'s overwrite guard is the second instance of the same rule, and
  stating it as a rule with two instances is stronger than one refusal that
  reads as a feature. The beat: of six tools, exactly one operation destroys
  work with no trace in the log; name it, gate it. The model's own words
  ("the one tool that can quietly destroy work") are the reader's intuition
  stated back.
- **ch3 search_files.** One sentence on `context_lines` and why grep's format
  (the model already parses it). Optional: the "a match costs a follow-up
  read_file" observation as the motivation.
- **ch4 §4.6/§4.7 tool_limits.** The one-shot rule is already stated. Add: the
  footgun in a one-shot setting is a silent misfire, so the consuming call
  reports the consumption. This is a clean instance of the chapter's
  loud-failure principle applied to the chapter's own mechanism, and it
  pre-empts the objection a careful reader will raise on their own.
- **Possible pull-quote.** The two standouts the model named are the two
  design decisions ch3 and ch4 argue for. Whether a live model's endorsement
  belongs in the text is the author's call; if used, frame it as the
  audience's reading of the description text, not as evidence.

### 6.3 If graded (proposal only; nothing built)

- `writeguard` in ch3, 5 pts from `mutatetools`: write an existing planted
  file without `overwrite` (error, file byte-identical), with `overwrite`
  (replaced), a new file (no flag needed). Same shape as `editcontract`.
  Worth grading: it is a safety contract, not an ergonomic.
- `context_lines`: recommend NOT graded. It is an ergonomic, any format is
  coherent, and grading grep's exact bytes would be representational.
- `toollimits` in ch4: add a leg requiring the consuming call's report to
  mention `tool_limits`. Loud consumption is the point; a student whose burn
  is silent has the footgun the chapter is warning about. P9: the deletion
  `fromPending && false` must drop the score.

### 6.4 Wire facts the text can state (verified)

- `write_file` refusal text and overwrite reply are as quoted in §1.
- `search_files` context format is as in §2; `--` also separates the first
  group of a file from a previous file's output.
- The consumption note is the first line of the consuming call's result, on
  every tool including the `NoJob` four; a second `tool_limits` call consumes
  the first and says so.
