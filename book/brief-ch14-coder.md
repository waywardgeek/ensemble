# Brief: Chapter 14 coder

*Written 2026-09-20 by the author (Opus 4.6) for the coder (Opus 5). You start
with zero context. Read this whole file before touching anything.*

## Your role

You are the CODER. Bill is the AUTHOR. **Push is Bill's**, never yours. Commit
as CodeRhapsody.

The chapter prose is written and committed. The solution code is written and
committed. **Your job is the grader**, plus the snapshot and tag. Do not
rewrite the solution; it is verified working and Bill signed off on the
behavior.

## What chapter 14 is

"The Channel Nobody Tested", an accessibility chapter. Two deliverables in the
student exercise:

1. A **speech pipeline** that turns a stream of arbitrary text fragments into
   utterances a person can listen to.
2. A **blind persona** for the chapter 13 virtual user, which perceives the
   agent through speech alone and has no DOM access.

The thesis, which explains the grader's weighting: an automated observer only
finds bugs in the channel it needs to succeed. Chapter 13's virtual user read
the DOM, so the speech channel went untested and accumulated six defects.

## Read these first

| file | why |
|---|---|
| `book/chapter-14.md` | The TL;DR section IS the grader contract. 15 numbered rules. |
| `book/chapter-14-outline.md` | Proposed check table, mutation requirements, open questions. |
| `agent/web/gui/tts.js` | The speech pipeline. The thing under test. |
| `agent/web/gui/artifact-scroll.js` | The four feed sites into speech. |
| `agent/web/gui/gui.js` | The pause gate (`updateGate`). |
| `agent/web/gui/mcp.js` | `tts_listen`, `tts_log`, `tts_queue` MCP tools. |
| `agent/cmd/virtual-user/main.go` | The blind persona, around line 156. |
| `internal/grade/ch13_checks.go` | The house grader style to match. |
| `internal/grade/ch13_harness.go` | `FakeMCPServer` pattern, if you need to drive the GUI. |

## The six commits that built the solution

```
a26b358  pause gate speaking half + speak errors
30a090e  blind persona (virtual user --persona blind)
75888a5  buffer to phrase boundaries, flush(), filter for the ear
a959042  resolve fences before splitting; newlines are spaces
918be35  speak parts that arrive whole (streaming-off silence)
6165961  apply settings by presence, not truthiness
```

`git show <hash>` each one if you need the before state. The pre-fix code is
the best source of mutation fixtures, because those mutations are real.

## Build the grader

### Mechanism (Bill approved this)

The speech pipeline is plain JavaScript with one browser dependency. **Load the
student's speech module under `node` with a stubbed `speechSynthesis`, feed a
fixed fragment sequence, assert on the transcript.** Deterministic, no API key,
no browser, no model, no flake.

Bill ruled node is an acceptable grader prerequisite. Emit a clear prerequisite
error if node is absent rather than silently skipping checks.

### The harness pattern, which is verified working

Two gotchas cost real time today. Both are non-obvious:

```javascript
// 1. eval does NOT leak the const into scope. Append the export yourself.
eval(fs.readFileSync('web/gui/tts.js','utf8') + '; global.TTS = TTS;');

// 2. global.window MUST carry speechSynthesis. queueChunk returns early on
//    !('speechSynthesis' in window), and the harness then silently produces
//    NOTHING, which looks exactly like a buffering bug.
global.window = { speechSynthesis: stub };
```

### The checks

| id | pts | property |
|---|---|---|
| `tts-buffers-fragments` | 20 | No utterance breaks a word. Text reassembles. |
| `tts-filters-markup` | 20 | No markup marker reaches the channel. A fence is named. |
| `tts-boundaries` | 15 | Newline is a space, blank line is a boundary, trailing text flushes. |
| `tts-speaks-unstreamed` | 15 | A whole part is spoken. A streamed part is not doubled. |
| `tts-expands-identifiers` | 10 | Identifiers split into words. Ordinary words untouched. |
| `tts-gate-both-causes` | 10 | Unpause only when both causes clear. A space counts as input. |
| `ch13-parity` | 10 | Chapter 13 still passes. |

Total 100.

### Fixtures that are known to discriminate

These are the exact cases verified against the reference today, 6 of 6 passing:

- `"Hel"`, `"lo wor"`, `"ld. Bye."` produces one utterance `"Hello world."`
  then `"Bye."`. Nothing is spoken until a boundary arrives.
- A fenced code block **as a single chunk** produces `"code block."` and no
  backtick.
- The same fence **split across deltas** produces the same thing.
- A sentence wrapped across a newline produces ONE utterance.
- A blank line produces TWO utterances.
- `"Next up is this:"` plus `flush()` is spoken rather than stranded.

Normalization, verified 9 of 9: bold, inline code, camelCase, snake_case,
HTTPServer, headings, bullets, links, and plain words correctly left alone.

### Two weighting decisions you should not quietly revise

`tts-filters-markup` and `tts-speaks-unstreamed` carry the most points because
those two defects survived a passing grader suite, an automated observer, and a
human listening to the output every working day.

**The fence fixture must run BOTH as one chunk and split across deltas.** The
single-chunk case is the one that disproved the streaming theory. A grader that
only tests the split case passes an implementation that still speaks backticks.

**The identifier check must assert ordinary words are untouched.** That is the
mutation guard. An implementation that inserts a space before every capital
letter passes the positive case and fails this one.

### Mutation tests (P9, required)

Each check must FAIL when its behavior is deleted from the reference. At
minimum:

- Remove the buffer, so fragments enqueue raw. `tts-buffers-fragments` fails.
- Make fence resolution per-line instead of whole-buffer. `tts-filters-markup`
  fails. This is the real pre-fix bug; see `a959042^`.
- Delete the `accumulated` guard in `_handleFinal`. `tts-speaks-unstreamed`
  must fail **on the doubling side, not only the silence side**. Test both
  directions or the check is half a check.
- Neuter the edge comparison in `updateGate`. `tts-gate-both-causes` fails.

Log the details of each mutation run. A mutant that matches zero tests is a
stale mutant, not a passing grader.

### Also required

- `make grade14` target in the Makefile. It does not exist yet.
- `solutions/ch14` snapshot. **Do NOT commit the `ensemble` binary or
  `settings.json` into it.** `solutions/ch09..ch13` currently track ~49MB of
  committed binaries and Bill's personal `tts_speed`. Pre-existing, flagged to
  him twice, not yours to fix, but do not extend it.
- Tag `ch14-solution`.
- A short review doc for the author at `book/review-ch14.md`, noting anything
  the chapter claims that the grader cannot actually check.

## Definition of done

- `make grade14` reports 100/100 against the reference.
- Every prior grader still passes: ch2, 3, 5, 7, 8, 9, 10, 11, 12, 13.
- `go vet ./...` clean, `go test ./... -race` clean.
- Mutation results recorded.
- Grader does not read the reference solution.

## Environment

Repo `~/projects/ensemble`.

```bash
cd agent && go build -o ensemble ./cmd/ && go build -o virtual-user ./cmd/virtual-user/
```

**Server start.** Three gotchas, each cost time today:

```bash
cd ~/projects/ensemble/agent
export EN_PRIMARY_SKILL=ensemble
export ANTHROPIC_API_KEY=$(python3 -c "import json; print(json.load(open('$HOME/.cr/settings.json'))['directClaudeAPIKey'])")
tail -f /dev/null | ./ensemble --port 8084 --skills-dir ./skills --gui-debug
```

1. Env vars must be **exported**, not prefixed on the pipeline. Prefixing puts
   them on `tail` and the agent starts with the wrong primary skill.
2. Stdin must stay **open**. Both `&` and `< /dev/null` make it exit instantly.
3. Run `lsof -nP -i :8084 | grep LISTEN` **first**. A stale server silently
   holds the port and your rebuilt binary then serves nothing. Kill it with the
   `kill_job` tool.
4. `virtual-user` needs `ANTHROPIC_API_KEY` exported in its own shell too.
5. Bill keeps a Chrome tab on `:8084`. `mcp.js` only exists in that tab, so MCP
   tools hang until he refreshes. Ask him; he refreshes on request.

**Graders**: `make grade{2,3,5,7,8,9,10,11,12,13}`. There is no grade1, grade4
or grade6 target; ch4 runs inside grade5's parity check. Score line greps as
`score: N/M`. Run `go clean -cache` after code changes, because the build cache
will serve a stale binary and you will grade the wrong thing.

**Commits**: never `git add -A`; `book/*` holds Bill's untracked drafts. Use a
heredoc, because backticks in `-m` get command-substituted:

```bash
git -c user.name='CodeRhapsody' -c user.email='coderhapsody@local' commit -F - <<'EOF'
...
EOF
```

## Traps, from today

**There are TWO tool dispatchers.** `Engine.Execute` in
`internal/llm/engine.go` serves ch3/ch4 via `eng.Ask()`. `Actor.dispatchTool`
in `internal/llm/actor.go` serves ch6+. Fixing only one left every ch4
behavioral check at 0/100, with a symptom (`Output: ""`, 0 bytes on disk) that
looked like broken output capture and was actually a lifetime bug.

**Silent debug output can mean wrong code path, not stale build.** A debug
print that never appeared sent me to `go clean -cache` when the real cause was
a second dispatcher.

**`grep -c` exits 1 on zero matches**, which reads as command failure in a
chained line.

**A stray `main` package file left in `agent/` breaks `go build ./cmd/`** with a
package collision. A previous virtual-user run left `hanoi.go` there.

**`edit_file` refuses an edit that removes a markdown heading.** Repeat the
heading verbatim in `new_text`, or pass `allow_heading_changes: true` for a
deliberate rename.

## Open questions for Bill

1. Whether `tts-gate-both-causes` should drive the real GUI through the MCP
   tunnel (like ch13's `FakeMCPServer`) or test `updateGate` in node like the
   rest. Node is simpler and has no browser dependency; the tunnel is more
   end-to-end. Recommend node for consistency.
2. Whether the blind-persona check should assert tool *absence* structurally
   (derived from the student's own tree, the way ch5's `framework-blind` check
   works) or behaviorally (the persona fails a DOM-only task). Structural is
   cheaper and less flaky.

## One thing the chapter claims that you should verify

Chapter 14 section 14.2 states that four call sites feed the speech channel,
all in `artifact-scroll.js`: streaming deltas, tool dispatch, a part arriving
complete, and errors. If you add or move a feed site while building the
grader, the chapter is now wrong and the author needs to know.
