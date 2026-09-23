# Brief: Chapter 11 grader + reference to the new TL;DR contract

Role: CODER. Repo: `/Users/bill/projects/ensemble` (always pass `cwd`).
The author rewrote the Chapter 11 TL;DR (book/chapter-11.md, from
`# Chapter 11` through the line before `## §11.1`). **That TL;DR is the
spec.** The current grader (`internal/grade/ch11_*`) and reference
(`solutions/ch11`) implement the OLD contract and must be brought to it.

Read first, in full:
1. book/chapter-11.md lines 1-89 (TL;DR, 8 rules, check table).
2. book/derive-ch11.md (old grader map, weaknesses, construction notes
   for each new check, coder cautions). Section "Ruling 2026-09-22".
3. internal/grade/ch11_checks.go, ch11_harness.go, and the existing
   ch11 grader test / mutation tests.
4. `grep -n 'P9' book/*.md` for the grader-audit policy.

## Deliverables

1. **Grader**: implement exactly the TL;DR check table (names, points,
   total 100): save-shape 15, default-load 20, replay-equals-snapshot 20,
   tail-applied-once 15, log-not-needed 10, bad-save-refused 5,
   ch10-parity 15. Constructions are in derive-ch11.md. Grade behavior,
   not code: the grader splices JSON fields between save files and
   compares captured vendor request bodies; it never parses the
   student's event internals and never trusts student-printed verdicts.
   `--load` no longer exists; `--save PATH` names the file for both load
   and save; no flag means `./save.json` in the working directory.
2. **Reference**: change `solutions/ch11` to obey all 8 rules. Edit the
   existing code incrementally. Do NOT rewrite files wholesale (a past
   rewrite silently deleted 309 lines of teaching comments while still
   grading 100). Check `git diff --stat` before each commit.
3. **Forward ripple**: find out whether later graders (ch12, ch13, ch14,
   and anything run as parity) exercise ch11 behavior, and whether
   `solutions/ch12..ch14` and `agent/` must adopt default load. If yes,
   port the change forward the same incremental way. Also confirm every
   harness from ch1 on runs the agent in a per-run temp cwd; a shared cwd
   plus default load makes runs load each other's conversations.
4. **Mutation audit (P9)**: for each check with points, delete the one
   behavior it protects from the reference and show that exactly that
   check fails (record the exact failing set). Include at least: skip the
   tail (apply nothing after as_of), apply the tail twice (replay from
   seq 0 over the snapshot), ignore default path, start fresh on a
   corrupt file, omit as_of. A check no mutant can fail is decorative:
   fix it.
5. **Review**: write `book/review-ch11.md` for the author: what was
   built, any TL;DR rule that was wrong, ambiguous or unimplementable
   (quote the rule), figures measured, the mutation table, anything the
   8 rules leave a student guessing. This is the ONLY file under book/
   you may create or edit.

## Known cautions (from derive-ch11.md)

- Rule 7: running config wins over saved config. Verify the reference.
- Rule 8 assumes rendered requests carry no timestamp/random field.
  Verify on the reference before trusting replay-equals-snapshot.
- Two full-replay functions exist (`Log.Replay`, event_log.go:62, and
  `Rebuild`, save.go). Keep one.

## Gate

`make grade-dir CH=11 DIR=./solutions/ch11` = 100/100, and every other
chapter grader that passed before still passes (run them; list scores in
the review). `go vet` clean. Use `go test -count=1`.

## House rules

- Commit as CodeRhapsody:
  `git -c user.name='CodeRhapsody' -c user.email='coderhapsody@coderhapsody.local' commit`
- Stage EXACT paths. Never `git add -A` or `git add .` (Bill has many
  untracked files). Never push. Always `git --no-pager`.
- Don't pipe test commands into `head` (it masks the exit code). macOS
  has no `timeout`.
- Never use the shell `kill`; use the kill_job tool.
- Report measured facts only; mark anything inferred as ASSUMED.
